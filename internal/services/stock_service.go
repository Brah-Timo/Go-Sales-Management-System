package services

import (
	"database/sql"
	"fmt"
	"gestion-commerciale/internal/models"
	"gestion-commerciale/pkg/utils"
)

// StockService gère les opérations de stock
type StockService struct {
	db *sql.DB
}

// NewStockService crée un service de stock
func NewStockService(db *sql.DB) *StockService {
	return &StockService{db: db}
}

// AddStockMovement ajoute un mouvement de stock
func (s *StockService) AddStockMovement(move *models.StockMovement, userID int) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.Exec(`
		INSERT INTO stock_movements (date, type, article_id, warehouse_id, quantity, unit_price, reference_doc_id, notes, created_by)
		VALUES (?,?,?,?,?,?,?,?,?)`,
		move.Date, move.Type, move.ArticleID, move.WarehouseID,
		move.Quantity, move.UnitPrice, move.ReferenceDocID, move.Notes, userID)
	if err != nil {
		return fmt.Errorf("AddStockMovement: %w", err)
	}

	// Mettre à jour la quantité en stock
	switch move.Type {
	case models.StockMovePurchaseIn, models.StockMoveReturnIn,
		models.StockMoveAdjustIn, models.StockMoveTransferIn:
		tx.Exec(`UPDATE articles SET stock_qty = stock_qty + ? WHERE id=?`,
			move.Quantity, move.ArticleID)
	case models.StockMoveSaleOut, models.StockMoveReturnOut,
		models.StockMoveAdjustOut, models.StockMoveTransferOut, models.StockMoveDamage:
		tx.Exec(`UPDATE articles SET stock_qty = stock_qty - ? WHERE id=?`,
			move.Quantity, move.ArticleID)
	}

	// Recalculer le CMUP si entrée
	if move.Type == models.StockMovePurchaseIn || move.Type == models.StockMoveReturnIn {
		s.recalculateCMUPInTx(tx, move.ArticleID)
	}

	// Mettre à jour le stock par entrepôt si spécifié
	if move.WarehouseID != nil {
		switch move.Type {
		case models.StockMovePurchaseIn, models.StockMoveReturnIn,
			models.StockMoveAdjustIn, models.StockMoveTransferIn:
			tx.Exec(`INSERT INTO warehouse_stock (article_id, warehouse_id, quantity) VALUES (?,?,?)
				ON CONFLICT(article_id, warehouse_id) DO UPDATE SET quantity = quantity + ?`,
				move.ArticleID, *move.WarehouseID, move.Quantity, move.Quantity)
		default:
			tx.Exec(`INSERT INTO warehouse_stock (article_id, warehouse_id, quantity) VALUES (?,?,0)
				ON CONFLICT(article_id, warehouse_id) DO UPDATE SET quantity = quantity - ?`,
				move.ArticleID, *move.WarehouseID, move.Quantity)
		}
	}

	return tx.Commit()
}

// recalculateCMUPInTx recalcule le CMUP dans une transaction
func (s *StockService) recalculateCMUPInTx(tx *sql.Tx, articleID int) {
	rows, err := tx.Query(`
		SELECT type, quantity, unit_price FROM stock_movements
		WHERE article_id=? ORDER BY date, id`, articleID)
	if err != nil {
		return
	}
	defer rows.Close()

	state := utils.NewCMUPState()
	for rows.Next() {
		var mvType string
		var qty, price float64
		rows.Scan(&mvType, &qty, &price)
		switch mvType {
		case "purchase_in", "adjustment_in", "return_in", "transfer_in":
			state.AddPurchase(qty, price)
		default:
			state.AddSale(qty)
		}
	}
	tx.Exec(`UPDATE articles SET cmup=? WHERE id=?`, state.CMUP(), articleID)
}

// TransferStock transfère du stock entre deux entrepôts
func (s *StockService) TransferStock(articleID, fromWarehouseID, toWarehouseID int, qty float64, userID int) error {
	if fromWarehouseID == toWarehouseID {
		return fmt.Errorf("les entrepôts source et destination sont identiques")
	}

	// Vérifier le stock disponible
	var available float64
	s.db.QueryRow(`SELECT COALESCE(quantity,0) FROM warehouse_stock WHERE article_id=? AND warehouse_id=?`,
		articleID, fromWarehouseID).Scan(&available)
	if available < qty {
		return fmt.Errorf("stock insuffisant: disponible %.2f, demandé %.2f", available, qty)
	}

	now := utils.NowString()

	// Sortie de l'entrepôt source
	outMove := &models.StockMovement{
		Date:        now,
		Type:        models.StockMoveTransferOut,
		ArticleID:   articleID,
		WarehouseID: &fromWarehouseID,
		Quantity:    qty,
	}
	if err := s.AddStockMovement(outMove, userID); err != nil {
		return err
	}

	// Entrée dans l'entrepôt destination
	var cmup float64
	s.db.QueryRow(`SELECT cmup FROM articles WHERE id=?`, articleID).Scan(&cmup)

	inMove := &models.StockMovement{
		Date:        now,
		Type:        models.StockMoveTransferIn,
		ArticleID:   articleID,
		WarehouseID: &toWarehouseID,
		Quantity:    qty,
		UnitPrice:   cmup,
	}
	return s.AddStockMovement(inMove, userID)
}

// GetStockValue retourne la valeur totale du stock
func (s *StockService) GetStockValue() float64 {
	var value float64
	s.db.QueryRow(`SELECT COALESCE(SUM(stock_qty * cmup), 0) FROM articles WHERE is_active=1`).Scan(&value)
	return value
}

// ConfirmInventory confirme un inventaire et ajuste le stock
func (s *StockService) ConfirmInventory(inventoryID, userID int) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Charger les lignes d'inventaire
	rows, err := tx.Query(`
		SELECT article_id, theoretical_qty, physical_qty, difference 
		FROM inventory_lines WHERE inventory_id=?`, inventoryID)
	if err != nil {
		return err
	}

	now := utils.NowString()
	var lines []struct {
		ArticleID int
		TheoQty   float64
		PhysQty   float64
		Diff      float64
	}
	for rows.Next() {
		var l struct {
			ArticleID int
			TheoQty   float64
			PhysQty   float64
			Diff      float64
		}
		rows.Scan(&l.ArticleID, &l.TheoQty, &l.PhysQty, &l.Diff)
		lines = append(lines, l)
	}
	rows.Close()

	for _, l := range lines {
		if l.Diff == 0 {
			continue
		}

		var moveType string
		var moveQty float64
		if l.Diff > 0 {
			moveType = models.StockMoveAdjustIn
			moveQty = l.Diff
		} else {
			moveType = models.StockMoveAdjustOut
			moveQty = -l.Diff
		}

		tx.Exec(`
			INSERT INTO stock_movements (date, type, article_id, quantity, notes, created_by)
			VALUES (?,?,?,?,?,?)`,
			now, moveType, l.ArticleID, moveQty,
			fmt.Sprintf("Ajustement inventaire #%d", inventoryID), userID)

		tx.Exec(`UPDATE articles SET stock_qty=? WHERE id=?`, l.PhysQty, l.ArticleID)
	}

	tx.Exec(`UPDATE inventories SET status='confirmed' WHERE id=?`, inventoryID)
	tx.Exec(`INSERT INTO audit_log (user_id, action_type, module, description) VALUES (?,?,?,?)`,
		userID, "confirm", "inventory", fmt.Sprintf("Confirmation inventaire #%d", inventoryID))

	return tx.Commit()
}
