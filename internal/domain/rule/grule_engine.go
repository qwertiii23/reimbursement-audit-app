// grule_engine.go Grule引擎封装
// 功能点：
// 1. 封装Grule规则引擎
// 2. 规则定义和解析
// 3. 规则执行和结果处理
// 4. 规则库管理
// 5. 规则执行上下文管理
// 6. 规则性能监控

package rule

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"reimbursement-audit/internal/pkg/logger"

	"github.com/hyperjumptech/grule-rule-engine/ast"
	"github.com/hyperjumptech/grule-rule-engine/builder"
	"github.com/hyperjumptech/grule-rule-engine/engine"
	"github.com/hyperjumptech/grule-rule-engine/model"
	"github.com/hyperjumptech/grule-rule-engine/pkg"
)

// GRuleEngine Grule规则引擎结构体
type GRuleEngine struct {
	ruleLibrary      map[string]*ast.KnowledgeBase // 规则库
	knowledgeLibrary *ast.KnowledgeLibrary         // Grule知识库
	repository       Repository                    // 规则仓库接口
	logger           logger.Logger                 // 日志记录器
	mu               sync.RWMutex                  // 读写锁
	stats            map[string]*EngineRuleStats   // 规则执行统计
}

// EngineRuleStats 引擎规则执行统计
type EngineRuleStats struct {
	RuleID         string        `json:"rule_id"`
	ExecutionCount int           `json:"execution_count"`
	SuccessCount   int           `json:"success_count"`
	FailureCount   int           `json:"failure_count"`
	LastExecution  time.Time     `json:"last_execution"`
	AverageTime    time.Duration `json:"average_time"`
}

// NewGRuleEngine 创建Grule规则引擎实例
func NewGRuleEngine(repository Repository, log logger.Logger) *GRuleEngine {
	return &GRuleEngine{
		ruleLibrary:      make(map[string]*ast.KnowledgeBase),
		knowledgeLibrary: ast.NewKnowledgeLibrary(),
		repository:       repository,
		logger:           log,
		stats:            make(map[string]*EngineRuleStats),
	}
}

// Initialize 初始化引擎，加载数据库中启用的规则
func (e *GRuleEngine) Initialize(ctx context.Context) error {
	e.logger.WithContext(ctx).Info("初始化Grule规则引擎")

	// 从数据库获取所有启用的规则
	filter := &RuleFilter{
		Enabled: &[]bool{true}[0], // 获取启用的规则
		Size:    1000,             // 设置较大的页面大小以获取所有规则
	}

	rules, _, err := e.repository.ListRules(ctx, filter)
	if err != nil {
		e.logger.WithContext(ctx).Error("获取启用规则失败",
			logger.NewField("error", err.Error()))
		return fmt.Errorf("获取启用规则失败: %w", err)
	}

	e.logger.WithContext(ctx).Info("开始加载规则到引擎",
		logger.NewField("规则数量", len(rules)))

	// 加载所有启用的规则
	for _, rule := range rules {
		if err := e.LoadRule(ctx, rule); err != nil {
			e.logger.WithContext(ctx).Error("加载规则失败",
				logger.NewField("规则ID", rule.ID),
				logger.NewField("规则名称", rule.Name),
				logger.NewField("error", err.Error()))
			// 继续加载其他规则，不中断初始化过程
			continue
		}
	}

	e.logger.WithContext(ctx).Info("Grule规则引擎初始化完成")
	return nil
}

