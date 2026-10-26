package invoice_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Massad/gin-boilerplate/invoice"
	"github.com/gin-gonic/gin"
)

// Use an independent PDF reader: searching the raw bytes would miss compressed
// page content and would not prove that the text is present in the document.
func pdfLines(t *testing.T, data []byte) []string {
	t.Helper()
	if !bytes.HasPrefix(data, []byte("%PDF-")) {
		t.Fatal("invalid PDF signature")
	}
	tool, err := exec.LookPath("pdftotext")
	if err != nil {
		t.Fatal("PDF content tests require pdftotext (apt-get install poppler-utils or brew install poppler)")
	}
	path := filepath.Join(t.TempDir(), "invoice.pdf")
	if err := os.WriteFile(path, data, 0600); err != nil {
		t.Fatal(err)
	}
	var stderr bytes.Buffer
	cmd := exec.Command(tool, "-layout", path, "-")
	cmd.Stderr = &stderr
	output, err := cmd.Output()
	if err != nil {
		t.Fatalf("extract PDF text: %v: %s", err, stderr.String())
	}
	var lines []string
	for _, line := range strings.Split(string(output), "\n") {
		if normalized := strings.Join(strings.Fields(line), " "); normalized != "" {
			lines = append(lines, normalized)
		}
	}
	return lines
}

func assertPDFLines(t *testing.T, data []byte, want []string) {
	t.Helper()
	lines := pdfLines(t, data)
	for _, expected := range want {
		found := false
		for _, line := range lines {
			if line == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("PDF missing line %q; extracted text:\n%s", expected, strings.Join(lines, "\n"))
		}
	}
}

// Keep the expected values literal so a regression in Total or the renderer
// cannot change both the actual PDF and its expected result.
func samplePDFLines() []string {
	return []string{
		"Invoice: INV-001 Date: 2026-03-17",
		"Description Qty Price Total",
		"Web Development 40 $150.00 $6000.00",
		"UI/UX Design 20 $120.00 $2400.00",
		"Server Setup 5 $200.00 $1000.00",
		"Total: $9400.00",
	}
}

func TestGeneratePDFContent(t *testing.T) {
	cases := []struct {
		name string
		inv  invoice.Invoice
		want []string
	}{
		{"sample", invoice.SampleInvoice(), samplePDFLines()},
		{
			name: "custom invoice with fractional prices",
			inv: invoice.Invoice{
				Number: "INV-REGRESSION-042", Date: "2026-10-26", DueDate: "2026-11-26",
				Company: invoice.Company{Name: "Example Supplier"},
				Client:  invoice.Company{Name: "Example Customer"},
				Items: []invoice.Item{
					{Description: "Support (monthly)", Quantity: 3, UnitPrice: 19.95},
					{Description: "Hosting & backups", Quantity: 2, UnitPrice: 4.50},
				},
			},
			want: []string{
				"Invoice: INV-REGRESSION-042 Date: 2026-10-26",
				"Support (monthly) 3 $19.95 $59.85",
				"Hosting & backups 2 $4.50 $9.00",
				"Total: $68.85",
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			data, err := invoice.GeneratePDF(tc.inv)
			if err != nil {
				t.Fatal(err)
			}
			assertPDFLines(t, data, tc.want)
		})
	}
}

func TestDownloadPDFContent(t *testing.T) {
	router := gin.New()
	router.GET("/v1/invoice/download", new(invoice.InvoiceController).Download)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/v1/invoice/download", nil))
	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d: %s", response.Code, http.StatusOK, response.Body.String())
	}
	if got := response.Header().Get("Content-Type"); got != "application/pdf" {
		t.Errorf("Content-Type = %q, want application/pdf", got)
	}
	if got := response.Header().Get("Content-Disposition"); got != `attachment; filename="INV-001.pdf"` {
		t.Errorf("unexpected Content-Disposition: %q", got)
	}
	assertPDFLines(t, response.Body.Bytes(), samplePDFLines())
}
