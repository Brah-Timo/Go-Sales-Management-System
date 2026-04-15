package services

import (
	"database/sql"
	"fmt"
	"gestion-commerciale/internal/database/queries"
)

// NumberingService gère la numérotation des documents
type NumberingService struct {
	db *sql.DB
}

// NewNumberingService crée un service de numérotation
func NewNumberingService(db *sql.DB) *NumberingService {
	return &NumberingService{db: db}
}

// NextNumber génère le prochain numéro pour un type de document
func (s *NumberingService) NextNumber(docType string, year int) (string, error) {
	return queries.GenerateDocNumber(s.db, docType, year)
}

// ResetYearly remet à zéro les compteurs pour une nouvelle année
func (s *NumberingService) ResetYearly() error {
	_, err := s.db.Exec(`UPDATE numbering_config SET current_number=0 WHERE reset_yearly=1`)
	return err
}

// GetAllConfigs retourne toutes les configurations de numérotation
func (s *NumberingService) GetAllConfigs() ([]struct {
	DocType       string
	Prefix        string
	CurrentNumber int
	ResetYearly   bool
	Label         string
}, error) {
	rows, err := s.db.Query(`SELECT doc_type, prefix, current_number, reset_yearly FROM numbering_config ORDER BY doc_type`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	labels := map[string]string{
		"FA":  "Facture de Vente",
		"FAC": "Facture d'Achat",
		"BL":  "Bon de Livraison",
		"BR":  "Bon de Réception",
		"DV":  "Devis",
		"PF":  "Proforma",
		"BCC": "Bon de Commande Client",
		"BCF": "Bon de Commande Fournisseur",
		"AV":  "Avoir",
		"BP":  "Bon de Paiement",
		"BRE": "Bon de Réception Espèces",
	}

	var configs []struct {
		DocType       string
		Prefix        string
		CurrentNumber int
		ResetYearly   bool
		Label         string
	}
	for rows.Next() {
		var c struct {
			DocType       string
			Prefix        string
			CurrentNumber int
			ResetYearly   bool
			Label         string
		}
		rows.Scan(&c.DocType, &c.Prefix, &c.CurrentNumber, &c.ResetYearly)
		c.Label = labels[c.DocType]
		configs = append(configs, c)
	}
	return configs, nil
}

// UpdateConfig met à jour la configuration d'un type de document
func (s *NumberingService) UpdateConfig(docType, prefix string, startNum int, resetYearly bool) error {
	resetVal := 0
	if resetYearly { resetVal = 1 }
	_, err := s.db.Exec(`
		UPDATE numbering_config SET prefix=?, current_number=?, reset_yearly=?
		WHERE doc_type=?`,
		prefix, startNum-1, resetVal, docType)
	return err
}

// PreviewNextNumber affiche un aperçu du prochain numéro
func (s *NumberingService) PreviewNextNumber(docType, prefix string, currentNum, year int) string {
	return fmt.Sprintf("%s-%d-%04d", prefix, year, currentNum+1)
}
