# Migration Guide: tablewriter v0.0.5 to v1.0.x
>NOTE
> This document is work in progress, use with `caution`
> 
The `tablewriter` library has undergone a substantial redesign between versions **v0.0.5** and **v1.0.x**, evolving from a method-driven API to a modular, configuration-driven framework. This guide provides a comprehensive roadmap for migrating your v0.0.5 codebase to v1.0.x, referencing the provided source files (`tablewriter.go`, `zoo.go`, `stream.go`, `config.go`, `tw/*`, etc.) for accuracy. It includes detailed mappings of old methods to new approaches, practical examples, and explanations of new features like streaming and hierarchical merging.

## Core Philosophy Changes in v1.0.x

Understanding these fundamental shifts is key to grasping the new API:

1. **Configuration-Driven Approach**:
    - **Old**: Used `table.SetXxx()` methods to incrementally modify table properties.
    - **New**: Table behavior and appearance are defined by a `Config` struct (`config.go:Config`) and a `tw.Rendition` struct (`tw/renderer.go:Rendition`) for the renderer, typically set at creation time or via a fluent `ConfigBuilder` (`config.go:ConfigBuilder`). This ensures atomic and predictable configuration.

2. **Decoupled Rendering Engine**:
    - **Old**: Rendering was tightly coupled with the `Table` struct.
    - **New**: The `tw.Renderer` interface (`tw/renderer.go:Renderer`) defines rendering logic, with `renderer.NewBlueprint()` as the default text-based renderer. Renderer appearance (borders, symbols) is managed by `tw.Rendition`, enabling future support for formats like HTML or Markdown.

3. **Unified Section Configuration**:
    - **New**: `tw.CellConfig` (`tw/cell.go:CellConfig`) consistently configures headers, rows, and footers, covering formatting (`tw.CellFormatting`), padding (`tw.CellPadding`), column widths (`tw.CellWidth`), and alignments.

4. **Fluent Builders**:
    - **New**: `tablewriter.NewConfigBuilder()` (`config.go:NewConfigBuilder`) offers a chained API for constructing `Config` objects, with nested builders for `Header()`, `Row()`, `Footer()`, and `ForColumn()`.

5. **Explicit Streaming Mode**:
    - **New**: Streaming (row-by-row rendering) is enabled via `tw.StreamConfig` (`tw/renderer.go:StreamConfig`) and managed with `Table.Start()`, `Table.Append()`/`Row()`, and `Table.Close()` (`stream.go`).

6. **Enhanced Error Handling**:
    - **New**: Methods like `Render()`, `Start()`, `Close()`, and `Append()` return errors, promoting robust applications (`tablewriter.go`, `stream.go`).

7. **Richer Type System**:
    - **New**: Types like `tw.State` (`tw/state.go:State`), `tw.Align` (`tw/types.go:Align`), and `tw.Position` (`tw/types.go:Position`) provide type safety and clarity, replacing magic constants or booleans.

## Detailed Migration Steps

This section maps common v0.0.5 functionalities to their v1.0.x equivalents, with side-by-side examples and references to the source files.

### 1. Table Initialization

The primary constructor has changed from `NewWriter` to `NewTable`, which accepts functional options for configuration.

**Old (v0.0.5):**
```go
import "github.com/olekukonko/tablewriter"
import "os"

table := tablewriter.NewWriter(os.Stdout)
```

**New (v1.0.x):**
```go
import (
    "os"
    "github.com/olekukonko/tablewriter"
    "github.com/olekukonko/tablewriter/renderer"
    "github.com/olekukonko/tablewriter/tw"
)

// Basic creation (default config and renderer)
table := tablewriter.NewTable(os.Stdout)

// With specific configuration
cfg := tablewriter.Config{
    Row: tw.CellConfig{
        Formatting: tw.CellFormatting{Alignment: tw.AlignRight},
    },
}
tableWithConfig := tablewriter.NewTable(os.Stdout, tablewriter.WithConfig(cfg))

// With ConfigBuilder
builder := tablewriter.NewConfigBuilder().
    WithRowAlignment(tw.AlignLeft).
    Header().Formatting().WithAlignment(tw.AlignCenter)
tableWithBuilder := tablewriter.NewTable(os.Stdout, tablewriter.WithConfig(builder.Build()))

// With custom renderer and rendition
rendition := tw.Rendition{
    Symbols: tw.NewSymbols(tw.StyleRounded),
    Borders: tw.Border{Top: tw.On, Bottom: tw.On, Left: tw.On, Right: tw.On},
}
tableWithRenderer := tablewriter.NewTable(os.Stdout,
    tablewriter.WithRenderer(renderer.NewBlueprint(rendition)),
)
```

