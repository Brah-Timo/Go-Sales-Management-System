package queries

import (
	"database/sql"
	"fmt"
	"gestion-commerciale/internal/models"
	"math"
)

// DashboardStats contient les statistiques du tableau de bord
type DashboardStats struct {
	TodaySales     float64
	MonthSales     float64
	MonthPurchases float64
	CashBalance    float64
	MonthProfit    float64
	ClientDebts    float64
	SupplierDebts  float64
	LowStockCount  int
}

// GetDashboardStats calcule toutes les statistiques du tableau de bord
func GetDashboardStats(db *sql.DB) DashboardStats {
	var stats DashboardStats

	db.QueryRow(`SELECT COALESCE(SUM(net_amount),0) FROM documents
		WHERE doc_type='FA' AND status IN ('confirmed','paid','partial') AND date(date)=date('now')`).
		Scan(&stats.TodaySales)

	db.QueryRow(`SELECT COALESCE(SUM(net_amount),0) FROM documents
		WHERE doc_type='FA' AND status IN ('confirmed','paid','partial')
		AND strftime('%Y-%m',date)=strftime('%Y-%m','now')`).
		Scan(&stats.MonthSales)

	db.QueryRow(`SELECT COALESCE(SUM(net_amount),0) FROM documents
		WHERE doc_type='FAC' AND status IN ('confirmed','paid','partial')
		AND strftime('%Y-%m',date)=strftime('%Y-%m','now')`).
		Scan(&stats.MonthPurchases)

	db.QueryRow(`SELECT COALESCE(SUM(CASE WHEN type='in' THEN amount ELSE -amount END),0) FROM cash_movements`).
		Scan(&stats.CashBalance)

	db.QueryRow(`SELECT COALESCE(SUM(balance),0) FROM clients WHERE balance > 0`).
		Scan(&stats.ClientDebts)

	db.QueryRow(`SELECT COALESCE(SUM(balance),0) FROM suppliers WHERE balance > 0`).
		Scan(&stats.SupplierDebts)

	db.QueryRow(`SELECT COUNT(*) FROM articles WHERE is_active=1 AND stock_qty <= stock_min`).
		Scan(&stats.LowStockCount)

	return stats
}

// SalesChartData retourne les données graphique ventes 30 jours
func SalesChartData(db *sql.DB) ([]string, []float64, error) {
	rows, err := db.Query(`
		SELECT date(date) as d, COALESCE(SUM(net_amount),0)
		FROM documents
		WHERE doc_type='FA' AND status IN ('confirmed','paid','partial')
		AND date >= date('now', '-30 days')
		GROUP BY d ORDER BY d`)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	var dates []string
	var amounts []float64
	for rows.Next() {
		var d string
		var a float64
		rows.Scan(&d, &a)
		dates = append(dates, d)
		amounts = append(amounts, a)
	}
	return dates, amounts, nil
}

// MonthlySalesVsPurchases retourne la comparaison mensuelle (12 mois)
func MonthlySalesVsPurchases(db *sql.DB) ([]string, []float64, []float64, error) {
	rows, err := db.Query(`
		SELECT strftime('%Y-%m', date) as month,
		  COALESCE(SUM(CASE WHEN doc_type='FA'  THEN net_amount ELSE 0 END),0),
		  COALESCE(SUM(CASE WHEN doc_type='FAC' THEN net_amount ELSE 0 END),0)
		FROM documents
		WHERE status IN ('confirmed','paid','partial')
		AND date >= date('now', '-12 months')
		GROUP BY month ORDER BY month`)
	if err != nil {
		return nil, nil, nil, err
	}
	defer rows.Close()

	var months []string
	var sales, purchases []float64
	for rows.Next() {
		var m string
		var s, p float64
		rows.Scan(&m, &s, &p)
		months = append(months, m)
		sales = append(sales, s)
		purchases = append(purchases, p)
	}
	return months, sales, purchases, nil
}

