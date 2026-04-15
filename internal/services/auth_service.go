package services

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"gestion-commerciale/internal/models"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// AuthService gère l'authentification
type AuthService struct {
	db *sql.DB
}

// NewAuthService crée un nouveau service d'authentification
func NewAuthService(db *sql.DB) *AuthService {
	return &AuthService{db: db}
}

// LoginAttempt représente une tentative de connexion
type LoginAttempt struct {
	Count    int
	LastTime time.Time
}

// loginAttempts suivi des tentatives par utilisateur
var loginAttempts = make(map[string]*LoginAttempt)

// MaxLoginAttempts nombre maximum de tentatives
const MaxLoginAttempts = 3

// LockoutDuration durée de verrouillage
const LockoutDuration = 30 * time.Second

// Login tente de connecter un utilisateur
func (s *AuthService) Login(username, password string) (*models.User, error) {
	// Vérifier le verrouillage
	if attempt, ok := loginAttempts[username]; ok {
		if attempt.Count >= MaxLoginAttempts {
			elapsed := time.Since(attempt.LastTime)
			if elapsed < LockoutDuration {
				remaining := int((LockoutDuration - elapsed).Seconds())
				return nil, fmt.Errorf("compte verrouillé. Réessayez dans %d secondes", remaining)
			}
			// Réinitialiser après la période de verrouillage
			delete(loginAttempts, username)
		}
	}

	// Chercher l'utilisateur
	var user models.User
	err := s.db.QueryRow(`
		SELECT id, username, full_name, password_hash, role, permissions_json, is_active
		FROM users WHERE username=? AND is_active=1`, username).
		Scan(&user.ID, &user.Username, &user.FullName, &user.PasswordHash,
			&user.Role, &user.PermissionsJSON, &user.IsActive)

	if err == sql.ErrNoRows {
		s.recordFailedAttempt(username)
		return nil, errors.New("nom d'utilisateur ou mot de passe incorrect")
	}
	if err != nil {
		return nil, fmt.Errorf("erreur base de données: %w", err)
	}

	// Vérifier le mot de passe
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		s.recordFailedAttempt(username)
		return nil, errors.New("nom d'utilisateur ou mot de passe incorrect")
	}

	// Connexion réussie - réinitialiser les tentatives
	delete(loginAttempts, username)

	// Charger les permissions
	user.Permissions = make(map[string]bool)
	if user.PermissionsJSON != "" {
		json.Unmarshal([]byte(user.PermissionsJSON), &user.Permissions)
	}

	// Mettre à jour la dernière connexion
	now := time.Now()
	s.db.Exec(`UPDATE users SET last_login=? WHERE id=?`, now.Format("2006-01-02 15:04:05"), user.ID)

	// Journaliser
	s.db.Exec(`INSERT INTO audit_log (user_id, action_type, module, description) VALUES (?,?,?,?)`,
		user.ID, "login", "auth", "Connexion réussie")

	return &user, nil
}

// recordFailedAttempt enregistre une tentative échouée
func (s *AuthService) recordFailedAttempt(username string) {
	if attempt, ok := loginAttempts[username]; ok {
		attempt.Count++
		attempt.LastTime = time.Now()
	} else {
		loginAttempts[username] = &LoginAttempt{Count: 1, LastTime: time.Now()}
	}
}

// Logout déconnecte l'utilisateur
func (s *AuthService) Logout(userID int) {
	s.db.Exec(`INSERT INTO audit_log (user_id, action_type, module, description) VALUES (?,?,?,?)`,
		userID, "logout", "auth", "Déconnexion")
}

// HashPassword génère le hash bcrypt d'un mot de passe
func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// VerifyPassword vérifie un mot de passe contre son hash
func VerifyPassword(password, hash string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}

// GetAllUsers retourne tous les utilisateurs
func (s *AuthService) GetAllUsers() ([]models.User, error) {
	rows, err := s.db.Query(`
		SELECT id, username, full_name, role, permissions_json, is_active, last_login
		FROM users ORDER BY role, full_name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var u models.User
		var lastLogin sql.NullString
		rows.Scan(&u.ID, &u.Username, &u.FullName, &u.Role,
			&u.PermissionsJSON, &u.IsActive, &lastLogin)
		u.Permissions = make(map[string]bool)
		json.Unmarshal([]byte(u.PermissionsJSON), &u.Permissions)
		users = append(users, u)
	}
	return users, nil
}

// CreateUser crée un nouvel utilisateur
func (s *AuthService) CreateUser(username, fullName, password, role string, permissions map[string]bool) error {
	hash, err := HashPassword(password)
	if err != nil {
		return err
	}

	permsJSON, _ := json.Marshal(permissions)

	_, err = s.db.Exec(`
		INSERT INTO users (username, full_name, password_hash, role, permissions_json, is_active)
		VALUES (?,?,?,?,?,1)`,
		username, fullName, hash, role, string(permsJSON))
	return err
}

// UpdateUserPermissions met à jour les permissions d'un utilisateur
func (s *AuthService) UpdateUserPermissions(userID int, permissions map[string]bool) error {
	permsJSON, _ := json.Marshal(permissions)
	_, err := s.db.Exec(`UPDATE users SET permissions_json=? WHERE id=?`, string(permsJSON), userID)
	return err
}

// ToggleUserActive active/désactive un utilisateur
func (s *AuthService) ToggleUserActive(userID int, active bool) error {
	val := 0
	if active { val = 1 }
	_, err := s.db.Exec(`UPDATE users SET is_active=? WHERE id=?`, val, userID)
	return err
}

// ChangePassword change le mot de passe d'un utilisateur
func (s *AuthService) ChangePassword(userID int, newPassword string) error {
	hash, err := HashPassword(newPassword)
	if err != nil {
		return err
	}
	_, err = s.db.Exec(`UPDATE users SET password_hash=? WHERE id=?`, hash, userID)
	return err
}

// HasPermission vérifie si un utilisateur a une permission donnée
func (s *AuthService) HasPermission(userID int, permission string) bool {
	var permsJSON string
	s.db.QueryRow(`SELECT permissions_json FROM users WHERE id=?`, userID).Scan(&permsJSON)

	var perms map[string]bool
	json.Unmarshal([]byte(permsJSON), &perms)
	return perms[permission]
}

// BuildSession construit une session depuis un utilisateur et des paramètres
func BuildSession(user *models.User, fiscalYear int, dbPath, companyName string) *models.Session {
	return &models.Session{
		UserID:      user.ID,
		Username:    user.Username,
		FullName:    user.FullName,
		Role:        user.Role,
		FiscalYear:  fiscalYear,
		CompanyName: companyName,
		Permissions: user.Permissions,
		DBPath:      dbPath,
	}
}
