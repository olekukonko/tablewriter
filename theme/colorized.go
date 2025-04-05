package theme

import (
	"fmt"
	"github.com/fatih/color"
	"github.com/olekukonko/tablewriter/symbols"
	"github.com/olekukonko/tablewriter/utils"
	"io"
	"os"
	"strings"
)

// ... (ALIGN_* constants unchanged) ...

// Colors represents color attributes from fatih/color
type Colors []color.Attribute

// Colorized implements colored ASCII table formatting
type Colorized struct {
	borders          Border
	centerSeparator  string
	rowSeparator     string
	columnSeparator  string
	headerAlignment  int
	footerAlignment  int
	alignment        int
	columnAlignments []int
	columnWidths     []int
	headerLine       bool
	newLine          string
	headerColors     []Colors
	columnColors     []Colors
	footerColors     []Colors
	syms             []string
	symbols          symbols.Symbols
	headerMaxWidth   int
	rowMaxWidth      int
	footerMaxWidth   int
	headerAutoWrap   bool
	rowAutoWrap      bool
	footerAutoWrap   bool
}

// ColorizedConfig holds configuration options for Colorized
type ColorizedConfig struct {
	Borders          Border
	HeaderAlignment  int
	FooterAlignment  int
	Alignment        int
	ColumnAlignments []int
	ColumnWidths     []int
	HeaderLine       bool
	CenterSeparator  string
	RowSeparator     string
	ColumnSeparator  string
	HeaderColors     []Colors
	ColumnColors     []Colors
	FooterColors     []Colors
	Symbols          symbols.Symbols
	HeaderMaxWidth   int
	RowMaxWidth      int
	FooterMaxWidth   int
	HeaderAutoWrap   bool
	RowAutoWrap      bool
	FooterAutoWrap   bool
}

// NewColorized creates a new Colorized formatter
func NewColorized(config ...ColorizedConfig) *Colorized {
	s := symbols.NewSymbols(symbols.StyleASCII)
	cfg := ColorizedConfig{
		Borders:         Border{Left: true, Right: true, Top: true, Bottom: true},
		HeaderAlignment: ALIGN_DEFAULT,
		FooterAlignment: ALIGN_DEFAULT,
		Alignment:       ALIGN_DEFAULT,
		HeaderLine:      true,
		Symbols:         s,
		HeaderAutoWrap:  true,
		RowAutoWrap:     true,
		FooterAutoWrap:  true,
	}
	if len(config) > 0 {
		cfg = config[0]
	}
	if cfg.Symbols == nil {
		cfg.Symbols = s
	}

	f := &Colorized{
		borders:          cfg.Borders,
		centerSeparator:  cfg.CenterSeparator,
		rowSeparator:     cfg.RowSeparator,
		columnSeparator:  cfg.ColumnSeparator,
		headerAlignment:  cfg.HeaderAlignment,
		footerAlignment:  cfg.FooterAlignment,
		alignment:        cfg.Alignment,
		columnAlignments: cfg.ColumnAlignments,
		columnWidths:     cfg.ColumnWidths,
		headerLine:       cfg.HeaderLine,
		newLine:          symbols.NEWLINE,
		headerColors:     cfg.HeaderColors,
		columnColors:     cfg.ColumnColors,
		footerColors:     cfg.FooterColors,
		syms:             simpleSyms(cfg.Symbols.Center(), cfg.Symbols.Row(), cfg.Symbols.Column()),
		symbols:          cfg.Symbols,
		headerMaxWidth:   cfg.HeaderMaxWidth,
		rowMaxWidth:      cfg.RowMaxWidth,
		footerMaxWidth:   cfg.FooterMaxWidth,
		headerAutoWrap:   cfg.HeaderAutoWrap,
		rowAutoWrap:      cfg.RowAutoWrap,
		footerAutoWrap:   cfg.FooterAutoWrap,
	}
	if f.centerSeparator == "" {
		f.centerSeparator = f.symbols.Center()
	}
	if f.rowSeparator == "" {
		f.rowSeparator = f.symbols.Row()
	}
	if f.columnSeparator == "" {
		f.columnSeparator = f.symbols.Column()
	}
	f.updateSymbols()
	fmt.Fprintf(os.Stderr, "Colorized Symbols: Center=%q, Row=%q, Column=%q\n", f.centerSeparator, f.rowSeparator, f.columnSeparator)
	return f
}

// FormatHeader renders the header row with multi-line support
func (f *Colorized) FormatHeader(w io.Writer, headers []string, colWidths map[int]int) {
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
		cells := make([]string, len(headers))
		for i := range headers {
			width := colWidths[i]
			if i < len(f.columnWidths) && f.columnWidths[i] > 0 {
				width = f.columnWidths[i]
			}
			var content string
			if lineIdx < len(splitHeaders[i]) {
				content = splitHeaders[i][lineIdx]
			}
			padFunc := f.pad(f.headerAlignment)
			cells[i] = padFunc(content, symbols.SPACE, width)
			if i < len(f.headerColors) && len(f.headerColors[i]) > 0 {
				c := color.New(f.headerColors[i]...).SprintFunc()
				cells[i] = c(cells[i])
			}
		}
		prefix := utils.ConditionString(f.borders.Left, f.columnSeparator, "")
		suffix := utils.ConditionString(f.borders.Right, f.columnSeparator, "")
		fmt.Fprintf(w, "%s%s%s%s", prefix, strings.Join(cells, f.columnSeparator), suffix, f.newLine)
	}
	if f.headerLine {
		f.FormatLine(w, colWidths, symbols.Middle)
	}
}

