package tests

import (
	"bytes"
	"testing"

	"github.com/olekukonko/tablewriter"
	"github.com/olekukonko/tablewriter/renderer"
	"github.com/olekukonko/tablewriter/tw"
)

func TestVerticalMerge(t *testing.T) {
	var buf bytes.Buffer
	table := tablewriter.NewTable(&buf, tablewriter.WithConfig(tablewriter.Config{
		Row: tw.CellConfig{
			Merging: tw.CellMerging{
				Mode: tw.MergeVertical,
			},
			Alignment: tw.CellAlignment{PerColumn: []tw.Align{tw.Skip, tw.Skip, tw.AlignRight, tw.AlignRight}},
		},
	}), tablewriter.WithRenderer(renderer.NewBlueprint(tw.Rendition{
		Settings: tw.Settings{
			Separators: tw.Separators{
				BetweenRows: tw.Off,
			},
		},
	})),
	)
	table.Header([]string{"Name", "Sign", "Rating"})
	table.Append([]string{"A", "The Good", "500"})
	table.Append([]string{"A", "The Very very Bad Man", "288"})
	table.Append([]string{"B", "", "120"})
	table.Append([]string{"B", "", "200"})
	table.Render()

	expected := `
	┌──────┬───────────────────────┬────────┐
	│ NAME │         SIGN          │ RATING │
	├──────┼───────────────────────┼────────┤
	│ A    │ The Good              │    500 │
	│      │ The Very very Bad Man │    288 │
	│ B    │                       │    120 │
	│      │                       │    200 │
	└──────┴───────────────────────┴────────┘
`

	if !visualCheck(t, "VerticalMerge", buf.String(), expected) {
		t.Error(table.Debug())
	}
}

func TestHorizontalMerge(t *testing.T) {
	var buf bytes.Buffer
	table := tablewriter.NewTable(&buf, tablewriter.WithConfig(tablewriter.Config{
		Row: tw.CellConfig{
			Merging: tw.CellMerging{
				Mode: tw.MergeHorizontal,
			},
			Alignment: tw.CellAlignment{PerColumn: []tw.Align{tw.AlignCenter, tw.AlignCenter, tw.AlignCenter, tw.AlignCenter}},
		},
	}), tablewriter.WithRenderer(renderer.NewBlueprint(tw.Rendition{
		Settings: tw.Settings{
			Separators: tw.Separators{
				BetweenRows: tw.Off,
			},
		},
	})),
	)
	table.Header([]string{"Col1", "Col2", "Col2"})
	table.Append([]string{"A", "B", "B"})
	table.Append([]string{"A", "A", "C"})
	table.Append([]string{"A", "B", "C"})
	table.Append([]string{"B", "C", "C"})
	table.Append([]string{"B", "C", "D"})
	table.Append([]string{"D", "D", "D"})
	table.Render()

	expected := `
     ┌───────┬───────┬───────┐
     │ COL 1 │ COL 2 │ COL 2 │
     ├───────┼───────┴───────┤
     │   A   │       B       │
     │       A       │   C   │
     │   A   │   B   │   C   │
     │   B   │       C       │
     │   B   │   C   │   D   │
     │           D           │
     └───────────────────────┘

`

	if !visualCheck(t, "TestHorizontalMerge", buf.String(), expected) {
		t.Error(table.Debug())
	}
}

