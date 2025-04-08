package tablewriter

import (
	"fmt"
	"github.com/olekukonko/errors"
	"github.com/olekukonko/tablewriter/renderer"
	"github.com/olekukonko/tablewriter/tw"
	"github.com/olekukonko/tablewriter/twfn"
	"io"
	"reflect"
	"strings"
)

// Filter defines a function type for processing cell content
type Filter func([]string) []string

// CellFormatting holds formatting options for table cells
type CellFormatting struct {
	Alignment  tw.Align // Text alignment within the cell
	AutoWrap   int      // Wrapping behavior (e.g., truncate, normal)
	AutoFormat bool     // Whether to automatically format text (e.g., title case)
	MaxWidth   int      // Maximum width of the cell content
	MergeMode  int      // Cell merging behavior (e.g., vertical, horizontal)
}

// CellPadding defines padding settings for cells
type CellPadding struct {
	Global    tw.Padding   // Default padding applied to all cells
	PerColumn []tw.Padding // Column-specific padding overrides
}

// CellCallbacks holds callback functions for cell processing
type CellCallbacks struct {
	Global    func()   // Global callback applied to all cells
	PerColumn []func() // Column-specific callbacks
}

// CellConfig combines formatting, padding, and callback settings for a table section
type CellConfig struct {
	Formatting   CellFormatting // Formatting options
	Padding      CellPadding    // Padding settings
	Callbacks    CellCallbacks  // Callback functions
	Filter       Filter         // Filter function for cell content
	ColumnAligns []tw.Align     // Per-column alignment overrides
	ColMaxWidths map[int]int    // Per-column maximum width overrides
}

// Config represents the overall table configuration
type Config struct {
	MaxWidth int        // Maximum width of the entire table
	Header   CellConfig // Configuration for header section
	Row      CellConfig // Configuration for row section
	Footer   CellConfig // Configuration for footer section
	Debug    bool       // Enable debug logging
}

// defaultConfig returns a default table configuration
func defaultConfig() Config {
	defaultPadding := tw.Padding{Left: tw.Space, Right: tw.Space, Top: tw.Empty, Bottom: tw.Empty}
	return Config{
		MaxWidth: 0,
		Header: CellConfig{
			Formatting: CellFormatting{
				MaxWidth:   0,
				AutoWrap:   tw.WrapTruncate,
				Alignment:  tw.AlignCenter,
				AutoFormat: true,
				MergeMode:  tw.MergeNone,
			},
			Padding: CellPadding{
				Global: defaultPadding,
			},
		},
		Row: CellConfig{
			Formatting: CellFormatting{
				MaxWidth:   0,
				AutoWrap:   tw.WrapNormal,
				Alignment:  tw.AlignLeft,
				AutoFormat: false,
				MergeMode:  tw.MergeNone,
			},
			Padding: CellPadding{
				Global: defaultPadding,
			},
		},
		Footer: CellConfig{
			Formatting: CellFormatting{
				MaxWidth:   0,
				AutoWrap:   tw.WrapNormal,
				Alignment:  tw.AlignRight,
				AutoFormat: false,
				MergeMode:  tw.MergeNone,
			},
			Padding: CellPadding{
				Global: defaultPadding,
			},
		},
		Debug: true,
	}
}

// Option defines a function to configure a Table instance
type Option func(*Table)

// WithHeader sets the table headers
func WithHeader(headers []string) Option {
	return func(t *Table) { t.SetHeader(headers) }
}

// WithFooter sets the table footers
func WithFooter(footers []string) Option {
	return func(t *Table) { t.SetFooter(footers) }
}

// WithRenderer sets a custom renderer for the table
func WithRenderer(f renderer.Renderer) Option {
	return func(t *Table) { t.renderer = f }
}

// WithConfig applies a custom configuration to the table
func WithConfig(cfg Config) Option {
	return func(t *Table) { t.config = mergeConfig(defaultConfig(), cfg) }
}

// WithStringer sets a custom stringer function for row conversion
func WithStringer[T any](s func(T) []string) Option {
	return func(t *Table) { t.stringer = s }
}

// Table represents a table with content and rendering capabilities
type Table struct {
	writer       io.Writer         // Output destination
	rows         [][][]string      // Row data (multi-line cells supported)
	headers      [][]string        // Header content
	footers      [][]string        // Footer content
	headerWidths map[int]int       // Calculated widths for header columns
	rowWidths    map[int]int       // Calculated widths for row columns
	footerWidths map[int]int       // Calculated widths for footer columns
	renderer     renderer.Renderer // Rendering engine
	config       Config            // Table configuration
	stringer     any               // Function to convert rows to strings
	newLine      string            // Newline character
	hasPrinted   bool              // Tracks if table has been rendered
	trace        []string          // Debug trace log
}

// NewTable creates a new table instance with optional configurations
func NewTable(w io.Writer, opts ...Option) *Table {
	t := &Table{
		writer:       w,
		headerWidths: make(map[int]int),
		rowWidths:    make(map[int]int),
		footerWidths: make(map[int]int),
		renderer:     renderer.NewDefault(),
		config:       defaultConfig(),
		newLine:      tw.NewLine,
		trace:        make([]string, 0, 100),
	}
	for _, opt := range opts {
		opt(t)
	}
	t.debug("Table initialized with %d options", len(opts))
	return t
}

// NewWriter creates a new table with default settings
func NewWriter(w io.Writer) *Table {
	t := NewTable(w)
	t.debug("NewWriter created table")
	return t
}

// Renderer returns the current renderer
func (t *Table) Renderer() renderer.Renderer {
	t.debug("Renderer requested")
	return t.renderer
}

// Debug returns the debug trace including renderer logs
func (t *Table) Debug() []string {
	t.debug("Debug trace requested, returning %d entries", len(t.trace))
	return append(t.trace, t.renderer.Debug()...)
}

