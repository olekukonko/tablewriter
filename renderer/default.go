package renderer

import (
	"fmt"
	"github.com/olekukonko/tablewriter/tw"
	"github.com/olekukonko/tablewriter/twfn"
	"io"
	"strings"
)

// Formatting represents the formatting context for an entire row in a table.
type Formatting struct {
	Row       RowContext // Detailed properties for the row and its cells
	Level     tw.Level   // Hierarchical level (LevelHeader, LevelBody, LevelFooter) - Used for Line drawing
	HasFooter bool       // True if the table includes a footer
	IsSubRow  bool       // True if this is a continuation line of a multi-line row or a padding line
}

// CellContext defines the formatting and state of an individual cell within a row.
type CellContext struct {
	Data    string     // The cell’s content. Should be populated by the caller.
	Align   tw.Align   // Alignment of the cell’s content.
	Padding tw.Padding // Padding around the cell’s content.
	Width   int        // Intrinsic width suggestion (often ignored, uses Row.Widths).
	Merge   MergeState // Merge details, indicating if and how this cell spans rows or columns.
}

// MergeState captures the merging properties of a cell.
type MergeState struct {
	Vertical   bool // True if this cell merges vertically
	Horizontal bool // True if this cell merges horizontally
	Span       int  // Number of rows/cols spanned
	Start      bool // True if this is the start cell of the merge
	End        bool // True if this is the end cell of the merge
}

// RowContext manages layout and relational properties for the row and its columns.
type RowContext struct {
	Position     tw.Position         // Row’s section in the table (Header, Row, Footer).
	Location     tw.Location         // Row’s boundary location (LocationFirst, LocationMiddle, LocationEnd).
	Current      map[int]CellContext // Cells in this row, keyed by column index.
	Previous     map[int]CellContext // Cells from the previous row (visually above); nil if none.
	Next         map[int]CellContext // Cells from the next row (visually below); nil if none.
	Widths       map[int]int         // Final computed width of each column.
	ColMaxWidths map[int]int         // Maximum allowed width for each column.
}

// Renderer defines the interface for rendering a table to an io.Writer.
type Renderer interface {
	Header(w io.Writer, headers [][]string, ctx Formatting)
	Row(w io.Writer, row []string, ctx Formatting)
	Footer(w io.Writer, footers [][]string, ctx Formatting)
	Line(w io.Writer, ctx Formatting)
	Debug() []string
	Config() DefaultConfig // Added for tablewriter to access config
}

// Separators controls separator visibility
type Separators struct {
	ShowHeader     tw.State // Not directly used by default renderer? Might be tablewriter config.
	ShowFooter     tw.State // Not directly used by default renderer? Might be tablewriter config.
	BetweenRows    tw.State // Used by tablewriter to decide when to call Line
	BetweenColumns tw.State // Used by renderer for rendering lines and segments
}

// Lines controls line visibility
type Lines struct {
	ShowTop        tw.State // Used by tablewriter to decide when to call Line
	ShowBottom     tw.State // Used by tablewriter to decide when to call Line
	ShowHeaderLine tw.State // Used by tablewriter to decide when to call Line
	ShowFooterLine tw.State // Used by tablewriter to decide when to call Line
}

// Settings holds rendering preferences
type Settings struct {
	Separators     Separators
	Lines          Lines
	TrimWhitespace tw.State
	CompactMode    tw.State // Not currently used?
}

// Border defines table border states
type Border struct {
	Left   tw.State
	Right  tw.State
	Top    tw.State
	Bottom tw.State
}

// DefaultConfig holds the default renderer configuration
type DefaultConfig struct {
	Borders  Border
	Symbols  tw.Symbols
	Settings Settings
	Debug    bool
}

type Default struct {
	config DefaultConfig
	trace  []string
}

// [Existing struct definitions remain unchanged: Formatting, CellContext, MergeState, RowContext, Renderer, Separators, Lines, Settings, Border, DefaultConfig, Default]

// Config returns the current configuration
func (f *Default) Config() DefaultConfig {
	return f.config
}

