package tablewriter

import (
	"bytes"
	"database/sql"
	"fmt"
	"github.com/olekukonko/ll"
	"github.com/olekukonko/tablewriter/tw"
	"os"
	"reflect"
	"sync"
	"testing"
)

// TestBuildPaddingLineContents tests the buildPaddingLineContents function with various configurations.
// It verifies padding line construction with different widths and merge states.
func TestBuildPaddingLineContents(t *testing.T) {
	table := NewTable(os.Stdout)
	widths := tw.NewMapper[int, int]()
	widths.Set(0, 5)
	widths.Set(1, 10)
	merges := map[int]tw.MergeState{
		1: {Horizontal: tw.MergeStateOption{Present: true, Start: false}},
	}
	t.Run("Basic Padding", func(t *testing.T) {
		line := table.buildPaddingLineContents("-", widths, 2, nil)
		expected := []string{"-----", "----------"}
		if !reflect.DeepEqual(line, expected) {
			t.Errorf("Expected %v, got %v", expected, line)
		}
	})
	t.Run("With Merge", func(t *testing.T) {
		line := table.buildPaddingLineContents("-", widths, 2, merges)
		expected := []string{"-----", ""}
		if !reflect.DeepEqual(line, expected) {
			t.Errorf("Expected %v, got %v", expected, line)
		}
	})
	t.Run("Zero Width", func(t *testing.T) {
		widths.Set(0, 0)
		line := table.buildPaddingLineContents("-", widths, 2, nil)
		expected := []string{"", "----------"}
		if !reflect.DeepEqual(line, expected) {
			t.Errorf("Expected %v, got %v", expected, line)
		}
	})
}

// TestCalculateContentMaxWidth tests the calculateContentMaxWidth function in batch and streaming modes.
// It verifies width calculations with various constraints and padding settings.
func TestCalculateContentMaxWidth(t *testing.T) {
	table := NewTable(os.Stdout)
	config := tw.CellConfig{
		Formatting: tw.CellFormatting{
			MaxWidth:   10,
			AutoWrap:   tw.WrapTruncate,
			Alignment:  tw.AlignLeft,
			AutoFormat: false,
		},
		Padding: tw.CellPadding{
			Global: tw.Padding{Left: " ", Right: " "},
		},
	}
	t.Run("Batch Mode with MaxWidth", func(t *testing.T) {
		got := table.calculateContentMaxWidth(0, config, 1, 1, false)
		if got != 8 { // 10 - 1 (left) - 1 (right)
			t.Errorf("Expected width 8, got %d", got)
		}
	})
	t.Run("Streaming Mode", func(t *testing.T) {
		table.streamWidths = map[int]int{0: 12}
		table.config.Stream.Enable = true
		table.hasPrinted = true
		got := table.calculateContentMaxWidth(0, config, 1, 1, true)
		if got != 10 { // 12 - 1 (left) - 1 (right)
			t.Errorf("Expected width 10, got %d", got)
		}
	})
	t.Run("No Constraint in Batch", func(t *testing.T) {
		config.Formatting.MaxWidth = 0
		got := table.calculateContentMaxWidth(0, config, 1, 1, false)
		if got != 0 {
			t.Errorf("Expected width 0, got %d", got)
		}
	})
}

// TestCallStringer tests the callStringer function with caching enabled and disabled.
// It verifies stringer invocation and cache behavior for custom types.
func TestCallStringer(t *testing.T) {
	table := &Table{
		logger:               ll.New("test"),
		stringer:             func(s interface{}) []string { return []string{fmt.Sprintf("%v", s)} },
		stringerCache:        map[reflect.Type]reflect.Value{},
		stringerCacheEnabled: true,
	}
	input := struct{ Name string }{Name: "test"}
	cells, err := table.callStringer(input)
	if err != nil {
		t.Errorf("callStringer failed: %v", err)
	}
	if len(cells) != 1 || cells[0] != "{test}" {
		t.Errorf("callStringer returned unexpected cells: %v", cells)
	}

	// Test cache hit
	cells, err = table.callStringer(input)
	if err != nil {
		t.Errorf("callStringer failed on cache hit: %v", err)
	}
	if len(cells) != 1 || cells[0] != "{test}" {
		t.Errorf("callStringer returned unexpected cells on cache hit: %v", cells)
	}

	// Test disabled cache
	table.stringerCacheEnabled = false
	cells, err = table.callStringer(input)
	if err != nil {
		t.Errorf("callStringer failed without cache: %v", err)
	}
	if len(cells) != 1 || cells[0] != "{test}" {
		t.Errorf("callStringer returned unexpected cells without cache: %v", cells)
	}
}