func TestHorizontalMergeEachLine(t *testing.T) {
	var buf bytes.Buffer
	table := tablewriter.NewTable(&buf, tablewriter.WithConfig(tablewriter.Config{
		Row: tw.CellConfig{
			Merging: tw.CellMerging{
				Mode: tw.MergeHorizontal,
			},
		},
		Footer: tw.CellConfig{
			Merging: tw.CellMerging{
				Mode: tw.MergeHorizontal,
			},
			Alignment: tw.CellAlignment{PerColumn: []tw.Align{tw.Skip, tw.Skip, tw.AlignRight, tw.AlignLeft}},
		},
	}),
		tablewriter.WithRenderer(renderer.NewBlueprint(tw.Rendition{
			// Symbols: tw.NewSymbols(tw.StyleMerger),
			Settings: tw.Settings{
				Separators: tw.Separators{
					BetweenRows: tw.On,
				},
			},
		})),
	)
	table.Header([]string{"Date", "Section A", "Section B", "Section C", "Section D", "Section E"})
	table.Append([]string{"1/1/2014", "apple", "boy", "cat", "dog", "elephant"})
	table.Append([]string{"1/1/2014", "apple", "apple", "boy", "dog", "elephant"})
	table.Append([]string{"1/1/2014", "apple", "boy", "boy", "cat", "dog"})
	table.Append([]string{"1/1/2014", "apple", "boy", "cat", "cat", "dog"})
	table.Render()

	expected := `
	┌──────────┬───────────┬───────────┬───────────┬───────────┬───────────┐
	│   DATE   │ SECTION A │ SECTION B │ SECTION C │ SECTION D │ SECTION E │
	├──────────┼───────────┼───────────┼───────────┼───────────┼───────────┤
	│ 1/1/2014 │ apple     │ boy       │ cat       │ dog       │ elephant  │
	├──────────┼───────────┴───────────┼───────────┼───────────┼───────────┤
	│ 1/1/2014 │ apple                 │ boy       │ dog       │ elephant  │
	├──────────┼───────────┬───────────┴───────────┼───────────┼───────────┤
	│ 1/1/2014 │ apple     │ boy                   │ cat       │ dog       │
	├──────────┼───────────┼───────────┬───────────┴───────────┼───────────┤
	│ 1/1/2014 │ apple     │ boy       │ cat                   │ dog       │
	└──────────┴───────────┴───────────┴───────────────────────┴───────────┘


`
	check := visualCheck(t, "HorizontalMergeEachLine", buf.String(), expected)
	if !check {
		t.Error(table.Debug())
	}
}

func TestHorizontalMergeEachLineCenter(t *testing.T) {
	var buf bytes.Buffer
	table := tablewriter.NewTable(&buf, tablewriter.WithConfig(tablewriter.Config{
		Row: tw.CellConfig{
			Merging: tw.CellMerging{
				Mode: tw.MergeHorizontal,
			},
			Alignment: tw.CellAlignment{Global: tw.AlignCenter},
		},
	}),
		tablewriter.WithRenderer(renderer.NewBlueprint(tw.Rendition{
			// Symbols: tw.NewSymbols(tw.StyleMerger),
			Settings: tw.Settings{
				Separators: tw.Separators{
					BetweenRows: tw.On,
				},
			},
		})),
	)
	table.Header([]string{"Date", "Section A", "Section B", "Section C", "Section D", "Section E"})
	table.Append([]string{"1/1/2014", "apple", "boy", "cat", "dog", "elephant"})
	table.Append([]string{"1/1/2014", "apple", "apple", "boy", "dog", "elephant"})
	table.Append([]string{"1/1/2014", "apple", "boy", "boy", "cat", "dog"})
	table.Append([]string{"1/1/2014", "apple", "boy", "cat", "cat", "dog"})
	table.Render()

	expected := `
		┌──────────┬───────────┬───────────┬───────────┬───────────┬───────────┐
		│   DATE   │ SECTION A │ SECTION B │ SECTION C │ SECTION D │ SECTION E │
		├──────────┼───────────┼───────────┼───────────┼───────────┼───────────┤
		│ 1/1/2014 │   apple   │    boy    │    cat    │    dog    │ elephant  │
		├──────────┼───────────┴───────────┼───────────┼───────────┼───────────┤
		│ 1/1/2014 │         apple         │    boy    │    dog    │ elephant  │
		├──────────┼───────────┬───────────┴───────────┼───────────┼───────────┤
		│ 1/1/2014 │   apple   │          boy          │    cat    │    dog    │
		├──────────┼───────────┼───────────┬───────────┴───────────┼───────────┤
		│ 1/1/2014 │   apple   │    boy    │          cat          │    dog    │
		└──────────┴───────────┴───────────┴───────────────────────┴───────────┘

`
	check := visualCheck(t, "HorizontalMergeEachLineCenter", buf.String(), expected)
	if !check {
		t.Error(table.Debug())
	}
}

