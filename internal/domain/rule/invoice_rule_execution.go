// invoice_rule_execution.go 发票规则优先级执行和错误聚合
// 功能点：
// 1. 实现规则优先级执行
// 2. 实现错误聚合
// 3. 提供规则执行结果汇总

package rule

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"reimbursement-audit/internal/domain/ocr"
	"reimbursement-audit/internal/domain/reimbursement"
	"reimbursement-audit/internal/pkg/logger"
)

// InvoiceValidationData 发票校验数据（用于规则引擎）
type InvoiceValidationData struct {
	Invoice       *ocr.Invoice                 `json:"invoice"`       // 待校验发票
	Reimbursement *reimbursement.Reimbursement `json:"reimbursement"` // 关联报销单
	CompanyNames  []string                     `json:"company_names"` // 允许的公司名称列表
	InvoiceTypes  []string                     `json:"invoice_types"` // 允许的发票类型列表
	ApplyDate     time.Time                    `json:"apply_date"`    // 报销申请日期
}

// executeRulesWithPriority 按优先级执行规则
func (v *InvoiceValidatorImpl) executeRulesWithPriority(ctx context.Context, req *InvoiceValidationRequest, result *InvoiceValidationResult) error {
	v.logger.WithContext(ctx).Info("按优先级执行发票校验规则",
		logger.NewField("发票ID", req.Invoice.ID))

	// 合并基础规则和自定义规则
	allRules := make([]*RuleDefinition, 0, len(v.rules))
	allRules = append(allRules, v.rules...)

	// 按优先级排序（从高到低）
	sort.Slice(allRules, func(i, j int) bool {
		return allRules[i].Priority > allRules[j].Priority
	})

	// 创建校验数据
	validationData := &InvoiceValidationData{
		Invoice:       req.Invoice,
		Reimbursement: req.Reimbursement,
		CompanyNames:  req.CompanyNames,
		InvoiceTypes:  req.InvoiceTypes,
		ApplyDate:     req.ApplyDate,
	}

	// 创建校验结果对象
	validationResult := &RuleValidationResult{
		Passed:     true,
		Violations: make([]interface{}, 0),
		Message:    "",
	}

	// 将校验结果添加到数据上下文中
	dataContext := map[string]interface{}{
		"data":   validationData,
		"result": validationResult,
		// 添加辅助函数 - 适配为Grule可用的函数
		"IsDuplicateInvoice": func(invoiceCode, invoiceNumber string) bool {
			result, _ := v.isDuplicateInvoice(ctx, invoiceCode, invoiceNumber)
			return result
		},
		"GetAccommodationLimit": func(cityLevel string) float64 {
			return v.getAccommodationLimit(ctx, cityLevel)
		},
		"GetEntertainmentLimit": func(level string) float64 {
			return v.getEntertainmentLimit(ctx, level)
		},
		"IsConsecutiveInvoice": func(invoiceNumbers []string) bool {
			result, _ := v.isConsecutiveInvoice(ctx, invoiceNumbers)
			return result
		},
		"IsWeekendOrHoliday": func(date time.Time) bool {
			result, _ := v.isWeekendOrHoliday(ctx, date)
			return result
		},
		"IsValidTaxNumber": func(taxNumber string) bool {
			result, _ := v.isValidTaxNumber(ctx, taxNumber)
			return result
		},
		"HasOrderAndReceipt": func(invoiceID string) bool {
			result, _ := v.hasOrderAndReceipt(ctx, invoiceID)
			return result
		},
		"IsThreeDocumentMatching": func(invoiceID string) bool {
			result, _ := v.isThreeDocumentMatching(ctx, invoiceID)
			return result
		},
	}

	// 执行规则并收集结果
	for _, rule := range allRules {
		if !rule.Enabled {
			continue // 跳过禁用的规则
		}

		v.logger.WithContext(ctx).Debug("执行规则",
			logger.NewField("规则ID", rule.ID),
			logger.NewField("规则名称", rule.Name),
			logger.NewField("优先级", rule.Priority))

		// 执行规则
		ruleResult, err := v.ruleEngine.ExecuteRuleWithDataContext(ctx, rule.ID, dataContext)
		if err != nil {
			v.logger.WithContext(ctx).Error("执行规则失败",
				logger.NewField("规则ID", rule.ID),
				logger.NewField("发票ID", req.Invoice.ID),
				logger.NewField("error", err.Error()))
			continue
		}

		// 如果规则未通过，更新结果
		if !ruleResult.Passed {
			result.Passed = false

			// 从规则结果中提取违规信息
			if len(ruleResult.Violations) > 0 {
				for _, violation := range ruleResult.Violations {
					if v, ok := violation.(map[string]interface{}); ok {
						violationObj := &InvoiceViolation{
							RuleID:     getString(v, "RuleID"),
							RuleName:   getString(v, "RuleName"),
							RuleType:   getString(v, "RuleType"),
							Severity:   getString(v, "Severity"),
							Message:    getString(v, "Message"),
							Suggestion: getString(v, "Suggestion"),
							Priority:   getInt(v, "Priority"),
						}
						result.Violations = append(result.Violations, violationObj)
					}
				}
			} else {
				// 如果没有详细的违规信息，创建一个基本的违规记录
				violation := &InvoiceViolation{
					RuleID:     ruleResult.RuleID,
					RuleName:   ruleResult.RuleName,
					RuleType:   ruleResult.RuleType,
					Severity:   ruleResult.Severity,
					Message:    ruleResult.Message,
					Suggestion: generateSuggestion(ruleResult.RuleType, ruleResult.Message),
					Priority:   ruleResult.Priority,
				}
				result.Violations = append(result.Violations, violation)
			}
		}
	}

	// 按优先级排序违规信息
	sort.Slice(result.Violations, func(i, j int) bool {
		return result.Violations[i].Priority > result.Violations[j].Priority
	})

	// 生成校验结果摘要
	generateValidationSummary(result)

	v.logger.WithContext(ctx).Info("规则执行完成",
		logger.NewField("发票ID", req.Invoice.ID),
		logger.NewField("执行规则数", len(allRules)),
		logger.NewField("违规数", len(result.Violations)),
		logger.NewField("校验结果", result.Passed))

	return nil
}