// TestCallStringerConcurrent tests the callStringer function under concurrent access.
// It verifies thread-safety of the stringer cache with multiple goroutines.
func TestCallStringerConcurrent(t *testing.T) {
	table := &Table{
		logger:               ll.New("test"),
		stringer:             func(s interface{}) []string { return []string{fmt.Sprintf("%v", s)} },
		stringerCacheEnabled: true,
		stringerCache:        map[reflect.Type]reflect.Value{},
	}
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			input := struct{ ID int }{ID: i}
			cells, err := table.callStringer(input)
			if err != nil {
				t.Errorf("callStringer failed for ID %d: %v", i, err)
			}
			expected := fmt.Sprintf("{%d}", i)
			if len(cells) != 1 || cells[0] != expected {
				t.Errorf("callStringer returned unexpected cells for ID %d: got %v, want [%s]", i, cells, expected)
			}
		}(i)
	}
	wg.Wait()
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
				t.Header([]string{"Name", "Email", "Age"})
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
				t.Header([]string{"Name", "Email", "Age"})
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
				t.Header([]string{"Name", "Email", "Age"})
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

// TestConvertToString tests the convertToString function with various input types.
// It verifies correct string conversion for nil, strings, SQL null types, and errors.
func TestConvertToString(t *testing.T) {
	table := &Table{logger: ll.New("test")}
	tests := []struct {
		input    interface{}
		expected string
	}{
		{nil, ""},
		{"test", "test"},
		{[]byte("bytes"), "bytes"},
		{fmt.Errorf("err"), "err"},
		{sql.NullString{String: "valid", Valid: true}, "valid"},
		{sql.NullString{Valid: false}, ""},
		// Add more cases
	}
	for _, tt := range tests {
		result := table.convertToString(tt.input)
		if tt.expected != result {
			t.Errorf("ConvertToString: expected %v, got '%v'", tt.expected, result)
		}
	}
}

// TestEnsureStreamWidthsCalculated tests the ensureStreamWidthsCalculated function.
// It verifies stream width initialization in streaming mode with various inputs.
func TestEnsureStreamWidthsCalculated(t *testing.T) {
	table := NewTable(os.Stdout, WithStreaming(tw.StreamConfig{Enable: true}))
	sampleData := []string{"A", "B"}
	config := tw.CellConfig{}
	t.Run("Already Initialized", func(t *testing.T) {
		table.streamWidths = tw.NewMapper[int, int]().Set(0, 5).Set(1, 5)
		table.streamNumCols = 2
		err := table.ensureStreamWidthsCalculated(sampleData, config)
		if err != nil {
			t.Errorf("Expected nil, got %v", err)
		}
	})
	t.Run("Initialize New", func(t *testing.T) {
		table.streamWidths = nil
		table.streamNumCols = 0
		err := table.ensureStreamWidthsCalculated(sampleData, config)
		if err != nil {
			t.Errorf("Expected nil, got %v", err)
		}
		if table.streamNumCols != 2 {
			t.Errorf("Expected streamNumCols=2, got %d", table.streamNumCols)
		}
		if table.streamWidths.Len() < 2 {
			t.Errorf("Expected at least 2 widths, got %d", table.streamWidths.Len())
		}
	})
	t.Run("Zero Columns", func(t *testing.T) {
		table.streamWidths = nil
		table.streamNumCols = 0
		err := table.ensureStreamWidthsCalculated([]string{}, config)
		if err == nil || err.Error() != "failed to determine column count for streaming" {
			t.Errorf("Expected error for zero columns, got %v", err)
		}
	})
}

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

