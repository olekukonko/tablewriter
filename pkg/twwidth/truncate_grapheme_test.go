package twwidth

import "testing"

func TestTruncateGraphemeClusters(t *testing.T) {
	for _, cluster := range []string{"👨‍👩‍👧‍👦", "🇬🇧", "❤️", "1️⃣"} {
		t.Run(cluster, func(t *testing.T) {
			if got := Truncate(cluster+"abc", Width(cluster)); got != cluster {
				t.Fatalf("got %q, want %q", got, cluster)
			}
			if got := Truncate(cluster+"abc", Width(cluster)+1, "."); got != cluster+"." {
				t.Fatalf("with suffix: got %q, want %q", got, cluster+".")
			}
			want := "\x1b[31m" + cluster + "\x1b[0m"
			if got := Truncate("\x1b[31m"+cluster+"abc", Width(cluster)); got != want {
				t.Fatalf("with ANSI: got %q, want %q", got, want)
			}
		})
	}
}
