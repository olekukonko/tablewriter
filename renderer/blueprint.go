package renderer

import (
	"fmt"
	"io"
	"strings"

	"github.com/olekukonko/tablewriter/tw"
	"github.com/olekukonko/tablewriter/twfn"
)

// Blueprint implements a basic default table renderer with customizable borders and alignments.
type Blueprint struct {
	config tw.RendererConfig // Rendering configuration
	trace  []string          // Debug trace messages
}

func NewBlueprint(configs ...tw.RendererConfig) *Blueprint {
	cfg := defaultOptions()
	cfg.Debug = true
	if len(configs) > 0 {
		userCfg := configs[0]
		if userCfg.Borders.Left != 0 {
			cfg.Borders.Left = userCfg.Borders.Left
		}
		if userCfg.Borders.Right != 0 {
			cfg.Borders.Right = userCfg.Borders.Right
		}
		if userCfg.Borders.Top != 0 {
			cfg.Borders.Top = userCfg.Borders.Top
		}
		if userCfg.Borders.Bottom != 0 {
			cfg.Borders.Bottom = userCfg.Borders.Bottom
		}
		if userCfg.Symbols != nil {
			cfg.Symbols = userCfg.Symbols
		}
		cfg.Settings = mergeSettings(cfg.Settings, userCfg.Settings)
	}
	return &Blueprint{config: cfg}
}

// Config returns the renderer's current configuration.
func (f *Blueprint) Config() tw.RendererConfig {
	return f.config
}

// debug logs a message to the trace if debugging is enabled.
func (f *Blueprint) debug(format string, a ...interface{}) {
	if f.config.Debug {
		msg := fmt.Sprintf(format, a...)
		f.trace = append(f.trace, fmt.Sprintf("[DEFAULT] %s", msg))
	}
}

// Debug returns the accumulated debug trace messages.
func (f *Blueprint) Debug() []string {
	return f.trace
}

// Header renders the table header section with configured formatting.
func (f *Blueprint) Header(w io.Writer, headers [][]string, ctx tw.Formatting) {
	f.debug("Starting Header render: IsSubRow=%v, Location=%v, Pos=%s, lines=%d, widths=%v",
		ctx.IsSubRow, ctx.Row.Location, ctx.Row.Position, len(ctx.Row.Current), ctx.Row.Widths)
	f.renderLine(w, ctx)
	f.debug("Completed Header render")
}

// Row renders a table data row with configured formatting.
func (f *Blueprint) Row(w io.Writer, row []string, ctx tw.Formatting) {
	f.debug("Starting Row render: IsSubRow=%v, Location=%v, Pos=%s, hasFooter=%v",
		ctx.IsSubRow, ctx.Row.Location, ctx.Row.Position, ctx.HasFooter)
	f.renderLine(w, ctx)
	f.debug("Completed Row render")
}

// Footer renders the table footer section with configured formatting.
func (f *Blueprint) Footer(w io.Writer, footers [][]string, ctx tw.Formatting) {
	f.debug("Starting Footer render: IsSubRow=%v, Location=%v, Pos=%s",
		ctx.IsSubRow, ctx.Row.Location, ctx.Row.Position)
	f.renderLine(w, ctx)
	f.debug("Completed Footer render")
}

