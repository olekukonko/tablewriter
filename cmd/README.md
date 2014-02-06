ASCII Table Writer Tool
=========

Generate ASCII table on the fly via command line ...  Installation is simple as

#### Get Tool

    go get  github.com/olekukonko/tablewriter/cmd

#### Install Tool

    go install  github.com/olekukonko/tablewriter/cmd


#### Usage

    go run tablewriter.go -f ../test.csv

#### Output

+------------+-----------+---------+
| FIRST NAME | LAST NAME |   SSN   |
+------------+-----------+---------+
|    John    |   Barry   |  123456 |
|   Kathy    |   Smith   |  687987 |
|    Bob     | McCornick | 3979870 |
+------------+-----------+---------+