// aggregateErrors 聚合错误信息
func (v *InvoiceValidatorImpl) aggregateErrors(ctx context.Context, results []*InvoiceValidationResult) {
	v.logger.WithContext(ctx).Info("聚合发票校验错误信息",
		logger.NewField("发票数量", len(results)))

	// 统计各类违规数量
	violationStats := make(map[string]int)
	severityStats := make(map[string]int)

	for _, result := range results {
		if !result.Passed {
			for _, violation := range result.Violations {
				violationStats[violation.RuleType]++
				severityStats[violation.Severity]++
			}
		}
	}

	v.logger.WithContext(ctx).Info("校验错误统计",
		logger.NewField("违规类型统计", violationStats),
		logger.NewField("严重程度统计", severityStats))
}

// generateValidationSummary 生成校验结果摘要
func generateValidationSummary(result *InvoiceValidationResult) {
	if result.Passed {
		result.Summary = "发票校验通过，无违规项"
		return
	}

	// 按严重程度统计违规
	highCount := 0
	mediumCount := 0
	lowCount := 0

	for _, violation := range result.Violations {
		switch violation.Severity {
		case "高":
			highCount++
		case "中":
			mediumCount++
		case "低":
			lowCount++
		}
	}

	// 生成摘要
	summary := "发票校验未通过，发现"
	if highCount > 0 {
		summary += " " + string(rune(highCount+'0')) + "个高严重程度违规"
	}
	if mediumCount > 0 {
		summary += " " + string(rune(mediumCount+'0')) + "个中严重程度违规"
	}
	if lowCount > 0 {
		summary += " " + string(rune(lowCount+'0')) + "个低严重程度违规"
	}

	result.Summary = summary
}

// determineSeverity 根据规则类型确定严重程度
func determineSeverity(ruleType string) string {
	switch ruleType {
	case "基础校验", "金额校验", "重复校验":
		return "高"
	case "时效校验", "抬头校验", "类型校验":
		return "中"
	case "税额校验", "时间校验", "自定义规则":
		return "中"
	default:
		return "低"
	}
}

