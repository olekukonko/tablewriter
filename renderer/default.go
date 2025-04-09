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

type Default struct {
	config DefaultConfig
	trace  []string
}

// DefaultConfig holds the default renderer configuration
type DefaultConfig struct {
	Borders  Border
	Symbols  tw.Symbols
	Settings Settings
	Debug    bool
}

//type JunctionRenderer struct {
//	ctx Formatting
//	sym tw.Symbols
//}
//
//// RenderLeft returns the left border symbol
//func (jr *JunctionRenderer) RenderLeft() string {
//	mergeBelow := getCellMergeState(jr.ctx.Row.Next, 0).Vertical
//	mergeAbove := getCellMergeState(jr.ctx.Row.Current, 0).Vertical
//
//	switch jr.ctx.Level {
//	case tw.LevelHeader:
//		return jr.sym.TopLeft()
//	case tw.LevelFooter:
//		if jr.ctx.Row.Location == tw.LocationFirst {
//			return jr.sym.MidLeft() // Footer separator starts with ├
//		}
//		if mergeAbove {
//			return jr.sym.Column()
//		}
//		return jr.sym.BottomLeft()
//	case tw.LevelBody:
//		if jr.ctx.Row.Location == tw.LocationFirst {
//			return jr.sym.TopLeft()
//		}
//		if jr.ctx.Row.Location == tw.LocationEnd {
//			if mergeBelow {
//				return jr.sym.Column() // Continue vertical merge downward
//			}
//			return jr.sym.BottomLeft() // Default to └ at table end
//		}
//		if mergeAbove && mergeBelow {
//			return jr.sym.Column()
//		} else if mergeAbove {
//			return jr.sym.Column()
//		} else if mergeBelow {
//			return jr.sym.MidLeft()
//		}
//		return jr.sym.MidLeft()
//	}
//	return jr.sym.MidLeft()
//}
//
//// RenderRight returns the right border symbol
//func (jr *JunctionRenderer) RenderRight(lastColIdx int) string {
//	if lastColIdx < 0 {
//		switch jr.ctx.Level {
//		case tw.LevelHeader:
//			return jr.sym.TopRight()
//		case tw.LevelFooter:
//			return jr.sym.BottomRight()
//		case tw.LevelBody:
//			return jr.sym.MidRight()
//		}
//	}
//	mergeBelow := getCellMergeState(jr.ctx.Row.Next, lastColIdx).Vertical
//	mergeAbove := getCellMergeState(jr.ctx.Row.Current, lastColIdx).Horizontal
//
//	switch jr.ctx.Level {
//	case tw.LevelHeader:
//		return jr.sym.TopRight()
//	case tw.LevelFooter:
//		if mergeAbove && getCellMergeState(jr.ctx.Row.Current, lastColIdx).End {
//			return jr.sym.Row()
//		}
//		return jr.sym.BottomRight()
//	case tw.LevelBody:
//		if jr.ctx.Row.Location == tw.LocationEnd {
//			// Check if entire row is horizontally merged
//			for col := 0; col <= lastColIdx; col++ {
//				mergeState := getCellMergeState(jr.ctx.Row.Current, col)
//				if mergeState.Horizontal && mergeState.Start && mergeState.Span > 1 {
//					if col+mergeState.Span-1 >= lastColIdx {
//						return jr.sym.Row() // Full row merge
//					}
//				}
//			}
//			if mergeAbove && mergeBelow {
//				return jr.sym.Column()
//			}
//			return jr.sym.BottomRight()
//		}
//		if mergeAbove && mergeBelow {
//			return jr.sym.Column()
//		}
//		return jr.sym.MidRight()
//	}
//	return jr.sym.MidRight()
//}
//
//// RenderJunction returns the junction symbol between two columns
//func (jr *JunctionRenderer) RenderJunction(leftColIdx, rightColIdx int) string {
//	vMergeLeftAbove := getCellMergeState(jr.ctx.Row.Current, leftColIdx).Vertical
//	vMergeLeftBelow := getCellMergeState(jr.ctx.Row.Next, leftColIdx).Vertical
//	vMergeRightAbove := getCellMergeState(jr.ctx.Row.Current, rightColIdx).Vertical
//	vMergeRightBelow := getCellMergeState(jr.ctx.Row.Next, rightColIdx).Vertical
//
//	hMergeLeft := getCellMergeState(jr.ctx.Row.Current, leftColIdx).Horizontal
//	hMergeRight := getCellMergeState(jr.ctx.Row.Current, rightColIdx).Horizontal
//	hMergeBelowLeft := getCellMergeState(jr.ctx.Row.Next, leftColIdx).Horizontal
//	hMergeBelowRight := getCellMergeState(jr.ctx.Row.Next, rightColIdx).Horizontal
//
//	hMergeAbove := hMergeLeft && hMergeRight
//	hMergeBelow := hMergeBelowLeft && hMergeBelowRight
//
//	switch jr.ctx.Level {
//	case tw.LevelHeader:
//		if jr.ctx.Row.Location == tw.LocationFirst {
//			return jr.sym.TopMid() // Top border uses ┬
//		}
//		// Check if next row has a horizontal merge starting here
//		if hMergeBelowRight && getCellMergeState(jr.ctx.Row.Next, rightColIdx).Start {
//			return jr.sym.BottomMid() // ┴ for merge below
//		}
//		return jr.sym.Center()
//	case tw.LevelBody:
//		if jr.ctx.Row.Location == tw.LocationFirst {
//			return jr.sym.TopMid()
//		}
//		if jr.ctx.Row.Location == tw.LocationEnd {
//			if hMergeAbove {
//				return jr.sym.Row()
//			}
//			return jr.sym.BottomMid()
//		}
//		if hMergeAbove && hMergeBelow {
//			return jr.sym.Row()
//		} else if hMergeAbove {
//			return jr.sym.BottomMid()
//		} else if hMergeBelow {
//			return jr.sym.TopMid()
//		} else if vMergeLeftAbove && vMergeLeftBelow && vMergeRightAbove && vMergeRightBelow {
//			return jr.sym.Column()
//		} else if vMergeLeftAbove && vMergeLeftBelow {
//			return jr.sym.MidRight()
//		} else if vMergeRightAbove && vMergeRightBelow {
//			return jr.sym.MidLeft()
//		}
//		return jr.sym.Center()
//	case tw.LevelFooter:
//		if jr.ctx.Row.Location == tw.LocationFirst {
//			return jr.sym.Center() // Footer separator uses ┼
//		}
//		if hMergeAbove {
//			return jr.sym.Row()
//		}
//		return jr.sym.BottomMid()
//	}
//	return jr.sym.Center()
//}
//
//// GetSegment returns the horizontal segment character
//func (jr *JunctionRenderer) GetSegment(colIdx int) string {
//	segment := jr.sym.Row()
//	if jr.ctx.Level == tw.LevelBody {
//		mergeAbove := getCellMergeState(jr.ctx.Row.Current, colIdx).Vertical
//		mergeBelow := getCellMergeState(jr.ctx.Row.Next, colIdx).Vertical
//		if mergeAbove && mergeBelow {
//			segment = ""
//		}
//	}
//	return segment
//}

