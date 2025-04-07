package renderer

import (
	"fmt"
	"github.com/olekukonko/tablewriter/tw"
	"github.com/olekukonko/tablewriter/twfn"
	"io"
	"strings"
)

// Formatting holds table rendering context
type Formatting struct {
	Position              tw.Position
	Level                 tw.Level
	IsFirst               bool
	IsLast                bool
	MaxWidth              int
	Widths                map[int]int
	ColMaxWidths          map[int]int
	Align                 tw.Align
	ColAligns             map[int]tw.Align
	Padding               tw.Padding
	ColPadding            map[int]tw.Padding
	HasFooter             bool
	MergedCols            tw.MapBool // Columns merged vertically
	MergedRows            tw.MapBool // Columns merged horizontally in this row
	IsMergeStart          bool       // Indicates start of a merge group
	IsMergeEnd            bool       // Indicates end of a merge group
	NextRowContinuesMerge tw.MapBool // Indicates if the NEXT row continues a vertical merge
	NextRowMergedRows     tw.MapBool // Horizontal merge state for the row BELOW the separator
}

// Renderer defines the interface for table rendering
type Renderer interface {
	Header(w io.Writer, headers [][]string, ctx Formatting)
	Row(w io.Writer, rows []string, ctx Formatting)
	Footer(w io.Writer, footers [][]string, ctx Formatting)
	Line(w io.Writer, ctx Formatting)
	Debug() []string
}

// Separators controls separator visibility
type Separators struct {
	ShowHeader     tw.State
	ShowFooter     tw.State
	BetweenRows    tw.State
	BetweenColumns tw.State
}

// Lines controls line visibility
type Lines struct {
	ShowTop        tw.State
	ShowBottom     tw.State
	ShowHeaderLine tw.State
	ShowFooterLine tw.State
}

