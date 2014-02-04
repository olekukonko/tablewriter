package table

import (
	"testing"
	"strings"
)

var text = "The quick brown fox jumps over the lazy dog."

func TestWrap(t *testing.T) {
	exp := []string{
		"The", "quick", "brown", "fox",
		"jumps", "over", "the", "lazy", "dog."}


	got , _ := WrapString(text, 6)
	if len(exp) != len(got) {
		t.Fail()
	}
}

func TestWrapOneLine(t *testing.T) {
	exp := "The quick brown fox jumps over the lazy dog."
	words , _ := WrapString(text, 500)
	got := strings.Join(words, string(sp))
	if exp != got {
		t.Fail()
	}
}
