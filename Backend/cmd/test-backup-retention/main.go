package main

import (
	"log"
	"time"

	"github.com/nessieaudio/ecommerce-backend/internal/backup"
	"github.com/nessieaudio/ecommerce-backend/internal/config"
	"github.com/nessieaudio/ecommerce-backend/internal/database"
)

func main() {
	log.Println("🔄 Testing Backup Retention Policy")
	log.Println("===================================")

	// Load config
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize database
	db, err := database.InitDB(cfg.DatabasePath)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// Initialize backup manager with low retention for testing
	backupManager, err := backup.NewManager(db, backup.Config{
		BackupDir:        "backups",
		DatabasePath:     cfg.DatabasePath,
		DailyRetention:   3, // Keep only 3 daily backups for testing
		MonthlyRetention: 2, // Keep only 2 monthly backups for testing
		CompressBackups:  true,
	})
	if err != nil {
		log.Fatalf("Failed to initialize backup manager: %v", err)
	}

	log.Println("\n✅ Testing with retention policy:")
	log.Println("   Daily retention: 3 backups")
	log.Println("   Monthly retention: 2 backups")

	// Create 5 daily backups (should keep only last 3)
	log.Println("\n📦 Creating 5 daily backups...")
	for i := 1; i <= 5; i++ {
		if err := backupManager.CreateBackup("daily"); err != nil {
			log.Fatalf("Failed to create backup %d: %v", i, err)
		}
		log.Printf("   Created backup %d/5", i)
		time.Sleep(100 * time.Millisecond) // Small delay between backups
	}

	// Count daily backups
	backups, _ := backupManager.ListBackups()
	dailyCount := 0
	for _, b := range backups {
		if len(b) > 13 && b[len(b)-13:len(b)-6] == "daily/" {
			dailyCount++
		}
	}

	log.Printf("\n📊 Daily backups remaining: %d (expected: 3)", dailyCount)
	if dailyCount == 3 {
		log.Println("   ✅ Retention policy working correctly!")
	} else {
		log.Printf("   ⚠️  Expected 3 backups, found %d", dailyCount)
	}

	// Create 4 monthly backups (should keep only last 2)
	log.Println("\n📦 Creating 4 monthly backups...")
	for i := 1; i <= 4; i++ {
		if err := backupManager.CreateBackup("monthly"); err != nil {
			log.Fatalf("Failed to create backup %d: %v", i, err)
		}
		log.Printf("   Created backup %d/4", i)
		time.Sleep(100 * time.Millisecond)
	}

	// Count monthly backups
	backups, _ = backupManager.ListBackups()
	monthlyCount := 0
	for _, b := range backups {
		if len(b) > 15 && b[len(b)-15:len(b)-6] == "monthly/" {
			monthlyCount++
		}
	}

	log.Printf("\n📊 Monthly backups remaining: %d (expected: 2)", monthlyCount)
	if monthlyCount == 2 {
		log.Println("   ✅ Retention policy working correctly!")
	} else {
		log.Printf("   ⚠️  Expected 2 backups, found %d", monthlyCount)
	}

	// Final summary
	log.Println("\n====================================")
	log.Println("✅ Retention policy tests complete!")
	log.Println("====================================")
	log.Println("\n💡 Old backups are automatically deleted")
	log.Println("💡 Only the most recent backups are kept")
}
