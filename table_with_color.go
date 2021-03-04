package tablewriter

import (
	"fmt"
	"strconv"
	"strings"
)

// ESC escape symbol
const ESC = "\033"

// SEP separate symbol
const SEP = ";"

const (
	// BgBlackColor black background color
	BgBlackColor int = iota + 40
	// BgRedColor red background color
	BgRedColor
	// BgGreenColor green background color
	BgGreenColor
	// BgYellowColor yellow background color
	BgYellowColor
	// BgBlueColor blue background color
	BgBlueColor
	// BgMagentaColor magenta background color
	BgMagentaColor
	// BgCyanColor cyan background color
	BgCyanColor
	// BgWhiteColor white background color
	BgWhiteColor
)

const (
	// FgBlackColor black foreground color
	FgBlackColor int = iota + 30
	// FgRedColor red foreground color
	FgRedColor
	// FgGreenColor green foreground color
	FgGreenColor
	// FgYellowColor yellow foreground color
	FgYellowColor
	// FgBlueColor blue foreground color
	FgBlueColor
	// FgMagentaColor magenta foreground color
	FgMagentaColor
	// FgCyanColor cyan foreground color
	FgCyanColor
	// FgWhiteColor white foreground color
	FgWhiteColor
)

const (
	// BgHiBlackColor black bright background color
	BgHiBlackColor int = iota + 100
	// BgHiRedColor red bright background color
	BgHiRedColor
	// BgHiGreenColor green bright background color
	BgHiGreenColor
	// BgHiYellowColor yellow bright background color
	BgHiYellowColor
	// BgHiBlueColor blue bright background color
	BgHiBlueColor
	// BgHiMagentaColor magenta bright background color
	BgHiMagentaColor
	// BgHiCyanColor cyan bright background color
	BgHiCyanColor
	// BgHiWhiteColor white bright background color
	BgHiWhiteColor
)

const (
	// FgHiBlackColor black bright foreground color
	FgHiBlackColor int = iota + 90
	// FgHiRedColor red bright foreground color
	FgHiRedColor
	// FgHiGreenColor green bright foreground color
	FgHiGreenColor
	// FgHiYellowColor yellow bright foreground color
	FgHiYellowColor
	// FgHiBlueColor blue bright foreground color
	FgHiBlueColor
	// FgHiMagentaColor magenta bright foreground color
	FgHiMagentaColor
	// FgHiCyanColor cyan bright foreground color
	FgHiCyanColor
	// FgHiWhiteColor white bright foreground color
	FgHiWhiteColor
)

const (
	// Normal normal style
	Normal = 0
	// Bold bold style
	Bold = 1
	// UnderlineSingle underline style
	UnderlineSingle = 4
	// Italic italic style
	Italic
)

// Colors contains colors
type Colors []int

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
	case Colors:
		seq = makeSequence(v)
	default:
		return s
	}

	if len(seq) == 0 {
		return s
	}
	return startFormat(seq) + s + stopFormat()
}

// SetHeaderColor adding header colors (ANSI codes)
func (t *Table) SetHeaderColor(colors ...Colors) {
	if t.colSize != len(colors) {
		panic("Number of header colors must be equal to number of headers.")
	}
	for i := 0; i < len(colors); i++ {
		t.headerParams = append(t.headerParams, makeSequence(colors[i]))
	}
}

// SetColumnColor adding column colors (ANSI codes)
func (t *Table) SetColumnColor(colors ...Colors) {
	if t.colSize != len(colors) {
		panic("Number of column colors must be equal to number of headers.")
	}
	for i := 0; i < len(colors); i++ {
		t.columnsParams = append(t.columnsParams, makeSequence(colors[i]))
	}
}

// SetFooterColor adding column colors (ANSI codes)
func (t *Table) SetFooterColor(colors ...Colors) {
	if len(t.footers) != len(colors) {
		panic("Number of footer colors must be equal to number of footer.")
	}
	for i := 0; i < len(colors); i++ {
		t.footerParams = append(t.footerParams, makeSequence(colors[i]))
	}
}

// Color returns colors
func Color(colors ...int) []int {
	return colors
}
