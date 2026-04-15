package screens

// ─────────────────────────────────────────────────────────────────────────────
// Lot 8 — TRÉSORERIE
//   BuildCashScreen          → Journal de caisse
//   BuildBankScreen          → Comptes bancaires + relevé mouvements
//   BuildChequesScreen       → Gestion chèques reçus / émis
//   BuildCollectionsScreen   → Encaissements clients
//   BuildDisbursementsScreen → Décaissements fournisseurs
//   BuildAgingScreen         → Balance âgée clients + fournisseurs
//   BuildExpensesScreen      → Dépenses diverses
// ─────────────────────────────────────────────────────────────────────────────

import (
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	appstate "gestion-commerciale/internal/app"
	"gestion-commerciale/internal/models"
	"gestion-commerciale/internal/services"
	"gestion-commerciale/pkg/utils"
)

// ─────────────────────────────────────────────────────────────────────────────
// HELPERS INTERNES TRÉSORERIE
// ─────────────────────────────────────────────────────────────────────────────

// buildTreasuryHeader construit un bandeau coloré pour les sous-écrans trésorerie
func buildTreasuryHeader(icon, title, subtitle, hexColor string) fyne.CanvasObject {
	return buildScreenHeader(icon+" "+title, subtitle, hexColor)
}

// buildPeriodFilter retourne un row avec deux date entries (from/to) + un bouton Filtrer
func buildPeriodFilter(fromEntry, toEntry *widget.Entry, onFilter func()) fyne.CanvasObject {
	fromEntry.SetPlaceHolder("AAAA-MM-JJ")
	toEntry.SetPlaceHolder("AAAA-MM-JJ")

	now := time.Now()
	firstDay := fmt.Sprintf("%d-%02d-01", now.Year(), now.Month())
	lastDay := utils.TodayString()
	fromEntry.SetText(firstDay)
	toEntry.SetText(lastDay)

	filterBtn := widget.NewButtonWithIcon("Filtrer", theme.SearchIcon(), onFilter)
	filterBtn.Importance = widget.HighImportance

	todayBtn := widget.NewButton("Aujourd'hui", func() {
		today := utils.TodayString()
		fromEntry.SetText(today)
		toEntry.SetText(today)
		onFilter()
	})
	monthBtn := widget.NewButton("Ce mois", func() {
		fromEntry.SetText(firstDay)
		toEntry.SetText(lastDay)
		onFilter()
	})
	yearBtn := widget.NewButton("Cette année", func() {
		fromEntry.SetText(fmt.Sprintf("%d-01-01", now.Year()))
		toEntry.SetText(lastDay)
		onFilter()
	})

	return container.NewHBox(
		widget.NewLabel("Du :"), fromEntry,
		widget.NewLabel("Au :"), toEntry,
		filterBtn,
		widget.NewSeparator(),
		todayBtn, monthBtn, yearBtn,
	)
}

// buildAmountRow construit une ligne KPI montant colorée
func buildAmountRow(label, amount, color string) fyne.CanvasObject {
	lbl := widget.NewLabel(label)
	lbl.TextStyle = fyne.TextStyle{Bold: true}
	amtLbl := widget.NewRichTextFromMarkdown("**" + amount + "**")
	return container.NewBorder(nil, nil, lbl, amtLbl)
}

// ─────────────────────────────────────────────────────────────────────────────
// 1. CAISSE
// ─────────────────────────────────────────────────────────────────────────────

// BuildCashScreen construit le journal de caisse
func BuildCashScreen() fyne.CanvasObject {
	db := appstate.GetDB()

	header := buildTreasuryHeader("🏦", "Journal de Caisse",
		"Mouvements d'entrées et sorties de caisse", "#27ae60")

	// ── Filtres période
	fromEntry := widget.NewEntry()
	toEntry := widget.NewEntry()

	// ── Résumé KPIs
	balanceLbl := widget.NewLabel("Solde: 0,00 DA")
	balanceLbl.TextStyle = fyne.TextStyle{Bold: true}
	totalInLbl := widget.NewLabel("Entrées: 0,00 DA")
	totalOutLbl := widget.NewLabel("Sorties: 0,00 DA")

	// ── Table des mouvements
	colHeaders := []string{"Date", "Type", "Catégorie", "Tiers", "Description", "Référence", "Entrée (DA)", "Sortie (DA)", "Solde (DA)"}
	colWidths := []float32{90, 60, 110, 120, 180, 110, 110, 110, 110}

	tableContainer := container.NewStack()

	var movements []models.CashMovement
	var currentBalance float64

	loadMovements := func() {
		if db == nil {
			return
		}
		svc := services.NewPaymentService(db)
		mvs, bal, err := svc.GetCashMovements(fromEntry.Text, toEntry.Text)
		if err == nil {
			movements = mvs
			currentBalance = bal
		}

		// KPIs
		var totalIn, totalOut float64
		for _, m := range movements {
			if m.Type == "in" {
				totalIn += m.Amount
			} else {
				totalOut += m.Amount
			}
		}
		balanceLbl.SetText(fmt.Sprintf("Solde : %s DA", utils.FormatAmount(currentBalance)))
		totalInLbl.SetText(fmt.Sprintf("Entrées : %s DA", utils.FormatAmount(totalIn)))
		totalOutLbl.SetText(fmt.Sprintf("Sorties : %s DA", utils.FormatAmount(totalOut)))

		// Construire le tableau
		headerRow := buildTableHeaderRow(colHeaders, colWidths)
		rows := container.NewVBox(headerRow)

		for _, m := range movements {
			entree, sortie := "", ""
			typeLabel, bgColor := "➕ Entrée", "#e8f5e9"
			if m.Type == "in" {
				entree = utils.FormatAmount(m.Amount)
			} else {
				sortie = utils.FormatAmount(m.Amount)
				typeLabel, bgColor = "➖ Sortie", "#fce4ec"
			}
			_ = bgColor

			cells := []string{
				utils.FormatDateFr(m.Date),
				typeLabel,
				m.Category,
				m.PartyName,
				utils.TruncateString(m.Description, 30),
				m.Reference,
				entree,
				sortie,
				utils.FormatAmount(m.Balance),
			}
			rows.Add(buildTableRow(cells, colWidths, false))
		}

		if len(movements) == 0 {
			rows.Add(widget.NewLabel("   Aucun mouvement pour la période sélectionnée."))
		}
		tableContainer.Objects = []fyne.CanvasObject{container.NewVScroll(rows)}
		tableContainer.Refresh()
	}

	periodFilter := buildPeriodFilter(fromEntry, toEntry, loadMovements)

	// ── Bouton Nouveau Mouvement
	newMvBtn := widget.NewButtonWithIcon("Nouveau Mouvement", theme.ContentAddIcon(), func() {
		showCashMovementDialog(db, loadMovements)
	})
	newMvBtn.Importance = widget.HighImportance

	// KPI bar
	kpiBar := container.NewHBox(
		balanceLbl, widget.NewSeparator(),
		totalInLbl, widget.NewSeparator(),
		totalOutLbl,
	)

	toolbar := container.NewBorder(nil, nil, newMvBtn, nil, kpiBar)

	loadMovements()

	return container.NewBorder(
		container.NewVBox(header, periodFilter, toolbar, widget.NewSeparator()),
		nil, nil, nil,
		tableContainer,
	)
}