// Append adds a single row to the table
func (t *Table) Append(row interface{}) error {
	t.ensureInitialized()
	t.debug("Appending row: %v", row)
	lines, err := t.toStringLines(row, t.config.Row)
	if err != nil {
		t.debug("Error appending row: %v", err)
		return err
	}
	t.rows = append(t.rows, lines)
	t.debug("Row appended, total rows: %d", len(t.rows))
	return nil
}

// Bulk adds multiple rows to the table
func (t *Table) Bulk(rows interface{}) error {
	t.debug("Starting Bulk operation")
	rv := reflect.ValueOf(rows)
	if rv.Kind() != reflect.Slice {
		err := errors.Newf("Bulk expects a slice, got %T", rows)
		t.debug("Bulk error: %v", err)
		return err
	}
	for i := 0; i < rv.Len(); i++ {
		row := rv.Index(i).Interface()
		t.debug("Processing bulk row %d: %v", i, row)
		if err := t.Append(row); err != nil {
			t.debug("Bulk append failed at index %d: %v", i, err)
			return err
		}
	}
	t.debug("Bulk completed, processed %d rows", rv.Len())
	return nil
}

// SetHeader configures the table header
func (t *Table) SetHeader(headers []string) {
	t.ensureInitialized()
	t.debug("Setting header: %v", headers)
	prepared := t.prepareContent(headers, t.config.Header)
	t.headers = prepared
	t.debug("Header set, lines: %d", len(prepared))
}

// SetFooter configures the table footer, padding if necessary
func (t *Table) SetFooter(footers []string) {
	t.ensureInitialized()
	t.debug("Setting footer: %v", footers)
	numCols := t.maxColumns()
	prepared := t.prepareContent(footers, t.config.Footer)
	if len(prepared) > 0 && len(prepared[0]) < numCols {
		t.debug("Padding footer to %d columns", numCols)
		for i := range prepared {
			for len(prepared[i]) < numCols {
				prepared[i] = append(prepared[i], tw.Empty)
			}
		}
	}
	t.footers = prepared
	t.debug("Footer set, lines: %d", len(prepared))
}

func (t *Table) Render() error {
	return t.render()
}

