// File: tests/caption_test.go
package tests

import (
	"bytes"
	"testing"

	"github.com/olekukonko/tablewriter/renderer"

	"github.com/olekukonko/tablewriter"
	"github.com/olekukonko/tablewriter/tw" // Assuming your tw types (CaptionPosition, Align) are here
)

// TestTableCaptions comprehensive tests for caption functionality
func TestTableCaptions(t *testing.T) {
	data := [][]string{
		{"Alice", "30", "New York"},
		{"Bob", "24", "San Francisco"},
		{"Charlie", "35", "London"},
	}
	headers := []string{"Name", "Age", "City"}
	shortCaption := "User Data"
	longCaption := "This is a detailed caption for the user data table, intended to demonstrate text wrapping and alignment features."

	baseTableSetup := func(buf *bytes.Buffer) *tablewriter.Table {
		table := tablewriter.NewTable(buf,
			tablewriter.WithRenderer(renderer.NewBlueprint(tw.Rendition{Symbols: tw.NewSymbols(tw.StyleASCII)})),
			tablewriter.WithDebug(true),
		)
		table.Header(headers)
		for _, v := range data {
			table.Append(v)
		}
		return table
	}

	t.Run("NoCaption", func(t *testing.T) {
		var buf bytes.Buffer
		table := baseTableSetup(&buf)
		table.Render()
		expected := `
+---------+-----+---------------+
|  NAME   | AGE |     CITY      |
+---------+-----+---------------+
| Alice   | 30  | New York      |
| Bob     | 24  | San Francisco |
| Charlie | 35  | London        |
+---------+-----+---------------+
`
		if !visualCheckCaption(t, t.Name(), buf.String(), expected) {
			t.Log(table.Debug().String())
		}
	})

	t.Run("LegacySetCaption_BottomCenter", func(t *testing.T) {
		var buf bytes.Buffer
		table := baseTableSetup(&buf)
		table.Caption(tw.Caption{Text: shortCaption}) // Legacy, defaults to BottomCenter, auto width
		table.Render()
		// Width of table: 7+3+15 + 4 borders/separators = 29. "User Data" is 9.
		// (29-9)/2 = 10 padding left.
		expected := `
+---------+-----+---------------+
|  NAME   | AGE |     CITY      |
+---------+-----+---------------+
| Alice   | 30  | New York      |
| Bob     | 24  | San Francisco |
| Charlie | 35  | London        |
+---------+-----+---------------+
			User Data    
`
		if !visualCheckCaption(t, t.Name(), buf.String(), expected) {
			t.Log(table.Debug().String())
		}
	})

	t.Run("CaptionBottomCenter_AutoWidthBottom", func(t *testing.T) {
		var buf bytes.Buffer
		table := baseTableSetup(&buf)
		table.Caption(tw.Caption{Text: shortCaption, Spot: tw.SpotBottomCenter, Align: tw.AlignCenter})
		table.Render()
		expected := `
+---------+-----+---------------+
|  NAME   | AGE |     CITY      |
+---------+-----+---------------+
| Alice   | 30  | New York      |
| Bob     | 24  | San Francisco |
| Charlie | 35  | London        |
+---------+-----+---------------+
          User Data
`
		if !visualCheckCaption(t, t.Name(), buf.String(), expected) {
			t.Log(table.Debug().String())
		}
	})

	t.Run("CaptionTopCenter_AutoWidthTop", func(t *testing.T) {
		var buf bytes.Buffer
		table := baseTableSetup(&buf)
		table.Caption(tw.Caption{Text: shortCaption, Spot: tw.SpotTopCenter, Align: tw.AlignCenter})
		table.Render()
		expected := `
			User Data            
+---------+-----+---------------+
|  NAME   | AGE |     CITY      |
+---------+-----+---------------+
| Alice   | 30  | New York      |
| Bob     | 24  | San Francisco |
| Charlie | 35  | London        |
+---------+-----+---------------+
`
		if !visualCheckCaption(t, t.Name(), buf.String(), expected) {
			t.Log(table.Debug().String())
		}
	})

	t.Run("CaptionBottomLeft_AutoWidth", func(t *testing.T) {
		var buf bytes.Buffer
		table := baseTableSetup(&buf)
		table.Caption(tw.Caption{Text: shortCaption, Spot: tw.SpotBottomLeft, Align: tw.AlignLeft})
		table.Render()
		expected := `
+---------+-----+---------------+
|  NAME   | AGE |     CITY      |
+---------+-----+---------------+
| Alice   | 30  | New York      |
| Bob     | 24  | San Francisco |
| Charlie | 35  | London        |
+---------+-----+---------------+
User Data  
`
		if !visualCheckCaption(t, t.Name(), buf.String(), expected) {
			t.Log(table.Debug().String())
		}
	})

	t.Run("CaptionTopRight_AutoWidth", func(t *testing.T) {
		var buf bytes.Buffer
		table := baseTableSetup(&buf)
		table.Caption(tw.Caption{Text: shortCaption, Spot: tw.SpotTopRight, Align: tw.AlignRight})
		table.Render()
		expected := `
                       User Data
+---------+-----+---------------+
|  NAME   | AGE |     CITY      |
+---------+-----+---------------+
| Alice   | 30  | New York      |
| Bob     | 24  | San Francisco |
| Charlie | 35  | London        |
+---------+-----+---------------+
`
		if !visualCheckCaption(t, t.Name(), buf.String(), expected) {
			t.Log(table.Debug().String())
		}
	})

	t.Run("CaptionBottomCenter_LongCaption_AutoWidth", func(t *testing.T) {
		var buf bytes.Buffer
		table := baseTableSetup(&buf)
		table.Caption(tw.Caption{Text: longCaption, Spot: tw.SpotBottomCenter, Align: tw.AlignCenter})
		table.Render()
		// Table width is 29. Long caption will wrap to this.
		// "This is a detailed caption for" (29)
		// "the user data table, intended" (29)
		// "to demonstrate text wrapping" (28)
		// "and alignment features." (25)
		expected := `
+---------+-----+---------------+
|  NAME   | AGE |     CITY      |
+---------+-----+---------------+
| Alice   | 30  | New York      |
| Bob     | 24  | San Francisco |
| Charlie | 35  | London        |
+---------+-----+---------------+
  This is a detailed caption for 
  the user data table, intended  
 to demonstrate text wrapping and
	   alignment features.  
`
		if !visualCheckCaption(t, t.Name(), buf.String(), expected) {
			t.Log(table.Debug().String())
		}
	})

	t.Run("CaptionTopLeft_LongCaption_MaxWidth20", func(t *testing.T) {
		var buf bytes.Buffer
		table := baseTableSetup(&buf)
		// captionMaxWidth 20, table width is 29. Caption wraps to 20. Padded to 29 (table width) for alignment.
		table.Caption(tw.Caption{Text: longCaption, Spot: tw.SpotTopLeft, Align: tw.AlignLeft, Width: 20})
		table.Render()

		// The visual check normalizes spaces, so the alignment padding to table width is tricky to test visually for left/right aligned captions.
		// It's more about the wrapping width for the caption text itself.
		// The printTopBottomCaption will align the *block* of wrapped text.
		// Let's adjust the expected output: the lines are padded to actualTableWidth.
		// The caption lines themselves are max 20 wide.
		expectedAdjusted := `
This is a detailed               
caption for the user             
data table, intended             
to demonstrate                   
text wrapping and                
alignment features.              
+---------+-----+---------------+
|  NAME   | AGE |     CITY      |
+---------+-----+---------------+
| Alice   | 30  | New York      |
| Bob     | 24  | San Francisco |
| Charlie | 35  | London        |
+---------+-----+---------------+
`
		if !visualCheckCaption(t, t.Name(), buf.String(), expectedAdjusted) {
			t.Log(table.Debug().String())
		}
	})

	t.Run("CaptionBottomCenter_EmptyTable", func(t *testing.T) {
		var buf bytes.Buffer
		table := tablewriter.NewWriter(&buf)
		// No header, no data
		table.Caption(tw.Caption{Text: shortCaption, Spot: tw.SpotBottomCenter, Align: tw.AlignCenter})
		table.Render()
		// Expected: table is empty box, caption centered to its own width or a default.
		// Empty table with default borders prints:
		// +--+
		// +--+
		// If actualTableWidth is 0, captionWrapWidth becomes natural width of caption (9 for "User Data")
		// Then paddingTargetWidth also becomes 9.
		expected := `
+--+
+--+
User Data
`
		if !visualCheckCaption(t, t.Name(), buf.String(), expected) {
			t.Log(table.Debug().String())
		}
	})

	t.Run("CaptionTopLeft_EmptyTable_MaxWidth10", func(t *testing.T) {
		var buf bytes.Buffer
		table := tablewriter.NewWriter(&buf)
		table.Caption(tw.Caption{Text: "A very long caption text.", Spot: tw.SpotTopLeft, Align: tw.AlignLeft, Width: 10})
		table.Render()
		// Table is empty, captionMaxWidth is 10.
		// "A very"
		// "long"
		// "caption"
		// "text."
		// Each line left-aligned within width 10.
		expected := `
A very
long
caption
text.
+--+
+--+
`
		if !visualCheckCaption(t, t.Name(), buf.String(), expected) {
			t.Log(table.Debug().String())
		}
	})
}
