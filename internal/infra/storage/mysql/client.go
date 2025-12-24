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
	"errors"
	"sync"

	"reimbursement-audit/internal/pkg/logger"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"
)

// Client MySQL客户端结构体
type Client struct {
	db     *gorm.DB
	config *Config
	logger logger.Logger
	mu     sync.RWMutex
}

// NewClient 创建MySQL客户端实例
func NewClient(logger logger.Logger) *Client {
	return &Client{
		logger: logger,
	}
}

// Connect 连接数据库
func (c *Client) Connect(ctx context.Context, config *Config) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// 构建数据源名称
	dsn := config.GetDSN()

	// 配置GORM日志级别
	var logLevel gormLogger.LogLevel
	switch config.LogLevel {
	case "silent":
		logLevel = gormLogger.Silent
	case "error":
		logLevel = gormLogger.Error
	case "warn":
		logLevel = gormLogger.Warn
	case "info":
		logLevel = gormLogger.Info
	default:
		logLevel = gormLogger.Warn
	}

	// 打开数据库连接
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: gormLogger.Default.LogMode(logLevel),
	})
	if err != nil {
		c.logger.WithContext(ctx).Error("打开数据库连接失败",
			logger.NewField("error", err.Error()))
		return errors.New("打开数据库连接失败")
	}

	// 获取底层sql.DB对象以配置连接池
	sqlDB, err := db.DB()
	if err != nil {
		c.logger.WithContext(ctx).Error("获取底层SQL数据库连接失败",
			logger.NewField("error", err.Error()))
		return errors.New("获取底层SQL数据库连接失败")
	}

	// 设置连接池参数
	sqlDB.SetMaxOpenConns(config.MaxOpenConns)
	sqlDB.SetMaxIdleConns(config.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(config.ConnMaxLifetime)
	sqlDB.SetConnMaxIdleTime(config.ConnMaxIdleTime)

	// 测试连接
	if err := sqlDB.PingContext(ctx); err != nil {
		c.logger.WithContext(ctx).Error("数据库连接测试失败",
			logger.NewField("error", err.Error()))
		return errors.New("数据库连接测试失败")
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
		c.logger.WithContext(ctx).Error("获取底层SQL数据库连接失败",
			logger.NewField("error", err.Error()))
		return errors.New("获取底层SQL数据库连接失败")
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
		c.logger.Error("获取底层SQL数据库连接失败",
			logger.NewField("error", err.Error()))
		return map[string]interface{}{
			"error": "获取底层SQL数据库连接失败",
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
			c.logger.Error("获取底层SQL数据库连接失败",
				logger.NewField("error", err.Error()))
			return errors.New("获取底层SQL数据库连接失败")
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
			c.logger.WithContext(ctx).Error("断开连接失败",
				logger.NewField("error", err.Error()))
			return errors.New("断开连接失败")
		}
	}

	if c.config == nil {
		c.logger.WithContext(ctx).Error("配置信息不存在，无法重新连接")
		return errors.New("配置信息不存在，无法重新连接")
	}

	return c.Connect(ctx, c.config)
}
