package screens

import (
	"database/sql"
	"fmt"
	"gestion-commerciale/internal/app"
	"gestion-commerciale/internal/database/queries"
	"gestion-commerciale/internal/models"
	"gestion-commerciale/internal/services"
	"gestion-commerciale/pkg/utils"
	"strconv"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// SalesUILine représente une ligne de document dans l'interface vente/achat
type SalesUILine struct {
	Article     *models.Article
	Designation string
	Qty         float64
	Unit        string
	UnitPriceHT float64
	DiscPct     float64
	AmountHT    float64
	TVARate     float64
	TVAAmount   float64
	AmountTTC   float64
	LotNumber   string
}

// ─────────────────────────────────────────────────────────────────────────────
// LISTE DES FACTURES DE VENTE
// ─────────────────────────────────────────────────────────────────────────────

// BuildSaleInvoicesListScreen construit l'écran liste des factures de vente
func BuildSaleInvoicesListScreen() fyne.CanvasObject {
	return buildDocumentListScreen(models.DocTypeFA, "🧾 Factures de Vente", "#27ae60", func(docID int) {
		app.Navigate(app.RouteSaleInvoice, docID)
	})
}

// buildDocumentListScreen construit une liste générique de documents
func buildDocumentListScreen(docType, title, color string, onOpen func(id int)) fyne.CanvasObject {
	db := app.GetDB()

	searchEntry := widget.NewEntry()
	searchEntry.SetPlaceHolder("🔍 N° document, client...")

	statusSelect := widget.NewSelect([]string{
		"Tous", "Brouillon", "Confirmé", "Payé", "Partiel", "Annulé",
	}, nil)
	statusSelect.SetSelected("Tous")

	statusDBMap := map[string]string{
		"Brouillon": "draft", "Confirmé": "confirmed",
		"Payé": "paid", "Partiel": "partial", "Annulé": "cancelled",
	}

	var docs []models.Document
	selectedRow := -1

	loadDocs := func(search, status string) {
		if db == nil {
			return
		}
		filters := map[string]string{"limit": "200"}
		if search != "" {
			filters["search"] = search
		}
		if s, ok := statusDBMap[status]; ok {
			filters["status"] = s
		}
		var err error
		docs, err = queries.GetDocumentsList(db, docType, filters)
		if err != nil {
			docs = nil
		}
	}
	loadDocs("", "Tous")

	// Colonnes selon le type
	isSupplierDoc := docType == models.DocTypeFAC || docType == "BR" || docType == "BCC"
	partyHeader := "Client"
	if isSupplierDoc {
		partyHeader = "Fournisseur"
	}

	headers := []string{"N° Document", "Date", partyHeader, "HT", "TVA", "Net à Payer", "Statut", "Mode Paiement"}
	colWidths := []float32{130, 90, 180, 100, 80, 110, 90, 110}

	table := widget.NewTable(
		func() (int, int) { return len(docs) + 1, len(headers) },
		func() fyne.CanvasObject {
			lbl := widget.NewLabel("")
			lbl.Truncation = fyne.TextTruncateEllipsis
			return lbl
		},
		func(id widget.TableCellID, cell fyne.CanvasObject) {
			lbl := cell.(*widget.Label)
			if id.Row == 0 {
				lbl.SetText(headers[id.Col])
				lbl.TextStyle = fyne.TextStyle{Bold: true}
				return
			}
			if id.Row-1 >= len(docs) {
				lbl.SetText("")
				return
			}
			d := docs[id.Row-1]
			switch id.Col {
			case 0:
				lbl.SetText(d.DocNumber)
			case 1:
				lbl.SetText(utils.FormatDateFr(d.Date))
			case 2:
				if isSupplierDoc {
					lbl.SetText(d.SupplierName)
				} else {
					lbl.SetText(d.ClientName)
				}
			case 3:
				lbl.SetText(utils.FormatMoney(d.TotalHT))
			case 4:
				lbl.SetText(utils.FormatMoney(d.TotalTVA))
			case 5:
				lbl.SetText(utils.FormatMoney(d.NetAmount))
			case 6:
				lbl.SetText(utils.StatusLabel(d.Status))
			case 7:
				lbl.SetText(utils.PaymentMethodLabel(d.PaymentMethod))
			}
		},
	)
	for i, w := range colWidths {
		table.SetColumnWidth(i, w)
	}
	table.OnSelected = func(id widget.TableCellID) {
		if id.Row > 0 {
			selectedRow = id.Row - 1
		}
	}

	countLabel := widget.NewLabel(fmt.Sprintf("%d document(s)", len(docs)))

	refresh := func() {
		loadDocs(searchEntry.Text, statusSelect.Selected)
		table.Refresh()
		countLabel.SetText(fmt.Sprintf("%d document(s)", len(docs)))
	}

	searchEntry.OnChanged = func(s string) { refresh() }
	statusSelect.OnChanged = func(s string) { refresh() }

	// Boutons
	newBtn := widget.NewButtonWithIcon("Nouveau", theme.ContentAddIcon(), func() {
		onOpen(0)
	})
	newBtn.Importance = widget.HighImportance

	openBtn := widget.NewButtonWithIcon("Ouvrir", theme.DocumentIcon(), func() {
		if selectedRow < 0 || selectedRow >= len(docs) {
			dialog.ShowInformation("Info", "Sélectionnez un document", app.MainWindow)
			return
		}
		onOpen(docs[selectedRow].ID)
	})

	duplicateBtn := widget.NewButtonWithIcon("Dupliquer", theme.ContentCopyIcon(), func() {
		if selectedRow < 0 || selectedRow >= len(docs) {
			return
		}
		d := docs[selectedRow]
		dialog.ShowConfirm("Dupliquer", fmt.Sprintf("Dupliquer '%s' ?", d.DocNumber), func(ok bool) {
			if !ok {
				return
			}
			svc := services.NewDocumentService(db)
			session := app.GetSession()
			year := 2025
			if session != nil {
				year = session.FiscalYear
			}
			// Convertir vers le même type (= dupliquer)
			userID := 0
			if session != nil {
				userID = session.UserID
			}
			newDoc, err := svc.ConvertDocument(d.ID, d.DocType, year, userID)
			if err != nil {
				dialog.ShowError(err, app.MainWindow)
				return
			}
			refresh()
			dialog.ShowInformation("Succès", fmt.Sprintf("Document %s créé ✅", newDoc.DocNumber), app.MainWindow)
		}, app.MainWindow)
	})

	printBtn := widget.NewButtonWithIcon("Imprimer", theme.DocumentPrintIcon(), func() {
		if selectedRow < 0 || selectedRow >= len(docs) {
			return
		}
		dialog.ShowInformation("Impression", fmt.Sprintf("Impression de %s en cours...", docs[selectedRow].DocNumber), app.MainWindow)
	})

	deleteBtn := widget.NewButtonWithIcon("Supprimer", theme.DeleteIcon(), func() {
		if selectedRow < 0 || selectedRow >= len(docs) {
			return
		}
		d := docs[selectedRow]
		if d.Status != models.StatusDraft {
			dialog.ShowError(fmt.Errorf("seuls les brouillons peuvent être supprimés"), app.MainWindow)
			return
		}
		dialog.ShowConfirm("Supprimer", fmt.Sprintf("Supprimer '%s' ?", d.DocNumber), func(ok bool) {
			if ok {
				db.Exec(`DELETE FROM document_lines WHERE document_id=?`, d.ID)
				db.Exec(`DELETE FROM documents WHERE id=?`, d.ID)
				refresh()
			}
		}, app.MainWindow)
	})
	deleteBtn.Importance = widget.DangerImportance

	// Totaux
	totalsLabel := widget.NewLabel("")
	updateTotals := func() {
		var totalHT, totalNet float64
		for _, d := range docs {
			totalHT += d.TotalHT
			totalNet += d.NetAmount
		}
		totalsLabel.SetText(fmt.Sprintf("Total HT: %s | Net à payer: %s",
			utils.FormatMoney(totalHT), utils.FormatMoney(totalNet)))
	}
	updateTotals()

	header := buildScreenHeader(title, fmt.Sprintf("Gestion des %s", title), color)
	toolbar := container.NewHBox(newBtn, openBtn, duplicateBtn, printBtn, widget.NewSeparator(), deleteBtn, widget.NewSeparator(), countLabel)
	filters := container.NewGridWithColumns(2,
		container.NewBorder(nil, nil, widget.NewLabel("Recherche: "), nil, searchEntry),
		container.NewBorder(nil, nil, widget.NewLabel("Statut: "), nil, statusSelect),
	)

	return container.NewBorder(
		container.NewVBox(header, toolbar, filters, totalsLabel),
		nil, nil, nil,
		table,
	)
}

// ─────────────────────────────────────────────────────────────────────────────
// FORMULAIRE FACTURE DE VENTE (+ Devis, BL, etc.)
// ─────────────────────────────────────────────────────────────────────────────

// BuildSaleInvoiceScreen construit l'écran de saisie de facture de vente
func BuildSaleInvoiceScreen(docID int) fyne.CanvasObject {
	return buildInvoiceScreen(docID, models.DocTypeFA, "🧾 Facture de Vente", "#27ae60")
}

// BuildQuotationsScreen construit l'écran des devis
func BuildQuotationsScreen() fyne.CanvasObject {
	return buildDocumentListScreen(models.DocTypeDV, "📝 Devis", "#8e44ad", func(id int) {
		app.Navigate("sales/quotation_edit", id)
	})
}

// BuildProformaScreen construit l'écran des factures proforma
func BuildProformaScreen() fyne.CanvasObject {
	return buildDocumentListScreen(models.DocTypePF, "📄 Factures Proforma", "#9b59b6", func(id int) {
		app.Navigate("sales/proforma_edit", id)
	})
}

// BuildDeliveryNotesScreen construit l'écran des bons de livraison
func BuildDeliveryNotesScreen() fyne.CanvasObject {
	return buildDocumentListScreen(models.DocTypeBL, "📦 Bons de Livraison", "#16a085", func(id int) {
		app.Navigate("sales/delivery_edit", id)
	})
}

// BuildClientOrdersScreen construit l'écran des commandes clients
func BuildClientOrdersScreen() fyne.CanvasObject {
	return buildDocumentListScreen(models.DocTypeBCC, "📑 Commandes Clients", "#2980b9", func(id int) {
		app.Navigate("sales/order_edit", id)
	})
}

// BuildCreditNotesScreen construit l'écran des avoirs
func BuildCreditNotesScreen() fyne.CanvasObject {
	return buildDocumentListScreen(models.DocTypeAV, "↩️ Avoirs Clients", "#e74c3c", func(id int) {
		app.Navigate("sales/credit_edit", id)
	})
}

// buildInvoiceScreen construit un écran de saisie de document (FA, DV, BL…)
func buildInvoiceScreen(docID int, docType, title, headerColor string) fyne.CanvasObject {
	db := app.GetDB()
	if db == nil {
		return PlaceholderScreen(title + " (DB non connectée)")
	}

	session := app.GetSession()
	userID := 1
	year := 2025
	if session != nil {
		userID = session.UserID
		year = session.FiscalYear
	}

	docSvc := services.NewDocumentService(db)

	// Charger ou créer le document
	var doc *models.Document
	if docID > 0 {
		var err error
		doc, err = queries.GetDocumentByID(db, docID)
		if err != nil {
			return PlaceholderScreen("Erreur chargement document: " + err.Error())
		}
	} else {
		d := docSvc.NewDocument(docType, year, userID)
		doc = d
	}

	// ── Champs en-tête
	docNumLabel := widget.NewLabel(doc.DocNumber)
	docNumLabel.TextStyle = fyne.TextStyle{Bold: true}

	dateEntry := widget.NewEntry()
	dateEntry.SetText(utils.FormatDateFr(doc.Date))

	// Clients (pour FA, DV, PF, BL, CC, AV)
	clients, _ := queries.GetAllClients(db, "")
	clientNames := []string{"-- Aucun --"}
	clientIDMap := map[string]int{"-- Aucun --": 0}
	for _, c := range clients {
		key := fmt.Sprintf("[%s] %s", c.Code, c.NameFr)
		clientNames = append(clientNames, key)
		clientIDMap[key] = c.ID
	}
	clientSelect := widget.NewSelect(clientNames, nil)
	clientSelect.SetSelected("-- Aucun --")
	if doc.ClientID != nil && *doc.ClientID > 0 {
		for _, c := range clients {
			if c.ID == *doc.ClientID {
				key := fmt.Sprintf("[%s] %s", c.Code, c.NameFr)
				clientSelect.SetSelected(key)
				break
			}
		}
	}

	// Mode de paiement
	payModeSelect := widget.NewSelect([]string{"Espèces", "Chèque", "Virement", "Crédit", "Mixte"}, nil)
	payModeDBMap := map[string]string{
		"Espèces": "cash", "Chèque": "cheque",
		"Virement": "transfer", "Crédit": "credit", "Mixte": "mixed",
	}
	payModeSelect.SetSelected("Espèces")
	for label, code := range payModeDBMap {
		if code == doc.PaymentMethod {
			payModeSelect.SetSelected(label)
			break
		}
	}

	// Remise globale
	globalDiscEntry := widget.NewEntry()
	globalDiscEntry.SetText(fmt.Sprintf("%.2f", doc.GlobalDiscountPct))
	globalDiscEntry.SetPlaceHolder("0.00")

	// Notes
	notesEntry := widget.NewMultiLineEntry()
	notesEntry.SetText(doc.Notes)
	notesEntry.SetMinRowsVisible(2)

	// ── Lignes de document
	// Table lignes
	lineHeaders := []string{"Réf.", "Désignation", "Qté", "Unité", "P.U. HT", "Remise%", "Montant HT", "TVA%", "TVA", "TTC"}
	lineColWidths := []float32{70, 180, 60, 50, 90, 60, 100, 50, 80, 100}

	// Variables réactives pour les lignes
	var uiLines []SalesUILine
	for _, l := range doc.Lines {
		uiLines = append(uiLines, SalesUILine{
			Designation: l.Designation,
			Qty:         l.Quantity,
			Unit:        l.Unit,
			UnitPriceHT: l.UnitPriceHT,
			DiscPct:     l.DiscountPercent,
			AmountHT:    l.AmountHT,
			TVARate:     l.TVARate,
			TVAAmount:   l.TVAAmount,
			AmountTTC:   l.AmountTTC,
			LotNumber:   l.LotNumber,
		})
	}

	// ── Totaux
	totalHTLabel := widget.NewLabel("0.00 DA")
	totalTVALabel := widget.NewLabel("0.00 DA")
	totalTTCLabel := widget.NewLabel("0.00 DA")
	timbreLabel := widget.NewLabel("0.00 DA")
	netLabel := widget.NewLabel("0.00 DA")
	netLabel.TextStyle = fyne.TextStyle{Bold: true}
	amountWordsLabel := widget.NewLabel("")
	amountWordsLabel.Wrapping = fyne.TextWrapWord

	// Recalcul des totaux
	recalculate := func() {
		taxCfg := loadTaxConfig(db)
		var totalHT, totalDisc, totalTVA, totalTTC float64
		for _, l := range uiLines {
			totalHT += l.AmountHT
			totalDisc += l.UnitPriceHT * l.Qty * l.DiscPct / 100
			totalTVA += l.TVAAmount
			totalTTC += l.AmountTTC
		}

		globalDisc, _ := strconv.ParseFloat(strings.TrimSpace(globalDiscEntry.Text), 64)
		globalDiscAmt := totalHT * globalDisc / 100
		netHT := totalHT - globalDiscAmt

		// Recalcul TVA proportionnelle
		var tva9, tva19 float64
		for _, l := range uiLines {
			if l.TVARate == 9 {
				tva9 += l.TVAAmount * (1 - globalDisc/100)
			} else if l.TVARate == 19 {
				tva19 += l.TVAAmount * (1 - globalDisc/100)
			}
		}
		newTVA := tva9 + tva19

		ttcAfterDisc := netHT + newTVA
		timbre := 0.0
		pmCode := payModeDBMap[payModeSelect.Selected]
		if taxCfg.AutoTimbre && pmCode == "cash" {
			timbre = utils.CalculateTimbre(ttcAfterDisc, pmCode, taxCfg.TimbreRate, taxCfg.TimbreMax, taxCfg.TimbreExemption, taxCfg.AutoTimbre)
		}
		net := ttcAfterDisc + timbre

		totalHTLabel.SetText(utils.FormatMoney(totalHT))
		totalTVALabel.SetText(utils.FormatMoney(newTVA))
		totalTTCLabel.SetText(utils.FormatMoney(ttcAfterDisc))
		timbreLabel.SetText(utils.FormatMoney(timbre))
		netLabel.SetText(utils.FormatMoney(net))
		amountWordsLabel.SetText("🖊 " + utils.NumberToWordsFr(net))
	}
	recalculate()

	// Table des lignes
	lineTable := widget.NewTable(
		func() (int, int) { return len(uiLines) + 1, len(lineHeaders) },
		func() fyne.CanvasObject {
			lbl := widget.NewLabel("")
			lbl.Truncation = fyne.TextTruncateEllipsis
			return lbl
		},
		func(id widget.TableCellID, cell fyne.CanvasObject) {
			lbl := cell.(*widget.Label)
			if id.Row == 0 {
				lbl.SetText(lineHeaders[id.Col])
				lbl.TextStyle = fyne.TextStyle{Bold: true}
				return
			}
			if id.Row-1 >= len(uiLines) {
				lbl.SetText("")
				return
			}
			l := uiLines[id.Row-1]
			switch id.Col {
			case 0:
				if l.Article != nil {
					lbl.SetText(l.Article.Reference)
				}
			case 1:
				lbl.SetText(l.Designation)
			case 2:
				lbl.SetText(fmt.Sprintf("%.2f", l.Qty))
			case 3:
				lbl.SetText(l.Unit)
			case 4:
				lbl.SetText(fmt.Sprintf("%.4f", l.UnitPriceHT))
			case 5:
				lbl.SetText(fmt.Sprintf("%.2f%%", l.DiscPct))
			case 6:
				lbl.SetText(utils.FormatMoney(l.AmountHT))
			case 7:
				lbl.SetText(fmt.Sprintf("%.0f%%", l.TVARate))
			case 8:
				lbl.SetText(utils.FormatMoney(l.TVAAmount))
			case 9:
				lbl.SetText(utils.FormatMoney(l.AmountTTC))
			}
		},
	)
	for i, w := range lineColWidths {
		lineTable.SetColumnWidth(i, w)
	}

	// ── Bouton: Ajouter une ligne
	addLineBtn := widget.NewButtonWithIcon("Ajouter Article", theme.ContentAddIcon(), func() {
		showArticleLineDialog(db, func(line SalesUILine) {
			uiLines = append(uiLines, line)
			lineTable.Refresh()
			recalculate()
		})
	})
	addLineBtn.Importance = widget.HighImportance

	// Modifier ligne sélectionnée
	selectedLineRow := -1
	lineTable.OnSelected = func(id widget.TableCellID) {
		if id.Row > 0 {
			selectedLineRow = id.Row - 1
		}
	}

	editLineBtn := widget.NewButtonWithIcon("Modifier Ligne", theme.DocumentCreateIcon(), func() {
		if selectedLineRow < 0 || selectedLineRow >= len(uiLines) {
			return
		}
		showEditLineDialog(uiLines[selectedLineRow], func(updated SalesUILine) {
			uiLines[selectedLineRow] = updated
			lineTable.Refresh()
			recalculate()
		})
	})

	deleteLineBtn := widget.NewButtonWithIcon("Suppr. Ligne", theme.DeleteIcon(), func() {
		if selectedLineRow < 0 || selectedLineRow >= len(uiLines) {
			return
		}
		uiLines = append(uiLines[:selectedLineRow], uiLines[selectedLineRow+1:]...)
		selectedLineRow = -1
		lineTable.Refresh()
		recalculate()
	})
	deleteLineBtn.Importance = widget.DangerImportance

	globalDiscEntry.OnChanged = func(s string) { recalculate() }
	payModeSelect.OnChanged = func(s string) { recalculate() }

	// ── Construire le document à partir des champs UI
	buildDoc := func() *models.Document {
		doc.Date = utils.TodayString()
		doc.Notes = notesEntry.Text
		doc.GlobalDiscountPct, _ = strconv.ParseFloat(strings.TrimSpace(globalDiscEntry.Text), 64)
		doc.PaymentMethod = payModeDBMap[payModeSelect.Selected]

		// Client
		selClient := clientSelect.Selected
		if cid, ok := clientIDMap[selClient]; ok && cid > 0 {
			doc.ClientID = &cid
		}

		// Lignes
		doc.Lines = nil
		for i, l := range uiLines {
			line := models.DocumentLine{
				DocumentID:      doc.ID,
				LineNumber:      i + 1,
				Designation:     l.Designation,
				Quantity:        l.Qty,
				Unit:            l.Unit,
				UnitPriceHT:     l.UnitPriceHT,
				DiscountPercent: l.DiscPct,
				AmountHT:        l.AmountHT,
				TVARate:         l.TVARate,
				TVAAmount:       l.TVAAmount,
				AmountTTC:       l.AmountTTC,
				LotNumber:       l.LotNumber,
			}
			if l.Article != nil {
				artID := l.Article.ID
				line.ArticleID = &artID
			}
			doc.Lines = append(doc.Lines, line)
		}
		return doc
	}

	// ── Boutons action
	saveDraftBtn := widget.NewButtonWithIcon("Sauvegarder", theme.DocumentSaveIcon(), func() {
		if len(uiLines) == 0 {
			dialog.ShowError(fmt.Errorf("ajoutez au moins une ligne"), app.MainWindow)
			return
		}
		d := buildDoc()
		taxCfg := loadTaxConfig(db)
		err := docSvc.SaveDraft(d, taxCfg)
		if err != nil {
			dialog.ShowError(err, app.MainWindow)
			return
		}
		docNumLabel.SetText(d.DocNumber)
		dialog.ShowInformation("Sauvegardé", fmt.Sprintf("Document %s sauvegardé ✅", d.DocNumber), app.MainWindow)
	})

	confirmBtn := widget.NewButtonWithIcon("Confirmer & Valider", theme.ConfirmIcon(), func() {
		if len(uiLines) == 0 {
			dialog.ShowError(fmt.Errorf("ajoutez au moins une ligne"), app.MainWindow)
			return
		}
		selClient := clientSelect.Selected
		if _, ok := clientIDMap[selClient]; !ok || clientIDMap[selClient] == 0 {
			if docType == models.DocTypeFA || docType == models.DocTypeBL {
				dialog.ShowError(fmt.Errorf("sélectionnez un client"), app.MainWindow)
				return
			}
		}
		dialog.ShowConfirm("Confirmer",
			"Confirmer ce document?\nLes stocks seront mis à jour.",
			func(ok bool) {
				if !ok {
					return
				}
				// Exécuter en goroutine pour ne pas bloquer l'UI
				go func() {
					d := buildDoc()
					taxCfg := loadTaxConfig(db)
					var err error
					if docType == models.DocTypeFA {
						err = docSvc.ConfirmSaleInvoice(d, taxCfg, userID)
					} else {
						err = docSvc.SaveDraft(d, taxCfg)
						if err == nil {
							db.Exec(`UPDATE documents SET status='confirmed' WHERE id=?`, d.ID)
						}
					}
					if err != nil {
						fyne.Do(func() { dialog.ShowError(err, app.MainWindow) })
						return
					}
					fyne.Do(func() {
						docNumLabel.SetText(d.DocNumber)
						dialog.ShowInformation("✅ Confirmé", fmt.Sprintf("%s confirmé avec succès!", d.DocNumber), app.MainWindow)
						app.Navigate(app.RouteSaleInvoices)
					})
				}()
			}, app.MainWindow)
	})
	confirmBtn.Importance = widget.HighImportance

	printBtn := widget.NewButtonWithIcon("Imprimer PDF", theme.DocumentPrintIcon(), func() {
		if doc.ID == 0 {
			dialog.ShowError(fmt.Errorf("sauvegardez d'abord le document"), app.MainWindow)
			return
		}
		dialog.ShowInformation("Impression", fmt.Sprintf("Génération PDF de %s...", doc.DocNumber), app.MainWindow)
	})

	convertBtn := widget.NewButtonWithIcon("Convertir →", theme.MailForwardIcon(), func() {
		if doc.ID == 0 {
			return
		}
		showConvertDialog(doc, db, docSvc, year)
	})

	cancelBtn := widget.NewButton("Annuler / Retour", func() {
		app.Navigate(app.RouteSaleInvoices)
	})

	// ── Layout en-tête document
	headerForm := widget.NewForm(
		widget.NewFormItem("N° Document:", container.NewHBox(docNumLabel)),
		widget.NewFormItem("Date:", dateEntry),
		widget.NewFormItem("Client:", clientSelect),
		widget.NewFormItem("Mode Paiement:", payModeSelect),
		widget.NewFormItem("Remise Globale (%):", globalDiscEntry),
		widget.NewFormItem("Notes:", notesEntry),
	)

	// ── Panel totaux
	totalsForm := widget.NewForm(
		widget.NewFormItem("Total HT:", totalHTLabel),
		widget.NewFormItem("TVA:", totalTVALabel),
		widget.NewFormItem("Total TTC:", totalTTCLabel),
		widget.NewFormItem("Timbre:", timbreLabel),
		widget.NewFormItem("Net à Payer:", netLabel),
	)

	lineToolbar := container.NewHBox(addLineBtn, editLineBtn, deleteLineBtn)

	rightPanel := container.NewVBox(
		widget.NewCard("Totaux", "", totalsForm),
		widget.NewCard("Montant en Lettres", "", amountWordsLabel),
		container.NewVBox(saveDraftBtn, confirmBtn, printBtn, convertBtn, cancelBtn),
	)

	// Onglets pour l'en-tête et les lignes
	mainContent := container.NewHSplit(
		container.NewBorder(
			container.NewVBox(
				widget.NewCard("En-tête Document", "", headerForm),
				lineToolbar,
			),
			nil, nil, nil,
			lineTable,
		),
		container.NewVScroll(rightPanel),
	)
	mainContent.SetOffset(0.72)

	screenHeader := buildScreenHeader(
		fmt.Sprintf("%s — %s", title, doc.DocNumber),
		"Saisie et gestion du document commercial",
		headerColor,
	)

	return container.NewBorder(screenHeader, nil, nil, nil, mainContent)
}

// showArticleLineDialog affiche la dialogue de saisie d'une ligne article
func showArticleLineDialog(db *sql.DB, onAdd func(line SalesUILine)) {
	articles, _ := queries.GetAllArticles(db, "", 0, "")
	artNames := make([]string, len(articles))
	for i, a := range articles {
		artNames[i] = fmt.Sprintf("[%s] %s — %s", a.Reference, a.NameFr, utils.FormatMoney(a.SalePriceTTC))
	}
	if len(artNames) == 0 {
		dialog.ShowInformation("Info", "Aucun article disponible", app.MainWindow)
		return
	}

	artSelect := widget.NewSelect(artNames, nil)
	artSelect.SetSelected(artNames[0])

	qtyEntry := widget.NewEntry()
	qtyEntry.SetText("1")

	discPctEntry := widget.NewEntry()
	discPctEntry.SetText("0")

	tvaSelect := widget.NewSelect([]string{"0", "9", "19"}, nil)
	tvaSelect.SetSelected("19")

	priceHTEntry := widget.NewEntry()
	priceTTCEntry := widget.NewEntry()
	amountHTLabel := widget.NewLabel("0.00")

	// Pré-remplir depuis l'article sélectionné
	fillFromArticle := func(idx int) {
		if idx < 0 || idx >= len(articles) {
			return
		}
		a := articles[idx]
		priceHTEntry.SetText(fmt.Sprintf("%.4f", a.SalePriceHT))
		priceTTCEntry.SetText(fmt.Sprintf("%.2f", a.SalePriceTTC))
		tvaSelect.SetSelected(fmt.Sprintf("%.0f", a.TVARate))
	}
	fillFromArticle(0)

	artSelect.OnChanged = func(s string) { fillFromArticle(artSelect.SelectedIndex()) }

	recalcLine := func() {
		qty, _ := strconv.ParseFloat(strings.TrimSpace(qtyEntry.Text), 64)
		pu, _ := strconv.ParseFloat(strings.TrimSpace(priceHTEntry.Text), 64)
		disc, _ := strconv.ParseFloat(strings.TrimSpace(discPctEntry.Text), 64)
		ht := qty * pu * (1 - disc/100)
		amountHTLabel.SetText(utils.FormatMoney(ht))
	}
	qtyEntry.OnChanged = func(s string) { recalcLine() }
	priceHTEntry.OnChanged = func(s string) { recalcLine() }
	discPctEntry.OnChanged = func(s string) { recalcLine() }

	content := widget.NewForm(
		widget.NewFormItem("Article *", artSelect),
		widget.NewFormItem("Quantité *", qtyEntry),
		widget.NewFormItem("Prix Unit. HT", priceHTEntry),
		widget.NewFormItem("Prix Unit. TTC", priceTTCEntry),
		widget.NewFormItem("TVA %", tvaSelect),
		widget.NewFormItem("Remise %", discPctEntry),
		widget.NewFormItem("Montant HT", amountHTLabel),
	)

	dialog.ShowCustomConfirm("Ajouter Article", "Ajouter", "Annuler", content, func(ok bool) {
		if !ok {
			return
		}
		idx := artSelect.SelectedIndex()
		if idx < 0 || idx >= len(articles) {
			return
		}
		art := articles[idx]
		qty, _ := strconv.ParseFloat(strings.TrimSpace(qtyEntry.Text), 64)
		if qty <= 0 {
			qty = 1
		}
		pu, _ := strconv.ParseFloat(strings.TrimSpace(priceHTEntry.Text), 64)
		disc, _ := strconv.ParseFloat(strings.TrimSpace(discPctEntry.Text), 64)
		tva, _ := strconv.ParseFloat(tvaSelect.Selected, 64)

		_, clAmountHT, clTVAAmount, clAmountTTC := utils.CalculateLineAmounts(pu, qty, disc, tva)
		line := SalesUILine{
			Article:     &articles[idx],
			Designation: art.NameFr,
			Qty:         qty,
			Unit:        art.UnitSymbol,
			UnitPriceHT: pu,
			DiscPct:     disc,
			AmountHT:    clAmountHT,
			TVARate:     tva,
			TVAAmount:   clTVAAmount,
			AmountTTC:   clAmountTTC,
		}
		onAdd(line)
	}, app.MainWindow)
}

// showEditLineDialog affiche le dialogue de modification d'une ligne
func showEditLineDialog(line SalesUILine, onUpdate func(updated SalesUILine)) {
	qtyEntry := widget.NewEntry()
	qtyEntry.SetText(fmt.Sprintf("%.4f", line.Qty))
	priceEntry := widget.NewEntry()
	priceEntry.SetText(fmt.Sprintf("%.4f", line.UnitPriceHT))
	discEntry := widget.NewEntry()
	discEntry.SetText(fmt.Sprintf("%.2f", line.DiscPct))
	tvaSelect := widget.NewSelect([]string{"0", "9", "19"}, nil)
	tvaSelect.SetSelected(fmt.Sprintf("%.0f", line.TVARate))

	content := widget.NewForm(
		widget.NewFormItem("Désignation", widget.NewLabel(line.Designation)),
		widget.NewFormItem("Quantité", qtyEntry),
		widget.NewFormItem("Prix Unit. HT", priceEntry),
		widget.NewFormItem("Remise %", discEntry),
		widget.NewFormItem("TVA %", tvaSelect),
	)
	dialog.ShowCustomConfirm("Modifier Ligne", "Enregistrer", "Annuler", content, func(ok bool) {
		if !ok {
			return
		}
		qty, _ := strconv.ParseFloat(strings.TrimSpace(qtyEntry.Text), 64)
		pu, _ := strconv.ParseFloat(strings.TrimSpace(priceEntry.Text), 64)
		disc, _ := strconv.ParseFloat(strings.TrimSpace(discEntry.Text), 64)
		tva, _ := strconv.ParseFloat(tvaSelect.Selected, 64)
		_, elAmountHT, elTVAAmount, elAmountTTC := utils.CalculateLineAmounts(pu, qty, disc, tva)
		updated := SalesUILine{
			Article:     line.Article,
			Designation: line.Designation,
			Qty:         qty,
			Unit:        line.Unit,
			UnitPriceHT: pu,
			DiscPct:     disc,
			AmountHT:    elAmountHT,
			TVARate:     tva,
			TVAAmount:   elTVAAmount,
			AmountTTC:   elAmountTTC,
			LotNumber:   line.LotNumber,
		}
		onUpdate(updated)
	}, app.MainWindow)
}

// showConvertDialog affiche le dialogue de conversion de document
func showConvertDialog(doc *models.Document, db *sql.DB, svc *services.DocumentService, year int) {
	convMap := map[string][]string{
		models.DocTypeDV: {models.DocTypeFA, models.DocTypePF, models.DocTypeBL},
		models.DocTypePF: {models.DocTypeFA, models.DocTypeBL},
		models.DocTypeFA: {models.DocTypeBL, models.DocTypeAV},
		models.DocTypeBL: {models.DocTypeFA},
		models.DocTypeBCC: {models.DocTypeFA, models.DocTypeBL},
	}

	targets, ok := convMap[doc.DocType]
	if !ok || len(targets) == 0 {
		dialog.ShowInformation("Conversion", "Pas de conversion disponible pour ce type", app.MainWindow)
		return
	}

	docTypeLabels := map[string]string{
		models.DocTypeFA:  "Facture de Vente (FA)",
		models.DocTypePF:  "Facture Proforma (PF)",
		models.DocTypeBL:  "Bon de Livraison (BL)",
		models.DocTypeAV:  "Avoir (AV)",
		models.DocTypeFAC: "Facture d'Achat (FAC)",
	}

	targetLabels := make([]string, len(targets))
	for i, t := range targets {
		if l, ok := docTypeLabels[t]; ok {
			targetLabels[i] = l
		} else {
			targetLabels[i] = t
		}
	}

	targetSelect := widget.NewSelect(targetLabels, nil)
	if len(targetLabels) > 0 {
		targetSelect.SetSelected(targetLabels[0])
	}

	dialog.ShowCustomConfirm(
		fmt.Sprintf("Convertir %s", doc.DocNumber),
		"Convertir", "Annuler",
		container.NewVBox(
			widget.NewLabel(fmt.Sprintf("Document source: %s (%s)", doc.DocNumber, doc.DocType)),
			widget.NewForm(widget.NewFormItem("Convertir en:", targetSelect)),
		),
		func(ok bool) {
			if !ok {
				return
			}
			idx := targetSelect.SelectedIndex()
			if idx < 0 || idx >= len(targets) {
				return
			}
			convUserID := 0
			if s2 := app.GetSession(); s2 != nil {
				convUserID = s2.UserID
			}
			newDoc, err := svc.ConvertDocument(doc.ID, targets[idx], year, convUserID)
			if err != nil {
				dialog.ShowError(err, app.MainWindow)
				return
			}
			dialog.ShowInformation("Converti", fmt.Sprintf("Nouveau document créé: %s ✅", newDoc.DocNumber), app.MainWindow)
		}, app.MainWindow)
}

// ─────────────────────────────────────────────────────────────────────────────
// POS — POINT DE VENTE
// ─────────────────────────────────────────────────────────────────────────────

// BuildPOSScreen construit l'écran Point de Vente
func BuildPOSScreen(w fyne.Window) fyne.CanvasObject {
	db := app.GetDB()
	if db == nil {
		return PlaceholderScreen("POS (DB non connectée)")
	}

	session := app.GetSession()
	userID := 1
	year := 2025
	if session != nil {
		userID = session.UserID
		year = session.FiscalYear
	}
	_ = userID
	_ = year

	// ── Recherche produit
	barcodeEntry := widget.NewEntry()
	barcodeEntry.SetPlaceHolder("🔍 Scannez code-barres ou recherchez un article...")

	// ── Panier
	type CartItem struct {
		Article models.Article
		Qty     float64
		PriceHT float64
		TVARate float64
		LineTTC float64
	}

	var cartItems []CartItem

	// ── Labels totaux
	subTotalLabel := widget.NewLabel("0.00 DA")
	tvaLabel := widget.NewLabel("0.00 DA")
	totalLabel := widget.NewLabel("0.00 DA")
	totalLabel.TextStyle = fyne.TextStyle{Bold: true}

	recalcCart := func() {
		var totalHT, totalTVA, totalTTC float64
		for _, item := range cartItems {
			_, cHT, cTVA, cTTC := utils.CalculateLineAmounts(item.PriceHT, item.Qty, 0, item.TVARate)
			totalHT += cHT
			totalTVA += cTVA
			totalTTC += cTTC
		}
		subTotalLabel.SetText(utils.FormatMoney(totalHT))
		tvaLabel.SetText(utils.FormatMoney(totalTVA))
		totalLabel.SetText(utils.FormatMoney(totalTTC))
	}

	// Table du panier
	cartHeaders := []string{"Article", "Qté", "P.U. TTC", "Total TTC", ""}
	cartColWidths := []float32{200, 60, 90, 100, 50}

	cartTable := widget.NewTable(
		func() (int, int) { return len(cartItems) + 1, len(cartHeaders) },
		func() fyne.CanvasObject { return widget.NewLabel("") },
		func(id widget.TableCellID, cell fyne.CanvasObject) {
			lbl := cell.(*widget.Label)
			if id.Row == 0 {
				lbl.SetText(cartHeaders[id.Col])
				lbl.TextStyle = fyne.TextStyle{Bold: true}
				return
			}
			if id.Row-1 >= len(cartItems) {
				lbl.SetText("")
				return
			}
			item := cartItems[id.Row-1]
			switch id.Col {
			case 0:
				lbl.SetText(item.Article.NameFr)
			case 1:
				lbl.SetText(fmt.Sprintf("%.2f", item.Qty))
			case 2:
				lbl.SetText(utils.FormatMoney(item.Article.SalePriceTTC))
			case 3:
				lbl.SetText(utils.FormatMoney(item.LineTTC))
			case 4:
				lbl.SetText("✕")
			}
		},
	)
	for i, cw := range cartColWidths {
		cartTable.SetColumnWidth(i, cw)
	}

	cartTable.OnSelected = func(id widget.TableCellID) {
		if id.Row > 0 && id.Col == 4 {
			// Supprimer article
			idx := id.Row - 1
			if idx < len(cartItems) {
				cartItems = append(cartItems[:idx], cartItems[idx+1:]...)
				cartTable.Refresh()
				recalcCart()
			}
		}
	}

	// Ajouter article au panier
	addToCart := func(code string) {
		if db == nil || code == "" {
			return
		}
		art, err := queries.FindArticleByBarcode(db, code)
		if err != nil {
			// Essayer une recherche textuelle
			arts, _ := queries.GetAllArticles(db, code, 0, "")
			if len(arts) == 0 {
				dialog.ShowError(fmt.Errorf("article non trouvé: %s", code), app.MainWindow)
				return
			}
			a := arts[0]
			art = &a
		}

		// Vérifier si déjà dans le panier
		for i, item := range cartItems {
			if item.Article.ID == art.ID {
				cartItems[i].Qty++
				_, _, _, cTTC1 := utils.CalculateLineAmounts(art.SalePriceHT, cartItems[i].Qty, 0, art.TVARate)
				cartItems[i].LineTTC = cTTC1
				cartTable.Refresh()
				recalcCart()
				barcodeEntry.SetText("")
				return
			}
		}

		_, _, _, cTTC2 := utils.CalculateLineAmounts(art.SalePriceHT, 1, 0, art.TVARate)
		cartItems = append(cartItems, CartItem{
			Article: *art,
			Qty:     1,
			PriceHT: art.SalePriceHT,
			TVARate: art.TVARate,
			LineTTC: cTTC2,
		})
		cartTable.Refresh()
		recalcCart()
		barcodeEntry.SetText("")
	}

	barcodeEntry.OnSubmitted = func(s string) {
		addToCart(strings.TrimSpace(s))
	}

	// Bouton recherche article (F2)
	searchBtn := widget.NewButtonWithIcon("F2 Rechercher", theme.SearchIcon(), func() {
		arts, _ := queries.GetAllArticles(db, strings.TrimSpace(barcodeEntry.Text), 0, "")
		if len(arts) == 0 {
			dialog.ShowInformation("Recherche", "Aucun article trouvé", app.MainWindow)
			return
		}
		artNames := make([]string, len(arts))
		for i, a := range arts {
			artNames[i] = fmt.Sprintf("[%s] %s — %s", a.Reference, a.NameFr, utils.FormatMoney(a.SalePriceTTC))
		}
		sel := widget.NewSelect(artNames, nil)
		sel.SetSelected(artNames[0])
		qtyE := widget.NewEntry()
		qtyE.SetText("1")
		dialog.ShowCustomConfirm("Sélectionner Article", "Ajouter", "Annuler",
			widget.NewForm(
				widget.NewFormItem("Article:", sel),
				widget.NewFormItem("Quantité:", qtyE),
			),
			func(ok bool) {
				if !ok {
					return
				}
				idx := sel.SelectedIndex()
				if idx < 0 {
					return
				}
				art := arts[idx]
				qty, _ := strconv.ParseFloat(strings.TrimSpace(qtyE.Text), 64)
				if qty <= 0 {
					qty = 1
				}
				_, _, _, cTTC3 := utils.CalculateLineAmounts(art.SalePriceHT, qty, 0, art.TVARate)
				cartItems = append(cartItems, CartItem{
					Article: art,
					Qty:     qty,
					PriceHT: art.SalePriceHT,
					TVARate: art.TVARate,
					LineTTC: cTTC3,
				})
				cartTable.Refresh()
				recalcCart()
			}, app.MainWindow)
	})

	// ── Paiement
	payBtn := widget.NewButtonWithIcon("F8 Payer Espèces", theme.ConfirmIcon(), func() {
		if len(cartItems) == 0 {
			dialog.ShowError(fmt.Errorf("le panier est vide"), app.MainWindow)
			return
		}
		// Calculer le montant total
		var totalTTC float64
		for _, item := range cartItems {
			totalTTC += item.LineTTC
		}

		givenEntry := widget.NewEntry()
		givenEntry.SetText(fmt.Sprintf("%.2f", totalTTC))
		changeLabel := widget.NewLabel("Rendu: 0.00 DA")
		givenEntry.OnChanged = func(s string) {
			given, _ := strconv.ParseFloat(strings.TrimSpace(s), 64)
			change := given - totalTTC
			if change < 0 {
				changeLabel.SetText(fmt.Sprintf("Manque: %s", utils.FormatMoney(-change)))
			} else {
				changeLabel.SetText(fmt.Sprintf("Rendu: %s", utils.FormatMoney(change)))
			}
		}

		dialog.ShowCustomConfirm("💵 Paiement Espèces",
			"Valider", "Annuler",
			container.NewVBox(
				widget.NewLabel(fmt.Sprintf("Montant à payer: %s", utils.FormatMoney(totalTTC))),
				widget.NewForm(widget.NewFormItem("Montant remis:", givenEntry)),
				changeLabel,
			),
			func(ok bool) {
				if !ok {
					return
				}
				given, _ := strconv.ParseFloat(strings.TrimSpace(givenEntry.Text), 64)
				if given < totalTTC {
					dialog.ShowError(fmt.Errorf("montant insuffisant"), app.MainWindow)
					return
				}
				// Snapshot du panier pour goroutine
				cartSnap := make([]CartItem, len(cartItems))
				copy(cartSnap, cartItems)
				ttcSnap := totalTTC
				go func() {
					posDocSvc := services.NewDocumentService(db)
					posDoc := posDocSvc.NewDocument(models.DocTypeFA, year, userID)
					posDoc.PaymentMethod = "cash"
					for i, item := range cartSnap {
						_, lHT, lTVA, lTTC := utils.CalculateLineAmounts(item.PriceHT, item.Qty, 0, item.TVARate)
						lid := item.Article.ID
						posDoc.Lines = append(posDoc.Lines, models.DocumentLine{
							LineNumber:  i + 1,
							ArticleID:   &lid,
							Designation: item.Article.NameFr,
							Quantity:    item.Qty,
							Unit:        item.Article.UnitSymbol,
							UnitPriceHT: item.PriceHT,
							AmountHT:    lHT,
							TVARate:     item.TVARate,
							TVAAmount:   lTVA,
							AmountTTC:   lTTC,
						})
					}
					taxCfg := loadTaxConfig(db)
					err := posDocSvc.ConfirmSaleInvoice(posDoc, taxCfg, userID)
					if err != nil {
						fyne.Do(func() { dialog.ShowError(err, app.MainWindow) })
						return
					}
					change := given - ttcSnap
					fyne.Do(func() {
						cartItems = nil
						cartTable.Refresh()
						recalcCart()
						msg := fmt.Sprintf("Facture %s créée ✅\nMontant: %s\nRendu monnaie: %s",
							posDoc.DocNumber, utils.FormatMoney(ttcSnap), utils.FormatMoney(change))
						dialog.ShowInformation("✅ Vente enregistrée", msg, app.MainWindow)
					})
				}()
			}, app.MainWindow)
	})
	payBtn.Importance = widget.HighImportance

	clearBtn := widget.NewButtonWithIcon("Vider Panier", theme.DeleteIcon(), func() {
		dialog.ShowConfirm("Vider le panier", "Annuler la vente en cours ?", func(ok bool) {
			if ok {
				cartItems = nil
				cartTable.Refresh()
				recalcCart()
			}
		}, app.MainWindow)
	})
	clearBtn.Importance = widget.DangerImportance

	// ── Layout POS
	searchBar := container.NewBorder(nil, nil, nil,
		container.NewHBox(searchBtn),
		barcodeEntry,
	)

	totalsPanel := container.NewVBox(
		widget.NewSeparator(),
		widget.NewForm(
			widget.NewFormItem("Sous-total HT:", subTotalLabel),
			widget.NewFormItem("TVA:", tvaLabel),
			widget.NewFormItem("TOTAL TTC:", totalLabel),
		),
		widget.NewSeparator(),
		container.NewGridWithColumns(2, clearBtn, payBtn),
	)

	posHeader := buildScreenHeader("🛒 Point de Vente",
		"Vente rapide avec lecture code-barres — F2: Recherche · F8: Paiement", "#c0392b")

	return container.NewBorder(
		container.NewVBox(posHeader, searchBar),
		totalsPanel,
		nil, nil,
		cartTable,
	)
}

// ─────────────────────────────────────────────────────────────────────────────
// Helpers internes
// ─────────────────────────────────────────────────────────────────────────────

// loadTaxConfig charge la config fiscale depuis la DB
func loadTaxConfig(db *sql.DB) models.TaxConfig {
	cfg := models.TaxConfig{
		TVANormal:       19,
		TVAReduced:      9,
		TimbreRate:      1,
		TimbreMax:       2500,
		TimbreExemption: 1000,
		AutoTimbre:      true,
	}
	if db == nil {
		return cfg
	}
	db.QueryRow(`SELECT COALESCE(value,'19') FROM settings WHERE key='tva_normal'`).Scan(&cfg.TVANormal)
	db.QueryRow(`SELECT COALESCE(value,'1') FROM settings WHERE key='timbre_rate'`).Scan(&cfg.TimbreRate)
	db.QueryRow(`SELECT COALESCE(value,'2500') FROM settings WHERE key='timbre_max'`).Scan(&cfg.TimbreMax)
	return cfg
}

// ─────────────────────────────────────────────────────────────────────────────
// LISTE FACTURES ACHAT
// ─────────────────────────────────────────────────────────────────────────────

// BuildPurchaseInvoicesListScreen construit la liste des factures d'achat
func BuildPurchaseInvoicesListScreen() fyne.CanvasObject {
	return buildDocumentListScreen(models.DocTypeFAC, "🛒 Factures d'Achat", "#2980b9", func(id int) {
		app.Navigate(app.RoutePurchaseInvoice, id)
	})
}

// BuildPurchaseInvoiceScreen construit l'écran de saisie de facture d'achat
func BuildPurchaseInvoiceScreen(docID int) fyne.CanvasObject {
	return buildPurchaseInvoiceScreen(docID)
}

func buildPurchaseInvoiceScreen(docID int) fyne.CanvasObject {
	db := app.GetDB()
	if db == nil {
		return PlaceholderScreen("Facture d'Achat (DB non connectée)")
	}

	session := app.GetSession()
	userID := 1
	year := 2025
	if session != nil {
		userID = session.UserID
		year = session.FiscalYear
	}

	docSvc := services.NewDocumentService(db)

	var doc *models.Document
	if docID > 0 {
		var err error
		doc, err = queries.GetDocumentByID(db, docID)
		if err != nil {
			return PlaceholderScreen("Erreur chargement: " + err.Error())
		}
	} else {
		d := docSvc.NewDocument(models.DocTypeFAC, year, userID)
		doc = d
	}

	docNumLabel := widget.NewLabel(doc.DocNumber)
	docNumLabel.TextStyle = fyne.TextStyle{Bold: true}

	dateEntry := widget.NewEntry()
	dateEntry.SetText(utils.FormatDateFr(doc.Date))

	supplierInvEntry := widget.NewEntry()
	supplierInvEntry.SetPlaceHolder("N° facture fournisseur...")
	if doc.SupplierInvoiceNumber != "" {
		supplierInvEntry.SetText(doc.SupplierInvoiceNumber)
	}

	// Fournisseurs
	clientSvc := services.NewClientService(db)
	suppliers, _ := clientSvc.GetAllSuppliers("")
	supNames := []string{"-- Aucun --"}
	supIDMap := map[string]int{"-- Aucun --": 0}
	for _, s := range suppliers {
		key := fmt.Sprintf("[%s] %s", s.Code, s.NameFr)
		supNames = append(supNames, key)
		supIDMap[key] = s.ID
	}
	supSelect := widget.NewSelect(supNames, nil)
	supSelect.SetSelected("-- Aucun --")
	if doc.SupplierID != nil && *doc.SupplierID > 0 {
		for _, s := range suppliers {
			if s.ID == *doc.SupplierID {
				key := fmt.Sprintf("[%s] %s", s.Code, s.NameFr)
				supSelect.SetSelected(key)
				break
			}
		}
	}

	payModeSelect := widget.NewSelect([]string{"Espèces", "Chèque", "Virement", "Crédit", "Mixte"}, nil)
	payModeSelect.SetSelected("Crédit")
	payModeDBMap := map[string]string{
		"Espèces": "cash", "Chèque": "cheque",
		"Virement": "transfer", "Crédit": "credit", "Mixte": "mixed",
	}

	notesEntry := widget.NewMultiLineEntry()
	notesEntry.SetText(doc.Notes)
	notesEntry.SetMinRowsVisible(2)

	// Totaux
	totalHTLabel := widget.NewLabel("0.00 DA")
	totalTVALabel := widget.NewLabel("0.00 DA")
	netLabel := widget.NewLabel("0.00 DA")
	netLabel.TextStyle = fyne.TextStyle{Bold: true}

	type PurchaseLine struct {
		Article     *models.Article
		Designation string
		Qty         float64
		Unit        string
		UnitPriceHT float64
		TVARate     float64
		AmountHT    float64
		TVAAmount   float64
		AmountTTC   float64
	}

	var purchaseLines []PurchaseLine
	for _, l := range doc.Lines {
		purchaseLines = append(purchaseLines, PurchaseLine{
			Designation: l.Designation,
			Qty:         l.Quantity,
			Unit:        l.Unit,
			UnitPriceHT: l.UnitPriceHT,
			TVARate:     l.TVARate,
			AmountHT:    l.AmountHT,
			TVAAmount:   l.TVAAmount,
			AmountTTC:   l.AmountTTC,
		})
	}

	recalc := func() {
		var ht, tva float64
		for _, l := range purchaseLines {
			ht += l.AmountHT
			tva += l.TVAAmount
		}
		totalHTLabel.SetText(utils.FormatMoney(ht))
		totalTVALabel.SetText(utils.FormatMoney(tva))
		netLabel.SetText(utils.FormatMoney(ht + tva))
	}
	recalc()

	lineHeaders := []string{"Désignation", "Qté", "P.U. HT", "TVA%", "Montant HT", "TVA", "TTC"}
	lineColWidths := []float32{200, 60, 100, 50, 110, 90, 110}

	lineTable := widget.NewTable(
		func() (int, int) { return len(purchaseLines) + 1, len(lineHeaders) },
		func() fyne.CanvasObject { return widget.NewLabel("") },
		func(id widget.TableCellID, cell fyne.CanvasObject) {
			lbl := cell.(*widget.Label)
			if id.Row == 0 {
				lbl.SetText(lineHeaders[id.Col])
				lbl.TextStyle = fyne.TextStyle{Bold: true}
				return
			}
			if id.Row-1 >= len(purchaseLines) {
				lbl.SetText("")
				return
			}
			l := purchaseLines[id.Row-1]
			switch id.Col {
			case 0:
				lbl.SetText(l.Designation)
			case 1:
				lbl.SetText(fmt.Sprintf("%.2f", l.Qty))
			case 2:
				lbl.SetText(fmt.Sprintf("%.4f", l.UnitPriceHT))
			case 3:
				lbl.SetText(fmt.Sprintf("%.0f%%", l.TVARate))
			case 4:
				lbl.SetText(utils.FormatMoney(l.AmountHT))
			case 5:
				lbl.SetText(utils.FormatMoney(l.TVAAmount))
			case 6:
				lbl.SetText(utils.FormatMoney(l.AmountTTC))
			}
		},
	)
	for i, cw := range lineColWidths {
		lineTable.SetColumnWidth(i, cw)
	}

	addLineBtn := widget.NewButtonWithIcon("Ajouter Article", theme.ContentAddIcon(), func() {
		desigE := widget.NewEntry()
		desigE.SetPlaceHolder("Description de l'article...")
		qtyE := widget.NewEntry()
		qtyE.SetText("1")
		puE := widget.NewEntry()
		puE.SetPlaceHolder("Prix unitaire HT")
		tvaS := widget.NewSelect([]string{"0", "9", "19"}, nil)
		tvaS.SetSelected("19")

		dialog.ShowCustomConfirm("Nouvelle Ligne d'Achat", "Ajouter", "Annuler",
			widget.NewForm(
				widget.NewFormItem("Désignation *", desigE),
				widget.NewFormItem("Quantité", qtyE),
				widget.NewFormItem("P.U. HT (DA)", puE),
				widget.NewFormItem("TVA %", tvaS),
			),
			func(ok bool) {
				if !ok || strings.TrimSpace(desigE.Text) == "" {
					return
				}
				qty, _ := strconv.ParseFloat(strings.TrimSpace(qtyE.Text), 64)
				pu, _ := strconv.ParseFloat(strings.TrimSpace(puE.Text), 64)
				tva, _ := strconv.ParseFloat(tvaS.Selected, 64)
				if qty <= 0 {
					qty = 1
				}
				_, pHT, pTVA, pTTC := utils.CalculateLineAmounts(pu, qty, 0, tva)
				purchaseLines = append(purchaseLines, PurchaseLine{
					Designation: desigE.Text,
					Qty:         qty,
					Unit:        "U",
					UnitPriceHT: pu,
					TVARate:     tva,
					AmountHT:    pHT,
					TVAAmount:   pTVA,
					AmountTTC:   pTTC,
				})
				lineTable.Refresh()
				recalc()
			}, app.MainWindow)
	})
	addLineBtn.Importance = widget.HighImportance

	saveDraftBtn := widget.NewButtonWithIcon("Sauvegarder", theme.DocumentSaveIcon(), func() {
		doc.Notes = notesEntry.Text
		doc.PaymentMethod = payModeDBMap[payModeSelect.Selected]
		if sn := strings.TrimSpace(supplierInvEntry.Text); sn != "" {
			doc.SupplierInvoiceNumber = sn
		}
		selSup := supSelect.Selected
		if sid, ok := supIDMap[selSup]; ok && sid > 0 {
			doc.SupplierID = &sid
		}
		doc.Lines = nil
		for i, l := range purchaseLines {
			doc.Lines = append(doc.Lines, models.DocumentLine{
				LineNumber:  i + 1,
				Designation: l.Designation,
				Quantity:    l.Qty,
				Unit:        l.Unit,
				UnitPriceHT: l.UnitPriceHT,
				TVARate:     l.TVARate,
				AmountHT:    l.AmountHT,
				TVAAmount:   l.TVAAmount,
				AmountTTC:   l.AmountTTC,
			})
		}
		taxCfg := loadTaxConfig(db)
		err := docSvc.SaveDraft(doc, taxCfg)
		if err != nil {
			dialog.ShowError(err, app.MainWindow)
			return
		}
		docNumLabel.SetText(doc.DocNumber)
		dialog.ShowInformation("Sauvegardé", fmt.Sprintf("%s sauvegardé ✅", doc.DocNumber), app.MainWindow)
	})

	confirmBtn := widget.NewButtonWithIcon("Confirmer Achat", theme.ConfirmIcon(), func() {
		if len(purchaseLines) == 0 {
			dialog.ShowError(fmt.Errorf("ajoutez au moins une ligne"), app.MainWindow)
			return
		}
		dialog.ShowConfirm("Confirmer", "Confirmer cet achat?\nLes stocks seront mis à jour.", func(ok bool) {
			if !ok {
				return
			}
			go func() {
				// Construire le doc
				doc.Notes = notesEntry.Text
				doc.PaymentMethod = payModeDBMap[payModeSelect.Selected]
				selSup := supSelect.Selected
				if sid, ok2 := supIDMap[selSup]; ok2 && sid > 0 {
					doc.SupplierID = &sid
				}
				doc.Lines = nil
				for i, l := range purchaseLines {
					doc.Lines = append(doc.Lines, models.DocumentLine{
						LineNumber:  i + 1,
						Designation: l.Designation,
						Quantity:    l.Qty,
						Unit:        l.Unit,
						UnitPriceHT: l.UnitPriceHT,
						TVARate:     l.TVARate,
						AmountHT:    l.AmountHT,
						TVAAmount:   l.TVAAmount,
						AmountTTC:   l.AmountTTC,
					})
				}
				taxCfg := loadTaxConfig(db)
				err := docSvc.ConfirmPurchaseInvoice(doc, taxCfg, userID, true)
				if err != nil {
					fyne.Do(func() { dialog.ShowError(err, app.MainWindow) })
					return
				}
				fyne.Do(func() {
					dialog.ShowInformation("✅ Confirmé", fmt.Sprintf("%s confirmé!\nStocks mis à jour", doc.DocNumber), app.MainWindow)
					app.Navigate(app.RoutePurchaseInvoices)
				})
			}()
		}, app.MainWindow)
	})
	confirmBtn.Importance = widget.HighImportance

	cancelBtn := widget.NewButton("Retour", func() {
		app.Navigate(app.RoutePurchaseInvoices)
	})

	headerForm := widget.NewForm(
		widget.NewFormItem("N° Document:", docNumLabel),
		widget.NewFormItem("Date:", dateEntry),
		widget.NewFormItem("Fournisseur:", supSelect),
		widget.NewFormItem("N° Fact. Fournisseur:", supplierInvEntry),
		widget.NewFormItem("Mode Paiement:", payModeSelect),
		widget.NewFormItem("Notes:", notesEntry),
	)

	totalsPanel := widget.NewForm(
		widget.NewFormItem("Total HT:", totalHTLabel),
		widget.NewFormItem("TVA:", totalTVALabel),
		widget.NewFormItem("Net à Payer:", netLabel),
	)

	rightPanel := container.NewVBox(
		widget.NewCard("Totaux", "", totalsPanel),
		saveDraftBtn, confirmBtn, cancelBtn,
	)

	mainSplit := container.NewHSplit(
		container.NewBorder(
			container.NewVBox(widget.NewCard("En-tête", "", headerForm), addLineBtn),
			nil, nil, nil, lineTable,
		),
		container.NewVScroll(rightPanel),
	)
	mainSplit.SetOffset(0.72)

	screenHeader := buildScreenHeader(
		fmt.Sprintf("🛒 Facture d'Achat — %s", doc.DocNumber),
		"Saisie des factures fournisseurs avec mise à jour des stocks", "#2980b9",
	)
	return container.NewBorder(screenHeader, nil, nil, nil, mainSplit)
}

// BuildReceptionNotesScreen — Bons de réception
func BuildReceptionNotesScreen() fyne.CanvasObject {
	return buildDocumentListScreen("BR", "📦 Bons de Réception", "#1abc9c", func(id int) {
		app.Navigate("purchases/reception_edit", id)
	})
}

// BuildSupplierOrdersScreen — Commandes fournisseurs
func BuildSupplierOrdersScreen() fyne.CanvasObject {
	return buildDocumentListScreen("BCF", "📑 Commandes Fournisseurs", "#3498db", func(id int) {
		app.Navigate("purchases/order_edit", id)
	})
}

// BuildPurchaseReturnsScreen — Retours fournisseurs
func BuildPurchaseReturnsScreen() fyne.CanvasObject {
	return buildDocumentListScreen("BCC", "↪️ Retours Fournisseurs", "#e74c3c", func(id int) {
		app.Navigate("purchases/return_edit", id)
	})
}
