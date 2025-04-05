package symbols

// Basic constants
const (
	SPACE   = " "
	NEWLINE = "\n"
)

const (
	Top    = "top"
	Middle = "middle"
	Bottom = "bottom"
)

// Symbols defines the interface for table border symbols
type Symbols interface {
	// Basic components
	Center() string // Junction symbol (where lines cross)
	Row() string    // Horizontal line symbol
	Column() string // Vertical line symbol

	// Corners and junctions
	TopLeft() string     // Top-left corner
	TopMid() string      // Top junction
	TopRight() string    // Top-right corner
	MidLeft() string     // Left junction
	MidRight() string    // Right junction
	BottomLeft() string  // Bottom-left corner
	BottomMid() string   // Bottom junction
	BottomRight() string // Bottom-right corner

	// Optional: Header separator specific symbols
	HeaderLeft() string
	HeaderMid() string
	HeaderRight() string
}

// BorderStyle defines different border styling options
type BorderStyle int

const (
	StyleNone BorderStyle = iota
	StyleASCII
	StyleLight
	StyleHeavy
	StyleDouble
	StyleLightHeavy
	StyleHeavyLight
	StyleLightDouble
	StyleDoubleLight
	StyleRounded
	StyleMarkdown
)

// String representation of border styles
func (s BorderStyle) String() string {
	return [...]string{
		"None",
		"ASCII",
		"Light",
		"Heavy",
		"Double",
		"LightHeavy",
		"HeavyLight",
		"LightDouble",
		"DoubleLight",
		"Rounded",
		"Markdown",
	}[s]
}

// NewSymbols creates a new Symbols instance with the specified style
func NewSymbols(style BorderStyle) Symbols {
	switch style {
	case StyleASCII:
		return &ASCII{}
	case StyleLight:
		return &Unicode{
			row:    "─",
			column: "│",
			center: "┼",
			corners: [9]string{
				"┌", "┬", "┐",
				"├", "┼", "┤",
				"└", "┴", "┘",
			},
		}
	case StyleHeavy:
		return &Unicode{
			row:    "━",
			column: "┃",
			center: "╋",
			corners: [9]string{
				"┏", "┳", "┓",
				"┣", "╋", "┫",
				"┗", "┻", "┛",
			},
		}
	case StyleDouble:
		return &Unicode{
			row:    "═",
			column: "║",
			center: "╬",
			corners: [9]string{
				"╔", "╦", "╗",
				"╠", "╬", "╣",
				"╚", "╩", "╝",
			},
		}
	case StyleLightHeavy:
		return &Unicode{
			row:    "─",
			column: "┃",
			center: "╂",
			corners: [9]string{
				"┍", "┯", "┑",
				"┝", "╂", "┥",
				"┕", "┷", "┙",
			},
		}
	case StyleHeavyLight:
		return &Unicode{
			row:    "━",
			column: "│",
			center: "┿",
			corners: [9]string{
				"┎", "┰", "┒",
				"┠", "┿", "┨",
				"┖", "┸", "┚",
			},
		}
	case StyleLightDouble:
		return &Unicode{
			row:    "─",
			column: "║",
			center: "╫",
			corners: [9]string{
				"╓", "╥", "╖",
				"╟", "╫", "╢",
				"╙", "╨", "╜",
			},
		}
	case StyleDoubleLight:
		return &Unicode{
			row:    "═",
			column: "│",
			center: "╪",
			corners: [9]string{
				"╒", "╤", "╕",
				"╞", "╪", "╡",
				"╘", "╧", "╛",
			},
		}
	case StyleRounded:
		return &Unicode{
			row:    "─",
			column: "│",
			center: "┼",
			corners: [9]string{
				"╭", "┬", "╮",
				"├", "┼", "┤",
				"╰", "┴", "╯",
			},
		}
	case StyleMarkdown:
		return &Markdown{}
	default:
		return &Empty{}
	}
}

// ASCII provides basic ASCII border symbols
type ASCII struct{}

func (s *ASCII) Center() string      { return "+" }
func (s *ASCII) Row() string         { return "-" }
func (s *ASCII) Column() string      { return "|" }
func (s *ASCII) TopLeft() string     { return "+" }
func (s *ASCII) TopMid() string      { return "+" }
func (s *ASCII) TopRight() string    { return "+" }
func (s *ASCII) MidLeft() string     { return "+" }
func (s *ASCII) MidRight() string    { return "+" }
func (s *ASCII) BottomLeft() string  { return "+" }
func (s *ASCII) BottomMid() string   { return "+" }
func (s *ASCII) BottomRight() string { return "+" }
func (s *ASCII) HeaderLeft() string  { return "+" }
func (s *ASCII) HeaderMid() string   { return "+" }
func (s *ASCII) HeaderRight() string { return "+" }

// Unicode provides configurable Unicode border symbols
type Unicode struct {
	row     string
	column  string
	center  string
	corners [9]string // [topLeft, topMid, topRight, midLeft, center, midRight, bottomLeft, bottomMid, bottomRight]
}

func (s *Unicode) Center() string      { return s.center }
func (s *Unicode) Row() string         { return s.row }
func (s *Unicode) Column() string      { return s.column }
func (s *Unicode) TopLeft() string     { return s.corners[0] }
func (s *Unicode) TopMid() string      { return s.corners[1] }
func (s *Unicode) TopRight() string    { return s.corners[2] }
func (s *Unicode) MidLeft() string     { return s.corners[3] }
func (s *Unicode) MidRight() string    { return s.corners[5] }
func (s *Unicode) BottomLeft() string  { return s.corners[6] }
func (s *Unicode) BottomMid() string   { return s.corners[7] }
func (s *Unicode) BottomRight() string { return s.corners[8] }
func (s *Unicode) HeaderLeft() string  { return s.MidLeft() }
func (s *Unicode) HeaderMid() string   { return s.Center() }
func (s *Unicode) HeaderRight() string { return s.MidRight() }

// Markdown provides symbols for Markdown-style tables
type Markdown struct{}

func (s *Markdown) Center() string      { return "|" }
func (s *Markdown) Row() string         { return "-" }
func (s *Markdown) Column() string      { return "|" }
func (s *Markdown) TopLeft() string     { return "" }
func (s *Markdown) TopMid() string      { return "" }
func (s *Markdown) TopRight() string    { return "" }
func (s *Markdown) MidLeft() string     { return "|" }
func (s *Markdown) MidRight() string    { return "|" }
func (s *Markdown) BottomLeft() string  { return "" }
func (s *Markdown) BottomMid() string   { return "" }
func (s *Markdown) BottomRight() string { return "" }
func (s *Markdown) HeaderLeft() string  { return "|" }
func (s *Markdown) HeaderMid() string   { return "|" }
func (s *Markdown) HeaderRight() string { return "|" }

// Empty provides no border symbols
type Empty struct{}

func (s *Empty) Center() string      { return "" }
func (s *Empty) Row() string         { return "" }
func (s *Empty) Column() string      { return "" }
func (s *Empty) TopLeft() string     { return "" }
func (s *Empty) TopMid() string      { return "" }
func (s *Empty) TopRight() string    { return "" }
func (s *Empty) MidLeft() string     { return "" }
func (s *Empty) MidRight() string    { return "" }
func (s *Empty) BottomLeft() string  { return "" }
func (s *Empty) BottomMid() string   { return "" }
func (s *Empty) BottomRight() string { return "" }
func (s *Empty) HeaderLeft() string  { return "" }
func (s *Empty) HeaderMid() string   { return "" }
func (s *Empty) HeaderRight() string { return "" }
