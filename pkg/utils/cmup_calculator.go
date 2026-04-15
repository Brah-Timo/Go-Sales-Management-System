package utils

import "math"

// CMUPState représente l'état du calcul CMUP
type CMUPState struct {
	TotalQty   float64
	TotalValue float64
}

// NewCMUPState crée un nouvel état CMUP
func NewCMUPState() *CMUPState {
	return &CMUPState{}
}

// AddPurchase ajoute une entrée de stock (achat, retour, ajustement positif)
func (c *CMUPState) AddPurchase(qty, unitPrice float64) {
	if qty <= 0 {
		return
	}
	c.TotalValue += qty * unitPrice
	c.TotalQty += qty
}

// AddSale ajoute une sortie de stock en appliquant le CMUP actuel
func (c *CMUPState) AddSale(qty float64) {
	if qty <= 0 || c.TotalQty <= 0 {
		return
	}
	cmup := c.CMUP()
	c.TotalValue -= qty * cmup
	c.TotalQty -= qty
	if c.TotalQty < 0 {
		c.TotalQty = 0
		c.TotalValue = 0
	}
}

// CMUP retourne le Coût Moyen Unitaire Pondéré actuel
func (c *CMUPState) CMUP() float64 {
	if c.TotalQty <= 0 {
		return 0
	}
	return math.Round(c.TotalValue/c.TotalQty*10000) / 10000
}

// StockMovementForCMUP structure simplifiée pour le calcul CMUP
type StockMovementForCMUP struct {
	Type      string
	Quantity  float64
	UnitPrice float64
}

// RecalculateCMUP recalcule le CMUP depuis l'historique complet des mouvements
func RecalculateCMUP(movements []StockMovementForCMUP) float64 {
	state := NewCMUPState()

	for _, m := range movements {
		switch m.Type {
		case "purchase_in", "adjustment_in", "return_in", "transfer_in":
			state.AddPurchase(m.Quantity, m.UnitPrice)
		case "sale_out", "adjustment_out", "return_out", "transfer_out", "damage":
			state.AddSale(m.Quantity)
		}
	}

	return state.CMUP()
}

// StockValueAtCMUP calcule la valeur du stock au CMUP
func StockValueAtCMUP(qty, cmup float64) float64 {
	return math.Round(qty*cmup*100) / 100
}

// Round2 arrondit à 2 décimales
func Round2(v float64) float64 {
	return math.Round(v*100) / 100
}
