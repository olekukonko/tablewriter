package tests

import (
	"bytes"
	"testing"

	"github.com/olekukonko/tablewriter"
	"github.com/olekukonko/tablewriter/renderer"
	"github.com/olekukonko/tablewriter/tw"
)

// TestHTMLBasicTable verifies that a basic HTML table with headers and rows is rendered correctly.
func TestHTMLBasicTable(t *testing.T) {
	var buf bytes.Buffer
	table := tablewriter.NewTable(&buf,
		tablewriter.WithRenderer(renderer.NewHTML(&buf, false)),
	)
	table.SetHeader([]string{"Name", "Age", "City"})
	table.Append([]string{"Alice", "25", "New York"})
	table.Append([]string{"Bob", "30", "Boston"})
	table.Render()

	expected := `
<table>
<thead>
<tr><th style="text-align: center;">NAME</th><th style="text-align: center;">AGE</th><th style="text-align: center;">CITY</th></tr>
</thead>
<tbody>
<tr><td style="text-align: left;">Alice</td><td style="text-align: left;">25</td><td style="text-align: left;">New York</td></tr>
<tr><td style="text-align: left;">Bob</td><td style="text-align: left;">30</td><td style="text-align: left;">Boston</td></tr>
</tbody>
</table>`
	if !visualCheckHTML(t, "HTMLBasicTable", buf.String(), expected) {
		for _, v := range table.Debug() {
			t.Error(v)
		}
	}
}

// TestHTMLWithFooterAndAlignment tests an HTML table with a footer and custom column alignments.
func TestHTMLWithFooterAndAlignment(t *testing.T) {
	var buf bytes.Buffer
	table := tablewriter.NewTable(&buf,
		tablewriter.WithRenderer(renderer.NewHTML(&buf, false)),
		tablewriter.WithHeaderConfig(tablewriter.CellConfig{
			Formatting: tablewriter.CellFormatting{Alignment: tw.AlignCenter},
		}),
		tablewriter.WithRowConfig(tablewriter.CellConfig{
			ColumnAligns: []tw.Align{tw.AlignLeft, tw.AlignRight, tw.AlignCenter},
		}),
		tablewriter.WithFooterConfig(tablewriter.CellConfig{
			Formatting: tablewriter.CellFormatting{Alignment: tw.AlignRight},
		}),
	)
	table.SetHeader([]string{"Item", "Qty", "Price"})
	table.Append([]string{"Apple", "5", "1.20"})
	table.Append([]string{"Banana", "12", "0.35"})
	table.SetFooter([]string{"", "Total", "7.20"})
	table.Render()

	expected := `
<table>
<thead>
<tr><th style="text-align: center;">Item</th><th style="text-align: center;">Qty</th><th style="text-align: center;">Price</th></tr>
</thead>
<tbody>
<tr><td style="text-align: left;">Apple</td><td style="text-align: right;">5</td><td style="text-align: center;">1.20</td></tr>
<tr><td style="text-align: left;">Banana</td><td style="text-align: right;">12</td><td style="text-align: center;">0.35</td></tr>
</tbody>
<tfoot>
<tr><td style="text-align: right;"></td><td style="text-align: right;">Total</td><td style="text-align: right;">7.20</td></tr>
</tfoot>
</table>`
	if !visualCheckHTML(t, "HTMLWithFooterAndAlignment", buf.String(), expected) {
		for _, v := range table.Debug() {
			t.Error(v)
		}
	}
}

