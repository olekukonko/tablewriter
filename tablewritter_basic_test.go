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
	┌───────┬─────┬──────────┐
	│ NAME  │ AGE │   CITY   │
	├───────┼─────┼──────────┤
	│ Alice │ 25  │ New York │
	│ Bob   │ 30  │ Boston   │
	└───────┴─────┴──────────┘
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
	+-------+-----+----------+
	| NAME  | AGE |   CITY   |
	+-------+-----+----------+
	| Alice | 25  | New York |
	| Bob   | 30  | Boston   |
	+-------+-----+----------+
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
	╭───────┬─────┬──────────╮
	│ NAME  │ AGE │   CITY   │
	├───────┼─────┼──────────┤
	│ Alice │ 25  │ New York │
	│ Bob   │ 30  │ Boston   │
	╰───────┴─────┴──────────╯
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
	╔═══════╦═════╦══════════╗
	║ NAME  ║ AGE ║   CITY   ║
	╠═══════╬═════╬══════════╣
	║ Alice ║ 25  ║ New York ║
	║ Bob   ║ 30  ║ Boston   ║
	╚═══════╩═════╩══════════╝
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
	table.Bulk(data)

	table.Render()

	expected := `
	│  NAME   │           AGE           │ CITY │
	├─────────┼─────────────────────────┼──────┤
	│ Regular │ regular line            │ 1    │
	│ Thick   │ particularly thick line │ 2    │
	│ Double  │ double line             │ 3    │
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
					Separators: renderer.Separators{
						BetweenColumns: renderer.On,
						BetweenRows:    renderer.On,
					},
					Lines: renderer.Lines{
						ShowHeaderLine: renderer.On,
					},
				},
			})),
		)
		table.SetHeader([]string{"Name", "Age", "City"})
		table.Bulk(data)
		table.Render()

		expected := `
        ┌─────────┬─────────────────────────┬──────┐
        │  NAME   │           AGE           │ CITY │
        ├─────────┼─────────────────────────┼──────┤
        │ Regular │ regular line            │ 1    │
        ├─────────┼─────────────────────────┼──────┤
        │ Thick   │ particularly thick line │ 2    │
        ├─────────┼─────────────────────────┼──────┤
        │ Double  │ double line             │ 3    │
        └─────────┴─────────────────────────┴──────┘
    `
		visualCheck(t, "HorizontalEnabled", buf.String(), expected)
	})

	t.Run("horizontal - disabled", func(t *testing.T) {
		var buf bytes.Buffer
		table := NewTable(&buf,
			WithRenderer(renderer.NewDefault(renderer.DefaultConfig{
				Settings: renderer.Settings{
					Separators: renderer.Separators{
						BetweenColumns: renderer.On,
						BetweenRows:    renderer.Off,
					},
					Lines: renderer.Lines{
						ShowHeaderLine: renderer.On,
					},
				},
			})),
		)
		table.SetHeader([]string{"Name", "Age", "City"})
		table.Bulk(data)
		table.Render()

		expected := `
		┌─────────┬─────────────────────────┬──────┐
		│  NAME   │           AGE           │ CITY │
		├─────────┼─────────────────────────┼──────┤
		│ Regular │ regular line            │ 1    │
		│ Thick   │ particularly thick line │ 2    │
		│ Double  │ double line             │ 3    │
		└─────────┴─────────────────────────┴──────┘
    `
		visualCheck(t, "HorizontalDisabled", buf.String(), expected)
	})

	t.Run("vertical - enabled", func(t *testing.T) {
		var buf bytes.Buffer
		table := NewTable(&buf,
			WithRenderer(renderer.NewDefault(renderer.DefaultConfig{
				Settings: renderer.Settings{
					Separators: renderer.Separators{
						BetweenColumns: renderer.On,
						BetweenRows:    renderer.Off,
					},
					Lines: renderer.Lines{
						ShowHeaderLine: renderer.On,
					},
				},
			})),
		)
		table.SetHeader([]string{"Name", "Age", "City"})
		table.Bulk(data)
		table.Render()

		expected := `
        ┌─────────┬─────────────────────────┬──────┐
        │  NAME   │           AGE           │ CITY │
        ├─────────┼─────────────────────────┼──────┤
        │ Regular │ regular line            │ 1    │
        │ Thick   │ particularly thick line │ 2    │
        │ Double  │ double line             │ 3    │
        └─────────┴─────────────────────────┴──────┘
    `
		visualCheck(t, "VerticalEnabled", buf.String(), expected)
	})

	t.Run("vertical - disabled", func(t *testing.T) {
		var buf bytes.Buffer
		table := NewTable(&buf,
			WithRenderer(renderer.NewDefault(renderer.DefaultConfig{
				Settings: renderer.Settings{
					Separators: renderer.Separators{
						BetweenColumns: renderer.Off,
						BetweenRows:    renderer.Off,
					},
					Lines: renderer.Lines{
						ShowHeaderLine: renderer.On,
					},
				},
			})),
		)
		table.SetHeader([]string{"Name", "Age", "City"})
		table.Bulk(data)
		table.Render()

		expected := `
        ┌────────────────────────────────────────┐
        │  NAME              AGE            CITY │
        ├────────────────────────────────────────┤
        │ Regular  regular line             1    │
        │ Thick    particularly thick line  2    │
        │ Double   double line              3    │
        └────────────────────────────────────────┘
    `
		visualCheck(t, "VerticalDisabled", buf.String(), expected)
	})
}

func TestLongHeaders(t *testing.T) {
	var buf bytes.Buffer

	t.Run("long-headers", func(t *testing.T) {
		c := Config{
			MaxWidth: 30,
			Header: CellConfig{Formatting: CellFormatting{
				AutoWrap: WrapTruncate,
			}},
		}
		buf.Reset()
		table := NewTable(&buf, WithConfig(c))
		table.SetHeader([]string{"Name", "Age", "This is a very long header, let see if this will be properly wrapped"})
		table.Append([]string{"Alice", "25", "New York"})
		table.Append([]string{"Bob", "30", "Boston"})
		table.Render()

		expected := `
		┌───────┬─────┬──────────────────────────────┐
		│ Name  │ Age │ This is a very long header,… │
		├───────┼─────┼──────────────────────────────┤
		│ Alice │ 25  │ New York                     │
		│ Bob   │ 30  │ Boston                       │
		└───────┴─────┴──────────────────────────────┘
