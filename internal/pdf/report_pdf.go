package pdf

import (
	"fmt"
	"gestion-commerciale/internal/models"
	"gestion-commerciale/internal/services"
	"gestion-commerciale/pkg/utils"
	"time"
)

// ─────────────────────────────────────────────────────────────────────────────
// Rapport de ventes
// ─────────────────────────────────────────────────────────────────────────────

func GenerateSalesReport(report *services.SalesReport, company *models.Company, from, to string) []byte {
	g := NewPDFGenerator()
	pageNum := 1
	y := g.AddPage()

	y = g.drawCompanyHeader(company)
	y += 5

	// Titre
	g.FilledRect(MarginL-3, y, ContentW+6, 10, ColorPrimary)
	g.setFont("Bold", 11)
	g.SetColor(ColorWhite)
	g.TextCenter(MarginL, y+1, ContentW, "RAPPORT DES VENTES")
	y += 12

	// Période + résumé
	g.setFont("Regular", 8)
	g.SetColor(ColorSubtext)
	g.Text(MarginL, y, fmt.Sprintf("Période: %s au %s  —  %d facture(s)",
		utils.FormatDateFr(from), utils.FormatDateFr(to), report.InvoiceCount))
	y += 5

	g.FilledRect(MarginL-3, y, ContentW+6, 9, ColorLight)
	g.setFont("Bold", 8)
	g.SetColor(ColorText)
	g.Text(MarginL, y+1.5,
		fmt.Sprintf("Total HT: %s  |  TVA: %s  |  TTC: %s  |  Timbre: %s",
			utils.FormatMoney(report.TotalHT),
			utils.FormatMoney(report.TotalTVA),
			utils.FormatMoney(report.TotalTTC),
			utils.FormatMoney(report.TotalTimbre)))
	y += 12

	cols := []TableColumn{
		{Header: "N° Facture", Width: 28, Align: "L"},
		{Header: "Date", Width: 18, Align: "C"},
		{Header: "Client", Width: 42, Align: "L"},
		{Header: "Total HT", Width: 24, Align: "R"},
		{Header: "TVA", Width: 20, Align: "R"},
		{Header: "Total TTC", Width: 24, Align: "R"},
		{Header: "Payé", Width: 18, Align: "R"},
		{Header: "Statut", Width: 16, Align: "C"},
	}

	y = g.drawTableHeader(cols, y)

	for i, row := range report.Rows {
		if g.needNewPage(y, 8) {
			g.drawFooter(company, pageNum)
			pageNum++
			y = g.AddPage()
			y = g.drawTableHeader(cols, y)
		}
		values := []string{
			row.DocNumber,
			utils.FormatDateFr(row.Date),
			truncStr(row.ClientName, 20),
			utils.FormatMoney(row.TotalHT),
			utils.FormatMoney(row.TotalTVA),
			utils.FormatMoney(row.TotalTTC),
			utils.FormatMoney(row.AmountPaid),
			utils.StatusLabel(row.Status),
		}
		y = g.drawTableRow(cols, values, y, i%2 == 0)
	}

	g.drawFooter(company, pageNum)
	return g.GetBytes()
}

// ─────────────────────────────────────────────────────────────────────────────
// Rapport de stock
// ─────────────────────────────────────────────────────────────────────────────

