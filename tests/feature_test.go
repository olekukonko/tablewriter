package tests

import (
	"bytes"
	"github.com/olekukonko/tablewriter"
	"github.com/olekukonko/tablewriter/renderer"
	"github.com/olekukonko/tablewriter/tw"
	"testing"
)

func TestBug260(t *testing.T) {
	var buf bytes.Buffer

	tableRendition := tw.Rendition{
		Borders: tw.BorderNone,
		Settings: tw.Settings{
			Separators: tw.Separators{
				ShowHeader:     tw.Off,
				ShowFooter:     tw.Off,
				BetweenRows:    tw.Off,
				BetweenColumns: tw.Off,
			},
		},
		Symbols: tw.NewSymbols(tw.StyleNone),
	}

	t.Run("Normal", func(t *testing.T) {
		buf.Reset()
		tableRenderer := renderer.NewBlueprint(tableRendition)
		table := tablewriter.NewTable(
			&buf,
			tablewriter.WithRenderer(tableRenderer),
			tablewriter.WithTableMax(120),
			tablewriter.WithTrimSpace(tw.Off),
			tablewriter.WithDebug(true),
			tablewriter.WithPadding(tw.PaddingNone),
		)

		table.Append([]string{"INFO:",
			"The original machine had a base-plate of prefabulated aluminite, surmounted by a malleable logarithmic casing in such a way that the two main spurving bearings were in a direct line with the pentametric fan.",
		})

		table.Append("INFO:",
			"The original machine had a base-plate of prefabulated aluminite, surmounted by a malleable logarithmic casing in such a way that the two main spurving bearings were in a direct line with the pentametric fan.",
		)

		table.Render()

		expected := `
		INFO:The original machine had a base-plate of prefabulated     
			 aluminite, surmounted by a malleable logarithmic casing in
			 such a way that the two main spurving bearings were in a  
			 direct line with the pentametric fan.                     
		INFO:The original machine had a base-plate of prefabulated     
			 aluminite, surmounted by a malleable logarithmic casing in
			 such a way that the two main spurving bearings were in a  
			 direct line with the pentametric fan.    
`
		debug := visualCheck(t, "TestBug252-Mixed", buf.String(), expected)
		if !debug {
			t.Error(table.Debug())
		}

	})

	t.Run("Mixed", func(t *testing.T) {
		buf.Reset()
		tableRenderer := renderer.NewBlueprint(tableRendition)
		table := tablewriter.NewTable(
			&buf,
			tablewriter.WithRenderer(tableRenderer),
			tablewriter.WithTableMax(120),
			tablewriter.WithTrimSpace(tw.Off),
			tablewriter.WithDebug(true),
			tablewriter.WithPadding(tw.PaddingNone),
		)

		table.Append([]string{"INFO:",
			"The original machine had a base-plate of prefabulated aluminite, surmounted by a malleable logarithmic casing in such a way that the two main spurving bearings were in a direct line with the pentametric fan.",
		})

		table.Append("INFO: ",
			"The original machine had a base-plate of prefabulated aluminite, surmounted by a malleable logarithmic casing in such a way that the two main spurving bearings were in a direct line with the pentametric fan.",
		)

		table.Render()

		expected := `
		INFO: The original machine had a base-plate of prefabulated     
			  aluminite, surmounted by a malleable logarithmic casing in
			  such a way that the two main spurving bearings were in a  
			  direct line with the pentametric fan.                     
		INFO: The original machine had a base-plate of prefabulated     
			  aluminite, surmounted by a malleable logarithmic casing in
			  such a way that the two main spurving bearings were in a  
			  direct line with the pentametric fan.  
`
		debug := visualCheck(t, "TestBug252-Mixed", buf.String(), expected)
		if !debug {
			t.Error(table.Debug())
		}

	})

}

