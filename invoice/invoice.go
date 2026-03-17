package invoice

import (
	"fmt"

	"github.com/johnfercher/maroto/v2"
	"github.com/johnfercher/maroto/v2/pkg/components/col"
	"github.com/johnfercher/maroto/v2/pkg/components/row"
	"github.com/johnfercher/maroto/v2/pkg/components/text"
	"github.com/johnfercher/maroto/v2/pkg/config"
	"github.com/johnfercher/maroto/v2/pkg/consts/align"
	"github.com/johnfercher/maroto/v2/pkg/consts/fontstyle"
	"github.com/johnfercher/maroto/v2/pkg/props"
)

// Invoice represents an invoice with line items.
type Invoice struct {
	Number  string
	Date    string
	DueDate string
	Company Company
	Client  Company
	Items   []Item
}

// Company represents a business entity.
type Company struct {
	Name    string
	Address string
	Email   string
}

// Item represents an invoice line item.
type Item struct {
	Description string
	Quantity    int
	UnitPrice   float64
}

// Total calculates the invoice total.
func (inv Invoice) Total() float64 {
	var total float64
	for _, item := range inv.Items {
		total += float64(item.Quantity) * item.UnitPrice
	}
	return total
}

// SampleInvoice returns hardcoded sample data for demonstration.
func SampleInvoice() Invoice {
	return Invoice{
		Number:  "INV-001",
		Date:    "2026-03-17",
		DueDate: "2026-04-17",
		Company: Company{
			Name:    "Acme Corp",
			Address: "123 Main St, New York, NY 10001",
			Email:   "billing@acme.com",
		},
		Client: Company{
			Name:    "Client Inc",
			Address: "456 Oak Ave, San Francisco, CA 94102",
			Email:   "accounts@client.com",
		},
		Items: []Item{
			{Description: "Web Development", Quantity: 40, UnitPrice: 150.00},
			{Description: "UI/UX Design", Quantity: 20, UnitPrice: 120.00},
			{Description: "Server Setup", Quantity: 5, UnitPrice: 200.00},
		},
	}
}

// GeneratePDF creates a PDF document from an Invoice and returns the bytes.
func GeneratePDF(inv Invoice) ([]byte, error) {
	cfg := config.NewBuilder().Build()

	m := maroto.New(cfg)

	m.AddRows(
		row.New(20).Add(
			col.New(12).Add(
				text.New(inv.Company.Name, props.Text{
					Size:  16,
					Style: fontstyle.Bold,
					Align: align.Left,
				}),
			),
		),
		row.New(8).Add(
			col.New(12).Add(
				text.New(inv.Company.Address, props.Text{Size: 10, Align: align.Left}),
			),
		),
		row.New(8).Add(
			col.New(12).Add(
				text.New(inv.Company.Email, props.Text{Size: 10, Align: align.Left}),
			),
		),
	)

	m.AddRows(
		row.New(15).Add(
			col.New(6).Add(
				text.New(fmt.Sprintf("Invoice: %s", inv.Number), props.Text{
					Size:  12,
					Style: fontstyle.Bold,
					Align: align.Left,
				}),
			),
			col.New(6).Add(
				text.New(fmt.Sprintf("Date: %s", inv.Date), props.Text{Size: 10, Align: align.Right}),
			),
		),
		row.New(8).Add(
			col.New(6).Add(
				text.New(fmt.Sprintf("Bill To: %s", inv.Client.Name), props.Text{Size: 10, Align: align.Left}),
			),
			col.New(6).Add(
				text.New(fmt.Sprintf("Due: %s", inv.DueDate), props.Text{Size: 10, Align: align.Right}),
			),
		),
		row.New(8).Add(
			col.New(12).Add(
				text.New(inv.Client.Address, props.Text{Size: 10, Align: align.Left}),
			),
		),
	)

	// Header row
	m.AddRows(
		row.New(10).Add(
			col.New(6).Add(text.New("Description", props.Text{Size: 10, Style: fontstyle.Bold})),
			col.New(2).Add(text.New("Qty", props.Text{Size: 10, Style: fontstyle.Bold, Align: align.Center})),
			col.New(2).Add(text.New("Price", props.Text{Size: 10, Style: fontstyle.Bold, Align: align.Right})),
			col.New(2).Add(text.New("Total", props.Text{Size: 10, Style: fontstyle.Bold, Align: align.Right})),
		),
	)

	// Line items
	for _, item := range inv.Items {
		lineTotal := float64(item.Quantity) * item.UnitPrice
		m.AddRows(
			row.New(8).Add(
				col.New(6).Add(text.New(item.Description, props.Text{Size: 10})),
				col.New(2).Add(text.New(fmt.Sprintf("%d", item.Quantity), props.Text{Size: 10, Align: align.Center})),
				col.New(2).Add(text.New(fmt.Sprintf("$%.2f", item.UnitPrice), props.Text{Size: 10, Align: align.Right})),
				col.New(2).Add(text.New(fmt.Sprintf("$%.2f", lineTotal), props.Text{Size: 10, Align: align.Right})),
			),
		)
	}

	// Total
	m.AddRows(
		row.New(15).Add(
			col.New(8),
			col.New(2).Add(text.New("Total:", props.Text{Size: 12, Style: fontstyle.Bold, Align: align.Right})),
			col.New(2).Add(text.New(fmt.Sprintf("$%.2f", inv.Total()), props.Text{Size: 12, Style: fontstyle.Bold, Align: align.Right})),
		),
	)

	doc, err := m.Generate()
	if err != nil {
		return nil, err
	}

	return doc.GetBytes(), nil
}
