#!/bin/bash

# Script to create a new database migration
# Usage: ./scripts/create-migration.sh migration_name

if [ -z "$1" ]; then
    echo "Usage: ./scripts/create-migration.sh <migration_name>"
    echo "Example: ./scripts/create-migration.sh add_user_preferences"
    exit 1
fi

MIGRATION_NAME=$1

# Create migration using migrate CLI
migrate create -ext sql -dir migrations -seq "$MIGRATION_NAME"

# Also copy to internal/migrations for embedding
cp migrations/*.sql internal/migrations/ 2>/dev/null || true

echo ""
echo "✅ Migration created successfully!"
echo ""
echo "Next steps:"
echo "1. Edit the generated .up.sql file with your schema changes"
echo "2. Edit the generated .down.sql file with the rollback logic"
echo "3. Copy the files to internal/migrations: cp migrations/*.sql internal/migrations/"
echo "4. Restart your server - migrations will run automatically"
echo ""
echo "Files created in migrations/ directory"
