package tablewriter

import (
	"github.com/olekukonko/tablewriter/tw"
	"io"
)

// ConfigBuilder provides a fluent interface for building a Config struct.
// It combines direct methods for common settings with nested builders for advanced configuration.
type ConfigBuilder struct {
	config Config // The configuration being built
}

// NewConfigBuilder creates a new ConfigBuilder initialized with default settings.
func NewConfigBuilder() *ConfigBuilder {
	return &ConfigBuilder{
		config: defaultConfig(),
	}
}

// Build returns the finalized Config struct after all modifications.
func (b *ConfigBuilder) Build() Config {
	return b.config
}

// --- Global Configuration Methods ---

// WithMaxWidth sets the maximum width for the entire table (0 means unlimited).
func (b *ConfigBuilder) WithMaxWidth(width int) *ConfigBuilder {
	b.config.MaxWidth = width
	return b
}

// WithDebug enables or disables debug logging for the table.
func (b *ConfigBuilder) WithDebug(debug bool) *ConfigBuilder {
	b.config.Debug = debug
	return b
}

// --- Direct Header Configuration Methods ---

// WithHeaderAlignment sets the text alignment for all header cells.
func (b *ConfigBuilder) WithHeaderAlignment(align tw.Align) *ConfigBuilder {
	b.config.Header.Formatting.Alignment = align
	return b
}

// WithHeaderAutoWrap sets the wrapping behavior for header cells (e.g., WrapNormal, WrapTruncate).
func (b *ConfigBuilder) WithHeaderAutoWrap(autoWrap int) *ConfigBuilder {
	b.config.Header.Formatting.AutoWrap = autoWrap
	return b
}

// WithHeaderAutoFormat enables or disables automatic formatting (e.g., title case) for header cells.
func (b *ConfigBuilder) WithHeaderAutoFormat(autoFormat bool) *ConfigBuilder {
	b.config.Header.Formatting.AutoFormat = autoFormat
	return b
}

// WithHeaderMergeMode sets the merge behavior for header cells (e.g., MergeHorizontal, MergeVertical).
func (b *ConfigBuilder) WithHeaderMergeMode(mergeMode int) *ConfigBuilder {
	b.config.Header.Formatting.MergeMode = mergeMode
	return b
}

// WithHeaderMaxWidth sets the maximum content width for header cells.
func (b *ConfigBuilder) WithHeaderMaxWidth(maxWidth int) *ConfigBuilder {
	b.config.Header.Formatting.MaxWidth = maxWidth
	return b
}

// WithHeaderGlobalPadding sets the global padding for all header cells.
func (b *ConfigBuilder) WithHeaderGlobalPadding(padding tw.Padding) *ConfigBuilder {
	b.config.Header.Padding.Global = padding
	return b
}

// --- Direct Row Configuration Methods ---

// WithRowAlignment sets the text alignment for all row cells.
func (b *ConfigBuilder) WithRowAlignment(align tw.Align) *ConfigBuilder {
	b.config.Row.Formatting.Alignment = align
	return b
}

// WithRowAutoWrap sets the wrapping behavior for row cells.
func (b *ConfigBuilder) WithRowAutoWrap(autoWrap int) *ConfigBuilder {
	b.config.Row.Formatting.AutoWrap = autoWrap
	return b
}

// WithRowAutoFormat enables or disables automatic formatting for row cells.
func (b *ConfigBuilder) WithRowAutoFormat(autoFormat bool) *ConfigBuilder {
	b.config.Row.Formatting.AutoFormat = autoFormat
	return b
}

// WithRowMergeMode sets the merge behavior for row cells.
func (b *ConfigBuilder) WithRowMergeMode(mergeMode int) *ConfigBuilder {
	b.config.Row.Formatting.MergeMode = mergeMode
	return b
}

// WithRowMaxWidth sets the maximum content width for row cells.
func (b *ConfigBuilder) WithRowMaxWidth(maxWidth int) *ConfigBuilder {
	b.config.Row.Formatting.MaxWidth = maxWidth
	return b
}

// WithRowGlobalPadding sets the global padding for all row cells.
func (b *ConfigBuilder) WithRowGlobalPadding(padding tw.Padding) *ConfigBuilder {
	b.config.Row.Padding.Global = padding
	return b
}

