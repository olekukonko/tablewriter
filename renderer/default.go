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

// Cond allows additional conditions for enabling, avoiding long if statements
func (o State) Cond(c func() bool) bool {
	if o.Enabled() {
		return c()
	}
	return false
}

// Or allows additional conditions for enabling, avoiding long if statements
func (o State) Or(c State) State {
	if o.Enabled() {
		return o
	}
	return c
}

// Or allows additional conditions for enabling, avoiding long if statements
func (o State) String() string {
	if o.Enabled() {
		return "on"
	}
	return "off"
}

type Settings struct {
	HeaderLine          State // Horizontal line below header (e.g., ├────┤)
	HeaderSeparator     State // Horizontal separator between header rows
	LineColumnSeparator State // Vertical separators between columns (headers, rows, footers)
	LineSeparator       State // Horizontal separators between rows
	FooterLine          State // Line below footer (if borders off)
	FooterSeparator     State // Vertical separators in footer row
}

// Border defines which borders to draw
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
}

type Default struct {
	config DefaultConfig
}

func (f *Default) FormatHeader(w io.Writer, headers []string, ctx Context) {
	// fmt.Println("DEBUG: FormatHeader called with headers =", headers, "HeaderLine =", f.config.Settings.HeaderLine, "HeaderSeparator =", f.config.Settings.HeaderSeparator)

	var cells []string
	for i, h := range headers {
		width := ctx.Widths[i]
		content := strings.TrimSpace(h)
		padding := ctx.Padding
		if customPad, ok := ctx.ColPadding[i]; ok {
			padding = customPad
		}
		colAlign := ""
		if i < len(ctx.ColAligns) && ctx.ColAligns[i] != "" {
			colAlign = ctx.ColAligns[i]
		}
		cells = append(cells, f.formatCell(content, width, ctx.Align, colAlign, padding))
	}
	// fmt.Println("DEBUG: Rendering header cells =", cells)
	f.renderLine(w, cells, ctx)

	// For now, assume single header row; HeaderSeparator would apply between multiple header rows
	// If multi-row headers are added later, this would loop over rows and use HeaderSeparator

	if f.config.Settings.HeaderLine.Enabled() {
		// fmt.Println("DEBUG: Rendering header line below")
		f.FormatLine(w, Context{Widths: ctx.Widths, Level: Middle})
	} else {
		// fmt.Println("DEBUG: HeaderLine disabled, skipping line below")
	}
}

func (f *Default) renderLine(w io.Writer, cells []string, ctx Context) {
	// fmt.Println("DEBUG: renderLine called with cells =", cells, "Level =", ctx.Level)
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
	if f.config.Settings.LineColumnSeparator.Enabled() {
		separator := f.config.Symbols.Column()
		// fmt.Println("DEBUG: Using separator =", separator)
		output = fmt.Sprintf("%s%s%s%s", prefix, strings.Join(cells, separator), suffix, symbols.NewLine)
	} else {
		output = fmt.Sprintf("%s%s%s%s", prefix, strings.Join(cells, ""), suffix, symbols.NewLine)
	}

	// fmt.Println("DEBUG: renderLine output =", output)
	fmt.Fprint(w, output)
}

func (f *Default) FormatLine(w io.Writer, ctx Context) {
	// fmt.Println("DEBUG: FormatLine called with Level =", ctx.Level, "b(top) = ", f.config.Borders.Top)

	if ctx.Level == Top && f.config.Borders.Top.Disabled() {
		// fmt.Println("DEBUG: Skipping top border")
		return
	}
	if ctx.Level == Bottom && f.config.Borders.Bottom.Disabled() {
		// fmt.Println("DEBUG: Skipping bottom border")
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
	if f.config.Settings.LineColumnSeparator.Enabled() {
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

	// fmt.Println("DEBUG: FormatLine output =", line.String())
	fmt.Fprintln(w, line.String())
}

func (f *Default) FormatRow(w io.Writer, row []string, ctx Context) {
	f.formatSection(w, row, ctx, false)
	if f.config.Settings.LineSeparator.Enabled() && !ctx.Last {
		// fmt.Println("DEBUG: Rendering row  separator")
		f.FormatLine(w, Context{Widths: ctx.Widths, Level: Middle})
	}
}

func (f *Default) formatSection(w io.Writer, cells []string, ctx Context, isFooter bool) {
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
			if colAlign, ok := ctx.ColAligns[i]; ok {
				align = colAlign
			}

			renderedCells = append(renderedCells, f.formatCell(content, width, align, "", padding))
		}
		f.renderLine(w, renderedCells, ctx)
	}
}