// LoadRule 加载规则
func (e *GRuleEngine) LoadRule(ctx context.Context, rule *Rule) error {
	if rule == nil {
		return errors.New("规则不能为空")
	}

	if !rule.Enabled {
		e.logger.WithContext(ctx).Warn("尝试加载未启用的规则",
			logger.NewField("规则ID", rule.ID))
		return nil
	}

	e.mu.Lock()
	defer e.mu.Unlock()

	// 验证规则语法
	if err := e.ValidateRule(rule.Definition); err != nil {
		e.logger.WithContext(ctx).Error("规则语法验证失败",
			logger.NewField("规则ID", rule.ID),
			logger.NewField("error", err.Error()))
		return fmt.Errorf("规则语法验证失败: %w", err)
	}

	// 使用builder编译规则
	ruleBuilder := builder.NewRuleBuilder(e.knowledgeLibrary)

	// 创建规则资源
	ruleResource := pkg.NewBytesResource([]byte(rule.Definition))

	// 解析规则定义
	err := ruleBuilder.BuildRuleFromResource(rule.RuleCode, "1.0", ruleResource)
	if err != nil {
		e.logger.WithContext(ctx).Error("编译规则失败",
			logger.NewField("规则ID", rule.ID),
			logger.NewField("error", err.Error()))
		return fmt.Errorf("编译规则失败: %w", err)
	}

	// 获取知识库实例
	knowledgeBase := e.knowledgeLibrary.GetKnowledgeBase(rule.RuleCode, "1.0")
	if knowledgeBase == nil {
		e.logger.WithContext(ctx).Error("获取知识库实例失败",
			logger.NewField("规则ID", rule.ID))
		return fmt.Errorf("获取知识库实例失败")
	}

	// 保存知识库到本地规则库
	e.ruleLibrary[rule.ID] = knowledgeBase

	// 初始化统计信息
	e.stats[rule.ID] = &EngineRuleStats{
		RuleID:         rule.ID,
		ExecutionCount: 0,
		SuccessCount:   0,
		FailureCount:   0,
	}

	e.logger.WithContext(ctx).Info("规则加载成功",
		logger.NewField("规则ID", rule.ID),
		logger.NewField("规则名称", rule.Name))

	return nil
}

// UnloadRule 卸载规则
func (e *GRuleEngine) UnloadRule(ctx context.Context, ruleID string) error {
	if ruleID == "" {
		return errors.New("规则ID不能为空")
	}

	e.mu.Lock()
	defer e.mu.Unlock()

	if _, exists := e.ruleLibrary[ruleID]; !exists {
		return fmt.Errorf("规则不存在: %s", ruleID)
	}

	// 从规则库中移除
	delete(e.ruleLibrary, ruleID)

	// 从统计信息中移除
	delete(e.stats, ruleID)

	e.logger.WithContext(ctx).Info("规则卸载成功",
		logger.NewField("规则ID", ruleID))

	return nil
}

// ExecuteRule 执行单个规则
func (e *GRuleEngine) ExecuteRule(ctx context.Context, ruleID string, data interface{}) (*RuleValidationResult, error) {
	if ruleID == "" {
		return nil, errors.New("规则ID不能为空")
	}

	e.mu.RLock()
	knowledgeBase, exists := e.ruleLibrary[ruleID]
	e.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("规则不存在: %s", ruleID)
	}

	// 记录执行开始时间
	startTime := time.Now()

	// 更新统计信息
	e.updateStatistics(ruleID, true, startTime, false)

	// 创建数据上下文
	dataContext := ast.NewDataContext()
	err := dataContext.Add("data", data)
	if err != nil {
		e.updateStatistics(ruleID, false, startTime, true)
		e.logger.WithContext(ctx).Error("创建数据上下文失败",
			logger.NewField("规则ID", ruleID),
			logger.NewField("error", err.Error()))
		return nil, fmt.Errorf("创建数据上下文失败: %w", err)
	}

	// 创建结果对象
	result := &RuleValidationResult{
		RuleID:    ruleID,
		Passed:    true,
		Message:   "规则执行初始化",
		Timestamp: time.Now(),
	}

	// 添加结果对象到上下文
	err = dataContext.Add("result", result)
	if err != nil {
		e.updateStatistics(ruleID, false, startTime, true)
		e.logger.WithContext(ctx).Error("添加结果对象到上下文失败",
			logger.NewField("规则ID", ruleID),
			logger.NewField("error", err.Error()))
		return nil, fmt.Errorf("添加结果对象到上下文失败: %w", err)
	}

	// 创建引擎实例
	gruleEngine := engine.NewGruleEngine()

	// 执行规则
	err = gruleEngine.Execute(dataContext, knowledgeBase)
	executionTime := time.Since(startTime)

	if err != nil {
		e.updateStatistics(ruleID, false, startTime, true)
		e.logger.WithContext(ctx).Error("规则执行失败",
			logger.NewField("规则ID", ruleID),
			logger.NewField("执行时间", executionTime.String()),
			logger.NewField("error", err.Error()))

		return &RuleValidationResult{
			RuleID:    ruleID,
			Passed:    false,
			Message:   fmt.Sprintf("规则执行失败: %s", err.Error()),
			Timestamp: time.Now(),
		}, nil
	}

	// 更新统计信息
	e.updateStatistics(ruleID, false, startTime, false)

	// 从上下文中获取结果
	resultNode := dataContext.Get("result")
	if resultNode != nil {
		// 尝试将结果转换为 RuleValidationResult
		if resultVal, ok := resultNode.(model.ValueNode); ok {
			if resultObj, err := resultVal.GetValue(); err == nil {
				if res, ok := resultObj.Interface().(*RuleValidationResult); ok {
					result = res
				}
			}
		}
	} else {
		// 如果获取结果失败，使用默认结果
		e.logger.WithContext(ctx).Warn("获取规则执行结果失败，使用默认结果",
			logger.NewField("规则ID", ruleID))
		result = &RuleValidationResult{
			RuleID:    ruleID,
			Passed:    true,
			Message:   "规则执行成功",
			Timestamp: time.Now(),
		}
	}

	e.logger.WithContext(ctx).Info("规则执行成功",
		logger.NewField("规则ID", ruleID),
		logger.NewField("执行时间", executionTime.String()),
		logger.NewField("结果", result.Passed))

	return result, nil
}

