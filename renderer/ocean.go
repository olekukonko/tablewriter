package renderer

import (
	"fmt"
	"io"
	"strings"

	"github.com/olekukonko/ll"
	"github.com/olekukonko/tablewriter/tw"
	"github.com/olekukonko/tablewriter/twfn"
)

// OceanConfig defines configuration specific to the Ocean renderer.
type OceanConfig struct {
	// ColumWidth int // Default width for auto-calculation
}

// defaultOcean returns a default RendererConfig for the Ocean streaming renderer.
func defaultOcean() tw.RendererConfig {
	cfg := defaultBlueprint()                    // Use Blueprint's defaults as base
	cfg.Streaming = true                         // Mark as streaming capable
	cfg.Settings.Separators.BetweenRows = tw.Off // No between-row separators
	cfg.Settings.Lines.ShowFooterLine = tw.On    // Enable footer separator by default
	return cfg
}

// defaultOceanConfig provides default Ocean-specific configuration.
func defaultOceanConfig() OceanConfig {
	return OceanConfig{
		// ColumWidth: 25, // Default width for columns
	}
}

// Ocean is a streaming table renderer that writes ASCII tables with fixed column widths.
type Ocean struct {
	config           tw.RendererConfig   // Base renderer configuration
	oceanConfig      OceanConfig         // Ocean-specific configuration
	calculatedWidths tw.Mapper[int, int] // Widths used for rendering
	widthsCalculated bool                // Flag indicating if widths are finalized
	tableStarted     bool                // Tracks if rendering has started
	logger           *ll.Logger          // Logger for debugging
}

// NewOcean initializes an Ocean renderer with optional configuration.
func NewOcean(oceanCfgs ...OceanConfig) *Ocean {
	baseCfg := defaultOcean() // Start with default Blueprint-based config
	oceanCfg := defaultOceanConfig()

	renderer := &Ocean{
		config:           baseCfg,
		oceanConfig:      oceanCfg,
		calculatedWidths: tw.NewMapper[int, int](),
		widthsCalculated: false,
		logger:           ll.New("ocean"),
	}

	if renderer.widthsCalculated {
		renderer.calculatedWidths.Each(func(key int, val int) {
			if val < 1 {
				renderer.logger.Warn("Explicit ColumnWidths[%d]=%d is less than 1. Adjusting to 1.", key, val)
				renderer.calculatedWidths.Set(key, 1)
			}
		})
	}

	return renderer
}

// Logger sets the logger for the renderer.
func (o *Ocean) Logger(logger *ll.Logger) {
	o.logger = logger
}

// Config returns the renderer's current configuration.
func (o *Ocean) Config() tw.RendererConfig {
	return o.config
}

// Start prepares the renderer for drawing.
func (o *Ocean) Start(w io.Writer) error {
	if o.tableStarted {
		o.logger.Warn("Start() called but already started. Ignoring.")
		return nil
	}

	o.tableStarted = true
	o.logger.Debug("Start() called. tableStarted = true.")

	if o.widthsCalculated && o.config.Borders.Top.Enabled() && o.config.Settings.Lines.ShowTop.Enabled() {
		o.logger.Debug("Start: Rendering top border with explicit widths.")
		o.Line(w, tw.Formatting{
			Row: tw.RowContext{
				Widths:   o.calculatedWidths,
				Location: tw.LocationFirst,
				Position: tw.Header,
			},
			Level: tw.LevelHeader,
			Debug: true,
		})
	}

	return nil
}

