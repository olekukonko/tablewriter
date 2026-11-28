---
title: Configuration
description: The Config struct, ConfigBuilder, and the With* options that control every table.
---

# Configuration

TableWriter v1 is configuration-driven: you describe the table you want up
front, and the renderer follows it. There are four ways to configure a
table, from the most explicit to the most convenient. They can be combined —
options passed to `NewTable` apply first, and `Configure` can adjust the
result afterwards.

## The Config struct

`Config` is the root configuration. It has one field per concern:

```go
type Config struct {
	MaxWidth int
	Header   tw.CellConfig
	Row      tw.CellConfig
	Footer   tw.CellConfig
	Debug    bool
	Stream   tw.StreamConfig
	Behavior tw.Behavior
	Widths   tw.CellWidth
	Counter  tw.Counter
}
```

- `MaxWidth` caps the total table width in characters.
- `Header`, `Row`, and `Footer` each hold a `tw.CellConfig` — the same
  template applied to one section of the table.
- `Debug` turns on debug logging; the buffered log is available from
  `Table.Debug()`.
- `Stream` holds a `tw.StreamConfig` for streaming mode.
- `Behavior` holds table-wide behavior: trimming, auto-hiding empty
  columns, header/footer visibility, compact merging, and struct handling.
- `Widths` holds a `tw.CellWidth` for fixed column widths.
- `Counter` holds a `tw.Counter` for line counting.

Pass a finished `Config` with `WithConfig`:

```go
table := tablewriter.NewTable(os.Stdout,
	tablewriter.WithConfig(tablewriter.Config{
		Header: tw.CellConfig{
			Formatting: tw.CellFormatting{AutoFormat: tw.On},
			Alignment:  tw.CellAlignment{Global: tw.AlignCenter},
		},
		Row: tw.CellConfig{
			Alignment: tw.CellAlignment{Global: tw.AlignLeft},
		},
		MaxWidth: 80,
		Behavior: tw.Behavior{TrimSpace: tw.On},
	}),
)
```

## CellConfig: one template per section

`tw.CellConfig` standardizes configuration across the header, row, and
footer sections. Each section gets the same set of knobs:

```go
type CellConfig struct {
	Formatting   CellFormatting // wrapping, auto-format
	Padding      CellPadding    // per-side and per-column padding
	Callbacks    CellCallbacks  // cell callbacks (placeholders today)
	Filter       CellFilter     // content filters
	Alignment    CellAlignment  // global and per-column alignment
	ColMaxWidths CellWidth      // per-column maximum widths
	Merging      CellMerging    // merge mode and column scope
}
```

### Formatting

`tw.CellFormatting` controls how cell text is shaped before rendering:

- `AutoWrap` — an `int` using the `tw.Wrap*` constants: `tw.WrapNormal`
  (word wrap), `tw.WrapTruncate` (cut with an ellipsis),
  `tw.WrapBreak` (break words to fit), or `tw.WrapNone`.
- `AutoFormat` — a `tw.State`: `tw.On` title-cases header text,
  `tw.Off` leaves it alone.

```go
Row: tw.CellConfig{
	Formatting: tw.CellFormatting{
		AutoWrap: tw.WrapNormal,
	},
}
```

### Alignment

`tw.CellAlignment` sets alignment globally, per column, or both. Per-column
entries override the global value. `tw.Skip` leaves a column on the
section default:

```go
Row: tw.CellConfig{
	Alignment: tw.CellAlignment{
		Global:    tw.AlignLeft,
		PerColumn: []tw.Align{tw.Skip, tw.AlignRight, tw.AlignCenter},
	},
}
```

The `tw.Align` values are `tw.AlignLeft`, `tw.AlignCenter`,
`tw.AlignRight`, and `tw.AlignNone`. `tw.AlignDefault` is an alias for
`tw.AlignLeft`.

### Padding

`tw.CellPadding` sets the characters around cell content, globally or per
column. `tw.Padding` has `Left`, `Right`, `Top`, and `Bottom` fields plus an
`Overwrite` flag:

```go
Footer: tw.CellConfig{
	Padding: tw.CellPadding{
		Global:    tw.Padding{Left: "*", Right: "*"},
		PerColumn: []tw.Padding{{}, {}, {Bottom: "^"}, {Bottom: "^"}},
	},
}
```

Ready-made presets exist in the `tw` package: `tw.PaddingNone`,
`tw.PaddingDefault`, `tw.BorderNone`, `tw.LinesNone`, and
`tw.SeparatorsNone`.

### Merging

`tw.CellMerging` has a `Mode` (`tw.MergeNone`, `tw.MergeVertical`,
`tw.MergeHorizontal`, `tw.MergeBoth`, `tw.MergeHierarchical`) and an
optional `ByColumnIndex` mapper that restricts merging to specific columns:

