// rule_repository.go MySQL规则仓储实现
// 功能点：
// 1. 实现规则仓储接口
// 2. 提供MySQL数据访问实现
// 3. 支持规则CRUD操作
// 4. 支持规则查询和筛选

package mysql

import (
	"context"
	"errors"
	"time"

	"reimbursement-audit/internal/domain/rule"
	"reimbursement-audit/internal/pkg/logger"

	"gorm.io/gorm"
)

// RuleRepository 规则仓储实现
type RuleRepository struct {
	client *Client
	logger logger.Logger
}

// NewRuleRepository 创建规则仓储实例
func NewRuleRepository(client *Client, logger logger.Logger) rule.Repository {
	return &RuleRepository{
		client: client,
		logger: logger,
	}
}

// CreateRule 创建规则
func (r *RuleRepository) CreateRule(ctx context.Context, rule *rule.Rule) error {
	// 检查规则编码是否已存在
	exists, err := r.CheckRuleCodeExists(ctx, rule.RuleCode, "")
	if err != nil {
		r.logger.WithContext(ctx).Error("检查规则编码唯一性失败",
			logger.NewField("error", err.Error()),
			logger.NewField("rule_code", rule.RuleCode))
		return err
	}
	if exists {
		r.logger.WithContext(ctx).Warn("规则编码已存在",
			logger.NewField("rule_code", rule.RuleCode))
		return errors.New("规则编码已存在")
	}

	// 设置创建时间和更新时间
	now := time.Now()
	rule.CreatedAt = now
	rule.UpdatedAt = now

	// 使用GORM创建规则
	result := r.client.GetDB().WithContext(ctx).Create(rule)
	if result.Error != nil {
		r.logger.WithContext(ctx).Error("创建规则失败",
			logger.NewField("error", result.Error.Error()),
			logger.NewField("rule_code", rule.RuleCode),
			logger.NewField("rule_name", rule.Name))
		return result.Error
	}

	r.logger.WithContext(ctx).Info("创建规则成功",
		logger.NewField("rule_id", rule.ID),
		logger.NewField("rule_code", rule.RuleCode),
		logger.NewField("rule_name", rule.Name))

	return nil
}

// GetRuleByID 根据ID获取规则
func (r *RuleRepository) GetRuleByID(ctx context.Context, id string) (*rule.Rule, error) {
	var rule rule.Rule

	// 使用GORM查询规则
	result := r.client.GetDB().WithContext(ctx).Where("id = ?", id).First(&rule)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			r.logger.WithContext(ctx).Warn("规则不存在",
				logger.NewField("rule_id", id))
			return nil, result.Error
		}
		r.logger.WithContext(ctx).Error("获取规则失败",
			logger.NewField("error", result.Error.Error()),
			logger.NewField("rule_id", id))
		return nil, result.Error
	}

	return &rule, nil
}

// GetRuleByCode 根据规则编码获取规则
func (r *RuleRepository) GetRuleByCode(ctx context.Context, ruleCode string) (*rule.Rule, error) {
	var rule rule.Rule

	// 使用GORM查询规则
	result := r.client.GetDB().WithContext(ctx).Where("rule_code = ?", ruleCode).First(&rule)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			r.logger.WithContext(ctx).Warn("规则不存在",
				logger.NewField("rule_code", ruleCode))
			return nil, result.Error
		}
		r.logger.WithContext(ctx).Error("获取规则失败",
			logger.NewField("error", result.Error.Error()),
			logger.NewField("rule_code", ruleCode))
		return nil, result.Error
	}

	return &rule, nil
}

// UpdateRule 更新规则
func (r *RuleRepository) UpdateRule(ctx context.Context, rule *rule.Rule) error {
	// 检查规则编码是否已存在（排除当前规则）
	exists, err := r.CheckRuleCodeExists(ctx, rule.RuleCode, rule.ID)
	if err != nil {
		r.logger.WithContext(ctx).Error("检查规则编码唯一性失败",
			logger.NewField("error", err.Error()),
			logger.NewField("rule_code", rule.RuleCode))
		return err
	}
	if exists {
		r.logger.WithContext(ctx).Warn("规则编码已存在",
			logger.NewField("rule_code", rule.RuleCode))
		return errors.New("规则编码已存在")
	}

	// 设置更新时间
	rule.UpdatedAt = time.Now()

	// 使用GORM更新规则
	result := r.client.GetDB().WithContext(ctx).Model(rule).Where("id = ?", rule.ID).Updates(rule)
	if result.Error != nil {
		r.logger.WithContext(ctx).Error("更新规则失败",
			logger.NewField("error", result.Error.Error()),
			logger.NewField("rule_id", rule.ID),
			logger.NewField("rule_code", rule.RuleCode))
		return result.Error
	}

	if result.RowsAffected == 0 {
		r.logger.WithContext(ctx).Warn("规则不存在，更新失败",
			logger.NewField("rule_id", rule.ID))
		return errors.New("规则不存在")
	}

	r.logger.WithContext(ctx).Info("更新规则成功",
		logger.NewField("rule_id", rule.ID),
		logger.NewField("rule_code", rule.RuleCode))

	return nil
}

// DeleteRule 删除规则
func (r *RuleRepository) DeleteRule(ctx context.Context, id string) error {
	// 使用GORM删除规则
	result := r.client.GetDB().WithContext(ctx).Delete(&rule.Rule{}, "id = ?", id)
	if result.Error != nil {
		r.logger.WithContext(ctx).Error("删除规则失败",
			logger.NewField("error", result.Error.Error()),
			logger.NewField("rule_id", id))
		return result.Error
	}

	if result.RowsAffected == 0 {
		r.logger.WithContext(ctx).Warn("规则不存在，删除失败",
			logger.NewField("rule_id", id))
		return errors.New("规则不存在")
	}

	r.logger.WithContext(ctx).Info("删除规则成功",
		logger.NewField("rule_id", id))

	return nil
}

