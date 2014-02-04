TextTable
=========

[![Build Status](https://travis-ci.org/olekukonko/TextTable.png?branch=master)](https://travis-ci.org/olekukonko/TextTable) [![Total views](https://sourcegraph.com/api/repos/github.com/olekukonko/TextTable/counters/views.png)](https://sourcegraph.com/github.com/olekukonko/TextTable)
ASCII Text Table

#### Features
- Automatic Padding
- Support Multiple Lines
- Supports Alignment
- Support Custom Separators
- Automatic Alignment of numbers & percentage
- Write Directly to http , file etc via `io.Reader`

#### TODO
- Import Directly from CSV
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

t := table.NewTable(os.Stdout)
t.SetHeader([]string{"Name", "Sign", "Rating"})
for _, v := range data {
    t.Append(v)
}
t.Render() // Send output
```

#### output
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
