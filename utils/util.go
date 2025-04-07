package utils

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

// ansi is a compiled regular expression used to filter out ANSI escape sequences.
var ansi = generateEscapeFilterRegex()

// generateEscapeFilterRegex generates a regular expression to filter out ANSI escape sequences.
// It returns a compiled regexp that matches both control sequences and operating system commands.
func generateEscapeFilterRegex() *regexp.Regexp {
	var regESC = "\x1b" // ASCII escape
	var regBEL = "\x07" // ASCII bell

	// String Terminator - ends ANSI sequences.
	var regST = "(" + regESC + "\\\\" + "|" + regBEL + ")"

	// Control Sequence Introducer - usually for color codes.
	// Matches: esc + [ + zero or more characters in [0x30-0x3f] + zero or more characters in [0x20-0x2f] + a single character in [0x40-0x7e].
	var regCSI = regESC + "\\[" + "[\x30-\x3f]*[\x20-\x2f]*[\x40-\x7e]"

	// Operating System Command - e.g., hyperlinks.
	// Matches: esc + ] + any characters (non-greedy) + string terminator.
	var regOSC = regESC + "\\]" + ".*?" + regST

	return regexp.MustCompile("(" + regCSI + "|" + regOSC + ")")
}

// RuneWidth returns the display width of a string by first removing any ANSI escape sequences.
// This is useful for calculating the true length of strings when formatting table outputs.
func RuneWidth(str string) int {
	return runewidth.StringWidth(ansi.ReplaceAllLiteralString(str, ""))
}

// ConditionOr returns the 'valid' string if the condition is true; otherwise, it returns 'inValid'.
// This is a simple helper to choose between two strings based on a boolean condition.
func ConditionOr(cond bool, valid, inValid string) string {
	if cond {
		return valid
	}
	return inValid
}

// Title formats a table header by replacing underscores and dots (where appropriate) with spaces,
// trimming whitespace, and converting the result to uppercase.
// If the resulting string is empty (but the original had characters), it returns a single space.
func Title(name string) string {
	origLen := len(name)
	rs := []rune(name)
	for i, r := range rs {
		switch r {
		case '_':
			rs[i] = ' '
		case '.':
			// Ignore a dot in a floating number (e.g., 0.0) if adjacent to numeric characters.
			if (i != 0 && !IsNumOrSpace(rs[i-1])) || (i != len(rs)-1 && !IsNumOrSpace(rs[i+1])) {
				rs[i] = ' '
			}
		}
	}
	name = string(rs)
	name = strings.TrimSpace(name)
	if len(name) == 0 && origLen > 0 {
		// Preserve at least one character for empty lines in multi-line headers/footers.
		name = " "
	}
	return strings.ToUpper(name)
}

// Pad centers the string 's' within a field of a given 'width' by padding it with the provided 'pad' string.
// If the string is shorter than 'width', the remaining space is split evenly (with left side receiving the extra space if needed).
func Pad(s, pad string, width int) string {
	gap := width - RuneWidth(s)
	if gap > 0 {
		gapLeft := int(math.Ceil(float64(gap) / 2))
		gapRight := gap - gapLeft
		return strings.Repeat(pad, gapLeft) + s + strings.Repeat(pad, gapRight)
	}
	return s
}

// PadRight pads the string 's' on the right with the 'pad' string until it reaches the specified 'width'.
// This effectively left-aligns the string.
func PadRight(s, pad string, width int) string {
	gap := width - RuneWidth(s)
	if gap > 0 {
		return s + strings.Repeat(pad, gap)
	}
	return s
}

// PadLeft pads the string 's' on the left with the 'pad' string until it reaches the specified 'width'.
// This effectively right-aligns the string.
func PadLeft(s, pad string, width int) string {
	gap := width - RuneWidth(s)
	if gap > 0 {
		return strings.Repeat(pad, gap) + s
	}
	return s
}

// IsNumOrSpace checks whether the rune 'r' is a numeric digit (0-9) or a space.
// It returns true if the rune meets these conditions.
func IsNumOrSpace(r rune) bool {
	return ('0' <= r && r <= '9') || r == ' '
}

// IsNumeric checks whether a string represents a numeric value.
// It returns true for valid integers or floating-point numbers, and false otherwise.
func IsNumeric(s string) bool {
	s = strings.TrimSpace(s)
	if s == "" {
		return false
	}
	// Check for plain integer numbers.
	_, err := strconv.Atoi(s)
	if err == nil {
		return true
	}
	// Check for floating-point numbers.
	_, err = strconv.ParseFloat(s, 64)
	return err == nil
}

// SplitCamelCase splits a camel case string into its constituent words.
// It handles transitions such as an uppercase letter followed by lowercase letters
// (e.g., "PDFLoader" becomes ["PDF", "Loader"]).
func SplitCamelCase(src string) (entries []string) {
	// Do not split if the string is invalid UTF-8.
	if !utf8.ValidString(src) {
		return []string{src}
	}
	entries = []string{}
	var runes [][]rune
	lastClass := 0
	class := 0
	// Split into fields based on the Unicode class of each character.
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
	// Handle transitions from uppercase sequences to lowercase (e.g., "PDFLoader").
	for i := 0; i < len(runes)-1; i++ {
		if unicode.IsUpper(runes[i][0]) && unicode.IsLower(runes[i+1][0]) {
			runes[i+1] = append([]rune{runes[i][len(runes[i])-1]}, runes[i+1]...)
			runes[i] = runes[i][:len(runes[i])-1]
		}
	}
	// Slice rune slices to strings and collect non-empty entries.
	for _, s := range runes {
		if len(s) > 0 {
			if strings.TrimSpace(string(s)) == "" {
				continue
			}
			entries = append(entries, string(s))
		}
	}
	return
}

// TruncateString truncates a string to fit within maxWidth while preserving ANSI codes
func TruncateString(s string, maxWidth int) string {
	if maxWidth <= 0 {
		return ""
	}

	// First strip ANSI codes to measure real content width
	stripped := ansi.ReplaceAllLiteralString(s, "")
	if runewidth.StringWidth(stripped) <= maxWidth {
		return s
	}

	// Preserve ANSI codes while truncating
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

	return buf.String()
}

// ConvertToSorted converts any map[int]int to a sorted slice of widths
func ConvertToSorted(m map[int]int) []int {
	// Get and sort the keys
	columns := make([]int, 0, len(m))
	for col := range m {
		columns = append(columns, col)
	}
	sort.Ints(columns)

	// Create the sorted result
	result := make([]int, 0, len(columns))
	for _, col := range columns {
		result = append(result, m[col])
	}
	return result
}

//// Example usage:
//func main() {
//	widths := make(Widths)
//	widths.Set(0, 10)
//	widths.Set(1, 20)
//	widths.Set(3, 15) // Column 2 is missing
//
//	// Convert to slice
//	slice := widths.Convert(4) // [10, 20, 0, 15]
//	fmt.Println(slice)
//
//	// Get max width
//	max := widths.Max() // 20
//	fmt.Println(max)
//
//	// Merge with another
//	other := make(Widths)
//	other.Set(1, 25)
//	other.Set(2, 5)
//	widths.Merge(other) // Now contains {0:10, 1:25, 2:5, 3:15}
//
//	// Check equality
//	fmt.Println(widths.Equal(other)) // false
//}
