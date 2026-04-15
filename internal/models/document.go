package models

import "time"

// Document représente tout type de document commercial (FA, FAC, BL, BR, DV...)
type Document struct {
	ID                      int       `db:"id"`
	DocType                 string    `db:"doc_type"`
	DocNumber               string    `db:"doc_number"`
	Date                    string    `db:"date"`
	FiscalYearID            *int      `db:"fiscal_year_id"`
	ClientID                *int      `db:"client_id"`
	SupplierID              *int      `db:"supplier_id"`
	WarehouseID             *int      `db:"warehouse_id"`
	PaymentMethod           string    `db:"payment_method"`
	PaymentTerms            string    `db:"payment_terms"`
	PriceListID             *int      `db:"price_list_id"`
	TotalHT                 float64   `db:"total_ht"`
	TotalDiscount           float64   `db:"total_discount"`
	GlobalDiscountPct       float64   `db:"global_discount_pct"`
	NetHT                   float64   `db:"net_ht"`
	TotalTVA                float64   `db:"total_tva"`
	TotalTTC                float64   `db:"total_ttc"`
	Timbre                  float64   `db:"timbre"`
	NetAmount               float64   `db:"net_amount"`
	AmountPaid              float64   `db:"amount_paid"`
	AmountRemaining         float64   `db:"amount_remaining"`
	Status                  string    `db:"status"`
	Notes                   string    `db:"notes"`
	DriverID                *int      `db:"driver_id"`
	DeliveryAddress         string    `db:"delivery_address"`
	SupplierInvoiceNumber   string    `db:"supplier_invoice_number"`
	ValidityDays            int       `db:"validity_days"`
	SourceDocID             *int      `db:"source_doc_id"`
	CreatedBy               *int      `db:"created_by"`
	CreatedAt               time.Time `db:"created_at"`
	UpdatedAt               time.Time `db:"updated_at"`

	// Relations chargées dynamiquement
	ClientName   string        `db:"client_name"`
	SupplierName string        `db:"supplier_name"`
	DriverName   string        `db:"driver_name"`
	Lines        []DocumentLine `db:"-"`

	// Totaux calculés par TVA
	TVA9         float64 `db:"-"`
	TVA19        float64 `db:"-"`

	// Montant en lettres
	AmountInWordsFr string `db:"-"`
	AmountInWordsAr string `db:"-"`
}

// DocumentLine représente une ligne d'un document commercial
type DocumentLine struct {
	ID              int     `db:"id"`
	DocumentID      int     `db:"document_id"`
	LineNumber      int     `db:"line_number"`
	ArticleID       *int    `db:"article_id"`
	Designation     string  `db:"designation"`
	Quantity        float64 `db:"quantity"`
	Unit            string  `db:"unit"`
	UnitPriceHT     float64 `db:"unit_price_ht"`
	DiscountPercent float64 `db:"discount_percent"`
	DiscountAmount  float64 `db:"discount_amount"`
	AmountHT        float64 `db:"amount_ht"`
	TVARate         float64 `db:"tva_rate"`
	TVAAmount       float64 `db:"tva_amount"`
	AmountTTC       float64 `db:"amount_ttc"`
	LotNumber       string  `db:"lot_number"`
	ExpiryDate      string  `db:"expiry_date"`

	// Données article en cache
	Reference string  `db:"reference"`
	Barcode   string  `db:"barcode"`
	CMUP      float64 `db:"cmup"`
}

// InvoiceSummary représente le résumé de calcul d'un document
type InvoiceSummary struct {
	TotalHT        float64
	TotalDiscount  float64
	NetHT          float64
	TVA9           float64
	TVA19          float64
	TotalTVA       float64
	TotalTTC       float64
	Timbre         float64
	NetToPay       float64
	AmountInWordsFr string
	AmountInWordsAr string
}

// Payment représente un encaissement ou décaissement
type Payment struct {
	ID            int     `db:"id"`
	Type          string  `db:"type"` // collection | disbursement
	Date          string  `db:"date"`
	ClientID      *int    `db:"client_id"`
	SupplierID    *int    `db:"supplier_id"`
	Amount        float64 `db:"amount"`
	PaymentMethod string  `db:"payment_method"`
	ChequeNumber  string  `db:"cheque_number"`
	BankName      string  `db:"bank_name"`
	Reference     string  `db:"reference"`
	Notes         string  `db:"notes"`
	CreatedBy     *int    `db:"created_by"`
	CreatedAt     string  `db:"created_at"`

	// Relations
	ClientName   string `db:"client_name"`
	SupplierName string `db:"supplier_name"`
}

// PaymentAllocation représente la ventilation d'un paiement sur les factures
type PaymentAllocation struct {
	ID         int     `db:"id"`
	PaymentID  int     `db:"payment_id"`
	DocumentID int     `db:"document_id"`
	Amount     float64 `db:"amount"`
}