**Notes:**
- `NewWriter` is deprecated but retained for compatibility (`tablewriter.go:NewWriter`), internally calling `NewTable`.
- `NewTable` (`tablewriter.go:NewTable`) accepts an `io.Writer` and variadic `Option` functions, such as:
    - `WithConfig(Config)`: Applies a `Config` struct (`config.go:WithConfig`).
    - `WithRenderer(tw.Renderer)`: Sets a custom renderer, defaulting to `renderer.NewBlueprint()` (`tablewriter.go:WithRenderer`).
    - `WithStreaming(tw.StreamConfig)`: Enables streaming mode (`tablewriter.go:WithStreaming`).
- The `ConfigBuilder` (`config.go:NewConfigBuilder`) provides a fluent interface for complex setups.

### 2. Data Input

#### 2.1. Setting Headers

Headers can now be set using variadic arguments or slices, with flexible type handling.

**Old (v0.0.5):**
```go
table.SetHeader([]string{"Name", "Sign", "Rating"})
```

**New (v1.0.x):**
```go
// Variadic arguments (preferred for simple headers)
table.Header("Name", "Sign", "Rating")

// Slice of strings
table.Header([]string{"Name", "Sign", "Rating"})

// Slice of any type
table.Header([]any{"Name", 123, "Rating"})
```

**Notes:**
- `Table.Header(elements ...any)` (`tablewriter.go:Header`) processes variadic arguments via `processVariadic` (`zoo.go:processVariadic`).
- Elements are converted to strings using `convertCellsToStrings` and formatted with `prepareContent` based on `Config.Header` settings (`zoo.go:convertCellsToStrings`, `zoo.go:prepareContent`).
- In streaming mode, `Header()` renders immediately via `streamRenderHeader` (`stream.go:streamRenderHeader`).

#### 2.2. Appending Rows

Rows can be added using `Row` (alias for `Append`) for single rows or `Bulk` for multiple rows, supporting diverse data types.

**Old (v0.0.5):**
```go
table.Append([]string{"A", "The Good", "500"})
table.Append([]string{"B", "The Very Bad", "288"})

// Multiple rows
data := [][]string{
    {"C", "The Ugly", "120"},
    {"D", "The Gopher", "800"},
}
table.AppendBulk(data)
```

**New (v1.0.x):**
```go
// Single row (variadic)
table.Row("A", "The Good", 500)
table.Append("B", "The Very Bad", "288") // Alias for Row

// Single row (slice)
table.Append([]any{"A", "The Good", 500})

// Multiple rows
data := [][]any{
    {"C", "The Ugly", 120},
    {"D", "The Gopher", 800},
    {struct{ Name, Role string }{Name: "E", Role: "Admin"}},
}
err := table.Bulk(data)
if err != nil {
    log.Fatalf("Bulk failed: %v", err)
}
```

**Notes:**
- `Table.Append(rows ...interface{})` (`tablewriter.go:Append`) and `Table.Row` (alias) handle:
    - Multiple arguments as cells of a single row.
    - A single slice (`[]string`, `[]any`) as cells of a single row.
    - A single struct, using exported fields or `tw.Formatter`/`fmt.Stringer` (`zoo.go:convertItemToCells`).
- `Table.Bulk(rows interface{})` (`tablewriter.go:Bulk`) processes a slice of rows, where each row is processed by `appendSingle` (`zoo.go:appendSingle`).
- In streaming mode, `Append`/`Row` renders immediately via `streamAppendRow` (`stream.go:streamAppendRow`).
- Data conversion supports basic types, `fmt.Stringer`, `tw.Formatter`, and structs (`zoo.go:convertCellsToStrings`).
- Errors are returned for invalid conversions or streaming issues.

#### 2.3. Setting Footers

Footers follow a similar pattern to headers, with variadic or slice inputs.

**Old (v0.0.5):**
```go
table.SetFooter([]string{"", "", "Total", "1408"})
```

