package tests

import (
	"bytes"
	"fmt"
	"github.com/olekukonko/tablewriter"
	"github.com/olekukonko/tablewriter/renderer"
	"github.com/olekukonko/tablewriter/tw"
	"strings"
	"testing"
	"time"
	"unicode"
)

// createStreamTable creates a TableStream with the  renderer for testing.
func createStreamTable(t *testing.T, w *bytes.Buffer, opts ...tablewriter.Option) *tablewriter.Table {
	t.Helper()
	opts = append(opts, tablewriter.WithRenderer(renderer.NewBlueprint()))
	return tablewriter.NewTable(w, opts...)
}

func TestStreamTableDefault(t *testing.T) {
	var buf bytes.Buffer

	t.Run("disabled", func(t *testing.T) {
		buf.Reset()
		table := tablewriter.NewTable(&buf, tablewriter.WithStreaming(tw.StreamConfig{Enable: false}))

		//err := table.Start()
		//if err != nil {
		//	t.Fatalf("Start failed: %v", err)
		//}

		table.Header([]string{"Name", "Age", "City"})
		table.Append([]string{"Alice", "25", "New York"})
		table.Append([]string{"Bob", "30", "Boston"})

		//err = table.Close()
		//if err != nil {
		//	t.Fatalf("End failed: %v", err)
		//}

		table.Render()
		expected := `
	┌───────┬─────┬──────────┐
	│ NAME  │ AGE │   CITY   │
	├───────┼─────┼──────────┤
	│ Alice │ 25  │ New York │
	│ Bob   │ 30  │ Boston   │
	└───────┴─────┴──────────┘
`
		debug := visualCheck(t, "BasicTableRendering", buf.String(), expected)
		if !debug {
			t.Error(table.Debug().String())
		}
	})

	t.Run("enabled", func(t *testing.T) {
		buf.Reset()
		table := tablewriter.NewTable(&buf, tablewriter.WithStreaming(tw.StreamConfig{Enable: true}), tablewriter.WithDebug(false))

		err := table.Start()
		if err != nil {
			t.Fatalf("Start failed: %v", err)
		}

		table.Header([]string{"Name", "Age", "City"})
		table.Append([]string{"Alice", "25", "New York"})
		table.Append([]string{"Bob", "30", "Boston"})

		err = table.Close()
		if err != nil {
			t.Fatalf("End failed: %v", err)
		}

		expected := `
		┌────────┬────────┬────────┐
		│  NAME  │  AGE   │  CITY  │
		├────────┼────────┼────────┤
		│ Alice  │ 25     │ New    │
		│        │        │ York   │
		│ Bob    │ 30     │ Boston │
		└────────┴────────┴────────┘
`
		debug := visualCheck(t, "BasicTableRendering", buf.String(), expected)
		if !debug {
			t.Error(table.Debug().String())
		}
	})

}

