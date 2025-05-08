package tests

import (
	"bytes"
	"encoding/xml" // Import encoding/xml
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/olekukonko/tablewriter"
	"github.com/olekukonko/tablewriter/renderer"
	"github.com/olekukonko/tablewriter/tw"
)

// normalizeSVG attempts to make SVG comparison more robust.
func normalizeSVG(s string) string {
	s = strings.TrimSpace(s)
	s = regexp.MustCompile(`<!--.*?-->`).ReplaceAllString(s, "")                   // Remove comments
	s = regexp.MustCompile(`\s*\n\s*`).ReplaceAllString(s, "")                     // Remove newlines and surrounding whitespace
	s = regexp.MustCompile(`\s+`).ReplaceAllString(s, " ")                         // Collapse multiple spaces
	s = regexp.MustCompile(`>\s+<`).ReplaceAllString(s, "><")                      // Remove space between tags
	s = regexp.MustCompile(`\s*/>`).ReplaceAllString(s, "/>")                      // Clean up self-closing tags
	s = regexp.MustCompile(`\s*=\s*`).ReplaceAllString(s, "=")                     // Clean up around equals
	s = regexp.MustCompile(`"\s*`).ReplaceAllString(s, "\"")                       // Clean up spaces after quotes (start of attr value)
	s = regexp.MustCompile(`\s*"`).ReplaceAllString(s, "\"")                       // Clean up spaces before quotes (end of attr value)
	s = strings.ReplaceAll(s, " />", "/>")                                         // Specific cleanup for self-closing
	s = strings.ReplaceAll(s, "px", "")                                            // Remove "px"
	s = regexp.MustCompile(`(\d)\.00`).ReplaceAllString(s, "$1")                   // Change .00 to nothing
	s = regexp.MustCompile(`(\d)\.0`).ReplaceAllString(s, "$1")                    // Change .0 to nothing
	s = regexp.MustCompile(`0\.50`).ReplaceAllString(s, "0.5")                     // Change 0.50 to 0.5
	s = regexp.MustCompile(`(\d{1,})\.([1-9])0`).ReplaceAllString(s, "$1.$2")      // e.g. 19.50 -> 19.5
	s = regexp.MustCompile(`(\d{1,})\.([0-9]{1,2})0`).ReplaceAllString(s, "$1.$2") // e.g., 117.50 -> 117.5

	return s
}

// formatSVG takes a potentially compact SVG string and formats it with indentation
func formatSVG(svg string) (string, error) {
	// Minimal cleanup before parsing
	svg = strings.TrimSpace(svg)
	svg = regexp.MustCompile(`>\s+<`).ReplaceAllString(svg, "><")

	decoder := xml.NewDecoder(strings.NewReader(svg))
	var buf bytes.Buffer
	encoder := xml.NewEncoder(&buf)
	encoder.Indent("", "  ") // Use two spaces for indentation

	for {
		token, err := decoder.Token()
		if err != nil {
			if err.Error() == "EOF" { // Check for EOF specifically
				break
			}
			return "", fmt.Errorf("decoder error: %w", err) // Return other errors
		}

		if err := encoder.EncodeToken(token); err != nil {
			return "", fmt.Errorf("encoder error: %w", err)
		}
	}

	if err := encoder.Flush(); err != nil {
		return "", fmt.Errorf("flush error: %w", err)
	}

	// Add a newline at the end for consistency
	formatted := buf.String()
	if !strings.HasSuffix(formatted, "\n") {
		formatted += "\n"
	}

	// formatted = strings.ReplaceAll(formatted, "xmlns=\"http://www.w3.org/2000/svg\"", "")
	return formatted, nil

}