`
		visualCheck(t, "BasicTableRendering", buf.String(), expected)
	})

	t.Run("long-headers-no-truncate", func(t *testing.T) {
		buf.Reset()

		c := Config{
			MaxWidth: 30,
			Header: CellConfig{Formatting: CellFormatting{
				AutoWrap: WrapNormal,
			}},
		}

		table := NewTable(&buf, WithConfig(c))
		table.SetHeader([]string{"Name", "Age", "This is a very long header, let see if this will be properly wrapped"})
		table.Append([]string{"Alice", "25", "New York"})
		table.Append([]string{"Bob", "30", "Boston"})
		table.Render()
		expected := `
        ┌───────┬─────┬─────────────────────────────┐
        │ Name  │ Age │ This is a very long header, │
        │       │     │   let see if this will be   │
        │       │     │      properly wrapped       │
        ├───────┼─────┼─────────────────────────────┤
        │ Alice │ 25  │ New York                    │
        │ Bob   │ 30  │ Boston                      │
        └───────┴─────┴─────────────────────────────┘
`
		visualCheck(t, "BasicTableRendering", buf.String(), expected)
	})
}

func TestLongValues(t *testing.T) {
	data := [][]string{
		{"1", "Learn East has computers with adapted keyboards with enlarged print etc", "Some Data", "Another Data"},
		{"2", "Instead of lining up the letters all", "the way across, he splits the keyboard in two", "Like most ergonomic keyboards"},
		{"3", "Nice", "Lorem Ipsum is simply dummy text of the printing and typesetting industry. Lorem Ipsum has been the industry's \n" +
			"standard dummy text ever since the 1500s, when an unknown printer took a galley of type and scrambled it to make a type specimen bok", "Like most ergonomic keyboards"},
	}

	c := Config{
		Header: CellConfig{
			Formatting: CellFormatting{
				MaxWidth:   30,
				Alignment:  renderer.AlignCenter,
				AutoFormat: true,
			},
		},
		Row: CellConfig{
			Formatting: CellFormatting{
				MaxWidth:  30,
				AutoWrap:  WrapNormal,
				Alignment: renderer.AlignLeft,
			},
		},
		Footer: CellConfig{
			Formatting: CellFormatting{
				MaxWidth:  30,
				Alignment: renderer.AlignRight,
			},
			ColumnAligns: []string{"", "", "", renderer.AlignLeft},
		},
	}

	var buf bytes.Buffer
	table := NewTable(&buf, WithConfig(c))
	table.SetHeader([]string{"No", "Comments", "Another", ""})
	table.SetFooter([]string{"", "", "---------->", "<---------"})
	table.Bulk(data)

	table.Render()

	expected := `

	┌────┬─────────────────────────────┬──────────────────────────────┬─────────────────────┐
	│ NO │          COMMENTS           │           ANOTHER            │                     │
	├────┼─────────────────────────────┼──────────────────────────────┼─────────────────────┤
	│ 1  │ Learn East has computers    │ Some Data                    │ Another Data        │
	│    │ with adapted keyboards with │                              │                     │
	│    │ enlarged print etc          │                              │                     │
	│ 2  │ Instead of lining up the    │ the way across, he splits    │ Like most ergonomic │
	│    │ letters all                 │ the keyboard in two          │ keyboards           │
	│ 3  │ Nice                        │ Lorem Ipsum is simply        │ Like most ergonomic │
	│    │                             │ dummy text of the printing   │ keyboards           │
	│    │                             │ and typesetting industry.    │                     │
	│    │                             │ Lorem Ipsum has been the     │                     │
	│    │                             │ industry's                   │                     │
	│    │                             │ standard dummy text ever     │                     │
	│    │                             │ since the 1500s, when an     │                     │
	│    │                             │ unknown printer took a       │                     │
	│    │                             │ galley of type and scrambled │                     │
	│    │                             │ it to make a type specimen   │                     │
	│    │                             │ bok                          │                     │
	├────┼─────────────────────────────┼──────────────────────────────┼─────────────────────┤
	│    │                             │                  ----------> │ <---------          │
	└────┴─────────────────────────────┴──────────────────────────────┴─────────────────────┘

