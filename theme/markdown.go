package theme

import (
	"fmt"
	"github.com/olekukonko/tablewriter/symbols"
	"github.com/olekukonko/tablewriter/utils"
	"io"
	"strings"
)

// Markdown implements Markdown table formatting
type Markdown struct {
	headerAlignment  int   // Alignment for header cells
	footerAlignment  int   // Alignment for footer cells
	alignment        int   // Default alignment for body (rows)
	columnAlignments []int // Per-column alignment overrides for body (rows)
	columnWidths     []int // Per-column width overrides
	newLine          string
}

// MarkdownConfig holds configuration options for Markdown
type MarkdownConfig struct {
	HeaderAlignment  int   // Alignment for header cells
	FooterAlignment  int   // Alignment for footer cells
	Alignment        int   // Default alignment for body (rows)
	ColumnAlignments []int // Per-column alignment overrides for body (rows)
	ColumnWidths     []int // Per-column width overrides (0 means auto-calculate)
}

// NewMarkdown creates a new Markdown formatter with the given configuration
func NewMarkdown(config ...MarkdownConfig) *Markdown {
	// Default configuration
	cfg := MarkdownConfig{
		HeaderAlignment: ALIGN_DEFAULT,
		FooterAlignment: ALIGN_DEFAULT,
		Alignment:       ALIGN_DEFAULT,
	}
	if len(config) > 0 {
		cfg = config[0] // Use provided config if present
	}

	f := &Markdown{
		headerAlignment:  cfg.HeaderAlignment,
		footerAlignment:  cfg.FooterAlignment,
		alignment:        cfg.Alignment,
		columnAlignments: cfg.ColumnAlignments,
		columnWidths:     cfg.ColumnWidths,
		newLine:          symbols.NEWLINE,
	}
	return f
}

func (f *Markdown) FormatHeader(w io.Writer, headers []string, colWidths map[int]int) {
	padFunc := f.pad(f.headerAlignment)
	cells := make([]string, len(headers))
	for i, h := range headers {
		width := colWidths[i]
		if i < len(f.columnWidths) && f.columnWidths[i] > 0 {
			width = f.columnWidths[i] // Override with specified width
		}
		cells[i] = padFunc(h, symbols.SPACE, width)
	}
	fmt.Fprintf(w, "| %s |%s", strings.Join(cells, " | "), f.newLine)
	f.FormatLine(w, colWidths, false)
}

func (f *Markdown) FormatRow(w io.Writer, row []string, colWidths map[int]int, isFirstRow bool) {
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
	}
	fmt.Fprintf(w, "| %s |%s", strings.Join(cells, " | "), f.newLine)
}

func (f *Markdown) FormatFooter(w io.Writer, footers []string, colWidths map[int]int) {
	padFunc := f.pad(f.footerAlignment)
	cells := make([]string, len(footers))
	for i, cell := range footers {
		width := colWidths[i]
		if i < len(f.columnWidths) && f.columnWidths[i] > 0 {
			width = f.columnWidths[i] // Override with specified width
		}
		cells[i] = padFunc(cell, symbols.SPACE, width)
	}
	fmt.Fprintf(w, "| %s |%s", strings.Join(cells, " | "), f.newLine)
}

func (f *Markdown) FormatLine(w io.Writer, colWidths map[int]int, isTop bool) {
	var separators []string
	for i := 0; i < len(colWidths); i++ {
		width := colWidths[i]
		if i < len(f.columnWidths) && f.columnWidths[i] > 0 {
			width = f.columnWidths[i] // Override with specified width
		}
		// Use body alignment for the separator line (under headers)
		align := f.alignment
		if i < len(f.columnAlignments) && f.columnAlignments[i] != ALIGN_DEFAULT {
			align = f.columnAlignments[i]
		}
		switch align {
		case ALIGN_CENTER:
			separators = append(separators, ":"+strings.Repeat("-", width-2)+":")
		case ALIGN_RIGHT:
			separators = append(separators, strings.Repeat("-", width-1)+":")
		case ALIGN_LEFT:
			separators = append(separators, ":"+strings.Repeat("-", width-1))
		default: // ALIGN_DEFAULT
			separators = append(separators, strings.Repeat("-", width))
		}
	}
	fmt.Fprintf(w, "| %s |%s", strings.Join(separators, " | "), f.newLine)
}

func (f *Markdown) GetColumnWidths() []int {
	return f.columnWidths
}

func (f *Markdown) Reset() {
	// No internal state to reset
}

func (f *Markdown) pad(align int) func(string, string, int) string {
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
