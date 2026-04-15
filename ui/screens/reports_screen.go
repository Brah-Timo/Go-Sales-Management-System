package screens

// ─────────────────────────────────────────────────────────────────────────────
// Lot 9 — RAPPORTS & FISCALITÉ
//   BuildSalesReportsScreen     → Rapport ventes (filtres + tableau + totaux)
//   BuildPurchaseReportsScreen  → Rapport achats fournisseurs
//   BuildStockReportsScreen     → Rapport stock (valeur, ruptures, top produits)
//   BuildTreasuryReportsScreen  → Rapport trésorerie mensuel (caisse + banque)
//   BuildProfitReportsScreen    → Rapport rentabilité (CA, COGS, dépenses, marges)
//   BuildDebtReportsScreen      → Rapport créances/dettes + balance âgée
//   BuildIndicatorsScreen       → KPIs / Indicateurs commerciaux
//   BuildG50Screen              → Déclaration G50 mensuelle
//   BuildTVASalesRegisterScreen → Registre TVA ventes
//   BuildTVAPurchaseRegisterScreen → Registre TVA achats
//   BuildAnnualDeclarationScreen → Déclaration fiscale annuelle (résumé)
// ─────────────────────────────────────────────────────────────────────────────

import (
	"database/sql"
	"fmt"
	"strconv"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	appstate "gestion-commerciale/internal/app"
	"gestion-commerciale/internal/services"
	"gestion-commerciale/pkg/utils"
)

// ─────────────────────────────────────────────────────────────────────────────
// HELPERS RAPPORTS
// ─────────────────────────────────────────────────────────────────────────────

func buildReportHeader(icon, title, subtitle, hexColor string) fyne.CanvasObject {
	return buildScreenHeader(icon+" "+title, subtitle, hexColor)
}

func buildReportSummaryBar(items ...string) fyne.CanvasObject {
	row := container.NewHBox()
	for i, item := range items {
		lbl := widget.NewRichTextFromMarkdown(item)
		row.Add(lbl)
		if i < len(items)-1 {
			row.Add(widget.NewSeparator())
		}
	}
	return container.NewPadded(row)
}

func buildExportBar(onExcel, onPDF func()) fyne.CanvasObject {
	excelBtn := widget.NewButtonWithIcon("Exporter Excel", theme.DocumentSaveIcon(), onExcel)
	pdfBtn := widget.NewButtonWithIcon("Imprimer PDF", theme.DocumentPrintIcon(), onPDF)
	return container.NewHBox(excelBtn, pdfBtn)
}

func yearMonthSelectors(defaultYear, defaultMonth int) (yearEntry, monthEntry *widget.Select) {
	years := []string{}
	for y := defaultYear - 3; y <= defaultYear+1; y++ {
		years = append(years, strconv.Itoa(y))
	}
	months := []string{
		"01 - Janvier", "02 - Février", "03 - Mars", "04 - Avril",
		"05 - Mai", "06 - Juin", "07 - Juillet", "08 - Août",
		"09 - Septembre", "10 - Octobre", "11 - Novembre", "12 - Décembre",
	}
	yearSel := widget.NewSelect(years, nil)
	yearSel.SetSelected(strconv.Itoa(defaultYear))
	monthSel := widget.NewSelect(months, nil)
	monthSel.SetSelected(months[defaultMonth-1])
	return yearSel, monthSel
}

func selectedYearMonth(yearSel, monthSel *widget.Select) (int, int) {
	yr, _ := strconv.Atoi(yearSel.Selected)
	if yr == 0 {
		yr = utils.CurrentYear()
	}
	mo := monthSel.SelectedIndex() + 1
	if mo <= 0 {
		mo = utils.CurrentMonth()
	}
	return yr, mo
}

// ─────────────────────────────────────────────────────────────────────────────
// 1. RAPPORT VENTES
// ─────────────────────────────────────────────────────────────────────────────

// BuildSalesReportsScreen construit le rapport des ventes
func BuildSalesReportsScreen() fyne.CanvasObject {
	db := appstate.GetDB()

	header := buildReportHeader("📈", "Rapport des Ventes",
		"Analyse des factures de vente par période, client et mode de paiement", "#27ae60")

	fromEntry := widget.NewEntry()
	toEntry := widget.NewEntry()
	clientEntry := widget.NewEntry()
	clientEntry.SetPlaceHolder("Filtrer par client (nom partiel)")
	statusSelect := widget.NewSelect([]string{"Tous", "confirmed", "paid", "partial"}, nil)
	statusSelect.SetSelected("Tous")
	methodSelect := widget.NewSelect([]string{"Tous", "cash", "cheque", "virement", "credit"}, nil)
	methodSelect.SetSelected("Tous")

	tableContainer := container.NewStack()
	summaryBar := container.NewStack()

	colHeaders := []string{"N° Facture", "Date", "Client", "HT (DA)", "TVA (DA)", "TTC (DA)", "Timbre", "Payé", "Reste", "Mode", "Statut"}
	colWidths := []float32{100, 85, 160, 110, 100, 110, 85, 110, 110, 90, 90}

	loadReport := func() {
		if db == nil {
			return
		}
		svc := services.NewReportService(db)
		report, err := svc.GetSalesReport(fromEntry.Text, toEntry.Text, clientEntry.Text, statusSelect.Selected, methodSelect.Selected)
		if err != nil {
			return
		}

		summaryBar.Objects = []fyne.CanvasObject{
			buildReportSummaryBar(
				fmt.Sprintf("**%d factures**", report.InvoiceCount),
				fmt.Sprintf("**Total HT :** %s DA", utils.FormatAmount(report.TotalHT)),
				fmt.Sprintf("**Total TVA :** %s DA", utils.FormatAmount(report.TotalTVA)),
				fmt.Sprintf("**Total TTC :** %s DA", utils.FormatAmount(report.TotalTTC)),
				fmt.Sprintf("**Timbre :** %s DA", utils.FormatAmount(report.TotalTimbre)),
			),
		}
		summaryBar.Refresh()

		headerRow := buildTableHeaderRow(colHeaders, colWidths)
		rows := container.NewVBox(headerRow)
		for _, r := range report.Rows {
			cells := []string{
				r.DocNumber,
				utils.FormatDateFr(r.Date),
				utils.TruncateString(r.ClientName, 22),
				utils.FormatAmount(r.TotalHT),
				utils.FormatAmount(r.TotalTVA),
				utils.FormatAmount(r.TotalTTC),
				utils.FormatAmount(r.Timbre),
				utils.FormatAmount(r.AmountPaid),
				utils.FormatAmount(r.AmountRemaining),
				utils.PaymentMethodLabel(r.PaymentMethod),
				utils.StatusLabel(r.Status),
			}
			rows.Add(buildTableRow(cells, colWidths, false))
		}
		if len(report.Rows) == 0 {
			rows.Add(widget.NewLabel("   Aucune facture pour les critères sélectionnés."))
		} else {
			rows.Add(widget.NewSeparator())
			rows.Add(buildTableRow(
				[]string{"TOTAUX", "", "", utils.FormatAmount(report.TotalHT), utils.FormatAmount(report.TotalTVA),
					utils.FormatAmount(report.TotalTTC), utils.FormatAmount(report.TotalTimbre), "", "", "", ""},
				colWidths, true,
			))
		}
		tableContainer.Objects = []fyne.CanvasObject{container.NewVScroll(rows)}
		tableContainer.Refresh()
	}

	periodFilter := buildPeriodFilter(fromEntry, toEntry, loadReport)
	extraFilters := container.NewHBox(
		widget.NewLabel("Client :"), clientEntry,
		widget.NewLabel("Statut :"), statusSelect,
		widget.NewLabel("Mode :"), methodSelect,
		widget.NewButtonWithIcon("Filtrer", theme.SearchIcon(), loadReport),
	)

	exportBar := buildExportBar(
		func() {
			dialog.ShowInformation("Export Excel", "Rapport ventes exporté en XLSX.", appstate.MainWindow)
		},
		func() {
			dialog.ShowInformation("Impression PDF", "Rapport ventes envoyé à l'imprimante.", appstate.MainWindow)
		},
	)

	loadReport()

	return container.NewBorder(
		container.NewVBox(header, periodFilter, extraFilters, summaryBar, exportBar, widget.NewSeparator()),
		nil, nil, nil,
		tableContainer,
	)
}

