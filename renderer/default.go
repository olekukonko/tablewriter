// default.go
package renderer

import (
	"fmt"
	"github.com/olekukonko/tablewriter/tw"
	"github.com/olekukonko/tablewriter/twfn"
	"io"
	"strings"
)

// Formatting encapsulates the complete formatting context for a table row.
// It provides all necessary information to render a row correctly within the table structure.
type Formatting struct {
	Row              RowContext // Detailed configuration for the row and its cells
	Level            tw.Level   // Hierarchical level (Header, Body, Footer) affecting line drawing
	HasFooter        bool       // Indicates if the table includes a footer section
	IsSubRow         bool       // Marks this as a continuation or padding line in multi-line rows
	Debug            bool       // Enables debug logging when true
	NormalizedWidths tw.Mapper[int, int]
}

// CellContext defines the properties and formatting state of an individual table cell.
type CellContext struct {
	Data    string     // Content to be displayed in the cell, provided by the caller
	Align   tw.Align   // Text alignment within the cell (Left, Right, Center, Skip)
	Padding tw.Padding // Padding characters surrounding the cell content
	Width   int        // Suggested width (often overridden by Row.Widths)
	Merge   MergeState // Details about cell spanning across rows or columns
}

// MergeStateOption represents common attributes for merging in a specific direction.
type MergeStateOption struct {
	Present bool // True if this merge direction is active
	Span    int  // Number of cells this merge spans
	Start   bool // True if this cell is the starting point of the merge
	End     bool // True if this cell is the ending point of the merge
}

// MergeState captures how a cell merges across different directions.
type MergeState struct {
	Vertical     MergeStateOption // Properties for vertical merging (across rows)
	Horizontal   MergeStateOption // Properties for horizontal merging (across columns)
	Hierarchical MergeStateOption // Properties for nested/hierarchical merging
}

// RowContext manages layout properties and relationships for a row and its columns.
// It maintains state about the current row and its neighbors for proper rendering.
type RowContext struct {
	Position     tw.Position         // Section of the table (Header, Row, Footer)
	Location     tw.Location         // Boundary position (First, Middle, End)
	Current      map[int]CellContext // Cells in this row, indexed by column
	Previous     map[int]CellContext // Cells from the row above; nil if none
	Next         map[int]CellContext // Cells from the row below; nil if none
	Widths       tw.Mapper[int, int] // Computed widths for each column
	ColMaxWidths map[int]int         // Maximum allowed width per column
}

// Renderer defines the interface for rendering tables to an io.Writer.
// Implementations must handle headers, rows, footers, and separator lines.
type Renderer interface {
	Header(w io.Writer, headers [][]string, ctx Formatting) // Renders table header
	Row(w io.Writer, row []string, ctx Formatting)          // Renders a single row
	Footer(w io.Writer, footers [][]string, ctx Formatting) // Renders table footer
	Line(w io.Writer, ctx Formatting)                       // Renders separator line
	Debug() []string                                        // Returns debug trace
	Config() DefaultConfig                                  // Returns renderer config
}

// Separators controls the visibility of separators in the table.
type Separators struct {
	ShowHeader     tw.State // Controls header separator visibility
	ShowFooter     tw.State // Controls footer separator visibility
	BetweenRows    tw.State // Determines if lines appear between rows
	BetweenColumns tw.State // Determines if separators appear between columns
}

// Lines manages the visibility of table boundary lines.
type Lines struct {
	ShowTop        tw.State // Top border visibility
	ShowBottom     tw.State // Bottom border visibility
	ShowHeaderLine tw.State // Header separator line visibility
	ShowFooterLine tw.State // Footer separator line visibility
}

// Settings holds configuration preferences for rendering behavior.
type Settings struct {
	Separators     Separators // Separator visibility settings
	Lines          Lines      // Line visibility settings
	TrimWhitespace tw.State   // Trims whitespace from cell content if enabled
	CompactMode    tw.State   // Reserved for future compact rendering (unused)
}

// Border defines the visibility states of table borders.
type Border struct {
	Left   tw.State // Left border visibility
	Right  tw.State // Right border visibility
	Top    tw.State // Top border visibility
	Bottom tw.State // Bottom border visibility
}

var BorderNone = Border{tw.Off, tw.Off, tw.Off, tw.Off}

// Default is the default implementation of the Renderer interface.
type Default struct {
	config DefaultConfig // Configuration for rendering
	trace  []string      // Debug trace messages
}

// DefaultConfig holds the configuration for the default renderer.
type DefaultConfig struct {
	Borders  Border     // Border visibility settings
	Symbols  tw.Symbols // Symbols used for table drawing
	Settings Settings   // Rendering behavior settings
	Debug    bool       // Enables debug mode
}

