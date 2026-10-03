package tablewriter

import (
	"io"
	"strings"

	"github.com/olekukonko/errors"
	"github.com/olekukonko/tablewriter/pkg/twwarp"
	"github.com/olekukonko/tablewriter/pkg/twwidth"
	"github.com/olekukonko/tablewriter/tw"
)

// prepareTableSection prepares either headers or footers for the table
func (t *Table) prepareTableSection(elements []any, config tw.CellConfig, sectionName string) [][]string {
	actualCellsToProcess := t.processVariadic(elements)
	t.logger.Debugf("%s(): Effective cells to process: %v", sectionName, actualCellsToProcess)

	stringsResult, err := t.convertCellsToStrings(actualCellsToProcess, config)
	if err != nil {
		t.logger.Errorf("%s(): Failed to convert elements to strings: %v", sectionName, err)
		stringsResult = []string{}
	}

	prepared := t.prepareContent(stringsResult, config, nil)
	numColsBatch := t.maxColumns()

	if len(prepared) > 0 {
		for i := range prepared {
			if len(prepared[i]) < numColsBatch {
				t.logger.Debugf("Padding %s line %d from %d to %d columns", sectionName, i, len(prepared[i]), numColsBatch)
				paddedLine := make([]string, numColsBatch)
				copy(paddedLine, prepared[i])
				for j := len(prepared[i]); j < numColsBatch; j++ {
					paddedLine[j] = tw.Empty
				}
				prepared[i] = paddedLine
			} else if len(prepared[i]) > numColsBatch {
				t.logger.Debugf("Truncating %s line %d from %d to %d columns", sectionName, i, len(prepared[i]), numColsBatch)
				prepared[i] = prepared[i][:numColsBatch]
			}
		}
	}

	return prepared
}

// processVariadic handles the common logic for processing variadic arguments
// that could be either individual elements or a slice of elements
func (t *Table) processVariadic(elements []any) []any {
	if len(elements) == 1 {
		switch v := elements[0].(type) {
		case []string:
			t.logger.Debugf("Detected single []string argument. Unpacking it (fast path).")
			out := make([]any, len(v))
			for i := range v {
				out[i] = v[i]
			}
			return out

		case []interface{}:
			t.logger.Debugf("Detected single []interface{} argument. Unpacking it (fast path).")
			out := make([]any, len(v))
			copy(out, v)
			return out
		}
	}

	t.logger.Debugf("Input has multiple elements or single non-slice. Using variadic elements as-is.")
	return elements
}

// prepareContent processes cell content with formatting and wrapping.
// Parameters include cells to process and config for formatting rules.
// Returns a slice of string slices representing processed lines.
func (t *Table) prepareContent(cells []string, config tw.CellConfig, resolvedWidths tw.Mapper[int, int]) [][]string {
	isStreaming := t.config.Stream.Enable && t.hasPrinted
	t.logger.Debugf("prepareContent: Processing cells=%v (streaming: %v)", cells, isStreaming)
	initialInputCellCount := len(cells)
	result := make([][]string, 0)

	effectiveNumCols := initialInputCellCount
	if isStreaming {
		if t.streamNumCols > 0 {
			effectiveNumCols = t.streamNumCols
			t.logger.Debugf("prepareContent: Streaming mode, using fixed streamNumCols: %d", effectiveNumCols)
			if len(cells) != effectiveNumCols {
				t.logger.Warnf("prepareContent: Streaming mode, input cell count (%d) does not match streamNumCols (%d). Input cells will be padded/truncated.", len(cells), effectiveNumCols)
				if len(cells) < effectiveNumCols {
					paddedCells := make([]string, effectiveNumCols)
					copy(paddedCells, cells)
					for i := len(cells); i < effectiveNumCols; i++ {
						paddedCells[i] = tw.Empty
					}
					cells = paddedCells
				} else if len(cells) > effectiveNumCols {
					cells = cells[:effectiveNumCols]
				}
			}
		} else {
			t.logger.Warnf("prepareContent: Streaming mode enabled but streamNumCols is 0. Using input cell count %d. Stream widths may not be available.", effectiveNumCols)
		}
	}

	for i := 0; i < effectiveNumCols; i++ {
		cellContent := ""
		if i < len(cells) {
			cellContent = cells[i]
		} else {
			cellContent = tw.Empty
		}

		cellContent = t.Trimmer(cellContent)

		if strings.Contains(cellContent, twwidth.TabString.String()) {
			// Get the detected width from the singleton
			width := twwidth.TabWidth()
			spaces := strings.Repeat(tw.Space, width)
			cellContent = strings.ReplaceAll(cellContent, twwidth.TabString.String(), spaces)
		}

		colPad := config.Padding.Global
		if i < len(config.Padding.PerColumn) && config.Padding.PerColumn[i].Paddable() {
			colPad = config.Padding.PerColumn[i]
		}

		padLeftWidth := twwidth.Width(colPad.Left)
		padRightWidth := twwidth.Width(colPad.Right)

		effectiveContentMaxWidth := t.calculateContentMaxWidth(i, config, padLeftWidth, padRightWidth, effectiveNumCols, resolvedWidths)

		if config.Formatting.AutoFormat.Enabled() {
			cellContent = tw.Title(strings.Join(tw.SplitCamelCase(cellContent), tw.Space))
		}

		lines := strings.Split(cellContent, "\n")
		finalLinesForCell := make([]string, 0)
		for _, line := range lines {
			if effectiveContentMaxWidth > 0 {
				switch config.Formatting.AutoWrap {
				case tw.WrapNormal:
					var wrapped []string
					if t.config.Behavior.TrimSpace.Enabled() && t.config.Behavior.TrimTab.Enabled() {
						wrapped, _ = twwarp.WrapString(line, effectiveContentMaxWidth)
					} else {
						wrapped, _ = twwarp.WrapStringWithSpaces(line, effectiveContentMaxWidth)
					}
					finalLinesForCell = append(finalLinesForCell, wrapped...)
				case tw.WrapTruncate:
					if twwidth.Width(line) > effectiveContentMaxWidth {
						ellipsisWidth := twwidth.Width(tw.CharEllipsis)
						if effectiveContentMaxWidth >= ellipsisWidth {
							finalLinesForCell = append(finalLinesForCell, twwidth.Truncate(line, effectiveContentMaxWidth-ellipsisWidth, tw.CharEllipsis))
						} else {
							finalLinesForCell = append(finalLinesForCell, twwidth.Truncate(line, effectiveContentMaxWidth, ""))
						}
					} else {
						finalLinesForCell = append(finalLinesForCell, line)
					}
				case tw.WrapBreak:
					wrapped := make([]string, 0)
					currentLine := line
					breakCharWidth := twwidth.Width(tw.CharBreak)
					for twwidth.Width(currentLine) > effectiveContentMaxWidth {
						targetWidth := max(effectiveContentMaxWidth-breakCharWidth, 0)
						breakPoint := tw.BreakPoint(currentLine, targetWidth)
						runes := []rune(currentLine)
						if breakPoint <= 0 || breakPoint > len(runes) {
							t.logger.Warnf("prepareContent: WrapBreak - Invalid BreakPoint %d for line '%s' at width %d. Attempting manual break.", breakPoint, currentLine, targetWidth)
							actualBreakRuneCount := 0
							tempWidth := 0
							for charIdx, r := range runes {
								runeStr := string(r)
								rw := twwidth.Width(runeStr)
								if tempWidth+rw > targetWidth && charIdx > 0 {
									break
								}
								tempWidth += rw
								actualBreakRuneCount = charIdx + 1
								if tempWidth >= targetWidth && charIdx == 0 {
									break
								}
							}
							if actualBreakRuneCount == 0 && len(runes) > 0 {
								actualBreakRuneCount = 1
							}
							if actualBreakRuneCount > 0 && actualBreakRuneCount <= len(runes) {
								wrapped = append(wrapped, string(runes[:actualBreakRuneCount])+tw.CharBreak)
								currentLine = string(runes[actualBreakRuneCount:])
							} else {
								t.logger.Warnf("prepareContent: WrapBreak - Cannot break line '%s'. Adding as is.", currentLine)
								wrapped = append(wrapped, currentLine)
								currentLine = ""
								break
							}
						} else {
							wrapped = append(wrapped, string(runes[:breakPoint])+tw.CharBreak)
							currentLine = string(runes[breakPoint:])
						}
					}
					if twwidth.Width(currentLine) > 0 {
						wrapped = append(wrapped, currentLine)
					}
					if len(wrapped) == 0 && twwidth.Width(line) > 0 && len(finalLinesForCell) == 0 {
						finalLinesForCell = append(finalLinesForCell, line)
					} else {
						finalLinesForCell = append(finalLinesForCell, wrapped...)
					}
				default:
					finalLinesForCell = append(finalLinesForCell, line)
				}
			} else {
				finalLinesForCell = append(finalLinesForCell, line)
			}
		}

		for len(result) < len(finalLinesForCell) {
			newRow := make([]string, effectiveNumCols)
			for j := range newRow {
				newRow[j] = tw.Empty
			}
			result = append(result, newRow)
		}

		for j := 0; j < len(result); j++ {
			cellLineContent := tw.Empty
			if j < len(finalLinesForCell) {
				cellLineContent = finalLinesForCell[j]
			}
			if i < len(result[j]) {
				result[j][i] = cellLineContent
			} else {
				t.logger.Warnf("prepareContent: Column index %d out of bounds (%d) during result matrix population. EffectiveNumCols: %d. This indicates a logic error.",
					i, len(result[j]), effectiveNumCols)
			}
		}
	}

	t.logger.Debugf("prepareContent: Content prepared, result %d lines.", len(result))
	return result
}

