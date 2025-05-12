package tests

import (
	"bytes"
	"fmt"
	"github.com/olekukonko/tablewriter"
	"github.com/olekukonko/tablewriter/renderer"
	"github.com/olekukonko/tablewriter/tw"
	"testing"
)

func TestBasicTableDefault(t *testing.T) {
	var buf bytes.Buffer

	table := tablewriter.NewTable(&buf)
	table.Header([]string{"Name", "Age", "City"})
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
	debug := visualCheck(t, "BasicTableRendering", buf.String(), expected)
	if !debug {
		t.Error(table.Debug().String())
	}
}

func TestBasicTableDefaultBorder(t *testing.T) {
	var buf bytes.Buffer

	//table := tablewriter.NewTable(&buf)
	//table.Header([]string{"Name", "Age", "City"})
	//table.Append([]string{"Alice", "25", "New York"})
	//table.Append([]string{"Bob", "30", "Boston"})
	//table.Render()

	t.Run("all-off", func(t *testing.T) {
		buf.Reset()
		table := tablewriter.NewTable(&buf,
			tablewriter.WithRenderer(renderer.NewBlueprint(tw.Rendition{
				Borders: tw.Border{Left: tw.Off, Right: tw.Off, Top: tw.Off, Bottom: tw.Off},
			})),
		)

		table.Header([]string{"Name", "Age", "City"})
		table.Append([]string{"Alice", "25", "New York"})
		table.Append([]string{"Bob", "30", "Boston"})
		table.Render()

		expected := `
         NAME  │ AGE │   CITY   
        ───────┼─────┼──────────
         Alice │ 25  │ New York 
         Bob   │ 30  │ Boston   
`

		visualCheck(t, "BasicTableRendering-all-off", buf.String(), expected)

	})

	t.Run("top-on", func(t *testing.T) {
		buf.Reset()
		table := tablewriter.NewTable(&buf,
			tablewriter.WithRenderer(renderer.NewBlueprint(tw.Rendition{
				Borders: tw.Border{Left: tw.Off, Right: tw.Off, Top: tw.On, Bottom: tw.Off},
			})),
		)

		table.Header([]string{"Name", "Age", "City"})
		table.Append([]string{"Alice", "25", "New York"})
		table.Append([]string{"Bob", "30", "Boston"})
		table.Render()

		expected := `
        ───────┬─────┬──────────
         NAME  │ AGE │   CITY   
        ───────┼─────┼──────────
         Alice │ 25  │ New York 
         Bob   │ 30  │ Boston  

`

		visualCheck(t, "BasicTableRendering-top-on", buf.String(), expected)
	})

	t.Run("mix", func(t *testing.T) {
		buf.Reset()
		table := tablewriter.NewTable(&buf,
			tablewriter.WithRenderer(renderer.NewBlueprint(tw.Rendition{
				Borders: tw.Border{Left: tw.Off, Right: tw.On, Top: tw.On, Bottom: tw.On},
			})),
		)

		table.Header([]string{"Name", "Age", "City"})
		table.Append([]string{"Alice", "25", "New York"})
		table.Append([]string{"Bob", "30", "Boston"})
		table.Render()

		expected := `
        ───────┬─────┬──────────┐
         NAME  │ AGE │   CITY   │
        ───────┼─────┼──────────┤
         Alice │ 25  │ New York │
         Bob   │ 30  │ Boston   │
        ───────┴─────┴──────────┘

`
		visualCheck(t, "BasicTableRendering-mix", buf.String(), expected)

	})
}