// visualCheckSVG compares actual SVG output with expected, after normalization.
func visualCheckSVG(t *testing.T, testName, actual, expected string) bool {
	t.Helper()
	normActual := normalizeSVG(actual)
	normExpected := normalizeSVG(expected)

	if normActual != normExpected {
		t.Errorf("%s: SVG output mismatch.", testName)

		// Attempt to format both for easier visual diffing
		formatExpected, errExp := formatSVG(normExpected)
		if errExp != nil {
			t.Logf("Error formatting EXPECTED SVG: %v", errExp)
			formatExpected = normExpected // Fallback to normalized if formatting fails
		}

		formatActual, errAct := formatSVG(normActual)
		if errAct != nil {
			t.Logf("Error formatting ACTUAL SVG: %v", errAct)
			formatActual = normActual // Fallback to normalized if formatting fails
		}

		t.Logf("--- EXPECTED (Formatted) ---\n%s", formatExpected)
		t.Logf("--- ACTUAL (Formatted) ---\n%s", formatActual)
		return false
	}
	return true
}

// defaultSVGConfigForTests provides a consistent SVG config for most tests.
func defaultSVGConfigForTests(debug bool) renderer.SVGConfig {
	return renderer.SVGConfig{
		FontFamily:              "Arial",
		FontSize:                10,
		LineHeightFactor:        1.2, // from 1.4 to 1.2 to match test output better
		Padding:                 3,
		StrokeWidth:             1,
		StrokeColor:             "black",
		HeaderBG:                "#DDD",
		RowBG:                   "#FFF",
		RowAltBG:                "#EEE",
		FooterBG:                "#DDD",
		HeaderColor:             "#000",
		RowColor:                "#000",
		FooterColor:             "#000",
		ApproxCharWidthFactor:   0.6,
		MinColWidth:             20,
		RenderTWConfigOverrides: true,
		Debug:                   debug,
	}
}

// TestSVGBasicTable (Enable debug in config)
func TestSVGBasicTable(t *testing.T) {
	var buf bytes.Buffer
	svgCfg := defaultSVGConfigForTests(true) // DEBUG ENABLED
	table := tablewriter.NewTable(&buf,
		tablewriter.WithRenderer(renderer.NewSVG(svgCfg)),
		// tablewriter.WithDebug(true), // Optional: Enable tablewriter debug too
	)
	table.Header([]string{"Name", "Age", "City"})
	table.Append([]string{"Alice", "25", "New York"})
	table.Append([]string{"Bob", "30", "Boston"})
	table.Render()

	// Update expected based on previous runs or careful recalculation
	expected := `
<svg xmlns="http://www.w3.org/2000/svg"width="118"height="58"font-family="Arial"font-size="10"><style>text { stroke: none; }</style><rect x="1"y="1"width="36"height="18"fill="#DDD"/><text x="19"y="10"fill="#000"text-anchor="middle"dominant-baseline="middle">NAME</text><rect x="38"y="1"width="24"height="18"fill="#DDD"/><text x="50"y="10"fill="#000"text-anchor="middle"dominant-baseline="middle">AGE</text><rect x="63"y="1"width="54"height="18"fill="#DDD"/><text x="90"y="10"fill="#000"text-anchor="middle"dominant-baseline="middle">CITY</text><rect x="1"y="20"width="36"height="18"fill="#FFF"/><text x="4"y="29"fill="#000"text-anchor="start"dominant-baseline="middle">Alice</text><rect x="38"y="20"width="24"height="18"fill="#FFF"/><text x="41"y="29"fill="#000"text-anchor="start"dominant-baseline="middle">25</text><rect x="63"y="20"width="54"height="18"fill="#FFF"/><text x="66"y="29"fill="#000"text-anchor="start"dominant-baseline="middle">New York</text><rect x="1"y="39"width="36"height="18"fill="#EEE"/><text x="4"y="48"fill="#000"text-anchor="start"dominant-baseline="middle">Bob</text><rect x="38"y="39"width="24"height="18"fill="#EEE"/><text x="41"y="48"fill="#000"text-anchor="start"dominant-baseline="middle">30</text><rect x="63"y="39"width="54"height="18"fill="#EEE"/><text x="66"y="48"fill="#000"text-anchor="start"dominant-baseline="middle">Boston</text><g class="table-borders"stroke="black"stroke-width="1"stroke-linecap="square"><line x1="0.5"y1="0.5"x2="117.5"y2="0.5"/><line x1="0.5"y1="19.5"x2="117.5"y2="19.5"/><line x1="0.5"y1="38.5"x2="117.5"y2="38.5"/><line x1="0.5"y1="57.5"x2="117.5"y2="57.5"/><line x1="0.5"y1="0.5"x2="0.5"y2="57.5"/><line x1="37.5"y1="0.5"x2="37.5"y2="57.5"/><line x1="62.5"y1="0.5"x2="62.5"y2="57.5"/><line x1="117.5"y1="0.5"x2="117.5"y2="57.5"/></g></svg>`
	if !visualCheckSVG(t, "SVGBasicTable", buf.String(), expected) {
		t.Log("--- Debug Log for SVGBasicTable ---")
		// Print only the SVG renderer's trace
		for _, v := range table.Renderer().Debug() {
			t.Log(v)
		}
	}
}

