// loader.go 配置加载器
// 功能点：
// 1. 从YAML文件加载配置
// 2. 从环境变量加载配置（覆盖YAML配置）
// 3. 支持多环境配置（dev/prod）
// 4. 提供配置热重载功能
// 5. 提供配置项获取方法
// 6. 支持配置项默认值设置

package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/goccy/go-yaml"
)

// Loader 配置加载器结构体
type Loader struct {
	config *Config
	path   string
}

// NewLoader 创建配置加载器实例
func NewLoader(path string) *Loader {
	return &Loader{
		path: path,
	}
}

// Load 加载配置
func (l *Loader) Load() (*Config, error) {
	// 尝试从YAML文件加载配置
	config, err := l.LoadFromYAML()
	if err != nil {
		// 如果YAML加载失败，返回默认配置
		config = l.getDefaultConfig()
	}

	// 从环境变量加载配置，覆盖YAML配置
	config = l.LoadFromEnv(config)

	// 验证配置
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("配置验证失败: %w", err)
	}

	l.config = config
	return config, nil
}

// LoadFromYAML 从YAML文件加载配置
func (l *Loader) LoadFromYAML() (*Config, error) {
	// 检查文件是否存在
	if _, err := os.Stat(l.path); os.IsNotExist(err) {
		return nil, fmt.Errorf("配置文件不存在: %s", l.path)
	}

	// 读取YAML文件
	data, err := os.ReadFile(l.path)
	if err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %w", err)
	}

	// 解析YAML
	config := &Config{}
	if err := yaml.Unmarshal(data, config); err != nil {
		return nil, fmt.Errorf("解析YAML配置失败: %w", err)
	}

	return config, nil
}

// LoadFromEnv 从环境变量加载配置
func (l *Loader) LoadFromEnv(config *Config) *Config {
	if config == nil {
		config = l.getDefaultConfig()
	}

	// 服务器配置
	if host := os.Getenv("SERVER_HOST"); host != "" {
		config.Server.Host = host
	}
	if port := os.Getenv("SERVER_PORT"); port != "" {
		if p, err := strconv.Atoi(port); err == nil {
			config.Server.Port = p
		}
	}

	// OCR配置
	if secretID := os.Getenv("OCR_SECRET_ID"); secretID != "" {
		config.OCR.SecretID = secretID
	}
	if secretKey := os.Getenv("OCR_SECRET_KEY"); secretKey != "" {
		config.OCR.SecretKey = secretKey
	}
	if region := os.Getenv("OCR_REGION"); region != "" {
		config.OCR.Region = region
	}

	return config
}

// GetConfig 获取配置
func (l *Loader) GetConfig() *Config {
	return l.config
}

// GetServerConfig 获取服务器配置
func (l *Loader) GetServerConfig() ServerConfig {
	if l.config == nil {
		return ServerConfig{}
	}
	return l.config.Server
}

// GetDatabaseConfig 获取数据库配置
func (l *Loader) GetDatabaseConfig() DatabaseConfig {
	if l.config == nil {
		return DatabaseConfig{}
	}
	return l.config.Database
}

// GetRedisConfig 获取Redis配置
func (l *Loader) GetRedisConfig() RedisConfig {
	if l.config == nil {
		return RedisConfig{}
	}
	return l.config.Redis
}

// GetLLMConfig 获取大模型配置
func (l *Loader) GetLLMConfig() LLMConfig {
	if l.config == nil {
		return LLMConfig{}
	}
	return l.config.LLM
}

// GetOCRConfig 获取OCR配置
func (l *Loader) GetOCRConfig() OCRConfig {
	if l.config == nil {
		return OCRConfig{}
	}
	return l.config.OCR
}

// GetStorageConfig 获取存储配置
func (l *Loader) GetStorageConfig() StorageConfig {
	if l.config == nil {
		return StorageConfig{}
	}
	return l.config.Storage
}

// GetLoggerConfig 获取日志配置
func (l *Loader) GetLoggerConfig() LoggerConfig {
	if l.config == nil {
		return LoggerConfig{}
	}
	return l.config.Logger
}

// GetSecurityConfig 获取安全配置
func (l *Loader) GetSecurityConfig() SecurityConfig {
	if l.config == nil {
		return SecurityConfig{}
	}
	return l.config.Security
}

// GetAppConfig 获取应用配置
func (l *Loader) GetAppConfig() AppConfig {
	if l.config == nil {
		return AppConfig{}
	}
	return l.config.App
}

// Reload 重新加载配置
func (l *Loader) Reload() error {
	// TODO: 实现配置重新加载逻辑
	return nil
}

// Save 保存配置到文件
func (l *Loader) Save(config *Config) error {
	// TODO: 实现配置保存逻辑
	return nil
}

// GetConfigPath 获取配置文件路径
func (l *Loader) GetConfigPath() string {
	return l.path
}

// getDefaultConfig 获取默认配置
func (l *Loader) getDefaultConfig() *Config {
	return &Config{
		Server: ServerConfig{
			Host: "localhost",
			Port: 8080,
		},
		Database: DatabaseConfig{
			Host:   "localhost",
			Port:   5432,
			DBName: "default",
		},
		Redis: RedisConfig{
			Host: "localhost",
			Port: 6379,
		},
		OCR: OCRConfig{
			Provider:   "tencent",
			Region:     "ap-beijing",
			Timeout:    30,
			MaxRetries: 3,
		},
		Logger: LoggerConfig{
			Level: "info",
		},
	}
}

// SetConfigPath 设置配置文件路径
func (l *Loader) SetConfigPath(path string) {
	l.path = path
}

// GetEnv 获取环境变量，支持默认值
func GetEnv(key, defaultValue string) string {
	// TODO: 实现获取环境变量逻辑
	return ""
}

// GetEnvAsInt 获取环境变量并转换为int类型
func GetEnvAsInt(key string, defaultValue int) int {
	// TODO: 实现获取环境变量并转换为int类型逻辑
	return 0
}

// GetEnvAsBool 获取环境变量并转换为bool类型
func GetEnvAsBool(key string, defaultValue bool) bool {
	// TODO: 实现获取环境变量并转换为bool类型逻辑
	return false
}

// GetConfigFile 根据环境获取配置文件路径
func GetConfigFile(env string) string {
	// TODO: 实现根据环境获取配置文件路径逻辑
	return ""
}

// EnsureConfigDir 确保配置目录存在
func EnsureConfigDir(path string) error {
	// TODO: 实现确保配置目录存在逻辑
	return nil
}
