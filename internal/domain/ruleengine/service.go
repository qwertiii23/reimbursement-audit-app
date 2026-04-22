package ruleengine

import (
	"context"
	"errors"
	"fmt"

	"reimbursement-audit/internal/domain/featurefunction"
	"reimbursement-audit/internal/pkg/logger"

	"github.com/google/uuid"
)

type RuleEngineService struct {
	ruleRepo                RuleRepository
	conditionRepo           ConditionRepository
	decisionRepo            DecisionRepository
	featureRepo             FeatureRepository
	featureValueRepo        FeatureValueRepository
	featureFunctionRegistry *featurefunction.FunctionRegistry
	logger                  logger.Logger
	reloadCallback          func(ctx context.Context) error // 规则重新加载回调函数
	engine                  *Engine                         // 规则引擎实例
}

func NewRuleEngineService(
	ruleRepo RuleRepository,
	conditionRepo ConditionRepository,
	decisionRepo DecisionRepository,
	featureRepo FeatureRepository,
	featureValueRepo FeatureValueRepository,
	featureFunctionRegistry *featurefunction.FunctionRegistry,
	logger logger.Logger,
) *RuleEngineService {
	return &RuleEngineService{
		ruleRepo:                ruleRepo,
		conditionRepo:           conditionRepo,
		decisionRepo:            decisionRepo,
		featureRepo:             featureRepo,
		featureValueRepo:        featureValueRepo,
		featureFunctionRegistry: featureFunctionRegistry,
		logger:                  logger,
	}
}

func (s *RuleEngineService) SetEngine(engine *Engine) {
	s.engine = engine
}

func (s *RuleEngineService) SetReloadCallback(callback func(ctx context.Context) error) {
	s.reloadCallback = callback
}

func (s *RuleEngineService) triggerReload(ctx context.Context) {
	if s.reloadCallback != nil {
		if err := s.reloadCallback(ctx); err != nil {
			s.logger.WithContext(ctx).Error("重新加载规则失败", logger.NewField("error", err.Error()))
		} else {
			s.logger.WithContext(ctx).Info("规则重新加载成功")
		}
	}
}

func (s *RuleEngineService) CreateRule(ctx context.Context, req *CreateRuleRequest) (*Rule, error) {
	if req.Name == "" {
		return nil, errors.New("规则名称不能为空")
	}
	if len(req.Conditions) == 0 {
		return nil, errors.New("规则至少需要一个条件")
	}
	if req.Decision.Type == "" {
		return nil, errors.New("决策类型不能为空")
	}

	rule := &Rule{
		ID:          uuid.New().String(),
		Name:        req.Name,
		Description: req.Description,
		Priority:    req.Priority,
		Enabled:     req.Enabled,
	}

	if err := s.ruleRepo.CreateRule(ctx, rule); err != nil {
		return nil, err
	}

	for _, cond := range req.Conditions {
		condition := &Condition{
			ID:        uuid.New().String(),
			RuleID:    rule.ID,
			FeatureID: cond.FeatureID,
			Operator:  cond.Operator,
			Value:     cond.Value,
			LogicOp:   cond.LogicOp,
			SortOrder: cond.SortOrder,
		}
		if err := s.conditionRepo.CreateCondition(ctx, condition); err != nil {
			s.logger.WithContext(ctx).Error("创建条件失败", logger.NewField("error", err.Error()))
			return nil, err
		}
	}

	decision := &Decision{
		ID:     uuid.New().String(),
		RuleID: rule.ID,
		Type:   req.Decision.Type,
		Reason: req.Decision.Reason,
	}
	if err := s.decisionRepo.CreateDecision(ctx, decision); err != nil {
		s.logger.WithContext(ctx).Error("创建决策失败", logger.NewField("error", err.Error()))
		return nil, err
	}

	s.logger.WithContext(ctx).Info("创建规则成功", logger.NewField("rule_id", rule.ID))
	s.triggerReload(ctx)
	return rule, nil
}

