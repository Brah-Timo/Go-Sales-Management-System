package services

import (
	"database/sql"
	"fmt"
	"gestion-commerciale/pkg/utils"
)

// ReportService gère les rapports et statistiques
type ReportService struct {
	db *sql.DB
}

// NewReportService crée un service de rapports
func NewReportService(db *sql.DB) *ReportService {
	return &ReportService{db: db}
}

// DashboardStats contient les statistiques pour le tableau de bord
type DashboardStats struct {
	TodaySales       float64
	MonthSales       float64
	MonthPurchases   float64
	CashBalance      float64
	MonthProfit      float64
	ClientDebt       float64
	SupplierDebt     float64
	LowStockCount    int
	RecentInvoices   []InvoiceRow
	DailySalesChart  []DailyAmount
	MonthlySalesChart []MonthlyAmount
	CategoryChart    []CategoryAmount
	TopProducts      []ProductSales
}

type InvoiceRow struct {
	DocType     string
	DocNumber   string
	Date        string
	PartyName   string
	NetAmount   float64
	Status      string
}

type DailyAmount struct {
	Date   string
	Amount float64
}

type MonthlyAmount struct {
	Month  string
	Sales  float64
	Purchases float64
}

type CategoryAmount struct {
	Category string
	Amount   float64
}

type ProductSales struct {
	Reference  string
	Name       string
	Quantity   float64
	Amount     float64
}

