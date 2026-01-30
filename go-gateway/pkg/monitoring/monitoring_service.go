package monitoring

import (
	"context"
	"fmt"
	"log"
	"net/http"
)

// MonitoringService 监控服务
type MonitoringService struct {
	server *http.Server
	port   int
}

// NewMonitoringService 创建监控服务实例
func NewMonitoringService(port int) *MonitoringService {
	return &MonitoringService{
		port: port,
	}
}

// Start 启动监控服务
func (ms *MonitoringService) Start() error {
	mux := http.NewServeMux()
	mux.Handle("/metrics", MetricsHandler())

	addr := fmt.Sprintf(":%d", ms.port)
	ms.server = &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	log.Printf("Starting monitoring service on %s", addr)
	return ms.server.ListenAndServe()
}

// Stop 停止监控服务
func (ms *MonitoringService) Stop(ctx context.Context) error {
	if ms.server != nil {
		return ms.server.Shutdown(ctx)
	}
	return nil
}

// StartAsync 异步启动监控服务
func (ms *MonitoringService) StartAsync() {
	go func() {
		if err := ms.Start(); err != nil && err != http.ErrServerClosed {
			log.Printf("Monitoring server error: %v", err)
		}
	}()
}
