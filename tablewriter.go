package tablewriter

import (
	"bytes"
	"io"
	"reflect"
	"runtime"
	"strings"
	"unicode"

	"github.com/olekukonko/errors"
	"github.com/olekukonko/ll"
	"github.com/olekukonko/ll/lh"
	"github.com/olekukonko/tablewriter/pkg/twcache"
	"github.com/olekukonko/tablewriter/pkg/twwarp"
	"github.com/olekukonko/tablewriter/pkg/twwidth"
	"github.com/olekukonko/tablewriter/renderer"
	"github.com/olekukonko/tablewriter/tw"
)

// Table represents a table instance with content and rendering capabilities.
type Table struct {
	writer       io.Writer           // Destination for table output
	counters     []tw.Counter        // Counters for indices
	rows         [][]string          // Row data, one slice of strings per logical row
	headers      [][]string          // Header content
	footers      [][]string          // Footer content
	rawHeaders   []string            // Raw header content
	rawFooters   []string            // Raw footer content
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

	// Caption fields
	caption tw.Caption

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
	stringerCache twcache.Cache[reflect.Type, reflect.Value] // Cache for stringer reflection

	batchRenderNumCols      int
	isBatchRenderNumColsSet bool
}

// renderContext holds the core state for rendering the table.
type renderContext struct {
	table           *Table                                      // Reference to the table instance
	renderer        tw.Renderer                                 // Renderer instance
	cfg             tw.Rendition                                // Renderer configuration
	numCols         int                                         // Total number of columns
	headerLines     [][]string                                  // Processed header lines
	rowLines        [][][]string                                // Processed row lines
	footerLines     [][]string                                  // Processed footer lines
	widths          tw.Mapper[tw.Position, tw.Mapper[int, int]] // Widths per section
	footerPrepared  bool                                        // Tracks if footer is prepared
	emptyColumns    []bool                                      // Tracks which original columns are empty (true if empty)
	visibleColCount int                                         // Count of columns that are NOT empty
	logger          *ll.Logger                                  // Debug trace log
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

		// Streaming
		streamWidths:           tw.NewMapper[int, int](), // Initialize empty mapper for streaming widths
		lastRenderedMergeState: tw.NewMapper[int, tw.MergeState](),
		headerRendered:         false,
		firstRowRendered:       false,
		lastRenderedPosition:   "",
		streamNumCols:          0,
		streamRowCounter:       0,

		//  Cache
		stringerCache: twcache.NewLRU[reflect.Type, reflect.Value](tw.DefaultCacheStringCapacity),
	}

	// set Options
	t.Options(opts...)

	t.logger.Infof("Table initialized with %d options", len(opts))
	return t
}

// NewWriter creates a new table with default settings for backward compatibility.
// It logs the creation if debugging is enabled.
func NewWriter(w io.Writer) *Table {
	t := NewTable(w)
	if t.logger != nil {
		t.logger.Debug("NewWriter created buffered Table")
	}
	return t
}

// Caption sets the table caption (legacy method).
// Defaults to BottomCenter alignment, wrapping to table width.
// Use SetCaptionOptions for more control.
func (t *Table) Caption(caption tw.Caption) *Table { // This is the one we modified
	originalSpot := caption.Spot
	originalAlign := caption.Align

	if caption.Spot == tw.SpotNone {
		caption.Spot = tw.SpotBottomCenter
		t.logger.Debugf("[Table.Caption] Input Spot was SpotNone, defaulting Spot to SpotBottomCenter (%d)", caption.Spot)
	}

	if caption.Align == "" || caption.Align == tw.AlignDefault || caption.Align == tw.AlignNone {
		switch caption.Spot {
		case tw.SpotTopLeft, tw.SpotBottomLeft:
			caption.Align = tw.AlignLeft
		case tw.SpotTopRight, tw.SpotBottomRight:
			caption.Align = tw.AlignRight
		default:
			caption.Align = tw.AlignCenter
		}
		t.logger.Debugf("[Table.Caption] Input Align was empty/default, defaulting Align to %s for Spot %d", caption.Align, caption.Spot)
	}

	t.caption = caption // t.caption on the struct is now updated.
	t.logger.Debugf("Caption method called: Input(Spot:%v, Align:%q), Final(Spot:%v, Align:%q), Text:'%.20s', MaxWidth:%d",
		originalSpot, originalAlign, t.caption.Spot, t.caption.Align, t.caption.Text, t.caption.Width)
	return t
}

