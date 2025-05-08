// tests/streamer_test.go
package tests

import (
	"bytes"
	"fmt"
	"io"
	"testing"
	"time" // Added for slow test

	"github.com/olekukonko/tablewriter"
	"github.com/olekukonko/tablewriter/streamer"
	"github.com/olekukonko/tablewriter/tw"
)

// Helper function (remains the same)
func createOceanStreamTable(t *testing.T, w io.Writer, cfg streamer.OceanConfig, debug bool) (*tablewriter.TableStream, *streamer.Ocean) {
	t.Helper()
	oceanRenderer, err := streamer.NewOcean(w, debug, cfg)
	if err != nil {
		t.Fatalf("Failed to create Ocean renderer: %v", err)
	}
	st, err := tablewriter.NewStreamTable(w, oceanRenderer)
	if err != nil {
		t.Fatalf("Failed to create TableStream: %v", err)
	}
	return st, oceanRenderer
}

// TestOceanStreamBasic (Passed before, should still pass)
func TestOceanStreamBasic(t *testing.T) {
	var buf bytes.Buffer
	// Using ColumnWidths consistent with the previous expected output
	cfg := streamer.OceanConfig{
		ColumnWidths:   []int{9, 7, 12}, // Adjusted widths for default padding " "
		ShowHeaderLine: true,
	}
	st, rendererInstance := createOceanStreamTable(t, &buf, cfg, false)

	st.Start()
	st.Header([]string{"Name", "Age", "City"})  // Content: 4, 3, 4. Avail: 7, 5, 10. Fits.
	st.Row([]string{"Alice", "25", "New York"}) // Content: 5, 2, 8. Avail: 7, 5, 10. Fits.
	st.Row([]string{"Bob", "30", "Boston"})     // Content: 3, 2, 6. Avail: 7, 5, 10. Fits.
	st.End()

	expected := `
        ┌─────────┬───────┬────────────┐
        │  Name   │  Age  │    City    │
        ├─────────┼───────┼────────────┤
        │ Alice   │ 25    │ New York   │
        │ Bob     │ 30    │ Boston     │
        └─────────┴───────┴────────────┘
`
	if !visualCheck(t, "OceanStreamBasic", buf.String(), expected) {
		fmt.Println("--- DEBUG LOG ---")
		for _, msg := range rendererInstance.Debug() {
			fmt.Println(msg)
		}
		t.Fail()
	}
}

// TestOceanStreamNoHeaderASCII (Fixing widths and expected)
func TestOceanStreamNoHeaderASCII(t *testing.T) {
	var buf bytes.Buffer
	// Let's use widths that allow the content + default padding " "
	// "Regular" (7) needs 1+7+1=9
	// "line" (4) needs 1+4+1=6
	// "1" (1) needs 1+1+1=3
	// "Thick" (5) needs 1+5+1=7 (use 9)
	// "thick" (5) needs 1+5+1=7 (use 7)
	// "2" (1) needs 1+1+1=3 (use 3)
	// Choose max: [9, 7, 3]
	cfg := streamer.OceanConfig{
		ColumnWidths: []int{9, 7, 3},
		Symbols:      tw.NewSymbols(tw.StyleASCII),
		Borders:      tw.Border{Left: tw.On, Right: tw.On, Top: tw.On, Bottom: tw.On},
		RowAlign:     tw.AlignLeft, // Explicit for clarity
	}
	st, _ := createOceanStreamTable(t, &buf, cfg, false)

	st.Start()
	st.Row([]string{"Regular", "line", "1"}) // Fits: 7<=7, 4<=5, 1<=1
	st.Row([]string{"Thick", "thick", "2"})  // Fits: 5<=7, 5<=5, 1<=1
	st.End()

	// Expected with widths [9, 7, 3] and left align
	expected := `
+---------+-------+---+
| Regular | line  | 1 |
| Thick   | thick | 2 |
+---------+-------+---+
`
	visualCheck(t, "OceanStreamNoHeaderASCII", buf.String(), expected)
}

