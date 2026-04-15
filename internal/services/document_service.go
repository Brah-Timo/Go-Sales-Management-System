package services

import (
	"database/sql"
	"fmt"
	"gestion-commerciale/internal/models"
	"gestion-commerciale/internal/database/queries"
	"gestion-commerciale/pkg/utils"
	"math"
	"time"
)

// DocumentService gère les opérations sur les documents commerciaux
type DocumentService struct {
	db          *sql.DB
	numbering   *NumberingService
	stock       *StockService
}

// NewDocumentService crée un service de documents
func NewDocumentService(db *sql.DB) *DocumentService {
	return &DocumentService{
		db:        db,
		numbering: NewNumberingService(db),
		stock:     NewStockService(db),
	}
}

// NewDocument crée un nouveau document vide
func (s *DocumentService) NewDocument(docType string, year, userID int) *models.Document {
	docNumber, _ := s.numbering.NextNumber(docType, year)
	now := time.Now()
	return &models.Document{
		DocType:       docType,
		DocNumber:     docNumber,
		Date:          now.Format("2006-01-02 15:04:05"),
		PaymentMethod: models.PaymentCash,
		PaymentTerms:  "immediate",
		Status:        models.StatusDraft,
		Lines:         []models.DocumentLine{},
		CreatedBy:     &userID,
	}
}

// CalculateSummary calcule les totaux d'un document
func (s *DocumentService) CalculateSummary(doc *models.Document, taxConfig models.TaxConfig) models.InvoiceSummary {
	var summary models.InvoiceSummary
	var lines []utils.LineForCalc

	for _, l := range doc.Lines {
		lines = append(lines, utils.LineForCalc{
			AmountHT:       l.AmountHT,
			DiscountAmount: l.DiscountAmount,
			TVARate:        l.TVARate,
			TVAAmount:      l.TVAAmount,
			AmountTTC:      l.AmountTTC,
		})
	}

	result := utils.CalculateDocumentSummary(
		lines, doc.GlobalDiscountPct,
		doc.PaymentMethod,
		taxConfig.TimbreRate, taxConfig.TimbreMax,
		taxConfig.TimbreExemption, taxConfig.AutoTimbre,
	)

	summary.TotalHT = result.TotalHT
	summary.TotalDiscount = result.TotalDiscount
	summary.NetHT = result.NetHT
	summary.TVA9 = result.TVA9
	summary.TVA19 = result.TVA19
	summary.TotalTVA = result.TotalTVA
	summary.TotalTTC = result.TotalTTC
	summary.Timbre = result.Timbre
	summary.NetToPay = result.NetToPay
	summary.AmountInWordsFr = utils.NumberToWordsFr(result.NetToPay)

	return summary
}

// ApplySummaryToDocument applique le résumé calculé au document
func (s *DocumentService) ApplySummaryToDocument(doc *models.Document, summary models.InvoiceSummary) {
	doc.TotalHT = summary.TotalHT
	doc.TotalDiscount = summary.TotalDiscount
	doc.NetHT = summary.NetHT
	doc.TotalTVA = summary.TotalTVA
	doc.TotalTTC = summary.TotalTTC
	doc.Timbre = summary.Timbre
	doc.NetAmount = summary.NetToPay
	doc.AmountRemaining = summary.NetToPay - doc.AmountPaid
	doc.TVA9 = summary.TVA9
	doc.TVA19 = summary.TVA19
	doc.AmountInWordsFr = summary.AmountInWordsFr
}

// SaveDraft sauvegarde un document en brouillon
func (s *DocumentService) SaveDraft(doc *models.Document, taxConfig models.TaxConfig) error {
	summary := s.CalculateSummary(doc, taxConfig)
	s.ApplySummaryToDocument(doc, summary)
	doc.Status = models.StatusDraft

	if err := queries.SaveDocument(s.db, doc); err != nil {
		return err
	}
	return queries.SaveDocumentLines(s.db, doc.ID, doc.Lines)
}

