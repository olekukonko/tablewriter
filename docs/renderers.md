---
title: Renderers
description: Every output format TableWriter supports, from Unicode tables to HTML, SVG, and color.
---

# Renderers

A renderer turns a configured table into output. You pick one with
`WithRenderer` when you create the table; everything else — data,
configuration, sections — stays the same. The default renderer is
Blueprint, which draws Unicode tables.

```go
table := tablewriter.NewTable(os.Stdout,
	tablewriter.WithRenderer(renderer.NewMarkdown()),
)
```

## The Renderer interface

All renderers implement `tw.Renderer`:

```go
type Renderer interface {
	Start(w io.Writer) error
	Header(headers [][]string, ctx Formatting)
	Row(row []string, ctx Formatting)
	Footer(footers [][]string, ctx Formatting)
	Line(ctx Formatting)
	Config() Rendition
	Close() error
	Logger(logger *ll.Logger)
}
```

`Start` and `Close` bracket a render (streaming renderers use them to emit
the opening and closing frame). `Header`, `Row`, and `Footer` emit one
section each, `Line` emits a separator, and `Config` returns the
`tw.Rendition` the renderer is using. If you need an output format that
does not ship with the library, implement this interface and pass your
renderer to `WithRenderer`. `Formatting`, `RowContext`, and `CellContext`
carry everything a custom renderer needs: the row's cells, their
alignments, widths, padding, and merge state.

## Rendition: how a renderer looks

`tw.Rendition` is the visual styling a renderer applies:

```go
type Rendition struct {
	Borders   Border   // outer frame visibility
	Symbols   Symbols  // border and junction characters
	Settings  Settings // separators and lines
	Streaming bool
}
```

`tw.Border` toggles each side of the outer frame (`Left`, `Right`, `Top`,
`Bottom`) and has an `Overwrite` flag. `tw.Settings` holds `Separators`
(`ShowHeader`, `ShowFooter`, `BetweenRows`, `BetweenColumns`) and `Lines`
(`ShowTop`, `ShowBottom`, `ShowHeaderLine`, `ShowFooterLine`), plus a
`CompactMode` state reserved for future use. Every field is a `tw.State`:
`tw.On`, `tw.Off`, or `tw.Unknown` for the default.

`tw.Symbols` is the set of characters used to draw the table. Build one
with `tw.NewSymbols(style)` — the styles are `tw.StyleASCII`,
`tw.StyleLight`, `tw.StyleHeavy`, `tw.StyleDouble`,
`tw.StyleDoubleLong`, `tw.StyleLightHeavy`, `tw.StyleHeavyLight`,
`tw.StyleLightDouble`, `tw.StyleDoubleLight`, `tw.StyleRounded`,
`tw.StyleMarkdown`, `tw.StyleGraphical`, `tw.StyleMerger`,
`tw.StyleDotted`, `tw.StyleArrow`, `tw.StyleStarry`, `tw.StyleHearts`,
`tw.StyleCircuit`, `tw.StyleNature`, `tw.StyleArtistic`, `tw.Style8Bit`,
`tw.StyleChaos`, `tw.StyleDots`, `tw.StyleBlocks`, `tw.StyleZen`,
`tw.StyleVintage`, `tw.StyleSketch`, `tw.StyleArrowDouble`,
`tw.StyleCelestial`, `tw.StyleCyber`, `tw.StyleRunic`,
`tw.StyleIndustrial`, `tw.StyleInk`, `tw.StyleArcade`, `tw.StyleBlossom`,
`tw.StyleDefault`, `tw.StyleFrosted`, `tw.StyleMosaic`, `tw.StyleUFO`,
`tw.StyleSteampunk`, `tw.StyleGalaxy`, `tw.StyleJazz`, `tw.StylePuzzle`,
`tw.StyleHypno`, and `tw.StyleNone`.

For full control, build a custom set with `tw.NewSymbolCustom(name)` and
its `With*` methods:

```go
symbols := tw.NewSymbolCustom("Nature").
	WithRow("~").
	WithColumn("|").
	WithTopLeft("🌱").
	WithTopMid("🌿").
	WithTopRight("🌱").
	WithMidLeft("🍃").
	WithCenter("❀").
	WithMidRight("🍃").
	WithBottomLeft("🌻").
	WithBottomMid("🌾").
	WithBottomRight("🌻")

table := tablewriter.NewTable(os.Stdout,
	tablewriter.WithRenderer(renderer.NewBlueprint(tw.Rendition{Symbols: symbols})),
)
```

