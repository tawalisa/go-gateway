package middleware

import (
	"net/http"

	"go-gateway/pkg/common"
)

// Middleware defines the middleware interface
type Middleware interface {
	Name() string
	PreHandle(ctx *GatewayContext) bool
	PostHandle(ctx *GatewayContext) error
	HandleError(ctx *GatewayContext, err error)
}

// MiddlewareChain represents a chain of middlewares
type MiddlewareChain struct {
	index    int
	handlers []Middleware
}

// NewMiddlewareChain creates a new middleware chain
func NewMiddlewareChain(handlers []Middleware) *MiddlewareChain {
	return &MiddlewareChain{
		handlers: handlers,
		index:    0,
	}
}

// Execute executes the middleware chain
func (mc *MiddlewareChain) Execute(ctx *GatewayContext) {
	for mc.index < len(mc.handlers) {
		handler := mc.handlers[mc.index]
		mc.index++

		if !handler.PreHandle(ctx) {
			// If PreHandle returns false, stop executing subsequent middlewares
			break
		}
	}

	// Execute post-processing (in reverse order)
	// Ensure not accessing negative index
	startIndex := mc.index - 1
	if startIndex < 0 {
		startIndex = 0
	}
	for i := len(mc.handlers) - 1; i >= startIndex; i-- {
		handler := mc.handlers[i]
		if err := handler.PostHandle(ctx); err != nil {
			handler.HandleError(ctx, err)
		}
	}
}

// ExecuteNext executes the next middleware
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

// GatewayContext defines the gateway request context
type GatewayContext struct {
	Request     *http.Request
	Response    http.ResponseWriter
	Route       *common.Route
	Attributes  map[string]interface{}
	StartTime   int64
	OriginalURL string
	Index       int // Current executing middleware index
	Handlers    []Middleware
}