// Append adds data to the current row being built for the table.
// This method always contributes to a single logical row in the table.
// To add multiple distinct rows, call Append multiple times (once for each row's data)
// or use the Bulk() method if providing a slice where each element is a row.
func (t *Table) Append(rows ...interface{}) error {
	t.ensureInitialized()

	if t.config.Stream.Enable && t.hasPrinted {
		// Streaming logic remains unchanged, as AutoHeader is a batch-mode concept.
		t.logger.Debugf("Append() called in streaming mode with %d items for a single row", len(rows))
		var rowItemForStream interface{}
		if len(rows) == 1 {
			rowItemForStream = rows[0]
		} else {
			rowItemForStream = rows
		}
		if err := t.streamAppendRow(rowItemForStream); err != nil {
			t.logger.Errorf("Error rendering streaming row: %v", err)
			return errors.Newf("failed to stream append row").Wrap(err)
		}
		return nil
	}

	// Batch Mode Logic
	t.logger.Debugf("Append (Batch) received %d arguments: %v", len(rows), rows)

	var cellsSource interface{}
	if len(rows) == 1 {
		cellsSource = rows[0]
	} else {
		cellsSource = rows
	}
	// Check if we should attempt to auto-generate headers from this append operation.
	// Conditions: AutoHeader is on, no headers are set yet, and this is the first data row.
	isFirstRow := len(t.rows) == 0
	if t.config.Behavior.Structs.AutoHeader.Enabled() && len(t.rawHeaders) == 0 && isFirstRow {
		t.logger.Debug("Append: Triggering AutoHeader for the first row.")
		headers := t.extractHeadersFromStruct(cellsSource)
		if len(headers) > 0 {
			// Set the extracted headers. The Header() method handles the rest.
			t.Header(headers)
		}
	}

	cells, err := t.convertCellsToStrings(cellsSource, t.config.Row)
	if err != nil {
		t.logger.Errorf("Append (Batch) failed for cellsSource %v: %v", cellsSource, err)
		return err
	}
	t.rows = append(t.rows, cells)

	t.logger.Debugf("Append (Batch) completed for one row, total rows in table: %d", len(t.rows))
	return nil
}

