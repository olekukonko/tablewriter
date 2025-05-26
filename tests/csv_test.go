package tests // Use _test package to test as a user

import (
	"bytes"
	"encoding/csv"
	"strings"
	"testing"

	"github.com/olekukonko/tablewriter"
	"github.com/olekukonko/tablewriter/renderer" // For direct renderer use if needed
	"github.com/olekukonko/tablewriter/tw"
)

const csvTestData = `Name,Department,Salary
Alice,Engineering,120000
Bob,Marketing,85000
Charlie,Engineering,135000
Diana,HR,70000
`

func getCSVReaderFromString(data string) *csv.Reader {
	stringReader := strings.NewReader(data)
	return csv.NewReader(stringReader)
}

func TestTable_Configure_Basic(t *testing.T) {
	var buf bytes.Buffer
	csvReader := getCSVReaderFromString(csvTestData)

	table, err := tablewriter.NewCSVReader(&buf, csvReader, true)
	if err != nil {
		t.Fatalf("NewCSVReader failed: %v", err)
	}

	// Check initial default config values (examples)
	if table.Config().Header.Alignment.Global != tw.AlignCenter {
		t.Errorf("Expected initial header alignment to be Center, got %s", table.Config().Header.Alignment.Global)
	}
	if table.Config().Behavior.TrimSpace != tw.On { // Default from defaultConfig()
		t.Errorf("Expected initial TrimSpace to be On, got %s", table.Config().Behavior.TrimSpace)
	}

	table.Configure(func(cfg *tablewriter.Config) {
		cfg.Header.Formatting.Alignment = tw.AlignLeft
		cfg.Row.Formatting.Alignment = tw.AlignRight
		cfg.Behavior.TrimSpace = tw.Off
		cfg.Debug = true // This should enable the logger
	})

	// Check that Table.config was updated
	if table.Config().Header.Formatting.Alignment != tw.AlignLeft {
		t.Errorf("Expected configured header alignment to be Left, got %s", table.Config().Header.Formatting.Alignment)
	}
	if table.Config().Row.Formatting.Alignment != tw.AlignRight {
		t.Errorf("Expected configured row alignment to be Right, got %s", table.Config().Row.Formatting.Alignment)
	}
	if table.Config().Behavior.TrimSpace != tw.Off {
		t.Errorf("Expected configured TrimSpace to be Off, got %s", table.Config().Behavior.TrimSpace)
	}
	if !table.Config().Debug {
		t.Errorf("Expected configured Debug to be true")
	}

	// Render and check output (visual check will confirm alignment and trimming)
	err = table.Render()
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}

	// What visualCheck will see (assuming Blueprint respects CellContext.Align passed from table.config):
	expectedAfterConfigure := `
┌─────────┬─────────────┬────────┐
│ NAME    │ DEPARTMENT  │ SALARY │
├─────────┼─────────────┼────────┤
│   Alice │ Engineering │ 120000 │
│     Bob │   Marketing │  85000 │
│ Charlie │ Engineering │ 135000 │
│   Diana │          HR │  70000 │
└─────────┴─────────────┴────────┘
`
	if !visualCheck(t, "TestTable_Configure_Basic", buf.String(), expectedAfterConfigure) {
		t.Logf("Debug trace from table:\n%s", table.Debug().String())
	}
}

