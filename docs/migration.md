---
title: Migration Guide
description: Moving from tablewriter v0.0.5 to v1 — what changed and how to update your code.
---

# Migration Guide: v0.0.5 to v1

TableWriter v1 is a redesign, not a patch. The v0.0.5 API was method-driven:
you created a table and called setters on it one at a time. The v1 API is
configuration-driven: you describe the table with a `Config`, pick a
`tw.Renderer`, and the renderer follows the description. The old methods
are still present in `deprecated.go` for compatibility, but new code should
use the v1 API throughout.

> The upstream migration guide notes it is a work in progress and to use it
> with caution. For anything this page does not cover, read the source or
> open an issue on the repository.

## Why migrate

- **Extensibility.** Rendering is decoupled from the table, so the same
  table renders as ASCII, Markdown, HTML, SVG, or colorized text — and you
  can implement `tw.Renderer` for formats of your own.
- **Robust configuration.** A central `Config` struct and the fluent
  `ConfigBuilder` make setups atomic and predictable instead of a sequence
  of stateful setter calls.
- **Streaming.** A dedicated API (`Start`, `Append`, `Close`) renders rows
  as they arrive.
- **Type safety.** `tw.State`, `tw.Align`, and friends replace integer
  constants and booleans.
- **Consistent API.** Headers, rows, and footers share one configuration
  template, `tw.CellConfig`.
- **New features.** Hierarchical merging, granular per-side padding, table
  captions, and fixed column widths.

## The philosophy changes

### Configuration instead of setters

Old code configured a table by calling setters one at a time. New code
defines behavior with a `tablewriter.Config` and visual styling with a
`tw.Rendition`, usually at creation time:

```go
// v0.0.5
table := tablewriter.NewWriter(os.Stdout)
table.SetHeader([]string{"Name", "Age"})
table.SetAlignment(tablewriter.ALIGN_LEFT)

// v1
table := tablewriter.NewTable(os.Stdout,
	tablewriter.WithConfig(tablewriter.Config{
		Header: tw.CellConfig{Alignment: tw.CellAlignment{Global: tw.AlignCenter}},
		Row:    tw.CellConfig{Alignment: tw.CellAlignment{Global: tw.AlignLeft}},
	}),
)
```

### Rendering is decoupled

In v0.0.5 the rendering logic lived inside the `Table` struct. In v1 the
`tw.Renderer` interface owns it, and `renderer.NewBlueprint()` is the
default text renderer. Appearance is controlled by `tw.Rendition`
(borders, symbols, settings), so swapping formats is a one-line change.

### One section template

Headers, rows, and footers used to have separate, inconsistent
configuration methods. They now share `tw.CellConfig`, which combines
formatting, padding, alignment, filtering, callbacks, widths, and merging
with global and per-column control.

### Builders instead of call chains

`tablewriter.NewConfigBuilder()` builds a `Config` through chained
builders — `Header()`, `Row()`, `Footer()`, `Behavior()`, `ForColumn()` —
with sub-builders for alignment, formatting, merging, padding, callbacks,
and filters. See [Configuration](configuration).

### Streaming is explicit

v0.0.5 had no streaming support. v1 enables it with
`Config.Stream.Enable` or `WithStreaming(tw.StreamConfig{Enable: true})`
and drives it with `Table.Start()`, `Table.Append()`, and `Table.Close()`.

### Errors are returned

In v0.0.5, `Render()` returned nothing. In v1, `Render()`, `Start()`,
`Close()`, `Append()`, and `Bulk()` all return errors. Check them.

### Types instead of magic constants

Integer constants like `ALIGN_CENTER` and bare booleans are gone from the
public API. The `tw` package provides `tw.Align`, `tw.State`,
`tw.Position`, `tw.CellConfig`, and the rest.

## Method mapping