func GenerateStockReport(report *services.StockReport, company *models.Company) []byte {
	g := NewPDFGenerator()
	pageNum := 1
	y := g.AddPage()

	y = g.drawCompanyHeader(company)
	y += 5

	g.FilledRect(MarginL-3, y, ContentW+6, 10, ColorPrimary)
	g.setFont("Bold", 11)
	g.SetColor(ColorWhite)
	g.TextCenter(MarginL, y+1, ContentW, "ÉTAT DES STOCKS")
	y += 12

	g.setFont("Regular", 8)
	g.SetColor(ColorSubtext)
	g.Text(MarginL, y, fmt.Sprintf("Date: %s  |  Ruptures: %d  |  Stock bas: %d  |  Valeur: %s",
		time.Now().Format("02/01/2006"),
		report.OutOfStockCount, report.LowStockCount,
		utils.FormatMoney(report.TotalValue)))
	y += 8

	cols := []TableColumn{
		{Header: "Réf.", Width: 18, Align: "L"},
		{Header: "Désignation", Width: 54, Align: "L"},
		{Header: "Famille", Width: 26, Align: "L"},
		{Header: "U", Width: 10, Align: "C"},
		{Header: "Stock", Width: 14, Align: "R"},
		{Header: "Min", Width: 12, Align: "R"},
		{Header: "CMUP", Width: 18, Align: "R"},
		{Header: "P.V. TTC", Width: 18, Align: "R"},
		{Header: "Valeur", Width: 20, Align: "R"},
		{Header: "État", Width: 14, Align: "C"},
	}

	y = g.drawTableHeader(cols, y)

	for i, row := range report.Rows {
		if g.needNewPage(y, 7) {
			g.drawFooter(company, pageNum)
			pageNum++
			y = g.AddPage()
			y = g.drawTableHeader(cols, y)
		}
		values := []string{
			row.Reference, truncStr(row.Name, 26), truncStr(row.Category, 13),
			row.Unit,
			utils.FormatQuantity(row.StockQty), utils.FormatQuantity(row.StockMin),
			utils.FormatMoney(row.CMUP), utils.FormatMoney(row.SalePriceTTC),
			utils.FormatMoney(row.StockValue), row.Status,
		}
		y = g.drawTableRow(cols, values, y, i%2 == 0)
	}

	y += 5
	if !g.needNewPage(y, 10) {
		g.setFont("Bold", 9)
		g.SetColor(ColorPrimary)
		g.Text(MarginL, y, "VALEUR TOTALE DU STOCK: "+utils.FormatMoney(report.TotalValue))
	}

	g.drawFooter(company, pageNum)
	return g.GetBytes()
}

// ─────────────────────────────────────────────────────────────────────────────
// Relevé de compte client
// ─────────────────────────────────────────────────────────────────────────────

func GenerateClientStatement(client *models.Client, lines []models.AccountStatement, company *models.Company) []byte {
	g := NewPDFGenerator()
	pageNum := 1
	y := g.AddPage()

	y = g.drawCompanyHeader(company)
	y += 5

	g.FilledRect(MarginL-3, y, ContentW+6, 10, ColorPrimary)
	g.setFont("Bold", 11)
	g.SetColor(ColorWhite)
	g.TextCenter(MarginL, y+1, ContentW, "RELEVÉ DE COMPTE CLIENT")
	y += 12

	g.setFont("Bold", 9)
	g.SetColor(ColorText)
	g.Text(MarginL, y, client.NameFr)
	g.setFont("Regular", 8)
	g.SetColor(ColorSubtext)
	g.Text(MarginL, y+6, fmt.Sprintf("Code: %s  |  Solde: %s", client.Code, utils.FormatMoney(client.Balance)))
	y += 14

	cols := []TableColumn{
		{Header: "Date", Width: 20, Align: "C"},
		{Header: "Type", Width: 14, Align: "C"},
		{Header: "N° Document", Width: 28, Align: "L"},
		{Header: "Description", Width: 54, Align: "L"},
		{Header: "Débit", Width: 22, Align: "R"},
		{Header: "Crédit", Width: 22, Align: "R"},
		{Header: "Solde", Width: 22, Align: "R"},
	}

	y = g.drawTableHeader(cols, y)

	for i, line := range lines {
		if g.needNewPage(y, 7) {
			g.drawFooter(company, pageNum)
			pageNum++
			y = g.AddPage()
			y = g.drawTableHeader(cols, y)
		}
		debitStr, creditStr := "", ""
		if line.Debit > 0 {
			debitStr = utils.FormatMoney(line.Debit)
		}
		if line.Credit > 0 {
			creditStr = utils.FormatMoney(line.Credit)
		}
		values := []string{
			utils.FormatDateFr(line.Date), line.DocType, line.DocNumber,
			truncStr(line.Description, 26), debitStr, creditStr,
			utils.FormatMoney(line.Balance),
		}
		y = g.drawTableRow(cols, values, y, i%2 == 0)
	}

	y += 5
	if !g.needNewPage(y, 8) {
		g.setFont("Bold", 9)
		g.SetColor(ColorPrimary)
		g.Text(MarginL, y, "SOLDE FINAL: "+utils.FormatMoney(client.Balance))
	}

	g.drawFooter(company, pageNum)
	return g.GetBytes()
}

