package excel

import (
	"fmt"
	"gestion-commerciale/internal/models"
	"gestion-commerciale/internal/services"
	"gestion-commerciale/pkg/utils"
	"strconv"
	"strings"

	"github.com/xuri/excelize/v2"
)

// ExcelService gère les imports/exports Excel
type ExcelService struct{}

// NewExcelService crée un service Excel
func NewExcelService() *ExcelService {
	return &ExcelService{}
}

// ─────────────────────────────────────────────────────────────────────────────
// EXPORT — Articles
// ─────────────────────────────────────────────────────────────────────────────

// ExportArticles exporte la liste des articles en Excel
func (s *ExcelService) ExportArticles(articles []models.Article) ([]byte, error) {
	f := excelize.NewFile()
	sheet := "Articles"
	f.SetSheetName("Sheet1", sheet)

	// En-têtes
	headers := []string{
		"Référence", "Code-barres", "Désignation FR", "Désignation AR",
		"Famille", "Marque", "Unité", "Prix Achat HT", "CMUP",
		"Prix Vente HT", "Prix Vente TTC", "Marge %", "TVA %",
		"Stock", "Stock Min", "Stock Max", "Statut",
	}

	styleHeader, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Color: "FFFFFF"},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"2980b9"}, Pattern: 1},
		Alignment: &excelize.Alignment{Horizontal: "center"},
	})

	for j, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(j+1, 1)
		f.SetCellValue(sheet, cell, h)
		f.SetCellStyle(sheet, cell, cell, styleHeader)
	}

	// Figer la première ligne
	f.SetPanes(sheet, &excelize.Panes{
		Freeze:      true,
		Split:       false,
		YSplit:      1,
		TopLeftCell: "A2",
		ActivePane:  "bottomLeft",
	})

	// Données
	styleEven, _ := f.NewStyle(&excelize.Style{
		Fill: excelize.Fill{Type: "pattern", Color: []string{"EBF5FB"}, Pattern: 1},
	})
	styleLow, _ := f.NewStyle(&excelize.Style{
		Fill: excelize.Fill{Type: "pattern", Color: []string{"FDEDEC"}, Pattern: 1},
		Font: &excelize.Font{Color: "E74C3C"},
	})

	for i, a := range articles {
		row := i + 2
		status := "Actif"
		if !a.IsActive {
			status = "Inactif"
		}

		values := []interface{}{
			a.Reference, a.Barcode, a.NameFr, a.NameAr,
			a.CategoryName, a.BrandName, a.UnitSymbol,
			a.PurchasePrice, a.CMUP,
			a.SalePriceHT, a.SalePriceTTC,
			a.MarginPercent, a.TVARate,
			a.StockQty, a.StockMin, a.StockMax, status,
		}

		for j, v := range values {
			cell, _ := excelize.CoordinatesToCellName(j+1, row)
			f.SetCellValue(sheet, cell, v)
		}

		// Couleur de ligne
		if a.StockQty <= a.StockMin {
			firstCell, _ := excelize.CoordinatesToCellName(1, row)
			lastCell, _ := excelize.CoordinatesToCellName(len(headers), row)
			f.SetCellStyle(sheet, firstCell, lastCell, styleLow)
		} else if i%2 == 0 {
			firstCell, _ := excelize.CoordinatesToCellName(1, row)
			lastCell, _ := excelize.CoordinatesToCellName(len(headers), row)
			f.SetCellStyle(sheet, firstCell, lastCell, styleEven)
		}
	}

	// Largeurs des colonnes
	colWidths := []float64{12, 14, 30, 30, 16, 14, 8, 12, 12, 12, 12, 8, 6, 8, 8, 8, 8}
	for j, w := range colWidths {
		col, _ := excelize.ColumnNumberToName(j + 1)
		f.SetColWidth(sheet, col, col, w)
	}

	buf, err := f.WriteToBuffer()
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// ─────────────────────────────────────────────────────────────────────────────
// EXPORT — Clients
// ─────────────────────────────────────────────────────────────────────────────

