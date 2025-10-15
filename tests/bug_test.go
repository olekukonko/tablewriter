package tests

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/fatih/color"
	"github.com/olekukonko/tablewriter"
	"github.com/olekukonko/tablewriter/renderer"
	"github.com/olekukonko/tablewriter/tw"
)

type Cleaner string

// Note: Format() overrides String() if both exist.
func (c Cleaner) Format() string {
	return clean(string(c))
}

type Age int

// Age int will be ignore and string will be used
func (a Age) String() string {
	return fmt.Sprintf("%dyrs", a)
}

type Person struct {
	Name string
	Age  int
	City string
}

type Profile struct {
	Name Cleaner
	Age  Age
	City string
}

func TestBug252(t *testing.T) {
	var buf bytes.Buffer
	type Person struct {
		Name string
		Age  int
		City string
	}

	header := []string{"Name", "Age", "City"}
	alice := Person{Name: "Alice", Age: 25, City: "New York"}
	bob := Profile{Name: Cleaner("Bo   b"), Age: Age(30), City: "Boston"}

	t.Run("Normal", func(t *testing.T) {
		buf.Reset()
		table := tablewriter.NewTable(&buf, tablewriter.WithDebug(true))
		table.Header(header)
		table.Append("Alice", "25", "New York")
		table.Append("Bob", "30", "Boston")
		table.Render()

		expected := `
	┌───────┬─────┬──────────┐
	│ NAME  │ AGE │   CITY   │
	├───────┼─────┼──────────┤
	│ Alice │ 25  │ New York │
	│ Bob   │ 30  │ Boston   │
	└───────┴─────┴──────────┘
`
		debug := visualCheck(t, "TestBug252-Normal", buf.String(), expected)
		if !debug {
			t.Error(table.Debug())
		}
	})

	t.Run("Mixed", func(t *testing.T) {
		buf.Reset()
		table := tablewriter.NewTable(&buf, tablewriter.WithDebug(true))
		table.Header(header)
		table.Append(alice)
		table.Append("Bob", "30", "Boston")
		table.Render()

		expected := `
	┌───────┬─────┬──────────┐
	│ NAME  │ AGE │   CITY   │
	├───────┼─────┼──────────┤
	│ Alice │ 25  │ New York │
	│ Bob   │ 30  │ Boston   │
	└───────┴─────┴──────────┘
`
		debug := visualCheck(t, "TestBug252-Mixed", buf.String(), expected)
		if !debug {
			// t.Error(table.Debug())
		}
	})

	t.Run("Profile", func(t *testing.T) {
		buf.Reset()
		table := tablewriter.NewTable(&buf, tablewriter.WithDebug(true))
		table.Header(header)
		table.Append(Cleaner("A   lice"), Cleaner("2   5 yrs"), "New York")
		table.Append(bob)
		table.Render()

		expected := `
		┌───────┬───────┬──────────┐
		│ NAME  │  AGE  │   CITY   │
		├───────┼───────┼──────────┤
		│ Alice │ 25yrs │ New York │
		│ Bob   │ 30yrs │ Boston   │
		└───────┴───────┴──────────┘

`
		debug := visualCheck(t, "TestBasicTableDefault", buf.String(), expected)
		if !debug {
			// t.Error(table.Debug())
		}
	})

	type Override struct {
		Fish string `json:"name"`
		Name string `json:"-"`
		Age  Age
		City string
	}
	t.Run("Override", func(t *testing.T) {
		buf.Reset()
		table := tablewriter.NewTable(&buf, tablewriter.WithDebug(true))
		table.Header(header)
		table.Append(Cleaner("A   lice"), Cleaner("2   5 yrs"), "New York")
		table.Append(Override{
			Fish: "Bob",
			Name: "Skip",
			Age:  Age(25),
			City: "Boston",
		})
		table.Render()

		expected := `
        ┌───────┬───────┬──────────┐
        │ NAME  │  AGE  │   CITY   │
        ├───────┼───────┼──────────┤
        │ Alice │ 25yrs │ New York │
        │ Bob   │ 25yrs │ Boston   │
        └───────┴───────┴──────────┘

`
		debug := visualCheck(t, "TestBug252-Override", buf.String(), expected)
		if !debug {
			// t.Error(table.Debug())
		}
	})

	t.Run("Override-Streaming", func(t *testing.T) {
		buf.Reset()
		table := tablewriter.NewTable(&buf, tablewriter.WithDebug(true), tablewriter.WithStreaming(tw.StreamConfig{Enable: true}))

		err := table.Start()
		if err != nil {
			t.Error(err)
		}

		table.Header(header)
		table.Append(Cleaner("A   lice"), Cleaner("2   5 yrs"), "New York")
		table.Append(Override{
			Fish: "Bob",
			Name: "Skip",
			Age:  Age(25),
			City: "Boston",
		})
		expected := `
        ┌────────┬────────┬────────┐
        │  NAME  │  AGE   │  CITY  │
        ├────────┼────────┼────────┤
        │ Alice  │ 25yrs  │ New    │
        │        │        │ York   │
        │ Bob    │ 25yrs  │ Boston │
        └────────┴────────┴────────┘


`
		err = table.Close()
		if err != nil {
			t.Error(err)
		}
		debug := visualCheck(t, "TestBug252-Override-Streaming", buf.String(), expected)
		if !debug {
			// t.Error(table.Debug())
		}
	})
}

