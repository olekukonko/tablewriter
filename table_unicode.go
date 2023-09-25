package tablewriter

import "errors"

type UnicodeLineStyle int

const (
	Regular UnicodeLineStyle = iota
	Thick
	Double
)

const (
	// Unicode symbol sets for pre-made themes. The HV sets are for the
	// SetUnicodeHV themes, while the OI sets are for the SetUnicodeOuterInner
	// themes. The first 3 apply for both, for cases where both styles are the
	// same. The last two letters are the initial letter of the corresponding
	// UnicodeLineStyle arguments to the theme-setting functions.
	symsRR   = "─│┌┐└┘├┤┬┴┼─│┌┐└┘├┤┬┴"
	symsTT   = "━┃┏┓┗┛┣┫┳┻╋━┃┏┓┗┛┣┫┳┻"
	symsDD   = "═║╔╗╚╝╠╣╦╩╬═║╔╗╚╝╠╣╦╩"
	symsHVRT = "─┃┎┒┖┚┠┨┰┸╂─┃┎┒┖┚┠┨┰┸"
	symsHVTR = "━│┍┑┕┙┝┥┯┷┿━│┍┑┕┙┝┥┯┷"
	symsHVRD = "─║╓╖╙╜╟╢╥╨╫─║╓╖╙╜╟╢╥╨"
	symsHVDR = "═│╒╕╘╛╞╡╤╧╪═│╒╕╘╛╞╡╤╧"
	symsOIRT = "━┃┏┓┗┛┣┫┳┻╋─│┌┐└┘┝┥┰┸"
	symsOITR = "─│┌┐└┘├┤┬┴┼━┃┏┓┗┛┠┨┯┷"
	symsOIRD = "═║╔╗╚╝╠╣╦╩╬─│┌┐└┘╞╡╥╨"
	symsOIDR = "─│┌┐└┘├┤┬┴┼═║╔╗╚╝╟╢╤╧"
)

func simpleSyms(center, row, column string) []string {
	return []string{row, column, center, center, center, center, center, center, center, center, center, row, column, center, center, center, center, center, center, center, center}
}

// Use unicode box drawing symbols to achieve the specified line styles for
// horizontal and for vertical lines. Note that combinations of thick and double
// lines are not supported. Will return an error in case of unsupported
// combinations.
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
		syms = symsHVRT
	case horizontal == Thick && vertical == Regular:
		syms = symsHVTR
	case horizontal == Regular && vertical == Double:
		syms = symsHVRD
	case horizontal == Double && vertical == Regular:
		syms = symsHVDR
	default:
		return errors.New("Unsupported combination of unicode line styles")
	}
	t.syms = make([]string, 0, 21)
	for _, sym := range []rune(syms) {
		t.syms = append(t.syms, string(sym))
	}
	return nil
}

// Use unicode box drawing symbols to achieve the specified line styles for
// outer boundary and for inner lines. Note that combinations of thick and
// double lines are not supported. Will return an error in case of unsupported
// combinations.
func (t *Table) SetUnicodeOuterInner(outer, inner UnicodeLineStyle) error {
	var syms string
	switch {
	case outer == Regular && inner == Regular:
		syms = symsRR
	case outer == Thick && inner == Thick:
		syms = symsTT
	case outer == Double && inner == Double:
		syms = symsDD
	case outer == Regular && inner == Thick:
		syms = symsOIRT
	case outer == Thick && inner == Regular:
		syms = symsOITR
	case outer == Regular && inner == Double:
		syms = symsOIRD
	case outer == Double && inner == Regular:
		syms = symsOIDR
	default:
		return errors.New("Unsupported combination of unicode line styles")
	}
	t.syms = make([]string, 0, 21)
	for _, sym := range []rune(syms) {
		t.syms = append(t.syms, string(sym))
	}
	return nil
}
