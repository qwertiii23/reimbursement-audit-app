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
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"reimbursement-audit/internal/api/middleware"
	"reimbursement-audit/internal/api/request"
	"reimbursement-audit/internal/api/response"
	"reimbursement-audit/internal/domain/reimbursement"
	storage "reimbursement-audit/internal/infra/storage/file"
	"reimbursement-audit/internal/infra/storage/mysql"
)

// JSONResponse 返回JSON响应的辅助函数
func JSONResponse(c *gin.Context, code int, message string, data interface{}) {
	// 获取traceId
	traceId := middleware.GetTraceId(c)

	// 构建响应数据
	responseData := gin.H{
		"code":    code,
		"message": message,
		"data":    data,
	}

	// 如果有traceId，添加到响应中
	if traceId != "" {
		responseData["trace_id"] = traceId
	}

	c.JSON(http.StatusOK, responseData)
}

// 返回错误响应的辅助函数
func ErrorResponse(c *gin.Context, code int, message string) {
	JSONResponse(c, code, message, nil)
}

// 返回成功响应的辅助函数
func SuccessResponse(c *gin.Context, data interface{}) {
	JSONResponse(c, response.CodeSuccess, "成功", data)
}

// UploadHandler 处理文件上传的结构体
type UploadHandler struct {
	reimbursementRepo reimbursement.Repository
	mysqlClient       *mysql.Client
	fileService       *storage.Service
}

// NewUploadHandler 创建上传处理器实例
func NewUploadHandler(mysqlClient *mysql.Client, fileService *storage.Service) *UploadHandler {
	// 创建报销单仓储实例
	var reimbursementRepo reimbursement.Repository
	if mysqlClient != nil {
		reimbursementRepo = mysql.NewReimbursementRepository(mysqlClient)
	}

	return &UploadHandler{
		reimbursementRepo: reimbursementRepo,
		mysqlClient:       mysqlClient,
		fileService:       fileService,
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
			ErrorResponse(c, response.CodeInvalidParams, "请求参数格式错误: "+err.Error())
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
			ErrorResponse(c, response.CodeInvalidParams, "请求参数格式错误: "+err.Error())
			return
		}
	}

	// 清理和标准化请求数据
	req.Sanitize()

	// 校验请求数据
	if err := req.Validate(); err != nil {
		// 记录参数校验错误日志
		middleware.LogError(c, "请求参数校验失败",
			"error", err.Error(),
			"user_id", req.UserID,
			"total_amount", req.TotalAmount,
			"context", ctx)
		// 返回校验错误响应
		ErrorResponse(c, response.CodeInvalidParams, "参数校验失败: "+err.Error())
		return
	}

	// 生成报销单UUID
	reimbursementID := uuid.New().String()
	middleware.LogDebug(c, "生成报销单UUID",
		"reimbursement_id", reimbursementID,
		"user_id", req.UserID,
		"total_amount", req.TotalAmount)

	// 解析日期字段
	var applyDate, expenseDate time.Time
	var err error

	// 如果提供了申请日期，解析它
	if req.ApplyDate != "" {
		middleware.LogDebug(c, "解析申请日期",
			"apply_date", req.ApplyDate)
		applyDate, err = time.Parse("2006-01-02", req.ApplyDate)
		if err != nil {
			middleware.LogError(c, "申请日期解析失败",
				"error", err.Error(),
				"apply_date", req.ApplyDate)
			ErrorResponse(c, response.CodeInvalidParams, "申请日期格式不正确，应为YYYY-MM-DD")
			return
		}
	} else {
		// 如果没有提供申请日期，使用当前日期
		applyDate = time.Now()
		middleware.LogDebug(c, "使用当前日期作为申请日期")
	}

	// 如果提供了费用发生日期，解析它
	if req.ExpenseDate != "" {
		middleware.LogDebug(c, "解析费用发生日期",
			"expense_date", req.ExpenseDate)
		expenseDate, err = time.Parse("2006-01-02", req.ExpenseDate)
		if err != nil {
			middleware.LogError(c, "费用发生日期解析失败",
				"error", err.Error(),
				"expense_date", req.ExpenseDate)
			ErrorResponse(c, response.CodeInvalidParams, "费用发生日期格式不正确，应为YYYY-MM-DD")
			return
		}
	} else {
		// 如果没有提供费用发生日期，使用申请日期
		expenseDate = applyDate
		middleware.LogDebug(c, "使用申请日期作为费用发生日期")
	}

	// 创建报销单领域模型
	now := time.Now()
	reimbursementModel := &reimbursement.Reimbursement{
		ID:          reimbursementID,
		UserID:      req.UserID,
		UserName:    req.UserName,
		Department:  req.Department,
		Type:        req.Category, // 使用Category作为Type
		Title:       req.Reason,   // 使用Reason作为Title
		Description: req.Description,
		TotalAmount: req.TotalAmount,
		Currency:    "CNY", // 默认使用人民币
		ApplyDate:   applyDate,
		ExpenseDate: expenseDate,
		Status:      "待提交", // 初始状态为"待提交"
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	middleware.LogDebug(c, "创建报销单模型",
		"reimbursement_id", reimbursementID,
		"user_id", req.UserID,
		"total_amount", req.TotalAmount,
		"status", "待提交")

	// 保存报销单到数据库
	if err := h.reimbursementRepo.CreateReimbursement(ctx, reimbursementModel); err != nil {
		middleware.LogError(c, "保存报销单到数据库失败",
			"error", err.Error(),
			"reimbursement_id", reimbursementID)
		// 返回数据库错误响应
		ErrorResponse(c, response.CodeInternalError, "保存报销单失败: "+err.Error())
		return
	}

	middleware.LogInfo(c, "报销单保存成功",
		"reimbursement_id", reimbursementID,
		"user_id", req.UserID)

	// 创建响应数据
	respData := response.NewReimbursementUploadResponse(
		reimbursementID,
		req.UserID,
		req.UserName,
		req.Category,
		req.TotalAmount,
		"待提交",
		now,
	)

	// 返回成功响应
	middleware.LogInfo(c, "报销单上传处理完成",
		"reimbursement_id", reimbursementID,
		"user_id", req.UserID)
	SuccessResponse(c, respData)
}