**New (v1.0.x):**
```go
// Variadic
table.Footer("", "", "Total", 1408)

// Slice
table.Footer([]any{"", "", "Total", 1408})
```

**Notes:**
- `Table.Footer(elements ...any)` (`tablewriter.go:Footer`) uses `processVariadic` and `convertCellsToStrings` (`zoo.go`).
- In streaming mode, `Footer` buffers data via `streamStoreFooter`, rendered by `Close` (`stream.go:streamStoreFooter`, `stream.go:streamRenderFooter`).
- Formatted using `Config.Footer` settings (`zoo.go:prepareTableSection`).

### 3. Rendering the Table

Rendering now supports both batch and streaming modes, with error handling.

**Old (v0.0.5):**
```go
table.Render()
```

**New (v1.0.x) - Batch Mode:**
```go
err := table.Render()
if err != nil {
    log.Fatalf("Render failed: %v", err)
}
```

**New (v1.0.x) - Streaming Mode:**
```go
table := tablewriter.NewTable(os.Stdout,
    tablewriter.WithStreaming(tw.StreamConfig{
        Enable: true,
        Widths: tw.CellWidth{Global: 12}, // Optional: fixed column widths
    }),
)
if err := table.Start(); err != nil {
    log.Fatalf("Start failed: %v", err)
}
table.Header("Col1", "Col2")
table.Row("Data1", "Data2")
if err := table.Close(); err != nil {
    log.Fatalf("Close failed: %v", err)
}
```

**Notes:**
- **Batch Mode**:
    - `Render` (`tablewriter.go:render`) processes all data, calculating widths (`zoo.go:prepareContexts`) and rendering via `renderHeader`, `renderRow`, `renderFooter` (`tablewriter.go`).
    - Calls `renderer.Start()` and `renderer.Close()` (`tw/renderer.go:Renderer`).
    - Returns errors for invalid configurations or I/O issues.
- **Streaming Mode**:
    - Enabled with `WithStreaming(tw.StreamConfig{Enable: true})` (`tablewriter.go:WithStreaming`).
    - `Start` initializes the stream, fixing column widths (`stream.go:Start`, `stream.go:streamCalculateWidths`).
    - `Header`, `Append`/`Row`, and `Footer` render or buffer data immediately (`stream.go`).
    - `Close` renders the footer and finalizes the stream (`stream.go:Close`).
    - All streaming methods return errors.

### 4. Styling and Appearance Configuration

Styling has shifted from setter methods to `tw.Rendition` (renderer settings) and `Config` (content processing).

#### 4.1. Borders and Separator Lines

**Old (v0.0.5):**
```go
table.SetBorder(false)
table.SetRowLine(true)
table.SetHeaderLine(true)
```

**New (v1.0.x):**
```go
rendition := tw.Rendition{
    Borders: tw.Border{Left: tw.Off, Right: tw.Off, Top: tw.Off, Bottom: tw.Off},
    Settings: tw.Settings{
        Lines: tw.Lines{ShowHeaderLine: tw.On, ShowFooterLine: tw.On},
        Separators: tw.Separators{BetweenRows: tw.On, BetweenColumns: tw.On},
    },
}
table := tablewriter.NewTable(os.Stdout,
    tablewriter.WithRenderer(renderer.NewBlueprint(rendition)),
)
```

**Notes:**
- Deprecated methods (`SetBorder`, `SetRowLine`, `SetHeaderLine`) are replaced by `tw.Rendition` (`deprecated.go:WithBorders`).
- `tw.Rendition.Borders` controls outer borders (`tw/renderer.go:Border`).
- `tw.Rendition.Settings.Lines` manages major separators (`ShowHeaderLine`, `ShowFooterLine`) (`tw/renderer.go:Lines`).
- `tw.Rendition.Settings.Separators` controls row/column separators (`tw/renderer.go:Separators`).
- Predefined constants like `tw.BorderNone` and `tw.LinesNone` simplify minimal configurations (`tw/renderer.go`).

#### 4.2. Separator Characters (Symbols)

**Old (v0.0.5):**
```go
table.SetCenterSeparator("*")
table.SetColumnSeparator("!")
table.SetRowSeparator("=")
```

