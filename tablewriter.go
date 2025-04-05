package tablewriter

import (
	"fmt"
	"github.com/olekukonko/tablewriter/symbols"
	"github.com/olekukonko/tablewriter/theme"
	"github.com/olekukonko/tablewriter/utils"
	"io"
	"os"
	"reflect"
	"strings"
)

// Iterator defines a generic iterator interface
type Iterator[T any] interface {
	Next() (T, bool)
}

// Table represents a text-based table writer
type Table struct {
	writer     io.Writer
	rows       [][][]string
	headers    []string
	footers    []string
	colWidths  map[int]int
	rowHeights map[int]int
	theme      theme.Structure
	config     theme.Config
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

// WithFormatter sets a custom theme
func WithFormatter(f theme.Structure) Option {
	return func(t *Table) { t.theme = f }
}

// WithConfig sets the table configuration
func WithConfig(cfg theme.Config) Option {
	return func(t *Table) {
		t.config = cfg
		t.theme = theme.NewDefault(cfg)
	}
}

// WithStringer sets a custom stringer callback for type T
func WithStringer[T any](s func(T) []string) Option {
	return func(t *Table) { t.stringer = s }
}

// NewTable creates a new table writer with a default theme
func NewTable(w io.Writer, opts ...Option) *Table {
	cfg := theme.DefaultConfig()
	t := &Table{
		writer:     w,
		colWidths:  make(map[int]int),
		rowHeights: make(map[int]int),
		theme:      theme.NewDefault(cfg),
		config:     cfg,
		newLine:    symbols.NEWLINE,
	}
	for _, opt := range opts {
		opt(t)
	}
	return t
}

// NewWriter creates a table with default configurations (legacy compatibility)
func NewWriter(w io.Writer) *Table {
	return NewTable(w)
}

// ensureInitialized ensures the table is in a valid state
func (t *Table) ensureInitialized() {
	if t.colWidths == nil {
		t.colWidths = make(map[int]int)
	}
	if t.rowHeights == nil {
		t.rowHeights = make(map[int]int)
	}
	if t.theme == nil {
		t.theme = theme.NewDefault()
	}
	if t.config.Symbols == nil {
		t.config.Symbols = symbols.NewSymbols(symbols.StyleASCII)
	}
}

// SetHeader sets the table header
func (t *Table) SetHeader(headers []string) {
	t.ensureInitialized()
	formatted := make([]string, len(headers))
	for i, h := range headers {
		if t.config.Header.AutoFormat {
			h = utils.Title(strings.Join(utils.SplitCamelCase(h), " "))
			fmt.Fprintf(os.Stderr, "Formatted header %d: %q\n", i, h)
		}
		if t.config.Header.AutoWrap && t.config.Header.MaxWidth > 0 {
			lines, _ := utils.WrapString(h, t.config.Header.MaxWidth)
			formatted[i] = strings.Join(lines, "\n")
		} else {
			formatted[i] = h
		}
	}
	t.headers = formatted
	t.parseDimension(t.headers, -1)
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
	t.footers = padded
	t.parseDimension(t.footers, len(t.rows))
}

// Append adds a row to the table
func (t *Table) Append(row interface{}) error {
	t.ensureInitialized()
	lines, err := t.toStringLines(row)
	if err != nil {
		return err
	}
	t.rows = append(t.rows, lines)
	t.parseDimension(lines[0], len(t.rows)-1)
	return nil
}

// AppendStructs adds rows from a slice of structs
func (t *Table) AppendStructs(v interface{}) error {
	t.ensureInitialized()
	if v == nil {
		return fmt.Errorf("nil value")
	}
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Slice && rv.Kind() != reflect.Array {
		return fmt.Errorf("input must be a slice or array, got %T", v)
	}
	if rv.Len() == 0 {
		return fmt.Errorf("empty slice")
	}

	first := rv.Index(0)
	if first.Kind() == reflect.Ptr && first.IsNil() {
		return fmt.Errorf("first element is nil")
	}
	typ := first.Type()
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}
	if typ.Kind() != reflect.Struct {
		return fmt.Errorf("elements must be structs or pointers to structs, got %s", typ.Kind())
	}

	numFields := typ.NumField()
	if len(t.headers) == 0 {
		headers := make([]string, numFields)
		for i := 0; i < numFields; i++ {
			field := typ.Field(i)
			if tag := field.Tag.Get("tablewriter"); tag != "" {
				headers[i] = tag
			} else {
				headers[i] = field.Name
			}
		}
		t.headers = headers
	}

	for i := 0; i < rv.Len(); i++ {
		item := rv.Index(i)
		if item.Kind() == reflect.Ptr && item.IsNil() {
			continue
		}
		if item.Kind() == reflect.Ptr {
			item = item.Elem()
		}
		if !item.IsValid() || item.Kind() != reflect.Struct {
			return fmt.Errorf("invalid item at index %d: %v", i, item.Kind())
		}

		cells := make([]string, numFields)
		for j := 0; j < numFields; j++ {
			field := item.Field(j)
			if field.Kind() == reflect.Ptr && field.IsNil() {
				cells[j] = "nil"
			} else {
				if field.Kind() == reflect.Ptr {
					field = field.Elem()
				}
				if s, ok := field.Interface().(fmt.Stringer); ok && field.IsValid() {
					cells[j] = s.String()
				} else if field.IsValid() {
					cells[j] = fmt.Sprint(field.Interface())
				} else {
					cells[j] = "nil"
				}
			}
		}
		lines := t.splitLines(cells)
		t.rows = append(t.rows, lines)
		t.parseDimension(cells, len(t.rows)-1)
	}
	return nil
}

