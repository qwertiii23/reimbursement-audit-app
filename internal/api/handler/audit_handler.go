// audit_handler.go 处理审核触发的控制器
// 功能点：
// 1. 接收审核请求（报销单ID）
// 2. 触发规则引擎校验（刚性规则）
// 3. 触发RAG检索和大模型分析（柔性问题）
// 4. 整合审核结果并生成审核报告
// 5. 返回审核状态和结果
// 6. 处理审核过程中的异常情况

package handler

import (
	"context"
	"reimbursement-audit/internal/api/middleware"
	"reimbursement-audit/internal/api/request"
	"reimbursement-audit/internal/api/response"
	"reimbursement-audit/internal/application/service"

	"github.com/gin-gonic/gin"
)

// AuditHandler 处理审核请求的结构体
type AuditHandler struct {
	auditService *service.AuditApplicationService
}

// NewAuditHandler 创建审核处理器实例
func NewAuditHandler(auditService *service.AuditApplicationService) *AuditHandler {
	return &AuditHandler{
		auditService: auditService,
	}
}

// StartAudit 触发报销单审核
func (h *AuditHandler) StartAudit(c *gin.Context) {
	middleware.LogInfo(c, "开始审核请求", "path", c.Request.URL.Path,
		"method", c.Request.Method, "remote_addr", c.ClientIP())
	traceId := middleware.GetTraceId(c)
	ctx := middleware.WithTraceId(context.Background(), traceId)

	var req request.StartAuditRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.LogError(c, "JSON数据绑定失败", "error", err.Error(), "context", ctx)
		response.ErrorResponse(c, response.CodeInvalidParams, err.Error())
		return
	}

	if err := req.Validate(); err != nil {
		middleware.LogError(c, "请求参数校验失败", "error", err.Error(), "context", ctx)
		response.ErrorResponse(c, response.CodeInvalidParams, err.Error())
		return
	}

	auditResponse, err := h.auditService.StartAudit(ctx, &req)
	if err != nil {
		middleware.LogError(c, "开始审核失败", "error", err.Error(), "context", ctx)
		response.ErrorResponse(c, response.CodeInternalError, err.Error())
		return
	}

	middleware.LogInfo(c, "开始审核成功", "audit_id", auditResponse.ID, "context", ctx)
	response.SuccessResponse(c, auditResponse)
}

// GetAuditStatus 获取审核状态
func (h *AuditHandler) GetAuditStatus(c *gin.Context) {
	middleware.LogInfo(c, "获取审核状态请求", "path", c.Request.URL.Path,
		"method", c.Request.Method, "remote_addr", c.ClientIP())
	traceId := middleware.GetTraceId(c)
	ctx := middleware.WithTraceId(context.Background(), traceId)

	auditID := c.Param("id")
	if auditID == "" {
		middleware.LogError(c, "缺少审核ID", "context", ctx)
		response.ErrorResponse(c, response.CodeInvalidParams, "缺少审核ID")
		return
	}

	statusResponse, err := h.auditService.GetAuditStatus(ctx, auditID)
	if err != nil {
		middleware.LogError(c, "获取审核状态失败", "error", err.Error(), "context", ctx)
		response.ErrorResponse(c, response.CodeInternalError, err.Error())
		return
	}

	middleware.LogInfo(c, "获取审核状态成功", "audit_id", auditID, "context", ctx)
	response.SuccessResponse(c, statusResponse)
}

// GetAuditResult 获取审核结果
func (h *AuditHandler) GetAuditResult(c *gin.Context) {
	middleware.LogInfo(c, "获取审核结果请求", "path", c.Request.URL.Path,
		"method", c.Request.Method, "remote_addr", c.ClientIP())
	traceId := middleware.GetTraceId(c)
	ctx := middleware.WithTraceId(context.Background(), traceId)

	auditID := c.Param("id")
	if auditID == "" {
		middleware.LogError(c, "缺少审核ID", "context", ctx)
		response.ErrorResponse(c, response.CodeInvalidParams, "缺少审核ID")
		return
	}

	resultResponse, err := h.auditService.GetAuditResult(ctx, auditID)
	if err != nil {
		middleware.LogError(c, "获取审核结果失败", "error", err.Error(), "context", ctx)
		response.ErrorResponse(c, response.CodeInternalError, err.Error())
		return
	}

	middleware.LogInfo(c, "获取审核结果成功", "audit_id", auditID, "context", ctx)
	response.SuccessResponse(c, resultResponse)
}

// RetryAudit 重试审核
func (h *AuditHandler) RetryAudit(c *gin.Context) {
	middleware.LogInfo(c, "重试审核请求", "path", c.Request.URL.Path,
		"method", c.Request.Method, "remote_addr", c.ClientIP())
	traceId := middleware.GetTraceId(c)
	ctx := middleware.WithTraceId(context.Background(), traceId)

	auditID := c.Param("id")
	if auditID == "" {
		middleware.LogError(c, "缺少审核ID", "context", ctx)
		response.ErrorResponse(c, response.CodeInvalidParams, "缺少审核ID")
		return
	}

	resultResponse, err := h.auditService.RetryAudit(ctx, auditID)
	if err != nil {
		middleware.LogError(c, "重试审核失败", "error", err.Error(), "context", ctx)
		response.ErrorResponse(c, response.CodeInternalError, err.Error())
		return
	}

	middleware.LogInfo(c, "重试审核成功", "audit_id", auditID, "context", ctx)
	response.SuccessResponse(c, resultResponse)
}