// GetDashboardStats calcule toutes les statistiques pour le tableau de bord
func (s *ReportService) GetDashboardStats() DashboardStats {
	var stats DashboardStats

	// Ventes du jour
	s.db.QueryRow(`
		SELECT COALESCE(SUM(net_amount),0) FROM documents
		WHERE doc_type='FA' AND status IN ('confirmed','paid','partial')
		  AND date(date)=date('now')`).Scan(&stats.TodaySales)

	// Ventes du mois
	s.db.QueryRow(`
		SELECT COALESCE(SUM(net_amount),0) FROM documents
		WHERE doc_type='FA' AND status IN ('confirmed','paid','partial')
		  AND strftime('%Y-%m',date)=strftime('%Y-%m','now')`).Scan(&stats.MonthSales)

	// Achats du mois
	s.db.QueryRow(`
		SELECT COALESCE(SUM(net_amount),0) FROM documents
		WHERE doc_type='FAC' AND status IN ('confirmed','paid','partial')
		  AND strftime('%Y-%m',date)=strftime('%Y-%m','now')`).Scan(&stats.MonthPurchases)

	// Solde caisse
	s.db.QueryRow(`
		SELECT COALESCE(SUM(CASE WHEN type='in' THEN amount ELSE -amount END),0)
		FROM cash_movements`).Scan(&stats.CashBalance)

	// Dettes clients
	s.db.QueryRow(`SELECT COALESCE(SUM(balance),0) FROM clients WHERE balance > 0`).Scan(&stats.ClientDebt)

	// Dettes fournisseurs
	s.db.QueryRow(`SELECT COALESCE(SUM(balance),0) FROM suppliers WHERE balance > 0`).Scan(&stats.SupplierDebt)

	// Articles sous stock minimum
	s.db.QueryRow(`SELECT COUNT(*) FROM articles WHERE is_active=1 AND stock_qty <= stock_min`).Scan(&stats.LowStockCount)

	// Bénéfice du mois (approx)
	var cogs float64
	s.db.QueryRow(`
		SELECT COALESCE(SUM(dl.quantity * a.cmup), 0)
		FROM document_lines dl
		JOIN documents d ON dl.document_id=d.id
		JOIN articles a ON dl.article_id=a.id
		WHERE d.doc_type='FA' AND d.status IN ('confirmed','paid','partial')
		  AND strftime('%Y-%m',d.date)=strftime('%Y-%m','now')`).Scan(&cogs)
	stats.MonthProfit = stats.MonthSales - cogs

	// 10 derniers documents
	rows, _ := s.db.Query(`
		SELECT d.doc_type, d.doc_number, d.date,
		  COALESCE(c.name_fr, su.name_fr, 'N/A') as party,
		  d.net_amount, d.status
		FROM documents d
		LEFT JOIN clients c ON d.client_id=c.id
		LEFT JOIN suppliers su ON d.supplier_id=su.id
		ORDER BY d.created_at DESC LIMIT 10`)
	if rows != nil {
		defer rows.Close()
		for rows.Next() {
			var r InvoiceRow
			rows.Scan(&r.DocType, &r.DocNumber, &r.Date, &r.PartyName, &r.NetAmount, &r.Status)
			stats.RecentInvoices = append(stats.RecentInvoices, r)
		}
	}

	// Ventes journalières (30 derniers jours)
	rows2, _ := s.db.Query(`
		SELECT date(date) as d, COALESCE(SUM(net_amount),0)
		FROM documents
		WHERE doc_type='FA' AND status IN ('confirmed','paid','partial')
		  AND date >= date('now','-30 days')
		GROUP BY d ORDER BY d`)
	if rows2 != nil {
		defer rows2.Close()
		for rows2.Next() {
			var da DailyAmount
			rows2.Scan(&da.Date, &da.Amount)
			stats.DailySalesChart = append(stats.DailySalesChart, da)
		}
	}

	// Ventes/Achats mensuels (12 mois)
	rows3, _ := s.db.Query(`
		SELECT strftime('%Y-%m', date) as m,
		  COALESCE(SUM(CASE WHEN doc_type='FA' AND status IN ('confirmed','paid','partial') THEN net_amount ELSE 0 END), 0) as sales,
		  COALESCE(SUM(CASE WHEN doc_type='FAC' AND status IN ('confirmed','paid','partial') THEN net_amount ELSE 0 END), 0) as purchases
		FROM documents
		WHERE date >= date('now','-12 months')
		GROUP BY m ORDER BY m`)
	if rows3 != nil {
		defer rows3.Close()
		for rows3.Next() {
			var ma MonthlyAmount
			rows3.Scan(&ma.Month, &ma.Sales, &ma.Purchases)
			stats.MonthlySalesChart = append(stats.MonthlySalesChart, ma)
		}
	}

	// Ventes par catégorie
	rows4, _ := s.db.Query(`
		SELECT COALESCE(c.name_fr,'Non classé') as cat, COALESCE(SUM(dl.amount_ttc),0)
		FROM document_lines dl
		JOIN documents d ON dl.document_id=d.id
		JOIN articles a ON dl.article_id=a.id
		LEFT JOIN categories c ON a.category_id=c.id
		WHERE d.doc_type='FA' AND d.status IN ('confirmed','paid','partial')
		  AND strftime('%Y',d.date)=strftime('%Y','now')
		GROUP BY c.id ORDER BY 2 DESC LIMIT 8`)
	if rows4 != nil {
		defer rows4.Close()
		for rows4.Next() {
			var ca CategoryAmount
			rows4.Scan(&ca.Category, &ca.Amount)
			stats.CategoryChart = append(stats.CategoryChart, ca)
		}
	}

	// Top 10 produits
	rows5, _ := s.db.Query(`
		SELECT a.reference, a.name_fr, COALESCE(SUM(dl.quantity),0) as qty, COALESCE(SUM(dl.amount_ttc),0) as amt
		FROM document_lines dl
		JOIN documents d ON dl.document_id=d.id
		JOIN articles a ON dl.article_id=a.id
		WHERE d.doc_type='FA' AND d.status IN ('confirmed','paid','partial')
		  AND strftime('%Y-%m',d.date)=strftime('%Y-%m','now')
		GROUP BY a.id ORDER BY qty DESC LIMIT 10`)
	if rows5 != nil {
		defer rows5.Close()
		for rows5.Next() {
			var ps ProductSales
			rows5.Scan(&ps.Reference, &ps.Name, &ps.Quantity, &ps.Amount)
			stats.TopProducts = append(stats.TopProducts, ps)
		}
	}

	return stats
}

