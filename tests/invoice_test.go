package tests

//import (
//	"bytes"
//	"github.com/olekukonko/tablewriter/renderer"
//	"github.com/olekukonko/tablewriter/symbols"
//	"testing"
//)
//
//func TestInvoiceRendererFlexibleFooter(t *testing.T) {
//	data := [][]string{
//		{"1/1/2014", "Domain name", "2233", "$10.98"},
//		{"1/1/2014", "January Hosting", "2233", "$54.95"},
//		{"", "    (empty)\n    (empty)", "", ""},
//		{"1/4/2014", "February Hosting", "2233", "$51.00"},
//		{"1/4/2014", "February Extra Bandwidth", "2233", "$30.00"},
//		{"1/4/2014", "    (Discount)", "2233", "-$1.00"},
//	}
//
//	// Test with ASCII and "TOTAL" footer
//	var bufASCII bytes.Buffer
//	tableASCII := NewTable(&bufASCII, WithRenderer(renderer.NewInvoice()))
//	tableASCII.Header([]string{"DATE", "DESCRIPTION", "CV2", "AMOUNT"})
//	tableASCII.Bulk(data)
//	tableASCII.Footer([]string{"", "", "", "TOTAL | $145.93"})
//	tableASCII.Render()
//
//	expected := `
//
//		DATE      | DESCRIPTION             | CV2  | AMOUNT
//		---------+-------------------------+------+-------
//		1/1/2014  | Domain name             | 2233 | $10.98
//		1/1/2014  | January Hosting         | 2233 | $54.95
//				  | (empty)                 |      |
//				  | (empty)                 |      |
//		1/4/2014  | February Hosting        | 2233 | $51.00
//		1/4/2014  | February Extra Bandwidth| 2233 | $30.00
//		1/4/2014  | (Discount)              | 2233 | -$1.00
//		---------+-------------------------+------+-------
//											 TOTAL | $145.93
//
//    `
//	visualCheck(t, "InvoiceRendererASCII", bufASCII.String(), expected)
//
//	// Test with Credit and Debit
//	var bufMulti bytes.Buffer
//	tableMulti := NewTable(&bufMulti, WithRenderer(renderer.NewInvoice(renderer.InvoiceConfig{
//		Symbols: symbols.NewSymbols(symbols.StyleLight),
//	})))
//
//	tableMulti.Header([]string{"DATE", "DESCRIPTION", "CV2", "AMOUNT"})
//	tableMulti.Bulk(data)
//	tableMulti.Footer([]string{"", "Credit", "", "$50.00"})
//	tableMulti.Render()
//
//	expectedMulti := `
//
//		DATE      │ DESCRIPTION             │ CV2  │ AMOUNT
//		 ─────────┼─────────────────────────┼──────┼───────
//		1/1/2014  │ Domain name             │ 2233 │ $10.98
//		1/1/2014  │ January Hosting         │ 2233 │ $54.95
//				  │ (empty)                 │      │
//				  │ (empty)                 │      │
//		1/4/2014  │ February Hosting        │ 2233 │ $51.00
//		1/4/2014  │ February Extra Bandwidth│ 2233 │ $30.00
//		1/4/2014  │ (Discount)              │ 2233 │ -$1.00
//		 ─────────┼─────────────────────────┼──────┼───────
//				  │ Credit                  │      │ $50.00
//
//    `
//	visualCheck(t, "InvoiceRendererMulti", bufMulti.String(), expectedMulti)
//
//	// Test with StyleLight and "Credit/Debit" footer
//	// Test with empty footer
//	var bufEmpty bytes.Buffer
//	tableEmpty := NewTable(&bufEmpty, WithRenderer(renderer.NewInvoice()))
//	tableEmpty.Header([]string{"DATE", "DESCRIPTION", "CV2", "AMOUNT"})
//	tableEmpty.Bulk(data)
//	tableEmpty.Footer([]string{"", "", "", ""})
//	tableEmpty.Render()
//
//	expected = `
//
//			DATE      | DESCRIPTION             | CV2  | AMOUNT
//			---------+-------------------------+------+-------
//			1/1/2014  | Domain name             | 2233 | $10.98
//			1/1/2014  | January Hosting         | 2233 | $54.95
//					  | (empty)                 |      |
//					  | (empty)                 |      |
//			1/4/2014  | February Hosting        | 2233 | $51.00
//			1/4/2014  | February Extra Bandwidth| 2233 | $30.00
//			1/4/2014  | (Discount)              | 2233 | -$1.00
//
//    `
//	visualCheck(t, "InvoiceRendererEmpty", bufEmpty.String(), expected)
//
//}
