// Copyright 2014 Oleku Konko All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// This module is a Table Writer  API for the Go Programming Language.
// The protocols were written in pure Go and works on windows and unix systems

package tablewriter

import (
	"os"
	"testing"
)

func ExampleShort() {

	data := [][]string{
		[]string{"A", "The Good", "500"},
		[]string{"B", "The Very very Bad Man", "288"},
		[]string{"C", "The Ugly", "120"},
		[]string{"D", "The Gopher", "800"},
	}

	table := NewWriter(os.Stdout)
	table.SetHeader([]string{"Name", "Sign", "Rating"})

	for _, v := range data {
		table.Append(v)
	}
	table.Render()

}

func ExampleLong() {

	data := [][]string{
		[]string{"Learn East has computers with adapted keyboards with enlarged print etc", "  Some Data  ", " Another Data"},
		[]string{"Instead of lining up the letters all ", "the way across, he splits the keyboard in two", "Like most ergonomic keyboards", "See Data"},
	}

	table := NewWriter(os.Stdout)
	table.SetHeader([]string{"Name", "Sign", "Rating"})
	table.SetCenterSeparator("*")
	table.SetRowSeparator("=")

	for _, v := range data {
		table.Append(v)
	}
	table.Render()

}

func ExampleCSV() {
	table, _ := NewCSV(os.Stdout, "test.csv")
	table.SetCenterSeparator("*")
	table.SetRowSeparator("=")
	table.Render()
}


func TestCSV(t *testing.T) {
	table, _ := NewCSV(os.Stdout, "test.csv")
	table.SetRowLine(true)
	table.SetCenterSeparator("*")
	table.SetColumnSeparator("‡")
	table.SetRowSeparator("-")
	table.SetAlignment(ALIGN_LEFT)
	table.Render()
}


func TestCSVInfo(t *testing.T) {
	table, _ := NewCSV(os.Stdout, "test_info.csv")
	table.SetAlignment(ALIGN_LEFT)
	table.Render()
}




