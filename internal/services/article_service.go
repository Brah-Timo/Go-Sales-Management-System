package services

import (
	"database/sql"
	"fmt"
	"gestion-commerciale/internal/database/queries"
	"gestion-commerciale/internal/models"
	"gestion-commerciale/pkg/utils"
	"math"
	"strings"
)

// ArticleService gère les opérations sur les articles
type ArticleService struct {
	db *sql.DB
}

// NewArticleService crée un service d'articles
func NewArticleService(db *sql.DB) *ArticleService {
	return &ArticleService{db: db}
}

// Search recherche des articles par texte libre
func (s *ArticleService) Search(query string, categoryID int, activeOnly bool) ([]models.Article, error) {
	return queries.SearchArticles(s.db, query, categoryID, activeOnly)
}

// GetByID retourne un article par son ID
func (s *ArticleService) GetByID(id int) (*models.Article, error) {
	return queries.GetArticleByID(s.db, id)
}

// GetByBarcode retourne un article par son code-barres
func (s *ArticleService) GetByBarcode(barcode string) (*models.Article, error) {
	var a models.Article
	err := s.db.QueryRow(`
		SELECT a.id, a.reference, a.barcode, a.name_ar, a.name_fr, a.description,
		  a.category_id, a.brand_id, a.unit_id, a.purchase_price, a.cmup,
		  a.sale_price_ht, a.sale_price_ttc, a.wholesale_price, a.semi_wholesale_price,
		  a.margin_percent, a.tva_rate, a.stock_qty, a.stock_min, a.stock_max,
		  a.valuation_method, a.warehouse_location, a.lot_tracking, a.expiry_tracking,
		  a.image_path, a.is_active,
		  COALESCE(c.name_fr,'') as category_name,
		  COALESCE(b.name,'') as brand_name,
		  COALESCE(u.symbol,'') as unit_symbol,
		  COALESCE(u.name_fr,'') as unit_name_fr
		FROM articles a
		LEFT JOIN categories c ON a.category_id = c.id
		LEFT JOIN brands b ON a.brand_id = b.id
		LEFT JOIN units u ON a.unit_id = u.id
		WHERE (a.barcode=? OR a.reference=?) AND a.is_active=1`, barcode, barcode).
		Scan(&a.ID, &a.Reference, &a.Barcode, &a.NameAr, &a.NameFr, &a.Description,
			&a.CategoryID, &a.BrandID, &a.UnitID, &a.PurchasePrice, &a.CMUP,
			&a.SalePriceHT, &a.SalePriceTTC, &a.WholesalePrice, &a.SemiWholesalePrice,
			&a.MarginPercent, &a.TVARate, &a.StockQty, &a.StockMin, &a.StockMax,
			&a.ValuationMethod, &a.WarehouseLocation, &a.LotTracking, &a.ExpiryTracking,
			&a.ImagePath, &a.IsActive,
			&a.CategoryName, &a.BrandName, &a.UnitSymbol, &a.UnitNameFr)
	if err != nil {
		return nil, fmt.Errorf("article non trouvé: %w", err)
	}
	return &a, nil
}

// Save crée ou met à jour un article
func (s *ArticleService) Save(a *models.Article, userID int) error {
	// Calculer la marge si possible
	if a.PurchasePrice > 0 && a.SalePriceHT > 0 {
		a.MarginPercent = utils.CalculateMargin(a.PurchasePrice, a.SalePriceHT)
	}
	// Calculer TTC si HT défini
	if a.SalePriceHT > 0 {
		a.SalePriceTTC = utils.HTToTTC(a.SalePriceHT, a.TVARate)
	}

	if a.ID == 0 {
		// Création
		res, err := s.db.Exec(`
			INSERT INTO articles (reference, barcode, name_ar, name_fr, description,
			  category_id, brand_id, unit_id, purchase_price, cmup,
			  sale_price_ht, sale_price_ttc, wholesale_price, semi_wholesale_price,
			  margin_percent, tva_rate, stock_qty, stock_min, stock_max,
			  valuation_method, warehouse_location, lot_tracking, expiry_tracking,
			  image_path, is_active)
			VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,1)`,
			a.Reference, a.Barcode, a.NameAr, a.NameFr, a.Description,
			a.CategoryID, a.BrandID, a.UnitID, a.PurchasePrice, a.CMUP,
			a.SalePriceHT, a.SalePriceTTC, a.WholesalePrice, a.SemiWholesalePrice,
			a.MarginPercent, a.TVARate, a.StockQty, a.StockMin, a.StockMax,
			a.ValuationMethod, a.WarehouseLocation, a.LotTracking, a.ExpiryTracking,
			a.ImagePath)
		if err != nil {
			return fmt.Errorf("erreur création article: %w", err)
		}
		id, _ := res.LastInsertId()
		a.ID = int(id)

		// Journal d'audit
		s.db.Exec(`INSERT INTO audit_log (user_id, action_type, module, description) VALUES (?,?,?,?)`,
			userID, "create", "articles", "Création article "+a.Reference)
	} else {
		// Mise à jour
		_, err := s.db.Exec(`
			UPDATE articles SET reference=?, barcode=?, name_ar=?, name_fr=?, description=?,
			  category_id=?, brand_id=?, unit_id=?, purchase_price=?,
			  sale_price_ht=?, sale_price_ttc=?, wholesale_price=?, semi_wholesale_price=?,
			  margin_percent=?, tva_rate=?, stock_min=?, stock_max=?,
			  valuation_method=?, warehouse_location=?, lot_tracking=?, expiry_tracking=?,
			  image_path=?
			WHERE id=?`,
			a.Reference, a.Barcode, a.NameAr, a.NameFr, a.Description,
			a.CategoryID, a.BrandID, a.UnitID, a.PurchasePrice,
			a.SalePriceHT, a.SalePriceTTC, a.WholesalePrice, a.SemiWholesalePrice,
			a.MarginPercent, a.TVARate, a.StockMin, a.StockMax,
			a.ValuationMethod, a.WarehouseLocation, a.LotTracking, a.ExpiryTracking,
			a.ImagePath, a.ID)
		if err != nil {
			return fmt.Errorf("erreur mise à jour article: %w", err)
		}

		s.db.Exec(`INSERT INTO audit_log (user_id, action_type, module, description) VALUES (?,?,?,?)`,
			userID, "update", "articles", "Modification article "+a.Reference)
	}
	return nil
}