func (t *Table) render() error {
	t.ensureInitialized()
	t.debug("Starting Render")

	// Calculate column widths
	t.headerWidths = make(map[int]int)
	t.rowWidths = make(map[int]int)
	t.footerWidths = make(map[int]int)
	for _, lines := range t.headers {
		t.updateWidths(lines, t.headerWidths, t.config.Header.Padding)
	}
	t.debug("Header widths calculated: %v", t.headerWidths)
	for _, row := range t.rows {
		for _, line := range row {
			t.updateWidths(line, t.rowWidths, t.config.Row.Padding)
		}
	}
	t.debug("Row widths calculated: %v", t.rowWidths)
	for _, lines := range t.footers {
		t.updateWidths(lines, t.footerWidths, t.config.Footer.Padding)
	}
	t.debug("Footer widths calculated: %v", t.footerWidths)

	// Normalize widths across sections
	numCols := t.maxColumns()
	t.debug("Normalizing widths for %d columns", numCols)
	for i := 0; i < numCols; i++ {
		maxWidth := 0
		for _, w := range []map[int]int{t.headerWidths, t.rowWidths, t.footerWidths} {
			if wd, ok := w[i]; ok && wd > maxWidth {
				maxWidth = wd
			}
		}
		t.headerWidths[i] = maxWidth
		t.rowWidths[i] = maxWidth
		t.footerWidths[i] = maxWidth
	}
	t.debug("Normalized widths: header=%v, row=%v, footer=%v", t.headerWidths, t.rowWidths, t.footerWidths)

	// Prepare content with merges
	t.debug("Preparing header content")
	headerLines, headerVertMerges, headerHorzMerges := t.prepareWithMerges(t.headers, t.config.Header, tw.Header)

	t.debug("Preparing row content for %d rows", len(t.rows))
	rowLines := make([][][]string, len(t.rows))
	rowVertMerges := make([]map[int]renderer.MergeState, len(t.rows))
	rowHorzMerges := make([]map[int]bool, len(t.rows))
	for i, row := range t.rows {
		preparedLines, vertMerges, horzMap := t.prepareWithMerges(row, t.config.Row, tw.Row)
		rowLines[i] = preparedLines
		rowVertMerges[i] = vertMerges
		rowHorzMerges[i] = horzMap
		t.debug("Row %d prepared: lines=%d", i, len(rowLines[i]))
	}

	// Apply vertical merges across all rows
	if t.config.Row.Formatting.MergeMode == tw.MergeVertical || t.config.Row.Formatting.MergeMode == tw.MergeBoth {
		t.debug("Applying vertical merges across %d rows", len(rowLines))
		previousContent := make(map[int]string)
		for i := 0; i < len(rowLines); i++ {
			for j := 0; j < len(rowLines[i]); j++ {
				for col := 0; col < numCols; col++ {
					currentVal := strings.TrimSpace(rowLines[i][j][col])
					prevVal, exists := previousContent[col]
					if exists && currentVal == prevVal && currentVal != "" {
						if _, ok := rowVertMerges[i][col]; !ok {
							rowVertMerges[i][col] = renderer.MergeState{}
						}
						mergeState := rowVertMerges[i][col]
						mergeState.Vertical = true
						rowVertMerges[i][col] = mergeState
						rowLines[i][j][col] = tw.Empty
						t.debug("Vertical merge at row %d, line %d, col %d: cleared content", i, j, col)
						for k := i - 1; k >= 0; k-- {
							if len(rowLines[k]) > 0 && strings.TrimSpace(rowLines[k][0][col]) == prevVal {
								if _, ok := rowVertMerges[k][col]; !ok {
									rowVertMerges[k][col] = renderer.MergeState{}
								}
								startMerge := rowVertMerges[k][col]
								startMerge.Vertical = true
								startMerge.Span = i - k + 1
								startMerge.Start = true
								rowVertMerges[k][col] = startMerge
								break
							}
						}
					} else if currentVal != "" {
						if _, ok := rowVertMerges[i][col]; !ok {
							rowVertMerges[i][col] = renderer.MergeState{
								Vertical: false,
								Span:     1,
								Start:    true,
								End:      false,
							}
						}
						previousContent[col] = currentVal
					}
				}
			}
		}
		for i := 0; i < len(rowLines); i++ {
			for col, mergeState := range rowVertMerges[i] {
				if mergeState.Vertical && i == len(rowLines)-1 {
					mergeState.End = true
					rowVertMerges[i][col] = mergeState
				}
			}
		}
	}

	t.debug("Preparing footer content")
	footerLines, footerVertMerges, footerHorzMerges := t.prepareWithMerges(t.footers, t.config.Footer, tw.Footer)

	// Render sections
	f := t.renderer
	cfg := f.Config()

	// Render table top border
	if cfg.Borders.Top.Enabled() && cfg.Settings.Lines.ShowTop.Enabled() {
		t.debug("Rendering table top border")
		f.Line(t.writer, renderer.Formatting{
			Row: renderer.RowContext{
				Widths:   t.headerWidths,
				Position: tw.Header,
				Location: tw.LocationFirst,
			},
			Level:    tw.LevelHeader,
			IsSubRow: false,
		})
	}

	// Header section
	if len(headerLines) > 0 {
		colAligns := t.buildAligns(t.config.Header)
		colPadding := t.buildPadding(t.config.Header.Padding)

		if t.config.Header.Padding.Global.Top != tw.Empty {
			t.debug("Rendering header top padding")
			topPadding := make([]string, numCols)
			for j := range topPadding {
				repeatCount := t.headerWidths[j] / twfn.DisplayWidth(t.config.Header.Padding.Global.Top)
				if repeatCount < 1 {
					repeatCount = 1
				}
				topPadding[j] = strings.Repeat(t.config.Header.Padding.Global.Top, repeatCount)
			}
			currentCells := make(map[int]renderer.CellContext)
			for j, cell := range topPadding {
				currentCells[j] = renderer.CellContext{
					Data:    cell,
					Align:   colAligns[j],
					Padding: colPadding[j],
					Width:   t.headerWidths[j],
					Merge:   headerVertMerges[j],
				}
			}
			f.Header(t.writer, nil, renderer.Formatting{
				Row: renderer.RowContext{
					Widths:       t.headerWidths,
					ColMaxWidths: t.config.Header.ColMaxWidths,
					Current:      currentCells,
					Position:     tw.Header,
					Location:     tw.LocationFirst,
				},
				Level:    tw.LevelHeader,
				IsSubRow: true,
			})
		}

		for i, line := range headerLines {
			currentCells := make(map[int]renderer.CellContext)
			for j, cell := range line {
				mergeState := headerVertMerges[j]
				if headerHorzMerges[j] {
					mergeState.Horizontal = true
					mergeState.Span = t.calculateHorizontalSpan(headerHorzMerges, j)
					mergeState.Start = j == 0 || !headerHorzMerges[j-1]
					mergeState.End = j == numCols-1 || !headerHorzMerges[j+1]
				}
				currentCells[j] = renderer.CellContext{
					Data:    cell,
					Align:   colAligns[j],
					Padding: colPadding[j],
					Width:   t.headerWidths[j],
					Merge:   mergeState,
				}
			}
			previousCells := make(map[int]renderer.CellContext)
			if i > 0 {
				for j, cell := range headerLines[i-1] {
					previousCells[j] = renderer.CellContext{Data: cell, Merge: headerVertMerges[j]}
				}
			}
			nextCells := make(map[int]renderer.CellContext)
			if i+1 < len(headerLines) {
				for j, cell := range headerLines[i+1] {
					nextCells[j] = renderer.CellContext{Data: cell, Merge: headerVertMerges[j]}
				}
			} else if len(rowLines) > 0 {
				for j, cell := range rowLines[0][0] {
					nextCells[j] = renderer.CellContext{Data: cell, Merge: rowVertMerges[0][j]}
				}
			}
			location := tw.LocationMiddle
			if i == 0 && t.config.Header.Padding.Global.Top == tw.Empty {
				location = tw.LocationFirst
			}
			if i == len(headerLines)-1 && t.config.Header.Padding.Global.Bottom == tw.Empty {
				location = tw.LocationEnd
			}
			f.Header(t.writer, headerLines, renderer.Formatting{
				Row: renderer.RowContext{
					Widths:       t.headerWidths,
					ColMaxWidths: t.config.Header.ColMaxWidths,
					Current:      currentCells,
					Previous:     previousCells,
					Next:         nextCells,
					Position:     tw.Header,
					Location:     location,
				},
				Level:    tw.LevelHeader,
				IsSubRow: i > 0 || t.config.Header.Padding.Global.Top != tw.Empty,
			})
		}

		if t.config.Header.Padding.Global.Bottom != tw.Empty {
			t.debug("Rendering header bottom padding")
			bottomPadding := make([]string, numCols)
			for j := range bottomPadding {
				repeatCount := t.headerWidths[j] / twfn.DisplayWidth(t.config.Header.Padding.Global.Bottom)
				if repeatCount < 1 {
					repeatCount = 1
				}
				bottomPadding[j] = strings.Repeat(t.config.Header.Padding.Global.Bottom, repeatCount)
			}
			currentCells := make(map[int]renderer.CellContext)
			for j, cell := range bottomPadding {
				currentCells[j] = renderer.CellContext{
					Data:    cell,
					Align:   colAligns[j],
					Padding: colPadding[j],
					Width:   t.headerWidths[j],
					Merge:   headerVertMerges[j],
				}
			}
			nextCells := make(map[int]renderer.CellContext)
			if len(rowLines) > 0 {
				for j, cell := range rowLines[0][0] {
					nextCells[j] = renderer.CellContext{Data: cell, Merge: rowVertMerges[0][j]}
				}
			} else if len(footerLines) > 0 {
				for j, cell := range footerLines[0] {
					nextCells[j] = renderer.CellContext{Data: cell, Merge: footerVertMerges[j]}
				}
			}
			f.Header(t.writer, nil, renderer.Formatting{
				Row: renderer.RowContext{
					Widths:       t.headerWidths,
					ColMaxWidths: t.config.Header.ColMaxWidths,
					Current:      currentCells,
					Next:         nextCells,
					Position:     tw.Header,
					Location:     tw.LocationEnd,
				},
				Level:    tw.LevelHeader,
				IsSubRow: true,
			})
		}

		if cfg.Settings.Lines.ShowHeaderLine.Enabled() && (len(rowLines) > 0 || len(footerLines) > 0) {
			t.debug("Rendering header separator line")
			currentCells := make(map[int]renderer.CellContext)
			for j := 0; j < numCols; j++ {
				mergeState := headerVertMerges[j]
				if headerHorzMerges[j] {
					mergeState.Horizontal = true
					mergeState.Span = t.calculateHorizontalSpan(headerHorzMerges, j)
					mergeState.Start = j == 0 || !headerHorzMerges[j-1]
					mergeState.End = j == numCols-1 || !headerHorzMerges[j+1]
				}
				currentCells[j] = renderer.CellContext{
					Data:  headerLines[len(headerLines)-1][j],
					Merge: mergeState,
				}
			}
			nextCells := make(map[int]renderer.CellContext)
			if len(rowLines) > 0 {
				for j, cell := range rowLines[0][0] {
					mergeState := rowVertMerges[0][j]
					if rowHorzMerges[0][j] {
						mergeState.Horizontal = true
						mergeState.Span = t.calculateHorizontalSpan(rowHorzMerges[0], j)
						mergeState.Start = j == 0 || !rowHorzMerges[0][j-1]
						mergeState.End = j == numCols-1 || !rowHorzMerges[0][j+1]
					}
					nextCells[j] = renderer.CellContext{Data: cell, Merge: mergeState}
				}
			} else if len(footerLines) > 0 {
				for j, cell := range footerLines[0] {
					nextCells[j] = renderer.CellContext{Data: cell, Merge: footerVertMerges[j]}
				}
			}
			f.Line(t.writer, renderer.Formatting{
				Row: renderer.RowContext{
					Widths:   t.headerWidths,
					Current:  currentCells,
					Next:     nextCells,
					Position: tw.Header,
					Location: tw.LocationMiddle,
				},
				Level:    tw.LevelBody,
				IsSubRow: false,
			})
		}
	}

	// Rows section
	for i, lines := range rowLines {
		colAligns := t.buildAligns(t.config.Row)
		colPadding := t.buildPadding(t.config.Row.Padding)

		if t.config.Row.Padding.Global.Top != tw.Empty {
			t.debug("Rendering row top padding for row %d", i)
			topPadding := make([]string, numCols)
			for j := range topPadding {
				repeatCount := t.rowWidths[j] / twfn.DisplayWidth(t.config.Row.Padding.Global.Top)
				if repeatCount < 1 {
					repeatCount = 1
				}
				topPadding[j] = strings.Repeat(t.config.Row.Padding.Global.Top, repeatCount)
			}
			currentCells := make(map[int]renderer.CellContext)
			for j, cell := range topPadding {
				currentCells[j] = renderer.CellContext{
					Data:    cell,
					Align:   colAligns[j],
					Padding: colPadding[j],
					Width:   t.rowWidths[j],
					Merge:   rowVertMerges[i][j],
				}
			}
			location := tw.LocationMiddle
			if i == 0 && len(t.headers) == 0 {
				location = tw.LocationFirst
			}
			f.Row(t.writer, nil, renderer.Formatting{
				Row: renderer.RowContext{
					Widths:   t.rowWidths,
					Current:  currentCells,
					Position: tw.Row,
					Location: location,
				},
				Level:    tw.LevelBody,
				IsSubRow: true,
			})
		}

		for j, line := range lines {
			currentCells := make(map[int]renderer.CellContext)
			for k := 0; k < numCols; k++ {
				mergeState := rowVertMerges[i][k]
				if rowHorzMerges[i][k] {
					mergeState.Horizontal = true
					mergeState.Span = t.calculateHorizontalSpan(rowHorzMerges[i], k)
					mergeState.Start = k == 0 || !rowHorzMerges[i][k-1]
					mergeState.End = k == numCols-1 || !rowHorzMerges[i][k+1]
				}
				currentCells[k] = renderer.CellContext{
					Data:    line[k],
					Align:   colAligns[k],
					Padding: colPadding[k],
					Width:   t.rowWidths[k],
					Merge:   mergeState,
				}
			}
			previousCells := make(map[int]renderer.CellContext)
			if i > 0 || j > 0 || t.config.Row.Padding.Global.Top != tw.Empty {
				prevLines := rowLines[twfn.Max(0, i-1)]
				prevLineIdx := len(prevLines) - 1
				if j > 0 {
					prevLineIdx = j - 1
					prevLines = lines
				} else if i == 0 && t.config.Row.Padding.Global.Top != tw.Empty {
					for k := range numCols {
						prevMerge := rowVertMerges[i][k]
						previousCells[k] = renderer.CellContext{
							Data:  strings.Repeat(t.config.Row.Padding.Global.Top, t.rowWidths[k]/twfn.DisplayWidth(t.config.Row.Padding.Global.Top)),
							Merge: prevMerge,
						}
					}
				} else if j == 0 && i > 0 {
					for k, cell := range prevLines[prevLineIdx] {
						previousCells[k] = renderer.CellContext{Data: cell, Merge: rowVertMerges[twfn.Max(0, i-1)][k]}
					}
				}
			} else if len(t.headers) > 0 {
				for k, cell := range headerLines[len(headerLines)-1] {
					previousCells[k] = renderer.CellContext{Data: cell, Merge: headerVertMerges[k]}
				}
			}
			nextCells := make(map[int]renderer.CellContext)
			if j+1 < len(lines) {
				for k, cell := range lines[j+1] {
					nextCells[k] = renderer.CellContext{Data: cell, Merge: rowVertMerges[i][k]}
				}
			} else if i+1 < len(rowLines) {
				for k, cell := range rowLines[i+1][0] {
					nextCells[k] = renderer.CellContext{Data: cell, Merge: rowVertMerges[i+1][k]}
				}
			} else if len(t.footers) > 0 {
				for k, cell := range footerLines[0] {
					nextCells[k] = renderer.CellContext{Data: cell, Merge: footerVertMerges[k]}
				}
			}
			location := tw.LocationMiddle
			if i == 0 && j == 0 && t.config.Row.Padding.Global.Top == tw.Empty && len(t.headers) == 0 {
				location = tw.LocationFirst
			}
			if i == len(t.rows)-1 && j == len(lines)-1 && t.config.Row.Padding.Global.Bottom == tw.Empty {
				location = tw.LocationEnd
			}
			f.Row(t.writer, line, renderer.Formatting{
				Row: renderer.RowContext{
					Widths:       t.rowWidths,
					ColMaxWidths: t.config.Row.ColMaxWidths,
					Current:      currentCells,
					Previous:     previousCells,
					Next:         nextCells,
					Position:     tw.Row,
					Location:     location,
				},
				Level:     tw.LevelBody,
				HasFooter: len(t.footers) > 0,
				IsSubRow:  j > 0 || t.config.Row.Padding.Global.Top != tw.Empty,
			})

			if cfg.Settings.Separators.BetweenRows.Enabled() && !(i == len(t.rows)-1 && j == len(lines)-1) {
				t.debug("Rendering between-rows separator")
				f.Line(t.writer, renderer.Formatting{
					Row: renderer.RowContext{
						Widths:   t.rowWidths,
						Current:  currentCells,
						Next:     nextCells,
						Position: tw.Row,
						Location: location,
					},
					Level:    tw.LevelBody,
					IsSubRow: false,
				})
			}
		}

		if t.config.Row.Padding.Global.Bottom != tw.Empty {
			t.debug("Rendering row bottom padding for row %d", i)
			bottomPadding := make([]string, numCols)
			for j := range bottomPadding {
				repeatCount := t.rowWidths[j] / twfn.DisplayWidth(t.config.Row.Padding.Global.Bottom)
				if repeatCount < 1 {
					repeatCount = 1
				}
				bottomPadding[j] = strings.Repeat(t.config.Row.Padding.Global.Bottom, repeatCount)
			}
			currentCells := make(map[int]renderer.CellContext)
			for j, cell := range bottomPadding {
				currentCells[j] = renderer.CellContext{
					Data:    cell,
					Align:   colAligns[j],
					Padding: colPadding[j],
					Width:   t.rowWidths[j],
					Merge:   rowVertMerges[i][j],
				}
			}
			nextCells := make(map[int]renderer.CellContext)
			if i+1 < len(rowLines) {
				for j, cell := range rowLines[i+1][0] {
					nextCells[j] = renderer.CellContext{Data: cell, Merge: rowVertMerges[i+1][j]}
				}
			} else if len(t.footers) > 0 {
				for j, cell := range footerLines[0] {
					nextCells[j] = renderer.CellContext{Data: cell, Merge: footerVertMerges[j]}
				}
			}
			location := tw.LocationMiddle
			if i == len(t.rows)-1 && len(t.footers) == 0 {
				location = tw.LocationEnd
			}
			f.Row(t.writer, nil, renderer.Formatting{
				Row: renderer.RowContext{
					Widths:   t.rowWidths,
					Current:  currentCells,
					Next:     nextCells,
					Position: tw.Row,
					Location: location,
				},
				Level:    tw.LevelBody,
				IsSubRow: true,
			})
		}

		if i == len(t.rows)-1 && len(t.footers) == 0 && cfg.Borders.Bottom.Enabled() && cfg.Settings.Lines.ShowBottom.Enabled() {
			t.debug("Rendering table bottom border (no footer)")
			currentCells := make(map[int]renderer.CellContext)
			lastLine := lines[len(lines)-1]
			if t.config.Row.Padding.Global.Bottom != tw.Empty {
				for j := range numCols {
					currentCells[j] = renderer.CellContext{
						Data:    strings.Repeat(t.config.Row.Padding.Global.Bottom, t.rowWidths[j]/twfn.DisplayWidth(t.config.Row.Padding.Global.Bottom)),
						Align:   colAligns[j],
						Padding: colPadding[j],
						Width:   t.rowWidths[j],
						Merge:   rowVertMerges[i][j],
					}
				}
			} else {
				for j := 0; j < numCols; j++ {
					mergeState := rowVertMerges[i][j]
					if rowHorzMerges[i][j] {
						mergeState.Horizontal = true
						mergeState.Span = t.calculateHorizontalSpan(rowHorzMerges[i], j)
						mergeState.Start = j == 0 || !rowHorzMerges[i][j-1]
						mergeState.End = j == numCols-1 || !rowHorzMerges[i][j+1]
					}
					currentCells[j] = renderer.CellContext{
						Data:    lastLine[j],
						Align:   colAligns[j],
						Padding: colPadding[j],
						Width:   t.rowWidths[j],
						Merge:   mergeState,
					}
				}
			}
			f.Line(t.writer, renderer.Formatting{
				Row: renderer.RowContext{
					Widths:   t.rowWidths,
					Current:  currentCells,
					Position: tw.Row,
					Location: tw.LocationEnd,
				},
				Level:    tw.LevelFooter,
				IsSubRow: false,
			})
		}
	}

	// Footer section
	if len(footerLines) > 0 {
		colAligns := t.buildAligns(t.config.Footer)
		colPadding := t.buildPadding(t.config.Footer.Padding)

		if cfg.Settings.Lines.ShowFooterLine.Enabled() && len(t.rows) > 0 {
			t.debug("Rendering footer separator line")
			previousCells := make(map[int]renderer.CellContext)
			lastRow := rowLines[len(t.rows)-1]
			for j, cell := range lastRow[len(lastRow)-1] {
				previousCells[j] = renderer.CellContext{Data: cell, Merge: rowVertMerges[len(t.rows)-1][j]}
			}
			nextCells := make(map[int]renderer.CellContext)
			for j, cell := range footerLines[0] {
				nextCells[j] = renderer.CellContext{Data: cell, Merge: footerVertMerges[j]}
			}
			f.Line(t.writer, renderer.Formatting{
				Row: renderer.RowContext{
					Widths:   t.footerWidths,
					Current:  previousCells,
					Next:     nextCells,
					Position: tw.Footer,
					Location: tw.LocationFirst,
				},
				Level:    tw.LevelBody,
				IsSubRow: false,
			})
		}

		if t.config.Footer.Padding.Global.Top != tw.Empty {
			t.debug("Rendering footer top padding")
			topPadding := make([]string, numCols)
			for i := range topPadding {
				repeatCount := t.footerWidths[i] / twfn.DisplayWidth(t.config.Footer.Padding.Global.Top)
				if repeatCount < 1 {
					repeatCount = 1
				}
				topPadding[i] = strings.Repeat(t.config.Footer.Padding.Global.Top, repeatCount)
			}
			currentCells := make(map[int]renderer.CellContext)
			for j, cell := range topPadding {
				currentCells[j] = renderer.CellContext{
					Data:    cell,
					Align:   colAligns[j],
					Padding: colPadding[j],
					Width:   t.footerWidths[j],
					Merge:   footerVertMerges[j],
				}
			}
			f.Footer(t.writer, nil, renderer.Formatting{
				Row: renderer.RowContext{
					Widths:   t.footerWidths,
					Current:  currentCells,
					Position: tw.Footer,
					Location: tw.LocationFirst,
				},
				Level:    tw.LevelFooter,
				IsSubRow: true,
			})
		}

		for i, line := range footerLines {
			currentCells := make(map[int]renderer.CellContext)
			for j, cell := range line {
				mergeState := footerVertMerges[j]
				if footerHorzMerges[j] {
					mergeState.Horizontal = true
					mergeState.Span = t.calculateHorizontalSpan(footerHorzMerges, j)
					mergeState.Start = j == 0 || !footerHorzMerges[j-1]
					mergeState.End = j == numCols-1 || !footerHorzMerges[j+1]
				}
				currentCells[j] = renderer.CellContext{
					Data:    cell,
					Align:   colAligns[j],
					Padding: colPadding[j],
					Width:   t.footerWidths[j],
					Merge:   mergeState,
				}
			}
			previousCells := make(map[int]renderer.CellContext)
			if i > 0 {
				for j, cell := range footerLines[i-1] {
					previousCells[j] = renderer.CellContext{Data: cell, Merge: footerVertMerges[j]}
				}
			} else if len(t.rows) > 0 {
				for j, cell := range rowLines[len(t.rows)-1][len(rowLines[len(t.rows)-1])-1] {
					previousCells[j] = renderer.CellContext{Data: cell, Merge: rowVertMerges[len(t.rows)-1][j]}
				}
			}
			location := tw.LocationFirst
			if i > 0 || t.config.Footer.Padding.Global.Top != tw.Empty {
				location = tw.LocationMiddle
			}
			if i == len(footerLines)-1 && t.config.Footer.Padding.Global.Bottom == tw.Empty {
				location = tw.LocationEnd
			}
			f.Footer(t.writer, footerLines, renderer.Formatting{
				Row: renderer.RowContext{
					Widths:       t.footerWidths,
					ColMaxWidths: t.config.Footer.ColMaxWidths,
					Current:      currentCells,
					Previous:     previousCells,
					Position:     tw.Footer,
					Location:     location,
				},
				Level:     tw.LevelFooter,
				HasFooter: true,
				IsSubRow:  i > 0 || t.config.Footer.Padding.Global.Top != tw.Empty,
			})
		}

		if t.config.Footer.Padding.Global.Bottom != tw.Empty {
			t.debug("Rendering footer bottom padding")
			bottomPadding := make([]string, numCols)
			for i := range bottomPadding {
				repeatCount := t.footerWidths[i] / twfn.DisplayWidth(t.config.Footer.Padding.Global.Bottom)
				if repeatCount < 1 {
					repeatCount = 1
				}
				bottomPadding[i] = strings.Repeat(t.config.Footer.Padding.Global.Bottom, repeatCount)
			}
			currentCells := make(map[int]renderer.CellContext)
			for j, cell := range bottomPadding {
				currentCells[j] = renderer.CellContext{
					Data:    cell,
					Align:   colAligns[j],
					Padding: colPadding[j],
					Width:   t.footerWidths[j],
					Merge:   footerVertMerges[j],
				}
			}
			f.Footer(t.writer, nil, renderer.Formatting{
				Row: renderer.RowContext{
					Widths:   t.footerWidths,
					Current:  currentCells,
					Position: tw.Footer,
					Location: tw.LocationEnd,
				},
				Level:    tw.LevelFooter,
				IsSubRow: true,
			})
		}

		if cfg.Borders.Bottom.Enabled() && cfg.Settings.Lines.ShowBottom.Enabled() {
			t.debug("Rendering table bottom border (with footer)")
			currentCells := make(map[int]renderer.CellContext)
			for j := 0; j < numCols; j++ {
				mergeState := footerVertMerges[j]
				if footerHorzMerges[j] {
					mergeState.Horizontal = true
					mergeState.Span = t.calculateHorizontalSpan(footerHorzMerges, j)
					mergeState.Start = j == 0 || !footerHorzMerges[j-1]
					mergeState.End = j == numCols-1 || !footerHorzMerges[j+1]
				}
				currentCells[j] = renderer.CellContext{
					Data:    footerLines[len(footerLines)-1][j],
					Align:   colAligns[j],
					Padding: colPadding[j],
					Width:   t.footerWidths[j],
					Merge:   mergeState,
				}
			}
			f.Line(t.writer, renderer.Formatting{
				Row: renderer.RowContext{
					Widths:   t.footerWidths,
					Current:  currentCells,
					Position: tw.Footer,
					Location: tw.LocationEnd,
				},
				Level:    tw.LevelFooter,
				IsSubRow: false,
			})
		}
	}

	if len(headerLines) == 0 && len(footerLines) == 0 && len(rowLines) > 0 {
		if cfg.Borders.Top.Enabled() && cfg.Settings.Lines.ShowTop.Enabled() {
			t.debug("Rendering table top border (rows only)")
			f.Line(t.writer, renderer.Formatting{
				Row: renderer.RowContext{
					Widths:   t.rowWidths,
					Position: tw.Row,
					Location: tw.LocationFirst,
				},
				Level:    tw.LevelHeader,
				IsSubRow: false,
			})
		}
	}

	t.hasPrinted = true
	t.debug("Render completed")
	return nil
}

