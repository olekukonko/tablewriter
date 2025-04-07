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
	Truncate     bool
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
	if f.config.Borders.Top.Enabled() {
		f.Line(w, Formatting{Widths: ctx.Widths, Level: Top, Position: Header})
	}

	if ctx.Padding.Top != "" {
		var topCells []string
		for i := range headers {
			width := ctx.Widths[i]
			topCells = append(topCells, strings.Repeat(ctx.Padding.Top, width))
		}
		f.renderLine(w, topCells, ctx)
	}

	var cells []string
	for i, h := range headers {
		width := ctx.Widths[i]
		align := ctx.Align
		if i < len(ctx.ColAligns) && ctx.ColAligns[i] != "" {
			align = ctx.ColAligns[i]
		}
		padding := ctx.Padding
		if customPad, ok := ctx.ColPadding[i]; ok {
			padding = customPad
		}

		cell := f.formatCell(strings.TrimSpace(h), width, align, "", padding, ctx, i)
		cells = append(cells, cell)
	}
	f.renderLine(w, cells, ctx)

	if ctx.Padding.Bottom != "" {
		var bottomCells []string
		for i := range headers {
			width := ctx.Widths[i]
			bottomCells = append(bottomCells, strings.Repeat(ctx.Padding.Bottom, width))
		}
		f.renderLine(w, bottomCells, ctx)
	}

	if f.config.Settings.Lines.ShowHeaderLine.Enabled() {
		f.Line(w, Formatting{Widths: ctx.Widths, Level: Middle, Position: Header})
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
	f.print("DEBUG: Line called with Level:", ctx.Level, "Position:", ctx.Position, "Borders:", f.config.Borders)
	if ctx.Level == Top && f.config.Borders.Top.Disabled() {
		f.print("DEBUG: Skipping top border")
		return
	}
	if ctx.Level == Bottom && f.config.Borders.Bottom.Disabled() {
		f.print("DEBUG: Skipping bottom border")
		return
	}
	if ctx.Level == Middle {
		if ctx.Position == Header && f.config.Settings.Lines.ShowHeaderLine.Disabled() {
			f.print("DEBUG: Skipping header separator")
			return
		}
		if ctx.Position == Row && f.config.Settings.Separators.BetweenRows.Disabled() {
			f.print("DEBUG: Skipping row separator")
			return
		}
		if ctx.Position == Footer && f.config.Settings.Lines.ShowFooterLine.Disabled() {
			f.print("DEBUG: Skipping footer separator")
			return
		}
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

	f.print("DEBUG: Line level:", ctx.Level, "position:", ctx.Position, "output:", line.String())
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
		f.Line(w, Formatting{Widths: ctx.Widths, Level: Bottom, Position: Row})
	} else if f.config.Settings.Separators.BetweenRows.Enabled() && !ctx.IsLast {
		f.Line(w, Formatting{Widths: ctx.Widths, Level: Middle, Position: Row})
	}
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
		f.Line(w, Formatting{Widths: ctx.Widths, Level: Middle, Position: Footer})
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
		f.Line(w, Formatting{Widths: ctx.Widths, Level: Bottom, Position: Footer})
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

func (f *Default) formatCell(content string, width int, defaultAlign string, colAlign string, padding symbols.Padding, ctx Formatting, colIndex int) string {
	align := defaultAlign
	if colAlign != "" {
		align = colAlign
	}

	content = strings.TrimSpace(content)
	runeWidth := utils.RuneWidth(content)

	padLeft := padding.Left
	padRight := padding.Right
	padLeftWidth := utils.RuneWidth(padLeft)
	padRightWidth := utils.RuneWidth(padRight)

	contentWidth := width - padLeftWidth - padRightWidth
	if contentWidth < 0 {
		contentWidth = 0
	}

	// Handle custom padding
	if customPad, ok := ctx.ColPadding[colIndex]; ok && (customPad.Left != "" || customPad.Right != "") {
		padLeft = customPad.Left
		padRight = customPad.Right
		padLeftWidth = utils.RuneWidth(padLeft)
		padRightWidth = utils.RuneWidth(padRight)
		contentWidth = width - padLeftWidth - padRightWidth
		if contentWidth < 0 {
			contentWidth = 0
		}
		if runeWidth > contentWidth {
			if ctx.AutoWrap && !ctx.Truncate {
				lines, _ := utils.WrapString(content, contentWidth)
				content = lines[0]
			} else {
				content = utils.TruncateString(content, contentWidth)
			}
			runeWidth = utils.RuneWidth(content)
		}
		extraSpaces := contentWidth - runeWidth
		leftSpaces := 0
		rightSpaces := 0
		switch align {
		case AlignCenter:
			leftSpaces = extraSpaces / 2
			rightSpaces = extraSpaces - leftSpaces
		case AlignRight:
			leftSpaces = extraSpaces
		case AlignLeft:
			rightSpaces = extraSpaces
		}
		var builder strings.Builder
		builder.WriteString(padLeft)
		builder.WriteString(strings.Repeat(padLeft, leftSpaces))
		builder.WriteString(content)
		builder.WriteString(strings.Repeat(padRight, rightSpaces))
		builder.WriteString(padRight)
		result := builder.String()
		f.print("DEBUG: formatCell - content:", content, "width:", width, "padLeft:", padLeft, "padRight:", padRight, "result:", result)
		return result
	}

	// Standard padding case
	if runeWidth > contentWidth {
		if ctx.AutoWrap && !ctx.Truncate {
			lines, _ := utils.WrapString(content, contentWidth)
			content = lines[0]
		} else {
			content = utils.TruncateString(content, contentWidth)
		}
		runeWidth = utils.RuneWidth(content)
	}

	extraSpaces := contentWidth - runeWidth
	leftSpaces := 0
	rightSpaces := 0
	switch align {
	case AlignCenter:
		leftSpaces = extraSpaces / 2
		rightSpaces = extraSpaces - leftSpaces
	case AlignRight:
		leftSpaces = extraSpaces
	case AlignLeft:
		rightSpaces = extraSpaces
	}

	var builder strings.Builder
	builder.WriteString(padLeft)
	builder.WriteString(strings.Repeat(" ", leftSpaces))
	builder.WriteString(content)
	builder.WriteString(strings.Repeat(" ", rightSpaces))
	builder.WriteString(padRight)

	f.print("DEBUG: formatCell - content:", content, "width:", width, "padLeft:", padLeft, "padRight:", padRight, "result:", builder.String())
	return builder.String()
}

func (f *Default) formatSection(w io.Writer, cells []string, ctx Formatting, isFooter bool) {
	maxLines := 1
	splitCells := make([][]string, len(cells))
	for i, cell := range cells {
		var lines []string
		padLeftWidth := utils.RuneWidth(ctx.Padding.Left)
		padRightWidth := utils.RuneWidth(ctx.Padding.Right)
		if customPad, ok := ctx.ColPadding[i]; ok {
			padLeftWidth = utils.RuneWidth(customPad.Left)
			padRightWidth = utils.RuneWidth(customPad.Right)
		}
		contentWidth := ctx.Widths[i] - padLeftWidth - padRightWidth
		if contentWidth < 0 {
			contentWidth = 0
		}
		if ctx.AutoWrap && !ctx.Truncate && contentWidth > 0 {
			lines, _ = utils.WrapString(cell, contentWidth)
		} else {
			lines = strings.Split(cell, "\n")
			if len(lines) == 1 && utils.RuneWidth(lines[0]) > contentWidth {
				lines[0] = utils.TruncateString(lines[0], contentWidth)
			}
		}
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

func NewDefault(configs ...DefaultConfig) *Default {
	cfg := defaultConfig()
	cfg.debug = true

	if len(configs) > 0 {
		userCfg := configs[0]
		fmt.Println("DEBUG: Input config - Borders:", userCfg.Borders, "Settings.Lines:", userCfg.Settings.Lines)
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
		fmt.Println("DEBUG: Final config - Borders:", cfg.Borders, "Settings.Lines:", cfg.Settings.Lines)
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
