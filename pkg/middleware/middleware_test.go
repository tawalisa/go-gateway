package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// GatewayContext 定义网关请求上下文
type GatewayContext struct {
	Request     *http.Request
	Response    http.ResponseWriter
	Route       *Route
	Attributes  map[string]interface{}
	StartTime   int64
	OriginalURL string
	Index       int // 当前执行的中间件索引
	Handlers    []Middleware
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

// Middleware 定义中间件接口
type Middleware interface {
	Name() string
	PreHandle(ctx *GatewayContext) bool
	PostHandle(ctx *GatewayContext) error
	HandleError(ctx *GatewayContext, err error)
}

// TestMiddlewareChain 测试中间件链功能
func TestMiddlewareChain(t *testing.T) {
	t.Run("TestBasicMiddlewareChain", func(t *testing.T) {
		var executedOrder []string

		// 创建测试中间件
		middleware1 := &testMiddleware{
			name: "middleware1",
			onPreHandle: func(ctx *GatewayContext) bool {
				executedOrder = append(executedOrder, "pre-middleware1")
				return true
			},
			onPostHandle: func(ctx *GatewayContext) error {
				executedOrder = append(executedOrder, "post-middleware1")
				return nil
			},
		}

		middleware2 := &testMiddleware{
			name: "middleware2",
			onPreHandle: func(ctx *GatewayContext) bool {
				executedOrder = append(executedOrder, "pre-middleware2")
				return true
			},
			onPostHandle: func(ctx *GatewayContext) error {
				executedOrder = append(executedOrder, "post-middleware2")
				return nil
			},
		}

		// 创建上下文并添加中间件
		req := httptest.NewRequest("GET", "http://localhost/test", nil)
		resp := httptest.NewRecorder()

		ctx := &GatewayContext{
			Request:  req,
			Response: resp,
			Handlers: []Middleware{middleware1, middleware2},
		}

		// 执行中间件链
		executeMiddlewareChain(ctx)

		// 验证执行顺序
		expectedOrder := []string{
			"pre-middleware1",
			"pre-middleware2",
			"post-middleware2",
			"post-middleware1",
		}

		if len(executedOrder) != len(expectedOrder) {
			t.Errorf("Expected %d executions, got %d", len(expectedOrder), len(executedOrder))
		}

		for i, expected := range expectedOrder {
			if i >= len(executedOrder) || executedOrder[i] != expected {
				t.Errorf("At index %d, expected '%s', got '%s'", i, expected, executedOrder[i])
			}
		}
	})

	t.Run("TestMiddlewareEarlyTermination", func(t *testing.T) {
		var executedOrder []string

		// 创建测试中间件，第二个中间件返回false终止执行
		middleware1 := &testMiddleware{
			name: "middleware1",
			onPreHandle: func(ctx *GatewayContext) bool {
				executedOrder = append(executedOrder, "pre-middleware1")
				return true
			},
			onPostHandle: func(ctx *GatewayContext) error {
				executedOrder = append(executedOrder, "post-middleware1")
				return nil
			},
		}

		middleware2 := &testMiddleware{
			name: "middleware2",
			onPreHandle: func(ctx *GatewayContext) bool {
				executedOrder = append(executedOrder, "pre-middleware2-stop")
				return false // 终止执行
			},
			onPostHandle: func(ctx *GatewayContext) error {
				executedOrder = append(executedOrder, "post-middleware2") // 不应该执行到这里
				return nil
			},
		}

		middleware3 := &testMiddleware{
			name: "middleware3",
			onPreHandle: func(ctx *GatewayContext) bool {
				executedOrder = append(executedOrder, "pre-middleware3") // 不应该执行到这里
				return true
			},
			onPostHandle: func(ctx *GatewayContext) error {
				executedOrder = append(executedOrder, "post-middleware3")
				return nil
			},
		}

		// 创建上下文并添加中间件
		req := httptest.NewRequest("GET", "http://localhost/test", nil)
		resp := httptest.NewRecorder()

		ctx := &GatewayContext{
			Request:  req,
			Response: resp,
			Handlers: []Middleware{middleware1, middleware2, middleware3},
		}

		// 执行中间件链
		executeMiddlewareChain(ctx)

		// 验证执行顺序 - 第三个中间件不应该被执行
		expectedOrder := []string{
			"pre-middleware1",
			"pre-middleware2-stop",
		}

		if len(executedOrder) != len(expectedOrder) {
			t.Errorf("Expected %d executions, got %d", len(expectedOrder), len(executedOrder))
		}

		for i, expected := range expectedOrder {
			if i >= len(executedOrder) || executedOrder[i] != expected {
				t.Errorf("At index %d, expected '%s', got '%s'", i, expected, executedOrder[i])
			}
		}
	})
}

// testMiddleware 实现中间件接口的测试中间件
type testMiddleware struct {
	name         string
	onPreHandle  func(*GatewayContext) bool
	onPostHandle func(*GatewayContext) error
}

func (tm *testMiddleware) Name() string {
	return tm.name
}

func (tm *testMiddleware) PreHandle(ctx *GatewayContext) bool {
	if tm.onPreHandle != nil {
		return tm.onPreHandle(ctx)
	}
	return true
}

func (tm *testMiddleware) PostHandle(ctx *GatewayContext) error {
	if tm.onPostHandle != nil {
		return tm.onPostHandle(ctx)
	}
	return nil
}

func (tm *testMiddleware) HandleError(ctx *GatewayContext, err error) {
	// 测试用的空实现
}

// executeMiddlewareChain 执行中间件链的辅助函数
func executeMiddlewareChain(ctx *GatewayContext) {
	// 执行前置处理器
	for i, handler := range ctx.Handlers {
		ctx.Index = i
		if !handler.PreHandle(ctx) {
			break
		}
	}

	// 执行后置处理器（逆序）
	for i := len(ctx.Handlers) - 1; i >= 0; i-- {
		handler := ctx.Handlers[i]
		handler.PostHandle(ctx)
	}
}