func TestBug254(t *testing.T) {
	var buf bytes.Buffer
	data := [][]string{
		{"  LEFT", "RIGHT  ", "  BOTH  "},
	}
	t.Run("Normal", func(t *testing.T) {
		buf.Reset()
		table := tablewriter.NewTable(&buf,
			tablewriter.WithRowMaxWidth(20),
			tablewriter.WithTrimSpace(tw.On),
			tablewriter.WithAlignment(tw.Alignment{tw.AlignCenter, tw.AlignCenter, tw.AlignCenter}),
		)
		table.Bulk(data)
		table.Render()

		expected := `
		┌──────┬───────┬──────┐
		│ LEFT │ RIGHT │ BOTH │
		└──────┴───────┴──────┘

`
		debug := visualCheck(t, "TestBug252-Normal", buf.String(), expected)
		if !debug {
			t.Error(table.Debug())
		}
	})

	t.Run("Mixed", func(t *testing.T) {
		buf.Reset()
		table := tablewriter.NewTable(&buf,
			tablewriter.WithRowMaxWidth(20),
			tablewriter.WithTrimSpace(tw.Off),
			tablewriter.WithAlignment(tw.Alignment{tw.AlignCenter, tw.AlignCenter, tw.AlignCenter}),
		)
		table.Bulk(data)
		table.Render()

		expected := `
	┌────────┬─────────┬──────────┐
	│   LEFT │ RIGHT   │   BOTH   │
	└────────┴─────────┴──────────┘
`
		debug := visualCheck(t, "TestBug252-Mixed", buf.String(), expected)
		if !debug {
			t.Error(table.Debug())
		}
	})
}

func TestBug260(t *testing.T) {
	var buf bytes.Buffer

	tableRendition := tw.Rendition{
		Borders: tw.BorderNone,
		Settings: tw.Settings{
			Separators: tw.Separators{
				ShowHeader:     tw.Off,
				ShowFooter:     tw.Off,
				BetweenRows:    tw.Off,
				BetweenColumns: tw.Off,
			},
		},
		Symbols: tw.NewSymbols(tw.StyleNone),
	}

	t.Run("Normal", func(t *testing.T) {
		buf.Reset()
		tableRenderer := renderer.NewBlueprint(tableRendition)
		table := tablewriter.NewTable(
			&buf,
			tablewriter.WithRenderer(tableRenderer),
			tablewriter.WithTableMax(120),
			tablewriter.WithTrimSpace(tw.Off),
			tablewriter.WithDebug(true),
			tablewriter.WithPadding(tw.PaddingNone),
		)

		table.Append([]string{
			"INFO:",
			"The original machine had a base-plate of prefabulated aluminite, surmounted by a malleable logarithmic casing in such a way that the two main spurving bearings were in a direct line with the pentametric fan.",
		})

		table.Append("INFO:",
			"The original machine had a base-plate of prefabulated aluminite, surmounted by a malleable logarithmic casing in such a way that the two main spurving bearings were in a direct line with the pentametric fan.",
		)

		table.Render()

		expected := `
		INFO:The original machine had a base-plate of prefabulated     
			 aluminite, surmounted by a malleable logarithmic casing in
			 such a way that the two main spurving bearings were in a  
			 direct line with the pentametric fan.                     
		INFO:The original machine had a base-plate of prefabulated     
			 aluminite, surmounted by a malleable logarithmic casing in
			 such a way that the two main spurving bearings were in a  
			 direct line with the pentametric fan.    
`
		debug := visualCheck(t, "TestBug252-Mixed", buf.String(), expected)
		if !debug {
			t.Error(table.Debug())
		}
	})

	t.Run("Mixed", func(t *testing.T) {
		buf.Reset()
		tableRenderer := renderer.NewBlueprint(tableRendition)
		table := tablewriter.NewTable(
			&buf,
			tablewriter.WithRenderer(tableRenderer),
			tablewriter.WithTableMax(120),
			tablewriter.WithTrimSpace(tw.Off),
			tablewriter.WithDebug(true),
			tablewriter.WithPadding(tw.PaddingNone),
		)

		table.Append([]string{
			"INFO:",
			"The original machine had a base-plate of prefabulated aluminite, surmounted by a malleable logarithmic casing in such a way that the two main spurving bearings were in a direct line with the pentametric fan.",
		})

		table.Append("INFO: ",
			"The original machine had a base-plate of prefabulated aluminite, surmounted by a malleable logarithmic casing in such a way that the two main spurving bearings were in a direct line with the pentametric fan.",
		)

		table.Render()

		expected := `
		INFO: The original machine had a base-plate of prefabulated     
			  aluminite, surmounted by a malleable logarithmic casing in
			  such a way that the two main spurving bearings were in a  
			  direct line with the pentametric fan.                     
		INFO: The original machine had a base-plate of prefabulated     
			  aluminite, surmounted by a malleable logarithmic casing in
			  such a way that the two main spurving bearings were in a  
			  direct line with the pentametric fan.  
`
		debug := visualCheck(t, "TestBug252-Mixed", buf.String(), expected)
		if !debug {
			t.Error(table.Debug())
		}
	})
}

