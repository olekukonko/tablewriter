package theme

import (
	"fmt"
	"github.com/olekukonko/tablewriter/symbols"
	"github.com/olekukonko/tablewriter/utils"
	"io"
	"os"
	"strings"
	"sync"
)

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

// Colors represents ANSI color codes
type Colors []int

// Colorized implements colored ASCII table formatting
type Colorized struct {
	borders          Border
	centerSeparator  string
	rowSeparator     string
	columnSeparator  string
	headerAlignment  int
	footerAlignment  int
	alignment        int   // Default body alignment
	columnAlignments []int // Per-column body alignment overrides
	columnWidths     []int // Per-column width overrides
	headerLine       bool
	newLine          string
	headerColors     []Colors
	columnColors     []Colors
	footerColors     []Colors
	syms             []string
	symbols          symbols.Symbols
	cache            colorCache
}

// ColorizedConfig holds configuration options for Colorized
type ColorizedConfig struct {
	Borders          Border          // Table border settings
	HeaderAlignment  int             // Alignment for header cells
	FooterAlignment  int             // Alignment for footer cells
	Alignment        int             // Default alignment for body (rows)
	ColumnAlignments []int           // Per-column alignment overrides for body (rows)
	ColumnWidths     []int           // Per-column width overrides (0 means auto-calculate)
	HeaderLine       bool            // Whether to draw a line under the header
	CenterSeparator  string          // Custom center separator (e.g., "+")
	RowSeparator     string          // Custom row separator (e.g., "-")
	ColumnSeparator  string          // Custom column separator (e.g., "|")
	HeaderColors     []Colors        // Colors for header cells
	ColumnColors     []Colors        // Colors for body columns
	FooterColors     []Colors        // Colors for footer cells
	Symbols          symbols.Symbols // Symbol set (e.g., ASCII, Unicode)
}

// NewColorized creates a new Colorized formatter with the given configuration
func NewColorized(config ...ColorizedConfig) *Colorized {
	s := symbols.Default{}
	// Default configuration
	cfg := ColorizedConfig{
		Borders:         Border{Left: true, Right: true, Top: true, Bottom: true},
		HeaderAlignment: ALIGN_DEFAULT,
		FooterAlignment: ALIGN_DEFAULT,
		Alignment:       ALIGN_DEFAULT,
		HeaderLine:      true,
		Symbols:         s,
	}
	if len(config) > 0 {
		cfg = config[0] // Use provided config if present
	}
	// Ensure symbols are set even if nil in config
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
	}
	// Apply symbols correctly, overriding with custom separators if provided
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

	// Debug: Print symbols to verify
	fmt.Fprintf(os.Stderr, "Colorized Symbols: Center=%q, Row=%q, Column=%q\n", f.centerSeparator, f.rowSeparator, f.columnSeparator)
	return f
}

func (f *Colorized) FormatHeader(w io.Writer, headers []string, colWidths map[int]int) {
	padFunc := f.pad(f.headerAlignment)
	cells := make([]string, len(headers))
	for i, h := range headers {
		width := colWidths[i]
		if i < len(f.columnWidths) && f.columnWidths[i] > 0 {
			width = f.columnWidths[i] // Override with specified width
		}
		cells[i] = padFunc(h, symbols.SPACE, width)
		if i < len(f.headerColors) && len(f.headerColors[i]) > 0 {
			cells[i] = f.format(cells[i], f.headerColors[i])
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
	cells := make([]string, len(row))
	for i, cell := range row {
		align := f.alignment // Use body alignment
		if i < len(f.columnAlignments) && f.columnAlignments[i] != ALIGN_DEFAULT {
			align = f.columnAlignments[i] // Override with per-column body alignment
		}
		width := colWidths[i]
		if i < len(f.columnWidths) && f.columnWidths[i] > 0 {
			width = f.columnWidths[i] // Override with specified width
		}
		padFunc := f.pad(align)
		cells[i] = padFunc(cell, symbols.SPACE, width)
		if i < len(f.columnColors) && len(f.columnColors[i]) > 0 {
			cells[i] = f.format(cells[i], f.columnColors[i])
		}
	}
	prefix := utils.ConditionString(f.borders.Left, f.columnSeparator, symbols.SPACE)
	suffix := utils.ConditionString(f.borders.Right, f.columnSeparator, symbols.SPACE)
	fmt.Fprintf(w, "%s %s %s%s", prefix, strings.Join(cells, " "+f.columnSeparator+" "), suffix, f.newLine)
}

func (f *Colorized) FormatFooter(w io.Writer, footers []string, colWidths map[int]int) {
	padFunc := f.pad(f.footerAlignment)
	cells := make([]string, len(footers))
	for i, cell := range footers {
		width := colWidths[i]
		if i < len(f.columnWidths) && f.columnWidths[i] > 0 {
			width = f.columnWidths[i] // Override with specified width
		}
		cells[i] = padFunc(cell, symbols.SPACE, width)
		if i < len(f.footerColors) && len(f.footerColors[i]) > 0 {
			cells[i] = f.format(cells[i], f.footerColors[i])
		}
	}
	prefix := utils.ConditionString(f.borders.Left, f.columnSeparator, symbols.SPACE)
	suffix := utils.ConditionString(f.borders.Right, f.columnSeparator, symbols.SPACE)
	fmt.Fprintf(w, "%s %s %s%s", prefix, strings.Join(cells, " "+f.columnSeparator+" "), suffix, f.newLine)
}

// FormatLine renders a table line with proper symbols
func (f *Colorized) FormatLine(w io.Writer, colWidths map[int]int, isTop bool) {
	var prefix, suffix, mid string
	if isTop && f.borders.Top {
		prefix = utils.ConditionString(f.borders.Left, f.symbols.TopLeft(), f.rowSeparator)
		suffix = utils.ConditionString(f.borders.Right, f.symbols.TopRight(), "")
		mid = f.symbols.TopMid()
	} else if !isTop && f.borders.Bottom {
		prefix = utils.ConditionString(f.borders.Left, f.symbols.BottomLeft(), f.rowSeparator)
		suffix = utils.ConditionString(f.borders.Right, f.symbols.BottomRight(), "")
		mid = f.symbols.BottomMid()
	} else {
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
	f.cache = colorCache{}
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