// prepareContexts initializes rendering and merge contexts.
// No parameters are required.
// Returns renderContext, mergeContext, and an error if initialization fails.
func (t *Table) prepareContexts() (*renderContext, *mergeContext, error) {
	numOriginalCols := t.maxColumns()
	t.logger.Debugf("prepareContexts: Original number of columns: %d", numOriginalCols)

	ctx := &renderContext{
		table:    t,
		renderer: t.renderer,
		cfg:      t.renderer.Config(),
		numCols:  numOriginalCols,
		widths: map[tw.Position]tw.Mapper[int, int]{
			tw.Header: tw.NewMapper[int, int](),
			tw.Row:    tw.NewMapper[int, int](),
			tw.Footer: tw.NewMapper[int, int](),
		},
		logger: t.logger,
	}

	// Process raw rows into visual, multi-line rows (First Pass without final constraints)
	processedRowLines := make([][][]string, len(t.rows))
	for i, rawRow := range t.rows {
		processedRowLines[i] = t.prepareContent(rawRow, t.config.Row, nil)
	}
	ctx.rowLines = processedRowLines

	if len(t.rawHeaders) > 0 {
		ctx.headerLines = t.prepareContent(t.rawHeaders, t.config.Header, nil)
	} else {
		ctx.headerLines = nil
	}

	if len(t.rawFooters) > 0 {
		ctx.footerLines = t.prepareContent(t.rawFooters, t.config.Footer, nil)
	} else {
		ctx.footerLines = nil
	}

	isEmpty, visibleCount := t.getEmptyColumnInfo(ctx.rowLines, numOriginalCols)
	ctx.emptyColumns = isEmpty
	ctx.visibleColCount = visibleCount

	mctx := &mergeContext{
		headerMerges: make(map[int]tw.MergeState),
		rowMerges:    make([]map[int]tw.MergeState, len(ctx.rowLines)),
		footerMerges: make(map[int]tw.MergeState),
		horzMerges:   make(map[tw.Position]map[int]bool),
	}
	for i := range mctx.rowMerges {
		mctx.rowMerges[i] = make(map[int]tw.MergeState)
	}

	if err := t.calculateAndNormalizeWidths(ctx); err != nil {
		t.logger.Debugf("Error during initial width calculation: %v", err)
		return nil, nil, err
	}
	t.logger.Debugf("Initial normalized widths (before hiding): H=%v, R=%v, F=%v",
		ctx.widths[tw.Header], ctx.widths[tw.Row], ctx.widths[tw.Footer])

	// ONLY run Pass 2 if a global width constraint is active, which means calculateAndNormalizeWidths
	// might have dynamically scaled columns down. If not active, the Pass 1 prepared content is perfectly sized.
	globalLimitActive := t.config.Widths.Global > 0 || t.config.MaxWidth > 0

	if globalLimitActive {
		if len(t.rawHeaders) > 0 {
			reWrappedHeader := t.prepareContent(t.rawHeaders, t.config.Header, ctx.widths[tw.Header])
			ctx.headerLines, mctx.headerMerges, _ = t.prepareWithMerges(reWrappedHeader, t.config.Header, tw.Header)
			t.headers = ctx.headerLines
		} else {
			t.headers = nil
		}

		// Re-process row lines for merges now that widths are fully known
		processedRowLinesWithMerges := make([][][]string, len(t.rows))
		for i, rawRow := range t.rows {
			if mctx.rowMerges[i] == nil {
				mctx.rowMerges[i] = make(map[int]tw.MergeState)
			}
			reWrappedRow := t.prepareContent(rawRow, t.config.Row, ctx.widths[tw.Row])
			processedRowLinesWithMerges[i], mctx.rowMerges[i], _ = t.prepareWithMerges(reWrappedRow, t.config.Row, tw.Row)
		}
		ctx.rowLines = processedRowLinesWithMerges
	} else {
		// Use Pass 1 contents and just compute merges
		if len(t.rawHeaders) > 0 {
			ctx.headerLines, mctx.headerMerges, _ = t.prepareWithMerges(ctx.headerLines, t.config.Header, tw.Header)
			t.headers = ctx.headerLines
		} else {
			t.headers = nil
		}

		processedRowLinesWithMerges := make([][][]string, len(t.rows))
		for i := range ctx.rowLines {
			if mctx.rowMerges[i] == nil {
				mctx.rowMerges[i] = make(map[int]tw.MergeState)
			}
			processedRowLinesWithMerges[i], mctx.rowMerges[i], _ = t.prepareWithMerges(ctx.rowLines[i], t.config.Row, tw.Row)
		}
		ctx.rowLines = processedRowLinesWithMerges
	}

	t.applyHorizontalMerges(tw.Header, ctx, mctx.headerMerges)

	mergeMode := t.config.Row.Merging.Mode
	if mergeMode == 0 {
		mergeMode = t.config.Row.Formatting.MergeMode
	}

	// Now check against the effective mode
	if mergeMode&tw.MergeVertical != 0 {
		t.applyVerticalMerges(ctx, mctx)
	}
	if mergeMode&tw.MergeHierarchical != 0 {
		t.applyHierarchicalMerges(ctx, mctx)
	}

	t.prepareFooter(ctx, mctx)
	t.logger.Debugf("Footer prepared. Widths before hiding: H=%v, R=%v, F=%v",
		ctx.widths[tw.Header], ctx.widths[tw.Row], ctx.widths[tw.Footer])

	if t.config.Behavior.AutoHide.Enabled() {
		t.logger.Debugf("Applying AutoHide: Adjusting widths for empty columns.")
		if ctx.emptyColumns == nil {
			t.logger.Debugf("Warning: ctx.emptyColumns is nil during width adjustment.")
		} else if len(ctx.emptyColumns) != ctx.numCols {
			t.logger.Debugf("Warning: Length mismatch between emptyColumns (%d) and numCols (%d). Skipping adjustment.", len(ctx.emptyColumns), ctx.numCols)
		} else {
			for colIdx := 0; colIdx < ctx.numCols; colIdx++ {
				if ctx.emptyColumns[colIdx] {
					t.logger.Debugf("AutoHide: Hiding column %d by setting width to 0.", colIdx)
					ctx.widths[tw.Header].Set(colIdx, 0)
					ctx.widths[tw.Row].Set(colIdx, 0)
					ctx.widths[tw.Footer].Set(colIdx, 0)
				}
			}
			t.logger.Debugf("Widths after AutoHide adjustment: H=%v, R=%v, F=%v",
				ctx.widths[tw.Header], ctx.widths[tw.Row], ctx.widths[tw.Footer])
		}
	} else {
		t.logger.Debugf("AutoHide is disabled, skipping width adjustment.")
	}
	t.logger.Debugf("prepareContexts completed all stages.")
	return ctx, mctx, nil
}

