package tests

import (
	"bytes"
	"strings"
	"testing"

	"github.com/fatih/color"
	"github.com/olekukonko/tablewriter"
	"github.com/olekukonko/tablewriter/renderer"
	"github.com/olekukonko/tablewriter/tw"
)

// withColor forces fatih/color output on for the duration of a test, since the
// test output is not a TTY and color is otherwise disabled.
func withColor(t *testing.T) {
	t.Helper()
	prev := color.NoColor
	color.NoColor = false
	t.Cleanup(func() { color.NoColor = prev })
}

func TestRGBApply(t *testing.T) {
	withColor(t)

	// Apply wraps the text in an opening SGR sequence and a reset. The opening
	// sequence is what these helpers construct; the exact reset is produced by
	// fatih/color, so the tests assert the opening sequence and that the text
	// itself is preserved.
	tests := []struct {
		name     string
		tint     renderer.Tint
		wantOpen string
	}{
		{
			name:     "foreground",
			tint:     renderer.Tint{FG: renderer.RGB(255, 128, 0)},
			wantOpen: "\x1b[38;2;255;128;0m",
		},
		{
			name:     "background",
			tint:     renderer.Tint{BG: renderer.BgRGB(0, 0, 0)},
			wantOpen: "\x1b[48;2;0;0;0m",
		},
		{
			name:     "foreground and background",
			tint:     renderer.Tint{FG: renderer.RGB(255, 128, 0), BG: renderer.BgRGB(0, 0, 0)},
			wantOpen: "\x1b[38;2;255;128;0;48;2;0;0;0m",
		},
		{
			name:     "combined with attribute via append",
			tint:     renderer.Tint{FG: append(renderer.RGB(255, 128, 0), color.Bold)},
			wantOpen: "\x1b[38;2;255;128;0;1m",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.tint.Apply("hi")
			if !strings.HasPrefix(got, tc.wantOpen) {
				t.Errorf("Apply() = %q, want prefix %q", got, tc.wantOpen)
			}
			if stripped := StripColors(got); stripped != "hi" {
				t.Errorf("Apply() stripped = %q, want %q", stripped, "hi")
			}
		})
	}
}

func TestRGBClamping(t *testing.T) {
	tests := []struct {
		name                string
		r, g, b             int
		wantR, wantG, wantB color.Attribute
	}{
		{"in range", 10, 20, 30, 10, 20, 30},
		{"upper bound", 255, 255, 255, 255, 255, 255},
		{"lower bound", 0, 0, 0, 0, 0, 0},
		{"above max clamps to 255", 300, 256, 999, 255, 255, 255},
		{"below min clamps to 0", -1, -50, -999, 0, 0, 0},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := renderer.RGB(tc.r, tc.g, tc.b)
			want := renderer.Colors{38, 2, tc.wantR, tc.wantG, tc.wantB}
			if !colorsEqual(got, want) {
				t.Errorf("RGB(%d,%d,%d) = %v, want %v", tc.r, tc.g, tc.b, got, want)
			}
		})
	}
}

func TestHex(t *testing.T) {
	tests := []struct {
		name    string
		in      string
		want    renderer.Colors
		wantErr bool
	}{
		{"full with hash", "#FF8000", renderer.Colors{38, 2, 255, 128, 0}, false},
		{"full without hash", "FF8000", renderer.Colors{38, 2, 255, 128, 0}, false},
		{"lowercase", "#ff8000", renderer.Colors{38, 2, 255, 128, 0}, false},
		{"shorthand", "#f80", renderer.Colors{38, 2, 255, 136, 0}, false},
		{"surrounding space", "  #FF8000  ", renderer.Colors{38, 2, 255, 128, 0}, false},
		{"black", "#000000", renderer.Colors{38, 2, 0, 0, 0}, false},
		{"white", "#ffffff", renderer.Colors{38, 2, 255, 255, 255}, false},
		{"empty", "", nil, true},
		{"too short", "#12", nil, true},
		{"wrong length", "#12345", nil, true},
		{"non-hex digit", "#gggggg", nil, true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := renderer.Hex(tc.in)
			if (err != nil) != tc.wantErr {
				t.Fatalf("Hex(%q) error = %v, wantErr = %v", tc.in, err, tc.wantErr)
			}
			if !colorsEqual(got, tc.want) {
				t.Errorf("Hex(%q) = %v, want %v", tc.in, got, tc.want)
			}
		})
	}
}

func TestBgHex(t *testing.T) {
	got, err := renderer.BgHex("#FF8000")
	if err != nil {
		t.Fatalf("BgHex returned error: %v", err)
	}
	want := renderer.Colors{48, 2, 255, 128, 0}
	if !colorsEqual(got, want) {
		t.Errorf("BgHex(#FF8000) = %v, want %v", got, want)
	}
}

// TestRGBColorizedTable checks that RGB tints survive rendering through the
// colorized renderer end to end.
func TestRGBColorizedTable(t *testing.T) {
	withColor(t)

	var buf bytes.Buffer
	table := tablewriter.NewTable(&buf,
		tablewriter.WithRenderer(renderer.NewColorized(renderer.ColorizedConfig{
			Borders: tw.Border{Left: tw.Off, Right: tw.Off, Top: tw.Off, Bottom: tw.Off},
			Header:  renderer.Tint{FG: renderer.RGB(255, 128, 0)},
			Column:  renderer.Tint{FG: renderer.RGB(0, 200, 100)},
		})),
	)
	table.Header([]string{"Name"})
	table.Append([]string{"Alice"})
	table.Render()

	out := buf.String()
	if !strings.Contains(out, "\x1b[38;2;255;128;0m") {
		t.Errorf("expected header RGB sequence in output, got:\n%q", out)
	}
	if !strings.Contains(out, "\x1b[38;2;0;200;100m") {
		t.Errorf("expected column RGB sequence in output, got:\n%q", out)
	}
	if stripped := StripColors(out); !strings.Contains(stripped, "Alice") {
		t.Errorf("expected cell content preserved, got:\n%q", stripped)
	}
}

func colorsEqual(a, b renderer.Colors) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