func TestHorizontalMergeAlignFooter(t *testing.T) {
	var buf bytes.Buffer
	table := tablewriter.NewTable(&buf, tablewriter.WithConfig(tablewriter.Config{
		Row: tw.CellConfig{
			Merging: tw.CellMerging{
				Mode: tw.MergeHorizontal,
			},
		},
		Footer: tw.CellConfig{
			Merging: tw.CellMerging{
				Mode: tw.MergeHorizontal,
			},
			Alignment: tw.CellAlignment{PerColumn: []tw.Align{tw.Skip, tw.Skip, tw.AlignRight, tw.AlignLeft}},
		},
	}),
		tablewriter.WithRenderer(renderer.NewBlueprint(tw.Rendition{
			Settings: tw.Settings{
				Separators: tw.Separators{
					BetweenRows: tw.On,
				},
			},
		})),
	)
	table.Header([]string{"Date", "Description", "Status", "Conclusion"})
	table.Append([]string{"1/1/2014", "Domain name", "Successful", "Successful"})
	table.Append([]string{"1/1/2014", "Domain name", "Pending", "Waiting"})
	table.Append([]string{"1/1/2014", "Domain name", "Successful", "Rejected"})
	table.Footer([]string{"", "", "TOTAL", "$145.93"}) // Fixed from Append
	table.Render()

	expected := `
  ┌──────────┬─────────────┬────────────┬────────────┐
  │   DATE   │ DESCRIPTION │   STATUS   │ CONCLUSION │
  ├──────────┼─────────────┼────────────┴────────────┤
  │ 1/1/2014 │ Domain name │ Successful              │
  ├──────────┼─────────────┼────────────┬────────────┤
  │ 1/1/2014 │ Domain name │ Pending    │ Waiting    │
  ├──────────┼─────────────┼────────────┼────────────┤
  │ 1/1/2014 │ Domain name │ Successful │ Rejected   │
  ├──────────┴─────────────┴────────────┼────────────┤
  │                               TOTAL │ $145.93    │
  └─────────────────────────────────────┴────────────┘
`
	check := visualCheck(t, "HorizontalMergeAlignFooter", buf.String(), expected)
	if !check {
		t.Error(table.Debug())
	}
}

func TestVerticalMergeLines(t *testing.T) {
	var buf bytes.Buffer
	table := tablewriter.NewTable(&buf, tablewriter.WithConfig(tablewriter.Config{
		Row: tw.CellConfig{
			Merging: tw.CellMerging{
				Mode: tw.MergeVertical,
			},
			Alignment: tw.CellAlignment{PerColumn: []tw.Align{tw.Skip, tw.Skip, tw.AlignRight, tw.AlignRight}},
		},
	}), tablewriter.WithRenderer(renderer.NewBlueprint(tw.Rendition{
		// Symbols: tw.NewSymbols(tw.StyleMerger),
		Settings: tw.Settings{
			Separators: tw.Separators{
				BetweenRows: tw.On,
			},
		},
	})))
	table.Header([]string{"Name", "Sign", "Rating"})
	table.Append([]string{"A", "The Good", "500"})
	table.Append([]string{"A", "The Very very Bad Man", "288"})
	table.Append([]string{"B", "C", "120"})
	table.Append([]string{"B", "C", "200"})
	table.Render()

	expected := `
       ┌──────┬───────────────────────┬────────┐
       │ NAME │         SIGN          │ RATING │
       ├──────┼───────────────────────┼────────┤
       │ A    │ The Good              │    500 │
       │      ├───────────────────────┼────────┤
       │      │ The Very very Bad Man │    288 │
       ├──────┼───────────────────────┼────────┤
       │ B    │ C                     │    120 │
       │      │                       ├────────┤
       │      │                       │    200 │
       └──────┴───────────────────────┴────────┘
`

	if !visualCheck(t, "VerticalMergeLines", buf.String(), expected) {
		t.Error(table.Debug())
	}
}

