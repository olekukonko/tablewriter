// formatter/default.go
package formatter

import (
	"fmt"
	"github.com/olekukonko/tablewriter/symbols"
	"github.com/olekukonko/tablewriter/utils"
	"io"
	"strings"
)

// DefaultFormatter implements classic ASCII table formatting
type DefaultFormatter struct {
	borders         Border
	centerSeparator string
	rowSeparator    string
	columnSeparator string
	headerAlignment int
	footerAlignment int
	alignment       int
	headerLine      bool
	newLine         string
	syms            []string
	symbols         symbols.Symbols
	autoFormat      bool
}

func NewDefaultFormatter() Formatter {
	s := symbols.DefaultSymbols{}
	return &DefaultFormatter{
		borders:         Border{Left: true, Right: true, Top: true, Bottom: true},
		centerSeparator: s.Center(),
		rowSeparator:    s.Row(),
		columnSeparator: s.Column(),
		headerAlignment: ALIGN_DEFAULT,
		footerAlignment: ALIGN_DEFAULT,
		alignment:       ALIGN_DEFAULT,
		headerLine:      true,
		newLine:         symbols.NEWLINE,
		syms:            simpleSyms(s.Center(), s.Row(), s.Column()),
		symbols:         s,
		autoFormat:      true,
	}
}

func (f *DefaultFormatter) FormatHeader(w io.Writer, headers []string, colWidths map[int]int) {
	padFunc := f.pad(f.headerAlignment)
	cells := make([]string, len(headers))
	copy(cells, headers)
	for i, h := range cells {
		if f.autoFormat {
			cells[i] = utils.Title(h)
		}
	}
	cells = f.padCells(cells, padFunc, colWidths)
	prefix := utils.ConditionString(f.borders.Left, f.columnSeparator, symbols.SPACE)
	suffix := utils.ConditionString(f.borders.Right, f.columnSeparator, symbols.SPACE)
	fmt.Fprintf(w, "%s %s %s%s", prefix, strings.Join(cells, " "+f.columnSeparator+" "), suffix, f.newLine)
	if f.headerLine {
		f.FormatLine(w, colWidths, false)
	}
}

func (f *DefaultFormatter) FormatRow(w io.Writer, row []string, colWidths map[int]int, isFirstRow bool) {
	padFunc := f.pad(f.alignment)
	cells := f.padCells(row, padFunc, colWidths)
	prefix := utils.ConditionString(f.borders.Left, f.columnSeparator, symbols.SPACE)
	suffix := utils.ConditionString(f.borders.Right, f.columnSeparator, symbols.SPACE)
	fmt.Fprintf(w, "%s %s %s%s", prefix, strings.Join(cells, " "+f.columnSeparator+" "), suffix, f.newLine)
}

func (f *DefaultFormatter) FormatFooter(w io.Writer, footers []string, colWidths map[int]int) {
	padFunc := f.pad(f.footerAlignment)
	cells := f.padCells(footers, padFunc, colWidths)
	prefix := utils.ConditionString(f.borders.Left, f.columnSeparator, symbols.SPACE)
	suffix := utils.ConditionString(f.borders.Right, f.columnSeparator, symbols.SPACE)
	fmt.Fprintf(w, "%s %s %s%s", prefix, strings.Join(cells, " "+f.columnSeparator+" "), suffix, f.newLine)
}

func (f *DefaultFormatter) FormatLine(w io.Writer, colWidths map[int]int, isTop bool) {
	prefix := f.centerSeparator
	if !f.borders.Left {
		prefix = f.rowSeparator
	}
	fmt.Fprint(w, prefix)
	for i := 0; i < len(colWidths); i++ {
		width := colWidths[i]
		fmt.Fprintf(w, "%s%s%s%s", f.rowSeparator, strings.Repeat(f.rowSeparator, width), f.rowSeparator, f.centerSeparator)
	}
	fmt.Fprint(w, f.newLine)
}

func (f *DefaultFormatter) Configure(opt Option) {
	opt(f)
}

func (f *DefaultFormatter) Reset() {
	// No internal state to reset
}

func (f *DefaultFormatter) padCells(cells []string, padFunc func(string, string, int) string, colWidths map[int]int) []string {
	padded := make([]string, len(cells))
	for i, cell := range cells {
		w := colWidths[i]
		padded[i] = padFunc(cell, symbols.SPACE, w)
	}
	return padded
}

func (f *DefaultFormatter) pad(align int) func(string, string, int) string {
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

func (f *DefaultFormatter) updateSymbols() {
	f.centerSeparator = f.symbols.Center()
	f.rowSeparator = f.symbols.Row()
	f.columnSeparator = f.symbols.Column()
	f.syms = simpleSyms(f.centerSeparator, f.rowSeparator, f.columnSeparator)
}
