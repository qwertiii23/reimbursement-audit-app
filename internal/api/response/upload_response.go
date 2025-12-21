// upload_response.go 上传响应结构体
// 功能点：
// 1. 定义报销单上传响应结构体
// 2. 定义发票上传响应结构体
// 3. 定义批量上传响应结构体
// 4. 提供响应数据转换方法

package response

import "time"

// ReimbursementUploadResponse 报销单上传响应
type ReimbursementUploadResponse struct {
	ReimbursementID string    `json:"reimbursement_id"` // 报销单ID
	UserID          string    `json:"user_id"`          // 用户ID
	UserName        string    `json:"user_name"`        // 用户姓名
	TotalAmount     float64   `json:"total_amount"`     // 总金额
	Category        string    `json:"category"`         // 报销类别
	Status          string    `json:"status"`           // 状态
	CreatedAt       time.Time `json:"created_at"`       // 创建时间
}

// InvoiceUploadResponse 发票上传响应
type InvoiceUploadResponse struct {
	InvoiceID       string `json:"invoice_id"`        // 发票ID
	ReimbursementID string `json:"reimbursement_id"`  // 报销单ID
	FilePath        string `json:"file_path"`         // 文件存储路径
	FileSize        int64  `json:"file_size"`         // 文件大小
	UploadStatus    string `json:"upload_status"`     // 上传状态
}

// BatchUploadResponse 批量上传响应
type BatchUploadResponse struct {
	BatchID       string                         `json:"batch_id"`        // 批次ID
	TotalCount    int                            `json:"total_count"`     // 总数量
	SuccessCount  int                            `json:"success_count"`   // 成功数量
	FailedCount   int                            `json:"failed_count"`    // 失败数量
	Reimbursements []ReimbursementUploadResponse `json:"reimbursements"`  // 报销单列表
	Invoices      []InvoiceUploadResponse        `json:"invoices"`        // 发票列表
	Errors        []string                       `json:"errors"`          // 错误信息
}

// NewReimbursementUploadResponse 创建报销单上传响应
func NewReimbursementUploadResponse(reimbursementID, userID, userName, category string, 
	totalAmount float64, status string, createdAt time.Time) *ReimbursementUploadResponse {
	return &ReimbursementUploadResponse{
		ReimbursementID: reimbursementID,
		UserID:          userID,
		UserName:        userName,
		TotalAmount:     totalAmount,
		Category:        category,
		Status:          status,
		CreatedAt:       createdAt,
	}
}

// NewInvoiceUploadResponse 创建发票上传响应
func NewInvoiceUploadResponse(invoiceID, reimbursementID, filePath string, 
	fileSize int64, uploadStatus string) *InvoiceUploadResponse {
	return &InvoiceUploadResponse{
		InvoiceID:       invoiceID,
		ReimbursementID: reimbursementID,
		FilePath:        filePath,
		FileSize:        fileSize,
		UploadStatus:    uploadStatus,
	}
}

// NewBatchUploadResponse 创建批量上传响应
func NewBatchUploadResponse(batchID string, totalCount, successCount, failedCount int) *BatchUploadResponse {
	return &BatchUploadResponse{
		BatchID:      batchID,
		TotalCount:   totalCount,
		SuccessCount: successCount,
		FailedCount:  failedCount,
		Reimbursements: make([]ReimbursementUploadResponse, 0),
		Invoices:     make([]InvoiceUploadResponse, 0),
		Errors:       make([]string, 0),
	}
}