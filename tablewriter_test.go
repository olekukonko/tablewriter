package tablewriter

import (
	"bytes"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"reflect"
	"sync"
	"testing"

	"github.com/olekukonko/ll"
	"github.com/olekukonko/tablewriter/tw"
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
			AutoWrap:   tw.WrapTruncate,
			Alignment:  tw.AlignLeft,
			AutoFormat: tw.Off,
		},
		ColMaxWidths: tw.CellWidth{Global: 10},
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
		config.ColMaxWidths.Global = 0
		got := table.calculateContentMaxWidth(0, config, 1, 1, false)
		if got != 0 {
			t.Errorf("Expected width 0, got %d", got)
		}
	})
}

// TestCallStringer tests the convertToStringer function with caching enabled and disabled.
// It verifies stringer invocation and cache behavior for custom types.
func TestCallStringer(t *testing.T) {
	table := &Table{
		logger:               ll.New("test"),
		stringer:             func(s interface{}) []string { return []string{fmt.Sprintf("%v", s)} },
		stringerCache:        map[reflect.Type]reflect.Value{},
		stringerCacheEnabled: true,
	}
	input := struct{ Name string }{Name: "test"}
	cells, err := table.convertToStringer(input)
	if err != nil {
		t.Errorf("convertToStringer failed: %v", err)
	}
	if len(cells) != 1 || cells[0] != "{test}" {
		t.Errorf("convertToStringer returned unexpected cells: %v", cells)
	}

	// Test cache hit
	cells, err = table.convertToStringer(input)
	if err != nil {
		t.Errorf("convertToStringer failed on cache hit: %v", err)
	}
	if len(cells) != 1 || cells[0] != "{test}" {
		t.Errorf("convertToStringer returned unexpected cells on cache hit: %v", cells)
	}

	// Test disabled cache
	table.stringerCacheEnabled = false
	cells, err = table.convertToStringer(input)
	if err != nil {
		t.Errorf("convertToStringer failed without cache: %v", err)
	}
	if len(cells) != 1 || cells[0] != "{test}" {
		t.Errorf("convertToStringer returned unexpected cells without cache: %v", cells)
	}
}