// renderLine renders a single line (header, row, or footer) with borders, separators, and merge handling.
func (f *Blueprint) renderLine(w io.Writer, ctx tw.Formatting) {
	sortedKeys := twfn.ConvertToSortedKeys(ctx.Row.Widths)
	numCols := 0
	if len(sortedKeys) > 0 {
		numCols = sortedKeys[len(sortedKeys)-1] + 1
	} else {
		prefix := ""
		if f.config.Borders.Left.Enabled() {
			prefix = f.config.Symbols.Column()
		}
		suffix := ""
		if f.config.Borders.Right.Enabled() {
			suffix = f.config.Symbols.Column()
		}
		if prefix != "" || suffix != "" {
			fmt.Fprintln(w, prefix+suffix)
		}
		f.debug("renderLine: Handled empty row/widths case.")
		return
	}

	columnSeparator := f.config.Symbols.Column()
	prefix := ""
	if f.config.Borders.Left.Enabled() {
		prefix = columnSeparator
	}
	suffix := ""
	if f.config.Borders.Right.Enabled() {
		suffix = columnSeparator
	}

	var output strings.Builder
	output.WriteString(prefix)

	colIndex := 0
	separatorDisplayWidth := 0
	if f.config.Settings.Separators.BetweenColumns.Enabled() {
		separatorDisplayWidth = twfn.DisplayWidth(columnSeparator)
	}

	for colIndex < numCols {
		// Add separator if applicable
		shouldAddSeparator := false
		if colIndex > 0 && f.config.Settings.Separators.BetweenColumns.Enabled() {
			cellCtx, ok := ctx.Row.Current[colIndex]
			if !ok || !(cellCtx.Merge.Horizontal.Present && !cellCtx.Merge.Horizontal.Start) {
				shouldAddSeparator = true
			}
		}
		if shouldAddSeparator {
			output.WriteString(columnSeparator)
			f.debug("renderLine: Added separator '%s' before col %d", columnSeparator, colIndex)
		} else if colIndex > 0 {
			f.debug("renderLine: Skipped separator before col %d due to HMerge continuation", colIndex)
		}

		// Fetch cell context
		cellCtx, ok := ctx.Row.Current[colIndex]

		// Calculate width and span
		visualWidth := 0
		isHMergeStart := ok && cellCtx.Merge.Horizontal.Present && cellCtx.Merge.Horizontal.Start
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
					if k > 0 && separatorDisplayWidth > 0 {
						dynamicTotalWidth += separatorDisplayWidth
					}
				}
				visualWidth = dynamicTotalWidth
				f.debug("renderLine: Row HMerge col %d, span %d, dynamic visualWidth %d", colIndex, span, visualWidth)
			} else {
				visualWidth = ctx.Row.Widths.Get(colIndex)
				f.debug("renderLine: H/F HMerge col %d, span %d, pre-adjusted visualWidth %d", colIndex, span, visualWidth)
			}
		} else {
			visualWidth = ctx.Row.Widths.Get(colIndex)
			f.debug("renderLine: Regular col %d, visualWidth %d", colIndex, visualWidth)
		}
		if visualWidth < 0 {
			visualWidth = 0
		}

		// Skip merge continuation cells
		if ok && cellCtx.Merge.Horizontal.Present && !cellCtx.Merge.Horizontal.Start {
			f.debug("renderLine: Skipping col %d processing (part of HMerge)", colIndex)
			colIndex++
			continue
		}

		// Handle empty columns
		if !ok {
			if visualWidth > 0 {
				output.WriteString(strings.Repeat(" ", visualWidth))
				f.debug("renderLine: No cell context for col %d, writing %d spaces", colIndex, visualWidth)
			} else {
				f.debug("renderLine: No cell context for col %d, visualWidth is 0, writing nothing", colIndex)
			}
			colIndex += span
			continue
		}

		// Process cell content
		padding := cellCtx.Padding
		align := cellCtx.Align
		if align == tw.AlignNone {
			if ctx.Row.Position == tw.Header {
				align = tw.AlignCenter
			} else if ctx.Row.Position == tw.Footer {
				align = tw.AlignRight
			} else {
				align = tw.AlignLeft
			}
			f.debug("renderLine: col %d (data: '%s') using renderer default align '%s' for position %s.", colIndex, cellCtx.Data, align, ctx.Row.Position)
		} else if align == tw.Skip {
			if ctx.Row.Position == tw.Header {
				align = tw.AlignCenter
			} else if ctx.Row.Position == tw.Footer {
				align = tw.AlignRight
			} else {
				align = tw.AlignLeft
			}
			f.debug("renderLine: col %d (data: '%s') cellCtx.Align was Skip/empty, falling back to basic default '%s'.", colIndex, cellCtx.Data, align)
		}

		// Override alignment for footer patterns
		isTotalPattern := false
		if colIndex == 0 && isHMergeStart && cellCtx.Merge.Horizontal.Span >= 3 && strings.TrimSpace(cellCtx.Data) == "TOTAL" {
			isTotalPattern = true
		}
		if (ctx.Row.Position == tw.Footer && isHMergeStart) || isTotalPattern {
			if align != tw.AlignRight {
				f.debug("renderLine: Applying AlignRight HMerge/TOTAL override for Footer col %d. Original/default align was: %s", colIndex, align)
				align = tw.AlignRight
			}
		}

		// Handle merge blanking
		cellData := cellCtx.Data
		if (cellCtx.Merge.Vertical.Present && !cellCtx.Merge.Vertical.Start) ||
			(cellCtx.Merge.Hierarchical.Present && !cellCtx.Merge.Hierarchical.Start) {
			cellData = ""
			f.debug("renderLine: Blanked data for col %d (non-start V/Hierarchical)", colIndex)
		}

		// Format and append cell
		formattedCell := f.formatCell(cellData, visualWidth, padding, align)
		if len(formattedCell) > 0 {
			output.WriteString(formattedCell)
		}

		if isHMergeStart {
			f.debug("renderLine: Rendered HMerge START col %d (span %d, visualWidth %d, align %v): '%s'",
				colIndex, span, visualWidth, align, formattedCell)
		} else {
			f.debug("renderLine: Rendered regular col %d (visualWidth %d, align %v): '%s'",
				colIndex, visualWidth, align, formattedCell)
		}
		colIndex += span
	}

	output.WriteString(suffix)
	output.WriteString(tw.NewLine)
	fmt.Fprint(w, output.String())
	f.debug("renderLine: Final rendered line: %s", strings.TrimSuffix(output.String(), tw.NewLine))
}

