package queries

import (
	"database/sql"
	"fmt"
	"gestion-commerciale/internal/models"
	"strings"
)

// GetAllArticles retourne tous les articles actifs avec leurs relations
func GetAllArticles(db *sql.DB, search string, categoryID int, statusFilter string) ([]models.Article, error) {
	query := `
		SELECT a.id, a.reference, a.barcode, a.name_ar, a.name_fr, a.description,
		       a.category_id, a.brand_id, a.unit_id,
		       a.purchase_price, a.cmup, a.sale_price_ht, a.sale_price_ttc,
		       a.wholesale_price, a.semi_wholesale_price, a.margin_percent,
		       a.tva_rate, a.stock_qty, a.stock_min, a.stock_max,
		       a.valuation_method, a.warehouse_location, a.lot_tracking,
		       a.expiry_tracking, a.image_path, a.is_active,
		       COALESCE(c.name_fr, '') as category_name,
		       COALESCE(b.name, '') as brand_name,
		       COALESCE(u.symbol, '') as unit_symbol,
		       COALESCE(u.name_fr, '') as unit_name_fr
		FROM articles a
		LEFT JOIN categories c ON a.category_id = c.id
		LEFT JOIN brands b     ON a.brand_id = b.id
		LEFT JOIN units u      ON a.unit_id = u.id
		WHERE a.is_active = 1`

	var args []interface{}

	if search != "" {
		query += ` AND (a.name_fr LIKE ? OR a.name_ar LIKE ? OR a.reference LIKE ? OR a.barcode LIKE ?)`
		s := "%" + search + "%"
		args = append(args, s, s, s, s)
	}

	if categoryID > 0 {
		query += ` AND a.category_id = ?`
		args = append(args, categoryID)
	}

	switch statusFilter {
	case "available":
		query += ` AND a.stock_qty > a.stock_min`
	case "out_of_stock":
		query += ` AND a.stock_qty <= 0`
	case "low_stock":
		query += ` AND a.stock_qty <= a.stock_min AND a.stock_qty > 0`
	}

	query += ` ORDER BY a.name_fr`

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("GetAllArticles: %w", err)
	}
	defer rows.Close()

	var articles []models.Article
	for rows.Next() {
		var a models.Article
		err := rows.Scan(
			&a.ID, &a.Reference, &a.Barcode, &a.NameAr, &a.NameFr, &a.Description,
			&a.CategoryID, &a.BrandID, &a.UnitID,
			&a.PurchasePrice, &a.CMUP, &a.SalePriceHT, &a.SalePriceTTC,
			&a.WholesalePrice, &a.SemiWholesalePrice, &a.MarginPercent,
			&a.TVARate, &a.StockQty, &a.StockMin, &a.StockMax,
			&a.ValuationMethod, &a.WarehouseLocation, &a.LotTracking,
			&a.ExpiryTracking, &a.ImagePath, &a.IsActive,
			&a.CategoryName, &a.BrandName, &a.UnitSymbol, &a.UnitNameFr,
		)
		if err != nil {
			return nil, fmt.Errorf("GetAllArticles scan: %w", err)
		}
		articles = append(articles, a)
	}
	return articles, nil
}

