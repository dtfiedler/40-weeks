# Build stage
FROM golang:1.20-bullseye AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=1 go build -o simple-go main.go

# Final stage
FROM debian:bullseye-slim

# Install runtime dependencies
RUN apt-get update && apt-get install -y \
    ca-certificates \
    sqlite3 \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /root/

# Copy the binary from builder stage
COPY --from=builder /app/simple-go .

# Copy public directory for static files
COPY --from=builder /app/public ./public

# Copy database migrations
COPY --from=builder /app/db ./db

# Copy migration script
COPY --from=builder /app/migrate.sh .
RUN chmod +x migrate.sh

# Create directory for database
RUN mkdir -p /data

# Expose port
EXPOSE 8080

# Set environment variables
ENV DATABASE_URL=/data/users.db
ENV PORT=8080

# Run migrations and start the application
CMD ["./simple-go"]