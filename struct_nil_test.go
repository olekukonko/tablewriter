package tablewriter

import (
	"github.com/olekukonko/tablewriter/tw"
	"reflect"
	"testing"
)

func TestBulkNilEmbeddedPointer(t *testing.T) {
	type Location struct {
		Zip    int
		Hidden string `tw:"-"`
	}
	type Details struct {
		City string `tw:"name=Town"`
		*Location
	}
	type Row struct {
		Name string
		*Details
		Status string `tw:"align=right"`
	}
	alice := Row{"Alice", &Details{"Paris", &Location{75000, "hidden"}}, "active"}
	bob := Row{Name: "Bob", Status: "inactive"}
	for _, rows := range [][]Row{{alice, bob}, {bob, alice}} {
		tbl := NewTable(nil)
		tbl.Configure(func(cfg *Config) { cfg.Behavior.Structs.AutoHeader = tw.On })
		if err := tbl.Bulk(rows); err != nil {
			t.Fatal(err)
		}
		for i, row := range rows {
			want := []string{"Bob", "", "", "inactive"}
			if row.Name == "Alice" {
				want = []string{"Alice", "Paris", "75000", "active"}
			}
			if !reflect.DeepEqual(tbl.rows[i], want) {
				t.Errorf("row %d = %q, want %q", i, tbl.rows[i], want)
			}
		}
		headers := tbl.extractHeadersFromStruct(bob)
		if want := []string{"NAME", "Town", "ZIP", "STATUS"}; !reflect.DeepEqual(headers, want) {
			t.Errorf("headers = %q, want %q", headers, want)
		}
		if got := tbl.config.Row.Alignment.PerColumn; len(got) != 4 || got[3] != tw.AlignRight {
			t.Errorf("alignment = %v", got)
		}
	}
}

func TestNilRecursiveEmbeddedPointer(t *testing.T) {
	type Node struct {
		*Node
		Value int
	}
	tbl := NewTable(nil)
	headers, values := tbl.extractFieldsAndValuesFromStruct(struct {
		*Node
		Status string
	}{Status: "ok"})
	if want := []string{"VALUE", "STATUS"}; !reflect.DeepEqual(headers, want) {
		t.Fatalf("headers = %q, want %q", headers, want)
	}
	if want := []string{"", "ok"}; !reflect.DeepEqual(values, want) {
		t.Fatalf("values = %q, want %q", values, want)
	}
}
