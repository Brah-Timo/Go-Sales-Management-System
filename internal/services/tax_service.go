package services

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"gestion-commerciale/internal/models"
	"math"
)

// TaxService gère les calculs fiscaux et les déclarations
type TaxService struct {
	db *sql.DB
}

// NewTaxService crée un service fiscal
func NewTaxService(db *sql.DB) *TaxService {
	return &TaxService{db: db}
}

// GetTaxConfig retourne la configuration fiscale depuis les paramètres
func (s *TaxService) GetTaxConfig() models.TaxConfig {
	cfg := models.TaxConfig{
		TVANormal:     19,
		TVAReduced:    9,
		TimbreRate:    1,
		TimbreMax:     2500,
		TimbreExemption: 0,
		TAPRate:       1,
		IsTVASubject:  true,
		TaxRegime:     "real",
		AutoTimbre:    true,
	}

	rows, _ := s.db.Query(`SELECT key, value FROM settings WHERE key IN
		('tva_normal','tva_reduced','timbre_rate','timbre_max','timbre_exemption',
		 'tap_rate','is_tva_subject','tax_regime','auto_timbre','auto_update_sale_price')`)
	if rows != nil {
		defer rows.Close()
		for rows.Next() {
			var k, v string
			rows.Scan(&k, &v)
			switch k {
			case "tva_normal":
				fmt.Sscanf(v, "%f", &cfg.TVANormal)
			case "tva_reduced":
				fmt.Sscanf(v, "%f", &cfg.TVAReduced)
			case "timbre_rate":
				fmt.Sscanf(v, "%f", &cfg.TimbreRate)
			case "timbre_max":
				fmt.Sscanf(v, "%f", &cfg.TimbreMax)
			case "timbre_exemption":
				fmt.Sscanf(v, "%f", &cfg.TimbreExemption)
			case "tap_rate":
				fmt.Sscanf(v, "%f", &cfg.TAPRate)
			case "is_tva_subject":
				cfg.IsTVASubject = v == "1"
			case "tax_regime":
				cfg.TaxRegime = v
			case "auto_timbre":
				cfg.AutoTimbre = v == "1"
			case "auto_update_sale_price":
				cfg.AutoUpdateSalePrice = v == "1"
			}
		}
	}
	return cfg
}

// SaveTaxConfig sauvegarde la configuration fiscale
func (s *TaxService) SaveTaxConfig(cfg models.TaxConfig) error {
	settings := map[string]string{
		"tva_normal":           fmt.Sprintf("%.0f", cfg.TVANormal),
		"tva_reduced":          fmt.Sprintf("%.0f", cfg.TVAReduced),
		"timbre_rate":          fmt.Sprintf("%.2f", cfg.TimbreRate),
		"timbre_max":           fmt.Sprintf("%.2f", cfg.TimbreMax),
		"timbre_exemption":     fmt.Sprintf("%.2f", cfg.TimbreExemption),
		"tap_rate":             fmt.Sprintf("%.2f", cfg.TAPRate),
		"is_tva_subject":       boolToStr(cfg.IsTVASubject),
		"tax_regime":           cfg.TaxRegime,
		"auto_timbre":          boolToStr(cfg.AutoTimbre),
		"auto_update_sale_price": boolToStr(cfg.AutoUpdateSalePrice),
	}
	for k, v := range settings {
		_, err := s.db.Exec(`INSERT OR REPLACE INTO settings (key, value) VALUES (?,?)`, k, v)
		if err != nil {
			return err
		}
	}
	return nil
}

// CalculateG50 calcule les données du formulaire G50 mensuel
func (s *TaxService) CalculateG50(year, month int) (models.G50Data, error) {
	g50 := models.G50Data{
		Year:   year,
		Month:  month,
		Period: fmt.Sprintf("%d-%02d", year, month),
	}

	// Chiffre d'affaires par taux TVA
	err := s.db.QueryRow(`
		SELECT
		  COALESCE(SUM(CASE WHEN dl.tva_rate = 19 THEN dl.amount_ht ELSE 0 END), 0),
		  COALESCE(SUM(CASE WHEN dl.tva_rate = 9  THEN dl.amount_ht ELSE 0 END), 0),
		  COALESCE(SUM(CASE WHEN dl.tva_rate = 0  THEN dl.amount_ht ELSE 0 END), 0)
		FROM document_lines dl
		JOIN documents d ON dl.document_id = d.id
		WHERE d.doc_type = 'FA'
		  AND d.status IN ('confirmed','paid','partial')
		  AND strftime('%Y-%m', d.date) = ?`, g50.Period).
		Scan(&g50.Revenue19, &g50.Revenue9, &g50.RevenueExempt)
	if err != nil {
		return g50, fmt.Errorf("erreur calcul CA: %w", err)
	}

	g50.TotalRevenue = g50.Revenue19 + g50.Revenue9 + g50.RevenueExempt

	// TVA collectée
	g50.TVACollected19 = math.Round(g50.Revenue19*19/100*100) / 100
	g50.TVACollected9 = math.Round(g50.Revenue9*9/100*100) / 100
	g50.TotalTVACollected = g50.TVACollected19 + g50.TVACollected9

	// TVA déductible sur achats
	s.db.QueryRow(`
		SELECT COALESCE(SUM(dl.tva_amount), 0)
		FROM document_lines dl
		JOIN documents d ON dl.document_id = d.id
		WHERE d.doc_type = 'FAC'
		  AND d.status IN ('confirmed','paid','partial')
		  AND strftime('%Y-%m', d.date) = ?`, g50.Period).
		Scan(&g50.TVADeductiblePurchases)

	// TVA nette
	totalCollected := g50.TotalTVACollected
	totalDeductible := g50.TVADeductiblePurchases + g50.TVADeductibleInvestments + g50.TVAPrecompte

	if totalCollected > totalDeductible {
		g50.TVADue = math.Round((totalCollected-totalDeductible)*100) / 100
	} else {
		g50.TVAPrecompteNext = math.Round((totalDeductible-totalCollected)*100) / 100
	}

	// TAP (Taxe sur l'Activité Professionnelle)
	cfg := s.GetTaxConfig()
	g50.TAP = math.Round(g50.TotalRevenue*cfg.TAPRate/100*100) / 100

	// Droit de timbre
	s.db.QueryRow(`
		SELECT COALESCE(SUM(timbre), 0)
		FROM documents
		WHERE doc_type='FA'
		  AND status IN ('confirmed','paid','partial')
		  AND payment_method='cash'
		  AND strftime('%Y-%m', date) = ?`, g50.Period).
		Scan(&g50.Timbre)

	// Total à payer
	g50.TotalDue = math.Round((g50.TVADue+g50.TAP+g50.Timbre+g50.IRGSalaries)*100) / 100

	return g50, nil
}

