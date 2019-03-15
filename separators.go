package tablewriter

const (
	CENTER  = "+"
	ROW     = "-"
	COLUMN  = "|"
	SPACE   = " "
	NEWLINE = "\n"
)

type SeparatorRow struct {
	Line  string
	Left  string
	Mid   string
	Right string
}

type SeparatorDataRow struct {
	Left  string
	Mid   string
	Right string
}

type Separators struct {
	Top    SeparatorRow
	Bottom SeparatorRow
	Middle SeparatorRow
	Data   SeparatorDataRow
}

var defaultTheme = Separators{
	Top: SeparatorRow{
		Line:  ROW,
		Left:  CENTER,
		Right: CENTER,
		Mid:   CENTER,
	},
	Bottom: SeparatorRow{
		Line:  ROW,
		Left:  CENTER,
		Right: CENTER,
		Mid:   CENTER,
	},
	Middle: SeparatorRow{
		Line:  ROW,
		Left:  CENTER,
		Right: CENTER,
		Mid:   CENTER,
	},
	Data: SeparatorDataRow{
		Left:  COLUMN,
		Right: COLUMN,
		Mid:   COLUMN,
	},
}

var cliTableTheme = Separators{
	Top: SeparatorRow{
		Line:  "─",
		Left:  "┌",
		Right: "┐",
		Mid:   "┬",
	},
	Bottom: SeparatorRow{
		Line:  "─",
		Left:  "└",
		Right: "┘",
		Mid:   "┴",
	},
	Middle: SeparatorRow{
		Line:  "─",
		Left:  "├",
		Right: "┤",
		Mid:   "┼",
	},
	Data: SeparatorDataRow{
		Left:  "│",
		Right: "│",
		Mid:   "│",
	},
}

var cliTableDoubleTheme = Separators{
	Top: SeparatorRow{
		Line:  "═",
		Left:  "╔",
		Right: "╗",
		Mid:   "╤",
	},
	Bottom: SeparatorRow{
		Line:  "═",
		Left:  "╚",
		Right: "╝",
		Mid:   "╧",
	},
	Middle: SeparatorRow{
		Line:  "─",
		Left:  "╟",
		Right: "╢",
		Mid:   "┼",
	},
	Data: SeparatorDataRow{
		Left:  "║",
		Right: "║",
		Mid:   "│",
	},
}

// Themes defines some out-of-the box themes
var Themes = struct {
	CliTable       Separators
	CliTableDouble Separators
	Default        Separators
}{
	CliTable:       cliTableTheme,
	CliTableDouble: cliTableDoubleTheme,
	Default:        defaultTheme,
}

type linePosition int

const (
	topLine    linePosition = 0
	middleLine linePosition = 1
	bottomLine linePosition = 2
	dataLine   linePosition = 3
)

func (s *Separators) getSeparatorsRow(lp linePosition) SeparatorRow {
	switch lp {
	case topLine:
		return s.Top
	case bottomLine:
		return s.Bottom
	case middleLine:
		return s.Middle
	}
	return s.Middle
}
