// junction.go
package renderer

import (
	"fmt"
	"github.com/olekukonko/tablewriter/tw"
	"github.com/olekukonko/tablewriter/twfn"
	"io"
	"strings"
)

// JunctionRenderer manages how table junction points (corners, intersections, etc.) are rendered.
// It factors in cell merges and row context to select appropriate symbols like ┼, ├, │, etc.
type JunctionRenderer struct {
	sym       tw.Symbols                            // Symbols used for rendering (e.g., corners, lines).
	ctx       Formatting                            // Current formatting context, including row and merge states.
	colIdx    int                                   // Index of the column being processed.
	debugging bool                                  // Enables debug logging if true, inherited from ctx.Debug.
	debug     func(format string, a ...interface{}) // Function for logging debug messages.
}

// NewJunctionRenderer creates a new JunctionRenderer for handling junction symbols in the table.
// It falls back to a no-op logger if no debug function is provided to avoid nil dereferencing.
func NewJunctionRenderer(symbols tw.Symbols, ctx Formatting, colIdx int, debug func(format string, a ...interface{})) *JunctionRenderer {
	if debug == nil {
		debug = func(format string, a ...interface{}) {}
	}
	return &JunctionRenderer{
		sym:       symbols,
		ctx:       ctx,
		colIdx:    colIdx,
		debugging: ctx.Debug,
		debug:     debug,
	}
}

// log logs debug messages prefixed with [JUNCTION], if debugging is enabled.
// Useful for tracing merge decisions and layout computations.
func (jr *JunctionRenderer) log(format string, a ...interface{}) {
	if jr.debugging {
		jr.debug("[JUNCTION] "+format, a...)
	}
}

// getMergeState returns the merge state of a specific column in a row.
// Handles edge cases gracefully by returning a zero-value MergeState.
func (jr *JunctionRenderer) getMergeState(row map[int]CellContext, colIdx int) MergeState {
	if row == nil || colIdx < 0 {
		return MergeState{}
	}
	return row[colIdx].Merge
}

// GetSegment decides whether to draw a horizontal line or leave space.
// It returns a row symbol (─) or an empty string based on vertical merge continuity.
func (jr *JunctionRenderer) GetSegment() string {
	currentMerge := jr.getMergeState(jr.ctx.Row.Current, jr.colIdx)
	nextMerge := jr.getMergeState(jr.ctx.Row.Next, jr.colIdx)

	vPassThruStrict := (currentMerge.Vertical.Present && nextMerge.Vertical.Present && !currentMerge.Vertical.End && !nextMerge.Vertical.Start) ||
		(currentMerge.Hierarchical.Present && nextMerge.Hierarchical.Present && !currentMerge.Hierarchical.End && !nextMerge.Hierarchical.Start)

	if vPassThruStrict {
		jr.log("GetSegment col %d: VPassThruStrict=%v -> Empty segment", jr.colIdx, vPassThruStrict)
		return ""
	}
	jr.log("GetSegment col %d: VPassThruStrict=%v -> Row symbol", jr.colIdx, vPassThruStrict)
	return jr.sym.Row()
}

// RenderLeft chooses the leftmost symbol for the current row line.
// Considers top, bottom, or middle positioning and vertical merge pass-throughs.
func (jr *JunctionRenderer) RenderLeft() string {
	mergeAbove := jr.getMergeState(jr.ctx.Row.Current, 0)
	mergeBelow := jr.getMergeState(jr.ctx.Row.Next, 0)

	isTopBorder := (jr.ctx.Level == tw.LevelHeader && jr.ctx.Row.Location == tw.LocationFirst) ||
		(jr.ctx.Level == tw.LevelBody && jr.ctx.Row.Location == tw.LocationFirst && jr.ctx.Row.Previous == nil)
	if isTopBorder {
		return jr.sym.TopLeft()
	}

	isBottom := jr.ctx.Level == tw.LevelBody && jr.ctx.Row.Location == tw.LocationEnd && !jr.ctx.HasFooter
	isFooter := jr.ctx.Level == tw.LevelFooter && jr.ctx.Row.Location == tw.LocationEnd
	if isBottom || isFooter {
		return jr.sym.BottomLeft()
	}

	isVPassThruStrict := (mergeAbove.Vertical.Present && mergeBelow.Vertical.Present && !mergeAbove.Vertical.End && !mergeBelow.Vertical.Start) ||
		(mergeAbove.Hierarchical.Present && mergeBelow.Hierarchical.Present && !mergeAbove.Hierarchical.End && !mergeBelow.Hierarchical.Start)
	if isVPassThruStrict {
		return jr.sym.Column()
	}

	return jr.sym.MidLeft()
}

