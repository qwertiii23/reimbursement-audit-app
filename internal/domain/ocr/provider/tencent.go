// tencent.go 腾讯云OCR提供商实现
// 功能点：
// 1. 使用腾讯云官方SDK实现OCR API调用
// 2. 处理图片Base64编码
// 3. 使用SDK处理API签名和认证
// 4. 解析OCR响应结果

package provider

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"reimbursement-audit/internal/domain/ocr"
	"reimbursement-audit/internal/pkg/logger"

	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	tccr "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/ocr/v20181119"
)

// TencentProvider 腾讯云OCR提供商
type TencentProvider struct {
	config ocr.Config
	logger logger.Logger
}

// NewTencentProvider 创建腾讯云OCR提供商
func NewTencentProvider(config ocr.Config, logger logger.Logger) *TencentProvider {
	return &TencentProvider{
		config: config,
		logger: logger,
	}
}

// ParseInvoice 解析发票图片
func (p *TencentProvider) ParseInvoice(ctx context.Context, imagePath string) (*ocr.InvoiceInfo, error) {
	p.logger.WithContext(ctx).Info("开始解析发票图片", logger.NewField("image_path", imagePath))

	// 从环境变量获取凭证，优先使用环境变量
	secretID := os.Getenv("TENCENTCLOUD_SECRET_ID")
	secretKey := os.Getenv("TENCENTCLOUD_SECRET_KEY")

	// 如果环境变量不存在，则使用配置中的值
	if secretID == "" {
		secretID = p.config.SecretID
	}
	if secretKey == "" {
		secretKey = p.config.SecretKey
	}

	// 创建凭证
	credential := common.NewCredential(secretID, secretKey)

	// 创建客户端配置
	cpf := profile.NewClientProfile()
	cpf.HttpProfile.Endpoint = "ocr.tencentcloudapi.com"

	// 创建OCR客户端
	client, err := tccr.NewClient(credential, p.config.Region, cpf)
	if err != nil {
		p.logger.WithContext(ctx).Error("创建OCR客户端失败",
			logger.NewField("error", err.Error()),
			logger.NewField("region", p.config.Region))
		return nil, fmt.Errorf("创建OCR客户端失败: %w", err)
	}

	// 读取图片文件并转换为Base64
	imageBase64, err := p.imageToBase64(imagePath)
	if err != nil {
		p.logger.WithContext(ctx).Error("读取图片文件失败",
			logger.NewField("error", err.Error()),
			logger.NewField("image_path", imagePath))
		return nil, fmt.Errorf("读取图片文件失败: %w", err)
	}

	// 创建请求
	request := tccr.NewVatInvoiceOCRRequest()
	request.ImageBase64 = common.StringPtr(imageBase64)

	// 发送请求
	response, err := client.VatInvoiceOCR(request)
	if err != nil {
		p.logger.WithContext(ctx).Error("发送OCR请求失败",
			logger.NewField("error", err.Error()),
			logger.NewField("image_path", imagePath))
		return nil, fmt.Errorf("发送OCR请求失败: %w", err)
	}

	// 解析响应
	invoiceInfo, err := p.parseResponse(response)
	if err != nil {
		p.logger.WithContext(ctx).Error("解析OCR响应失败",
			logger.NewField("error", err.Error()),
			logger.NewField("image_path", imagePath))
		return nil, fmt.Errorf("解析OCR响应失败: %w", err)
	}

	p.logger.WithContext(ctx).Info("发票图片解析成功",
		logger.NewField("image_path", imagePath),
		logger.NewField("invoice_number", invoiceInfo.InvoiceNumber),
		logger.NewField("total_amount", invoiceInfo.TotalAmount))

	return invoiceInfo, nil
}

// imageToBase64 将图片文件转换为Base64编码
func (p *TencentProvider) imageToBase64(imagePath string) (string, error) {
	// 检查文件是否存在
	if _, err := os.Stat(imagePath); os.IsNotExist(err) {
		return "", fmt.Errorf("图片文件不存在: %s", imagePath)
	}

	// 读取图片文件
	imageData, err := os.ReadFile(imagePath)
	if err != nil {
		return "", fmt.Errorf("读取图片文件失败: %w", err)
	}

	// 转换为Base64编码
	base64Str := base64.StdEncoding.EncodeToString(imageData)
	return base64Str, nil
}

// parseResponse 解析OCR响应
func (p *TencentProvider) parseResponse(response *tccr.VatInvoiceOCRResponse) (*ocr.InvoiceInfo, error) {
	// 创建发票信息结构体
	invoiceInfo := &ocr.InvoiceInfo{
		ParseTime: time.Now(),
		IsValid:   true,
		RawText:   p.getRawText(response),
	}

	// 解析发票基本信息
	if response.Response.VatInvoiceInfos != nil {
		for _, item := range response.Response.VatInvoiceInfos {
			if item.Name != nil && item.Value != nil {
				name := *item.Name
				value := *item.Value

				switch name {
				case "发票代码":
					invoiceInfo.InvoiceCode = value
				case "发票号码":
					invoiceInfo.InvoiceNumber = value
				case "发票类型":
					invoiceInfo.InvoiceType = value
				case "开票日期":
					invoiceInfo.InvoiceDate = value
				case "合计金额":
					invoiceInfo.TotalAmount = p.parseFloat(value)
				case "合计税额":
					invoiceInfo.TaxAmount = p.parseFloat(value)
				case "价税合计":
					invoiceInfo.TotalWithTax = p.parseFloat(value)
				case "购买方名称":
					invoiceInfo.BuyerName = value
				case "购买方识别号":
					invoiceInfo.BuyerTaxNumber = value
				case "销售方名称":
					invoiceInfo.SellerName = value
				case "销售方识别号":
					invoiceInfo.SellerTaxNumber = value
				case "校验码":
					invoiceInfo.CheckCode = value
				case "密码区":
					invoiceInfo.PasswordArea = value
				}
			}
		}
	}

	return invoiceInfo, nil
}

// getRawText 获取OCR原始文本
func (p *TencentProvider) getRawText(response *tccr.VatInvoiceOCRResponse) string {
	// 将整个响应转换为JSON字符串作为原始文本
	// 这里简化处理，实际应用中可以根据需要调整
	rawText := fmt.Sprintf("%+v", response.Response)
	p.logger.Debug("获取OCR原始文本", logger.NewField("text_length", len(rawText)))
	return rawText
}

// parseFloat 解析浮点数
func (p *TencentProvider) parseFloat(s string) float64 {
	// 移除可能的逗号和其他非数字字符
	cleaned := strings.ReplaceAll(s, ",", "")
	cleaned = strings.ReplaceAll(cleaned, " ", "")

	result, err := strconv.ParseFloat(cleaned, 64)
	if err != nil {
		return 0
	}
	return result
}
