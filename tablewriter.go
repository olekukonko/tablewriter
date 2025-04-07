package tablewriter

import (
	"fmt"
	"io"
	"reflect"
	"strings"

	"github.com/olekukonko/errors"
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
	MaxWidth int
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
	headers      [][]string
	footers      [][]string
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
	return func(t *Table) { t.SetHeader(headers) }
}

func WithFooter(footers []string) Option {
	return func(t *Table) { t.SetFooter(footers) }
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
	lines, err := t.toStringLines(row, t.config.Row)
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
		t.parseDimension(line, renderer.Row, t.config.Row.Padding)
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
	formatted := make([][]string, len(headers))
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

		contentWidth := effectiveMaxWidth - utils.RuneWidth(t.config.Header.Padding.Global.Left) - utils.RuneWidth(t.config.Header.Padding.Global.Right)
		if contentWidth < 1 {
			contentWidth = 1
		}

		var lines []string
		if t.config.Header.Formatting.AutoWrap && effectiveMaxWidth > 0 {
			wrappedLines, _ := utils.WrapString(h, contentWidth)
			if t.config.Header.Formatting.Truncate && len(wrappedLines) > 0 {
				lines = []string{wrappedLines[0]}
				fmt.Printf("DEBUG: Truncated header[%d]: %s -> %s (contentWidth=%d)\n", i, h, lines[0], contentWidth)
			} else {
				lines = wrappedLines
				fmt.Printf("DEBUG: Wrapped header[%d]: %s -> %v (contentWidth=%d)\n", i, h, lines, contentWidth)
			}
		} else {
			lines = []string{h}
			fmt.Printf("DEBUG: Unwrapped header[%d]: %s (contentWidth=%d)\n", i, h, contentWidth)
		}
		formatted[i] = lines
	}

	if t.config.Header.Filter != nil {
		for i := range formatted {
			formatted[i] = t.config.Header.Filter(formatted[i])
		}
	}
	t.headers = formatted
	for _, lines := range formatted {
		t.parseDimension(lines, renderer.Header, t.config.Header.Padding)
	}
}

func (t *Table) SetFooter(footers []string) {
	t.ensureInitialized()
	numCols := len(t.headerWidths)
	if len(t.rowWidths) > numCols {
		numCols = len(t.rowWidths)
	}
	padded := make([][]string, numCols)
	effectiveMaxWidth := t.config.MaxWidth
	if t.config.Footer.Formatting.MaxWidth > 0 {
		effectiveMaxWidth = t.config.Footer.Formatting.MaxWidth
	}

	for i, f := range footers {
		if i >= numCols {
			break
		}
		colMaxWidth := effectiveMaxWidth
		if customMax, ok := t.config.Footer.ColMaxWidths[i]; ok && customMax > 0 {
			colMaxWidth = customMax
		}
		contentWidth := colMaxWidth - utils.RuneWidth(t.config.Footer.Padding.Global.Left) - utils.RuneWidth(t.config.Footer.Padding.Global.Right)
		if contentWidth < 1 {
			contentWidth = 1
		}
		var lines []string
		if t.config.Footer.Formatting.AutoWrap && colMaxWidth > 0 {
			wrappedLines, _ := utils.WrapString(f, contentWidth)
			if t.config.Footer.Formatting.Truncate && len(wrappedLines) > 0 {
				lines = []string{wrappedLines[0]}
			} else {
				lines = wrappedLines
			}
		} else {
			lines = []string{f}
		}
		padded[i] = lines
	}
	for i := len(footers); i < numCols; i++ {
		padded[i] = []string{""}
	}

	if t.config.Footer.Filter != nil {
		for i := range padded {
			padded[i] = t.config.Footer.Filter(padded[i])
		}
	}
	t.footers = padded
	for _, lines := range padded {
		t.parseDimension(lines, renderer.Footer, t.config.Footer.Padding)
	}
}

