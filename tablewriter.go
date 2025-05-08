package tablewriter

import (
	"database/sql"
	"fmt"
	"github.com/olekukonko/errors"
	"github.com/olekukonko/tablewriter/pkg/twwarp"
	"github.com/olekukonko/tablewriter/renderer"
	"github.com/olekukonko/tablewriter/tw"
	"github.com/olekukonko/tablewriter/twfn"
	"io"
	"reflect"
	"strings"
)

// Config represents the overall configuration for a table.
type Config struct {
	MaxWidth int           // Maximum width of the entire table (0 for unlimited)
	Header   tw.CellConfig // Configuration for the header section
	Row      tw.CellConfig // Configuration for the row section
	Footer   tw.CellConfig // Configuration for the footer section
	Debug    bool          // Enables debug logging when true
	AutoHide bool
}

// Table represents a table instance with content and rendering capabilities.
type Table struct {
	writer       io.Writer           // Destination for table output
	rows         [][][]string        // Row data, supporting multi-line cells
	headers      [][]string          // Header content
	footers      [][]string          // Footer content
	headerWidths tw.Mapper[int, int] // Computed widths for header columns
	rowWidths    tw.Mapper[int, int] // Computed widths for row columns
	footerWidths tw.Mapper[int, int] // Computed widths for footer columns
	renderer     tw.Renderer         // Engine for rendering the table
	config       Config              // Table configuration settings
	stringer     any                 // Function to convert rows to strings
	newLine      string              // Newline character (e.g., "\n")
	hasPrinted   bool                // Indicates if the table has been rendered
	trace        []string            // Debug trace log
}

// renderContext holds the core state for rendering the table.
type renderContext struct {
	table          *Table                                // Reference to the table instance
	renderer       tw.Renderer                           // Renderer instance
	cfg            tw.RendererConfig                     // Renderer configuration
	numCols        int                                   // Total number of columns
	headerLines    [][]string                            // Processed header lines
	rowLines       [][][]string                          // Processed row lines
	footerLines    [][]string                            // Processed footer lines
	widths         map[tw.Position]tw.Mapper[int, int]   // Widths per section
	debug          func(format string, a ...interface{}) // Debug logging function
	footerPrepared bool                                  // Tracks if footer is prepared

	emptyColumns    []bool // Tracks which original columns are empty (true if empty)
	visibleColCount int    // Count of columns that are NOT empty
}

// mergeContext holds state related to cell merging.
type mergeContext struct {
	headerMerges map[int]tw.MergeState        // Merge states for header columns
	rowMerges    []map[int]tw.MergeState      // Merge states for each row
	footerMerges map[int]tw.MergeState        // Merge states for footer columns
	horzMerges   map[tw.Position]map[int]bool // Tracks horizontal merges (unused)
}

// helperContext holds additional data for rendering helpers.
type helperContext struct {
	position tw.Position // Section being processed (Header, Row, Footer)
	rowIdx   int         // Row index within section
	lineIdx  int         // Line index within row
	location tw.Location // Boundary location (First, Middle, End)
	line     []string    // Current line content
}

// renderMergeResponse holds cell context data from rendering operations.
type renderMergeResponse struct {
	cells     map[int]tw.CellContext // Current line cells
	prevCells map[int]tw.CellContext // Previous line cells
	nextCells map[int]tw.CellContext // Next line cells
}

// ---- Public Methods ----

