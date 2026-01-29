package route

import (
	"regexp"
	"sort"
	"strings"

	"go-gateway/pkg/common"
)

// Router manages routing
type Router struct {
	routes []*common.Route
}

// NewRouter creates a new router instance
func NewRouter() *Router {
	return &Router{
		routes: make([]*common.Route, 0),
	}
}

// AddRoute adds a route
func (r *Router) AddRoute(route *common.Route) {
	r.routes = append(r.routes, route)
	// 按照优先级排序
	sort.Slice(r.routes, func(i, j int) bool {
		return r.routes[i].Order < r.routes[j].Order
	})
}

// Match matches a route by path
func (r *Router) Match(path string) *common.Route {
	for _, route := range r.routes {
		if matchRoute(route, path) {
			return route
		}
	}
	return nil
}

// matchRoute checks if a route matches the given path
func matchRoute(route *common.Route, path string) bool {
	for _, predicate := range route.Predicates {
		if predicate.Name == "Path" {
			pattern, ok := predicate.Args.(map[string]string)["pattern"]
			if !ok {
				continue
			}

			if pathMatch(pattern, path) {
				return true
			}
		}
	}
	return false
}

// pathMatch checks if the path matches the pattern
func pathMatch(pattern string, path string) bool {
	// Handle /** wildcard (match any length sub-path)
	if strings.HasSuffix(pattern, "/**") {
		basePath := strings.TrimSuffix(pattern, "/**")
		return strings.HasPrefix(path, basePath)
	}

	// Handle * wildcard (match single level path)
	if strings.Contains(pattern, "*") {
		// 简单的通配符处理：将*替换为.*并使用正则匹配
		escapedPattern := regexp.QuoteMeta(pattern)
		// 将转义的*替换回.*
		regexPattern := strings.Replace(escapedPattern, "\\*", ".*", -1)

		matched, err := regexp.MatchString("^"+regexPattern+"$", path)
		return err == nil && matched
	}

	// Exact match
	return pattern == path
}