`
	visualCheck(t, "UnicodeWithoutHeader", buf.String(), expected)
}

func TestWrapping(t *testing.T) {
	data := [][]string{
		{"1", "https://github.com/olekukonko/ruta", "routing websocket"},
		{"2", "https://github.com/olekukonko/error", "better error"},
		{"3", "https://github.com/olekukonko/tablewriter", "terminal\ntable"},
	}

	c := Config{
		Header: CellConfig{
			Formatting: CellFormatting{
				Alignment:  renderer.AlignCenter,
				AutoFormat: true,
			},
		},
		Row: CellConfig{
			Formatting: CellFormatting{
				MaxWidth:  30,
				AutoWrap:  WrapBreak,
				Alignment: renderer.AlignLeft,
			},
		},
		Footer: CellConfig{
			Formatting: CellFormatting{
				MaxWidth:  30,
				Alignment: renderer.AlignRight,
			},
		},
	}

	var buf bytes.Buffer
	table := NewTable(&buf, WithConfig(c))
	table.SetHeader([]string{"No", "Package", "Comments"})
	table.Bulk(data)
	table.Render()

	expected := `
        ┌────┬───────────────────────────────┬───────────────────┐
        │ NO │            PACKAGE            │     COMMENTS      │
        ├────┼───────────────────────────────┼───────────────────┤
        │ 1  │ https://github.com/olekukonk↩ │ routing websocket │
        │    │ o/ruta                        │                   │
        │ 2  │ https://github.com/olekukonk↩ │ better error      │
        │    │ o/error                       │                   │
        │ 3  │ https://github.com/olekukonk↩ │ terminal          │
        │    │ o/tablewriter                 │ table             │
        └────┴───────────────────────────────┴───────────────────┘