// Bulk adds multiple rows from a slice to the table.
// If Behavior.AutoHeader is enabled, no headers set, and rows is a slice of structs,
// automatically extracts/sets headers from the first struct.
func (t *Table) Bulk(rows interface{}) error {
	rv := reflect.ValueOf(rows)
	if rv.Kind() != reflect.Slice {
		return errors.Newf("Bulk expects a slice, got %T", rows)
	}
	if rv.Len() == 0 {
		return nil
	}

	// AutoHeader logic remains here, as it's a "Bulk" operation concept.
	if t.config.Behavior.Structs.AutoHeader.Enabled() && len(t.rawHeaders) == 0 {
		first := rv.Index(0).Interface()
		// We can now correctly get headers from pointers or embedded structs
		headers := t.extractHeadersFromStruct(first)
		if len(headers) > 0 {
			t.Header(headers)
		}
	}

	// The rest of the logic is now just a loop over Append.
	for i := 0; i < rv.Len(); i++ {
		row := rv.Index(i).Interface()
		if err := t.Append(row); err != nil { // Use Append
			return err
		}
	}
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
func (t *Table) Configure(fn func(cfg *Config)) *Table {
	fn(&t.config) // Let the user modify the config directly
	// Handle any immediate side-effects of config changes, e.g., logger state
	if t.config.Debug {
		t.logger.Enable()
		t.logger.Resume() // in case it was suspended
	} else {
		t.logger.Disable()
		t.logger.Suspend() // suspend totally, especially because of tight loops
	}
	t.logger.Debugf("Configure complete. New t.config: %+v", t.config)
	return t
}

// Debug retrieves the accumulated debug trace logs.
// No parameters are required.
// Returns a slice of debug messages including renderer logs.
func (t *Table) Debug() *bytes.Buffer {
	return t.trace
}

// Header sets the table's header content, padding to match column count.
// Parameter elements is a slice of strings for header content.
// No return value.
// In streaming mode, this processes and renders the header immediately.
func (t *Table) Header(elements ...any) {
	t.ensureInitialized()
	t.logger.Debugf("Header() method called with raw variadic elements: %v (len %d). Streaming: %v, Started: %v", elements, len(elements), t.config.Stream.Enable, t.hasPrinted)

	// just forget
	if t.config.Behavior.Header.Hide.Enabled() {
		return
	}

	// add come common default
	if t.config.Header.Formatting.AutoFormat == tw.Unknown {
		t.config.Header.Formatting.AutoFormat = tw.On
	}

	if t.config.Stream.Enable && t.hasPrinted {
		//  Streaming Path
		actualCellsToProcess := t.processVariadic(elements)
		headersAsStrings, err := t.convertCellsToStrings(actualCellsToProcess, t.config.Header)
		if err != nil {
			t.logger.Errorf("Header(): Failed to convert header elements to strings for streaming: %v", err)
			headersAsStrings = []string{} // Use empty on error
		}
		errStream := t.streamRenderHeader(headersAsStrings) // streamRenderHeader handles padding to streamNumCols internally
		if errStream != nil {
			t.logger.Errorf("Error rendering streaming header: %v", errStream)
		}
		return
	}

	//  Batch Path
	processedElements := t.processVariadic(elements)
	t.logger.Debugf("Header() (Batch): Effective cells to process: %v", processedElements)

	headersAsStrings, err := t.convertCellsToStrings(processedElements, t.config.Header)
	if err != nil {
		t.logger.Errorf("Header() (Batch): Failed to convert to strings: %v", err)
		t.rawHeaders = []string{} // Set to empty on error
		return
	}

	t.rawHeaders = headersAsStrings

	t.logger.Debugf("Header set (batch mode).")
}

// Footer sets the table's footer content, padding to match column count.
// Parameter footers is a slice of strings for footer content.
// No return value.
// Footer sets the table's footer content.
// Parameter footers is a slice of strings for footer content.
// In streaming mode, this processes and stores the footer for rendering by Close().
func (t *Table) Footer(elements ...any) {
	t.ensureInitialized()
	t.logger.Debugf("Footer() method called with raw variadic elements: %v (len %d). Streaming: %v, Started: %v", elements, len(elements), t.config.Stream.Enable, t.hasPrinted)

	// just forget
	if t.config.Behavior.Footer.Hide.Enabled() {
		return
	}

	if t.config.Stream.Enable && t.hasPrinted {
		//  Streaming Path
		actualCellsToProcess := t.processVariadic(elements)
		footersAsStrings, err := t.convertCellsToStrings(actualCellsToProcess, t.config.Footer)
		if err != nil {
			t.logger.Errorf("Footer(): Failed to convert footer elements to strings for streaming: %v", err)
			footersAsStrings = []string{} // Use empty on error
		}
		errStream := t.streamStoreFooter(footersAsStrings) // streamStoreFooter handles padding to streamNumCols internally
		if errStream != nil {
			t.logger.Errorf("Error processing streaming footer: %v", errStream)
		}
		return
	}

	//  Batch Path
	processedElements := t.processVariadic(elements)
	t.logger.Debugf("Footer() (Batch): Effective cells to process: %v", processedElements)

	footersAsStrings, err := t.convertCellsToStrings(processedElements, t.config.Footer)
	if err != nil {
		t.logger.Errorf("Footer() (Batch): Failed to convert to strings: %v", err)
		t.rawFooters = []string{} // Set to empty on error
		return
	}

	t.rawFooters = footersAsStrings

	t.logger.Debugf("Footer set (batch mode).")
}

// Options updates the table's Options using a provided function.
// Parameter opts is a function that modifies the Table struct.
// Returns the Table instance for method chaining.
func (t *Table) Options(opts ...Option) *Table {
	// add logger
	if t.logger == nil {
		t.logger = ll.New("table").Handler(lh.NewTextHandler(t.trace))
	}

	// Disable and suspend the logger before applying options to prevent premature
	// debug output from renderer methods (e.g., Blueprint.Rendition) triggered by
	// options like WithRendition. Without this, a previously-enabled logger would
	// still be active on the renderer during option application, causing debug
	// messages even when WithDebug(false) is being applied.
	t.logger.Disable()
	t.logger.Suspend()
	t.renderer.Logger(t.logger)

	// loop through options
	for _, opt := range opts {
		opt(t)
	}

	// force debugging mode if set
	if t.config.Debug {
		t.logger.Enable()
		t.logger.Resume()
	} else {
		t.logger.Disable()
		t.logger.Suspend()
	}

	// Get additional system information for debugging
	goVersion := runtime.Version()
	goOS := runtime.GOOS
	goArch := runtime.GOARCH
	numCPU := runtime.NumCPU()

	// Use the new struct-based info.
	// No type assertions or magic strings needed.
	info := twwidth.Debugging()

	t.logger.Infof("Go Runtime: Version=%s, OS=%s, Arch=%s, CPUs=%d",
		goVersion, goOS, goArch, numCPU)

	t.logger.Infof("Environment: LC_CTYPE=%s, LANG=%s, TERM=%s, TERM_PROGRAM=%s",
		info.Raw.LC_CTYPE,
		info.Raw.LANG,
		info.Raw.TERM,
		info.Raw.TERM_PROGRAM,
	)

	t.logger.Infof("East Asian Detection: Auto=%v, Mode=%s, ModernEnv=%v, CJKLocale=%v",
		info.AutoUseEastAsian,
		info.DetectionMode,
		info.Derived.IsModernEnv,
		info.Derived.IsCJKLocale,
	)

	// send logger to renderer
	t.renderer.Logger(t.logger)
	return t
}

// Reset clears all data (headers, rows, footers, caption) and rendering state
// from the table, allowing the Table instance to be reused for a new table
// with the same configuration and writer.
// It does NOT reset the configuration itself (set by NewTable options or Configure)
// or the underlying io.Writer.
func (t *Table) Reset() {
	t.logger.Debug("Reset() called. Clearing table data and render state.")

	// Clear data slices
	t.rows = nil    // Or t.rows = make([][]string, 0)
	t.headers = nil // Or t.headers = make([][]string, 0)
	t.footers = nil // Or t.footers = make([][]string, 0)
	t.rawHeaders = nil
	t.rawFooters = nil

	// Reset width mappers (important for recalculating widths for the new table)
	t.headerWidths = tw.NewMapper[int, int]()
	t.rowWidths = tw.NewMapper[int, int]()
	t.footerWidths = tw.NewMapper[int, int]()

	// Reset caption
	t.caption = tw.Caption{} // Reset to zero value

	// Reset rendering state flags
	t.hasPrinted = false // Critical for allowing Render() or stream Start() again

	// Reset streaming-specific state

	t.streamWidths = tw.NewMapper[int, int]()
	t.streamFooterLines = nil
	t.headerRendered = false
	t.firstRowRendered = false
	t.lastRenderedLineContent = nil
	t.lastRenderedMergeState = tw.NewMapper[int, tw.MergeState]() // Re-initialize
	t.lastRenderedPosition = ""
	t.streamNumCols = 0
	t.streamRowCounter = 0

	// The stringer and its cache are part of the table's configuration,
	if t.stringerCache == nil {
		t.stringerCache = twcache.NewLRU[reflect.Type, reflect.Value](tw.DefaultCacheStringCapacity)
		t.logger.Debug("Reset(): Stringer cache reset to default capacity.")
	} else {
		t.stringerCache.Purge()
		t.logger.Debug("Reset(): Stringer cache cleared.")
	}

	// If the renderer has its own state that needs resetting after a table is done,
	// this would be the place to call a renderer.Reset() method if it existed.
	// Most current renderers are stateless per render call or reset in their Start/Close.
	// For instance, HTML and SVG renderers have their own Reset method.
	// It might be good practice to call it if available.
	if r, ok := t.renderer.(interface{ Reset() }); ok {
		t.logger.Debug("Reset(): Calling Reset() on the current renderer.")
		r.Reset()
	}

	t.logger.Info("Table instance has been reset.")
}

// Render triggers the table rendering process to the configured writer.
// No parameters are required.
// Returns an error if rendering fails.
func (t *Table) Render() error {
	return t.render()
}

// Lines returns the total number of lines rendered.
// This method is only effective if the WithLineCounter() option was used during
// table initialization and must be called *after* Render().
// It actively searches for the default tw.LineCounter among all active counters.
// It returns -1 if the line counter was not enabled.
func (t *Table) Lines() int {
	for _, counter := range t.counters {
		if lc, ok := counter.(*tw.LineCounter); ok {
			return lc.Total()
		}
	}
	// use -1 to indicate no line counter is attached
	return -1
}

// Counters returns the slice of all active counter instances.
// This is useful when multiple counters are enabled.
// It must be called *after* Render().
func (t *Table) Counters() []tw.Counter {
	return t.counters
}

// Trimmer trims whitespace from a string based on the Table’s configuration.
// It conditionally applies trimming based on TrimSpace and TrimTab settings.
//
// Behavior Matrix:
// - TrimSpace=On, TrimTab=On:   Uses strings.TrimSpace (removes all Unicode space including \t).
// - TrimSpace=On, TrimTab=Off:  Removes spaces/newlines but PRESERVES tabs.
// - TrimSpace=Off, TrimTab=On:  Removes only tabs.
// - TrimSpace=Off, TrimTab=Off: Returns string unchanged.
func (t *Table) Trimmer(str string) string {
	trimSpace := t.config.Behavior.TrimSpace.Enabled()
	trimTab := t.config.Behavior.TrimTab.Enabled()

	// Fast Path 1: If both are enabled (Default), use the stdlib optimized TrimSpace
	if trimSpace && trimTab {
		return strings.TrimSpace(str)
	}

	// Fast Path 2: If both are disabled, return raw string
	if !trimSpace && !trimTab {
		return str
	}

	// Granular Trimming via TrimFunc
	return strings.TrimFunc(str, func(r rune) bool {
		if twwidth.IsTab(r) {
			return trimTab // Return true to trim if TrimTab is On
		}
		if trimSpace {
			return unicode.IsSpace(r) // Trim other whitespace if TrimSpace is On
		}
		return false
	})
}

// appendSingle adds a single row to the table's row data.
// Parameter row is the data to append, converted via stringer if needed.
// Returns an error if conversion or appending fails.
func (t *Table) appendSingle(row interface{}) error {
	t.ensureInitialized() // Already here

	if t.config.Stream.Enable && t.hasPrinted { // If streaming is active
		t.logger.Debugf("appendSingle: Dispatching to streamAppendRow for row: %v", row)
		return t.streamAppendRow(row) // Call the streaming render function
	}

	t.logger.Debugf("appendSingle: Processing for batch mode, row: %v", row)
	cells, err := t.convertCellsToStrings(row, t.config.Row)
	if err != nil {
		t.logger.Debugf("Error in convertCellsToStrings (batch mode): %v", err)
		return err
	}
	t.rows = append(t.rows, cells) // Add to batch storage
	t.logger.Debugf("Row appended to batch t.rows, total batch rows: %d", len(t.rows))
	return nil
}

// buildAligns constructs a map of column alignments from configuration.
// Parameter config provides alignment settings for the section.
// Returns a map of column indices to alignment settings.
func (t *Table) buildAligns(config tw.CellConfig) map[int]tw.Align {
	// Start with global alignment, preferring deprecated Formatting.Alignment
	effectiveGlobalAlign := config.Formatting.Alignment
	if effectiveGlobalAlign == tw.Empty || effectiveGlobalAlign == tw.Skip {
		effectiveGlobalAlign = config.Alignment.Global
		if config.Formatting.Alignment != tw.Empty && config.Formatting.Alignment != tw.Skip {
			t.logger.Warnf("Using deprecated CellFormatting.Alignment (%s). Migrate to CellConfig.Alignment.Global.", config.Formatting.Alignment)
		}
	}

	// Use per-column alignments, preferring deprecated ColumnAligns
	effectivePerColumn := config.ColumnAligns
	if len(effectivePerColumn) == 0 && len(config.Alignment.PerColumn) > 0 {
		effectivePerColumn = make([]tw.Align, len(config.Alignment.PerColumn))
		copy(effectivePerColumn, config.Alignment.PerColumn)
		if len(config.ColumnAligns) > 0 {
			t.logger.Warnf("Using deprecated CellConfig.ColumnAligns (%v). Migrate to CellConfig.Alignment.PerColumn.", config.ColumnAligns)
		}
	}

	// Log input for debugging
	t.logger.Debugf("buildAligns INPUT: deprecated Formatting.Alignment=%s, deprecated ColumnAligns=%v, config.Alignment.Global=%s, config.Alignment.PerColumn=%v",
		config.Formatting.Alignment, config.ColumnAligns, config.Alignment.Global, config.Alignment.PerColumn)

	numColsToUse := t.getNumColsToUse()
	colAlignsResult := make(map[int]tw.Align)
	for i := 0; i < numColsToUse; i++ {
		currentAlign := effectiveGlobalAlign
		if i < len(effectivePerColumn) && effectivePerColumn[i] != tw.Empty && effectivePerColumn[i] != tw.Skip {
			currentAlign = effectivePerColumn[i]
		}
		// Skip validation here; rely on rendering to handle invalid alignments
		colAlignsResult[i] = currentAlign
	}

	t.logger.Debugf("Aligns built: %v (length %d)", colAlignsResult, len(colAlignsResult))
	return colAlignsResult
}

// buildPadding constructs a map of column padding settings from configuration.
// Parameter padding provides padding settings for the section.
// Returns a map of column indices to padding settings.
func (t *Table) buildPadding(padding tw.CellPadding) map[int]tw.Padding {
	numColsToUse := t.getNumColsToUse()
	colPadding := make(map[int]tw.Padding)
	for i := 0; i < numColsToUse; i++ {
		if i < len(padding.PerColumn) && padding.PerColumn[i].Paddable() {
			colPadding[i] = padding.PerColumn[i]
		} else {
			colPadding[i] = padding.Global
		}
	}
	t.logger.Debugf("Padding built: %v (length %d)", colPadding, len(colPadding))
	return colPadding
}

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
		ctx.logger.Debugf("Hierarchical merge FINALIZE WARNING: Invalid block col %d, start %d > end %d", col, startRow, endRow)
		return
	}
	if startRow < 0 || endRow < 0 {
		ctx.logger.Debugf("Hierarchical merge FINALIZE WARNING: Negative row indices col %d, start %d, end %d", col, startRow, endRow)
		return
	}
	requiredLen := endRow + 1
	if requiredLen > len(mctx.rowMerges) {
		ctx.logger.Debugf("Hierarchical merge FINALIZE WARNING: rowMerges slice too short (len %d) for endRow %d", len(mctx.rowMerges), endRow)
		return
	}
	if mctx.rowMerges[startRow] == nil {
		mctx.rowMerges[startRow] = make(map[int]tw.MergeState)
	}
	if mctx.rowMerges[endRow] == nil {
		mctx.rowMerges[endRow] = make(map[int]tw.MergeState)
	}

	finalSpan := (endRow - startRow) + 1
	ctx.logger.Debugf("Finalizing H-merge block: col=%d, startRow=%d, endRow=%d, span=%d", col, startRow, endRow, finalSpan)

	startState := mctx.rowMerges[startRow][col]
	if startState.Hierarchical.Present && startState.Hierarchical.Start {
		startState.Hierarchical.Span = finalSpan
		startState.Hierarchical.End = finalSpan == 1
		mctx.rowMerges[startRow][col] = startState
		ctx.logger.Debugf(" -> Updated start state: %+v", startState.Hierarchical)
	} else {
		ctx.logger.Debugf("Hierarchical merge FINALIZE WARNING: col %d, startRow %d was not marked as Present/Start? Current state: %+v. Attempting recovery.", col, startRow, startState.Hierarchical)
		startState.Hierarchical.Present = true
		startState.Hierarchical.Start = true
		startState.Hierarchical.Span = finalSpan
		startState.Hierarchical.End = finalSpan == 1
		mctx.rowMerges[startRow][col] = startState
	}

	if endRow > startRow {
		endState := mctx.rowMerges[endRow][col]
		if endState.Hierarchical.Present && !endState.Hierarchical.Start {
			endState.Hierarchical.End = true
			endState.Hierarchical.Span = finalSpan
			mctx.rowMerges[endRow][col] = endState
			ctx.logger.Debugf(" -> Updated end state: %+v", endState.Hierarchical)
		} else {
			ctx.logger.Debugf("Hierarchical merge FINALIZE WARNING: col %d, endRow %d was not marked as Present/Continuation? Current state: %+v. Attempting recovery.", col, endRow, endState.Hierarchical)
			endState.Hierarchical.Present = true
			endState.Hierarchical.Start = false
			endState.Hierarchical.End = true
			endState.Hierarchical.Span = finalSpan
			mctx.rowMerges[endRow][col] = endState
		}
	} else {
		ctx.logger.Debugf(" -> Span is 1, startRow is also endRow.")
	}
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
	hasContent := len(t.rawFooters) > 0 || len(t.footers) > 0
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

