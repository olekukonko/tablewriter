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

func BenchmarkWidthFunction(b *testing.B) {
	eastAsianSettings := []bool{false, true}

	for name, str := range benchmarkStrings {
		for _, eaSetting := range eastAsianSettings {
			SetEastAsian(eaSetting)

			b.Run(fmt.Sprintf("%s_EA%v_NoCache", name, eaSetting), func(b *testing.B) {
				b.ReportAllocs()
				for i := 0; i < b.N; i++ {
					_ = WidthNoCache(str)
				}
			})

			b.Run(fmt.Sprintf("%s_EA%v_CacheMiss", name, eaSetting), func(b *testing.B) {
				b.ReportAllocs()
				for i := 0; i < b.N; i++ {
					_ = Width(str + strconv.Itoa(i))
				}
			})
			resetGlobalCache()

			b.Run(fmt.Sprintf("%s_EA%v_CacheHit", name, eaSetting), func(b *testing.B) {
				b.ReportAllocs()
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
