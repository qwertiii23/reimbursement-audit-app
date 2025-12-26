package response

import (
	"net/http"

	"reimbursement-audit/internal/api/middleware"

	"github.com/gin-gonic/gin"
)

// JSONResponse 返回JSON响应的辅助函数
func JSONResponse(c *gin.Context, code int, message string, data interface{}) {
	traceId := middleware.GetTraceId(c)

	responseData := gin.H{
		"code":    code,
		"message": message,
		"data":    data,
	}

	if traceId != "" {
		responseData["trace_id"] = traceId
	}

	c.JSON(http.StatusOK, responseData)
}

// ErrorResponse 返回错误响应的辅助函数
func ErrorResponse(c *gin.Context, code int, message string) {
	JSONResponse(c, code, message, nil)
}

// SuccessResponse 返回成功响应的辅助函数
func SuccessResponse(c *gin.Context, data interface{}) {
	JSONResponse(c, CodeSuccess, "成功", data)
}
