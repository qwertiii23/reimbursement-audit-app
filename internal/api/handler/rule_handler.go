// rule_handler.go 处理规则管理的控制器
// 功能点：
// 1. 规则CRUD操作（新增、编辑、删除、查询）
// 2. 规则启用/禁用状态管理
// 3. 规则优先级设置
// 4. 规则分类管理（金额校验、频次校验、发票信息校验等）
// 5. 规则导入/导出
// 6. 规则测试和验证

package handler

import (
	"context"
	"reimbursement-audit/internal/api/middleware"
	"reimbursement-audit/internal/api/request"
	"reimbursement-audit/internal/api/response"
	"reimbursement-audit/internal/domain/rule"
	"strconv"

	"github.com/gin-gonic/gin"
)

// RuleHandler 处理规则管理请求的结构体
type RuleHandler struct {
	ruleService *rule.RuleService
}

// NewRuleHandler 创建规则管理处理器实例
func NewRuleHandler(ruleService *rule.RuleService) *RuleHandler {
	return &RuleHandler{
		ruleService: ruleService,
	}
}

// CreateRule 创建新规则
func (h *RuleHandler) CreateRule(c *gin.Context) {
	middleware.LogInfo(c, "创建新规则请求", "path", c.Request.URL.Path,
		"method", c.Request.Method, "body", c.Request.Body, "remote_addr", c.ClientIP())
	// 获取traceId
	traceId := middleware.GetTraceId(c)
	// 创建上下文，用于数据库操作
	ctx := middleware.WithTraceId(context.Background(), traceId)
	var req request.CreateRuleRequest
	if err := c.ShouldBind(&req); err != nil {
		middleware.LogError(c, "JSON数据绑定失败", "error", err.Error(), "context", ctx)
		response.ErrorResponse(c, response.CodeInvalidParams, err.Error())
		return
	}
	rule, err := h.ruleService.CreateRule(ctx, &req)
	if err != nil {
		middleware.LogError(c, "创建规则失败", "error", err.Error(), "context", ctx)
		response.ErrorResponse(c, response.CodeInternalError, err.Error())
		return
	}
	middleware.LogInfo(c, "创建新规则成功", "rule_id", rule.ID, "context", ctx)
	response.SuccessResponse(c, "规则创建成功")
}

// UpdateRule 更新规则
func (h *RuleHandler) UpdateRule(c *gin.Context) {
	middleware.LogInfo(c, "更新规则请求", "path", c.Request.URL.Path,
		"method", c.Request.Method, "remote_addr", c.ClientIP())
	traceId := middleware.GetTraceId(c)
	ctx := middleware.WithTraceId(context.Background(), traceId)

	var req request.UpdateRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.LogError(c, "JSON数据绑定失败", "error", err.Error(), "context", ctx)
		response.ErrorResponse(c, response.CodeInvalidParams, err.Error())
		return
	}

	rule, err := h.ruleService.UpdateRule(ctx, &req)
	if err != nil {
		middleware.LogError(c, "更新规则失败", "error", err.Error(), "context", ctx)
		response.ErrorResponse(c, response.CodeInternalError, err.Error())
		return
	}

	middleware.LogInfo(c, "更新规则成功", "rule_id", rule.ID, "context", ctx)
	response.SuccessResponse(c, rule)
}

// DeleteRule 删除规则
func (h *RuleHandler) DeleteRule(c *gin.Context) {
	middleware.LogInfo(c, "删除规则请求", "path", c.Request.URL.Path,
		"method", c.Request.Method, "remote_addr", c.ClientIP())
	traceId := middleware.GetTraceId(c)
	ctx := middleware.WithTraceId(context.Background(), traceId)

	ruleID := c.Param("id")
	if ruleID == "" {
		middleware.LogError(c, "缺少规则ID", "context", ctx)
		response.ErrorResponse(c, response.CodeInvalidParams, "缺少规则ID")
		return
	}

	if err := h.ruleService.DeleteRule(ctx, ruleID); err != nil {
		middleware.LogError(c, "删除规则失败", "error", err.Error(), "context", ctx)
		response.ErrorResponse(c, response.CodeInternalError, err.Error())
		return
	}

	middleware.LogInfo(c, "删除规则成功", "rule_id", ruleID, "context", ctx)
	response.SuccessResponse(c, "规则删除成功")
}

