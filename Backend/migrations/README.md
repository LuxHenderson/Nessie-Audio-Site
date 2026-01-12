# Database Migrations

This directory contains database migration files for the Nessie Audio eCommerce backend.

## Overview

Migrations are automatically applied when the server starts. The system uses [golang-migrate](https://github.com/golang-migrate/migrate) to manage schema changes.

## How It Works

- Migrations are numbered sequentially (e.g., `000001_`, `000002_`, etc.)
- Each migration has two files:
  - `.up.sql` - Applies the schema change
  - `.down.sql` - Reverts the schema change
- The `schema_migrations` table tracks which migrations have been applied
- Migrations run automatically on server startup
- Only pending migrations are applied (safe to restart server multiple times)

## Creating a New Migration

### Using the Helper Script (Recommended)

```bash
./scripts/create-migration.sh add_user_preferences
```

This will:
1. Create numbered migration files in `migrations/`
2. Copy them to `internal/migrations/` for embedding
3. Provide next steps

### Manual Creation

```bash
migrate create -ext sql -dir migrations -seq migration_name
```

Then copy the files to `internal/migrations/`:
```bash
cp migrations/*.sql internal/migrations/
```

## Migration File Structure

**Up migration** (`*_name.up.sql`):
```sql
-- Add new feature
CREATE TABLE IF NOT EXISTS user_preferences (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL,
    theme TEXT DEFAULT 'light',
    created_at DATETIME NOT NULL,
    FOREIGN KEY (user_id) REFERENCES users(id)
);

CREATE INDEX IF NOT EXISTS idx_user_prefs_user ON user_preferences(user_id);
```

**Down migration** (`*_name.down.sql`):
```sql
-- Revert changes
DROP INDEX IF EXISTS idx_user_prefs_user;
DROP TABLE IF EXISTS user_preferences;
```

## Best Practices

1. **Always use IF EXISTS/IF NOT EXISTS** - Makes migrations idempotent
2. **Test rollback** - Ensure `.down.sql` properly reverts changes
3. **One logical change per migration** - Don't combine unrelated schema changes
4. **Never modify applied migrations** - Create a new migration instead
5. **Use transactions** - SQLite automatically wraps each migration in a transaction

## Important Notes

- Migrations are embedded in the Go binary at compile time
- After creating a migration, you must:
  1. Copy it to `internal/migrations/`
  2. Rebuild the server: `go build -o bin/server ./cmd/server`
  3. Restart the server
- The system will automatically detect and apply new migrations

## Checking Migration Status

View current migration version:
```bash
sqlite3 nessie_store.db "SELECT * FROM schema_migrations;"
```

## Manual Migration Commands

If you need to run migrations manually:

```bash
# Check current version
migrate -path migrations -database "sqlite3://nessie_store.db" version

# Apply all pending migrations
migrate -path migrations -database "sqlite3://nessie_store.db" up

# Rollback last migration
migrate -path migrations -database "sqlite3://nessie_store.db" down 1

# Go to specific version
migrate -path migrations -database "sqlite3://nessie_store.db" goto 2

# Force set version (use with caution!)
migrate -path migrations -database "sqlite3://nessie_store.db" force 1
```

## Troubleshooting

### Dirty Database State

If a migration fails partway through, the database may be marked as "dirty":

```
⚠️  WARNING: Database is in dirty state at version X
```

**Solution:**
1. Manually fix the database to match the expected state
2. Force the version: `migrate -path migrations -database "sqlite3://nessie_store.db" force X`

### Migration Already Applied

If you see "migration already applied", this is normal - the system detected the migration was already run.

### Development vs Production

- **Development**: Migrations run automatically on every server restart
- **Production**: Same behavior - migrations run automatically on deployment
- Both environments use the same migration files

## Example Workflow

1. **Add new feature requiring schema change**:
   ```bash
   ./scripts/create-migration.sh add_product_reviews
   ```

2. **Edit the generated files**:
   - `migrations/000002_add_product_reviews.up.sql` - Add reviews table
   - `migrations/000002_add_product_reviews.down.sql` - Drop reviews table

3. **Copy to internal package**:
   ```bash
   cp migrations/*.sql internal/migrations/
   ```

4. **Rebuild and restart**:
   ```bash
   go build -o bin/server ./cmd/server
   ./bin/server
   ```

5. **Verify**:
   ```
   📊 Current migration version: 2
   ✅ Successfully applied migrations (now at version 2)
   ```

The migration system is now part of your deployment process - no manual database changes needed!
