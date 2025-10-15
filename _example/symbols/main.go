package main

import (
	"fmt"
	"os"

	"github.com/olekukonko/ll"
	"github.com/olekukonko/tablewriter"
	"github.com/olekukonko/tablewriter/renderer"
	"github.com/olekukonko/tablewriter/tw"
)

func main() {
	data := [][]string{
		{"Engineering", "Backend", "API Team", "Alice"},
		{"Engineering", "Backend", "Database Team", "Bob"},
		{"Engineering", "Frontend", "UI Team", "Charlie"},
		{"Marketing", "Digital", "SEO Team", "Dave"},
		{"Marketing", "Digital", "Content Team", "Eve"},
	}

	cnf := tablewriter.Config{
		Header: tw.CellConfig{
			Alignment: tw.CellAlignment{Global: tw.AlignCenter},
		},
		Row: tw.CellConfig{
			Merging:   tw.CellMerging{Mode: tw.MergeHierarchical},
			Alignment: tw.CellAlignment{Global: tw.AlignLeft},
		},
		Debug: false,
	}

	// Create a custom border style
	DottedStyle := []tw.Symbols{
		tw.NewSymbolCustom("Dotted").
			WithRow("·").
			WithColumn(":").
			WithTopLeft(".").
			WithTopMid("·").
			WithTopRight(".").
			WithMidLeft(":").
			WithCenter("+").
			WithMidRight(":").
			WithBottomLeft("'").
			WithBottomMid("·").
			WithBottomRight("'"),

		// arrow style
		tw.NewSymbolCustom("Arrow").
			WithRow("→").
			WithColumn("↓").
			WithTopLeft("↗").
			WithTopMid("↑").
			WithTopRight("↖").
			WithMidLeft("→").
			WithCenter("↔").
			WithMidRight("←").
			WithBottomLeft("↘").
			WithBottomMid("↓").
			WithBottomRight("↙"),

		// start style
		tw.NewSymbolCustom("Starry").
			WithRow("★").
			WithColumn("☆").
			WithTopLeft("✧").
			WithTopMid("✯").
			WithTopRight("✧").
			WithMidLeft("✦").
			WithCenter("✶").
			WithMidRight("✦").
			WithBottomLeft("✧").
			WithBottomMid("✯").
			WithBottomRight("✧"),

		tw.NewSymbolCustom("Hearts").
			WithRow("♥").
			WithColumn("❤").
			WithTopLeft("❥").
			WithTopMid("♡").
			WithTopRight("❥").
			WithMidLeft("❣").
			WithCenter("✚").
			WithMidRight("❣").
			WithBottomLeft("❦").
			WithBottomMid("♡").
			WithBottomRight("❦"),

		tw.NewSymbolCustom("Tech").
			WithRow("=").
			WithColumn("||").
			WithTopLeft("/*").
			WithTopMid("##").
			WithTopRight("*/").
			WithMidLeft("//").
			WithCenter("<>").
			WithMidRight("\\").
			WithBottomLeft("\\*").
			WithBottomMid("##").
			WithBottomRight("*/"),

		tw.NewSymbolCustom("Nature").
			WithRow("~").
			WithColumn("|").
			WithTopLeft("🌱").
			WithTopMid("🌿").
			WithTopRight("🌱").
			WithMidLeft("🍃").
			WithCenter("❀").
			WithMidRight("🍃").
			WithBottomLeft("🌻").
			WithBottomMid("🌾").
			WithBottomRight("🌻"),

		tw.NewSymbolCustom("Artistic").
			WithRow("▬").
			WithColumn("▐").
			WithTopLeft("◈").
			WithTopMid("◊").
			WithTopRight("◈").
			WithMidLeft("◀").
			WithCenter("⬔").
			WithMidRight("▶").
			WithBottomLeft("◭").
			WithBottomMid("▣").
			WithBottomRight("◮"),

		tw.NewSymbolCustom("8-Bit").
			WithRow("■").
			WithColumn("█").
			WithTopLeft("╔").
			WithTopMid("▲").
			WithTopRight("╗").
			WithMidLeft("◄").
			WithCenter("♦").
			WithMidRight("►").
			WithBottomLeft("╚").
			WithBottomMid("▼").
			WithBottomRight("╝"),

		tw.NewSymbolCustom("Chaos").
			WithRow("≈").
			WithColumn("§").
			WithTopLeft("⌘").
			WithTopMid("∞").
			WithTopRight("⌥").
			WithMidLeft("⚡").
			WithCenter("☯").
			WithMidRight("♞").
			WithBottomLeft("⌂").
			WithBottomMid("∆").
			WithBottomRight("◊"),

		tw.NewSymbolCustom("Dots").
			WithRow("·").
			WithColumn(" "). // Invisible column lines
			WithTopLeft("·").
			WithTopMid("·").
			WithTopRight("·").
			WithMidLeft(" ").
			WithCenter("·").
			WithMidRight(" ").
			WithBottomLeft("·").
			WithBottomMid("·").
			WithBottomRight("·"),

		tw.NewSymbolCustom("Blocks").
			WithRow("▀").
			WithColumn("█").
			WithTopLeft("▛").
			WithTopMid("▀").
			WithTopRight("▜").
			WithMidLeft("▌").
			WithCenter("█").
			WithMidRight("▐").
			WithBottomLeft("▙").
			WithBottomMid("▄").
			WithBottomRight("▟"),

		tw.NewSymbolCustom("Zen").
			WithRow("~").
			WithColumn(" ").
			WithTopLeft(" ").
			WithTopMid("♨").
			WithTopRight(" ").
			WithMidLeft(" ").
			WithCenter("☯").
			WithMidRight(" ").
			WithBottomLeft(" ").
			WithBottomMid("♨").
			WithBottomRight(" "),
	}

	var table *tablewriter.Table
	for _, style := range DottedStyle {
		ll.Info(style.Name() + " style")
		table = tablewriter.NewTable(os.Stdout,
			tablewriter.WithRenderer(renderer.NewBlueprint(tw.Rendition{Symbols: style})),
			tablewriter.WithConfig(cnf),
		)
		table.Header([]string{"Department", "Division", "Team", "Lead"})
		table.Bulk(data)
		table.Render()

		fmt.Println()
	}
}