// Header renders the table header section.
func (o *Ocean) Header(w io.Writer, headers [][]string, ctx tw.Formatting) {
	o.logger.Debug("Header() called: IsSubRow=%v, Location=%v, Pos=%s, lines=%d, widths=%v",
		ctx.IsSubRow, ctx.Row.Location, ctx.Row.Position, len(headers), ctx.Row.Widths)

	if len(headers) == 0 || len(headers[0]) == 0 {
		o.logger.Debug("Header: No headers to render")
		return
	}

	headerRow := headers[0]

	if !o.widthsCalculated {
		o.logger.Debug("Header: Calculating widths based on header data.")
		o.calculateWidths(headerRow)
		o.widthsCalculated = true
		o.logger.Debug("Calculated widths: %v", o.calculatedWidths)

		if o.config.Borders.Top.Enabled() && o.config.Settings.Lines.ShowTop.Enabled() {
			o.logger.Debug("Header: Rendering delayed top border.")
			o.Line(w, tw.Formatting{
				Row: tw.RowContext{
					Widths:   o.calculatedWidths,
					Location: tw.LocationFirst,
					Position: tw.Header,
				},
				Level: tw.LevelHeader,
				Debug: true,
			})
		}
	}

	ctx.Row.Widths = o.calculatedWidths // Override ctx widths with fixed widths
	o.renderLine(w, ctx, headerRow)

	if o.config.Settings.Lines.ShowHeaderLine.Enabled() {
		o.logger.Debug("Header: Rendering header separator line.")
		sepCtx := ctx
		sepCtx.Row.Widths = o.calculatedWidths
		sepCtx.Row.Location = tw.LocationMiddle
		sepCtx.Level = tw.LevelBody
		sepCtx.Row.Current = make(map[int]tw.CellContext)
		o.Line(w, sepCtx)
	}
}

// Row renders a table data row.
func (o *Ocean) Row(w io.Writer, row []string, ctx tw.Formatting) {
	o.logger.Debug("Row() called: IsSubRow=%v, Location=%v, Pos=%s, hasFooter=%v",
		ctx.IsSubRow, ctx.Row.Location, ctx.Row.Position, ctx.HasFooter)

	if len(row) == 0 {
		o.logger.Debug("Row: No data to render")
		return
	}

	if !o.widthsCalculated {
		o.logger.Debug("Row: Calculating widths based on first row data.")
		o.calculateWidths(row)
		o.widthsCalculated = true
		o.logger.Debug("Calculated widths: %v", o.calculatedWidths)

		if o.config.Borders.Top.Enabled() && o.config.Settings.Lines.ShowTop.Enabled() {
			o.logger.Debug("Row: Rendering delayed top border.")
			o.Line(w, tw.Formatting{
				Row: tw.RowContext{
					Widths:   o.calculatedWidths,
					Location: tw.LocationFirst,
					Position: tw.Row,
				},
				Level: tw.LevelHeader,
				Debug: true,
			})
		}

		if o.config.Settings.Lines.ShowHeaderLine.Enabled() {
			o.logger.Debug("Row: Rendering delayed header separator (no header).")
			o.Line(w, tw.Formatting{
				Row: tw.RowContext{
					Widths:   o.calculatedWidths,
					Location: tw.LocationMiddle,
					Position: tw.Header,
				},
				Level: tw.LevelBody,
				Debug: true,
			})
		}
	}

	ctx.Row.Widths = o.calculatedWidths // Override ctx widths with fixed widths
	o.renderLine(w, ctx, row)
}

// Footer renders the table footer section.
func (o *Ocean) Footer(w io.Writer, footers [][]string, ctx tw.Formatting) {
	o.logger.Debug("Footer() called: IsSubRow=%v, Location=%v, Pos=%s",
		ctx.IsSubRow, ctx.Row.Location, ctx.Row.Position)

	if len(footers) == 0 || len(footers[0]) == 0 {
		o.logger.Debug("Footer: No footers to render")
		return
	}

	footerRow := footers[0]

	if !o.widthsCalculated {
		o.logger.Warn("Footer: Widths not calculated by header or row. Calculating based on footer.")
		o.calculateWidths(footerRow)
		o.widthsCalculated = true
		o.logger.Debug("Calculated widths: %v", o.calculatedWidths)

		if o.config.Borders.Top.Enabled() && o.config.Settings.Lines.ShowTop.Enabled() {
			o.logger.Debug("Footer: Rendering delayed top border.")
			o.Line(w, tw.Formatting{
				Row: tw.RowContext{
					Widths:   o.calculatedWidths,
					Location: tw.LocationFirst,
					Position: tw.Footer,
				},
				Level: tw.LevelHeader,
				Debug: true,
			})
		}

		if o.config.Settings.Lines.ShowHeaderLine.Enabled() {
			o.logger.Debug("Footer: Rendering delayed header separator.")
			o.Line(w, tw.Formatting{
				Row: tw.RowContext{
					Widths:   o.calculatedWidths,
					Location: tw.LocationMiddle,
					Position: tw.Header,
				},
				Level: tw.LevelBody,
				Debug: true,
			})
		}
	}

	ctx.Row.Widths = o.calculatedWidths // Override ctx widths with fixed widths

	// Render footer separator line if enabled
	if o.config.Settings.Lines.ShowFooterLine.Enabled() {
		o.logger.Debug("Footer: Rendering footer separator line.")
		sepCtx := ctx
		sepCtx.Row.Widths = o.calculatedWidths
		sepCtx.Row.Location = tw.LocationMiddle
		sepCtx.Level = tw.LevelFooter
		sepCtx.Row.Current = make(map[int]tw.CellContext)
		o.Line(w, sepCtx)
	}

	o.renderLine(w, ctx, footerRow)

	if o.config.Borders.Bottom.Enabled() && o.config.Settings.Lines.ShowBottom.Enabled() {
		o.logger.Debug("Footer: Rendering bottom border.")
		o.Line(w, tw.Formatting{
			Row: tw.RowContext{
				Widths:   o.calculatedWidths,
				Location: tw.LocationEnd,
				Position: tw.Footer,
			},
			Level: tw.LevelFooter,
			Debug: true,
		})
	}
}

