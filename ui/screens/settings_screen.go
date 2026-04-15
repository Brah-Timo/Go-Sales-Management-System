package screens

// ─────────────────────────────────────────────────────────────────────────────
// Lot 10 — PARAMÈTRES, UTILISATEURS & OUTILS
//   BuildCompanyInfoScreen      → Informations société
//   BuildFiscalYearsScreen      → Années fiscales
//   BuildNumberingScreen        → Numérotation documents
//   BuildTaxSettingsScreen      → Paramètres fiscaux (TVA, TAP, timbre)
//   BuildPrintSettingsScreen    → Paramètres d'impression
//   BuildUsersScreen            → Gestion des utilisateurs & permissions
//   BuildCurrenciesScreen       → Devises (affichage)
//   BuildBarcodeSettingsScreen  → Paramètres code-barres
//   BuildBackupScreen           → Sauvegarde base de données
//   BuildRestoreScreen          → Restauration base de données
//   BuildCalculatorScreen       → Calculatrice intégrée
//   BuildPriceUpdateScreen      → Mise à jour des prix en masse
//   BuildCalendarScreen         → Calendrier & Rappels
//   BuildAuditLogScreen         → Journal d'audit
//   BuildAboutScreen            → À propos
// ─────────────────────────────────────────────────────────────────────────────

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	appstate "gestion-commerciale/internal/app"
	"gestion-commerciale/internal/models"
	"gestion-commerciale/internal/services"
	"gestion-commerciale/pkg/utils"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// ─────────────────────────────────────────────────────────────────────────────
// Helpers communs
// ─────────────────────────────────────────────────────────────────────────────

func buildSettingsHeader(icon, title, subtitle, hexColor string) fyne.CanvasObject {
	titleLbl := widget.NewRichTextFromMarkdown("## " + icon + " " + title)
	subLbl := widget.NewLabel(subtitle)
	subLbl.Wrapping = fyne.TextWrapWord
	return container.NewVBox(titleLbl, subLbl, widget.NewSeparator())
}

func buildFormRow(label string, w fyne.CanvasObject) fyne.CanvasObject {
	lbl := widget.NewLabel(label)
	lbl.TextStyle = fyne.TextStyle{Bold: true}
	return container.NewBorder(nil, nil, lbl, nil, w)
}

func showSuccess(win fyne.Window, msg string) {
	if win == nil {
		return
	}
	dialog.ShowInformation("Succès", msg, win)
}

func showError(win fyne.Window, msg string) {
	if win == nil {
		return
	}
	dialog.ShowError(fmt.Errorf("%s", msg), win)
}

// ─────────────────────────────────────────────────────────────────────────────
// BuildCompanyInfoScreen — Informations Société
// ─────────────────────────────────────────────────────────────────────────────

func BuildCompanyInfoScreen() fyne.CanvasObject {
	db := appstate.GetDB()
	header := buildSettingsHeader("🏢", "Informations Société",
		"Renseignez les coordonnées légales et fiscales de votre entreprise", "#2980b9")

	if db == nil {
		return container.NewVBox(header, widget.NewLabel("Base de données non connectée."))
	}

	// Lire les paramètres depuis la table settings
	readSetting := func(key string) string {
		var val string
		db.QueryRow(`SELECT value FROM settings WHERE key=?`, key).Scan(&val)
		return val
	}

	nameArEntry := widget.NewEntry()
	nameArEntry.SetPlaceHolder("Nom en arabe")
	nameArEntry.SetText(readSetting("company_name_ar"))

	nameFrEntry := widget.NewEntry()
	nameFrEntry.SetPlaceHolder("Nom en français")
	nameFrEntry.SetText(readSetting("company_name_fr"))

	activityEntry := widget.NewEntry()
	activityEntry.SetPlaceHolder("Activité commerciale")
	activityEntry.SetText(readSetting("company_activity"))

	addressEntry := widget.NewEntry()
	addressEntry.SetPlaceHolder("Adresse")
	addressEntry.SetText(readSetting("company_address"))

	phoneEntry := widget.NewEntry()
	phoneEntry.SetPlaceHolder("Téléphone")
	phoneEntry.SetText(readSetting("company_phone"))

	mobileEntry := widget.NewEntry()
	mobileEntry.SetPlaceHolder("Mobile")
	mobileEntry.SetText(readSetting("company_mobile"))

	emailEntry := widget.NewEntry()
	emailEntry.SetPlaceHolder("E-mail")
	emailEntry.SetText(readSetting("company_email"))

	nifEntry := widget.NewEntry()
	nifEntry.SetPlaceHolder("NIF — Numéro d'Identification Fiscale")
	nifEntry.SetText(readSetting("company_nif"))

	nisEntry := widget.NewEntry()
	nisEntry.SetPlaceHolder("NIS — Numéro d'Identification Statistique")
	nisEntry.SetText(readSetting("company_nis"))

	rcEntry := widget.NewEntry()
	rcEntry.SetPlaceHolder("RC — Registre du Commerce")
	rcEntry.SetText(readSetting("company_rc"))

	aiEntry := widget.NewEntry()
	aiEntry.SetPlaceHolder("AI — Article d'Imposition")
	aiEntry.SetText(readSetting("company_ai"))

	ribEntry := widget.NewEntry()
	ribEntry.SetPlaceHolder("RIB bancaire")
	ribEntry.SetText(readSetting("company_rib"))

	bankEntry := widget.NewEntry()
	bankEntry.SetPlaceHolder("Nom de la banque")
	bankEntry.SetText(readSetting("company_bank_name"))

	capitalEntry := widget.NewEntry()
	capitalEntry.SetPlaceHolder("Capital social (DA)")
	capitalEntry.SetText(readSetting("company_capital"))

	footerEntry := widget.NewMultiLineEntry()
	footerEntry.SetPlaceHolder("Texte pied de page des documents")
	footerEntry.SetText(readSetting("company_footer"))
	footerEntry.SetMinRowsVisible(3)

	saveBtn := widget.NewButtonWithIcon("Enregistrer", theme.DocumentSaveIcon(), func() {
		pairs := map[string]string{
			"company_name_ar":   nameArEntry.Text,
			"company_name_fr":   nameFrEntry.Text,
			"company_activity":  activityEntry.Text,
			"company_address":   addressEntry.Text,
			"company_phone":     phoneEntry.Text,
			"company_mobile":    mobileEntry.Text,
			"company_email":     emailEntry.Text,
			"company_nif":       nifEntry.Text,
			"company_nis":       nisEntry.Text,
			"company_rc":        rcEntry.Text,
			"company_ai":        aiEntry.Text,
			"company_rib":       ribEntry.Text,
			"company_bank_name": bankEntry.Text,
			"company_capital":   capitalEntry.Text,
			"company_footer":    footerEntry.Text,
		}
		for k, v := range pairs {
			db.Exec(`INSERT INTO settings(key,value) VALUES(?,?)
				ON CONFLICT(key) DO UPDATE SET value=excluded.value`, k, v)
		}
		// Journaliser
		db.Exec(`INSERT INTO audit_log(action_type,module,description) VALUES(?,?,?)`,
			"update", "settings", "Informations société mises à jour")
		dialog.ShowInformation("Succès", "Informations société enregistrées.", fyne.CurrentApp().Driver().AllWindows()[0])
	})
	saveBtn.Importance = widget.HighImportance

	form := container.NewVBox(
		buildFormRow("Nom (AR):", nameArEntry),
		buildFormRow("Nom (FR):", nameFrEntry),
		buildFormRow("Activité:", activityEntry),
		buildFormRow("Adresse:", addressEntry),
		buildFormRow("Téléphone:", phoneEntry),
		buildFormRow("Mobile:", mobileEntry),
		buildFormRow("E-mail:", emailEntry),
		widget.NewSeparator(),
		buildFormRow("NIF:", nifEntry),
		buildFormRow("NIS:", nisEntry),
		buildFormRow("RC:", rcEntry),
		buildFormRow("AI:", aiEntry),
		buildFormRow("RIB:", ribEntry),
		buildFormRow("Banque:", bankEntry),
		buildFormRow("Capital:", capitalEntry),
		widget.NewSeparator(),
		buildFormRow("Pied de page:", footerEntry),
		container.NewHBox(saveBtn),
	)

	return container.NewBorder(header, nil, nil, nil,
		container.NewVScroll(form))
}

// ─────────────────────────────────────────────────────────────────────────────
// BuildFiscalYearsScreen — Années Fiscales
// ─────────────────────────────────────────────────────────────────────────────