// showCashMovementDialog affiche le dialogue de saisie d'un mouvement de caisse
func showCashMovementDialog(_ interface{}, onSave func()) {
	appDB := appstate.GetDB()
	if appDB == nil {
		return
	}

	typeSelect := widget.NewSelect([]string{"Entrée (in)", "Sortie (out)"}, nil)
	typeSelect.SetSelected("Entrée (in)")
	dateEntry := widget.NewEntry()
	dateEntry.SetText(utils.TodayString())
	categoryEntry := widget.NewEntry()
	categoryEntry.SetPlaceHolder("Catégorie (ex: Vente, Dépense...)")
	descEntry := widget.NewEntry()
	descEntry.SetPlaceHolder("Description du mouvement")
	refEntry := widget.NewEntry()
	refEntry.SetPlaceHolder("Référence (numéro pièce)")
	partyEntry := widget.NewEntry()
	partyEntry.SetPlaceHolder("Nom tiers (client/fournisseur)")
	amountEntry := widget.NewEntry()
	amountEntry.SetPlaceHolder("0,00")

	content := container.NewVBox(
		widget.NewForm(
			widget.NewFormItem("Type *", typeSelect),
			widget.NewFormItem("Date *", dateEntry),
			widget.NewFormItem("Catégorie", categoryEntry),
			widget.NewFormItem("Description", descEntry),
			widget.NewFormItem("Référence", refEntry),
			widget.NewFormItem("Tiers", partyEntry),
			widget.NewFormItem("Montant (DA) *", amountEntry),
		),
	)

	win := appstate.MainWindow
	dlg := dialog.NewCustomConfirm("Nouveau Mouvement de Caisse", "Enregistrer", "Annuler",
		container.NewScroll(content),
		func(ok bool) {
			if !ok {
				return
			}
			amount, err := strconv.ParseFloat(strings.ReplaceAll(amountEntry.Text, ",", "."), 64)
			if err != nil || amount <= 0 {
				dialog.ShowError(fmt.Errorf("montant invalide"), win)
				return
			}
			mvType := "in"
			if typeSelect.Selected == "Sortie (out)" {
				mvType = "out"
			}

			session := appstate.GetSession()
			userID := 0
			if session != nil {
				userID = session.UserID
			}

			move := &models.CashMovement{
				Date:        dateEntry.Text,
				Type:        mvType,
				Category:    categoryEntry.Text,
				Description: descEntry.Text,
				Reference:   refEntry.Text,
				PartyName:   partyEntry.Text,
				Amount:      amount,
			}
			moveCopy := *move
			go func(mv models.CashMovement) {
				svc := services.NewCashService(appDB)
				err := svc.AddCashMovement(&mv, userID)
				fyne.Do(func() {
					if err != nil {
						dialog.ShowError(err, win)
						return
					}
					dialog.ShowInformation("Succès", "Mouvement de caisse enregistré.", win)
					onSave()
				})
			}(moveCopy)
		}, win)
	dlg.Resize(fyne.NewSize(500, 400))
	dlg.Show()
}

// ─────────────────────────────────────────────────────────────────────────────
// 2. BANQUE
// ─────────────────────────────────────────────────────────────────────────────

