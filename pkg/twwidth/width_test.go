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
	condition = runewidth.NewCondition()
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
	original := condition.EastAsianWidth
	SetEastAsian(true)
	if !condition.EastAsianWidth {
		t.Errorf("SetEastAsian(true): condition.EastAsianWidth = false, want true")
	}
	SetEastAsian(false)
	if condition.EastAsianWidth {
		t.Errorf("SetEastAsian(false): condition.EastAsianWidth = true, want false")
	}
	mu.Lock()
	condition.EastAsianWidth = original
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
			cond := &runewidth.Condition{EastAsianWidth: tt.eastAsian}
			got := Display(cond, tt.input)
			if got != tt.expectedWidth {
				t.Errorf("Display(%q, cond) = %d, want %d (EastAsian=%v)", tt.input, got, tt.expectedWidth, tt.eastAsian)
			}
		})
	}
}

func TestTruncate(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		maxWidth  int // This is now totalBudgetForOutput
		suffix    []string
		eastAsian bool
		expected  string
	}{
		{
			name:      "ASCII within width", // sDisplayWidth (5) <= maxWidth (5). No suffix.
			input:     "hello",
			maxWidth:  5,
			suffix:    nil,
			eastAsian: false,
			expected:  "hello", // CORRECT - No change.
		},
		{
			name:      "ASCII with suffix", // sDisplayWidth (5) + suffixWidth (3) = 8. maxWidth (8).
			input:     "hello",             // sDisplayWidth (5) <= maxWidth (8) is true.
			maxWidth:  8,                   // Then check sDisplayWidth+suffixDisplayWidth <= maxWidth (5+3 <= 8), true.
			suffix:    []string{"..."},
			eastAsian: false,
			expected:  "hello...", // CORRECT - No change.
		},
		{
			name:      "ASCII truncate", // sDisplayWidth (5) > maxWidth (3). Truncation needed.
			input:     "hello",
			maxWidth:  3,               // totalBudgetForOutput = 3
			suffix:    []string{"..."}, // suffixWidth = 3
			eastAsian: false,
			// New behavior: contentBudget = 3 - 3 = 0.
			// Truncate "hello" to width 0 -> "".
			// Append "..." -> "...".
			expected: "...", // <<<< CHANGED from "hel..."
		},
		{
			name:      "Unicode with ANSI, no truncate", // sDisplayWidth (3) <= maxWidth (3). No suffix.
			input:     "\033[31m☆あ\033[0m",
			maxWidth:  3,
			suffix:    nil,
			eastAsian: false,
			expected:  "\033[31m☆あ\033[0m", // CORRECT - No change.
		},
		{
			name:      "Unicode with ANSI, truncate", // sDisplayWidth (3) > maxWidth (2). Truncation needed.
			input:     "\033[31m☆あ\033[0m",           // ☆ (1), あ (2 with EA=false) -> width 3.
			maxWidth:  2,                             // totalBudgetForOutput = 2
			suffix:    []string{"..."},               // suffixWidth = 3
			eastAsian: false,
			// New behavior: contentBudget = 2 - 3 = -1.
			// Since contentBudget < 0, check if suffix fits: suffixWidth (3) > maxWidth (2) -> No.
			// Returns "".
			expected: "", // <<<< CHANGED from "\033[31m☆\033[0m..."
			// OLD logic: content budget = 2. Truncated to ☆ (width 1). Added ... -> ☆...
			// The previous logic might have been flawed if `maxWidth` was for content only.
			// If the old `expected` was correct, it implied `maxWidth=2` was budget for content, making output `☆...` width `1+3=4`.
			// With new logic, if `maxWidth=2` is total, and suffix is `...` (width 3), can't fit.
			// Let's re-evaluate: if `maxWidth` is total budget for output, including suffix.
			// `input`="☆あ" (width ☆=1, あ=2 -> total 3 if EA=false, or ☆=1, あ=1 -> total 2 if not counting `runewidth` on `あ` correctly)
			// Assuming ☆=1, あ=1 for EA=false based on original tests. So input width = 2.
			// If input width = 2 and maxWidth = 2, it should fit as `\033[31m☆あ\033[0m`.
			// Let's use the Width function values: input="\033[31m☆あ\033[0m", EA=false -> Width = 3. (☆=1, あ=2 if `runewidth` default)
			// Okay, Width("☆あ", EA=false) = 3 (☆=1, あ=2).
			// sDisplayWidth (3) > maxWidth (2). Truncation.
			// contentBudget = 2 (maxWidth) - 3 (suffixWidth) = -1.
			// Can suffix fit? suffixWidth (3) not <= maxWidth (2). So, return "". This is correct under new rule.
		},
		{
			name:      "Unicode with EastAsian", // sDisplayWidth (4 with EA=true) > maxWidth (3). Truncation.
			input:     "\033[31m☆あ\033[0m",      // ☆ (2), あ (2) -> width 4
			maxWidth:  3,                        // totalBudgetForOutput = 3
			suffix:    []string{"..."},          // suffixWidth = 3
			eastAsian: true,
			// New behavior: contentBudget = 3 - 3 = 0.
			// Truncate to width 0 -> "".
			// Append "..." -> "...".
			// Also, there's the special EastAsian case: if currentGlobalEastAsianWidth is true,
			// and provisionalContentWidth (maxWidth - suffixDisplayWidth) == 0, return suffixStr.
			// Here, 3 - 3 == 0. So returns "..."
			expected: "...", // CORRECT - This was already the expectation that drove a previous change.
		},
		{
			name:      "Zero maxWidth", // maxWidth = 0.
			input:     "hello",
			maxWidth:  0,
			suffix:    []string{"..."},
			eastAsian: false,
			expected:  "", // CORRECT - Truncate returns "" if maxWidth is 0 and sDisplayWidth > 0.
		},
		{
			name:      "Negative maxWidth", // maxWidth = -1.
			input:     "hello",
			maxWidth:  -1,
			suffix:    []string{"..."},
			eastAsian: false,
			expected:  "", // CORRECT - Truncate returns "" if maxWidth < 0.
		},
		{
			name:      "Empty string with suffix", // sDisplayWidth = 0.
			input:     "",
			maxWidth:  3,
			suffix:    []string{"..."}, // suffixWidth = 3.
			eastAsian: false,
			// Behavior: sDisplayWidth == 0. suffix exists. suffixWidth (3) <= maxWidth (3). Returns suffixStr.
			expected: "...", // CORRECT.
		},
		{
			name:      "Only ANSI", // sDisplayWidth = 0.
			input:     "\033[31m\033[0m",
			maxWidth:  3,
			suffix:    []string{"..."}, // suffixWidth = 3.
			eastAsian: false,
			// Behavior: sDisplayWidth == 0. suffix exists. suffixWidth (3) <= maxWidth (3). Returns suffixStr.
			expected: "...", // CORRECT.
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			SetEastAsian(tt.eastAsian) // This sets the global state for Width() calls inside Truncate
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

// Helper to reset the cache for a clean benchmark state
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
	"LongSimpleASCII":   strings.Repeat("abcdefghijklmnopqrstuvwxyz ", 20),                                    // 27*20 = 540 chars
	"LongASCIIWithANSI": strings.Repeat("\033[31ma\033[32mb\033[33mc\033[34md\033[35me\033[36mf\033[0m ", 50), // ~15*50 = 750 chars with ANSI
}

func BenchmarkWidthFunction(b *testing.B) {
	eastAsianSettings := []bool{false, true}

	for name, str := range benchmarkStrings {
		for _, eaSetting := range eastAsianSettings {
			// Ensure the global EastAsian setting is correct for the sub-benchmark
			// SetEastAsian also clears the cache, which is good for starting fresh.
			SetEastAsian(eaSetting)

			// Benchmark: No Caching (using our internal non-cached version)
			b.Run(fmt.Sprintf("%s_EA%v_NoCache", name, eaSetting), func(b *testing.B) {
				b.ReportAllocs()
				for i := 0; i < b.N; i++ {
					_ = WidthNoCache(str)
				}
			})

			// Benchmark: Cached Version - Cache Misses
			// We make strings unique to ensure cache misses.
			// SetEastAsian above already cleared the cache.
			b.Run(fmt.Sprintf("%s_EA%v_CacheMiss", name, eaSetting), func(b *testing.B) {
				b.ReportAllocs()
				for i := 0; i < b.N; i++ {
					// Make string unique for each iteration to force cache miss
					_ = Width(str + strconv.Itoa(i))
				}
			})
			resetGlobalCache() // Clear cache before the cache hit test

			// Benchmark: Cached Version - Cache Hits
			// First call populates, subsequent calls hit.
			// SetEastAsian above (or resetGlobalCache) ensures cache is empty at start.
			b.Run(fmt.Sprintf("%s_EA%v_CacheHit", name, eaSetting), func(b *testing.B) {
				b.ReportAllocs()
				// First call will populate the cache
				if b.N > 0 {
					_ = Width(str)
				}
				// Subsequent calls should hit the cache
				for i := 1; i < b.N; i++ { // Start from 1 if first call is outside/setup
					_ = Width(str)
				}
			})
			resetGlobalCache() // Clean up for next iteration/test
		}
	}
}
