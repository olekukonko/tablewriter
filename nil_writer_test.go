package tablewriter

import "testing"

func TestNewTableNilWriter(t *testing.T) {
	// must not panic
	tbl := NewTable(nil)
	tbl.Append([]string{"a", "b"})
	if err := tbl.Render(); err != nil {
		// Render may error depending on config; panic is the failure mode we care about
		t.Log(err)
	}
}
