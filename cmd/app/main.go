package main

import (
	"fmt"
	_ "embed"
	"log"
	"os"
	"path/filepath"
	"time"

	"fyne.io/fyne/v2"
	fyneapp "fyne.io/fyne/v2/app"

	appstate "gestion-commerciale/internal/app"
	"gestion-commerciale/internal/database"
	"gestion-commerciale/ui/screens"
	gctheme "gestion-commerciale/ui/theme"
)

//go:embed icon.png
var iconBytes []byte

func main() {
	// Configurer le log
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.SetOutput(os.Stdout)

	// Créer les répertoires nécessaires
	dirs := []string{"data", "data/backups", "assets/fonts", "assets/icons"}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			log.Printf("Avertissement création répertoire %s: %v", dir, err)
		}
	}

	// Créer l'application Fyne
	a := fyneapp.New()
	// Appliquer le thème personnalisé avec police Cairo (support arabe)
	a.Settings().SetTheme(&gctheme.GestionTheme{})

	// Définir l'icône de l'application (visible dans la barre des tâches et l'explorateur)
	appIcon := fyne.NewStaticResource("icon.png", iconBytes)
	a.SetIcon(appIcon)

	// Créer la fenêtre principale (cachée au départ)
	mainWin := a.NewWindow("Gestion Commerciale Pro")
	mainWin.Resize(fyne.NewSize(1280, 768))
	mainWin.SetIcon(appIcon)
	mainWin.SetMaster()
	appstate.MainWindow = mainWin

	// Créer la fenêtre de login
	loginWin := a.NewWindow("Connexion — Gestion Commerciale Pro")
	loginWin.Resize(fyne.NewSize(480, 600))
	loginWin.SetFixedSize(true)
	loginWin.SetIcon(appIcon)
	loginWin.CenterOnScreen()

	// Fonction de rappel après connexion réussie
	onLoginSuccess := func(dbPath string, session *appstate.LoginResult) {
		// Ouvrir la base de données
		db, err := database.InitDatabase(dbPath)
		if err != nil {
			log.Fatalf("Erreur ouverture DB: %v", err)
		}

		// Configurer la session globale
		appstate.SetDB(db)
		appstate.SetSession(session.Session)
		appstate.RefreshTaxConfig()

		// Fermer la fenêtre de login
		loginWin.Hide()

		// Construire et afficher le layout principal
		mainContent := screens.BuildMainLayout(mainWin, a)
		mainWin.SetContent(mainContent)
		mainWin.SetTitle(fmt.Sprintf("Gestion Commerciale Pro — %s — Année %d",
			session.Session.CompanyName, session.Session.FiscalYear))

		// Naviguer vers le dashboard
		appstate.Navigate(appstate.RouteDashboard)

		// Afficher la fenêtre principale
		mainWin.Show()

		// Gestionnaire de fermeture
		mainWin.SetOnClosed(func() {
			backupMode := database.GetSetting(db, "backup_mode")
			if backupMode == "daily" {
				backupPath := filepath.Join("data", "backups",
					fmt.Sprintf("backup_%s.zip", time.Now().Format("20060102")))
				if _, err := os.Stat(backupPath); os.IsNotExist(err) {
					log.Println("Sauvegarde automatique...")
				}
			}
			database.Close(db)
			log.Println("Application fermée proprement")
		})
	}

	// Construire l'écran de login
	loginContent := screens.BuildLoginScreen(loginWin, a, onLoginSuccess)
	loginWin.SetContent(loginContent)
	loginWin.Show()

	// Démarrer l'application
	a.Run()
}