// [Unchanged methods: Config, debug, Debug, Header, Row, Footer, renderLine, Line, formatCell, defaultConfig, NewDefault, mergeSettings, getCellMergeState]

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

func (f *Default) Header(w io.Writer, headers [][]string, ctx Formatting) {
	f.debug("Starting Header render: IsSubRow=%v, Location=%v, Pos=%s, lines=%d, widths=%v", ctx.IsSubRow, ctx.Row.Location, ctx.Row.Position, len(ctx.Row.Current), ctx.Row.Widths)
	f.renderLine(w, ctx)
	f.debug("Completed Header render")
}

func (f *Default) Row(w io.Writer, row []string, ctx Formatting) {
	f.debug("Starting Row render: IsSubRow=%v, Location=%v, Pos=%s, hasFooter=%v", ctx.IsSubRow, ctx.Row.Location, ctx.Row.Position, ctx.HasFooter)
	f.renderLine(w, ctx)
	f.debug("Completed Row render")
}

func (f *Default) Footer(w io.Writer, footers [][]string, ctx Formatting) {
	f.debug("Starting Footer render: IsSubRow=%v, Location=%v, Pos=%s", ctx.IsSubRow, ctx.Row.Location, ctx.Row.Position)
	f.renderLine(w, ctx)
	f.debug("Completed Footer render")
}

