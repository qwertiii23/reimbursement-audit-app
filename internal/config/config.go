// config.go 配置结构体
// 功能点：
// 1. 定义系统配置结构体
// 2. 定义数据库配置结构体
// 3. 定义大模型API配置结构体
// 4. 定义存储配置结构体
// 5. 定义日志配置结构体
// 6. 提供配置验证方法

package config

import "fmt"

// Config 系统配置结构体
type Config struct {
	Server   ServerConfig   `json:"server" yaml:"server"`     // 服务器配置
	Database DatabaseConfig `json:"database" yaml:"database"` // 数据库配置
	Redis    RedisConfig    `json:"redis" yaml:"redis"`       // Redis配置
	LLM      LLMConfig      `json:"llm" yaml:"llm"`           // 大模型配置
	OCR      OCRConfig      `json:"ocr" yaml:"ocr"`           // OCR配置
	Storage  StorageConfig  `json:"storage" yaml:"storage"`   // 存储配置
	Logger   LoggerConfig   `json:"logger" yaml:"logger"`     // 日志配置
	Security SecurityConfig `json:"security" yaml:"security"` // 安全配置
	App      AppConfig      `json:"app" yaml:"app"`           // 应用配置
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Host         string `json:"host" yaml:"host"`                   // 服务器主机
	Port         int    `json:"port" yaml:"port"`                   // 服务器端口
	ReadTimeout  int    `json:"read_timeout" yaml:"read_timeout"`   // 读超时时间(秒)
	WriteTimeout int    `json:"write_timeout" yaml:"write_timeout"` // 写超时时间(秒)
	IdleTimeout  int    `json:"idle_timeout" yaml:"idle_timeout"`   // 空闲超时时间(秒)
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Host         string `json:"host" yaml:"host"`                     // 数据库主机
	Port         int    `json:"port" yaml:"port"`                     // 数据库端口
	Username     string `json:"username" yaml:"username"`             // 用户名
	Password     string `json:"password" yaml:"password"`             // 密码
	DBName       string `json:"dbname" yaml:"dbname"`                 // 数据库名
	SSLMode      string `json:"sslmode" yaml:"sslmode"`               // SSL模式
	MaxOpenConns int    `json:"max_open_conns" yaml:"max_open_conns"` // 最大打开连接数
	MaxIdleConns int    `json:"max_idle_conns" yaml:"max_idle_conns"` // 最大空闲连接数
}

// RedisConfig Redis配置
type RedisConfig struct {
	Host     string `json:"host" yaml:"host"`         // Redis主机
	Port     int    `json:"port" yaml:"port"`         // Redis端口
	Password string `json:"password" yaml:"password"` // Redis密码
	DB       int    `json:"db" yaml:"db"`             // Redis数据库
}

// LLMConfig 大模型配置
type LLMConfig struct {
	Provider    string  `json:"provider" yaml:"provider"`       // 提供商(zhipu/wenxin等)
	APIKey      string  `json:"api_key" yaml:"api_key"`         // API密钥
	BaseURL     string  `json:"base_url" yaml:"base_url"`       // 基础URL
	Model       string  `json:"model" yaml:"model"`             // 模型名称
	MaxTokens   int     `json:"max_tokens" yaml:"max_tokens"`   // 最大令牌数
	Temperature float64 `json:"temperature" yaml:"temperature"` // 温度参数
	Timeout     int     `json:"timeout" yaml:"timeout"`         // 超时时间(秒)
}

// OCRConfig OCR配置
type OCRConfig struct {
	Provider   string `json:"provider" yaml:"provider"`       // OCR提供商(tencent)
	SecretID   string `json:"secret_id" yaml:"secret_id"`     // 腾讯云SecretId
	SecretKey  string `json:"secret_key" yaml:"secret_key"`   // 腾讯云SecretKey
	Region     string `json:"region" yaml:"region"`           // 腾讯云地域
	Timeout    int    `json:"timeout" yaml:"timeout"`         // 超时时间(秒)
	MaxRetries int    `json:"max_retries" yaml:"max_retries"` // 最大重试次数
}

