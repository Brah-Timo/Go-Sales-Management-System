package models

import "time"

// Company représente l'entreprise/le commerce
type Company struct {
	ID            int       `db:"id"`
	NameAr        string    `db:"name_ar"`
	NameFr        string    `db:"name_fr"`
	Activity      string    `db:"activity"`
	Address       string    `db:"address"`
	Wilaya        string    `db:"wilaya"`
	Commune       string    `db:"commune"`
	PostalCode    string    `db:"postal_code"`
	Phone         string    `db:"phone"`
	Mobile        string    `db:"mobile"`
	Fax           string    `db:"fax"`
	Email         string    `db:"email"`
	Website       string    `db:"website"`
	NIF           string    `db:"nif"`
	NIS           string    `db:"nis"`
	RC            string    `db:"rc"`
	AI            string    `db:"ai"`
	RIB           string    `db:"rib"`
	BankName      string    `db:"bank_name"`
	Capital       float64   `db:"capital"`
	LogoPath      string    `db:"logo_path"`
	StampPath     string    `db:"stamp_path"`
	SignaturePath string    `db:"signature_path"`
	FooterText    string    `db:"footer_text"`
	CreatedAt     time.Time `db:"created_at"`
	UpdatedAt     time.Time `db:"updated_at"`
}

// FiscalYear représente une année fiscale
type FiscalYear struct {
	ID           int       `db:"id"`
	Year         int       `db:"year"`
	StartDate    string    `db:"start_date"`
	EndDate      string    `db:"end_date"`
	Status       string    `db:"status"`
	InvoiceCount int       `db:"invoice_count"`
	Revenue      float64   `db:"revenue"`
	CreatedAt    time.Time `db:"created_at"`
}

// Settings stocke les paramètres clé-valeur
type Setting struct {
	ID    int    `db:"id"`
	Key   string `db:"key"`
	Value string `db:"value"`
}

// TaxConfig contient la configuration fiscale
type TaxConfig struct {
	TVANormal          float64
	TVAReduced         float64
	TimbreRate         float64
	TimbreMax          float64
	TimbreExemption    float64
	TAPRate            float64
	IsTVASubject       bool
	TaxRegime          string
	AutoTimbre         bool
	AutoUpdateSalePrice bool
}
