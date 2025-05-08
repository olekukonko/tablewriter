package streamer

import (
	"fmt"
	"io"
	"strings"

	"github.com/olekukonko/errors"
	"github.com/olekukonko/tablewriter/tw"
	"github.com/olekukonko/tablewriter/twfn"
)

// OceanConfig defines the configuration for the Ocean table renderer.
type OceanConfig struct {
	ColumnWidths   []int      // Widths for each column
	Symbols        tw.Symbols // Symbols for table borders and separators
	Borders        tw.Border  // Border visibility settings
	ColumnAligns   []tw.Align // Per-column alignment overrides
	HeaderAlign    tw.Align   // Default alignment for header cells
	RowAlign       tw.Align   // Default alignment for row cells
	FooterAlign    tw.Align   // Default alignment for footer cells
	Padding        tw.Padding // Padding characters for cells
	TrimWhitespace tw.State   // Whether to trim whitespace from cell content
	ShowHeaderLine bool       // Whether to render a separator line after the header
	ShowFooterLine bool       // Whether to render a separator line before the footer
}

// Ocean is a streaming table renderer that writes ASCII tables with fixed column widths.
type Ocean struct {
	config OceanConfig // Renderer configuration
	writer io.Writer   // Output writer
	trace  []string    // Debug trace messages
	debug  bool        // Enables debug logging

	tableStarted    bool // Tracks if table rendering has started
	headerRendered  bool // Tracks if header has been rendered
	lastRowRendered bool // Tracks if the last rendered line was a row
}

// NewOcean initializes an Ocean renderer with the given writer, debug setting, and configuration.
// It validates column widths and applies defaults for unset fields.
func NewOcean(w io.Writer, debug bool, config OceanConfig) (*Ocean, error) {
	if w == nil {
		return nil, errors.New("Ocean renderer requires a non-nil writer")
	}
	if len(config.ColumnWidths) == 0 {
		return nil, errors.New("OceanConfig requires ColumnWidths to be set")
	}
	numCols := len(config.ColumnWidths)

	// Apply defaults for padding validation
	effectivePadding := config.Padding
	if effectivePadding.Left == "" && effectivePadding.Right == "" && effectivePadding.Top == "" && effectivePadding.Bottom == "" {
		effectivePadding = tw.Padding{Left: " ", Right: " "}
	}
	padLStr := effectivePadding.Left
	if padLStr == "" {
		padLStr = " "
	}
	padRStr := effectivePadding.Right
	if padRStr == "" {
		padRStr = " "
	}

	// Validate column widths
	for i, cw := range config.ColumnWidths {
		minRequiredCellWidth := twfn.DisplayWidth(padLStr) + twfn.DisplayWidth(padRStr)
		if cw < minRequiredCellWidth {
			return nil, fmt.Errorf("OceanConfig: ColumnWidths[%d]=%d is too small; needs at least %d for padding ('%s' + content + '%s')", i, cw, minRequiredCellWidth, padLStr, padRStr)
		}
	}

	// Validate column alignments
	if config.ColumnAligns != nil && len(config.ColumnAligns) != numCols {
		return nil, fmt.Errorf("OceanConfig: ColumnAligns length (%d) must match ColumnWidths length (%d)", len(config.ColumnAligns), numCols)
	}

	// Apply defaults for unset fields
	if config.Symbols == nil {
		config.Symbols = tw.NewSymbols(tw.StyleLight)
	}
	if config.Borders == (tw.Border{}) {
		config.Borders = tw.Border{Left: tw.On, Right: tw.On, Top: tw.On, Bottom: tw.On}
	}
	if config.HeaderAlign == "" {
		config.HeaderAlign = tw.AlignCenter
	}
	if config.RowAlign == "" {
		config.RowAlign = tw.AlignLeft
	}
	if config.FooterAlign == "" {
		config.FooterAlign = tw.AlignRight
	}
	if config.Padding == (tw.Padding{}) {
		config.Padding = tw.Padding{Left: " ", Right: " "}
	}
	if config.TrimWhitespace == 0 {
		config.TrimWhitespace = tw.On
	}

	renderer := &Ocean{
		config:          config,
		writer:          w,
		debug:           debug,
		trace:           make([]string, 0, 50),
		tableStarted:    false,
		headerRendered:  false,
		lastRowRendered: false,
	}
	renderer.debugLog("Initialized Ocean renderer. NumCols: %d, Widths: %v, Padding L:'%s' R:'%s'",
		numCols, config.ColumnWidths, config.Padding.Left, config.Padding.Right)
	return renderer, nil
}

