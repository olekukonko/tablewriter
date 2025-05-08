package renderer

import (
	"fmt"
	"html" // For escaping text content
	"io"
	"strings"

	"github.com/olekukonko/tablewriter/tw"
)

// SVGConfig holds configuration specific to the pure SVG renderer.
type SVGConfig struct {
	FontFamily              string  // e.g., "Arial, sans-serif"
	FontSize                float64 // Base font size in SVG units (often pixels)
	LineHeightFactor        float64 // Factor for line height (e.g., 1.2 for 1.2 * FontSize line height)
	Padding                 float64 // Padding inside cells (SVG units)
	StrokeWidth             float64 // Line width for borders/separators
	StrokeColor             string  // Color for strokes (e.g., "black", "#FF0000")
	HeaderBG                string  // Background color for header cells
	RowBG                   string  // Background color for row cells
	RowAltBG                string  // Optional alternating row background color
	FooterBG                string  // Background color for footer cells
	HeaderColor             string  // Text color for header
	RowColor                string  // Text color for rows
	FooterColor             string  // Text color for footer
	ApproxCharWidthFactor   float64 // Approximate width of a char relative to FontSize (e.g., 0.6)
	MinColWidth             float64 // Minimum column width in SVG units
	RenderTWConfigOverrides bool    // If true, tablewriter alignments override SVG text-anchor
	Debug                   bool
}

// SVGRenderer implements the tw.Renderer interface for pure SVG output.
type SVGRenderer struct {
	config SVGConfig
	trace  []string

	allVisualLineData [][][]string      // [sectionType][visualLineIdx][cellContent]
	allVisualLineCtx  [][]tw.Formatting // [sectionType][visualLineIdx]FormattingContext

	maxCols             int
	calculatedColWidths []float64
	svgElements         strings.Builder
	currentY            float64
	dataRowCounter      int
	vMergeTrack         map[int]int // Tracks for each COLUMN, how many more visual rows its current v-merge spans
	numVisualRowsDrawn  int
}

const (
	sectionTypeHeader = 0
	sectionTypeRow    = 1
	sectionTypeFooter = 2
)

// NewSVG creates a new pure SVG renderer.
func NewSVG(configs ...SVGConfig) *SVGRenderer {
	cfg := SVGConfig{
		FontFamily:              "sans-serif",
		FontSize:                12.0,
		LineHeightFactor:        1.4,
		Padding:                 5.0,
		StrokeWidth:             1.0,
		StrokeColor:             "black",
		HeaderBG:                "#F0F0F0",
		RowBG:                   "white",
		RowAltBG:                "#F9F9F9",
		FooterBG:                "#F0F0F0",
		HeaderColor:             "black",
		RowColor:                "black",
		FooterColor:             "black",
		ApproxCharWidthFactor:   0.6,
		MinColWidth:             30.0,
		RenderTWConfigOverrides: true,
		Debug:                   false,
	}
	if len(configs) > 0 {
		userCfg := configs[0]
		if userCfg.FontFamily != "" {
			cfg.FontFamily = userCfg.FontFamily
		}
		if userCfg.FontSize > 0 {
			cfg.FontSize = userCfg.FontSize
		}
		if userCfg.LineHeightFactor > 0 {
			cfg.LineHeightFactor = userCfg.LineHeightFactor
		}
		if userCfg.Padding >= 0 {
			cfg.Padding = userCfg.Padding
		}
		if userCfg.StrokeWidth > 0 {
			cfg.StrokeWidth = userCfg.StrokeWidth
		}
		if userCfg.StrokeColor != "" {
			cfg.StrokeColor = userCfg.StrokeColor
		}
		if userCfg.HeaderBG != "" {
			cfg.HeaderBG = userCfg.HeaderBG
		}
		if userCfg.RowBG != "" {
			cfg.RowBG = userCfg.RowBG
		}
		cfg.RowAltBG = userCfg.RowAltBG
		if userCfg.FooterBG != "" {
			cfg.FooterBG = userCfg.FooterBG
		}
		if userCfg.HeaderColor != "" {
			cfg.HeaderColor = userCfg.HeaderColor
		}
		if userCfg.RowColor != "" {
			cfg.RowColor = userCfg.RowColor
		}
		if userCfg.FooterColor != "" {
			cfg.FooterColor = userCfg.FooterColor
		}
		if userCfg.ApproxCharWidthFactor > 0 {
			cfg.ApproxCharWidthFactor = userCfg.ApproxCharWidthFactor
		}
		if userCfg.MinColWidth >= 0 {
			cfg.MinColWidth = userCfg.MinColWidth
		}
		cfg.RenderTWConfigOverrides = userCfg.RenderTWConfigOverrides
		cfg.Debug = userCfg.Debug
	}

	r := &SVGRenderer{
		config:            cfg,
		trace:             make([]string, 0, 50),
		allVisualLineData: make([][][]string, 3), // Header, Row, Footer
		allVisualLineCtx:  make([][]tw.Formatting, 3),
		vMergeTrack:       make(map[int]int),
	}
	for i := 0; i < 3; i++ {
		r.allVisualLineData[i] = make([][]string, 0)
		r.allVisualLineCtx[i] = make([]tw.Formatting, 0)
	}
	return r
}

