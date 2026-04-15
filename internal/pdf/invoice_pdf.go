package pdf

import (
	"fmt"
	"gestion-commerciale/internal/models"
	"gestion-commerciale/pkg/utils"
)

// ─────────────────────────────────────────────────────────────────────────────
// Factures et documents commerciaux
// ─────────────────────────────────────────────────────────────────────────────

func GenerateSaleInvoice(doc *models.Document, company *models.Company) []byte {
	return generateCommercialDocument(doc, company, "FACTURE DE VENTE")
}
func GenerateQuotation(doc *models.Document, company *models.Company) []byte {
	return generateCommercialDocument(doc, company, "DEVIS")
}
func GenerateProforma(doc *models.Document, company *models.Company) []byte {
	return generateCommercialDocument(doc, company, "FACTURE PROFORMA")
}
func GeneratePurchaseInvoice(doc *models.Document, company *models.Company) []byte {
	return generateCommercialDocument(doc, company, "FACTURE D'ACHAT")
}
func GenerateCreditNote(doc *models.Document, company *models.Company) []byte {
	return generateCommercialDocument(doc, company, "AVOIR / NOTE DE CRÉDIT")
}
func GenerateDeliveryNote(doc *models.Document, company *models.Company) []byte {
	return generateDeliveryDoc(doc, company, "BON DE LIVRAISON")
}
func GenerateReceptionNote(doc *models.Document, company *models.Company) []byte {
	return generateDeliveryDoc(doc, company, "BON DE RÉCEPTION")
}
func GenerateClientOrder(doc *models.Document, company *models.Company) []byte {
	return generateCommercialDocument(doc, company, "BON DE COMMANDE CLIENT")
}
func GenerateSupplierOrder(doc *models.Document, company *models.Company) []byte {
	return generateCommercialDocument(doc, company, "BON DE COMMANDE FOURNISSEUR")
}

// ─────────────────────────────────────────────────────────────────────────────
// Document commercial générique (avec totaux TVA)
// ─────────────────────────────────────────────────────────────────────────────

func generateCommercialDocument(doc *models.Document, company *models.Company, title string) []byte {
	g := NewPDFGenerator()
	pageNum := 1
	y := g.AddPage()

	// En-tête société
	y = g.drawCompanyHeader(company)
	y += 3

	// Titre du document
	y = g.drawDocumentTitle(title, doc.DocNumber, doc.Date, y)
	y += 3

	// Infos supplémentaires (côté gauche)
	g.setFont("Regular", 8)
	g.SetColor(ColorSubtext)
	if doc.SupplierInvoiceNumber != "" {
		g.Text(MarginL, y, "N° Facture fournisseur: "+doc.SupplierInvoiceNumber)
		y += 6
	}
	if doc.ValidityDays > 0 {
		g.Text(MarginL, y, fmt.Sprintf("Validité: %d jour(s)", doc.ValidityDays))
		y += 6
	}
	g.SetColor(ColorText)

	// Bloc client/fournisseur (droite)
	partyLabel := "CLIENT"
	partyName := doc.ClientName
	partyNIF := ""
	partyRC := ""
	if doc.SupplierName != "" {
		partyLabel = "FOURNISSEUR"
		partyName = doc.SupplierName
	}
	y = g.drawClientBlock(partyLabel, partyName, "", partyNIF, partyRC, y)
	y += 2

	// Info paiement
	y = g.drawPaymentInfo(doc.PaymentMethod, doc.PaymentTerms, MarginL, y)
	y += 2

	// ── Tableau des lignes ──
	cols := saleInvoiceColumns()
	y = g.drawTableHeader(cols, y)

	for i, line := range doc.Lines {
		if g.needNewPage(y, 10) {
			g.drawFooter(company, pageNum)
			pageNum++
			y = g.AddPage()
			y = g.drawTableHeader(cols, y)
		}
		values := []string{
			fmt.Sprintf("%d", i+1),
			line.Reference,
			truncStr(line.Designation, 30),
			utils.FormatQuantity(line.Quantity),
			line.Unit,
			utils.FormatMoney(line.UnitPriceHT),
			fmt.Sprintf("%.1f%%", line.DiscountPercent),
			utils.FormatMoney(line.AmountHT),
			fmt.Sprintf("%.0f%%", line.TVARate),
			utils.FormatMoney(line.TVAAmount),
			utils.FormatMoney(line.AmountTTC),
		}
		y = g.drawTableRow(cols, values, y, i%2 == 0)
	}
	y += 5

	// Vérifier l'espace pour les totaux
	if g.needNewPage(y, 80) {
		g.drawFooter(company, pageNum)
		pageNum++
		y = g.AddPage()
	}

	// Totaux
	y = g.drawTotalsBlock(doc, y)
	y += 3

	// Montant en lettres
	y = g.drawAmountInWords(doc.NetAmount, y)
	y += 3

	// Notes
	if doc.Notes != "" {
		g.setFont("Regular", 8)
		g.SetColor(ColorSubtext)
		g.Text(MarginL, y, "Observations: "+truncStr(doc.Notes, 80))
		y += 7
	}

	// Signature
	g.drawSignatureStamp(company, y)

	// Pied de page
	g.drawFooter(company, pageNum)

	return g.GetBytes()
}

