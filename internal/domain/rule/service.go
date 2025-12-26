// service.go 规则校验逻辑
// 功能点：
// 1. 规则校验流程编排
// 2. 规则优先级排序和执行
// 3. 规则冲突处理
// 4. 规则校验结果整合
// 5. 规则动态加载和更新
// 6. 规则测试和验证

package rule

import (
	"context"
	"errors"
	"fmt"
	"time"

	"reimbursement-audit/internal/api/request"
	"reimbursement-audit/internal/pkg/logger"

	"github.com/google/uuid"
)

// RuleService 规则服务结构体
type RuleService struct {
	repo   Repository
	logger logger.Logger
	engine *GRuleEngine
}

// NewRuleService 创建规则服务实例
func NewRuleService(repo Repository, logger logger.Logger, engine *GRuleEngine) *RuleService {
	return &RuleService{
		repo:   repo,
		logger: logger,
		engine: engine,
	}
}

// generateRuleCode 生成规则编码
// 格式: RULE_YYYYMMDD_HHMMSS_UUID
func (s *RuleService) generateRuleCode() string {
	now := time.Now()
	timeStr := now.Format("20060102_150405")
	uuidStr := uuid.New().String()[:8] // 取UUID前8位
	return fmt.Sprintf("RULE_%s_%s", timeStr, uuidStr)
}

// CreateRule 创建规则
func (s *RuleService) CreateRule(ctx context.Context, req *request.CreateRuleRequest) (*Rule, error) {
	// 参数验证
	if req.Name == "" {
		s.logger.WithContext(ctx).Error("规则名称不能为空")
		return nil, errors.New("规则名称不能为空")
	}
	if req.Type == "" {
		s.logger.WithContext(ctx).Error("规则类型不能为空")
		return nil, errors.New("规则类型不能为空")
	}

	// 生成规则编码，最多重试3次
	var ruleCode string
	var exists bool
	var err error

	for i := 0; i < 3; i++ {
		ruleCode = s.generateRuleCode()
		exists, err = s.repo.CheckRuleCodeExists(ctx, ruleCode, "")
		if err != nil {
			s.logger.WithContext(ctx).Error("检查规则编码唯一性失败",
				logger.NewField("error", err.Error()),
				logger.NewField("rule_code", ruleCode))
			return nil, err
		}
		if !exists {
			break // 找到未使用的规则编码
		}
	}

	if exists {
		s.logger.WithContext(ctx).Error("生成唯一规则编码失败，已重试3次")
		return nil, errors.New("生成唯一规则编码失败")
	}

	// 创建规则模型
	now := time.Now()
	rule := &Rule{
		ID:          uuid.New().String(),
		RuleCode:    ruleCode,
		Name:        req.Name,
		Description: req.Description,
		Type:        req.Type,
		Category:    req.Category,
		Status:      RuleStatusDraft, // 默认状态为草稿
		Definition:  req.Definition,
		Priority:    req.Priority,
		Enabled:     false, // 默认禁用
		CreatedBy:   req.CreatedBy,
		UpdatedAt:   now,
		CreatedAt:   now,
		Version:     1,
	}

	// 保存规则
	if err := s.repo.CreateRule(ctx, rule); err != nil {
		s.logger.WithContext(ctx).Error("创建规则失败",
			logger.NewField("error", err.Error()),
			logger.NewField("rule_code", ruleCode))
		return nil, err
	}

	s.logger.WithContext(ctx).Info("创建规则成功",
		logger.NewField("rule_id", rule.ID),
		logger.NewField("rule_code", rule.RuleCode))

	return rule, nil
}

