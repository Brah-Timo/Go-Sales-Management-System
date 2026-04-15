package queries

import (
	"database/sql"
	"fmt"
	"gestion-commerciale/internal/models"
)

// GetDocumentByID retourne un document complet avec ses lignes
func GetDocumentByID(db *sql.DB, id int) (*models.Document, error) {
	var doc models.Document
	err := db.QueryRow(`
		SELECT d.id, d.doc_type, d.doc_number, d.date,
		       d.fiscal_year_id, d.client_id, d.supplier_id, d.warehouse_id,
		       d.payment_method, d.payment_terms, d.price_list_id,
		       d.total_ht, d.total_discount, d.global_discount_pct, d.net_ht,
		       d.total_tva, d.total_ttc, d.timbre, d.net_amount,
		       d.amount_paid, d.amount_remaining,
		       d.status, d.notes, d.driver_id, d.delivery_address,
		       d.supplier_invoice_number, d.validity_days, d.source_doc_id,
		       d.created_by,
		       COALESCE(cl.name_fr,'') as client_name,
		       COALESCE(su.name_fr,'') as supplier_name,
		       COALESCE(dr.name,'') as driver_name
		FROM documents d
		LEFT JOIN clients cl   ON d.client_id = cl.id
		LEFT JOIN suppliers su ON d.supplier_id = su.id
		LEFT JOIN drivers dr   ON d.driver_id = dr.id
		WHERE d.id = ?`, id).Scan(
		&doc.ID, &doc.DocType, &doc.DocNumber, &doc.Date,
		&doc.FiscalYearID, &doc.ClientID, &doc.SupplierID, &doc.WarehouseID,
		&doc.PaymentMethod, &doc.PaymentTerms, &doc.PriceListID,
		&doc.TotalHT, &doc.TotalDiscount, &doc.GlobalDiscountPct, &doc.NetHT,
		&doc.TotalTVA, &doc.TotalTTC, &doc.Timbre, &doc.NetAmount,
		&doc.AmountPaid, &doc.AmountRemaining,
		&doc.Status, &doc.Notes, &doc.DriverID, &doc.DeliveryAddress,
		&doc.SupplierInvoiceNumber, &doc.ValidityDays, &doc.SourceDocID,
		&doc.CreatedBy,
		&doc.ClientName, &doc.SupplierName, &doc.DriverName,
	)
	if err != nil {
		return nil, fmt.Errorf("GetDocumentByID(%d): %w", id, err)
	}

	// Charger les lignes
	lines, err := GetDocumentLines(db, id)
	if err != nil {
		return nil, err
	}
	doc.Lines = lines

	return &doc, nil
}