// TestStreamBasic tests basic streaming table rendering with header, rows, and footer.
func TestStreamBasic(t *testing.T) {
	var buf bytes.Buffer

	t.Run("TestStreamBasic", func(t *testing.T) {
		buf.Reset()
		st := createStreamTable(t, &buf,
			tablewriter.WithConfig(tablewriter.Config{
				Header: tw.CellConfig{Formatting: tw.CellFormatting{Alignment: tw.AlignCenter}},
				Row:    tw.CellConfig{Formatting: tw.CellFormatting{Alignment: tw.AlignLeft}},
				Footer: tw.CellConfig{Formatting: tw.CellFormatting{Alignment: tw.AlignLeft}}}),
			tablewriter.WithDebug(false),
			tablewriter.WithStreaming(tw.StreamConfig{Enable: true}),
		)

		err := st.Start()
		if err != nil {
			t.Fatalf("Start failed: %v", err)
		}
		st.Header([]string{"Name", "Age", "City"})
		st.Append([]string{"Alice", "25", "New York"})
		st.Append([]string{"Bob", "30", "Boston"})
		st.Footer([]string{"Total", "55", "*"})
		err = st.Close()
		if err != nil {
			t.Fatalf("End failed: %v", err)
		}

		// Widths: Name(5)+2=7, Age(3)+2=5, City(8)+2=10
		expected := `
		┌────────┬────────┬────────┐
		│  NAME  │  AGE   │  CITY  │
		├────────┼────────┼────────┤
		│ Alice  │ 25     │ New    │
		│        │        │ York   │
		│ Bob    │ 30     │ Boston │
		├────────┼────────┼────────┤
		│ Total  │ 55     │ *      │
		└────────┴────────┴────────┘
`
		if !visualCheck(t, "StreamBasic", buf.String(), expected) {
			fmt.Println(st.Debug())
			t.Fail()
		}

	})

	t.Run("TestStreamBasicGlobal", func(t *testing.T) {
		buf.Reset()
		st := createStreamTable(t, &buf,
			tablewriter.WithConfig(tablewriter.Config{
				Header: tw.CellConfig{Formatting: tw.CellFormatting{Alignment: tw.AlignCenter}},
				Row:    tw.CellConfig{Formatting: tw.CellFormatting{Alignment: tw.AlignLeft}},
				Footer: tw.CellConfig{Formatting: tw.CellFormatting{Alignment: tw.AlignRight}}}),
			tablewriter.WithDebug(false),
			tablewriter.WithStreaming(tw.StreamConfig{Enable: true}),
		)

		err := st.Start()
		if err != nil {
			t.Fatalf("Start failed: %v", err)
		}
		st.Header([]string{"Name", "Age", "City"})
		st.Append([]string{"Alice", "25", "New York"})
		st.Append([]string{"Bob", "30", "Boston"})
		st.Footer([]string{"Total", "55", "*"})

		err = st.Close()
		if err != nil {
			t.Fatalf("End failed: %v", err)
		}

		// Widths: Name(5)+2=7, Age(3)+2=5, City(8)+2=10
		expected := `
		┌────────┬────────┬────────┐
		│  NAME  │  AGE   │  CITY  │
		├────────┼────────┼────────┤
		│ Alice  │ 25     │ New    │
		│        │        │ York   │
		│ Bob    │ 30     │ Boston │
		├────────┼────────┼────────┤
		│  Total │     55 │      * │
		└────────┴────────┴────────┘
`
		if !visualCheck(t, "StreamBasic", buf.String(), expected) {
			fmt.Println(st.Debug())
			t.Fail()
		}

	})

}

// TestStreamWithFooterAlign tests streaming table with footer and custom alignments.
func TestStreamWithFooterAlign(t *testing.T) {
	var buf bytes.Buffer
	st := createStreamTable(t, &buf, tablewriter.WithConfig(tablewriter.Config{
		Header: tw.CellConfig{Formatting: tw.CellFormatting{Alignment: tw.AlignCenter}},
		Row: tw.CellConfig{
			Formatting:   tw.CellFormatting{Alignment: tw.AlignLeft},
			ColumnAligns: []tw.Align{tw.AlignLeft, tw.AlignCenter, tw.AlignRight},
		},
		Footer: tw.CellConfig{
			Formatting:   tw.CellFormatting{Alignment: tw.AlignRight},
			ColumnAligns: []tw.Align{tw.AlignLeft, tw.AlignCenter, tw.AlignRight},
		},
	}),
		tablewriter.WithDebug(false),
		tablewriter.WithStreaming(tw.StreamConfig{Enable: true}))

	err := st.Start()
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	st.Header([]string{"Item", "Qty", "Price"})   // Widths: 4+2=6, 3+2=5, 5+2=7
	st.Append([]string{"Item 1", "2", "1000.00"}) // Needs: 6+2=8, 1+2=3, 7+2=9
	st.Append([]string{"Item 2", "10", "25.50"})  // Needs: 6+2=8, 2+2=4, 5+2=7
	st.Footer([]string{"", "Total", "1025.50"})   // Needs: 0+2=2, 5+2=7, 7+2=9
	err = st.Close()
	if err != nil {
		t.Fatalf("End failed: %v", err)
	}

	// Max widths: [8, 7, 9]
	expected := `
		┌────────┬────────┬─────────┐
		│  ITEM  │  QTY   │  PRICE  │
		├────────┼────────┼─────────┤
		│ Item 1 │   2    │ 1000.00 │
		│ Item 2 │   10   │   25.50 │
		├────────┼────────┼─────────┤
		│        │ Total  │ 1025.50 │
		└────────┴────────┴─────────┘
`
	if !visualCheck(t, "StreamWithFooterAlign", buf.String(), expected) {
		fmt.Println("--- DEBUG LOG ---")
		fmt.Println(st.Debug().String())
		t.Fail()
	}
}

