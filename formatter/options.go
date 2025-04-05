// formatter/options.go
package formatter

import (
	"github.com/olekukonko/tablewriter/symbols"
)

// Option defines a configuration function for formatters
type Option func(Formatter)

// WithBorders sets the table borders
func WithBorders(borders Border) Option {
	return func(f Formatter) {
		switch ff := f.(type) {
		case *DefaultFormatter:
			ff.borders = borders
		case *ColorFormatter:
			ff.borders = borders
		}
	}
}

// WithHeaderAlignment sets the header alignment
func WithHeaderAlignment(align int) Option {
	return func(f Formatter) {
		switch ff := f.(type) {
		case *DefaultFormatter:
			ff.headerAlignment = align
		case *ColorFormatter:
			ff.headerAlignment = align
		}
	}
}

// WithFooterAlignment sets the footer alignment
func WithFooterAlignment(align int) Option {
	return func(f Formatter) {
		switch ff := f.(type) {
		case *DefaultFormatter:
			ff.footerAlignment = align
		case *ColorFormatter:
			ff.footerAlignment = align
		}
	}
}

// WithAlignment sets the default cell alignment
func WithAlignment(align int) Option {
	return func(f Formatter) {
		switch ff := f.(type) {
		case *DefaultFormatter:
			ff.alignment = align
		case *MarkdownFormatter:
			ff.alignment = align
		case *ColorFormatter:
			ff.alignment = align
		}
	}
}

// WithHeaderLine enables/disables the header line
func WithHeaderLine(enabled bool) Option {
	return func(f Formatter) {
		switch ff := f.(type) {
		case *DefaultFormatter:
			ff.headerLine = enabled
		case *ColorFormatter:
			ff.headerLine = enabled
		}
	}
}

// WithCenterSeparator sets the center separator
func WithCenterSeparator(sep string) Option {
	return func(f Formatter) {
		switch ff := f.(type) {
		case *DefaultFormatter:
			ff.centerSeparator = sep
			ff.updateSymbols()
		case *ColorFormatter:
			ff.centerSeparator = sep
			ff.updateSymbols()
		}
	}
}

// WithRowSeparator sets the row separator
func WithRowSeparator(sep string) Option {
	return func(f Formatter) {
		switch ff := f.(type) {
		case *DefaultFormatter:
			ff.rowSeparator = sep
			ff.updateSymbols()
		case *ColorFormatter:
			ff.rowSeparator = sep
			ff.updateSymbols()
		}
	}
}

// WithColumnSeparator sets the column separator
func WithColumnSeparator(sep string) Option {
	return func(f Formatter) {
		switch ff := f.(type) {
		case *DefaultFormatter:
			ff.columnSeparator = sep
			ff.updateSymbols()
		case *ColorFormatter:
			ff.columnSeparator = sep
			ff.updateSymbols()
		}
	}
}

// WithAutoFormatHeaders enables/disables auto-formatting of headers
func WithAutoFormatHeaders(auto bool) Option {
	return func(f Formatter) {
		if df, ok := f.(*DefaultFormatter); ok {
			df.autoFormat = auto
		}
	}
}

// WithSymbols sets custom symbols
func WithSymbols(s symbols.Symbols) Option {
	return func(f Formatter) {
		switch ff := f.(type) {
		case *DefaultFormatter:
			ff.symbols = s
			ff.updateSymbols()
		case *ColorFormatter:
			ff.symbols = s
			ff.updateSymbols()
		}
	}
}

// WithHeaderColors sets header colors (ColorFormatter-specific)
func WithHeaderColors(colors []Colors) Option {
	return func(f Formatter) {
		if cf, ok := f.(*ColorFormatter); ok {
			cf.headerColors = colors
		}
	}
}

// WithColumnColors sets column colors (ColorFormatter-specific)
func WithColumnColors(colors []Colors) Option {
	return func(f Formatter) {
		if cf, ok := f.(*ColorFormatter); ok {
			cf.columnColors = colors
		}
	}
}

// WithFooterColors sets footer colors (ColorFormatter-specific)
func WithFooterColors(colors []Colors) Option {
	return func(f Formatter) {
		if cf, ok := f.(*ColorFormatter); ok {
			cf.footerColors = colors
		}
	}
}
