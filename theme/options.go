// formatter/options.go
package theme

import (
	"github.com/olekukonko/tablewriter/symbols"
)

// Option defines a configuration function for formatters
type Option func(Structure)

// WithBorders sets the table borders
func WithBorders(borders Border) Option {
	return func(f Structure) {
		switch ff := f.(type) {
		case *DefaultFormatter:
			ff.borders = borders
		case *Colorized:
			ff.borders = borders
		}
	}
}

// WithHeaderAlignment sets the header alignment
func WithHeaderAlignment(align int) Option {
	return func(f Structure) {
		switch ff := f.(type) {
		case *DefaultFormatter:
			ff.headerAlignment = align
		case *Colorized:
			ff.headerAlignment = align
		}
	}
}

// WithFooterAlignment sets the footer alignment
func WithFooterAlignment(align int) Option {
	return func(f Structure) {
		switch ff := f.(type) {
		case *DefaultFormatter:
			ff.footerAlignment = align
		case *Colorized:
			ff.footerAlignment = align
		}
	}
}

// WithAlignment sets the default cell alignment
func WithAlignment(align int) Option {
	return func(f Structure) {
		switch ff := f.(type) {
		case *DefaultFormatter:
			ff.alignment = align
		case *Markdown:
			ff.alignment = align
		case *Colorized:
			ff.alignment = align
		}
	}
}

// WithHeaderLine enables/disables the header line
func WithHeaderLine(enabled bool) Option {
	return func(f Structure) {
		switch ff := f.(type) {
		case *DefaultFormatter:
			ff.headerLine = enabled
		case *Colorized:
			ff.headerLine = enabled
		}
	}
}

// WithCenterSeparator sets the center separator
func WithCenterSeparator(sep string) Option {
	return func(f Structure) {
		switch ff := f.(type) {
		case *DefaultFormatter:
			ff.centerSeparator = sep
			ff.updateSymbols()
		case *Colorized:
			ff.centerSeparator = sep
			ff.updateSymbols()
		}
	}
}

// WithRowSeparator sets the row separator
func WithRowSeparator(sep string) Option {
	return func(f Structure) {
		switch ff := f.(type) {
		case *DefaultFormatter:
			ff.rowSeparator = sep
			ff.updateSymbols()
		case *Colorized:
			ff.rowSeparator = sep
			ff.updateSymbols()
		}
	}
}

// WithColumnSeparator sets the column separator
func WithColumnSeparator(sep string) Option {
	return func(f Structure) {
		switch ff := f.(type) {
		case *DefaultFormatter:
			ff.columnSeparator = sep
			ff.updateSymbols()
		case *Colorized:
			ff.columnSeparator = sep
			ff.updateSymbols()
		}
	}
}

// WithAutoFormatHeaders enables/disables auto-formatting of headers
func WithAutoFormatHeaders(auto bool) Option {
	return func(f Structure) {
		if df, ok := f.(*DefaultFormatter); ok {
			df.autoFormat = auto
		}
	}
}

// WithSymbols sets custom symbols
func WithSymbols(s symbols.Symbols) Option {
	return func(f Structure) {
		switch ff := f.(type) {
		case *DefaultFormatter:
			ff.symbols = s
			ff.updateSymbols()
		case *Colorized:
			ff.symbols = s
			ff.updateSymbols()
		}
	}
}

// WithHeaderColors sets header colors (Colorized-specific)
func WithHeaderColors(colors []Colors) Option {
	return func(f Structure) {
		if cf, ok := f.(*Colorized); ok {
			cf.headerColors = colors
		}
	}
}

// WithColumnColors sets column colors (Colorized-specific)
func WithColumnColors(colors []Colors) Option {
	return func(f Structure) {
		if cf, ok := f.(*Colorized); ok {
			cf.columnColors = colors
		}
	}
}

// WithFooterColors sets footer colors (Colorized-specific)
func WithFooterColors(colors []Colors) Option {
	return func(f Structure) {
		if cf, ok := f.(*Colorized); ok {
			cf.footerColors = colors
		}
	}
}
