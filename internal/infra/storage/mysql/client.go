// client.go MySQL客户端实现
// 功能点：
// 1. 实现MySQL数据库连接（使用GORM）
// 2. 实现连接池管理
// 3. 实现事务管理
// 4. 提供数据库操作方法
// 5. 支持上下文传递
// 6. 支持健康检查

package mysql

import (
	"context"
	"fmt"
	"sync"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Client MySQL客户端结构体
type Client struct {
	db     *gorm.DB
	config *Config
	mu     sync.RWMutex
}

// NewClient 创建MySQL客户端实例
func NewClient() *Client {
	return &Client{}
}

// Connect 连接数据库
func (c *Client) Connect(ctx context.Context, config *Config) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// 构建数据源名称
	dsn := config.GetDSN()

	// 配置GORM日志级别
	var logLevel logger.LogLevel
	switch config.LogLevel {
	case "silent":
		logLevel = logger.Silent
	case "error":
		logLevel = logger.Error
	case "warn":
		logLevel = logger.Warn
	case "info":
		logLevel = logger.Info
	default:
		logLevel = logger.Warn
	}

	// 打开数据库连接
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logLevel),
	})
	if err != nil {
		return fmt.Errorf("打开数据库连接失败: %w", err)
	}

	// 获取底层sql.DB对象以配置连接池
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("获取底层SQL数据库连接失败: %w", err)
	}

	// 设置连接池参数
	sqlDB.SetMaxOpenConns(config.MaxOpenConns)
	sqlDB.SetMaxIdleConns(config.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(config.ConnMaxLifetime)
	sqlDB.SetConnMaxIdleTime(config.ConnMaxIdleTime)

	// 测试连接
	if err := sqlDB.PingContext(ctx); err != nil {
		return fmt.Errorf("数据库连接测试失败: %w", err)
	}

	c.db = db
	c.config = config

	return nil
}

// Disconnect 断开数据库连接
func (c *Client) Disconnect(ctx context.Context) error {
	return c.Close()
}

// Ping 检查数据库连接
func (c *Client) Ping(ctx context.Context) error {
	sqlDB, err := c.GetDB().DB()
	if err != nil {
		return fmt.Errorf("获取底层SQL数据库连接失败: %w", err)
	}
	return sqlDB.PingContext(ctx)
}

// GetDB 获取GORM数据库连接
func (c *Client) GetDB() *gorm.DB {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.db
}

// Begin 开始事务
func (c *Client) Begin(ctx context.Context) *gorm.DB {
	return c.GetDB().Begin()
}

// Execute 执行SQL语句（使用GORM）
func (c *Client) Execute(ctx context.Context, query string, args ...interface{}) (int64, error) {
	result := c.GetDB().Exec(query, args...)
	if result.Error != nil {
		return 0, result.Error
	}
	return result.RowsAffected, nil
}

// Migrate 执行数据库迁移
func (c *Client) Migrate(ctx context.Context, models ...interface{}) error {
	return c.GetDB().AutoMigrate(models...)
}

// GetConnectionStats 获取连接统计信息
func (c *Client) GetConnectionStats() map[string]interface{} {
	sqlDB, err := c.GetDB().DB()
	if err != nil {
		return map[string]interface{}{
			"error": fmt.Sprintf("获取底层SQL数据库连接失败: %v", err),
		}
	}

	stats := sqlDB.Stats()
	return map[string]interface{}{
		"MaxOpenConnections": stats.MaxOpenConnections,
		"OpenConnections":    stats.OpenConnections,
		"InUse":              stats.InUse,
		"Idle":               stats.Idle,
		"WaitCount":          stats.WaitCount,
		"WaitDuration":       stats.WaitDuration,
		"MaxIdleClosed":      stats.MaxIdleClosed,
		"MaxIdleTimeClosed":  stats.MaxIdleTimeClosed,
		"MaxLifetimeClosed":  stats.MaxLifetimeClosed,
	}
}

// Close 关闭数据库连接
func (c *Client) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.db != nil {
		sqlDB, err := c.db.DB()
		if err != nil {
			return fmt.Errorf("获取底层SQL数据库连接失败: %w", err)
		}
		return sqlDB.Close()
	}

	return nil
}

// IsConnected 检查是否已连接
func (c *Client) IsConnected() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.db != nil
}

// Reconnect 重新连接数据库
func (c *Client) Reconnect(ctx context.Context) error {
	if c.IsConnected() {
		if err := c.Disconnect(ctx); err != nil {
			return fmt.Errorf("断开连接失败: %w", err)
		}
	}

	if c.config == nil {
		return fmt.Errorf("配置信息不存在，无法重新连接")
	}

	return c.Connect(ctx, c.config)
}
