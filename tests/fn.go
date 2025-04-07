package tests

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
)

// visualCheck compares rendered output against exact expected lines
func visualCheck(t *testing.T, name string, output string, expected string) {
	t.Helper()

	// Normalize line endings and split into lines
	normalize := func(s string) []string {
		s = strings.ReplaceAll(s, "\r\n", "\n")
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

	// Compare line counts
	if len(outputLines) != len(expectedLines) {
		t.Errorf("%s: line count mismatch - expected %d, got %d", name, len(expectedLines), len(outputLines))
		t.Errorf("Expected:\n%s\n", strings.Join(expectedLines, "\n"))
		t.Errorf("Got:\n%s\n", strings.Join(outputLines, "\n"))
		return
	}

	// Compare each line
	type mismatch struct {
		Line     int    `json:"line"`
		Expected string `json:"expected"`
		Got      string `json:"got"`
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
	}
}

// Filter Presets
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

func MaskPassword(cells []string) []string {
	for i, cell := range cells {
		if len(cell) > 0 && (strings.Contains(strings.ToLower(cell), "pass") || len(cell) >= 8) {
			cells[i] = strings.Repeat("*", len(cell))
		}
	}
	return cells
}

func MaskCard(cells []string) []string {
	for i, cell := range cells {
		// Simple check for card-like numbers (16 digits or with dashes)
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
