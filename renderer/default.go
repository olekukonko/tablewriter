package renderer

import (
	"fmt"
	"io"
	"strings"

	"github.com/olekukonko/tablewriter/symbols"
	"github.com/olekukonko/tablewriter/utils"
)

const (
	Fail    = -1
	Success = 1
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
	HasFooter    bool
}

type Renderer interface {
	Header(w io.Writer, headers [][]string, ctx Formatting)
	Row(w io.Writer, rows []string, ctx Formatting)
	Footer(w io.Writer, footers [][]string, ctx Formatting)
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

func (f *Default) Header(w io.Writer, headers [][]string, ctx Formatting) {
	if f.config.Borders.Top.Enabled() {
		f.Line(w, Formatting{Widths: ctx.Widths, Level: Top, Position: Header})
	}

	if ctx.Padding.Top != "" {
		var topCells []string
		for i := range ctx.Widths {
			topCells = append(topCells, strings.Repeat(ctx.Padding.Top, ctx.Widths[i]))
		}
		f.renderLine(w, topCells, ctx)
	}

	maxLines := 0
	for _, lines := range headers {
		if len(lines) > maxLines {
			maxLines = len(lines)
		}
	}
	fmt.Printf("DEBUG: Header maxLines=%d\n", maxLines)
	for lineIdx := 0; lineIdx < maxLines; lineIdx++ {
		var cells []string
		for i := 0; i < len(ctx.Widths); i++ {
			content := ""
			if i < len(headers) && lineIdx < len(headers[i]) {
				content = headers[i][lineIdx]
			}
			padding := ctx.Padding
			if customPad, ok := ctx.ColPadding[i]; ok {
				padding = customPad
			}
			align := AlignCenter
			if colAlign, ok := ctx.ColAligns[i]; ok && colAlign != "" {
				align = colAlign
			}
			cell := f.formatCell(content, ctx.Widths[i], padding, align)
			cells = append(cells, cell)
			fmt.Printf("DEBUG: Header cell[%d][%d]: content='%s', width=%d, align=%s, result='%s'\n", i, lineIdx, content, ctx.Widths[i], align, cell)
		}
		f.renderLine(w, cells, ctx)
	}

	if ctx.Padding.Bottom != "" {
		var bottomCells []string
		for i := range ctx.Widths {
			bottomCells = append(bottomCells, strings.Repeat(ctx.Padding.Bottom, ctx.Widths[i]))
		}
		f.renderLine(w, bottomCells, ctx)
	}

	if f.config.Settings.Lines.ShowHeaderLine.Enabled() && len(headers) > 0 {
		f.Line(w, Formatting{Widths: ctx.Widths, Level: Middle, Position: Header})
	}
}

func (f *Default) renderLine(w io.Writer, cells []string, ctx Formatting) {
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
	if ctx.Level == Top && f.config.Borders.Top.Disabled() {
		return
	}
	if ctx.Level == Bottom && f.config.Borders.Bottom.Disabled() {
		return
	}
	if ctx.Level == Middle {
		if ctx.Position == Header && f.config.Settings.Lines.ShowHeaderLine.Disabled() {
			return
		}
		if ctx.Position == Row && f.config.Settings.Separators.BetweenRows.Disabled() {
			return
		}
		if ctx.Position == Footer && f.config.Settings.Lines.ShowFooterLine.Disabled() {
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

	line.WriteString(prefix)
	if f.config.Settings.Separators.BetweenColumns.Enabled() {
		junction := f.config.Symbols.TopMid()
		if ctx.Level == Middle {
			junction = f.config.Symbols.Center()
		} else if ctx.Level == Bottom {
			junction = f.config.Symbols.BottomMid()
		}
		for i, width := range widths {
			if i > 0 {
				line.WriteString(junction)
			}
			line.WriteString(strings.Repeat(rowChar, width))
		}
	} else {
		totalWidth := 0
		for _, w := range widths {
			totalWidth += w
		}
		line.WriteString(strings.Repeat(rowChar, totalWidth))
	}
	line.WriteString(suffix)

	f.print("DEBUG: Line output:", line.String())
	fmt.Fprintln(w, line.String())
}

func (f *Default) Row(w io.Writer, row []string, ctx Formatting) {
	if ctx.Padding.Top != "" {
		var topCells []string
		for i := range ctx.Widths {
			topCells = append(topCells, strings.Repeat(ctx.Padding.Top, ctx.Widths[i]))
		}
		f.renderLine(w, topCells, ctx)
	}

	f.formatSection(w, row, ctx)

	if ctx.Padding.Bottom != "" {
		var bottomCells []string
		for i := range ctx.Widths {
			bottomCells = append(bottomCells, strings.Repeat(ctx.Padding.Bottom, ctx.Widths[i]))
		}
		f.renderLine(w, bottomCells, ctx)
	}

	if ctx.IsLast && !ctx.HasFooter && f.config.Borders.Bottom.Enabled() {
		f.Line(w, Formatting{Widths: ctx.Widths, Level: Bottom, Position: Row})
	} else if f.config.Settings.Separators.BetweenRows.Enabled() && !ctx.IsLast {
		f.Line(w, Formatting{Widths: ctx.Widths, Level: Middle, Position: Row})
	}
}

func (f *Default) Footer(w io.Writer, footers [][]string, ctx Formatting) {
	if ctx.Padding.Top != "" {
		var topCells []string
		for i := range ctx.Widths {
			topCells = append(topCells, strings.Repeat(ctx.Padding.Top, ctx.Widths[i]))
		}
		f.renderLine(w, topCells, ctx)
	}

	if f.config.Settings.Lines.ShowFooterLine.Enabled() && len(footers) > 0 {
		f.Line(w, Formatting{Widths: ctx.Widths, Level: Middle, Position: Footer})
	}

	maxLines := 0
	for _, lines := range footers {
		if len(lines) > maxLines {
			maxLines = len(lines)
		}
	}
	fmt.Printf("DEBUG: Footer maxLines=%d\n", maxLines)
	for lineIdx := 0; lineIdx < maxLines; lineIdx++ {
		var cells []string
		for i := 0; i < len(ctx.Widths); i++ {
			content := ""
			if i < len(footers) && lineIdx < len(footers[i]) {
				content = footers[i][lineIdx]
			}
			padding := ctx.Padding
			if customPad, ok := ctx.ColPadding[i]; ok {
				padding = customPad
			}
			align := AlignRight
			if colAlign, ok := ctx.ColAligns[i]; ok && colAlign != "" {
				align = colAlign
			}
			cell := f.formatCell(content, ctx.Widths[i], padding, align)
			cells = append(cells, cell)
			fmt.Printf("DEBUG: Footer cell[%d][%d]: content='%s', width=%d, align=%s, result='%s'\n", i, lineIdx, content, ctx.Widths[i], align, cell)
		}
		f.renderLine(w, cells, ctx)
	}

	if ctx.Padding.Bottom != "" {
		var bottomCells []string
		for i := range ctx.Widths {
			bottomCells = append(bottomCells, strings.Repeat(ctx.Padding.Bottom, ctx.Widths[i]))
		}
		f.renderLine(w, bottomCells, ctx)
	}

	if ctx.HasFooter && f.config.Borders.Bottom.Enabled() {
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

func (f *Default) formatCell(content string, width int, padding symbols.Padding, align string) string {
	content = strings.TrimSpace(content)
	runeWidth := utils.RuneWidth(content)
	padLeftWidth := utils.RuneWidth(padding.Left)
	padRightWidth := utils.RuneWidth(padding.Right)
	totalPadding := padLeftWidth + padRightWidth

	if runeWidth > width-totalPadding && width > 0 {
		content = utils.TruncateString(content, width-totalPadding)
		runeWidth = utils.RuneWidth(content)
	}

	var builder strings.Builder
	switch align {
	case AlignLeft:
		builder.WriteString(padding.Left)
		builder.WriteString(content)
		remaining := width - runeWidth - padLeftWidth - padRightWidth
		if remaining > 0 {
			builder.WriteString(strings.Repeat(padding.Right, remaining))
		}
		builder.WriteString(padding.Right)
	case AlignRight:
		remaining := width - runeWidth - padLeftWidth - padRightWidth
		if remaining > 0 {
			builder.WriteString(strings.Repeat(padding.Left, remaining))
		}
		builder.WriteString(padding.Left)
		builder.WriteString(content)
		builder.WriteString(padding.Right)
	case AlignCenter:
		remaining := width - runeWidth - padLeftWidth - padRightWidth
		leftPad := remaining / 2
		rightPad := remaining - leftPad
		if leftPad > 0 {
			builder.WriteString(strings.Repeat(padding.Left, leftPad))
		}
		builder.WriteString(padding.Left)
		builder.WriteString(content)
		if rightPad > 0 {
			builder.WriteString(strings.Repeat(padding.Right, rightPad))
		}
		builder.WriteString(padding.Right)
	default:
		builder.WriteString(padding.Left)
		builder.WriteString(content)
		remaining := width - runeWidth - padLeftWidth - padRightWidth
		if remaining > 0 {
			builder.WriteString(strings.Repeat(padding.Right, remaining))
		}
		builder.WriteString(padding.Right)
	}
	result := builder.String()
	fmt.Printf("DEBUG: formatCell - content='%s', width=%d, align=%s, runeWidth=%d, totalPadding=%d, result='%s'\n", content, width, align, runeWidth, totalPadding, result)
	return result
}

func (f *Default) formatSection(w io.Writer, cells []string, ctx Formatting) {
	var renderedCells []string
	for i := 0; i < len(ctx.Widths); i++ {
		content := ""
		if i < len(cells) {
			content = cells[i]
		}
		padding := ctx.Padding
		if customPad, ok := ctx.ColPadding[i]; ok {
			padding = customPad
		}
		// Use AlignLeft for rows as per defaultConfig
		align := AlignLeft
		if colAlign, ok := ctx.ColAligns[i]; ok && colAlign != "" {
			align = colAlign
		}
		cell := f.formatCell(content, ctx.Widths[i], padding, align)
		renderedCells = append(renderedCells, cell)
	}
	f.renderLine(w, renderedCells, ctx)
}

func NewDefault(configs ...DefaultConfig) *Default {
	cfg := defaultConfig()
	cfg.debug = false

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
