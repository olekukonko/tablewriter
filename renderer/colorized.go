package renderer

import (
	"fmt"
	"github.com/fatih/color"
	"io"
	"strings"

	"github.com/olekukonko/tablewriter/tw"
	"github.com/olekukonko/tablewriter/twfn"
)

// Colors is a slice of color attributes for use with fatih/color, such as color.FgWhite or color.Bold.
type Colors []color.Attribute

// Tint defines foreground and background color settings for table elements, with optional per-column overrides.
type Tint struct {
	FG      Colors // Foreground color attributes
	BG      Colors // Background color attributes
	Columns []Tint // Per-column color settings
}

// Apply applies the Tint's foreground and background colors to the given text, returning the text unchanged if no colors are set.
func (t Tint) Apply(text string) string {
	if len(t.FG) == 0 && len(t.BG) == 0 {
		return text
	}
	combinedColors := append(t.FG, t.BG...)
	c := color.New(combinedColors...).SprintFunc()
	return c(text)
}

// ColorizedConfig holds configuration for the Colorized table renderer.
type ColorizedConfig struct {
	Borders   tw.Border   // Border visibility settings
	Settings  tw.Settings // Rendering behavior settings (e.g., separators, whitespace)
	Header    Tint        // Colors for header cells
	Column    Tint        // Colors for row cells
	Footer    Tint        // Colors for footer cells
	Border    Tint        // Colors for borders and lines
	Separator Tint        // Colors for column separators
	Symbols   tw.Symbols  // Symbols for table drawing (e.g., corners, lines)
	Debug     bool        // Enables debug logging
}

// NewColorized creates a Colorized renderer with the specified configuration, falling back to defaults if none provided.
// Only the first config is used if multiple are passed.
func NewColorized(configs ...ColorizedConfig) *Colorized {
	baseCfg := defaultColorized()

	if len(configs) > 0 {
		userCfg := configs[0]

		// Merge borders
		if userCfg.Borders.Left != 0 {
			baseCfg.Borders.Left = userCfg.Borders.Left
		}
		if userCfg.Borders.Right != 0 {
			baseCfg.Borders.Right = userCfg.Borders.Right
		}
		if userCfg.Borders.Top != 0 {
			baseCfg.Borders.Top = userCfg.Borders.Top
		}
		if userCfg.Borders.Bottom != 0 {
			baseCfg.Borders.Bottom = userCfg.Borders.Bottom
		}

		// Merge settings
		baseCfg.Settings.Separators = mergeSeparators(baseCfg.Settings.Separators, userCfg.Settings.Separators)
		baseCfg.Settings.Lines = mergeLines(baseCfg.Settings.Lines, userCfg.Settings.Lines)
		if userCfg.Settings.TrimWhitespace != 0 {
			baseCfg.Settings.TrimWhitespace = userCfg.Settings.TrimWhitespace
		}
		if userCfg.Settings.CompactMode != 0 {
			baseCfg.Settings.CompactMode = userCfg.Settings.CompactMode
		}

		// Replace tints if any part is set
		if len(userCfg.Header.FG) > 0 || len(userCfg.Header.BG) > 0 || userCfg.Header.Columns != nil {
			baseCfg.Header = userCfg.Header
		}
		if len(userCfg.Column.FG) > 0 || len(userCfg.Column.BG) > 0 || userCfg.Column.Columns != nil {
			baseCfg.Column = userCfg.Column
		}
		if len(userCfg.Footer.FG) > 0 || len(userCfg.Footer.BG) > 0 || userCfg.Footer.Columns != nil {
			baseCfg.Footer = userCfg.Footer
		}
		if len(userCfg.Border.FG) > 0 || len(userCfg.Border.BG) > 0 || userCfg.Border.Columns != nil {
			baseCfg.Border = userCfg.Border
		}
		if len(userCfg.Separator.FG) > 0 || len(userCfg.Separator.BG) > 0 || userCfg.Separator.Columns != nil {
			baseCfg.Separator = userCfg.Separator
		}

		if userCfg.Symbols != nil {
			baseCfg.Symbols = userCfg.Symbols
		}
		baseCfg.Debug = userCfg.Debug
	}

	cfg := baseCfg
	if cfg.Symbols == nil {
		cfg.Symbols = tw.NewSymbols(tw.StyleLight)
	}

	f := &Colorized{
		config:  cfg,
		newLine: tw.NewLine,
		defaultAlign: map[tw.Position]tw.Align{
			tw.Header: tw.AlignCenter,
			tw.Row:    tw.AlignLeft,
			tw.Footer: tw.AlignRight,
		},
	}
	f.debug("Initialized Colorized renderer with symbols: Center=%q, Row=%q, Column=%q", f.config.Symbols.Center(), f.config.Symbols.Row(), f.config.Symbols.Column())
	f.debug("Final ColorizedConfig.Settings.Lines: %+v", f.config.Settings.Lines)
	f.debug("Final ColorizedConfig.Borders: %+v", f.config.Borders)
	return f
}