// BuildBankScreen construit l'écran de gestion bancaire
func BuildBankScreen() fyne.CanvasObject {
	db := appstate.GetDB()

	header := buildTreasuryHeader("🏛️", "Banque",
		"Gestion des comptes bancaires et mouvements", "#2980b9")

	// ── Liste des comptes bancaires
	var accounts []models.BankAccount
	var selectedAccount *models.BankAccount

	accountList := widget.NewList(
		func() int { return len(accounts) },
		func() fyne.CanvasObject {
			return container.NewHBox(
				widget.NewLabel("Banque"),
				widget.NewLabel("Solde"),
			)
		},
		func(i widget.ListItemID, o fyne.CanvasObject) {
			if i >= len(accounts) {
				return
			}
			a := accounts[i]
			row := o.(*fyne.Container)
			row.Objects[0].(*widget.Label).SetText(fmt.Sprintf("%s — %s", a.BankName, a.AccountNumber))
			row.Objects[1].(*widget.Label).SetText(utils.FormatAmount(a.Balance) + " DA")
		},
	)

	// ── Panel détail compte + mouvements
	detailContainer := container.NewStack()
	fromEntry := widget.NewEntry()
	toEntry := widget.NewEntry()

	var bankMovements []models.BankMovement

	var loadBankMovements func()
	loadBankMovements = func() {
		if db == nil || selectedAccount == nil {
			return
		}
		svc := services.NewCashService(db)
		mvs, err := svc.GetBankMovements(selectedAccount.ID, fromEntry.Text, toEntry.Text)
		if err == nil {
			bankMovements = mvs
		}

		colHeaders := []string{"Date", "Type", "Description", "Référence", "Débit (DA)", "Crédit (DA)", "Rapprochement"}
		colWidths := []float32{90, 100, 200, 120, 120, 120, 120}

		headerRow := buildTableHeaderRow(colHeaders, colWidths)
		rows := container.NewVBox(headerRow)

		var totalDebit, totalCredit float64
		for _, m := range bankMovements {
			totalDebit += m.Debit
			totalCredit += m.Credit
			recon := "Non"
			if m.IsReconciled {
				recon = "✓ Oui"
			}
			cells := []string{
				utils.FormatDateFr(m.Date),
				m.Type,
				utils.TruncateString(m.Description, 30),
				m.Reference,
				utils.FormatAmount(m.Debit),
				utils.FormatAmount(m.Credit),
				recon,
			}
			rows.Add(buildTableRow(cells, colWidths, false))
		}
		// Ligne totaux
		rows.Add(widget.NewSeparator())
		rows.Add(buildTableRow(
			[]string{"", "", "", "TOTAL", utils.FormatAmount(totalDebit), utils.FormatAmount(totalCredit), ""},
			colWidths, true,
		))

		if len(bankMovements) == 0 {
			rows.Add(widget.NewLabel("   Aucun mouvement pour la période."))
		}

		// KPI solde
		soldeLbl := widget.NewRichTextFromMarkdown(fmt.Sprintf(
			"**Solde compte :** %s DA", utils.FormatAmount(selectedAccount.Balance)))

		addMvBtn := widget.NewButtonWithIcon("Ajouter Mouvement", theme.ContentAddIcon(), func() {
			showBankMovementDialog(db, selectedAccount, loadBankMovements)
		})
		addMvBtn.Importance = widget.HighImportance

		periodFilter := buildPeriodFilter(fromEntry, toEntry, loadBankMovements)

		detailContainer.Objects = []fyne.CanvasObject{
			container.NewBorder(
				container.NewVBox(
					widget.NewRichTextFromMarkdown("### "+selectedAccount.BankName+" — "+selectedAccount.AccountNumber),
					container.NewHBox(soldeLbl, widget.NewSeparator(), addMvBtn),
					periodFilter,
					widget.NewSeparator(),
				),
				nil, nil, nil,
				container.NewVScroll(rows),
			),
		}
		detailContainer.Refresh()
	}

	accountList.OnSelected = func(id widget.ListItemID) {
		if id < len(accounts) {
			selectedAccount = &accounts[id]
			loadBankMovements()
		}
	}

	loadAccounts := func() {
		if db == nil {
			return
		}
		svc := services.NewCashService(db)
		accs, err := svc.GetAllBankAccounts()
		if err == nil {
			accounts = accs
		}
		accountList.Refresh()
	}

	// ── Bouton Nouveau Compte
	newAccBtn := widget.NewButtonWithIcon("Nouveau Compte", theme.ContentAddIcon(), func() {
		showBankAccountDialog(db, loadAccounts)
	})
	newAccBtn.Importance = widget.HighImportance

	accountPanel := container.NewBorder(
		container.NewVBox(
			widget.NewRichTextFromMarkdown("### Comptes Bancaires"),
			newAccBtn,
			widget.NewSeparator(),
		),
		nil, nil, nil,
		accountList,
	)

	loadAccounts()

	split := container.NewHSplit(
		container.NewPadded(accountPanel),
		container.NewPadded(detailContainer),
	)
	split.Offset = 0.28

	return container.NewBorder(header, nil, nil, nil, split)
}

func showBankAccountDialog(_ interface{}, onSave func()) {
	appDB := appstate.GetDB()
	if appDB == nil {
		return
	}
	win := appstate.MainWindow

	bankEntry := widget.NewEntry()
	bankEntry.SetPlaceHolder("Nom de la banque (ex: BEA, BNA...)")
	branchEntry := widget.NewEntry()
	branchEntry.SetPlaceHolder("Agence")
	accountNumEntry := widget.NewEntry()
	accountNumEntry.SetPlaceHolder("N° de compte")
	ribEntry := widget.NewEntry()
	ribEntry.SetPlaceHolder("RIB / IBAN")
	balanceEntry := widget.NewEntry()
	balanceEntry.SetText("0")

	content := widget.NewForm(
		widget.NewFormItem("Banque *", bankEntry),
		widget.NewFormItem("Agence", branchEntry),
		widget.NewFormItem("N° Compte *", accountNumEntry),
		widget.NewFormItem("RIB", ribEntry),
		widget.NewFormItem("Solde initial (DA)", balanceEntry),
	)

	dlg := dialog.NewCustomConfirm("Nouveau Compte Bancaire", "Créer", "Annuler",
		container.NewScroll(content),
		func(ok bool) {
			if !ok {
				return
			}
			if strings.TrimSpace(bankEntry.Text) == "" || strings.TrimSpace(accountNumEntry.Text) == "" {
				dialog.ShowError(fmt.Errorf("banque et numéro de compte requis"), win)
				return
			}
			bal, _ := strconv.ParseFloat(strings.ReplaceAll(balanceEntry.Text, ",", "."), 64)
			acc := &models.BankAccount{
				BankName:      bankEntry.Text,
				Branch:        branchEntry.Text,
				AccountNumber: accountNumEntry.Text,
				RIB:           ribEntry.Text,
				Balance:       bal,
			}
			accCopy := *acc
			go func(a models.BankAccount) {
				svc := services.NewCashService(appDB)
				err := svc.SaveBankAccount(&a)
				fyne.Do(func() {
					if err != nil {
						dialog.ShowError(err, win)
						return
					}
					dialog.ShowInformation("Succès", "Compte bancaire créé.", win)
					onSave()
				})
			}(accCopy)
		}, win)
	dlg.Resize(fyne.NewSize(480, 360))
	dlg.Show()
}

func showBankMovementDialog(_ interface{}, acc *models.BankAccount, onSave func()) {
	appDB := appstate.GetDB()
	if appDB == nil || acc == nil {
		return
	}
	win := appstate.MainWindow

	typeSelect := widget.NewSelect([]string{"Crédit (dépôt)", "Débit (retrait)", "Virement émis", "Virement reçu", "Frais bancaires"}, nil)
	typeSelect.SetSelected("Crédit (dépôt)")
	dateEntry := widget.NewEntry()
	dateEntry.SetText(utils.TodayString())
	descEntry := widget.NewEntry()
	descEntry.SetPlaceHolder("Description")
	refEntry := widget.NewEntry()
	refEntry.SetPlaceHolder("N° opération")
	debitEntry := widget.NewEntry()
	debitEntry.SetText("0")
	creditEntry := widget.NewEntry()
	creditEntry.SetText("0")

	content := widget.NewForm(
		widget.NewFormItem("Type *", typeSelect),
		widget.NewFormItem("Date *", dateEntry),
		widget.NewFormItem("Description", descEntry),
		widget.NewFormItem("Référence", refEntry),
		widget.NewFormItem("Débit (DA)", debitEntry),
		widget.NewFormItem("Crédit (DA)", creditEntry),
	)

	dlg := dialog.NewCustomConfirm("Nouveau Mouvement Bancaire — "+acc.BankName, "Enregistrer", "Annuler",
		container.NewScroll(content),
		func(ok bool) {
			if !ok {
				return
			}
			debit, _ := strconv.ParseFloat(strings.ReplaceAll(debitEntry.Text, ",", "."), 64)
			credit, _ := strconv.ParseFloat(strings.ReplaceAll(creditEntry.Text, ",", "."), 64)
			if debit == 0 && credit == 0 {
				dialog.ShowError(fmt.Errorf("débit ou crédit doit être > 0"), win)
				return
			}
			mv := &models.BankMovement{
				BankAccountID: acc.ID,
				Date:          dateEntry.Text,
				Type:          typeSelect.Selected,
				Description:   descEntry.Text,
				Reference:     refEntry.Text,
				Debit:         debit,
				Credit:        credit,
			}
			mvCopy := *mv
			go func(m models.BankMovement) {
				svc := services.NewCashService(appDB)
				err := svc.AddBankMovement(&m)
				fyne.Do(func() {
					if err != nil {
						dialog.ShowError(err, win)
						return
					}
					dialog.ShowInformation("Succès", "Mouvement bancaire enregistré.", win)
					onSave()
				})
			}(mvCopy)
		}, win)
	dlg.Resize(fyne.NewSize(480, 380))
	dlg.Show()
}