// ExecuteRuleWithDataContext 执行单个规则，支持自定义数据上下文
func (e *GRuleEngine) ExecuteRuleWithDataContext(ctx context.Context, ruleID string, dataContext map[string]interface{}) (*RuleValidationResult, error) {
	if ruleID == "" {
		return nil, errors.New("规则ID不能为空")
	}

	e.mu.RLock()
	knowledgeBase, exists := e.ruleLibrary[ruleID]
	e.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("规则不存在: %s", ruleID)
	}

	// 记录执行开始时间
	startTime := time.Now()

	// 创建数据上下文
	dc := ast.NewDataContext()

	// 添加所有数据上下文项
	for key, value := range dataContext {
		err := dc.Add(key, value)
		if err != nil {
			e.logger.WithContext(ctx).Error("添加数据上下文项失败",
				logger.NewField("规则ID", ruleID),
				logger.NewField("上下文键", key),
				logger.NewField("error", err.Error()))
			return nil, fmt.Errorf("添加数据上下文项失败: %w", err)
		}
	}

	// 创建结果对象
	result := &RuleValidationResult{
		RuleID:    ruleID,
		Passed:    true,
		Message:   "规则执行初始化",
		Timestamp: time.Now(),
	}

	// 添加结果对象到上下文
	err := dc.Add("result", result)
	if err != nil {
		e.logger.WithContext(ctx).Error("添加结果对象到上下文失败",
			logger.NewField("规则ID", ruleID),
			logger.NewField("error", err.Error()))
		return nil, fmt.Errorf("添加结果对象到上下文失败: %w", err)
	}

	// 创建引擎实例
	gruleEngine := engine.NewGruleEngine()

	// 执行规则
	err = gruleEngine.Execute(dc, knowledgeBase)
	executionTime := time.Since(startTime)

	if err != nil {
		e.logger.WithContext(ctx).Error("规则执行失败",
			logger.NewField("规则ID", ruleID),
			logger.NewField("执行时间", executionTime.String()),
			logger.NewField("error", err.Error()))

		return &RuleValidationResult{
			RuleID:    ruleID,
			Passed:    false,
			Message:   fmt.Sprintf("规则执行失败: %s", err.Error()),
			Timestamp: time.Now(),
		}, nil
	}

	// 从上下文中获取结果
	resultNode := dc.Get("result")
	if resultNode != nil {
		// 尝试将结果转换为 RuleValidationResult
		if resultVal, ok := resultNode.(model.ValueNode); ok {
			if resultObj, err := resultVal.GetValue(); err == nil {
				if res, ok := resultObj.Interface().(*RuleValidationResult); ok {
					result = res
				}
			}
		}
	} else {
		// 如果获取结果失败，使用默认结果
		e.logger.WithContext(ctx).Warn("获取规则执行结果失败，使用默认结果",
			logger.NewField("规则ID", ruleID))
		result = &RuleValidationResult{
			RuleID:    ruleID,
			Passed:    true,
			Message:   "规则执行成功",
			Timestamp: time.Now(),
		}
	}

	e.logger.WithContext(ctx).Info("规则执行成功",
		logger.NewField("规则ID", ruleID),
		logger.NewField("执行时间", executionTime.String()),
		logger.NewField("结果", result.Passed))

	return result, nil
}