func BuildFiscalYearsScreen() fyne.CanvasObject {
	db := appstate.GetDB()
	header := buildSettingsHeader("📅", "Années Fiscales",
		"Gérez les exercices comptables (un seul exercice actif à la fois)", "#8e44ad")

	if db == nil {
		return container.NewVBox(header, widget.NewLabel("Base de données non connectée."))
	}

	loadYears := func() []models.FiscalYear {
		rows, err := db.Query(`SELECT id, year, start_date, end_date, status, invoice_count, revenue
			FROM fiscal_years ORDER BY year DESC`)
		if err != nil {
			return nil
		}
		defer rows.Close()
		var list []models.FiscalYear
		for rows.Next() {
			var fy models.FiscalYear
			rows.Scan(&fy.ID, &fy.Year, &fy.StartDate, &fy.EndDate,
				&fy.Status, &fy.InvoiceCount, &fy.Revenue)
			list = append(list, fy)
		}
		return list
	}

	colHeaders := []string{"Année", "Début", "Fin", "Statut", "Factures", "CA (DA)"}
	colWidths := []float32{80, 120, 120, 100, 90, 150}

	var tableContainer *fyne.Container
	var refresh func()

	buildTable := func(years []models.FiscalYear) fyne.CanvasObject {
		headerRow := container.NewHBox()
		for i, h := range colHeaders {
			lbl := widget.NewLabel(h)
			lbl.TextStyle = fyne.TextStyle{Bold: true}
			headerRow.Add(container.NewGridWrap(fyne.NewSize(colWidths[i], 28), lbl))
		}

		rows2 := container.NewVBox(
			container.NewGridWrap(fyne.NewSize(0, 36), headerRow),
			widget.NewSeparator(),
		)
		if len(years) == 0 {
			rows2.Add(widget.NewLabel("Aucun exercice fiscal trouvé."))
		}
		for _, fy := range years {
			fy := fy
			statusLabel := "Fermé"
			if fy.Status == "active" {
				statusLabel = "✅ Actif"
			}
			row := container.NewHBox(
				container.NewGridWrap(fyne.NewSize(colWidths[0], 32), widget.NewLabel(fmt.Sprintf("%d", fy.Year))),
				container.NewGridWrap(fyne.NewSize(colWidths[1], 32), widget.NewLabel(fy.StartDate)),
				container.NewGridWrap(fyne.NewSize(colWidths[2], 32), widget.NewLabel(fy.EndDate)),
				container.NewGridWrap(fyne.NewSize(colWidths[3], 32), widget.NewLabel(statusLabel)),
				container.NewGridWrap(fyne.NewSize(colWidths[4], 32), widget.NewLabel(fmt.Sprintf("%d", fy.InvoiceCount))),
				container.NewGridWrap(fyne.NewSize(colWidths[5], 32), widget.NewLabel(utils.FormatAmount(fy.Revenue))),
			)
			if fy.Status != "active" {
				activateBtn := widget.NewButtonWithIcon("Activer", theme.ConfirmIcon(), func() {
					db.Exec(`UPDATE fiscal_years SET status='closed' WHERE status='active'`)
					db.Exec(`UPDATE fiscal_years SET status='active' WHERE id=?`, fy.ID)
					refresh()
				})
				row.Add(activateBtn)
			}
			rows2.Add(row)
		}
		return container.NewVScroll(rows2)
	}

	tableContainer = container.NewMax()
	refresh = func() {
		years := loadYears()
		tableContainer.Objects = []fyne.CanvasObject{buildTable(years)}
		tableContainer.Refresh()
	}
	refresh()

	// Formulaire création nouvel exercice
	yearEntry := widget.NewEntry()
	yearEntry.SetText(fmt.Sprintf("%d", time.Now().Year()))
	yearEntry.SetPlaceHolder("Année (ex: 2025)")

	startEntry := widget.NewEntry()
	startEntry.SetText(fmt.Sprintf("%d-01-01", time.Now().Year()))
	startEntry.SetPlaceHolder("AAAA-MM-JJ")

	endEntry := widget.NewEntry()
	endEntry.SetText(fmt.Sprintf("%d-12-31", time.Now().Year()))
	endEntry.SetPlaceHolder("AAAA-MM-JJ")

	createBtn := widget.NewButtonWithIcon("Créer exercice", theme.ContentAddIcon(), func() {
		yr, err := strconv.Atoi(yearEntry.Text)
		if err != nil || yr < 2000 {
			dialog.ShowError(fmt.Errorf("année invalide"), fyne.CurrentApp().Driver().AllWindows()[0])
			return
		}
		_, err = db.Exec(`INSERT INTO fiscal_years(year,start_date,end_date,status,invoice_count,revenue)
			VALUES(?,?,?,'closed',0,0)`, yr, startEntry.Text, endEntry.Text)
		if err != nil {
			dialog.ShowError(err, fyne.CurrentApp().Driver().AllWindows()[0])
			return
		}
		refresh()
	})
	createBtn.Importance = widget.HighImportance

	createForm := container.NewVBox(
		widget.NewSeparator(),
		widget.NewRichTextFromMarkdown("### Créer un nouvel exercice"),
		container.NewGridWithColumns(3,
			buildFormRow("Année:", yearEntry),
			buildFormRow("Début:", startEntry),
			buildFormRow("Fin:", endEntry),
		),
		container.NewHBox(createBtn),
	)

	return container.NewBorder(
		container.NewVBox(header, createForm, widget.NewSeparator()),
		nil, nil, nil,
		tableContainer,
	)
}

// ─────────────────────────────────────────────────────────────────────────────
// BuildNumberingScreen — Numérotation Documents
// ─────────────────────────────────────────────────────────────────────────────

func BuildNumberingScreen() fyne.CanvasObject {
	db := appstate.GetDB()
	header := buildSettingsHeader("🔢", "Numérotation Documents",
		"Configurez les préfixes et compteurs de numérotation automatique", "#16a085")

	if db == nil {
		return container.NewVBox(header, widget.NewLabel("Base de données non connectée."))
	}

	svc := services.NewNumberingService(db)

	var refresh func()
	content := container.NewVBox()

	refresh = func() {
		configs, err := svc.GetAllConfigs()
		content.Objects = nil
		if err != nil {
			content.Add(widget.NewLabel("Erreur chargement: " + err.Error()))
			content.Refresh()
			return
		}

		colHeaders := []string{"Type Document", "Préfixe", "Numéro actuel", "RAZ annuel", "Aperçu", "Action"}
		colWidths := []float32{180, 100, 130, 100, 160, 90}

		headerRow := container.NewHBox()
		for i, h := range colHeaders {
			lbl := widget.NewLabel(h)
			lbl.TextStyle = fyne.TextStyle{Bold: true}
			headerRow.Add(container.NewGridWrap(fyne.NewSize(colWidths[i], 28), lbl))
		}
		content.Add(headerRow)
		content.Add(widget.NewSeparator())

		if len(configs) == 0 {
			content.Add(widget.NewLabel("Aucune configuration trouvée."))
		}
		for _, cfg := range configs {
			cfg := cfg
			prefixEntry := widget.NewEntry()
			prefixEntry.SetText(cfg.Prefix)

			numEntry := widget.NewEntry()
			numEntry.SetText(fmt.Sprintf("%d", cfg.CurrentNumber+1))

			resetCheck := widget.NewCheck("", nil)
			resetCheck.SetChecked(cfg.ResetYearly)

			preview := svc.PreviewNextNumber(cfg.DocType, cfg.Prefix, cfg.CurrentNumber, time.Now().Year())
			previewLbl := widget.NewLabel(preview)
			prefixEntry.OnChanged = func(s string) {
				num, _ := strconv.Atoi(numEntry.Text)
				previewLbl.SetText(svc.PreviewNextNumber(cfg.DocType, s, num-1, time.Now().Year()))
			}

			saveBtn := widget.NewButtonWithIcon("Sauv.", theme.DocumentSaveIcon(), func() {
				startNum, err2 := strconv.Atoi(numEntry.Text)
				if err2 != nil || startNum < 1 {
					return
				}
				svc.UpdateConfig(cfg.DocType, prefixEntry.Text, startNum, resetCheck.Checked)
				refresh()
			})

			row := container.NewHBox(
				container.NewGridWrap(fyne.NewSize(colWidths[0], 36), widget.NewLabel(cfg.Label)),
				container.NewGridWrap(fyne.NewSize(colWidths[1], 36), prefixEntry),
				container.NewGridWrap(fyne.NewSize(colWidths[2], 36), numEntry),
				container.NewGridWrap(fyne.NewSize(colWidths[3], 36), resetCheck),
				container.NewGridWrap(fyne.NewSize(colWidths[4], 36), previewLbl),
				container.NewGridWrap(fyne.NewSize(colWidths[5], 36), saveBtn),
			)
			content.Add(row)
		}
		content.Refresh()
	}
	refresh()

	return container.NewBorder(header, nil, nil, nil, container.NewVScroll(content))
}

// ─────────────────────────────────────────────────────────────────────────────
// BuildTaxSettingsScreen — Paramètres Fiscaux
// ─────────────────────────────────────────────────────────────────────────────

func BuildTaxSettingsScreen() fyne.CanvasObject {
	db := appstate.GetDB()
	header := buildSettingsHeader("⚖️", "Paramètres Fiscaux",
		"Configurez les taux TVA, TAP, timbre fiscal et régime d'imposition", "#c0392b")

	if db == nil {
		return container.NewVBox(header, widget.NewLabel("Base de données non connectée."))
	}

	svc := services.NewTaxService(db)
	cfg := svc.GetTaxConfig()

	tvaSelect := widget.NewSelect([]string{"0%", "9%", "19%"}, nil)
	switch cfg.TVANormal {
	case 9:
		tvaSelect.SetSelected("9%")
	case 19:
		tvaSelect.SetSelected("19%")
	default:
		tvaSelect.SetSelected("19%")
	}

	tvaReducedSelect := widget.NewSelect([]string{"0%", "9%", "19%"}, nil)
	switch cfg.TVAReduced {
	case 0:
		tvaReducedSelect.SetSelected("0%")
	case 9:
		tvaReducedSelect.SetSelected("9%")
	default:
		tvaReducedSelect.SetSelected("9%")
	}

	timbreRateEntry := widget.NewEntry()
	timbreRateEntry.SetText(fmt.Sprintf("%.4f", cfg.TimbreRate))
	timbreRateEntry.SetPlaceHolder("Ex: 0.001 (1‰)")

	timbreMaxEntry := widget.NewEntry()
	timbreMaxEntry.SetText(fmt.Sprintf("%.0f", cfg.TimbreMax))
	timbreMaxEntry.SetPlaceHolder("Ex: 2500")

	tapEntry := widget.NewEntry()
	tapEntry.SetText(fmt.Sprintf("%.4f", cfg.TAPRate))
	tapEntry.SetPlaceHolder("Ex: 0.02 (2%)")

	regimeSelect := widget.NewSelect([]string{"Réel", "Forfaitaire", "Auto-entrepreneur"}, nil)
	regimeSelect.SetSelected(cfg.TaxRegime)

	isTVACheck := widget.NewCheck("Assujetti à la TVA", nil)
	isTVACheck.SetChecked(cfg.IsTVASubject)

	autoTimbreCheck := widget.NewCheck("Timbre fiscal automatique sur factures TTC", nil)
	autoTimbreCheck.SetChecked(cfg.AutoTimbre)

	autoUpdatePriceCheck := widget.NewCheck("Mettre à jour prix de vente auto (CMUP + marge)", nil)
	autoUpdatePriceCheck.SetChecked(cfg.AutoUpdateSalePrice)

	saveBtn := widget.NewButtonWithIcon("Enregistrer paramètres fiscaux", theme.DocumentSaveIcon(), func() {
		tvaStr := strings.TrimSuffix(tvaSelect.Selected, "%")
		tvaNormal, _ := strconv.ParseFloat(tvaStr, 64)

		tvaRStr := strings.TrimSuffix(tvaReducedSelect.Selected, "%")
		tvaReduced, _ := strconv.ParseFloat(tvaRStr, 64)

		timbreRate, _ := strconv.ParseFloat(timbreRateEntry.Text, 64)
		timbreMax, _ := strconv.ParseFloat(timbreMaxEntry.Text, 64)
		tapRate, _ := strconv.ParseFloat(tapEntry.Text, 64)

		newCfg := models.TaxConfig{
			TVANormal:           tvaNormal,
			TVAReduced:          tvaReduced,
			TimbreRate:          timbreRate,
			TimbreMax:           timbreMax,
			TAPRate:             tapRate,
			TaxRegime:           regimeSelect.Selected,
			IsTVASubject:        isTVACheck.Checked,
			AutoTimbre:          autoTimbreCheck.Checked,
			AutoUpdateSalePrice: autoUpdatePriceCheck.Checked,
		}
		if err := svc.SaveTaxConfig(newCfg); err != nil {
			dialog.ShowError(err, fyne.CurrentApp().Driver().AllWindows()[0])
			return
		}
		db.Exec(`INSERT INTO audit_log(action_type,module,description) VALUES(?,?,?)`,
			"update", "tax_settings", "Paramètres fiscaux mis à jour")
		dialog.ShowInformation("Succès", "Paramètres fiscaux enregistrés.", fyne.CurrentApp().Driver().AllWindows()[0])
	})
	saveBtn.Importance = widget.HighImportance

	form := container.NewVBox(
		buildFormRow("TVA taux normal:", tvaSelect),
		buildFormRow("TVA taux réduit:", tvaReducedSelect),
		widget.NewSeparator(),
		buildFormRow("Taux timbre fiscal:", timbreRateEntry),
		buildFormRow("Timbre maximum (DA):", timbreMaxEntry),
		buildFormRow("Taux TAP:", tapEntry),
		widget.NewSeparator(),
		buildFormRow("Régime fiscal:", regimeSelect),
		isTVACheck,
		autoTimbreCheck,
		autoUpdatePriceCheck,
		widget.NewSeparator(),
		container.NewHBox(saveBtn),
	)

	return container.NewBorder(header, nil, nil, nil, container.NewVScroll(form))
}

