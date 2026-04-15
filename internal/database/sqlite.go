package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

// DB est l'instance globale de la base de données
var DB *sql.DB

// Open ouvre (ou crée) la base de données SQLite
func Open(dbPath string) (*sql.DB, error) {
	// Créer le répertoire si nécessaire
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("impossible de créer le répertoire %s: %w", dir, err)
	}

	// Ouvrir la connexion SQLite avec options optimales
	dsn := dbPath + "?_journal_mode=WAL&_foreign_keys=ON&_busy_timeout=5000&_cache_size=-20000"
	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return nil, fmt.Errorf("erreur ouverture SQLite: %w", err)
	}

	// Configuration des connexions
	db.SetMaxOpenConns(1)    // SQLite ne supporte qu'une connexion en écriture
	db.SetMaxIdleConns(1)
	db.SetConnMaxLifetime(0) // Pas de limite de temps

	// Pragmas d'optimisation
	pragmas := []string{
		"PRAGMA journal_mode=WAL",
		"PRAGMA foreign_keys=ON",
		"PRAGMA busy_timeout=5000",
		"PRAGMA cache_size=-20000",
		"PRAGMA synchronous=NORMAL",
		"PRAGMA temp_store=MEMORY",
	}
	for _, p := range pragmas {
		if _, err := db.Exec(p); err != nil {
			log.Printf("Avertissement pragma %s: %v", p, err)
		}
	}

	// Tester la connexion
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("impossible de se connecter à la base de données: %w", err)
	}

	return db, nil
}

// InitDatabase initialise la base de données avec le schéma et les données de base
func InitDatabase(dbPath string) (*sql.DB, error) {
	db, err := Open(dbPath)
	if err != nil {
		return nil, err
	}

	// Créer les tables
	if err := CreateSchema(db); err != nil {
		return nil, fmt.Errorf("erreur création schéma: %w", err)
	}

	// Insérer les données de base
	if err := SeedDatabase(db); err != nil {
		return nil, fmt.Errorf("erreur initialisation données: %w", err)
	}

	log.Printf("Base de données initialisée: %s", dbPath)
	return db, nil
}

// CreateSchema crée toutes les tables
func CreateSchema(db *sql.DB) error {
	_, err := db.Exec(SchemaSQL)
	if err != nil {
		return fmt.Errorf("erreur création schéma: %w", err)
	}
	return nil
}

// Close ferme proprement la connexion
func Close(db *sql.DB) {
	if db != nil {
		if err := db.Close(); err != nil {
			log.Printf("Erreur fermeture DB: %v", err)
		}
	}
}

// GetDatabaseFiles liste tous les fichiers .db dans le répertoire data/
func GetDatabaseFiles(dataDir string) []string {
	var files []string
	entries, err := os.ReadDir(dataDir)
	if err != nil {
		return files
	}
	for _, e := range entries {
		if !e.IsDir() && filepath.Ext(e.Name()) == ".db" {
			name := e.Name()[:len(e.Name())-3] // Enlever .db
			files = append(files, name)
		}
	}
	return files
}

// CreateNewCompanyDB crée une nouvelle base de données pour une nouvelle société
func CreateNewCompanyDB(dataDir, companyName string) (string, error) {
	// Nom de fichier sécurisé
	safeName := sanitizeFileName(companyName)
	dbPath := filepath.Join(dataDir, safeName+".db")

	// Vérifier si le fichier existe déjà
	if _, err := os.Stat(dbPath); err == nil {
		return "", fmt.Errorf("une société avec ce nom existe déjà: %s", safeName)
	}

	db, err := InitDatabase(dbPath)
	if err != nil {
		return "", err
	}
	defer db.Close()

	return dbPath, nil
}

// sanitizeFileName nettoie le nom pour usage comme nom de fichier
func sanitizeFileName(name string) string {
	safe := ""
	for _, r := range name {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') ||
			(r >= '0' && r <= '9') || r == '_' || r == '-' {
			safe += string(r)
		} else {
			safe += "_"
		}
	}
	if safe == "" {
		safe = "company"
	}
	return safe
}

// GetSetting récupère un paramètre par clé
func GetSetting(db *sql.DB, key string) string {
	var value string
	db.QueryRow(`SELECT value FROM settings WHERE key=?`, key).Scan(&value)
	return value
}

// SetSetting enregistre un paramètre
func SetSetting(db *sql.DB, key, value string) error {
	_, err := db.Exec(`INSERT OR REPLACE INTO settings (key, value) VALUES (?,?)`, key, value)
	return err
}

// GetCurrentFiscalYearID retourne l'ID de l'année fiscale courante
func GetCurrentFiscalYearID(db *sql.DB, year int) (int, error) {
	var id int
	err := db.QueryRow(`SELECT id FROM fiscal_years WHERE year=?`, year).Scan(&id)
	if err == sql.ErrNoRows {
		// Créer l'année si elle n'existe pas
		result, err := db.Exec(`
			INSERT INTO fiscal_years (year, start_date, end_date, status)
			VALUES (?, ?||'-01-01', ?||'-12-31', 'open')`,
			year, year, year)
		if err != nil {
			return 0, err
		}
		insertedID, _ := result.LastInsertId()
		return int(insertedID), nil
	}
	return id, err
}

// BeginTx démarre une transaction
func BeginTx(db *sql.DB) (*sql.Tx, error) {
	return db.Begin()
}
