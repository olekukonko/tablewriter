package tablewriter

import (
	"bytes"
	"fmt"
	"github.com/olekukonko/errors"
	"github.com/olekukonko/ll"
	"github.com/olekukonko/ll/lh"
	"github.com/olekukonko/tablewriter/renderer"
	"github.com/olekukonko/tablewriter/tw"
	"github.com/olekukonko/tablewriter/twfn"
	"io"
	"reflect"
	"strings"
	"sync"
)

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
	logger       *ll.Logger          // Debug trace log
	trace        *bytes.Buffer       // Debug trace log

	// streaming
	streamWidths            tw.Mapper[int, int]           // Fixed column widths for streaming mode, calculated once
	streamFooterLines       [][]string                    // Processed footer lines for streaming, stored until Close().
	headerRendered          bool                          // Tracks if header has been rendered in streaming mode
	firstRowRendered        bool                          // Tracks if the first data row has been rendered in streaming mode
	lastRenderedLineContent []string                      // Content of the very last line rendered (for Previous context in streaming)
	lastRenderedMergeState  tw.Mapper[int, tw.MergeState] // Merge state of the very last line rendered (for Previous context in streaming)
	lastRenderedPosition    tw.Position                   // Position (Header/Row/Footer/Separator) of the last line rendered (for Previous context in streaming)
	streamNumCols           int                           // The derived number of columns in streaming mode
	streamRowCounter        int                           // Counter for rows rendered in streaming mode (0-indexed logical rows)

	// cache
	stringerCache        map[reflect.Type]reflect.Value // Cache for stringer reflection
	stringerCacheMu      sync.RWMutex                   // Mutex for thread-safe cache access
	stringerCacheEnabled bool                           // Flag to enable/disable caching
}

// renderContext holds the core state for rendering the table.
type renderContext struct {
	table       *Table                              // Reference to the table instance
	renderer    tw.Renderer                         // Renderer instance
	cfg         tw.RendererConfig                   // Renderer configuration
	numCols     int                                 // Total number of columns
	headerLines [][]string                          // Processed header lines
	rowLines    [][][]string                        // Processed row lines
	footerLines [][]string                          // Processed footer lines
	widths      map[tw.Position]tw.Mapper[int, int] // Widths per section
	//debug           func(format string, a ...interface{}) // Debug logging function
	footerPrepared  bool       // Tracks if footer is prepared
	emptyColumns    []bool     // Tracks which original columns are empty (true if empty)
	visibleColCount int        // Count of columns that are NOT empty
	logger          *ll.Logger // Debug trace log
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
	cells        map[int]tw.CellContext // Current line cells
	prevCells    map[int]tw.CellContext // Previous line cells
	nextCells    map[int]tw.CellContext // Next line cells
	location     tw.Location            // Determined Location for this line
	cellsContent []string
}

// NewTable creates a new table instance with specified writer and options.
// Parameters include writer for output and optional configuration options.
// Returns a pointer to the initialized Table instance.
func NewTable(w io.Writer, opts ...Option) *Table {
	t := &Table{
		writer:       w,
		headerWidths: tw.NewMapper[int, int](),
		rowWidths:    tw.NewMapper[int, int](),
		footerWidths: tw.NewMapper[int, int](),
		renderer:     renderer.NewBlueprint(),
		config:       defaultConfig(),
		newLine:      tw.NewLine,
		trace:        &bytes.Buffer{},

		// --- ADDED INITIALIZATION START ---
		streamWidths:           tw.NewMapper[int, int](), // Initialize empty mapper for streaming widths
		lastRenderedMergeState: tw.NewMapper[int, tw.MergeState](),
		headerRendered:         false,
		firstRowRendered:       false,
		lastRenderedPosition:   "",
		streamNumCols:          0,
		streamRowCounter:       0,

		//  Cache
		stringerCache:        make(map[reflect.Type]reflect.Value),
		stringerCacheEnabled: false, // Disabled by default
	}

	// add logger
	t.logger = ll.New("table").Handler(lh.NewTextHandler(t.trace))

	for _, opt := range opts {
		opt(t)
	}

	// send logger to renderer
	// this will overwrite the default logger
	t.renderer.Logger(t.logger)
	t.logger.Info("Table initialized with %d options", len(opts))
	return t
}

// ---- Public Methods ----

// Append adds rows to the table, supporting various input types.
// Parameter rows accepts one or more rows, with stringer for custom types.
// Returns an error if any row fails to append.
func (t *Table) Append(rows ...interface{}) error {
	t.ensureInitialized() // Ensure initialized regardless of mode

	// --- ADDED STREAMING LOGIC START ---
	// Check if streaming is enabled and started
	if t.config.Stream.Enable && t.hasPrinted {
		t.logger.Debug("Append() called in streaming mode with %d rows", len(rows))
		for i, row := range rows {
			// Delegate rendering of each row to the streaming helper
			if err := t.streamAppendRow(row); err != nil {
				t.logger.Error("Error rendering streaming row %d: %v", i, err)
				// Decide error handling: stop appending? log and continue?
				// For now, stop and return the error.
				return fmt.Errorf("failed to stream append row %d: %w", i, err)
			}
		}
		t.logger.Debug("Append() completed in streaming mode, %d rows processed.", len(rows))
		return nil // Exit the function after handling streaming
	}
	// --- ADDED STREAMING LOGIC END ---

	// Existing batch rendering logic:
	t.logger.Debug("Starting batch Append operation with %d rows", len(rows))
	for i, row := range rows {
		if err := t.appendSingle(row); err != nil {
			t.logger.Debug("Append failed at index %d (batch mode): %v", i, err)
			return err
		}
	}
	t.logger.Debug("Append completed (batch mode), total rows: %d", len(t.rows))
	return nil
}

// Bulk adds multiple rows from a slice to the table (legacy method).
// Parameter rows must be a slice compatible with stringer or []string.
// Returns an error if the input is invalid or appending fails.
func (t *Table) Bulk(rows interface{}) error {
	t.logger.Debug("Starting Bulk operation")
	rv := reflect.ValueOf(rows)
	if rv.Kind() != reflect.Slice {
		err := errors.Newf("Bulk expects a slice, got %T", rows)
		t.logger.Debug("Bulk error: %v", err)
		return err
	}
	for i := 0; i < rv.Len(); i++ {
		row := rv.Index(i).Interface()
		t.logger.Debug("Processing bulk row %d: %v", i, row)
		if err := t.appendSingle(row); err != nil {
			t.logger.Debug("Bulk append failed at index %d: %v", i, err)
			return err
		}
	}
	t.logger.Debug("Bulk completed, processed %d rows", rv.Len())
	return nil
}

// Config returns the current table configuration.
// No parameters are required.
// Returns the Config struct with current settings.
func (t *Table) Config() Config {
	return t.config
}

