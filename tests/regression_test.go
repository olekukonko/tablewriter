package tests

import (
	"bytes"
	"strings"
	"testing"

	"github.com/olekukonko/tablewriter"
	"github.com/olekukonko/tablewriter/pkg/twwidth"
	"github.com/olekukonko/tablewriter/tw"
)

// / Regression Test
// https://github.com/olekukonko/tablewriter/pull/335
func TestPR335AutoFormatWrapBreakPreservesContentAtGlobalWidth20(t *testing.T) {
	var buf bytes.Buffer
	st := createStreamTable(t, &buf, tablewriter.WithConfig(tablewriter.Config{
		Header: tw.CellConfig{Formatting: tw.CellFormatting{AutoWrap: tw.WrapBreak, AutoFormat: tw.On}},
		Stream: tw.StreamConfig{Enable: true},
		Widths: tw.CellWidth{Global: 20},
	}), tablewriter.WithDebug(false))

	if err := st.Start(); err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	st.Header([]string{"x", "fooBarBazQux"})
	if err := st.Close(); err != nil {
		t.Fatalf("Close failed: %v", err)
	}

	for _, line := range strings.Split(buf.String(), "\n") {
		if width := twwidth.Width(line); width > 20 {
			t.Errorf("rendered line is %d columns wide, exceeding the configured width of 20: %q", width, line)
		}
	}

	var secondColumn strings.Builder
	for _, line := range strings.Split(buf.String(), "\n") {
		if !strings.Contains(line, "│") {
			continue
		}
		cells := strings.Split(strings.Trim(line, "│"), "│")
		if len(cells) < 2 {
			continue
		}
		secondColumn.WriteString(strings.TrimSpace(cells[1]))
	}
	got := strings.NewReplacer(" ", "", tw.CharBreak, "").Replace(secondColumn.String())
	if got != "FOOBARBAZQUX" {
		t.Fatalf("formatted content was not preserved: got %q; table:\n%s", got, buf.String())
	}
}
