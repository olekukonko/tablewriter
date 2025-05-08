package tests

import (
	"bytes"
	"testing"

	"github.com/fatih/color"
	"github.com/olekukonko/tablewriter"
	"github.com/olekukonko/tablewriter/renderer"
	"github.com/olekukonko/tablewriter/tw"
)

func TestColorizedBasicTable(t *testing.T) {
	var buf bytes.Buffer
	table := tablewriter.NewTable(&buf,
		tablewriter.WithRenderer(renderer.NewColorized()),
	)
	table.SetHeader([]string{"Name", "Age", "City"})
	table.Append([]string{"Alice", "25", "New York"})
	table.Append([]string{"Bob", "30", "Boston"})
	table.Render()

	// Expected colors: Headers (white/bold on black), Rows (cyan on black), Borders/Separators (white on black)
	expected := `
┌───────┬─────┬──────────┐
│ NAME  │ AGE │   CITY   │
├───────┼─────┼──────────┤
│ Alice │ 25  │ New York │
│ Bob   │ 30  │ Boston   │
└───────┴─────┴──────────┘
`
	visualCheck(t, "ColorizedBasicTable", buf.String(), expected)

}

func TestColorizedNoBorders(t *testing.T) {
	var buf bytes.Buffer
	table := tablewriter.NewTable(&buf,
		tablewriter.WithRenderer(renderer.NewColorized(renderer.ColorizedConfig{
			Borders: tw.Border{Left: tw.Off, Right: tw.Off, Top: tw.Off, Bottom: tw.Off},
		})),
	)
	table.SetHeader([]string{"Name", "Age", "City"})
	table.Append([]string{"Alice", "25", "New York"})
	table.Append([]string{"Bob", "30", "Boston"})
	table.Render()

	// Expected colors: Headers (white/bold on black), Rows (cyan on black), Separators (white on black)
	expected := `
 NAME  │ AGE │   CITY   
───────┼─────┼──────────
 Alice │ 25  │ New York 
 Bob   │ 30  │ Boston   
`
	visualCheck(t, "ColorizedNoBorders", buf.String(), expected)
}

func TestColorizedCustomColors(t *testing.T) {
	var buf bytes.Buffer
	table := tablewriter.NewTable(&buf,
		tablewriter.WithRenderer(renderer.NewColorized(renderer.ColorizedConfig{
			Header: renderer.Tint{
				FG: renderer.Colors{color.FgGreen, color.Bold},
				BG: renderer.Colors{color.BgBlue},
				Columns: []renderer.Tint{
					{FG: renderer.Colors{color.FgRed}, BG: renderer.Colors{color.BgBlue}},
					{FG: renderer.Colors{color.FgYellow}, BG: renderer.Colors{color.BgBlue}},
				},
			},
			Column: renderer.Tint{
				FG: renderer.Colors{color.FgBlue},
				BG: renderer.Colors{color.BgBlack},
				Columns: []renderer.Tint{
					{FG: renderer.Colors{color.FgMagenta}, BG: renderer.Colors{color.BgBlack}},
				},
			},
			Footer: renderer.Tint{
				FG: renderer.Colors{color.FgYellow},
				BG: renderer.Colors{color.BgBlue},
			},
			Border: renderer.Tint{
				FG: renderer.Colors{color.FgWhite},
				BG: renderer.Colors{color.BgBlue},
			},
			Separator: renderer.Tint{
				FG: renderer.Colors{color.FgWhite},
				BG: renderer.Colors{color.BgBlue},
			},
		})),
		tablewriter.WithFooterConfig(tablewriter.CellConfig{
			ColumnAligns: []tw.Align{tw.AlignRight, tw.AlignCenter}, // Align "1" to center
		}),
	)
	table.SetHeader([]string{"Name", "Age"})
	table.Append([]string{"Alice", "25"})
	table.SetFooter([]string{"Total", "1"})
	table.Render()

	// Expected colors: Headers (red, yellow on blue), Rows (magenta, blue on black), Footers (yellow on blue), Borders/Separators (white on blue)
	expected := `
┌───────┬─────┐
│ NAME  │ AGE │
├───────┼─────┤
│ Alice │ 25  │
├───────┼─────┤
│ Total │  1  │
└───────┴─────┘
`
	if !visualCheck(t, "ColorizedCustomColors", buf.String(), expected) {
		for _, v := range table.Debug() {
			t.Error(v)
		}
	}
}

func TestColorizedLongValues(t *testing.T) {
	var buf bytes.Buffer
	c := tablewriter.Config{
		Row: tablewriter.CellConfig{
			Formatting: tablewriter.CellFormatting{
				MaxWidth:  20,
				AutoWrap:  tw.WrapNormal,
				Alignment: tw.AlignLeft,
			},
		},
	}
	table := tablewriter.NewTable(&buf,
		tablewriter.WithConfig(c),
		tablewriter.WithRenderer(renderer.NewColorized()),
	)
	table.SetHeader([]string{"No", "Description", "Note"})
	table.Append([]string{"1", "This is a very long description that should wrap", "Short"})
	table.Append([]string{"2", "Short desc", "Another note"})
	table.Render()

	// Expected colors: Headers (white/bold on black), Rows (cyan on black), Borders/Separators (white on black)
	expected := `
        ┌────┬──────────────────┬──────────────┐
        │ NO │   DESCRIPTION    │     NOTE     │
        ├────┼──────────────────┼──────────────┤
        │ 1  │ This is a very   │ Short        │
        │    │ long description │              │
        │    │ that should wrap │              │
        │ 2  │ Short desc       │ Another note │
        └────┴──────────────────┴──────────────┘
`
	visualCheck(t, "ColorizedLongValues", buf.String(), expected)
}

func TestColorizedHorizontalMerge(t *testing.T) {
	var buf bytes.Buffer
	c := tablewriter.Config{
		Header: tablewriter.CellConfig{
			Formatting: tablewriter.CellFormatting{
				MergeMode: tw.MergeHorizontal,
			},
		},
		Row: tablewriter.CellConfig{
			Formatting: tablewriter.CellFormatting{
				MergeMode: tw.MergeHorizontal,
			},
		},
	}
	table := tablewriter.NewTable(&buf,
		tablewriter.WithConfig(c),
		tablewriter.WithRenderer(renderer.NewColorized()),
	)
	table.SetHeader([]string{"Merged", "Merged", "Normal"})
	table.Append([]string{"Same", "Same", "Unique"})
	table.Render()

	// Expected colors: Headers (white/bold on black), Rows (cyan on black), Borders/Separators (white on black)
	expected := `
        ┌─────────────────┬────────┐
        │     MERGED      │ NORMAL │
        ├─────────────────┼────────┤
        │ Same            │ Unique │
        └─────────────────┴────────┘
`
	if !visualCheck(t, "ColorizedHorizontalMerge", buf.String(), expected) {
		for _, v := range table.Debug() {
			t.Error(v)
		}
	}
}
