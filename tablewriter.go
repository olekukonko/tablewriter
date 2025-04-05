// tablewriter.go
package tablewriter

import (
	"fmt"
	"io"
	"reflect"
	"strings"

	"github.com/olekukonko/tablewriter/formatter"
	"github.com/olekukonko/tablewriter/symbols"
	"github.com/olekukonko/tablewriter/utils"
)

// Iterator defines a generic iterator interface
type Iterator[T any] interface {
	Next() (T, bool)
}

// Table represents a text-based table writer
type Table struct {
	writer     io.Writer
	rows       [][][]string // Multi-line rows
	headers    []string
	footers    []string
	colWidths  map[int]int
	formatter  formatter.Formatter
	stringer   any // func(T) []string for custom types
	maxWidth   int
	newLine    string
	hasPrinted bool // Tracks if any content has been rendered
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

// WithFormatter sets a custom formatter
func WithFormatter(f formatter.Formatter) Option {
	return func(t *Table) { t.formatter = f }
}

// WithStringer sets a custom stringer callback for type T
func WithStringer[T any](s func(T) []string) Option {
	return func(t *Table) { t.stringer = s }
}

// WithMaxWidth sets the maximum column width
func WithMaxWidth(width int) Option {
	return func(t *Table) { t.maxWidth = width }
}

// NewTable creates a new table writer with a default formatter
func NewTable(w io.Writer, opts ...Option) *Table {
	t := &Table{
		writer:    w,
		colWidths: make(map[int]int),
		formatter: formatter.NewDefaultFormatter(),
		maxWidth:  formatter.DefaultMaxWidth,
		newLine:   symbols.NEWLINE,
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
	if t.formatter == nil {
		t.formatter = formatter.NewDefaultFormatter()
	}
}

// Append adds a row to the table
func (t *Table) Append(row interface{}) error {
	t.ensureInitialized()
	lines, err := t.toStringLines(row)
	if err != nil {
		return err
	}
	t.rows = append(t.rows, lines)
	for _, line := range lines {
		t.updateWidths(line)
	}
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

	// Set headers from the first element
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

	// Append rows
	for i := 0; i < rv.Len(); i++ {
		item := rv.Index(i)
		if item.Kind() == reflect.Ptr && item.IsNil() {
			continue // Skip nil pointers
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
		for _, line := range lines {
			t.updateWidths(line)
		}
	}
	return nil
}

// StreamRow renders a row immediately
func (t *Table) StreamRow(row interface{}) error {
	t.ensureInitialized()
	f := t.formatter
	if !t.hasPrinted {
		f.FormatLine(t.writer, t.colWidths, true) // Top line for all formatters
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
		t.updateWidths(line)
		f.FormatRow(t.writer, line, t.colWidths, i == 0)
	}
	return nil
}

// RenderFromIterator renders the table from an iterator
func (t *Table) RenderFromIterator(iter interface{}) error {
	t.ensureInitialized()
	f := t.formatter
	f.FormatLine(t.writer, t.colWidths, true) // Top line for all formatters
	if len(t.headers) > 0 {
		f.FormatHeader(t.writer, t.headers, t.colWidths)
	}
	t.hasPrinted = true

	switch v := iter.(type) {
	case Iterator[[]string]:
		for row, ok := v.Next(); ok; row, ok = v.Next() {
			lines := t.splitLines(row)
			for i, line := range lines {
				t.updateWidths(line)
				f.FormatRow(t.writer, line, t.colWidths, i == 0)
			}
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
				return fmt.Errorf("Next() must return (T, bool)")
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
				t.updateWidths(line)
				f.FormatRow(t.writer, line, t.colWidths, i == 0)
			}
		}
	}

	if len(t.footers) > 0 {
		f.FormatFooter(t.writer, t.footers, t.colWidths)
	}
	f.FormatLine(t.writer, t.colWidths, false) // Bottom line for all formatters
	return nil
}

// Render outputs the table to the writer
func (t *Table) Render() error {
	t.ensureInitialized()
	f := t.formatter
	f.FormatLine(t.writer, t.colWidths, true) // Top line for all formatters
	if len(t.headers) > 0 {
		f.FormatHeader(t.writer, t.headers, t.colWidths)
	}
	for i, lines := range t.rows {
		for j, line := range lines {
			f.FormatRow(t.writer, line, t.colWidths, j == 0 && i == 0)
		}
	}
	if len(t.footers) > 0 {
		f.FormatFooter(t.writer, t.footers, t.colWidths)
	}
	f.FormatLine(t.writer, t.colWidths, false) // Bottom line for all formatters
	return nil
}

// toStringLines converts a row to a slice of string slices (multi-line support)
func (t *Table) toStringLines(row interface{}) ([][]string, error) {
	switch v := row.(type) {
	case []string:
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

// updateWidths updates column widths based on a row
func (t *Table) updateWidths(row []string) {
	for i, cell := range row {
		w := utils.DisplayWidth(cell)
		if curr, ok := t.colWidths[i]; !ok || w > curr {
			t.colWidths[i] = w
		}
	}
}

// Formatter returns the current formatter for configuration
func (t *Table) Formatter() formatter.Formatter {
	return t.formatter
}