// ensureInitialized ensures all required fields are initialized
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
	t.debug("ensureInitialized called")
}

// toStringLines converts a row to string lines using stringer or direct cast
func (t *Table) toStringLines(row interface{}, config CellConfig) ([][]string, error) {
	t.debug("Converting row to string lines: %v", row)
	var cells []string
	switch v := row.(type) {
	case []string:
		cells = v
		t.debug("Row is already []string")
	default:
		if t.stringer == nil {
			err := errors.Newf("no stringer provided for type %T", row)
			t.debug("Stringer error: %v", err)
			return nil, err
		}
		rv := reflect.ValueOf(t.stringer)
		if rv.Kind() != reflect.Func || rv.Type().NumIn() != 1 || rv.Type().NumOut() != 1 {
			err := errors.Newf("stringer must be a func(T) []string")
			t.debug("Stringer format error: %v", err)
			return nil, err
		}
		in := []reflect.Value{reflect.ValueOf(row)}
		out := rv.Call(in)
		if len(out) != 1 || out[0].Kind() != reflect.Slice || out[0].Type().Elem().Kind() != reflect.String {
			err := errors.Newf("stringer must return []string")
			t.debug("Stringer return error: %v", err)
			return nil, err
		}
		cells = out[0].Interface().([]string)
		t.debug("Converted row using stringer: %v", cells)
	}

	if config.Filter != nil {
		t.debug("Applying filter to cells")
		cells = config.Filter(cells)
	}

	result := t.prepareContent(cells, config)
	t.debug("Prepared content: %v", result)
	return result, nil
}