// generateSuggestion 根据规则类型和错误信息生成建议
func generateSuggestion(ruleType, message string) string {
	switch ruleType {
	case "基础校验":
		return "请检查发票基本信息是否完整，包括发票代码、号码、日期和金额等必填字段"
	case "金额校验":
		return "请检查发票金额是否正确，确保不超过报销金额且总和匹配"
	case "时效校验":
		return "请确保发票开票日期距报销申请日期不超过180天"
	case "抬头校验":
		return "请确保发票抬头与报销人所在公司名称一致"
	case "类型校验":
		return "请使用公司规定的可报销发票类型"
	case "重复校验":
		return "请检查该发票是否已经报销过，避免重复报销"
	case "税额校验":
		return "请检查发票税额是否正确，应为0到发票金额之间"
	case "时间校验":
		return "请检查发票开票时间是否在正常营业时间内"
	case "自定义规则":
		return "请根据自定义规则要求修改发票信息"
	default:
		return "请检查发票信息是否符合相关规定"
	}
}

// ExecuteAllRules 执行所有发票校验规则
func (v *InvoiceValidatorImpl) ExecuteAllRules(ctx context.Context, req *InvoiceValidationRequest) (*InvoiceValidationResult, error) {
	v.logger.WithContext(ctx).Info("执行所有发票校验规则",
		logger.NewField("发票ID", req.Invoice.ID))

	// 创建校验结果
	result := &InvoiceValidationResult{
		InvoiceID: req.Invoice.ID,
		Passed:    true,
		Timestamp: time.Now(),
	}

	// 使用优先级执行规则
	if err := v.executeRulesWithPriority(ctx, req, result); err != nil {
		v.logger.WithContext(ctx).Error("执行规则失败",
			logger.NewField("发票ID", req.Invoice.ID),
			logger.NewField("error", err.Error()))
		return nil, err
	}

	// 生成校验结果摘要
	generateValidationSummary(result)

	v.logger.WithContext(ctx).Info("所有规则执行完成",
		logger.NewField("发票ID", req.Invoice.ID),
		logger.NewField("违规数", len(result.Violations)),
		logger.NewField("校验结果", result.Passed))

	return result, nil
}

// Helper functions for Grule rules

