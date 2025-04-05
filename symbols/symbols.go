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

// DefaultSymbols provides the default ASCII symbols
type DefaultSymbols struct{}

func (s DefaultSymbols) Center() string { return "+" }
func (s DefaultSymbols) Row() string    { return "-" }
func (s DefaultSymbols) Column() string { return "|" }

// UnicodeSymbols provides Unicode box-drawing symbols
type UnicodeSymbols struct {
	style UnicodeStyle
}

// NewUnicodeSymbols creates a new UnicodeSymbols instance with the specified style
func NewUnicodeSymbols(style UnicodeStyle) *UnicodeSymbols {
	return &UnicodeSymbols{style: style}
}

func (s *UnicodeSymbols) Center() string {
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

func (s *UnicodeSymbols) Row() string {
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

func (s *UnicodeSymbols) Column() string {
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
