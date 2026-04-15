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
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"image/color"
)

// ─────────────────────────────────────────────────────────────────────────────
// ÉCRAN LISTE DES ARTICLES
// ─────────────────────────────────────────────────────────────────────────────

// BuildArticlesListScreen construit l'écran de liste des articles
func BuildArticlesListScreen() fyne.CanvasObject {
	db := app.GetDB()

	// ── Barre de recherche & filtres
	searchEntry := widget.NewEntry()
	searchEntry.SetPlaceHolder("🔍 Rechercher par nom, référence, code-barres...")

	categorySelect := widget.NewSelect([]string{"Toutes les catégories"}, nil)
	categorySelect.SetSelected("Toutes les catégories")

	// Charger les catégories
	if db != nil {
		cats, _ := queries.GetAllCategories(db)
		opts := []string{"Toutes les catégories"}
		for _, c := range cats {
			opts = append(opts, c.NameFr)
		}
		categorySelect.Options = opts
	}

	// ── Tableau des articles
	headers := []string{"Réf.", "Désignation", "Catégorie", "Stock", "P.V. HT", "P.V. TTC", "TVA%", "État"}
	colWidths := []float32{80, 200, 120, 70, 100, 100, 60, 70}

	var articles []models.Article
	selectedRow := -1

	loadArticles := func(search, category string) {
		if db == nil {
			return
		}
		catID := 0
		if category != "Toutes les catégories" {
			cats, _ := queries.GetAllCategories(db)
			for _, c := range cats {
				if c.NameFr == category {
					catID = c.ID
					break
				}
			}
		}
		svc := services.NewArticleService(db)
		var err error
		articles, err = svc.Search(search, catID, false)
		if err != nil {
			articles = nil
		}
	}

	// Table widget
	table := widget.NewTable(
		func() (int, int) { return len(articles) + 1, len(headers) },
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
			if id.Row-1 >= len(articles) {
				lbl.SetText("")
				return
			}
			a := articles[id.Row-1]
			switch id.Col {
			case 0:
				lbl.SetText(a.Reference)
			case 1:
				lbl.SetText(a.NameFr)
			case 2:
				lbl.SetText(a.CategoryName)
			case 3:
				lbl.SetText(fmt.Sprintf("%.2f", a.StockQty))
			case 4:
				lbl.SetText(utils.FormatMoney(a.SalePriceHT))
			case 5:
				lbl.SetText(utils.FormatMoney(a.SalePriceTTC))
			case 6:
				lbl.SetText(fmt.Sprintf("%.0f%%", a.TVARate))
			case 7:
				if a.IsActive {
					lbl.SetText("✅ Actif")
				} else {
					lbl.SetText("❌ Inactif")
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

	// Charger initialement
	loadArticles("", "Toutes les catégories")
	table.Refresh()

	// Compteur
	countLabel := widget.NewLabel(fmt.Sprintf("%d article(s)", len(articles)))

	// Recherche dynamique
	searchEntry.OnChanged = func(s string) {
		loadArticles(s, categorySelect.Selected)
		table.Refresh()
		countLabel.SetText(fmt.Sprintf("%d article(s)", len(articles)))
	}
	categorySelect.OnChanged = func(s string) {
		loadArticles(searchEntry.Text, s)
		table.Refresh()
		countLabel.SetText(fmt.Sprintf("%d article(s)", len(articles)))
	}

	// ── Boutons d'action
	newBtn := widget.NewButtonWithIcon("Nouveau", theme.ContentAddIcon(), func() {
		showArticleForm(nil, db, func() {
			loadArticles(searchEntry.Text, categorySelect.Selected)
			table.Refresh()
			countLabel.SetText(fmt.Sprintf("%d article(s)", len(articles)))
		})
	})
	newBtn.Importance = widget.HighImportance

	editBtn := widget.NewButtonWithIcon("Modifier", theme.DocumentCreateIcon(), func() {
		if selectedRow < 0 || selectedRow >= len(articles) {
			dialog.ShowInformation("Info", "Sélectionnez un article à modifier", app.MainWindow)
			return
		}
		a := articles[selectedRow]
		full, err := queries.GetArticleByID(db, a.ID)
		if err != nil {
			dialog.ShowError(err, app.MainWindow)
			return
		}
		showArticleForm(full, db, func() {
			loadArticles(searchEntry.Text, categorySelect.Selected)
			table.Refresh()
		})
	})

	deleteBtn := widget.NewButtonWithIcon("Supprimer", theme.DeleteIcon(), func() {
		if selectedRow < 0 || selectedRow >= len(articles) {
			dialog.ShowInformation("Info", "Sélectionnez un article à supprimer", app.MainWindow)
			return
		}
		a := articles[selectedRow]
		dialog.ShowConfirm("Supprimer", fmt.Sprintf("Supprimer l'article '%s' ?", a.NameFr),
			func(ok bool) {
				if !ok {
					return
				}
				err := queries.DeleteArticle(db, a.ID)
				if err != nil {
					dialog.ShowError(err, app.MainWindow)
					return
				}
				loadArticles(searchEntry.Text, categorySelect.Selected)
				table.Refresh()
			}, app.MainWindow)
	})
	deleteBtn.Importance = widget.DangerImportance

	barcodeBtn := widget.NewButtonWithIcon("Chercher Barcode", theme.SearchIcon(), func() {
		showBarcodeLookup(db)
	})

	// ── Layout
	toolbar := container.NewHBox(
		newBtn, editBtn, deleteBtn, widget.NewSeparator(),
		barcodeBtn, widget.NewSeparator(),
		countLabel,
	)

	filters := container.NewGridWithColumns(2,
		container.NewBorder(nil, nil, widget.NewLabel("Recherche: "), nil, searchEntry),
		container.NewBorder(nil, nil, widget.NewLabel("Catégorie: "), nil, categorySelect),
	)

	header := buildScreenHeader("📦 Catalogue Articles",
		"Gérez votre catalogue de produits et services", "#2980b9")

	return container.NewBorder(
		container.NewVBox(header, toolbar, filters),
		nil, nil, nil,
		table,
	)
}

// ─────────────────────────────────────────────────────────────────────────────
// FORMULAIRE ARTICLE (Nouveau / Modifier)
// ─────────────────────────────────────────────────────────────────────────────

func showArticleForm(article *models.Article, db *sql.DB, onSave func()) {
	if db == nil {
		db = app.GetDB()
	}
	if db == nil {
		return
	}

	isNew := article == nil
	title := "Nouveau Article"
	if !isNew {
		title = "Modifier Article: " + article.NameFr
	}

	// Champs du formulaire
	refEntry := widget.NewEntry()
	refEntry.SetPlaceHolder("ART-0001")

	nameArEntry := widget.NewEntry()
	nameArEntry.SetPlaceHolder("الاسم بالعربية")

	nameFrEntry := widget.NewEntry()
	nameFrEntry.SetPlaceHolder("Nom en français")

	barcodeEntry := widget.NewEntry()
	barcodeEntry.SetPlaceHolder("Code-barres EAN13")

	descEntry := widget.NewMultiLineEntry()
	descEntry.SetPlaceHolder("Description de l'article...")
	descEntry.SetMinRowsVisible(3)

	// Catégorie
	cats, _ := queries.GetAllCategories(db)
	catNames := []string{"Aucune"}
	catMap := map[string]int{"Aucune": 0}
	for _, c := range cats {
		catNames = append(catNames, c.NameFr)
		catMap[c.NameFr] = c.ID
	}
	catSelect := widget.NewSelect(catNames, nil)
	catSelect.SetSelected("Aucune")

	// Unités
	units, _ := queries.GetAllUnits(db)
	unitNames := []string{"Aucune"}
	unitMap := map[string]int{"Aucune": 0}
	for _, u := range units {
		unitNames = append(unitNames, u.NameFr)
		unitMap[u.NameFr] = u.ID
	}
	unitSelect := widget.NewSelect(unitNames, nil)
	unitSelect.SetSelected("Aucune")

	// TVA
	tvaSelect := widget.NewSelect([]string{"0", "9", "19"}, nil)
	tvaSelect.SetSelected("19")

	// Prix
	purchasePriceEntry := widget.NewEntry()
	purchasePriceEntry.SetPlaceHolder("0.00")

	salePriceHTEntry := widget.NewEntry()
	salePriceHTEntry.SetPlaceHolder("0.00")

	salePriceTTCEntry := widget.NewEntry()
	salePriceTTCEntry.SetPlaceHolder("0.00")

	marginEntry := widget.NewEntry()
	marginEntry.SetPlaceHolder("0.00")

	// Stock
	stockQtyEntry := widget.NewEntry()
	stockQtyEntry.SetPlaceHolder("0")

	stockMinEntry := widget.NewEntry()
	stockMinEntry.SetPlaceHolder("5")

	stockMaxEntry := widget.NewEntry()
	stockMaxEntry.SetPlaceHolder("100")

	// Options
	isActiveCheck := widget.NewCheck("Article actif", nil)
	isActiveCheck.SetChecked(true)

	lotTrackingCheck := widget.NewCheck("Suivi par lots", nil)
	expiryTrackingCheck := widget.NewCheck("Suivi dates d'expiration", nil)

	// Méthode valorisation
	valuationSelect := widget.NewSelect([]string{"CMUP", "FIFO"}, nil)
	valuationSelect.SetSelected("CMUP")

	// Calcul automatique prix TTC depuis HT
	salePriceHTEntry.OnChanged = func(s string) {
		ht, err := strconv.ParseFloat(strings.TrimSpace(s), 64)
		if err != nil {
			return
		}
		tvaRate, _ := strconv.ParseFloat(tvaSelect.Selected, 64)
		ttc := utils.HTToTTC(ht, tvaRate)
		salePriceTTCEntry.SetText(fmt.Sprintf("%.2f", ttc))
	}

	// Calcul depuis marge
	marginEntry.OnChanged = func(s string) {
		margin, err := strconv.ParseFloat(strings.TrimSpace(s), 64)
		if err != nil {
			return
		}
		purchase, _ := strconv.ParseFloat(strings.TrimSpace(purchasePriceEntry.Text), 64)
		if purchase > 0 {
			saleHT := utils.SalePriceFromMargin(purchase, margin)
			salePriceHTEntry.SetText(fmt.Sprintf("%.2f", saleHT))
		}
	}

	// Pré-remplissage si modification
	if !isNew {
		refEntry.SetText(article.Reference)
		nameArEntry.SetText(article.NameAr)
		nameFrEntry.SetText(article.NameFr)
		barcodeEntry.SetText(article.Barcode)
		descEntry.SetText(article.Description)
		purchasePriceEntry.SetText(fmt.Sprintf("%.2f", article.PurchasePrice))
		salePriceHTEntry.SetText(fmt.Sprintf("%.2f", article.SalePriceHT))
		salePriceTTCEntry.SetText(fmt.Sprintf("%.2f", article.SalePriceTTC))
		marginEntry.SetText(fmt.Sprintf("%.2f", article.MarginPercent))
		stockQtyEntry.SetText(fmt.Sprintf("%.2f", article.StockQty))
		stockMinEntry.SetText(fmt.Sprintf("%.2f", article.StockMin))
		stockMaxEntry.SetText(fmt.Sprintf("%.2f", article.StockMax))
		isActiveCheck.SetChecked(article.IsActive)
		lotTrackingCheck.SetChecked(article.LotTracking)
		expiryTrackingCheck.SetChecked(article.ExpiryTracking)
		tvaSelect.SetSelected(fmt.Sprintf("%.0f", article.TVARate))
		valuationSelect.SetSelected(article.ValuationMethod)

		// Catégorie
		for _, c := range cats {
			if article.CategoryID != nil && c.ID == *article.CategoryID {
				catSelect.SetSelected(c.NameFr)
				break
			}
		}
		// Unité
		for _, u := range units {
			if article.UnitID != nil && u.ID == *article.UnitID {
				unitSelect.SetSelected(u.NameFr)
				break
			}
		}
	} else {
		// Générer référence auto
		ref := queries.NextArticleReference(db)
		refEntry.SetText(ref)
	}

	// Formulaire en onglets
	infoTab := container.NewVBox(
		widget.NewForm(
			widget.NewFormItem("Référence *", refEntry),
			widget.NewFormItem("Code-barres", barcodeEntry),
			widget.NewFormItem("Nom (FR) *", nameFrEntry),
			widget.NewFormItem("Nom (AR)", nameArEntry),
			widget.NewFormItem("Catégorie", catSelect),
			widget.NewFormItem("Unité", unitSelect),
			widget.NewFormItem("Description", descEntry),
		),
	)

	priceTab := container.NewVBox(
		widget.NewForm(
			widget.NewFormItem("TVA (%)", tvaSelect),
			widget.NewFormItem("Prix d'achat (DA)", purchasePriceEntry),
			widget.NewFormItem("Marge (%)", marginEntry),
			widget.NewFormItem("Prix Vente HT (DA)", salePriceHTEntry),
			widget.NewFormItem("Prix Vente TTC (DA)", salePriceTTCEntry),
		),
	)

	stockTab := container.NewVBox(
		widget.NewForm(
			widget.NewFormItem("Méthode valorisation", valuationSelect),
			widget.NewFormItem("Stock actuel", stockQtyEntry),
			widget.NewFormItem("Stock minimum", stockMinEntry),
			widget.NewFormItem("Stock maximum", stockMaxEntry),
		),
		widget.NewSeparator(),
		isActiveCheck,
		lotTrackingCheck,
		expiryTrackingCheck,
	)

	tabs := container.NewAppTabs(
		container.NewTabItem("📋 Informations", infoTab),
		container.NewTabItem("💰 Prix & TVA", priceTab),
		container.NewTabItem("📊 Stock & Options", stockTab),
	)

	// Boutons
	cancelBtn := widget.NewButton("Annuler", nil)
	var dlg *dialog.CustomDialog

	saveBtn := widget.NewButtonWithIcon("Enregistrer", theme.DocumentSaveIcon(), func() {
		// Validation
		if strings.TrimSpace(nameFrEntry.Text) == "" {
			dialog.ShowError(fmt.Errorf("le nom en français est obligatoire"), app.MainWindow)
			return
		}
		if strings.TrimSpace(refEntry.Text) == "" {
			dialog.ShowError(fmt.Errorf("la référence est obligatoire"), app.MainWindow)
			return
		}

		// Construire l'article
		var a models.Article
		if !isNew {
			a.ID = article.ID
		}
		a.Reference = strings.TrimSpace(refEntry.Text)
		a.Barcode = strings.TrimSpace(barcodeEntry.Text)
		a.NameFr = strings.TrimSpace(nameFrEntry.Text)
		a.NameAr = strings.TrimSpace(nameArEntry.Text)
		a.Description = strings.TrimSpace(descEntry.Text)
		if cid := catMap[catSelect.Selected]; cid != 0 {
			a.CategoryID = &cid
		}
		if uid := unitMap[unitSelect.Selected]; uid != 0 {
			a.UnitID = &uid
		}
		a.TVARate, _ = strconv.ParseFloat(tvaSelect.Selected, 64)
		a.PurchasePrice, _ = strconv.ParseFloat(strings.TrimSpace(purchasePriceEntry.Text), 64)
		a.SalePriceHT, _ = strconv.ParseFloat(strings.TrimSpace(salePriceHTEntry.Text), 64)
		a.SalePriceTTC, _ = strconv.ParseFloat(strings.TrimSpace(salePriceTTCEntry.Text), 64)
		a.MarginPercent, _ = strconv.ParseFloat(strings.TrimSpace(marginEntry.Text), 64)
		a.StockQty, _ = strconv.ParseFloat(strings.TrimSpace(stockQtyEntry.Text), 64)
		a.StockMin, _ = strconv.ParseFloat(strings.TrimSpace(stockMinEntry.Text), 64)
		a.StockMax, _ = strconv.ParseFloat(strings.TrimSpace(stockMaxEntry.Text), 64)
		a.IsActive = isActiveCheck.Checked
		a.LotTracking = lotTrackingCheck.Checked
		a.ExpiryTracking = expiryTrackingCheck.Checked
		a.ValuationMethod = valuationSelect.Selected

		go func(art models.Article) {
			err := queries.SaveArticle(db, &art)
			fyne.Do(func() {
				if err != nil {
					dialog.ShowError(err, app.MainWindow)
					return
				}
				if dlg != nil {
					dlg.Hide()
				}
				onSave()
			})
		}(a)
	})
	saveBtn.Importance = widget.HighImportance

	content := container.NewBorder(
		nil,
		container.NewHBox(cancelBtn, saveBtn),
		nil, nil,
		tabs,
	)

	dlg = dialog.NewCustom(title, "Fermer", content, app.MainWindow)
	dlg.Resize(fyne.NewSize(600, 500))
	cancelBtn.OnTapped = func() { dlg.Hide() }
	dlg.Show()
}

// showBarcodeLookup ouvre la recherche par code-barres
func showBarcodeLookup(db *sql.DB) {
	if db == nil {
		db = app.GetDB()
	}
	if db == nil {
		return
	}

	entry := widget.NewEntry()
	entry.SetPlaceHolder("Scannez ou saisissez le code-barres...")
	resultLabel := widget.NewLabel("")

	doSearch := func() {
		code := strings.TrimSpace(entry.Text)
		if code == "" {
			return
		}
		a, err := queries.FindArticleByBarcode(db, code)
		if err != nil {
			resultLabel.SetText("❌ Article non trouvé: " + code)
			return
		}
		resultLabel.SetText(fmt.Sprintf(
			"✅ %s — %s\nStock: %.2f | P.V. TTC: %s",
			a.Reference, a.NameFr, a.StockQty, utils.FormatMoney(a.SalePriceTTC),
		))
	}

	entry.OnSubmitted = func(s string) { doSearch() }

	content := container.NewVBox(
		widget.NewLabel("Entrez le code-barres de l'article:"),
		entry,
		widget.NewButtonWithIcon("Rechercher", theme.SearchIcon(), doSearch),
		widget.NewSeparator(),
		resultLabel,
	)

	dlg := dialog.NewCustom("🏷️ Recherche par Code-barres", "Fermer", content, app.MainWindow)
	dlg.Resize(fyne.NewSize(450, 250))
	dlg.Show()
	app.MainWindow.Canvas().Focus(entry)
}

// ─────────────────────────────────────────────────────────────────────────────
// CATÉGORIES
// ─────────────────────────────────────────────────────────────────────────────

// BuildCategoriesScreen construit l'écran de gestion des catégories
func BuildCategoriesScreen() fyne.CanvasObject {
	db := app.GetDB()

	var categories []models.Category
	selectedRow := -1

	loadCats := func() {
		if db == nil {
			return
		}
		categories, _ = queries.GetAllCategories(db)
	}
	loadCats()

	table := widget.NewTable(
		func() (int, int) { return len(categories) + 1, 3 },
		func() fyne.CanvasObject { return widget.NewLabel("") },
		func(id widget.TableCellID, cell fyne.CanvasObject) {
			lbl := cell.(*widget.Label)
			headers := []string{"ID", "Nom (FR)", "Nom (AR)"}
			if id.Row == 0 {
				lbl.SetText(headers[id.Col])
				lbl.TextStyle = fyne.TextStyle{Bold: true}
				return
			}
			if id.Row-1 >= len(categories) {
				lbl.SetText("")
				return
			}
			c := categories[id.Row-1]
			switch id.Col {
			case 0:
				lbl.SetText(fmt.Sprintf("%d", c.ID))
			case 1:
				lbl.SetText(c.NameFr)
			case 2:
				lbl.SetText(c.NameAr)
			}
		},
	)
	table.SetColumnWidth(0, 50)
	table.SetColumnWidth(1, 200)
	table.SetColumnWidth(2, 200)
	table.OnSelected = func(id widget.TableCellID) {
		if id.Row > 0 {
			selectedRow = id.Row - 1
		}
	}

	showCatForm := func(cat *models.Category) {
		isNew := cat == nil
		title := "Nouvelle Catégorie"
		if !isNew {
			title = "Modifier: " + cat.NameFr
		}

		nameFrEntry := widget.NewEntry()
		nameArEntry := widget.NewEntry()
		descEntry := widget.NewEntry()

		if !isNew {
			nameFrEntry.SetText(cat.NameFr)
			nameArEntry.SetText(cat.NameAr)
			descEntry.SetText(cat.Description)
		}

		content := widget.NewForm(
			widget.NewFormItem("Nom (FR) *", nameFrEntry),
			widget.NewFormItem("Nom (AR)", nameArEntry),
			widget.NewFormItem("Description", descEntry),
		)

		dialog.ShowCustomConfirm(title, "Enregistrer", "Annuler", content,
			func(ok bool) {
				if !ok || strings.TrimSpace(nameFrEntry.Text) == "" {
					return
				}
				nFr2, nAr2, desc2 := nameFrEntry.Text, nameArEntry.Text, descEntry.Text
				go func() {
					if isNew {
						db.Exec(`INSERT INTO categories (name_fr, name_ar, description) VALUES (?,?,?)`,
						nFr2, nAr2, desc2)
					} else {
						db.Exec(`UPDATE categories SET name_fr=?, name_ar=?, description=? WHERE id=?`,
							nFr2, nAr2, desc2, cat.ID)
					}
					fyne.Do(func() { loadCats(); table.Refresh() })
				}()
			}, app.MainWindow)
	}

	newBtn := widget.NewButtonWithIcon("Nouvelle", theme.ContentAddIcon(), func() { showCatForm(nil) })
	newBtn.Importance = widget.HighImportance

	editBtn := widget.NewButtonWithIcon("Modifier", theme.DocumentCreateIcon(), func() {
		if selectedRow < 0 || selectedRow >= len(categories) {
			return
		}
		showCatForm(&categories[selectedRow])
	})

	deleteBtn := widget.NewButtonWithIcon("Supprimer", theme.DeleteIcon(), func() {
		if selectedRow < 0 || selectedRow >= len(categories) {
			return
		}
		c := categories[selectedRow]
		dialog.ShowConfirm("Supprimer", fmt.Sprintf("Supprimer '%s' ?", c.NameFr),
			func(ok bool) {
				if !ok {
					return
				}
				cID := c.ID
				go func() {
					db.Exec(`DELETE FROM categories WHERE id=?`, cID)
					fyne.Do(func() { loadCats(); table.Refresh() })
				}()
			}, app.MainWindow)
	})
	deleteBtn.Importance = widget.DangerImportance

	header := buildScreenHeader("📂 Familles / Catégories",
		"Gérez les familles et catégories de produits", "#8e44ad")

	toolbar := container.NewHBox(newBtn, editBtn, deleteBtn)

	return container.NewBorder(
		container.NewVBox(header, toolbar),
		nil, nil, nil,
		table,
	)
}

// ─────────────────────────────────────────────────────────────────────────────
// MARQUES
// ─────────────────────────────────────────────────────────────────────────────

// BuildBrandsScreen construit l'écran de gestion des marques
func BuildBrandsScreen() fyne.CanvasObject {
	db := app.GetDB()

	var brands []models.Brand
	selectedRow := -1

	loadBrands := func() {
		if db == nil {
			return
		}
		brands, _ = queries.GetAllBrands(db)
	}
	loadBrands()

	table := widget.NewTable(
		func() (int, int) { return len(brands) + 1, 3 },
		func() fyne.CanvasObject { return widget.NewLabel("") },
		func(id widget.TableCellID, cell fyne.CanvasObject) {
			lbl := cell.(*widget.Label)
			headers := []string{"Marque", "Pays", "Nb Produits"}
			if id.Row == 0 {
				lbl.SetText(headers[id.Col])
				lbl.TextStyle = fyne.TextStyle{Bold: true}
				return
			}
			if id.Row-1 >= len(brands) {
				lbl.SetText("")
				return
			}
			b := brands[id.Row-1]
			switch id.Col {
			case 0:
				lbl.SetText(b.Name)
			case 1:
				lbl.SetText(b.Country)
			case 2:
				lbl.SetText(fmt.Sprintf("%d", b.ProductCount))
			}
		},
	)
	table.SetColumnWidth(0, 200)
	table.SetColumnWidth(1, 150)
	table.SetColumnWidth(2, 100)
	table.OnSelected = func(id widget.TableCellID) {
		if id.Row > 0 {
			selectedRow = id.Row - 1
		}
	}

	showBrandForm := func(brand *models.Brand) {
		isNew := brand == nil
		title := "Nouvelle Marque"
		if !isNew {
			title = "Modifier: " + brand.Name
		}
		nameEntry := widget.NewEntry()
		countryEntry := widget.NewEntry()
		if !isNew {
			nameEntry.SetText(brand.Name)
			countryEntry.SetText(brand.Country)
		}
		content := widget.NewForm(
			widget.NewFormItem("Nom *", nameEntry),
			widget.NewFormItem("Pays", countryEntry),
		)
		dialog.ShowCustomConfirm(title, "Enregistrer", "Annuler", content, func(ok bool) {
			if !ok || strings.TrimSpace(nameEntry.Text) == "" {
				return
			}
			nm2, ct2 := nameEntry.Text, countryEntry.Text
			go func() {
				if isNew {
					db.Exec(`INSERT INTO brands (name, country) VALUES (?,?)`, nm2, ct2)
				} else {
					db.Exec(`UPDATE brands SET name=?, country=? WHERE id=?`, nm2, ct2, brand.ID)
				}
				fyne.Do(func() { loadBrands(); table.Refresh() })
			}()
		}, app.MainWindow)
	}

	newBtn := widget.NewButtonWithIcon("Nouvelle", theme.ContentAddIcon(), func() { showBrandForm(nil) })
	newBtn.Importance = widget.HighImportance
	editBtn := widget.NewButtonWithIcon("Modifier", theme.DocumentCreateIcon(), func() {
		if selectedRow >= 0 && selectedRow < len(brands) {
			showBrandForm(&brands[selectedRow])
		}
	})
	deleteBtn := widget.NewButtonWithIcon("Supprimer", theme.DeleteIcon(), func() {
		if selectedRow < 0 || selectedRow >= len(brands) {
			return
		}
		b := brands[selectedRow]
		dialog.ShowConfirm("Supprimer", fmt.Sprintf("Supprimer '%s' ?", b.Name), func(ok bool) {
			if !ok {
				return
			}
			bID := b.ID
			go func() {
				db.Exec(`DELETE FROM brands WHERE id=?`, bID)
				fyne.Do(func() { loadBrands(); table.Refresh() })
			}()
		}, app.MainWindow)
	})
	deleteBtn.Importance = widget.DangerImportance

	header := buildScreenHeader("🏷️ Marques", "Gérez les marques de vos produits", "#16a085")
	return container.NewBorder(
		container.NewVBox(header, container.NewHBox(newBtn, editBtn, deleteBtn)),
		nil, nil, nil, table,
	)
}

// ─────────────────────────────────────────────────────────────────────────────
// UNITÉS DE MESURE
// ─────────────────────────────────────────────────────────────────────────────

// BuildUnitsScreen construit l'écran des unités de mesure
func BuildUnitsScreen() fyne.CanvasObject {
	db := app.GetDB()

	var units []models.Unit
	selectedRow := -1

	loadUnits := func() {
		if db == nil {
			return
		}
		units, _ = queries.GetAllUnits(db)
	}
	loadUnits()

	table := widget.NewTable(
		func() (int, int) { return len(units) + 1, 3 },
		func() fyne.CanvasObject { return widget.NewLabel("") },
		func(id widget.TableCellID, cell fyne.CanvasObject) {
			lbl := cell.(*widget.Label)
			headers := []string{"Nom (FR)", "Nom (AR)", "Symbole"}
			if id.Row == 0 {
				lbl.SetText(headers[id.Col])
				lbl.TextStyle = fyne.TextStyle{Bold: true}
				return
			}
			if id.Row-1 >= len(units) {
				lbl.SetText("")
				return
			}
			u := units[id.Row-1]
			switch id.Col {
			case 0:
				lbl.SetText(u.NameFr)
			case 1:
				lbl.SetText(u.NameAr)
			case 2:
				lbl.SetText(u.Symbol)
			}
		},
	)
	table.SetColumnWidth(0, 200)
	table.SetColumnWidth(1, 200)
	table.SetColumnWidth(2, 100)
	table.OnSelected = func(id widget.TableCellID) {
		if id.Row > 0 {
			selectedRow = id.Row - 1
		}
	}

	showUnitForm := func(unit *models.Unit) {
		isNew := unit == nil
		title := "Nouvelle Unité"
		if !isNew {
			title = "Modifier: " + unit.NameFr
		}
		nameFrEntry := widget.NewEntry()
		nameArEntry := widget.NewEntry()
		symbolEntry := widget.NewEntry()
		if !isNew {
			nameFrEntry.SetText(unit.NameFr)
			nameArEntry.SetText(unit.NameAr)
			symbolEntry.SetText(unit.Symbol)
		}
		content := widget.NewForm(
			widget.NewFormItem("Nom (FR) *", nameFrEntry),
			widget.NewFormItem("Nom (AR)", nameArEntry),
			widget.NewFormItem("Symbole", symbolEntry),
		)
		dialog.ShowCustomConfirm(title, "Enregistrer", "Annuler", content, func(ok bool) {
			if !ok || strings.TrimSpace(nameFrEntry.Text) == "" {
				return
			}
			nFr3, nAr3, sym3 := nameFrEntry.Text, nameArEntry.Text, symbolEntry.Text
			go func() {
				if isNew {
					db.Exec(`INSERT INTO units (name_fr, name_ar, symbol) VALUES (?,?,?)`,
						nFr3, nAr3, sym3)
				} else {
					db.Exec(`UPDATE units SET name_fr=?, name_ar=?, symbol=? WHERE id=?`,
						nFr3, nAr3, sym3, unit.ID)
				}
				fyne.Do(func() { loadUnits(); table.Refresh() })
			}()
		}, app.MainWindow)
	}

	newBtn := widget.NewButtonWithIcon("Nouvelle", theme.ContentAddIcon(), func() { showUnitForm(nil) })
	newBtn.Importance = widget.HighImportance
	editBtn := widget.NewButtonWithIcon("Modifier", theme.DocumentCreateIcon(), func() {
		if selectedRow >= 0 && selectedRow < len(units) {
			showUnitForm(&units[selectedRow])
		}
	})

	header := buildScreenHeader("📏 Unités de Mesure", "Définissez les unités pour vos articles", "#27ae60")
	return container.NewBorder(
		container.NewVBox(header, container.NewHBox(newBtn, editBtn)),
		nil, nil, nil, table,
	)
}

// ─────────────────────────────────────────────────────────────────────────────
// INVENTAIRE
// ─────────────────────────────────────────────────────────────────────────────

// BuildInventoryScreen construit l'écran d'inventaire
func BuildInventoryScreen() fyne.CanvasObject {
	db := app.GetDB()

	var inventories []models.Inventory
	selectedRow := -1

	loadInventories := func() {
		if db == nil {
			return
		}
		svc := services.NewInventoryService(db)
		var err error
		inventories, err = svc.GetInventoryList()
		if err != nil {
			inventories = nil
		}
	}
	loadInventories()

	table := widget.NewTable(
		func() (int, int) { return len(inventories) + 1, 5 },
		func() fyne.CanvasObject { return widget.NewLabel("") },
		func(id widget.TableCellID, cell fyne.CanvasObject) {
			lbl := cell.(*widget.Label)
			headers := []string{"Date", "Type", "Statut", "Créé par", "Notes"}
			if id.Row == 0 {
				lbl.SetText(headers[id.Col])
				lbl.TextStyle = fyne.TextStyle{Bold: true}
				return
			}
			if id.Row-1 >= len(inventories) {
				lbl.SetText("")
				return
			}
			inv := inventories[id.Row-1]
			switch id.Col {
			case 0:
				lbl.SetText(utils.FormatDateFr(inv.Date))
			case 1:
				if inv.Type == "full" {
					lbl.SetText("Complet")
				} else {
					lbl.SetText("Partiel")
				}
			case 2:
				switch inv.Status {
				case "draft":
					lbl.SetText("🔄 Brouillon")
				case "confirmed":
					lbl.SetText("✅ Confirmé")
				case "cancelled":
					lbl.SetText("❌ Annulé")
				default:
					lbl.SetText(inv.Status)
				}
			case 3:
				if inv.CreatedBy != nil {
					lbl.SetText(fmt.Sprintf("user#%d", *inv.CreatedBy))
				}
			case 4:
				lbl.SetText(utils.TruncateString(inv.Notes, 50))
			}
		},
	)
	table.SetColumnWidth(0, 100)
	table.SetColumnWidth(1, 100)
	table.SetColumnWidth(2, 110)
	table.SetColumnWidth(3, 150)
	table.SetColumnWidth(4, 250)
	table.OnSelected = func(id widget.TableCellID) {
		if id.Row > 0 {
			selectedRow = id.Row - 1
		}
	}

	// Nouveau inventaire
	newBtn := widget.NewButtonWithIcon("Nouvel Inventaire", theme.ContentAddIcon(), func() {
		if db == nil {
			return
		}
		session := app.GetSession()
		userID := 1
		if session != nil {
			userID = session.UserID
		}

		typeSelect := widget.NewSelect([]string{"Complet", "Partiel"}, nil)
		typeSelect.SetSelected("Complet")

		// Sélection catégorie pour inventaire partiel
		cats, _ := queries.GetAllCategories(db)
		catNames := []string{"Toutes"}
		catMap := map[string]int{"Toutes": 0}
		for _, c := range cats {
			catNames = append(catNames, c.NameFr)
			catMap[c.NameFr] = c.ID
		}
		catSelect := widget.NewSelect(catNames, nil)
		catSelect.SetSelected("Toutes")

		content := widget.NewForm(
			widget.NewFormItem("Type", typeSelect),
			widget.NewFormItem("Catégorie (Partiel)", catSelect),
		)

		dialog.ShowCustomConfirm("Nouvel Inventaire", "Créer", "Annuler", content,
			func(ok bool) {
				if !ok {
					return
				}
				invType := "full"
				if typeSelect.Selected == "Partiel" {
					invType = "partial"
				}
				catID := catMap[catSelect.Selected]

				svc := services.NewInventoryService(db)
				_, err := svc.CreateInventory(invType, catID, userID)
				if err != nil {
					dialog.ShowError(err, app.MainWindow)
					return
				}
				loadInventories()
				table.Refresh()
				dialog.ShowInformation("Succès", "Inventaire créé ✅", app.MainWindow)
			}, app.MainWindow)
	})
	newBtn.Importance = widget.HighImportance

	// Ouvrir inventaire pour saisie
	openBtn := widget.NewButtonWithIcon("Ouvrir / Saisir", theme.DocumentCreateIcon(), func() {
		if selectedRow < 0 || selectedRow >= len(inventories) {
			dialog.ShowInformation("Info", "Sélectionnez un inventaire à ouvrir", app.MainWindow)
			return
		}
		inv := inventories[selectedRow]
		if inv.Status != "draft" {
			dialog.ShowInformation("Info", "Seuls les inventaires en brouillon peuvent être modifiés", app.MainWindow)
			return
		}
		showInventoryDetail(&inv, db, func() {
			loadInventories()
			table.Refresh()
		})
	})

	confirmBtn := widget.NewButtonWithIcon("Confirmer", theme.ConfirmIcon(), func() {
		if selectedRow < 0 || selectedRow >= len(inventories) {
			return
		}
		inv := inventories[selectedRow]
		if inv.Status != "draft" {
			dialog.ShowInformation("Info", "Seuls les brouillons peuvent être confirmés", app.MainWindow)
			return
		}
		dialog.ShowConfirm("Confirmer inventaire",
			"Confirmer cet inventaire va ajuster les stocks. Continuer?",
			func(ok bool) {
				if !ok {
					return
				}
				invID := inv.ID
				go func() {
					session := app.GetSession()
					userID := 1
					if session != nil {
						userID = session.UserID
					}
					svc := services.NewInventoryService(db)
					err := svc.ConfirmInventory(invID, userID)
					if err != nil {
						fyne.Do(func() { dialog.ShowError(err, app.MainWindow) })
						return
					}
					fyne.Do(func() {
						loadInventories()
						table.Refresh()
						dialog.ShowInformation("✅ Succès", "Inventaire confirmé, stocks ajustés!", app.MainWindow)
					})
				}()
			}, app.MainWindow)
	})
	confirmBtn.Importance = widget.HighImportance

	header := buildScreenHeader("📊 Inventaire", "Gérez vos inventaires de stock", "#e67e22")
	toolbar := container.NewHBox(newBtn, openBtn, confirmBtn)

	return container.NewBorder(
		container.NewVBox(header, toolbar),
		nil, nil, nil,
		table,
	)
}

// showInventoryDetail ouvre le formulaire de saisie d'un inventaire
func showInventoryDetail(inv *models.Inventory, db *sql.DB, onSave func()) {
	if db == nil {
		db = app.GetDB()
	}
	if db == nil {
		return
	}

	svc := services.NewInventoryService(db)
	fullInv, err := svc.GetInventory(inv.ID)
	if err != nil || fullInv == nil {
		dialog.ShowError(fmt.Errorf("impossible de charger l'inventaire: %v", err), app.MainWindow)
		return
	}

	lines := fullInv.Lines

	// Table des lignes
	lineTable := widget.NewTable(
		func() (int, int) { return len(lines) + 1, 5 },
		func() fyne.CanvasObject { return widget.NewLabel("") },
		func(id widget.TableCellID, cell fyne.CanvasObject) {
			lbl := cell.(*widget.Label)
			headers := []string{"Réf.", "Article", "Qté Théorique", "Qté Physique", "Écart"}
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
				lbl.SetText(l.Reference)
			case 1:
				lbl.SetText(l.ArticleName)
			case 2:
				lbl.SetText(fmt.Sprintf("%.2f", l.TheoreticalQty))
			case 3:
				lbl.SetText(fmt.Sprintf("%.2f", l.PhysicalQty))
			case 4:
				diff := l.Difference
				if diff > 0 {
					lbl.SetText(fmt.Sprintf("+%.2f", diff))
				} else {
					lbl.SetText(fmt.Sprintf("%.2f", diff))
				}
			}
		},
	)
	lineTable.SetColumnWidth(0, 80)
	lineTable.SetColumnWidth(1, 200)
	lineTable.SetColumnWidth(2, 120)
	lineTable.SetColumnWidth(3, 120)
	lineTable.SetColumnWidth(4, 80)

	// Saisie quantité physique
	lineTable.OnSelected = func(id widget.TableCellID) {
		if id.Row == 0 || id.Row-1 >= len(lines) {
			return
		}
		l := lines[id.Row-1]
		qtyEntry := widget.NewEntry()
		qtyEntry.SetText(fmt.Sprintf("%.2f", l.PhysicalQty))

		dialog.ShowCustomConfirm(
			fmt.Sprintf("Saisir quantité: %s", l.ArticleName),
			"Mettre à jour", "Annuler",
			container.NewVBox(
				widget.NewLabel(fmt.Sprintf("Qté théorique: %.2f", l.TheoreticalQty)),
				widget.NewForm(widget.NewFormItem("Qté physique:", qtyEntry)),
			),
			func(ok bool) {
				if !ok {
					return
				}
				qty, err := strconv.ParseFloat(strings.TrimSpace(qtyEntry.Text), 64)
				if err != nil {
					return
				}
				svc.UpdateInventoryLine(inv.ID, l.ArticleID, qty, "")
				// Recharger
				fullInv2, _ := svc.GetInventory(inv.ID)
				if fullInv2 != nil {
					lines = fullInv2.Lines
				}
				lineTable.Refresh()
			}, app.MainWindow)
	}

	// Recherche par barcode
	barcodeEntry := widget.NewEntry()
	barcodeEntry.SetPlaceHolder("Scanner code-barres...")
	addQtyEntry := widget.NewEntry()
	addQtyEntry.SetText("1")

	barcodeEntry.OnSubmitted = func(code string) {
		qty, _ := strconv.ParseFloat(strings.TrimSpace(addQtyEntry.Text), 64)
		if qty <= 0 {
			qty = 1
		}
		line, err := svc.UpdateInventoryByBarcode(inv.ID, code, qty)
		if err != nil {
			dialog.ShowError(err, app.MainWindow)
			return
		}
		barcodeEntry.SetText("")
		// Recharger lignes
		fullInv2, _ := svc.GetInventory(inv.ID)
		if fullInv2 != nil {
			lines = fullInv2.Lines
		}
		lineTable.Refresh()
		dialog.ShowInformation("✅", fmt.Sprintf("%s: %.2f", line.ArticleName, line.PhysicalQty), app.MainWindow)
	}

	barcodeBar := container.NewBorder(nil, nil,
		widget.NewLabel("Barcode:"),
		container.NewHBox(widget.NewLabel("Qté:"), addQtyEntry),
		barcodeEntry,
	)

	content := container.NewBorder(
		container.NewVBox(
			widget.NewLabel(fmt.Sprintf("Inventaire du %s — %d articles", utils.FormatDateFr(inv.Date), len(lines))),
			barcodeBar,
		),
		nil,
		nil, nil,
		lineTable,
	)

	dlg := dialog.NewCustom("📊 Saisie Inventaire", "Fermer", content, app.MainWindow)
	dlg.Resize(fyne.NewSize(700, 500))
	dlg.SetOnClosed(func() {
		onSave()
	})
	dlg.Show()
	app.MainWindow.Canvas().Focus(barcodeEntry)
}

// ─────────────────────────────────────────────────────────────────────────────
// MOUVEMENTS DE STOCK
// ─────────────────────────────────────────────────────────────────────────────

// BuildStockMovementsScreen construit l'écran des mouvements de stock
func BuildStockMovementsScreen() fyne.CanvasObject {
	db := app.GetDB()

	searchEntry := widget.NewEntry()
	searchEntry.SetPlaceHolder("🔍 Rechercher par article, référence...")

	moveTypeSelect := widget.NewSelect(
		[]string{"Tous", "Entrée achat", "Sortie vente", "Entrée retour", "Sortie retour",
			"Ajustement +", "Ajustement -", "Dommage"},
		nil,
	)
	moveTypeSelect.SetSelected("Tous")

	var movements []models.StockMovement

	typeDBMap := map[string]string{
		"Entrée achat":  "purchase_in",
		"Sortie vente":  "sale_out",
		"Entrée retour": "return_in",
		"Sortie retour": "return_out",
		"Ajustement +":  "adjustment_in",
		"Ajustement -":  "adjustment_out",
		"Dommage":       "damage",
	}

	loadMovements := func(search, moveType string) {
		if db == nil {
			return
		}
		q := `SELECT sm.id, sm.date, sm.type, sm.article_id, a.name_fr, sm.quantity, sm.unit_price,
			         COALESCE(sm.ref_doc_number,''), COALESCE(sm.notes,''), COALESCE(sm.created_by, 0)
			  FROM stock_movements sm
			  JOIN articles a ON sm.article_id = a.id
			  WHERE 1=1`
		args := []interface{}{}
		if search != "" {
			q += ` AND (a.name_fr LIKE ? OR a.reference LIKE ?)`
			s := "%" + search + "%"
			args = append(args, s, s)
		}
		if t, ok := typeDBMap[moveType]; ok {
			q += ` AND sm.type=?`
			args = append(args, t)
		}
		q += ` ORDER BY sm.date DESC, sm.id DESC LIMIT 200`

		rows, err := db.Query(q, args...)
		if err != nil {
			return
		}
		defer rows.Close()
		movements = nil
		for rows.Next() {
			var m models.StockMovement
			var createdByInt int
			rows.Scan(&m.ID, &m.Date, &m.Type, &m.ArticleID, &m.ArticleName,
				&m.Quantity, &m.UnitPrice, &m.RefDocNumber, &m.Notes, &createdByInt)
			m.CreatedBy = &createdByInt
			movements = append(movements, m)
		}
	}
	loadMovements("", "Tous")

	headers := []string{"Date", "Type", "Article", "Quantité", "P.U.", "Ref. Doc", "Créé par"}
	colWidths := []float32{90, 130, 200, 80, 100, 120, 120}

	table := widget.NewTable(
		func() (int, int) { return len(movements) + 1, len(headers) },
		func() fyne.CanvasObject { return widget.NewLabel("") },
		func(id widget.TableCellID, cell fyne.CanvasObject) {
			lbl := cell.(*widget.Label)
			if id.Row == 0 {
				lbl.SetText(headers[id.Col])
				lbl.TextStyle = fyne.TextStyle{Bold: true}
				return
			}
			if id.Row-1 >= len(movements) {
				lbl.SetText("")
				return
			}
			m := movements[id.Row-1]
			switch id.Col {
			case 0:
				lbl.SetText(utils.FormatDateFr(m.Date))
			case 1:
				lbl.SetText(utils.StockMoveLabel(m.Type))
			case 2:
				lbl.SetText(m.ArticleName)
			case 3:
				lbl.SetText(fmt.Sprintf("%.2f", m.Quantity))
			case 4:
				lbl.SetText(utils.FormatMoney(m.UnitPrice))
			case 5:
				lbl.SetText(m.RefDocNumber)
			case 6:
				if m.CreatedBy != nil {
					lbl.SetText(fmt.Sprintf("user#%d", *m.CreatedBy))
				}
			}
		},
	)
	for i, w := range colWidths {
		table.SetColumnWidth(i, w)
	}

	countLabel := widget.NewLabel(fmt.Sprintf("%d mouvement(s)", len(movements)))

	searchEntry.OnChanged = func(s string) {
		loadMovements(s, moveTypeSelect.Selected)
		table.Refresh()
		countLabel.SetText(fmt.Sprintf("%d mouvement(s)", len(movements)))
	}
	moveTypeSelect.OnChanged = func(s string) {
		loadMovements(searchEntry.Text, s)
		table.Refresh()
		countLabel.SetText(fmt.Sprintf("%d mouvement(s)", len(movements)))
	}

	// Ajustement manuel
	adjustBtn := widget.NewButtonWithIcon("Ajustement Manuel", theme.ContentAddIcon(), func() {
		if db == nil {
			return
		}
		showManualStockAdjustment(db, func() {
			loadMovements(searchEntry.Text, moveTypeSelect.Selected)
			table.Refresh()
			countLabel.SetText(fmt.Sprintf("%d mouvement(s)", len(movements)))
		})
	})

	header := buildScreenHeader("🔄 Mouvements de Stock",
		"Historique complet des entrées et sorties de stock", "#16a085")

	filters := container.NewGridWithColumns(3,
		container.NewBorder(nil, nil, widget.NewLabel("Recherche: "), nil, searchEntry),
		container.NewBorder(nil, nil, widget.NewLabel("Type: "), nil, moveTypeSelect),
		container.NewHBox(adjustBtn, widget.NewSeparator(), countLabel),
	)

	return container.NewBorder(
		container.NewVBox(header, filters),
		nil, nil, nil,
		table,
	)
}

// showManualStockAdjustment affiche le formulaire d'ajustement de stock
func showManualStockAdjustment(db *sql.DB, onSave func()) {
	if db == nil {
		return
	}

	// Sélection article
	articles, _ := queries.GetAllArticles(db, "", 0, "")
	artNames := make([]string, len(articles))
	for i, a := range articles {
		artNames[i] = fmt.Sprintf("[%s] %s (stock: %.2f)", a.Reference, a.NameFr, a.StockQty)
	}
	if len(artNames) == 0 {
		artNames = []string{"Aucun article"}
	}

	artSelect := widget.NewSelect(artNames, nil)
	if len(artNames) > 0 {
		artSelect.SetSelected(artNames[0])
	}

	adjTypeSelect := widget.NewSelect([]string{"Entrée (+)", "Sortie (-)"}, nil)
	adjTypeSelect.SetSelected("Entrée (+)")

	qtyEntry := widget.NewEntry()
	qtyEntry.SetPlaceHolder("0.00")

	notesEntry := widget.NewEntry()
	notesEntry.SetPlaceHolder("Motif de l'ajustement...")

	content := widget.NewForm(
		widget.NewFormItem("Article *", artSelect),
		widget.NewFormItem("Type", adjTypeSelect),
		widget.NewFormItem("Quantité *", qtyEntry),
		widget.NewFormItem("Notes", notesEntry),
	)

	dialog.ShowCustomConfirm("Ajustement Manuel de Stock", "Confirmer", "Annuler", content,
		func(ok bool) {
			if !ok {
				return
			}
			qty, err := strconv.ParseFloat(strings.TrimSpace(qtyEntry.Text), 64)
			if err != nil || qty <= 0 {
				dialog.ShowError(fmt.Errorf("quantité invalide"), app.MainWindow)
				return
			}
			idx := artSelect.SelectedIndex()
			if idx < 0 || idx >= len(articles) {
				return
			}
			art := articles[idx]
			moveType := "adjustment_in"
			if adjTypeSelect.Selected == "Sortie (-)" {
				moveType = "adjustment_out"
			}
			session := app.GetSession()
			createdBy := "admin"
			if session != nil {
				createdBy = session.Username
			}
			notes := notesEntry.Text
			go func() {
				db.Exec(`INSERT INTO stock_movements (date, type, article_id, quantity, unit_price, notes, created_by)
					VALUES (date('now'), ?, ?, ?, ?, ?, ?)`,
					moveType, art.ID, qty, art.CMUP, notes, createdBy)
				if moveType == "adjustment_in" {
					db.Exec(`UPDATE articles SET stock_qty = stock_qty + ? WHERE id=?`, qty, art.ID)
				} else {
					db.Exec(`UPDATE articles SET stock_qty = stock_qty - ? WHERE id=?`, qty, art.ID)
				}
				fyne.Do(func() {
					onSave()
					dialog.ShowInformation("✅ Succès", fmt.Sprintf("Ajustement de %.2f appliqué!", qty), app.MainWindow)
				})
			}()
		}, app.MainWindow)
}

// ─────────────────────────────────────────────────────────────────────────────
// DÉPÔTS / MAGASINS
// ─────────────────────────────────────────────────────────────────────────────

// BuildWarehousesScreen construit l'écran des dépôts
func BuildWarehousesScreen() fyne.CanvasObject {
	db := app.GetDB()

	var warehouses []models.Warehouse
	selectedRow := -1

	loadWarehouses := func() {
		if db == nil {
			return
		}
		warehouses, _ = queries.GetAllWarehouses(db)
	}
	loadWarehouses()

	table := widget.NewTable(
		func() (int, int) { return len(warehouses) + 1, 5 },
		func() fyne.CanvasObject { return widget.NewLabel("") },
		func(id widget.TableCellID, cell fyne.CanvasObject) {
			lbl := cell.(*widget.Label)
			headers := []string{"Nom", "Adresse", "Responsable", "Nb Articles", "Valeur Stock"}
			if id.Row == 0 {
				lbl.SetText(headers[id.Col])
				lbl.TextStyle = fyne.TextStyle{Bold: true}
				return
			}
			if id.Row-1 >= len(warehouses) {
				lbl.SetText("")
				return
			}
			w := warehouses[id.Row-1]
			switch id.Col {
			case 0:
				lbl.SetText(w.Name)
			case 1:
				lbl.SetText(w.Address)
			case 2:
				lbl.SetText(w.Manager)
			case 3:
				lbl.SetText(fmt.Sprintf("%d", w.ProductCount))
			case 4:
				lbl.SetText(utils.FormatMoney(w.TotalValue))
			}
		},
	)
	table.SetColumnWidth(0, 150)
	table.SetColumnWidth(1, 200)
	table.SetColumnWidth(2, 150)
	table.SetColumnWidth(3, 100)
	table.SetColumnWidth(4, 130)
	table.OnSelected = func(id widget.TableCellID) {
		if id.Row > 0 {
			selectedRow = id.Row - 1
		}
	}

	showWhForm := func(wh *models.Warehouse) {
		isNew := wh == nil
		title := "Nouveau Dépôt"
		if !isNew {
			title = "Modifier: " + wh.Name
		}
		nameEntry := widget.NewEntry()
		addrEntry := widget.NewEntry()
		mgmtEntry := widget.NewEntry()
		if !isNew {
			nameEntry.SetText(wh.Name)
			addrEntry.SetText(wh.Address)
			mgmtEntry.SetText(wh.Manager)
		}
		content := widget.NewForm(
			widget.NewFormItem("Nom *", nameEntry),
			widget.NewFormItem("Adresse", addrEntry),
			widget.NewFormItem("Responsable", mgmtEntry),
		)
		dialog.ShowCustomConfirm(title, "Enregistrer", "Annuler", content, func(ok bool) {
			if !ok || strings.TrimSpace(nameEntry.Text) == "" {
				return
			}
			nm4, addr4, mgr4 := nameEntry.Text, addrEntry.Text, mgmtEntry.Text
			go func() {
				if isNew {
					db.Exec(`INSERT INTO warehouses (name, address, manager) VALUES (?,?,?)`,
						nm4, addr4, mgr4)
				} else {
					db.Exec(`UPDATE warehouses SET name=?, address=?, manager=? WHERE id=?`,
						nm4, addr4, mgr4, wh.ID)
				}
				fyne.Do(func() { loadWarehouses(); table.Refresh() })
			}()
		}, app.MainWindow)
	}

	newBtn := widget.NewButtonWithIcon("Nouveau", theme.ContentAddIcon(), func() { showWhForm(nil) })
	newBtn.Importance = widget.HighImportance
	editBtn := widget.NewButtonWithIcon("Modifier", theme.DocumentCreateIcon(), func() {
		if selectedRow >= 0 && selectedRow < len(warehouses) {
			showWhForm(&warehouses[selectedRow])
		}
	})
	deleteBtn := widget.NewButtonWithIcon("Supprimer", theme.DeleteIcon(), func() {
		if selectedRow < 0 || selectedRow >= len(warehouses) {
			return
		}
		w := warehouses[selectedRow]
		dialog.ShowConfirm("Supprimer", fmt.Sprintf("Supprimer '%s' ?", w.Name), func(ok bool) {
			if !ok {
				return
			}
			wID := w.ID
			go func() {
				db.Exec(`DELETE FROM warehouses WHERE id=?`, wID)
				fyne.Do(func() { loadWarehouses(); table.Refresh() })
			}()
		}, app.MainWindow)
	})
	deleteBtn.Importance = widget.DangerImportance

	header := buildScreenHeader("🏬 Dépôts / Magasins",
		"Gérez vos points de stockage et leurs inventaires", "#2c3e50")

	return container.NewBorder(
		container.NewVBox(header, container.NewHBox(newBtn, editBtn, deleteBtn)),
		nil, nil, nil,
		table,
	)
}

// ─────────────────────────────────────────────────────────────────────────────
// LISTES DE PRIX
// ─────────────────────────────────────────────────────────────────────────────

// BuildPriceListsScreen construit l'écran des listes de prix
func BuildPriceListsScreen() fyne.CanvasObject {
	db := app.GetDB()

	var priceLists []models.PriceList
	selectedRow := -1

	loadPriceLists := func() {
		if db == nil {
			return
		}
		priceLists, _ = queries.GetAllPriceLists(db)
	}
	loadPriceLists()

	table := widget.NewTable(
		func() (int, int) { return len(priceLists) + 1, 2 },
		func() fyne.CanvasObject { return widget.NewLabel("") },
		func(id widget.TableCellID, cell fyne.CanvasObject) {
			lbl := cell.(*widget.Label)
			headers := []string{"Nom", "Description"}
			if id.Row == 0 {
				lbl.SetText(headers[id.Col])
				lbl.TextStyle = fyne.TextStyle{Bold: true}
				return
			}
			if id.Row-1 >= len(priceLists) {
				lbl.SetText("")
				return
			}
			pl := priceLists[id.Row-1]
			switch id.Col {
			case 0:
				lbl.SetText(pl.Name)
			case 1:
				lbl.SetText(pl.Description)
			}
		},
	)
	table.SetColumnWidth(0, 200)
	table.SetColumnWidth(1, 400)
	table.OnSelected = func(id widget.TableCellID) {
		if id.Row > 0 {
			selectedRow = id.Row - 1
		}
	}

	showPLForm := func(pl *models.PriceList) {
		isNew := pl == nil
		title := "Nouvelle Liste de Prix"
		if !isNew {
			title = "Modifier: " + pl.Name
		}
		nameEntry := widget.NewEntry()
		descEntry := widget.NewEntry()
		if !isNew {
			nameEntry.SetText(pl.Name)
			descEntry.SetText(pl.Description)
		}
		content := widget.NewForm(
			widget.NewFormItem("Nom *", nameEntry),
			widget.NewFormItem("Description", descEntry),
		)
		dialog.ShowCustomConfirm(title, "Enregistrer", "Annuler", content, func(ok bool) {
			if !ok || strings.TrimSpace(nameEntry.Text) == "" {
				return
			}
			nm5, desc5 := nameEntry.Text, descEntry.Text
			go func() {
				if isNew {
					db.Exec(`INSERT INTO price_lists (name, description) VALUES (?,?)`, nm5, desc5)
				} else {
					db.Exec(`UPDATE price_lists SET name=?, description=? WHERE id=?`, nm5, desc5, pl.ID)
				}
				fyne.Do(func() { loadPriceLists(); table.Refresh() })
			}()
		}, app.MainWindow)
	}

	newBtn := widget.NewButtonWithIcon("Nouvelle", theme.ContentAddIcon(), func() { showPLForm(nil) })
	newBtn.Importance = widget.HighImportance
	editBtn := widget.NewButtonWithIcon("Modifier", theme.DocumentCreateIcon(), func() {
		if selectedRow >= 0 && selectedRow < len(priceLists) {
			showPLForm(&priceLists[selectedRow])
		}
	})

	header := buildScreenHeader("💲 Listes de Prix",
		"Configurez des grilles tarifaires pour différents clients", "#f39c12")

	return container.NewBorder(
		container.NewVBox(header, container.NewHBox(newBtn, editBtn)),
		nil, nil, nil,
		table,
	)
}

// ─────────────────────────────────────────────────────────────────────────────
// buildScreenHeader — En-tête coloré réutilisable
// ─────────────────────────────────────────────────────────────────────────────

// buildScreenHeader crée un en-tête coloré pour les écrans
func buildScreenHeader(title, subtitle, hexColor string) fyne.CanvasObject {
	bg := canvas.NewRectangle(parseHexColor(hexColor))
	bg.CornerRadius = 6

	titleLabel := canvas.NewText(title, color.White)
	titleLabel.TextSize = 16
	titleLabel.TextStyle = fyne.TextStyle{Bold: true}

	subtitleLabel := canvas.NewText(subtitle, color.RGBA{220, 220, 220, 255})
	subtitleLabel.TextSize = 12

	content := container.NewVBox(
		container.NewPadded(titleLabel),
		container.NewPadded(subtitleLabel),
	)

	return container.NewStack(bg, container.NewPadded(content))
}