func (s *RuleEngineService) UpdateRule(ctx context.Context, req *UpdateRuleRequest) (*Rule, error) {
	if req.ID == "" {
		return nil, errors.New("规则ID不能为空")
	}

	rule, err := s.ruleRepo.GetRuleByID(ctx, req.ID)
	if err != nil {
		return nil, err
	}

	rule.Name = req.Name
	rule.Description = req.Description
	rule.Priority = req.Priority
	rule.Enabled = req.Enabled

	if err := s.ruleRepo.UpdateRule(ctx, rule); err != nil {
		return nil, err
	}

	if err := s.conditionRepo.DeleteConditionsByRuleID(ctx, req.ID); err != nil {
		return nil, err
	}

	for _, cond := range req.Conditions {
		condition := &Condition{
			ID:        uuid.New().String(),
			RuleID:    rule.ID,
			FeatureID: cond.FeatureID,
			Operator:  cond.Operator,
			Value:     cond.Value,
			LogicOp:   cond.LogicOp,
			SortOrder: cond.SortOrder,
		}
		if err := s.conditionRepo.CreateCondition(ctx, condition); err != nil {
			s.logger.WithContext(ctx).Error("创建条件失败", logger.NewField("error", err.Error()))
			return nil, err
		}
	}

	decision, err := s.decisionRepo.GetDecisionByRuleID(ctx, req.ID)
	if err != nil {
		decision = &Decision{
			ID:     uuid.New().String(),
			RuleID: rule.ID,
		}
	}

	decision.Type = req.Decision.Type
	decision.Reason = req.Decision.Reason

	if err := s.decisionRepo.UpdateDecision(ctx, decision); err != nil {
		return nil, err
	}

	s.logger.WithContext(ctx).Info("更新规则成功", logger.NewField("rule_id", rule.ID))
	s.triggerReload(ctx)
	return rule, nil
}

func (s *RuleEngineService) DeleteRule(ctx context.Context, id string) error {
	if err := s.conditionRepo.DeleteConditionsByRuleID(ctx, id); err != nil {
		return err
	}
	if err := s.ruleRepo.DeleteRule(ctx, id); err != nil {
		return err
	}
	s.triggerReload(ctx)
	return nil
}

func (s *RuleEngineService) GetRules(ctx context.Context, filter *RuleFilter) ([]*Rule, int64, error) {
	return s.ruleRepo.ListRules(ctx, filter)
}

func (s *RuleEngineService) GetRuleByID(ctx context.Context, id string) (*Rule, error) {
	return s.ruleRepo.GetRuleByID(ctx, id)
}

func (s *RuleEngineService) EnableRule(ctx context.Context, id string) error {
	return s.ruleRepo.EnableRule(ctx, id)
}

func (s *RuleEngineService) DisableRule(ctx context.Context, id string) error {
	return s.ruleRepo.DisableRule(ctx, id)
}

func (s *RuleEngineService) CreateFeature(ctx context.Context, req *CreateFeatureRequest) (*Feature, error) {
	if req.Name == "" {
		return nil, errors.New("特征名称不能为空")
	}
	if req.Code == "" {
		return nil, errors.New("特征编码不能为空")
	}
	if req.Type == "" {
		return nil, errors.New("特征类型不能为空")
	}
	if req.ValueType == "" {
		return nil, errors.New("值类型不能为空")
	}

	feature := &Feature{
		ID:             uuid.New().String(),
		Name:           req.Name,
		Code:           req.Code,
		Description:    req.Description,
		Type:           req.Type,
		ValueType:      req.ValueType,
		Category:       req.Category,
		Enabled:        req.Enabled,
		FunctionName:   req.FunctionName,
		FunctionConfig: req.FunctionConfig,
	}

	if err := s.featureRepo.CreateFeature(ctx, feature); err != nil {
		return nil, err
	}

	for _, val := range req.Values {
		featureValue := &FeatureValue{
			ID:        uuid.New().String(),
			FeatureID: feature.ID,
			Value:     val.Value,
			Label:     val.Label,
			SortOrder: val.SortOrder,
			Enabled:   val.Enabled,
		}
		if err := s.featureValueRepo.CreateFeatureValue(ctx, featureValue); err != nil {
			s.logger.WithContext(ctx).Error("创建特征值失败", logger.NewField("error", err.Error()))
			return nil, err
		}
	}

	s.logger.WithContext(ctx).Info("创建特征成功", logger.NewField("feature_id", feature.ID))
	return feature, nil
}

func (s *RuleEngineService) UpdateFeature(ctx context.Context, req *UpdateFeatureRequest) (*Feature, error) {
	if req.ID == "" {
		return nil, errors.New("特征ID不能为空")
	}

	feature, err := s.featureRepo.GetFeatureByID(ctx, req.ID)
	if err != nil {
		return nil, err
	}

	feature.Name = req.Name
	feature.Code = req.Code
	feature.Type = req.Type
	feature.ValueType = req.ValueType
	feature.Category = req.Category
	feature.Enabled = req.Enabled
	feature.FunctionName = req.FunctionName
	feature.FunctionConfig = req.FunctionConfig
	feature.Description = req.Description

	if err := s.featureRepo.UpdateFeature(ctx, feature); err != nil {
		return nil, err
	}

	if err := s.featureValueRepo.DeleteFeatureValuesByFeatureID(ctx, req.ID); err != nil {
		return nil, err
	}

	for _, val := range req.Values {
		featureValue := &FeatureValue{
			ID:        uuid.New().String(),
			FeatureID: feature.ID,
			Value:     val.Value,
			Label:     val.Label,
			SortOrder: val.SortOrder,
			Enabled:   val.Enabled,
		}
		if err := s.featureValueRepo.CreateFeatureValue(ctx, featureValue); err != nil {
			s.logger.WithContext(ctx).Error("创建特征值失败", logger.NewField("error", err.Error()))
			return nil, err
		}
	}

	s.logger.WithContext(ctx).Info("更新特征成功", logger.NewField("feature_id", feature.ID))
	return feature, nil
}