// UpdateRule 更新规则
func (s *RuleService) UpdateRule(ctx context.Context, req *request.UpdateRuleRequest) (*Rule, error) {
	// 参数验证
	if req.ID == "" {
		s.logger.WithContext(ctx).Error("规则ID不能为空")
		return nil, errors.New("规则ID不能为空")
	}

	// 获取现有规则
	existingRule, err := s.repo.GetRuleByID(ctx, req.ID)
	if err != nil {
		s.logger.WithContext(ctx).Error("获取规则失败",
			logger.NewField("error", err.Error()),
			logger.NewField("rule_id", req.ID))
		return nil, err
	}

	// 处理规则编码
	var newRuleCode string
	if req.RuleCode == "" {
		// 如果没有提供规则编码，生成一个新的
		var exists bool
		for i := 0; i < 3; i++ {
			newRuleCode = s.generateRuleCode()
			exists, err = s.repo.CheckRuleCodeExists(ctx, newRuleCode, req.ID)
			if err != nil {
				s.logger.WithContext(ctx).Error("检查规则编码唯一性失败",
					logger.NewField("error", err.Error()),
					logger.NewField("rule_code", newRuleCode))
				return nil, err
			}
			if !exists {
				break // 找到未使用的规则编码
			}
		}

		if exists {
			s.logger.WithContext(ctx).Error("生成唯一规则编码失败，已重试3次")
			return nil, errors.New("生成唯一规则编码失败")
		}
	} else {
		// 如果提供了规则编码，检查是否已被其他规则使用
		if req.RuleCode != existingRule.RuleCode {
			exists, err := s.repo.CheckRuleCodeExists(ctx, req.RuleCode, req.ID)
			if err != nil {
				s.logger.WithContext(ctx).Error("检查规则编码唯一性失败",
					logger.NewField("error", err.Error()),
					logger.NewField("rule_code", req.RuleCode))
				return nil, err
			}
			if exists {
				s.logger.WithContext(ctx).Warn("规则编码已存在",
					logger.NewField("rule_code", req.RuleCode))
				return nil, errors.New("规则编码已存在")
			}
		}
		newRuleCode = req.RuleCode
	}

	// 更新规则字段
	existingRule.RuleCode = newRuleCode
	existingRule.Name = req.Name
	existingRule.Description = req.Description
	existingRule.Type = req.Type
	existingRule.Category = req.Category
	existingRule.Status = req.Status
	existingRule.Definition = req.Definition
	existingRule.Priority = req.Priority
	existingRule.UpdatedBy = req.UpdatedBy
	existingRule.Version = existingRule.Version + 1

	// 保存更新
	if err := s.repo.UpdateRule(ctx, existingRule); err != nil {
		s.logger.WithContext(ctx).Error("更新规则失败",
			logger.NewField("error", err.Error()),
			logger.NewField("rule_id", req.ID))
		return nil, err
	}

	s.logger.WithContext(ctx).Info("更新规则成功",
		logger.NewField("rule_id", existingRule.ID),
		logger.NewField("rule_code", existingRule.RuleCode))

	return existingRule, nil
}

// GetRules 获取规则列表
func (s *RuleService) GetRules(ctx context.Context, filter *RuleFilter) ([]*Rule, int64, error) {
	// 设置默认分页参数
	if filter != nil {
		if filter.Page <= 0 {
			filter.Page = 1
		}
		if filter.Size <= 0 {
			filter.Size = 10
		}
	} else {
		filter = &RuleFilter{
			Page: 1,
			Size: 10,
		}
	}

	// 查询规则列表
	rules, total, err := s.repo.ListRules(ctx, filter)
	if err != nil {
		s.logger.WithContext(ctx).Error("查询规则列表失败",
			logger.NewField("error", err.Error()))
		return nil, 0, err
	}

	s.logger.WithContext(ctx).Info("查询规则列表成功",
		logger.NewField("total", total),
		logger.NewField("count", len(rules)))

	return rules, total, nil
}

// GetRuleByID 根据ID获取规则
func (s *RuleService) GetRuleByID(ctx context.Context, id string) (*Rule, error) {
	if id == "" {
		s.logger.WithContext(ctx).Error("规则ID不能为空")
		return nil, errors.New("规则ID不能为空")
	}

	rule, err := s.repo.GetRuleByID(ctx, id)
	if err != nil {
		s.logger.WithContext(ctx).Error("获取规则失败",
			logger.NewField("error", err.Error()),
			logger.NewField("rule_id", id))
		return nil, err
	}

	return rule, nil
}

// GetRuleByCode 根据规则编码获取规则
func (s *RuleService) GetRuleByCode(ctx context.Context, ruleCode string) (*Rule, error) {
	if ruleCode == "" {
		s.logger.WithContext(ctx).Error("规则编码不能为空")
		return nil, errors.New("规则编码不能为空")
	}

	rule, err := s.repo.GetRuleByCode(ctx, ruleCode)
	if err != nil {
		s.logger.WithContext(ctx).Error("获取规则失败",
			logger.NewField("error", err.Error()),
			logger.NewField("rule_code", ruleCode))
		return nil, err
	}

	return rule, nil
}

// DeleteRule 删除规则
func (s *RuleService) DeleteRule(ctx context.Context, id string) error {
	if id == "" {
		s.logger.WithContext(ctx).Error("规则ID不能为空")
		return errors.New("规则ID不能为空")
	}

	// 检查规则是否存在
	_, err := s.repo.GetRuleByID(ctx, id)
	if err != nil {
		s.logger.WithContext(ctx).Error("规则不存在",
			logger.NewField("error", err.Error()),
			logger.NewField("rule_id", id))
		return err
	}

	// 删除规则
	if err := s.repo.DeleteRule(ctx, id); err != nil {
		s.logger.WithContext(ctx).Error("删除规则失败",
			logger.NewField("error", err.Error()),
			logger.NewField("rule_id", id))
		return err
	}

	s.logger.WithContext(ctx).Info("删除规则成功",
		logger.NewField("rule_id", id))

	return nil
}