// TestHTMLEscaping verifies HTML content escaping behavior with and without EscapeContent enabled.
func TestHTMLEscaping(t *testing.T) {
	// Test case 1: Default (EscapeContent = true)
	var buf bytes.Buffer
	table := tablewriter.NewTable(&buf,
		tablewriter.WithRenderer(renderer.NewHTML(&buf, false)),
	)
	table.SetHeader([]string{"Tag", "Attribute"})
	table.Append([]string{"<br>", "should escape < & >"})
	table.Render()

	expectedEscaped := `
<table>
<thead>
<tr><th style="text-align: center;">TAG</th><th style="text-align: center;">ATTRIBUTE</th></tr>
</thead>
<tbody>
<tr><td style="text-align: left;">&lt;br&gt;</td><td style="text-align: left;">should escape &lt; &amp; &gt;</td></tr>
</tbody>
</table>
`
	if !visualCheckHTML(t, "HTMLEscaping_Default", buf.String(), expectedEscaped) {
		t.Log("--- Debug Log for HTMLEscaping_Default ---")
		for _, v := range table.Debug() {
			t.Log(v)
		}
	}

	// Test case 2: EscapeContent = false
	buf.Reset()
	tableNoEscape := tablewriter.NewTable(&buf,
		tablewriter.WithRenderer(renderer.NewHTML(&buf, false, renderer.HTMLConfig{EscapeContent: false})),
	)
	tableNoEscape.SetHeader([]string{"Tag", "Attribute"})
	tableNoEscape.Append([]string{"<br>", "should NOT escape < & >"})
	tableNoEscape.Render()

	expectedUnescaped := `
<table>
<thead>
<tr><th style="text-align: center;">TAG</th><th style="text-align: center;">ATTRIBUTE</th></tr>
</thead>
<tbody>
<tr><td style="text-align: left;"><br></td><td style="text-align: left;">should NOT escape < & ></td></tr>
</tbody>
</table>`
	if !visualCheckHTML(t, "HTMLEscaping_Disabled", buf.String(), expectedUnescaped) {
		t.Log("--- Debug Log for HTMLEscaping_Disabled ---")
		for _, v := range tableNoEscape.Debug() {
			t.Log(v)
		}
	}
}

// TestHTMLMultiLine tests HTML table rendering with multiline cell content, noting that newlines create separate rows.
func TestHTMLMultiLine(t *testing.T) {
	// Test case 1: Default behavior (newlines split into separate rows)
	var buf bytes.Buffer
	table := tablewriter.NewTable(&buf,
		tablewriter.WithRenderer(renderer.NewHTML(&buf, false)),
	)
	table.Append([]string{"Line 1\nLine 2", "Single Line"})
	table.Render()

	expected := `
<table>
<tbody>
<tr><td style="text-align: left;">Line 1</td><td style="text-align: left;">Single Line</td></tr>
<tr><td style="text-align: left;">Line 2</td><td style="text-align: left;"></td></tr>
</tbody>
</table>`
	if !visualCheckHTML(t, "HTMLMultiLine_Default", buf.String(), expected) {
		t.Logf("MultiLine Default Output: %s", buf.String())
		for _, v := range table.Debug() {
			t.Error(v)
		}
	}

	// Test case 2: With AddLinesTag (no effect due to newline pre-splitting)
	buf.Reset()
	tableLinesTag := tablewriter.NewTable(&buf,
		tablewriter.WithRenderer(renderer.NewHTML(&buf, false, renderer.HTMLConfig{AddLinesTag: true})),
	)
	tableLinesTag.Append([]string{"Line 1\nLine 2", "Single Line"})
	tableLinesTag.Render()

	expectedLinesTag := `
<table>
<tbody>
<tr><td style="text-align: left;">Line 1</td><td style="text-align: left;">Single Line</td></tr>
<tr><td style="text-align: left;">Line 2</td><td style="text-align: left;"></td></tr>
</tbody>
</table>`
	if !visualCheckHTML(t, "HTMLMultiLine_LinesTag", buf.String(), expectedLinesTag) {
		t.Logf("MultiLine LinesTag Output: %s", buf.String())
		for _, v := range tableLinesTag.Debug() {
			t.Error(v)
		}
	}
}

