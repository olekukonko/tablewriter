package twwidth

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"testing"

	"github.com/mattn/go-runewidth"
)

// TestMain is the entry point for tests. It includes special logic to handle
// re-executing the test binary as a sub-process. This is the standard pattern
// for testing environment-dependent init() functions.
func TestMain(m *testing.M) {
	// Check if we are in the sub-process execution.
	if os.Getenv("GO_TEST_SUBPROCESS") == "1" {
		helperProcess()
		return
	}
	os.Exit(m.Run())
}

// helperProcess is run when the test is executed as a sub-process.
// It checks the initial state of IsEastAsian() (which is set by init())
// and prints it to stdout for the parent test process to capture.
func helperProcess() {
	if IsEastAsian() {
		fmt.Fprint(os.Stdout, "true")
	} else {
		fmt.Fprint(os.Stdout, "false")
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
	// Defer restoring the original state to ensure other tests are not affected.
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

func resetGlobalCache() {
	mu.Lock()
	widthCache = newLRUCache(cacheCapacity)
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

// TestInitRespectsEnvironment verifies that the package's init() function
// correctly reads environment variables like RUNEWIDTH_EASTASIAN.
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
			// Get the path to the current test binary.
			// This ensures the sub-process runs TestMain (which we need), but doesn't
			// re-run this same test, which would cause an infinite loop.
			cmd := exec.Command(os.Args[0], "-test.run=^TestHelperProcess$")

			// Set the environment for the sub-process.
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
