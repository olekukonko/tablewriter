package tw

import "github.com/olekukonko/errors"

// Position Type and Constants
// Position defines where formatting applies in the table
type Position string

func (pos Position) Validate() error {
	switch pos {
	case Header, Footer, Row:
		return nil
	}

	return errors.New("invalid position")
}

// Filter defines a function type for processing cell content.
// It takes a slice of strings and returns a processed slice.
type Filter func([]string) []string

type Formatter interface {
	Format() string
}
