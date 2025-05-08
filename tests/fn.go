package tests

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"testing"
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