func TestUnicodeWithoutHeader(t *testing.T) {
	data := [][]string{
		{"Regular", "regular line", "1"},
		{"Thick", "particularly thick line", "2"},
		{"Double", "double line", "3"},
	}

	var buf bytes.Buffer
	table := tablewriter.NewTable(&buf,
		tablewriter.WithRenderer(renderer.NewBlueprint(tw.Rendition{
			Borders: tw.Border{Left: tw.On, Right: tw.On, Top: tw.Off, Bottom: tw.Off},
		})),
	)
	table.Header([]string{"Name", "Age", "City"})
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

func TestBasicTableASCII(t *testing.T) {
	var buf bytes.Buffer
	table := tablewriter.NewTable(&buf,
		tablewriter.WithRenderer(renderer.NewBlueprint(tw.Rendition{
			Symbols: tw.NewSymbols(tw.StyleASCII),
		})),
	)
	table.Header([]string{"Name", "Age", "City"})
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
	table := tablewriter.NewTable(&buf,
		tablewriter.WithRenderer(renderer.NewBlueprint(tw.Rendition{
			Symbols: tw.NewSymbols(tw.StyleRounded),
		})),
	)
	table.Header([]string{"Name", "Age", "City"})
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
	table := tablewriter.NewTable(&buf,
		tablewriter.WithRenderer(renderer.NewBlueprint(tw.Rendition{
			Symbols: tw.NewSymbols(tw.StyleDouble),
		})),
	)
	table.Header([]string{"Name", "Age", "City"})
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

func TestSeparator(t *testing.T) {
	data := [][]string{
		{"Regular", "regular line", "1"},
		{"Thick", "particularly thick line", "2"},
		{"Double", "double line", "3"},
	}

	t.Run("horizontal - enabled", func(t *testing.T) {
		var buf bytes.Buffer
		table := tablewriter.NewTable(&buf,
			tablewriter.WithRenderer(renderer.NewBlueprint(tw.Rendition{
				Settings: tw.Settings{
					Separators: tw.Separators{
						BetweenColumns: tw.On,
						BetweenRows:    tw.On,
					},
					Lines: tw.Lines{
						ShowHeaderLine: tw.On,
					},
				},
			})),
		)
		table.Header([]string{"Name", "Age", "City"})
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
		table := tablewriter.NewTable(&buf,
			tablewriter.WithRenderer(renderer.NewBlueprint(tw.Rendition{
				Settings: tw.Settings{
					Separators: tw.Separators{
						BetweenColumns: tw.On,
						BetweenRows:    tw.Off,
					},
					Lines: tw.Lines{
						ShowHeaderLine: tw.On,
					},
				},
			})),
		)
		table.Header([]string{"Name", "Age", "City"})
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
		visualCheck(t, "Separator", buf.String(), expected)
	})

	t.Run("vertical - enabled", func(t *testing.T) {
		var buf bytes.Buffer
		table := tablewriter.NewTable(&buf,
			tablewriter.WithRenderer(renderer.NewBlueprint(tw.Rendition{
				Settings: tw.Settings{
					Separators: tw.Separators{
						BetweenColumns: tw.On,
						BetweenRows:    tw.Off,
					},
					Lines: tw.Lines{
						ShowHeaderLine: tw.On,
					},
				},
			})),
		)
		table.Header([]string{"Name", "Age", "City"})
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
		table := tablewriter.NewTable(&buf,
			tablewriter.WithRenderer(renderer.NewBlueprint(tw.Rendition{
				Settings: tw.Settings{
					Separators: tw.Separators{
						BetweenColumns: tw.Off,
						BetweenRows:    tw.Off,
					},
					Lines: tw.Lines{
						ShowHeaderLine: tw.On,
					},
				},
			})),
		)
		table.Header([]string{"Name", "Age", "City"})
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
		c := tablewriter.Config{
			MaxWidth: 30,
			Header: tw.CellConfig{Formatting: tw.CellFormatting{
				AutoWrap: tw.WrapTruncate,
			}},
			Debug: false,
		}
		buf.Reset()
		table := tablewriter.NewTable(&buf, tablewriter.WithConfig(c), tablewriter.WithDebug(false))
		table.Header([]string{"Name", "Age", "This is a very long header, let see if this will be properly wrapped"})
		table.Append([]string{"Alice", "25", "New York"})
		table.Append([]string{"Bob", "30", "Boston"})
		table.Render()

		expected := `
            ┌───────┬─────┬─────────────────────────────┐
            │ NAME  │ AGE │ THIS IS A VERY LONG HEADER… │
            ├───────┼─────┼─────────────────────────────┤
            │ Alice │ 25  │ New York                    │
            │ Bob   │ 30  │ Boston                      │
            └───────┴─────┴─────────────────────────────┘
`
		if !visualCheck(t, "BasicTableRendering", buf.String(), expected) {
			t.Log(table.Debug())
		}

	})

	t.Run("long-headers-no-truncate", func(t *testing.T) {
		buf.Reset()

		c := tablewriter.Config{
			MaxWidth: 30,
			Header: tw.CellConfig{Formatting: tw.CellFormatting{
				AutoWrap: tw.WrapNormal,
			}},
		}

		table := tablewriter.NewTable(&buf, tablewriter.WithConfig(c))
		table.Header([]string{"Name", "Age", "This is a very long header, let see if this will be properly wrapped"})
		table.Append([]string{"Alice", "25", "New York"})
		table.Append([]string{"Bob", "30", "Boston"})
		table.Render()
		expected := `
        ┌───────┬─────┬────────────────────────────┐
        │ NAME  │ AGE │ THIS IS A VERY LONG HEADER │
        │       │     │ , LET SEE IF THIS WILL BE  │
        │       │     │      PROPERLY WRAPPED      │
        ├───────┼─────┼────────────────────────────┤
        │ Alice │ 25  │ New York                   │
        │ Bob   │ 30  │ Boston                     │
        └───────┴─────┴────────────────────────────┘
`
		if !visualCheck(t, "LongHeaders", buf.String(), expected) {
			t.Log(table.Debug())
		}

	})
}

func TestLongValues(t *testing.T) {
	data := [][]string{
		{"1", "Learn East has computers with adapted keyboards with enlarged print etc", "Some Data", "Another Data"},
		{"2", "Instead of lining up the letters all", "the way across, he splits the keyboard in two", "Like most ergonomic keyboards"},
		{"3", "Nice", "Lorem Ipsum is simply dummy text of the printing and typesetting industry. Lorem Ipsum has been the industry's \n" +
			"standard dummy text ever since the 1500s, when an unknown printer took a galley of type and scrambled it to make a type specimen bok", "Like most ergonomic keyboards"},
	}

	c := tablewriter.Config{
		Header: tw.CellConfig{
			Formatting: tw.CellFormatting{
				MaxWidth:   30,
				Alignment:  tw.AlignCenter,
				AutoFormat: true,
			},
		},
		Row: tw.CellConfig{
			Formatting: tw.CellFormatting{
				MaxWidth:  30,
				AutoWrap:  tw.WrapNormal,
				Alignment: tw.AlignLeft,
			},
		},
		Footer: tw.CellConfig{
			Formatting: tw.CellFormatting{
				MaxWidth:  30,
				Alignment: tw.AlignRight,
			},
			ColumnAligns: []tw.Align{tw.Skip, tw.Skip, tw.Skip, tw.AlignLeft},
		},
	}

	var buf bytes.Buffer
	table := tablewriter.NewTable(&buf, tablewriter.WithConfig(c))
	table.Header([]string{"No", "Comments", "Another", ""})
	table.Footer([]string{"", "", "---------->", "<---------"})
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
	if !visualCheck(t, "LongValues", buf.String(), expected) {
		t.Error(table.Debug().String())
	}
}

func TestWrapping(t *testing.T) {
	data := [][]string{
		{"1", "https://github.com/olekukonko/ruta", "routing websocket"},
		{"2", "https://github.com/olekukonko/error", "better error"},
		{"3", "https://github.com/olekukonko/tablewriter", "terminal\ntable"},
	}

	c := tablewriter.Config{
		Header: tw.CellConfig{
			Formatting: tw.CellFormatting{
				Alignment:  tw.AlignCenter,
				AutoFormat: true,
			},
		},
		Row: tw.CellConfig{
			Formatting: tw.CellFormatting{
				MaxWidth:  30,
				AutoWrap:  tw.WrapBreak,
				Alignment: tw.AlignLeft,
			},
		},
		Footer: tw.CellConfig{
			Formatting: tw.CellFormatting{
				MaxWidth:  30,
				Alignment: tw.AlignRight,
			},
		},
	}

	var buf bytes.Buffer
	table := tablewriter.NewTable(&buf, tablewriter.WithConfig(c))
	table.Header([]string{"No", "Package", "Comments"})
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
	visualCheck(t, "Wrapping", buf.String(), expected)
}

func TestTableWithCustomPadding(t *testing.T) {
	data := [][]string{
		{"Regular", "regular line", "1"},
		{"Thick", "particularly thick line", "2"},
		{"Double", "double line", "3"},
	}

	c := tablewriter.Config{
		Header: tw.CellConfig{
			Formatting: tw.CellFormatting{
				Alignment:  tw.AlignCenter,
				AutoFormat: true,
			},
			Padding: tw.CellPadding{
				Global: tw.Padding{Left: " ", Right: " ", Top: "^", Bottom: "^"},
			},
		},
		Row: tw.CellConfig{
			Formatting: tw.CellFormatting{
				Alignment: tw.AlignCenter,
			},
			Padding: tw.CellPadding{
				Global: tw.Padding{Left: "L", Right: "R", Top: "T", Bottom: "B"},
			},
		},
		Footer: tw.CellConfig{
			Formatting: tw.CellFormatting{
				Alignment:  tw.AlignCenter,
				AutoFormat: true,
			},
			Padding: tw.CellPadding{
				Global: tw.Padding{Left: "*", Right: "*", Top: "", Bottom: ""},
			},
		},
	}

	var buf bytes.Buffer
	table := tablewriter.NewTable(&buf, tablewriter.WithConfig(c))
	table.Header([]string{"Name", "Age", "City"})
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
	visualCheck(t, "TableWithCustomPadding", buf.String(), expected)
}

func TestStreamBorders(t *testing.T) {
	data := [][]string{{"A", "B"}, {"C", "D"}}
	widths := map[int]int{0: 3, 1: 3} // Content (1) + padding (1+1) = 3

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
				tablewriter.WithStreaming(tw.StreamConfig{Enable: true, Widths: tw.CellWidth{PerColumn: widths}}),
			)
			err := st.Start()
			if err != nil {
				t.Fatalf("Start failed: %v", err)
			}
			st.Append(data[0])
			st.Append(data[1])
			err = st.Close()
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
