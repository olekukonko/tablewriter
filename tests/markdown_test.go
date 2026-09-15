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
|:------|:----|:---------|
| Alice | 25  | New York |
| Bob   | 30  | Boston   |
`
	if !visualCheck(t, "MarkdownBasicTable", buf.String(), expected) {
		t.Error(table.Debug())
	}
}

func TestMarkdownRule1NoExplicitAlignment(t *testing.T) {
	var buf bytes.Buffer
	table := tablewriter.NewTable(&buf,
		tablewriter.WithRenderer(renderer.NewMarkdown()),
		tablewriter.WithConfig(tablewriter.Config{
			Debug: true,
		}),
	)
	table.Header([]string{"Change", "Resource Type", "Resource Name", "Address"})
	table.Append([]string{"CREATE", "myresource", "foo", "myresource.foo"})
	table.Append([]string{"REPLACE", "myresource", "foobar", "myresource.foobar"})
	table.Render()

	expected := `
| CHANGE  | RESOURCE TYPE | RESOURCE NAME |      ADDRESS      |
|:--------|:--------------|:--------------|:------------------|
| CREATE  | myresource    | foo           | myresource.foo    |
| REPLACE | myresource    | foobar        | myresource.foobar |
`
	if !visualCheck(t, "TestMarkdownRule1NoExplicitAlignment", buf.String(), expected) {
		t.Error(table.Debug())
	}
}

func TestMarkdownRule2BodyOnly(t *testing.T) {
	var buf bytes.Buffer
	table := tablewriter.NewTable(&buf,
		tablewriter.WithRenderer(renderer.NewMarkdown()),
		tablewriter.WithConfig(tablewriter.Config{
			Row: tw.CellConfig{
				Alignment: tw.CellAlignment{Global: tw.AlignLeft},
			},
			Debug: true,
		}),
	)
	table.Header([]string{"Change", "Resource Type", "Resource Name", "Address"})
	table.Append([]string{"CREATE", "myresource", "foo", "myresource.foo"})
	table.Append([]string{"REPLACE", "myresource", "foobar", "myresource.foobar"})
	table.Render()

	expected := `
| CHANGE  | RESOURCE TYPE | RESOURCE NAME |      ADDRESS      |
|:--------|:--------------|:--------------|:------------------|
| CREATE  | myresource    | foo           | myresource.foo    |
| REPLACE | myresource    | foobar        | myresource.foobar |
`
	if !visualCheck(t, "TestMarkdownRule2BodyOnly", buf.String(), expected) {
		t.Error(table.Debug())
	}
}

func TestMarkdownRule3BothSet(t *testing.T) {
	var buf bytes.Buffer
	table := tablewriter.NewTable(&buf,
		tablewriter.WithRenderer(renderer.NewMarkdown()),
		tablewriter.WithConfig(tablewriter.Config{
			Header: tw.CellConfig{
				Alignment: tw.CellAlignment{Global: tw.AlignLeft},
			},
			Row: tw.CellConfig{
				Alignment: tw.CellAlignment{Global: tw.AlignRight},
			},
			Debug: true,
		}),
	)
	table.Header([]string{"Name", "Age", "City"})
	table.Append([]string{"Alice", "25", "New York"})
	table.Append([]string{"Bob", "30", "Boston"})
	table.Render()

	expected := `
| NAME  | AGE | CITY     |
|------:|----:|---------:|
| Alice |  25 | New York |
|   Bob |  30 |   Boston |
`
	if !visualCheck(t, "TestMarkdownRule3BothSet", buf.String(), expected) {
		t.Error(table.Debug())
	}
}

func TestMarkdownRule4HeaderOnly(t *testing.T) {
	var buf bytes.Buffer
	table := tablewriter.NewTable(&buf,
		tablewriter.WithRenderer(renderer.NewMarkdown()),
		tablewriter.WithConfig(tablewriter.Config{
			Header: tw.CellConfig{
				Alignment: tw.CellAlignment{Global: tw.AlignLeft},
			},
			Debug: true,
		}),
	)
	table.Header([]string{"Name", "Age", "City"})
	table.Append([]string{"Alice", "25", "New York"})
	table.Append([]string{"Bob", "30", "Boston"})
	table.Render()

	expected := `
| NAME  | AGE | CITY     |
|:------|:----|:---------|
| Alice | 25  | New York |
| Bob   | 30  | Boston   |
`
	if !visualCheck(t, "TestMarkdownRule4HeaderOnly", buf.String(), expected) {
		t.Error(table.Debug())
	}
}

func TestMarkdownPerColumnAlignment(t *testing.T) {
	var buf bytes.Buffer
	table := tablewriter.NewTable(&buf,
		tablewriter.WithRenderer(renderer.NewMarkdown()),
		tablewriter.WithConfig(tablewriter.Config{
			Header: tw.CellConfig{
				Alignment: tw.CellAlignment{PerColumn: []tw.Align{tw.AlignLeft, tw.AlignCenter, tw.AlignRight}},
			},
			Row: tw.CellConfig{
				Alignment: tw.CellAlignment{PerColumn: []tw.Align{tw.AlignRight, tw.AlignLeft, tw.AlignCenter}},
			},
			Debug: true,
		}),
	)
	table.Header([]string{"Name", "Age", "City"})
	table.Append([]string{"Alice", "25", "New York"})
	table.Render()

	expected := `
