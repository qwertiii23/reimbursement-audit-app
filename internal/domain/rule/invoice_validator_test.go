package rule

import (
	"context"
	"fmt"
	"io"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"reimbursement-audit/internal/domain/ocr"
	"reimbursement-audit/internal/domain/reimbursement"
	"reimbursement-audit/internal/pkg/logger"
)

type MockRuleRepository struct {
	rules map[string]*Rule
}

func NewMockRuleRepository() *MockRuleRepository {
	return &MockRuleRepository{
		rules: make(map[string]*Rule),
	}
}

func (m *MockRuleRepository) AddTestRule(rule *Rule) {
	m.rules[rule.ID] = rule
}

func (m *MockRuleRepository) AddTestRules(rules []*Rule) {
	for _, rule := range rules {
		m.rules[rule.ID] = rule
	}
}

func (m *MockRuleRepository) CreateRule(ctx context.Context, rule *Rule) error {
	m.rules[rule.ID] = rule
	return nil
}

func (m *MockRuleRepository) Create(ctx context.Context, rule *Rule) error {
	m.rules[rule.ID] = rule
	return nil
}

func (m *MockRuleRepository) GetRuleByID(ctx context.Context, id string) (*Rule, error) {
	if rule, ok := m.rules[id]; ok {
		return rule, nil
	}
	return nil, nil
}

func (m *MockRuleRepository) GetByID(ctx context.Context, id string) (*Rule, error) {
	if rule, ok := m.rules[id]; ok {
		return rule, nil
	}
	return nil, nil
}

func (m *MockRuleRepository) GetRuleByCode(ctx context.Context, ruleCode string) (*Rule, error) {
	for _, rule := range m.rules {
		if rule.RuleCode == ruleCode {
			return rule, nil
		}
	}
	return nil, nil
}

func (m *MockRuleRepository) GetByRuleCode(ctx context.Context, ruleCode string) (*Rule, error) {
	for _, rule := range m.rules {
		if rule.RuleCode == ruleCode {
			return rule, nil
		}
	}
	return nil, nil
}

func (m *MockRuleRepository) UpdateRule(ctx context.Context, rule *Rule) error {
	if _, ok := m.rules[rule.ID]; ok {
		m.rules[rule.ID] = rule
	}
	return nil
}

func (m *MockRuleRepository) Update(ctx context.Context, rule *Rule) error {
	if _, ok := m.rules[rule.ID]; ok {
		m.rules[rule.ID] = rule
	}
	return nil
}

func (m *MockRuleRepository) DeleteRule(ctx context.Context, id string) error {
	delete(m.rules, id)
	return nil
}

func (m *MockRuleRepository) Delete(ctx context.Context, id string) error {
	delete(m.rules, id)
	return nil
}

func (m *MockRuleRepository) ListRules(ctx context.Context, filter *RuleFilter) ([]*Rule, int64, error) {
	rules := make([]*Rule, 0, len(m.rules))
	for _, rule := range m.rules {
		if filter.Type != "" && rule.Type != filter.Type {
			continue
		}
		if filter.Status != "" {
			enabled := rule.Enabled
			if (filter.Status == "enabled" && !enabled) || (filter.Status == "disabled" && enabled) {
				continue
			}
		}
		rules = append(rules, rule)
	}
	return rules, int64(len(rules)), nil
}

func (m *MockRuleRepository) CountRules(ctx context.Context, filter *RuleFilter) (int64, error) {
	rules, _, err := m.ListRules(ctx, filter)
	if err != nil {
		return 0, err
	}
	return int64(len(rules)), nil
}

func (m *MockRuleRepository) EnableRule(ctx context.Context, id string) error {
	if rule, ok := m.rules[id]; ok {
		rule.Enabled = true
	}
	return nil
}

func (m *MockRuleRepository) DisableRule(ctx context.Context, id string) error {
	if rule, ok := m.rules[id]; ok {
		rule.Enabled = false
	}
	return nil
}

func (m *MockRuleRepository) CheckRuleCodeExists(ctx context.Context, ruleCode string, excludeID string) (bool, error) {
	for _, rule := range m.rules {
		if rule.RuleCode == ruleCode && rule.ID != excludeID {
			return true, nil
		}
	}
	return false, nil
}