// ─────────────────────────────────────────────────────────────────────────────
// Journal de caisse
// ─────────────────────────────────────────────────────────────────────────────

func GenerateCashJournal(date string, movements []models.CashMovement, company *models.Company) []byte {
	g := NewPDFGenerator()
	pageNum := 1
	y := g.AddPage()

	y = g.drawCompanyHeader(company)
	y += 5

	g.FilledRect(MarginL-3, y, ContentW+6, 10, ColorPrimary)
	g.setFont("Bold", 11)
	g.SetColor(ColorWhite)
	g.TextCenter(MarginL, y+1, ContentW, "JOURNAL DE CAISSE — "+utils.FormatDateFr(date))
	y += 12

	var totalIn, totalOut float64
	for _, m := range movements {
		if m.Type == "in" {
			totalIn += m.Amount
		} else {
			totalOut += m.Amount
		}
	}

	g.setFont("Regular", 8)
	g.SetColor(ColorSubtext)
	g.Text(MarginL, y, fmt.Sprintf("Entrées: %s  |  Sorties: %s  |  Solde: %s",
		utils.FormatMoney(totalIn), utils.FormatMoney(totalOut),
		utils.FormatMoney(totalIn-totalOut)))
	y += 8

	cols := []TableColumn{
		{Header: "Heure", Width: 14, Align: "C"},
		{Header: "Sens", Width: 14, Align: "C"},
		{Header: "Catégorie", Width: 26, Align: "L"},
		{Header: "Description", Width: 48, Align: "L"},
		{Header: "Référence", Width: 20, Align: "L"},
		{Header: "Partie", Width: 24, Align: "L"},
		{Header: "Entrée", Width: 18, Align: "R"},
		{Header: "Sortie", Width: 18, Align: "R"},
		{Header: "Solde", Width: 18, Align: "R"},
	}

	y = g.drawTableHeader(cols, y)

	for i, m := range movements {
		if g.needNewPage(y, 7) {
			g.drawFooter(company, pageNum)
			pageNum++
			y = g.AddPage()
			y = g.drawTableHeader(cols, y)
		}
		entree, sortie := "", ""
		if m.Type == "in" {
			entree = utils.FormatMoney(m.Amount)
		} else {
			sortie = utils.FormatMoney(m.Amount)
		}
		heure := ""
		if len(m.Date) >= 16 {
			heure = m.Date[11:16]
		}
		typeStr := "↑ Entrée"
		if m.Type == "out" {
			typeStr = "↓ Sortie"
		}
		values := []string{
			heure, typeStr,
			truncStr(m.Category, 13), truncStr(m.Description, 24),
			m.Reference, truncStr(m.PartyName, 12),
			entree, sortie, utils.FormatMoney(m.Balance),
		}
		y = g.drawTableRow(cols, values, y, i%2 == 0)
	}

	y += 5
	if !g.needNewPage(y, 12) {
		g.FilledRect(MarginL-3, y, ContentW+6, 11, ColorLight)
		g.setFont("Bold", 9)
		g.SetColor(ColorText)
		g.Text(MarginL, y+2, fmt.Sprintf("ENTRÉES: %s  |  SORTIES: %s  |  SOLDE: %s",
			utils.FormatMoney(totalIn), utils.FormatMoney(totalOut),
			utils.FormatMoney(totalIn-totalOut)))
	}

	g.drawFooter(company, pageNum)
	return g.GetBytes()
}