// TestOceanStreamWithFooterAlign (Fixing widths and expected)
func TestOceanStreamWithFooterAlign(t *testing.T) {
	var buf bytes.Buffer
	// Header: ITEM(4) QTY(3) PRICE(5). Center Align. Widths needed: 6, 5, 7
	// Row 1: Item 1(6) 2(1) 1000.00(7). Align L,C,R. Widths needed: 8, 3, 9
	// Row 2: Item 2(6) 10(2) 25.50(5). Align L,C,R. Widths needed: 8, 4, 7
	// Footer: ""(-) Total(5) 1025.50(7). Align L,C,R. Widths needed: 2, 7, 9
	// Max widths needed: [8, 7, 9]
	cfg := streamer.OceanConfig{
		ColumnWidths:   []int{8, 7, 9},
		ColumnAligns:   []tw.Align{tw.AlignLeft, tw.AlignCenter, tw.AlignRight}, // Specific column aligns
		HeaderAlign:    tw.AlignCenter,                                          // Default for Header
		FooterAlign:    tw.AlignRight,                                           // Default for Footer (but ColumnAligns overrides)
		ShowHeaderLine: true,
		ShowFooterLine: true,
	}
	st, rendererInstance := createOceanStreamTable(t, &buf, cfg, false)

	st.Start()
	st.Header([]string{"Item", "Qty", "Price"})
	st.Row([]string{"Item 1", "2", "1000.00"})
	st.Row([]string{"Item 2", "10", "25.50"})
	st.Footer([]string{"", "Total", "1025.50"})
	st.End()

	// Expected with widths [8, 7, 9] and alignments L, C, R
	expected := `
        ┌────────┬───────┬─────────┐
        │ Item   │  Qty  │   Price │
        ├────────┼───────┼─────────┤
        │ Item 1 │   2   │ 1000.00 │
        │ Item 2 │  10   │   25.50 │
        ├────────┼───────┼─────────┤
        │        │ Total │ 1025.50 │
        └────────┴───────┴─────────┘
`
	if !visualCheck(t, "OceanStreamWithFooterAlign", buf.String(), expected) {
		fmt.Println("--- DEBUG LOG ---")
		for _, msg := range rendererInstance.Debug() {
			fmt.Println(msg)
		}
		t.Fail()
	}
}

// TestOceanStreamTruncation (Fixing widths and expected)
func TestOceanStreamTruncation(t *testing.T) {
	var buf bytes.Buffer
	// Choose widths: ID(4), Desc(15), Status(6) -> Default padding " "
	// Avail content: 2, 13, 4
	cfg := streamer.OceanConfig{
		ColumnWidths:   []int{4, 15, 8}, // Increased status width slightly
		ShowHeaderLine: true,
		HeaderAlign:    tw.AlignCenter,
		RowAlign:       tw.AlignLeft,
	}
	st, rendererInstance := createOceanStreamTable(t, &buf, cfg, false)

	st.Start()
	st.Header([]string{"ID", "Description", "Status"})                  // Fits: 2<=2, 11<=13, 6<=6
	st.Row([]string{"1", "This description is quite long", "OK"})       // Truncate: 1<=2, 30 > 13 -> "This descript…"(13), 2<=6
	st.Row([]string{"2", "Short desc", "DONE"})                         // Fits: 1<=2, 10<=13, 4<=6
	st.Row([]string{"3", "Another long one needing truncation", "ERR"}) // Truncate: 1<=2, 35 > 13 -> "Another long …"(13), 3<=6
	st.End()

	// Expected with widths [4, 15, 8]
	expected := `
        ┌────┬───────────────┬────────┐
        │ ID │  Description  │ Status │
        ├────┼───────────────┼────────┤
        │ 1  │ This descrip… │ OK     │
        │ 2  │ Short desc    │ DONE   │
        │ 3  │ Another long… │ ERR    │
        └────┴───────────────┴────────┘
`
	if !visualCheck(t, "OceanStreamTruncation", buf.String(), expected) {
		fmt.Println("--- DEBUG LOG ---")
		for _, msg := range rendererInstance.Debug() {
			fmt.Println(msg)
		}
		t.Fail()
	}
}