// formatCell formats a cell's content with specified width, padding, and alignment, returning an empty string if width is non-positive.
func (f *Blueprint) formatCell(content string, width int, padding tw.Padding, align tw.Align) string {
	if width <= 0 {
		return ""
	}

	f.debug("Formatting cell: content='%s', width=%d, align=%s, padding={L:'%s' R:'%s'}",
		content, width, align, padding.Left, padding.Right)
	if f.config.Settings.TrimWhitespace.Enabled() {
		content = strings.TrimSpace(content)
		f.debug("Trimmed content: '%s'", content)
	}

	runeWidth := twfn.DisplayWidth(content)
	padLeftWidth := twfn.DisplayWidth(padding.Left)
	padRightWidth := twfn.DisplayWidth(padding.Right)
	totalPaddingWidth := padLeftWidth + padRightWidth

	availableContentWidth := width - totalPaddingWidth
	if availableContentWidth < 0 {
		availableContentWidth = 0
	}
	f.debug("Available content width: %d", availableContentWidth)

	if runeWidth > availableContentWidth {
		content = twfn.TruncateString(content, availableContentWidth)
		runeWidth = twfn.DisplayWidth(content)
		f.debug("Truncated content to fit %d: '%s' (new width %d)", availableContentWidth, content, runeWidth)
	}

	remainingSpace := width - runeWidth - totalPaddingWidth
	if remainingSpace < 0 {
		remainingSpace = 0
	}
	f.debug("Remaining space for alignment padding: %d", remainingSpace)

	leftPadChar := padding.Left
	rightPadChar := padding.Right
	if leftPadChar == "" {
		leftPadChar = tw.Space
	}
	if rightPadChar == "" {
		rightPadChar = tw.Space
	}
	leftPadCharWidth := twfn.DisplayWidth(leftPadChar)
	if leftPadCharWidth <= 0 {
		leftPadCharWidth = 1
	}
	rightPadCharWidth := twfn.DisplayWidth(rightPadChar)
	if rightPadCharWidth <= 0 {
		rightPadCharWidth = 1
	}

	var result strings.Builder
	var leftSpaces, rightSpaces int

	switch align {
	case tw.AlignLeft:
		leftSpaces = padLeftWidth
		rightSpaces = width - runeWidth - leftSpaces
	case tw.AlignRight:
		rightSpaces = padRightWidth
		leftSpaces = width - runeWidth - rightSpaces
	case tw.AlignCenter:
		leftSpaces = padLeftWidth + remainingSpace/2
		rightSpaces = width - runeWidth - leftSpaces
	default:
		leftSpaces = padLeftWidth
		rightSpaces = width - runeWidth - leftSpaces
	}

	if leftSpaces < 0 {
		leftSpaces = 0
	}
	if rightSpaces < 0 {
		rightSpaces = 0
	}

	if leftPadCharWidth > 0 {
		leftRepeat := leftSpaces / leftPadCharWidth
		result.WriteString(strings.Repeat(leftPadChar, leftRepeat))
		result.WriteString(strings.Repeat(" ", leftSpaces%leftPadCharWidth))
	} else {
		result.WriteString(strings.Repeat(" ", leftSpaces))
	}

	result.WriteString(content)

	if rightPadCharWidth > 0 {
		rightRepeat := rightSpaces / rightPadCharWidth
		result.WriteString(strings.Repeat(rightPadChar, rightRepeat))
		result.WriteString(strings.Repeat(" ", rightSpaces%rightPadCharWidth))
	} else {
		result.WriteString(strings.Repeat(" ", rightSpaces))
	}

	output := result.String()
	finalWidth := twfn.DisplayWidth(output)
	if finalWidth > width {
		output = twfn.TruncateString(output, width)
		f.debug("formatCell: Final check truncated output to width %d", width)
	} else if finalWidth < width {
		output += strings.Repeat(" ", width-finalWidth)
		f.debug("formatCell: Final check added %d spaces to meet width %d", width-finalWidth, width)
	}

	if f.config.Debug && twfn.DisplayWidth(output) != width {
		f.debug("formatCell Warning: Final width %d does not match target %d for result '%s'",
			twfn.DisplayWidth(output), width, output)
	}

	f.debug("Formatted cell final result: '%s' (target width %d)", output, width)
	return output
}