// SaveG50 sauvegarde une déclaration G50
func (s *TaxService) SaveG50(g50 models.G50Data, status string) error {
	dataJSON, _ := json.Marshal(g50)

	var existing int
	s.db.QueryRow(`SELECT id FROM tax_declarations WHERE type='G50' AND year=? AND month=?`,
		g50.Year, g50.Month).Scan(&existing)

	if existing > 0 {
		_, err := s.db.Exec(`UPDATE tax_declarations SET data_json=?, status=? WHERE id=?`,
			string(dataJSON), status, existing)
		return err
	}

	_, err := s.db.Exec(`
		INSERT INTO tax_declarations (type, year, month, data_json, status)
		VALUES ('G50', ?, ?, ?, ?)`,
		g50.Year, g50.Month, string(dataJSON), status)
	return err
}

// GetG50History retourne l'historique des déclarations G50
func (s *TaxService) GetG50History(year int) ([]models.TaxDeclaration, error) {
	rows, err := s.db.Query(`
		SELECT id, type, year, month, data_json, status, created_at
		FROM tax_declarations WHERE type='G50' AND year=?
		ORDER BY month DESC`, year)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var decls []models.TaxDeclaration
	for rows.Next() {
		var d models.TaxDeclaration
		rows.Scan(&d.ID, &d.Type, &d.Year, &d.Month, &d.DataJSON, &d.Status, &d.CreatedAt)
		decls = append(decls, d)
	}
	return decls, nil
}

// GetTVASalesRegister retourne le registre TVA des ventes
func (s *TaxService) GetTVASalesRegister(year, month int) ([]map[string]interface{}, error) {
	period := fmt.Sprintf("%d-%02d", year, month)
	rows, err := s.db.Query(`
		SELECT d.doc_number, d.date,
		  COALESCE(c.name_fr,'Client de passage') as client_name,
		  COALESCE(c.nif,'') as nif,
		  d.net_ht, d.total_tva, d.total_ttc, d.timbre
		FROM documents d
		LEFT JOIN clients c ON d.client_id=c.id
		WHERE d.doc_type='FA'
		  AND d.status IN ('confirmed','paid','partial')
		  AND strftime('%Y-%m', d.date)=?
		ORDER BY d.date`, period)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []map[string]interface{}
	for rows.Next() {
		row := make(map[string]interface{})
		var docNum, date, clientName, nif string
		var netHT, totalTVA, totalTTC, timbre float64
		rows.Scan(&docNum, &date, &clientName, &nif, &netHT, &totalTVA, &totalTTC, &timbre)
		row["doc_number"] = docNum
		row["date"] = date
		row["client_name"] = clientName
		row["nif"] = nif
		row["net_ht"] = netHT
		row["total_tva"] = totalTVA
		row["total_ttc"] = totalTTC
		row["timbre"] = timbre
		result = append(result, row)
	}
	return result, nil
}

// GetTVAPurchasesRegister retourne le registre TVA des achats
func (s *TaxService) GetTVAPurchasesRegister(year, month int) ([]map[string]interface{}, error) {
	period := fmt.Sprintf("%d-%02d", year, month)
	rows, err := s.db.Query(`
		SELECT d.doc_number, d.supplier_invoice_number, d.date,
		  COALESCE(sup.name_fr,'') as supplier_name,
		  COALESCE(sup.nif,'') as nif,
		  d.net_ht, d.total_tva, d.total_ttc
		FROM documents d
		LEFT JOIN suppliers sup ON d.supplier_id=sup.id
		WHERE d.doc_type='FAC'
		  AND d.status IN ('confirmed','paid','partial')
		  AND strftime('%Y-%m', d.date)=?
		ORDER BY d.date`, period)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []map[string]interface{}
	for rows.Next() {
		row := make(map[string]interface{})
		var docNum, suppInvNum, date, supplierName, nif string
		var netHT, totalTVA, totalTTC float64
		rows.Scan(&docNum, &suppInvNum, &date, &supplierName, &nif, &netHT, &totalTVA, &totalTTC)
		row["doc_number"] = docNum
		row["supplier_invoice_number"] = suppInvNum
		row["date"] = date
		row["supplier_name"] = supplierName
		row["nif"] = nif
		row["net_ht"] = netHT
		row["total_tva"] = totalTVA
		row["total_ttc"] = totalTTC
		result = append(result, row)
	}
	return result, nil
}

func boolToStr(b bool) string {
	if b {
		return "1"
	}
	return "0"
}
