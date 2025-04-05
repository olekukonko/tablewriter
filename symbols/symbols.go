package symbols

// Default symbols
const (
	SPACE   = " "
	NEWLINE = "\n"
)

// Symbols defines the interface for table symbols
type Symbols interface {
	Center() string
	Row() string
	Column() string
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

func (s Default) Center() string { return "+" }
func (s Default) Row() string    { return "-" }
func (s Default) Column() string { return "|" }

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

//@todo add Terminal
//var cliTableDoubleTheme = Separators{
//	Top: SeparatorRow{
//		Line:  "═",
//		Left:  "╔",
//		Right: "╗",
//		Mid:   "╤",
//	},
//	Bottom: SeparatorRow{
//		Line:  "═",
//		Left:  "╚",
//		Right: "╝",
//		Mid:   "╧",
//	},
//	Middle: SeparatorRow{
//		Line:  "─",
//		Left:  "╟",
//		Right: "╢",
//		Mid:   "┼",
//	},
//	Data: SeparatorDataRow{
//		Left:  "║",
//		Right: "║",
//		Mid:   "│",
//	},
//}
