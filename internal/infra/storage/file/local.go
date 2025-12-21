// local.go 本地文件存储实现
// 功能点：
// 1. 实现本地文件存储
// 2. 支持文件上传、下载、删除
// 3. 自动创建目录结构

package storage

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"reimbursement-audit/internal/api/middleware"
	"strings"
	"time"
)

// LocalStorage 本地存储实现
type LocalStorage struct {
	basePath string // 基础存储路径
	baseURL  string // 基础访问URL
}

// NewLocalStorage 创建本地存储实例
func NewLocalStorage(basePath, baseURL string) *LocalStorage {
	// 确保基础路径存在
	if err := os.MkdirAll(basePath, 0755); err != nil {
		panic(fmt.Sprintf("创建本地存储目录失败: %v", err))
	}

	return &LocalStorage{
		basePath: basePath,
		baseURL:  baseURL,
	}
}

// UploadFile 上传文件
func (ls *LocalStorage) UploadFile(ctx context.Context, file *multipart.FileHeader, path string) (*FileInfo, error) {
	// 获取traceId用于日志追踪
	traceId := middleware.GetTraceIdFromContext(ctx)

	// 打开上传的文件
	src, err := file.Open()
	if err != nil {
		return nil, fmt.Errorf("打开上传文件失败: %w, traceId: %s", err, traceId)
	}
	defer src.Close()

	// 构建完整文件路径
	fullPath := filepath.Join(ls.basePath, path)

	// 确保目录存在
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("创建目录失败: %w, traceId: %s", err, traceId)
	}

	// 创建目标文件
	dst, err := os.Create(fullPath)
	if err != nil {
		return nil, fmt.Errorf("创建目标文件失败: %w, traceId: %s", err, traceId)
	}
	defer dst.Close()

	// 复制文件内容
	if _, err := io.Copy(dst, src); err != nil {
		return nil, fmt.Errorf("保存文件失败: %w, traceId: %s", err, traceId)
	}

	// 构建文件信息
	fileInfo := &FileInfo{
		ID:         generateFileID(path),
		Name:       file.Filename,
		Size:       file.Size,
		Path:       path,
		URL:        ls.buildURL(path),
		MimeType:   file.Header.Get("Content-Type"),
		UploadedAt: time.Now(),
	}

	return fileInfo, nil
}

// UploadFileFromBytes 从字节数组上传文件
func (ls *LocalStorage) UploadFileFromBytes(ctx context.Context, data []byte, filename, path, mimeType string) (*FileInfo, error) {
	// 构建完整文件路径
	fullPath := filepath.Join(ls.basePath, path)

	// 确保目录存在
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("创建目录失败: %w", err)
	}

	// 创建目标文件
	dst, err := os.Create(fullPath)
	if err != nil {
		return nil, fmt.Errorf("创建目标文件失败: %w", err)
	}
	defer dst.Close()

	// 写入文件内容
	if _, err := dst.Write(data); err != nil {
		return nil, fmt.Errorf("保存文件失败: %w", err)
	}

	// 构建文件信息
	fileInfo := &FileInfo{
		ID:         generateFileID(path),
		Name:       filename,
		Size:       int64(len(data)),
		Path:       path,
		URL:        ls.buildURL(path),
		MimeType:   mimeType,
		UploadedAt: time.Now(),
	}

	return fileInfo, nil
}

// GetFile 获取文件
func (ls *LocalStorage) GetFile(ctx context.Context, path string) (io.ReadCloser, *FileInfo, error) {
	// 构建完整文件路径
	fullPath := filepath.Join(ls.basePath, path)

	// 打开文件
	file, err := os.Open(fullPath)
	if err != nil {
		return nil, nil, fmt.Errorf("打开文件失败: %w", err)
	}

	// 获取文件信息
	stat, err := file.Stat()
	if err != nil {
		file.Close()
		return nil, nil, fmt.Errorf("获取文件信息失败: %w", err)
	}

	// 构建文件信息
	fileInfo := &FileInfo{
		ID:         generateFileID(path),
		Name:       filepath.Base(path),
		Size:       stat.Size(),
		Path:       path,
		URL:        ls.buildURL(path),
		UploadedAt: stat.ModTime(),
	}

	return file, fileInfo, nil
}

// DeleteFile 删除文件
func (ls *LocalStorage) DeleteFile(ctx context.Context, path string) error {
	// 构建完整文件路径
	fullPath := filepath.Join(ls.basePath, path)

	// 删除文件
	if err := os.Remove(fullPath); err != nil {
		return fmt.Errorf("删除文件失败: %w", err)
	}

	return nil
}

// GetFileURL 获取文件访问URL
func (ls *LocalStorage) GetFileURL(ctx context.Context, path string, expires time.Duration) (string, error) {
	// 本地存储不需要生成过期URL，直接返回固定URL
	return ls.buildURL(path), nil
}

// buildURL 构建文件访问URL
func (ls *LocalStorage) buildURL(path string) string {
	if ls.baseURL == "" {
		return ""
	}

	// 确保路径以/开头
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	// 移除baseURL末尾的/
	baseURL := strings.TrimSuffix(ls.baseURL, "/")

	return baseURL + path
}

// generateFileID 生成文件ID
func generateFileID(path string) string {
	// 使用路径的哈希作为文件ID
	return fmt.Sprintf("%x", path)
}
