// engine.go 规则引擎实现
// 功能点：
// 1. 规则执行
// 2. 规则加载和卸载
// 3. 规则统计

package ruleengine

import (
	"context"
	"fmt"
	"sync"
	"time"

	"reimbursement-audit/internal/domain/featurefunction"
	"reimbursement-audit/internal/pkg/logger"

	"github.com/google/uuid"
)

// Engine 规则引擎结构体
type Engine struct {
	ruleRepository          RuleRepository
	featureRepository       FeatureRepository
	logger                  logger.Logger
	mu                      sync.RWMutex
	loadedRules             map[string]*Rule
	stats                   map[string]*EngineStats
	featureFunctionRegistry *featurefunction.FunctionRegistry
}

// EngineStats 引擎统计
type EngineStats struct {
	RuleID         string        `json:"rule_id"`
	ExecutionCount int           `json:"execution_count"`
	SuccessCount   int           `json:"success_count"`
	FailureCount   int           `json:"failure_count"`
	LastExecution  time.Time     `json:"last_execution"`
	AverageTime    time.Duration `json:"average_time"`
}

// NewEngine 创建规则引擎实例
func NewEngine(repository RuleRepository, featureRepository FeatureRepository, log logger.Logger, featureFunctionRegistry *featurefunction.FunctionRegistry) *Engine {
	return &Engine{
		ruleRepository:          repository,
		featureRepository:       featureRepository,
		logger:                  log,
		loadedRules:             make(map[string]*Rule),
		stats:                   make(map[string]*EngineStats),
		featureFunctionRegistry: featureFunctionRegistry,
	}
}

// SetRepository 设置规则仓储
func (e *Engine) SetRepository(repository RuleRepository) {
	e.ruleRepository = repository
}

// Initialize 初始化引擎
func (e *Engine) Initialize(ctx context.Context) error {
	e.logger.WithContext(ctx).Info("初始化规则引擎")

	filter := &RuleFilter{
		Enabled: &[]bool{true}[0],
		Size:    1000,
	}

	rules, _, err := e.ruleRepository.ListRules(ctx, filter)
	if err != nil {
		return fmt.Errorf("获取启用规则失败: %w", err)
	}

	for _, rule := range rules {
		e.loadedRules[rule.ID] = rule
		e.stats[rule.ID] = &EngineStats{
			RuleID: rule.ID,
		}
	}

	e.logger.WithContext(ctx).Info("规则引擎初始化完成",
		logger.NewField("规则数量", len(rules)))

	return nil
}

// Reload 重新加载规则
func (e *Engine) Reload(ctx context.Context) error {
	e.logger.WithContext(ctx).Info("重新加载规则引擎")

	e.mu.Lock()
	e.loadedRules = make(map[string]*Rule)
	e.stats = make(map[string]*EngineStats)
	e.mu.Unlock()

	return e.Initialize(ctx)
}

// ExecuteAllRules 执行所有规则
func (e *Engine) ExecuteAllRules(ctx context.Context, data interface{}) ([]*RuleValidationResult, error) {
	e.mu.RLock()
	rules := make([]*Rule, 0, len(e.loadedRules))
	for _, rule := range e.loadedRules {
		rules = append(rules, rule)
	}
	e.mu.RUnlock()

	results := make([]*RuleValidationResult, 0, len(rules))

	for _, rule := range rules {
		result := e.evaluateRule(ctx, rule, data)
		results = append(results, result)
	}

	return results, nil
}

// evaluateRule 评估规则
func (e *Engine) evaluateRule(ctx context.Context, rule *Rule, data interface{}) *RuleValidationResult {
	startTime := time.Now()

	result := &RuleValidationResult{
		RuleID:    rule.ID,
		RuleName:  rule.Name,
		RuleType:  "rule_engine",
		Passed:    true,
		Message:   "规则执行成功",
		Timestamp: time.Now(),
	}

	if len(rule.Conditions) == 0 {
		result.Passed = false
		result.Message = "规则没有条件"
		return result
	}

	passed := e.evaluateConditions(ctx, rule.Conditions, data)
	result.Passed = passed

	if passed {
		result.Message = fmt.Sprintf("规则命中: %s", rule.Decision.Reason)
	} else {
		result.Message = "规则未命中"
	}

	executionTime := time.Since(startTime)
	result.ExecutionTime = executionTime.Milliseconds()

	e.updateStats(rule.ID, passed, executionTime)

	return result
}