// ExecuteRules 执行多个规则
func (e *GRuleEngine) ExecuteRules(ctx context.Context, ruleIDs []string, data interface{}) ([]*RuleValidationResult, error) {
	if len(ruleIDs) == 0 {
		return nil, errors.New("规则ID列表不能为空")
	}

	results := make([]*RuleValidationResult, 0, len(ruleIDs))

	for _, ruleID := range ruleIDs {
		result, err := e.ExecuteRule(ctx, ruleID, data)
		if err != nil {
			e.logger.WithContext(ctx).Error("执行规则失败",
				logger.NewField("规则ID", ruleID),
				logger.NewField("error", err.Error()))
			// 继续执行其他规则
			results = append(results, &RuleValidationResult{
				RuleID:    ruleID,
				Passed:    false,
				Message:   fmt.Sprintf("规则执行失败: %s", err.Error()),
				Timestamp: time.Now(),
			})
			continue
		}
		results = append(results, result)
	}

	return results, nil
}

// ExecuteAllRules 执行所有规则
func (e *GRuleEngine) ExecuteAllRules(ctx context.Context, data interface{}) ([]*RuleValidationResult, error) {
	e.mu.RLock()
	ruleIDs := make([]string, 0, len(e.ruleLibrary))
	for ruleID := range e.ruleLibrary {
		ruleIDs = append(ruleIDs, ruleID)
	}
	e.mu.RUnlock()

	if len(ruleIDs) == 0 {
		return []*RuleValidationResult{}, nil
	}

	return e.ExecuteRules(ctx, ruleIDs, data)
}

// ValidateRule 验证规则语法
func (e *GRuleEngine) ValidateRule(ruleDefinition string) error {
	if ruleDefinition == "" {
		return errors.New("规则定义不能为空")
	}

	// 创建临时知识库进行验证
	tempKnowledgeLibrary := ast.NewKnowledgeLibrary()
	ruleBuilder := builder.NewRuleBuilder(tempKnowledgeLibrary)

	// 创建规则资源
	ruleResource := pkg.NewBytesResource([]byte(ruleDefinition))

	// 尝试构建规则
	err := ruleBuilder.BuildRuleFromResource("validation", "1.0", ruleResource)
	if err != nil {
		return fmt.Errorf("规则语法错误: %w", err)
	}

	return nil
}

// ParseRuleDefinition 解析规则定义
func (e *GRuleEngine) ParseRuleDefinition(ruleDefinition string) (*Rule, error) {
	if ruleDefinition == "" {
		return nil, errors.New("规则定义不能为空")
	}

	// 这里可以实现更复杂的解析逻辑，目前只返回基本信息
	// 在实际应用中，可能需要解析规则定义中的元数据
	rule := &Rule{
		ID:         e.generateUUID(),
		Definition: ruleDefinition,
		Enabled:    true,
	}

	return rule, nil
}

// GetRuleLibrary 获取规则库
func (e *GRuleEngine) GetRuleLibrary() map[string]*ast.KnowledgeBase {
	e.mu.RLock()
	defer e.mu.RUnlock()

	// 返回规则库的副本
	result := make(map[string]*ast.KnowledgeBase)
	for k, v := range e.ruleLibrary {
		result[k] = v
	}

	return result
}

// ClearRuleLibrary 清空规则库
func (e *GRuleEngine) ClearRuleLibrary() {
	e.mu.Lock()
	defer e.mu.Unlock()

	e.ruleLibrary = make(map[string]*ast.KnowledgeBase)
	e.knowledgeLibrary = ast.NewKnowledgeLibrary()
	e.stats = make(map[string]*EngineRuleStats)
}

// ReloadRuleLibrary 重新加载规则库
func (e *GRuleEngine) ReloadRuleLibrary(ctx context.Context, rules []*Rule) error {
	e.logger.WithContext(ctx).Info("重新加载规则库")

	// 清空当前规则库
	e.ClearRuleLibrary()

	// 重新加载所有规则
	for _, rule := range rules {
		if err := e.LoadRule(ctx, rule); err != nil {
			e.logger.WithContext(ctx).Error("重新加载规则失败",
				logger.NewField("规则ID", rule.ID),
				logger.NewField("error", err.Error()))
			// 继续加载其他规则
			continue
		}
	}

	e.logger.WithContext(ctx).Info("规则库重新加载完成")
	return nil
}

