package tests

import (
	"bytes"
	"testing"

	"github.com/olekukonko/tablewriter"
	"github.com/olekukonko/tablewriter/renderer"
	"github.com/olekukonko/tablewriter/tw"
)

func TestMarkdownBasicTable(t *testing.T) {
	var buf bytes.Buffer
	table := tablewriter.NewTable(&buf,
		tablewriter.WithRenderer(renderer.NewMarkdown()),
	)
	table.Header([]string{"Name", "Age", "City"})
	table.Append([]string{"Alice", "25", "New York"})
	table.Append([]string{"Bob", "30", "Boston"})
	table.Render()

	expected := `
	| NAME  | AGE |   CITY   |
	|:-----:|:---:|:--------:|
	| Alice | 25  | New York |
	|  Bob  | 30  |  Boston  |
`
	if !visualCheck(t, "MarkdownBasicTable", buf.String(), expected) {
		t.Error(table.Debug())
	}
}

func TestMarkdownAlignment(t *testing.T) {
	var buf bytes.Buffer
	table := tablewriter.NewTable(&buf,
		tablewriter.WithRenderer(renderer.NewMarkdown()),
		tablewriter.WithConfig(tablewriter.Config{
			Header: tw.CellConfig{
				Alignment: tw.CellAlignment{PerColumn: []tw.Align{tw.AlignLeft, tw.AlignRight, tw.AlignRight, tw.AlignCenter}},
			},
		}),
	)
	table.Header([]string{"Name", "Age", "City", "Status"})
	table.Append([]string{"Alice", "25", "New York", "OK"})
	table.Append([]string{"Bob", "30", "Boston", "ERROR"})
	table.Render()

	expected := `
	| NAME  | AGE |     CITY | STATUS |
	|:------|----:|---------:|:------:|
	| Alice |  25 | New York |   OK   |
	| Bob   |  30 |   Boston | ERROR  |
`
	if !visualCheck(t, "MarkdownBasicTable", buf.String(), expected) {
		t.Error(table.Debug())
	}
}

func TestMarkdownNoBorders(t *testing.T) {
	var buf bytes.Buffer
	table := tablewriter.NewTable(&buf,
		tablewriter.WithRenderer(renderer.NewMarkdown(tw.Rendition{
			Borders: tw.Border{Left: tw.Off, Right: tw.Off, Top: tw.Off, Bottom: tw.Off},
		})),
		tablewriter.WithConfig(tablewriter.Config{
			Header: tw.CellConfig{
				Alignment: tw.CellAlignment{PerColumn: []tw.Align{tw.AlignLeft}},
			},
		}),
	)

	table.Header([]string{"Name", "Age", "City"})
	table.Append([]string{"Alice", "25", "New York"})
	table.Append([]string{"Bob", "30", "Boston"})
	table.Render()

	expected := `
 NAME  | AGE |   CITY   
:------|:---:|:--------:
 Alice | 25  | New York 
 Bob   | 30  |  Boston   
`
	visualCheck(t, "MarkdownNoBorders", buf.String(), expected)
}

func TestMarkdownUnicode(t *testing.T) {
	var buf bytes.Buffer
	table := tablewriter.NewTable(&buf,
		tablewriter.WithRenderer(renderer.NewMarkdown()),
	)
	table.Header([]string{"Name", "Age", "City"})
	table.Append([]string{"Bøb", "30", "Tōkyō"})
	table.Append([]string{"José", "28", "México"})
	table.Append([]string{"张三", "35", "北京"})
	table.Render()

	expected := `
		| NAME | AGE |  CITY  |
		|:----:|:---:|:------:|
		| Bøb  | 30  | Tōkyō  |
		| José | 28  | México |
		| 张三 | 35  |  北京  |
`
	visualCheck(t, "MarkdownUnicode", buf.String(), expected)
}

func TestMarkdownLongHeaders(t *testing.T) {
	var buf bytes.Buffer
	c := tablewriter.Config{
		Header: tw.CellConfig{
			Formatting: tw.CellFormatting{
				AutoWrap: tw.WrapTruncate,
			},
			ColMaxWidths: tw.CellWidth{Global: 20},
		},
	}
	table := tablewriter.NewTable(&buf,
		tablewriter.WithConfig(c),
		tablewriter.WithRenderer(renderer.NewMarkdown()),
		tablewriter.WithAlignment(tw.MakeAlign(3, tw.AlignLeft)),
	)
	table.Header([]string{"Name", "Age", "Very Long Header That Needs Truncation"})
	table.Append([]string{"Alice", "25", "New York"})
	table.Append([]string{"Bob", "30", "Boston"})
	table.Render()

	expected := `
        | NAME  | AGE | VERY LONG HEADER… |
        |:------|:----|:------------------|
        | Alice | 25  | New York          |
        | Bob   | 30  | Boston            |
`
	visualCheck(t, "MarkdownLongHeaders", buf.String(), expected)
}