func (f *Default) debug(format string, a ...interface{}) {
	if f.config.Debug {
		msg := fmt.Sprintf(format, a...)
		f.trace = append(f.trace, fmt.Sprintf("[DEFAULT] %s", msg))
	}
}

func (f *Default) Debug() []string {
	return f.trace
}

// Header renders the table header.
func (f *Default) Header(w io.Writer, headers [][]string, ctx Formatting) {
	f.debug("Starting Header render: IsSubRow=%v, Location=%v, Pos=%s, lines=%d, widths=%v", ctx.IsSubRow, ctx.Row.Location, ctx.Row.Position, len(ctx.Row.Current), ctx.Row.Widths)
	f.renderLine(w, ctx)
	f.debug("Completed Header render")
}

// Row renders a table row.
func (f *Default) Row(w io.Writer, row []string, ctx Formatting) {
	f.debug("Starting Row render: IsSubRow=%v, Location=%v, Pos=%s, hasFooter=%v", ctx.IsSubRow, ctx.Row.Location, ctx.Row.Position, ctx.HasFooter)
	f.renderLine(w, ctx)
	f.debug("Completed Row render")
}

// Footer renders the table footer.
func (f *Default) Footer(w io.Writer, footers [][]string, ctx Formatting) {
	f.debug("Starting Footer render: IsSubRow=%v, Location=%v, Pos=%s", ctx.IsSubRow, ctx.Row.Location, ctx.Row.Position)
	f.renderLine(w, ctx)
	f.debug("Completed Footer render")
}

// renderLine renders a single line of cells based on ctx.Row.Current.
func (f *Default) renderLine(w io.Writer, ctx Formatting) {
	sortedKeys := twfn.ConvertToSortedKeys(ctx.Row.Widths)

	numCols := 0
	if len(sortedKeys) > 0 {
		maxCol := sortedKeys[len(sortedKeys)-1]
		numCols = maxCol + 1
	}
	f.debug("Starting renderLine: numCols=%d, position=%s, isSubRow=%v, widths=%v", numCols, ctx.Row.Position, ctx.IsSubRow, ctx.Row.Widths)

	if numCols == 0 && (!f.config.Borders.Left.Enabled() || !f.config.Borders.Right.Enabled()) {
		f.debug("renderLine: No columns and no side borders, rendering empty line.")
		fmt.Fprintln(w)
		return
	}

	prefix := ""
	if f.config.Borders.Left.Enabled() {
		prefix = f.config.Symbols.Column()
		f.debug("Added left border prefix: %s", prefix)
	}
	suffix := ""
	if f.config.Borders.Right.Enabled() {
		suffix = f.config.Symbols.Column()
		f.debug("Added right border suffix: %s", suffix)
	}

	if numCols == 0 {
		f.debug("renderLine: No columns, rendering side borders only.")
		fmt.Fprintln(w, prefix+suffix)
		return
	}

	formattedCells := make([]string, numCols)
	colIndex := 0
	for colIndex < numCols {
		cellCtx, ok := ctx.Row.Current[colIndex]
		if !ok {
			f.debug("renderLine Warning: Missing CellContext for column %d. Rendering empty.", colIndex)
			width := 0
			if w, wOK := ctx.Row.Widths[colIndex]; wOK {
				width = w
			}
			formattedCells[colIndex] = f.formatCell("", width, tw.Padding{}, tw.AlignLeft)
			colIndex++
			continue
		}

		padding := cellCtx.Padding
		align := cellCtx.Align
		cellData := cellCtx.Data
		targetWidth := ctx.Row.Widths[colIndex]

		if cellCtx.Merge.Vertical && !cellCtx.Merge.Start {
			cellData = ""
			f.debug("Cell %d content cleared due to vertical merge continuation", colIndex)
		}

		if cellCtx.Merge.Horizontal && cellCtx.Merge.Start {
			f.debug("Processing horizontal merge starting at cell %d, span=%d", colIndex, cellCtx.Merge.Span)
			calculatedSpanWidth := 0
			separatorWidth := 0
			if f.config.Settings.Separators.BetweenColumns.Enabled() {
				separatorWidth = twfn.DisplayWidth(f.config.Symbols.Column())
			}
			for j := 0; j < cellCtx.Merge.Span && colIndex+j < numCols; j++ {
				mergeColIndex := colIndex + j
				if w, wOK := ctx.Row.Widths[mergeColIndex]; wOK {
					calculatedSpanWidth += w
					if j > 0 {
						calculatedSpanWidth += separatorWidth
					}
				}
			}
			targetWidth = calculatedSpanWidth
			formattedCells[colIndex] = f.formatCell(cellData, targetWidth, padding, align)
			for k := 1; k < cellCtx.Merge.Span && colIndex+k < numCols; k++ {
				formattedCells[colIndex+k] = "" // Skip continuation cells
				f.debug("Cell %d skipped due to horizontal merge continuation", colIndex+k)
			}
			colIndex += cellCtx.Merge.Span
		} else {
			formattedCells[colIndex] = f.formatCell(cellData, targetWidth, padding, align)
			f.debug("Formatted cell %d: data='%s', width=%d, align=%s -> '%s'", colIndex, cellData, targetWidth, align, formattedCells[colIndex])
			colIndex++
		}
	}

	var output strings.Builder
	output.WriteString(prefix)
	if f.config.Settings.Separators.BetweenColumns.Enabled() {
		separator := f.config.Symbols.Column()
		firstVisible := true
		for i := 0; i < numCols; i++ {
			if formattedCells[i] != "" { // Only render non-empty cells (skips horizontal merge continuations)
				if !firstVisible {
					output.WriteString(separator)
				}
				output.WriteString(formattedCells[i])
				firstVisible = false
			}
		}
		f.debug("Constructed output with column separators")
	} else {
		for i := 0; i < numCols; i++ {
			if formattedCells[i] != "" {
				output.WriteString(formattedCells[i])
			}
		}
		f.debug("Constructed output without column separators")
	}
	output.WriteString(suffix)
	output.WriteString(tw.NewLine)
	fmt.Fprint(w, output.String())
	f.debug("renderLine completed, output: %q", output.String())
}