// ListRules 根据过滤条件查询规则列表
func (r *RuleRepository) ListRules(ctx context.Context, filter *rule.RuleFilter) ([]*rule.Rule, int64, error) {
	var rules []*rule.Rule
	var total int64

	// 构建查询
	db := r.client.GetDB().WithContext(ctx).Model(&rule.Rule{})

	// 应用过滤条件
	if filter != nil {
		if filter.RuleCode != "" {
			db = db.Where("rule_code LIKE ?", "%"+filter.RuleCode+"%")
		}
		if filter.Type != "" {
			db = db.Where("type = ?", filter.Type)
		}
		if filter.Category != "" {
			db = db.Where("category = ?", filter.Category)
		}
		if filter.Status != "" {
			db = db.Where("status = ?", filter.Status)
		}
		if filter.Enabled != nil {
			db = db.Where("enabled = ?", *filter.Enabled)
		}
	}

	// 获取总数
	if err := db.Count(&total).Error; err != nil {
		r.logger.WithContext(ctx).Error("统计规则数量失败",
			logger.NewField("error", err.Error()))
		return nil, 0, err
	}

	// 应用分页
	if filter != nil && filter.Page > 0 && filter.Size > 0 {
		offset := (filter.Page - 1) * filter.Size
		db = db.Offset(offset).Limit(filter.Size)
	}

	// 排序
	db = db.Order("priority DESC, updated_at DESC")

	// 查询数据
	if err := db.Find(&rules).Error; err != nil {
		r.logger.WithContext(ctx).Error("查询规则列表失败",
			logger.NewField("error", err.Error()))
		return nil, 0, err
	}

	r.logger.WithContext(ctx).Info("查询规则列表成功",
		logger.NewField("total", total),
		logger.NewField("count", len(rules)))

	return rules, total, nil
}

// CountRules 根据过滤条件统计规则数量
func (r *RuleRepository) CountRules(ctx context.Context, filter *rule.RuleFilter) (int64, error) {
	var count int64

	// 构建查询
	db := r.client.GetDB().WithContext(ctx).Model(&rule.Rule{})

	// 应用过滤条件
	if filter != nil {
		if filter.RuleCode != "" {
			db = db.Where("rule_code LIKE ?", "%"+filter.RuleCode+"%")
		}
		if filter.Type != "" {
			db = db.Where("type = ?", filter.Type)
		}
		if filter.Category != "" {
			db = db.Where("category = ?", filter.Category)
		}
		if filter.Status != "" {
			db = db.Where("status = ?", filter.Status)
		}
		if filter.Enabled != nil {
			db = db.Where("enabled = ?", *filter.Enabled)
		}
	}

	// 获取总数
	if err := db.Count(&count).Error; err != nil {
		r.logger.WithContext(ctx).Error("统计规则数量失败",
			logger.NewField("error", err.Error()))
		return 0, err
	}

	return count, nil
}

// EnableRule 启用规则
func (r *RuleRepository) EnableRule(ctx context.Context, id string) error {
	// 使用GORM更新规则状态
	result := r.client.GetDB().WithContext(ctx).Model(&rule.Rule{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"enabled":    true,
			"status":     rule.RuleStatusEnabled,
			"updated_at": time.Now(),
		})

	if result.Error != nil {
		r.logger.WithContext(ctx).Error("启用规则失败",
			logger.NewField("error", result.Error.Error()),
			logger.NewField("rule_id", id))
		return result.Error
	}

	if result.RowsAffected == 0 {
		r.logger.WithContext(ctx).Warn("规则不存在，启用失败",
			logger.NewField("rule_id", id))
		return errors.New("规则不存在")
	}

	r.logger.WithContext(ctx).Info("启用规则成功",
		logger.NewField("rule_id", id))

	return nil
}

// DisableRule 禁用规则
func (r *RuleRepository) DisableRule(ctx context.Context, id string) error {
	// 使用GORM更新规则状态
	result := r.client.GetDB().WithContext(ctx).Model(&rule.Rule{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"enabled":    false,
			"status":     rule.RuleStatusDisabled,
			"updated_at": time.Now(),
		})

	if result.Error != nil {
		r.logger.WithContext(ctx).Error("禁用规则失败",
			logger.NewField("error", result.Error.Error()),
			logger.NewField("rule_id", id))
		return result.Error
	}

	if result.RowsAffected == 0 {
		r.logger.WithContext(ctx).Warn("规则不存在，禁用失败",
			logger.NewField("rule_id", id))
		return errors.New("规则不存在")
	}

	r.logger.WithContext(ctx).Info("禁用规则成功",
		logger.NewField("rule_id", id))

	return nil
}

// CheckRuleCodeExists 检查规则编码是否存在
func (r *RuleRepository) CheckRuleCodeExists(ctx context.Context, ruleCode string, excludeID string) (bool, error) {
	var count int64

	// 构建查询
	db := r.client.GetDB().WithContext(ctx).Model(&rule.Rule{}).Where("rule_code = ?", ruleCode)

	// 如果提供了排除ID，则添加排除条件
	if excludeID != "" {
		db = db.Where("id != ?", excludeID)
	}

	// 获取数量
	if err := db.Count(&count).Error; err != nil {
		r.logger.WithContext(ctx).Error("检查规则编码唯一性失败",
			logger.NewField("error", err.Error()),
			logger.NewField("rule_code", ruleCode))
		return false, err
	}

	return count > 0, nil
}
