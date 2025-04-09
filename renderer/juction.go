package renderer

import (
	"fmt"
	"github.com/olekukonko/tablewriter/tw"
	"github.com/olekukonko/tablewriter/twfn"
	"io"
	"strings"
)

type JunctionRenderer struct {
	symbols   tw.Symbols
	ctx       Formatting
	colIdx    int
	debugging bool
	debug     func(format string, a ...interface{}) // Debug callback
}

// NewJunctionRenderer creates a new JunctionRenderer with a debug callback
func NewJunctionRenderer(symbols tw.Symbols, ctx Formatting, colIdx int, debug func(format string, a ...interface{})) *JunctionRenderer {
	return &JunctionRenderer{
		symbols:   symbols,
		ctx:       ctx,
		colIdx:    colIdx,
		debug:     debug,
		debugging: true,
	}
}

func (jr *JunctionRenderer) log(format string, a ...interface{}) {
	if jr.debugging {
		jr.debug("[JUNCTION] "+format, a...)
	}
}

func (jr *JunctionRenderer) DetermineJunction(nextColIdx int) string {
	current := jr.getMergeState(0, 0)
	next := jr.getMergeState(0, nextColIdx-jr.colIdx)
	prev := jr.getMergeState(0, -1)
	above := jr.getMergeState(-1, 0)
	below := jr.getMergeState(1, 0)

	jr.log("DetermineJunction: colIdx=%d, nextColIdx=%d, current=%v (span=%d), next=%v, prev=%v, above=%v, below=%v",
		jr.colIdx, nextColIdx, current, current.Span, next, prev, above, below)

	// Handle horizontal merges
	if current.Horizontal {
		if current.Start && nextColIdx <= jr.colIdx+current.Span {
			jr.log("Horizontal merge start at col %d, span %d, returning Row", jr.colIdx, current.Span)
			return jr.symbols.Row()
		}
		if current.End && jr.colIdx == nextColIdx-1 {
			jr.log("Horizontal merge end at col %d, returning MidRight", jr.colIdx)
			return jr.symbols.MidRight()
		}
		if !current.Start && !current.End && nextColIdx <= jr.colIdx+current.Span {
			jr.log("Horizontal merge middle at col %d, returning Row", jr.colIdx)
			return jr.symbols.Row()
		}
	}

	hasLeft := prev.Horizontal || jr.getMergeState(0, -1).Horizontal
	hasRight := next.Horizontal || jr.getMergeState(0, nextColIdx-jr.colIdx).Horizontal
	hasAbove := above.Vertical || (current.Vertical && current.Start)
	hasBelow := below.Vertical || (current.Vertical && !current.Start)

	jr.log("Flags: hasLeft=%v, hasRight=%v, hasAbove=%v, hasBelow=%v", hasLeft, hasRight, hasAbove, hasBelow)

	// Handle transitions for horizontal merges in adjacent rows
	if !current.Horizontal && next.Horizontal && next.Start {
		jr.log("Next row has horizontal merge start at col %d, returning BottomMid", nextColIdx)
		return jr.symbols.BottomMid()
	}
	if !current.Horizontal && prev.Horizontal && prev.End {
		jr.log("Previous row has horizontal merge end at col %d, returning TopMid", jr.colIdx-1)
		return jr.symbols.TopMid()
	}

	switch {
	case jr.ctx.Level == tw.LevelHeader && jr.ctx.Row.Location == tw.LocationFirst:
		jr.log("Header first row, rendering top border junction")
		return jr.renderTopBorderJunction(hasRight, hasLeft)
	case jr.ctx.Level == tw.LevelFooter && jr.ctx.Row.Location == tw.LocationEnd:
		jr.log("Footer last row, rendering bottom border junction")
		return jr.renderBottomBorderJunction(hasRight, hasLeft)
	case jr.ctx.Row.Location == tw.LocationFirst && jr.ctx.Level == tw.LevelBody:
		jr.log("Body first row, rendering header separator junction")
		return jr.renderHeaderSeparatorJunction(hasAbove, hasBelow, hasRight, hasLeft)
	case jr.ctx.Row.Location == tw.LocationEnd && jr.ctx.Level == tw.LevelBody:
		jr.log("Body last row, rendering bottom separator junction")
		return jr.renderBottomSeparatorJunction(hasAbove, hasRight, hasLeft)
	default:
		jr.log("Default case, rendering standard junction")
		return jr.renderStandardJunction(hasAbove, hasBelow, hasRight, hasLeft)
	}
}

