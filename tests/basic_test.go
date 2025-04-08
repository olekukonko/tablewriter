package tests

import (
	"bytes"
	"github.com/olekukonko/tablewriter"
	"github.com/olekukonko/tablewriter/renderer"
	"github.com/olekukonko/tablewriter/tw"
	"testing"
)

func TestBasicTableDefault(t *testing.T) {
	var buf bytes.Buffer

	table := tablewriter.NewTable(&buf)
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
	debug := visualCheck(t, "BasicTableRendering", buf.String(), expected)
	if !debug {
		for _, v := range table.Debug() {
			t.Error(v)
		}
	}
}

func TestBasicTableDefaultBorder(t *testing.T) {
	var buf bytes.Buffer

	//table := tablewriter.NewTable(&buf)
	//table.SetHeader([]string{"Name", "Age", "City"})
	//table.Append([]string{"Alice", "25", "New York"})
	//table.Append([]string{"Bob", "30", "Boston"})
	//table.Render()

	t.Run("all-off", func(t *testing.T) {
		buf.Reset()
		table := tablewriter.NewTable(&buf,
			tablewriter.WithRenderer(renderer.NewDefault(renderer.DefaultConfig{
				Borders: renderer.Border{Left: tw.Off, Right: tw.Off, Top: tw.Off, Bottom: tw.Off},
			})),
		)

		table.SetHeader([]string{"Name", "Age", "City"})
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
			tablewriter.WithRenderer(renderer.NewDefault(renderer.DefaultConfig{
				Borders: renderer.Border{Left: tw.Off, Right: tw.Off, Top: tw.On, Bottom: tw.Off},
			})),
		)

		table.SetHeader([]string{"Name", "Age", "City"})
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
			tablewriter.WithRenderer(renderer.NewDefault(renderer.DefaultConfig{
				Borders: renderer.Border{Left: tw.Off, Right: tw.On, Top: tw.On, Bottom: tw.On},
			})),
		)

		table.SetHeader([]string{"Name", "Age", "City"})
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
		tablewriter.WithRenderer(renderer.NewDefault(renderer.DefaultConfig{
			Borders: renderer.Border{Left: tw.On, Right: tw.On, Top: tw.Off, Bottom: tw.Off},
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

//func TestUnicodeTableDefault(t *testing.T) {
//	var buf bytes.Buffer
//
//	table := tablewriter.NewTable(&buf)
//	table.SetHeader([]string{"Name", "Age", "City"})
//	table.Append([]string{"Alice", "25", "New York"})
//	table.Append([]string{"Bøb", "30", "Tōkyō"})    // Contains ø and ō
//	table.Append([]string{"José", "28", "México"}) // Contains é and accented e (e + combining acute)
//	table.Append([]string{"张三", "35", "北京"})        // Chinese characters
//	table.Append([]string{"अनु", "40", "मुंबई"})    // Devanagari script
//	table.Render()
//
//	expected := `
//	┌───────┬─────┬──────────┐
//	│ NAME  │ AGE │   CITY   │
//	├───────┼─────┼──────────┤
//	│ Alice │ 25  │ New York │
//	│ Bøb   │ 30  │ Tōkyō    │
//	│ José  │ 28  │ México   │
//	│ 张三   │ 35  │ 北京     │
//	│ अनु    │ 40  │ मुंबई      │
//	└───────┴─────┴──────────┘
//`
//	visualCheck(t, "UnicodeTableRendering", buf.String(), expected)
//}

func TestBasicTableASCII(t *testing.T) {
	var buf bytes.Buffer
	table := tablewriter.NewTable(&buf,
		tablewriter.WithRenderer(renderer.NewDefault(renderer.DefaultConfig{
			Symbols: tw.NewSymbols(tw.StyleASCII),
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
	table := tablewriter.NewTable(&buf,
		tablewriter.WithRenderer(renderer.NewDefault(renderer.DefaultConfig{
			Symbols: tw.NewSymbols(tw.StyleRounded),
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
	table := tablewriter.NewTable(&buf,
		tablewriter.WithRenderer(renderer.NewDefault(renderer.DefaultConfig{
			Symbols: tw.NewSymbols(tw.StyleDouble),
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

func TestSeparator(t *testing.T) {
	data := [][]string{
		{"Regular", "regular line", "1"},
		{"Thick", "particularly thick line", "2"},
		{"Double", "double line", "3"},
	}

	t.Run("horizontal - enabled", func(t *testing.T) {
		var buf bytes.Buffer
		table := tablewriter.NewTable(&buf,
			tablewriter.WithRenderer(renderer.NewDefault(renderer.DefaultConfig{
				Settings: renderer.Settings{
					Separators: renderer.Separators{
						BetweenColumns: tw.On,
						BetweenRows:    tw.On,
					},
					Lines: renderer.Lines{
						ShowHeaderLine: tw.On,
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
		table := tablewriter.NewTable(&buf,
			tablewriter.WithRenderer(renderer.NewDefault(renderer.DefaultConfig{
				Settings: renderer.Settings{
					Separators: renderer.Separators{
						BetweenColumns: tw.On,
						BetweenRows:    tw.Off,
					},
					Lines: renderer.Lines{
						ShowHeaderLine: tw.On,
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
		visualCheck(t, "Separator", buf.String(), expected)
	})

	t.Run("vertical - enabled", func(t *testing.T) {
		var buf bytes.Buffer
		table := tablewriter.NewTable(&buf,
			tablewriter.WithRenderer(renderer.NewDefault(renderer.DefaultConfig{
				Settings: renderer.Settings{
					Separators: renderer.Separators{
						BetweenColumns: tw.On,
						BetweenRows:    tw.Off,
					},
					Lines: renderer.Lines{
						ShowHeaderLine: tw.On,
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
		table := tablewriter.NewTable(&buf,
			tablewriter.WithRenderer(renderer.NewDefault(renderer.DefaultConfig{
				Settings: renderer.Settings{
					Separators: renderer.Separators{
						BetweenColumns: tw.Off,
						BetweenRows:    tw.Off,
					},
					Lines: renderer.Lines{
						ShowHeaderLine: tw.On,
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
		c := tablewriter.Config{
			MaxWidth: 30,
			Header: tablewriter.CellConfig{Formatting: tablewriter.CellFormatting{
				AutoWrap: tw.WrapTruncate,
			}},
		}
		buf.Reset()
		table := tablewriter.NewTable(&buf, tablewriter.WithConfig(c))
		table.SetHeader([]string{"Name", "Age", "This is a very long header, let see if this will be properly wrapped"})
		table.Append([]string{"Alice", "25", "New York"})
		table.Append([]string{"Bob", "30", "Boston"})
		table.Render()

		expected := `
        ┌───────┬─────┬──────────────────────────────┐
        │ NAME  │ AGE │ THIS IS A VERY LONG HEADER … │
        ├───────┼─────┼──────────────────────────────┤
        │ Alice │ 25  │ New York                     │
        │ Bob   │ 30  │ Boston                       │
        └───────┴─────┴──────────────────────────────┘
`
		visualCheck(t, "BasicTableRendering", buf.String(), expected)
	})

	t.Run("long-headers-no-truncate", func(t *testing.T) {
		buf.Reset()

		c := tablewriter.Config{
			MaxWidth: 30,
			Header: tablewriter.CellConfig{Formatting: tablewriter.CellFormatting{
				AutoWrap: tw.WrapNormal,
			}},
		}

		table := tablewriter.NewTable(&buf, tablewriter.WithConfig(c))
		table.SetHeader([]string{"Name", "Age", "This is a very long header, let see if this will be properly wrapped"})
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
		visualCheck(t, "LongHeaders", buf.String(), expected)
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
		Header: tablewriter.CellConfig{
			Formatting: tablewriter.CellFormatting{
				MaxWidth:   30,
				Alignment:  tw.AlignCenter,
				AutoFormat: true,
			},
		},
		Row: tablewriter.CellConfig{
			Formatting: tablewriter.CellFormatting{
				MaxWidth:  30,
				AutoWrap:  tw.WrapNormal,
				Alignment: tw.AlignLeft,
			},
		},
		Footer: tablewriter.CellConfig{
			Formatting: tablewriter.CellFormatting{
				MaxWidth:  30,
				Alignment: tw.AlignRight,
			},
			ColumnAligns: []tw.Align{tw.Skip, tw.Skip, tw.Skip, tw.AlignLeft},
		},
	}

	var buf bytes.Buffer
	table := tablewriter.NewTable(&buf, tablewriter.WithConfig(c))
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
	if !visualCheck(t, "LongValues", buf.String(), expected) {
		for _, v := range table.Debug() {
			t.Error(v)
		}
	}
}

func TestWrapping(t *testing.T) {
	data := [][]string{
		{"1", "https://github.com/olekukonko/ruta", "routing websocket"},
		{"2", "https://github.com/olekukonko/error", "better error"},
		{"3", "https://github.com/olekukonko/tablewriter", "terminal\ntable"},
	}

	c := tablewriter.Config{
		Header: tablewriter.CellConfig{
			Formatting: tablewriter.CellFormatting{
				Alignment:  tw.AlignCenter,
				AutoFormat: true,
			},
		},
		Row: tablewriter.CellConfig{
			Formatting: tablewriter.CellFormatting{
				MaxWidth:  30,
				AutoWrap:  tw.WrapBreak,
				Alignment: tw.AlignLeft,
			},
		},
		Footer: tablewriter.CellConfig{
			Formatting: tablewriter.CellFormatting{
				MaxWidth:  30,
				Alignment: tw.AlignRight,
			},
		},
	}

	var buf bytes.Buffer
	table := tablewriter.NewTable(&buf, tablewriter.WithConfig(c))
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
	visualCheck(t, "Wrapping", buf.String(), expected)
}

func TestTableWithCustomPadding(t *testing.T) {
	data := [][]string{
		{"Regular", "regular line", "1"},
		{"Thick", "particularly thick line", "2"},
		{"Double", "double line", "3"},
	}

	c := tablewriter.Config{
		Header: tablewriter.CellConfig{
			Formatting: tablewriter.CellFormatting{
				Alignment:  tw.AlignCenter,
				AutoFormat: true,
			},
			Padding: tablewriter.CellPadding{
				Global: tw.Padding{Left: " ", Right: " ", Top: "^", Bottom: "^"},
			},
		},
		Row: tablewriter.CellConfig{
			Formatting: tablewriter.CellFormatting{
				Alignment: tw.AlignCenter,
			},
			Padding: tablewriter.CellPadding{
				Global: tw.Padding{Left: "L", Right: "R", Top: "T", Bottom: "B"},
			},
		},
		Footer: tablewriter.CellConfig{
			Formatting: tablewriter.CellFormatting{
				Alignment:  tw.AlignCenter,
				AutoFormat: true,
			},
			Padding: tablewriter.CellPadding{
				Global: tw.Padding{Left: "*", Right: "*", Top: "", Bottom: ""},
			},
		},
	}

	var buf bytes.Buffer
	table := tablewriter.NewTable(&buf, tablewriter.WithConfig(c))
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
	visualCheck(t, "TableWithCustomPadding", buf.String(), expected)
}