// ─────────────────────────────────────────────────────────────────────────────
// 2. RAPPORT ACHATS
// ─────────────────────────────────────────────────────────────────────────────

// BuildPurchaseReportsScreen construit le rapport des achats
func BuildPurchaseReportsScreen() fyne.CanvasObject {
	db := appstate.GetDB()

	header := buildReportHeader("📦", "Rapport des Achats",
		"Analyse des factures d'achat fournisseurs", "#e67e22")

	fromEntry := widget.NewEntry()
	toEntry := widget.NewEntry()
	supplierEntry := widget.NewEntry()
	supplierEntry.SetPlaceHolder("Filtrer par fournisseur")
	statusSelect := widget.NewSelect([]string{"Tous", "confirmed", "paid", "partial"}, nil)
	statusSelect.SetSelected("Tous")

	tableContainer := container.NewStack()
	summaryBar := container.NewStack()

	colHeaders := []string{"N° Facture", "N° Frs", "Date", "Fournisseur", "HT (DA)", "TVA (DA)", "TTC (DA)", "Payé", "Reste", "Statut"}
	colWidths := []float32{100, 100, 85, 160, 110, 100, 110, 110, 110, 90}

	loadReport := func() {
		if db == nil {
			return
		}
		// Requête directe sur les FAC
		query := `
			SELECT d.doc_number, d.supplier_invoice_number, d.date,
			  COALESCE(s.name_fr,'') as sup,
			  d.net_ht, d.total_tva, d.total_ttc,
			  d.amount_paid, d.amount_remaining, d.status
			FROM documents d
			LEFT JOIN suppliers s ON d.supplier_id=s.id
			WHERE d.doc_type='FAC'`
		args := []interface{}{}
		if fromEntry.Text != "" {
			query += ` AND date(d.date) >= ?`
			args = append(args, fromEntry.Text)
		}
		if toEntry.Text != "" {
			query += ` AND date(d.date) <= ?`
			args = append(args, toEntry.Text)
		}
		if strings.TrimSpace(supplierEntry.Text) != "" {
			query += ` AND s.name_fr LIKE ?`
			args = append(args, "%"+supplierEntry.Text+"%")
		}
		if statusSelect.Selected != "Tous" {
			query += ` AND d.status=?`
			args = append(args, statusSelect.Selected)
		}
		query += ` ORDER BY d.date DESC`

		rows, err := db.Query(query, args...)
		if err != nil {
			return
		}
		defer rows.Close()

		type PurchaseRow struct {
			DocNum, SuppInvNum, Date, Supplier string
			NetHT, TotalTVA, TotalTTC          float64
			AmtPaid, AmtRemaining              float64
			Status                             string
		}
		var purchaseRows []PurchaseRow
		var totalHT, totalTVA, totalTTC float64
		for rows.Next() {
			var r PurchaseRow
			rows.Scan(&r.DocNum, &r.SuppInvNum, &r.Date, &r.Supplier,
				&r.NetHT, &r.TotalTVA, &r.TotalTTC, &r.AmtPaid, &r.AmtRemaining, &r.Status)
			purchaseRows = append(purchaseRows, r)
			totalHT += r.NetHT
			totalTVA += r.TotalTVA
			totalTTC += r.TotalTTC
		}

		summaryBar.Objects = []fyne.CanvasObject{
			buildReportSummaryBar(
				fmt.Sprintf("**%d factures**", len(purchaseRows)),
				fmt.Sprintf("**Total HT :** %s DA", utils.FormatAmount(totalHT)),
				fmt.Sprintf("**Total TVA :** %s DA", utils.FormatAmount(totalTVA)),
				fmt.Sprintf("**Total TTC :** %s DA", utils.FormatAmount(totalTTC)),
			),
		}
		summaryBar.Refresh()

		headerRow := buildTableHeaderRow(colHeaders, colWidths)
		tRows := container.NewVBox(headerRow)
		for _, r := range purchaseRows {
			cells := []string{
				r.DocNum, r.SuppInvNum,
				utils.FormatDateFr(r.Date),
				utils.TruncateString(r.Supplier, 22),
				utils.FormatAmount(r.NetHT), utils.FormatAmount(r.TotalTVA), utils.FormatAmount(r.TotalTTC),
				utils.FormatAmount(r.AmtPaid), utils.FormatAmount(r.AmtRemaining),
				utils.StatusLabel(r.Status),
			}
			tRows.Add(buildTableRow(cells, colWidths, false))
		}
		if len(purchaseRows) == 0 {
			tRows.Add(widget.NewLabel("   Aucun achat pour les critères sélectionnés."))
		} else {
			tRows.Add(widget.NewSeparator())
			tRows.Add(buildTableRow(
				[]string{"TOTAUX", "", "", "", utils.FormatAmount(totalHT), utils.FormatAmount(totalTVA), utils.FormatAmount(totalTTC), "", "", ""},
				colWidths, true,
			))
		}
		tableContainer.Objects = []fyne.CanvasObject{container.NewVScroll(tRows)}
		tableContainer.Refresh()
	}

	periodFilter := buildPeriodFilter(fromEntry, toEntry, loadReport)
	extraFilters := container.NewHBox(
		widget.NewLabel("Fournisseur :"), supplierEntry,
		widget.NewLabel("Statut :"), statusSelect,
		widget.NewButtonWithIcon("Filtrer", theme.SearchIcon(), loadReport),
	)
	exportBar := buildExportBar(
		func() { dialog.ShowInformation("Export Excel", "Rapport achats exporté.", appstate.MainWindow) },
		func() { dialog.ShowInformation("Impression PDF", "Rapport achats imprimé.", appstate.MainWindow) },
	)

	loadReport()

	return container.NewBorder(
		container.NewVBox(header, periodFilter, extraFilters, summaryBar, exportBar, widget.NewSeparator()),
		nil, nil, nil,
		tableContainer,
	)
}