// TopSellingArticles retourne les 10 articles les plus vendus
func TopSellingArticles(db *sql.DB) ([]struct {
	Name     string
	Quantity float64
	Amount   float64
}, error) {
	rows, err := db.Query(`
		SELECT a.name_fr, COALESCE(SUM(dl.quantity),0), COALESCE(SUM(dl.amount_ttc),0)
		FROM document_lines dl
		JOIN documents d ON dl.document_id = d.id
		JOIN articles a ON dl.article_id = a.id
		WHERE d.doc_type='FA' AND d.status IN ('confirmed','paid','partial')
		GROUP BY a.id ORDER BY SUM(dl.quantity) DESC LIMIT 10`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []struct {
		Name     string
		Quantity float64
		Amount   float64
	}
	for rows.Next() {
		var r struct {
			Name     string
			Quantity float64
			Amount   float64
		}
		rows.Scan(&r.Name, &r.Quantity, &r.Amount)
		results = append(results, r)
	}
	return results, nil
}

// GetAgingReport retourne la balance âgée des clients
func GetAgingReport(db *sql.DB) ([]models.AgingLine, error) {
	rows, err := db.Query(`
		SELECT c.name_fr, c.balance,
		  COALESCE(SUM(CASE WHEN julianday('now') - julianday(d.date) <= 30 THEN d.amount_remaining ELSE 0 END),0),
		  COALESCE(SUM(CASE WHEN julianday('now') - julianday(d.date) BETWEEN 31 AND 60 THEN d.amount_remaining ELSE 0 END),0),
		  COALESCE(SUM(CASE WHEN julianday('now') - julianday(d.date) BETWEEN 61 AND 90 THEN d.amount_remaining ELSE 0 END),0),
		  COALESCE(SUM(CASE WHEN julianday('now') - julianday(d.date) > 90 THEN d.amount_remaining ELSE 0 END),0)
		FROM clients c
		JOIN documents d ON d.client_id=c.id AND d.doc_type='FA'
		  AND d.status IN ('confirmed','partial') AND d.amount_remaining > 0
		WHERE c.balance > 0
		GROUP BY c.id ORDER BY c.balance DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var lines []models.AgingLine
	for rows.Next() {
		var l models.AgingLine
		rows.Scan(&l.ClientName, &l.Balance, &l.Age0_30, &l.Age31_60, &l.Age61_90, &l.Age90Plus)
		lines = append(lines, l)
	}
	return lines, nil
}

// CalculateG50 calcule les données du formulaire G50
func CalculateG50(db *sql.DB, year, month int) (models.G50Data, error) {
	period := fmt.Sprintf("%d-%02d", year, month)
	var g models.G50Data
	g.Year = year
	g.Month = month
	g.Period = period

	// Chiffre d'affaires par taux TVA
	err := db.QueryRow(`
		SELECT
		  COALESCE(SUM(CASE WHEN dl.tva_rate=19 THEN dl.amount_ht ELSE 0 END),0),
		  COALESCE(SUM(CASE WHEN dl.tva_rate=9  THEN dl.amount_ht ELSE 0 END),0),
		  COALESCE(SUM(CASE WHEN dl.tva_rate=0  THEN dl.amount_ht ELSE 0 END),0)
		FROM document_lines dl
		JOIN documents d ON dl.document_id = d.id
		WHERE d.doc_type='FA' AND d.status IN ('confirmed','paid','partial')
		AND strftime('%Y-%m', d.date)=?`, period,
	).Scan(&g.Revenue19, &g.Revenue9, &g.RevenueExempt)
	if err != nil {
		return g, err
	}

	g.TotalRevenue = g.Revenue19 + g.Revenue9 + g.RevenueExempt
	g.TVACollected19 = r2(g.Revenue19 * 0.19)
	g.TVACollected9 = r2(g.Revenue9 * 0.09)
	g.TotalTVACollected = g.TVACollected19 + g.TVACollected9

	db.QueryRow(`
		SELECT COALESCE(SUM(dl.tva_amount),0)
		FROM document_lines dl
		JOIN documents d ON dl.document_id=d.id
		WHERE d.doc_type='FAC' AND d.status IN ('confirmed','paid','partial')
		AND strftime('%Y-%m', d.date)=?`, period,
	).Scan(&g.TVADeductiblePurchases)

	totalDeductible := g.TVADeductiblePurchases + g.TVADeductibleInvestments + g.TVAPrecompte
	if g.TotalTVACollected > totalDeductible {
		g.TVADue = r2(g.TotalTVACollected - totalDeductible)
	} else {
		g.TVAPrecompteNext = r2(totalDeductible - g.TotalTVACollected)
	}

	var tapRateStr string
	db.QueryRow(`SELECT COALESCE(value,'1') FROM settings WHERE key='tap_rate'`).Scan(&tapRateStr)
	tapRate := 1.0
	fmt.Sscanf(tapRateStr, "%f", &tapRate)
	g.TAP = r2(g.TotalRevenue * tapRate / 100)

	db.QueryRow(`
		SELECT COALESCE(SUM(timbre),0) FROM documents
		WHERE doc_type='FA' AND status IN ('confirmed','paid','partial')
		AND payment_method='cash' AND strftime('%Y-%m', date)=?`, period,
	).Scan(&g.Timbre)

	g.TotalDue = r2(g.TVADue + g.TAP + g.Timbre + g.IRGSalaries)
	return g, nil
}

// GetCashBalance retourne le solde actuel de la caisse
func GetCashBalance(db *sql.DB) float64 {
	var balance float64
	db.QueryRow(`SELECT COALESCE(SUM(CASE WHEN type='in' THEN amount ELSE -amount END),0) FROM cash_movements`).
		Scan(&balance)
	return balance
}

// GetBankBalance retourne le solde d'un compte bancaire
func GetBankBalance(db *sql.DB, accountID int) float64 {
	var balance float64
	db.QueryRow(`SELECT COALESCE(SUM(credit - debit),0) FROM bank_movements WHERE bank_account_id=?`, accountID).
		Scan(&balance)
	return balance
}

// r2 arrondit à 2 décimales
func r2(v float64) float64 {
	return math.Round(v*100) / 100
}