type MockLogger struct {
	debugMessages []string
	infoMessages  []string
	warnMessages  []string
	errorMessages []string
}

func (m *MockLogger) Debug(msg string, fields ...logger.Field) {
	logMsg := msg
	for _, f := range fields {
		logMsg += fmt.Sprintf(" %s=%v", f.Key, f.Value)
	}
	m.debugMessages = append(m.debugMessages, logMsg)
}
func (m *MockLogger) Info(msg string, fields ...logger.Field) {
	logMsg := msg
	for _, f := range fields {
		logMsg += fmt.Sprintf(" %s=%v", f.Key, f.Value)
	}
	m.infoMessages = append(m.infoMessages, logMsg)
}
func (m *MockLogger) Warn(msg string, fields ...logger.Field) {
	logMsg := msg
	for _, f := range fields {
		logMsg += fmt.Sprintf(" %s=%v", f.Key, f.Value)
	}
	m.warnMessages = append(m.warnMessages, logMsg)
}
func (m *MockLogger) Error(msg string, fields ...logger.Field) {
	logMsg := msg
	for _, f := range fields {
		logMsg += fmt.Sprintf(" %s=%v", f.Key, f.Value)
	}
	m.errorMessages = append(m.errorMessages, logMsg)
}
func (m *MockLogger) Fatal(msg string, fields ...logger.Field) {}
func (m *MockLogger) WithContext(ctx context.Context) logger.Logger {
	return m
}
func (m *MockLogger) WithFields(fields ...logger.Field) logger.Logger {
	return m
}
func (m *MockLogger) WithField(key string, value interface{}) logger.Logger {
	return m
}
func (m *MockLogger) SetLevel(level logger.Level) {}
func (m *MockLogger) GetLevel() logger.Level {
	return logger.InfoLevel
}
func (m *MockLogger) SetOutput(w io.Writer) {}
func (m *MockLogger) Close() error {
	return nil
}

func (m *MockLogger) PrintAllLogs(t *testing.T) {
	t.Logf("=== Debug Messages (%d) ===", len(m.debugMessages))
	for _, msg := range m.debugMessages {
		t.Logf("  DEBUG: %s", msg)
	}
	t.Logf("=== Info Messages (%d) ===", len(m.infoMessages))
	for _, msg := range m.infoMessages {
		t.Logf("  INFO: %s", msg)
	}
	t.Logf("=== Warn Messages (%d) ===", len(m.warnMessages))
	for _, msg := range m.warnMessages {
		t.Logf("  WARN: %s", msg)
	}
	t.Logf("=== Error Messages (%d) ===", len(m.errorMessages))
	for _, msg := range m.errorMessages {
		t.Logf("  ERROR: %s", msg)
	}
}

