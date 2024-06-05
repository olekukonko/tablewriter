package tablewriter

import (
	"strings"
	"testing"
)

func TestCleanHyperlinksInTerminalEmulators(t *testing.T) {
	testInput := "\033]8;;http://example.com\033\\This is a link\033]8;;\033\\\n"
	expectedOutput := "This is a link\n"
	actualOutput := CleanHyperlinksInTerminalEmulators(testInput)
	if actualOutput != expectedOutput {
		t.Errorf("Expected %s, got %s", expectedOutput, actualOutput)
	}
	expectedOutputSize := len(expectedOutput)
	actualOutputSize := len(actualOutput)
	if actualOutputSize != expectedOutputSize {
		t.Errorf("Expected size %d, got size %d", expectedOutputSize, actualOutputSize)
	}
	testInput2 := strings.Repeat(testInput, 10)
	expectedOutput2 := strings.Repeat(expectedOutput, 10)
	actualOutput2 := CleanHyperlinksInTerminalEmulators(testInput2)
	if actualOutput2 != expectedOutput2 {
		t.Errorf("Expected %s, got %s", expectedOutput2, actualOutput2)
	}
	expectedOutputSize2 := len(expectedOutput2)
	actualOutputSize2 := len(actualOutput2)
	if actualOutputSize2 != expectedOutputSize2 {
		t.Errorf("Expected size %d, got size %d", expectedOutputSize2, actualOutputSize2)
	}
}