// TestOceanStreamBorders (Fixing widths and expected)
func TestOceanStreamBorders(t *testing.T) {
	data := [][]string{{"A", "B"}, {"C", "D"}}
	// Use width 3 => content avail 1. 'A', 'B', 'C', 'D' fit.
	widths := []int{3, 3}

	tests := []struct {
		name     string
		borders  tw.Border
		expected string
	}{
		{
			"All Off",
			tw.Border{Left: tw.Off, Right: tw.Off, Top: tw.Off, Bottom: tw.Off},
			` A │ B 
 C │ D 
`, // Space for default padding around A, B, C, D
		},
		{
			"No Left/Right",
			tw.Border{Left: tw.Off, Right: tw.Off, Top: tw.On, Bottom: tw.On},
			`───┬───
 A │ B 
 C │ D 
───┴───
`,
		},
		{
			"No Top/Bottom",
			tw.Border{Left: tw.On, Right: tw.On, Top: tw.Off, Bottom: tw.Off},
			`│ A │ B │
│ C │ D │
`,
		},
		{
			"Only Left",
			tw.Border{Left: tw.On, Right: tw.Off, Top: tw.Off, Bottom: tw.Off},
			`│ A │ B 
│ C │ D 
`,
		},
		{
			"Default (All On in Ocean's default)",
			tw.Border{},
			`┌───┬───┐
│ A │ B │
│ C │ D │
└───┴───┘
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			cfg := streamer.OceanConfig{
				ColumnWidths: widths,
				RowAlign:     tw.AlignLeft,
			}
			if tt.name != "Default (All On in Ocean's default)" {
				cfg.Borders = tt.borders
			}
			st, rendererInstance := createOceanStreamTable(t, &buf, cfg, false)

			st.Start()
			st.Row(data[0])
			st.Row(data[1])
			st.End()

			if !visualCheck(t, "OceanStreamBorders_"+tt.name, buf.String(), tt.expected) {
				fmt.Printf("--- DEBUG LOG for %s ---\n", tt.name)
				for _, msg := range rendererInstance.Debug() {
					fmt.Println(msg)
				}
				t.Fail()
			}
		})
	}
}

// TestOceanStreamCustomPadding (Fixing widths and expected)
func TestOceanStreamCustomPadding(t *testing.T) {
	var buf bytes.Buffer
	// Padding L=">>"(2), R="<<"(2). Total pad=4.
	// Let's target width 7 for the cell. Avail content = 7-2-2 = 3.
	// Header "Head1"(5) > 3. Truncate. avail(3) >= suffix(1). Truncate to 3-1=2. "He"+... = "He…" (width 3).
	// Row "R1C1"(4) > 3. Truncate. avail(3) >= suffix(1). Truncate to 3-1=2. "R1"+... = "R1…" (width 3).
	cfg := streamer.OceanConfig{
		ColumnWidths:   []int{7, 7},
		Padding:        tw.Padding{Left: ">>", Right: "<<"},
		ShowHeaderLine: true,
		HeaderAlign:    tw.AlignCenter,
		RowAlign:       tw.AlignLeft,
	}
	st, rendererInstance := createOceanStreamTable(t, &buf, cfg, false)

	st.Start()
	st.Header([]string{"Head1", "Head2"})
	st.Row([]string{"R1C1", "R1C2"})
	st.End()

	// Expected cell content: "He…" and "R1…"
	// Final cell string: ">>" + "He…" + "<<" = ">>He…<<" (width 2+3+2=7)
	// Final cell string: ">>" + "R1…" + "<<" = ">>R1…<<" (width 2+3+2=7)
	expected := `
        ┌───────┬───────┐
        │>>He…<<│>>He…<<│
        ├───────┼───────┤
        │>>R1…<<│>>R1…<<│
        └───────┴───────┘
`
	if !visualCheck(t, "OceanStreamCustomPadding", buf.String(), expected) {
		fmt.Println("--- DEBUG LOG ---")
		for _, msg := range rendererInstance.Debug() { // Use rendererInstance
			fmt.Println(msg)
		}
		t.Fail()
	}
}

// TestOceanStreamErrors (Should pass as is, no expected output changes needed)
func TestOceanStreamErrors(t *testing.T) {
	var buf bytes.Buffer

	t.Run("NewOcean No ColumnWidths", func(t *testing.T) {
		_, err := streamer.NewOcean(&buf, false, streamer.OceanConfig{})
		if err == nil {
			t.Error("Expected error for missing ColumnWidths, got nil")
		} else {
			t.Logf("Got expected error: %v", err)
		}
	})

	t.Run("NewOcean Invalid ColumnWidth (zero)", func(t *testing.T) {
		_, err := streamer.NewOcean(&buf, false, streamer.OceanConfig{ColumnWidths: []int{5, 0}})
		if err == nil {
			t.Error("Expected error for zero ColumnWidth, got nil")
		} else {
			// The message should reflect the padding check now
			t.Logf("Got expected error: %v", err)
		}
	})

	t.Run("NewOcean Invalid ColumnWidth (too small for padding)", func(t *testing.T) {
		_, err := streamer.NewOcean(&buf, false, streamer.OceanConfig{
			ColumnWidths: []int{5, 1},
			Padding:      tw.Padding{Left: " ", Right: " "},
		})
		if err == nil {
			t.Error("Expected error for too small ColumnWidth with padding, got nil")
		} else {
			t.Logf("Got expected error: %v", err)
		}
	})

	t.Run("NewOcean Mismatched ColumnAligns", func(t *testing.T) {
		_, err := streamer.NewOcean(&buf, false, streamer.OceanConfig{
			ColumnWidths: []int{5, 5},
			ColumnAligns: []tw.Align{tw.AlignLeft},
		})
		if err == nil {
			t.Error("Expected error for mismatched ColumnAligns length, got nil")
		} else {
			t.Logf("Got expected error: %v", err)
		}
	})

	t.Run("NewStreamTable Nil Renderer", func(t *testing.T) {
		_, err := tablewriter.NewStreamTable(&buf, nil)
		if err == nil {
			t.Error("Expected error for nil renderer, got nil")
		} else {
			t.Logf("Got expected error: %v", err)
		}
	})

	t.Run("NewStreamTable Nil Writer", func(t *testing.T) {
		dummyOceanWriter := &bytes.Buffer{}
		oceanCfg := streamer.OceanConfig{ColumnWidths: []int{10}}
		r, _ := streamer.NewOcean(dummyOceanWriter, false, oceanCfg)
		_, err := tablewriter.NewStreamTable(nil, r)
		if err == nil {
			t.Error("Expected error for nil writer, got nil")
		} else {
			t.Logf("Got expected error: %v", err)
		}
	})

	t.Run("TableStream Call Before Start", func(t *testing.T) {
		var streamBuf bytes.Buffer
		oceanCfg := streamer.OceanConfig{ColumnWidths: []int{10}}
		oceanRenderer, _ := streamer.NewOcean(&streamBuf, false, oceanCfg)
		st, _ := tablewriter.NewStreamTable(&streamBuf, oceanRenderer)

		err := st.Header([]string{"H"})
		if err == nil {
			t.Error("Expected error calling Header before Start, got nil")
		} else {
			t.Logf("Got expected Header error: %v", err)
		}
		err = st.Row([]string{"R"})
		if err == nil {
			t.Error("Expected error calling Row before Start, got nil")
		} else {
			t.Logf("Got expected Row error: %v", err)
		}
		err = st.Footer([]string{"F"})
		if err == nil {
			t.Error("Expected error calling Footer before Start, got nil")
		} else {
			t.Logf("Got expected Footer error: %v", err)
		}
		err = st.End()
		if err == nil {
			t.Error("Expected error calling End before Start, got nil")
		} else {
			t.Logf("Got expected End error: %v", err)
		}
	})
}

// TestOceanStreamEmptyCells (Fixing widths and expected)
func TestOceanStreamEmptyCells(t *testing.T) {
	var buf bytes.Buffer
	// Choose widths [5, 6, 5]. Avail content [3, 4, 3]
	cfg := streamer.OceanConfig{
		ColumnWidths:   []int{5, 6, 5},
		ShowHeaderLine: true,
		HeaderAlign:    tw.AlignCenter,
		RowAlign:       tw.AlignLeft,
	}
	st, rendererInstance := createOceanStreamTable(t, &buf, cfg, false)

	st.Start()
	st.Header([]string{"H1", "", "H3"})  // Fits: 2<=3, 0<=4, 2<=3
	st.Row([]string{"", "R1C2", ""})     // Fits: 0<=3, 4<=4, 0<=3
	st.Row([]string{"R2C1", "", "R2C3"}) // Fits: 4>3 ->"R2…", 0<=4, 4>3 ->"R2…"
	st.Row([]string{"", "", ""})         // Fits: 0<=3, 0<=4, 0<=3
	st.End()

	// Expected with widths [5, 6, 5]
	expected := `
┌─────┬──────┬─────┐
│ H1  │      │ H3  │
├─────┼──────┼─────┤
│     │ R1C2 │     │
│ R2… │      │ R2… │
│     │      │     │
└─────┴──────┴─────┘
`
	if !visualCheck(t, "OceanStreamEmptyCells", buf.String(), expected) {
		fmt.Println("--- DEBUG LOG ---")
		for _, msg := range rendererInstance.Debug() {
			fmt.Println(msg)
		}
		t.Fail()
	}
}

// TestOceanStreamOnlyHeader (Fixing widths and expected)
func TestOceanStreamOnlyHeader(t *testing.T) {
	var buf bytes.Buffer
	// Widths [9, 9]. Avail content [7, 7]
	cfg := streamer.OceanConfig{
		ColumnWidths:   []int{9, 9},
		ShowHeaderLine: true,
		HeaderAlign:    tw.AlignCenter,
	}
	st, _ := createOceanStreamTable(t, &buf, cfg, false)

	st.Start()
	st.Header([]string{"Header1", "Header2"}) // Fits: 7<=7
	st.End()

	// Expected with widths [9, 9]
	expected := `
┌─────────┬─────────┐
│ Header1 │ Header2 │
├─────────┼─────────┤
└─────────┴─────────┘
`
	visualCheck(t, "OceanStreamOnlyHeader", buf.String(), expected)
}

// TestOceanStreamOnlyHeaderNoHeaderLine (Fixing widths and expected)
func TestOceanStreamOnlyHeaderNoHeaderLine(t *testing.T) {
	var buf bytes.Buffer
	// Widths [9, 9]. Avail content [7, 7]
	cfg := streamer.OceanConfig{
		ColumnWidths:   []int{9, 9},
		ShowHeaderLine: false,
		HeaderAlign:    tw.AlignCenter,
	}
	st, _ := createOceanStreamTable(t, &buf, cfg, false)

	st.Start()
	st.Header([]string{"Header1", "Header2"}) // Fits: 7<=7
	st.End()

	// Expected with widths [9, 9]
	expected := `
┌─────────┬─────────┐
│ Header1 │ Header2 │
└─────────┴─────────┘
`
	visualCheck(t, "OceanStreamOnlyHeaderNoHeaderLine", buf.String(), expected)
}

// TestOceanStreamSlowOutput (No expected output changes needed, just observation)
func TestOceanStreamSlowOutput(t *testing.T) {
	var buf bytes.Buffer
	// Using widths that should fit the content comfortably
	cfg := streamer.OceanConfig{
		ColumnWidths:   []int{15, 15},
		ShowHeaderLine: true,
		HeaderAlign:    tw.AlignCenter,
		RowAlign:       tw.AlignLeft,
	}
	st, rendererInstance := createOceanStreamTable(t, &buf, cfg, false)

	fmt.Println("Starting slow stream test...") // User prompt

	err := st.Start()
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	fmt.Println("Output after Start():")
	fmt.Print(buf.String()) // Print incremental output
	buf.Reset()
	time.Sleep(1 * time.Second)

	err = st.Header([]string{"Event", "Timestamp"})
	if err != nil {
		t.Fatalf("Header failed: %v", err)
	}
	fmt.Println("\nOutput after Header():")
	fmt.Print(buf.String())
	buf.Reset()
	time.Sleep(1 * time.Second)

	for i := 1; i <= 3; i++ {
		err = st.Row([]string{fmt.Sprintf("Data Row %d", i), time.Now().Format("15:04:05.000")})
		if err != nil {
			t.Fatalf("Row %d failed: %v", i, err)
		}
		fmt.Printf("\nOutput after Row %d:\n", i)
		fmt.Print(buf.String())
		buf.Reset()
		time.Sleep(1 * time.Second)
	}

	err = st.End()
	if err != nil {
		t.Fatalf("End failed: %v", err)
	}
	fmt.Println("\nOutput after End():")
	fmt.Print(buf.String()) // Print final part

	fmt.Println("\n--- Full Output (Conceptual) ---") // Explain full output not captured incrementally

	// Conceptual check if needed - build the expected structure manually
	// var finalBuf bytes.Buffer
	// stFinal, _ := createOceanStreamTable(t, &finalBuf, cfg, false)
	// stFinal.Start()
	// stFinal.Header([]string{"Event", "Timestamp"})
	// stFinal.Row([]string{"Data Row 1", "HH:MM:SS.mmm"})
	// stFinal.Row([]string{"Data Row 2", "HH:MM:SS.mmm"})
	// stFinal.Row([]string{"Data Row 3", "HH:MM:SS.mmm"})
	// stFinal.End()
	// expectedFull := `...`
	// visualCheck(t, "OceanStreamSlowOutput Full", finalBuf.String(), expectedFull) // Would fail on timestamps

	t.Log("Slow stream test completed. Observe terminal output.")
	if len(rendererInstance.Debug()) > 0 {
		fmt.Println("--- DEBUG LOG ---")
		for _, msg := range rendererInstance.Debug() {
			fmt.Println(msg)
		}
	}
}