func (s *SVGRenderer) debug(format string, a ...interface{}) {
	if s.config.Debug {
		msg := fmt.Sprintf(format, a...)
		s.trace = append(s.trace, fmt.Sprintf("[SVG] %s", msg))
	}
}

func (s *SVGRenderer) Debug() []string {
	return s.trace
}

func (s *SVGRenderer) Config() tw.RendererConfig {
	return tw.RendererConfig{
		Borders:  tw.Border{Left: tw.On, Right: tw.On, Top: tw.On, Bottom: tw.On},
		Settings: tw.Settings{TrimWhitespace: tw.On},
		Debug:    s.config.Debug,
	}
}

func (s *SVGRenderer) Start(w io.Writer) error {
	s.debug("Start called")
	s.Reset()
	return nil
}

func (s *SVGRenderer) Reset() {
	s.debug("Resetting internal state")
	s.trace = make([]string, 0, 50)
	for i := 0; i < 3; i++ {
		s.allVisualLineData[i] = s.allVisualLineData[i][:0]
		s.allVisualLineCtx[i] = s.allVisualLineCtx[i][:0]
	}
	s.maxCols = 0
	s.calculatedColWidths = nil
	s.svgElements.Reset()
	s.currentY = 0
	s.dataRowCounter = 0
	s.vMergeTrack = make(map[int]int)
	s.numVisualRowsDrawn = 0
}

func (s *SVGRenderer) estimateTextWidth(text string) float64 {
	runeCount := float64(len([]rune(text)))
	return runeCount * s.config.FontSize * s.config.ApproxCharWidthFactor
}

