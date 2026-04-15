package pdf

import (
	"database/sql"
	"fmt"
	"gestion-commerciale/internal/models"
	"gestion-commerciale/internal/services"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

// PDFService est le service principal de génération PDF
type PDFService struct {
	db      *sql.DB
	company *models.Company
}

// NewPDFService crée un service PDF
func NewPDFService(db *sql.DB) *PDFService {
	return &PDFService{db: db}
}

// SetCompany définit les informations de la société
func (s *PDFService) SetCompany(company *models.Company) {
	s.company = company
}

// LoadCompany charge les informations de la société depuis la DB
func (s *PDFService) LoadCompany() (*models.Company, error) {
	var c models.Company
	err := s.db.QueryRow(`
		SELECT id, name_ar, name_fr, activity, address, wilaya, commune,
		  postal_code, phone, mobile, fax, email, website,
		  nif, nis, rc, ai, rib, bank_name, capital,
		  logo_path, stamp_path, signature_path,
		  COALESCE((SELECT value FROM settings WHERE key='footer_text'),'') as footer_text
		FROM companies WHERE id=1`).
		Scan(&c.ID, &c.NameAr, &c.NameFr, &c.Activity, &c.Address, &c.Wilaya, &c.Commune,
			&c.PostalCode, &c.Phone, &c.Mobile, &c.Fax, &c.Email, &c.Website,
			&c.NIF, &c.NIS, &c.RC, &c.AI, &c.RIB, &c.BankName, &c.Capital,
			&c.LogoPath, &c.StampPath, &c.SignaturePath, &c.FooterText)
	if err != nil {
		return nil, err
	}
	s.company = &c
	return &c, nil
}

// ─────────────────────────────────────────────────────────────────────────────
// Méthodes de génération
// ─────────────────────────────────────────────────────────────────────────────

// GenerateDocumentPDF génère le PDF d'un document selon son type
func (s *PDFService) GenerateDocumentPDF(doc *models.Document) ([]byte, error) {
	company, err := s.getCompany()
	if err != nil {
		return nil, err
	}

	switch doc.DocType {
	case models.DocTypeFA:
		return GenerateSaleInvoice(doc, company), nil
	case models.DocTypeFAC:
		return GeneratePurchaseInvoice(doc, company), nil
	case models.DocTypeDV:
		return GenerateQuotation(doc, company), nil
	case models.DocTypePF:
		return GenerateProforma(doc, company), nil
	case models.DocTypeBL:
		return GenerateDeliveryNote(doc, company), nil
	case models.DocTypeBR:
		return GenerateReceptionNote(doc, company), nil
	case models.DocTypeBCC:
		return GenerateClientOrder(doc, company), nil
	case models.DocTypeBCF:
		return GenerateSupplierOrder(doc, company), nil
	case models.DocTypeAV:
		return GenerateCreditNote(doc, company), nil
	default:
		return GenerateSaleInvoice(doc, company), nil
	}
}

// GenerateTicketPDF génère un ticket de caisse
func (s *PDFService) GenerateTicketPDF(doc *models.Document) ([]byte, error) {
	company, err := s.getCompany()
	if err != nil {
		return nil, err
	}
	return GenerateTicket(doc, company), nil
}

// GenerateReceiptPDF génère un reçu de paiement
func (s *PDFService) GenerateReceiptPDF(payment *models.Payment) ([]byte, error) {
	company, err := s.getCompany()
	if err != nil {
		return nil, err
	}
	return GenerateReceipt(payment, company), nil
}

// GenerateSalesReportPDF génère un rapport de ventes
func (s *PDFService) GenerateSalesReportPDF(report *services.SalesReport, from, to string) ([]byte, error) {
	company, err := s.getCompany()
	if err != nil {
		return nil, err
	}
	return GenerateSalesReport(report, company, from, to), nil
}

// GenerateStockReportPDF génère un rapport de stock
func (s *PDFService) GenerateStockReportPDF(report *services.StockReport) ([]byte, error) {
	company, err := s.getCompany()
	if err != nil {
		return nil, err
	}
	return GenerateStockReport(report, company), nil
}

// GenerateClientStatementPDF génère un relevé de compte client
func (s *PDFService) GenerateClientStatementPDF(client *models.Client, lines []models.AccountStatement) ([]byte, error) {
	company, err := s.getCompany()
	if err != nil {
		return nil, err
	}
	return GenerateClientStatement(client, lines, company), nil
}

// GenerateCashJournalPDF génère le journal de caisse
func (s *PDFService) GenerateCashJournalPDF(date string, movements []models.CashMovement) ([]byte, error) {
	company, err := s.getCompany()
	if err != nil {
		return nil, err
	}
	return GenerateCashJournal(date, movements, company), nil
}

// GenerateG50PDF génère le formulaire G50
func (s *PDFService) GenerateG50PDF(g50 *models.G50Data) ([]byte, error) {
	company, err := s.getCompany()
	if err != nil {
		return nil, err
	}
	return GenerateG50(g50, company), nil
}

// GenerateInventoryPDF génère la fiche d'inventaire
func (s *PDFService) GenerateInventoryPDF(inventory *models.Inventory) ([]byte, error) {
	company, err := s.getCompany()
	if err != nil {
		return nil, err
	}
	return GenerateInventorySheet(inventory, company), nil
}

// ─────────────────────────────────────────────────────────────────────────────
// Impression
// ─────────────────────────────────────────────────────────────────────────────

// PrintDocument imprime un document sur l'imprimante par défaut
func (s *PDFService) PrintDocument(pdfBytes []byte, printerName string) error {
	// Sauvegarder dans un fichier temporaire
	tmpFile := filepath.Join(os.TempDir(), "gc_print_temp.pdf")
	if err := os.WriteFile(tmpFile, pdfBytes, 0644); err != nil {
		return fmt.Errorf("erreur création fichier temporaire: %w", err)
	}
	defer os.Remove(tmpFile)

	return printFile(tmpFile, printerName)
}

// printFile lance l'impression d'un fichier PDF
func printFile(filePath, printerName string) error {
	switch runtime.GOOS {
	case "windows":
		// Essayer SumatraPDF d'abord (silencieux)
		sumatrapdf, _ := exec.LookPath("SumatraPDF.exe")
		if sumatrapdf != "" {
			args := []string{"-print-to", printerName, filePath}
			if printerName == "" {
				args = []string{"-print-to-default", filePath}
			}
			return exec.Command(sumatrapdf, args...).Run()
		}
		// Sinon ouvrir avec le lecteur par défaut
		return exec.Command("cmd", "/c", "start", "/wait", "", filePath).Run()

	case "linux":
		if printerName != "" {
			return exec.Command("lpr", "-P", printerName, filePath).Run()
		}
		return exec.Command("lpr", filePath).Run()

	case "darwin":
		if printerName != "" {
			return exec.Command("lpr", "-P", printerName, filePath).Run()
		}
		return exec.Command("open", "-a", "Preview", filePath).Run()

	default:
		return fmt.Errorf("système d'exploitation non supporté pour l'impression: %s", runtime.GOOS)
	}
}

// OpenPDF ouvre un PDF dans le lecteur par défaut
func OpenPDF(pdfBytes []byte, filename string) error {
	tmpDir := filepath.Join(os.TempDir(), "gestion_commerciale")
	os.MkdirAll(tmpDir, 0755)

	filePath := filepath.Join(tmpDir, filename)
	if err := os.WriteFile(filePath, pdfBytes, 0644); err != nil {
		return err
	}

	switch runtime.GOOS {
	case "windows":
		return exec.Command("cmd", "/c", "start", "", filePath).Run()
	case "linux":
		return exec.Command("xdg-open", filePath).Run()
	case "darwin":
		return exec.Command("open", filePath).Run()
	}
	return nil
}

// SavePDF sauvegarde un PDF dans un fichier
func SavePDF(pdfBytes []byte, destPath string) error {
	return os.WriteFile(destPath, pdfBytes, 0644)
}

// GetSystemPrinters retourne la liste des imprimantes système
func GetSystemPrinters() []string {
	var printers []string

	switch runtime.GOOS {
	case "windows":
		out, err := exec.Command("wmic", "printer", "get", "name").Output()
		if err == nil {
			lines := splitLines(string(out))
			for i, l := range lines {
				if i == 0 {
					continue
				} // Skip header
				if l = trimStr(l); l != "" {
					printers = append(printers, l)
				}
			}
		}

	case "linux", "darwin":
		out, err := exec.Command("lpstat", "-a").Output()
		if err == nil {
			for _, l := range splitLines(string(out)) {
				if parts := splitBySpace(l); len(parts) > 0 {
					printers = append(printers, parts[0])
				}
			}
		}
	}

	if len(printers) == 0 {
		printers = []string{"Imprimante par défaut"}
	}
	return printers
}

// ─────────────────────────────────────────────────────────────────────────────
// Helpers internes
// ─────────────────────────────────────────────────────────────────────────────

func (s *PDFService) getCompany() (*models.Company, error) {
	if s.company != nil {
		return s.company, nil
	}
	return s.LoadCompany()
}

func splitLines(s string) []string {
	var lines []string
	current := ""
	for _, c := range s {
		if c == '\n' || c == '\r' {
			if current != "" {
				lines = append(lines, current)
				current = ""
			}
		} else {
			current += string(c)
		}
	}
	if current != "" {
		lines = append(lines, current)
	}
	return lines
}

func trimStr(s string) string {
	start, end := 0, len(s)
	for start < end && (s[start] == ' ' || s[start] == '\t') {
		start++
	}
	for end > start && (s[end-1] == ' ' || s[end-1] == '\t') {
		end--
	}
	return s[start:end]
}

func splitBySpace(s string) []string {
	var parts []string
	inWord := false
	word := ""
	for _, c := range s {
		if c == ' ' || c == '\t' {
			if inWord {
				parts = append(parts, word)
				word = ""
				inWord = false
			}
		} else {
			word += string(c)
			inWord = true
		}
	}
	if inWord {
		parts = append(parts, word)
	}
	return parts
}