// ─────────────────────────────────────────────────────────────────────────────
// 3. CHÈQUES
// ─────────────────────────────────────────────────────────────────────────────

// BuildChequesScreen construit l'écran de gestion des chèques
func BuildChequesScreen() fyne.CanvasObject {
	db := appstate.GetDB()

	header := buildTreasuryHeader("📋", "Gestion des Chèques",
		"Chèques reçus (clients) et émis (fournisseurs)", "#8e44ad")

	// ── Filtres
	typeSelect := widget.NewSelect([]string{"Tous", "Reçus (received)", "Émis (issued)"}, nil)
	typeSelect.SetSelected("Tous")
	statusSelect := widget.NewSelect([]string{"Tous", "pending", "deposited", "cleared", "rejected"}, nil)
	statusSelect.SetSelected("Tous")

	// ── Table chèques
	colHeaders := []string{"N° Chèque", "Type", "Date", "Échéance", "Tiers", "Banque", "Montant (DA)", "Statut", "Actions"}
	colWidths := []float32{110, 80, 90, 90, 140, 110, 120, 90, 130}

	tableContainer := container.NewStack()
	var cheques []models.Cheque

	var loadCheques func()
	loadCheques = func() {
		if db == nil {
			return
		}
		chqType := ""
		if typeSelect.Selected == "Reçus (received)" {
			chqType = "received"
		} else if typeSelect.Selected == "Émis (issued)" {
			chqType = "issued"
		}
		status := ""
		if statusSelect.Selected != "Tous" {
			status = statusSelect.Selected
		}
		svc := services.NewCashService(db)
		chqs, err := svc.GetCheques(chqType, status)
		if err == nil {
			cheques = chqs
		}

		headerRow := buildTableHeaderRow(colHeaders, colWidths)
		rows := container.NewVBox(headerRow)

		for i, c := range cheques {
			idx := i
			chq := c

			statusLabel := chequeSatutsLabel(chq.Status)
			typeLabel := "📥 Reçu"
			if chq.Type == "issued" {
				typeLabel = "📤 Émis"
			}

			// Boutons d'action
			var actionBtn *widget.Button
			if chq.Status == "pending" || chq.Status == "deposited" {
				actionBtn = widget.NewButton("Changer statut", func() {
					showChequeStatusDialog(db, &cheques[idx], loadCheques)
				})
			} else {
				actionBtn = widget.NewButton("Détails", func() {
					showChequeDetails(&cheques[idx])
				})
			}

			cells := []string{
				chq.ChequeNumber,
				typeLabel,
				utils.FormatDateFr(chq.Date),
				utils.FormatDateFr(chq.DueDate),
				chq.PayerPayee,
				chq.BankName,
				utils.FormatAmount(chq.Amount),
				statusLabel,
				"",
			}
			row := container.NewHBox()
			for j, cell := range cells[:8] {
				lbl := widget.NewLabel(cell)
				lbl.Wrapping = fyne.TextWrapOff
				lbl.Resize(fyne.NewSize(colWidths[j], 28))
				row.Add(lbl)
			}
			row.Add(actionBtn)
			rows.Add(row)
		}

		if len(cheques) == 0 {
			rows.Add(widget.NewLabel("   Aucun chèque trouvé."))
		}

		tableContainer.Objects = []fyne.CanvasObject{container.NewVScroll(rows)}
		tableContainer.Refresh()
	}

	// Alertes échéances proches
	var alertsBanner fyne.CanvasObject
	if db != nil {
		svc := services.NewCashService(db)
		nearDue, _ := svc.GetChequesNearDue()
		if len(nearDue) > 0 {
			alertsBanner = widget.NewRichTextFromMarkdown(fmt.Sprintf(
				"> ⚠️ **%d chèque(s) arrivent à échéance dans les 7 prochains jours !**", len(nearDue)))
		}
	}

	// ── Bouton Nouveau Chèque
	newChqBtn := widget.NewButtonWithIcon("Nouveau Chèque", theme.ContentAddIcon(), func() {
		showNewChequeDialog(db, loadCheques)
	})
	newChqBtn.Importance = widget.HighImportance

	filterRow := container.NewHBox(
		widget.NewLabel("Type:"), typeSelect,
		widget.NewLabel("Statut:"), statusSelect,
		widget.NewButtonWithIcon("Filtrer", theme.SearchIcon(), loadCheques),
		widget.NewSeparator(),
		newChqBtn,
	)

	headerSection := container.NewVBox(header, filterRow)
	if alertsBanner != nil {
		headerSection.Add(alertsBanner)
	}
	headerSection.Add(widget.NewSeparator())

	loadCheques()

	return container.NewBorder(headerSection, nil, nil, nil, tableContainer)
}

func chequeSatutsLabel(status string) string {
	switch status {
	case "pending":
		return "⏳ En attente"
	case "deposited":
		return "🏦 Déposé"
	case "cleared":
		return "✅ Encaissé"
	case "rejected":
		return "❌ Rejeté"
	default:
		return status
	}
}