func (s *RuleEngineService) DeleteFeature(ctx context.Context, id string) error {
	if err := s.featureValueRepo.DeleteFeatureValuesByFeatureID(ctx, id); err != nil {
		return err
	}
	return s.featureRepo.DeleteFeature(ctx, id)
}

func (s *RuleEngineService) GetFeatures(ctx context.Context, filter *FeatureFilter) ([]*Feature, int64, error) {
	return s.featureRepo.ListFeatures(ctx, filter)
}

func (s *RuleEngineService) GetFeatureByID(ctx context.Context, id string) (*Feature, error) {
	return s.featureRepo.GetFeatureByID(ctx, id)
}

func (s *RuleEngineService) GetFeatureByCode(ctx context.Context, code string) (*Feature, error) {
	return s.featureRepo.GetFeatureByCode(ctx, code)
}

func (s *RuleEngineService) EnableFeature(ctx context.Context, id string) error {
	return s.featureRepo.EnableFeature(ctx, id)
}

func (s *RuleEngineService) DisableFeature(ctx context.Context, id string) error {
	return s.featureRepo.DisableFeature(ctx, id)
}

func (s *RuleEngineService) EvaluateRule(ctx context.Context, ruleID string, data map[string]interface{}) (*RuleEvaluationResult, error) {
	rule, err := s.ruleRepo.GetRuleByID(ctx, ruleID)
	if err != nil {
		return nil, err
	}

	if !rule.Enabled {
		return &RuleEvaluationResult{
			RuleID:  rule.ID,
			Passed:  false,
			Message: "规则未启用",
		}, nil
	}

	result := &RuleEvaluationResult{
		RuleID:   rule.ID,
		RuleName: rule.Name,
		Passed:   false,
		Message:  "",
	}

	if len(rule.Conditions) == 0 {
		result.Passed = false
		result.Message = "规则没有条件"
		return result, nil
	}

	passed, err := s.evaluateConditions(ctx, rule.Conditions, data, LogicOpAnd)
	if err != nil {
		return nil, err
	}

	result.Passed = passed
	if passed {
		result.DecisionType = rule.Decision.Type
		result.DecisionReason = rule.Decision.Reason
		result.Message = "规则命中"
	} else {
		result.Message = "规则未命中"
	}

	return result, nil
}

func (s *RuleEngineService) evaluateConditions(ctx context.Context, conditions []Condition, data map[string]interface{}, defaultLogicOp string) (bool, error) {
	if len(conditions) == 0 {
		return true, nil
	}

	result := true
	currentLogicOp := defaultLogicOp

	for i, condition := range conditions {
		conditionPassed, err := s.evaluateCondition(ctx, &condition, data)
		if err != nil {
			return false, err
		}

		if i > 0 {
			currentLogicOp = condition.LogicOp
		}

		if currentLogicOp == LogicOpAnd {
			result = result && conditionPassed
			if !result {
				return false, nil
			}
		} else {
			result = result || conditionPassed
			if result {
				return true, nil
			}
		}
	}

	return result, nil
}

func (s *RuleEngineService) evaluateCondition(ctx context.Context, condition *Condition, data map[string]interface{}) (bool, error) {
	value, exists := data[condition.FeatureID]
	if !exists {
		return false, nil
	}

	feature, err := s.featureRepo.GetFeatureByID(ctx, condition.FeatureID)
	if err == nil && feature.FunctionName != "" {
		fn, fnExists := s.featureFunctionRegistry.Get(feature.FunctionName)
		if fnExists {
			functionConfig := make(map[string]interface{})
			if feature.FunctionConfig != nil {
				for k, v := range feature.FunctionConfig {
					functionConfig[k] = v
				}
			}

			fnInput := &featurefunction.FunctionInput{
				Config:      functionConfig,
				InvoiceData: data,
				InvoiceID:   fmt.Sprintf("%v", data["invoice_id"]),
			}

			fnOutput, err := fn.Execute(ctx, fnInput)
			if err != nil {
				s.logger.WithContext(ctx).Error("特征函数执行失败",
					logger.NewField("error", err.Error()),
					logger.NewField("function", feature.FunctionName))
				return false, nil
			}

			if fnOutput.Error != "" {
				s.logger.WithContext(ctx).Error("特征函数返回错误",
					logger.NewField("error", fnOutput.Error),
					logger.NewField("function", feature.FunctionName))
				return false, nil
			}

			value = fnOutput.Value
		}
	}

	switch condition.Operator {
	case OperatorEqual:
		return fmt.Sprintf("%v", value) == condition.Value, nil
	case OperatorNotEqual:
		return fmt.Sprintf("%v", value) != condition.Value, nil
	case OperatorGreaterThan:
		return s.compareNumbers(value, condition.Value, ">")
	case OperatorGreaterEqual:
		return s.compareNumbers(value, condition.Value, ">=")
	case OperatorLessThan:
		return s.compareNumbers(value, condition.Value, "<")
	case OperatorLessEqual:
		return s.compareNumbers(value, condition.Value, "<=")
	case OperatorContains:
		return s.containsString(value, condition.Value, true)
	case OperatorNotContains:
		return s.containsString(value, condition.Value, false)
	default:
		return false, errors.New("不支持的运算符: " + condition.Operator)
	}
}