// Line renders a horizontal separator line.
func (o *Ocean) Line(w io.Writer, ctx tw.Formatting) {
	if !o.widthsCalculated || o.calculatedWidths.Len() == 0 {
		o.logger.Debug("Line: No calculated widths set; skipping line.")
		return
	}

	ctx.Row.Widths = o.calculatedWidths // Ensure fixed widths are used
	o.logger.Debug("Line: Starting with Level=%v, Location=%v, IsSubRow=%v, Widths=%v",
		ctx.Level, ctx.Row.Location, ctx.IsSubRow, ctx.Row.Widths)

	jr := NewJunction(JunctionContext{
		Symbols:       o.config.Symbols,
		Ctx:           ctx,
		ColIdx:        0,
		BorderTint:    Tint{},
		SeparatorTint: Tint{},
		Logger:        o.logger,
	})

	var line strings.Builder
	sortedKeys := twfn.ConvertToSortedKeys(ctx.Row.Widths)
	numCols := 0
	if len(sortedKeys) > 0 {
		numCols = sortedKeys[len(sortedKeys)-1] + 1
	}

	if numCols == 0 {
		prefix := ""
		suffix := ""
		if o.config.Borders.Left.Enabled() {
			prefix = jr.RenderLeft()
		}
		if o.config.Borders.Right.Enabled() {
			suffix = jr.RenderRight(-1)
		}
		if prefix != "" || suffix != "" {
			line.WriteString(prefix + suffix + tw.NewLine)
			fmt.Fprint(w, line.String())
		}
		o.logger.Debug("Line: Handled empty row/widths case")
		return
	}

	if o.config.Borders.Left.Enabled() {
		line.WriteString(jr.RenderLeft())
	}

	totalWidth := 0
	visibleColIndices := make([]int, 0)
	for i, colIdx := range sortedKeys {
		colWidth := ctx.Row.Widths.Get(colIdx)
		if colWidth > 0 {
			visibleColIndices = append(visibleColIndices, colIdx)
			totalWidth += colWidth
			if i > 0 && o.config.Settings.Separators.BetweenColumns.Enabled() {
				prevColIdx := sortedKeys[i-1]
				prevColWidth := ctx.Row.Widths.Get(prevColIdx)
				if prevColWidth > 0 {
					totalWidth += twfn.DisplayWidth(jr.sym.Column())
				}
			}
		}
	}

	o.logger.Debug("Line: sortedKeys=%v, Widths=%v, visibleColIndices=%v", sortedKeys, ctx.Row.Widths, visibleColIndices)
	for keyIndex, currentColIdx := range visibleColIndices {
		jr.colIdx = currentColIdx
		segment := jr.GetSegment()
		colWidth := ctx.Row.Widths.Get(currentColIdx)
		o.logger.Debug("Line: colIdx=%d, segment='%s', width=%d", currentColIdx, segment, colWidth)
		if segment == "" {
			line.WriteString(strings.Repeat(" ", colWidth))
		} else {
			repeat := colWidth / twfn.DisplayWidth(segment)
			if repeat < 1 && colWidth > 0 {
				repeat = 1
			}
			line.WriteString(strings.Repeat(segment, repeat))
		}

		isLast := keyIndex == len(visibleColIndices)-1
		if !isLast && o.config.Settings.Separators.BetweenColumns.Enabled() {
			nextColIdx := visibleColIndices[keyIndex+1]
			junction := jr.RenderJunction(currentColIdx, nextColIdx)
			o.logger.Debug("Line: Junction between %d and %d: '%s'", currentColIdx, nextColIdx, junction)
			line.WriteString(junction)
		}
	}

	if o.config.Borders.Right.Enabled() && len(visibleColIndices) > 0 {
		lastIdx := visibleColIndices[len(visibleColIndices)-1]
		line.WriteString(jr.RenderRight(lastIdx))
		actualWidth := twfn.DisplayWidth(line.String()) - twfn.DisplayWidth(jr.RenderLeft()) - twfn.DisplayWidth(jr.RenderRight(lastIdx))
		if actualWidth > totalWidth {
			lineStr := line.String()
			line.Reset()
			excess := actualWidth - totalWidth
			line.WriteString(lineStr[:len(lineStr)-excess-twfn.DisplayWidth(jr.RenderRight(lastIdx))] + jr.RenderRight(lastIdx))
		}
	}

	line.WriteString(tw.NewLine)
	fmt.Fprint(w, line.String())
	o.logger.Debug("Line rendered: %s", strings.TrimSuffix(line.String(), tw.NewLine))
}

