package screens

import (
	"fmt"
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"gestion-commerciale/internal/app"
	"gestion-commerciale/internal/database/queries"
	"gestion-commerciale/pkg/utils"
)

// BuildDashboardScreen construit le tableau de bord principal
func BuildDashboardScreen() fyne.CanvasObject {
	db := app.GetDB()

	var stats queries.DashboardStats
	if db != nil {
		stats = queries.GetDashboardStats(db)
	}

	session := app.GetSession()
	welcomeMsg := "Bienvenue"
	companyMsg := "Gestion Commerciale Pro"
	if session != nil {
		welcomeMsg = fmt.Sprintf("Bienvenue, %s", session.FullName)
		companyMsg = fmt.Sprintf("%s — Exercice %d", session.CompanyName, session.FiscalYear)
	}

	// ── Bannière supérieure ───────────────────────────────────────────────────
	headerBg := canvas.NewRectangle(color.RGBA{R: 0x1a, G: 0x3a, B: 0x6e, A: 0xff})
	headerBg.CornerRadius = 10

	titleTxt := canvas.NewText("📊  "+welcomeMsg, color.White)
	titleTxt.TextSize = 20
	titleTxt.TextStyle = fyne.TextStyle{Bold: true}

	subTxt := canvas.NewText(companyMsg, color.RGBA{R: 0xd0, G: 0xdc, B: 0xf0, A: 0xff})
	subTxt.TextSize = 13

	headerContent := container.NewVBox(
		container.NewPadded(titleTxt),
		container.NewPadded(subTxt),
	)
	banner := container.NewStack(headerBg, container.NewPadded(headerContent))

	// ── Cartes KPI ───────────────────────────────────────────────────────────
	kpiCards := container.NewGridWithColumns(4,
		buildKPICard("Ventes du Jour", utils.FormatMoney(stats.TodaySales), "↗", "#27ae60", "#1e8449"),
		buildKPICard("Ventes du Mois", utils.FormatMoney(stats.MonthSales), "📈", "#2ecc71", "#239b56"),
		buildKPICard("Achats du Mois", utils.FormatMoney(stats.MonthPurchases), "📉", "#2980b9", "#1a6fa3"),
		buildKPICard("Solde Caisse", utils.FormatMoney(stats.CashBalance), "💵", "#f39c12", "#d68910"),
	)
	kpiCards2 := container.NewGridWithColumns(3,
		buildKPICard("Créances Clients", utils.FormatMoney(stats.ClientDebts), "👤", "#e74c3c", "#cb4335"),
		buildKPICard("Dettes Fournisseurs", utils.FormatMoney(stats.SupplierDebts), "🏭", "#9b59b6", "#884ea0"),
		buildKPICard("Articles Sous Seuil", fmt.Sprintf("%d articles", stats.LowStockCount), "⚠️", "#e67e22", "#ca6f1e"),
	)

	// ── Accès rapide ──────────────────────────────────────────────────────────
	quickLabel := buildSectionTitle("⚡  Accès Rapide")

	makeQuick := func(icon, label, route string, c color.RGBA) fyne.CanvasObject {
		bg := canvas.NewRectangle(c)
		bg.CornerRadius = 8
		btn := widget.NewButton(icon+"  "+label, func() { app.Navigate(route) })
		btn.Importance = widget.MediumImportance
		return container.NewStack(bg, btn)
	}

	quickBtns := container.NewGridWithColumns(4,
		makeQuick("🧾", "Nouvelle Facture", "sales/invoice", color.RGBA{R: 0x27, G: 0xae, B: 0x60, A: 0xff}),
		makeQuick("🛒", "Nouvel Achat", "purchases/invoice", color.RGBA{R: 0x29, G: 0x80, B: 0xb9, A: 0xff}),
		makeQuick("👤", "Nouveau Client", "tiers/clients", color.RGBA{R: 0x9b, G: 0x59, B: 0xb6, A: 0xff}),
		makeQuick("📦", "Nouvel Article", "articles", color.RGBA{R: 0xe6, G: 0x7e, B: 0x22, A: 0xff}),
		makeQuick("💵", "Encaissement", "treasury/collections", color.RGBA{R: 0xf3, G: 0x9c, B: 0x12, A: 0xff}),
		makeQuick("💸", "Décaissement", "treasury/disbursements", color.RGBA{R: 0xe7, G: 0x4c, B: 0x3c, A: 0xff}),
		makeQuick("📊", "Rapport Ventes", "reports/sales", color.RGBA{R: 0x1a, G: 0x3a, B: 0x6e, A: 0xff}),
		makeQuick("🛒", "Point de Vente", "pos", color.RGBA{R: 0x16, G: 0xa0, B: 0x85, A: 0xff}),
	)

	// ── Dernières opérations ──────────────────────────────────────────────────
	recentLabel := buildSectionTitle("📋  Dernières Opérations")

	headers := []string{"Type", "Numéro", "Date", "Client / Fournisseur", "Montant (DA)", "Statut"}
	colWidths := []float32{70, 150, 100, 200, 130, 100}

	headerRow := buildDashTableHeader(headers, colWidths)
	rowsBox := container.NewVBox(headerRow)

	if db != nil {
		docs, _ := queries.GetRecentDocuments(db, 10)
		for i, doc := range docs {
			d := doc
			bg := color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff}
			if i%2 == 0 {
				bg = color.RGBA{R: 0xf4, G: 0xf6, B: 0xf9, A: 0xff}
			}
			cells := []string{
				d.DocType,
				d.DocNumber,
				utils.FormatDateFr(d.Date),
				d.ClientName,
				utils.FormatMoney(d.NetAmount),
				utils.StatusLabel(d.Status),
			}
			rowsBox.Add(buildDashTableRow(cells, colWidths, bg))
		}
	}

	recentCard := buildCard(container.NewVBox(recentLabel, rowsBox))

	// ── Assemblage final ──────────────────────────────────────────────────────
	content := container.NewVBox(
		banner,
		widget.NewLabel(""), // spacer
		kpiCards,
		kpiCards2,
		widget.NewLabel(""), // spacer
		buildCard(container.NewVBox(quickLabel, quickBtns)),
		widget.NewLabel(""), // spacer
		recentCard,
	)

	return container.NewPadded(container.NewVScroll(content))
}

