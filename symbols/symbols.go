package symbols

// Default symbols
const (
	SPACE   = " "
	NEWLINE = "\n"
)

// Symbols defines the interface for table symbols
type Symbols interface {
	Center() string      // e.g., "╬" for junctions
	Row() string         // e.g., "═" for horizontal lines
	Column() string      // e.g., "║" for vertical lines
	TopLeft() string     // e.g., "╔"
	TopMid() string      // e.g., "╦"
	TopRight() string    // e.g., "╗"
	MidLeft() string     // e.g., "╠"
	MidRight() string    // e.g., "╣"
	BottomLeft() string  // e.g., "╚"
	BottomMid() string   // e.g., "╩"
	BottomRight() string // e.g., "╝"
}

// UnicodeStyle represents different Unicode box drawing styles
type UnicodeStyle int

// Unicode line style constants
const (
	RegularRegular UnicodeStyle = iota
	ThickThick
	DoubleDouble
	RegularThick
	ThickRegular
	RegularDouble
	DoubleRegular
)

// String representation of Unicode styles
func (s UnicodeStyle) String() string {
	return [...]string{
		"RegularRegular",
		"ThickThick",
		"DoubleDouble",
		"RegularThick",
		"ThickRegular",
		"RegularDouble",
		"DoubleRegular",
	}[s]
}

// Default provides the default ASCII symbols
type Default struct{}

func (s Default) Center() string      { return "+" }
func (s Default) Row() string         { return "-" }
func (s Default) Column() string      { return "|" }
func (s Default) TopLeft() string     { return "+" }
func (s Default) TopMid() string      { return "+" }
func (s Default) TopRight() string    { return "+" }
func (s Default) MidLeft() string     { return "+" }
func (s Default) MidRight() string    { return "+" }
func (s Default) BottomLeft() string  { return "+" }
func (s Default) BottomMid() string   { return "+" }
func (s Default) BottomRight() string { return "+" }

// Unicode provides Unicode box-drawing symbols
type Unicode struct {
	style UnicodeStyle
}

// NewUnicodeSymbols creates a new Unicode instance with the specified style
func NewUnicodeSymbols(style UnicodeStyle) *Unicode {
	return &Unicode{style: style}
}

func (s *Unicode) Center() string {
	switch s.style {
	case RegularRegular:
		return "┼"
	case ThickThick:
		return "╋"
	case DoubleDouble:
		return "╬"
	case RegularThick:
		return "╂"
	case ThickRegular:
		return "┿"
	case RegularDouble:
		return "╫"
	case DoubleRegular:
		return "╪"
	default:
		return "+"
	}
}

func (s *Unicode) Row() string {
	switch s.style {
	case RegularRegular, RegularThick, RegularDouble:
		return "─"
	case ThickThick, ThickRegular:
		return "━"
	case DoubleDouble, DoubleRegular:
		return "═"
	default:
		return "-"
	}
}

func (s *Unicode) Column() string {
	switch s.style {
	case RegularRegular, ThickRegular, DoubleRegular:
		return "│"
	case ThickThick, RegularThick:
		return "┃"
	case DoubleDouble, RegularDouble:
		return "║"
	default:
		return "|"
	}
}

func (s *Unicode) TopLeft() string {
	switch s.style {
	case DoubleDouble:
		return "╔"
	default:
		return "+"
	}
}

func (s *Unicode) TopMid() string {
	switch s.style {
	case DoubleDouble:
		return "╦"
	default:
		return "+"
	}
}

func (s *Unicode) TopRight() string {
	switch s.style {
	case DoubleDouble:
		return "╗"
	default:
		return "+"
	}
}

func (s *Unicode) MidLeft() string {
	switch s.style {
	case DoubleDouble:
		return "╠"
	default:
		return "+"
	}
}

func (s *Unicode) MidRight() string {
	switch s.style {
	case DoubleDouble:
		return "╣"
	default:
		return "+"
	}
}

func (s *Unicode) BottomLeft() string {
	switch s.style {
	case DoubleDouble:
		return "╚"
	default:
		return "+"
	}
}

func (s *Unicode) BottomMid() string {
	switch s.style {
	case DoubleDouble:
		return "╩"
	default:
		return "+"
	}
}

func (s *Unicode) BottomRight() string {
	switch s.style {
	case DoubleDouble:
		return "╝"
	default:
		return "+"
	}
}
