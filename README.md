ASCII Table Writer
=========

[![Build Status](https://travis-ci.org/olekukonko/tablewriter.png?branch=master)](https://travis-ci.org/olekukonko/tablewriter) [![Total views](https://sourcegraph.com/api/repos/github.com/olekukonko/tablewriter/counters/views.png)](https://sourcegraph.com/github.com/olekukonko/tablewriter)

Generate ASCII table on the fly ...  Installation is simple as

    go get  github.com/olekukonko/tablewriter


#### Features
- Automatic Padding
- Support Multiple Lines
- Supports Alignment
- Support Custom Separators
- Automatic Alignment of numbers & percentage
- Write Directly to http , file etc via `io.Reader`
- Read directly from CSV file
- Optional Row line via `SetRowLine` - `new`
- Normalise Table Header - `new`

#### TODO
- ~~Import Directly from CSV~~  - `done`
- Support custom alignment
- ~~Support table with uneven rows~~ - `done`
- General Improvement & Optimisation
- `NewHTML` Parse table from HTML


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

#### Output  2
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

#### Example 2 - CSV
```go
table, _ := tablewriter.NewCSV(os.Stdout, "test_info.csv")
table.SetAlignment(table.ALIGN_LEFT)   // Set Alignment
table.Render()
```

#### Output 2
```
+----------+--------------+------+-----+---------+----------------+
|  FIELD   |     TYPE     | NULL | KEY | DEFAULT |     EXTRA      |
+----------+--------------+------+-----+---------+----------------+
| user_id  | smallint(5)  | NO   | PRI | NULL    | auto_increment |
| username | varchar(10)  | NO   |     | NULL    |                |
| password | varchar(100) | NO   |     | NULL    |                |
+----------+--------------+------+-----+---------+----------------+
```

#### Example 3  - Separator
```go
table, _ := tablewriter.NewCSV(os.Stdout, "test.csv")
table.SetRowLine(true)         // Enable row line

// Change table lines
table.SetCenterSeparator("*")
table.SetColumnSeparator("‡")
table.SetRowSeparator("-")

table.SetAlignment(tablewriter.ALIGN_LEFT)
table.Render()
```

### Output 3
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