package table

import (
	"errors"
	"fmt"
	"io"
	"strings"
	"unicode/utf8"
	"regexp"
)

const (
	MAX_ROW_WIDTH = 30
)

const (
	CENTRE = "+"
	ROW    = "-"
	COLUMN = "|"
)


const (
	ALIGN_DEFAULT = iota
	ALIGN_CENTRE
	ALIGN_RIGHT
	ALIGN_LEFT
)

var (
	decimal = regexp.MustCompile(`^[0-9]+.[0-9]+$`)
	percent = regexp.MustCompile(`^[0-9]+.[0-9]+%$`)
)

type table struct {
	out     io.Writer
	rows    [][]string
	lines   [][][]string
	cs      map[int]int
	rs      map[int]int
	headers []string
	mW      int
	pCenter string
	pRow    string
	pColumn string

	tColumn int
	tRow    int
	align   int
}

func NewTable(writer io.Writer) *table {
	t := &table{
		out:     writer,
		rows:    [][]string{},
		lines:   [][][]string{},
		cs:      make(map[int]int),
		rs:      make(map[int]int),
		headers: []string{},
		mW:      MAX_ROW_WIDTH,
		pCenter: CENTRE,
		pRow:    ROW,
		pColumn: COLUMN,
		tColumn: -1,
		tRow:    -1,
		align : ALIGN_DEFAULT}

	return t
}

// Render table output
func (t table) Render() {
	t.printLine(true)
	t.printHeading()
	t.printRows()
	t.printLine(true)
	fmt.Fprint(t.out, "\nDone\n")
}

// Set table header
func (t *table) SetHeader(keys []string) {
	for i, v := range keys {
		t.setWidth(v, i)
		t.headers = append(t.headers, strings.ToUpper(v))
	}
}

// Set the Default column width
func (t *table) SetColWidth(width int) {
	t.mW = width
}

// Set the Column Separator
func (t *table) SetColumnSeparator(sep string) {
	t.pColumn = sep
}

// Set the Row Separator
func (t *table) SetRowSeparator(sep string) {
	t.pRow = sep
}

// Set the center Separator
func (t *table) SetCenterSeparator(sep string) {
	t.pCenter = sep
}

// Set Table Alignment
func (t *table) SetAlignment(align int) {
	t.align = align
}

// Append row to table
func (t *table) Append(row []string) error {

	h := len(t.headers)
	if h > 0 && h != len(row) {
		return errors.New("Heder length does not match")
	}

	line := [][]string{}
	for i, v := range row {
		// Detect string  width
		t.setWidth(v, i)

		// Detect String height
		out := t.setHeight(v, i)

		// Append broken words
		line = append(line, out)
	}

	t.lines = append(t.lines, line)
	return nil
}

func (t table) printLine(nl bool) {
	fmt.Fprint(t.out, t.pCenter)
	for _, v := range t.cs {
		fmt.Printf("%s%s%s%s",
			t.pRow,
			strings.Repeat(string(t.pRow), v),
			t.pRow,
			t.pCenter)
	}
	if nl {
		fmt.Println()
	}
}

func (t table) printHeading() {
	if len(t.headers) < 1 {
		return
	}

	fmt.Fprint(t.out, t.pColumn)
	for i, v := range t.cs {
		fmt.Fprintf(t.out, " %s %s",
			Pad(t.headers[i], " ", v),
			t.pColumn)
	}

	fmt.Println()
	t.printLine(true)
}

func (t table) printRows() {
	for i, lines := range t.lines {
		t.printRow(lines, i)
	}

}
func (t table) printRow(lines [][]string, colKey int) {
	// Get Maximum Height
	max := t.rs[colKey]
	total := len(lines)
	// Pad Each Height
	align := []int{}
	for i, line := range lines {
		length := len(line)
		pad := max - length
		align = append(align, pad) // save align Information
		for n := 0; n < pad; n++ {
			lines[i] = append(lines[i], "  ")
		}
	}

	for x := 0; x < max; x++ {
		for y := 0; y < total; y++ {
			fmt.Fprint(t.out, t.pColumn)
			fmt.Fprintf(t.out, " ")

			str := lines[y][x]

			switch t.align{
			case ALIGN_CENTRE : //
				fmt.Fprintf(t.out, "%s", Pad(str, " ", t.cs[y]))
			case ALIGN_LEFT :
				fmt.Fprintf(t.out, "%s", PadLeft(str, " ", t.cs[y]))
			case ALIGN_RIGHT  :
				fmt.Fprintf(t.out, "%s", PadRight(str, " ", t.cs[y]))
			default :
				if decimal.MatchString(strings.TrimSpace(str)) || percent.MatchString(strings.TrimSpace(str)) {
					fmt.Fprintf(t.out, "%s", PadLeft(str, " ", t.cs[y]))
				} else {
					if align[y] >= 0 {
						fmt.Fprintf(t.out, "%s", Pad(str, " ", t.cs[y]))
					} else {
						fmt.Fprintf(t.out, "%s", PadRight(str, " ", t.cs[y]))
					}
				}
			}

			fmt.Printf(" ")
		}
		fmt.Fprint(t.out, t.pColumn)
		fmt.Fprintln(t.out)
	}

}

func (t *table) setWidth(str string, colKey int) string {
	w := utf8.RuneCountInString(str)
	if w > t.mW {
		w = t.mW
	}

	v, ok := t.cs[colKey]
	if !ok || v < w || v == 0 {
		t.cs[colKey] = w
	}

	//fmt.Println(w, strings.IndexAny(str, " \n"), t.cs[colKey])
	return str
}

func (t *table) setHeight(str string, colKey int) []string {

	raw, max := WrapString(str, t.cs[colKey])
	// Make sure the with is the same length as maximum word
	if max > t.cs[colKey] {
		t.cs[colKey] = max
	}

	h := len(raw)
	v, ok := t.rs[colKey]
	if !ok || v < h {
		t.rs[colKey] = h
	}
	//fmt.Printf("Raw %+v %d\n", raw, len(raw))
	return raw
}