func (t *Table) Render() error {
	t.ensureInitialized()
	t.headerWidths = make(map[int]int)
	t.rowWidths = make(map[int]int)
	t.footerWidths = make(map[int]int)

	if len(t.headers) > 0 {
		for _, lines := range t.headers {
			t.parseDimension(lines, renderer.Header, t.config.Header.Padding)
		}
	}
	for _, lines := range t.rows {
		for _, line := range lines {
			t.parseDimension(line, renderer.Row, t.config.Row.Padding)
		}
	}
	if len(t.footers) > 0 {
		for _, lines := range t.footers {
			t.parseDimension(lines, renderer.Footer, t.config.Footer.Padding)
		}
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
		if i < len(t.headers) && len(t.headers[i]) > 0 {
			for _, h := range t.headers[i] {
				minWidth := utils.RuneWidth(strings.TrimSpace(h)) + utils.RuneWidth(t.config.Header.Padding.Global.Left) + utils.RuneWidth(t.config.Header.Padding.Global.Right)
				if minWidth > maxWidth {
					maxWidth = minWidth
				}
			}
		}
		for _, lines := range t.rows {
			if i < len(lines) {
				for _, r := range lines {
					if i < len(r) {
						minWidth := utils.RuneWidth(strings.TrimSpace(r[i])) + utils.RuneWidth(t.config.Row.Padding.Global.Left) + utils.RuneWidth(t.config.Row.Padding.Global.Right)
						if minWidth > maxWidth {
							maxWidth = minWidth
						}
					}
				}
			}
		}
		if i < len(t.footers) && len(t.footers[i]) > 0 {
			for _, f := range t.footers[i] {
				minWidth := utils.RuneWidth(strings.TrimSpace(f)) + utils.RuneWidth(t.config.Footer.Padding.Global.Left) + utils.RuneWidth(t.config.Footer.Padding.Global.Right)
				if minWidth > maxWidth {
					maxWidth = minWidth
				}
			}
		}
		fmt.Printf("DEBUG: Column %d width - hWidth=%d, rWidth=%d, fWidth=%d, maxWidth=%d\n", i, hWidth, rWidth, fWidth, maxWidth)
		t.headerWidths[i] = maxWidth
		t.rowWidths[i] = maxWidth
		t.footerWidths[i] = maxWidth
	}

	fmt.Println("DEBUG: Normalized widths - header:", t.headerWidths, "row:", t.rowWidths, "footer:", t.footerWidths)

	f := t.renderer

	if len(t.headers) > 0 {
		colAligns := make(map[int]string)
		for i := 0; i < numCols; i++ {
			if i < len(t.config.Header.ColumnAligns) && t.config.Header.ColumnAligns[i] != "" {
				colAligns[i] = t.config.Header.ColumnAligns[i]
			} else {
				colAligns[i] = t.config.Header.Formatting.Alignment
			}
		}
		f.Header(t.writer, t.headers, renderer.Formatting{
			Widths:     t.headerWidths,
			Padding:    t.config.Header.Padding.Global,
			ColPadding: make(map[int]symbols.Padding),
			ColAligns:  colAligns,
			HasFooter:  len(t.footers) > 0,
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
			lastRow := i == len(t.rows)-1 && j == len(lines)-1
			f.Row(t.writer, line, renderer.Formatting{
				Widths:     t.rowWidths,
				Level:      renderer.Middle,
				Padding:    t.config.Row.Padding.Global,
				ColPadding: colPadding,
				ColAligns:  colAligns,
				IsFirst:    i == 0 && j == 0,
				IsLast:     lastRow,
				HasFooter:  len(t.footers) > 0,
			})
		}
	}

	if len(t.footers) > 0 {
		colAligns := make(map[int]string)
		for i := 0; i < numCols; i++ {
			if i < len(t.config.Footer.ColumnAligns) && t.config.Footer.ColumnAligns[i] != "" {
				colAligns[i] = t.config.Footer.ColumnAligns[i]
			} else {
				colAligns[i] = t.config.Footer.Formatting.Alignment
			}
		}
		f.Footer(t.writer, t.footers, renderer.Formatting{
			Widths:     t.footerWidths,
			Padding:    t.config.Footer.Padding.Global,
			ColPadding: make(map[int]symbols.Padding),
			ColAligns:  colAligns,
			HasFooter:  true,
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

func (t *Table) toStringLines(row interface{}, config CellConfig) ([][]string, error) {
	switch v := row.(type) {
	case []string:
		result := make([][]string, len(v))
		for i, cell := range v {
			effectiveMaxWidth := t.config.MaxWidth
			if config.Formatting.MaxWidth > 0 {
				effectiveMaxWidth = config.Formatting.MaxWidth
			}
			if colMaxWidth, ok := config.ColMaxWidths[i]; ok && colMaxWidth > 0 {
				effectiveMaxWidth = colMaxWidth
			}
			contentWidth := effectiveMaxWidth - utils.RuneWidth(config.Padding.Global.Left) - utils.RuneWidth(config.Padding.Global.Right)
			if contentWidth < 1 {
				contentWidth = 1
			}
			var lines []string
			if config.Formatting.AutoWrap && effectiveMaxWidth > 0 {
				wrappedLines, _ := utils.WrapString(cell, contentWidth)
				if config.Formatting.Truncate && len(wrappedLines) > 0 {
					lines = []string{wrappedLines[0]}
				} else {
					lines = wrappedLines
				}
			} else {
				lines = strings.Split(cell, "\n")
			}
			result[i] = lines
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
		return t.toStringLines(out[0].Interface().([]string), config)
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

func (t *Table) parseDimension(row []string, position renderer.Position, padding CellPadding) int {
	var targetWidths map[int]int

	switch position {
	case renderer.Header:
		targetWidths = t.headerWidths
	case renderer.Footer:
		targetWidths = t.footerWidths
	default:
		targetWidths = t.rowWidths
	}

	maxWidth := 0
	for i, cell := range row {
		padLeftWidth := utils.RuneWidth(padding.Global.Left)
		padRightWidth := utils.RuneWidth(padding.Global.Right)
		if i < len(padding.PerColumn) && padding.PerColumn[i] != (symbols.Padding{}) {
			padLeftWidth = utils.RuneWidth(padding.PerColumn[i].Left)
			padRightWidth = utils.RuneWidth(padding.PerColumn[i].Right)
		}
		cellWidth := utils.RuneWidth(strings.TrimSpace(cell))
		// Ensure minimum width includes content + 1 left + 1 right padding
		totalWidth := cellWidth + padLeftWidth + padRightWidth
		if totalWidth > maxWidth {
			maxWidth = totalWidth
		}
		if current, exists := targetWidths[i]; !exists || totalWidth > current {
			targetWidths[i] = totalWidth
		}
	}
	return maxWidth
}
