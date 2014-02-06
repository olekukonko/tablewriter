package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"github.com/olekukonko/tablewriter"
	"os"
	"unicode/utf8"
)

var (
	fileName  = flag.String("f", "", "Set file with  eg. sample.csv")
	delimiter = flag.String("d", ",", "Set CSV File delimiter eg. ,|;|\t ")
	header    = flag.Bool("h", true, "Set header options eg. true|false ")
	align     = flag.String("a", "none", "Set aligmement with eg. none|left|right|centre")
)

func main() {
	flag.Parse()

	fmt.Println()

	if *fileName == "" {
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
		flag.PrintDefaults()
		fmt.Println()
		os.Exit(1)
	}

	process()
	fmt.Println()

}

func process() {
	file, err := os.Open(*fileName)
	if err != nil {
		exit(err)
	}
	defer file.Close()
	csvReader := csv.NewReader(file)

	rune, size := utf8.DecodeRuneInString(*delimiter)

	if size == 0 {
		rune = ','
	}
	csvReader.Comma = rune

	table, err := tablewriter.NewCSVReader(os.Stdout, csvReader, *header)

	if err != nil {
		exit(err)
	}

	switch *align {
	case "left":
		table.SetAlignment(tablewriter.ALIGN_LEFT)
	case "right":
		table.SetAlignment(tablewriter.ALIGN_RIGHT)
	case "center":
		table.SetAlignment(tablewriter.ALIGN_CENTRE)
	}

	table.Render()
}

func exit(err error) {
	fmt.Fprintf(os.Stderr, "#Error : %s", err)
	os.Exit(1)
}
