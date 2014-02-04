package table

import (
	"os"
	"testing"
)

func TestLong(t *testing.T) {

	data := [][]string{
		[]string{"Learn East has computers with adapted keyboards with enlarged print etc. There is also a portable loop system and a minicom", "  y  ", " y "},
		[]string{"Instead of lining up the letters all the way across, he splits the keyboard in two", "Like most ergonomic keyboards", "x"},
	}


		data = [][]string{
			[]string{"The Good man", "B", "500"},
			[]string{"D", "The Very very Bad Man", "288"},
			[]string{"D", " E ", "288"},
		}



	table := NewTable(os.Stdout)
	table.SetHeader([]string{"Name", "Sign", "Rating"})
	table.SetCenterSeparator("*")
	table.SetRowSeparator("=");

	for _, v := range data {
		table.Append(v)
	}
	table.Render()

}
