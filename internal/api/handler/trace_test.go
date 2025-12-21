package handler

import (
	"bytes"
	"context"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"reimbursement-audit/internal/api/middleware"
	storage "reimbursement-audit/internal/infra/storage/file"
	"reimbursement-audit/internal/pkg/logger"

	"github.com/gin-gonic/gin"
)

func TestTraceIdPropagation(t *testing.T) {
	// 创建context并添加traceId
	ctx := context.Background()
	traceId := "test-trace-id-12345"
	ctx = context.WithValue(ctx, middleware.TraceIdKey, traceId)

	// 从context中获取traceId
	retrievedTraceId := middleware.GetTraceIdFromContext(ctx)

	// 验证traceId是否正确传播
	if retrievedTraceId != traceId {
		t.Errorf("traceId应该正确传播，期望 %s，实际 %s", traceId, retrievedTraceId)
	}
}

func TestTraceIdInResponse(t *testing.T) {
	// 创建Gin测试环境
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// 注册trace中间件
	router.Use(middleware.TraceMiddleware())

	// 创建日志记录器
	loggerImpl, err := logger.NewLogger(logger.DefaultConfig())
	if err != nil {
		t.Fatalf("创建日志记录器失败: %v", err)
	}

	// 注册日志中间件，用于将带有traceId的logger注入到Gin上下文中
	router.Use(middleware.LoggerMiddleware(loggerImpl))

	// 创建文件服务
	tempDir := t.TempDir()
	fileStorage := storage.NewLocalStorage(tempDir, "http://localhost:8080/files")
	fileService := storage.NewService(fileStorage)

	// 创建上传处理器（不使用数据库）
	uploadHandler := NewUploadHandler(nil, fileService)

	// 注册路由
	router.POST("/api/v1/invoices/upload", uploadHandler.UploadInvoices)

	// 准备测试文件
	testFilePath := filepath.Join(tempDir, "test.jpg")
	testFileContent := []byte("test file content")
	err = os.WriteFile(testFilePath, testFileContent, 0644)
	if err != nil {
		t.Fatalf("创建测试文件失败: %v", err)
	}

	// 创建multipart表单
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("invoice", "test.jpg")
	if err != nil {
		t.Fatalf("创建表单文件失败: %v", err)
	}
	_, err = part.Write(testFileContent)
	if err != nil {
		t.Fatalf("写入表单文件内容失败: %v", err)
	}
	err = writer.Close()
	if err != nil {
		t.Fatalf("关闭表单写入器失败: %v", err)
	}

	// 创建HTTP请求
	req, err := http.NewRequest("POST", "/api/v1/invoices/upload", body)
	if err != nil {
		t.Fatalf("创建HTTP请求失败: %v", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// 执行请求
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// 检查响应头中是否包含X-Trace-Id
	traceIdHeader := w.Header().Get("X-Trace-Id")
	if traceIdHeader == "" {
		t.Errorf("响应头中应该包含X-Trace-Id")
	}

	// 检查响应状态码
	if w.Code != http.StatusOK {
		t.Errorf("响应状态码应该是200，实际 %d", w.Code)
	}

	fmt.Printf("测试通过! TraceId: %s\n", traceIdHeader)
}

func TestTraceIdInBatchUpload(t *testing.T) {
	// 创建Gin测试环境
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// 注册trace中间件
	router.Use(middleware.TraceMiddleware())

	// 创建日志记录器
	loggerImpl, err := logger.NewLogger(logger.DefaultConfig())
	if err != nil {
		t.Fatalf("创建日志记录器失败: %v", err)
	}

	// 注册日志中间件，用于将带有traceId的logger注入到Gin上下文中
	router.Use(middleware.LoggerMiddleware(loggerImpl))

	// 创建文件服务
	tempDir := t.TempDir()
	fileStorage := storage.NewLocalStorage(tempDir, "http://localhost:8080/files")
	fileService := storage.NewService(fileStorage)

	// 创建上传处理器（不使用数据库）
	uploadHandler := NewUploadHandler(nil, fileService)

	// 注册路由
	router.POST("/api/v1/invoices/batch-upload", uploadHandler.BatchUpload)

	// 准备测试文件
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// 添加多个文件
	for i := 0; i < 2; i++ {
		testFilePath := filepath.Join(tempDir, fmt.Sprintf("test%d.jpg", i))
		testFileContent := []byte(fmt.Sprintf("test file content %d", i))
		err = os.WriteFile(testFilePath, testFileContent, 0644)
		if err != nil {
			t.Fatalf("创建测试文件失败: %v", err)
		}

		part, err := writer.CreateFormFile("invoices", fmt.Sprintf("test%d.jpg", i))
		if err != nil {
			t.Fatalf("创建表单文件失败: %v", err)
		}
		_, err = part.Write(testFileContent)
		if err != nil {
			t.Fatalf("写入表单文件内容失败: %v", err)
		}
	}

	err = writer.Close()
	if err != nil {
		t.Fatalf("关闭表单写入器失败: %v", err)
	}

	// 创建HTTP请求
	req, err := http.NewRequest("POST", "/api/v1/invoices/batch-upload", body)
	if err != nil {
		t.Fatalf("创建HTTP请求失败: %v", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// 执行请求
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// 检查响应头中是否包含X-Trace-Id
	traceIdHeader := w.Header().Get("X-Trace-Id")
	if traceIdHeader == "" {
		t.Errorf("响应头中应该包含X-Trace-Id")
	}

	// 检查响应状态码
	if w.Code != http.StatusOK {
		t.Errorf("响应状态码应该是200，实际 %d", w.Code)
	}

	fmt.Printf("批量上传测试通过! TraceId: %s\n", traceIdHeader)
}
