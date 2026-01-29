# Go-Gateway

A high-performance API gateway based on Go language, similar to Spring Cloud Gateway functionality.

## Project Features

- 🚀 **High Performance**: Based on Go language's concurrency advantage
- 🔧 **Extensible**: Modular design, easy to extend
- 📝 **Easy Configuration**: Supports JSON format configuration
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
- JSON format configuration file
- Configuration hot update support

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

See [example-config.json](example-config.json) file.

## Usage Instructions

For detailed usage instructions, please refer to [USAGE.md](USAGE.md) document.

## Project Structure

```
go-gateway/
├── main.go                 # 主应用程序入口
├── README.md              # 项目说明
├── USAGE.md               # 使用说明
├── example-config.json    # 示例配置文件
├── start-gateway.bat      # Windows启动脚本
├── go.mod                # Go模块文件
├── go.sum                # Go依赖校验文件
├── pkg/                  # 功能包
│   ├── config/           # 配置管理
│   ├── loadbalancer/     # 负载均衡器
│   ├── middleware/       # 中间件系统
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