func (s *SVGRenderer) calculateAllColumnWidths() {
	s.debug("Calculating all column widths...")
	tempMaxCols := 0

	// --- Pass 1: Determine Max Columns from Context ---
	for sectionIdx := 0; sectionIdx < 3; sectionIdx++ {
		// s.debug("Checking maxCols for section %d, num lines: %d", sectionIdx, len(s.allVisualLineCtx[sectionIdx])) // Redundant Debug
		for lineIdx, lineCtx := range s.allVisualLineCtx[sectionIdx] {
			if lineCtx.Row.Current == nil {
				// s.debug("Section %d Line %d: lineCtx.Row.Current is nil", sectionIdx, lineIdx) // Redundant Debug
				// Fallback check
				if lineIdx < len(s.allVisualLineData[sectionIdx]) {
					rawDataLen := len(s.allVisualLineData[sectionIdx][lineIdx])
					if rawDataLen > tempMaxCols {
						tempMaxCols = rawDataLen
						// s.debug("Updated maxCols to %d from raw data length (section %d line %d)", tempMaxCols, sectionIdx, lineIdx) // Redundant
					}
				}
				continue
			}
			// s.debug("Section %d Line %d: Processing %d cells in context", sectionIdx, lineIdx, len(lineCtx.Row.Current)) // Redundant Debug
			for colIdx, cellCtx := range lineCtx.Row.Current {
				currentMaxReach := colIdx + 1
				if cellCtx.Merge.Horizontal.Present && cellCtx.Merge.Horizontal.Start {
					span := cellCtx.Merge.Horizontal.Span
					if span <= 0 {
						span = 1
					}
					currentMaxReach = colIdx + span
				}
				if currentMaxReach > tempMaxCols {
					tempMaxCols = currentMaxReach
					// s.debug("Updated maxCols to %d from context (section %d line %d col %d span %d)", tempMaxCols, sectionIdx, lineIdx, colIdx, currentMaxReach-colIdx) // Redundant
				}
			}
		}
	}

	s.maxCols = tempMaxCols
	s.debug("Max columns determined: %d", s.maxCols)

	if s.maxCols == 0 {
		s.calculatedColWidths = []float64{}
		s.debug("MaxCols is 0, returning empty calculated widths.")
		return
	}

	// Initialize widths to minimum
	s.calculatedColWidths = make([]float64, s.maxCols)
	for i := range s.calculatedColWidths {
		s.calculatedColWidths[i] = s.config.MinColWidth
	}

	// --- Pass 2: Calculate Widths Based on Content and Merges ---
	processSectionForWidth := func(sectionIdx int) {
		for lineIdx, visualLineData := range s.allVisualLineData[sectionIdx] {
			if lineIdx >= len(s.allVisualLineCtx[sectionIdx]) {
				s.debug("Warning: Context missing for section %d line %d during width calculation. Skipping line.", sectionIdx, lineIdx)
				continue
			}
			lineCtx := s.allVisualLineCtx[sectionIdx][lineIdx]
			currentTableCol := 0
			currentVisualCol := 0

			for currentVisualCol < len(visualLineData) && currentTableCol < s.maxCols {
				cellContent := visualLineData[currentVisualCol]

				cellCtx := tw.CellContext{}
				if lineCtx.Row.Current != nil {
					if c, ok := lineCtx.Row.Current[currentTableCol]; ok {
						cellCtx = c
					}
				}

				hSpan := 1
				if cellCtx.Merge.Horizontal.Present {
					if cellCtx.Merge.Horizontal.Start {
						hSpan = cellCtx.Merge.Horizontal.Span
						if hSpan <= 0 {
							hSpan = 1
						}
					} else { // Consumed by previous merge start
						currentTableCol++
						continue
					}
				}

				textPixelWidth := s.estimateTextWidth(cellContent)
				contentAndPaddingWidth := textPixelWidth + (2 * s.config.Padding)

				if hSpan == 1 {
					if currentTableCol < len(s.calculatedColWidths) { // Bounds check
						if contentAndPaddingWidth > s.calculatedColWidths[currentTableCol] {
							s.calculatedColWidths[currentTableCol] = contentAndPaddingWidth
						}
					}
				} else { // Horizontal Merge
					currentSpannedContentWidth := 0.0
					colsInSpan := 0
					for i := 0; i < hSpan && (currentTableCol+i) < s.maxCols && (currentTableCol+i) < len(s.calculatedColWidths); i++ {
						currentSpannedContentWidth += s.calculatedColWidths[currentTableCol+i]
						colsInSpan++
					}

					if contentAndPaddingWidth > currentSpannedContentWidth {
						if currentTableCol < len(s.calculatedColWidths) {
							neededExtra := contentAndPaddingWidth - currentSpannedContentWidth
							s.calculatedColWidths[currentTableCol] += neededExtra
							s.debug("Col %d (HMerge Span %d): Added %.2f extra width for content '%s'. New width for col %d: %.2f",
								currentTableCol, hSpan, neededExtra, cellContent, currentTableCol, s.calculatedColWidths[currentTableCol])
						}
					}
				}
				currentTableCol += hSpan
				currentVisualCol++
			}
		}
	}

	processSectionForWidth(sectionTypeHeader)
	processSectionForWidth(sectionTypeRow)
	processSectionForWidth(sectionTypeFooter)

	for i := range s.calculatedColWidths {
		if s.calculatedColWidths[i] < s.config.MinColWidth {
			s.calculatedColWidths[i] = s.config.MinColWidth
		}
	}
	s.debug("Final calculated pixel column widths: %v", s.calculatedColWidths)
}

