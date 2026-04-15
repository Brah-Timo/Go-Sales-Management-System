package services

import (
	"database/sql"
	"fmt"
	"gestion-commerciale/internal/models"
)

// ClientService gère les opérations sur les clients et fournisseurs
type ClientService struct {
	db *sql.DB
}

// NewClientService crée un service clients
func NewClientService(db *sql.DB) *ClientService {
	return &ClientService{db: db}
}

// ─────────────────────────────────────────────────────────────────────────────
// CLIENTS
// ─────────────────────────────────────────────────────────────────────────────

// GetAllClients retourne tous les clients
func (s *ClientService) GetAllClients(search string) ([]models.Client, error) {
	query := `
		SELECT c.id, c.code, c.name_ar, c.name_fr, c.type, c.address, c.wilaya,
		  c.commune, c.phone, c.mobile, c.fax, c.email, c.nif, c.nis, c.rc, c.ai,
		  c.price_list_id, c.credit_limit, c.payment_terms, c.discount_rate,
		  c.balance, c.is_blocked, c.notes,
		  COALESCE(pl.name,'') as price_list_name
		FROM clients c
		LEFT JOIN price_lists pl ON c.price_list_id=pl.id`

	args := []interface{}{}
	if search != "" {
		query += ` WHERE (c.name_fr LIKE ? OR c.name_ar LIKE ? OR c.code LIKE ? OR c.phone LIKE ?)`
		s := "%" + search + "%"
		args = append(args, s, s, s, s)
	}
	query += ` ORDER BY c.name_fr`

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var clients []models.Client
	for rows.Next() {
		var c models.Client
		rows.Scan(&c.ID, &c.Code, &c.NameAr, &c.NameFr, &c.Type, &c.Address, &c.Wilaya,
			&c.Commune, &c.Phone, &c.Mobile, &c.Fax, &c.Email, &c.NIF, &c.NIS, &c.RC, &c.AI,
			&c.PriceListID, &c.CreditLimit, &c.PaymentTerms, &c.DiscountRate,
			&c.Balance, &c.IsBlocked, &c.Notes, &c.PriceListName)
		clients = append(clients, c)
	}
	return clients, nil
}

// GetClientByID retourne un client par son ID
func (s *ClientService) GetClientByID(id int) (*models.Client, error) {
	var c models.Client
	err := s.db.QueryRow(`
		SELECT c.id, c.code, c.name_ar, c.name_fr, c.type, c.address, c.wilaya,
		  c.commune, c.phone, c.mobile, c.fax, c.email, c.nif, c.nis, c.rc, c.ai,
		  c.price_list_id, c.credit_limit, c.payment_terms, c.discount_rate,
		  c.balance, c.is_blocked, c.notes,
		  COALESCE(pl.name,'') as price_list_name
		FROM clients c
		LEFT JOIN price_lists pl ON c.price_list_id=pl.id
		WHERE c.id=?`, id).
		Scan(&c.ID, &c.Code, &c.NameAr, &c.NameFr, &c.Type, &c.Address, &c.Wilaya,
			&c.Commune, &c.Phone, &c.Mobile, &c.Fax, &c.Email, &c.NIF, &c.NIS, &c.RC, &c.AI,
			&c.PriceListID, &c.CreditLimit, &c.PaymentTerms, &c.DiscountRate,
			&c.Balance, &c.IsBlocked, &c.Notes, &c.PriceListName)
	if err != nil {
		return nil, fmt.Errorf("client introuvable: %w", err)
	}
	return &c, nil
}