// UploadInvoices 处理发票图片上传
func (h *UploadHandler) UploadInvoices(c *gin.Context) {
	// 记录请求开始日志
	middleware.LogInfo(c, "开始处理发票单文件上传请求",
		"path", c.Request.URL.Path,
		"method", c.Request.Method,
		"remote_addr", c.ClientIP())

	// 从请求中获取文件
	file, err := c.FormFile("invoice")
	if err != nil {
		middleware.LogError(c, "获取上传文件失败",
			"error", err.Error(),
			"form_field", "invoice")
		ErrorResponse(c, response.CodeInvalidParams, "获取文件失败: "+err.Error())
		return
	}

	middleware.LogDebug(c, "获取到上传文件",
		"filename", file.Filename,
		"size", file.Size,
		"header", file.Header)

	// 上传发票文件
	fileInfo, err := h.fileService.UploadInvoice(c.Request.Context(), file)
	if err != nil {
		middleware.LogError(c, "上传文件失败",
			"error", err.Error(),
			"filename", file.Filename)
		ErrorResponse(c, response.CodeInternalError, "上传文件失败: "+err.Error())
		return
	}

	middleware.LogDebug(c, "文件上传成功",
		"file_id", fileInfo.ID,
		"file_path", fileInfo.Path)

	// 创建发票记录
	invoice := &reimbursement.Invoice{
		ID:        fileInfo.ID,
		ImagePath: fileInfo.Path,
		Status:    "待识别", // 初始状态为待识别，等待OCR处理
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	middleware.LogDebug(c, "创建发票记录",
		"invoice_id", invoice.ID,
		"status", invoice.Status)

	// 保存发票记录到数据库
	if h.reimbursementRepo != nil {
		if err := h.reimbursementRepo.CreateInvoice(c.Request.Context(), invoice); err != nil {
			middleware.LogError(c, "保存发票记录到数据库失败",
				"error", err.Error(),
				"invoice_id", invoice.ID)
			ErrorResponse(c, response.CodeInternalError, "保存发票记录失败: "+err.Error())
			return
		}
	} else {
		middleware.LogDebug(c, "跳过数据库保存，因为repository为空",
			"invoice_id", invoice.ID)
	}

	middleware.LogInfo(c, "发票记录保存成功",
		"invoice_id", invoice.ID,
		"image_path", invoice.ImagePath)

	// 返回成功响应
	middleware.LogInfo(c, "发票上传处理完成",
		"invoice_id", invoice.ID)
	SuccessResponse(c, gin.H{
		"invoice_id":  invoice.ID,
		"image_path":  invoice.ImagePath,
		"file_url":    fileInfo.URL,
		"upload_time": fileInfo.UploadedAt,
	})
}

// BatchUpload 批量上传处理
func (h *UploadHandler) BatchUpload(c *gin.Context) {
	// 记录请求开始日志
	middleware.LogInfo(c, "开始处理发票批量上传请求",
		"path", c.Request.URL.Path,
		"method", c.Request.Method,
		"remote_addr", c.ClientIP())

	// 解析多文件上传
	form, err := c.MultipartForm()
	if err != nil {
		middleware.LogError(c, "解析多文件表单失败",
			"error", err.Error())
		ErrorResponse(c, response.CodeInvalidParams, "解析表单失败: "+err.Error())
		return
	}

	// 获取上传的文件列表
	files := form.File["invoices"]
	if len(files) == 0 {
		middleware.LogWarn(c, "批量上传请求中未找到文件")
		ErrorResponse(c, response.CodeInvalidParams, "未找到上传文件")
		return
	}

	middleware.LogInfo(c, "获取到批量上传文件",
		"file_count", len(files))

	// 限制批量上传数量
	maxBatchSize := 10
	if len(files) > maxBatchSize {
		middleware.LogWarn(c, "批量上传文件数量超过限制",
			"file_count", len(files),
			"max_batch_size", maxBatchSize)
		ErrorResponse(c, response.CodeInvalidParams, fmt.Sprintf("批量上传文件数量不能超过%d个", maxBatchSize))
		return
	}

	// 处理结果
	results := make([]gin.H, 0, len(files))
	var successCount, failureCount int

	// 存储成功上传的发票信息，用于批量保存
	var successfulInvoices []*reimbursement.Invoice

	// 逐个处理文件上传
	for _, fileHeader := range files {
		middleware.LogDebug(c, "处理上传文件",
			"filename", fileHeader.Filename,
			"size", fileHeader.Size)

		// 上传发票文件
		fileInfo, err := h.fileService.UploadInvoice(c.Request.Context(), fileHeader)
		if err != nil {
			middleware.LogError(c, "上传文件失败",
				"error", err.Error(),
				"filename", fileHeader.Filename)
			// 记录失败信息
			results = append(results, gin.H{
				"filename": fileHeader.Filename,
				"success":  false,
				"error":    err.Error(),
			})
			failureCount++
			continue
		}

		middleware.LogDebug(c, "文件上传成功",
			"filename", fileHeader.Filename,
			"file_id", fileInfo.ID)

		// 创建发票记录
		invoice := &reimbursement.Invoice{
			ID:        fileInfo.ID,
			ImagePath: fileInfo.Path,
			Status:    "待识别", // 初始状态为待识别，等待OCR处理
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		// 添加到成功列表，稍后批量保存
		successfulInvoices = append(successfulInvoices, invoice)

		// 记录成功信息（先不保存到数据库）
		results = append(results, gin.H{
			"filename":   fileHeader.Filename,
			"success":    true,
			"invoice_id": invoice.ID,
			"image_path": invoice.ImagePath,
			"file_url":   fileInfo.URL,
		})
		successCount++
	}

	middleware.LogInfo(c, "文件上传处理完成",
		"success_count", successCount,
		"failure_count", failureCount)

	// 批量保存成功的发票记录到数据库
	if len(successfulInvoices) > 0 {
		middleware.LogInfo(c, "开始批量保存发票记录",
			"invoice_count", len(successfulInvoices))

		if h.reimbursementRepo != nil {
			if err := h.reimbursementRepo.CreateInvoices(c.Request.Context(), successfulInvoices); err != nil {
				middleware.LogError(c, "批量保存发票记录失败",
					"error", err.Error(),
					"invoice_count", len(successfulInvoices))
				// 批量保存失败，需要将所有成功标记的记录改为失败
				for i := range results {
					if results[i]["success"].(bool) {
						results[i]["success"] = false
						results[i]["error"] = "批量保存发票记录失败: " + err.Error()
					}
				}
				successCount = 0
				failureCount = len(files)
			} else {
				middleware.LogInfo(c, "批量保存发票记录成功",
					"invoice_count", len(successfulInvoices))
			}
		} else {
			middleware.LogDebug(c, "跳过数据库保存，因为repository为空",
				"invoice_count", len(successfulInvoices))
		}
	}

	// 返回批量处理结果
	middleware.LogInfo(c, "批量上传处理完成",
		"total", len(files),
		"success_count", successCount,
		"failure_count", failureCount)
	SuccessResponse(c, gin.H{
		"total":         len(files),
		"success_count": successCount,
		"failure_count": failureCount,
		"results":       results,
	})
}
