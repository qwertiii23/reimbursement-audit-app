package handler

import (
	"context"
	"reimbursement-audit/internal/api/middleware"
	"reimbursement-audit/internal/api/response"
	ruleenginedomain "reimbursement-audit/internal/domain/ruleengine"
	"strconv"

	"github.com/gin-gonic/gin"
)

type RuleEngineHandler struct {
	service *ruleenginedomain.RuleEngineService
}

func NewRuleEngineHandler(service *ruleenginedomain.RuleEngineService) *RuleEngineHandler {
	return &RuleEngineHandler{
		service: service,
	}
}

func (h *RuleEngineHandler) CreateRule(c *gin.Context) {
	ctx := context.Background()

	var req ruleenginedomain.CreateRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.LogError(c, "请求参数错误", "error", err.Error())
		response.ErrorResponse(c, response.CodeInvalidParams, "请求参数错误")
		return
	}

	rule, err := h.service.CreateRule(ctx, &req)
	if err != nil {
		middleware.LogError(c, "创建规则失败", "error", err.Error())
		response.ErrorResponse(c, response.CodeInternalError, "创建规则失败")
		return
	}

	response.SuccessResponse(c, gin.H{
		"rule": rule,
	})
}

func (h *RuleEngineHandler) UpdateRule(c *gin.Context) {
	ctx := context.Background()

	ruleId := c.Param("id")
	if ruleId == "" {
		response.ErrorResponse(c, response.CodeInvalidParams, "规则ID不能为空")
		return
	}

	var req ruleenginedomain.CreateRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.LogError(c, "请求参数错误", "error", err.Error())
		response.ErrorResponse(c, response.CodeInvalidParams, "请求参数错误")
		return
	}

	updateReq := &ruleenginedomain.UpdateRuleRequest{
		ID:          ruleId,
		Name:        req.Name,
		Description: req.Description,
		Priority:    req.Priority,
		Enabled:     req.Enabled,
		Conditions:  req.Conditions,
		Decision:    req.Decision,
	}

	rule, err := h.service.UpdateRule(ctx, updateReq)
	if err != nil {
		middleware.LogError(c, "更新规则失败", "error", err.Error())
		response.ErrorResponse(c, response.CodeInternalError, "更新规则失败")
		return
	}

	response.SuccessResponse(c, gin.H{
		"rule": rule,
	})
}

func (h *RuleEngineHandler) DeleteRule(c *gin.Context) {
	ctx := context.Background()

	ruleId := c.Param("id")
	if ruleId == "" {
		response.ErrorResponse(c, response.CodeInvalidParams, "规则ID不能为空")
		return
	}

	err := h.service.DeleteRule(ctx, ruleId)
	if err != nil {
		middleware.LogError(c, "删除规则失败", "error", err.Error())
		response.ErrorResponse(c, response.CodeInternalError, "删除规则失败")
		return
	}

	response.SuccessResponse(c, gin.H{
		"message": "删除成功",
	})
}

func (h *RuleEngineHandler) ToggleRuleStatus(c *gin.Context) {
	ctx := context.Background()

	ruleId := c.Param("id")
	if ruleId == "" {
		response.ErrorResponse(c, response.CodeInvalidParams, "规则ID不能为空")
		return
	}

	rule, err := h.service.GetRuleByID(ctx, ruleId)
	if err != nil {
		middleware.LogError(c, "获取规则失败", "error", err.Error())
		response.ErrorResponse(c, response.CodeInternalError, "获取规则失败")
		return
	}

	if rule.Enabled {
		err = h.service.DisableRule(ctx, ruleId)
	} else {
		err = h.service.EnableRule(ctx, ruleId)
	}

	if err != nil {
		middleware.LogError(c, "切换规则状态失败", "error", err.Error())
		response.ErrorResponse(c, response.CodeInternalError, "切换规则状态失败")
		return
	}

	response.SuccessResponse(c, gin.H{
		"message": "状态切换成功",
	})
}