// GetRules 获取规则列表
func (h *RuleHandler) GetRules(c *gin.Context) {
	middleware.LogInfo(c, "获取规则列表请求", "path", c.Request.URL.Path,
		"method", c.Request.Method, "remote_addr", c.ClientIP())
	traceId := middleware.GetTraceId(c)
	ctx := middleware.WithTraceId(context.Background(), traceId)

	filter := &rule.RuleFilter{
		RuleCode: c.Query("rule_code"),
		Type:     c.Query("type"),
		Category: c.Query("category"),
		Status:   c.Query("status"),
		Page:     1,
		Size:     10,
	}

	if page := c.Query("page"); page != "" {
		if p, err := strconv.Atoi(page); err == nil {
			filter.Page = p
		}
	}

	if size := c.Query("size"); size != "" {
		if s, err := strconv.Atoi(size); err == nil {
			filter.Size = s
		}
	}

	rules, total, err := h.ruleService.GetRules(ctx, filter)
	if err != nil {
		middleware.LogError(c, "获取规则列表失败", "error", err.Error(), "context", ctx)
		response.ErrorResponse(c, response.CodeInternalError, err.Error())
		return
	}

	middleware.LogInfo(c, "获取规则列表成功", "total", total, "count", len(rules), "context", ctx)
	response.SuccessResponse(c, gin.H{
		"rules": rules,
		"total": total,
	})
}

// EnableRule 启用规则
func (h *RuleHandler) EnableRule(c *gin.Context) {
	middleware.LogInfo(c, "启用规则请求", "path", c.Request.URL.Path,
		"method", c.Request.Method, "remote_addr", c.ClientIP())
	traceId := middleware.GetTraceId(c)
	ctx := middleware.WithTraceId(context.Background(), traceId)

	ruleID := c.Param("id")
	if ruleID == "" {
		middleware.LogError(c, "缺少规则ID", "context", ctx)
		response.ErrorResponse(c, response.CodeInvalidParams, "缺少规则ID")
		return
	}

	if err := h.ruleService.EnableRule(ctx, ruleID); err != nil {
		middleware.LogError(c, "启用规则失败", "error", err.Error(), "context", ctx)
		response.ErrorResponse(c, response.CodeInternalError, err.Error())
		return
	}

	middleware.LogInfo(c, "启用规则成功", "rule_id", ruleID, "context", ctx)
	response.SuccessResponse(c, "规则启用成功")
}

// DisableRule 禁用规则
func (h *RuleHandler) DisableRule(c *gin.Context) {
	middleware.LogInfo(c, "禁用规则请求", "path", c.Request.URL.Path,
		"method", c.Request.Method, "remote_addr", c.ClientIP())
	traceId := middleware.GetTraceId(c)
	ctx := middleware.WithTraceId(context.Background(), traceId)

	ruleID := c.Param("id")
	if ruleID == "" {
		middleware.LogError(c, "缺少规则ID", "context", ctx)
		response.ErrorResponse(c, response.CodeInvalidParams, "缺少规则ID")
		return
	}

	if err := h.ruleService.DisableRule(ctx, ruleID); err != nil {
		middleware.LogError(c, "禁用规则失败", "error", err.Error(), "context", ctx)
		response.ErrorResponse(c, response.CodeInternalError, err.Error())
		return
	}

	middleware.LogInfo(c, "禁用规则成功", "rule_id", ruleID, "context", ctx)
	response.SuccessResponse(c, "规则禁用成功")
}

// TestRule 测试规则
func (h *RuleHandler) TestRule(c *gin.Context) {
	middleware.LogInfo(c, "测试规则请求", "path", c.Request.URL.Path,
		"method", c.Request.Method, "remote_addr", c.ClientIP())
	traceId := middleware.GetTraceId(c)
	ctx := middleware.WithTraceId(context.Background(), traceId)

	ruleID := c.Param("id")
	if ruleID == "" {
		middleware.LogError(c, "缺少规则ID", "context", ctx)
		response.ErrorResponse(c, response.CodeInvalidParams, "缺少规则ID")
		return
	}

	response.ErrorResponse(c, response.CodeInternalError, "测试规则功能暂未实现")
}