func TestMergeBoth(t *testing.T) {
	var buf bytes.Buffer

	c := tablewriter.Config{
		Row: tw.CellConfig{
			Merging: tw.CellMerging{
				Mode: tw.MergeBoth,
			},
		},
		Footer: tw.CellConfig{
			Merging: tw.CellMerging{
				Mode: tw.MergeHorizontal,
			},
			Alignment: tw.CellAlignment{PerColumn: []tw.Align{tw.AlignRight, tw.AlignRight, tw.AlignRight, tw.AlignLeft}},
		},
	}

	r := renderer.NewBlueprint(tw.Rendition{
		Settings: tw.Settings{
			Separators: tw.Separators{
				BetweenRows: tw.On,
			},
		},
	})

	t.Run("mixed-1", func(t *testing.T) {
		buf.Reset()
		table := tablewriter.NewTable(&buf, tablewriter.WithConfig(c), tablewriter.WithRenderer(r))
		table.Header([]string{"Date", "Description", "Status", "Conclusion"})
		table.Append([]string{"1/1/2014", "Domain name", "Successful", "Successful"})
		table.Append([]string{"1/1/2014", "Domain name", "Pending", "Waiting"})
		table.Append([]string{"1/1/2014", "Domain name", "Successful", "Rejected"})
		table.Footer([]string{"TOTAL", "TOTAL", "TOTAL", "$145.93"})
		table.Render()

		expected := `
        ┌──────────┬─────────────┬────────────┬────────────┐
        │   DATE   │ DESCRIPTION │   STATUS   │ CONCLUSION │
        ├──────────┼─────────────┼────────────┴────────────┤
        │ 1/1/2014 │ Domain name │ Successful              │
        │          │             ├────────────┬────────────┤
        │          │             │ Pending    │ Waiting    │
        │          │             ├────────────┼────────────┤
        │          │             │ Successful │ Rejected   │
        ├──────────┴─────────────┴────────────┼────────────┤
        │                               TOTAL │ $145.93    │
        └─────────────────────────────────────┴────────────┘

`
		check := visualCheck(t, "TestMergeBoth-mixed-1", buf.String(), expected)
		if !check {
			t.Log(table.Debug())
		}
	})

	t.Run("mixed-2", func(t *testing.T) {
		buf.Reset()
		table := tablewriter.NewTable(&buf, tablewriter.WithConfig(c), tablewriter.WithRenderer(r))
		table.Header([]string{"Date", "Description", "Status", "Conclusion"})
		table.Append([]string{"1/1/2014", "Domain name", "Successful", "Successful"})
		table.Append([]string{"1/1/2014", "Domain name", "Pending", "Waiting"})
		table.Append([]string{"1/1/2015", "Domain name", "Successful", "Rejected"})
		table.Footer([]string{"TOTAL", "TOTAL", "TOTAL", "$145.93"})
		table.Render()

		expected := `
        ┌──────────┬─────────────┬────────────┬────────────┐
        │   DATE   │ DESCRIPTION │   STATUS   │ CONCLUSION │
        ├──────────┼─────────────┼────────────┴────────────┤
        │ 1/1/2014 │ Domain name │ Successful              │
        │          │             ├────────────┬────────────┤
        │          │             │ Pending    │ Waiting    │
        ├──────────┤             ├────────────┼────────────┤
        │ 1/1/2015 │             │ Successful │ Rejected   │
        ├──────────┴─────────────┴────────────┼────────────┤
        │                               TOTAL │ $145.93    │
        └─────────────────────────────────────┴────────────┘
		`
		check := visualCheck(t, "TestMergeBoth-mixed-2", buf.String(), expected)
		if !check {
			t.Log(table.Debug())
		}
	})
}

func TestMergeHierarchical(t *testing.T) {
	var buf bytes.Buffer
	table := tablewriter.NewTable(&buf, tablewriter.WithConfig(tablewriter.Config{
		Row: tw.CellConfig{
			Merging: tw.CellMerging{
				Mode: tw.MergeHierarchical,
			},
		},
	}),
		tablewriter.WithRenderer(renderer.NewBlueprint(tw.Rendition{
			Symbols: tw.NewSymbols(tw.StyleASCII),
			Settings: tw.Settings{
				Separators: tw.Separators{
					BetweenRows: tw.On,
				},
			},
		})),
	)
	table.Header([]string{"0", "1", "2", "3"})
	table.Append([]string{"A", "a", "c", "-"})
	table.Append([]string{"A", "b", "c", "-"})
	table.Append([]string{"A", "b", "d", "-"})
	table.Append([]string{"B", "b", "d", "-"})
	table.Render()

	expected := `
	+---+---+---+---+
	| 0 | 1 | 2 | 3 |
	+---+---+---+---+
	| A | a | c | - |
	|   +---+---+---+
	|   | b | c | - |
	|   |   +---+---+
	|   |   | d | - |
	+---+---+---+---+
	| B | b | d | - |
	+---+---+---+---+
`
	check := visualCheck(t, "MergeHierarchical", buf.String(), expected)
	if !check {
		t.Error(table.Debug())
	}
}

