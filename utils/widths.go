package utils

import "sort"

type Widths map[int]int

// Convert converts the column widths map into a slice of widths.
// numCols specifies the total number of columns expected.
// Returns a slice where each index corresponds to the column width.
func (cw Widths) Convert(numCols int) []int {
	slice := make([]int, numCols)
	for i := 0; i < numCols; i++ {
		slice[i] = cw.Get(i) // Use Get() to handle missing columns
	}
	return slice
}

// ConvertSorted returns a slice of widths sorted by column index in ascending order.
// The result only includes columns that exist in the map (no zero padding).
func (cw Widths) ConvertSorted() []int {
	// Get all column indices and sort them
	columns := make([]int, 0, len(cw))
	for col := range cw {
		columns = append(columns, col)
	}
	sort.Ints(columns)

	// Create result slice with widths in column order
	result := make([]int, 0, len(columns))
	for _, col := range columns {
		result = append(result, cw[col])
	}
	return result
}

// Get returns the width for a specific column.
// Returns 0 if the column doesn't exist in the map.
func (cw Widths) Get(col int) int {
	if width, exists := cw[col]; exists {
		return width
	}
	return 0 // Default width
}

// Set sets the width for a specific column.
func (cw Widths) Set(col, width int) {
	cw[col] = width
}

// Max returns the maximum column width in the map.
func (cw Widths) Max() int {
	max := 0
	for _, width := range cw {
		if width > max {
			max = width
		}
	}
	return max
}

// Total returns the sum of all column widths.
func (cw Widths) Total() int {
	sum := 0
	for _, width := range cw {
		sum += width
	}
	return sum
}

// Merge combines another Widths into this one,
// keeping the larger width when columns exist in both.
func (cw Widths) Merge(other Widths) {
	for col, width := range other {
		if current, exists := cw[col]; !exists || width > current {
			cw[col] = width
		}
	}
}

// Equal compares if two Widths have the same widths for the same columns.
func (cw Widths) Equal(other Widths) bool {
	if len(cw) != len(other) {
		return false
	}
	for col, width := range cw {
		if otherWidth, exists := other[col]; !exists || otherWidth != width {
			return false
		}
	}
	return true
}

// Clone creates a deep copy of the Widths.
func (cw Widths) Clone() Widths {
	newCW := make(Widths, len(cw))
	for col, width := range cw {
		newCW[col] = width
	}
	return newCW
}
