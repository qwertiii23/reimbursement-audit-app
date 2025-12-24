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

	// 扩展字段 - 支持更丰富的报销规则
	Category           string    `json:"category" gorm:"type:varchar(50);column:category"`                                     // 发票类别(差旅费/办公费/招待费/培训费等)
	SubCategory        string    `json:"sub_category" gorm:"type:varchar(50);column:sub_category"`                             // 发票子类别(住宿费/交通费/餐饮费等)
	ExpenseType        string    `json:"expense_type" gorm:"type:varchar(50);column:expense_type"`                             // 费用类型(日常/紧急/计划内等)
	PaymentMethod      string    `json:"payment_method" gorm:"type:varchar(50);column:payment_method"`                         // 支付方式(现金/信用卡/公司账户等)
	MerchantType       string    `json:"merchant_type" gorm:"type:varchar(50);column:merchant_type"`                           // 商户类型(酒店/餐厅/航空公司等)
	MerchantCode       string    `json:"merchant_code" gorm:"type:varchar(50);column:merchant_code"`                           // 商户编码
	Location           string    `json:"location" gorm:"type:varchar(100);column:location"`                                    // 消费地点
	City               string    `json:"city" gorm:"type:varchar(50);column:city"`                                             // 消费城市
	Province           string    `json:"province" gorm:"type:varchar(50);column:province"`                                     // 消费省份
	Country            string    `json:"country" gorm:"type:varchar(50);default:'中国';column:country"`                          // 消费国家
	Purpose            string    `json:"purpose" gorm:"type:varchar(200);column:purpose"`                                      // 消费目的
	Description        string    `json:"description" gorm:"type:text;column:description"`                                      // 发票描述
	ProjectCode        string    `json:"project_code" gorm:"type:varchar(50);column:project_code"`                             // 项目编码
	DepartmentCode     string    `json:"department_code" gorm:"type:varchar(50);column:department_code"`                       // 部门编码
	CostCenter         string    `json:"cost_center" gorm:"type:varchar(50);column:cost_center"`                               // 成本中心
	ContractNumber     string    `json:"contract_number" gorm:"type:varchar(50);column:contract_number"`                       // 合同编号
	ApprovalLevel      string    `json:"approval_level" gorm:"type:varchar(20);column:approval_level"`                         // 审批级别(普通/重要/重大)
	IsReimbursable     bool      `json:"is_reimbursable" gorm:"type:boolean;default:true;column:is_reimbursable"`              // 是否可报销
	IsPersonal         bool      `json:"is_personal" gorm:"type:boolean;default:false;column:is_personal"`                     // 是否个人消费
	IsVAT              bool      `json:"is_vat" gorm:"type:boolean;default:false;column:is_vat"`                               // 是否增值税发票
	VATRate            float64   `json:"vat_rate" gorm:"type:decimal(5,2);column:vat_rate"`                                    // 增值税率
	ExchangeRate       float64   `json:"exchange_rate" gorm:"type:decimal(10,4);default:1.0;column:exchange_rate"`             // 汇率
	OriginalAmount     float64   `json:"original_amount" gorm:"type:decimal(10,2);column:original_amount"`                     // 原币金额
	OriginalCurrency   string    `json:"original_currency" gorm:"type:varchar(10);column:original_currency"`                   // 原币种
	ReceiptNumber      string    `json:"receipt_number" gorm:"type:varchar(50);column:receipt_number"`                         // 收据编号
	InvoiceSeries      string    `json:"invoice_series" gorm:"type:varchar(50);column:invoice_series"`                         // 发票系列
	BatchNumber        string    `json:"batch_number" gorm:"type:varchar(50);column:batch_number"`                             // 批次号
	ValidFrom          time.Time `json:"valid_from" gorm:"type:date;column:valid_from"`                                        // 有效期开始
	ValidTo            time.Time `json:"valid_to" gorm:"type:date;column:valid_to"`                                            // 有效期结束
	IsElectronic       bool      `json:"is_electronic" gorm:"type:boolean;default:false;column:is_electronic"`                 // 是否电子发票
	IsDuplicate        bool      `json:"is_duplicate" gorm:"type:boolean;default:false;column:is_duplicate"`                   // 是否重复发票
	RelatedInvoiceID   string    `json:"related_invoice_id" gorm:"type:varchar(36);column:related_invoice_id"`                 // 关联发票ID(红字发票关联)
	VerificationStatus string    `json:"verification_status" gorm:"type:varchar(20);default:'未验证';column:verification_status"` // 验证状态
	VerificationTime   time.Time `json:"verification_time" gorm:"type:datetime;column:verification_time"`                      // 验证时间
	Remarks            string    `json:"remarks" gorm:"type:text;column:remarks"`                                              // 备注
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
