package tablewriter

import (
	"fmt"
	"github.com/olekukonko/errors"
	"io"
	"reflect"
	"strings"

	"github.com/olekukonko/tablewriter/renderer"
	"github.com/olekukonko/tablewriter/symbols"
	"github.com/olekukonko/tablewriter/utils"
)

type Iterator[T any] interface {
	Next() (T, bool)
}

type Filter func([]string) []string

type CellFormatting struct {
	Alignment  string
	AutoWrap   bool
	AutoFormat bool
	AutoMerge  bool
	Truncate   bool
	MaxWidth   int
}

type CellPadding struct {
	Global    symbols.Padding
	PerColumn []symbols.Padding
}

type CellCallbacks struct {
	Global    func()
	PerColumn []func()
}

type CellConfig struct {
	Formatting   CellFormatting
	Padding      CellPadding
	Callbacks    CellCallbacks
	Filter       Filter
	ColumnAligns []string
	ColMaxWidths map[int]int
}

type Config struct {
	MaxWidth int // Global max width default
	Header   CellConfig
	Row      CellConfig
	Footer   CellConfig
}

func defaultConfig() Config {
	defaultPadding := symbols.Padding{Left: " ", Right: " ", Top: "", Bottom: ""}
	return Config{
		MaxWidth: 0,
		Header: CellConfig{
			Formatting: CellFormatting{
				MaxWidth:   0,
				AutoWrap:   true,
				Alignment:  renderer.AlignCenter,
				AutoFormat: true,
				Truncate:   true,
			},
			Padding: CellPadding{
				Global: defaultPadding,
			},
		},
		Row: CellConfig{
			Formatting: CellFormatting{
				MaxWidth:  0,
				AutoWrap:  true,
				Alignment: renderer.AlignLeft,
				Truncate:  false,
			},
			Padding: CellPadding{
				Global: defaultPadding,
			},
		},
		Footer: CellConfig{
			Formatting: CellFormatting{
				MaxWidth:  0,
				AutoWrap:  true,
				Alignment: renderer.AlignRight,
				Truncate:  false,
			},
			Padding: CellPadding{
				Global: defaultPadding,
			},
		},
	}
}

type Table struct {
	writer       io.Writer
	rows         [][][]string
	headers      []string
	footers      []string
	headerWidths map[int]int
	rowWidths    map[int]int
	footerWidths map[int]int
	rowHeights   map[int]int
	renderer     renderer.Renderer
	config       Config
	stringer     any
	newLine      string
	hasPrinted   bool
}

type Option func(*Table)

func WithHeader(headers []string) Option {
	return func(t *Table) { t.headers = headers }
}

func WithFooter(footers []string) Option {
	return func(t *Table) { t.footers = footers }
}

func WithRenderer(f renderer.Renderer) Option {
	return func(t *Table) { t.renderer = f }
}

func WithConfig(cfg Config) Option {
	return func(t *Table) { t.config = cfg }
}

func WithStringer[T any](s func(T) []string) Option {
	return func(t *Table) { t.stringer = s }
}

func NewTable(w io.Writer, opts ...Option) *Table {
	t := &Table{
		writer:       w,
		headerWidths: make(map[int]int),
		rowWidths:    make(map[int]int),
		footerWidths: make(map[int]int),
		rowHeights:   make(map[int]int),
		renderer:     renderer.NewDefault(),
		config:       defaultConfig(),
		newLine:      symbols.NewLine,
	}
	for _, opt := range opts {
		opt(t)
	}
	return t
}

func NewWriter(w io.Writer) *Table {
	return NewTable(w)
}

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

func (t *Table) Bulk(rows interface{}) error {
	rv := reflect.ValueOf(rows)
	if rv.Kind() != reflect.Slice {
		return errors.Newf("Bulk expects a slice, got %T", rows)
	}
	for i := 0; i < rv.Len(); i++ {
		row := rv.Index(i).Interface()
		if err := t.Append(row); err != nil {
			return err
		}
	}
	return nil
}

func (t *Table) SetHeader(headers []string) {
	t.ensureInitialized()
	formatted := make([]string, len(headers))
	for i, h := range headers {
		if t.config.Header.Formatting.AutoFormat {
			h = utils.Title(strings.Join(utils.SplitCamelCase(h), " "))
		}
		effectiveMaxWidth := t.config.MaxWidth
		if t.config.Header.Formatting.MaxWidth > 0 {
			effectiveMaxWidth = t.config.Header.Formatting.MaxWidth
		}
		if colMaxWidth, ok := t.config.Header.ColMaxWidths[i]; ok && colMaxWidth > 0 {
			effectiveMaxWidth = colMaxWidth
		}
		if effectiveMaxWidth == 0 {
			effectiveMaxWidth = 30 // Default for wrapping
		}
		if t.config.Header.Formatting.AutoWrap && effectiveMaxWidth > 0 {
			padLeftWidth := utils.RuneWidth(t.config.Header.Padding.Global.Left)
			padRightWidth := utils.RuneWidth(t.config.Header.Padding.Global.Right)
			lines, wrappedWidth := utils.WrapString(h, effectiveMaxWidth-padLeftWidth-padRightWidth)
			fmt.Println("DEBUG: SetHeader col", i, "content:", h, "wrapped lines:", lines, "wrapped width:", wrappedWidth)
			if t.config.Header.Formatting.Truncate {
				h = lines[0]
			} else {
				h = strings.Join(lines, "\n")
			}
		}
		formatted[i] = h
	}
	if t.config.Header.Filter != nil {
		formatted = t.config.Header.Filter(formatted)
	}
	t.headers = formatted
	t.parseDimension(t.headers, renderer.Header)
	fmt.Println("DEBUG: SetHeader widths:", t.headerWidths)
}