func TestMergeHierarchicalUnicode(t *testing.T) {
	var buf bytes.Buffer
	table := tablewriter.NewTable(&buf, tablewriter.WithConfig(tablewriter.Config{
		Row: tw.CellConfig{
			Merging: tw.CellMerging{
				Mode: tw.MergeHierarchical,
			},
		},
	}),
		tablewriter.WithRenderer(renderer.NewBlueprint(tw.Rendition{
			// Symbols: tw.NewSymbols(tw.StyleRounded),
			Settings: tw.Settings{
				Separators: tw.Separators{
					BetweenRows: tw.On,
				},
			},
		})),
	)
	table.Header([]string{"0", "1", "2", "3"})
	table.Append([]string{"A", "a", "c", "-"})
	table.Append([]string{"A", "b", "c", "-"})
	table.Append([]string{"A", "b", "d", "-"})
	table.Append([]string{"B", "b", "d", "-"})
	table.Render()

	expected := `
        ┌───┬───┬───┬───┐
        │ 0 │ 1 │ 2 │ 3 │
        ├───┼───┼───┼───┤
        │ A │ a │ c │ - │
        │   ├───┼───┼───┤
        │   │ b │ c │ - │
        │   │   ├───┼───┤
        │   │   │ d │ - │
        ├───┼───┼───┼───┤
        │ B │ b │ d │ - │
        └───┴───┴───┴───┘
`
	check := visualCheck(t, "MergeHierarchicalUnicode", buf.String(), expected)
	if !check {
		t.Error(table.Debug())
	}
}

func TestMergeWithPadding(t *testing.T) {
	var buf bytes.Buffer

	table := tablewriter.NewTable(&buf, tablewriter.WithConfig(tablewriter.Config{
		Row: tw.CellConfig{
			Merging: tw.CellMerging{
				Mode: tw.MergeBoth,
			},
			Alignment: tw.CellAlignment{PerColumn: []tw.Align{tw.Skip, tw.Skip, tw.AlignRight, tw.AlignLeft}},
		},
		Footer: tw.CellConfig{
			Padding: tw.CellPadding{
				Global:    tw.Padding{Left: "*", Right: "*", Top: "", Bottom: ""},
				PerColumn: []tw.Padding{{}, {}, {Bottom: "^"}, {Bottom: "."}},
			},
			Alignment: tw.CellAlignment{PerColumn: []tw.Align{tw.Skip, tw.Skip, tw.AlignRight, tw.AlignLeft}},
		},
	}), tablewriter.WithRenderer(renderer.NewBlueprint(tw.Rendition{
		// Symbols: tw.NewSymbols(tw.StyleASCII),
		Settings: tw.Settings{
			Separators: tw.Separators{
				BetweenRows: tw.On,
			},
		},
	})))

	table.Header([]string{"Date", "Description", "Status", "Conclusion"})
	table.Append([]string{"1/1/2014", "Domain name", "Successful", "Successful"})
	table.Append([]string{"1/1/2014", "Domain name", "Pending", "Waiting"})
	table.Append([]string{"1/1/2014", "Domain name", "Successful", "Rejected"})
	table.Append([]string{"", "", "TOTAL", "$145.93"})
	table.Render()

	expected := `
	┌──────────┬─────────────┬────────────┬────────────┐
	│   DATE   │ DESCRIPTION │   STATUS   │ CONCLUSION │
	├──────────┼─────────────┼────────────┴────────────┤
	│ 1/1/2014 │ Domain name │              Successful │
	│          │             ├────────────┬────────────┤
	│          │             │    Pending │ Waiting    │
	│          │             ├────────────┼────────────┤
	│          │             │ Successful │ Rejected   │
	├──────────┼─────────────┼────────────┼────────────┤
	│          │             │      TOTAL │ $145.93    │
	│          │             │^^^^^^^^^^^^│............│
	└──────────┴─────────────┴────────────┴────────────┘
`
	visualCheck(t, "MergeWithPadding", buf.String(), expected)
}