// TestHTMLHorizontalMerge verifies HTML table rendering with horizontal cell merges in headers, rows, and footers.
func TestHTMLHorizontalMerge(t *testing.T) {
	var buf bytes.Buffer
	table := tablewriter.NewTable(&buf,
		tablewriter.WithRenderer(renderer.NewHTML(&buf, false)),
		tablewriter.WithConfig(tablewriter.Config{
			Header: tablewriter.CellConfig{Formatting: tablewriter.CellFormatting{MergeMode: tw.MergeHorizontal}},
			Row:    tablewriter.CellConfig{Formatting: tablewriter.CellFormatting{MergeMode: tw.MergeHorizontal}},
			Footer: tablewriter.CellConfig{Formatting: tablewriter.CellFormatting{MergeMode: tw.MergeHorizontal}},
		}),
	)
	table.SetHeader([]string{"A", "Merged Header", "Merged Header"})
	table.Append([]string{"Data 1", "Data 2", "Data 2"})
	table.Append([]string{"Merged Row", "Merged Row", "Data 3"})
	table.SetFooter([]string{"Footer 1", "Merged Footer", "Merged Footer"})
	table.Render()

	expected := `
<table>
<thead>
<tr><th style="text-align: center;">A</th><th colspan="2" style="text-align: center;">MERGED HEADER</th></tr>
</thead>
<tbody>
<tr><td style="text-align: left;">Data 1</td><td colspan="2" style="text-align: left;">Data 2</td></tr>
<tr><td colspan="2" style="text-align: left;">Merged Row</td><td style="text-align: left;">Data 3</td></tr>
</tbody>
<tfoot>
<tr><td style="text-align: right;">Footer 1</td><td colspan="2" style="text-align: right;">Merged Footer</td></tr>
</tfoot>
</table>`
	if !visualCheckHTML(t, "HTMLHorizontalMerge", buf.String(), expected) {
		for _, v := range table.Debug() {
			t.Error(v)
		}
	}
}

// TestHTMLVerticalMerge tests HTML table rendering with vertical cell merges based on repeated values.
func TestHTMLVerticalMerge(t *testing.T) {
	var buf bytes.Buffer
	table := tablewriter.NewTable(&buf,
		tablewriter.WithRenderer(renderer.NewHTML(&buf, false)),
		tablewriter.WithConfig(tablewriter.Config{
			Row: tablewriter.CellConfig{Formatting: tablewriter.CellFormatting{MergeMode: tw.MergeVertical}},
		}),
	)
	table.SetHeader([]string{"Category", "Item", "Value"})
	table.Append([]string{"Fruit", "Apple", "10"})
	table.Append([]string{"Fruit", "Banana", "5"})
	table.Append([]string{"Fruit", "Orange", "8"})
	table.Append([]string{"Dairy", "Milk", "2"})
	table.Append([]string{"Dairy", "Cheese", "4"})
	table.Append([]string{"Other", "Bread", "3"})
	table.Render()

	expected := `
<table>
<thead>
<tr><th style="text-align: center;">CATEGORY</th><th style="text-align: center;">ITEM</th><th style="text-align: center;">VALUE</th></tr>
</thead>
<tbody>
<tr><td rowspan="3" style="text-align: left;">Fruit</td><td style="text-align: left;">Apple</td><td style="text-align: left;">10</td></tr>
<tr><td style="text-align: left;">Banana</td><td style="text-align: left;">5</td></tr>
<tr><td style="text-align: left;">Orange</td><td style="text-align: left;">8</td></tr>
<tr><td rowspan="2" style="text-align: left;">Dairy</td><td style="text-align: left;">Milk</td><td style="text-align: left;">2</td></tr>
<tr><td style="text-align: left;">Cheese</td><td style="text-align: left;">4</td></tr>
<tr><td style="text-align: left;">Other</td><td style="text-align: left;">Bread</td><td style="text-align: left;">3</td></tr>
</tbody>
</table>`
	if !visualCheckHTML(t, "HTMLVerticalMerge", buf.String(), expected) {
		for _, v := range table.Debug() {
			t.Error(v)
		}
	}
}

