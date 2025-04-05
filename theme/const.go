// formatter/const.go
package theme

import "io"

// Alignment constants
const (
	ALIGN_DEFAULT = iota
	ALIGN_CENTER
	ALIGN_RIGHT
	ALIGN_LEFT
)

// DefaultMaxWidth is the default maximum column width
const DefaultMaxWidth = 30

// Structure defines the interface for formatting table elements
type Structure interface {
	FormatHeader(w io.Writer, headers []string, colWidths map[int]int)
	FormatRow(w io.Writer, row []string, colWidths map[int]int, isFirstRow bool)
	FormatFooter(w io.Writer, footers []string, colWidths map[int]int)
	FormatLine(w io.Writer, colWidths map[int]int, isTop bool)
	GetColumnWidths() []int // Returns per-column width overrides
	Reset()                 // Clears internal state

}

// Border defines table border settings
type Border struct {
	Left   bool
	Right  bool
	Top    bool
	Bottom bool
}
