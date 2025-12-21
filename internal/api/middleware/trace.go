package middleware

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// TraceIdKey 上下文中存储traceId的键
const TraceIdKey = "trace_id"

// TraceMiddleware 生成traceId并添加到上下文中的中间件
func TraceMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 生成traceId
		traceId := uuid.New().String()

		// 将traceId添加到请求上下文
		c.Set(TraceIdKey, traceId)

		// 将traceId添加到响应头
		c.Header("X-Trace-Id", traceId)

		// 继续处理请求
		c.Next()
	}
}

// GetTraceId 从上下文中获取traceId
func GetTraceId(c *gin.Context) string {
	if traceId, exists := c.Get(TraceIdKey); exists {
		if id, ok := traceId.(string); ok {
			return id
		}
	}
	return ""
}

// GetTraceIdFromContext 从context.Context中获取traceId
func GetTraceIdFromContext(ctx context.Context) string {
	if traceId := ctx.Value(TraceIdKey); traceId != nil {
		if id, ok := traceId.(string); ok {
			return id
		}
	}
	return ""
}

// WithTraceId 将traceId添加到context.Context中
func WithTraceId(ctx context.Context, traceId string) context.Context {
	return context.WithValue(ctx, TraceIdKey, traceId)
}
