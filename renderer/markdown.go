package renderer

import (
	"fmt"
	"github.com/olekukonko/tablewriter/utils"
	"io"
	"strings"

	"github.com/olekukonko/tablewriter/symbols"
)

// Markdown implements a simple Markdown table formatter
type Markdown struct{}

// NewMarkdown creates a new Markdown renderer
func NewMarkdown() *Markdown {
	return &Markdown{}
}

func (m *Markdown) formatCell(content string, width int, align string, padding symbols.Padding) string {
	content = strings.TrimSpace(content)
	runeWidth := utils.RuneWidth(content)
	padLeftWidth := utils.RuneWidth(padding.Left)
	padRightWidth := utils.RuneWidth(padding.Right)
	totalContentWidth := runeWidth + padLeftWidth + padRightWidth

	if totalContentWidth > width {
		width = totalContentWidth
	}

	gap := width - runeWidth - padLeftWidth - padRightWidth
	if gap < 0 {
		gap = 0
	}

	leftPadChar := " "
	if padding.Left != "" {
		leftPadChar = padding.Left
	}
	rightPadChar := " "
	if padding.Right != "" {
		rightPadChar = padding.Right
	}

	switch align {
	case AlignCenter:
		leftGap := gap / 2
		rightGap := gap - leftGap
		return padding.Left + strings.Repeat(leftPadChar, leftGap) + content + strings.Repeat(rightPadChar, rightGap) + padding.Right
	case AlignRight:
		return padding.Left + strings.Repeat(leftPadChar, gap) + content + padding.Right
	default: // Left or default
		return padding.Left + content + strings.Repeat(rightPadChar, gap) + padding.Right
	}
}

func (m *Markdown) FormatHeader(w io.Writer, headers []string, ctx Context) {
	fmt.Printf("DEBUG: headers = %v\n", headers) // Add this for debugging
	// Format header cells with padding
	cells := make([]string, len(headers))
	for i, h := range headers {
		width := ctx.Widths[i]
		align := ctx.Align
		if colAlign, ok := ctx.ColAligns[i]; ok {
			align = colAlign
		}
		padding := ctx.Padding
		if customPad, ok := ctx.ColPadding[i]; ok {
			padding = customPad
		}
		cells[i] = m.formatCell(h, width, align, padding)
	}
	fmt.Fprintf(w, "|%s|\n", strings.Join(cells, "|"))

	// Format separator row
	separators := make([]string, len(headers))
	for i := range headers {
		width := ctx.Widths[i]
		align := ctx.Align
		if colAlign, ok := ctx.ColAligns[i]; ok {
			align = colAlign
		}

		switch align {
		case AlignCenter:
			separators[i] = ":" + strings.Repeat("-", width-2) + ":"
		case AlignRight:
			separators[i] = strings.Repeat("-", width-1) + ":"
		default: // Left or default
			separators[i] = ":" + strings.Repeat("-", width-1)
		}
	}
	fmt.Fprintf(w, "|%s|\n", strings.Join(separators, "|"))
}

func (m *Markdown) FormatRow(w io.Writer, row []string, ctx Context) {
	cells := make([]string, len(row))
	for i, cell := range row {
		width := ctx.Widths[i]
		align := ctx.Align
		if colAlign, ok := ctx.ColAligns[i]; ok {
			align = colAlign
		}
		padding := ctx.Padding
		if customPad, ok := ctx.ColPadding[i]; ok {
			padding = customPad
		}
		cells[i] = m.formatCell(cell, width, align, padding)
	}
	fmt.Fprintf(w, "|%s|\n", strings.Join(cells, "|"))
}

func (m *Markdown) FormatFooter(w io.Writer, footers []string, ctx Context) {
	cells := make([]string, len(footers))
	for i, f := range footers {
		width := ctx.Widths[i]
		align := ctx.Align
		if colAlign, ok := ctx.ColAligns[i]; ok {
			align = colAlign
		}
		padding := ctx.Padding
		if customPad, ok := ctx.ColPadding[i]; ok {
			padding = customPad
		}
		cells[i] = m.formatCell(f, width, align, padding)
	}
	fmt.Fprintf(w, "|%s|\n", strings.Join(cells, "|"))
}

func (m *Markdown) FormatLine(w io.Writer, ctx Context) {
	// Markdown tables don't need additional lines
	return
}

func (m *Markdown) GetColumnWidths() []int {
	return nil
}

func (m *Markdown) Reset() {}