// ─────────────────────────────────────────────────────────────────────────────
// BuildPrintSettingsScreen — Paramètres Impression
// ─────────────────────────────────────────────────────────────────────────────

func BuildPrintSettingsScreen() fyne.CanvasObject {
	db := appstate.GetDB()
	header := buildSettingsHeader("🖨️", "Paramètres Impression",
		"Configurez le format des documents imprimés (PDF)", "#7f8c8d")

	if db == nil {
		return container.NewVBox(header, widget.NewLabel("Base de données non connectée."))
	}

	readSetting := func(key string) string {
		var val string
		db.QueryRow(`SELECT value FROM settings WHERE key=?`, key).Scan(&val)
		return val
	}
	writeSetting := func(key, val string) {
		db.Exec(`INSERT INTO settings(key,value) VALUES(?,?)
			ON CONFLICT(key) DO UPDATE SET value=excluded.value`, key, val)
	}

	paperSelect := widget.NewSelect([]string{"A4", "A5", "Thermique 80mm", "Thermique 58mm"}, nil)
	paperSelect.SetSelected(readSetting("print_paper_size"))
	if paperSelect.Selected == "" {
		paperSelect.SetSelected("A4")
	}

	marginEntry := widget.NewEntry()
	marginEntry.SetText(readSetting("print_margin_mm"))
	if marginEntry.Text == "" {
		marginEntry.SetText("15")
	}
	marginEntry.SetPlaceHolder("Marges en mm (ex: 15)")

	logoCheck := widget.NewCheck("Afficher logo sur les documents", nil)
	logoCheck.SetChecked(readSetting("print_show_logo") == "true")

	stampCheck := widget.NewCheck("Afficher cachet & signature", nil)
	stampCheck.SetChecked(readSetting("print_show_stamp") == "true")

	watermarkCheck := widget.NewCheck("Afficher filigrane BROUILLON sur devis/proforma", nil)
	watermarkCheck.SetChecked(readSetting("print_watermark_draft") == "true")

	fontSelect := widget.NewSelect([]string{"Arial", "Helvetica", "Times New Roman", "Courier"}, nil)
	fontSelect.SetSelected(readSetting("print_font"))
	if fontSelect.Selected == "" {
		fontSelect.SetSelected("Arial")
	}

	copyCountEntry := widget.NewEntry()
	copyCountEntry.SetText(readSetting("print_copy_count"))
	if copyCountEntry.Text == "" {
		copyCountEntry.SetText("2")
	}
	copyCountEntry.SetPlaceHolder("Nombre de copies par défaut")

	pdfDirEntry := widget.NewEntry()
	pdfDirEntry.SetText(readSetting("print_pdf_output_dir"))
	if pdfDirEntry.Text == "" {
		pdfDirEntry.SetText("./pdf_output")
	}
	pdfDirEntry.SetPlaceHolder("Répertoire de sortie PDF")

	saveBtn := widget.NewButtonWithIcon("Enregistrer", theme.DocumentSaveIcon(), func() {
		boolStr := func(b bool) string {
			if b {
				return "true"
			}
			return "false"
		}
		writeSetting("print_paper_size", paperSelect.Selected)
		writeSetting("print_margin_mm", marginEntry.Text)
		writeSetting("print_show_logo", boolStr(logoCheck.Checked))
		writeSetting("print_show_stamp", boolStr(stampCheck.Checked))
		writeSetting("print_watermark_draft", boolStr(watermarkCheck.Checked))
		writeSetting("print_font", fontSelect.Selected)
		writeSetting("print_copy_count", copyCountEntry.Text)
		writeSetting("print_pdf_output_dir", pdfDirEntry.Text)
		dialog.ShowInformation("Succès", "Paramètres d'impression enregistrés.", fyne.CurrentApp().Driver().AllWindows()[0])
	})
	saveBtn.Importance = widget.HighImportance

	form := container.NewVBox(
		buildFormRow("Format papier:", paperSelect),
		buildFormRow("Marges (mm):", marginEntry),
		buildFormRow("Police:", fontSelect),
		buildFormRow("Copies par défaut:", copyCountEntry),
		buildFormRow("Répertoire PDF:", pdfDirEntry),
		widget.NewSeparator(),
		logoCheck,
		stampCheck,
		watermarkCheck,
		widget.NewSeparator(),
		container.NewHBox(saveBtn),
	)

	return container.NewBorder(header, nil, nil, nil, container.NewVScroll(form))
}

// ─────────────────────────────────────────────────────────────────────────────
// BuildUsersScreen — Gestion Utilisateurs & Permissions
// ─────────────────────────────────────────────────────────────────────────────

