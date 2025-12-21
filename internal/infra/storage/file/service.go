// service.go 文件服务
// 功能点：
// 1. 文件格式和大小校验
// 2. 生成文件UUID
// 3. 处理文件上传和存储

package storage

import (
	"context"
	"fmt"
	"mime/multipart"
	"path/filepath"
	"reimbursement-audit/internal/api/middleware"
	"strings"
	"time"

	"github.com/google/uuid"
)

// Service 文件服务
type Service struct {
	storage Storage // 文件存储接口
}

// NewService 创建文件服务实例
func NewService(storage Storage) *Service {
	return &Service{
		storage: storage,
	}
}

// AllowedFileTypes 允许的文件类型
var AllowedFileTypes = map[string]bool{
	".jpg":  true,
	".jpeg": true,
	".png":  true,
	".pdf":  true,
}

// MaxFileSize 最大文件大小 (10MB)
const MaxFileSize = 10 * 1024 * 1024

// ValidateFile 校验文件
func (s *Service) ValidateFile(file *multipart.FileHeader) error {
	// 检查文件大小
	if file.Size > MaxFileSize {
		return fmt.Errorf("文件大小超过限制，最大允许 %d MB", MaxFileSize/(1024*1024))
	}

	// 检查文件类型
	ext := strings.ToLower(filepath.Ext(file.Filename))
	if !AllowedFileTypes[ext] {
		return fmt.Errorf("不支持的文件类型: %s，仅支持 JPG、PNG、PDF", ext)
	}

	return nil
}

// GenerateFileUUID 生成文件UUID
func (s *Service) GenerateFileUUID() string {
	return uuid.New().String()
}

// GenerateFilePath 生成文件存储路径
func (s *Service) GenerateFilePath(fileID, filename string) string {
	ext := filepath.Ext(filename)
	// 按日期创建目录结构
	date := time.Now().Format("2006/01/02")
	return fmt.Sprintf("invoices/%s/%s%s", date, fileID, ext)
}

// UploadInvoice 上传发票文件
func (s *Service) UploadInvoice(ctx context.Context, file *multipart.FileHeader) (*FileInfo, error) {
	// 获取traceId用于日志追踪
	traceId := middleware.GetTraceIdFromContext(ctx)

	// 校验文件
	if err := s.ValidateFile(file); err != nil {
		return nil, fmt.Errorf("%w, traceId: %s", err, traceId)
	}

	// 生成文件UUID
	fileID := s.GenerateFileUUID()

	// 生成文件存储路径
	filePath := s.GenerateFilePath(fileID, file.Filename)

	// 上传文件
	fileInfo, err := s.storage.UploadFile(ctx, file, filePath)
	if err != nil {
		return nil, fmt.Errorf("上传文件失败: %w, traceId: %s", err, traceId)
	}

	return fileInfo, nil
}
