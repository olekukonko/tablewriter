package renderer

import (
	"fmt"
	"github.com/olekukonko/tablewriter/tw"
	"github.com/olekukonko/tablewriter/twfn"
	"io"
	"strings"
)

// JunctionRenderer handles junction rendering with row context
type JunctionRenderer struct {
	sym       tw.Symbols
	ctx       Formatting
	colIdx    int
	debugging bool
	debug     func(format string, a ...interface{})
}

// NewJunctionRenderer creates a new JunctionRenderer instance
func NewJunctionRenderer(symbols tw.Symbols, ctx Formatting, colIdx int, debug func(format string, a ...interface{})) *JunctionRenderer {
	return &JunctionRenderer{
		sym:       symbols,
		ctx:       ctx,
		colIdx:    colIdx,
		debugging: true,
		debug:     debug,
	}
}

// log logs debug messages if debugging is enabled
func (jr *JunctionRenderer) log(format string, a ...interface{}) {
	if jr.debugging {
		jr.debug("[JUNCTION] "+format, a...)
	}
}

// getMergeState safely retrieves the merge state for a column
func (jr *JunctionRenderer) getMergeState(row map[int]CellContext, colIdx int) MergeState {
	if row == nil || colIdx < 0 || colIdx >= len(jr.ctx.Row.Widths) {
		return MergeState{}
	}
	if cell, ok := row[colIdx]; ok {
		// Check for full-span merges
		if cell.Merge.Horizontal && cell.Merge.Span == len(jr.ctx.Row.Widths) {
			return MergeState{Horizontal: true, Start: true, End: true, Span: cell.Merge.Span}
		}
		return cell.Merge
	}
	return MergeState{}
}

// findMergeStart finds the start column index of a merge that includes the given column
func (jr *JunctionRenderer) findMergeStart(row map[int]CellContext, colIdx int) int {
	if row == nil || colIdx < 0 {
		return colIdx
	}

	merge := jr.getMergeState(row, colIdx)
	if !merge.Horizontal || merge.Start {
		return colIdx
	}

	// Look backwards to find the start column
	for i := colIdx - 1; i >= 0; i-- {
		m := jr.getMergeState(row, i)
		if m.Horizontal && m.Start {
			return i
		}
	}
	return colIdx
}

// isWithinMerge checks if the column is within a horizontal merge
func (jr *JunctionRenderer) isWithinMerge(row map[int]CellContext, colIdx int, nextColIdx int) bool {
	if row == nil {
		return false
	}

	startIdx := jr.findMergeStart(row, colIdx)
	merge := jr.getMergeState(row, startIdx)

	if !merge.Horizontal {
		return false
	}

	// Check if both columns are within the same merge span
	return colIdx >= startIdx && nextColIdx < startIdx+merge.Span
}

// isMergeJunction checks if this is a junction at the start or end of a merge
func (jr *JunctionRenderer) isMergeJunction(row map[int]CellContext, colIdx int, nextColIdx int) (bool, bool) {
	if row == nil {
		return false, false
	}

	// Check if this is a merge start junction
	startMerge := false
	endMerge := false

	leftMerge := jr.getMergeState(row, colIdx)
	rightMerge := jr.getMergeState(row, nextColIdx)

	// If the left column isn't in a merge but the right column starts a merge
	if (!leftMerge.Horizontal || leftMerge.End) && rightMerge.Horizontal && rightMerge.Start {
		startMerge = true
	}

	// If the left column is the end of a merge
	if leftMerge.Horizontal && leftMerge.End {
		endMerge = true
	}

	return startMerge, endMerge
}

// RenderLeft returns the left border symbol based on level and location
func (jr *JunctionRenderer) RenderLeft() string {
	jr.log("Rendering left border: level=%d, location=%v", jr.ctx.Level, jr.ctx.Row.Location)
	switch jr.ctx.Level {
	case tw.LevelHeader:
		if jr.ctx.Row.Location == tw.LocationFirst {
			return jr.sym.TopLeft() // ┌
		}
		return jr.sym.MidLeft() // ├
	case tw.LevelBody:
		if jr.ctx.Row.Location == tw.LocationFirst {
			return jr.sym.MidLeft() // ├
		}
		if jr.ctx.Row.Location == tw.LocationEnd {
			return jr.sym.BottomLeft() // └
		}
		return jr.sym.MidLeft() // ├
	case tw.LevelFooter:
		if jr.ctx.Row.Location == tw.LocationEnd {
			return jr.sym.BottomLeft() // └
		}
		return jr.sym.MidLeft() // ├
	}
	return jr.sym.MidLeft() // Default: ├
}

// RenderRight returns the right border symbol based on level and location
func (jr *JunctionRenderer) RenderRight(lastColIdx int) string {
	jr.log("Rendering right border: level=%d, location=%v, lastColIdx=%d", jr.ctx.Level, jr.ctx.Row.Location, lastColIdx)
	currentMerge := jr.getMergeState(jr.ctx.Row.Current, lastColIdx)
	isFullSpanMerge := currentMerge.Horizontal && currentMerge.End && currentMerge.Span == len(jr.ctx.Row.Widths)

	switch jr.ctx.Level {
	case tw.LevelHeader:
		if jr.ctx.Row.Location == tw.LocationFirst {
			return jr.sym.TopRight() // ┐
		}
		return jr.sym.MidRight() // ┤
	case tw.LevelBody:
		if jr.ctx.Row.Location == tw.LocationEnd {
			if isFullSpanMerge {
				return jr.sym.BottomRight() // ┘ for full-span merge
			}
			return jr.sym.BottomRight() // ┘
		}
		return jr.sym.MidRight() // ┤
	case tw.LevelFooter:
		if jr.ctx.Row.Location == tw.LocationEnd {
			if isFullSpanMerge {
				return jr.sym.BottomRight() // ┘ for full-span merge
			}
			return jr.sym.BottomRight() // ┘
		}
		return jr.sym.MidRight() // ┤
	}
	return jr.sym.MidRight() // Default: ┤
}