func TestMarkdownLongValues(t *testing.T) {
	var buf bytes.Buffer
	c := tablewriter.Config{
		Row: tw.CellConfig{
			Formatting: tw.CellFormatting{
				AutoWrap: tw.WrapNormal,
			},
			Alignment:    tw.CellAlignment{Global: tw.AlignLeft},
			ColMaxWidths: tw.CellWidth{Global: 20},
		},
	}
	table := tablewriter.NewTable(&buf,
		tablewriter.WithConfig(c),
		tablewriter.WithRenderer(renderer.NewMarkdown()),
		tablewriter.WithAlignment(tw.MakeAlign(3, tw.AlignLeft)),
	)
	table.Header([]string{"No", "Description", "Note"})
	table.Append([]string{"1", "This is a very long description that should wrap", "Short"})
	table.Append([]string{"2", "Short desc", "Another note"})
	table.Render()

	expected := `
        | NO | DESCRIPTION      | NOTE         |
        |:---|:-----------------|:-------------|
        | 1  | This is a very   | Short        |
        |    | long description |              |
        |    | that should wrap |              |
        | 2  | Short desc       | Another note |
`
	visualCheck(t, "MarkdownLongValues", buf.String(), expected)
}

func TestMarkdownCustomPadding(t *testing.T) {
	var buf bytes.Buffer
	c := tablewriter.Config{
		Header: tw.CellConfig{
			Padding: tw.CellPadding{
				Global: tw.Padding{Left: "*", Right: "*", Top: "", Bottom: ""},
			},
		},
		Row: tw.CellConfig{
			Padding: tw.CellPadding{
				Global: tw.Padding{Left: ">", Right: "<", Top: "", Bottom: ""},
			},
		},
	}
	table := tablewriter.NewTable(&buf,
		tablewriter.WithConfig(c),
		tablewriter.WithRenderer(renderer.NewMarkdown()),
	)
	table.Header([]string{"Name", "Age", "City"})
	table.Append([]string{"Alice", "25", "New York"})
	table.Append([]string{"Bob", "30", "Boston"})
	table.Render()

	expected := `
        |*NAME**|*AGE*|***CITY***|
        |:-----:|:---:|:--------:|
        |>Alice<|>25<<|>New York<|
        |>>Bob<<|>30<<|>>Boston<<|
`
	visualCheck(t, "MarkdownCustomPadding", buf.String(), expected)
}

func TestMarkdownHorizontalMerge(t *testing.T) {
	var buf bytes.Buffer
	c := tablewriter.Config{
		Header: tw.CellConfig{
			Merging: tw.CellMerging{
				Mode: tw.MergeHorizontal,
			},
		},
		Row: tw.CellConfig{
			Merging: tw.CellMerging{
				Mode: tw.MergeHorizontal,
			},
		},
	}
	table := tablewriter.NewTable(&buf,
		tablewriter.WithConfig(c),
		tablewriter.WithRenderer(renderer.NewMarkdown()),
	)
	table.Header([]string{"Merged", "Merged", "Normal"})
	table.Append([]string{"Same", "Same", "Unique"})
	table.Render()

	expected := `
        |     MERGED      | NORMAL |
        |:---------------:|:------:|
        |      Same       | Unique |
`
	visualCheck(t, "MarkdownHorizontalMerge", buf.String(), expected)
}

func TestMarkdownEmptyTable(t *testing.T) {
	var buf bytes.Buffer
	table := tablewriter.NewTable(&buf,
		tablewriter.WithRenderer(renderer.NewMarkdown()),
	)
	table.Render()

	expected := ""
	visualCheck(t, "MarkdownEmptyTable", buf.String(), expected)
}

func TestMarkdownWithFooter(t *testing.T) {
	var buf bytes.Buffer
	c := tablewriter.Config{
		Footer: tw.CellConfig{
			Alignment: tw.CellAlignment{Global: tw.AlignRight},
		},
	}
	table := tablewriter.NewTable(&buf,
		tablewriter.WithConfig(c),
		tablewriter.WithRenderer(renderer.NewMarkdown()),
	)
	table.Header([]string{"Name", "Age", "City"})
	table.Append([]string{"Alice", "25", "New York"})
	table.Append([]string{"Bob", "30", "Boston"})
	table.Footer([]string{"Total", "2", ""})
	table.Render()

	expected := `
        | NAME  | AGE |   CITY   |
        |:-----:|:---:|:--------:|
        | Alice | 25  | New York |
        |  Bob  | 30  |  Boston  |
        | Total |  2  |          |
`
	visualCheck(t, "MarkdownWithFooter", buf.String(), expected)
}

func TestMarkdownAlignmentNone(t *testing.T) {
	t.Run("AlignNone", func(t *testing.T) {
		var buf bytes.Buffer
		table := tablewriter.NewTable(&buf, tablewriter.WithRenderer(renderer.NewMarkdown()))
		table.Configure(func(cfg *tablewriter.Config) {
			cfg.Header.Alignment.PerColumn = []tw.Align{tw.AlignNone}
			cfg.Row.Alignment.PerColumn = []tw.Align{tw.AlignNone}
			cfg.Debug = true
		})
		table.Header([]string{"Header"})
		table.Append([]string{"Data"})
		table.Render()

		expected := `
            | HEADER |
            |--------|
            | Data   |


`
		if !visualCheck(t, "AlignNone", buf.String(), expected) {
			t.Fatal(table.Debug())
		}
	})
}
