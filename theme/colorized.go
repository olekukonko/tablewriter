// formatter/color.go
package theme

import (
	"fmt"
	"github.com/olekukonko/tablewriter/utils"
	"io"
	"strings"
	"sync"

	"github.com/olekukonko/tablewriter/symbols"
)

// Colorized implements colored ASCII table formatting
type Colorized struct {
	borders         Border
	centerSeparator string
	rowSeparator    string
	columnSeparator string
	headerAlignment int
	footerAlignment int
	alignment       int
	headerLine      bool
	newLine         string
	headerColors    []Colors
	columnColors    []Colors
	footerColors    []Colors
	syms            []string
	symbols         symbols.Symbols
	cache           colorCache
}

// Colors represents ANSI color codes
type Colors []int

const (
	BgBlackColor int = iota + 40
	BgRedColor
	BgGreenColor
	BgYellowColor
	BgBlueColor
	BgMagentaColor
	BgCyanColor
	BgWhiteColor
)

const (
	FgBlackColor int = iota + 30
	FgRedColor
	FgGreenColor
	FgYellowColor
	FgBlueColor
	FgMagentaColor
	FgCyanColor
	FgWhiteColor
)

const (
	BgHiBlackColor int = iota + 100
	BgHiRedColor
	BgHiGreenColor
	BgHiYellowColor
	BgHiBlueColor
	BgHiMagentaColor
	BgHiCyanColor
	BgHiWhiteColor
)

const (
	FgHiBlackColor int = iota + 90
	FgHiRedColor
	FgHiGreenColor
	FgHiYellowColor
	FgHiBlueColor
	FgHiMagentaColor
	FgHiCyanColor
	FgHiWhiteColor
)

const (
	Normal          = 0
	Bold            = 1
	UnderlineSingle = 4
	Italic
)

func NewColorFormatter() Structure {
	s := symbols.Default{}
	return &Colorized{
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
	}
}

func (f *Colorized) FormatHeader(w io.Writer, headers []string, colWidths map[int]int) {
	padFunc := f.pad(f.headerAlignment)
	cells := f.padCells(headers, padFunc, colWidths)
	for i, cell := range cells {
		if i < len(f.headerColors) && len(f.headerColors[i]) > 0 {
			cells[i] = f.format(cell, f.headerColors[i])
		}
	}
	prefix := utils.ConditionString(f.borders.Left, f.columnSeparator, symbols.SPACE)
	suffix := utils.ConditionString(f.borders.Right, f.columnSeparator, symbols.SPACE)
	fmt.Fprintf(w, "%s %s %s%s", prefix, strings.Join(cells, " "+f.columnSeparator+" "), suffix, f.newLine)
	if f.headerLine {
		f.FormatLine(w, colWidths, false)
	}
}

func (f *Colorized) FormatRow(w io.Writer, row []string, colWidths map[int]int, isFirstRow bool) {
	padFunc := f.pad(f.alignment)
	cells := f.padCells(row, padFunc, colWidths)
	for i, cell := range cells {
		if i < len(f.columnColors) && len(f.columnColors[i]) > 0 {
			cells[i] = f.format(cell, f.columnColors[i])
		}
	}
	prefix := utils.ConditionString(f.borders.Left, f.columnSeparator, symbols.SPACE)
	suffix := utils.ConditionString(f.borders.Right, f.columnSeparator, symbols.SPACE)
	fmt.Fprintf(w, "%s %s %s%s", prefix, strings.Join(cells, " "+f.columnSeparator+" "), suffix, f.newLine)
}

func (f *Colorized) FormatFooter(w io.Writer, footers []string, colWidths map[int]int) {
	padFunc := f.pad(f.footerAlignment)
	cells := f.padCells(footers, padFunc, colWidths)
	for i, cell := range cells {
		if i < len(f.footerColors) && len(f.footerColors[i]) > 0 {
			cells[i] = f.format(cell, f.footerColors[i])
		}
	}
	prefix := utils.ConditionString(f.borders.Left, f.columnSeparator, symbols.SPACE)
	suffix := utils.ConditionString(f.borders.Right, f.columnSeparator, symbols.SPACE)
	fmt.Fprintf(w, "%s %s %s%s", prefix, strings.Join(cells, " "+f.columnSeparator+" "), suffix, f.newLine)
}

func (f *Colorized) FormatLine(w io.Writer, colWidths map[int]int, isTop bool) {
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

func (f *Colorized) Configure(opt Option) {
	opt(f)
}

func (f *Colorized) Reset() {
	f.cache = colorCache{}
}

func (f *Colorized) padCells(cells []string, padFunc func(string, string, int) string, colWidths map[int]int) []string {
	padded := make([]string, len(cells))
	for i, cell := range cells {
		w := colWidths[i]
		padded[i] = padFunc(cell, symbols.SPACE, w)
	}
	return padded
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

func (f *Colorized) format(s string, codes Colors) string {
	if len(codes) == 0 {
		return s
	}
	key := fmt.Sprintf("%s|%v", s, codes)
	if cached, ok := f.cache.Get(key); ok {
		return cached
	}
	seq := f.makeSequence(codes)
	formatted := f.startFormat(seq) + s + f.stopFormat()
	f.cache.Store(key, formatted)
	return formatted
}

func (f *Colorized) makeSequence(codes []int) string {
	codesInString := make([]string, len(codes))
	for i, code := range codes {
		codesInString[i] = fmt.Sprintf("%d", code)
	}
	return strings.Join(codesInString, ";")
}

func (f *Colorized) startFormat(seq string) string {
	return fmt.Sprintf("\033[%sm", seq)
}

func (f *Colorized) stopFormat() string {
	return fmt.Sprintf("\033[%dm", Normal)
}

func (f *Colorized) updateSymbols() {
	f.centerSeparator = f.symbols.Center()
	f.rowSeparator = f.symbols.Row()
	f.columnSeparator = f.symbols.Column()
	f.syms = simpleSyms(f.centerSeparator, f.rowSeparator, f.columnSeparator)
}

// colorCache caches formatted strings for performance
type colorCache struct {
	sync.Map
}

func (c *colorCache) Get(key string) (string, bool) {
	if v, ok := c.Load(key); ok {
		return v.(string), true
	}
	return "", false
}

func (c *colorCache) Store(key, value string) {
	c.Map.Store(key, value)
}
