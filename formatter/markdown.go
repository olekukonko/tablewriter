// formatter/markdown.go
package formatter

import (
	"fmt"
	"github.com/olekukonko/tablewriter/utils"
	"io"
	"strings"

	"github.com/olekukonko/tablewriter/symbols"
)

type MarkdownFormatter struct {
	alignment int
	newLine   string
}

func NewMarkdownFormatter() Formatter {
	return &MarkdownFormatter{
		alignment: ALIGN_DEFAULT,
		newLine:   symbols.NEWLINE,
	}
}

func (f *MarkdownFormatter) FormatHeader(w io.Writer, headers []string, colWidths map[int]int) {
	padFunc := f.pad(f.alignment)
	cells := f.padCells(headers, padFunc, colWidths)
	fmt.Fprintf(w, "| %s |%s", strings.Join(cells, " | "), f.newLine)
	f.FormatLine(w, colWidths, false)
}

func (f *MarkdownFormatter) FormatRow(w io.Writer, row []string, colWidths map[int]int, isFirstRow bool) {
	padFunc := f.pad(f.alignment)
	cells := f.padCells(row, padFunc, colWidths)
	fmt.Fprintf(w, "| %s |%s", strings.Join(cells, " | "), f.newLine)
}

func (f *MarkdownFormatter) FormatFooter(w io.Writer, footers []string, colWidths map[int]int) {
	padFunc := f.pad(f.alignment)
	cells := f.padCells(footers, padFunc, colWidths)
	fmt.Fprintf(w, "| %s |%s", strings.Join(cells, " | "), f.newLine)
}

func (f *MarkdownFormatter) FormatLine(w io.Writer, colWidths map[int]int, isTop bool) {
	var separators []string
	for i := 0; i < len(colWidths); i++ {
		w := colWidths[i]
		separators = append(separators, strings.Repeat("-", w))
	}
	fmt.Fprintf(w, "| %s |%s", strings.Join(separators, " | "), f.newLine)
}

func (f *MarkdownFormatter) Configure(opt Option) {
	opt(f)
}

func (f *MarkdownFormatter) Reset() {
	// No internal state to reset
}

func (f *MarkdownFormatter) padCells(cells []string, padFunc func(string, string, int) string, colWidths map[int]int) []string {
	padded := make([]string, len(cells))
	for i, cell := range cells {
		w := colWidths[i]
		padded[i] = padFunc(cell, symbols.SPACE, w)
	}
	return padded
}

func (f *MarkdownFormatter) pad(align int) func(string, string, int) string {
	switch align {
	case ALIGN_CENTER:
		return utils.Pad
	case ALIGN_RIGHT:
		return utils.PadLeft
	case ALIGN_LEFT:
		return utils.PadRight
	default:
		return utils.PadRight
	}
}
