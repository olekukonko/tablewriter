// tablewriter/tablewriter.go
package tablewriter

import (
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

const (
	WrapNone     = iota // No wrapping
	WrapNormal          // Normal wrapping around word boundaries
	WrapTruncate        // Truncate content with ellipsis
	WrapBreak           // Break at character boundaries
)

const (
	CharEllipsis = "…"
	CharBreak    = "↩"
)

type CellFormatting struct {
	Alignment  string
	AutoWrap   int
	AutoFormat bool
	AutoMerge  bool
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
				AutoWrap:   WrapTruncate,
				Alignment:  renderer.AlignCenter,
				AutoFormat: true,
			},
			Padding: CellPadding{
				Global: defaultPadding,
			},
		},
		Row: CellConfig{
			Formatting: CellFormatting{
				MaxWidth:   0,
				AutoWrap:   WrapNormal,
				Alignment:  renderer.AlignLeft,
				AutoFormat: false,
			},
			Padding: CellPadding{
				Global: defaultPadding,
			},
		},
		Footer: CellConfig{
			Formatting: CellFormatting{
				MaxWidth:   0,
				AutoWrap:   WrapNormal,
				Alignment:  renderer.AlignRight,
				AutoFormat: false,
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
	return func(t *Table) {
		t.config = mergeConfig(defaultConfig(), cfg)
	}
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
	t.rows = append(t.rows, lines)
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
	prepared := t.prepareContent(headers, t.config.Header)
	t.headers = prepared
}

func (t *Table) SetFooter(footers []string) {
	t.ensureInitialized()
	numCols := t.maxColumns()
	prepared := t.prepareContent(footers, t.config.Footer)
	if len(prepared) > 0 && len(prepared[0]) < numCols {
		for i := range prepared {
			for len(prepared[i]) < numCols {
				prepared[i] = append(prepared[i], "")
			}
		}
	}
	t.footers = prepared
}

func (t *Table) Render() error {
	t.ensureInitialized()

	// Calculate widths
	t.headerWidths = make(map[int]int)
	t.rowWidths = make(map[int]int)
	t.footerWidths = make(map[int]int)

	for _, lines := range t.headers {
		t.updateWidths(lines, t.headerWidths, t.config.Header.Padding)
	}
	for _, row := range t.rows {
		for _, line := range row {
			t.updateWidths(line, t.rowWidths, t.config.Row.Padding)
		}
	}
	for _, lines := range t.footers {
		t.updateWidths(lines, t.footerWidths, t.config.Footer.Padding)
	}

	// Normalize widths
	numCols := t.maxColumns()
	for i := 0; i < numCols; i++ {
		maxWidth := 0
		for _, w := range []map[int]int{t.headerWidths, t.rowWidths, t.footerWidths} {
			if w[i] > maxWidth {
				maxWidth = w[i]
			}
		}
		t.headerWidths[i] = maxWidth
		t.rowWidths[i] = maxWidth
		t.footerWidths[i] = maxWidth
	}

	// Add padding lines after width calculation
	headerLines := t.addPaddingLines(t.headers, t.config.Header, renderer.Header)
	rowLines := make([][][]string, len(t.rows))
	for i, row := range t.rows {
		rowLines[i] = t.addPaddingLines(row, t.config.Row, renderer.Row)
	}
	footerLines := t.addPaddingLines(t.footers, t.config.Footer, renderer.Footer)

	f := t.renderer

	if len(headerLines) > 0 {
		colAligns := t.buildAligns(t.config.Header)
		f.Header(t.writer, headerLines, renderer.Formatting{
			Widths:     t.headerWidths,
			Padding:    t.config.Header.Padding.Global,
			ColPadding: t.buildPadding(t.config.Header.Padding),
			ColAligns:  colAligns,
			HasFooter:  len(t.footers) > 0,
		})
	}

	for i, lines := range rowLines {
		colAligns := t.buildAligns(t.config.Row)
		colPadding := t.buildPadding(t.config.Row.Padding)
		for j, line := range lines {
			f.Row(t.writer, line, renderer.Formatting{
				Widths:     t.rowWidths,
				Padding:    t.config.Row.Padding.Global,
				ColPadding: colPadding,
				ColAligns:  colAligns,
				IsFirst:    i == 0 && j == 0,
				IsLast:     i == len(t.rows)-1 && j == len(lines)-1,
				HasFooter:  len(t.footers) > 0,
			})
		}
	}

	if len(footerLines) > 0 {
		colAligns := t.buildAligns(t.config.Footer)
		f.Footer(t.writer, footerLines, renderer.Formatting{
			Widths:     t.footerWidths,
			Padding:    t.config.Footer.Padding.Global,
			ColPadding: t.buildPadding(t.config.Footer.Padding),
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
	if t.renderer == nil {
		t.renderer = renderer.NewDefault()
	}
}

func (t *Table) toStringLines(row interface{}, config CellConfig) ([][]string, error) {
	var cells []string
	switch v := row.(type) {
	case []string:
		cells = v
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
		cells = out[0].Interface().([]string)
	}

	if config.Filter != nil {
		cells = config.Filter(cells)
	}

	return t.prepareContent(cells, config), nil
}

func (t *Table) prepareContent(cells []string, config CellConfig) [][]string {
	result := make([][]string, 0)
	numCols := len(cells)

	for i, cell := range cells {
		effectiveMaxWidth := t.config.MaxWidth
		if config.Formatting.MaxWidth > 0 {
			effectiveMaxWidth = config.Formatting.MaxWidth
		}
		if colMaxWidth, ok := config.ColMaxWidths[i]; ok && colMaxWidth > 0 {
			effectiveMaxWidth = colMaxWidth
		}

		padLeftWidth := utils.RuneWidth(config.Padding.Global.Left)
		padRightWidth := utils.RuneWidth(config.Padding.Global.Right)
		if i < len(config.Padding.PerColumn) && config.Padding.PerColumn[i] != (symbols.Padding{}) {
			padLeftWidth = utils.RuneWidth(config.Padding.PerColumn[i].Left)
			padRightWidth = utils.RuneWidth(config.Padding.PerColumn[i].Right)
		}

		contentWidth := effectiveMaxWidth - padLeftWidth - padRightWidth
		if contentWidth < 1 {
			contentWidth = 1
		}

		if config.Formatting.AutoFormat {
			cell = utils.Title(strings.Join(utils.SplitCamelCase(cell), " "))
		}

		// Split on newlines first
		lines := strings.Split(cell, "\n")
		finalLines := make([]string, 0)

		for _, line := range lines {
			if effectiveMaxWidth > 0 {
				switch config.Formatting.AutoWrap {
				case WrapNormal:
					wrapped, _ := utils.WrapString(line, contentWidth)
					finalLines = append(finalLines, wrapped...)
				case WrapTruncate:
					if utils.RuneWidth(line) > contentWidth {
						finalLines = append(finalLines, utils.TruncateString(line, contentWidth-1)+CharEllipsis)
					} else {
						finalLines = append(finalLines, line)
					}
				case WrapBreak:
					wrapped := make([]string, 0)
					for len(line) > contentWidth {
						wrapped = append(wrapped, line[:contentWidth]+CharBreak)
						line = line[contentWidth:]
					}
					if len(line) > 0 {
						wrapped = append(wrapped, line)
					}
					finalLines = append(finalLines, wrapped...)
				default: // WrapNone
					finalLines = append(finalLines, line)
				}
			} else {
				finalLines = append(finalLines, line)
			}
		}

		for len(result) < len(finalLines) {
			newRow := make([]string, numCols)
			for j := range newRow {
				newRow[j] = ""
			}
			result = append(result, newRow)
		}

		for j, line := range finalLines {
			result[j][i] = line
		}
	}

	return result
}

func (t *Table) addPaddingLines(content [][]string, config CellConfig, position renderer.Position) [][]string {
	if len(content) == 0 {
		return content
	}

	result := make([][]string, 0)
	numCols := len(content[0])

	if config.Padding.Global.Top != "" {
		topPadding := make([]string, numCols)
		for i := range topPadding {
			var padWidth int
			switch position {
			case renderer.Header:
				padWidth = t.headerWidths[i]
			case renderer.Row:
				padWidth = t.rowWidths[i]
			case renderer.Footer:
				padWidth = t.footerWidths[i]
			}
			if padWidth == 0 {
				padWidth = utils.RuneWidth(config.Padding.Global.Top)
			}
			repeatCount := (padWidth + utils.RuneWidth(config.Padding.Global.Top) - 1) / utils.RuneWidth(config.Padding.Global.Top)
			if repeatCount < 1 {
				repeatCount = 1
			}
			topPadding[i] = strings.Repeat(config.Padding.Global.Top, repeatCount)
		}
		result = append(result, topPadding)
	}

	result = append(result, content...)

	if config.Padding.Global.Bottom != "" {
		bottomPadding := make([]string, numCols)
		for i := range bottomPadding {
			var padWidth int
			switch position {
			case renderer.Header:
				padWidth = t.headerWidths[i]
			case renderer.Row:
				padWidth = t.rowWidths[i]
			case renderer.Footer:
				padWidth = t.footerWidths[i]
			}
			if padWidth == 0 {
				padWidth = utils.RuneWidth(config.Padding.Global.Bottom)
			}
			repeatCount := (padWidth + utils.RuneWidth(config.Padding.Global.Bottom) - 1) / utils.RuneWidth(config.Padding.Global.Bottom)
			if repeatCount < 1 {
				repeatCount = 1
			}
			bottomPadding[i] = strings.Repeat(config.Padding.Global.Bottom, repeatCount)
		}
		result = append(result, bottomPadding)
	}

	return result
}

func (t *Table) updateWidths(row []string, widths map[int]int, padding CellPadding) {
	for i, cell := range row {
		padLeftWidth := utils.RuneWidth(padding.Global.Left)
		padRightWidth := utils.RuneWidth(padding.Global.Right)
		if i < len(padding.PerColumn) && padding.PerColumn[i] != (symbols.Padding{}) {
			padLeftWidth = utils.RuneWidth(padding.PerColumn[i].Left)
			padRightWidth = utils.RuneWidth(padding.PerColumn[i].Right)
		}
		totalWidth := utils.RuneWidth(strings.TrimSpace(cell)) + padLeftWidth + padRightWidth
		if current, exists := widths[i]; !exists || totalWidth > current {
			widths[i] = totalWidth
		}
	}
}

func (t *Table) maxColumns() int {
	max := 0
	if len(t.headers) > 0 && len(t.headers[0]) > max {
		max = len(t.headers[0])
	}
	for _, row := range t.rows {
		if len(row) > 0 && len(row[0]) > max {
			max = len(row[0])
		}
	}
	if len(t.footers) > 0 && len(t.footers[0]) > max {
		max = len(t.footers[0])
	}
	return max
}

func (t *Table) buildAligns(config CellConfig) map[int]string {
	colAligns := make(map[int]string)
	numCols := t.maxColumns()
	for i := 0; i < numCols; i++ {
		if i < len(config.ColumnAligns) && config.ColumnAligns[i] != "" {
			colAligns[i] = config.ColumnAligns[i]
		} else {
			colAligns[i] = config.Formatting.Alignment
		}
	}
	return colAligns
}

func (t *Table) buildPadding(padding CellPadding) map[int]symbols.Padding {
	colPadding := make(map[int]symbols.Padding)
	numCols := t.maxColumns()
	for i := 0; i < numCols; i++ {
		if i < len(padding.PerColumn) && padding.PerColumn[i] != (symbols.Padding{}) {
			colPadding[i] = padding.PerColumn[i]
		} else {
			colPadding[i] = padding.Global
		}
	}
	return colPadding
}

// Add this merge function to handle nested struct merging
func mergeConfig(dst, src Config) Config {
	// Merge MaxWidth
	if src.MaxWidth != 0 {
		dst.MaxWidth = src.MaxWidth
	}

	// Merge Header config
	dst.Header = mergeCellConfig(dst.Header, src.Header)

	// Merge Row config
	dst.Row = mergeCellConfig(dst.Row, src.Row)

	// Merge Footer config
	dst.Footer = mergeCellConfig(dst.Footer, src.Footer)

	return dst
}

func mergeCellConfig(dst, src CellConfig) CellConfig {
	// Merge Formatting
	if src.Formatting.Alignment != "" {
		dst.Formatting.Alignment = src.Formatting.Alignment
	}
	if src.Formatting.AutoWrap != 0 {
		dst.Formatting.AutoWrap = src.Formatting.AutoWrap
	}
	if src.Formatting.MaxWidth != 0 {
		dst.Formatting.MaxWidth = src.Formatting.MaxWidth
	}
	dst.Formatting.AutoFormat = src.Formatting.AutoFormat
	dst.Formatting.AutoMerge = src.Formatting.AutoMerge

	// Merge Padding
	if src.Padding.Global != (symbols.Padding{}) {
		dst.Padding.Global = src.Padding.Global
	}
	if len(src.Padding.PerColumn) > 0 {
		if dst.Padding.PerColumn == nil {
			dst.Padding.PerColumn = make([]symbols.Padding, len(src.Padding.PerColumn))
		}
		for i, pad := range src.Padding.PerColumn {
			if pad != (symbols.Padding{}) {
				dst.Padding.PerColumn[i] = pad
			}
		}
	}

	// Merge Callbacks
	if src.Callbacks.Global != nil {
		dst.Callbacks.Global = src.Callbacks.Global
	}
	if len(src.Callbacks.PerColumn) > 0 {
		if dst.Callbacks.PerColumn == nil {
			dst.Callbacks.PerColumn = make([]func(), len(src.Callbacks.PerColumn))
		}
		for i, cb := range src.Callbacks.PerColumn {
			if cb != nil {
				dst.Callbacks.PerColumn[i] = cb
			}
		}
	}

	// Merge Filter
	if src.Filter != nil {
		dst.Filter = src.Filter
	}

	// Merge ColumnAligns
	if len(src.ColumnAligns) > 0 {
		dst.ColumnAligns = src.ColumnAligns
	}

	// Merge ColMaxWidths
	if len(src.ColMaxWidths) > 0 {
		if dst.ColMaxWidths == nil {
			dst.ColMaxWidths = make(map[int]int)
		}
		for k, v := range src.ColMaxWidths {
			dst.ColMaxWidths[k] = v
		}
	}

	return dst
}
