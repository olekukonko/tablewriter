// Copyright 2014 Oleku Konko All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// This module is a Table Writer  API for the Go Programming Language.
// The protocols were written in pure Go and works on windows and unix systems

package tablewriter

import (
	"bytes"
	"os"
	"testing"
)

func TestSetHeaderColorTTY(t *testing.T) {
	data := [][]string{
		{"A", "The Good", "500"},
		{"B", "The Very very Bad Man", "288"},
		{"C", "The Ugly", "120"},
		{"D", "The Gopher", "800"},
	}

	var buf bytes.Buffer
	table := NewWriter(&buf)
	table.SetHeader([]string{"Name", "Sign", "Rating"})
	table.SetHeaderColor(Colors{Bold, FgHiYellowColor}, Colors{Bold, FgHiYellowColor}, Colors{Bold, FgHiYellowColor})

	for _, v := range data {
		table.Append(v)
	}

	table.SetAutoMergeCells(true)
	table.SetRowLine(true)
	table.Render()

	// table := NewWriter(os.Stdout)
	var want []string
	want = append(want, "1;93", "1;93", "1;93")

	// The color codes are added in case of TTY output.
	got := table.headerParams

	checkEqual(t, got, want, "SetHeaderColor when TTY is attached failed")
}

func createTempFile(t *testing.T) *os.File {
	tempFile, err := os.CreateTemp("", "output-*.txt")
	if err != nil {
		t.Fatalf("Failed to create a temporary file: %v", err)
	}
	return tempFile
}

func TestSetHeaderColorNonTTY(t *testing.T) {
	data := [][]string{
		{"A", "The Good", "500"},
		{"B", "The Very very Bad Man", "288"},
		{"C", "The Ugly", "120"},
		{"D", "The Gopher", "800"},
	}

	os.Stdout = createTempFile(t)
	table := NewWriter(os.Stdout)

	table.SetHeader([]string{"Name", "Sign", "Rating"})
	want := table.headerParams
	table.SetHeaderColor(Colors{Bold, FgHiYellowColor}, Colors{Bold, FgHiYellowColor}, Colors{Bold, FgHiYellowColor})

	for _, v := range data {
		table.Append(v)
	}

	table.SetAutoMergeCells(true)
	table.SetRowLine(true)
	table.Render()

	// The color codes are not added in case of non TTY output.
	got := table.headerParams

	checkEqual(t, got, want, "SetHeaderColor when TTY is not attached failed")
}