// Delete désactive un article (suppression logique)
func (s *ArticleService) Delete(articleID, userID int) error {
	// Vérifier s'il a des mouvements
	var count int
	s.db.QueryRow(`SELECT COUNT(*) FROM document_lines WHERE article_id=?`, articleID).Scan(&count)
	if count > 0 {
		return fmt.Errorf("impossible de supprimer: l'article a %d ligne(s) dans des documents", count)
	}

	_, err := s.db.Exec(`UPDATE articles SET is_active=0 WHERE id=?`, articleID)
	if err != nil {
		return err
	}
	s.db.Exec(`INSERT INTO audit_log (user_id, action_type, module, description) VALUES (?,?,?,?)`,
		userID, "delete", "articles", fmt.Sprintf("Suppression article ID %d", articleID))
	return nil
}

// GetCategories retourne toutes les catégories
func (s *ArticleService) GetCategories() ([]models.Category, error) {
	rows, err := s.db.Query(`SELECT id, name_fr, name_ar, COALESCE(parent_id, 0) as parent_id, description FROM categories ORDER BY parent_id, name_fr`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var cats []models.Category
	for rows.Next() {
		var c models.Category
		var parentID int
		rows.Scan(&c.ID, &c.NameFr, &c.NameAr, &parentID, &c.Description)
		if parentID > 0 {
			c.ParentID = &parentID
		}
		cats = append(cats, c)
	}
	return cats, nil
}

// GetBrands retourne toutes les marques
func (s *ArticleService) GetBrands() ([]models.Brand, error) {
	rows, err := s.db.Query(`
		SELECT b.id, b.name, b.country, COUNT(a.id) as product_count
		FROM brands b LEFT JOIN articles a ON a.brand_id=b.id AND a.is_active=1
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

// GetUnits retourne toutes les unités
func (s *ArticleService) GetUnits() ([]models.Unit, error) {
	rows, err := s.db.Query(`SELECT id, name_fr, name_ar, symbol FROM units ORDER BY name_fr`)
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

// GetWarehouses retourne tous les dépôts
func (s *ArticleService) GetWarehouses() ([]models.Warehouse, error) {
	rows, err := s.db.Query(`
		SELECT w.id, w.name, w.address, w.manager,
		  COUNT(DISTINCT ws.article_id) as product_count,
		  COALESCE(SUM(ws.quantity * a.cmup), 0) as total_value
		FROM warehouses w
		LEFT JOIN warehouse_stock ws ON ws.warehouse_id=w.id
		LEFT JOIN articles a ON ws.article_id=a.id
		GROUP BY w.id ORDER BY w.name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var warehouses []models.Warehouse
	for rows.Next() {
		var w models.Warehouse
		rows.Scan(&w.ID, &w.Name, &w.Address, &w.Manager, &w.ProductCount, &w.TotalValue)
		warehouses = append(warehouses, w)
	}
	return warehouses, nil
}

// GetLowStockArticles retourne les articles sous le seuil minimum
func (s *ArticleService) GetLowStockArticles() ([]models.Article, error) {
	rows, err := s.db.Query(`
		SELECT a.id, a.reference, a.name_fr, a.stock_qty, a.stock_min, a.stock_max
		FROM articles a
		WHERE a.is_active=1 AND a.stock_qty <= a.stock_min
		ORDER BY (a.stock_min - a.stock_qty) DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var articles []models.Article
	for rows.Next() {
		var a models.Article
		rows.Scan(&a.ID, &a.Reference, &a.NameFr, &a.StockQty, &a.StockMin, &a.StockMax)
		articles = append(articles, a)
	}
	return articles, nil
}

// BulkUpdatePrices met à jour les prix en masse
func (s *ArticleService) BulkUpdatePrices(categoryID int, updateType string, value float64, userID int) error {
	query := `SELECT id, sale_price_ht, sale_price_ttc, tva_rate, cmup, margin_percent FROM articles WHERE is_active=1`
	args := []interface{}{}
	if categoryID > 0 {
		query += ` AND category_id=?`
		args = append(args, categoryID)
	}

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return err
	}
	defer rows.Close()

	tx, _ := s.db.Begin()
	defer tx.Rollback()

	for rows.Next() {
		var id int
		var ht, ttc, tvaRate, cmup, marginPct float64
		rows.Scan(&id, &ht, &ttc, &tvaRate, &cmup, &marginPct)

		var newHT float64
		switch updateType {
		case "increase_percent":
			newHT = math.Round(ht*(1+value/100)*100) / 100
		case "increase_amount":
			newHT = math.Round((ht+value)*100) / 100
		case "decrease_percent":
			newHT = math.Round(ht*(1-value/100)*100) / 100
		case "from_cmup":
			if cmup > 0 {
				newHT = math.Round(cmup*(1+value/100)*100) / 100
			} else {
				continue
			}
		case "set_margin":
			if cmup > 0 {
				newHT = math.Round(cmup*(1+value/100)*100) / 100
			} else {
				continue
			}
		default:
			continue
		}

		newTTC := math.Round(newHT*(1+tvaRate/100)*100) / 100
		newMargin := utils.CalculateMargin(cmup, newHT)
		tx.Exec(`UPDATE articles SET sale_price_ht=?, sale_price_ttc=?, margin_percent=? WHERE id=?`,
			newHT, newTTC, newMargin, id)
	}

	s.db.Exec(`INSERT INTO audit_log (user_id, action_type, module, description) VALUES (?,?,?,?)`,
		userID, "update", "articles", fmt.Sprintf("Mise à jour prix en masse: %s %.2f", updateType, value))

	return tx.Commit()
}

// GenerateReference génère une référence unique
func (s *ArticleService) GenerateReference() string {
	var maxRef string
	s.db.QueryRow(`SELECT COALESCE(MAX(reference),'ART-0000') FROM articles WHERE reference LIKE 'ART-%'`).Scan(&maxRef)

	var num int
	fmt.Sscanf(strings.TrimPrefix(maxRef, "ART-"), "%d", &num)
	num++
	return fmt.Sprintf("ART-%04d", num)
}

// GetPriceLists retourne toutes les listes de prix
func (s *ArticleService) GetPriceLists() ([]models.PriceList, error) {
	rows, err := s.db.Query(`SELECT id, name, description FROM price_lists ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var pls []models.PriceList
	for rows.Next() {
		var pl models.PriceList
		rows.Scan(&pl.ID, &pl.Name, &pl.Description)
		pls = append(pls, pl)
	}
	return pls, nil
}

// SaveCategory sauvegarde une catégorie
func (s *ArticleService) SaveCategory(c *models.Category) error {
	if c.ID == 0 {
		_, err := s.db.Exec(`INSERT INTO categories (name_fr, name_ar, parent_id, description) VALUES (?,?,?,?)`,
			c.NameFr, c.NameAr, c.ParentID, c.Description)
		return err
	}
	_, err := s.db.Exec(`UPDATE categories SET name_fr=?, name_ar=?, parent_id=?, description=? WHERE id=?`,
		c.NameFr, c.NameAr, c.ParentID, c.Description, c.ID)
	return err
}

// SaveBrand sauvegarde une marque
func (s *ArticleService) SaveBrand(b *models.Brand) error {
	if b.ID == 0 {
		_, err := s.db.Exec(`INSERT INTO brands (name, country) VALUES (?,?)`, b.Name, b.Country)
		return err
	}
	_, err := s.db.Exec(`UPDATE brands SET name=?, country=? WHERE id=?`, b.Name, b.Country, b.ID)
	return err
}

// SaveUnit sauvegarde une unité
func (s *ArticleService) SaveUnit(u *models.Unit) error {
	if u.ID == 0 {
		_, err := s.db.Exec(`INSERT INTO units (name_fr, name_ar, symbol) VALUES (?,?,?)`, u.NameFr, u.NameAr, u.Symbol)
		return err
	}
	_, err := s.db.Exec(`UPDATE units SET name_fr=?, name_ar=?, symbol=? WHERE id=?`, u.NameFr, u.NameAr, u.Symbol, u.ID)
	return err
}