// SalesReport contient les données d'un rapport de ventes
type SalesReport struct {
	Rows      []SalesRow
	TotalHT   float64
	TotalTVA  float64
	TotalTTC  float64
	TotalTimbre float64
	InvoiceCount int
}

type SalesRow struct {
	DocNumber   string
	Date        string
	ClientName  string
	TotalHT     float64
	TotalTVA    float64
	TotalTTC    float64
	Timbre      float64
	AmountPaid  float64
	AmountRemaining float64
	PaymentMethod string
	Status      string
}

// GetSalesReport génère un rapport de ventes
func (s *ReportService) GetSalesReport(from, to, clientName, status, paymentMethod string) (*SalesReport, error) {
	query := `
		SELECT d.doc_number, d.date,
		  COALESCE(c.name_fr,'Client de passage') as client_name,
		  d.net_ht, d.total_tva, d.total_ttc, d.timbre,
		  d.amount_paid, d.amount_remaining, d.payment_method, d.status
		FROM documents d
		LEFT JOIN clients c ON d.client_id=c.id
		WHERE d.doc_type='FA'`

	args := []interface{}{}
	if from != "" {
		query += ` AND date(d.date) >= ?`
		args = append(args, from)
	}
	if to != "" {
		query += ` AND date(d.date) <= ?`
		args = append(args, to)
	}
	if clientName != "" {
		query += ` AND c.name_fr LIKE ?`
		args = append(args, "%"+clientName+"%")
	}
	if status != "" && status != "Tous" {
		query += ` AND d.status=?`
		args = append(args, status)
	}
	if paymentMethod != "" && paymentMethod != "Tous" {
		query += ` AND d.payment_method=?`
		args = append(args, paymentMethod)
	}
	query += ` ORDER BY d.date DESC`

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	report := &SalesReport{}
	for rows.Next() {
		var r SalesRow
		rows.Scan(&r.DocNumber, &r.Date, &r.ClientName, &r.TotalHT, &r.TotalTVA, &r.TotalTTC,
			&r.Timbre, &r.AmountPaid, &r.AmountRemaining, &r.PaymentMethod, &r.Status)
		report.Rows = append(report.Rows, r)
		report.TotalHT += r.TotalHT
		report.TotalTVA += r.TotalTVA
		report.TotalTTC += r.TotalTTC
		report.TotalTimbre += r.Timbre
		report.InvoiceCount++
	}
	return report, nil
}

// StockReport contient les données d'un rapport de stock
type StockReport struct {
	Rows           []StockRow
	TotalValue     float64
	LowStockCount  int
	OutOfStockCount int
}

type StockRow struct {
	Reference   string
	Name        string
	Category    string
	Unit        string
	StockQty    float64
	StockMin    float64
	PurchasePrice float64
	CMUP        float64
	SalePriceTTC float64
	StockValue  float64
	Status      string
}

// GetStockReport génère un rapport de stock
func (s *ReportService) GetStockReport(categoryID int, stockFilter string) (*StockReport, error) {
	query := `
		SELECT a.reference, a.name_fr,
		  COALESCE(c.name_fr,'') as category,
		  COALESCE(u.symbol,'') as unit,
		  a.stock_qty, a.stock_min, a.purchase_price, a.cmup, a.sale_price_ttc
		FROM articles a
		LEFT JOIN categories c ON a.category_id=c.id
		LEFT JOIN units u ON a.unit_id=u.id
		WHERE a.is_active=1`

	args := []interface{}{}
	if categoryID > 0 {
		query += ` AND a.category_id=?`
		args = append(args, categoryID)
	}
	switch stockFilter {
	case "low":
		query += ` AND a.stock_qty <= a.stock_min`
	case "out":
		query += ` AND a.stock_qty <= 0`
	case "ok":
		query += ` AND a.stock_qty > a.stock_min`
	}
	query += ` ORDER BY a.name_fr`

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	report := &StockReport{}
	for rows.Next() {
		var r StockRow
		rows.Scan(&r.Reference, &r.Name, &r.Category, &r.Unit,
			&r.StockQty, &r.StockMin, &r.PurchasePrice, &r.CMUP, &r.SalePriceTTC)
		r.StockValue = r.StockQty * r.CMUP
		if r.StockQty <= 0 {
			r.Status = "Rupture"
			report.OutOfStockCount++
		} else if r.StockQty <= r.StockMin {
			r.Status = "Bas"
			report.LowStockCount++
		} else {
			r.Status = "OK"
		}
		report.TotalValue += r.StockValue
		report.Rows = append(report.Rows, r)
	}
	return report, nil
}