// Start begins the table stream, rendering the top border if enabled.
// The provided writer is ignored in favor of the instance's writer.
func (s *Ocean) Start(w io.Writer) error {
	if s.writer == nil {
		return errors.New("Ocean renderer writer not initialized")
	}
	if s.tableStarted {
		s.debugLog("Start() called, but table already started. Resetting internal state.")
		s.Reset()
	} else {
		s.trace = make([]string, 0, 50)
	}
	s.tableStarted = true
	s.debugLog("Ocean.Start() called.")

	if s.config.Borders.Top.Enabled() {
		s.debugLog("Start: Rendering top border")
		s.renderBorderLine(s.writer, tw.LocationFirst)
	}
	return nil
}

// Header renders the header row and an optional separator line.
func (s *Ocean) Header(w io.Writer, headerRow []string) error {
	if !s.tableStarted {
		return errors.New("Ocean.Header() called before Start()")
	}
	if s.writer == nil {
		return errors.New("Ocean renderer writer not initialized")
	}

	s.debugLog("Header: Rendering content row: %v", headerRow)
	s.renderContentLine(s.writer, headerRow, tw.Header)
	s.headerRendered = true
	s.lastRowRendered = false

	if s.config.ShowHeaderLine {
		s.debugLog("Header: Rendering separator line")
		s.renderBorderLine(s.writer, tw.LocationMiddle)
	}
	return nil
}

// Row renders a data row.
func (s *Ocean) Row(w io.Writer, row []string) error {
	if !s.tableStarted {
		return errors.New("Ocean.Row() called before Start()")
	}
	if s.writer == nil {
		return errors.New("Ocean renderer writer not initialized")
	}
	s.debugLog("Row: Rendering content row: %v", row)
	s.renderContentLine(s.writer, row, tw.Row)
	s.lastRowRendered = true
	return nil
}

// Footer renders the footer row with an optional preceding separator line.
func (s *Ocean) Footer(w io.Writer, footerRow []string) error {
	if !s.tableStarted {
		return errors.New("Ocean.Footer() called before Start()")
	}
	if s.writer == nil {
		return errors.New("Ocean renderer writer not initialized")
	}

	if s.config.ShowFooterLine && s.lastRowRendered {
		s.debugLog("Footer: Rendering separator line")
		s.renderBorderLine(s.writer, tw.LocationMiddle)
	}

	s.debugLog("Footer: Rendering content row: %v", footerRow)
	s.renderContentLine(s.writer, footerRow, tw.Footer)
	s.lastRowRendered = false
	return nil
}

// End finalizes the table stream, rendering the bottom border if enabled.
func (s *Ocean) End(w io.Writer) error {
	if !s.tableStarted {
		return errors.New("Ocean.End() called before Start()")
	}
	if s.writer == nil {
		return errors.New("Ocean renderer writer not initialized")
	}

	if !s.headerRendered && !s.lastRowRendered {
		s.debugLog("End: Context indicates an empty table (no header/rows/footer rendered). Top border should have been drawn by Start.")
	}

	if s.config.Borders.Bottom.Enabled() {
		s.debugLog("End: Rendering bottom border")
		s.renderBorderLine(s.writer, tw.LocationEnd)
	}
	s.tableStarted = false
	s.headerRendered = false
	s.lastRowRendered = false
	return nil
}