// Configure updates the table's configuration using a provided function.
// Parameter fn is a function that modifies the Config struct.
// Returns the Table instance for method chaining.
func (t *Table) Configure(fn func(*Config)) *Table {
	fn(&t.config)
	return t
}

// Debug retrieves the accumulated debug trace logs.
// No parameters are required.
// Returns a slice of debug messages including renderer logs.
func (t *Table) Debug() *bytes.Buffer {
	return t.trace
}

// Footer sets the table's footer content, padding to match column count.
// Parameter footers is a slice of strings for footer content.
// No return value.
// Footer sets the table's footer content.
// Parameter footers is a slice of strings for footer content.
// In streaming mode, this processes and stores the footer for rendering by Close().
func (t *Table) Footer(elements ...any) {
	t.ensureInitialized()
	t.logger.Debug("Footer() method called with raw variadic elements: %v (len %d). Streaming: %v, Started: %v",
		elements, len(elements), t.config.Stream.Enable, t.hasPrinted)

	if t.config.Stream.Enable && t.hasPrinted {
		// --- Streaming Path ---
		actualCellsToProcess := t.processVariadicElements(elements)
		footersAsStrings, err := t.rawCellsToStrings(actualCellsToProcess, t.config.Footer)
		if err != nil {
			t.logger.Error("Footer(): Failed to convert footer elements to strings for streaming: %v", err)
			footersAsStrings = []string{} // Use empty on error
		}
		errStream := t.streamStoreFooter(footersAsStrings) // streamStoreFooter handles padding to streamNumCols internally
		if errStream != nil {
			t.logger.Error("Error processing streaming footer: %v", errStream)
		}
		return
	}

	// --- Batch Path ---
	actualCellsToProcess := t.processVariadicElements(elements)
	t.logger.Debug("Footer() (Batch): Effective cells to process: %v", actualCellsToProcess)

	footersAsStrings, err := t.rawCellsToStrings(actualCellsToProcess, t.config.Footer)
	if err != nil {
		t.logger.Error("Footer() (Batch): Failed to convert to strings: %v", err)
		t.footers = [][]string{} // Set to empty on error
		return
	}

	preparedFooterLines := t.prepareContent(footersAsStrings, t.config.Footer)
	t.footers = preparedFooterLines // Store directly. Padding to t.maxColumns() will happen in prepareContexts.

	t.logger.Debug("Footer set (batch mode), lines stored: %d. First line if exists: %v",
		len(t.footers), func() []string {
			if len(t.footers) > 0 {
				return t.footers[0]
			} else {
				return nil
			}
		}())
}

func (t *Table) Header(elements ...any) {
	t.ensureInitialized()
	t.logger.Debug("Header() method called with raw variadic elements: %v (len %d). Streaming: %v, Started: %v",
		elements, len(elements), t.config.Stream.Enable, t.hasPrinted)

	if t.config.Stream.Enable && t.hasPrinted {
		// --- Streaming Path ---
		actualCellsToProcess := t.processVariadicElements(elements)
		headersAsStrings, err := t.rawCellsToStrings(actualCellsToProcess, t.config.Header)
		if err != nil {
			t.logger.Error("Header(): Failed to convert header elements to strings for streaming: %v", err)
			headersAsStrings = []string{} // Use empty on error
		}
		errStream := t.streamRenderHeader(headersAsStrings) // streamRenderHeader handles padding to streamNumCols internally
		if errStream != nil {
			t.logger.Error("Error rendering streaming header: %v", errStream)
		}
		return
	}

	// --- Batch Path ---
	actualCellsToProcess := t.processVariadicElements(elements)
	t.logger.Debug("Header() (Batch): Effective cells to process: %v", actualCellsToProcess)

	headersAsStrings, err := t.rawCellsToStrings(actualCellsToProcess, t.config.Header)
	if err != nil {
		t.logger.Error("Header() (Batch): Failed to convert to strings: %v", err)
		t.headers = [][]string{} // Set to empty on error
		return
	}

	// prepareContent uses t.config.Header for AutoFormat and MaxWidth constraints.
	// It processes based on the number of columns in headersAsStrings.
	preparedHeaderLines := t.prepareContent(headersAsStrings, t.config.Header)
	t.headers = preparedHeaderLines // Store directly. Padding to t.maxColumns() will happen in prepareContexts.

	t.logger.Debug("Header set (batch mode), lines stored: %d. First line if exists: %v",
		len(t.headers), func() []string {
			if len(t.headers) > 0 {
				return t.headers[0]
			} else {
				return nil
			}
		}())
}

func (t *Table) Logger() *ll.Logger {
	return t.logger
}

// Render triggers the table rendering process to the configured writer.
// No parameters are required.
// Returns an error if rendering fails.
func (t *Table) Render() error {
	return t.render()
}

// Renderer retrieves the current renderer instance used by the table.
// No parameters are required.
// Returns the tw.Renderer interface instance.
func (t *Table) Renderer() tw.Renderer {
	t.logger.Debug("Renderer requested")
	return t.renderer
}

// ---- Private Methods ----

// appendSingle adds a single row to the table's row data.
// Parameter row is the data to append, converted via stringer if needed.
// Returns an error if conversion or appending fails.
func (t *Table) appendSingle(row interface{}) error {
	t.ensureInitialized() // Already here

	if t.config.Stream.Enable && t.hasPrinted { // If streaming is active
		t.logger.Debug("appendSingle: Dispatching to streamAppendRow for row: %v", row)
		return t.streamAppendRow(row) // Call the streaming render function
	}
	// Existing batch logic:
	t.logger.Debug("appendSingle: Processing for batch mode, row: %v", row)
	// toStringLines now uses the new rawCellsToStrings internally, then prepareContent.
	// This is fine for batch.
	lines, err := t.toStringLines(row, t.config.Row)
	if err != nil {
		t.logger.Debug("Error in toStringLines (batch mode): %v", err)
		return err
	}
	t.rows = append(t.rows, lines) // Add to batch storage
	t.logger.Debug("Row appended to batch t.rows, total batch rows: %d", len(t.rows))
	return nil
}

