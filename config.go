package tablewriter

import (
	"github.com/olekukonko/ll"
	"github.com/olekukonko/ll/lx"
	"github.com/olekukonko/tablewriter/tw"
	"io"
)

// Config represents the overall configuration for a table.
type Config struct {
	MaxWidth int             // Maximum width of the entire table (0 for unlimited)
	Header   tw.CellConfig   // Configuration for the header section
	Row      tw.CellConfig   // Configuration for the row section
	Footer   tw.CellConfig   // Configuration for the footer section
	Debug    bool            // Enables debug logging when true
	Stream   tw.StreamConfig // Configuration only valid for stream
	AutoHide bool            // Auto-hide empty columns (ignored when Stream.Enable is true)
}

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

// --- Global Configuration Methods (for ConfigBuilder) ---

// WithMaxWidth sets the maximum width for the entire table (0 means unlimited).
// Negative values are treated as 0.
func (b *ConfigBuilder) WithMaxWidth(width int) *ConfigBuilder {
	if width < 0 {
		b.config.MaxWidth = 0
	} else {
		b.config.MaxWidth = width
	}
	return b
}

// WithDebug enables or disables debug logging for the table.
func (b *ConfigBuilder) WithDebug(debug bool) *ConfigBuilder {
	b.config.Debug = debug
	return b
}

// WithAutoHide enables or disables automatic hiding of empty columns.
// This is ignored when streaming is enabled (Stream.Enable is true).
func (b *ConfigBuilder) WithAutoHide(hide bool) *ConfigBuilder {
	b.config.AutoHide = hide
	return b
}

// --- Direct Header Configuration Methods (for ConfigBuilder) ---

// WithHeaderAlignment sets the text alignment for all header cells.
func (b *ConfigBuilder) WithHeaderAlignment(align tw.Align) *ConfigBuilder {
	if align != tw.AlignLeft && align != tw.AlignRight && align != tw.AlignCenter && align != tw.AlignNone {
		return b // Ignore invalid alignment
	}
	b.config.Header.Formatting.Alignment = align
	return b
}

// WithHeaderAutoWrap sets the wrapping behavior for header cells (e.g., WrapNormal, WrapTruncate).
func (b *ConfigBuilder) WithHeaderAutoWrap(autoWrap int) *ConfigBuilder {
	if autoWrap < tw.WrapNone || autoWrap > tw.WrapBreak {
		return b // Ignore invalid wrap mode
	}
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
	if mergeMode < tw.MergeNone || mergeMode > tw.MergeHierarchical {
		return b // Ignore invalid merge mode
	}
	b.config.Header.Formatting.MergeMode = mergeMode
	return b
}

// WithHeaderMaxWidth sets the maximum content width for header cells.
func (b *ConfigBuilder) WithHeaderMaxWidth(maxWidth int) *ConfigBuilder {
	if maxWidth < 0 {
		return b // Ignore negative width
	}
	b.config.Header.Formatting.MaxWidth = maxWidth
	return b
}

// WithHeaderGlobalPadding sets the global padding for all header cells.
func (b *ConfigBuilder) WithHeaderGlobalPadding(padding tw.Padding) *ConfigBuilder {
	b.config.Header.Padding.Global = padding
	return b
}

// --- Direct Row Configuration Methods (for ConfigBuilder) ---

// WithRowAlignment sets the text alignment for all row cells.
func (b *ConfigBuilder) WithRowAlignment(align tw.Align) *ConfigBuilder {
	if align != tw.AlignLeft && align != tw.AlignRight && align != tw.AlignCenter && align != tw.AlignNone {
		return b // Ignore invalid alignment
	}
	b.config.Row.Formatting.Alignment = align
	return b
}

// WithRowAutoWrap sets the wrapping behavior for row cells.
func (b *ConfigBuilder) WithRowAutoWrap(autoWrap int) *ConfigBuilder {
	if autoWrap < tw.WrapNone || autoWrap > tw.WrapBreak {
		return b // Ignore invalid wrap mode
	}
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
	if mergeMode < tw.MergeNone || mergeMode > tw.MergeHierarchical {
		return b // Ignore invalid merge mode
	}
	b.config.Row.Formatting.MergeMode = mergeMode
	return b
}

