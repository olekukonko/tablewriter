package tests

import (
	"bytes"
	"github.com/olekukonko/tablewriter"
	"github.com/olekukonko/tablewriter/renderer"
	"github.com/olekukonko/tablewriter/tw"
	"testing"
)

func TestVerticalMerge(t *testing.T) {
	var buf bytes.Buffer
	table := tablewriter.NewTable(&buf, tablewriter.WithConfig(tablewriter.Config{
		Row: tablewriter.CellConfig{
			Formatting: tablewriter.CellFormatting{
				MergeMode: tw.MergeVertical,
			},
			ColumnAligns: []tw.Align{tw.Skip, tw.Skip, tw.AlignRight, tw.AlignRight},
		},
	}), tablewriter.WithRenderer(renderer.NewDefault(renderer.DefaultConfig{
		Settings: renderer.Settings{
			Separators: renderer.Separators{
				BetweenRows: tw.Off,
			},
		},
	})),
	)
	table.SetHeader([]string{"Name", "Sign", "Rating"})
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
		for _, v := range table.Debug() {
			t.Error(v)
		}
	}
}

func TestHorizontalMerge(t *testing.T) {
	var buf bytes.Buffer
	table := tablewriter.NewTable(&buf, tablewriter.WithConfig(tablewriter.Config{
		Row: tablewriter.CellConfig{
			Formatting: tablewriter.CellFormatting{
				MergeMode: tw.MergeHorizontal,
			},
			ColumnAligns: []tw.Align{tw.AlignCenter, tw.AlignCenter, tw.AlignCenter, tw.AlignCenter},
		},
	}), tablewriter.WithRenderer(renderer.NewDefault(renderer.DefaultConfig{
		Settings: renderer.Settings{
			Separators: renderer.Separators{
				BetweenRows: tw.Off,
			},
		},
	})),
	)
	table.SetHeader([]string{"Col1", "Col2", "Col2"})
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
		for _, v := range table.Debug() {
			t.Error(v)
		}
	}
}

func TestHorizontalMergeEachLine(t *testing.T) {
	var buf bytes.Buffer
	table := tablewriter.NewTable(&buf, tablewriter.WithConfig(tablewriter.Config{
		Row: tablewriter.CellConfig{
			Formatting: tablewriter.CellFormatting{
				MergeMode: tw.MergeHorizontal,
			},
		},
		Footer: tablewriter.CellConfig{
			Formatting: tablewriter.CellFormatting{
				MergeMode: tw.MergeHorizontal,
			},
			ColumnAligns: []tw.Align{tw.Skip, tw.Skip, tw.AlignRight, tw.AlignLeft},
		},
	}),
		tablewriter.WithRenderer(renderer.NewDefault(renderer.DefaultConfig{
			// Symbols: tw.NewSymbols(tw.StyleMerger),
			Settings: renderer.Settings{
				Separators: renderer.Separators{
					BetweenRows: tw.On,
				},
			},
		})),
	)
	table.SetHeader([]string{"Date", "Section A", "Section B", "Section C", "Section D", "Section E"})
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
		for _, v := range table.Debug() {
			t.Error(v)
		}
	}
}

func TestHorizontalMergeEachLineCenter(t *testing.T) {
	var buf bytes.Buffer
	table := tablewriter.NewTable(&buf, tablewriter.WithConfig(tablewriter.Config{
		Row: tablewriter.CellConfig{
			Formatting: tablewriter.CellFormatting{
				Alignment: tw.AlignCenter,
				MergeMode: tw.MergeHorizontal,
			},
		},
	}),
		tablewriter.WithRenderer(renderer.NewDefault(renderer.DefaultConfig{
			// Symbols: tw.NewSymbols(tw.StyleMerger),
			Settings: renderer.Settings{
				Separators: renderer.Separators{
					BetweenRows: tw.On,
				},
			},
		})),
	)
	table.SetHeader([]string{"Date", "Section A", "Section B", "Section C", "Section D", "Section E"})
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
		for _, v := range table.Debug() {
			t.Error(v)
		}
	}
}