func (jr *JunctionRenderer) getMergeState(rowOffset, colOffset int) MergeState {
	targetCol := jr.colIdx + colOffset
	if targetCol < 0 || targetCol >= len(jr.ctx.Row.Widths) {
		jr.log("getMergeState: targetCol %d out of bounds, returning empty MergeState", targetCol)
		return MergeState{}
	}

	var cellCtx CellContext
	var ok bool
	if rowOffset < 0 {
		if jr.ctx.Row.Previous != nil && targetCol < len(jr.ctx.Row.Previous) {
			cellCtx, ok = jr.ctx.Row.Previous[targetCol]
		}
	} else if rowOffset > 0 {
		if jr.ctx.Row.Next != nil && targetCol < len(jr.ctx.Row.Next) {
			cellCtx, ok = jr.ctx.Row.Next[targetCol]
		}
	} else {
		if targetCol < len(jr.ctx.Row.Current) {
			cellCtx, ok = jr.ctx.Row.Current[targetCol]
		}
	}

	if ok {
		jr.log("getMergeState: rowOffset=%d, colOffset=%d, found merge state %v", rowOffset, colOffset, cellCtx.Merge)
		return cellCtx.Merge
	}
	jr.log("getMergeState: rowOffset=%d, colOffset=%d, no cell found, returning empty MergeState", rowOffset, colOffset)
	return MergeState{}
}

func (jr *JunctionRenderer) renderTopBorderJunction(right, left bool) string {
	switch {
	case right && left:
		return jr.symbols.TopMid()
	case right:
		return jr.symbols.TopLeft()
	case left:
		return jr.symbols.TopRight()
	default:
		return jr.symbols.TopMid()
	}
}

func (jr *JunctionRenderer) renderBottomBorderJunction(right, left bool) string {
	switch {
	case right && left:
		return jr.symbols.BottomMid()
	case right:
		return jr.symbols.BottomLeft()
	case left:
		return jr.symbols.BottomRight()
	default:
		return jr.symbols.BottomMid()
	}
}

func (jr *JunctionRenderer) renderHeaderSeparatorJunction(above, below, right, left bool) string {
	switch {
	case above && below && right && left:
		return jr.symbols.Center()
	case above && below && right:
		return jr.symbols.MidLeft()
	case above && below && left:
		return jr.symbols.MidRight()
	case right && left:
		return jr.symbols.Center()
	case below && right:
		return jr.symbols.MidLeft()
	case below && left:
		return jr.symbols.MidRight()
	default:
		return jr.symbols.Center()
	}
}

func (jr *JunctionRenderer) renderBottomSeparatorJunction(above, right, left bool) string {
	current := jr.getMergeState(0, 0)
	if current.Horizontal {
		if current.Start {
			return jr.symbols.MidLeft()
		}
		if current.End {
			return jr.symbols.MidRight()
		}
		return jr.symbols.Row()
	}

	switch {
	case above && right && left:
		return jr.symbols.Row()
	case right && left:
		return jr.symbols.BottomMid()
	case above && right:
		return jr.symbols.MidLeft()
	case above && left:
		return jr.symbols.MidRight()
	default:
		return jr.symbols.BottomMid()
	}
}

func (jr *JunctionRenderer) renderStandardJunction(above, below, right, left bool) string {
	switch {
	case above && below && right && left:
		return jr.symbols.Center()
	case above && below && right:
		return jr.symbols.MidLeft()
	case above && below && left:
		return jr.symbols.MidRight()
	case above && right && left:
		return jr.symbols.BottomMid()
	case below && right && left:
		return jr.symbols.TopMid()
	case above && right:
		return jr.symbols.BottomLeft()
	case above && left:
		return jr.symbols.BottomRight()
	case below && right:
		return jr.symbols.TopLeft()
	case below && left:
		return jr.symbols.TopRight()
	case above && below:
		return jr.symbols.Column()
	case right && left:
		return jr.symbols.Row()
	default:
		return jr.symbols.Center()
	}
}

func (jr *JunctionRenderer) renderLeftBorder() string {
	current := jr.getMergeState(0, 0)
	above := jr.getMergeState(-1, 0)
	below := jr.getMergeState(1, 0)
	hasAbove := above.Vertical || (current.Vertical && current.Start)
	hasBelow := below.Vertical || (current.Vertical && !current.Start)

	jr.log("renderLeftBorder: current=%v, hasAbove=%v, hasBelow=%v, level=%d, location=%v",
		current, hasAbove, hasBelow, jr.ctx.Level, jr.ctx.Row.Location)

	// Footer last row should always use BottomLeft unless overridden by a full-span horizontal merge
	if jr.ctx.Level == tw.LevelFooter && jr.ctx.Row.Location == tw.LocationEnd {
		jr.log("Left border: footer last, returning BottomLeft")
		return jr.symbols.BottomLeft()
	}

	// Vertical merge continuation should use MidLeft
	if current.Vertical && !current.Start {
		jr.log("Left border: vertical merge continuation, returning MidLeft")
		return jr.symbols.MidLeft() // Use ├ for continuation to maintain structure
	}

	switch {
	case jr.ctx.Level == tw.LevelHeader && jr.ctx.Row.Location == tw.LocationFirst:
		jr.log("Left border: header first, returning TopLeft")
		return jr.symbols.TopLeft()
	case jr.ctx.Row.Location == tw.LocationFirst && jr.ctx.Level == tw.LevelBody:
		jr.log("Left border: body first, returning MidLeft")
		return jr.symbols.MidLeft()
	case hasAbove && hasBelow:
		jr.log("Left border: hasAbove && hasBelow, returning Column")
		return jr.symbols.Column()
	default:
		jr.log("Left border: default, returning MidLeft")
		return jr.symbols.MidLeft()
	}
}

