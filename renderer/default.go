package renderer

import (
	"fmt"
	"io"
	"strings"

	"github.com/olekukonko/tablewriter/symbols"
	"github.com/olekukonko/tablewriter/utils"
)

// DefaultConfig holds configuration specific to the Default theme
type DefaultConfig struct {
	Borders    Border          // Defines which borders to draw (left, right, top, bottom)
	HeaderLine bool            // Indicates whether to draw a line after the header
	Symbols    symbols.Symbols // Symbols used for table borders and separators
}

// Default implements classic table formatting
type Default struct {
	config DefaultConfig // Configuration for the Default theme
}

// --- Private Methods ---

// defaultConfig returns a default configuration for the Default theme
func defaultConfig() DefaultConfig {
	s := symbols.NewSymbols(symbols.StyleASCII)
	return DefaultConfig{
		Borders:    Border{Left: true, Right: true, Top: true, Bottom: true},
		HeaderLine: true,
		Symbols:    s,
	}
}

// formatCell formats a single cell with alignment and padding
// Returns the formatted string with content aligned and padded as specified
func (f *Default) formatCell(content string, width int, defaultAlign string, colAlign string, padding symbols.Padding) string {
	// Determine alignment: column-specific overrides default
	align := defaultAlign
	if colAlign != "" {
		align = colAlign
	}

	// Trim content and calculate widths
	content = strings.TrimSpace(content)
	runeWidth := utils.RuneWidth(content)
	padLeftWidth := utils.RuneWidth(padding.Left)
	padRightWidth := utils.RuneWidth(padding.Right)
	totalContentWidth := runeWidth + padLeftWidth + padRightWidth

	// Ensure width accommodates content and padding
	if totalContentWidth > width {
		width = totalContentWidth
	}

	// Calculate remaining gap
	gap := width - runeWidth - padLeftWidth - padRightWidth
	if gap < 0 {
		gap = 0
	}

	// Use padding characters for gaps, falling back to spaces if empty
	leftPadChar := " "
	if padding.Left != "" {
		leftPadChar = padding.Left
	}
	rightPadChar := " "
	if padding.Right != "" {
		rightPadChar = padding.Right
	}

	// Apply alignment
	switch align {
	case AlignCenter:
		leftGap := gap / 2
		rightGap := gap - leftGap
		return padding.Left + strings.Repeat(leftPadChar, leftGap) + content + strings.Repeat(rightPadChar, rightGap) + padding.Right
	case AlignRight:
		return padding.Left + strings.Repeat(leftPadChar, gap) + content + padding.Right
	default: // Left/Default
		return padding.Left + content + strings.Repeat(rightPadChar, gap) + padding.Right
	}
}

// formatSection renders a section (row or footer) with multi-line support
// Writes formatted cells to the writer, handling rows or footers based on isFooter flag
func (f *Default) formatSection(w io.Writer, cells []string, ctx Context, isFooter bool) {
	maxLines := 1
	splitCells := make([][]string, len(cells))
	for i, cell := range cells {
		lines := strings.Split(cell, "\n")
		splitCells[i] = lines
		if len(lines) > maxLines {
			maxLines = len(lines)
		}
	}

	for lineIdx := 0; lineIdx < maxLines; lineIdx++ {
		var renderedCells []string
		for i := range cells {
			width := ctx.Widths[i]
			content := ""
			if lineIdx < len(splitCells[i]) {
				content = splitCells[i][lineIdx]
			}

			// Apply custom padding if specified
			padding := ctx.Padding
			if customPad, ok := ctx.ColPadding[i]; ok {
				padding = customPad
			}

			// Apply column-specific alignment if specified
			align := ctx.Align
			if colAlign, ok := ctx.ColAligns[i]; ok {
				align = colAlign
			}

			renderedCells = append(renderedCells, f.formatCell(content, width, align, "", padding))
		}
		f.renderLine(w, renderedCells, ctx)
	}
}

// renderLine renders a single line of cells to the writer
// Handles Markdown-specific rendering and bordered table formatting
func (f *Default) renderLine(w io.Writer, cells []string, ctx Context) {
	// Special case for Markdown: always use "|" as column separator
	if f.config.Symbols.Name() == symbols.NameMarkdown {
		fmt.Fprintf(w, "|%s|%s", strings.Join(cells, "|"), symbols.NewLine)
		return
	}

	// Determine prefix and suffix based on border settings
	prefix := f.config.Symbols.Column()
	if !f.config.Borders.Left {
		prefix = ""
	}
	suffix := f.config.Symbols.Column()
	if !f.config.Borders.Right {
		suffix = ""
	}
	fmt.Fprintf(w, "%s%s%s%s", prefix, strings.Join(cells, f.config.Symbols.Column()), suffix, symbols.NewLine)
}

// --- Public Methods ---