// StreamRow renders a row immediately
func (t *Table) StreamRow(row interface{}) error {
	t.ensureInitialized()
	f := t.theme
	if !t.hasPrinted {
		f.FormatLine(t.writer, t.colWidths, symbols.Top)
		if len(t.headers) > 0 {
			f.FormatHeader(t.writer, t.headers, t.colWidths)
		}
		t.hasPrinted = true
	}
	lines, err := t.toStringLines(row)
	if err != nil {
		return err
	}
	for i, line := range lines {
		t.parseDimension(line, len(t.rows))
		f.FormatRow(t.writer, line, t.colWidths, i == 0)
	}
	t.rows = append(t.rows, lines)
	return nil
}

// RenderFromIterator renders the table from an iterator
func (t *Table) RenderFromIterator(iter interface{}) error {
	t.ensureInitialized()
	f := t.theme
	f.FormatLine(t.writer, t.colWidths, symbols.Top)
	if len(t.headers) > 0 {
		f.FormatHeader(t.writer, t.headers, t.colWidths)
	}
	t.hasPrinted = true

	switch v := iter.(type) {
	case Iterator[[]string]:
		for row, ok := v.Next(); ok; row, ok = v.Next() {
			lines := t.splitLines(row)
			for i, line := range lines {
				t.parseDimension(line, len(t.rows))
				f.FormatRow(t.writer, line, t.colWidths, i == 0)
			}
			t.rows = append(t.rows, lines)
		}
	default:
		rv := reflect.ValueOf(iter)
		if rv.Kind() != reflect.Interface || rv.IsNil() {
			return fmt.Errorf("invalid iterator type: %T", iter)
		}
		iterType := rv.Elem().Type()
		if iterType.Kind() != reflect.Struct {
			return fmt.Errorf("iterator must be a struct implementing Iterator[T]")
		}
		nextMethod := rv.MethodByName("Next")
		if !nextMethod.IsValid() {
			return fmt.Errorf("iterator must implement Next() (T, bool)")
		}
		for {
			results := nextMethod.Call(nil)
			if len(results) != 2 {
				return fmt.Errorf("next() must return (T, bool)")
			}
			ok := results[1].Bool()
			if !ok {
				break
			}
			lines, err := t.toStringLines(results[0].Interface())
			if err != nil {
				return err
			}
			for i, line := range lines {
				t.parseDimension(line, len(t.rows))
				f.FormatRow(t.writer, line, t.colWidths, i == 0)
			}
			t.rows = append(t.rows, lines)
		}
	}

	if len(t.footers) > 0 {
		f.FormatLine(t.writer, t.colWidths, symbols.Middle)
		f.FormatFooter(t.writer, t.footers, t.colWidths)
	}
	f.FormatLine(t.writer, t.colWidths, symbols.Bottom)
	return nil
}

// Render outputs the table to the writer
func (t *Table) Render() error {
	t.ensureInitialized()
	t.colWidths = make(map[int]int)
	for i, w := range t.theme.GetColumnWidths() {
		if w > 0 {
			t.colWidths[i] = w
		}
	}
	if len(t.headers) > 0 {
		t.parseDimension(t.headers, -1)
	}
	for i, lines := range t.rows {
		for _, line := range lines {
			t.parseDimension(line, i)
		}
	}
	if len(t.footers) > 0 {
		t.parseDimension(t.footers, len(t.rows))
	}

	f := t.theme
	f.FormatLine(t.writer, t.colWidths, symbols.Top)
	if len(t.headers) > 0 {
		f.FormatHeader(t.writer, t.headers, t.colWidths)
	}
	for i, lines := range t.rows {
		for j, line := range lines {
			f.FormatRow(t.writer, line, t.colWidths, j == 0 && i == 0)
		}
	}
	if len(t.footers) > 0 {
		f.FormatLine(t.writer, t.colWidths, symbols.Middle)
		f.FormatFooter(t.writer, t.footers, t.colWidths)
		f.FormatLine(t.writer, t.colWidths, symbols.Bottom)
	} else {
		f.FormatLine(t.writer, t.colWidths, symbols.Bottom)
	}
	return nil
}

