package pdf

import (
	"fmt"
	"gestion-commerciale/internal/models"
	"gestion-commerciale/pkg/utils"
	"math"
	"os"
	"path/filepath"

	"github.com/signintech/gopdf"
)

// ─────────────────────────────────────────────────────────────────────────────
// Constantes de mise en page A4 (en mm)
// ─────────────────────────────────────────────────────────────────────────────

const (
	PageWidth  = 210.0
	PageHeight = 297.0
	MarginL    = 15.0
	MarginR    = 15.0
	MarginT    = 15.0
	MarginB    = 15.0
	ContentW   = PageWidth - MarginL - MarginR
)

// Couleur RGB simplifiée
type RGB struct{ R, G, B uint8 }

var (
	ColorPrimary   = RGB{41, 128, 185}
	ColorSecondary = RGB{52, 73, 94}
	ColorSuccess   = RGB{39, 174, 96}
	ColorDanger    = RGB{231, 76, 60}
	ColorWarning   = RGB{243, 156, 18}
	ColorLight     = RGB{236, 240, 241}
	ColorTableHead = RGB{52, 73, 94}
	ColorTableRow  = RGB{248, 249, 250}
	ColorText      = RGB{33, 33, 33}
	ColorSubtext   = RGB{100, 100, 100}
	ColorWhite     = RGB{255, 255, 255}
)

// ─────────────────────────────────────────────────────────────────────────────
// PDFGenerator — Générateur PDF
// ─────────────────────────────────────────────────────────────────────────────

// PDFGenerator encapsule gopdf
type PDFGenerator struct {
	pdf     *gopdf.GoPdf
	fontDir string
	pageNum int
}

// NewPDFGenerator crée un nouveau générateur PDF A4
func NewPDFGenerator() *PDFGenerator {
	g := &PDFGenerator{
		pdf:     &gopdf.GoPdf{},
		fontDir: filepath.Join("assets", "fonts"),
	}
	g.pdf.Start(gopdf.Config{PageSize: gopdf.Rect{W: PageWidth, H: PageHeight}})
	g.loadFonts()
	return g
}

// loadFonts charge les polices
func (g *PDFGenerator) loadFonts() {
	// Essayer de charger les polices depuis le répertoire assets/fonts
	fonts := []struct{ name, file string }{
		{"Regular", "LiberationSans-Regular.ttf"},
		{"Bold", "LiberationSans-Bold.ttf"},
		{"Italic", "LiberationSans-Italic.ttf"},
	}
	for _, f := range fonts {
		path := filepath.Join(g.fontDir, f.file)
		if _, err := os.Stat(path); err == nil {
			g.pdf.AddTTFFont(f.name, path)
		}
	}
	// Police de secours (intégrée gopdf)
	g.setFont("Regular", 10)
}

// setFont définit la police (ignore l'erreur si la police n'existe pas)
func (g *PDFGenerator) setFont(style string, size float64) {
	if err := g.pdf.SetFont(style, "", size); err != nil {
		// Fallback — ignorer silencieusement
		_ = err
	}
}

// AddPage ajoute une page et retourne la position Y initiale
func (g *PDFGenerator) AddPage() float64 {
	g.pdf.AddPage()
	g.pageNum++
	return MarginT
}

// SetColor définit la couleur du texte
func (g *PDFGenerator) SetColor(c RGB) {
	g.pdf.SetTextColor(c.R, c.G, c.B)
}

// SetFill définit la couleur de remplissage
func (g *PDFGenerator) SetFill(c RGB) {
	g.pdf.SetFillColor(c.R, c.G, c.B)
}

// SetStroke définit la couleur de contour
func (g *PDFGenerator) SetStroke(c RGB) {
	g.pdf.SetStrokeColor(c.R, c.G, c.B)
}