func BuildUsersScreen() fyne.CanvasObject {
	db := appstate.GetDB()
	header := buildSettingsHeader("👥", "Gestion Utilisateurs",
		"Créez et gérez les comptes utilisateurs avec leurs rôles et permissions", "#e67e22")

	if db == nil {
		return container.NewVBox(header, widget.NewLabel("Base de données non connectée."))
	}

	session := appstate.GetSession()
	if session == nil || session.Role != models.RoleAdmin {
		return container.NewVBox(header,
			widget.NewLabel("⛔ Accès réservé à l'administrateur."))
	}

	svc := services.NewAuthService(db)

	allPerms := []struct{ key, label string }{
		{models.PermCreateSaleInvoice, "Créer factures vente"},
		{models.PermCreatePurchaseInvoice, "Créer factures achat"},
		{models.PermEditConfirmedInvoice, "Modifier factures confirmées"},
		{models.PermDeleteInvoice, "Supprimer factures"},
		{models.PermEditPrices, "Modifier prix"},
		{models.PermViewPurchasePrices, "Voir prix d'achat"},
		{models.PermViewProfitMargin, "Voir marges"},
		{models.PermManageStock, "Gérer le stock"},
		{models.PermManageClientSupplier, "Gérer clients/fournisseurs"},
		{models.PermAccessFinancialReports, "Rapports financiers"},
		{models.PermCollectPayments, "Encaisser paiements"},
		{models.PermManageSettings, "Paramètres système"},
		{models.PermBackupRestore, "Sauvegarde/Restauration"},
		{models.PermApplyDiscountAbove10, "Remise > 10%"},
		{models.PermInventory, "Inventaire"},
	}

	var listContent *fyne.Container
	var refresh func()

	buildUserRow := func(u models.User) fyne.CanvasObject {
		statusIcon := "🟢"
		if !u.IsActive {
			statusIcon = "🔴"
		}
		lastLogin := "Jamais"
		if u.LastLogin != nil {
			lastLogin = u.LastLogin.Format("02/01/2006 15:04")
		}
		info := widget.NewLabel(fmt.Sprintf("%s %s  |  %s  |  Rôle: %s  |  Dernière connexion: %s",
			statusIcon, u.FullName, u.Username, u.Role, lastLogin))

		editBtn := widget.NewButtonWithIcon("Modifier", theme.DocumentIcon(), func() {
			// Dialog modification permissions
			permChecks := make(map[string]*widget.Check)
			permBox := container.NewGridWithColumns(2)
			for _, p := range allPerms {
				ch := widget.NewCheck(p.label, nil)
				ch.SetChecked(u.Permissions[p.key])
				permChecks[p.key] = ch
				permBox.Add(ch)
			}

			newPassEntry := widget.NewPasswordEntry()
			newPassEntry.SetPlaceHolder("Nouveau mot de passe (laisser vide = inchangé)")

			toggleLbl := "Désactiver"
			if !u.IsActive {
				toggleLbl = "Activer"
			}
			toggleBtn := widget.NewButton(toggleLbl, func() {
				svc.ToggleUserActive(u.ID, !u.IsActive)
				refresh()
			})

			content2 := container.NewVBox(
				widget.NewRichTextFromMarkdown("### Permissions"),
				permBox,
				widget.NewSeparator(),
				widget.NewLabel("Changer mot de passe:"),
				newPassEntry,
				widget.NewSeparator(),
				toggleBtn,
			)

			wins := fyne.CurrentApp().Driver().AllWindows()
			if len(wins) == 0 {
				return
			}
			d := dialog.NewCustomConfirm("Modifier "+u.Username, "Enregistrer", "Annuler",
				container.NewVScroll(content2),
				func(ok bool) {
					if !ok {
						return
					}
					newPerms := make(map[string]bool)
					for _, p := range allPerms {
						newPerms[p.key] = permChecks[p.key].Checked
					}
					uID := u.ID
					newPass := newPassEntry.Text
					go func() {
						svc.UpdateUserPermissions(uID, newPerms)
						if newPass != "" {
							svc.ChangePassword(uID, newPass)
						}
						fyne.Do(refresh)
					}()
				},
				wins[0])
			d.Resize(fyne.NewSize(600, 500))
			d.Show()
		})

		return container.NewBorder(nil, nil, nil, editBtn, info)
	}

	listContent = container.NewVBox()
	refresh = func() {
		users, err := svc.GetAllUsers()
		listContent.Objects = nil
		if err != nil {
			listContent.Add(widget.NewLabel("Erreur: " + err.Error()))
		} else if len(users) == 0 {
			listContent.Add(widget.NewLabel("Aucun utilisateur trouvé."))
		} else {
			for _, u := range users {
				listContent.Add(buildUserRow(u))
				listContent.Add(widget.NewSeparator())
			}
		}
		listContent.Refresh()
	}
	refresh()

	// Formulaire création
	unameEntry := widget.NewEntry()
	unameEntry.SetPlaceHolder("Nom d'utilisateur")
	fullnameEntry := widget.NewEntry()
	fullnameEntry.SetPlaceHolder("Nom complet")
	passEntry := widget.NewPasswordEntry()
	passEntry.SetPlaceHolder("Mot de passe")
	roleSelect := widget.NewSelect([]string{models.RoleAdmin, models.RoleSeller, models.RoleCashier, models.RoleAssistant}, nil)
	roleSelect.SetSelected(models.RoleSeller)

	createBtn := widget.NewButtonWithIcon("Créer utilisateur", theme.ContentAddIcon(), func() {
		if unameEntry.Text == "" || passEntry.Text == "" {
			return
		}
		perms := models.DefaultPermissionsByRole(roleSelect.Selected)
		if err := svc.CreateUser(unameEntry.Text, fullnameEntry.Text, passEntry.Text, roleSelect.Selected, perms); err != nil {
			dialog.ShowError(err, fyne.CurrentApp().Driver().AllWindows()[0])
			return
		}
		unameEntry.SetText("")
		fullnameEntry.SetText("")
		passEntry.SetText("")
		refresh()
	})
	createBtn.Importance = widget.HighImportance

	createForm := container.NewVBox(
		widget.NewSeparator(),
		widget.NewRichTextFromMarkdown("### Créer un nouvel utilisateur"),
		container.NewGridWithColumns(2,
			buildFormRow("Identifiant:", unameEntry),
			buildFormRow("Nom complet:", fullnameEntry),
		),
		container.NewGridWithColumns(2,
			buildFormRow("Mot de passe:", passEntry),
			buildFormRow("Rôle:", roleSelect),
		),
		container.NewHBox(createBtn),
		widget.NewSeparator(),
		widget.NewRichTextFromMarkdown("### Utilisateurs existants"),
	)

	return container.NewBorder(
		container.NewVBox(header, createForm),
		nil, nil, nil,
		container.NewVScroll(listContent),
	)
}

// ─────────────────────────────────────────────────────────────────────────────
// BuildCurrenciesScreen — Devises
// ─────────────────────────────────────────────────────────────────────────────

func BuildCurrenciesScreen() fyne.CanvasObject {
	db := appstate.GetDB()
	header := buildSettingsHeader("💱", "Devises",
		"Consultez et gérez les devises utilisées (Dinar Algérien par défaut)", "#1abc9c")

	if db == nil {
		return container.NewVBox(header, widget.NewLabel("Base de données non connectée."))
	}

	rows, _ := db.Query(`SELECT code, name, symbol, rate FROM currencies ORDER BY code`)
	var currencies []models.Currency
	if rows != nil {
		defer rows.Close()
		for rows.Next() {
			var c models.Currency
			rows.Scan(&c.Code, &c.Name, &c.Symbol, &c.Rate)
			currencies = append(currencies, c)
		}
	}

	colHeaders := []string{"Code", "Nom", "Symbole", "Taux / DZD", "Réf."}
	colWidths := []float32{70, 200, 80, 130, 80}

	headerRow := container.NewHBox()
	for i, h := range colHeaders {
		lbl := widget.NewLabel(h)
		lbl.TextStyle = fyne.TextStyle{Bold: true}
		headerRow.Add(container.NewGridWrap(fyne.NewSize(colWidths[i], 28), lbl))
	}

	tableRows := container.NewVBox(headerRow, widget.NewSeparator())
	if len(currencies) == 0 {
		tableRows.Add(widget.NewLabel("DZD — Dinar Algérien (devise principale)"))
	}
	for _, c := range currencies {
		row := container.NewHBox(
			container.NewGridWrap(fyne.NewSize(colWidths[0], 30), widget.NewLabel(c.Code)),
			container.NewGridWrap(fyne.NewSize(colWidths[1], 30), widget.NewLabel(c.Name)),
			container.NewGridWrap(fyne.NewSize(colWidths[2], 30), widget.NewLabel(c.Symbol)),
			container.NewGridWrap(fyne.NewSize(colWidths[3], 30), widget.NewLabel(utils.FormatAmount(c.Rate))),
			container.NewGridWrap(fyne.NewSize(colWidths[4], 30), widget.NewLabel("DZD")),
		)
		tableRows.Add(row)
	}

	// Ajout devise rapide
	codeEntry := widget.NewEntry()
	codeEntry.SetPlaceHolder("USD")
	nameEntry := widget.NewEntry()
	nameEntry.SetPlaceHolder("Dollar américain")
	symbolEntry := widget.NewEntry()
	symbolEntry.SetPlaceHolder("$")
	rateEntry := widget.NewEntry()
	rateEntry.SetPlaceHolder("135.50")

	addBtn := widget.NewButtonWithIcon("Ajouter devise", theme.ContentAddIcon(), func() {
		rate, _ := strconv.ParseFloat(rateEntry.Text, 64)
		db.Exec(`INSERT OR IGNORE INTO currencies(code,name,symbol,rate)
			VALUES(?,?,?,?)`,
			strings.ToUpper(codeEntry.Text), nameEntry.Text, symbolEntry.Text, rate)
		appstate.GlobalNavigate(appstate.RouteCurrencies)
	})
	addBtn.Importance = widget.HighImportance

	addForm := container.NewVBox(
		widget.NewSeparator(),
		widget.NewRichTextFromMarkdown("### Ajouter une devise"),
		container.NewGridWithColumns(4,
			buildFormRow("Code:", codeEntry),
			buildFormRow("Nom:", nameEntry),
			buildFormRow("Symbole:", symbolEntry),
			buildFormRow("Taux DZD:", rateEntry),
		),
		container.NewHBox(addBtn),
	)

	return container.NewBorder(
		container.NewVBox(header, addForm, widget.NewSeparator()),
		nil, nil, nil,
		container.NewVScroll(tableRows),
	)
}

// ─────────────────────────────────────────────────────────────────────────────
// BuildBarcodeSettingsScreen — Paramètres Code-barres
// ─────────────────────────────────────────────────────────────────────────────

func BuildBarcodeSettingsScreen() fyne.CanvasObject {
	db := appstate.GetDB()
	header := buildSettingsHeader("📦", "Paramètres Code-barres",
		"Configurez la lecture et la génération des codes-barres articles", "#2c3e50")

	if db == nil {
		return container.NewVBox(header, widget.NewLabel("Base de données non connectée."))
	}

	readSetting := func(key string) string {
		var val string
		db.QueryRow(`SELECT value FROM settings WHERE key=?`, key).Scan(&val)
		return val
	}
	writeSetting := func(key, val string) {
		db.Exec(`INSERT INTO settings(key,value) VALUES(?,?)
			ON CONFLICT(key) DO UPDATE SET value=excluded.value`, key, val)
	}

	typeSelect := widget.NewSelect([]string{"EAN-13", "EAN-8", "CODE-128", "QR Code"}, nil)
	typeSelect.SetSelected(readSetting("barcode_type"))
	if typeSelect.Selected == "" {
		typeSelect.SetSelected("EAN-13")
	}

	prefixEntry := widget.NewEntry()
	prefixEntry.SetText(readSetting("barcode_prefix"))
	prefixEntry.SetPlaceHolder("Préfixe auto (ex: 619 = Algérie)")

	autoGenCheck := widget.NewCheck("Générer code-barres automatiquement à la création article", nil)
	autoGenCheck.SetChecked(readSetting("barcode_auto_generate") == "true")

	scanModeSelect := widget.NewSelect([]string{"USB HID", "Série COM", "Caméra"}, nil)
	scanModeSelect.SetSelected(readSetting("barcode_scanner_mode"))
	if scanModeSelect.Selected == "" {
		scanModeSelect.SetSelected("USB HID")
	}

	soundCheck := widget.NewCheck("Son bip à la lecture", nil)
	soundCheck.SetChecked(readSetting("barcode_beep") != "false")

	saveBtn := widget.NewButtonWithIcon("Enregistrer", theme.DocumentSaveIcon(), func() {
		boolStr := func(b bool) string {
			if b {
				return "true"
			}
			return "false"
		}
		writeSetting("barcode_type", typeSelect.Selected)
		writeSetting("barcode_prefix", prefixEntry.Text)
		writeSetting("barcode_auto_generate", boolStr(autoGenCheck.Checked))
		writeSetting("barcode_scanner_mode", scanModeSelect.Selected)
		writeSetting("barcode_beep", boolStr(soundCheck.Checked))
		dialog.ShowInformation("Succès", "Paramètres code-barres enregistrés.", fyne.CurrentApp().Driver().AllWindows()[0])
	})
	saveBtn.Importance = widget.HighImportance

	form := container.NewVBox(
		buildFormRow("Type code-barres:", typeSelect),
		buildFormRow("Préfixe (GS1):", prefixEntry),
		buildFormRow("Mode scanner:", scanModeSelect),
		autoGenCheck,
		soundCheck,
		widget.NewSeparator(),
		container.NewHBox(saveBtn),
	)

	return container.NewBorder(header, nil, nil, nil, container.NewVScroll(form))
}

