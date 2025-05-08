package tablewriter

import (
	"fmt" // Added for potential future errors/warnings
	"github.com/olekukonko/errors"
	"github.com/olekukonko/tablewriter/tw"
	"io"
)

// TableStream provides an interface for incrementally rendering table rows
// using a compatible Streamer. It requires pre-defined column widths
// or other configuration needed by the specific streaming renderer.
type TableStream struct {
	writer   io.Writer
	renderer tw.Streamer // Must be a streaming-capable renderer
}

// NewStreamTable creates a new table configured for streaming output.
// It takes the output writer and a pre-configured Streamer.
func NewStreamTable(w io.Writer, renderer tw.Streamer) (*TableStream, error) {
	if w == nil {
		return nil, errors.New("NewStreamTable requires a non-nil writer")
	}
	if renderer == nil {
		return nil, errors.New("NewStreamTable requires a non-nil Streamer")
	}
	st := &TableStream{
		writer:   w,
		renderer: renderer,
	}
	return st, nil
}

// Start initializes the table stream and renders the top border/opening tags.
// This should be called once before streaming any headers or rows.
func (st *TableStream) Start() error {
	return st.renderer.Start(st.writer)
}

// Header renders a single header row. Can be called multiple times if needed,
// although most streaming renderers will assume a single primary header row.
func (st *TableStream) Header(header []string) error {
	if header == nil {
		return fmt.Errorf("header received nil slice")
	}
	return st.renderer.Header(st.writer, header)
}

// Row renders a single data row. Call this repeatedly for each row in the stream.
func (st *TableStream) Row(row []string) error {
	if row == nil {
		return fmt.Errorf("row received nil slice")
	}
	return st.renderer.Row(st.writer, row)
}

// Footer renders a single footer row. Can be called multiple times.
func (st *TableStream) Footer(footer []string) error {
	if footer == nil {
		return fmt.Errorf("footer received nil slice")
	}
	return st.renderer.Footer(st.writer, footer)
}

// End finalizes the table stream, rendering the bottom border/closing tags.
// This should be called once after all headers, rows, and footers have been streamed.
func (st *TableStream) End() error {
	return st.renderer.End(st.writer)
}

// Debug returns the debug trace from the underlying renderer.
func (st *TableStream) Debug() []string {
	if st.renderer != nil {
		return st.renderer.Debug()
	}
	return nil
}