// --- Direct Footer Configuration Methods ---

// WithFooterAlignment sets the text alignment for all footer cells.
func (b *ConfigBuilder) WithFooterAlignment(align tw.Align) *ConfigBuilder {
	b.config.Footer.Formatting.Alignment = align
	return b
}

// WithFooterAutoWrap sets the wrapping behavior for footer cells.
func (b *ConfigBuilder) WithFooterAutoWrap(autoWrap int) *ConfigBuilder {
	b.config.Footer.Formatting.AutoWrap = autoWrap
	return b
}

// WithFooterAutoFormat enables or disables automatic formatting for footer cells.
func (b *ConfigBuilder) WithFooterAutoFormat(autoFormat bool) *ConfigBuilder {
	b.config.Footer.Formatting.AutoFormat = autoFormat
	return b
}

// WithFooterMergeMode sets the merge behavior for footer cells.
func (b *ConfigBuilder) WithFooterMergeMode(mergeMode int) *ConfigBuilder {
	b.config.Footer.Formatting.MergeMode = mergeMode
	return b
}

// WithFooterMaxWidth sets the maximum content width for footer cells.
func (b *ConfigBuilder) WithFooterMaxWidth(maxWidth int) *ConfigBuilder {
	b.config.Footer.Formatting.MaxWidth = maxWidth
	return b
}

// WithFooterGlobalPadding sets the global padding for all footer cells.
func (b *ConfigBuilder) WithFooterGlobalPadding(padding tw.Padding) *ConfigBuilder {
	b.config.Footer.Padding.Global = padding
	return b
}

// --- Nested Builders for Advanced Configuration ---

// Header returns a builder for advanced header configuration.
func (b *ConfigBuilder) Header() *HeaderConfigBuilder {
	return &HeaderConfigBuilder{
		parent:  b,
		config:  &b.config.Header,
		section: "header",
	}
}

// Row returns a builder for advanced row configuration.
func (b *ConfigBuilder) Row() *RowConfigBuilder {
	return &RowConfigBuilder{
		parent:  b,
		config:  &b.config.Row,
		section: "row",
	}
}

// Footer returns a builder for advanced footer configuration.
func (b *ConfigBuilder) Footer() *FooterConfigBuilder {
	return &FooterConfigBuilder{
		parent:  b,
		config:  &b.config.Footer,
		section: "footer",
	}
}

// ForColumn returns a builder for column-specific overrides across all sections.
func (b *ConfigBuilder) ForColumn(col int) *ColumnConfigBuilder {
	return &ColumnConfigBuilder{
		parent: b,
		col:    col,
	}
}

// --- Nested Builder Definitions ---

// HeaderConfigBuilder provides advanced configuration for the header section.
type HeaderConfigBuilder struct {
	parent  *ConfigBuilder
	config  *CellConfig
	section string
}

// Build returns to the parent ConfigBuilder.
func (h *HeaderConfigBuilder) Build() *ConfigBuilder {
	return h.parent
}

// Formatting returns a builder for header formatting settings.
func (h *HeaderConfigBuilder) Formatting() *HeaderFormattingBuilder {
	return &HeaderFormattingBuilder{
		parent:  h,
		config:  &h.config.Formatting,
		section: h.section,
	}
}

// Padding returns a builder for header padding settings.
func (h *HeaderConfigBuilder) Padding() *HeaderPaddingBuilder {
	return &HeaderPaddingBuilder{
		parent:  h,
		config:  &h.config.Padding,
		section: h.section,
	}
}

// HeaderFormattingBuilder configures formatting options for the header.
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

// HeaderPaddingBuilder configures padding options for the header.
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

// RowConfigBuilder provides advanced configuration for the row section.
// (Similar structure to HeaderConfigBuilder, omitted for brevity but follows the same pattern)
type RowConfigBuilder struct {
	parent  *ConfigBuilder
	config  *CellConfig
	section string
}

func (r *RowConfigBuilder) Build() *ConfigBuilder {
	return r.parent
}

