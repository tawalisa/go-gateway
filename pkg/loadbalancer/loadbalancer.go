package loadbalancer

import (
	"math/rand"
	"sync"
	"time"
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

// RoundRobinBalancer 轮询负载均衡器
type RoundRobinBalancer struct {
	mutex        sync.RWMutex
	servers      []Server
	currentIndex int
}

// NewRoundRobinBalancer 创建新的轮询负载均衡器
func NewRoundRobinBalancer() *RoundRobinBalancer {
	return &RoundRobinBalancer{
		servers:      make([]Server, 0),
		currentIndex: 0,
	}
}

// AddServer 添加服务器
func (rr *RoundRobinBalancer) AddServer(server Server) {
	rr.mutex.Lock()
	defer rr.mutex.Unlock()
	rr.servers = append(rr.servers, server)
}

// RemoveServer 移除服务器
func (rr *RoundRobinBalancer) RemoveServer(url string) {
	rr.mutex.Lock()
	defer rr.mutex.Unlock()

	for i, server := range rr.servers {
		if server.URL == url {
			rr.servers = append(rr.servers[:i], rr.servers[i+1:]...)
			if rr.currentIndex >= len(rr.servers) && len(rr.servers) > 0 {
				rr.currentIndex = len(rr.servers) - 1
			}
			break
		}
	}
}

// UpdateServer 更新服务器
func (rr *RoundRobinBalancer) UpdateServer(server Server) {
	rr.mutex.Lock()
	defer rr.mutex.Unlock()

	for i, s := range rr.servers {
		if s.URL == server.URL {
			rr.servers[i] = server
			break
		}
	}
}

// GetServers 获取所有服务器
func (rr *RoundRobinBalancer) GetServers() []Server {
	rr.mutex.RLock()
	defer rr.mutex.RUnlock()

	// 返回副本以避免外部修改
	result := make([]Server, len(rr.servers))
	copy(result, rr.servers)
	return result
}

// ChooseServer 选择服务器
func (rr *RoundRobinBalancer) ChooseServer(servers []Server) *Server {
	rr.mutex.Lock()
	defer rr.mutex.Unlock()

	if len(servers) == 0 {
		return nil
	}

	// 过滤掉重复的服务器以避免重复计算
	uniqueServers := make([]Server, 0)
	seen := make(map[string]bool)

	for _, server := range servers {
		if !seen[server.URL] {
			seen[server.URL] = true
			uniqueServers = append(uniqueServers, server)
		}
	}

	if len(uniqueServers) == 0 {
		return nil
	}

	// 简单轮询选择
	server := &uniqueServers[rr.currentIndex%len(uniqueServers)]
	rr.currentIndex++
	return server
}

// RandomBalancer 随机负载均衡器
type RandomBalancer struct {
	mutex   sync.RWMutex
	servers []Server
	rand    *rand.Rand
}

// NewRandomBalancer 创建新的随机负载均衡器
func NewRandomBalancer() *RandomBalancer {
	return &RandomBalancer{
		servers: make([]Server, 0),
		rand:    rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// AddServer 添加服务器
func (rb *RandomBalancer) AddServer(server Server) {
	rb.mutex.Lock()
	defer rb.mutex.Unlock()
	rb.servers = append(rb.servers, server)
}

// RemoveServer 移除服务器
func (rb *RandomBalancer) RemoveServer(url string) {
	rb.mutex.Lock()
	defer rb.mutex.Unlock()

	for i, server := range rb.servers {
		if server.URL == url {
			rb.servers = append(rb.servers[:i], rb.servers[i+1:]...)
			break
		}
	}
}

// UpdateServer 更新服务器
func (rb *RandomBalancer) UpdateServer(server Server) {
	rb.mutex.Lock()
	defer rb.mutex.Unlock()

	for i, s := range rb.servers {
		if s.URL == server.URL {
			rb.servers[i] = server
			break
		}
	}
}

// GetServers 获取所有服务器
func (rb *RandomBalancer) GetServers() []Server {
	rb.mutex.RLock()
	defer rb.mutex.RUnlock()

	result := make([]Server, len(rb.servers))
	copy(result, rb.servers)
	return result
}

// ChooseServer 随机选择服务器
func (rb *RandomBalancer) ChooseServer(servers []Server) *Server {
	rb.mutex.RLock()
	defer rb.mutex.RUnlock()

	if len(servers) == 0 {
		return nil
	}

	// 过滤唯一服务器
	uniqueServers := make([]Server, 0)
	seen := make(map[string]bool)

	for _, server := range servers {
		if !seen[server.URL] {
			seen[server.URL] = true
			uniqueServers = append(uniqueServers, server)
		}
	}

	if len(uniqueServers) == 0 {
		return nil
	}

	// 随机选择
	index := rb.rand.Intn(len(uniqueServers))
	server := &uniqueServers[index]
	return server
}

// WeightedRoundRobinBalancer 加权轮询负载均衡器
type WeightedRoundRobinBalancer struct {
	mutex        sync.RWMutex
	servers      []Server
	currentIndex int
	totalWeight  int
}

// NewWeightedRoundRobinBalancer 创建新的加权轮询负载均衡器
func NewWeightedRoundRobinBalancer() *WeightedRoundRobinBalancer {
	return &WeightedRoundRobinBalancer{
		servers:      make([]Server, 0),
		currentIndex: 0,
		totalWeight:  0,
	}
}

// AddServer 添加服务器
func (wrr *WeightedRoundRobinBalancer) AddServer(server Server) {
	wrr.mutex.Lock()
	defer wrr.mutex.Unlock()
	wrr.servers = append(wrr.servers, server)
	wrr.totalWeight += server.Weight
}

// RemoveServer 移除服务器
func (wrr *WeightedRoundRobinBalancer) RemoveServer(url string) {
	wrr.mutex.Lock()
	defer wrr.mutex.Unlock()

	for i, server := range wrr.servers {
		if server.URL == url {
			wrr.totalWeight -= server.Weight
			wrr.servers = append(wrr.servers[:i], wrr.servers[i+1:]...)
			if wrr.currentIndex >= len(wrr.servers) && len(wrr.servers) > 0 {
				wrr.currentIndex = len(wrr.servers) - 1
			}
			break
		}
	}
}

// UpdateServer 更新服务器
func (wrr *WeightedRoundRobinBalancer) UpdateServer(server Server) {
	wrr.mutex.Lock()
	defer wrr.mutex.Unlock()

	for i, s := range wrr.servers {
		if s.URL == server.URL {
			wrr.totalWeight -= s.Weight
			wrr.servers[i] = server
			wrr.totalWeight += server.Weight
			break
		}
	}
}

// GetServers 获取所有服务器
func (wrr *WeightedRoundRobinBalancer) GetServers() []Server {
	wrr.mutex.RLock()
	defer wrr.mutex.RUnlock()

	result := make([]Server, len(wrr.servers))
	copy(result, wrr.servers)
	return result
}

// ChooseServer 加权轮询选择服务器
func (wrr *WeightedRoundRobinBalancer) ChooseServer(servers []Server) *Server {
	wrr.mutex.Lock()
	defer wrr.mutex.Unlock()

	if len(servers) == 0 {
		return nil
	}

	// 过滤唯一服务器
	uniqueServers := make([]Server, 0)
	seen := make(map[string]bool)

	for _, server := range servers {
		if !seen[server.URL] {
			seen[server.URL] = true
			uniqueServers = append(uniqueServers, server)
		}
	}

	if len(uniqueServers) == 0 {
		return nil
	}

	// 简单的加权轮询实现
	// 这里我们使用一个简化的算法：按权重分配选择机会
	totalWeight := 0
	for _, server := range uniqueServers {
		totalWeight += server.Weight
	}

	if totalWeight == 0 {
		// 如果所有权重都是0，则退回到普通轮询
		server := &uniqueServers[wrr.currentIndex%len(uniqueServers)]
		wrr.currentIndex++
		return server
	}

	// 按权重选择
	currentWeight := 0
	for _, server := range uniqueServers {
		currentWeight += server.Weight
		if wrr.currentIndex%totalWeight < currentWeight {
			result := server
			wrr.currentIndex = (wrr.currentIndex + 1) % totalWeight
			return &result
		}
	}

	// 默认返回第一个
	server := &uniqueServers[0]
	wrr.currentIndex = (wrr.currentIndex + 1) % totalWeight
	return server
}
