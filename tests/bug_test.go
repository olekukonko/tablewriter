package tests

import (
	"bytes"
	"fmt"
	"github.com/olekukonko/tablewriter"
	"github.com/olekukonko/tablewriter/tw"
	"testing"
)

type Cleaner string

// Note: Format() overrides String() if both exist.
func (c Cleaner) Format() string {
	return clean(string(c))
}

type Age int

// Age int will be ignore and string will be used
func (a Age) String() string {
	return fmt.Sprintf("%dyrs", a)
}

type Person struct {
	Name string
	Age  int
	City string
}

type Profile struct {
	Name Cleaner
	Age  Age
	City string
}

func TestBug252(t *testing.T) {
	var buf bytes.Buffer
	type Person struct {
		Name string
		Age  int
		City string
	}

	header := []string{"Name", "Age", "City"}
	alice := Person{Name: "Alice", Age: 25, City: "New York"}
	bob := Profile{Name: Cleaner("Bo   b"), Age: Age(30), City: "Boston"}

	t.Run("Normal", func(t *testing.T) {
		buf.Reset()
		table := tablewriter.NewTable(&buf, tablewriter.WithDebug(true))
		table.Header(header)
		table.Append("Alice", "25", "New York")
		table.Append("Bob", "30", "Boston")
		table.Render()

		expected := `
	┌───────┬─────┬──────────┐
	│ NAME  │ AGE │   CITY   │
	├───────┼─────┼──────────┤
	│ Alice │ 25  │ New York │
	│ Bob   │ 30  │ Boston   │
	└───────┴─────┴──────────┘
`
		debug := visualCheck(t, "TestBug252-Normal", buf.String(), expected)
		if !debug {
			t.Error(table.Debug())
		}

	})

	t.Run("Mixed", func(t *testing.T) {
		buf.Reset()
		table := tablewriter.NewTable(&buf, tablewriter.WithDebug(true))
		table.Header(header)
		table.Append(alice)
		table.Append("Bob", "30", "Boston")
		table.Render()

		expected := `
	┌───────┬─────┬──────────┐
	│ NAME  │ AGE │   CITY   │
	├───────┼─────┼──────────┤
	│ Alice │ 25  │ New York │
	│ Bob   │ 30  │ Boston   │
	└───────┴─────┴──────────┘
`
		debug := visualCheck(t, "TestBug252-Mixed", buf.String(), expected)
		if !debug {
			// t.Error(table.Debug())
		}

	})

	t.Run("Profile", func(t *testing.T) {
		buf.Reset()
		table := tablewriter.NewTable(&buf, tablewriter.WithDebug(true))
		table.Header(header)
		table.Append(Cleaner("A   lice"), Cleaner("2   5 yrs"), "New York")
		table.Append(bob)
		table.Render()

		expected := `
		┌───────┬───────┬──────────┐
		│ NAME  │  AGE  │   CITY   │
		├───────┼───────┼──────────┤
		│ Alice │ 25yrs │ New York │
		│ Bob   │ 30yrs │ Boston   │
		└───────┴───────┴──────────┘

`
		debug := visualCheck(t, "TestBasicTableDefault", buf.String(), expected)
		if !debug {
			// t.Error(table.Debug())
		}
	})

	type Override struct {
		Fish string `json:"name"`
		Name string `json:"-"`
		Age  Age
		City string
	}
	t.Run("Override", func(t *testing.T) {
		buf.Reset()
		table := tablewriter.NewTable(&buf, tablewriter.WithDebug(true))
		table.Header(header)
		table.Append(Cleaner("A   lice"), Cleaner("2   5 yrs"), "New York")
		table.Append(Override{
			Fish: "Bob",
			Name: "Skip",
			Age:  Age(25),
			City: "Boston",
		})
		table.Render()

		expected := `
        ┌───────┬───────┬──────────┐
        │ NAME  │  AGE  │   CITY   │
        ├───────┼───────┼──────────┤
        │ Alice │ 25yrs │ New York │
        │ Bob   │ 25yrs │ Boston   │
        └───────┴───────┴──────────┘

`
		debug := visualCheck(t, "TestBug252-Override", buf.String(), expected)
		if !debug {
			// t.Error(table.Debug())
		}
	})

	t.Run("Override-Streaming", func(t *testing.T) {
		buf.Reset()
		table := tablewriter.NewTable(&buf, tablewriter.WithDebug(true), tablewriter.WithStreaming(tw.StreamConfig{Enable: true}))

		err := table.Start()
		if err != nil {
			t.Error(err)
		}

		table.Header(header)
		table.Append(Cleaner("A   lice"), Cleaner("2   5 yrs"), "New York")
		table.Append(Override{
			Fish: "Bob",
			Name: "Skip",
			Age:  Age(25),
			City: "Boston",
		})
		expected := `
        ┌────────┬────────┬────────┐
        │  NAME  │  AGE   │  CITY  │
        ├────────┼────────┼────────┤
        │ Alice  │ 25yrs  │ New    │
        │        │        │ York   │
        │ Bob    │ 25yrs  │ Boston │
        └────────┴────────┴────────┘


`
		err = table.Close()
		if err != nil {
			t.Error(err)
		}
		debug := visualCheck(t, "TestBug252-Override-Streaming", buf.String(), expected)
		if !debug {
			// t.Error(table.Debug())
		}
	})

}

