# Database Backup System

## Overview

Automated database backup system for the Nessie Audio eCommerce platform using SQLite's VACUUM INTO for consistent snapshots with gzip compression.

## Features

- **Automated Daily Backups**: Runs at 3:00 AM every day
- **Monthly Archives**: Created on the 1st of each month
- **Gzip Compression**: ~90% size reduction (216 KB → 22 KB)
- **Retention Policy**:
  - Last 30 daily backups
  - Last 12 monthly archives
- **Startup Backup**: Creates backup when server starts
- **Consistent Snapshots**: Uses SQLite VACUUM INTO for safe backups while database is in use

## Storage Requirements

Based on current database size (216 KB):

- **Current**: ~22 KB per compressed backup
- **30 daily backups**: ~660 KB
- **12 monthly archives**: ~264 KB
- **Total**: < 1 MB

Even with 10,000 orders (100 MB database):
- Compressed backup: ~30 MB
- 30 daily + 12 monthly: ~1.2 GB total

## Backup Locations

```
backups/
├── daily/          # Last 30 daily backups
│   └── daily_2026-01-02_16-06-47.db.gz
├── monthly/        # Last 12 monthly backups
│   └── monthly_2026-01-02_16-06-48.db.gz
└── backup_*.db.gz  # Manual backups (not auto-deleted)
```

## Testing

Run the backup test suite:

```bash
go run cmd/test-backup/main.go
```

Expected output:
- ✓ Backup manager initialized
- ✓ Daily backup created
- ✓ Monthly backup created
- ✓ Manual backup created
- ✓ Backups listed successfully

## Manual Backup

To create a manual backup at any time:

```go
backupManager.CreateBackup("daily")   // Daily backup
backupManager.CreateBackup("monthly") // Monthly backup
backupManager.CreateBackup("manual")  // Manual (not auto-deleted)
```

## Restore from Backup

To restore the database from a backup:

```go
err := backupManager.RestoreBackup("backups/daily/daily_2026-01-02_16-06-47.db.gz")
```

**Important**:
- Server must be stopped before restore
- Current database is backed up to `.before_restore` before overwriting
- Compressed backups (.gz) are automatically decompressed

## Scheduled Backups

Backups are automatically scheduled:

- **Daily**: 3:00 AM every day
- **Monthly**: 3:00 AM on the 1st of each month
- **Startup**: When server starts (one-time)

## Manual Backup Commands

List all backups:
```bash
ls -lh backups/daily/
ls -lh backups/monthly/
```

Decompress a backup manually:
```bash
gunzip backups/daily/daily_2026-01-02_16-06-47.db.gz
```

Copy backup to safe location:
```bash
cp backups/daily/daily_2026-01-02_16-06-47.db.gz ~/safe-location/
```

## Files

- **internal/backup/backup.go** - Backup manager implementation
- **cmd/test-backup/main.go** - Backup testing tool
- **BACKUP.md** - This documentation

## Production Recommendations

1. **Cloud Storage**: Upload daily backups to S3/Google Cloud Storage for off-site redundancy
2. **Monitoring**: Add email alerts if backups fail
3. **Verification**: Periodically test restore process
4. **Retention**: Adjust retention based on compliance requirements

## Configuration

Backup settings in `cmd/server/main.go`:

```go
backupManager, err := backup.NewManager(db, backup.Config{
    BackupDir:        "backups",      // Backup directory
    DatabasePath:     cfg.DatabasePath, // Database file path
    DailyRetention:   30,              // Keep last 30 daily backups
    MonthlyRetention: 12,              // Keep last 12 monthly backups
    CompressBackups:  true,            // Enable gzip compression
})
```

## Troubleshooting

**Backup fails with "database is locked":**
- SQLite VACUUM INTO handles this automatically
- Backups are safe to run while database is in use

**Compressed files won't open:**
- Use `gunzip` or compatible tool
- Verify file wasn't corrupted during transfer

**Old backups not being deleted:**
- Check retention policy settings
- Verify backup directory permissions
- Check logs for cleanup errors

## Security

- Backup files contain full database (including customer data)
- Ensure backup directory has restricted permissions (0755)
- Do not expose backup directory via web server
- Encrypt backups if storing off-site
