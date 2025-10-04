# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Development Commands

### Running the application
```bash
go run main.go
```
Server runs on port 8080 by default (configurable via PORT environment variable).

### Building the application
```bash
go build -o simple-go main.go
```

### Database migrations (golang-migrate)
```bash
# Run all pending migrations
./migrate.sh up

# Rollback last migration
./migrate.sh down

# Check migration status
./migrate.sh status

# Create new migration
./migrate.sh create "migration_description"

# Force migration version (use with caution)
./migrate.sh force <version>
```

Alternative using migrate CLI directly:
```bash
# Install migrate CLI (one-time setup)
go install -tags 'sqlite3' github.com/golang-migrate/migrate/v4/cmd/migrate@v4.17.1

# Run migrations
~/go/bin/migrate -path db/migrations -database sqlite3://./users.db up
~/go/bin/migrate -path db/migrations -database sqlite3://./users.db down 1
~/go/bin/migrate -path db/migrations -database sqlite3://./users.db version
```

### Dependencies
```bash
go mod tidy
go mod download
```

## Architecture Overview

This is a Go HTTP service with JWT authentication, SQLite database, and a custom migration system.

### Key Components

- **Main Application (`main.go`)**: Sets up routes, initializes database, starts HTTP server on configurable port
- **Database Layer (`db/`)**: SQLite database with bcrypt password hashing, user management functions
- **Configuration (`config/`)**: Environment-based config with defaults (JWT secret, server port, database URL)
- **Migration System (`migrations/`)**: Custom migration runner with transactional support, up/down migrations, and status tracking
- **Middleware (`middleware/`)**: JWT authentication middleware and CORS handling
- **Routes (`routes/`)**: Separate handlers for API endpoints, authentication, and static file serving
- **Public Assets (`public/`)**: HTML files for login, register, and dashboard pages

### Authentication Flow
- JWT tokens for session management
- Password hashing with bcrypt
- Protected routes use auth middleware
- Demo admin user (admin/password) seeded via migration

### Route Structure
- Public: `/health`, `/login`, `/register`, `/api/login`, `/api/register`
- Protected: `/api/users`, `/api/profile`, `/dashboard`
- Static files served from `/static/*` mapped to `public/` directory
- Root redirects to login page

### Database
- SQLite database (default: `./users.db`)
- golang-migrate for database migrations with transaction support
- Migration files in `db/migrations/` with sequential numbering
- Users table with bcrypt-hashed passwords

### Environment Variables
- `JWT_SECRET`: JWT signing key (default: "your-secret-key-change-this")
- `PORT`: Server port (default: "8080")
- `DATABASE_URL`: SQLite database file path (default: "./users.db")