package app

import (
	"database/sql"
	"gestion-commerciale/internal/models"
	"sync"
)

// CurrentSession est la session utilisateur globale
var CurrentSession *models.Session

// AppDB est la connexion à la base de données active
var AppDB *sql.DB

// AppMu protège l'accès concurrent à AppDB
var AppMu sync.RWMutex

// SetSession initialise la session utilisateur
func SetSession(s *models.Session) {
	AppMu.Lock()
	defer AppMu.Unlock()
	CurrentSession = s
}

// GetSession retourne la session courante
func GetSession() *models.Session {
	AppMu.RLock()
	defer AppMu.RUnlock()
	return CurrentSession
}

// SetDB définit la base de données active
func SetDB(db *sql.DB) {
	AppMu.Lock()
	defer AppMu.Unlock()
	AppDB = db
}

// GetDB retourne la base de données active
func GetDB() *sql.DB {
	AppMu.RLock()
	defer AppMu.RUnlock()
	return AppDB
}

// HasPermission vérifie si l'utilisateur courant a une permission
func HasPermission(perm string) bool {
	s := GetSession()
	if s == nil {
		return false
	}
	return s.HasPermission(perm)
}

// IsAdmin vérifie si l'utilisateur est administrateur
func IsAdmin() bool {
	s := GetSession()
	return s != nil && s.Role == models.RoleAdmin
}
