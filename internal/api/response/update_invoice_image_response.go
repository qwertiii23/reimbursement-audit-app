package response

import "time"

// UpdateInvoiceImageResponse 发票图片修改响应
type UpdateInvoiceImageResponse struct {
	InvoiceID   string    `json:"invoice_id"`   // 发票ID
	FilePath    string    `json:"file_path"`    // 文件存储路径
	FileSize    int64     `json:"file_size"`    // 文件大小
	UpdatedAt   time.Time `json:"updated_at"`   // 更新时间
	OCRStatus   string    `json:"ocr_status"`   // OCR状态
}

// NewUpdateInvoiceImageResponse 创建发票图片修改响应
func NewUpdateInvoiceImageResponse(
	invoiceID, filePath string,
	fileSize int64,
	updatedAt time.Time,
	ocrStatus string,
) *UpdateInvoiceImageResponse {
	return &UpdateInvoiceImageResponse{
		InvoiceID: invoiceID,
		FilePath:  filePath,
		FileSize:  fileSize,
		UpdatedAt: updatedAt,
		OCRStatus: ocrStatus,
	}
}