func TestMergeWithMultipleLines(t *testing.T) {
	var buf bytes.Buffer

	data := [][]string{
		{"Module", "Description", "Version", "Status"},
		{"core\nutils", "Utility\nfunctions", "v1.0.0", "stable"},
		{"core\nutils", "Helper\nroutines", "v1.1.0", "beta"},
		{"web\nserver", "HTTP\nserver", "v2.0.0", "stable"},
		{"web\nserver", "", "v2.1.0", "testing"},
		{"db\nclient", "Database\naccess", "v3.0.0", ""},
	}

	t.Run("Horizontal", func(t *testing.T) {
		buf.Reset()
		table := tablewriter.NewTable(&buf, tablewriter.WithConfig(tablewriter.Config{
			Row: tw.CellConfig{
				Merging: tw.CellMerging{
					Mode: tw.MergeHorizontal,
				},
				Alignment: tw.CellAlignment{PerColumn: []tw.Align{tw.AlignLeft, tw.AlignLeft, tw.AlignLeft, tw.AlignLeft}},
			},
		}), tablewriter.WithRenderer(renderer.NewBlueprint(tw.Rendition{
			Settings: tw.Settings{
				Separators: tw.Separators{
					BetweenRows: tw.On,
				},
			},
		})))
		table.Header(data[0])
		table.Bulk(data[1:])
		table.Render()

		expected := `
┌────────┬─────────────┬─────────┬─────────┐
│ MODULE │ DESCRIPTION │ VERSION │ STATUS  │
├────────┼─────────────┼─────────┼─────────┤
│ core   │ Utility     │ v1.0.0  │ stable  │
│ utils  │ functions   │         │         │
├────────┼─────────────┼─────────┼─────────┤
│ core   │ Helper      │ v1.1.0  │ beta    │
│ utils  │ routines    │         │         │
├────────┼─────────────┼─────────┼─────────┤
│ web    │ HTTP        │ v2.0.0  │ stable  │
│ server │ server      │         │         │
├────────┼─────────────┼─────────┼─────────┤
│ web    │             │ v2.1.0  │ testing │
│ server │             │         │         │
├────────┼─────────────┼─────────┼─────────┤
│ db     │ Database    │ v3.0.0  │         │
│ client │ access      │         │         │
└────────┴─────────────┴─────────┴─────────┘
`
		// t.Log("====== LOG", table.Logger().Enabled())
		if !visualCheck(t, "Horizontal", buf.String(), expected) {
			t.Error(table.Debug())
		}
	})

	t.Run("Vertical", func(t *testing.T) {
		buf.Reset()
		table := tablewriter.NewTable(&buf, tablewriter.WithConfig(tablewriter.Config{
			Row: tw.CellConfig{
				Merging: tw.CellMerging{
					Mode: tw.MergeVertical,
				},
				Alignment: tw.CellAlignment{PerColumn: []tw.Align{tw.AlignLeft, tw.AlignLeft, tw.AlignLeft, tw.AlignLeft}},
			},
			Debug: true,
		}), tablewriter.WithRenderer(renderer.NewBlueprint(tw.Rendition{
			Settings: tw.Settings{
				Separators: tw.Separators{
					BetweenRows: tw.On,
				},
			},
		})))
		table.Header(data[0])
		table.Bulk(data[1:])
		table.Render()

		expected := `
┌────────┬─────────────┬─────────┬─────────┐
│ MODULE │ DESCRIPTION │ VERSION │ STATUS  │
├────────┼─────────────┼─────────┼─────────┤
│ core   │ Utility     │ v1.0.0  │ stable  │
│ utils  │ functions   │         │         │
│        ├─────────────┼─────────┼─────────┤
│        │ Helper      │ v1.1.0  │ beta    │
│        │ routines    │         │         │
├────────┼─────────────┼─────────┼─────────┤
│ web    │ HTTP        │ v2.0.0  │ stable  │
│ server │ server      │         │         │
│        ├─────────────┼─────────┼─────────┤
│        │             │ v2.1.0  │ testing │
├────────┼─────────────┼─────────┼─────────┤
│ db     │ Database    │ v3.0.0  │         │
│ client │ access      │         │         │
└────────┴─────────────┴─────────┴─────────┘
`
		if !visualCheck(t, "Vertical", buf.String(), expected) {
			t.Error(table.Debug())
		}
	})

	t.Run("Hierarch", func(t *testing.T) {
		buf.Reset()
		table := tablewriter.NewTable(&buf, tablewriter.WithConfig(tablewriter.Config{
			Row: tw.CellConfig{
				Merging: tw.CellMerging{
					Mode: tw.MergeHierarchical,
				},
				Alignment: tw.CellAlignment{PerColumn: []tw.Align{tw.AlignLeft, tw.AlignLeft, tw.AlignLeft, tw.AlignLeft}},
			},
		}), tablewriter.WithRenderer(renderer.NewBlueprint(tw.Rendition{
			Settings: tw.Settings{
				Separators: tw.Separators{
					BetweenRows: tw.On,
				},
			},
		})))
		table.Header(data[0])
		table.Bulk(data[1:])
		table.Render()

		expected := `
        ┌────────┬─────────────┬─────────┬─────────┐
        │ MODULE │ DESCRIPTION │ VERSION │ STATUS  │
        ├────────┼─────────────┼─────────┼─────────┤
        │ core   │ Utility     │ v1.0.0  │ stable  │
        │ utils  │ functions   │         │         │
        │        ├─────────────┼─────────┼─────────┤
        │        │ Helper      │ v1.1.0  │ beta    │
        │        │ routines    │         │         │
        ├────────┼─────────────┼─────────┼─────────┤
        │ web    │ HTTP        │ v2.0.0  │ stable  │
        │ server │ server      │         │         │
        │        ├─────────────┼─────────┼─────────┤
        │        │             │ v2.1.0  │ testing │
        ├────────┼─────────────┼─────────┼─────────┤
        │ db     │ Database    │ v3.0.0  │         │
        │ client │ access      │         │         │
        └────────┴─────────────┴─────────┴─────────┘
`
		if !visualCheck(t, "Hierarch", buf.String(), expected) {
			// t.Error(table.Debug())
		}
	})
}

