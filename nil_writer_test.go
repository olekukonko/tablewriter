package tablewriter

import (
	"bytes"
	"strings"
	"testing"

	"github.com/olekukonko/tablewriter/tw"
)

func TestRenderNilWriter(t *testing.T) {
	tb := NewTable(nil)
	tb.Append([]string{"a", "b"})
	err := tb.Render()
	if err == nil || !strings.Contains(err.Error(), "nil") {
		t.Fatalf("expected nil writer error, got %v", err)
	}
}

func TestStartNilWriter(t *testing.T) {
	tb := NewTable(nil, WithStreaming(tw.StreamConfig{Enable: true}))
	err := tb.Start()
	if err == nil || !strings.Contains(err.Error(), "nil") {
		t.Fatalf("expected nil writer error, got %v", err)
	}
}

func TestRenderOK(t *testing.T) {
	var buf bytes.Buffer
	tb := NewTable(&buf)
	tb.Append([]string{"x", "y"})
	if err := tb.Render(); err != nil {
		t.Fatal(err)
	}
	if buf.Len() == 0 {
		t.Fatal("expected output")
	}
}
