package theme

import (
	"fmt"
	"github.com/olekukonko/tablewriter/symbols"
	"github.com/olekukonko/tablewriter/utils"
	"io"
	"os"
	"strings"
)

// Default implements classic ASCII table formatting
type Default struct {
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
	syms             []string
	symbols          symbols.Symbols
	autoFormat       bool
}

// DefaultConfig holds configuration options for Default
type DefaultConfig struct {
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
	AutoFormat       bool            // Whether to auto-format headers (e.g., capitalize)
	Symbols          symbols.Symbols // Symbol set (e.g., ASCII, Unicode)
}

// NewDefault creates a new Default with the given configuration
func NewDefault(config ...DefaultConfig) *Default {
	s := symbols.Default{}
	// Default configuration
	cfg := DefaultConfig{
		Borders:         Border{Left: true, Right: true, Top: true, Bottom: true},
		HeaderAlignment: ALIGN_DEFAULT,
		FooterAlignment: ALIGN_DEFAULT,
		Alignment:       ALIGN_DEFAULT,
		HeaderLine:      true,
		AutoFormat:      true,
		Symbols:         s,
	}
	if len(config) > 0 {
		cfg = config[0] // Use provided config if present
	}
	// Ensure symbols are set even if nil in config
	if cfg.Symbols == nil {
		cfg.Symbols = s
	}

	f := &Default{
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
		syms:             simpleSyms(cfg.Symbols.Center(), cfg.Symbols.Row(), cfg.Symbols.Column()),
		symbols:          cfg.Symbols,
		autoFormat:       cfg.AutoFormat,
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
	fmt.Fprintf(os.Stderr, "Default Symbols: Center=%q, Row=%q, Column=%q\n", f.centerSeparator, f.rowSeparator, f.columnSeparator)
	return f
}

func (f *Default) FormatHeader(w io.Writer, headers []string, colWidths map[int]int) {
	padFunc := f.pad(f.headerAlignment)
	cells := make([]string, len(headers))
	copy(cells, headers)
	for i, h := range cells {
		if f.autoFormat {
			cells[i] = utils.Title(h)
		}
		width := colWidths[i]
		if i < len(f.columnWidths) && f.columnWidths[i] > 0 {
			width = f.columnWidths[i]
		}
		cells[i] = padFunc(cells[i], symbols.SPACE, width)
	}
	prefix := utils.ConditionString(f.borders.Left, f.columnSeparator, symbols.SPACE)
	suffix := utils.ConditionString(f.borders.Right, f.columnSeparator, symbols.SPACE)
	fmt.Fprintf(w, "%s %s %s%s", prefix, strings.Join(cells, " "+f.columnSeparator+" "), suffix, f.newLine)
	if f.headerLine {
		f.FormatLine(w, colWidths, false)
	}
}

func (f *Default) FormatRow(w io.Writer, row []string, colWidths map[int]int, isFirstRow bool) {
	cells := make([]string, len(row))
	for i, cell := range row {
		align := f.alignment
		if i < len(f.columnAlignments) && f.columnAlignments[i] != ALIGN_DEFAULT {
			align = f.columnAlignments[i]
		}
		width := colWidths[i]
		if i < len(f.columnWidths) && f.columnWidths[i] > 0 {
			width = f.columnWidths[i]
		}
		padFunc := f.pad(align)
		cells[i] = padFunc(cell, symbols.SPACE, width)
	}
	prefix := utils.ConditionString(f.borders.Left, f.columnSeparator, symbols.SPACE)
	suffix := utils.ConditionString(f.borders.Right, f.columnSeparator, symbols.SPACE)
	fmt.Fprintf(w, "%s %s %s%s", prefix, strings.Join(cells, " "+f.columnSeparator+" "), suffix, f.newLine)
}

func (f *Default) FormatFooter(w io.Writer, footers []string, colWidths map[int]int) {
	padFunc := f.pad(f.footerAlignment)
	cells := make([]string, len(footers))
	for i, cell := range footers {
		width := colWidths[i]
		if i < len(f.columnWidths) && f.columnWidths[i] > 0 {
			width = f.columnWidths[i]
		}
		cells[i] = padFunc(cell, symbols.SPACE, width)
	}
	prefix := utils.ConditionString(f.borders.Left, f.columnSeparator, symbols.SPACE)
	suffix := utils.ConditionString(f.borders.Right, f.columnSeparator, symbols.SPACE)
	fmt.Fprintf(w, "%s %s %s%s", prefix, strings.Join(cells, " "+f.columnSeparator+" "), suffix, f.newLine)
}

func (f *Default) FormatLine(w io.Writer, colWidths map[int]int, isTop bool) {
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

func (f *Default) GetColumnWidths() []int {
	return f.columnWidths
}

func (f *Default) Reset() {
	// No internal state to reset
}

func (f *Default) pad(align int) func(string, string, int) string {
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

func (f *Default) updateSymbols() {
	f.syms = simpleSyms(f.centerSeparator, f.rowSeparator, f.columnSeparator)
}

// simpleSyms generates a basic symbol set
func simpleSyms(center, row, column string) []string {
	return []string{row, column, center, center, center, center, center, center, center, center, center}
}