func showChequeDetails(c *models.Cheque) {
	win := appstate.MainWindow
	info := fmt.Sprintf(
		"N° : %s\nType : %s\nDate : %s\nÉchéance : %s\nTiers : %s\nBanque : %s\nMontant : %s DA\nStatut : %s\nNote : %s",
		c.ChequeNumber, c.Type, c.Date, c.DueDate, c.PayerPayee,
		c.BankName, utils.FormatAmount(c.Amount), c.Status, c.Notes,
	)
	dialog.ShowInformation("Détails Chèque — "+c.ChequeNumber, info, win)
}

func showChequeStatusDialog(db interface{}, c *models.Cheque, onSave func()) {
	appDB := appstate.GetDB()
	win := appstate.MainWindow

	statusSelect := widget.NewSelect([]string{"deposited", "cleared", "rejected"}, nil)
	statusSelect.SetSelected("cleared")
	reasonEntry := widget.NewEntry()
	reasonEntry.SetPlaceHolder("Motif de rejet (si rejeté)")

	bankAccounts := []models.BankAccount{}
	var bankIDs []int
	bankOpts := []string{"Aucun"}
	if appDB != nil {
		svc := services.NewCashService(appDB)
		accs, _ := svc.GetAllBankAccounts()
		bankAccounts = accs
	}
	for _, a := range bankAccounts {
		bankOpts = append(bankOpts, a.BankName+" — "+a.AccountNumber)
		bankIDs = append(bankIDs, a.ID)
	}
	bankSelect := widget.NewSelect(bankOpts, nil)
	bankSelect.SetSelected("Aucun")

	content := widget.NewForm(
		widget.NewFormItem("Nouveau statut", statusSelect),
		widget.NewFormItem("Motif rejet", reasonEntry),
		widget.NewFormItem("Compte bancaire (si dépôt)", bankSelect),
	)

	dlg := dialog.NewCustomConfirm("Changer statut — Chèque "+c.ChequeNumber, "Valider", "Annuler",
		content,
		func(ok bool) {
			if !ok || appDB == nil {
				return
			}
			bankID := 0
			idx := bankSelect.SelectedIndex()
			if idx > 0 && idx-1 < len(bankIDs) {
				bankID = bankIDs[idx-1]
			}
			cID, newStatus, reason := c.ID, statusSelect.Selected, reasonEntry.Text
			go func() {
				svc := services.NewCashService(appDB)
				err := svc.UpdateChequeStatus(cID, newStatus, reason, bankID)
				fyne.Do(func() {
					if err != nil {
						dialog.ShowError(err, win)
						return
					}
					dialog.ShowInformation("Succès", "Statut chèque mis à jour.", win)
					onSave()
				})
			}()
		}, win)
	dlg.Resize(fyne.NewSize(440, 320))
	dlg.Show()
}

func showNewChequeDialog(db interface{}, onSave func()) {
	appDB := appstate.GetDB()
	win := appstate.MainWindow

	typeSelect := widget.NewSelect([]string{"received", "issued"}, nil)
	typeSelect.SetSelected("received")
	numEntry := widget.NewEntry()
	numEntry.SetPlaceHolder("N° de chèque")
	dateEntry := widget.NewEntry()
	dateEntry.SetText(utils.TodayString())
	dueDateEntry := widget.NewEntry()
	dueDateEntry.SetText(utils.TodayString())
	amountEntry := widget.NewEntry()
	amountEntry.SetPlaceHolder("0,00")
	payerEntry := widget.NewEntry()
	payerEntry.SetPlaceHolder("Nom donneur / bénéficiaire")
	bankEntry := widget.NewEntry()
	bankEntry.SetPlaceHolder("Banque tirée")
	notesEntry := widget.NewMultiLineEntry()
	notesEntry.SetPlaceHolder("Notes...")

	content := container.NewVBox(widget.NewForm(
		widget.NewFormItem("Type *", typeSelect),
		widget.NewFormItem("N° Chèque *", numEntry),
		widget.NewFormItem("Date *", dateEntry),
		widget.NewFormItem("Échéance *", dueDateEntry),
		widget.NewFormItem("Montant (DA) *", amountEntry),
		widget.NewFormItem("Tiers *", payerEntry),
		widget.NewFormItem("Banque", bankEntry),
		widget.NewFormItem("Notes", notesEntry),
	))

	dlg := dialog.NewCustomConfirm("Nouveau Chèque", "Enregistrer", "Annuler",
		container.NewScroll(content),
		func(ok bool) {
			if !ok || appDB == nil {
				return
			}
			amount, err := strconv.ParseFloat(strings.ReplaceAll(amountEntry.Text, ",", "."), 64)
			if err != nil || amount <= 0 || numEntry.Text == "" || payerEntry.Text == "" {
				dialog.ShowError(fmt.Errorf("champs obligatoires manquants ou montant invalide"), win)
				return
			}
			chq := &models.Cheque{
				Type:         typeSelect.Selected,
				ChequeNumber: numEntry.Text,
				Date:         dateEntry.Text,
				DueDate:      dueDateEntry.Text,
				Amount:       amount,
				PayerPayee:   payerEntry.Text,
				BankName:     bankEntry.Text,
				Status:       "pending",
				Notes:        notesEntry.Text,
			}
			chqCopy := *chq
			go func(ch models.Cheque) {
				svc := services.NewCashService(appDB)
				err := svc.SaveCheque(&ch)
				fyne.Do(func() {
					if err != nil {
						dialog.ShowError(err, win)
						return
					}
					dialog.ShowInformation("Succès", "Chèque enregistré.", win)
					onSave()
				})
			}(chqCopy)
		}, win)
	dlg.Resize(fyne.NewSize(500, 480))
	dlg.Show()
}

// ─────────────────────────────────────────────────────────────────────────────
// 4. ENCAISSEMENTS CLIENTS
// ─────────────────────────────────────────────────────────────────────────────

