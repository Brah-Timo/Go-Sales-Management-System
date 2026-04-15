package screens

import (
	"fmt"
	"gestion-commerciale/internal/app"
	"gestion-commerciale/internal/models"
	"gestion-commerciale/internal/services"
	"gestion-commerciale/pkg/utils"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// ─────────────────────────────────────────────────────────────────────────────
// CLIENTS
// ─────────────────────────────────────────────────────────────────────────────

// BuildClientsScreen construit l'écran de gestion des clients
func BuildClientsScreen() fyne.CanvasObject {
	db := app.GetDB()
	if db == nil {
		return PlaceholderScreen("Clients (DB non connectée)")
	}

	session := app.GetSession()
	userID := 1
	if session != nil {
		userID = session.UserID
	}

	svc := services.NewClientService(db)

	searchEntry := widget.NewEntry()
	searchEntry.SetPlaceHolder("🔍 Rechercher par nom, code, téléphone...")

	var clients []models.Client
	selectedRow := -1

	headers := []string{"Code", "Nom (Fr)", "Nom (Ar)", "Téléphone", "Wilaya", "Solde (DA)", "Statut"}
	colWidths := []float32{70, 180, 150, 110, 120, 110, 80}

	table := widget.NewTable(
		func() (int, int) { return len(clients) + 1, len(headers) },
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
			if id.Row-1 >= len(clients) {
				lbl.SetText("")
				return
			}
			c := clients[id.Row-1]
			lbl.TextStyle = fyne.TextStyle{}
			switch id.Col {
			case 0:
				lbl.SetText(c.Code)
			case 1:
				lbl.SetText(c.NameFr)
			case 2:
				lbl.SetText(c.NameAr)
			case 3:
				ph := c.Phone
				if ph == "" {
					ph = c.Mobile
				}
				lbl.SetText(ph)
			case 4:
				lbl.SetText(c.Wilaya)
			case 5:
				lbl.SetText(utils.FormatMoney(c.Balance))
			case 6:
				if c.IsBlocked {
					lbl.SetText("🔒 Bloqué")
				} else {
					lbl.SetText("✅ Actif")
				}
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

	loadClients := func() {
		clients, _ = svc.GetAllClients(strings.TrimSpace(searchEntry.Text))
		selectedRow = -1
		table.Refresh()
	}
	loadClients()

	searchEntry.OnChanged = func(s string) { loadClients() }

	// ── Formulaire Client
	showClientForm := func(client *models.Client) {
		isNew := client == nil
		if isNew {
			client = &models.Client{Type: "company"}
		}

		nameFrEntry := widget.NewEntry()
		nameFrEntry.SetText(client.NameFr)
		nameArEntry := widget.NewEntry()
		nameArEntry.SetText(client.NameAr)

		typeSelect := widget.NewSelect([]string{"Entreprise", "Particulier"}, nil)
		if client.Type == "person" {
			typeSelect.SetSelected("Particulier")
		} else {
			typeSelect.SetSelected("Entreprise")
		}

		phoneEntry := widget.NewEntry()
		phoneEntry.SetText(client.Phone)
		mobileEntry := widget.NewEntry()
		mobileEntry.SetText(client.Mobile)
		emailEntry := widget.NewEntry()
		emailEntry.SetText(client.Email)

		wilayaSelect := widget.NewSelect(utils.WilayaNames(), nil)
		if client.Wilaya != "" {
			wilayaSelect.SetSelected(client.Wilaya)
		}
		communeEntry := widget.NewEntry()
		communeEntry.SetText(client.Commune)
		addressEntry := widget.NewMultiLineEntry()
		addressEntry.SetText(client.Address)
		addressEntry.SetMinRowsVisible(2)

		nifEntry := widget.NewEntry()
		nifEntry.SetText(client.NIF)
		nisEntry := widget.NewEntry()
		nisEntry.SetText(client.NIS)
		rcEntry := widget.NewEntry()
		rcEntry.SetText(client.RC)
		aiEntry := widget.NewEntry()
		aiEntry.SetText(client.AI)

		creditLimitEntry := widget.NewEntry()
		creditLimitEntry.SetText(fmt.Sprintf("%.2f", client.CreditLimit))
		discountEntry := widget.NewEntry()
		discountEntry.SetText(fmt.Sprintf("%.2f", client.DiscountRate))

		payTermsSelect := widget.NewSelect(
			[]string{"Immédiat", "30 jours", "60 jours", "90 jours"}, nil,
		)
		if client.PaymentTerms != "" {
			payTermsSelect.SetSelected(client.PaymentTerms)
		} else {
			payTermsSelect.SetSelected("Immédiat")
		}

		notesEntry := widget.NewMultiLineEntry()
		notesEntry.SetText(client.Notes)
		notesEntry.SetMinRowsVisible(2)

		// Tabs: Général | Fiscal | Conditions
		generalTab := container.NewTabItem("Général",
			widget.NewForm(
				widget.NewFormItem("Nom (Français) *", nameFrEntry),
				widget.NewFormItem("Nom (Arabe)", nameArEntry),
				widget.NewFormItem("Type", typeSelect),
				widget.NewFormItem("Téléphone", phoneEntry),
				widget.NewFormItem("Mobile", mobileEntry),
				widget.NewFormItem("Email", emailEntry),
				widget.NewFormItem("Wilaya", wilayaSelect),
				widget.NewFormItem("Commune", communeEntry),
				widget.NewFormItem("Adresse", addressEntry),
				widget.NewFormItem("Notes", notesEntry),
			),
		)

		fiscalTab := container.NewTabItem("Fiscal",
			widget.NewForm(
				widget.NewFormItem("NIF", nifEntry),
				widget.NewFormItem("NIS", nisEntry),
				widget.NewFormItem("Registre Commerce", rcEntry),
				widget.NewFormItem("Article Imposition", aiEntry),
			),
		)

		conditionsTab := container.NewTabItem("Conditions",
			widget.NewForm(
				widget.NewFormItem("Plafond Crédit (DA)", creditLimitEntry),
				widget.NewFormItem("Remise (%)", discountEntry),
				widget.NewFormItem("Délai Paiement", payTermsSelect),
			),
		)

		tabs := container.NewAppTabs(generalTab, fiscalTab, conditionsTab)

		title := "Nouveau Client"
		if !isNew {
			title = fmt.Sprintf("Modifier Client — %s", client.Code)
		}

		dlg := dialog.NewCustomConfirm(title, "Enregistrer", "Annuler", tabs,
			func(ok bool) {
				if !ok {
					return
				}
				if strings.TrimSpace(nameFrEntry.Text) == "" {
					dialog.ShowError(fmt.Errorf("le nom (Français) est obligatoire"), app.MainWindow)
					return
				}
				client.NameFr = strings.TrimSpace(nameFrEntry.Text)
				client.NameAr = strings.TrimSpace(nameArEntry.Text)
				if typeSelect.Selected == "Particulier" {
					client.Type = "person"
				} else {
					client.Type = "company"
				}
				client.Phone = strings.TrimSpace(phoneEntry.Text)
				client.Mobile = strings.TrimSpace(mobileEntry.Text)
				client.Email = strings.TrimSpace(emailEntry.Text)
				client.Wilaya = wilayaSelect.Selected
				client.Commune = strings.TrimSpace(communeEntry.Text)
				client.Address = strings.TrimSpace(addressEntry.Text)
				client.NIF = strings.TrimSpace(nifEntry.Text)
				client.NIS = strings.TrimSpace(nisEntry.Text)
				client.RC = strings.TrimSpace(rcEntry.Text)
				client.AI = strings.TrimSpace(aiEntry.Text)
				client.Notes = strings.TrimSpace(notesEntry.Text)
				client.PaymentTerms = payTermsSelect.Selected
				fmt.Sscanf(creditLimitEntry.Text, "%f", &client.CreditLimit)
				fmt.Sscanf(discountEntry.Text, "%f", &client.DiscountRate)

				clientCopy := *client
				go func(cl models.Client, newRecord bool) {
					err := svc.SaveClient(&cl, userID)
					fyne.Do(func() {
						if err != nil {
							dialog.ShowError(err, app.MainWindow)
							return
						}
						loadClients()
						if newRecord {
							dialog.ShowInformation("Succès", fmt.Sprintf("Client %s créé ✅", cl.Code), app.MainWindow)
						} else {
							dialog.ShowInformation("Succès", "Client mis à jour ✅", app.MainWindow)
						}
					})
				}(clientCopy, isNew)
			}, app.MainWindow)
		dlg.Resize(fyne.NewSize(560, 480))
		dlg.Show()
	}

	// Boutons
	newBtn := widget.NewButtonWithIcon("Nouveau Client", theme.ContentAddIcon(), func() {
		showClientForm(nil)
	})
	newBtn.Importance = widget.HighImportance

	editBtn := widget.NewButtonWithIcon("Modifier", theme.DocumentCreateIcon(), func() {
		if selectedRow < 0 || selectedRow >= len(clients) {
			dialog.ShowInformation("Info", "Sélectionnez un client", app.MainWindow)
			return
		}
		c := clients[selectedRow]
		showClientForm(&c)
	})

	deleteBtn := widget.NewButtonWithIcon("Supprimer", theme.DeleteIcon(), func() {
		if selectedRow < 0 || selectedRow >= len(clients) {
			return
		}
		c := clients[selectedRow]
		dialog.ShowConfirm("Supprimer",
			fmt.Sprintf("Supprimer le client %s - %s ?", c.Code, c.NameFr),
			func(ok bool) {
				if !ok {
					return
				}
				cID2 := c.ID
				go func() {
					err := svc.DeleteClient(cID2, userID)
					fyne.Do(func() {
						if err != nil {
							dialog.ShowError(err, app.MainWindow)
							return
						}
						loadClients()
					})
				}()
			}, app.MainWindow)
	})
	deleteBtn.Importance = widget.DangerImportance

	blockBtn := widget.NewButtonWithIcon("Bloquer/Débloquer", theme.CancelIcon(), func() {
		if selectedRow < 0 || selectedRow >= len(clients) {
			return
		}
		c := clients[selectedRow]
		err := svc.ToggleBlockClient(c.ID, !c.IsBlocked)
		if err != nil {
			dialog.ShowError(err, app.MainWindow)
			return
		}
		loadClients()
	})

	// Relevé de compte
	statementBtn := widget.NewButtonWithIcon("Relevé Compte", theme.ListIcon(), func() {
		if selectedRow < 0 || selectedRow >= len(clients) {
			dialog.ShowInformation("Info", "Sélectionnez un client", app.MainWindow)
			return
		}
		c := clients[selectedRow]
		fromEntry := widget.NewEntry()
		fromEntry.SetText("2025-01-01")
		toEntry := widget.NewEntry()
		toEntry.SetText(utils.TodayString())

		dialog.ShowCustomConfirm("Relevé de Compte", "Afficher", "Annuler",
			widget.NewForm(
				widget.NewFormItem("De:", fromEntry),
				widget.NewFormItem("Au:", toEntry),
			),
			func(ok bool) {
				if !ok {
					return
				}
				lines, err := svc.GetClientStatement(c.ID, fromEntry.Text, toEntry.Text)
				if err != nil || len(lines) == 0 {
					dialog.ShowInformation("Relevé", "Aucune opération trouvée", app.MainWindow)
					return
				}
				showClientStatement(c.NameFr, lines)
			}, app.MainWindow)
	})

	toolbar := container.NewHBox(newBtn, editBtn, deleteBtn, blockBtn, statementBtn)

	screenHdr := buildScreenHeader(
		"👥 Clients",
		"Gestion des clients: création, modification, relevé de compte",
		"#27ae60",
	)

	return container.NewBorder(
		container.NewVBox(screenHdr, searchEntry, toolbar),
		nil, nil, nil,
		table,
	)
}

// showClientStatement affiche la fenêtre du relevé de compte client
func showClientStatement(clientName string, lines []models.AccountStatement) {
	headers := []string{"Date", "Type", "N° Doc", "Description", "Débit", "Crédit", "Solde"}
	colWidths := []float32{90, 50, 100, 150, 100, 100, 110}

	stTable := widget.NewTable(
		func() (int, int) { return len(lines) + 1, len(headers) },
		func() fyne.CanvasObject { return widget.NewLabel("") },
		func(id widget.TableCellID, cell fyne.CanvasObject) {
			lbl := cell.(*widget.Label)
			if id.Row == 0 {
				lbl.SetText(headers[id.Col])
				lbl.TextStyle = fyne.TextStyle{Bold: true}
				return
			}
			if id.Row-1 >= len(lines) {
				lbl.SetText("")
				return
			}
			l := lines[id.Row-1]
			switch id.Col {
			case 0:
				lbl.SetText(l.Date)
			case 1:
				lbl.SetText(l.DocType)
			case 2:
				lbl.SetText(l.DocNumber)
			case 3:
				lbl.SetText(l.Description)
			case 4:
				if l.Debit > 0 {
					lbl.SetText(utils.FormatMoney(l.Debit))
				} else {
					lbl.SetText("")
				}
			case 5:
				if l.Credit > 0 {
					lbl.SetText(utils.FormatMoney(l.Credit))
				} else {
					lbl.SetText("")
				}
			case 6:
				lbl.SetText(utils.FormatMoney(l.Balance))
			}
		},
	)
	for i, w := range colWidths {
		stTable.SetColumnWidth(i, w)
	}

	dlg := dialog.NewCustom(
		fmt.Sprintf("📋 Relevé de Compte — %s", clientName),
		"Fermer",
		container.NewBorder(nil, nil, nil, nil, stTable),
		app.MainWindow,
	)
	dlg.Resize(fyne.NewSize(780, 500))
	dlg.Show()
}

// ─────────────────────────────────────────────────────────────────────────────
// FOURNISSEURS
// ─────────────────────────────────────────────────────────────────────────────

// BuildSuppliersScreen construit l'écran de gestion des fournisseurs
func BuildSuppliersScreen() fyne.CanvasObject {
	db := app.GetDB()
	if db == nil {
		return PlaceholderScreen("Fournisseurs (DB non connectée)")
	}

	session := app.GetSession()
	userID := 1
	if session != nil {
		userID = session.UserID
	}

	svc := services.NewClientService(db)

	searchEntry := widget.NewEntry()
	searchEntry.SetPlaceHolder("🔍 Rechercher par nom, code...")

	var suppliers []models.Supplier
	selectedRow := -1

	headers := []string{"Code", "Nom (Fr)", "Téléphone", "Wilaya", "Solde (DA)", "Livraison ⭐", "Qualité ⭐"}
	colWidths := []float32{70, 200, 110, 120, 110, 90, 90}

	table := widget.NewTable(
		func() (int, int) { return len(suppliers) + 1, len(headers) },
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
			if id.Row-1 >= len(suppliers) {
				lbl.SetText("")
				return
			}
			s := suppliers[id.Row-1]
			lbl.TextStyle = fyne.TextStyle{}
			switch id.Col {
			case 0:
				lbl.SetText(s.Code)
			case 1:
				lbl.SetText(s.NameFr)
			case 2:
				ph := s.Phone
				if ph == "" {
					ph = s.Mobile
				}
				lbl.SetText(ph)
			case 3:
				lbl.SetText(s.Wilaya)
			case 4:
				lbl.SetText(utils.FormatMoney(s.Balance))
			case 5:
				lbl.SetText(ratingStars(s.RatingDelivery))
			case 6:
				lbl.SetText(ratingStars(s.RatingQuality))
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

	loadSuppliers := func() {
		suppliers, _ = svc.GetAllSuppliers(strings.TrimSpace(searchEntry.Text))
		selectedRow = -1
		table.Refresh()
	}
	loadSuppliers()
	searchEntry.OnChanged = func(s string) { loadSuppliers() }

	// ── Formulaire Fournisseur
	showSupplierForm := func(sup *models.Supplier) {
		isNew := sup == nil
		if isNew {
			sup = &models.Supplier{}
		}

		nameFrEntry := widget.NewEntry()
		nameFrEntry.SetText(sup.NameFr)
		nameArEntry := widget.NewEntry()
		nameArEntry.SetText(sup.NameAr)
		phoneEntry := widget.NewEntry()
		phoneEntry.SetText(sup.Phone)
		mobileEntry := widget.NewEntry()
		mobileEntry.SetText(sup.Mobile)
		emailEntry := widget.NewEntry()
		emailEntry.SetText(sup.Email)
		wilayaSelect := widget.NewSelect(utils.WilayaNames(), nil)
		if sup.Wilaya != "" {
			wilayaSelect.SetSelected(sup.Wilaya)
		}
		addressEntry := widget.NewMultiLineEntry()
		addressEntry.SetText(sup.Address)
		addressEntry.SetMinRowsVisible(2)

		nifEntry := widget.NewEntry()
		nifEntry.SetText(sup.NIF)
		nisEntry := widget.NewEntry()
		nisEntry.SetText(sup.NIS)
		rcEntry := widget.NewEntry()
		rcEntry.SetText(sup.RC)
		aiEntry := widget.NewEntry()
		aiEntry.SetText(sup.AI)

		payTermsSelect := widget.NewSelect(
			[]string{"Immédiat", "30 jours", "60 jours", "90 jours"}, nil,
		)
		if sup.PaymentTerms != "" {
			payTermsSelect.SetSelected(sup.PaymentTerms)
		} else {
			payTermsSelect.SetSelected("30 jours")
		}

		ratingDeliverySelect := widget.NewSelect([]string{"1", "2", "3", "4", "5"}, nil)
		ratingDeliverySelect.SetSelected(fmt.Sprintf("%d", max(sup.RatingDelivery, 3)))
		ratingQualitySelect := widget.NewSelect([]string{"1", "2", "3", "4", "5"}, nil)
		ratingQualitySelect.SetSelected(fmt.Sprintf("%d", max(sup.RatingQuality, 3)))

		notesEntry := widget.NewMultiLineEntry()
		notesEntry.SetText(sup.Notes)
		notesEntry.SetMinRowsVisible(2)

		generalTab := container.NewTabItem("Général",
			widget.NewForm(
				widget.NewFormItem("Nom (Français) *", nameFrEntry),
				widget.NewFormItem("Nom (Arabe)", nameArEntry),
				widget.NewFormItem("Téléphone", phoneEntry),
				widget.NewFormItem("Mobile", mobileEntry),
				widget.NewFormItem("Email", emailEntry),
				widget.NewFormItem("Wilaya", wilayaSelect),
				widget.NewFormItem("Adresse", addressEntry),
				widget.NewFormItem("Notes", notesEntry),
			),
		)
		fiscalTab := container.NewTabItem("Fiscal",
			widget.NewForm(
				widget.NewFormItem("NIF", nifEntry),
				widget.NewFormItem("NIS", nisEntry),
				widget.NewFormItem("Registre Commerce", rcEntry),
				widget.NewFormItem("Article Imposition", aiEntry),
				widget.NewFormItem("Délai Paiement", payTermsSelect),
			),
		)
		ratingTab := container.NewTabItem("Évaluation",
			widget.NewForm(
				widget.NewFormItem("Note Livraison (1-5)", ratingDeliverySelect),
				widget.NewFormItem("Note Qualité (1-5)", ratingQualitySelect),
			),
		)

		tabs := container.NewAppTabs(generalTab, fiscalTab, ratingTab)

		title := "Nouveau Fournisseur"
		if !isNew {
			title = fmt.Sprintf("Modifier Fournisseur — %s", sup.Code)
		}

		dlg := dialog.NewCustomConfirm(title, "Enregistrer", "Annuler", tabs,
			func(ok bool) {
				if !ok {
					return
				}
				if strings.TrimSpace(nameFrEntry.Text) == "" {
					dialog.ShowError(fmt.Errorf("le nom est obligatoire"), app.MainWindow)
					return
				}
				sup.NameFr = strings.TrimSpace(nameFrEntry.Text)
				sup.NameAr = strings.TrimSpace(nameArEntry.Text)
				sup.Phone = strings.TrimSpace(phoneEntry.Text)
				sup.Mobile = strings.TrimSpace(mobileEntry.Text)
				sup.Email = strings.TrimSpace(emailEntry.Text)
				sup.Wilaya = wilayaSelect.Selected
				sup.Address = strings.TrimSpace(addressEntry.Text)
				sup.NIF = strings.TrimSpace(nifEntry.Text)
				sup.NIS = strings.TrimSpace(nisEntry.Text)
				sup.RC = strings.TrimSpace(rcEntry.Text)
				sup.AI = strings.TrimSpace(aiEntry.Text)
				sup.PaymentTerms = payTermsSelect.Selected
				sup.Notes = strings.TrimSpace(notesEntry.Text)
				fmt.Sscanf(ratingDeliverySelect.Selected, "%d", &sup.RatingDelivery)
				fmt.Sscanf(ratingQualitySelect.Selected, "%d", &sup.RatingQuality)

				supCopy := *sup
				go func(s models.Supplier, newRec bool) {
					err := svc.SaveSupplier(&s, userID)
					fyne.Do(func() {
						if err != nil {
							dialog.ShowError(err, app.MainWindow)
							return
						}
						loadSuppliers()
						if newRec {
							dialog.ShowInformation("Succès", fmt.Sprintf("Fournisseur %s créé ✅", s.Code), app.MainWindow)
						} else {
							dialog.ShowInformation("Succès", "Fournisseur mis à jour ✅", app.MainWindow)
						}
					})
				}(supCopy, isNew)
			}, app.MainWindow)
		dlg.Resize(fyne.NewSize(540, 460))
		dlg.Show()
	}

	newBtn := widget.NewButtonWithIcon("Nouveau Fournisseur", theme.ContentAddIcon(), func() {
		showSupplierForm(nil)
	})
	newBtn.Importance = widget.HighImportance

	editBtn := widget.NewButtonWithIcon("Modifier", theme.DocumentCreateIcon(), func() {
		if selectedRow < 0 || selectedRow >= len(suppliers) {
			dialog.ShowInformation("Info", "Sélectionnez un fournisseur", app.MainWindow)
			return
		}
		s := suppliers[selectedRow]
		showSupplierForm(&s)
	})

	deleteBtn := widget.NewButtonWithIcon("Supprimer", theme.DeleteIcon(), func() {
		if selectedRow < 0 || selectedRow >= len(suppliers) {
			return
		}
		s := suppliers[selectedRow]
		dialog.ShowConfirm("Supprimer",
			fmt.Sprintf("Supprimer le fournisseur %s - %s ?", s.Code, s.NameFr),
			func(ok bool) {
				if !ok {
					return
				}
				sID := s.ID
				go func() {
					err := svc.DeleteSupplier(sID, userID)
					fyne.Do(func() {
						if err != nil {
							dialog.ShowError(err, app.MainWindow)
							return
						}
						loadSuppliers()
					})
				}()
			}, app.MainWindow)
	})
	deleteBtn.Importance = widget.DangerImportance

	toolbar := container.NewHBox(newBtn, editBtn, deleteBtn)

	screenHdr := buildScreenHeader(
		"🏭 Fournisseurs",
		"Gestion des fournisseurs: création, modification, évaluation",
		"#2980b9",
	)

	return container.NewBorder(
		container.NewVBox(screenHdr, searchEntry, toolbar),
		nil, nil, nil,
		table,
	)
}

// ─────────────────────────────────────────────────────────────────────────────
// CHAUFFEURS / LIVREURS
// ─────────────────────────────────────────────────────────────────────────────

// BuildDriversScreen construit l'écran de gestion des chauffeurs
func BuildDriversScreen() fyne.CanvasObject {
	db := app.GetDB()
	if db == nil {
		return PlaceholderScreen("Chauffeurs (DB non connectée)")
	}

	svc := services.NewClientService(db)

	var drivers []models.Driver
	selectedRow := -1

	headers := []string{"Nom Complet", "Téléphone", "Immatriculation", "Livraisons"}
	colWidths := []float32{200, 130, 130, 100}

	table := widget.NewTable(
		func() (int, int) { return len(drivers) + 1, len(headers) },
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
			if id.Row-1 >= len(drivers) {
				lbl.SetText("")
				return
			}
			d := drivers[id.Row-1]
			lbl.TextStyle = fyne.TextStyle{}
			switch id.Col {
			case 0:
				lbl.SetText(d.Name)
			case 1:
				lbl.SetText(d.Phone)
			case 2:
				lbl.SetText(d.VehiclePlate)
			case 3:
				lbl.SetText(fmt.Sprintf("%d BL", d.DeliveryCount))
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

	loadDrivers := func() {
		drivers, _ = svc.GetAllDrivers()
		selectedRow = -1
		table.Refresh()
	}
	loadDrivers()

	showDriverForm := func(driver *models.Driver) {
		isNew := driver == nil
		if isNew {
			driver = &models.Driver{}
		}

		nameEntry := widget.NewEntry()
		nameEntry.SetText(driver.Name)
		nameEntry.SetPlaceHolder("Prénom et Nom...")
		phoneEntry := widget.NewEntry()
		phoneEntry.SetText(driver.Phone)
		phoneEntry.SetPlaceHolder("0XXX XX XX XX")
		plateEntry := widget.NewEntry()
		plateEntry.SetText(driver.VehiclePlate)
		plateEntry.SetPlaceHolder("123 ABC 16...")

		content := widget.NewForm(
			widget.NewFormItem("Nom Complet *", nameEntry),
			widget.NewFormItem("Téléphone", phoneEntry),
			widget.NewFormItem("Immatriculation", plateEntry),
		)

		title := "Nouveau Chauffeur / Livreur"
		if !isNew {
			title = fmt.Sprintf("Modifier — %s", driver.Name)
		}

		dlg := dialog.NewCustomConfirm(title, "Enregistrer", "Annuler", content,
			func(ok bool) {
				if !ok {
					return
				}
				if strings.TrimSpace(nameEntry.Text) == "" {
					dialog.ShowError(fmt.Errorf("le nom est obligatoire"), app.MainWindow)
					return
				}
				driver.Name = strings.TrimSpace(nameEntry.Text)
				driver.Phone = strings.TrimSpace(phoneEntry.Text)
				driver.VehiclePlate = strings.TrimSpace(plateEntry.Text)

				drvCopy := *driver
				go func(drv models.Driver, newRec bool) {
					err := svc.SaveDriver(&drv)
					fyne.Do(func() {
						if err != nil {
							dialog.ShowError(err, app.MainWindow)
							return
						}
						loadDrivers()
						if newRec {
							dialog.ShowInformation("Succès", fmt.Sprintf("Chauffeur %s ajouté ✅", drv.Name), app.MainWindow)
						} else {
							dialog.ShowInformation("Succès", "Chauffeur mis à jour ✅", app.MainWindow)
						}
					})
				}(drvCopy, isNew)
			}, app.MainWindow)
		dlg.Resize(fyne.NewSize(400, 260))
		dlg.Show()
	}

	newBtn := widget.NewButtonWithIcon("Nouveau Chauffeur", theme.ContentAddIcon(), func() {
		showDriverForm(nil)
	})
	newBtn.Importance = widget.HighImportance

	editBtn := widget.NewButtonWithIcon("Modifier", theme.DocumentCreateIcon(), func() {
		if selectedRow < 0 || selectedRow >= len(drivers) {
			dialog.ShowInformation("Info", "Sélectionnez un chauffeur", app.MainWindow)
			return
		}
		d := drivers[selectedRow]
		showDriverForm(&d)
	})

	deleteBtn := widget.NewButtonWithIcon("Supprimer", theme.DeleteIcon(), func() {
		if selectedRow < 0 || selectedRow >= len(drivers) {
			return
		}
		d := drivers[selectedRow]
		dialog.ShowConfirm("Supprimer",
			fmt.Sprintf("Supprimer le chauffeur %s ?", d.Name),
			func(ok bool) {
				if !ok {
					return
				}
				dID := d.ID
				go func() {
					err := svc.DeleteDriver(dID)
					fyne.Do(func() {
						if err != nil {
							dialog.ShowError(err, app.MainWindow)
							return
						}
						loadDrivers()
					})
				}()
			}, app.MainWindow)
	})
	deleteBtn.Importance = widget.DangerImportance

	toolbar := container.NewHBox(newBtn, editBtn, deleteBtn)

	screenHdr := buildScreenHeader(
		"🚚 Chauffeurs / Livreurs",
		"Gestion des livreurs: ajout, modification, statistiques de livraison",
		"#16a085",
	)

	return container.NewBorder(
		container.NewVBox(screenHdr, toolbar),
		nil, nil, nil,
		table,
	)
}

// ─────────────────────────────────────────────────────────────────────────────
// Helpers
// ─────────────────────────────────────────────────────────────────────────────

// ratingStars converts a numeric rating to a star string
func ratingStars(r int) string {
	if r <= 0 {
		return "—"
	}
	stars := ""
	for i := 0; i < r && i < 5; i++ {
		stars += "★"
	}
	for i := r; i < 5; i++ {
		stars += "☆"
	}
	return stars
}

// max returns the larger of a or b (int)
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
