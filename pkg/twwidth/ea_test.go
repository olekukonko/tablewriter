package twwidth

import (
	"os"
	"testing"
)

func TestDetectEastAsian_Logic(t *testing.T) {
	// We cannot run this in parallel (t.Parallel()) because it modifies
	// process-level environment variables and global state.

	// Helper struct to define the environment state for a test case
	type envConfig struct {
		lcAll       string
		lcCtype     string
		lang        string
		runeWidth   string // RUNEWIDTH_EASTASIAN
		termProg    string // TERM_PROGRAM
		term        string // TERM
		forceLegacy bool   // Global legacy switch
	}

	tests := []struct {
		name     string
		env      envConfig
		expected bool
	}{

		// LOCALE ONLY (Legacy Behavior / Non-Modern Terminal)
		{
			name:     "Locale: Chinese (Simplified)",
			env:      envConfig{lang: "zh_CN.UTF-8"},
			expected: true,
		},
		{
			name:     "Locale: Japanese",
			env:      envConfig{lang: "ja_JP.UTF-8"},
			expected: true,
		},
		{
			name:     "Locale: English",
			env:      envConfig{lang: "en_US.UTF-8"},
			expected: false,
		},
		{
			name:     "Locale Priority: LC_ALL overrides LANG",
			env:      envConfig{lcAll: "ja_JP", lang: "en_US"},
			expected: true,
		},
		{
			name:     "Locale Region: English in Hong Kong (en_HK)",
			env:      envConfig{lang: "en_HK"},
			expected: true,
		},

		// MODERN ENVIRONMENT (Should force Narrow/False)
		{
			name:     "Modern: VSCode with Chinese Locale",
			env:      envConfig{lang: "zh_CN.UTF-8", termProg: "vscode"},
			expected: false, // Modern env implies single-width font
		},
		{
			name:     "Modern: iTerm2 with Japanese Locale",
			env:      envConfig{lang: "ja_JP.UTF-8", termProg: "iTerm.app"},
			expected: false,
		},
		{
			name: "Modern: Windows Terminal via WT_PROFILE_ID (Simulated by TERM checks in this test structure)",
			// Note: We can't easily mock runtime.GOOS, so we stick to env vars
			// that work cross-platform in the heuristic function.
			// Let's test Alacritty which is checked via env var.
			env:      envConfig{lang: "zh_CN.UTF-8", termProg: "Alacritty"},
			expected: false,
		},
		{
			name:     "Modern: TERM=xterm-kitty with Chinese Locale",
			env:      envConfig{lang: "zh_CN.UTF-8", term: "xterm-kitty"},
			expected: false,
		},

		// USER OVERRIDE (Highest Priority)
		{
			name:     "Override: Force ON (1) in English Env",
			env:      envConfig{lang: "en_US.UTF-8", runeWidth: "1"},
			expected: true,
		},
		{
			name:     "Override: Force ON (true) overrides Modern Env",
			env:      envConfig{lang: "en_US.UTF-8", termProg: "vscode", runeWidth: "true"},
			expected: true,
		},
		{
			name:     "Override: Force OFF (0) in Chinese Env",
			env:      envConfig{lang: "zh_CN.UTF-8", runeWidth: "0"},
			expected: false,
		},
		{
			name:     "Override: Force OFF (false) overrides Legacy Detection",
			env:      envConfig{lang: "zh_CN.UTF-8", runeWidth: "false"},
			expected: false,
		},

		// LEGACY FORCE SWITCH (Programmatic Override)
		{
			name:     "Legacy Force: Ignores Modern Env",
			env:      envConfig{lang: "zh_CN.UTF-8", termProg: "vscode", forceLegacy: true},
			expected: true, // Should use Locale (True) despite VSCode
		},
		{
			name:     "Legacy Force: Still respects User Override (RW=0)",
			env:      envConfig{lang: "zh_CN.UTF-8", runeWidth: "0", forceLegacy: true},
			expected: false, // User env var is supreme
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save current state
			saveEnv := func(key string) string { return os.Getenv(key) }
			oldLCAll := saveEnv(EnvLCAll)
			oldLCCtype := saveEnv(EnvLCCtype)
			oldLang := saveEnv(EnvLang)
			oldRuneWidth := saveEnv(EnvRuneWidthEastAsian)
			oldTermProg := saveEnv(EnvTermProgram)
			oldTerm := saveEnv(EnvTerm)

			// Helper to set/unset
			setEnv := func(key, val string) {
				if val == "" {
					os.Unsetenv(key)
				} else {
					os.Setenv(key, val)
				}
			}

			// Restore after test
			defer func() {
				setEnv(EnvLCAll, oldLCAll)
				setEnv(EnvLCCtype, oldLCCtype)
				setEnv(EnvLang, oldLang)
				setEnv(EnvRuneWidthEastAsian, oldRuneWidth)
				setEnv(EnvTermProgram, oldTermProg)
				setEnv(EnvTerm, oldTerm)
				EastAsianForceLegacy(false) // Reset global flag
			}()

			// Apply test configuration
			setEnv(EnvLCAll, tt.env.lcAll)
			setEnv(EnvLCCtype, tt.env.lcCtype)
			setEnv(EnvLang, tt.env.lang)
			setEnv(EnvRuneWidthEastAsian, tt.env.runeWidth)
			setEnv(EnvTermProgram, tt.env.termProg)
			setEnv(EnvTerm, tt.env.term)
			EastAsianForceLegacy(tt.env.forceLegacy)

			// Call internal logic directly to bypass sync.Once
			if got := detectEastAsian(); got != tt.expected {
				t.Errorf("detectEastAsian() = %v, want %v\nEnv: %+v", got, tt.expected, tt.env)
			}
		})
	}
}

func TestAutoUseEastAsian_Cache(t *testing.T) {
	// This test verifies that the result is cached (sync.Once).
	// Since 'eastAsianOnce' is private and global, we assume this is the
	// first call in this process execution, OR we implicitly accept checking
	// the behavior of the *existing* singleton state if unrelated tests ran before.

	// IMPORTANT: Because other tests might have run, we can't guarantee `eastAsianOnce`
	// is fresh. This test is best effort to ensure stability.

	// Snapshot Result
	firstResult := EastAsian()

	// Change Environmental factors to the OPPOSITE of what triggered firstResult
	if firstResult {
		// If currently True (e.g. CJK), try to force False
		os.Setenv(EnvLang, "en_US.UTF-8")
		os.Setenv(EnvRuneWidthEastAsian, "0")
	} else {
		// If currently False, try to force True
		os.Setenv(EnvLang, "zh_CN.UTF-8")
		os.Setenv(EnvRuneWidthEastAsian, "1")
	}
	// Ensure we clean up
	defer func() {
		os.Unsetenv(EnvLang)
		os.Unsetenv(EnvRuneWidthEastAsian)
	}()

	// Second call
	secondResult := EastAsian()

	if firstResult != secondResult {
		t.Errorf("EastAsian() did not cache result. First=%v, Second=%v", firstResult, secondResult)
	}
}
