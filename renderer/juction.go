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
	switch jr.ctx.Level {
	case tw.LevelHeader:
		if jr.ctx.Row.Location == tw.LocationFirst {
			return jr.sym.TopRight() // ┐
		}
		return jr.sym.MidRight() // ┤
	case tw.LevelBody:
		if jr.ctx.Row.Location == tw.LocationEnd {
			return jr.sym.BottomRight() // ┘
		}
		return jr.sym.MidRight() // ┤
	case tw.LevelFooter:
		if jr.ctx.Row.Location == tw.LocationEnd {
			return jr.sym.BottomRight() // ┘
		}
		return jr.sym.MidRight() // ┤
	}
	return jr.sym.MidRight() // Default: ┤
}

// RenderJunction returns the junction symbol between two columns
func (jr *JunctionRenderer) RenderJunction(nextColIdx int) string {
	jr.log("Rendering junction: colIdx=%d, nextColIdx=%d, level=%d, location=%v", jr.colIdx, nextColIdx, jr.ctx.Level, jr.ctx.Row.Location)
	switch jr.ctx.Level {
	case tw.LevelHeader:
		if jr.ctx.Row.Location == tw.LocationFirst {
			return jr.sym.TopMid() // ┬ for top border
		}
		return jr.sym.Center() // ┼ for header separator
	case tw.LevelBody:
		if jr.ctx.Row.Location == tw.LocationFirst {
			return jr.sym.Center() // ┼ for first row separator
		}
		if jr.ctx.Row.Location == tw.LocationEnd {
			return jr.sym.BottomMid() // ┴ for last row separator
		}
		return jr.sym.Center() // ┼ for middle rows
	case tw.LevelFooter:
		if jr.ctx.Row.Location == tw.LocationFirst {
			return jr.sym.Center() // ┼ for footer separator
		}
		return jr.sym.BottomMid() // ┴ for bottom border
	}
	return jr.sym.Center() // Default: ┼
}

// GetSegment returns the horizontal segment character
func (jr *JunctionRenderer) GetSegment() string {
	jr.log("Getting segment for colIdx=%d", jr.colIdx)
	return jr.sym.Row() // ─ for basic rendering
}

// Line renders a horizontal line with junctions
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

	// Left border
	if f.config.Borders.Left.Enabled() {
		jr := NewJunctionRenderer(f.config.Symbols, ctx, sortedKeys[0], f.debug)
		line.WriteString(jr.RenderLeft())
	}

	// Columns and junctions
	for i, colIdx := range sortedKeys {
		jr := NewJunctionRenderer(f.config.Symbols, ctx, colIdx, f.debug)
		segment := jr.GetSegment()
		if width, ok := ctx.Row.Widths[colIdx]; ok && width > 0 {
			line.WriteString(strings.Repeat(segment, width))
		}
		if i < len(sortedKeys)-1 && f.config.Settings.Separators.BetweenColumns.Enabled() {
			nextColIdx := sortedKeys[i+1]
			line.WriteString(jr.RenderJunction(nextColIdx))
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
