package services

import (
	"database/sql"
	"fmt"
	"gestion-commerciale/internal/models"
	"gestion-commerciale/pkg/utils"
	"math"
)

// PaymentService gère les paiements
type PaymentService struct {
	db *sql.DB
}

// NewPaymentService crée un service de paiement
func NewPaymentService(db *sql.DB) *PaymentService {
	return &PaymentService{db: db}
}

// ProcessClientPayment traite un encaissement client avec ventilation FIFO
func (s *PaymentService) ProcessClientPayment(payment *models.Payment, userID int) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Insérer le paiement
	result, err := tx.Exec(`
		INSERT INTO payments (type, date, client_id, amount, payment_method,
		  cheque_number, bank_name, reference, notes, created_by)
		VALUES (?,?,?,?,?,?,?,?,?,?)`,
		"collection", payment.Date, payment.ClientID, payment.Amount,
		payment.PaymentMethod, payment.ChequeNumber, payment.BankName,
		payment.Reference, payment.Notes, userID)
	if err != nil {
		return fmt.Errorf("ProcessClientPayment insert: %w", err)
	}
	paymentID, _ := result.LastInsertId()

	// Ventilation FIFO sur les factures impayées
	remaining := payment.Amount
	rows, err := tx.Query(`
		SELECT id, amount_remaining FROM documents
		WHERE client_id=? AND doc_type='FA' AND status IN ('confirmed','partial') AND amount_remaining > 0
		ORDER BY date ASC`, *payment.ClientID)
	if err != nil {
		return err
	}

	var allocations []struct {
		DocID     int
		Allocated float64
	}
	for rows.Next() && remaining > 0.01 {
		var docID int
		var docRemaining float64
		rows.Scan(&docID, &docRemaining)

		allocated := math.Min(remaining, docRemaining)
		allocations = append(allocations, struct {
			DocID     int
			Allocated float64
		}{docID, allocated})
		remaining -= allocated
	}
	rows.Close()

	// Appliquer les ventilations
	for _, alloc := range allocations {
		tx.Exec(`INSERT INTO payment_allocations (payment_id, document_id, amount) VALUES (?,?,?)`,
			paymentID, alloc.DocID, alloc.Allocated)

		var newPaid, netAmount float64
		tx.QueryRow(`SELECT amount_paid, net_amount FROM documents WHERE id=?`, alloc.DocID).
			Scan(&newPaid, &netAmount)

		newPaid += alloc.Allocated
		newRemaining := netAmount - newPaid
		status := "partial"
		if newRemaining <= 0.01 {
			status = "paid"
			newRemaining = 0
		}

		tx.Exec(`UPDATE documents SET amount_paid=?, amount_remaining=?, status=? WHERE id=?`,
			newPaid, newRemaining, status, alloc.DocID)
	}

	// Mettre à jour le solde client
	tx.Exec(`UPDATE clients SET balance = balance - ? WHERE id=?`, payment.Amount, *payment.ClientID)

	// Mouvement de caisse si espèces
	if payment.PaymentMethod == "cash" {
		var clientName string
		tx.QueryRow(`SELECT name_fr FROM clients WHERE id=?`, *payment.ClientID).Scan(&clientName)
		tx.Exec(`
			INSERT INTO cash_movements (date, type, category, description, reference, party_name, amount, created_by)
			VALUES (?,?,?,?,?,?,?,?)`,
			payment.Date, "in", "Encaissement client",
			"Encaissement "+clientName, payment.Reference, clientName, payment.Amount, userID)
	}

	// Enregistrement chèque si paiement par chèque
	if payment.PaymentMethod == "cheque" && payment.ChequeNumber != "" {
		var clientName string
		tx.QueryRow(`SELECT name_fr FROM clients WHERE id=?`, *payment.ClientID).Scan(&clientName)
		tx.Exec(`
			INSERT INTO cheques (type, cheque_number, date, due_date, amount, payer_payee, bank_name, status, related_payment_id)
			VALUES (?,?,?,?,?,?,?,?,?)`,
			"received", payment.ChequeNumber, payment.Date, payment.Date,
			payment.Amount, clientName, payment.BankName, "pending", paymentID)
	}

	return tx.Commit()
}

