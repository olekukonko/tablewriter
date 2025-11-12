package twwidth

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"testing"

	"github.com/mattn/go-runewidth"
)

func TestMain(m *testing.M) {
	mu.Lock()
	globalOptions = newOptions()
	mu.Unlock()
	os.Exit(m.Run())
}

func TestFilter(t *testing.T) {
	ansi := Filter()
	tests := []struct {
		input    string
		expected bool
	}{
		{"\033[31m", true},
		{"\033]8;;http://example.com\007", true},
		{"hello", false},
		{"\033[m", true},
		{"\033[1;34;40m", true},
		{"\033invalid", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if got := ansi.MatchString(tt.input); got != tt.expected {
				t.Errorf("Filter().MatchString(%q) = %v, want %v", tt.input, got, tt.expected)
			}
		})
	}
}

func TestSetEastAsian(t *testing.T) {
	original := globalOptions.EastAsianWidth
	SetEastAsian(true)
	if !globalOptions.EastAsianWidth {
		t.Errorf("SetEastAsian(true): condition.EastAsianWidth = false, want true")
	}
	SetEastAsian(false)
	if globalOptions.EastAsianWidth {
		t.Errorf("SetEastAsian(false): condition.EastAsianWidth = true, want false")
	}
	mu.Lock()
	globalOptions.EastAsianWidth = original
	mu.Unlock()
}

func TestWidth(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		eastAsian     bool
		expectedWidth int
	}{
		{
			name:          "ASCII",
			input:         "hello",
			eastAsian:     false,
			expectedWidth: 5,
		},
		{
			name:          "Unicode with ANSI",
			input:         "\033[31m☆あ\033[0m",
			eastAsian:     false,
			expectedWidth: 3,
		},
		{
			name:          "Unicode with EastAsian",
			input:         "\033[31m☆あ\033[0m",
			eastAsian:     true,
			expectedWidth: 4,
		},
		{
			name:          "Empty string",
			input:         "",
			eastAsian:     false,
			expectedWidth: 0,
		},
		{
			name:          "Only ANSI",
			input:         "\033[31m\033[0m",
			eastAsian:     false,
			expectedWidth: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			SetEastAsian(tt.eastAsian)
			got := Width(tt.input)
			if got != tt.expectedWidth {
				t.Errorf("Width(%q) = %d, want %d (EastAsian=%v)", tt.input, got, tt.expectedWidth, tt.eastAsian)
			}
		})
	}
}

func TestDisplay(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		eastAsian     bool
		expectedWidth int
	}{
		{
			name:          "ASCII",
			input:         "hello",
			eastAsian:     false,
			expectedWidth: 5,
		},
		{
			name:          "Unicode with ANSI",
			input:         "\033[31m☆あ\033[0m",
			eastAsian:     false,
			expectedWidth: 3,
		},
		{
			name:          "Unicode with EastAsian",
			input:         "\033[31m☆あ\033[0m",
			eastAsian:     true,
			expectedWidth: 4,
		},
		{
			name:          "Empty string",
			input:         "",
			eastAsian:     false,
			expectedWidth: 0,
		},
		{
			name:          "Only ANSI",
			input:         "\033[31m\033[0m",
			eastAsian:     false,
			expectedWidth: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cond := &runewidth.Condition{
				EastAsianWidth: tt.eastAsian,
			}
			got := Display(cond, tt.input)
			if got != tt.expectedWidth {
				t.Errorf("Display(%q, options) = %d, want %d (EastAsian=%v)", tt.input, got, tt.expectedWidth, tt.eastAsian)
			}
		})
	}
}