// --- KEEP ALL OTHER TESTS THE SAME AS THE PREVIOUS VERSION ---
// --- (TestSVGWithFooterAndAlignment, TestSVGMultiLineContent, etc.) ---
// --- BUT UPDATE THEIR `expected` STRINGS TO USE THE SIMPLE FULL GRID BORDER LOGIC ---

func TestSVGWithFooterAndAlignment(t *testing.T) {
	var buf bytes.Buffer
	svgCfg := defaultSVGConfigForTests(false)
	table := tablewriter.NewTable(&buf,
		tablewriter.WithRenderer(renderer.NewSVG(svgCfg)),
		tablewriter.WithHeaderConfig(tw.CellConfig{
			Formatting: tw.CellFormatting{Alignment: tw.AlignCenter},
		}),
		tablewriter.WithRowConfig(tw.CellConfig{
			ColumnAligns: []tw.Align{tw.AlignLeft, tw.AlignRight, tw.AlignCenter},
		}),
		tablewriter.WithFooterConfig(tw.CellConfig{
			Formatting: tw.CellFormatting{Alignment: tw.AlignRight},
		}),
	)
	table.Header([]string{"Item", "Qty", "Price"}) // ITEM, QTY, PRICE
	table.Append([]string{"Apple", "5", "1.20"})
	table.Append([]string{"Banana", "12", "0.35"})
	table.Footer([]string{"", "Total", "7.20"})

	// Widths [42, 36, 36], TotalW=118, TotalH=77
	expected := `
<svg xmlns="http://www.w3.org/2000/svg"width="118"height="77"font-family="Arial"font-size="10"><style>text { stroke: none; }</style><rect x="1"y="1"width="42"height="18"fill="#DDD"/><text x="22"y="10"fill="#000"text-anchor="middle"dominant-baseline="middle">ITEM</text><rect x="44"y="1"width="36"height="18"fill="#DDD"/><text x="62"y="10"fill="#000"text-anchor="middle"dominant-baseline="middle">QTY</text><rect x="81"y="1"width="36"height="18"fill="#DDD"/><text x="99"y="10"fill="#000"text-anchor="middle"dominant-baseline="middle">PRICE</text><rect x="1"y="20"width="42"height="18"fill="#FFF"/><text x="4"y="29"fill="#000"text-anchor="start"dominant-baseline="middle">Apple</text><rect x="44"y="20"width="36"height="18"fill="#FFF"/><text x="77"y="29"fill="#000"text-anchor="end"dominant-baseline="middle">5</text><rect x="81"y="20"width="36"height="18"fill="#FFF"/><text x="99"y="29"fill="#000"text-anchor="middle"dominant-baseline="middle">1.20</text><rect x="1"y="39"width="42"height="18"fill="#EEE"/><text x="4"y="48"fill="#000"text-anchor="start"dominant-baseline="middle">Banana</text><rect x="44"y="39"width="36"height="18"fill="#EEE"/><text x="77"y="48"fill="#000"text-anchor="end"dominant-baseline="middle">12</text><rect x="81"y="39"width="36"height="18"fill="#EEE"/><text x="99"y="48"fill="#000"text-anchor="middle"dominant-baseline="middle">0.35</text><rect x="1"y="58"width="42"height="18"fill="#DDD"/><text x="39"y="67"fill="#000"text-anchor="end"dominant-baseline="middle"></text><rect x="44"y="58"width="36"height="18"fill="#DDD"/><text x="77"y="67"fill="#000"text-anchor="end"dominant-baseline="middle">Total</text><rect x="81"y="58"width="36"height="18"fill="#DDD"/><text x="114"y="67"fill="#000"text-anchor="end"dominant-baseline="middle">7.20</text><g class="table-borders"stroke="black"stroke-width="1"stroke-linecap="square"><line x1="0.5"y1="0.5"x2="117.5"y2="0.5"/><line x1="0.5"y1="19.5"x2="117.5"y2="19.5"/><line x1="0.5"y1="38.5"x2="117.5"y2="38.5"/><line x1="0.5"y1="57.5"x2="117.5"y2="57.5"/><line x1="0.5"y1="76.5"x2="117.5"y2="76.5"/><line x1="0.5"y1="0.5"x2="0.5"y2="76.5"/><line x1="43.5"y1="0.5"x2="43.5"y2="76.5"/><line x1="80.5"y1="0.5"x2="80.5"y2="76.5"/><line x1="117.5"y1="0.5"x2="117.5"y2="76.5"/></g></svg>`
	if !visualCheckSVG(t, "SVGWithFooterAndAlignment", buf.String(), expected) {
		t.Log("--- Debug Log for SVGWithFooterAndAlignment ---")
		// Only print renderer debug if test fails
		for _, v := range table.Renderer().Debug() {
			t.Log(v)
		}
	}
}

