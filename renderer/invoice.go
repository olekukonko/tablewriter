package renderer

import (
	"fmt"
	"io"
	"strings"

	"github.com/olekukonko/tablewriter/symbols"
	"github.com/olekukonko/tablewriter/utils"
)

type InvoiceConfig struct {
	Symbols symbols.Symbols
}

type Invoice struct {
	config InvoiceConfig
}

func NewInvoice(config ...InvoiceConfig) *Invoice {
	cfg := InvoiceConfig{
		Symbols: symbols.NewSymbols(symbols.StyleASCII),
	}
	if len(config) > 0 && config[0].Symbols != nil {
		cfg.Symbols = config[0].Symbols
	}
	return &Invoice{config: cfg}
}

func (i *Invoice) formatCell(content string, width int) string {
	content = strings.TrimSpace(content)
	if content == "" {
		return strings.Repeat(" ", width)
	}
	if utils.RuneWidth(content) > width {
		width = utils.RuneWidth(content)
	}
	return fmt.Sprintf("%-*s", width, content)
}

func (i *Invoice) Header(w io.Writer, headers []string, ctx Formatting) {
	cells := make([]string, len(headers))
	for j, h := range headers {
		cells[j] = i.formatCell(h, ctx.Widths[j])
	}
	fmt.Fprintf(w, "    %s %s", strings.Join(cells, " "+i.config.Symbols.Column()+" "), symbols.NewLine)
	i.Line(w, ctx)
}

func (i *Invoice) Row(w io.Writer, row []string, ctx Formatting) {
	maxLines := 1
	splitCells := make([][]string, len(row))
	for j, cell := range row {
		lines := strings.Split(cell, "\n")
		splitCells[j] = lines
		if len(lines) > maxLines {
			maxLines = len(lines)
		}
	}

	for lineIdx := 0; lineIdx < maxLines; lineIdx++ {
		cells := make([]string, len(row))
		for j := range row {
			content := ""
			if lineIdx < len(splitCells[j]) {
				content = splitCells[j][lineIdx]
			}
			cells[j] = i.formatCell(content, ctx.Widths[j])
		}
		fmt.Fprintf(w, "    %s %s", strings.Join(cells, " "+i.config.Symbols.Column()+" "), symbols.NewLine)
	}
}

func (i *Invoice) Footer(w io.Writer, footers []string, ctx Formatting) {
	if len(footers) == 0 {
		return
	}

	cells := make([]string, len(ctx.Widths))
	hasContent := false
	for j := 0; j < len(ctx.Widths) && j < len(footers); j++ {
		cells[j] = i.formatCell(footers[j], ctx.Widths[j])
		if cells[j] != strings.Repeat(" ", ctx.Widths[j]) {
			hasContent = true
		}
	}

	if hasContent {
		i.Line(w, ctx)
		fmt.Fprintf(w, "    %s %s", strings.Join(cells, " "+i.config.Symbols.Column()+" "), symbols.NewLine)
	}
}

func (i *Invoice) Line(w io.Writer, ctx Formatting) {
	separators := make([]string, len(ctx.Widths))
	for j := range ctx.Widths {
		separators[j] = strings.Repeat(i.config.Symbols.Row(), ctx.Widths[j])
	}
	fmt.Fprintf(w, "    %s%s", strings.Join(separators, i.config.Symbols.Center()), symbols.NewLine)
}

func (i *Invoice) GetColumnWidths() []int { return nil }
func (i *Invoice) Reset()                 {}
