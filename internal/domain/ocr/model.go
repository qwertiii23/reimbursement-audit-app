// model.go OCR领域模型
// 功能点：
// 1. 定义OCR解析的发票信息结构
// 2. 定义OCR配置结构
// 3. 提供领域相关的验证方法

package ocr

import (
	"strconv"
	"time"
)

// InvoiceInfo 发票信息领域模型
type InvoiceInfo struct {
	// 发票基本信息
	InvoiceCode   string `json:"invoice_code"`   // 发票代码
	InvoiceNumber string `json:"invoice_number"` // 发票号码
	InvoiceType   string `json:"invoice_type"`   // 发票类型
	InvoiceDate   string `json:"invoice_date"`   // 开票日期

	// 金额信息
	TotalAmount  float64 `json:"total_amount"`   // 金额合计(不含税)
	TaxAmount    float64 `json:"tax_amount"`     // 税额
	TotalWithTax float64 `json:"total_with_tax"` // 价税合计

	// 购方信息
	BuyerName      string `json:"buyer_name"`       // 购买方名称
	BuyerTaxNumber string `json:"buyer_tax_number"` // 购买方识别号

	// 销方信息
	SellerName      string `json:"seller_name"`       // 销售方名称
	SellerTaxNumber string `json:"seller_tax_number"` // 销售方识别号

	// 校验信息
	CheckCode    string `json:"check_code"`    // 校验码
	PasswordArea string `json:"password_area"` // 密码区

	// 其他信息
	IsValid      bool      `json:"is_valid"`      // 是否有效
	ErrorMessage string    `json:"error_message"` // 错误信息
	RawText      string    `json:"raw_text"`      // OCR原始文本
	ParseTime    time.Time `json:"parse_time"`    // 解析时间
}

// Invoice 发票模型
type Invoice struct {
	ID              string    `json:"id" gorm:"primaryKey;type:varchar(36);column:id"`                                                      // 发票ID
	ReimbursementID string    `json:"reimbursement_id" gorm:"type:varchar(36);not null;index:idx_reimbursement_id;column:reimbursement_id"` // 报销单ID
	Type            string    `json:"type" gorm:"type:varchar(50);column:type"`                                                             // 发票类型(增值税发票/定额发票等)
	Code            string    `json:"code" gorm:"type:varchar(50);column:code"`                                                             // 发票代码
	Number          string    `json:"number" gorm:"type:varchar(50);column:number"`                                                         // 发票号码
	Date            time.Time `json:"date" gorm:"type:date;column:date"`                                                                    // 开票日期
	Amount          float64   `json:"amount" gorm:"type:decimal(10,2);not null;column:amount"`                                              // 发票金额
	TaxAmount       float64   `json:"tax_amount" gorm:"type:decimal(10,2);column:tax_amount"`                                               // 税额
	Payer           string    `json:"payer" gorm:"type:varchar(100);column:payer"`                                                          // 付款方
	Payee           string    `json:"payee" gorm:"type:varchar(100);column:payee"`                                                          // 收款方
	BuyerName       string    `json:"buyer_name" gorm:"type:varchar(100);column:buyer_name"`                                                // 购买方名称
	BuyerTaxNo      string    `json:"buyer_tax_no" gorm:"type:varchar(50);column:buyer_tax_no"`                                             // 购买方税号
	SellerName      string    `json:"seller_name" gorm:"type:varchar(100);column:seller_name"`                                              // 销售方名称
	SellerTaxNo     string    `json:"seller_tax_no" gorm:"type:varchar(50);column:seller_tax_no"`                                           // 销售方税号
	CommodityName   string    `json:"commodity_name" gorm:"type:varchar(200);column:commodity_name"`                                        // 商品名称
	Specification   string    `json:"specification" gorm:"type:varchar(100);column:specification"`                                          // 规格型号
	Unit            string    `json:"unit" gorm:"type:varchar(20);column:unit"`                                                             // 单位
	Quantity        float64   `json:"quantity" gorm:"type:decimal(10,2);column:quantity"`                                                   // 数量
	Price           float64   `json:"price" gorm:"type:decimal(10,2);column:price"`                                                         // 单价
	ImagePath       string    `json:"image_path" gorm:"type:varchar(500);column:image_path"`                                                // 发票图片路径
	OCRResult       string    `json:"ocr_result" gorm:"type:text;column:ocr_result"`                                                        // OCR识别结果
	Status          string    `json:"status" gorm:"type:varchar(20);not null;default:'待识别';column:status"`                                  // 状态(待识别/已识别/识别失败)
	CreatedAt       time.Time `json:"created_at" gorm:"type:datetime;not null;column:created_at"`                                           // 创建时间
	UpdatedAt       time.Time `json:"updated_at" gorm:"type:datetime;not null;column:updated_at"`                                           // 更新时间
}