func TestBug267(t *testing.T) {
	var buf bytes.Buffer
	data := [][]string{
		{"a", "b", "c"},
		{"aa", "bb", "cc"},
	}
	t.Run("WithoutMaxWith", func(t *testing.T) {
		buf.Reset()
		table := tablewriter.NewTable(&buf,
			tablewriter.WithTrimSpace(tw.On),
			tablewriter.WithConfig(tablewriter.Config{Row: tw.CellConfig{Padding: tw.CellPadding{
				Global: tw.PaddingNone,
			}}}),
		)
		table.Bulk(data)
		table.Render()

		expected := `
		┌──┬──┬──┐
		│a │b │c │
		│aa│bb│cc│
		└──┴──┴──┘

`
		debug := visualCheck(t, "TestBug252-Normal", buf.String(), expected)
		if !debug {
			t.Error(table.Debug())
		}
	})

	t.Run("WithMaxWidth", func(t *testing.T) {
		buf.Reset()
		table := tablewriter.NewTable(&buf,
			tablewriter.WithRowMaxWidth(20),
			tablewriter.WithTrimSpace(tw.Off),
			tablewriter.WithAlignment(tw.Alignment{tw.AlignCenter, tw.AlignCenter, tw.AlignCenter}),
			tablewriter.WithDebug(false),
			tablewriter.WithConfig(tablewriter.Config{Row: tw.CellConfig{Padding: tw.CellPadding{
				Global: tw.PaddingNone,
			}}}),
		)
		table.Bulk(data)
		table.Render()

		expected := `
            ┌──┬──┬──┐
            │a │b │c │
            │aa│bb│cc│
            └──┴──┴──┘
`
		debug := visualCheck(t, "TestBug252-Mixed", buf.String(), expected)
		if !debug {
			t.Error(table.Debug())
		}
	})
}

func TestBug271(t *testing.T) {
	var buf bytes.Buffer
	table := tablewriter.NewTable(&buf, tablewriter.WithConfig(tablewriter.Config{
		Header: tw.CellConfig{
			Merging: tw.CellMerging{
				Mode: tw.MergeHorizontal,
			},
		},
		Footer: tw.CellConfig{
			Merging: tw.CellMerging{
				Mode: tw.MergeHorizontal,
			},
			Alignment: tw.CellAlignment{PerColumn: []tw.Align{tw.Skip, tw.Skip, tw.AlignRight, tw.AlignLeft}},
		},
	}),
		tablewriter.WithRenderer(renderer.NewBlueprint(tw.Rendition{
			Settings: tw.Settings{
				Separators: tw.Separators{
					BetweenRows: tw.On,
				},
			},
		})),
	)
	table.Header([]string{"Info", "Info", "Info", "Info"})
	table.Append([]string{"1/1/2014", "Domain name", "Successful", "Successful"})
	table.Footer([]string{"", "", "TOTAL", "$145.93"}) // Fixed from Append
	table.Render()

	expected := `
        ┌──────────────────────────────────────────────────┐
        │                       INFO                       │
        ├──────────┬─────────────┬────────────┬────────────┤
        │ 1/1/2014 │ Domain name │ Successful │ Successful │
        ├──────────┴─────────────┴────────────┼────────────┤
        │                               TOTAL │ $145.93    │
        └─────────────────────────────────────┴────────────┘

`
	check := visualCheck(t, "HorizontalMergeAlignFooter", buf.String(), expected)
	if !check {
		t.Error(table.Debug())
	}
}

