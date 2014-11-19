// Copyright 2014 Oleku Konko All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// This module is a Table Writer  API for the Go Programming Language.
// The protocols were written in pure Go and works on windows and unix systems

// Create & Generate text based table
package tablewriter

import (
	"fmt"
	"io"
	"regexp"
	"strings"
)

const (
	MAX_ROW_WIDTH = 30
)

const (
	CENTER = "+"
	ROW    = "-"
	COLUMN = "|"
	SPACE  = " "
)

const (
	ALIGN_DEFAULT = iota
	ALIGN_CENTER
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
	footers []string
	mW      int
	pCenter string
	pRow    string
	pColumn string
	tColumn int
	tRow    int
	align   int
	rowLine bool
	border  bool
	colSize int
}

// Start New Table
// Take io.Writer Directly
func NewWriter(writer io.Writer) *table {
	t := &table{
		out:     writer,
		rows:    [][]string{},
		lines:   [][][]string{},
		cs:      make(map[int]int),
		rs:      make(map[int]int),
		headers: []string{},
		footers: []string{},
		mW:      MAX_ROW_WIDTH,
		pCenter: CENTER,
		pRow:    ROW,
		pColumn: COLUMN,
		tColumn: -1,
		tRow:    -1,
		align:   ALIGN_DEFAULT,
		rowLine: false,
		border:  true,
		colSize: -1}
	return t
}

// Render table output
func (t table) Render() {
	if t.border {
		t.printLine(true)
	}
	t.printHeading()
	t.printRows()

	if !t.rowLine && t.border {
		t.printLine(true)
	}
	t.printFooter()

}

// Set table header
func (t *table) SetHeader(keys []string) {
	t.colSize = len(keys)
	for i, v := range keys {
		t.parseDimension(v, i, -1)
		t.headers = append(t.headers, Title(v))
	}
}

