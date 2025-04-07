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

// -----------------------------------
// Type Definitions
// -----------------------------------

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

// -----------------------------------
// Configuration Defaults
// -----------------------------------

// defaultConfig returns a default table configuration
func defaultConfig() Config {
	defaultPadding := tw.Padding{Left: tw.Space, Right: tw.Space, Top: tw.Empty, Bottom: tw.Empty}
	return Config{
		MaxWidth: 0, // No maximum width by default
		Header: CellConfig{
			Formatting: CellFormatting{
				MaxWidth:   0,               // No width limit
				AutoWrap:   tw.WrapTruncate, // Truncate long text
				Alignment:  tw.AlignCenter,  // Center-aligned headers
				AutoFormat: true,            // Auto-format header text
				MergeMode:  tw.MergeNone,    // No merging
			},
			Padding: CellPadding{
				Global: defaultPadding, // Standard spacing
			},
		},
		Row: CellConfig{
			Formatting: CellFormatting{
				MaxWidth:   0,             // No width limit
				AutoWrap:   tw.WrapNormal, // Wrap text normally
				Alignment:  tw.AlignLeft,  // Left-aligned rows
				AutoFormat: false,         // No auto-formatting
				MergeMode:  tw.MergeNone,  // No merging
			},
			Padding: CellPadding{
				Global: defaultPadding, // Standard spacing
			},
		},
		Footer: CellConfig{
			Formatting: CellFormatting{
				MaxWidth:   0,             // No width limit
				AutoWrap:   tw.WrapNormal, // Wrap text normally
				Alignment:  tw.AlignRight, // Right-aligned footers
				AutoFormat: false,         // No auto-formatting
				MergeMode:  tw.MergeNone,  // No merging
			},
			Padding: CellPadding{
				Global: defaultPadding, // Standard spacing
			},
		},
		Debug: true, // Debug mode enabled by default
	}
}

// -----------------------------------
// Table Options
// -----------------------------------

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

// -----------------------------------
// Table Structure
// -----------------------------------

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

// -----------------------------------
// Table Constructors
// -----------------------------------