func (s *SVGRenderer) storeVisualLine(sectionIdx int, lineData []string, ctx tw.Formatting) {
	copiedLineData := make([]string, len(lineData))
	copy(copiedLineData, lineData)
	s.allVisualLineData[sectionIdx] = append(s.allVisualLineData[sectionIdx], copiedLineData)
	s.allVisualLineCtx[sectionIdx] = append(s.allVisualLineCtx[sectionIdx], ctx)
	// **** ADDED DEBUG ****
	hasCurrent := "false"
	if ctx.Row.Current != nil {
		hasCurrent = fmt.Sprintf("true (%d cells)", len(ctx.Row.Current))
	}
	s.debug("Stored line in section %d. Context has Current? %s", sectionIdx, hasCurrent)
	// **** END ADDED DEBUG ****
}

func (s *SVGRenderer) Header(w io.Writer, headers [][]string, ctx tw.Formatting) {
	s.debug("Header called, buffering %d visual lines.", len(headers))
	for i, line := range headers {
		currentCtx := ctx
		currentCtx.IsSubRow = (i > 0)
		s.storeVisualLine(sectionTypeHeader, line, currentCtx)
	}
}

func (s *SVGRenderer) Row(w io.Writer, rowLine []string, ctx tw.Formatting) {
	s.debug("Row called, buffering visual line (IsSubRow: %v)", ctx.IsSubRow)
	s.storeVisualLine(sectionTypeRow, rowLine, ctx)
}

func (s *SVGRenderer) Footer(w io.Writer, footers [][]string, ctx tw.Formatting) {
	s.debug("Footer called, buffering %d visual lines.", len(footers))
	for i, line := range footers {
		currentCtx := ctx
		currentCtx.IsSubRow = (i > 0)
		s.storeVisualLine(sectionTypeFooter, line, currentCtx)
	}
}

func (s *SVGRenderer) Line(w io.Writer, ctx tw.Formatting) {
	s.debug("Line called (ignored by SVG renderer)")
}

func (s *SVGRenderer) renderBufferedData() {
	s.debug("Rendering buffered data to SVG elements string builder")
	s.currentY = s.config.StrokeWidth
	s.dataRowCounter = 0
	s.vMergeTrack = make(map[int]int)
	s.numVisualRowsDrawn = 0 // Reset before counting during render

	renderSection := func(sectionIdx int, position tw.Position) {
		s.debug("Rendering section %d (%s), %d visual lines", sectionIdx, position, len(s.allVisualLineData[sectionIdx]))
		for visualLineIdx, visualLineData := range s.allVisualLineData[sectionIdx] {
			if visualLineIdx >= len(s.allVisualLineCtx[sectionIdx]) {
				s.debug("Error: Context missing for section %d line %d", sectionIdx, visualLineIdx)
				continue
			}
			s.renderVisualLine(visualLineData, s.allVisualLineCtx[sectionIdx][visualLineIdx], position)
		}
	}

	renderSection(sectionTypeHeader, tw.Header)
	renderSection(sectionTypeRow, tw.Row)
	renderSection(sectionTypeFooter, tw.Footer)
	s.debug("Finished renderBufferedData, numVisualRowsDrawn: %d", s.numVisualRowsDrawn)
}

