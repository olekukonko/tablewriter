package renderer

//
//import (
//	"fmt"
//	"github.com/fatih/color"
//	"github.com/olekukonko/tablewriter/symbols"
//	"github.com/olekukonko/tablewriter/utils"
//	"io"
//	"os"
//	"strings"
//)
//
//// Colors represents color attributes from fatih/color
//type Colors []color.Attribute
//
//// Colorized implements colored ASCII table formatting
//type Colorized struct {
//	borders      Border
//	headerLine   bool
//	newLine      string
//	headerColors []Colors
//	columnColors []Colors
//	footerColors []Colors
//	symbols      symbols.Symbols
//}
//
//// ColorizedConfig holds configuration specific to the Colorized theme
//type ColorizedConfig struct {
//	Borders      Border
//	HeaderLine   bool
//	HeaderColors []Colors
//	ColumnColors []Colors
//	FooterColors []Colors
//	Symbols      symbols.Symbols
//}
//
//// defaultColorizedConfig returns a default configuration for the Colorized theme
//func defaultColorizedConfig() ColorizedConfig {
//	s := symbols.NewSymbols(symbols.StyleASCII)
//	return ColorizedConfig{
//		Borders:      Border{Left: true, Right: true, Top: true, Bottom: true},
//		HeaderLine:   true,
//		HeaderColors: nil,
//		ColumnColors: nil,
//		FooterColors: nil,
//		Symbols:      s,
//	}
//}
//
//// NewColorized creates a new Colorized formatter
//func NewColorized(config ...ColorizedConfig) *Colorized {
//	cfg := defaultColorizedConfig()
//	if len(config) > 0 {
//		cfg = config[0]
//	}
//	if cfg.Symbols == nil {
//		cfg.Symbols = symbols.NewSymbols(symbols.StyleASCII)
//	}
//
//	f := &Colorized{
//		borders:      cfg.Borders,
//		headerLine:   cfg.HeaderLine,
//		newLine:      symbols.NEWLINE,
//		headerColors: cfg.HeaderColors,
//		columnColors: cfg.ColumnColors,
//		footerColors: cfg.FooterColors,
//		symbols:      cfg.Symbols,
//	}
//	fmt.Fprintf(os.Stderr, "Colorized Symbols: Center=%q, Row=%q, Column=%q\n", f.symbols.Center(), f.symbols.Row(), f.symbols.Column())
//	return f
//}
//
//// Header renders the header row with color
//func (f *Colorized) Header(w io.Writer, headers []string, colWidths map[int]int) {
//	maxLines := 1
//	splitHeaders := make([][]string, len(headers))
//	for i, h := range headers {
//		lines := strings.Split(h, "\n")
//		splitHeaders[i] = lines
//		if len(lines) > maxLines {
//			maxLines = len(lines)
//		}
//	}
//
//	for lineIdx := 0; lineIdx < maxLines; lineIdx++ {
//		cells := make([]string, len(headers))
//		for i := range headers {
//			width := colWidths[i]
//			var content string
//			if lineIdx < len(splitHeaders[i]) {
//				content = splitHeaders[i][lineIdx]
//			}
//			cells[i] = utils.Pad(content, symbols.SPACE, width) // Center alignment for headers
//			if i < len(f.headerColors) && len(f.headerColors[i]) > 0 {
//				c := color.New(f.headerColors[i]...).SprintFunc()
//				cells[i] = c(cells[i])
//			}
//		}
//		prefix := utils.ConditionString(f.borders.Left, f.symbols.Column(), "")
//		suffix := utils.ConditionString(f.borders.Right, f.symbols.Column(), "")
//		fmt.Fprintf(w, "%s%s%s%s", prefix, strings.Join(cells, f.symbols.Column()), suffix, f.newLine)
//	}
//	if f.headerLine {
//		f.Lines(w, colWidths, Row)
//	}
//}
//
//// Row renders a data row with color
//func (f *Colorized) Row(w io.Writer, row []string, colWidths map[int]int, isFirstRow bool) {
//	maxLines := 1
//	splitRows := make([][]string, len(row))
//	for i, cell := range row {
//		lines := strings.Split(cell, "\n")
//		splitRows[i] = lines
//		if len(lines) > maxLines {
//			maxLines = len(lines)
//		}
//	}
//
//	for lineIdx := 0; lineIdx < maxLines; lineIdx++ {
//		cells := make([]string, len(row))
//		for i := range row {
//			width := colWidths[i]
//			var content string
//			if lineIdx < len(splitRows[i]) {
//				content = splitRows[i][lineIdx]
//			}
//			cells[i] = utils.PadRight(content, symbols.SPACE, width) // Left alignment for rows
//			if i < len(f.columnColors) && len(f.columnColors[i]) > 0 {
//				c := color.New(f.columnColors[i]...).SprintFunc()
//				cells[i] = c(cells[i])
//			}
//		}
//		prefix := utils.ConditionString(f.borders.Left, f.symbols.Column(), "")
//		suffix := utils.ConditionString(f.borders.Right, f.symbols.Column(), "")
//		fmt.Fprintf(w, "%s%s%s%s", prefix, strings.Join(cells, f.symbols.Column()), suffix, f.newLine)
//	}
//}
//
//// Footer renders the footer row with color
//func (f *Colorized) Footer(w io.Writer, footers []string, colWidths map[int]int) {
//	maxLines := 1
//	splitFooters := make([][]string, len(footers))
//	for i, cell := range footers {
//		lines := strings.Split(cell, "\n")
//		splitFooters[i] = lines
//		if len(lines) > maxLines {
//			maxLines = len(lines)
//		}
//	}
//	for lineIdx := 0; lineIdx < maxLines; lineIdx++ {
//		cells := make([]string, len(footers))
//		for i := range footers {
//			width := colWidths[i]
//			var content string
//			if lineIdx < len(splitFooters[i]) {
//				content = splitFooters[i][lineIdx]
//			}
//			if utils.IsNumeric(strings.TrimSpace(footers[i])) {
//				cells[i] = utils.PadLeft(content, symbols.SPACE, width) // Right alignment for numeric footers
//			} else {
//				cells[i] = utils.PadRight(content, symbols.SPACE, width) // Left alignment otherwise
//			}
//			if i < len(f.footerColors) && len(f.footerColors[i]) > 0 {
//				c := color.New(f.footerColors[i]...).SprintFunc()
//				cells[i] = c(cells[i])
//			}
//		}
//		prefix := utils.ConditionString(f.borders.Left, f.symbols.Column(), "")
//		suffix := utils.ConditionString(f.borders.Right, f.symbols.Column(), "")
//		fmt.Fprintf(w, "%s%s%s%s", prefix, strings.Join(cells, f.symbols.Column()), suffix, f.newLine)
//	}
//}
//
//// Lines renders a table line with proper symbols
//func (f *Colorized) Lines(w io.Writer, colWidths map[int]int, lineType string) {
//	var prefix, suffix, mid string
//	switch lineType {
//	case Header:
//		prefix = utils.ConditionString(f.borders.Left, f.symbols.TopLeft(), f.symbols.Row())
//		suffix = utils.ConditionString(f.borders.Right, f.symbols.TopRight(), "")
//		mid = f.symbols.TopMid()
//	case Footer:
//		prefix = utils.ConditionString(f.borders.Left, f.symbols.BottomLeft(), f.symbols.Row())
//		suffix = utils.ConditionString(f.borders.Right, f.symbols.BottomRight(), "")
//		mid = f.symbols.BottomMid()
//	case Row:
//		prefix = utils.ConditionString(f.borders.Left, f.symbols.MidLeft(), f.symbols.Row())
//		suffix = utils.ConditionString(f.borders.Right, f.symbols.MidRight(), "")
//		mid = f.symbols.Center()
//	}
//
//	line := prefix
//	for i := 0; i < len(colWidths); i++ {
//		width := colWidths[i]
//		if i > 0 {
//			line += mid
//		}
//		line += strings.Repeat(f.symbols.Row(), width)
//	}
//	line += suffix
//	fmt.Fprintf(w, "%s%s", line, f.newLine)
//}
//
//func (f *Colorized) GetColumnWidths() []int {
//	return nil // Colorized doesn’t enforce column widths; tablewriter handles this
//}
//
//func (f *Colorized) Reset() {
//	// No state to reset
//}
