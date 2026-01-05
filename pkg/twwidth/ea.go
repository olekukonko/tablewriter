package twwidth

import (
	"os"
	"strings"
	"sync"
)

// Environment Variable Constants
const (
	EnvLCAll   = "LC_ALL"
	EnvLCCtype = "LC_CTYPE"
	EnvLang    = "LANG"
)

// CJK Language Codes (Prefixes)
// Covers ISO 639-1 (2-letter) and common full names used in some systems.
var cjkPrefixes = []string{
	"zh", "ja", "ko", // Standard: Chinese, Japanese, Korean
	"chi", "zho", // ISO 639-2/B and T for Chinese
	"jpn", "kor", // ISO 639-2 for Japanese, Korean
	"chinese", "japanese", "korean", // Full names (rare but possible in some legacy systems)
}

// CJK Region Codes
// Checks for specific regions that imply CJK font usage (e.g., en_HK).
var cjkRegions = map[string]bool{
	"cn": true, // China
	"tw": true, // Taiwan
	"hk": true, // Hong Kong
	"mo": true, // Macau
	"jp": true, // Japan
	"kr": true, // South Korea
	"kp": true, // North Korea
	"sg": true, // Singapore (Often uses CJK fonts)
}

var (
	eastAsianOnce sync.Once
	eastAsianVal  bool
)

// AutoUseEastAsian checks the environment variables to determine if
// East Asian width calculations should be enabled.
// The result is cached after the first call.
func AutoUseEastAsian() bool {
	eastAsianOnce.Do(func() {
		eastAsianVal = detectEastAsian()
	})
	return eastAsianVal
}

func detectEastAsian() bool {
	// Check Env Vars in POSIX priority order
	var locale string
	if loc := os.Getenv(EnvLCAll); loc != "" {
		locale = loc
	} else if loc := os.Getenv(EnvLCCtype); loc != "" {
		locale = loc
	} else if loc := os.Getenv(EnvLang); loc != "" {
		locale = loc
	}

	// Fast fail for empty or standard C/POSIX locales
	if locale == "" || locale == "C" || locale == "POSIX" {
		return false
	}

	// Normalize the string
	// Remove encoding suffix (e.g., ".UTF-8")
	if idx := strings.IndexByte(locale, '.'); idx != -1 {
		locale = locale[:idx]
	}
	// Remove modifiers (e.g., "@currency=CNY", "@euro")
	if idx := strings.IndexByte(locale, '@'); idx != -1 {
		locale = locale[:idx]
	}

	// Lowercase for table lookups
	locale = strings.ToLower(locale)

	// Check Language Prefix (e.g., "zh_CN" -> checks "zh")
	for _, prefix := range cjkPrefixes {
		if strings.HasPrefix(locale, prefix) {
			return true
		}
	}

	// Check Regions (e.g., "en_HK")
	// Split by underscore to handle formats like "lang_REGION" or "lang_Script_REGION"
	parts := strings.Split(locale, "_")
	if len(parts) > 1 {
		// Iterate through parts to find a matching region.
		// Usually the region is the last part, or the second part.
		for _, part := range parts[1:] {
			if cjkRegions[part] {
				return true
			}
		}
	}
	return false
}
