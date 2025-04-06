package renderer

import (
	"fmt"
	"io"
	"strings"

	"github.com/olekukonko/tablewriter/symbols"
	"github.com/olekukonko/tablewriter/utils"
)

const (
	Fail    = -1 // Explicitly disabled
	Success = 1  // Explicitly enabled
)

const (
	On  State = Success
	Off       = Fail
)

type State int

func (o State) Enabled() bool  { return o == Success }
func (o State) Disabled() bool { return o == Fail }
func (o State) Toggle() State {
	if o == Fail {
		return Success
	}
	return Fail
}
func (o State) Cond(c func() bool) bool {
	if o.Enabled() {
		return c()
	}
	return false
}
func (o State) Or(c State) State {
	if o.Enabled() {
		return o
	}
	return c
}
func (o State) String() string {
	if o.Enabled() {
		return "on"
	}
	return "off"
}

const (
	AlignCenter  = "center"
	AlignRight   = "right"
	AlignLeft    = "left"
	AlignDefault = AlignLeft
)

type Position string

const (
	Header Position = "header"
	Row             = "row"
	Footer          = "footer"
)

type Level int

const (
	Top = iota
	Middle
	Bottom
)

type Formatting struct {
	Position     Position
	Level        Level
	IsFirst      bool
	IsLast       bool
	MaxWidth     int
	Widths       map[int]int
	ColMaxWidths map[int]int
	Align        string
	ColAligns    map[int]string
	Padding      symbols.Padding
	ColPadding   map[int]symbols.Padding
	AutoWrap     bool
	HasFooter    bool
}

type Renderer interface {
	Header(w io.Writer, headers []string, ctx Formatting)
	Row(w io.Writer, row []string, ctx Formatting)
	Footer(w io.Writer, footers []string, ctx Formatting)
	Line(w io.Writer, ctx Formatting)
}

type Separators struct {
	ShowHeader     State
	ShowFooter     State
	BetweenRows    State
	BetweenColumns State
}

type Lines struct {
	ShowTop        State
	ShowBottom     State
	ShowHeaderLine State
	ShowFooterLine State
}

type Settings struct {
	Separators     Separators
	Lines          Lines
	TrimWhitespace State
	CompactMode    State
}

type Border struct {
	Left   State
	Right  State
	Top    State
	Bottom State
}

type DefaultConfig struct {
	Borders  Border
	Symbols  symbols.Symbols
	Settings Settings
	debug    bool
}

type Default struct {
	config DefaultConfig
}

func (f *Default) Header(w io.Writer, headers []string, ctx Formatting) {
	f.print("DEBUG: Entering Header, HasFooter:", ctx.HasFooter)
	if f.config.Borders.Top.Enabled() {
		f.Line(w, Formatting{Widths: ctx.Widths, Level: Top})
	}

	sortedWidth := utils.ConvertToSorted(ctx.Widths)
	if ctx.Padding.Top != "" {
		var topCells []string
		for _, width := range sortedWidth {
			topCells = append(topCells, strings.Repeat(ctx.Padding.Top, width))
		}
		f.renderLine(w, topCells, ctx)
	}

	var cells []string
	for i, h := range headers {
		width := ctx.Widths[i]
		content := strings.TrimSpace(h)
		padding := symbols.Padding{Left: ctx.Padding.Left, Right: ctx.Padding.Right}
		if customPad, ok := ctx.ColPadding[i]; ok {
			padding = customPad
		}
		colAlign := ""
		if i < len(ctx.ColAligns) && ctx.ColAligns[i] != "" {
			colAlign = ctx.ColAligns[i]
		}
		cell := f.formatCell(content, width, ctx.Align, colAlign, padding, ctx, i)
		cells = append(cells, cell)
		f.print("DEBUG: Header cell", i, "width:", width, "content:", content, "rendered:", cell)
	}
	f.renderLine(w, cells, ctx)

	if ctx.Padding.Bottom != "" {
		var bottomCells []string
		for _, width := range sortedWidth {
			bottomCells = append(bottomCells, strings.Repeat(ctx.Padding.Bottom, width))
		}
		f.renderLine(w, bottomCells, ctx)
	}

	if f.config.Settings.Lines.ShowHeaderLine.Enabled() {
		f.Line(w, Formatting{Widths: ctx.Widths, Level: Middle})
	}
}