| NAME  | AGE |     CITY |
|------:|:----|:--------:|
| Alice | 25  | New York |
`
	if !visualCheck(t, "TestMarkdownPerColumnAlignment", buf.String(), expected) {
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
|:------|:----|:---------|:-------|
| Alice | 25  | New York | OK     |
| Bob   | 30  | Boston   | ERROR  |
`
	if !visualCheck(t, "MarkdownAlignment", buf.String(), expected) {
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
:------|:----|:---------
Alice | 25  | New York 
Bob   | 30  | Boston   
`
	if !visualCheck(t, "MarkdownNoBorders", buf.String(), expected) {
		t.Error(table.Debug())
	}
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
|:-----|:----|:-------|
| Bøb  | 30  | Tōkyō  |
| José | 28  | México |
| 张三 | 35  | 北京   |
`
	if !visualCheck(t, "MarkdownUnicode", buf.String(), expected) {
		t.Error(table.Debug())
	}
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
	if !visualCheck(t, "MarkdownLongHeaders", buf.String(), expected) {
		t.Error(table.Debug())
	}
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
	)
	table.Header([]string{"No", "Description", "Note"})
	table.Append([]string{"1", "This is a very long description that should wrap", "Short"})
	table.Append([]string{"2", "Short desc", "Another note"})
	table.Render()

	expected := `
| NO |   DESCRIPTION    |     NOTE     |
|:---|:-----------------|:-------------|
| 1  | This is a very   | Short        |
|    | long description |              |
|    | that should wrap |              |
| 2  | Short desc       | Another note |
`
	if !visualCheck(t, "MarkdownLongValues", buf.String(), expected) {
		t.Error(table.Debug())
	}
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
			Alignment: tw.CellAlignment{Global: tw.AlignLeft},
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
|:------|:----|:---------|
|>Alice<|>25<<|>New York<|
|>Bob<<<|>30<<|>Boston<<<|
`
	if !visualCheck(t, "MarkdownCustomPadding", buf.String(), expected) {
		t.Error(table.Debug())
	}
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
|:----------------|:-------|
| Same            | Unique |
`
	if !visualCheck(t, "MarkdownHorizontalMerge", buf.String(), expected) {
		t.Error(table.Debug())
	}
}

func TestMarkdownEmptyTable(t *testing.T) {
	var buf bytes.Buffer
	table := tablewriter.NewTable(&buf,
		tablewriter.WithRenderer(renderer.NewMarkdown()),
	)
	table.Render()

	expected := ""
	if !visualCheck(t, "MarkdownEmptyTable", buf.String(), expected) {
		t.Error(table.Debug())
	}
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
|:------|:----|:---------|
| Alice | 25  | New York |
| Bob   | 30  | Boston   |
| Total | 2   |          |
`
	if !visualCheck(t, "MarkdownWithFooter", buf.String(), expected) {
		t.Error(table.Debug())
	}
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
|:------:|
|  Data  |
`
		if !visualCheck(t, "AlignNone", buf.String(), expected) {
			t.Fatal(table.Debug())
		}
	})
}

func TestMarkdownMixedAlignments(t *testing.T) {
	var buf bytes.Buffer
	table := tablewriter.NewTable(&buf,
		tablewriter.WithRenderer(renderer.NewMarkdown()),
		tablewriter.WithConfig(tablewriter.Config{
			Header: tw.CellConfig{
				Alignment: tw.CellAlignment{PerColumn: []tw.Align{tw.AlignLeft, tw.AlignNone, tw.AlignRight}},
			},
			Row: tw.CellConfig{
				Alignment: tw.CellAlignment{PerColumn: []tw.Align{tw.AlignNone, tw.AlignCenter, tw.AlignNone}},
			},
		}),
	)
	table.Header([]string{"Left", "Default", "Right"})
	table.Append([]string{"A", "B", "C"})
	table.Render()

	expected := `
| LEFT | DEFAULT | RIGHT |
|:-----|:-------:|------:|
| A    |    B    |     C |
`
	if !visualCheck(t, "TestMarkdownMixedAlignments", buf.String(), expected) {
		t.Error(table.Debug())
	}
}

func TestMarkdownCenterAlignment(t *testing.T) {
	var buf bytes.Buffer
	table := tablewriter.NewTable(&buf,
		tablewriter.WithRenderer(renderer.NewMarkdown()),
		tablewriter.WithConfig(tablewriter.Config{
			Header: tw.CellConfig{
				Alignment: tw.CellAlignment{Global: tw.AlignCenter},
			},
			Row: tw.CellConfig{
				Alignment: tw.CellAlignment{Global: tw.AlignCenter},
			},
		}),
	)
	table.Header([]string{"Name", "Age"})
	table.Append([]string{"Alice", "25"})
	table.Render()

	expected := `
| NAME  | AGE |
|:-----:|:---:|
| Alice | 25  |
`
	if !visualCheck(t, "TestMarkdownCenterAlignment", buf.String(), expected) {
		t.Error(table.Debug())
	}
}