func TestTable_Options_WithRendition_Borderless(t *testing.T) {
	var buf bytes.Buffer
	csvReader := getCSVReaderFromString(csvTestData) // Ensure csvTestData is defined

	table, err := tablewriter.NewCSVReader(&buf, csvReader, true, tablewriter.WithDebug(true))
	if err != nil {
		t.Fatalf("NewCSVReader failed: %v", err)
	}

	// Initially, it should have default borders
	table.Render()
	initialOutputWithBorders := buf.String()
	buf.Reset() // Clear buffer for next render

	if !strings.Contains(initialOutputWithBorders, "┌───") { // Basic check for default border
		t.Errorf("Expected initial render to have borders, but got:\n%s", initialOutputWithBorders)
	}

	// Define a TRULY borderless and line-less rendition
	borderlessRendition := tw.Rendition{
		Borders: tw.Border{ // Explicitly set all borders to Off
			Left:   tw.Off,
			Right:  tw.Off,
			Top:    tw.Off,
			Bottom: tw.Off,
		},
		// Using StyleNone for symbols means no visible characters for borders/lines if they were on.
		// For a "markdown-like but no lines" look, you might use StyleMarkdown and then turn off lines/separators.
		// For true "no visual structure", StyleNone is good.
		Symbols: tw.NewSymbols(tw.StyleNone),
		Settings: tw.Settings{
			Lines: tw.Lines{ // Explicitly set all line drawing to Off
				ShowTop:        tw.Off,
				ShowBottom:     tw.Off,
				ShowHeaderLine: tw.Off,
				ShowFooterLine: tw.Off,
			},
			Separators: tw.Separators{ // Explicitly set all separators to Off
				ShowHeader:     tw.Off,
				ShowFooter:     tw.Off,
				BetweenRows:    tw.Off,
				BetweenColumns: tw.Off,
			},
		},
	}

	table.Options(
		tablewriter.WithRendition(borderlessRendition),
	)

	// Render again
	err = table.Render()
	if err != nil {
		t.Fatalf("Render after WithRendition failed: %v", err)
	}

	// Expected output: Plain text, no borders, no lines, no separators.
	// Content alignment will be default (Header:Center, Row:Left) because
	// Table.config was not modified for alignments in this test.
	expectedOutputBorderless := `
          NAME    DEPARTMENT   SALARY 
         Alice    Engineering  120000 
         Bob      Marketing    85000  
         Charlie  Engineering  135000 
         Diana    HR           70000  
`
	if !visualCheck(t, "TestTable_Options_WithRendition_Borderless", buf.String(), expectedOutputBorderless) {
		t.Logf("Initial output with borders was:\n%s", initialOutputWithBorders) // For context
		t.Logf("Debug trace from table after borderless rendition:\n%s", table.Debug().String())
	}

	// Verify renderer's internal config was changed
	if bp, ok := table.Renderer().(*renderer.Blueprint); ok {
		currentRendererCfg := bp.Config()

		if currentRendererCfg.Borders.Left.Enabled() {
			t.Errorf("Blueprint Borders.Left should be OFF, but is ON")
		}
		if currentRendererCfg.Borders.Right.Enabled() {
			t.Errorf("Blueprint Borders.Right should be OFF, but is ON")
		}
		if currentRendererCfg.Borders.Top.Enabled() {
			t.Errorf("Blueprint Borders.Top should be OFF, but is ON")
		}
		if currentRendererCfg.Borders.Bottom.Enabled() {
			t.Errorf("Blueprint Borders.Bottom should be OFF, but is ON")
		}

		if currentRendererCfg.Settings.Lines.ShowHeaderLine.Enabled() {
			t.Errorf("Blueprint Settings.Lines.ShowHeaderLine should be OFF, but is ON")
		}
		if currentRendererCfg.Settings.Lines.ShowTop.Enabled() {
			t.Errorf("Blueprint Settings.Lines.ShowTop should be OFF, but is ON")
		}
		if currentRendererCfg.Settings.Separators.BetweenColumns.Enabled() {
			t.Errorf("Blueprint Settings.Separators.BetweenColumns should be OFF, but is ON")
		}

		// Check symbols if relevant (StyleNone should have empty symbols)
		if currentRendererCfg.Symbols.Column() != "" {
			t.Errorf("Blueprint Symbols.Column should be empty for StyleNone, got '%s'", currentRendererCfg.Symbols.Column())
		}
	} else {
		t.Logf("Renderer is not *renderer.Blueprint, skipping detailed internal config check. Type is %T", table.Renderer())
	}
}

// Assume csvTestData and getCSVReaderFromString are defined as in previous examples:
const csvTestDataForPartial = `Name,Department,Salary
Alice,Engineering,120000
Bob,Marketing,85000
`

func getCSVReaderFromStringForPartial(data string) *csv.Reader {
	stringReader := strings.NewReader(data)
	return csv.NewReader(stringReader)
}

// Assume csvTestDataForPartial and getCSVReaderFromStringForPartial are defined
const csvTestDataForPartialUpdate = `Name,Department,Salary
Alice,Engineering,120000
Bob,Marketing,85000
`

func getCSVReaderFromStringForPartialUpdate(data string) *csv.Reader {
	stringReader := strings.NewReader(data)
	return csv.NewReader(stringReader)
}