// Logger retrieves the table's logger instance.
// No parameters are required.
// Returns the ll.Logger instance used for debug tracing.
func (t *Table) Logger() *ll.Logger {
	return t.logger
}

// Renderer retrieves the current renderer instance used by the table.
// No parameters are required.
// Returns the tw.Renderer interface instance.
func (t *Table) Renderer() tw.Renderer {
	t.logger.Debug("Renderer requested")
	return t.renderer
}

// maxColumns calculates the maximum column count across sections.
// No parameters are required.
// Returns the highest number of columns found.
func (t *Table) maxColumns() int {
	m := 0
	if len(t.rawHeaders) > m {
		m = len(t.rawHeaders)
	}
	for _, row := range t.rows {
		if len(row) > m {
			m = len(row)
		}
	}
	if len(t.rawFooters) > m {
		m = len(t.rawFooters)
	}
	t.logger.Debugf("Max columns: %d", m)
	return m
}

// printTopBottomCaption prints the table's caption at the specified top or bottom position.
// It wraps the caption text to fit the table width or a user-defined width, aligns it according
// to the specified alignment, and writes it to the provided writer. If the caption text is empty
// or the spot is invalid, it logs the issue and returns without printing. The function handles
// wrapping errors by falling back to splitting on newlines or using the original text.
func (t *Table) printTopBottomCaption(w io.Writer, actualTableWidth int) {
	t.logger.Debugf("[printCaption Entry] Text=%q, Spot=%v (type %T), Align=%q, UserWidth=%d, ActualTableWidth=%d",
		t.caption.Text, t.caption.Spot, t.caption.Spot, t.caption.Align, t.caption.Width, actualTableWidth)

	currentCaptionSpot := t.caption.Spot
	isValidSpot := currentCaptionSpot >= tw.SpotTopLeft && currentCaptionSpot <= tw.SpotBottomRight
	if t.caption.Text == "" || !isValidSpot {
		t.logger.Debugf("[printCaption] Aborting: Text empty OR Spot invalid...")
		return
	}

	var captionWrapWidth int
	if t.caption.Width > 0 {
		captionWrapWidth = t.caption.Width
		t.logger.Debugf("[printCaption] Using user-defined caption.Width %d for wrapping.", captionWrapWidth)
	} else if actualTableWidth <= 4 {
		captionWrapWidth = twwidth.Width(t.caption.Text)
		t.logger.Debugf("[printCaption] Empty table, no user caption.Width: Using natural caption width %d.", captionWrapWidth)
	} else {
		captionWrapWidth = actualTableWidth
		t.logger.Debugf("[printCaption] Non-empty table, no user caption.Width: Using actualTableWidth %d for wrapping.", captionWrapWidth)
	}

	if captionWrapWidth <= 0 {
		captionWrapWidth = 10
		t.logger.Warnf("[printCaption] captionWrapWidth was %d (<=0). Setting to minimum %d.", captionWrapWidth, 10)
	}
	t.logger.Debugf("[printCaption] Final captionWrapWidth to be used by twwarp: %d", captionWrapWidth)

	wrappedCaptionLines, count := twwarp.WrapString(t.caption.Text, captionWrapWidth)
	if count == 0 {
		t.logger.Errorf("[printCaption] Error from twwarp.WrapString (width %d): %v. Text: %q", captionWrapWidth, count, t.caption.Text)
		if strings.Contains(t.caption.Text, "\n") {
			wrappedCaptionLines = strings.Split(t.caption.Text, "\n")
		} else {
			wrappedCaptionLines = []string{t.caption.Text}
		}
		t.logger.Debugf("[printCaption] Fallback: using %d lines from original text.", len(wrappedCaptionLines))
	}

	if len(wrappedCaptionLines) == 0 && t.caption.Text != "" {
		t.logger.Warn("[printCaption] Wrapping resulted in zero lines for non-empty text. Using fallback.")
		if strings.Contains(t.caption.Text, "\n") {
			wrappedCaptionLines = strings.Split(t.caption.Text, "\n")
		} else {
			wrappedCaptionLines = []string{t.caption.Text}
		}
	} else if t.caption.Text != "" {
		t.logger.Debugf("[printCaption] Wrapped caption into %d lines: %v", len(wrappedCaptionLines), wrappedCaptionLines)
	}

	paddingTargetWidth := actualTableWidth
	if t.caption.Width > 0 {
		paddingTargetWidth = t.caption.Width
	} else if actualTableWidth <= 4 {
		paddingTargetWidth = captionWrapWidth
	}
	t.logger.Debugf("[printCaption] Final paddingTargetWidth for tw.Pad: %d", paddingTargetWidth)

	for i, line := range wrappedCaptionLines {
		align := t.caption.Align
		if align == "" || align == tw.AlignDefault || align == tw.AlignNone {
			switch t.caption.Spot {
			case tw.SpotTopLeft, tw.SpotBottomLeft:
				align = tw.AlignLeft
			case tw.SpotTopRight, tw.SpotBottomRight:
				align = tw.AlignRight
			default:
				align = tw.AlignCenter
			}
			t.logger.Debugf("[printCaption] Line %d: Alignment defaulted to %s based on Spot %v", i, align, t.caption.Spot)
		}
		paddedLine := tw.Pad(line, " ", paddingTargetWidth, align)
		t.logger.Debugf("[printCaption] Printing line %d: InputLine=%q, Align=%s, PaddingTargetWidth=%d, PaddedLine=%q",
			i, line, align, paddingTargetWidth, paddedLine)
		w.Write([]byte(paddedLine))
		w.Write([]byte(tw.NewLine))
	}

	t.logger.Debugf("[printCaption] Finished printing all caption lines.")
}

