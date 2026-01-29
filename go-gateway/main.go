package main

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"sync"

	"go-gateway/pkg/common"
	"go-gateway/pkg/route"

	"go-gateway/pkg/config"
	"go-gateway/pkg/loadbalancer"
	"go-gateway/pkg/middleware"
)

// Gateway represents gateway instance
type Gateway struct {
	configManager *config.StaticConfigManager
	router        *route.Router
	loadBalancer  loadbalancer.LoadBalancer
	middlewares   []middleware.Middleware
	mutex         sync.RWMutex
}

// NewGateway creates new gateway instance
func NewGateway() *Gateway {
	return &Gateway{
		configManager: config.NewStaticConfigManager(),
		router:        route.NewRouter(),
		loadBalancer:  loadbalancer.NewRoundRobinBalancer(),
		middlewares:   make([]middleware.Middleware, 0),
	}
}

// LoadConfig loads config from config file
func (g *Gateway) LoadConfig(configPath string) error {
	err := g.configManager.Load(configPath)
	if err != nil {
		return err
	}

	// Reload routes
	g.reloadRoutes()

	return nil
}

// reloadRoutes reloads routes
func (g *Gateway) reloadRoutes() {
	// Clear existing routes
	g.router = route.NewRouter()

	// Load routes from config
	for _, routeConfig := range g.configManager.GetRoutes() {
		// Need to convert config.Route to common.Route
		internalRoute := &common.Route{
			ID:         routeConfig.ID,
			URI:        routeConfig.URI,
			Predicates: convertPredicates(routeConfig.Predicates),
			Filters:    convertFilters(routeConfig.Filters),
			Order:      routeConfig.Order,
			Metadata:   routeConfig.Metadata,
		}
		g.router.AddRoute(internalRoute)
	}
}

// ServeHTTP implements HTTP handler interface
func (g *Gateway) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Match route
	matchedRoute := g.router.Match(r.URL.Path)
	if matchedRoute == nil {
		http.NotFound(w, r)
		return
	}

	// Create gateway context
	gatewayCtx := &middleware.GatewayContext{
		Request:     r,
		Response:    w,
		Route:       matchedRoute, // Now this is compatible with common.Route
		Attributes:  make(map[string]interface{}),
		StartTime:   0, // Should set current time in actual use
		OriginalURL: r.URL.String(),
		Handlers:    g.middlewares,
		Index:       0,
	}

	// Execute middleware chain
	chain := middleware.NewMiddlewareChain(g.middlewares)
	chain.Execute(gatewayCtx)

	// Determine target URL based on route URI
	targetURL := matchedRoute.URI
	if strings.HasPrefix(targetURL, "lb://") {
		// If it's load balancer identifier, select a backend server
		_ = strings.TrimPrefix(targetURL, "lb://") // Service name, temporarily unused
		// Simplified processing here, should get server list by service name in reality
		servers := g.loadBalancer.GetServers()
		if len(servers) > 0 {
			chosenServer := g.loadBalancer.ChooseServer(servers)
			if chosenServer != nil {
				targetURL = chosenServer.URL
			}
		}
	}

	// Parse target URL
	target, err := url.Parse(targetURL)
	if err != nil {
		http.Error(w, "Invalid target URL", http.StatusInternalServerError)
		return
	}

	// Create reverse proxy
	proxy := httputil.NewSingleHostReverseProxy(target)

	// Forward request
	proxy.ServeHTTP(w, r)
}

// Run starts gateway service
func (g *Gateway) Run(port int) error {
	addr := fmt.Sprintf(":%d", port)

	return http.ListenAndServe(addr, g)
}

// convertPredicates converts predicates
func convertPredicates(predicates []common.Predicate) []common.Predicate {
	result := make([]common.Predicate, len(predicates))
	for i, p := range predicates {
		result[i] = common.Predicate{
			Name: p.Name,
			Args: p.Args,
		}
	}
	return result
}

// convertFilters converts filters
func convertFilters(filters []common.Filter) []common.Filter {
	result := make([]common.Filter, len(filters))
	for i, f := range filters {
		result[i] = common.Filter{
			Name: f.Name,
			Args: f.Args,
		}
	}
	return result
}

func main() {
	gateway := NewGateway()

	// Initialize default config
	defaultConfig := config.Config{
		Routes: []common.Route{
			{
				ID:  "bing-redirect",
				URI: "http://localhost:18081", // Default redirect to CN Bing
				Predicates: []common.Predicate{
					{
						Name: "Path",
						Args: map[string]string{"pattern": "/**"}, // All paths
					},
				},
				Filters: []common.Filter{},
				Order:   999, // Low priority, serves as fallback route
			},
		},
		Port: 8080,
	}
	gateway.configManager.SetConfig(defaultConfig)
	gateway.reloadRoutes()

	log.Println("Starting gateway on :8080")
	if err := gateway.Run(8080); err != nil {
		log.Fatal("Gateway failed to start: ", err)
	}
}
