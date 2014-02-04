package table

import (
	"math"
	"strings"
	"unicode/utf8"
)

func Pad(s, pad string, width int) string {
	gap := width - utf8.RuneCountInString(s)
	if gap > 0 {
		gapLeft := int(math.Ceil(float64(gap / 2)))
		gapRight := gap - gapLeft
		return strings.Repeat(string(pad), gapLeft) + s + strings.Repeat(string(pad), gapRight)
	}
	return s
}

func PadRight(s, pad string, width int) string {
	gap := width - utf8.RuneCountInString(s)
	if gap > 0 {
		return s + strings.Repeat(string(pad), gap)
	}
	return s
}

func PadLeft(s, pad string, width int) string {
	gap := width - utf8.RuneCountInString(s)
	if gap > 0 {
		return strings.Repeat(string(pad), gap) + s
	}
	return s
}