// NewTable creates a new table instance with the specified writer and options.
// Options can customize the table's configuration.
func NewTable(w io.Writer, opts ...Option) *Table {
	t := &Table{
		writer:       w,
		headerWidths: tw.NewMapper[int, int](),
		rowWidths:    tw.NewMapper[int, int](),
		footerWidths: tw.NewMapper[int, int](),
		renderer:     renderer.NewBlueprint(),
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

func (t *Table) Configure(fn func(*Config)) *Table {
	fn(&t.config)
	return t
}

// Append adds one or more rows to the table.
// Rows can be of any type if a stringer is provided.
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

// Bulk adds multiple rows from a slice to the table (legacy method).
// Expects a slice of rows compatible with the stringer or []string.
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

// Header configures the table's header content.
// Multi-line headers are supported via prepareContent.
func (t *Table) Header(headers []string) {
	t.ensureInitialized()
	t.debug("Setting header: %v", headers)
	prepared := t.prepareContent(headers, t.config.Header)
	t.headers = prepared
	t.debug("Header set, lines: %d", len(prepared))
}

// Footer configures the table's footer content, padding to match column count.
// Multi-line footers are supported.
func (t *Table) Footer(footers []string) {
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

// Render triggers the table rendering process to the writer.
// Returns an error if rendering fails.
func (t *Table) Render() error {
	return t.render()
}

// Renderer returns the current renderer instance used by the table.
func (t *Table) Renderer() tw.Renderer {
	t.debug("Renderer requested")
	return t.renderer
}

// Debug returns the accumulated debug trace, including renderer logs.
func (t *Table) Debug() []string {
	t.debug("Debug trace requested, returning %d entries", len(t.trace))
	return append(t.trace, t.renderer.Debug()...)
}

func (t *Table) Config() Config {
	return t.config
}

// render generates the table output using the configured renderer, handling headers, rows, footers, and cleanup.
// It initializes the renderer, renders sections, and ensures proper closure, continuing through section errors to attempt cleanup.
func (t *Table) render() error {
	t.ensureInitialized()

	// Prepare contexts first
	ctx, mctx, err := t.prepareContexts()
	if err != nil {
		return err
	}

	// Always call Start() before rendering sections
	ctx.debug("Calling renderer Start()")
	if err := ctx.renderer.Start(t.writer); err != nil {
		ctx.debug("Renderer Start() error: %v", err)
		return fmt.Errorf("renderer start failed: %w", err)
	}

	renderError := false
	for _, renderFn := range []func(*renderContext, *mergeContext) error{
		t.renderHeader,
		t.renderRow,
		t.renderFooter,
	} {
		if err := renderFn(ctx, mctx); err != nil {
			ctx.debug("Renderer section error: %v", err)
			renderError = true
			// Optional: break here if you don't want to attempt subsequent sections
		}
	}

	// Always call Close() after attempting to render sections
	ctx.debug("Calling renderer Close()")
	closeErr := ctx.renderer.Close(t.writer)
	if closeErr != nil {
		ctx.debug("Renderer Close() error: %v", closeErr)
		if !renderError { // Prioritize returning renderError if it happened
			return fmt.Errorf("renderer close failed: %w", closeErr)
		}
	}

	// If a rendering error happened earlier, return it now
	if renderError {
		return errors.New("table rendering failed in one or more sections")
	}

	t.hasPrinted = true
	ctx.debug("Render completed")
	return nil // Success
}

// renderHeader renders the table's header section, including borders and padding.
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

	// Render top border if enabled
	if cfg.Borders.Top.Enabled() && cfg.Settings.Lines.ShowTop.Enabled() {
		ctx.debug("Rendering table top border")
		nextCells := make(map[int]tw.CellContext)
		if len(ctx.headerLines) > 0 {
			for j, cell := range ctx.headerLines[0] {
				nextCells[j] = tw.CellContext{Data: cell, Merge: mctx.headerMerges[j]}
			}
		}
		f.Line(t.writer, tw.Formatting{
			Row: tw.RowContext{
				Widths:   ctx.widths[tw.Header],
				Next:     nextCells,
				Position: tw.Header,
				Location: tw.LocationFirst,
			},
			Level:    tw.LevelHeader,
			IsSubRow: false,
		})
	}

	// Render top padding if specified
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

	// Render header content lines with callbacks
	for i, line := range ctx.headerLines {
		hctx.rowIdx = 0
		hctx.lineIdx = i
		hctx.line = padLine(line, ctx.numCols)
		hctx.location = t.determineLocation(i, len(ctx.headerLines), t.config.Header.Padding.Global.Top, t.config.Header.Padding.Global.Bottom)

		// Execute callbacks before rendering each line
		if t.config.Header.Callbacks.Global != nil {
			ctx.debug("Executing global header callback for line %d", i)
			t.config.Header.Callbacks.Global()
		}
		for colIdx, cb := range t.config.Header.Callbacks.PerColumn {
			if colIdx < ctx.numCols && cb != nil {
				ctx.debug("Executing per-column header callback for line %d, col %d", i, colIdx)
				cb()
			}
		}

		if err := t.renderLine(ctx, mctx, hctx, colAligns, colPadding); err != nil {
			return err
		}
	}

	// Render bottom padding if specified
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

	// Render header separator if applicable
	if cfg.Settings.Lines.ShowHeaderLine.Enabled() && (len(ctx.rowLines) > 0 || len(ctx.footerLines) > 0) {
		ctx.debug("Rendering header separator line")
		hctx.rowIdx = 0
		hctx.lineIdx = len(ctx.headerLines) - 1
		hctx.line = padLine(ctx.headerLines[hctx.lineIdx], ctx.numCols)
		hctx.location = tw.LocationMiddle
		resp := t.buildCellContexts(ctx, mctx, hctx, colAligns, colPadding)
		f.Line(t.writer, tw.Formatting{
			Row: tw.RowContext{
				Widths:   ctx.widths[tw.Header],
				Current:  resp.cells,
				Previous: resp.prevCells,
				Next:     resp.nextCells,
				Position: tw.Header,
				Location: tw.LocationMiddle,
			},
			Level:    tw.LevelBody,
			IsSubRow: false,
		})
	}
	return nil
}

// renderRow renders the table's row section, including borders and padding.
func (t *Table) renderRow(ctx *renderContext, mctx *mergeContext) error {
	if len(ctx.rowLines) == 0 {
		return nil
	}
	ctx.debug("Rendering row section (total rows: %d)", len(ctx.rowLines))

	f := ctx.renderer
	cfg := ctx.cfg
	colAligns := t.buildAligns(t.config.Row)
	colPadding := t.buildPadding(t.config.Row.Padding)
	hctx := &helperContext{position: tw.Row}

	// Render top border for rows-only table
	footerIsEmptyOrNonExistent := !t.hasFooterElements()
	if len(ctx.headerLines) == 0 && footerIsEmptyOrNonExistent && cfg.Borders.Top.Enabled() && cfg.Settings.Lines.ShowTop.Enabled() {
		ctx.debug("Rendering table top border (rows only table)")
		nextCells := make(map[int]tw.CellContext)
		if len(ctx.rowLines) > 0 && len(ctx.rowLines[0]) > 0 && len(mctx.rowMerges) > 0 {
			firstLine := ctx.rowLines[0][0]
			firstMerges := mctx.rowMerges[0]
			for j, cell := range padLine(firstLine, ctx.numCols) {
				mergeState := tw.MergeState{}
				if firstMerges != nil {
					mergeState = firstMerges[j]
				}
				nextCells[j] = tw.CellContext{Data: cell, Merge: mergeState, Width: ctx.widths[tw.Row].Get(j)}
			}
		}
		f.Line(t.writer, tw.Formatting{
			Row: tw.RowContext{
				Widths:   ctx.widths[tw.Row],
				Next:     nextCells,
				Position: tw.Row,
				Location: tw.LocationFirst,
			},
			Level:    tw.LevelHeader,
			IsSubRow: false,
			Debug:    t.config.Debug,
		})
	}

	// Render all rows with padding and separators
	for i, lines := range ctx.rowLines {
		// Top row padding
		rowHasTopPadding := t.config.Row.Padding.Global.Top != tw.Empty
		if rowHasTopPadding {
			hctx.rowIdx = i
			hctx.lineIdx = -1
			if i == 0 && len(ctx.headerLines) == 0 {
				hctx.location = tw.LocationFirst
			} else {
				hctx.location = tw.LocationMiddle
			}

			topPadChar := t.config.Row.Padding.Global.Top
			topPaddingLineContent := make([]string, ctx.numCols)
			topPadWidth := twfn.DisplayWidth(topPadChar)
			if topPadWidth < 1 {
				topPadWidth = 1
			}

			for j := 0; j < ctx.numCols; j++ {
				colWd := ctx.widths[tw.Row].Get(j)
				mergeState := tw.MergeState{}
				if i < len(mctx.rowMerges) && mctx.rowMerges[i] != nil {
					mergeState = mctx.rowMerges[i][j]
				}
				if mergeState.Horizontal.Present && !mergeState.Horizontal.Start {
					topPaddingLineContent[j] = ""
					continue
				}
				repeatCount := 0
				if colWd > 0 && topPadWidth > 0 {
					repeatCount = colWd / topPadWidth
				}
				if colWd > 0 && repeatCount < 1 {
					repeatCount = 1
				}
				if colWd == 0 {
					repeatCount = 0
				}
				rawPaddingContent := strings.Repeat(topPadChar, repeatCount)
				currentWd := twfn.DisplayWidth(rawPaddingContent)
				if currentWd < colWd {
					rawPaddingContent += strings.Repeat(" ", colWd-currentWd)
				}
				if currentWd > colWd && colWd > 0 {
					rawPaddingContent = twfn.TruncateString(rawPaddingContent, colWd)
				}
				if colWd == 0 {
					rawPaddingContent = ""
				}
				topPaddingLineContent[j] = rawPaddingContent
			}
			hctx.line = topPaddingLineContent

			ctx.debug("Calling renderPadding for Row Top Padding (row %d): %v (loc: %v)", i, hctx.line, hctx.location)
			if err := t.renderPadding(ctx, mctx, hctx, topPadChar); err != nil {
				return err
			}
		}

		// Row content lines
		footerExists := t.hasFooterElements()
		rowHasBottomPadding := t.config.Row.Padding.Global.Bottom != tw.Empty

		for j, line := range lines {
			hctx.rowIdx = i
			hctx.lineIdx = j
			hctx.line = padLine(line, ctx.numCols)

			isFirstRow := i == 0
			isLastRow := i == len(ctx.rowLines)-1
			isFirstLineOfRow := j == 0
			isLastLineOfRow := j == len(lines)-1

			if isFirstRow && isFirstLineOfRow && !rowHasTopPadding && len(ctx.headerLines) == 0 {
				hctx.location = tw.LocationFirst
			} else if isLastRow && isLastLineOfRow && !rowHasBottomPadding && !footerExists {
				hctx.location = tw.LocationEnd
			} else {
				hctx.location = tw.LocationMiddle
			}

			ctx.debug("Rendering row %d line %d with location %v", i, j, hctx.location)
			if err := t.renderLine(ctx, mctx, hctx, colAligns, colPadding); err != nil {
				return err
			}

			shouldDrawSeparator := cfg.Settings.Separators.BetweenRows.Enabled() &&
				!(isLastRow && isLastLineOfRow && (footerExists || rowHasBottomPadding || (cfg.Borders.Bottom.Enabled() && cfg.Settings.Lines.ShowBottom.Enabled())))

			if shouldDrawSeparator {
				ctx.debug("Rendering between-rows separator after row %d line %d", i, j)
				resp := t.buildCellContexts(ctx, mctx, hctx, colAligns, colPadding)

				nextCells := make(map[int]tw.CellContext)
				nextRowIdx := i
				nextLineIdx := j + 1
				var nextRowMerges map[int]tw.MergeState

				if nextLineIdx >= len(lines) {
					nextRowIdx = i + 1
					nextLineIdx = 0
				}

				if nextRowIdx < len(ctx.rowLines) && nextRowIdx < len(mctx.rowMerges) {
					if len(ctx.rowLines[nextRowIdx]) > nextLineIdx {
						nextLineData := ctx.rowLines[nextRowIdx][nextLineIdx]
						nextRowMerges = mctx.rowMerges[nextRowIdx]
						for k, cell := range padLine(nextLineData, ctx.numCols) {
							mergeState := tw.MergeState{}
							if nextRowMerges != nil {
								mergeState = nextRowMerges[k]
							}
							nextCells[k] = tw.CellContext{Data: cell, Merge: mergeState, Width: ctx.widths[tw.Row].Get(k)}
						}
						ctx.debug("Separator context: Next is row %d line %d", nextRowIdx, nextLineIdx)
					} else if nextLineIdx == 0 && len(ctx.rowLines[nextRowIdx]) == 0 {
						ctx.debug("Separator context: Next row %d is empty", nextRowIdx)
						nextCells = nil
					} else {
						ctx.debug("Separator context: Unexpected end of lines for next row %d", nextRowIdx)
						nextCells = nil
					}
				} else {
					ctx.debug("Separator context: No next row.")
					nextCells = nil
				}

				f.Line(t.writer, tw.Formatting{
					Row: tw.RowContext{
						Widths:   ctx.widths[tw.Row],
						Current:  resp.cells,
						Previous: resp.prevCells,
						Next:     nextCells,
						Position: tw.Row,
						Location: tw.LocationMiddle,
					},
					Level:     tw.LevelBody,
					IsSubRow:  false,
					HasFooter: footerExists,
					Debug:     t.config.Debug,
				})
			} else if cfg.Settings.Separators.BetweenRows.Enabled() && isLastRow && isLastLineOfRow {
				ctx.debug("Skipping between-rows separator after last row %d line %d (footerExists=%v, rowHasBottomPadding=%v, bottomBorderEnabled=%v)",
					i, j, footerExists, rowHasBottomPadding, cfg.Borders.Bottom.Enabled() && cfg.Settings.Lines.ShowBottom.Enabled())
			}
		}

		// Bottom row padding
		if rowHasBottomPadding {
			hctx.rowIdx = i
			hctx.lineIdx = len(lines)
			if i == len(ctx.rowLines)-1 && !footerExists {
				hctx.location = tw.LocationEnd
			} else {
				hctx.location = tw.LocationMiddle
			}

			bottomPadChar := t.config.Row.Padding.Global.Bottom
			bottomPaddingLineContent := make([]string, ctx.numCols)
			bottomPadWidth := twfn.DisplayWidth(bottomPadChar)
			if bottomPadWidth < 1 {
				bottomPadWidth = 1
			}

			for j := 0; j < ctx.numCols; j++ {
				colWd := ctx.widths[tw.Row].Get(j)
				mergeState := tw.MergeState{}
				if i < len(mctx.rowMerges) && mctx.rowMerges[i] != nil {
					mergeState = mctx.rowMerges[i][j]
				}
				if mergeState.Horizontal.Present && !mergeState.Horizontal.Start {
					bottomPaddingLineContent[j] = ""
					continue
				}
				repeatCount := 0
				if colWd > 0 && bottomPadWidth > 0 {
					repeatCount = colWd / bottomPadWidth
				}
				if colWd > 0 && repeatCount < 1 {
					repeatCount = 1
				}
				if colWd == 0 {
					repeatCount = 0
				}
				rawPaddingContent := strings.Repeat(bottomPadChar, repeatCount)
				currentWd := twfn.DisplayWidth(rawPaddingContent)
				if currentWd < colWd {
					rawPaddingContent += strings.Repeat(" ", colWd-currentWd)
				}
				if currentWd > colWd && colWd > 0 {
					rawPaddingContent = twfn.TruncateString(rawPaddingContent, colWd)
				}
				if colWd == 0 {
					rawPaddingContent = ""
				}
				bottomPaddingLineContent[j] = rawPaddingContent
			}
			hctx.line = bottomPaddingLineContent

			ctx.debug("Calling renderPadding for Row Bottom Padding (row %d): %v (loc: %v)", i, hctx.line, hctx.location)
			if err := t.renderPadding(ctx, mctx, hctx, bottomPadChar); err != nil {
				return err
			}
		}
	}
	return nil
}

// renderFooter renders the table's footer section, including borders and padding.
func (t *Table) renderFooter(ctx *renderContext, mctx *mergeContext) error {
	if !ctx.footerPrepared {
		t.prepareFooter(ctx, mctx)
	}

	f := ctx.renderer
	cfg := ctx.cfg
	colAligns := t.buildAligns(t.config.Footer)
	colPadding := t.buildPadding(t.config.Footer.Padding)
	hctx := &helperContext{position: tw.Footer}

	hasContent := len(ctx.footerLines) > 0
	hasTopPadding := t.config.Footer.Padding.Global.Top != tw.Empty
	hasBottomPaddingConfig := t.config.Footer.Padding.Global.Bottom != tw.Empty || t.hasPerColumnBottomPadding()
	hasAnyFooterElement := hasContent || hasTopPadding || hasBottomPaddingConfig

	if !hasAnyFooterElement {
		hasContentAbove := len(ctx.rowLines) > 0 || len(ctx.headerLines) > 0
		if hasContentAbove && cfg.Borders.Bottom.Enabled() && cfg.Settings.Lines.ShowBottom.Enabled() {
			ctx.debug("Footer is empty, rendering table bottom border based on last row/header")
			var lastLineAboveCtx *helperContext
			var lastLineAligns map[int]tw.Align
			var lastLinePadding map[int]tw.Padding

			if len(ctx.rowLines) > 0 {
				lastRowIdx := len(ctx.rowLines) - 1
				lastRowLineIdx := -1
				var lastRowLine []string
				if lastRowIdx >= 0 && len(ctx.rowLines[lastRowIdx]) > 0 {
					lastRowLineIdx = len(ctx.rowLines[lastRowIdx]) - 1
					lastRowLine = padLine(ctx.rowLines[lastRowIdx][lastRowLineIdx], ctx.numCols)
				} else {
					lastRowLine = make([]string, ctx.numCols)
				}
				lastLineAboveCtx = &helperContext{
					position: tw.Row,
					rowIdx:   lastRowIdx,
					lineIdx:  lastRowLineIdx,
					line:     lastRowLine,
					location: tw.LocationEnd,
				}
				lastLineAligns = t.buildAligns(t.config.Row)
				lastLinePadding = t.buildPadding(t.config.Row.Padding)
			} else {
				lastHeaderLineIdx := -1
				var lastHeaderLine []string
				if len(ctx.headerLines) > 0 {
					lastHeaderLineIdx = len(ctx.headerLines) - 1
					lastHeaderLine = padLine(ctx.headerLines[lastHeaderLineIdx], ctx.numCols)
				} else {
					lastHeaderLine = make([]string, ctx.numCols)
				}
				lastLineAboveCtx = &helperContext{
					position: tw.Header,
					rowIdx:   0,
					lineIdx:  lastHeaderLineIdx,
					line:     lastHeaderLine,
					location: tw.LocationEnd,
				}
				lastLineAligns = t.buildAligns(t.config.Header)
				lastLinePadding = t.buildPadding(t.config.Header.Padding)
			}

			resp := t.buildCellContexts(ctx, mctx, lastLineAboveCtx, lastLineAligns, lastLinePadding)
			ctx.debug("Bottom border: Using Widths=%v", ctx.widths[tw.Row])
			f.Line(t.writer, tw.Formatting{
				Row: tw.RowContext{
					Widths:       ctx.widths[tw.Row],
					Current:      resp.cells,
					Previous:     resp.prevCells,
					Position:     lastLineAboveCtx.position,
					Location:     tw.LocationEnd,
					ColMaxWidths: t.getColMaxWidths(tw.Footer),
				},
				Level:    tw.LevelFooter,
				IsSubRow: false,
				Debug:    t.config.Debug,
			})
		} else {
			ctx.debug("Footer is empty and no content above or borders disabled, skipping footer render")
		}
		return nil
	}

	ctx.debug("Rendering footer section (has elements)")
	hasContentAbove := len(ctx.rowLines) > 0 || len(ctx.headerLines) > 0
	if hasContentAbove && cfg.Settings.Lines.ShowFooterLine.Enabled() && !hasTopPadding && len(ctx.footerLines) > 0 {
		ctx.debug("Rendering footer separator line")
		var lastLineAboveCtx *helperContext
		var lastLineAligns map[int]tw.Align
		var lastLinePadding map[int]tw.Padding
		var lastLinePosition tw.Position

		if len(ctx.rowLines) > 0 {
			lastRowIdx := len(ctx.rowLines) - 1
			lastRowLineIdx := -1
			var lastRowLine []string
			if lastRowIdx >= 0 && len(ctx.rowLines[lastRowIdx]) > 0 {
				lastRowLineIdx = len(ctx.rowLines[lastRowIdx]) - 1
				lastRowLine = padLine(ctx.rowLines[lastRowIdx][lastRowLineIdx], ctx.numCols)
			} else {
				lastRowLine = make([]string, ctx.numCols)
			}
			lastLineAboveCtx = &helperContext{
				position: tw.Row,
				rowIdx:   lastRowIdx,
				lineIdx:  lastRowLineIdx,
				line:     lastRowLine,
				location: tw.LocationMiddle,
			}
			lastLineAligns = t.buildAligns(t.config.Row)
			lastLinePadding = t.buildPadding(t.config.Row.Padding)
			lastLinePosition = tw.Row
		} else {
			lastHeaderLineIdx := -1
			var lastHeaderLine []string
			if len(ctx.headerLines) > 0 {
				lastHeaderLineIdx = len(ctx.headerLines) - 1
				lastHeaderLine = padLine(ctx.headerLines[lastHeaderLineIdx], ctx.numCols)
			} else {
				lastHeaderLine = make([]string, ctx.numCols)
			}
			lastLineAboveCtx = &helperContext{
				position: tw.Header,
				rowIdx:   0,
				lineIdx:  lastHeaderLineIdx,
				line:     lastHeaderLine,
				location: tw.LocationMiddle,
			}
			lastLineAligns = t.buildAligns(t.config.Header)
			lastLinePadding = t.buildPadding(t.config.Header.Padding)
			lastLinePosition = tw.Header
		}

		resp := t.buildCellContexts(ctx, mctx, lastLineAboveCtx, lastLineAligns, lastLinePadding)
		var nextCells map[int]tw.CellContext
		if hasContent {
			nextCells = make(map[int]tw.CellContext)
			for j, cellData := range padLine(ctx.footerLines[0], ctx.numCols) {
				mergeState := tw.MergeState{}
				if mctx.footerMerges != nil {
					mergeState = mctx.footerMerges[j]
				}
				nextCells[j] = tw.CellContext{Data: cellData, Merge: mergeState, Width: ctx.widths[tw.Footer].Get(j)}
			}
		}
		ctx.debug("Footer separator: Using Widths=%v", ctx.widths[tw.Row])
		f.Line(t.writer, tw.Formatting{
			Row: tw.RowContext{
				Widths:       ctx.widths[tw.Row],
				Current:      resp.cells,
				Previous:     resp.prevCells,
				Next:         nextCells,
				Position:     lastLinePosition,
				Location:     tw.LocationMiddle,
				ColMaxWidths: t.getColMaxWidths(tw.Footer),
			},
			Level:     tw.LevelFooter,
			IsSubRow:  false,
			HasFooter: true,
			Debug:     t.config.Debug,
		})
	}

	if hasTopPadding {
		topPadChar := t.config.Footer.Padding.Global.Top
		topPaddingLineContent := make([]string, ctx.numCols)
		topPadWidth := twfn.DisplayWidth(topPadChar)
		if topPadWidth < 1 {
			topPadWidth = 1
		}
		ctx.debug("Constructing Footer Global Top Padding line content")
		for j := 0; j < ctx.numCols; j++ {
			colWd := ctx.widths[tw.Footer].Get(j)
			mergeState := tw.MergeState{}
			if mctx.footerMerges != nil {
				mergeState = mctx.footerMerges[j]
			}
			if mergeState.Horizontal.Present && !mergeState.Horizontal.Start {
				topPaddingLineContent[j] = ""
				continue
			}
			repeatCount := 0
			if colWd > 0 && topPadWidth > 0 {
				repeatCount = colWd / topPadWidth
			}
			if colWd > 0 && repeatCount < 1 {
				repeatCount = 1
			}
			if colWd == 0 {
				repeatCount = 0
			}
			rawPaddingContent := strings.Repeat(topPadChar, repeatCount)
			currentWd := twfn.DisplayWidth(rawPaddingContent)
			if currentWd < colWd {
				rawPaddingContent += strings.Repeat(" ", colWd-currentWd)
			}
			if currentWd > colWd && colWd > 0 {
				rawPaddingContent = twfn.TruncateString(rawPaddingContent, colWd)
			}
			if colWd == 0 {
				rawPaddingContent = ""
			}
			topPaddingLineContent[j] = rawPaddingContent
		}
		hctx.rowIdx = 0
		hctx.lineIdx = -1
		hctx.line = topPaddingLineContent
		if !(hasContentAbove && cfg.Settings.Lines.ShowFooterLine.Enabled()) {
			hctx.location = tw.LocationFirst
		} else {
			hctx.location = tw.LocationMiddle
		}
		ctx.debug("Calling renderPadding for Footer Top Padding line: %v (loc: %v)", hctx.line, hctx.location)
		if err := t.renderPadding(ctx, mctx, hctx, topPadChar); err != nil {
			return err
		}
	}

	lastRenderedLineIdx := -2
	if hasTopPadding {
		lastRenderedLineIdx = -1
	}
	for i, line := range ctx.footerLines {
		hctx.rowIdx = 0
		hctx.lineIdx = i
		hctx.line = padLine(line, ctx.numCols)
		isFirstContentLine := i == 0
		isLastContentLine := i == len(ctx.footerLines)-1
		if isFirstContentLine && !hasTopPadding && !(hasContentAbove && cfg.Settings.Lines.ShowFooterLine.Enabled()) {
			hctx.location = tw.LocationFirst
		} else if isLastContentLine && !hasBottomPaddingConfig {
			hctx.location = tw.LocationEnd
		} else {
			hctx.location = tw.LocationMiddle
		}
		ctx.debug("Rendering footer content line %d with location %v", i, hctx.location)
		if err := t.renderLine(ctx, mctx, hctx, colAligns, colPadding); err != nil {
			return err
		}
		lastRenderedLineIdx = i
	}

	var paddingLineContentForContext []string
	if hasBottomPaddingConfig {
		paddingLineContentForContext = make([]string, ctx.numCols)
		formattedPaddingCells := make([]string, ctx.numCols)
		var representativePadChar string = " "
		ctx.debug("Constructing Footer Bottom Padding line content strings")
		for j := 0; j < ctx.numCols; j++ {
			colWd := ctx.widths[tw.Footer].Get(j)
			mergeState := tw.MergeState{}
			if mctx.footerMerges != nil {
				mergeState = mctx.footerMerges[j]
			}
			if mergeState.Horizontal.Present && !mergeState.Horizontal.Start {
				paddingLineContentForContext[j] = ""
				formattedPaddingCells[j] = ""
				continue
			}
			padChar := " "
			if j < len(t.config.Footer.Padding.PerColumn) && t.config.Footer.Padding.PerColumn[j].Bottom != tw.Empty {
				padChar = t.config.Footer.Padding.PerColumn[j].Bottom
			} else if t.config.Footer.Padding.Global.Bottom != tw.Empty {
				padChar = t.config.Footer.Padding.Global.Bottom
			}
			paddingLineContentForContext[j] = padChar
			if j == 0 || representativePadChar == " " {
				representativePadChar = padChar
			}
			padWidth := twfn.DisplayWidth(padChar)
			if padWidth < 1 {
				padWidth = 1
			}
			repeatCount := 0
			if colWd > 0 && padWidth > 0 {
				repeatCount = colWd / padWidth
			}
			if colWd > 0 && repeatCount < 1 && padChar != " " {
				repeatCount = 1
			}
			if colWd == 0 {
				repeatCount = 0
			}
			rawPaddingContent := strings.Repeat(padChar, repeatCount)
			currentWd := twfn.DisplayWidth(rawPaddingContent)
			if currentWd < colWd {
				rawPaddingContent += strings.Repeat(" ", colWd-currentWd)
			}
			if currentWd > colWd && colWd > 0 {
				rawPaddingContent = twfn.TruncateString(rawPaddingContent, colWd)
			}
			if colWd == 0 {
				rawPaddingContent = ""
			}
			formattedPaddingCells[j] = rawPaddingContent
		}
		ctx.debug("Manually rendering Footer Bottom Padding line (char like '%s')", representativePadChar)
		var paddingLineOutput strings.Builder
		if cfg.Borders.Left.Enabled() {
			paddingLineOutput.WriteString(cfg.Symbols.Column())
		}
		for colIdx := 0; colIdx < ctx.numCols; {
			if colIdx > 0 && cfg.Settings.Separators.BetweenColumns.Enabled() {
				shouldAddSeparator := true
				if prevMergeState, ok := mctx.footerMerges[colIdx-1]; ok {
					if prevMergeState.Horizontal.Present && !prevMergeState.Horizontal.End {
						shouldAddSeparator = false
					}
				}
				if shouldAddSeparator {
					paddingLineOutput.WriteString(cfg.Symbols.Column())
				}
			}
			if colIdx < len(formattedPaddingCells) {
				paddingLineOutput.WriteString(formattedPaddingCells[colIdx])
			}
			currentMergeState := tw.MergeState{}
			if mctx.footerMerges != nil {
				if state, ok := mctx.footerMerges[colIdx]; ok {
					currentMergeState = state
				}
			}
			if currentMergeState.Horizontal.Present && currentMergeState.Horizontal.Start {
				colIdx += currentMergeState.Horizontal.Span
			} else {
				colIdx++
			}
		}
		if cfg.Borders.Right.Enabled() {
			paddingLineOutput.WriteString(cfg.Symbols.Column())
		}
		paddingLineOutput.WriteString(t.newLine)
		fmt.Fprint(t.writer, paddingLineOutput.String())
		ctx.debug("Manually rendered Footer Bottom Padding line: %s", strings.TrimSuffix(paddingLineOutput.String(), t.newLine))
		hctx.rowIdx = 0
		hctx.lineIdx = len(ctx.footerLines)
		hctx.line = paddingLineContentForContext
		hctx.location = tw.LocationEnd
		lastRenderedLineIdx = hctx.lineIdx
	}

	if cfg.Borders.Bottom.Enabled() && cfg.Settings.Lines.ShowBottom.Enabled() {
		ctx.debug("Rendering final table bottom border")
		if lastRenderedLineIdx == len(ctx.footerLines) {
			hctx.rowIdx = 0
			hctx.lineIdx = lastRenderedLineIdx
			hctx.line = paddingLineContentForContext
			hctx.location = tw.LocationEnd
			ctx.debug("Setting border context based on bottom padding line")
		} else if lastRenderedLineIdx >= 0 {
			hctx.rowIdx = 0
			hctx.lineIdx = lastRenderedLineIdx
			hctx.line = padLine(ctx.footerLines[hctx.lineIdx], ctx.numCols)
			hctx.location = tw.LocationEnd
			ctx.debug("Setting border context based on last content line idx %d", hctx.lineIdx)
		} else if lastRenderedLineIdx == -1 {
			hctx.rowIdx = 0
			hctx.lineIdx = -1
			if hctx.line == nil {
				hctx.line = make([]string, ctx.numCols)
			}
			hctx.location = tw.LocationEnd
			ctx.debug("Setting border context based on top padding line")
		} else {
			hctx.rowIdx = 0
			hctx.lineIdx = -2
			hctx.line = make([]string, ctx.numCols)
			hctx.location = tw.LocationEnd
			ctx.debug("Warning: Cannot determine context for bottom border")
		}
		resp := t.buildCellContexts(ctx, mctx, hctx, colAligns, colPadding)
		ctx.debug("Bottom border: Using Widths=%v", ctx.widths[tw.Row])
		f.Line(t.writer, tw.Formatting{
			Row: tw.RowContext{
				Widths:       ctx.widths[tw.Row],
				Current:      resp.cells,
				Previous:     resp.prevCells,
				Position:     tw.Footer,
				Location:     tw.LocationEnd,
				ColMaxWidths: t.getColMaxWidths(tw.Footer),
			},
			Level:    tw.LevelFooter,
			IsSubRow: false,
			Debug:    t.config.Debug,
		})
	} else {
		ctx.debug("Skipping final table bottom border rendering (disabled or not applicable)")
	}

	return nil
}

// buildCellContexts constructs CellContext objects for a given line.
func (t *Table) buildCellContexts(ctx *renderContext, mctx *mergeContext, hctx *helperContext, aligns map[int]tw.Align, padding map[int]tw.Padding) renderMergeResponse {
	cells := make(map[int]tw.CellContext)
	var merges map[int]tw.MergeState

	// Determine the correct merge map based on the current position
	switch hctx.position {
	case tw.Header:
		merges = mctx.headerMerges
	case tw.Row:
		// Safety check for row index bounds
		if hctx.rowIdx >= 0 && hctx.rowIdx < len(mctx.rowMerges) && mctx.rowMerges[hctx.rowIdx] != nil {
			merges = mctx.rowMerges[hctx.rowIdx]
		} else {
			merges = make(map[int]tw.MergeState) // Use empty if out of bounds or nil
			t.debug("Warning: Invalid row index %d or nil merges in buildCellContexts", hctx.rowIdx)
		}
	case tw.Footer:
		merges = mctx.footerMerges
	default:
		merges = make(map[int]tw.MergeState) // Default to empty map for unknown position
		t.debug("Warning: Invalid position '%s' in buildCellContexts", hctx.position)
	}

	// Ensure merges map is not nil
	if merges == nil {
		merges = make(map[int]tw.MergeState)
		t.debug("Warning: merges map was nil in buildCellContexts after switch, using empty map")
	}

	// Build CellContext for each column
	for j := 0; j < ctx.numCols; j++ {
		mergeState := merges[j] // Get merge state for the column
		cellData := ""
		if j < len(hctx.line) {
			cellData = hctx.line[j] // Get data if available
		}

		// Get the potentially adjusted width for this column and position
		finalColWidth := ctx.widths[hctx.position].Get(j)

		cells[j] = tw.CellContext{
			Data:    cellData,
			Align:   aligns[j],     // Use provided aligns map
			Padding: padding[j],    // Use provided padding map
			Width:   finalColWidth, // Use the FINAL adjusted width
			Merge:   mergeState,
		}
	}

	// Build contexts for adjacent cells (they also need adjusted widths)
	prevCells := t.buildAdjacentCells(ctx, mctx, hctx, -1)
	nextCells := t.buildAdjacentCells(ctx, mctx, hctx, +1)

	return renderMergeResponse{
		cells:     cells,
		prevCells: prevCells,
		nextCells: nextCells,
	}
}

// buildAdjacentCells constructs cell contexts for adjacent lines (previous or next).
func (t *Table) buildAdjacentCells(ctx *renderContext, mctx *mergeContext, hctx *helperContext, direction int) map[int]tw.CellContext {
	adjCells := make(map[int]tw.CellContext)
	var adjLine []string
	var adjMerges map[int]tw.MergeState
	found := false
	adjPosition := hctx.position // Assume adjacent line is in the same section initially

	switch hctx.position {
	case tw.Header:
		targetLineIdx := hctx.lineIdx + direction
		if direction < 0 { // Previous
			if targetLineIdx >= 0 && targetLineIdx < len(ctx.headerLines) {
				adjLine = ctx.headerLines[targetLineIdx]
				adjMerges = mctx.headerMerges
				found = true
			} // If targetLineIdx < 0, there's no previous line (return nil later)
		} else { // Next
			if targetLineIdx < len(ctx.headerLines) {
				adjLine = ctx.headerLines[targetLineIdx]
				adjMerges = mctx.headerMerges
				found = true
			} else if len(ctx.rowLines) > 0 && len(ctx.rowLines[0]) > 0 && len(mctx.rowMerges) > 0 { // Transition to first row
				adjLine = ctx.rowLines[0][0]
				adjMerges = mctx.rowMerges[0]
				adjPosition = tw.Row
				found = true
			} else if len(ctx.footerLines) > 0 { // Transition to footer (if no rows)
				adjLine = ctx.footerLines[0]
				adjMerges = mctx.footerMerges
				adjPosition = tw.Footer
				found = true
			}
		}
	case tw.Row:
		targetLineIdx := hctx.lineIdx + direction
		// Safety check row index
		if hctx.rowIdx < 0 || hctx.rowIdx >= len(ctx.rowLines) || hctx.rowIdx >= len(mctx.rowMerges) {
			t.debug("Warning: Invalid row index %d in buildAdjacentCells", hctx.rowIdx)
			return nil // Cannot determine adjacent safely
		}
		currentRowLines := ctx.rowLines[hctx.rowIdx]
		currentMerges := mctx.rowMerges[hctx.rowIdx]

		if direction < 0 { // Previous
			if targetLineIdx >= 0 && targetLineIdx < len(currentRowLines) { // Previous line within the same logical row
				adjLine = currentRowLines[targetLineIdx]
				adjMerges = currentMerges
				found = true
			} else if targetLineIdx < 0 { // Need previous logical row or header
				targetRowIdx := hctx.rowIdx - 1
				if targetRowIdx >= 0 && targetRowIdx < len(ctx.rowLines) && targetRowIdx < len(mctx.rowMerges) { // Previous logical row exists
					prevRowLines := ctx.rowLines[targetRowIdx]
					if len(prevRowLines) > 0 { // If prev row has lines
						adjLine = prevRowLines[len(prevRowLines)-1] // Last line of prev row
						adjMerges = mctx.rowMerges[targetRowIdx]
						found = true
					} // If prev row is empty, keep searching upwards? Assume transitions handled by header check.
				} else if len(ctx.headerLines) > 0 { // Transition to last header line
					adjLine = ctx.headerLines[len(ctx.headerLines)-1]
					adjMerges = mctx.headerMerges
					adjPosition = tw.Header
					found = true
				}
			}
		} else { // Next
			if targetLineIdx >= 0 && targetLineIdx < len(currentRowLines) { // Next line within the same logical row
				adjLine = currentRowLines[targetLineIdx]
				adjMerges = currentMerges
				found = true
			} else if targetLineIdx >= len(currentRowLines) { // Need next logical row or footer
				targetRowIdx := hctx.rowIdx + 1
				// Check if next logical row exists and has lines
				if targetRowIdx < len(ctx.rowLines) && targetRowIdx < len(mctx.rowMerges) && len(ctx.rowLines[targetRowIdx]) > 0 {
					adjLine = ctx.rowLines[targetRowIdx][0] // First line of next row
					adjMerges = mctx.rowMerges[targetRowIdx]
					found = true
				} else if len(ctx.footerLines) > 0 { // Transition to first footer line
					adjLine = ctx.footerLines[0]
					adjMerges = mctx.footerMerges
					adjPosition = tw.Footer
					found = true
				}
			}
		}
	case tw.Footer:
		targetLineIdx := hctx.lineIdx + direction
		if direction < 0 { // Previous
			if targetLineIdx >= 0 && targetLineIdx < len(ctx.footerLines) {
				adjLine = ctx.footerLines[targetLineIdx]
				adjMerges = mctx.footerMerges
				found = true
			} else if targetLineIdx < 0 { // Transition to last row or last header
				if len(ctx.rowLines) > 0 {
					lastRowIdx := len(ctx.rowLines) - 1
					if lastRowIdx < len(mctx.rowMerges) && len(ctx.rowLines[lastRowIdx]) > 0 {
						lastRowLines := ctx.rowLines[lastRowIdx]
						adjLine = lastRowLines[len(lastRowLines)-1]
						adjMerges = mctx.rowMerges[lastRowIdx]
						adjPosition = tw.Row
						found = true
					}
				} else if len(ctx.headerLines) > 0 { // Fallback to header if no rows
					adjLine = ctx.headerLines[len(ctx.headerLines)-1]
					adjMerges = mctx.headerMerges
					adjPosition = tw.Header
					found = true
				}
			}
		} else { // Next
			if targetLineIdx >= 0 && targetLineIdx < len(ctx.footerLines) {
				adjLine = ctx.footerLines[targetLineIdx]
				adjMerges = mctx.footerMerges
				found = true
			} // If targetLineIdx >= len(ctx.footerLines), there is no next line (return nil later)
		}
	}

	if !found {
		return nil // No adjacent line exists
	}

	// Ensure adjMerges is not nil (safety)
	if adjMerges == nil {
		adjMerges = make(map[int]tw.MergeState)
		t.debug("Warning: adjMerges was nil in buildAdjacentCells despite found=true")
	}

	// Pad the adjacent line to the full original column count
	paddedAdjLine := padLine(adjLine, ctx.numCols)

	// Build CellContext for each column using the *adjusted* widths
	for j := 0; j < ctx.numCols; j++ {
		mergeState := adjMerges[j] // Get merge state for the column in the adjacent line
		cellData := paddedAdjLine[j]

		// Get the potentially adjusted width for the *adjacent cell's* position and column
		finalAdjColWidth := ctx.widths[adjPosition].Get(j)

		adjCells[j] = tw.CellContext{
			Data:  cellData,
			Merge: mergeState,
			Width: finalAdjColWidth, // Use the FINAL adjusted width for the adjacent cell
			// Align and Padding are not typically needed for context, only Width and Merge
		}
	}
	return adjCells
}

// ---- Helpers ----

// defaultConfig returns the default configuration for a table.
func defaultConfig() Config {
	defaultPadding := tw.Padding{Left: tw.Space, Right: tw.Space, Top: tw.Empty, Bottom: tw.Empty}
	return Config{
		MaxWidth: 0,
		Header: tw.CellConfig{
			Formatting: tw.CellFormatting{
				MaxWidth:   0,
				AutoWrap:   tw.WrapTruncate,
				Alignment:  tw.AlignCenter,
				AutoFormat: true,
				MergeMode:  tw.MergeNone,
			},
			Padding: tw.CellPadding{
				Global: defaultPadding,
			},
		},
		Row: tw.CellConfig{
			Formatting: tw.CellFormatting{
				MaxWidth:   0,
				AutoWrap:   tw.WrapNormal,
				Alignment:  tw.AlignLeft,
				AutoFormat: false,
				MergeMode:  tw.MergeNone,
			},
			Padding: tw.CellPadding{
				Global: defaultPadding,
			},
		},
		Footer: tw.CellConfig{
			Formatting: tw.CellFormatting{
				MaxWidth:   0,
				AutoWrap:   tw.WrapNormal,
				Alignment:  tw.AlignRight,
				AutoFormat: false,
				MergeMode:  tw.MergeNone,
			},
			Padding: tw.CellPadding{
				Global: defaultPadding,
			},
		},
		Debug:    true,
		AutoHide: false,
	}
}

// prepareContexts initializes rendering and merge contexts for the table.
func (t *Table) prepareContexts() (*renderContext, *mergeContext, error) {
	// Determine original number of columns FIRST
	numOriginalCols := t.maxColumns()
	t.debug("prepareContexts: Original number of columns: %d", numOriginalCols)

	// Initialize renderContext
	ctx := &renderContext{
		table:    t,
		renderer: t.renderer,
		cfg:      t.renderer.Config(),
		numCols:  numOriginalCols, // Store original count initially
		widths: map[tw.Position]tw.Mapper[int, int]{
			tw.Header: tw.NewMapper[int, int](),
			tw.Row:    tw.NewMapper[int, int](),
			tw.Footer: tw.NewMapper[int, int](),
		},
		debug: t.debug,
	}

	// Detect empty columns based on t.rows and t.config.AutoHide
	isEmpty, visibleCount := t.getEmptyColumnInfo(numOriginalCols)
	ctx.emptyColumns = isEmpty
	ctx.visibleColCount = visibleCount

	// Initialize merge context
	mctx := &mergeContext{
		headerMerges: make(map[int]tw.MergeState),
		rowMerges:    make([]map[int]tw.MergeState, len(t.rows)),
		footerMerges: make(map[int]tw.MergeState),
		horzMerges:   make(map[tw.Position]map[int]bool),
	}
	for i := range mctx.rowMerges {
		mctx.rowMerges[i] = make(map[int]tw.MergeState)
	}

	// Prepare content (needed for width calculation)
	// Assign internal data to context for processing
	ctx.headerLines = t.headers
	ctx.rowLines = t.rows
	ctx.footerLines = t.footers

	// Calculate initial widths based on ALL content (original columns)
	if err := t.calculateAndNormalizeWidths(ctx); err != nil {
		t.debug("Error during initial width calculation: %v", err)
		return nil, nil, err
	}
	t.debug("Initial normalized widths (before hiding): H=%v, R=%v, F=%v",
		ctx.widths[tw.Header], ctx.widths[tw.Row], ctx.widths[tw.Footer])

	// Prepare merges based on original content structure and column count
	preparedHeaderLines, headerMerges, _ := t.prepareWithMerges(ctx.headerLines, t.config.Header, tw.Header)
	ctx.headerLines = preparedHeaderLines // Update ctx with processed lines
	mctx.headerMerges = headerMerges

	processedRowLines := make([][][]string, len(ctx.rowLines))
	for i, row := range ctx.rowLines {
		if mctx.rowMerges[i] == nil {
			mctx.rowMerges[i] = make(map[int]tw.MergeState)
		}
		processedRowLines[i], mctx.rowMerges[i], _ = t.prepareWithMerges(row, t.config.Row, tw.Row)
	}
	ctx.rowLines = processedRowLines // Update ctx with processed lines

	// Apply H-Merge Widths based on calculated merges (modifies ctx.widths)
	t.applyHorizontalMergeWidths(tw.Header, ctx, mctx.headerMerges)

	// Apply V/H Merges (modifies mctx.rowMerges)
	if t.config.Row.Formatting.MergeMode&tw.MergeVertical != 0 {
		t.applyVerticalMerges(ctx, mctx)
	}
	if t.config.Row.Formatting.MergeMode&tw.MergeHierarchical != 0 {
		t.applyHierarchicalMerges(ctx, mctx)
	}

	// Prepare Footer (updates ctx.footerLines, mctx.footerMerges, ctx.widths[tw.Footer])
	t.prepareFooter(ctx, mctx)
	t.debug("Footer prepared. Widths before hiding: H=%v, R=%v, F=%v",
		ctx.widths[tw.Header], ctx.widths[tw.Row], ctx.widths[tw.Footer])

	// --- Step 3 Logic: Adjust widths for hidden columns ---
	if t.config.AutoHide {
		t.debug("Applying AutoHide: Adjusting widths for empty columns.")
		if ctx.emptyColumns == nil {
			t.debug("Warning: ctx.emptyColumns is nil during width adjustment.")
		} else if len(ctx.emptyColumns) != ctx.numCols {
			t.debug("Warning: Length mismatch between emptyColumns (%d) and numCols (%d). Skipping adjustment.", len(ctx.emptyColumns), ctx.numCols)
		} else {
			for colIdx := 0; colIdx < ctx.numCols; colIdx++ {
				// Check the flag determined purely by t.rows content
				if ctx.emptyColumns[colIdx] {
					// If the data rows designated this column as empty, zero out its width everywhere
					t.debug("AutoHide: Hiding column %d by setting width to 0.", colIdx)
					ctx.widths[tw.Header].Set(colIdx, 0)
					ctx.widths[tw.Row].Set(colIdx, 0) // Affects normalized widths used by renderers
					ctx.widths[tw.Footer].Set(colIdx, 0)
				}
			}
			t.debug("Widths after AutoHide adjustment: H=%v, R=%v, F=%v",
				ctx.widths[tw.Header], ctx.widths[tw.Row], ctx.widths[tw.Footer])
		}
	} else {
		t.debug("AutoHide is disabled, skipping width adjustment.")
	}
	// --- End Step 3 Logic ---

	t.debug("prepareContexts completed all stages.")
	return ctx, mctx, nil
}

// renderLine renders a single line with callbacks and normalized widths.
func (t *Table) renderLine(ctx *renderContext, mctx *mergeContext, hctx *helperContext, aligns map[int]tw.Align, padding map[int]tw.Padding) error {
	resp := t.buildCellContexts(ctx, mctx, hctx, aligns, padding)
	f := ctx.renderer

	isPaddingLine := false
	sectionConfig := t.config.Row
	switch hctx.position {
	case tw.Header:
		sectionConfig = t.config.Header
		isPaddingLine = (hctx.lineIdx == -1 && sectionConfig.Padding.Global.Top != tw.Empty) ||
			(hctx.lineIdx == len(ctx.headerLines) && sectionConfig.Padding.Global.Bottom != tw.Empty)
	case tw.Footer:
		sectionConfig = t.config.Footer
		isPaddingLine = (hctx.lineIdx == -1 && sectionConfig.Padding.Global.Top != tw.Empty) ||
			(hctx.lineIdx == len(ctx.footerLines) && (sectionConfig.Padding.Global.Bottom != tw.Empty || t.hasPerColumnBottomPadding()))
	case tw.Row:
		if hctx.rowIdx >= 0 && hctx.rowIdx < len(ctx.rowLines) {
			isPaddingLine = (hctx.lineIdx == -1 && sectionConfig.Padding.Global.Top != tw.Empty) ||
				(hctx.lineIdx == len(ctx.rowLines[hctx.rowIdx]) && sectionConfig.Padding.Global.Bottom != tw.Empty)
		}
	}

	sectionWidths := ctx.widths[hctx.position]
	normalizedWidths := ctx.widths[tw.Row]

	formatting := tw.Formatting{
		Row: tw.RowContext{
			Widths:       sectionWidths,
			ColMaxWidths: t.getColMaxWidths(hctx.position),
			Current:      resp.cells,
			Previous:     resp.prevCells,
			Next:         resp.nextCells,
			Position:     hctx.position,
			Location:     hctx.location,
		},
		Level:            t.getLevel(hctx.position),
		IsSubRow:         hctx.lineIdx > 0 || isPaddingLine,
		Debug:            t.config.Debug,
		NormalizedWidths: normalizedWidths,
	}

	if hctx.position == tw.Row {
		formatting.HasFooter = len(ctx.footerLines) > 0
	}

	switch hctx.position {
	case tw.Header:
		f.Header(t.writer, [][]string{hctx.line}, formatting)
	case tw.Row:
		f.Row(t.writer, hctx.line, formatting)
	case tw.Footer:
		f.Footer(t.writer, [][]string{hctx.line}, formatting)
	}
	return nil
}

// renderPadding renders padding lines.
func (t *Table) renderPadding(ctx *renderContext, mctx *mergeContext, hctx *helperContext, padChar string) error {
	ctx.debug("Rendering padding line for %s (using char like '%s')", hctx.position, padChar)

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

// appendSingle adds a single row to the table's row data.
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

// toStringLines converts a row to string lines.
// Handles []string, []any, []interface{}, or types compatible with a custom stringer.
func (t *Table) toStringLines(row interface{}, config tw.CellConfig) ([][]string, error) {
	t.debug("Converting row to string lines: %v (type: %T)", row, row)
	var cells []string

	switch v := row.(type) {
	case []string:
		cells = v
		t.debug("Row is already []string")
	case []any: // Handle []any
		t.debug("Row is []any, converting elements")
		cells = make([]string, len(v))
		for i, element := range v {
			cells[i] = t.elementToString(element)
		}
	default:
		// Try using the custom stringer if provided
		if t.stringer != nil {
			t.debug("Attempting conversion using custom stringer for type %T", row)
			rv := reflect.ValueOf(t.stringer)
			// Basic validation of the stringer function signature
			if rv.Kind() != reflect.Func || rv.Type().NumIn() != 1 || rv.Type().NumOut() != 1 {
				err := errors.Newf("stringer must be a func(T) []string, got %T", t.stringer)
				t.debug("Stringer format error: %v", err)
				return nil, err
			}
			inType := rv.Type().In(0)
			rowType := reflect.TypeOf(row)
			if !rowType.AssignableTo(inType) {
				err := errors.Newf("cannot assign row type %T to stringer input type %s", row, inType)
				t.debug("Stringer type mismatch error: %v", err)
				return nil, err
			}

			in := []reflect.Value{reflect.ValueOf(row)}
			out := rv.Call(in)

			// Basic validation of the stringer function output
			if len(out) != 1 || out[0].Kind() != reflect.Slice || out[0].Type().Elem().Kind() != reflect.String {
				err := errors.Newf("stringer must return []string, got %T", out[0].Interface())
				t.debug("Stringer return error: %v", err)
				return nil, err
			}
			cells = out[0].Interface().([]string)
			t.debug("Converted row using stringer: %v", cells)
		} else {
			// If no stringer and not a known slice type, report error
			err := errors.Newf("cannot convert row type %T to []string; provide a stringer via WithStringer", row)
			t.debug("Conversion error: %v", err)
			return nil, err
		}
	}

	// Apply global filter if present
	if config.Filter.Global != nil {
		t.debug("Applying global filter to cells: %v", cells)
		cells = config.Filter.Global(cells)
		t.debug("Cells after global filter: %v", cells)
	}

	// Apply per-column filters if present
	if len(config.Filter.PerColumn) > 0 {
		t.debug("Applying per-column filters to cells")
		numFilters := len(config.Filter.PerColumn)
		for i, cell := range cells {
			if i < numFilters && config.Filter.PerColumn[i] != nil {
				originalCell := cell
				cells[i] = config.Filter.PerColumn[i](cell)
				if cells[i] != originalCell {
					t.debug("  Col %d filter applied: '%s' -> '%s'", i, originalCell, cells[i])
				}
			}
		}
	}

	// Prepare content (wrapping, splitting multi-line cells)
	result := t.prepareContent(cells, config)
	t.debug("Prepared content lines: %v", result)
	return result, nil
}

// elementToString converts a single element to its string representation.
// It prioritizes the tw.Formatter interface if implemented.
func (t *Table) elementToString(element interface{}) string {
	if element == nil {
		return ""
	}

	// 1. Check for custom formatter (tw.Formatter)
	if formatter, ok := element.(tw.Formatter); ok {
		return formatter.Format()
	}

	// 2. Handle io.Reader (with fixed buffer to prevent OOM)
	if reader, ok := element.(io.Reader); ok {
		const maxReadSize = 512 // Prevent huge reads
		var buf strings.Builder
		_, err := io.CopyN(&buf, reader, maxReadSize)
		if err != nil && err != io.EOF {
			return fmt.Sprintf("[reader error: %v]", err)
		}
		if buf.Len() == maxReadSize {
			buf.WriteString("...") // Indicate truncation
		}
		return buf.String()
	}

	// 3. Handle SQL nullable types (sql.NullString, sql.NullInt64, etc.)
	switch v := element.(type) {
	case sql.NullString:
		if v.Valid {
			return v.String
		}
		return "" // NULL in SQL → empty string
	case sql.NullInt64:
		if v.Valid {
			return fmt.Sprintf("%d", v.Int64)
		}
		return ""
	case sql.NullFloat64:
		if v.Valid {
			return fmt.Sprintf("%f", v.Float64)
		}
		return ""
	case sql.NullBool:
		if v.Valid {
			return fmt.Sprintf("%t", v.Bool)
		}
		return ""
	case sql.NullTime:
		if v.Valid {
			return v.Time.String() // Or format as needed
		}
		return ""
	}

	// 4. Handle []byte (common in DB results)
	if b, ok := element.([]byte); ok {
		return string(b)
	}

	// 5. Handle errors (implements Error())
	if err, ok := element.(error); ok {
		return err.Error()
	}

	// 6. Handle fmt.Stringer (like time.Time, net.IP, etc.)
	if stringer, ok := element.(fmt.Stringer); ok {
		return stringer.String()
	}

	// 7. Fallback: Use %v, but avoid panics on weird types
	defer func() {
		if r := recover(); r != nil {
			return // Return empty string on format panic
		}
	}()
	return fmt.Sprintf("%v", element)
}

// prepareContent processes cell content with formatting, wrapping, and splitting.
func (t *Table) prepareContent(cells []string, config tw.CellConfig) [][]string {
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
					wrapped, _ := twwarp.WrapString(line, contentWidth)
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

// prepareFooter processes footer content, applying merges and updating widths.
func (t *Table) prepareFooter(ctx *renderContext, mctx *mergeContext) {
	if len(t.footers) == 0 {
		ctx.debug("Skipping footer preparation - no footer data")
		if ctx.widths[tw.Footer] == nil {
			ctx.widths[tw.Footer] = tw.NewMapper[int, int]()
		}
		numCols := ctx.numCols
		for i := 0; i < numCols; i++ {
			ctx.widths[tw.Footer].Set(i, ctx.widths[tw.Row].Get(i))
		}
		t.debug("Initialized empty footer widths based on row widths: %v", ctx.widths[tw.Footer])
		ctx.footerPrepared = true
		return
	}

	t.debug("Preparing footer with merge mode: %d", t.config.Footer.Formatting.MergeMode)
	preparedLines, mergeStates, _ := t.prepareWithMerges(t.footers, t.config.Footer, tw.Footer)
	t.footers = preparedLines
	mctx.footerMerges = mergeStates
	ctx.footerLines = t.footers
	t.debug("Base footer widths (normalized from rows/header): %v", ctx.widths[tw.Footer])
	t.applyHorizontalMergeWidths(tw.Footer, ctx, mctx.footerMerges)
	ctx.footerPrepared = true
	t.debug("Footer preparation completed. Final footer widths: %v", ctx.widths[tw.Footer])
}

// prepareWithMerges processes content and detects horizontal merges.
func (t *Table) prepareWithMerges(content [][]string, config tw.CellConfig, position tw.Position) ([][]string, map[int]tw.MergeState, map[int]bool) {
	t.debug("PrepareWithMerges START: position=%s, mergeMode=%d", position, config.Formatting.MergeMode)
	if len(content) == 0 {
		t.debug("PrepareWithMerges END: No content.")
		return content, nil, nil
	}

	numCols := 0
	if len(content) > 0 && len(content[0]) > 0 {
		numCols = len(content[0])
	} else {
		for _, line := range content {
			if len(line) > numCols {
				numCols = len(line)
			}
		}
		if numCols == 0 {
			numCols = t.maxColumns()
		}
	}

	if numCols == 0 {
		t.debug("PrepareWithMerges END: numCols is zero.")
		return content, nil, nil
	}

	horzMergeMap := make(map[int]bool)
	mergeMap := make(map[int]tw.MergeState)
	result := make([][]string, len(content))
	for i := range content {
		result[i] = padLine(content[i], numCols)
	}

	if config.Formatting.MergeMode&tw.MergeHorizontal != 0 {
		t.debug("Checking for horizontal merges in %d lines", len(content))

		if position == tw.Footer {
			for lineIdx := 0; lineIdx < len(content); lineIdx++ {
				originalLine := padLine(content[lineIdx], numCols)
				currentLineResult := result[lineIdx]

				firstContentIdx := -1
				var firstContent string
				for c := 0; c < numCols; c++ {
					if c >= len(originalLine) {
						break
					}
					trimmedVal := strings.TrimSpace(originalLine[c])
					if trimmedVal != "" && trimmedVal != "-" {
						firstContentIdx = c
						firstContent = originalLine[c]
						break
					} else if trimmedVal == "-" {
						break
					}
				}

				if firstContentIdx > 0 {
					span := firstContentIdx + 1
					startCol := 0

					allEmptyBefore := true
					for c := 0; c < firstContentIdx; c++ {
						if c >= len(originalLine) || strings.TrimSpace(originalLine[c]) != "" {
							allEmptyBefore = false
							break
						}
					}

					if allEmptyBefore {
						t.debug("Footer lead-merge applied line %d: content '%s' from col %d moved to col %d, span %d",
							lineIdx, firstContent, firstContentIdx, startCol, span)

						if startCol < len(currentLineResult) {
							currentLineResult[startCol] = firstContent
						}
						for k := startCol + 1; k < startCol+span; k++ {
							if k < len(currentLineResult) {
								currentLineResult[k] = tw.Empty
							}
						}

						startState := mergeMap[startCol]
						startState.Horizontal = tw.MergeStateOption{Present: true, Span: span, Start: true, End: span == 1}
						mergeMap[startCol] = startState
						horzMergeMap[startCol] = true

						for k := startCol + 1; k < startCol+span; k++ {
							colState := mergeMap[k]
							colState.Horizontal = tw.MergeStateOption{Present: true, Span: span, Start: false, End: k == startCol+span-1}
							mergeMap[k] = colState
							horzMergeMap[k] = true
						}
					}
				}
			}
		}

		for lineIdx := 0; lineIdx < len(content); lineIdx++ {
			originalLine := padLine(content[lineIdx], numCols)
			currentLineResult := result[lineIdx]
			col := 0
			for col < numCols {
				if horzMergeMap[col] {
					leadMergeHandled := false
					if mergeState, ok := mergeMap[col]; ok && mergeState.Horizontal.Present && !mergeState.Horizontal.Start {
						tempCol := col - 1
						startCol := -1
						startState := tw.MergeState{}
						for tempCol >= 0 {
							if state, okS := mergeMap[tempCol]; okS && state.Horizontal.Present && state.Horizontal.Start {
								startCol = tempCol
								startState = state
								break
							}
							tempCol--
						}
						if startCol != -1 {
							skipToCol := startCol + startState.Horizontal.Span
							if skipToCol > col {
								t.debug("Skipping standard H-merge check from col %d to %d (part of detected H-merge)", col, skipToCol-1)
								col = skipToCol
								leadMergeHandled = true
							}
						}
					}
					if leadMergeHandled {
						continue
					}
				}

				if col >= len(currentLineResult) {
					break
				}
				currentVal := strings.TrimSpace(currentLineResult[col])

				if currentVal == "" || currentVal == "-" || (mergeMap[col].Horizontal.Present && mergeMap[col].Horizontal.Start) {
					col++
					continue
				}

				span := 1
				startCol := col
				for nextCol := col + 1; nextCol < numCols; nextCol++ {
					if nextCol >= len(originalLine) {
						break
					}
					originalNextVal := strings.TrimSpace(originalLine[nextCol])

					if currentVal == originalNextVal && !horzMergeMap[nextCol] && originalNextVal != "-" {
						span++
					} else {
						break
					}
				}

				if span > 1 {
					t.debug("Standard horizontal merge at line %d, col %d, span %d", lineIdx, startCol, span)
					startState := mergeMap[startCol]
					startState.Horizontal = tw.MergeStateOption{Present: true, Span: span, Start: true, End: (span == 1)}
					mergeMap[startCol] = startState
					horzMergeMap[startCol] = true

					for k := startCol + 1; k < startCol+span; k++ {
						if k < len(currentLineResult) {
							currentLineResult[k] = tw.Empty
						}
						colState := mergeMap[k]
						colState.Horizontal = tw.MergeStateOption{Present: true, Span: span, Start: false, End: k == startCol+span-1}
						mergeMap[k] = colState
						horzMergeMap[k] = true
					}
					col += span
				} else {
					col++
				}
			}
		}
	}

	t.debug("PrepareWithMerges END: position=%s, lines=%d", position, len(result))
	return result, mergeMap, horzMergeMap
}

// applyVerticalMerges applies vertical merges to row content.
func (t *Table) applyVerticalMerges(ctx *renderContext, mctx *mergeContext) {
	ctx.debug("Applying vertical merges across %d rows", len(ctx.rowLines))
	numCols := ctx.numCols

	mergeStartRow := make(map[int]int)
	mergeStartContent := make(map[int]string)

	for i := 0; i < len(ctx.rowLines); i++ {
		if i >= len(mctx.rowMerges) {
			newRowMerges := make([]map[int]tw.MergeState, i+1)
			copy(newRowMerges, mctx.rowMerges)
			for k := len(mctx.rowMerges); k <= i; k++ {
				newRowMerges[k] = make(map[int]tw.MergeState)
			}
			mctx.rowMerges = newRowMerges
			ctx.debug("Extended rowMerges to index %d", i)
		} else if mctx.rowMerges[i] == nil {
			mctx.rowMerges[i] = make(map[int]tw.MergeState)
		}

		if len(ctx.rowLines[i]) == 0 {
			continue
		}
		currentLineContent := ctx.rowLines[i][0]
		paddedLine := padLine(currentLineContent, numCols)

		for col := 0; col < numCols; col++ {
			currentVal := strings.TrimSpace(paddedLine[col])
			startRow, ongoingMerge := mergeStartRow[col]
			startContent := mergeStartContent[col]
			mergeState := mctx.rowMerges[i][col]

			if ongoingMerge && currentVal == startContent && currentVal != "" {
				mergeState.Vertical = tw.MergeStateOption{
					Present: true,
					Span:    0,
					Start:   false,
					End:     false,
				}
				mctx.rowMerges[i][col] = mergeState
				for lineIdx := range ctx.rowLines[i] {
					if col < len(ctx.rowLines[i][lineIdx]) {
						ctx.rowLines[i][lineIdx][col] = tw.Empty
					}
				}
				ctx.debug("Vertical merge continued at row %d, col %d", i, col)
			} else {
				if ongoingMerge {
					endedRow := i - 1
					if endedRow >= 0 && endedRow >= startRow {
						startState := mctx.rowMerges[startRow][col]
						startState.Vertical.Span = (endedRow - startRow) + 1
						mctx.rowMerges[startRow][col] = startState

						endState := mctx.rowMerges[endedRow][col]
						endState.Vertical.End = true
						endState.Vertical.Span = startState.Vertical.Span
						mctx.rowMerges[endedRow][col] = endState
						ctx.debug("Vertical merge ended at row %d, col %d, span %d", endedRow, col, startState.Vertical.Span)
					}
					delete(mergeStartRow, col)
					delete(mergeStartContent, col)
				}

				if currentVal != "" {
					mergeState.Vertical = tw.MergeStateOption{
						Present: true,
						Span:    1,
						Start:   true,
						End:     false,
					}
					mctx.rowMerges[i][col] = mergeState
					mergeStartRow[col] = i
					mergeStartContent[col] = currentVal
					ctx.debug("Vertical merge started at row %d, col %d", i, col)
				} else if !mergeState.Horizontal.Present {
					mergeState.Vertical = tw.MergeStateOption{}
					mctx.rowMerges[i][col] = mergeState
				}
			}
		}
	}

	lastRowIdx := len(ctx.rowLines) - 1
	if lastRowIdx >= 0 {
		for col, startRow := range mergeStartRow {
			startState := mctx.rowMerges[startRow][col]
			finalSpan := (lastRowIdx - startRow) + 1
			startState.Vertical.Span = finalSpan
			mctx.rowMerges[startRow][col] = startState

			endState := mctx.rowMerges[lastRowIdx][col]
			endState.Vertical.Present = true
			endState.Vertical.End = true
			endState.Vertical.Span = finalSpan
			if startRow != lastRowIdx {
				endState.Vertical.Start = false
			}
			mctx.rowMerges[lastRowIdx][col] = endState
			ctx.debug("Vertical merge finalized at row %d, col %d, span %d", lastRowIdx, col, finalSpan)
		}
	}
	ctx.debug("Vertical merges completed")
}

// applyHorizontalMergeWidths recalculates widths for a section after H-merges are known.
func (t *Table) applyHorizontalMergeWidths(position tw.Position, ctx *renderContext, mergeStates map[int]tw.MergeState) {
	if mergeStates == nil {
		t.debug("applyHorizontalMergeWidths: Skipping %s - no merge states", position)
		return
	}
	t.debug("applyHorizontalMergeWidths: Applying HMerge width recalc for %s", position)

	numCols := ctx.numCols
	targetWidthsMap := ctx.widths[position]
	originalNormalizedWidths := tw.NewMapper[int, int]()
	for i := 0; i < numCols; i++ {
		originalNormalizedWidths.Set(i, targetWidthsMap.Get(i))
	}

	separatorWidth := 0
	if t.renderer != nil {
		rendererConfig := t.renderer.Config()
		if rendererConfig.Settings.Separators.BetweenColumns.Enabled() {
			separatorWidth = twfn.DisplayWidth(rendererConfig.Symbols.Column())
		}
	}

	processedCols := make(map[int]bool)

	for col := 0; col < numCols; col++ {
		if processedCols[col] {
			continue
		}

		state, exists := mergeStates[col]
		if !exists {
			continue
		}

		if state.Horizontal.Present && state.Horizontal.Start {
			totalWidth := 0
			span := state.Horizontal.Span
			t.debug("  -> HMerge detected: startCol=%d, span=%d, separatorWidth=%d", col, span, separatorWidth)

			for i := 0; i < span && (col+i) < numCols; i++ {
				currentColIndex := col + i
				normalizedWidth := originalNormalizedWidths.Get(currentColIndex)
				totalWidth += normalizedWidth
				t.debug("      -> col %d: adding normalized width %d", currentColIndex, normalizedWidth)

				if i > 0 && separatorWidth > 0 {
					totalWidth += separatorWidth
					t.debug("      -> col %d: adding separator width %d", currentColIndex, separatorWidth)
				}
			}

			targetWidthsMap.Set(col, totalWidth)
			t.debug("  -> Set %s col %d width to %d (merged)", position, col, totalWidth)
			processedCols[col] = true

			for i := 1; i < span && (col+i) < numCols; i++ {
				targetWidthsMap.Set(col+i, 0)
				t.debug("  -> Set %s col %d width to 0 (part of merge)", position, col+i)
				processedCols[col+i] = true
			}
		}
	}
	ctx.debug("applyHorizontalMergeWidths: Final widths for %s: %v", position, targetWidthsMap)
}

// applyHierarchicalMerges applies hierarchical merges to row content using a snapshot.
func (t *Table) applyHierarchicalMerges(ctx *renderContext, mctx *mergeContext) {
	ctx.debug("Applying hierarchical merges (left-to-right vertical flow - snapshot comparison)")
	if len(ctx.rowLines) <= 1 {
		ctx.debug("Skipping hierarchical merges - less than 2 rows")
		return
	}
	numCols := ctx.numCols

	originalRowLines := make([][][]string, len(ctx.rowLines))
	for i, row := range ctx.rowLines {
		originalRowLines[i] = make([][]string, len(row))
		for j, line := range row {
			originalRowLines[i][j] = make([]string, len(line))
			copy(originalRowLines[i][j], line)
		}
	}
	ctx.debug("Created snapshot of original row data for hierarchical merge comparison.")

	hMergeStartRow := make(map[int]int)

	for r := 1; r < len(ctx.rowLines); r++ {
		leftCellContinuedHierarchical := false

		for c := 0; c < numCols; c++ {
			if mctx.rowMerges[r] == nil {
				mctx.rowMerges[r] = make(map[int]tw.MergeState)
			}
			if mctx.rowMerges[r-1] == nil {
				mctx.rowMerges[r-1] = make(map[int]tw.MergeState)
			}

			canCompare := r > 0 &&
				len(originalRowLines[r]) > 0 && len(originalRowLines[r][0]) > c &&
				len(originalRowLines[r-1]) > 0 && len(originalRowLines[r-1][0]) > c

			if !canCompare {
				currentState := mctx.rowMerges[r][c]
				currentState.Hierarchical = tw.MergeStateOption{}
				mctx.rowMerges[r][c] = currentState
				ctx.debug("HCompare Skipped: r=%d, c=%d - Insufficient data in snapshot", r, c)
				leftCellContinuedHierarchical = false
				continue
			}

			currentVal := strings.TrimSpace(originalRowLines[r][0][c])
			aboveVal := strings.TrimSpace(originalRowLines[r-1][0][c])
			currentState := mctx.rowMerges[r][c]
			prevStateAbove := mctx.rowMerges[r-1][c]

			valuesMatch := (currentVal == aboveVal && currentVal != "" && currentVal != "-")
			hierarchyAllowed := (c == 0 || leftCellContinuedHierarchical)
			shouldContinue := valuesMatch && hierarchyAllowed

			ctx.debug("HCompare: r=%d, c=%d; current='%s', above='%s'; match=%v; leftCont=%v; shouldCont=%v",
				r, c, currentVal, aboveVal, valuesMatch, leftCellContinuedHierarchical, shouldContinue)

			if shouldContinue {
				currentState.Hierarchical.Present = true
				currentState.Hierarchical.Start = false

				if prevStateAbove.Hierarchical.Present && !prevStateAbove.Hierarchical.End {
					startRow, ok := hMergeStartRow[c]
					if !ok {
						ctx.debug("Hierarchical merge WARNING: Recovering lost start row at r=%d, c=%d. Assuming r-1 was start.", r, c)
						startRow = r - 1
						hMergeStartRow[c] = startRow
						startState := mctx.rowMerges[startRow][c]
						startState.Hierarchical.Present = true
						startState.Hierarchical.Start = true
						startState.Hierarchical.End = false
						mctx.rowMerges[startRow][c] = startState
					}
					ctx.debug("Hierarchical merge CONTINUED row %d, col %d. Block previously started row %d", r, c, startRow)
				} else {
					startRow := r - 1
					hMergeStartRow[c] = startRow
					startState := mctx.rowMerges[startRow][c]
					startState.Hierarchical.Present = true
					startState.Hierarchical.Start = true
					startState.Hierarchical.End = false
					mctx.rowMerges[startRow][c] = startState
					ctx.debug("Hierarchical merge START detected for block ending at or after row %d, col %d (started at row %d)", r, c, startRow)
				}

				for lineIdx := range ctx.rowLines[r] {
					if c < len(ctx.rowLines[r][lineIdx]) {
						ctx.rowLines[r][lineIdx][c] = tw.Empty
					}
				}

				leftCellContinuedHierarchical = true
			} else {
				currentState.Hierarchical = tw.MergeStateOption{}

				if startRow, ok := hMergeStartRow[c]; ok {
					t.finalizeHierarchicalMergeBlock(ctx, mctx, c, startRow, r-1)
					delete(hMergeStartRow, c)
				}

				leftCellContinuedHierarchical = false
			}

			mctx.rowMerges[r][c] = currentState
		}
	}

	lastRowIdx := len(ctx.rowLines) - 1
	if lastRowIdx >= 0 {
		for c, startRow := range hMergeStartRow {
			t.finalizeHierarchicalMergeBlock(ctx, mctx, c, startRow, lastRowIdx)
		}
	}
	ctx.debug("Hierarchical merge processing completed")
}

// finalizeHierarchicalMergeBlock sets the final Span and End flags for a completed H-merge block.
func (t *Table) finalizeHierarchicalMergeBlock(ctx *renderContext, mctx *mergeContext, col, startRow, endRow int) {
	if endRow < startRow {
		ctx.debug("Hierarchical merge FINALIZE WARNING: Invalid block col %d, start %d > end %d", col, startRow, endRow)
		return
	}
	if startRow < 0 || endRow < 0 {
		ctx.debug("Hierarchical merge FINALIZE WARNING: Negative row indices col %d, start %d, end %d", col, startRow, endRow)
		return
	}
	requiredLen := endRow + 1
	if requiredLen > len(mctx.rowMerges) {
		ctx.debug("Hierarchical merge FINALIZE WARNING: rowMerges slice too short (len %d) for endRow %d", len(mctx.rowMerges), endRow)
		return
	}
	if mctx.rowMerges[startRow] == nil {
		mctx.rowMerges[startRow] = make(map[int]tw.MergeState)
	}
	if mctx.rowMerges[endRow] == nil {
		mctx.rowMerges[endRow] = make(map[int]tw.MergeState)
	}

	finalSpan := (endRow - startRow) + 1
	ctx.debug("Finalizing H-merge block: col=%d, startRow=%d, endRow=%d, span=%d", col, startRow, endRow, finalSpan)

	startState := mctx.rowMerges[startRow][col]
	if startState.Hierarchical.Present && startState.Hierarchical.Start {
		startState.Hierarchical.Span = finalSpan
		startState.Hierarchical.End = finalSpan == 1
		mctx.rowMerges[startRow][col] = startState
		ctx.debug(" -> Updated start state: %+v", startState.Hierarchical)
	} else {
		ctx.debug("Hierarchical merge FINALIZE WARNING: col %d, startRow %d was not marked as Present/Start? Current state: %+v. Attempting recovery.", col, startRow, startState.Hierarchical)
		startState.Hierarchical.Present = true
		startState.Hierarchical.Start = true
		startState.Hierarchical.Span = finalSpan
		startState.Hierarchical.End = (finalSpan == 1)
		mctx.rowMerges[startRow][col] = startState
	}

	if endRow > startRow {
		endState := mctx.rowMerges[endRow][col]
		if endState.Hierarchical.Present && !endState.Hierarchical.Start {
			endState.Hierarchical.End = true
			endState.Hierarchical.Span = finalSpan
			mctx.rowMerges[endRow][col] = endState
			ctx.debug(" -> Updated end state: %+v", endState.Hierarchical)
		} else {
			ctx.debug("Hierarchical merge FINALIZE WARNING: col %d, endRow %d was not marked as Present/Continuation? Current state: %+v. Attempting recovery.", col, endRow, endState.Hierarchical)
			endState.Hierarchical.Present = true
			endState.Hierarchical.Start = false
			endState.Hierarchical.End = true
			endState.Hierarchical.Span = finalSpan
			mctx.rowMerges[endRow][col] = endState
		}
	} else {
		ctx.debug(" -> Span is 1, startRow is also endRow.")
	}
}

// calculateAndNormalizeWidths computes and normalizes column widths across sections.
func (t *Table) calculateAndNormalizeWidths(ctx *renderContext) error {
	ctx.debug("Calculating and normalizing widths")

	t.headerWidths = tw.NewMapper[int, int]()
	t.rowWidths = tw.NewMapper[int, int]()
	t.footerWidths = tw.NewMapper[int, int]()

	for _, lines := range ctx.headerLines {
		t.updateWidths(lines, t.headerWidths, t.config.Header.Padding)
	}
	ctx.debug("Initial Header widths: %v", t.headerWidths)
	for _, row := range ctx.rowLines {
		for _, line := range row {
			t.updateWidths(line, t.rowWidths, t.config.Row.Padding)
		}
	}
	ctx.debug("Initial Row widths: %v", t.rowWidths)
	for _, lines := range ctx.footerLines {
		t.updateWidths(lines, t.footerWidths, t.config.Footer.Padding)
	}
	ctx.debug("Initial Footer widths: %v", t.footerWidths)

	ctx.debug("Normalizing widths for %d columns", ctx.numCols)
	ctx.widths[tw.Header] = tw.NewMapper[int, int]()
	ctx.widths[tw.Row] = tw.NewMapper[int, int]()
	ctx.widths[tw.Footer] = tw.NewMapper[int, int]()

	for i := 0; i < ctx.numCols; i++ {
		maxWidth := 0
		for _, w := range []tw.Mapper[int, int]{t.headerWidths, t.rowWidths, t.footerWidths} {
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

// maxColumns calculates the maximum number of columns across all sections.
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
	t.debug("Max columns: %d", m)
	return m
}

// buildAligns constructs a map of column alignments based on config.
func (t *Table) buildAligns(config tw.CellConfig) map[int]tw.Align {
	t.debug("Building aligns for section. Section Default: '%s', ColumnAligns: %v", config.Formatting.Alignment, config.ColumnAligns)
	colAlignsResult := make(map[int]tw.Align)
	numCols := t.maxColumns()

	for i := 0; i < numCols; i++ {
		// Default to the section's configured alignment
		currentAlign := config.Formatting.Alignment
		// t.debug("  Col %d: Initial align from section default: '%s'", i, currentAlign)

		// Check for a per-column override
		if i < len(config.ColumnAligns) {
			colSpecificAlign := config.ColumnAligns[i]
			// t.debug("  Col %d: Found ColumnAligns value: '%s'", i, colSpecificAlign)
			if colSpecificAlign == tw.Skip { // If explicitly "skip", signal renderer to use its default
				currentAlign = tw.AlignNone
				// t.debug("  Col %d: ColumnAligns is tw.Skip, setting to tw.AlignNone for renderer default.", i)
			} else if colSpecificAlign != "" { // If any other non-empty specific alignment, use it
				currentAlign = colSpecificAlign
				// t.debug("  Col %d: ColumnAligns provides specific override: '%s'", i, currentAlign)
			}
			// If colSpecificAlign is "" but not tw.Skip (e.g. an uninitialized entry), it won't override,
			// and we stick with section default. However, tw.Skip is also "", so this case is covered by the above.
			// If colSpecificAlign is tw.AlignNone ("none"), that's a valid specific override.
		}
		colAlignsResult[i] = currentAlign
	}
	t.debug("Aligns built: %v", colAlignsResult)
	return colAlignsResult
}

// buildPadding constructs a map of column padding settings based on config.
func (t *Table) buildPadding(padding tw.CellPadding) map[int]tw.Padding {
	t.debug("Building padding")
	colPadding := make(map[int]tw.Padding)
	numCols := t.maxColumns()
	for i := 0; i < numCols; i++ {
		if i < len(padding.PerColumn) && padding.PerColumn[i] != (tw.Padding{}) {
			colPadding[i] = padding.PerColumn[i]
			t.debug("Col %d: Using per-column padding: %+v", i, padding.PerColumn[i])
		} else {
			colPadding[i] = padding.Global
			t.debug("Col %d: Using global padding: %+v", i, padding.Global)
		}
	}
	t.debug("Padding built: %v", colPadding)
	return colPadding
}

// debug logs a message to the trace if debug mode is enabled.
func (t *Table) debug(format string, a ...interface{}) {
	if t.config.Debug {
		msg := fmt.Sprintf(format, a...)
		t.trace = append(t.trace, fmt.Sprintf("[TABLE] %s", msg))
	}
}

// ensureInitialized ensures all required fields are initialized before use.
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
		t.renderer = renderer.NewBlueprint()
	}
	t.debug("ensureInitialized called")
}

// getColMaxWidths retrieves the maximum widths for columns in a section.
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

// getLevel maps a position to its corresponding rendering level.
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

// determineLocation determines the boundary location for a line in headers or footers.
func (t *Table) determineLocation(lineIdx, totalLines int, topPad, bottomPad string) tw.Location {
	if lineIdx == 0 && topPad == tw.Empty {
		return tw.LocationFirst
	}
	if lineIdx == totalLines-1 && bottomPad == tw.Empty {
		return tw.LocationEnd
	}
	return tw.LocationMiddle
}

// updateWidths updates the width map based on cell content and padding.
func (t *Table) updateWidths(row []string, widths tw.Mapper[int, int], padding tw.CellPadding) {
	t.debug("Updating widths for row: %v", row)
	for i, cell := range row {
		// Determine effective padding for this column
		colPad := padding.Global // Start with global
		if i < len(padding.PerColumn) && padding.PerColumn[i] != (tw.Padding{}) {
			// Use per-column padding ONLY if it's explicitly set (not the zero value)
			colPad = padding.PerColumn[i]
			t.debug("  Col %d: Using per-column padding: L:'%s' R:'%s'", i, colPad.Left, colPad.Right)
		} else {
			t.debug("  Col %d: Using global padding: L:'%s' R:'%s'", i, padding.Global.Left, padding.Global.Right)
		}

		// Calculate display widths OF THE ACTUAL PADDING CHARACTERS specified
		// DisplayWidth("") correctly returns 0.
		padLeftWidth := twfn.DisplayWidth(colPad.Left)
		padRightWidth := twfn.DisplayWidth(colPad.Right)

		// Calculate content width - assume trimming for width calculation consistency
		contentWidth := twfn.DisplayWidth(strings.TrimSpace(cell))

		// Total width required for this cell including padding space
		totalWidth := contentWidth + padLeftWidth + padRightWidth

		// Ensure minimum width IS the calculated padding width, even if content is empty.
		// The minimum space required is just the space for the padding characters themselves.
		minRequiredPaddingWidth := padLeftWidth + padRightWidth
		if contentWidth == 0 && totalWidth < minRequiredPaddingWidth {
			// If content is empty, the width must be AT LEAST the padding width.
			t.debug("  Col %d: Empty content, ensuring width >= padding width (%d). Setting totalWidth to %d.", i, minRequiredPaddingWidth, minRequiredPaddingWidth)
			totalWidth = minRequiredPaddingWidth
		}

		// A cell should visually occupy at least 1 unit if it exists, even if content
		// and padding are empty/zero-width. This prevents columns from collapsing entirely.
		if totalWidth < 1 {
			t.debug("  Col %d: Calculated totalWidth is zero, setting minimum width to 1.", i)
			totalWidth = 1
		}

		// Update the map if this cell requires more width than previously recorded
		currentMax, _ := widths.OK(i) // Get current max width for this column
		if totalWidth > currentMax {
			widths.Set(i, totalWidth)
			t.debug("  Col %d: Updated width from %d to %d (content:%d + padL:%d + padR:%d) for cell '%s'", i, currentMax, totalWidth, contentWidth, padLeftWidth, padRightWidth, cell)
		} else {
			t.debug("  Col %d: Width %d not greater than current max %d for cell '%s'", i, totalWidth, currentMax, cell)
		}
	}
}

// hasPerColumnBottomPadding checks if any per-column bottom padding is defined.
func (t *Table) hasPerColumnBottomPadding() bool {
	if t.config.Footer.Padding.PerColumn == nil {
		return false
	}
	for _, pad := range t.config.Footer.Padding.PerColumn {
		if pad.Bottom != tw.Empty {
			return true
		}
	}
	return false
}

// hasFooterElements checks if footer has any renderable elements (content or padding).
func (t *Table) hasFooterElements() bool {
	hasContent := len(t.footers) > 0
	hasTopPadding := t.config.Footer.Padding.Global.Top != tw.Empty
	hasBottomPaddingConfig := t.config.Footer.Padding.Global.Bottom != tw.Empty || t.hasPerColumnBottomPadding()
	return hasContent || hasTopPadding || hasBottomPaddingConfig
}

// getEmptyColumnInfo checks data rows (t.rows) to determine which columns
// contain only empty or whitespace content. It ignores headers and footers.
// It returns a boolean slice where true indicates an empty column,
// and the count of non-empty (visible) columns.
func (t *Table) getEmptyColumnInfo(numOriginalCols int) (isEmpty []bool, visibleColCount int) {
	// Initialize tracker: assume all columns are empty initially.
	isEmpty = make([]bool, numOriginalCols)
	for i := range isEmpty {
		isEmpty[i] = true
	}

	if !t.config.AutoHide {
		// If feature is disabled, consider all columns non-empty
		t.debug("getEmptyColumnInfo: AutoHide disabled, marking all %d columns as visible.", numOriginalCols)
		for i := range isEmpty {
			isEmpty[i] = false
		}
		visibleColCount = numOriginalCols
		return isEmpty, visibleColCount
	}

	t.debug("getEmptyColumnInfo: Checking %d rows for %d columns...", len(t.rows), numOriginalCols)

	// Iterate through actual row data (t.rows)
	// Structure: [logical_row_idx][visual_line_idx][cell_idx]
	for rowIdx, logicalRow := range t.rows {
		for lineIdx, visualLine := range logicalRow {
			// Process each cell in the visual line
			for colIdx, cellContent := range visualLine {
				// Important: Only check columns within the original bounds
				if colIdx >= numOriginalCols {
					continue // Should not happen if data is consistent, but safe check
				}

				// If a column is already marked as not empty, skip further checks for it
				if !isEmpty[colIdx] {
					continue
				}

				// Check if the cell content is non-whitespace
				if strings.TrimSpace(cellContent) != "" {
					// Found content, mark this column as not empty
					isEmpty[colIdx] = false
					t.debug("getEmptyColumnInfo: Found content in row %d, line %d, col %d ('%s'). Marked as not empty.", rowIdx, lineIdx, colIdx, cellContent)
				}
			}
		}
	}

	// Count visible columns
	visibleColCount = 0
	for _, empty := range isEmpty {
		if !empty {
			visibleColCount++
		}
	}

	t.debug("getEmptyColumnInfo: Detection complete. isEmpty: %v, visibleColCount: %d", isEmpty, visibleColCount)
	return isEmpty, visibleColCount
}

// padLine pads a line to the specified number of columns with empty strings.
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