// ProcessSupplierPayment traite un décaissement fournisseur
func (s *PaymentService) ProcessSupplierPayment(payment *models.Payment, userID int) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	result, err := tx.Exec(`
		INSERT INTO payments (type, date, supplier_id, amount, payment_method,
		  cheque_number, bank_name, reference, notes, created_by)
		VALUES (?,?,?,?,?,?,?,?,?,?)`,
		"disbursement", payment.Date, payment.SupplierID, payment.Amount,
		payment.PaymentMethod, payment.ChequeNumber, payment.BankName,
		payment.Reference, payment.Notes, userID)
	if err != nil {
		return err
	}
	paymentID, _ := result.LastInsertId()

	// Ventilation sur factures fournisseur
	remaining := payment.Amount
	rows, _ := tx.Query(`
		SELECT id, amount_remaining FROM documents
		WHERE supplier_id=? AND doc_type='FAC' AND status IN ('confirmed','partial') AND amount_remaining > 0
		ORDER BY date ASC`, *payment.SupplierID)

	if rows != nil {
		for rows.Next() && remaining > 0.01 {
			var docID int
			var docRemaining float64
			rows.Scan(&docID, &docRemaining)

			allocated := math.Min(remaining, docRemaining)
			tx.Exec(`INSERT INTO payment_allocations (payment_id, document_id, amount) VALUES (?,?,?)`,
				paymentID, docID, allocated)

			var newPaid, netAmount float64
			tx.QueryRow(`SELECT amount_paid, net_amount FROM documents WHERE id=?`, docID).
				Scan(&newPaid, &netAmount)
			newPaid += allocated
			newRemaining := netAmount - newPaid
			status := "partial"
			if newRemaining <= 0.01 { status = "paid"; newRemaining = 0 }
			tx.Exec(`UPDATE documents SET amount_paid=?, amount_remaining=?, status=? WHERE id=?`,
				newPaid, newRemaining, status, docID)
			remaining -= allocated
		}
		rows.Close()
	}

	// Mettre à jour solde fournisseur
	tx.Exec(`UPDATE suppliers SET balance = balance - ? WHERE id=?`, payment.Amount, *payment.SupplierID)

	// Mouvement de caisse
	if payment.PaymentMethod == "cash" {
		var supplierName string
		tx.QueryRow(`SELECT name_fr FROM suppliers WHERE id=?`, *payment.SupplierID).Scan(&supplierName)
		tx.Exec(`
			INSERT INTO cash_movements (date, type, category, description, reference, party_name, amount, created_by)
			VALUES (?,?,?,?,?,?,?,?)`,
			payment.Date, "out", "Paiement fournisseur",
			"Paiement "+supplierName, payment.Reference, supplierName, payment.Amount, userID)
	}

	// Chèque émis
	if payment.PaymentMethod == "cheque" && payment.ChequeNumber != "" {
		var supplierName string
		tx.QueryRow(`SELECT name_fr FROM suppliers WHERE id=?`, *payment.SupplierID).Scan(&supplierName)
		tx.Exec(`
			INSERT INTO cheques (type, cheque_number, date, due_date, amount, payer_payee, bank_name, status, related_payment_id)
			VALUES (?,?,?,?,?,?,?,?,?)`,
			"issued", payment.ChequeNumber, payment.Date, payment.Date,
			payment.Amount, supplierName, payment.BankName, "pending", paymentID)
	}

	return tx.Commit()
}

// AddCashMovement ajoute un mouvement de caisse manuel
func (s *PaymentService) AddCashMovement(move *models.CashMovement, userID int) error {
	_, err := s.db.Exec(`
		INSERT INTO cash_movements (date, type, category, description, reference, party_name, amount, created_by)
		VALUES (?,?,?,?,?,?,?,?)`,
		move.Date, move.Type, move.Category, move.Description,
		move.Reference, move.PartyName, move.Amount, userID)
	return err
}

// GetCashMovements retourne les mouvements de caisse avec solde cumulé
func (s *PaymentService) GetCashMovements(dateFrom, dateTo string) ([]models.CashMovement, float64, error) {
	query := `
		SELECT cm.id, cm.date, cm.type, cm.category, cm.description, cm.reference, cm.party_name, cm.amount,
		  (SELECT COALESCE(SUM(CASE WHEN cm2.type='in' THEN cm2.amount ELSE -cm2.amount END),0)
		   FROM cash_movements cm2 WHERE cm2.id <= cm.id) as running_balance
		FROM cash_movements cm WHERE 1=1`

	var args []interface{}
	if dateFrom != "" {
		query += ` AND date(cm.date) >= ?`
		args = append(args, dateFrom)
	}
	if dateTo != "" {
		query += ` AND date(cm.date) <= ?`
		args = append(args, dateTo)
	}
	query += ` ORDER BY cm.date, cm.id`

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var moves []models.CashMovement
	for rows.Next() {
		var m models.CashMovement
		rows.Scan(&m.ID, &m.Date, &m.Type, &m.Category, &m.Description,
			&m.Reference, &m.PartyName, &m.Amount, &m.Balance)
		moves = append(moves, m)
	}

	balance := utils.FormatAmount(0)
	if len(moves) > 0 {
		balance = utils.FormatAmount(moves[len(moves)-1].Balance)
	}
	_ = balance

	var totalBalance float64
	if len(moves) > 0 {
		totalBalance = moves[len(moves)-1].Balance
	}

	return moves, totalBalance, nil
}