// FormatFooter renders the footer section of the table
// Writes multi-line footer content with proper alignment and padding
func (f *Default) FormatFooter(w io.Writer, footers []string, ctx Context) {
	maxLines := 1
	splitFooters := make([][]string, len(footers))
	for i, footer := range footers {
		lines := strings.Split(footer, "\n")
		splitFooters[i] = lines
		if len(lines) > maxLines {
			maxLines = len(lines)
		}
	}

	for lineIdx := 0; lineIdx < maxLines; lineIdx++ {
		var cells []string
		for i := range footers {
			width := ctx.Widths[i]
			content := ""
			if lineIdx < len(splitFooters[i]) {
				content = splitFooters[i][lineIdx]
			}
			padding := ctx.Padding
			if customPad, ok := ctx.ColPadding[i]; ok {
				padding = customPad
			}
			cells = append(cells, f.formatCell(content, width, ctx.Align, ctx.ColAligns[i], padding))
		}
		f.renderLine(w, cells, ctx)
	}
}

// FormatHeader renders the header section of the table
// Writes multi-line headers with an optional separator line if configured
func (f *Default) FormatHeader(w io.Writer, headers []string, ctx Context) {
	maxLines := 1
	splitHeaders := make([][]string, len(headers))
	for i, h := range headers {
		lines := strings.Split(h, "\n")
		splitHeaders[i] = lines
		if len(lines) > maxLines {
			maxLines = len(lines)
		}
	}

	for lineIdx := 0; lineIdx < maxLines; lineIdx++ {
		var cells []string
		for i := range headers {
			width := ctx.Widths[i]
			content := ""
			if lineIdx < len(splitHeaders[i]) {
				content = splitHeaders[i][lineIdx]
			}
			padding := ctx.Padding
			if customPad, ok := ctx.ColPadding[i]; ok {
				padding = customPad
			}
			cells = append(cells, f.formatCell(content, width, ctx.Align, ctx.ColAligns[i], padding))
		}
		f.renderLine(w, cells, ctx)
	}
	if f.config.HeaderLine {
		f.FormatLine(w, Context{Widths: ctx.Widths, Level: Middle})
	}
}

// FormatLine renders a horizontal separator line
// Supports top, middle, and bottom lines with appropriate border symbols
func (f *Default) FormatLine(w io.Writer, ctx Context) {
	var line strings.Builder

	// Left border or column symbol
	if f.config.Borders.Left {
		switch ctx.Level {
		case Top:
			line.WriteString(f.config.Symbols.TopLeft())
		case Middle:
			line.WriteString(f.config.Symbols.MidLeft())
		case Bottom:
			line.WriteString(f.config.Symbols.BottomLeft())
		}
	} else {
		line.WriteString(f.config.Symbols.Column()) // "|" for Markdown
	}

	// Columns and separators
	for i, width := range ctx.Widths {
		if i > 0 {
			switch ctx.Level {
			case Top:
				if f.config.Borders.Top {
					line.WriteString(f.config.Symbols.TopMid())
				} else {
					line.WriteString(f.config.Symbols.Center())
				}
			case Middle:
				line.WriteString(f.config.Symbols.Center())
			case Bottom:
				if f.config.Borders.Bottom {
					line.WriteString(f.config.Symbols.BottomMid())
				} else {
					line.WriteString(f.config.Symbols.Center())
				}
			}
		}
		line.WriteString(strings.Repeat(f.config.Symbols.Row(), width))
	}

	// Right border or column symbol
	if f.config.Borders.Right {
		switch ctx.Level {
		case Top:
			line.WriteString(f.config.Symbols.TopRight())
		case Middle:
			line.WriteString(f.config.Symbols.MidRight())
		case Bottom:
			line.WriteString(f.config.Symbols.BottomRight())
		}
	} else {
		line.WriteString(f.config.Symbols.Column()) // "|" for Markdown
	}

	fmt.Fprintln(w, line.String())
}

// FormatRow renders a row of the table
// Delegates to formatSection for consistent row rendering
func (f *Default) FormatRow(w io.Writer, row []string, ctx Context) {
	f.formatSection(w, row, ctx, false)
}

// GetColumnWidths returns predefined column widths
// Returns nil as Default does not predefine widths
func (f *Default) GetColumnWidths() []int {
	return nil
}

// NewDefault creates a new Default theme renderer
// Accepts an optional configuration; uses default if none provided
func NewDefault(config ...DefaultConfig) *Default {
	cfg := defaultConfig()
	if len(config) > 0 {
		cfg = config[0]
	}
	if cfg.Symbols == nil {
		cfg.Symbols = symbols.NewSymbols(symbols.StyleASCII)
	}
	return &Default{config: cfg}
}

// Reset resets the renderer state
// No-op for Default as it maintains no mutable state
func (f *Default) Reset() {}
