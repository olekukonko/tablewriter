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

type Colors []int

// CellColor Color information (ie FgHiBlackColor, FgHiRedColor, etc...)
// For the column and row position in your table
type CellColor struct {
	column int
	row    int
	color  string
}

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
func (t *Table) SetHeaderColor(colors ...Colors) {
	if t.colSize != len(colors) {
		panic("Number of header colors must be equal to number of headers.")
	}
	for i := 0; i < len(colors); i++ {
		t.headerParams = append(t.headerParams, makeSequence(colors[i]))
	}
}

// Adding column colors (ANSI codes)
func (t *Table) SetColumnColor(colors ...Colors) {
	if t.colSize != len(colors) {
		panic("Number of column colors must be equal to number of headers.")
	}
	for i := 0; i < len(colors); i++ {
		t.columnsParams = append(t.columnsParams, makeSequence(colors[i]))
	}
}

// Adding column colors (ANSI codes)
func (t *Table) SetFooterColor(colors ...Colors) {
	if len(t.footers) != len(colors) {
		panic("Number of footer colors must be equal to number of footer.")
	}
	for i := 0; i < len(colors); i++ {
		t.footerParams = append(t.footerParams, makeSequence(colors[i]))
	}
}

// Adding color for cell (ANSI codes) {
// Column and Rows start at 0
// This does NOT do anything for header/footers
func (t *Table) SetCellColor(column, row int, color Colors) {
	if column < 0 || column > t.colSize-1 {
		panic(fmt.Sprintf("Column Value: %d is outside column size %d", column, t.colSize-1))
	} else if row < 0 || row > len(t.lines)-1 {
		panic(fmt.Sprintf("Row Value: %d is outside row size %d", row, len(t.lines)-1))
	}

	t.cellParams = append(t.cellParams, CellColor{
		column: column,
		row:    row,
		color:  makeSequence(color),
	})
}

// Returns makeSequence version of color by cell indices
func (t *Table) getColorForCell(column, row int) string {
	for _, colorInfo := range t.cellParams {
		if column == colorInfo.column && row == colorInfo.row {
			return colorInfo.color
		}
	}

	return ""
}

func Color(colors ...int) []int {
	return colors
}
