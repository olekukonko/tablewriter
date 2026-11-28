---
title: Data Input
description: Feed tables from slices, structs, CSV files, and live streams.
---

# Data Input

TableWriter accepts data in several shapes. The shape you choose depends
on where the data lives: in your code, in a struct, in a CSV file, or
arriving over time. This page covers each one.

## Slices

The most common input is a slice of rows. Rows can be slices of strings or
slices of `any`, and `Header`, `Append`, and `Bulk` all accept them:

```go
data := [][]string{
	{"Alice", "25", "New York"},
	{"Bob", "30", "Boston"},
}

table := tablewriter.NewTable(os.Stdout)
table.Header([]string{"Name", "Age", "City"})
table.Bulk(data)
table.Render()
```

`Append` adds one row at a time and `Bulk` adds many. Both return an
`error`. Mixing them is fine — bulk-add a batch, then append rows as they
arrive.

Values are not limited to strings. `Append` and `Bulk` accept
`[][]any`, and each value renders through its `String()` method when it
has one. A value with a `Format()` method uses `Format()` instead —
`Format()` wins when both exist. This is how the `Age` type in the
Quickstart prints `25 yrs` instead of `25`.

## Structs

Pass a struct to `Append`, or a slice of structs to `Bulk`, and
TableWriter expands each field into a column:

```go
type Employee struct {
	ID         int
	Name       string
	Age        int
	Department string
	Salary     float64
}

func employeeStringer(e interface{}) []string {
	emp, ok := e.(Employee)
	if !ok {
		return []string{"Error: Invalid type"}
	}
	return []string{
		fmt.Sprintf("%d", emp.ID),
		emp.Name,
		fmt.Sprintf("%d", emp.Age),
		emp.Department,
		fmt.Sprintf("%.2f", emp.Salary),
	}
}

employees := []Employee{
	{ID: 1, Name: "Alice Smith", Age: 28, Department: "Engineering", Salary: 75000.50},
	{ID: 2, Name: "Bob Johnson", Age: 34, Department: "Marketing", Salary: 62000.00},
}

table := tablewriter.NewTable(os.Stdout,
	tablewriter.WithRenderer(renderer.NewBlueprint(tw.Rendition{
		Symbols: tw.NewSymbols(tw.StyleRounded),
	})),
	tablewriter.WithStringer(employeeStringer),
	tablewriter.WithConfig(tablewriter.Config{
		Header: tw.CellConfig{
			Formatting: tw.CellFormatting{AutoFormat: tw.On},
			Alignment:  tw.CellAlignment{Global: tw.AlignCenter},
		},
		Row:    tw.CellConfig{Alignment: tw.CellAlignment{Global: tw.AlignLeft}},
		Footer: tw.CellConfig{Alignment: tw.CellAlignment{Global: tw.AlignRight}},
	}),
)
table.Header([]string{"ID", "Name", "Age", "Department", "Salary"})

for _, emp := range employees {
	table.Append(emp)
}
table.Render()
```

### How struct fields become columns

When you `Bulk` a slice of structs, TableWriter reflects over the struct
type: each exported field becomes a column, in declaration order. Embedded
structs are expanded inline, and nil embedded pointers are replaced with
zero structs so the schema stays stable. Headers come from the field names
unless `Behavior.Structs.AutoHeader` is off, and the header names are taken
from struct tags first — the `Tags` list (default `["json", "db"]`) is
checked in order, and the first tag found on a field wins. Unexported
fields and `json:"-"` fields are skipped.

### The tw struct tag

Fields can carry a `tw` tag that controls how the column renders. The
recognized keys are:

| Key | Purpose | Example |
|---|---|---|
| `name` | Override the column header | `` tw:"name=Full Name" `` |
| `align` | Column alignment (`left`, `right`, `center`) | `` tw:"align=right" `` |
| `header_align` | Header alignment for this column | `` tw:"header_align=center" `` |
| `max_width` | Maximum width for the column | `` tw:"max_width=20" `` |
| `pad_left` / `pad_right` | Padding characters | `` tw:"pad_left= " `` |
| `trim_space` | Trim spaces in this column | `` tw:"trim_space=true" `` |
| `trim_tab` | Trim tabs in this column | `` tw:"trim_tab=true" `` |
| `auto_format` | Title-case the header | `` tw:"auto_format=true" `` |
| `wrap` | Wrap mode for the column | `` tw:"wrap=truncate" `` |
| `-` | Skip the field entirely | `` tw:"-" `` |

```go
type Product struct {
	Name  string  `tw:"name=Product,align=left"`
	Price float64 `tw:"name=Price,align=right,max_width=12"`
	Stock int     `tw:"-"`
}
```

### Custom stringers

`WithStringer` takes a function that converts one value into a row of
strings. It is the escape hatch for types whose default rendering is not
what you want — database rows, types from another package, or values that
need formatting logic:

```go
table := tablewriter.NewTable(os.Stdout,
	tablewriter.WithStringer(employeeStringer),
)
```

`WithStringerCache()` caches stringer results so repeated values are only
converted once. `WithStringerCacheCustom` takes a custom cache — a
`twcache.Cache[reflect.Type, reflect.Value]` built with `twcache.NewLRU`
or `twcache.NewLRUEvict`. Caching pays off when the same values appear in
many rows; skip it when rows are unique.

## CSV files

`NewCSV` builds a table directly from a CSV file, and `NewCSVReader` does
the same from an `encoding/csv.Reader` you already configured:

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

The arguments are the writer, the file name (or reader), and `hasHeader` —
pass `true` when the first CSV row is a header. Both constructors accept
the same `With*` options as `NewTable`, so the CSV table is configured
exactly like any other table.

A CSV reader gives you control over the input before TableWriter sees it —
custom delimiters, comment characters, lazy quotes, and ragged rows are all
`encoding/csv.Reader` settings. Build the reader, then hand it over:

```go
f, _ := os.Open("data.csv")
defer f.Close()

r := csv.NewReader(f)
r.Comma = ';'
r.FieldsPerRecord = -1

table, err := tablewriter.NewCSVReader(os.Stdout, r, true)
if err != nil {
	log.Fatal(err)
}
table.Render()
```

## Streaming input

When rows arrive over time, enable streaming and append as they come — the
table prints each row as it arrives instead of buffering everything:

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
	time.Sleep(500 * time.Millisecond)
}
table.Footer([]string{"", "Total", "3"})
```

`tw.StreamConfig.StrictColumns` makes `Append` return an error when a row
has the wrong number of cells; with it off (the default), short rows are
padded and long rows truncated, with a warning log. Streaming pairs with
the Ocean renderer. See [Rendering](rendering) for the full lifecycle.

## Database rows

TableWriter is database-friendly: `sql.Null*` values render natively from
structs, and a query result maps naturally onto `Append` in a loop:

```go
rows, err := db.Query("SELECT name, age, city FROM users")
if err != nil {
	log.Fatal(err)
}
defer rows.Close()

table := tablewriter.NewTable(os.Stdout)
table.Header("Name", "Age", "City")
for rows.Next() {
	var name, city string
	var age int
	if err := rows.Scan(&name, &age, &city); err != nil {
		log.Fatal(err)
	}
	table.Append([]string{name, fmt.Sprint(age), city})
}
table.Render()
```

For repeated queries, define a struct for the row shape and use
`WithStringer` once instead of formatting every row by hand.
