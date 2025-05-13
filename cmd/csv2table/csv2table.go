package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"github.com/olekukonko/ll"
	"github.com/olekukonko/ll/lh"
	"github.com/olekukonko/ts" // For terminal size
	"io"
	"math"
	"os"
	"strings"
	"unicode/utf8"

	"github.com/olekukonko/tablewriter"
	"github.com/olekukonko/tablewriter/renderer"
	"github.com/olekukonko/tablewriter/tw"
)

var (
	fileName      = flag.String("f", "", "Set CSV file path (e.g., sample.csv). If empty and -p is not set, STDIN is used.")
	delimiter     = flag.String("d", ",", "Set CSV delimiter (e.g., \",\" \"|\" \"\\t\"). For tab, use actual tab or '\\t'.")
	header        = flag.Bool("h", true, "Treat the first row as a header.")
	align         = flag.String("a", "none", "Set global cell alignment (none|left|right|center). 'none' uses renderer defaults.")
	pipe          = flag.Bool("p", false, "Read CSV data from STDIN (standard input). Overrides -f if both are set.")
	border        = flag.Bool("b", true, "Enable/disable table borders and lines (overall structure).")
	streaming     = flag.Bool("s", false, "Enable streaming mode (processes row-by-row). Might not support all features like AutoHide.")
	rendererType  = flag.String("renderer", "blueprint", "Set table renderer (blueprint|colorized|markdown|html|svg|ocean).")
	symbolStyle   = flag.String("symbols", "light", "Set border symbol style (light|ascii|heavy|double|rounded|markdown|graphical|dotted|arrow|starry|hearts|tech|nature|artistic|8-bit|chaos|dots|blocks|zen|none).")
	rowAutoWrap   = flag.String("wrap", "normal", "Set row auto-wrap mode (normal|truncate|break|none).")
	inferColumns  = flag.Bool("infer", true, "Attempt to infer and normalize column counts if CSV rows are ragged. If false, CSV parsing errors on mismatched columns will halt.")
	tableMaxWidth = flag.Int("maxwidth", 0, "Set maximum table width in characters (0 for auto based on 90% terminal width or content).")
	debug         = flag.Bool("debug", false, "Enable debug logging for tablewriter operations.")

	// add namespace
	logger = ll.Namespace("csv2table").Handler(lh.NewColorizedHandler(os.Stdout))
)

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options] [file]\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Reads CSV data from a file or STDIN and renders it as a formatted table.\n\n")
		fmt.Fprintf(os.Stderr, "If [file] is provided, it overrides the -f flag.\n")
		fmt.Fprintf(os.Stderr, "If no [file] and no -f is provided, and -p is not set, STDIN is used.\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
	}
	flag.Parse()

	// Handle non-flag filename argument
	if flag.NArg() > 0 {
		*fileName = flag.Arg(0)
		logger.Info("Using filename from argument: %s", *fileName)
	}

	// Determine input source
	var inputReader io.Reader
	var err error

	if *pipe {
		logger.Info("Reading CSV from STDIN (pipe mode).")
		inputReader = os.Stdin
	} else if *fileName != "" {
		logger.Info("Reading CSV from file: %s", *fileName)
		file, errFile := os.Open(*fileName)
		if errFile != nil {
			logger.Fatal("failed to open file '%s': %w", *fileName, errFile)
		}
		defer file.Close()
		inputReader = file
	} else {
		logger.Info("No file specified and pipe mode not active. Reading CSV from STDIN.")
		inputReader = os.Stdin
	}

	// Leading newline for cleaner output, unless it's HTML/SVG etc.
	if !isGraphicalRenderer(*rendererType) {
		fmt.Println()
	}

	err = process(inputReader)
	if err != nil {
		logger.Fatal(err) // process will return specific errors
	}

	if !isGraphicalRenderer(*rendererType) {
		fmt.Println() // Trailing newline
	}
}

