package tablewriter

import (
	"bytes"
	"github.com/olekukonko/tablewriter/tw"
	"testing"
)

// TestMergeCellConfig tests the mergeCellConfig function with various input configurations.
// It includes tests for the hybrid configuration approach using ConfigBuilder and Option functions.
func TestMergeCellConfig(t *testing.T) {
	// Default configuration used as the base for merging, aligned with defaultConfig() from tablewriter.go
	defaultCfg := tw.CellConfig{
		Formatting: tw.CellFormatting{
			Alignment:  tw.AlignLeft,
			AutoWrap:   tw.WrapNormal, // 1
			AutoFormat: false,
			MergeMode:  tw.MergeNone,
			MaxWidth:   0,
		},
		Padding: tw.CellPadding{
			Global: tw.Padding{Left: " ", Right: " "},
		},
	}

	tests := []struct {
		name     string
		input    tw.CellConfig
		expected tw.CellConfig
	}{
		// Test case: Empty input should preserve defaults
		{
			name:  "EmptyConfig",
			input: tw.CellConfig{},
			expected: tw.CellConfig{
				Formatting: tw.CellFormatting{
					Alignment:  tw.AlignLeft,
					AutoWrap:   tw.WrapNormal,
					AutoFormat: false,
					MergeMode:  tw.MergeNone,
					MaxWidth:   0,
				},
				Padding: tw.CellPadding{
					Global: tw.Padding{Left: " ", Right: " "},
				},
			},
		},
		// Test case: Override MergeMode to None
		{
			name: "OverrideMergeModeNone",
			input: tw.CellConfig{
				Formatting: tw.CellFormatting{
					MergeMode: tw.MergeNone,
				},
			},
			expected: tw.CellConfig{
				Formatting: tw.CellFormatting{
					Alignment:  tw.AlignLeft,
					AutoWrap:   tw.WrapNormal,
					AutoFormat: false,
					MergeMode:  tw.MergeNone,
					MaxWidth:   0,
				},
				Padding: tw.CellPadding{
					Global: tw.Padding{Left: " ", Right: " "},
				},
			},
		},
		// Test case: Override MergeMode to Vertical
		{
			name: "OverrideMergeModeVertical",
			input: tw.CellConfig{
				Formatting: tw.CellFormatting{
					MergeMode: tw.MergeVertical,
				},
			},
			expected: tw.CellConfig{
				Formatting: tw.CellFormatting{
					Alignment:  tw.AlignLeft,
					AutoWrap:   tw.WrapNormal,
					AutoFormat: false,
					MergeMode:  tw.MergeVertical,
					MaxWidth:   0,
				},
				Padding: tw.CellPadding{
					Global: tw.Padding{Left: " ", Right: " "},
				},
			},
		},
		// Test case: Override MergeMode to Horizontal
		{
			name: "OverrideMergeModeHorizontal",
			input: tw.CellConfig{
				Formatting: tw.CellFormatting{
					MergeMode: tw.MergeHorizontal,
				},
			},
			expected: tw.CellConfig{
				Formatting: tw.CellFormatting{
					Alignment:  tw.AlignLeft,
					AutoWrap:   tw.WrapNormal,
					AutoFormat: false,
					MergeMode:  tw.MergeHorizontal,
					MaxWidth:   0,
				},
				Padding: tw.CellPadding{
					Global: tw.Padding{Left: " ", Right: " "},
				},
			},
		},
		// Test case: Override MergeMode to Both
		{
			name: "OverrideMergeModeBoth",
			input: tw.CellConfig{
				Formatting: tw.CellFormatting{
					MergeMode: tw.MergeBoth,
				},
			},
			expected: tw.CellConfig{
				Formatting: tw.CellFormatting{
					Alignment:  tw.AlignLeft,
					AutoWrap:   tw.WrapNormal,
					AutoFormat: false,
					MergeMode:  tw.MergeBoth,
					MaxWidth:   0,
				},
				Padding: tw.CellPadding{
					Global: tw.Padding{Left: " ", Right: " "},
				},
			},
		},
		// Test case: Override MergeMode to Hierarchical
		{
			name: "OverrideMergeModeHierarchical",
			input: tw.CellConfig{
				Formatting: tw.CellFormatting{
					MergeMode: tw.MergeHierarchical,
				},
			},
			expected: tw.CellConfig{
				Formatting: tw.CellFormatting{
					Alignment:  tw.AlignLeft,
					AutoWrap:   tw.WrapNormal,
					AutoFormat: false,
					MergeMode:  tw.MergeHierarchical,
					MaxWidth:   0,
				},
				Padding: tw.CellPadding{
					Global: tw.Padding{Left: " ", Right: " "},
				},
			},
		},
		// Test case: Merge with ConfigBuilder using flattened methods
		// Adjusted to match defaultConfig() Header defaults
		{
			name: "ConfigBuilderFlattened",
			input: NewConfigBuilder().
				WithHeaderAlignment(tw.AlignCenter).
				WithHeaderMergeMode(tw.MergeHorizontal).
				Build().Header,
			expected: tw.CellConfig{
				Formatting: tw.CellFormatting{
					Alignment:  tw.AlignCenter,     // From builder
					AutoWrap:   tw.WrapTruncate,    // From defaultConfig() Header
					AutoFormat: true,               // From defaultConfig() Header
					MergeMode:  tw.MergeHorizontal, // From builder
					MaxWidth:   0,
				},
				Padding: tw.CellPadding{
					Global: tw.Padding{Left: " ", Right: " "},
				},
			},
		},
		// Test case: Merge with ConfigBuilder using nested methods
		// Adjusted to match defaultConfig() Header defaults
		//{
		//	name: "ConfigBuilderNested",
		//	input: NewConfigBuilder().
		//		Header().
		//		Formatting().
		//		WithAlignment(tw.AlignRight).
		//		WithMaxWidth(20).
		//		Build().
		//		Build().Header,
		//	expected: CellConfig{
		//		Formatting: CellFormatting{
		//			Alignment:  tw.AlignRight,   // From nested builder
		//			AutoWrap:   tw.WrapTruncate, // From defaultConfig() Header
		//			AutoFormat: true,            // From defaultConfig() Header
		//			MergeMode:  tw.MergeNone,
		//			MaxWidth:   20, // From nested builder
		//		},
		//		Padding: CellPadding{
		//			Global: tw.Padding{Left: " ", Right: " "},
		//		},
		//	},
		//},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Merge into a fresh copy of defaultCfg
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

// TestCallbacks tests the CellCallbacks functionality with the hybrid configuration approach.
// It verifies callbacks are triggered during rendering using WithConfig, Configure, and ConfigBuilder.
func TestCallbacks(t *testing.T) {
	tests := []struct {
		name         string
		setup        func(*Table) // How to configure the table
		expectedGlob int          // Expected global callback count
		expectedCol0 int          // Expected column 0 callback count
	}{
		// Test case: Using WithConfig Option
		{
			name: "WithConfig",
			setup: func(t *Table) {
				t.SetHeader([]string{"Name", "Email", "Age"})
				t.Append([]string{"Alice", "alice@example.com", "25"})
			},
			expectedGlob: 1, // One header line
			expectedCol0: 1, // One callback for column 0
		},
		// Test case: Using Configure method
		{
			name: "Configure",
			setup: func(t *Table) {
				t.Configure(func(cfg *Config) {
					cfg.Header.Callbacks = tw.CellCallbacks{
						Global: t.config.Header.Callbacks.Global, // Preserve from base
						PerColumn: []func(){
							t.config.Header.Callbacks.PerColumn[0], // Preserve column 0
							nil,
							nil,
						},
					}
				})
				t.SetHeader([]string{"Name", "Email", "Age"})
				t.Append([]string{"Bob", "bob@example.com", "30"})
			},
			expectedGlob: 1,
			expectedCol0: 1,
		},
		// Test case: Using ConfigBuilder
		{
			name: "ConfigBuilder",
			setup: func(t *Table) {
				config := NewConfigBuilder().
					Header().
					Build().
					Build()
				t.config = mergeConfig(t.config, config) // Apply builder config
				t.SetHeader([]string{"Name", "Email", "Age"})
				t.Append([]string{"Charlie", "charlie@example.com", "35"})
			},
			expectedGlob: 1,
			expectedCol0: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			globalCount := 0
			col0Count := 0

			// Base configuration with callbacks
			baseConfig := Config{
				Header: tw.CellConfig{
					Callbacks: tw.CellCallbacks{
						Global: func() { globalCount++ },
						PerColumn: []func(){
							func() { col0Count++ }, // Callback for column 0
							nil,
							nil,
						},
					},
				},
			}

			// Create table with base config
			table := NewTable(&buf, WithConfig(baseConfig))

			// Apply test-specific setup
			tt.setup(table)

			// Render to trigger callbacks
			table.Render()

			// Verify callback counts
			if globalCount != tt.expectedGlob {
				t.Errorf("%s: Expected global callback to run %d time(s), got %d", tt.name, tt.expectedGlob, globalCount)
			}
			if col0Count != tt.expectedCol0 {
				t.Errorf("%s: Expected column 0 callback to run %d time(s), got %d", tt.name, tt.expectedCol0, col0Count)
			}
		})
	}
}