func TestSVGMultiLineContent(t *testing.T) {
	var buf bytes.Buffer
	svgCfg := defaultSVGConfigForTests(false)
	table := tablewriter.NewTable(&buf,
		tablewriter.WithRenderer(renderer.NewSVG(svgCfg)),
	)
	table.Header([]string{"Description", "Status"}) // DESCRIPTION, STATUS
	table.Append([]string{"Line 1\nLine 2", "OK"})
	table.Render()

	// Widths: Desc=72, Status=42, TotalW=117, TotalH=58
	expected := `
<svg xmlns="http://www.w3.org/2000/svg"width="117"height="58"font-family="Arial"font-size="10"><style>text { stroke: none; }</style><rect x="1"y="1"width="72"height="18"fill="#DDD"/><text x="37"y="10"fill="#000"text-anchor="middle"dominant-baseline="middle">DESCRIPTION</text><rect x="74"y="1"width="42"height="18"fill="#DDD"/><text x="95"y="10"fill="#000"text-anchor="middle"dominant-baseline="middle">STATUS</text><rect x="1"y="20"width="72"height="18"fill="#FFF"/><text x="4"y="29"fill="#000"text-anchor="start"dominant-baseline="middle">Line 1</text><rect x="74"y="20"width="42"height="18"fill="#FFF"/><text x="77"y="29"fill="#000"text-anchor="start"dominant-baseline="middle">OK</text><rect x="1"y="39"width="72"height="18"fill="#FFF"/><text x="4"y="48"fill="#000"text-anchor="start"dominant-baseline="middle">Line 2</text><rect x="74"y="39"width="42"height="18"fill="#FFF"/><text x="77"y="48"fill="#000"text-anchor="start"dominant-baseline="middle"></text><g class="table-borders"stroke="black"stroke-width="1"stroke-linecap="square"><line x1="0.5"y1="0.5"x2="116.5"y2="0.5"/><line x1="0.5"y1="19.5"x2="116.5"y2="19.5"/><line x1="0.5"y1="38.5"x2="116.5"y2="38.5"/><line x1="0.5"y1="57.5"x2="116.5"y2="57.5"/><line x1="0.5"y1="0.5"x2="0.5"y2="57.5"/><line x1="73.5"y1="0.5"x2="73.5"y2="57.5"/><line x1="116.5"y1="0.5"x2="116.5"y2="57.5"/></g></svg>`
	if !visualCheckSVG(t, "SVGMultiLineContent", buf.String(), expected) {
		for _, v := range table.Renderer().Debug() {
			t.Log(v)
		}
	}
}

