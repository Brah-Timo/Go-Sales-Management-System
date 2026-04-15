package pdf

import (
	"fmt"
	"gestion-commerciale/internal/models"
	"gestion-commerciale/pkg/utils"
	"strings"

	"github.com/signintech/gopdf"
)

// ─────────────────────────────────────────────────────────────────────────────
// Ticket de caisse (format thermique 80mm)
// ─────────────────────────────────────────────────────────────────────────────

const (
	TicketWidth   = 80.0
	TicketMargin  = 3.0
	TicketContent = TicketWidth - TicketMargin*2
)

// GenerateTicket génère un ticket de caisse format 80mm
func GenerateTicket(doc *models.Document, company *models.Company) []byte {
	lineCount := len(doc.Lines)
	estimatedH := 100.0 + float64(lineCount)*10 + 60

	g := &PDFGenerator{pdf: &gopdf.GoPdf{}, fontDir: "assets/fonts"}
	g.pdf.Start(gopdf.Config{PageSize: gopdf.Rect{W: TicketWidth, H: estimatedH}})
	g.loadFonts()
	g.pdf.AddPage()

	y := TicketMargin

	// En-tête
	g.setFont("Bold", 10)
	g.SetColor(ColorText)
	g.TextCenter(TicketMargin, y, TicketContent, company.NameFr)
	y += 8

	g.setFont("Regular", 7)
	g.SetColor(ColorSubtext)
	if company.Address != "" {
		g.TextCenter(TicketMargin, y, TicketContent, company.Address)
		y += 5
	}
	if company.Phone != "" {
		g.TextCenter(TicketMargin, y, TicketContent, "Tél: "+company.Phone)
		y += 5
	}
	if company.NIF != "" {
		g.TextCenter(TicketMargin, y, TicketContent, "NIF: "+company.NIF)
		y += 5
	}

	g.SetColor(ColorText)
	y = ticketLine(g, y, "=")
	y += 1

	// Titre
	g.setFont("Bold", 8)
	g.TextCenter(TicketMargin, y, TicketContent, "TICKET DE CAISSE")
	y += 6

	g.setFont("Regular", 7)
	g.Text(TicketMargin, y, "N°: "+doc.DocNumber)
	g.TextRight(TicketMargin, y, TicketContent, utils.FormatDateFr(doc.Date))
	y += 5

	if doc.ClientName != "" {
		g.Text(TicketMargin, y, "Client: "+truncStr(doc.ClientName, 20))
		y += 5
	}

	y = ticketLine(g, y, "-")

	// En-têtes colonnes
	g.setFont("Bold", 7)
	g.Text(TicketMargin, y, "DÉSIGNATION")
	g.TextRight(TicketMargin, y, TicketContent, "MONTANT")
	y += 5
	y = ticketLine(g, y, "-")

	// Lignes produits
	g.setFont("Regular", 7)
	for _, line := range doc.Lines {
		name := truncStr(line.Designation, 22)
		g.SetColor(ColorText)
		g.Text(TicketMargin, y, name)
		g.TextRight(TicketMargin, y, TicketContent, utils.FormatMoney(line.AmountTTC))
		y += 5

		if line.Quantity != 1 {
			detail := fmt.Sprintf("  %.0f x %s", line.Quantity, utils.FormatMoney(line.UnitPriceHT))
			if line.DiscountPercent > 0 {
				detail += fmt.Sprintf(" (-%0.f%%)", line.DiscountPercent)
			}
			g.SetColor(ColorSubtext)
			g.Text(TicketMargin, y, detail)
			y += 4
		}
	}

	g.SetColor(ColorText)
	y = ticketLine(g, y, "=")
	y += 1

	// Totaux
	g.setFont("Regular", 7)
	if doc.TotalDiscount > 0 {
		ticketRow(g, TicketMargin, TicketContent, y, "Remises:", "-"+utils.FormatMoney(doc.TotalDiscount))
		y += 5
	}
	if doc.TotalTVA > 0 {
		ticketRow(g, TicketMargin, TicketContent, y, "TVA:", utils.FormatMoney(doc.TotalTVA))
		y += 5
	}
	if doc.Timbre > 0 {
		ticketRow(g, TicketMargin, TicketContent, y, "Timbre:", utils.FormatMoney(doc.Timbre))
		y += 5
	}

	y = ticketLine(g, y, "=")
	y += 1

	// Total net
	g.setFont("Bold", 10)
	g.SetColor(ColorText)
	ticketRow(g, TicketMargin, TicketContent, y, "TOTAL:", utils.FormatMoney(doc.NetAmount))
	y += 8

	// Mode paiement
	g.setFont("Regular", 7)
	g.SetColor(ColorSubtext)
	g.TextCenter(TicketMargin, y, TicketContent, utils.PaymentMethodLabel(doc.PaymentMethod))
	y += 5

	y = ticketLine(g, y, "-")
	y += 2

	// Message
	g.setFont("Italic", 7)
	thanks := "Merci de votre visite!"
	if company.FooterText != "" {
		thanks = company.FooterText
	}
	g.TextCenter(TicketMargin, y, TicketContent, thanks)
	y += 5
	g.TextCenter(TicketMargin, y, TicketContent, "À bientôt!")

	g.SetColor(ColorText)
	return g.GetBytes()
}