// TestHTMLCombinedMerge verifies HTML table rendering with both horizontal and vertical cell merges.
func TestHTMLCombinedMerge(t *testing.T) {
	var buf bytes.Buffer
	table := tablewriter.NewTable(&buf,
		tablewriter.WithRenderer(renderer.NewHTML(&buf, false)),
		tablewriter.WithConfig(tablewriter.Config{
			Row: tablewriter.CellConfig{Formatting: tablewriter.CellFormatting{MergeMode: tw.MergeBoth}},
		}),
	)
	table.SetHeader([]string{"Region", "Quarter", "Sales", "Target"})
	table.Append([]string{"North", "Q1", "1000", "900"})
	table.Append([]string{"North", "Q2", "1200", "1100"})
	table.Append([]string{"South", "Q1+Q2", "Q1+Q2", "2000"})
	table.Append([]string{"East", "Q1", "800", "850"})
	table.Append([]string{"East", "Q2", "950", "850"})
	table.Render()

	expected := `
<table>
<thead>
<tr><th style="text-align: center;">REGION</th><th style="text-align: center;">QUARTER</th><th style="text-align: center;">SALES</th><th style="text-align: center;">TARGET</th></tr>
</thead>
<tbody>
<tr><td rowspan="2" style="text-align: left;">North</td><td style="text-align: left;">Q1</td><td style="text-align: left;">1000</td><td style="text-align: left;">900</td></tr>
<tr><td style="text-align: left;">Q2</td><td style="text-align: left;">1200</td><td style="text-align: left;">1100</td></tr>
<tr><td style="text-align: left;">South</td><td colspan="2" style="text-align: left;">Q1+Q2</td><td style="text-align: left;">2000</td></tr>
<tr><td rowspan="2" style="text-align: left;">East</td><td style="text-align: left;">Q1</td><td style="text-align: left;">800</td><td rowspan="2" style="text-align: left;">850</td></tr>
<tr><td style="text-align: left;">Q2</td><td style="text-align: left;">950</td></tr>
</tbody>
</table>`
	if !visualCheckHTML(t, "HTMLCombinedMerge", buf.String(), expected) {
		t.Logf("Combined Merge Output: %s", buf.String())
		for _, v := range table.Debug() {
			t.Error(v)
		}
	}
}

// TestHTMLHierarchicalMerge tests HTML table rendering with hierarchical cell merges.
func TestHTMLHierarchicalMerge(t *testing.T) {
	var buf bytes.Buffer
	table := tablewriter.NewTable(&buf,
		tablewriter.WithRenderer(renderer.NewHTML(&buf, false)),
		tablewriter.WithConfig(tablewriter.Config{
			Row: tablewriter.CellConfig{Formatting: tablewriter.CellFormatting{MergeMode: tw.MergeHierarchical}},
		}),
		tablewriter.WithHeaderConfig(tablewriter.CellConfig{
			Formatting: tablewriter.CellFormatting{
				AutoFormat: false,
				Alignment:  tw.AlignCenter,
			},
		}),
	)
	table.SetHeader([]string{"L1", "L2", "L3"})
	table.Append([]string{"A", "a", "1"})
	table.Append([]string{"A", "b", "2"})
	table.Append([]string{"A", "b", "3"})
	table.Append([]string{"B", "c", "4"})
	table.Render()

	expected := `
<table>
<thead>
<tr><th style="text-align: center;">L1</th><th style="text-align: center;">L2</th><th style="text-align: center;">L3</th></tr>
</thead>
<tbody>
<tr><td rowspan="3" style="text-align: left;">A</td><td style="text-align: left;">a</td><td style="text-align: left;">1</td></tr>
<tr><td rowspan="2" style="text-align: left;">b</td><td style="text-align: left;">2</td></tr>
<tr><td style="text-align: left;">3</td></tr>
<tr><td style="text-align: left;">B</td><td style="text-align: left;">c</td><td style="text-align: left;">4</td></tr>
</tbody>
</table>`
	if !visualCheckHTML(t, "HTMLHierarchicalMerge", buf.String(), expected) {
		for _, v := range table.Debug() {
			t.Error(v)
		}
	}
}