// StorageConfig 存储配置
type StorageConfig struct {
	Type  string             `json:"type" yaml:"type"`   // 存储类型(local/minio)
	Local LocalStorageConfig `json:"local" yaml:"local"` // 本地存储配置
	MinIO MinIOConfig        `json:"minio" yaml:"minio"` // MinIO存储配置
}

// LocalStorageConfig 本地存储配置
type LocalStorageConfig struct {
	Path string `json:"path" yaml:"path"` // 存储路径
}

// MinIOConfig MinIO配置
type MinIOConfig struct {
	Endpoint  string `json:"endpoint" yaml:"endpoint"`     // MinIO端点
	AccessKey string `json:"access_key" yaml:"access_key"` // 访问密钥
	SecretKey string `json:"secret_key" yaml:"secret_key"` // 秘密密钥
	Bucket    string `json:"bucket" yaml:"bucket"`         // 存储桶
	UseSSL    bool   `json:"use_ssl" yaml:"use_ssl"`       // 是否使用SSL
}

// LoggerConfig 日志配置
type LoggerConfig struct {
	Level      string `json:"level" yaml:"level"`             // 日志级别
	Format     string `json:"format" yaml:"format"`           // 日志格式(json/text)
	Output     string `json:"output" yaml:"output"`           // 输出路径(stdout/file)
	Filename   string `json:"filename" yaml:"filename"`       // 文件名
	MaxSize    int    `json:"max_size" yaml:"max_size"`       // 最大文件大小(MB)
	MaxBackups int    `json:"max_backups" yaml:"max_backups"` // 最大备份文件数
	MaxAge     int    `json:"max_age" yaml:"max_age"`         // 最大保存天数
	Compress   bool   `json:"compress" yaml:"compress"`       // 是否压缩
}

// SecurityConfig 安全配置
type SecurityConfig struct {
	JWTSecret    string   `json:"jwt_secret" yaml:"jwt_secret"`       // JWT密钥
	JWTExpire    int      `json:"jwt_expire" yaml:"jwt_expire"`       // JWT过期时间(小时)
	PasswordSalt string   `json:"password_salt" yaml:"password_salt"` // 密码盐值
	EnableHTTPS  bool     `json:"enable_https" yaml:"enable_https"`   // 是否启用HTTPS
	CertFile     string   `json:"cert_file" yaml:"cert_file"`         // 证书文件
	KeyFile      string   `json:"key_file" yaml:"key_file"`           // 私钥文件
	EnableCORS   bool     `json:"enable_cors" yaml:"enable_cors"`     // 是否启用CORS
	TrustedIPs   []string `json:"trusted_ips" yaml:"trusted_ips"`     // 信任IP列表
}

// AppConfig 应用配置
type AppConfig struct {
	Name        string `json:"name" yaml:"name"`               // 应用名称
	Version     string `json:"version" yaml:"version"`         // 应用版本
	Environment string `json:"environment" yaml:"environment"` // 运行环境
	Debug       bool   `json:"debug" yaml:"debug"`             // 调试模式
	TimeZone    string `json:"timezone" yaml:"timezone"`       // 时区
}

// Validate 验证配置
func (c *Config) Validate() error {
	if c == nil {
		return fmt.Errorf("配置不能为空")
	}

	// 验证服务器配置
	if c.Server.Host == "" {
		return fmt.Errorf("服务器主机不能为空")
	}
	if c.Server.Port <= 0 || c.Server.Port > 65535 {
		return fmt.Errorf("服务器端口必须在1-65535范围内")
	}

	return nil
}

// IsProduction 是否为生产环境
func (c *Config) IsProduction() bool {
	// TODO: 实现环境判断逻辑
	return false
}

// IsDevelopment 是否为开发环境
func (c *Config) IsDevelopment() bool {
	// TODO: 实现环境判断逻辑
	return false
}