// Line renders a full horizontal row line with junctions and segments.
func (f *Blueprint) Line(w io.Writer, ctx tw.Formatting) {
	jr := NewJunction(JunctionContext{
		Symbols:       f.config.Symbols,
		Ctx:           ctx,
		ColIdx:        0,
		Debugging:     false,
		Debug:         f.debug,
		BorderTint:    Tint{},
		SeparatorTint: Tint{},
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
		if f.config.Borders.Left.Enabled() {
			prefix = jr.RenderLeft()
		}
		if f.config.Borders.Right.Enabled() {
			suffix = jr.RenderRight(-1)
		}
		if prefix != "" || suffix != "" {
			line.WriteString(prefix + suffix + tw.NewLine)
			fmt.Fprint(w, line.String())
		}
		f.debug("Line: Handled empty row/widths case")
		return
	}

	if f.config.Borders.Left.Enabled() {
		line.WriteString(jr.RenderLeft())
	}

	totalWidth := 0
	for i, colIdx := range sortedKeys {
		totalWidth += ctx.Row.Widths[colIdx]
		if i > 0 && f.config.Settings.Separators.BetweenColumns.Enabled() {
			totalWidth += twfn.DisplayWidth(jr.sym.Column())
		}
	}

	f.debug("Line: sortedKeys=%v, Widths=%v", sortedKeys, ctx.Row.Widths)
	for keyIndex, currentColIdx := range sortedKeys {
		jr.colIdx = currentColIdx
		segment := jr.GetSegment()
		colWidth := ctx.Row.Widths[currentColIdx]
		f.debug("Line: colIdx=%d, segment='%s', width=%d", currentColIdx, segment, colWidth)
		if segment == "" {
			line.WriteString(strings.Repeat(" ", colWidth))
		} else {
			repeat := colWidth / twfn.DisplayWidth(segment)
			if repeat < 1 && colWidth > 0 {
				repeat = 1
			}
			line.WriteString(strings.Repeat(segment, repeat))
		}

		isLast := keyIndex == len(sortedKeys)-1
		if !isLast && f.config.Settings.Separators.BetweenColumns.Enabled() {
			nextColIdx := sortedKeys[keyIndex+1]
			junction := jr.RenderJunction(currentColIdx, nextColIdx)
			f.debug("Line: Junction between %d and %d: '%s'", currentColIdx, nextColIdx, junction)
			line.WriteString(junction)
		}
	}

	if f.config.Borders.Right.Enabled() {
		lastIdx := sortedKeys[len(sortedKeys)-1]
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
	f.debug("Line rendered: %s", strings.TrimSuffix(line.String(), tw.NewLine))
}

func (f *Blueprint) Start(w io.Writer) error {
	f.debug("Blueprint.Start() called (no-op).")
	return nil
}

func (f *Blueprint) Close(w io.Writer) error {
	f.debug("Blueprint.Close() called (no-op).")
	return nil
}
