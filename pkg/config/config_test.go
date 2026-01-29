package config

import (
	"os"
	"testing"
)

// Route 定义路由结构
type Route struct {
	ID         string            `json:"id"`
	URI        string            `json:"uri"`
	Predicates []Predicate       `json:"predicates"`
	Filters    []Filter          `json:"filters"`
	Order      int               `json:"order"`
	Metadata   map[string]string `json:"metadata"`
}

// Predicate 定义谓词结构
type Predicate struct {
	Name string      `json:"name"`
	Args interface{} `json:"args"`
}

// Filter 定义过滤器结构
type Filter struct {
	Name string      `json:"name"`
	Args interface{} `json:"args"`
}

// Config 定义配置结构
type Config struct {
	Routes        []Route        `json:"routes"`
	GlobalFilters []GlobalFilter `json:"global_filters"`
	Port          int            `json:"port"`
}

// GlobalFilter 定义全局过滤器
type GlobalFilter struct {
	Name string      `json:"name"`
	Args interface{} `json:"args"`
}

// ConfigManager 配置管理器接口
type ConfigManager interface {
	Load(configPath string) error
	Save(configPath string) error
	GetRoutes() []Route
	AddRoute(route Route)
	UpdateRoute(route Route) error
	DeleteRoute(id string) error
	GetConfig() Config
	SetConfig(config Config)
}