```go
Row: tw.CellConfig{
	Merging: tw.CellMerging{
		Mode:          tw.MergeHierarchical,
		ByColumnIndex: tw.NewMapper[int, bool](),
	},
}
```

See [Width & Truncation](width-truncation) for merging in depth.

### Widths

`tw.CellWidth` constrains column widths:

```go
type CellWidth struct {
	Global    int
	PerColumn Mapper[int, int]
}
```

Set it per section through `ColMaxWidths`, or table-wide through
`Config.Widths`. `CellWidth.Constrained()` reports whether any limit is
set. See [Width & Truncation](width-truncation).

### Filters

`tw.CellFilter` transforms cell content at render time. `Global` receives
the whole row; `PerColumn` holds one function per column:

```go
Row: tw.CellConfig{
	Filter: tw.CellFilter{
		PerColumn: []func(string) string{
			strings.ToUpper,
			nil,
			func(s string) string { return "$" + s },
		},
	},
}
```

`tw.CellCallbacks` exists in the struct but is a placeholder today — its
`Global` and `PerColumn` callbacks are not fully implemented yet, so do not
rely on them.

## Behavior

`tw.Behavior` controls table-wide behavior:

```go
type Behavior struct {
	AutoHide  State // hide columns that are entirely empty
	TrimSpace State // trim leading and trailing spaces
	TrimLine  State // collapse empty visual lines inside cells
	TrimTab   State // trim leading and trailing tabs
	Header    Control // header visibility
	Footer    Control // footer visibility
	Compact   Compact // compact width calculation for merged cells
	Structs   Struct  // how struct data is processed
}
```

`tw.Control` has a single field, `Hide` — a `tw.State` that hides the
header or footer section. `tw.Compact` has a `Merge` state that enables
compact width calculation for merged cells. `tw.Struct` has `AutoHeader`
(whether headers are extracted from struct fields when `Bulk` receives a
slice of structs — on by default) and `Tags`, the priority-ordered list of
struct tag keys checked for header names, defaulting to `["json", "db"]`.

States use the `tw.State` type: `tw.On` enables a feature, `tw.Off`
disables it, and `tw.Unknown` leaves the default in place. `State` has
`Enabled()`, `Disabled()`, `Default()`, `Toggle()`, `Cond()`, and `Or()`
helpers.

## The With* options

Functional options are the quickest way to configure a table at creation.
They fall into groups:

**Sections** — one option per setting, per section. The header, row, and
footer variants share the same suffixes:

| Option | What it sets |
|---|---|
| `WithHeader(headers []string)` / `WithFooter(footers []string)` | Section content |
| `WithHeaderAlignment` / `WithRowAlignment` / `WithFooterAlignment` | `tw.Align` for the section |
| `WithHeaderAlignmentConfig` / `WithRowAlignmentConfig` / `WithFooterAlignmentConfig` | Full `tw.CellAlignment` |
| `WithHeaderAutoFormat` / `WithRowAutoFormat` / `WithFooterAutoFormat` | `tw.State` for title-casing |
| `WithHeaderAutoWrap` / `WithRowAutoWrap` / `WithFooterAutoWrap` | Wrap mode (`tw.Wrap*`) |
| `WithHeaderConfig` / `WithRowConfig` / `WithFooterConfig` | Whole `tw.CellConfig` |
| `WithHeaderControl` / `WithFooterControl` | `tw.Control` (hide the section) |
| `WithHeaderFilter` / `WithRowFilter` / `WithFooterFilter` | `tw.CellFilter` |
| `WithHeaderCallbacks` / `WithRowCallbacks` / `WithFooterCallbacks` | `tw.CellCallbacks` |
| `WithHeaderMaxWidth` / `WithRowMaxWidth` / `WithFooterMaxWidth` | Section max width |
| `WithHeaderMergeMode` / `WithRowMergeMode` / `WithFooterMergeMode` | Merge mode |
| `WithHeaderPaddingPerColumn` / `WithRowPaddingPerColumn` / `WithFooterPaddingPerColumn` | `[]tw.Padding` |

**Table-wide:**

