# Go-Gateway Configuration Guide

## Port Configuration Notes

### Important Reminder
**Gateway listening port (port) must be different from backend service port (uri)!**

If the gateway listening port and backend service port are the same, the following issues will occur:
- Circular calls: Gateway requests processed by itself
- Port conflicts: Cannot bind to occupied port
- Service unavailable: Requests cannot be properly forwarded

### Correct Configuration Example

```json
{
  "routes": [
    {
      "id": "service-a",
      "uri": "http://localhost:9090",  // Backend service uses port 9090
      "predicates": [
        {
          "name": "Path",
          "args": {
            "pattern": "/api/service-a/**"
          }
        }
      ]
    }
  ],
  "port": 8080  // Gateway listens on port 8080
}
```

### Incorrect Configuration Example

```json
{
  "routes": [
    {
      "id": "service-a",
      "uri": "http://localhost:8080",  // Error: Backend service also on port 8080
      "predicates": [
        {
          "name": "Path",
          "args": {
            "pattern": "/api/service-a/**"
          }
        }
      ]
    }
  ],
  "port": 8080  // Gateway also on port 8080, causing conflict
}
```

## Configuration File Details

### routes - Route Configuration
- `id`: Unique route identifier
- `uri`: Backend service address, format `http://host:port`
- `predicates`: Matching conditions, currently supports Path predicate
- `filters`: Filter list
- `order`: Priority, smaller number means higher priority
- `metadata`: Metadata information

### predicates - Matching Conditions

#### Path Predicate
```json
{
  "name": "Path",
  "args": {
    "pattern": "/api/users/**"
  }
}
```

Supported matching patterns:
- `/exact/path` - Exact match
- `/api/*` - Single-level wildcard match
- `/api/**` - Multi-level wildcard match

### filters - Filters

#### RateLimiter
```json
{
  "name": "RateLimiter",
  "args": {
    "permitsPerSecond": 100,
    "burstCapacity": 200
  }
}
```

### global_filters - Global Filters
Filters that apply to all requests.

### port - Listening Port
Port number the gateway service listens on.

## Common Configuration Scenarios

### Scenario 1: Multiple Microservice Routes
```json
{
  "routes": [
    {
      "id": "users-service",
      "uri": "http://users-service:8081",
      "predicates": [
        {
          "name": "Path",
          "args": {
            "pattern": "/api/users/**"
          }
        }
      ],
      "order": 1
    },
    {
      "id": "orders-service",
      "uri": "http://orders-service:8082",
      "predicates": [
        {
          "name": "Path",
          "args": {
            "pattern": "/api/orders/**"
          }
        }
      ],
      "order": 2
    }
  ],
  "port": 8080
}
```

### Scenario 2: Routes with Rate Limiting
```json
{
  "routes": [
    {
      "id": "protected-api",
      "uri": "http://backend:9000",
      "predicates": [
        {
          "name": "Path",
          "args": {
            "pattern": "/api/protected/**"
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
      ]
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

## Startup Configuration

### Method 1: Modify main.go
Edit the [main.go](file://D:\code\go-gateway\go-gateway\main.go) file to add configuration loading logic:

```go
func main() {
    gateway := NewGateway()
    
    // Load external configuration file
    err := gateway.LoadConfig("example-config.json")
    if err != nil {
        log.Fatal("Failed to load config: ", err)
    }
    
    port := gateway.configManager.GetConfig().Port
    log.Printf("Starting gateway on :%d", port)
    if err := gateway.Run(port); err != nil {
        log.Fatal("Gateway failed to start: ", err)
    }
}
```

### Method 2: Using Default Configuration
If no external configuration file is loaded, the gateway will use default configuration with empty route list, requiring route rules to be defined through external configuration file.

## Troubleshooting

### Issue 1: Port Occupied
**Phenomenon**: `bind: address already in use`
**Solution**: Change gateway listening port or stop other services occupying the port

### Issue 2: Circular Calls
**Phenomenon**: Request timeout or infinite redirection
**Solution**: Check route configuration to ensure backend service port is different from gateway listening port

### Issue 3: Cannot Connect to Backend Service
**Phenomenon**: `connection refused`
**Solution**: Confirm backend service is running and network is reachable