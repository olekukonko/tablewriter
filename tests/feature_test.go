package tests

import (
	"bytes"
	"github.com/olekukonko/tablewriter"
	"github.com/olekukonko/tablewriter/renderer"
	"github.com/olekukonko/tablewriter/tw"
	"testing"
)

func TestBug260(t *testing.T) {
	var buf bytes.Buffer

	tableRendition := tw.Rendition{
		Borders: tw.BorderNone,
		Settings: tw.Settings{
			Separators: tw.Separators{
				ShowHeader:     tw.Off,
				ShowFooter:     tw.Off,
				BetweenRows:    tw.Off,
				BetweenColumns: tw.Off,
			},
		},
		Symbols: tw.NewSymbols(tw.StyleNone),
	}

	t.Run("Normal", func(t *testing.T) {
		buf.Reset()
		tableRenderer := renderer.NewBlueprint(tableRendition)
		table := tablewriter.NewTable(
			&buf,
			tablewriter.WithRenderer(tableRenderer),
			tablewriter.WithTableMax(120),
			tablewriter.WithTrimSpace(tw.Off),
			tablewriter.WithDebug(true),
			tablewriter.WithPadding(tw.PaddingNone),
		)

		table.Append([]string{"INFO:",
			"The original machine had a base-plate of prefabulated aluminite, surmounted by a malleable logarithmic casing in such a way that the two main spurving bearings were in a direct line with the pentametric fan.",
		})

		table.Append("INFO:",
			"The original machine had a base-plate of prefabulated aluminite, surmounted by a malleable logarithmic casing in such a way that the two main spurving bearings were in a direct line with the pentametric fan.",
		)

		table.Render()

		expected := `
		INFO:The original machine had a base-plate of prefabulated     
			 aluminite, surmounted by a malleable logarithmic casing in
			 such a way that the two main spurving bearings were in a  
			 direct line with the pentametric fan.                     
		INFO:The original machine had a base-plate of prefabulated     
			 aluminite, surmounted by a malleable logarithmic casing in
			 such a way that the two main spurving bearings were in a  
			 direct line with the pentametric fan.    
`
		debug := visualCheck(t, "TestBug252-Mixed", buf.String(), expected)
		if !debug {
			t.Error(table.Debug())
		}

	})

	t.Run("Mixed", func(t *testing.T) {
		buf.Reset()
		tableRenderer := renderer.NewBlueprint(tableRendition)
		table := tablewriter.NewTable(
			&buf,
			tablewriter.WithRenderer(tableRenderer),
			tablewriter.WithTableMax(120),
			tablewriter.WithTrimSpace(tw.Off),
			tablewriter.WithDebug(true),
			tablewriter.WithPadding(tw.PaddingNone),
		)

		table.Append([]string{"INFO:",
			"The original machine had a base-plate of prefabulated aluminite, surmounted by a malleable logarithmic casing in such a way that the two main spurving bearings were in a direct line with the pentametric fan.",
		})

		table.Append("INFO: ",
			"The original machine had a base-plate of prefabulated aluminite, surmounted by a malleable logarithmic casing in such a way that the two main spurving bearings were in a direct line with the pentametric fan.",
		)

		table.Render()

		expected := `
		INFO: The original machine had a base-plate of prefabulated     
			  aluminite, surmounted by a malleable logarithmic casing in
			  such a way that the two main spurving bearings were in a  
			  direct line with the pentametric fan.                     
		INFO: The original machine had a base-plate of prefabulated     
			  aluminite, surmounted by a malleable logarithmic casing in
			  such a way that the two main spurving bearings were in a  
			  direct line with the pentametric fan.  
`
		debug := visualCheck(t, "TestBug252-Mixed", buf.String(), expected)
		if !debug {
			t.Error(table.Debug())
		}

	})

}
