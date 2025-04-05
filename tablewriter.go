package tablewriter

import (
	"github.com/olekukonko/errors"
	"io"
	"reflect"
	"strings"

	"github.com/olekukonko/tablewriter/renderer"
	"github.com/olekukonko/tablewriter/symbols"
	"github.com/olekukonko/tablewriter/utils"
)

// Iterator defines a generic iterator interface
type Iterator[T any] interface {
	Next() (T, bool)
}

// CellConfig defines pre-processing settings for a table section
type CellConfig struct {
	MaxWidth     int
	AutoWrap     bool
	Alignment    string
	ColumnAligns []string
	AutoFormat   bool
	Filter       func([]string) []string // Updated to return []string
	Callback     func()
	Padding      symbols.Padding
	AutoMerge    bool // Added to support cell merging
}

// Config holds the shared configuration for table content preparation
type Config struct {
	Header  CellConfig
	Row     CellConfig
	Footer  CellConfig
	Symbols symbols.Symbols
}

// DefaultConfig returns a default configuration
func DefaultConfig() Config {
	s := symbols.NewSymbols(symbols.StyleASCII)
	defaultPadding := symbols.Padding{Left: " ", Right: " ", Top: "", Bottom: ""}
	return Config{
		Header: CellConfig{
			MaxWidth:   0,
			AutoWrap:   true,
			Alignment:  renderer.AlignCenter,
			AutoFormat: true,
			Padding:    defaultPadding,
		},
		Row: CellConfig{
			MaxWidth:  0,
			AutoWrap:  true,
			Alignment: renderer.AlignLeft,
			Padding:   defaultPadding,
		},
		Footer: CellConfig{
			MaxWidth:  0,
			AutoWrap:  true,
			Alignment: renderer.AlignRight,
			Padding:   defaultPadding,
		},
		Symbols: s,
	}
}

// Table represents a text-based table writer
type Table struct {
	writer     io.Writer
	rows       [][][]string
	headers    []string
	footers    []string
	colWidths  map[int]int
	rowHeights map[int]int
	renderer   renderer.Structure
	config     Config
	stringer   any
	newLine    string
	hasPrinted bool
}

// Option defines a configuration function for Table
type Option func(*Table)

// WithHeader sets the table header
func WithHeader(headers []string) Option {
	return func(t *Table) { t.headers = headers }
}

// WithFooter sets the table footer
func WithFooter(footers []string) Option {
	return func(t *Table) { t.footers = footers }
}

// WithFormatter sets a custom renderer
func WithFormatter(f renderer.Structure) Option {
	return func(t *Table) { t.renderer = f }
}

// WithConfig sets the table configuration
func WithConfig(cfg Config) Option {
	return func(t *Table) { t.config = cfg }
}

// WithStringer sets a custom stringer callback for type T
func WithStringer[T any](s func(T) []string) Option {
	return func(t *Table) { t.stringer = s }
}

// NewTable creates a new table writer with a default renderer
func NewTable(w io.Writer, opts ...Option) *Table {
	t := &Table{
		writer:     w,
		colWidths:  make(map[int]int),
		rowHeights: make(map[int]int),
		renderer:   renderer.NewDefault(),
		config:     DefaultConfig(),
		newLine:    symbols.NewLine,
	}
	for _, opt := range opts {
		opt(t)
	}
	return t
}

// Append adds a single row to the table
func (t *Table) Append(row interface{}) error {
	t.ensureInitialized()
	lines, err := t.toStringLines(row)
	if err != nil {
		return err
	}
	if t.config.Row.Filter != nil {
		for i := range lines {
			lines[i] = t.config.Row.Filter(lines[i])
		}
	}
	t.rows = append(t.rows, lines)
	for _, line := range lines {
		t.parseDimension(line, renderer.Row)
	}
	return nil
}

// AppendBulk adds multiple rows to the table at once
func (t *Table) AppendBulk(rows []interface{}) error {
	t.ensureInitialized()
	for _, row := range rows {
		if err := t.Append(row); err != nil {
			return err
		}
	}
	return nil
}

// SetHeader sets the table header
func (t *Table) SetHeader(headers []string) {
	t.ensureInitialized()
	formatted := make([]string, len(headers))
	for i, h := range headers {
		if t.config.Header.AutoFormat {
			h = utils.Title(strings.Join(utils.SplitCamelCase(h), " "))
		}
		if t.config.Header.AutoWrap && t.config.Header.MaxWidth > 0 {
			lines, _ := utils.WrapString(h, t.config.Header.MaxWidth)
			h = strings.Join(lines, t.newLine)
		}
		if t.config.Header.Filter != nil {
			h = t.config.Header.Filter([]string{h})[0]
		}
		formatted[i] = h
	}
	t.headers = formatted
	t.parseDimension(t.headers, renderer.Header)
}