// render generates the table output using the configured renderer.
// No parameters are required.
// Returns an error if rendering fails in any section.
func (t *Table) render() error {
	t.ensureInitialized()

	// Save the original writer and schedule its restoration upon function exit.
	// This guarantees the table's writer is restored even if errors occur.
	originalWriter := t.writer
	defer func() {
		t.writer = originalWriter
	}()

	// If a counter is active, wrap the writer in a MultiWriter.
	if len(t.counters) > 0 {
		// The slice must be of type io.Writer.
		// Start it with the original destination writer.
		allWriters := []io.Writer{originalWriter}

		// Append each counter to the slice of writers.
		for _, c := range t.counters {
			allWriters = append(allWriters, c)
		}

		// Create a MultiWriter that broadcasts to the original writer AND all counters.
		t.writer = io.MultiWriter(allWriters...)
	}

	if t.config.Stream.Enable {
		t.logger.Warn("Render() called in streaming mode. Use Start/Append/Close methods instead.")
		return errors.New("render called in streaming mode; use Start/Append/Close")
	}

	// Calculate and cache the column count for this specific batch render pass.
	t.batchRenderNumCols = t.maxColumns()
	t.isBatchRenderNumColsSet = true
	defer func() {
		t.isBatchRenderNumColsSet = false
		t.logger.Debugf("Render(): Cleared isBatchRenderNumColsSet to false (batchRenderNumCols was %d).", t.batchRenderNumCols)
	}()

	hasCaption := t.caption.Text != "" && t.caption.Spot != tw.SpotNone
	isTopOrBottomCaption := hasCaption &&
		(t.caption.Spot >= tw.SpotTopLeft && t.caption.Spot <= tw.SpotBottomRight)

	var tableStringBuffer *strings.Builder
	targetWriter := t.writer // Can be the original writer or the MultiWriter.

	// If a caption is present, the main table content must be rendered to an
	// in-memory buffer first to calculate its final width.
	if isTopOrBottomCaption {
		tableStringBuffer = &strings.Builder{}
		targetWriter = tableStringBuffer
		t.logger.Debugf("Top/Bottom caption detected. Rendering table core to buffer first.")
	} else {
		t.logger.Debugf("No caption detected. Rendering table core directly to writer.")
	}

	// Point the table's writer to the target (either the final destination or the buffer).
	t.writer = targetWriter
	ctx, mctx, err := t.prepareContexts()
	if err != nil {
		t.logger.Errorf("prepareContexts failed: %v", err)
		return errors.Newf("failed to prepare table contexts").Wrap(err)
	}

	if err := ctx.renderer.Start(t.writer); err != nil {
		t.logger.Errorf("Renderer Start() error: %v", err)
		return errors.Newf("renderer start failed").Wrap(err)
	}

	renderError := false
	var firstRenderErr error
	renderFuncs := []func(*renderContext, *mergeContext) error{
		t.renderHeader,
		t.renderRow,
		t.renderFooter,
	}
	for i, renderFn := range renderFuncs {
		sectionName := []string{"Header", "Row", "Footer"}[i]
		if renderErr := renderFn(ctx, mctx); renderErr != nil {
			t.logger.Errorf("Renderer section error (%s): %v", sectionName, renderErr)
			if !renderError {
				firstRenderErr = errors.Newf("failed to render %s section", sectionName).Wrap(renderErr)
			}
			renderError = true
			break
		}
	}

	if closeErr := ctx.renderer.Close(); closeErr != nil {
		t.logger.Errorf("Renderer Close() error: %v", closeErr)
		if !renderError {
			firstRenderErr = errors.Newf("renderer close failed").Wrap(closeErr)
		}
		renderError = true
	}

	// Restore the writer to the original for the caption-handling logic.
	// This is necessary because the caption must be written to the final
	// destination, not the temporary buffer used for the table body.
	t.writer = originalWriter

	if renderError {
		return firstRenderErr
	}

	// Caption Handling & Final Output
	if isTopOrBottomCaption {
		renderedTableContent := tableStringBuffer.String()
		t.logger.Debugf("[Render] Table core buffer length: %d", len(renderedTableContent))

		// Handle edge case where table is empty but should have borders.
		shouldHaveBorders := t.renderer != nil && (t.renderer.Config().Borders.Top.Enabled() || t.renderer.Config().Borders.Bottom.Enabled())
		if len(renderedTableContent) == 0 && shouldHaveBorders {
			var sb strings.Builder
			if t.renderer.Config().Borders.Top.Enabled() {
				sb.WriteString("+--+")
				sb.WriteString(t.newLine)
			}
			if t.renderer.Config().Borders.Bottom.Enabled() {
				sb.WriteString("+--+")
			}
			renderedTableContent = sb.String()
			t.logger.Warnf("[Render] Table buffer was empty despite enabled borders. Manually generated minimal output: %q", renderedTableContent)
		}

		actualTableWidth := 0
		trimmedBuffer := strings.TrimRight(renderedTableContent, "\r\n \t")
		for _, line := range strings.Split(trimmedBuffer, "\n") {
			w := twwidth.Width(line)
			if w > actualTableWidth {
				actualTableWidth = w
			}
		}
		t.logger.Debugf("[Render] Calculated actual table width: %d (from content: %q)", actualTableWidth, renderedTableContent)

		isTopCaption := t.caption.Spot >= tw.SpotTopLeft && t.caption.Spot <= tw.SpotTopRight

		if isTopCaption {
			t.logger.Debugf("[Render] Printing Top Caption.")
			t.printTopBottomCaption(t.writer, actualTableWidth)
		}

		if len(renderedTableContent) > 0 {
			t.logger.Debugf("[Render] Printing table content (length %d) to final writer.", len(renderedTableContent))
			t.writer.Write([]byte(renderedTableContent))
			if !isTopCaption && t.caption.Text != "" && !strings.HasSuffix(renderedTableContent, t.newLine) {
				t.writer.Write([]byte(tw.NewLine))
				t.logger.Debugf("[Render] Added trailing newline after table content before bottom caption.")
			}
		} else {
			t.logger.Debugf("[Render] No table content (original buffer or generated) to print.")
		}

		if !isTopCaption {
			t.logger.Debugf("[Render] Calling printTopBottomCaption for Bottom Caption. Width: %d", actualTableWidth)
			t.printTopBottomCaption(t.writer, actualTableWidth)
			t.logger.Debugf("[Render] Returned from printTopBottomCaption for Bottom Caption.")
		}
	}

	t.hasPrinted = true
	t.logger.Info("Render() completed.")
	return nil
}