**New (v1.0.x):**
```go
// Predefined style
rendition := tw.Rendition{
    Symbols: tw.NewSymbols(tw.StyleRounded),
}
table := tablewriter.NewTable(os.Stdout,
    tablewriter.WithRenderer(renderer.NewBlueprint(rendition)),
)

// Custom symbols
mySymbols := tw.NewSymbolCustom("my-style").
    WithCenter("*").
    WithColumn("!").
    WithRow("=").
    WithTopLeft("/").
	WithTopMid("-").
	WithTopRight("\\").
    WithMidLeft("[").
	WithMidRight("]").
    WithBottomLeft("\\").
	WithBottomMid("_").
	WithBottomRight("/")

renditionCustom := tw.Rendition{Symbols: mySymbols}
tableCustom := tablewriter.NewTable(os.Stdout,tablewriter.WithRenderer(renderer.NewBlueprint(renditionCustom)))
```

**Notes:**
- Old separator methods are deprecated (`deprecated.go`).
- `tw.Rendition.Symbols` uses the `tw.Symbols` interface (`tw/symbols.go:Symbols`).
- `tw.NewSymbols(tw.BorderStyle)` provides predefined styles (e.g., `StyleASCII`, `StyleRounded`) (`tw/symbols.go:NewSymbols`).
- `tw.SymbolCustom` allows fully custom symbols (`tw/symbols.go:SymbolCustom`).

#### 4.3. Alignment

**Old (v0.0.5):**
```go
table.SetAlignment(tablewriter.ALIGN_CENTER)
table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
```

**New (v1.0.x):**
```go
// ConfigBuilder
builder := tablewriter.NewConfigBuilder().
    WithRowAlignment(tw.AlignCenter).
    Header().Formatting().
	WithAlignment(tw.AlignLeft).
    Footer().Formatting().
	WithAlignment(tw.AlignRight).
    ForColumn(0).
	WithAlignment(tw.AlignLeft)
table := tablewriter.NewTable(os.Stdout, tablewriter.WithConfig(builder.Build()))

// Direct Config
cfg := tablewriter.Config{
    Header: tw.CellConfig{
        Formatting: tw.CellFormatting{Alignment: tw.AlignLeft},
        ColumnAligns: []tw.Align{tw.AlignLeft, tw.AlignCenter},
    },
    Row: tw.CellConfig{
        Formatting: tw.CellFormatting{Alignment: tw.AlignCenter},
    },
}
tableDirect := tablewriter.NewTable(os.Stdout, tablewriter.WithConfig(cfg))

// Option shortcut
tableOpt := tablewriter.NewTable(os.Stdout,
    tablewriter.WithHeaderAlignment(tw.AlignLeft),
    tablewriter.WithAlignment(tw.AlignCenter),
)
```

**Notes:**
- Old `ALIGN_XXX` constants are replaced by `tw.AlignLeft`, `tw.AlignCenter`, etc. (`tw/tw.go`).
- Global alignment is set via `Config.<Section>.Formatting.Alignment` (`tw/cell.go:CellFormatting`).
- Per-column alignment uses `Config.<Section>.ColumnAligns` (`tw/cell.go:CellConfig`).
- `ConfigBuilder.ForColumn(idx).WithAlignment()` sets header alignment (`config.go:ColumnConfigBuilder`).
- `WithAlignment` sets `ColumnAligns` for all sections (`tablewriter.go:WithAlignment`).

#### 4.4. Auto-Formatting (Headers)

**Old (v0.0.5):**
```go
table.SetAutoFormatHeaders(true)
```

**New (v1.0.x):**
```go
// ConfigBuilder
builder := tablewriter.NewConfigBuilder().
    WithHeaderAutoFormat(tw.On)
table := tablewriter.NewTable(os.Stdout, tablewriter.WithConfig(builder.Build()))

// Direct Config
cfg := tablewriter.Config{
    Header: tw.CellConfig{
        Formatting: tw.CellFormatting{AutoFormat: tw.On},
    },
}
tableDirect := tablewriter.NewTable(os.Stdout, tablewriter.WithConfig(cfg))

// Option
tableOpt := tablewriter.NewTable(os.Stdout, tablewriter.WithHeaderAutoFormat(tw.On))
```