// TestHTMLEmptyTable verifies HTML rendering for empty tables and tables with only headers.
func TestHTMLEmptyTable(t *testing.T) {
	// Test case 1: Completely empty table
	var buf bytes.Buffer
	table := tablewriter.NewTable(&buf,
		tablewriter.WithRenderer(renderer.NewHTML(&buf, false)),
	)
	table.Render()

	expected := `
<table>
</table>`
	if !visualCheckHTML(t, "HTMLEmptyTable", buf.String(), expected) {
		t.Logf("Empty table output: '%s'", buf.String())
		for _, v := range table.Debug() {
			t.Error(v)
		}
	}

	// Test case 2: Header-only table
	buf.Reset()
	tableHeaderOnly := tablewriter.NewTable(&buf,
		tablewriter.WithRenderer(renderer.NewHTML(&buf, false)),
	)
	tableHeaderOnly.SetHeader([]string{"Col A"})
	tableHeaderOnly.Render()

	expectedHeaderOnly := `
<table>
<thead>
<tr><th style="text-align: center;">COL A</th></tr>
</thead>
</table>`
	if !visualCheckHTML(t, "HTMLEmptyTable_HeaderOnly", buf.String(), expectedHeaderOnly) {
		for _, v := range tableHeaderOnly.Debug() {
			t.Error(v)
		}
	}
}

// TestHTMLCSSClasses tests HTML table rendering with custom CSS classes for table, sections, and rows.
func TestHTMLCSSClasses(t *testing.T) {
	var buf bytes.Buffer
	htmlCfg := renderer.HTMLConfig{
		TableClass: "my-table", HeaderClass: "my-thead", BodyClass: "my-tbody",
		FooterClass: "my-tfoot", RowClass: "my-row", HeaderRowClass: "my-header-row",
		FooterRowClass: "my-footer-row",
	}
	table := tablewriter.NewTable(&buf,
		tablewriter.WithRenderer(renderer.NewHTML(&buf, false, htmlCfg)),
		tablewriter.WithHeaderConfig(tablewriter.CellConfig{
			Formatting: tablewriter.CellFormatting{AutoFormat: false, Alignment: tw.AlignCenter},
		}),
	)
	table.SetHeader([]string{"H1"})
	table.Append([]string{"R1"})
	table.SetFooter([]string{"F1"})
	table.Render()

	expected := `
<table class="my-table">
<thead class="my-thead">
<tr class="my-header-row"><th style="text-align: center;">H1</th></tr>
</thead>
<tbody class="my-tbody">
<tr class="my-row"><td style="text-align: left;">R1</td></tr>
</tbody>
<tfoot class="my-tfoot">
<tr class="my-footer-row"><td style="text-align: right;">F1</td></tr>
</tfoot>
</table>`
	if !visualCheckHTML(t, "HTMLCSSClasses", buf.String(), expected) {
		for _, v := range table.Debug() {
			t.Error(v)
		}
	}
}

// TestHTMLStructureStrict verifies the exact HTML structure of a table without whitespace or formatting variations.
func TestHTMLStructureStrict(t *testing.T) {
	var buf bytes.Buffer
	table := tablewriter.NewTable(&buf,
		tablewriter.WithRenderer(renderer.NewHTML(&buf, false)),
	)
	table.SetHeader([]string{"A", "B"})
	table.Append([]string{"1", "2"})
	table.Append([]string{"3", "4"})
	table.SetFooter([]string{"F1", "F2"})
	table.Render()

	expectedStructure := `<table><thead><tr><th style="text-align: center;">A</th><th style="text-align: center;">B</th></tr></thead><tbody><tr><td style="text-align: left;">1</td><td style="text-align: left;">2</td></tr><tr><td style="text-align: left;">3</td><td style="text-align: left;">4</td></tr></tbody><tfoot><tr><td style="text-align: right;">F1</td><td style="text-align: right;">F2</td></tr></tfoot></table>`

	outputNormalized := normalizeHTMLStrict(buf.String())
	if outputNormalized != expectedStructure {
		t.Errorf("HTMLStructureStrict: Mismatch")
		t.Errorf("Expected Structure:\n%s", expectedStructure)
		t.Errorf("Got Normalized Output:\n%s", outputNormalized)
		t.Errorf("Got Raw Output:\n%s", buf.String())
	}
}