func TestMergeHierarchicalCombined(t *testing.T) {
	var buf bytes.Buffer
	table := tablewriter.NewTable(&buf, tablewriter.WithConfig(tablewriter.Config{
		Header: tw.CellConfig{Alignment: tw.CellAlignment{Global: tw.AlignCenter}},
		Row: tw.CellConfig{
			Merging:   tw.CellMerging{Mode: tw.MergeHierarchical},
			Alignment: tw.CellAlignment{Global: tw.AlignLeft},
		},
		Footer: tw.CellConfig{
			Alignment: tw.CellAlignment{Global: tw.AlignLeft},
			Merging: tw.CellMerging{
				Mode: tw.MergeHorizontal,
			},
		},
		Debug: true,
	}),
		tablewriter.WithRenderer(renderer.NewBlueprint(tw.Rendition{
			Settings: tw.Settings{Separators: tw.Separators{BetweenRows: tw.On}},
		})),
	)

	data := [][]string{
		{"Engineering", "Backend", "API Team", "Alice"},
		{"Engineering", "Backend", "Database Team", "Bob"},
		{"Engineering", "Frontend", "UI Team", "Charlie"},
		{"Marketing", "Digital", "SEO Team", "Dave"},
		{"Marketing", "Digital", "Content Team", "Eve"},
	}

	table.Header([]string{"Department", "Division", "Team", "Lead"})
	table.Bulk(data)
	table.Footer([]string{"Total Teams", "5", "5", "5"})
	table.Render()

	expected := `
		┌─────────────┬──────────┬───────────────┬─────────┐
		│ DEPARTMENT  │ DIVISION │     TEAM      │  LEAD   │
		├─────────────┼──────────┼───────────────┼─────────┤
		│ Engineering │ Backend  │ API Team      │ Alice   │
		│             │          ├───────────────┼─────────┤
		│             │          │ Database Team │ Bob     │
		│             ├──────────┼───────────────┼─────────┤
		│             │ Frontend │ UI Team       │ Charlie │
		├─────────────┼──────────┼───────────────┼─────────┤
		│ Marketing   │ Digital  │ SEO Team      │ Dave    │
		│             │          ├───────────────┼─────────┤
		│             │          │ Content Team  │ Eve     │
		├─────────────┼──────────┴───────────────┴─────────┤
		│ Total Teams │ 5                                  │
		└─────────────┴────────────────────────────────────┘

`
	check := visualCheck(t, "MergeHierarchical", buf.String(), expected)
	if !check {
		t.Error(table.Debug())
	}
}

