package mysql

import (
	"context"
	"errors"
	"time"

	"reimbursement-audit/internal/domain/ruleengine"
	"reimbursement-audit/internal/pkg/logger"

	"gorm.io/gorm"
)

type RuleEngineRepository struct {
	client *Client
	logger logger.Logger
}

func NewRuleEngineRepository(client *Client, logger logger.Logger) *RuleEngineRepository {
	return &RuleEngineRepository{
		client: client,
		logger: logger,
	}
}

func (r *RuleEngineRepository) CreateRule(ctx context.Context, rule *ruleengine.Rule) error {
	now := time.Now()
	rule.CreatedAt = now
	rule.UpdatedAt = now

	result := r.client.GetDB().WithContext(ctx).Create(rule)
	if result.Error != nil {
		r.logger.WithContext(ctx).Error("创建规则失败",
			logger.NewField("error", result.Error.Error()),
			logger.NewField("rule_name", rule.Name))
		return result.Error
	}

	r.logger.WithContext(ctx).Info("创建规则成功",
		logger.NewField("rule_id", rule.ID),
		logger.NewField("rule_name", rule.Name))
	return nil
}

func (r *RuleEngineRepository) GetRuleByID(ctx context.Context, id string) (*ruleengine.Rule, error) {
	var rule ruleengine.Rule
	result := r.client.GetDB().WithContext(ctx).Preload("Conditions").Preload("Decision").Where("id = ?", id).First(&rule)
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

func (r *RuleEngineRepository) UpdateRule(ctx context.Context, rule *ruleengine.Rule) error {
	rule.UpdatedAt = time.Now()

	result := r.client.GetDB().WithContext(ctx).Model(rule).Where("id = ?", rule.ID).Updates(rule)
	if result.Error != nil {
		r.logger.WithContext(ctx).Error("更新规则失败",
			logger.NewField("error", result.Error.Error()),
			logger.NewField("rule_id", rule.ID))
		return result.Error
	}

	if result.RowsAffected == 0 {
		r.logger.WithContext(ctx).Warn("规则不存在，更新失败",
			logger.NewField("rule_id", rule.ID))
		return errors.New("规则不存在")
	}

	r.logger.WithContext(ctx).Info("更新规则成功",
		logger.NewField("rule_id", rule.ID))
	return nil
}

func (r *RuleEngineRepository) DeleteRule(ctx context.Context, id string) error {
	result := r.client.GetDB().WithContext(ctx).Delete(&ruleengine.Rule{}, "id = ?", id)
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

func (r *RuleEngineRepository) ListRules(ctx context.Context, filter *ruleengine.RuleFilter) ([]*ruleengine.Rule, int64, error) {
	var rules []*ruleengine.Rule
	var total int64

	db := r.client.GetDB().WithContext(ctx).Model(&ruleengine.Rule{})

	if filter != nil {
		if filter.Name != "" {
			db = db.Where("name LIKE ?", "%"+filter.Name+"%")
		}
		if filter.Enabled != nil {
			db = db.Where("enabled = ?", *filter.Enabled)
		}
		if filter.FeatureID != "" {
			db = db.Joins("JOIN conditions ON conditions.rule_id = rules.id").
				Where("conditions.feature_id = ?", filter.FeatureID)
		}
	}

	if err := db.Count(&total).Error; err != nil {
		r.logger.WithContext(ctx).Error("统计规则数量失败",
			logger.NewField("error", err.Error()))
		return nil, 0, err
	}

	if filter != nil && filter.Page > 0 && filter.Size > 0 {
		offset := (filter.Page - 1) * filter.Size
		db = db.Offset(offset).Limit(filter.Size)
	}

	db = db.Order("priority DESC, updated_at DESC")

	if err := db.Preload("Conditions").Preload("Decision").Find(&rules).Error; err != nil {
		r.logger.WithContext(ctx).Error("查询规则列表失败",
			logger.NewField("error", err.Error()))
		return nil, 0, err
	}

	for _, rule := range rules {
		var conditions []ruleengine.Condition
		if err := r.client.GetDB().WithContext(ctx).Where("rule_id = ?", rule.ID).Order("sort_order ASC").Find(&conditions).Error; err == nil {
			rule.Conditions = conditions
		}
	}

	return rules, total, nil
}

func (r *RuleEngineRepository) EnableRule(ctx context.Context, id string) error {
	result := r.client.GetDB().WithContext(ctx).Model(&ruleengine.Rule{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"enabled":    true,
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

func (r *RuleEngineRepository) DisableRule(ctx context.Context, id string) error {
	result := r.client.GetDB().WithContext(ctx).Model(&ruleengine.Rule{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"enabled":    false,
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

func (r *RuleEngineRepository) CreateCondition(ctx context.Context, condition *ruleengine.Condition) error {
	result := r.client.GetDB().WithContext(ctx).Create(condition)
	if result.Error != nil {
		r.logger.WithContext(ctx).Error("创建条件失败",
			logger.NewField("error", result.Error.Error()))
		return result.Error
	}

	r.logger.WithContext(ctx).Info("创建条件成功",
		logger.NewField("condition_id", condition.ID))
	return nil
}

func (r *RuleEngineRepository) GetConditionByID(ctx context.Context, id string) (*ruleengine.Condition, error) {
	var condition ruleengine.Condition
	result := r.client.GetDB().WithContext(ctx).Where("id = ?", id).First(&condition)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			r.logger.WithContext(ctx).Warn("条件不存在",
				logger.NewField("condition_id", id))
			return nil, result.Error
		}
		r.logger.WithContext(ctx).Error("获取条件失败",
			logger.NewField("error", result.Error.Error()),
			logger.NewField("condition_id", id))
		return nil, result.Error
	}
	return &condition, nil
}

func (r *RuleEngineRepository) UpdateCondition(ctx context.Context, condition *ruleengine.Condition) error {
	result := r.client.GetDB().WithContext(ctx).Model(condition).Where("id = ?", condition.ID).Updates(condition)
	if result.Error != nil {
		r.logger.WithContext(ctx).Error("更新条件失败",
			logger.NewField("error", result.Error.Error()))
		return result.Error
	}

	if result.RowsAffected == 0 {
		r.logger.WithContext(ctx).Warn("条件不存在，更新失败",
			logger.NewField("condition_id", condition.ID))
		return errors.New("条件不存在")
	}

	r.logger.WithContext(ctx).Info("更新条件成功",
		logger.NewField("condition_id", condition.ID))
	return nil
}

func (r *RuleEngineRepository) DeleteCondition(ctx context.Context, id string) error {
	result := r.client.GetDB().WithContext(ctx).Delete(&ruleengine.Condition{}, "id = ?", id)
	if result.Error != nil {
		r.logger.WithContext(ctx).Error("删除条件失败",
			logger.NewField("error", result.Error.Error()))
		return result.Error
	}

	if result.RowsAffected == 0 {
		r.logger.WithContext(ctx).Warn("条件不存在，删除失败",
			logger.NewField("condition_id", id))
		return errors.New("条件不存在")
	}

	r.logger.WithContext(ctx).Info("删除条件成功",
		logger.NewField("condition_id", id))
	return nil
}

func (r *RuleEngineRepository) ListConditions(ctx context.Context, filter *ruleengine.ConditionFilter) ([]*ruleengine.Condition, error) {
	var conditions []*ruleengine.Condition

	db := r.client.GetDB().WithContext(ctx).Model(&ruleengine.Condition{})

	if filter != nil {
		if filter.RuleID != "" {
			db = db.Where("rule_id = ?", filter.RuleID)
		}
		if filter.FeatureID != "" {
			db = db.Where("feature_id = ?", filter.FeatureID)
		}
	}

	if err := db.Order("sort_order ASC").Find(&conditions).Error; err != nil {
		r.logger.WithContext(ctx).Error("查询条件列表失败",
			logger.NewField("error", err.Error()))
		return nil, err
	}

	return conditions, nil
}

func (r *RuleEngineRepository) DeleteConditionsByRuleID(ctx context.Context, ruleID string) error {
	result := r.client.GetDB().WithContext(ctx).Where("rule_id = ?", ruleID).Delete(&ruleengine.Condition{})
	if result.Error != nil {
		r.logger.WithContext(ctx).Error("删除规则条件失败",
			logger.NewField("error", result.Error.Error()),
			logger.NewField("rule_id", ruleID))
		return result.Error
	}

	r.logger.WithContext(ctx).Info("删除规则条件成功",
		logger.NewField("rule_id", ruleID))
	return nil
}

func (r *RuleEngineRepository) CreateDecision(ctx context.Context, decision *ruleengine.Decision) error {
	now := time.Now()
	decision.CreatedAt = now
	decision.UpdatedAt = now

	result := r.client.GetDB().WithContext(ctx).Create(decision)
	if result.Error != nil {
		r.logger.WithContext(ctx).Error("创建决策失败",
			logger.NewField("error", result.Error.Error()))
		return result.Error
	}

	r.logger.WithContext(ctx).Info("创建决策成功",
		logger.NewField("decision_id", decision.ID))
	return nil
}

func (r *RuleEngineRepository) GetDecisionByRuleID(ctx context.Context, ruleID string) (*ruleengine.Decision, error) {
	var decision ruleengine.Decision
	result := r.client.GetDB().WithContext(ctx).Where("rule_id = ?", ruleID).First(&decision)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		r.logger.WithContext(ctx).Error("获取决策失败",
			logger.NewField("error", result.Error.Error()),
			logger.NewField("rule_id", ruleID))
		return nil, result.Error
	}
	return &decision, nil
}

func (r *RuleEngineRepository) UpdateDecision(ctx context.Context, decision *ruleengine.Decision) error {
	decision.UpdatedAt = time.Now()

	result := r.client.GetDB().WithContext(ctx).Model(decision).Where("id = ?", decision.ID).Updates(decision)
	if result.Error != nil {
		r.logger.WithContext(ctx).Error("更新决策失败",
			logger.NewField("error", result.Error.Error()))
		return result.Error
	}

	if result.RowsAffected == 0 {
		r.logger.WithContext(ctx).Warn("决策不存在，更新失败",
			logger.NewField("decision_id", decision.ID))
		return errors.New("决策不存在")
	}

	r.logger.WithContext(ctx).Info("更新决策成功",
		logger.NewField("decision_id", decision.ID))
	return nil
}

func (r *RuleEngineRepository) DeleteDecision(ctx context.Context, id string) error {
	result := r.client.GetDB().WithContext(ctx).Delete(&ruleengine.Decision{}, "id = ?", id)
	if result.Error != nil {
		r.logger.WithContext(ctx).Error("删除决策失败",
			logger.NewField("error", result.Error.Error()))
		return result.Error
	}

	if result.RowsAffected == 0 {
		r.logger.WithContext(ctx).Warn("决策不存在，删除失败",
			logger.NewField("decision_id", id))
		return errors.New("决策不存在")
	}

	r.logger.WithContext(ctx).Info("删除决策成功",
		logger.NewField("decision_id", id))
	return nil
}

func (r *RuleEngineRepository) CreateFeature(ctx context.Context, feature *ruleengine.Feature) error {
	now := time.Now()
	feature.CreatedAt = now
	feature.UpdatedAt = now

	result := r.client.GetDB().WithContext(ctx).Create(feature)
	if result.Error != nil {
		r.logger.WithContext(ctx).Error("创建特征失败",
			logger.NewField("error", result.Error.Error()),
			logger.NewField("feature_name", feature.Name))
		return result.Error
	}

	r.logger.WithContext(ctx).Info("创建特征成功",
		logger.NewField("feature_id", feature.ID),
		logger.NewField("feature_name", feature.Name))
	return nil
}

func (r *RuleEngineRepository) GetFeatureByID(ctx context.Context, id string) (*ruleengine.Feature, error) {
	var feature ruleengine.Feature
	result := r.client.GetDB().WithContext(ctx).Preload("Values").Where("id = ?", id).First(&feature)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			r.logger.WithContext(ctx).Warn("特征不存在",
				logger.NewField("feature_id", id))
			return nil, result.Error
		}
		r.logger.WithContext(ctx).Error("获取特征失败",
			logger.NewField("error", result.Error.Error()),
			logger.NewField("feature_id", id))
		return nil, result.Error
	}
	return &feature, nil
}

func (r *RuleEngineRepository) GetFeatureByCode(ctx context.Context, code string) (*ruleengine.Feature, error) {
	var feature ruleengine.Feature
	result := r.client.GetDB().WithContext(ctx).Preload("Values").Where("code = ?", code).First(&feature)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			r.logger.WithContext(ctx).Warn("特征不存在",
				logger.NewField("feature_code", code))
			return nil, result.Error
		}
		r.logger.WithContext(ctx).Error("获取特征失败",
			logger.NewField("error", result.Error.Error()),
			logger.NewField("feature_code", code))
		return nil, result.Error
	}
	return &feature, nil
}

func (r *RuleEngineRepository) UpdateFeature(ctx context.Context, feature *ruleengine.Feature) error {
	feature.UpdatedAt = time.Now()

	result := r.client.GetDB().WithContext(ctx).Model(feature).Where("id = ?", feature.ID).Updates(feature)
	if result.Error != nil {
		r.logger.WithContext(ctx).Error("更新特征失败",
			logger.NewField("error", result.Error.Error()),
			logger.NewField("feature_id", feature.ID))
		return result.Error
	}

	if result.RowsAffected == 0 {
		r.logger.WithContext(ctx).Warn("特征不存在，更新失败",
			logger.NewField("feature_id", feature.ID))
		return errors.New("特征不存在")
	}

	r.logger.WithContext(ctx).Info("更新特征成功",
		logger.NewField("feature_id", feature.ID))
	return nil
}

func (r *RuleEngineRepository) DeleteFeature(ctx context.Context, id string) error {
	result := r.client.GetDB().WithContext(ctx).Delete(&ruleengine.Feature{}, "id = ?", id)
	if result.Error != nil {
		r.logger.WithContext(ctx).Error("删除特征失败",
			logger.NewField("error", result.Error.Error()),
			logger.NewField("feature_id", id))
		return result.Error
	}

	if result.RowsAffected == 0 {
		r.logger.WithContext(ctx).Warn("特征不存在，删除失败",
			logger.NewField("feature_id", id))
		return errors.New("规则不存在")
	}

	r.logger.WithContext(ctx).Info("删除特征成功",
		logger.NewField("feature_id", id))
	return nil
}

func (r *RuleEngineRepository) ListFeatures(ctx context.Context, filter *ruleengine.FeatureFilter) ([]*ruleengine.Feature, int64, error) {
	var features []*ruleengine.Feature
	var total int64

	db := r.client.GetDB().WithContext(ctx).Model(&ruleengine.Feature{})

	if filter != nil {
		if filter.Name != "" {
			db = db.Where("name LIKE ?", "%"+filter.Name+"%")
		}
		if filter.Code != "" {
			db = db.Where("code LIKE ?", "%"+filter.Code+"%")
		}
		if filter.Category != "" {
			db = db.Where("category = ?", filter.Category)
		}
		if filter.Type != "" {
			db = db.Where("type = ?", filter.Type)
		}
		if filter.Enabled != nil {
			db = db.Where("enabled = ?", *filter.Enabled)
		}
	}

	if err := db.Count(&total).Error; err != nil {
		r.logger.WithContext(ctx).Error("统计特征数量失败",
			logger.NewField("error", err.Error()))
		return nil, 0, err
	}

	if filter != nil && filter.Page > 0 && filter.Size > 0 {
		offset := (filter.Page - 1) * filter.Size
		db = db.Offset(offset).Limit(filter.Size)
	}

	db = db.Order("created_at DESC")

	if err := db.Preload("Values").Find(&features).Error; err != nil {
		r.logger.WithContext(ctx).Error("查询特征列表失败",
			logger.NewField("error", err.Error()))
		return nil, 0, err
	}

	return features, total, nil
}

func (r *RuleEngineRepository) EnableFeature(ctx context.Context, id string) error {
	result := r.client.GetDB().WithContext(ctx).Model(&ruleengine.Feature{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"enabled":    true,
			"updated_at": time.Now(),
		})

	if result.Error != nil {
		r.logger.WithContext(ctx).Error("启用特征失败",
			logger.NewField("error", result.Error.Error()),
			logger.NewField("feature_id", id))
		return result.Error
	}

	if result.RowsAffected == 0 {
		r.logger.WithContext(ctx).Warn("特征不存在，启用失败",
			logger.NewField("feature_id", id))
		return errors.New("特征不存在")
	}

	r.logger.WithContext(ctx).Info("启用特征成功",
		logger.NewField("feature_id", id))
	return nil
}

func (r *RuleEngineRepository) DisableFeature(ctx context.Context, id string) error {
	result := r.client.GetDB().WithContext(ctx).Model(&ruleengine.Feature{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"enabled":    false,
			"updated_at": time.Now(),
		})

	if result.Error != nil {
		r.logger.WithContext(ctx).Error("禁用特征失败",
			logger.NewField("error", result.Error.Error()),
			logger.NewField("feature_id", id))
		return result.Error
	}

	if result.RowsAffected == 0 {
		r.logger.WithContext(ctx).Warn("特征不存在，禁用失败",
			logger.NewField("feature_id", id))
		return errors.New("规则不存在")
	}

	r.logger.WithContext(ctx).Info("禁用特征成功",
		logger.NewField("feature_id", id))
	return nil
}

func (r *RuleEngineRepository) CreateFeatureValue(ctx context.Context, value *ruleengine.FeatureValue) error {
	now := time.Now()
	value.CreatedAt = now
	value.UpdatedAt = now

	result := r.client.GetDB().WithContext(ctx).Create(value)
	if result.Error != nil {
		r.logger.WithContext(ctx).Error("创建特征值失败",
			logger.NewField("error", result.Error.Error()))
		return result.Error
	}

	r.logger.WithContext(ctx).Info("创建特征值成功",
		logger.NewField("feature_value_id", value.ID))
	return nil
}

func (r *RuleEngineRepository) GetFeatureValueByID(ctx context.Context, id string) (*ruleengine.FeatureValue, error) {
	var value ruleengine.FeatureValue
	result := r.client.GetDB().WithContext(ctx).Where("id = ?", id).First(&value)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			r.logger.WithContext(ctx).Warn("特征值不存在",
				logger.NewField("feature_value_id", id))
			return nil, result.Error
		}
		r.logger.WithContext(ctx).Error("获取特征值失败",
			logger.NewField("error", result.Error.Error()),
			logger.NewField("feature_value_id", id))
		return nil, result.Error
	}
	return &value, nil
}