// Line draws a horizontal line based on context provided by tablewriter.
func (f *Default) Line(w io.Writer, ctx Formatting) {
	f.debug("Starting Line render: level=%d, pos=%s, widths=%v", ctx.Level, ctx.Row.Position, ctx.Row.Widths)
	sortedKeys := twfn.ConvertToSortedKeys(ctx.Row.Widths)
	if len(sortedKeys) == 0 && (!f.config.Borders.Left.Enabled() || !f.config.Borders.Right.Enabled()) {
		f.debug("Line render warning: No widths and no side borders. Skipping line.")
		fmt.Fprintln(w)
		return
	}

	var line strings.Builder
	sym := f.config.Symbols

	if f.config.Borders.Left.Enabled() {
		if ctx.Row.Location == tw.LocationEnd {
			line.WriteString(sym.BottomLeft())
		} else {
			line.WriteString(f.getLeftBorderSymbol(ctx, sym))
		}
	}

	if len(sortedKeys) == 0 {
		if f.config.Borders.Right.Enabled() {
			if ctx.Row.Location == tw.LocationEnd {
				line.WriteString(sym.BottomRight())
			} else {
				line.WriteString(f.getRightBorderSymbol(ctx, sym, -1))
			}
		}
		fmt.Fprintln(w, line.String())
		return
	}

	if f.config.Settings.Separators.BetweenColumns.Enabled() {
		for i, colIdx := range sortedKeys {
			if i > 0 {
				prevColIdx := sortedKeys[i-1]
				mergeAbove := getCellMergeState(ctx.Row.Current, prevColIdx).Vertical
				mergeBelow := getCellMergeState(ctx.Row.Next, prevColIdx).Vertical
				hMergeAbove := getCellMergeState(ctx.Row.Current, prevColIdx).Horizontal && getCellMergeState(ctx.Row.Current, colIdx).Horizontal
				hMergeBelow := getCellMergeState(ctx.Row.Next, prevColIdx).Horizontal && getCellMergeState(ctx.Row.Next, colIdx).Horizontal
				if hMergeAbove && !hMergeBelow {
					line.WriteString(sym.BottomMid())
				} else if !hMergeAbove && hMergeBelow {
					line.WriteString(sym.TopMid())
				} else if hMergeAbove && hMergeBelow {
					line.WriteString(sym.Row())
				} else if mergeAbove && mergeBelow {
					line.WriteString(sym.Column())
				} else {
					if ctx.Row.Location == tw.LocationEnd {
						line.WriteString(sym.BottomMid())
					} else {
						line.WriteString(f.getJunctionSymbol(prevColIdx, colIdx, ctx, sym))
					}
				}
			}
			segmentChar := f.getSegmentSymbol(colIdx, ctx, sym)
			if width := ctx.Row.Widths[colIdx]; width > 0 {
				line.WriteString(strings.Repeat(segmentChar, width))
			}
		}
	} else {
		for _, colIdx := range sortedKeys {
			segmentChar := sym.Row()
			mergeAbove := getCellMergeState(ctx.Row.Current, colIdx).Vertical
			mergeBelow := getCellMergeState(ctx.Row.Next, colIdx).Vertical
			if ctx.Level == tw.LevelBody && mergeAbove && mergeBelow {
				segmentChar = ""
			}
			if width, ok := ctx.Row.Widths[colIdx]; ok && width > 0 {
				line.WriteString(strings.Repeat(segmentChar, width))
			}
		}
	}

	if f.config.Borders.Right.Enabled() {
		lastColIndex := -1
		if len(sortedKeys) > 0 {
			lastColIndex = sortedKeys[len(sortedKeys)-1]
		}
		if ctx.Row.Location == tw.LocationEnd {
			mergeAbove := getCellMergeState(ctx.Row.Current, lastColIndex).Horizontal
			if mergeAbove && getCellMergeState(ctx.Row.Current, lastColIndex).End {
				line.WriteString(sym.BottomRight())
			} else {
				line.WriteString(sym.BottomRight())
			}
		} else {
			line.WriteString(f.getRightBorderSymbol(ctx, sym, lastColIndex))
		}
	}

	fmt.Fprintln(w, line.String())
	f.debug("Line render completed: [%s]", line.String())
}