func TestBug254(t *testing.T) {
	var buf bytes.Buffer
	data := [][]string{
		{"  LEFT", "RIGHT  ", "  BOTH  "},
	}
	t.Run("Normal", func(t *testing.T) {
		buf.Reset()
		table := tablewriter.NewTable(&buf,
			tablewriter.WithRowMaxWidth(20),
			tablewriter.WithTrimSpace(tw.On),
			tablewriter.WithAlignment(tw.Alignment{tw.AlignCenter, tw.AlignCenter, tw.AlignCenter}),
		)
		table.Bulk(data)
		table.Render()

		expected := `
		┌──────┬───────┬──────┐
		│ LEFT │ RIGHT │ BOTH │
		└──────┴───────┴──────┘

`
		debug := visualCheck(t, "TestBug252-Normal", buf.String(), expected)
		if !debug {
			t.Error(table.Debug())
		}

	})

	t.Run("Mixed", func(t *testing.T) {
		buf.Reset()
		table := tablewriter.NewTable(&buf,
			tablewriter.WithRowMaxWidth(20),
			tablewriter.WithTrimSpace(tw.Off),
			tablewriter.WithAlignment(tw.Alignment{tw.AlignCenter, tw.AlignCenter, tw.AlignCenter}),
		)
		table.Bulk(data)
		table.Render()

		expected := `
	┌────────┬─────────┬──────────┐
	│   LEFT │ RIGHT   │   BOTH   │
	└────────┴─────────┴──────────┘
`
		debug := visualCheck(t, "TestBug252-Mixed", buf.String(), expected)
		if !debug {
			t.Error(table.Debug())
		}

	})

}

func TestBug267(t *testing.T) {
	var buf bytes.Buffer
	data := [][]string{
		{"a", "b", "c"},
		{"aa", "bb", "cc"},
	}
	t.Run("WithoutMaxWith", func(t *testing.T) {
		buf.Reset()
		table := tablewriter.NewTable(&buf,
			tablewriter.WithTrimSpace(tw.On),
			tablewriter.WithConfig(tablewriter.Config{Row: tw.CellConfig{Padding: tw.CellPadding{
				Global: tw.PaddingNone,
			}}}),
		)
		table.Bulk(data)
		table.Render()

		expected := `
		┌──┬──┬──┐
		│a │b │c │
		│aa│bb│cc│
		└──┴──┴──┘

`
		debug := visualCheck(t, "TestBug252-Normal", buf.String(), expected)
		if !debug {
			t.Error(table.Debug())
		}

	})

	t.Run("WithMaxWidth", func(t *testing.T) {
		buf.Reset()
		table := tablewriter.NewTable(&buf,
			tablewriter.WithRowMaxWidth(20),
			tablewriter.WithTrimSpace(tw.Off),
			tablewriter.WithAlignment(tw.Alignment{tw.AlignCenter, tw.AlignCenter, tw.AlignCenter}),
			tablewriter.WithDebug(false),
			tablewriter.WithConfig(tablewriter.Config{Row: tw.CellConfig{Padding: tw.CellPadding{
				Global: tw.PaddingNone,
			}}}),
		)
		table.Bulk(data)
		table.Render()

		expected := `
            ┌──┬──┬──┐
            │a │b │c │
            │aa│bb│cc│
            └──┴──┴──┘
`
		debug := visualCheck(t, "TestBug252-Mixed", buf.String(), expected)
		if !debug {
			t.Error(table.Debug())
		}

	})

}
