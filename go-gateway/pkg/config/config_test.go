package config

import (
	"os"
	"testing"

	"go-gateway/pkg/common"
)

// TestStaticConfigManager tests static config manager
func TestStaticConfigManager(t *testing.T) {
	t.Run("TestLoadAndSaveConfig", func(t *testing.T) {
		// Create temporary config file
		tempConfigFile := "temp_config.json"

		// Create test config
		testConfig := Config{
			Routes: []common.Route{
				{
					ID:  "test-route-1",
					URI: "http://backend1:8080",
					Predicates: []common.Predicate{
						{
							Name: "Path",
							Args: map[string]string{"pattern": "/api/test1"},
						},
					},
					Filters: []common.Filter{
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
					Predicates: []common.Predicate{
						{
							Name: "Path",
							Args: map[string]string{"pattern": "/api/test2"},
						},
					},
					Filters: []common.Filter{},
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

		// Create config manager and save config
		configMgr := NewStaticConfigManager()
		configMgr.SetConfig(testConfig)

		err := configMgr.Save(tempConfigFile)
		if err != nil {
			t.Fatalf("Failed to save config: %v", err)
		}

		// Verify file is created
		if _, err := os.Stat(tempConfigFile); os.IsNotExist(err) {
			t.Fatalf("Config file was not created: %s", tempConfigFile)
		}

		// Create new config manager and load config
		newConfigMgr := NewStaticConfigManager()
		err = newConfigMgr.Load(tempConfigFile)
		if err != nil {
			t.Fatalf("Failed to load config: %v", err)
		}

		// Verify loaded config
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

		// Clean up temporary file
		os.Remove(tempConfigFile)
	})

	t.Run("TestAddRoute", func(t *testing.T) {
		configMgr := NewStaticConfigManager()

		initialRoutes := configMgr.GetRoutes()
		if len(initialRoutes) != 0 {
			t.Errorf("Expected empty routes initially, got %d", len(initialRoutes))
		}

		// 添加路由
		newRoute := common.Route{
			ID:  "new-test-route",
			URI: "http://new-backend:8080",
			Predicates: []common.Predicate{
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

		// Add several routes
		route1 := common.Route{
			ID:  "route-to-delete",
			URI: "http://backend1:8080",
			Predicates: []common.Predicate{
				{
					Name: "Path",
					Args: map[string]string{"pattern": "/api/delete"},
				},
			},
			Order: 1,
		}

		route2 := common.Route{
			ID:  "route-to-keep",
			URI: "http://backend2:8080",
			Predicates: []common.Predicate{
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

		// Delete route
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

		// Try to delete non-existent route
		err = configMgr.DeleteRoute("non-existent-route")
		if err == nil {
			t.Error("Expected error when deleting non-existent route")
		}
	})

	t.Run("TestUpdateRoute", func(t *testing.T) {
		configMgr := NewStaticConfigManager()

		// 添加路由
		initialRoute := common.Route{
			ID:  "updatable-route",
			URI: "http://old-backend:8080",
			Predicates: []common.Predicate{
				{
					Name: "Path",
					Args: map[string]string{"pattern": "/api/update"},
				},
			},
			Order: 1,
		}

		configMgr.AddRoute(initialRoute)

		// Update route
		updatedRoute := common.Route{
			ID:  "updatable-route",
			URI: "http://new-backend:8080",
			Predicates: []common.Predicate{
				{
					Name: "Path",
					Args: map[string]string{"pattern": "/api/updated"},
				},
			},
			Filters: []common.Filter{
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

		// Verify updated route
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

		// Try to update non-existent route
		nonExistentRoute := common.Route{
			ID:  "non-existent-route",
			URI: "http://dummy:8080",
		}

		err = configMgr.UpdateRoute(nonExistentRoute)
		if err == nil {
			t.Error("Expected error when updating non-existent route")
		}
	})
}