func (f *Default) renderLine(w io.Writer, cells []string, ctx Formatting) {
	if f.config.Symbols.Name() == symbols.NameMarkdown {
		fmt.Fprintf(w, "|%s|%s", strings.Join(cells, "|"), symbols.NewLine)
		return
	}

	prefix := f.config.Symbols.Column()
	if f.config.Borders.Left.Disabled() {
		prefix = ""
	}
	suffix := f.config.Symbols.Column()
	if f.config.Borders.Right.Disabled() {
		suffix = ""
	}

	var output string
	if f.config.Settings.Separators.BetweenColumns.Enabled() {
		separator := f.config.Symbols.Column()
		output = fmt.Sprintf("%s%s%s%s", prefix, strings.Join(cells, separator), suffix, symbols.NewLine)
	} else {
		output = fmt.Sprintf("%s%s%s%s", prefix, strings.Join(cells, ""), suffix, symbols.NewLine)
	}
	f.print("DEBUG: renderLine output:", output)
	fmt.Fprint(w, output)
}

func (f *Default) Line(w io.Writer, ctx Formatting) {
	f.print("DEBUG: Line called with Level:", ctx.Level, "Borders:", f.config.Borders)
	if ctx.Level == Top && f.config.Borders.Top.Disabled() {
		return
	}
	if ctx.Level == Bottom && f.config.Borders.Bottom.Disabled() {
		return
	}

	var line strings.Builder
	widths := utils.ConvertToSorted(ctx.Widths)
	rowChar := f.config.Symbols.Row()

	prefix := rowChar
	if ctx.Level == Top && f.config.Borders.Left.Enabled() {
		prefix = f.config.Symbols.TopLeft()
	} else if ctx.Level == Middle && f.config.Borders.Left.Enabled() {
		prefix = f.config.Symbols.MidLeft()
	} else if ctx.Level == Bottom && f.config.Borders.Left.Enabled() {
		prefix = f.config.Symbols.BottomLeft()
	}

	suffix := rowChar
	if ctx.Level == Top && f.config.Borders.Right.Enabled() {
		suffix = f.config.Symbols.TopRight()
	} else if ctx.Level == Middle && f.config.Borders.Right.Enabled() {
		suffix = f.config.Symbols.MidRight()
	} else if ctx.Level == Bottom && f.config.Borders.Right.Enabled() {
		suffix = f.config.Symbols.BottomRight()
	}

	totalWidth := 0
	for _, w := range widths {
		totalWidth += w
	}

	line.WriteString(prefix)
	if f.config.Settings.Separators.BetweenColumns.Enabled() {
		junction := " "
		if ctx.Level == Top {
			junction = f.config.Symbols.TopMid()
		} else if ctx.Level == Middle {
			junction = f.config.Symbols.Center()
		} else if ctx.Level == Bottom {
			junction = f.config.Symbols.BottomMid()
		}
		for i := 0; i < len(widths); i++ {
			if i > 0 {
				line.WriteString(junction)
			}
			line.WriteString(strings.Repeat(rowChar, widths[i]))
		}
	} else {
		line.WriteString(strings.Repeat(rowChar, totalWidth))
	}
	line.WriteString(suffix)

	f.print("DEBUG: Line level:", ctx.Level, "output:", line.String())
	fmt.Fprintln(w, line.String())
}