// Colorized renders colored ASCII tables with customizable borders, colors, and alignments.
type Colorized struct {
	config       ColorizedConfig          // Renderer configuration
	trace        []string                 // Debug trace messages
	newLine      string                   // Newline character
	defaultAlign map[tw.Position]tw.Align // Default alignments for header, row, and footer
}

// Config returns the renderer's configuration as a RendererConfig.
func (f *Colorized) Config() tw.RendererConfig {
	return tw.RendererConfig{
		Borders:  f.config.Borders,
		Settings: f.config.Settings,
		Symbols:  f.config.Symbols,
		Debug:    f.config.Debug,
	}
}

// debug logs a message to the trace if debugging is enabled.
func (f *Colorized) debug(format string, a ...interface{}) {
	if f.config.Debug {
		msg := fmt.Sprintf(format, a...)
		f.trace = append(f.trace, fmt.Sprintf("[COLORIZED] %s", msg))
	}
}

// Debug returns the accumulated debug trace messages.
func (f *Colorized) Debug() []string {
	return f.trace
}

// Reset clears the renderer's internal state, including debug traces.
func (f *Colorized) Reset() {
	f.trace = nil
	f.debug("Reset: Cleared debug trace")
}

// formatCell formats a cell's content with color, width, padding, and alignment, handling whitespace trimming and truncation.
func (f *Colorized) formatCell(content string, width int, padding tw.Padding, align tw.Align, tint Tint) string {
	if f.config.Settings.TrimWhitespace.Enabled() {
		content = strings.TrimSpace(content)
	}
	f.debug("Formatting cell: content='%s', width=%d, align=%s, paddingL='%s', paddingR='%s', tintFG=%v, tintBG=%v",
		content, width, align, padding.Left, padding.Right, tint.FG, tint.BG)

	if width <= 0 {
		f.debug("formatCell: width %d <= 0, returning empty string", width)
		return ""
	}

	contentVisualWidth := twfn.DisplayWidth(content)

	// Use default padding if not specified
	padLeftCharStr := padding.Left
	if padLeftCharStr == "" {
		padLeftCharStr = tw.Space
	}
	padRightCharStr := padding.Right
	if padRightCharStr == "" {
		padRightCharStr = tw.Space
	}

	// Calculate padding widths
	definedPadLeftWidth := twfn.DisplayWidth(padLeftCharStr)
	definedPadRightWidth := twfn.DisplayWidth(padRightCharStr)
	availableForContentAndAlign := width - definedPadLeftWidth - definedPadRightWidth
	if availableForContentAndAlign < 0 {
		availableForContentAndAlign = 0
	}

	// Truncate content if too wide
	if contentVisualWidth > availableForContentAndAlign {
		content = twfn.TruncateString(content, availableForContentAndAlign)
		contentVisualWidth = twfn.DisplayWidth(content)
		f.debug("Truncated content to fit %d: '%s' (new width %d)", availableForContentAndAlign, content, contentVisualWidth)
	}

	// Calculate alignment padding
	remainingSpaceForAlignment := availableForContentAndAlign - contentVisualWidth
	if remainingSpaceForAlignment < 0 {
		remainingSpaceForAlignment = 0
	}

	leftAlignmentPadSpaces := ""
	rightAlignmentPadSpaces := ""
	switch align {
	case tw.AlignLeft:
		rightAlignmentPadSpaces = strings.Repeat(tw.Space, remainingSpaceForAlignment)
	case tw.AlignRight:
		leftAlignmentPadSpaces = strings.Repeat(tw.Space, remainingSpaceForAlignment)
	case tw.AlignCenter:
		leftSpacesCount := remainingSpaceForAlignment / 2
		rightSpacesCount := remainingSpaceForAlignment - leftSpacesCount
		leftAlignmentPadSpaces = strings.Repeat(tw.Space, leftSpacesCount)
		rightAlignmentPadSpaces = strings.Repeat(tw.Space, rightSpacesCount)
	default:
		rightAlignmentPadSpaces = strings.Repeat(tw.Space, remainingSpaceForAlignment)
	}

	// Apply tinting to components
	coloredContent := tint.Apply(content)
	coloredPadLeft := padLeftCharStr
	coloredPadRight := padRightCharStr
	coloredAlignPadLeft := leftAlignmentPadSpaces
	coloredAlignPadRight := rightAlignmentPadSpaces

	if len(tint.BG) > 0 {
		bgTint := Tint{BG: tint.BG}
		if len(tint.FG) > 0 && padLeftCharStr != tw.Space {
			coloredPadLeft = tint.Apply(padLeftCharStr)
		} else {
			coloredPadLeft = bgTint.Apply(padLeftCharStr)
		}
		if len(tint.FG) > 0 && padRightCharStr != tw.Space {
			coloredPadRight = tint.Apply(padRightCharStr)
		} else {
			coloredPadRight = bgTint.Apply(padRightCharStr)
		}
		if leftAlignmentPadSpaces != "" {
			coloredAlignPadLeft = bgTint.Apply(leftAlignmentPadSpaces)
		}
		if rightAlignmentPadSpaces != "" {
			coloredAlignPadRight = bgTint.Apply(rightAlignmentPadSpaces)
		}
	} else if len(tint.FG) > 0 {
		if padLeftCharStr != tw.Space {
			coloredPadLeft = tint.Apply(padLeftCharStr)
		}
		if padRightCharStr != tw.Space {
			coloredPadRight = tint.Apply(padRightCharStr)
		}
	}

	// Build final cell string
	var sb strings.Builder
	sb.WriteString(coloredPadLeft)
	sb.WriteString(coloredAlignPadLeft)
	sb.WriteString(coloredContent)
	sb.WriteString(coloredAlignPadRight)
	sb.WriteString(coloredPadRight)
	output := sb.String()

	// Ensure correct visual width
	currentVisualWidth := twfn.DisplayWidth(output)
	if currentVisualWidth != width {
		f.debug("formatCell MISMATCH: content='%s', target_w=%d. Calculated parts width = %d. String: '%s'",
			content, width, currentVisualWidth, output)
		if currentVisualWidth > width {
			output = twfn.TruncateString(output, width)
		} else {
			paddingSpacesStr := strings.Repeat(" ", width-currentVisualWidth)
			if len(tint.BG) > 0 {
				output += Tint{BG: tint.BG}.Apply(paddingSpacesStr)
			} else {
				output += paddingSpacesStr
			}
		}
		f.debug("formatCell Post-Correction: Target %d, New Visual width %d. Output: '%s'", width, twfn.DisplayWidth(output), output)
	}

	f.debug("Formatted cell final result: '%s' (target width %d, display width %d)", output, width, twfn.DisplayWidth(output))
	return output
}

