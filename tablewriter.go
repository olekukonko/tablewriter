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

// Table represents a table with content and rendering capabilities
type Table struct {
	writer       io.Writer           // Output destination
	rows         [][][]string        // Row data (multi-line cells supported)
	headers      [][]string          // Header content
	footers      [][]string          // Footer content
	headerWidths tw.Mapper[int, int] // Calculated widths for header columns
	rowWidths    tw.Mapper[int, int] // Calculated widths for row columns
	footerWidths tw.Mapper[int, int] // Calculated widths for footer columns
	renderer     renderer.Renderer   // Rendering engine
	config       Config              // Table configuration
	stringer     any                 // Function to convert rows to strings
	newLine      string              // Newline character
	hasPrinted   bool                // Tracks if table has been rendered
	trace        []string            // Debug trace log
}

// renderContext holds core rendering state
type renderContext struct {
	table       *Table
	renderer    renderer.Renderer
	cfg         renderer.DefaultConfig
	numCols     int
	headerLines [][]string
	rowLines    [][][]string
	footerLines [][]string
	widths      map[tw.Position]tw.Mapper[int, int]
	debug       func(format string, a ...interface{})
}

// mergeContext holds merge-related state
type mergeContext struct {
	headerMerges map[int]renderer.MergeState
	rowMerges    []map[int]renderer.MergeState
	footerMerges map[int]renderer.MergeState
	horzMerges   map[tw.Position]map[int]bool
}

// helperContext holds additional helper data
type helperContext struct {
	position tw.Position
	rowIdx   int // Tracks row index
	lineIdx  int
	location tw.Location
	line     []string
}

// renderMergeResponse holds response data from rendering operations
type renderMergeResponse struct {
	cells     map[int]renderer.CellContext
	prevCells map[int]renderer.CellContext
	nextCells map[int]renderer.CellContext
}

// ---- Public Methods ----

