package tablewriter

import (
	"github.com/olekukonko/tablewriter/renderer"
	"github.com/olekukonko/tablewriter/symbols"
	"testing"
)

func TestMergeCellConfig(t *testing.T) {
	defaultCfg := CellConfig{
		Formatting: CellFormatting{
			Alignment:  renderer.AlignLeft,
			AutoWrap:   WrapNormal,
			AutoFormat: false,
			AutoMerge:  false,
			MaxWidth:   0,
		},
		Padding: CellPadding{
			Global: symbols.Padding{Left: " ", Right: " "},
		},
	}

	tests := []struct {
		name     string
		input    CellConfig
		expected CellConfig
	}{
		{
			name:  "EmptyConfig",
			input: CellConfig{},
			expected: CellConfig{
				Formatting: CellFormatting{
					Alignment:  renderer.AlignLeft,
					AutoWrap:   WrapNormal,
					AutoFormat: false,
					AutoMerge:  false,
					MaxWidth:   0,
				},
				Padding: CellPadding{
					Global: symbols.Padding{Left: " ", Right: " "},
				},
			},
		},
		{
			name: "OverrideAlignment",
			input: CellConfig{
				Formatting: CellFormatting{
					Alignment: renderer.AlignCenter,
				},
			},
			expected: CellConfig{
				Formatting: CellFormatting{
					Alignment:  renderer.AlignCenter,
					AutoWrap:   WrapNormal,
					AutoFormat: false,
					AutoMerge:  false,
					MaxWidth:   0,
				},
				Padding: CellPadding{
					Global: symbols.Padding{Left: " ", Right: " "},
				},
			},
		},
		{
			name: "OverridePadding",
			input: CellConfig{
				Padding: CellPadding{
					Global: symbols.Padding{Left: ">", Right: "<"},
				},
			},
			expected: CellConfig{
				Formatting: CellFormatting{
					Alignment:  renderer.AlignLeft,
					AutoWrap:   WrapNormal,
					AutoFormat: false,
					AutoMerge:  false,
					MaxWidth:   0,
				},
				Padding: CellPadding{
					Global: symbols.Padding{Left: ">", Right: "<"},
				},
			},
		},
		{
			name: "AddPerColumnPadding",
			input: CellConfig{
				Padding: CellPadding{
					PerColumn: []symbols.Padding{
						{Left: "|", Right: "|"},
					},
				},
			},
			expected: CellConfig{
				Formatting: CellFormatting{
					Alignment:  renderer.AlignLeft,
					AutoWrap:   WrapNormal,
					AutoFormat: false,
					AutoMerge:  false,
					MaxWidth:   0,
				},
				Padding: CellPadding{
					Global:    symbols.Padding{Left: " ", Right: " "},
					PerColumn: []symbols.Padding{{Left: "|", Right: "|"}},
				},
			},
		},
		{
			name: "OverrideAutoWrap",
			input: CellConfig{
				Formatting: CellFormatting{
					AutoWrap: WrapTruncate,
				},
			},
			expected: CellConfig{
				Formatting: CellFormatting{
					Alignment:  renderer.AlignLeft,
					AutoWrap:   WrapTruncate,
					AutoFormat: false,
					AutoMerge:  false,
					MaxWidth:   0,
				},
				Padding: CellPadding{
					Global: symbols.Padding{Left: " ", Right: " "},
				},
			},
		},
		{
			name: "AddColumnAligns",
			input: CellConfig{
				ColumnAligns: []string{renderer.AlignRight},
			},
			expected: CellConfig{
				Formatting: CellFormatting{
					Alignment:  renderer.AlignLeft,
					AutoWrap:   WrapNormal,
					AutoFormat: false,
					AutoMerge:  false,
					MaxWidth:   0,
				},
				Padding: CellPadding{
					Global: symbols.Padding{Left: " ", Right: " "},
				},
				ColumnAligns: []string{renderer.AlignRight},
			},
		},
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