// getLeftBorderSymbol gets the left border symbol based on line level and merge state below.
func (f *Default) getLeftBorderSymbol(ctx Formatting, sym tw.Symbols) string {
	mergeBelow := getCellMergeState(ctx.Row.Next, 0).Vertical
	mergeAbove := getCellMergeState(ctx.Row.Current, 0).Vertical

	switch ctx.Level {
	case tw.LevelHeader:
		return sym.TopLeft()
	case tw.LevelFooter:
		if mergeAbove {
			return sym.Column()
		}
		return sym.BottomLeft()
	case tw.LevelBody:
		if mergeAbove && mergeBelow {
			return sym.Column()
		} else if mergeAbove {
			return sym.Column()
		} else if mergeBelow {
			return sym.MidLeft()
		}
		return sym.MidLeft()
	}
	f.debug("getLeftBorderSymbol: Unknown level %d", ctx.Level)
	return sym.MidLeft()
}

// getJunctionSymbol gets junction symbol based on merge states around the junction.
func (f *Default) getJunctionSymbol(leftColIndex, rightColIndex int, ctx Formatting, sym tw.Symbols) string {
	vMergeLeftAbove := getCellMergeState(ctx.Row.Current, leftColIndex).Vertical
	vMergeLeftBelow := getCellMergeState(ctx.Row.Next, leftColIndex).Vertical
	vMergeRightAbove := getCellMergeState(ctx.Row.Current, rightColIndex).Vertical
	vMergeRightBelow := getCellMergeState(ctx.Row.Next, rightColIndex).Vertical

	hMergeLeft := getCellMergeState(ctx.Row.Current, leftColIndex).Horizontal
	hMergeRight := getCellMergeState(ctx.Row.Current, rightColIndex).Horizontal
	hMergeBelowLeft := getCellMergeState(ctx.Row.Next, leftColIndex).Horizontal
	hMergeBelowRight := getCellMergeState(ctx.Row.Next, rightColIndex).Horizontal

	hMergeAbove := hMergeLeft && hMergeRight
	hMergeBelow := hMergeBelowLeft && hMergeBelowRight

	hasContentBelow := ctx.Row.Next != nil && len(ctx.Row.Next) > 0

	switch ctx.Level {
	case tw.LevelHeader:
		if ctx.Row.Location == tw.LocationFirst {
			return sym.TopMid()
		}
		if hMergeBelow {
			return sym.TopMid()
		}
		return sym.Center()
	case tw.LevelBody:
		if hMergeAbove && hMergeBelow {
			return sym.Row() // Continuous line if merged above and below
		} else if hMergeAbove {
			return sym.BottomMid() // Merge above ends here
		} else if hMergeBelow {
			return sym.TopMid() // Merge below starts here
		} else if vMergeLeftAbove && vMergeLeftBelow && vMergeRightAbove && vMergeRightBelow {
			return sym.Column()
		} else if vMergeLeftAbove && vMergeLeftBelow {
			return sym.MidRight()
		} else if vMergeRightAbove && vMergeRightBelow {
			return sym.MidLeft()
		}
		if hasContentBelow || ctx.HasFooter {
			return sym.Center()
		}
		return sym.BottomMid()
	case tw.LevelFooter:
		if hMergeAbove {
			return sym.Row()
		}
		return sym.BottomMid()
	}
	return sym.Center() // Fallback
}