// ─────────────────────────────────────────────────────────────────────────────
// Formulaire G50
// ─────────────────────────────────────────────────────────────────────────────

func GenerateG50(g50 *models.G50Data, company *models.Company) []byte {
	gen := NewPDFGenerator()
	y := gen.AddPage()

	y = gen.drawCompanyHeader(company)
	y += 5

	gen.FilledRect(MarginL-3, y, ContentW+6, 12, ColorPrimary)
	gen.setFont("Bold", 12)
	gen.SetColor(ColorWhite)
	gen.TextCenter(MarginL, y+2, ContentW/2,
		fmt.Sprintf("DÉCLARATION G50 — %02d/%d", g50.Month, g50.Year))
	gen.TextRight(MarginL+ContentW/2, y+2, ContentW/2, "NIF: "+company.NIF)
	y += 15

	y = gen.g50Section("I. CHIFFRE D'AFFAIRES", y)
	rows := [][]string{
		{"CA soumis à TVA 19%", utils.FormatMoney(g50.Revenue19)},
		{"CA soumis à TVA 9%", utils.FormatMoney(g50.Revenue9)},
		{"CA exonéré / hors TVA", utils.FormatMoney(g50.RevenueExempt)},
		{"Total CA", utils.FormatMoney(g50.TotalRevenue)},
	}
	for i, r := range rows {
		y = gen.g50Row(r[0], r[1], y, i == len(rows)-1)
	}

	y += 5
	y = gen.g50Section("II. TVA", y)
	tvaRows := [][]string{
		{"TVA collectée 19%", utils.FormatMoney(g50.TVACollected19)},
		{"TVA collectée 9%", utils.FormatMoney(g50.TVACollected9)},
		{"Total TVA collectée", utils.FormatMoney(g50.TotalTVACollected)},
		{"(-) TVA déductible sur achats", utils.FormatMoney(g50.TVADeductiblePurchases)},
		{"(-) TVA précompte précédente", utils.FormatMoney(g50.TVAPrecompte)},
		{"TVA nette à payer", utils.FormatMoney(g50.TVADue)},
		{"Crédit TVA à reporter", utils.FormatMoney(g50.TVAPrecompteNext)},
	}
	for i, r := range tvaRows {
		y = gen.g50Row(r[0], r[1], y, i >= 2 && i == len(tvaRows)-2)
	}

	y += 5
	y = gen.g50Section("III. AUTRES TAXES", y)
	otherRows := [][]string{
		{"TAP (Taxe Activité Professionnelle)", utils.FormatMoney(g50.TAP)},
		{"Droit de timbre", utils.FormatMoney(g50.Timbre)},
		{"IRG/IBS sur salaires", utils.FormatMoney(g50.IRGSalaries)},
	}
	for _, r := range otherRows {
		y = gen.g50Row(r[0], r[1], y, false)
	}

	y += 5
	gen.FilledRect(MarginL-3, y, ContentW+6, 12, ColorPrimary)
	gen.setFont("Bold", 11)
	gen.SetColor(ColorWhite)
	gen.Text(MarginL+2, y+2, "TOTAL GÉNÉRAL À PAYER:")
	gen.TextRight(MarginL, y+2, ContentW-2, utils.FormatMoney(g50.TotalDue))
	y += 15

	gen.setFont("Regular", 8)
	gen.SetColor(ColorSubtext)
	gen.Text(MarginL, y, fmt.Sprintf("Fait à ____________  le %s", time.Now().Format("02/01/2006")))
	gen.Text(MarginL+ContentW*0.55, y, "Signature et cachet:")

	gen.drawFooter(company, 1)
	return gen.GetBytes()
}

