package tw

import "sort"

type Widths map[int]int

// Get returns the width for a specific column.
// Returns 0 if the column doesn't exist in the map.
func (w Widths) Get(col int) int {
	if width, exists := w[col]; exists {
		return width
	}
	return 0 // Default width
}

// Convert converts the column widths map into a slice of widths.
// numCols specifies the total number of columns expected.
// Returns a slice where each index corresponds to the column width.
func (w Widths) Convert(numCols int) []int {
	slice := make([]int, numCols)
	for i := 0; i < numCols; i++ {
		slice[i] = w.Get(i) // Use Get() to handle missing columns
	}
	return slice
}

// ConvertSorted returns a slice of widths sorted by column index in ascending order.
// The result only includes columns that exist in the map (no zero padding).
func (w Widths) ConvertSorted() []int {
	// Get all column indices and sort them
	columns := make([]int, 0, len(w))
	for col := range w {
		columns = append(columns, col)
	}
	sort.Ints(columns)

	// Create result slice with widths in column order
	result := make([]int, 0, len(columns))
	for _, col := range columns {
		result = append(result, w[col])
	}
	return result
}

// Set sets the width for a specific column.
func (w Widths) Set(col, width int) {
	w[col] = width
}

// Max returns the maximum column width in the map.
func (w Widths) Max() int {
	max := 0
	for _, width := range w {
		if width > max {
			max = width
		}
	}
	return max
}

// Total returns the sum of all column widths.
func (w Widths) Total() int {
	sum := 0
	for _, width := range w {
		sum += width
	}
	return sum
}

// Merge combines another Widths into this one,
// keeping the larger width when columns exist in both.
func (w Widths) Merge(other Widths) {
	for col, width := range other {
		if current, exists := w[col]; !exists || width > current {
			w[col] = width
		}
	}
}

// Equal compares if two Widths have the same widths for the same columns.
func (w Widths) Equal(other Widths) bool {
	if len(w) != len(other) {
		return false
	}
	for col, width := range w {
		if otherWidth, exists := other[col]; !exists || otherWidth != width {
			return false
		}
	}
	return true
}

// Clone creates a deep copy of the Widths.
func (w Widths) Clone() Widths {
	newCW := make(Widths, len(w))
	for col, width := range w {
		newCW[col] = width
	}
	return newCW
}
