package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"testing"
	"unicode"
)

// mismatch represents a discrepancy between expected and actual output lines in a test.
type mismatch struct {
	Line     int    `json:"line"`     // Line number (1-based)
	Expected string `json:"expected"` // Expected line content and length
	Got      string `json:"got"`      // Actual line content and length
}

// MaskEmail masks email addresses in a slice of strings, replacing all but the first character of the local part with asterisks.
func MaskEmail(cells []string) []string {
	for i, cell := range cells {
		if strings.Contains(cell, "@") {
			parts := strings.Split(cell, "@")
			if len(parts) == 2 {
				masked := parts[0][:1] + strings.Repeat("*", len(parts[0])-1) + "@" + parts[1]
				cells[i] = masked
			}
		}
	}
	return cells
}

// MaskPassword masks strings that resemble passwords (containing "pass" or 8+ characters) with asterisks.
func MaskPassword(cells []string) []string {
	for i, cell := range cells {
		if len(cell) > 0 && (strings.Contains(strings.ToLower(cell), "pass") || len(cell) >= 8) {
			cells[i] = strings.Repeat("*", len(cell))
		}
	}
	return cells
}

// MaskCard masks credit card-like numbers, keeping only the last four digits visible.
func MaskCard(cells []string) []string {
	for i, cell := range cells {
		// Check for card-like numbers (12+ digits, with or without dashes/spaces)
		if len(cell) >= 12 && (strings.Contains(cell, "-") || len(strings.ReplaceAll(cell, " ", "")) >= 12) {
			parts := strings.FieldsFunc(cell, func(r rune) bool { return r == '-' || r == ' ' })
			masked := ""
			for j, part := range parts {
				if j < len(parts)-1 {
					masked += strings.Repeat("*", len(part))
				} else {
					masked += part // Keep last 4 digits visible
				}
				if j < len(parts)-1 {
					masked += "-"
				}
			}
			cells[i] = masked
		}
	}
	return cells
}

// visualCheck compares rendered output against expected lines, reporting mismatches in a test.
// It normalizes line endings, strips ANSI colors, and trims empty lines before comparison.
func visualCheck(t *testing.T, name string, output string, expected string) bool {
	t.Helper()

	// Normalize line endings and split into lines
	normalize := func(s string) []string {
		s = strings.ReplaceAll(s, "\r\n", "\n")
		s = StripColors(s)
		return strings.Split(s, "\n")
	}

	expectedLines := normalize(expected)
	outputLines := normalize(output)

	// Trim empty lines from start and end
	trimEmpty := func(lines []string) []string {
		start, end := 0, len(lines)
		for start < end && strings.TrimSpace(lines[start]) == "" {
			start++
		}
		for end > start && strings.TrimSpace(lines[end-1]) == "" {
			end--
		}
		return lines[start:end]
	}

	expectedLines = trimEmpty(expectedLines)
	outputLines = trimEmpty(outputLines)

	// Check line counts
	if len(outputLines) != len(expectedLines) {
		ex := strings.Join(expectedLines, "\n")
		ot := strings.Join(outputLines, "\n")
		t.Errorf("%s: line count mismatch - expected %d, got %d", name, len(expectedLines), len(outputLines))
		t.Errorf("Expected:\n%s\n", ex)
		t.Errorf("Got:\n%s\n", ot)
		return false
	}

	var mismatches []mismatch
	for i := 0; i < len(expectedLines) && i < len(outputLines); i++ {
		exp := strings.TrimSpace(expectedLines[i])
		got := strings.TrimSpace(outputLines[i])
		if exp != got {
			mismatches = append(mismatches, mismatch{
				Line:     i + 1,
				Expected: fmt.Sprintf("%s (%d)", exp, len(exp)),
				Got:      fmt.Sprintf("%s (%d)", got, len(got)),
			})
		}
	}

	// Report mismatches
	if len(mismatches) > 0 {
		diff, _ := json.MarshalIndent(mismatches, "", "  ")
		t.Errorf("%s: %d mismatches found:\n%s", name, len(mismatches), diff)
		t.Errorf("Full expected output:\n%s", expected)
		t.Errorf("Full actual output:\n%s", output)
		return false
	}

	return true
}

// visualCheckHTML compares rendered HTML output against expected lines,
// trimming whitespace per line and ignoring blank lines.
func visualCheckHTML(t *testing.T, name string, output string, expected string) bool {
	t.Helper()

	normalizeHTML := func(s string) []string {
		s = strings.ReplaceAll(s, "\r\n", "\n") // Normalize line endings
		lines := strings.Split(s, "\n")
		trimmedLines := make([]string, 0, len(lines))
		for _, line := range lines {
			trimmed := strings.TrimSpace(line)
			if trimmed != "" { // Only keep non-blank lines
				trimmedLines = append(trimmedLines, trimmed)
			}
		}
		return trimmedLines
	}

	expectedLines := normalizeHTML(expected)
	outputLines := normalizeHTML(output)

	// Compare line counts
	if len(outputLines) != len(expectedLines) {
		t.Errorf("%s: line count mismatch - expected %d, got %d", name, len(expectedLines), len(outputLines))
		t.Errorf("Expected (trimmed):\n%s", strings.Join(expectedLines, "\n"))
		t.Errorf("Got (trimmed):\n%s", strings.Join(outputLines, "\n"))
		// Optionally print full untrimmed for debugging exact whitespace
		// t.Errorf("Full Expected:\n%s", expected)
		// t.Errorf("Full Got:\n%s", output)
		return false
	}

	// Compare each line
	mismatches := []mismatch{} // Use mismatch struct from fn.go
	for i := 0; i < len(expectedLines); i++ {
		if expectedLines[i] != outputLines[i] {
			mismatches = append(mismatches, mismatch{
				Line:     i + 1,
				Expected: expectedLines[i],
				Got:      outputLines[i],
			})
		}
	}

	if len(mismatches) > 0 {
		t.Errorf("%s: %d mismatches found:", name, len(mismatches))
		for _, mm := range mismatches {
			t.Errorf("  Line %d:\n    Expected: %s\n    Got:      %s", mm.Line, mm.Expected, mm.Got)
		}
		// Optionally print full outputs again on mismatch
		// t.Errorf("Full Expected:\n%s", expected)
		// t.Errorf("Full Got:\n%s", output)
		return false
	}

	return true
}

