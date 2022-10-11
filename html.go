package tablewriter

import (
	"io"
	"os"
	"golang.org/x/net/html"
)

func NewHTML(writer io.Writer, fileName string, headingsNum int) (*Table, error) {
	file, err := os.Open(fileName)

	if err != nil {
		return  &Table{}, err
	}
	defer file.Close()
	// Parse the html
	htmlParser := html.NewTokenizer(file) // file must utf-8 encoded html or error

	var isTh bool
	var isTd bool
	var n int // n is the number of th tags in the html file
	var depth int
	var headings []string
	var records []string
	table := NewWriter(writer)
	
	// start a loop tokenizing the html
	for {
		tt := htmlParser.Next()
		switch {
		case tt == html.ErrorToken:
			return  &Table{}, nil
		// if the token gets to start tag that is td or th, make isTd and isTh true
		case tt == html.StartTagToken:
			t := htmlParser.Token()
			isTd = t.Data == "td"
			isTh = t.Data == "th"
		case tt == html.TextToken:
			t := htmlParser.Token()
			if isTd {
				records = append(records, t.Data)
				depth++	
			}
			if isTh {
				headings = append(headings, t.Data)
				n++
			}
			if isTd && depth % headingsNum == 0 {
				table.Append(records)
				records = nil
			}
			if isTh && n % headingsNum == 0 {
				table.SetHeader(headings)
			}
			isTd = false
			isTh = false
		}
	}
}