// ExportClients exporte la liste des clients en Excel
func (s *ExcelService) ExportClients(clients []models.Client) ([]byte, error) {
	f := excelize.NewFile()
	sheet := "Clients"
	f.SetSheetName("Sheet1", sheet)

	headers := []string{
		"Code", "Nom FR", "Nom AR", "Type", "Adresse", "Wilaya",
		"Téléphone", "Mobile", "Email", "NIF", "RC", "NIS",
		"Solde", "Plafond", "Conditions", "Remise %", "Statut",
	}

	styleHeader, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Color: "FFFFFF"},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"27ae60"}, Pattern: 1},
		Alignment: &excelize.Alignment{Horizontal: "center"},
	})

	for j, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(j+1, 1)
		f.SetCellValue(sheet, cell, h)
		f.SetCellStyle(sheet, cell, cell, styleHeader)
	}

	styleDebt, _ := f.NewStyle(&excelize.Style{
		Fill: excelize.Fill{Type: "pattern", Color: []string{"FDEDEC"}, Pattern: 1},
		Font: &excelize.Font{Color: "E74C3C"},
	})

	for i, c := range clients {
		row := i + 2
		status := "Actif"
		if c.IsBlocked {
			status = "Bloqué"
		}
		typeLabel := "Particulier"
		if c.Type == "company" {
			typeLabel = "Société"
		}

		values := []interface{}{
			c.Code, c.NameFr, c.NameAr, typeLabel, c.Address, c.Wilaya,
			c.Phone, c.Mobile, c.Email, c.NIF, c.RC, c.NIS,
			c.Balance, c.CreditLimit, c.PaymentTerms, c.DiscountRate, status,
		}

		for j, v := range values {
			cell, _ := excelize.CoordinatesToCellName(j+1, row)
			f.SetCellValue(sheet, cell, v)
		}

		if c.Balance > 0 {
			firstCell, _ := excelize.CoordinatesToCellName(1, row)
			lastCell, _ := excelize.CoordinatesToCellName(len(headers), row)
			f.SetCellStyle(sheet, firstCell, lastCell, styleDebt)
		}
	}

	buf, err := f.WriteToBuffer()
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// ─────────────────────────────────────────────────────────────────────────────
// EXPORT — Rapport de ventes
// ─────────────────────────────────────────────────────────────────────────────

// ExportSalesReport exporte un rapport de ventes en Excel
func (s *ExcelService) ExportSalesReport(report *services.SalesReport, from, to string) ([]byte, error) {
	f := excelize.NewFile()
	sheet := "Rapport Ventes"
	f.SetSheetName("Sheet1", sheet)

	headers := []string{
		"N° Facture", "Date", "Client", "Total HT", "TVA", "Total TTC",
		"Payé", "Reste", "Mode Paiement", "Statut",
	}

	styleHeader, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Color: "FFFFFF"},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"2980b9"}, Pattern: 1},
		Alignment: &excelize.Alignment{Horizontal: "center"},
	})

	// Titre
	f.SetCellValue(sheet, "A1", fmt.Sprintf("Rapport Ventes — %s au %s",
		utils.FormatDateFr(from), utils.FormatDateFr(to)))
	titleStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true, Size: 14},
	})
	f.SetCellStyle(sheet, "A1", "J1", titleStyle)
	f.MergeCell(sheet, "A1", "J1")

	for j, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(j+1, 2)
		f.SetCellValue(sheet, cell, h)
		f.SetCellStyle(sheet, cell, cell, styleHeader)
	}

	styleEven, _ := f.NewStyle(&excelize.Style{
		Fill: excelize.Fill{Type: "pattern", Color: []string{"EBF5FB"}, Pattern: 1},
	})

	for i, row := range report.Rows {
		r := i + 3
		values := []interface{}{
			row.DocNumber, utils.FormatDateFr(row.Date), row.ClientName,
			row.TotalHT, row.TotalTVA, row.TotalTTC,
			row.AmountPaid, row.AmountRemaining,
			utils.PaymentMethodLabel(row.PaymentMethod),
			utils.StatusLabel(row.Status),
		}
		for j, v := range values {
			cell, _ := excelize.CoordinatesToCellName(j+1, r)
			f.SetCellValue(sheet, cell, v)
		}
		if i%2 == 0 {
			firstCell, _ := excelize.CoordinatesToCellName(1, r)
			lastCell, _ := excelize.CoordinatesToCellName(len(headers), r)
			f.SetCellStyle(sheet, firstCell, lastCell, styleEven)
		}
	}

	// Totaux
	lastRow := len(report.Rows) + 3
	totalStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"F7DC6F"}, Pattern: 1},
	})
	totals := []interface{}{"TOTAL", "", fmt.Sprintf("%d factures", report.InvoiceCount),
		report.TotalHT, report.TotalTVA, report.TotalTTC, "", "", "", ""}
	for j, v := range totals {
		cell, _ := excelize.CoordinatesToCellName(j+1, lastRow)
		f.SetCellValue(sheet, cell, v)
		f.SetCellStyle(sheet, cell, cell, totalStyle)
	}

	buf, err := f.WriteToBuffer()
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// ─────────────────────────────────────────────────────────────────────────────
// EXPORT — Stock
// ─────────────────────────────────────────────────────────────────────────────