// TestCallStringerConcurrent tests the convertToStringer function under concurrent access.
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
			cells, err := table.convertToStringer(input)
			if err != nil {
				t.Errorf("convertToStringer failed for ID %d: %v", i, err)
			}
			expected := fmt.Sprintf("{%d}", i)
			if len(cells) != 1 || cells[0] != expected {
				t.Errorf("convertToStringer returned unexpected cells for ID %d: got %v, want [%s]", i, cells, expected)
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
		{errors.New("err"), "err"},
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

// Helper to get a fresh default CellConfig for a section.
// This uses the defaultConfig() function from the tablewriter package.
func getTestSectionDefaultConfig(section string) tw.CellConfig {
	fullDefaultCfg := defaultConfig()
	switch section {
	case "header":
		return fullDefaultCfg.Header
	case "row":
		return fullDefaultCfg.Row
	case "footer":
		return fullDefaultCfg.Footer
	default:
		return fullDefaultCfg.Row
	}
}

// TestMergeCellConfig comprehensively tests the mergeCellConfig function.
func TestMergeCellConfig(t *testing.T) {
	tests := []struct {
		name           string
		baseConfig     func() tw.CellConfig
		inputConfig    tw.CellConfig
		expectedConfig func() tw.CellConfig
	}{
		{
			name:        "EmptyInput_OnRowDefaultBase",
			baseConfig:  func() tw.CellConfig { return getTestSectionDefaultConfig("row") },
			inputConfig: tw.CellConfig{},
			expectedConfig: func() tw.CellConfig {
				cfg := getTestSectionDefaultConfig("row")
				// src.AutoFormat is 0 (Pending) from empty CellConfig.
				// mergeCellConfig assigns src.AutoFormat to dst.AutoFormat.
				cfg.Formatting.AutoFormat = tw.Pending // Should be 0
				return cfg
			},
		},
		{
			name:       "OverrideMergeModeNone_OnRowDefaultBase",
			baseConfig: func() tw.CellConfig { return getTestSectionDefaultConfig("row") },
			inputConfig: tw.CellConfig{
				Merging: tw.CellMerging{Mode: tw.MergeNone},
			},
			expectedConfig: func() tw.CellConfig {
				cfg := getTestSectionDefaultConfig("row")
				cfg.Merging.Mode = tw.MergeNone
				cfg.Formatting.AutoFormat = tw.Pending // src.AutoFormat is 0
				return cfg
			},
		},
		{
			name:       "OverrideMergeModeVertical_OnRowDefaultBase",
			baseConfig: func() tw.CellConfig { return getTestSectionDefaultConfig("row") },
			inputConfig: tw.CellConfig{
				Formatting: tw.CellFormatting{MergeMode: tw.MergeVertical},
			},
			expectedConfig: func() tw.CellConfig {
				cfg := getTestSectionDefaultConfig("row")
				cfg.Formatting.MergeMode = tw.MergeVertical
				cfg.Formatting.AutoFormat = tw.Pending // src.AutoFormat is 0
				return cfg
			},
		},
		{
			name:       "OverrideMergeModeHorizontal_OnRowDefaultBase",
			baseConfig: func() tw.CellConfig { return getTestSectionDefaultConfig("row") },
			inputConfig: tw.CellConfig{
				Formatting: tw.CellFormatting{MergeMode: tw.MergeHorizontal},
			},
			expectedConfig: func() tw.CellConfig {
				cfg := getTestSectionDefaultConfig("row")
				cfg.Formatting.MergeMode = tw.MergeHorizontal
				cfg.Formatting.AutoFormat = tw.Pending // src.AutoFormat is 0
				return cfg
			},
		},
		{
			name:       "OverrideMergeModeBoth_OnRowDefaultBase",
			baseConfig: func() tw.CellConfig { return getTestSectionDefaultConfig("row") },
			inputConfig: tw.CellConfig{
				Formatting: tw.CellFormatting{MergeMode: tw.MergeBoth},
			},
			expectedConfig: func() tw.CellConfig {
				cfg := getTestSectionDefaultConfig("row")
				cfg.Formatting.MergeMode = tw.MergeBoth
				cfg.Formatting.AutoFormat = tw.Pending // src.AutoFormat is 0
				return cfg
			},
		},
		{
			name:       "OverrideMergeModeHierarchical_OnRowDefaultBase",
			baseConfig: func() tw.CellConfig { return getTestSectionDefaultConfig("row") },
			inputConfig: tw.CellConfig{
				Formatting: tw.CellFormatting{MergeMode: tw.MergeHierarchical},
			},
			expectedConfig: func() tw.CellConfig {
				cfg := getTestSectionDefaultConfig("row")
				cfg.Formatting.MergeMode = tw.MergeHierarchical
				cfg.Formatting.AutoFormat = tw.Pending // src.AutoFormat is 0
				return cfg
			},
		},
		{
			name:       "ConfigBuilderOutput_MergingIntoRowDefault",
			baseConfig: func() tw.CellConfig { return getTestSectionDefaultConfig("row") }, // Base AutoFormat = tw.Off (-1)
			inputConfig: NewConfigBuilder().
				WithHeaderAlignment(tw.AlignCenter).
				WithHeaderMergeMode(tw.MergeHorizontal).
				Build().Header, // Src AutoFormat = tw.On (1) from defaultConfig().Header
			expectedConfig: func() tw.CellConfig {
				cfg := getTestSectionDefaultConfig("row")
				cfg.Alignment.Global = tw.AlignCenter
				cfg.Formatting.AutoWrap = tw.WrapTruncate // from defaultConfig().Header
				cfg.Formatting.AutoFormat = tw.On         // from src (Builder's Header)
				cfg.Formatting.MergeMode = tw.MergeHorizontal
				// Padding.Global is same in default row and header, so it's effectively overwritten by itself.
				return cfg
			},
		},
		{
			name:        "Row_EmptyInput_OnRowBase",
			baseConfig:  func() tw.CellConfig { return getTestSectionDefaultConfig("row") },
			inputConfig: tw.CellConfig{},
			expectedConfig: func() tw.CellConfig {
				cfg := getTestSectionDefaultConfig("row")
				cfg.Formatting.AutoFormat = tw.Pending // src.AutoFormat is 0
				return cfg
			},
		},
		{
			name:       "Row_OverrideAlignment_OnRowBase",
			baseConfig: func() tw.CellConfig { return getTestSectionDefaultConfig("row") },
			inputConfig: tw.CellConfig{
				Formatting: tw.CellFormatting{Alignment: tw.AlignCenter},
			},
			expectedConfig: func() tw.CellConfig {
				cfg := getTestSectionDefaultConfig("row")
				cfg.Formatting.Alignment = tw.AlignCenter
				cfg.Formatting.AutoFormat = tw.Pending // src.AutoFormat is 0
				return cfg
			},
		},
		{
			name:       "Row_OverrideColumnAligns_OnRowBase_WithSkip",
			baseConfig: func() tw.CellConfig { return getTestSectionDefaultConfig("row") },
			inputConfig: tw.CellConfig{
				ColumnAligns: []tw.Align{tw.AlignRight, tw.Skip, tw.AlignLeft},
			},
			expectedConfig: func() tw.CellConfig {
				cfg := getTestSectionDefaultConfig("row")
				cfg.ColumnAligns = []tw.Align{tw.AlignRight, tw.Empty, tw.AlignLeft}
				cfg.Formatting.AutoFormat = tw.Pending // src.AutoFormat is 0
				return cfg
			},
		},
		{
			name:       "Row_OverrideAutoWrap_OnRowBase",
			baseConfig: func() tw.CellConfig { return getTestSectionDefaultConfig("row") },
			inputConfig: tw.CellConfig{
				Formatting: tw.CellFormatting{AutoWrap: tw.WrapTruncate},
			},
			expectedConfig: func() tw.CellConfig {
				cfg := getTestSectionDefaultConfig("row")
				cfg.Formatting.AutoWrap = tw.WrapTruncate
				cfg.Formatting.AutoFormat = tw.Pending // src.AutoFormat is 0
				return cfg
			},
		},
		{
			name:       "Row_OverrideAutoFormat_InputOff_BaseOff",
			baseConfig: func() tw.CellConfig { return getTestSectionDefaultConfig("row") }, // Base AutoFormat = tw.Off (-1)
			inputConfig: tw.CellConfig{
				Formatting: tw.CellFormatting{AutoFormat: tw.Off}, // Src AutoFormat = tw.Off (-1)
			},
			expectedConfig: func() tw.CellConfig {
				cfg := getTestSectionDefaultConfig("row")
				cfg.Formatting.AutoFormat = tw.Off // Expected -1
				return cfg
			},
		},
		{
			name:       "Row_OverrideAutoFormat_InputOn_BaseOff",
			baseConfig: func() tw.CellConfig { return getTestSectionDefaultConfig("row") }, // Base AutoFormat = tw.Off (-1)
			inputConfig: tw.CellConfig{
				Formatting: tw.CellFormatting{AutoFormat: tw.On}, // Src AutoFormat = tw.On (1)
			},
			expectedConfig: func() tw.CellConfig {
				cfg := getTestSectionDefaultConfig("row")
				cfg.Formatting.AutoFormat = tw.On // Expected 1
				return cfg
			},
		},
		{
			name:       "Header_ConfigBuilderOutput_MergingIntoHeaderDefault",
			baseConfig: func() tw.CellConfig { return getTestSectionDefaultConfig("header") }, // Base AutoFormat = tw.On (1)
			inputConfig: NewConfigBuilder().
				WithHeaderAlignment(tw.AlignCenter).
				WithHeaderMergeMode(tw.MergeHorizontal).
				Build().Header, // Src AutoFormat = tw.On (1)
			expectedConfig: func() tw.CellConfig {
				cfg := getTestSectionDefaultConfig("header")
				cfg.Formatting.MergeMode = tw.MergeHorizontal
				cfg.Formatting.AutoFormat = tw.On
				return cfg
			},
		},
		{
			name:       "OverrideColMaxWidthGlobal_OnRowDefault",
			baseConfig: func() tw.CellConfig { return getTestSectionDefaultConfig("row") },
			inputConfig: tw.CellConfig{
				ColMaxWidths: tw.CellWidth{Global: 50},
			},
			expectedConfig: func() tw.CellConfig {
				cfg := getTestSectionDefaultConfig("row")
				cfg.ColMaxWidths.Global = 50
				cfg.Formatting.AutoFormat = tw.Pending // src.AutoFormat is 0
				return cfg
			},
		},
		{
			name:       "MergeColMaxWidthPerColumn_NewEntries_OnRowDefault",
			baseConfig: func() tw.CellConfig { return getTestSectionDefaultConfig("row") },
			inputConfig: tw.CellConfig{
				ColMaxWidths: tw.CellWidth{PerColumn: map[int]int{0: 10, 2: 20}},
			},
			expectedConfig: func() tw.CellConfig {
				cfg := getTestSectionDefaultConfig("row")
				cfg.ColMaxWidths.PerColumn = map[int]int{0: 10, 2: 20}
				cfg.Formatting.AutoFormat = tw.Pending // src.AutoFormat is 0
				return cfg
			},
		},
		{
			name: "MergeColMaxWidthPerColumn_OverwriteAndAdd_OnExisting",
			baseConfig: func() tw.CellConfig {
				cfg := getTestSectionDefaultConfig("row")
				cfg.ColMaxWidths.PerColumn = map[int]int{0: 5, 1: 15}
				return cfg
			},
			inputConfig: tw.CellConfig{
				ColMaxWidths: tw.CellWidth{PerColumn: map[int]int{0: 10, 2: 20, 1: 0}},
			},
			expectedConfig: func() tw.CellConfig {
				cfg := getTestSectionDefaultConfig("row")
				cfg.ColMaxWidths.PerColumn = map[int]int{0: 10, 1: 15, 2: 20}
				cfg.Formatting.AutoFormat = tw.Pending // src.AutoFormat is 0
				return cfg
			},
		},
		{
			name:       "MergePaddingPerColumn_NewEntries_OnRowDefault",
			baseConfig: func() tw.CellConfig { return getTestSectionDefaultConfig("row") },
			inputConfig: tw.CellConfig{
				Padding: tw.CellPadding{PerColumn: []tw.Padding{
					{Left: "L0", Right: "R0"},
					{},
					{Left: "L2", Right: "R2"},
				}},
			},
			expectedConfig: func() tw.CellConfig {
				cfg := getTestSectionDefaultConfig("row")
				cfg.Padding.PerColumn = []tw.Padding{
					{Left: "L0", Right: "R0"},
					{},
					{Left: "L2", Right: "R2"},
				}
				cfg.Formatting.AutoFormat = tw.Pending // src.AutoFormat is 0
				return cfg
			},
		},
		{
			name: "MergePaddingPerColumn_OverwriteExtendAndPreserve_OnExisting",
			baseConfig: func() tw.CellConfig {
				cfg := getTestSectionDefaultConfig("row")
				cfg.Padding.PerColumn = []tw.Padding{
					{Left: "BASE_L0"}, {Left: "BASE_L1"}, {Left: "BASE_L2"},
				}
				return cfg
			},
			inputConfig: tw.CellConfig{
				Padding: tw.CellPadding{PerColumn: []tw.Padding{
					{Left: "SRC_L0"},
					{},
				}},
			},
			expectedConfig: func() tw.CellConfig {
				cfg := getTestSectionDefaultConfig("row")
				cfg.Padding.PerColumn = []tw.Padding{
					{Left: "SRC_L0"}, {Left: "BASE_L1"}, {Left: "BASE_L2"},
				}
				cfg.Formatting.AutoFormat = tw.Pending // src.AutoFormat is 0
				return cfg
			},
		},
		{
			name: "MergeColumnAligns_SrcShorterThanDst_WithSkipAndEmpty",
			baseConfig: func() tw.CellConfig {
				cfg := getTestSectionDefaultConfig("row")
				cfg.ColumnAligns = []tw.Align{tw.AlignCenter, tw.AlignRight, tw.AlignCenter}
				return cfg
			},
			inputConfig: tw.CellConfig{
				ColumnAligns: []tw.Align{tw.AlignLeft, tw.Skip},
			},
			expectedConfig: func() tw.CellConfig {
				cfg := getTestSectionDefaultConfig("row")
				cfg.ColumnAligns = []tw.Align{tw.AlignLeft, tw.AlignRight, tw.AlignCenter}
				cfg.Formatting.AutoFormat = tw.Pending // src.AutoFormat is 0
				return cfg
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			baseForTest := tt.baseConfig()
			expected := tt.expectedConfig()

			got := mergeCellConfig(baseForTest, tt.inputConfig)

			// Use a more verbose comparison for Formatting if DeepEqual fails to pinpoint AutoFormat
			if !reflect.DeepEqual(got.Formatting, expected.Formatting) {
				if got.Formatting.Alignment != expected.Formatting.Alignment ||
					got.Formatting.AutoWrap != expected.Formatting.AutoWrap ||
					got.Formatting.MergeMode != expected.Formatting.MergeMode ||
					got.Formatting.AutoFormat != expected.Formatting.AutoFormat { // Key comparison
					t.Errorf("Formatting mismatch:\nexpected: Alignment:%v, AutoWrap:%d, MergeMode:%d, AutoFormat:%d (%s)\ngot:      Alignment:%v, AutoWrap:%d, MergeMode:%d, AutoFormat:%d (%s)",
						expected.Formatting.Alignment, expected.Formatting.AutoWrap, expected.Formatting.MergeMode, expected.Formatting.AutoFormat, expected.Formatting.AutoFormat.String(),
						got.Formatting.Alignment, got.Formatting.AutoWrap, got.Formatting.MergeMode, got.Formatting.AutoFormat, got.Formatting.AutoFormat.String())
				} else {
					// Fallback if the above doesn't catch it (shouldn't happen for simple fields)
					t.Errorf("Formatting mismatch (DeepEqual failed):\nexpected: %+v\ngot:      %+v", expected.Formatting, got.Formatting)
				}
			}

			if !reflect.DeepEqual(got.Padding.Global, expected.Padding.Global) {
				t.Errorf("Padding.Global mismatch\nexpected: %+v\ngot:      %+v", expected.Padding.Global, got.Padding.Global)
			}
			if !reflect.DeepEqual(got.Padding.PerColumn, expected.Padding.PerColumn) {
				t.Errorf("Padding.PerColumn mismatch\nexpected: %#v\ngot:      %#v", expected.Padding.PerColumn, got.Padding.PerColumn)
			}
			if !reflect.DeepEqual(got.ColMaxWidths.Global, expected.ColMaxWidths.Global) {
				t.Errorf("ColMaxWidths.Global mismatch\nexpected: %d\ngot:      %d", expected.ColMaxWidths.Global, got.ColMaxWidths.Global)
			}
			if !reflect.DeepEqual(got.ColMaxWidths.PerColumn, expected.ColMaxWidths.PerColumn) {
				t.Errorf("ColMaxWidths.PerColumn mismatch\nexpected: %#v\ngot:      %#v", expected.ColMaxWidths.PerColumn, got.ColMaxWidths.PerColumn)
			}
			if !reflect.DeepEqual(got.ColumnAligns, expected.ColumnAligns) {
				t.Errorf("ColumnAligns mismatch\nexpected: %#v\ngot:      %#v", expected.ColumnAligns, got.ColumnAligns)
			}

			if (got.Callbacks.Global == nil) != (expected.Callbacks.Global == nil) {
				t.Errorf("Callbacks.Global nilness mismatch\nexpected nil: %t, got nil: %t",
					expected.Callbacks.Global == nil, got.Callbacks.Global == nil)
			}
			if len(got.Callbacks.PerColumn) != len(expected.Callbacks.PerColumn) {
				t.Errorf("Callbacks.PerColumn length mismatch\nexpected: %d, got: %d",
					len(expected.Callbacks.PerColumn), len(got.Callbacks.PerColumn))
			} else {
				for i := range got.Callbacks.PerColumn {
					if (got.Callbacks.PerColumn[i] == nil) != (expected.Callbacks.PerColumn[i] == nil) {
						t.Errorf("Callbacks.PerColumn[%d] nilness mismatch\nexpected nil: %t, got nil: %t",
							i, expected.Callbacks.PerColumn[i] == nil, got.Callbacks.PerColumn[i] == nil)
						break
					}
				}
			}
			if (got.Filter.Global == nil) != (expected.Filter.Global == nil) {
				t.Errorf("Filter.Global nilness mismatch\nexpected nil: %t, got nil: %t",
					expected.Filter.Global == nil, got.Filter.Global == nil)
			}
			if len(got.Filter.PerColumn) != len(expected.Filter.PerColumn) {
				t.Errorf("Filter.PerColumn length mismatch\nexpected: %d, got: %d",
					len(expected.Filter.PerColumn), len(got.Filter.PerColumn))
			} else {
				for i := range got.Filter.PerColumn {
					if (got.Filter.PerColumn[i] == nil) != (expected.Filter.PerColumn[i] == nil) {
						t.Errorf("Filter.PerColumn[%d] nilness mismatch\nexpected nil: %t, got nil: %t",
							i, expected.Filter.PerColumn[i] == nil, got.Filter.PerColumn[i] == nil)
						break
					}
				}
			}
		})
	}
}