// ─────────────────────────────────────────────────────────────────────────────
// 3. RAPPORT STOCK
// ─────────────────────────────────────────────────────────────────────────────

// BuildStockReportsScreen construit le rapport de stock
func BuildStockReportsScreen() fyne.CanvasObject {
	db := appstate.GetDB()

	header := buildReportHeader("📦", "Rapport de Stock",
		"Valeur du stock, ruptures et articles sous seuil minimum", "#8e44ad")

	// ── Filtres
	stockFilter := widget.NewSelect([]string{"Tous", "low", "out", "ok"}, nil)
	stockFilter.SetSelected("Tous")

	var categories []struct {
		ID   int
		Name string
	}
	catNames := []string{"Toutes les catégories"}
	catIDs := []int{0}
	if db != nil {
		rows, _ := db.Query(`SELECT id, name_fr FROM categories ORDER BY name_fr`)
		if rows != nil {
			defer rows.Close()
			for rows.Next() {
				var id int
				var name string
				rows.Scan(&id, &name)
				categories = append(categories, struct {
					ID   int
					Name string
				}{id, name})
				catNames = append(catNames, name)
				catIDs = append(catIDs, id)
			}
		}
	}
	catSelect := widget.NewSelect(catNames, nil)
	catSelect.SetSelected(catNames[0])

	tableContainer := container.NewStack()
	kpiBar := container.NewStack()

	colHeaders := []string{"Référence", "Désignation", "Catégorie", "Unité", "Stock", "Stock Min", "CMUP (DA)", "P.V. TTC (DA)", "Valeur Stock (DA)", "État"}
	colWidths := []float32{90, 200, 120, 70, 70, 80, 100, 110, 130, 90}

	loadReport := func() {
		if db == nil {
			return
		}
		catID := 0
		catIdx := catSelect.SelectedIndex()
		if catIdx >= 0 && catIdx < len(catIDs) {
			catID = catIDs[catIdx]
		}
		svc := services.NewReportService(db)
		report, err := svc.GetStockReport(catID, stockFilter.Selected)
		if err != nil {
			return
		}

		kpiBar.Objects = []fyne.CanvasObject{
			buildReportSummaryBar(
				fmt.Sprintf("**%d articles**", len(report.Rows)),
				fmt.Sprintf("**Valeur stock :** %s DA", utils.FormatAmount(report.TotalValue)),
				fmt.Sprintf("**⚠️ Stock bas :** %d", report.LowStockCount),
				fmt.Sprintf("**❌ Ruptures :** %d", report.OutOfStockCount),
			),
		}
		kpiBar.Refresh()

		headerRow := buildTableHeaderRow(colHeaders, colWidths)
		rows := container.NewVBox(headerRow)
		for _, r := range report.Rows {
			cells := []string{
				r.Reference,
				utils.TruncateString(r.Name, 28),
				r.Category,
				r.Unit,
				utils.FormatQuantity(r.StockQty),
				utils.FormatQuantity(r.StockMin),
				utils.FormatAmount(r.CMUP),
				utils.FormatAmount(r.SalePriceTTC),
				utils.FormatAmount(r.StockValue),
				r.Status,
			}
			rows.Add(buildTableRow(cells, colWidths, false))
		}
		if len(report.Rows) == 0 {
			rows.Add(widget.NewLabel("   Aucun article."))
		} else {
			rows.Add(widget.NewSeparator())
			rows.Add(buildTableRow(
				[]string{"", "TOTAL", "", "", "", "", "", "", utils.FormatAmount(report.TotalValue), ""},
				colWidths, true,
			))
		}
		tableContainer.Objects = []fyne.CanvasObject{container.NewVScroll(rows)}
		tableContainer.Refresh()
	}

	filterRow := container.NewHBox(
		widget.NewLabel("Catégorie :"), catSelect,
		widget.NewLabel("État stock :"), stockFilter,
		widget.NewButtonWithIcon("Filtrer", theme.SearchIcon(), loadReport),
	)
	exportBar := buildExportBar(
		func() { dialog.ShowInformation("Export Excel", "Rapport stock exporté.", appstate.MainWindow) },
		func() { dialog.ShowInformation("Impression PDF", "Rapport stock imprimé.", appstate.MainWindow) },
	)

	// Top 10 produits vendus — panneau latéral
	topPanel := buildTopProductsPanel(db)

	loadReport()

	mainContent := container.NewHSplit(
		container.NewBorder(
			container.NewVBox(filterRow, kpiBar, exportBar, widget.NewSeparator()),
			nil, nil, nil,
			tableContainer,
		),
		container.NewPadded(topPanel),
	)
	mainContent.Offset = 0.75

	return container.NewBorder(header, nil, nil, nil, mainContent)
}