func process(r io.Reader) error {
	// --- CSV Reader Configuration ---
	csvInputReader := csv.NewReader(r)
	if *delimiter != "" {
		// Handle literal \t for tab delimiter
		d := *delimiter
		if d == "\\t" {
			d = "\t"
		}
		runeVal, size := utf8.DecodeRuneInString(d)
		if size == 0 {
			logger.Warn("Invalid or empty delimiter specified, using default comma ','.")
			runeVal = ','
		}
		csvInputReader.Comma = runeVal
	}
	// If inferring columns, we need to allow variable fields in the first pass.
	// If not inferring, `FieldsPerRecord = 0` will cause csv.Reader to error on inconsistent rows.
	if !*inferColumns {
		csvInputReader.FieldsPerRecord = 0 // Standard Go CSV behavior: error on inconsistent fields after first.
	} else {
		csvInputReader.FieldsPerRecord = -1 // Allow variable fields for the first pass if inferring.
	}

	// --- Symbol Selection ---
	var selectedSymbols tw.Symbols
	// (Full switch statement for symbolStyle as provided previously)
	switch strings.ToLower(*symbolStyle) {
	case "ascii":
		selectedSymbols = tw.NewSymbols(tw.StyleASCII)
	case "light", "default":
		selectedSymbols = tw.NewSymbols(tw.StyleLight)
	case "heavy":
		selectedSymbols = tw.NewSymbols(tw.StyleHeavy)
	case "double":
		selectedSymbols = tw.NewSymbols(tw.StyleDouble)
	case "rounded":
		selectedSymbols = tw.NewSymbols(tw.StyleRounded)
	case "markdown":
		selectedSymbols = tw.NewSymbols(tw.StyleMarkdown)
	case "graphical":
		selectedSymbols = tw.NewSymbols(tw.StyleGraphical)
	case "dotted":
		selectedSymbols = tw.NewSymbolCustom("Dotted").WithRow("·").WithColumn(":").WithTopLeft(".").WithTopMid("·").WithTopRight(".").WithMidLeft(":").WithCenter("+").WithMidRight(":").WithBottomLeft("'").WithBottomMid("·").WithBottomRight("'")
	case "arrow":
		selectedSymbols = tw.NewSymbolCustom("Arrow").WithRow("→").WithColumn("↓").WithTopLeft("↗").WithTopMid("↑").WithTopRight("↖").WithMidLeft("→").WithCenter("↔").WithMidRight("←").WithBottomLeft("↘").WithBottomMid("↓").WithBottomRight("↙")
	case "starry":
		selectedSymbols = tw.NewSymbolCustom("Starry").WithRow("★").WithColumn("☆").WithTopLeft("✧").WithTopMid("✯").WithTopRight("✧").WithMidLeft("✦").WithCenter("✶").WithMidRight("✦").WithBottomLeft("✧").WithBottomMid("✯").WithBottomRight("✧")
	case "hearts":
		selectedSymbols = tw.NewSymbolCustom("Hearts").WithRow("♥").WithColumn("❤").WithTopLeft("❥").WithTopMid("♡").WithTopRight("❥").WithMidLeft("❣").WithCenter("✚").WithMidRight("❣").WithBottomLeft("❦").WithBottomMid("♡").WithBottomRight("❦")
	case "tech":
		selectedSymbols = tw.NewSymbolCustom("Tech").WithRow("=").WithColumn("||").WithTopLeft("/*").WithTopMid("##").WithTopRight("*/").WithMidLeft("//").WithCenter("<>").WithMidRight("\\").WithBottomLeft("\\*").WithBottomMid("##").WithBottomRight("*/")
	case "nature":
		selectedSymbols = tw.NewSymbolCustom("Nature").WithRow("~").WithColumn("|").WithTopLeft("🌱").WithTopMid("🌿").WithTopRight("🌱").WithMidLeft("🍃").WithCenter("❀").WithMidRight("🍃").WithBottomLeft("🌻").WithBottomMid("🌾").WithBottomRight("🌻")
	case "artistic":
		selectedSymbols = tw.NewSymbolCustom("Artistic").WithRow("▬").WithColumn("▐").WithTopLeft("◈").WithTopMid("◊").WithTopRight("◈").WithMidLeft("◀").WithCenter("⬔").WithMidRight("▶").WithBottomLeft("◭").WithBottomMid("▣").WithBottomRight("◮")
	case "8-bit":
		selectedSymbols = tw.NewSymbolCustom("8-Bit").WithRow("■").WithColumn("█").WithTopLeft("╔").WithTopMid("▲").WithTopRight("╗").WithMidLeft("◄").WithCenter("♦").WithMidRight("►").WithBottomLeft("╚").WithBottomMid("▼").WithBottomRight("╝")
	case "chaos":
		selectedSymbols = tw.NewSymbolCustom("Chaos").WithRow("≈").WithColumn("§").WithTopLeft("⌘").WithTopMid("∞").WithTopRight("⌥").WithMidLeft("⚡").WithCenter("☯").WithMidRight("♞").WithBottomLeft("⌂").WithBottomMid("∆").WithBottomRight("◊")
	case "dots":
		selectedSymbols = tw.NewSymbolCustom("Dots").WithRow("·").WithColumn(" ").WithTopLeft("·").WithTopMid("·").WithTopRight("·").WithMidLeft(" ").WithCenter("·").WithMidRight(" ").WithBottomLeft("·").WithBottomMid("·").WithBottomRight("·")
	case "blocks":
		selectedSymbols = tw.NewSymbolCustom("Blocks").WithRow("▀").WithColumn("█").WithTopLeft("▛").WithTopMid("▀").WithTopRight("▜").WithMidLeft("▌").WithCenter("█").WithMidRight("▐").WithBottomLeft("▙").WithBottomMid("▄").WithBottomRight("▟")
	case "zen":
		selectedSymbols = tw.NewSymbolCustom("Zen").WithRow("~").WithColumn(" ").WithTopLeft(" ").WithTopMid("♨").WithTopRight(" ").WithMidLeft(" ").WithCenter("☯").WithMidRight(" ").WithBottomLeft(" ").WithBottomMid("♨").WithBottomRight(" ")
	case "none":
		selectedSymbols = tw.NewSymbols(tw.StyleNone)
	default:
		logger.Warn(fmt.Sprintf("Unknown symbol style '%s', using default (Light).", *symbolStyle))
		selectedSymbols = tw.NewSymbols(tw.StyleLight)
	}

	// --- Base Rendition Configuration ---
	borderCfg := tw.Border{Left: tw.Off, Right: tw.Off, Top: tw.Off, Bottom: tw.Off}
	linesCfg := tw.Lines{ShowTop: tw.Off, ShowBottom: tw.Off, ShowHeaderLine: tw.Off, ShowFooterLine: tw.Off}
	separatorsCfg := tw.Separators{BetweenColumns: tw.Off, ShowHeader: tw.Off, ShowFooter: tw.Off, BetweenRows: tw.Off}

	if *border {
		borderCfg = tw.Border{Left: tw.On, Right: tw.On, Top: tw.On, Bottom: tw.On}
		linesCfg = tw.Lines{ShowTop: tw.On, ShowBottom: tw.On, ShowHeaderLine: tw.On, ShowFooterLine: tw.On}
		separatorsCfg = tw.Separators{ShowHeader: tw.On, ShowFooter: tw.On, BetweenRows: tw.Off, BetweenColumns: tw.On}
	}

	rendererConfiguredSpecifically := false
	switch strings.ToLower(*rendererType) {
	case "markdown":
		selectedSymbols = tw.NewSymbols(tw.StyleMarkdown)
		borderCfg = tw.Border{Left: tw.On, Right: tw.On, Top: tw.Off, Bottom: tw.Off}
		linesCfg = tw.Lines{ShowTop: tw.Off, ShowBottom: tw.Off, ShowHeaderLine: tw.On, ShowFooterLine: tw.Off}
		separatorsCfg = tw.Separators{BetweenColumns: tw.On, ShowHeader: tw.On, BetweenRows: tw.Off, ShowFooter: tw.Off}
		if !*border {
			borderCfg = tw.Border{Left: tw.Off, Right: tw.Off, Top: tw.Off, Bottom: tw.Off}
			linesCfg.ShowHeaderLine = tw.Off
			separatorsCfg.BetweenColumns = tw.Off
		}
		rendererConfiguredSpecifically = true
	case "html", "svg": // These renderers manage their own structure
		borderCfg = tw.Border{}
		linesCfg = tw.Lines{}
		separatorsCfg = tw.Separators{}
		selectedSymbols = tw.NewSymbols(tw.StyleNone)
		rendererConfiguredSpecifically = true
	}

	baseRendition := tw.Rendition{
		Borders:  borderCfg,
		Settings: tw.Settings{Separators: separatorsCfg, Lines: linesCfg, CompactMode: tw.Off},
		Symbols:  selectedSymbols,
	}

	// --- Renderer Instantiation ---
	var selectedRenderer tw.Renderer
	// For CLI, os.Stdout is the writer. For HTML/SVG, their renderers handle this.
	outputTarget := io.Writer(os.Stdout) // Default to os.Stdout

	switch strings.ToLower(*rendererType) {
	case "markdown":
		selectedRenderer = renderer.NewMarkdown(baseRendition)
	case "html":
		selectedRenderer = renderer.NewHTML(outputTarget, *debug, renderer.HTMLConfig{EscapeContent: true})
	case "svg":
		selectedRenderer = renderer.NewSVG(renderer.SVGConfig{FontSize: 12, Padding: 5, Debug: *debug})
	case "colorized":
		selectedRenderer = renderer.NewColorized()
		if r, ok := selectedRenderer.(tw.Renditioning); ok && !rendererConfiguredSpecifically {
			r.Rendition(baseRendition)
		}
	case "ocean":
		selectedRenderer = renderer.NewOcean()
		if r, ok := selectedRenderer.(tw.Renditioning); ok && !rendererConfiguredSpecifically {
			r.Rendition(baseRendition)
		}
	case "blueprint":
		fallthrough
	default:
		if *rendererType != "" && strings.ToLower(*rendererType) != "blueprint" {
			logger.Warn(fmt.Sprintf("Unknown renderer type '%s', using Blueprint.", *rendererType))
		}
		selectedRenderer = renderer.NewBlueprint(baseRendition)
	}

	// --- Table Options & Creation ---
	calculatedMaxWidth := 0
	if *tableMaxWidth > 0 {
		calculatedMaxWidth = *tableMaxWidth
	} else {
		termSize, err := ts.GetSize()
		if err == nil && termSize.Col() > 0 {
			calculatedMaxWidth = int(math.Floor(float64(termSize.Col()) * 0.90))
		}
		// If termSize fails or is 0, calculatedMaxWidth remains 0 (content-based width)
	}
	if calculatedMaxWidth > 0 {
		logger.Info(fmt.Sprintf("Calculated table max width: %d", calculatedMaxWidth))
	}

	tableOpts := []tablewriter.Option{
		tablewriter.WithDebug(*debug),
		tablewriter.WithHeaderConfig(getHeaderConfig(*align, *rowAutoWrap)), // Can use same wrap for header or add specific
		tablewriter.WithRowConfig(getRowConfig(*align, *rowAutoWrap)),
		tablewriter.WithRenderer(selectedRenderer),
		tablewriter.WithTableMax(calculatedMaxWidth), // Apply max width
	}

	if *streaming {
		tableOpts = append(tableOpts, tablewriter.WithStreaming(tw.StreamConfig{Enable: true}))
		logger.Info("Streaming mode ENABLED.")
	} else {
		logger.Info("Streaming mode DISABLED (batch mode).")
	}

	table := tablewriter.NewTable(outputTarget, tableOpts...)

	// --- Data Ingestion and Normalization (Two-Pass if inferring) ---
	var headerData []string
	var dataRecords [][]string

	if *inferColumns {
		logger.Info("Inferring columns (two-pass CSV read).")
		// Pass 1: Get all records and find max columns
		firstPassReader := csv.NewReader(r) // r is the original io.Reader
		// Re-apply delimiter for firstPassReader
		if *delimiter != "" {
			d := *delimiter
			if d == "\\t" {
				d = "\t"
			}
			runeVal, size := utf8.DecodeRuneInString(d)
			if size == 0 {
				runeVal = ','
			}
			firstPassReader.Comma = runeVal
		}
		firstPassReader.FieldsPerRecord = -1 // Allow variable fields

		allRawRecords, errRead := firstPassReader.ReadAll()
		if errRead != nil {
			return fmt.Errorf("error reading CSV during inference pass: %w", errRead)
		}
		if len(allRawRecords) == 0 {
			fmt.Println("No data to display (CSV empty or unreadable in inference pass).")
			return nil
		}

		maxCols := 0
		if *header && len(allRawRecords) > 0 {
			headerData = allRawRecords[0]
			if len(headerData) > maxCols {
				maxCols = len(headerData)
			}
			if len(allRawRecords) > 1 {
				dataRecords = allRawRecords[1:]
			}
		} else {
			dataRecords = allRawRecords
		}

		for _, rec := range dataRecords {
			if len(rec) > maxCols {
				maxCols = len(rec)
			}
		}
		if maxCols == 0 && len(headerData) > 0 { // Only header was present
			maxCols = len(headerData)
		}
		logger.Info(fmt.Sprintf("Inferred max columns: %d", maxCols))

		// Normalize header
		if *header && len(headerData) > 0 {
			normHeader := make([]string, maxCols)
			copy(normHeader, headerData)
			// Padding with empty strings is implicit if normHeader is shorter
			table.Header(normHeader)
		}
		// Normalize data records
		for i := range dataRecords {
			normRecord := make([]string, maxCols)
			copy(normRecord, dataRecords[i])
			// Padding with empty strings is implicit if normRecord is shorter
			dataRecords[i] = normRecord
		}
	} else {
		logger.Info("Not inferring columns (standard CSV parsing). Errors on inconsistent rows are fatal.")
		// If not inferring, use the original csvInputReader (already configured)
		// For batch mode, ReadAll is natural. For streaming, we'll read line by line.
		if !*streaming {
			allRawRecords, errRead := csvInputReader.ReadAll()
			if errRead != nil {
				return fmt.Errorf("error reading CSV (ReadAll, no inference): %w.\nIf CSV has ragged rows, try enabling -infer flag.", errRead)
			}
			if len(allRawRecords) == 0 {
				fmt.Println("No data to display.")
				return nil
			}
			if *header && len(allRawRecords) > 0 {
				headerData = allRawRecords[0]
				table.Header(headerData)
				if len(allRawRecords) > 1 {
					dataRecords = allRawRecords[1:]
				}
			} else {
				dataRecords = allRawRecords
			}
		}
		// Streaming mode without inference will handle records directly from csvInputReader later.
	}

	// --- Table Population and Rendering ---
	if table.Config().Stream.Enable {
		logger.Info("Populating table in STREAMING mode.")
		if err := table.Start(); err != nil {
			return fmt.Errorf("error starting streaming table: %w", err)
		}
		defer table.Close() // Ensure close is called

		if *inferColumns { // We already have normalized headerData and dataRecords
			// Header already set if *header was true
			for i, record := range dataRecords {
				if err := table.Append(record); err != nil {
					return fmt.Errorf("error appending stream record %d (inferred): %w", i, err)
				}
			}
		} else { // Not inferring, read directly
			lineNum := 1
			if *header { // Header would be the first record read by csvInputReader
				headerRow, errH := csvInputReader.Read()
				if errH != nil && errH != io.EOF {
					return fmt.Errorf("error reading header for streaming (no inference): %w", errH)
				}
				if errH != io.EOF && len(headerRow) > 0 {
					table.Header(headerRow)
				}
				lineNum = 2
			}
			for {
				record, errL := csvInputReader.Read()
				if errL == io.EOF {
					break
				}
				if errL != nil {
					return fmt.Errorf("error reading CSV record for streaming on data line approx %d (no inference): %w", lineNum, errL)
				}
				if errA := table.Append(record); errA != nil {
					return fmt.Errorf("error appending stream record on data line approx %d (no inference): %w", lineNum, errA)
				}
				lineNum++
			}
		}
	} else { // Batch mode
		logger.Info("Populating table in BATCH mode.")
		// If inferring, headerData and dataRecords are already prepared and normalized.
		// If not inferring, they were populated from ReadAll().
		// Header was already set if *header was true.
		if err := table.Bulk(dataRecords); err != nil {
			return fmt.Errorf("error appending batch records: %w", err)
		}
		if err := table.Render(); err != nil {
			return fmt.Errorf("error rendering batch table: %w", err)
		}
	}
	return nil
}

