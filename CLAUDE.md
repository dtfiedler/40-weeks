# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Architecture Overview

This is a **single Go HTTP service** that serves both the API backend and the frontend HTML pages. The frontend uses **Tailwind CSS** and **shadcn/ui-inspired components** for a modern, professional look.

## Development Commands

### Running the application
```bash
cd api && go run main.go
```
Server runs on port 8080 by default (configurable via PORT environment variable).

### Building the application
```bash
cd api && go build -o simple-go main.go
```

### Database migrations (golang-migrate)
```bash
cd api

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
cd api

# Install migrate CLI (one-time setup)
go install -tags 'sqlite3' github.com/golang-migrate/migrate/v4/cmd/migrate@v4.17.1

# Run migrations
~/go/bin/migrate -path db/migrations -database sqlite3://./data/sqlite/core.db up
~/go/bin/migrate -path db/migrations -database sqlite3://./data/sqlite/core.db down 1
~/go/bin/migrate -path db/migrations -database sqlite3://./data/sqlite/core.db version
```

### Dependencies
```bash
cd api && go mod tidy && go mod download
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

## Architecture Details

### Single Service Design
This application uses a **unified architecture** where one Go service handles both:
- **API Backend**: JWT authentication, database operations, user management
- **Frontend Serving**: HTML pages with Tailwind CSS and shadcn/ui components

### Key Components

#### Backend (`api/`)
- **Main Application (`api/main.go`)**: Sets up routes, initializes database, starts HTTP server
- **Database Layer (`api/db/`)**: SQLite database with bcrypt password hashing, user management functions
- **Configuration (`api/config/`)**: Environment-based config with defaults (JWT secret, server port, database URL)
- **Migration System (`api/migrations/`)**: Custom migration runner with transactional support, up/down migrations, and status tracking
- **Middleware (`api/middleware/`)**: JWT authentication middleware and CORS handling
- **Routes (`api/routes/`)**: Separate handlers for API endpoints, authentication, and static file serving

#### Frontend (`api/public/`)
- **login.html**: Modern login page with Tailwind CSS and shadcn/ui styling
- **register.html**: User registration page with form validation
- **dashboard.html**: Protected dashboard showing user profile and system users

### Frontend Technology Stack
- **Tailwind CSS**: Utility-first CSS framework via CDN for rapid styling
- **shadcn/ui Inspired**: Custom CSS components using design system patterns
- **Vanilla JavaScript**: Clean, modern JS for authentication and API calls
- **CSS Variables**: Comprehensive design system with light/dark mode support
- **Responsive Design**: Mobile-first approach with proper breakpoints

### Authentication Flow
- JWT tokens for session management
- Password hashing with bcrypt
- Protected routes use auth middleware
- Demo admin user (admin/password) seeded via migration
- Token storage in localStorage with automatic cleanup

### Route Structure
#### Public Routes
- `/health` - API health check
- `/login` - Login page (HTML)
- `/register` - Registration page (HTML)
- `/api/login` - Login API endpoint
- `/api/register` - Registration API endpoint

#### Protected Routes  
- `/api/users` - Get all users (requires JWT)
- `/api/profile` - Get current user profile (requires JWT)
- `/dashboard` - Dashboard page (HTML with JWT validation in JavaScript)

#### Static Files
- `/static/*` - Static files served from `api/public/` directory
- Root `/` redirects to login page

### Database
- SQLite database (default: `./api/data/sqlite/core.db`)
- golang-migrate for database migrations with transaction support
- Migration files in `api/db/migrations/` with sequential numbering
- Users table with bcrypt-hashed passwords
- Demo admin user: `admin/password`

### Frontend Features
- **Modern UI**: Clean, professional design using Tailwind CSS
- **Component System**: Reusable CSS classes following shadcn/ui patterns
- **Authentication**: JWT token handling with automatic redirects
- **Error Handling**: User-friendly error messages and loading states
- **Responsive**: Works on desktop and mobile devices
- **Accessibility**: Proper focus states, semantic HTML, keyboard navigation

### Environment Variables
- `JWT_SECRET`: JWT signing key (default: "your-secret-key-change-this")
- `PORT`: Server port (default: "8080")
- `DATABASE_URL`: SQLite database file path (default: "./data/sqlite/core.db")
- `IMAGES_DIRECTORY`: Directory for uploaded images (default: "./data/images")
- `VIDEOS_DIRECTORY`: Directory for uploaded videos (default: "./data/videos")
- `EMAIL_ENABLED`: Enable email notifications (default: false)
- `AWS_REGION`: AWS region for SES (default: "us-east-1")
- `AWS_ACCESS_KEY_ID`: AWS access key for SES
- `AWS_SECRET_ACCESS_KEY`: AWS secret key for SES
- `SENDER_EMAIL`: From email address (default: "noreply@40weeks.app")
- `SENDER_NAME`: From name (default: "40Weeks")

### Development Workflow
1. Make changes to Go code in `api/` directory
2. Update HTML/CSS in `api/public/` directory
3. Run `go run main.go` from `api/` directory
4. Access frontend at `http://localhost:8080`
5. Test API endpoints at `http://localhost:8080/api/*`

This unified approach provides:
- **Simplicity**: Single service to deploy and maintain
- **Performance**: No network latency between frontend and backend
- **Security**: Templates served securely with CSRF protection
- **Development Speed**: No build steps for frontend, immediate feedback

## Email Notification System

The application includes a comprehensive email notification system powered by AWS SES:

### Features
- **Automatic notifications** for pregnancy updates and milestones
- **Welcome emails** for new village members  
- **Professional HTML templates** with responsive design
- **Delivery tracking** and email analytics
- **Background processing** to avoid blocking API responses

### Email Types
- **Update notifications**: Sent when new pregnancy updates are posted
- **Milestone notifications**: Celebrate pregnancy milestones
- **Welcome emails**: Onboard new village members
- **Test emails**: Verify configuration

### API Endpoints
- `GET /api/email/config-test` - Test SES configuration
- `POST /api/email/test` - Send test email
- `GET /api/email/notifications` - Get notification history
- `GET /api/email/statistics` - Email delivery statistics
- `POST /api/email/send-update` - Manual update notification

### Setup Requirements
1. AWS SES account with verified sender domain/email
2. AWS credentials with SES send permissions
3. Environment variables configured (see above)
4. Set `EMAIL_ENABLED=true` to activate

### Development
When `EMAIL_ENABLED=false`, emails are logged but not sent, making development safe.

## Marketing Notes

### Tagline Options
- Current: "Share Your Pregnancy Journey" 
- Alternative to test: "From Bump to Book" - emphasizes the complete journey from pregnancy tracking to physical keepsake
