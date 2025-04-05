// formatter/const.go
package formatter

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

// Formatter defines the interface for formatting table elements
type Formatter interface {
	FormatHeader(w io.Writer, headers []string, colWidths map[int]int)
	FormatRow(w io.Writer, row []string, colWidths map[int]int, isFirstRow bool)
	FormatFooter(w io.Writer, footers []string, colWidths map[int]int)
	FormatLine(w io.Writer, colWidths map[int]int, isTop bool)
	Configure(opt Option) // Apply table options to the formatter
	Reset()               // Clears internal state
}

// Border defines table border settings
type Border struct {
	Left   bool
	Right  bool
	Top    bool
	Bottom bool
}

// simpleSyms generates a basic symbol set (placeholder from original)
func simpleSyms(center, row, column string) []string {
	return []string{row, column, center, center, center, center, center, center, center, center, center}
}