// FilledRect dessine un rectangle rempli (x,y = coin haut-gauche)
func (g *PDFGenerator) FilledRect(x, y, w, h float64, fill RGB) {
	g.SetFill(fill)
	g.pdf.RectFromUpperLeftWithStyle(x, y, w, h, "F")
}

// OutlineRect dessine un rectangle avec contour
func (g *PDFGenerator) OutlineRect(x, y, w, h float64, stroke RGB) {
	g.SetStroke(stroke)
	g.pdf.RectFromUpperLeftWithStyle(x, y, w, h, "D")
}

// HLine dessine une ligne horizontale
func (g *PDFGenerator) HLine(x, y, w float64, stroke RGB) {
	g.SetStroke(stroke)
	g.pdf.Line(x, y, x+w, y)
}

// Text écrit du texte à (x, y)
func (g *PDFGenerator) Text(x, y float64, text string) {
	g.pdf.SetXY(x, y)
	g.pdf.Cell(nil, text)
}

// TextRight écrit du texte aligné à droite dans la largeur w
func (g *PDFGenerator) TextRight(x, y, w float64, text string) {
	tw, err := g.pdf.MeasureTextWidth(text)
	if err != nil || tw < 0 {
		tw = float64(len(text)) * 2.5 // approximation
	}
	g.pdf.SetXY(x+w-tw, y)
	g.pdf.Cell(nil, text)
}

// TextCenter écrit du texte centré dans la largeur w
func (g *PDFGenerator) TextCenter(x, y, w float64, text string) {
	tw, err := g.pdf.MeasureTextWidth(text)
	if err != nil || tw < 0 {
		tw = float64(len(text)) * 2.5
	}
	offset := (w - tw) / 2
	if offset < 0 {
		offset = 0
	}
	g.pdf.SetXY(x+offset, y)
	g.pdf.Cell(nil, text)
}

// GetBytes retourne le PDF en bytes
func (g *PDFGenerator) GetBytes() []byte {
	return g.pdf.GetBytesPdf()
}

// needNewPage vérifie si on a besoin d'une nouvelle page
func (g *PDFGenerator) needNewPage(y, needed float64) bool {
	return y+needed > PageHeight-MarginB-15
}

// ─────────────────────────────────────────────────────────────────────────────
// En-tête société commune à tous les documents
// ─────────────────────────────────────────────────────────────────────────────

// drawCompanyHeader dessine l'en-tête société, retourne la nouvelle position Y
func (g *PDFGenerator) drawCompanyHeader(company *models.Company) float64 {
	x := MarginL
	y := MarginT

	// Fond gris clair pour l'en-tête
	g.FilledRect(x-3, y-3, ContentW+6, 52, ColorLight)

	// Ligne bleue en bas de l'en-tête
	g.pdf.SetLineWidth(0.8)
	g.HLine(x-3, y+49, ContentW+6, ColorPrimary)

	// Logo
	logoW := 0.0
	if company.LogoPath != "" {
		if _, err := os.Stat(company.LogoPath); err == nil {
			g.pdf.Image(company.LogoPath, x, y+2, &gopdf.Rect{W: 40, H: 38})
			logoW = 44
		}
	}

	startX := x + logoW

	// Nom société
	g.setFont("Bold", 13)
	g.SetColor(ColorPrimary)
	g.Text(startX, y+3, company.NameFr)

	g.setFont("Regular", 9)
	g.SetColor(ColorSubtext)
	lineY := y + 13

	if company.NameAr != "" {
		g.Text(startX, lineY, company.NameAr)
		lineY += 7
	}
	if company.Activity != "" {
		g.setFont("Italic", 8)
		g.Text(startX, lineY, "Activité: "+company.Activity)
		g.setFont("Regular", 8)
		lineY += 6
	}
	if company.Address != "" {
		addr := company.Address
		if company.Wilaya != "" {
			addr += " — " + company.Wilaya
		}
		g.Text(startX, lineY, addr)
		lineY += 6
	}

	// Téléphone / Email
	contact := ""
	if company.Phone != "" {
		contact += "Tél: " + company.Phone
	}
	if company.Mobile != "" {
		if contact != "" {
			contact += "  |  "
		}
		contact += "Mob: " + company.Mobile
	}
	if company.Email != "" {
		if contact != "" {
			contact += "  |  "
		}
		contact += company.Email
	}
	if contact != "" {
		g.Text(startX, lineY, contact)
		lineY += 6
	}

	// Identifiants fiscaux
	g.setFont("Bold", 8)
	g.SetColor(ColorSecondary)
	fiscal := ""
	if company.NIF != "" {
		fiscal += "NIF: " + company.NIF
	}
	if company.NIS != "" {
		if fiscal != "" {
			fiscal += "   |   "
		}
		fiscal += "NIS: " + company.NIS
	}
	if fiscal != "" {
		g.Text(startX, lineY, fiscal)
		lineY += 6
	}

	fiscal2 := ""
	if company.RC != "" {
		fiscal2 += "RC: " + company.RC
	}
	if company.AI != "" {
		if fiscal2 != "" {
			fiscal2 += "   |   "
		}
		fiscal2 += "Art.Imp.: " + company.AI
	}
	if fiscal2 != "" {
		g.Text(startX, lineY, fiscal2)
	}

	g.SetColor(ColorText)
	return y + 55
}

