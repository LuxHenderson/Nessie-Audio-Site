# Database Migrations - Quick Start Guide

## What You Get

Your database schema changes are now **fully automated**:
- ✅ Migrations run automatically on server startup
- ✅ No manual SQL commands needed
- ✅ Version controlled schema changes
- ✅ Safe rollback capability
- ✅ Works in dev and production

## Common Tasks

### Creating a New Migration

When you need to change the database schema:

```bash
./scripts/create-migration.sh add_feature_name
```

Example:
```bash
./scripts/create-migration.sh add_user_preferences
```

This creates two files:
- `migrations/000002_add_user_preferences.up.sql` - Add the feature
- `migrations/000002_add_user_preferences.down.sql` - Remove the feature

### Editing Migration Files

**Up migration** (add_user_preferences.up.sql):
```sql
CREATE TABLE IF NOT EXISTS user_preferences (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL,
    theme TEXT DEFAULT 'light',
    notifications_enabled BOOLEAN DEFAULT 1,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL,
    FOREIGN KEY (user_id) REFERENCES customers(id)
);

CREATE INDEX IF NOT EXISTS idx_user_prefs_user ON user_preferences(user_id);
```

**Down migration** (add_user_preferences.down.sql):
```sql
DROP INDEX IF EXISTS idx_user_prefs_user;
DROP TABLE IF EXISTS user_preferences;
```

### Applying Migrations

After editing your migration files:

```bash
# 1. Copy to internal package
cp migrations/*.sql internal/migrations/

# 2. Rebuild server
go build -o bin/server ./cmd/server

# 3. Restart server (migrations apply automatically)
./bin/server
```

You'll see:
```
🔄 Checking for database migrations...
📊 Current migration version: 1
✅ Successfully applied migrations (now at version 2)
```

## That's It!

The migration system handles everything automatically:
- ✅ Tracks which migrations have run
- ✅ Only applies new migrations
- ✅ Safe to restart server multiple times
- ✅ Works on deployment to production

## Checking Status

View current version:
```bash
sqlite3 nessie_store.db "SELECT * FROM schema_migrations;"
```

View all tables:
```bash
sqlite3 nessie_store.db ".tables"
```

## Manual Control (If Needed)

Rollback last migration:
```bash
migrate -path migrations -database "sqlite3://nessie_store.db" down 1
```

Force version (emergency only):
```bash
migrate -path migrations -database "sqlite3://nessie_store.db" force 1
```

## Best Practices

1. **Always use IF EXISTS/IF NOT EXISTS** - Makes migrations safe to re-run
2. **Test both up and down** - Ensure rollback works
3. **One change per migration** - Keep migrations focused
4. **Never modify applied migrations** - Create a new one instead

## Example Workflow

Let's say you want to add product reviews:

```bash
# 1. Create migration
./scripts/create-migration.sh add_product_reviews

# 2. Edit migrations/000002_add_product_reviews.up.sql
cat > migrations/000002_add_product_reviews.up.sql << 'EOF'
CREATE TABLE IF NOT EXISTS product_reviews (
    id TEXT PRIMARY KEY,
    product_id TEXT NOT NULL,
    customer_id TEXT NOT NULL,
    rating INTEGER NOT NULL CHECK(rating >= 1 AND rating <= 5),
    review_text TEXT,
    created_at DATETIME NOT NULL,
    FOREIGN KEY (product_id) REFERENCES products(id),
    FOREIGN KEY (customer_id) REFERENCES customers(id)
);

CREATE INDEX IF NOT EXISTS idx_reviews_product ON product_reviews(product_id);
CREATE INDEX IF NOT EXISTS idx_reviews_customer ON product_reviews(customer_id);
EOF

# 3. Edit migrations/000002_add_product_reviews.down.sql
cat > migrations/000002_add_product_reviews.down.sql << 'EOF'
DROP INDEX IF EXISTS idx_reviews_customer;
DROP INDEX IF EXISTS idx_reviews_product;
DROP TABLE IF EXISTS product_reviews;
EOF

# 4. Apply
cp migrations/*.sql internal/migrations/
go build -o bin/server ./cmd/server
./bin/server
```

Done! The reviews table is now in your database.

## Troubleshooting

**"Database in dirty state"**
- A migration failed partway through
- Manually fix the database, then force the version

**"No such table: schema_migrations"**
- First run will create this automatically
- This tracks which migrations have run

**Migration didn't apply**
- Check it's in `internal/migrations/`
- Verify you rebuilt the server
- Check server logs for errors

## More Details

See [migrations/README.md](migrations/README.md) for comprehensive documentation.