func (h *RuleEngineHandler) GetRules(c *gin.Context) {
	ctx := context.Background()

	name := c.Query("name")
	enabledStr := c.Query("enabled")

	var enabled *bool
	if enabledStr != "" {
		enabledVal := enabledStr == "true"
		enabled = &enabledVal
	}

	page := 1
	size := 10
	if pageStr := c.Query("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}
	if sizeStr := c.Query("size"); sizeStr != "" {
		if s, err := strconv.Atoi(sizeStr); err == nil && s > 0 {
			size = s
		}
	}

	filter := &ruleenginedomain.RuleFilter{
		Name:    name,
		Enabled: enabled,
		Page:    page,
		Size:    size,
	}

	rules, total, err := h.service.GetRules(ctx, filter)
	if err != nil {
		middleware.LogError(c, "获取规则列表失败", "error", err.Error())
		response.ErrorResponse(c, response.CodeInternalError, "获取规则列表失败")
		return
	}

	response.SuccessResponse(c, gin.H{
		"rules": h.toRuleResponses(rules),
		"total": total,
		"page":  page,
		"size":  size,
	})
}

func (h *RuleEngineHandler) GetRuleByID(c *gin.Context) {
	ctx := context.Background()

	ruleId := c.Param("id")
	if ruleId == "" {
		response.ErrorResponse(c, response.CodeInvalidParams, "规则ID不能为空")
		return
	}

	rule, err := h.service.GetRuleByID(ctx, ruleId)
	if err != nil {
		middleware.LogError(c, "获取规则详情失败", "error", err.Error())
		response.ErrorResponse(c, response.CodeInternalError, "获取规则详情失败")
		return
	}

	response.SuccessResponse(c, gin.H{
		"rule": h.toRuleResponse(rule),
	})
}

func (h *RuleEngineHandler) GetFeatures(c *gin.Context) {
	ctx := context.Background()

	category := c.Query("category")
	enabledStr := c.Query("enabled")

	var enabled *bool
	if enabledStr != "" {
		enabledVal := enabledStr == "true"
		enabled = &enabledVal
	}

	page := 1
	size := 100
	if pageStr := c.Query("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}
	if sizeStr := c.Query("size"); sizeStr != "" {
		if s, err := strconv.Atoi(sizeStr); err == nil && s > 0 {
			size = s
		}
	}

	filter := &ruleenginedomain.FeatureFilter{
		Category: category,
		Enabled:  enabled,
		Page:     page,
		Size:     size,
	}

	features, total, err := h.service.GetFeatures(ctx, filter)
	if err != nil {
		middleware.LogError(c, "获取特征列表失败", "error", err.Error())
		response.ErrorResponse(c, response.CodeInternalError, "获取特征列表失败")
		return
	}

	response.SuccessResponse(c, gin.H{
		"features": h.toFeatureResponses(features),
		"total":    total,
		"page":     page,
		"size":     size,
	})
}

func (h *RuleEngineHandler) toRuleResponse(rule *ruleenginedomain.Rule) *ruleenginedomain.RuleResponse {
	if rule == nil {
		return nil
	}

	return &ruleenginedomain.RuleResponse{
		ID:          rule.ID,
		Name:        rule.Name,
		Description: rule.Description,
		Conditions:  h.toConditionResponses(rule.Conditions),
		Decision:    *h.toDecisionResponse(&rule.Decision),
		Priority:    rule.Priority,
		Enabled:     rule.Enabled,
		CreatedAt:   rule.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:   rule.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
}

func (h *RuleEngineHandler) toRuleResponses(rules []*ruleenginedomain.Rule) []*ruleenginedomain.RuleResponse {
	responses := make([]*ruleenginedomain.RuleResponse, 0, len(rules))
	for _, rule := range rules {
		responses = append(responses, h.toRuleResponse(rule))
	}
	return responses
}

func (h *RuleEngineHandler) toConditionResponse(condition *ruleenginedomain.Condition) *ruleenginedomain.ConditionResponse {
	if condition == nil {
		return nil
	}

	return &ruleenginedomain.ConditionResponse{
		ID:        condition.ID,
		FeatureID: condition.FeatureID,
		Operator:  condition.Operator,
		Value:     condition.Value,
		LogicOp:   condition.LogicOp,
	}
}

func (h *RuleEngineHandler) toConditionResponses(conditions []ruleenginedomain.Condition) []*ruleenginedomain.ConditionResponse {
	responses := make([]*ruleenginedomain.ConditionResponse, 0, len(conditions))
	for i := range conditions {
		responses = append(responses, h.toConditionResponse(&conditions[i]))
	}
	return responses
}

func (h *RuleEngineHandler) toDecisionResponse(decision *ruleenginedomain.Decision) *ruleenginedomain.DecisionResponse {
	if decision == nil {
		return nil
	}

	return &ruleenginedomain.DecisionResponse{
		ID:        decision.ID,
		Type:      decision.Type,
		Reason:    decision.Reason,
		CreatedAt: decision.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt: decision.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
}

func (h *RuleEngineHandler) toFeatureResponse(feature *ruleenginedomain.Feature) *ruleenginedomain.FeatureResponse {
	if feature == nil {
		return nil
	}

	return &ruleenginedomain.FeatureResponse{
		ID:             feature.ID,
		Name:           feature.Name,
		Code:           feature.Code,
		Description:    feature.Description,
		Type:           feature.Type,
		ValueType:      feature.ValueType,
		Category:       feature.Category,
		Enabled:        feature.Enabled,
		FunctionName:   feature.FunctionName,
		FunctionConfig: feature.FunctionConfig,
		Values:         h.toFeatureValueResponses(feature.Values),
		CreatedAt:      feature.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:      feature.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
}

func (h *RuleEngineHandler) toFeatureResponses(features []*ruleenginedomain.Feature) []*ruleenginedomain.FeatureResponse {
	responses := make([]*ruleenginedomain.FeatureResponse, 0, len(features))
	for _, feature := range features {
		responses = append(responses, h.toFeatureResponse(feature))
	}
	return responses
}

func (h *RuleEngineHandler) toFeatureValueResponse(value *ruleenginedomain.FeatureValue) *ruleenginedomain.FeatureValueResponse {
	if value == nil {
		return nil
	}

	return &ruleenginedomain.FeatureValueResponse{
		ID:        value.ID,
		FeatureID: value.FeatureID,
		Value:     value.Value,
		Label:     value.Label,
		SortOrder: value.SortOrder,
		Enabled:   value.Enabled,
	}
}

func (h *RuleEngineHandler) toFeatureValueResponses(values []ruleenginedomain.FeatureValue) []*ruleenginedomain.FeatureValueResponse {
	responses := make([]*ruleenginedomain.FeatureValueResponse, 0, len(values))
	for i := range values {
		responses = append(responses, h.toFeatureValueResponse(&values[i]))
	}
	return responses
}

func (h *RuleEngineHandler) TestRules(c *gin.Context) {
	ctx := context.Background()

	var testData map[string]interface{}
	if err := c.ShouldBindJSON(&testData); err != nil {
		middleware.LogError(c, "请求参数错误", "error", err.Error())
		response.ErrorResponse(c, response.CodeInvalidParams, "请求参数错误")
		return
	}

	results, err := h.service.ExecuteAllRules(ctx, testData)
	if err != nil {
		middleware.LogError(c, "执行规则测试失败", "error", err.Error())
		response.ErrorResponse(c, response.CodeInternalError, "执行规则测试失败")
		return
	}

	response.SuccessResponse(c, gin.H{
		"results": results,
		"total":   len(results),
	})
}