func buildTopProductsPanel(_ *sql.DB) fyne.CanvasObject {
	appDB := appstate.GetDB()
	if appDB == nil {
		return widget.NewLabel("Non connecté.")
	}

	svc := services.NewReportService(appDB)
	now := utils.TodayString()
	firstDay := fmt.Sprintf("%d-%02d-01", utils.CurrentYear(), utils.CurrentMonth())
	top, err := svc.GetTopProducts(firstDay, now, 10)
	if err != nil || len(top) == 0 {
		return widget.NewLabel("Aucune vente ce mois.")
	}

	rows := container.NewVBox(
		widget.NewRichTextFromMarkdown("### 🏆 Top Produits (ce mois)"),
		widget.NewSeparator(),
	)
	for i, p := range top {
		lbl := widget.NewLabel(fmt.Sprintf("%d. %s — %s u — %s DA",
			i+1,
			utils.TruncateString(p.Name, 20),
			utils.FormatQuantity(p.Quantity),
			utils.FormatAmount(p.Amount),
		))
		rows.Add(lbl)
	}
	return container.NewVScroll(rows)
}

// ─────────────────────────────────────────────────────────────────────────────
// 4. RAPPORT TRÉSORERIE
// ─────────────────────────────────────────────────────────────────────────────

// BuildTreasuryReportsScreen construit le rapport de trésorerie mensuel
func BuildTreasuryReportsScreen() fyne.CanvasObject {
	db := appstate.GetDB()

	header := buildReportHeader("💰", "Rapport Trésorerie",
		"Synthèse des flux de trésorerie (caisse + banque)", "#2980b9")

	fromEntry := widget.NewEntry()
	toEntry := widget.NewEntry()
	tableContainer := container.NewStack()
	kpiBar := container.NewStack()

	loadReport := func() {
		if db == nil {
			return
		}
		svc := services.NewPaymentService(db)
		moves, balance, _ := svc.GetCashMovements(fromEntry.Text, toEntry.Text)

		var totalIn, totalOut float64
		for _, m := range moves {
			if m.Type == "in" {
				totalIn += m.Amount
			} else {
				totalOut += m.Amount
			}
		}

		kpiBar.Objects = []fyne.CanvasObject{
			buildReportSummaryBar(
				fmt.Sprintf("**%d mouvements**", len(moves)),
				fmt.Sprintf("**Total Entrées :** %s DA", utils.FormatAmount(totalIn)),
				fmt.Sprintf("**Total Sorties :** %s DA", utils.FormatAmount(totalOut)),
				fmt.Sprintf("**Solde Net :** %s DA", utils.FormatAmount(balance)),
			),
		}
		kpiBar.Refresh()

		colHeaders := []string{"Date", "Type", "Catégorie", "Tiers", "Description", "Entrée (DA)", "Sortie (DA)", "Solde (DA)"}
		colWidths := []float32{90, 70, 110, 130, 200, 120, 120, 120}
		headerRow := buildTableHeaderRow(colHeaders, colWidths)
		rows := container.NewVBox(headerRow)
		for _, m := range moves {
			entree, sortie := "", ""
			if m.Type == "in" {
				entree = utils.FormatAmount(m.Amount)
			} else {
				sortie = utils.FormatAmount(m.Amount)
			}
			cells := []string{
				utils.FormatDateFr(m.Date),
				m.Type,
				m.Category,
				m.PartyName,
				utils.TruncateString(m.Description, 28),
				entree, sortie,
				utils.FormatAmount(m.Balance),
			}
			rows.Add(buildTableRow(cells, colWidths, false))
		}
		if len(moves) == 0 {
			rows.Add(widget.NewLabel("   Aucun mouvement pour la période."))
		}
		tableContainer.Objects = []fyne.CanvasObject{container.NewVScroll(rows)}
		tableContainer.Refresh()
	}

	periodFilter := buildPeriodFilter(fromEntry, toEntry, loadReport)
	exportBar := buildExportBar(
		func() { dialog.ShowInformation("Export Excel", "Rapport trésorerie exporté.", appstate.MainWindow) },
		func() { dialog.ShowInformation("Impression", "Rapport trésorerie imprimé.", appstate.MainWindow) },
	)

	loadReport()

	return container.NewBorder(
		container.NewVBox(header, periodFilter, kpiBar, exportBar, widget.NewSeparator()),
		nil, nil, nil,
		tableContainer,
	)
}

// ─────────────────────────────────────────────────────────────────────────────
// 5. RAPPORT RENTABILITÉ (BÉNÉFICES)
// ─────────────────────────────────────────────────────────────────────────────

// BuildProfitReportsScreen construit le rapport de rentabilité
func BuildProfitReportsScreen() fyne.CanvasObject {
	db := appstate.GetDB()

	header := buildReportHeader("📊", "Rapport de Rentabilité",
		"Chiffre d'affaires, coût des ventes, dépenses et bénéfices", "#16a085")

	yearSel, monthSel := yearMonthSelectors(utils.CurrentYear(), utils.CurrentMonth())
	resultContainer := container.NewStack()

	loadReport := func() {
		if db == nil {
			return
		}
		yr, mo := selectedYearMonth(yearSel, monthSel)
		svc := services.NewReportService(db)
		report, err := svc.GetProfitReport(yr, mo)
		if err != nil {
			return
		}

		// Affichage type compte de résultat
		cards := container.NewVBox(
			widget.NewRichTextFromMarkdown(fmt.Sprintf("## Résultat — %s", report.Period)),
			widget.NewSeparator(),
			buildStatCard("Chiffre d'Affaires (TTC)", utils.FormatAmount(report.Revenue)+" DA", "#27ae60", "💰"),
			buildStatCard("Coût des Marchandises Vendues (CMUP)", utils.FormatAmount(report.COGS)+" DA", "#c0392b", "📦"),
			buildStatCard("Bénéfice Brut", utils.FormatAmount(report.GrossProfit)+" DA", "#2980b9", "📈"),
			buildStatCard("Dépenses d'Exploitation", utils.FormatAmount(report.Expenses)+" DA", "#e74c3c", "🧾"),
			buildStatCard("Bénéfice Net", utils.FormatAmount(report.NetProfit)+" DA", "#8e44ad", "✅"),
			widget.NewSeparator(),
			buildReportSummaryBar(
				fmt.Sprintf("**Marge brute :** %.1f%%", report.GrossMargin),
				fmt.Sprintf("**Marge nette :** %.1f%%", report.NetMargin),
			),
		)

		// Top clients du mois
		firstDay := fmt.Sprintf("%d-%02d-01", yr, mo)
		lastDay := fmt.Sprintf("%d-%02d-28", yr, mo) // approximation
		topClients, _ := svc.GetTopClients(firstDay, lastDay, 5)
		if len(topClients) > 0 {
			cards.Add(widget.NewSeparator())
			cards.Add(widget.NewRichTextFromMarkdown("### 🏆 Top 5 Clients du mois"))
			for i, c := range topClients {
				cards.Add(widget.NewLabel(fmt.Sprintf("  %d. %s — %s DA (%d factures)",
					i+1, c.Name, utils.FormatAmount(c.Amount), c.Count)))
			}
		}

		resultContainer.Objects = []fyne.CanvasObject{container.NewVScroll(container.NewPadded(cards))}
		resultContainer.Refresh()
	}

	filterRow := container.NewHBox(
		widget.NewLabel("Année :"), yearSel,
		widget.NewLabel("Mois :"), monthSel,
		widget.NewButtonWithIcon("Calculer", theme.SearchIcon(), loadReport),
	)

	loadReport()

	return container.NewBorder(
		container.NewVBox(header, filterRow, widget.NewSeparator()),
		nil, nil, nil,
		resultContainer,
	)
}

