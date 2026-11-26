---
title: Introduction
description: What TableWriter is, what it can do, and how to install it.
---

# Introduction

TableWriter is a Go library for generating rich text-based tables. It renders
data as ASCII and Unicode tables in the terminal, and can also produce
Markdown, HTML, SVG, and ANSI-colored output from the same table definition.
It fits CLI tools, logs, reports, and anywhere else you need readable
tabular output without reaching for a spreadsheet.

```go
table := tablewriter.NewTable(os.Stdout)
table.Header("Name", "Age", "City")
table.Append([]string{"Alice", "25", "New York"})
table.Append([]string{"Bob", "30", "Boston"})
table.Render()
```

```
┌───────┬─────┬──────────┐
│ NAME  │ AGE │   CITY   │
├───────┼─────┼──────────┤
│ Alice │ 25  │ New York │
│ Bob   │ 30  │ Boston   │
└───────┴─────┴──────────┘
```

## What it does

- **Multi-format rendering.** One table definition renders as ASCII,
  Unicode, Markdown, HTML, SVG, or ANSI-colored text. Pick a renderer when
  you create the table; the data and configuration stay the same.
- **Advanced styling.** Cell merging (vertical, horizontal, both, and
  hierarchical), per-column and global alignment, per-side padding, custom
  border symbols, and captions.
- **Flexible input.** Feed a table from slices, structs, CSV files, CSV
  readers, or a live stream of rows. Struct fields can carry `tw` tags that
  control alignment, width, and padding per column.
- **Streaming.** Rows can be rendered as they arrive, so a long-running job
  can show a live table instead of waiting for all data before printing
  anything.
- **High performance.** Minimal allocations and buffer reuse keep rendering
  cheap even for large tables.
- **Modern Go.** Generics support throughout the helper types (`tw.Mapper`,
  `tw.Slicer`), a type-safe configuration system built on `tw.State` and
  `tw.Align`, and a decoupled `tw.Renderer` interface you can implement
  yourself.

## When to use it

TableWriter shines whenever a program needs to show a human a small-to-medium
set of rows:

- CLI tools printing status, inventory, or query results.
- Log summaries and end-of-run reports.
- Database query output, including `sql.Null*` values, which render
  natively from structs.
- Markdown tables inside generated documentation.
- HTML or SVG tables embedded in web pages and reports.

If your output is a fixed-width report, a dashboard, or a data export,
TableWriter is the right weight: richer than `fmt.Printf` alignment, far
lighter than a full reporting engine.

## Installation

TableWriter requires Go 1.21 or newer. The module is
`github.com/olekukonko/tablewriter`.

Install the latest v1 release:

```bash
go get github.com/olekukonko/tablewriter@latest
```

Or pin the current stable version:

```bash
go get github.com/olekukonko/tablewriter@v1.1.5
```

Then import it:

```go
import (
	"github.com/olekukonko/tablewriter"
	"github.com/olekukonko/tablewriter/renderer"
	"github.com/olekukonko/tablewriter/tw"
)
```

The main package is `tablewriter`. Renderers live in
`github.com/olekukonko/tablewriter/renderer`, and the shared types —
alignment, state, symbols, mappers, and the `Renderer` interface — live in
`github.com/olekukonko/tablewriter/tw`. Width helpers live in
`github.com/olekukonko/tablewriter/pkg/twwidth`.

## v1 and the legacy version

There are two lines of TableWriter in the wild:

- **v1 (current).** The redesigned, configuration-driven API documented
  here: `NewTable`, `Config`, `ConfigBuilder`, `With*` options, the
  `tw.Renderer` interface, and streaming. Use `@latest` for new projects.
- **v0.0.5 (legacy).** The original method-driven API (`NewWriter`,
  `SetHeader`, `AppendBulk`, and friends). It still works and is kept for
  existing applications. Install it explicitly if you need it:

  ```bash
  go get github.com/olekukonko/tablewriter@v0.0.5
  ```

Do not use `v1.0.0`: it is missing functionality and should be skipped in
favor of v1.1.x. If you are coming from v0.0.5, read the
[Migration](migration) guide — it maps every legacy method to its v1
equivalent.

## How the pieces fit

A table has three layers:

- **Data.** Headers, rows, and footers, added with `Header`, `Append`,
  `Bulk`, and `Footer`. Data can be strings, any value with a `String()`
  method, structs, or CSV records.
- **Configuration.** A `Config` controls behavior and cell layout: alignment,
  wrapping, padding, merging, widths, and trimming. You build it with a
  `Config` struct literal, `Configure`, `With*` options, or the fluent
  `ConfigBuilder`. See [Configuration](configuration).
- **Rendering.** A `tw.Renderer` turns the configured table into output. The
  default is `renderer.NewBlueprint`, which draws Unicode tables. Swap it
  with `WithRenderer` for Markdown, HTML, SVG, colorized, or streaming
  output. See [Renderers](renderers).

The next page walks through your first table end to end:
[Quickstart](quickstart).
