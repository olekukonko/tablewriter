package tablewriter

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/olekukonko/tablewriter/renderer"
	"github.com/olekukonko/tablewriter/symbols"
	"strings"
	"testing"
)

// visualCheck compares rendered output against exact expected lines
func visualCheck(t *testing.T, name string, output string, expected string) {
	t.Helper()

	// Normalize line endings and split into lines
	normalize := func(s string) []string {
		s = strings.ReplaceAll(s, "\r\n", "\n")
		return strings.Split(s, "\n")
	}

	expectedLines := normalize(expected)
	outputLines := normalize(output)

	// Trim empty lines from start and end
	trimEmpty := func(lines []string) []string {
		start, end := 0, len(lines)
		for start < end && strings.TrimSpace(lines[start]) == "" {
			start++
		}
		for end > start && strings.TrimSpace(lines[end-1]) == "" {
			end--
		}
		return lines[start:end]
	}

	expectedLines = trimEmpty(expectedLines)
	outputLines = trimEmpty(outputLines)

	// Compare line counts
	if len(outputLines) != len(expectedLines) {
		t.Errorf("%s: line count mismatch - expected %d, got %d", name, len(expectedLines), len(outputLines))
		t.Errorf("Expected:\n%s\n", strings.Join(expectedLines, "\n"))
		t.Errorf("Got:\n%s\n", strings.Join(outputLines, "\n"))
		return
	}

	// Compare each line
	type mismatch struct {
		Line     int    `json:"line"`
		Expected string `json:"expected"`
		Got      string `json:"got"`
	}
	var mismatches []mismatch

	for i := 0; i < len(expectedLines) && i < len(outputLines); i++ {
		exp := strings.TrimSpace(expectedLines[i])
		got := strings.TrimSpace(outputLines[i])
		if exp != got {
			mismatches = append(mismatches, mismatch{
				Line:     i + 1,
				Expected: fmt.Sprintf("%s (%d)", exp, len(exp)),
				Got:      fmt.Sprintf("%s (%d)", got, len(got)),
			})
		}
	}

	// Report mismatches
	if len(mismatches) > 0 {
		diff, _ := json.MarshalIndent(mismatches, "", "  ")
		t.Errorf("%s: %d mismatches found:\n%s", name, len(mismatches), diff)
		t.Errorf("Full expected output:\n%s", expected)
		t.Errorf("Full actual output:\n%s", output)
	}
}

func TestBasicTableDefault(t *testing.T) {
	var buf bytes.Buffer

	table := NewTable(&buf)
	table.SetHeader([]string{"Name", "Age", "City"})
	table.Append([]string{"Alice", "25", "New York"})
	table.Append([]string{"Bob", "30", "Boston"})
	table.Render()

	expected := `
	┌─────────┬──────┬────────────┐
	│  NAME   │ AGE  │    CITY    │
	├─────────┼──────┼────────────┤
	│ Alice   │ 25   │ New York   │
	│ Bob     │ 30   │ Boston     │
	└─────────┴──────┴────────────┘

`
	visualCheck(t, "BasicTableRendering", buf.String(), expected)
}

func TestBasicTableASCII(t *testing.T) {
	var buf bytes.Buffer
	table := NewTable(&buf,
		WithRenderer(renderer.NewDefault(renderer.DefaultConfig{
			Symbols: symbols.NewSymbols(symbols.StyleASCII),
		})),
	)
	table.SetHeader([]string{"Name", "Age", "City"})
	table.Append([]string{"Alice", "25", "New York"})
	table.Append([]string{"Bob", "30", "Boston"})
	table.Render()

	expected := `
	+---------+------+------------+
	|  NAME   | AGE  |    CITY    |
	+---------+------+------------+
	| Alice   | 25   | New York   |
	| Bob     | 30   | Boston     |
	+---------+------+------------+
`
	visualCheck(t, "BasicTableASCII", buf.String(), expected)
}

func TestBasicTableUnicodeRounded(t *testing.T) {
	var buf bytes.Buffer
	table := NewTable(&buf,
		WithRenderer(renderer.NewDefault(renderer.DefaultConfig{
			Symbols: symbols.NewSymbols(symbols.StyleRounded),
		})),
	)
	table.SetHeader([]string{"Name", "Age", "City"})
	table.Append([]string{"Alice", "25", "New York"})
	table.Append([]string{"Bob", "30", "Boston"})
	table.Render()

	expected := `
	╭─────────┬──────┬────────────╮
	│  NAME   │ AGE  │    CITY    │
	├─────────┼──────┼────────────┤
	│ Alice   │ 25   │ New York   │
	│ Bob     │ 30   │ Boston     │
	╰─────────┴──────┴────────────╯
`
	visualCheck(t, "BasicTableUnicodeRounded", buf.String(), expected)
}

func TestBasicTableUnicodeDouble(t *testing.T) {
	var buf bytes.Buffer
	table := NewTable(&buf,
		WithRenderer(renderer.NewDefault(renderer.DefaultConfig{
			Symbols: symbols.NewSymbols(symbols.StyleDouble),
		})),
	)
	table.SetHeader([]string{"Name", "Age", "City"})
	table.Append([]string{"Alice", "25", "New York"})
	table.Append([]string{"Bob", "30", "Boston"})
	table.Render()

	expected := `
	╔═════════╦══════╦════════════╗
	║  NAME   ║ AGE  ║    CITY    ║
	╠═════════╬══════╬════════════╣
	║ Alice   ║ 25   ║ New York   ║
	║ Bob     ║ 30   ║ Boston     ║
	╚═════════╩══════╩════════════╝
`
	visualCheck(t, "TableUnicodeDouble", buf.String(), expected)
}

