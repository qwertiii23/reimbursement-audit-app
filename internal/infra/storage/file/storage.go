// storage.go 文件存储服务接口
// 功能点：
// 1. 定义文件存储接口
// 2. 支持本地存储和MinIO存储
// 3. 提供文件上传、下载、删除功能

package storage

import (
	"context"
	"io"
	"mime/multipart"
	"time"
)

// FileInfo 文件信息
type FileInfo struct {
	ID       string    `json:"id"`        // 文件ID
	Name     string    `json:"name"`      // 文件名
	Size     int64     `json:"size"`      // 文件大小
	Path     string    `json:"path"`      // 存储路径
	URL      string    `json:"url"`       // 访问URL
	MimeType string    `json:"mime_type"` // MIME类型
	UploadedAt time.Time `json:"uploaded_at"` // 上传时间
}

// Storage 文件存储接口
type Storage interface {
	// UploadFile 上传文件
	UploadFile(ctx context.Context, file *multipart.FileHeader, path string) (*FileInfo, error)
	
	// UploadFileFromBytes 从字节数组上传文件
	UploadFileFromBytes(ctx context.Context, data []byte, filename, path, mimeType string) (*FileInfo, error)
	
	// GetFile 获取文件
	GetFile(ctx context.Context, path string) (io.ReadCloser, *FileInfo, error)
	
	// DeleteFile 删除文件
	DeleteFile(ctx context.Context, path string) error
	
	// GetFileURL 获取文件访问URL
	GetFileURL(ctx context.Context, path string, expires time.Duration) (string, error)
}