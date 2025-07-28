package tests

import (
	"bytes"
	"fmt"
	"github.com/olekukonko/tablewriter"
	"github.com/olekukonko/tablewriter/renderer"
	"github.com/olekukonko/tablewriter/tw"
	"testing"
)

// Employee represents a struct for employee data, simulating a database record.
type Employee struct {
	ID         int
	Name       string
	Age        int
	Department string
	Salary     float64
}

// dummyDatabase simulates a database with employee records.
type dummyDatabase struct {
	records []Employee
}

// fetchEmployees simulates fetching data from a database.
func (db *dummyDatabase) fetchEmployees() []Employee {
	return db.records
}

// employeeStringer converts an Employee struct to a slice of strings for table rendering.
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

// TestStructTableWithDB tests rendering a table from struct data fetched from a dummy database.
func TestStructTableWithDB(t *testing.T) {
	// Initialize dummy database with sample data
	db := &dummyDatabase{
		records: []Employee{
			{ID: 1, Name: "Alice Smith", Age: 28, Department: "Engineering", Salary: 75000.50},
			{ID: 2, Name: "Bob Johnson", Age: 34, Department: "Marketing", Salary: 62000.00},
			{ID: 3, Name: "Charlie Brown", Age: 45, Department: "HR", Salary: 80000.75},
		},
	}

	// Configure table with custom settings
	config := tablewriter.Config{
		Header: tw.CellConfig{
			Formatting: tw.CellFormatting{
				AutoFormat: tw.On,
			},
			Alignment: tw.CellAlignment{Global: tw.AlignCenter},
		},
		Row: tw.CellConfig{
			Alignment: tw.CellAlignment{Global: tw.AlignLeft},
		},
		Footer: tw.CellConfig{
			Alignment: tw.CellAlignment{Global: tw.AlignRight},
		},
	}

	// Create table with buffer and custom renderer
	var buf bytes.Buffer
	table := tablewriter.NewTable(&buf,
		tablewriter.WithConfig(config),
		tablewriter.WithRenderer(renderer.NewBlueprint(tw.Rendition{
			Symbols: tw.NewSymbols(tw.StyleRounded), // Use rounded Unicode style
			Settings: tw.Settings{
				Separators: tw.Separators{
					BetweenColumns: tw.On,
					BetweenRows:    tw.Off,
				},
				Lines: tw.Lines{
					ShowHeaderLine: tw.On,
				},
			},
		})),
		tablewriter.WithStringer(employeeStringer),
	)

	// Set the stringer for converting Employee structs

	// Set header
	table.Header([]string{"ID", "Name", "Age", "Department", "Salary"})

	// Fetch data from "database" and append to table
	employees := db.fetchEmployees()
	for _, emp := range employees {
		if err := table.Append(emp); err != nil {
			t.Fatalf("Failed to append employee: %v", err)
		}
	}

	// Add a footer with a total salary
	totalSalary := 0.0
	for _, emp := range employees {
		totalSalary += emp.Salary
	}
	table.Footer([]string{"", "", "", "Total", fmt.Sprintf("%.2f", totalSalary)})

	// Render the table
	if err := table.Render(); err != nil {
		t.Fatalf("Failed to render table: %v", err)
	}

	// Expected output
	expected := `
        ╭────┬───────────────┬─────┬─────────────┬───────────╮
        │ ID │     NAME      │ AGE │ DEPARTMENT  │  SALARY   │
        ├────┼───────────────┼─────┼─────────────┼───────────┤
        │ 1  │ Alice Smith   │ 28  │ Engineering │ 75000.50  │
        │ 2  │ Bob Johnson   │ 34  │ Marketing   │ 62000.00  │
        │ 3  │ Charlie Brown │ 45  │ HR          │ 80000.75  │
        ├────┼───────────────┼─────┼─────────────┼───────────┤
        │    │               │     │       Total │ 217001.25 │
        ╰────┴───────────────┴─────┴─────────────┴───────────╯
`

	// Visual check
	if !visualCheck(t, "StructTableWithDB", buf.String(), expected) {
		t.Log(table.Debug())
	}
}

