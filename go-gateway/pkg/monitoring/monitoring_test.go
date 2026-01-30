package monitoring

import (
	"testing"
)

// TestMetricsInitialization 测试指标初始化
func TestMetricsInitialization(t *testing.T) {
	// 验证指标变量已被初始化
	if RequestTotal == nil {
		t.Error("RequestTotal should be initialized")
	}

	if RequestDuration == nil {
		t.Error("RequestDuration should be initialized")
	}

	if ActiveConnections == nil {
		t.Error("ActiveConnections should be initialized")
	}

	if BackendRequestTotal == nil {
		t.Error("BackendRequestTotal should be initialized")
	}

	if RouteHitTotal == nil {
		t.Error("RouteHitTotal should be initialized")
	}

	if ErrorTotal == nil {
		t.Error("ErrorTotal should be initialized")
	}
}

// TestMetricsMiddlewareCreation 测试监控中间件创建
func TestMetricsMiddlewareCreation(t *testing.T) {
	middleware := NewMetricsMiddleware()

	if middleware == nil {
		t.Error("MetricsMiddleware should be created successfully")
	}

	if middleware.Name() != "MetricsMiddleware" {
		t.Errorf("Expected middleware name 'MetricsMiddleware', got '%s'", middleware.Name())
	}
}
