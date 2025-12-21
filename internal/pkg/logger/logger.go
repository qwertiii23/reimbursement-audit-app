package logger

import (
	"context"
	"io"
)

// Level 日志级别
type Level int

const (
	// DebugLevel 调试级别
	DebugLevel Level = iota
	// InfoLevel 信息级别
	InfoLevel
	// WarnLevel 警告级别
	WarnLevel
	// ErrorLevel 错误级别
	ErrorLevel
	// FatalLevel 致命级别
	FatalLevel
)

// String 返回日志级别的字符串表示
func (l Level) String() string {
	switch l {
	case DebugLevel:
		return "DEBUG"
	case InfoLevel:
		return "INFO"
	case WarnLevel:
		return "WARN"
	case ErrorLevel:
		return "ERROR"
	case FatalLevel:
		return "FATAL"
	default:
		return "UNKNOWN"
	}
}

// Field 日志字段
type Field struct {
	Key   string
	Value interface{}
}

// Logger 日志接口
type Logger interface {
	// Debug 记录调试日志
	Debug(msg string, fields ...Field)
	// Info 记录信息日志
	Info(msg string, fields ...Field)
	// Warn 记录警告日志
	Warn(msg string, fields ...Field)
	// Error 记录错误日志
	Error(msg string, fields ...Field)
	// Fatal 记录致命日志
	Fatal(msg string, fields ...Field)

	// WithContext 添加上下文
	WithContext(ctx context.Context) Logger
	// WithFields 添加字段
	WithFields(fields ...Field) Logger
	// WithField 添加单个字段
	WithField(key string, value interface{}) Logger

	// SetLevel 设置日志级别
	SetLevel(level Level)
	// GetLevel 获取日志级别
	GetLevel() Level

	// SetOutput 设置输出
	SetOutput(w io.Writer)
	// Close 关闭日志器
	Close() error
}

// Config 日志配置
type Config struct {
	Level      Level  `json:"level"`       // 日志级别
	Format     string `json:"format"`      // 输出格式 (json/text)
	Output     string `json:"output"`      // 输出目标 (stdout/stderr/file)
	Filename   string `json:"filename"`    // 文件名 (当output为file时)
	MaxSize    int    `json:"max_size"`    // 单个日志文件最大大小(MB)
	MaxBackups int    `json:"max_backups"` // 保留的旧日志文件数量
	MaxAge     int    `json:"max_age"`     // 保留日志文件的最大天数
	Compress   bool   `json:"compress"`    // 是否压缩旧日志文件
}

// DefaultConfig 返回默认配置
func DefaultConfig() *Config {
	return &Config{
		Level:      InfoLevel,
		Format:     "json",
		Output:     "stdout",
		MaxSize:    100,
		MaxBackups: 3,
		MaxAge:     28,
		Compress:   true,
	}
}

// Validate 验证配置
func (c *Config) Validate() error {
	// TODO: 实现配置验证逻辑
	return nil
}

// NewField 创建日志字段
func NewField(key string, value interface{}) Field {
	return Field{
		Key:   key,
		Value: value,
	}
}