// ── Helpers visuels ───────────────────────────────────────────────────────────

// buildKPICard construit une carte KPI avec dégradé simulé
func buildKPICard(title, value, icon, hexTop, hexBottom string) fyne.CanvasObject {
	bgTop := canvas.NewRectangle(parseHexColor(hexTop))
	bgTop.CornerRadius = 10

	// Bande décorative en bas
	bgBottom := canvas.NewRectangle(parseHexColor(hexBottom))
	bgBottom.CornerRadius = 10

	iconTxt := canvas.NewText(icon, color.White)
	iconTxt.TextSize = 28

	valueTxt := canvas.NewText(value, color.White)
	valueTxt.TextSize = 17
	valueTxt.TextStyle = fyne.TextStyle{Bold: true}

	titleTxt := canvas.NewText(title, color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xcc})
	titleTxt.TextSize = 11

	content := container.NewVBox(
		container.NewHBox(container.NewPadded(iconTxt)),
		container.NewPadded(valueTxt),
		container.NewPadded(titleTxt),
	)

	return container.NewStack(
		bgBottom,
		bgTop,
		container.NewPadded(content),
	)
}

// buildCard entoure un contenu d'une carte blanche avec ombre simulée
func buildCard(content fyne.CanvasObject) fyne.CanvasObject {
	shadow := canvas.NewRectangle(color.RGBA{R: 0x1a, G: 0x3a, B: 0x6e, A: 0x18})
	shadow.CornerRadius = 12

	bg := canvas.NewRectangle(color.White)
	bg.CornerRadius = 10

	return container.NewStack(shadow, bg, container.NewPadded(content))
}

// buildSectionTitle construit un titre de section
func buildSectionTitle(title string) fyne.CanvasObject {
	lbl := widget.NewLabel(title)
	lbl.TextStyle = fyne.TextStyle{Bold: true}
	sep := widget.NewSeparator()
	return container.NewVBox(lbl, sep)
}

// buildDashTableHeader construit l'en-tête du tableau de bord
func buildDashTableHeader(headers []string, widths []float32) fyne.CanvasObject {
	hdrBg := canvas.NewRectangle(color.RGBA{R: 0x1a, G: 0x3a, B: 0x6e, A: 0xff})
	row := container.NewHBox()
	for i, h := range headers {
		lbl := canvas.NewText(h, color.White)
		lbl.TextStyle = fyne.TextStyle{Bold: true}
		lbl.TextSize = 12
		cell := container.NewStack(
			canvas.NewRectangle(color.Transparent),
			container.NewPadded(lbl),
		)
		cell.Resize(fyne.NewSize(widths[i], 30))
		row.Add(cell)
	}
	return container.NewStack(hdrBg, row)
}

// buildDashTableRow construit une ligne du tableau de bord
func buildDashTableRow(cells []string, widths []float32, bg color.RGBA) fyne.CanvasObject {
	bgRect := canvas.NewRectangle(bg)
	row := container.NewHBox()
	for i, c := range cells {
		lbl := widget.NewLabel(c)
		lbl.Wrapping = fyne.TextWrapOff
		cell := container.NewStack(
			canvas.NewRectangle(color.Transparent),
			container.NewPadded(lbl),
		)
		cell.Resize(fyne.NewSize(widths[i], 28))
		row.Add(cell)
	}
	return container.NewStack(bgRect, row)
}

// parseHexColor convertit une couleur hexadécimale en color.RGBA
func parseHexColor(hex string) color.RGBA {
	if len(hex) != 7 || hex[0] != '#' {
		return color.RGBA{100, 100, 100, 255}
	}
	var r, g, b uint8
	fmt.Sscanf(hex[1:3], "%02x", &r)
	fmt.Sscanf(hex[3:5], "%02x", &g)
	fmt.Sscanf(hex[5:7], "%02x", &b)
	return color.RGBA{r, g, b, 255}
}

// Garder l'import theme utilisé
var _ = theme.HomeIcon