// BuildCollectionsScreen construit l'écran des encaissements clients
func BuildCollectionsScreen() fyne.CanvasObject {
	db := appstate.GetDB()

	header := buildTreasuryHeader("💰", "Encaissements Clients",
		"Saisir et suivre les règlements clients", "#27ae60")

	// ── Formulaire encaissement
	var clients []models.Client
	clientNames := []string{"— Sélectionner client —"}
	var clientIDs []int

	if db != nil {
		rows, err := db.Query(`SELECT id, name_fr, balance FROM clients ORDER BY name_fr`)
		if err == nil {
			defer rows.Close()
			for rows.Next() {
				var c models.Client
				rows.Scan(&c.ID, &c.NameFr, &c.Balance)
				clients = append(clients, c)
				clientNames = append(clientNames, fmt.Sprintf("%s (solde: %s DA)", c.NameFr, utils.FormatAmount(c.Balance)))
				clientIDs = append(clientIDs, c.ID)
			}
		}
	}

	clientSelect := widget.NewSelect(clientNames, nil)
	clientSelect.SetSelected(clientNames[0])
	dateEntry := widget.NewEntry()
	dateEntry.SetText(utils.TodayString())
	amountEntry := widget.NewEntry()
	amountEntry.SetPlaceHolder("Montant à encaisser (DA)")
	methodSelect := widget.NewSelect([]string{"cash", "cheque", "virement", "mixte"}, nil)
	methodSelect.SetSelected("cash")
	chequeNumEntry := widget.NewEntry()
	chequeNumEntry.SetPlaceHolder("N° chèque (si paiement chèque)")
	bankEntry := widget.NewEntry()
	bankEntry.SetPlaceHolder("Banque (si chèque/virement)")
	refEntry := widget.NewEntry()
	refEntry.SetPlaceHolder("Référence interne")
	notesEntry := widget.NewEntry()
	notesEntry.SetPlaceHolder("Notes")

	// Solde client sélectionné
	clientSoldeLbl := widget.NewLabel("")
	clientSelect.OnChanged = func(s string) {
		idx := clientSelect.SelectedIndex()
		if idx > 0 && idx-1 < len(clients) {
			c := clients[idx-1]
			clientSoldeLbl.SetText(fmt.Sprintf("Solde actuel : %s DA", utils.FormatAmount(c.Balance)))
		} else {
			clientSoldeLbl.SetText("")
		}
	}

	win := appstate.MainWindow
	payBtn := widget.NewButtonWithIcon("Enregistrer Encaissement", theme.ConfirmIcon(), func() {
		if db == nil {
			return
		}
		idx := clientSelect.SelectedIndex()
		if idx <= 0 {
			dialog.ShowError(fmt.Errorf("sélectionner un client"), win)
			return
		}
		clientID := clientIDs[idx-1]
		amount, err := strconv.ParseFloat(strings.ReplaceAll(amountEntry.Text, ",", "."), 64)
		if err != nil || amount <= 0 {
			dialog.ShowError(fmt.Errorf("montant invalide"), win)
			return
		}
		session := appstate.GetSession()
		userID := 0
		if session != nil {
			userID = session.UserID
		}
		payment := &models.Payment{
			Type:          "collection",
			Date:          dateEntry.Text,
			ClientID:      &clientID,
			Amount:        amount,
			PaymentMethod: methodSelect.Selected,
			ChequeNumber:  chequeNumEntry.Text,
			BankName:      bankEntry.Text,
			Reference:     refEntry.Text,
			Notes:         notesEntry.Text,
		}
		pmtCopy := *payment
		amtStr := utils.FormatAmount(amount)
		go func(p models.Payment) {
			svc := services.NewPaymentService(db)
			err := svc.ProcessClientPayment(&p, userID)
			fyne.Do(func() {
				if err != nil {
					dialog.ShowError(err, win)
					return
				}
				dialog.ShowInformation("Encaissement enregistré",
					fmt.Sprintf("Paiement de %s DA enregistré pour ce client.\nVentilation FIFO automatique sur les factures impayées.", amtStr), win)
				amountEntry.SetText("")
				chequeNumEntry.SetText("")
				bankEntry.SetText("")
				notesEntry.SetText("")
				clientSelect.OnChanged(clientSelect.Selected)
			})
		}(pmtCopy)
	})
	payBtn.Importance = widget.HighImportance

	form := widget.NewForm(
		widget.NewFormItem("Client *", clientSelect),
		widget.NewFormItem("", clientSoldeLbl),
		widget.NewFormItem("Date *", dateEntry),
		widget.NewFormItem("Montant (DA) *", amountEntry),
		widget.NewFormItem("Mode de paiement *", methodSelect),
		widget.NewFormItem("N° Chèque", chequeNumEntry),
		widget.NewFormItem("Banque", bankEntry),
		widget.NewFormItem("Référence", refEntry),
		widget.NewFormItem("Notes", notesEntry),
	)

	infoCard := widget.NewRichTextFromMarkdown(`
**Fonctionnement :**
- Le paiement est ventilé automatiquement (FIFO) sur les factures impayées les plus anciennes
- Si paiement en espèces → mouvement de caisse créé automatiquement
- Si paiement par chèque → chèque reçu créé avec statut "En attente"
- Le solde client est mis à jour immédiatement
`)

	leftPane := container.NewVBox(header, form, payBtn)
	rightPane := container.NewPadded(infoCard)

	return container.NewHSplit(
		container.NewPadded(leftPane),
		rightPane,
	)
}

// ─────────────────────────────────────────────────────────────────────────────
// 5. DÉCAISSEMENTS FOURNISSEURS
// ─────────────────────────────────────────────────────────────────────────────