// WithRowMaxWidth sets the maximum content width for row cells.
func (b *ConfigBuilder) WithRowMaxWidth(maxWidth int) *ConfigBuilder {
	if maxWidth < 0 {
		return b // Ignore negative width
	}
	b.config.Row.Formatting.MaxWidth = maxWidth
	return b
}

// WithRowGlobalPadding sets the global padding for all row cells.
func (b *ConfigBuilder) WithRowGlobalPadding(padding tw.Padding) *ConfigBuilder {
	b.config.Row.Padding.Global = padding
	return b
}

// --- Direct Footer Configuration Methods (for ConfigBuilder) ---

// WithFooterAlignment sets the text alignment for all footer cells.
func (b *ConfigBuilder) WithFooterAlignment(align tw.Align) *ConfigBuilder {
	if align != tw.AlignLeft && align != tw.AlignRight && align != tw.AlignCenter && align != tw.AlignNone {
		return b // Ignore invalid alignment
	}
	b.config.Footer.Formatting.Alignment = align
	return b
}

// WithFooterAutoWrap sets the wrapping behavior for footer cells.
func (b *ConfigBuilder) WithFooterAutoWrap(autoWrap int) *ConfigBuilder {
	if autoWrap < tw.WrapNone || autoWrap > tw.WrapBreak {
		return b // Ignore invalid wrap mode
	}
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
	if mergeMode < tw.MergeNone || mergeMode > tw.MergeHierarchical {
		return b // Ignore invalid merge mode
	}
	b.config.Footer.Formatting.MergeMode = mergeMode
	return b
}

// WithFooterMaxWidth sets the maximum content width for footer cells.
func (b *ConfigBuilder) WithFooterMaxWidth(maxWidth int) *ConfigBuilder {
	if maxWidth < 0 {
		return b // Ignore negative width
	}
	b.config.Footer.Formatting.MaxWidth = maxWidth
	return b
}

// WithFooterGlobalPadding sets the global padding for all footer cells.
func (b *ConfigBuilder) WithFooterGlobalPadding(padding tw.Padding) *ConfigBuilder {
	b.config.Footer.Padding.Global = padding
	return b
}

// --- Nested Builders for Advanced Configuration (for ConfigBuilder) ---

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

// --- Nested Builder Definitions (for ConfigBuilder) ---

// HeaderConfigBuilder provides advanced configuration for the header section.
type HeaderConfigBuilder struct {
	parent  *ConfigBuilder
	config  *tw.CellConfig
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
	config  *tw.CellFormatting
	section string
}

func (hf *HeaderFormattingBuilder) WithAlignment(align tw.Align) *HeaderFormattingBuilder {
	if align != tw.AlignLeft && align != tw.AlignRight && align != tw.AlignCenter && align != tw.AlignNone {
		return hf // Ignore invalid alignment
	}
	hf.config.Alignment = align
	return hf
}

func (hf *HeaderFormattingBuilder) WithAutoWrap(autoWrap int) *HeaderFormattingBuilder {
	if autoWrap < tw.WrapNone || autoWrap > tw.WrapBreak {
		return hf // Ignore invalid wrap mode
	}
	hf.config.AutoWrap = autoWrap
	return hf
}

func (hf *HeaderFormattingBuilder) WithAutoFormat(autoFormat bool) *HeaderFormattingBuilder {
	hf.config.AutoFormat = autoFormat
	return hf
}

func (hf *HeaderFormattingBuilder) WithMergeMode(mergeMode int) *HeaderFormattingBuilder {
	if mergeMode < tw.MergeNone || mergeMode > tw.MergeHierarchical {
		return hf // Ignore invalid merge mode
	}
	hf.config.MergeMode = mergeMode
	return hf
}