// renderFooter renders the table's footer section with borders and padding.
// Parameters ctx and mctx hold rendering and merge state.
// Returns an error if rendering fails.
func (t *Table) renderFooter(ctx *renderContext, mctx *mergeContext) error {
	if !ctx.footerPrepared {
		t.prepareFooter(ctx, mctx)
	}

	f := ctx.renderer
	cfg := ctx.cfg

	hasContent := len(ctx.footerLines) > 0
	hasTopPadding := t.config.Footer.Padding.Global.Top != tw.Empty
	hasBottomPaddingConfig := t.config.Footer.Padding.Global.Bottom != tw.Empty || t.hasPerColumnBottomPadding()
	hasAnyFooterElement := hasContent || hasTopPadding || hasBottomPaddingConfig

	if !hasAnyFooterElement {
		hasContentAbove := len(ctx.rowLines) > 0 || len(ctx.headerLines) > 0
		if hasContentAbove && cfg.Borders.Bottom.Enabled() && cfg.Settings.Lines.ShowBottom.Enabled() {
			ctx.logger.Debugf("Footer is empty, rendering table bottom border based on last row/header")
			var lastLineAboveCtx *helperContext
			var lastLineAligns map[int]tw.Align
			var lastLinePadding map[int]tw.Padding

			if len(ctx.rowLines) > 0 {
				lastRowIdx := len(ctx.rowLines) - 1
				lastRowLineIdx := -1
				var lastRowLine []string
				if lastRowIdx >= 0 && len(ctx.rowLines[lastRowIdx]) > 0 {
					lastRowLineIdx = len(ctx.rowLines[lastRowIdx]) - 1
					lastRowLine = padLine(ctx.rowLines[lastRowIdx][lastRowLineIdx], ctx.numCols)
				} else {
					lastRowLine = make([]string, ctx.numCols)
				}
				lastLineAboveCtx = &helperContext{
					position: tw.Row,
					rowIdx:   lastRowIdx,
					lineIdx:  lastRowLineIdx,
					line:     lastRowLine,
					location: tw.LocationEnd,
				}
				lastLineAligns = t.buildAligns(t.config.Row)
				lastLinePadding = t.buildPadding(t.config.Row.Padding)
			} else {
				lastHeaderLineIdx := -1
				var lastHeaderLine []string
				if len(ctx.headerLines) > 0 {
					lastHeaderLineIdx = len(ctx.headerLines) - 1
					lastHeaderLine = padLine(ctx.headerLines[lastHeaderLineIdx], ctx.numCols)
				} else {
					lastHeaderLine = make([]string, ctx.numCols)
				}
				lastLineAboveCtx = &helperContext{
					position: tw.Header,
					rowIdx:   0,
					lineIdx:  lastHeaderLineIdx,
					line:     lastHeaderLine,
					location: tw.LocationEnd,
				}
				lastLineAligns = t.buildAligns(t.config.Header)
				lastLinePadding = t.buildPadding(t.config.Header.Padding)
			}

			resp := t.buildCellContexts(ctx, mctx, lastLineAboveCtx, lastLineAligns, lastLinePadding)
			ctx.logger.Debugf("Bottom border: Using Widths=%v", ctx.widths[tw.Row])
			f.Line(tw.Formatting{
				Row: tw.RowContext{
					Widths:       ctx.widths[tw.Row],
					Current:      resp.cells,
					Previous:     resp.prevCells,
					Position:     lastLineAboveCtx.position,
					Location:     tw.LocationEnd,
					ColMaxWidths: t.getColMaxWidths(tw.Footer),
				},
				Level:    tw.LevelFooter,
				IsSubRow: false,
			})
		} else {
			ctx.logger.Debugf("Footer is empty and no content above or borders disabled, skipping footer render")
		}
		return nil
	}

	ctx.logger.Debugf("Rendering footer section (has elements)")
	hasContentAbove := len(ctx.rowLines) > 0 || len(ctx.headerLines) > 0
	colAligns := t.buildAligns(t.config.Footer)
	colPadding := t.buildPadding(t.config.Footer.Padding)
	hctx := &helperContext{position: tw.Footer}
	// Declare paddingLineContentForContext with a default value
	paddingLineContentForContext := make([]string, ctx.numCols)

	if hasContentAbove && cfg.Settings.Lines.ShowFooterLine.Enabled() && !hasTopPadding && len(ctx.footerLines) > 0 {
		ctx.logger.Debugf("Rendering footer separator line")
		var lastLineAboveCtx *helperContext
		var lastLineAligns map[int]tw.Align
		var lastLinePadding map[int]tw.Padding
		var lastLinePosition tw.Position

		if len(ctx.rowLines) > 0 {
			lastRowIdx := len(ctx.rowLines) - 1
			lastRowLineIdx := -1
			var lastRowLine []string
			if lastRowIdx >= 0 && len(ctx.rowLines[lastRowIdx]) > 0 {
				lastRowLineIdx = len(ctx.rowLines[lastRowIdx]) - 1
				lastRowLine = padLine(ctx.rowLines[lastRowIdx][lastRowLineIdx], ctx.numCols)
			} else {
				lastRowLine = make([]string, ctx.numCols)
			}
			lastLineAboveCtx = &helperContext{
				position: tw.Row,
				rowIdx:   lastRowIdx,
				lineIdx:  lastRowLineIdx,
				line:     lastRowLine,
				location: tw.LocationMiddle,
			}
			lastLineAligns = t.buildAligns(t.config.Row)
			lastLinePadding = t.buildPadding(t.config.Row.Padding)
			lastLinePosition = tw.Row
		} else {
			lastHeaderLineIdx := -1
			var lastHeaderLine []string
			if len(ctx.headerLines) > 0 {
				lastHeaderLineIdx = len(ctx.headerLines) - 1
				lastHeaderLine = padLine(ctx.headerLines[lastHeaderLineIdx], ctx.numCols)
			} else {
				lastHeaderLine = make([]string, ctx.numCols)
			}
			lastLineAboveCtx = &helperContext{
				position: tw.Header,
				rowIdx:   0,
				lineIdx:  lastHeaderLineIdx,
				line:     lastHeaderLine,
				location: tw.LocationMiddle,
			}
			lastLineAligns = t.buildAligns(t.config.Header)
			lastLinePadding = t.buildPadding(t.config.Header.Padding)
			lastLinePosition = tw.Header
		}

		resp := t.buildCellContexts(ctx, mctx, lastLineAboveCtx, lastLineAligns, lastLinePadding)
		var nextCells map[int]tw.CellContext
		if hasContent {
			nextCells = make(map[int]tw.CellContext)
			for j, cellData := range padLine(ctx.footerLines[0], ctx.numCols) {
				mergeState := tw.MergeState{}
				if mctx.footerMerges != nil {
					mergeState = mctx.footerMerges[j]
				}
				nextCells[j] = tw.CellContext{Data: cellData, Merge: mergeState, Width: ctx.widths[tw.Footer].Get(j)}
			}
		}
		ctx.logger.Debugf("Footer separator: Using Widths=%v", ctx.widths[tw.Row])
		f.Line(tw.Formatting{
			Row: tw.RowContext{
				Widths:       ctx.widths[tw.Row],
				Current:      resp.cells,
				Previous:     resp.prevCells,
				Next:         nextCells,
				Position:     lastLinePosition,
				Location:     tw.LocationMiddle,
				ColMaxWidths: t.getColMaxWidths(tw.Footer),
			},
			Level:     tw.LevelFooter,
			IsSubRow:  false,
			HasFooter: true,
		})
	}

	if hasTopPadding {
		hctx.rowIdx = 0
		hctx.lineIdx = -1
		if !hasContentAbove || !cfg.Settings.Lines.ShowFooterLine.Enabled() {
			hctx.location = tw.LocationFirst
		} else {
			hctx.location = tw.LocationMiddle
		}
		hctx.line = t.buildPaddingLineContents(t.config.Footer.Padding.Global.Top, ctx.widths[tw.Footer], ctx.numCols, mctx.footerMerges)
		ctx.logger.Debugf("Calling renderPadding for Footer Top Padding line: %v (loc: %v)", hctx.line, hctx.location)
		if err := t.renderPadding(ctx, mctx, hctx, t.config.Footer.Padding.Global.Top); err != nil {
			return err
		}
	}

	lastRenderedLineIdx := -2
	if hasTopPadding {
		lastRenderedLineIdx = -1
	}
	for i, line := range ctx.footerLines {
		hctx.rowIdx = 0
		hctx.lineIdx = i
		hctx.line = padLine(line, ctx.numCols)
		isFirstContentLine := i == 0
		isLastContentLine := i == len(ctx.footerLines)-1
		if isFirstContentLine && !hasTopPadding && (!hasContentAbove || !cfg.Settings.Lines.ShowFooterLine.Enabled()) {
			hctx.location = tw.LocationFirst
		} else if isLastContentLine && !hasBottomPaddingConfig {
			hctx.location = tw.LocationEnd
		} else {
			hctx.location = tw.LocationMiddle
		}
		ctx.logger.Debugf("Rendering footer content line %d with location %v", i, hctx.location)
		if err := t.renderLine(ctx, mctx, hctx, colAligns, colPadding); err != nil {
			return err
		}
		lastRenderedLineIdx = i
	}

	if hasBottomPaddingConfig {
		paddingLineContentForContext = make([]string, ctx.numCols)
		formattedPaddingCells := make([]string, ctx.numCols)
		representativePadChar := " "
		ctx.logger.Debugf("Constructing Footer Bottom Padding line content strings")
		for j := 0; j < ctx.numCols; j++ {
			colWd := ctx.widths[tw.Footer].Get(j)
			mergeState := tw.MergeState{}
			if mctx.footerMerges != nil {
				if state, ok := mctx.footerMerges[j]; ok {
					mergeState = state
				}
			}
			if mergeState.Horizontal.Present && !mergeState.Horizontal.Start {
				paddingLineContentForContext[j] = ""
				formattedPaddingCells[j] = ""
				continue
			}
			padChar := " "
			if j < len(t.config.Footer.Padding.PerColumn) && t.config.Footer.Padding.PerColumn[j].Bottom != tw.Empty {
				padChar = t.config.Footer.Padding.PerColumn[j].Bottom
			} else if t.config.Footer.Padding.Global.Bottom != tw.Empty {
				padChar = t.config.Footer.Padding.Global.Bottom
			}
			paddingLineContentForContext[j] = padChar
			if j == 0 || representativePadChar == " " {
				representativePadChar = padChar
			}
			padWidth := max(twwidth.Width(padChar), 1)
			repeatCount := 0
			if colWd > 0 && padWidth > 0 {
				repeatCount = colWd / padWidth
			}
			if colWd > 0 && repeatCount < 1 && padChar != " " {
				repeatCount = 1
			}
			if colWd == 0 {
				repeatCount = 0
			}
			rawPaddingContent := strings.Repeat(padChar, repeatCount)
			currentWd := twwidth.Width(rawPaddingContent)
			if currentWd < colWd {
				rawPaddingContent += strings.Repeat(" ", colWd-currentWd)
			}
			if currentWd > colWd && colWd > 0 {
				rawPaddingContent = twwidth.Truncate(rawPaddingContent, colWd)
			}
			if colWd == 0 {
				rawPaddingContent = ""
			}
			formattedPaddingCells[j] = rawPaddingContent
		}
		ctx.logger.Debugf("Manually rendering Footer Bottom Padding line (char like '%s')", representativePadChar)
		var paddingLineOutput strings.Builder
		if cfg.Borders.Left.Enabled() {
			paddingLineOutput.WriteString(cfg.Symbols.Column())
		}
		for colIdx := 0; colIdx < ctx.numCols; {
			if colIdx > 0 && cfg.Settings.Separators.BetweenColumns.Enabled() {
				shouldAddSeparator := true
				if prevMergeState, ok := mctx.footerMerges[colIdx-1]; ok {
					if prevMergeState.Horizontal.Present && !prevMergeState.Horizontal.End {
						shouldAddSeparator = false
					}
				}
				if shouldAddSeparator {
					paddingLineOutput.WriteString(cfg.Symbols.Column())
				}
			}
			if colIdx < len(formattedPaddingCells) {
				paddingLineOutput.WriteString(formattedPaddingCells[colIdx])
			}
			currentMergeState := tw.MergeState{}
			if mctx.footerMerges != nil {
				if state, ok := mctx.footerMerges[colIdx]; ok {
					currentMergeState = state
				}
			}
			if currentMergeState.Horizontal.Present && currentMergeState.Horizontal.Start {
				colIdx += currentMergeState.Horizontal.Span
			} else {
				colIdx++
			}
		}
		if cfg.Borders.Right.Enabled() {
			paddingLineOutput.WriteString(cfg.Symbols.Column())
		}
		paddingLineOutput.WriteString(t.newLine)
		t.writer.Write([]byte(paddingLineOutput.String()))
		ctx.logger.Debugf("Manually rendered Footer Bottom Padding line: %s", strings.TrimSuffix(paddingLineOutput.String(), t.newLine))
		hctx.rowIdx = 0
		hctx.lineIdx = len(ctx.footerLines)
		hctx.line = paddingLineContentForContext
		hctx.location = tw.LocationEnd
		lastRenderedLineIdx = hctx.lineIdx
	}

	if cfg.Borders.Bottom.Enabled() && cfg.Settings.Lines.ShowBottom.Enabled() {
		ctx.logger.Debugf("Rendering final table bottom border")
		if lastRenderedLineIdx == len(ctx.footerLines) {
			hctx.rowIdx = 0
			hctx.lineIdx = lastRenderedLineIdx
			hctx.line = paddingLineContentForContext
			hctx.location = tw.LocationEnd
			ctx.logger.Debugf("Setting border context based on bottom padding line")
		} else if lastRenderedLineIdx >= 0 {
			hctx.rowIdx = 0
			hctx.lineIdx = lastRenderedLineIdx
			hctx.line = padLine(ctx.footerLines[hctx.lineIdx], ctx.numCols)
			hctx.location = tw.LocationEnd
			ctx.logger.Debugf("Setting border context based on last content line idx %d", hctx.lineIdx)
		} else if lastRenderedLineIdx == -1 {
			hctx.rowIdx = 0
			hctx.lineIdx = -1
			hctx.line = paddingLineContentForContext
			hctx.location = tw.LocationEnd
			ctx.logger.Debugf("Setting border context based on top padding line")
		} else {
			hctx.rowIdx = 0
			hctx.lineIdx = -2
			hctx.line = make([]string, ctx.numCols)
			hctx.location = tw.LocationEnd
			ctx.logger.Debugf("Warning: Cannot determine context for bottom border")
		}
		resp := t.buildCellContexts(ctx, mctx, hctx, colAligns, colPadding)
		ctx.logger.Debugf("Bottom border: Using Widths=%v", ctx.widths[tw.Row])
		f.Line(tw.Formatting{
			Row: tw.RowContext{
				Widths:       ctx.widths[tw.Row],
				Current:      resp.cells,
				Previous:     resp.prevCells,
				Position:     tw.Footer,
				Location:     tw.LocationEnd,
				ColMaxWidths: t.getColMaxWidths(tw.Footer),
			},
			Level:    tw.LevelFooter,
			IsSubRow: false,
		})
	}

	return nil
}