func (r *RuleEngineRepository) UpdateFeatureValue(ctx context.Context, value *ruleengine.FeatureValue) error {
	value.UpdatedAt = time.Now()

	result := r.client.GetDB().WithContext(ctx).Model(value).Where("id = ?", value.ID).Updates(value)
	if result.Error != nil {
		r.logger.WithContext(ctx).Error("更新特征值失败",
			logger.NewField("error", result.Error.Error()))
		return result.Error
	}

	if result.RowsAffected == 0 {
		r.logger.WithContext(ctx).Warn("特征值不存在，更新失败",
			logger.NewField("feature_value_id", value.ID))
		return errors.New("特征不存在")
	}

	r.logger.WithContext(ctx).Info("更新特征值成功",
		logger.NewField("feature_value_id", value.ID))
	return nil
}

func (r *RuleEngineRepository) DeleteFeatureValue(ctx context.Context, id string) error {
	result := r.client.GetDB().WithContext(ctx).Delete(&ruleengine.FeatureValue{}, "id = ?", id)
	if result.Error != nil {
		r.logger.WithContext(ctx).Error("删除特征值失败",
			logger.NewField("error", result.Error.Error()))
		return result.Error
	}

	if result.RowsAffected == 0 {
		r.logger.WithContext(ctx).Warn("特征值不存在，删除失败",
			logger.NewField("feature_value_id", id))
		return errors.New("特征不存在")
	}

	r.logger.WithContext(ctx).Info("删除特征值成功",
		logger.NewField("feature_value_id", id))
	return nil
}