func (hf *HeaderFormattingBuilder) WithMaxWidth(maxWidth int) *HeaderFormattingBuilder {
	if maxWidth < 0 {
		return hf // Ignore negative width
	}
	hf.config.MaxWidth = maxWidth
	return hf
}

func (hf *HeaderFormattingBuilder) Build() *HeaderConfigBuilder {
	return hf.parent
}

// HeaderPaddingBuilder configures padding options for the header.
type HeaderPaddingBuilder struct {
	parent  *HeaderConfigBuilder
	config  *tw.CellPadding
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
type RowConfigBuilder struct {
	parent  *ConfigBuilder
	config  *tw.CellConfig
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
	config  *tw.CellFormatting
	section string
}

func (rf *RowFormattingBuilder) WithAlignment(align tw.Align) *RowFormattingBuilder {
	if align != tw.AlignLeft && align != tw.AlignRight && align != tw.AlignCenter && align != tw.AlignNone {
		return rf // Ignore invalid alignment
	}
	rf.config.Alignment = align
	return rf
}

func (rf *RowFormattingBuilder) WithAutoWrap(autoWrap int) *RowFormattingBuilder {
	if autoWrap < tw.WrapNone || autoWrap > tw.WrapBreak {
		return rf // Ignore invalid wrap mode
	}
	rf.config.AutoWrap = autoWrap
	return rf
}

func (rf *RowFormattingBuilder) WithAutoFormat(autoFormat bool) *RowFormattingBuilder {
	rf.config.AutoFormat = autoFormat
	return rf
}

func (rf *RowFormattingBuilder) WithMergeMode(mergeMode int) *RowFormattingBuilder {
	if mergeMode < tw.MergeNone || mergeMode > tw.MergeHierarchical {
		return rf // Ignore invalid merge mode
	}
	rf.config.MergeMode = mergeMode
	return rf
}

func (rf *RowFormattingBuilder) WithMaxWidth(maxWidth int) *RowFormattingBuilder {
	if maxWidth < 0 {
		return rf // Ignore negative width
	}
	rf.config.MaxWidth = maxWidth
	return rf
}

func (rf *RowFormattingBuilder) Build() *RowConfigBuilder {
	return rf.parent
}

// RowPaddingBuilder configures padding options for rows.
type RowPaddingBuilder struct {
	parent  *RowConfigBuilder
	config  *tw.CellPadding
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
type FooterConfigBuilder struct {
	parent  *ConfigBuilder
	config  *tw.CellConfig
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
	config  *tw.CellFormatting
	section string
}

func (ff *FooterFormattingBuilder) WithAlignment(align tw.Align) *FooterFormattingBuilder {
	if align != tw.AlignLeft && align != tw.AlignRight && align != tw.AlignCenter && align != tw.AlignNone {
		return ff // Ignore invalid alignment
	}
	ff.config.Alignment = align
	return ff
}

func (ff *FooterFormattingBuilder) WithAutoWrap(autoWrap int) *FooterFormattingBuilder {
	if autoWrap < tw.WrapNone || autoWrap > tw.WrapBreak {
		return ff // Ignore invalid wrap mode
	}
	ff.config.AutoWrap = autoWrap
	return ff
}

func (ff *FooterFormattingBuilder) WithAutoFormat(autoFormat bool) *FooterFormattingBuilder {
	ff.config.AutoFormat = autoFormat
	return ff
}

func (ff *FooterFormattingBuilder) WithNewarkMode(mergeMode int) *FooterFormattingBuilder {
	if mergeMode < tw.MergeNone || mergeMode > tw.MergeHierarchical {
		return ff // Ignore invalid merge mode
	}
	ff.config.MergeMode = mergeMode
	return ff
}

func (ff *FooterFormattingBuilder) WithMaxWidth(maxWidth int) *FooterFormattingBuilder {
	if maxWidth < 0 {
		return ff // Ignore negative width
	}
	ff.config.MaxWidth = maxWidth
	return ff
}

func (ff *FooterFormattingBuilder) Build() *FooterConfigBuilder {
	return ff.parent
}

// FooterPaddingBuilder configures padding options for the footer.
type FooterPaddingBuilder struct {
	parent  *FooterConfigBuilder
	config  *tw.CellPadding
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
	if width < 0 {
		return c // Ignore negative width
	}
	if c.parent.config.Header.ColMaxWidths.PerColumn == nil {
		c.parent.config.Header.ColMaxWidths.PerColumn = make(map[int]int)
		c.parent.config.Row.ColMaxWidths.PerColumn = make(map[int]int)
		c.parent.config.Footer.ColMaxWidths.PerColumn = make(map[int]int)
	}
	c.parent.config.Header.ColMaxWidths.PerColumn[c.col] = width
	c.parent.config.Row.ColMaxWidths.PerColumn[c.col] = width
	c.parent.config.Footer.ColMaxWidths.PerColumn[c.col] = width
	return c
}

// WithAlignment sets the alignment for a specific column in the header section only.
func (c *ColumnConfigBuilder) WithAlignment(align tw.Align) *ColumnConfigBuilder {
	if align != tw.AlignLeft && align != tw.AlignRight && align != tw.AlignCenter && align != tw.AlignNone {
		return c // Ignore invalid alignment
	}
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

// mergeStreamConfig merges a source StreamConfig into a destination StreamConfig.
// It prioritizes non-zero/non-default source values.
func mergeStreamConfig(dst, src tw.StreamConfig) tw.StreamConfig {
	if src.Enable {
		dst.Enable = true
	}
	if src.Widths.Global != 0 {
		dst.Widths.Global = src.Widths.Global
	}
	if len(src.Widths.PerColumn) > 0 {
		if dst.Widths.PerColumn == nil {
			dst.Widths.PerColumn = make(map[int]int)
		}
		for k, v := range src.Widths.PerColumn {
			if v != 0 {
				dst.Widths.PerColumn[k] = v
			}
		}
	}
	//if src.BufferSize != 0 {
	//	dst.BufferSize = src.BufferSize
	//}
	//if src.ChunkSize != 0 {
	//	dst.ChunkSize = src.ChunkSize
	//}
	//if src.RefreshRate != 0 {
	//	dst.RefreshRate = src.RefreshRate
	//}
	//if src.ColorMode != "" {
	//	dst.ColorMode = src.ColorMode
	//}
	//if src.Throttle {
	//	dst.Throttle = true
	//}
	//if src.MaxLines != 0 {
	//	dst.MaxLines = src.MaxLines
	//}
	//if src.Silent {
	//	dst.Silent = true
	//}
	//if src.Interactive {
	//	dst.Interactive = true
	//}
	return dst
}

// mergeConfig merges a source Config into a destination Config.
// It prioritizes non-zero/non-default source values for fields and performs deep merging for complex types.
func mergeConfig(dst, src Config) Config {
	if src.MaxWidth != 0 {
		dst.MaxWidth = src.MaxWidth
	}
	dst.Debug = src.Debug || dst.Debug
	dst.AutoHide = src.AutoHide
	dst.Header = mergeCellConfig(dst.Header, src.Header)
	dst.Row = mergeCellConfig(dst.Row, src.Row)
	dst.Footer = mergeCellConfig(dst.Footer, src.Footer)
	dst.Stream = mergeStreamConfig(dst.Stream, src.Stream)
	return dst
}

// mergeCellConfig merges a source CellConfig into a destination CellConfig.
// It prioritizes non-zero/non-default source values and optimizes slice merging.
func mergeCellConfig(dst, src tw.CellConfig) tw.CellConfig {
	// Formatting
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

	// Padding
	if src.Padding.Global != (tw.Padding{}) {
		dst.Padding.Global = src.Padding.Global
	}
	if len(src.Padding.PerColumn) > 0 {
		if dst.Padding.PerColumn == nil {
			dst.Padding.PerColumn = make([]tw.Padding, len(src.Padding.PerColumn))
		} else if len(src.Padding.PerColumn) > len(dst.Padding.PerColumn) {
			dst.Padding.PerColumn = append(dst.Padding.PerColumn, make([]tw.Padding, len(src.Padding.PerColumn)-len(dst.Padding.PerColumn))...)
		}
		for i, pad := range src.Padding.PerColumn {
			if pad != (tw.Padding{}) {
				dst.Padding.PerColumn[i] = pad
			}
		}
	}

	// Callbacks
	if src.Callbacks.Global != nil {
		dst.Callbacks.Global = src.Callbacks.Global
	}
	if len(src.Callbacks.PerColumn) > 0 {
		if dst.Callbacks.PerColumn == nil {
			dst.Callbacks.PerColumn = make([]func(), len(src.Callbacks.PerColumn))
		} else if len(src.Callbacks.PerColumn) > len(dst.Callbacks.PerColumn) {
			dst.Callbacks.PerColumn = append(dst.Callbacks.PerColumn, make([]func(), len(src.Callbacks.PerColumn)-len(dst.Callbacks.PerColumn))...)
		}
		for i, cb := range src.Callbacks.PerColumn {
			if cb != nil {
				dst.Callbacks.PerColumn[i] = cb
			}
		}
	}

	// Filter
	if src.Filter.Global != nil {
		dst.Filter.Global = src.Filter.Global
	}
	if len(src.Filter.PerColumn) > 0 {
		if dst.Filter.PerColumn == nil {
			dst.Filter.PerColumn = make([]func(string) string, len(src.Filter.PerColumn))
		} else if len(src.Filter.PerColumn) > len(dst.Filter.PerColumn) {
			dst.Filter.PerColumn = append(dst.Filter.PerColumn, make([]func(string) string, len(src.Filter.PerColumn)-len(dst.Filter.PerColumn))...)
		}
		for i, filter := range src.Filter.PerColumn {
			if filter != nil {
				dst.Filter.PerColumn[i] = filter
			}
		}
	}

	// ColumnAligns
	if len(src.ColumnAligns) > 0 {
		if dst.ColumnAligns == nil {
			dst.ColumnAligns = make([]tw.Align, len(src.ColumnAligns))
		} else if len(src.ColumnAligns) > len(dst.ColumnAligns) {
			dst.ColumnAligns = append(dst.ColumnAligns, make([]tw.Align, len(src.ColumnAligns)-len(dst.ColumnAligns))...)
		}
		for i, align := range src.ColumnAligns {
			if align != tw.Empty && align != tw.Skip {
				dst.ColumnAligns[i] = align
			}
		}
	}

	// ColMaxWidths
	if len(src.ColMaxWidths.PerColumn) > 0 {
		if dst.ColMaxWidths.PerColumn == nil {
			dst.ColMaxWidths.PerColumn = make(map[int]int)
		}
		for k, v := range src.ColMaxWidths.PerColumn {
			if v != 0 {
				dst.ColMaxWidths.PerColumn[k] = v
			}
		}
	}

	return dst
}

// --- Option Functions ---

// Option defines a function to configure a Table instance.
type Option func(target *Table)

// WithHeader sets the table headers.
func WithHeader(headers []string) Option {
	return func(target *Table) {
		target.Header(headers)
	}
}

// WithFooter sets the table footers.
func WithFooter(footers []string) Option {
	return func(target *Table) {
		target.Footer(footers)
	}
}

// WithRenderer sets a custom renderer for the table.
func WithRenderer(f tw.Renderer) Option {
	return func(target *Table) {
		target.renderer = f
		if target.logger != nil {
			target.logger.Debug("Option: WithRenderer applied to Table: %T", f)
			f.Logger(target.logger)
		}
	}
}

// WithConfig applies a custom configuration to the table.
func WithConfig(cfg Config) Option {
	return func(target *Table) {
		target.config = mergeConfig(defaultConfig(), cfg)
		if target.logger != nil {
			target.logger.Debug("Option: WithConfig applied to Table.")
		}
	}
}

// WithStringer sets a custom stringer function for row conversion.
func WithStringer[T any](s func(T) []string) Option {
	return func(target *Table) {
		target.stringer = s
	}
}

// WithDebug enables or disables debug logging.
func WithDebug(debug bool) Option {
	return func(target *Table) {
		target.config.Debug = debug
		if target.logger != nil {
			if debug {
				target.logger.Enable()
				target.logger.Level(lx.LevelDebug)
			} else {
				target.logger.Level(lx.LevelInfo)
			}
			target.logger.Debug("Option: WithDebug applied to Table: %v", debug)
			if target.renderer != nil {
				target.renderer.Logger(target.logger)
			}
		}
	}
}

// WithLogger adds a custom logger.
func WithLogger(logger *ll.Logger) Option {
	return func(target *Table) {
		target.logger = logger
		if target.logger != nil {
			target.logger.Debug("Option: WithLogger applied to Table.")
			if target.renderer != nil {
				target.renderer.Logger(target.logger)
			}
		}
	}
}

// WithAutoHide enables or disables automatic hiding of columns with empty data rows.
func WithAutoHide(hide bool) Option {
	return func(target *Table) {
		target.config.AutoHide = hide
		if target.logger != nil {
			target.logger.Debug("Option: WithAutoHide applied to Table: %v", hide)
		}
	}
}

// WithHeaderAlignment sets the header alignment.
func WithHeaderAlignment(align tw.Align) Option {
	return func(target *Table) {
		if align != tw.AlignLeft && align != tw.AlignRight && align != tw.AlignCenter && align != tw.AlignNone {
			return
		}
		target.config.Header.Formatting.Alignment = align
		if target.logger != nil {
			target.logger.Debug("Option: WithHeaderAlignment applied to Table: %v", align)
		}
	}
}

// WithRowMaxWidth sets the row max width.
func WithRowMaxWidth(maxWidth int) Option {
	return func(target *Table) {
		if maxWidth < 0 {
			return
		}
		target.config.Row.Formatting.MaxWidth = maxWidth
		if target.logger != nil {
			target.logger.Debug("Option: WithRowMaxWidth applied to Table: %v", maxWidth)
		}
	}
}

// WithFooterMergeMode sets the footer merge mode.
func WithFooterMergeMode(mergeMode int) Option {
	return func(target *Table) {
		if mergeMode < tw.MergeNone || mergeMode > tw.MergeHierarchical {
			return
		}
		target.config.Footer.Formatting.MergeMode = mergeMode
		if target.logger != nil {
			target.logger.Debug("Option: WithFooterMergeMode applied to Table: %v", mergeMode)
		}
	}
}

// WithHeaderConfig applies a full header configuration.
func WithHeaderConfig(config tw.CellConfig) Option {
	return func(target *Table) {
		target.config.Header = config
		if target.logger != nil {
			target.logger.Debug("Option: WithHeaderConfig applied to Table.")
		}
	}
}

// WithStreaming applies a streaming configuration.
func WithStreaming(c tw.StreamConfig) Option {
	return func(target *Table) {
		target.config.Stream = mergeStreamConfig(target.config.Stream, c)
		if target.logger != nil {
			target.logger.Debug("Option: WithStreaming applied to Table.")
		}
	}
}

// WithRowConfig applies a full row configuration.
func WithRowConfig(config tw.CellConfig) Option {
	return func(target *Table) {
		target.config.Row = config
		if target.logger != nil {
			target.logger.Debug("Option: WithRowConfig applied to Table.")
		}
	}
}

// WithFooterConfig applies a full footer configuration.
func WithFooterConfig(config tw.CellConfig) Option {
	return func(target *Table) {
		target.config.Footer = config
		if target.logger != nil {
			target.logger.Debug("Option: WithFooterConfig applied to Table.")
		}
	}
}

// WithColumnWidths sets explicit column widths for the table.
func WithColumnWidths(widths map[int]int) Option {
	return func(target *Table) {
		for k, v := range widths {
			if v < 0 {
				delete(widths, k) // Ignore negative widths
			}
		}
		target.config.Stream.Widths.PerColumn = widths
		if target.logger != nil {
			target.logger.Debug("Option: WithColumnWidths applied to Table: %v", widths)
		}
	}
}

// WithColumnMax sets explicit column max widths for the table.
func WithColumnMax(width int) Option {
	return func(target *Table) {
		if width < 0 {
			return
		}
		target.config.Stream.Widths.Global = width
		if target.logger != nil {
			target.logger.Debug("Option: WithColumnMax applied to Table: %v", width)
		}
	}
}

// WithSymbols sets the symbols used for table drawing.
// This updates the renderer's configuration if supported.
func WithSymbols(symbols tw.Symbols) Option {
	return func(target *Table) {
		if target.renderer != nil {
			cfg := target.renderer.Config()
			cfg.Symbols = symbols
			// Note: Assumes renderer supports updating config; may need custom logic
			if target.logger != nil {
				target.logger.Debug("Option: WithSymbols applied to Table.")
			}
		}
	}
}

// WithBorders sets the border configuration for the table.
// This updates the renderer's configuration if supported.
func WithBorders(borders tw.Border) Option {
	return func(target *Table) {
		if target.renderer != nil {
			cfg := target.renderer.Config()
			cfg.Borders = borders
			if target.logger != nil {
				target.logger.Debug("Option: WithBorders applied to Table: %+v", borders)
			}
		}
	}
}

// WithRendererSettings sets additional renderer settings (e.g., separators, lines).
// This updates the renderer's configuration if supported.
func WithRendererSettings(settings tw.Settings) Option {
	return func(target *Table) {
		if target.renderer != nil {
			cfg := target.renderer.Config()
			cfg.Settings = settings
			if target.logger != nil {
				target.logger.Debug("Option: WithRendererSettings applied to Table: %+v", settings)
			}
		}
	}
}

// NewWriter creates a new table with default settings for backward compatibility.
func NewWriter(w io.Writer) *Table {
	t := NewTable(w)
	if t.logger != nil {
		t.logger.Debug("NewWriter created buffered Table")
	}
	return t
}

// defaultConfig returns a default Config with sensible settings.
func defaultConfig() Config {
	defaultPadding := tw.Padding{Left: tw.Space, Right: tw.Space, Top: tw.Empty, Bottom: tw.Empty}
	return Config{
		MaxWidth: 0,
		Header: tw.CellConfig{
			Formatting: tw.CellFormatting{
				MaxWidth:   0,
				AutoWrap:   tw.WrapTruncate,
				Alignment:  tw.AlignCenter,
				AutoFormat: true,
				MergeMode:  tw.MergeNone,
			},
			Padding: tw.CellPadding{
				Global: defaultPadding,
			},
		},
		Row: tw.CellConfig{
			Formatting: tw.CellFormatting{
				MaxWidth:   0,
				AutoWrap:   tw.WrapNormal,
				Alignment:  tw.AlignLeft,
				AutoFormat: false,
				MergeMode:  tw.MergeNone,
			},
			Padding: tw.CellPadding{
				Global: defaultPadding,
			},
		},
		Footer: tw.CellConfig{
			Formatting: tw.CellFormatting{
				MaxWidth:   0,
				AutoWrap:   tw.WrapNormal,
				Alignment:  tw.AlignRight,
				AutoFormat: false,
				MergeMode:  tw.MergeNone,
			},
			Padding: tw.CellPadding{
				Global: defaultPadding,
			},
		},
		Stream:   tw.StreamConfig{},
		Debug:    true,
		AutoHide: false,
	}
}

// padLine pads a line to the specified column count.
// Returns the padded line with empty strings as needed.
func padLine(line []string, numCols int) []string {
	if len(line) >= numCols {
		return line
	}
	padded := make([]string, numCols)
	copy(padded, line)
	for i := len(line); i < numCols; i++ {
		padded[i] = tw.Empty
	}
	return padded
}
