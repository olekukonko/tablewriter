package tw

import (
	"io"
)

// Streamer defines the interface for renderers that support
// incremental, row-by-row output without batching, typically requiring
// pre-defined column widths or other fixed layout information.
type Streamer interface {
	// Start initializes the stream and writes any leading elements (e.g., top border, <table>).
	// It should also reset any internal state of the renderer.
	Start(w io.Writer) error

	// Header renders a single header row. May be called multiple times
	// for multi-line headers if the specific implementation supports it.
	// Should handle drawing separators below if configured.
	Header(w io.Writer, header []string) error

	// Row RenderRow renders a single data row.
	Row(w io.Writer, row []string) error

	// Footer renders a single footer row. May be called multiple times.
	// Should handle drawing separators above if configured.
	Footer(w io.Writer, footer []string) error

	// End finalizes the stream and writes any trailing elements (e.g., bottom border, </table>).
	End(w io.Writer) error

	// Debug returns trace information accumulated during rendering.
	Debug() []string

	// Reset clears internal state. Typically called by Start.
	Reset()

	// Config provides basic configuration details back to the core if needed,
	// though its utility is limited for streaming renderers compared to batch renderers.
	// It might return info like Symbols or Border settings used.
	Config() RendererConfig
}
