package logger

import (
	"context"
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	"gopkg.in/natefinch/lumberjack.v2"
)

// loggerImpl 日志实现
type loggerImpl struct {
	config  *Config
	output  io.Writer
	mu      sync.RWMutex
	fields  []Field
	context context.Context
}

// NewLogger 创建日志器实例
func NewLogger(config *Config) (Logger, error) {
	if config == nil {
		config = DefaultConfig()
	}

	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid logger config: %w", err)
	}

	l := &loggerImpl{
		config:  config,
		context: context.Background(),
	}

	// 设置输出
	if err := l.setOutput(); err != nil {
		return nil, err
	}

	return l, nil
}

// setOutput 设置输出
func (l *loggerImpl) setOutput() error {
	switch l.config.Output {
	case "stdout":
		l.output = os.Stdout
	case "stderr":
		l.output = os.Stderr
	case "file":
		if l.config.Filename == "" {
			return fmt.Errorf("filename is required when output is file")
		}
		l.output = &lumberjack.Logger{
			Filename:   l.config.Filename,
			MaxSize:    l.config.MaxSize,
			MaxBackups: l.config.MaxBackups,
			MaxAge:     l.config.MaxAge,
			Compress:   l.config.Compress,
		}
	default:
		return fmt.Errorf("unsupported output: %s", l.config.Output)
	}
	return nil
}

// Debug 记录调试日志
func (l *loggerImpl) Debug(msg string, fields ...Field) {
	l.log(DebugLevel, msg, fields...)
}

// Info 记录信息日志
func (l *loggerImpl) Info(msg string, fields ...Field) {
	l.log(InfoLevel, msg, fields...)
}

// Warn 记录警告日志
func (l *loggerImpl) Warn(msg string, fields ...Field) {
	l.log(WarnLevel, msg, fields...)
}

// Error 记录错误日志
func (l *loggerImpl) Error(msg string, fields ...Field) {
	l.log(ErrorLevel, msg, fields...)
}

// Fatal 记录致命日志
func (l *loggerImpl) Fatal(msg string, fields ...Field) {
	l.log(FatalLevel, msg, fields...)
	os.Exit(1)
}

// log 记录日志
func (l *loggerImpl) log(level Level, msg string, fields ...Field) {
	// 检查日志级别
	if level < l.config.Level {
		return
	}

	// 合并字段
	allFields := make([]Field, 0, len(l.fields)+len(fields))
	allFields = append(allFields, l.fields...)
	allFields = append(allFields, fields...)

	// 自动从上下文中提取traceId
	if l.context != nil {
		if traceId := l.context.Value("trace_id"); traceId != nil {
			if id, ok := traceId.(string); ok && id != "" {
				// 检查是否已经有traceId字段，避免重复
				hasTraceId := false
				for _, field := range allFields {
					if field.Key == "trace_id" {
						hasTraceId = true
						break
					}
				}
				// 如果没有traceId字段，则自动添加
				if !hasTraceId {
					allFields = append(allFields, Field{Key: "trace_id", Value: id})
				}
			}
		}
	}

	// 格式化日志
	var logLine string
	if l.config.Format == "json" {
		logLine = l.formatJSON(level, msg, allFields)
	} else {
		logLine = l.formatText(level, msg, allFields)
	}

	// 写入日志
	l.mu.Lock()
	defer l.mu.Unlock()

	fmt.Fprintln(l.output, logLine)
}

// formatJSON 格式化为JSON
func (l *loggerImpl) formatJSON(level Level, msg string, fields []Field) string {
	// 创建基础JSON结构
	jsonStr := fmt.Sprintf(`{"level":"%s","time":"%s","message":"%s"`,
		level.String(), time.Now().Format(time.RFC3339), msg)

	// 添加字段
	if len(fields) > 0 {
		jsonStr += `,"fields":{`
		for i, field := range fields {
			if i > 0 {
				jsonStr += ","
			}
			jsonStr += fmt.Sprintf(`"%s":"%v"`, field.Key, field.Value)
		}
		jsonStr += "}"
	}

	jsonStr += "}"
	return jsonStr
}

// formatText 格式化为文本
func (l *loggerImpl) formatText(level Level, msg string, fields []Field) string {
	// 基础文本格式
	textStr := fmt.Sprintf("[%s] %s %s",
		level.String(), time.Now().Format(time.RFC3339), msg)

	// 添加字段
	if len(fields) > 0 {
		textStr += " |"
		for _, field := range fields {
			textStr += fmt.Sprintf(" %s=%v", field.Key, field.Value)
		}
	}

	return textStr
}

// WithContext 添加上下文
func (l *loggerImpl) WithContext(ctx context.Context) Logger {
	newLogger := &loggerImpl{
		config:  l.config,
		output:  l.output,
		fields:  l.fields,
		context: ctx,
	}
	return newLogger
}

// WithFields 添加字段
func (l *loggerImpl) WithFields(fields ...Field) Logger {
	newLogger := &loggerImpl{
		config:  l.config,
		output:  l.output,
		fields:  append(l.fields, fields...),
		context: l.context,
	}
	return newLogger
}

// WithField 添加单个字段
func (l *loggerImpl) WithField(key string, value interface{}) Logger {
	return l.WithFields(NewField(key, value))
}

// SetLevel 设置日志级别
func (l *loggerImpl) SetLevel(level Level) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.config.Level = level
}

// GetLevel 获取日志级别
func (l *loggerImpl) GetLevel() Level {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.config.Level
}

// SetOutput 设置输出
func (l *loggerImpl) SetOutput(w io.Writer) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.output = w
}

// Close 关闭日志器
func (l *loggerImpl) Close() error {
	// 如果是文件输出，关闭文件
	if closer, ok := l.output.(io.Closer); ok {
		return closer.Close()
	}
	return nil
}