// drawDocumentTitle dessine la barre de titre du document
func (g *PDFGenerator) drawDocumentTitle(title, docNumber, date string, y float64) float64 {
	g.FilledRect(MarginL-3, y, ContentW+6, 12, ColorPrimary)
	g.setFont("Bold", 12)
	g.SetColor(ColorWhite)
	g.TextCenter(MarginL, y+2, ContentW*0.5, title)

	g.setFont("Regular", 9)
	g.TextRight(MarginL+ContentW*0.5, y+1, ContentW*0.5-3, "N°: "+docNumber)
	g.TextRight(MarginL+ContentW*0.5, y+6, ContentW*0.5-3, "Date: "+utils.FormatDateFr(date))
	g.SetColor(ColorText)
	return y + 14
}

// drawClientBlock dessine le bloc client/fournisseur (côté droit)
func (g *PDFGenerator) drawClientBlock(label, name, address, nif, rc string, y float64) float64 {
	bX := MarginL + ContentW*0.52
	bW := ContentW * 0.48
	bH := 35.0

	// Entête du bloc
	g.FilledRect(bX, y, bW, 8, ColorPrimary)
	g.setFont("Bold", 8)
	g.SetColor(ColorWhite)
	g.TextCenter(bX, y+1, bW, label)

	// Contenu
	g.OutlineRect(bX, y+8, bW, bH-8, ColorLight)
	g.setFont("Bold", 9)
	g.SetColor(ColorText)
	g.Text(bX+3, y+11, truncStr(name, 28))

	g.setFont("Regular", 8)
	g.SetColor(ColorSubtext)
	ly := y + 19
	if address != "" {
		g.Text(bX+3, ly, truncStr(address, 28))
		ly += 6
	}
	ids := ""
	if nif != "" {
		ids += "NIF: " + nif
	}
	if rc != "" {
		if ids != "" {
			ids += "  |  "
		}
		ids += "RC: " + rc
	}
	if ids != "" {
		g.Text(bX+3, ly, ids)
	}

	g.SetColor(ColorText)
	return y + bH + 3
}

// drawPaymentInfo dessine les infos de paiement
func (g *PDFGenerator) drawPaymentInfo(paymentMethod, paymentTerms string, x, y float64) float64 {
	g.setFont("Regular", 8)
	g.SetColor(ColorSubtext)
	info := "Mode de paiement: " + utils.PaymentMethodLabel(paymentMethod)
	if paymentTerms != "" && paymentTerms != "immediate" {
		info += "   |   Délai: " + paymentTerms
	}
	g.Text(x, y, info)
	g.SetColor(ColorText)
	return y + 6
}

