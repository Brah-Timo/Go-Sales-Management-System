package utils

import "math"

// TVACalculator gère les calculs de TVA

// HTToTTC calcule le montant TTC depuis le montant HT
func HTToTTC(ht, tvaRate float64) float64 {
	return math.Round(ht*(1+tvaRate/100)*100) / 100
}

// TTCToHT calcule le montant HT depuis le montant TTC
func TTCToHT(ttc, tvaRate float64) float64 {
	if tvaRate == 0 {
		return ttc
	}
	return math.Round(ttc/(1+tvaRate/100)*100) / 100
}

// TVAAmount calcule le montant de TVA
func TVAAmount(ht, tvaRate float64) float64 {
	return math.Round(ht*tvaRate/100*100) / 100
}

// CalculateTimbre calcule le droit de timbre fiscal algérien
func CalculateTimbre(totalTTC float64, paymentMethod string, timbreRate, timbreMax, timbreExemption float64, autoTimbre bool) float64 {
	if paymentMethod != "cash" || !autoTimbre {
		return 0
	}
	if totalTTC < timbreExemption {
		return 0
	}
	timbre := math.Round(totalTTC*timbreRate/100*100) / 100
	if timbre > timbreMax {
		timbre = timbreMax
	}
	return timbre
}

// CalculateMargin calcule le taux de marge en %
func CalculateMargin(purchasePrice, salePriceHT float64) float64 {
	if purchasePrice <= 0 {
		return 0
	}
	return math.Round(((salePriceHT-purchasePrice)/purchasePrice)*100*100) / 100
}

// SalePriceFromMargin calcule le prix de vente HT depuis la marge %
func SalePriceFromMargin(purchasePrice, marginPercent float64) float64 {
	return math.Round(purchasePrice*(1+marginPercent/100)*100) / 100
}

// CalculateLineAmounts calcule tous les montants d'une ligne de document
func CalculateLineAmounts(unitPriceHT, quantity, discountPercent, tvaRate float64) (
	discountAmount, amountHT, tvaAmount, amountTTC float64) {

	grossHT := math.Round(unitPriceHT*quantity*100) / 100
	discountAmount = math.Round(grossHT*discountPercent/100*100) / 100
	amountHT = math.Round((grossHT-discountAmount)*100) / 100
	tvaAmount = math.Round(amountHT*tvaRate/100*100) / 100
	amountTTC = math.Round((amountHT+tvaAmount)*100) / 100
	return
}

// InvoiceSummaryCalc calcule le résumé d'un document à partir de ses lignes
type LineForCalc struct {
	AmountHT       float64
	DiscountAmount float64
	TVARate        float64
	TVAAmount      float64
	AmountTTC      float64
}

type SummaryResult struct {
	TotalHT       float64
	TotalDiscount float64
	NetHT         float64
	TVA9          float64
	TVA19         float64
	TotalTVA      float64
	TotalTTC      float64
	Timbre        float64
	NetToPay      float64
}

func CalculateDocumentSummary(lines []LineForCalc, globalDiscountPct float64,
	paymentMethod string, timbreRate, timbreMax, timbreExemption float64, autoTimbre bool) SummaryResult {

	var r SummaryResult

	for _, l := range lines {
		r.TotalHT += l.AmountHT + l.DiscountAmount
		r.TotalDiscount += l.DiscountAmount
		if l.TVARate == 9 {
			r.TVA9 += l.TVAAmount
		} else if l.TVARate == 19 {
			r.TVA19 += l.TVAAmount
		}
		r.TotalTVA += l.TVAAmount
	}

	r.NetHT = r.TotalHT - r.TotalDiscount

	// Remise globale supplémentaire
	if globalDiscountPct > 0 {
		globalDiscountAmt := math.Round(r.NetHT*globalDiscountPct/100*100) / 100
		r.TotalDiscount += globalDiscountAmt
		r.NetHT -= globalDiscountAmt

		// Recalcul TVA proportionnel
		factor := r.NetHT / (r.TotalHT - (r.TotalDiscount - globalDiscountAmt))
		if factor > 0 && factor <= 1 {
			r.TVA9 = math.Round(r.TVA9*factor*100) / 100
			r.TVA19 = math.Round(r.TVA19*factor*100) / 100
			r.TotalTVA = r.TVA9 + r.TVA19
		}
	}

	r.TotalTTC = math.Round((r.NetHT+r.TotalTVA)*100) / 100

	r.Timbre = CalculateTimbre(r.TotalTTC, paymentMethod, timbreRate, timbreMax, timbreExemption, autoTimbre)

	r.NetToPay = math.Round((r.TotalTTC+r.Timbre)*100) / 100

	return r
}
