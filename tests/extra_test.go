package tests

import (
	"bytes"
	"github.com/olekukonko/tablewriter"
	"github.com/olekukonko/tablewriter/renderer"
	"github.com/olekukonko/tablewriter/tw"
	"testing"
)

func TestFilterMasking(t *testing.T) {
	tests := []struct {
		name     string
		filter   tw.Filter
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
			table := tablewriter.NewTable(&buf, tablewriter.WithConfig(tablewriter.Config{
				Header: tw.CellConfig{
					Formatting: tw.CellFormatting{Alignment: tw.AlignCenter, AutoFormat: true},
					Padding:    tw.CellPadding{Global: tw.Padding{Left: " ", Right: " "}},
				},
				Row: tw.CellConfig{
					Formatting: tw.CellFormatting{Alignment: tw.AlignLeft},
					Padding:    tw.CellPadding{Global: tw.Padding{Left: " ", Right: " "}},
					Filter: tw.CellFilter{
						Global: tt.filter,
					},
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
			table.Header(header)
			table.Bulk(tt.data)
			table.Render()
			visualCheck(t, tt.name, buf.String(), tt.expected)
		})
	}
}

func TestMasterClass(t *testing.T) {
	var buf bytes.Buffer
	littleConfig := tablewriter.Config{
		MaxWidth: 30,
		Row: tw.CellConfig{
			Formatting: tw.CellFormatting{
				Alignment: tw.AlignCenter,
			},
			Padding: tw.CellPadding{
				Global: tw.Padding{Left: tw.Skip, Right: tw.Skip, Top: tw.Skip, Bottom: tw.Skip},
			},
		},
	}

	bigConfig := tablewriter.Config{
		MaxWidth: 50,
		Header: tw.CellConfig{Formatting: tw.CellFormatting{
			AutoWrap: tw.WrapTruncate,
		}},
		Row: tw.CellConfig{
			Formatting: tw.CellFormatting{
				Alignment: tw.AlignCenter,
			},
			Padding: tw.CellPadding{
				Global: tw.Padding{Left: tw.Skip, Right: tw.Skip, Top: tw.Skip, Bottom: tw.Skip},
			},
		},
	}

	little := func(s string) string {
		var b bytes.Buffer
		table := tablewriter.NewTable(&b,
			tablewriter.WithConfig(littleConfig),
			tablewriter.WithRenderer(renderer.NewBlueprint(tw.Rendition{
				Borders: tw.BorderNone,
				Settings: tw.Settings{
					Separators: tw.Separators{
						ShowHeader:     tw.Off,
						ShowFooter:     tw.Off,
						BetweenRows:    tw.On,
						BetweenColumns: tw.Off,
					},
					Lines: tw.Lines{
						ShowTop:        tw.Off,
						ShowBottom:     tw.Off,
						ShowHeaderLine: tw.Off,
						ShowFooterLine: tw.On,
					},
				},
			})),
		)
		table.Append([]string{s, s})
		table.Append([]string{s, s})
		table.Render()

		return b.String()
	}

	table := tablewriter.NewTable(&buf,
		tablewriter.WithConfig(bigConfig),
		tablewriter.WithDebug(true),
		tablewriter.WithRenderer(renderer.NewBlueprint(tw.Rendition{
			Borders: tw.BorderNone,
			Settings: tw.Settings{
				Separators: tw.Separators{
					ShowHeader:     tw.Off,
					ShowFooter:     tw.Off,
					BetweenRows:    tw.Off,
					BetweenColumns: tw.On,
				},
				Lines: tw.Lines{
					ShowTop:        tw.Off,
					ShowBottom:     tw.Off,
					ShowHeaderLine: tw.Off,
					ShowFooterLine: tw.Off,
				},
			},
		})),
	)
	table.Append([]string{little("A"), little("B")})
	table.Append([]string{little("C"), little("D")})
	table.Render()

	expected := `
          A A   │  B B   
         ────── │ ────── 
          A A   │  B B   
          C C   │  D D   
         ────── │ ────── 
          C C   │  D D  
`
	visualCheck(t, "Master Class", buf.String(), expected)

}

func TestConfigAutoHideDefault(t *testing.T) {
	var buf bytes.Buffer
	table := tablewriter.NewTable(&buf)

	// Use the new exported Config() method
	cfg := table.Config()
	if cfg.Behavior.AutoHide.Enabled() {
		t.Errorf("Expected AutoHide default to be false, got true")
	}
}

func TestAutoHideFeature(t *testing.T) {
	data := [][]string{
		{"A", "The Good", ""},    // Rating is empty
		{"B", "The Bad", " "},    // Rating is whitespace
		{"C", "The Ugly", "   "}, // Rating is whitespace
		{"D", "The Gopher", ""},  // Rating is empty
		// Add a row where Rating is NOT empty to test the opposite case
		{"E", "The Tester", "999"},
	}

	//  Test Case 1: Hide Empty Column
	t.Run("HideWhenEmpty", func(t *testing.T) {
		var buf bytes.Buffer
		table := tablewriter.NewTable(&buf,
			tablewriter.WithAutoHide(tw.On), // Enable the feature
			tablewriter.WithDebug(false),
		)
		table.Header([]string{"Name", "Sign", "Rating"}) // Header IS included

		// Use only data where the last column IS empty
		emptyData := [][]string{
			{"A", "The Good", ""},
			{"B", "The Bad", " "},
			{"C", "The Ugly", "   "},
			{"D", "The Gopher", ""},
		}
		for _, v := range emptyData {
			table.Append(v)
		}

		table.Render()

		// Expected output: Rating column should be completely gone
		expected := `
            ┌──────┬────────────┐
            │ NAME │    SIGN    │
            ├──────┼────────────┤
            │ A    │ The Good   │
            │ B    │ The Bad    │
            │ C    │ The Ugly   │
            │ D    │ The Gopher │
            └──────┴────────────┘
`
		// Use visualCheck, expect it might fail initially if Blueprint isn't perfect yet
		if !visualCheck(t, "AutoHide_HideWhenEmpty", buf.String(), expected) {
			t.Log("Output for HideWhenEmpty was not as expected (might be OK if Blueprint needs more fixes):")
			t.Error(buf.String())
			// Log debug info if helpful
			// for _, v := range table.Debug() {
			// 	t.Log(v)
			// }
		}
	})

	//  Test Case 2: Show Column When Not Empty
	t.Run("ShowWhenNotEmpty", func(t *testing.T) {
		var buf bytes.Buffer
		table := tablewriter.NewTable(&buf,
			tablewriter.WithAutoHide(tw.On), // Feature enabled
			// tablewriter.WithRenderer(renderer.NewBlueprint()),
		)
		table.Header([]string{"Name", "Sign", "Rating"})

		// Use data where at least one row has content in the last column
		for _, v := range data { // Use the original data mix
			table.Append(v)
		}

		table.Render()

		// Expected output: Rating column should be present because row "E" has content
		expected := `
            ┌──────┬────────────┬────────┐
            │ NAME │    SIGN    │ RATING │
            ├──────┼────────────┼────────┤
            │ A    │ The Good   │        │
            │ B    │ The Bad    │        │
            │ C    │ The Ugly   │        │
            │ D    │ The Gopher │        │
            │ E    │ The Tester │ 999    │
            └──────┴────────────┴────────┘
`
		if !visualCheck(t, "AutoHide_ShowWhenNotEmpty", buf.String(), expected) {
			t.Log("Output for ShowWhenNotEmpty was not as expected (might be OK if Blueprint needs more fixes):")
			t.Log(buf.String())
			// Log debug info if helpful
			// for _, v := range table.Debug() {
			// 	t.Log(v)
			// }
		}
	})

	//  Test Case 3: Feature Disabled
	t.Run("DisabledShowsEmpty", func(t *testing.T) {
		var buf bytes.Buffer
		table := tablewriter.NewTable(&buf,
			tablewriter.WithAutoHide(tw.Off), // Feature explicitly disabled
			// tablewriter.WithRenderer(renderer.NewBlueprint()),
		)
		table.Header([]string{"Name", "Sign", "Rating"})

		// Use only data where the last column IS empty
		emptyData := [][]string{
			{"A", "The Good", ""},
			{"B", "The Bad", " "},
			{"C", "The Ugly", "   "},
			{"D", "The Gopher", ""},
		}
		for _, v := range emptyData {
			table.Append(v)
		}

		table.Render()

		// Expected output: Rating column should be present but empty
		expected := `
            ┌──────┬────────────┬────────┐
            │ NAME │    SIGN    │ RATING │
            ├──────┼────────────┼────────┤
            │ A    │ The Good   │        │
            │ B    │ The Bad    │        │
            │ C    │ The Ugly   │        │
            │ D    │ The Gopher │        │
            └──────┴────────────┴────────┘
`
		// This one should ideally PASS if the default behavior is preserved
		if !visualCheck(t, "AutoHide_DisabledShowsEmpty", buf.String(), expected) {
			t.Errorf("AutoHide disabled test failed!")
			t.Log(buf.String())
			// Log debug info if helpful
			// for _, v := range table.Debug() {
			// 	t.Log(v)
			// }
		}
	})
}

func TestEmojiTable(t *testing.T) {
	var buf bytes.Buffer

	table := tablewriter.NewTable(&buf)
	table.Header([]string{"Name 😺", "Age 🎂", "City 🌍"})
	data := [][]string{
		{"Alice 😊", "25", "New York 🌆"},
		{"Bob 😎", "30", "Boston 🏙️"},
		{"Charlie 🤓", "28", "Tokyo 🗼"},
	}
	table.Bulk(data)
	table.Footer([]string{"", "Total 👥", "3"})
	table.Configure(func(config *tablewriter.Config) {
		config.Row.Formatting.Alignment = tw.AlignLeft
		config.Footer.Formatting.Alignment = tw.AlignRight
	})
	table.Render()

	expected := `
┌────────────┬──────────┬─────────────┐
│  NAME  😺  │ AGE  🎂  │  CITY  🌍   │
├────────────┼──────────┼─────────────┤
│ Alice 😊   │ 25       │ New York 🌆 │
│ Bob 😎     │ 30       │ Boston 🏙️    │
│ Charlie 🤓 │ 28       │ Tokyo 🗼    │
├────────────┼──────────┼─────────────┤
│            │ Total 👥 │           3 │
└────────────┴──────────┴─────────────┘

`
	if !visualCheck(t, "EmojiTable", buf.String(), expected) {
		t.Error(table.Debug().String())
	}
}

func TestUnicodeTableDefault(t *testing.T) {
	var buf bytes.Buffer

	table := tablewriter.NewTable(&buf)
	table.Header([]string{"Name", "Age", "City"})
	table.Append([]string{"Alice", "25", "New York"})
	table.Append([]string{"Bøb", "30", "Tōkyō"})    // Contains ø and ō
	table.Append([]string{"José", "28", "México"}) // Contains é and accented e (e + combining acute)
	table.Append([]string{"张三", "35", "北京"})        // Chinese characters
	table.Append([]string{"अनु", "40", "मुंबई"})    // Devanagari script
	table.Render()

	expected := `
┌───────┬─────┬──────────┐
│ NAME  │ AGE │   CITY   │
├───────┼─────┼──────────┤
│ Alice │ 25  │ New York │
│ Bøb   │ 30  │ Tōkyō    │
│ José  │ 28  │ México   │
│ 张三  │ 35  │ 北京     │
│ अनु    │ 40  │ मुंबई      │
└───────┴─────┴──────────┘

`
	visualCheck(t, "UnicodeTableRendering", buf.String(), expected)
}

func TestSpaces(t *testing.T) {
	var buf bytes.Buffer
	var data = [][]string{
		{"No", "Age", "    City"},
		{"    1", "25", "New York"},
		{"2", "30", "x"},
		{"       3", "28", "     Lagos"},
	}
	t.Run("Trim", func(t *testing.T) {
		buf.Reset()
		table := tablewriter.NewTable(&buf, tablewriter.WithDebug(false), tablewriter.WithTrimSpace(tw.On))
		table.Header(data[0])
		table.Bulk(data[1:])
		table.Render()

		expected := `
           ┌────┬─────┬──────────┐
           │ NO │ AGE │   CITY   │
           ├────┼─────┼──────────┤
           │ 1  │ 25  │ New York │
           │ 2  │ 30  │ x        │
           │ 3  │ 28  │ Lagos    │
           └────┴─────┴──────────┘
`
		if !visualCheck(t, "UnicodeTableRendering", buf.String(), expected) {
			t.Log(table.Debug())
		}
	})

	t.Run("NoTrim", func(t *testing.T) {
		buf.Reset()
		table := tablewriter.NewTable(&buf, tablewriter.WithTrimSpace(tw.Off))
		table.Header(data[0])
		table.Bulk(data[1:])
		table.Render()

		expected := `
       ┌──────────┬─────┬────────────┐
       │    NO    │ AGE │    CITY    │
       ├──────────┼─────┼────────────┤
       │     1    │ 25  │ New York   │
       │ 2        │ 30  │ x          │
       │        3 │ 28  │      Lagos │
       └──────────┴─────┴────────────┘

`
		visualCheck(t, "UnicodeTableRendering", buf.String(), expected)
	})

}