func (s *SVGRenderer) renderVisualLine(visualLineData []string, ctx tw.Formatting, position tw.Position) {
	if s.maxCols == 0 || len(s.calculatedColWidths) == 0 {
		s.debug("Skipping visual line rendering - maxCols (%d) or calculatedColWidths (%d) is zero/nil.", s.maxCols, len(s.calculatedColWidths))
		return
	}
	s.numVisualRowsDrawn++
	s.debug("renderVisualLine: Drawing visual row %d", s.numVisualRowsDrawn)

	singleVisualRowHeight := s.config.FontSize*s.config.LineHeightFactor + (2 * s.config.Padding)

	bgColor := ""
	textColor := ""
	defaultTextAnchor := "start"

	switch position {
	case tw.Header:
		bgColor = s.config.HeaderBG
		textColor = s.config.HeaderColor
		defaultTextAnchor = "middle"
	case tw.Footer:
		bgColor = s.config.FooterBG
		textColor = s.config.FooterColor
		defaultTextAnchor = "end"
	default: // tw.Row
		textColor = s.config.RowColor
		if !ctx.IsSubRow {
			if s.config.RowAltBG != "" && s.dataRowCounter%2 != 0 {
				bgColor = s.config.RowAltBG
			} else {
				bgColor = s.config.RowBG
			}
			s.dataRowCounter++
		} else {
			parentDataRowStripeIndex := s.dataRowCounter - 1
			if parentDataRowStripeIndex < 0 {
				parentDataRowStripeIndex = 0
			}
			if s.config.RowAltBG != "" && parentDataRowStripeIndex%2 != 0 {
				bgColor = s.config.RowAltBG
			} else {
				bgColor = s.config.RowBG
			}
		}
	}

	currentX := s.config.StrokeWidth
	currentVisualCellIdx := 0

	for tableColIdx := 0; tableColIdx < s.maxCols; {
		if tableColIdx >= len(s.calculatedColWidths) {
			s.debug("renderVisualLine: Table Col Idx %d out of bounds for calculatedColWidths (len %d)", tableColIdx, len(s.calculatedColWidths))
			tableColIdx++
			continue
		}

		if remainingVSpan, isMerging := s.vMergeTrack[tableColIdx]; isMerging && remainingVSpan > 1 {
			// s.debug("Table Col %d: Under VMerge, span remaining %d. Skipping cell draw.", tableColIdx, remainingVSpan-1) // Verbose
			s.vMergeTrack[tableColIdx]--
			if s.vMergeTrack[tableColIdx] <= 1 {
				delete(s.vMergeTrack, tableColIdx)
			}
			currentX += s.calculatedColWidths[tableColIdx] + s.config.StrokeWidth
			tableColIdx++
			continue
		}

		cellContentFromVisualLine := ""
		if currentVisualCellIdx < len(visualLineData) {
			cellContentFromVisualLine = visualLineData[currentVisualCellIdx]
		}

		cellCtx := tw.CellContext{}
		if ctx.Row.Current != nil {
			if c, ok := ctx.Row.Current[tableColIdx]; ok {
				cellCtx = c
			} else {
				// s.debug("renderVisualLine: No context found for tableColIdx %d", tableColIdx) // Can be noisy
			}
		} else {
			// s.debug("renderVisualLine: ctx.Row.Current is nil") // Can be noisy
		}

		textToRender := cellContentFromVisualLine
		if cellCtx.Data != "" {
			if !((cellCtx.Merge.Vertical.Present && !cellCtx.Merge.Vertical.Start) || (cellCtx.Merge.Hierarchical.Present && !cellCtx.Merge.Hierarchical.Start)) {
				textToRender = cellCtx.Data
			} else {
				textToRender = ""
			}
		} else if (cellCtx.Merge.Vertical.Present && !cellCtx.Merge.Vertical.Start) || (cellCtx.Merge.Hierarchical.Present && !cellCtx.Merge.Hierarchical.Start) {
			textToRender = ""
		}

		hSpan := 1
		if cellCtx.Merge.Horizontal.Present {
			if cellCtx.Merge.Horizontal.Start {
				hSpan = cellCtx.Merge.Horizontal.Span
				if hSpan <= 0 {
					hSpan = 1
				}
			} else { // !Start && Present -> consumed by previous H-merge start
				currentX += s.calculatedColWidths[tableColIdx] + s.config.StrokeWidth
				tableColIdx++
				continue // Don't advance visual cell idx here
			}
		}

		vSpan := 1
		isVSpanStart := false
		if cellCtx.Merge.Vertical.Present && cellCtx.Merge.Vertical.Start {
			vSpan = cellCtx.Merge.Vertical.Span
			isVSpanStart = true
		} else if cellCtx.Merge.Hierarchical.Present && cellCtx.Merge.Hierarchical.Start {
			vSpan = cellCtx.Merge.Hierarchical.Span
			isVSpanStart = true
		}
		if vSpan <= 0 {
			vSpan = 1
		}

		rectWidth := 0.0
		for hs := 0; hs < hSpan; hs++ {
			if (tableColIdx + hs) < len(s.calculatedColWidths) {
				rectWidth += s.calculatedColWidths[tableColIdx+hs]
			} else {
				// This span goes beyond calculated columns, use MinColWidth
				rectWidth += s.config.MinColWidth
			}
		}
		if hSpan > 1 {
			rectWidth += float64(hSpan-1) * s.config.StrokeWidth
		}

		if rectWidth <= 0 {
			// s.debug("Table Col %d: Calculated rectWidth is zero. Advancing.", tableColIdx) // Verbose
			tableColIdx += hSpan
			// Only advance visual index if hSpan > 0, otherwise it was likely a column skipped due to H-merge continuation handled above
			if hSpan > 0 {
				currentVisualCellIdx++
			}
			continue
		}

		rectHeight := singleVisualRowHeight
		if isVSpanStart && vSpan > 1 {
			rectHeight = float64(vSpan)*singleVisualRowHeight + float64(vSpan-1)*s.config.StrokeWidth
			// Track v-merge state for *all* columns covered by hSpan, starting from tableColIdx
			for hs := 0; hs < hSpan; hs++ {
				if (tableColIdx + hs) < s.maxCols {
					s.vMergeTrack[tableColIdx+hs] = vSpan
					if hs > 0 { // Log columns affected by H-merge combined with V-merge start
						s.debug("Table Col %d also part of VMerge span %d due to HMerge start at %d", tableColIdx+hs, vSpan, tableColIdx)
					}
				}
			}
			if hSpan == 1 { // Log only if it's a simple V-merge start
				s.debug("Table Col %d: Starting VMerge, vSpan: %d, rectHeight: %.2f", tableColIdx, vSpan, rectHeight)
			}
		}

		fmt.Fprintf(&s.svgElements, `  <rect x="%.2f" y="%.2f" width="%.2f" height="%.2f" fill="%s"/>%s`,
			currentX, s.currentY, rectWidth, rectHeight, html.EscapeString(bgColor), "\n")

		cellTextAnchor := defaultTextAnchor
		if s.config.RenderTWConfigOverrides {
			if al := s.getSVGAnchorFromTW(cellCtx.Align); al != "" {
				cellTextAnchor = al
			}
		}

		textX := currentX + s.config.Padding
		if cellTextAnchor == "middle" {
			textX = currentX + rectWidth/2.0
		} else if cellTextAnchor == "end" {
			textX = currentX + rectWidth - s.config.Padding
		}
		// Adjust Y for vertical centering within the potentially taller merged cell rectangle
		textY := s.currentY + rectHeight/2.0

		escapedCell := html.EscapeString(textToRender)
		fmt.Fprintf(&s.svgElements, `  <text x="%.2f" y="%.2f" fill="%s" text-anchor="%s" dominant-baseline="middle">%s</text>%s`,
			textX, textY, html.EscapeString(textColor), cellTextAnchor, escapedCell, "\n")

		currentX += rectWidth + s.config.StrokeWidth
		tableColIdx += hSpan
		currentVisualCellIdx++ // This visual cell data is processed
	}
	s.currentY += singleVisualRowHeight + s.config.StrokeWidth
}

