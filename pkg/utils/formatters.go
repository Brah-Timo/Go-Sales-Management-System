package utils

import (
	"fmt"
	"math"
	"strings"
	"time"
)

// FormatMoney formate un montant en dinars algériens
func FormatMoney(amount float64) string {
	amount = math.Round(amount*100) / 100
	negative := amount < 0
	if negative { amount = -amount }

	intPart := int64(amount)
	decPart := math.Round((amount-float64(intPart))*100)

	intStr := formatIntWithSeparator(intPart, " ")
	result := fmt.Sprintf("%s,%02.0f DA", intStr, decPart)
	if negative { result = "-" + result }
	return result
}

// FormatAmount est un alias de FormatMoney (compatibilité)
func FormatAmount(amount float64) string {
	return FormatMoney(amount)
}

// FormatMoneyNoSymbol formate sans le symbole DA
func FormatMoneyNoSymbol(amount float64) string {
	amount = math.Round(amount*100) / 100
	intPart := int64(amount)
	decPart := math.Round((amount-float64(intPart))*100)
	return fmt.Sprintf("%s,%02.0f", formatIntWithSeparator(intPart, " "), decPart)
}

// FormatPercent formate un pourcentage
func FormatPercent(value float64) string {
	return fmt.Sprintf("%.2f%%", value)
}

// FormatDate formate une date
func FormatDate(t time.Time) string { return t.Format("02/01/2006") }

// FormatDateTime formate une date et heure
func FormatDateTime(t time.Time) string { return t.Format("02/01/2006 15:04") }

// ParseDate analyse une date au format DD/MM/YYYY
func ParseDate(s string) (time.Time, error) { return time.Parse("02/01/2006", s) }

// TodayString retourne la date du jour en format SQLite
func TodayString() string { return time.Now().Format("2006-01-02") }

// NowString retourne la date et heure actuelles en format SQLite
func NowString() string { return time.Now().Format("2006-01-02 15:04:05") }

// FormatDateFr formate une date SQLite en format français
func FormatDateFr(sqliteDate string) string {
	if sqliteDate == "" { return "" }
	t, err := time.Parse("2006-01-02", sqliteDate[:10])
	if err != nil { return sqliteDate }
	return t.Format("02/01/2006")
}

// FormatDateTimeFr formate une datetime SQLite en format français
func FormatDateTimeFr(sqliteDateTime string) string {
	if sqliteDateTime == "" { return "" }
	formats := []string{"2006-01-02 15:04:05", "2006-01-02T15:04:05", "2006-01-02"}
	for _, f := range formats {
		if len(sqliteDateTime) >= len(f) {
			if t, err := time.Parse(f, sqliteDateTime[:len(f)]); err == nil {
				if len(f) > 10 { return t.Format("02/01/2006 15:04") }
				return t.Format("02/01/2006")
			}
		}
	}
	return sqliteDateTime
}

// formatIntWithSeparator formate un entier avec séparateur de milliers
func formatIntWithSeparator(n int64, sep string) string {
	s := fmt.Sprintf("%d", n)
	if len(s) <= 3 { return s }
	var result strings.Builder
	start := len(s) % 3
	if start > 0 { result.WriteString(s[:start]) }
	for i := start; i < len(s); i += 3 {
		if i > 0 { result.WriteString(sep) }
		result.WriteString(s[i : i+3])
	}
	return result.String()
}

// CurrentYear retourne l'année courante
func CurrentYear() int { return time.Now().Year() }

// CurrentMonth retourne le mois courant
func CurrentMonth() int { return int(time.Now().Month()) }

// CurrentYearMonth retourne YYYY-MM
func CurrentYearMonth() string { return time.Now().Format("2006-01") }

// FormatQuantity formate une quantité
func FormatQuantity(qty float64) string {
	if qty == math.Trunc(qty) { return fmt.Sprintf("%.0f", qty) }
	return fmt.Sprintf("%.3f", qty)
}

// TruncateString tronque une chaîne
func TruncateString(s string, maxLen int) string {
	runes := []rune(s)
	if len(runes) <= maxLen { return s }
	return string(runes[:maxLen-3]) + "..."
}

// StatusColor retourne la couleur d'un statut
func StatusColor(status string) string {
	switch status {
	case "paid":      return "#27ae60"
	case "confirmed": return "#2980b9"
	case "partial":   return "#f39c12"
	case "draft":     return "#95a5a6"
	case "cancelled": return "#e74c3c"
	default:          return "#7f8c8d"
	}
}

// StatusLabel retourne le libellé d'un statut
func StatusLabel(status string) string {
	labels := map[string]string{
		"paid": "Payé", "confirmed": "Confirmé",
		"partial": "Partiel", "draft": "Brouillon", "cancelled": "Annulé",
	}
	if l, ok := labels[status]; ok { return l }
	return status
}

// PaymentMethodLabel retourne le libellé d'un mode de paiement
func PaymentMethodLabel(method string) string {
	labels := map[string]string{
		"cash": "Espèces", "cheque": "Chèque",
		"transfer": "Virement", "credit": "Crédit", "mixed": "Mixte",
	}
	if l, ok := labels[method]; ok { return l }
	return method
}

// StockMoveLabel retourne le libellé d'un type de mouvement
func StockMoveLabel(moveType string) string {
	labels := map[string]string{
		"purchase_in": "↑ Entrée Achat", "sale_out": "↓ Sortie Vente",
		"transfer_in": "→ Entrée Transfert", "transfer_out": "← Sortie Transfert",
		"return_in": "↩ Retour Client", "return_out": "↪ Retour Fournisseur",
		"adjustment_in": "+ Ajustement Positif", "adjustment_out": "- Ajustement Négatif",
		"damage": "✗ Perte/Dommage",
	}
	if l, ok := labels[moveType]; ok { return l }
	return moveType
}
