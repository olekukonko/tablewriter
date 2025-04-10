package tablewriter

import (
	"bytes"
	"github.com/olekukonko/tablewriter/tw"
	"testing"
)

func TestMergeCellConfig(t *testing.T) {
	defaultCfg := CellConfig{
		Formatting: CellFormatting{
			Alignment:  tw.AlignLeft,
			AutoWrap:   tw.WrapNormal,
			AutoFormat: false,
			MergeMode:  tw.MergeNone,
			MaxWidth:   0,
		},
		Padding: CellPadding{
			Global: tw.Padding{Left: " ", Right: " "},
		},
	}

	tests := []struct {
		name     string
		input    CellConfig
		expected CellConfig
	}{
		// Existing cases...
		{
			name:  "EmptyConfig",
			input: CellConfig{},
			expected: CellConfig{
				Formatting: CellFormatting{
					Alignment:  tw.AlignLeft,
					AutoWrap:   tw.WrapNormal,
					AutoFormat: false,
					MergeMode:  tw.MergeNone,
					MaxWidth:   0,
				},
				Padding: CellPadding{
					Global: tw.Padding{Left: " ", Right: " "},
				},
			},
		},

		// Test cases for MergeMode
		{
			name: "OverrideMergeModeNone",
			input: CellConfig{
				Formatting: CellFormatting{
					MergeMode: tw.MergeNone,
				},
			},
			expected: CellConfig{
				Formatting: CellFormatting{
					Alignment:  tw.AlignLeft,
					AutoWrap:   tw.WrapNormal,
					AutoFormat: false,
					MergeMode:  tw.MergeNone,
					MaxWidth:   0,
				},
				Padding: CellPadding{
					Global: tw.Padding{Left: " ", Right: " "},
				},
			},
		},
		{
			name: "OverrideMergeModeVertical",
			input: CellConfig{
				Formatting: CellFormatting{
					MergeMode: tw.MergeVertical,
				},
			},
			expected: CellConfig{
				Formatting: CellFormatting{
					Alignment:  tw.AlignLeft,
					AutoWrap:   tw.WrapNormal,
					AutoFormat: false,
					MergeMode:  tw.MergeVertical,
					MaxWidth:   0,
				},
				Padding: CellPadding{
					Global: tw.Padding{Left: " ", Right: " "},
				},
			},
		},
		{
			name: "OverrideMergeModeHorizontal",
			input: CellConfig{
				Formatting: CellFormatting{
					MergeMode: tw.MergeHorizontal,
				},
			},
			expected: CellConfig{
				Formatting: CellFormatting{
					Alignment:  tw.AlignLeft,
					AutoWrap:   tw.WrapNormal,
					AutoFormat: false,
					MergeMode:  tw.MergeHorizontal,
					MaxWidth:   0,
				},
				Padding: CellPadding{
					Global: tw.Padding{Left: " ", Right: " "},
				},
			},
		},
		{
			name: "OverrideMergeModeBoth",
			input: CellConfig{
				Formatting: CellFormatting{
					MergeMode: tw.MergeBoth,
				},
			},
			expected: CellConfig{
				Formatting: CellFormatting{
					Alignment:  tw.AlignLeft,
					AutoWrap:   tw.WrapNormal,
					AutoFormat: false,
					MergeMode:  tw.MergeBoth,
					MaxWidth:   0,
				},
				Padding: CellPadding{
					Global: tw.Padding{Left: " ", Right: " "},
				},
			},
		},
		{
			name: "OverrideMergeModeHierarchical",
			input: CellConfig{
				Formatting: CellFormatting{
					MergeMode: tw.MergeHierarchical,
				},
			},
			expected: CellConfig{
				Formatting: CellFormatting{
					Alignment:  tw.AlignLeft,
					AutoWrap:   tw.WrapNormal,
					AutoFormat: false,
					MergeMode:  tw.MergeHierarchical,
					MaxWidth:   0,
				},
				Padding: CellPadding{
					Global: tw.Padding{Left: " ", Right: " "},
				},
			},
		},
		// Additional test cases...
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a copy of the default config to merge into
			baseCfg := defaultCfg
			got := mergeCellConfig(baseCfg, tt.input)

			// Compare Formatting
			if got.Formatting != tt.expected.Formatting {
				t.Errorf("%s: Formatting mismatch\nexpected: %+v\ngot:      %+v",
					tt.name, tt.expected.Formatting, got.Formatting)
			}

			// Compare Padding.Global
			if got.Padding.Global != tt.expected.Padding.Global {
				t.Errorf("%s: Padding.Global mismatch\nexpected: %+v\ngot:      %+v",
					tt.name, tt.expected.Padding.Global, got.Padding.Global)
			}

			// Compare Padding.PerColumn
			if len(got.Padding.PerColumn) != len(tt.expected.Padding.PerColumn) {
				t.Errorf("%s: Padding.PerColumn length mismatch - expected %d, got %d",
					tt.name, len(tt.expected.Padding.PerColumn), len(got.Padding.PerColumn))
			} else {
				for i, pad := range tt.expected.Padding.PerColumn {
					if got.Padding.PerColumn[i] != pad {
						t.Errorf("%s: Padding.PerColumn[%d] mismatch - expected %+v, got %+v",
							tt.name, i, pad, got.Padding.PerColumn[i])
					}
				}
			}

			// Compare ColumnAligns
			if len(got.ColumnAligns) != len(tt.expected.ColumnAligns) {
				t.Errorf("%s: ColumnAligns length mismatch - expected %d, got %d",
					tt.name, len(tt.expected.ColumnAligns), len(got.ColumnAligns))
			} else {
				for i, align := range tt.expected.ColumnAligns {
					if got.ColumnAligns[i] != align {
						t.Errorf("%s: ColumnAligns[%d] mismatch - expected %s, got %s",
							tt.name, i, align, got.ColumnAligns[i])
					}
				}
			}
		})
	}
}

func TestCallbacks(t *testing.T) {
	var buf bytes.Buffer
	globalCount := 0
	col0Count := 0
	table := NewTable(&buf, WithConfig(Config{
		Header: CellConfig{
			Callbacks: CellCallbacks{
				Global: func() { globalCount++ },
				PerColumn: []func(){
					func() { col0Count++ }, // Callback for column 0
					nil,                    // No callback for column 1
					nil,                    // No callback for column 2
				},
			},
		},
	}))
	table.SetHeader([]string{"Name", "Email", "Age"})
	table.Append([]string{"Alice", "alice@example.com", "25"})
	table.Render()

	if globalCount != 1 {
		t.Errorf("Expected global callback to run 1 time, got %d", globalCount)
	}
	if col0Count != 1 {
		t.Errorf("Expected column 0 callback to run 1 time, got %d", col0Count)
	}
}