func (f *Default) Row(w io.Writer, row []string, ctx Formatting) {
	f.print("DEBUG: Entering Row, IsLast:", ctx.IsLast, "HasFooter:", ctx.HasFooter)
	sortedWidth := utils.ConvertToSorted(ctx.Widths)
	if ctx.Padding.Top != "" {
		var topCells []string
		for _, width := range sortedWidth {
			topCells = append(topCells, strings.Repeat(ctx.Padding.Top, width))
		}
		f.renderLine(w, topCells, ctx)
	}

	f.formatSection(w, row, ctx, false)

	if ctx.Padding.Bottom != "" {
		var bottomCells []string
		for _, width := range sortedWidth {
			bottomCells = append(bottomCells, strings.Repeat(ctx.Padding.Bottom, width))
		}
		f.renderLine(w, bottomCells, ctx)
	}

	if ctx.IsLast && !ctx.HasFooter && f.config.Borders.Bottom.Enabled() {
		f.print("DEBUG: Rendering bottom border from Row")
		f.Line(w, Formatting{Widths: ctx.Widths, Level: Bottom})
	} else if f.config.Settings.Separators.BetweenRows.Enabled() && !ctx.IsLast {
		f.Line(w, Formatting{Widths: ctx.Widths, Level: Middle})
	}
}

func (f *Default) formatSection(w io.Writer, cells []string, ctx Formatting, isFooter bool) {
	maxLines := 1
	splitCells := make([][]string, len(cells))
	for i, cell := range cells {
		lines := strings.Split(cell, "\n")
		splitCells[i] = lines
		if len(lines) > maxLines {
			maxLines = len(lines)
		}
	}

	for lineIdx := 0; lineIdx < maxLines; lineIdx++ {
		var renderedCells []string
		for i := range cells {
			width := ctx.Widths[i]
			content := ""
			if lineIdx < len(splitCells[i]) {
				content = splitCells[i][lineIdx]
			}

			padding := ctx.Padding
			if customPad, ok := ctx.ColPadding[i]; ok {
				padding = customPad
			}

			align := ctx.Align
			if colAlign, ok := ctx.ColAligns[i]; ok && colAlign != "" {
				align = colAlign
			}

			cell := f.formatCell(content, width, align, "", padding, ctx, i)
			renderedCells = append(renderedCells, cell)
			f.print("DEBUG: formatSection cell", i, "line", lineIdx, "width:", width, "content:", content, "rendered:", cell)
		}
		f.renderLine(w, renderedCells, ctx)
	}
}

func (f *Default) formatCell(content string, width int, defaultAlign string, colAlign string, padding symbols.Padding, ctx Formatting, colIndex int) string {
	align := defaultAlign
	if colAlign != "" {
		align = colAlign
	}

	content = strings.TrimSpace(content)
	runeWidth := utils.RuneWidth(content)

	// Use only explicitly defined padding, default to empty if not set
	padLeftWidth := utils.RuneWidth(padding.Left)
	padRightWidth := utils.RuneWidth(padding.Right)
	totalContentWidth := runeWidth + padLeftWidth + padRightWidth

	// Apply max width constraints
	effectiveMaxWidth := ctx.MaxWidth
	if colMaxWidth, ok := ctx.ColMaxWidths[colIndex]; ok && colMaxWidth > 0 {
		effectiveMaxWidth = colMaxWidth
	}
	if effectiveMaxWidth > 0 && totalContentWidth > effectiveMaxWidth {
		usableWidth := effectiveMaxWidth - padLeftWidth - padRightWidth
		if usableWidth > 0 {
			if ctx.AutoWrap {
				lines, wrappedWidth := utils.WrapString(content, usableWidth)
				f.print("DEBUG: formatCell col", colIndex, "content:", content, "wrapped lines:", lines, "wrapped width:", wrappedWidth)
				content = strings.Join(lines, "\n")
				runeWidth = utils.RuneWidth(lines[0])
				totalContentWidth = runeWidth + padLeftWidth + padRightWidth
			} else {
				content = utils.TruncateString(content, usableWidth) + "…"
				runeWidth = utils.RuneWidth(content)
				totalContentWidth = runeWidth + padLeftWidth + padRightWidth
			}
		}
	}

	// Ensure width matches content + explicit padding
	if width < totalContentWidth {
		width = totalContentWidth
	}

	// Use padding characters, no extra spaces unless in padding
	leftPadChar := padding.Left
	if leftPadChar == "" {
		leftPadChar = ""
	}
	rightPadChar := padding.Right
	if rightPadChar == "" {
		rightPadChar = ""
	}

	// Build result with only explicit padding
	var result string
	switch align {
	case AlignCenter:
		extraSpace := width - runeWidth - padLeftWidth - padRightWidth
		leftExtra := extraSpace / 2
		rightExtra := extraSpace - leftExtra
		result = leftPadChar + content + rightPadChar + strings.Repeat(" ", rightExtra)
	case AlignRight:
		extraSpace := width - runeWidth - padLeftWidth - padRightWidth
		result = leftPadChar + strings.Repeat(" ", extraSpace) + content + rightPadChar
	default: // AlignLeft
		extraSpace := width - runeWidth - padLeftWidth - padRightWidth
		result = leftPadChar + content + rightPadChar + strings.Repeat(" ", extraSpace)
	}

	f.print("DEBUG: formatCell col", colIndex, "width:", width, "content:", content, "align:", align, "result:", result)
	return result
}

