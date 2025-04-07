package tablewriter

import (
	"github.com/olekukonko/tablewriter/tw"
)

// ConfigBuilder provides a fluent interface for building Config
type ConfigBuilder struct {
	config Config
}

// NewConfigBuilder creates a new ConfigBuilder with defaults
func NewConfigBuilder() *ConfigBuilder {
	return &ConfigBuilder{
		config: defaultConfig(),
	}
}

// Build returns the finalized Config
func (b *ConfigBuilder) Build() Config {
	return b.config
}

// Global settings ------------------------------------------------------------

func (b *ConfigBuilder) WithMaxWidth(width int) *ConfigBuilder {
	b.config.MaxWidth = width
	return b
}

// Header configuration -------------------------------------------------------

func (b *ConfigBuilder) Header() *HeaderConfigBuilder {
	return &HeaderConfigBuilder{
		parent:  b,
		config:  &b.config.Header,
		section: "header",
	}
}

type HeaderConfigBuilder struct {
	parent  *ConfigBuilder
	config  *CellConfig
	section string
}

func (h *HeaderConfigBuilder) Build() *ConfigBuilder {
	return h.parent
}

func (h *HeaderConfigBuilder) Formatting() *HeaderFormattingBuilder {
	return &HeaderFormattingBuilder{
		parent:  h,
		config:  &h.config.Formatting,
		section: h.section,
	}
}

type HeaderFormattingBuilder struct {
	parent  *HeaderConfigBuilder
	config  *CellFormatting
	section string
}

func (hf *HeaderFormattingBuilder) WithAlignment(align tw.Align) *HeaderFormattingBuilder {
	hf.config.Alignment = align
	return hf
}

func (hf *HeaderFormattingBuilder) WithAutoWrap(autoWrap int) *HeaderFormattingBuilder {
	hf.config.AutoWrap = autoWrap
	return hf
}

func (hf *HeaderFormattingBuilder) WithAutoFormat(autoFormat bool) *HeaderFormattingBuilder {
	hf.config.AutoFormat = autoFormat
	return hf
}

func (hf *HeaderFormattingBuilder) WithMergeMode(mergeMode int) *HeaderFormattingBuilder {
	hf.config.MergeMode = mergeMode
	return hf
}

func (hf *HeaderFormattingBuilder) WithMaxWidth(maxWidth int) *HeaderFormattingBuilder {
	hf.config.MaxWidth = maxWidth
	return hf
}

func (hf *HeaderFormattingBuilder) Build() *HeaderConfigBuilder {
	return hf.parent
}

func (h *HeaderConfigBuilder) Padding() *HeaderPaddingBuilder {
	return &HeaderPaddingBuilder{
		parent:  h,
		config:  &h.config.Padding,
		section: h.section,
	}
}

type HeaderPaddingBuilder struct {
	parent  *HeaderConfigBuilder
	config  *CellPadding
	section string
}

func (hp *HeaderPaddingBuilder) WithGlobal(padding tw.Padding) *HeaderPaddingBuilder {
	hp.config.Global = padding
	return hp
}

func (hp *HeaderPaddingBuilder) WithPerColumn(padding []tw.Padding) *HeaderPaddingBuilder {
	hp.config.PerColumn = padding
	return hp
}

func (hp *HeaderPaddingBuilder) AddColumnPadding(padding tw.Padding) *HeaderPaddingBuilder {
	hp.config.PerColumn = append(hp.config.PerColumn, padding)
	return hp
}

func (hp *HeaderPaddingBuilder) Build() *HeaderConfigBuilder {
	return hp.parent
}

// Row configuration ----------------------------------------------------------

func (b *ConfigBuilder) Row() *RowConfigBuilder {
	return &RowConfigBuilder{
		parent:  b,
		config:  &b.config.Row,
		section: "row",
	}
}

type RowConfigBuilder struct {
	parent  *ConfigBuilder
	config  *CellConfig
	section string
}

func (r *RowConfigBuilder) Build() *ConfigBuilder {
	return r.parent
}

// ... similar builder methods for Row as Header ...

// Footer configuration -------------------------------------------------------

func (b *ConfigBuilder) Footer() *FooterConfigBuilder {
	return &FooterConfigBuilder{
		parent:  b,
		config:  &b.config.Footer,
		section: "footer",
	}
}

type FooterConfigBuilder struct {
	parent  *ConfigBuilder
	config  *CellConfig
	section string
}

func (f *FooterConfigBuilder) Build() *ConfigBuilder {
	return f.parent
}

// ... similar builder methods for Footer as Header ...

// Column-specific overrides --------------------------------------------------

func (b *ConfigBuilder) ForColumn(col int) *ColumnConfigBuilder {
	return &ColumnConfigBuilder{
		parent: b,
		col:    col,
	}
}

type ColumnConfigBuilder struct {
	parent *ConfigBuilder
	col    int
}

func (c *ColumnConfigBuilder) WithMaxWidth(width int) *ColumnConfigBuilder {
	// Initialize maps if needed
	if c.parent.config.Header.ColMaxWidths == nil {
		c.parent.config.Header.ColMaxWidths = make(map[int]int)
		c.parent.config.Row.ColMaxWidths = make(map[int]int)
		c.parent.config.Footer.ColMaxWidths = make(map[int]int)
	}

	c.parent.config.Header.ColMaxWidths[c.col] = width
	c.parent.config.Row.ColMaxWidths[c.col] = width
	c.parent.config.Footer.ColMaxWidths[c.col] = width
	return c
}

func (c *ColumnConfigBuilder) WithAlignment(align tw.Align) *ColumnConfigBuilder {
	// Ensure slice is large enough
	if len(c.parent.config.Header.ColumnAligns) <= c.col {
		newAligns := make([]tw.Align, c.col+1)
		copy(newAligns, c.parent.config.Header.ColumnAligns)
		c.parent.config.Header.ColumnAligns = newAligns
	}

	c.parent.config.Header.ColumnAligns[c.col] = align
	return c
}

func (c *ColumnConfigBuilder) Build() *ConfigBuilder {
	return c.parent
}

//// Example usage:
//func ExampleBuilderUsage() {
//	cfg := NewConfigBuilder().
//		WithMaxWidth(120).
//		Header().
//		Formatting().
//		WithAlignment(AlignCenter).
//		WithAutoWrap(true).
//		Build().
//		Padding().
//		WithGlobal(symbols.Padding{Left: " ", Right: " "}).
//		Build().
//		Build().
//		Row().
//		Formatting().
//		WithAlignment(AlignLeft).
//		Build().
//		Build().
//		ForColumn(0).
//		WithMaxWidth(30).
//		WithAlignment(AlignRight).
//		Build().
//		Build()
//
//	_ = cfg // use the config
//}