// ReloadRulesFromDatabase 从数据库重新加载规则
func (e *GRuleEngine) ReloadRulesFromDatabase(ctx context.Context) error {
	// 从数据库获取所有启用的规则
	filter := &RuleFilter{
		Enabled: &[]bool{true}[0], // 获取启用的规则
		Size:    1000,             // 设置较大的页面大小以获取所有规则
	}

	rules, _, err := e.repository.ListRules(ctx, filter)
	if err != nil {
		e.logger.WithContext(ctx).Error("获取启用规则失败",
			logger.NewField("error", err.Error()))
		return fmt.Errorf("获取启用规则失败: %w", err)
	}

	return e.ReloadRuleLibrary(ctx, rules)
}

// GetLoadedRules 获取已加载的规则列表
func (e *GRuleEngine) GetLoadedRules() []string {
	e.mu.RLock()
	defer e.mu.RUnlock()

	ruleIDs := make([]string, 0, len(e.ruleLibrary))
	for ruleID := range e.ruleLibrary {
		ruleIDs = append(ruleIDs, ruleID)
	}

	return ruleIDs
}

// IsRuleLoaded 检查规则是否已加载
func (e *GRuleEngine) IsRuleLoaded(ruleID string) bool {
	e.mu.RLock()
	defer e.mu.RUnlock()

	_, exists := e.ruleLibrary[ruleID]
	return exists
}

// GetRuleStatistics 获取规则执行统计信息
func (e *GRuleEngine) GetRuleStatistics() map[string]*EngineRuleStats {
	e.mu.RLock()
	defer e.mu.RUnlock()

	// 返回统计信息的副本
	result := make(map[string]*EngineRuleStats)
	for k, v := range e.stats {
		result[k] = v
	}

	return result
}

// CreateExecutionContext 创建执行上下文
func (e *GRuleEngine) CreateExecutionContext(data interface{}) ast.IDataContext {
	dataContext := ast.NewDataContext()

	// 添加数据到上下文
	err := dataContext.Add("data", data)
	if err != nil {
		e.logger.Error("创建执行上下文失败",
			logger.NewField("error", err.Error()))
		return nil
	}

	// 添加结果对象到上下文
	result := &RuleValidationResult{
		Passed:  true,
		Message: "规则执行初始化",
	}
	err = dataContext.Add("result", result)
	if err != nil {
		e.logger.Error("添加结果对象到上下文失败",
			logger.NewField("error", err.Error()))
		return nil
	}

	return dataContext
}

// ResetStatistics 重置统计信息
func (e *GRuleEngine) ResetStatistics() {
	e.mu.Lock()
	defer e.mu.Unlock()

	for _, stat := range e.stats {
		stat.ExecutionCount = 0
		stat.SuccessCount = 0
		stat.FailureCount = 0
		stat.LastExecution = time.Time{}
		stat.AverageTime = 0
	}
}

// updateStatistics 更新规则执行统计信息
func (e *GRuleEngine) updateStatistics(ruleID string, isStart bool, startTime time.Time, isError bool) {
	e.mu.Lock()
	defer e.mu.Unlock()

	stat, exists := e.stats[ruleID]
	if !exists {
		// 如果统计信息不存在，创建新的
		stat = &EngineRuleStats{
			RuleID: ruleID,
		}
		e.stats[ruleID] = stat
	}

	if isStart {
		// 记录开始时间
		stat.LastExecution = startTime
		stat.ExecutionCount++
	} else {
		// 计算执行时间
		executionTime := time.Since(startTime)

		// 更新平均执行时间
		if stat.ExecutionCount == 1 {
			stat.AverageTime = executionTime
		} else {
			// 计算新的平均时间
			totalTime := time.Duration(stat.ExecutionCount-1) * stat.AverageTime
			stat.AverageTime = (totalTime + executionTime) / time.Duration(stat.ExecutionCount)
		}

		// 更新成功/失败计数
		if isError {
			stat.FailureCount++
		} else {
			stat.SuccessCount++
		}
	}
}

// generateUUID 生成UUID
func (e *GRuleEngine) generateUUID() string {
	// 简单的UUID生成实现，实际项目中可以使用更完善的库
	return fmt.Sprintf("%d", time.Now().UnixNano())
}