func (f *Default) Footer(w io.Writer, footers []string, ctx Formatting) {
	f.print("DEBUG: Entering Footer, HasFooter:", ctx.HasFooter)
	sortedWidth := utils.ConvertToSorted(ctx.Widths)
	if ctx.Padding.Top != "" {
		var topCells []string
		for _, width := range sortedWidth {
			topCells = append(topCells, strings.Repeat(ctx.Padding.Top, width))
		}
		f.renderLine(w, topCells, ctx)
	}

	if f.config.Settings.Lines.ShowFooterLine.Enabled() {
		f.Line(w, Formatting{Widths: ctx.Widths, Level: Middle})
	}

	f.formatSection(w, footers, ctx, true)

	if ctx.Padding.Bottom != "" {
		var bottomCells []string
		for _, width := range sortedWidth {
			bottomCells = append(bottomCells, strings.Repeat(ctx.Padding.Bottom, width))
		}
		f.renderLine(w, bottomCells, ctx)
	}

	if ctx.HasFooter && f.config.Borders.Bottom.Enabled() {
		f.print("DEBUG: Rendering bottom border from Footer")
		f.Line(w, Formatting{Widths: ctx.Widths, Level: Bottom})
	}
}

func defaultConfig() DefaultConfig {
	return DefaultConfig{
		Borders: Border{
			Left:   On,
			Right:  On,
			Top:    On,
			Bottom: On,
		},
		Settings: Settings{
			Separators: Separators{
				ShowHeader:     On,
				ShowFooter:     On,
				BetweenRows:    Off,
				BetweenColumns: On,
			},
			Lines: Lines{
				ShowTop:        On,
				ShowBottom:     On,
				ShowHeaderLine: On,
				ShowFooterLine: On,
			},
			TrimWhitespace: On,
			CompactMode:    Off,
		},
		Symbols: symbols.NewSymbols(symbols.StyleLight),
	}
}

func NewDefault(configs ...DefaultConfig) *Default {
	cfg := defaultConfig()
	cfg.debug = true // Enable debug for visibility

	if len(configs) > 0 {
		userCfg := configs[0]
		fmt.Println("DEBUG: Renderer config - Borders:", userCfg.Borders)
		// Fully override Borders
		cfg.Borders = userCfg.Borders
		// Override Symbols if provided
		if userCfg.Symbols != nil {
			cfg.Symbols = userCfg.Symbols
		}
		// Merge Settings
		cfg.Settings = mergeSettings(cfg.Settings, userCfg.Settings)
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

func (f *Default) Config() DefaultConfig {
	return f.config
}

func (f *Default) print(a ...interface{}) {
	if f.config.debug {
		fmt.Println(a...)
	}
}