func TestTruncate(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		maxWidth  int
		suffix    []string
		eastAsian bool
		expected  string
	}{
		{
			name:      "ASCII within width",
			input:     "hello",
			maxWidth:  5,
			suffix:    nil,
			eastAsian: false,
			expected:  "hello",
		},
		{
			name:      "ASCII with suffix",
			input:     "hello",
			maxWidth:  8,
			suffix:    []string{"..."},
			eastAsian: false,
			expected:  "hello...",
		},
		{
			name:      "ASCII truncate",
			input:     "hello",
			maxWidth:  3,
			suffix:    []string{"..."},
			eastAsian: false,
			expected:  "...",
		},
		{
			name:      "Unicode with ANSI, no truncate",
			input:     "\033[31m☆あ\033[0m",
			maxWidth:  3,
			suffix:    nil,
			eastAsian: false,
			expected:  "\033[31m☆あ\033[0m",
		},
		{
			name:      "Unicode with ANSI, truncate",
			input:     "\033[31m☆あ\033[0m",
			maxWidth:  2,
			suffix:    []string{"..."},
			eastAsian: false,
			expected:  "",
		},
		{
			name:      "Unicode with EastAsian",
			input:     "\033[31m☆あ\033[0m",
			maxWidth:  3,
			suffix:    []string{"..."},
			eastAsian: true,
			expected:  "...",
		},
		{
			name:      "Zero maxWidth",
			input:     "hello",
			maxWidth:  0,
			suffix:    []string{"..."},
			eastAsian: false,
			expected:  "",
		},
		{
			name:      "Negative maxWidth",
			input:     "hello",
			maxWidth:  -1,
			suffix:    []string{"..."},
			eastAsian: false,
			expected:  "",
		},
		{
			name:      "Empty string with suffix",
			input:     "",
			maxWidth:  3,
			suffix:    []string{"..."},
			eastAsian: false,
			expected:  "...",
		},
		{
			name:      "Only ANSI",
			input:     "\033[31m\033[0m",
			maxWidth:  3,
			suffix:    []string{"..."},
			eastAsian: false,
			expected:  "...",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			SetEastAsian(tt.eastAsian)
			got := Truncate(tt.input, tt.maxWidth, tt.suffix...)
			if got != tt.expected {
				t.Errorf("Truncate(%q, %d, %v) (EA=%v) = %q, want %q", tt.input, tt.maxWidth, tt.suffix, tt.eastAsian, got, tt.expected)
			}
		})
	}
}

func TestConcurrentSetEastAsian(t *testing.T) {
	var wg sync.WaitGroup
	numGoroutines := 10
	iterations := 100

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(enable bool) {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				SetEastAsian(enable)
			}
		}(i%2 == 0)
	}

	wg.Wait()
}

func TestWidthWithEnvironment(t *testing.T) {
	original := os.Getenv("RUNEWIDTH_EASTASIAN")
	defer func() {
		if original == "" {
			os.Unsetenv("RUNEWIDTH_EASTASIAN")
		} else {
			os.Setenv("RUNEWIDTH_EASTASIAN", original)
		}
	}()

	os.Setenv("RUNEWIDTH_EASTASIAN", "0")
	SetEastAsian(false)
	if got := Width("☆あ"); got != 3 {
		t.Errorf("Width(☆あ) with RUNEWIDTH_EASTASIAN=0 = %d, want 3", got)
	}

	os.Setenv("RUNEWIDTH_EASTASIAN", "1")
	SetEastAsian(true)
	if got := Width("☆あ"); got != 4 {
		t.Errorf("Width(☆あ) with RUNEWIDTH_EASTASIAN=1 = %d, want 4", got)
	}
}

func resetGlobalCache() {
	mu.Lock()
	widthCache = make(map[cacheKey]int)
	mu.Unlock()
}

var benchmarkStrings = map[string]string{
	"SimpleASCII":       "hello world, this is a test string.",
	"ASCIIWithANSI":     "\033[31mhello\033[0m \033[34mworld\033[0m, this is \033[1ma\033[0m test string.",
	"EastAsian":         "こんにちは世界、これはテスト文字列です。",
	"EastAsianWithANSI": "\033[32mこんにちは\033[0m \033[35m世界\033[0m、これは\033[4mテスト\033[0m文字列です。",
	"LongSimpleASCII":   strings.Repeat("abcdefghijklmnopqrstuvwxyz ", 20),
	"LongASCIIWithANSI": strings.Repeat("\033[31ma\033[32mb\033[33mc\033[34md\033[35me\033[36mf\033[0m ", 50),
}