// ExportStockReport exporte un rapport de stock en Excel
func (s *ExcelService) ExportStockReport(report *services.StockReport) ([]byte, error) {
	f := excelize.NewFile()
	sheet := "Stock"
	f.SetSheetName("Sheet1", sheet)

	headers := []string{
		"Référence", "Désignation", "Famille", "Unité",
		"Stock", "Min", "CMUP", "P.V. TTC", "Valeur Stock", "État",
	}

	styleHeader, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Color: "FFFFFF"},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"8e44ad"}, Pattern: 1},
		Alignment: &excelize.Alignment{Horizontal: "center"},
	})

	for j, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(j+1, 1)
		f.SetCellValue(sheet, cell, h)
		f.SetCellStyle(sheet, cell, cell, styleHeader)
	}

	styleLow, _ := f.NewStyle(&excelize.Style{
		Fill: excelize.Fill{Type: "pattern", Color: []string{"FADBD8"}, Pattern: 1},
		Font: &excelize.Font{Color: "C0392B"},
	})

	for i, row := range report.Rows {
		r := i + 2
		values := []interface{}{
			row.Reference, row.Name, row.Category, row.Unit,
			row.StockQty, row.StockMin, row.CMUP, row.SalePriceTTC,
			row.StockValue, row.Status,
		}
		for j, v := range values {
			cell, _ := excelize.CoordinatesToCellName(j+1, r)
			f.SetCellValue(sheet, cell, v)
		}
		if row.Status != "OK" {
			firstCell, _ := excelize.CoordinatesToCellName(1, r)
			lastCell, _ := excelize.CoordinatesToCellName(len(headers), r)
			f.SetCellStyle(sheet, firstCell, lastCell, styleLow)
		}
	}

	// Total valeur
	lastRow := len(report.Rows) + 2
	totalStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"F7DC6F"}, Pattern: 1},
	})
	for j := 0; j < len(headers); j++ {
		cell, _ := excelize.CoordinatesToCellName(j+1, lastRow)
		if j == 0 {
			f.SetCellValue(sheet, cell, "VALEUR TOTALE STOCK")
		} else if j == 8 {
			f.SetCellValue(sheet, cell, report.TotalValue)
		}
		f.SetCellStyle(sheet, cell, cell, totalStyle)
	}

	buf, err := f.WriteToBuffer()
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// ─────────────────────────────────────────────────────────────────────────────
// IMPORT — Articles (depuis template Excel)
// ─────────────────────────────────────────────────────────────────────────────