func (s *SVGRenderer) getSVGAnchorFromTW(align tw.Align) string {
	switch align {
	case tw.AlignLeft:
		return "start"
	case tw.AlignCenter:
		return "middle"
	case tw.AlignRight:
		return "end"
	case tw.AlignNone, tw.Skip:
		return ""
	}
	return ""
}

func (s *SVGRenderer) Close(w io.Writer) error {
	s.debug("Close called - Assembling SVG.")
	s.calculateAllColumnWidths()
	s.renderBufferedData()

	s.debug("After renderBufferedData: numVisualRowsDrawn=%d, maxCols=%d", s.numVisualRowsDrawn, s.maxCols) // ADDED DEBUG

	if s.numVisualRowsDrawn == 0 && s.maxCols == 0 {
		s.debug("Condition met: No data rendered, writing empty SVG placeholder.")
		fmt.Fprintf(w, `<svg xmlns="http://www.w3.org/2000/svg" width="%.2f" height="%.2f"></svg>`, s.config.StrokeWidth*2, s.config.StrokeWidth*2)
		return nil
	}

	// --- Calculate total dimensions ---
	totalWidth := s.config.StrokeWidth
	if len(s.calculatedColWidths) > 0 {
		for _, cw := range s.calculatedColWidths {
			colWidthForCalc := cw
			if colWidthForCalc <= 0 {
				colWidthForCalc = s.config.MinColWidth
			}
			totalWidth += colWidthForCalc + s.config.StrokeWidth
		}
	} else if s.maxCols > 0 {
		for i := 0; i < s.maxCols; i++ {
			totalWidth += s.config.MinColWidth + s.config.StrokeWidth
		}
	} else {
		totalWidth = s.config.StrokeWidth * 2
	}

	totalHeight := s.currentY
	singleVisualRowHeight := s.config.FontSize*s.config.LineHeightFactor + (2 * s.config.Padding)
	if s.numVisualRowsDrawn == 0 {
		if s.maxCols > 0 {
			totalHeight = s.config.StrokeWidth + singleVisualRowHeight + s.config.StrokeWidth
		} else {
			totalHeight = s.config.StrokeWidth * 2
		}
	}

	s.debug("Final SVG dimensions: W=%.2f H=%.2f (numVisualRowsDrawn: %d, maxCols: %d)", totalWidth, totalHeight, s.numVisualRowsDrawn, s.maxCols)

	// --- Write SVG Header ---
	fmt.Fprintf(w, `<svg xmlns="http://www.w3.org/2000/svg" width="%.2f" height="%.2f" font-family="%s" font-size="%.2fpx">`,
		totalWidth, totalHeight, html.EscapeString(s.config.FontFamily), s.config.FontSize)
	fmt.Fprintln(w)
	fmt.Fprintln(w, "<style>text { stroke: none; }</style>")

	// --- Write Cell Rects and Text ---
	_, err := io.WriteString(w, s.svgElements.String())
	if err != nil {
		s.debug("Error writing buffered SVG elements: %v", err)
		fmt.Fprintln(w, `</svg>`)
		return fmt.Errorf("failed to write SVG elements: %w", err)
	}

	// --- Draw Borders (Simple Full Grid) ---
	if s.maxCols > 0 || s.numVisualRowsDrawn > 0 {
		fmt.Fprintf(w, `  <g class="table-borders" stroke="%s" stroke-width="%.2f" stroke-linecap="square">`,
			html.EscapeString(s.config.StrokeColor), s.config.StrokeWidth)
		fmt.Fprintln(w)

		// --- Horizontal Lines ---
		yPos := s.config.StrokeWidth / 2.0
		borderRowsToDraw := s.numVisualRowsDrawn
		if borderRowsToDraw == 0 && s.maxCols > 0 {
			borderRowsToDraw = 1
		}
		lineStartX := s.config.StrokeWidth / 2.0
		lineEndX := totalWidth - s.config.StrokeWidth/2.0

		for i := 0; i <= borderRowsToDraw; i++ {
			fmt.Fprintf(w, `    <line x1="%.2f" y1="%.2f" x2="%.2f" y2="%.2f" />%s`,
				lineStartX, yPos, lineEndX, yPos, "\n")
			if i < borderRowsToDraw {
				yPos += singleVisualRowHeight + s.config.StrokeWidth
			}
		}

		// --- Vertical Lines ---
		xPos := s.config.StrokeWidth / 2.0
		borderLineStartY := s.config.StrokeWidth / 2.0
		borderLineEndY := totalHeight - (s.config.StrokeWidth / 2.0)

		for i := 0; i <= s.maxCols; i++ {
			fmt.Fprintf(w, `    <line x1="%.2f" y1="%.2f" x2="%.2f" y2="%.2f" />%s`,
				xPos, borderLineStartY, xPos, borderLineEndY, "\n")
			if i < s.maxCols {
				colWidth := s.config.MinColWidth
				if i < len(s.calculatedColWidths) {
					if s.calculatedColWidths[i] > 0 {
						colWidth = s.calculatedColWidths[i]
					} else if s.maxCols > 0 && s.calculatedColWidths[i] <= 0 {
						// Use MinColWidth if explicitly calculated as zero but part of the table structure
						colWidth = s.config.MinColWidth
					}
				}
				xPos += colWidth + s.config.StrokeWidth
			}
		}
		fmt.Fprintln(w, "  </g>")
	}

	fmt.Fprintln(w, `</svg>`)
	s.debug("SVG generation complete.")
	return nil
}

// padLineSVG remains the same
func padLineSVG(line []string, numCols int) []string {
	if numCols <= 0 {
		return []string{}
	}
	currentLen := len(line)
	if currentLen == numCols {
		return line
	}
	if currentLen > numCols {
		return line[:numCols]
	}
	padded := make([]string, numCols)
	copy(padded, line)
	return padded
}