// applyHierarchicalMerges applies hierarchical merges to row content.
// Parameters ctx and mctx hold rendering and merge state.
// No return value.
func (t *Table) applyHierarchicalMerges(ctx *renderContext, mctx *mergeContext) {
	ctx.logger.Debug("Applying hierarchical merges (left-to-right vertical flow - snapshot comparison)")
	if len(ctx.rowLines) <= 1 {
		ctx.logger.Debug("Skipping hierarchical merges - less than 2 rows")
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
	ctx.logger.Debug("Created snapshot of original row data for hierarchical merge comparison.")

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
				ctx.logger.Debug("HCompare Skipped: r=%d, c=%d - Insufficient data in snapshot", r, c)
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

			ctx.logger.Debug("HCompare: r=%d, c=%d; current='%s', above='%s'; match=%v; leftCont=%v; shouldCont=%v",
				r, c, currentVal, aboveVal, valuesMatch, leftCellContinuedHierarchical, shouldContinue)

			if shouldContinue {
				currentState.Hierarchical.Present = true
				currentState.Hierarchical.Start = false

				if prevStateAbove.Hierarchical.Present && !prevStateAbove.Hierarchical.End {
					startRow, ok := hMergeStartRow[c]
					if !ok {
						ctx.logger.Debug("Hierarchical merge WARNING: Recovering lost start row at r=%d, c=%d. Assuming r-1 was start.", r, c)
						startRow = r - 1
						hMergeStartRow[c] = startRow
						startState := mctx.rowMerges[startRow][c]
						startState.Hierarchical.Present = true
						startState.Hierarchical.Start = true
						startState.Hierarchical.End = false
						mctx.rowMerges[startRow][c] = startState
					}
					ctx.logger.Debug("Hierarchical merge CONTINUED row %d, col %d. Block previously started row %d", r, c, startRow)
				} else {
					startRow := r - 1
					hMergeStartRow[c] = startRow
					startState := mctx.rowMerges[startRow][c]
					startState.Hierarchical.Present = true
					startState.Hierarchical.Start = true
					startState.Hierarchical.End = false
					mctx.rowMerges[startRow][c] = startState
					ctx.logger.Debug("Hierarchical merge START detected for block ending at or after row %d, col %d (started at row %d)", r, c, startRow)
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
	ctx.logger.Debug("Hierarchical merge processing completed")
}

// applyHorizontalMergeWidths adjusts column widths for horizontal merges.
// Parameters include position, ctx for rendering, and mergeStates for merges.
// No return value.
func (t *Table) applyHorizontalMergeWidths(position tw.Position, ctx *renderContext, mergeStates map[int]tw.MergeState) {
	if mergeStates == nil {
		t.logger.Debug("applyHorizontalMergeWidths: Skipping %s - no merge states", position)
		return
	}
	t.logger.Debug("applyHorizontalMergeWidths: Applying HMerge width recalc for %s", position)

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
			t.logger.Debug("  -> HMerge detected: startCol=%d, span=%d, separatorWidth=%d", col, span, separatorWidth)

			for i := 0; i < span && (col+i) < numCols; i++ {
				currentColIndex := col + i
				normalizedWidth := originalNormalizedWidths.Get(currentColIndex)
				totalWidth += normalizedWidth
				t.logger.Debug("      -> col %d: adding normalized width %d", currentColIndex, normalizedWidth)

				if i > 0 && separatorWidth > 0 {
					totalWidth += separatorWidth
					t.logger.Debug("      -> col %d: adding separator width %d", currentColIndex, separatorWidth)
				}
			}

			targetWidthsMap.Set(col, totalWidth)
			t.logger.Debug("  -> Set %s col %d width to %d (merged)", position, col, totalWidth)
			processedCols[col] = true

			for i := 1; i < span && (col+i) < numCols; i++ {
				targetWidthsMap.Set(col+i, 0)
				t.logger.Debug("  -> Set %s col %d width to 0 (part of merge)", position, col+i)
				processedCols[col+i] = true
			}
		}
	}
	ctx.logger.Debug("applyHorizontalMergeWidths: Final widths for %s: %v", position, targetWidthsMap)
}

// applyVerticalMerges applies vertical merges to row content.
// Parameters ctx and mctx hold rendering and merge state.
// No return value.
func (t *Table) applyVerticalMerges(ctx *renderContext, mctx *mergeContext) {
	ctx.logger.Debug("Applying vertical merges across %d rows", len(ctx.rowLines))
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
			ctx.logger.Debug("Extended rowMerges to index %d", i)
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
				ctx.logger.Debug("Vertical merge continued at row %d, col %d", i, col)
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
						ctx.logger.Debug("Vertical merge ended at row %d, col %d, span %d", endedRow, col, startState.Vertical.Span)
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
					ctx.logger.Debug("Vertical merge started at row %d, col %d", i, col)
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
			ctx.logger.Debug("Vertical merge finalized at row %d, col %d, span %d", lastRowIdx, col, finalSpan)
		}
	}
	ctx.logger.Debug("Vertical merges completed")
}

// buildAdjacentCells constructs cell contexts for adjacent lines.
// Parameters include ctx, mctx, hctx, and direction (-1 for prev, +1 for next).
// Returns a map of column indices to CellContext for the adjacent line.
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
			}
		} else { // Next
			if targetLineIdx < len(ctx.headerLines) {
				adjLine = ctx.headerLines[targetLineIdx]
				adjMerges = mctx.headerMerges
				found = true
			} else if len(ctx.rowLines) > 0 && len(ctx.rowLines[0]) > 0 && len(mctx.rowMerges) > 0 {
				adjLine = ctx.rowLines[0][0]
				adjMerges = mctx.rowMerges[0]
				adjPosition = tw.Row
				found = true
			} else if len(ctx.footerLines) > 0 {
				adjLine = ctx.footerLines[0]
				adjMerges = mctx.footerMerges
				adjPosition = tw.Footer
				found = true
			}
		}
	case tw.Row:
		targetLineIdx := hctx.lineIdx + direction
		if hctx.rowIdx < 0 || hctx.rowIdx >= len(ctx.rowLines) || hctx.rowIdx >= len(mctx.rowMerges) {
			t.logger.Debug("Warning: Invalid row index %d in buildAdjacentCells", hctx.rowIdx)
			return nil
		}
		currentRowLines := ctx.rowLines[hctx.rowIdx]
		currentMerges := mctx.rowMerges[hctx.rowIdx]

		if direction < 0 { // Previous
			if targetLineIdx >= 0 && targetLineIdx < len(currentRowLines) {
				adjLine = currentRowLines[targetLineIdx]
				adjMerges = currentMerges
				found = true
			} else if targetLineIdx < 0 {
				targetRowIdx := hctx.rowIdx - 1
				if targetRowIdx >= 0 && targetRowIdx < len(ctx.rowLines) && targetRowIdx < len(mctx.rowMerges) {
					prevRowLines := ctx.rowLines[targetRowIdx]
					if len(prevRowLines) > 0 {
						adjLine = prevRowLines[len(prevRowLines)-1]
						adjMerges = mctx.rowMerges[targetRowIdx]
						found = true
					}
				} else if len(ctx.headerLines) > 0 {
					adjLine = ctx.headerLines[len(ctx.headerLines)-1]
					adjMerges = mctx.headerMerges
					adjPosition = tw.Header
					found = true
				}
			}
		} else { // Next
			if targetLineIdx >= 0 && targetLineIdx < len(currentRowLines) {
				adjLine = currentRowLines[targetLineIdx]
				adjMerges = currentMerges
				found = true
			} else if targetLineIdx >= len(currentRowLines) {
				targetRowIdx := hctx.rowIdx + 1
				if targetRowIdx < len(ctx.rowLines) && targetRowIdx < len(mctx.rowMerges) && len(ctx.rowLines[targetRowIdx]) > 0 {
					adjLine = ctx.rowLines[targetRowIdx][0]
					adjMerges = mctx.rowMerges[targetRowIdx]
					found = true
				} else if len(ctx.footerLines) > 0 {
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
			} else if targetLineIdx < 0 {
				if len(ctx.rowLines) > 0 {
					lastRowIdx := len(ctx.rowLines) - 1
					if lastRowIdx < len(mctx.rowMerges) && len(ctx.rowLines[lastRowIdx]) > 0 {
						lastRowLines := ctx.rowLines[lastRowIdx]
						adjLine = lastRowLines[len(lastRowLines)-1]
						adjMerges = mctx.rowMerges[lastRowIdx]
						adjPosition = tw.Row
						found = true
					}
				} else if len(ctx.headerLines) > 0 {
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
			}
		}
	}

	if !found {
		return nil
	}

	if adjMerges == nil {
		adjMerges = make(map[int]tw.MergeState)
		t.logger.Debug("Warning: adjMerges was nil in buildAdjacentCells despite found=true")
	}

	paddedAdjLine := padLine(adjLine, ctx.numCols)

	for j := 0; j < ctx.numCols; j++ {
		mergeState := adjMerges[j]
		cellData := paddedAdjLine[j]
		finalAdjColWidth := ctx.widths[adjPosition].Get(j)

		adjCells[j] = tw.CellContext{
			Data:  cellData,
			Merge: mergeState,
			Width: finalAdjColWidth,
		}
	}
	return adjCells
}

