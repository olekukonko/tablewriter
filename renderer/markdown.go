package renderer

//// Markdown implements a simple Markdown table formatter
//type Markdown struct{}
//
//// NewMarkdown creates a new Markdown renderer
//func NewMarkdown() *Markdown {
//	return &Markdown{}
//}
//
//func (m *Markdown) formatCell(content string, width int, align tw.Align, padding tw.Padding) string {
//	content = strings.TrimSpace(content)
//	runeWidth := twfn.DisplayWidth(content)
//	padLeftWidth := twfn.DisplayWidth(padding.Left)
//	padRightWidth := twfn.DisplayWidth(padding.Right)
//	totalContentWidth := runeWidth + padLeftWidth + padRightWidth
//
//	if totalContentWidth > width {
//		width = totalContentWidth
//	}
//
//	gap := width - runeWidth - padLeftWidth - padRightWidth
//	if gap < 0 {
//		gap = 0
//	}
//
//	leftPadChar := " "
//	if padding.Left != "" {
//		leftPadChar = padding.Left
//	}
//	rightPadChar := " "
//	if padding.Right != "" {
//		rightPadChar = padding.Right
//	}
//
//	switch align {
//	case tw.AlignCenter:
//		leftGap := gap / 2
//		rightGap := gap - leftGap
//		return padding.Left + strings.Repeat(leftPadChar, leftGap) + content + strings.Repeat(rightPadChar, rightGap) + padding.Right
//	case tw.AlignRight:
//		return padding.Left + strings.Repeat(leftPadChar, gap) + content + padding.Right
//	default: // Left or default
//		return padding.Left + content + strings.Repeat(rightPadChar, gap) + padding.Right
//	}
//}
//
//func (m *Markdown) Header(w io.Writer, headers []string, ctx Formatting) {
//	fmt.Printf("DEBUG: headers = %v\n", headers) // Add this for debugging
//	// Format header cells with padding
//	cells := make([]string, len(headers))
//	for i, h := range headers {
//		width := ctx.Widths[i]
//		align := ctx.Align
//		if colAlign, ok := ctx.ColAligns[i]; ok {
//			align = colAlign
//		}
//		padding := ctx.Padding
//		if customPad, ok := ctx.ColPadding[i]; ok {
//			padding = customPad
//		}
//		cells[i] = m.formatCell(h, width, align, padding)
//	}
//	fmt.Fprintf(w, "|%s|\n", strings.Join(cells, "|"))
//
//	// Format separator row
//	separators := make([]string, len(headers))
//	for i := range headers {
//		width := ctx.Widths[i]
//		align := ctx.Align
//		if colAlign, ok := ctx.ColAligns[i]; ok {
//			align = colAlign
//		}
//
//		switch align {
//		case tw.AlignCenter:
//			separators[i] = ":" + strings.Repeat("-", width-2) + ":"
//		case tw.AlignRight:
//			separators[i] = strings.Repeat("-", width-1) + ":"
//		default: // Left or default
//			separators[i] = ":" + strings.Repeat("-", width-1)
//		}
//	}
//	fmt.Fprintf(w, "|%s|\n", strings.Join(separators, "|"))
//}
//
//func (m *Markdown) Row(w io.Writer, row []string, ctx Formatting) {
//	cells := make([]string, len(row))
//	for i, cell := range row {
//		width := ctx.Widths[i]
//		align := ctx.Align
//		if colAlign, ok := ctx.ColAligns[i]; ok {
//			align = colAlign
//		}
//		padding := ctx.Padding
//		if customPad, ok := ctx.ColPadding[i]; ok {
//			padding = customPad
//		}
//		cells[i] = m.formatCell(cell, width, align, padding)
//	}
//	fmt.Fprintf(w, "|%s|\n", strings.Join(cells, "|"))
//}
//
//func (m *Markdown) Footer(w io.Writer, footers []string, ctx Formatting) {
//	cells := make([]string, len(footers))
//	for i, f := range footers {
//		width := ctx.Widths[i]
//		align := ctx.Align
//		if colAlign, ok := ctx.ColAligns[i]; ok {
//			align = colAlign
//		}
//		padding := ctx.Padding
//		if customPad, ok := ctx.ColPadding[i]; ok {
//			padding = customPad
//		}
//		cells[i] = m.formatCell(f, width, align, padding)
//	}
//	fmt.Fprintf(w, "|%s|\n", strings.Join(cells, "|"))
//}
//
//func (m *Markdown) Line(w io.Writer, ctx Formatting) {
//	// SymbolMarkdown tables don't need additional lines
//	return
//}
//
//func (m *Markdown) GetColumnWidths() []int {
//	return nil
//}
//
//func (m *Markdown) Reset() {}
