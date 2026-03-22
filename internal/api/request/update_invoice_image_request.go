package request

import (
	"errors"
	"strings"
)

// UpdateInvoiceImageRequest 发票图片修改请求
type UpdateInvoiceImageRequest struct {
	InvoiceID string `json:"invoice_id" binding:"required"` // 发票ID，必填
}

// Validate 校验发票图片修改请求
func (r *UpdateInvoiceImageRequest) Validate() error {
	if strings.TrimSpace(r.InvoiceID) == "" {
		return errors.New("发票ID不能为空")
	}
	return nil
}