func (g *PDFGenerator) g50Section(title string, y float64) float64 {
	g.FilledRect(MarginL-3, y, ContentW+6, 8, ColorSecondary)
	g.setFont("Bold", 9)
	g.SetColor(ColorWhite)
	g.Text(MarginL+1, y+1, title)
	g.SetColor(ColorText)
	return y + 9
}

func (g *PDFGenerator) g50Row(label, value string, y float64, bold bool) float64 {
	if bold {
		g.setFont("Bold", 9)
		g.FilledRect(MarginL-3, y, ContentW+6, 7, ColorLight)
	} else {
		g.setFont("Regular", 8)
	}
	g.SetColor(ColorSubtext)
	g.Text(MarginL+2, y+1, label)
	g.SetColor(ColorText)
	g.TextRight(MarginL, y+1, ContentW-2, value)
	g.pdf.SetLineWidth(0.1)
	g.HLine(MarginL-3, y+7, ContentW+6, ColorLight)
	return y + 7
}

// ─────────────────────────────────────────────────────────────────────────────
// Fiche d'inventaire
// ─────────────────────────────────────────────────────────────────────────────

func GenerateInventorySheet(inventory *models.Inventory, company *models.Company) []byte {
	g := NewPDFGenerator()
	pageNum := 1
	y := g.AddPage()

	y = g.drawCompanyHeader(company)
	y += 5

	title := "FEUILLE D'INVENTAIRE"
	if inventory.Status == "confirmed" {
		title = "RÉSULTATS D'INVENTAIRE"
	}

	g.FilledRect(MarginL-3, y, ContentW+6, 10, ColorPrimary)
	g.setFont("Bold", 11)
	g.SetColor(ColorWhite)
	g.TextCenter(MarginL, y+1, ContentW, title)
	y += 12

	g.setFont("Regular", 8)
	g.SetColor(ColorSubtext)
	g.Text(MarginL, y, fmt.Sprintf("Date: %s  |  Type: %s  |  Statut: %s",
		utils.FormatDateFr(inventory.Date), inventory.Type, inventory.Status))
	y += 8

	showDiff := inventory.Status == "confirmed"
	cols := []TableColumn{
		{Header: "N°", Width: 10, Align: "C"},
		{Header: "Réf.", Width: 18, Align: "L"},
		{Header: "Code-barres", Width: 28, Align: "L"},
		{Header: "Désignation", Width: 58, Align: "L"},
		{Header: "Qté théorique", Width: 22, Align: "R"},
		{Header: "Qté physique", Width: 22, Align: "R"},
	}
	if showDiff {
		cols = append(cols,
			TableColumn{Header: "Écart", Width: 18, Align: "R"},
			TableColumn{Header: "Valeur", Width: 14, Align: "R"},
		)
	}

	y = g.drawTableHeader(cols, y)

	for i, line := range inventory.Lines {
		if g.needNewPage(y, 7) {
			g.drawFooter(company, pageNum)
			pageNum++
			y = g.AddPage()
			y = g.drawTableHeader(cols, y)
		}
		physQty := ""
		if showDiff {
			physQty = utils.FormatQuantity(line.PhysicalQty)
		}
		values := []string{
			fmt.Sprintf("%d", i+1),
			line.Reference, line.Barcode,
			truncStr(line.ArticleName, 28),
			utils.FormatQuantity(line.TheoreticalQty),
			physQty,
		}
		if showDiff {
			values = append(values,
				fmt.Sprintf("%+.2f", line.Difference),
				utils.FormatMoney(line.Value))
		}
		y = g.drawTableRow(cols, values, y, i%2 == 0)
	}

	g.drawFooter(company, pageNum)
	return g.GetBytes()
}