// ─────────────────────────────────────────────────────────────────────────────
// Bon de livraison / réception (sans prix)
// ─────────────────────────────────────────────────────────────────────────────

func generateDeliveryDoc(doc *models.Document, company *models.Company, title string) []byte {
	g := NewPDFGenerator()
	pageNum := 1
	y := g.AddPage()

	y = g.drawCompanyHeader(company)
	y += 3
	y = g.drawDocumentTitle(title, doc.DocNumber, doc.Date, y)
	y += 3

	// Infos livraison
	g.setFont("Regular", 8)
	g.SetColor(ColorSubtext)
	if doc.DeliveryAddress != "" {
		g.Text(MarginL, y, "Adresse: "+truncStr(doc.DeliveryAddress, 50))
		y += 6
	}
	if doc.DriverName != "" {
		g.Text(MarginL, y, "Livreur: "+doc.DriverName)
		y += 6
	}
	g.SetColor(ColorText)

	partyName := doc.ClientName
	if doc.SupplierName != "" {
		partyName = doc.SupplierName
	}
	y = g.drawClientBlock("DESTINATAIRE", partyName, "", "", "", y)
	y += 3

	cols := deliveryColumns()
	y = g.drawTableHeader(cols, y)

	for i, line := range doc.Lines {
		if g.needNewPage(y, 8) {
			g.drawFooter(company, pageNum)
			pageNum++
			y = g.AddPage()
			y = g.drawTableHeader(cols, y)
		}
		values := []string{
			fmt.Sprintf("%d", i+1),
			line.Reference,
			truncStr(line.Designation, 44),
			utils.FormatQuantity(line.Quantity),
			line.Unit,
			"",
		}
		y = g.drawTableRow(cols, values, y, i%2 == 0)
	}

	y += 10

	// Zone signature destinataire
	g.setFont("Regular", 8)
	g.SetColor(ColorSubtext)
	g.Text(MarginL, y, "Signature du destinataire (bon pour réception):")
	g.OutlineRect(MarginL, y+5, ContentW*0.42, 22, ColorLight)

	g.drawSignatureStamp(company, y)
	g.drawFooter(company, pageNum)
	return g.GetBytes()
}

// ─────────────────────────────────────────────────────────────────────────────
// Colonnes des tableaux
// ─────────────────────────────────────────────────────────────────────────────

func saleInvoiceColumns() []TableColumn {
	return []TableColumn{
		{Header: "N°", Width: 8, Align: "C"},
		{Header: "Réf.", Width: 16, Align: "L"},
		{Header: "Désignation", Width: 50, Align: "L"},
		{Header: "Qté", Width: 12, Align: "R"},
		{Header: "U", Width: 9, Align: "C"},
		{Header: "P.U. HT", Width: 19, Align: "R"},
		{Header: "Rem%", Width: 11, Align: "R"},
		{Header: "Mnt HT", Width: 19, Align: "R"},
		{Header: "TVA", Width: 9, Align: "C"},
		{Header: "TVA Mnt", Width: 17, Align: "R"},
		{Header: "Mnt TTC", Width: 20, Align: "R"},
	}
}

func deliveryColumns() []TableColumn {
	return []TableColumn{
		{Header: "N°", Width: 10, Align: "C"},
		{Header: "Réf.", Width: 20, Align: "L"},
		{Header: "Désignation", Width: 85, Align: "L"},
		{Header: "Quantité", Width: 25, Align: "R"},
		{Header: "Unité", Width: 20, Align: "C"},
		{Header: "Observation", Width: 30, Align: "L"},
	}
}
