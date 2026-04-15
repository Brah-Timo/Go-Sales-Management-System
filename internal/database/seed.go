package database

import "database/sql"

// SeedDatabase insère les données de base si elles n'existent pas encore
func SeedDatabase(db *sql.DB) error {
	// Unités de mesure
	units := []struct{ fr, ar, sym string }{
		{"Pièce", "قطعة", "U"},
		{"Kilogramme", "كيلوغرام", "Kg"},
		{"Gramme", "غرام", "g"},
		{"Litre", "لتر", "L"},
		{"Mètre", "متر", "m"},
		{"Mètre carré", "متر مربع", "m²"},
		{"Boîte", "علبة", "Bte"},
		{"Carton", "كرتون", "Crt"},
		{"Sac", "كيس", "Sac"},
		{"Tonne", "طن", "T"},
		{"Lot", "باقة", "Lot"},
	}
	for _, u := range units {
		db.Exec(`INSERT OR IGNORE INTO units (name_fr, name_ar, symbol) VALUES (?,?,?)`,
			u.fr, u.ar, u.sym)
	}

	// Catégories principales
	mainCats := []struct{ fr, ar string }{
		{"Alimentation", "مواد غذائية"},
		{"Boissons", "مشروبات"},
		{"Produits d'entretien", "مواد التنظيف"},
		{"Articles ménagers", "مستلزمات منزلية"},
		{"Fournitures de bureau", "أدوات مكتبية"},
		{"Matériaux de construction", "مواد بناء"},
		{"Pièces de rechange", "قطع غيار"},
		{"Vêtements", "ألبسة"},
		{"Électronique", "أجهزة إلكترونية"},
		{"Autres", "أخرى"},
	}
	for _, c := range mainCats {
		db.Exec(`INSERT OR IGNORE INTO categories (name_fr, name_ar, parent_id) VALUES (?,?,NULL)`,
			c.fr, c.ar)
	}

	// Sous-catégories alimentation (parent_id=1)
	subCats := []struct{ fr, ar string }{
		{"Céréales et pâtes", "حبوب وعجائن"},
		{"Huiles et beurre", "زيوت وسمن"},
		{"Sucre et dérivés", "سكر ومشتقاته"},
		{"Lait et produits laitiers", "حليب ومشتقاته"},
		{"Conserves", "معلبات"},
		{"Eaux minérales", "مياه معدنية"},
		{"Jus", "عصائر"},
		{"Boissons gazeuses", "مشروبات غازية"},
	}
	var alimID, boissonID int
	db.QueryRow(`SELECT id FROM categories WHERE name_fr='Alimentation'`).Scan(&alimID)
	db.QueryRow(`SELECT id FROM categories WHERE name_fr='Boissons'`).Scan(&boissonID)
	for i, sc := range subCats {
		parentID := alimID
		if i >= 5 { parentID = boissonID }
		db.Exec(`INSERT OR IGNORE INTO categories (name_fr, name_ar, parent_id) VALUES (?,?,?)`,
			sc.fr, sc.ar, parentID)
	}

	// Configuration de numérotation
	configs := []struct{ docType, prefix string }{
		{"FA", "FA"}, {"FAC", "FAC"}, {"BL", "BL"}, {"BR", "BR"},
		{"DV", "DV"}, {"PF", "PF"}, {"BCC", "BCC"}, {"BCF", "BCF"},
		{"AV", "AV"}, {"BP", "BP"}, {"BRE", "BRE"},
	}
	for _, c := range configs {
		db.Exec(`INSERT OR IGNORE INTO numbering_config (doc_type, prefix, current_number, reset_yearly) VALUES (?,?,0,1)`,
			c.docType, c.prefix)
	}

	// Listes de prix
	priceLists := []struct{ name, desc string }{
		{"Détail", "Prix de détail standard"},
		{"Demi-gros", "Prix demi-gros"},
		{"Gros", "Prix de gros"},
		{"Spécial 1", "Prix spécial catégorie 1"},
		{"Spécial 2", "Prix promotionnel"},
	}
	for _, p := range priceLists {
		db.Exec(`INSERT OR IGNORE INTO price_lists (name, description) VALUES (?,?)`, p.name, p.desc)
	}

	// Devises
	currencies := []struct{ name, code, symbol string; rate float64 }{
		{"Dinar Algérien", "DZD", "DA", 1.0},
		{"Euro", "EUR", "€", 147.50},
		{"Dollar US", "USD", "$", 135.20},
	}
	for _, c := range currencies {
		db.Exec(`INSERT OR IGNORE INTO currencies (name, code, symbol, rate) VALUES (?,?,?,?)`,
			c.name, c.code, c.symbol, c.rate)
	}

	// Catégories de dépenses
	expenseCats := []string{
		"Carburant", "Transport", "Repas", "Fournitures de bureau",
		"Maintenance", "Loyer", "Électricité & Gaz", "Téléphone & Internet", "Autres",
	}
	for _, ec := range expenseCats {
		db.Exec(`INSERT OR IGNORE INTO expense_categories (name) VALUES (?)`, ec)
	}

	// Dépôt principal
	db.Exec(`INSERT OR IGNORE INTO warehouses (id, name, address) VALUES (1, 'Dépôt Principal', 'Siège Social')`)

	// Paramètres par défaut
	settings := map[string]string{
		"tva_normal":             "19",
		"tva_reduced":            "9",
		"timbre_rate":            "1",
		"timbre_max":             "2500",
		"timbre_exemption":       "0",
		"tap_rate":               "1",
		"is_tva_subject":         "1",
		"tax_regime":             "real",
		"auto_timbre":            "1",
		"auto_update_sale_price": "0",
		"language":               "fr",
		"backup_mode":            "daily",
		"print_config":           `{"paper_size":"A4","orientation":"portrait","include_logo":true,"include_stamp":true,"include_amount_words":true,"invoice_language":"bilingual"}`,
		"barcode_config":         `{"type":"EAN-13","label_size":"38x25","columns":3}`,
		"app_version":            "1.0.0",
	}
	for k, v := range settings {
		db.Exec(`INSERT OR IGNORE INTO settings (key, value) VALUES (?,?)`, k, v)
	}

	// Compte bancaire par défaut
	db.Exec(`INSERT OR IGNORE INTO bank_accounts (id, bank_name, branch) VALUES (1, 'Banque Principale', 'Agence Principale')`)

	// Année fiscale courante
	db.Exec(`INSERT OR IGNORE INTO fiscal_years (year, start_date, end_date, status) 
		VALUES (strftime('%Y','now'), strftime('%Y','now')||'-01-01', strftime('%Y','now')||'-12-31', 'open')`)

	// Informations société par défaut
	db.Exec(`INSERT OR IGNORE INTO companies (id, name_fr, name_ar, activity) 
		VALUES (1, 'Mon Commerce', 'تجارتي', 'Commerce Général')`)

	// Utilisateur admin par défaut (mot de passe: admin123)
	// Hash bcrypt généré avec bcrypt.DefaultCost — vérifié OK
	adminHash := "$2a$10$dLMnHzeDzQf0wiCXuqc/r.3Wv67Pv/Na7Vv6rD1a265wYXt7qiYxm"
	adminPerms := `{"create_sale_invoice":true,"create_purchase_invoice":true,"edit_confirmed_invoice":true,"delete_invoice":true,"edit_prices":true,"view_purchase_prices":true,"view_profit_margin":true,"manage_stock":true,"manage_clients_suppliers":true,"access_financial_reports":true,"collect_payments":true,"manage_settings":true,"backup_restore":true,"apply_discount_above_10":true,"inventory":true}`
	db.Exec(`INSERT OR IGNORE INTO users (id, username, full_name, password_hash, role, permissions_json, is_active)
		VALUES (1, 'admin', 'Administrateur', ?, 'admin', ?, 1)`,
		adminHash, adminPerms)
	// Migration: corriger le hash si l'ancien hash incorrect est présent
	db.Exec(`UPDATE users SET password_hash=? WHERE username='admin' AND password_hash='$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy'`,
		adminHash)

	// Client de passage (client par défaut pour les ventes au comptant)
	db.Exec(`INSERT OR IGNORE INTO clients (id, code, name_fr, name_ar, type) VALUES (0, 'C000', 'Client de Passage', 'زبون عابر', 'person')`)

	return nil
}
