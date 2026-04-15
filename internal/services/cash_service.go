package services

import (
	"database/sql"
	"fmt"
	"gestion-commerciale/internal/models"
)

// CashService gère la caisse, la banque et les chèques
type CashService struct {
	db *sql.DB
}

// NewCashService crée un service caisse
func NewCashService(db *sql.DB) *CashService {
	return &CashService{db: db}
}

// ─────────────────────────────────────────────────────────────────────────────
// CAISSE
// ─────────────────────────────────────────────────────────────────────────────

// GetCashBalance retourne le solde actuel de la caisse
func (s *CashService) GetCashBalance() float64 {
	var balance float64
	s.db.QueryRow(`
		SELECT COALESCE(SUM(CASE WHEN type='in' THEN amount ELSE -amount END), 0)
		FROM cash_movements`).Scan(&balance)
	return balance
}

// GetCashMovements retourne les mouvements de caisse pour une période
func (s *CashService) GetCashMovements(from, to string) ([]models.CashMovement, error) {
	rows, err := s.db.Query(`
		SELECT id, date, type, category, description, reference, party_name, amount
		FROM cash_movements
		WHERE date(date) BETWEEN ? AND ?
		ORDER BY date, id`, from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var movements []models.CashMovement
	var runningBalance float64
	for rows.Next() {
		var m models.CashMovement
		rows.Scan(&m.ID, &m.Date, &m.Type, &m.Category, &m.Description,
			&m.Reference, &m.PartyName, &m.Amount)
		if m.Type == "in" {
			runningBalance += m.Amount
		} else {
			runningBalance -= m.Amount
		}
		m.Balance = runningBalance
		movements = append(movements, m)
	}
	return movements, nil
}

// AddCashMovement ajoute un mouvement de caisse manuel
func (s *CashService) AddCashMovement(m *models.CashMovement, userID int) error {
	_, err := s.db.Exec(`
		INSERT INTO cash_movements (date, type, category, description, reference, party_name, amount, created_by)
		VALUES (?,?,?,?,?,?,?,?)`,
		m.Date, m.Type, m.Category, m.Description, m.Reference, m.PartyName, m.Amount, userID)
	if err != nil {
		return fmt.Errorf("erreur ajout mouvement caisse: %w", err)
	}
	s.db.Exec(`INSERT INTO audit_log (user_id, action_type, module, description) VALUES (?,?,?,?)`,
		userID, "create", "cash", fmt.Sprintf("Mouvement caisse %s %.2f DA", m.Type, m.Amount))
	return nil
}

// GetDailyCashSummary retourne le résumé journalier de la caisse
func (s *CashService) GetDailyCashSummary(date string) (totalIn, totalOut float64, err error) {
	err = s.db.QueryRow(`
		SELECT
		  COALESCE(SUM(CASE WHEN type='in' THEN amount ELSE 0 END), 0),
		  COALESCE(SUM(CASE WHEN type='out' THEN amount ELSE 0 END), 0)
		FROM cash_movements
		WHERE date(date) = ?`, date).Scan(&totalIn, &totalOut)
	return
}

// ─────────────────────────────────────────────────────────────────────────────
// BANQUE
// ─────────────────────────────────────────────────────────────────────────────

// GetAllBankAccounts retourne tous les comptes bancaires
func (s *CashService) GetAllBankAccounts() ([]models.BankAccount, error) {
	rows, err := s.db.Query(`SELECT id, bank_name, branch, account_number, rib, balance FROM bank_accounts ORDER BY bank_name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var accounts []models.BankAccount
	for rows.Next() {
		var a models.BankAccount
		rows.Scan(&a.ID, &a.BankName, &a.Branch, &a.AccountNumber, &a.RIB, &a.Balance)
		accounts = append(accounts, a)
	}
	return accounts, nil
}

// GetBankMovements retourne les mouvements d'un compte bancaire
func (s *CashService) GetBankMovements(bankAccountID int, from, to string) ([]models.BankMovement, error) {
	rows, err := s.db.Query(`
		SELECT id, bank_account_id, date, type, description, reference, debit, credit, is_reconciled
		FROM bank_movements
		WHERE bank_account_id=? AND date(date) BETWEEN ? AND ?
		ORDER BY date, id`, bankAccountID, from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var movements []models.BankMovement
	for rows.Next() {
		var m models.BankMovement
		rows.Scan(&m.ID, &m.BankAccountID, &m.Date, &m.Type, &m.Description,
			&m.Reference, &m.Debit, &m.Credit, &m.IsReconciled)
		movements = append(movements, m)
	}
	return movements, nil
}

// AddBankMovement ajoute un mouvement bancaire
func (s *CashService) AddBankMovement(m *models.BankMovement) error {
	_, err := s.db.Exec(`
		INSERT INTO bank_movements (bank_account_id, date, type, description, reference, debit, credit)
		VALUES (?,?,?,?,?,?,?)`,
		m.BankAccountID, m.Date, m.Type, m.Description, m.Reference, m.Debit, m.Credit)
	if err != nil {
		return err
	}
	// Mettre à jour le solde du compte
	_, err = s.db.Exec(`UPDATE bank_accounts SET balance = balance + ? - ? WHERE id=?`,
		m.Credit, m.Debit, m.BankAccountID)
	return err
}

// SaveBankAccount sauvegarde un compte bancaire
func (s *CashService) SaveBankAccount(a *models.BankAccount) error {
	if a.ID == 0 {
		res, err := s.db.Exec(`INSERT INTO bank_accounts (bank_name, branch, account_number, rib, balance) VALUES (?,?,?,?,?)`,
			a.BankName, a.Branch, a.AccountNumber, a.RIB, a.Balance)
		if err != nil {
			return err
		}
		id, _ := res.LastInsertId()
		a.ID = int(id)
		return nil
	}
	_, err := s.db.Exec(`UPDATE bank_accounts SET bank_name=?, branch=?, account_number=?, rib=? WHERE id=?`,
		a.BankName, a.Branch, a.AccountNumber, a.RIB, a.ID)
	return err
}

// ─────────────────────────────────────────────────────────────────────────────
// CHÈQUES
// ─────────────────────────────────────────────────────────────────────────────

// GetCheques retourne tous les chèques par type
func (s *CashService) GetCheques(chequeType, status string) ([]models.Cheque, error) {
	query := `SELECT id, type, cheque_number, date, due_date, amount, payer_payee, bank_name, status, reject_reason, notes FROM cheques WHERE 1=1`
	args := []interface{}{}

	if chequeType != "" {
		query += ` AND type=?`
		args = append(args, chequeType)
	}
	if status != "" {
		query += ` AND status=?`
		args = append(args, status)
	}
	query += ` ORDER BY due_date`

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cheques []models.Cheque
	for rows.Next() {
		var c models.Cheque
		rows.Scan(&c.ID, &c.Type, &c.ChequeNumber, &c.Date, &c.DueDate,
			&c.Amount, &c.PayerPayee, &c.BankName, &c.Status, &c.RejectReason, &c.Notes)
		cheques = append(cheques, c)
	}
	return cheques, nil
}

// SaveCheque sauvegarde un chèque
func (s *CashService) SaveCheque(c *models.Cheque) error {
	if c.ID == 0 {
		res, err := s.db.Exec(`
			INSERT INTO cheques (type, cheque_number, date, due_date, amount, payer_payee, bank_name, status, notes)
			VALUES (?,?,?,?,?,?,?,?,?)`,
			c.Type, c.ChequeNumber, c.Date, c.DueDate, c.Amount, c.PayerPayee, c.BankName, c.Status, c.Notes)
		if err != nil {
			return err
		}
		id, _ := res.LastInsertId()
		c.ID = int(id)
		return nil
	}
	_, err := s.db.Exec(`
		UPDATE cheques SET status=?, reject_reason=?, notes=? WHERE id=?`,
		c.Status, c.RejectReason, c.Notes, c.ID)
	return err
}

// UpdateChequeStatus met à jour le statut d'un chèque
func (s *CashService) UpdateChequeStatus(chequeID int, status, reason string, bankAccountID int) error {
	tx, _ := s.db.Begin()
	defer tx.Rollback()

	_, err := tx.Exec(`UPDATE cheques SET status=?, reject_reason=? WHERE id=?`, status, reason, chequeID)
	if err != nil {
		return err
	}

	// Si déposé en banque
	if status == "deposited" && bankAccountID > 0 {
		var amount float64
		var payerPayee string
		tx.QueryRow(`SELECT amount, payer_payee FROM cheques WHERE id=?`, chequeID).Scan(&amount, &payerPayee)

		tx.Exec(`
			INSERT INTO bank_movements (bank_account_id, date, type, description, reference, credit)
			VALUES (?, datetime('now'), 'cheque_received', ?, ?, ?)`,
			bankAccountID, "Chèque "+payerPayee, fmt.Sprintf("CHQ-%d", chequeID), amount)

		tx.Exec(`UPDATE bank_accounts SET balance=balance+? WHERE id=?`, amount, bankAccountID)
	}

	return tx.Commit()
}

// GetChequesNearDue retourne les chèques à échéance dans les 7 prochains jours
func (s *CashService) GetChequesNearDue() ([]models.Cheque, error) {
	rows, err := s.db.Query(`
		SELECT id, type, cheque_number, date, due_date, amount, payer_payee, bank_name, status
		FROM cheques
		WHERE status IN ('pending','deposited')
		  AND due_date BETWEEN date('now') AND date('now','+7 days')
		ORDER BY due_date`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var cheques []models.Cheque
	for rows.Next() {
		var c models.Cheque
		rows.Scan(&c.ID, &c.Type, &c.ChequeNumber, &c.Date, &c.DueDate,
			&c.Amount, &c.PayerPayee, &c.BankName, &c.Status)
		cheques = append(cheques, c)
	}
	return cheques, nil
}

// ─────────────────────────────────────────────────────────────────────────────
// DÉPENSES
// ─────────────────────────────────────────────────────────────────────────────

// GetExpenses retourne les dépenses pour une période
func (s *CashService) GetExpenses(from, to string) ([]models.Expense, error) {
	rows, err := s.db.Query(`
		SELECT e.id, e.date, e.category_id, COALESCE(ec.name,'') as category_name,
		  e.description, e.amount, e.payment_method
		FROM expenses e
		LEFT JOIN expense_categories ec ON e.category_id=ec.id
		WHERE date(e.date) BETWEEN ? AND ?
		ORDER BY e.date DESC`, from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var expenses []models.Expense
	for rows.Next() {
		var exp models.Expense
		rows.Scan(&exp.ID, &exp.Date, &exp.CategoryID, &exp.CategoryName,
			&exp.Description, &exp.Amount, &exp.PaymentMethod)
		expenses = append(expenses, exp)
	}
	return expenses, nil
}

// SaveExpense sauvegarde une dépense
func (s *CashService) SaveExpense(exp *models.Expense, userID int) error {
	if exp.ID == 0 {
		res, err := s.db.Exec(`
			INSERT INTO expenses (date, category_id, description, amount, payment_method, created_by)
			VALUES (?,?,?,?,?,?)`,
			exp.Date, exp.CategoryID, exp.Description, exp.Amount, exp.PaymentMethod, userID)
		if err != nil {
			return err
		}
		id, _ := res.LastInsertId()
		exp.ID = int(id)

		// Mouvement de caisse si espèces
		if exp.PaymentMethod == "cash" {
			s.db.Exec(`
				INSERT INTO cash_movements (date, type, category, description, amount, created_by)
				VALUES (?,?,?,?,?,?)`,
				exp.Date, "out", "Dépense", exp.Description, exp.Amount, userID)
		}
		return nil
	}
	_, err := s.db.Exec(`
		UPDATE expenses SET date=?, category_id=?, description=?, amount=?, payment_method=? WHERE id=?`,
		exp.Date, exp.CategoryID, exp.Description, exp.Amount, exp.PaymentMethod, exp.ID)
	return err
}

// GetExpenseCategories retourne toutes les catégories de dépenses
func (s *CashService) GetExpenseCategories() ([]models.ExpenseCategory, error) {
	rows, err := s.db.Query(`SELECT id, name FROM expense_categories ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var cats []models.ExpenseCategory
	for rows.Next() {
		var c models.ExpenseCategory
		rows.Scan(&c.ID, &c.Name)
		cats = append(cats, c)
	}
	return cats, nil
}

// ─────────────────────────────────────────────────────────────────────────────
// BALANCE ÂGÉE
// ─────────────────────────────────────────────────────────────────────────────

// GetClientAging retourne la balance âgée des clients
func (s *CashService) GetClientAging() ([]models.AgingLine, error) {
	rows, err := s.db.Query(`
		SELECT c.name_fr,
		  COALESCE(SUM(d.amount_remaining), 0) as balance,
		  COALESCE(SUM(CASE WHEN julianday('now')-julianday(d.date)<=30 THEN d.amount_remaining ELSE 0 END), 0) as age_0_30,
		  COALESCE(SUM(CASE WHEN julianday('now')-julianday(d.date) BETWEEN 31 AND 60 THEN d.amount_remaining ELSE 0 END), 0) as age_31_60,
		  COALESCE(SUM(CASE WHEN julianday('now')-julianday(d.date) BETWEEN 61 AND 90 THEN d.amount_remaining ELSE 0 END), 0) as age_61_90,
		  COALESCE(SUM(CASE WHEN julianday('now')-julianday(d.date)>90 THEN d.amount_remaining ELSE 0 END), 0) as age_90_plus
		FROM clients c
		JOIN documents d ON d.client_id=c.id AND d.doc_type='FA' AND d.status IN ('confirmed','partial')
		WHERE c.balance > 0
		GROUP BY c.id ORDER BY balance DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var lines []models.AgingLine
	for rows.Next() {
		var l models.AgingLine
		rows.Scan(&l.ClientName, &l.Balance, &l.Age0_30, &l.Age31_60, &l.Age61_90, &l.Age90Plus)
		lines = append(lines, l)
	}
	return lines, nil
}

// GetSupplierAging retourne la balance âgée des fournisseurs
func (s *CashService) GetSupplierAging() ([]models.AgingLine, error) {
	rows, err := s.db.Query(`
		SELECT sup.name_fr,
		  COALESCE(SUM(d.amount_remaining), 0) as balance,
		  COALESCE(SUM(CASE WHEN julianday('now')-julianday(d.date)<=30 THEN d.amount_remaining ELSE 0 END), 0) as age_0_30,
		  COALESCE(SUM(CASE WHEN julianday('now')-julianday(d.date) BETWEEN 31 AND 60 THEN d.amount_remaining ELSE 0 END), 0) as age_31_60,
		  COALESCE(SUM(CASE WHEN julianday('now')-julianday(d.date) BETWEEN 61 AND 90 THEN d.amount_remaining ELSE 0 END), 0) as age_61_90,
		  COALESCE(SUM(CASE WHEN julianday('now')-julianday(d.date)>90 THEN d.amount_remaining ELSE 0 END), 0) as age_90_plus
		FROM suppliers sup
		JOIN documents d ON d.supplier_id=sup.id AND d.doc_type='FAC' AND d.status IN ('confirmed','partial')
		WHERE sup.balance > 0
		GROUP BY sup.id ORDER BY balance DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var lines []models.AgingLine
	for rows.Next() {
		var l models.AgingLine
		rows.Scan(&l.ClientName, &l.Balance, &l.Age0_30, &l.Age31_60, &l.Age61_90, &l.Age90Plus)
		lines = append(lines, l)
	}
	return lines, nil
}
