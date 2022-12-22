package tablewriter

import "errors"

type UnicodeLineStyle int

const (
	Regular UnicodeLineStyle = iota
	Thick
	Double
)

const (
	symsRR = "─│┌┐└┘├┤┬┴┼"
	symsTT = "━┃┏┓┗┛┣┫┳┻╋"
	symsDD = "═║╔╗╚╝╠╣╦╩╬"
	symsRT = "─┃┎┒┖┚┠┨┰┸╂"
	symsTR = "━│┍┑┕┙┝┥┯┷┿"
	symsRD = "─║╓╖╙╜╟╢╥╨╫"
	symsDR = "═│╒╕╘╛╞╡╤╧╪"
)

func simpleSyms(center, row, column string) []string {
	return []string{row, column, center, center, center, center, center, center, center, center, center}
}

// Use unicode box drawing symbols to achieve the specified line styles.
// Note that combinations of thick and double lines are not supported.
// Will return an error in case of unsupported combinations.
func (t *Table) SetUnicodeHV(horizontal, vertical UnicodeLineStyle) error {
	var syms string
	switch {
	case horizontal == Regular && vertical == Regular:
		syms = symsRR
	case horizontal == Thick && vertical == Thick:
		syms = symsTT
	case horizontal == Double && vertical == Double:
		syms = symsDD
	case horizontal == Regular && vertical == Thick:
		syms = symsRT
	case horizontal == Thick && vertical == Regular:
		syms = symsTR
	case horizontal == Regular && vertical == Double:
		syms = symsRD
	case horizontal == Double && vertical == Regular:
		syms = symsDR
	default:
		return errors.New("Unsupported combination of unicode line styles")
	}
	t.syms = make([]string, 0, 11)
	for _, sym := range []rune(syms) {
		t.syms = append(t.syms, string(sym))
	}
	return nil
}