func (f *Default) formatCell(content string, width int, defaultAlign string, colAlign string, padding symbols.Padding) string {
	align := defaultAlign
	if colAlign != "" {
		align = colAlign
	}

	content = strings.TrimSpace(content)
	runeWidth := utils.RuneWidth(content)
	padLeftWidth := utils.RuneWidth(padding.Left)
	padRightWidth := utils.RuneWidth(padding.Right)
	totalContentWidth := runeWidth + padLeftWidth + padRightWidth

	if totalContentWidth > width {
		width = totalContentWidth
	}

	gap := width - runeWidth - padLeftWidth - padRightWidth
	if gap < 0 {
		gap = 0
	}

	leftPadChar := " "
	if padding.Left != "" {
		leftPadChar = padding.Left
	}
	rightPadChar := " "
	if padding.Right != "" {
		rightPadChar = padding.Right
	}

	switch align {
	case AlignCenter:
		leftGap := gap / 2
		rightGap := gap - leftGap
		return padding.Left + strings.Repeat(leftPadChar, leftGap) + content + strings.Repeat(rightPadChar, rightGap) + padding.Right
	case AlignRight:
		return padding.Left + strings.Repeat(leftPadChar, gap) + content + padding.Right
	default:
		return padding.Left + content + strings.Repeat(rightPadChar, gap) + padding.Right
	}
}

func (f *Default) FormatFooter(w io.Writer, footers []string, ctx Context) {
	f.formatSection(w, footers, ctx, true)
}

func (f *Default) GetColumnWidths() []int {
	return nil
}

func defaultConfig() DefaultConfig {
	return DefaultConfig{
		Borders: Border{Left: On, Right: On, Top: On, Bottom: On},
		Settings: Settings{
			HeaderLine:          On,  // Header row and horizontal separator enabled
			HeaderSeparator:     On,  // Enable horizontal row separators by default
			LineColumnSeparator: On,  // Vertical separators across all sections
			LineSeparator:       Off, // No horizontal row separators by default
			FooterLine:          On,  // Footer line enabled
			FooterSeparator:     On,  // Vertical separators in footer
		},
	}
}

func NewDefault(config ...DefaultConfig) *Default {
	cfg := defaultConfig()

	if len(config) > 0 {
		userCfg := config[0]

		// Merge Borders - directly apply user settings if provided
		if userCfg.Borders != (Border{}) {
			if userCfg.Borders != (Border{}) {
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
			}
		}

		// Merge Symbols
		if userCfg.Symbols != nil {
			cfg.Symbols = userCfg.Symbols
		}

		// Merge Settings
		if userCfg.Settings != (Settings{}) {
			if userCfg.Settings.HeaderLine != 0 {
				cfg.Settings.HeaderLine = userCfg.Settings.HeaderLine
			}
			if userCfg.Settings.HeaderSeparator != 0 {
				cfg.Settings.HeaderSeparator = userCfg.Settings.HeaderSeparator
			}
			if userCfg.Settings.LineColumnSeparator != 0 {
				cfg.Settings.LineColumnSeparator = userCfg.Settings.LineColumnSeparator
			}
			if userCfg.Settings.LineSeparator != 0 {
				cfg.Settings.LineSeparator = userCfg.Settings.LineSeparator
			}
			if userCfg.Settings.FooterLine != 0 {
				cfg.Settings.FooterLine = userCfg.Settings.FooterLine
			}
			if userCfg.Settings.FooterSeparator != 0 {
				cfg.Settings.FooterSeparator = userCfg.Settings.FooterSeparator
			}
		}
	}

	if cfg.Symbols == nil {
		cfg.Symbols = symbols.NewSymbols(symbols.StyleLight)
	}

	return &Default{config: cfg}
}

func (f *Default) Reset() {}

func (f *Default) Config() DefaultConfig {
	return f.config
}