// NewTable creates a new table instance with optional configurations
func NewTable(w io.Writer, opts ...Option) *Table {
	t := &Table{
		writer:       w,
		headerWidths: tw.NewMapper[int, int](),
		rowWidths:    tw.NewMapper[int, int](),
		footerWidths: tw.NewMapper[int, int](),
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

// Append adds one or more rows to the table (variadic)
func (t *Table) Append(rows ...interface{}) error {
	for i, row := range rows {
		if err := t.appendSingle(row); err != nil {
			t.debug("Append failed at index %d: %v", i, err)
			return err
		}
	}
	t.debug("Appended %d rows, total rows: %d", len(rows), len(t.rows))
	return nil
}

// Bulk adds multiple rows to the table (legacy, kept for compatibility)
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
		if err := t.appendSingle(row); err != nil {
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

// Render invokes the rendering process
func (t *Table) Render() error {
	return t.render()
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

// ---- Renderers ----

// render orchestrates the table rendering process
func (t *Table) render() error {
	t.ensureInitialized()
	ctx, mctx, err := t.prepareContexts()
	if err != nil {
		return err
	}

	for _, renderFn := range []func(*renderContext, *mergeContext) error{
		t.renderHeader,
		t.renderRow,
		t.renderFooter,
	} {
		if err := renderFn(ctx, mctx); err != nil {
			return err
		}
	}

	t.hasPrinted = true
	ctx.debug("Render completed")
	return nil
}

// renderHeader renders the header section
func (t *Table) renderHeader(ctx *renderContext, mctx *mergeContext) error {
	if len(ctx.headerLines) == 0 {
		return nil
	}
	ctx.debug("Rendering header section")

	f := ctx.renderer
	cfg := ctx.cfg
	colAligns := t.buildAligns(t.config.Header)
	colPadding := t.buildPadding(t.config.Header.Padding)
	hctx := &helperContext{position: tw.Header}

	// Top border
	if cfg.Borders.Top.Enabled() && cfg.Settings.Lines.ShowTop.Enabled() {
		ctx.debug("Rendering table top border")
		nextCells := make(map[int]renderer.CellContext)
		if len(ctx.headerLines) > 0 {
			for j, cell := range ctx.headerLines[0] {
				nextCells[j] = renderer.CellContext{Data: cell, Merge: mctx.headerMerges[j]}
			}
		}
		f.Line(t.writer, renderer.Formatting{
			Row: renderer.RowContext{
				Widths:   ctx.widths[tw.Header],
				Next:     nextCells,
				Position: tw.Header,
				Location: tw.LocationFirst,
			},
			Level:    tw.LevelHeader,
			IsSubRow: false,
		})
	}

	// Top padding
	if t.config.Header.Padding.Global.Top != tw.Empty {
		hctx.location = tw.LocationFirst
		hctx.line = make([]string, ctx.numCols)
		for j := 0; j < ctx.numCols; j++ {
			repeatCount := ctx.widths[tw.Header].Get(j) / twfn.DisplayWidth(t.config.Header.Padding.Global.Top)
			if repeatCount < 1 {
				repeatCount = 1
			}
			hctx.line[j] = strings.Repeat(t.config.Header.Padding.Global.Top, repeatCount)
		}
		if err := t.renderPadding(ctx, mctx, hctx, t.config.Header.Padding.Global.Top); err != nil {
			return err
		}
	}

	// Header content
	for i, line := range ctx.headerLines {
		hctx.rowIdx = 0 // Single-row section
		hctx.lineIdx = i
		hctx.line = padLine(line, ctx.numCols)
		hctx.location = t.determineLocation(i, len(ctx.headerLines), t.config.Header.Padding.Global.Top, t.config.Header.Padding.Global.Bottom)
		if err := t.renderLine(ctx, mctx, hctx, colAligns, colPadding); err != nil {
			return err
		}
	}

	// Bottom padding
	if t.config.Header.Padding.Global.Bottom != tw.Empty {
		hctx.location = tw.LocationEnd
		hctx.line = make([]string, ctx.numCols)
		for j := 0; j < ctx.numCols; j++ {
			repeatCount := ctx.widths[tw.Header].Get(j) / twfn.DisplayWidth(t.config.Header.Padding.Global.Bottom)
			if repeatCount < 1 {
				repeatCount = 1
			}
			hctx.line[j] = strings.Repeat(t.config.Header.Padding.Global.Bottom, repeatCount)
		}
		if err := t.renderPadding(ctx, mctx, hctx, t.config.Header.Padding.Global.Bottom); err != nil {
			return err
		}
	}

	// Header separator
	if cfg.Settings.Lines.ShowHeaderLine.Enabled() && (len(ctx.rowLines) > 0 || len(ctx.footerLines) > 0) {
		ctx.debug("Rendering header separator line")
		hctx.rowIdx = 0
		hctx.lineIdx = len(ctx.headerLines) - 1
		hctx.line = padLine(ctx.headerLines[hctx.lineIdx], ctx.numCols)
		hctx.location = tw.LocationMiddle
		resp := t.buildCellContexts(ctx, mctx, hctx, colAligns, colPadding)
		f.Line(t.writer, renderer.Formatting{
			Row: renderer.RowContext{
				Widths:   ctx.widths[tw.Header],
				Current:  resp.cells, // Last header line
				Previous: resp.prevCells,
				Next:     resp.nextCells, // First row or footer
				Position: tw.Header,
				Location: tw.LocationMiddle,
			},
			Level:    tw.LevelBody,
			IsSubRow: false,
		})
	}
	return nil
}

// renderRow renders the row section
func (t *Table) renderRow(ctx *renderContext, mctx *mergeContext) error {
	if len(ctx.rowLines) == 0 {
		return nil
	}
	ctx.debug("Rendering row section")

	f := ctx.renderer
	cfg := ctx.cfg
	colAligns := t.buildAligns(t.config.Row)
	colPadding := t.buildPadding(t.config.Row.Padding)
	hctx := &helperContext{position: tw.Row}

	// Rows-only top border
	if len(ctx.headerLines) == 0 && len(ctx.footerLines) == 0 && cfg.Borders.Top.Enabled() && cfg.Settings.Lines.ShowTop.Enabled() {
		ctx.debug("Rendering table top border (rows only)")
		nextCells := make(map[int]renderer.CellContext)
		if len(ctx.rowLines) > 0 {
			for j, cell := range ctx.rowLines[0][0] {
				nextCells[j] = renderer.CellContext{Data: cell, Merge: mctx.rowMerges[0][j]}
			}
		}
		f.Line(t.writer, renderer.Formatting{
			Row: renderer.RowContext{
				Widths:   ctx.widths[tw.Row],
				Next:     nextCells,
				Position: tw.Row,
				Location: tw.LocationFirst,
			},
			Level:    tw.LevelHeader,
			IsSubRow: false,
		})
	}

	// Row content
	for i, lines := range ctx.rowLines {
		if t.config.Row.Padding.Global.Top != tw.Empty {
			hctx.rowIdx = i
			hctx.lineIdx = -1 // Special index for padding
			hctx.location = tw.LocationMiddle
			if i == 0 && len(ctx.headerLines) == 0 {
				hctx.location = tw.LocationFirst
			}
			hctx.line = make([]string, ctx.numCols)
			for j := 0; j < ctx.numCols; j++ {
				repeatCount := ctx.widths[tw.Row].Get(j) / twfn.DisplayWidth(t.config.Row.Padding.Global.Top)
				if repeatCount < 1 {
					repeatCount = 1
				}
				hctx.line[j] = strings.Repeat(t.config.Row.Padding.Global.Top, repeatCount)
			}
			if err := t.renderPadding(ctx, mctx, hctx, t.config.Row.Padding.Global.Top); err != nil {
				return err
			}
		}

		for j, line := range lines {
			hctx.rowIdx = i
			hctx.lineIdx = j
			hctx.line = padLine(line, ctx.numCols)
			hctx.location = tw.LocationMiddle
			if i == 0 && j == 0 && len(ctx.headerLines) == 0 && t.config.Row.Padding.Global.Top == tw.Empty {
				hctx.location = tw.LocationFirst
			}
			if i == len(ctx.rowLines)-1 && j == len(lines)-1 && len(ctx.footerLines) == 0 && t.config.Row.Padding.Global.Bottom == tw.Empty {
				hctx.location = tw.LocationEnd
			}
			if err := t.renderLine(ctx, mctx, hctx, colAligns, colPadding); err != nil {
				return err
			}

			// Between rows
			if cfg.Settings.Separators.BetweenRows.Enabled() && !(i == len(ctx.rowLines)-1 && j == len(lines)-1) {
				ctx.debug("Rendering between-rows separator")
				resp := t.buildCellContexts(ctx, mctx, hctx, colAligns, colPadding)
				f.Line(t.writer, renderer.Formatting{
					Row: renderer.RowContext{
						Widths:   ctx.widths[tw.Row],
						Current:  resp.cells,
						Previous: resp.prevCells,
						Next:     resp.nextCells,
						Position: tw.Row,
						Location: hctx.location,
					},
					Level:    tw.LevelBody,
					IsSubRow: false,
				})
			}
		}

		if t.config.Row.Padding.Global.Bottom != tw.Empty {
			hctx.rowIdx = i
			hctx.lineIdx = len(lines) // Special index for padding
			hctx.location = tw.LocationMiddle
			if i == len(ctx.rowLines)-1 && len(ctx.footerLines) == 0 {
				hctx.location = tw.LocationEnd
			}
			hctx.line = make([]string, ctx.numCols)
			for j := 0; j < ctx.numCols; j++ {
				repeatCount := ctx.widths[tw.Row].Get(j) / twfn.DisplayWidth(t.config.Row.Padding.Global.Bottom)
				if repeatCount < 1 {
					repeatCount = 1
				}
				hctx.line[j] = strings.Repeat(t.config.Row.Padding.Global.Bottom, repeatCount)
			}
			if err := t.renderPadding(ctx, mctx, hctx, t.config.Row.Padding.Global.Bottom); err != nil {
				return err
			}
		}

		// Bottom border (no footer)
		if i == len(ctx.rowLines)-1 && len(ctx.footerLines) == 0 && cfg.Borders.Bottom.Enabled() && cfg.Settings.Lines.ShowBottom.Enabled() {
			ctx.debug("Rendering table bottom border (no footer)")
			hctx.rowIdx = i
			hctx.lineIdx = len(lines) - 1
			hctx.line = padLine(lines[hctx.lineIdx], ctx.numCols)
			hctx.location = tw.LocationEnd
			resp := t.buildCellContexts(ctx, mctx, hctx, colAligns, colPadding)
			f.Line(t.writer, renderer.Formatting{
				Row: renderer.RowContext{
					Widths:   ctx.widths[tw.Row],
					Current:  resp.cells,
					Previous: resp.prevCells,
					Position: tw.Row,
					Location: tw.LocationEnd,
				},
				Level:    tw.LevelFooter,
				IsSubRow: false,
			})
		}
	}

	return nil
}

// renderFooter renders the footer section
func (t *Table) renderFooter(ctx *renderContext, mctx *mergeContext) error {
	if len(ctx.footerLines) == 0 {
		return nil
	}
	ctx.debug("Rendering footer section")

	f := ctx.renderer
	cfg := ctx.cfg
	colAligns := t.buildAligns(t.config.Footer)
	colPadding := t.buildPadding(t.config.Footer.Padding)
	hctx := &helperContext{position: tw.Footer}

	// Footer separator
	if cfg.Settings.Lines.ShowFooterLine.Enabled() && len(ctx.rowLines) > 0 {
		ctx.debug("Rendering footer separator line")
		prevCells := make(map[int]renderer.CellContext)
		if len(ctx.rowLines) > 0 {
			lastRow := ctx.rowLines[len(ctx.rowLines)-1]
			for j, cell := range lastRow[len(lastRow)-1] {
				prevCells[j] = renderer.CellContext{Data: cell, Merge: mctx.rowMerges[len(ctx.rowLines)-1][j]}
			}
		}
		hctx.rowIdx = 0
		hctx.lineIdx = 0
		hctx.line = padLine(ctx.footerLines[0], ctx.numCols)
		hctx.location = tw.LocationFirst
		resp := t.buildCellContexts(ctx, mctx, hctx, colAligns, colPadding)
		f.Line(t.writer, renderer.Formatting{
			Row: renderer.RowContext{
				Widths:   ctx.widths[tw.Footer],
				Current:  prevCells,
				Next:     resp.cells,
				Position: tw.Footer,
				Location: tw.LocationFirst,
			},
			Level:    tw.LevelBody,
			IsSubRow: false,
		})
	}

	// Footer content
	if t.config.Footer.Padding.Global.Top != tw.Empty {
		hctx.rowIdx = 0
		hctx.lineIdx = -1 // Special index for padding
		hctx.location = tw.LocationFirst
		hctx.line = make([]string, ctx.numCols)
		for j := 0; j < ctx.numCols; j++ {
			repeatCount := ctx.widths[tw.Footer].Get(j) / twfn.DisplayWidth(t.config.Footer.Padding.Global.Top)
			if repeatCount < 1 {
				repeatCount = 1
			}
			hctx.line[j] = strings.Repeat(t.config.Footer.Padding.Global.Top, repeatCount)
		}
		if err := t.renderPadding(ctx, mctx, hctx, t.config.Footer.Padding.Global.Top); err != nil {
			return err
		}
	}

	for i, line := range ctx.footerLines {
		hctx.rowIdx = 0
		hctx.lineIdx = i
		hctx.line = padLine(line, ctx.numCols)
		hctx.location = t.determineLocation(i, len(ctx.footerLines), t.config.Footer.Padding.Global.Top, t.config.Footer.Padding.Global.Bottom)
		if err := t.renderLine(ctx, mctx, hctx, colAligns, colPadding); err != nil {
			return err
		}
	}

	if t.config.Footer.Padding.Global.Bottom != tw.Empty {
		hctx.rowIdx = 0
		hctx.lineIdx = len(ctx.footerLines) // Special index for padding
		hctx.location = tw.LocationEnd
		hctx.line = make([]string, ctx.numCols)
		for j := 0; j < ctx.numCols; j++ {
			repeatCount := ctx.widths[tw.Footer].Get(j) / twfn.DisplayWidth(t.config.Footer.Padding.Global.Bottom)
			if repeatCount < 1 {
				repeatCount = 1
			}
			hctx.line[j] = strings.Repeat(t.config.Footer.Padding.Global.Bottom, repeatCount)
		}
		if err := t.renderPadding(ctx, mctx, hctx, t.config.Footer.Padding.Global.Bottom); err != nil {
			return err
		}
	}

	// Bottom border
	if cfg.Borders.Bottom.Enabled() && cfg.Settings.Lines.ShowBottom.Enabled() {
		ctx.debug("Rendering table bottom border (with footer)")
		hctx.rowIdx = 0
		hctx.lineIdx = len(ctx.footerLines) - 1
		hctx.line = padLine(ctx.footerLines[hctx.lineIdx], ctx.numCols)
		hctx.location = tw.LocationEnd
		resp := t.buildCellContexts(ctx, mctx, hctx, colAligns, colPadding)
		f.Line(t.writer, renderer.Formatting{
			Row: renderer.RowContext{
				Widths:   ctx.widths[tw.Footer],
				Current:  resp.cells,
				Previous: resp.prevCells,
				Position: tw.Footer,
				Location: tw.LocationEnd,
			},
			Level:    tw.LevelFooter,
			IsSubRow: false,
		})
	}

	return nil
}

// ---- Builders ----

// buildCellContexts constructs CellContext for a given line
func (t *Table) buildCellContexts(ctx *renderContext, mctx *mergeContext, hctx *helperContext, aligns map[int]tw.Align, padding map[int]tw.Padding) renderMergeResponse {
	cells := make(map[int]renderer.CellContext)
	var merges map[int]renderer.MergeState
	switch hctx.position {
	case tw.Header:
		merges = mctx.headerMerges
	case tw.Row:
		merges = mctx.rowMerges[hctx.rowIdx]
	case tw.Footer:
		merges = mctx.footerMerges
	}

	// Handle horizontal merges by adjusting cell content and widths
	for j := 0; j < ctx.numCols; j++ {
		mergeState := merges[j]
		if mctx.horzMerges[hctx.position][j] && mergeState.Horizontal {
			if mergeState.Start {
				// Start of a merged cell: combine content and width
				mergedWidth := ctx.widths[hctx.position].Get(j)
				mergedContent := hctx.line[j]
				for k := j + 1; k < j+mergeState.Span && k < ctx.numCols; k++ {
					mergedWidth += ctx.widths[hctx.position].Get(k)
					if hctx.line[k] != tw.Empty {
						mergedContent = hctx.line[k] // Use the first non-empty content
					}
				}
				cells[j] = renderer.CellContext{
					Data:    mergedContent,
					Align:   aligns[j],
					Padding: padding[j],
					Width:   mergedWidth,
					Merge:   mergeState,
				}
			}
			// Skip subsequent merged cells (they're covered by the start cell)
		} else {
			cells[j] = renderer.CellContext{
				Data:    hctx.line[j],
				Align:   aligns[j],
				Padding: padding[j],
				Width:   ctx.widths[hctx.position].Get(j),
				Merge:   mergeState,
			}
		}
	}

	return renderMergeResponse{
		cells:     cells,
		prevCells: t.buildPreviousCells(ctx, mctx, hctx),
		nextCells: t.buildNextCells(ctx, mctx, hctx),
	}
}

// buildPreviousCells constructs previous cells for any section
func (t *Table) buildPreviousCells(ctx *renderContext, mctx *mergeContext, hctx *helperContext) map[int]renderer.CellContext {
	prevCells := make(map[int]renderer.CellContext)

	switch hctx.position {
	case tw.Header:
		if hctx.lineIdx > 0 {
			for j, cell := range ctx.headerLines[hctx.lineIdx-1] {
				prevCells[j] = renderer.CellContext{Data: cell, Merge: mctx.headerMerges[j]}
			}
			return prevCells
		}
	case tw.Row:
		lines := ctx.rowLines[hctx.rowIdx]
		if hctx.lineIdx >= 0 { // Normal row content
			if hctx.lineIdx > 0 {
				for j, cell := range lines[hctx.lineIdx-1] {
					prevCells[j] = renderer.CellContext{Data: cell, Merge: mctx.rowMerges[hctx.rowIdx][j]}
				}
				return prevCells
			} else if hctx.rowIdx == 0 && len(ctx.headerLines) > 0 {
				for j, cell := range ctx.headerLines[len(ctx.headerLines)-1] {
					prevCells[j] = renderer.CellContext{Data: cell, Merge: mctx.headerMerges[j]}
				}
				return prevCells
			} else if hctx.rowIdx > 0 {
				prevLines := ctx.rowLines[hctx.rowIdx-1]
				for j, cell := range prevLines[len(prevLines)-1] {
					prevCells[j] = renderer.CellContext{Data: cell, Merge: mctx.rowMerges[hctx.rowIdx-1][j]}
				}
				return prevCells
			}
		}
		// Padding case (hctx.lineIdx < 0 or > len(lines)) returns nil
	case tw.Footer:
		if hctx.lineIdx > 0 {
			for j, cell := range ctx.footerLines[hctx.lineIdx-1] {
				prevCells[j] = renderer.CellContext{Data: cell, Merge: mctx.footerMerges[j]}
			}
			return prevCells
		} else if hctx.rowIdx == 0 && len(ctx.rowLines) > 0 {
			lastRow := ctx.rowLines[len(ctx.rowLines)-1]
			for j, cell := range lastRow[len(lastRow)-1] {
				prevCells[j] = renderer.CellContext{Data: cell, Merge: mctx.rowMerges[len(ctx.rowLines)-1][j]}
			}
			return prevCells
		}
	}
	return nil
}

// buildNextCells constructs next cells for any section
func (t *Table) buildNextCells(ctx *renderContext, mctx *mergeContext, hctx *helperContext) map[int]renderer.CellContext {
	nextCells := make(map[int]renderer.CellContext)

	switch hctx.position {
	case tw.Header:
		if hctx.lineIdx+1 < len(ctx.headerLines) {
			for j, cell := range ctx.headerLines[hctx.lineIdx+1] {
				nextCells[j] = renderer.CellContext{Data: cell, Merge: mctx.headerMerges[j]}
			}
			return nextCells
		} else if len(ctx.rowLines) > 0 {
			for j, cell := range ctx.rowLines[0][0] {
				nextCells[j] = renderer.CellContext{Data: cell, Merge: mctx.rowMerges[0][j]}
			}
			return nextCells
		}
	case tw.Row:
		lines := ctx.rowLines[hctx.rowIdx]
		if hctx.lineIdx >= 0 { // Normal row content
			if hctx.lineIdx+1 < len(lines) {
				for j, cell := range lines[hctx.lineIdx+1] {
					nextCells[j] = renderer.CellContext{Data: cell, Merge: mctx.rowMerges[hctx.rowIdx][j]}
				}
				return nextCells
			} else if hctx.rowIdx+1 < len(ctx.rowLines) {
				for j, cell := range ctx.rowLines[hctx.rowIdx+1][0] {
					nextCells[j] = renderer.CellContext{Data: cell, Merge: mctx.rowMerges[hctx.rowIdx+1][j]}
				}
				return nextCells
			} else if len(ctx.footerLines) > 0 {
				for j, cell := range ctx.footerLines[0] {
					nextCells[j] = renderer.CellContext{Data: cell, Merge: mctx.footerMerges[j]}
				}
				return nextCells
			}
		}
		// Padding case (hctx.lineIdx < 0 or > len(lines)) returns nil
	case tw.Footer:
		if hctx.lineIdx+1 < len(ctx.footerLines) {
			for j, cell := range ctx.footerLines[hctx.lineIdx+1] {
				nextCells[j] = renderer.CellContext{Data: cell, Merge: mctx.footerMerges[j]}
			}
			return nextCells
		}
	}
	return nil
}

// ---- Helpers ----

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

// prepareContexts initializes the rendering and merge contexts
func (t *Table) prepareContexts() (*renderContext, *mergeContext, error) {
	ctx := &renderContext{
		table:       t,
		renderer:    t.renderer,
		cfg:         t.renderer.Config(),
		numCols:     t.maxColumns(),
		headerLines: t.headers,
		rowLines:    t.rows,
		footerLines: t.footers,
		widths: map[tw.Position]tw.Mapper[int, int]{
			tw.Header: tw.NewMapper[int, int](),
			tw.Row:    tw.NewMapper[int, int](),
			tw.Footer: tw.NewMapper[int, int](),
		},
		debug: t.debug,
	}

	mctx := &mergeContext{
		headerMerges: make(map[int]renderer.MergeState),
		rowMerges:    make([]map[int]renderer.MergeState, len(t.rows)),
		footerMerges: make(map[int]renderer.MergeState),
		horzMerges: map[tw.Position]map[int]bool{
			tw.Header: make(map[int]bool),
			tw.Row:    make(map[int]bool),
			tw.Footer: make(map[int]bool),
		},
	}

	if err := t.calculateAndNormalizeWidths(ctx); err != nil {
		return nil, nil, err
	}

	ctx.headerLines, mctx.headerMerges, mctx.horzMerges[tw.Header] = t.prepareWithMerges(t.headers, t.config.Header, tw.Header)
	ctx.rowLines = make([][][]string, len(t.rows))
	for i, row := range t.rows {
		ctx.rowLines[i], mctx.rowMerges[i], mctx.horzMerges[tw.Row] = t.prepareWithMerges(row, t.config.Row, tw.Row)
	}
	if t.config.Row.Formatting.MergeMode == tw.MergeVertical || t.config.Row.Formatting.MergeMode == tw.MergeBoth {
		t.applyVerticalMerges(ctx, mctx)
	}
	ctx.footerLines, mctx.footerMerges, mctx.horzMerges[tw.Footer] = t.prepareWithMerges(t.footers, t.config.Footer, tw.Footer)

	return ctx, mctx, nil
}

// renderLine renders a single line for a section
func (t *Table) renderLine(ctx *renderContext, mctx *mergeContext, hctx *helperContext, aligns map[int]tw.Align, padding map[int]tw.Padding) error {
	resp := t.buildCellContexts(ctx, mctx, hctx, aligns, padding)
	f := ctx.renderer
	formatting := renderer.Formatting{
		Row: renderer.RowContext{
			Widths:       ctx.widths[hctx.position],
			ColMaxWidths: t.getColMaxWidths(hctx.position),
			Current:      resp.cells,
			Previous:     resp.prevCells,
			Next:         resp.nextCells,
			Position:     hctx.position,
			Location:     hctx.location,
		},
		Level:    t.getLevel(hctx.position),
		IsSubRow: hctx.lineIdx > 0 || t.config.Row.Padding.Global.Top != tw.Empty,
	}
	if hctx.position == tw.Row {
		formatting.HasFooter = len(ctx.footerLines) > 0
	}
	switch hctx.position {
	case tw.Header:
		f.Header(t.writer, ctx.headerLines, formatting)
	case tw.Row:
		f.Row(t.writer, hctx.line, formatting)
	case tw.Footer:
		f.Footer(t.writer, ctx.footerLines, formatting)
	}
	return nil
}

// renderPadding renders padding for a section
func (t *Table) renderPadding(ctx *renderContext, mctx *mergeContext, hctx *helperContext, padChar string) error {
	ctx.debug("Rendering %s padding for %s", padChar, hctx.position)
	colAligns := t.buildAligns(t.config.Row)
	colPadding := t.buildPadding(t.config.Row.Padding)
	switch hctx.position {
	case tw.Header:
		colAligns = t.buildAligns(t.config.Header)
		colPadding = t.buildPadding(t.config.Header.Padding)
	case tw.Footer:
		colAligns = t.buildAligns(t.config.Footer)
		colPadding = t.buildPadding(t.config.Footer.Padding)
	}
	return t.renderLine(ctx, mctx, hctx, colAligns, colPadding)
}

// appendSingle adds a single row to the table (internal helper)
func (t *Table) appendSingle(row interface{}) error {
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
						finalLines = append(finalLines, twfn.TruncateString(line, contentWidth-1, tw.CharEllipsis))
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

// applyVerticalMerges applies vertical merges across rows
func (t *Table) applyVerticalMerges(ctx *renderContext, mctx *mergeContext) {
	ctx.debug("Applying vertical merges across %d rows", len(ctx.rowLines))
	previousContent := make(map[int]string)
	for i := 0; i < len(ctx.rowLines); i++ {
		for j := 0; j < len(ctx.rowLines[i]); j++ {
			for col := 0; col < ctx.numCols; col++ {
				currentVal := strings.TrimSpace(ctx.rowLines[i][j][col])
				prevVal, exists := previousContent[col]
				if exists && currentVal == prevVal && currentVal != "" {
					if _, ok := mctx.rowMerges[i][col]; !ok {
						mctx.rowMerges[i][col] = renderer.MergeState{}
					}
					mergeState := mctx.rowMerges[i][col]
					mergeState.Vertical = true
					mctx.rowMerges[i][col] = mergeState
					ctx.rowLines[i][j][col] = tw.Empty
					ctx.debug("Vertical merge at row %d, line %d, col %d: cleared content", i, j, col)
					for k := i - 1; k >= 0; k-- {
						if len(ctx.rowLines[k]) > 0 && strings.TrimSpace(ctx.rowLines[k][0][col]) == prevVal {
							if _, ok := mctx.rowMerges[k][col]; !ok {
								mctx.rowMerges[k][col] = renderer.MergeState{}
							}
							startMerge := mctx.rowMerges[k][col]
							startMerge.Vertical = true
							startMerge.Span = i - k + 1
							startMerge.Start = true
							mctx.rowMerges[k][col] = startMerge
							break
						}
					}
				} else if currentVal != "" {
					if _, ok := mctx.rowMerges[i][col]; !ok {
						mctx.rowMerges[i][col] = renderer.MergeState{
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
	for i := 0; i < len(ctx.rowLines); i++ {
		for col, mergeState := range mctx.rowMerges[i] {
			if mergeState.Vertical && i == len(ctx.rowLines)-1 {
				mergeState.End = true
				mctx.rowMerges[i][col] = mergeState
			}
		}
	}
}

// calculateHorizontalSpan calculates the span of a horizontal merge
func (t *Table) calculateHorizontalSpan(horzMerges map[int]bool, startCol int) int {
	span := 1
	for col := startCol + 1; horzMerges[col]; col++ {
		span++
	}
	return span
}

// calculateAndNormalizeWidths computes and normalizes column widths
func (t *Table) calculateAndNormalizeWidths(ctx *renderContext) error {
	ctx.debug("Calculating and normalizing widths")
	for _, lines := range ctx.headerLines {
		t.updateWidths(lines, ctx.widths[tw.Header], t.config.Header.Padding)
	}
	ctx.debug("Header widths calculated: %v", ctx.widths[tw.Header])
	for _, row := range ctx.rowLines {
		for _, line := range row {
			t.updateWidths(line, ctx.widths[tw.Row], t.config.Row.Padding)
		}
	}
	ctx.debug("Row widths calculated: %v", ctx.widths[tw.Row])
	for _, lines := range ctx.footerLines {
		t.updateWidths(lines, ctx.widths[tw.Footer], t.config.Footer.Padding)
	}
	ctx.debug("Footer widths calculated: %v", ctx.widths[tw.Footer])

	ctx.debug("Normalizing widths for %d columns", ctx.numCols)
	for i := 0; i < ctx.numCols; i++ {
		maxWidth := 0
		for _, w := range []tw.Mapper[int, int]{ctx.widths[tw.Header], ctx.widths[tw.Row], ctx.widths[tw.Footer]} {
			if wd := w.Get(i); wd > maxWidth {
				maxWidth = wd
			}
		}
		ctx.widths[tw.Header].Set(i, maxWidth)
		ctx.widths[tw.Row].Set(i, maxWidth)
		ctx.widths[tw.Footer].Set(i, maxWidth)
	}
	ctx.debug("Normalized widths: header=%v, row=%v, footer=%v", ctx.widths[tw.Header], ctx.widths[tw.Row], ctx.widths[tw.Footer])
	return nil
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

// ensureInitialized ensures all required fields are initialized
func (t *Table) ensureInitialized() {
	if t.headerWidths == nil {
		t.headerWidths = tw.NewMapper[int, int]()
	}
	if t.rowWidths == nil {
		t.rowWidths = tw.NewMapper[int, int]()
	}
	if t.footerWidths == nil {
		t.footerWidths = tw.NewMapper[int, int]()
	}
	if t.renderer == nil {
		t.renderer = renderer.NewDefault()
	}
	t.debug("ensureInitialized called")
}

// getColMaxWidths retrieves the column max widths for a position
func (t *Table) getColMaxWidths(position tw.Position) map[int]int {
	switch position {
	case tw.Header:
		return t.config.Header.ColMaxWidths
	case tw.Row:
		return t.config.Row.ColMaxWidths
	case tw.Footer:
		return t.config.Footer.ColMaxWidths
	default:
		return nil
	}
}

// getLevel determines the rendering level for a position
func (t *Table) getLevel(position tw.Position) tw.Level {
	switch position {
	case tw.Header:
		return tw.LevelHeader
	case tw.Row:
		return tw.LevelBody
	case tw.Footer:
		return tw.LevelFooter
	default:
		return tw.LevelBody
	}
}

// determineLocation determines the location for headers or footers
func (t *Table) determineLocation(lineIdx, totalLines int, topPad, bottomPad string) tw.Location {
	if lineIdx == 0 && topPad == tw.Empty {
		return tw.LocationFirst
	}
	if lineIdx == totalLines-1 && bottomPad == tw.Empty {
		return tw.LocationEnd
	}
	return tw.LocationMiddle
}

// updateWidths calculates and updates column widths
func (t *Table) updateWidths(row []string, widths tw.Mapper[int, int], padding CellPadding) {
	t.debug("Updating widths for row: %v", row)
	for i, cell := range row {
		padLeftWidth := twfn.DisplayWidth(padding.Global.Left)
		padRightWidth := twfn.DisplayWidth(padding.Global.Right)
		if i < len(padding.PerColumn) && padding.PerColumn[i] != (tw.Padding{}) {
			padLeftWidth = twfn.DisplayWidth(padding.PerColumn[i].Left)
			padRightWidth = twfn.DisplayWidth(padding.PerColumn[i].Right)
		}
		totalWidth := twfn.DisplayWidth(strings.TrimSpace(cell)) + padLeftWidth + padRightWidth
		if current := widths.Get(i); totalWidth > current {
			widths.Set(i, totalWidth)
		}
	}
}

// padLine ensures a line is padded to the correct number of columns
func padLine(line []string, numCols int) []string {
	if len(line) >= numCols {
		return line
	}
	padded := make([]string, numCols)
	copy(padded, line)
	for i := len(line); i < numCols; i++ {
		padded[i] = tw.Empty
	}
	return padded
}