`
	visualCheck(t, "UnicodeWithoutHeader", buf.String(), expected)
}

func TestTableWithCustomPadding(t *testing.T) {
	data := [][]string{
		{"Regular", "regular line", "1"},
		{"Thick", "particularly thick line", "2"},
		{"Double", "double line", "3"},
	}

	c := Config{
		Header: CellConfig{
			Formatting: CellFormatting{
				Alignment:  renderer.AlignCenter,
				AutoFormat: true,
			},
			Padding: CellPadding{
				Global: symbols.Padding{Left: " ", Right: " ", Top: "^", Bottom: "^"},
			},
		},
		Row: CellConfig{
			Formatting: CellFormatting{
				Alignment: renderer.AlignCenter,
			},
			Padding: CellPadding{
				Global: symbols.Padding{Left: "L", Right: "R", Top: "T", Bottom: "B"},
			},
		},
		Footer: CellConfig{
			Formatting: CellFormatting{
				Alignment:  renderer.AlignCenter,
				AutoFormat: true,
				AutoMerge:  false,
			},
			Padding: CellPadding{
				Global: symbols.Padding{Left: "*", Right: "*", Top: "", Bottom: ""},
			},
		},
	}

	var buf bytes.Buffer
	table := NewTable(&buf, WithConfig(c))
	table.SetHeader([]string{"Name", "Age", "City"})
	table.Bulk(data)
	table.Render()

	expected := `
        ┌─────────┬─────────────────────────┬──────┐
        │ ^^^^^^^ │ ^^^^^^^^^^^^^^^^^^^^^^^ │ ^^^^ │
        │  NAME   │           AGE           │ CITY │
        │ ^^^^^^^ │ ^^^^^^^^^^^^^^^^^^^^^^^ │ ^^^^ │
        ├─────────┼─────────────────────────┼──────┤
        │LTTTTTTTR│LTTTTTTTTTTTTTTTTTTTTTTTR│LTTTTR│
        │LRegularR│LLLLLLregular lineRRRRRRR│LL1RRR│
        │LBBBBBBBR│LBBBBBBBBBBBBBBBBBBBBBBBR│LBBBBR│
        │LTTTTTTTR│LTTTTTTTTTTTTTTTTTTTTTTTR│LTTTTR│
        │LLThickRR│Lparticularly thick lineR│LL2RRR│
        │LBBBBBBBR│LBBBBBBBBBBBBBBBBBBBBBBBR│LBBBBR│
        │LTTTTTTTR│LTTTTTTTTTTTTTTTTTTTTTTTR│LTTTTR│
        │LDoubleRR│LLLLLLLdouble lineRRRRRRR│LL3RRR│
        │LBBBBBBBR│LBBBBBBBBBBBBBBBBBBBBBBBR│LBBBBR│
        └─────────┴─────────────────────────┴──────┘
