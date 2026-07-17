package ticket

import (
	"bytes"
	"fmt"
	"time"

	"github.com/jung-kurt/gofpdf"
)

// TicketData holds the context needed to populate our beautiful PDF voucher.
type TicketData struct {
	BookingID   string
	EventTitle  string
	EventDate   time.Time
	UserName    string
	SeatsCount  int
	TicketPrice float64
}

// GenerateTicketPDF creates a highly-polished, professional ticket in memory
func GenerateTicketPDF(data TicketData) ([]byte, error) {
	// Initialize custom-sized PDF (Standard Ticket/Envelope size works best for single vouchers)
	// We'll use a wide Landscape card format: 200mm wide x 100mm high
	pdf := gofpdf.NewCustom(&gofpdf.InitType{
		UnitStr: "mm",
		Size:    gofpdf.SizeType{Wd: 200, Ht: 100},
	})

	pdf.SetMargins(0, 0, 0)
	pdf.AddPage()

	// 1. Draw a Header Accent Bar (Primary Brand Color: Deep Navy #1A365D)
	pdf.SetFillColor(26, 54, 93)
	pdf.Rect(0, 0, 200, 15, "F")

	// Header Text
	pdf.SetTextColor(255, 255, 255)
	pdf.SetFont("Helvetica", "B", 14)
	pdf.Text(10, 10, "OFFICIAL EVENT ADMISSION TICKET")

	// 2. Draw Side Stub Accent (Contrasting Accent Color: Coral/Gold #DD6B20)
	pdf.SetFillColor(221, 107, 32)
	pdf.Rect(165, 15, 35, 85, "F")

	// Draw dashed stub divider line
	pdf.SetDrawColor(200, 200, 200)
	pdf.SetLineWidth(0.5)
	pdf.Line(165, 15, 165, 100)

	// 3. Body Typography & Info (Left Content Area)
	pdf.SetTextColor(45, 55, 72) // Off-black/Charcoal for premium readability

	// Event Title (Big & Bold)
	pdf.SetFont("Helvetica", "B", 18)
	pdf.MoveTo(10, 25)
	pdf.MultiCell(145, 8, data.EventTitle, "", "L", false)

	// Details grid
	pdf.SetFont("Helvetica", "", 10)
	pdf.SetTextColor(113, 128, 150) // Slate gray secondary text

	// Left Column Labels & Values
	pdf.Text(10, 52, "ATTENDEE")
	pdf.SetFont("Helvetica", "B", 11)
	pdf.SetTextColor(45, 55, 72)
	pdf.Text(10, 57, data.UserName)

	pdf.SetFont("Helvetica", "", 10)
	pdf.SetTextColor(113, 128, 150)
	pdf.Text(10, 70, "DATE & TIME")
	pdf.SetFont("Helvetica", "B", 11)
	pdf.SetTextColor(45, 55, 72)
	pdf.Text(10, 75, data.EventDate.Format("January 02, 2006 at 3:04 PM"))

	// Right Column Labels & Values
	pdf.SetFont("Helvetica", "", 10)
	pdf.SetTextColor(113, 128, 150)
	pdf.Text(90, 52, "RESERVATION ID")
	pdf.SetFont("Helvetica", "B", 11)
	pdf.SetTextColor(45, 55, 72)
	pdf.Text(90, 57, fmt.Sprintf("#%s", data.BookingID))

	pdf.SetFont("Helvetica", "", 10)
	pdf.SetTextColor(113, 128, 150)
	pdf.Text(90, 70, "SEATS ALLOCATED")
	pdf.SetFont("Helvetica", "B", 11)
	pdf.SetTextColor(45, 55, 72)
	pdf.Text(90, 75, fmt.Sprintf("%d Ticket(s)", data.SeatsCount))

	// 4. Ticket Stub Details (Right Colored Area)
	pdf.SetTextColor(255, 255, 255)
	
	// Rotated vertical stub layout
	pdf.SetFont("Helvetica", "B", 12)
	pdf.Text(171, 30, "ADMIT ONE")

	pdf.SetFont("Helvetica", "", 9)
	pdf.Text(171, 45, "BOOKING ID:")
	pdf.SetFont("Helvetica", "B", 10)
	// Truncate stub ID if too long
	stubID := data.BookingID
	if len(stubID) > 8 {
		stubID = stubID[:8]
	}
	pdf.Text(171, 50, stubID)

	// Mock Barcode Lines on the stub footer
	pdf.SetFillColor(255, 255, 255)
	barcodeY := 70.0
	barcodeX := 170.0
	widths := []float64{1.5, 0.5, 2.0, 1.0, 0.5, 1.5, 2.5, 0.5, 1.0, 2.0, 0.5, 1.5}
	
	for i, w := range widths {
		if i%2 == 0 {
			pdf.Rect(barcodeX, barcodeY, w, 15, "F")
		}
		barcodeX += w
	}

	// 5. Build and Return Byte Buffer
	var buf bytes.Buffer
	err := pdf.Output(&buf)
	if err != nil {
		return nil, fmt.Errorf("failed writing PDF to stream buffer: %w", err)
	}

	return buf.Bytes(), nil
}