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
	visualCheck(t, "VerticalMerge", buf.String(), expected)
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
		Symbols: tw.NewSymbols(tw.StyleMerger),
		Settings: renderer.Settings{
			Separators: renderer.Separators{
				BetweenRows: tw.On,
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
	┌──────.───────────────────────.────────┐
	│ NAME │         SIGN          │ RATING │
	.──────.───────────────────────.────────.
	│ A    │ The Good              │    500 │
	│      .───────────────────────.────────.
	│      │ The Very very Bad Man │    288 │
	│──────.───────────────────────.────────.
	│ B    │                       │    120 │
	│      │                       .────────.
	│      │                       │    200 │
	└──────.───────────────────────.────────┘
`
	visualCheck(t, "VerticalMergeLines", buf.String(), expected)
}

func TestHorizontalMerge(t *testing.T) {
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
			Symbols: tw.NewSymbols(tw.StyleASCII),
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
	table.Append([]string{"", "", "TOTAL", "$145.93"})
	table.Render()

	expected := `
	+----------+-------------+------------+------------+
	|   DATE   | DESCRIPTION |   STATUS   | CONCLUSION |
	+----------+-------------+------------+------------+
	| 1/1/2014 | Domain name | Successful              |
	+----------+-------------+------------+------------+
	| 1/1/2014 | Domain name | Pending    | Waiting    |
	+----------+-------------+------------+------------+
	| 1/1/2014 | Domain name | Successful | Rejected   |
	+----------+-------------+------------+------------+
	|          |             | TOTAL      | $145.93    |
	+----------+-------------+------------+------------+
`
	visualCheck(t, "HorizontalMerge", buf.String(), expected)
}

//func TestMergeBoth(t *testing.T) {
//	var buf bytes.Buffer
//	table := NewTable(&buf, WithConfig(Config{
//		Row: CellConfig{
//			Formatting: CellFormatting{
//				MergeMode: MergeVertical,
//			},
//			ColumnAligns: []string{renderer.AlignLeft, renderer.AlignLeft, renderer.AlignRight, renderer.AlignRight},
//		},
//	}))
//	table.SetHeader([]string{"Name", "Sign", "Rating", "Score"})
//	table.Append([]string{"A", "The Good", "500", "500"})
//	table.Append([]string{"A", "The Very very Bad Man", "288", "120"})
//	table.Append([]string{"B", "", "120", "150"})
//	table.Append([]string{"B", "", "200", "530"})
//	table.Render()
//
//	expected := `
//	┌──────┬───────────────────────┬────────┬───────┐
//	│ NAME │         SIGN          │ RATING │ SCORE │
//	├──────┼───────────────────────┼────────┴───────┤
//	│ A    │ The Good              │            500 │
//	│      ├───────────────────────┼────────┬───────┤
//	│      │ The Very very Bad Man │    288 │   120 │
//	├──────┼───────────────────────┼────────┼───────┤
//	│ B    │                       │    120 │   150 │
//	│      │                       ├────────┼───────┤
//	│      │                       │    200 │   530 │
//	└──────┴───────────────────────┴────────┴───────┘
//`
//	visualCheck(t, "MergeBoth", buf.String(), expected)
//}
//
//func TestMergeHierarchical(t *testing.T) {
//	var buf bytes.Buffer
//	table := NewTable(&buf, WithConfig(Config{
//		Row: CellConfig{
//			Formatting: CellFormatting{
//				MergeMode: MergeHierarchical,
//			},
//		},
//	}),
//		WithRenderer(renderer.NewDefault(renderer.DefaultConfig{
//			Symbols: symbols.NewSymbols(symbols.StyleASCII),
//			Settings: renderer.Settings{
//				Separators: renderer.Separators{
//					BetweenRows: renderer.On,
//				},
//			},
//		})),
//	)
//	table.SetHeader([]string{"0", "1", "2", "3"})
//	table.Append([]string{"A", "a", "c", "-"})
//	table.Append([]string{"A", "b", "c", "-"})
//	table.Append([]string{"A", "b", "d", "-"})
//	table.Append([]string{"B", "b", "d", "-"})
//	table.Render()
//
//	expected := `
//		+---+---+---+---+
//		| 0 | 1 | 2 | 3 |
//		+---+---+---+---+
//		| A | a | c | - |
//		+   +---+---+---+
//		|   | b | c | - |
//		+   +   +---+---+
//		|   |   | d | - |
//		+---+---+---+---+
//		| B | b | d | - |
//		+---+---+---+---+
//`
//	visualCheck(t, "MergeHierarchical", buf.String(), expected)
//}
