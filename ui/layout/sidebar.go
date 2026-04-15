package layout

// sidebar.go — Barre latérale de navigation
// Implémentation complète dans le Lot 4

// Ce fichier contient la définition de la structure de la sidebar
// L'implémentation Fyne complète sera dans le Lot 4

// SidebarItem représente un élément du menu latéral
type SidebarItem struct {
	ID       string
	Label    string
	Icon     string
	Route    string
	Children []SidebarItem
}

// GetSidebarStructure retourne la structure complète du menu
func GetSidebarStructure() []SidebarItem {
	return []SidebarItem{
		{
			ID: "dashboard", Label: "Tableau de Bord",
			Icon: "home", Route: "dashboard",
		},
		{
			ID: "settings", Label: "Paramètres",
			Icon: "settings",
			Children: []SidebarItem{
				{ID: "company", Label: "Informations Société", Route: "settings/company"},
				{ID: "fiscal_years", Label: "Années Fiscales", Route: "settings/fiscal_years"},
				{ID: "numbering", Label: "Numérotation", Route: "settings/numbering"},
				{ID: "taxes", Label: "Taxes & TVA", Route: "settings/taxes"},
				{ID: "print", Label: "Impression", Route: "settings/print"},
				{ID: "users", Label: "Utilisateurs", Route: "settings/users"},
				{ID: "currencies", Label: "Devises", Route: "settings/currencies"},
				{ID: "barcode", Label: "Code-barres", Route: "settings/barcode"},
			},
		},
		{
			ID: "articles", Label: "Articles & Stock",
			Icon: "package",
			Children: []SidebarItem{
				{ID: "articles_list", Label: "Catalogue Articles", Route: "articles"},
				{ID: "categories", Label: "Familles / Catégories", Route: "articles/categories"},
				{ID: "brands", Label: "Marques", Route: "articles/brands"},
				{ID: "units", Label: "Unités de Mesure", Route: "articles/units"},
				{ID: "inventory", Label: "Inventaire", Route: "articles/inventory"},
				{ID: "stock_movements", Label: "Mouvements de Stock", Route: "articles/stock_movements"},
				{ID: "warehouses", Label: "Dépôts / Magasins", Route: "articles/warehouses"},
				{ID: "price_lists", Label: "Listes de Prix", Route: "articles/price_lists"},
			},
		},
		{
			ID: "sales", Label: "Ventes",
			Icon: "sales",
			Children: []SidebarItem{
				{ID: "sale_invoice", Label: "Facture de Vente", Route: "sales/invoice"},
				{ID: "sale_invoices", Label: "Liste Factures Vente", Route: "sales/invoices"},
				{ID: "quotations", Label: "Devis", Route: "sales/quotations"},
				{ID: "proforma", Label: "Facture Proforma", Route: "sales/proforma"},
				{ID: "delivery_notes", Label: "Bons de Livraison", Route: "sales/delivery_notes"},
				{ID: "client_orders", Label: "Commandes Clients", Route: "sales/client_orders"},
				{ID: "credit_notes", Label: "Avoirs", Route: "sales/credit_notes"},
			},
		},
		{
			ID: "purchases", Label: "Achats",
			Icon: "cart",
			Children: []SidebarItem{
				{ID: "purchase_invoice", Label: "Facture d'Achat", Route: "purchases/invoice"},
				{ID: "purchase_invoices", Label: "Liste Factures Achat", Route: "purchases/invoices"},
				{ID: "reception_notes", Label: "Bons de Réception", Route: "purchases/reception_notes"},
				{ID: "supplier_orders", Label: "Commandes Fournisseurs", Route: "purchases/supplier_orders"},
				{ID: "purchase_returns", Label: "Retours Fournisseurs", Route: "purchases/returns"},
			},
		},
		{
			ID: "tiers", Label: "Tiers",
			Icon: "people",
			Children: []SidebarItem{
				{ID: "clients", Label: "Clients", Route: "tiers/clients"},
				{ID: "suppliers", Label: "Fournisseurs", Route: "tiers/suppliers"},
				{ID: "drivers", Label: "Chauffeurs / Livreurs", Route: "tiers/drivers"},
			},
		},
		{
			ID: "treasury", Label: "Trésorerie",
			Icon: "money",
			Children: []SidebarItem{
				{ID: "cash", Label: "Caisse", Route: "treasury/cash"},
				{ID: "bank", Label: "Banque", Route: "treasury/bank"},
				{ID: "cheques", Label: "Chèques", Route: "treasury/cheques"},
				{ID: "collections", Label: "Encaissements", Route: "treasury/collections"},
				{ID: "disbursements", Label: "Décaissements", Route: "treasury/disbursements"},
				{ID: "aging", Label: "Balance Âgée", Route: "treasury/aging"},
				{ID: "expenses", Label: "Dépenses Diverses", Route: "treasury/expenses"},
			},
		},
		{
			ID: "tax", Label: "Fiscalité",
			Icon: "tax",
			Children: []SidebarItem{
				{ID: "g50", Label: "Déclaration G50", Route: "tax/g50"},
				{ID: "tva_sales", Label: "Registre TVA Ventes", Route: "tax/tva_sales"},
				{ID: "tva_purchases", Label: "Registre TVA Achats", Route: "tax/tva_purchases"},
				{ID: "annual", Label: "Déclaration Annuelle", Route: "tax/annual"},
			},
		},
		{
			ID: "reports", Label: "Rapports",
			Icon: "chart",
			Children: []SidebarItem{
				{ID: "sales_reports", Label: "Rapports Ventes", Route: "reports/sales"},
				{ID: "purchase_reports", Label: "Rapports Achats", Route: "reports/purchases"},
				{ID: "stock_reports", Label: "Rapports Stock", Route: "reports/stock"},
				{ID: "treasury_reports", Label: "Rapports Trésorerie", Route: "reports/treasury"},
				{ID: "profit_reports", Label: "Rapports Bénéfices", Route: "reports/profits"},
				{ID: "debt_reports", Label: "Rapports Dettes", Route: "reports/debts"},
			},
		},
		{
			ID: "pos", Label: "Point de Vente (POS)",
			Icon: "pos", Route: "pos",
		},
		{
			ID: "print_center", Label: "Centre d'Impression",
			Icon: "printer", Route: "print_center",
		},
		{
			ID: "tools", Label: "Outils",
			Icon: "tools",
			Children: []SidebarItem{
				{ID: "backup", Label: "Sauvegarde", Route: "tools/backup"},
				{ID: "restore", Label: "Restauration", Route: "tools/restore"},
				{ID: "calculator", Label: "Calculatrice TVA", Route: "tools/calculator"},
				{ID: "price_update", Label: "Mise à Jour Prix", Route: "tools/price_update"},
				{ID: "indicators", Label: "Indicateurs Commerciaux", Route: "tools/indicators"},
				{ID: "calendar", Label: "Calendrier & Rappels", Route: "tools/calendar"},
				{ID: "audit_log", Label: "Journal d'Audit", Route: "tools/audit_log"},
			},
		},
		{
			ID: "help", Label: "Aide",
			Icon: "help",
			Children: []SidebarItem{
				{ID: "shortcuts", Label: "Raccourcis Clavier", Route: "help/shortcuts"},
				{ID: "about", Label: "À Propos", Route: "help/about"},
			},
		},
	}
}