// prepareFooter processes footer content and applies merges.
// Parameters ctx and mctx hold rendering and merge state.
// No return value.
func (t *Table) prepareFooter(ctx *renderContext, mctx *mergeContext) {
	if len(t.rawFooters) == 0 {
		ctx.logger.Debugf("Skipping footer preparation - no footer data")
		if ctx.widths[tw.Footer] == nil {
			ctx.widths[tw.Footer] = tw.NewMapper[int, int]()
		}
		numCols := ctx.numCols
		for i := 0; i < numCols; i++ {
			ctx.widths[tw.Footer].Set(i, ctx.widths[tw.Row].Get(i))
		}
		t.logger.Debug("Initialized empty footer widths based on row widths: %v", ctx.widths[tw.Footer])
		ctx.footerPrepared = true
		return
	}

	t.logger.Debugf("Preparing footer with merge mode: %d", t.config.Footer.Formatting.MergeMode)

	globalLimitActive := t.config.Widths.Global > 0 || t.config.MaxWidth > 0
	var footerContent [][]string
	if globalLimitActive {
		footerContent = t.prepareContent(t.rawFooters, t.config.Footer, ctx.widths[tw.Footer])
	} else {
		footerContent = ctx.footerLines
	}

	preparedLines, mergeStates, _ := t.prepareWithMerges(footerContent, t.config.Footer, tw.Footer)
	t.footers = preparedLines
	mctx.footerMerges = mergeStates
	ctx.footerLines = t.footers
	t.logger.Debugf("Base footer widths (normalized from rows/header): %v", ctx.widths[tw.Footer])
	t.applyHorizontalMerges(tw.Footer, ctx, mctx.footerMerges)
	ctx.footerPrepared = true
	t.logger.Debugf("Footer preparation completed. Final footer widths: %v", ctx.widths[tw.Footer])
}