// ansiColorRegex matches ANSI color escape sequences.
var ansiColorRegex = regexp.MustCompile(`\x1b\[[0-9;]*m`)

// StripColors removes ANSI color codes from a string.
func StripColors(s string) string {
	return ansiColorRegex.ReplaceAllString(s, "")
}

// Regex to remove leading/trailing whitespace from lines AND blank lines for HTML comparison
var htmlWhitespaceRegex = regexp.MustCompile(`(?m)^\s+|\s+$`)
var blankLineRegex = regexp.MustCompile(`(?m)^\s*\n`)

func normalizeHTMLStrict(s string) string {
	s = strings.ReplaceAll(s, "\r\n", "\n")
	s = strings.ReplaceAll(s, "\n", "") // Remove all newlines
	s = strings.ReplaceAll(s, "\t", "") // Remove tabs

	// Remove spaces after > and before <, effectively compacting tags
	s = regexp.MustCompile(`>\s+`).ReplaceAllString(s, ">")
	s = regexp.MustCompile(`\s+<`).ReplaceAllString(s, "<")

	// Trim overall leading/trailing space that might be left.
	return strings.TrimSpace(s)
}

// visualCheckCaption (helper function, potentially shared or adapted from your existing visualCheck)
// Ensure this helper normalizes expected and got strings for reliable comparison
// (e.g., trim spaces from each line, normalize newlines)
func visualCheckCaption(t *testing.T, testName, got, expected string) bool {
	t.Helper()
	normalize := func(s string) string {
		s = strings.ReplaceAll(s, "\r\n", "\n") // Normalize newlines
		lines := strings.Split(s, "\n")
		var trimmedLines []string
		for _, l := range lines {
			trimmedLines = append(trimmedLines, strings.TrimSpace(l))
		}
		// Join, then trim overall to handle cases where expected might have leading/trailing blank lines
		// but individual lines should keep their relative structure.
		return strings.TrimSpace(strings.Join(trimmedLines, "\n"))
	}

	gotNormalized := normalize(got)
	expectedNormalized := normalize(expected)

	if gotNormalized != expectedNormalized {
		// Use a more detailed diff output if available, or just print both.
		t.Errorf("%s: outputs do not match.\nExpected:\n```\n%s\n```\nGot:\n```\n%s\n```\n---Diff---\n%s",
			testName, expected, got, getDiff(expectedNormalized, gotNormalized)) // You might need a diff utility
		return false
	}
	return true
}

// A simple diff helper (replace with a proper library if needed)
func getDiff(expected, actual string) string {
	expectedLines := strings.Split(expected, "\n")
	actualLines := strings.Split(actual, "\n")
	maxLen := len(expectedLines)
	if len(actualLines) > maxLen {
		maxLen = len(actualLines)
	}
	var diff strings.Builder
	diff.WriteString("Line | Expected                         | Actual\n")
	diff.WriteString("-----|----------------------------------|----------------------------------\n")
	for i := 0; i < maxLen; i++ {
		eLine := ""
		if i < len(expectedLines) {
			eLine = expectedLines[i]
		}
		aLine := ""
		if i < len(actualLines) {
			aLine = actualLines[i]
		}
		marker := " "
		if eLine != aLine {
			marker = "!"
		}
		diff.WriteString(fmt.Sprintf("%4d %s| %-32s | %-32s\n", i+1, marker, eLine, aLine))
	}
	return diff.String()
}

func getLastContentLine(buf *bytes.Buffer) string {
	content := buf.String()
	lines := strings.Split(content, "\n")

	// Search backwards for first non-border, non-empty line
	for i := len(lines) - 1; i >= 0; i-- {
		line := strings.TrimSpace(lines[i])
		if line == "" || strings.Contains(line, "─") ||
			strings.Contains(line, "┌") || strings.Contains(line, "└") {
			continue
		}
		return line
	}
	return ""
}

type Name struct {
	First string
	Last  string
}

// this will be ignored since  Format() is present
func (n Name) String() string {
	return fmt.Sprintf("%s %s", n.First, n.Last)
}

// Note: Format() overrides String() if both exist.
func (n Name) Format() string {
	return fmt.Sprintf("%s %s", clean(n.First), clean(n.Last))
}

// clean ensures the first letter is capitalized and the rest are lowercase
func clean(s string) string {
	s = strings.TrimSpace(strings.ToLower(s))
	words := strings.Fields(s)
	s = strings.Join(words, "")

	if s == "" {
		return s
	}
	// Capitalize the first letter
	runes := []rune(s)
	runes[0] = unicode.ToUpper(runes[0])
	return string(runes)
}