func TestNewOptions(t *testing.T) {
	// Test that newOptions() correctly picks up settings from go-runewidth
	cond := runewidth.NewCondition()
	options := newOptions()

	// Verify that EastAsianWidth is correctly copied from runewidth condition
	if options.EastAsianWidth != cond.EastAsianWidth {
		t.Errorf("newOptions().EastAsianWidth = %v, want %v (from runewidth.NewCondition())",
			options.EastAsianWidth, cond.EastAsianWidth)
	}
}

func TestNewOptionsWithEnvironment(t *testing.T) {
	// Test that newOptions() respects environment variables that affect go-runewidth
	testCases := []struct {
		name        string
		runewidthEA string
		locale      string
		expectedEA  bool
	}{
		{
			name:        "Default environment",
			runewidthEA: "",
			locale:      "",
			expectedEA:  false, // Default behavior
		},
		{
			name:        "RUNEWIDTH_EASTASIAN=1",
			runewidthEA: "1",
			locale:      "",
			expectedEA:  true,
		},
		{
			name:        "RUNEWIDTH_EASTASIAN=0",
			runewidthEA: "0",
			locale:      "",
			expectedEA:  false,
		},
		{
			name:        "Japanese locale",
			runewidthEA: "",
			locale:      "ja_JP.UTF-8",
			expectedEA:  true, // Japanese locale typically enables East Asian width
		},
		{
			name:        "Chinese locale",
			runewidthEA: "",
			locale:      "zh_CN.UTF-8",
			expectedEA:  true, // Chinese locale typically enables East Asian width
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Use t.Setenv for automatic cleanup (Go 1.17+)
			if tc.runewidthEA != "" {
				t.Setenv("RUNEWIDTH_EASTASIAN", tc.runewidthEA)
			} else {
				t.Setenv("RUNEWIDTH_EASTASIAN", "") // This effectively unsets it
			}

			if tc.locale != "" {
				t.Setenv("LC_ALL", tc.locale)
			} else {
				t.Setenv("LC_ALL", "") // This effectively unsets it
			}

			// Create a new runewidth condition with current environment
			cond := runewidth.NewCondition()

			// Get options from our function
			options := newOptions()

			// Verify that our function matches the runewidth condition
			if options.EastAsianWidth != cond.EastAsianWidth {
				t.Errorf("newOptions().EastAsianWidth = %v, want %v (from runewidth.NewCondition() with env: RUNEWIDTH_EASTASIAN=%s, LC_ALL=%s)",
					options.EastAsianWidth, cond.EastAsianWidth, tc.runewidthEA, tc.locale)
			}
		})
	}
}

func BenchmarkWidthFunction(b *testing.B) {
	eastAsianSettings := []bool{false, true}

	for name, str := range benchmarkStrings {
		for _, eaSetting := range eastAsianSettings {
			SetEastAsian(eaSetting)

			b.Run(fmt.Sprintf("%s_EA%v_NoCache", name, eaSetting), func(b *testing.B) {
				b.SetBytes(int64(len(str)))
				b.ReportAllocs()
				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					_ = WidthNoCache(str)
				}
			})

			b.Run(fmt.Sprintf("%s_EA%v_CacheMiss", name, eaSetting), func(b *testing.B) {
				b.SetBytes(int64(len(str)))
				b.ReportAllocs()
				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					_ = Width(str + strconv.Itoa(i))
				}
			})
			resetGlobalCache()

			b.Run(fmt.Sprintf("%s_EA%v_CacheHit", name, eaSetting), func(b *testing.B) {
				b.SetBytes(int64(len(str)))
				b.ReportAllocs()
				b.ResetTimer()
				if b.N > 0 {
					_ = Width(str)
				}
				for i := 1; i < b.N; i++ {
					_ = Width(str)
				}
			})
			resetGlobalCache()
		}
	}
}
