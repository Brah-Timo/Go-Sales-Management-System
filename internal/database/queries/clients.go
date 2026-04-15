package queries

import (
	"database/sql"
	"fmt"
	"gestion-commerciale/internal/models"
)

// GetAllClients retourne tous les clients
func GetAllClients(db *sql.DB, search string) ([]models.Client, error) {
	query := `
		SELECT c.id, c.code, c.name_ar, c.name_fr, c.type,
		       c.address, c.wilaya, c.commune, c.phone, c.mobile, c.fax, c.email,
		       c.nif, c.nis, c.rc, c.ai,
		       c.price_list_id, c.credit_limit, c.payment_terms,
		       c.discount_rate, c.balance, c.is_blocked, c.notes,
		       COALESCE(pl.name,'') as price_list_name
		FROM clients c
		LEFT JOIN price_lists pl ON c.price_list_id = pl.id
		WHERE 1=1`

	var args []interface{}
	if search != "" {
		query += ` AND (c.name_fr LIKE ? OR c.name_ar LIKE ? OR c.code LIKE ? OR c.phone LIKE ?)`
		s := "%" + search + "%"
		args = append(args, s, s, s, s)
	}
	query += ` ORDER BY c.name_fr`

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var clients []models.Client
	for rows.Next() {
		var c models.Client
		rows.Scan(
			&c.ID, &c.Code, &c.NameAr, &c.NameFr, &c.Type,
			&c.Address, &c.Wilaya, &c.Commune, &c.Phone, &c.Mobile, &c.Fax, &c.Email,
			&c.NIF, &c.NIS, &c.RC, &c.AI,
			&c.PriceListID, &c.CreditLimit, &c.PaymentTerms,
			&c.DiscountRate, &c.Balance, &c.IsBlocked, &c.Notes,
			&c.PriceListName,
		)
		clients = append(clients, c)
	}
	return clients, nil
}