// renderHeader renders the table's header section with borders and padding.
// Parameters ctx and mctx hold rendering and merge state.
// Returns an error if rendering fails.
func (t *Table) renderHeader(ctx *renderContext, mctx *mergeContext) error {
	if len(ctx.headerLines) == 0 {
		return nil
	}
	ctx.logger.Debug("Rendering header section")

	f := ctx.renderer
	cfg := ctx.cfg
	colAligns := t.buildAligns(t.config.Header)
	colPadding := t.buildPadding(t.config.Header.Padding)
	hctx := &helperContext{position: tw.Header}

	if cfg.Borders.Top.Enabled() && cfg.Settings.Lines.ShowTop.Enabled() {
		ctx.logger.Debug("Rendering table top border")
		nextCells := make(map[int]tw.CellContext)
		if len(ctx.headerLines) > 0 {
			for j, cell := range ctx.headerLines[0] {
				nextCells[j] = tw.CellContext{Data: cell, Merge: mctx.headerMerges[j]}
			}
		}
		f.Line(tw.Formatting{
			Row: tw.RowContext{
				Widths:   ctx.widths[tw.Header],
				Next:     nextCells,
				Position: tw.Header,
				Location: tw.LocationFirst,
			},
			Level:    tw.LevelHeader,
			IsSubRow: false,
		})
	}

	if t.config.Header.Padding.Global.Top != tw.Empty {
		hctx.location = tw.LocationFirst
		hctx.line = t.buildPaddingLineContents(t.config.Header.Padding.Global.Top, ctx.widths[tw.Header], ctx.numCols, mctx.headerMerges)
		if err := t.renderPadding(ctx, mctx, hctx, t.config.Header.Padding.Global.Top); err != nil {
			return err
		}
	}

	for i, line := range ctx.headerLines {
		hctx.rowIdx = 0
		hctx.lineIdx = i
		hctx.line = padLine(line, ctx.numCols)
		hctx.location = t.determineLocation(i, len(ctx.headerLines), t.config.Header.Padding.Global.Top, t.config.Header.Padding.Global.Bottom)

		if t.config.Header.Callbacks.Global != nil {
			ctx.logger.Debug("Executing global header callback for line %d", i)
			t.config.Header.Callbacks.Global()
		}
		for colIdx, cb := range t.config.Header.Callbacks.PerColumn {
			if colIdx < ctx.numCols && cb != nil {
				ctx.logger.Debug("Executing per-column header callback for line %d, col %d", i, colIdx)
				cb()
			}
		}

		if err := t.renderLine(ctx, mctx, hctx, colAligns, colPadding); err != nil {
			return err
		}
	}

	if t.config.Header.Padding.Global.Bottom != tw.Empty {
		hctx.location = tw.LocationEnd
		hctx.line = t.buildPaddingLineContents(t.config.Header.Padding.Global.Bottom, ctx.widths[tw.Header], ctx.numCols, mctx.headerMerges)
		if err := t.renderPadding(ctx, mctx, hctx, t.config.Header.Padding.Global.Bottom); err != nil {
			return err
		}
	}

	if cfg.Settings.Lines.ShowHeaderLine.Enabled() && (len(ctx.rowLines) > 0 || len(ctx.footerLines) > 0) {
		ctx.logger.Debug("Rendering header separator line")
		resp := t.buildCellContexts(ctx, mctx, hctx, colAligns, colPadding)

		var nextSectionCells map[int]tw.CellContext
		var nextSectionWidths tw.Mapper[int, int]

		if len(ctx.rowLines) > 0 {
			nextSectionWidths = ctx.widths[tw.Row]
			rowColAligns := t.buildAligns(t.config.Row)
			rowColPadding := t.buildPadding(t.config.Row.Padding)
			firstRowHctx := &helperContext{
				position: tw.Row,
				rowIdx:   0,
				lineIdx:  0,
			}
			if len(ctx.rowLines[0]) > 0 {
				firstRowHctx.line = padLine(ctx.rowLines[0][0], ctx.numCols)
			} else {
				firstRowHctx.line = make([]string, ctx.numCols)
			}
			firstRowResp := t.buildCellContexts(ctx, mctx, firstRowHctx, rowColAligns, rowColPadding)
			nextSectionCells = firstRowResp.cells
		} else if len(ctx.footerLines) > 0 {
			nextSectionWidths = ctx.widths[tw.Row]
			footerColAligns := t.buildAligns(t.config.Footer)
			footerColPadding := t.buildPadding(t.config.Footer.Padding)
			firstFooterHctx := &helperContext{
				position: tw.Footer,
				rowIdx:   0,
				lineIdx:  0,
			}
			if len(ctx.footerLines) > 0 {
				firstFooterHctx.line = padLine(ctx.footerLines[0], ctx.numCols)
			} else {
				firstFooterHctx.line = make([]string, ctx.numCols)
			}
			firstFooterResp := t.buildCellContexts(ctx, mctx, firstFooterHctx, footerColAligns, footerColPadding)
			nextSectionCells = firstFooterResp.cells
		} else {
			nextSectionWidths = ctx.widths[tw.Header]
			nextSectionCells = nil
		}

		f.Line(tw.Formatting{
			Row: tw.RowContext{
				Widths:   nextSectionWidths,
				Current:  resp.cells,
				Previous: resp.prevCells,
				Next:     nextSectionCells,
				Position: tw.Header,
				Location: tw.LocationMiddle,
			},
			Level:    tw.LevelBody,
			IsSubRow: false,
		})
	}
	return nil
}

