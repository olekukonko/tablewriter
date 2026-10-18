package renderer

import (
	"bytes"
	"strings"
	"testing"

	"github.com/olekukonko/tablewriter/tw"
)

func TestHTMLClassEscaping(t *testing.T) {
	const class = `a"&b`
	h := NewHTML(HTMLConfig{
		TableClass: class, HeaderClass: class, BodyClass: class, FooterClass: class,
		RowClass: class, HeaderRowClass: class, FooterRowClass: class,
	})
	var buf bytes.Buffer
	if err := h.Start(&buf); err != nil {
		t.Fatal(err)
	}
	h.Header([][]string{{"head"}}, tw.Formatting{})
	h.Row([]string{"body"}, tw.Formatting{})
	h.Footer([][]string{{"foot"}}, tw.Formatting{})
	if got := strings.Count(buf.String(), `class="a&#34;&amp;b"`); got != 7 {
		t.Fatalf("escaped %d classes, want 7: %s", got, buf.String())
	}
}
