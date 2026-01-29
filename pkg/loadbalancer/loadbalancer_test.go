package loadbalancer

import (
	"fmt"
	"testing"
)

// Server 定义后端服务器结构
type Server struct {
	URL    string
	Weight int
}

// LoadBalancer 定义负载均衡器接口
type LoadBalancer interface {
	ChooseServer(servers []Server) *Server
	AddServer(server Server)
	RemoveServer(url string)
	UpdateServer(server Server)
	GetServers() []Server
}

// TestRoundRobinLoadBalancer 测试轮询负载均衡器
func TestRoundRobinLoadBalancer(t *testing.T) {
	t.Run("TestRoundRobinSelection", func(t *testing.T) {
		servers := []Server{
			{URL: "http://server1:8080", Weight: 1},
			{URL: "http://server2:8080", Weight: 1},
			{URL: "http://server3:8080", Weight: 1},
		}

		lb := NewRoundRobinBalancer()

		// 添加服务器
		for _, server := range servers {
			lb.AddServer(server)
		}

		// 测试轮询选择
		selectedServers := make([]string, 0, 6)
		for i := 0; i < 6; i++ {
			server := lb.ChooseServer(lb.GetServers())
			if server != nil {
				selectedServers = append(selectedServers, server.URL)
			}
		}

		// 验证轮询效果：应该是循环选择三个服务器
		expectedSequence := []string{
			"http://server1:8080",
			"http://server2:8080",
			"http://server3:8080",
			"http://server1:8080",
			"http://server2:8080",
			"http://server3:8080",
		}

		if len(selectedServers) != len(expectedSequence) {
			t.Errorf("Expected %d selections, got %d", len(expectedSequence), len(selectedServers))
		}

		for i, expected := range expectedSequence {
			if i >= len(selectedServers) {
				t.Errorf("Missing selection at index %d", i)
				continue
			}
			if selectedServers[i] != expected {
				t.Errorf("At index %d, expected '%s', got '%s'", i, expected, selectedServers[i])
			}
		}
	})

	t.Run("TestEmptyServerList", func(t *testing.T) {
		lb := NewRoundRobinBalancer()
		server := lb.ChooseServer([]Server{})

		if server != nil {
			t.Errorf("Expected nil for empty server list, got %v", server)
		}
	})

	t.Run("TestSingleServer", func(t *testing.T) {
		servers := []Server{
			{URL: "http://only-server:8080", Weight: 1},
		}

		lb := NewRoundRobinBalancer()
		for _, server := range servers {
			lb.AddServer(server)
		}

		// 多次选择应该总是返回同一个服务器
		for i := 0; i < 5; i++ {
			selected := lb.ChooseServer(lb.GetServers())
			if selected == nil || selected.URL != "http://only-server:8080" {
				t.Errorf("Expected 'http://only-server:8080', got %v", selected)
			}
		}
	})
}

// TestRandomLoadBalancer 测试随机负载均衡器
func TestRandomLoadBalancer(t *testing.T) {
	t.Run("TestRandomSelection", func(t *testing.T) {
		servers := []Server{
			{URL: "http://server1:8080", Weight: 1},
			{URL: "http://server2:8080", Weight: 1},
			{URL: "http://server3:8080", Weight: 1},
		}

		lb := NewRandomBalancer()

		// 添加服务器
		for _, server := range servers {
			lb.AddServer(server)
		}

		// 测试多次随机选择，确保至少选择了不同的服务器
		selectedCount := make(map[string]int)
		selectionCount := 100

		for i := 0; i < selectionCount; i++ {
			server := lb.ChooseServer(lb.GetServers())
			if server != nil {
				selectedCount[server.URL]++
			}
		}

		// 验证所有服务器都被选中过（概率上应该如此）
		if len(selectedCount) != 3 {
			t.Errorf("Expected 3 different servers to be selected, got %d", len(selectedCount))
		}

		for _, server := range servers {
			if selectedCount[server.URL] == 0 {
				t.Errorf("Server %s was never selected", server.URL)
			}
		}
	})

	t.Run("TestRandomWithEmptyList", func(t *testing.T) {
		lb := NewRandomBalancer()
		server := lb.ChooseServer([]Server{})

		if server != nil {
			t.Errorf("Expected nil for empty server list, got %v", server)
		}
	})
}

// TestWeightedRoundRobinLoadBalancer 测试加权轮询负载均衡器
func TestWeightedRoundRobinLoadBalancer(t *testing.T) {
	t.Run("TestWeightedRoundRobin", func(t *testing.T) {
		servers := []Server{
			{URL: "http://high-weight:8080", Weight: 3},
			{URL: "http://low-weight:8080", Weight: 1},
		}

		lb := NewWeightedRoundRobinBalancer()

		// 添加服务器
		for _, server := range servers {
			lb.AddServer(server)
		}

		// 执行多次选择以验证权重
		selectedCount := make(map[string]int)
		selectionCount := 8 // 应该是4次高权重，4次低权重的倍数

		for i := 0; i < selectionCount; i++ {
			server := lb.ChooseServer(lb.GetServers())
			if server != nil {
				selectedCount[server.URL]++
			}
		}

		// 在加权轮询中，高权重服务器应该被选中更多次
		highWeightSelected := selectedCount["http://high-weight:8080"]
		lowWeightSelected := selectedCount["http://low-weight:8080"]

		// 理想情况下，比例应该是3:1，但我们允许一定的误差范围
		if highWeightSelected <= lowWeightSelected {
			t.Errorf("High weight server should be selected more often. High: %d, Low: %d",
				highWeightSelected, lowWeightSelected)
		}
	})
}

// MockServerHealthChecker 模拟服务器健康检查器
type MockServerHealthChecker struct {
	healthyServers map[string]bool
}

func NewMockServerHealthChecker() *MockServerHealthChecker {
	return &MockServerHealthChecker{
		healthyServers: make(map[string]bool),
	}
}

func (mhc *MockServerHealthChecker) SetHealthy(url string, healthy bool) {
	mhc.healthyServers[url] = healthy
}

func (mhc *MockServerHealthChecker) IsHealthy(url string) bool {
	if healthy, exists := mhc.healthyServers[url]; exists {
		return healthy
	}
	// 默认认为是健康的
	return true
}

// TestHealthAwareLoadBalancer 测试带健康检查的负载均衡器
func TestHealthAwareLoadBalancer(t *testing.T) {
	t.Run("TestSkipUnhealthyServer", func(t *testing.T) {
		servers := []Server{
			{URL: "http://healthy-server:8080", Weight: 1},
			{URL: "http://unhealthy-server:8080", Weight: 1},
		}

		lb := NewRoundRobinBalancer()

		// 添加服务器
		for _, server := range servers {
			lb.AddServer(server)
		}

		// 模拟健康检查
		mockChecker := NewMockServerHealthChecker()
		mockChecker.SetHealthy("http://healthy-server:8080", true)
		mockChecker.SetHealthy("http://unhealthy-server:8080", false)

		// 测试选择逻辑 - 应该只选择健康服务器
		for i := 0; i < 10; i++ {
			selected := lb.ChooseServer(lb.GetServers())
			if selected == nil {
				t.Errorf("Got nil server selection")
				break
			}
			if selected.URL != "http://healthy-server:8080" {
				t.Errorf("Expected healthy server, got %s", selected.URL)
			}
		}
	})
}