func TestTable_Options_WithRendition_PartialUpdate(t *testing.T) {
	var buf bytes.Buffer
	csvReader := getCSVReaderFromStringForPartialUpdate(csvTestDataForPartialUpdate)

	// 1. Define an explicitly borderless and line-less initial rendition
	initiallyAllOffRendition := tw.Rendition{
		Borders: tw.Border{
			Left:   tw.Off,
			Right:  tw.Off,
			Top:    tw.Off,
			Bottom: tw.Off,
		},
		Symbols: tw.NewSymbols(tw.StyleNone), // StyleNone should render no visible symbols
		Settings: tw.Settings{
			Lines: tw.Lines{
				ShowTop:        tw.Off,
				ShowBottom:     tw.Off,
				ShowHeaderLine: tw.Off,
				ShowFooterLine: tw.Off,
			},
			Separators: tw.Separators{
				ShowHeader:     tw.Off,
				ShowFooter:     tw.Off,
				BetweenRows:    tw.Off,
				BetweenColumns: tw.Off,
			},
		},
	}

	// Create table with this explicitly "all off" rendition
	table, err := tablewriter.NewCSVReader(&buf, csvReader, true,
		tablewriter.WithDebug(true),
		tablewriter.WithRenderer(renderer.NewBlueprint(initiallyAllOffRendition)),
	)
	if err != nil {
		t.Fatalf("NewCSVReader with initial 'all off' rendition failed: %v", err)
	}

	// Render to confirm initial state (should be very plain)
	table.Render()
	outputAfterInitialAllOff := buf.String()
	buf.Reset() // Clear buffer for the next render

	// Check the initial plain output (content only, no borders/lines)
	expectedInitialPlainOutput := `
         NAME   DEPARTMENT   SALARY 
         Alice  Engineering  120000 
         Bob    Marketing    85000  
`
	if !visualCheck(t, "TestTable_Options_WithRendition_PartialUpdate_InitialState", outputAfterInitialAllOff, expectedInitialPlainOutput) {
		t.Errorf("Initial render was not plain as expected.")
		t.Logf("Initial 'all off' output was:\n%s", outputAfterInitialAllOff)
	}

	partialRenditionUpdate := tw.Rendition{
		Borders: tw.Border{Top: tw.On, Bottom: tw.On}, // Left/Right are 0 (unspecified in this struct literal)
		Symbols: tw.NewSymbols(tw.StyleHeavy),
		Settings: tw.Settings{
			Lines: tw.Lines{ShowTop: tw.On, ShowBottom: tw.On}, // Enable drawing of these lines
			// Separators are zero-value, so they will remain Off from 'initiallyAllOffRendition'
		},
	}

	// Apply the partial update using Options
	table.Options(
		tablewriter.WithRendition(partialRenditionUpdate),
	)

	// Render again
	err = table.Render()
	if err != nil {
		t.Fatalf("Render after partial WithRendition failed: %v", err)
	}
	outputAfterPartialUpdate := buf.String()

	expectedOutputPartialBorders := `
        ━━━━━━━━━━━━━━━━━━━━━━━━━━━━
         NAME   DEPARTMENT   SALARY 
         Alice  Engineering  120000 
         Bob    Marketing    85000  
        ━━━━━━━━━━━━━━━━━━━━━━━━━━━━
`

	if !visualCheck(t, "TestTable_Options_WithRendition_PartialUpdate_FinalState", outputAfterPartialUpdate, expectedOutputPartialBorders) {
		t.Logf("Initial 'all off' output was:\n%s", outputAfterInitialAllOff) // For context
		t.Logf("Debug trace from table after partial update:\n%s", table.Debug().String())
	}

	// 3. Verify the renderer's internal configuration reflects the partial update correctly.
	if bp, ok := table.Renderer().(*renderer.Blueprint); ok {
		currentRendererCfg := bp.Config()

		if !currentRendererCfg.Borders.Top.Enabled() {
			t.Errorf("Blueprint Borders.Top should be ON, but is OFF")
		}
		if !currentRendererCfg.Borders.Bottom.Enabled() {
			t.Errorf("Blueprint Borders.Bottom should be ON, but is OFF")
		}
		if currentRendererCfg.Borders.Left.Enabled() {
			t.Errorf("Blueprint Borders.Left should remain OFF, but is ON")
		}
		if currentRendererCfg.Borders.Right.Enabled() {
			t.Errorf("Blueprint Borders.Right should remain OFF, but is ON")
		}

		if currentRendererCfg.Symbols.Row() != "━" { // From StyleHeavy
			t.Errorf("Blueprint Symbols.Row is not '━' (Heavy), got '%s'", currentRendererCfg.Symbols.Row())
		}
		// Column symbol check might be less relevant if BetweenColumns is Off, but good for completeness.
		if currentRendererCfg.Symbols.Column() != "┃" { // From StyleHeavy
			t.Errorf("Blueprint Symbols.Column is not '┃' (Heavy), got '%s'", currentRendererCfg.Symbols.Column())
		}

		// Check Settings.Lines
		if !currentRendererCfg.Settings.Lines.ShowTop.Enabled() {
			t.Errorf("Blueprint Settings.Lines.ShowTop should be ON, but is OFF")
		}
		if !currentRendererCfg.Settings.Lines.ShowBottom.Enabled() {
			t.Errorf("Blueprint Settings.Lines.ShowBottom should be ON, but is OFF")
		}
		if currentRendererCfg.Settings.Lines.ShowHeaderLine.Enabled() {
			t.Errorf("Blueprint Settings.Lines.ShowHeaderLine should remain OFF, but is ON")
		}

		// Check Settings.Separators
		if currentRendererCfg.Settings.Separators.BetweenColumns.Enabled() {
			t.Errorf("Blueprint Settings.Separators.BetweenColumns should remain OFF, but is ON")
		}
	} else {
		t.Logf("Renderer is not *renderer.Blueprint, skipping detailed internal config check. Type is %T", table.Renderer())
	}
}
