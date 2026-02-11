package twwidth

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"
	"testing"

	"github.com/mattn/go-runewidth"
	"github.com/olekukonko/tablewriter/pkg/twcache"
)

func helperProcess() {
	if IsEastAsian() {
		fmt.Fprint(os.Stdout, "true")
	} else {
		fmt.Fprint(os.Stdout, "false")
	}
}

func resetGlobalCache() {
	mu.Lock()
	widthCache = twcache.NewLRU[cacheKey, int](cacheCapacity)
	mu.Unlock()
}

func TestMain(m *testing.M) {
	if os.Getenv("GO_TEST_SUBPROCESS") == "1" {
		helperProcess()
		return
	}
	os.Exit(m.Run())
}

func TestInitRespectsEnvironment(t *testing.T) {
	testCases := []struct {
		name       string
		envVar     string
		wantOutput string
	}{
		{"RUNEWIDTH_EASTASIAN=1", "1", "true"},
		{"RUNEWIDTH_EASTASIAN=0", "0", "false"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cmd := exec.Command(os.Args[0], "-test.run=^TestHelperProcess$")

			cmd.Env = append(os.Environ(), "GO_TEST_SUBPROCESS=1", "RUNEWIDTH_EASTASIAN="+tc.envVar)

			output, err := cmd.CombinedOutput()
			if err != nil {
				t.Fatalf("command failed: %v\nOutput: %s", err, output)
			}

			got := strings.TrimSpace(string(output))
			if got != tc.wantOutput {
				t.Errorf("with RUNEWIDTH_EASTASIAN=%s, IsEastAsian() was %s, want %s", tc.envVar, got, tc.wantOutput)
			}
		})
	}
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
	original := IsEastAsian()
	t.Cleanup(func() {
		SetEastAsian(original)
	})

	SetEastAsian(true)
	if !IsEastAsian() {
		t.Errorf("SetEastAsian(true): IsEastAsian() = false, want true")
	}
	SetEastAsian(false)
	if IsEastAsian() {
		t.Errorf("SetEastAsian(false): IsEastAsian() = true, want false")
	}
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

func TestCoverageWidth(t *testing.T) {
	t.Run("Direct Width Functions", func(t *testing.T) {
		w := WidthNoCache("abc")
		if w != 3 {
			t.Errorf("WidthNoCache('abc') = %d, want 3", w)
		}

		opts := Options{EastAsianWidth: true}
		w2 := WidthWithOptions("abc", opts)
		if w2 != 3 {
			t.Errorf("WidthWithOptions result wrong")
		}
	})

	t.Run("SetCondition", func(t *testing.T) {
		original := IsEastAsian()
		defer SetEastAsian(original)

		cond := &runewidth.Condition{EastAsianWidth: !original}
		SetCondition(cond)

		if IsEastAsian() == original {
			t.Error("SetCondition failed to update global state")
		}
	})

	t.Run("Truncate ANSI Reset", func(t *testing.T) {
		input := "\x1b[31mHello World"
		got := Truncate(input, 5)

		if !strings.HasPrefix(got, "\x1b[31m") {
			t.Error("Lost initial color code")
		}
		if !strings.HasSuffix(got, "\x1b[0m") {
			t.Errorf("Expected ANSI reset at end, got: %q", got)
		}

		got = Truncate("\x1b[31mHi", 2)
		if !strings.HasSuffix(got, "\x1b[0m") {
			t.Error("Should append reset even on exact fit if color was active")
		}
	})

	t.Run("Truncate Suffix Logic", func(t *testing.T) {
		got := Truncate("Hello", 2, "...")
		if got != "" {
			t.Errorf("Expected empty string when suffix doesn't fit, got %q", got)
		}

		got = Truncate("Hello", 3, "...")
		if got != "..." {
			t.Errorf("Expected only suffix, got %q", got)
		}
	})

	t.Run("Global Cache Management", func(t *testing.T) {
		mu.Lock()
		origCache := widthCache
		mu.Unlock()
		defer func() {
			mu.Lock()
			widthCache = origCache
			mu.Unlock()
		}()

		SetCacheCapacity(0)
		size, cap, _ := GetCacheStats()
		if size != 0 || cap != 0 {
			t.Error("Cache should be disabled (stats 0,0)")
		}

		if w := Width("abc"); w != 3 {
			t.Errorf("Width('abc') = %d, want 3", w)
		}

		SetCacheCapacity(10)
		Width("ab")
		size, cap, _ = GetCacheStats()
		if size != 1 || cap != 10 {
			t.Errorf("Stats mismatch: size=%d, cap=%d", size, cap)
		}
	})

}

func TestBug_ForceNarrow_MultiChar(t *testing.T) {
	// Setup the specific condition causing the bug:
	// EastAsianWidth is ON (simulating RUNEWIDTH_EASTASIAN=1)
	// ForceNarrowBorders is ON (simulating Modern Terminal detection)
	buggyOptions := Options{
		EastAsianWidth:     true,
		ForceNarrowBorders: true,
	}
	SetOptions(buggyOptions)

	// Define a border string commonly used by the renderer (length > 1)
	// The current bug fails because it only checks len(r) == 1
	input := "─────" // 5 chars
	want := 5        // Should be 5 width (1 per char)

	// 3. Measure
	got := Width(input)

	// 4. Assert
	if got != want {
		t.Fatalf("Regression found: ForceNarrowBorders failed on multi-char string.\nInput: %q\nGot Width: %d (Likely Double Width)\nWant Width: %d", input, got, want)
	}
}
