package screens

import (
	"fmt"
	"gestion-commerciale/internal/app"
	"gestion-commerciale/internal/database/queries"
	"image/color"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// BuildTopToolbar construit la barre d'outils supérieure avec un design moderne
func BuildTopToolbar(w fyne.Window, navigate func(route string, params ...interface{})) fyne.CanvasObject {
	session := app.GetSession()

	// Fond de la toolbar
	toolbarBg := canvas.NewRectangle(color.RGBA{R: 0x1a, G: 0x3a, B: 0x6e, A: 0xff})

	// Informations session (côté gauche)
	companyName := "Mon Commerce"
	fiscalYear := "2026"
	userName := "Administrateur"
	userRole := "admin"

	if session != nil {
		companyName = session.CompanyName
		fiscalYear = fmt.Sprintf("%d", session.FiscalYear)
		userName = session.FullName
		userRole = session.Role
	}

	companyLbl := canvas.NewText("🏪  "+companyName, color.White)
	companyLbl.TextSize = 14
	companyLbl.TextStyle = fyne.TextStyle{Bold: true}

	yearLbl := canvas.NewText("📅  Exercice "+fiscalYear, color.RGBA{R: 0xd0, G: 0xdc, B: 0xf4, A: 0xff})
	yearLbl.TextSize = 12

	leftInfo := container.NewHBox(
		container.NewPadded(companyLbl),
		container.NewPadded(yearLbl),
	)

	// Boutons d'action (côté droit)
	newInvoiceBtn := widget.NewButtonWithIcon("Nouvelle Facture", theme.DocumentCreateIcon(), func() {
		navigate("sales/invoice")
	})
	newInvoiceBtn.Importance = widget.HighImportance

	posBtn := widget.NewButtonWithIcon("POS", theme.ComputerIcon(), func() {
		navigate("pos")
	})

	notifBtn := widget.NewButtonWithIcon("🔔 Alertes", theme.InfoIcon(), func() {
		db := app.GetDB()
		if db == nil {
			return
		}
		go func() {
			stats := queries.GetDashboardStats(db)
			msg := fmt.Sprintf(
				"⚠️  Articles sous seuil: %d\n"+
					"👤  Créances clients: %s DA\n"+
					"🏭  Dettes fournisseurs: %s DA",
				stats.LowStockCount,
				fmt.Sprintf("%.2f", stats.ClientDebts),
				fmt.Sprintf("%.2f", stats.SupplierDebts),
			)
			fyne.Do(func() {
				dialog.ShowInformation("🔔 Notifications", msg, w)
			})
		}()
	})

	userLbl := canvas.NewText("👤  "+userName+" ("+userRole+")", color.RGBA{R: 0xf0, G: 0xb4, B: 0x29, A: 0xff})
	userLbl.TextSize = 12

	logoutBtn := widget.NewButtonWithIcon("Déconnecter", theme.LogoutIcon(), func() {
		dialog.ShowConfirm("Déconnexion",
			"Voulez-vous vous déconnecter?",
			func(ok bool) {
				if ok {
					db := app.GetDB()
					if db != nil {
						db.Close()
					}
					w.Close()
				}
			}, w)
	})
	logoutBtn.Importance = widget.DangerImportance

	rightInfo := container.NewHBox(
		container.NewPadded(userLbl),
		notifBtn,
		newInvoiceBtn,
		posBtn,
		logoutBtn,
	)

	toolbarContent := container.NewBorder(nil, nil, leftInfo, rightInfo)

	return container.NewStack(
		toolbarBg,
		container.NewPadded(toolbarContent),
	)
}

// BuildStatusBar construit la barre de statut inférieure moderne
func BuildStatusBar() fyne.CanvasObject {
	statusBg := canvas.NewRectangle(color.RGBA{R: 0x0d, G: 0x21, B: 0x42, A: 0xff})

	dateTimeLabel := canvas.NewText(time.Now().Format("📅  02/01/2006   🕐  15:04:05"), color.RGBA{R: 0xb8, G: 0xcc, B: 0xee, A: 0xff})
	dateTimeLabel.TextSize = 11

	// Mise à jour toutes les secondes en arrière-plan
	go func() {
		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()
		for range ticker.C {
			t := time.Now().Format("📅  02/01/2006   🕐  15:04:05")
			fyne.Do(func() { dateTimeLabel.Text = t; dateTimeLabel.Refresh() })
		}
	}()

	dbStatus := "✅  Base de données connectée"
	if session := app.GetSession(); session != nil && session.DBPath != "" {
		dbStatus = "✅  " + session.DBPath
	}

	dbLbl := canvas.NewText(dbStatus, color.RGBA{R: 0x7a, G: 0xb4, B: 0x7a, A: 0xff})
	dbLbl.TextSize = 11

	versionLbl := canvas.NewText("Gestion Commerciale Pro  v2.0", color.RGBA{R: 0x7a, G: 0x9a, B: 0xcc, A: 0xff})
	versionLbl.TextSize = 11

	barContent := container.NewBorder(
		nil, nil,
		container.NewHBox(container.NewPadded(dateTimeLabel), container.NewPadded(dbLbl)),
		container.NewPadded(versionLbl),
	)

	return container.NewStack(statusBg, container.NewPadded(barContent))
}
