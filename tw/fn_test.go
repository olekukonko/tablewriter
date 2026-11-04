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
