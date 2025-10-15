package tests

import (
	"bytes"
	"io"
	"testing"

	"github.com/olekukonko/tablewriter"
	"github.com/olekukonko/tablewriter/renderer"
	"github.com/olekukonko/tablewriter/tw"
)

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
				AutoWrap: tw.WrapTruncate,
			},
			Merging: tw.CellMerging{
				Mode: tw.MergeHorizontal,
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
		Header:   tw.CellConfig{Merging: tw.CellMerging{Mode: tw.MergeHorizontal}},
		Behavior: tw.Behavior{Compact: tw.Compact{Merge: tw.On}},
		Debug:    true,
	}))

	x := "This is a long header that makes the table look too wide"
	table.Header([]string{x, x})
	table.Append([]string{"Key", "Value"})
	table.Render()

	expected := `
        ┌──────────────────────────────────────────────────────────┐
        │ THIS IS A LONG HEADER THAT MAKES THE TABLE LOOK TOO WIDE │
        ├────────────────────────────┬─────────────────────────────┤
        │ Key                        │ Value                       │
        └────────────────────────────┴─────────────────────────────┘
`
	if !visualCheck(t, "TestCompatMode", buf.String(), expected) {
		t.Error(table.Debug())
	}
}

func TestTrimLine(t *testing.T) {
	var buf bytes.Buffer

	table := tablewriter.NewTable(
		&buf,
		tablewriter.WithRenderer(
			renderer.NewBlueprint(tw.Rendition{
				Settings: tw.Settings{
					Separators: tw.Separators{BetweenRows: tw.On},
				},
			}),
		),
		tablewriter.WithRowAutoFormat(tw.WrapNone),
		tablewriter.WithTrimLine(tw.Off), // you will be able to do this
	)
	_ = table.Append([]string{"Row1", "Cell\n\n\nWith Newlines"})
	_ = table.Append([]string{"Row2", "Cell\nWith Newlines"})
	_ = table.Append([]string{"Row3", "Cell\n\nWith Newlines"})
	table.Render()

	expected := `
		┌──────┬───────────────┐
		│ Row1 │ Cell          │
		│      │               │
		│      │               │
		│      │ With Newlines │
		├──────┼───────────────┤
		│ Row2 │ Cell          │
		│      │ With Newlines │
		├──────┼───────────────┤
		│ Row3 │ Cell          │
		│      │               │
		│      │ With Newlines │
		└──────┴───────────────┘
`
	if !visualCheck(t, "TestCompatMode", buf.String(), expected) {
		t.Error(table.Debug())
	}
}

// A simple ByteCounter to demonstrate a custom counter implementation.
type ByteCounter struct {
	count int
}

func (bc *ByteCounter) Write(p []byte) (n int, err error) {
	bc.count += len(p)
	return len(p), nil
}
func (bc *ByteCounter) Total() int {
	return bc.count
}

// TestLinesCounter verifies the functionality of the WithLineCounter and WithCounters options.
func TestLinesCounter(t *testing.T) {

	data := [][]string{
		{"A", "The Good", "500"},
		{"B", "The Very Very Bad Man", "288"},
		{"C", "The Ugly", "120"},
		{"D", "The Gopher", "800"},
	}

	// Test Case 1: Default line counting on a standard table using the new API.
	t.Run("WithLineCounter", func(t *testing.T) {
		table := tablewriter.NewTable(io.Discard,
			tablewriter.WithLineCounter(), // Use the new, explicit function.
		)
		table.Header("Name", "Sign", "Rating")
		table.Bulk(data)
		table.Render()

		// Expected: 1 Top border + 1 Header + 1 Separator + 4 Rows + 1 Bottom border = 8
		expectedLines := 8
		if got := table.Lines(); got != expectedLines {
			t.Errorf("expected %d lines, but got %d", expectedLines, got)
		}
	})

	// Test Case 2: Line counting with auto-wrapping enabled.
	t.Run("LineCounterWithWrapping", func(t *testing.T) {
		table := tablewriter.NewTable(io.Discard,
			tablewriter.WithLineCounter(), // Use the new, explicit function.
			tablewriter.WithRowAutoWrap(tw.WrapNormal),
			tablewriter.WithMaxWidth(40),
		)
		table.Header("Name", "Sign", "Rating")
		table.Bulk(data)
		table.Render()

		// Expected: 1 Top border + 1 Header + 1 Separator + 1+3+1+1 Rows + 1 Bottom border = 10
		expectedLines := 10
		if got := table.Lines(); got != expectedLines {
			t.Errorf("expected %d lines with wrapping, but got %d", expectedLines, got)
		}
	})

	// Test Case 3: Ensure Lines() returns -1 when no counter is enabled at all.
	t.Run("NoCounters", func(t *testing.T) {
		table := tablewriter.NewTable(io.Discard) // No counter options
		table.Header("Name", "Sign")
		table.Append("A", "B")
		table.Render()

		expected := -1
		if got := table.Lines(); got != expected {
			t.Errorf("expected %d when no counter is used, but got %d", expected, got)
		}
	})

	// Test Case 4: Use a custom counter and verify it's retrieved via Counters().
	t.Run("WithCustomCounter", func(t *testing.T) {
		byteCounter := &ByteCounter{}
		var buf bytes.Buffer

		table := tablewriter.NewTable(&buf,
			tablewriter.WithCounters(byteCounter), // Use the new plural function for custom counters.
		)
		table.Header("A", "B")
		table.Append("1", "2")
		table.Render()

		// Crucial Test: Lines() should return -1 because no *LineCounter* was added.
		if got := table.Lines(); got != -1 {
			t.Errorf("expected Lines() to return -1 when only a custom counter is used, but got %d", got)
		}

		// Verify the custom counter via the Counters() method.
		allCounters := table.Counters()
		if len(allCounters) != 1 {
			t.Fatalf("expected 1 counter, but found %d", len(allCounters))
		}

		if custom, ok := allCounters[0].(*ByteCounter); ok {
			if custom.Total() <= 0 {
				t.Errorf("expected a positive byte count from custom counter, but got %d", custom.Total())
			}
			if custom.Total() != buf.Len() {
				t.Errorf("byte counter total (%d) does not match buffer length (%d)", custom.Total(), buf.Len())
			}
		} else {
			t.Error("expected the first counter to be of type *ByteCounter")
		}
	})

	// Test Case 5: Ensure Lines() finds the line counter even when mixed with others.
	t.Run("LinesWithMixedCounters", func(t *testing.T) {
		byteCounter := &ByteCounter{}

		// Add counters in a specific order: custom first, then default.
		table := tablewriter.NewTable(io.Discard,
			tablewriter.WithCounters(byteCounter),
			tablewriter.WithLineCounter(),
		)
		table.Header("Name", "Sign", "Rating")
		table.Bulk(data)
		table.Render()

		// Lines() should still find the line count correctly, regardless of order.
		expectedLines := 8
		if got := table.Lines(); got != expectedLines {
			t.Errorf("expected %d lines even with mixed counters, but got %d", expectedLines, got)
		}
	})
}
