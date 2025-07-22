package tests

import (
	"io"
	"strings"
	"testing"

	"github.com/olekukonko/tablewriter"
	"github.com/olekukonko/tablewriter/renderer"
	"github.com/olekukonko/tablewriter/tw"
)

type Country string

func (c Country) String() string { return strings.ToUpper(string(c)) }

func BenchmarkBlueprint(b *testing.B) {
	table := tablewriter.NewTable(io.Discard, tablewriter.WithRenderer(renderer.NewBlueprint()))
	table.Header([]string{"Name", "Age", "City"})
	for i := 0; i < b.N; i++ {
		table.Append([]any{"Alice", Age(25), Country("New York")})
		table.Append([]string{"Bob", "30", "Boston"})
		table.Render()
	}
}

func BenchmarkOcean(b *testing.B) {
	table := tablewriter.NewTable(io.Discard, tablewriter.WithRenderer(renderer.NewOcean()))
	table.Header([]string{"Name", "Age", "City"})
	for i := 0; i < b.N; i++ {
		table.Append([]any{"Alice", Age(25), Country("New York")})
		table.Append([]string{"Bob", "30", "Boston"})
		table.Render()
	}
}

func BenchmarkMarkdown(b *testing.B) {
	table := tablewriter.NewTable(io.Discard, tablewriter.WithRenderer(renderer.NewMarkdown()))
	table.Header([]string{"Name", "Age", "City"})
	for i := 0; i < b.N; i++ {
		table.Append([]any{"Alice", Age(25), Country("New York")})
		table.Append([]string{"Bob", "30", "Boston"})
		table.Render()
	}
}

func BenchmarkColorized(b *testing.B) {
	table := tablewriter.NewTable(io.Discard, tablewriter.WithRenderer(renderer.NewColorized()))
	table.Header([]string{"Name", "Age", "City"})
	for i := 0; i < b.N; i++ {
		table.Append([]any{"Alice", Age(25), Country("New York")})
		table.Append([]string{"Bob", "30", "Boston"})
		table.Render()
	}
}

func BenchmarkStreamBlueprint(b *testing.B) {
	table := tablewriter.NewTable(io.Discard,
		tablewriter.WithRenderer(renderer.NewBlueprint()),
		tablewriter.WithStreaming(tw.StreamConfig{Enable: true}))

	err := table.Start()
	if err != nil {
		b.Fatal(err)
	}
	table.Header([]string{"Name", "Age", "City"})
	for i := 0; i < b.N; i++ {
		table.Append([]any{"Alice", Age(25), Country("New York")})
		table.Append([]string{"Bob", "30", "Boston"})

	}

	err = table.Close()
	if err != nil {
		b.Fatal(err)
	}
}

func BenchmarkStreamOcean(b *testing.B) {
	table := tablewriter.NewTable(io.Discard,
		tablewriter.WithRenderer(renderer.NewOcean()),
		tablewriter.WithStreaming(tw.StreamConfig{Enable: true}))

	err := table.Start()
	if err != nil {
		b.Fatal(err)
	}
	table.Header([]string{"Name", "Age", "City"})
	for i := 0; i < b.N; i++ {
		table.Append([]any{"Alice", Age(25), Country("New York")})
		table.Append([]string{"Bob", "30", "Boston"})

	}

	err = table.Close()
	if err != nil {
		b.Fatal(err)
	}
}