// Set table Footer
func (t *table) SetFooter(keys []string) {
	//t.colSize = len(keys)
	for i, v := range keys {
		t.parseDimension(v, i, -1)
		t.footers = append(t.footers, Title(v))
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

// Set Row Line
// This would enable / disable a line on each row of the table
func (t *table) SetRowLine(line bool) {
	t.rowLine = line
}

// Set Table Border
// This would enable / disable line around the table
func (t *table) SetBorder(border bool) {
	t.border = border
}

// Append row to table
func (t *table) Append(row []string) error {
	rowSize := len(t.headers)
	if rowSize > t.colSize {
		t.colSize = rowSize
	}

	n := len(t.lines)
	line := [][]string{}
	for i, v := range row {

		// Detect string  width
		// Detect String height
		// Break strings into words
		out := t.parseDimension(v, i, n)

		// Append broken words
		line = append(line, out)
	}
	t.lines = append(t.lines, line)
	return nil
}

// Allow Support for Bulk Append
// Eliminates repeated for loops
func (t *table) AppendBulk(rows [][]string) (err error) {
	for _, row := range rows {
		err = t.Append(row)
		if err != nil {
			return err
		}
	}
	return nil
}

// Print line based on row width
func (t table) printLine(nl bool) {
	fmt.Fprint(t.out, t.pCenter)
	for i := 0; i < len(t.cs); i++ {
		v := t.cs[i]
		fmt.Fprintf(t.out, "%s%s%s%s",
			t.pRow,
			strings.Repeat(string(t.pRow), v),
			t.pRow,
			t.pCenter)
	}
	if nl {
		fmt.Fprintln(t.out)
	}
}

// Print heading information
func (t table) printHeading() {
	// Check if headers is available
	if len(t.headers) < 1 {
		return
	}

	// Check if border is set
	// Replace with space if not set
	fmt.Fprint(t.out, ConditionString(t.border, t.pColumn, SPACE))

	// Identify last column
	end := len(t.cs) - 1

	// Print Heading column
	for i := 0; i <= end; i++ {
		v := t.cs[i]
		pad := ConditionString((i == end && !t.border), SPACE, t.pColumn)
		fmt.Fprintf(t.out, " %s %s",
			Pad(t.headers[i], SPACE, v),
			pad)
	}
	// Next line
	fmt.Fprintln(t.out)
	t.printLine(true)
}

// Print heading information
func (t table) printFooter() {
	// Check if headers is available
	if len(t.footers) < 1 {
		return
	}

	// Only print line if border is not set
	if !t.border {
		t.printLine(true)
	}
	// Check if border is set
	// Replace with space if not set
	fmt.Fprint(t.out, ConditionString(t.border, t.pColumn, SPACE))

	// Identify last column
	end := len(t.cs) - 1

	// Print Heading column
	for i := 0; i <= end; i++ {
		v := t.cs[i]
		pad := ConditionString((i == end && !t.border), SPACE, t.pColumn)

		if len(t.footers[i]) == 0 {
			pad = SPACE
		}
		fmt.Fprintf(t.out, " %s %s",
			Pad(t.footers[i], SPACE, v),
			pad)
	}
	// Next line
	fmt.Fprintln(t.out)
	//t.printLine(true)

	hasPrinted := false

	for i := 0; i <= end; i++ {
		v := t.cs[i]
		pad := t.pRow
		center := t.pCenter
		length := len(t.footers[i])

		if length > 0 {
			hasPrinted = true
		}

		// Set center to be space if length is 0
		if length == 0 && !t.border {
			center = SPACE
		}

		// Print first junction
		if i == 0 {
			fmt.Fprint(t.out, center)
		}

		// Pad With space of length is 0
		if length == 0 {
			pad = SPACE
		}
		// Ignore left space of it has printed before
		if hasPrinted || t.border {
			pad = t.pRow
			center = t.pCenter
		}

		// Change Center start position
		if center == SPACE {
			if i < end && len(t.footers[i+1]) != 0 {
				center = t.pCenter
			}
		}

		// Print the footer
		fmt.Fprintf(t.out, "%s%s%s%s",
			pad,
			strings.Repeat(string(pad), v),
			pad,
			center)

	}

	fmt.Fprintln(t.out)

}

func (t table) printRows() {
	for i, lines := range t.lines {
		t.printRow(lines, i)
	}

}

// Print Row Information
// Adjust column alignment based on type

func (t table) printRow(columns [][]string, colKey int) {
	// Get Maximum Height
	max := t.rs[colKey]
	total := len(columns)

	// TODO Fix uneven col size
	// if total < t.colSize {
	//	for n := t.colSize - total; n < t.colSize ; n++ {
	//		columns = append(columns, []string{SPACE})
	//		t.cs[n] = t.mW
	//	}
	//}

	// Pad Each Height
	// pads := []int{}
	pads := []int{}

	for i, line := range columns {
		length := len(line)
		pad := max - length
		pads = append(pads, pad)
		for n := 0; n < pad; n++ {
			columns[i] = append(columns[i], "  ")
		}
	}
	//fmt.Println(max, "\n")
	for x := 0; x < max; x++ {
		for y := 0; y < total; y++ {

			// Check if border is set
			fmt.Fprint(t.out, ConditionString((!t.border && y == 0), SPACE, t.pColumn))

			fmt.Fprintf(t.out, SPACE)
			str := columns[y][x]

			// This would print alignment
			// Default alignment  would use multiple configuration
			switch t.align {
			case ALIGN_CENTER: //
				fmt.Fprintf(t.out, "%s", Pad(str, SPACE, t.cs[y]))
			case ALIGN_RIGHT:
				fmt.Fprintf(t.out, "%s", PadLeft(str, SPACE, t.cs[y]))
			case ALIGN_LEFT:
				fmt.Fprintf(t.out, "%s", PadRight(str, SPACE, t.cs[y]))
			default:
				if decimal.MatchString(strings.TrimSpace(str)) || percent.MatchString(strings.TrimSpace(str)) {
					fmt.Fprintf(t.out, "%s", PadLeft(str, SPACE, t.cs[y]))
				} else {
					fmt.Fprintf(t.out, "%s", PadRight(str, SPACE, t.cs[y]))

					// TODO Custom alignment per column
					//if max == 1 || pads[y] > 0 {
					//	fmt.Fprintf(t.out, "%s", Pad(str, SPACE, t.cs[y]))
					//} else {
					//	fmt.Fprintf(t.out, "%s", PadRight(str, SPACE, t.cs[y]))
					//}

				}
			}
			fmt.Fprintf(t.out, SPACE)
		}
		// Check if border is set
		// Replace with space if not set
		fmt.Fprint(t.out, ConditionString(t.border, t.pColumn, SPACE))
		fmt.Fprintln(t.out)
	}

	if t.rowLine {
		t.printLine(true)
	}

}

func (t *table) parseDimension(str string, colKey, rowKey int) []string {
	var (
		raw []string
		max int
	)
	w := DisplayWidth(str)
	// Calculate Width
	// Check if with is grater than maximum width
	if w > t.mW {
		w = t.mW
	}

	// Check if width exists
	v, ok := t.cs[colKey]
	if !ok || v < w || v == 0 {
		t.cs[colKey] = w
	}

	if rowKey == -1 {
		return raw
	}
	// Calculate Height
	raw, max = WrapString(str, t.cs[colKey])

	// Make sure the with is the same length as maximum word
	// Important for cases where the width is smaller than maxu word
	if max > t.cs[colKey] {
		t.cs[colKey] = max
	}

	h := len(raw)
	v, ok = t.rs[rowKey]

	if !ok || v < h || v == 0 {
		t.rs[rowKey] = h
	}
	//fmt.Printf("Raw %+v %d\n", raw, len(raw))
	return raw
}