**Notes:**
- `AutoFormat` is a `tw.State` (`tw/state.go:State`) in `tw.CellFormatting` (`tw/cell.go`).
- Defaults: `Header.AutoFormat = tw.On`, `Row/Footer.AutoFormat = tw.Off` (`config.go:defaultConfig`).
- Applies `tw.Title` (uppercase, underscore/dot to space) (`zoo.go:prepareContent`, `tw/fn.go:Title`).

#### 4.5. Text Wrapping

**Old (v0.0.5):**
```go
table.SetAutoWrapText(true) // Normal wrap
```

**New (v1.0.x):**
```go
// ConfigBuilder
builder := tablewriter.NewConfigBuilder().
    WithRowAutoWrap(tw.WrapNone).
    WithHeaderAutoWrap(tw.WrapTruncate)
table := tablewriter.NewTable(os.Stdout, tablewriter.WithConfig(builder.Build()))

// Direct Config
cfg := tablewriter.Config{
    Row: tw.CellConfig{
        Formatting: tw.CellFormatting{AutoWrap: tw.WrapNone},
    },
    Header: tw.CellConfig{
        Formatting: tw.CellFormatting{AutoWrap: tw.WrapTruncate},
    },
}
tableDirect := tablewriter.NewTable(os.Stdout, tablewriter.WithConfig(cfg))
```

**Notes:**
- `AutoWrap` uses constants like `tw.WrapNone`, `tw.WrapNormal`, `tw.WrapTruncate`, `tw.WrapBreak` (`tw/tw.go`).
- Defaults: `Header = tw.WrapTruncate`, `Row/Footer = tw.WrapNormal` (`config.go:defaultConfig`).
- Wrapping depends on `ColMaxWidths` or `Config.MaxWidth` (`zoo.go:calculateContentMaxWidth`, `zoo.go:prepareContent`).

#### 4.6. Padding

**Old (v0.0.5):**
```go
table.SetTablePadding("\t")
```

**New (v1.0.x):**
```go
// ConfigBuilder
builder := tablewriter.NewConfigBuilder().
    WithRowGlobalPadding(tw.Padding{Left: " ", Right: " ", Top: "~"}).
    Header().Padding().WithGlobal(tw.Padding{Left: ">", Right: "<"}).
    Header().Padding().AddColumnPadding(tw.Padding{Left: ">>", Right: "<<"}) // Column 0
table := tablewriter.NewTable(os.Stdout, tablewriter.WithConfig(builder.Build()))

// Direct Config
cfg := tablewriter.Config{
    Header: tw.CellConfig{
        Padding: tw.CellPadding{Global: tw.Padding{Left: "[", Right: "]"}},
    },
    Row: tw.CellConfig{
        Padding: tw.CellPadding{Global: tw.Padding{Left: " ", Right: " "}},
    },
}
tableDirect := tablewriter.NewTable(os.Stdout, tablewriter.WithConfig(cfg))

// Option
tableOpt := tablewriter.NewTable(os.Stdout,
    tablewriter.WithPadding(tw.Padding{Left: " ", Right: " "}),
)
```

**Notes:**
- `tw.Padding` (`tw/types.go:Padding`) defines `Left`, `Right`, `Top`, `Bottom` strings.
- `tw.CellPadding` (`tw/cell.go:CellPadding`) includes `Global` and `PerColumn` settings.
- `tw.PaddingNone` removes padding (`tw/renderer.go:PaddingNone`).
- Padding affects width calculations (`zoo.go:updateWidths`).

#### 4.7. Column Widths (Max Width)

**Old (v0.0.5):**
```go
// Limited control, possibly via SetColMaxWidth (if available)
```

**New (v1.0.x):**
```go
// ConfigBuilder
builder := tablewriter.NewConfigBuilder().
    WithMaxWidth(80).
    WithRowMaxWidth(15).
    ForColumn(0).WithMaxWidth(10)
table := tablewriter.NewTable(os.Stdout, tablewriter.WithConfig(builder.Build()))

// Direct Config
cfg := tablewriter.Config{
    MaxWidth: 80,
    Header: tw.CellConfig{
        ColMaxWidths: tw.CellWidth{
            Global:    20,
            PerColumn: tw.NewMapper[int, int]().Set(0, 10),
        },
    },
    Row: tw.CellConfig{
        ColMaxWidths: tw.CellWidth{Global: 15},
    },
}
tableDirect := tablewriter.NewTable(os.Stdout, tablewriter.WithConfig(cfg))

// Streaming mode
streamCfg := tw.StreamConfig{
    Enable: true,
    Widths: tw.CellWidth{
        Global:    12,
        PerColumn: tw.NewMapper[int, int]().Set(0, 20),
    },
}
tableStream := tablewriter.NewTable(os.Stdout, tablewriter.WithStreaming(streamCfg))
```