// buildAligns constructs a map of column alignments from configuration.
// Parameter config provides alignment settings for the section.
// Returns a map of column indices to alignment settings.
func (t *Table) buildAligns(config tw.CellConfig) map[int]tw.Align {
	numColsToUse := t.getNumColsToUse()
	colAlignsResult := make(map[int]tw.Align)
	for i := 0; i < numColsToUse; i++ {
		currentAlign := config.Formatting.Alignment
		if i < len(config.ColumnAligns) {
			colSpecificAlign := config.ColumnAligns[i]
			if colSpecificAlign != tw.Empty && colSpecificAlign != tw.Skip {
				currentAlign = colSpecificAlign
			}
		}
		colAlignsResult[i] = currentAlign
	}
	t.logger.Debug("Aligns built: %v (length %d)", colAlignsResult, len(colAlignsResult))
	return colAlignsResult
}

// buildCellContexts creates CellContext objects for a given line.
// Parameters include ctx, mctx, hctx, aligns, and padding for rendering.
// Returns a renderMergeResponse with current, previous, and next cell contexts.
func (t *Table) buildCellContexts(ctx *renderContext, mctx *mergeContext, hctx *helperContext, aligns map[int]tw.Align, padding map[int]tw.Padding) renderMergeResponse {
	cells := make(map[int]tw.CellContext)
	var merges map[int]tw.MergeState

	switch hctx.position {
	case tw.Header:
		merges = mctx.headerMerges
	case tw.Row:
		if hctx.rowIdx >= 0 && hctx.rowIdx < len(mctx.rowMerges) && mctx.rowMerges[hctx.rowIdx] != nil {
			merges = mctx.rowMerges[hctx.rowIdx]
		} else {
			merges = make(map[int]tw.MergeState)
			t.logger.Debug("Warning: Invalid row index %d or nil merges in buildCellContexts", hctx.rowIdx)
		}
	case tw.Footer:
		merges = mctx.footerMerges
	default:
		merges = make(map[int]tw.MergeState)
		t.logger.Debug("Warning: Invalid position '%s' in buildCellContexts", hctx.position)
	}

	if merges == nil {
		merges = make(map[int]tw.MergeState)
		t.logger.Debug("Warning: merges map was nil in buildCellContexts after switch, using empty map")
	}

	for j := 0; j < ctx.numCols; j++ {
		mergeState := merges[j]
		cellData := ""
		if j < len(hctx.line) {
			cellData = hctx.line[j]
		}
		finalColWidth := ctx.widths[hctx.position].Get(j)

		cells[j] = tw.CellContext{
			Data:    cellData,
			Align:   aligns[j],
			Padding: padding[j],
			Width:   finalColWidth,
			Merge:   mergeState,
		}
	}

	prevCells := t.buildAdjacentCells(ctx, mctx, hctx, -1)
	nextCells := t.buildAdjacentCells(ctx, mctx, hctx, +1)

	return renderMergeResponse{
		cells:     cells,
		prevCells: prevCells,
		nextCells: nextCells,
	}
}

// buildPadding constructs a map of column padding settings from configuration.
// Parameter padding provides padding settings for the section.
// Returns a map of column indices to padding settings.
func (t *Table) buildPadding(padding tw.CellPadding) map[int]tw.Padding {
	numColsToUse := t.getNumColsToUse()
	colPadding := make(map[int]tw.Padding)
	for i := 0; i < numColsToUse; i++ {
		if i < len(padding.PerColumn) && padding.PerColumn[i] != (tw.Padding{}) {
			colPadding[i] = padding.PerColumn[i]
		} else {
			colPadding[i] = padding.Global
		}
	}
	t.logger.Debug("Padding built: %v (length %d)", colPadding, len(colPadding))
	return colPadding
}

// calculateAndNormalizeWidths computes and normalizes column widths.
// Parameter ctx holds rendering state with width maps.
// Returns an error if width calculation fails.
func (t *Table) calculateAndNormalizeWidths(ctx *renderContext) error {
	ctx.logger.Debug("Calculating and normalizing widths")

	t.headerWidths = tw.NewMapper[int, int]()
	t.rowWidths = tw.NewMapper[int, int]()
	t.footerWidths = tw.NewMapper[int, int]()

	for _, lines := range ctx.headerLines {
		t.updateWidths(lines, t.headerWidths, t.config.Header.Padding)
	}
	ctx.logger.Debug("Initial Header widths: %v", t.headerWidths)
	for _, row := range ctx.rowLines {
		for _, line := range row {
			t.updateWidths(line, t.rowWidths, t.config.Row.Padding)
		}
	}
	ctx.logger.Debug("Initial Row widths: %v", t.rowWidths)
	for _, lines := range ctx.footerLines {
		t.updateWidths(lines, t.footerWidths, t.config.Footer.Padding)
	}
	ctx.logger.Debug("Initial Footer widths: %v", t.footerWidths)

	ctx.logger.Debug("Normalizing widths for %d columns", ctx.numCols)
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
	ctx.logger.Debug("Normalized widths: header=%v, row=%v, footer=%v", ctx.widths[tw.Header], ctx.widths[tw.Row], ctx.widths[tw.Footer])
	return nil
}

