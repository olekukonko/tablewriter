// renderer/default.go
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

	for _, line := range headers {
		f.renderLine(w, line, ctx)
	}

	if f.config.Settings.Lines.ShowHeaderLine.Enabled() && len(headers) > 0 {
		f.Line(w, Formatting{Widths: ctx.Widths, Level: Middle, Position: Header})
	}
}

func (f *Default) Row(w io.Writer, row []string, ctx Formatting) {
	f.renderLine(w, row, ctx)

	if ctx.IsLast && !ctx.HasFooter && f.config.Borders.Bottom.Enabled() {
		f.Line(w, Formatting{Widths: ctx.Widths, Level: Bottom, Position: Row})
	} else if f.config.Settings.Separators.BetweenRows.Enabled() && !ctx.IsLast {
		f.Line(w, Formatting{Widths: ctx.Widths, Level: Middle, Position: Row})
	}
}

func (f *Default) Footer(w io.Writer, footers [][]string, ctx Formatting) {
	if f.config.Settings.Lines.ShowFooterLine.Enabled() && len(footers) > 0 {
		f.Line(w, Formatting{Widths: ctx.Widths, Level: Middle, Position: Footer})
	}

	for _, line := range footers {
		f.renderLine(w, line, ctx)
	}

	if ctx.HasFooter && f.config.Borders.Bottom.Enabled() {
		f.Line(w, Formatting{Widths: ctx.Widths, Level: Bottom, Position: Footer})
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

	formattedCells := make([]string, len(cells))
	for i, content := range cells {
		padding := ctx.Padding
		if customPad, ok := ctx.ColPadding[i]; ok {
			padding = customPad
		}
		align := ctx.Align
		if colAlign, ok := ctx.ColAligns[i]; ok && colAlign != "" {
			align = colAlign
		}
		formattedCells[i] = f.formatCell(content, ctx.Widths[i], padding, align)
	}

	var output string
	if f.config.Settings.Separators.BetweenColumns.Enabled() {
		separator := f.config.Symbols.Column()
		output = fmt.Sprintf("%s%s%s%s", prefix, strings.Join(formattedCells, separator), suffix, symbols.NewLine)
	} else {
		output = fmt.Sprintf("%s%s%s%s", prefix, strings.Join(formattedCells, ""), suffix, symbols.NewLine)
	}
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
	fmt.Fprintln(w, line.String())
}

func (f *Default) formatCell(content string, width int, padding symbols.Padding, align string) string {
	content = strings.TrimSpace(content)
	runeWidth := utils.RuneWidth(content)
	padLeftWidth := utils.RuneWidth(padding.Left)
	padRightWidth := utils.RuneWidth(padding.Right)
	totalPadding := padLeftWidth + padRightWidth

	availableWidth := width - totalPadding
	if availableWidth < 0 {
		availableWidth = 0
	}

	if runeWidth > availableWidth && availableWidth > 0 {
		content = utils.TruncateString(content, availableWidth)
		runeWidth = utils.RuneWidth(content)
	}

	var builder strings.Builder
	remaining := width - runeWidth - totalPadding

	switch align {
	case AlignLeft:
		builder.WriteString(padding.Left)
		builder.WriteString(content)
		if remaining > 0 {
			builder.WriteString(strings.Repeat(padding.Right, remaining))
		}
		builder.WriteString(padding.Right)
	case AlignRight:
		if remaining > 0 {
			builder.WriteString(strings.Repeat(padding.Left, remaining))
		}
		builder.WriteString(padding.Left)
		builder.WriteString(content)
		builder.WriteString(padding.Right)
	case AlignCenter:
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
	default:
		builder.WriteString(padding.Left)
		builder.WriteString(content)
		if remaining > 0 {
			builder.WriteString(strings.Repeat(padding.Left, remaining))
		}
		builder.WriteString(padding.Right)
	}
	return builder.String()
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
