package twwidth

import (
	"os"
	"testing"
)

func TestDetectEastAsian_Logic(t *testing.T) {
	// We cannot run this in parallel (t.Parallel()) because it modifies
	// process-level environment variables.

	tests := []struct {
		name     string
		lcAll    string // Sets LC_ALL
		lcCtype  string // Sets LC_CTYPE
		lang     string // Sets LANG
		expected bool
	}{
		// --- Basic Language Checks ---
		{"Chinese (Simplified)", "", "", "zh_CN.UTF-8", true},
		{"Chinese (Traditional)", "", "", "zh_TW", true},
		{"Japanese", "", "", "ja_JP.UTF-8", true},
		{"Korean", "", "", "ko_KR.UTF-8", true},
		{"English", "", "", "en_US.UTF-8", false},
		{"German", "", "", "de_DE.UTF-8", false},

		// --- Priority Order Checks (LC_ALL > LC_CTYPE > LANG) ---
		{
			name:     "LC_ALL overrides others",
			lcAll:    "ja_JP", // Should be true
			lcCtype:  "en_US",
			lang:     "en_US",
			expected: true,
		},
		{
			name:     "LC_CTYPE overrides LANG",
			lcAll:    "",
			lcCtype:  "zh_CN", // Should be true
			lang:     "en_US",
			expected: true,
		},
		{
			name:     "LANG is fallback",
			lcAll:    "",
			lcCtype:  "",
			lang:     "ko_KR", // Should be true
			expected: true,
		},
		{
			name:     "Non-CJK LC_ALL disables CJK LANG",
			lcAll:    "en_US", // Should be false
			lcCtype:  "zh_CN",
			lang:     "zh_CN",
			expected: false,
		},

		// --- Suffix Handling ---
		{"Strip Encoding", "", "", "ja_JP.EUC-JP", true},
		{"Strip Modifier", "", "", "zh_CN@currency=CNY", true},
		{"Strip Both", "", "", "zh_TW.UTF-8@modifier", true},

		// --- Region Checks (e.g. English users in East Asia) ---
		{"English in Hong Kong", "", "", "en_HK", true},
		{"English in Japan", "", "", "en_JP", true},
		{"English in Singapore", "", "", "en_SG.UTF-8", true}, // 'sg' was added in previous steps
		{"English in US", "", "", "en_US", false},

		// --- Edge Cases ---
		{"Empty", "", "", "", false},
		{"C Locale", "", "", "C", false},
		{"POSIX Locale", "", "", "POSIX", false},
		{"Case Insensitive Input", "", "", "ZH_cn.utf-8", true},
		{"Short string", "", "", "z", false},
		{"Just language", "", "", "ja", true},
		{"ISO 3-letter (chi)", "", "", "chi_CN", true},
		{"Full name (japanese)", "", "", "japanese", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Helper to save and restore env vars
			saveEnv := func(key string) string { return os.Getenv(key) }
			restoreEnv := func(key, val string) {
				if val == "" {
					os.Unsetenv(key)
				} else {
					os.Setenv(key, val)
				}
			}

			oldLCAll := saveEnv("LC_ALL")
			oldLCCtype := saveEnv("LC_CTYPE")
			oldLang := saveEnv("LANG")

			// Cleanup after test
			defer func() {
				restoreEnv("LC_ALL", oldLCAll)
				restoreEnv("LC_CTYPE", oldLCCtype)
				restoreEnv("LANG", oldLang)
			}()

			// Set Environment for test
			if tt.lcAll != "" {
				os.Setenv("LC_ALL", tt.lcAll)
			} else {
				os.Unsetenv("LC_ALL")
			}

			if tt.lcCtype != "" {
				os.Setenv("LC_CTYPE", tt.lcCtype)
			} else {
				os.Unsetenv("LC_CTYPE")
			}

			if tt.lang != "" {
				os.Setenv("LANG", tt.lang)
			} else {
				os.Unsetenv("LANG")
			}

			// Call internal logic directly to bypass sync.Once
			if got := detectEastAsian(); got != tt.expected {
				t.Errorf("detectEastAsian() = %v, want %v (LC_ALL=%q LC_CTYPE=%q LANG=%q)",
					got, tt.expected, tt.lcAll, tt.lcCtype, tt.lang)
			}
		})
	}
}

func TestAutoUseEastAsian_Cache(t *testing.T) {
	// This test verifies that the result is cached (sync.Once).
	// NOTE: This test might conflict if AutoUseEastAsian was called by other tests
	// in the same package run. Ideally, reset 'eastAsianOnce' or run separately.
	// Since 'eastAsianOnce' is private, we assume this is the first call in this process execution
	// or we accept we are testing the behavior of the *first* call logic.

	// Set Environment to CJK
	os.Setenv("LANG", "ja_JP.UTF-8")
	defer os.Unsetenv("LANG")

	// Since we can't easily reset sync.Once in a black-box test,
	// we will verify that changing the Env Var AFTER the first call
	// DOES NOT change the result.

	initial := AutoUseEastAsian() // Calculates based on ja_JP -> True

	// Change Environment to US
	os.Setenv("LANG", "en_US.UTF-8")

	// Second call: Should still be what the first call was (True)
	// because logic is cached.
	cached := AutoUseEastAsian()

	if initial != cached {
		t.Errorf("AutoUseEastAsian() did not cache result. Got %v, expected %v", cached, initial)
	}
}
