package main

import (
	"encoding/json"
	"net/http/httptest"
	"testing"

	"go-gateway/pkg/config"
	"go-gateway/pkg/loadbalancer"
	"go-gateway/pkg/middleware"
	"go-gateway/pkg/route"
)

// TestGatewayIntegration 测试网关集成
func TestGatewayIntegration(t *testing.T) {
	t.Run("TestBasicRoutingAndLoadBalancing", func(t *testing.T) {
		// 创建配置管理器
		configMgr := config.NewStaticConfigManager()

		// 设置测试配置
		testConfig := config.Config{
			Routes: []config.Route{
				{
					ID:  "service-a",
					URI: "lb://service-a", // 使用负载均衡标识
					Predicates: []config.Predicate{
						{
							Name: "Path",
							Args: map[string]string{"pattern": "/api/service-a/**"},
						},
					},
					Filters: []config.Filter{
						{
							Name: "RateLimiter",
							Args: map[string]interface{}{
								"permitsPerSecond": 100.0,
								"burstCapacity":    200.0,
							},
						},
					},
					Order: 1,
				},
				{
					ID:  "service-b",
					URI: "http://specific-backend:8080",
					Predicates: []config.Predicate{
						{
							Name: "Path",
							Args: map[string]string{"pattern": "/api/service-b/**"},
						},
					},
					Order: 2,
				},
			},
			Port: 8080,
		}

		configMgr.SetConfig(testConfig)

		// 创建路由器
		router := route.NewRouter()

		// 从配置加载路由
		for _, routeConfig := range configMgr.GetRoutes() {
			// 转换配置路由到内部路由结构
			internalRoute := &route.Route{
				ID:         routeConfig.ID,
				URI:        routeConfig.URI,
				Predicates: convertPredicates(routeConfig.Predicates),
				Filters:    convertFilters(routeConfig.Filters),
				Order:      routeConfig.Order,
				Metadata:   routeConfig.Metadata,
			}
			router.AddRoute(internalRoute)
		}

		// 创建负载均衡器
		lb := loadbalancer.NewRoundRobinBalancer()

		// 添加后端服务器
		servers := []loadbalancer.Server{
			{URL: "http://backend1:8080", Weight: 1},
			{URL: "http://backend2:8080", Weight: 1},
		}

		for _, server := range servers {
			lb.AddServer(server)
		}

		// 测试路由匹配
		matchedRoute := router.Match("/api/service-a/test")
		if matchedRoute == nil {
			t.Errorf("Expected route to match /api/service-a/test, got nil")
		} else if matchedRoute.ID != "service-a" {
			t.Errorf("Expected route ID 'service-a', got %s", matchedRoute.ID)
		}

		// 测试负载均衡选择
		chosenServer := lb.ChooseServer(lb.GetServers())
		if chosenServer == nil {
			t.Errorf("Expected a server to be chosen, got nil")
		}

		// 验证所选服务器在列表中
		found := false
		for _, server := range servers {
			if server.URL == chosenServer.URL {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Chosen server %s not in original server list", chosenServer.URL)
		}
	})

	t.Run("TestRouteConfigurationManagement", func(t *testing.T) {
		// 创建配置管理器
		configMgr := config.NewStaticConfigManager()

		// 初始路由数量
		initialRoutes := configMgr.GetRoutes()
		if len(initialRoutes) != 0 {
			t.Errorf("Expected 0 routes initially, got %d", len(initialRoutes))
		}

		// 添加路由
		newRoute := config.Route{
			ID:  "dynamic-route",
			URI: "http://dynamic-backend:8080",
			Predicates: []config.Predicate{
				{
					Name: "Path",
					Args: map[string]string{"pattern": "/api/dynamic/**"},
				},
			},
			Order: 1,
		}

		configMgr.AddRoute(newRoute)

		// 验证路由已添加
		updatedRoutes := configMgr.GetRoutes()
		if len(updatedRoutes) != 1 {
			t.Errorf("Expected 1 route after adding, got %d", len(updatedRoutes))
		}

		// 更新路由
		updatedRoute := config.Route{
			ID:  "dynamic-route",
			URI: "http://updated-backend:8080",
			Predicates: []config.Predicate{
				{
					Name: "Path",
					Args: map[string]string{"pattern": "/api/updated/**"},
				},
			},
			Filters: []config.Filter{
				{
					Name: "AuthFilter",
					Args: map[string]interface{}{"required": true},
				},
			},
			Order: 2,
		}

		err := configMgr.UpdateRoute(updatedRoute)
		if err != nil {
			t.Fatalf("Failed to update route: %v", err)
		}

		// 验证路由已更新
		finalRoutes := configMgr.GetRoutes()
		if len(finalRoutes) != 1 {
			t.Errorf("Expected 1 route after update, got %d", len(finalRoutes))
		}

		if finalRoutes[0].URI != "http://updated-backend:8080" {
			t.Errorf("Expected updated URI, got %s", finalRoutes[0].URI)
		}

		if finalRoutes[0].Order != 2 {
			t.Errorf("Expected updated order 2, got %d", finalRoutes[0].Order)
		}
	})

	t.Run("TestMiddlewareIntegration", func(t *testing.T) {
		// 创建网关上下文
		req := httptest.NewRequest("GET", "http://localhost/api/test", nil)
		resp := httptest.NewRecorder()

		gatewayCtx := &middleware.GatewayContext{
			Request:    req,
			Response:   resp,
			Attributes: make(map[string]interface{}),
		}

		// 验证上下文创建成功
		if gatewayCtx.Request == nil {
			t.Error("Expected request in gateway context, got nil")
		}

		if gatewayCtx.Response == nil {
			t.Error("Expected response in gateway context, got nil")
		}

		if gatewayCtx.Attributes == nil {
			t.Error("Expected attributes map in gateway context, got nil")
		}
	})
}

// 辅助函数：转换谓词
func convertPredicates(predicates []config.Predicate) []route.Predicate {
	result := make([]route.Predicate, len(predicates))
	for i, p := range predicates {
		result[i] = route.Predicate{
			Name: p.Name,
			Args: p.Args,
		}
	}
	return result
}

// 辅助函数：转换过滤器
func convertFilters(filters []config.Filter) []route.Filter {
	result := make([]route.Filter, len(filters))
	for i, f := range filters {
		result[i] = route.Filter{
			Name: f.Name,
			Args: f.Args,
		}
	}
	return result
}

// TestConfigSerialization 测试配置序列化
func TestConfigSerialization(t *testing.T) {
	// 创建测试配置
	testConfig := config.Config{
		Routes: []config.Route{
			{
				ID:  "serialized-route",
				URI: "http://serialized-backend:8080",
				Predicates: []config.Predicate{
					{
						Name: "Path",
						Args: map[string]string{"pattern": "/api/serialize/**"},
					},
				},
				Filters: []config.Filter{
					{
						Name: "RateLimiter",
						Args: map[string]interface{}{
							"permitsPerSecond": 50.0,
							"burstCapacity":    100.0,
						},
					},
				},
				Order: 1,
				Metadata: map[string]string{
					"description": "Test route for serialization",
					"version":     "1.0",
				},
			},
		},
		GlobalFilters: []config.GlobalFilter{
			{
				Name: "GlobalLogFilter",
			},
			{
				Name: "GlobalMetricsFilter",
				Args: map[string]interface{}{"enabled": true},
			},
		},
		Port: 9090,
	}

	// 序列化配置
	jsonData, err := json.MarshalIndent(testConfig, "", "  ")
	if err != nil {
		t.Fatalf("Failed to serialize config: %v", err)
	}

	// 反序列化配置
	var deserializedConfig config.Config
	err = json.Unmarshal(jsonData, &deserializedConfig)
	if err != nil {
		t.Fatalf("Failed to deserialize config: %v", err)
	}

	// 验证反序列化的配置
	if deserializedConfig.Port != 9090 {
		t.Errorf("Expected port 9090, got %d", deserializedConfig.Port)
	}

	if len(deserializedConfig.Routes) != 1 {
		t.Errorf("Expected 1 route, got %d", len(deserializedConfig.Routes))
	}

	if len(deserializedConfig.GlobalFilters) != 2 {
		t.Errorf("Expected 2 global filters, got %d", len(deserializedConfig.GlobalFilters))
	}

	// 验证特定路由属性
	route := deserializedConfig.Routes[0]
	if route.ID != "serialized-route" {
		t.Errorf("Expected route ID 'serialized-route', got %s", route.ID)
	}

	if route.URI != "http://serialized-backend:8080" {
		t.Errorf("Expected URI 'http://serialized-backend:8080', got %s", route.URI)
	}

	// 验证元数据
	description, exists := route.Metadata["description"]
	if !exists || description != "Test route for serialization" {
		t.Errorf("Expected metadata 'description' to be 'Test route for serialization', got %s", description)
	}

	version, exists := route.Metadata["version"]
	if !exists || version != "1.0" {
		t.Errorf("Expected metadata 'version' to be '1.0', got %s", version)
	}
}