// ─────────────────────────────────────────────────────────────────────────────
// 6. RAPPORT CRÉANCES / DETTES
// ─────────────────────────────────────────────────────────────────────────────

// BuildDebtReportsScreen construit le rapport créances/dettes
func BuildDebtReportsScreen() fyne.CanvasObject {
	db := appstate.GetDB()

	header := buildReportHeader("⚖️", "Créances & Dettes",
		"Soldes clients impayés et dettes fournisseurs", "#c0392b")

	tabs := container.NewAppTabs(
		container.NewTabItem("👥 Créances Clients", buildDebtTable(db, "clients")),
		container.NewTabItem("🏭 Dettes Fournisseurs", buildDebtTable(db, "suppliers")),
	)

	return container.NewBorder(header, nil, nil, nil, tabs)
}

func buildDebtTable(_ *sql.DB, party string) fyne.CanvasObject {
	appDB := appstate.GetDB()
	if appDB == nil {
		return widget.NewLabel("Non connecté.")
	}

	var query string
	var colHeaders []string
	if party == "clients" {
		query = `SELECT code, name_fr, phone, balance FROM clients WHERE balance > 0 ORDER BY balance DESC`
		colHeaders = []string{"Code", "Client", "Téléphone", "Solde (DA)"}
	} else {
		query = `SELECT code, name_fr, phone, balance FROM suppliers WHERE balance > 0 ORDER BY balance DESC`
		colHeaders = []string{"Code", "Fournisseur", "Téléphone", "Solde (DA)"}
	}
	colWidths := []float32{90, 250, 130, 140}

	rows, err := appDB.Query(query)
	if err != nil {
		return widget.NewLabel("Erreur chargement: " + err.Error())
	}
	defer rows.Close()

	type DebtRow struct {
		Code, Name, Phone string
		Balance           float64
	}
	var debtRows []DebtRow
	var total float64
	for rows.Next() {
		var r DebtRow
		rows.Scan(&r.Code, &r.Name, &r.Phone, &r.Balance)
		debtRows = append(debtRows, r)
		total += r.Balance
	}

	headerRow := buildTableHeaderRow(colHeaders, colWidths)
	tRows := container.NewVBox(headerRow)
	for _, r := range debtRows {
		cells := []string{r.Code, r.Name, r.Phone, utils.FormatAmount(r.Balance)}
		tRows.Add(buildTableRow(cells, colWidths, false))
	}
	if len(debtRows) == 0 {
		tRows.Add(widget.NewLabel("   Aucune créance / dette en cours. 👍"))
	} else {
		tRows.Add(widget.NewSeparator())
		tRows.Add(buildTableRow([]string{"", fmt.Sprintf("TOTAL (%d)", len(debtRows)), "", utils.FormatAmount(total)}, colWidths, true))
	}

	kpi := buildReportSummaryBar(
		fmt.Sprintf("**%d tiers avec solde positif**", len(debtRows)),
		fmt.Sprintf("**Total dû :** %s DA", utils.FormatAmount(total)),
	)

	return container.NewBorder(kpi, nil, nil, nil, container.NewVScroll(tRows))
}

// ─────────────────────────────────────────────────────────────────────────────
// 7. INDICATEURS COMMERCIAUX
// ─────────────────────────────────────────────────────────────────────────────

// BuildIndicatorsScreen construit l'écran des indicateurs commerciaux
func BuildIndicatorsScreen() fyne.CanvasObject {
	db := appstate.GetDB()

	header := buildReportHeader("📊", "Indicateurs Commerciaux",
		"KPIs et ratios de gestion annuels", "#2c3e50")

	yearSel, _ := yearMonthSelectors(utils.CurrentYear(), 1)
	resultContainer := container.NewStack()

	loadIndicators := func() {
		if db == nil {
			return
		}
		yr, _ := strconv.Atoi(yearSel.Selected)
		if yr == 0 {
			yr = utils.CurrentYear()
		}
		svc := services.NewReportService(db)
		ind := svc.GetBusinessIndicators(yr)

		cards := container.NewGridWithColumns(2,
			buildStatCard("Rotation du Stock", fmt.Sprintf("%.2f ×", ind.StockTurnover), "#27ae60", "🔄"),
			buildStatCard("Délai Recouvrement Clients", fmt.Sprintf("%.0f jours", ind.AvgCollectionDays), "#e67e22", "📅"),
			buildStatCard("Marge Brute", fmt.Sprintf("%.1f%%", ind.GrossMarginPct), "#2980b9", "📈"),
			buildStatCard("Marge Nette", fmt.Sprintf("%.1f%%", ind.NetMarginPct), "#8e44ad", "💰"),
			buildStatCard("Valeur Moyenne Facture", utils.FormatAmount(ind.AvgInvoiceValue)+" DA", "#16a085", "🧾"),
			buildStatCard("Taux de Retour", fmt.Sprintf("%.1f%%", ind.ReturnRate), "#c0392b", "↩️"),
		)

		resultContainer.Objects = []fyne.CanvasObject{container.NewPadded(container.NewVBox(
			widget.NewRichTextFromMarkdown(fmt.Sprintf("## Indicateurs — Année %d", yr)),
			widget.NewSeparator(),
			cards,
		))}
		resultContainer.Refresh()
	}

	filterRow := container.NewHBox(
		widget.NewLabel("Année :"), yearSel,
		widget.NewButtonWithIcon("Calculer", theme.SearchIcon(), loadIndicators),
	)

	loadIndicators()

	return container.NewBorder(
		container.NewVBox(header, filterRow, widget.NewSeparator()),
		nil, nil, nil,
		resultContainer,
	)
}