// BuildDisbursementsScreen construit l'écran des décaissements fournisseurs
func BuildDisbursementsScreen() fyne.CanvasObject {
	db := appstate.GetDB()

	header := buildTreasuryHeader("💸", "Décaissements Fournisseurs",
		"Saisir les règlements fournisseurs", "#e67e22")

	var suppliers []models.Supplier
	supplierNames := []string{"— Sélectionner fournisseur —"}
	var supplierIDs []int

	if db != nil {
		rows, err := db.Query(`SELECT id, name_fr, balance FROM suppliers ORDER BY name_fr`)
		if err == nil {
			defer rows.Close()
			for rows.Next() {
				var s models.Supplier
				rows.Scan(&s.ID, &s.NameFr, &s.Balance)
				suppliers = append(suppliers, s)
				supplierNames = append(supplierNames, fmt.Sprintf("%s (dû: %s DA)", s.NameFr, utils.FormatAmount(s.Balance)))
				supplierIDs = append(supplierIDs, s.ID)
			}
		}
	}

	supplierSelect := widget.NewSelect(supplierNames, nil)
	supplierSelect.SetSelected(supplierNames[0])
	dateEntry := widget.NewEntry()
	dateEntry.SetText(utils.TodayString())
	amountEntry := widget.NewEntry()
	amountEntry.SetPlaceHolder("Montant à payer (DA)")
	methodSelect := widget.NewSelect([]string{"cash", "cheque", "virement"}, nil)
	methodSelect.SetSelected("virement")
	chequeNumEntry := widget.NewEntry()
	chequeNumEntry.SetPlaceHolder("N° chèque")
	bankEntry := widget.NewEntry()
	bankEntry.SetPlaceHolder("Banque")
	refEntry := widget.NewEntry()
	refEntry.SetPlaceHolder("Référence")
	notesEntry := widget.NewEntry()
	notesEntry.SetPlaceHolder("Notes")

	supplierSoldeLbl := widget.NewLabel("")
	supplierSelect.OnChanged = func(s string) {
		idx := supplierSelect.SelectedIndex()
		if idx > 0 && idx-1 < len(suppliers) {
			sup := suppliers[idx-1]
			supplierSoldeLbl.SetText(fmt.Sprintf("Montant dû : %s DA", utils.FormatAmount(sup.Balance)))
		} else {
			supplierSoldeLbl.SetText("")
		}
	}

	win := appstate.MainWindow
	payBtn := widget.NewButtonWithIcon("Enregistrer Décaissement", theme.ConfirmIcon(), func() {
		if db == nil {
			return
		}
		idx := supplierSelect.SelectedIndex()
		if idx <= 0 {
			dialog.ShowError(fmt.Errorf("sélectionner un fournisseur"), win)
			return
		}
		supplierID := supplierIDs[idx-1]
		amount, err := strconv.ParseFloat(strings.ReplaceAll(amountEntry.Text, ",", "."), 64)
		if err != nil || amount <= 0 {
			dialog.ShowError(fmt.Errorf("montant invalide"), win)
			return
		}
		session := appstate.GetSession()
		userID := 0
		if session != nil {
			userID = session.UserID
		}
		payment := &models.Payment{
			Type:          "disbursement",
			Date:          dateEntry.Text,
			SupplierID:    &supplierID,
			Amount:        amount,
			PaymentMethod: methodSelect.Selected,
			ChequeNumber:  chequeNumEntry.Text,
			BankName:      bankEntry.Text,
			Reference:     refEntry.Text,
			Notes:         notesEntry.Text,
		}
		pmtCopy2 := *payment
		amtStr2 := utils.FormatAmount(amount)
		go func(p models.Payment) {
			svc := services.NewPaymentService(db)
			err := svc.ProcessSupplierPayment(&p, userID)
			fyne.Do(func() {
				if err != nil {
					dialog.ShowError(err, win)
					return
				}
				dialog.ShowInformation("Décaissement enregistré",
					fmt.Sprintf("Paiement de %s DA enregistré.", amtStr2), win)
				amountEntry.SetText("")
				supplierSelect.OnChanged(supplierSelect.Selected)
			})
		}(pmtCopy2)
	})
	payBtn.Importance = widget.HighImportance

	form := widget.NewForm(
		widget.NewFormItem("Fournisseur *", supplierSelect),
		widget.NewFormItem("", supplierSoldeLbl),
		widget.NewFormItem("Date *", dateEntry),
		widget.NewFormItem("Montant (DA) *", amountEntry),
		widget.NewFormItem("Mode de paiement *", methodSelect),
		widget.NewFormItem("N° Chèque", chequeNumEntry),
		widget.NewFormItem("Banque", bankEntry),
		widget.NewFormItem("Référence", refEntry),
		widget.NewFormItem("Notes", notesEntry),
	)

	return container.NewBorder(
		header, nil, nil, nil,
		container.NewPadded(container.NewVBox(form, payBtn)),
	)
}

// ─────────────────────────────────────────────────────────────────────────────
// 6. BALANCE ÂGÉE
// ─────────────────────────────────────────────────────────────────────────────

// BuildAgingScreen construit l'écran balance âgée
func BuildAgingScreen() fyne.CanvasObject {
	db := appstate.GetDB()

	header := buildTreasuryHeader("📊", "Balance Âgée",
		"Créances clients et dettes fournisseurs par ancienneté", "#c0392b")

	// ── Onglets clients / fournisseurs
	clientTab := buildAgingTable(db, "clients")
	supplierTab := buildAgingTable(db, "suppliers")

	tabs := container.NewAppTabs(
		container.NewTabItem("👥 Clients — Créances", clientTab),
		container.NewTabItem("🏭 Fournisseurs — Dettes", supplierTab),
	)

	return container.NewBorder(header, nil, nil, nil, tabs)
}

func buildAgingTable(_ *sql.DB, party string) fyne.CanvasObject {
	appDB := appstate.GetDB()
	if appDB == nil {
		return widget.NewLabel("Base de données non connectée.")
	}

	colHeaders := []string{"Tiers", "Solde Total (DA)", "0-30 j", "31-60 j", "61-90 j", "> 90 j"}
	colWidths := []float32{220, 140, 130, 130, 130, 130}

	svc := services.NewCashService(appDB)
	var lines []models.AgingLine
	if party == "clients" {
		lines, _ = svc.GetClientAging()
	} else {
		lines, _ = svc.GetSupplierAging()
	}

	headerRow := buildTableHeaderRow(colHeaders, colWidths)
	rows := container.NewVBox(headerRow)

	var totalBalance, total030, total3160, total6190, total90 float64
	for _, l := range lines {
		cells := []string{
			l.ClientName,
			utils.FormatAmount(l.Balance),
			utils.FormatAmount(l.Age0_30),
			utils.FormatAmount(l.Age31_60),
			utils.FormatAmount(l.Age61_90),
			utils.FormatAmount(l.Age90Plus),
		}
		rows.Add(buildTableRow(cells, colWidths, false))
		totalBalance += l.Balance
		total030 += l.Age0_30
		total3160 += l.Age31_60
		total6190 += l.Age61_90
		total90 += l.Age90Plus
	}

	if len(lines) == 0 {
		rows.Add(widget.NewLabel("   Aucune créance en cours."))
	} else {
		rows.Add(widget.NewSeparator())
		rows.Add(buildTableRow(
			[]string{"TOTAL", utils.FormatAmount(totalBalance), utils.FormatAmount(total030),
				utils.FormatAmount(total3160), utils.FormatAmount(total6190), utils.FormatAmount(total90)},
			colWidths, true,
		))
	}

	return container.NewVScroll(rows)
}

// ─────────────────────────────────────────────────────────────────────────────
// 7. DÉPENSES DIVERSES
// ─────────────────────────────────────────────────────────────────────────────

