# Simple Go Service

A Go HTTP service with JWT authentication, user management, and SQLite database.

## Features

- **JWT Authentication**: Secure token-based authentication
- **User Management**: Registration, login, and profile management
- **SQLite Database**: Lightweight database with automated migrations
- **Static File Serving**: HTML pages for login, register, and dashboard
- **Docker Support**: Containerized deployment with Docker Compose
- **Health Monitoring**: Health check endpoint for service monitoring

## Quick Start with Docker (Recommended)

### Prerequisites
- Docker and Docker Compose installed

### Steps

1. Clone the repository:
```bash
git clone https://github.com/dtfiedler/simple-go.git
cd simple-go
```

2. Start the service:
```bash
docker compose up -d
```

3. Access the application:
- Web Interface: http://localhost:8081/login
- Health Check: http://localhost:8081/health
- Demo credentials: `admin` / `password`

4. View logs:
```bash
docker compose logs -f simple-go
```

5. Stop the service:
```bash
docker compose down
```

## Running Locally (Development)

### Prerequisites
- Go 1.20 or higher installed

### Steps

1. Install dependencies:
```bash
go mod download
```

2. Run database migrations:
```bash
./migrate.sh up
```

3. Start the service:
```bash
go run main.go
```

The service will start on port 8080.

## API Endpoints

### Public Endpoints
- `GET /health` - Service health check
- `GET /login` - Login page
- `GET /register` - Registration page
- `POST /api/login` - User authentication
- `POST /api/register` - User registration

### Protected Endpoints (Require JWT Token)
- `GET /api/users` - List all users
- `GET /api/profile` - Current user profile
- `GET /dashboard` - User dashboard page

### Static Files
- `/static/*` - Serves files from `public/` directory

## Authentication

The service uses JWT tokens for authentication. Include the token in the Authorization header:

```bash
Authorization: Bearer <your-jwt-token>
```

### Demo User
A demo admin user is automatically created:
- **Username**: `admin`
- **Password**: `password`

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `8080` | Server port |
| `JWT_SECRET` | `your-secret-key-change-this` | JWT signing key |
| `DATABASE_URL` | `./data/sqlite/core.db` | SQLite database file path |

## Database

- **Engine**: SQLite
- **Migrations**: Automated via `golang-migrate`
- **Location**: `./data/sqlite/core.db` (configurable)

### Migration Commands

```bash
# Run migrations
./migrate.sh up

# Rollback migrations
./migrate.sh down

# Check status
./migrate.sh status

# Create new migration
./migrate.sh create "migration_name"
```

## Docker Configuration

The service includes a multi-stage Docker build optimized for production:

- **Base Image**: Debian Bullseye (for SQLite compatibility)
- **Port Mapping**: 8081:8080
- **Persistent Storage**: `./data` volume mount
- **Health Checks**: Automatic monitoring
- **Security**: Non-root container execution

## Development

For local development, see `CLAUDE.md` for detailed commands and architecture information.

## License

MIT License - see LICENSE file for details.