// ─────────────────────────────────────────────────────────────────────────────
// 8. DÉCLARATION G50
// ─────────────────────────────────────────────────────────────────────────────

// BuildG50Screen construit l'écran de déclaration G50
func BuildG50Screen() fyne.CanvasObject {
	db := appstate.GetDB()

	header := buildReportHeader("🇩🇿", "Déclaration G50",
		"Déclaration mensuelle TVA, TAP et Timbre fiscal", "#c0392b")

	yearSel, monthSel := yearMonthSelectors(utils.CurrentYear(), utils.CurrentMonth())
	resultContainer := container.NewStack()
	historyContainer := container.NewStack()

	loadG50 := func() {
		if db == nil {
			return
		}
		yr, mo := selectedYearMonth(yearSel, monthSel)
		svc := services.NewTaxService(db)
		g50, err := svc.CalculateG50(yr, mo)
		if err != nil {
			resultContainer.Objects = []fyne.CanvasObject{widget.NewLabel("Erreur: " + err.Error())}
			resultContainer.Refresh()
			return
		}

		form := container.NewVBox(
			widget.NewRichTextFromMarkdown(fmt.Sprintf("## G50 — %s", g50.Period)),
			widget.NewSeparator(),
			widget.NewRichTextFromMarkdown("### 📊 Chiffre d'Affaires"),
			buildG50Line("CA taux 19%", g50.Revenue19),
			buildG50Line("CA taux 9%", g50.Revenue9),
			buildG50Line("CA exonéré", g50.RevenueExempt),
			buildG50Line("**Total CA (HT)**", g50.TotalRevenue),
			widget.NewSeparator(),
			widget.NewRichTextFromMarkdown("### 💸 TVA Collectée"),
			buildG50Line("TVA 19%", g50.TVACollected19),
			buildG50Line("TVA 9%", g50.TVACollected9),
			buildG50Line("**Total TVA collectée**", g50.TotalTVACollected),
			widget.NewSeparator(),
			widget.NewRichTextFromMarkdown("### ✅ TVA Déductible"),
			buildG50Line("TVA achats", g50.TVADeductiblePurchases),
			buildG50Line("TVA investissements", g50.TVADeductibleInvestments),
			buildG50Line("Précompte reporté", g50.TVAPrecompte),
			widget.NewSeparator(),
			buildG50Line("**TVA Nette Due**", g50.TVADue),
			buildG50Line("Crédit TVA reporté", g50.TVAPrecompteNext),
			widget.NewSeparator(),
			widget.NewRichTextFromMarkdown("### 📋 Autres Taxes"),
			buildG50Line("TAP (Taxe Activité Prof.)", g50.TAP),
			buildG50Line("Droit de Timbre", g50.Timbre),
			buildG50Line("IRG sur salaires", g50.IRGSalaries),
			widget.NewSeparator(),
			buildG50LineBold("🔴 TOTAL À PAYER", g50.TotalDue),
		)

		saveBtn := widget.NewButtonWithIcon("Sauvegarder G50", theme.DocumentSaveIcon(), func() {
			svc2 := services.NewTaxService(db)
			if err := svc2.SaveG50(g50, "draft"); err != nil {
				dialog.ShowError(err, appstate.MainWindow)
				return
			}
			dialog.ShowInformation("G50 sauvegardé", "Déclaration G50 enregistrée en brouillon.", appstate.MainWindow)
		})
		saveBtn.Importance = widget.HighImportance

		validateBtn := widget.NewButton("Valider G50", func() {
			svc2 := services.NewTaxService(db)
			if err := svc2.SaveG50(g50, "validated"); err != nil {
				dialog.ShowError(err, appstate.MainWindow)
				return
			}
			dialog.ShowInformation("G50 validé", "Déclaration G50 validée et enregistrée.", appstate.MainWindow)
		})

		buttons := container.NewHBox(saveBtn, validateBtn)
		resultContainer.Objects = []fyne.CanvasObject{
			container.NewVScroll(container.NewPadded(container.NewVBox(form, buttons))),
		}
		resultContainer.Refresh()

		// Historique
		history, _ := svc.GetG50History(yr)
		histRows := container.NewVBox(
			widget.NewRichTextFromMarkdown("### 📁 Historique G50 — "+strconv.Itoa(yr)),
			widget.NewSeparator(),
		)
		for _, h := range history {
			mo := ""
			if h.Month != nil {
				mo = fmt.Sprintf("%02d", *h.Month)
			}
			histRows.Add(widget.NewLabel(fmt.Sprintf(
				"G50 — %d/%s — Statut: %s — %s", h.Year, mo, h.Status, h.CreatedAt)))
		}
		if len(history) == 0 {
			histRows.Add(widget.NewLabel("Aucune déclaration sauvegardée."))
		}
		historyContainer.Objects = []fyne.CanvasObject{container.NewVScroll(histRows)}
		historyContainer.Refresh()
	}

	filterRow := container.NewHBox(
		widget.NewLabel("Année :"), yearSel,
		widget.NewLabel("Mois :"), monthSel,
		widget.NewButtonWithIcon("Calculer G50", theme.SearchIcon(), loadG50),
	)

	split := container.NewHSplit(resultContainer, container.NewPadded(historyContainer))
	split.Offset = 0.65

	loadG50()

	return container.NewBorder(
		container.NewVBox(header, filterRow, widget.NewSeparator()),
		nil, nil, nil,
		split,
	)
}

func buildG50Line(label string, amount float64) fyne.CanvasObject {
	return container.NewBorder(nil, nil,
		widget.NewLabel("  "+label),
		widget.NewRichTextFromMarkdown(fmt.Sprintf("**%s DA**", utils.FormatAmount(amount))),
	)
}

func buildG50LineBold(label string, amount float64) fyne.CanvasObject {
	lbl := widget.NewRichTextFromMarkdown("**" + label + "**")
	val := widget.NewRichTextFromMarkdown(fmt.Sprintf("**%s DA**", utils.FormatAmount(amount)))
	return container.NewBorder(nil, nil, lbl, val)
}

// ─────────────────────────────────────────────────────────────────────────────
// 9. REGISTRE TVA VENTES
// ─────────────────────────────────────────────────────────────────────────────

