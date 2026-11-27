package twwidth

import "testing"

func TestTruncateClosesHyperlink(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		width  int
		suffix string
		want   string
	}{
		{
			name:  "ST terminator",
			input: "\x1b]8;;https://example.com\x1b\\example.com\x1b]8;;\x1b\\",
			width: 4,
			want:  "\x1b]8;;https://example.com\x1b\\exam\x1b]8;;\x1b\\\x1b[0m",
		},
		{
			name:  "BEL terminator",
			input: "\x1b]8;;https://example.com\x07example.com\x1b]8;;\x07",
			width: 4,
			want:  "\x1b]8;;https://example.com\x07exam\x1b]8;;\x07\x1b[0m",
		},
		{
			name:   "suffix stays outside the link",
			input:  "\x1b]8;id=1;https://example.com\x1b\\example.com\x1b]8;;\x1b\\",
			width:  5,
			suffix: "…",
			want:   "\x1b]8;id=1;https://example.com\x1b\\exam\x1b]8;;\x1b\\\x1b[0m…",
		},
		{
			name:  "link closed before the cut",
			input: "\x1b]8;;https://example.com\x1b\\ex\x1b]8;;\x1b\\ample",
			width: 4,
			want:  "\x1b]8;;https://example.com\x1b\\ex\x1b]8;;\x1b\\am\x1b[0m",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got string
			if tt.suffix != "" {
				got = Truncate(tt.input, tt.width, tt.suffix)
			} else {
				got = Truncate(tt.input, tt.width)
			}
			if got != tt.want {
				t.Fatalf("got %q, want %q", got, tt.want)
			}
		})
	}
}
