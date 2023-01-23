package tablewriter

import (
	"io"
	"os"
	"golang.org/x/net/html"
)

func NewHTML(writer io.Writer, fileName string, columnNum int) (*Table, error) {
	file, err := os.Open(fileName)

	if err != nil {
		return  &Table{}, err
	}
	defer file.Close()
	
	htmlParser := html.NewTokenizer(file) 

	var isTh bool
	var isTd bool
	var depth int
	var headings []string
	var cellData []string
	table := NewWriter(writer)
	
	for {
		tt := htmlParser.Next()
		switch {
		case tt == html.ErrorToken:
			return  table, nil
		case tt == html.StartTagToken:
			t := htmlParser.Token()
			isTd = t.Data == "td"
			isTh = t.Data == "th"
		case tt == html.TextToken:
			t := htmlParser.Token()
			if isTd {
				cellData = append(cellData, t.Data)
				depth++
			}
			if isTh {
				headings = append(headings, t.Data)
				depth++
			}
			if isTd && depth % columnNum == 0 {
				table.Append(cellData)
				cellData = nil
			}
			if isTh && depth % columnNum == 0 {
				table.SetHeader(headings)
			}
			isTd = false
			isTh = false
		}
	}
}