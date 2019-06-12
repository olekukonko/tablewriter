ASCII Table Writer
=========

[![Build Status](https://travis-ci.org/olekukonko/tablewriter.png?branch=master)](https://travis-ci.org/olekukonko/tablewriter)
[![Total views](https://img.shields.io/sourcegraph/rrc/github.com/olekukonko/tablewriter.svg)](https://sourcegraph.com/github.com/olekukonko/tablewriter)
[![Godoc](https://godoc.org/github.com/olekukonko/tablewriter?status.svg)](https://godoc.org/github.com/olekukonko/tablewriter)

Generate ASCII table on the fly ...  Installation is simple as

    go get github.com/olekukonko/tablewriter


#### Features
- Automatic Padding
- Support Multiple Lines
- Supports Alignment
- Support Custom Separators
- Automatic Alignment of numbers & percentage
- Write directly to http , file etc via `io.Writer`
- Read directly from CSV file
- Optional row line via `SetRowLine`
- Normalise table header
- Make CSV Headers optional
- Enable or disable table border
- Set custom footer support
- Optional identical cells merging
- Set custom caption
- Optional reflowing of paragrpahs in multi-line cells.
- Parse from html.

#### Example   1 - Basic
```go
data := [][]string{
    []string{"A", "The Good", "500"},
    []string{"B", "The Very very Bad Man", "288"},
    []string{"C", "The Ugly", "120"},
    []string{"D", "The Gopher", "800"},
}

table := tablewriter.NewWriter(os.Stdout)
table.SetHeader([]string{"Name", "Sign", "Rating"})

for _, v := range data {
    table.Append(v)
}
table.Render() // Send output
```

##### Output  1
```
+------+-----------------------+--------+
| NAME |         SIGN          | RATING |
+------+-----------------------+--------+
|  A   |       The Good        |    500 |
|  B   | The Very very Bad Man |    288 |
|  C   |       The Ugly        |    120 |
|  D   |      The Gopher       |    800 |
+------+-----------------------+--------+
```

#### Example 2 - Without Border / Footer / Bulk Append
```go
data := [][]string{
    []string{"1/1/2014", "Domain name", "2233", "$10.98"},
    []string{"1/1/2014", "January Hosting", "2233", "$54.95"},
    []string{"1/4/2014", "February Hosting", "2233", "$51.00"},
    []string{"1/4/2014", "February Extra Bandwidth", "2233", "$30.00"},
}

table := tablewriter.NewWriter(os.Stdout)
table.SetHeader([]string{"Date", "Description", "CV2", "Amount"})
table.SetFooter([]string{"", "", "Total", "$146.93"}) // Add Footer
table.SetBorder(false)                                // Set Border to false
table.AppendBulk(data)                                // Add Bulk Data
table.Render()
```

##### Output 2
```

    DATE   |       DESCRIPTION        |  CV2  | AMOUNT
-----------+--------------------------+-------+----------
  1/1/2014 | Domain name              |  2233 | $10.98
  1/1/2014 | January Hosting          |  2233 | $54.95
  1/4/2014 | February Hosting         |  2233 | $51.00
  1/4/2014 | February Extra Bandwidth |  2233 | $30.00
-----------+--------------------------+-------+----------
                                        TOTAL | $146 93
                                      --------+----------

```


#### Example 3 - CSV
```go
table, _ := tablewriter.NewCSV(os.Stdout, "testdata/test_info.csv", true)
table.SetAlignment(tablewriter.ALIGN_LEFT)   // Set Alignment
table.Render()
```

##### Output 3
```
+----------+--------------+------+-----+---------+----------------+
|  FIELD   |     TYPE     | NULL | KEY | DEFAULT |     EXTRA      |
+----------+--------------+------+-----+---------+----------------+
| user_id  | smallint(5)  | NO   | PRI | NULL    | auto_increment |
| username | varchar(10)  | NO   |     | NULL    |                |
| password | varchar(100) | NO   |     | NULL    |                |
+----------+--------------+------+-----+---------+----------------+
```

#### Example 4  - Custom Separator
```go
table, _ := tablewriter.NewCSV(os.Stdout, "testdata/test.csv", true)
table.SetRowLine(true)         // Enable row line

// Change table lines
table.SetCenterSeparator("*")
table.SetColumnSeparator("╪")
table.SetRowSeparator("-")

table.SetAlignment(tablewriter.ALIGN_LEFT)
table.Render()
```

##### Output 4
```
*------------*-----------*---------*
╪ FIRST NAME ╪ LAST NAME ╪   SSN   ╪
*------------*-----------*---------*
╪ John       ╪ Barry     ╪ 123456  ╪
*------------*-----------*---------*
╪ Kathy      ╪ Smith     ╪ 687987  ╪
*------------*-----------*---------*
╪ Bob        ╪ McCornick ╪ 3979870 ╪
*------------*-----------*---------*
```

#### Example 5 - Markdown Format
```go
data := [][]string{
	[]string{"1/1/2014", "Domain name", "2233", "$10.98"},
	[]string{"1/1/2014", "January Hosting", "2233", "$54.95"},
	[]string{"1/4/2014", "February Hosting", "2233", "$51.00"},
	[]string{"1/4/2014", "February Extra Bandwidth", "2233", "$30.00"},
}

table := tablewriter.NewWriter(os.Stdout)
table.SetHeader([]string{"Date", "Description", "CV2", "Amount"})
table.SetBorders(tablewriter.Border{Left: true, Top: false, Right: true, Bottom: false})
table.SetCenterSeparator("|")
table.AppendBulk(data) // Add Bulk Data
table.Render()
```

##### Output 5
```
|   DATE   |       DESCRIPTION        | CV2  | AMOUNT |
|----------|--------------------------|------|--------|
| 1/1/2014 | Domain name              | 2233 | $10.98 |
| 1/1/2014 | January Hosting          | 2233 | $54.95 |
| 1/4/2014 | February Hosting         | 2233 | $51.00 |
| 1/4/2014 | February Extra Bandwidth | 2233 | $30.00 |
```

#### Example 6  - Identical cells merging
```go
data := [][]string{
  []string{"1/1/2014", "Domain name", "1234", "$10.98"},
  []string{"1/1/2014", "January Hosting", "2345", "$54.95"},
  []string{"1/4/2014", "February Hosting", "3456", "$51.00"},
  []string{"1/4/2014", "February Extra Bandwidth", "4567", "$30.00"},
}

table := tablewriter.NewWriter(os.Stdout)
table.SetHeader([]string{"Date", "Description", "CV2", "Amount"})
table.SetFooter([]string{"", "", "Total", "$146.93"})
table.SetAutoMergeCells(true)
table.SetRowLine(true)
table.AppendBulk(data)
table.Render()
```

##### Output 6
```
+----------+--------------------------+-------+---------+
|   DATE   |       DESCRIPTION        |  CV2  | AMOUNT  |
+----------+--------------------------+-------+---------+
| 1/1/2014 | Domain name              |  1234 | $10.98  |
+          +--------------------------+-------+---------+
|          | January Hosting          |  2345 | $54.95  |
+----------+--------------------------+-------+---------+
| 1/4/2014 | February Hosting         |  3456 | $51.00  |
+          +--------------------------+-------+---------+
|          | February Extra Bandwidth |  4567 | $30.00  |
+----------+--------------------------+-------+---------+
|                                       TOTAL | $146 93 |
+----------+--------------------------+-------+---------+
```


#### Example 7  - Table with color
```go
data := [][]string{
	[]string{"1/1/2014", "Domain name", "2233", "$10.98"},
	[]string{"1/1/2014", "January Hosting", "2233", "$54.95"},
	[]string{"1/4/2014", "February Hosting", "2233", "$51.00"},
	[]string{"1/4/2014", "February Extra Bandwidth", "2233", "$30.00"},
}

table := tablewriter.NewWriter(os.Stdout)
table.SetHeader([]string{"Date", "Description", "CV2", "Amount"})
table.SetFooter([]string{"", "", "Total", "$146.93"}) // Add Footer
table.SetBorder(false)                                // Set Border to false

table.SetHeaderColor(tablewriter.Colors{tablewriter.Bold, tablewriter.BgGreenColor},
	tablewriter.Colors{tablewriter.FgHiRedColor, tablewriter.Bold, tablewriter.BgBlackColor},
	tablewriter.Colors{tablewriter.BgRedColor, tablewriter.FgWhiteColor},
	tablewriter.Colors{tablewriter.BgCyanColor, tablewriter.FgWhiteColor})

table.SetColumnColor(tablewriter.Colors{tablewriter.Bold, tablewriter.FgHiBlackColor},
	tablewriter.Colors{tablewriter.Bold, tablewriter.FgHiRedColor},
	tablewriter.Colors{tablewriter.Bold, tablewriter.FgHiBlackColor},
	tablewriter.Colors{tablewriter.Bold, tablewriter.FgBlackColor})

table.SetFooterColor(tablewriter.Colors{}, tablewriter.Colors{},
	tablewriter.Colors{tablewriter.Bold},
	tablewriter.Colors{tablewriter.FgHiRedColor})

table.AppendBulk(data)
table.Render()
```

##### Output 7
![Table with Color](https://cloud.githubusercontent.com/assets/6460392/21101956/bbc7b356-c0a1-11e6-9f36-dba694746efc.png)

#### Example 8 - Set table caption
```go
data := [][]string{
    []string{"A", "The Good", "500"},
    []string{"B", "The Very very Bad Man", "288"},
    []string{"C", "The Ugly", "120"},
    []string{"D", "The Gopher", "800"},
}

table := tablewriter.NewWriter(os.Stdout)
table.SetHeader([]string{"Name", "Sign", "Rating"})
table.SetCaption(true, "Movie ratings.")

for _, v := range data {
    table.Append(v)
}
table.Render() // Send output
```

Note: Caption text will wrap with total width of rendered table.

##### Output 8
```
+------+-----------------------+--------+
| NAME |         SIGN          | RATING |
+------+-----------------------+--------+
|  A   |       The Good        |    500 |
|  B   | The Very very Bad Man |    288 |
|  C   |       The Ugly        |    120 |
|  D   |      The Gopher       |    800 |
+------+-----------------------+--------+
Movie ratings.
```

#### Render table into a string

Instead of rendering the table to `io.Stdout` you can also render it into a string. Go 1.10 introduced the `strings.Builder` type which implements the `io.Writer` interface and can therefore be used for this task. Example:

```go
package main

import (
    "strings"
    "fmt"

    "github.com/olekukonko/tablewriter"
)

func main() {
    tableString := &strings.Builder{}
	table := tablewriter.NewWriter(tableString)

    /*
     * Code to fill the table
     */

    table.Render()

    fmt.Println(tableString.String())
}
```

#### Example 9 - Parse from html
```go
	html := `
<!DOCTYPE html>
<html>
<body>

<table style="width:100%">
  <tr>
    <th>Firstname</th>
    <th>Lastname</th>
    <th>Age</th>
    <td class="print_ignore"></td>
  </tr>
  <tr>
    <td>Jill</td>
    <td>Smith</td>
    <td>50</td>
  </tr>
  <tr>
    <td>Eve</td>
    <td>Jackson</td>
    <td>94</td>
  </tr>
  <tr>
    <td>John</td>
    <td>Doe</td>
    <td>80</td>
  </tr>
</table>

</body>
</html>
`
	table := tablewriter.NewWriter(os.Stdout)
	table.FromHTML(html)
	table.Render() // Send output
```

Note: Caption text will wrap with total width of rendered table.

##### Output 9
```
+-----------+----------+-----+
| FIRSTNAME | LASTNAME | AGE |
+-----------+----------+-----+
| Jill      | Smith    |  50 |
| Eve       | Jackson  |  94 |
| John      | Doe      |  80 |
+-----------+----------+-----+
```

#### TODO
- ~~Import Directly from CSV~~  - `done`
- ~~Support for `SetFooter`~~  - `done`
- ~~Support for `SetBorder`~~  - `done`
- ~~Support table with uneven rows~~ - `done`
- ~~Support custom alignment~~
- ~~`FromHTML` Parse table from HTML~~
- General Improvement & Optimisation