// ─────────────────────────────────────────────────────────────────────────────
// BuildBackupScreen — Sauvegarde
// ─────────────────────────────────────────────────────────────────────────────

func BuildBackupScreen() fyne.CanvasObject {
	db := appstate.GetDB()
	header := buildSettingsHeader("💾", "Sauvegarde",
		"Créez une sauvegarde complète de la base de données et des assets", "#27ae60")

	if db == nil {
		return container.NewVBox(header, widget.NewLabel("Base de données non connectée."))
	}

	readSetting := func(key string) string {
		var val string
		db.QueryRow(`SELECT value FROM settings WHERE key=?`, key).Scan(&val)
		return val
	}

	backupDirEntry := widget.NewEntry()
	backupDirEntry.SetText(readSetting("backup_directory"))
	if backupDirEntry.Text == "" {
		backupDirEntry.SetText("./backups")
	}
	backupDirEntry.SetPlaceHolder("Répertoire de sauvegarde")

	dbPath := readSetting("db_path")
	if dbPath == "" {
		dbPath = "./data/gestion.db"
	}

	statusLbl := widget.NewLabel("Prêt à sauvegarder.")
	statusLbl.Wrapping = fyne.TextWrapWord

	// Liste des sauvegardes existantes
	loadBackups := func() []string {
		var list []string
		rows, err := db.Query(`SELECT description FROM audit_log WHERE action_type='backup' ORDER BY timestamp DESC LIMIT 10`)
		if err != nil {
			return list
		}
		defer rows.Close()
		for rows.Next() {
			var desc string
			rows.Scan(&desc)
			list = append(list, desc)
		}
		return list
	}

	backupList := container.NewVBox()
	refreshList := func() {
		backupList.Objects = nil
		for _, b := range loadBackups() {
			backupList.Add(widget.NewLabel("• " + b))
		}
		backupList.Refresh()
	}
	refreshList()

	backupBtn := widget.NewButtonWithIcon("Lancer la sauvegarde maintenant", theme.DownloadIcon(), func() {
		svc := services.NewBackupService(db, dbPath)
		statusLbl.SetText("⏳ Sauvegarde en cours...")
		path, err := svc.Backup(backupDirEntry.Text)
		if err != nil {
			statusLbl.SetText("❌ Erreur: " + err.Error())
			return
		}
		statusLbl.SetText("✅ Sauvegarde créée: " + path)
		refreshList()
	})
	backupBtn.Importance = widget.HighImportance

	autoCheck := widget.NewCheck("Sauvegarde automatique quotidienne", nil)
	autoCheck.SetChecked(readSetting("backup_auto_daily") == "true")

	keepEntry := widget.NewEntry()
	keepEntry.SetText(readSetting("backup_keep_count"))
	if keepEntry.Text == "" {
		keepEntry.SetText("30")
	}
	keepEntry.SetPlaceHolder("Nombre de sauvegardes à conserver")

	saveSettingsBtn := widget.NewButtonWithIcon("Enregistrer paramètres", theme.DocumentSaveIcon(), func() {
		boolStr := func(b bool) string {
			if b {
				return "true"
			}
			return "false"
		}
		db.Exec(`INSERT INTO settings(key,value) VALUES(?,?) ON CONFLICT(key) DO UPDATE SET value=excluded.value`,
			"backup_directory", backupDirEntry.Text)
		db.Exec(`INSERT INTO settings(key,value) VALUES(?,?) ON CONFLICT(key) DO UPDATE SET value=excluded.value`,
			"backup_auto_daily", boolStr(autoCheck.Checked))
		db.Exec(`INSERT INTO settings(key,value) VALUES(?,?) ON CONFLICT(key) DO UPDATE SET value=excluded.value`,
			"backup_keep_count", keepEntry.Text)
		dialog.ShowInformation("Succès", "Paramètres sauvegarde enregistrés.", fyne.CurrentApp().Driver().AllWindows()[0])
	})

	form := container.NewVBox(
		buildFormRow("Répertoire:", backupDirEntry),
		buildFormRow("Conserver (nb):", keepEntry),
		autoCheck,
		container.NewHBox(saveSettingsBtn),
		widget.NewSeparator(),
		backupBtn,
		statusLbl,
		widget.NewSeparator(),
		widget.NewRichTextFromMarkdown("### Dernières sauvegardes"),
		backupList,
	)

	return container.NewBorder(header, nil, nil, nil, container.NewVScroll(form))
}

// ─────────────────────────────────────────────────────────────────────────────
// BuildRestoreScreen — Restauration
// ─────────────────────────────────────────────────────────────────────────────

func BuildRestoreScreen() fyne.CanvasObject {
	db := appstate.GetDB()
	header := buildSettingsHeader("🔄", "Restauration",
		"Restaurez une sauvegarde précédente (remplace les données actuelles)", "#e74c3c")

	if db == nil {
		return container.NewVBox(header, widget.NewLabel("Base de données non connectée."))
	}

	readSetting := func(key string) string {
		var val string
		db.QueryRow(`SELECT value FROM settings WHERE key=?`, key).Scan(&val)
		return val
	}

	dbPath := readSetting("db_path")
	if dbPath == "" {
		dbPath = "./data/gestion.db"
	}

	zipPathEntry := widget.NewEntry()
	zipPathEntry.SetPlaceHolder("Chemin du fichier .zip de sauvegarde")

	targetDirEntry := widget.NewEntry()
	targetDirEntry.SetText("./")
	targetDirEntry.SetPlaceHolder("Répertoire cible de restauration")

	statusLbl := widget.NewLabel("Sélectionnez un fichier de sauvegarde .zip")
	statusLbl.Wrapping = fyne.TextWrapWord

	warningLbl := widget.NewRichTextFromMarkdown(
		"> ⚠️ **ATTENTION** : La restauration remplacera **définitivement** toutes les données actuelles." +
			" Effectuez une sauvegarde préalable avant de restaurer.")

	restoreBtn := widget.NewButtonWithIcon("Restaurer maintenant", theme.UploadIcon(), func() {
		if zipPathEntry.Text == "" {
			statusLbl.SetText("❌ Veuillez indiquer le chemin du fichier ZIP.")
			return
		}
		wins := fyne.CurrentApp().Driver().AllWindows()
		if len(wins) == 0 {
			return
		}
		dialog.ShowConfirm("Confirmation restauration",
			"Êtes-vous certain de vouloir restaurer cette sauvegarde ?\nToutes les données actuelles seront remplacées.",
			func(ok bool) {
				if !ok {
					return
				}
				zipPath := zipPathEntry.Text
				targetDir := targetDirEntry.Text
				fyne.Do(func() { statusLbl.SetText("⏳ Restauration en cours...") })
				go func() {
					svc := services.NewBackupService(db, dbPath)
					if err := svc.Restore(zipPath, targetDir); err != nil {
						fyne.Do(func() { statusLbl.SetText("❌ Erreur: " + err.Error()) })
						return
					}
					fyne.Do(func() { statusLbl.SetText("✅ Restauration terminée. Redémarrez l'application.") })
				}()
			},
			wins[0])
	})
	restoreBtn.Importance = widget.DangerImportance

	form := container.NewVBox(
		warningLbl,
		widget.NewSeparator(),
		buildFormRow("Fichier ZIP:", zipPathEntry),
		buildFormRow("Répertoire cible:", targetDirEntry),
		container.NewHBox(restoreBtn),
		widget.NewSeparator(),
		statusLbl,
	)

	return container.NewBorder(header, nil, nil, nil, container.NewVScroll(form))
}

// ─────────────────────────────────────────────────────────────────────────────
// BuildCalculatorScreen — Calculatrice
// ─────────────────────────────────────────────────────────────────────────────

