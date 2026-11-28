---
title: Rendering
description: The render lifecycle, streaming tables, captions, and reusing a table.
---

# Rendering

Rendering is the moment a table turns its data and configuration into
output. This page covers the lifecycle — creating, filling, and rendering a
table — plus streaming mode, captions, and what you can do after a render.

## The lifecycle

A table goes through four stages:

```go
// 1. Create — pick a writer and any options.
table := tablewriter.NewTable(os.Stdout)

// 2. Configure — optional; see Configuration.
table.Configure(func(config *tablewriter.Config) {
	config.Row.Alignment.Global = tw.AlignLeft
})

// 3. Add data — header, rows, footer, in any order.
table.Header("Name", "Age", "City")
table.Append([]string{"Alice", "25", "New York"})
table.Bulk(data)
table.Footer([]string{"", "Total", "2"})

// 4. Render — writes everything to the writer.
if err := table.Render(); err != nil {
	log.Fatal(err)
}
```

`Render()` returns an `error`. Check it: rendering can fail if the writer
fails, if the configuration is inconsistent, or if a custom renderer
reports a problem. In v1, `Append` and `Bulk` also return errors — handle
them the same way instead of ignoring them.

## Constructors

- `NewTable(w io.Writer, opts ...Option) *Table` — the main constructor.
  Pass `With*` options to configure the table as you create it. The
  default renderer is Blueprint (Unicode borders).
- `NewWriter(w io.Writer) *Table` — creates a table with default settings.
  Kept for compatibility with v0.0.5 code; prefer `NewTable` for new code.

Both write to any `io.Writer`: `os.Stdout`, a file, a `bytes.Buffer`, or a
network connection.

## Adding data

| Method | What it does |
|---|---|
| `Header(elements ...any)` | Sets the header row. Accepts variadic values or a single slice. |
| `Footer(elements ...any)` | Sets the footer row. Same calling conventions as `Header`. |
| `Append(rows ...interface{}) error` | Adds one row per argument. Each row can be a slice, a struct, or variadic values. |
| `Bulk(rows interface{}) error` | Adds many rows at once from a slice of rows or a slice of structs. |
| `Reset()` | Clears all data and state so the table can be reused. |
| `Lines() int` | Reports how many lines the last render produced. |
| `Counters() []tw.Counter` | Returns the attached line counters. |
| `Debug() *bytes.Buffer` | Returns the buffered debug log (only when `Debug` is on). |
| `Trimmer(str string) string` | Applies the table's trimming rules to a string. |
| `Config() Config` | Returns a copy of the current configuration. |
| `Configure(fn func(cfg *Config)) *Table` | Mutates the live configuration. |
| `Options(opts ...Option) *Table` | Applies more options to an existing table. |
| `Renderer() tw.Renderer` | Returns the active renderer. |
| `Logger() *ll.Logger` | Returns the table's logger. |

Rows accept more than strings. Any value with a `String()` method renders
through it, and a value with a `Format()` method uses `Format()` instead.
Structs are expanded field by field — see [Data Input](data-input) for
the full rules, including `tw` struct tags.

## Captions

`Caption` attaches a caption to the table. Build it with the fluent
`tw.Caption` methods:

```go
table.Caption(tw.Caption{}.WithText("Monthly sales by region").WithSpot(tw.SpotTopCenter))
```

`Caption.WithText` sets the text, `WithAlign` sets its alignment,
`WithSpot` sets its placement, and `WithWidth` caps its width. Captions
render outside the table frame, above or below it depending on the spot.

## Streaming

Streaming renders rows as they arrive instead of waiting for the full
dataset. Enable it with `WithStreaming`, then drive the table with
`Start`, `Append`, and `Close`:

```go
table := tablewriter.NewTable(os.Stdout,
	tablewriter.WithStreaming(tw.StreamConfig{Enable: true}),
)

if err := table.Start(); err != nil {
	log.Fatalf("Start failed: %v", err)
}
defer table.Close()

table.Header([]string{"ID", "Description", "Status"})

for _, row := range data {
	table.Append(row)
	time.Sleep(500 * time.Millisecond) // simulate a live feed
}

table.Footer([]string{"", "Total", "3"})
```

The header prints as soon as `Header` is called, each row prints as it is
appended, and the footer and bottom border print when `Close` runs. The
output appears incrementally, which is what makes streaming right for
progress displays, log tails, and long-running jobs.

`tw.StreamConfig` has three fields:

- `Enable` — turns streaming on.
- `StrictColumns` — when true, `Append` returns an error if a row's cell
  count does not match the established column count. When false (the
  default), mismatched rows are padded or truncated with a warning log.
- `Widths` — a `tw.CellWidth` for streaming width control. Deprecated in
  favor of top-level `Config.Widths`, but still respected when
  `Config.Widths` is not set.

Two behaviors differ in streaming mode: auto-hiding empty columns is
ignored (the renderer cannot know a column is empty until the stream
ends), and trimming rules that need the whole cell are limited. The
Ocean renderer is built for this mode — see [Renderers](renderers).

## Nested tables

Because a table writes to any `io.Writer`, you can render a small table
into a `bytes.Buffer` and use the result as a cell in a bigger table:

```go
createSubTable := func(s string) string {
	var buf bytes.Buffer
	table := tablewriter.NewTable(&buf,
		tablewriter.WithRenderer(renderer.NewBlueprint(tw.Rendition{
			Borders: tw.BorderNone,
			Symbols: tw.NewSymbols(tw.StyleASCII),
			Settings: tw.Settings{
				Separators: tw.Separators{BetweenRows: tw.On},
				Lines:      tw.Lines{ShowFooterLine: tw.On},
			},
		})),
		tablewriter.WithConfig(tablewriter.Config{
			MaxWidth: 10,
			Row:      tw.CellConfig{Alignment: tw.CellAlignment{Global: tw.AlignCenter}},
		}),
	)
	table.Append([]string{s, s})
	table.Append([]string{s, s})
	table.Render()
	return buf.String()
}

table := tablewriter.NewTable(os.Stdout,
	tablewriter.WithRenderer(renderer.NewBlueprint(tw.Rendition{
		Borders:  tw.BorderNone,
		Settings: tw.Settings{Separators: tw.Separators{BetweenColumns: tw.On}},
	})),
	tablewriter.WithConfig(tablewriter.Config{
		MaxWidth: 30,
		Row:      tw.CellConfig{Alignment: tw.CellAlignment{Global: tw.AlignCenter}},
	}),
)
table.Append([]string{createSubTable("A"), createSubTable("B")})
table.Append([]string{createSubTable("C"), createSubTable("D")})
table.Render()
```

```
  A | A  │  B | B
 ---+--- │ ---+---
  A | A  │  B | B
  C | C  │  D | D
 ---+--- │ ---+---
  C | C  │  D | D
```

## Reusing a table

`Reset()` clears headers, rows, footers, and render state. The
configuration and renderer stay in place, so you can refill the same table
with a new dataset and call `Render()` again. This is useful for periodic
reports that share one layout.

## Errors and debugging

Turn on `Debug` (via `WithDebug(true)` or `Config.Debug`) to record what
the renderer does. `Table.Debug()` returns the log as a `*bytes.Buffer`.
The log is most useful when a table renders empty, columns come out the
wrong width, or a renderer reports an error — it shows the configuration
the renderer actually saw, which is often not the configuration you thought
you set.
