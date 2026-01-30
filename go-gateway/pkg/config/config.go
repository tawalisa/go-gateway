package config

import (
	"fmt"
	"sync"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
	"go-gateway/pkg/common"
)

// Config defines configuration structure
type Config struct {
	Routes        []common.Route `json:"routes" mapstructure:"routes"`
	GlobalFilters []GlobalFilter `json:"global_filters" mapstructure:"global_filters"`
	Port          int            `json:"port" mapstructure:"port"`
}

// GlobalFilter defines global filter
type GlobalFilter struct {
	Name string      `json:"name" mapstructure:"name"`
	Args interface{} `json:"args" mapstructure:"args"`
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

// ViperConfigManager viper-based configuration manager
type ViperConfigManager struct {
	mutex  sync.RWMutex
	config Config
	viper  *viper.Viper
}

// NewViperConfigManager creates a new viper config manager
func NewViperConfigManager() *ViperConfigManager {
	v := viper.New()
	return &ViperConfigManager{
		config: Config{
			Routes:        make([]common.Route, 0),
			GlobalFilters: make([]GlobalFilter, 0),
			Port:          8080, // 默认端口
		},
		viper: v,
	}
}

// Load loads config from file using viper
func (vcm *ViperConfigManager) Load(configPath string) error {
	vcm.viper.SetConfigFile(configPath)

	if err := vcm.viper.ReadInConfig(); err != nil {
		return fmt.Errorf("error reading config file: %w", err)
	}

	// 将配置解码到结构体中
	var config Config
	if err := vcm.viper.Unmarshal(&config); err != nil {
		return fmt.Errorf("error unmarshaling config: %w", err)
	}

	vcm.mutex.Lock()
	defer vcm.mutex.Unlock()
	vcm.config = config

	return nil
}

// Save saves config to file using viper
func (vcm *ViperConfigManager) Save(configPath string) error {
	vcm.mutex.RLock()
	defer vcm.mutex.RUnlock()

	// 设置配置值
	vcm.viper.Set("routes", vcm.config.Routes)
	vcm.viper.Set("global_filters", vcm.config.GlobalFilters)
	vcm.viper.Set("port", vcm.config.Port)

	// 写入文件
	if err := vcm.viper.WriteConfigAs(configPath); err != nil {
		// 如果配置文件不存在，使用SafeWriteConfigAs创建它
		if err := vcm.viper.SafeWriteConfigAs(configPath); err != nil {
			return fmt.Errorf("error saving config file: %w", err)
		}
	}

	return nil
}

// GetRoutes gets all routes
func (vcm *ViperConfigManager) GetRoutes() []common.Route {
	vcm.mutex.RLock()
	defer vcm.mutex.RUnlock()

	// 返回副本以避免外部修改
	routes := make([]common.Route, len(vcm.config.Routes))
	copy(routes, vcm.config.Routes)
	return routes
}

// AddRoute adds a route
func (vcm *ViperConfigManager) AddRoute(route common.Route) {
	vcm.mutex.Lock()
	defer vcm.mutex.Unlock()

	// 检查是否已存在相同ID的路由
	for i, r := range vcm.config.Routes {
		if r.ID == route.ID {
			// 如果存在，则替换
			vcm.config.Routes[i] = route
			return
		}
	}

	// 添加新路由
	vcm.config.Routes = append(vcm.config.Routes, route)
}

// UpdateRoute updates a route
func (vcm *ViperConfigManager) UpdateRoute(route common.Route) error {
	vcm.mutex.Lock()
	defer vcm.mutex.Unlock()

	for i, r := range vcm.config.Routes {
		if r.ID == route.ID {
			vcm.config.Routes[i] = route
			return nil
		}
	}

	return fmt.Errorf("route with id %s not found", route.ID)
}

// DeleteRoute deletes a route
func (vcm *ViperConfigManager) DeleteRoute(id string) error {
	vcm.mutex.Lock()
	defer vcm.mutex.Unlock()

	for i, route := range vcm.config.Routes {
		if route.ID == id {
			// 从切片中移除元素
			vcm.config.Routes = append(vcm.config.Routes[:i], vcm.config.Routes[i+1:]...)
			return nil
		}
	}

	return fmt.Errorf("route with id %s not found", id)
}

// GetConfig gets full config
func (vcm *ViperConfigManager) GetConfig() Config {
	vcm.mutex.RLock()
	defer vcm.mutex.RUnlock()

	// 返回副本
	config := vcm.config

	// 复制路由切片
	config.Routes = make([]common.Route, len(vcm.config.Routes))
	copy(config.Routes, vcm.config.Routes)

	// 复制全局过滤器切片
	config.GlobalFilters = make([]GlobalFilter, len(vcm.config.GlobalFilters))
	copy(config.GlobalFilters, vcm.config.GlobalFilters)

	return config
}

// SetConfig sets full config
func (vcm *ViperConfigManager) SetConfig(config Config) {
	vcm.mutex.Lock()
	defer vcm.mutex.Unlock()

	vcm.config = config
}

// 监听配置变化的功能
func (vcm *ViperConfigManager) WatchConfig(onChange func()) {
	vcm.viper.OnConfigChange(func(e fsnotify.Event) {
		fmt.Println("Config file changed:", e.Name)
		if onChange != nil {
			onChange()
		}
	})
	vcm.viper.WatchConfig()
}