// SaveClient crée ou met à jour un client
func (s *ClientService) SaveClient(c *models.Client, userID int) error {
	if c.Code == "" {
		c.Code = s.generateClientCode()
	}
	if c.ID == 0 {
		res, err := s.db.Exec(`
			INSERT INTO clients (code, name_ar, name_fr, type, address, wilaya, commune,
			  phone, mobile, fax, email, nif, nis, rc, ai, price_list_id,
			  credit_limit, payment_terms, discount_rate, balance, is_blocked, notes)
			VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,0,0,?)`,
			c.Code, c.NameAr, c.NameFr, c.Type, c.Address, c.Wilaya, c.Commune,
			c.Phone, c.Mobile, c.Fax, c.Email, c.NIF, c.NIS, c.RC, c.AI, c.PriceListID,
			c.CreditLimit, c.PaymentTerms, c.DiscountRate, c.Notes)
		if err != nil {
			return fmt.Errorf("erreur création client: %w", err)
		}
		id, _ := res.LastInsertId()
		c.ID = int(id)

		s.db.Exec(`INSERT INTO audit_log (user_id, action_type, module, description) VALUES (?,?,?,?)`,
			userID, "create", "clients", "Création client "+c.Code)
	} else {
		_, err := s.db.Exec(`
			UPDATE clients SET name_ar=?, name_fr=?, type=?, address=?, wilaya=?, commune=?,
			  phone=?, mobile=?, fax=?, email=?, nif=?, nis=?, rc=?, ai=?, price_list_id=?,
			  credit_limit=?, payment_terms=?, discount_rate=?, is_blocked=?, notes=?
			WHERE id=?`,
			c.NameAr, c.NameFr, c.Type, c.Address, c.Wilaya, c.Commune,
			c.Phone, c.Mobile, c.Fax, c.Email, c.NIF, c.NIS, c.RC, c.AI, c.PriceListID,
			c.CreditLimit, c.PaymentTerms, c.DiscountRate, c.IsBlocked, c.Notes, c.ID)
		if err != nil {
			return fmt.Errorf("erreur mise à jour client: %w", err)
		}
		s.db.Exec(`INSERT INTO audit_log (user_id, action_type, module, description) VALUES (?,?,?,?)`,
			userID, "update", "clients", "Modification client "+c.Code)
	}
	return nil
}

// DeleteClient supprime un client (vérification des documents)
func (s *ClientService) DeleteClient(clientID, userID int) error {
	var count int
	s.db.QueryRow(`SELECT COUNT(*) FROM documents WHERE client_id=?`, clientID).Scan(&count)
	if count > 0 {
		return fmt.Errorf("impossible de supprimer: le client a %d document(s)", count)
	}
	_, err := s.db.Exec(`DELETE FROM clients WHERE id=?`, clientID)
	if err != nil {
		return err
	}
	s.db.Exec(`INSERT INTO audit_log (user_id, action_type, module, description) VALUES (?,?,?,?)`,
		userID, "delete", "clients", fmt.Sprintf("Suppression client ID %d", clientID))
	return nil
}

// ToggleBlockClient bloque/débloque un client
func (s *ClientService) ToggleBlockClient(clientID int, blocked bool) error {
	val := 0
	if blocked {
		val = 1
	}
	_, err := s.db.Exec(`UPDATE clients SET is_blocked=? WHERE id=?`, val, clientID)
	return err
}