// TestStaticConfigManager 测试静态配置管理器
func TestStaticConfigManager(t *testing.T) {
	t.Run("TestLoadAndSaveConfig", func(t *testing.T) {
		// 创建临时配置文件
		tempConfigFile := "temp_config.json"

		// 创建测试配置
		testConfig := Config{
			Routes: []Route{
				{
					ID:  "test-route-1",
					URI: "http://backend1:8080",
					Predicates: []Predicate{
						{
							Name: "Path",
							Args: map[string]string{"pattern": "/api/test1"},
						},
					},
					Filters: []Filter{
						{
							Name: "RateLimiter",
							Args: map[string]interface{}{
								"permitsPerSecond": float64(100),
								"burstCapacity":    float64(200),
							},
						},
					},
					Order: 1,
				},
				{
					ID:  "test-route-2",
					URI: "http://backend2:8080",
					Predicates: []Predicate{
						{
							Name: "Path",
							Args: map[string]string{"pattern": "/api/test2"},
						},
					},
					Filters: []Filter{},
					Order:   2,
				},
			},
			GlobalFilters: []GlobalFilter{
				{
					Name: "GlobalLogFilter",
				},
				{
					Name: "GlobalMetricsFilter",
				},
			},
			Port: 8080,
		}

		// 创建配置管理器并保存配置
		configMgr := NewStaticConfigManager()
		configMgr.SetConfig(testConfig)

		err := configMgr.Save(tempConfigFile)
		if err != nil {
			t.Fatalf("Failed to save config: %v", err)
		}

		// 验证文件已创建
		if _, err := os.Stat(tempConfigFile); os.IsNotExist(err) {
			t.Fatalf("Config file was not created: %s", tempConfigFile)
		}

		// 创建新的配置管理器并加载配置
		newConfigMgr := NewStaticConfigManager()
		err = newConfigMgr.Load(tempConfigFile)
		if err != nil {
			t.Fatalf("Failed to load config: %v", err)
		}

		// 验证加载的配置
		loadedConfig := newConfigMgr.GetConfig()
		if len(loadedConfig.Routes) != 2 {
			t.Errorf("Expected 2 routes, got %d", len(loadedConfig.Routes))
		}

		if loadedConfig.Port != 8080 {
			t.Errorf("Expected port 8080, got %d", loadedConfig.Port)
		}

		if len(loadedConfig.GlobalFilters) != 2 {
			t.Errorf("Expected 2 global filters, got %d", len(loadedConfig.GlobalFilters))
		}

		// 清理临时文件
		os.Remove(tempConfigFile)
	})

	t.Run("TestAddRoute", func(t *testing.T) {
		configMgr := NewStaticConfigManager()

		initialRoutes := configMgr.GetRoutes()
		if len(initialRoutes) != 0 {
			t.Errorf("Expected empty routes initially, got %d", len(initialRoutes))
		}

		// 添加路由
		newRoute := Route{
			ID:  "new-test-route",
			URI: "http://new-backend:8080",
			Predicates: []Predicate{
				{
					Name: "Path",
					Args: map[string]string{"pattern": "/api/new"},
				},
			},
			Order: 1,
		}

		configMgr.AddRoute(newRoute)

		updatedRoutes := configMgr.GetRoutes()
		if len(updatedRoutes) != 1 {
			t.Errorf("Expected 1 route after adding, got %d", len(updatedRoutes))
		}

		if updatedRoutes[0].ID != "new-test-route" {
			t.Errorf("Expected route ID 'new-test-route', got %s", updatedRoutes[0].ID)
		}
	})

	t.Run("TestDeleteRoute", func(t *testing.T) {
		configMgr := NewStaticConfigManager()

		// 添加几个路由
		route1 := Route{
			ID:  "route-to-delete",
			URI: "http://backend1:8080",
			Predicates: []Predicate{
				{
					Name: "Path",
					Args: map[string]string{"pattern": "/api/delete"},
				},
			},
			Order: 1,
		}

		route2 := Route{
			ID:  "route-to-keep",
			URI: "http://backend2:8080",
			Predicates: []Predicate{
				{
					Name: "Path",
					Args: map[string]string{"pattern": "/api/keep"},
				},
			},
			Order: 2,
		}

		configMgr.AddRoute(route1)
		configMgr.AddRoute(route2)

		initialRoutes := configMgr.GetRoutes()
		if len(initialRoutes) != 2 {
			t.Errorf("Expected 2 routes initially, got %d", len(initialRoutes))
		}

		// 删除路由
		err := configMgr.DeleteRoute("route-to-delete")
		if err != nil {
			t.Fatalf("Failed to delete route: %v", err)
		}

		remainingRoutes := configMgr.GetRoutes()
		if len(remainingRoutes) != 1 {
			t.Errorf("Expected 1 route after deletion, got %d", len(remainingRoutes))
		}

		if remainingRoutes[0].ID != "route-to-keep" {
			t.Errorf("Expected to keep route 'route-to-keep', got %s", remainingRoutes[0].ID)
		}

		// 尝试删除不存在的路由
		err = configMgr.DeleteRoute("non-existent-route")
		if err == nil {
			t.Error("Expected error when deleting non-existent route")
		}
	})

	t.Run("TestUpdateRoute", func(t *testing.T) {
		configMgr := NewStaticConfigManager()

		// 添加路由
		initialRoute := Route{
			ID:  "updatable-route",
			URI: "http://old-backend:8080",
			Predicates: []Predicate{
				{
					Name: "Path",
					Args: map[string]string{"pattern": "/api/update"},
				},
			},
			Order: 1,
		}

		configMgr.AddRoute(initialRoute)

		// 更新路由
		updatedRoute := Route{
			ID:  "updatable-route",
			URI: "http://new-backend:8080",
			Predicates: []Predicate{
				{
					Name: "Path",
					Args: map[string]string{"pattern": "/api/updated"},
				},
			},
			Filters: []Filter{
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

		// 验证更新后的路由
		routes := configMgr.GetRoutes()
		if len(routes) != 1 {
			t.Errorf("Expected 1 route, got %d", len(routes))
		}

		updated := routes[0]
		if updated.URI != "http://new-backend:8080" {
			t.Errorf("Expected URI 'http://new-backend:8080', got %s", updated.URI)
		}

		if updated.Order != 2 {
			t.Errorf("Expected order 2, got %d", updated.Order)
		}

		if len(updated.Filters) != 1 {
			t.Errorf("Expected 1 filter, got %d", len(updated.Filters))
		}

		// 尝试更新不存在的路由
		nonExistentRoute := Route{
			ID:  "non-existent-route",
			URI: "http://dummy:8080",
		}

		err = configMgr.UpdateRoute(nonExistentRoute)
		if err == nil {
			t.Error("Expected error when updating non-existent route")
		}
	})
}
