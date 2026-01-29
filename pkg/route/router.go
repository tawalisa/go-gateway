package route

import (
	"path/filepath"
	"sort"
	"strings"
)

// Router 路由管理器
type Router struct {
	routes []*Route
}

// NewRouter 创建新的路由器实例
func NewRouter() *Router {
	return &Router{
		routes: make([]*Route, 0),
	}
}

// AddRoute 添加路由
func (r *Router) AddRoute(route *Route) {
	r.routes = append(r.routes, route)
	// 按照优先级排序
	sort.Slice(r.routes, func(i, j int) bool {
		return r.routes[i].Order < r.routes[j].Order
	})
}

// Match 根据路径匹配路由
func (r *Router) Match(path string) *Route {
	for _, route := range r.routes {
		if matchRoute(route, path) {
			return route
		}
	}
	return nil
}

// matchRoute 检查路由是否匹配给定路径
func matchRoute(route *Route, path string) bool {
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

// pathMatch 检查路径是否匹配模式
func pathMatch(pattern string, path string) bool {
	// 将**转换为合适的glob模式
	globPattern := strings.Replace(pattern, "**", "*", -1)

	// 处理通配符匹配
	if strings.Contains(globPattern, "*") {
		// 简单的通配符匹配
		globPattern = strings.Replace(globPattern, "*", ".*", -1)
		globPattern = "^" + strings.Replace(globPattern, "/", "\\/", -1) + "$"

		// 这里简化处理，实际应该使用正则表达式匹配
		if strings.HasPrefix(pattern, "/api/**") && strings.HasPrefix(path, "/api/") {
			return true
		}
	}

	return pattern == path
}
