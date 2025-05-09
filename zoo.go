package tablewriter

import (
	"database/sql"
	"fmt"
	"github.com/olekukonko/errors"
	"github.com/olekukonko/tablewriter/pkg/twwarp"
	"github.com/olekukonko/tablewriter/tw"
	"github.com/olekukonko/tablewriter/twfn"
	"io"
	"reflect"
	"strings"
)

// prepareContent processes cell content with formatting and wrapping.
// Parameters include cells to process and config for formatting rules.
// Returns a slice of string slices representing processed lines.
func (t *Table) prepareContent(cells []string, config tw.CellConfig) [][]string {
	isStreaming := t.config.Stream.Enable && t.hasPrinted
	t.logger.Debug("prepareContent: Processing cells=%v (streaming: %v)", cells, isStreaming)
	initialInputCellCount := len(cells)
	result := make([][]string, 0)

	effectiveNumCols := initialInputCellCount
	if isStreaming {
		if t.streamNumCols > 0 {
			effectiveNumCols = t.streamNumCols
			t.logger.Debug("prepareContent: Streaming mode, using fixed streamNumCols: %d", effectiveNumCols)
			if len(cells) != effectiveNumCols {
				t.logger.Warn("prepareContent: Streaming mode, input cell count (%d) does not match streamNumCols (%d). Input cells will be padded/truncated.", len(cells), effectiveNumCols)
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
			t.logger.Warn("prepareContent: Streaming mode enabled but streamNumCols is 0. Using input cell count %d. Stream widths may not be available.", effectiveNumCols)
		}
	}

	for i := 0; i < effectiveNumCols; i++ {
		cellContent := ""
		if i < len(cells) {
			cellContent = cells[i]
		} else {
			cellContent = tw.Empty
		}

		colPad := config.Padding.Global
		if i < len(config.Padding.PerColumn) && config.Padding.PerColumn[i] != (tw.Padding{}) {
			colPad = config.Padding.PerColumn[i]
		}
		padLeftWidth := twfn.DisplayWidth(colPad.Left)
		padRightWidth := twfn.DisplayWidth(colPad.Right)

		effectiveContentMaxWidth := t.calculateContentMaxWidth(i, config, padLeftWidth, padRightWidth, isStreaming)

		if config.Formatting.AutoFormat {
			cellContent = twfn.Title(strings.Join(twfn.SplitCamelCase(cellContent), tw.Space))
		}

		lines := strings.Split(cellContent, "\n")
		finalLinesForCell := make([]string, 0)
		for _, line := range lines {
			if effectiveContentMaxWidth > 0 {
				switch config.Formatting.AutoWrap {
				case tw.WrapNormal:
					wrapped, _ := twwarp.WrapString(line, effectiveContentMaxWidth)
					finalLinesForCell = append(finalLinesForCell, wrapped...)
				case tw.WrapTruncate:
					if twfn.DisplayWidth(line) > effectiveContentMaxWidth {
						ellipsisWidth := twfn.DisplayWidth(tw.CharEllipsis)
						if effectiveContentMaxWidth >= ellipsisWidth {
							finalLinesForCell = append(finalLinesForCell, twfn.TruncateString(line, effectiveContentMaxWidth-ellipsisWidth, tw.CharEllipsis))
						} else {
							finalLinesForCell = append(finalLinesForCell, twfn.TruncateString(line, effectiveContentMaxWidth, ""))
						}
					} else {
						finalLinesForCell = append(finalLinesForCell, line)
					}
				case tw.WrapBreak:
					wrapped := make([]string, 0)
					currentLine := line
					for twfn.DisplayWidth(currentLine) > effectiveContentMaxWidth {
						breakPoint := twfn.BreakPoint(currentLine, effectiveContentMaxWidth)
						if breakPoint <= 0 {
							t.logger.Warn("prepareContent: WrapBreak - BreakPoint <= 0 for line '%s' at width %d. Attempting manual break.", currentLine, effectiveContentMaxWidth)
							runes := []rune(currentLine)
							actualBreakRuneCount := 0
							tempWidth := 0
							for charIdx, r := range currentLine {
								runeStr := string(r)
								rw := twfn.DisplayWidth(runeStr)
								if tempWidth+rw > effectiveContentMaxWidth && charIdx > 0 {
									break
								}
								tempWidth += rw
								actualBreakRuneCount = charIdx + 1
								if tempWidth >= effectiveContentMaxWidth && charIdx == 0 {
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
								if twfn.DisplayWidth(currentLine) > 0 {
									wrapped = append(wrapped, currentLine)
									currentLine = ""
								}
								break
							}
						} else {
							runes := []rune(currentLine)
							if breakPoint <= len(runes) {
								wrapped = append(wrapped, string(runes[:breakPoint])+tw.CharBreak)
								currentLine = string(runes[breakPoint:])
							} else {
								t.logger.Warn("prepareContent: WrapBreak - BreakPoint (%d) out of bounds for line runes (%d). Adding full line.", breakPoint, len(runes))
								wrapped = append(wrapped, currentLine)
								currentLine = ""
								break
							}
						}
					}
					if twfn.DisplayWidth(currentLine) > 0 {
						wrapped = append(wrapped, currentLine)
					}
					if len(wrapped) == 0 && twfn.DisplayWidth(line) > 0 && len(finalLinesForCell) == 0 {
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
				t.logger.Warn("prepareContent: Column index %d out of bounds (%d) during result matrix population.", i, len(result[j]))
			}
		}
	}

	t.logger.Debug("prepareContent: Content prepared, result %d lines.", len(result))
	return result
}

// prepareContexts initializes rendering and merge contexts.
// No parameters are required.
// Returns renderContext, mergeContext, and an error if initialization fails.
func (t *Table) prepareContexts() (*renderContext, *mergeContext, error) {
	numOriginalCols := t.maxColumns()
	t.logger.Debug("prepareContexts: Original number of columns: %d", numOriginalCols)

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

	isEmpty, visibleCount := t.getEmptyColumnInfo(numOriginalCols)
	ctx.emptyColumns = isEmpty
	ctx.visibleColCount = visibleCount

	mctx := &mergeContext{
		headerMerges: make(map[int]tw.MergeState),
		rowMerges:    make([]map[int]tw.MergeState, len(t.rows)),
		footerMerges: make(map[int]tw.MergeState),
		horzMerges:   make(map[tw.Position]map[int]bool),
	}
	for i := range mctx.rowMerges {
		mctx.rowMerges[i] = make(map[int]tw.MergeState)
	}

	ctx.headerLines = t.headers
	ctx.rowLines = t.rows
	ctx.footerLines = t.footers

	if err := t.calculateAndNormalizeWidths(ctx); err != nil {
		t.logger.Debug("Error during initial width calculation: %v", err)
		return nil, nil, err
	}
	t.logger.Debug("Initial normalized widths (before hiding): H=%v, R=%v, F=%v",
		ctx.widths[tw.Header], ctx.widths[tw.Row], ctx.widths[tw.Footer])

	preparedHeaderLines, headerMerges, _ := t.prepareWithMerges(ctx.headerLines, t.config.Header, tw.Header)
	ctx.headerLines = preparedHeaderLines
	mctx.headerMerges = headerMerges

	processedRowLines := make([][][]string, len(ctx.rowLines))
	for i, row := range ctx.rowLines {
		if mctx.rowMerges[i] == nil {
			mctx.rowMerges[i] = make(map[int]tw.MergeState)
		}
		processedRowLines[i], mctx.rowMerges[i], _ = t.prepareWithMerges(row, t.config.Row, tw.Row)
	}
	ctx.rowLines = processedRowLines

	t.applyHorizontalMergeWidths(tw.Header, ctx, mctx.headerMerges)

	if t.config.Row.Formatting.MergeMode&tw.MergeVertical != 0 {
		t.applyVerticalMerges(ctx, mctx)
	}
	if t.config.Row.Formatting.MergeMode&tw.MergeHierarchical != 0 {
		t.applyHierarchicalMerges(ctx, mctx)
	}

	t.prepareFooter(ctx, mctx)
	t.logger.Debug("Footer prepared. Widths before hiding: H=%v, R=%v, F=%v",
		ctx.widths[tw.Header], ctx.widths[tw.Row], ctx.widths[tw.Footer])

	if t.config.AutoHide {
		t.logger.Debug("Applying AutoHide: Adjusting widths for empty columns.")
		if ctx.emptyColumns == nil {
			t.logger.Debug("Warning: ctx.emptyColumns is nil during width adjustment.")
		} else if len(ctx.emptyColumns) != ctx.numCols {
			t.logger.Debug("Warning: Length mismatch between emptyColumns (%d) and numCols (%d). Skipping adjustment.", len(ctx.emptyColumns), ctx.numCols)
		} else {
			for colIdx := 0; colIdx < ctx.numCols; colIdx++ {
				if ctx.emptyColumns[colIdx] {
					t.logger.Debug("AutoHide: Hiding column %d by setting width to 0.", colIdx)
					ctx.widths[tw.Header].Set(colIdx, 0)
					ctx.widths[tw.Row].Set(colIdx, 0)
					ctx.widths[tw.Footer].Set(colIdx, 0)
				}
			}
			t.logger.Debug("Widths after AutoHide adjustment: H=%v, R=%v, F=%v",
				ctx.widths[tw.Header], ctx.widths[tw.Row], ctx.widths[tw.Footer])
		}
	} else {
		t.logger.Debug("AutoHide is disabled, skipping width adjustment.")
	}
	t.logger.Debug("prepareContexts completed all stages.")
	return ctx, mctx, nil
}

// prepareFooter processes footer content and applies merges.
// Parameters ctx and mctx hold rendering and merge state.
// No return value.
func (t *Table) prepareFooter(ctx *renderContext, mctx *mergeContext) {
	if len(t.footers) == 0 {
		ctx.logger.Debug("Skipping footer preparation - no footer data")
		if ctx.widths[tw.Footer] == nil {
			ctx.widths[tw.Footer] = tw.NewMapper[int, int]()
		}
		numCols := ctx.numCols
		for i := 0; i < numCols; i++ {
			ctx.widths[tw.Footer].Set(i, ctx.widths[tw.Row].Get(i))
		}
		t.logger.Debug("Initialized empty footer widths based on row widths: %v", ctx.widths[tw.Footer])
		ctx.footerPrepared = true
		return
	}

	t.logger.Debug("Preparing footer with merge mode: %d", t.config.Footer.Formatting.MergeMode)
	preparedLines, mergeStates, _ := t.prepareWithMerges(t.footers, t.config.Footer, tw.Footer)
	t.footers = preparedLines
	mctx.footerMerges = mergeStates
	ctx.footerLines = t.footers
	t.logger.Debug("Base footer widths (normalized from rows/header): %v", ctx.widths[tw.Footer])
	t.applyHorizontalMergeWidths(tw.Footer, ctx, mctx.footerMerges)
	ctx.footerPrepared = true
	t.logger.Debug("Footer preparation completed. Final footer widths: %v", ctx.widths[tw.Footer])
}

// prepareWithMerges processes content and detects horizontal merges.
// Parameters include content, config, and position (Header, Row, Footer).
// Returns processed lines, merge states, and horizontal merge map.
func (t *Table) prepareWithMerges(content [][]string, config tw.CellConfig, position tw.Position) ([][]string, map[int]tw.MergeState, map[int]bool) {
	t.logger.Debug("PrepareWithMerges START: position=%s, mergeMode=%d", position, config.Formatting.MergeMode)
	if len(content) == 0 {
		t.logger.Debug("PrepareWithMerges END: No content.")
		return content, nil, nil
	}

	numCols := 0
	if len(content) > 0 && len(content[0]) > 0 {
		numCols = len(content[0])
	} else {
		for _, line := range content {
			if len(line) > numCols {
				numCols = len(line)
			}
		}
		if numCols == 0 {
			numCols = t.maxColumns()
		}
	}

	if numCols == 0 {
		t.logger.Debug("PrepareWithMerges END: numCols is zero.")
		return content, nil, nil
	}

	horzMergeMap := make(map[int]bool)
	mergeMap := make(map[int]tw.MergeState)
	result := make([][]string, len(content))
	for i := range content {
		result[i] = padLine(content[i], numCols)
	}

	if config.Formatting.MergeMode&tw.MergeHorizontal != 0 {
		t.logger.Debug("Checking for horizontal merges in %d lines", len(content))

		if position == tw.Footer {
			for lineIdx := 0; lineIdx < len(content); lineIdx++ {
				originalLine := padLine(content[lineIdx], numCols)
				currentLineResult := result[lineIdx]

				firstContentIdx := -1
				var firstContent string
				for c := 0; c < numCols; c++ {
					if c >= len(originalLine) {
						break
					}
					trimmedVal := strings.TrimSpace(originalLine[c])
					if trimmedVal != "" && trimmedVal != "-" {
						firstContentIdx = c
						firstContent = originalLine[c]
						break
					} else if trimmedVal == "-" {
						break
					}
				}

				if firstContentIdx > 0 {
					span := firstContentIdx + 1
					startCol := 0

					allEmptyBefore := true
					for c := 0; c < firstContentIdx; c++ {
						if c >= len(originalLine) || strings.TrimSpace(originalLine[c]) != "" {
							allEmptyBefore = false
							break
						}
					}

					if allEmptyBefore {
						t.logger.Debug("Footer lead-merge applied line %d: content '%s' from col %d moved to col %d, span %d",
							lineIdx, firstContent, firstContentIdx, startCol, span)

						if startCol < len(currentLineResult) {
							currentLineResult[startCol] = firstContent
						}
						for k := startCol + 1; k < startCol+span; k++ {
							if k < len(currentLineResult) {
								currentLineResult[k] = tw.Empty
							}
						}

						startState := mergeMap[startCol]
						startState.Horizontal = tw.MergeStateOption{Present: true, Span: span, Start: true, End: span == 1}
						mergeMap[startCol] = startState
						horzMergeMap[startCol] = true

						for k := startCol + 1; k < startCol+span; k++ {
							colState := mergeMap[k]
							colState.Horizontal = tw.MergeStateOption{Present: true, Span: span, Start: false, End: k == startCol+span-1}
							mergeMap[k] = colState
							horzMergeMap[k] = true
						}
					}
				}
			}
		}

		for lineIdx := 0; lineIdx < len(content); lineIdx++ {
			originalLine := padLine(content[lineIdx], numCols)
			currentLineResult := result[lineIdx]
			col := 0
			for col < numCols {
				if horzMergeMap[col] {
					leadMergeHandled := false
					if mergeState, ok := mergeMap[col]; ok && mergeState.Horizontal.Present && !mergeState.Horizontal.Start {
						tempCol := col - 1
						startCol := -1
						startState := tw.MergeState{}
						for tempCol >= 0 {
							if state, okS := mergeMap[tempCol]; okS && state.Horizontal.Present && state.Horizontal.Start {
								startCol = tempCol
								startState = state
								break
							}
							tempCol--
						}
						if startCol != -1 {
							skipToCol := startCol + startState.Horizontal.Span
							if skipToCol > col {
								t.logger.Debug("Skipping standard H-merge check from col %d to %d (part of detected H-merge)", col, skipToCol-1)
								col = skipToCol
								leadMergeHandled = true
							}
						}
					}
					if leadMergeHandled {
						continue
					}
				}

				if col >= len(currentLineResult) {
					break
				}
				currentVal := strings.TrimSpace(currentLineResult[col])

				if currentVal == "" || currentVal == "-" || (mergeMap[col].Horizontal.Present && mergeMap[col].Horizontal.Start) {
					col++
					continue
				}

				span := 1
				startCol := col
				for nextCol := col + 1; nextCol < numCols; nextCol++ {
					if nextCol >= len(originalLine) {
						break
					}
					originalNextVal := strings.TrimSpace(originalLine[nextCol])

					if currentVal == originalNextVal && !horzMergeMap[nextCol] && originalNextVal != "-" {
						span++
					} else {
						break
					}
				}

				if span > 1 {
					t.logger.Debug("Standard horizontal merge at line %d, col %d, span %d", lineIdx, startCol, span)
					startState := mergeMap[startCol]
					startState.Horizontal = tw.MergeStateOption{Present: true, Span: span, Start: true, End: (span == 1)}
					mergeMap[startCol] = startState
					horzMergeMap[startCol] = true

					for k := startCol + 1; k < startCol+span; k++ {
						if k < len(currentLineResult) {
							currentLineResult[k] = tw.Empty
						}
						colState := mergeMap[k]
						colState.Horizontal = tw.MergeStateOption{Present: true, Span: span, Start: false, End: k == startCol+span-1}
						mergeMap[k] = colState
						horzMergeMap[k] = true
					}
					col += span
				} else {
					col++
				}
			}
		}
	}

	t.logger.Debug("PrepareWithMerges END: position=%s, lines=%d", position, len(result))
	return result, mergeMap, horzMergeMap
}

// rawCellsToStrings converts a row to its raw string representation using specified cell config for filters.
// 'rowInput' can be []string, []any, or a custom type if t.stringer is set.
func (t *Table) rawCellsToStrings(rowInput interface{}, cellCfg tw.CellConfig) ([]string, error) {
	t.logger.Debug("rawCellsToStrings: Converting row: %v (type: %T) using filters from specified CellConfig", rowInput, rowInput)
	var cells []string
	var conversionError error

	switch v := rowInput.(type) {
	case []string:
		cells = v
		t.logger.Debug("rawCellsToStrings: Input is []string.")
	case []any:
		cells = make([]string, len(v))
		for i, element := range v {
			cells[i] = t.convertToString(element)
		}
		t.logger.Debug("rawCellsToStrings: Input is []any, processed elements.")
	default:
		// var ok bool
		if t.stringer != nil {
			cells, conversionError = t.callStringer(rowInput)
			if conversionError != nil {
				t.logger.Debug("rawCellsToStrings: Stringer error: %v", conversionError)
				return nil, conversionError
			}
			t.logger.Debug("rawCellsToStrings: Input (custom type) converted to: %v", cells)
		} else if stringer, ok := rowInput.(fmt.Stringer); ok {
			cells = []string{stringer.String()}
			t.logger.Debug("rawCellsToStrings: Input is fmt.Stringer, used String() for single cell")
		} else {
			conversionError = errors.Newf("cannot convert row type %T to []string; not []string, []any, no t.stringer, not fmt.Stringer", rowInput)
			t.logger.Debug("rawCellsToStrings: Conversion error: %v", conversionError)
			return nil, conversionError
		}
	}

	if cellCfg.Filter.Global != nil {
		cells = cellCfg.Filter.Global(cells)
		t.logger.Debug("rawCellsToStrings: Applied global filter. Result  Result: %v", cells)
	}

	if len(cellCfg.Filter.PerColumn) > 0 {
		originalCellsForLog := append([]string(nil), cells...)
		changedByFilter := false
		for i := 0; i < len(cells) && i < len(cellCfg.Filter.PerColumn); i++ {
			if filter := cellCfg.Filter.PerColumn[i]; filter != nil {
				newCell := filter(cells[i])
				if newCell != cells[i] {
					changedByFilter = true
				}
				cells[i] = newCell
			}
		}
		if changedByFilter {
			t.logger.Debug("rawCellsToStrings: Applied per-column filters. Original: %v, Result: %v", originalCellsForLog, cells)
		}
	}
	return cells, nil
}

// toStringLines converts raw cells to formatted lines for table output
func (t *Table) toStringLines(row interface{}, config tw.CellConfig) ([][]string, error) {
	cells, err := t.rawCellsToStrings(row, config)
	if err != nil {
		return nil, err
	}
	return t.prepareContent(cells, config), nil
}

// processVariadicElements handles the common logic for processing variadic arguments
// that could be either individual elements or a slice of elements
func (t *Table) processVariadicElements(elements []any) []any {
	// Check if the input looks like a single slice was passed
	if len(elements) == 1 {
		firstArg := elements[0]
		// Try to assert to common slice types users might pass
		switch v := firstArg.(type) {
		case []string:
			t.logger.Debug("Detected single []string argument. Unpacking it.")
			result := make([]any, len(v))
			for i, val := range v {
				result[i] = val
			}
			return result
		case []interface{}:
			t.logger.Debug("Detected single []interface{} argument. Unpacking it.")
			result := make([]any, len(v))
			for i, val := range v {
				result[i] = val
			}
			return result
		}
	}

	// Either multiple arguments were passed, or a single non-slice argument
	t.logger.Debug("Input has multiple elements, is empty, or is a single non-slice element. Using variadic elements as is.")
	return elements
}

// prepareTableSection prepares either headers or footers for the table
func (t *Table) prepareTableSection(elements []any, config tw.CellConfig, sectionName string) [][]string {
	actualCellsToProcess := t.processVariadicElements(elements)
	t.logger.Debug("%s(): Effective cells to process: %v", sectionName, actualCellsToProcess)

	stringsResult, err := t.rawCellsToStrings(actualCellsToProcess, config)
	if err != nil {
		t.logger.Error("%s(): Failed to convert elements to strings: %v", sectionName, err)
		stringsResult = []string{}
	}

	prepared := t.prepareContent(stringsResult, config)
	numColsBatch := t.maxColumns()

	if len(prepared) > 0 {
		for i := range prepared {
			if len(prepared[i]) < numColsBatch {
				t.logger.Debug("Padding %s line %d from %d to %d columns", sectionName, i, len(prepared[i]), numColsBatch)
				paddedLine := make([]string, numColsBatch)
				copy(paddedLine, prepared[i])
				for j := len(prepared[i]); j < numColsBatch; j++ {
					paddedLine[j] = tw.Empty
				}
				prepared[i] = paddedLine
			} else if len(prepared[i]) > numColsBatch {
				t.logger.Debug("Truncating %s line %d from %d to %d columns", sectionName, i, len(prepared[i]), numColsBatch)
				prepared[i] = prepared[i][:numColsBatch]
			}
		}
	}

	return prepared
}

// convertToString converts a value to its string representation.
func (t *Table) convertToString(value interface{}) string {
	if value == nil {
		return ""
	}
	switch v := value.(type) {
	case tw.Formatter:
		return v.Format()
	case io.Reader:
		const maxReadSize = 512
		var buf strings.Builder
		_, err := io.CopyN(&buf, v, maxReadSize)
		if err != nil && err != io.EOF {
			return fmt.Sprintf("[reader error: %v]", err)
		}
		if buf.Len() == maxReadSize {
			buf.WriteString("...")
		}
		return buf.String()
	case sql.NullString:
		if v.Valid {
			return v.String
		}
		return ""
	case sql.NullInt64:
		if v.Valid {
			return fmt.Sprintf("%d", v.Int64)
		}
		return ""
	case sql.NullFloat64:
		if v.Valid {
			return fmt.Sprintf("%f", v.Float64)
		}
		return ""
	case sql.NullBool:
		if v.Valid {
			return fmt.Sprintf("%t", v.Bool)
		}
		return ""
	case sql.NullTime:
		if v.Valid {
			return v.Time.String()
		}
		return ""
	case []byte:
		return string(v)
	case error:
		return v.Error()
	case fmt.Stringer:
		return v.String()
	default:
		defer func() {
			if r := recover(); r != nil {
				t.logger.Debug("convertToString: Recovered panic for value %v: %v", value, r)
			}
		}()
		return fmt.Sprintf("%v", value)
	}
}

// callStringer invokes the table's stringer function with optional caching.
func (t *Table) callStringer(input interface{}) ([]string, error) {
	if !t.stringerCacheEnabled {
		// Fallback to original reflection logic
		rv := reflect.ValueOf(t.stringer)
		stringerType := rv.Type()
		inputValue := reflect.ValueOf(input)

		if !(rv.Kind() == reflect.Func && stringerType.NumIn() == 1 && stringerType.NumOut() == 1 &&
			inputValue.Type().AssignableTo(stringerType.In(0)) &&
			stringerType.Out(0) == reflect.TypeOf([]string{})) {
			return nil, errors.Newf("stringer must be func(T) []string where T is assignable from %T, got %T", input, t.stringer)
		}
		result := rv.Call([]reflect.Value{inputValue})
		cells, ok := result[0].Interface().([]string)
		if !ok {
			return nil, errors.Newf("stringer for type %T did not return []string", input)
		}
		return cells, nil
	}

	// Cached path
	inputType := reflect.TypeOf(input)
	t.stringerCacheMu.RLock()
	if rv, ok := t.stringerCache[inputType]; ok {
		t.stringerCacheMu.RUnlock()
		t.logger.Debug("callStringer: Cache hit for type %v", inputType)
		result := rv.Call([]reflect.Value{reflect.ValueOf(input)})
		cells, ok := result[0].Interface().([]string)
		if !ok {
			return nil, errors.Newf("cached stringer for type %T did not return []string", input)
		}
		return cells, nil
	}
	t.stringerCacheMu.RUnlock()

	t.logger.Debug("callStringer: Cache miss for type %v", inputType)
	rv := reflect.ValueOf(t.stringer)
	stringerType := rv.Type()
	if !(rv.Kind() == reflect.Func && stringerType.NumIn() == 1 && stringerType.NumOut() == 1 &&
		inputType.AssignableTo(stringerType.In(0)) &&
		stringerType.Out(0) == reflect.TypeOf([]string{})) {
		return nil, errors.Newf("stringer must be func(T) []string where T is assignable from %T, got %T", input, t.stringer)
	}

	t.stringerCacheMu.Lock()
	t.stringerCache[inputType] = rv
	t.stringerCacheMu.Unlock()

	result := rv.Call([]reflect.Value{reflect.ValueOf(input)})
	cells, ok := result[0].Interface().([]string)
	if !ok {
		return nil, errors.Newf("stringer for type %T did not return []string", input)
	}
	return cells, nil
}

// buildPaddingLineContents constructs a padding line for a given section, respecting column widths and horizontal merges.
// It generates a []string where each element is the padding content for a column, using the specified padChar.
func (t *Table) buildPaddingLineContents(padChar string, widths tw.Mapper[int, int], numCols int, merges map[int]tw.MergeState) []string {
	line := make([]string, numCols)
	padWidth := twfn.DisplayWidth(padChar)
	if padWidth < 1 {
		padWidth = 1
	}
	for j := 0; j < numCols; j++ {
		mergeState := tw.MergeState{}
		if merges != nil {
			if state, ok := merges[j]; ok {
				mergeState = state
			}
		}
		if mergeState.Horizontal.Present && !mergeState.Horizontal.Start {
			line[j] = tw.Empty
			continue
		}
		colWd := widths.Get(j)
		repeatCount := 0
		if colWd > 0 && padWidth > 0 {
			repeatCount = colWd / padWidth
		}
		if colWd > 0 && repeatCount < 1 {
			repeatCount = 1
		}
		content := strings.Repeat(padChar, repeatCount)
		line[j] = content
	}
	if t.logger.Enabled() {
		t.logger.Debug("Built padding line with char '%s' for %d columns", padChar, numCols)
	}
	return line
}

// calculateContentMaxWidth computes the maximum content width for a column, accounting for padding and mode-specific constraints.
// Returns the effective content width (after subtracting padding) for the given column index.
func (t *Table) calculateContentMaxWidth(colIdx int, config tw.CellConfig, padLeftWidth, padRightWidth int, isStreaming bool) int {
	var effectiveContentMaxWidth int
	if isStreaming {
		totalColumnWidthFromStream := t.streamWidths.Get(colIdx)
		if totalColumnWidthFromStream < 0 {
			totalColumnWidthFromStream = 0
		}
		effectiveContentMaxWidth = totalColumnWidthFromStream - padLeftWidth - padRightWidth
		if effectiveContentMaxWidth < 1 && totalColumnWidthFromStream > (padLeftWidth+padRightWidth) {
			effectiveContentMaxWidth = 1
		} else if effectiveContentMaxWidth < 0 {
			effectiveContentMaxWidth = 0
		}
		if totalColumnWidthFromStream == 0 {
			effectiveContentMaxWidth = 0
		}
		t.logger.Debug("calculateContentMaxWidth: Streaming col %d, TotalColWd=%d, PadL=%d, PadR=%d -> ContentMaxWd=%d",
			colIdx, totalColumnWidthFromStream, padLeftWidth, padRightWidth, effectiveContentMaxWidth)
	} else {
		hasConstraint := false
		constraintTotalCellWidth := 0
		if config.ColMaxWidths.PerColumn != nil {
			if colMax, ok := config.ColMaxWidths.PerColumn.OK(colIdx); ok && colMax > 0 {
				constraintTotalCellWidth = colMax
				hasConstraint = true
				t.logger.Debug("calculateContentMaxWidth: Batch col %d using config.ColMaxWidths.PerColumn (as total cell width constraint): %d", colIdx, constraintTotalCellWidth)
			}
		}
		if !hasConstraint && config.Formatting.MaxWidth > 0 {
			constraintTotalCellWidth = config.Formatting.MaxWidth
			hasConstraint = true
			t.logger.Debug("calculateContentMaxWidth: Batch col %d using config.Formatting.MaxWidth (as total cell width constraint): %d", colIdx, constraintTotalCellWidth)
		}
		if !hasConstraint && t.config.MaxWidth > 0 && config.Formatting.AutoWrap != tw.WrapNone {
			constraintTotalCellWidth = t.config.MaxWidth
			hasConstraint = true
			t.logger.Debug("calculateContentMaxWidth: Batch col %d using t.config.MaxWidth (as total cell width constraint, due to AutoWrap != WrapNone): %d", colIdx, constraintTotalCellWidth)
		}
		if hasConstraint {
			effectiveContentMaxWidth = constraintTotalCellWidth - padLeftWidth - padRightWidth
			if effectiveContentMaxWidth < 1 && constraintTotalCellWidth > (padLeftWidth+padRightWidth) {
				effectiveContentMaxWidth = 1
			} else if effectiveContentMaxWidth < 0 {
				effectiveContentMaxWidth = 0
			}
			t.logger.Debug("calculateContentMaxWidth: Batch col %d, ConstraintTotalCellWidth=%d, PadL=%d, PadR=%d -> EffectiveContentMaxWidth=%d",
				colIdx, constraintTotalCellWidth, padLeftWidth, padRightWidth, effectiveContentMaxWidth)
		} else {
			effectiveContentMaxWidth = 0
			t.logger.Debug("calculateContentMaxWidth: Batch col %d, No applicable MaxWidth constraint. EffectiveContentMaxWidth set to 0 (unlimited for this stage).", colIdx)
		}
	}
	return effectiveContentMaxWidth
}

// ensureStreamWidthsCalculated ensures that stream widths and column count are initialized for streaming mode.
// It uses sampleData and sectionConfig to calculate widths if not already set.
// Returns an error if the column count cannot be determined.
func (t *Table) ensureStreamWidthsCalculated(sampleData []string, sectionConfig tw.CellConfig) error {
	if t.streamWidths != nil && t.streamWidths.Len() > 0 {
		t.logger.Debug("Stream widths already set: %v", t.streamWidths)
		return nil
	}
	t.calculateStreamWidths(sampleData, sectionConfig)
	if t.streamNumCols == 0 {
		t.logger.Warn("Failed to determine column count from sample data")
		return errors.New("failed to determine column count for streaming")
	}
	for i := 0; i < t.streamNumCols; i++ {
		if _, ok := t.streamWidths.OK(i); !ok {
			t.streamWidths.Set(i, 0)
		}
	}
	t.logger.Debug("Initialized stream widths: %v", t.streamWidths)
	return nil
}
