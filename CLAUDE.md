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
~/go/bin/migrate -path db/migrations -database sqlite3://./data/sqlite/core.db up
~/go/bin/migrate -path db/migrations -database sqlite3://./data/sqlite/core.db down 1
~/go/bin/migrate -path db/migrations -database sqlite3://./data/sqlite/core.db version
```

### Dependencies
```bash
go mod tidy
go mod download
```

## Docker Development

### Running with Docker Compose (Recommended)
```bash
# Build and start the container
docker compose up -d

# View logs
docker compose logs simple-go

# Follow logs in real-time
docker compose logs -f simple-go

# Stop the container
docker compose down

# Rebuild and start (after code changes)
docker compose up --build -d
```

### Manual Docker Commands
```bash
# Build the Docker image
docker build -t simple-go .

# Run the container manually
docker run -d -p 8081:8080 \
  -e JWT_SECRET=your-secret-key \
  -e DATABASE_URL=/data/sqlite/core.db \
  -v $(pwd)/data:/data \
  --name simple-go-container \
  simple-go

# Stop and remove container
docker stop simple-go-container
docker rm simple-go-container
```

### Docker Configuration
- **Port Mapping**: Host port 8081 → Container port 8080
- **Volume Mount**: `./data:/data` for persistent database storage
- **Health Check**: Automatically monitors `/health` endpoint
- **Environment Variables**: Configurable JWT secret, port, and database path
- **Multi-stage Build**: Optimized Debian-based build for SQLite compatibility

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
- SQLite database (default: `./data/sqlite/core.db`)
- golang-migrate for database migrations with transaction support
- Migration files in `db/migrations/` with sequential numbering
- Users table with bcrypt-hashed passwords

### Environment Variables
- `JWT_SECRET`: JWT signing key (default: "your-secret-key-change-this")
- `PORT`: Server port (default: "8080")
- `DATABASE_URL`: SQLite database file path (default: "./data/sqlite/core.db")