// GetArticleByID retourne un article par son ID
func GetArticleByID(db *sql.DB, id int) (*models.Article, error) {
	query := `
		SELECT a.id, a.reference, a.barcode, a.name_ar, a.name_fr, a.description,
		       a.category_id, a.brand_id, a.unit_id,
		       a.purchase_price, a.cmup, a.sale_price_ht, a.sale_price_ttc,
		       a.wholesale_price, a.semi_wholesale_price, a.margin_percent,
		       a.tva_rate, a.stock_qty, a.stock_min, a.stock_max,
		       a.valuation_method, a.warehouse_location, a.lot_tracking,
		       a.expiry_tracking, a.image_path, a.is_active,
		       COALESCE(c.name_fr,'') as category_name,
		       COALESCE(b.name,'') as brand_name,
		       COALESCE(u.symbol,'') as unit_symbol,
		       COALESCE(u.name_fr,'') as unit_name_fr
		FROM articles a
		LEFT JOIN categories c ON a.category_id = c.id
		LEFT JOIN brands b     ON a.brand_id = b.id
		LEFT JOIN units u      ON a.unit_id = u.id
		WHERE a.id = ?`

	var a models.Article
	err := db.QueryRow(query, id).Scan(
		&a.ID, &a.Reference, &a.Barcode, &a.NameAr, &a.NameFr, &a.Description,
		&a.CategoryID, &a.BrandID, &a.UnitID,
		&a.PurchasePrice, &a.CMUP, &a.SalePriceHT, &a.SalePriceTTC,
		&a.WholesalePrice, &a.SemiWholesalePrice, &a.MarginPercent,
		&a.TVARate, &a.StockQty, &a.StockMin, &a.StockMax,
		&a.ValuationMethod, &a.WarehouseLocation, &a.LotTracking,
		&a.ExpiryTracking, &a.ImagePath, &a.IsActive,
		&a.CategoryName, &a.BrandName, &a.UnitSymbol, &a.UnitNameFr,
	)
	if err != nil {
		return nil, fmt.Errorf("GetArticleByID(%d): %w", id, err)
	}
	return &a, nil
}