// prepareContent processes cell content with formatting and wrapping
func (t *Table) prepareContent(cells []string, config CellConfig) [][]string {
	t.debug("Preparing content: cells=%v", cells)
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

		padLeftWidth := twfn.DisplayWidth(config.Padding.Global.Left)
		padRightWidth := twfn.DisplayWidth(config.Padding.Global.Right)
		if i < len(config.Padding.PerColumn) && config.Padding.PerColumn[i] != (tw.Padding{}) {
			padLeftWidth = twfn.DisplayWidth(config.Padding.PerColumn[i].Left)
			padRightWidth = twfn.DisplayWidth(config.Padding.PerColumn[i].Right)
		}

		contentWidth := effectiveMaxWidth - padLeftWidth - padRightWidth
		if contentWidth < 1 {
			contentWidth = 1
		}

		if config.Formatting.AutoFormat {
			cell = twfn.Title(strings.Join(twfn.SplitCamelCase(cell), tw.Space))
		}

		lines := strings.Split(cell, "\n")
		finalLines := make([]string, 0)

		for _, line := range lines {
			if effectiveMaxWidth > 0 {
				switch config.Formatting.AutoWrap {
				case tw.WrapNormal:
					wrapped, _ := twfn.WrapString(line, contentWidth)
					finalLines = append(finalLines, wrapped...)
				case tw.WrapTruncate:
					if twfn.DisplayWidth(line) > contentWidth {
						finalLines = append(finalLines, twfn.TruncateString(line, contentWidth-1)+tw.CharEllipsis)
					} else {
						finalLines = append(finalLines, line)
					}
				case tw.WrapBreak:
					wrapped := make([]string, 0)
					for len(line) > contentWidth {
						wrapped = append(wrapped, line[:contentWidth]+tw.CharBreak)
						line = line[contentWidth:]
					}
					if len(line) > 0 {
						wrapped = append(wrapped, line)
					}
					finalLines = append(finalLines, wrapped...)
				default:
					finalLines = append(finalLines, line)
				}
			} else {
				finalLines = append(finalLines, line)
			}
		}

		for len(result) < len(finalLines) {
			newRow := make([]string, numCols)
			for j := range newRow {
				newRow[j] = tw.Empty
			}
			result = append(result, newRow)
		}

		for j, line := range finalLines {
			result[j][i] = line
		}
	}

	t.debug("Content prepared: %v", result)
	return result
}

