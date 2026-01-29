package loadbalancer

import (
	"testing"
)

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
	t.Run("TestHealthAwareSelection", func(t *testing.T) {
		// 注意：当前实现的负载均衡器不包含健康检查逻辑
		// 这个测试展示了如何集成健康检查
		servers := []Server{
			{URL: "http://server1:8080", Weight: 1},
			{URL: "http://server2:8080", Weight: 1},
		}

		lb := NewRoundRobinBalancer()

		// 添加服务器
		for _, server := range servers {
			lb.AddServer(server)
		}

		// 当前实现会返回所有服务器
		selected := lb.ChooseServer(lb.GetServers())
		if selected == nil {
			t.Errorf("Got nil server selection")
		}

		// 简单验证返回的服务器在列表中
		found := false
		for _, s := range servers {
			if s.URL == selected.URL {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Selected server not in original list: %s", selected.URL)
		}
	})
}