func TestBatchPerColumnWidths(t *testing.T) {
	var buf bytes.Buffer
	table := tablewriter.NewTable(&buf, tablewriter.WithConfig(tablewriter.Config{
		Widths: tw.CellWidth{
			PerColumn: tw.NewMapper[int, int]().Set(0, 8).Set(1, 10).Set(2, 15), // Total widths: 8, 5, 15
		},
		Row: tw.CellConfig{
			Formatting: tw.CellFormatting{
				AutoWrap: tw.WrapTruncate, // Truncate content to fit
			},
		},
	}), tablewriter.WithRenderer(renderer.NewBlueprint(tw.Rendition{
		Settings: tw.Settings{
			Separators: tw.Separators{
				BetweenColumns: tw.On, // Separator width = 1
			},
		},
	})))

	table.Header([]string{"Name", "Age", "City"})
	table.Append([]string{"Alice Smith", "25", "New York City"})
	table.Append([]string{"Bob Johnson", "30", "Boston"})
	table.Render()

	// Expected widths:
	// Col 0: 8 (content=6, pad=1+1, sep=1 for next column)
	// Col 1: 5 (content=3, pad=1+1, sep=1 for next column)
	// Col 2: 15 (content=13, pad=1+1, no sep at end)
	expected := `
	┌────────┬──────────┬───────────────┐
	│  NAME  │   AGE    │     CITY      │
	├────────┼──────────┼───────────────┤
	│ Alic…  │ 25       │ New York City │
	│ Bob …  │ 30       │ Boston        │
	└────────┴──────────┴───────────────┘

`
	if !visualCheck(t, "BatchPerColumnWidths", buf.String(), expected) {
		t.Error(table.Debug())
	}
}

func TestBatchGlobalWidthScaling(t *testing.T) {
	var buf bytes.Buffer
	table := tablewriter.NewTable(&buf, tablewriter.WithConfig(tablewriter.Config{
		Widths: tw.CellWidth{
			Global: 20, // Total table width, including padding and separators
		},
		Row: tw.CellConfig{
			Formatting: tw.CellFormatting{
				AutoWrap: tw.WrapNormal,
			},
		},
	}), tablewriter.WithRenderer(renderer.NewBlueprint(tw.Rendition{
		Settings: tw.Settings{
			Separators: tw.Separators{
				BetweenColumns: tw.On, // Separator width = 1
			},
		},
	})))

	table.Header([]string{"Name", "Age", "City"})
	table.Append([]string{"Alice Smith", "25", "New York City"})
	table.Append([]string{"Bob Johnson", "30", "Boston"})
	table.Render()

	// Expected widths:
	// Total width = 20, with 2 separators (2x1 = 2)
	// Available for columns = 20 - 2 = 18
	// 3 columns, so each ~6 (18/3), adjusted for padding and separators
	// Col 0: 6 (content=4, pad=1+1, sep=1)
	// Col 1: 6 (content=4, pad=1+1, sep=1)
	// Col 2: 6 (content=4, pad=1+1)
	expected := `
	┌──────┬─────┬───────┐
	│ NAME │ AGE │ CITY  │
	├──────┼─────┼───────┤
	│ Alic │ 25  │ New Y │
	│ Bob  │ 30  │ Bosto │
	└──────┴─────┴───────┘
`
	if !visualCheck(t, "BatchGlobalWidthScaling", buf.String(), expected) {
		t.Error(table.Debug())
	}
}