// GetDocumentLines retourne les lignes d'un document
func GetDocumentLines(db *sql.DB, docID int) ([]models.DocumentLine, error) {
	rows, err := db.Query(`
		SELECT dl.id, dl.document_id, dl.line_number, dl.article_id,
		       dl.designation, dl.quantity, dl.unit, dl.unit_price_ht,
		       dl.discount_percent, dl.discount_amount, dl.amount_ht,
		       dl.tva_rate, dl.tva_amount, dl.amount_ttc,
		       dl.lot_number, COALESCE(dl.expiry_date,''),
		       COALESCE(a.reference,''), COALESCE(a.barcode,''), COALESCE(a.cmup,0)
		FROM document_lines dl
		LEFT JOIN articles a ON dl.article_id = a.id
		WHERE dl.document_id = ?
		ORDER BY dl.line_number`, docID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var lines []models.DocumentLine
	for rows.Next() {
		var l models.DocumentLine
		rows.Scan(
			&l.ID, &l.DocumentID, &l.LineNumber, &l.ArticleID,
			&l.Designation, &l.Quantity, &l.Unit, &l.UnitPriceHT,
			&l.DiscountPercent, &l.DiscountAmount, &l.AmountHT,
			&l.TVARate, &l.TVAAmount, &l.AmountTTC,
			&l.LotNumber, &l.ExpiryDate,
			&l.Reference, &l.Barcode, &l.CMUP,
		)
		lines = append(lines, l)
	}
	return lines, nil
}

// GetDocumentsList retourne la liste des documents avec filtres
func GetDocumentsList(db *sql.DB, docType string, filters map[string]string) ([]models.Document, error) {
	query := `
		SELECT d.id, d.doc_type, d.doc_number, d.date,
		       d.total_ht, d.total_tva, d.total_ttc, d.timbre, d.net_amount,
		       d.amount_paid, d.amount_remaining,
		       d.payment_method, d.status,
		       COALESCE(cl.name_fr,'') as client_name,
		       COALESCE(su.name_fr,'') as supplier_name
		FROM documents d
		LEFT JOIN clients cl   ON d.client_id = cl.id
		LEFT JOIN suppliers su ON d.supplier_id = su.id
		WHERE d.doc_type = ?`

	args := []interface{}{docType}

	if v, ok := filters["date_from"]; ok && v != "" {
		query += ` AND d.date >= ?`
		args = append(args, v)
	}
	if v, ok := filters["date_to"]; ok && v != "" {
		query += ` AND d.date <= ?`
		args = append(args, v+" 23:59:59")
	}
	if v, ok := filters["client_id"]; ok && v != "" {
		query += ` AND d.client_id = ?`
		args = append(args, v)
	}
	if v, ok := filters["supplier_id"]; ok && v != "" {
		query += ` AND d.supplier_id = ?`
		args = append(args, v)
	}
	if v, ok := filters["status"]; ok && v != "" {
		query += ` AND d.status = ?`
		args = append(args, v)
	}
	if v, ok := filters["payment_method"]; ok && v != "" {
		query += ` AND d.payment_method = ?`
		args = append(args, v)
	}
	if v, ok := filters["search"]; ok && v != "" {
		query += ` AND (d.doc_number LIKE ? OR cl.name_fr LIKE ? OR su.name_fr LIKE ?)`
		s := "%" + v + "%"
		args = append(args, s, s, s)
	}

	query += ` ORDER BY d.date DESC, d.id DESC`

	if v, ok := filters["limit"]; ok && v != "" {
		query += ` LIMIT ` + v
	}

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("GetDocumentsList: %w", err)
	}
	defer rows.Close()

	var docs []models.Document
	for rows.Next() {
		var d models.Document
		rows.Scan(
			&d.ID, &d.DocType, &d.DocNumber, &d.Date,
			&d.TotalHT, &d.TotalTVA, &d.TotalTTC, &d.Timbre, &d.NetAmount,
			&d.AmountPaid, &d.AmountRemaining,
			&d.PaymentMethod, &d.Status,
			&d.ClientName, &d.SupplierName,
		)
		docs = append(docs, d)
	}
	return docs, nil
}

// GenerateDocNumber génère le numéro de document suivant
func GenerateDocNumber(db *sql.DB, docType string, year int) (string, error) {
	tx, err := db.Begin()
	if err != nil {
		return "", err
	}
	defer tx.Rollback()

	var prefix string
	var currentNum int
	var resetYearly bool

	err = tx.QueryRow(`
		SELECT prefix, current_number, reset_yearly 
		FROM numbering_config WHERE doc_type=?`, docType).Scan(&prefix, &currentNum, &resetYearly)
	if err != nil {
		return "", fmt.Errorf("GenerateDocNumber: config non trouvée pour %s", docType)
	}

	nextNum := currentNum + 1
	docNumber := fmt.Sprintf("%s-%d-%04d", prefix, year, nextNum)

	_, err = tx.Exec(`UPDATE numbering_config SET current_number=? WHERE doc_type=?`, nextNum, docType)
	if err != nil {
		return "", err
	}

	return docNumber, tx.Commit()
}