// Close finalizes the renderer.
func (o *Ocean) Close(w io.Writer) error {
	o.logger.Debug("Close() called (no-op).")
	return nil
}

// renderLine renders a single line (header, row, or footer) with borders and separators.
func (o *Ocean) renderLine(w io.Writer, ctx tw.Formatting, line []string) {
	ctx.Row.Widths = o.calculatedWidths // Ensure fixed widths are used
	sortedKeys := twfn.ConvertToSortedKeys(ctx.Row.Widths)
	numCols := 0
	if len(sortedKeys) > 0 {
		numCols = sortedKeys[len(sortedKeys)-1] + 1
	} else {
		prefix := ""
		if o.config.Borders.Left.Enabled() {
			prefix = o.config.Symbols.Column()
		}
		suffix := ""
		if o.config.Borders.Right.Enabled() {
			suffix = o.config.Symbols.Column()
		}
		if prefix != "" || suffix != "" {
			fmt.Fprintln(w, prefix+suffix)
		}
		o.logger.Debug("renderLine: Handled empty row/widths case.")
		return
	}

	columnSeparator := o.config.Symbols.Column()
	prefix := ""
	if o.config.Borders.Left.Enabled() {
		prefix = columnSeparator
	}
	suffix := ""
	if o.config.Borders.Right.Enabled() {
		suffix = columnSeparator
	}

	var output strings.Builder
	output.WriteString(prefix)

	colIndex := 0
	separatorDisplayWidth := 0
	if o.config.Settings.Separators.BetweenColumns.Enabled() {
		separatorDisplayWidth = twfn.DisplayWidth(columnSeparator)
	}

	for colIndex < numCols {
		visualWidth := ctx.Row.Widths.Get(colIndex)
		cellCtx, ok := ctx.Row.Current[colIndex]
		isHMergeStart := ok && cellCtx.Merge.Horizontal.Present && cellCtx.Merge.Horizontal.Start
		if visualWidth == 0 && !isHMergeStart {
			o.logger.Debug("renderLine: Skipping col %d (zero width, not HMerge start)", colIndex)
			colIndex++
			continue
		}

		shouldAddSeparator := false
		if colIndex > 0 && o.config.Settings.Separators.BetweenColumns.Enabled() {
			prevWidth := ctx.Row.Widths.Get(colIndex - 1)
			prevCellCtx, prevOk := ctx.Row.Current[colIndex-1]
			prevIsHMergeEnd := prevOk && prevCellCtx.Merge.Horizontal.Present && prevCellCtx.Merge.Horizontal.End
			if (prevWidth > 0 || prevIsHMergeEnd) && (!ok || !(cellCtx.Merge.Horizontal.Present && !cellCtx.Merge.Horizontal.Start)) {
				shouldAddSeparator = true
			}
		}
		if shouldAddSeparator {
			output.WriteString(columnSeparator)
			o.logger.Debug("renderLine: Added separator '%s' before col %d", columnSeparator, colIndex)
		} else if colIndex > 0 {
			o.logger.Debug("renderLine: Skipped separator before col %d due to zero-width prev col or HMerge continuation", colIndex)
		}

		span := 1
		if isHMergeStart {
			span = cellCtx.Merge.Horizontal.Span
			if ctx.Row.Position == tw.Row {
				dynamicTotalWidth := 0
				for k := 0; k < span && colIndex+k < numCols; k++ {
					normWidth := ctx.NormalizedWidths.Get(colIndex + k)
					if normWidth < 0 {
						normWidth = 0
					}
					dynamicTotalWidth += normWidth
					if k > 0 && separatorDisplayWidth > 0 && ctx.NormalizedWidths.Get(colIndex+k) > 0 {
						dynamicTotalWidth += separatorDisplayWidth
					}
				}
				visualWidth = dynamicTotalWidth
				o.logger.Debug("renderLine: Row HMerge col %d, span %d, dynamic visualWidth %d", colIndex, span, visualWidth)
			} else {
				visualWidth = ctx.Row.Widths.Get(colIndex)
				o.logger.Debug("renderLine: H/F HMerge col %d, span %d, pre-adjusted visualWidth %d", colIndex, span, visualWidth)
			}
		} else {
			visualWidth = ctx.Row.Widths.Get(colIndex)
			o.logger.Debug("renderLine: Regular col %d, visualWidth %d", colIndex, visualWidth)
		}
		if visualWidth < 0 {
			visualWidth = 0
		}

		if ok && cellCtx.Merge.Horizontal.Present && !cellCtx.Merge.Horizontal.Start {
			o.logger.Debug("renderLine: Skipping col %d processing (part of HMerge)", colIndex)
			colIndex++
			continue
		}

		if !ok {
			if visualWidth > 0 {
				output.WriteString(strings.Repeat(" ", visualWidth))
				o.logger.Debug("renderLine: No cell context for col %d, writing %d spaces", colIndex, visualWidth)
			}
			colIndex += span
			continue
		}

		padding := cellCtx.Padding
		if padding.Left == "" {
			padding.Left = " "
		}
		if padding.Right == "" {
			padding.Right = " "
		}
		align := cellCtx.Align
		if align == tw.AlignNone {
			if ctx.Row.Position == tw.Header {
				align = tw.AlignCenter
			} else if ctx.Row.Position == tw.Footer {
				align = tw.AlignRight
			} else {
				align = tw.AlignLeft
			}
			o.logger.Debug("renderLine: col %d (data: '%s') using renderer default align '%s' for position %s.", colIndex, cellCtx.Data, align, ctx.Row.Position)
		} else if align == tw.Skip {
			if ctx.Row.Position == tw.Header {
				align = tw.AlignCenter
			} else if ctx.Row.Position == tw.Footer {
				align = tw.AlignRight
			} else {
				align = tw.AlignLeft
			}
			o.logger.Debug("renderLine: col %d (data: '%s') cellCtx.Align was Skip/empty, falling back to basic default '%s'.", colIndex, cellCtx.Data, align)
		}

		isTotalPattern := false
		if colIndex == 0 && isHMergeStart && cellCtx.Merge.Horizontal.Span >= 3 && strings.TrimSpace(cellCtx.Data) == "TOTAL" {
			isTotalPattern = true
		}
		if (ctx.Row.Position == tw.Footer && isHMergeStart) || isTotalPattern {
			if align != tw.AlignRight {
				o.logger.Debug("renderLine: Applying AlignRight HMerge/TOTAL override for Footer col %d. Original/default align was: %s", colIndex, align)
				align = tw.AlignRight
			}
		}

		cellData := cellCtx.Data
		if (cellCtx.Merge.Vertical.Present && !cellCtx.Merge.Vertical.Start) ||
			(cellCtx.Merge.Hierarchical.Present && !cellCtx.Merge.Hierarchical.Start) {
			o.logger.Warn("renderLine: Ignoring vertical/hierarchical merge for col %d in streaming mode.", colIndex)
			cellData = ""
		}

		formattedCell := o.formatCell(cellData, visualWidth, padding, align)
		if len(formattedCell) > 0 {
			output.WriteString(formattedCell)
		}

		if isHMergeStart {
			o.logger.Debug("renderLine: Rendered HMerge START col %d (span %d, visualWidth %d, align %v): '%s'",
				colIndex, span, visualWidth, align, formattedCell)
		} else {
			o.logger.Debug("renderLine: Rendered regular col %d (visualWidth %d, align %v): '%s'",
				colIndex, visualWidth, align, formattedCell)
		}
		colIndex += span
	}

	if output.Len() > len(prefix) || o.config.Borders.Right.Enabled() {
		output.WriteString(suffix)
	}
	output.WriteString(tw.NewLine)
	fmt.Fprint(w, output.String())
	o.logger.Debug("renderLine: Final rendered line: %s", strings.TrimSuffix(output.String(), tw.NewLine))
}