func (r *RowConfigBuilder) Formatting() *RowFormattingBuilder {
	return &RowFormattingBuilder{
		parent:  r,
		config:  &r.config.Formatting,
		section: r.section,
	}
}

func (r *RowConfigBuilder) Padding() *RowPaddingBuilder {
	return &RowPaddingBuilder{
		parent:  r,
		config:  &r.config.Padding,
		section: r.section,
	}
}

// RowFormattingBuilder configures formatting options for rows.
type RowFormattingBuilder struct {
	parent  *RowConfigBuilder
	config  *CellFormatting
	section string
}

func (rf *RowFormattingBuilder) WithAlignment(align tw.Align) *RowFormattingBuilder {
	rf.config.Alignment = align
	return rf
}

func (rf *RowFormattingBuilder) WithAutoWrap(autoWrap int) *RowFormattingBuilder {
	rf.config.AutoWrap = autoWrap
	return rf
}

func (rf *RowFormattingBuilder) WithAutoFormat(autoFormat bool) *RowFormattingBuilder {
	rf.config.AutoFormat = autoFormat
	return rf
}

func (rf *RowFormattingBuilder) WithMergeMode(mergeMode int) *RowFormattingBuilder {
	rf.config.MergeMode = mergeMode
	return rf
}

func (rf *RowFormattingBuilder) WithMaxWidth(maxWidth int) *RowFormattingBuilder {
	rf.config.MaxWidth = maxWidth
	return rf
}

func (rf *RowFormattingBuilder) Build() *RowConfigBuilder {
	return rf.parent
}

// RowPaddingBuilder configures padding options for rows.
type RowPaddingBuilder struct {
	parent  *RowConfigBuilder
	config  *CellPadding
	section string
}

func (rp *RowPaddingBuilder) WithGlobal(padding tw.Padding) *RowPaddingBuilder {
	rp.config.Global = padding
	return rp
}

func (rp *RowPaddingBuilder) WithPerColumn(padding []tw.Padding) *RowPaddingBuilder {
	rp.config.PerColumn = padding
	return rp
}

func (rp *RowPaddingBuilder) AddColumnPadding(padding tw.Padding) *RowPaddingBuilder {
	rp.config.PerColumn = append(rp.config.PerColumn, padding)
	return rp
}

func (rp *RowPaddingBuilder) Build() *RowConfigBuilder {
	return rp.parent
}

// FooterConfigBuilder provides advanced configuration for the footer section.
// (Similar structure to HeaderConfigBuilder, omitted for brevity)
type FooterConfigBuilder struct {
	parent  *ConfigBuilder
	config  *CellConfig
	section string
}

func (f *FooterConfigBuilder) Build() *ConfigBuilder {
	return f.parent
}

func (f *FooterConfigBuilder) Formatting() *FooterFormattingBuilder {
	return &FooterFormattingBuilder{
		parent:  f,
		config:  &f.config.Formatting,
		section: f.section,
	}
}

func (f *FooterConfigBuilder) Padding() *FooterPaddingBuilder {
	return &FooterPaddingBuilder{
		parent:  f,
		config:  &f.config.Padding,
		section: f.section,
	}
}

// FooterFormattingBuilder configures formatting options for the footer.
type FooterFormattingBuilder struct {
	parent  *FooterConfigBuilder
	config  *CellFormatting
	section string
}

func (ff *FooterFormattingBuilder) WithAlignment(align tw.Align) *FooterFormattingBuilder {
	ff.config.Alignment = align
	return ff
}

func (ff *FooterFormattingBuilder) WithAutoWrap(autoWrap int) *FooterFormattingBuilder {
	ff.config.AutoWrap = autoWrap
	return ff
}

func (ff *FooterFormattingBuilder) WithAutoFormat(autoFormat bool) *FooterFormattingBuilder {
	ff.config.AutoFormat = autoFormat
	return ff
}

func (ff *FooterFormattingBuilder) WithMergeMode(mergeMode int) *FooterFormattingBuilder {
	ff.config.MergeMode = mergeMode
	return ff
}

func (ff *FooterFormattingBuilder) WithMaxWidth(maxWidth int) *FooterFormattingBuilder {
	ff.config.MaxWidth = maxWidth
	return ff
}

