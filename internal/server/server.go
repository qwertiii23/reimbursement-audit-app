package server

import (
	"context"
	"fmt"
	"net/http"
	"reimbursement-audit/internal/config"
	"time"

	"github.com/gin-gonic/gin"
)

// Server 服务器接口
type Server interface {
	// Start 启动服务器
	Start() error
	// Stop 停止服务器
	Stop(ctx context.Context) error
	// GetEngine 获取Gin引擎
	GetEngine() *gin.Engine
	// GetConfig 获取服务器配置
	GetConfig() *Config
	// SetConfig 设置服务器配置
	SetConfig(config *Config)
	// SetAppConfig 设置应用配置
	SetAppConfig(config *config.Config)
	// RegisterRoutes 注册路由
	RegisterRoutes()
}

// Config 服务器配置
type Config struct {
	Host         string        `json:"host"`          // 服务器主机
	Port         int           `json:"port"`          // 服务器端口
	ReadTimeout  time.Duration `json:"read_timeout"`  // 读取超时时间
	WriteTimeout time.Duration `json:"write_timeout"` // 写入超时时间
	IdleTimeout  time.Duration `json:"idle_timeout"`  // 空闲超时时间
	Mode         string        `json:"mode"`          // 运行模式 (debug/release/test)
	TLS          bool          `json:"tls"`           // 是否启用TLS
	CertFile     string        `json:"cert_file"`     // TLS证书文件路径
	KeyFile      string        `json:"key_file"`      // TLS私钥文件路径
}

// DefaultConfig 返回默认配置
func DefaultConfig() *Config {
	return &Config{
		Host:         "0.0.0.0",
		Port:         8080,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
		Mode:         gin.ReleaseMode,
		TLS:          false,
	}
}

// Validate 验证配置
func (c *Config) Validate() error {
	if c.Port <= 0 || c.Port > 65535 {
		return fmt.Errorf("invalid port: %d", c.Port)
	}

	if c.TLS {
		if c.CertFile == "" {
			return fmt.Errorf("cert file is required when TLS is enabled")
		}
		if c.KeyFile == "" {
			return fmt.Errorf("key file is required when TLS is enabled")
		}
	}

	return nil
}

// GetAddress 获取服务器地址
func (c *Config) GetAddress() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

// IsTLS 检查是否启用TLS
func (c *Config) IsTLS() bool {
	return c.TLS
}

// SetMode 设置运行模式
func (c *Config) SetMode(mode string) {
	c.Mode = mode
}

// GetMode 获取运行模式
func (c *Config) GetMode() string {
	return c.Mode
}

// SetTimeouts 设置超时时间
func (c *Config) SetTimeouts(readTimeout, writeTimeout, idleTimeout time.Duration) {
	c.ReadTimeout = readTimeout
	c.WriteTimeout = writeTimeout
	c.IdleTimeout = idleTimeout
}

// NewServer 创建服务器实例
func NewServer(config *Config) Server {
	if config == nil {
		config = DefaultConfig()
	}

	// 设置Gin模式
	gin.SetMode(config.Mode)

	// 创建Gin引擎
	engine := gin.New()

	// 添加中间件
	engine.Use(gin.Logger())
	engine.Use(gin.Recovery())

	return &serverImpl{
		config: config,
		engine: engine,
	}
}

// RunServer 运行服务器
func RunServer(config *Config) error {
	server := NewServer(config)
	server.RegisterRoutes()
	return server.Start()
}

// RunServerWithContext 带上下文运行服务器
func RunServerWithContext(ctx context.Context, config *Config) error {
	server := NewServer(config)
	server.RegisterRoutes()

	// 启动服务器
	go func() {
		if err := server.Start(); err != nil && err != http.ErrServerClosed {
			panic(err)
		}
	}()

	// 等待上下文取消
	<-ctx.Done()

	// 停止服务器
	return server.Stop(context.Background())
}

// HealthCheck 健康检查
func HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "ok",
		"timestamp": time.Now().Unix(),
	})
}

// ReadyCheck 就绪检查
func ReadyCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "ready",
		"timestamp": time.Now().Unix(),
	})
}

// VersionCheck 版本检查
func VersionCheck(version string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"version":   version,
			"timestamp": time.Now().Unix(),
		})
	}
}
