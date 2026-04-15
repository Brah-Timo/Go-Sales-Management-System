package models

import "time"

// Category représente une famille de produits (hiérarchique)
type Category struct {
	ID          int    `db:"id"`
	NameFr      string `db:"name_fr"`
	NameAr      string `db:"name_ar"`
	ParentID    *int   `db:"parent_id"`
	Description string `db:"description"`
	Children    []*Category `db:"-"`
}

// Brand représente une marque commerciale
type Brand struct {
	ID           int    `db:"id"`
	Name         string `db:"name"`
	Country      string `db:"country"`
	ProductCount int    `db:"product_count"`
}

// Unit représente une unité de mesure
type Unit struct {
	ID     int    `db:"id"`
	NameFr string `db:"name_fr"`
	NameAr string `db:"name_ar"`
	Symbol string `db:"symbol"`
}

// UnitConversion représente un facteur de conversion entre unités
type UnitConversion struct {
	ID         int     `db:"id"`
	ArticleID  *int    `db:"article_id"`
	FromUnitID int     `db:"from_unit_id"`
	ToUnitID   int     `db:"to_unit_id"`
	Factor     float64 `db:"factor"`
}

// Article représente un produit/marchandise
type Article struct {
	ID                  int       `db:"id"`
	Reference           string    `db:"reference"`
	Barcode             string    `db:"barcode"`
	NameAr              string    `db:"name_ar"`
	NameFr              string    `db:"name_fr"`
	Description         string    `db:"description"`
	CategoryID          *int      `db:"category_id"`
	BrandID             *int      `db:"brand_id"`
	UnitID              *int      `db:"unit_id"`
	PurchasePrice       float64   `db:"purchase_price"`
	CMUP                float64   `db:"cmup"`
	SalePriceHT         float64   `db:"sale_price_ht"`
	SalePriceTTC        float64   `db:"sale_price_ttc"`
	WholesalePrice      float64   `db:"wholesale_price"`
	SemiWholesalePrice  float64   `db:"semi_wholesale_price"`
	MarginPercent       float64   `db:"margin_percent"`
	TVARate             float64   `db:"tva_rate"`
	StockQty            float64   `db:"stock_qty"`
	StockMin            float64   `db:"stock_min"`
	StockMax            float64   `db:"stock_max"`
	ValuationMethod     string    `db:"valuation_method"`
	WarehouseLocation   string    `db:"warehouse_location"`
	LotTracking         bool      `db:"lot_tracking"`
	ExpiryTracking      bool      `db:"expiry_tracking"`
	ImagePath           string    `db:"image_path"`
	IsActive            bool      `db:"is_active"`
	CreatedAt           time.Time `db:"created_at"`
	UpdatedAt           time.Time `db:"updated_at"`

	// Relations chargées dynamiquement
	CategoryName string  `db:"category_name"`
	BrandName    string  `db:"brand_name"`
	UnitSymbol   string  `db:"unit_symbol"`
	UnitNameFr   string  `db:"unit_name_fr"`
}

// ArticleLot représente un lot/batch d'un produit
type ArticleLot struct {
	ID             int     `db:"id"`
	ArticleID      int     `db:"article_id"`
	LotNumber      string  `db:"lot_number"`
	ProductionDate string  `db:"production_date"`
	ExpiryDate     string  `db:"expiry_date"`
	Quantity       float64 `db:"quantity"`
}

// PriceList représente une liste de prix
type PriceList struct {
	ID          int    `db:"id"`
	Name        string `db:"name"`
	Description string `db:"description"`
}

// ArticlePrice représente le prix d'un article dans une liste de prix
type ArticlePrice struct {
	ID          int     `db:"id"`
	ArticleID   int     `db:"article_id"`
	PriceListID int     `db:"price_list_id"`
	Price       float64 `db:"price"`
}

// Warehouse représente un dépôt/magasin
type Warehouse struct {
	ID           int     `db:"id"`
	Name         string  `db:"name"`
	Address      string  `db:"address"`
	Manager      string  `db:"manager"`
	ProductCount int     `db:"product_count"`
	TotalValue   float64 `db:"total_value"`
}

// WarehouseStock représente le stock par entrepôt
type WarehouseStock struct {
	ID          int     `db:"id"`
	ArticleID   int     `db:"article_id"`
	WarehouseID int     `db:"warehouse_id"`
	Quantity    float64 `db:"quantity"`
}

// StockMovement représente un mouvement de stock
type StockMovement struct {
	ID             int     `db:"id"`
	Date           string  `db:"date"`
	Type           string  `db:"type"`
	ArticleID      int     `db:"article_id"`
	ArticleName    string  `db:"article_name"`
	WarehouseID    *int    `db:"warehouse_id"`
	WarehouseName  string  `db:"warehouse_name"`
	Quantity       float64 `db:"quantity"`
	UnitPrice      float64 `db:"unit_price"`
	ReferenceDocID *int    `db:"reference_doc_id"`
	RefDocNumber   string  `db:"ref_doc_number"`
	Notes          string  `db:"notes"`
	CreatedBy      *int    `db:"created_by"`
}

// StockMovement libellés
var StockMoveLabels = map[string]string{
	StockMovePurchaseIn:  "Entrée Achat",
	StockMoveSaleOut:     "Sortie Vente",
	StockMoveTransferIn:  "Entrée Transfert",
	StockMoveTransferOut: "Sortie Transfert",
	StockMoveReturnIn:    "Retour Client",
	StockMoveReturnOut:   "Retour Fournisseur",
	StockMoveAdjustIn:    "Ajustement Positif",
	StockMoveAdjustOut:   "Ajustement Négatif",
	StockMoveDamage:      "Perte/Dommage",
}

// Inventory représente une opération d'inventaire
type Inventory struct {
	ID        int    `db:"id"`
	Date      string `db:"date"`
	Type      string `db:"type"`
	Status    string `db:"status"`
	Notes     string `db:"notes"`
	CreatedBy *int   `db:"created_by"`
	CreatedAt string `db:"created_at"`
	Lines     []InventoryLine `db:"-"`
}

// InventoryLine représente une ligne d'inventaire
type InventoryLine struct {
	ID             int     `db:"id"`
	InventoryID    int     `db:"inventory_id"`
	ArticleID      int     `db:"article_id"`
	ArticleName    string  `db:"article_name"`
	Reference      string  `db:"reference"`
	Barcode        string  `db:"barcode"`
	TheoreticalQty float64 `db:"theoretical_qty"`
	PhysicalQty    float64 `db:"physical_qty"`
	Difference     float64 `db:"difference"`
	Value          float64 `db:"value"`
	Note           string  `db:"note"`
}