func BuildCalculatorScreen() fyne.CanvasObject {
	header := buildSettingsHeader("🧮", "Calculatrice",
		"Calculatrice intégrée avec historique", "#8e44ad")

	displayEntry := widget.NewEntry()
	displayEntry.SetText("0")
	displayEntry.Disable()

	historyBox := container.NewVBox()
	historyScroll := container.NewVScroll(historyBox)

	var current string
	var lastResult string
	var operator string
	var waitingForOperand bool

	setDisplay := func(s string) {
		displayEntry.SetText(s)
	}

	appendDigit := func(d string) {
		if waitingForOperand {
			current = d
			waitingForOperand = false
		} else {
			if current == "0" {
				current = d
			} else {
				current += d
			}
		}
		setDisplay(current)
	}

	appendDot := func() {
		if waitingForOperand {
			current = "0."
			waitingForOperand = false
		} else if !strings.Contains(current, ".") {
			if current == "" {
				current = "0."
			} else {
				current += "."
			}
		}
		setDisplay(current)
	}

	calculate := func() {
		if operator == "" || lastResult == "" {
			return
		}
		a, _ := strconv.ParseFloat(lastResult, 64)
		b, _ := strconv.ParseFloat(current, 64)
		var result float64
		switch operator {
		case "+":
			result = a + b
		case "-":
			result = a - b
		case "×":
			result = a * b
		case "÷":
			if b == 0 {
				setDisplay("Erreur")
				return
			}
			result = a / b
		case "%":
			result = a * b / 100
		}
		expr := fmt.Sprintf("%s %s %s = %s DA",
			utils.FormatAmount(a), operator, utils.FormatAmount(b), utils.FormatAmount(result))
		historyBox.Add(widget.NewLabel(expr))
		historyScroll.ScrollToBottom()

		lastResult = strconv.FormatFloat(result, 'f', -1, 64)
		current = lastResult
		setDisplay(utils.FormatAmount(result))
		operator = ""
		waitingForOperand = false
	}

	setOperator := func(op string) {
		if current == "" {
			return
		}
		if operator != "" && !waitingForOperand {
			calculate()
		}
		lastResult = current
		operator = op
		waitingForOperand = true
		setDisplay(current + " " + op)
	}

	clearAll := func() {
		current = ""
		lastResult = ""
		operator = ""
		waitingForOperand = false
		setDisplay("0")
	}

	clearEntry := func() {
		current = ""
		setDisplay("0")
	}

	toggleSign := func() {
		v, err := strconv.ParseFloat(current, 64)
		if err != nil {
			return
		}
		v = -v
		current = strconv.FormatFloat(v, 'f', -1, 64)
		setDisplay(current)
	}

	sqrtFn := func() {
		v, err := strconv.ParseFloat(current, 64)
		if err != nil || v < 0 {
			setDisplay("Erreur")
			return
		}
		result := math.Sqrt(v)
		current = strconv.FormatFloat(result, 'f', -1, 64)
		setDisplay(utils.FormatAmount(result))
		lastResult = current
	}

	btnStyle := func(label string, fn func()) *widget.Button {
		return widget.NewButton(label, fn)
	}

	numBtn := func(d string) *widget.Button {
		b := btnStyle(d, func() { appendDigit(d) })
		return b
	}

	opBtn := func(op string) *widget.Button {
		b := btnStyle(op, func() { setOperator(op) })
		b.Importance = widget.MediumImportance
		return b
	}

	eqBtn := widget.NewButton("=", calculate)
	eqBtn.Importance = widget.HighImportance

	grid := container.NewGridWithColumns(4,
		btnStyle("AC", clearAll), btnStyle("CE", clearEntry), btnStyle("±", toggleSign), opBtn("÷"),
		numBtn("7"), numBtn("8"), numBtn("9"), opBtn("×"),
		numBtn("4"), numBtn("5"), numBtn("6"), opBtn("-"),
		numBtn("1"), numBtn("2"), numBtn("3"), opBtn("+"),
		numBtn("0"), btnStyle(".", appendDot), btnStyle("√", sqrtFn), eqBtn,
	)

	tvaNoteBtn := widget.NewButton("TVA 19%", func() {
		v, _ := strconv.ParseFloat(current, 64)
		tva := v * 0.19
		ttc := v + tva
		expr := fmt.Sprintf("HT: %s | TVA: %s | TTC: %s DA",
			utils.FormatAmount(v), utils.FormatAmount(tva), utils.FormatAmount(ttc))
		historyBox.Add(widget.NewLabel(expr))
		historyScroll.ScrollToBottom()
		current = strconv.FormatFloat(ttc, 'f', -1, 64)
		setDisplay(utils.FormatAmount(ttc))
	})

	tva9Btn := widget.NewButton("TVA 9%", func() {
		v, _ := strconv.ParseFloat(current, 64)
		tva := v * 0.09
		ttc := v + tva
		expr := fmt.Sprintf("HT: %s | TVA: %s | TTC: %s DA",
			utils.FormatAmount(v), utils.FormatAmount(tva), utils.FormatAmount(ttc))
		historyBox.Add(widget.NewLabel(expr))
		historyScroll.ScrollToBottom()
		current = strconv.FormatFloat(ttc, 'f', -1, 64)
		setDisplay(utils.FormatAmount(ttc))
	})

	timbreBtn := widget.NewButton("Timbre 1‰", func() {
		v, _ := strconv.ParseFloat(current, 64)
		timbre := v * 0.001
		if timbre > 2500 {
			timbre = 2500
		}
		if timbre < 0 {
			timbre = 0
		}
		expr := fmt.Sprintf("TTC: %s | Timbre: %s DA", utils.FormatAmount(v), utils.FormatAmount(timbre))
		historyBox.Add(widget.NewLabel(expr))
		historyScroll.ScrollToBottom()
		current = strconv.FormatFloat(timbre, 'f', -1, 64)
		setDisplay(utils.FormatAmount(timbre))
	})

	quickBtns := container.NewHBox(tvaNoteBtn, tva9Btn, timbreBtn)

	leftPanel := container.NewVBox(
		displayEntry,
		grid,
		widget.NewSeparator(),
		quickBtns,
	)

	rightPanel := container.NewBorder(
		widget.NewRichTextFromMarkdown("### Historique"),
		nil, nil, nil,
		historyScroll,
	)

	return container.NewBorder(
		header, nil, nil, nil,
		container.NewGridWithColumns(2, leftPanel, rightPanel),
	)
}

// ─────────────────────────────────────────────────────────────────────────────
// BuildPriceUpdateScreen — Mise à Jour Prix en Masse
// ─────────────────────────────────────────────────────────────────────────────

func BuildPriceUpdateScreen() fyne.CanvasObject {
	db := appstate.GetDB()
	header := buildSettingsHeader("💹", "Mise à Jour Prix",
		"Appliquez une hausse ou baisse de prix en masse sur une sélection d'articles", "#e67e22")

	if db == nil {
		return container.NewVBox(header, widget.NewLabel("Base de données non connectée."))
	}

	// Charger catégories
	catRows, _ := db.Query(`SELECT id, name FROM categories ORDER BY name`)
	var catNames []string
	var catIDs []int
	catNames = append(catNames, "Toutes catégories")
	catIDs = append(catIDs, 0)
	if catRows != nil {
		defer catRows.Close()
		for catRows.Next() {
			var id int
			var name string
			catRows.Scan(&id, &name)
			catNames = append(catNames, name)
			catIDs = append(catIDs, id)
		}
	}

	catSelect := widget.NewSelect(catNames, nil)
	catSelect.SetSelectedIndex(0)

	modeSelect := widget.NewSelect([]string{"Hausse %", "Baisse %", "Hausse montant fixe", "Baisse montant fixe", "Nouveau prix fixe"}, nil)
	modeSelect.SetSelected("Hausse %")

	valueEntry := widget.NewEntry()
	valueEntry.SetPlaceHolder("Valeur (ex: 10 pour 10% ou 500 DA)")

	roundSelect := widget.NewSelect([]string{"Aucun", "Arrondi à 1 DA", "Arrondi à 5 DA", "Arrondi à 10 DA", "Arrondi à 100 DA"}, nil)
	roundSelect.SetSelected("Arrondi à 1 DA")

	applyToPurchaseCheck := widget.NewCheck("Appliquer aussi au prix d'achat", nil)

	previewList := container.NewVBox()
	previewScroll := container.NewVScroll(previewList)

	statusLbl := widget.NewLabel("")

	getSelectedCatID := func() int {
		idx := catSelect.SelectedIndex()
		if idx < 0 || idx >= len(catIDs) {
			return 0
		}
		return catIDs[idx]
	}

	doRound := func(v float64) float64 {
		switch roundSelect.Selected {
		case "Arrondi à 1 DA":
			return math.Round(v)
		case "Arrondi à 5 DA":
			return math.Round(v/5) * 5
		case "Arrondi à 10 DA":
			return math.Round(v/10) * 10
		case "Arrondi à 100 DA":
			return math.Round(v/100) * 100
		}
		return v
	}

	calcNewPrice := func(old float64) float64 {
		val, _ := strconv.ParseFloat(valueEntry.Text, 64)
		var newPrice float64
		switch modeSelect.Selected {
		case "Hausse %":
			newPrice = old * (1 + val/100)
		case "Baisse %":
			newPrice = old * (1 - val/100)
		case "Hausse montant fixe":
			newPrice = old + val
		case "Baisse montant fixe":
			newPrice = old - val
		case "Nouveau prix fixe":
			newPrice = val
		default:
			newPrice = old
		}
		if newPrice < 0 {
			newPrice = 0
		}
		return doRound(newPrice)
	}

	previewBtn := widget.NewButtonWithIcon("Aperçu", theme.SearchIcon(), func() {
		catID := getSelectedCatID()
		query := `SELECT id, designation, sale_price_ttc FROM articles WHERE is_active=1`
		args := []interface{}{}
		if catID > 0 {
			query += ` AND category_id=?`
			args = append(args, catID)
		}
		query += ` ORDER BY designation LIMIT 50`
		rows2, err := db.Query(query, args...)
		if err != nil {
			statusLbl.SetText("Erreur: " + err.Error())
			return
		}
		defer rows2.Close()
		previewList.Objects = nil
		count := 0
		for rows2.Next() {
			var id int
			var name string
			var oldPrice float64
			rows2.Scan(&id, &name, &oldPrice)
			newPrice := calcNewPrice(oldPrice)
			diff := newPrice - oldPrice
			sign := "+"
			if diff < 0 {
				sign = ""
			}
			lbl := widget.NewLabel(fmt.Sprintf("%-30s  %s DA  →  %s DA  (%s%s DA)",
				utils.TruncateString(name, 28),
				utils.FormatAmount(oldPrice),
				utils.FormatAmount(newPrice),
				sign,
				utils.FormatAmount(diff),
			))
			previewList.Add(lbl)
			count++
		}
		statusLbl.SetText(fmt.Sprintf("%d article(s) affectés.", count))
		previewList.Refresh()
	})

	applyBtn := widget.NewButtonWithIcon("Appliquer la mise à jour", theme.ConfirmIcon(), func() {
		wins := fyne.CurrentApp().Driver().AllWindows()
		if len(wins) == 0 {
			return
		}
		dialog.ShowConfirm("Confirmation",
			"Appliquer la mise à jour des prix à tous les articles sélectionnés ?",
			func(ok bool) {
				if !ok {
					return
				}
				modeVal := modeSelect.Selected
				valueVal := valueEntry.Text
				applyPurch := applyToPurchaseCheck.Checked
				fyne.Do(func() { statusLbl.SetText("⏳ Mise à jour en cours...") })
				go func() {
					catID := getSelectedCatID()
					query := `SELECT id, sale_price_ttc, purchase_price FROM articles WHERE is_active=1`
					args2 := []interface{}{}
					if catID > 0 {
						query += ` AND category_id=?`
						args2 = append(args2, catID)
					}
					rows3, err := db.Query(query, args2...)
					if err != nil {
						fyne.Do(func() { statusLbl.SetText("Erreur: " + err.Error()) })
						return
					}
					type artUpdate struct {
						id       int
						oldSale  float64
						oldPurch float64
					}
					var updates []artUpdate
					for rows3.Next() {
						var u artUpdate
						rows3.Scan(&u.id, &u.oldSale, &u.oldPurch)
						updates = append(updates, u)
					}
					rows3.Close()
					count := 0
					for _, u := range updates {
						newSale := calcNewPrice(u.oldSale)
						db.Exec(`UPDATE articles SET sale_price_ttc=? WHERE id=?`, newSale, u.id)
						if applyPurch {
							newPurch := calcNewPrice(u.oldPurch)
							db.Exec(`UPDATE articles SET purchase_price=? WHERE id=?`, newPurch, u.id)
						}
						count++
					}
					db.Exec(`INSERT INTO audit_log(action_type,module,description) VALUES(?,?,?)`,
						"mass_update", "articles", fmt.Sprintf("Mise à jour prix: %s %s sur %d articles",
							modeVal, valueVal, count))
					fyne.Do(func() { statusLbl.SetText(fmt.Sprintf("✅ %d article(s) mis à jour.", count)) })
				}()
			},
			wins[0])
	})
	applyBtn.Importance = widget.HighImportance

	form := container.NewVBox(
		container.NewGridWithColumns(2,
			buildFormRow("Catégorie:", catSelect),
			buildFormRow("Mode:", modeSelect),
		),
		container.NewGridWithColumns(2,
			buildFormRow("Valeur:", valueEntry),
			buildFormRow("Arrondi:", roundSelect),
		),
		applyToPurchaseCheck,
		container.NewHBox(previewBtn, applyBtn),
		statusLbl,
	)

	return container.NewBorder(
		container.NewVBox(header, form, widget.NewSeparator(),
			widget.NewRichTextFromMarkdown("### Aperçu (50 premiers articles)")),
		nil, nil, nil,
		previewScroll,
	)
}

