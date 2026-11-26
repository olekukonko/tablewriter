---
title: Width & Truncation
description: Column widths, wrapping, truncation, East Asian widths, and cell merging.
---

# Width & Truncation

Column widths decide whether a table is readable. TableWriter computes
them from content by default, lets you cap or fix them, wraps or truncates
text that does not fit, and measures widths correctly for wide characters,
tabs, and ANSI color codes. This page covers all of it, plus the cell
merging that interacts with widths.

## How widths are computed

By default, each column is as wide as its widest cell, plus padding.
Nothing else to do — the table sizes itself. When content is long, the
table grows; when you need it to stop growing, the knobs below take over.

`tw.CellWidth` is the shared width type:

```go
type CellWidth struct {
	Global    int
	PerColumn Mapper[int, int]
}
```

`CellWidth.Constrained()` reports whether any limit is set.

## Capping the table

`Config.MaxWidth` caps the total table width in characters. Set it with
`WithMaxWidth` or `WithTableMax`:

```go
table := tablewriter.NewTable(os.Stdout,
	tablewriter.WithMaxWidth(80),
)
```

When the cap binds, columns shrink and long cells wrap or truncate
according to their `AutoWrap` mode. `Config.Widths` (set with
`WithWidths`) fixes widths table-wide instead of capping the total:

```go
table := tablewriter.NewTable(os.Stdout,
	tablewriter.WithConfig(tablewriter.Config{
		Widths: tw.CellWidth{
			PerColumn: tw.NewMapper[int, int](),
		},
	}),
)
```

## Per-column widths

`WithColumnWidths` takes a `tw.Mapper[int, int]` of column index to width,
and `WithColumnMax` sets one maximum for every column:

```go
table := tablewriter.NewTable(os.Stdout,
	tablewriter.WithColumnWidths(tw.NewMapper[int, int]().Set(0, 20).Set(2, 12)),
	tablewriter.WithColumnMax(30),
)
```

Per-section caps live on `CellConfig` as `ColMaxWidths` — the header, row,
and footer sections can each have different limits:

```go
Row: tw.CellConfig{
	ColMaxWidths: tw.CellWidth{Global: 25},
}
```

The `ConfigBuilder` exposes the same settings as `WithMaxWidth`,
`WithHeaderMaxWidth`, `WithRowMaxWidth`, `WithFooterMaxWidth`, and
`ForColumn(col).WithMaxWidth(width)`.

## Wrapping and truncation

`CellFormatting.AutoWrap` decides what happens when text is wider than its
column. The modes are `tw.WrapNormal` (word wrap), `tw.WrapTruncate`
(cut with an ellipsis), `tw.WrapBreak` (break words to fit), and
`tw.WrapNone`:

```go
table := tablewriter.NewTable(os.Stdout,
	tablewriter.WithRenderer(renderer.NewColorized(colorCfg)),
	tablewriter.WithConfig(tablewriter.Config{
		Row: tw.CellConfig{
			Formatting:   tw.CellFormatting{AutoWrap: tw.WrapNormal},
			Alignment:    tw.CellAlignment{Global: tw.AlignLeft},
			ColMaxWidths: tw.CellWidth{Global: 25},
		},
	}),
)
```

With `WrapNormal`, a long description flows onto continuation lines inside
its cell. With `WrapTruncate`, it is cut and ends with `…`. The per-section
options `WithHeaderAutoWrap`, `WithRowAutoWrap`, and
`WithFooterAutoWrap` set the mode for one section, and the `tw` struct tag
`wrap` sets it for a single column.

In streaming mode, truncation is the norm: the Ocean renderer cuts long
cells with `…` because it cannot reflow rows it has already printed. See
[Rendering](rendering).

## Measuring width: the twwidth package

The `pkg/twwidth` package measures display width — the number of terminal
columns a string occupies — which is not the same as its byte or rune
length once wide characters, tabs, and ANSI escapes are involved:

- `twwidth.Width(str string) int` — cached display width.
- `twwidth.WidthNoCache(str string) int` — the same measurement without
  the cache.
- `twwidth.WidthWithOptions(str string, opts Options) int` — measured with
  explicit options.
- `twwidth.Display(cond *runewidth.Condition, str string) int` — width
  under a given `runewidth` condition.
- `twwidth.Truncate(s string, maxWidth int, suffix ...string) string` —
  cut a string to a display width, appending the suffix (typically `…`)
  inside the budget. Returns `""` for a negative width and the original
  string when it already fits.
- `twwidth.Filter() *regexp.Regexp` — the expression used to strip ANSI
  sequences before measuring.

Width results are cached; `twwidth.GetCacheStats()` reports the cache size,
capacity, and hit rate, and `twwidth.SetCacheCapacity(capacity int)` tunes
it.

### East Asian wide characters

CJK characters occupy two terminal columns. `twwidth` detects the
environment and adjusts:

- `twwidth.IsEastAsian() bool` and `twwidth.EastAsianMode() string` report
  the current detection.
- `twwidth.EastAsianDetect() bool` and
  `twwidth.EastAsianConservative() bool` report the detection details.