func (jr *JunctionRenderer) renderRightBorder() string {
	current := jr.getMergeState(0, 0)
	above := jr.getMergeState(-1, 0)
	below := jr.getMergeState(1, 0)
	hasAbove := above.Vertical || (current.Vertical && current.Start)
	hasBelow := below.Vertical || (current.Vertical && !current.Start)

	jr.log("renderRightBorder: current=%v, hasAbove=%v, hasBelow=%v, level=%d, location=%v",
		current, hasAbove, hasBelow, jr.ctx.Level, jr.ctx.Row.Location)

	if current.Horizontal && current.End && current.Span == len(jr.ctx.Row.Widths) {
		jr.log("Right border: full-span horizontal merge end, returning BottomRight")
		return jr.symbols.BottomRight()
	}

	switch {
	case jr.ctx.Level == tw.LevelHeader && jr.ctx.Row.Location == tw.LocationFirst:
		return jr.symbols.TopRight()
	case jr.ctx.Level == tw.LevelFooter && jr.ctx.Row.Location == tw.LocationEnd:
		return jr.symbols.BottomRight()
	case jr.ctx.Row.Location == tw.LocationFirst && jr.ctx.Level == tw.LevelBody:
		return jr.symbols.MidRight()
	case hasAbove && hasBelow:
		return jr.symbols.Column()
	default:
		return jr.symbols.MidRight()
	}
}

func (jr *JunctionRenderer) getSegment() string {
	current := jr.getMergeState(0, 0)
	if current.Horizontal {
		jr.log("getSegment: horizontal merge, returning Row")
		return jr.symbols.Row()
	}
	jr.log("getSegment: default, returning Row")
	return jr.symbols.Row()
}

func (f *Default) Line(w io.Writer, ctx Formatting) {
	f.debug("Starting Line render: level=%d, pos=%s, loc=%v, widths=%v",
		ctx.Level, ctx.Row.Position, ctx.Row.Location, ctx.Row.Widths)

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

	if f.config.Borders.Left.Enabled() {
		jr := NewJunctionRenderer(f.config.Symbols, ctx, sortedKeys[0], f.debug)
		line.WriteString(jr.renderLeftBorder())
	}

	fullSpanMerge := false
	var mergeStartIdx, mergeEndIdx int
	if len(ctx.Row.Current) > 0 {
		for colIdx, cell := range ctx.Row.Current {
			if cell.Merge.Horizontal && cell.Merge.Start && cell.Merge.Span > 1 {
				fullSpanMerge = cell.Merge.Span == len(sortedKeys)
				mergeStartIdx = colIdx
				mergeEndIdx = colIdx + cell.Merge.Span - 1
				f.debug("Detected horizontal merge: start=%d, end=%d, span=%d, fullSpan=%v", mergeStartIdx, mergeEndIdx, cell.Merge.Span, fullSpanMerge)
				break
			}
		}
	}

	for i, colIdx := range sortedKeys {
		jr := NewJunctionRenderer(f.config.Symbols, ctx, colIdx, f.debug)
		current := jr.getMergeState(0, 0)

		if width, ok := ctx.Row.Widths[colIdx]; ok && width > 0 {
			line.WriteString(strings.Repeat(jr.getSegment(), width))
		}

		// Add junction between columns unless within a horizontal merge span and not at the end
		if i < len(sortedKeys)-1 && f.config.Settings.Separators.BetweenColumns.Enabled() {
			if fullSpanMerge && ctx.Row.Location == tw.LocationEnd {
				// Skip all junctions for a full-span merge on the last row
				continue
			}
			if current.Horizontal && colIdx >= mergeStartIdx && colIdx < mergeEndIdx {
				// Skip junction within a horizontal merge span, except at the end
				continue
			}
			nextColIdx := sortedKeys[i+1]
			line.WriteString(jr.DetermineJunction(nextColIdx))
		}
	}

	if f.config.Borders.Right.Enabled() {
		jr := NewJunctionRenderer(f.config.Symbols, ctx, lastColIdx, f.debug)
		if fullSpanMerge && ctx.Row.Location == tw.LocationEnd {
			f.debug("Full-span merge at end, using BottomRight")
			line.WriteString(jr.symbols.BottomRight())
		} else {
			line.WriteString(jr.renderRightBorder())
		}
	}

	fmt.Fprintln(w, line.String())
	f.debug("Line render completed: [%s]", line.String())
}