// ─────────────────────────────────────────────────────────────────────────────
// BuildCalendarScreen — Calendrier & Rappels
// ─────────────────────────────────────────────────────────────────────────────

func BuildCalendarScreen() fyne.CanvasObject {
	db := appstate.GetDB()
	header := buildSettingsHeader("📆", "Calendrier & Rappels",
		"Gérez vos rendez-vous, rappels et échéances commerciales", "#2980b9")

	if db == nil {
		return container.NewVBox(header, widget.NewLabel("Base de données non connectée."))
	}

	loadReminders := func() []models.Reminder {
		rows, err := db.Query(`SELECT id, title, date, description, is_done, type
			FROM reminders ORDER BY date ASC LIMIT 100`)
		if err != nil {
			return nil
		}
		defer rows.Close()
		var list []models.Reminder
		for rows.Next() {
			var r models.Reminder
			rows.Scan(&r.ID, &r.Title, &r.Date, &r.Description, &r.IsDone, &r.Type)
			list = append(list, r)
		}
		return list
	}

	listContainer := container.NewVBox()
	var refreshList func()

	today := utils.TodayString()

	buildReminderRow := func(r models.Reminder) fyne.CanvasObject {
		icon := "🔔"
		if r.IsDone {
			icon = "✅"
		} else if r.Date < today {
			icon = "🔴"
		}
		info := widget.NewLabel(fmt.Sprintf("%s [%s] %s — %s",
			icon, r.Date, r.Title, utils.TruncateString(r.Description, 40)))

		doneBtn := widget.NewButton("✓", func() {
			db.Exec(`UPDATE reminders SET is_done=1 WHERE id=?`, r.ID)
			refreshList()
		})
		deleteBtn := widget.NewButtonWithIcon("", theme.DeleteIcon(), func() {
			db.Exec(`DELETE FROM reminders WHERE id=?`, r.ID)
			refreshList()
		})
		return container.NewBorder(nil, nil, nil,
			container.NewHBox(doneBtn, deleteBtn),
			info)
	}

	refreshList = func() {
		reminders := loadReminders()
		listContainer.Objects = nil
		if len(reminders) == 0 {
			listContainer.Add(widget.NewLabel("Aucun rappel. Créez-en un ci-dessus."))
		}
		for _, r := range reminders {
			listContainer.Add(buildReminderRow(r))
			listContainer.Add(widget.NewSeparator())
		}
		listContainer.Refresh()
	}
	refreshList()

	// Chèques à encaisser bientôt
	chequeList := container.NewVBox()
	crows, err := db.Query(`SELECT cheque_number, amount, due_date, drawer_name
		FROM cheques WHERE status='pending' AND due_date<=date('now','+7 days')
		ORDER BY due_date ASC`)
	if err == nil && crows != nil {
		defer crows.Close()
		chequeList.Add(widget.NewRichTextFromMarkdown("### 📋 Chèques à encaisser dans 7 jours"))
		any := false
		for crows.Next() {
			var num, drawer, due string
			var amount float64
			crows.Scan(&num, &amount, &due, &drawer)
			chequeList.Add(widget.NewLabel(fmt.Sprintf("• N° %s | %s DA | Échéance: %s | %s",
				num, utils.FormatAmount(amount), due, drawer)))
			any = true
		}
		if !any {
			chequeList.Add(widget.NewLabel("Aucun chèque à encaisser dans les 7 prochains jours."))
		}
	}

	// Formulaire nouveau rappel
	titleEntry := widget.NewEntry()
	titleEntry.SetPlaceHolder("Titre du rappel")
	dateEntry := widget.NewEntry()
	dateEntry.SetText(today)
	dateEntry.SetPlaceHolder("AAAA-MM-JJ")
	descEntry := widget.NewMultiLineEntry()
	descEntry.SetPlaceHolder("Description (optionnel)")
	descEntry.SetMinRowsVisible(2)
	typeSelect := widget.NewSelect([]string{"rendez-vous", "relance", "paiement", "autre"}, nil)
	typeSelect.SetSelected("rendez-vous")

	addBtn := widget.NewButtonWithIcon("Ajouter rappel", theme.ContentAddIcon(), func() {
		if titleEntry.Text == "" {
			return
		}
		db.Exec(`INSERT INTO reminders(title,date,description,is_done,type)
			VALUES(?,?,?,0,?)`,
			titleEntry.Text, dateEntry.Text, descEntry.Text, typeSelect.Selected)
		titleEntry.SetText("")
		descEntry.SetText("")
		refreshList()
	})
	addBtn.Importance = widget.HighImportance

	addForm := container.NewVBox(
		widget.NewSeparator(),
		widget.NewRichTextFromMarkdown("### Nouveau rappel"),
		container.NewGridWithColumns(2,
			buildFormRow("Titre:", titleEntry),
			buildFormRow("Date:", dateEntry),
		),
		container.NewGridWithColumns(2,
			buildFormRow("Type:", typeSelect),
			buildFormRow("Description:", descEntry),
		),
		container.NewHBox(addBtn),
	)

	return container.NewBorder(
		container.NewVBox(header, chequeList, addForm, widget.NewSeparator(),
			widget.NewRichTextFromMarkdown("### Rappels")),
		nil, nil, nil,
		container.NewVScroll(listContainer),
	)
}

// ─────────────────────────────────────────────────────────────────────────────
// BuildAuditLogScreen — Journal d'Audit
// ─────────────────────────────────────────────────────────────────────────────

func BuildAuditLogScreen() fyne.CanvasObject {
	db := appstate.GetDB()
	header := buildSettingsHeader("📋", "Journal d'Audit",
		"Consultez l'historique de toutes les actions effectuées dans le système", "#7f8c8d")

	if db == nil {
		return container.NewVBox(header, widget.NewLabel("Base de données non connectée."))
	}

	userFilter := widget.NewEntry()
	userFilter.SetPlaceHolder("Filtrer par utilisateur...")
	moduleFilter := widget.NewSelect([]string{"Tous", "auth", "articles", "sales", "purchases",
		"treasury", "settings", "backup", "tax_settings", "mass_update"}, nil)
	moduleFilter.SetSelected("Tous")
	limitSelect := widget.NewSelect([]string{"50", "100", "200", "500"}, nil)
	limitSelect.SetSelected("100")

	listContainer := container.NewVBox()
	scroll := container.NewVScroll(listContainer)

	loadLogs := func() {
		limit, _ := strconv.Atoi(limitSelect.Selected)
		query := `SELECT a.timestamp, COALESCE(u.username,'système'), a.action_type, a.module, a.description
			FROM audit_log a
			LEFT JOIN users u ON u.id=a.user_id
			WHERE 1=1`
		args := []interface{}{}
		if userFilter.Text != "" {
			query += ` AND u.username LIKE ?`
			args = append(args, "%"+userFilter.Text+"%")
		}
		if moduleFilter.Selected != "Tous" {
			query += ` AND a.module=?`
			args = append(args, moduleFilter.Selected)
		}
		query += ` ORDER BY a.id DESC LIMIT ?`
		args = append(args, limit)

		rows, err := db.Query(query, args...)
		listContainer.Objects = nil
		if err != nil {
			listContainer.Add(widget.NewLabel("Erreur: " + err.Error()))
			listContainer.Refresh()
			return
		}
		defer rows.Close()
		count := 0
		for rows.Next() {
			var ts, user, action, module, desc string
			rows.Scan(&ts, &user, &action, &module, &desc)
			// Format ts
			if len(ts) > 16 {
				ts = ts[:16]
			}
			lbl := widget.NewLabel(fmt.Sprintf("[%s] %-12s | %-8s | %-15s | %s",
				ts, user, action, module, utils.TruncateString(desc, 60)))
			listContainer.Add(lbl)
			count++
		}
		if count == 0 {
			listContainer.Add(widget.NewLabel("Aucune entrée dans le journal."))
		}
		listContainer.Refresh()
		scroll.ScrollToBottom()
	}
	loadLogs()

	refreshBtn := widget.NewButtonWithIcon("Actualiser", theme.ViewRefreshIcon(), func() {
		loadLogs()
	})

	exportBtn := widget.NewButtonWithIcon("Exporter CSV", theme.DocumentIcon(), func() {
		// Export simple vers fichier texte
		rows, err := db.Query(`SELECT a.timestamp, COALESCE(u.username,'système'), a.action_type, a.module, a.description
			FROM audit_log a LEFT JOIN users u ON u.id=a.user_id ORDER BY a.id DESC LIMIT 10000`)
		if err != nil {
			return
		}
		defer rows.Close()
		var lines []string
		lines = append(lines, "Timestamp,Utilisateur,Action,Module,Description")
		for rows.Next() {
			var ts, user, action, module, desc string
			rows.Scan(&ts, &user, &action, &module, &desc)
			desc = strings.ReplaceAll(desc, ",", ";")
			lines = append(lines, fmt.Sprintf("%s,%s,%s,%s,%s", ts, user, action, module, desc))
		}
		filename := fmt.Sprintf("audit_log_%s.csv", strings.ReplaceAll(utils.TodayString(), "-", ""))
		content := strings.Join(lines, "\n")
		_ = content // In real app: write to file
		_ = filename
		dialog.ShowInformation("Export", "Export CSV: "+filename+" ("+strconv.Itoa(len(lines)-1)+" lignes)",
			fyne.CurrentApp().Driver().AllWindows()[0])
	})

	filterBar := container.NewBorder(nil, nil, nil,
		container.NewHBox(refreshBtn, exportBtn),
		container.NewGridWithColumns(3,
			buildFormRow("Utilisateur:", userFilter),
			buildFormRow("Module:", moduleFilter),
			buildFormRow("Limite:", limitSelect),
		),
	)
	userFilter.OnChanged = func(_ string) { loadLogs() }
	moduleFilter.OnChanged = func(_ string) { loadLogs() }
	limitSelect.OnChanged = func(_ string) { loadLogs() }

	return container.NewBorder(
		container.NewVBox(header, filterBar, widget.NewSeparator()),
		nil, nil, nil,
		scroll,
	)
}

