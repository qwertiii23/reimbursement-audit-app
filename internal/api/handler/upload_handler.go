// upload_handler.go 处理报销单和发票上传的控制器
// 功能点：
// 1. 处理报销单JSON/表单数据上传
// 2. 处理发票图片上传（支持单文件/多文件批量上传）
// 3. 文件格式校验和大小限制
// 4. 上传文件临时存储
// 5. 调用OCR服务解析发票信息
// 6. 返回上传结果和初步解析信息

package handler

import (
	"context"

	"github.com/gin-gonic/gin"

	"reimbursement-audit/internal/api/middleware"
	"reimbursement-audit/internal/api/request"
	"reimbursement-audit/internal/api/response"
	"reimbursement-audit/internal/application/service"
)

// UploadHandler 处理文件上传的结构体
type UploadHandler struct {
	reimbursementAppService *service.ReimbursementApplicationService
}

// NewUploadHandler 创建上传处理器实例
func NewUploadHandler(reimbursementAppService *service.ReimbursementApplicationService) *UploadHandler {
	return &UploadHandler{
		reimbursementAppService: reimbursementAppService,
	}
}

// UploadReimbursement 处理报销单上传
// 支持JSON格式和表单格式的报销单数据上传
// 校验必填字段，生成UUID，存储到数据库
func (h *UploadHandler) UploadReimbursement(c *gin.Context) {
	// 记录请求开始日志
	middleware.LogInfo(c, "开始处理报销单上传请求",
		"path", c.Request.URL.Path,
		"method", c.Request.Method,
		"remote_addr", c.ClientIP())

	// 获取traceId
	traceId := middleware.GetTraceId(c)

	// 创建上下文，用于数据库操作
	ctx := middleware.WithTraceId(context.Background(), traceId)

	// 创建报销单上传请求结构体
	var req request.ReimbursementUploadRequest

	// 根据Content-Type判断数据格式
	contentType := c.GetHeader("Content-Type")
	middleware.LogDebug(c, "请求内容类型",
		"content_type", contentType)

	// 处理JSON格式数据
	if contentType == "application/json" {
		middleware.LogDebug(c, "处理JSON格式数据")
		// 绑定JSON数据到请求结构体
		if err := c.ShouldBindJSON(&req); err != nil {
			// 记录参数绑定错误日志
			middleware.LogError(c, "JSON数据绑定失败",
				"error", err.Error(),
				"context", ctx)
			// 返回参数错误响应
			response.ErrorResponse(c, response.CodeInvalidParams, "请求参数格式错误: "+err.Error())
			return
		}
	} else {
		middleware.LogDebug(c, "处理表单格式数据")
		// 处理表单格式数据
		if err := c.ShouldBind(&req); err != nil {
			// 记录参数绑定错误日志
			middleware.LogError(c, "表单数据绑定失败",
				"error", err.Error(),
				"context", ctx)
			// 返回参数错误响应
			response.ErrorResponse(c, response.CodeInvalidParams, "请求参数格式错误: "+err.Error())
			return
		}
	}

	// 调用应用服务处理业务逻辑
	result, err := h.reimbursementAppService.CreateReimbursement(ctx, &req)
	if err != nil {
		middleware.LogError(c, "创建报销单失败",
			"error", err.Error(),
			"user_id", req.UserID,
			"context", ctx)
		response.ErrorResponse(c, response.CodeInternalError, err.Error())
		return
	}

	// 返回成功响应
	middleware.LogInfo(c, "报销单上传处理完成",
		"reimbursement_id", result.ReimbursementID,
		"user_id", req.UserID)
	response.SuccessResponse(c, result)
}

// UploadInvoices 处理发票图片上传
func (h *UploadHandler) UploadInvoices(c *gin.Context) {
	// 记录请求开始日志
	middleware.LogInfo(c, "开始处理发票单文件上传请求",
		"path", c.Request.URL.Path,
		"method", c.Request.Method,
		"remote_addr", c.ClientIP())

	// 获取traceId
	traceId := middleware.GetTraceId(c)

	// 创建上下文，用于数据库操作
	ctx := middleware.WithTraceId(context.Background(), traceId)

	// 从请求中获取文件
	file, err := c.FormFile("invoice")
	if err != nil {
		middleware.LogError(c, "获取上传文件失败",
			"error", err.Error(),
			"form_field", "invoice")
		response.ErrorResponse(c, response.CodeInvalidParams, "获取文件失败: "+err.Error())
		return
	}

	// 从表单中获取reimbursement_id
	reimbursementID := c.PostForm("reimbursement_id")
	if reimbursementID == "" {
		response.ErrorResponse(c, response.CodeInvalidParams, "缺少报销单ID reimbursement_id")
		return
	}

	// 调用应用服务处理业务逻辑
	result, err := h.reimbursementAppService.UploadInvoice(ctx, reimbursementID, file)
	if err != nil {
		middleware.LogError(c, "上传发票失败",
			"error", err.Error(),
			"reimbursement_id", reimbursementID,
			"filename", file.Filename,
			"context", ctx)
		response.ErrorResponse(c, response.CodeInternalError, err.Error())
		return
	}

	// 返回成功响应
	middleware.LogInfo(c, "发票上传处理完成",
		"invoice_id", result.InvoiceID,
		"reimbursement_id", reimbursementID)
	response.SuccessResponse(c, result)
}

// BatchUpload 批量上传处理
func (h *UploadHandler) BatchUpload(c *gin.Context) {
	// 记录请求开始日志
	middleware.LogInfo(c, "开始处理发票批量上传请求",
		"path", c.Request.URL.Path,
		"method", c.Request.Method,
		"remote_addr", c.ClientIP())

	// 获取traceId
	traceId := middleware.GetTraceId(c)

	// 创建上下文，用于数据库操作
	ctx := middleware.WithTraceId(context.Background(), traceId)

	// 解析多文件上传
	form, err := c.MultipartForm()
	if err != nil {
		middleware.LogError(c, "解析多文件表单失败",
			"error", err.Error())
		response.ErrorResponse(c, response.CodeInvalidParams, "解析表单失败: "+err.Error())
		return
	}

	// 获取上传的文件列表
	files := form.File["invoices"]
	if len(files) == 0 {
		middleware.LogWarn(c, "批量上传请求中未找到文件")
		response.ErrorResponse(c, response.CodeInvalidParams, "未找到上传文件")
		return
	}

	// 从表单中获取reimbursement_id
	reimbursementID := c.PostForm("reimbursement_id")
	if reimbursementID == "" {
		response.ErrorResponse(c, response.CodeInvalidParams, "缺少报销单ID reimbursement_id")
		return
	}

	// 调用应用服务处理业务逻辑
	// 转换文件类型以匹配应用服务接口
	fileHeaders := make([]interface{}, len(files))
	for i, file := range files {
		fileHeaders[i] = file
	}

	result, err := h.reimbursementAppService.BatchUploadInvoices(ctx, reimbursementID, fileHeaders)
	if err != nil {
		middleware.LogError(c, "批量上传发票失败",
			"error", err.Error(),
			"reimbursement_id", reimbursementID,
			"file_count", len(files),
			"context", ctx)
		response.ErrorResponse(c, response.CodeInternalError, err.Error())
		return
	}

	// 返回成功响应
	middleware.LogInfo(c, "批量上传处理完成",
		"batch_id", result.BatchID,
		"total", result.TotalCount,
		"success_count", result.SuccessCount,
		"failure_count", result.FailedCount)
	response.SuccessResponse(c, result)
}
