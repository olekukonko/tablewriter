---
title: The tw Package
description: Shared types, helpers, and presets — alignment, state, symbols, mappers, and slicers.
---

# The tw Package

The `tw` package (`github.com/olekukonko/tablewriter/tw`) holds the types
and helpers shared by the table, its configuration, and its renderers.
You already use it whenever you write `tw.AlignLeft` or `tw.On`. This page
tours the whole package so you know what is available.

## State

`tw.State` is a three-valued on/off switch. `tw.On` enables a feature,
`tw.Off` disables it, and `tw.Unknown` leaves the default in place. It is
the type behind every behavior flag, separator toggle, and line toggle in
the library.

```go
Settings: tw.Settings{
	Separators: tw.Separators{BetweenRows: tw.On},
	Lines:      tw.Lines{ShowFooterLine: tw.On},
}
```

`State` has helpers: `Enabled()` and `Disabled()` test the value,
`Default()` reports whether it is still unknown, `Toggle()` flips on to
off and back, `Cond(fn)` runs a function only when enabled, and `Or(other)`
falls back to another state when disabled. `String()` renders `"on"`,
`"off"`, or `"undefined"`.

## Alignment

`tw.Align` is a string type with four values: `tw.AlignLeft`,
`tw.AlignCenter`, `tw.AlignRight`, and `tw.AlignNone`.
`tw.AlignDefault` is an alias for `tw.AlignLeft`. `tw.Skip` is the empty
alignment, used in per-column lists to leave a column on the section
default. `Align.Validate()` reports an error for an unknown value.

`tw.Alignment` is the computed alignment for a cell — the result of
combining global and per-column settings. It has `Add`, `Set`, `Copy`, and
`String` methods, and `tw.MakeAlign(l int, align Align) Alignment` builds
one for a column of a given width.

`tw.CellAlignment` is the configuration form: a `Global` align plus a
`PerColumn` slice of overrides. See [Configuration](configuration).

## Symbols and border styles

`tw.Symbols` is the interface a renderer uses to draw borders and
junctions: `Name`, `Center`, `Row`, `Column`, and the corner and junction
symbols `TopLeft`, `TopMid`, `TopRight`, `MidLeft`, `MidRight`,
`BottomLeft`, `BottomMid`, `BottomRight`, plus `HeaderLeft`, `HeaderMid`,
and `HeaderRight`.

`tw.NewSymbols(style tw.BorderStyle) Symbols` returns a built-in set. The
styles are:

`StyleNone`, `StyleASCII`, `StyleLight`, `StyleHeavy`, `StyleDouble`,
`StyleDoubleLong`, `StyleLightHeavy`, `StyleHeavyLight`,
`StyleLightDouble`, `StyleDoubleLight`, `StyleRounded`, `StyleMarkdown`,
`StyleGraphical`, `StyleMerger`, `StyleDotted`, `StyleArrow`,
`StyleStarry`, `StyleHearts`, `StyleCircuit`, `StyleNature`,
`StyleArtistic`, `Style8Bit`, `StyleChaos`, `StyleDots`, `StyleBlocks`,
`StyleZen`, `StyleVintage`, `StyleSketch`, `StyleArrowDouble`,
`StyleCelestial`, `StyleCyber`, `StyleRunic`, `StyleIndustrial`,
`StyleInk`, `StyleArcade`, `StyleBlossom`.

`StyleName` is the string form of the same list (`StyleNameASCII`,
`StyleNameLight`, and so on) for selecting a style by name.

For a fully custom look, `tw.NewSymbolCustom(name string) *SymbolCustom`
builds a set one symbol at a time with `WithRow`, `WithColumn`,
`WithTopLeft`, `WithTopMid`, `WithTopRight`, `WithMidLeft`, `WithCenter`,
`WithMidRight`, `WithBottomLeft`, `WithBottomMid`, and `WithBottomRight`,
plus `WithHeaderLeft`, `WithHeaderMid`, and `WithHeaderRight`. The matching
getters (`Row()`, `Column()`, `Center()`, and the rest) read the values
back, and `Name()` returns the set's name.

## Presets

A few ready-made values cover the common "turn it off" cases:

```go
tw.PaddingNone    // Padding{Left: "", Right: "", Top: "", Bottom: "", Overwrite: true}
tw.PaddingDefault // Padding{Left: " ", Right: " ", Overwrite: true}
tw.BorderNone     // Border{Left: Off, Right: Off, Top: Off, Bottom: Off}
tw.LinesNone      // Lines{ShowTop: Off, ShowBottom: Off, ShowHeaderLine: Off, ShowFooterLine: Off}
tw.SeparatorsNone // Separators{ShowHeader: Off, ShowFooter: Off, BetweenRows: Off, BetweenColumns: Off}
```

## Mapper and Slicer