// Cheque représente un chèque émis ou reçu
type Cheque struct {
	ID               int     `db:"id"`
	Type             string  `db:"type"` // issued | received
	ChequeNumber     string  `db:"cheque_number"`
	Date             string  `db:"date"`
	DueDate          string  `db:"due_date"`
	Amount           float64 `db:"amount"`
	PayerPayee       string  `db:"payer_payee"`
	BankName         string  `db:"bank_name"`
	Status           string  `db:"status"`
	RejectReason     string  `db:"reject_reason"`
	RelatedPaymentID *int    `db:"related_payment_id"`
	Notes            string  `db:"notes"`
	CreatedAt        string  `db:"created_at"`
}

// CashMovement représente un mouvement de caisse
type CashMovement struct {
	ID          int     `db:"id"`
	Date        string  `db:"date"`
	Type        string  `db:"type"` // in | out
	Category    string  `db:"category"`
	Description string  `db:"description"`
	Reference   string  `db:"reference"`
	PartyName   string  `db:"party_name"`
	Amount      float64 `db:"amount"`
	Balance     float64 `db:"balance"` // solde cumulé
	CreatedBy   *int    `db:"created_by"`
}

// BankAccount représente un compte bancaire
type BankAccount struct {
	ID            int     `db:"id"`
	BankName      string  `db:"bank_name"`
	Branch        string  `db:"branch"`
	AccountNumber string  `db:"account_number"`
	RIB           string  `db:"rib"`
	Balance       float64 `db:"balance"`
}

// BankMovement représente un mouvement bancaire
type BankMovement struct {
	ID            int     `db:"id"`
	BankAccountID int     `db:"bank_account_id"`
	Date          string  `db:"date"`
	Type          string  `db:"type"`
	Description   string  `db:"description"`
	Reference     string  `db:"reference"`
	Debit         float64 `db:"debit"`
	Credit        float64 `db:"credit"`
	IsReconciled  bool    `db:"is_reconciled"`
	CreatedAt     string  `db:"created_at"`
}

// Expense représente une dépense diverse
type Expense struct {
	ID            int     `db:"id"`
	Date          string  `db:"date"`
	CategoryID    *int    `db:"category_id"`
	CategoryName  string  `db:"category_name"`
	Description   string  `db:"description"`
	Amount        float64 `db:"amount"`
	PaymentMethod string  `db:"payment_method"`
	CreatedBy     *int    `db:"created_by"`
	CreatedAt     string  `db:"created_at"`
}

// ExpenseCategory représente une catégorie de dépenses
type ExpenseCategory struct {
	ID   int    `db:"id"`
	Name string `db:"name"`
}

// TaxDeclaration représente une déclaration fiscale (G50, annuelle)
type TaxDeclaration struct {
	ID        int    `db:"id"`
	Type      string `db:"type"` // G50 | annual
	Year      int    `db:"year"`
	Month     *int   `db:"month"`
	DataJSON  string `db:"data_json"`
	Status    string `db:"status"`
	CreatedAt string `db:"created_at"`
}

// G50Data contient les données du formulaire G50
type G50Data struct {
	Year      int
	Month     int
	Period    string

	// Chiffre d'affaires par taux TVA
	Revenue19    float64
	Revenue9     float64
	RevenueExempt float64
	TotalRevenue float64

	// TVA collectée
	TVACollected19 float64
	TVACollected9  float64
	TotalTVACollected float64

	// TVA déductible
	TVADeductiblePurchases    float64
	TVADeductibleInvestments  float64
	TVAPrecompte              float64

	// TVA due ou crédit
	TVADue         float64
	TVAPrecompteNext float64

	// TAP
	TAP float64

	// Timbre
	Timbre float64

	// IRG/IBS sur salaires
	IRGSalaries float64

	// Total à payer
	TotalDue float64
}

// Reminder représente un rappel/alerte
type Reminder struct {
	ID          int    `db:"id"`
	Date        string `db:"date"`
	Title       string `db:"title"`
	Description string `db:"description"`
	Type        string `db:"type"` // custom | tax | cheque | delivery
	RelatedID   *int   `db:"related_id"`
	IsDone      bool   `db:"is_done"`
	CreatedBy   *int   `db:"created_by"`
}

// Currency représente une devise
type Currency struct {
	ID     int     `db:"id"`
	Name   string  `db:"name"`
	Code   string  `db:"code"`
	Symbol string  `db:"symbol"`
	Rate   float64 `db:"rate"`
}

// NumberingConfig représente la configuration de numérotation
type NumberingConfig struct {
	ID            int    `db:"id"`
	DocType       string `db:"doc_type"`
	Prefix        string `db:"prefix"`
	CurrentNumber int    `db:"current_number"`
	ResetYearly   bool   `db:"reset_yearly"`
}