func (f *Default) renderLine(w io.Writer, ctx Formatting) {
	sortedKeys := twfn.ConvertToSortedKeys(ctx.Row.Widths)
	numCols := len(sortedKeys)
	if numCols > 0 {
		numCols = sortedKeys[len(sortedKeys)-1] + 1
	}
	f.debug("Starting renderLine: numCols=%d, position=%s, isSubRow=%v, widths=%v", numCols, ctx.Row.Position, ctx.IsSubRow, ctx.Row.Widths)

	if numCols == 0 && (!f.config.Borders.Left.Enabled() || !f.config.Borders.Right.Enabled()) {
		fmt.Fprintln(w)
		return
	}

	prefix := ""
	if f.config.Borders.Left.Enabled() {
		prefix = f.config.Symbols.Column()
	}
	suffix := ""
	if f.config.Borders.Right.Enabled() {
		suffix = f.config.Symbols.Column()
	}

	if numCols == 0 {
		fmt.Fprintln(w, prefix+suffix)
		return
	}

	formattedCells := make([]string, numCols)
	colIndex := 0
	for colIndex < numCols {
		cellCtx, ok := ctx.Row.Current[colIndex]
		if !ok {
			width := ctx.Row.Widths[colIndex]
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
		}

		if cellCtx.Merge.Horizontal && cellCtx.Merge.Start {
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
				formattedCells[colIndex+k] = ""
			}
			colIndex += cellCtx.Merge.Span
		} else {
			formattedCells[colIndex] = f.formatCell(cellData, targetWidth, padding, align)
			colIndex++
		}
	}

	var output strings.Builder
	output.WriteString(prefix)
	if f.config.Settings.Separators.BetweenColumns.Enabled() {
		separator := f.config.Symbols.Column()
		firstVisible := true
		for i := 0; i < numCols; i++ {
			if formattedCells[i] != "" {
				if !firstVisible {
					output.WriteString(separator)
				}
				output.WriteString(formattedCells[i])
				firstVisible = false
			}
		}
	} else {
		for i := 0; i < numCols; i++ {
			if formattedCells[i] != "" {
				output.WriteString(formattedCells[i])
			}
		}
	}
	output.WriteString(suffix)
	output.WriteString(tw.NewLine)
	fmt.Fprint(w, output.String())
}

//func (f *Default) Line(w io.Writer, ctx Formatting) {
//	f.debug("Starting Line render: level=%d, pos=%s, widths=%v", ctx.Level, ctx.Row.Position, ctx.Row.Widths)
//	sortedKeys := twfn.ConvertToSortedKeys(ctx.Row.Widths)
//	if len(sortedKeys) == 0 && (!f.config.Borders.Left.Enabled() || !f.config.Borders.Right.Enabled()) {
//		fmt.Fprintln(w)
//		return
//	}
//
//	var line strings.Builder
//	jr := &JunctionRenderer{ctx: ctx, sym: f.config.Symbols}
//
//	if f.config.Borders.Left.Enabled() {
//		line.WriteString(jr.RenderLeft())
//	}
//
//	if len(sortedKeys) == 0 {
//		if f.config.Borders.Right.Enabled() {
//			line.WriteString(jr.RenderRight(-1))
//		}
//		fmt.Fprintln(w, line.String())
//		return
//	}
//
//	if f.config.Settings.Separators.BetweenColumns.Enabled() {
//		for i, colIdx := range sortedKeys {
//			if i > 0 {
//				line.WriteString(jr.RenderJunction(sortedKeys[i-1], colIdx))
//			}
//			segmentChar := jr.GetSegment(colIdx)
//			if width := ctx.Row.Widths[colIdx]; width > 0 {
//				line.WriteString(strings.Repeat(segmentChar, width))
//			}
//		}
//	} else {
//		for _, colIdx := range sortedKeys {
//			segmentChar := jr.GetSegment(colIdx)
//			if width := ctx.Row.Widths[colIdx]; width > 0 {
//				line.WriteString(strings.Repeat(segmentChar, width))
//			}
//		}
//	}
//
//	if f.config.Borders.Right.Enabled() {
//		lastColIdx := sortedKeys[len(sortedKeys)-1]
//		line.WriteString(jr.RenderRight(lastColIdx))
//	}
//
//	fmt.Fprintln(w, line.String())
//	f.debug("Line render completed: [%s]", line.String())
//}

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

//func getCellMergeState(rowCtx map[int]CellContext, colIndex int) MergeState {
//	if rowCtx == nil || colIndex < 0 {
//		return MergeState{}
//	}
//	if cellCtx, ok := rowCtx[colIndex]; ok {
//		return cellCtx.Merge
//	}
//	return MergeState{}
//}