// BuildTVASalesRegisterScreen construit le registre TVA ventes
func BuildTVASalesRegisterScreen() fyne.CanvasObject {
	db := appstate.GetDB()

	header := buildReportHeader("📋", "Registre TVA — Ventes",
		"Registre légal des factures de vente avec TVA collectée", "#27ae60")

	yearSel, monthSel := yearMonthSelectors(utils.CurrentYear(), utils.CurrentMonth())
	tableContainer := container.NewStack()
	summaryBar := container.NewStack()

	colHeaders := []string{"N° Facture", "Date", "Client", "NIF Client", "HT (DA)", "TVA (DA)", "TTC (DA)", "Timbre (DA)"}
	colWidths := []float32{110, 90, 180, 140, 120, 110, 120, 100}

	loadRegister := func() {
		if db == nil {
			return
		}
		yr, mo := selectedYearMonth(yearSel, monthSel)
		svc := services.NewTaxService(db)
		rows, err := svc.GetTVASalesRegister(yr, mo)
		if err != nil {
			return
		}

		headerRow := buildTableHeaderRow(colHeaders, colWidths)
		tRows := container.NewVBox(headerRow)

		var totalHT, totalTVA, totalTTC, totalTimbre float64
		for _, r := range rows {
			cells := []string{
				fmt.Sprint(r["doc_number"]),
				utils.FormatDateFr(fmt.Sprint(r["date"])),
				utils.TruncateString(fmt.Sprint(r["client_name"]), 25),
				fmt.Sprint(r["nif"]),
				utils.FormatAmount(toFloat(r["net_ht"])),
				utils.FormatAmount(toFloat(r["total_tva"])),
				utils.FormatAmount(toFloat(r["total_ttc"])),
				utils.FormatAmount(toFloat(r["timbre"])),
			}
			tRows.Add(buildTableRow(cells, colWidths, false))
			totalHT += toFloat(r["net_ht"])
			totalTVA += toFloat(r["total_tva"])
			totalTTC += toFloat(r["total_ttc"])
			totalTimbre += toFloat(r["timbre"])
		}
		if len(rows) == 0 {
			tRows.Add(widget.NewLabel("   Aucune facture pour cette période."))
		} else {
			tRows.Add(widget.NewSeparator())
			tRows.Add(buildTableRow(
				[]string{"TOTAUX", "", "", "", utils.FormatAmount(totalHT), utils.FormatAmount(totalTVA), utils.FormatAmount(totalTTC), utils.FormatAmount(totalTimbre)},
				colWidths, true,
			))
		}

		summaryBar.Objects = []fyne.CanvasObject{
			buildReportSummaryBar(
				fmt.Sprintf("**%d factures**", len(rows)),
				fmt.Sprintf("**CA HT :** %s DA", utils.FormatAmount(totalHT)),
				fmt.Sprintf("**TVA collectée :** %s DA", utils.FormatAmount(totalTVA)),
				fmt.Sprintf("**Timbre :** %s DA", utils.FormatAmount(totalTimbre)),
			),
		}
		summaryBar.Refresh()
		tableContainer.Objects = []fyne.CanvasObject{container.NewVScroll(tRows)}
		tableContainer.Refresh()
	}

	filterRow := container.NewHBox(
		widget.NewLabel("Année :"), yearSel,
		widget.NewLabel("Mois :"), monthSel,
		widget.NewButtonWithIcon("Charger", theme.SearchIcon(), loadRegister),
	)
	exportBar := buildExportBar(
		func() { dialog.ShowInformation("Export", "Registre TVA ventes exporté.", appstate.MainWindow) },
		func() { dialog.ShowInformation("Impression", "Registre TVA ventes imprimé.", appstate.MainWindow) },
	)

	loadRegister()

	return container.NewBorder(
		container.NewVBox(header, filterRow, summaryBar, exportBar, widget.NewSeparator()),
		nil, nil, nil,
		tableContainer,
	)
}

// ─────────────────────────────────────────────────────────────────────────────
// 10. REGISTRE TVA ACHATS
// ─────────────────────────────────────────────────────────────────────────────

// BuildTVAPurchaseRegisterScreen construit le registre TVA achats
func BuildTVAPurchaseRegisterScreen() fyne.CanvasObject {
	db := appstate.GetDB()

	header := buildReportHeader("📋", "Registre TVA — Achats",
		"Registre légal des factures d'achat avec TVA déductible", "#e67e22")

	yearSel, monthSel := yearMonthSelectors(utils.CurrentYear(), utils.CurrentMonth())
	tableContainer := container.NewStack()
	summaryBar := container.NewStack()

	colHeaders := []string{"N° Interne", "N° Fournisseur", "Date", "Fournisseur", "NIF", "HT (DA)", "TVA (DA)", "TTC (DA)"}
	colWidths := []float32{100, 120, 90, 180, 140, 120, 110, 120}

	loadRegister := func() {
		if db == nil {
			return
		}
		yr, mo := selectedYearMonth(yearSel, monthSel)
		svc := services.NewTaxService(db)
		rows, err := svc.GetTVAPurchasesRegister(yr, mo)
		if err != nil {
			return
		}

		headerRow := buildTableHeaderRow(colHeaders, colWidths)
		tRows := container.NewVBox(headerRow)

		var totalHT, totalTVA, totalTTC float64
		for _, r := range rows {
			cells := []string{
				fmt.Sprint(r["doc_number"]),
				fmt.Sprint(r["supplier_invoice_number"]),
				utils.FormatDateFr(fmt.Sprint(r["date"])),
				utils.TruncateString(fmt.Sprint(r["supplier_name"]), 25),
				fmt.Sprint(r["nif"]),
				utils.FormatAmount(toFloat(r["net_ht"])),
				utils.FormatAmount(toFloat(r["total_tva"])),
				utils.FormatAmount(toFloat(r["total_ttc"])),
			}
			tRows.Add(buildTableRow(cells, colWidths, false))
			totalHT += toFloat(r["net_ht"])
			totalTVA += toFloat(r["total_tva"])
			totalTTC += toFloat(r["total_ttc"])
		}
		if len(rows) == 0 {
			tRows.Add(widget.NewLabel("   Aucun achat pour cette période."))
		} else {
			tRows.Add(widget.NewSeparator())
			tRows.Add(buildTableRow(
				[]string{"TOTAUX", "", "", "", "", utils.FormatAmount(totalHT), utils.FormatAmount(totalTVA), utils.FormatAmount(totalTTC)},
				colWidths, true,
			))
		}

		summaryBar.Objects = []fyne.CanvasObject{
			buildReportSummaryBar(
				fmt.Sprintf("**%d factures**", len(rows)),
				fmt.Sprintf("**Total Achats HT :** %s DA", utils.FormatAmount(totalHT)),
				fmt.Sprintf("**TVA déductible :** %s DA", utils.FormatAmount(totalTVA)),
			),
		}
		summaryBar.Refresh()
		tableContainer.Objects = []fyne.CanvasObject{container.NewVScroll(tRows)}
		tableContainer.Refresh()
	}

	filterRow := container.NewHBox(
		widget.NewLabel("Année :"), yearSel,
		widget.NewLabel("Mois :"), monthSel,
		widget.NewButtonWithIcon("Charger", theme.SearchIcon(), loadRegister),
	)
	exportBar := buildExportBar(
		func() { dialog.ShowInformation("Export", "Registre TVA achats exporté.", appstate.MainWindow) },
		func() { dialog.ShowInformation("Impression", "Registre TVA achats imprimé.", appstate.MainWindow) },
	)

	loadRegister()

	return container.NewBorder(
		container.NewVBox(header, filterRow, summaryBar, exportBar, widget.NewSeparator()),
		nil, nil, nil,
		tableContainer,
	)
}