Pass a `Rendition` to `WithRendition` to change the look without swapping
the renderer, or to `renderer.NewBlueprint` / `renderer.NewMarkdown`
directly.

## Blueprint (default)

`renderer.NewBlueprint(configs ...tw.Rendition) *Blueprint` draws the
classic Unicode table — the output you saw on the Quickstart page. It is
the default when you do not pass a renderer, and the right choice for
terminal output. Pass a `Rendition` to change its borders, symbols, and
separator lines:

```go
table := tablewriter.NewTable(os.Stdout,
	tablewriter.WithRenderer(renderer.NewBlueprint(tw.Rendition{
		Borders: tw.BorderNone,
		Symbols: tw.NewSymbols(tw.StyleASCII),
		Settings: tw.Settings{
			Separators: tw.Separators{BetweenRows: tw.On},
			Lines:      tw.Lines{ShowFooterLine: tw.On},
		},
	})),
)
```

With `StyleASCII` the same table renders with `+`, `-`, and `|` instead of
Unicode box-drawing characters — useful when the output may pass through
an ASCII-only channel.

## Markdown

`renderer.NewMarkdown(configs ...tw.Rendition) *Markdown` renders a
GitHub-flavored Markdown table, ready to paste into documentation:

```go
table := tablewriter.NewTable(os.Stdout,
	tablewriter.WithRenderer(renderer.NewMarkdown()),
)
table.Header([]string{"Name", "Age", "City"})
table.Bulk(data)
table.Render()
```

```
|    NAME    |  AGE   |   CITY   |
|:-----------|:-------|:---------|
| Alice Mask | 25 yrs | New York |
| Bob Marley | 30 yrs | Boston   |
```

Alignment markers in the separator row follow the column alignment, and
cell content is escaped so Markdown-significant characters do not break
the table. Use it for generated docs, README snippets, and issue reports.

## HTML

`renderer.NewHTML(configs ...HTMLConfig) *HTML` renders a `<table>` element
with per-section CSS classes and optional content escaping:

```go
htmlCfg := renderer.HTMLConfig{
	TableClass:     "sales-table",
	HeaderClass:    "table-header",
	BodyClass:      "table-body",
	FooterClass:    "table-footer",
	RowClass:       "table-row",
	HeaderRowClass: "header-row",
	FooterRowClass: "footer-row",
	EscapeContent:  true,
}

table := tablewriter.NewTable(os.Stdout,
	tablewriter.WithRenderer(renderer.NewHTML(htmlCfg)),
)
```

`HTMLConfig` fields:

| Field | Purpose |
|---|---|
| `EscapeContent` | Escape HTML characters in cell content (`&` becomes `&amp;`) |
| `AddLinesTag` | Wrap multiline content in `<lines>` tags |
| `TableClass` | CSS class for `<table>` |
| `HeaderClass` | CSS class for `<thead>` |
| `BodyClass` | CSS class for `<tbody>` |
| `FooterClass` | CSS class for `<tfoot>` |
| `RowClass` | CSS class for `<tr>` in the body |
| `HeaderRowClass` | CSS class for `<tr>` in the header |
| `FooterRowClass` | CSS class for `<tr>` in the footer |

Merged cells become `colspan` attributes, and alignment becomes an inline
`text-align` style. Keep `EscapeContent` on when cell data comes from users
or external input — it is the default-safe choice.

## SVG

`renderer.NewSVG(configs ...SVGConfig) *SVG` renders the table as a
standalone SVG document: rectangles for cells, text elements for content,
and lines for borders. Write it to a file and it opens in any browser or
image viewer:

```go
file, err := os.OpenFile("out.svg", os.O_CREATE|os.O_WRONLY, 0644)
if err != nil {
	ll.Fatal(err)
}
defer file.Close()

table := tablewriter.NewTable(file, tablewriter.WithRenderer(renderer.NewSVG()))
table.Header("Name", "Age", "City")
table.Bulk(data)
table.Render()
```

`SVGConfig` fields:

| Field | Purpose |
|---|---|
| `FontFamily` | Font family, e.g. `"Arial, sans-serif"` |
| `FontSize` | Base font size in SVG units |
| `LineHeightFactor` | Line-height multiplier, e.g. `1.2` |
| `Padding` | Padding inside cells |
| `StrokeWidth` | Line width for borders |
| `StrokeColor` | Stroke color, e.g. `"black"` |
| `HeaderBG` | Header background color |
| `RowBG` | Row background color |
| `RowAltBG` | Alternating row background color |
| `FooterBG` | Footer background color |
| `Debug` | Enable debug output |