func TestHorizontalMergeAlignFooter(t *testing.T) {
	var buf bytes.Buffer
	table := tablewriter.NewTable(&buf, tablewriter.WithConfig(tablewriter.Config{
		Row: tablewriter.CellConfig{
			Formatting: tablewriter.CellFormatting{
				MergeMode: tw.MergeHorizontal,
			},
		},
		Footer: tablewriter.CellConfig{
			Formatting: tablewriter.CellFormatting{
				MergeMode: tw.MergeHorizontal,
			},
			ColumnAligns: []tw.Align{tw.Skip, tw.Skip, tw.AlignRight, tw.AlignLeft},
		},
	}),
		tablewriter.WithRenderer(renderer.NewDefault(renderer.DefaultConfig{
			Settings: renderer.Settings{
				Separators: renderer.Separators{
					BetweenRows: tw.On,
				},
			},
		})),
	)
	table.SetHeader([]string{"Date", "Description", "Status", "Conclusion"})
	table.Append([]string{"1/1/2014", "Domain name", "Successful", "Successful"})
	table.Append([]string{"1/1/2014", "Domain name", "Pending", "Waiting"})
	table.Append([]string{"1/1/2014", "Domain name", "Successful", "Rejected"})
	table.SetFooter([]string{"", "", "TOTAL", "$145.93"}) // Fixed from Append
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
		for _, v := range table.Debug() {
			t.Error(v)
		}
	}
}

func TestVerticalMergeLines(t *testing.T) {
	var buf bytes.Buffer
	table := tablewriter.NewTable(&buf, tablewriter.WithConfig(tablewriter.Config{
		Row: tablewriter.CellConfig{
			Formatting: tablewriter.CellFormatting{
				MergeMode: tw.MergeVertical,
			},
			ColumnAligns: []tw.Align{tw.Skip, tw.Skip, tw.AlignRight, tw.AlignRight},
		},
	}), tablewriter.WithRenderer(renderer.NewDefault(renderer.DefaultConfig{
		// Symbols: tw.NewSymbols(tw.StyleMerger),
		Settings: renderer.Settings{
			Separators: renderer.Separators{
				BetweenRows: tw.On,
			},
		},
	})))
	table.SetHeader([]string{"Name", "Sign", "Rating"})
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
		for _, v := range table.Debug() {
			t.Error(v)
		}
	}
}

func TestMergeBoth(t *testing.T) {
	var buf bytes.Buffer

	c := tablewriter.Config{
		Row: tablewriter.CellConfig{
			Formatting: tablewriter.CellFormatting{
				MergeMode: tw.MergeBoth,
			},
		},
		Footer: tablewriter.CellConfig{
			Formatting: tablewriter.CellFormatting{
				MergeMode: tw.MergeHorizontal,
			},
			ColumnAligns: []tw.Align{tw.Skip, tw.Skip, tw.AlignRight, tw.AlignLeft},
		},
	}

	r := renderer.NewDefault(renderer.DefaultConfig{
		Settings: renderer.Settings{
			Separators: renderer.Separators{
				BetweenRows: tw.On,
			},
		},
	})

	t.Run("mixed-1", func(t *testing.T) {
		buf.Reset()
		table := tablewriter.NewTable(&buf, tablewriter.WithConfig(c), tablewriter.WithRenderer(r))
		table.SetHeader([]string{"Date", "Description", "Status", "Conclusion"})
		table.Append([]string{"1/1/2014", "Domain name", "Successful", "Successful"})
		table.Append([]string{"1/1/2014", "Domain name", "Pending", "Waiting"})
		table.Append([]string{"1/1/2014", "Domain name", "Successful", "Rejected"})
		table.Append([]string{"TOTAL", "TOTAL", "TOTAL", "$145.93"})
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
			for _, v := range table.Debug() {
				t.Error(v)
			}
		}

	})

	t.Run("mixed-2", func(t *testing.T) {
		buf.Reset()
		table := tablewriter.NewTable(&buf, tablewriter.WithConfig(c), tablewriter.WithRenderer(r))
		table.SetHeader([]string{"Date", "Description", "Status", "Conclusion"})
		table.Append([]string{"1/1/2014", "Domain name", "Successful", "Successful"})
		table.Append([]string{"1/1/2014", "Domain name", "Pending", "Waiting"})
		table.Append([]string{"1/1/2015", "Domain name", "Successful", "Rejected"})
		table.Append([]string{"TOTAL", "TOTAL", "TOTAL", "$145.93"})
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
			for _, v := range table.Debug() {
				t.Error(v)
			}
		}

	})
}