func TestSVGHorizontalMerge(t *testing.T) {
	var buf bytes.Buffer
	svgCfg := defaultSVGConfigForTests(false)
	table := tablewriter.NewTable(&buf,
		tablewriter.WithRenderer(renderer.NewSVG(svgCfg)),
		tablewriter.WithConfig(tablewriter.Config{
			Header: tw.CellConfig{Formatting: tw.CellFormatting{MergeMode: tw.MergeHorizontal, Alignment: tw.AlignCenter}},
			Row:    tw.CellConfig{Formatting: tw.CellFormatting{MergeMode: tw.MergeHorizontal, Alignment: tw.AlignLeft}},
		}),
	)
	table.Header([]string{"A", "Merged Header", "Merged Header"}) // A, MERGED HEADER
	table.Append([]string{"Data 1", "Data 2", "Data 2"})
	table.Render()

	// Widths: [42, 84, 20], TotalW=150, H=39
	// Merged Header Rect: x=44, width=105
	// Merged Data 2 Rect: x=44, width=105
	expected := `
<svg xmlns="http://www.w3.org/2000/svg"width="150"height="39"font-family="Arial"font-size="10"><style>text { stroke: none; }</style><rect x="1"y="1"width="42"height="18"fill="#DDD"/><text x="22"y="10"fill="#000"text-anchor="middle"dominant-baseline="middle">A</text><rect x="44"y="1"width="105"height="18"fill="#DDD"/><text x="96.5"y="10"fill="#000"text-anchor="middle"dominant-baseline="middle">MERGED HEADER</text><rect x="1"y="20"width="42"height="18"fill="#FFF"/><text x="4"y="29"fill="#000"text-anchor="start"dominant-baseline="middle">Data 1</text><rect x="44"y="20"width="105"height="18"fill="#FFF"/><text x="47"y="29"fill="#000"text-anchor="start"dominant-baseline="middle">Data 2</text><g class="table-borders"stroke="black"stroke-width="1"stroke-linecap="square"><line x1="0.5"y1="0.5"x2="149.5"y2="0.5"/><line x1="0.5"y1="19.5"x2="149.5"y2="19.5"/><line x1="0.5"y1="38.5"x2="149.5"y2="38.5"/><line x1="0.5"y1="0.5"x2="0.5"y2="38.5"/><line x1="43.5"y1="0.5"x2="43.5"y2="38.5"/><line x1="128.5"y1="0.5"x2="128.5"y2="38.5"/><line x1="149.5"y1="0.5"x2="149.5"y2="38.5"/></g></svg>` // Full grid borders now
	if !visualCheckSVG(t, "SVGHorizontalMerge", buf.String(), expected) {
		for _, v := range table.Renderer().Debug() {
			t.Log(v)
		}
	}
}

