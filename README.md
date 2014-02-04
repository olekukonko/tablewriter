ASCII Table Writer
=========

[![Build Status](https://travis-ci.org/olekukonko/TextTable.png?branch=master)](https://travis-ci.org/olekukonko/TextTable) [![Total views](https://sourcegraph.com/api/repos/github.com/olekukonko/TextTable/counters/views.png)](https://sourcegraph.com/github.com/olekukonko/TextTable)

Generate ASCII table on the fly ...

#### Features
- Automatic Padding
- Support Multiple Lines
- Supports Alignment
- Support Custom Separators
- Automatic Alignment of numbers & percentage
- Write Directly to http , file etc via `io.Reader`

#### TODO
- <strike>Import Directly from CSV</strike>
- Support custom alignment
- Support table with uneven elements
- Support pyramid structure
- General Improvement & Optimisation

#### Example
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

#### Output
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


#### Example
```go
	table, _ := tablewriter.NewCSV(os.Stdout, "test.csv")
	table.SetCenterSeparator("*")
	table.SetRowSeparator("=")
	table.SetAlignment(ALIGN_LEFT)
	table.Render()
```

#### Output
```
*============*===========*=========*
| FIRST_NAME | LAST_NAME |   SSN   |
*============*===========*=========*
| John       | Barry     | 123456  |
| Kathy      | Smith     | 687987  |
| Bob        | McCornick | 3979870 |
*============*===========*=========*
```