// Settings holds rendering preferences
type Settings struct {
	Separators     Separators
	Lines          Lines
	TrimWhitespace tw.State
	CompactMode    tw.State
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

// Default is the default table renderer implementation
type Default struct {
	config DefaultConfig
	trace  []string
}

// Private: Internal debugging utility
func (f *Default) debug(format string, a ...interface{}) {
	if f.config.Debug {
		msg := fmt.Sprintf(format, a...)
		traceEntry := fmt.Sprintf("[DEFAULT] %s", msg)
		f.trace = append(f.trace, traceEntry)
	}
}

// Debug returns the debug trace
func (f *Default) Debug() []string {
	return f.trace
}

// Header renders the table header
func (f *Default) Header(w io.Writer, headers [][]string, ctx Formatting) {
	f.debug("Starting Header render: lines=%d, widths=%v", len(headers), ctx.Widths)
	if f.config.Borders.Top.Enabled() {
		f.debug("Rendering top border with level=Top")
		f.Line(w, Formatting{Widths: ctx.Widths, Level: tw.Top, Position: tw.Header, MergedCols: ctx.MergedCols, MergedRows: ctx.MergedRows})
	}

	for i, line := range headers {
		f.debug("Processing header line %d: content=%v, mergedCols=%v", i, line, ctx.MergedCols)
		f.renderLine(w, line, ctx)
		if i < len(headers)-1 && f.config.Settings.Separators.BetweenRows.Enabled() {
			f.debug("Adding row separator between header lines %d and %d", i, i+1)
			f.Line(w, Formatting{Widths: ctx.Widths, Level: tw.Middle, Position: tw.Header, MergedCols: ctx.MergedCols, MergedRows: ctx.MergedRows})
		}
	}

	if f.config.Settings.Lines.ShowHeaderLine.Enabled() && len(headers) > 0 {
		f.debug("Rendering header line separator with %d headers", len(headers))
		f.Line(w, Formatting{Widths: ctx.Widths, Level: tw.Middle, Position: tw.Header, MergedCols: ctx.MergedCols, MergedRows: ctx.MergedRows})
	}
	f.debug("Completed Header render")
}

// Row renders a table row
func (f *Default) Row(w io.Writer, row []string, ctx Formatting) {
	f.debug("Starting Row render: content=%v, isLast=%v, hasFooter=%v, mergedCols=%v, nextRowMerge=%v",
		row, ctx.IsLast, ctx.HasFooter, ctx.MergedCols, ctx.NextRowContinuesMerge)

	f.renderLine(w, row, ctx)

	betweenRowsState := f.config.Settings.Separators.BetweenRows
	isEnabled := betweenRowsState.Enabled()
	f.debug("Row Sep Check: State=%v (%d), Enabled()=%v, !ctx.IsLast=%v",
		betweenRowsState, int(betweenRowsState), isEnabled, !ctx.IsLast)

	shouldConsiderSeparator := isEnabled && !ctx.IsLast

	if shouldConsiderSeparator {
		f.debug("Considering between-rows separator draw after row (isLast=%v)", ctx.IsLast)
		f.Line(w, Formatting{
			Widths:                ctx.Widths,
			Level:                 tw.Middle,
			Position:              tw.Row,
			MergedCols:            ctx.MergedCols,
			MergedRows:            ctx.MergedRows,
			NextRowContinuesMerge: ctx.NextRowContinuesMerge,
		})
	} else {
		f.debug("Not considering between-rows separator draw (isLast=%v, separatorEnabled=%v)",
			ctx.IsLast, isEnabled)
	}

	if ctx.IsLast && !ctx.HasFooter && f.config.Borders.Bottom.Enabled() {
		f.debug("Rendering bottom border for last row (no footer)")
		f.Line(w, Formatting{
			Widths:     ctx.Widths,
			Level:      tw.Bottom,
			Position:   tw.Row,
			MergedCols: nil,
			MergedRows: nil,
		})
	}
	f.debug("Completed Row render")
}

// Footer renders the table footer
func (f *Default) Footer(w io.Writer, footers [][]string, ctx Formatting) {
	f.debug("Starting Footer render: lines=%d", len(footers))
	if f.config.Settings.Lines.ShowFooterLine.Enabled() && len(footers) > 0 {
		f.debug("Rendering footer line separator")
		f.Line(w, Formatting{Widths: ctx.Widths, Level: tw.Middle, Position: tw.Footer, MergedCols: ctx.MergedCols, MergedRows: ctx.MergedRows})
	}

	for i, line := range footers {
		f.debug("Processing footer line %d: content=%v", i, line)
		f.renderLine(w, line, ctx)
		if i < len(footers)-1 && f.config.Settings.Separators.BetweenRows.Enabled() {
			f.debug("Adding separator between footer lines %d and %d", i, i+1)
			f.Line(w, Formatting{Widths: ctx.Widths, Level: tw.Middle, Position: tw.Footer, MergedCols: ctx.MergedCols, MergedRows: ctx.MergedRows})
		}
	}

	if ctx.HasFooter && f.config.Borders.Bottom.Enabled() {
		f.debug("Rendering bottom border for footer")
		f.Line(w, Formatting{Widths: ctx.Widths, Level: tw.Bottom, Position: tw.Footer, MergedCols: ctx.MergedCols, MergedRows: ctx.MergedRows})
	}
	f.debug("Completed Footer render")
}

// Private: Renders a single line of cells
func (f *Default) renderLine(w io.Writer, cells []string, ctx Formatting) {
	f.debug("Starting renderLine: cells=%v, position=%s", cells, ctx.Position)
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

	formattedCells := make([]string, len(cells))
	merged := make(map[int]bool)
	for i := 0; i < len(cells); i++ {
		if merged[i] {
			formattedCells[i] = ""
			f.debug("Cell %d skipped due to previous merge", i)
			continue
		}
		padding := ctx.Padding
		if customPad, ok := ctx.ColPadding[i]; ok {
			padding = customPad
			f.debug("Using custom padding for cell %d: left=%s, right=%s", i, padding.Left, padding.Right)
		}
		align := ctx.Align
		if colAlign, ok := ctx.ColAligns[i]; ok && colAlign != "" {
			align = colAlign
			f.debug("Using column-specific alignment for cell %d: %s", i, align)
		}
		mergedWidth := ctx.Widths[i]

		if ctx.MergedRows != nil && ctx.MergedRows[i] {
			f.debug("Processing horizontal merge starting at cell %d", i)
			for j := i + 1; j < len(cells) && ctx.MergedRows[j]; j++ {
				mergedWidth += ctx.Widths[j]
				if f.config.Settings.Separators.BetweenColumns.Enabled() {
					mergedWidth += twfn.DisplayWidth(f.config.Symbols.Column())
				}
				merged[j] = true
				formattedCells[j] = ""
				f.debug("Merged cell %d into width calculation, new width=%d", j, mergedWidth)
			}
			formattedCells[i] = f.formatCell(cells[i], mergedWidth, padding, align)
			f.debug("Formatted merged cell %d: %s", i, formattedCells[i])
			continue
		}

		if ctx.MergedCols != nil && ctx.MergedCols[i] {
			if ctx.IsMergeStart && strings.TrimSpace(cells[i]) != "" {
				formattedCells[i] = f.formatCell(cells[i], mergedWidth, padding, align)
				f.debug("Formatted cell %d as merge start: %s", i, formattedCells[i])
			} else {
				formattedCells[i] = f.formatCell("", mergedWidth, padding, align)
				f.debug("Cell %d formatted as empty due to vertical merge continuation", i)
			}
		} else {
			formattedCells[i] = f.formatCell(cells[i], mergedWidth, padding, align)
			f.debug("Formatted cell %d: %s", i, formattedCells[i])
		}
	}

	var output string
	if f.config.Settings.Separators.BetweenColumns.Enabled() {
		separator := f.config.Symbols.Column()
		parts := make([]string, 0, len(formattedCells))
		for i, cell := range formattedCells {
			if cell != "" || (!merged[i] && !(ctx.MergedCols != nil && ctx.MergedCols[i])) {
				parts = append(parts, cell)
			}
		}
		output = prefix + strings.Join(parts, separator) + suffix + tw.NewLine
		f.debug("Constructed output with column separators: %s", output)
	} else {
		output = prefix + strings.Join(formattedCells, "") + suffix + tw.NewLine
		f.debug("Constructed output without column separators: %s", output)
	}
	fmt.Fprint(w, output)
	f.debug("renderLine completed")
}

// Line draws a horizontal line based on context
func (f *Default) Line(w io.Writer, ctx Formatting) {
	f.debug("Starting Line render: level=%d, pos=%s, vAbove=%v, vBelow=%v, hAbove=%v, hBelow=%v",
		ctx.Level, ctx.Position, ctx.MergedCols, ctx.NextRowContinuesMerge, ctx.MergedRows, ctx.NextRowMergedRows)

	if !f.shouldDrawLine(ctx) {
		f.debug("Skipping line render - disabled by config or context")
		return
	}

	widths := twfn.ConvertToSorted(ctx.Widths)
	if len(widths) == 0 {
		f.debug("Line render warning: No widths provided.")
		return
	}

	var line strings.Builder
	sym := f.config.Symbols

	if f.config.Borders.Left.Enabled() {
		line.WriteString(f.getLeftBorderSymbol(ctx, sym))
	}

	if f.config.Settings.Separators.BetweenColumns.Enabled() {
		f.buildLineWithSeparators(&line, widths, ctx, sym)
	} else {
		f.buildSolidLine(&line, widths, sym)
	}

	if f.config.Borders.Right.Enabled() {
		line.WriteString(f.getRightBorderSymbol(ctx, sym, len(widths)))
	}

	fmt.Fprintln(w, line.String())
	f.debug("Line render completed: [%s]", line.String())
}

// Private: Determines if a line should be drawn
func (f *Default) shouldDrawLine(ctx Formatting) bool {
	switch ctx.Level {
	case tw.Top:
		return f.config.Borders.Top.Enabled()
	case tw.Bottom:
		return f.config.Borders.Bottom.Enabled()
	case tw.Middle:
		switch ctx.Position {
		case tw.Header:
			return f.config.Settings.Lines.ShowHeaderLine.Enabled() || f.config.Settings.Separators.BetweenRows.Enabled()
		case tw.Row:
			return f.config.Settings.Separators.BetweenRows.Enabled()
		case tw.Footer:
			return f.config.Settings.Lines.ShowFooterLine.Enabled() || f.config.Settings.Separators.BetweenRows.Enabled()
		}
	}
	f.debug("shouldDrawLine: Condition not met (level=%d, pos=%s)", ctx.Level, ctx.Position)
	return false
}

// Private: Gets the left border symbol
func (f *Default) getLeftBorderSymbol(ctx Formatting, sym tw.Symbols) string {
	defaultMap := map[tw.Level]string{tw.Top: sym.TopLeft(), tw.Middle: sym.MidLeft(), tw.Bottom: sym.BottomLeft()}
	symbol := defaultMap[ctx.Level]

	if ctx.Level == tw.Middle {
		vMergeAboveFirstCol := ctx.MergedCols.Get(0)
		vMergeBelowFirstCol := ctx.NextRowContinuesMerge.Get(0)
		if vMergeAboveFirstCol || vMergeBelowFirstCol {
			symbol = sym.Column()
			f.debug("Left Border Symbol: Col 0 V-Merge Active -> '%s'", symbol)
		} else {
			f.debug("Left Border Symbol: Col 0 No V-Merge -> Default '%s'", symbol)
		}
	}
	return symbol
}

// Private: Builds line with separators
func (f *Default) buildLineWithSeparators(line *strings.Builder, widths []int, ctx Formatting, sym tw.Symbols) {
	for i, width := range widths {
		if i > 0 {
			line.WriteString(f.getJunctionSymbol(i, ctx, sym))
		}
		segmentChar := f.getSegmentSymbol(i, ctx, sym)
		if width > 0 {
			line.WriteString(strings.Repeat(segmentChar, width))
		}
	}
}

// Private: Gets junction symbol based on merge states
func (f *Default) getJunctionSymbol(colIndex int, ctx Formatting, sym tw.Symbols) string {
	prevColIndex := colIndex - 1
	junctionChar := ""

	vMergeAboveLeft := ctx.MergedCols.Get(prevColIndex)
	vMergeBelowLeft := ctx.NextRowContinuesMerge.Get(prevColIndex)
	vMergeAboveRight := ctx.MergedCols.Get(colIndex)
	vMergeBelowRight := ctx.NextRowContinuesMerge.Get(colIndex)

	hMergeAbove := ctx.MergedRows.Get(colIndex)
	hMergeBelow := ctx.NextRowMergedRows.Get(colIndex)

	if ctx.Level == tw.Middle {
		isVerticallyMergedLeft := vMergeAboveLeft || vMergeBelowLeft
		isVerticallyMergedRight := vMergeAboveRight || vMergeBelowRight

		f.debug("getJunctionSymbol (Middle, Before col %d): vAL:%t vBL:%t vAR:%t vBR:%t hA:%t hB:%t",
			colIndex, vMergeAboveLeft, vMergeBelowLeft, vMergeAboveRight, vMergeBelowRight, hMergeAbove, hMergeBelow)

		if hMergeAbove && hMergeBelow {
			junctionChar = sym.Row()
		} else if hMergeAbove && !hMergeBelow {
			junctionChar = sym.BottomMid()
		} else if !hMergeAbove && hMergeBelow {
			junctionChar = sym.TopMid()
		} else {
			if isVerticallyMergedLeft && isVerticallyMergedRight {
				junctionChar = sym.Column()
			} else if isVerticallyMergedLeft && !isVerticallyMergedRight {
				junctionChar = sym.MidRight()
			} else if !isVerticallyMergedLeft && isVerticallyMergedRight {
				junctionChar = sym.MidLeft()
			} else {
				junctionChar = sym.Center()
			}
		}
	} else {
		f.debug("getJunctionSymbol (Top/Bottom, Before col %d): hA:%t hB:%t", colIndex, hMergeAbove, hMergeBelow)
		if ctx.Level == tw.Top && hMergeBelow {
			junctionChar = sym.TopMid()
		} else if ctx.Level == tw.Bottom && hMergeAbove {
			junctionChar = sym.BottomMid()
		} else {
			defaultMap := map[tw.Level]string{tw.Top: sym.TopMid(), tw.Bottom: sym.BottomMid()}
			junctionChar = defaultMap[ctx.Level]
		}
	}

	if junctionChar == "" {
		f.debug("Junction Warning: Character not assigned, using fallback")
		fallbackMap := map[tw.Level]string{tw.Top: sym.TopMid(), tw.Middle: sym.Center(), tw.Bottom: sym.BottomMid()}
		junctionChar = fallbackMap[ctx.Level]
	}
	f.debug("Selected Junction Char: '%s'", junctionChar)
	return junctionChar
}

// Private: Gets segment symbol for horizontal lines
func (f *Default) getSegmentSymbol(colIndex int, ctx Formatting, sym tw.Symbols) string {
	segmentChar := sym.Row()
	if ctx.Level == tw.Middle {
		vMergeBelow := ctx.NextRowContinuesMerge.Get(colIndex)
		if vMergeBelow {
			segmentChar = tw.Space
			f.debug("getSegmentSymbol (Col %d): Using space due to VMerge below", colIndex)
		} else {
			f.debug("getSegmentSymbol (Col %d): Using row char '%s'", colIndex, segmentChar)
		}
	}
	return segmentChar
}

// Private: Builds a solid line without separators
func (f *Default) buildSolidLine(line *strings.Builder, widths []int, sym tw.Symbols) {
	totalWidth := 0
	for _, w := range widths {
		if w > 0 {
			totalWidth += w
		}
	}
	if totalWidth > 0 {
		line.WriteString(strings.Repeat(sym.Row(), totalWidth))
	}
	f.debug("buildSolidLine: totalWidth=%d", totalWidth)
}

// Private: Gets the right border symbol
func (f *Default) getRightBorderSymbol(ctx Formatting, sym tw.Symbols, numCols int) string {
	defaultMap := map[tw.Level]string{tw.Top: sym.TopRight(), tw.Middle: sym.MidRight(), tw.Bottom: sym.BottomRight()}
	symbol := defaultMap[ctx.Level]

	lastColIndex := numCols - 1
	if lastColIndex >= 0 && ctx.Level == tw.Middle {
		vMergeAboveLastCol := ctx.MergedCols.Get(lastColIndex)
		vMergeBelowLastCol := ctx.NextRowContinuesMerge.Get(lastColIndex)
		if vMergeAboveLastCol || vMergeBelowLastCol {
			symbol = sym.Column()
			f.debug("Right Border Symbol: Last Col V-Merge Active -> '%s'", symbol)
		} else {
			f.debug("Right Border Symbol: Last Col No V-Merge -> Default '%s'", symbol)
		}
	}
	return symbol
}

// Private: Formats a single cell
func (f *Default) formatCell(content string, width int, padding tw.Padding, align tw.Align) string {
	f.debug("Formatting cell: content='%s', width=%d, align=%s", content, width, align)
	content = strings.TrimSpace(content)
	runeWidth := twfn.DisplayWidth(content)
	padLeftWidth := twfn.DisplayWidth(padding.Left)
	padRightWidth := twfn.DisplayWidth(padding.Right)
	totalPadding := padLeftWidth + padRightWidth
	f.debug("Calculated widths: content=%d, padding=%d+%d", runeWidth, padLeftWidth, padRightWidth)

	availableWidth := width - totalPadding
	if availableWidth < 0 {
		availableWidth = 0
	}

	if runeWidth > availableWidth && availableWidth > 0 {
		content = twfn.TruncateString(content, availableWidth)
		runeWidth = twfn.DisplayWidth(content)
		f.debug("Truncated content to width %d: %s", availableWidth, content)
	}

	var builder strings.Builder
	remaining := width - runeWidth - totalPadding
	f.debug("Remaining space: %d", remaining)

	switch align {
	case tw.AlignLeft:
		builder.WriteString(padding.Left)
		builder.WriteString(content)
		if remaining > 0 {
			builder.WriteString(strings.Repeat(padding.Right, remaining))
		}
		builder.WriteString(padding.Right)
		f.debug("Left aligned cell")
	case tw.AlignRight:
		if remaining > 0 {
			builder.WriteString(strings.Repeat(padding.Left, remaining))
		}
		builder.WriteString(padding.Left)
		builder.WriteString(content)
		builder.WriteString(padding.Right)
		f.debug("Right aligned cell")
	case tw.AlignCenter:
		leftPad := remaining / 2
		rightPad := remaining - leftPad
		if leftPad > 0 {
			builder.WriteString(strings.Repeat(padding.Left, leftPad))
		}
		builder.WriteString(padding.Left)
		builder.WriteString(content)
		builder.WriteString(padding.Right)
		if rightPad > 0 {
			builder.WriteString(strings.Repeat(padding.Right, rightPad))
		}
		f.debug("Center aligned cell: leftPad=%d, rightPad=%d", leftPad, rightPad)
	default:
		builder.WriteString(padding.Left)
		builder.WriteString(content)
		if remaining > 0 {
			builder.WriteString(strings.Repeat(padding.Left, remaining))
		}
		builder.WriteString(padding.Right)
		f.debug("Default (left) aligned cell")
	}
	result := builder.String()
	f.debug("Formatted cell result: %s", result)
	return result
}

// Private: Returns default configuration
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

// NewDefault creates a new Default renderer instance
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
	}
	return &Default{config: cfg}
}

// Private: Merges user settings with defaults
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

// Config returns the current configuration
func (f *Default) Config() DefaultConfig {
	return f.config
}
