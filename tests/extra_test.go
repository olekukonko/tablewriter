package tests

import (
	"bytes"
	"testing"

	"github.com/olekukonko/tablewriter"
	"github.com/olekukonko/tablewriter/renderer"
	"github.com/olekukonko/tablewriter/tw"
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
					Formatting: tw.CellFormatting{AutoFormat: tw.On},
					Alignment:  tw.CellAlignment{Global: tw.AlignCenter},
					Padding:    tw.CellPadding{Global: tw.Padding{Left: " ", Right: " "}},
				},
				Row: tw.CellConfig{
					Alignment: tw.CellAlignment{Global: tw.AlignLeft},
					Padding:   tw.CellPadding{Global: tw.Padding{Left: " ", Right: " "}},
					Filter: tw.CellFilter{
						Global: tt.filter,
					},
				},
			}))
			header := []string{"Name", tt.name, "Age"}
			switch tt.name {
			case "MaskEmail":
				header[1] = "Email"
			case "MaskPassword":
				header[1] = "Password"
			case "MaskCard":
				header[1] = "Credit Card"
			}
			table.Header(header)
			table.Bulk(tt.data)
			table.Render()

			if !visualCheck(t, tt.name, buf.String(), tt.expected) {
				t.Error(table.Debug())
			}
		})
	}
}

func TestMasterClass(t *testing.T) {
	var buf bytes.Buffer
	littleConfig := tablewriter.Config{
		MaxWidth: 30,
		Row: tw.CellConfig{
			Alignment: tw.CellAlignment{Global: tw.AlignCenter},
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
			Alignment: tw.CellAlignment{Global: tw.AlignCenter},
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
	if !visualCheck(t, "Master Class", buf.String(), expected) {
		t.Error(table.Debug())
	}
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
			t.Error(table.Debug())
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
			t.Error(table.Debug())
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
			t.Error(table.Debug())
		}
	})
}

func TestEmojiTable(t *testing.T) {
	// Original Boston Test (Updated expectation as discussed)
	t.Run("Cityscape", func(t *testing.T) {
		var buf bytes.Buffer
		table := tablewriter.NewTable(&buf, tablewriter.WithDebug(true))
		table.Header([]string{"Name 😺", "Age 🎂", "City 🌍"})
		data := [][]string{
			{"Alice 😊", "25", "New York 🌆"},
			{"Bob 😎", "30", "Boston 🏙️"},
			{"Charlie 🤓", "28", "Tokyo 🗼"},
		}
		table.Bulk(data)
		table.Footer([]string{"", "Total 👥", "3"})
		table.Configure(func(config *tablewriter.Config) {
			config.Row.Alignment.Global = tw.AlignLeft
			config.Footer.Alignment.Global = tw.AlignRight
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
		if !visualCheck(t, "Cityscape", buf.String(), expected) {
			t.Error(table.Debug())
		}
	})

	// New Test: Common Emojis
	// This verifies if standard faces/objects also calculate as Width 1
	// causing the "extra space" visual effect in the padding.
	t.Run("CommonEmojis", func(t *testing.T) {
		var buf bytes.Buffer
		table := tablewriter.NewTable(&buf)
		// "Status" is 6 chars wide.
		// "Icon" is 4 chars wide.
		table.Header([]string{"Item", "Status", "Icon"})

		table.Append([]string{"Work", "Done", "✅"})   // Check mark
		table.Append([]string{"Happy", "Great", "😀"}) // Grinning Face
		table.Append([]string{"Food", "Yum", "🍕"})    // Pizza
		table.Append([]string{"Tech", "Fast", "🚀"})   // Rocket

		table.Render()
		expected := `
┌───────┬────────┬──────┐
│ ITEM  │ STATUS │ ICON │
├───────┼────────┼──────┤
│ Work  │ Done   │ ✅   │
│ Happy │ Great  │ 😀   │
│ Food  │ Yum    │ 🍕   │
│ Tech  │ Fast   │ 🚀   │
└───────┴────────┴──────┘
`

		if !visualCheck(t, "CommonEmojis", buf.String(), expected) {
			t.Error(table.Debug())
		}
	})
}

func TestUnicodeTableDefault(t *testing.T) {
	// Original test case: Mixed ASCII, Latin-1, Emoji/Symbols, and Devanagari
	t.Run("Mixed", func(t *testing.T) {
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
		if !visualCheck(t, "UnicodeTableRendering_Mixed", buf.String(), expected) {
			t.Error(table.Debug())
		}
	})

	// New test case: 100% Wide Characters (Chinese)
	// This verifies alignment when every single character is Double Width (2 cells).
	t.Run("PureChinese", func(t *testing.T) {
		var buf bytes.Buffer

		table := tablewriter.NewTable(&buf)
		// Headers: Name, Age, Hometown
		table.Header([]string{"姓名", "年龄", "籍贯"})

		// Row 1: Zhang San, Twenty-Five, Beijing
		table.Append([]string{"张三", "二十五", "北京"})

		// Row 2: Li Si, Thirty, Shanghai
		// Note: "三十" (30) is 2 chars (4 width), "二十五" (25) is 3 chars (6 width).
		// This forces alignment padding on the "30" cell.
		table.Append([]string{"李四", "三十", "上海"})

		table.Render()

		// Logic:
		// Col 1: Max width 4 ("张三"). Visual: 1 space + 4 content + 1 space = 6.
		// Col 2: Max width 6 ("二十五"). Visual: 1 space + 6 content + 1 space = 8.
		// Col 3: Max width 4 ("北京"). Visual: 1 space + 4 content + 1 space = 6.
		// "三十" needs 2 extra spaces padding to match "二十五".
		expected := `
┌──────┬────────┬──────┐
│ 姓名 │  年龄  │ 籍贯 │
├──────┼────────┼──────┤
│ 张三 │ 二十五 │ 北京 │
│ 李四 │ 三十   │ 上海 │
└──────┴────────┴──────┘
`
		if !visualCheck(t, "UnicodeTableRendering_PureChinese", buf.String(), expected) {
			t.Error(table.Debug())
		}
	})
}

func TestSpaces(t *testing.T) {
	var buf bytes.Buffer
	data := [][]string{
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
			t.Error(table.Debug())
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

func TestControl(t *testing.T) {
	var buf bytes.Buffer
	data := [][]string{
		{"No", "Age", "    City"},
		{"    1", "25", "New York"},
		{"2", "30", "x"},
		{"       3", "28", "     Lagos"},
	}
	t.Run("Trim", func(t *testing.T) {
		buf.Reset()
		table := tablewriter.NewTable(&buf,
			tablewriter.WithDebug(false),
			tablewriter.WithTrimSpace(tw.On),
			tablewriter.WithHeaderControl(tw.Control{Hide: tw.On}),
		)
		table.Header(data[0])
		table.Bulk(data[1:])
		table.Render()

		expected := `
		┌───┬────┬──────────┐
		│ 1 │ 25 │ New York │
		│ 2 │ 30 │ x        │
		│ 3 │ 28 │ Lagos    │
		└───┴────┴──────────┘

`
		if !visualCheck(t, "UnicodeTableRendering", buf.String(), expected) {
			t.Log(table.Debug())
		}
	})

	t.Run("NoTrim", func(t *testing.T) {
		buf.Reset()
		table := tablewriter.NewTable(&buf,
			tablewriter.WithTrimSpace(tw.On),
			tablewriter.WithHeaderControl(tw.Control{Hide: tw.Off}),
		)
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
			t.Error(table.Debug())
		}
	})
}