// TestStreamNoHeaderASCII tests streaming table without header using ASCII symbols.
func TestStreamNoHeaderASCII(t *testing.T) {
	var buf bytes.Buffer
	st := tablewriter.NewTable(&buf,
		tablewriter.WithConfig(tablewriter.Config{Row: tw.CellConfig{Formatting: tw.CellFormatting{Alignment: tw.AlignLeft}}}),
		tablewriter.WithRenderer(renderer.NewBlueprint(tw.Rendition{Symbols: tw.NewSymbols(tw.StyleASCII)})),
		tablewriter.WithDebug(false),
		tablewriter.WithStreaming(tw.StreamConfig{Enable: true}),
	)
	err := st.Start()
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	st.Append([]string{"Regular", "line", "1"}) // Widths: 7+2=9, 4+2=6, 1+2=3
	st.Append([]string{"Thick", "thick", "2"})

	err = st.Close()
	if err != nil {
		t.Fatalf("End failed: %v", err)
	}

	expected := `
	+-----------+--------+--------+
	| Regular   | line   | 1      |
	| Thick     | thick  | 2      |
	+-----------+--------+--------+
`
	if !visualCheck(t, "StreamNoHeaderASCII", buf.String(), expected) {
		fmt.Println(st.Debug().String())
		t.Fail()
	}
}

