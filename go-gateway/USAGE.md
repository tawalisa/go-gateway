# Go-Gateway Usage Guide

## Project Overview

Go-Gateway is a high-performance API gateway similar to Spring Cloud Gateway, developed using Go language. The gateway provides routing, load balancing, authentication, and rate limiting functions.

## Directory Structure

```
go-gateway/
├── main.go             # Main application entry point
├── USAGE.md            # Usage guide
├── example-config.json # Example configuration file
├── go.mod             # Go module file
├── go.sum             # Go dependency checksums
├── core/              # Core components
├── pkg/               # Feature packages
│   ├── config/        # Configuration management
│   ├── loadbalancer/  # Load balancer
│   ├── middleware/    # Middleware system
│   └── route/         # Routing management
└── tests/             # Test files
```

## Quick Start

### Prerequisites

- Go 1.19+
- Windows/Linux/macOS

### Build Project

```bash
# Enter project directory
cd D:\code\go-gateway\go-gateway

# Build project
go build -o gateway.exe .

# Or run directly
go run .
```

### Running Gateway

```bash
# Run directly
go run .

# Or run built executable
./gateway.exe
```

The gateway will start on `:8080` port.

## Configuration Guide

The gateway supports configuration via JSON configuration file. The configuration file includes the following sections:

### Route Configuration (Routes)

```json
{
  "routes": [
    {
      "id": "service-a",
      "uri": "http://localhost:9090",  // Target service address
      "predicates": [                  // Matching conditions
        {
          "name": "Path",
          "args": {
            "pattern": "/api/service-a/**"
          }
        }
      ],
      "filters": [                     // Filters
        {
          "name": "RateLimiter",
          "args": {
            "permitsPerSecond": 100,
            "burstCapacity": 200
          }
        }
      ],
      "order": 1,                      // Priority
      "metadata": {                    // Metadata
        "description": "Service A route"
      }
    }
  ]
}
```

### Route Field Description

- `id`: Unique route identifier
- `uri`: Target service address
  - `http://host:port` - Directly specify service address
  - `lb://service-name` - Use load balancing (service discovery not fully implemented yet)
- `predicates`: Matching condition array
  - `name`: Predicate name (currently only supports Path)
  - `args`: Arguments object
- `filters`: Filter array
- `order`: Priority (smaller number means higher priority)
- `metadata`: Metadata information

### Path Matching Patterns

- `/exact/path` - Exact match
- `/api/*` - Single-level wildcard match
- `/api/**` - Multi-level wildcard match (matches all sub-paths)

### Global Filters

```json
{
  "global_filters": [
    {
      "name": "GlobalLogFilter"
    },
    {
      "name": "GlobalMetricsFilter",
      "args": {
        "enabled": true
      }
    }
  ]
}
```

## Features

### 1. Route Matching

- Supports exact path matching
- Supports wildcard matching (`*` and `**`)
- Supports route priority sorting

### 2. Load Balancer

- **Round Robin Algorithm** - Distribute requests sequentially
- **Random Algorithm** - Randomly select backend services
- **Weighted Round Robin** - Distribute requests by weight

### 3. Middleware System

- Pluggable middleware mechanism
- Supports pre and post processing
- Supports middleware chaining

### 4. Configuration Management

- Supports JSON format configuration
- Configuration hot update (requires restart in current version)
- Configuration validation

## Usage Examples

### Example Configuration File

```json
{
  "routes": [
    {
      "id": "users-service",
      "uri": "http://localhost:8081",
      "predicates": [
        {
          "name": "Path",
          "args": {
            "pattern": "/api/users/**"
          }
        }
      ],
      "filters": [],
      "order": 1,
      "metadata": {
        "description": "Users service endpoint"
      }
    },
    {
      "id": "orders-service",
      "uri": "http://localhost:8082",
      "predicates": [
        {
          "name": "Path",
          "args": {
            "pattern": "/api/orders/**"
          }
        }
      ],
      "filters": [
        {
          "name": "RateLimiter",
          "args": {
            "permitsPerSecond": 50,
            "burstCapacity": 100
          }
        }
      ],
      "order": 2,
      "metadata": {
        "description": "Orders service endpoint with rate limiting"
      }
    }
  ],
  "global_filters": [
    {
      "name": "GlobalLogFilter"
    }
  ],
  "port": 8080
}
```