// ─────────────────────────────────────────────────────────────────────────────
// Tableau des lignes
// ─────────────────────────────────────────────────────────────────────────────

// TableColumn définit une colonne de tableau
type TableColumn struct {
	Header string
	Width  float64
	Align  string // "L", "C", "R"
}

// drawTableHeader dessine l'entête du tableau
func (g *PDFGenerator) drawTableHeader(cols []TableColumn, y float64) float64 {
	rowH := 8.0
	g.FilledRect(MarginL-3, y, ContentW+6, rowH, ColorTableHead)

	g.setFont("Bold", 7)
	g.SetColor(ColorWhite)

	x := MarginL
	for _, col := range cols {
		switch col.Align {
		case "R":
			g.TextRight(x, y+1, col.Width, col.Header)
		case "C":
			g.TextCenter(x, y+1, col.Width, col.Header)
		default:
			g.Text(x+1, y+1, col.Header)
		}
		x += col.Width
	}
	g.SetColor(ColorText)
	return y + rowH
}

// drawTableRow dessine une ligne du tableau
func (g *PDFGenerator) drawTableRow(cols []TableColumn, values []string, y float64, isEven bool) float64 {
	rowH := 7.0
	if isEven {
		g.FilledRect(MarginL-3, y, ContentW+6, rowH, ColorTableRow)
	}

	g.setFont("Regular", 7)
	g.SetColor(ColorText)

	x := MarginL
	for i, col := range cols {
		text := ""
		if i < len(values) {
			text = values[i]
		}
		switch col.Align {
		case "R":
			g.TextRight(x, y+1, col.Width, text)
		case "C":
			g.TextCenter(x, y+1, col.Width, text)
		default:
			g.Text(x+1, y+1, text)
		}
		x += col.Width
	}

	// Séparateur
	g.pdf.SetLineWidth(0.1)
	g.HLine(MarginL-3, y+rowH, ContentW+6, ColorLight)
	return y + rowH
}

// ─────────────────────────────────────────────────────────────────────────────
// Bloc des totaux
// ─────────────────────────────────────────────────────────────────────────────

// drawTotalsBlock dessine le bloc des totaux
func (g *PDFGenerator) drawTotalsBlock(doc *models.Document, y float64) float64 {
	totX := MarginL + ContentW*0.52
	totW := ContentW * 0.48
	lineH := 6.5

	rowCount := 5 // HT, Discount, NetHT, TVA, TTC
	if doc.Timbre > 0 {
		rowCount++
	}
	if doc.TVA9 > 0 {
		rowCount++
	}
	if doc.TVA19 > 0 {
		rowCount++
	}

	blockH := float64(rowCount)*lineH + 12
	g.FilledRect(totX-2, y-1, totW+2, blockH, ColorLight)

	drawTotRow := func(label, value string, bold bool, rowY float64) {
		if bold {
			g.setFont("Bold", 9)
		} else {
			g.setFont("Regular", 8)
		}
		g.SetColor(ColorSubtext)
		g.Text(totX, rowY, label)
		g.SetColor(ColorText)
		g.TextRight(totX, rowY, totW, value)
	}

	rowY := y
	drawTotRow("Total HT:", utils.FormatMoney(doc.TotalHT), false, rowY)
	rowY += lineH

	if doc.TotalDiscount > 0 {
		drawTotRow("Remises:", "-"+utils.FormatMoney(doc.TotalDiscount), false, rowY)
		rowY += lineH
	}

	drawTotRow("Net HT:", utils.FormatMoney(doc.NetHT), false, rowY)
	rowY += lineH

	if doc.TVA9 > 0 {
		drawTotRow("TVA 9%:", utils.FormatMoney(doc.TVA9), false, rowY)
		rowY += lineH
	}
	if doc.TVA19 > 0 {
		drawTotRow("TVA 19%:", utils.FormatMoney(doc.TVA19), false, rowY)
		rowY += lineH
	}

	drawTotRow("Total TVA:", utils.FormatMoney(doc.TotalTVA), false, rowY)
	rowY += lineH

	// Séparateur
	g.pdf.SetLineWidth(0.5)
	g.HLine(totX-2, rowY, totW+2, ColorPrimary)
	rowY += 1

	drawTotRow("Total TTC:", utils.FormatMoney(doc.TotalTTC), true, rowY)
	rowY += lineH

	if doc.Timbre > 0 {
		drawTotRow("Droit de timbre:", utils.FormatMoney(doc.Timbre), false, rowY)
		rowY += lineH
	}

	// Net à payer (fond bleu)
	g.FilledRect(totX-2, rowY-1, totW+2, 10, ColorPrimary)
	g.setFont("Bold", 9)
	g.SetColor(ColorWhite)
	g.Text(totX+1, rowY+0.5, "NET À PAYER:")
	g.TextRight(totX-1, rowY+0.5, totW, utils.FormatMoney(doc.NetAmount))
	g.SetColor(ColorText)
	return rowY + 11
}

