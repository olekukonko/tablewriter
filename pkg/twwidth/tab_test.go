package twwidth_test

import (
	"fmt"
	"testing"

	"github.com/olekukonko/tablewriter/pkg/twwidth"
)

func TestTabinalSetWidth(t *testing.T) {
	t.Setenv("TABWIDTH", "3")

	for _, initialized := range []bool{false, true} {
		for _, width := range []int{1, 2, 32, 0, -1, 33} {
			t.Run(fmt.Sprintf("initialized=%v/width=%d", initialized, width), func(t *testing.T) {
				var tab twwidth.Tabinal
				if initialized {
					if got := tab.Size(); got != 3 {
						t.Fatalf("initial Size() = %d, want 3", got)
					}
				}

				tab.SetWidth(width)
				want := width
				if width <= 0 || width > 32 {
					want = 3
				}
				if got := tab.Size(); got != want {
					t.Fatalf("Size() after SetWidth(%d) = %d, want %d", width, got, want)
				}

				tab.SetWidth(5)
				tab.SetWidth(0)
				if got := tab.Size(); got != 5 {
					t.Fatalf("Size() after changing width = %d, want 5", got)
				}
			})
		}
	}
}