func TestBatchWidthsWithHorizontalMerge(t *testing.T) {
	var buf bytes.Buffer
	table := tablewriter.NewTable(&buf, tablewriter.WithConfig(tablewriter.Config{
		Widths: tw.CellWidth{
			PerColumn: tw.NewMapper[int, int]().Set(0, 10).Set(1, 8).Set(2, 8), // Total widths: 10, 8, 8
		},
		Row: tw.CellConfig{
			Formatting: tw.CellFormatting{
				MergeMode: tw.MergeHorizontal,
				AutoWrap:  tw.WrapTruncate,
			},
		},
	}), tablewriter.WithRenderer(renderer.NewBlueprint(tw.Rendition{
		Settings: tw.Settings{
			Separators: tw.Separators{
				BetweenColumns: tw.On, // Separator width = 1
				BetweenRows:    tw.On,
			},
		},
	})))

	table.Header([]string{"Name", "Status", "Status"})
	table.Append([]string{"Alice", "Active", "Active"})  // Should merge Status columns
	table.Append([]string{"Bob", "Inactive", "Pending"}) // No merge
	table.Render()

	// Expected widths:
	// Col 0: 10 (content=8, pad=1+1, sep=1)
	// Col 1: 8 (content=6, pad=1+1, sep=1)
	// Col 2: 8 (content=6, pad=1+1)
	// Merged Col 1+2: 8 + 8 - 1 (no separator between) = 15 (content=13, pad=1+1)
	expected := `
        ┌──────────┬────────┬────────┐
        │   NAME   │ STATUS │ STATUS │
        ├──────────┼────────┴────────┤
        │ Alice    │ Active          │
        ├──────────┼────────┬────────┤
        │ Bob      │ Inac…  │ Pend…  │
        └──────────┴────────┴────────┘
`
	if !visualCheck(t, "BatchWidthsWithHorizontalMerge", buf.String(), expected) {
		t.Error(table.Debug())
	}
}

func TestWrapBreakWithConstrainedWidthsNoRightPadding(t *testing.T) {
	var buf bytes.Buffer
	table := tablewriter.NewTable(&buf,
		tablewriter.WithTrimSpace(tw.Off),
		tablewriter.WithHeaderAutoFormat(tw.Off),
		tablewriter.WithConfig(tablewriter.Config{
			Widths: tw.CellWidth{
				PerColumn: tw.NewMapper[int, int]().Set(0, 4).Set(1, 4).Set(2, 6).Set(3, 7),
			},
			Row: tw.CellConfig{
				Formatting: tw.CellFormatting{
					AutoWrap: tw.WrapBreak,
				},
				Padding: tw.CellPadding{
					Global: tw.PaddingNone,
				},
			},
		}),
	)

	headers := []string{"a", "b", "c", "d"}
	table.Header(headers)

	data := [][]string{
		{"aa", "bb", "cc", "dd"},
		{"aaa", "bbb", "ccc", "ddd"},
		{"aaaa", "bbbb", "cccc", "dddd"},
		{"aaaaa", "bbbbb", "ccccc", "ddddd"},
	}
	table.Bulk(data)

	table.Render()

	expected := `
	┌────┬────┬──────┬───────┐
	│ A  │ B  │  C   │   D   │
	├────┼────┼──────┼───────┤
	│aa  │bb  │cc    │dd     │
	│aaa │bbb │ccc   │ddd    │
	│aaaa│bbbb│cccc  │dddd   │
	│aaa↩│bbb↩│ccccc │ddddd  │
	│aa  │bb  │      │       │
	└────┴────┴──────┴───────┘
`
	if !visualCheck(t, "WrapBreakWithConstrainedWidthsNoRightPadding", buf.String(), expected) {
		t.Error(table.Debug())
	}
}

func TestCompatMode(t *testing.T) {
	var buf bytes.Buffer
	table := tablewriter.NewTable(&buf, tablewriter.WithConfig(tablewriter.Config{
		Header:   tw.CellConfig{Formatting: tw.CellFormatting{MergeMode: tw.MergeHorizontal}},
		Behavior: tw.Behavior{Compact: tw.Compact{Merge: tw.On}},
	}))

	x := "This is a long header that makes the table look too wide"
	table.Header([]string{x, x})
	table.Append([]string{"Key", "Value"})
	table.Render()

	expected := `
	┌──────────────────────────────────────────────────────────┐
	│ THIS IS A LONG HEADER THAT MAKES THE TABLE LOOK TOO WIDE │
	├───────────────────────────┬──────────────────────────────┤
	│ Key                       │ Value                        │
	└───────────────────────────┴──────────────────────────────┘
`
	if !visualCheck(t, "BatchGlobalWidthScaling", buf.String(), expected) {
		t.Error(table.Debug())
	}
}
