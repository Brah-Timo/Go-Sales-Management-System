package screens

import (
	"fmt"
	"gestion-commerciale/internal/app"
	"gestion-commerciale/internal/database"
	"gestion-commerciale/internal/models"
	"gestion-commerciale/internal/services"
	"gestion-commerciale/pkg/utils"
	"image/color"
	"path/filepath"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	fyneapp "fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// OnLoginSuccess callback après connexion réussie
type OnLoginSuccess func(dbPath string, result *app.LoginResult)

// BuildLoginScreen construit l'écran de connexion moderne
func BuildLoginScreen(w fyne.Window, a fyne.App, onSuccess OnLoginSuccess) fyne.CanvasObject {

	// ── Panneau gauche (décoratif) ────────────────────────────────────────────
	leftBg := canvas.NewRectangle(color.RGBA{R: 0x1a, G: 0x3a, B: 0x6e, A: 0xff})

	logoCircle := canvas.NewCircle(color.RGBA{R: 0xf0, G: 0xb4, B: 0x29, A: 0xff})
	logoCircle.Resize(fyne.NewSize(80, 80))

	logoText := canvas.NewText("G", color.White)
	logoText.TextSize = 48
	logoText.TextStyle = fyne.TextStyle{Bold: true}

	logoStack := container.NewStack(
		container.NewCenter(logoCircle),
		container.NewCenter(logoText),
	)
	logoStack.Resize(fyne.NewSize(80, 80))

	appTitle := canvas.NewText("Gestion Commerciale Pro", color.White)
	appTitle.TextSize = 22
	appTitle.TextStyle = fyne.TextStyle{Bold: true}

	appSubtitle := canvas.NewText("Système de Gestion Algérien", color.RGBA{R: 0xb8, G: 0xcc, B: 0xee, A: 0xff})
	appSubtitle.TextSize = 13

	sep1 := canvas.NewLine(color.RGBA{R: 0xf0, G: 0xb4, B: 0x29, A: 0x88})
	sep1.StrokeWidth = 2

	feat1 := makeFeatureItem("📦", "Gestion Articles & Stock")
	feat2 := makeFeatureItem("🧾", "Facturation Ventes & Achats")
	feat3 := makeFeatureItem("💰", "Trésorerie & Finances")
	feat4 := makeFeatureItem("📊", "Rapports & Statistiques")
	feat5 := makeFeatureItem("👥", "Clients & Fournisseurs")

	version := canvas.NewText("v2.0 — Conforme aux normes DGI", color.RGBA{R: 0x7a, G: 0x9a, B: 0xcc, A: 0xff})
	version.TextSize = 11

	leftContent := container.NewVBox(
		container.NewCenter(logoStack),
		container.NewCenter(appTitle),
		container.NewCenter(appSubtitle),
		container.NewPadded(sep1),
		feat1, feat2, feat3, feat4, feat5,
		widget.NewLabel(""),
		container.NewCenter(version),
	)

	leftPanel := container.NewStack(leftBg, container.NewPadded(leftContent))

	// ── Panneau droit (formulaire) ─────────────────────────────────────────────
	rightBg := canvas.NewRectangle(color.RGBA{R: 0xf8, G: 0xfa, B: 0xfd, A: 0xff})

	formTitle := canvas.NewText("Connexion", color.RGBA{R: 0x1a, G: 0x3a, B: 0x6e, A: 0xff})
	formTitle.TextSize = 26
	formTitle.TextStyle = fyne.TextStyle{Bold: true}

	formSubtitle := canvas.NewText("Entrez vos identifiants pour accéder au système", color.RGBA{R: 0x5a, G: 0x6a, B: 0x85, A: 0xff})
	formSubtitle.TextSize = 12

	// Liste des sociétés
	dataDir := "data"
	companies := database.GetDatabaseFiles(dataDir)
	if len(companies) == 0 {
		database.InitDatabase(filepath.Join(dataDir, "default_company.db"))
		companies = database.GetDatabaseFiles(dataDir)
	}
	companyNames := companies
	if len(companyNames) == 0 {
		companyNames = []string{"default_company"}
	}
	companySelect := widget.NewSelect(companyNames, nil)
	if len(companyNames) > 0 {
		companySelect.SetSelected(companyNames[0])
	}

	// Années
	currentYear := utils.CurrentYear()
	years := []string{
		fmt.Sprintf("%d", currentYear-1),
		fmt.Sprintf("%d", currentYear),
		fmt.Sprintf("%d", currentYear+1),
	}
	yearSelect := widget.NewSelect(years, nil)
	yearSelect.SetSelected(fmt.Sprintf("%d", currentYear))

	// Champs
	usernameEntry := widget.NewEntry()
	usernameEntry.SetPlaceHolder("Nom d'utilisateur")

	passwordEntry := widget.NewPasswordEntry()
	passwordEntry.SetPlaceHolder("Mot de passe")

	// Message d'état
	statusLabel := widget.NewLabel("")
	statusLabel.Alignment = fyne.TextAlignCenter
	statusLabel.Importance = widget.DangerImportance

	// Compteur de tentatives
	loginAttempts := 0
	lockUntil := time.Time{}

	var loginButton *widget.Button

	loginButton = widget.NewButtonWithIcon("  Se Connecter  ", theme.LoginIcon(), func() {
		if !lockUntil.IsZero() && time.Now().Before(lockUntil) {
			remaining := int(time.Until(lockUntil).Seconds())
			statusLabel.SetText(fmt.Sprintf("⛔ Verrouillé — réessayez dans %d s", remaining))
			return
		}
		if !lockUntil.IsZero() && time.Now().After(lockUntil) {
			loginAttempts = 0
			lockUntil = time.Time{}
			statusLabel.SetText("")
		}

		username := strings.TrimSpace(usernameEntry.Text)
		password := passwordEntry.Text
		companyFile := companySelect.Selected

		if username == "" || password == "" {
			statusLabel.SetText("⚠️ Saisissez le nom d'utilisateur et le mot de passe")
			return
		}

		statusLabel.SetText("⏳ Connexion en cours...")
		loginButton.Disable()

		dbPath := filepath.Join(dataDir, companyFile+".db")
		selectedYear := yearSelect.Selected

		go func() {
			db, err := database.InitDatabase(dbPath)
			if err != nil {
				fyne.Do(func() {
					statusLabel.SetText("❌ Impossible d'ouvrir la base de données")
					loginButton.Enable()
				})
				return
			}

			authSvc := services.NewAuthService(db)
			user, err := authSvc.Login(username, password)
			if err != nil {
				db.Close()
				fyne.Do(func() {
					loginButton.Enable()
					loginAttempts++
					if loginAttempts >= 3 {
						lockUntil = time.Now().Add(30 * time.Second)
						statusLabel.SetText("🔒 Trop de tentatives. Verrouillé 30 secondes.")
						go func() {
							time.Sleep(31 * time.Second)
							fyne.Do(func() {
								loginAttempts = 0
								lockUntil = time.Time{}
								statusLabel.SetText("🔓 Verrouillage levé. Vous pouvez réessayer.")
							})
						}()
					} else {
						statusLabel.SetText(fmt.Sprintf("❌ Identifiants incorrects (%d/3)", loginAttempts))
					}
				})
				return
			}

			var companyName string
			db.QueryRow(`SELECT COALESCE(name_fr,'Mon Commerce') FROM companies WHERE id=1`).Scan(&companyName)
			db.Close()

			var year int
			fmt.Sscanf(selectedYear, "%d", &year)

			session := services.BuildSession(user, year, dbPath, companyName)
			result := &app.LoginResult{Session: session, DBPath: dbPath}

			fyne.Do(func() {
				loginAttempts = 0
				lockUntil = time.Time{}
				statusLabel.SetText("")
				loginButton.Enable()
				onSuccess(dbPath, result)
			})
		}()
	})
	loginButton.Importance = widget.HighImportance

	// Touche Entrée
	passwordEntry.OnSubmitted = func(s string) { loginButton.OnTapped() }
	usernameEntry.OnSubmitted = func(s string) { w.Canvas().Focus(passwordEntry) }

	// Bouton Quitter
	cancelButton := widget.NewButtonWithIcon("Quitter", theme.CancelIcon(), func() { a.Quit() })

	// Bouton Nouvelle Société
	newCompanyButton := widget.NewButtonWithIcon("Nouvelle Société", theme.FolderNewIcon(), func() {
		showNewCompanyDialog(w, dataDir, func(name string) {
			companies := database.GetDatabaseFiles(dataDir)
			companySelect.Options = companies
			if len(companies) > 0 {
				companySelect.SetSelected(companies[len(companies)-1])
			}
			companySelect.Refresh()
		})
	})

	// Bouton Nouvel Utilisateur
	newUserButton := widget.NewButtonWithIcon("Nouvel Utilisateur", theme.AccountIcon(), func() {
		showNewUserOnLoginDialog(w, dataDir, companySelect.Selected)
	})

	// Séparateur décoratif
	formSep := canvas.NewLine(color.RGBA{R: 0x1a, G: 0x3a, B: 0x6e, A: 0x44})
	formSep.StrokeWidth = 1

	demoInfo := widget.NewLabel("💡 Connexion démo : admin / admin123")
	demoInfo.Alignment = fyne.TextAlignCenter
	demoInfo.TextStyle = fyne.TextStyle{Italic: true}

	// Formulaire
	form := widget.NewForm(
		widget.NewFormItem("🏢  Société :", companySelect),
		widget.NewFormItem("📅  Année :", yearSelect),
		widget.NewFormItem("👤  Utilisateur :", usernameEntry),
		widget.NewFormItem("🔑  Mot de passe :", passwordEntry),
	)

	bottomRow := container.NewGridWithColumns(3, cancelButton, newCompanyButton, newUserButton)

	rightContent := container.NewVBox(
		widget.NewLabel(""),
		container.NewCenter(formTitle),
		container.NewCenter(formSubtitle),
		widget.NewLabel(""),
		container.NewPadded(formSep),
		container.NewPadded(form),
		statusLabel,
		widget.NewLabel(""),
		container.NewCenter(loginButton),
		widget.NewLabel(""),
		widget.NewSeparator(),
		bottomRow,
		widget.NewSeparator(),
		container.NewCenter(demoInfo),
		widget.NewLabel(""),
	)

	rightPanel := container.NewStack(
		rightBg,
		container.NewPadded(rightContent),
	)

	// ── Split gauche / droite ──────────────────────────────────────────────────
	split := container.NewHSplit(leftPanel, rightPanel)
	split.SetOffset(0.40)

	return split
}

// makeFeatureItem crée un élément de liste de fonctionnalité pour le panneau gauche
func makeFeatureItem(icon, label string) fyne.CanvasObject {
	icn := canvas.NewText(icon+"  ", color.RGBA{R: 0xf0, G: 0xb4, B: 0x29, A: 0xff})
	icn.TextSize = 14

	lbl := canvas.NewText(label, color.RGBA{R: 0xd0, G: 0xdc, B: 0xf4, A: 0xff})
	lbl.TextSize = 13

	return container.NewHBox(icn, lbl)
}

// showNewUserOnLoginDialog — crée un utilisateur directement depuis l'écran de login
func showNewUserOnLoginDialog(w fyne.Window, dataDir, selectedCompany string) {
	if selectedCompany == "" {
		dialog.ShowInformation("Attention", "Sélectionnez d'abord une société.", w)
		return
	}
	dbPath := filepath.Join(dataDir, selectedCompany+".db")
	db, err := database.InitDatabase(dbPath)
	if err != nil {
		dialog.ShowError(fmt.Errorf("impossible d'ouvrir la base: %v", err), w)
		return
	}

	// Champs du formulaire
	userEntry := widget.NewEntry()
	userEntry.SetPlaceHolder("ex: vendeur1")
	fullNameEntry := widget.NewEntry()
	fullNameEntry.SetPlaceHolder("ex: Ahmed Benali")
	passEntry := widget.NewPasswordEntry()
	passEntry.SetPlaceHolder("Mot de passe")
	pass2Entry := widget.NewPasswordEntry()
	pass2Entry.SetPlaceHolder("Confirmer le mot de passe")
	roleSelect := widget.NewSelect([]string{"admin", "seller", "cashier", "assistant"}, nil)
	roleSelect.SetSelected("seller")

	adminPassEntry := widget.NewPasswordEntry()
	adminPassEntry.SetPlaceHolder("Mot de passe admin pour autoriser")

	items := []*widget.FormItem{
		widget.NewFormItem("Nom d'utilisateur *", userEntry),
		widget.NewFormItem("Nom complet *", fullNameEntry),
		widget.NewFormItem("Rôle", roleSelect),
		widget.NewFormItem("Mot de passe *", passEntry),
		widget.NewFormItem("Confirmer mot de passe *", pass2Entry),
		widget.NewFormItem("─── Autorisation admin ───", widget.NewLabel("")),
		widget.NewFormItem("Mot de passe admin *", adminPassEntry),
	}

	dialog.ShowForm("➕ Créer un nouvel utilisateur", "Créer", "Annuler", items,
		func(ok bool) {
			if !ok {
				db.Close()
				return
			}

			username := strings.TrimSpace(userEntry.Text)
			fullName := strings.TrimSpace(fullNameEntry.Text)
			pass := passEntry.Text
			pass2 := pass2Entry.Text
			adminPass := adminPassEntry.Text
			role := roleSelect.Selected

			if username == "" || fullName == "" || pass == "" {
				dialog.ShowError(fmt.Errorf("tous les champs obligatoires (*) doivent être remplis"), w)
				db.Close()
				return
			}
			if pass != pass2 {
				dialog.ShowError(fmt.Errorf("les mots de passe ne correspondent pas"), w)
				db.Close()
				return
			}
			if len(pass) < 4 {
				dialog.ShowError(fmt.Errorf("mot de passe trop court (minimum 4 caractères)"), w)
				db.Close()
				return
			}

			go func() {
				authSvc := services.NewAuthService(db)
				adminUser, err := authSvc.Login("admin", adminPass)
				if err != nil || adminUser == nil {
					fyne.Do(func() {
						dialog.ShowError(fmt.Errorf("mot de passe admin incorrect — création refusée"), w)
						db.Close()
					})
					return
				}

				perms := models.DefaultPermissionsByRole(role)
				err = authSvc.CreateUser(username, fullName, pass, role, perms)
				db.Close()
				fyne.Do(func() {
					if err != nil {
						dialog.ShowError(fmt.Errorf("erreur création: %v", err), w)
						return
					}
					dialog.ShowInformation("✅ Succès",
						fmt.Sprintf("Utilisateur '%s' créé avec succès!\nRôle: %s", username, role), w)
				})
			}()
		}, w)
}

// showNewCompanyDialog — crée une nouvelle société
func showNewCompanyDialog(w fyne.Window, dataDir string, onCreated func(name string)) {
	nameEntry := widget.NewEntry()
	nameEntry.SetPlaceHolder("Nom de la société (ex: Ma_Boutique)")

	dialog.ShowForm("Nouvelle Société", "Créer", "Annuler",
		[]*widget.FormItem{
			widget.NewFormItem("Nom:", nameEntry),
		},
		func(ok bool) {
			if !ok || strings.TrimSpace(nameEntry.Text) == "" {
				return
			}
			name := strings.ReplaceAll(strings.TrimSpace(nameEntry.Text), " ", "_")
			dbPath := filepath.Join(dataDir, name+".db")
			go func() {
				_, err := database.InitDatabase(dbPath)
				fyne.Do(func() {
					if err != nil {
						dialog.ShowError(fmt.Errorf("impossible de créer: %v", err), w)
						return
					}
					dialog.ShowInformation("✅ Succès", fmt.Sprintf("Société '%s' créée!", name), w)
					onCreated(name)
				})
			}()
		}, w)
}

// Éviter import inutilisé
var _ = fyneapp.New
