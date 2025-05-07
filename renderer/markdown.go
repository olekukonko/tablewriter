package renderer

import (
	"fmt"
	"io"
	"strings"

	"github.com/olekukonko/tablewriter/tw"
	"github.com/olekukonko/tablewriter/twfn"
)

// Markdown implements a Markdown table formatter.
type Markdown struct {
	config DefaultConfig // Configuration for rendering
	trace  []string      // Debug trace messages
}

// NewMarkdown creates a new Markdown renderer with default settings.
func NewMarkdown(configs ...DefaultConfig) *Markdown {
	cfg := defaultConfig()
	cfg.Symbols = tw.NewSymbols(tw.StyleMarkdown)
	cfg.Borders = Border{Left: tw.On, Right: tw.On, Top: tw.Off, Bottom: tw.Off}
	cfg.Settings.Separators.BetweenColumns = tw.On
	cfg.Settings.Lines.ShowHeaderLine = tw.On
	cfg.Settings.Lines.ShowTop = tw.Off
	cfg.Settings.Lines.ShowBottom = tw.Off
	if len(configs) > 0 {
		userCfg := configs[0]
		cfg = mergeConfig(cfg, userCfg)
	}
	return &Markdown{config: cfg}
}

// mergeConfig merges user-provided config with defaults, preserving non-zero values.
func mergeConfig(defaults, overrides DefaultConfig) DefaultConfig {
	if overrides.Borders.Left != 0 {
		defaults.Borders.Left = overrides.Borders.Left
	}
	if overrides.Borders.Right != 0 {
		defaults.Borders.Right = overrides.Borders.Right
	}
	if overrides.Borders.Top != 0 {
		defaults.Borders.Top = overrides.Borders.Top
	}
	if overrides.Borders.Bottom != 0 {
		defaults.Borders.Bottom = overrides.Borders.Bottom
	}
	if overrides.Symbols != nil {
		defaults.Symbols = overrides.Symbols
	}
	defaults.Settings = mergeSettings(defaults.Settings, overrides.Settings)
	return defaults
}

// Config returns the current renderer configuration.
func (m *Markdown) Config() DefaultConfig {
	return m.config
}

// debug logs a debug message if debugging is enabled.
func (m *Markdown) debug(format string, a ...interface{}) {
	if m.config.Debug {
		msg := fmt.Sprintf(format, a...)
		m.trace = append(m.trace, fmt.Sprintf("[MARKDOWN] %s", msg))
	}
}

// Debug returns the accumulated debug trace.
func (m *Markdown) Debug() []string {
	return m.trace
}

