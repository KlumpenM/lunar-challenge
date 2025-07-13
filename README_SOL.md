# Lunar Rocket Tracking API

A high-performance rocket tracking system built with Go 1.22+ that handles out-of-order message processing, message deduplication, and real-time rocket state management.

## Features

- **Message Processing**
  - Out-of-order message handling
  - Message deduplication (at-least-once guarantee)
  - Real-time state management
  - Thread-safe concurrent operations

- **API Design**
  - RESTful endpoints with proper HTTP status codes
  - Comprehensive input validation
  - Detailed error responses
  - Debug endpoints for monitoring
  - Go 1.22+ ServeMux with method-specific patterns

- **Production Ready**
  - Graceful shutdown
  - Middleware support (error handling)
  - Comprehensive test coverage
  - Performance benchmarks

## Project Structure

```
.
├── cmd/
│   └── main.go                 # Application entry point
├── internal/
│   ├── api/                    # HTTP handlers and tests
│   ├── errors/                 # Custom error types
│   ├── middleware/             # HTTP middleware
│   ├── models/                 # Data structures
│   ├── sorting/                # Sorting utilities
│   ├── storage/                # Repository implementation
│   └── validation/             # Input validation
├── docs/                       # API documentation (Swagger)
├── test/                       # Integration tests
└── bin/                        # Compiled binaries
```

## Quick Start

### Prerequisites
- Go 1.22 or later

### Installation
Extract the zip file and find the `README_SOL`.
```bash
# Install dependencies
go mod download

# Run server
go run cmd/main.go
```

The server starts on port 8088 with the following endpoints:
- POST /messages - Process rocket messages
- GET /rockets - List all rockets
- GET /rockets/{id} - Get specific rocket
- GET /debug/rockets - Debug info for all rockets
- GET /debug/rockets/{id} - Debug info for specific rocket
- GET /health - Health check

## API Documentation

### Message Processing

Process rocket messages through the `/messages` endpoint:

```json
POST /messages
{
  "metadata": {
    "channel": "rocket-id-12345",
    "messageNumber": 1,
    "messageTime": "2024-01-15T10:30:00Z",
    "messageType": "RocketLaunched"
  },
  "message": {
    "type": "Falcon Heavy",
    "mission": "Mars Mission Alpha",
    "launchSpeed": 1500
  }
}
```

Supported message types:
- RocketLaunched - Initial rocket launch
- RocketSpeedIncreased - Speed increase
- RocketSpeedDecreased - Speed decrease
- RocketMissionChanged - Mission update
- RocketExploded - Rocket failure

### Rocket State Management

Get rocket information:

```bash
# List all rockets
GET /rockets

# Get specific rocket
GET /rockets/{id}

# Get debug information
GET /debug/rockets/{id}
```

### Error Handling

Standard error response format:
```json
{
  "error": {
    "code": 400,
    "message": "Brief error description",
    "details": "Detailed error information"
  }
}
```

HTTP Status Codes:
- 200: Success
- 400: Invalid request/validation error
- 404: Resource not found
- 422: Message processing error
- 500: Server error

## Testing

Run tests:
```bash
# All tests
go test ./...

# Specific packages
go test ./internal/api
go test ./internal/storage
go test ./internal/sorting
go test ./test
```

Test coverage includes:
- Unit tests for all packages
- Integration tests for message flow
- Concurrent operation tests
- Error handling scenarios

## Development

### Code Style
- Follow Go standard conventions
- Use gofmt for formatting
- Run golint for style checking

## Security & Performance

### Current Implementation
- Input validation and sanitization
- JSON payload size limits
- Thread-safe operations
- Graceful error handling


---

Built with Go 1.22+ following modern best practices