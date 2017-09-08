package tablewriter

import (
	"fmt"
	"strconv"
	"strings"
)

const ESC = "\033"
const SEP = ";"

const (
	BgBlackColor int = iota + 40
	BgRedColor
	BgGreenColor
	BgYellowColor
	BgBlueColor
	BgMagentaColor
	BgCyanColor
	BgWhiteColor
)

const (
	FgBlackColor int = iota + 30
	FgRedColor
	FgGreenColor
	FgYellowColor
	FgBlueColor
	FgMagentaColor
	FgCyanColor
	FgWhiteColor
)

const (
	BgHiBlackColor int = iota + 100
	BgHiRedColor
	BgHiGreenColor
	BgHiYellowColor
	BgHiBlueColor
	BgHiMagentaColor
	BgHiCyanColor
	BgHiWhiteColor
)

const (
	FgHiBlackColor int = iota + 90
	FgHiRedColor
	FgHiGreenColor
	FgHiYellowColor
	FgHiBlueColor
	FgHiMagentaColor
	FgHiCyanColor
	FgHiWhiteColor
)

const (
	Normal          = 0
	Bold            = 1
	UnderlineSingle = 4
	Italic
)

type Values []int

func startFormat(seq string) string {
	return fmt.Sprintf("%s[%sm", ESC, seq)
}

func stopFormat() string {
	return fmt.Sprintf("%s[%dm", ESC, Normal)
}

// Making the SGR (Select Graphic Rendition) sequence.
func makeSequence(codes []int) string {
	codesInString := []string{}
	for _, code := range codes {
		codesInString = append(codesInString, strconv.Itoa(code))
	}
	return strings.Join(codesInString, SEP)
}

// Adding ANSI escape  sequences before and after string
func format(s string, codes interface{}) string {
	var seq string

	switch v := codes.(type) {

	case string:
		seq = v
	case []int:
		seq = makeSequence(v)
	default:
		return s
	}

	if len(seq) == 0 {
		return s
	}
	return startFormat(seq) + s + stopFormat()
}

// Adding header colors (ANSI codes)
func (t *Table) SetHeaderColor(args ...Values) {
	if t.colSize != len(args) {
		panic("Number of header colors must be equal to number of headers.")
	}
	for i := 0; i < len(args); i++ {
		t.headerParams = append(t.headerParams, makeSequence(args[i]))
	}
}

// Adding column colors (ANSI codes)
func (t *Table) SetColumnColor(args ...Values) {
	if t.colSize != len(args) {
		panic("Number of column colors must be equal to number of headers.")
	}
	for i := 0; i < len(args); i++ {
		t.columnsParams = append(t.columnsParams, makeSequence(args[i]))
	}
}

// Adding column colors (ANSI codes)
func (t *Table) SetFooterColor(args ...Values) {
	if len(t.footers) != len(args) {
		panic("Number of footer colors must be equal to number of footer.")
	}
	for i := 0; i < len(args); i++ {
		t.footerParams = append(t.footerParams, makeSequence(args[i]))
	}
}

func Add(values ...int) []int {
	return values
}