// ConfirmSaleInvoice confirme une facture de vente
func (s *DocumentService) ConfirmSaleInvoice(doc *models.Document, taxConfig models.TaxConfig, userID int) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Calculer et appliquer les totaux
	summary := s.CalculateSummary(doc, taxConfig)
	s.ApplySummaryToDocument(doc, summary)
	doc.Status = models.StatusConfirmed

	// Sauvegarder le document
	if err := queries.SaveDocument(s.db, doc); err != nil {
		return err
	}
	if err := queries.SaveDocumentLines(s.db, doc.ID, doc.Lines); err != nil {
		return err
	}

	// Mouvements de stock
	for _, line := range doc.Lines {
		if line.ArticleID == nil {
			continue
		}
		// Sortie stock
		_, err = tx.Exec(`
			INSERT INTO stock_movements (date, type, article_id, warehouse_id, quantity, unit_price, reference_doc_id, created_by)
			VALUES (?,?,?,?,?,?,?,?)`,
			doc.Date, models.StockMoveSaleOut, *line.ArticleID, doc.WarehouseID,
			line.Quantity, line.UnitPriceHT, doc.ID, userID)
		if err != nil {
			return fmt.Errorf("erreur mouvement stock: %w", err)
		}

		// Mettre à jour le stock article
		_, err = tx.Exec(`UPDATE articles SET stock_qty = stock_qty - ? WHERE id=?`,
			line.Quantity, *line.ArticleID)
		if err != nil {
			return fmt.Errorf("erreur mise à jour stock: %w", err)
		}
	}

	// Mettre à jour le solde client
	if doc.ClientID != nil && *doc.ClientID > 0 {
		_, err = tx.Exec(`UPDATE clients SET balance = balance + ? WHERE id=?`,
			doc.NetAmount, *doc.ClientID)
		if err != nil {
			return err
		}
	}

	// Si paiement immédiat en espèces, créer mouvement de caisse
	if doc.PaymentMethod == models.PaymentCash && doc.PaymentTerms == "immediate" {
		clientName := "Client de passage"
		if doc.ClientID != nil {
			s.db.QueryRow(`SELECT name_fr FROM clients WHERE id=?`, *doc.ClientID).Scan(&clientName)
		}
		_, err = tx.Exec(`
			INSERT INTO cash_movements (date, type, category, description, reference, party_name, amount, created_by)
			VALUES (?,?,?,?,?,?,?,?)`,
			doc.Date, "in", "Encaissement client",
			"Vente "+doc.DocNumber, doc.DocNumber, clientName, doc.NetAmount, userID)
		if err != nil {
			return err
		}

		// Marquer comme payé
		_, err = tx.Exec(`UPDATE documents SET status='paid', amount_paid=?, amount_remaining=0 WHERE id=?`,
			doc.NetAmount, doc.ID)
		if err != nil {
			return err
		}
		doc.Status = models.StatusPaid
		doc.AmountPaid = doc.NetAmount
		doc.AmountRemaining = 0

		// Réduire le solde client
		if doc.ClientID != nil && *doc.ClientID > 0 {
			tx.Exec(`UPDATE clients SET balance = balance - ? WHERE id=?`, doc.NetAmount, *doc.ClientID)
		}
	}

	// Journal d'audit
	tx.Exec(`INSERT INTO audit_log (user_id, action_type, module, description) VALUES (?,?,?,?)`,
		userID, "confirm", "documents", "Confirmation facture "+doc.DocNumber)

	return tx.Commit()
}

// ConfirmPurchaseInvoice confirme une facture d'achat
func (s *DocumentService) ConfirmPurchaseInvoice(doc *models.Document, taxConfig models.TaxConfig, userID int, autoUpdatePrice bool) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	summary := s.CalculateSummary(doc, taxConfig)
	s.ApplySummaryToDocument(doc, summary)
	doc.Status = models.StatusConfirmed

	if err := queries.SaveDocument(s.db, doc); err != nil {
		return err
	}
	if err := queries.SaveDocumentLines(s.db, doc.ID, doc.Lines); err != nil {
		return err
	}

	for _, line := range doc.Lines {
		if line.ArticleID == nil {
			continue
		}

		// Entrée stock
		tx.Exec(`
			INSERT INTO stock_movements (date, type, article_id, warehouse_id, quantity, unit_price, reference_doc_id, created_by)
			VALUES (?,?,?,?,?,?,?,?)`,
			doc.Date, models.StockMovePurchaseIn, *line.ArticleID, doc.WarehouseID,
			line.Quantity, line.UnitPriceHT, doc.ID, userID)

		// Mise à jour stock
		tx.Exec(`UPDATE articles SET stock_qty = stock_qty + ?, purchase_price = ? WHERE id=?`,
			line.Quantity, line.UnitPriceHT, *line.ArticleID)

		// Recalculer CMUP
		s.recalculateCMUP(tx, *line.ArticleID)

		// Mise à jour prix de vente si option activée
		if autoUpdatePrice {
			var cmup, marginPct, tvaRate float64
			tx.QueryRow(`SELECT cmup, margin_percent, tva_rate FROM articles WHERE id=?`, *line.ArticleID).
				Scan(&cmup, &marginPct, &tvaRate)
			if cmup > 0 {
				newHT := math.Round(cmup*(1+marginPct/100)*100) / 100
				newTTC := math.Round(newHT*(1+tvaRate/100)*100) / 100
				tx.Exec(`UPDATE articles SET sale_price_ht=?, sale_price_ttc=? WHERE id=?`, newHT, newTTC, *line.ArticleID)
			}
		}
	}

	// Mettre à jour solde fournisseur
	if doc.SupplierID != nil {
		tx.Exec(`UPDATE suppliers SET balance = balance + ? WHERE id=?`, doc.NetAmount, *doc.SupplierID)
	}

	tx.Exec(`INSERT INTO audit_log (user_id, action_type, module, description) VALUES (?,?,?,?)`,
		userID, "confirm", "documents", "Confirmation facture achat "+doc.DocNumber)

	return tx.Commit()
}