// SetFooter sets the table footer
func (t *Table) SetFooter(footers []string) {
	t.ensureInitialized()
	numCols := len(t.colWidths)
	padded := make([]string, numCols)
	if t.config.Footer.AutoWrap && t.config.Footer.MaxWidth > 0 {
		for i, f := range footers {
			if i < numCols {
				lines, _ := utils.WrapString(f, t.config.Footer.MaxWidth)
				padded[i] = strings.Join(lines, t.newLine)
			}
		}
	} else {
		copy(padded, footers)
		for i := len(footers); i < numCols; i++ {
			padded[i] = ""
		}
	}
	if t.config.Footer.Filter != nil {
		padded = t.config.Footer.Filter(padded)
	}
	t.footers = padded
	t.parseDimension(t.footers, renderer.Footer)
}

// Render outputs the table to the writer
func (t *Table) Render() error {
	t.ensureInitialized()
	t.colWidths = make(map[int]int)

	// Calculate widths for all sections
	if len(t.headers) > 0 {
		t.parseDimension(t.headers, renderer.Header)
	}
	for _, lines := range t.rows {
		for _, line := range lines {
			t.parseDimension(line, renderer.Row)
		}
	}
	if len(t.footers) > 0 {
		t.parseDimension(t.footers, renderer.Footer)
	}

	f := t.renderer
	f.FormatLine(t.writer, renderer.Context{Widths: t.colWidths, Level: renderer.Top})

	// Header
	if len(t.headers) > 0 {
		colAligns := make(map[int]string)
		colPadding := make(map[int]symbols.Padding)
		for i := range t.headers {
			colAligns[i] = t.config.Header.Alignment
			colPadding[i] = t.config.Header.Padding
		}
		f.FormatHeader(t.writer, t.headers, renderer.Context{
			Widths:     t.colWidths,
			Align:      t.config.Header.Alignment,
			Padding:    t.config.Header.Padding,
			ColPadding: colPadding,
			ColAligns:  colAligns,
		})
	}

	// Rows with optional cell merging
	for i, lines := range t.rows {
		for j, line := range lines {
			colPadding := make(map[int]symbols.Padding)
			colAligns := make(map[int]string)
			for colKey := range line {
				colPadding[colKey] = t.config.Row.Padding
				if colKey < len(t.config.Row.ColumnAligns) && t.config.Row.ColumnAligns[colKey] != "" {
					colAligns[colKey] = t.config.Row.ColumnAligns[colKey]
				} else {
					colAligns[colKey] = t.config.Row.Alignment
				}
			}
			// Apply cell merging if enabled
			if t.config.Row.AutoMerge && i > 0 && j == 0 {
				prevLines := t.rows[i-1]
				if len(prevLines) > 0 && len(prevLines[0]) == len(line) {
					for colKey := range line {
						if colKey == 0 && line[colKey] == prevLines[0][colKey] {
							line[colKey] = "" // Merge by clearing duplicate value
						}
					}
				}
			}
			f.FormatRow(t.writer, line, renderer.Context{
				Widths:     t.colWidths,
				Level:      renderer.Middle,
				Align:      t.config.Row.Alignment,
				Padding:    t.config.Row.Padding,
				ColPadding: colPadding,
				ColAligns:  colAligns,
				First:      i == 0 && j == 0,
			})
		}
	}

	// Footer
	if len(t.footers) > 0 {
		f.FormatLine(t.writer, renderer.Context{Widths: t.colWidths, Level: renderer.Middle})
		colPadding := make(map[int]symbols.Padding)
		colAligns := make(map[int]string)
		for i := range t.footers {
			colPadding[i] = t.config.Footer.Padding
			if i < len(t.config.Footer.ColumnAligns) && t.config.Footer.ColumnAligns[i] != "" {
				colAligns[i] = t.config.Footer.ColumnAligns[i]
			} else {
				colAligns[i] = t.config.Footer.Alignment
			}
		}
		f.FormatFooter(t.writer, t.footers, renderer.Context{
			Widths:     t.colWidths,
			Align:      t.config.Footer.Alignment,
			Padding:    t.config.Footer.Padding,
			ColPadding: colPadding,
			ColAligns:  colAligns,
		})
		f.FormatLine(t.writer, renderer.Context{Widths: t.colWidths, Level: renderer.Bottom})
	} else {
		f.FormatLine(t.writer, renderer.Context{Widths: t.colWidths, Level: renderer.Bottom})
	}
	t.hasPrinted = true
	return nil
}