// RenderRight selects the rightmost junction or border character for the table.
// Handles top and bottom cases and vertical pass-through merges.
func (jr *JunctionRenderer) RenderRight(lastColIdx int) string {
	if lastColIdx < 0 {
		switch jr.ctx.Level {
		case tw.LevelHeader:
			return jr.sym.TopRight()
		case tw.LevelFooter:
			return jr.sym.BottomRight()
		default:
			if jr.ctx.Row.Location == tw.LocationFirst {
				return jr.sym.TopRight()
			}
			if jr.ctx.Row.Location == tw.LocationEnd {
				return jr.sym.BottomRight()
			}
			return jr.sym.MidRight()
		}
	}

	mergeAbove := jr.getMergeState(jr.ctx.Row.Current, lastColIdx)
	mergeBelow := jr.getMergeState(jr.ctx.Row.Next, lastColIdx)

	isTopBorder := (jr.ctx.Level == tw.LevelHeader && jr.ctx.Row.Location == tw.LocationFirst) ||
		(jr.ctx.Level == tw.LevelBody && jr.ctx.Row.Location == tw.LocationFirst && jr.ctx.Row.Previous == nil)
	if isTopBorder {
		return jr.sym.TopRight()
	}

	isBottom := jr.ctx.Level == tw.LevelBody && jr.ctx.Row.Location == tw.LocationEnd && !jr.ctx.HasFooter
	isFooter := jr.ctx.Level == tw.LevelFooter && jr.ctx.Row.Location == tw.LocationEnd
	if isBottom || isFooter {
		return jr.sym.BottomRight()
	}

	isVPassThruStrict := (mergeAbove.Vertical.Present && mergeBelow.Vertical.Present && !mergeAbove.Vertical.End && !mergeBelow.Vertical.Start) ||
		(mergeAbove.Hierarchical.Present && mergeBelow.Hierarchical.Present && !mergeAbove.Hierarchical.End && !mergeBelow.Hierarchical.Start)
	if isVPassThruStrict {
		return jr.sym.Column()
	}

	return jr.sym.MidRight()
}

// RenderJunction determines the junction symbol between two adjacent columns.
// Accounts for horizontal spans, merge transitions, and border roles.

func (jr *JunctionRenderer) RenderJunction(leftColIdx, rightColIdx int) string {
	mergeCurrentL := jr.getMergeState(jr.ctx.Row.Current, leftColIdx)
	mergeCurrentR := jr.getMergeState(jr.ctx.Row.Current, rightColIdx)
	mergeNextL := jr.getMergeState(jr.ctx.Row.Next, leftColIdx)
	mergeNextR := jr.getMergeState(jr.ctx.Row.Next, rightColIdx)

	isSpannedCurrent := mergeCurrentL.Horizontal.Present && !mergeCurrentL.Horizontal.End
	isSpannedNext := mergeNextL.Horizontal.Present && !mergeNextL.Horizontal.End

	vPassThruLStrict := (mergeCurrentL.Vertical.Present && mergeNextL.Vertical.Present && !mergeCurrentL.Vertical.End && !mergeNextL.Vertical.Start) ||
		(mergeCurrentL.Hierarchical.Present && mergeNextL.Hierarchical.Present && !mergeCurrentL.Hierarchical.End && !mergeNextL.Hierarchical.Start)
	vPassThruRStrict := (mergeCurrentR.Vertical.Present && mergeNextR.Vertical.Present && !mergeCurrentR.Vertical.End && !mergeNextR.Vertical.Start) ||
		(mergeCurrentR.Hierarchical.Present && mergeNextR.Hierarchical.Present && !mergeCurrentR.Hierarchical.End && !mergeNextR.Hierarchical.Start)

	isTop := (jr.ctx.Level == tw.LevelHeader && jr.ctx.Row.Location == tw.LocationFirst) ||
		(jr.ctx.Level == tw.LevelBody && jr.ctx.Row.Location == tw.LocationFirst && len(jr.ctx.Row.Previous) == 0)
	isBottom := (jr.ctx.Level == tw.LevelFooter && jr.ctx.Row.Location == tw.LocationEnd) ||
		(jr.ctx.Level == tw.LevelBody && jr.ctx.Row.Location == tw.LocationEnd && !jr.ctx.HasFooter)
	isPreFooter := (jr.ctx.Level == tw.LevelFooter && (jr.ctx.Row.Position == tw.Row || jr.ctx.Row.Position == tw.Header))

	if isTop {
		if isSpannedNext {
			return jr.sym.Row()
		}
		return jr.sym.TopMid()
	}

	if isBottom {
		if vPassThruLStrict && vPassThruRStrict {
			return jr.sym.Column()
		}
		if vPassThruLStrict {
			return jr.sym.MidLeft()
		}
		if vPassThruRStrict {
			return jr.sym.MidRight()
		}
		if isSpannedCurrent {
			return jr.sym.Row()
		}
		return jr.sym.BottomMid()
	}

	if isPreFooter {
		if vPassThruLStrict && vPassThruRStrict {
			return jr.sym.Column()
		}
		if vPassThruLStrict {
			return jr.sym.MidLeft()
		}
		if vPassThruRStrict {
			return jr.sym.MidRight()
		}
		if mergeCurrentL.Horizontal.Present {
			if !mergeCurrentL.Horizontal.End && mergeCurrentR.Horizontal.Present && !mergeCurrentR.Horizontal.End {
				jr.log("Footer separator: H-merge continues from col %d to %d (mid-span), using BottomMid", leftColIdx, rightColIdx)
				return jr.sym.BottomMid() // "┴" between merged columns (0→1)
			}
			if !mergeCurrentL.Horizontal.End && mergeCurrentR.Horizontal.Present && mergeCurrentR.Horizontal.End {
				jr.log("Footer separator: H-merge ends at col %d, using BottomMid", rightColIdx)
				return jr.sym.BottomMid() // "┴" at end of merge within span (1→2)
			}
			if mergeCurrentL.Horizontal.End && !mergeCurrentR.Horizontal.Present {
				jr.log("Footer separator: H-merge ends at col %d, next col %d not merged, using Center", leftColIdx, rightColIdx)
				return jr.sym.Center() // "┼" when merge ends before non-merged column (2→3)
			}
		}
		if isSpannedNext {
			return jr.sym.BottomMid()
		}
		if isSpannedCurrent {
			return jr.sym.TopMid()
		}
		return jr.sym.Center()
	}

	if vPassThruLStrict && vPassThruRStrict {
		return jr.sym.Column()
	}
	if vPassThruLStrict {
		return jr.sym.MidLeft()
	}
	if vPassThruRStrict {
		return jr.sym.MidRight()
	}
	if isSpannedCurrent && isSpannedNext {
		return jr.sym.Row()
	}
	if isSpannedCurrent {
		return jr.sym.TopMid()
	}
	if isSpannedNext {
		return jr.sym.BottomMid()
	}

	return jr.sym.Center()
}

