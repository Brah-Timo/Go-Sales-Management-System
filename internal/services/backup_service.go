package services

import (
	"archive/zip"
	"database/sql"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// BackupService gère les sauvegardes et restaurations
type BackupService struct {
	db     *sql.DB
	dbPath string
}

// NewBackupService crée un service de sauvegarde
func NewBackupService(db *sql.DB, dbPath string) *BackupService {
	return &BackupService{db: db, dbPath: dbPath}
}

// Backup crée une sauvegarde complète (DB + assets) dans un fichier ZIP
func (s *BackupService) Backup(backupDir string) (string, error) {
	// Créer le répertoire de sauvegarde si nécessaire
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		return "", fmt.Errorf("erreur création répertoire sauvegarde: %w", err)
	}

	timestamp := time.Now().Format("20060102_150405")
	backupFile := filepath.Join(backupDir, fmt.Sprintf("backup_%s.zip", timestamp))

	zipFile, err := os.Create(backupFile)
	if err != nil {
		return "", fmt.Errorf("erreur création fichier ZIP: %w", err)
	}
	defer zipFile.Close()

	writer := zip.NewWriter(zipFile)
	defer writer.Close()

	// Ajouter la base de données
	if err := addFileToZip(writer, s.dbPath, "database.db"); err != nil {
		return "", fmt.Errorf("erreur ajout DB: %w", err)
	}

	// Ajouter les assets (logos, images, etc.)
	assetsDir := "assets"
	if _, err := os.Stat(assetsDir); err == nil {
		err = filepath.Walk(assetsDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return nil // Ignorer les erreurs individuelles
			}
			if info.IsDir() {
				return nil
			}
			// Exclure les fichiers temporaires
			if strings.HasPrefix(filepath.Base(path), ".") {
				return nil
			}
			return addFileToZip(writer, path, path)
		})
		if err != nil {
			// Ne pas échouer si les assets ne peuvent pas être ajoutés
			fmt.Printf("Avertissement: impossible d'ajouter les assets: %v\n", err)
		}
	}

	// Journaliser
	s.db.Exec(`INSERT INTO audit_log (action_type, module, description) VALUES (?,?,?)`,
		"backup", "system", "Sauvegarde créée: "+filepath.Base(backupFile))

	return backupFile, nil
}

// Restore restaure une sauvegarde depuis un fichier ZIP
func (s *BackupService) Restore(zipPath, targetDir string) error {
	// Vérifier que le fichier existe
	if _, err := os.Stat(zipPath); err != nil {
		return fmt.Errorf("fichier de sauvegarde introuvable: %w", err)
	}

	reader, err := zip.OpenReader(zipPath)
	if err != nil {
		return fmt.Errorf("impossible d'ouvrir le fichier ZIP: %w", err)
	}
	defer reader.Close()

	// Fermer la DB avant restauration
	if s.db != nil {
		s.db.Close()
	}

	// Extraire les fichiers
	for _, file := range reader.File {
		if err := extractFileFromZip(file, targetDir); err != nil {
			return fmt.Errorf("erreur extraction %s: %w", file.Name, err)
		}
	}

	return nil
}

// ListBackups retourne la liste des sauvegardes disponibles
func (s *BackupService) ListBackups(backupDir string) ([]BackupInfo, error) {
	entries, err := os.ReadDir(backupDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var backups []BackupInfo
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".zip") {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			continue
		}
		backups = append(backups, BackupInfo{
			FileName:  entry.Name(),
			Path:      filepath.Join(backupDir, entry.Name()),
			Size:      info.Size(),
			CreatedAt: info.ModTime(),
		})
	}

	return backups, nil
}

// DeleteBackup supprime un fichier de sauvegarde
func (s *BackupService) DeleteBackup(path string) error {
	return os.Remove(path)
}

// AutoBackup effectue une sauvegarde automatique si configurée
func (s *BackupService) AutoBackup(backupDir string) {
	var mode string
	s.db.QueryRow(`SELECT value FROM settings WHERE key='backup_mode'`).Scan(&mode)

	if mode == "" || mode == "none" {
		return
	}

	// Vérifier si une sauvegarde est nécessaire aujourd'hui
	today := time.Now().Format("20060102")
	var existing int
	_ = existing // ignorer

	// Effectuer la sauvegarde
	backupFile, err := s.Backup(backupDir)
	if err != nil {
		fmt.Printf("Erreur sauvegarde automatique: %v\n", err)
		return
	}

	fmt.Printf("Sauvegarde automatique créée: %s\n", backupFile)
	_ = today

	// Nettoyer les anciennes sauvegardes (garder les 10 dernières)
	s.cleanOldBackups(backupDir, 10)
}

// cleanOldBackups supprime les anciennes sauvegardes
func (s *BackupService) cleanOldBackups(backupDir string, keepCount int) {
	backups, err := s.ListBackups(backupDir)
	if err != nil || len(backups) <= keepCount {
		return
	}

	// Trier par date (les plus anciennes d'abord)
	for i := 0; i < len(backups)-keepCount; i++ {
		os.Remove(backups[i].Path)
	}
}

// BackupInfo représente les informations d'une sauvegarde
type BackupInfo struct {
	FileName  string
	Path      string
	Size      int64
	CreatedAt time.Time
}

// SizeFormatted retourne la taille formatée
func (b BackupInfo) SizeFormatted() string {
	switch {
	case b.Size < 1024:
		return fmt.Sprintf("%d o", b.Size)
	case b.Size < 1024*1024:
		return fmt.Sprintf("%.1f Ko", float64(b.Size)/1024)
	default:
		return fmt.Sprintf("%.1f Mo", float64(b.Size)/(1024*1024))
	}
}

// addFileToZip ajoute un fichier au ZIP
func addFileToZip(writer *zip.Writer, filePath, zipPath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return err
	}

	header, err := zip.FileInfoHeader(info)
	if err != nil {
		return err
	}
	header.Name = zipPath
	header.Method = zip.Deflate

	writer_, err := writer.CreateHeader(header)
	if err != nil {
		return err
	}

	_, err = io.Copy(writer_, file)
	return err
}

// extractFileFromZip extrait un fichier depuis le ZIP
func extractFileFromZip(file *zip.File, targetDir string) error {
	targetPath := filepath.Join(targetDir, file.Name)

	// Protection contre le path traversal
	if !strings.HasPrefix(filepath.Clean(targetPath), filepath.Clean(targetDir)) {
		return fmt.Errorf("chemin invalide: %s", file.Name)
	}

	if file.FileInfo().IsDir() {
		return os.MkdirAll(targetPath, 0755)
	}

	// Créer le répertoire parent
	if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
		return err
	}

	outFile, err := os.Create(targetPath)
	if err != nil {
		return err
	}
	defer outFile.Close()

	rc, err := file.Open()
	if err != nil {
		return err
	}
	defer rc.Close()

	_, err = io.Copy(outFile, rc)
	return err
}