func (t *Table) SetFooter(footers []string) {
	t.ensureInitialized()
	numCols := len(t.headerWidths)
	if len(t.rowWidths) > numCols {
		numCols = len(t.rowWidths)
	}
	padded := make([]string, numCols)
	effectiveMaxWidth := t.config.MaxWidth
	if t.config.Footer.Formatting.MaxWidth > 0 {
		effectiveMaxWidth = t.config.Footer.Formatting.MaxWidth
	}
	if t.config.Footer.Formatting.AutoWrap && effectiveMaxWidth > 0 {
		for i, f := range footers {
			if i < numCols {
				colMaxWidth := effectiveMaxWidth
				if customMax, ok := t.config.Footer.ColMaxWidths[i]; ok && customMax > 0 {
					colMaxWidth = customMax
				}
				lines, wrappedWidth := utils.WrapString(f, colMaxWidth-utils.RuneWidth(t.config.Footer.Padding.Global.Left)-utils.RuneWidth(t.config.Footer.Padding.Global.Right))
				fmt.Println("DEBUG: SetFooter col", i, "content:", f, "wrapped lines:", lines, "wrapped width:", wrappedWidth)
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
	fmt.Println("DEBUG: SetFooter widths:", t.footerWidths)
}

func (t *Table) Render() error {
	t.ensureInitialized()
	t.headerWidths = make(map[int]int)
	t.rowWidths = make(map[int]int)
	t.footerWidths = make(map[int]int)

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

	numCols := 0
	if len(t.headers) > numCols {
		numCols = len(t.headers)
	}
	for _, lines := range t.rows {
		for _, line := range lines {
			if len(line) > numCols {
				numCols = len(line)
			}
		}
	}
	if len(t.footers) > numCols {
		numCols = len(t.footers)
	}

	for i := 0; i < numCols; i++ {
		hWidth := t.headerWidths[i]
		rWidth := t.rowWidths[i]
		fWidth := t.footerWidths[i]
		maxWidth := hWidth
		if rWidth > maxWidth {
			maxWidth = rWidth
		}
		if fWidth > maxWidth {
			maxWidth = fWidth
		}
		t.headerWidths[i] = maxWidth
		t.rowWidths[i] = maxWidth
		t.footerWidths[i] = maxWidth
	}

	fmt.Println("DEBUG: Normalized widths - header:", t.headerWidths, "row:", t.rowWidths, "footer:", t.footerWidths)

	f := t.renderer

	if len(t.headers) > 0 {
		f.Header(t.writer, t.headers, renderer.Formatting{
			Widths:       t.headerWidths,
			Align:        t.config.Header.Formatting.Alignment,
			Padding:      t.config.Header.Padding.Global,
			ColPadding:   make(map[int]symbols.Padding),
			ColAligns:    make(map[int]string),
			MaxWidth:     t.config.MaxWidth,
			ColMaxWidths: t.config.Header.ColMaxWidths,
			AutoWrap:     t.config.Header.Formatting.AutoWrap,
			HasFooter:    len(t.footers) > 0,
		})
	}

	for i, lines := range t.rows {
		for j, line := range lines {
			colPadding := make(map[int]symbols.Padding)
			colAligns := make(map[int]string)
			for colKey := range line {
				colPadding[colKey] = t.config.Row.Padding.Global
				if colKey < len(t.config.Row.Padding.PerColumn) && t.config.Row.Padding.PerColumn[colKey] != (symbols.Padding{}) {
					colPadding[colKey] = t.config.Row.Padding.PerColumn[colKey]
				}
				if colKey < len(t.config.Row.ColumnAligns) && t.config.Row.ColumnAligns[colKey] != "" {
					colAligns[colKey] = t.config.Row.ColumnAligns[colKey]
				} else {
					colAligns[colKey] = t.config.Row.Formatting.Alignment
				}
			}
			lastRow := (i == len(t.rows)-1 && j == len(lines)-1)
			f.Row(t.writer, line, renderer.Formatting{
				Widths:       t.rowWidths,
				Level:        renderer.Middle,
				Align:        t.config.Row.Formatting.Alignment,
				Padding:      t.config.Row.Padding.Global,
				ColPadding:   colPadding,
				ColAligns:    colAligns,
				IsFirst:      i == 0 && j == 0,
				IsLast:       lastRow,
				MaxWidth:     t.config.MaxWidth,
				ColMaxWidths: t.config.Row.ColMaxWidths,
				AutoWrap:     t.config.Row.Formatting.AutoWrap,
				HasFooter:    len(t.footers) > 0,
			})
		}
	}

	if len(t.footers) > 0 {
		colPadding := make(map[int]symbols.Padding)
		colAligns := make(map[int]string)
		for i := range t.footers {
			colPadding[i] = t.config.Footer.Padding.Global
			if i < len(t.config.Footer.Padding.PerColumn) && t.config.Footer.Padding.PerColumn[i] != (symbols.Padding{}) {
				colPadding[i] = t.config.Footer.Padding.PerColumn[i]
			}
			if i < len(t.config.Footer.ColumnAligns) && t.config.Footer.ColumnAligns[i] != "" {
				colAligns[i] = t.config.Footer.ColumnAligns[i]
			} else {
				colAligns[i] = t.config.Footer.Formatting.Alignment
			}
		}
		f.Footer(t.writer, t.footers, renderer.Formatting{
			Widths:       t.footerWidths,
			Align:        t.config.Footer.Formatting.Alignment,
			Padding:      t.config.Footer.Padding.Global,
			ColPadding:   colPadding,
			ColAligns:    colAligns,
			MaxWidth:     t.config.MaxWidth,
			ColMaxWidths: t.config.Footer.ColMaxWidths,
			AutoWrap:     t.config.Footer.Formatting.AutoWrap,
			HasFooter:    true,
		})
	}

	t.hasPrinted = true
	return nil
}

func (t *Table) ensureInitialized() {
	if t.headerWidths == nil {
		t.headerWidths = make(map[int]int)
	}
	if t.rowWidths == nil {
		t.rowWidths = make(map[int]int)
	}
	if t.footerWidths == nil {
		t.footerWidths = make(map[int]int)
	}
	if t.rowHeights == nil {
		t.rowHeights = make(map[int]int)
	}
	if t.renderer == nil {
		t.renderer = renderer.NewDefault()
	}
}

func (t *Table) toStringLines(row interface{}) ([][]string, error) {
	switch v := row.(type) {
	case []string:
		result := make([][]string, len(v))
		for i, cell := range v {
			effectiveMaxWidth := t.config.MaxWidth
			if t.config.Row.Formatting.MaxWidth > 0 {
				effectiveMaxWidth = t.config.Row.Formatting.MaxWidth
			}
			if colMaxWidth, ok := t.config.Row.ColMaxWidths[i]; ok && colMaxWidth > 0 {
				effectiveMaxWidth = colMaxWidth
			}
			var lines []string
			if t.config.Row.Formatting.AutoWrap && effectiveMaxWidth > 0 {
				wrappedLines, wrappedWidth := utils.WrapString(cell, effectiveMaxWidth-utils.RuneWidth(t.config.Row.Padding.Global.Left)-utils.RuneWidth(t.config.Row.Padding.Global.Right))
				fmt.Println("DEBUG: toStringLines col", i, "content:", cell, "wrapped lines:", wrappedLines, "wrapped width:", wrappedWidth)
				lines = wrappedLines
			} else {
				lines = strings.Split(cell, "\n")
			}
			// Remove padding here; let formatCell handle it
			result[i] = lines
			fmt.Println("DEBUG: toStringLines cell", i, "lines:", lines)
		}
		normalized := t.normalizeLines(result)
		fmt.Println("DEBUG: toStringLines normalized:", normalized)
		return normalized, nil
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
				result[i][j] = ""
			}
		}
	}
	return result
}

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

func (t *Table) parseDimension(row []string, position renderer.Position) int {
	var targetWidths map[int]int
	var padding symbols.Padding
	switch position {
	case renderer.Header:
		targetWidths = t.headerWidths
		padding = t.config.Header.Padding.Global
	case renderer.Footer:
		targetWidths = t.footerWidths
		padding = t.config.Footer.Padding.Global
	default:
		targetWidths = t.rowWidths
		padding = t.config.Row.Padding.Global
	}

	maxWidth := 0
	for i, cell := range row {
		lines := strings.Split(cell, "\n")
		cellWidth := 0
		for _, line := range lines {
			lineWidth := utils.RuneWidth(line)
			if lineWidth > cellWidth {
				cellWidth = lineWidth
			}
		}

		totalWidth := cellWidth + utils.RuneWidth(padding.Left) + utils.RuneWidth(padding.Right)
		if totalWidth > maxWidth {
			maxWidth = totalWidth
		}
		if current, exists := targetWidths[i]; !exists || totalWidth > current {
			targetWidths[i] = totalWidth
		}
	}
	return maxWidth
}