// prepareWithMerges processes content and detects horizontal merges.
// Parameters include content, config, and position (Header, Row, Footer).
// Returns processed lines, merge states, and horizontal merge map.
func (t *Table) prepareWithMerges(content [][]string, config tw.CellConfig, position tw.Position) ([][]string, map[int]tw.MergeState, map[int]bool) {
	t.logger.Debugf("PrepareWithMerges START: position=%s, mergeMode=%d", position, config.Formatting.MergeMode)
	if len(content) == 0 {
		t.logger.Debugf("PrepareWithMerges END: No content.")
		return content, nil, nil
	}

	numCols := 0
	if len(content) > 0 && len(content[0]) > 0 { // Assumes content[0] exists and has items
		numCols = len(content[0])
	} else { // Fallback if first line is empty or content is empty
		for _, line := range content { // Find max columns from any line
			if len(line) > numCols {
				numCols = len(line)
			}
		}
		if numCols == 0 { // If still 0, try table-wide max (batch mode context)
			numCols = t.maxColumns()
		}
	}

	if numCols == 0 {
		t.logger.Debugf("PrepareWithMerges END: numCols is zero.")
		return content, nil, nil
	}

	horzMergeMap := make(map[int]bool)      // Tracks if a column is part of any horizontal merge for this logical row
	mergeMap := make(map[int]tw.MergeState) // Final merge states for this logical row

	// Ensure all lines in 'content' are padded to numCols for consistent processing
	// This result is what will be modified and returned.
	result := make([][]string, len(content))
	for i := range content {
		result[i] = padLine(content[i], numCols)
	}

	if config.Formatting.MergeMode&tw.MergeHorizontal != 0 {
		t.logger.Debugf("Checking for horizontal merges (logical cell comparison) for %d visual lines, %d columns", len(content), numCols)

		// Special handling for footer lead merge (often for "TOTAL" spanning empty cells)
		// This logic only applies if it's a footer and typically to the first (often only) visual line.
		if position == tw.Footer && len(content) > 0 {
			lineIdx := 0                                       // Assume footer lead merge applies to the first visual line primarily
			originalLine := padLine(content[lineIdx], numCols) // Use original content for decision
			currentLineResult := result[lineIdx]               // Modify the result line

			firstContentIdx := -1
			var firstContent string
			for c := 0; c < numCols; c++ {
				if c >= len(originalLine) {
					break
				}
				trimmedVal := t.Trimmer(originalLine[c])

				if trimmedVal != "" && trimmedVal != "-" { // "-" is often a placeholder not to merge over
					firstContentIdx = c
					firstContent = originalLine[c] // Store the raw content for placement
					break
				} else if trimmedVal == "-" { // Stop if we hit a hard non-mergeable placeholder
					break
				}
			}

			if firstContentIdx > 0 { // If content starts after the first column
				span := firstContentIdx + 1 // Merge from col 0 up to and including firstContentIdx
				startCol := 0

				allEmptyBefore := true
				for c := 0; c < firstContentIdx; c++ {
					originalLine[c] = t.Trimmer(originalLine[c])
					if c >= len(originalLine) || originalLine[c] != "" {
						allEmptyBefore = false
						break
					}
				}

				if allEmptyBefore {
					t.logger.Debugf("Footer lead-merge applied line %d: content '%s' from col %d moved to col %d, span %d",
						lineIdx, firstContent, firstContentIdx, startCol, span)

					if startCol < len(currentLineResult) {
						currentLineResult[startCol] = firstContent // Place the original content
					}
					for k := startCol + 1; k < startCol+span; k++ { // Clear out other cells in the span
						if k < len(currentLineResult) {
							currentLineResult[k] = tw.Empty
						}
					}

					// Update mergeMap for all visual lines of this logical row
					for visualLine := 0; visualLine < len(result); visualLine++ {
						// Only apply the data move to the line where it was detected,
						// but the merge state should apply to the logical cell (all its visual lines).
						if visualLine != lineIdx { // For other visual lines, just clear the cells in the span
							if startCol < len(result[visualLine]) {
								result[visualLine][startCol] = tw.Empty // Typically empty for other lines in a lead merge
							}
							for k := startCol + 1; k < startCol+span; k++ {
								if k < len(result[visualLine]) {
									result[visualLine][k] = tw.Empty
								}
							}
						}
					}

					// Set merge state for the starting column
					startState := mergeMap[startCol]
					startState.Horizontal = tw.MergeStateOption{Present: true, Span: span, Start: true, End: (span == 1)}
					mergeMap[startCol] = startState
					horzMergeMap[startCol] = true // Mark this column as processed by a merge

					// Set merge state for subsequent columns in the span
					for k := startCol + 1; k < startCol+span; k++ {
						colState := mergeMap[k]
						colState.Horizontal = tw.MergeStateOption{Present: true, Span: span, Start: false, End: k == startCol+span-1}
						mergeMap[k] = colState
						horzMergeMap[k] = true // Mark as processed
					}
				}
			}
		}

		// Standard horizontal merge logic based on full logical cell content
		col := 0
		for col < numCols {
			if horzMergeMap[col] { // If already part of a footer lead-merge, skip
				col++
				continue
			}

			// Get full content of logical cell 'col'
			var currentLogicalCellContentBuilder strings.Builder
			for lineIdx := 0; lineIdx < len(content); lineIdx++ {
				if col < len(content[lineIdx]) {
					currentLogicalCellContentBuilder.WriteString(content[lineIdx][col])
				}
			}

			currentLogicalCellTrimmed := t.Trimmer(currentLogicalCellContentBuilder.String())
			if currentLogicalCellTrimmed == "" || currentLogicalCellTrimmed == "-" {
				col++
				continue
			}

			span := 1
			for nextCol := col + 1; nextCol < numCols; nextCol++ {
				if horzMergeMap[nextCol] { // Don't merge into an already merged (e.g. footer lead) column
					break
				}
				var nextLogicalCellContentBuilder strings.Builder
				for lineIdx := 0; lineIdx < len(content); lineIdx++ {
					if nextCol < len(content[lineIdx]) {
						nextLogicalCellContentBuilder.WriteString(content[lineIdx][nextCol])
					}
				}

				nextLogicalCellTrimmed := t.Trimmer(nextLogicalCellContentBuilder.String())
				if currentLogicalCellTrimmed == nextLogicalCellTrimmed && nextLogicalCellTrimmed != "-" {
					span++
				} else {
					break
				}
			}

			if span > 1 {
				t.logger.Debugf("Standard horizontal merge (logical cell): startCol %d, span %d for content '%s'", col, span, currentLogicalCellTrimmed)
				startState := mergeMap[col]
				startState.Horizontal = tw.MergeStateOption{Present: true, Span: span, Start: true, End: (span == 1)}
				mergeMap[col] = startState
				horzMergeMap[col] = true

				// For all visual lines, clear out the content of the merged-over cells
				for lineIdx := 0; lineIdx < len(result); lineIdx++ {
					for k := col + 1; k < col+span; k++ {
						if k < len(result[lineIdx]) {
							result[lineIdx][k] = tw.Empty
						}
					}
				}

				// Set merge state for subsequent columns in the span
				for k := col + 1; k < col+span; k++ {
					colState := mergeMap[k]
					colState.Horizontal = tw.MergeStateOption{Present: true, Span: span, Start: false, End: k == col+span-1}
					mergeMap[k] = colState
					horzMergeMap[k] = true
				}
				col += span
			} else {
				col++
			}
		}
	}

	t.logger.Debugf("PrepareWithMerges END: position=%s, lines=%d, mergeMapH: %v", position, len(result), func() map[int]tw.MergeStateOption {
		m := make(map[int]tw.MergeStateOption)
		for k, v := range mergeMap {
			m[k] = v.Horizontal
		}
		return m
	}())
	return result, mergeMap, horzMergeMap
}
