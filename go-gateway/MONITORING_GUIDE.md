# 监控系统指南

本文档介绍如何使用Prometheus监控Go-Gateway。

## 特性

Go-Gateway内置了对Prometheus监控的支持，提供以下关键指标：

- **请求计数**: 记录所有传入请求的数量
- **响应时间**: 记录请求处理的时间分布
- **活跃连接数**: 当前活跃的连接数
- **后端请求**: 记录发送到后端服务的请求数
- **路由命中率**: 每个路由的访问次数
- **错误统计**: 不同类型的错误计数

## 指标详情

### gateway_requests_total
- 类型: Counter
- 标签: method, path, status
- 描述: 网关处理的总请求数

### gateway_request_duration_seconds
- 类型: Histogram
- 标签: method, path
- 描述: 请求处理时间（秒），包含预设的时间桶

### gateway_active_connections
- 类型: Gauge
- 描述: 当前活跃连接数

### gateway_backend_requests_total
- 类型: Counter
- 标签: backend_url, route_id
- 描述: 发送到后端服务的总请求数

### gateway_route_hits_total
- 类型: Counter
- 标签: route_id
- 描述: 每个路由的命中次数

### gateway_errors_total
- 类型: Counter
- 标签: type, route_id
- 描述: 错误计数，按类型和路由分组

## 配置Prometheus

要将Go-Gateway与Prometheus集成，请在Prometheus配置文件中添加以下job：

```yaml
scrape_configs:
  - job_name: 'go-gateway'
    static_configs:
      - targets: ['localhost:9090']
    metrics_path: '/metrics'
    scrape_interval: 5s
    scrape_timeout: 4s
```

## Grafana仪表板

您可以创建一个Grafana仪表板来可视化这些指标。以下是一些有用的查询示例：

### 请求速率
```
rate(gateway_requests_total[5m])
```

### 平均响应时间
```
rate(gateway_request_duration_seconds_sum[5m]) / rate(gateway_request_duration_seconds_count[5m])
```

### 错误率
```
rate(gateway_errors_total[5m])
```

### 活跃连接数
```
gateway_active_connections
```

## 启动监控

当您启动Go-Gateway时，监控服务会在端口9090上自动启动，并在`/metrics`路径下暴露指标。

直接访问 `http://localhost:9090/metrics` 来查看原始指标数据。

## 集成到现有系统

如果您希望自定义监控配置，可以在代码中这样做：

```go
// 创建监控服务
monitoringService := monitoring.NewMonitoringService(9090)

// 启动监控服务
go func() {
    if err := monitoringService.Start(); err != nil {
        log.Printf("Monitoring server error: %v", err)
    }
}()

// 添加监控中间件
metricsMiddleware := monitoring.NewMetricsMiddleware()
gateway.middlewares = append(gateway.middlewares, metricsMiddleware)
```