// formatCell formats a cell's content with specified width, padding, and alignment.
func (o *Ocean) formatCell(content string, width int, padding tw.Padding, align tw.Align) string {
	if width <= 0 {
		return ""
	}

	o.logger.Debug("Formatting cell: content='%s', width=%d, align=%s, padding={L:'%s' R:'%s'}",
		content, width, align, padding.Left, padding.Right)

	if o.config.Settings.TrimWhitespace.Enabled() {
		content = strings.TrimSpace(content)
		o.logger.Debug("Trimmed content: '%s'", content)
	}

	runeWidth := twfn.DisplayWidth(content)

	leftPadChar := padding.Left
	rightPadChar := padding.Right
	if leftPadChar == "" {
		leftPadChar = " "
	}
	if rightPadChar == "" {
		rightPadChar = " "
	}

	padLeftWidth := twfn.DisplayWidth(leftPadChar)
	padRightWidth := twfn.DisplayWidth(rightPadChar)

	availableContentWidth := width - padLeftWidth - padRightWidth
	if availableContentWidth < 0 {
		availableContentWidth = 0
	}
	o.logger.Debug("Available content width: %d", availableContentWidth)

	if runeWidth > availableContentWidth {
		content = twfn.TruncateString(content, availableContentWidth)
		runeWidth = twfn.DisplayWidth(content)
		o.logger.Debug("Truncated content to fit %d: '%s' (new width %d)", availableContentWidth, content, runeWidth)
	}

	totalPaddingWidth := width - runeWidth
	if totalPaddingWidth < 0 {
		totalPaddingWidth = 0
	}
	o.logger.Debug("Total padding width: %d", totalPaddingWidth)

	var result strings.Builder
	var leftPaddingWidth, rightPaddingWidth int

	switch align {
	case tw.AlignLeft:
		result.WriteString(leftPadChar)
		result.WriteString(content)
		rightPaddingWidth = totalPaddingWidth - padLeftWidth
		if rightPaddingWidth > 0 {
			result.WriteString(strings.Repeat(rightPadChar, rightPaddingWidth/padRightWidth))
			leftover := rightPaddingWidth % padRightWidth
			if leftover > 0 {
				result.WriteString(strings.Repeat(" ", leftover))
				o.logger.Debug("Added %d leftover spaces for right padding", leftover)
			}
			o.logger.Debug("Applied right padding: '%s' for %d width", rightPadChar, rightPaddingWidth)
		}
	case tw.AlignRight:
		leftPaddingWidth = totalPaddingWidth - padRightWidth
		if leftPaddingWidth > 0 {
			result.WriteString(strings.Repeat(leftPadChar, leftPaddingWidth/padLeftWidth))
			leftover := leftPaddingWidth % padLeftWidth
			if leftover > 0 {
				result.WriteString(strings.Repeat(" ", leftover))
				o.logger.Debug("Added %d leftover spaces for left padding", leftover)
			}
			o.logger.Debug("Applied left padding: '%s' for %d width", leftPadChar, leftPaddingWidth)
		}
		result.WriteString(content)
		result.WriteString(rightPadChar)
	case tw.AlignCenter:
		leftPaddingWidth = (totalPaddingWidth-padLeftWidth-padRightWidth)/2 + padLeftWidth
		rightPaddingWidth = totalPaddingWidth - leftPaddingWidth
		if leftPaddingWidth > padLeftWidth {
			result.WriteString(strings.Repeat(leftPadChar, (leftPaddingWidth-padLeftWidth)/padLeftWidth))
			leftover := (leftPaddingWidth - padLeftWidth) % padLeftWidth
			if leftover > 0 {
				result.WriteString(strings.Repeat(" ", leftover))
				o.logger.Debug("Added %d leftover spaces for left centering", leftover)
			}
			o.logger.Debug("Applied left centering padding: '%s' for %d width", leftPadChar, leftPaddingWidth-padLeftWidth)
		}
		result.WriteString(leftPadChar)
		result.WriteString(content)
		result.WriteString(rightPadChar)
		if rightPaddingWidth > padRightWidth {
			result.WriteString(strings.Repeat(rightPadChar, (rightPaddingWidth-padRightWidth)/padRightWidth))
			leftover := (rightPaddingWidth - padRightWidth) % padRightWidth
			if leftover > 0 {
				result.WriteString(strings.Repeat(" ", leftover))
				o.logger.Debug("Added %d leftover spaces for right centering", leftover)
			}
			o.logger.Debug("Applied right centering padding: '%s' for %d width", rightPadChar, rightPaddingWidth-padRightWidth)
		}
	default:
		result.WriteString(leftPadChar)
		result.WriteString(content)
		rightPaddingWidth = totalPaddingWidth - padLeftWidth
		if rightPaddingWidth > 0 {
			result.WriteString(strings.Repeat(rightPadChar, rightPaddingWidth/padRightWidth))
			leftover := rightPaddingWidth % padRightWidth
			if leftover > 0 {
				result.WriteString(strings.Repeat(" ", leftover))
				o.logger.Debug("Added %d leftover spaces for right padding", leftover)
			}
			o.logger.Debug("Applied right padding: '%s' for %d width", rightPadChar, rightPaddingWidth)
		}
	}

	output := result.String()
	finalWidth := twfn.DisplayWidth(output)
	if finalWidth > width {
		output = twfn.TruncateString(output, width)
		o.logger.Debug("formatCell: Truncated output to width %d", width)
	} else if finalWidth < width {
		result.WriteString(strings.Repeat(" ", width-finalWidth))
		output = result.String()
		o.logger.Debug("formatCell: Added %d spaces to meet width %d", width-finalWidth, width)
	}

	if o.logger.Enabled() && twfn.DisplayWidth(output) != width {
		o.logger.Debug("formatCell Warning: Final width %d does not match target %d for result '%s'",
			twfn.DisplayWidth(output), width, output)
	}

	o.logger.Debug("Formatted cell final result: '%s' (target width %d)", output, width)
	return output
}

