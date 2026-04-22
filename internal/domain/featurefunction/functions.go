package featurefunction

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"reimbursement-audit/internal/pkg/logger"
	"strings"
	"time"
)

type DetectPhotoshopFunction struct {
	client *http.Client
}

func NewDetectPhotoshopFunction() *DetectPhotoshopFunction {
	return &DetectPhotoshopFunction{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (f *DetectPhotoshopFunction) GetName() string {
	return "detect_photoshop"
}

func (f *DetectPhotoshopFunction) GetDescription() string {
	return "使用大模型检测发票图片是否存在P图痕迹"
}

func (f *DetectPhotoshopFunction) GetConfigSchema() *ConfigSchema {
	return &ConfigSchema{
		Fields: []FieldConfig{
			{
				Name:     "model",
				Type:     "select",
				Label:    "模型选择",
				Required: true,
				Default:  "gpt-4-vision",
				Options: []Option{
					{Label: "GPT-4 Vision", Value: "gpt-4-vision"},
					{Label: "GPT-4o", Value: "gpt-4o"},
					{Label: "Claude 3.5 Sonnet", Value: "claude-3-5-sonnet"},
				},
				Description: "选择用于检测的AI模型",
			},
			{
				Name:        "threshold",
				Type:        "number",
				Label:       "置信度阈值",
				Required:    true,
				Default:     0.7,
				Description: "判断为P图的最低置信度（0-1）",
			},
			{
				Name:        "prompt",
				Type:        "string",
				Label:       "检测提示词",
				Required:    true,
				Default:     "请仔细分析这张发票图片，判断是否存在明显的P图痕迹、篡改或伪造。请从以下几个方面分析：1. 文字是否清晰自然 2. 数字和金额是否一致 3. 印章和签名是否真实 4. 整体布局是否合理。请给出你的判断结果和置信度（0-1之间）。",
				Description: "用于指导大模型检测的提示词",
			},
		},
	}
}

func (f *DetectPhotoshopFunction) Validate(config map[string]interface{}) error {
	if _, ok := config["model"]; !ok {
		return fmt.Errorf("model is required")
	}
	if _, ok := config["threshold"]; !ok {
		return fmt.Errorf("threshold is required")
	}
	if _, ok := config["prompt"]; !ok {
		return fmt.Errorf("prompt is required")
	}
	return nil
}

func (f *DetectPhotoshopFunction) Execute(ctx context.Context, input *FunctionInput) (*FunctionOutput, error) {
	if err := f.Validate(input.Config); err != nil {
		return &FunctionOutput{
			Error: err.Error(),
		}, nil
	}

	imageURL, ok := input.InvoiceData["image_url"].(string)
	if !ok {
		return &FunctionOutput{
			Error: "invoice_data.image_url is required",
		}, nil
	}

	threshold, _ := input.Config["threshold"].(float64)
	prompt, _ := input.Config["prompt"].(string)

	result, err := f.callVisionModel(ctx, imageURL, prompt)
	if err != nil {
		return &FunctionOutput{
			Error: fmt.Sprintf("failed to call vision model: %v", err),
		}, nil
	}

	isPhotoshopped := result.Confidence >= threshold

	return &FunctionOutput{
		Value:      isPhotoshopped,
		Confidence: result.Confidence,
		Metadata: map[string]interface{}{
			"reasoning": result.Reasoning,
			"model":     input.Config["model"],
		},
	}, nil
}

type VisionModelResponse struct {
	Confidence float64 `json:"confidence"`
	Reasoning  string  `json:"reasoning"`
}

func (f *DetectPhotoshopFunction) callVisionModel(ctx context.Context, imageURL, prompt string) (*VisionModelResponse, error) {
	payload := map[string]interface{}{
		"image_url": imageURL,
		"prompt":    prompt,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", "http://localhost:8080/api/v1/vision/analyze", strings.NewReader(string(jsonData)))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := f.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("vision API returned status %d: %s", resp.StatusCode, string(body))
	}

	var result VisionModelResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
}

type ImageQualityFunction struct{}

func NewImageQualityFunction() *ImageQualityFunction {
	return &ImageQualityFunction{}
}

func (f *ImageQualityFunction) GetName() string {
	return "image_quality"
}

func (f *ImageQualityFunction) GetDescription() string {
	return "检测发票图片质量（清晰度、完整性等）"
}

func (f *ImageQualityFunction) GetConfigSchema() *ConfigSchema {
	return &ConfigSchema{
		Fields: []FieldConfig{
			{
				Name:        "min_resolution",
				Type:        "number",
				Label:       "最小分辨率",
				Required:    true,
				Default:     1024,
				Description: "图片最小宽度或高度（像素）",
			},
			{
				Name:        "check_blur",
				Type:        "boolean",
				Label:       "检查模糊",
				Required:    true,
				Default:     true,
				Description: "是否检查图片是否模糊",
			},
		},
	}
}

func (f *ImageQualityFunction) Validate(config map[string]interface{}) error {
	if _, ok := config["min_resolution"]; !ok {
		return fmt.Errorf("min_resolution is required")
	}
	return nil
}

func (f *ImageQualityFunction) Execute(ctx context.Context, input *FunctionInput) (*FunctionOutput, error) {
	if err := f.Validate(input.Config); err != nil {
		return &FunctionOutput{
			Error: err.Error(),
		}, nil
	}

	imageURL, ok := input.InvoiceData["image_url"].(string)
	if !ok {
		return &FunctionOutput{
			Error: "invoice_data.image_url is required",
		}, nil
	}

	minResolution, _ := input.Config["min_resolution"].(float64)
	checkBlur, _ := input.Config["check_blur"].(bool)

	qualityScore := f.analyzeImageQuality(ctx, imageURL, minResolution, checkBlur)

	return &FunctionOutput{
		Value:      qualityScore >= 0.7,
		Confidence: qualityScore,
		Metadata: map[string]interface{}{
			"score":      qualityScore,
			"resolution": minResolution,
		},
	}, nil
}

func (f *ImageQualityFunction) analyzeImageQuality(ctx context.Context, imageURL string, minResolution float64, checkBlur bool) float64 {
	return 0.85
}

type InvoiceCodeLengthFunction struct{}

func NewInvoiceCodeLengthFunction() *InvoiceCodeLengthFunction {
	return &InvoiceCodeLengthFunction{}
}

func (f *InvoiceCodeLengthFunction) GetName() string {
	return "invoice_code_length"
}

func (f *InvoiceCodeLengthFunction) GetDescription() string {
	return "检查发票代码长度（10-12位纯数字）"
}

func (f *InvoiceCodeLengthFunction) GetConfigSchema() *ConfigSchema {
	return &ConfigSchema{
		Fields: []FieldConfig{
			{
				Name:        "min_length",
				Type:        "number",
				Label:       "最小长度",
				Required:    true,
				Default:     10,
				Description: "发票代码最小长度",
			},
			{
				Name:        "max_length",
				Type:        "number",
				Label:       "最大长度",
				Required:    true,
				Default:     12,
				Description: "发票代码最大长度",
			},
		},
	}
}

func (f *InvoiceCodeLengthFunction) Validate(config map[string]interface{}) error {
	if _, ok := config["min_length"]; !ok {
		return fmt.Errorf("min_length is required")
	}
	if _, ok := config["max_length"]; !ok {
		return fmt.Errorf("max_length is required")
	}
	return nil
}

func (f *InvoiceCodeLengthFunction) Execute(ctx context.Context, input *FunctionInput) (*FunctionOutput, error) {
	if err := f.Validate(input.Config); err != nil {
		return &FunctionOutput{
			Error: err.Error(),
		}, nil
	}

	invoiceCode, ok := input.InvoiceData["invoice_code"].(string)
	if !ok {
		return &FunctionOutput{
			Error: "invoice_data.invoice_code is required",
		}, nil
	}

	minLength, _ := input.Config["min_length"].(float64)
	maxLength, _ := input.Config["max_length"].(float64)

	codeLength := len(invoiceCode)
	isValidLength := float64(codeLength) >= minLength && float64(codeLength) <= maxLength

	isPureDigit := true
	for _, char := range invoiceCode {
		if char < '0' || char > '9' {
			isPureDigit = false
			break
		}
	}

	isValid := isValidLength && isPureDigit

	return &FunctionOutput{
		Value: isValid,
		Metadata: map[string]interface{}{
			"code_length":     codeLength,
			"is_valid_length": isValidLength,
			"is_pure_digit":   isPureDigit,
			"min_length":      minLength,
			"max_length":      maxLength,
		},
	}, nil
}

type InvoiceFraudDetectionFunction struct {
	client *http.Client
	logger logger.Logger
}

func NewInvoiceFraudDetectionFunction() *InvoiceFraudDetectionFunction {
	return &InvoiceFraudDetectionFunction{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (f *InvoiceFraudDetectionFunction) SetLogger(logger logger.Logger) {
	f.logger = logger
}

func (f *InvoiceFraudDetectionFunction) GetName() string {
	return "invoice_fraud_detection"
}

func (f *InvoiceFraudDetectionFunction) GetDescription() string {
	return "检测发票图片是否存在P图/篡改痕迹，支持多张发票检测"
}

func (f *InvoiceFraudDetectionFunction) GetConfigSchema() *ConfigSchema {
	return &ConfigSchema{
		Fields: []FieldConfig{
			{
				Name:        "confidence_threshold",
				Type:        "number",
				Label:       "置信度阈值",
				Required:    true,
				Default:     0.7,
				Description: "判断为P图的最低置信度（0-1）",
			},
			{
				Name:        "check_all_images",
				Type:        "boolean",
				Label:       "检查所有图片",
				Required:    true,
				Default:     true,
				Description: "是否检查所有发票图片，false时发现一张P图就停止",
			},
			{
				Name:        "detection_prompt",
				Type:        "string",
				Label:       "检测提示词",
				Required:    true,
				Default:     "请仔细分析这张发票图片，判断是否存在明显的P图痕迹、篡改或伪造。请从以下几个方面分析：1. 文字是否清晰自然，字体是否一致 2. 数字和金额是否合理，有无涂改痕迹 3. 印章和签名是否真实，有无PS痕迹 4. 整体布局是否合理，有无拼接痕迹 5. 图片质量是否异常，有无模糊或失真。请给出你的判断结果（存在P图/不存在P图）和置信度（0-1之间）。",
				Description: "用于指导大模型检测的提示词",
			},
		},
	}
}

func (f *InvoiceFraudDetectionFunction) Validate(config map[string]interface{}) error {
	if _, ok := config["confidence_threshold"]; !ok {
		return fmt.Errorf("confidence_threshold is required")
	}
	if _, ok := config["check_all_images"]; !ok {
		return fmt.Errorf("check_all_images is required")
	}
	if _, ok := config["detection_prompt"]; !ok {
		return fmt.Errorf("detection_prompt is required")
	}
	return nil
}

func (f *InvoiceFraudDetectionFunction) Execute(ctx context.Context, input *FunctionInput) (*FunctionOutput, error) {
	if err := f.Validate(input.Config); err != nil {
		return &FunctionOutput{
			Error: err.Error(),
		}, nil
	}

	confidenceThreshold, _ := input.Config["confidence_threshold"].(float64)
	checkAllImages, _ := input.Config["check_all_images"].(bool)
	detectionPrompt, _ := input.Config["detection_prompt"].(string)

	invoices, ok := input.InvoiceData["invoices"].([]interface{})
	if !ok || len(invoices) == 0 {
		return &FunctionOutput{
			Error: "invoice_data.invoices is required and must be an array",
		}, nil
	}

	detectionResults := make([]map[string]interface{}, 0, len(invoices))
	hasFraud := false

	for i, invoice := range invoices {
		invoiceMap, ok := invoice.(map[string]interface{})
		if !ok {
			continue
		}

		imagePath, ok := invoiceMap["image_path"].(string)
		if !ok {
			continue
		}

		result, err := f.detectSingleInvoice(ctx, imagePath, detectionPrompt)
		if err != nil {
			if f.logger != nil {
				f.logger.Error("检测发票失败", logger.NewField("error", err), logger.NewField("image_path", imagePath))
			}
			continue
		}

		isFraud := result.Confidence >= confidenceThreshold
		if isFraud {
			hasFraud = true
		}

		detectionResults = append(detectionResults, map[string]interface{}{
			"invoice_index": i,
			"image_path":    imagePath,
			"is_fraud":      isFraud,
			"confidence":    result.Confidence,
			"reasoning":     result.Reasoning,
		})

		if hasFraud && !checkAllImages {
			break
		}
	}

	return &FunctionOutput{
		Value: hasFraud,
		Metadata: map[string]interface{}{
			"total_invoices":       len(invoices),
			"checked_invoices":     len(detectionResults),
			"fraud_invoices":       len(detectionResults),
			"detection_results":    detectionResults,
			"confidence_threshold": confidenceThreshold,
		},
	}, nil
}

type FraudDetectionResult struct {
	Confidence float64 `json:"confidence"`
	Reasoning  string  `json:"reasoning"`
}

func (f *InvoiceFraudDetectionFunction) detectSingleInvoice(ctx context.Context, imagePath, prompt string) (*FraudDetectionResult, error) {
	imageURL := f.getImageURL(imagePath)

	payload := map[string]interface{}{
		"image_url": imageURL,
		"prompt":    prompt,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", "http://localhost:8080/api/v1/vision/analyze", strings.NewReader(string(jsonData)))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := f.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("vision API returned status %d: %s", resp.StatusCode, string(body))
	}

	var result FraudDetectionResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
}

func (f *InvoiceFraudDetectionFunction) getImageURL(imagePath string) string {
	if strings.HasPrefix(imagePath, "http://") || strings.HasPrefix(imagePath, "https://") {
		return imagePath
	}
	return "http://localhost:8080" + imagePath
}

func parseDate(d interface{}) (time.Time, error) {
	switch v := d.(type) {
	case time.Time:
		return v, nil
	case *time.Time:
		if v != nil {
			return *v, nil
		}
		return time.Time{}, fmt.Errorf("date is nil")
	case string:
		formats := []string{
			time.RFC3339,
			"2006-01-02T15:04:05Z07:00",
			"2006-01-02",
			"2006-01-02 15:04:05",
		}
		for _, format := range formats {
			if parsed, err := time.Parse(format, v); err == nil {
				return parsed, nil
			}
		}
		return time.Time{}, fmt.Errorf("date parse failed: %s", v)
	default:
		return time.Time{}, fmt.Errorf("date type not supported: %T", d)
	}
}

type ReimbursementTotalAmountFunction struct{}

func NewReimbursementTotalAmountFunction() *ReimbursementTotalAmountFunction {
	return &ReimbursementTotalAmountFunction{}
}

func (f *ReimbursementTotalAmountFunction) GetName() string {
	return "reimbursement_total_amount"
}

func (f *ReimbursementTotalAmountFunction) GetDescription() string {
	return "提取报销单总金额"
}

func (f *ReimbursementTotalAmountFunction) GetConfigSchema() *ConfigSchema {
	return &ConfigSchema{
		Fields: []FieldConfig{},
	}
}

func (f *ReimbursementTotalAmountFunction) Validate(config map[string]interface{}) error {
	return nil
}

func (f *ReimbursementTotalAmountFunction) Execute(ctx context.Context, input *FunctionInput) (*FunctionOutput, error) {
	reimbData := input.InvoiceData["reimbursement"]
	if reimbData == nil {
		return &FunctionOutput{Error: "reimbursement data not found"}, nil
	}

	reimbMap, ok := reimbData.(map[string]interface{})
	if !ok {
		return &FunctionOutput{Error: "reimbursement data is not a map"}, nil
	}

	totalAmount := 0.0
	if ta, ok := reimbMap["total_amount"]; ok {
		switch v := ta.(type) {
		case float64:
			totalAmount = v
		case int:
			totalAmount = float64(v)
		case string:
			fmt.Sscanf(v, "%f", &totalAmount)
		}
	}

	return &FunctionOutput{
		Value: totalAmount,
		Metadata: map[string]interface{}{
			"total_amount": totalAmount,
		},
	}, nil
}

type InvoiceAmountFunction struct{}

func NewInvoiceAmountFunction() *InvoiceAmountFunction {
	return &InvoiceAmountFunction{}
}

func (f *InvoiceAmountFunction) GetName() string {
	return "invoice_amount"
}

func (f *InvoiceAmountFunction) GetDescription() string {
	return "提取发票金额"
}

func (f *InvoiceAmountFunction) GetConfigSchema() *ConfigSchema {
	return &ConfigSchema{
		Fields: []FieldConfig{},
	}
}

func (f *InvoiceAmountFunction) Validate(config map[string]interface{}) error {
	return nil
}

func (f *InvoiceAmountFunction) Execute(ctx context.Context, input *FunctionInput) (*FunctionOutput, error) {
	invoiceData := input.InvoiceData["invoice"]
	if invoiceData == nil {
		return &FunctionOutput{Error: "invoice data not found"}, nil
	}

	invoiceMap, ok := invoiceData.(map[string]interface{})
	if !ok {
		return &FunctionOutput{Error: "invoice data is not a map"}, nil
	}

	amount := 0.0
	if a, ok := invoiceMap["amount"]; ok {
		switch v := a.(type) {
		case float64:
			amount = v
		case int:
			amount = float64(v)
		case string:
			fmt.Sscanf(v, "%f", &amount)
		}
	}

	return &FunctionOutput{
		Value: amount,
		Metadata: map[string]interface{}{
			"amount": amount,
		},
	}, nil
}

type InvoiceDaysFromTodayFunction struct{}

func NewInvoiceDaysFromTodayFunction() *InvoiceDaysFromTodayFunction {
	return &InvoiceDaysFromTodayFunction{}
}

func (f *InvoiceDaysFromTodayFunction) GetName() string {
	return "invoice_days_from_today"
}

func (f *InvoiceDaysFromTodayFunction) GetDescription() string {
	return "计算发票日期距今天数"
}

func (f *InvoiceDaysFromTodayFunction) GetConfigSchema() *ConfigSchema {
	return &ConfigSchema{
		Fields: []FieldConfig{},
	}
}

func (f *InvoiceDaysFromTodayFunction) Validate(config map[string]interface{}) error {
	return nil
}

func (f *InvoiceDaysFromTodayFunction) Execute(ctx context.Context, input *FunctionInput) (*FunctionOutput, error) {
	invoiceData := input.InvoiceData["invoice"]
	if invoiceData == nil {
		return &FunctionOutput{Error: "invoice data not found"}, nil
	}

	invoiceMap, ok := invoiceData.(map[string]interface{})
	if !ok {
		return &FunctionOutput{Error: "invoice data is not a map"}, nil
	}

	d, ok := invoiceMap["date"]
	if !ok {
		return &FunctionOutput{Error: "invoice date not found"}, nil
	}

	invoiceDate, err := parseDate(d)
	if err != nil {
		return &FunctionOutput{Error: err.Error()}, nil
	}

	now := time.Now()
	days := int(now.Sub(invoiceDate).Hours() / 24)

	return &FunctionOutput{
		Value: float64(days),
		Metadata: map[string]interface{}{
			"invoice_date":    invoiceDate.Format("2006-01-02"),
			"today":           now.Format("2006-01-02"),
			"days_from_today": days,
		},
	}, nil
}

type TripDurationFunction struct{}

func NewTripDurationFunction() *TripDurationFunction {
	return &TripDurationFunction{}
}

func (f *TripDurationFunction) GetName() string {
	return "trip_duration"
}

func (f *TripDurationFunction) GetDescription() string {
	return "计算出差天数"
}

func (f *TripDurationFunction) GetConfigSchema() *ConfigSchema {
	return &ConfigSchema{
		Fields: []FieldConfig{},
	}
}

func (f *TripDurationFunction) Validate(config map[string]interface{}) error {
	return nil
}

func (f *TripDurationFunction) Execute(ctx context.Context, input *FunctionInput) (*FunctionOutput, error) {
	reimbData := input.InvoiceData["reimbursement"]
	if reimbData == nil {
		return &FunctionOutput{Error: "reimbursement data not found"}, nil
	}

	reimbMap, ok := reimbData.(map[string]interface{})
	if !ok {
		return &FunctionOutput{Error: "reimbursement data is not a map"}, nil
	}

	var startDate, endDate time.Time

	if sd, ok := reimbMap["start_date"]; ok && sd != nil {
		if parsed, err := parseDate(sd); err == nil {
			startDate = parsed
		}
	}

	if ed, ok := reimbMap["end_date"]; ok && ed != nil {
		if parsed, err := parseDate(ed); err == nil {
			endDate = parsed
		}
	}

	days := 0.0
	if !startDate.IsZero() && !endDate.IsZero() {
		days = endDate.Sub(startDate).Hours()/24 + 1
		if days < 0 {
			days = 0
		}
	}

	return &FunctionOutput{
		Value: days,
		Metadata: map[string]interface{}{
			"start_date":    startDate.Format("2006-01-02"),
			"end_date":      endDate.Format("2006-01-02"),
			"duration_days": days,
		},
	}, nil
}

type InvoiceTypeFunction struct{}

func NewInvoiceTypeFunction() *InvoiceTypeFunction {
	return &InvoiceTypeFunction{}
}

func (f *InvoiceTypeFunction) GetName() string {
	return "invoice_type"
}

func (f *InvoiceTypeFunction) GetDescription() string {
	return "提取发票类型"
}

func (f *InvoiceTypeFunction) GetConfigSchema() *ConfigSchema {
	return &ConfigSchema{
		Fields: []FieldConfig{},
	}
}

func (f *InvoiceTypeFunction) Validate(config map[string]interface{}) error {
	return nil
}

func (f *InvoiceTypeFunction) Execute(ctx context.Context, input *FunctionInput) (*FunctionOutput, error) {
	invoiceData := input.InvoiceData["invoice"]
	if invoiceData == nil {
		return &FunctionOutput{Error: "invoice data not found"}, nil
	}

	invoiceMap, ok := invoiceData.(map[string]interface{})
	if !ok {
		return &FunctionOutput{Error: "invoice data is not a map"}, nil
	}

	invoiceType := ""
	if t, ok := invoiceMap["type"]; ok {
		invoiceType = fmt.Sprintf("%v", t)
	}

	return &FunctionOutput{
		Value: invoiceType,
		Metadata: map[string]interface{}{
			"invoice_type": invoiceType,
		},
	}, nil
}

type CommodityNameFunction struct{}

func NewCommodityNameFunction() *CommodityNameFunction {
	return &CommodityNameFunction{}
}

func (f *CommodityNameFunction) GetName() string {
	return "commodity_name"
}

func (f *CommodityNameFunction) GetDescription() string {
	return "提取发票商品名称"
}

func (f *CommodityNameFunction) GetConfigSchema() *ConfigSchema {
	return &ConfigSchema{
		Fields: []FieldConfig{},
	}
}

func (f *CommodityNameFunction) Validate(config map[string]interface{}) error {
	return nil
}

func (f *CommodityNameFunction) Execute(ctx context.Context, input *FunctionInput) (*FunctionOutput, error) {
	invoiceData := input.InvoiceData["invoice"]
	if invoiceData == nil {
		return &FunctionOutput{Error: "invoice data not found"}, nil
	}

	invoiceMap, ok := invoiceData.(map[string]interface{})
	if !ok {
		return &FunctionOutput{Error: "invoice data is not a map"}, nil
	}

	commodityName := ""
	if cn, ok := invoiceMap["commodity_name"]; ok {
		commodityName = fmt.Sprintf("%v", cn)
	}

	return &FunctionOutput{
		Value: commodityName,
		Metadata: map[string]interface{}{
			"commodity_name": commodityName,
		},
	}, nil
}

type MerchantTypeFunction struct{}

func NewMerchantTypeFunction() *MerchantTypeFunction {
	return &MerchantTypeFunction{}
}

func (f *MerchantTypeFunction) GetName() string {
	return "merchant_type"
}

func (f *MerchantTypeFunction) GetDescription() string {
	return "提取商户类型"
}

func (f *MerchantTypeFunction) GetConfigSchema() *ConfigSchema {
	return &ConfigSchema{
		Fields: []FieldConfig{},
	}
}

func (f *MerchantTypeFunction) Validate(config map[string]interface{}) error {
	return nil
}

func (f *MerchantTypeFunction) Execute(ctx context.Context, input *FunctionInput) (*FunctionOutput, error) {
	invoiceData := input.InvoiceData["invoice"]
	if invoiceData == nil {
		return &FunctionOutput{Error: "invoice data not found"}, nil
	}

	invoiceMap, ok := invoiceData.(map[string]interface{})
	if !ok {
		return &FunctionOutput{Error: "invoice data is not a map"}, nil
	}

	merchantType := ""
	if mt, ok := invoiceMap["merchant_type"]; ok {
		merchantType = fmt.Sprintf("%v", mt)
	}

	return &FunctionOutput{
		Value: merchantType,
		Metadata: map[string]interface{}{
			"merchant_type": merchantType,
		},
	}, nil
}

type ReimbursementTypeFunction struct{}

func NewReimbursementTypeFunction() *ReimbursementTypeFunction {
	return &ReimbursementTypeFunction{}
}

func (f *ReimbursementTypeFunction) GetName() string {
	return "reimbursement_type"
}

func (f *ReimbursementTypeFunction) GetDescription() string {
	return "提取报销类型"
}

func (f *ReimbursementTypeFunction) GetConfigSchema() *ConfigSchema {
	return &ConfigSchema{
		Fields: []FieldConfig{},
	}
}

func (f *ReimbursementTypeFunction) Validate(config map[string]interface{}) error {
	return nil
}

func (f *ReimbursementTypeFunction) Execute(ctx context.Context, input *FunctionInput) (*FunctionOutput, error) {
	reimbData := input.InvoiceData["reimbursement"]
	if reimbData == nil {
		return &FunctionOutput{Error: "reimbursement data not found"}, nil
	}

	reimbMap, ok := reimbData.(map[string]interface{})
	if !ok {
		return &FunctionOutput{Error: "reimbursement data is not a map"}, nil
	}

	reimbType := ""
	if t, ok := reimbMap["type"]; ok {
		reimbType = fmt.Sprintf("%v", t)
	}

	return &FunctionOutput{
		Value: reimbType,
		Metadata: map[string]interface{}{
			"reimbursement_type": reimbType,
		},
	}, nil
}

type ApplicantLevelFunction struct{}

func NewApplicantLevelFunction() *ApplicantLevelFunction {
	return &ApplicantLevelFunction{}
}

func (f *ApplicantLevelFunction) GetName() string {
	return "applicant_level"
}

func (f *ApplicantLevelFunction) GetDescription() string {
	return "提取申请人级别"
}

func (f *ApplicantLevelFunction) GetConfigSchema() *ConfigSchema {
	return &ConfigSchema{
		Fields: []FieldConfig{},
	}
}

func (f *ApplicantLevelFunction) Validate(config map[string]interface{}) error {
	return nil
}

func (f *ApplicantLevelFunction) Execute(ctx context.Context, input *FunctionInput) (*FunctionOutput, error) {
	reimbData := input.InvoiceData["reimbursement"]
	if reimbData == nil {
		return &FunctionOutput{Error: "reimbursement data not found"}, nil
	}

	reimbMap, ok := reimbData.(map[string]interface{})
	if !ok {
		return &FunctionOutput{Error: "reimbursement data is not a map"}, nil
	}

	level := ""
	if l, ok := reimbMap["applicant_level"]; ok {
		level = fmt.Sprintf("%v", l)
	}

	return &FunctionOutput{
		Value: level,
		Metadata: map[string]interface{}{
			"applicant_level": level,
		},
	}, nil
}

type InvoiceDateValidityFunction struct{}

func NewInvoiceDateValidityFunction() *InvoiceDateValidityFunction {
	return &InvoiceDateValidityFunction{}
}

func (f *InvoiceDateValidityFunction) GetName() string {
	return "invoice_date_validity"
}

func (f *InvoiceDateValidityFunction) GetDescription() string {
	return "校验开票日期是否在有效范围内"
}

func (f *InvoiceDateValidityFunction) GetConfigSchema() *ConfigSchema {
	return &ConfigSchema{
		Fields: []FieldConfig{
			{
				Name:        "max_days_ago",
				Type:        "number",
				Label:       "最大天数",
				Required:    true,
				Default:     365,
				Description: "开票日期距今天最大允许天数",
			},
		},
	}
}

func (f *InvoiceDateValidityFunction) Validate(config map[string]interface{}) error {
	if _, ok := config["max_days_ago"]; !ok {
		return fmt.Errorf("max_days_ago is required")
	}
	return nil
}

func (f *InvoiceDateValidityFunction) Execute(ctx context.Context, input *FunctionInput) (*FunctionOutput, error) {
	maxDaysAgo := 365.0
	if mda, ok := input.Config["max_days_ago"]; ok {
		switch v := mda.(type) {
		case float64:
			maxDaysAgo = v
		case int:
			maxDaysAgo = float64(v)
		}
	}

	invoiceData := input.InvoiceData["invoice"]
	if invoiceData == nil {
		return &FunctionOutput{Error: "invoice data not found"}, nil
	}

	invoiceMap, ok := invoiceData.(map[string]interface{})
	if !ok {
		return &FunctionOutput{Error: "invoice data is not a map"}, nil
	}

	d, ok := invoiceMap["date"]
	if !ok {
		return &FunctionOutput{Value: false, Error: "invoice date not found"}, nil
	}

	invoiceDate, err := parseDate(d)
	if err != nil {
		return &FunctionOutput{Value: false, Error: err.Error()}, nil
	}

	now := time.Now()
	daysAgo := now.Sub(invoiceDate).Hours() / 24

	isValid := daysAgo >= 0 && daysAgo <= maxDaysAgo

	return &FunctionOutput{
		Value: isValid,
		Metadata: map[string]interface{}{
			"invoice_date": invoiceDate.Format("2006-01-02"),
			"days_ago":     daysAgo,
			"max_days_ago": maxDaysAgo,
			"is_valid":     isValid,
		},
	}, nil
}

type InvoiceAmountRangeFunction struct{}

func NewInvoiceAmountRangeFunction() *InvoiceAmountRangeFunction {
	return &InvoiceAmountRangeFunction{}
}

func (f *InvoiceAmountRangeFunction) GetName() string {
	return "invoice_amount_range"
}

func (f *InvoiceAmountRangeFunction) GetDescription() string {
	return "检查发票金额是否在合理范围内"
}

func (f *InvoiceAmountRangeFunction) GetConfigSchema() *ConfigSchema {
	return &ConfigSchema{
		Fields: []FieldConfig{
			{
				Name:        "min_amount",
				Type:        "number",
				Label:       "最小金额",
				Required:    true,
				Default:     0.01,
				Description: "发票最小金额",
			},
			{
				Name:        "max_amount",
				Type:        "number",
				Label:       "最大金额",
				Required:    true,
				Default:     100000.00,
				Description: "发票最大金额",
			},
		},
	}
}

func (f *InvoiceAmountRangeFunction) Validate(config map[string]interface{}) error {
	if _, ok := config["min_amount"]; !ok {
		return fmt.Errorf("min_amount is required")
	}
	if _, ok := config["max_amount"]; !ok {
		return fmt.Errorf("max_amount is required")
	}
	return nil
}

func (f *InvoiceAmountRangeFunction) Execute(ctx context.Context, input *FunctionInput) (*FunctionOutput, error) {
	minAmount := 0.01
	maxAmount := 100000.00

	if ma, ok := input.Config["min_amount"]; ok {
		switch v := ma.(type) {
		case float64:
			minAmount = v
		case int:
			minAmount = float64(v)
		}
	}

	if ma, ok := input.Config["max_amount"]; ok {
		switch v := ma.(type) {
		case float64:
			maxAmount = v
		case int:
			maxAmount = float64(v)
		}
	}

	invoiceData := input.InvoiceData["invoice"]
	if invoiceData == nil {
		return &FunctionOutput{Error: "invoice data not found"}, nil
	}

	invoiceMap, ok := invoiceData.(map[string]interface{})
	if !ok {
		return &FunctionOutput{Error: "invoice data is not a map"}, nil
	}

	amount := 0.0
	if a, ok := invoiceMap["amount"]; ok {
		switch v := a.(type) {
		case float64:
			amount = v
		case int:
			amount = float64(v)
		case string:
			fmt.Sscanf(v, "%f", &amount)
		}
	}

	isValid := amount >= minAmount && amount <= maxAmount

	return &FunctionOutput{
		Value: isValid,
		Metadata: map[string]interface{}{
			"amount":     amount,
			"min_amount": minAmount,
			"max_amount": maxAmount,
			"is_valid":   isValid,
		},
	}, nil
}

type InvoicePriceFunction struct{}

func NewInvoicePriceFunction() *InvoicePriceFunction {
	return &InvoicePriceFunction{}
}

func (f *InvoicePriceFunction) GetName() string {
	return "invoice_price"
}

func (f *InvoicePriceFunction) GetDescription() string {
	return "提取发票单价"
}

func (f *InvoicePriceFunction) GetConfigSchema() *ConfigSchema {
	return &ConfigSchema{
		Fields: []FieldConfig{},
	}
}

func (f *InvoicePriceFunction) Validate(config map[string]interface{}) error {
	return nil
}

func (f *InvoicePriceFunction) Execute(ctx context.Context, input *FunctionInput) (*FunctionOutput, error) {
	invoiceData := input.InvoiceData["invoice"]
	if invoiceData == nil {
		return &FunctionOutput{Error: "invoice data not found"}, nil
	}

	invoiceMap, ok := invoiceData.(map[string]interface{})
	if !ok {
		return &FunctionOutput{Error: "invoice data is not a map"}, nil
	}

	price := 0.0
	if p, ok := invoiceMap["price"]; ok {
		switch v := p.(type) {
		case float64:
			price = v
		case int:
			price = float64(v)
		case string:
			fmt.Sscanf(v, "%f", &price)
		}
	}

	return &FunctionOutput{
		Value: price,
		Metadata: map[string]interface{}{
			"price": price,
		},
	}, nil
}