- `twwidth.SetEastAsian(enable bool)` forces wide-character handling on or
  off, and `twwidth.EastAsianForceLegacy(force bool)` forces the legacy
  behavior.
- `twwidth.SetForceNarrow(enable bool)` forces box-drawing characters to
  count as narrow regardless of the East Asian setting.
- `twwidth.SetOptions(opts Options)` sets both knobs at once through
  `twwidth.Options{EastAsianWidth, ForceNarrowBorders}`.
- `twwidth.Debugging() Detection` reports the detection state, and
  `twwidth.SetCondition(cond *runewidth.Condition)` installs a custom
  width condition (the same condition `WithCondition` passes to a table).

Enable East Asian handling on a table with `WithEastAsian(tw.On)`.

### Tabs

`twwidth.SetTabWidth(w int)` and `twwidth.TabWidth()` control how wide a
tab counts; `twwidth.IsTab(r rune) bool` tests a rune, and
`twwidth.TabInstance() *Tabinal` returns the shared tab handler.
`WithTrimTab(tw.On)` trims tabs from cell edges at the table level.

## Cell merging

Merging joins identical adjacent cells so repeated values print once,
spanning the rows or columns they cover. The mode lives on
`tw.CellMerging.Mode`:

- `tw.MergeNone` — no merging.
- `tw.MergeVertical` — merge identical values down a column.
- `tw.MergeHorizontal` — merge identical values across a row.
- `tw.MergeBoth` — both directions.
- `tw.MergeHierarchical` — tree-style merging for nested data: a value
  spans the rows of every distinct value beneath it.

`CellMerging.ByColumnIndex` (a `Mapper[int, bool]`) restricts merging to
specific columns; `ByRowIndex` is reserved for future use. The per-section
options are `WithHeaderMergeMode`, `WithRowMergeMode`, and
`WithFooterMergeMode`, and the builder form is `Row().Merging().WithMode(...).ByColumnIndex([]int{...})`.

Hierarchical merging turns flat rows into a visible tree:

```go
data := [][]string{
	{"Engineering", "Backend", "API Team", "Alice"},
	{"Engineering", "Backend", "Database Team", "Bob"},
	{"Engineering", "Frontend", "UI Team", "Charlie"},
	{"Marketing", "Digital", "SEO Team", "Dave"},
	{"Marketing", "Digital", "Content Team", "Eve"},
}

table := tablewriter.NewTable(os.Stdout,
	tablewriter.WithRenderer(renderer.NewBlueprint(tw.Rendition{
		Settings: tw.Settings{Separators: tw.Separators{BetweenRows: tw.On}},
	})),
	tablewriter.WithConfig(tablewriter.Config{
		Header: tw.CellConfig{Alignment: tw.CellAlignment{Global: tw.AlignCenter}},
		Row: tw.CellConfig{
			Merging:   tw.CellMerging{Mode: tw.MergeHierarchical},
			Alignment: tw.CellAlignment{Global: tw.AlignLeft},
		},
	}),
)
table.Header([]string{"Department", "Division", "Team", "Lead"})
table.Bulk(data)
table.Render()
```

```
┌─────────────┬──────────┬───────────────┬─────────┐
│ DEPARTMENT  │ DIVISION │     TEAM      │  LEAD   │
├─────────────┼──────────┼───────────────┼─────────┤
│ Engineering │ Backend  │ API Team      │ Alice   │
│             │          ├───────────────┼─────────┤
│             │          │ Database Team │ Bob     │
│             ├──────────┼───────────────┼─────────┤
│             │ Frontend │ UI Team       │ Charlie │
├─────────────┼──────────┼───────────────┼─────────┤
│ Marketing   │ Digital  │ SEO Team      │ Dave    │
│             │          ├───────────────┼─────────┤
│             │          │ Content Team  │ Eve     │
└─────────────┴──────────┴───────────────┴─────────┘
```

"Engineering" spans all of its teams; "Backend" spans its two teams.
Horizontal merging works the same way across a row — the HTML example in
[Renderers](renderers) uses `tw.MergeHorizontal` to collapse repeated
"Quarter" header cells with a `colspan`.

Merging and widths interact: a merged cell's width is computed from the
span it covers, and `Behavior.Compact` (`Compact.Merge`) enables compact
width calculation for merged cells so the table does not over-allocate
space to spans. If merged columns come out wider than expected, that is
the first setting to check.

## Padding and merging together

Padding is per cell, so merged cells keep their padding on the visible
edges. A footer total often wants different padding from the body — set it
per section:

```go
Footer: tw.CellConfig{
	Padding: tw.CellPadding{
		Global:    tw.Padding{Left: "*", Right: "*"},
		PerColumn: []tw.Padding{{}, {}, {Bottom: "^"}, {Bottom: "^"}},
	},
	Alignment: tw.CellAlignment{PerColumn: []tw.Align{tw.Skip, tw.Skip, tw.AlignRight, tw.AlignLeft}},
},
```

Combined with `tw.MergeBoth` on the rows, this produces merged date and
description cells with a padded, right-aligned total row — the shape used
for invoice-style tables.