// getSegmentSymbol gets segment symbol for horizontal lines within a column.
func (f *Default) getSegmentSymbol(colIndex int, ctx Formatting, sym tw.Symbols) string {
	segment := sym.Row()
	if ctx.Level == tw.LevelBody {
		mergeAbove := getCellMergeState(ctx.Row.Current, colIndex).Vertical
		mergeBelow := getCellMergeState(ctx.Row.Next, colIndex).Vertical
		if mergeAbove && mergeBelow {
			segment = ""
		}
	}
	return segment
}

// getRightBorderSymbol gets the right border symbol based on line level and merge state below.
func (f *Default) getRightBorderSymbol(ctx Formatting, sym tw.Symbols, lastColIndex int) string {
	if lastColIndex < 0 {
		switch ctx.Level {
		case tw.LevelHeader:
			return sym.TopRight()
		case tw.LevelFooter:
			return sym.BottomRight()
		case tw.LevelBody:
			return sym.MidRight()
		default:
			return sym.MidRight()
		}
	}

	mergeBelow := getCellMergeState(ctx.Row.Next, lastColIndex).Vertical
	mergeAbove := getCellMergeState(ctx.Row.Current, lastColIndex).Vertical

	switch ctx.Level {
	case tw.LevelHeader:
		return sym.TopRight()
	case tw.LevelFooter:
		if mergeAbove {
			return sym.Column()
		}
		return sym.BottomRight()
	case tw.LevelBody:
		if mergeAbove && mergeBelow {
			return sym.Column()
		}
		return sym.MidRight()
	}
	f.debug("getRightBorderSymbol: Unknown level %d", ctx.Level)
	return sym.MidRight()
}

// getCellMergeState safely retrieves merge state from a row context map.
func getCellMergeState(rowCtx map[int]CellContext, colIndex int) MergeState {
	if rowCtx == nil || colIndex < 0 {
		return MergeState{}
	}
	if cellCtx, ok := rowCtx[colIndex]; ok {
		return cellCtx.Merge
	}
	return MergeState{}
}