// defaultConfig provides the default configuration for a table.
// No parameters are required.
// Returns a Config struct with default settings.

// determineLocation determines the boundary location for a line.
// Parameters include lineIdx, totalLines, topPad, and bottomPad.
// Returns a tw.Location indicating First, Middle, or End.
func (t *Table) determineLocation(lineIdx, totalLines int, topPad, bottomPad string) tw.Location {
	if lineIdx == 0 && topPad == tw.Empty {
		return tw.LocationFirst
	}
	if lineIdx == totalLines-1 && bottomPad == tw.Empty {
		return tw.LocationEnd
	}
	return tw.LocationMiddle
}

// elementToString converts an element to its string representation.
// Parameter element is the data to convert, supporting various types.
// Returns the string representation of the element.
//func (t *Table) elementToString(element interface{}) string {
//	if element == nil {
//		return ""
//	}
//
//	if formatter, ok := element.(tw.Formatter); ok {
//		return formatter.Format()
//	}
//
//	if reader, ok := element.(io.Reader); ok {
//		const maxReadSize = 512
//		var buf strings.Builder
//		_, err := io.CopyN(&buf, reader, maxReadSize)
//		if err != nil && err != io.EOF {
//			return fmt.Sprintf("[reader error: %v]", err)
//		}
//		if buf.Len() == maxReadSize {
//			buf.WriteString("...")
//		}
//		return buf.String()
//	}
//
//	switch v := element.(type) {
//	case sql.NullString:
//		if v.Valid {
//			return v.String
//		}
//		return ""
//	case sql.NullInt64:
//		if v.Valid {
//			return fmt.Sprintf("%d", v.Int64)
//		}
//		return ""
//	case sql.NullFloat64:
//		if v.Valid {
//			return fmt.Sprintf("%f", v.Float64)
//		}
//		return ""
//	case sql.NullBool:
//		if v.Valid {
//			return fmt.Sprintf("%t", v.Bool)
//		}
//		return ""
//	case sql.NullTime:
//		if v.Valid {
//			return v.Time.String()
//		}
//		return ""
//	}
//
//	if b, ok := element.([]byte); ok {
//		return string(b)
//	}
//
//	if err, ok := element.(error); ok {
//		return err.Error()
//	}
//
//	if stringer, ok := element.(fmt.Stringer); ok {
//		return stringer.String()
//	}
//
//	defer func() {
//		if r := recover(); r != nil {
//			return
//		}
//	}()
//	return fmt.Sprintf("%v", element)
//}

// ensureInitialized initializes required fields before use.
// No parameters are required.
// No return value.
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
	t.logger.Debug("ensureInitialized called")
}

// finalizeHierarchicalMergeBlock sets Span and End for hierarchical merges.
// Parameters include ctx, mctx, col, startRow, and endRow.
// No return value.
func (t *Table) finalizeHierarchicalMergeBlock(ctx *renderContext, mctx *mergeContext, col, startRow, endRow int) {
	if endRow < startRow {
		ctx.logger.Debug("Hierarchical merge FINALIZE WARNING: Invalid block col %d, start %d > end %d", col, startRow, endRow)
		return
	}
	if startRow < 0 || endRow < 0 {
		ctx.logger.Debug("Hierarchical merge FINALIZE WARNING: Negative row indices col %d, start %d, end %d", col, startRow, endRow)
		return
	}
	requiredLen := endRow + 1
	if requiredLen > len(mctx.rowMerges) {
		ctx.logger.Debug("Hierarchical merge FINALIZE WARNING: rowMerges slice too short (len %d) for endRow %d", len(mctx.rowMerges), endRow)
		return
	}
	if mctx.rowMerges[startRow] == nil {
		mctx.rowMerges[startRow] = make(map[int]tw.MergeState)
	}
	if mctx.rowMerges[endRow] == nil {
		mctx.rowMerges[endRow] = make(map[int]tw.MergeState)
	}

	finalSpan := (endRow - startRow) + 1
	ctx.logger.Debug("Finalizing H-merge block: col=%d, startRow=%d, endRow=%d, span=%d", col, startRow, endRow, finalSpan)

	startState := mctx.rowMerges[startRow][col]
	if startState.Hierarchical.Present && startState.Hierarchical.Start {
		startState.Hierarchical.Span = finalSpan
		startState.Hierarchical.End = finalSpan == 1
		mctx.rowMerges[startRow][col] = startState
		ctx.logger.Debug(" -> Updated start state: %+v", startState.Hierarchical)
	} else {
		ctx.logger.Debug("Hierarchical merge FINALIZE WARNING: col %d, startRow %d was not marked as Present/Start? Current state: %+v. Attempting recovery.", col, startRow, startState.Hierarchical)
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
			ctx.logger.Debug(" -> Updated end state: %+v", endState.Hierarchical)
		} else {
			ctx.logger.Debug("Hierarchical merge FINALIZE WARNING: col %d, endRow %d was not marked as Present/Continuation? Current state: %+v. Attempting recovery.", col, endRow, endState.Hierarchical)
			endState.Hierarchical.Present = true
			endState.Hierarchical.Start = false
			endState.Hierarchical.End = true
			endState.Hierarchical.Span = finalSpan
			mctx.rowMerges[endRow][col] = endState
		}
	} else {
		ctx.logger.Debug(" -> Span is 1, startRow is also endRow.")
	}
}

// getColMaxWidths retrieves maximum column widths for a section.
// Parameter position specifies the section (Header, Row, Footer).
// Returns a map of column indices to maximum widths.
func (t *Table) getColMaxWidths(position tw.Position) tw.CellWidth {
	switch position {
	case tw.Header:
		return t.config.Header.ColMaxWidths
	case tw.Row:
		return t.config.Row.ColMaxWidths
	case tw.Footer:
		return t.config.Footer.ColMaxWidths
	default:
		return tw.CellWidth{}
	}
}

func (t *Table) getNumColsToUse() int {
	if t.config.Stream.Enable && t.hasPrinted {
		t.logger.Debug("getNumColsToUse: Using streamNumCols: %d", t.streamNumCols)
		return t.streamNumCols
	}
	numCols := t.maxColumns()
	t.logger.Debug("getNumColsToUse: Using maxColumns: %d", numCols)
	return numCols
}

