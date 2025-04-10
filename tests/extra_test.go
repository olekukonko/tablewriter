package tests

import (
	"bytes"
	"github.com/olekukonko/tablewriter"
	"github.com/olekukonko/tablewriter/tw"
	"testing"
)

func TestFilterMasking(t *testing.T) {
	tests := []struct {
		name     string
		filter   tablewriter.Filter
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
				Header: tablewriter.CellConfig{
					Formatting: tablewriter.CellFormatting{Alignment: tw.AlignCenter, AutoFormat: true},
					Padding:    tablewriter.CellPadding{Global: tw.Padding{Left: " ", Right: " "}},
				},
				Row: tablewriter.CellConfig{
					Formatting: tablewriter.CellFormatting{Alignment: tw.AlignLeft},
					Padding:    tablewriter.CellPadding{Global: tw.Padding{Left: " ", Right: " "}},
					Filter: tablewriter.CellFilter{
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
			table.SetHeader(header)
			table.Bulk(tt.data)
			table.Render()
			visualCheck(t, tt.name, buf.String(), tt.expected)
		})
	}
}

//func TestMasterClass(t *testing.T) {
//	var buf bytes.Buffer
//
//	littleConfig := tablewriter.Config{
//		MaxWidth: 30,
//		Row: tablewriter.CellConfig{
//			Formatting: tablewriter.CellFormatting{
//				Alignment: tw.AlignCenter,
//			},
//			Padding: tablewriter.CellPadding{
//				Global: tw.Padding{Left: tw.Skip, Right: tw.Skip, Top: tw.Skip, Bottom: tw.Skip},
//			},
//		},
//	}
//
//	bigConfig := tablewriter.Config{
//		MaxWidth: 50,
//		Header: tablewriter.CellConfig{Formatting: tablewriter.CellFormatting{
//			AutoWrap: tw.WrapTruncate,
//		}},
//	}
//
//	little := func(s string) string {
//		var b bytes.Buffer
//		table := tablewriter.NewTable(&b,
//			tablewriter.WithConfig(littleConfig),
//			tablewriter.WithRenderer(renderer.NewDefault(renderer.DefaultConfig{
//				Borders: renderer.BorderNone,
//				Settings: renderer.Settings{
//					Separators: renderer.Separators{
//						ShowHeader:     tw.Off,
//						ShowFooter:     tw.Off,
//						BetweenRows:    tw.Off,
//						BetweenColumns: tw.Off,
//					},
//					Lines: renderer.Lines{
//						ShowTop:        tw.Off,
//						ShowBottom:     tw.Off,
//						ShowHeaderLine: tw.Off,
//						ShowFooterLine: tw.Off,
//					},
//				},
//				Debug: false,
//			})),
//		)
//		table.Append([]string{s, s})
//		table.Append([]string{s, s})
//		table.Render()
//
//		return b.String()
//	}
//
//	table := tablewriter.NewTable(&buf,
//		tablewriter.WithConfig(bigConfig),
//		tablewriter.WithRenderer(renderer.NewDefault(renderer.DefaultConfig{
//			//Borders: renderer.BorderNone,
//			Settings: renderer.Settings{
//				Separators: renderer.Separators{
//					ShowHeader:     tw.Off,
//					ShowFooter:     tw.Off,
//					BetweenRows:    tw.Off,
//					BetweenColumns: tw.Off,
//				},
//				Lines: renderer.Lines{
//					ShowTop:        tw.Off,
//					ShowBottom:     tw.Off,
//					ShowHeaderLine: tw.Off,
//					ShowFooterLine: tw.Off,
//				},
//			},
//			Debug: false,
//		})),
//	)
//	table.Append([]string{little("A"), little("B")})
//	table.Append([]string{little("C"), little("D")})
//	table.Render()
//
//	expected := `
//
//`
//	visualCheck(t, "BasicTableRendering", buf.String(), expected)
//
//}
