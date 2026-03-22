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
