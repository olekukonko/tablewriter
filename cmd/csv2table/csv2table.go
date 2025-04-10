package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"os"
	"unicode/utf8"

	"github.com/olekukonko/tablewriter"
	"github.com/olekukonko/tablewriter/renderer"
	"github.com/olekukonko/tablewriter/tw"
)

var (
	fileName  = flag.String("f", "", "Set file (e.g., sample.csv)")
	delimiter = flag.String("d", ",", "Set CSV delimiter (e.g., ,|;|\t)")
	header    = flag.Bool("h", true, "Enable/disable header row")
	align     = flag.String("a", "none", "Set alignment (none|left|right|center)")
	pipe      = flag.Bool("p", false, "Read from STDIN")
	border    = flag.Bool("b", true, "Enable/disable table borders")
)

func main() {
	flag.Parse()
	fmt.Println() // Leading newline for clean output
	if *pipe {
		process(os.Stdin)
	} else {
		if *fileName == "" {
			flag.Usage()
			os.Exit(1)
		}
		processFile()
	}
	fmt.Println() // Trailing newline for clean output
}

func processFile() {
	file, err := os.Open(*fileName)
	if err != nil {
		exit(err)
	}
	defer file.Close()
	process(file)
}

func process(r io.Reader) {
	// Configure CSV reader
	csvReader := csv.NewReader(r)
	if *delimiter != "" {
		rune, size := utf8.DecodeRuneInString(*delimiter)
		if size == 0 {
			rune = ',' // Default to comma if invalid
		}
		csvReader.Comma = rune
	}

	// Define border configurations
	on := renderer.Border{
		Left:   tw.On,
		Right:  tw.On,
		Top:    tw.On,
		Bottom: tw.On,
	}
	off := renderer.Border{
		Left:   tw.Off,
		Right:  tw.Off,
		Top:    tw.Off,
		Bottom: tw.Off,
	}

	// Set border based on flag (corrected logic)
	b := off
	if *border {
		b = on
	}

	// Configure renderer with default symbols and settings
	renderConfig := renderer.DefaultConfig{
		Borders: b,
		Settings: renderer.Settings{
			Separators: renderer.Separators{
				ShowHeader:     tw.On,
				ShowFooter:     tw.On,
				BetweenRows:    tw.Off,
				BetweenColumns: tw.On,
			},
			Lines: renderer.Lines{
				ShowTop:        tw.On,
				ShowBottom:     tw.On,
				ShowHeaderLine: tw.On,
				ShowFooterLine: tw.On,
			},
			TrimWhitespace: tw.On,
			CompactMode:    tw.Off,
		},
		Debug: false,
	}

	// Create table with configuration
	table := tablewriter.NewTable(os.Stdout,
		tablewriter.WithDebug(false),
		tablewriter.WithHeaderConfig(getHeaderConfig()),
		tablewriter.WithRowConfig(getRowConfig()),
		tablewriter.WithFooterConfig(getFooterConfig()),
		tablewriter.WithRenderer(renderer.NewDefault(renderConfig)),
	)

	// Read and process CSV data
	records, err := csvReader.ReadAll()
	if err != nil {
		exit(err)
	}

	if len(records) == 0 {
		fmt.Println("No data to display")
		return
	}

	// Set header if enabled
	if *header && len(records) > 0 {
		table.SetHeader(records[0])
		records = records[1:]
	}

	// Add rows to table
	for _, record := range records {
		if err := table.Append(record); err != nil {
			exit(err)
		}
	}

	// Render the table
	if err := table.Render(); err != nil {
		exit(err)
	}
}

func getHeaderConfig() tablewriter.CellConfig {
	cfg := tablewriter.CellConfig{
		Formatting: tablewriter.CellFormatting{
			Alignment:  tw.AlignCenter,
			AutoFormat: true,
			AutoWrap:   tw.WrapTruncate,
		},
		Padding: tablewriter.CellPadding{
			Global: tw.Padding{Left: " ", Right: " ", Top: "", Bottom: ""},
		},
	}

	switch *align {
	case "left":
		cfg.Formatting.Alignment = tw.AlignLeft
	case "right":
		cfg.Formatting.Alignment = tw.AlignRight
	case "center":
		cfg.Formatting.Alignment = tw.AlignCenter
	}

	return cfg
}

func getRowConfig() tablewriter.CellConfig {
	cfg := tablewriter.CellConfig{
		Formatting: tablewriter.CellFormatting{
			AutoWrap: tw.WrapNormal,
		},
		Padding: tablewriter.CellPadding{
			Global: tw.Padding{Left: " ", Right: " ", Top: "", Bottom: ""},
		},
	}

	switch *align {
	case "left":
		cfg.Formatting.Alignment = tw.AlignLeft
	case "right":
		cfg.Formatting.Alignment = tw.AlignRight
	case "center":
		cfg.Formatting.Alignment = tw.AlignCenter
	default:
		cfg.Formatting.Alignment = tw.AlignLeft
	}
	return cfg
}

func getFooterConfig() tablewriter.CellConfig {
	return tablewriter.CellConfig{
		Formatting: tablewriter.CellFormatting{
			Alignment: tw.AlignRight,
			AutoWrap:  tw.WrapNormal,
		},
	}
}

func exit(err error) {
	fmt.Fprintf(os.Stderr, "Error: %v\n", err)
	os.Exit(1)
}
