package tw

import (
	"reflect"
	"testing"
)

func TestSplitCase(t *testing.T) {
	tests := []struct {
		input    string
		expected []string
	}{
		{
			input:    "",
			expected: []string{},
		},
		{
			input:    "snake_Case",
			expected: []string{"snake", "Case"},
		},
		{
			input:    "PascalCase",
			expected: []string{"Pascal", "Case"},
		},
		{
			input:    "camelCase",
			expected: []string{"camel", "Case"},
		},
		{
			input:    "_snake_CasePascalCase_camelCase123",
			expected: []string{"snake", "Case", "Pascal", "Case", "camel", "Case", "123"},
		},
		{
			input:    "ㅤ",
			expected: []string{"ㅤ"},
		},
		{
			input:    " \r\n\t",
			expected: []string{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if output := SplitCamelCase(tt.input); !reflect.DeepEqual(output, tt.expected) {
				t.Errorf("SplitCamelCase(%q) = %v, want %v", tt.input, output, tt.expected)
			}
		})
	}
}

func TestBreakPoint(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		limit    int
		expected int
	}{
		{
			name:     "empty string",
			input:    "",
			limit:    5,
			expected: 0,
		},
		{
			name:     "the whole string fits",
			input:    "hello",
			limit:    5,
			expected: 5,
		},
		{
			name:     "the string is longer than the limit",
			input:    "hello world",
			limit:    5,
			expected: 5,
		},
		{
			name:     "a limit of zero fits nothing",
			input:    "hello",
			limit:    0,
			expected: 0,
		},
		{
			name:     "an escape sequence takes no width",
			input:    "\x1b[31mabc",
			limit:    2,
			expected: 7,
		},
		{
			name:     "an escape sequence between two words",
			input:    "ab\x1b[0mcd",
			limit:    3,
			expected: 7,
		},
		{
			name:     "an OSC sequence takes no width",
			input:    "\x1b]8;;http://example.com\x1b\\link",
			limit:    2,
			expected: 27,
		},
		{
			name:     "a wide rune after an escape sequence",
			input:    "\x1b[31m漢",
			limit:    1,
			expected: 6,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if output := BreakPoint(tt.input, tt.limit); output != tt.expected {
				t.Errorf("BreakPoint(%q, %d) = %d, want %d", tt.input, tt.limit, output, tt.expected)
			}
		})
	}
}

func TestNormalizeNewLines(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{input: "", expected: ""},
		{input: "a\nb", expected: "a\nb"},
		{input: "a\r\nb\r\n", expected: "a\nb\n"},
		{input: "a\r\n\r\nb", expected: "a\n\nb"},
		{input: "a\rb", expected: "a\rb"},
		{input: "a\r\r\nb", expected: "a\r\nb"},
	}
	for _, tt := range tests {
		if output := NormalizeNewLines(tt.input); output != tt.expected {
			t.Errorf("NormalizeNewLines(%q) = %q, want %q", tt.input, output, tt.expected)
		}
	}
}
