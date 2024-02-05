// Copyright 2014 Oleku Konko All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// This module is a Table Writer  API for the Go Programming Language.
// The protocols were written in pure Go and works on windows and unix systems

package tablewriter

import (
	"math"
	"regexp"
	"strings"

	"github.com/mattn/go-runewidth"
)

var ansi = generateEscapeFilterRegex()

// generateEscapeFilterRegex builds a regex to remove non-printing ANSI escape codes from strings so
// that their display width can be determined accurately. The regex is complicated enough that it's
// better to build it programmatically than to write it by hand.
// Based on https://en.wikipedia.org/wiki/ANSI_escape_code#Fe_Escape_sequences
func generateEscapeFilterRegex() *regexp.Regexp {
	var regESC = "\x1b" // ASCII escape
	var regBEL = "\x07" // ASCII bell

	// String Terminator - ends ANSI sequences
	var regST = "(" + regESC + "\\\\" + "|" + regBEL + ")"

	// Control Sequence Introducer - usually color codes
	// esc + [ + zero or more 0x30-0x3f + zero or more 0x20-0x2f and a single 0x40-0x7e
	var regCSI = regESC + "\\[" + "[\x30-\x3f]*[\x20-\x2f]*[\x40-\x7e]"

	// Operating System Command - hyperlinks
	// esc + ] + any number of any chars + ST
	var regOSC = regESC + "\\]" + ".*?" + regST

	return regexp.MustCompile("(" + regCSI + "|" + regOSC + ")")
}

func DisplayWidth(str string) int {
	return runewidth.StringWidth(ansi.ReplaceAllLiteralString(str, ""))
}

// ConditionString Simple Condition for string
// Returns value based on condition
func ConditionString(cond bool, valid, inValid string) string {
	if cond {
		return valid
	}
	return inValid
}

func isNumOrSpace(r rune) bool {
	return ('0' <= r && r <= '9') || r == ' '
}

// Title Format Table Header
// Replace _ , . and spaces
func Title(name string) string {
	origLen := len(name)
	rs := []rune(name)
	for i, r := range rs {
		switch r {
		case '_':
			rs[i] = ' '
		case '.':
			// ignore floating number 0.0
			if (i != 0 && !isNumOrSpace(rs[i-1])) || (i != len(rs)-1 && !isNumOrSpace(rs[i+1])) {
				rs[i] = ' '
			}
		}
	}
	name = string(rs)
	name = strings.TrimSpace(name)
	if len(name) == 0 && origLen > 0 {
		// Keep at least one character. This is important to preserve
		// empty lines in multi-line headers/footers.
		name = " "
	}
	return strings.ToUpper(name)
}

// Pad String
// Attempts to place string in the center
func Pad(s, pad string, width int) string {
	gap := width - DisplayWidth(s)
	if gap > 0 {
		gapLeft := int(math.Ceil(float64(gap / 2)))
		gapRight := gap - gapLeft
		return strings.Repeat(string(pad), gapLeft) + s + strings.Repeat(string(pad), gapRight)
	}
	return s
}

// PadRight Pad String Right position
// This would place string at the left side of the screen
func PadRight(s, pad string, width int) string {
	gap := width - DisplayWidth(s)
	if gap > 0 {
		return s + strings.Repeat(string(pad), gap)
	}
	return s
}

// PadLeft Pad String Left position
// This would place string at the right side of the screen
func PadLeft(s, pad string, width int) string {
	gap := width - DisplayWidth(s)
	if gap > 0 {
		return strings.Repeat(string(pad), gap) + s
	}
	return s
}