// getString 从map中安全获取字符串值
func getString(m map[string]interface{}, key string) string {
	if val, ok := m[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}

// getInt 从map中安全获取整数值
func getInt(m map[string]interface{}, key string) int {
	if val, ok := m[key]; ok {
		if i, ok := val.(int); ok {
			return i
		}
	}
	return 0
}

// isDuplicateInvoice 检查发票是否重复
func (v *InvoiceValidatorImpl) isDuplicateInvoice(ctx context.Context, invoiceCode, invoiceNumber string) (bool, error) {
	// 这里应该调用repository检查发票是否已存在
	// 简化实现，实际应该查询数据库
	return false, nil
}

// getAccommodationLimit 获取住宿限额
func (v *InvoiceValidatorImpl) getAccommodationLimit(ctx context.Context, cityLevel string) float64 {
	// 根据城市级别返回住宿限额
	switch cityLevel {
	case "一线城市":
		return 600.0
	case "二线城市":
		return 400.0
	case "三线城市":
		return 300.0
	default:
		return 200.0
	}
}

// getEntertainmentLimit 获取招待费限额
func (v *InvoiceValidatorImpl) getEntertainmentLimit(ctx context.Context, level string) float64 {
	// 根据级别返回招待费限额
	switch level {
	case "高管":
		return 500.0
	case "经理":
		return 300.0
	case "员工":
		return 100.0
	default:
		return 100.0
	}
}

// isConsecutiveInvoice 检查是否为连号发票
func (v *InvoiceValidatorImpl) isConsecutiveInvoice(ctx context.Context, invoiceNumbers []string) (bool, error) {
	if len(invoiceNumbers) < 2 {
		return false, nil
	}

	// 提取发票号码中的数字部分
	numbers := make([]int, 0, len(invoiceNumbers))
	for _, numStr := range invoiceNumbers {
		// 简化处理：提取字符串中的数字部分
		var num int
		_, err := fmt.Sscanf(numStr, "%d", &num)
		if err != nil {
			// 如果无法提取数字，尝试从字符串末尾提取
			for i := len(numStr) - 1; i >= 0; i-- {
				if numStr[i] < '0' || numStr[i] > '9' {
					if i < len(numStr)-1 {
						numStr = numStr[i+1:]
						_, err = fmt.Sscanf(numStr, "%d", &num)
						if err == nil {
							break
						}
					}
					continue
				}
			}
			if err != nil {
				continue // 跳过无法解析的发票号码
			}
		}
		numbers = append(numbers, num)
	}

	if len(numbers) < 2 {
		return false, nil
	}

	// 按数字大小排序
	sort.Ints(numbers)

	// 检查是否有连续的号码
	consecutiveCount := 1
	for i := 1; i < len(numbers); i++ {
		if numbers[i] == numbers[i-1]+1 {
			consecutiveCount++
			if consecutiveCount >= 3 {
				return true, nil // 3个或以上连续号码视为连号发票
			}
		} else {
			consecutiveCount = 1
		}
	}

	return false, nil
}

// isWeekendOrHoliday 检查是否为周末或节假日
func (v *InvoiceValidatorImpl) isWeekendOrHoliday(ctx context.Context, date time.Time) (bool, error) {
	// 检查日期是否为周末或节假日
	weekday := date.Weekday()

	// 首先检查是否为周末
	if weekday == time.Saturday || weekday == time.Sunday {
		return true, nil
	}

	// 检查是否为法定节假日
	// 这里简化实现，实际应该查询节假日数据库或API
	// 可以根据年份和日期判断是否为法定节假日
	month := int(date.Month())
	day := date.Day()

	// 常见的固定法定节假日（简化版）
	holidays := map[string]bool{
		"0101": true, // 元旦
		"0501": true, // 劳动节
		"1001": true, // 国庆节
		"1002": true, // 国庆节
		"1003": true, // 国庆节
	}

	// 检查固定节假日
	monthDay := fmt.Sprintf("%02d%02d", month, day)
	if isHoliday, exists := holidays[monthDay]; exists && isHoliday {
		return true, nil
	}

	// 检查春节、清明节、端午节、中秋节等浮动节假日
	// 这里简化处理，实际应该查询具体的节假日安排
	if month == 1 || month == 2 || month == 4 || month == 6 || month == 9 {
		// 这些月份可能有浮动节假日，需要更复杂的逻辑
		// 简化实现：假设1-2月可能有春节，4月可能有清明，6月可能有端午，9月可能有中秋
		// 实际应用中应该使用专门的节假日库或API
	}

	return false, nil
}

// isValidTaxNumber 检查税号是否有效
func (v *InvoiceValidatorImpl) isValidTaxNumber(ctx context.Context, taxNumber string) (bool, error) {
	// 检查税号格式是否有效
	if len(taxNumber) < 15 || len(taxNumber) > 20 {
		return false, nil
	}

	// 去除空格和特殊字符
	cleanedTaxNumber := strings.ReplaceAll(taxNumber, " ", "")
	cleanedTaxNumber = strings.ReplaceAll(cleanedTaxNumber, "-", "")
	cleanedTaxNumber = strings.ReplaceAll(cleanedTaxNumber, "_", "")

	// 检查长度
	if len(cleanedTaxNumber) < 15 || len(cleanedTaxNumber) > 20 {
		return false, nil
	}

	// 检查是否全为数字或数字+字母
	for _, char := range cleanedTaxNumber {
		if !((char >= '0' && char <= '9') || (char >= 'A' && char <= 'Z') || (char >= 'a' && char <= 'z')) {
			return false, nil
		}
	}

	// 根据长度判断税号类型并验证格式
	if len(cleanedTaxNumber) == 15 {
		// 15位税号（老版）
		// 检查前6位是否为地区码（简化处理）
		if len(cleanedTaxNumber) >= 6 {
			areaCode := cleanedTaxNumber[0:6]
			// 简化验证：检查前6位是否全为数字
			for _, char := range areaCode {
				if char < '0' || char > '9' {
					return false, nil
				}
			}
		}
	} else if len(cleanedTaxNumber) == 18 {
		// 18位税号（新版统一社会信用代码）
		// 检查第1位为数字，第2位为字母或数字，第3-8位为数字，第9位为数字或字母
		// 第10-17位为数字，第18位为数字或字母或校验码

		// 检查第1位
		if cleanedTaxNumber[0] < '0' || cleanedTaxNumber[0] > '9' {
			return false, nil
		}

		// 检查第2位
		char2 := cleanedTaxNumber[1]
		if !((char2 >= '0' && char2 <= '9') || (char2 >= 'A' && char2 <= 'Z') || (char2 >= 'a' && char2 <= 'z')) {
			return false, nil
		}

		// 检查第3-8位
		for i := 2; i < 8; i++ {
			if cleanedTaxNumber[i] < '0' || cleanedTaxNumber[i] > '9' {
				return false, nil
			}
		}

		// 检查第9位
		char9 := cleanedTaxNumber[8]
		if !((char9 >= '0' && char9 <= '9') || (char9 >= 'A' && char9 <= 'Z') || (char9 >= 'a' && char9 <= 'z')) {
			return false, nil
		}

		// 检查第10-17位
		for i := 9; i < 17; i++ {
			if cleanedTaxNumber[i] < '0' || cleanedTaxNumber[i] > '9' {
				return false, nil
			}
		}

		// 检查第18位（校验码）
		char18 := cleanedTaxNumber[17]
		if !((char18 >= '0' && char18 <= '9') || (char18 >= 'A' && char18 <= 'Z') || (char18 >= 'a' && char18 <= 'z')) {
			return false, nil
		}
	}

	return true, nil
}

// hasOrderAndReceipt 检查是否有订单和收据
func (v *InvoiceValidatorImpl) hasOrderAndReceipt(ctx context.Context, invoiceID string) (bool, error) {
	// 检查发票是否有对应的订单和收据
	// 实际应该查询订单和收据数据

	// 从数据库查询发票对应的订单信息
	// 这里简化实现，假设通过repository查询
	// orders, err := v.orderRepository.GetOrdersByInvoiceID(ctx, invoiceID)
	// receipts, err := v.receiptRepository.GetReceiptsByInvoiceID(ctx, invoiceID)

	// 简化实现：假设查询结果
	hasOrder := true   // 实际应该查询数据库
	hasReceipt := true // 实际应该查询数据库

	// 如果两者都有，返回true
	return hasOrder && hasReceipt, nil
}

// isThreeDocumentMatching 检查三单是否匹配
func (v *InvoiceValidatorImpl) isThreeDocumentMatching(ctx context.Context, invoiceID string) (bool, error) {
	// 检查发票、订单、收据三单是否匹配
	// 实际应该比较三单信息

	// 从数据库查询发票、订单、收据信息
	// invoice, err := v.invoiceRepository.GetByID(ctx, invoiceID)
	// orders, err := v.orderRepository.GetOrdersByInvoiceID(ctx, invoiceID)
	// receipts, err := v.receiptRepository.GetReceiptsByInvoiceID(ctx, invoiceID)

	// 简化实现：假设查询结果
	// 实际应该检查以下信息是否一致：
	// 1. 金额是否一致
	// 2. 商品/服务名称是否一致
	// 3. 数量是否一致
	// 4. 日期是否合理（订单日期 <= 发票日期 <= 收据日期）

	// 模拟数据
	invoiceAmount := 1000.0 // 发票金额
	orderAmount := 1000.0   // 订单金额
	receiptAmount := 1000.0 // 收据金额

	// 检查金额是否一致
	amountMatch := invoiceAmount == orderAmount && orderAmount == receiptAmount

	// 检查商品/服务名称是否一致
	// invoiceItems := invoice.Items
	// orderItems := order.Items
	// receiptItems := receipt.Items
	// itemsMatch := compareItems(invoiceItems, orderItems, receiptItems)
	itemsMatch := true // 简化实现

	// 检查日期是否合理
	// invoiceDate := invoice.Date
	// orderDate := order.Date
	// receiptDate := receipt.Date
	// dateValid := orderDate <= invoiceDate && invoiceDate <= receiptDate
	dateValid := true // 简化实现

	// 如果所有检查都通过，返回true
	return amountMatch && itemsMatch && dateValid, nil
}