// ImportArticles importe des articles depuis un fichier Excel
func (s *ExcelService) ImportArticles(data []byte) ([]models.Article, []string, error) {
	f, err := excelize.OpenReader(strings.NewReader(string(data)))
	if err != nil {
		// Essayer comme fichier
		f, err = excelize.OpenFile(string(data))
		if err != nil {
			return nil, nil, fmt.Errorf("impossible d'ouvrir le fichier Excel: %w", err)
		}
	}

	rows, err := f.GetRows("Articles")
	if err != nil {
		// Essayer la première feuille
		sheets := f.GetSheetList()
		if len(sheets) == 0 {
			return nil, nil, fmt.Errorf("aucune feuille trouvée")
		}
		rows, err = f.GetRows(sheets[0])
		if err != nil {
			return nil, nil, err
		}
	}

	var articles []models.Article
	var errors []string

	for i, row := range rows {
		if i == 0 {
			continue
		} // Skip header

		if len(row) < 4 {
			continue
		}

		a := models.Article{
			Reference:   safeGet(row, 0),
			Barcode:     safeGet(row, 1),
			NameFr:      safeGet(row, 2),
			NameAr:      safeGet(row, 3),
			IsActive:    true,
			TVARate:     19,
			ValuationMethod: "CMUP",
		}

		if a.Reference == "" || a.NameFr == "" {
			errors = append(errors, fmt.Sprintf("Ligne %d: référence ou désignation manquante", i+1))
			continue
		}

		// Prix
		if len(row) > 7 {
			a.PurchasePrice = parseFloat(safeGet(row, 7))
		}
		if len(row) > 9 {
			a.SalePriceHT = parseFloat(safeGet(row, 9))
		}
		if len(row) > 10 {
			a.SalePriceTTC = parseFloat(safeGet(row, 10))
		}
		if len(row) > 12 {
			a.TVARate = parseFloat(safeGet(row, 12))
		}
		if len(row) > 13 {
			a.StockQty = parseFloat(safeGet(row, 13))
		}
		if len(row) > 14 {
			a.StockMin = parseFloat(safeGet(row, 14))
		}

		articles = append(articles, a)
	}

	return articles, errors, nil
}

// ─────────────────────────────────────────────────────────────────────────────
// TEMPLATES
// ─────────────────────────────────────────────────────────────────────────────

// GenerateArticlesTemplate génère le template Excel pour l'import d'articles
func (s *ExcelService) GenerateArticlesTemplate() ([]byte, error) {
	f := excelize.NewFile()
	f.SetSheetName("Sheet1", "Articles")

	headers := []string{
		"Référence*", "Code-barres", "Désignation FR*", "Désignation AR",
		"Famille", "Marque", "Unité", "Prix Achat HT", "CMUP",
		"Prix Vente HT", "Prix Vente TTC", "Marge %", "TVA %",
		"Stock Initial", "Stock Min", "Stock Max",
	}

	styleHeader, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Color: "FFFFFF"},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"2980b9"}, Pattern: 1},
		Alignment: &excelize.Alignment{Horizontal: "center"},
	})

	for j, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(j+1, 1)
		f.SetCellValue("Articles", cell, h)
		f.SetCellStyle("Articles", cell, cell, styleHeader)
	}

	// Ligne d'exemple
	example := []interface{}{
		"ART-0001", "1234567890123", "Exemple Produit", "مثال منتج",
		"Alimentation", "Marque X", "U", 100.0, 100.0,
		130.0, 154.7, 30.0, 19.0, 0.0, 5.0, 100.0,
	}
	styleExample, _ := f.NewStyle(&excelize.Style{
		Fill: excelize.Fill{Type: "pattern", Color: []string{"EBF5FB"}, Pattern: 1},
		Font: &excelize.Font{Italic: true, Color: "7F8C8D"},
	})
	for j, v := range example {
		cell, _ := excelize.CoordinatesToCellName(j+1, 2)
		f.SetCellValue("Articles", cell, v)
		f.SetCellStyle("Articles", cell, cell, styleExample)
	}

	// Instructions
	f.NewSheet("Instructions")
	f.SetCellValue("Instructions", "A1", "Instructions d'import des articles")
	f.SetCellValue("Instructions", "A3", "1. Les champs marqués * sont obligatoires")
	f.SetCellValue("Instructions", "A4", "2. La référence doit être unique")
	f.SetCellValue("Instructions", "A5", "3. Le code-barres doit être unique (EAN-13 recommandé)")
	f.SetCellValue("Instructions", "A6", "4. TVA: 0, 9 ou 19")
	f.SetCellValue("Instructions", "A7", "5. Utilisez le point (.) comme séparateur décimal")

	buf, err := f.WriteToBuffer()
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// ─────────────────────────────────────────────────────────────────────────────
// Helpers
// ─────────────────────────────────────────────────────────────────────────────

func safeGet(row []string, idx int) string {
	if idx < len(row) {
		return strings.TrimSpace(row[idx])
	}
	return ""
}

func parseFloat(s string) float64 {
	s = strings.ReplaceAll(s, ",", ".")
	s = strings.ReplaceAll(s, " ", "")
	v, _ := strconv.ParseFloat(s, 64)
	return v
}