// Header renders the table header with configured colors and formatting.
func (f *Colorized) Header(w io.Writer, headers [][]string, ctx tw.Formatting) {
	f.debug("Starting Header render: IsSubRow=%v, Location=%v, Pos=%s, lines=%d, widths=%v",
		ctx.IsSubRow, ctx.Row.Location, ctx.Row.Position, len(headers), ctx.Row.Widths)

	if len(headers) == 0 || len(headers[0]) == 0 {
		f.debug("Header: No headers to render")
		return
	}

	f.renderLine(w, ctx, headers[0], f.config.Header)
	f.debug("Completed Header render")
}

// Row renders a table data row with configured colors and formatting.
func (f *Colorized) Row(w io.Writer, row []string, ctx tw.Formatting) {
	f.debug("Starting Row render: IsSubRow=%v, Location=%v, Pos=%s, hasFooter=%v",
		ctx.IsSubRow, ctx.Row.Location, ctx.Row.Position, ctx.HasFooter)

	if len(row) == 0 {
		f.debug("Row: No data to render")
		return
	}

	f.renderLine(w, ctx, row, f.config.Column)
	f.debug("Completed Row render")
}

// Footer renders the table footer with configured colors and formatting.
func (f *Colorized) Footer(w io.Writer, footers [][]string, ctx tw.Formatting) {
	f.debug("Starting Footer render: IsSubRow=%v, Location=%v, Pos=%s",
		ctx.IsSubRow, ctx.Row.Location, ctx.Row.Position)

	if len(footers) == 0 || len(footers[0]) == 0 {
		f.debug("Footer: No footers to render")
		return
	}

	f.renderLine(w, ctx, footers[0], f.config.Footer)
	f.debug("Completed Footer render")
}