// formatCell formats a single cell's content to fit the given width, applying padding and alignment.
func (f *Default) formatCell(content string, width int, padding tw.Padding, align tw.Align) string {
	if width < 0 {
		f.debug("formatCell Warning: Received negative width %d, using 0.", width)
		width = 0
	}

	f.debug("Formatting cell: content='%s', width=%d, align=%s, padding={L:'%s' R:'%s'}", content, width, align, padding.Left, padding.Right)
	if f.config.Settings.TrimWhitespace.Enabled() {
		content = strings.TrimSpace(content)
		f.debug("Trimmed content: '%s'", content)
	}

	runeWidth := twfn.DisplayWidth(content)
	padLeftWidth := twfn.DisplayWidth(padding.Left)
	padRightWidth := twfn.DisplayWidth(padding.Right)
	totalPaddingWidth := padLeftWidth + padRightWidth

	availableContentWidth := width - totalPaddingWidth
	if availableContentWidth < 0 {
		availableContentWidth = 0
	}
	f.debug("Available content width: %d", availableContentWidth)

	if runeWidth > availableContentWidth {
		content = twfn.TruncateString(content, availableContentWidth)
		runeWidth = twfn.DisplayWidth(content)
		f.debug("Truncated content to fit %d: '%s' (new width %d)", availableContentWidth, content, runeWidth)
	}

	remainingSpace := width - runeWidth - totalPaddingWidth
	if remainingSpace < 0 {
		remainingSpace = 0
	}
	f.debug("Remaining space for alignment padding: %d", remainingSpace)

	leftPad := padding.Left
	rightPad := padding.Right
	if leftPad == "" {
		leftPad = tw.Space
	}
	if rightPad == "" {
		rightPad = tw.Space
	}

	var result strings.Builder
	switch align {
	case tw.AlignLeft:
		result.WriteString(leftPad)
		result.WriteString(content)
		repeatCount := (width - runeWidth - padLeftWidth) / twfn.DisplayWidth(rightPad)
		if repeatCount > 0 {
			result.WriteString(strings.Repeat(rightPad, repeatCount))
		}
	case tw.AlignRight:
		repeatCount := (width - runeWidth - padRightWidth) / twfn.DisplayWidth(leftPad)
		if repeatCount > 0 {
			result.WriteString(strings.Repeat(leftPad, repeatCount))
		}
		result.WriteString(content)
		result.WriteString(rightPad)
	case tw.AlignCenter:
		extraSpace := remainingSpace
		leftExtra := extraSpace / 2
		rightExtra := extraSpace - leftExtra
		leftRepeat := (padLeftWidth + leftExtra) / twfn.DisplayWidth(leftPad)
		if leftRepeat < 1 {
			leftRepeat = 1
		}
		rightRepeat := (padRightWidth + rightExtra) / twfn.DisplayWidth(rightPad)
		if rightRepeat < 1 {
			rightRepeat = 1
		}
		result.WriteString(strings.Repeat(leftPad, leftRepeat))
		result.WriteString(content)
		result.WriteString(strings.Repeat(rightPad, rightRepeat))
	default:
		result.WriteString(leftPad)
		result.WriteString(content)
		repeatCount := (width - runeWidth - padLeftWidth) / twfn.DisplayWidth(rightPad)
		if repeatCount > 0 {
			result.WriteString(strings.Repeat(rightPad, repeatCount))
		}
	}

	finalWidth := twfn.DisplayWidth(result.String())
	if finalWidth < width {
		switch align {
		case tw.AlignLeft:
			result.WriteString(strings.Repeat(rightPad, (width-finalWidth)/twfn.DisplayWidth(rightPad)))
		case tw.AlignRight:
			pad := strings.Repeat(leftPad, (width-finalWidth)/twfn.DisplayWidth(leftPad))
			result = strings.Builder{}
			result.WriteString(pad)
			result.WriteString(result.String())
		case tw.AlignCenter:
			extra := width - finalWidth
			leftExtra := extra / 2
			result.WriteString(strings.Repeat(rightPad, (extra-leftExtra)/twfn.DisplayWidth(rightPad)))
			pad := strings.Repeat(leftPad, leftExtra/twfn.DisplayWidth(leftPad))
			temp := result.String()
			result = strings.Builder{}
			result.WriteString(pad)
			result.WriteString(temp)
		}
	}

	output := result.String()
	if f.config.Debug && twfn.DisplayWidth(output) != width && width > 0 {
		f.debug("formatCell Warning: Final width %d does not match target %d for result '%s'", twfn.DisplayWidth(output), width, output)
	}

	f.debug("Formatted cell final result: '%s' (target width %d)", output, width)
	return output
}

