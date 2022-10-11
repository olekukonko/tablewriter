package tablewriter

import (
	"io"
	"os"
	"golang.org/x/net/html"
)

func NewHTML(writer io.Writer, fileName string, hasHeader bool) (*Table, error) {
	file, err := os.Open(fileName)

	if err != nil {
		return  &Table{}, err
	}
	defer file.Close()
	// Parse the html
	htmlParser := html.NewTokenizer(file)

	var isTh bool
	var isTd bool
	var n int
	var headings []string
	var records []string
	table := NewWriter(writer)

	// start a loop tokenizing the html
	for {
		tt := htmlParser.Next()
		switch {
		case tt == html.ErrorToken:
			return  &Table{}, err
		// if the token gets to start tag that is td or th, make isTd and isTh true
		case tt == html.StartTagToken:
			t := htmlParser.Token()
			isTd = t.Data == "td"
			isTh = t.Data == "th"
		case tt == html.TextToken:
			t := htmlParser.Token()
			if isTd {
				records = append(records, t.Data)
				n++
			}
			if isTh {
				headings = append(headings, t.Data)
				n++
			}
			if isTd && n % 4 == 0 {
				table.Append(records)
				records = nil
			}
			isTd = false
			isTh = false
		}
	}

	if hasHeader {
		table.SetHeader(headings)
	}

	return table, nil
}