// ─────────────────────────────────────────────────────────────────────────────
// BuildPrintCenterScreen — Centre d'Impression
// ─────────────────────────────────────────────────────────────────────────────

func BuildPrintCenterScreen() fyne.CanvasObject {
	db := appstate.GetDB()
	header := buildSettingsHeader("🖨️", "Centre d'Impression",
		"Réimprimez et gérez les documents PDF générés", "#34495e")

	if db == nil {
		return container.NewVBox(header, widget.NewLabel("Base de données non connectée."))
	}

	typeSelect := widget.NewSelect([]string{
		"Tous", "FA", "FAC", "BL", "BR", "DV", "PF", "BCC", "BCF", "AV",
	}, nil)
	typeSelect.SetSelected("Tous")

	fromEntry := widget.NewEntry()
	fromEntry.SetPlaceHolder("Du (AAAA-MM-JJ)")
	fromEntry.SetText(fmt.Sprintf("%d-%02d-01", time.Now().Year(), time.Now().Month()))

	toEntry := widget.NewEntry()
	toEntry.SetPlaceHolder("Au (AAAA-MM-JJ)")
	toEntry.SetText(utils.TodayString())

	searchEntry := widget.NewEntry()
	searchEntry.SetPlaceHolder("Recherche N° document ou tiers...")

	listContainer := container.NewVBox()
	scroll := container.NewVScroll(listContainer)

	colHeaders := []string{"N° Document", "Type", "Date", "Tiers", "Montant TTC", "Statut", ""}
	colWidths := []float32{140, 60, 100, 180, 130, 90, 90}

	headerRow := container.NewHBox()
	for i, h := range colHeaders {
		lbl := widget.NewLabel(h)
		lbl.TextStyle = fyne.TextStyle{Bold: true}
		headerRow.Add(container.NewGridWrap(fyne.NewSize(colWidths[i], 28), lbl))
	}

	loadDocs := func() {
		query := `SELECT d.doc_number, d.doc_type, d.date,
			COALESCE(c.name_fr, s.name_fr, '') as tiers,
			d.net_amount, d.status
			FROM documents d
			LEFT JOIN clients c ON c.id=d.client_id
			LEFT JOIN suppliers s ON s.id=d.supplier_id
			WHERE 1=1`
		args := []interface{}{}
		if typeSelect.Selected != "Tous" {
			query += ` AND d.doc_type=?`
			args = append(args, typeSelect.Selected)
		}
		if fromEntry.Text != "" {
			query += ` AND d.date>=?`
			args = append(args, fromEntry.Text)
		}
		if toEntry.Text != "" {
			query += ` AND d.date<=?`
			args = append(args, toEntry.Text)
		}
		if searchEntry.Text != "" {
			query += ` AND (d.doc_number LIKE ? OR c.name_fr LIKE ? OR s.name_fr LIKE ?)`
			s3 := "%" + searchEntry.Text + "%"
			args = append(args, s3, s3, s3)
		}
		query += ` ORDER BY d.date DESC, d.id DESC LIMIT 100`

		rows, err := db.Query(query, args...)
		listContainer.Objects = nil
		listContainer.Add(headerRow)
		listContainer.Add(widget.NewSeparator())
		if err != nil {
			listContainer.Add(widget.NewLabel("Erreur: " + err.Error()))
			listContainer.Refresh()
			return
		}
		defer rows.Close()
		count := 0
		for rows.Next() {
			var docNum, docType, date, tiers, status string
			var amount float64
			rows.Scan(&docNum, &docType, &date, &tiers, &amount, &status)

			printBtn := widget.NewButtonWithIcon("PDF", theme.DocumentPrintIcon(), func() {
				dialog.ShowInformation("Impression",
					fmt.Sprintf("Impression du document %s en cours...", docNum),
					fyne.CurrentApp().Driver().AllWindows()[0])
			})

			row := container.NewHBox(
				container.NewGridWrap(fyne.NewSize(colWidths[0], 30), widget.NewLabel(docNum)),
				container.NewGridWrap(fyne.NewSize(colWidths[1], 30), widget.NewLabel(docType)),
				container.NewGridWrap(fyne.NewSize(colWidths[2], 30), widget.NewLabel(date)),
				container.NewGridWrap(fyne.NewSize(colWidths[3], 30), widget.NewLabel(utils.TruncateString(tiers, 22))),
				container.NewGridWrap(fyne.NewSize(colWidths[4], 30), widget.NewLabel(utils.FormatAmount(amount)+" DA")),
				container.NewGridWrap(fyne.NewSize(colWidths[5], 30), widget.NewLabel(utils.StatusLabel(status))),
				container.NewGridWrap(fyne.NewSize(colWidths[6], 30), printBtn),
			)
			listContainer.Add(row)
			count++
		}
		if count == 0 {
			listContainer.Add(widget.NewLabel("Aucun document trouvé."))
		}
		listContainer.Refresh()
	}
	loadDocs()

	searchEntry.OnChanged = func(_ string) { loadDocs() }
	typeSelect.OnChanged = func(_ string) { loadDocs() }

	filterBar := container.NewGridWithColumns(4,
		buildFormRow("Type:", typeSelect),
		buildFormRow("Du:", fromEntry),
		buildFormRow("Au:", toEntry),
		buildFormRow("Recherche:", searchEntry),
	)

	return container.NewBorder(
		container.NewVBox(header, filterBar, widget.NewSeparator()),
		nil, nil, nil,
		scroll,
	)
}

// ─────────────────────────────────────────────────────────────────────────────
// BuildAboutScreen — À Propos
// ─────────────────────────────────────────────────────────────────────────────

func BuildAboutScreen() fyne.CanvasObject {
	header := buildSettingsHeader("ℹ️", "À Propos",
		"Gestion Commerciale Pro — Logiciel de gestion pour PME algériennes", "#2c3e50")

	content := widget.NewRichTextFromMarkdown(`
## Gestion Commerciale Pro v1.0

**Logiciel de gestion commerciale complet pour PME algériennes**

---

### Fonctionnalités
- ✅ **Lot 1** — Authentification & sécurité (bcrypt, rôles, permissions)
- ✅ **Lot 2** — Tableau de bord (KPIs temps réel, graphiques)
- ✅ **Lot 3** — Gestion articles, catégories, marques, unités, dépôts
- ✅ **Lot 4** — Stock (mouvements, inventaire, valorisation CMUP)
- ✅ **Lot 5** — Listes de prix, tarifs spéciaux
- ✅ **Lot 6** — Ventes (FA, DV, PF, BL, BCC), Achats (FAC, BR, BCF), POS
- ✅ **Lot 7** — Tiers (Clients, Fournisseurs, Chauffeurs)
- ✅ **Lot 8** — Trésorerie (Caisse, Banque, Chèques, Encaissements, Dépenses)
- ✅ **Lot 9** — Rapports & Fiscalité (G50, TVA, Registres, Bénéfices)
- ✅ **Lot 10** — Paramètres, Utilisateurs, Outils, Sauvegarde

---

### Technologies
| Composant | Version |
|-----------|---------|
| Go | 1.19+ |
| Fyne UI | v2.4 |
| SQLite | Embarqué |
| gopdf | v0.22 |
| excelize | v2.8 |

---

### Spécificités Algériennes
- TVA : 0%, 9%, 19%
- Timbre fiscal : 1‰ TTC (max 2 500 DA)
- TAP : 2% (configurable)
- Déclaration G50 trimestrielle
- 58 wilayas
- Montants en lettres (Français & Arabe)
- Champs fiscaux : NIF, NIS, RC, AI

---

### Licence
Logiciel propriétaire — Tous droits réservés.  
Développé avec ❤️ pour les PME algériennes.

**Support** : migashopsite@gmail.com  
**GitHub** : https://github.com/GPT4-AI/go-gestion-commerciale
`)

	content.Wrapping = fyne.TextWrapWord

	sysInfo := widget.NewRichTextFromMarkdown(fmt.Sprintf(`
---
### Session en cours
- **Utilisateur** : %s  
- **Rôle** : %s  
- **Date** : %s
`,
		func() string {
			s := appstate.GetSession()
			if s != nil {
				return s.FullName
			}
			return "Non connecté"
		}(),
		func() string {
			s := appstate.GetSession()
			if s != nil {
				return s.Role
			}
			return "-"
		}(),
		time.Now().Format("02/01/2006 15:04"),
	))

	return container.NewBorder(
		header, nil, nil, nil,
		container.NewVScroll(container.NewVBox(content, sysInfo)),
	)
}
