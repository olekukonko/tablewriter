package twfn

import (
	"bytes"
	"math"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/mattn/go-runewidth"
)

//
// ────────────────────────────────────────────────────────────
//  ANSI FILTERING AND TERMINAL WIDTH UTILITIES
// ────────────────────────────────────────────────────────────
//

// ansi is a compiled regex pattern used to strip ANSI escape codes.
// These codes are used in terminal output for styling and are invisible in rendered text.
var ansi = CompileANSIFilter()

// CompileANSIFilter constructs and compiles a regex for matching ANSI sequences.
// It supports both control sequences and operating system commands like hyperlinks.
func CompileANSIFilter() *regexp.Regexp {
	var regESC = "\x1b" // ASCII escape character
	var regBEL = "\x07" // ASCII bell character

	var regST = "(" + regESC + "\\\\" + "|" + regBEL + ")"              // ANSI string terminator
	var regCSI = regESC + "\\[" + "[\x30-\x3f]*[\x20-\x2f]*[\x40-\x7e]" // Control codes
	var regOSC = regESC + "\\]" + ".*?" + regST                         // OSC codes like hyperlinks

	return regexp.MustCompile("(" + regCSI + "|" + regOSC + ")")
}

// DisplayWidth calculates the visual width of a string.
// ANSI escape sequences are stripped before measurement for accuracy.
func DisplayWidth(str string) int {
	return runewidth.StringWidth(ansi.ReplaceAllLiteralString(str, ""))
}

// TruncateString shortens a string to a max width while preserving ANSI color codes.
// An optional suffix (like "...") can be appended to indicate truncation.
func TruncateString(s string, maxWidth int, suffix ...string) string {
	if maxWidth <= 0 {
		return ""
	}

	stripped := ansi.ReplaceAllLiteralString(s, "")
	if runewidth.StringWidth(stripped) <= maxWidth {
		return s
	}

	var buf bytes.Buffer
	var currentWidth int
	ansiEnabled := false

	for _, r := range s {
		if r == '\x1b' {
			ansiEnabled = true
		}
		buf.WriteRune(r)

		if !ansiEnabled {
			currentWidth += runewidth.RuneWidth(r)
			if currentWidth >= maxWidth {
				break
			}
		}

		if ansiEnabled && r == 'm' {
			ansiEnabled = false
		}
	}

	buf.WriteString(strings.Join(suffix, " "))
	return buf.String()
}

//
// ────────────────────────────────────────────────────────────
//  STRING FORMATTING AND ALIGNMENT
// ────────────────────────────────────────────────────────────
//

// Title normalizes and uppercases a label string for use in headers.
// It replaces underscores and select dots with spaces, trimming whitespace.
func Title(name string) string {
	origLen := len(name)
	rs := []rune(name)
	for i, r := range rs {
		switch r {
		case '_':
			rs[i] = ' '
		case '.':
			if (i != 0 && !IsNumOrSpace(rs[i-1])) || (i != len(rs)-1 && !IsNumOrSpace(rs[i+1])) {
				rs[i] = ' '
			}
		}
	}
	name = string(rs)
	name = strings.TrimSpace(name)
	if len(name) == 0 && origLen > 0 {
		name = " "
	}
	return strings.ToUpper(name)
}

// PadCenter centers the input string within a fixed width using the pad character.
// If the string is smaller, extra padding is split between left and right.
func PadCenter(s, pad string, width int) string {
	gap := width - DisplayWidth(s)
	if gap > 0 {
		gapLeft := int(math.Ceil(float64(gap) / 2))
		gapRight := gap - gapLeft
		return strings.Repeat(pad, gapLeft) + s + strings.Repeat(pad, gapRight)
	}
	return s
}

// PadRight left-aligns the string within the specified width.
// The remaining space on the right is filled using the pad string.
func PadRight(s, pad string, width int) string {
	gap := width - DisplayWidth(s)
	if gap > 0 {
		return s + strings.Repeat(pad, gap)
	}
	return s
}

// PadLeft right-aligns the string within the specified width.
// The remaining space on the left is filled using the pad string.
func PadLeft(s, pad string, width int) string {
	gap := width - DisplayWidth(s)
	if gap > 0 {
		return strings.Repeat(pad, gap) + s
	}
	return s
}

//
// ────────────────────────────────────────────────────────────
//  STRING AND CHARACTER UTILITIES
// ────────────────────────────────────────────────────────────
//

// IsNumOrSpace checks if a rune is a digit or space character.
// It is used for safely replacing characters in formatting logic.
func IsNumOrSpace(r rune) bool {
	return ('0' <= r && r <= '9') || r == ' '
}

// IsNumeric returns true if a string represents a valid number.
// It supports both integers and floating-point values.
func IsNumeric(s string) bool {
	s = strings.TrimSpace(s)
	if s == "" {
		return false
	}
	if _, err := strconv.Atoi(s); err == nil {
		return true
	}
	_, err := strconv.ParseFloat(s, 64)
	return err == nil
}

// SplitCamelCase breaks a camelCase or PascalCase string into word segments.
// It handles transitions between uppercase and lowercase characters.
func SplitCamelCase(src string) (entries []string) {
	if !utf8.ValidString(src) {
		return []string{src}
	}
	entries = []string{}
	var runes [][]rune
	lastClass := 0
	class := 0
	for _, r := range src {
		switch {
		case unicode.IsLower(r):
			class = 1
		case unicode.IsUpper(r):
			class = 2
		case unicode.IsDigit(r):
			class = 3
		default:
			class = 4
		}
		if class == lastClass {
			runes[len(runes)-1] = append(runes[len(runes)-1], r)
		} else {
			runes = append(runes, []rune{r})
		}
		lastClass = class
	}
	for i := 0; i < len(runes)-1; i++ {
		if unicode.IsUpper(runes[i][0]) && unicode.IsLower(runes[i+1][0]) {
			runes[i+1] = append([]rune{runes[i][len(runes[i])-1]}, runes[i+1]...)
			runes[i] = runes[i][:len(runes[i])-1]
		}
	}
	for _, s := range runes {
		if len(s) > 0 && strings.TrimSpace(string(s)) != "" {
			entries = append(entries, string(s))
		}
	}
	return
}

//
// ────────────────────────────────────────────────────────────
//  MAP TRANSFORMATION UTILITIES
// ────────────────────────────────────────────────────────────
//

// ConvertToSorted returns a sorted slice of map values by key order.
// It is useful for converting maps into ordered table structures.
func ConvertToSorted(m map[int]int) []int {
	columns := make([]int, 0, len(m))
	for col := range m {
		columns = append(columns, col)
	}
	sort.Ints(columns)

	result := make([]int, 0, len(columns))
	for _, col := range columns {
		result = append(result, m[col])
	}
	return result
}

// ConvertToSortedKeys returns sorted integer keys of a generic map.
// This helps when iterating over maps in a consistent order.
func ConvertToSortedKeys[V any](m map[int]V) []int {
	keys := make([]int, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Ints(keys)
	return keys
}

//
// ────────────────────────────────────────────────────────────
//  MISCELLANEOUS UTILITIES
// ────────────────────────────────────────────────────────────
//

// Or returns 'valid' if cond is true; otherwise returns 'inValid'.
// It simplifies ternary-like decisions for string output.
func Or(cond bool, valid, inValid string) string {
	if cond {
		return valid
	}
	return inValid
}

// Max returns the greater of two integer values.
// Simple helper for comparison logic.
func Max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// MapKeys GetMapKeys returns a slice containing all keys from the input map
func MapKeys[K comparable, V any](m map[K]V) []K {
	keys := make([]K, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