// prepareWithMerges handles content preparation including merges
// prepareWithMerges handles content preparation including horizontal merges only
func (t *Table) prepareWithMerges(content [][]string, config CellConfig, position tw.Position) ([][]string, map[int]renderer.MergeState, map[int]bool) {
	t.debug("PrepareWithMerges START: position=%s, mergeMode=%d", position, config.Formatting.MergeMode)
	if len(content) == 0 {
		t.debug("PrepareWithMerges END: No content.")
		return content, nil, nil
	}

	numCols := len(content[0])
	horzMergeMap := make(map[int]bool)
	vertMergeMap := make(map[int]renderer.MergeState)
	result := make([][]string, len(content))
	for i := range content {
		result[i] = make([]string, len(content[i]))
		copy(result[i], content[i])
	}

	// Horizontal merges
	if config.Formatting.MergeMode == tw.MergeHorizontal || config.Formatting.MergeMode == tw.MergeBoth {
		t.debug("Checking for horizontal merges in %d rows", len(content))
		for row := 0; row < len(content); row++ {
			currentRowLen := len(result[row])
			if currentRowLen < numCols {
				for k := currentRowLen; k < numCols; k++ {
					result[row] = append(result[row], tw.Empty)
				}
			}
			col := 0
			for col < numCols-1 {
				currentVal := strings.TrimSpace(result[row][col])
				if currentVal == "" {
					col++
					continue
				}
				span := 1
				startCol := col
				for nextCol := col + 1; nextCol < numCols; nextCol++ {
					nextVal := strings.TrimSpace(result[row][nextCol])
					if currentVal == nextVal && currentVal != "" {
						span++
						result[row][nextCol] = tw.Empty
						horzMergeMap[nextCol] = true
						t.debug("Horizontal merge detected at row %d, col %d -> col %d", row, startCol, nextCol)
					} else {
						break
					}
				}
				if span > 1 {
					horzMergeMap[startCol] = true
					vertMergeMap[startCol] = renderer.MergeState{
						Horizontal: true,
						Span:       span,
						Start:      true,
						End:        false,
					}
					vertMergeMap[startCol+span-1] = renderer.MergeState{
						Horizontal: true,
						Span:       span,
						Start:      false,
						End:        true,
					}
					for k := startCol + 1; k < startCol+span-1; k++ {
						vertMergeMap[k] = renderer.MergeState{
							Horizontal: true,
							Span:       span,
							Start:      false,
							End:        false,
						}
					}
				}
				col += span
			}
		}
	}

	t.debug("PrepareWithMerges END: position=%s, lines=%d", position, len(result))
	return result, vertMergeMap, horzMergeMap
}