// EnableRule 启用规则
func (s *RuleService) EnableRule(ctx context.Context, id string) error {
	if id == "" {
		s.logger.WithContext(ctx).Error("规则ID不能为空")
		return errors.New("规则ID不能为空")
	}

	// 检查规则是否存在
	rule, err := s.repo.GetRuleByID(ctx, id)
	if err != nil {
		s.logger.WithContext(ctx).Error("规则不存在",
			logger.NewField("error", err.Error()),
			logger.NewField("rule_id", id))
		return err
	}

	// 检查规则是否已启用
	if rule.Enabled {
		s.logger.WithContext(ctx).Warn("规则已启用",
			logger.NewField("rule_id", id))
		return nil
	}

	// 启用规则
	if err := s.repo.EnableRule(ctx, id); err != nil {
		s.logger.WithContext(ctx).Error("启用规则失败",
			logger.NewField("error", err.Error()),
			logger.NewField("rule_id", id))
		return err
	}

	s.logger.WithContext(ctx).Info("启用规则成功",
		logger.NewField("rule_id", id))

	return nil
}

// DisableRule 禁用规则
func (s *RuleService) DisableRule(ctx context.Context, id string) error {
	if id == "" {
		s.logger.WithContext(ctx).Error("规则ID不能为空")
		return errors.New("规则ID不能为空")
	}

	// 检查规则是否存在
	rule, err := s.repo.GetRuleByID(ctx, id)
	if err != nil {
		s.logger.WithContext(ctx).Error("规则不存在",
			logger.NewField("error", err.Error()),
			logger.NewField("rule_id", id))
		return err
	}

	// 检查规则是否已禁用
	if !rule.Enabled {
		s.logger.WithContext(ctx).Warn("规则已禁用",
			logger.NewField("rule_id", id))
		return nil
	}

	// 禁用规则
	if err := s.repo.DisableRule(ctx, id); err != nil {
		s.logger.WithContext(ctx).Error("禁用规则失败",
			logger.NewField("error", err.Error()),
			logger.NewField("rule_id", id))
		return err
	}

	s.logger.WithContext(ctx).Info("禁用规则成功",
		logger.NewField("rule_id", id))

	return nil
}

// ValidateRules 执行规则校验
func (s *RuleService) ValidateRules(ctx context.Context, data interface{}, ruleIDs []string) ([]*RuleValidationResult, error) {
	// TODO: 实现规则校验逻辑
	return nil, nil
}

// ValidateAllRules 执行所有规则校验
func (s *RuleService) ValidateAllRules(ctx context.Context, data interface{}) ([]*RuleValidationResult, error) {
	// TODO: 实现所有规则校验逻辑
	return nil, nil
}

// ValidateRuleByType 按类型执行规则校验
func (s *RuleService) ValidateRuleByType(ctx context.Context, data interface{}, ruleType string) ([]*RuleValidationResult, error) {
	// TODO: 实现按类型规则校验逻辑
	return nil, nil
}

// TestRule 测试规则
func (s *RuleService) TestRule(ctx context.Context, rule *Rule, testData interface{}) (*RuleValidationResult, error) {
	// TODO: 实现测试规则逻辑
	return nil, nil
}

// LoadRules 加载规则
func (s *RuleService) LoadRules(ctx context.Context) error {
	// TODO: 实现加载规则逻辑
	return nil
}

// ReloadRules 重新加载规则
func (s *RuleService) ReloadRules(ctx context.Context) error {
	// TODO: 实现重新加载规则逻辑
	return nil
}

// GetRuleTypes 获取规则类型列表
func (s *RuleService) GetRuleTypes(ctx context.Context) ([]string, error) {
	// TODO: 实现获取规则类型列表逻辑
	return []string{RuleTypeAmount, RuleTypeFrequency, RuleTypeInvoice, RuleTypeCompliance, RuleTypeCustom}, nil
}

// ResolveRuleConflicts 解决规则冲突
func (s *RuleService) ResolveRuleConflicts(results []*RuleValidationResult) []*RuleValidationResult {
	// TODO: 实现解决规则冲突逻辑
	return nil
}

// SortRulesByPriority 按优先级排序规则
func (s *RuleService) SortRulesByPriority(rules []*Rule) []*Rule {
	// TODO: 实现按优先级排序规则逻辑
	return nil
}
