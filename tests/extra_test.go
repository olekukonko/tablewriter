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