// GetSegment returns the horizontal segment character
func (jr *JunctionRenderer) GetSegment() string {
	currentMerge := jr.getMergeState(jr.ctx.Row.Current, jr.colIdx)
	jr.log("Getting segment for colIdx=%d, merge=%v", jr.colIdx, currentMerge)
	if currentMerge.Horizontal {
		return jr.sym.Row() // ─ for horizontal merge
	}
	return jr.sym.Row() // ─ for basic rendering
}

// RenderJunction returns the junction character based on the merge state
// RenderJunction returns the junction character based on the merge state
func (jr *JunctionRenderer) RenderJunction(nextColIdx int) string {
	jr.log("Rendering junction: colIdx=%d, nextColIdx=%d, level=%d, location=%v", jr.colIdx, nextColIdx, jr.ctx.Level, jr.ctx.Row.Location)

	// Determine if there’s a vertical line above (no merge in current row)
	var hasVerticalAbove bool
	if jr.ctx.Row.Current == nil {
		hasVerticalAbove = false // Top border case
	} else {
		hasVerticalAbove = !jr.isWithinMerge(jr.ctx.Row.Current, jr.colIdx, nextColIdx)
	}

	// Determine if there’s a vertical line below (no merge in next row)
	var hasVerticalBelow bool
	if jr.ctx.Row.Next == nil {
		hasVerticalBelow = false // Bottom border case
	} else {
		hasVerticalBelow = !jr.isWithinMerge(jr.ctx.Row.Next, jr.colIdx, nextColIdx)
	}

	// Select junction character based on vertical line presence
	if hasVerticalAbove && hasVerticalBelow {
		return jr.sym.Center() // ┼
	} else if hasVerticalAbove && !hasVerticalBelow {
		return jr.sym.BottomMid() // ┴
	} else if !hasVerticalAbove && hasVerticalBelow {
		return jr.sym.TopMid() // ┬
	} else {
		return jr.sym.Row() // ─
	}
}

// Line function (unchanged from your last working version)
func (f *Default) Line(w io.Writer, ctx Formatting) {
	f.debug("Starting Line render: level=%d, pos=%s, loc=%v, widths=%v", ctx.Level, ctx.Row.Position, ctx.Row.Location, ctx.Row.Widths)
	sortedKeys := twfn.ConvertToSortedKeys(ctx.Row.Widths)
	if len(sortedKeys) == 0 {
		if f.config.Borders.Left.Enabled() {
			fmt.Fprint(w, f.config.Symbols.Column())
		}
		if f.config.Borders.Right.Enabled() {
			fmt.Fprint(w, f.config.Symbols.Column())
		}
		fmt.Fprintln(w)
		return
	}

	var line strings.Builder
	lastColIdx := sortedKeys[len(sortedKeys)-1]

	// Full-span merge check
	fullSpanMerge := false
	totalWidth := 0
	if ctx.Row.Location == tw.LocationEnd && ctx.Row.Current != nil {
		jr := NewJunctionRenderer(f.config.Symbols, ctx, 0, f.debug)
		for colIdx := range ctx.Row.Current {
			merge := jr.getMergeState(ctx.Row.Current, colIdx)
			if merge.Horizontal && merge.Start && merge.Span == len(sortedKeys) {
				fullSpanMerge = true
				break
			}
		}
	}
	if fullSpanMerge {
		for _, width := range ctx.Row.Widths {
			totalWidth += width
		}
		totalWidth += (len(sortedKeys) - 1)
	}
	f.debug("Full-span merge detected: %v, totalWidth=%d", fullSpanMerge, totalWidth)

	// Left border
	if f.config.Borders.Left.Enabled() {
		jr := NewJunctionRenderer(f.config.Symbols, ctx, sortedKeys[0], f.debug)
		line.WriteString(jr.RenderLeft())
	}

	// Columns and junctions
	if fullSpanMerge {
		jr := NewJunctionRenderer(f.config.Symbols, ctx, sortedKeys[0], f.debug)
		line.WriteString(strings.Repeat(jr.GetSegment(), totalWidth))
	} else {
		for i, colIdx := range sortedKeys {
			jr := NewJunctionRenderer(f.config.Symbols, ctx, colIdx, f.debug)
			if width, ok := ctx.Row.Widths[colIdx]; ok && width > 0 {
				line.WriteString(strings.Repeat(jr.GetSegment(), width))
			}
			if i < len(sortedKeys)-1 && f.config.Settings.Separators.BetweenColumns.Enabled() {
				nextColIdx := sortedKeys[i+1]
				line.WriteString(jr.RenderJunction(nextColIdx))
			}
		}
	}

	// Right border
	if f.config.Borders.Right.Enabled() {
		jr := NewJunctionRenderer(f.config.Symbols, ctx, lastColIdx, f.debug)
		line.WriteString(jr.RenderRight(lastColIdx))
	}

	fmt.Fprintln(w, line.String())
	f.debug("Line render completed: [%s]", line.String())
}