// getEmptyColumnInfo identifies empty columns in row data.
// Parameter numOriginalCols specifies the total column count.
// Returns a boolean slice (true for empty) and visible column count.
func (t *Table) getEmptyColumnInfo(numOriginalCols int) (isEmpty []bool, visibleColCount int) {
	isEmpty = make([]bool, numOriginalCols)
	for i := range isEmpty {
		isEmpty[i] = true
	}

	if !t.config.AutoHide {
		t.logger.Debug("getEmptyColumnInfo: AutoHide disabled, marking all %d columns as visible.", numOriginalCols)
		for i := range isEmpty {
			isEmpty[i] = false
		}
		visibleColCount = numOriginalCols
		return isEmpty, visibleColCount
	}

	t.logger.Debug("getEmptyColumnInfo: Checking %d rows for %d columns...", len(t.rows), numOriginalCols)

	for rowIdx, logicalRow := range t.rows {
		for lineIdx, visualLine := range logicalRow {
			for colIdx, cellContent := range visualLine {
				if colIdx >= numOriginalCols {
					continue
				}
				if !isEmpty[colIdx] {
					continue
				}
				if strings.TrimSpace(cellContent) != "" {
					isEmpty[colIdx] = false
					t.logger.Debug("getEmptyColumnInfo: Found content in row %d, line %d, col %d ('%s'). Marked as not empty.", rowIdx, lineIdx, colIdx, cellContent)
				}
			}
		}
	}

	visibleColCount = 0
	for _, empty := range isEmpty {
		if !empty {
			visibleColCount++
		}
	}

	t.logger.Debug("getEmptyColumnInfo: Detection complete. isEmpty: %v, visibleColCount: %d", isEmpty, visibleColCount)
	return isEmpty, visibleColCount
}

// getLevel maps a position to its rendering level.
// Parameter position specifies the section (Header, Row, Footer).
// Returns the corresponding tw.Level (Header, Body, Footer).
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

// hasFooterElements checks if the footer has renderable elements.
// No parameters are required.
// Returns true if footer has content or padding, false otherwise.
func (t *Table) hasFooterElements() bool {
	hasContent := len(t.footers) > 0
	hasTopPadding := t.config.Footer.Padding.Global.Top != tw.Empty
	hasBottomPaddingConfig := t.config.Footer.Padding.Global.Bottom != tw.Empty || t.hasPerColumnBottomPadding()
	return hasContent || hasTopPadding || hasBottomPaddingConfig
}

// hasPerColumnBottomPadding checks for per-column bottom padding in footer.
// No parameters are required.
// Returns true if any per-column bottom padding is defined.
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

// maxColumns calculates the maximum column count across sections.
// No parameters are required.
// Returns the highest number of columns found.
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
	t.logger.Debug("Max columns: %d", m)
	return m
}

// render generates the table output using the configured renderer.
// No parameters are required.
// Returns an error if rendering fails in any section.
func (t *Table) render() error {
	t.ensureInitialized() // This is needed in both modes, good place to keep it.

	// --- ADDED LOGIC START ---
	// If streaming is enabled, this batch rendering path should not be used.
	// The Start/Append/Close methods handle streaming.
	if t.config.Stream.Enable {
		t.logger.Warn("Internal render() method called when streaming is enabled. This indicates incorrect usage of Render() instead of Start/Append/Close.")
		// It's safer to just return an error here, as the table state is likely not set up for batch.
		return errors.New("internal batch render called in streaming mode")
	}
	// --- ADDED LOGIC END ---

	// The rest of the original batch rendering logic follows:
	ctx, mctx, err := t.prepareContexts() // This is batch-specific
	if err != nil {
		t.logger.Error("prepareContexts failed: %v", err)
		return err
	}

	ctx.logger.Debug("Calling renderer Start()")
	// In batch mode, renderer.Start is called here.
	if err := ctx.renderer.Start(t.writer); err != nil {
		ctx.logger.Debug("Renderer Start() error: %v", err)
		return fmt.Errorf("renderer start failed: %w", err)
	}

	renderError := false
	// Render sections (Header, Row, Footer) sequentially using batch context
	for _, renderFn := range []func(*renderContext, *mergeContext) error{
		t.renderHeader,
		t.renderRow,
		t.renderFooter,
	} {
		if err := renderFn(ctx, mctx); err != nil {
			ctx.logger.Error("Renderer section error: %v", err)
			renderError = true
		}
	}

	ctx.logger.Debug("Calling renderer Close()")
	// In batch mode, renderer.Close is called here.
	closeErr := ctx.renderer.Close(t.writer)
	if closeErr != nil {
		ctx.logger.Error("Renderer Close() error: %v", closeErr)
		if !renderError {
			return fmt.Errorf("renderer close failed: %w", closeErr)
		}
	}

	if renderError {
		return errors.New("table rendering failed in one or more sections")
	}

	// This flag needs careful consideration with streaming.
	// For now, it marks that rendering finished (batch mode).
	t.hasPrinted = true
	ctx.logger.Debug("Render completed")
	return nil
}