// calculateHorizontalSpan calculates the span of a horizontal merge
func (t *Table) calculateHorizontalSpan(horzMerges map[int]bool, startCol int) int {
	span := 1
	for col := startCol + 1; horzMerges[col]; col++ {
		span++
	}
	return span
}

// updateWidths calculates and updates column widths
func (t *Table) updateWidths(row []string, widths map[int]int, padding CellPadding) {
	t.debug("Updating widths for row: %v", row)
	for i, cell := range row {
		padLeftWidth := twfn.DisplayWidth(padding.Global.Left)
		padRightWidth := twfn.DisplayWidth(padding.Global.Right)
		if i < len(padding.PerColumn) && padding.PerColumn[i] != (tw.Padding{}) {
			padLeftWidth = twfn.DisplayWidth(padding.PerColumn[i].Left)
			padRightWidth = twfn.DisplayWidth(padding.PerColumn[i].Right)
		}
		totalWidth := twfn.DisplayWidth(strings.TrimSpace(cell)) + padLeftWidth + padRightWidth
		if current, exists := widths[i]; !exists || totalWidth > current {
			widths[i] = totalWidth
		}
	}
}

// maxColumns determines the maximum number of columns in the table
func (t *Table) maxColumns() int {
	m := 0
	if len(t.headers) > 0 && len(t.headers[0]) > m {
		m = len(t.headers[0])
	}
	for _, row := range t.rows {
		if len(row) > 0 && len(row[0]) > m {
			m = len(row[0])
		}
	}
	if len(t.footers) > 0 && len(t.footers[0]) > m {
		m = len(t.footers[0])
	}
	t.debug("Max columns calculated: %d", m)
	return m
}

