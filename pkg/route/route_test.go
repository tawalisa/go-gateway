package route

import (
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

// TestRouteMatch 测试路由匹配功能
func TestRouteMatch(t *testing.T) {
	t.Run("TestExactPathMatch", func(t *testing.T) {
		routes := []*Route{
			{
				ID:  "test-route",
				URI: "http://backend-service",
				Predicates: []Predicate{
					{
						Name: "Path",
						Args: map[string]string{"pattern": "/api/test"},
					},
				},
			},
		}

		router := NewRouter()
		for _, route := range routes {
			router.AddRoute(route)
		}

		matchedRoute := router.Match("/api/test")
		if matchedRoute == nil {
			t.Errorf("Expected route to match /api/test, got nil")
		} else if matchedRoute.ID != "test-route" {
			t.Errorf("Expected route ID 'test-route', got %s", matchedRoute.ID)
		}
	})

	t.Run("TestWildcardPathMatch", func(t *testing.T) {
		routes := []*Route{
			{
				ID:  "wildcard-route",
				URI: "http://backend-service",
				Predicates: []Predicate{
					{
						Name: "Path",
						Args: map[string]string{"pattern": "/api/**"},
					},
				},
			},
		}

		router := NewRouter()
		for _, route := range routes {
			router.AddRoute(route)
		}

		matchedRoute := router.Match("/api/users/123")
		if matchedRoute == nil {
			t.Errorf("Expected route to match /api/users/123, got nil")
		} else if matchedRoute.ID != "wildcard-route" {
			t.Errorf("Expected route ID 'wildcard-route', got %s", matchedRoute.ID)
		}
	})

	t.Run("TestNoMatch", func(t *testing.T) {
		routes := []*Route{
			{
				ID:  "test-route",
				URI: "http://backend-service",
				Predicates: []Predicate{
					{
						Name: "Path",
						Args: map[string]string{"pattern": "/api/test"},
					},
				},
			},
		}

		router := NewRouter()
		for _, route := range routes {
			router.AddRoute(route)
		}

		matchedRoute := router.Match("/nonexistent/path")
		if matchedRoute != nil {
			t.Errorf("Expected no route to match /nonexistent/path, got %s", matchedRoute.ID)
		}
	})
}

// TestRoutePriority 测试路由优先级
func TestRoutePriority(t *testing.T) {
	t.Run("TestHigherPriorityRouteMatchesFirst", func(t *testing.T) {
		routes := []*Route{
			{
				ID:    "low-priority",
				URI:   "http://backend1",
				Order: 1,
				Predicates: []Predicate{
					{
						Name: "Path",
						Args: map[string]string{"pattern": "/api/**"},
					},
				},
			},
			{
				ID:    "high-priority",
				URI:   "http://backend2",
				Order: 0, // Higher priority (lower number)
				Predicates: []Predicate{
					{
						Name: "Path",
						Args: map[string]string{"pattern": "/api/specific"},
					},
				},
			},
		}

		router := NewRouter()
		for _, route := range routes {
			router.AddRoute(route)
		}

		matchedRoute := router.Match("/api/specific")
		if matchedRoute == nil {
			t.Errorf("Expected route to match /api/specific, got nil")
		} else if matchedRoute.ID != "high-priority" {
			t.Errorf("Expected high priority route 'high-priority', got %s", matchedRoute.ID)
		}
	})
}
