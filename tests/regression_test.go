package tests

import (
	"bytes"
	"strings"
	"testing"
	"unicode/utf8"

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

// Regression Test
// With TrimSpace off the row wrapper preserves the spaces around the words, so
// the cell goes through twwarp.WrapStringWithSpaces. That function measured the
// last word on a rune cut in two, so a cell ending in a multi-byte character
// was wrapped to a line wider than the column and the renderer dropped the word.
func TestWrapNormalKeepsMultiByteLastWord(t *testing.T) {
	var buf bytes.Buffer
	columnWidths := tw.NewMapper[int, int]()
	columnWidths.Set(0, 7)

	table := tablewriter.NewTable(&buf, tablewriter.WithConfig(tablewriter.Config{
		Behavior: tw.Behavior{TrimSpace: tw.Off, TrimTab: tw.On},
		Widths:   tw.CellWidth{PerColumn: columnWidths},
		Row:      tw.CellConfig{Formatting: tw.CellFormatting{AutoWrap: tw.WrapNormal}},
	}))
	table.Header([]string{"col"})
	table.Append([]string{"aaa あ"})

	if err := table.Render(); err != nil {
		t.Fatalf("Render failed: %v", err)
	}

	rendered := buf.String()
	if !strings.Contains(rendered, "あ") {
		t.Fatalf("the last word of the cell did not reach the output; table:\n%s", rendered)
	}

	lines := strings.Split(strings.TrimRight(rendered, "\n"), "\n")
	want := twwidth.Width(lines[0])
	for i, line := range lines[1:] {
		if got := twwidth.Width(line); got != want {
			t.Errorf("line %d is %d columns wide, want %d; table:\n%s", i+1, got, want, rendered)
		}
	}
}

// Regression Test
// The last word of a cell that ends in a multi-byte character followed by
// spaces. The byte index of the last non-space rune points at its first byte,
// so the content of the cell was cut inside that rune and the spaces were read
// from the rest of it. The broken tail reached the wrapper as part of the last
// word, the word no longer fitted the column, and the renderer dropped it: the
// cell "aaa あ " was rendered as "aaa".
func TestWrapNormalKeepsMultiByteLastRuneBeforeTrailingSpaces(t *testing.T) {
	const columnWidth = 7
	for _, tc := range []struct {
		name string
		cell string
	}{
		{"three byte rune and one trailing space", "aaa \u3042 "},
		{"three byte rune and two trailing spaces", "aaa \u3042  "},
		{"four byte rune and one trailing space", "aaa \U0001F600 "},
		{"two byte rune and two trailing spaces", "caf\u00e9  "},
		{"two wide runes and two trailing spaces", "\u3042\u3042  "},
		{"rune only and one trailing space", "\u3042 "},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var buf bytes.Buffer
			columnWidths := tw.NewMapper[int, int]()
			columnWidths.Set(0, columnWidth)

			table := tablewriter.NewTable(&buf, tablewriter.WithConfig(tablewriter.Config{
				Behavior: tw.Behavior{TrimSpace: tw.Off, TrimTab: tw.On},
				Widths:   tw.CellWidth{PerColumn: columnWidths},
				Row:      tw.CellConfig{Formatting: tw.CellFormatting{AutoWrap: tw.WrapNormal}},
			}))
			table.Header([]string{"col"})
			table.Append([]string{tc.cell})

			if err := table.Render(); err != nil {
				t.Fatalf("Render failed: %v", err)
			}
			rendered := buf.String()
			if !utf8.ValidString(rendered) {
				t.Errorf("the output is not valid UTF-8: %q", rendered)
			}

			// Every rune of the cell is preserved, whole. The wrap may break
			// the line between two words and the trailing spaces may be
			// trimmed, so each word of the cell is looked for on its own: a
			// broken rune would not be found as part of any word.
			content := strings.TrimRight(tc.cell, " ")
			for _, word := range strings.Fields(content) {
				if !strings.Contains(rendered, word) {
					t.Errorf("the word %q did not reach the output whole; table:\n%s", word, rendered)
				}
			}

			// The data of the table is the row after the header separator, and
			// it holds one line per wrapped line of the cell.
			lines := strings.Split(strings.TrimRight(rendered, "\n"), "\n")
			rows := lines[3 : len(lines)-1]
			if len(rows) < 2 && twwidth.Width(content) > columnWidth {
				t.Errorf("the cell %q is wider than the column %d but was not wrapped; table:\n%s", content, columnWidth, rendered)
			}
			for i, row := range rows {
				for _, r := range content {
					if !strings.ContainsRune(row, r) && !strings.ContainsRune(strings.Join(rows, ""), r) {
						t.Errorf("the rune %q is not in the data of the table; table:\n%s", r, rendered)
					}
				}
				if got := twwidth.Width(row); got != twwidth.Width(lines[0]) {
					t.Errorf("row line %d is %d columns wide, want %d; table:\n%s", i+1, got, twwidth.Width(lines[0]), rendered)
				}
			}

			// The wrapped content stays inside the column, so every line of
			// the table is as wide as the top border.
			border := twwidth.Width(lines[0])
			if border > columnWidth+4 {
				t.Errorf("the table is %d columns wide, above %d; table:\n%s", border, columnWidth+4, rendered)
			}
			for i, line := range lines {
				if got := twwidth.Width(line); got != border {
					t.Errorf("line %d is %d columns wide, want %d; table:\n%s", i+1, got, border, rendered)
				}
			}
		})
	}
}