// ProfitReport contient les données d'un rapport de rentabilité
type ProfitReport struct {
	Period      string
	Revenue     float64
	COGS        float64
	GrossProfit float64
	Expenses    float64
	NetProfit   float64
	GrossMargin float64
	NetMargin   float64
}

// GetProfitReport génère un rapport de rentabilité mensuel
func (s *ReportService) GetProfitReport(year, month int) (*ProfitReport, error) {
	report := &ProfitReport{
		Period: fmt.Sprintf("%d/%02d", year, month),
	}
	period := fmt.Sprintf("%d-%02d", year, month)

	// Chiffre d'affaires
	s.db.QueryRow(`
		SELECT COALESCE(SUM(net_amount),0) FROM documents
		WHERE doc_type='FA' AND status IN ('confirmed','paid','partial')
		  AND strftime('%Y-%m',date)=?`, period).Scan(&report.Revenue)

	// Coût des marchandises vendues
	s.db.QueryRow(`
		SELECT COALESCE(SUM(dl.quantity * a.cmup),0)
		FROM document_lines dl
		JOIN documents d ON dl.document_id=d.id
		JOIN articles a ON dl.article_id=a.id
		WHERE d.doc_type='FA' AND d.status IN ('confirmed','paid','partial')
		  AND strftime('%Y-%m',d.date)=?`, period).Scan(&report.COGS)

	// Dépenses
	s.db.QueryRow(`
		SELECT COALESCE(SUM(amount),0) FROM expenses
		WHERE strftime('%Y-%m',date)=?`, period).Scan(&report.Expenses)

	report.GrossProfit = report.Revenue - report.COGS
	report.NetProfit = report.GrossProfit - report.Expenses

	if report.Revenue > 0 {
		report.GrossMargin = report.GrossProfit / report.Revenue * 100
		report.NetMargin = report.NetProfit / report.Revenue * 100
	}

	return report, nil
}

