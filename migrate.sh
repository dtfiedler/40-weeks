#!/bin/bash

# Migration script using golang-migrate CLI
# Usage: ./migrate.sh [up|down|status|create]

MIGRATE_CMD=~/go/bin/migrate
DATABASE_URL="sqlite3://./users.db"
MIGRATIONS_PATH="db/migrations"

case "$1" in
    "up")
        echo "Running migrations..."
        $MIGRATE_CMD -path $MIGRATIONS_PATH -database $DATABASE_URL up
        ;;
    "down")
        echo "Rolling back last migration..."
        $MIGRATE_CMD -path $MIGRATIONS_PATH -database $DATABASE_URL down 1
        ;;
    "status")
        echo "Migration status:"
        $MIGRATE_CMD -path $MIGRATIONS_PATH -database $DATABASE_URL version
        ;;
    "create")
        if [ -z "$2" ]; then
            echo "Usage: ./migrate.sh create <migration_name>"
            exit 1
        fi
        echo "Creating migration: $2"
        $MIGRATE_CMD create -ext sql -dir $MIGRATIONS_PATH -seq "$2"
        ;;
    "force")
        if [ -z "$2" ]; then
            echo "Usage: ./migrate.sh force <version>"
            exit 1
        fi
        echo "Forcing migration version: $2"
        $MIGRATE_CMD -path $MIGRATIONS_PATH -database $DATABASE_URL force "$2"
        ;;
    *)
        echo "Usage: ./migrate.sh [up|down|status|create|force]"
        echo ""
        echo "Commands:"
        echo "  up       - Run all pending migrations"
        echo "  down     - Rollback last migration"
        echo "  status   - Show current migration version"
        echo "  create   - Create new migration file"
        echo "  force    - Force set migration version (use with caution)"
        echo ""
        echo "Examples:"
        echo "  ./migrate.sh up"
        echo "  ./migrate.sh down"
        echo "  ./migrate.sh status"
        echo "  ./migrate.sh create add_user_roles"
        ;;
esac