// TestMergeVerticalByColumnIndex verifies that vertical merging
// only applies to the columns specified by the new API.
func TestMergeVerticalByColumnIndex(t *testing.T) {
	var buf bytes.Buffer

	// Data where columns 0, 1, and 3 have duplicates that could be merged.
	data := [][]string{
		{"A", "The Good", "500", "OK"},
		{"A", "The Bad", "288", "OK"},
		{"B", "The Ugly", "120", "FAIL"},
		{"B", "The Ugly", "200", "FAIL"},
		{"B", "The Ugly", "300", "OK"},
	}

	// Use the new fluent builder to configure column-specific merging.
	b := tablewriter.NewConfigBuilder()
	b.Row().Merging().
		WithMode(tw.MergeVertical). // Enable Vertical Merging
		ByColumnIndex([]int{0, 3}). // ONLY apply it to column 0 and 3
		Build()

	table := tablewriter.NewTable(&buf,
		tablewriter.WithConfig(b.Build()),
		tablewriter.WithRenderer(renderer.NewBlueprint(tw.Rendition{
			Settings: tw.Settings{
				Separators: tw.Separators{
					BetweenRows: tw.On,
				},
			},
		})))

	table.Header([]string{"Name", "Sign", "Rating", "Status"})
	table.Bulk(data)
	table.Render()

	// EXPECTED:
	// - Column 0 ("Name") SHOULD be merged.
	// - Column 1 ("Sign") should NOT be merged, even though "The Ugly" repeats.
	// - Column 3 ("Status") SHOULD be merged.
	expected := `
			┌──────┬──────────┬────────┬────────┐
			│ NAME │   SIGN   │ RATING │ STATUS │
			├──────┼──────────┼────────┼────────┤
			│ A    │ The Good │ 500    │ OK     │
			│      ├──────────┼────────┤        │
			│      │ The Bad  │ 288    │        │
			├──────┼──────────┼────────┼────────┤
			│ B    │ The Ugly │ 120    │ FAIL   │
			│      ├──────────┼────────┤        │
			│      │ The Ugly │ 200    │        │
			│      ├──────────┼────────┼────────┤
			│      │ The Ugly │ 300    │ OK     │
			└──────┴──────────┴────────┴────────┘
`

	if !visualCheck(t, "MergeVerticalByColumnIndex", buf.String(), expected) {
		t.Error(table.Debug())
	}
}

// TestMergeHierarchicalByColumnIndex verifies that hierarchical merging
// correctly respects the column index filter.
func TestMergeHierarchicalByColumnIndex(t *testing.T) {
	var buf bytes.Buffer

	data := [][]string{
		{"Engineering", "Backend", "API Team", "Alice"},
		{"Engineering", "Backend", "Database Team", "Bob"},
		{"Engineering", "Frontend", "UI Team", "Charlie"},
		{"Marketing", "Digital", "SEO Team", "Dave"},
		{"Marketing", "Digital", "Content Team", "Eve"},
		{"Marketing", "Brand", "PR Team", "Frank"},
	}

	// Use the new fluent builder to enable hierarchical merging for columns 0 and 1 only.
	b := tablewriter.NewConfigBuilder()
	b.Row().Merging().
		WithMode(tw.MergeHierarchical).
		ByColumnIndex([]int{0, 1}).
		Build()

	table := tablewriter.NewTable(&buf,
		tablewriter.WithConfig(b.Build()),
		tablewriter.WithRenderer(renderer.NewBlueprint(tw.Rendition{
			Settings: tw.Settings{
				Separators: tw.Separators{
					BetweenRows: tw.On,
				},
			},
		})))

	table.Header([]string{"Department", "Division", "Team", "Lead"})
	table.Bulk(data)
	table.Render()

	// EXPECTED:
	// - Column 0 ("Department") SHOULD merge.
	// - Column 1 ("Division") SHOULD merge hierarchically (e.g., "Backend" and "Digital").
	// - Column 2 ("Team") should NOT merge because it's not in the index list.
	//   This will break the hierarchical chain for column 3.
	// - Column 3 ("Lead") should NOT merge.
	expected := `
		┌─────────────┬──────────┬───────────────┬─────────┐
		│ DEPARTMENT  │ DIVISION │     TEAM      │  LEAD   │
		├─────────────┼──────────┼───────────────┼─────────┤
		│ Engineering │ Backend  │ API Team      │ Alice   │
		│             │          ├───────────────┼─────────┤
		│             │          │ Database Team │ Bob     │
		│             ├──────────┼───────────────┼─────────┤
		│             │ Frontend │ UI Team       │ Charlie │
		├─────────────┼──────────┼───────────────┼─────────┤
		│ Marketing   │ Digital  │ SEO Team      │ Dave    │
		│             │          ├───────────────┼─────────┤
		│             │          │ Content Team  │ Eve     │
		│             ├──────────┼───────────────┼─────────┤
		│             │ Brand    │ PR Team       │ Frank   │
		└─────────────┴──────────┴───────────────┴─────────┘
`
	if !visualCheck(t, "MergeHierarchicalByColumnIndex", buf.String(), expected) {
		t.Error(table.Debug())
	}
}