// GetTopClients retourne les meilleurs clients
func (s *ReportService) GetTopClients(from, to string, limit int) ([]struct {
	Name   string
	Amount float64
	Count  int
}, error) {
	rows, err := s.db.Query(`
		SELECT COALESCE(c.name_fr,'Client de passage') as name,
		  COALESCE(SUM(d.net_amount),0) as amount,
		  COUNT(d.id) as cnt
		FROM documents d
		LEFT JOIN clients c ON d.client_id=c.id
		WHERE d.doc_type='FA' AND d.status IN ('confirmed','paid','partial')
		  AND date(d.date) BETWEEN ? AND ?
		GROUP BY d.client_id ORDER BY amount DESC LIMIT ?`,
		from, to, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []struct {
		Name   string
		Amount float64
		Count  int
	}
	for rows.Next() {
		var r struct {
			Name   string
			Amount float64
			Count  int
		}
		rows.Scan(&r.Name, &r.Amount, &r.Count)
		result = append(result, r)
	}
	return result, nil
}

// GetTopProducts retourne les produits les plus vendus
func (s *ReportService) GetTopProducts(from, to string, limit int) ([]ProductSales, error) {
	rows, err := s.db.Query(`
		SELECT a.reference, a.name_fr,
		  COALESCE(SUM(dl.quantity),0) as qty,
		  COALESCE(SUM(dl.amount_ttc),0) as amt
		FROM document_lines dl
		JOIN documents d ON dl.document_id=d.id
		JOIN articles a ON dl.article_id=a.id
		WHERE d.doc_type='FA' AND d.status IN ('confirmed','paid','partial')
		  AND date(d.date) BETWEEN ? AND ?
		GROUP BY a.id ORDER BY qty DESC LIMIT ?`,
		from, to, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []ProductSales
	for rows.Next() {
		var p ProductSales
		rows.Scan(&p.Reference, &p.Name, &p.Quantity, &p.Amount)
		products = append(products, p)
	}
	return products, nil
}

// BusinessIndicators contient les indicateurs clés
type BusinessIndicators struct {
	StockTurnover     float64
	AvgCollectionDays float64
	AvgPaymentDays    float64
	GrossMarginPct    float64
	NetMarginPct      float64
	ReturnRate        float64
	AvgInvoiceValue   float64
}

// GetBusinessIndicators calcule les indicateurs clés de gestion
func (s *ReportService) GetBusinessIndicators(year int) BusinessIndicators {
	var ind BusinessIndicators

	// Ventes annuelles
	var revenue, cogs float64
	s.db.QueryRow(`
		SELECT COALESCE(SUM(net_amount),0) FROM documents
		WHERE doc_type='FA' AND status IN ('confirmed','paid','partial')
		  AND strftime('%Y',date)=?`, fmt.Sprint(year)).Scan(&revenue)

	s.db.QueryRow(`
		SELECT COALESCE(SUM(dl.quantity * a.cmup),0)
		FROM document_lines dl
		JOIN documents d ON dl.document_id=d.id
		JOIN articles a ON dl.article_id=a.id
		WHERE d.doc_type='FA' AND d.status IN ('confirmed','paid','partial')
		  AND strftime('%Y',d.date)=?`, fmt.Sprint(year)).Scan(&cogs)

	// Valeur du stock
	var stockValue float64
	s.db.QueryRow(`SELECT COALESCE(SUM(stock_qty * cmup),0) FROM articles WHERE is_active=1`).Scan(&stockValue)

	// Rotation du stock
	if stockValue > 0 {
		ind.StockTurnover = utils.Round2(cogs / stockValue)
	}

	// Marge brute
	if revenue > 0 {
		ind.GrossMarginPct = utils.Round2((revenue - cogs) / revenue * 100)
	}

	// Dépenses annuelles
	var expenses float64
	s.db.QueryRow(`
		SELECT COALESCE(SUM(amount),0) FROM expenses
		WHERE strftime('%Y',date)=?`, fmt.Sprint(year)).Scan(&expenses)

	netProfit := revenue - cogs - expenses
	if revenue > 0 {
		ind.NetMarginPct = utils.Round2(netProfit / revenue * 100)
	}

	// Délai moyen de recouvrement (DSO)
	var totalDebt, avgMonthSales float64
	s.db.QueryRow(`SELECT COALESCE(SUM(balance),0) FROM clients WHERE balance > 0`).Scan(&totalDebt)
	avgMonthSales = revenue / 12
	if avgMonthSales > 0 {
		ind.AvgCollectionDays = utils.Round2(totalDebt / avgMonthSales * 30)
	}

	// Valeur moyenne facture
	var invoiceCount int
	s.db.QueryRow(`
		SELECT COUNT(*) FROM documents
		WHERE doc_type='FA' AND status IN ('confirmed','paid','partial')
		  AND strftime('%Y',date)=?`, fmt.Sprint(year)).Scan(&invoiceCount)
	if invoiceCount > 0 {
		ind.AvgInvoiceValue = utils.Round2(revenue / float64(invoiceCount))
	}

	return ind
}