// Config OCR服务配置
type Config struct {
	// 腾讯云OCR配置
	SecretID  string `json:"secret_id" yaml:"secret_id"`
	SecretKey string `json:"secret_key" yaml:"secret_key"`
	Region    string `json:"region" yaml:"region"`

	// 请求配置
	Timeout    int `json:"timeout" yaml:"timeout"`         // 超时时间(秒)
	MaxRetries int `json:"max_retries" yaml:"max_retries"` // 最大重试次数
}


// Validate 验证发票信息是否有效
func (i *InvoiceInfo) Validate() (bool, string) {
	// 检查必填字段
	if i.InvoiceCode == "" {
		return false, "发票代码为空"
	}
	if i.InvoiceNumber == "" {
		return false, "发票号码为空"
	}
	if i.InvoiceDate == "" {
		return false, "开票日期为空"
	}
	if i.TotalAmount <= 0 {
		return false, "金额无效"
	}

	// 验证发票代码格式（通常为10位或12位数字）
	if !isNumeric(i.InvoiceCode) || (len(i.InvoiceCode) != 10 && len(i.InvoiceCode) != 12) {
		return false, "发票代码格式不正确"
	}

	// 验证发票号码格式（通常为8位数字）
	if !isNumeric(i.InvoiceNumber) || len(i.InvoiceNumber) != 8 {
		return false, "发票号码格式不正确"
	}

	// 验证开票日期格式
	if !isValidDate(i.InvoiceDate) {
		return false, "开票日期格式不正确"
	}

	return true, ""
}

// isNumeric 检查字符串是否只包含数字
func isNumeric(str string) bool {
	for _, c := range str {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}

// isValidDate 检查日期格式是否有效（支持YYYYMMDD和YYYY-MM-DD格式）
func isValidDate(dateStr string) bool {
	// 尝试YYYYMMDD格式
	if len(dateStr) == 8 {
		if year, err := strconv.Atoi(dateStr[:4]); err == nil && year > 0 {
			if month, err := strconv.Atoi(dateStr[4:6]); err == nil && month >= 1 && month <= 12 {
				if day, err := strconv.Atoi(dateStr[6:8]); err == nil && day >= 1 && day <= 31 {
					return true
				}
			}
		}
	}

	// 尝试YYYY-MM-DD格式
	if len(dateStr) == 10 && dateStr[4] == '-' && dateStr[7] == '-' {
		if year, err := strconv.Atoi(dateStr[:4]); err == nil && year > 0 {
			if month, err := strconv.Atoi(dateStr[5:7]); err == nil && month >= 1 && month <= 12 {
				if day, err := strconv.Atoi(dateStr[8:10]); err == nil && day >= 1 && day <= 31 {
					return true
				}
			}
		}
	}

	// 尝试其他常见格式
	formats := []string{
		"20060102",
		"2006-01-02",
		"2006/01/02",
		"2006.01.02",
	}

	for _, format := range formats {
		if _, err := time.Parse(format, dateStr); err == nil {
			return true
		}
	}

	return false
}
