package screens

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// ── Couleurs sidebar ──────────────────────────────────────────────────────────
var (
	sidebarBgColor   = color.RGBA{R: 0x10, G: 0x24, B: 0x48, A: 0xff} // bleu nuit profond
	sidebarTextColor = color.RGBA{R: 0xd8, G: 0xe6, B: 0xf8, A: 0xff} // bleu pâle
	sidebarGoldColor = color.RGBA{R: 0xf0, G: 0xb4, B: 0x29, A: 0xff} // or
	sidebarDivColor  = color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0x18} // blanc 10%
)

// BuildSidebar construit la barre latérale de navigation moderne
func BuildSidebar(navigate func(route string, params ...interface{})) fyne.CanvasObject {

	// ── En-tête ───────────────────────────────────────────────────────────────
	headerBg := canvas.NewRectangle(color.RGBA{R: 0x08, G: 0x14, B: 0x2c, A: 0xff})

	circBg := canvas.NewCircle(sidebarGoldColor)
	circBg.Resize(fyne.NewSize(52, 52))

	logoLetter := canvas.NewText("G", color.White)
	logoLetter.TextSize = 30
	logoLetter.TextStyle = fyne.TextStyle{Bold: true}

	logoStack := container.NewStack(
		container.NewCenter(circBg),
		container.NewCenter(logoLetter),
	)

	appNameLbl := canvas.NewText("Gestion Commerciale", sidebarTextColor)
	appNameLbl.TextSize = 12
	appNameLbl.TextStyle = fyne.TextStyle{Bold: true}

	proLbl := canvas.NewText("Pro v2.0", sidebarGoldColor)
	proLbl.TextSize = 10

	headerContent := container.NewVBox(
		container.NewCenter(logoStack),
		container.NewCenter(appNameLbl),
		container.NewCenter(proLbl),
	)

	header := container.NewStack(
		headerBg,
		container.NewPadded(headerContent),
	)

	// ── Helpers ───────────────────────────────────────────────────────────────
	makeSection := func(label string) fyne.CanvasObject {
		divLine := canvas.NewLine(sidebarDivColor)
		divLine.StrokeWidth = 1

		lbl := canvas.NewText("  " + label, sidebarGoldColor)
		lbl.TextSize = 9
		lbl.TextStyle = fyne.TextStyle{Bold: true}

		return container.NewVBox(
			container.NewPadded(divLine),
			container.NewPadded(lbl),
		)
	}

	makeMenuBtn := func(icon, label, route string) fyne.CanvasObject {
		btn := widget.NewButton(icon+"  "+label, func() {
			navigate(route)
		})
		btn.Alignment = widget.ButtonAlignLeading
		return btn
	}

	makeHighBtn := func(icon, label, route string) fyne.CanvasObject {
		btn := widget.NewButton(icon+"  "+label, func() {
			navigate(route)
		})
		btn.Importance = widget.HighImportance
		btn.Alignment = widget.ButtonAlignLeading
		return btn
	}

	// ── Menu items ────────────────────────────────────────────────────────────
	menuItems := container.NewVBox(

		makeHighBtn("🏠", "Tableau de Bord", "dashboard"),

		makeSection("ARTICLES & STOCK"),
		makeMenuBtn("📦", "Catalogue Articles", "articles"),
		makeMenuBtn("📂", "Familles / Catégories", "articles/categories"),
		makeMenuBtn("🏷️", "Marques", "articles/brands"),
		makeMenuBtn("📏", "Unités de Mesure", "articles/units"),
		makeMenuBtn("📊", "Inventaire", "articles/inventory"),
		makeMenuBtn("🔄", "Mouvements de Stock", "articles/stock_movements"),
		makeMenuBtn("🏬", "Dépôts / Magasins", "articles/warehouses"),
		makeMenuBtn("💲", "Listes de Prix", "articles/price_lists"),

		makeSection("VENTES"),
		makeHighBtn("🧾", "Nouvelle Facture Vente", "sales/invoice"),
		makeMenuBtn("📋", "Liste Factures Vente", "sales/invoices"),
		makeMenuBtn("📝", "Devis", "sales/quotations"),
		makeMenuBtn("📄", "Facture Proforma", "sales/proforma"),
		makeMenuBtn("📦", "Bons de Livraison", "sales/delivery_notes"),
		makeMenuBtn("📑", "Commandes Clients", "sales/client_orders"),
		makeMenuBtn("↩️", "Avoirs / Retours", "sales/credit_notes"),

		makeSection("ACHATS"),
		makeHighBtn("🛒", "Nouvelle Facture Achat", "purchases/invoice"),
		makeMenuBtn("📋", "Liste Factures Achat", "purchases/invoices"),
		makeMenuBtn("📦", "Bons de Réception", "purchases/reception_notes"),
		makeMenuBtn("📑", "Commandes Fournisseurs", "purchases/supplier_orders"),
		makeMenuBtn("↪️", "Retours Fournisseurs", "purchases/returns"),

		makeSection("TIERS"),
		makeMenuBtn("👤", "Clients", "tiers/clients"),
		makeMenuBtn("🏭", "Fournisseurs", "tiers/suppliers"),
		makeMenuBtn("🚛", "Chauffeurs / Livreurs", "tiers/drivers"),

		makeSection("TRÉSORERIE"),
		makeMenuBtn("💵", "Caisse", "treasury/cash"),
		makeMenuBtn("🏛️", "Banque", "treasury/bank"),
		makeMenuBtn("📋", "Chèques", "treasury/cheques"),
		makeMenuBtn("💰", "Encaissements", "treasury/collections"),
		makeMenuBtn("💸", "Décaissements", "treasury/disbursements"),
		makeMenuBtn("📊", "Balance Âgée", "treasury/aging"),
		makeMenuBtn("🧾", "Dépenses Diverses", "treasury/expenses"),

		makeSection("FISCALITÉ"),
		makeMenuBtn("📋", "Déclaration G50", "tax/g50"),
		makeMenuBtn("📊", "Registre TVA Ventes", "tax/tva_sales"),
		makeMenuBtn("📊", "Registre TVA Achats", "tax/tva_purchases"),
		makeMenuBtn("📅", "Déclaration Annuelle", "tax/annual"),

		makeSection("RAPPORTS"),
		makeMenuBtn("📈", "Rapports Ventes", "reports/sales"),
		makeMenuBtn("📉", "Rapports Achats", "reports/purchases"),
		makeMenuBtn("📦", "Rapports Stock", "reports/stock"),
		makeMenuBtn("💰", "Rapports Trésorerie", "reports/treasury"),
		makeMenuBtn("💹", "Bénéfices & Marges", "reports/profits"),
		makeMenuBtn("📊", "Rapports Dettes", "reports/debts"),
		makeMenuBtn("📊", "Indicateurs KPI", "tools/indicators"),

		makeSection("VENTE RAPIDE"),
		makeHighBtn("🛒", "Point de Vente (POS)", "pos"),

		makeSection("PARAMÈTRES"),
		makeMenuBtn("🏢", "Informations Société", "settings/company"),
		makeMenuBtn("📅", "Années Fiscales", "settings/fiscal_years"),
		makeMenuBtn("🔢", "Numérotation", "settings/numbering"),
		makeMenuBtn("💰", "Taxes & TVA", "settings/taxes"),
		makeMenuBtn("🖨️", "Impression", "settings/print"),
		makeMenuBtn("👥", "Utilisateurs", "settings/users"),
		makeMenuBtn("💱", "Devises", "settings/currencies"),
		makeMenuBtn("🏷️", "Code-barres", "settings/barcode"),

		makeSection("OUTILS"),
		makeMenuBtn("💾", "Sauvegarde", "tools/backup"),
		makeMenuBtn("📂", "Restauration", "tools/restore"),
		makeMenuBtn("🧮", "Calculatrice TVA", "tools/calculator"),
		makeMenuBtn("💲", "Mise à Jour Prix", "tools/price_update"),
		makeMenuBtn("📅", "Calendrier & Rappels", "tools/calendar"),
		makeMenuBtn("📋", "Journal d'Audit", "tools/audit_log"),
		makeMenuBtn("🖨️", "Centre d'Impression", "print_center"),

		widget.NewSeparator(),
		makeMenuBtn("ℹ️", "À Propos", "help/about"),
	)

	scroll := container.NewVScroll(menuItems)
	scroll.SetMinSize(fyne.NewSize(240, 0))

	bg := canvas.NewRectangle(sidebarBgColor)

	return container.NewStack(
		bg,
		container.NewBorder(header, nil, nil, nil, scroll),
	)
}