// BuildExpensesScreen construit l'écran des dépenses diverses
func BuildExpensesScreen() fyne.CanvasObject {
	db := appstate.GetDB()

	header := buildTreasuryHeader("🧾", "Dépenses Diverses",
		"Charges et frais d'exploitation", "#e74c3c")

	fromEntry := widget.NewEntry()
	toEntry := widget.NewEntry()

	tableContainer := container.NewStack()
	var expenses []models.Expense

	var expenseCategories []models.ExpenseCategory
	catNames := []string{"Toutes"}
	catIDs := []int{0}
	if db != nil {
		svc := services.NewCashService(db)
		cats, _ := svc.GetExpenseCategories()
		expenseCategories = cats
		for _, c := range expenseCategories {
			catNames = append(catNames, c.Name)
			catIDs = append(catIDs, c.ID)
		}
	}

	loadExpenses := func() {
		if db == nil {
			return
		}
		svc := services.NewCashService(db)
		exps, err := svc.GetExpenses(fromEntry.Text, toEntry.Text)
		if err == nil {
			expenses = exps
		}

		colHeaders := []string{"Date", "Catégorie", "Description", "Mode Paiement", "Montant (DA)"}
		colWidths := []float32{90, 130, 250, 120, 120}

		headerRow := buildTableHeaderRow(colHeaders, colWidths)
		rows := container.NewVBox(headerRow)

		var total float64
		for _, e := range expenses {
			total += e.Amount
			cells := []string{
				utils.FormatDateFr(e.Date),
				e.CategoryName,
				utils.TruncateString(e.Description, 40),
				utils.PaymentMethodLabel(e.PaymentMethod),
				utils.FormatAmount(e.Amount),
			}
			rows.Add(buildTableRow(cells, colWidths, false))
		}
		if len(expenses) > 0 {
			rows.Add(widget.NewSeparator())
			rows.Add(buildTableRow(
				[]string{"", "", "", "TOTAL", utils.FormatAmount(total)},
				colWidths, true,
			))
		}
		if len(expenses) == 0 {
			rows.Add(widget.NewLabel("   Aucune dépense pour la période."))
		}
		tableContainer.Objects = []fyne.CanvasObject{container.NewVScroll(rows)}
		tableContainer.Refresh()
	}

	periodFilter := buildPeriodFilter(fromEntry, toEntry, loadExpenses)

	newExpBtn := widget.NewButtonWithIcon("Nouvelle Dépense", theme.ContentAddIcon(), func() {
		showExpenseDialog(db, catNames, catIDs, loadExpenses)
	})
	newExpBtn.Importance = widget.HighImportance

	loadExpenses()

	return container.NewBorder(
		container.NewVBox(header, periodFilter, newExpBtn, widget.NewSeparator()),
		nil, nil, nil,
		tableContainer,
	)
}

func showExpenseDialog(_ interface{}, catNames []string, catIDs []int, onSave func()) {
	appDB := appstate.GetDB()
	win := appstate.MainWindow

	catSelect := widget.NewSelect(catNames, nil)
	if len(catNames) > 0 {
		catSelect.SetSelected(catNames[0])
	}
	dateEntry := widget.NewEntry()
	dateEntry.SetText(utils.TodayString())
	descEntry := widget.NewMultiLineEntry()
	descEntry.SetPlaceHolder("Description de la dépense")
	amountEntry := widget.NewEntry()
	amountEntry.SetPlaceHolder("0,00")
	methodSelect := widget.NewSelect([]string{"cash", "cheque", "virement"}, nil)
	methodSelect.SetSelected("cash")

	content := container.NewVBox(widget.NewForm(
		widget.NewFormItem("Catégorie", catSelect),
		widget.NewFormItem("Date *", dateEntry),
		widget.NewFormItem("Description *", descEntry),
		widget.NewFormItem("Montant (DA) *", amountEntry),
		widget.NewFormItem("Mode paiement", methodSelect),
	))

	dlg := dialog.NewCustomConfirm("Nouvelle Dépense", "Enregistrer", "Annuler",
		container.NewScroll(content),
		func(ok bool) {
			if !ok || appDB == nil {
				return
			}
			amount, err := strconv.ParseFloat(strings.ReplaceAll(amountEntry.Text, ",", "."), 64)
			if err != nil || amount <= 0 || strings.TrimSpace(descEntry.Text) == "" {
				dialog.ShowError(fmt.Errorf("montant invalide ou description manquante"), win)
				return
			}
			catID := 0
			catIdx := catSelect.SelectedIndex()
			if catIdx >= 0 && catIdx < len(catIDs) {
				catID = catIDs[catIdx]
			}

			session := appstate.GetSession()
			userID := 0
			if session != nil {
				userID = session.UserID
			}

			var catIDPtr *int
			if catID > 0 {
				catIDPtr = &catID
			}
			exp := &models.Expense{
				Date:          dateEntry.Text,
				CategoryID:    catIDPtr,
				Description:   descEntry.Text,
				Amount:        amount,
				PaymentMethod: methodSelect.Selected,
			}
			expCopy := *exp
			go func(e models.Expense) {
				svc := services.NewCashService(appDB)
				err := svc.SaveExpense(&e, userID)
				fyne.Do(func() {
					if err != nil {
						dialog.ShowError(err, win)
						return
					}
					dialog.ShowInformation("Succès", "Dépense enregistrée.", win)
					onSave()
				})
			}(expCopy)
		}, win)
	dlg.Resize(fyne.NewSize(460, 380))
	dlg.Show()
}

// ─────────────────────────────────────────────────────────────────────────────
// HELPERS TABLE (réutilisés par Lot 9 aussi)
// ─────────────────────────────────────────────────────────────────────────────

// buildTableHeaderRow construit une ligne d'en-tête de tableau
func buildTableHeaderRow(headers []string, widths []float32) fyne.CanvasObject {
	row := container.NewHBox()
	for i, h := range headers {
		lbl := widget.NewLabel(h)
		lbl.TextStyle = fyne.TextStyle{Bold: true}
		lbl.Wrapping = fyne.TextWrapOff
		if i < len(widths) {
			lbl.Resize(fyne.NewSize(widths[i], 28))
		}
		row.Add(lbl)
	}
	return container.NewVBox(row, widget.NewSeparator())
}

// buildTableRow construit une ligne de données dans un tableau
func buildTableRow(cells []string, widths []float32, bold bool) fyne.CanvasObject {
	row := container.NewHBox()
	for i, cell := range cells {
		lbl := widget.NewLabel(cell)
		lbl.TextStyle = fyne.TextStyle{Bold: bold}
		lbl.Wrapping = fyne.TextWrapOff
		if i < len(widths) {
			lbl.Resize(fyne.NewSize(widths[i], 28))
		}
		row.Add(lbl)
	}
	return row
}