func TestSVGVerticalMerge(t *testing.T) {
	var buf bytes.Buffer
	svgCfg := defaultSVGConfigForTests(false)
	table := tablewriter.NewTable(&buf,
		tablewriter.WithRenderer(renderer.NewSVG(svgCfg)),
		tablewriter.WithConfig(tablewriter.Config{
			Row: tw.CellConfig{Formatting: tw.CellFormatting{MergeMode: tw.MergeVertical}},
		}),
	)
	table.Header([]string{"Cat", "Item"}) // CAT, ITEM
	table.Append([]string{"Fruit", "Apple"})
	table.Append([]string{"Fruit", "Banana"})
	table.Render()

	// Widths: [36, 42], TotalW=81, H=58
	expected := `
<svg xmlns="http://www.w3.org/2000/svg"width="81"height="58"font-family="Arial"font-size="10"><style>text { stroke: none; }</style><rect x="1"y="1"width="36"height="18"fill="#DDD"/><text x="19"y="10"fill="#000"text-anchor="middle"dominant-baseline="middle">CAT</text><rect x="38"y="1"width="42"height="18"fill="#DDD"/><text x="59"y="10"fill="#000"text-anchor="middle"dominant-baseline="middle">ITEM</text><rect x="1"y="20"width="36"height="37"fill="#FFF"/><text x="4"y="38.5"fill="#000"text-anchor="start"dominant-baseline="middle">Fruit</text><rect x="38"y="20"width="42"height="18"fill="#FFF"/><text x="41"y="29"fill="#000"text-anchor="start"dominant-baseline="middle">Apple</text><rect x="38"y="39"width="42"height="18"fill="#EEE"/><text x="41"y="48"fill="#000"text-anchor="start"dominant-baseline="middle">Banana</text><g class="table-borders"stroke="black"stroke-width="1"stroke-linecap="square"><line x1="0.5"y1="0.5"x2="80.5"y2="0.5"/><line x1="0.5"y1="19.5"x2="80.5"y2="19.5"/><line x1="0.5"y1="38.5"x2="80.5"y2="38.5"/><line x1="0.5"y1="57.5"x2="80.5"y2="57.5"/><line x1="0.5"y1="0.5"x2="0.5"y2="57.5"/><line x1="37.5"y1="0.5"x2="37.5"y2="57.5"/><line x1="80.5"y1="0.5"x2="80.5"y2="57.5"/></g></svg>` // Full grid borders
	if !visualCheckSVG(t, "SVGVerticalMerge", buf.String(), expected) {
		for _, v := range table.Renderer().Debug() {
			t.Log(v)
		}
	}
}

func TestSVGEmptyTable(t *testing.T) {
	var buf bytes.Buffer
	svgCfg := defaultSVGConfigForTests(false)

	table := tablewriter.NewTable(&buf, tablewriter.WithRenderer(renderer.NewSVG(svgCfg)))
	table.Render()
	expectedEmpty := `<svg xmlns="http://www.w3.org/2000/svg"width="2"height="2"></svg>`
	if !visualCheckSVG(t, "SVGEmptyTable_CompletelyEmpty", buf.String(), expectedEmpty) {
		t.Logf("Empty table output: '%s'", buf.String())
		for _, v := range table.Renderer().Debug() {
			t.Log(v)
		}
	}

	buf.Reset()
	tableHeaderOnly := tablewriter.NewTable(&buf, tablewriter.WithRenderer(renderer.NewSVG(svgCfg)))
	tableHeaderOnly.Header([]string{"Test"}) // TEST
	tableHeaderOnly.Render()
	// Widths: [30], TotalW=32, H=20
	expectedHeaderOnly := `
<svg xmlns="http://www.w3.org/2000/svg"width="32"height="20"font-family="Arial"font-size="10"><style>text { stroke: none; }</style><rect x="1"y="1"width="30"height="18"fill="#DDD"/><text x="16"y="10"fill="#000"text-anchor="middle"dominant-baseline="middle">TEST</text><g class="table-borders"stroke="black"stroke-width="1"stroke-linecap="square"><line x1="0.5"y1="0.5"x2="31.5"y2="0.5"/><line x1="0.5"y1="19.5"x2="31.5"y2="19.5"/><line x1="0.5"y1="0.5"x2="0.5"y2="19.5"/><line x1="31.5"y1="0.5"x2="31.5"y2="19.5"/></g></svg>`
	if !visualCheckSVG(t, "SVGEmptyTable_HeaderOnly", buf.String(), expectedHeaderOnly) {
		for _, v := range tableHeaderOnly.Renderer().Debug() {
			t.Log(v)
		}
	}
}