// renderLine renders a single line with callbacks and normalized widths.
// Parameters include ctx, mctx, hctx, aligns, and padding for rendering.
// Returns an error if rendering fails.
func (t *Table) renderLine(ctx *renderContext, mctx *mergeContext, hctx *helperContext, aligns map[int]tw.Align, padding map[int]tw.Padding) error {
	resp := t.buildCellContexts(ctx, mctx, hctx, aligns, padding)
	f := ctx.renderer

	isPaddingLine := false
	sectionConfig := t.config.Row
	switch hctx.position {
	case tw.Header:
		sectionConfig = t.config.Header
		isPaddingLine = (hctx.lineIdx == -1 && sectionConfig.Padding.Global.Top != tw.Empty) ||
			(hctx.lineIdx == len(ctx.headerLines) && sectionConfig.Padding.Global.Bottom != tw.Empty)
	case tw.Footer:
		sectionConfig = t.config.Footer
		isPaddingLine = (hctx.lineIdx == -1 && sectionConfig.Padding.Global.Top != tw.Empty) ||
			(hctx.lineIdx == len(ctx.footerLines) && (sectionConfig.Padding.Global.Bottom != tw.Empty || t.hasPerColumnBottomPadding()))
	case tw.Row:
		if hctx.rowIdx >= 0 && hctx.rowIdx < len(ctx.rowLines) {
			isPaddingLine = (hctx.lineIdx == -1 && sectionConfig.Padding.Global.Top != tw.Empty) ||
				(hctx.lineIdx == len(ctx.rowLines[hctx.rowIdx]) && sectionConfig.Padding.Global.Bottom != tw.Empty)
		}
	}

	sectionWidths := ctx.widths[hctx.position]
	normalizedWidths := ctx.widths[tw.Row]

	formatting := tw.Formatting{
		Row: tw.RowContext{
			Widths:       sectionWidths,
			ColMaxWidths: t.getColMaxWidths(hctx.position),
			Current:      resp.cells,
			Previous:     resp.prevCells,
			Next:         resp.nextCells,
			Position:     hctx.position,
			Location:     hctx.location,
		},
		Level:            t.getLevel(hctx.position),
		IsSubRow:         hctx.lineIdx > 0 || isPaddingLine,
		NormalizedWidths: normalizedWidths,
	}

	if hctx.position == tw.Row {
		formatting.HasFooter = len(ctx.footerLines) > 0
	}

	switch hctx.position {
	case tw.Header:
		f.Header([][]string{hctx.line}, formatting)
	case tw.Row:
		f.Row(hctx.line, formatting)
	case tw.Footer:
		f.Footer([][]string{hctx.line}, formatting)
	}
	return nil
}