func createTestRules() []*Rule {
	return []*Rule{
		{
			ID:          "rule-date-valid-001",
			RuleCode:    "RULE_INVOICE_DATE_VALID",
			Name:        "发票开票日期必须在半年内",
			Type:        "invoice_validation",
			Category:    "时效性校验",
			Status:      "enabled",
			Description: "发票开票日期必须在半年（180天）内，超过期限无法报销",
			Definition: `rule invoice_date_valid "发票开票日期时效性检查" salience 100 {
			when
				data.IsInvoiceDateOlderThan6Months == true && result.Passed == true
			then
				result.Passed = false;
				result.Message = "发票开票日期超过半年，无法报销";
				result.Severity = "high";
				result.Priority = 100;
			}`,
			Priority: 100,
			Enabled:  true,
		},
		{
			ID:          "rule-accommodation-limit-001",
			RuleCode:    "RULE_ACCOMMODATION_LIMIT",
			Name:        "住宿费单次限额200元",
			Type:        "invoice_validation",
			Category:    "金额校验",
			Status:      "enabled",
			Description: "住宿费单次不得超过200元，超出部分需特殊审批",
			Definition: `rule accommodation_limit "住宿费限额检查" salience 90 {
			when
				data.Invoice.Category == "差旅费" && data.Invoice.SubCategory == "住宿费" && data.Invoice.Amount > 200.0 && result.Passed == true
			then
				result.Passed = false;
				result.Message = "住宿费单次不得超过200元，超出部分需特殊审批";
				result.Severity = "high";
				result.Priority = 90;
			}`,
			Priority: 90,
			Enabled:  true,
		},
		{
			ID:          "rule-missing-fields-001",
			RuleCode:    "RULE_MISSING_REQUIRED_FIELDS",
			Name:        "发票必填字段检查",
			Type:        "invoice_validation",
			Category:    "基础校验",
			Status:      "enabled",
			Description: "发票代码、发票号码、开票日期为必填字段",
			Definition: `rule missing_required_fields "必填字段检查" salience 110 {
			when
				(data.Invoice.Code == "" || data.Invoice.Number == "" || data.Invoice.Date == nil) && result.Passed == true
			then
				result.Passed = false;
				result.Message = "发票缺少必填字段";
				result.Severity = "high";
				result.Priority = 110;
			}`,
			Priority: 110,
			Enabled:  true,
		},
		{
			ID:          "rule-large-amount-001",
			RuleCode:    "RULE_LARGE_AMOUNT_APPROVAL",
			Name:        "大额发票审批规则",
			Type:        "invoice_validation",
			Category:    "金额校验",
			Status:      "enabled",
			Description: "单张发票金额超过5000元需要特殊审批",
			Definition: `rule large_amount_approval "大额发票审批检查" salience 80 {
			when
				data.Invoice.Amount > 5000.0 && result.Passed == true
			then
				result.Passed = false;
				result.Message = "单张发票金额超过5000元，需要特殊审批";
				result.Severity = "high";
				result.Priority = 80;
			}`,
			Priority: 80,
			Enabled:  true,
		},
	}
}

func setupTestValidator(t *testing.T) (InvoiceValidator, *MockRuleRepository, *MockLogger) {
	ctx := context.Background()
	mockLogger := &MockLogger{}
	mockRepo := NewMockRuleRepository()

	testRules := createTestRules()
	mockRepo.AddTestRules(testRules)

	ruleEngine := NewGRuleEngine(mockRepo, mockLogger)
	validator := NewInvoiceValidator(ruleEngine, mockRepo, mockLogger)

	err := validator.LoadRules(ctx)
	require.NoError(t, err, "加载规则应该成功")

	return validator, mockRepo, mockLogger
}

func TestInvoiceValidator_DateValidation(t *testing.T) {
	validator, _, mockLogger := setupTestValidator(t)
	ctx := context.Background()

	oldDate := time.Now().AddDate(0, -7, 0)
	applyDate := time.Now()

	t.Logf("发票日期: %v", oldDate)
	t.Logf("申请日期: %v", applyDate)
	t.Logf("发票日期+6个月: %v", oldDate.AddDate(0, 6, 0))
	t.Logf("申请日期是否在发票日期+6个月之后: %v", applyDate.After(oldDate.AddDate(0, 6, 0)))

	invoice := &ocr.Invoice{
		ID:          "test-invoice-001",
		Code:        "12345",
		Number:      "67890",
		Date:        &oldDate,
		Amount:      100.0,
		Category:    "差旅费",
		SubCategory: "交通费",
	}

	reimbursement := &reimbursement.Reimbursement{
		ID:     "test-reimbursement-001",
		UserID: "user001",
	}

	req := &InvoiceValidationRequest{
		Invoice:       invoice,
		Reimbursement: reimbursement,
		ApplyDate:     applyDate,
	}

	t.Logf("已加载规则数量: %d", len(validator.GetRuleDefinitions()))
	for _, rule := range validator.GetRuleDefinitions() {
		t.Logf("规则: %s (ID: %s, Enabled: %v)", rule.Name, rule.ID, rule.Enabled)
	}

	result, err := validator.ValidateSingle(ctx, req)
	require.NoError(t, err)
	assert.NotNil(t, result)

	t.Logf("校验结果 - Passed: %v", result.Passed)
	t.Logf("违规数量: %d", len(result.Violations))
	for i, violation := range result.Violations {
		t.Logf("违规 %d: %s - %s", i+1, violation.RuleName, violation.Message)
	}

	mockLogger.PrintAllLogs(t)

	assert.False(t, result.Passed, "发票日期超过半年应该不通过校验")

	if len(result.Violations) > 0 {
		t.Logf("发现违规: %s", result.Violations[0].Message)
	}
}