// Line renders a full horizontal row line with junctions and segments.
func (f *Default) Line(w io.Writer, ctx Formatting) {
	jr := NewJunctionRenderer(f.config.Symbols, ctx, 0, f.debug)
	var line strings.Builder
	sortedKeys := twfn.ConvertToSortedKeys(ctx.Row.Widths)
	numCols := 0
	if len(sortedKeys) > 0 {
		numCols = sortedKeys[len(sortedKeys)-1] + 1
	}

	if numCols == 0 {
		prefix := ""
		suffix := ""
		if f.config.Borders.Left.Enabled() {
			prefix = jr.RenderLeft()
		}
		if f.config.Borders.Right.Enabled() {
			suffix = jr.RenderRight(-1)
		}
		if prefix != "" || suffix != "" {
			line.WriteString(prefix + suffix + tw.NewLine)
			fmt.Fprint(w, line.String())
		}
		f.debug("Line: Handled empty row/widths case")
		return
	}

	if f.config.Borders.Left.Enabled() {
		line.WriteString(jr.RenderLeft())
	}

	totalWidth := 0
	for i, colIdx := range sortedKeys {
		totalWidth += ctx.Row.Widths[colIdx]
		if i > 0 && f.config.Settings.Separators.BetweenColumns.Enabled() {
			totalWidth += twfn.DisplayWidth(jr.sym.Column())
		}
	}

	f.debug("Line: sortedKeys=%v, Widths=%v", sortedKeys, ctx.Row.Widths)
	for keyIndex, currentColIdx := range sortedKeys {
		jr.colIdx = currentColIdx
		segment := jr.GetSegment()
		colWidth := ctx.Row.Widths[currentColIdx]
		f.debug("Line: colIdx=%d, segment='%s', width=%d", currentColIdx, segment, colWidth)
		if segment == "" {
			line.WriteString(strings.Repeat(" ", colWidth))
		} else {
			repeat := colWidth / twfn.DisplayWidth(segment)
			if repeat < 1 && colWidth > 0 {
				repeat = 1
			}
			line.WriteString(strings.Repeat(segment, repeat))
		}

		isLast := keyIndex == len(sortedKeys)-1
		if !isLast && f.config.Settings.Separators.BetweenColumns.Enabled() {
			nextColIdx := sortedKeys[keyIndex+1]
			junction := jr.RenderJunction(currentColIdx, nextColIdx)
			f.debug("Line: Junction between %d and %d: '%s'", currentColIdx, nextColIdx, junction)
			line.WriteString(junction)
		}
	}

	if f.config.Borders.Right.Enabled() {
		lastIdx := sortedKeys[len(sortedKeys)-1]
		line.WriteString(jr.RenderRight(lastIdx))
		actualWidth := twfn.DisplayWidth(line.String()) - twfn.DisplayWidth(jr.RenderLeft()) - twfn.DisplayWidth(jr.RenderRight(lastIdx))
		if actualWidth > totalWidth {
			lineStr := line.String()
			line.Reset()
			excess := actualWidth - totalWidth
			line.WriteString(lineStr[:len(lineStr)-excess-twfn.DisplayWidth(jr.RenderRight(lastIdx))] + jr.RenderRight(lastIdx))
		}
	}

	line.WriteString(tw.NewLine)
	fmt.Fprint(w, line.String())
	f.debug("Line rendered: %s", strings.TrimSuffix(line.String(), tw.NewLine))
}