| Option | What it sets |
|---|---|
| `WithConfig(cfg Config)` | Whole `Config` at once |
| `WithRenderer(f tw.Renderer)` | The renderer |
| `WithRendererSettings(settings tw.Settings)` | Separator and line visibility |
| `WithRendition(rendition tw.Rendition)` | Borders, symbols, settings |
| `WithBorders(borders tw.Border)` | Border visibility (kept for compatibility; prefer `WithRendition`) |
| `WithSymbols(symbols tw.Symbols)` | Border symbols (kept for compatibility; prefer `WithRendition`) |
| `WithAlignment(alignment tw.Alignment)` | Alignment builder value |
| `WithBehavior(behavior tw.Behavior)` | Whole `Behavior` |
| `WithAutoHide(state tw.State)` | Hide empty columns |
| `WithTrimSpace` / `WithTrimLine` / `WithTrimTab` | Trimming states |
| `WithEastAsian(state tw.State)` | East Asian wide-character handling |
| `WithCondition(cond *runewidth.Condition)` | Custom width condition |
| `WithDebug(debug bool)` | Debug logging |
| `WithLogger(logger *ll.Logger)` | Custom logger |
| `WithMaxWidth(width int)` | Total table width cap |
| `WithTableMax(width int)` | Same cap, table-level |
| `WithColumnMax(width int)` | Per-column max width |
| `WithColumnWidths(widths tw.Mapper[int, int])` | Fixed column widths |
| `WithWidths(width tw.CellWidth)` | `CellWidth` for the table |
| `WithPadding(padding tw.Padding)` | Global cell padding |
| `WithCounters(counters ...tw.Counter)` | Line counters |
| `WithLineCounter()` | Attach a line counter |
| `WithStreaming(c tw.StreamConfig)` | Enable streaming mode |
| `WithStringer(stringer interface{})` | Custom struct-to-row function |
| `WithStringerCache()` | Cache stringer results |
| `WithStringerCacheCustom(cache twcache.Cache[reflect.Type, reflect.Value])` | Custom stringer cache |

## Configure after creation

`Table.Configure` adjusts the live config with a function. It is handy for
tweaks that depend on data you only learn after creating the table:

```go
table, err := tablewriter.NewCSV(os.Stdout, "test.csv", true)
if err != nil {
	log.Fatalf("Error: %v", err)
}

table.Configure(func(config *tablewriter.Config) {
	config.Row.Alignment.Global = tw.AlignLeft
})
table.Render()
```

`Table.Config()` returns a copy of the current config for inspection, and
`Table.Options(opts ...Option)` applies more options to an existing table.

## ConfigBuilder: the fluent way

`NewConfigBuilder` builds a `Config` through chained builders — one per
section, plus behavior and per-column builders. Each section builder ends
with `Build()`, which returns the parent builder, so you can keep chaining
until the final `Build()` returns the `Config`:

```go
b := tablewriter.NewConfigBuilder()
b.Row().Merging().
	WithMode(tw.MergeVertical).
	ByColumnIndex([]int{0, 3}).
	Build()

table := tablewriter.NewTable(&buf,
	tablewriter.WithConfig(b.Build()),
	tablewriter.WithRenderer(renderer.NewBlueprint(tw.Rendition{
		Settings: tw.Settings{
			Separators: tw.Separators{BetweenRows: tw.On},
		},
	})),
)
```

The section builders are `Header()`, `Row()`, and `Footer()`, each exposing
`Alignment()`, `Callbacks()`, `Filter()`, `Formatting()`, `Merging()`, and
`Padding()` sub-builders. The alignment builder offers `WithGlobal` and
`WithPerColumn`; the formatting builder offers `WithAutoFormat`,
`WithAutoWrap`, and `WithMergeMode`; the merging builder offers `WithMode`
and `ByColumnIndex`; the padding builder offers `WithGlobal`,
`WithPerColumn`, and `AddColumnPadding`; the callbacks builder offers
`WithGlobal`, `WithPerColumn`, and `AddColumnCallback`; the filter builder
offers `WithGlobal`, `WithPerColumn`, and `AddColumnFilter`.

`Behavior()` returns a behavior builder with `WithAutoHeader`,
`WithAutoHide`, `WithCompactMerge`, `WithFooterHide`, `WithHeaderHide`,
`WithTrimSpace`, and `WithTrimTab`. `ForColumn(col)` returns a column
builder with `WithAlignment` and `WithMaxWidth`. The root builder also has
direct `With*` methods for common settings: `WithMaxWidth`,
`WithHeaderAlignment`, `WithRowAlignment`, `WithFooterAlignment`,
`WithHeaderAutoFormat`, `WithRowAutoFormat`, `WithFooterAutoFormat`,
`WithHeaderAutoWrap`, `WithRowAutoWrap`, `WithFooterAutoWrap`,
`WithHeaderGlobalPadding`, `WithRowGlobalPadding`,
`WithFooterGlobalPadding`, `WithHeaderMaxWidth`, `WithRowMaxWidth`,
`WithFooterMaxWidth`, `WithHeaderMergeMode`, `WithRowMergeMode`,
`WithFooterMergeMode`, `WithAutoHide`, `WithDebug`, `WithTrimSpace`, and
`WithTrimTab`.

Use `ConfigBuilder` when a table has several interacting settings — the
chain reads like the table it produces. Use a `Config` literal when the
configuration is fixed and shared, and `With*` options for one-off tweaks.