// renderFooter renders the table's footer section with borders and padding.
// Parameters ctx and mctx hold rendering and merge state.
// Returns an error if rendering fails.
func (t *Table) renderFooter(ctx *renderContext, mctx *mergeContext) error {
	if !ctx.footerPrepared {
		t.prepareFooter(ctx, mctx)
	}

	f := ctx.renderer
	cfg := ctx.cfg

	hasContent := len(ctx.footerLines) > 0
	hasTopPadding := t.config.Footer.Padding.Global.Top != tw.Empty
	hasBottomPaddingConfig := t.config.Footer.Padding.Global.Bottom != tw.Empty || t.hasPerColumnBottomPadding()
	hasAnyFooterElement := hasContent || hasTopPadding || hasBottomPaddingConfig

	if !hasAnyFooterElement {
		hasContentAbove := len(ctx.rowLines) > 0 || len(ctx.headerLines) > 0
		if hasContentAbove && cfg.Borders.Bottom.Enabled() && cfg.Settings.Lines.ShowBottom.Enabled() {
			ctx.logger.Debug("Footer is empty, rendering table bottom border based on last row/header")
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
			ctx.logger.Debug("Bottom border: Using Widths=%v", ctx.widths[tw.Row])
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
			ctx.logger.Debug("Footer is empty and no content above or borders disabled, skipping footer render")
		}
		return nil
	}

	ctx.logger.Debug("Rendering footer section (has elements)")
	hasContentAbove := len(ctx.rowLines) > 0 || len(ctx.headerLines) > 0
	colAligns := t.buildAligns(t.config.Footer)
	colPadding := t.buildPadding(t.config.Footer.Padding)
	hctx := &helperContext{position: tw.Footer}
	// Declare paddingLineContentForContext with a default value
	paddingLineContentForContext := make([]string, ctx.numCols)

	if hasContentAbove && cfg.Settings.Lines.ShowFooterLine.Enabled() && !hasTopPadding && len(ctx.footerLines) > 0 {
		ctx.logger.Debug("Rendering footer separator line")
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
		ctx.logger.Debug("Footer separator: Using Widths=%v", ctx.widths[tw.Row])
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
		hctx.rowIdx = 0
		hctx.lineIdx = -1
		if !(hasContentAbove && cfg.Settings.Lines.ShowFooterLine.Enabled()) {
			hctx.location = tw.LocationFirst
		} else {
			hctx.location = tw.LocationMiddle
		}
		hctx.line = t.buildPaddingLineContents(t.config.Footer.Padding.Global.Top, ctx.widths[tw.Footer], ctx.numCols, mctx.footerMerges)
		ctx.logger.Debug("Calling renderPadding for Footer Top Padding line: %v (loc: %v)", hctx.line, hctx.location)
		if err := t.renderPadding(ctx, mctx, hctx, t.config.Footer.Padding.Global.Top); err != nil {
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
		ctx.logger.Debug("Rendering footer content line %d with location %v", i, hctx.location)
		if err := t.renderLine(ctx, mctx, hctx, colAligns, colPadding); err != nil {
			return err
		}
		lastRenderedLineIdx = i
	}

	if hasBottomPaddingConfig {
		paddingLineContentForContext = make([]string, ctx.numCols)
		formattedPaddingCells := make([]string, ctx.numCols)
		var representativePadChar string = " "
		ctx.logger.Debug("Constructing Footer Bottom Padding line content strings")
		for j := 0; j < ctx.numCols; j++ {
			colWd := ctx.widths[tw.Footer].Get(j)
			mergeState := tw.MergeState{}
			if mctx.footerMerges != nil {
				if state, ok := mctx.footerMerges[j]; ok {
					mergeState = state
				}
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
		ctx.logger.Debug("Manually rendering Footer Bottom Padding line (char like '%s')", representativePadChar)
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
		ctx.logger.Debug("Manually rendered Footer Bottom Padding line: %s", strings.TrimSuffix(paddingLineOutput.String(), t.newLine))
		hctx.rowIdx = 0
		hctx.lineIdx = len(ctx.footerLines)
		hctx.line = paddingLineContentForContext
		hctx.location = tw.LocationEnd
		lastRenderedLineIdx = hctx.lineIdx
	}

	if cfg.Borders.Bottom.Enabled() && cfg.Settings.Lines.ShowBottom.Enabled() {
		ctx.logger.Debug("Rendering final table bottom border")
		if lastRenderedLineIdx == len(ctx.footerLines) {
			hctx.rowIdx = 0
			hctx.lineIdx = lastRenderedLineIdx
			hctx.line = paddingLineContentForContext
			hctx.location = tw.LocationEnd
			ctx.logger.Debug("Setting border context based on bottom padding line")
		} else if lastRenderedLineIdx >= 0 {
			hctx.rowIdx = 0
			hctx.lineIdx = lastRenderedLineIdx
			hctx.line = padLine(ctx.footerLines[hctx.lineIdx], ctx.numCols)
			hctx.location = tw.LocationEnd
			ctx.logger.Debug("Setting border context based on last content line idx %d", hctx.lineIdx)
		} else if lastRenderedLineIdx == -1 {
			hctx.rowIdx = 0
			hctx.lineIdx = -1
			hctx.line = paddingLineContentForContext
			hctx.location = tw.LocationEnd
			ctx.logger.Debug("Setting border context based on top padding line")
		} else {
			hctx.rowIdx = 0
			hctx.lineIdx = -2
			hctx.line = make([]string, ctx.numCols)
			hctx.location = tw.LocationEnd
			ctx.logger.Debug("Warning: Cannot determine context for bottom border")
		}
		resp := t.buildCellContexts(ctx, mctx, hctx, colAligns, colPadding)
		ctx.logger.Debug("Bottom border: Using Widths=%v", ctx.widths[tw.Row])
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
	}

	return nil
}

// renderHeader renders the table's header section with borders and padding.
// Parameters ctx and mctx hold rendering and merge state.
// Returns an error if rendering fails.
func (t *Table) renderHeader(ctx *renderContext, mctx *mergeContext) error {
	if len(ctx.headerLines) == 0 {
		return nil
	}
	ctx.logger.Debug("Rendering header section")

	f := ctx.renderer
	cfg := ctx.cfg
	colAligns := t.buildAligns(t.config.Header)
	colPadding := t.buildPadding(t.config.Header.Padding)
	hctx := &helperContext{position: tw.Header}

	if cfg.Borders.Top.Enabled() && cfg.Settings.Lines.ShowTop.Enabled() {
		ctx.logger.Debug("Rendering table top border")
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

	if t.config.Header.Padding.Global.Top != tw.Empty {
		hctx.location = tw.LocationFirst
		hctx.line = t.buildPaddingLineContents(t.config.Header.Padding.Global.Top, ctx.widths[tw.Header], ctx.numCols, mctx.headerMerges)
		if err := t.renderPadding(ctx, mctx, hctx, t.config.Header.Padding.Global.Top); err != nil {
			return err
		}
	}

	for i, line := range ctx.headerLines {
		hctx.rowIdx = 0
		hctx.lineIdx = i
		hctx.line = padLine(line, ctx.numCols)
		hctx.location = t.determineLocation(i, len(ctx.headerLines), t.config.Header.Padding.Global.Top, t.config.Header.Padding.Global.Bottom)

		if t.config.Header.Callbacks.Global != nil {
			ctx.logger.Debug("Executing global header callback for line %d", i)
			t.config.Header.Callbacks.Global()
		}
		for colIdx, cb := range t.config.Header.Callbacks.PerColumn {
			if colIdx < ctx.numCols && cb != nil {
				ctx.logger.Debug("Executing per-column header callback for line %d, col %d", i, colIdx)
				cb()
			}
		}

		if err := t.renderLine(ctx, mctx, hctx, colAligns, colPadding); err != nil {
			return err
		}
	}

	if t.config.Header.Padding.Global.Bottom != tw.Empty {
		hctx.location = tw.LocationEnd
		hctx.line = t.buildPaddingLineContents(t.config.Header.Padding.Global.Bottom, ctx.widths[tw.Header], ctx.numCols, mctx.headerMerges)
		if err := t.renderPadding(ctx, mctx, hctx, t.config.Header.Padding.Global.Bottom); err != nil {
			return err
		}
	}

	if cfg.Settings.Lines.ShowHeaderLine.Enabled() && (len(ctx.rowLines) > 0 || len(ctx.footerLines) > 0) {
		ctx.logger.Debug("Rendering header separator line")
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

// renderLine renders a single line with callbacks and normalized widths.
// Parameters include ctx, mctx, hctx, aligns, and padding for rendering.
// Returns an error if rendering fails.
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

// renderPadding renders padding lines for a section.
// Parameters include ctx, mctx, hctx, and padChar for padding content.
// Returns an error if rendering fails.
func (t *Table) renderPadding(ctx *renderContext, mctx *mergeContext, hctx *helperContext, padChar string) error {
	ctx.logger.Debug("Rendering padding line for %s (using char like '%s')", hctx.position, padChar)

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

// renderRow renders the table's row section with borders and padding.
// Parameters ctx and mctx hold rendering and merge state.
// Returns an error if rendering fails.
func (t *Table) renderRow(ctx *renderContext, mctx *mergeContext) error {
	if len(ctx.rowLines) == 0 {
		return nil
	}
	ctx.logger.Debug("Rendering row section (total rows: %d)", len(ctx.rowLines))

	f := ctx.renderer
	cfg := ctx.cfg
	colAligns := t.buildAligns(t.config.Row)
	colPadding := t.buildPadding(t.config.Row.Padding)
	hctx := &helperContext{position: tw.Row}

	footerIsEmptyOrNonExistent := !t.hasFooterElements()
	if len(ctx.headerLines) == 0 && footerIsEmptyOrNonExistent && cfg.Borders.Top.Enabled() && cfg.Settings.Lines.ShowTop.Enabled() {
		ctx.logger.Debug("Rendering table top border (rows only table)")
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

	for i, lines := range ctx.rowLines {
		rowHasTopPadding := t.config.Row.Padding.Global.Top != tw.Empty
		if rowHasTopPadding {
			hctx.rowIdx = i
			hctx.lineIdx = -1
			if i == 0 && len(ctx.headerLines) == 0 {
				hctx.location = tw.LocationFirst
			} else {
				hctx.location = tw.LocationMiddle
			}
			hctx.line = t.buildPaddingLineContents(t.config.Row.Padding.Global.Top, ctx.widths[tw.Row], ctx.numCols, mctx.rowMerges[i])
			ctx.logger.Debug("Calling renderPadding for Row Top Padding (row %d): %v (loc: %v)", i, hctx.line, hctx.location)
			if err := t.renderPadding(ctx, mctx, hctx, t.config.Row.Padding.Global.Top); err != nil {
				return err
			}
		}

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

			ctx.logger.Debug("Rendering row %d line %d with location %v", i, j, hctx.location)
			if err := t.renderLine(ctx, mctx, hctx, colAligns, colPadding); err != nil {
				return err
			}

			shouldDrawSeparator := cfg.Settings.Separators.BetweenRows.Enabled() &&
				!(isLastRow && isLastLineOfRow && (footerExists || rowHasBottomPadding || (cfg.Borders.Bottom.Enabled() && cfg.Settings.Lines.ShowBottom.Enabled())))

			if shouldDrawSeparator {
				ctx.logger.Debug("Rendering between-rows separator after row %d line %d", i, j)
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
						ctx.logger.Debug("Separator context: Next is row %d line %d", nextRowIdx, nextLineIdx)
					} else if nextLineIdx == 0 && len(ctx.rowLines[nextRowIdx]) == 0 {
						ctx.logger.Debug("Separator context: Next row %d is empty", nextRowIdx)
						nextCells = nil
					} else {
						ctx.logger.Debug("Separator context: Unexpected end of lines for next row %d", nextRowIdx)
						nextCells = nil
					}
				} else {
					ctx.logger.Debug("Separator context: No next row.")
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
				ctx.logger.Debug("Skipping between-rows separator after last row %d line %d (footerExists=%v, rowHasBottomPadding=%v, bottomBorderEnabled=%v)",
					i, j, footerExists, rowHasBottomPadding, cfg.Borders.Bottom.Enabled() && cfg.Settings.Lines.ShowBottom.Enabled())
			}
		}

		if rowHasBottomPadding {
			hctx.rowIdx = i
			hctx.lineIdx = len(lines)
			if i == len(ctx.rowLines)-1 && !footerExists {
				hctx.location = tw.LocationEnd
			} else {
				hctx.location = tw.LocationMiddle
			}
			hctx.line = t.buildPaddingLineContents(t.config.Row.Padding.Global.Bottom, ctx.widths[tw.Row], ctx.numCols, mctx.rowMerges[i])
			ctx.logger.Debug("Calling renderPadding for Row Bottom Padding (row %d): %v (loc: %v)", i, hctx.line, hctx.location)
			if err := t.renderPadding(ctx, mctx, hctx, t.config.Row.Padding.Global.Bottom); err != nil {
				return err
			}
		}
	}
	return nil
}

// updateWidths updates the width map based on cell content and padding.
// Parameters include row content, widths map, and padding configuration.
// No return value.
func (t *Table) updateWidths(row []string, widths tw.Mapper[int, int], padding tw.CellPadding) {
	t.logger.Debug("Updating widths for row: %v", row)
	for i, cell := range row {
		colPad := padding.Global
		if i < len(padding.PerColumn) && padding.PerColumn[i] != (tw.Padding{}) {
			colPad = padding.PerColumn[i]
			t.logger.Debug("  Col %d: Using per-column padding: L:'%s' R:'%s'", i, colPad.Left, colPad.Right)
		} else {
			t.logger.Debug("  Col %d: Using global padding: L:'%s' R:'%s'", i, padding.Global.Left, padding.Global.Right)
		}

		padLeftWidth := twfn.DisplayWidth(colPad.Left)
		padRightWidth := twfn.DisplayWidth(colPad.Right)
		contentWidth := twfn.DisplayWidth(strings.TrimSpace(cell))
		totalWidth := contentWidth + padLeftWidth + padRightWidth
		minRequiredPaddingWidth := padLeftWidth + padRightWidth

		if contentWidth == 0 && totalWidth < minRequiredPaddingWidth {
			t.logger.Debug("  Col %d: Empty content, ensuring width >= padding width (%d). Setting totalWidth to %d.", i, minRequiredPaddingWidth, minRequiredPaddingWidth)
			totalWidth = minRequiredPaddingWidth
		}

		if totalWidth < 1 {
			t.logger.Debug("  Col %d: Calculated totalWidth is zero, setting minimum width to 1.", i)
			totalWidth = 1
		}

		currentMax, _ := widths.OK(i)
		if totalWidth > currentMax {
			widths.Set(i, totalWidth)
			t.logger.Debug("  Col %d: Updated width from %d to %d (content:%d + padL:%d + padR:%d) for cell '%s'", i, currentMax, totalWidth, contentWidth, padLeftWidth, padRightWidth, cell)
		} else {
			t.logger.Debug("  Col %d: Width %d not greater than current max %d for cell '%s'", i, totalWidth, currentMax, cell)
		}
	}
}