func getHeaderConfig(alignFlag string, wrapFlag string) tw.CellConfig {
	cfgFmt := tw.CellFormatting{Alignment: tw.AlignCenter, AutoFormat: true}
	switch strings.ToLower(alignFlag) {
	case "left":
		cfgFmt.Alignment = tw.AlignLeft
	case "right":
		cfgFmt.Alignment = tw.AlignRight
	case "center":
		cfgFmt.Alignment = tw.AlignCenter
	}
	switch strings.ToLower(wrapFlag) {
	case "truncate":
		cfgFmt.AutoWrap = tw.WrapTruncate
	case "break":
		cfgFmt.AutoWrap = tw.WrapBreak
	case "none":
		cfgFmt.AutoWrap = tw.WrapNone
	case "normal":
		cfgFmt.AutoWrap = tw.WrapNormal
	default:
		cfgFmt.AutoWrap = tw.WrapTruncate // Default for headers
	}
	return tw.CellConfig{
		Formatting: cfgFmt,
		Padding:    tw.CellPadding{Global: tw.Padding{Left: " ", Right: " "}},
	}
}

func getRowConfig(alignFlag string, wrapFlag string) tw.CellConfig {
	cfgFmt := tw.CellFormatting{}
	switch strings.ToLower(alignFlag) {
	case "left":
		cfgFmt.Alignment = tw.AlignLeft
	case "right":
		cfgFmt.Alignment = tw.AlignRight
	case "center":
		cfgFmt.Alignment = tw.AlignCenter
	default:
		cfgFmt.Alignment = tw.AlignLeft
	}
	switch strings.ToLower(wrapFlag) {
	case "truncate":
		cfgFmt.AutoWrap = tw.WrapTruncate
	case "break":
		cfgFmt.AutoWrap = tw.WrapBreak
	case "none":
		cfgFmt.AutoWrap = tw.WrapNone
	default:
		cfgFmt.AutoWrap = tw.WrapNormal
	}
	return tw.CellConfig{
		Formatting: cfgFmt,
		Padding:    tw.CellPadding{Global: tw.Padding{Left: " ", Right: " "}},
	}
}

func getFooterConfig() tw.CellConfig { // Footer doesn't currently take wrap/align flags from CLI
	return tw.CellConfig{
		Formatting: tw.CellFormatting{Alignment: tw.AlignRight, AutoWrap: tw.WrapNormal},
	}
}

func isGraphicalRenderer(rendererName string) bool {
	name := strings.ToLower(rendererName)
	return name == "html" || name == "svg"
}

//func exit(err error) {
//	// Using logger.Error instead of Fprintf for consistent logging if ll is used elsewhere
//	logger.Error("Error: %v", err)
//	os.Exit(1)
//}