// SaveDocument insère ou met à jour un document (sans ses lignes)
func SaveDocument(db *sql.DB, doc *models.Document) error {
	if doc.ID == 0 {
		result, err := db.Exec(`
			INSERT INTO documents (doc_type, doc_number, date, fiscal_year_id,
			  client_id, supplier_id, warehouse_id, payment_method, payment_terms,
			  price_list_id, total_ht, total_discount, global_discount_pct, net_ht,
			  total_tva, total_ttc, timbre, net_amount, amount_paid, amount_remaining,
			  status, notes, driver_id, delivery_address, supplier_invoice_number,
			  validity_days, source_doc_id, created_by)
			VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`,
			doc.DocType, doc.DocNumber, doc.Date, doc.FiscalYearID,
			doc.ClientID, doc.SupplierID, doc.WarehouseID, doc.PaymentMethod, doc.PaymentTerms,
			doc.PriceListID, doc.TotalHT, doc.TotalDiscount, doc.GlobalDiscountPct, doc.NetHT,
			doc.TotalTVA, doc.TotalTTC, doc.Timbre, doc.NetAmount, doc.AmountPaid, doc.AmountRemaining,
			doc.Status, doc.Notes, doc.DriverID, doc.DeliveryAddress, doc.SupplierInvoiceNumber,
			doc.ValidityDays, doc.SourceDocID, doc.CreatedBy,
		)
		if err != nil {
			return fmt.Errorf("SaveDocument insert: %w", err)
		}
		id, _ := result.LastInsertId()
		doc.ID = int(id)
	} else {
		_, err := db.Exec(`
			UPDATE documents SET date=?, client_id=?, supplier_id=?, warehouse_id=?,
			  payment_method=?, payment_terms=?, price_list_id=?,
			  total_ht=?, total_discount=?, global_discount_pct=?, net_ht=?,
			  total_tva=?, total_ttc=?, timbre=?, net_amount=?,
			  amount_paid=?, amount_remaining=?, status=?, notes=?,
			  driver_id=?, delivery_address=?, supplier_invoice_number=?,
			  validity_days=?, source_doc_id=?
			WHERE id=?`,
			doc.Date, doc.ClientID, doc.SupplierID, doc.WarehouseID,
			doc.PaymentMethod, doc.PaymentTerms, doc.PriceListID,
			doc.TotalHT, doc.TotalDiscount, doc.GlobalDiscountPct, doc.NetHT,
			doc.TotalTVA, doc.TotalTTC, doc.Timbre, doc.NetAmount,
			doc.AmountPaid, doc.AmountRemaining, doc.Status, doc.Notes,
			doc.DriverID, doc.DeliveryAddress, doc.SupplierInvoiceNumber,
			doc.ValidityDays, doc.SourceDocID,
			doc.ID,
		)
		if err != nil {
			return fmt.Errorf("SaveDocument update: %w", err)
		}
	}
	return nil
}

// SaveDocumentLines supprime les anciennes lignes et insère les nouvelles
func SaveDocumentLines(db *sql.DB, docID int, lines []models.DocumentLine) error {
	// Supprimer les anciennes lignes
	_, err := db.Exec(`DELETE FROM document_lines WHERE document_id=?`, docID)
	if err != nil {
		return err
	}

	// Insérer les nouvelles
	for i, l := range lines {
		_, err = db.Exec(`
			INSERT INTO document_lines (document_id, line_number, article_id,
			  designation, quantity, unit, unit_price_ht, discount_percent,
			  discount_amount, amount_ht, tva_rate, tva_amount, amount_ttc,
			  lot_number, expiry_date)
			VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`,
			docID, i+1, l.ArticleID,
			l.Designation, l.Quantity, l.Unit, l.UnitPriceHT, l.DiscountPercent,
			l.DiscountAmount, l.AmountHT, l.TVARate, l.TVAAmount, l.AmountTTC,
			l.LotNumber, l.ExpiryDate,
		)
		if err != nil {
			return fmt.Errorf("SaveDocumentLines line %d: %w", i+1, err)
		}
	}
	return nil
}

// GetRecentDocuments retourne les derniers documents (toutes types)
func GetRecentDocuments(db *sql.DB, limit int) ([]models.Document, error) {
	rows, err := db.Query(`
		SELECT d.id, d.doc_type, d.doc_number, d.date, d.net_amount, d.status,
		       COALESCE(cl.name_fr, su.name_fr, 'N/A') as party_name
		FROM documents d
		LEFT JOIN clients cl   ON d.client_id = cl.id
		LEFT JOIN suppliers su ON d.supplier_id = su.id
		ORDER BY d.created_at DESC LIMIT ?`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var docs []models.Document
	for rows.Next() {
		var d models.Document
		rows.Scan(&d.ID, &d.DocType, &d.DocNumber, &d.Date, &d.NetAmount, &d.Status, &d.ClientName)
		docs = append(docs, d)
	}
	return docs, nil
}