func (s *RuleEngineService) compareNumbers(value interface{}, target string, operator string) (bool, error) {
	valueStr := fmt.Sprintf("%v", value)
	var num1, num2 float64
	_, err1 := fmt.Sscanf(valueStr, "%f", &num1)
	_, err2 := fmt.Sscanf(target, "%f", &num2)

	if err1 != nil || err2 != nil {
		return false, nil
	}

	switch operator {
	case ">":
		return num1 > num2, nil
	case ">=":
		return num1 >= num2, nil
	case "<":
		return num1 < num2, nil
	case "<=":
		return num1 <= num2, nil
	default:
		return false, errors.New("不支持的比较运算符: " + operator)
	}
}

func (s *RuleEngineService) containsString(value interface{}, target string, shouldContain bool) (bool, error) {
	valueStr := fmt.Sprintf("%v", value)
	if shouldContain {
		return len(valueStr) > 0 && len(target) > 0 &&
			(valueStr == target || len(valueStr) >= len(target) &&
				valueStr[:len(target)] == target), nil
	}
	return !(len(valueStr) > 0 && len(target) > 0 &&
		(valueStr == target || len(valueStr) >= len(target) &&
			valueStr[:len(target)] == target)), nil
}

func (s *RuleEngineService) ExecuteAllRules(ctx context.Context, data interface{}) ([]*RuleValidationResult, error) {
	if s.engine == nil {
		return nil, errors.New("规则引擎未初始化")
	}
	return s.engine.ExecuteAllRules(ctx, data)
}

type CreateRuleRequest struct {
	Name        string             `json:"name"`
	Description string             `json:"description"`
	Conditions  []ConditionRequest `json:"conditions"`
	Decision    DecisionRequest    `json:"decision"`
	Priority    int                `json:"priority"`
	Enabled     bool               `json:"enabled"`
}

type UpdateRuleRequest struct {
	ID          string             `json:"id"`
	Name        string             `json:"name"`
	Description string             `json:"description"`
	Conditions  []ConditionRequest `json:"conditions"`
	Decision    DecisionRequest    `json:"decision"`
	Priority    int                `json:"priority"`
	Enabled     bool               `json:"enabled"`
}

type ConditionRequest struct {
	FeatureID string `json:"feature_id"`
	Operator  string `json:"operator"`
	Value     string `json:"value"`
	LogicOp   string `json:"logic_op"`
	SortOrder int    `json:"sort_order"`
}

type DecisionRequest struct {
	Type   string `json:"type"`
	Reason string `json:"reason"`
}

type CreateFeatureRequest struct {
	Name           string                `json:"name"`
	Code           string                `json:"code"`
	Description    string                `json:"description"`
	Type           string                `json:"type"`
	ValueType      string                `json:"value_type"`
	Category       string                `json:"category"`
	Enabled        bool                  `json:"enabled"`
	FunctionName   string                `json:"function_name"`
	FunctionConfig FeatureConfig         `json:"function_config"`
	Values         []FeatureValueRequest `json:"values"`
}

type UpdateFeatureRequest struct {
	ID             string                `json:"id"`
	Name           string                `json:"name"`
	Code           string                `json:"code"`
	Description    string                `json:"description"`
	Type           string                `json:"type"`
	ValueType      string                `json:"value_type"`
	Category       string                `json:"category"`
	Enabled        bool                  `json:"enabled"`
	FunctionName   string                `json:"function_name"`
	FunctionConfig FeatureConfig         `json:"function_config"`
	Values         []FeatureValueRequest `json:"values"`
}

type FeatureValueRequest struct {
	Value     string `json:"value"`
	Label     string `json:"label"`
	SortOrder int    `json:"sort_order"`
	Enabled   bool   `json:"enabled"`
}

type RuleEvaluationResult struct {
	RuleID         string
	RuleName       string
	Passed         bool
	Message        string
	DecisionType   string
	DecisionReason string
}