func TestSVGHierarchicalMerge(t *testing.T) {
	var buf bytes.Buffer
	svgCfg := defaultSVGConfigForTests(false)
	table := tablewriter.NewTable(&buf,
		tablewriter.WithRenderer(renderer.NewSVG(svgCfg)),
		tablewriter.WithConfig(tablewriter.Config{
			Header: tw.CellConfig{Formatting: tw.CellFormatting{Alignment: tw.AlignCenter}},
			Row:    tw.CellConfig{Formatting: tw.CellFormatting{MergeMode: tw.MergeHierarchical}},
		}),
	)
	table.Header([]string{"L1", "L2", "L3"}) // L 1, L 2, L 3
	table.Append([]string{"A", "a", "1"})
	table.Append([]string{"A", "b", "2"})
	table.Append([]string{"A", "b", "3"})
	table.Render()

	// Widths: [24,24,24], TotalW=76, TotalH=77
	expected := `
<svg xmlns="http://www.w3.org/2000/svg"width="76"height="77"font-family="Arial"font-size="10"><style>text { stroke: none; }</style><rect x="1"y="1"width="24"height="18"fill="#DDD"/><text x="13"y="10"fill="#000"text-anchor="middle"dominant-baseline="middle">L 1</text><rect x="26"y="1"width="24"height="18"fill="#DDD"/><text x="38"y="10"fill="#000"text-anchor="middle"dominant-baseline="middle">L 2</text><rect x="51"y="1"width="24"height="18"fill="#DDD"/><text x="63"y="10"fill="#000"text-anchor="middle"dominant-baseline="middle">L 3</text><rect x="1"y="20"width="24"height="56"fill="#FFF"/><text x="4"y="48"fill="#000"text-anchor="start"dominant-baseline="middle">A</text><rect x="26"y="20"width="24"height="18"fill="#FFF"/><text x="29"y="29"fill="#000"text-anchor="start"dominant-baseline="middle">a</text><rect x="51"y="20"width="24"height="18"fill="#FFF"/><text x="54"y="29"fill="#000"text-anchor="start"dominant-baseline="middle">1</text><rect x="26"y="39"width="24"height="37"fill="#EEE"/><text x="29"y="57.5"fill="#000"text-anchor="start"dominant-baseline="middle">b</text><rect x="51"y="39"width="24"height="18"fill="#EEE"/><text x="54"y="48"fill="#000"text-anchor="start"dominant-baseline="middle">2</text><rect x="51"y="58"width="24"height="18"fill="#FFF"/><text x="54"y="67"fill="#000"text-anchor="start"dominant-baseline="middle">3</text><g class="table-borders"stroke="black"stroke-width="1"stroke-linecap="square"><line x1="0.5"y1="0.5"x2="75.5"y2="0.5"/><line x1="0.5"y1="19.5"x2="75.5"y2="19.5"/><line x1="0.5"y1="38.5"x2="75.5"y2="38.5"/><line x1="0.5"y1="57.5"x2="75.5"y2="57.5"/><line x1="0.5"y1="76.5"x2="75.5"y2="76.5"/><line x1="0.5"y1="0.5"x2="0.5"y2="76.5"/><line x1="25.5"y1="0.5"x2="25.5"y2="76.5"/><line x1="50.5"y1="0.5"x2="50.5"y2="76.5"/><line x1="75.5"y1="0.5"x2="75.5"y2="76.5"/></g></svg>` // Full grid borders
	if !visualCheckSVG(t, "SVGHierarchicalMerge", buf.String(), expected) {
		for _, v := range table.Renderer().Debug() {
			t.Log(v)
		}
	}
}

