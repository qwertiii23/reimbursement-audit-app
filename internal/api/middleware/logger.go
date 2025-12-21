package middleware

import (
	"context"
	"fmt"

	"github.com/gin-gonic/gin"
	"reimbursement-audit/internal/pkg/logger"
)

// 定义logger在Gin上下文中的键
const loggerKey = "logger"

// LoggerMiddleware 日志中间件，将带有traceId的logger注入到Gin上下文中
func LoggerMiddleware(baseLogger logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取traceId
		traceId := GetTraceId(c)

		// 创建带有traceId的上下文
		ctx := WithTraceId(context.Background(), traceId)

		// 创建带有上下文的logger
		loggerWithContext := baseLogger.WithContext(ctx)

		// 将logger注入到Gin上下文中
		c.Set(loggerKey, loggerWithContext)

		// 继续处理请求
		c.Next()
	}
}

// GetLogger 从Gin上下文中获取带有traceId的logger
func GetLogger(c *gin.Context) logger.Logger {
	if log, exists := c.Get(loggerKey); exists {
		if logger, ok := log.(logger.Logger); ok {
			return logger
		}
	}

	// 如果中间件未设置，返回一个默认logger（这种情况不应该发生）
	return nil
}

// LogInfo 从Gin上下文中获取logger并记录信息日志
func LogInfo(c *gin.Context, msg string, keyValuePairs ...interface{}) {
	if log := GetLogger(c); log != nil {
		fields := convertToFields(keyValuePairs...)
		log.Info(msg, fields...)
	}
}

// LogDebug 从Gin上下文中获取logger并记录调试日志
func LogDebug(c *gin.Context, msg string, keyValuePairs ...interface{}) {
	if log := GetLogger(c); log != nil {
		fields := convertToFields(keyValuePairs...)
		log.Debug(msg, fields...)
	}
}

// LogWarn 从Gin上下文中获取logger并记录警告日志
func LogWarn(c *gin.Context, msg string, keyValuePairs ...interface{}) {
	if log := GetLogger(c); log != nil {
		fields := convertToFields(keyValuePairs...)
		log.Warn(msg, fields...)
	}
}

// LogError 从Gin上下文中获取logger并记录错误日志
func LogError(c *gin.Context, msg string, keyValuePairs ...interface{}) {
	if log := GetLogger(c); log != nil {
		fields := convertToFields(keyValuePairs...)
		log.Error(msg, fields...)
	}
}

// LogFatal 从Gin上下文中获取logger并记录致命日志
func LogFatal(c *gin.Context, msg string, keyValuePairs ...interface{}) {
	if log := GetLogger(c); log != nil {
		fields := convertToFields(keyValuePairs...)
		log.Fatal(msg, fields...)
	}
}

// convertToFields 将可变参数转换为Field数组
// 支持两种格式：
// 1. "key1", value1, "key2", value2, ...
// 2. logger.Field{Key: "key1", Value: value1}, logger.Field{Key: "key2", Value: value2}, ...
func convertToFields(keyValuePairs ...interface{}) []logger.Field {
	var fields []logger.Field

	// 如果参数为空，返回空数组
	if len(keyValuePairs) == 0 {
		return fields
	}

	// 如果第一个参数是Field类型，假设所有参数都是Field类型
	if len(keyValuePairs) > 0 {
		if _, ok := keyValuePairs[0].(logger.Field); ok {
			for _, pair := range keyValuePairs {
				if field, ok := pair.(logger.Field); ok {
					fields = append(fields, field)
				}
			}
			return fields
		}
	}

	// 否则，假设参数是key-value对格式
	for i := 0; i < len(keyValuePairs); i += 2 {
		// 确保有足够的参数
		if i+1 < len(keyValuePairs) {
			// 尝试将key转换为字符串
			var key string
			if k, ok := keyValuePairs[i].(string); ok {
				key = k
			} else {
				// 如果不是字符串，使用fmt.Sprintf转换
				key = fmt.Sprintf("%v", keyValuePairs[i])
			}

			fields = append(fields, logger.Field{
				Key:   key,
				Value: keyValuePairs[i+1],
			})
		}
	}

	return fields
}
