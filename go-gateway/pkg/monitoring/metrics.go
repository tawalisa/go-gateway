package monitoring

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	// RequestTotal 请求总数计数器
	RequestTotal *prometheus.CounterVec

	// RequestDuration 请求持续时间直方图
	RequestDuration *prometheus.HistogramVec

	// ActiveConnections 活跃连接数计数器
	ActiveConnections prometheus.Gauge

	// BackendRequestTotal 后端服务请求计数器
	BackendRequestTotal *prometheus.CounterVec

	// RouteHitTotal 路由命中计数器
	RouteHitTotal *prometheus.CounterVec

	// ErrorTotal 错误计数器
	ErrorTotal *prometheus.CounterVec
)

// 初始化监控指标
func init() {
	RequestTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "gateway_requests_total",
			Help: "Total number of requests processed by the gateway",
		},
		[]string{"method", "path", "status"},
	)
	prometheus.MustRegister(RequestTotal)

	RequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "gateway_request_duration_seconds",
			Help:    "Request duration in seconds",
			Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1.0, 2.5, 5.0, 10.0},
		},
		[]string{"method", "path"},
	)
	prometheus.MustRegister(RequestDuration)

	ActiveConnections = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "gateway_active_connections",
			Help: "Current number of active connections",
		},
	)
	prometheus.MustRegister(ActiveConnections)

	BackendRequestTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "gateway_backend_requests_total",
			Help: "Total number of requests to backend services",
		},
		[]string{"backend_url", "route_id"},
	)
	prometheus.MustRegister(BackendRequestTotal)

	RouteHitTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "gateway_route_hits_total",
			Help: "Total number of hits per route",
		},
		[]string{"route_id"},
	)
	prometheus.MustRegister(RouteHitTotal)

	ErrorTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "gateway_errors_total",
			Help: "Total number of errors",
		},
		[]string{"type", "route_id"},
	)
	prometheus.MustRegister(ErrorTotal)
}

// MetricsHandler 返回Prometheus指标处理器
func MetricsHandler() http.Handler {
	return promhttp.Handler()
}