// FindArticleByBarcode recherche un article par code-barres ou référence
func FindArticleByBarcode(db *sql.DB, code string) (*models.Article, error) {
	code = strings.TrimSpace(code)
	if code == "" {
		return nil, nil
	}

	query := `
		SELECT a.id, a.reference, a.barcode, a.name_ar, a.name_fr,
		       a.sale_price_ht, a.sale_price_ttc, a.purchase_price, a.cmup,
		       a.tva_rate, a.stock_qty, a.margin_percent,
		       COALESCE(u.symbol,'') as unit_symbol,
		       COALESCE(u.name_fr,'') as unit_name_fr,
		       a.category_id, a.unit_id, a.brand_id, a.image_path,
		       a.wholesale_price, a.semi_wholesale_price,
		       a.stock_min, a.stock_max, a.is_active
		FROM articles a
		LEFT JOIN units u ON a.unit_id = u.id
		WHERE (a.barcode = ? OR a.reference = ?) AND a.is_active = 1
		LIMIT 1`

	var a models.Article
	err := db.QueryRow(query, code, code).Scan(
		&a.ID, &a.Reference, &a.Barcode, &a.NameAr, &a.NameFr,
		&a.SalePriceHT, &a.SalePriceTTC, &a.PurchasePrice, &a.CMUP,
		&a.TVARate, &a.StockQty, &a.MarginPercent,
		&a.UnitSymbol, &a.UnitNameFr,
		&a.CategoryID, &a.UnitID, &a.BrandID, &a.ImagePath,
		&a.WholesalePrice, &a.SemiWholesalePrice,
		&a.StockMin, &a.StockMax, &a.IsActive,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &a, nil
}

// SaveArticle insère ou met à jour un article
func SaveArticle(db *sql.DB, a *models.Article) error {
	if a.ID == 0 {
		// INSERT
		result, err := db.Exec(`
			INSERT INTO articles (reference, barcode, name_ar, name_fr, description,
			  category_id, brand_id, unit_id, purchase_price, cmup,
			  sale_price_ht, sale_price_ttc, wholesale_price, semi_wholesale_price,
			  margin_percent, tva_rate, stock_qty, stock_min, stock_max,
			  valuation_method, warehouse_location, lot_tracking, expiry_tracking,
			  image_path, is_active)
			VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`,
			a.Reference, a.Barcode, a.NameAr, a.NameFr, a.Description,
			a.CategoryID, a.BrandID, a.UnitID, a.PurchasePrice, a.CMUP,
			a.SalePriceHT, a.SalePriceTTC, a.WholesalePrice, a.SemiWholesalePrice,
			a.MarginPercent, a.TVARate, a.StockQty, a.StockMin, a.StockMax,
			a.ValuationMethod, a.WarehouseLocation, a.LotTracking, a.ExpiryTracking,
			a.ImagePath, a.IsActive,
		)
		if err != nil {
			return fmt.Errorf("SaveArticle insert: %w", err)
		}
		id, _ := result.LastInsertId()
		a.ID = int(id)
	} else {
		// UPDATE
		_, err := db.Exec(`
			UPDATE articles SET reference=?, barcode=?, name_ar=?, name_fr=?,
			  description=?, category_id=?, brand_id=?, unit_id=?,
			  purchase_price=?, sale_price_ht=?, sale_price_ttc=?,
			  wholesale_price=?, semi_wholesale_price=?,
			  margin_percent=?, tva_rate=?, stock_min=?, stock_max=?,
			  valuation_method=?, warehouse_location=?, lot_tracking=?,
			  expiry_tracking=?, image_path=?, is_active=?
			WHERE id=?`,
			a.Reference, a.Barcode, a.NameAr, a.NameFr,
			a.Description, a.CategoryID, a.BrandID, a.UnitID,
			a.PurchasePrice, a.SalePriceHT, a.SalePriceTTC,
			a.WholesalePrice, a.SemiWholesalePrice,
			a.MarginPercent, a.TVARate, a.StockMin, a.StockMax,
			a.ValuationMethod, a.WarehouseLocation, a.LotTracking,
			a.ExpiryTracking, a.ImagePath, a.IsActive,
			a.ID,
		)
		if err != nil {
			return fmt.Errorf("SaveArticle update: %w", err)
		}
	}
	return nil
}

// DeleteArticle supprime un article (vérifie s'il a des mouvements)
func DeleteArticle(db *sql.DB, id int) error {
	var count int
	db.QueryRow(`SELECT COUNT(*) FROM document_lines WHERE article_id=?`, id).Scan(&count)
	if count > 0 {
		return fmt.Errorf("impossible de supprimer: cet article a %d lignes de document associées", count)
	}

	db.QueryRow(`SELECT COUNT(*) FROM stock_movements WHERE article_id=?`, id).Scan(&count)
	if count > 0 {
		return fmt.Errorf("impossible de supprimer: cet article a %d mouvements de stock associés", count)
	}

	_, err := db.Exec(`DELETE FROM articles WHERE id=?`, id)
	return err
}

// GetLowStockArticles retourne les articles sous le seuil minimum
func GetLowStockArticles(db *sql.DB) ([]models.Article, error) {
	query := `
		SELECT a.id, a.reference, a.name_fr, a.stock_qty, a.stock_min, a.stock_max,
		       COALESCE(u.symbol,'') as unit_symbol
		FROM articles a
		LEFT JOIN units u ON a.unit_id = u.id
		WHERE a.is_active=1 AND a.stock_qty <= a.stock_min
		ORDER BY (a.stock_qty - a.stock_min) ASC`

	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var articles []models.Article
	for rows.Next() {
		var a models.Article
		rows.Scan(&a.ID, &a.Reference, &a.NameFr, &a.StockQty, &a.StockMin, &a.StockMax, &a.UnitSymbol)
		articles = append(articles, a)
	}
	return articles, nil
}

// GetArticleStockMovements retourne les mouvements de stock d'un article
func GetArticleStockMovements(db *sql.DB, articleID int) ([]models.StockMovement, error) {
	rows, err := db.Query(`
		SELECT sm.id, sm.date, sm.type, sm.article_id,
		       COALESCE(a.name_fr,'') as article_name,
		       sm.warehouse_id, COALESCE(w.name,'') as warehouse_name,
		       sm.quantity, sm.unit_price, sm.reference_doc_id,
		       COALESCE(d.doc_number,'') as ref_doc_number,
		       sm.notes
		FROM stock_movements sm
		JOIN articles a ON sm.article_id = a.id
		LEFT JOIN warehouses w ON sm.warehouse_id = w.id
		LEFT JOIN documents d ON sm.reference_doc_id = d.id
		WHERE sm.article_id = ?
		ORDER BY sm.date DESC, sm.id DESC`, articleID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var moves []models.StockMovement
	for rows.Next() {
		var m models.StockMovement
		rows.Scan(&m.ID, &m.Date, &m.Type, &m.ArticleID, &m.ArticleName,
			&m.WarehouseID, &m.WarehouseName, &m.Quantity, &m.UnitPrice,
			&m.ReferenceDocID, &m.RefDocNumber, &m.Notes)
		moves = append(moves, m)
	}
	return moves, nil
}

// GetAllCategories retourne toutes les catégories
func GetAllCategories(db *sql.DB) ([]models.Category, error) {
	rows, err := db.Query(`SELECT id, name_fr, name_ar, parent_id, description FROM categories ORDER BY parent_id NULLS FIRST, name_fr`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cats []models.Category
	for rows.Next() {
		var c models.Category
		rows.Scan(&c.ID, &c.NameFr, &c.NameAr, &c.ParentID, &c.Description)
		cats = append(cats, c)
	}
	return cats, nil
}

// GetAllUnits retourne toutes les unités
func GetAllUnits(db *sql.DB) ([]models.Unit, error) {
	rows, err := db.Query(`SELECT id, name_fr, name_ar, symbol FROM units ORDER BY name_fr`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var units []models.Unit
	for rows.Next() {
		var u models.Unit
		rows.Scan(&u.ID, &u.NameFr, &u.NameAr, &u.Symbol)
		units = append(units, u)
	}
	return units, nil
}

// GetAllBrands retourne toutes les marques
func GetAllBrands(db *sql.DB) ([]models.Brand, error) {
	rows, err := db.Query(`
		SELECT b.id, b.name, b.country, COUNT(a.id) as product_count
		FROM brands b LEFT JOIN articles a ON a.brand_id = b.id
		GROUP BY b.id ORDER BY b.name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var brands []models.Brand
	for rows.Next() {
		var b models.Brand
		rows.Scan(&b.ID, &b.Name, &b.Country, &b.ProductCount)
		brands = append(brands, b)
	}
	return brands, nil
}

// GetAllWarehouses retourne tous les dépôts
func GetAllWarehouses(db *sql.DB) ([]models.Warehouse, error) {
	rows, err := db.Query(`
		SELECT w.id, w.name, w.address, w.manager,
		  COUNT(DISTINCT ws.article_id) as product_count,
		  COALESCE(SUM(ws.quantity * a.cmup), 0) as total_value
		FROM warehouses w
		LEFT JOIN warehouse_stock ws ON ws.warehouse_id = w.id
		LEFT JOIN articles a ON ws.article_id = a.id
		GROUP BY w.id ORDER BY w.name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var whs []models.Warehouse
	for rows.Next() {
		var w models.Warehouse
		rows.Scan(&w.ID, &w.Name, &w.Address, &w.Manager, &w.ProductCount, &w.TotalValue)
		whs = append(whs, w)
	}
	return whs, nil
}

// GetAllPriceLists retourne toutes les listes de prix
func GetAllPriceLists(db *sql.DB) ([]models.PriceList, error) {
	rows, err := db.Query(`SELECT id, name, description FROM price_lists ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var lists []models.PriceList
	for rows.Next() {
		var p models.PriceList
		rows.Scan(&p.ID, &p.Name, &p.Description)
		lists = append(lists, p)
	}
	return lists, nil
}

// SearchArticles recherche des articles (alias de GetAllArticles)
func SearchArticles(db *sql.DB, query string, categoryID int, activeOnly bool) ([]models.Article, error) {
	return GetAllArticles(db, query, categoryID, "")
}

// NextArticleReference génère la prochaine référence d'article
func NextArticleReference(db *sql.DB) string {
	var maxRef string
	db.QueryRow(`SELECT COALESCE(MAX(reference),'ART-0000') FROM articles WHERE reference LIKE 'ART-%'`).Scan(&maxRef)
	// Parse et incrémenter
	num := 1
	if len(maxRef) > 4 {
		fmt.Sscanf(maxRef[4:], "%d", &num)
		num++
	}
	return fmt.Sprintf("ART-%04d", num)
}