// renderPadding renders padding lines for a section.
// Parameters include ctx, mctx, hctx, and padChar for padding content.
// Returns an error if rendering fails.
func (t *Table) renderPadding(ctx *renderContext, mctx *mergeContext, hctx *helperContext, padChar string) error {
	ctx.logger.Debug("Rendering padding line for %s (using char like '%s')", hctx.position, padChar)

	colAligns := t.buildAligns(t.config.Row)
	colPadding := t.buildPadding(t.config.Row.Padding)

	switch hctx.position {
	case tw.Header:
		colAligns = t.buildAligns(t.config.Header)
		colPadding = t.buildPadding(t.config.Header.Padding)
	case tw.Footer:
		colAligns = t.buildAligns(t.config.Footer)
		colPadding = t.buildPadding(t.config.Footer.Padding)
	}

	return t.renderLine(ctx, mctx, hctx, colAligns, colPadding)
}

// renderRow renders the table's row section with borders and padding.
// Parameters ctx and mctx hold rendering and merge state.
// Returns an error if rendering fails.
func (t *Table) renderRow(ctx *renderContext, mctx *mergeContext) error {
	if len(ctx.rowLines) == 0 {
		return nil
	}
	ctx.logger.Debugf("Rendering row section (total rows: %d)", len(ctx.rowLines))

	f := ctx.renderer
	cfg := ctx.cfg
	colAligns := t.buildAligns(t.config.Row)
	colPadding := t.buildPadding(t.config.Row.Padding)
	hctx := &helperContext{position: tw.Row}

	footerIsEmptyOrNonExistent := !t.hasFooterElements()
	if len(ctx.headerLines) == 0 && footerIsEmptyOrNonExistent && cfg.Borders.Top.Enabled() && cfg.Settings.Lines.ShowTop.Enabled() {
		ctx.logger.Debug("Rendering table top border (rows only table)")
		nextCells := make(map[int]tw.CellContext)
		if len(ctx.rowLines) > 0 && len(ctx.rowLines[0]) > 0 && len(mctx.rowMerges) > 0 {
			firstLine := ctx.rowLines[0][0]
			firstMerges := mctx.rowMerges[0]
			for j, cell := range padLine(firstLine, ctx.numCols) {
				mergeState := tw.MergeState{}
				if firstMerges != nil {
					mergeState = firstMerges[j]
				}
				nextCells[j] = tw.CellContext{Data: cell, Merge: mergeState, Width: ctx.widths[tw.Row].Get(j)}
			}
		}
		f.Line(tw.Formatting{
			Row: tw.RowContext{
				Widths:   ctx.widths[tw.Row],
				Next:     nextCells,
				Position: tw.Row,
				Location: tw.LocationFirst,
			},
			Level:    tw.LevelHeader,
			IsSubRow: false,
		})
	}

	for i, lines := range ctx.rowLines {
		rowHasTopPadding := t.config.Row.Padding.Global.Top != tw.Empty
		if rowHasTopPadding {
			hctx.rowIdx = i
			hctx.lineIdx = -1
			if i == 0 && len(ctx.headerLines) == 0 {
				hctx.location = tw.LocationFirst
			} else {
				hctx.location = tw.LocationMiddle
			}
			hctx.line = t.buildPaddingLineContents(t.config.Row.Padding.Global.Top, ctx.widths[tw.Row], ctx.numCols, mctx.rowMerges[i])
			ctx.logger.Debug("Calling renderPadding for Row Top Padding (row %d): %v (loc: %v)", i, hctx.line, hctx.location)
			if err := t.renderPadding(ctx, mctx, hctx, t.config.Row.Padding.Global.Top); err != nil {
				return err
			}
		}

		footerExists := t.hasFooterElements()
		rowHasBottomPadding := t.config.Row.Padding.Global.Bottom != tw.Empty
		isLastRow := i == len(ctx.rowLines)-1

		for j, visualLineData := range lines {
			hctx.rowIdx = i
			hctx.lineIdx = j
			hctx.line = padLine(visualLineData, ctx.numCols)

			if t.config.Behavior.TrimLine.Enabled() {
				if j > 0 {
					visualLineHasActualContent := false
					for kCellIdx, cellContentInVisualLine := range hctx.line {
						if t.Trimmer(cellContentInVisualLine) != "" {
							visualLineHasActualContent = true
							ctx.logger.Debug("Visual line [%d][%d] has content in cell %d: '%s'. Not skipping.", i, j, kCellIdx, cellContentInVisualLine)
							break
						}
					}

					if !visualLineHasActualContent {
						ctx.logger.Debug("Skipping visual line [%d][%d] as it's entirely blank after trimming. Line: %q", i, j, hctx.line)
						continue
					}
				}
			}

			isFirstRow := i == 0
			isLastLineOfRow := j == len(lines)-1

			if isFirstRow && j == 0 && !rowHasTopPadding && len(ctx.headerLines) == 0 {
				hctx.location = tw.LocationFirst
			} else if isLastRow && isLastLineOfRow && !rowHasBottomPadding && !footerExists {
				hctx.location = tw.LocationEnd
			} else {
				hctx.location = tw.LocationMiddle
			}

			ctx.logger.Debugf("Rendering row %d line %d with location %v. Content: %q", i, j, hctx.location, hctx.line)
			if err := t.renderLine(ctx, mctx, hctx, colAligns, colPadding); err != nil {
				return err
			}
		}

		if rowHasBottomPadding {
			hctx.rowIdx = i
			hctx.lineIdx = len(lines)
			if isLastRow && !footerExists {
				hctx.location = tw.LocationEnd
			} else {
				hctx.location = tw.LocationMiddle
			}
			hctx.line = t.buildPaddingLineContents(t.config.Row.Padding.Global.Bottom, ctx.widths[tw.Row], ctx.numCols, mctx.rowMerges[i])
			ctx.logger.Debug("Calling renderPadding for Row Bottom Padding (row %d): %v (loc: %v)", i, hctx.line, hctx.location)
			if err := t.renderPadding(ctx, mctx, hctx, t.config.Row.Padding.Global.Bottom); err != nil {
				return err
			}
		}

		if cfg.Settings.Separators.BetweenRows.Enabled() && !isLastRow {
			ctx.logger.Debug("Rendering between-rows separator after logical row %d", i)
			respCurrent := t.buildCellContexts(ctx, mctx, hctx, colAligns, colPadding)

			var nextCellsForSeparator map[int]tw.CellContext = nil
			nextRowIdx := i + 1
			if nextRowIdx < len(ctx.rowLines) && nextRowIdx < len(mctx.rowMerges) {
				hctxNext := &helperContext{position: tw.Row, rowIdx: nextRowIdx, location: tw.LocationMiddle}
				nextRowActualLines := ctx.rowLines[nextRowIdx]
				nextRowMerges := mctx.rowMerges[nextRowIdx]

				if t.config.Row.Padding.Global.Top != tw.Empty {
					hctxNext.lineIdx = -1
					hctxNext.line = t.buildPaddingLineContents(t.config.Row.Padding.Global.Top, ctx.widths[tw.Row], ctx.numCols, nextRowMerges)
				} else if len(nextRowActualLines) > 0 {
					hctxNext.lineIdx = 0
					hctxNext.line = padLine(nextRowActualLines[0], ctx.numCols)
				} else {
					hctxNext.lineIdx = 0
					hctxNext.line = make([]string, ctx.numCols)
				}
				respNext := t.buildCellContexts(ctx, mctx, hctxNext, colAligns, colPadding)
				nextCellsForSeparator = respNext.cells
			} else {
				ctx.logger.Debug("Separator context: No next logical row for separator after row %d.", i)
			}

			f.Line(tw.Formatting{
				Row: tw.RowContext{
					Widths:       ctx.widths[tw.Row],
					Current:      respCurrent.cells,
					Previous:     respCurrent.prevCells,
					Next:         nextCellsForSeparator,
					Position:     tw.Row,
					Location:     tw.LocationMiddle,
					ColMaxWidths: t.getColMaxWidths(tw.Row),
				},
				Level:     tw.LevelBody,
				IsSubRow:  false,
				HasFooter: footerExists,
			})
		}
	}
	return nil
}
