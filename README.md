# Simple Go Service

A minimal Go HTTP service with a health check endpoint.

## Running Locally

### Prerequisites
- Go 1.21 or higher installed

### Steps

1. Clone the repository:
```bash
git clone <repository-url>
cd simple-go
```

2. Run the service:
```bash
go run main.go
```

The service will start on port 8080.

### Testing the Endpoint

Once the service is running, you can test the health endpoint:

```bash
curl http://localhost:8080/health
```

Expected response:
```json
{
  "message": "Hello from Go service!",
  "status": "healthy"
}
```

## API Endpoints

- `GET /health` - Returns a JSON response indicating the service status