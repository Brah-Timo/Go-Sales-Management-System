package services

import (
	"database/sql"
	"fmt"
	"gestion-commerciale/internal/models"
	"time"
)

// InventoryService gère les opérations d'inventaire
type InventoryService struct {
	db *sql.DB
}

// NewInventoryService crée un service d'inventaire
func NewInventoryService(db *sql.DB) *InventoryService {
	return &InventoryService{db: db}
}

// CreateInventory crée une nouvelle opération d'inventaire
func (s *InventoryService) CreateInventory(inventoryType string, categoryID int, userID int) (*models.Inventory, error) {
	tx, err := s.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	inv := &models.Inventory{
		Date:      time.Now().Format("2006-01-02"),
		Type:      inventoryType,
		Status:    "draft",
		CreatedBy: &userID,
	}

	res, err := tx.Exec(`
		INSERT INTO inventories (date, type, status, created_by)
		VALUES (?,?,?,?)`,
		inv.Date, inv.Type, inv.Status, userID)
	if err != nil {
		return nil, err
	}
	id, _ := res.LastInsertId()
	inv.ID = int(id)

	// Créer les lignes d'inventaire
	query := `SELECT id, reference, barcode, name_fr, stock_qty, cmup FROM articles WHERE is_active=1`
	args := []interface{}{}
	if categoryID > 0 {
		query += ` AND category_id=?`
		args = append(args, categoryID)
	}
	query += ` ORDER BY name_fr`

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var artID int
		var ref, barcode, name string
		var stockQty, cmup float64
		rows.Scan(&artID, &ref, &barcode, &name, &stockQty, &cmup)

		line := models.InventoryLine{
			InventoryID:    inv.ID,
			ArticleID:      artID,
			Reference:      ref,
			Barcode:        barcode,
			ArticleName:    name,
			TheoreticalQty: stockQty,
			PhysicalQty:    0,
			Difference:     -stockQty,
			Value:          -stockQty * cmup,
		}

		tx.Exec(`
			INSERT INTO inventory_lines (inventory_id, article_id, theoretical_qty, physical_qty, difference, value)
			VALUES (?,?,?,?,?,?)`,
			inv.ID, artID, stockQty, 0, -stockQty, -stockQty*cmup)

		inv.Lines = append(inv.Lines, line)
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	s.db.Exec(`INSERT INTO audit_log (user_id, action_type, module, description) VALUES (?,?,?,?)`,
		userID, "create", "inventory", fmt.Sprintf("Création inventaire ID %d", inv.ID))

	return inv, nil
}

// GetInventory retourne un inventaire avec ses lignes
func (s *InventoryService) GetInventory(inventoryID int) (*models.Inventory, error) {
	var inv models.Inventory
	err := s.db.QueryRow(`SELECT id, date, type, status, notes, created_by, created_at FROM inventories WHERE id=?`, inventoryID).
		Scan(&inv.ID, &inv.Date, &inv.Type, &inv.Status, &inv.Notes, &inv.CreatedBy, &inv.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("inventaire introuvable: %w", err)
	}

	rows, err := s.db.Query(`
		SELECT il.id, il.inventory_id, il.article_id,
		  a.name_fr, a.reference, a.barcode,
		  il.theoretical_qty, il.physical_qty, il.difference, il.value, il.note
		FROM inventory_lines il
		JOIN articles a ON il.article_id=a.id
		WHERE il.inventory_id=?
		ORDER BY a.name_fr`, inventoryID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var l models.InventoryLine
		rows.Scan(&l.ID, &l.InventoryID, &l.ArticleID, &l.ArticleName, &l.Reference, &l.Barcode,
			&l.TheoreticalQty, &l.PhysicalQty, &l.Difference, &l.Value, &l.Note)
		inv.Lines = append(inv.Lines, l)
	}

	return &inv, nil
}

// UpdateInventoryLine met à jour la quantité physique d'une ligne
func (s *InventoryService) UpdateInventoryLine(inventoryID, articleID int, physicalQty float64, note string) error {
	var theoreticalQty, cmup float64
	s.db.QueryRow(`SELECT theoretical_qty FROM inventory_lines WHERE inventory_id=? AND article_id=?`,
		inventoryID, articleID).Scan(&theoreticalQty)
	s.db.QueryRow(`SELECT cmup FROM articles WHERE id=?`, articleID).Scan(&cmup)

	difference := physicalQty - theoreticalQty
	value := difference * cmup

	_, err := s.db.Exec(`
		UPDATE inventory_lines
		SET physical_qty=?, difference=?, value=?, note=?
		WHERE inventory_id=? AND article_id=?`,
		physicalQty, difference, value, note, inventoryID, articleID)
	return err
}