// buildAligns constructs alignment settings for columns
func (t *Table) buildAligns(config CellConfig) map[int]tw.Align {
	t.debug("Building aligns")
	colAligns := make(map[int]tw.Align)
	numCols := t.maxColumns()
	for i := 0; i < numCols; i++ {
		if i < len(config.ColumnAligns) && config.ColumnAligns[i] != tw.Empty {
			colAligns[i] = config.ColumnAligns[i]
		} else {
			colAligns[i] = config.Formatting.Alignment
		}
	}
	t.debug("Aligns built: %v", colAligns)
	return colAligns
}

// buildPadding constructs padding settings for columns
func (t *Table) buildPadding(padding CellPadding) map[int]tw.Padding {
	t.debug("Building padding")
	colPadding := make(map[int]tw.Padding)
	numCols := t.maxColumns()
	for i := 0; i < numCols; i++ {
		if i < len(padding.PerColumn) && padding.PerColumn[i] != (tw.Padding{}) {
			colPadding[i] = padding.PerColumn[i]
		} else {
			colPadding[i] = padding.Global
		}
	}
	t.debug("Padding built: %v", colPadding)
	return colPadding
}

// debug logs a message to the trace if debug mode is enabled
func (t *Table) debug(format string, a ...interface{}) {
	if t.config.Debug {
		msg := fmt.Sprintf(format, a...)
		t.trace = append(t.trace, fmt.Sprintf("[TABLE] %s", msg))
	}
}
