package tests

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"testing"

	"github.com/olekukonko/tablewriter"
	"github.com/olekukonko/tablewriter/renderer"
	"github.com/olekukonko/tablewriter/tw"
)

type failingTableWriter struct {
	bytes.Buffer
	failAt int
	calls  int
	err    error
	short  bool
}

func (w *failingTableWriter) Write(p []byte) (int, error) {
	w.calls++
	if w.failAt > 0 && w.calls >= w.failAt {
		if w.short {
			return len(p) - 1, nil
		}
		return 0, w.err
	}
	return w.Buffer.Write(p)
}

func TestRenderPropagatesWriteErrors(t *testing.T) {
	broken := errors.New("output unavailable")
	for _, kind := range []string{"blueprint", "markdown", "html", "svg"} {
		for _, caption := range []tw.Spot{tw.SpotNone, tw.SpotTopLeft, tw.SpotBottomLeft} {
			for _, short := range []bool{false, true} {
				t.Run(fmt.Sprintf("%s/caption=%d/short=%t", kind, caption, short), func(t *testing.T) {
					w := &failingTableWriter{failAt: 2, err: broken, short: short}
					var opts []tablewriter.Option
					switch kind {
					case "markdown":
						opts = append(opts, tablewriter.WithRenderer(renderer.NewMarkdown()))
					case "html":
						opts = append(opts, tablewriter.WithRenderer(renderer.NewHTML()))
					case "svg":
						opts = append(opts, tablewriter.WithRenderer(renderer.NewSVG()))
					}
					table := tablewriter.NewTable(w, opts...)
					table.Header("name")
					if err := table.Append("test"); err != nil {
						t.Fatal(err)
					}
					if caption != tw.SpotNone {
						table.Caption(tw.Caption{Text: "caption", Spot: caption})
					}
					// SVG emits its complete document in one write.
					if kind == "svg" {
						w.failAt = 1
					}
					want := broken
					if short {
						want = io.ErrShortWrite
					}
					if err := table.Render(); !errors.Is(err, want) {
						t.Fatalf("Render() error=%v, want %v", err, want)
					}
					if w.calls != w.failAt {
						t.Fatalf("writer called %d times, want stop at %d", w.calls, w.failAt)
					}
					// Each Render must use a fresh error state and retain the caller's writer.
					w.failAt = 0
					w.Buffer.Reset()
					if err := table.Render(); err != nil {
						t.Fatalf("retry Render: %v", err)
					}
					if w.Len() == 0 {
						t.Fatal("retry produced no output")
					}
				})
			}
		}
	}
}

// A cleanup error must not hide an earlier failure to write the table.
type closeErrorRenderer struct {
	tw.Renderer
	err error
}

func (r closeErrorRenderer) Close() error {
	_ = r.Renderer.Close()
	return r.err
}
func TestRenderWriteErrorBeforeCloseError(t *testing.T) {
	writeErr := errors.New("write failed")
	closeErr := errors.New("close failed")
	for _, failAt := range []int{0, 1} {
		w := &failingTableWriter{failAt: failAt, err: writeErr}
		r := closeErrorRenderer{Renderer: renderer.NewBlueprint(), err: closeErr}
		table := tablewriter.NewTable(w, tablewriter.WithRenderer(r))
		table.Header("name")
		if err := table.Append("test"); err != nil {
			t.Fatal(err)
		}
		want := closeErr
		if failAt > 0 {
			want = writeErr
		}
		if err := table.Render(); !errors.Is(err, want) {
			t.Errorf("Render error=%v, want %v", err, want)
		}
	}
}
