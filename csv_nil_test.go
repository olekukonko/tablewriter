package tablewriter

import (
	"bytes"
	"testing"
)

func TestNewCSVReaderNil(t *testing.T) {
	var buf bytes.Buffer
	tbl, err := NewCSVReader(&buf, nil, true)
	if err == nil {
		t.Fatal("expected error for nil csv.Reader")
	}
	if tbl != nil {
		t.Fatal("expected nil table on error")
	}
}