Use SVG when a table needs to be an image — embedded in a report, a slide,
or a page that cannot run the library.

## Colorized

`renderer.NewColorized(configs ...ColorizedConfig) *Colorized` renders
ANSI-colored ASCII tables for terminals. Colors are set per section with
`renderer.Tint` values:

```go
colorCfg := renderer.ColorizedConfig{
	Header: renderer.Tint{
		FG: renderer.Colors{color.FgGreen, color.Bold},
		BG: renderer.Colors{color.BgHiWhite},
	},
	Column: renderer.Tint{
		FG: renderer.Colors{color.FgCyan},
		Columns: []renderer.Tint{
			{FG: renderer.Colors{color.FgMagenta}},
			{},
			{FG: renderer.Colors{color.FgHiRed}},
		},
	},
	Footer: renderer.Tint{
		FG: renderer.Colors{color.FgYellow, color.Bold},
	},
	Border:    renderer.Tint{FG: renderer.Colors{color.FgWhite}},
	Separator: renderer.Tint{FG: renderer.Colors{color.FgWhite}},
}

table := tablewriter.NewTable(os.Stdout,
	tablewriter.WithRenderer(renderer.NewColorized(colorCfg)),
)
```

`ColorizedConfig` fields: `Borders` (`tw.Border`), `Settings`
(`tw.Settings`), `Header`, `Column`, `Footer`, `Border`, and `Separator`
(all `renderer.Tint`), and `Symbols` (`tw.Symbols`).

A `Tint` has `FG` and `BG` (`renderer.Colors`) plus a `Columns` slice of
per-column `Tint` overrides — an empty entry inherits the section default.
`Tint.Apply(text)` returns the text wrapped in its color codes, unchanged
when no colors are set.

### 24-bit color

Besides the named `color.Fg*` / `color.Bg*` attributes from
`github.com/fatih/color`, tints accept 24-bit true color. `renderer.RGB`
and `renderer.BgRGB` take red, green, and blue channels (0–255; out-of-range
values are clamped), while `renderer.Hex` and `renderer.BgHex` parse a
`#RRGGBB` or `#RGB` string, with or without the leading `#`. All of them
return a `renderer.Colors`, so they slot in wherever named attributes do:

```go
orange, _ := renderer.Hex("#ff8800")

colorCfg := renderer.ColorizedConfig{
	Header: renderer.Tint{
		FG: append(orange, color.Bold),
		BG: renderer.BgRGB(30, 30, 30),
	},
	Column: renderer.Tint{FG: renderer.RGB(0, 200, 180)},
}
```

Terminals without true-color support may approximate or ignore these
colors.

## Ocean (streaming)

`renderer.NewOcean(oceanConfig ...OceanConfig) *Ocean` is the streaming
renderer: it writes rows as they arrive and keeps column widths stable
while the stream is open. Pair it with `WithStreaming`:

```go
table := tablewriter.NewTable(os.Stdout,
	tablewriter.WithRenderer(renderer.NewOcean()),
	tablewriter.WithStreaming(tw.StreamConfig{Enable: true}),
)
table.Start()
table.Header([]string{"ID", "Description", "Status"})
table.Append([]string{"1", "Short desc", "DONE"})
table.Close()
```

`OceanConfig` is currently an empty struct — the renderer takes no
configuration beyond the table's own `Config`. Use Ocean whenever output
is produced incrementally: progress displays, live feeds, and long-running
jobs. See [Rendering](rendering) for the streaming lifecycle.

## Junction (color helper)

`renderer.NewJunction(ctx JunctionContext) *Junction` renders the junction
points of a table — corners and intersections — with color support. It is
a helper used by the colorized rendering path, not a standalone table
renderer: you do not pass it to `WithRenderer`. `JunctionContext` carries
the `Symbols`, the current `tw.Formatting`, the column index, a logger,
and the `BorderTint` and `SeparatorTint` to apply. Its `GetSegment`
method decides, from the merge state of the current and next rows, whether
a junction renders as a colored line or as empty space. You only need it
if you are building a custom colored renderer and want its junctions to
match the rest of the table.

## Choosing a renderer

| Need | Renderer |
|---|---|
| Terminal output, default look | Blueprint |
| ASCII-only output | Blueprint with `tw.StyleASCII` |
| Documentation or README snippets | Markdown |
| Web pages, styled with CSS | HTML |
| Tables as images | SVG |
| Colored terminal output | Colorized |
| Rows arriving over time | Ocean |
| A format that does not exist yet | Implement `tw.Renderer` |