// GetClientByID retourne un client par son ID
func GetClientByID(db *sql.DB, id int) (*models.Client, error) {
	var c models.Client
	err := db.QueryRow(`
		SELECT c.id, c.code, c.name_ar, c.name_fr, c.type,
		       c.address, c.wilaya, c.commune, c.phone, c.mobile, c.fax, c.email,
		       c.nif, c.nis, c.rc, c.ai,
		       c.price_list_id, c.credit_limit, c.payment_terms,
		       c.discount_rate, c.balance, c.is_blocked, c.notes,
		       COALESCE(pl.name,'') as price_list_name
		FROM clients c
		LEFT JOIN price_lists pl ON c.price_list_id = pl.id
		WHERE c.id=?`, id).Scan(
		&c.ID, &c.Code, &c.NameAr, &c.NameFr, &c.Type,
		&c.Address, &c.Wilaya, &c.Commune, &c.Phone, &c.Mobile, &c.Fax, &c.Email,
		&c.NIF, &c.NIS, &c.RC, &c.AI,
		&c.PriceListID, &c.CreditLimit, &c.PaymentTerms,
		&c.DiscountRate, &c.Balance, &c.IsBlocked, &c.Notes,
		&c.PriceListName,
	)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

// SaveClient insère ou met à jour un client
func SaveClient(db *sql.DB, c *models.Client) error {
	if c.ID == 0 {
		// Générer le code client
		if c.Code == "" {
			c.Code = NextClientCode(db)
		}
		result, err := db.Exec(`
			INSERT INTO clients (code, name_ar, name_fr, type, address, wilaya, commune,
			  phone, mobile, fax, email, nif, nis, rc, ai,
			  price_list_id, credit_limit, payment_terms, discount_rate,
			  balance, is_blocked, notes)
			VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`,
			c.Code, c.NameAr, c.NameFr, c.Type, c.Address, c.Wilaya, c.Commune,
			c.Phone, c.Mobile, c.Fax, c.Email, c.NIF, c.NIS, c.RC, c.AI,
			c.PriceListID, c.CreditLimit, c.PaymentTerms, c.DiscountRate,
			c.Balance, c.IsBlocked, c.Notes,
		)
		if err != nil {
			return fmt.Errorf("SaveClient insert: %w", err)
		}
		id, _ := result.LastInsertId()
		c.ID = int(id)
	} else {
		_, err := db.Exec(`
			UPDATE clients SET name_ar=?, name_fr=?, type=?, address=?, wilaya=?, commune=?,
			  phone=?, mobile=?, fax=?, email=?, nif=?, nis=?, rc=?, ai=?,
			  price_list_id=?, credit_limit=?, payment_terms=?, discount_rate=?,
			  is_blocked=?, notes=?
			WHERE id=?`,
			c.NameAr, c.NameFr, c.Type, c.Address, c.Wilaya, c.Commune,
			c.Phone, c.Mobile, c.Fax, c.Email, c.NIF, c.NIS, c.RC, c.AI,
			c.PriceListID, c.CreditLimit, c.PaymentTerms, c.DiscountRate,
			c.IsBlocked, c.Notes, c.ID,
		)
		if err != nil {
			return fmt.Errorf("SaveClient update: %w", err)
		}
	}
	return nil
}

// NextClientCode génère le prochain code client
func NextClientCode(db *sql.DB) string {
	var max int
	db.QueryRow(`SELECT COUNT(*) FROM clients`).Scan(&max)
	return fmt.Sprintf("C%03d", max+1)
}

// GetClientBalance retourne le solde d'un client
func GetClientBalance(db *sql.DB, clientID int) float64 {
	var balance float64
	db.QueryRow(`SELECT COALESCE(balance, 0) FROM clients WHERE id=?`, clientID).Scan(&balance)
	return balance
}

// UpdateClientBalance met à jour le solde d'un client
func UpdateClientBalance(db *sql.DB, clientID int, amount float64) error {
	_, err := db.Exec(`UPDATE clients SET balance = balance + ? WHERE id=?`, amount, clientID)
	return err
}

// GetAllSuppliers retourne tous les fournisseurs
func GetAllSuppliers(db *sql.DB, search string) ([]models.Supplier, error) {
	query := `
		SELECT id, code, name_ar, name_fr, address, wilaya, phone, mobile,
		       fax, email, nif, nis, rc, ai, payment_terms, balance,
		       rating_delivery, rating_quality, rating_pricing, notes
		FROM suppliers WHERE 1=1`

	var args []interface{}
	if search != "" {
		query += ` AND (name_fr LIKE ? OR name_ar LIKE ? OR code LIKE ?)`
		s := "%" + search + "%"
		args = append(args, s, s, s)
	}
	query += ` ORDER BY name_fr`

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var suppliers []models.Supplier
	for rows.Next() {
		var s models.Supplier
		rows.Scan(
			&s.ID, &s.Code, &s.NameAr, &s.NameFr, &s.Address, &s.Wilaya,
			&s.Phone, &s.Mobile, &s.Fax, &s.Email, &s.NIF, &s.NIS, &s.RC, &s.AI,
			&s.PaymentTerms, &s.Balance, &s.RatingDelivery, &s.RatingQuality,
			&s.RatingPricing, &s.Notes,
		)
		suppliers = append(suppliers, s)
	}
	return suppliers, nil
}

// SaveSupplier insère ou met à jour un fournisseur
func SaveSupplier(db *sql.DB, s *models.Supplier) error {
	if s.ID == 0 {
		if s.Code == "" {
			var count int
			db.QueryRow(`SELECT COUNT(*) FROM suppliers`).Scan(&count)
			s.Code = fmt.Sprintf("F%03d", count+1)
		}
		result, err := db.Exec(`
			INSERT INTO suppliers (code, name_ar, name_fr, address, wilaya,
			  phone, mobile, fax, email, nif, nis, rc, ai,
			  payment_terms, balance, rating_delivery, rating_quality, rating_pricing, notes)
			VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`,
			s.Code, s.NameAr, s.NameFr, s.Address, s.Wilaya,
			s.Phone, s.Mobile, s.Fax, s.Email, s.NIF, s.NIS, s.RC, s.AI,
			s.PaymentTerms, s.Balance, s.RatingDelivery, s.RatingQuality, s.RatingPricing, s.Notes,
		)
		if err != nil {
			return err
		}
		id, _ := result.LastInsertId()
		s.ID = int(id)
	} else {
		_, err := db.Exec(`
			UPDATE suppliers SET name_ar=?, name_fr=?, address=?, wilaya=?,
			  phone=?, mobile=?, fax=?, email=?, nif=?, nis=?, rc=?, ai=?,
			  payment_terms=?, rating_delivery=?, rating_quality=?, rating_pricing=?, notes=?
			WHERE id=?`,
			s.NameAr, s.NameFr, s.Address, s.Wilaya,
			s.Phone, s.Mobile, s.Fax, s.Email, s.NIF, s.NIS, s.RC, s.AI,
			s.PaymentTerms, s.RatingDelivery, s.RatingQuality, s.RatingPricing, s.Notes,
			s.ID,
		)
		if err != nil {
			return err
		}
	}
	return nil
}

// GetUnpaidInvoices retourne les factures non payées d'un client
func GetUnpaidInvoices(db *sql.DB, clientID int) ([]models.Document, error) {
	rows, err := db.Query(`
		SELECT id, doc_number, date, net_amount, amount_paid, amount_remaining, status
		FROM documents
		WHERE client_id=? AND doc_type='FA' AND status IN ('confirmed','partial')
		AND amount_remaining > 0
		ORDER BY date ASC`, clientID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var docs []models.Document
	for rows.Next() {
		var d models.Document
		rows.Scan(&d.ID, &d.DocNumber, &d.Date, &d.NetAmount, &d.AmountPaid, &d.AmountRemaining, &d.Status)
		docs = append(docs, d)
	}
	return docs, nil
}