// FormatRow renders a data row with multi-line support
func (f *Colorized) FormatRow(w io.Writer, row []string, colWidths map[int]int, isFirstRow bool) {
	maxLines := 1
	splitRows := make([][]string, len(row))
	for i, cell := range row {
		lines := strings.Split(cell, "\n")
		splitRows[i] = lines
		if len(lines) > maxLines {
			maxLines = len(lines)
		}
	}

	for lineIdx := 0; lineIdx < maxLines; lineIdx++ {
		cells := make([]string, len(row))
		for i := range row {
			align := f.alignment
			if i < len(f.columnAlignments) && f.columnAlignments[i] != ALIGN_DEFAULT {
				align = f.columnAlignments[i]
			}
			width := colWidths[i]
			if i < len(f.columnWidths) && f.columnWidths[i] > 0 {
				width = f.columnWidths[i]
			}
			var content string
			if lineIdx < len(splitRows[i]) {
				content = splitRows[i][lineIdx]
			}
			padFunc := f.pad(align)
			cells[i] = padFunc(content, symbols.SPACE, width)
			if i < len(f.columnColors) && len(f.columnColors[i]) > 0 {
				c := color.New(f.columnColors[i]...).SprintFunc()
				cells[i] = c(cells[i])
			}
		}
		prefix := utils.ConditionString(f.borders.Left, f.columnSeparator, "")
		suffix := utils.ConditionString(f.borders.Right, f.columnSeparator, "")
		fmt.Fprintf(w, "%s%s%s%s", prefix, strings.Join(cells, f.columnSeparator), suffix, f.newLine)
	}
}

// FormatFooter renders the footer row with multi-line support
func (f *Colorized) FormatFooter(w io.Writer, footers []string, colWidths map[int]int) {
	maxLines := 1
	splitFooters := make([][]string, len(footers))
	for i, cell := range footers {
		lines := strings.Split(cell, "\n")
		splitFooters[i] = lines
		if len(lines) > maxLines {
			maxLines = len(lines)
		}
	}
	for lineIdx := 0; lineIdx < maxLines; lineIdx++ {
		cells := make([]string, len(footers))
		for i := range footers {
			width := colWidths[i]
			if i < len(f.columnWidths) && f.columnWidths[i] > 0 {
				width = f.columnWidths[i]
			}
			var content string
			if lineIdx < len(splitFooters[i]) {
				content = splitFooters[i][lineIdx]
			}
			padFunc := f.pad(f.footerAlignment)
			cells[i] = padFunc(content, symbols.SPACE, width)
			if i < len(f.footerColors) && len(f.footerColors[i]) > 0 {
				c := color.New(f.footerColors[i]...).SprintFunc()
				cells[i] = c(cells[i])
			}
		}
		prefix := utils.ConditionString(f.borders.Left, f.columnSeparator, "")
		suffix := utils.ConditionString(f.borders.Right, f.columnSeparator, "")
		fmt.Fprintf(w, "%s%s%s%s", prefix, strings.Join(cells, f.columnSeparator), suffix, f.newLine)
	}
}

// FormatLine renders a table line with proper symbols
func (f *Colorized) FormatLine(w io.Writer, colWidths map[int]int, lineType string) {
	var prefix, suffix, mid string
	switch lineType {
	case symbols.Top:
		prefix = utils.ConditionString(f.borders.Left, f.symbols.TopLeft(), f.rowSeparator)
		suffix = utils.ConditionString(f.borders.Right, f.symbols.TopRight(), "")
		mid = f.symbols.TopMid()
	case symbols.Bottom:
		prefix = utils.ConditionString(f.borders.Left, f.symbols.BottomLeft(), f.rowSeparator)
		suffix = utils.ConditionString(f.borders.Right, f.symbols.BottomRight(), "")
		mid = f.symbols.BottomMid()
	case symbols.Middle:
		prefix = utils.ConditionString(f.borders.Left, f.symbols.MidLeft(), f.rowSeparator)
		suffix = utils.ConditionString(f.borders.Right, f.symbols.MidRight(), "")
		mid = f.centerSeparator
	}

	line := prefix
	for i := 0; i < len(colWidths); i++ {
		width := colWidths[i]
		if i < len(f.columnWidths) && f.columnWidths[i] > 0 {
			width = f.columnWidths[i]
		}
		if i > 0 {
			line += mid
		}
		line += strings.Repeat(f.rowSeparator, width)
	}
	line += suffix
	fmt.Fprintf(w, "%s%s", line, f.newLine)
}

func (f *Colorized) GetColumnWidths() []int {
	return f.columnWidths
}

func (f *Colorized) Reset() {
	// No cache to reset with fatih/color
}

func (f *Colorized) updateSymbols() {
	f.centerSeparator = f.symbols.Center()
	f.rowSeparator = f.symbols.Row()
	f.columnSeparator = f.symbols.Column()
	f.syms = simpleSyms(f.centerSeparator, f.rowSeparator, f.columnSeparator)
}

func (f *Colorized) pad(align int) func(string, string, int) string {
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

func simpleSyms(center, row, column string) []string {
	return []string{row, column, center, center, center, center, center, center, center, center, center}
}
