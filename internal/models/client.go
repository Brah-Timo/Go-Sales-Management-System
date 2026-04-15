package models

import "time"

// Client représente un client
type Client struct {
	ID           int       `db:"id"`
	Code         string    `db:"code"`
	NameAr       string    `db:"name_ar"`
	NameFr       string    `db:"name_fr"`
	Type         string    `db:"type"` // person | company
	Address      string    `db:"address"`
	Wilaya       string    `db:"wilaya"`
	Commune      string    `db:"commune"`
	Phone        string    `db:"phone"`
	Mobile       string    `db:"mobile"`
	Fax          string    `db:"fax"`
	Email        string    `db:"email"`
	NIF          string    `db:"nif"`
	NIS          string    `db:"nis"`
	RC           string    `db:"rc"`
	AI           string    `db:"ai"`
	PriceListID  *int      `db:"price_list_id"`
	CreditLimit  float64   `db:"credit_limit"`
	PaymentTerms string    `db:"payment_terms"`
	DiscountRate float64   `db:"discount_rate"`
	Balance      float64   `db:"balance"`
	IsBlocked    bool      `db:"is_blocked"`
	Notes        string    `db:"notes"`
	CreatedAt    time.Time `db:"created_at"`

	// Relations
	PriceListName string `db:"price_list_name"`
}

// Supplier représente un fournisseur
type Supplier struct {
	ID             int       `db:"id"`
	Code           string    `db:"code"`
	NameAr         string    `db:"name_ar"`
	NameFr         string    `db:"name_fr"`
	Address        string    `db:"address"`
	Wilaya         string    `db:"wilaya"`
	Phone          string    `db:"phone"`
	Mobile         string    `db:"mobile"`
	Fax            string    `db:"fax"`
	Email          string    `db:"email"`
	NIF            string    `db:"nif"`
	NIS            string    `db:"nis"`
	RC             string    `db:"rc"`
	AI             string    `db:"ai"`
	PaymentTerms   string    `db:"payment_terms"`
	Balance        float64   `db:"balance"`
	RatingDelivery int       `db:"rating_delivery"`
	RatingQuality  int       `db:"rating_quality"`
	RatingPricing  int       `db:"rating_pricing"`
	Notes          string    `db:"notes"`
	CreatedAt      time.Time `db:"created_at"`
}

// Driver représente un livreur/chauffeur
type Driver struct {
	ID            int    `db:"id"`
	Name          string `db:"name"`
	Phone         string `db:"phone"`
	VehiclePlate  string `db:"vehicle_plate"`
	DeliveryCount int    `db:"delivery_count"`
}

// AccountStatement représente une ligne de relevé de compte
type AccountStatement struct {
	Date          string  `db:"date"`
	DocType       string  `db:"doc_type"`
	DocNumber     string  `db:"doc_number"`
	Description   string  `db:"description"`
	Debit         float64 `db:"debit"`
	Credit        float64 `db:"credit"`
	Balance       float64 `db:"balance"`
}

// AgingLine représente une ligne de balance âgée
type AgingLine struct {
	ClientName string  `db:"name_ar"`
	Balance    float64 `db:"balance"`
	Age0_30    float64 `db:"age_0_30"`
	Age31_60   float64 `db:"age_31_60"`
	Age61_90   float64 `db:"age_61_90"`
	Age90Plus  float64 `db:"age_90_plus"`
}
