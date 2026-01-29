package config

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"sync"

	"go-gateway/pkg/common"
)

// Config defines configuration structure
type Config struct {
	Routes        []common.Route `json:"routes"`
	GlobalFilters []GlobalFilter `json:"global_filters"`
	Port          int            `json:"port"`
}

// GlobalFilter defines global filter
type GlobalFilter struct {
	Name string      `json:"name"`
	Args interface{} `json:"args"`
}

// ConfigManager configuration manager interface
type ConfigManager interface {
	Load(configPath string) error
	Save(configPath string) error
	GetRoutes() []common.Route
	AddRoute(route common.Route)
	UpdateRoute(route common.Route) error
	DeleteRoute(id string) error
	GetConfig() Config
	SetConfig(config Config)
}

// StaticConfigManager static configuration manager
type StaticConfigManager struct {
	mutex  sync.RWMutex
	config Config
}

// NewStaticConfigManager creates a new static config manager
func NewStaticConfigManager() *StaticConfigManager {
	return &StaticConfigManager{
		config: Config{
			Routes:        make([]common.Route, 0),
			GlobalFilters: make([]GlobalFilter, 0),
			Port:          8080, // 默认端口
		},
	}
}

// Load loads config from file
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

// Save saves config to file
func (scm *StaticConfigManager) Save(configPath string) error {
	scm.mutex.RLock()
	defer scm.mutex.RUnlock()

	data, err := json.MarshalIndent(scm.config, "", "  ")
	if err != nil {
		return err
	}

	return ioutil.WriteFile(configPath, data, 0644)
}

// GetRoutes gets all routes
func (scm *StaticConfigManager) GetRoutes() []common.Route {
	scm.mutex.RLock()
	defer scm.mutex.RUnlock()

	// 返回副本以避免外部修改
	routes := make([]common.Route, len(scm.config.Routes))
	copy(routes, scm.config.Routes)
	return routes
}

// AddRoute adds a route
func (scm *StaticConfigManager) AddRoute(route common.Route) {
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

// UpdateRoute updates a route
func (scm *StaticConfigManager) UpdateRoute(route common.Route) error {
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

// DeleteRoute deletes a route
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

// GetConfig gets full config
func (scm *StaticConfigManager) GetConfig() Config {
	scm.mutex.RLock()
	defer scm.mutex.RUnlock()

	// 返回副本
	config := scm.config

	// 复制路由切片
	config.Routes = make([]common.Route, len(scm.config.Routes))
	copy(config.Routes, scm.config.Routes)

	// 复制全局过滤器切片
	config.GlobalFilters = make([]GlobalFilter, len(scm.config.GlobalFilters))
	copy(config.GlobalFilters, scm.config.GlobalFilters)

	return config
}

// SetConfig sets full config
func (scm *StaticConfigManager) SetConfig(config Config) {
	scm.mutex.Lock()
	defer scm.mutex.Unlock()

	scm.config = config
}