// Config returns the current renderer configuration.
func (f *Default) Config() DefaultConfig {
	return f.config
}

// debug logs a debug message if debugging is enabled.
func (f *Default) debug(format string, a ...interface{}) {
	if f.config.Debug {
		msg := fmt.Sprintf(format, a...)
		f.trace = append(f.trace, fmt.Sprintf("[DEFAULT] %s", msg))
	}
}

// Debug returns the accumulated debug trace.
func (f *Default) Debug() []string {
	return f.trace
}

// Header renders the table header section.
func (f *Default) Header(w io.Writer, headers [][]string, ctx Formatting) {
	f.debug("Starting Header render: IsSubRow=%v, Location=%v, Pos=%s, lines=%d, widths=%v",
		ctx.IsSubRow, ctx.Row.Location, ctx.Row.Position, len(ctx.Row.Current), ctx.Row.Widths)
	f.renderLine(w, ctx)
	f.debug("Completed Header render")
}

// Row renders a single table row.
func (f *Default) Row(w io.Writer, row []string, ctx Formatting) {
	f.debug("Starting Row render: IsSubRow=%v, Location=%v, Pos=%s, hasFooter=%v",
		ctx.IsSubRow, ctx.Row.Location, ctx.Row.Position, ctx.HasFooter)
	f.renderLine(w, ctx)
	f.debug("Completed Row render")
}

// renderer/default.go

// renderer/default.go