| v0.0.5 | v1 equivalent | Notes |
|---|---|---|
| `NewWriter(w)` | `NewTable(w, opts...)` | `NewWriter` is kept as a wrapper around `NewTable` |
| `SetHeader([]string)` | `Header(...any)` | Variadic or slice; any type |
| `Append([]string)` | `Append(...any)` | Variadic, slice, or struct |
| `AppendBulk([][]string)` | `Bulk([]any)` | Slice of rows; any type |
| `SetFooter([]string)` | `Footer(...any)` | Variadic or slice; any type |
| `Render()` | `Render()` (returns `error`) | Batch mode; streaming uses `Start`/`Close` |
| `SetBorder(bool)` | `WithRendition(tw.Rendition{Borders: ...})` or `WithBorders(tw.Border)` | Prefer `WithRendition`; `WithBorders` is kept for compatibility |
| `SetRowLine(bool)` | `WithRendition(tw.Rendition{Settings: tw.Settings{Separators: tw.Separators{BetweenRows: tw.On}}})` | |
| `SetHeaderLine(bool)` | `WithRendition(tw.Rendition{Settings: tw.Settings{Lines: tw.Lines{ShowHeaderLine: tw.On}}})` | |
| `SetColumnSeparator(string)` | `WithRendition(tw.Rendition{Symbols: ...})` or `WithSymbols(tw.Symbols)` | `tw.NewSymbols` or a custom `tw.Symbols` |
| `SetCenterSeparator(string)` | `WithRendition(tw.Rendition{Symbols: ...})` or `WithSymbols(tw.Symbols)` | |
| `SetRowSeparator(string)` | `WithRendition(tw.Rendition{Symbols: ...})` or `WithSymbols(tw.Symbols)` | |
| `SetAlignment(int)` | `WithRowAlignment(tw.Align)` or `Config.Row.Alignment.Global` | |
| `SetHeaderAlignment(int)` | `WithHeaderAlignment(tw.Align)` or `Config.Header.Alignment.Global` | |
| `SetAutoFormatHeaders(bool)` | `WithHeaderAutoFormat(tw.State)` or `Config.Header.Formatting.AutoFormat` | |
| `SetAutoWrapText(bool)` | `WithRowAutoWrap(int)` or `Config.Row.Formatting.AutoWrap` | Uses the `tw.Wrap*` constants |
| `SetAutoMergeCells(bool)` | `WithRowMergeMode(int)` or `Config.Row.Formatting.MergeMode` | Supports vertical and hierarchical merging |
| `SetColMinWidth(col, w)` | `WithColumnWidths(tw.NewMapper[int,int]().Set(col, w))` or `Config.Widths.PerColumn` | Fixed widths via `Config.Widths` |
| `SetTablePadding(string)` | `Config.<Section>.Padding.Global` with `tw.Padding` | No direct equivalent; manage per section |
| `SetDebug(bool)` | `WithDebug(bool)` or `Config.Debug` | Inspect with `Table.Debug()` |
| `Clear()` | `Reset()` | Clears data and state |
| `SetNoWhiteSpace(bool)` | `WithTrimSpace(tw.Off)` (if `true`) or `WithPadding(tw.PaddingNone)` | Via `Config.Behavior.TrimSpace` or padding |
| `SetColumnColor(Colors)` | Embed ANSI codes in cell data, use `tw.Formatter`, or `Config.<Section>.Filter` | Color now comes from data, formatters, or the Colorized renderer |
| `SetHeaderColor(Colors)` | Same as above, for the header | |
| `SetFooterColor(Colors)` | Same as above, for the footer | |
| (no equivalent) | `Start()` / `Close()` | Streaming lifecycle |
| (no equivalent) | `Caption(tw.Caption)` | Table captions |
| (no equivalent) | `WithStreaming(tw.StreamConfig)` | Streaming configuration |

## Migration steps

### 1. Initialization

Replace the constructor and move setter calls into options or a `Config`:

```go
// Before
table := tablewriter.NewWriter(os.Stdout)
table.SetHeader([]string{"Name", "Status"})
table.SetBorder(false)

// After
table := tablewriter.NewTable(os.Stdout,
	tablewriter.WithRenderer(renderer.NewBlueprint(tw.Rendition{
		Borders: tw.BorderNone,
	})),
)
table.Header("Name", "Status")
```

`NewWriter` still exists, so a minimal migration can keep it and change
only the calls that no longer compile — but `NewTable` with options is the
form new code should use.

### 2. Data input