// Debug returns the accumulated debug trace messages.
func (s *Ocean) Debug() []string {
	return s.trace
}

// Reset clears the renderer's internal state, including debug traces.
func (s *Ocean) Reset() {
	s.debugLog("Ocean.Reset() called.")
	s.tableStarted = false
	s.headerRendered = false
	s.lastRowRendered = false
	s.trace = make([]string, 0, 50)
}

// Config returns a RendererConfig representation of the current configuration.
func (s *Ocean) Config() tw.RendererConfig {
	return tw.RendererConfig{
		Borders:  s.config.Borders,
		Symbols:  s.config.Symbols,
		Settings: tw.Settings{TrimWhitespace: s.config.TrimWhitespace},
		Debug:    s.debug,
	}
}

// debugLog appends a formatted message to the debug trace if debugging is enabled.
func (s *Ocean) debugLog(format string, a ...interface{}) {
	if s.debug {
		s.trace = append(s.trace, fmt.Sprintf("[OCEANSTRM] %s", fmt.Sprintf(format, a...)))
	}
}

// renderContentLine renders a header, row, or footer line with aligned content and borders.
func (s *Ocean) renderContentLine(w io.Writer, data []string, position tw.Position) {
	var sb strings.Builder
	numCols := len(s.config.ColumnWidths)

	if s.config.Borders.Left.Enabled() {
		sb.WriteString(s.config.Symbols.Column())
	}

	for i := 0; i < numCols; i++ {
		if i > 0 {
			sb.WriteString(s.config.Symbols.Column())
		}

		content := ""
		if i < len(data) {
			content = data[i]
		}
		align := s.getDefaultAlign(position)
		if s.config.ColumnAligns != nil && i < len(s.config.ColumnAligns) && s.config.ColumnAligns[i] != "" && s.config.ColumnAligns[i] != tw.AlignNone {
			align = s.config.ColumnAligns[i]
		}
		targetWidth := s.config.ColumnWidths[i]
		formattedCell := s.formatStreamCell(content, targetWidth, align)
		sb.WriteString(formattedCell)
	}

	if s.config.Borders.Right.Enabled() {
		sb.WriteString(s.config.Symbols.Column())
	}
	fmt.Fprintln(w, sb.String())
}

// renderBorderLine renders a horizontal border line (top, middle, or bottom).
func (s *Ocean) renderBorderLine(w io.Writer, location tw.Location) {
	var sb strings.Builder
	numCols := len(s.config.ColumnWidths)
	sym := s.config.Symbols
	leftSym, midSym, rightSym, rowSym := sym.MidLeft(), sym.Center(), sym.MidRight(), sym.Row()

	switch location {
	case tw.LocationFirst:
		leftSym, midSym, rightSym = sym.TopLeft(), sym.TopMid(), sym.TopRight()
	case tw.LocationEnd:
		leftSym, midSym, rightSym = sym.BottomLeft(), sym.BottomMid(), sym.BottomRight()
	}

	if s.config.Borders.Left.Enabled() {
		sb.WriteString(leftSym)
	}
	for i := 0; i < numCols; i++ {
		if i > 0 {
			sb.WriteString(midSym)
		}
		width := s.config.ColumnWidths[i]
		if width > 0 {
			sb.WriteString(strings.Repeat(rowSym, width))
		}
	}
	if s.config.Borders.Right.Enabled() {
		sb.WriteString(rightSym)
	}
	fmt.Fprintln(w, sb.String())
}