### Startup Configuration

To use a custom configuration file to start the gateway, modify the configuration loading logic in [main.go](file://D:\code\go-gateway\go-gateway\main.go):

```go
func main() {
    gateway := NewGateway()
    
    // Load configuration file
    err := gateway.LoadConfig("example-config.json")
    if err != nil {
        log.Fatal("Failed to load config: ", err)
    }
    
    log.Println("Starting gateway on :8080")
    if err := gateway.Run(8080); err != nil {
        log.Fatal("Gateway failed to start: ", err)
    }
}
```

## Development Guide

### Adding New Route

```go
// Create route configuration
routeConfig := config.Route{
    ID:  "new-service",
    URI: "http://localhost:9090",
    Predicates: []config.Predicate{
        {
            Name: "Path",
            Args: map[string]string{"pattern": "/api/new-service/**"},
        },
    },
    Order: 3,
}

// Add to configuration manager
configManager.AddRoute(routeConfig)
```

### Adding New Middleware

```go
// Implement middleware interface
type CustomMiddleware struct{}

func (cm *CustomMiddleware) Name() string {
    return "CustomMiddleware"
}

func (cm *CustomMiddleware) PreHandle(ctx *middleware.GatewayContext) bool {
    // Pre-processing logic
    return true
}

func (cm *CustomMiddleware) PostHandle(ctx *middleware.GatewayContext) error {
    // Post-processing logic
    return nil
}

func (cm *CustomMiddleware) HandleError(ctx *middleware.GatewayContext, err error) {
    // Error handling logic
}
```

## Testing

### Run All Tests

```bash
# Run all tests
go test ./... -v

# Run tests for specific package
go test ./pkg/route/... -v
go test ./pkg/middleware/... -v
go test ./pkg/loadbalancer/... -v
go test ./pkg/config/... -v
go test ./tests/... -v
```

### Test Coverage

- Router: Path matching, priority sorting
- Middleware System: Middleware chains, pre/post processing
- Load Balancer: Round-robin, random, weighted round-robin algorithms
- Configuration Manager: Loading, saving, CRUD operations
- Integration Tests: Component collaboration validation

## Deployment

### Production Deployment

1. Build executable file
```bash
go build -o gateway .
```

2. Prepare configuration file
3. Start service
```bash
./gateway
```

### Docker Deployment (Future Implementation)

Consider adding Docker support in the future:

```dockerfile
FROM golang:alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o gateway .

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/gateway .
COPY --from=builder /app/example-config.json .
CMD ["./gateway"]
```

## API Interface

As a reverse proxy, the gateway forwards requests to corresponding backend services based on configured routing rules.

For example, if configured route:
```json
{
  "id": "api-service",
  "uri": "http://backend:8080",
  "predicates": [{
    "name": "Path",
    "args": {"pattern": "/api/**"}
  }]
}
```

Then all requests sent to `/api/*` will be forwarded to `http://backend:8080`.

## Troubleshooting

### Common Issues

1. **Port Occupied**
   - Check if port is occupied by other processes
   - Modify port number in configuration file

2. **Route Not Working**
   - Check if path matching pattern is correct
   - Confirm route priority settings

3. **Cannot Connect to Backend Service**
   - Confirm backend service is running normally
   - Check network connectivity

### Log Viewing

Current version logs output to standard output, which can be viewed as follows:

```bash
./gateway 2>&1 | tee gateway.log
```

## Version Information

- **Version**: 1.0.0
- **Development Language**: Go
- **Framework**: Native Go net/http
- **License**: MIT

## Contributing

Welcome to submit Issues and Pull Requests to improve the project.

## Acknowledgments

Thanks to all developers who contributed to the project.