func TestAutoHeaderScenarios(t *testing.T) {
	type Basic struct {
		Foo int
		Bar string
		baz bool // unexported
	}

	type WithTags struct {
		ID   int    `json:"id"`
		Name string `json:"name,omitempty"`
		Age  int    `json:"-"`
		City string `json:"location"`
	}

	type Omitted struct {
		SkipMe string `json:"-"`
		KeepMe string
	}

	type NoTags struct {
		Field1 string
		field2 int // unexported
	}

	type Embedded struct {
		WithTags
		Extra string
	}

	type PointerTest struct {
		Value string
	}

	tests := []struct {
		name       string
		data       interface{}
		enable     bool
		preHeaders []string
		expected   string
	}{
		{
			name:   "BasicStruct",
			data:   []Basic{{1, "test", true}, {2, "test2", false}},
			enable: true,
			expected: `
			┌─────┬───────┐
			│ FOO │  BAR  │
			├─────┼───────┤
			│ 1   │ test  │
			│ 2   │ test2 │
			└─────┴───────┘

`,
		},
		{
			name:   "WithTags",
			data:   []WithTags{{1, "John", 30, "NY"}, {2, "", 0, "LA"}},
			enable: true,
			expected: `
			┌────┬──────┬──────────┐
			│ ID │ NAME │ LOCATION │
			├────┼──────┼──────────┤
			│ 1  │ John │ NY       │
			│ 2  │      │ LA       │
			└────┴──────┴──────────┘
`,
		},
		{
			name:   "OmittedFields",
			data:   []Omitted{{"skip", "keep"}, {"skip2", "keep2"}},
			enable: true,
			expected: `
			┌────────┐
			│ KEEPME │
			├────────┤
			│ keep   │
			│ keep2  │
			└────────┘

`,
		},
		{
			name:   "NoTags",
			data:   []NoTags{{"val1", 42}, {"val2", 43}},
			enable: true,
			expected: `
			┌─────────┐
			│ FIELD 1 │
			├─────────┤
			│ val1    │
			│ val2    │
			└─────────┘

`,
		},
		{
			name:   "Embedded",
			data:   []Embedded{{WithTags{1, "John", 30, "NY"}, "Extra"}, {WithTags{2, "Doe", 25, "LA"}, "Value"}},
			enable: true,
			expected: `
		┌────┬──────┬──────────┬───────┐
		│ ID │ NAME │ LOCATION │ EXTRA │
		├────┼──────┼──────────┼───────┤
		│ 1  │ John │ NY       │ Extra │
		│ 2  │ Doe  │ LA       │ Value │
		└────┴──────┴──────────┴───────┘
`,
		},
		{
			name:   "PointerToStruct",
			data:   []*PointerTest{{"Value1"}, {"Value2"}},
			enable: true,
			expected: `
			┌────────┐
			│ VALUE  │
			├────────┤
			│ Value1 │
			│ Value2 │
			└────────┘
`,
		},
		{
			name:   "SliceOfPointers",
			data:   []*WithTags{{1, "John", 30, "NY"}, {2, "Doe", 25, "LA"}},
			enable: true,
			expected: `
			┌────┬──────┬──────────┐
			│ ID │ NAME │ LOCATION │
			├────┼──────┼──────────┤
			│ 1  │ John │ NY       │
			│ 2  │ Doe  │ LA       │
			└────┴──────┴──────────┘
`,
		},
		{
			name:   "NonStruct",
			data:   [][]string{{"A", "B"}, {"C", "D"}},
			enable: false,
			expected: `
┌───┬───┐
│ A │ B │
│ C │ D │
└───┴───┘
`,
		},
		{
			name:     "EmptySlice",
			data:     []WithTags{},
			expected: ``,
		},
		{
			name:   "enabled",
			data:   []WithTags{{1, "John", 30, "NY"}},
			enable: true,
			expected: `
			┌────┬──────┬──────────┐
			│ ID │ NAME │ LOCATION │
			├────┼──────┼──────────┤
			│ 1  │ John │ NY       │
			└────┴──────┴──────────┘


`, // No header, falls back to string reps
		},
		{
			name:   "Disabled",
			data:   []WithTags{{1, "John", 30, "NY"}},
			enable: false,
			expected: `
			┌───┬──────┬────┐
			│ 1 │ John │ NY │
			└───┴──────┴────┘

`, // No header, falls back to string reps
		},
		{
			name:       "PreExistingHeaders",
			data:       []WithTags{{1, "John", 30, "NY"}},
			preHeaders: []string{"CustomID", "CustomName", "CustomCity"},
			enable:     true,
			expected: `
            ┌───────────┬─────────────┬─────────────┐
            │ CUSTOM ID │ CUSTOM NAME │ CUSTOM CITY │
            ├───────────┼─────────────┼─────────────┤
            │ 1         │ John        │ NY          │
            └───────────┴─────────────┴─────────────┘

`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			table := tablewriter.NewTable(&buf)
			if tt.enable {
				table.Configure(func(cfg *tablewriter.Config) {
					cfg.Behavior.Structs.AutoHeader = tw.On
				})
			}

			if len(tt.preHeaders) > 0 {
				table.Header(tt.preHeaders)
			}
			err := table.Bulk(tt.data)
			if err != nil {
				t.Fatalf("Bulk failed: %v", err)
			}
			table.Render()

			visualCheck(t, tt.name, buf.String(), tt.expected)
		})
	}
}
