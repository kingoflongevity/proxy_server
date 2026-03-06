package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"

	"proxy_server/pkg/logger"
)

// Config 应用配置
type Config struct {
	Server   ServerConfig   `json:"server"`
	Database DatabaseConfig `json:"database"`
	Proxy    ProxyConfig    `json:"proxy"`
	Log      LogConfig      `json:"log"`
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Port         int    `json:"port"`
	Host         string `json:"host"`
	ReadTimeout  int    `json:"read_timeout"`  // 秒
	WriteTimeout int    `json:"write_timeout"` // 秒
	Mode         string `json:"mode"`          // debug, release, test
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Type string `json:"type"` // json, mysql, postgres
	Path string `json:"path"` // 数据文件路径
}

// ProxyConfig 代理配置
type ProxyConfig struct {
	LocalPort    int    `json:"local_port"`    // 本地代理端口
	SocksPort    int    `json:"socks_port"`    // SOCKS5端口
	HTTPPort     int    `json:"http_port"`     // HTTP代理端口
	EnableDNS    bool   `json:"enable_dns"`    // 启用DNS代理
	DNSPort      int    `json:"dns_port"`      // DNS端口
	LogLevel     string `json:"log_level"`     // xray日志级别
}

// LogConfig 日志配置
type LogConfig struct {
	Level  string `json:"level"`   // debug, info, warn, error
	Format string `json:"format"`  // text, json
	Output string `json:"output"`  // stdout, file
	Path   string `json:"path"`    // 日志文件路径
}

var (
	configInstance *Config
	configOnce     sync.Once
)

// GetConfig 获取配置实例（单例模式）
func GetConfig() *Config {
	configOnce.Do(func() {
		configInstance = &Config{
			Server: ServerConfig{
				Port:         8000,
				Host:         "0.0.0.0",
				ReadTimeout:  60,
				WriteTimeout: 60,
				Mode:         "debug",
			},
			Database: DatabaseConfig{
				Type: "json",
				Path: "./data",
			},
			Proxy: ProxyConfig{
				LocalPort: 10808,
				SocksPort: 1080,
				HTTPPort:  8080,
				EnableDNS: false,
				DNSPort:   53,
				LogLevel:  "warning",
			},
			Log: LogConfig{
				Level:  "info",
				Format: "text",
				Output: "stdout",
			},
		}
	})
	return configInstance
}

// LoadConfig 从文件加载配置
func LoadConfig(configPath string) error {
	config := GetConfig()
	
	// 如果配置文件不存在，使用默认配置
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		logger.Info("配置文件不存在，使用默认配置: %s", configPath)
		return config.SaveConfig(configPath)
	}
	
	// 读取配置文件
	data, err := os.ReadFile(configPath)
	if err != nil {
		return err
	}
	
	// 解析配置
	if err := json.Unmarshal(data, config); err != nil {
		return err
	}
	
	logger.Info("配置加载成功: %s", configPath)
	return nil
}

// SaveConfig 保存配置到文件
func (c *Config) SaveConfig(configPath string) error {
	// 确保目录存在
	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	
	// 序列化配置
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	
	// 写入文件
	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return err
	}
	
	logger.Info("配置保存成功: %s", configPath)
	return nil
}

// GetDefaultConfigPath 获取默认配置文件路径
func GetDefaultConfigPath() string {
	return "./data/config.json"
}
