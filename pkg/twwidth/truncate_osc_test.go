package twwidth

import "testing"

func TestTruncateOSCStringTerminator(t *testing.T) {
	const osc = "\x1b]0;title\x1b\\"
	const want = osc + "abc\x1b[0m"
	if got := Truncate(osc+"abcdef", 3); got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}