func TestUnicodeWithoutHeader(t *testing.T) {
	data := [][]string{
		{"Regular", "regular line", "1"},
		{"Thick", "particularly thick line", "2"},
		{"Double", "double line", "3"},
	}

	var buf bytes.Buffer
	table := NewTable(&buf,
		WithRenderer(renderer.NewDefault(renderer.DefaultConfig{
			Borders: renderer.Border{Left: renderer.On, Right: renderer.On, Top: renderer.Off, Bottom: renderer.Off},
		})),
	)
	table.SetHeader([]string{"Name", "Age", "City"})
	table.AppendBulk(data)

	table.Render()

	expected := `

	│   NAME    │            AGE            │ CITY │
	├───────────┼───────────────────────────┼──────┤
	│ Regular   │ regular line              │ 1    │
	│ Thick     │ particularly thick line   │ 2    │
	│ Double    │ double line               │ 3    │

`
	visualCheck(t, "UnicodeWithoutHeader", buf.String(), expected)
}

func TestDisableSeparator(t *testing.T) {
	data := [][]string{
		{"Regular", "regular line", "1"},
		{"Thick", "particularly thick line", "2"},
		{"Double", "double line", "3"},
	}

	t.Run("horizontal - enabled", func(t *testing.T) {
		var buf bytes.Buffer
		table := NewTable(&buf,
			WithRenderer(renderer.NewDefault(renderer.DefaultConfig{
				Settings: renderer.Settings{
					HeaderLine:          renderer.On, // Header separator on
					LineColumnSeparator: renderer.On, // Vertical separators on
					LineSeparator:       renderer.On, // Horizontal separators on
				},
			})),
		)
		table.SetHeader([]string{"Name", "Age", "City"})
		table.AppendBulk(data)
		table.Render()

		expected := `
    ┌───────────┬───────────────────────────┬──────┐
    │   NAME    │            AGE            │ CITY │
    ├───────────┼───────────────────────────┼──────┤
    │ Regular   │ regular line              │ 1    │
    ├───────────┼───────────────────────────┼──────┤
    │ Thick     │ particularly thick line   │ 2    │
    ├───────────┼───────────────────────────┼──────┤
    │ Double    │ double line               │ 3    │
    └───────────┴───────────────────────────┴──────┘
    `
		visualCheck(t, "HorizontalEnabled", buf.String(), expected)
	})

	t.Run("horizontal - disabled", func(t *testing.T) {
		var buf bytes.Buffer
		table := NewTable(&buf,
			WithRenderer(renderer.NewDefault(renderer.DefaultConfig{
				Settings: renderer.Settings{
					HeaderLine:          renderer.On,  // Header separator on
					LineColumnSeparator: renderer.On,  // Vertical separators on
					LineSeparator:       renderer.Off, // Horizontal separators off
				},
			})),
		)
		table.SetHeader([]string{"Name", "Age", "City"})
		table.AppendBulk(data)
		table.Render()

		expected := `
    ┌───────────┬───────────────────────────┬──────┐
    │   NAME    │            AGE            │ CITY │
    ├───────────┼───────────────────────────┼──────┤
    │ Regular   │ regular line              │ 1    │
    │ Thick     │ particularly thick line   │ 2    │
    │ Double    │ double line               │ 3    │
    └───────────┴───────────────────────────┴──────┘
    `
		visualCheck(t, "HorizontalDisabled", buf.String(), expected)
	})

	t.Run("vertical - enabled", func(t *testing.T) {
		var buf bytes.Buffer
		table := NewTable(&buf,
			WithRenderer(renderer.NewDefault(renderer.DefaultConfig{
				Settings: renderer.Settings{
					HeaderLine:          renderer.On,  // Header separator on
					LineColumnSeparator: renderer.On,  // Vertical separators on
					LineSeparator:       renderer.Off, // Horizontal separators off
				},
			})),
		)
		table.SetHeader([]string{"Name", "Age", "City"})
		table.AppendBulk(data)
		table.Render()

		expected := `
    ┌───────────┬───────────────────────────┬──────┐
    │   NAME    │            AGE            │ CITY │
    ├───────────┼───────────────────────────┼──────┤
    │ Regular   │ regular line              │ 1    │
    │ Thick     │ particularly thick line   │ 2    │
    │ Double    │ double line               │ 3    │
    └───────────┴───────────────────────────┴──────┘
    `
		visualCheck(t, "VerticalEnabled", buf.String(), expected)
	})

	t.Run("vertical - disabled", func(t *testing.T) {
		var buf bytes.Buffer
		table := NewTable(&buf,
			WithRenderer(renderer.NewDefault(renderer.DefaultConfig{
				Settings: renderer.Settings{
					HeaderLine:          renderer.On,  // Header separator on
					LineColumnSeparator: renderer.Off, // Vertical separators off
					LineSeparator:       renderer.Off, // Horizontal separators off
				},
			})),
		)
		table.SetHeader([]string{"Name", "Age", "City"})
		table.AppendBulk(data)
		table.Render()

		expected := `
        ┌────────────────────────────────────────────┐
        │   NAME                AGE             CITY │
        ├────────────────────────────────────────────┤
        │ Regular    regular line               1    │
        │ Thick      particularly thick line    2    │
        │ Double     double line                3    │
        └────────────────────────────────────────────┘
    `
		visualCheck(t, "VerticalDisabled", buf.String(), expected)
	})
}