// GetClientStatement retourne le relevé de compte d'un client
func (s *ClientService) GetClientStatement(clientID int, from, to string) ([]models.AccountStatement, error) {
	rows, err := s.db.Query(`
		SELECT d.date, d.doc_type, d.doc_number,
		  CASE WHEN d.doc_type IN ('FA','BL') THEN 'Facture/BL' ELSE d.doc_type END as description,
		  CASE WHEN d.doc_type IN ('FA','BL') THEN d.net_amount ELSE 0 END as debit,
		  0 as credit,
		  0 as balance
		FROM documents d
		WHERE d.client_id=? AND d.status IN ('confirmed','paid','partial')
		  AND date(d.date) BETWEEN ? AND ?
		UNION ALL
		SELECT p.date, 'REGL', 'REGL-'||p.id, 'Règlement', 0, p.amount, 0
		FROM payments p
		WHERE p.client_id=? AND p.type='collection'
		  AND date(p.date) BETWEEN ? AND ?
		ORDER BY 1, 2`,
		clientID, from, to, clientID, from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var lines []models.AccountStatement
	var runningBalance float64
	for rows.Next() {
		var l models.AccountStatement
		rows.Scan(&l.Date, &l.DocType, &l.DocNumber, &l.Description, &l.Debit, &l.Credit, &l.Balance)
		runningBalance += l.Debit - l.Credit
		l.Balance = runningBalance
		lines = append(lines, l)
	}
	return lines, nil
}

// GetClientNames retourne la liste des noms de clients pour les selects
func (s *ClientService) GetClientNames() []string {
	rows, _ := s.db.Query(`SELECT name_fr FROM clients WHERE is_blocked=0 ORDER BY name_fr`)
	if rows == nil {
		return nil
	}
	defer rows.Close()
	var names []string
	for rows.Next() {
		var n string
		rows.Scan(&n)
		names = append(names, n)
	}
	return names
}

// generateClientCode génère un code client unique
func (s *ClientService) generateClientCode() string {
	var maxCode string
	s.db.QueryRow(`SELECT COALESCE(MAX(code),'C-0000') FROM clients WHERE code LIKE 'C-%'`).Scan(&maxCode)
	var num int
	fmt.Sscanf(maxCode[2:], "%d", &num)
	num++
	return fmt.Sprintf("C-%04d", num)
}

// ─────────────────────────────────────────────────────────────────────────────
// FOURNISSEURS
// ─────────────────────────────────────────────────────────────────────────────

// GetAllSuppliers retourne tous les fournisseurs
func (s *ClientService) GetAllSuppliers(search string) ([]models.Supplier, error) {
	query := `
		SELECT id, code, name_ar, name_fr, address, wilaya, phone, mobile, fax, email,
		  nif, nis, rc, ai, payment_terms, balance,
		  rating_delivery, rating_quality, rating_pricing, notes
		FROM suppliers`
	args := []interface{}{}
	if search != "" {
		query += ` WHERE (name_fr LIKE ? OR name_ar LIKE ? OR code LIKE ? OR phone LIKE ?)`
		ss := "%" + search + "%"
		args = append(args, ss, ss, ss, ss)
	}
	query += ` ORDER BY name_fr`

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var suppliers []models.Supplier
	for rows.Next() {
		var sup models.Supplier
		rows.Scan(&sup.ID, &sup.Code, &sup.NameAr, &sup.NameFr, &sup.Address, &sup.Wilaya,
			&sup.Phone, &sup.Mobile, &sup.Fax, &sup.Email,
			&sup.NIF, &sup.NIS, &sup.RC, &sup.AI, &sup.PaymentTerms, &sup.Balance,
			&sup.RatingDelivery, &sup.RatingQuality, &sup.RatingPricing, &sup.Notes)
		suppliers = append(suppliers, sup)
	}
	return suppliers, nil
}

// GetSupplierByID retourne un fournisseur par son ID
func (s *ClientService) GetSupplierByID(id int) (*models.Supplier, error) {
	var sup models.Supplier
	err := s.db.QueryRow(`
		SELECT id, code, name_ar, name_fr, address, wilaya, phone, mobile, fax, email,
		  nif, nis, rc, ai, payment_terms, balance,
		  rating_delivery, rating_quality, rating_pricing, notes
		FROM suppliers WHERE id=?`, id).
		Scan(&sup.ID, &sup.Code, &sup.NameAr, &sup.NameFr, &sup.Address, &sup.Wilaya,
			&sup.Phone, &sup.Mobile, &sup.Fax, &sup.Email,
			&sup.NIF, &sup.NIS, &sup.RC, &sup.AI, &sup.PaymentTerms, &sup.Balance,
			&sup.RatingDelivery, &sup.RatingQuality, &sup.RatingPricing, &sup.Notes)
	if err != nil {
		return nil, fmt.Errorf("fournisseur introuvable: %w", err)
	}
	return &sup, nil
}

// SaveSupplier crée ou met à jour un fournisseur
func (s *ClientService) SaveSupplier(sup *models.Supplier, userID int) error {
	if sup.Code == "" {
		sup.Code = s.generateSupplierCode()
	}
	if sup.ID == 0 {
		res, err := s.db.Exec(`
			INSERT INTO suppliers (code, name_ar, name_fr, address, wilaya, phone, mobile,
			  fax, email, nif, nis, rc, ai, payment_terms, balance,
			  rating_delivery, rating_quality, rating_pricing, notes)
			VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,0,?,?,?,?)`,
			sup.Code, sup.NameAr, sup.NameFr, sup.Address, sup.Wilaya, sup.Phone, sup.Mobile,
			sup.Fax, sup.Email, sup.NIF, sup.NIS, sup.RC, sup.AI, sup.PaymentTerms,
			sup.RatingDelivery, sup.RatingQuality, sup.RatingPricing, sup.Notes)
		if err != nil {
			return fmt.Errorf("erreur création fournisseur: %w", err)
		}
		id, _ := res.LastInsertId()
		sup.ID = int(id)
		s.db.Exec(`INSERT INTO audit_log (user_id, action_type, module, description) VALUES (?,?,?,?)`,
			userID, "create", "suppliers", "Création fournisseur "+sup.Code)
	} else {
		_, err := s.db.Exec(`
			UPDATE suppliers SET name_ar=?, name_fr=?, address=?, wilaya=?, phone=?, mobile=?,
			  fax=?, email=?, nif=?, nis=?, rc=?, ai=?, payment_terms=?,
			  rating_delivery=?, rating_quality=?, rating_pricing=?, notes=?
			WHERE id=?`,
			sup.NameAr, sup.NameFr, sup.Address, sup.Wilaya, sup.Phone, sup.Mobile,
			sup.Fax, sup.Email, sup.NIF, sup.NIS, sup.RC, sup.AI, sup.PaymentTerms,
			sup.RatingDelivery, sup.RatingQuality, sup.RatingPricing, sup.Notes, sup.ID)
		if err != nil {
			return fmt.Errorf("erreur mise à jour fournisseur: %w", err)
		}
		s.db.Exec(`INSERT INTO audit_log (user_id, action_type, module, description) VALUES (?,?,?,?)`,
			userID, "update", "suppliers", "Modification fournisseur "+sup.Code)
	}
	return nil
}

// DeleteSupplier supprime un fournisseur
func (s *ClientService) DeleteSupplier(supplierID, userID int) error {
	var count int
	s.db.QueryRow(`SELECT COUNT(*) FROM documents WHERE supplier_id=?`, supplierID).Scan(&count)
	if count > 0 {
		return fmt.Errorf("impossible de supprimer: le fournisseur a %d document(s)", count)
	}
	_, err := s.db.Exec(`DELETE FROM suppliers WHERE id=?`, supplierID)
	return err
}

// GetSupplierNames retourne les noms de fournisseurs
func (s *ClientService) GetSupplierNames() []string {
	rows, _ := s.db.Query(`SELECT name_fr FROM suppliers ORDER BY name_fr`)
	if rows == nil {
		return nil
	}
	defer rows.Close()
	var names []string
	for rows.Next() {
		var n string
		rows.Scan(&n)
		names = append(names, n)
	}
	return names
}

func (s *ClientService) generateSupplierCode() string {
	var maxCode string
	s.db.QueryRow(`SELECT COALESCE(MAX(code),'F-0000') FROM suppliers WHERE code LIKE 'F-%'`).Scan(&maxCode)
	var num int
	fmt.Sscanf(maxCode[2:], "%d", &num)
	num++
	return fmt.Sprintf("F-%04d", num)
}

// ─────────────────────────────────────────────────────────────────────────────
// CHAUFFEURS
// ─────────────────────────────────────────────────────────────────────────────

// GetAllDrivers retourne tous les chauffeurs
func (s *ClientService) GetAllDrivers() ([]models.Driver, error) {
	rows, err := s.db.Query(`
		SELECT d.id, d.name, d.phone, d.vehicle_plate,
		  (SELECT COUNT(*) FROM documents WHERE driver_id=d.id AND doc_type='BL') as delivery_count
		FROM drivers d ORDER BY d.name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var drivers []models.Driver
	for rows.Next() {
		var d models.Driver
		rows.Scan(&d.ID, &d.Name, &d.Phone, &d.VehiclePlate, &d.DeliveryCount)
		drivers = append(drivers, d)
	}
	return drivers, nil
}

// SaveDriver sauvegarde un chauffeur
func (s *ClientService) SaveDriver(d *models.Driver) error {
	if d.ID == 0 {
		res, err := s.db.Exec(`INSERT INTO drivers (name, phone, vehicle_plate) VALUES (?,?,?)`,
			d.Name, d.Phone, d.VehiclePlate)
		if err != nil {
			return err
		}
		id, _ := res.LastInsertId()
		d.ID = int(id)
		return nil
	}
	_, err := s.db.Exec(`UPDATE drivers SET name=?, phone=?, vehicle_plate=? WHERE id=?`,
		d.Name, d.Phone, d.VehiclePlate, d.ID)
	return err
}

// DeleteDriver supprime un chauffeur
func (s *ClientService) DeleteDriver(driverID int) error {
	var count int
	s.db.QueryRow(`SELECT COUNT(*) FROM documents WHERE driver_id=?`, driverID).Scan(&count)
	if count > 0 {
		return fmt.Errorf("impossible de supprimer: le chauffeur a %d livraison(s)", count)
	}
	_, err := s.db.Exec(`DELETE FROM drivers WHERE id=?`, driverID)
	return err
}

// GetDriverNames retourne les noms de chauffeurs
func (s *ClientService) GetDriverNames() []string {
	rows, _ := s.db.Query(`SELECT name FROM drivers ORDER BY name`)
	if rows == nil {
		return nil
	}
	defer rows.Close()
	var names []string
	for rows.Next() {
		var n string
		rows.Scan(&n)
		names = append(names, n)
	}
	return names
}
