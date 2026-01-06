package main

import (
	"log"
	"time"

	"github.com/nessieaudio/ecommerce-backend/internal/backup"
	"github.com/nessieaudio/ecommerce-backend/internal/config"
	"github.com/nessieaudio/ecommerce-backend/internal/database"
)

func main() {
	log.Println("🗄️  Testing Database Backup System")
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

	log.Printf("Database: %s\n", cfg.DatabasePath)

	// Initialize backup manager
	backupManager, err := backup.NewManager(db, backup.Config{
		BackupDir:        "backups",
		DatabasePath:     cfg.DatabasePath,
		DailyRetention:   30,
		MonthlyRetention: 12,
		CompressBackups:  true,
	})
	if err != nil {
		log.Fatalf("Failed to initialize backup manager: %v", err)
	}

	log.Println("\n✅ Backup manager initialized")
	log.Println("   Backup directory: backups/")
	log.Println("   Daily retention: 30 backups")
	log.Println("   Monthly retention: 12 backups")
	log.Println("   Compression: enabled (gzip)")

	// Test 1: Create daily backup
	log.Println("\n📦 Test 1: Creating daily backup...")
	if err := backupManager.CreateBackup("daily"); err != nil {
		log.Fatalf("Failed to create daily backup: %v", err)
	}
	log.Println("   ✓ Daily backup created")

	// Small delay
	time.Sleep(500 * time.Millisecond)

	// Test 2: Create monthly backup
	log.Println("\n📦 Test 2: Creating monthly backup...")
	if err := backupManager.CreateBackup("monthly"); err != nil {
		log.Fatalf("Failed to create monthly backup: %v", err)
	}
	log.Println("   ✓ Monthly backup created")

	// Test 3: List all backups
	log.Println("\n📋 Test 3: Listing all backups...")
	backups, err := backupManager.ListBackups()
	if err != nil {
		log.Fatalf("Failed to list backups: %v", err)
	}

	if len(backups) == 0 {
		log.Println("   ⚠️  No backups found")
	} else {
		log.Printf("   Found %d backup(s):\n", len(backups))
		for i, b := range backups {
			log.Printf("   %d. %s\n", i+1, b)
		}
	}

	// Test 4: Create manual backup
	log.Println("\n📦 Test 4: Creating manual backup...")
	if err := backupManager.CreateBackup("manual"); err != nil {
		log.Fatalf("Failed to create manual backup: %v", err)
	}
	log.Println("   ✓ Manual backup created")

	// Summary
	log.Println("\n====================================")
	log.Println("✅ All backup tests passed!")
	log.Println("====================================")
	log.Println("\n📁 Check the backup directory:")
	log.Println("   ls -lh backups/daily/")
	log.Println("   ls -lh backups/monthly/")
	log.Println("\n💡 Backups are compressed with gzip (.gz)")
	log.Println("💡 Scheduled backups run daily at 3:00 AM")
	log.Println("💡 Monthly backups created on 1st of each month")
}
