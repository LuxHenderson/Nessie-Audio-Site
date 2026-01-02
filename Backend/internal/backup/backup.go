package backup

import (
	"compress/gzip"
	"database/sql"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// Config holds backup configuration
type Config struct {
	BackupDir         string
	DatabasePath      string
	DailyRetention    int // Number of daily backups to keep
	MonthlyRetention  int // Number of monthly backups to keep
	CompressBackups   bool
}

// Manager handles database backups
type Manager struct {
	config Config
	db     *sql.DB
}

// NewManager creates a new backup manager
func NewManager(db *sql.DB, config Config) (*Manager, error) {
	// Set defaults if not provided
	if config.BackupDir == "" {
		config.BackupDir = "backups"
	}
	if config.DailyRetention == 0 {
		config.DailyRetention = 30
	}
	if config.MonthlyRetention == 0 {
		config.MonthlyRetention = 12
	}
	config.CompressBackups = true // Always compress

	// Ensure backup directory exists
	if err := os.MkdirAll(config.BackupDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create backup directory: %w", err)
	}

	// Create subdirectories for daily and monthly backups
	dailyDir := filepath.Join(config.BackupDir, "daily")
	monthlyDir := filepath.Join(config.BackupDir, "monthly")

	if err := os.MkdirAll(dailyDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create daily backup directory: %w", err)
	}
	if err := os.MkdirAll(monthlyDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create monthly backup directory: %w", err)
	}

	return &Manager{
		config: config,
		db:     db,
	}, nil
}

// CreateBackup creates a new database backup
func (m *Manager) CreateBackup(backupType string) error {
	timestamp := time.Now().Format("2006-01-02_15-04-05")

	var backupDir string
	var filename string

	switch backupType {
	case "daily":
		backupDir = filepath.Join(m.config.BackupDir, "daily")
		filename = fmt.Sprintf("daily_%s.db", timestamp)
	case "monthly":
		backupDir = filepath.Join(m.config.BackupDir, "monthly")
		filename = fmt.Sprintf("monthly_%s.db", timestamp)
	default:
		backupDir = m.config.BackupDir
		filename = fmt.Sprintf("backup_%s.db", timestamp)
	}

	backupPath := filepath.Join(backupDir, filename)

	// Use SQLite's backup API for consistent snapshots
	if err := m.backupDatabase(backupPath); err != nil {
		return fmt.Errorf("backup failed: %w", err)
	}

	log.Printf("Database backup created: %s", backupPath)

	// Compress the backup
	if m.config.CompressBackups {
		if err := m.compressBackup(backupPath); err != nil {
			log.Printf("Warning: Failed to compress backup: %v", err)
		} else {
			// Remove uncompressed backup after successful compression
			os.Remove(backupPath)
			log.Printf("Backup compressed: %s.gz", backupPath)
		}
	}

	// Clean up old backups
	if err := m.cleanupOldBackups(backupType); err != nil {
		log.Printf("Warning: Failed to cleanup old backups: %v", err)
	}

	return nil
}

// backupDatabase uses SQLite's backup API to create a consistent snapshot
func (m *Manager) backupDatabase(destPath string) error {
	// SQLite backup using file copy with VACUUM INTO (SQLite 3.27.0+)
	// This is more reliable than direct file copying
	query := fmt.Sprintf("VACUUM INTO '%s'", destPath)

	_, err := m.db.Exec(query)
	if err != nil {
		// Fallback to manual copy if VACUUM INTO fails (older SQLite versions)
		return m.copyDatabaseFile(destPath)
	}

	return nil
}

// copyDatabaseFile performs a manual file copy as fallback
func (m *Manager) copyDatabaseFile(destPath string) error {
	sourceFile, err := os.Open(m.config.DatabasePath)
	if err != nil {
		return fmt.Errorf("failed to open source database: %w", err)
	}
	defer sourceFile.Close()

	destFile, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("failed to create backup file: %w", err)
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	if err != nil {
		return fmt.Errorf("failed to copy database: %w", err)
	}

	return nil
}

// compressBackup compresses a backup file using gzip
func (m *Manager) compressBackup(filepath string) error {
	// Open original file
	sourceFile, err := os.Open(filepath)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	// Create compressed file
	destFile, err := os.Create(filepath + ".gz")
	if err != nil {
		return err
	}
	defer destFile.Close()

	// Create gzip writer
	gzipWriter := gzip.NewWriter(destFile)
	defer gzipWriter.Close()

	// Copy and compress
	_, err = io.Copy(gzipWriter, sourceFile)
	return err
}

// cleanupOldBackups removes backups exceeding retention policy
func (m *Manager) cleanupOldBackups(backupType string) error {
	var backupDir string
	var retention int

	switch backupType {
	case "daily":
		backupDir = filepath.Join(m.config.BackupDir, "daily")
		retention = m.config.DailyRetention
	case "monthly":
		backupDir = filepath.Join(m.config.BackupDir, "monthly")
		retention = m.config.MonthlyRetention
	default:
		return nil // Don't clean up manual backups
	}

	// Get all backup files
	files, err := os.ReadDir(backupDir)
	if err != nil {
		return err
	}

	// Filter backup files
	var backupFiles []os.DirEntry
	for _, file := range files {
		if !file.IsDir() && (strings.HasSuffix(file.Name(), ".db") || strings.HasSuffix(file.Name(), ".db.gz")) {
			backupFiles = append(backupFiles, file)
		}
	}

	// Sort by modification time (newest first)
	sort.Slice(backupFiles, func(i, j int) bool {
		infoI, _ := backupFiles[i].Info()
		infoJ, _ := backupFiles[j].Info()
		return infoI.ModTime().After(infoJ.ModTime())
	})

	// Remove old backups
	for i := retention; i < len(backupFiles); i++ {
		filePath := filepath.Join(backupDir, backupFiles[i].Name())
		if err := os.Remove(filePath); err != nil {
			log.Printf("Warning: Failed to remove old backup %s: %v", filePath, err)
		} else {
			log.Printf("Removed old backup: %s", filePath)
		}
	}

	return nil
}

// StartScheduledBackups starts automated backup scheduler
func (m *Manager) StartScheduledBackups() {
	// Daily backup at 3:00 AM
	go func() {
		for {
			now := time.Now()

			// Calculate next 3:00 AM
			next := time.Date(now.Year(), now.Month(), now.Day(), 3, 0, 0, 0, now.Location())
			if now.After(next) {
				next = next.Add(24 * time.Hour)
			}

			// Wait until 3:00 AM
			time.Sleep(time.Until(next))

			// Create daily backup
			if err := m.CreateBackup("daily"); err != nil {
				log.Printf("Scheduled daily backup failed: %v", err)
			}

			// Check if it's the first day of the month for monthly backup
			if time.Now().Day() == 1 {
				if err := m.CreateBackup("monthly"); err != nil {
					log.Printf("Scheduled monthly backup failed: %v", err)
				}
			}
		}
	}()

	log.Println("Scheduled backups started (daily at 3:00 AM)")
}

// BackupAfterOrder creates a backup after a successful order (optional)
func (m *Manager) BackupAfterOrder() error {
	// Create a timestamped backup in the daily folder
	// This won't count against retention as it will be cleaned up with daily backups
	return m.CreateBackup("daily")
}

// RestoreBackup restores a database from a backup file
func (m *Manager) RestoreBackup(backupPath string) error {
	// Check if backup is compressed
	var sourceFile *os.File
	var err error

	if strings.HasSuffix(backupPath, ".gz") {
		// Decompress first
		gzFile, err := os.Open(backupPath)
		if err != nil {
			return fmt.Errorf("failed to open compressed backup: %w", err)
		}
		defer gzFile.Close()

		gzReader, err := gzip.NewReader(gzFile)
		if err != nil {
			return fmt.Errorf("failed to create gzip reader: %w", err)
		}
		defer gzReader.Close()

		// Create temporary uncompressed file
		tempFile, err := os.CreateTemp("", "restore_*.db")
		if err != nil {
			return fmt.Errorf("failed to create temp file: %w", err)
		}
		defer os.Remove(tempFile.Name())
		defer tempFile.Close()

		// Decompress to temp file
		if _, err := io.Copy(tempFile, gzReader); err != nil {
			return fmt.Errorf("failed to decompress backup: %w", err)
		}

		// Reopen for reading
		tempFile.Close()
		sourceFile, err = os.Open(tempFile.Name())
		if err != nil {
			return fmt.Errorf("failed to reopen temp file: %w", err)
		}
	} else {
		sourceFile, err = os.Open(backupPath)
		if err != nil {
			return fmt.Errorf("failed to open backup: %w", err)
		}
	}
	defer sourceFile.Close()

	// Close current database connection
	if err := m.db.Close(); err != nil {
		return fmt.Errorf("failed to close database: %w", err)
	}

	// Backup current database before overwriting
	currentBackup := m.config.DatabasePath + ".before_restore"
	if err := os.Rename(m.config.DatabasePath, currentBackup); err != nil {
		log.Printf("Warning: Could not backup current database: %v", err)
	}

	// Restore backup
	destFile, err := os.Create(m.config.DatabasePath)
	if err != nil {
		// Try to restore the original
		os.Rename(currentBackup, m.config.DatabasePath)
		return fmt.Errorf("failed to create database file: %w", err)
	}
	defer destFile.Close()

	if _, err := io.Copy(destFile, sourceFile); err != nil {
		destFile.Close()
		// Try to restore the original
		os.Remove(m.config.DatabasePath)
		os.Rename(currentBackup, m.config.DatabasePath)
		return fmt.Errorf("failed to restore backup: %w", err)
	}

	log.Printf("Database restored from: %s", backupPath)
	log.Printf("Previous database saved as: %s", currentBackup)

	return nil
}

// ListBackups returns all available backups
func (m *Manager) ListBackups() ([]string, error) {
	var backups []string

	// Check daily backups
	dailyDir := filepath.Join(m.config.BackupDir, "daily")
	dailyFiles, _ := os.ReadDir(dailyDir)
	for _, file := range dailyFiles {
		if !file.IsDir() {
			backups = append(backups, filepath.Join(dailyDir, file.Name()))
		}
	}

	// Check monthly backups
	monthlyDir := filepath.Join(m.config.BackupDir, "monthly")
	monthlyFiles, _ := os.ReadDir(monthlyDir)
	for _, file := range monthlyFiles {
		if !file.IsDir() {
			backups = append(backups, filepath.Join(monthlyDir, file.Name()))
		}
	}

	return backups, nil
}
