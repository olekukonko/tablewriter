---
title: csv2table Command-Line Tool
description: Render CSV files as tables from the terminal, with no Go code.
---

# csv2table Command-Line Tool

`csv2table` is the command-line front end for TableWriter. It reads CSV
data from a file or standard input and prints a formatted table — no Go
code required. It is the fastest way to try the library, and a handy tool
in its own right for inspecting CSV files.

## Install

```bash
go install github.com/olekukonko/tablewriter/cmd/csv2table@latest
```

## Basic usage

Render a file:

```bash
csv2table -f test.csv
```

A positional file argument works too, and overrides `-f`:

```bash
csv2table test.csv
```

Pipe data in:

```bash
cat test.csv | csv2table -p=true
```

With no file and no `-p`, `csv2table` reads from standard input anyway, so
this also works:

```bash
cat test.csv | csv2table
```

Treat the first row as data instead of a header:

```bash
echo dance,with,me | csv2table -p=true -h=false
```

```
+-------+------+-----+
| dance | with | me  |
+-------+------+-----+
```

A typical run looks like this:

```bash
csv2table -f test.csv -h true -a left
```

## Flags

| Flag | Default | Purpose |
|---|---|---|
| `-f` | `""` | CSV file path, e.g. `-f sample.csv`. Empty means standard input unless `-p` is set. |
| `-d` | `","` | CSV delimiter, e.g. `"\|"` or `"\t"`. For tab, pass an actual tab or the literal `\t`. |
| `-h` | `true` | Treat the first row as a header. |
| `-a` | `"none"` | Global cell alignment: `none`, `left`, `right`, or `center`. `none` uses the renderer defaults. |
| `-p` | `false` | Read CSV from standard input. Overrides `-f` when both are set. |
| `-b` | `true` | Enable table borders and lines (the overall structure). |
| `-s` | `false` | Streaming mode: process rows one at a time. May not support all features, such as auto-hiding empty columns. |
| `-renderer` | `"blueprint"` | Table renderer: `blueprint`, `colorized`, `markdown`, `html`, `svg`, or `ocean`. |
| `-symbols` | `"light"` | Border symbol style: `light`, `ascii`, `heavy`, `double`, `rounded`, `markdown`, `graphical`, `dotted`, `arrow`, `starry`, `hearts`, `tech`, `nature`, `artistic`, `8-bit`, `chaos`, `dots`, `blocks`, `zen`, `none`. |
| `-wrap` | `"normal"` | Row auto-wrap mode: `normal`, `truncate`, `break`, or `none`. |
| `-infer` | `true` | Infer and normalize column counts when CSV rows are ragged. When `false`, a row with the wrong number of fields halts parsing with a CSV error. |
| `-maxwidth` | `0` | Maximum table width in characters. `0` means automatic: 90% of the terminal width, or content-based when the terminal size cannot be read. |
| `-debug` | `false` | Enable debug logging for TableWriter operations. |

The usage text is printed with `-h` — note that `-h` is already the
header flag, so there is no separate help flag; run `csv2table` with an
unknown flag to see the usage summary.

## Renderers

The `-renderer` flag picks the output format. `blueprint` is the default
Unicode table. `markdown` prints a Markdown table for pasting into docs,
`html` prints a `<table>` element, `svg` prints an SVG document, and
`colorized` prints an ANSI-colored table. `ocean` is the streaming
renderer — combine it with `-s` for row-by-row output.

Graphical renderers (`html`, `svg`, and friends) print without the leading
and trailing blank lines that the text renderers add for readability.

## Symbol styles

The `-symbols` flag selects the border characters. `light` is the default
Unicode set; `ascii` uses `+`, `-`, and `|`; `heavy` and `double` use
heavier line weights; `rounded` uses curved corners; `markdown` uses the
pipe style of Markdown tables. The named pictorial styles — `dotted`,
`arrow`, `starry`, `hearts`, `nature`, `artistic`, `8-bit`, `chaos`,
`dots`, `blocks`, `zen`, `tech` — are built from custom symbol sets. Use
`none` for a borderless table.

## Ragged CSV files

Real CSV files do not always have the same number of fields in every row.
With `-infer` on (the default), `csv2table` makes two passes: it allows
variable field counts, infers the column count, and normalizes short rows.
With `-infer=false`, the standard `encoding/csv` behavior applies and a
mismatched row stops the run with an error. Turn inference off when you
want malformed input to fail loudly instead of being padded silently.

Custom delimiters work the same way: `-d "|"` reads pipe-separated files,
and `-d "\t"` reads tab-separated files.

## Streaming mode

With `-s`, rows are rendered as they arrive instead of all at once:

```bash
tail -f access.log | csv2table -p=true -s=true -renderer ocean
```

Streaming mode has limits: features that need the whole table up front —
auto-hiding empty columns, for example — are not available while
streaming. Widths are fixed from the rows seen so far, and long cells are
truncated rather than reflowed.

## Examples

Render a CSV file as Markdown for documentation:

```bash
csv2table -f sales.csv -renderer markdown
```

Render a pipe-delimited file with heavy borders:

```bash
csv2table -f data.psv -d "|" -symbols heavy
```

Render a file with a capped width and truncated cells:

```bash
csv2table -f long.csv -maxwidth 100 -wrap truncate
```

Watch a growing file as a live table:

```bash
tail -f events.csv | csv2table -p=true -s=true
```

## How it works

`csv2table` is a thin wrapper around the library: it builds a CSV reader
from the file or standard input, picks symbols and a renderer from the
flags, computes the table width (from `-maxwidth`, or 90% of the terminal
width when the flag is `0`), and renders through `tablewriter` with
`WithRenderer` and `WithTableMax`. Everything the tool can do, the library
can do — if you outgrow the flags, the same calls in Go give you full
control. See [Configuration](configuration) and
[Renderers](renderers).
