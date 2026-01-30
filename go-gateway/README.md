# Go-Gateway

A high-performance API gateway based on Go language, similar to Spring Cloud Gateway functionality.

## Project Features

- 🚀 **High Performance**: Based on Go language's concurrency advantage
- 🔧 **Extensible**: Modular design, easy to extend
- 📝 **Easy Configuration**: Supports multiple formats (JSON, YAML, TOML, etc.) via Viper
- 📊 **Observability**: Built-in monitoring and logging functions
- ✅ **High Reliability**: Complete test coverage

## Architecture Design

```
客户端请求 -> 路由匹配 -> 中间件链 -> 负载均衡 -> 目标服务
     ↑                                    ↓
   认证授权                            响应处理
     ↑                                    ↓
   限流控制                           日志记录
                                        ↓
                                     监控上报
```

## Core Functions

### 1. Route Management
- Supports path matching (exact match, wildcard match)
- Supports route priority
- Supports dynamic route configuration

### 2. Load Balancing
- Round Robin Algorithm
- Random Algorithm
- Weighted Round Robin Algorithm

### 3. Middleware System
- Pluggable middleware mechanism
- Supports pre/post processing
- Middleware chained calls

### 4. Configuration Management
- **Multiple Format Support**: Supports JSON, YAML, TOML, INI, env files via Viper
- **Configuration Hot Update**: Automatically reloads configuration upon file changes
- **Remote Configuration**: Supports remote configuration sources (etcd, Consul, etc.)
- **Environment Variables**: Seamlessly integrates with environment variables
- **Default Values**: Supports default configuration values

### 5. Monitoring System
- **Prometheus Integration**: Exposes metrics in Prometheus format
- **Key Metrics**: Request count, response time, active connections, route hits, error rates
- **Monitoring Endpoint**: Available at `/metrics` on port 9090 by default
- **Grafana Ready**: Metrics formatted for easy visualization with Grafana

## 快速开始

### Prerequisites
- Go 1.19+

### Important Note

**Please pay attention to port configuration**: The gateway listens on port 8080 by default, you need to ensure this port is not occupied by other services. Also, the gateway listening port and backend service port must be different to avoid circular calls or port conflicts.

### Installation and Running

1. Clone project
```bash
git clone <repository-url>
cd go-gateway/go-gateway
```

2. Build project
```bash
go build -o gateway .
```

3. Run gateway
```bash
./gateway
```

Or run directly:
```bash
go run .
```

The gateway listens on port 8080 by default, you can modify this setting through configuration file.

## Configuration Example

See [example-config.json](example-config.json) or [example-viper-config.json](example-viper-config.json) file.

## Monitoring

The gateway exposes Prometheus metrics at `http://localhost:9090/metrics`. Key metrics include:

- `gateway_requests_total`: Total number of requests processed
- `gateway_request_duration_seconds`: Request duration histogram
- `gateway_active_connections`: Current number of active connections
- `gateway_backend_requests_total`: Total requests to backend services
- `gateway_route_hits_total`: Hits per route
- `gateway_errors_total`: Error counts by type

For more details about monitoring, see [MONITORING_GUIDE.md](MONITORING_GUIDE.md).

## Usage Instructions

For detailed usage instructions, please refer to [USAGE.md](USAGE.md) document.

## Project Structure

```
go-gateway/
├── main.go                 # 主应用程序入口
├── README.md              # 项目说明
├── USAGE.md               # 使用说明
├── CONFIG_GUIDE.md        # 配置管理指南
├── MONITORING_GUIDE.md    # 监控系统指南
├── example-config.json    # 示例配置文件
├── example-viper-config.json # Viper配置示例文件
├── prometheus.yml         # Prometheus配置示例
├── start-gateway.bat      # Windows启动脚本
├── go.mod                # Go模块文件
├── go.sum                # Go依赖校验文件
├── pkg/                  # 功能包
│   ├── common/           # 公共类型定义
│   ├── config/           # 配置管理
│   ├── loadbalancer/     # 负载均衡器
│   ├── middleware/       # 中间件系统
│   ├── monitoring/       # 监控系统
│   └── route/            # 路由管理
└── tests/                # 测试文件
```

## Testing

Run all tests:
```bash
go test ./... -v
```

## Contribution

Welcome to submit Issues and Pull Requests to improve the project.

## License

MIT License