// evaluateConditions 评估条件
func (e *Engine) evaluateConditions(ctx context.Context, conditions []Condition, data interface{}) bool {
	if len(conditions) == 0 {
		return true
	}

	dataMap, ok := data.(map[string]interface{})
	if !ok {
		return false
	}

	// 先计算所有需要的特征值
	for _, condition := range conditions {
		if _, exists := dataMap[condition.FeatureID]; !exists {
			// 如果特征值不存在，尝试调用特征函数计算
			if e.featureFunctionRegistry != nil && e.featureRepository != nil {
				// 通过特征ID查找特征信息
				feature, err := e.featureRepository.GetFeatureByID(ctx, condition.FeatureID)
				if err != nil {
					e.logger.WithContext(ctx).Error("获取特征失败",
						logger.NewField("error", err.Error()),
						logger.NewField("feature_id", condition.FeatureID))
				} else if feature == nil {
					e.logger.WithContext(ctx).Error("特征不存在",
						logger.NewField("feature_id", condition.FeatureID))
				} else if feature.FunctionName == "" {
					// 特征没有配置函数，这是正常的，跳过处理
					e.logger.WithContext(ctx).Debug("特征没有配置函数，跳过处理",
						logger.NewField("feature_id", condition.FeatureID))
				} else {
					// 使用特征函数名称调用特征函数
					if featureFunc, exists := e.featureFunctionRegistry.Get(feature.FunctionName); exists {
						input := &featurefunction.FunctionInput{
							InvoiceData: dataMap,
							Config:      feature.FunctionConfig,
						}
						output, err := featureFunc.Execute(ctx, input)
						if err != nil {
							e.logger.WithContext(ctx).Error("特征函数执行失败",
								logger.NewField("error", err.Error()),
								logger.NewField("function", feature.FunctionName),
								logger.NewField("feature_id", condition.FeatureID))
						} else if output.Error != "" {
							e.logger.WithContext(ctx).Error("特征函数返回错误",
								logger.NewField("error", output.Error),
								logger.NewField("function", feature.FunctionName),
								logger.NewField("feature_id", condition.FeatureID))
						} else {
							dataMap[condition.FeatureID] = output.Value
							e.logger.WithContext(ctx).Info("特征函数执行成功",
								logger.NewField("function", feature.FunctionName),
								logger.NewField("feature_id", condition.FeatureID),
								logger.NewField("value", output.Value))
						}
					} else {
						e.logger.WithContext(ctx).Error("特征函数不存在",
							logger.NewField("function", feature.FunctionName),
							logger.NewField("feature_id", condition.FeatureID))
					}
				}
			}
		}
	}

	for _, condition := range conditions {
		if !e.evaluateCondition(ctx, condition, dataMap) {
			return false
		}
	}

	return true
}

// evaluateCondition 评估单个条件
func (e *Engine) evaluateCondition(ctx context.Context, condition Condition, data interface{}) bool {
	dataMap, ok := data.(map[string]interface{})
	if !ok {
		return false
	}

	value, exists := dataMap[condition.FeatureID]
	if !exists {
		return false
	}

	// 如果值是字符串"true"或"false"，转换为布尔值
	if strValue, ok := value.(string); ok {
		if strValue == "true" {
			value = true
		} else if strValue == "false" {
			value = false
		}
	}

	switch condition.Operator {
	case OperatorEqual:
		return fmt.Sprintf("%v", value) == condition.Value
	case OperatorNotEqual:
		return fmt.Sprintf("%v", value) != condition.Value
	case OperatorGreaterThan:
		return e.compareNumbers(value, condition.Value, ">")
	case OperatorGreaterEqual:
		return e.compareNumbers(value, condition.Value, ">=")
	case OperatorLessThan:
		return e.compareNumbers(value, condition.Value, "<")
	case OperatorLessEqual:
		return e.compareNumbers(value, condition.Value, "<=")
	case OperatorContains:
		return e.containsString(value, condition.Value, true)
	case OperatorNotContains:
		return e.containsString(value, condition.Value, false)
	default:
		return false
	}
}

// compareNumbers 比较数字
func (e *Engine) compareNumbers(value interface{}, target string, operator string) bool {
	valueStr := fmt.Sprintf("%v", value)
	targetStr := target

	var num1, num2 float64
	_, err1 := fmt.Sscanf(valueStr, "%f", &num1)
	_, err2 := fmt.Sscanf(targetStr, "%f", &num2)

	if err1 != nil || err2 != nil {
		return false
	}

	switch operator {
	case ">":
		return num1 > num2
	case ">=":
		return num1 >= num2
	case "<":
		return num1 < num2
	case "<=":
		return num1 <= num2
	default:
		return false
	}
}

// containsString 检查字符串包含
func (e *Engine) containsString(value interface{}, target string, shouldContain bool) bool {
	valueStr := fmt.Sprintf("%v", value)
	if shouldContain {
		return len(valueStr) > 0 && len(target) > 0 &&
			(valueStr == target || len(valueStr) >= len(target) &&
				valueStr[:len(target)] == target)
	}
	return !(len(valueStr) > 0 && len(target) > 0 &&
		(valueStr == target || len(valueStr) >= len(target) &&
			valueStr[:len(target)] == target))
}

// updateStats 更新统计信息
func (e *Engine) updateStats(ruleID string, passed bool, executionTime time.Duration) {
	e.mu.Lock()
	defer e.mu.Unlock()

	stat, exists := e.stats[ruleID]
	if !exists {
		return
	}

	stat.ExecutionCount++
	stat.LastExecution = time.Now()

	if passed {
		stat.SuccessCount++
	} else {
		stat.FailureCount++
	}

	totalTime := stat.AverageTime * time.Duration(stat.ExecutionCount-1)
	stat.AverageTime = (totalTime + executionTime) / time.Duration(stat.ExecutionCount)
}

// GetStats 获取统计信息
func (e *Engine) GetStats() map[string]*EngineStats {
	e.mu.RLock()
	defer e.mu.RUnlock()

	result := make(map[string]*EngineStats)
	for k, v := range e.stats {
		result[k] = v
	}

	return result
}

// generateUUID 生成UUID
func (e *Engine) generateUUID() string {
	return uuid.New().String()
}