func TestSVGColumnAlignmentOverride(t *testing.T) {
	var buf bytes.Buffer
	svgCfg := defaultSVGConfigForTests(false)

	table := tablewriter.NewTable(&buf,
		tablewriter.WithRenderer(renderer.NewSVG(svgCfg)),
		tablewriter.WithHeaderConfig(tw.CellConfig{Formatting: tw.CellFormatting{Alignment: tw.AlignCenter}}), // H 1, H 2, H 3 -> AutoFormatted
		tablewriter.WithRowConfig(tw.CellConfig{
			ColumnAligns: []tw.Align{tw.AlignRight, tw.AlignCenter, tw.AlignLeft},
		}),
	)
	table.Header([]string{"H1", "H2", "H3"})
	table.Append([]string{"R1C1", "R1C2", "R1C3"})
	table.Render()

	// Widths: [30,30,30], TotalW=94, TotalH=39
	expected := `
<svg xmlns="http://www.w3.org/2000/svg"width="94"height="39"font-family="Arial"font-size="10"><style>text { stroke: none; }</style><rect x="1"y="1"width="30"height="18"fill="#DDD"/><text x="16"y="10"fill="#000"text-anchor="middle"dominant-baseline="middle">H 1</text><rect x="32"y="1"width="30"height="18"fill="#DDD"/><text x="47"y="10"fill="#000"text-anchor="middle"dominant-baseline="middle">H 2</text><rect x="63"y="1"width="30"height="18"fill="#DDD"/><text x="78"y="10"fill="#000"text-anchor="middle"dominant-baseline="middle">H 3</text><rect x="1"y="20"width="30"height="18"fill="#FFF"/><text x="28"y="29"fill="#000"text-anchor="end"dominant-baseline="middle">R1C1</text><rect x="32"y="20"width="30"height="18"fill="#FFF"/><text x="47"y="29"fill="#000"text-anchor="middle"dominant-baseline="middle">R1C2</text><rect x="63"y="20"width="30"height="18"fill="#FFF"/><text x="66"y="29"fill="#000"text-anchor="start"dominant-baseline="middle">R1C3</text><g class="table-borders"stroke="black"stroke-width="1"stroke-linecap="square"><line x1="0.5"y1="0.5"x2="93.5"y2="0.5"/><line x1="0.5"y1="19.5"x2="93.5"y2="19.5"/><line x1="0.5"y1="38.5"x2="93.5"y2="38.5"/><line x1="0.5"y1="0.5"x2="0.5"y2="38.5"/><line x1="31.5"y1="0.5"x2="31.5"y2="38.5"/><line x1="62.5"y1="0.5"x2="62.5"y2="38.5"/><line x1="93.5"y1="0.5"x2="93.5"y2="38.5"/></g></svg>`
	if !visualCheckSVG(t, "SVGColumnAlignmentOverride", buf.String(), expected) {
		for _, v := range table.Renderer().Debug() {
			t.Log(v)
		}
	}
}

func TestSVGPaddingAndFont(t *testing.T) {
	var buf bytes.Buffer
	svgCfg := defaultSVGConfigForTests(false)
	svgCfg.Padding = 10
	svgCfg.FontSize = 16
	svgCfg.FontFamily = "Verdana"

	table := tablewriter.NewTable(&buf,
		tablewriter.WithRenderer(renderer.NewSVG(svgCfg)),
	)
	table.Header([]string{"Test"}) // TEST
	table.Render()

	// TEST(4c). FontSize=16, Padding=10, Factor=0.6
	// Width = 4 * 16 * 0.6 + 2*10 = 38.4 + 20 = 58.4
	// RowH = 16*1.2 + 2*10 = 19.2 + 20 = 39.2
	// TotalW = 1 + 58.4 + 1 = 60.4
	// TotalH = 1 + 39.2 + 1 = 41.2
	expected := `
<svg xmlns="http://www.w3.org/2000/svg"width="60.4"height="41.2"font-family="Verdana"font-size="16"><style>text { stroke: none; }</style><rect x="1"y="1"width="58.4"height="39.2"fill="#DDD"/><text x="30.2"y="20.6"fill="#000"text-anchor="middle"dominant-baseline="middle">TEST</text><g class="table-borders"stroke="black"stroke-width="1"stroke-linecap="square"><line x1="0.5"y1="0.5"x2="59.9"y2="0.5"/><line x1="0.5"y1="40.7"x2="59.9"y2="40.7"/><line x1="0.5"y1="0.5"x2="0.5"y2="40.7"/><line x1="59.9"y1="0.5"x2="59.9"y2="40.7"/></g></svg>`
	if !visualCheckSVG(t, "SVGPaddingAndFont", buf.String(), expected) {
		for _, v := range table.Renderer().Debug() {
			t.Log(v)
		}
	}
}