// renderLine renders a single line (header, row, or footer) with colors, handling merges and separators.
func (f *Colorized) renderLine(w io.Writer, ctx tw.Formatting, line []string, tint Tint) {
	numCols := 0
	if len(ctx.Row.Current) > 0 {
		maxKey := -1
		for k := range ctx.Row.Current {
			if k > maxKey {
				maxKey = k
			}
		}
		numCols = maxKey + 1
	} else {
		maxKey := -1
		for k := range ctx.Row.Widths {
			if k > maxKey {
				maxKey = k
			}
		}
		numCols = maxKey + 1
	}

	var output strings.Builder

	// Render left border
	prefix := ""
	if f.config.Borders.Left.Enabled() {
		prefix = f.config.Border.Apply(f.config.Symbols.Column())
	}
	output.WriteString(prefix)

	// Calculate separator width
	separatorDisplayWidth := 0
	separatorString := ""
	if f.config.Settings.Separators.BetweenColumns.Enabled() {
		separatorString = f.config.Separator.Apply(f.config.Symbols.Column())
		separatorDisplayWidth = twfn.DisplayWidth(f.config.Symbols.Column())
	}

	for i := 0; i < numCols; {
		// Add separator if applicable
		shouldAddSeparator := false
		if i > 0 && f.config.Settings.Separators.BetweenColumns.Enabled() {
			cellCtx, ok := ctx.Row.Current[i]
			if !ok || !(cellCtx.Merge.Horizontal.Present && !cellCtx.Merge.Horizontal.Start) {
				shouldAddSeparator = true
			}
		}
		if shouldAddSeparator {
			output.WriteString(separatorString)
			f.debug("renderLine: Added separator '%s' before col %d", separatorString, i)
		} else if i > 0 {
			f.debug("renderLine: Skipped separator before col %d due to HMerge continuation", i)
		}

		// Get cell context
		cellCtx, ok := ctx.Row.Current[i]
		if !ok {
			cellCtx = tw.CellContext{
				Data:    "",
				Align:   f.defaultAlign[ctx.Row.Position],
				Padding: tw.Padding{Left: " ", Right: " "},
				Width:   ctx.Row.Widths.Get(i),
				Merge:   tw.MergeState{},
			}
		}

		// Determine cell width and span
		visualWidth := 0
		span := 1
		isHMergeStart := ok && cellCtx.Merge.Horizontal.Present && cellCtx.Merge.Horizontal.Start

		if isHMergeStart {
			span = cellCtx.Merge.Horizontal.Span
			if ctx.Row.Position == tw.Row {
				dynamicTotalWidth := 0
				for k := 0; k < span && i+k < numCols; k++ {
					colToSum := i + k
					normWidth := ctx.NormalizedWidths.Get(colToSum)
					if normWidth < 0 {
						normWidth = 0
					}
					dynamicTotalWidth += normWidth
					if k > 0 && separatorDisplayWidth > 0 {
						dynamicTotalWidth += separatorDisplayWidth
					}
				}
				visualWidth = dynamicTotalWidth
				f.debug("renderLine: Row HMerge col %d, span %d, dynamic visualWidth %d", i, span, visualWidth)
			} else {
				visualWidth = ctx.Row.Widths.Get(i)
				f.debug("renderLine: H/F HMerge col %d, span %d, pre-adjusted visualWidth %d", i, span, visualWidth)
			}
		} else {
			visualWidth = ctx.Row.Widths.Get(i)
			f.debug("renderLine: Regular col %d, visualWidth %d", i, visualWidth)
		}
		if visualWidth < 0 {
			visualWidth = 0
		}

		// Skip non-start merge cells
		if ok && cellCtx.Merge.Horizontal.Present && !cellCtx.Merge.Horizontal.Start {
			f.debug("renderLine: Skipping col %d processing (part of HMerge)", i)
			i++
			continue
		}

		// Handle empty columns
		if !ok && visualWidth > 0 {
			spaces := strings.Repeat(" ", visualWidth)
			if len(tint.BG) > 0 {
				output.WriteString(Tint{BG: tint.BG}.Apply(spaces))
			} else {
				output.WriteString(spaces)
			}
			f.debug("renderLine: No cell context for col %d, writing %d spaces", i, visualWidth)
			i += span
			continue
		}

		// Process cell content
		padding := cellCtx.Padding
		align := cellCtx.Align
		if align == tw.AlignNone {
			align = f.defaultAlign[ctx.Row.Position]
			f.debug("renderLine: col %d using default renderer align '%s' for position %s because cellCtx.Align was AlignNone", i, align, ctx.Row.Position)
		}

		// Apply alignment overrides for specific patterns
		isTotalPattern := false
		if i == 0 && isHMergeStart && cellCtx.Merge.Horizontal.Span >= 3 && strings.TrimSpace(cellCtx.Data) == "TOTAL" {
			isTotalPattern = true
			f.debug("renderLine: Detected 'TOTAL' HMerge pattern at col 0")
		}
		if (ctx.Row.Position == tw.Footer && isHMergeStart) || isTotalPattern {
			if align != tw.AlignRight {
				f.debug("renderLine: Applying AlignRight override for Footer HMerge/TOTAL pattern at col %d. Original/default align was: %s", i, align)
				align = tw.AlignRight
			}
		}

		// Handle merge blanking
		content := cellCtx.Data
		if (cellCtx.Merge.Vertical.Present && !cellCtx.Merge.Vertical.Start) ||
			(cellCtx.Merge.Hierarchical.Present && !cellCtx.Merge.Hierarchical.Start) {
			content = ""
			f.debug("renderLine: Blanked data for col %d (non-start V/Hierarchical)", i)
		}

		// Apply per-column tint if available
		cellTint := tint
		if i < len(tint.Columns) {
			columnTint := tint.Columns[i]
			if len(columnTint.FG) > 0 || len(columnTint.BG) > 0 {
				cellTint = columnTint
			}
		}

		// Format and append cell
		formattedCell := f.formatCell(content, visualWidth, padding, align, cellTint)
		if len(formattedCell) > 0 {
			output.WriteString(formattedCell)
		} else if visualWidth == 0 && isHMergeStart {
			f.debug("renderLine: Rendered HMerge START col %d resulted in 0 visual width, wrote nothing.", i)
		} else if visualWidth == 0 {
			f.debug("renderLine: Rendered regular col %d resulted in 0 visual width, wrote nothing.", i)
		}

		if isHMergeStart {
			f.debug("renderLine: Rendered HMerge START col %d (span %d, visualWidth %d, align %s): '%s'",
				i, span, visualWidth, align, formattedCell)
		} else {
			f.debug("renderLine: Rendered regular col %d (visualWidth %d, align %s): '%s'",
				i, visualWidth, align, formattedCell)
		}

		i += span
	}

	// Render right border
	suffix := ""
	if f.config.Borders.Right.Enabled() {
		suffix = f.config.Border.Apply(f.config.Symbols.Column())
	}
	output.WriteString(suffix)

	output.WriteString(f.newLine)
	fmt.Fprint(w, output.String())
	f.debug("renderLine: Final rendered line: %s", strings.TrimSuffix(output.String(), f.newLine))
}