// Metrics returns statistics about the table
func (t *Table) Metrics() struct {
	HeaderCount int
	RowCount    int
	FooterCount int
} {
	t.ensureInitialized()
	return struct {
		HeaderCount int
		RowCount    int
		FooterCount int
	}{
		HeaderCount: len(t.headers),
		RowCount:    len(t.rows),
		FooterCount: len(t.footers),
	}
}

// toStringLines converts a row to a slice of string slices (multi-line support)
func (t *Table) toStringLines(row interface{}) ([][]string, error) {
	switch v := row.(type) {
	case []string:
		if t.config.Row.AutoWrap && t.config.Row.MaxWidth > 0 {
			wrapped := make([]string, len(v))
			for i, cell := range v {
				lines, _ := utils.WrapString(cell, t.config.Row.MaxWidth)
				wrapped[i] = strings.Join(lines, "\n")
			}
			return t.splitLines(wrapped), nil
		}
		return t.splitLines(v), nil
	default:
		if t.stringer == nil {
			return nil, fmt.Errorf("no stringer provided for type %T", row)
		}
		rv := reflect.ValueOf(t.stringer)
		if rv.Kind() != reflect.Func || rv.Type().NumIn() != 1 || rv.Type().NumOut() != 1 {
			return nil, fmt.Errorf("stringer must be a func(T) []string")
		}
		in := []reflect.Value{reflect.ValueOf(row)}
		out := rv.Call(in)
		if len(out) != 1 || out[0].Kind() != reflect.Slice || out[0].Type().Elem().Kind() != reflect.String {
			return nil, fmt.Errorf("stringer must return []string")
		}
		return t.splitLines(out[0].Interface().([]string)), nil
	}
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

// parseDimension parses table dimensions for a row
func (t *Table) parseDimension(row []string, rowKey int) {
	var cfg *theme.CellConfig
	switch rowKey {
	case -1: // Headers
		cfg = &t.config.Header
	case len(t.rows): // Footers
		cfg = &t.config.Footer
	default: // Rows
		cfg = &t.config.Row
	}
	for colKey, cell := range row {
		lines := strings.Split(cell, "\n")
		maxWidth := 0
		for _, line := range lines {
			w := utils.DisplayWidth(line)
			if cfg.AutoWrap && cfg.MaxWidth > 0 {
				wrapped, _ := utils.WrapString(line, cfg.MaxWidth)
				for _, wrappedLine := range wrapped {
					if w := utils.DisplayWidth(wrappedLine); w > maxWidth {
						maxWidth = w
					}
				}
			} else if w > maxWidth {
				maxWidth = w
			}
		}
		// Calculate rendered width with padding
		align := cfg.Alignment
		if colKey < len(cfg.ColumnAligns) && cfg.ColumnAligns[colKey] != theme.ALIGN_DEFAULT {
			align = cfg.ColumnAligns[colKey]
		}
		var padded string
		switch align {
		case theme.ALIGN_CENTER:
			padded = utils.Pad(strings.Join(lines, "\n"), symbols.SPACE, maxWidth)
		case theme.ALIGN_RIGHT:
			padded = utils.PadLeft(strings.Join(lines, "\n"), symbols.SPACE, maxWidth)
		case theme.ALIGN_LEFT, theme.ALIGN_DEFAULT:
			padded = utils.PadRight(strings.Join(lines, "\n"), symbols.SPACE, maxWidth)
		}
		if w := utils.DisplayWidth(padded); w > maxWidth {
			maxWidth = w
		}
		// Apply theme-based width overrides
		if widths := t.theme.GetColumnWidths(); len(widths) > colKey && widths[colKey] > 0 {
			if maxWidth < widths[colKey] {
				maxWidth = widths[colKey]
			}
		}
		// Enforce maxWidth cap from config
		if cfg.MaxWidth > 0 && maxWidth > cfg.MaxWidth {
			maxWidth = cfg.MaxWidth
		}
		// Update column width
		if current, exists := t.colWidths[colKey]; !exists || maxWidth > current {
			t.colWidths[colKey] = maxWidth
		}
		// Update row height
		if h := len(lines); h > 0 {
			if curr, ok := t.rowHeights[rowKey]; !ok || h > curr {
				t.rowHeights[rowKey] = h
			}
		}
	}
}

// Theme returns the current theme for configuration
func (t *Table) Theme() theme.Structure {
	return t.theme
}