func (f *Default) renderLine(w io.Writer, ctx Formatting) {
	// Input ctx.Row.Widths contains:
	// - Header/Footer: Pre-adjusted merged widths
	// - Row: Normalized widths (unadjusted for merges)
	// ctx.NormalizedWidths contains:
	// - Globally normalized widths (unadjusted for merges)

	sortedKeys := twfn.ConvertToSortedKeys(ctx.Row.Widths) // Use keys from the potentially adjusted map
	numCols := 0
	if len(sortedKeys) > 0 {
		numCols = sortedKeys[len(sortedKeys)-1] + 1
	} else {
		// Handle empty row case... (unchanged)
		prefix := ""
		if f.config.Borders.Left.Enabled() {
			prefix = f.config.Symbols.Column()
		}
		suffix := ""
		if f.config.Borders.Right.Enabled() {
			suffix = f.config.Symbols.Column()
		}
		if prefix != "" || suffix != "" {
			fmt.Fprintln(w, prefix+suffix)
		}
		f.debug("renderLine: Handled empty row/widths case.")
		return
	}

	columnSeparator := f.config.Symbols.Column()
	prefix := ""
	if f.config.Borders.Left.Enabled() {
		prefix = columnSeparator
	}
	suffix := ""
	if f.config.Borders.Right.Enabled() {
		suffix = columnSeparator
	}

	var output strings.Builder
	output.WriteString(prefix)

	colIndex := 0
	separatorDisplayWidth := 0
	if f.config.Settings.Separators.BetweenColumns.Enabled() {
		separatorDisplayWidth = twfn.DisplayWidth(columnSeparator)
	}

	for colIndex < numCols {
		// Add column separator IF:
		// - Not the first column
		// - Separators enabled
		// - Current column is NOT horizontally merged-into from left
		shouldAddSeparator := false
		if colIndex > 0 && f.config.Settings.Separators.BetweenColumns.Enabled() {
			cellCtx, ok := ctx.Row.Current[colIndex]
			if !ok || !(cellCtx.Merge.Horizontal.Present && !cellCtx.Merge.Horizontal.Start) {
				shouldAddSeparator = true
			}
		}
		if shouldAddSeparator {
			output.WriteString(columnSeparator)
			f.debug("renderLine: Added separator '%s' before col %d", columnSeparator, colIndex)
		} else if colIndex > 0 {
			f.debug("renderLine: Skipped separator before col %d due to HMerge continuation", colIndex)
		}

		cellCtx, ok := ctx.Row.Current[colIndex]

		// Determine the correct *visual* width for formatting this cell
		visualWidth := 0
		isHMergeStart := ok && cellCtx.Merge.Horizontal.Present && cellCtx.Merge.Horizontal.Start
		span := 1

		if isHMergeStart {
			span = cellCtx.Merge.Horizontal.Span
			if ctx.Row.Position == tw.Row {
				// Calculate dynamic width for ROW H-merge using NORMALIZED base widths
				dynamicTotalWidth := 0
				for k := 0; k < span && colIndex+k < numCols; k++ {
					colToSum := colIndex + k
					normWidth := ctx.NormalizedWidths.Get(colToSum) // Use normalized
					if normWidth < 0 {
						normWidth = 0
					}
					dynamicTotalWidth += normWidth
					if k > 0 && separatorDisplayWidth > 0 {
						dynamicTotalWidth += separatorDisplayWidth
					}
				}
				visualWidth = dynamicTotalWidth
				f.debug("renderLine: Row HMerge col %d, span %d, dynamic visualWidth %d", colIndex, span, visualWidth)
			} else {
				// For Header/Footer, the pre-adjusted width IS the visual width
				visualWidth = ctx.Row.Widths.Get(colIndex)
				f.debug("renderLine: H/F HMerge col %d, span %d, pre-adjusted visualWidth %d", colIndex, span, visualWidth)
			}
		} else {
			// For regular cells, visual width comes from the section's width map
			visualWidth = ctx.Row.Widths.Get(colIndex)
			f.debug("renderLine: Regular col %d, visualWidth %d", colIndex, visualWidth)
		}
		if visualWidth < 0 {
			visualWidth = 0
		}

		// Skip processing for cells that are visually part of a preceding H-merge
		// These should have visualWidth=0 based on how applyHorizontalMergeWidths works for H/F
		// or because they weren't the starting 'colIndex' in the row case.
		// Need a reliable check: use the merge state directly.
		if ok && cellCtx.Merge.Horizontal.Present && !cellCtx.Merge.Horizontal.Start {
			f.debug("renderLine: Skipping col %d processing (part of HMerge)", colIndex)
			colIndex++
			continue
		}

		// If no cell context, just draw spaces for the calculated visual width (can be 0)
		if !ok {
			if visualWidth > 0 {
				output.WriteString(strings.Repeat(" ", visualWidth))
				f.debug("renderLine: No cell context for col %d, writing %d spaces", colIndex, visualWidth)
			} else {
				f.debug("renderLine: No cell context for col %d, visualWidth is 0, writing nothing", colIndex)
			}
			colIndex += span // Advance by span (usually 1 if !ok)
			continue
		}

		// --- We have cell context ---

		// Handle alignment, padding, blanking for V/H merges
		padding := cellCtx.Padding
		align := cellCtx.Align
		if align == tw.Skip {
			align = tw.AlignLeft
		}

		// Override alignment for Footer/TOTAL pattern
		isTotalPattern := false
		if colIndex == 0 && isHMergeStart && cellCtx.Merge.Horizontal.Span >= 3 && strings.TrimSpace(cellCtx.Data) == "TOTAL" {
			isTotalPattern = true
			f.debug("renderLine: Detected 'TOTAL' HMerge pattern at col 0")
		}
		if (ctx.Row.Position == tw.Footer && isHMergeStart) || isTotalPattern {
			f.debug("renderLine: Applying AlignRight override for Footer/TOTAL pattern at col %d. Original align was: %v", colIndex, align)
			align = tw.AlignRight
		}

		cellData := cellCtx.Data
		if (cellCtx.Merge.Vertical.Present && !cellCtx.Merge.Vertical.Start) ||
			(cellCtx.Merge.Hierarchical.Present && !cellCtx.Merge.Hierarchical.Start) {
			cellData = ""
			f.debug("renderLine: Blanked data for col %d (non-start V/Hierarchical)", colIndex)
		}

		// Format and write using the determined visualWidth
		formattedCell := f.formatCell(cellData, visualWidth, padding, align)
		// Only write if non-empty (formatCell returns "" for width 0)
		if len(formattedCell) > 0 {
			output.WriteString(formattedCell)
		}

		if isHMergeStart {
			f.debug("renderLine: Rendered HMerge START col %d (span %d, visualWidth %d, align %v): '%s'",
				colIndex, span, visualWidth, align, formattedCell)
		} else {
			f.debug("renderLine: Rendered regular col %d (visualWidth %d, align %v): '%s'",
				colIndex, visualWidth, align, formattedCell)
		}
		// Advance index by the span (1 for regular, >1 for HMerge start)
		colIndex += span
	}

	output.WriteString(suffix)
	output.WriteString(tw.NewLine)
	fmt.Fprint(w, output.String())
	f.debug("renderLine: Final rendered line: %s", strings.TrimSuffix(output.String(), tw.NewLine))
}