**Notes:**
- **Batch Mode**:
    - `Config.MaxWidth` constrains total table width (`zoo.go:prepareContent`).
    - `Config.<Section>.ColMaxWidths.Global/PerColumn` sets content width limits (`tw/cell.go:CellWidth`).
    - Processed by `calculateContentMaxWidth` (`zoo.go:calculateContentMaxWidth`).
- **Streaming Mode**:
    - `StreamConfig.Widths.Global/PerColumn` fixes total cell widths (content + padding) (`stream.go:streamCalculateWidths`).
    - Set via `WithColumnMax` or `WithColumnWidths` (`tablewriter.go`).

#### 4.8. Colors

**Old (v0.0.5):**
```go
table.SetColumnColor(
    tablewriter.Colors{tablewriter.Bold, tablewriter.FgGreenColor},
    tablewriter.Colors{tablewriter.FgRedColor},
)
```

**New (v1.0.x):**
```go
import "fmt"

const (
    Reset   = "\033[0m"
    Bold    = "\033[1m"
    FgGreen = "\033[32m"
    FgRed   = "\033[31m"
)

func gray(value string) string {
    return fmt.Sprintf("\u001B[38;5;240m%s\u001B[0m", value)
}

func main() {
    data := [][]any{
        {"String", "NULL"},
        {"Invalid", gray("NULL")},
    }
    table := tablewriter.NewTable(os.Stdout)
    table.Bulk(data)
    table.Render()
}

// Using tw.Formatter
type Status string

func (s Status) Format() string {
    color := FgGreen
    if s == "Inactive" {
        color = FgRed
    }
    return color + string(s) + Reset
}

table.Append("Bob", Status("Inactive"))

// Using per-column filter
statusFilter := func(cell string) string {
    if cell == "Ready" {
        return FgGreen + cell + Reset
    }
    return FgRed + cell + Reset
}
cfg := tablewriter.Config{
    Row: tw.CellConfig{
        Filter: tw.CellFilter{
            PerColumn: []func(string) string{statusFilter},
        },
    },
}
tableWithFilters := tablewriter.NewTable(os.Stdout, tablewriter.WithConfig(cfg))
tableWithFilters.Append("Node1", "Ready")
```

**Notes:**
- Colors are applied via ANSI escape codes in cell content (`tw/fn.go:TruncateString`, `tw/fn.go:DisplayWidth`).
- Use `tw.Formatter` or filters (`tw/cell.go:CellFilter`) for programmatic styling (`zoo.go:convertCellsToStrings`).

### 5. Cell Merging

**Old (v0.0.5):**
```go
table.SetAutoMergeCells(true) // Horizontal merging
```

**New (v1.0.x):**
```go
// Horizontal merging
cfgHorizontal := tablewriter.Config{
    Row: tw.CellConfig{
        Formatting: tw.CellFormatting{MergeMode: tw.MergeHorizontal},
    },
}
tableH := tablewriter.NewTable(os.Stdout, tablewriter.WithConfig(cfgHorizontal))
tableH.Append("A", "B", "B", "C") // "B" cells merge

// Vertical merging
cfgVertical := tablewriter.Config{
    Row: tw.CellConfig{
        Formatting: tw.CellFormatting{MergeMode: tw.MergeVertical},
    },
}
tableV := tablewriter.NewTable(os.Stdout, tablewriter.WithConfig(cfgVertical))
tableV.Append("A", "X")
tableV.Append("A", "Y") // "A" merges vertically

// Hierarchical merging (see example below)
```

**Notes:**
- `MergeMode` supports `tw.MergeHorizontal`, `tw.MergeVertical`, `tw.MergeHierarchical` (`tw/tw.go`).
- Handled by `prepareWithMerges` (horizontal), `applyVerticalMerges`, and `applyHierarchicalMerges` (`zoo.go`).
- `applyHorizontalMergeWidths` adjusts column widths (`zoo.go`).
- Renderer uses `MergeState` in `CellContext` for correct border drawing (`tw/renderer.go:CellContext`).