`Append` and `Bulk` accept the same shapes as before, plus structs and
`any` values. The main change is that both now return an `error`:

```go
// Before
table.AppendBulk(data)

// After
if err := table.Bulk(data); err != nil {
	log.Fatal(err)
}
```

Struct input is new in v1: pass a struct to `Append` or a slice of structs
to `Bulk`, and optionally describe columns with `tw` struct tags. See
[Data Input](data-input).

### 3. Rendering

`Render()` now returns an error, and streaming tables use `Start()` and
`Close()` instead:

```go
// Batch
if err := table.Render(); err != nil {
	log.Fatal(err)
}

// Streaming
if err := table.Start(); err != nil {
	log.Fatal(err)
}
table.Header("ID", "Status")
table.Append([]string{"1", "OK"})
if err := table.Close(); err != nil {
	log.Fatal(err)
}
```

### 4. Styling

Styling moved from setters into `Config` and `tw.Rendition`:

- **Table styles** — borders, symbols, separators, and lines live in
  `tw.Rendition`, applied with `WithRendition` or passed to
  `renderer.NewBlueprint`.
- **Alignment** — `tw.Align` values on `Config.<Section>.Alignment`, or the
  `With*Alignment` options.
- **Auto-formatting** — `Config.Header.Formatting.AutoFormat` with a
  `tw.State`. Note that header auto-format behavior changed in v1; if
  headers come out differently than in v0.0.5, this is the first setting
  to check.
- **Text wrapping** — `Config.Row.Formatting.AutoWrap` with the `tw.Wrap*`
  constants.
- **Padding** — `Config.<Section>.Padding` with `tw.Padding`, globally or
  per column.
- **Column widths** — `Config.Widths` and `Config.<Section>.ColMaxWidths`.
- **Colors** — there is no per-column color setter anymore. Color comes
  from the Colorized renderer (`renderer.NewColorized` with
  `renderer.ColorizedConfig`), from ANSI codes embedded in cell data, or
  from a `tw.Formatter` / section `Filter`.

### 5. Advanced features

Features that did not exist in v0.0.5 and are worth adopting while you
migrate:

- **Cell merging** — `tw.MergeVertical`, `tw.MergeHorizontal`,
  `tw.MergeBoth`, and `tw.MergeHierarchical` on `CellMerging.Mode`, with
  `ByColumnIndex` scoping.
- **Table captions** — `table.Caption(tw.Caption{}.WithText(...))`.
- **Filters** — `CellFilter.Global` and `CellFilter.PerColumn` transform
  cell content at render time.
- **Stringers and caching** — `WithStringer`, `WithStringerCache`, and
  `WithStringerCacheCustom` for custom value rendering.
- **Custom renderers** — implement `tw.Renderer` for output formats the
  library does not ship.

## Common pitfalls

- **Do not use v1.0.0.** It is missing functionality. Migrate to v1.1.x
  (or `@latest`) instead.
- **Check `Render()` errors.** Code that ignored the old void `Render()`
  will silently drop failures if the error is not handled.
- **`AppendBulk` is now `Bulk`.** The old name does not exist in v1.
- **Setters are gone.** `SetHeader`, `SetFooter`, `SetAlignment`,
  `SetBorder`, and friends exist only in `deprecated.go` for
  compatibility. Replace them; do not build new code on them.
- **Alignment constants changed type.** `ALIGN_CENTER` and friends were
  integers. v1 uses `tw.Align` values (`tw.AlignLeft`, `tw.AlignCenter`,
  `tw.AlignRight`).
- **Auto-format changed.** Header title-casing behavior differs from
  v0.0.5. If headers look wrong after migrating, check
  `Config.Header.Formatting.AutoFormat`.
- **Counting changed.** Line counting behavior changed in v1; if you rely
  on `Table.Lines()` or counters, verify the numbers against the new
  version.
- **CSV output is a custom renderer now.** The old built-in CSV output is
  gone as a mode; implement `tw.Renderer` (or use the pattern in the
  migration guide's custom-renderer example) if you need it.
- **Legacy docs.** The v0.0.5 documentation is preserved as
  `README_LEGACY.md` in the repository for reference while you migrate.