// drawAmountInWords dessine le montant en lettres
func (g *PDFGenerator) drawAmountInWords(amount float64, y float64) float64 {
	g.setFont("Italic", 8)
	g.SetColor(ColorSubtext)
	words := "Arrêtée la présente à la somme de: " + utils.NumberToWordsFr(amount)
	g.Text(MarginL, y, truncStr(words, 80))
	g.SetColor(ColorText)
	return y + 7
}

// drawSignatureStamp dessine la zone signature/cachet
func (g *PDFGenerator) drawSignatureStamp(company *models.Company, y float64) float64 {
	sigX := MarginL + ContentW*0.52
	sigW := ContentW * 0.48

	g.setFont("Regular", 8)
	g.SetColor(ColorSubtext)
	g.Text(sigX, y, "Signature et cachet:")
	g.SetColor(ColorText)

	g.OutlineRect(sigX, y+5, sigW, 25, ColorLight)

	if company.StampPath != "" {
		if _, err := os.Stat(company.StampPath); err == nil {
			g.pdf.Image(company.StampPath, sigX+sigW-30, y+5, &gopdf.Rect{W: 28, H: 23})
		}
	}
	if company.SignaturePath != "" {
		if _, err := os.Stat(company.SignaturePath); err == nil {
			g.pdf.Image(company.SignaturePath, sigX+2, y+8, &gopdf.Rect{W: 25, H: 20})
		}
	}
	return y + 32
}

// drawFooter dessine le pied de page
func (g *PDFGenerator) drawFooter(company *models.Company, pageNum int) {
	footerY := PageHeight - MarginB - 12

	g.pdf.SetLineWidth(0.3)
	g.HLine(MarginL-3, footerY, ContentW+6, ColorPrimary)

	g.setFont("Regular", 7)
	g.SetColor(ColorSubtext)

	if company.FooterText != "" {
		g.TextCenter(MarginL, footerY+2, ContentW, company.FooterText)
	}
	if company.RIB != "" {
		rib := "RIB: " + company.RIB
		if company.BankName != "" {
			rib += " — " + company.BankName
		}
		g.TextCenter(MarginL, footerY+6, ContentW, rib)
	}

	g.TextRight(MarginL, footerY+2, ContentW, fmt.Sprintf("Page %d", pageNum))
	g.SetColor(ColorText)
}

// ─────────────────────────────────────────────────────────────────────────────
// Utilitaires
// ─────────────────────────────────────────────────────────────────────────────

func truncStr(s string, maxLen int) string {
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}
	return string(runes[:maxLen-2]) + ".."
}

func roundMoney(v float64) float64 {
	return math.Round(v*100) / 100
}
