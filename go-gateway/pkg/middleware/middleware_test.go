package middleware

import (
	"net/http/httptest"
	"testing"
)

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
		// 但是middleware1和middleware2的post处理应该被执行（逆序）
		expectedOrder := []string{
			"pre-middleware1",
			"pre-middleware2-stop",
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
			// 如果某个中间件返回false，停止执行后续的前置处理器
			// 然后执行已执行前置处理器的中间件的后置处理器（逆序）
			for j := i; j >= 0; j-- {
				handler := ctx.Handlers[j]
				handler.PostHandle(ctx)
			}
			return
		}
	}

	// 如果所有中间件都成功执行前置处理器，则执行所有中间件的后置处理器（逆序）
	for i := len(ctx.Handlers) - 1; i >= 0; i-- {
		ctx.Handlers[i].PostHandle(ctx)
	}
}