### 6. Caption

**Old (v0.0.5):**
```go
// Limited or no direct support
```

**New (v1.0.x):**
```go
table.Caption(tw.Caption{
    Text:  "Table Caption",
    Spot:  tw.SpotTopCenter,
    Align: tw.AlignCenter,
    Width: 50,
})
```

**Notes:**
- `Table.Caption(tw.Caption)` (`tablewriter.go:Caption`) uses `tw.Caption` (`tw/types.go:Caption`).
- `Spot` defaults to `SpotBottomCenter` if `SpotNone` (`tablewriter.go:Caption`).
- Rendered via `printTopBottomCaption` (`tablewriter.go:printTopBottomCaption`).

### 7. Miscellaneous Methods

- **Clear() → Reset()**:
    - **Old**: `Clear()` cleared data.
    - **New**: `Table.Reset()` clears data and rendering state (`tablewriter.go:Reset`).
- **SetDebug(bool) → WithDebug(bool)**:
    - **New**: Enable via `WithDebug(true)` or `Config.Debug = true` (`config.go:WithDebug`).
    - Logs accessed via `table.Debug()` (`tablewriter.go:Debug`).
- **SetNoWhiteSpace(bool), SetTablePadding(string)**:
    - **New**: Use `tw.Rendition` and `tw.CellPadding` for kubectl-style output (see Section 7 in original guide).

### 8. Adopting New Features

- **Fluent Builders**: Use `NewConfigBuilder()` for readable configuration (`config.go`).
- **Advanced Padding**: `tw.Padding` supports top/bottom padding as extra lines (`tw/types.go:Padding`).
- **Per-Column Configuration**: `ConfigBuilder.ForColumn()` for alignment/widths (`config.go:ColumnConfigBuilder`).
- **Renderer Customization**: Implement `tw.Renderer` or customize `tw.Rendition` (`tw/renderer.go`).
- **Filters**: `tw.CellFilter` for data transformation (`tw/cell.go:CellFilter`).
- **Enhanced Symbols**: Predefined styles via `tw.NewSymbols` (`tw/symbols.go`).

## Example: Hierarchical Merging

```go
package main

import (
	"github.com/olekukonko/tablewriter"
	"github.com/olekukonko/tablewriter/renderer"
	"github.com/olekukonko/tablewriter/tw"
	"os"
)

func main() {
	data := [][]string{
		{"Package", "Version", "Status"},
		{"table\nwriter", "v0.0.1", "legacy"},
		{"table\nwriter", "v0.0.2", "legacy"},
		{"table\nwriter", "v0.0.2", "legacy"},
		{"table\nwriter", "v0.0.2", "legacy"},
		{"table\nwriter", "v0.0.5", "legacy"},
		{"table\nwriter", "v1.0.6", "latest"},
	}

	r := tablewriter.WithRenderer(renderer.NewBlueprint(tw.Rendition{
		Settings: tw.Settings{
			Separators: tw.Separators{BetweenRows: tw.On},
			Lines:      tw.Lines{ShowFooterLine: tw.On},
		},
	}))
	table := tablewriter.NewTable(os.Stdout, r,
		tablewriter.WithConfig(tablewriter.Config{
			Row: tw.CellConfig{
				Formatting: tw.CellFormatting{
					MergeMode:  tw.MergeHierarchical,
					Alignment:  tw.AlignCenter,
					AutoWrap:   tw.WrapNone,
					AutoFormat: tw.Off,
				},
			},
		}),
	)

	table.Header(data[0])
	table.Bulk(data[1:])
	table.Render()

	table = tablewriter.NewTable(os.Stdout, r,
		tablewriter.WithConfig(tablewriter.Config{
			Row: tw.CellConfig{
				Formatting: tw.CellFormatting{
					MergeMode: tw.MergeVertical,
					Alignment: tw.AlignCenter,
					AutoWrap:  tw.WrapNone,
				},
			},
		}),
	)

	table.Header(data[0])
	table.Bulk(data[1:])
	table.Render()

}
```

