package theme

import (
	"fmt"
	"github.com/olekukonko/tablewriter/symbols"
	"github.com/olekukonko/tablewriter/utils"
	"io"
	"os"
	"strings"
)

// CellConfig defines configuration for a specific part of the table
type CellConfig struct {
	Alignment    int   // ALIGN_LEFT, ALIGN_RIGHT, ALIGN_CENTER, ALIGN_DEFAULT
	ColumnAligns []int // Per-column alignment overrides
	MaxWidth     int   // Maximum width for wrapping (0 = no limit)
	AutoWrap     bool  // Enable text wrapping
	AutoFormat   bool  // Auto-format text (e.g., capitalize headers)
	ColumnWidths []int // Predefined column widths (0 = auto-calculate)
}

// Config holds the full table configuration
type Config struct {
	Borders         Border
	Header          CellConfig
	Row             CellConfig
	Footer          CellConfig
	HeaderLine      bool            // Draw a line under the header
	CenterSeparator string          // Custom center separator (e.g., "+")
	RowSeparator    string          // Custom row separator (e.g., "-")
	ColumnSeparator string          // Custom column separator (e.g., "|")
	Symbols         symbols.Symbols // Symbol set (e.g., ASCII, Unicode)
}

// DefaultConfig returns a default configuration
func DefaultConfig() Config {
	s := symbols.NewSymbols(symbols.StyleASCII)
	return Config{
		Borders: Border{Left: true, Right: true, Top: true, Bottom: true},
		Header: CellConfig{
			Alignment:    ALIGN_CENTER,
			AutoFormat:   true,
			AutoWrap:     true,
			MaxWidth:     0, // No limit by default
			ColumnWidths: []int{},
		},
		Row: CellConfig{
			Alignment:    ALIGN_LEFT,
			AutoWrap:     true,
			MaxWidth:     0,
			ColumnWidths: []int{},
		},
		Footer: CellConfig{
			Alignment:    ALIGN_RIGHT,
			AutoWrap:     true,
			MaxWidth:     0,
			ColumnWidths: []int{},
		},
		HeaderLine:      true,
		CenterSeparator: "",
		RowSeparator:    "",
		ColumnSeparator: "",
		Symbols:         s,
	}
}

// Default implements classic table formatting
type Default struct {
	config Config
}

// NewDefault creates a new Default theme with the given configuration
func NewDefault(config ...Config) *Default {
	cfg := DefaultConfig()
	if len(config) > 0 {
		cfg = config[0]
	}
	if cfg.Symbols == nil {
		cfg.Symbols = symbols.NewSymbols(symbols.StyleASCII)
	}
	if cfg.CenterSeparator == "" {
		cfg.CenterSeparator = cfg.Symbols.Center()
	}
	if cfg.RowSeparator == "" {
		cfg.RowSeparator = cfg.Symbols.Row()
	}
	if cfg.ColumnSeparator == "" {
		cfg.ColumnSeparator = cfg.Symbols.Column()
	}
	fmt.Fprintf(os.Stderr, "Default Symbols: Center=%q, Row=%q, Column=%q\n", cfg.CenterSeparator, cfg.RowSeparator, cfg.ColumnSeparator)
	return &Default{config: cfg}
}

// FormatHeader formats the header row
func (f *Default) FormatHeader(w io.Writer, headers []string, colWidths map[int]int) {
	maxLines := 1
	splitHeaders := make([][]string, len(headers))
	for i, h := range headers {
		// Headers are already wrapped by SetHeader, so just split them
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
			var content string
			if lineIdx < len(splitHeaders[i]) {
				content = splitHeaders[i][lineIdx]
			}
			// Apply padding without extra spaces
			switch f.config.Header.Alignment {
			case ALIGN_CENTER:
				cells[i] = utils.Pad(content, " ", width)
			case ALIGN_RIGHT:
				cells[i] = utils.PadLeft(content, " ", width)
			case ALIGN_LEFT, ALIGN_DEFAULT:
				cells[i] = utils.PadRight(content, " ", width)
			}
		}
		prefix := utils.ConditionString(f.config.Borders.Left, f.config.ColumnSeparator, "")
		suffix := utils.ConditionString(f.config.Borders.Right, f.config.ColumnSeparator, "")
		fmt.Fprintf(w, "%s%s%s%s", prefix, strings.Join(cells, f.config.ColumnSeparator), suffix, symbols.NEWLINE)
	}
	if f.config.HeaderLine {
		f.FormatLine(w, colWidths, symbols.Middle)
	}
}

