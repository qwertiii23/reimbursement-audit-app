// upload_request.go 上传请求结构体和参数校验
// 功能点：
// 1. 定义报销单上传请求结构体
// 2. 定义发票上传请求结构体
// 3. 实现参数校验规则
// 4. 支持文件格式和大小校验
// 5. 支持自定义校验规则
// 6. 提供参数绑定和校验方法

package request

import (
	"errors"
	"fmt"
	"mime/multipart"
	"regexp"
	"strings"
	"time"
)

// ReimbursementUploadRequest 报销单上传请求
type ReimbursementUploadRequest struct {
	UserID      string  `json:"user_id" form:"user_id"`           // 用户ID，必填
	UserName    string  `json:"user_name" form:"user_name"`       // 用户姓名，必填
	TotalAmount float64 `json:"total_amount" form:"total_amount"` // 总金额，必填，大于0
	Category    string  `json:"category" form:"category"`         // 报销类别，必填
	Reason      string  `json:"reason" form:"reason"`             // 报销事由，必填
	Department  string  `json:"department" form:"department"`     // 所属部门，可选
	ApplyDate   string  `json:"apply_date" form:"apply_date"`     // 申请日期，可选，格式：YYYY-MM-DD
	ExpenseDate string  `json:"expense_date" form:"expense_date"` // 费用发生日期，可选，格式：YYYY-MM-DD
	Description string  `json:"description" form:"description"`   // 报销描述，可选
}

// InvoiceUploadRequest 发票上传请求
type InvoiceUploadRequest struct {
	ReimbursementID string                `form:"reimbursement_id"` // 报销单ID，必填
	File            *multipart.FileHeader `form:"file"`             // 上传的文件，必填
}

// BatchUploadRequest 批量上传请求
type BatchUploadRequest struct {
	Reimbursements []ReimbursementUploadRequest `json:"reimbursements"` // 报销单列表
	Invoices       []InvoiceUploadRequest       `json:"invoices"`       // 发票列表
}

// FileUploadInfo 文件上传信息
type FileUploadInfo struct {
	File     multipart.File
	Header   *multipart.FileHeader
	FileType string
}

// Validate 校验报销单上传请求
func (r *ReimbursementUploadRequest) Validate() error {
	// 校验用户ID
	if strings.TrimSpace(r.UserID) == "" {
		return errors.New("用户ID不能为空")
	}

	// 校验用户姓名
	if strings.TrimSpace(r.UserName) == "" {
		return errors.New("用户姓名不能为空")
	}

	// 校验总金额
	if r.TotalAmount <= 0 {
		return errors.New("总金额必须大于0")
	}

	// 校验报销类别
	if strings.TrimSpace(r.Category) == "" {
		return errors.New("报销类别不能为空")
	}

	// 校验报销事由
	if strings.TrimSpace(r.Reason) == "" {
		return errors.New("报销事由不能为空")
	}

	// 校验日期格式
	if r.ApplyDate != "" {
		if _, err := time.Parse("2006-01-02", r.ApplyDate); err != nil {
			return errors.New("申请日期格式不正确，应为YYYY-MM-DD")
		}
	}

	if r.ExpenseDate != "" {
		if _, err := time.Parse("2006-01-02", r.ExpenseDate); err != nil {
			return errors.New("费用发生日期格式不正确，应为YYYY-MM-DD")
		}
	}

	return nil
}

// Validate 校验发票上传请求
func (r *InvoiceUploadRequest) Validate() error {
	// 校验报销单ID
	if strings.TrimSpace(r.ReimbursementID) == "" {
		return errors.New("报销单ID不能为空")
	}

	// 校验文件
	if r.File == nil {
		return errors.New("上传文件不能为空")
	}

	return nil
}

// Validate 校验批量上传请求
func (r *BatchUploadRequest) Validate() error {
	// 校验报销单列表
	if len(r.Reimbursements) == 0 && len(r.Invoices) == 0 {
		return errors.New("报销单列表和发票列表不能都为空")
	}

	// 校验每个报销单
	for i, reimbursement := range r.Reimbursements {
		if err := reimbursement.Validate(); err != nil {
			return fmt.Errorf("第%d个报销单校验失败: %v", i+1, err)
		}
	}

	// 校验每个发票
	for i, invoice := range r.Invoices {
		if err := invoice.Validate(); err != nil {
			return fmt.Errorf("第%d个发票校验失败: %v", i+1, err)
		}
	}

	return nil
}

// ValidateFile 校验上传文件
func (f *FileUploadInfo) ValidateFile() error {
	if f.File == nil {
		return errors.New("文件不能为空")
	}

	if f.Header == nil {
		return errors.New("文件头信息不能为空")
	}

	// 校验文件大小（10MB限制）
	const maxFileSize = 10 * 1024 * 1024 // 10MB
	if f.Header.Size > maxFileSize {
		return errors.New("文件大小不能超过10MB")
	}

	// 校验文件格式（支持JPG/PNG/PDF）
	allowedTypes := []string{"image/jpeg", "image/jpg", "image/png", "application/pdf"}
	isAllowedType := false
	for _, allowedType := range allowedTypes {
		if f.Header.Header.Get("Content-Type") == allowedType {
			isAllowedType = true
			break
		}
	}

	if !isAllowedType {
		// 尝试通过文件扩展名校验
		filename := f.Header.Filename
		if filename != "" {
			ext := strings.ToLower(filename[strings.LastIndex(filename, ".")+1:])
			if ext != "jpg" && ext != "jpeg" && ext != "png" && ext != "pdf" {
				return errors.New("文件格式不支持，仅支持JPG、PNG、PDF格式")
			}
		} else {
			return errors.New("文件格式不支持，仅支持JPG、PNG、PDF格式")
		}
	}

	return nil
}

// IsEmpty 检查报销单上传请求是否为空
func (r *ReimbursementUploadRequest) IsEmpty() bool {
	return strings.TrimSpace(r.UserID) == "" &&
		strings.TrimSpace(r.UserName) == "" &&
		r.TotalAmount == 0 &&
		strings.TrimSpace(r.Category) == "" &&
		strings.TrimSpace(r.Reason) == ""
}

// Sanitize 清理和标准化报销单上传请求数据
func (r *ReimbursementUploadRequest) Sanitize() {
	r.UserID = strings.TrimSpace(r.UserID)
	r.UserName = strings.TrimSpace(r.UserName)
	r.Category = strings.TrimSpace(r.Category)
	r.Reason = strings.TrimSpace(r.Reason)
	r.Department = strings.TrimSpace(r.Department)
	r.Description = strings.TrimSpace(r.Description)
}

// IsValidUserID 校验用户ID格式
func IsValidUserID(userID string) bool {
	// 用户ID应为字母数字组合，长度4-20
	pattern := `^[a-zA-Z0-9]{4,20}$`
	matched, _ := regexp.MatchString(pattern, userID)
	return matched
}

// IsValidAmount 校验金额格式（保留两位小数）
func IsValidAmount(amount float64) bool {
	// 金额应为正数，最多两位小数
	return amount > 0 && amount*100 == float64(int(amount*100))
}