// calculateWidths calculates column widths based on the provided data.
func (o *Ocean) calculateWidths(data []string) {
	if o.widthsCalculated {
		o.logger.Warn("calculateWidths called but widths already set (%v). Skipping.", o.calculatedWidths)
		return
	}

	o.logger.Debug("calculateWidths: Calculating widths based on data: %v", data)

	o.calculatedWidths = tw.NewMapper[int, int]()
	numColsInData := len(data)

	for i := 0; i < numColsInData; i++ {
		content := ""
		if i < len(data) {
			content = data[i]
		}

		contentWidth := twfn.DisplayWidth(strings.TrimSpace(content))
		totalWidth := contentWidth + twfn.DisplayWidth(" ") + twfn.DisplayWidth(" ") // Default padding (space left + right)

		//
		//if contentWidth == 0 && totalWidth < o.oceanConfig.ColumWidth {
		//	totalWidth = o.oceanConfig.ColumWidth
		//}

		if totalWidth < 1 {
			totalWidth = 1
		}

		o.calculatedWidths.Set(i, totalWidth)
		o.logger.Debug("calculateWidths: Col %d calculated width %d (content:%d + padding)", i, totalWidth, contentWidth)
	}

	o.widthsCalculated = true
	o.logger.Debug("calculateWidths: Final calculated widths: %v", o.calculatedWidths)
}