func (r *RuleEngineRepository) ListFeatureValues(ctx context.Context, filter *ruleengine.FeatureValueFilter) ([]*ruleengine.FeatureValue, error) {
	var values []*ruleengine.FeatureValue

	db := r.client.GetDB().WithContext(ctx).Model(&ruleengine.FeatureValue{})

	if filter != nil {
		if filter.FeatureID != "" {
			db = db.Where("feature_id = ?", filter.FeatureID)
		}
		if filter.Enabled != nil {
			db = db.Where("enabled = ?", *filter.Enabled)
		}
	}

	if err := db.Order("sort_order ASC").Find(&values).Error; err != nil {
		r.logger.WithContext(ctx).Error("查询特征值列表失败",
			logger.NewField("error", err.Error()))
		return nil, err
	}

	return values, nil
}

func (r *RuleEngineRepository) DeleteFeatureValuesByFeatureID(ctx context.Context, featureID string) error {
	result := r.client.GetDB().WithContext(ctx).Where("feature_id = ?", featureID).Delete(&ruleengine.FeatureValue{})
	if result.Error != nil {
		r.logger.WithContext(ctx).Error("删除特征值失败",
			logger.NewField("error", result.Error.Error()),
			logger.NewField("feature_id", featureID))
		return result.Error
	}

	r.logger.WithContext(ctx).Info("删除特征值成功",
		logger.NewField("feature_id", featureID))
	return nil
}