**Output:**
```
┌─────────┬─────────┬────────┐
│ PACKAGE │ VERSION │ STATUS │
├─────────┼─────────┼────────┤
│  table  │ v0.0.1  │ legacy │
│ writer  │         │        │
│         ├─────────┼────────┤
│         │ v0.0.2  │ legacy │
│         │         │        │
│         │         │        │
│         │         │        │
│         │         │        │
│         ├─────────┼────────┤
│         │ v0.0.5  │ legacy │
│         ├─────────┼────────┤
│         │ v1.0.6  │ latest │
└─────────┴─────────┴────────┘
┌─────────┬─────────┬────────┐
│ PACKAGE │ VERSION │ STATUS │
├─────────┼─────────┼────────┤
│  table  │ v0.0.1  │ legacy │
│ writer  │         │        │
│         ├─────────┤        │
│         │ v0.0.2  │        │
│         │         │        │
│         │         │        │
│         │         │        │
│         │         │        │
│         ├─────────┤        │
│         │ v0.0.5  │        │
│         ├─────────┼────────┤
│         │ v1.0.6  │ latest │
└─────────┴─────────┴────────┘

```

## Example: Kubectl-Style Output

```go
package main

import (
    "os"
    "sync"
    "github.com/olekukonko/tablewriter"
    "github.com/olekukonko/tablewriter/renderer"
    "github.com/olekukonko/tablewriter/tw"
)

var wg sync.WaitGroup

func main() {
    data := [][]any{
        {"node1.example.com", "Ready", "compute", "1.11"},
        {"node2.example.com", "Ready", "compute", "1.11"},
        {"node3.example.com", "Ready", "compute", "1.11"},
        {"node4.example.com", "NotReady", "compute", "1.11"},
    }

    table := tablewriter.NewTable(os.Stdout,
        tablewriter.WithRenderer(renderer.NewBlueprint(tw.Rendition{
            Borders: tw.BorderNone,
            Settings: tw.Settings{
                Separators: tw.SeparatorsNone,
                Lines:      tw.LinesNone,
            },
        })),
        tablewriter.WithConfig(tablewriter.Config{
            Header: tw.CellConfig{
                Formatting: tw.CellFormatting{Alignment: tw.AlignLeft},
            },
            Row: tw.CellConfig{
                Formatting: tw.CellFormatting{Alignment: tw.AlignLeft},
                Padding:    tw.CellPadding{Global: tw.PaddingNone},
            },
        }),
    )
    table.Header("Name", "Status", "Role", "Version")
    table.Bulk(data)
    table.Render()
}
```

**Output:**
```
NAME               STATUS    ROLE     VERSION
node1.example.com  Ready     compute  1.11
node2.example.com  Ready     compute  1.11
node3.example.com  Ready     compute  1.11
node4.example.com  NotReady  compute  1.11
```

## Troubleshooting & Common Pitfalls

- **Error Handling**: Check errors from `Render`, `Start`, `Close`, `Append`, and `Bulk` (`tablewriter.go`, `stream.go`).
- **Config vs. Rendition**:
    - `Config`: Manages data processing, formatting, padding, and merging (`config.go:Config`).
    - `Rendition`: Controls visual rendering (borders, symbols) (`tw/renderer.go:Rendition`).
- **Streaming Widths**: Fixed at `Start` based on initial data or `StreamConfig.Widths` (`stream.go:streamCalculateWidths`).
- **Defaults**: `defaultConfig` sets `Header.AutoFormat = tw.On`, `Row.AutoWrap = tw.WrapNormal` (`config.go:defaultConfig`).
- **Alignment Precedence**: `ColumnAligns` overrides `Formatting.Alignment` (`tw/cell.go:CellConfig`).
- **Padding and Widths**: Padding is included in column widths; `ColMaxWidths` applies to content (`zoo.go:updateWidths`).

## Additional Notes

- **Performance**: Use `WithStringerCache` for custom stringers (`tablewriter.go:WithStringerCache`).
- **Debugging**: Enable with `WithDebug(true)` and access via `table.Debug()` (`config.go:WithDebug`).
- **Backward Compatibility**: Deprecated methods like `NewWriter` are retained but should be replaced (`tablewriter.go:NewWriter`, `deprecated.go`).

For further details, consult the library documentation or source files (`tablewriter.go`, `zoo.go`, `stream.go`, `tests/*`).