// Line renders a horizontal row line with colored junctions and segments, skipping zero-width columns.
func (f *Colorized) Line(w io.Writer, ctx tw.Formatting) {
	f.debug("Line: Starting with Level=%v, Location=%v, IsSubRow=%v, Widths=%v", ctx.Level, ctx.Row.Location, ctx.IsSubRow, ctx.Row.Widths)

	jr := NewJunction(JunctionContext{
		Symbols:       f.config.Symbols,
		Ctx:           ctx,
		ColIdx:        0,
		Debugging:     true,
		Debug:         f.debug,
		BorderTint:    f.config.Border,
		SeparatorTint: f.config.Separator,
	})

	var line strings.Builder

	// Filter effective columns
	allSortedKeys := twfn.ConvertToSortedKeys(ctx.Row.Widths)
	effectiveKeys := []int{}
	keyWidthMap := make(map[int]int)

	for _, k := range allSortedKeys {
		width := ctx.Row.Widths.Get(k)
		keyWidthMap[k] = width
		if width > 0 {
			effectiveKeys = append(effectiveKeys, k)
		}
	}
	f.debug("Line: All keys=%v, Effective keys (width>0)=%v", allSortedKeys, effectiveKeys)

	// Handle empty table
	if len(effectiveKeys) == 0 {
		prefix := ""
		suffix := ""
		if f.config.Borders.Left.Enabled() {
			prefix = jr.RenderLeft()
		}
		if f.config.Borders.Right.Enabled() {
			originalLastColIdx := -1
			if len(allSortedKeys) > 0 {
				originalLastColIdx = allSortedKeys[len(allSortedKeys)-1]
			}
			suffix = jr.RenderRight(originalLastColIdx)
		}
		if prefix != "" || suffix != "" {
			line.WriteString(prefix + suffix + tw.NewLine)
			fmt.Fprint(w, line.String())
		}
		f.debug("Line: Handled empty row/widths case (no effective keys)")
		return
	}

	// Render left border
	if f.config.Borders.Left.Enabled() {
		line.WriteString(jr.RenderLeft())
	}

	// Render segments and junctions
	for keyIndex, currentColIdx := range effectiveKeys {
		jr.colIdx = currentColIdx
		segment := jr.GetSegment()
		colWidth := keyWidthMap[currentColIdx]
		f.debug("Line: Drawing segment for Effective colIdx=%d, segment='%s', width=%d", currentColIdx, segment, colWidth)

		if segment == "" {
			line.WriteString(strings.Repeat(" ", colWidth))
		} else {
			segmentWidth := twfn.DisplayWidth(segment)
			if segmentWidth <= 0 {
				segmentWidth = 1
			}
			repeat := 0
			if colWidth > 0 && segmentWidth > 0 {
				repeat = colWidth / segmentWidth
			}
			drawnSegment := strings.Repeat(segment, repeat)
			line.WriteString(drawnSegment)

			actualDrawnWidth := twfn.DisplayWidth(drawnSegment)
			if actualDrawnWidth < colWidth {
				missingWidth := colWidth - actualDrawnWidth
				spaces := strings.Repeat(" ", missingWidth)
				if len(f.config.Border.BG) > 0 {
					line.WriteString(Tint{BG: f.config.Border.BG}.Apply(spaces))
				} else {
					line.WriteString(spaces)
				}
				f.debug("Line: colIdx=%d corrected segment width, added %d spaces", currentColIdx, missingWidth)
			} else if actualDrawnWidth > colWidth {
				f.debug("Line: WARNING colIdx=%d segment draw width %d > target %d", currentColIdx, actualDrawnWidth, colWidth)
			}
		}

		// Render junction
		isLastVisible := keyIndex == len(effectiveKeys)-1
		if !isLastVisible && f.config.Settings.Separators.BetweenColumns.Enabled() {
			nextVisibleColIdx := effectiveKeys[keyIndex+1]
			originalPrecedingCol := -1
			foundCurrent := false
			for _, k := range allSortedKeys {
				if k == currentColIdx {
					foundCurrent = true
				}
				if foundCurrent && k < nextVisibleColIdx {
					originalPrecedingCol = k
				}
				if k >= nextVisibleColIdx {
					break
				}
			}

			if originalPrecedingCol != -1 {
				jr.colIdx = originalPrecedingCol
				junction := jr.RenderJunction(originalPrecedingCol, nextVisibleColIdx)
				f.debug("Line: Junction between visible %d (orig preceding %d) and next visible %d: '%s'", currentColIdx, originalPrecedingCol, nextVisibleColIdx, junction)
				line.WriteString(junction)
			} else {
				f.debug("Line: Could not determine original preceding column for junction before visible %d", nextVisibleColIdx)
				line.WriteString(f.config.Separator.Apply(jr.sym.Center()))
			}
		}
	}

	// Render right border
	if f.config.Borders.Right.Enabled() {
		originalLastColIdx := -1
		if len(allSortedKeys) > 0 {
			originalLastColIdx = allSortedKeys[len(allSortedKeys)-1]
		}
		jr.colIdx = originalLastColIdx
		line.WriteString(jr.RenderRight(originalLastColIdx))
	}

	line.WriteString(f.newLine)
	fmt.Fprint(w, line.String())
	f.debug("Line rendered: %s", strings.TrimSuffix(line.String(), f.newLine))
}