// Fix 13: formatCell return empty string for width <= 0 (Re-verify)
func (f *Default) formatCell(content string, width int, padding tw.Padding, align tw.Align) string {
	// Return immediately if width is non-positive
	if width <= 0 {
		// f.debug("formatCell: width %d <= 0, returning empty string", width) // Debug line removed for brevity
		return "" // Ensure it returns truly empty string
	}

	// --- Rest of the function remains the same ---
	f.debug("Formatting cell: content='%s', width=%d, align=%s, padding={L:'%s' R:'%s'}",
		content, width, align, padding.Left, padding.Right)
	if f.config.Settings.TrimWhitespace.Enabled() {
		content = strings.TrimSpace(content)
		f.debug("Trimmed content: '%s'", content)
	}

	runeWidth := twfn.DisplayWidth(content)
	padLeftWidth := twfn.DisplayWidth(padding.Left)
	padRightWidth := twfn.DisplayWidth(padding.Right)
	totalPaddingWidth := padLeftWidth + padRightWidth

	availableContentWidth := width - totalPaddingWidth
	if availableContentWidth < 0 {
		availableContentWidth = 0
	}
	f.debug("Available content width: %d", availableContentWidth)

	if runeWidth > availableContentWidth {
		content = twfn.TruncateString(content, availableContentWidth)
		runeWidth = twfn.DisplayWidth(content)
		f.debug("Truncated content to fit %d: '%s' (new width %d)", availableContentWidth, content, runeWidth)
	}

	remainingSpace := width - runeWidth - totalPaddingWidth
	if remainingSpace < 0 {
		remainingSpace = 0
	}
	f.debug("Remaining space for alignment padding: %d", remainingSpace)

	leftPadChar := padding.Left // Use variable for clarity
	rightPadChar := padding.Right
	// Default padding character if empty
	if leftPadChar == "" {
		leftPadChar = tw.Space
	}
	if rightPadChar == "" {
		rightPadChar = tw.Space
	}
	leftPadCharWidth := twfn.DisplayWidth(leftPadChar)
	if leftPadCharWidth <= 0 {
		leftPadCharWidth = 1
	} // Safety
	rightPadCharWidth := twfn.DisplayWidth(rightPadChar)
	if rightPadCharWidth <= 0 {
		rightPadCharWidth = 1
	} // Safety

	var result strings.Builder
	var leftSpaces, rightSpaces int

	switch align {
	case tw.AlignLeft:
		leftSpaces = padLeftWidth
		rightSpaces = width - runeWidth - leftSpaces
	case tw.AlignRight:
		rightSpaces = padRightWidth
		leftSpaces = width - runeWidth - rightSpaces
	case tw.AlignCenter:
		leftSpaces = padLeftWidth + remainingSpace/2
		rightSpaces = width - runeWidth - leftSpaces
	default: // Default to AlignLeft
		leftSpaces = padLeftWidth
		rightSpaces = width - runeWidth - leftSpaces
	}

	// Ensure non-negative counts
	if leftSpaces < 0 {
		leftSpaces = 0
	}
	if rightSpaces < 0 {
		rightSpaces = 0
	}

	// Write left padding
	if leftPadCharWidth > 0 {
		leftRepeat := leftSpaces / leftPadCharWidth
		result.WriteString(strings.Repeat(leftPadChar, leftRepeat))
		result.WriteString(strings.Repeat(" ", leftSpaces%leftPadCharWidth)) // Remainder spaces
	} else {
		result.WriteString(strings.Repeat(" ", leftSpaces)) // Use space if pad char has no width
	}

	// Write content
	result.WriteString(content)

	// Write right padding
	if rightPadCharWidth > 0 {
		rightRepeat := rightSpaces / rightPadCharWidth
		result.WriteString(strings.Repeat(rightPadChar, rightRepeat))
		result.WriteString(strings.Repeat(" ", rightSpaces%rightPadCharWidth)) // Remainder spaces
	} else {
		result.WriteString(strings.Repeat(" ", rightSpaces)) // Use space if pad char has no width
	}

	output := result.String()
	// Ensure final output matches target width precisely, truncating if necessary
	finalWidth := twfn.DisplayWidth(output)
	if finalWidth > width {
		output = twfn.TruncateString(output, width)
		f.debug("formatCell: Final check truncated output to width %d", width)
	} else if finalWidth < width {
		output += strings.Repeat(" ", width-finalWidth) // Pad with spaces only
		f.debug("formatCell: Final check added %d spaces to meet width %d", width-finalWidth, width)
	}

	if f.config.Debug && twfn.DisplayWidth(output) != width {
		f.debug("formatCell Warning: Final width %d does not match target %d for result '%s'",
			twfn.DisplayWidth(output), width, output)
	}

	f.debug("Formatted cell final result: '%s' (target width %d)", output, width)
	return output
}

// Footer renders the table footer section.
func (f *Default) Footer(w io.Writer, footers [][]string, ctx Formatting) {
	f.debug("Starting Footer render: IsSubRow=%v, Location=%v, Pos=%s",
		ctx.IsSubRow, ctx.Row.Location, ctx.Row.Position)
	f.renderLine(w, ctx)
	f.debug("Completed Footer render")
}

// formatCell formats a cell's content according to width, padding, and alignment.
// It handles truncation and padding adjustments as needed.
