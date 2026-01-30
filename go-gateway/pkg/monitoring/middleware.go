package monitoring

import (
	"time"

	"go-gateway/pkg/middleware"
)

// MetricsMiddleware 监控中间件，用于收集请求指标
type MetricsMiddleware struct{}

// NewMetricsMiddleware 创建监控中间件实例
func NewMetricsMiddleware() *MetricsMiddleware {
	return &MetricsMiddleware{}
}

// Name 返回中间件名称
func (mm *MetricsMiddleware) Name() string {
	return "MetricsMiddleware"
}

// PreHandle 预处理请求并收集指标
func (mm *MetricsMiddleware) PreHandle(ctx *middleware.GatewayContext) bool {
	// 记录开始时间
	ctx.Attributes["start_time"] = time.Now()

	// 记录活跃连接数增加
	ActiveConnections.Inc()

	return true // 继续执行后续中间件
}

// PostHandle 后处理请求并收集指标
func (mm *MetricsMiddleware) PostHandle(ctx *middleware.GatewayContext) error {
	// 获取开始时间
	startTime, ok := ctx.Attributes["start_time"].(time.Time)
	if !ok {
		startTime = time.Now()
	}

	// 计算请求持续时间
	duration := time.Since(startTime).Seconds()

	// 获取路由ID，如果可用
	routeID := "unknown"
	if ctx.Route != nil {
		routeID = ctx.Route.ID
	}

	// 记录请求持续时间
	RequestDuration.WithLabelValues(
		ctx.Request.Method,
		ctx.Request.URL.Path,
	).Observe(duration)

	// 记录请求总数 - 使用适当的HTTP状态码
	RequestTotal.WithLabelValues(
		ctx.Request.Method,
		ctx.Request.URL.Path,
		"200",
	).Inc()

	// 记录路由命中
	RouteHitTotal.WithLabelValues(routeID).Inc()

	return nil
}

// HandleError 处理错误
func (mm *MetricsMiddleware) HandleError(ctx *middleware.GatewayContext, err error) {
	// 记录错误
	routeID := "unknown"
	if ctx.Route != nil {
		routeID = ctx.Route.ID
	}

	ErrorTotal.WithLabelValues("middleware_error", routeID).Inc()
}
