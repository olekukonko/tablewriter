package tests

import (
	"bytes"
	"encoding/xml"
	"math"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/olekukonko/tablewriter"
	"github.com/olekukonko/tablewriter/renderer"
	"github.com/olekukonko/tablewriter/tw"
)

// defaultSVGConfigForTests provides a consistent SVG config for tests.
// Parameter debug enables debug logging if true.
// Returns an SVGConfig struct with default settings.
func defaultSVGConfigForTests(debug bool) renderer.SVGConfig {
	return renderer.SVGConfig{
		FontFamily:              "Arial",
		FontSize:                10,
		LineHeightFactor:        1.2,
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

// xmlNode represents a node in the XML tree for structural comparison
type xmlNode struct {
	XMLName xml.Name
	Attrs   []xml.Attr `xml:",any,attr"`
	Content string     `xml:",chardata"`
	Nodes   []xmlNode  `xml:",any"`
}

// normalizeSVG performs lightweight normalization for fast comparison
func normalizeSVG(s string) string {
	s = strings.TrimSpace(s)

	// Remove comments
	s = regexp.MustCompile(`<!--.*?-->`).ReplaceAllString(s, "")

	// Normalize whitespace
	s = regexp.MustCompile(`\s+`).ReplaceAllString(s, " ")
	s = regexp.MustCompile(`>\s+<`).ReplaceAllString(s, "><")

	// Normalize self-closing tags
	s = regexp.MustCompile(`\s*/>`).ReplaceAllString(s, "/>")

	// Normalize numeric values (1.0 -> 1, 0.500 -> 0.5)
	s = regexp.MustCompile(`(\d+)\.0([^0-9])`).ReplaceAllString(s, "$1$2")
	s = regexp.MustCompile(`(\d+\.\d+)0([^0-9])`).ReplaceAllString(s, "$1$2")
	s = regexp.MustCompile(`\.0([^0-9])`).ReplaceAllString(s, "$1")

	return s
}

// visualCheckSVG is the main comparison function
func visualCheckSVG(t *testing.T, testName, actual, expected string) bool {
	t.Helper()

	// First try normalized string comparison
	normActual := normalizeSVG(actual)
	normExpected := normalizeSVG(expected)

	if normActual == normExpected {
		return true
	}

	// If strings differ, try structural comparison
	ok, err := compareSVGStructure(normActual, normExpected)
	if err != nil {
		t.Logf("Structural comparison failed: %v", err)
	} else if ok {
		return true
	}

	// Both comparisons failed - show formatted diff
	t.Errorf("%s: SVG output mismatch", testName)
	showFormattedDiff(t, normActual, normExpected)
	return false
}

// compareSVGStructure does structural XML comparison
func compareSVGStructure(actual, expected string) (bool, error) {
	var actualNode, expectedNode xmlNode

	if err := xml.Unmarshal([]byte(actual), &actualNode); err != nil {
		return false, err
	}
	if err := xml.Unmarshal([]byte(expected), &expectedNode); err != nil {
		return false, err
	}

	return compareXMLNodes(actualNode, expectedNode), nil
}

// compareXMLNodes recursively compares XML nodes
func compareXMLNodes(a, b xmlNode) bool {
	if a.XMLName != b.XMLName {
		return false
	}

	// Compare attributes
	if len(a.Attrs) != len(b.Attrs) {
		return false
	}

	attrMap := make(map[xml.Name]string)
	for _, attr := range a.Attrs {
		attrMap[attr.Name] = attr.Value
	}

	for _, attr := range b.Attrs {
		aVal, ok := attrMap[attr.Name]
		if !ok {
			return false
		}

		if isNumericAttribute(attr.Name.Local) {
			if !compareFloatStrings(aVal, attr.Value) {
				return false
			}
		} else if aVal != attr.Value {
			return false
		}
	}

	// Compare content
	if strings.TrimSpace(a.Content) != strings.TrimSpace(b.Content) {
		return false
	}

	// Compare children
	if len(a.Nodes) != len(b.Nodes) {
		return false
	}

	for i := range a.Nodes {
		if !compareXMLNodes(a.Nodes[i], b.Nodes[i]) {
			return false
		}
	}

	return true
}

// Helper functions
func isNumericAttribute(name string) bool {
	numericAttrs := map[string]bool{
		"x": true, "y": true, "width": true, "height": true,
		"x1": true, "y1": true, "x2": true, "y2": true,
		"font-size": true, "stroke-width": true,
	}
	return numericAttrs[name]
}

func compareFloatStrings(a, b string) bool {
	f1, err1 := strconv.ParseFloat(a, 64)
	f2, err2 := strconv.ParseFloat(b, 64)

	if err1 != nil || err2 != nil {
		return a == b
	}

	const epsilon = 0.0001
	return math.Abs(f1-f2) < epsilon
}

func showFormattedDiff(t *testing.T, actual, expected string) {
	t.Helper()
	format := func(s string) string {
		var buf bytes.Buffer
		enc := xml.NewEncoder(&buf)
		enc.Indent("", "  ")
		if err := enc.Encode(xml.NewDecoder(strings.NewReader(s))); err != nil {
			return s
		}
		return buf.String()
	}

	t.Logf("--- EXPECTED (Formatted) ---\n%s", format(expected))
	t.Logf("--- ACTUAL (Formatted) ---\n%s", format(actual))
}

// TestSVGBasicTable tests rendering a basic SVG table.
// Parameter t is the testing context.
// No return value; logs debug info on failure.
func TestSVGBasicTable(t *testing.T) {
	var buf bytes.Buffer
	svgCfg := defaultSVGConfigForTests(true)
	table := tablewriter.NewTable(&buf, tablewriter.WithRenderer(renderer.NewSVG(svgCfg)))
	table.Header([]string{"Name", "Age", "City"})
	table.Append([]string{"Alice", "25", "New York"})
	table.Append([]string{"Bob", "30", "Boston"})
	table.Render()
	expected := `<svg xmlns="http://www.w3.org/2000/svg"width="118"height="58"font-family="Arial"font-size="10"><style>text { stroke: none; }</style><rect x="1"y="1"width="36"height="18"fill="#DDD"/><text x="19"y="10"fill="#000"text-anchor="middle"dominant-baseline="middle">NAME</text><rect x="38"y="1"width="24"height="18"fill="#DDD"/><text x="50"y="10"fill="#000"text-anchor="middle"dominant-baseline="middle">AGE</text><rect x="63"y="1"width="54"height="18"fill="#DDD"/><text x="90"y="10"fill="#000"text-anchor="middle"dominant-baseline="middle">CITY</text><rect x="1"y="20"width="36"height="18"fill="#FFF"/><text x="4"y="29"fill="#000"text-anchor="start"dominant-baseline="middle">Alice</text><rect x="38"y="20"width="24"height="18"fill="#FFF"/><text x="41"y="29"fill="#000"text-anchor="start"dominant-baseline="middle">25</text><rect x="63"y="20"width="54"height="18"fill="#FFF"/><text x="66"y="29"fill="#000"text-anchor="start"dominant-baseline="middle">New York</text><rect x="1"y="39"width="36"height="18"fill="#EEE"/><text x="4"y="48"fill="#000"text-anchor="start"dominant-baseline="middle">Bob</text><rect x="38"y="39"width="24"height="18"fill="#EEE"/><text x="41"y="48"fill="#000"text-anchor="start"dominant-baseline="middle">30</text><rect x="63"y="39"width="54"height="18"fill="#EEE"/><text x="66"y="48"fill="#000"text-anchor="start"dominant-baseline="middle">Boston</text><g class="table-borders"stroke="black"stroke-width="1"stroke-linecap="square"><line x1="0.5"y1="0.5"x2="117.5"y2="0.5"/><line x1="0.5"y1="19.5"x2="117.5"y2="19.5"/><line x1="0.5"y1="38.5"x2="117.5"y2="38.5"/><line x1="0.5"y1="57.5"x2="117.5"y2="57.5"/><line x1="0.5"y1="0.5"x2="0.5"y2="57.5"/><line x1="37.5"y1="0.5"x2="37.5"y2="57.5"/><line x1="62.5"y1="0.5"x2="62.5"y2="57.5"/><line x1="117.5"y1="0.5"x2="117.5"y2="57.5"/></g></svg>`
	if !visualCheckSVG(t, "SVGBasicTable", buf.String(), expected) {
		t.Log("--- Debug Log for SVGBasicTable ---")
		t.Log(table.Debug())
	}
}

// TestSVGEmptyTable tests rendering empty and header-only SVG tables.
// Parameter t is the testing context.
// No return value; logs debug info on failure.
func TestSVGEmptyTable(t *testing.T) {
	var buf bytes.Buffer
	svgCfg := defaultSVGConfigForTests(false)
	table := tablewriter.NewTable(&buf, tablewriter.WithRenderer(renderer.NewSVG(svgCfg)))
	table.Render()
	expectedEmpty := `<svg xmlns="http://www.w3.org/2000/svg"width="2"height="2"></svg>`
	if !visualCheckSVG(t, "SVGEmptyTable_CompletelyEmpty", buf.String(), expectedEmpty) {
		t.Logf("Empty table output: '%s'", buf.String())
		t.Log(table.Debug())
	}
	buf.Reset()
	tableHeaderOnly := tablewriter.NewTable(&buf, tablewriter.WithRenderer(renderer.NewSVG(svgCfg)))
	tableHeaderOnly.Header([]string{"Test"})
	tableHeaderOnly.Render()
	expectedHeaderOnly := `<svg xmlns="http://www.w3.org/2000/svg"width="32"height="20"font-family="Arial"font-size="10"><style>text { stroke: none; }</style><rect x="1"y="1"width="30"height="18"fill="#DDD"/><text x="16"y="10"fill="#000"text-anchor="middle"dominant-baseline="middle">TEST</text><g class="table-borders"stroke="black"stroke-width="1"stroke-linecap="square"><line x1="0.5"y1="0.5"x2="31.5"y2="0.5"/><line x1="0.5"y1="19.5"x2="31.5"y2="19.5"/><line x1="0.5"y1="0.5"x2="0.5"y2="19.5"/><line x1="31.5"y1="0.5"x2="31.5"y2="19.5"/></g></svg>`
	if !visualCheckSVG(t, "SVGEmptyTable_HeaderOnly", buf.String(), expectedHeaderOnly) {
		t.Log(table.Debug())
	}
}

// TestSVGHierarchicalMerge tests SVG rendering with hierarchical merging.
// Parameter t is the testing context.
// No return value; logs debug info on failure.
func TestSVGHierarchicalMerge(t *testing.T) {
	var buf bytes.Buffer
	svgCfg := defaultSVGConfigForTests(false)
	table := tablewriter.NewTable(&buf,
		tablewriter.WithRenderer(renderer.NewSVG(svgCfg)),
		tablewriter.WithConfig(tablewriter.Config{
			Header: tw.CellConfig{Alignment: tw.CellAlignment{Global: tw.AlignCenter}},
			Row:    tw.CellConfig{Merging: tw.CellMerging{Mode: tw.MergeHierarchical}},
		}),
	)
	table.Header([]string{"L1", "L2", "L3"})
	table.Append([]string{"A", "a", "1"})
	table.Append([]string{"A", "b", "2"})
	table.Append([]string{"A", "b", "3"})
	table.Render()
	expected := `<svg xmlns="http://www.w3.org/2000/svg"width="76"height="77"font-family="Arial"font-size="10"><style>text { stroke: none; }</style><rect x="1"y="1"width="24"height="18"fill="#DDD"/><text x="13"y="10"fill="#000"text-anchor="middle"dominant-baseline="middle">L 1</text><rect x="26"y="1"width="24"height="18"fill="#DDD"/><text x="38"y="10"fill="#000"text-anchor="middle"dominant-baseline="middle">L 2</text><rect x="51"y="1"width="24"height="18"fill="#DDD"/><text x="63"y="10"fill="#000"text-anchor="middle"dominant-baseline="middle">L 3</text><rect x="1"y="20"width="24"height="56"fill="#FFF"/><text x="4"y="48"fill="#000"text-anchor="start"dominant-baseline="middle">A</text><rect x="26"y="20"width="24"height="18"fill="#FFF"/><text x="29"y="29"fill="#000"text-anchor="start"dominant-baseline="middle">a</text><rect x="51"y="20"width="24"height="18"fill="#FFF"/><text x="54"y="29"fill="#000"text-anchor="start"dominant-baseline="middle">1</text><rect x="26"y="39"width="24"height="37"fill="#EEE"/><text x="29"y="57.5"fill="#000"text-anchor="start"dominant-baseline="middle">b</text><rect x="51"y="39"width="24"height="18"fill="#EEE"/><text x="54"y="48"fill="#000"text-anchor="start"dominant-baseline="middle">2</text><rect x="51"y="58"width="24"height="18"fill="#FFF"/><text x="54"y="67"fill="#000"text-anchor="start"dominant-baseline="middle">3</text><g class="table-borders"stroke="black"stroke-width="1"stroke-linecap="square"><line x1="0.5"y1="0.5"x2="75.5"y2="0.5"/><line x1="0.5"y1="19.5"x2="75.5"y2="19.5"/><line x1="0.5"y1="38.5"x2="75.5"y2="38.5"/><line x1="0.5"y1="57.5"x2="75.5"y2="57.5"/><line x1="0.5"y1="76.5"x2="75.5"y2="76.5"/><line x1="0.5"y1="0.5"x2="0.5"y2="76.5"/><line x1="25.5"y1="0.5"x2="25.5"y2="76.5"/><line x1="50.5"y1="0.5"x2="50.5"y2="76.5"/><line x1="75.5"y1="0.5"x2="75.5"y2="76.5"/></g></svg>`
	if !visualCheckSVG(t, "SVGHierarchicalMerge", buf.String(), expected) {
		t.Log(table.Debug())
	}
}

// TestSVGMultiLineContent tests SVG rendering with multi-line content.
// Parameter t is the testing context.
// No return value; logs debug info on failure.
func TestSVGMultiLineContent(t *testing.T) {
	var buf bytes.Buffer
	svgCfg := defaultSVGConfigForTests(false)
	table := tablewriter.NewTable(&buf, tablewriter.WithRenderer(renderer.NewSVG(svgCfg)))
	table.Header([]string{"Description", "Status"})
	table.Append([]string{"Line 1\nLine 2", "OK"})
	table.Render()
	expected := `<svg xmlns="http://www.w3.org/2000/svg"width="117"height="58"font-family="Arial"font-size="10"><style>text { stroke: none; }</style><rect x="1"y="1"width="72"height="18"fill="#DDD"/><text x="37"y="10"fill="#000"text-anchor="middle"dominant-baseline="middle">DESCRIPTION</text><rect x="74"y="1"width="42"height="18"fill="#DDD"/><text x="95"y="10"fill="#000"text-anchor="middle"dominant-baseline="middle">STATUS</text><rect x="1"y="20"width="72"height="18"fill="#FFF"/><text x="4"y="29"fill="#000"text-anchor="start"dominant-baseline="middle">Line 1</text><rect x="74"y="20"width="42"height="18"fill="#FFF"/><text x="77"y="29"fill="#000"text-anchor="start"dominant-baseline="middle">OK</text><rect x="1"y="39"width="72"height="18"fill="#FFF"/><text x="4"y="48"fill="#000"text-anchor="start"dominant-baseline="middle">Line 2</text><rect x="74"y="39"width="42"height="18"fill="#FFF"/><text x="77"y="48"fill="#000"text-anchor="start"dominant-baseline="middle"></text><g class="table-borders"stroke="black"stroke-width="1"stroke-linecap="square"><line x1="0.5"y1="0.5"x2="116.5"y2="0.5"/><line x1="0.5"y1="19.5"x2="116.5"y2="19.5"/><line x1="0.5"y1="38.5"x2="116.5"y2="38.5"/><line x1="0.5"y1="57.5"x2="116.5"y2="57.5"/><line x1="0.5"y1="0.5"x2="0.5"y2="57.5"/><line x1="73.5"y1="0.5"x2="73.5"y2="57.5"/><line x1="116.5"y1="0.5"x2="116.5"y2="57.5"/></g></svg>`
	if !visualCheckSVG(t, "SVGMultiLineContent", buf.String(), expected) {
		t.Log(table.Debug())
	}
}

// TestSVGPaddingAndFont tests SVG rendering with custom padding and font.
// Parameter t is the testing context.
// No return value; logs debug info on failure.
func TestSVGPaddingAndFont(t *testing.T) {
	var buf bytes.Buffer
	svgCfg := defaultSVGConfigForTests(false)
	svgCfg.Padding = 10
	svgCfg.FontSize = 16
	svgCfg.FontFamily = "Verdana"
	table := tablewriter.NewTable(&buf, tablewriter.WithRenderer(renderer.NewSVG(svgCfg)))
	table.Header([]string{"Test"})
	table.Render()
	expected := `<svg xmlns="http://www.w3.org/2000/svg"width="60.4"height="41.2"font-family="Verdana"font-size="16"><style>text { stroke: none; }</style><rect x="1"y="1"width="58.4"height="39.2"fill="#DDD"/><text x="30.2"y="20.6"fill="#000"text-anchor="middle"dominant-baseline="middle">TEST</text><g class="table-borders"stroke="black"stroke-width="1"stroke-linecap="square"><line x1="0.5"y1="0.5"x2="59.9"y2="0.5"/><line x1="0.5"y1="40.7"x2="59.9"y2="40.7"/><line x1="0.5"y1="0.5"x2="0.5"y2="40.7"/><line x1="59.9"y1="0.5"x2="59.9"y2="40.7"/></g></svg>`
	if !visualCheckSVG(t, "SVGPaddingAndFont", buf.String(), expected) {
		t.Log(table.Debug())
	}
}

// TestSVGVerticalMerge tests SVG rendering with vertical merging.
// Parameter t is the testing context.
// No return value; logs debug info on failure.
func TestSVGVerticalMerge(t *testing.T) {
	var buf bytes.Buffer
	svgCfg := defaultSVGConfigForTests(false)
	table := tablewriter.NewTable(&buf,
		tablewriter.WithRenderer(renderer.NewSVG(svgCfg)),
		tablewriter.WithConfig(tablewriter.Config{
			Row: tw.CellConfig{Merging: tw.CellMerging{Mode: tw.MergeVertical}},
		}),
	)
	table.Header([]string{"Cat", "Item"})
	table.Append([]string{"Fruit", "Apple"})
	table.Append([]string{"Fruit", "Banana"})
	table.Render()
	expected := `<svg xmlns="http://www.w3.org/2000/svg"width="81"height="58"font-family="Arial"font-size="10"><style>text { stroke: none; }</style><rect x="1"y="1"width="36"height="18"fill="#DDD"/><text x="19"y="10"fill="#000"text-anchor="middle"dominant-baseline="middle">CAT</text><rect x="38"y="1"width="42"height="18"fill="#DDD"/><text x="59"y="10"fill="#000"text-anchor="middle"dominant-baseline="middle">ITEM</text><rect x="1"y="20"width="36"height="37"fill="#FFF"/><text x="4"y="38.5"fill="#000"text-anchor="start"dominant-baseline="middle">Fruit</text><rect x="38"y="20"width="42"height="18"fill="#FFF"/><text x="41"y="29"fill="#000"text-anchor="start"dominant-baseline="middle">Apple</text><rect x="38"y="39"width="42"height="18"fill="#EEE"/><text x="41"y="48"fill="#000"text-anchor="start"dominant-baseline="middle">Banana</text><g class="table-borders"stroke="black"stroke-width="1"stroke-linecap="square"><line x1="0.5"y1="0.5"x2="80.5"y2="0.5"/><line x1="0.5"y1="19.5"x2="80.5"y2="19.5"/><line x1="0.5"y1="38.5"x2="80.5"y2="38.5"/><line x1="0.5"y1="57.5"x2="80.5"y2="57.5"/><line x1="0.5"y1="0.5"x2="0.5"y2="57.5"/><line x1="37.5"y1="0.5"x2="37.5"y2="57.5"/><line x1="80.5"y1="0.5"x2="80.5"y2="57.5"/></g></svg>`
	if !visualCheckSVG(t, "SVGVerticalMerge", buf.String(), expected) {
		t.Log(table.Debug())
	}
}

// TestSVGWithFooterAndAlignment tests SVG with footer and alignments.
// Parameter t is the testing context.
// No return value; skips test as experimental.
func TestSVGWithFooterAndAlignment(t *testing.T) {
	var buf bytes.Buffer
	svgCfg := defaultSVGConfigForTests(false)
	table := tablewriter.NewTable(&buf,
		tablewriter.WithRenderer(renderer.NewSVG(svgCfg)),
		tablewriter.WithHeaderConfig(tw.CellConfig{
			Formatting: tw.CellFormatting{AutoFormat: tw.On},
			Alignment:  tw.CellAlignment{Global: tw.AlignCenter},
		}),
		tablewriter.WithRowConfig(tw.CellConfig{
			Alignment: tw.CellAlignment{PerColumn: []tw.Align{tw.AlignLeft, tw.AlignRight, tw.AlignCenter}},
		}),
		tablewriter.WithFooterConfig(tw.CellConfig{
			Alignment: tw.CellAlignment{Global: tw.AlignRight},
		}),
	)
	table.Header([]string{"Item", "Qty", "Price"})
	table.Append([]string{"Apple", "5", "1.20"})
	table.Append([]string{"Banana", "12", "0.35"})
	table.Footer([]string{"", "Total", "7.20"})
	table.Render()
	expected := `<svg xmlns="http://www.w3.org/2000/svg"width="118"height="77"font-family="Arial"font-size="10"><style>text { stroke: none; }</style><rect x="1"y="1"width="42"height="18"fill="#DDD"/><text x="22"y="10"fill="#000"text-anchor="middle"dominant-baseline="middle">ITEM</text><rect x="44"y="1"width="36"height="18"fill="#DDD"/><text x="62"y="10"fill="#000"text-anchor="middle"dominant-baseline="middle">QTY</text><rect x="81"y="1"width="36"height="18"fill="#DDD"/><text x="99"y="10"fill="#000"text-anchor="middle"dominant-baseline="middle">PRICE</text><rect x="1"y="20"width="42"height="18"fill="#FFF"/><text x="4"y="29"fill="#000"text-anchor="start"dominant-baseline="middle">Apple</text><rect x="44"y="20"width="36"height="18"fill="#FFF"/><text x="77"y="29"fill="#000"text-anchor="end"dominant-baseline="middle">5</text><rect x="81"y="20"width="36"height="18"fill="#FFF"/><text x="99"y="29"fill="#000"text-anchor="middle"dominant-baseline="middle">1.20</text><rect x="1"y="39"width="42"height="18"fill="#EEE"/><text x="4"y="48"fill="#000"text-anchor="start"dominant-baseline="middle">Banana</text><rect x="44"y="39"width="36"height="18"fill="#EEE"/><text x="77"y="48"fill="#000"text-anchor="end"dominant-baseline="middle">12</text><rect x="81"y="39"width="36"height="18"fill="#EEE"/><text x="99"y="48"fill="#000"text-anchor="middle"dominant-baseline="middle">0.35</text><rect x="1"y="58"width="42"height="18"fill="#DDD"/><text x="40" y="67"fill="#000"text-anchor="end"dominant-baseline="middle"></text><rect x="44"y="58"width="36"height="18"fill="#DDD"/><text x="77"y="67"fill="#000"text-anchor="end"dominant-baseline="middle">Total</text><rect x="81"y="58"width="36"height="18"fill="#DDD"/><text x="114"y="67"fill="#000"text-anchor="end"dominant-baseline="middle">7.20</text><g class="table-borders"stroke="black"stroke-width="1"stroke-linecap="square"><line x1="0.5"y1="0.5"x2="117.5"y2="0.5"/><line x1="0.5"y1="19.5"x2="117.5"y2="19.5"/><line x1="0.5"y1="38.5"x2="117.5"y2="38.5"/><line x1="0.5"y1="57.5"x2="117.5"y2="57.5"/><line x1="0.5"y1="76.5"x2="117.5"y2="76.5"/><line x1="0.5"y1="0.5"x2="0.5"y2="76.5"/><line x1="43.5"y1="0.5"x2="43.5"y2="76.5"/><line x1="80.5"y1="0.5"x2="80.5"y2="76.5"/><line x1="117.5"y1="0.5"x2="117.5"y2="76.5"/></g></svg>`
	if !visualCheckSVG(t, "SVGWithFooterAndAlignment", buf.String(), expected) {
		t.Log("--- Debug Log for SVGWithFooterAndAlignment ---")
		t.Log(table.Debug())
	}
}

// TestSVGColumnAlignmentOverride tests SVG with column alignment overrides.
// Parameter t is the testing context.
// No return value;
func TestSVGColumnAlignmentOverride(t *testing.T) {
	var buf bytes.Buffer
	svgCfg := defaultSVGConfigForTests(false)
	table := tablewriter.NewTable(&buf,
		tablewriter.WithRenderer(renderer.NewSVG(svgCfg)),
		tablewriter.WithHeaderConfig(tw.CellConfig{Alignment: tw.CellAlignment{Global: tw.AlignCenter}, Formatting: tw.CellFormatting{AutoFormat: tw.Off}}),
		tablewriter.WithRowConfig(tw.CellConfig{
			Alignment: tw.CellAlignment{PerColumn: []tw.Align{tw.AlignRight, tw.AlignCenter, tw.AlignLeft}},
		}),
	)
	table.Header([]string{"H1", "H2", "H3"})
	table.Append([]string{"R1C1", "R1C2", "R1C3"})
	table.Render()
	expected := `<svg xmlns="http://www.w3.org/2000/svg"width="94"height="39"font-family="Arial"font-size="10"><style>text { stroke: none; }</style><rect x="1"y="1"width="30"height="18"fill="#DDD"/><text x="16"y="10"fill="#000"text-anchor="middle"dominant-baseline="middle">H1</text><rect x="32"y="1"width="30"height="18"fill="#DDD"/><text x="47"y="10"fill="#000"text-anchor="middle"dominant-baseline="middle">H2</text><rect x="63"y="1"width="30"height="18"fill="#DDD"/><text x="78"y="10"fill="#000"text-anchor="middle"dominant-baseline="middle">H3</text><rect x="1"y="20"width="30"height="18"fill="#FFF"/><text x="28"y="29"fill="#000"text-anchor="end"dominant-baseline="middle">R1C1</text><rect x="32"y="20"width="30"height="18"fill="#FFF"/><text x="47"y="29"fill="#000"text-anchor="middle"dominant-baseline="middle">R1C2</text><rect x="63"y="20"width="30"height="18"fill="#FFF"/><text x="66"y="29"fill="#000"text-anchor="start"dominant-baseline="middle">R1C3</text><g class="table-borders"stroke="black"stroke-width="1"stroke-linecap="square"><line x1="0.5"y1="0.5"x2="93.5"y2="0.5"/><line x1="0.5"y1="19.5"x2="93.5"y2="19.5"/><line x1="0.5"y1="38.5"x2="93.5"y2="38.5"/><line x1="0.5"y1="0.5"x2="0.5"y2="38.5"/><line x1="31.5"y1="0.5"x2="31.5"y2="38.5"/><line x1="62.5"y1="0.5"x2="62.5"y2="38.5"/><line x1="93.5"y1="0.5"x2="93.5"y2="38.5"/></g></svg>`
	if !visualCheckSVG(t, "SVGColumnAlignmentOverride", buf.String(), expected) {
		t.Log(table.Debug())
	}
}

// TestSVGHorizontalMerge tests SVG with horizontal merging.
// Parameter t is the testing context.
// No return value;
func TestSVGHorizontalMerge(t *testing.T) {
	var buf bytes.Buffer
	svgCfg := defaultSVGConfigForTests(false)
	table := tablewriter.NewTable(&buf,
		tablewriter.WithRenderer(renderer.NewSVG(svgCfg)),
		tablewriter.WithConfig(tablewriter.Config{
			Header: tw.CellConfig{
				Merging:   tw.CellMerging{Mode: tw.MergeHorizontal},
				Alignment: tw.CellAlignment{Global: tw.AlignCenter},
			},
			Row: tw.CellConfig{Merging: tw.CellMerging{Mode: tw.MergeHorizontal}, Alignment: tw.CellAlignment{Global: tw.AlignLeft}},
		}),
	)
	table.Header([]string{"A", "Merged Header", "Merged Header"})
	table.Append([]string{"Data 1", "Data 2", "Data 2"})
	table.Render()
	expected := `<svg xmlns="http://www.w3.org/2000/svg" width="135" height="39" font-family="Arial" font-size="10"><style>text { stroke: none; }</style><rect x="1" y="1" width="42" height="18" fill="#DDD"/><text x="22" y="10" fill="#000" text-anchor="middle" dominant-baseline="middle">A</text><rect x="44" y="1" width="91" height="18" fill="#DDD"/><text x="89.5" y="10" fill="#000" text-anchor="middle" dominant-baseline="middle">MERGED HEADER</text><rect x="1" y="20" width="42" height="18" fill="#FFF"/><text x="4" y="29" fill="#000" text-anchor="start" dominant-baseline="middle">Data 1</text><rect x="44" y="20" width="91" height="18" fill="#FFF"/><text x="47" y="29" fill="#000" text-anchor="start" dominant-baseline="middle">Data 2</text><g class="table-borders" stroke="black" stroke-width="1" stroke-linecap="square"><line x1="0.5" y1="0.5" x2="134.5" y2="0.5"/><line x1="0.5" y1="19.5" x2="134.5" y2="19.5"/><line x1="0.5" y1="38.5" x2="134.5" y2="38.5"/><line x1="0.5" y1="0.5" x2="0.5" y2="38.5"/><line x1="43.5" y1="0.5" x2="43.5" y2="38.5"/><line x1="134.5" y1="0.5" x2="134.5" y2="38.5"/></g></svg>`
	if !visualCheckSVG(t, "SVGHorizontalMerge", buf.String(), expected) {
		t.Log("--- Debug Log for SVGHorizontalMerge ---")
		t.Log(table.Debug())
	}
}