// formatCell formats a cell's content according to width, padding, and alignment.
func (m *Markdown) formatCell(content string, width int, align tw.Align, padding tw.Padding) string {
	if m.config.Settings.TrimWhitespace.Enabled() {
		content = strings.TrimSpace(content)
	}
	m.debug("Formatting cell: content='%s', width=%d, align=%s", content, width, align)

	runeWidth := twfn.DisplayWidth(content)
	padLeftWidth := twfn.DisplayWidth(padding.Left)
	padRightWidth := twfn.DisplayWidth(padding.Right)
	totalContentWidth := runeWidth + padLeftWidth + padRightWidth

	if totalContentWidth > width && width > 0 {
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

	var result strings.Builder
	switch align {
	case tw.AlignCenter:
		leftGap := gap / 2
		rightGap := gap - leftGap
		result.WriteString(padding.Left)
		result.WriteString(strings.Repeat(leftPadChar, leftGap))
		result.WriteString(content)
		result.WriteString(strings.Repeat(rightPadChar, rightGap))
		result.WriteString(padding.Right)
	case tw.AlignRight:
		result.WriteString(padding.Left)
		result.WriteString(strings.Repeat(leftPadChar, gap))
		result.WriteString(content)
		result.WriteString(padding.Right)
	default: // Left or default
		result.WriteString(padding.Left)
		result.WriteString(content)
		result.WriteString(strings.Repeat(rightPadChar, gap))
		result.WriteString(padding.Right)
	}

	output := result.String()
	m.debug("Formatted cell result: '%s' (width %d)", output, twfn.DisplayWidth(output))
	return output
}

// Header renders the table header section with Markdown formatting.
func (m *Markdown) Header(w io.Writer, headers [][]string, ctx Formatting) {
	if len(headers) == 0 || len(headers[0]) == 0 {
		m.debug("Header: No headers to render")
		return
	}

	m.debug("Rendering header with %d lines, widths=%v", len(headers), ctx.Row.Widths)
	numCols := len(ctx.Row.Current)
	headerLine := headers[0]

	// Render header row
	cells := make([]string, numCols)
	for i := 0; i < numCols; i++ {
		content := ""
		if i < len(headerLine) {
			content = headerLine[i]
		}
		cellCtx, ok := ctx.Row.Current[i]
		if !ok {
			align := tw.AlignLeft
			if ctx.Row.Position == tw.Header {
				align = tw.AlignCenter
			}
			cellCtx = CellContext{
				Data:    content,
				Align:   align,
				Padding: tw.Padding{Left: " ", Right: " "},
				Width:   ctx.Row.Widths.Get(i),
			}
		}
		if cellCtx.Merge.Horizontal.Present && !cellCtx.Merge.Horizontal.Start {
			cells[i] = ""
			continue
		}
		width := ctx.Row.Widths.Get(i)
		if colMaxWidth, ok := ctx.Row.ColMaxWidths[i]; ok && colMaxWidth > 0 && colMaxWidth < width {
			width = colMaxWidth
		}
		cells[i] = m.formatCell(content, width, cellCtx.Align, cellCtx.Padding)
	}
	prefix := m.config.Symbols.HeaderLeft()
	suffix := m.config.Symbols.HeaderRight()
	fmt.Fprintf(w, "%s%s%s%s", prefix, strings.Join(cells, m.config.Symbols.HeaderMid()), suffix, tw.NewLine)

	// Render separator row
	if m.config.Settings.Lines.ShowHeaderLine.Enabled() {
		separators := make([]string, numCols)
		for i := 0; i < numCols; {
			cellCtx, ok := ctx.Row.Current[i]
			width := ctx.Row.Widths.Get(i)
			if ok && cellCtx.Merge.Horizontal.Present && cellCtx.Merge.Horizontal.Start {
				span := cellCtx.Merge.Horizontal.Span
				totalWidth := width
				for j := 1; j < span && i+j < numCols; j++ {
					totalWidth += ctx.Row.Widths.Get(i + j)
					if m.config.Settings.Separators.BetweenColumns.Enabled() {
						totalWidth += twfn.DisplayWidth(m.config.Symbols.Column())
					}
				}
				separators[i] = m.formatSeparator(totalWidth, cellCtx.Align)
				for j := 1; j < span && i+j < numCols; j++ {
					separators[i+j] = ""
				}
				i += span
			} else {
				separators[i] = m.formatSeparator(width, cellCtx.Align)
				i++
			}
		}
		fmt.Fprintf(w, "%s%s%s%s", prefix, strings.Join(separators, m.config.Symbols.HeaderMid()), suffix, tw.NewLine)
	}
}

// formatSeparator generates a Markdown separator string based on alignment and width.
func (m *Markdown) formatSeparator(width int, align tw.Align) string {
	if width <= 0 {
		return "---"
	}
	// Ensure minimum width for separator (at least 3 dashes)
	if width < 3 {
		width = 3
	}
	switch align {
	case tw.AlignCenter:
		return ":" + strings.Repeat("-", width-2) + ":"
	case tw.AlignRight:
		return strings.Repeat("-", width-1) + ":"
	default: // Left or default
		return ":" + strings.Repeat("-", width-1)
	}
}

// Row renders a table row with Markdown formatting.
func (m *Markdown) Row(w io.Writer, row []string, ctx Formatting) {
	if len(row) == 0 {
		m.debug("Row: No data to render")
		return
	}

	m.debug("Rendering row with data=%v, widths=%v", row, ctx.Row.Widths)
	numCols := len(ctx.Row.Current)
	cells := make([]string, numCols)
	for i := 0; i < numCols; i++ {
		content := ""
		if i < len(row) {
			content = row[i]
		}
		cellCtx, ok := ctx.Row.Current[i]
		if !ok {
			cellCtx = CellContext{
				Data:    content,
				Align:   tw.AlignLeft,
				Padding: tw.Padding{Left: " ", Right: " "},
				Width:   ctx.Row.Widths.Get(i),
			}
		}
		if cellCtx.Merge.Horizontal.Present && !cellCtx.Merge.Horizontal.Start {
			cells[i] = ""
			continue
		}
		width := ctx.Row.Widths.Get(i)
		if colMaxWidth, ok := ctx.Row.ColMaxWidths[i]; ok && colMaxWidth > 0 && colMaxWidth < width {
			width = colMaxWidth
		}
		cells[i] = m.formatCell(content, width, cellCtx.Align, cellCtx.Padding)
	}
	prefix := m.config.Symbols.MidLeft()
	suffix := m.config.Symbols.MidRight()
	fmt.Fprintf(w, "%s%s%s%s", prefix, strings.Join(cells, m.config.Symbols.Column()), suffix, tw.NewLine)
}

// Footer renders the table footer section with Markdown formatting.
func (m *Markdown) Footer(w io.Writer, footers [][]string, ctx Formatting) {
	if len(footers) == 0 || len(footers[0]) == 0 {
		m.debug("Footer: No footers to render")
		return
	}

	m.debug("Rendering footer with %d lines, widths=%v", len(footers), ctx.Row.Widths)
	numCols := len(ctx.Row.Current)
	footerLine := footers[0]
	cells := make([]string, numCols)
	for i := 0; i < numCols; i++ {
		content := ""
		if i < len(footerLine) {
			content = footerLine[i]
		}
		cellCtx, ok := ctx.Row.Current[i]
		if !ok {
			cellCtx = CellContext{
				Data:    content,
				Align:   tw.AlignRight,
				Padding: tw.Padding{Left: " ", Right: " "},
				Width:   ctx.Row.Widths.Get(i),
			}
		}
		if cellCtx.Merge.Horizontal.Present && !cellCtx.Merge.Horizontal.Start {
			cells[i] = ""
			continue
		}
		width := ctx.Row.Widths.Get(i)
		if colMaxWidth, ok := ctx.Row.ColMaxWidths[i]; ok && colMaxWidth > 0 && colMaxWidth < width {
			width = colMaxWidth
		}
		cells[i] = m.formatCell(content, width, cellCtx.Align, cellCtx.Padding)
	}
	prefix := m.config.Symbols.MidLeft()
	suffix := m.config.Symbols.MidRight()
	fmt.Fprintf(w, "%s%s%s%s", prefix, strings.Join(cells, m.config.Symbols.Column()), suffix, tw.NewLine)
}

// Line renders a separator line (not used in Markdown tables).
func (m *Markdown) Line(w io.Writer, ctx Formatting) {
	m.debug("Line: Markdown tables do not render separator lines")
	// Markdown tables do not use additional separator lines beyond header separator
}

// Reset clears any internal state (no-op for Markdown).
func (m *Markdown) Reset() {
	m.trace = nil
	m.debug("Reset: Cleared debug trace")
}
