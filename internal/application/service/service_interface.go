package service

import (
	"context"
	"mime/multipart"

	"reimbursement-audit/internal/api/request"
	"reimbursement-audit/internal/api/response"
	"reimbursement-audit/internal/domain/reimbursement"
)

// ReimbursementApplicationServiceInterface 报销单应用服务接口
type ReimbursementApplicationServiceInterface interface {
	GetReimbursementDetail(ctx context.Context, id string) (*reimbursement.Reimbursement, error)
	CreateReimbursement(ctx context.Context, req *request.ReimbursementUploadRequest) (*response.ReimbursementUploadResponse, error)
	UpdateReimbursement(ctx context.Context, req *request.UpdateReimbursementRequest) (*response.UpdateReimbursementResponse, error)
	UploadInvoice(ctx context.Context, reimbursementID string, category string, fileHeader *multipart.FileHeader) (*response.InvoiceUploadResponse, error)
	BatchUploadInvoices(ctx context.Context, reimbursementID string, category string, fileHeaders []interface{}) (*response.BatchUploadResponse, error)
	ProcessOCRAsync(ctx context.Context, invoiceID string)
	UpdateInvoiceCategory(ctx context.Context, invoiceID string, category string) error
	UpdateInvoiceImage(ctx context.Context, invoiceID string, fileHeader *multipart.FileHeader) (*response.UpdateInvoiceImageResponse, error)
}
