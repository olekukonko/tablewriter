---
title: Quickstart
description: Build and render your first table in under a minute.
---

# Quickstart

Every table follows the same four steps: create it, configure it (optional),
add data, render it. This page runs through a complete example and explains
what each call does.

## Your first table

```go
package main

import (
	"os"

	"github.com/olekukonko/tablewriter"
)

func main() {
	data := [][]string{
		{"Package", "Version", "Status"},
		{"tablewriter", "v0.0.5", "legacy"},
		{"tablewriter", "v1.1.5", "latest"},
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.Header(data[0])
	table.Bulk(data[1:])
	table.Render()
}
```

Run it and you get:

```
┌─────────────┬─────────┬────────┐
│   PACKAGE   │ VERSION │ STATUS │
├─────────────┼─────────┼────────┤
│ tablewriter │ v0.0.5  │ legacy │
│ tablewriter │ v1.1.5  │ latest │
└─────────────┴─────────┴────────┘
```

Four things happened:

1. `NewWriter(os.Stdout)` created a table writing to standard output with
   default settings. `NewTable` does the same thing and also accepts
   configuration options — prefer it for anything beyond the simplest case.
2. `Header` set the header row. Headers are auto-formatted to title case by
   default, which is why `Package` prints as `PACKAGE`.
3. `Bulk` added all data rows at once. It accepts a slice of rows — slices
   of strings, slices of `any`, or a slice of structs.
4. `Render` wrote the finished table. It returns an `error`, so check it in
   real code: `if err := table.Render(); err != nil { ... }`.

## The same table with NewTable

`NewTable` is the constructor you will use most. It takes the same writer
plus any number of `With*` options:

```go
package main

import (
	"fmt"
	"os"

	"github.com/olekukonko/tablewriter"
)

type Age int

func (a Age) String() string {
	return fmt.Sprintf("%d yrs", a)
}

func main() {
	data := [][]any{
		{"Alice", Age(25), "New York"},
		{"Bob", Age(30), "Boston"},
	}

	table := tablewriter.NewTable(os.Stdout)
	table.Header("Name", "Age", "City")
	table.Bulk(data)
	table.Render()
}
```

```
┌───────┬────────┬──────────┐
│ NAME  │  AGE   │   CITY   │
├───────┼────────┼──────────┤
│ Alice │ 25 yrs │ New York │
│ Bob   │ 30 yrs │ Boston   │
└───────┴────────┴──────────┘
```

Two details worth noticing:

- `Header` is variadic: `table.Header("Name", "Age", "City")` works the same
  as passing a slice.
- Values are not limited to strings. `Age(25)` is rendered through its
  `String()` method, so the table shows `25 yrs`. Any value with a
  `String()` method works this way; values with a `Format()` method use
  that instead — `Format()` wins when both exist.

## Adding rows one at a time

Use `Append` when rows arrive one by one, for example inside a loop:

```go
table := tablewriter.NewTable(os.Stdout)
table.Header("Name", "Age", "City")
table.Append([]string{"Alice", "25", "New York"})
table.Append([]string{"Bob", "30", "Boston"})
table.Render()
```

`Append` accepts a single row as a slice, a struct, or a set of variadic
values, and it returns an `error`. `Bulk` is the batch version of the same
idea. Both can be mixed freely — append some rows, bulk-add the rest.

## Adding a footer

`Footer` adds a closing row, typically a total or summary:

```go
table := tablewriter.NewTable(os.Stdout)
table.Header("Item", "Qty", "Price")
table.Append([]string{"Apples", "4", "$2.00"})
table.Append([]string{"Bread", "1", "$3.50"})
table.Footer([]string{"", "Total", "$5.50"})
table.Render()
```

```
┌────────┬───────┬───────┐
│  ITEM  │  QTY  │ PRICE │
├────────┼───────┼───────┤
│ Apples │ 4     │ $2.00 │
│ Bread  │ 1     │ $3.50 │
├────────┼───────┼───────┤
│        │ Total │ $5.50 │
└────────┴───────┴───────┘
```

## Writing somewhere else

The writer is any `io.Writer`. Render into a buffer to capture the output
as a string — useful for tests, for embedding a table inside another
document, or for nested tables:

```go
var buf bytes.Buffer
table := tablewriter.NewTable(&buf)
table.Header("Name", "Age")
table.Append([]string{"Alice", "25"})
table.Render()
fmt.Print(buf.String())
```

## What the defaults give you

A table created with `NewTable(os.Stdout)` and no options renders with the
Blueprint renderer: Unicode borders, centered auto-formatted headers,
left-aligned rows, and a single space of padding on each side of every cell.
Column widths come from the widest cell in each column. That default covers
most cases; when it does not, the next pages show how to change it:

- [Configuration](configuration) — alignment, wrapping, padding, merging,
  and widths.
- [Rendering](rendering) — the render lifecycle, streaming, and captions.
- [Renderers](renderers) — Markdown, HTML, SVG, colorized, and streaming
  output.