// NewTable creates a new table instance with optional configurations
func NewTable(w io.Writer, opts ...Option) *Table {
	t := &Table{
		writer:       w,
		headerWidths: make(map[int]int),
		rowWidths:    make(map[int]int),
		footerWidths: make(map[int]int),
		renderer:     renderer.NewDefault(),  // Default renderer
		config:       defaultConfig(),        // Default configuration
		newLine:      tw.NewLine,             // Standard newline
		trace:        make([]string, 0, 100), // Pre-allocated debug trace
	}
	for _, opt := range opts {
		opt(t) // Apply each option
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

// -----------------------------------
// Debug Utilities
// -----------------------------------

// debug logs a message to the trace if debug mode is enabled
func (t *Table) debug(format string, a ...interface{}) {
	if t.config.Debug {
		msg := fmt.Sprintf(format, a...)
		traceEntry := fmt.Sprintf("[TABLE] %s", msg)
		t.trace = append(t.trace, traceEntry)
	}
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

// -----------------------------------
// Data Management
// -----------------------------------

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

// -----------------------------------
// Rendering Logic
// -----------------------------------

// Render generates and outputs the table
func (t *Table) Render() error {
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
	headerLines, _, headerHorzMerges, _ := t.prepareWithMerges(t.headers, t.config.Header, tw.Header)
	t.debug("Header prepared: lines=%d, horzMerges=%v", len(headerLines), headerHorzMerges)

	rowLines := make([][][]string, len(t.rows))
	rowHorzMerges := make([]map[int]bool, len(t.rows))
	t.debug("Preparing row content for %d rows", len(t.rows))
	for i, row := range t.rows {
		originalRowCopy := make([][]string, len(row))
		for rIdx, rLine := range row {
			originalRowCopy[rIdx] = make([]string, len(rLine))
			copy(originalRowCopy[rIdx], rLine)
		}

		// Ensure consistent column count
		if len(row) > 0 && len(row[0]) < numCols {
			t.debug("Row %d padded to %d columns", i, numCols)
			for lineIdx := range row {
				for len(row[lineIdx]) < numCols {
					row[lineIdx] = append(row[lineIdx], tw.Empty)
				}
			}
		} else if len(row) > 0 && len(row[0]) > numCols {
			t.debug("Row %d truncated to %d columns", i, numCols)
			for lineIdx := range row {
				row[lineIdx] = row[lineIdx][:numCols]
			}
		}

		preparedLines, _, horzMap, _ := t.prepareWithMerges(row, t.config.Row, tw.Row)
		rowLines[i] = preparedLines
		rowHorzMerges[i] = horzMap
		t.debug("Row %d prepared: lines=%d, horzMerges=%v", i, len(rowLines[i]), rowHorzMerges[i])
	}

	t.debug("Preparing footer content")
	footerLines, _, footerHorzMerges, _ := t.prepareWithMerges(t.footers, t.config.Footer, tw.Footer)
	t.debug("Footer prepared: lines=%d, horzMerges=%v", len(footerLines), footerHorzMerges)

	// Vertical merge detection
	rowVertMerges := make([]tw.MapBool, len(t.rows))
	rowIsMergeStart := make([]tw.MapBool, len(t.rows))
	if t.config.Row.Formatting.MergeMode == tw.MergeVertical || t.config.Row.Formatting.MergeMode == tw.MergeBoth {
		t.debug("Starting Vertical Merge Processing for %d columns", numCols)
		lastRowRawValues := make(map[int]string)
		for i, lines := range rowLines {
			rowVertMerges[i] = make(tw.MapBool)
			rowIsMergeStart[i] = make(tw.MapBool)
			isPreviousRowMerged := make(tw.MapBool)
			if i > 0 {
				isPreviousRowMerged = rowVertMerges[i-1]
			}
			isPreviousRowMergeStart := make(tw.MapBool)
			if i > 0 {
				isPreviousRowMergeStart = rowIsMergeStart[i-1]
			}

			for colIndex := 0; colIndex < numCols; colIndex++ {
				currentValue := tw.Empty
				if len(lines) > 0 && colIndex < len(lines[0]) {
					currentValue = strings.TrimSpace(lines[0][colIndex])
				}
				previousValue, _ := lastRowRawValues[colIndex]
				shouldMerge := i > 0 && currentValue == previousValue &&
					(isPreviousRowMergeStart.Get(colIndex) || isPreviousRowMerged.Get(colIndex))

				if shouldMerge {
					t.debug("Vertical Merge: Merging Row %d, Col %d", i, colIndex)
					rowVertMerges[i][colIndex] = true
					for lineIdx := range rowLines[i] {
						if colIndex < len(rowLines[i][lineIdx]) {
							rowLines[i][lineIdx][colIndex] = tw.Empty
						}
					}
				} else {
					rowIsMergeStart[i][colIndex] = true
				}
				lastRowRawValues[colIndex] = currentValue
			}
		}
		t.debug("Finished Vertical Merge Processing")
	} else {
		t.debug("Vertical Merge Processing SKIPPED")
		for i := range rowLines {
			rowVertMerges[i] = make(map[int]bool)
			rowIsMergeStart[i] = make(map[int]bool)
		}
	}

	// Render sections
	f := t.renderer
	if len(headerLines) > 0 {
		colAligns := t.buildAligns(t.config.Header)
		colPadding := t.buildPadding(t.config.Header.Padding)
		nextRowHMerge := make(map[int]bool)
		if len(t.rows) > 0 {
			nextRowHMerge = rowHorzMerges[0]
		} else if len(t.footers) > 0 {
			nextRowHMerge = footerHorzMerges
		}
		t.debug("Rendering header")
		f.Header(t.writer, headerLines, renderer.Formatting{
			Widths: t.headerWidths, Padding: t.config.Header.Padding.Global, ColPadding: colPadding,
			ColAligns: colAligns, HasFooter: len(t.footers) > 0, MergedRows: headerHorzMerges,
			ColMaxWidths: t.config.Header.ColMaxWidths, NextRowMergedRows: nextRowHMerge,
		})
	}

	for i, lines := range rowLines {
		colAligns := t.buildAligns(t.config.Row)
		colPadding := t.buildPadding(t.config.Row.Padding)
		nextRowContinuesMergeMap := make(map[int]bool)
		if i+1 < len(rowVertMerges) {
			for colIndex := 0; colIndex < numCols; colIndex++ {
				if rowVertMerges[i+1].Get(colIndex) {
					nextRowContinuesMergeMap[colIndex] = true
				}
			}
		}
		nextRowHorzMergeMap := make(map[int]bool)
		if i+1 < len(rowHorzMerges) {
			nextRowHorzMergeMap = rowHorzMerges[i+1]
		} else if i+1 == len(t.rows) && len(t.footers) > 0 {
			nextRowHorzMergeMap = footerHorzMerges
		}

		t.debug("Rendering row %d", i)
		for j, line := range lines {
			f.Row(t.writer, line, renderer.Formatting{
				Widths: t.rowWidths, Padding: t.config.Row.Padding.Global, ColPadding: colPadding,
				ColAligns: colAligns, IsFirst: i == 0 && j == 0, IsLast: i == len(t.rows)-1 && j == len(lines)-1,
				HasFooter: len(t.footers) > 0, MergedCols: rowVertMerges[i], MergedRows: rowHorzMerges[i],
				NextRowContinuesMerge: nextRowContinuesMergeMap, NextRowMergedRows: nextRowHorzMergeMap,
				ColMaxWidths: t.config.Row.ColMaxWidths,
			})
		}
	}

	if len(footerLines) > 0 {
		colAligns := t.buildAligns(t.config.Footer)
		colPadding := t.buildPadding(t.config.Footer.Padding)
		t.debug("Rendering footer")
		f.Footer(t.writer, footerLines, renderer.Formatting{
			Widths: t.footerWidths, Padding: t.config.Footer.Padding.Global, ColPadding: colPadding,
			ColAligns: colAligns, HasFooter: true, MergedRows: footerHorzMerges,
			ColMaxWidths: t.config.Footer.ColMaxWidths,
		})
	}

	t.hasPrinted = true
	t.debug("Render completed")
	return nil
}

// -----------------------------------
// Helper Functions
// -----------------------------------

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

// prepareWithMerges handles content preparation including horizontal merges
func (t *Table) prepareWithMerges(content [][]string, config CellConfig, position tw.Position) ([][]string, map[int]bool, map[int]bool, []map[int]bool) {
	t.debug("PrepareWithMerges START: position=%s, mergeMode=%d", position, config.Formatting.MergeMode)
	if len(content) == 0 {
		return content, nil, nil, nil
	}

	numCols := len(content[0])
	horzMergeMap := make(map[int]bool)
	result := make([][]string, len(content))
	mergeStarts := make([]map[int]bool, len(content))

	for i := range content {
		result[i] = make([]string, numCols)
		copy(result[i], content[i])
		mergeStarts[i] = make(map[int]bool)
	}

	if config.Formatting.MergeMode == tw.MergeHorizontal || config.Formatting.MergeMode == tw.MergeBoth {
		t.debug("Checking for horizontal merges")
		for row := 0; row < len(content); row++ {
			for col := 0; col < numCols-1; col++ {
				currentVal := strings.TrimSpace(content[row][col])
				nextVal := strings.TrimSpace(content[row][col+1])
				if currentVal != tw.Empty && currentVal == nextVal {
					horzMergeMap[col] = true
					horzMergeMap[col+1] = true
					result[row][col+1] = tw.Empty
					t.debug("Horizontal merge at row %d, col %d-%d", row, col, col+1)
				}
			}
		}
	}

	result = t.addPaddingLines(result, config, position)
	t.debug("PrepareWithMerges END: lines=%d", len(result))
	return result, nil, horzMergeMap, nil
}

// addPaddingLines adds top and bottom padding to content
func (t *Table) addPaddingLines(content [][]string, config CellConfig, position tw.Position) [][]string {
	t.debug("Adding padding lines: position=%s", position)
	if len(content) == 0 {
		return content
	}

	result := make([][]string, 0)
	numCols := len(content[0])

	if config.Padding.Global.Top != tw.Empty {
		t.debug("Adding top padding")
		topPadding := make([]string, numCols)
		for i := range topPadding {
			padWidth := t.getWidth(i, position)
			repeatCount := (padWidth + twfn.DisplayWidth(config.Padding.Global.Top) - 1) / twfn.DisplayWidth(config.Padding.Global.Top)
			if repeatCount < 1 {
				repeatCount = 1
			}
			topPadding[i] = strings.Repeat(config.Padding.Global.Top, repeatCount)
		}
		result = append(result, topPadding)
	}

	result = append(result, content...)

	if config.Padding.Global.Bottom != tw.Empty {
		t.debug("Adding bottom padding")
		bottomPadding := make([]string, numCols)
		for i := range bottomPadding {
			padWidth := t.getWidth(i, position)
			repeatCount := (padWidth + twfn.DisplayWidth(config.Padding.Global.Bottom) - 1) / twfn.DisplayWidth(config.Padding.Global.Bottom)
			if repeatCount < 1 {
				repeatCount = 1
			}
			bottomPadding[i] = strings.Repeat(config.Padding.Global.Bottom, repeatCount)
		}
		result = append(result, bottomPadding)
	}

	t.debug("Padding lines added: total lines=%d", len(result))
	return result
}

// getWidth retrieves the width for a column based on position
func (t *Table) getWidth(i int, position tw.Position) int {
	switch position {
	case tw.Header:
		if w, ok := t.headerWidths[i]; ok {
			return w
		}
	case tw.Row:
		if w, ok := t.rowWidths[i]; ok {
			return w
		}
	case tw.Footer:
		if w, ok := t.footerWidths[i]; ok {
			return w
		}
	}
	return twfn.DisplayWidth(t.config.Header.Padding.Global.Top) // Fallback
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
	t.debug("Max columns calculated: %d", max)
	return max
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

// -----------------------------------
// Configuration Merging
// -----------------------------------

// mergeConfig merges a source configuration into a destination configuration
func mergeConfig(dst, src Config) Config {
	t := &Table{config: dst}
	t.debug("Merging config: src.MaxWidth=%d", src.MaxWidth)
	if src.MaxWidth != 0 {
		dst.MaxWidth = src.MaxWidth
	}
	dst.Header = mergeCellConfig(dst.Header, src.Header)
	dst.Row = mergeCellConfig(dst.Row, src.Row)
	dst.Footer = mergeCellConfig(dst.Footer, src.Footer)
	t.debug("Config merged")
	return dst
}

// mergeCellConfig merges a source cell configuration into a destination
func mergeCellConfig(dst, src CellConfig) CellConfig {
	t := &Table{config: Config{Debug: true}}
	t.debug("Merging cell config")
	if src.Formatting.Alignment != tw.Empty {
		dst.Formatting.Alignment = src.Formatting.Alignment
	}
	if src.Formatting.AutoWrap != 0 {
		dst.Formatting.AutoWrap = src.Formatting.AutoWrap
	}
	if src.Formatting.MaxWidth != 0 {
		dst.Formatting.MaxWidth = src.Formatting.MaxWidth
	}
	if src.Formatting.MergeMode != 0 {
		dst.Formatting.MergeMode = src.Formatting.MergeMode
	}
	dst.Formatting.AutoFormat = src.Formatting.AutoFormat || dst.Formatting.AutoFormat

	if src.Padding.Global != (tw.Padding{}) {
		dst.Padding.Global = src.Padding.Global
	}
	if len(src.Padding.PerColumn) > 0 {
		if dst.Padding.PerColumn == nil {
			dst.Padding.PerColumn = make([]tw.Padding, len(src.Padding.PerColumn))
		}
		for i, pad := range src.Padding.PerColumn {
			if pad != (tw.Padding{}) {
				if i < len(dst.Padding.PerColumn) {
					dst.Padding.PerColumn[i] = pad
				} else {
					dst.Padding.PerColumn = append(dst.Padding.PerColumn, pad)
				}
			}
		}
	}

	if src.Callbacks.Global != nil {
		dst.Callbacks.Global = src.Callbacks.Global
	}
	if len(src.Callbacks.PerColumn) > 0 {
		if dst.Callbacks.PerColumn == nil {
			dst.Callbacks.PerColumn = make([]func(), len(src.Callbacks.PerColumn))
		}
		for i, cb := range src.Callbacks.PerColumn {
			if cb != nil {
				if i < len(dst.Callbacks.PerColumn) {
					dst.Callbacks.PerColumn[i] = cb
				} else {
					dst.Callbacks.PerColumn = append(dst.Callbacks.PerColumn, cb)
				}
			}
		}
	}

	if src.Filter != nil {
		dst.Filter = src.Filter
	}

	if len(src.ColumnAligns) > 0 {
		dst.ColumnAligns = src.ColumnAligns
	}

	if len(src.ColMaxWidths) > 0 {
		if dst.ColMaxWidths == nil {
			dst.ColMaxWidths = make(map[int]int)
		}
		for k, v := range src.ColMaxWidths {
			dst.ColMaxWidths[k] = v
		}
	}

	t.debug("Cell config merged")
	return dst
}
