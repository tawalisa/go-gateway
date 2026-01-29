package middleware

import (
	"net/http"
)

// Middleware 定义中间件接口
type Middleware interface {
	Name() string
	PreHandle(ctx *GatewayContext) bool
	PostHandle(ctx *GatewayContext) error
	HandleError(ctx *GatewayContext, err error)
}

// MiddlewareChain 表示中间件链
type MiddlewareChain struct {
	index    int
	handlers []Middleware
}

// NewMiddlewareChain 创建新的中间件链
func NewMiddlewareChain(handlers []Middleware) *MiddlewareChain {
	return &MiddlewareChain{
		handlers: handlers,
		index:    0,
	}
}

// Execute 执行中间件链
func (mc *MiddlewareChain) Execute(ctx *GatewayContext) {
	for mc.index < len(mc.handlers) {
		handler := mc.handlers[mc.index]
		mc.index++

		if !handler.PreHandle(ctx) {
			// 如果PreHandle返回false，停止执行后续中间件
			break
		}
	}

	// 执行后置处理（逆序）
	for i := len(mc.handlers) - 1; i >= mc.index-1; i-- {
		handler := mc.handlers[i]
		if err := handler.PostHandle(ctx); err != nil {
			handler.HandleError(ctx, err)
		}
	}
}

// ExecuteNext 执行下一个中间件
func (mc *MiddlewareChain) ExecuteNext(ctx *GatewayContext) bool {
	if mc.index >= len(mc.handlers) {
		return false
	}

	handler := mc.handlers[mc.index]
	mc.index++

	result := handler.PreHandle(ctx)
	if result {
		// 继续执行下一个中间件
		if mc.ExecuteNext(ctx) {
			// 后置处理
			mc.index--
			if err := handler.PostHandle(ctx); err != nil {
				handler.HandleError(ctx, err)
			}
		} else {
			mc.index--
			if err := handler.PostHandle(ctx); err != nil {
				handler.HandleError(ctx, err)
			}
		}
	} else {
		mc.index--
	}

	return result
}

// GatewayContext 定义网关请求上下文
type GatewayContext struct {
	Request     *http.Request
	Response    http.ResponseWriter
	Route       *Route
	Attributes  map[string]interface{}
	StartTime   int64
	OriginalURL string
}

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