// ─────────────────────────────────────────────────────────────────────────────
// Reçu de paiement
// ─────────────────────────────────────────────────────────────────────────────

// GenerateReceipt génère un reçu de paiement (A6)
func GenerateReceipt(payment *models.Payment, company *models.Company) []byte {
	g := &PDFGenerator{pdf: &gopdf.GoPdf{}, fontDir: "assets/fonts"}
	g.pdf.Start(gopdf.Config{PageSize: gopdf.Rect{W: 148, H: 105}})
	g.loadFonts()
	g.pdf.AddPage()

	w := 148.0
	m := 6.0
	cw := w - m*2
	y := m

	// Titre
	g.FilledRect(m-2, y, cw+4, 10, ColorPrimary)
	g.setFont("Bold", 10)
	g.SetColor(ColorWhite)
	g.TextCenter(m, y+1, cw, "REÇU DE PAIEMENT")
	y += 13

	// Société
	g.setFont("Bold", 9)
	g.SetColor(ColorText)
	g.TextCenter(m, y, cw, company.NameFr)
	y += 6

	g.setFont("Regular", 7)
	g.SetColor(ColorSubtext)
	if company.Address != "" {
		g.TextCenter(m, y, cw, company.Address)
		y += 5
	}
	y += 2

	// Infos paiement
	rows := [][]string{
		{"Reçu le:", utils.FormatDateFr(payment.Date)},
		{"De:", truncStr(payment.ClientName, 30)},
		{"Montant:", utils.FormatMoney(payment.Amount)},
		{"Mode:", utils.PaymentMethodLabel(payment.PaymentMethod)},
	}
	if payment.ChequeNumber != "" {
		rows = append(rows, []string{"N° Chèque:", payment.ChequeNumber})
	}
	if payment.BankName != "" {
		rows = append(rows, []string{"Banque:", payment.BankName})
	}

	for _, r := range rows {
		g.setFont("Bold", 8)
		g.SetColor(ColorSubtext)
		g.Text(m, y, r[0])
		g.setFont("Regular", 8)
		g.SetColor(ColorText)
		g.Text(m+28, y, r[1])
		y += 6
	}

	// Montant en lettres
	y += 2
	g.setFont("Italic", 7)
	g.SetColor(ColorSubtext)
	g.Text(m, y, truncStr("En lettres: "+utils.NumberToWordsFr(payment.Amount), 55))
	y += 8

	// Signatures
	g.setFont("Regular", 7)
	g.SetColor(ColorSubtext)
	g.Text(m, y, "Le caissier:")
	g.TextRight(m, y, cw, "La partie versante:")
	y += 3
	g.pdf.SetLineWidth(0.3)
	g.HLine(m, y+10, cw*0.4, ColorLight)
	g.HLine(m+cw*0.6, y+10, cw*0.4, ColorLight)

	return g.GetBytes()
}

// ─────────────────────────────────────────────────────────────────────────────
// Utilitaires ticket
// ─────────────────────────────────────────────────────────────────────────────

func ticketLine(g *PDFGenerator, y float64, char string) float64 {
	g.setFont("Regular", 7)
	g.SetColor(ColorLight)
	line := strings.Repeat(char, 28)
	g.Text(TicketMargin, y, line)
	g.SetColor(ColorText)
	return y + 4
}

func ticketRow(g *PDFGenerator, x, w, y float64, label, value string) {
	g.Text(x, y, label)
	g.TextRight(x, y, w, value)
}