// FormatRow formats a data row
func (f *Default) FormatRow(w io.Writer, row []string, colWidths map[int]int, isFirstRow bool) {
	cells := make([]string, len(row))
	for i, cell := range row {
		align := f.config.Row.Alignment
		if i < len(f.config.Row.ColumnAligns) && f.config.Row.ColumnAligns[i] != ALIGN_DEFAULT {
			align = f.config.Row.ColumnAligns[i]
		}
		width := colWidths[i]
		switch align {
		case ALIGN_CENTER:
			cells[i] = utils.Pad(cell, symbols.SPACE, width)
		case ALIGN_RIGHT:
			cells[i] = utils.PadLeft(cell, symbols.SPACE, width)
		case ALIGN_LEFT, ALIGN_DEFAULT:
			cells[i] = utils.PadRight(cell, symbols.SPACE, width)
		}
	}
	prefix := utils.ConditionString(f.config.Borders.Left, f.config.ColumnSeparator, "")
	suffix := utils.ConditionString(f.config.Borders.Right, f.config.ColumnSeparator, "")
	fmt.Fprintf(w, "%s%s%s%s", prefix, strings.Join(cells, f.config.ColumnSeparator), suffix, symbols.NEWLINE)
}

// FormatFooter formats the footer row
func (f *Default) FormatFooter(w io.Writer, footers []string, colWidths map[int]int) {
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
			align := f.config.Footer.Alignment
			if i < len(f.config.Footer.ColumnAligns) && f.config.Footer.ColumnAligns[i] != ALIGN_DEFAULT {
				align = f.config.Footer.ColumnAligns[i]
			}
			if utils.IsNumeric(strings.TrimSpace(footers[i])) {
				align = ALIGN_RIGHT
			}
			width := colWidths[i]
			var content string
			if lineIdx < len(splitFooters[i]) {
				content = splitFooters[i][lineIdx]
			}
			switch align {
			case ALIGN_CENTER:
				cells[i] = utils.Pad(content, symbols.SPACE, width)
			case ALIGN_RIGHT:
				cells[i] = utils.PadLeft(content, symbols.SPACE, width)
			case ALIGN_LEFT, ALIGN_DEFAULT:
				cells[i] = utils.PadRight(content, symbols.SPACE, width)
			}
		}
		prefix := utils.ConditionString(f.config.Borders.Left, f.config.ColumnSeparator, "")
		suffix := utils.ConditionString(f.config.Borders.Right, f.config.ColumnSeparator, "")
		fmt.Fprintf(w, "%s%s%s%s", prefix, strings.Join(cells, f.config.ColumnSeparator), suffix, symbols.NEWLINE)
	}
}

// FormatLine formats a separator line
func (f *Default) FormatLine(w io.Writer, colWidths map[int]int, lineType string) {
	var line strings.Builder
	isMarkdown := f.config.Symbols.Center() == "|" && f.config.Symbols.Row() == "-" && f.config.Symbols.TopMid() == ""
	if isMarkdown {
		for i := 0; i < len(colWidths); i++ {
			if i > 0 {
				line.WriteString(f.config.Symbols.Column())
			}
			width := colWidths[i]
			line.WriteString(strings.Repeat(f.config.Symbols.Row(), width))
		}
	} else {
		if f.config.Borders.Left {
			switch lineType {
			case symbols.Top:
				line.WriteString(f.config.Symbols.TopLeft())
			case symbols.Middle:
				line.WriteString(f.config.Symbols.MidLeft())
			case symbols.Bottom:
				line.WriteString(f.config.Symbols.BottomLeft())
			}
		}
		numCols := len(colWidths)
		for i := 0; i < numCols; i++ {
			if i > 0 {
				switch lineType {
				case symbols.Top:
					line.WriteString(f.config.Symbols.TopMid())
				case symbols.Middle:
					line.WriteString(f.config.Symbols.Center())
				case symbols.Bottom:
					line.WriteString(f.config.Symbols.BottomMid())
				}
			}
			width := colWidths[i]
			line.WriteString(strings.Repeat(f.config.Symbols.Row(), width))
		}
		if f.config.Borders.Right {
			switch lineType {
			case symbols.Top:
				line.WriteString(f.config.Symbols.TopRight())
			case symbols.Middle:
				line.WriteString(f.config.Symbols.MidRight())
			case symbols.Bottom:
				line.WriteString(f.config.Symbols.BottomRight())
			}
		}
	}
	fmt.Fprintln(w, line.String())
}

// GetColumnWidths returns predefined column widths
func (f *Default) GetColumnWidths() []int {
	return f.config.Row.ColumnWidths
}

// Reset resets the theme (no-op for Default)
func (f *Default) Reset() {}