// UpdateInventoryByBarcode met à jour une ligne par scan de code-barres
func (s *InventoryService) UpdateInventoryByBarcode(inventoryID int, barcode string, addQty float64) (*models.InventoryLine, error) {
	var articleID int
	err := s.db.QueryRow(`SELECT id FROM articles WHERE barcode=? OR reference=?`, barcode, barcode).Scan(&articleID)
	if err != nil {
		return nil, fmt.Errorf("article non trouvé: %s", barcode)
	}

	// Vérifier que la ligne existe dans l'inventaire
	var existingPhysical float64
	err = s.db.QueryRow(`SELECT physical_qty FROM inventory_lines WHERE inventory_id=? AND article_id=?`,
		inventoryID, articleID).Scan(&existingPhysical)
	if err != nil {
		return nil, fmt.Errorf("article non présent dans cet inventaire")
	}

	newPhysical := existingPhysical + addQty
	if err := s.UpdateInventoryLine(inventoryID, articleID, newPhysical, ""); err != nil {
		return nil, err
	}

	// Retourner la ligne mise à jour
	var line models.InventoryLine
	s.db.QueryRow(`
		SELECT il.id, il.inventory_id, il.article_id, a.name_fr, a.reference, a.barcode,
		  il.theoretical_qty, il.physical_qty, il.difference, il.value, il.note
		FROM inventory_lines il
		JOIN articles a ON il.article_id=a.id
		WHERE il.inventory_id=? AND il.article_id=?`, inventoryID, articleID).
		Scan(&line.ID, &line.InventoryID, &line.ArticleID, &line.ArticleName, &line.Reference, &line.Barcode,
			&line.TheoreticalQty, &line.PhysicalQty, &line.Difference, &line.Value, &line.Note)

	return &line, nil
}

// ConfirmInventory confirme l'inventaire et ajuste le stock
func (s *InventoryService) ConfirmInventory(inventoryID, userID int) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Vérifier le statut
	var status string
	tx.QueryRow(`SELECT status FROM inventories WHERE id=?`, inventoryID).Scan(&status)
	if status != "draft" {
		return fmt.Errorf("cet inventaire n'est pas en statut brouillon")
	}

	// Appliquer les ajustements
	rows, _ := tx.Query(`
		SELECT il.article_id, il.physical_qty, il.difference, il.value, a.cmup
		FROM inventory_lines il
		JOIN articles a ON il.article_id=a.id
		WHERE il.inventory_id=? AND il.difference != 0`, inventoryID)
	if rows != nil {
		defer rows.Close()
		for rows.Next() {
			var artID int
			var physQty, diff, value, cmup float64
			rows.Scan(&artID, &physQty, &diff, &value, &cmup)

			// Mettre à jour le stock
			tx.Exec(`UPDATE articles SET stock_qty=? WHERE id=?`, physQty, artID)

			// Créer le mouvement de stock
			mvType := "adjustment_in"
			if diff < 0 {
				mvType = "adjustment_out"
				diff = -diff
			}
			tx.Exec(`
				INSERT INTO stock_movements (date, type, article_id, quantity, unit_price, reference_doc_id, notes, created_by)
				VALUES (datetime('now'), ?, ?, ?, ?, ?, 'Ajustement inventaire', ?)`,
				mvType, artID, diff, cmup, inventoryID, userID)
		}
	}

	// Confirmer l'inventaire
	_, err = tx.Exec(`UPDATE inventories SET status='confirmed' WHERE id=?`, inventoryID)
	if err != nil {
		return err
	}

	tx.Exec(`INSERT INTO audit_log (user_id, action_type, module, description) VALUES (?,?,?,?)`,
		userID, "confirm", "inventory", fmt.Sprintf("Confirmation inventaire ID %d", inventoryID))

	return tx.Commit()
}

// GetInventoryList retourne la liste des inventaires
func (s *InventoryService) GetInventoryList() ([]models.Inventory, error) {
	rows, err := s.db.Query(`
		SELECT id, date, type, status, notes, created_by, created_at
		FROM inventories ORDER BY date DESC, id DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var inventories []models.Inventory
	for rows.Next() {
		var inv models.Inventory
		rows.Scan(&inv.ID, &inv.Date, &inv.Type, &inv.Status, &inv.Notes, &inv.CreatedBy, &inv.CreatedAt)
		inventories = append(inventories, inv)
	}
	return inventories, nil
}

// SaveInventoryNotes sauvegarde les notes d'un inventaire
func (s *InventoryService) SaveInventoryNotes(inventoryID int, notes string) error {
	_, err := s.db.Exec(`UPDATE inventories SET notes=? WHERE id=?`, notes, inventoryID)
	return err
}

// GetInventoryDifferenceSummary retourne le résumé des écarts
func (s *InventoryService) GetInventoryDifferenceSummary(inventoryID int) (totalPositive, totalNegative, totalValue float64, positiveCount, negativeCount int) {
	s.db.QueryRow(`
		SELECT
		  COALESCE(SUM(CASE WHEN difference > 0 THEN value ELSE 0 END), 0),
		  COALESCE(SUM(CASE WHEN difference < 0 THEN ABS(value) ELSE 0 END), 0),
		  COALESCE(SUM(ABS(value)), 0),
		  COUNT(CASE WHEN difference > 0 THEN 1 END),
		  COUNT(CASE WHEN difference < 0 THEN 1 END)
		FROM inventory_lines WHERE inventory_id=?`, inventoryID).
		Scan(&totalPositive, &totalNegative, &totalValue, &positiveCount, &negativeCount)
	return
}
