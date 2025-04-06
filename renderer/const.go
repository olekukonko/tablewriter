// formatter/const.go
package renderer

import (
	"github.com/olekukonko/tablewriter/symbols"
	"io"
)

// Alignment constants
const (
	AlignCenter = "center"
	AlignRight  = "right"
	AlignLeft   = "left"

	// AlignDefault make left default
	AlignDefault = AlignLeft
)

type Position string

const (
	Header Position = "header"
	Row             = "row"
	Footer          = "footer"
)

const (
	Top = iota
	Middle
	Bottom
)

// Context carries rendering information to the renderer
type Context struct {
	Level      int
	Align      string
	First      bool
	Last       bool
	Widths     map[int]int
	Padding    symbols.Padding
	ColPadding map[int]symbols.Padding
	ColAligns  map[int]string // Per-column alignments
}

// Structure defines the interface for table renderers
type Structure interface {
	FormatHeader(w io.Writer, headers []string, ctx Context)
	FormatRow(w io.Writer, row []string, ctx Context)
	FormatFooter(w io.Writer, footers []string, ctx Context)
	FormatLine(w io.Writer, ctx Context)
	GetColumnWidths() []int
	Reset()
}