// [Remaining functions remain unchanged: defaultConfig, NewDefault, mergeSettings]

// defaultConfig returns default configuration
func defaultConfig() DefaultConfig {
	return DefaultConfig{
		Borders: Border{
			Left:   tw.On,
			Right:  tw.On,
			Top:    tw.On,
			Bottom: tw.On,
		},
		Settings: Settings{
			Separators: Separators{
				ShowHeader:     tw.On,
				ShowFooter:     tw.On,
				BetweenRows:    tw.Off,
				BetweenColumns: tw.On,
			},
			Lines: Lines{
				ShowTop:        tw.On,
				ShowBottom:     tw.On,
				ShowHeaderLine: tw.On,
				ShowFooterLine: tw.On,
			},
			TrimWhitespace: tw.On,
			CompactMode:    tw.Off,
		},
		Symbols: tw.NewSymbols(tw.StyleLight),
		Debug:   true,
	}
}

// NewDefault creates a new Default renderer instance, merging provided config with defaults.
func NewDefault(configs ...DefaultConfig) *Default {
	cfg := defaultConfig()
	if len(configs) > 0 {
		userCfg := configs[0]
		if userCfg.Borders.Left != 0 {
			cfg.Borders.Left = userCfg.Borders.Left
		}
		if userCfg.Borders.Right != 0 {
			cfg.Borders.Right = userCfg.Borders.Right
		}
		if userCfg.Borders.Top != 0 {
			cfg.Borders.Top = userCfg.Borders.Top
		}
		if userCfg.Borders.Bottom != 0 {
			cfg.Borders.Bottom = userCfg.Borders.Bottom
		}
		if userCfg.Symbols != nil {
			cfg.Symbols = userCfg.Symbols
		}
		cfg.Settings = mergeSettings(cfg.Settings, userCfg.Settings)
		cfg.Debug = userCfg.Debug
	}
	return &Default{config: cfg}
}

// mergeSettings merges user settings with defaults for the Settings struct.
func mergeSettings(defaults, overrides Settings) Settings {
	if overrides.Separators.ShowHeader != 0 {
		defaults.Separators.ShowHeader = overrides.Separators.ShowHeader
	}
	if overrides.Separators.ShowFooter != 0 {
		defaults.Separators.ShowFooter = overrides.Separators.ShowFooter
	}
	if overrides.Separators.BetweenRows != 0 {
		defaults.Separators.BetweenRows = overrides.Separators.BetweenRows
	}
	if overrides.Separators.BetweenColumns != 0 {
		defaults.Separators.BetweenColumns = overrides.Separators.BetweenColumns
	}
	if overrides.Lines.ShowTop != 0 {
		defaults.Lines.ShowTop = overrides.Lines.ShowTop
	}
	if overrides.Lines.ShowBottom != 0 {
		defaults.Lines.ShowBottom = overrides.Lines.ShowBottom
	}
	if overrides.Lines.ShowHeaderLine != 0 {
		defaults.Lines.ShowHeaderLine = overrides.Lines.ShowHeaderLine
	}
	if overrides.Lines.ShowFooterLine != 0 {
		defaults.Lines.ShowFooterLine = overrides.Lines.ShowFooterLine
	}
	if overrides.TrimWhitespace != 0 {
		defaults.TrimWhitespace = overrides.TrimWhitespace
	}
	if overrides.CompactMode != 0 {
		defaults.CompactMode = overrides.CompactMode
	}
	return defaults
}