func TestBorders(t *testing.T) {
	data := [][]string{{"A", "B"}, {"C", "D"}}
	//widths := map[int]int{0: 3, 1: 3} // Content (1) + padding (1+1) = 3

	tests := []struct {
		name     string
		borders  tw.Border
		expected string
	}{
		{
			name:    "All Off",
			borders: tw.Border{Left: tw.Off, Right: tw.Off, Top: tw.Off, Bottom: tw.Off},
			expected: `
A │ B
C │ D
`,
		},
		{
			name:    "No Left/Right",
			borders: tw.Border{Left: tw.Off, Right: tw.Off, Top: tw.On, Bottom: tw.On},
			expected: `
───┬───
A │ B
C │ D
───┴───
`,
		},
		{
			name:    "No Top/Bottom",
			borders: tw.Border{Left: tw.On, Right: tw.On, Top: tw.Off, Bottom: tw.Off},
			expected: `
│ A │ B │
│ C │ D │
`,
		},
		{
			name:    "Only Left",
			borders: tw.Border{Left: tw.On, Right: tw.Off, Top: tw.Off, Bottom: tw.Off},
			expected: `
│ A │ B
│ C │ D
`,
		},
		{
			name:    "Default",
			borders: tw.Border{Left: tw.On, Right: tw.On, Top: tw.On, Bottom: tw.On},
			expected: `
┌───┬───┐
│ A │ B │
│ C │ D │
└───┴───┘
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer

			r := renderer.NewBlueprint(
				tw.Rendition{
					Borders: tt.borders,
				},
			)
			st := tablewriter.NewTable(&buf,
				tablewriter.WithConfig(tablewriter.Config{Row: tw.CellConfig{Formatting: tw.CellFormatting{Alignment: tw.AlignLeft}}}),
				tablewriter.WithRenderer(r),
				tablewriter.WithDebug(false),
			)

			st.Append(data[0])
			st.Append(data[1])
			err := st.Render()
			if err != nil {
				t.Fatalf("End failed: %v", err)
			}

			if !visualCheck(t, "StreamBorders_"+tt.name, buf.String(), tt.expected) {
				fmt.Printf("--- DEBUG LOG for %s ---\n", tt.name)
				fmt.Println(st.Debug().String())
				t.Fail()
			}
		})
	}
}

// TestStreamTruncation tests streaming table with long content truncation.
func TestStreamTruncation(t *testing.T) {
	var buf bytes.Buffer
	st := createStreamTable(t, &buf,
		tablewriter.WithConfig(
			tablewriter.Config{
				Header: tw.CellConfig{Formatting: tw.CellFormatting{Alignment: tw.AlignCenter}},
				Row: tw.CellConfig{
					Formatting: tw.CellFormatting{Alignment: tw.AlignLeft, AutoWrap: tw.WrapTruncate, MaxWidth: 13},
				},
				Stream: tw.StreamConfig{
					Enable: true,
					Widths: tw.CellWidth{
						PerColumn: map[int]int{0: 4, 1: 15, 2: 8},
					},
				},
			}))

	err := st.Start()
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	st.Header([]string{"ID", "Description", "Status"})                     // Fits: 2<=4, 11<=15, 6<=8
	st.Append([]string{"1", "This description is quite long", "OK"})       // Truncates: 1<=4, 30>15 -> "This descript…", 2<=8
	st.Append([]string{"2", "Short desc", "DONE"})                         // Fits: 1<=4, 10<=15, 4<=8
	st.Append([]string{"3", "Another long one needing truncation", "ERR"}) // Truncates: 1<=4, 35>15 -> "Another long …", 3<=8
	err = st.Close()
	if err != nil {
		t.Fatalf("End failed: %v", err)
	}

	// Widths: [4, 15, 8]
	expected := `
        ┌────┬───────────────┬────────┐
        │ ID │  DESCRIPTION  │ STATUS │
        ├────┼───────────────┼────────┤
        │ 1  │ This descri…  │ OK     │
        │ 2  │ Short desc    │ DONE   │
        │ 3  │ Another lon…  │ ERR    │
        └────┴───────────────┴────────┘
`
	if !visualCheck(t, "StreamTruncation", buf.String(), expected) {
		fmt.Println("--- DEBUG LOG ---")
		fmt.Println(st.Debug().String())
		t.Fail()
	}
}

// TestStreamCustomPadding tests streaming table with custom padding.
func TestStreamCustomPadding(t *testing.T) {
	var buf bytes.Buffer
	st := createStreamTable(t, &buf, tablewriter.WithConfig(tablewriter.Config{
		Header: tw.CellConfig{
			Padding: tw.CellPadding{Global: tw.Padding{Left: ">>", Right: "<<"}},
		},
		Row: tw.CellConfig{
			Padding: tw.CellPadding{Global: tw.Padding{Left: ">>", Right: "<<"}},
		},
		Stream: tw.StreamConfig{
			Enable: true,
			Widths: tw.CellWidth{
				PerColumn: map[int]int{0: 7, 1: 7},
			},
		},
	}),
		tablewriter.WithDebug(false))

	err := st.Start()
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	st.Header([]string{"Head1", "Head2"}) // Truncates: 5>3 -> "He…"
	st.Append([]string{"R1C1", "R1C2"})   // Truncates: 4>3 -> "R1…"
	err = st.Close()
	if err != nil {
		t.Fatalf("End failed: %v", err)
	}

	expected := `
		┌───────┬───────┐
		│>>H…<<<│>>H…<<<│
		├───────┼───────┤
		│>>R1C<<│>>R1C<<│
		└───────┴───────┘
`
	if !visualCheck(t, "StreamCustomPadding", buf.String(), expected) {
		fmt.Println("--- DEBUG LOG ---")
		fmt.Println(st.Debug().String())
		t.Fail()
	}
}

// TestStreamEmptyCells tests streaming table with empty and sparse cells.
func TestStreamEmptyCells(t *testing.T) {
	var buf bytes.Buffer
	st := createStreamTable(t, &buf, tablewriter.WithConfig(tablewriter.Config{
		Header: tw.CellConfig{Formatting: tw.CellFormatting{Alignment: tw.AlignCenter}},
		Row:    tw.CellConfig{Formatting: tw.CellFormatting{Alignment: tw.AlignLeft}},
		Stream: tw.StreamConfig{
			Enable: true,
			Widths: tw.CellWidth{
				Global: 20,
			}},
	}),
		tablewriter.WithDebug(false))

	err := st.Start()
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	st.Header([]string{"H1", "", "H3"})     // Widths: 2+2=4, 0+2=2->3, 2+2=4
	st.Append([]string{"", "R1C2", ""})     // Needs: 0+2=2, 4+2=6, 0+2=2
	st.Append([]string{"R2C1", "", "R2C3"}) // Needs: 4+2=6, 0+2=2, 4+2=6
	st.Append([]string{"", "", ""})         // Needs: 0+2=2, 0+2=2, 0+2=2
	err = st.Close()
	if err != nil {
		t.Fatalf("End failed: %v", err)
	}

	// Max widths: [6, 6, 6]
	expected := `
		┌──────┬──────┬──────┐
		│ H 1  │      │ H 3  │
		├──────┼──────┼──────┤
		│      │ R1C2 │      │
		│ R2C1 │      │ R2C3 │
		│      │      │      │
		└──────┴──────┴──────┘
`
	if !visualCheck(t, "StreamEmptyCells", buf.String(), expected) {
		fmt.Println("--- DEBUG LOG ---")
		fmt.Println(st.Debug().String())
		t.Fail()
	}
}

// TestStreamOnlyHeader tests streaming table with only a header.
func TestStreamOnlyHeader(t *testing.T) {
	var buf bytes.Buffer
	st := createStreamTable(t, &buf, tablewriter.WithConfig(tablewriter.Config{
		Header: tw.CellConfig{Formatting: tw.CellFormatting{Alignment: tw.AlignCenter}},
	}),
		tablewriter.WithDebug(false),
		tablewriter.WithStreaming(tw.StreamConfig{Enable: true}))

	err := st.Start()
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	st.Header([]string{"Header1", "Header2"}) // Widths: 7+2=9, 7+2=9
	err = st.Close()
	if err != nil {
		t.Fatalf("End failed: %v", err)
	}

	expected := `
        ┌───────────┬───────────┐
        │ HEADER 1  │ HEADER 2  │
        └───────────┴───────────┘
`
	if !visualCheck(t, "StreamOnlyHeader", buf.String(), expected) {
		fmt.Println("--- DEBUG LOG ---")
		fmt.Println(st.Debug().String())
		t.Fail()
	}
}

// TestStreamOnlyHeaderNoHeaderLine tests streaming table with only a header and no header line.
func TestStreamOnlyHeaderNoHeaderLine(t *testing.T) {
	var buf bytes.Buffer
	st := createStreamTable(t, &buf, tablewriter.WithConfig(tablewriter.Config{
		Header: tw.CellConfig{Formatting: tw.CellFormatting{Alignment: tw.AlignCenter}},
	}),
		tablewriter.WithStreaming(tw.StreamConfig{Enable: true}),
	)

	err := st.Start()
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	st.Header([]string{"Header1", "Header2"}) // Fits: 7<=9, 7<=9
	err = st.Close()
	if err != nil {
		t.Fatalf("End failed: %v", err)
	}

	expected := `
        ┌───────────┬───────────┐
        │ HEADER 1  │ HEADER 2  │
        └───────────┴───────────┘
`
	if !visualCheck(t, "StreamOnlyHeaderNoHeaderLine", buf.String(), expected) {
		fmt.Println("--- DEBUG LOG ---")
		fmt.Println(st.Debug().String())
		t.Fail()
	}
}

// TestStreamSlowOutput tests streaming table with incremental output and delays.
func TestStreamSlowOutput(t *testing.T) {
	var buf bytes.Buffer
	st := createStreamTable(t, &buf, tablewriter.WithConfig(tablewriter.Config{
		Header: tw.CellConfig{Formatting: tw.CellFormatting{Alignment: tw.AlignCenter}},
		Row:    tw.CellConfig{Formatting: tw.CellFormatting{Alignment: tw.AlignLeft}},
	}),
		tablewriter.WithStreaming(tw.StreamConfig{Enable: true}),
	)

	err := st.Start()
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	fmt.Println("Output after Start():")
	fmt.Print(buf.String())
	buf.Reset()
	time.Sleep(100 * time.Millisecond)

	st.Header([]string{"Event", "Timestamp"}) // Widths: 5+2=7, 9+2=11
	fmt.Println("\nOutput after Header():")
	fmt.Print(buf.String())
	buf.Reset()
	time.Sleep(100 * time.Millisecond)

	for i := 1; i <= 3; i++ {
		err = st.Append([]string{fmt.Sprintf("Data Row %d", i), time.Now().Format("15:04:05.000")})
		if err != nil {
			t.Fatalf("Row %d failed: %v", i, err)
		}
		fmt.Printf("\nOutput after Row %d:\n", i)
		fmt.Print(buf.String())
		buf.Reset()
		time.Sleep(100 * time.Millisecond)
	}

	err = st.Close()
	if err != nil {
		t.Fatalf("End failed: %v", err)
	}
	fmt.Println("\nOutput after End():")
	fmt.Print(buf.String())

	t.Log("Slow stream test completed. Observe terminal output.")
	if st.Logger().Len() > 0 {
		//fmt.Println("--- DEBUG LOG ---")
		//fmt.Println(st.Debug().String())
	}
}

type Name struct {
	First string
	Last  string
}

// this will be ignored since  Format() is present
func (n Name) String() string {
	return fmt.Sprintf("%s %s", n.First, n.Last)
}

// Note: Format() overrides String() if both exist.
func (n Name) Format() string {
	return fmt.Sprintf("%s %s", n.clean(n.First), n.clean(n.Last))
}

// clean ensures the first letter is capitalized and the rest are lowercase
func (n Name) clean(s string) string {
	s = strings.TrimSpace(strings.ToLower(s))
	words := strings.Fields(s)
	s = strings.Join(words, "")

	if s == "" {
		return s
	}
	// Capitalize the first letter
	runes := []rune(s)
	runes[0] = unicode.ToUpper(runes[0])
	return string(runes)
}

type Age int

// Age int will be ignore and string will be used
func (a Age) String() string {
	return fmt.Sprintf("%d yrs", a)
}

func TestStreamFormating(t *testing.T) {
	var buf bytes.Buffer

	st := createStreamTable(t, &buf,
		tablewriter.WithConfig(tablewriter.Config{
			Header: tw.CellConfig{Formatting: tw.CellFormatting{Alignment: tw.AlignCenter}},
			Row:    tw.CellConfig{Formatting: tw.CellFormatting{Alignment: tw.AlignLeft}},
			Footer: tw.CellConfig{Formatting: tw.CellFormatting{Alignment: tw.AlignLeft}}}),
		tablewriter.WithDebug(false),
		tablewriter.WithStreaming(tw.StreamConfig{
			Enable: true,
			Widths: tw.CellWidth{PerColumn: map[int]int{0: 12, 1: 8, 2: 10}},
		}),
	)

	err := st.Start()
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	data := [][]any{
		{Name{"Al  i  CE", " Ma  SK"}, Age(25), "New York"},
		{Name{"bOb", "mar   le   y"}, Age(30), "Boston"},
	}

	st.Header([]string{"Name", "Age", "City"})
	st.Bulk(data)

	err = st.Close()
	if err != nil {
		t.Fatalf("End failed: %v", err)
	}

	// Widths: Name(5)+2=7, Age(3)+2=5, City(8)+2=10
	expected := `
		┌────────────┬────────┬──────────┐
		│    NAME    │  AGE   │   CITY   │
		├────────────┼────────┼──────────┤
		│ Alice Mask │ 25 yrs │ New York │
		│ Bob Marley │ 30 yrs │ Boston   │
		└────────────┴────────┴──────────┘
`
	if !visualCheck(t, "StreamBasic", buf.String(), expected) {
		fmt.Println(st.Debug())
		t.Fail()
	}

}