`tw.Mapper[K, V]` is a generic ordered map used for per-column settings —
widths, merge scopes, and alignments all travel through it. Build one with
`tw.NewMapper[K, V]()`, `tw.NewMapperWithKeys[K, V](keys ...K)`, or
`tw.NewMapperFromPairs[K, V](pairs ...any)`. Convenience constructors
cover the common value types: `tw.NewIntMapper[K](keys ...K)`,
`tw.NewBoolMapper[K](keys ...K)`, and `tw.NewIdentityMapper[K](keys ...K)`.

```go
widths := tw.NewMapper[int, int]()
widths.Set(0, 20)
widths.Set(2, 12)

if w, ok := widths.Get(0); ok {
	fmt.Println("column 0 width:", w)
}
fmt.Println("columns:", widths.Keys())
fmt.Println("sorted:", widths.SortedKeys())
```

`Mapper` methods: `Set`, `Get`, `OK`, `Has`, `Delete`, `Clear`, `Len`,
`Keys`, `SortedKeys`, `Values`, `Each`, `Filter`, `MapValues`, `Clone`,
and `Slicer` (which converts the mapper to a `Slicer[KeyValuePair[K, V]]`).

`tw.Slicer[T]` is the matching generic slice type. `tw.NewSlicer[T]()`
creates one; `Append` and `Prepend` grow it, `Get` and `GetOK` read by
index, `First` and `Last` read the ends, `Len` and `IsEmpty` measure it,
`Has` tests an index, `Each`, `Filter`, `Map`, `Contains`, and `Find`
iterate and search it, and `Clone` copies it. `tw.SlicerToMapper` converts
a `Slicer[KeyValuePair[K, V]]` back into a `Mapper[K, V]`.

## Renderer types

The rendering surface lives in `tw` so custom renderers do not need to
import the `renderer` package:

- `tw.Renderer` — the interface every renderer implements (`Start`,
  `Header`, `Row`, `Footer`, `Line`, `Config`, `Close`, `Logger`).
- `tw.Rendition` — visual styling: `Borders`, `Symbols`, `Settings`, and
  `Streaming`.
- `tw.Renditioning` — the optional `Rendition(r Rendition)` interface for
  renderers that accept a new rendition after creation.
- `tw.Formatting` — the full formatting context handed to a renderer for a
  row: the `RowContext`, the `Level` (header, body, footer), whether the
  table has a footer, whether the row is a sub-row, and the normalized
  widths.
- `tw.RowContext` — layout state for a row: `Position`, `Location`,
  `Current`/`Previous`/`Next` cell maps, `Widths`, and `ColMaxWidths`.
  `GetCell(col)` reads one cell.
- `tw.CellContext` — one cell: `Data`, `Align`, `Padding`, `Width`, and
  `Merge`.
- `tw.MergeState` — how a cell spans: `Vertical`, `Horizontal`, and
  `Hierarchical`, each a `MergeStateOption` with `Present`, `Span`,
  `Start`, and `End`.
- `tw.Settings`, `tw.Separators`, `tw.Lines`, `tw.Border` — the visibility
  knobs described in [Renderers](renderers).
- `tw.Position`, `tw.Location`, `tw.Level` — where a row sits in the table,
  each with a `Validate()` method.
- `tw.Caption` — table captions, built with `WithText`, `WithAlign`,
  `WithSpot`, and `WithWidth`.

## Configuration types

The configuration surface also lives in `tw`:

- `tw.CellConfig`, `tw.CellFormatting`, `tw.CellPadding`,
  `tw.CellAlignment`, `tw.CellWidth`, `tw.CellMerging`, `tw.CellFilter`,
  `tw.CellCallbacks` — the per-section template. See
  [Configuration](configuration).
- `tw.Behavior` — table-wide behavior, including `Structs` (`tw.Struct`
  with `AutoHeader` and `Tags`).
- `tw.StreamConfig` — streaming mode: `Enable`, `StrictColumns`, and the
  deprecated `Widths`.
- `tw.Counter` and `tw.LineCounter` — line counting. `LineCounter` has
  `Total()` and `Write()`; attach one with `WithLineCounter()` and read it
  back with `Table.Counters()`.
- `tw.Control` — a single `Hide` state for a section. `tw.Compact` — a
  `Merge` state for compact width calculation on merged cells.

## Text helpers

Small functions used across the library, available for your own code too:

- `tw.Pad(s, padChar string, totalWidth int, alignment Align) string`,
  plus `tw.PadLeft`, `tw.PadRight`, and `tw.PadCenter`.
- `tw.Title(name string) string` — title-case a name.
- `tw.SplitCamelCase(src string) (entries []string)` — split a camel-case
  identifier into words.
- `tw.NormalizeNewLines(s string) string` — normalize line endings.
- `tw.IsNumeric(s string) bool` and `tw.IsIsNumericOrSpace(r rune) bool`.
- `tw.Max(a, b int) int` and `tw.Min(a, b int) int`.
- `tw.Or(cond bool, valid, inValid string) string` — pick one of two
  strings.
- `tw.BreakPoint(s string, limit int) int` — find a wrap point.
- `tw.Glyphs` — glyph tables used by the symbol styles.