// TestMergeCellConfig2 tests the mergeCellConfig function with section-specific default configurations.
// It verifies merging behavior for header, row, and footer sections with various input overrides.
func TestMergeCellConfig2(t *testing.T) {
	tests := []struct {
		name        string
		baseSection string // "header", "row", or "footer" to pick the correct default base
		input       tw.CellConfig
		expected    tw.CellConfig
	}{
		// --- Test Cases for ROW section (using defaultConfig().Row as base) ---
		{
			name:        "Row_EmptyInput",
			baseSection: "row",
			input:       tw.CellConfig{},
			expected:    getDefaultSectionConfig("row"), // Expected is just the default
		},
		{
			name:        "Row_OverrideAlignment",
			baseSection: "row",
			input: tw.CellConfig{
				Formatting: tw.CellFormatting{Alignment: tw.AlignCenter},
			},
			expected: func() tw.CellConfig {
				cfg := getDefaultSectionConfig("row")
				cfg.Formatting.Alignment = tw.AlignCenter
				return cfg
			}(),
		},
		{
			name:        "Row_OverrideColumnAligns",
			baseSection: "row",
			input: tw.CellConfig{
				ColumnAligns: []tw.Align{tw.AlignRight, tw.Skip, tw.AlignLeft},
			},
			expected: func() tw.CellConfig {
				cfg := getDefaultSectionConfig("row")
				cfg.ColumnAligns = []tw.Align{tw.AlignRight, tw.Skip, tw.AlignLeft}
				return cfg
			}(),
		},
		{
			name:        "Row_OverrideAutoWrap",
			baseSection: "row",
			input: tw.CellConfig{
				Formatting: tw.CellFormatting{AutoWrap: tw.WrapTruncate},
			},
			expected: func() tw.CellConfig {
				cfg := getDefaultSectionConfig("row")
				cfg.Formatting.AutoWrap = tw.WrapTruncate
				return cfg
			}(),
		},
		{
			name:        "Row_OverrideAutoFormatTrue",
			baseSection: "row",
			input: tw.CellConfig{
				Formatting: tw.CellFormatting{AutoFormat: true}, // default Row is false
			},
			expected: func() tw.CellConfig {
				cfg := getDefaultSectionConfig("row")
				cfg.Formatting.AutoFormat = true // Because of current OR logic in merge
				return cfg
			}(),
		},
		// --- Test Cases for HEADER section (using defaultConfig().Header as base) ---
		{
			name:        "Header_FromBuilderFlattened",
			baseSection: "header",
			input:       NewConfigBuilder().WithHeaderAlignment(tw.AlignCenter).WithHeaderMergeMode(tw.MergeHorizontal).Build().Header,
			expected: func() tw.CellConfig {
				cfg := getDefaultSectionConfig("header") // Starts with Header defaults
				cfg.Formatting.Alignment = tw.AlignCenter
				cfg.Formatting.MergeMode = tw.MergeHorizontal
				return cfg
			}(),
		},
		// Add more tests for other fields (MaxWidth, Padding.PerColumn, Filters, Callbacks)
		// for different sections if their defaults differ significantly.
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			baseCfg := getDefaultSectionConfig(tt.baseSection)
			// Make a deep copy of baseCfg to avoid modification across tests if it contains slices/maps
			// For structs with simple types, direct assignment is fine for 'baseCfg'.
			// If CellConfig had map fields, a proper deep copy would be safer.
			// Slices like ColumnAligns are handled by mergeCellConfig replacing them.

			got := mergeCellConfig(baseCfg, tt.input)

			// Using reflect.DeepEqual for simpler comparison of entire structs.
			// Fallback to field-by-field if DeepEqual gives unhelpful diffs or has issues with func types.
			if !reflect.DeepEqual(got.Formatting, tt.expected.Formatting) {
				t.Errorf("Formatting mismatch\nexpected: %+v\ngot:      %+v", tt.expected.Formatting, got.Formatting)
			}
			if !reflect.DeepEqual(got.Padding, tt.expected.Padding) {
				t.Errorf("Padding mismatch\nexpected: %+v\ngot:      %+v", tt.expected.Padding, got.Padding)
			}
			if !reflect.DeepEqual(got.Callbacks, tt.expected.Callbacks) {
				t.Errorf("Callbacks mismatch\nexpected: %+v\ngot:      %+v", tt.expected.Callbacks, got.Callbacks)
			}
			if !reflect.DeepEqual(got.Filter, tt.expected.Filter) {
				t.Errorf("Filter mismatch\nexpected: %+v\ngot:      %+v", tt.expected.Filter, got.Filter)
			}
			if !reflect.DeepEqual(got.ColumnAligns, tt.expected.ColumnAligns) {
				t.Errorf("ColumnAligns mismatch\nexpected: %#v\ngot:      %#v", tt.expected.ColumnAligns, got.ColumnAligns)
			}
			if !reflect.DeepEqual(got.ColMaxWidths, tt.expected.ColMaxWidths) {
				t.Errorf("ColMaxWidths mismatch\nexpected: %+v\ngot:      %+v", tt.expected.ColMaxWidths, got.ColMaxWidths)
			}
		})
	}
}

// Helper to get a fresh default CellConfig for a section
func getDefaultSectionConfig(section string) tw.CellConfig {
	fullDefaultCfg := defaultConfig() // Assuming defaultConfig() is accessible
	switch section {
	case "header":
		return fullDefaultCfg.Header
	case "row":
		return fullDefaultCfg.Row
	case "footer":
		return fullDefaultCfg.Footer
	default:
		panic("unknown section: " + section)
	}
}