func (ff *FooterFormattingBuilder) Build() *FooterConfigBuilder {
	return ff.parent
}

// FooterPaddingBuilder configures padding options for the footer.
type FooterPaddingBuilder struct {
	parent  *FooterConfigBuilder
	config  *CellPadding
	section string
}

func (fp *FooterPaddingBuilder) WithGlobal(padding tw.Padding) *FooterPaddingBuilder {
	fp.config.Global = padding
	return fp
}

func (fp *FooterPaddingBuilder) WithPerColumn(padding []tw.Padding) *FooterPaddingBuilder {
	fp.config.PerColumn = padding
	return fp
}

func (fp *FooterPaddingBuilder) AddColumnPadding(padding tw.Padding) *FooterPaddingBuilder {
	fp.config.PerColumn = append(fp.config.PerColumn, padding)
	return fp
}

func (fp *FooterPaddingBuilder) Build() *FooterConfigBuilder {
	return fp.parent
}

// ColumnConfigBuilder configures column-specific overrides across all sections.
type ColumnConfigBuilder struct {
	parent *ConfigBuilder
	col    int
}

// WithMaxWidth sets the maximum width for a specific column in all sections.
func (c *ColumnConfigBuilder) WithMaxWidth(width int) *ColumnConfigBuilder {
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

// WithAlignment sets the alignment for a specific column in the header section only.
// (Could be expanded to affect all sections if desired)
func (c *ColumnConfigBuilder) WithAlignment(align tw.Align) *ColumnConfigBuilder {
	if len(c.parent.config.Header.ColumnAligns) <= c.col {
		newAligns := make([]tw.Align, c.col+1)
		copy(newAligns, c.parent.config.Header.ColumnAligns)
		c.parent.config.Header.ColumnAligns = newAligns
	}
	c.parent.config.Header.ColumnAligns[c.col] = align
	return c
}

// Build returns to the parent ConfigBuilder.
func (c *ColumnConfigBuilder) Build() *ConfigBuilder {
	return c.parent
}

// --- Configuration Merging ---

// mergeConfig merges a source Config into a destination Config, preserving non-zero values.
func mergeConfig(dst, src Config) Config {
	t := &Table{config: dst}
	t.debug("Merging config: src.MaxWidth=%d", src.MaxWidth)
	if src.MaxWidth != 0 {
		dst.MaxWidth = src.MaxWidth
	}
	dst.Header = mergeCellConfig(dst.Header, src.Header)
	dst.Row = mergeCellConfig(dst.Row, src.Row)
	dst.Footer = mergeCellConfig(dst.Footer, src.Footer)
	dst.Debug = src.Debug || dst.Debug // Debug is a special case, merge with OR
	t.debug("Config merged")
	return dst
}

// mergeCellConfig merges a source CellConfig into a destination CellConfig.
func mergeCellConfig(dst, src CellConfig) CellConfig {
	t := &Table{config: Config{Debug: true}}
	t.debug("Merging cell config")
	if src.Formatting.Alignment != tw.Empty {
		dst.Formatting.Alignment = src.Formatting.Alignment
	}
	if src.Formatting.AutoWrap != 0 {
		dst.Formatting.AutoWrap = src.Formatting.AutoWrap
	}
	if src.Formatting.MaxWidth != 0 {
		dst.Formatting.MaxWidth = src.Formatting.MaxWidth
	}
	if src.Formatting.MergeMode != 0 {
		dst.Formatting.MergeMode = src.Formatting.MergeMode
	}
	dst.Formatting.AutoFormat = src.Formatting.AutoFormat || dst.Formatting.AutoFormat

	if src.Padding.Global != (tw.Padding{}) {
		dst.Padding.Global = src.Padding.Global
	}
	if len(src.Padding.PerColumn) > 0 {
		if dst.Padding.PerColumn == nil {
			dst.Padding.PerColumn = make([]tw.Padding, len(src.Padding.PerColumn))
		}
		for i, pad := range src.Padding.PerColumn {
			if pad != (tw.Padding{}) {
				if i < len(dst.Padding.PerColumn) {
					dst.Padding.PerColumn[i] = pad
				} else {
					dst.Padding.PerColumn = append(dst.Padding.PerColumn, pad)
				}
			}
		}
	}

	if src.Callbacks.Global != nil {
		dst.Callbacks.Global = src.Callbacks.Global
	}
	if len(src.Callbacks.PerColumn) > 0 {
		if dst.Callbacks.PerColumn == nil {
			dst.Callbacks.PerColumn = make([]func(), len(src.Callbacks.PerColumn))
		}
		for i, cb := range src.Callbacks.PerColumn {
			if cb != nil {
				if i < len(dst.Callbacks.PerColumn) {
					dst.Callbacks.PerColumn[i] = cb
				} else {
					dst.Callbacks.PerColumn = append(dst.Callbacks.PerColumn, cb)
				}
			}
		}
	}

	if src.Filter.Global != nil {
		dst.Filter.Global = src.Filter.Global
	}
	if len(src.Filter.PerColumn) > 0 {
		if dst.Filter.PerColumn == nil {
			dst.Filter.PerColumn = make([]func(string) string, len(src.Filter.PerColumn))
		}
		for i, filter := range src.Filter.PerColumn {
			if filter != nil {
				if i < len(dst.Filter.PerColumn) {
					dst.Filter.PerColumn[i] = filter
				} else {
					dst.Filter.PerColumn = append(dst.Filter.PerColumn, filter)
				}
			}
		}
	}

	if len(src.ColumnAligns) > 0 {
		dst.ColumnAligns = src.ColumnAligns
	}

	if len(src.ColMaxWidths) > 0 {
		if dst.ColMaxWidths == nil {
			dst.ColMaxWidths = make(map[int]int)
		}
		for k, v := range src.ColMaxWidths {
			dst.ColMaxWidths[k] = v
		}
	}

	t.debug("Cell config merged")
	return dst
}

// --- Option Functions ---

// Option defines a function to configure a Table instance.
type Option func(*Table)

// WithHeader sets the table headers.
func WithHeader(headers []string) Option {
	return func(t *Table) { t.SetHeader(headers) }
}

// WithFooter sets the table footers.
func WithFooter(footers []string) Option {
	return func(t *Table) { t.SetFooter(footers) }
}

// WithRenderer sets a custom renderer for the table.
func WithRenderer(f tw.Renderer) Option {
	return func(t *Table) { t.renderer = f }
}

// WithConfig applies a custom configuration to the table.
func WithConfig(cfg Config) Option {
	return func(t *Table) { t.config = mergeConfig(defaultConfig(), cfg) }
}

// WithStringer sets a custom stringer function for row conversion.
func WithStringer[T any](s func(T) []string) Option {
	return func(t *Table) { t.stringer = s }
}

// WithDebug enables or disables debug logging.
func WithDebug(debug bool) Option {
	return func(t *Table) { t.config.Debug = debug }
}

// WithHeaderAlignment sets the header alignment directly.
func WithHeaderAlignment(align tw.Align) Option {
	return func(t *Table) { t.config.Header.Formatting.Alignment = align }
}

// WithRowMaxWidth sets the row max width directly.
func WithRowMaxWidth(maxWidth int) Option {
	return func(t *Table) { t.config.Row.Formatting.MaxWidth = maxWidth }
}

// WithFooterMergeMode sets the footer merge mode directly.
func WithFooterMergeMode(mergeMode int) Option {
	return func(t *Table) { t.config.Footer.Formatting.MergeMode = mergeMode }
}

// WithHeaderConfig applies a full header configuration.
func WithHeaderConfig(config CellConfig) Option {
	return func(t *Table) { t.config.Header = config }
}

// WithRowConfig applies a full row configuration.
func WithRowConfig(config CellConfig) Option {
	return func(t *Table) { t.config.Row = config }
}

// WithFooterConfig applies a full footer configuration.
func WithFooterConfig(config CellConfig) Option {
	return func(t *Table) { t.config.Footer = config }
}

// --- Legacy Compatibility ---

// NewWriter creates a new table with default settings for backward compatibility.
func NewWriter(w io.Writer) *Table {
	t := NewTable(w)
	t.debug("NewWriter created table")
	return t
}