func TestMergeHierarchical(t *testing.T) {
	var buf bytes.Buffer
	table := tablewriter.NewTable(&buf, tablewriter.WithConfig(tablewriter.Config{
		Row: tablewriter.CellConfig{
			Formatting: tablewriter.CellFormatting{
				MergeMode: tw.MergeHierarchical,
			},
		},
	}),
		tablewriter.WithRenderer(renderer.NewDefault(renderer.DefaultConfig{
			Symbols: tw.NewSymbols(tw.StyleASCII),
			Settings: renderer.Settings{
				Separators: renderer.Separators{
					BetweenRows: tw.On,
				},
			},
		})),
	)
	table.SetHeader([]string{"0", "1", "2", "3"})
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
		for _, v := range table.Debug() {
			t.Error(v)
		}
	}
}

func TestMergeHierarchicalUnicode(t *testing.T) {
	var buf bytes.Buffer
	table := tablewriter.NewTable(&buf, tablewriter.WithConfig(tablewriter.Config{
		Row: tablewriter.CellConfig{
			Formatting: tablewriter.CellFormatting{
				MergeMode: tw.MergeHierarchical,
			},
		},
	}),
		tablewriter.WithRenderer(renderer.NewDefault(renderer.DefaultConfig{
			// Symbols: tw.NewSymbols(tw.StyleRounded),
			Settings: renderer.Settings{
				Separators: renderer.Separators{
					BetweenRows: tw.On,
				},
			},
		})),
	)
	table.SetHeader([]string{"0", "1", "2", "3"})
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
		for _, v := range table.Debug() {
			t.Error(v)
		}
	}
}

func TestMergeWithPadding(t *testing.T) {
	var buf bytes.Buffer

	table := tablewriter.NewTable(&buf, tablewriter.WithConfig(tablewriter.Config{
		Row: tablewriter.CellConfig{
			Formatting: tablewriter.CellFormatting{
				MergeMode: tw.MergeBoth,
			},
			ColumnAligns: []tw.Align{tw.Skip, tw.Skip, tw.AlignRight, tw.AlignLeft},
		},
		Footer: tablewriter.CellConfig{
			Padding: tablewriter.CellPadding{
				Global:    tw.Padding{Left: "*", Right: "*", Top: "", Bottom: ""},
				PerColumn: []tw.Padding{tw.Padding{}, tw.Padding{}, tw.Padding{Bottom: "^"}, tw.Padding{Bottom: "."}},
			},
			ColumnAligns: []tw.Align{tw.Skip, tw.Skip, tw.AlignRight, tw.AlignLeft},
		},
	}), tablewriter.WithRenderer(renderer.NewDefault(renderer.DefaultConfig{
		//Symbols: tw.NewSymbols(tw.StyleASCII),
		Settings: renderer.Settings{
			Separators: renderer.Separators{
				BetweenRows: tw.On,
			},
		},
	})))

	table.SetHeader([]string{"Date", "Description", "Status", "Conclusion"})
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