// formatStreamCell formats a cell's content with padding, alignment, and optional truncation.
func (s *Ocean) formatStreamCell(content string, targetCellWidth int, align tw.Align) string {
	padLeftChar := s.config.Padding.Left
	if padLeftChar == "" {
		padLeftChar = " "
	}
	padRightChar := s.config.Padding.Right
	if padRightChar == "" {
		padRightChar = " "
	}

	padLeftWidth := twfn.DisplayWidth(padLeftChar)
	padRightWidth := twfn.DisplayWidth(padRightChar)

	if s.config.TrimWhitespace.Enabled() {
		content = strings.TrimSpace(content)
	}
	contentVisualWidth := twfn.DisplayWidth(content)

	availableForContent := targetCellWidth - padLeftWidth - padRightWidth
	if availableForContent < 0 {
		s.debugLog("formatStreamCell Warning: targetCellWidth %d is less than padding width %d. Cell content: '%s'. Truncating padding.", targetCellWidth, padLeftWidth+padRightWidth, content)

		var sb strings.Builder
		currentLen := 0
		for _, r := range padLeftChar {
			rWidth := twfn.DisplayWidth(string(r))
			if currentLen+rWidth <= targetCellWidth {
				sb.WriteRune(r)
				currentLen += rWidth
			} else {
				break
			}
		}
		if currentLen < targetCellWidth {
			sb.WriteString(strings.Repeat(" ", targetCellWidth-currentLen))
		}

		res := sb.String()
		if twfn.DisplayWidth(res) > targetCellWidth {
			return twfn.TruncateString(res, targetCellWidth)
		}
		return res
	}

	suffix := "…"
	suffixWidth := twfn.DisplayWidth(suffix)

	if contentVisualWidth > availableForContent {
		if availableForContent >= suffixWidth {
			content = twfn.TruncateString(content, availableForContent-suffixWidth) + suffix
		} else {
			content = twfn.TruncateString(content, availableForContent)
		}
		contentVisualWidth = twfn.DisplayWidth(content)
	}

	internalPaddingNeeded := availableForContent - contentVisualWidth
	if internalPaddingNeeded < 0 {
		internalPaddingNeeded = 0
	}

	leftInternalPadCount, rightInternalPadCount := 0, 0
	internalPadChar := " "

	switch align {
	case tw.AlignRight:
		leftInternalPadCount = internalPaddingNeeded
	case tw.AlignCenter:
		leftInternalPadCount = internalPaddingNeeded / 2
		rightInternalPadCount = internalPaddingNeeded - leftInternalPadCount
	default:
		rightInternalPadCount = internalPaddingNeeded
	}

	var sb strings.Builder
	sb.WriteString(padLeftChar)
	sb.WriteString(strings.Repeat(internalPadChar, leftInternalPadCount))
	sb.WriteString(content)
	sb.WriteString(strings.Repeat(internalPadChar, rightInternalPadCount))
	sb.WriteString(padRightChar)

	result := sb.String()

	finalVisualWidth := twfn.DisplayWidth(result)
	if finalVisualWidth != targetCellWidth {
		s.debugLog("formatStreamCell Post-Build MISMATCH: Initial target_w=%d. Built: '%s' (visual_w=%d). Content was '%s'. AvailForContent: %d",
			targetCellWidth, result, finalVisualWidth, content, availableForContent)
		if finalVisualWidth < targetCellWidth {
			spacesToAdj := targetCellWidth - finalVisualWidth
			switch align {
			case tw.AlignRight:
				result = strings.Repeat(" ", spacesToAdj) + result
			case tw.AlignCenter:
				lAdj := spacesToAdj / 2
				rAdj := spacesToAdj - lAdj
				result = strings.Repeat(" ", lAdj) + result + strings.Repeat(" ", rAdj)
			default:
				result = result + strings.Repeat(" ", spacesToAdj)
			}
		} else {
			result = twfn.TruncateString(result, targetCellWidth)
		}
		s.debugLog("formatStreamCell Post-Build Corrected: target_w=%d -> result='%s' (new_visual_w=%d)",
			targetCellWidth, result, twfn.DisplayWidth(result))
	}
	return result
}

// getDefaultAlign returns the default alignment for a given table position.
func (s *Ocean) getDefaultAlign(position tw.Position) tw.Align {
	switch position {
	case tw.Header:
		return s.config.HeaderAlign
	case tw.Footer:
		return s.config.FooterAlign
	default:
		return s.config.RowAlign
	}
}
