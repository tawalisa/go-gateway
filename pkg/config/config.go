package config

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"sync"
)

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

// Config 定义配置结构
type Config struct {
	Routes        []Route        `json:"routes"`
	GlobalFilters []GlobalFilter `json:"global_filters"`
	Port          int            `json:"port"`
}

// GlobalFilter 定义全局过滤器
type GlobalFilter struct {
	Name string      `json:"name"`
	Args interface{} `json:"args"`
}

// ConfigManager 配置管理器接口
type ConfigManager interface {
	Load(configPath string) error
	Save(configPath string) error
	GetRoutes() []Route
	AddRoute(route Route)
	UpdateRoute(route Route) error
	DeleteRoute(id string) error
	GetConfig() Config
	SetConfig(config Config)
}

// StaticConfigManager 静态配置管理器
type StaticConfigManager struct {
	mutex  sync.RWMutex
	config Config
}

// NewStaticConfigManager 创建新的静态配置管理器
func NewStaticConfigManager() *StaticConfigManager {
	return &StaticConfigManager{
		config: Config{
			Routes:        make([]Route, 0),
			GlobalFilters: make([]GlobalFilter, 0),
			Port:          8080, // 默认端口
		},
	}
}

// Load 从文件加载配置
func (scm *StaticConfigManager) Load(configPath string) error {
	scm.mutex.Lock()
	defer scm.mutex.Unlock()

	data, err := ioutil.ReadFile(configPath)
	if err != nil {
		return err
	}

	err = json.Unmarshal(data, &scm.config)
	if err != nil {
		return err
	}

	return nil
}

// Save 将配置保存到文件
func (scm *StaticConfigManager) Save(configPath string) error {
	scm.mutex.RLock()
	defer scm.mutex.RUnlock()

	data, err := json.MarshalIndent(scm.config, "", "  ")
	if err != nil {
		return err
	}

	return ioutil.WriteFile(configPath, data, 0644)
}

// GetRoutes 获取所有路由
func (scm *StaticConfigManager) GetRoutes() []Route {
	scm.mutex.RLock()
	defer scm.mutex.RUnlock()

	// 返回副本以避免外部修改
	routes := make([]Route, len(scm.config.Routes))
	copy(routes, scm.config.Routes)
	return routes
}

// AddRoute 添加路由
func (scm *StaticConfigManager) AddRoute(route Route) {
	scm.mutex.Lock()
	defer scm.mutex.Unlock()

	// 检查是否已存在相同ID的路由
	for i, r := range scm.config.Routes {
		if r.ID == route.ID {
			// 如果存在，则替换
			scm.config.Routes[i] = route
			return
		}
	}

	// 添加新路由
	scm.config.Routes = append(scm.config.Routes, route)
}

// UpdateRoute 更新路由
func (scm *StaticConfigManager) UpdateRoute(route Route) error {
	scm.mutex.Lock()
	defer scm.mutex.Unlock()

	for i, r := range scm.config.Routes {
		if r.ID == route.ID {
			scm.config.Routes[i] = route
			return nil
		}
	}

	return os.ErrNotExist
}

// DeleteRoute 删除路由
func (scm *StaticConfigManager) DeleteRoute(id string) error {
	scm.mutex.Lock()
	defer scm.mutex.Unlock()

	for i, route := range scm.config.Routes {
		if route.ID == id {
			// 从切片中移除元素
			scm.config.Routes = append(scm.config.Routes[:i], scm.config.Routes[i+1:]...)
			return nil
		}
	}

	return os.ErrNotExist
}

// GetConfig 获取完整配置
func (scm *StaticConfigManager) GetConfig() Config {
	scm.mutex.RLock()
	defer scm.mutex.RUnlock()

	// 返回副本
	config := scm.config

	// 复制路由切片
	config.Routes = make([]Route, len(scm.config.Routes))
	copy(config.Routes, scm.config.Routes)

	// 复制全局过滤器切片
	config.GlobalFilters = make([]GlobalFilter, len(scm.config.GlobalFilters))
	copy(config.GlobalFilters, scm.config.GlobalFilters)

	return config
}

// SetConfig 设置完整配置
func (scm *StaticConfigManager) SetConfig(config Config) {
	scm.mutex.Lock()
	defer scm.mutex.Unlock()

	scm.config = config
}