// ensureInitialized ensures the table is in a valid state
func (t *Table) ensureInitialized() {
	if t.colWidths == nil {
		t.colWidths = make(map[int]int)
	}
	if t.rowHeights == nil {
		t.rowHeights = make(map[int]int)
	}
	if t.renderer == nil {
		t.renderer = renderer.NewDefault()
	}
	if t.config.Symbols == nil {
		t.config.Symbols = symbols.NewSymbols(symbols.StyleASCII)
	}
}

// toStringLines converts a row to a slice of string slices (multi-line support)
func (t *Table) toStringLines(row interface{}) ([][]string, error) {
	switch v := row.(type) {
	case []string:
		result := make([][]string, len(v))
		for i, cell := range v {
			var lines []string
			if t.config.Row.AutoWrap && t.config.Row.MaxWidth > 0 {
				wrapped, _ := utils.WrapString(cell, t.config.Row.MaxWidth)
				lines = wrapped
			} else {
				lines = strings.Split(cell, "\n")
			}
			paddedLines := make([]string, len(lines))
			for j, line := range lines {
				paddedLines[j] = t.config.Row.Padding.Left + line + t.config.Row.Padding.Right
			}
			result[i] = paddedLines
		}
		return t.normalizeLines(result), nil
	default:
		if t.stringer == nil {
			return nil, errors.Newf("no stringer provided for type %T", row)
		}
		rv := reflect.ValueOf(t.stringer)
		if rv.Kind() != reflect.Func || rv.Type().NumIn() != 1 || rv.Type().NumOut() != 1 {
			return nil, errors.Newf("stringer must be a func(T) []string")
		}
		in := []reflect.Value{reflect.ValueOf(row)}
		out := rv.Call(in)
		if len(out) != 1 || out[0].Kind() != reflect.Slice || out[0].Type().Elem().Kind() != reflect.String {
			return nil, errors.Newf("stringer must return []string")
		}
		return t.splitLines(out[0].Interface().([]string)), nil
	}
}

// normalizeLines ensures all columns have the same number of lines
func (t *Table) normalizeLines(lines [][]string) [][]string {
	maxLines := 0
	for _, col := range lines {
		if len(col) > maxLines {
			maxLines = len(col)
		}
	}
	result := make([][]string, maxLines)
	for i := 0; i < maxLines; i++ {
		result[i] = make([]string, len(lines))
		for j, col := range lines {
			if i < len(col) {
				result[i][j] = col[i]
			} else {
				result[i][j] = t.config.Row.Padding.Left + t.config.Row.Padding.Right
			}
		}
	}
	return result
}

// splitLines splits multi-line strings into separate lines
func (t *Table) splitLines(row []string) [][]string {
	var maxLines int
	var splitRows [][]string
	for _, cell := range row {
		lines := strings.Split(cell, "\n")
		splitRows = append(splitRows, lines)
		if len(lines) > maxLines {
			maxLines = len(lines)
		}
	}
	result := make([][]string, maxLines)
	for i := 0; i < maxLines; i++ {
		result[i] = make([]string, len(row))
		for j, lines := range splitRows {
			if i < len(lines) {
				result[i][j] = lines[i]
			} else {
				result[i][j] = ""
			}
		}
	}
	return result
}

// parseDimension calculates column widths and prepares content without padding
func (t *Table) parseDimension(row []string, position renderer.Position) int {
	var cfg *CellConfig
	switch position {
	case renderer.Header:
		cfg = &t.config.Header
	case renderer.Footer:
		cfg = &t.config.Footer
	default:
		cfg = &t.config.Row
	}

	maxWidth := 0
	for i, cell := range row {
		lines := strings.Split(cell, "\n")
		for _, line := range lines {
			contentWidth := utils.RuneWidth(line)
			totalWidth := contentWidth +
				utils.RuneWidth(cfg.Padding.Left) +
				utils.RuneWidth(cfg.Padding.Right)

			if cfg.MaxWidth > 0 && totalWidth > cfg.MaxWidth {
				totalWidth = cfg.MaxWidth
			}

			if current, exists := t.colWidths[i]; !exists || totalWidth > current {
				t.colWidths[i] = totalWidth
			}
			if totalWidth > maxWidth {
				maxWidth = totalWidth
			}
		}
	}
	return maxWidth
}