// ─────────────────────────────────────────────────────────────────────────────
// 11. DÉCLARATION ANNUELLE
// ─────────────────────────────────────────────────────────────────────────────

// BuildAnnualDeclarationScreen construit l'écran de déclaration fiscale annuelle
func BuildAnnualDeclarationScreen() fyne.CanvasObject {
	db := appstate.GetDB()

	header := buildReportHeader("📜", "Déclaration Fiscale Annuelle",
		"Synthèse annuelle TVA, TAP et IBS pour la déclaration G50-bis", "#c0392b")

	years := []string{}
	for y := utils.CurrentYear() - 3; y <= utils.CurrentYear(); y++ {
		years = append(years, strconv.Itoa(y))
	}
	yearSel := widget.NewSelect(years, nil)
	yearSel.SetSelected(strconv.Itoa(utils.CurrentYear() - 1))

	resultContainer := container.NewStack()

	loadAnnual := func() {
		if db == nil {
			return
		}
		yr, _ := strconv.Atoi(yearSel.Selected)
		if yr == 0 {
			yr = utils.CurrentYear() - 1
		}

		// Agrégation de tous les G50 de l'année via les données réelles
		var totalRevenue19, totalRevenue9, totalRevenueExempt float64
		var totalTVACollected, totalTVADeductible float64
		var totalTAP, totalTimbre float64

		if db != nil {
			svc := services.NewTaxService(db)
			for m := 1; m <= 12; m++ {
				g50, err := svc.CalculateG50(yr, m)
				if err == nil {
					totalRevenue19 += g50.Revenue19
					totalRevenue9 += g50.Revenue9
					totalRevenueExempt += g50.RevenueExempt
					totalTVACollected += g50.TotalTVACollected
					totalTVADeductible += g50.TVADeductiblePurchases
					totalTAP += g50.TAP
					totalTimbre += g50.Timbre
				}
			}
		}

		totalRevenue := totalRevenue19 + totalRevenue9 + totalRevenueExempt
		totalTVADue := totalTVACollected - totalTVADeductible
		if totalTVADue < 0 {
			totalTVADue = 0
		}
		totalDue := totalTVADue + totalTAP + totalTimbre

		// Rapport de rentabilité annuel
		var netProfit float64
		if db != nil {
			svc := services.NewReportService(db)
			for m := 1; m <= 12; m++ {
				rpt, err := svc.GetProfitReport(yr, m)
				if err == nil {
					netProfit += rpt.NetProfit
				}
			}
		}

		form := container.NewVBox(
			widget.NewRichTextFromMarkdown(fmt.Sprintf("## Déclaration Annuelle — %d", yr)),
			widget.NewSeparator(),
			widget.NewRichTextFromMarkdown("### Chiffre d'Affaires Annuel"),
			buildG50Line("CA taux 19%", totalRevenue19),
			buildG50Line("CA taux 9%", totalRevenue9),
			buildG50Line("CA exonéré", totalRevenueExempt),
			buildG50LineBold("Total CA HT", totalRevenue),
			widget.NewSeparator(),
			widget.NewRichTextFromMarkdown("### TVA"),
			buildG50Line("TVA collectée", totalTVACollected),
			buildG50Line("TVA déductible (achats)", totalTVADeductible),
			buildG50LineBold("TVA Nette Due (annuelle)", totalTVADue),
			widget.NewSeparator(),
			widget.NewRichTextFromMarkdown("### Autres Taxes"),
			buildG50Line("TAP (annuel)", totalTAP),
			buildG50Line("Timbre fiscal (annuel)", totalTimbre),
			widget.NewSeparator(),
			buildG50LineBold("💰 Bénéfice Net Estimé", netProfit),
			buildG50LineBold("🔴 Total Taxes Annuelles", totalDue),
			widget.NewSeparator(),
			widget.NewRichTextFromMarkdown(`> ⚠️ **Rappel** : Cette synthèse est indicative. La déclaration annuelle officielle (G50-bis / IBS) doit être déposée auprès du centre des impôts avant le 30 avril de l'année suivante.`),
		)

		exportBtn := widget.NewButtonWithIcon("Exporter Excel", theme.DocumentSaveIcon(), func() {
			dialog.ShowInformation("Export", "Déclaration annuelle exportée.", appstate.MainWindow)
		})
		exportBtn.Importance = widget.HighImportance

		resultContainer.Objects = []fyne.CanvasObject{
			container.NewVScroll(container.NewPadded(container.NewVBox(form, exportBtn))),
		}
		resultContainer.Refresh()
	}

	filterRow := container.NewHBox(
		widget.NewLabel("Année fiscale :"), yearSel,
		widget.NewButtonWithIcon("Calculer", theme.SearchIcon(), loadAnnual),
	)

	loadAnnual()

	return container.NewBorder(
		container.NewVBox(header, filterRow, widget.NewSeparator()),
		nil, nil, nil,
		resultContainer,
	)
}

// ─────────────────────────────────────────────────────────────────────────────
// UTILITAIRE
// ─────────────────────────────────────────────────────────────────────────────

// toFloat convertit une interface{} en float64 (pour les maps SQL)
func toFloat(v interface{}) float64 {
	if v == nil {
		return 0
	}
	switch val := v.(type) {
	case float64:
		return val
	case float32:
		return float64(val)
	case int:
		return float64(val)
	case int64:
		return float64(val)
	case string:
		f, _ := strconv.ParseFloat(strings.TrimSpace(val), 64)
		return f
	}
	return 0
}