// recalculateCMUP recalcule le CMUP d'un article dans une transaction
func (s *DocumentService) recalculateCMUP(tx *sql.Tx, articleID int) {
	rows, _ := tx.Query(`
		SELECT type, quantity, unit_price FROM stock_movements
		WHERE article_id=? ORDER BY date, id`, articleID)
	if rows == nil {
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

	cmup := state.CMUP()
	tx.Exec(`UPDATE articles SET cmup=? WHERE id=?`, cmup, articleID)
}

// CancelDocument annule un document
func (s *DocumentService) CancelDocument(docID, userID int) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var doc models.Document
	err = tx.QueryRow(`SELECT id, doc_type, doc_number, status, client_id, supplier_id, net_amount 
		FROM documents WHERE id=?`, docID).
		Scan(&doc.ID, &doc.DocType, &doc.DocNumber, &doc.Status, &doc.ClientID, &doc.SupplierID, &doc.NetAmount)
	if err != nil {
		return err
	}

	if doc.Status == models.StatusPaid {
		return fmt.Errorf("impossible d'annuler une facture déjà payée")
	}

	_, err = tx.Exec(`UPDATE documents SET status='cancelled' WHERE id=?`, docID)
	if err != nil {
		return err
	}

	// Annuler les mouvements de stock associés
	if doc.Status == models.StatusConfirmed {
		// Récupérer les lignes pour inverser le stock
		rows, _ := tx.Query(`SELECT article_id, quantity, unit_price FROM document_lines WHERE document_id=?`, docID)
		if rows != nil {
			for rows.Next() {
				var artID *int
				var qty, price float64
				rows.Scan(&artID, &qty, &price)
				if artID != nil {
					if doc.DocType == models.DocTypeFA {
						tx.Exec(`UPDATE articles SET stock_qty = stock_qty + ? WHERE id=?`, qty, *artID)
					} else if doc.DocType == models.DocTypeFAC {
						tx.Exec(`UPDATE articles SET stock_qty = stock_qty - ? WHERE id=?`, qty, *artID)
					}
				}
			}
			rows.Close()
		}

		// Inverser les soldes
		if doc.ClientID != nil {
			tx.Exec(`UPDATE clients SET balance = balance - ? WHERE id=?`, doc.NetAmount, *doc.ClientID)
		}
		if doc.SupplierID != nil {
			tx.Exec(`UPDATE suppliers SET balance = balance - ? WHERE id=?`, doc.NetAmount, *doc.SupplierID)
		}
	}

	tx.Exec(`INSERT INTO audit_log (user_id, action_type, module, description) VALUES (?,?,?,?)`,
		userID, "cancel", "documents", "Annulation document "+doc.DocNumber)

	return tx.Commit()
}

// ConvertDocument convertit un document vers un autre type
func (s *DocumentService) ConvertDocument(sourceDocID int, targetType string, year, userID int) (*models.Document, error) {
	source, err := queries.GetDocumentByID(s.db, sourceDocID)
	if err != nil {
		return nil, err
	}

	docNumber, err := s.numbering.NextNumber(targetType, year)
	if err != nil {
		return nil, err
	}

	newDoc := &models.Document{
		DocType:       targetType,
		DocNumber:     docNumber,
		Date:          time.Now().Format("2006-01-02 15:04:05"),
		ClientID:      source.ClientID,
		SupplierID:    source.SupplierID,
		WarehouseID:   source.WarehouseID,
		PaymentMethod: source.PaymentMethod,
		PaymentTerms:  source.PaymentTerms,
		PriceListID:   source.PriceListID,
		Notes:         source.Notes,
		Status:        models.StatusDraft,
		SourceDocID:   &source.ID,
		CreatedBy:     &userID,
		Lines:         make([]models.DocumentLine, len(source.Lines)),
	}

	// Copier les lignes
	for i, l := range source.Lines {
		newLine := l
		newLine.ID = 0
		newLine.DocumentID = 0
		newDoc.Lines[i] = newLine
	}

	return newDoc, nil
}
