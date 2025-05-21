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