`
	visualCheck(t, "UnicodeWithoutHeader", buf.String(), expected)
}

func TestFilterMasking(t *testing.T) {
	tests := []struct {
		name     string
		filter   Filter
		data     [][]string
		expected string
	}{
		{
			name:   "MaskEmail",
			filter: MaskEmail,
			data: [][]string{
				{"Alice", "alice@example.com", "25"},
				{"Bob", "bob.test@domain.org", "30"},
			},
			expected: `
        ┌───────┬─────────────────────┬─────┐
        │ NAME  │        EMAIL        │ AGE │
        ├───────┼─────────────────────┼─────┤
        │ Alice │ a****@example.com   │ 25  │
        │ Bob   │ b*******@domain.org │ 30  │
        └───────┴─────────────────────┴─────┘
`,
		},
		{
			name:   "MaskPassword",
			filter: MaskPassword,
			data: [][]string{
				{"Alice", "secretpassword", "25"},
				{"Bob", "pass1234", "30"},
			},
			expected: `
        ┌───────┬────────────────┬─────┐
        │ NAME  │    PASSWORD    │ AGE │
        ├───────┼────────────────┼─────┤
        │ Alice │ ************** │ 25  │
        │ Bob   │ ********       │ 30  │
        └───────┴────────────────┴─────┘
`,
		},
		{
			name:   "MaskCard",
			filter: MaskCard,
			data: [][]string{
				{"Alice", "4111-1111-1111-1111", "25"},
				{"Bob", "5105105105105100", "30"},
			},
			expected: `
        ┌───────┬─────────────────────┬─────┐
        │ NAME  │     CREDIT CARD     │ AGE │
        ├───────┼─────────────────────┼─────┤
        │ Alice │ ****-****-****-1111 │ 25  │
        │ Bob   │ 5105105105105100    │ 30  │
        └───────┴─────────────────────┴─────┘
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			table := NewTable(&buf, WithConfig(Config{
				Header: CellConfig{
					Formatting: CellFormatting{Alignment: renderer.AlignCenter, AutoFormat: true},
					Padding:    CellPadding{Global: symbols.Padding{Left: " ", Right: " "}},
				},
				Row: CellConfig{
					Formatting: CellFormatting{Alignment: renderer.AlignLeft},
					Padding:    CellPadding{Global: symbols.Padding{Left: " ", Right: " "}},
					Filter:     tt.filter,
				},
			}))
			header := []string{"Name", tt.name, "Age"}
			if tt.name == "MaskEmail" {
				header[1] = "Email"
			} else if tt.name == "MaskPassword" {
				header[1] = "Password"
			} else if tt.name == "MaskCard" {
				header[1] = "Credit Card"
			}
			table.SetHeader(header)
			table.Bulk(tt.data)
			table.Render()
			visualCheck(t, tt.name, buf.String(), tt.expected)
		})
	}
}

// Filter Presets
func MaskEmail(cells []string) []string {
	for i, cell := range cells {
		if strings.Contains(cell, "@") {
			parts := strings.Split(cell, "@")
			if len(parts) == 2 {
				masked := parts[0][:1] + strings.Repeat("*", len(parts[0])-1) + "@" + parts[1]
				cells[i] = masked
			}
		}
	}
	return cells
}

func MaskPassword(cells []string) []string {
	for i, cell := range cells {
		if len(cell) > 0 && (strings.Contains(strings.ToLower(cell), "pass") || len(cell) >= 8) {
			cells[i] = strings.Repeat("*", len(cell))
		}
	}
	return cells
}

func MaskCard(cells []string) []string {
	for i, cell := range cells {
		// Simple check for card-like numbers (16 digits or with dashes)
		if len(cell) >= 12 && (strings.Contains(cell, "-") || len(strings.ReplaceAll(cell, " ", "")) >= 12) {
			parts := strings.FieldsFunc(cell, func(r rune) bool { return r == '-' || r == ' ' })
			masked := ""
			for j, part := range parts {
				if j < len(parts)-1 {
					masked += strings.Repeat("*", len(part))
				} else {
					masked += part // Keep last 4 digits visible
				}
				if j < len(parts)-1 {
					masked += "-"
				}
			}
			cells[i] = masked
		}
	}
	return cells
}