func TestInvoiceValidator_AccommodationLimit(t *testing.T) {
	validator, _, _ := setupTestValidator(t)
	ctx := context.Background()

	invoiceDate := time.Now().AddDate(0, -1, 0)
	applyDate := time.Now()

	invoice := &ocr.Invoice{
		ID:          "test-invoice-002",
		Code:        "12345",
		Number:      "67890",
		Date:        &invoiceDate,
		Amount:      300.0,
		Category:    "差旅费",
		SubCategory: "住宿费",
	}

	reimbursement := &reimbursement.Reimbursement{
		ID:     "test-reimbursement-002",
		UserID: "user002",
	}

	req := &InvoiceValidationRequest{
		Invoice:       invoice,
		Reimbursement: reimbursement,
		ApplyDate:     applyDate,
	}

	result, err := validator.ValidateSingle(ctx, req)
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.False(t, result.Passed, "住宿费超过200元应该不通过校验")

	if len(result.Violations) > 0 {
		t.Logf("发现违规: %s", result.Violations[0].Message)
	}
}

func TestInvoiceValidator_ValidInvoice(t *testing.T) {
	validator, _, _ := setupTestValidator(t)
	ctx := context.Background()

	invoiceDate := time.Now().AddDate(0, -1, 0)
	applyDate := time.Now()

	invoice := &ocr.Invoice{
		ID:          "test-invoice-003",
		Code:        "12345",
		Number:      "67890",
		Date:        &invoiceDate,
		Amount:      150.0,
		Category:    "差旅费",
		SubCategory: "住宿费",
	}

	reimbursement := &reimbursement.Reimbursement{
		ID:     "test-reimbursement-003",
		UserID: "user003",
	}

	req := &InvoiceValidationRequest{
		Invoice:       invoice,
		Reimbursement: reimbursement,
		ApplyDate:     applyDate,
	}

	result, err := validator.ValidateSingle(ctx, req)
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.True(t, result.Passed, "合规的发票应该通过校验")

	t.Logf("校验结果: %s", result.Summary)
}

func TestInvoiceValidator_MissingRequiredFields(t *testing.T) {
	validator, _, _ := setupTestValidator(t)
	ctx := context.Background()

	invoiceDate := time.Now().AddDate(0, -1, 0)
	applyDate := time.Now()

	invoice := &ocr.Invoice{
		ID:          "test-invoice-004",
		Code:        "",
		Number:      "",
		Date:        &invoiceDate,
		Amount:      100.0,
		Category:    "差旅费",
		SubCategory: "交通费",
	}

	reimbursement := &reimbursement.Reimbursement{
		ID:     "test-reimbursement-004",
		UserID: "user004",
	}

	req := &InvoiceValidationRequest{
		Invoice:       invoice,
		Reimbursement: reimbursement,
		ApplyDate:     applyDate,
	}

	result, err := validator.ValidateSingle(ctx, req)
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.False(t, result.Passed, "缺少必填字段的发票应该不通过校验")

	if len(result.Violations) > 0 {
		for _, violation := range result.Violations {
			t.Logf("发现违规: %s - %s", violation.RuleName, violation.Message)
		}
	}
}

func TestInvoiceValidator_LargeAmountApproval(t *testing.T) {
	validator, _, _ := setupTestValidator(t)
	ctx := context.Background()

	invoiceDate := time.Now().AddDate(0, -1, 0)
	applyDate := time.Now()

	invoice := &ocr.Invoice{
		ID:          "test-invoice-005",
		Code:        "12345",
		Number:      "67890",
		Date:        &invoiceDate,
		Amount:      6000.0,
		Category:    "办公费",
		SubCategory: "设备采购",
	}

	reimbursement := &reimbursement.Reimbursement{
		ID:     "test-reimbursement-005",
		UserID: "user005",
	}

	req := &InvoiceValidationRequest{
		Invoice:       invoice,
		Reimbursement: reimbursement,
		ApplyDate:     applyDate,
	}

	result, err := validator.ValidateSingle(ctx, req)
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.False(t, result.Passed, "大额发票应该需要特殊审批")

	if len(result.Violations) > 0 {
		t.Logf("发现违规: %s", result.Violations[0].Message)
	}
}