func TestBug289(t *testing.T) {
	var buf bytes.Buffer

	data := [][]string{
		{"Name", "Version", "Rev", "Tracking", "Publisher", "Notes"},
		{"amberol", "0.10.3", "30", "latest/stable", "alexmurray✪", "-"},
		{"android-studio", "2023.1.1", "148", "latest/stable", "snapcrafters✪", "classic"},
		{"arianna", "23.08.3", "37", "latest/stable", "kde✓", "-"},
		{"ascii-draw", "0.2.0", "66", "latest/stable", "nokse22", "-"},
		{"bare", "1.0", "5", "latest/stable", "canonical✓", "base"},
		{"beekeeper-studio", "4.1.10", "244", "latest/stable", "matthew-rathbone", "-"},
		{"blender", "4.0.2", "4300", "latest/stable", "blenderfoundati✓", "classic"},
	}

	colorCfg := renderer.ColorizedConfig{
		Header: renderer.Tint{
			FG: renderer.Colors{color.Bold}, // Bold headers
			BG: renderer.Colors{},
		},
		Column: renderer.Tint{
			FG: renderer.Colors{color.Reset},
			BG: renderer.Colors{color.Reset},
		},
		Footer: renderer.Tint{
			FG: renderer.Colors{color.Bold},
			BG: renderer.Colors{},
		},
		Borders: tw.BorderNone,
		Settings: tw.Settings{
			Separators: tw.Separators{
				ShowHeader:     tw.Off,
				ShowFooter:     tw.Off,
				BetweenRows:    tw.Off,
				BetweenColumns: tw.Off,
			},
			Lines: tw.Lines{
				ShowTop:        tw.Off,
				ShowBottom:     tw.Off,
				ShowHeaderLine: tw.Off,
				ShowFooterLine: tw.Off,
			},
		},
	}

	options := []tablewriter.Option{
		tablewriter.WithRenderer(renderer.NewColorized(colorCfg)),
		tablewriter.WithConfig(tablewriter.Config{
			MaxWidth: 80,
			Header: tw.CellConfig{
				Alignment: tw.CellAlignment{
					Global:    tw.AlignLeft,
					PerColumn: []tw.Align{tw.Skip, tw.Skip, tw.AlignRight, tw.Skip, tw.Skip},
				},
				Formatting: tw.CellFormatting{
					AutoWrap:   tw.WrapNone,
					AutoFormat: tw.On,
				},
				Merging: tw.CellMerging{
					Mode: tw.MergeNone,
				},
				Padding: tw.CellPadding{
					Global: tw.Padding{
						Left:      tw.Empty,
						Right:     "  ",
						Top:       tw.Empty,
						Bottom:    tw.Empty,
						Overwrite: true,
					},
				},
			},
			Row: tw.CellConfig{
				Formatting: tw.CellFormatting{AutoWrap: tw.WrapNormal}, // Wrap long content
				Alignment: tw.CellAlignment{
					Global:    tw.AlignLeft,
					PerColumn: []tw.Align{tw.Skip, tw.Skip, tw.AlignRight, tw.Skip, tw.Skip},
				},
				Padding: tw.CellPadding{
					Global: tw.Padding{
						Left:      tw.Empty,
						Right:     "  ",
						Top:       tw.Empty,
						Bottom:    tw.Empty,
						Overwrite: true,
					},
				},
			},
			Footer: tw.CellConfig{
				Alignment: tw.CellAlignment{Global: tw.AlignRight},
			},
		}),
	}

	table := tablewriter.NewTable(&buf, options...)
	table.Header(data[0])
	table.Bulk(data[1:])
	table.Render()

	expected := `
	NAME              VERSION    REV  TRACKING       PUBLISHER         NOTES    
	amberol           0.10.3      30  latest/stable  alexmurray✪       -        
	android-studio    2023.1.1   148  latest/stable  snapcrafters✪     classic  
	arianna           23.08.3     37  latest/stable  kde✓              -        
	ascii-draw        0.2.0       66  latest/stable  nokse22           -        
	bare              1.0          5  latest/stable  canonical✓        base     
	beekeeper-studio  4.1.10     244  latest/stable  matthew-rathbone  -        
	blender           4.0.2     4300  latest/stable  blenderfoundati✓  classic
`
	check := visualCheck(t, "HorizontalMergeAlignFooter", buf.String(), expected)
	if !check {
		t.Error(table.Debug())
	}
}
