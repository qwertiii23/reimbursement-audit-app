package handler

import (
	"reimbursement-audit/internal/api/middleware"
	"reimbursement-audit/internal/api/request"
	"reimbursement-audit/internal/api/response"
	"reimbursement-audit/internal/domain/knowledge"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type KnowledgeHandler struct {
	knowledgeService knowledge.KnowledgeService
}

func NewKnowledgeHandler(knowledgeService knowledge.KnowledgeService) *KnowledgeHandler {
	return &KnowledgeHandler{
		knowledgeService: knowledgeService,
	}
}

func (h *KnowledgeHandler) UploadFile(c *gin.Context) {
	middleware.LogInfo(c, "上传知识库文件", "path", c.Request.URL.Path,
		"method", c.Request.Method, "remote_addr", c.ClientIP())
	traceId := middleware.GetTraceId(c)
	ctx := middleware.WithTraceId(c.Request.Context(), traceId)

	fileHeader, err := c.FormFile("file")
	if err != nil {
		middleware.LogError(c, "获取上传文件失败", "error", err.Error(), "context", ctx)
		response.ErrorResponse(c, response.CodeInvalidParams, "请选择要上传的文件")
		return
	}

	file, err := fileHeader.Open()
	if err != nil {
		middleware.LogError(c, "打开文件失败", "error", err.Error(), "context", ctx)
		response.ErrorResponse(c, response.CodeInternalError, "打开文件失败")
		return
	}
	defer file.Close()

	content := make([]byte, fileHeader.Size)
	if _, err := file.Read(content); err != nil {
		middleware.LogError(c, "读取文件内容失败", "error", err.Error(), "context", ctx)
		response.ErrorResponse(c, response.CodeInternalError, "读取文件内容失败")
		return
	}

	fileType := fileHeader.Header.Get("Content-Type")
	if fileType == "" {
		fileType = "application/octet-stream"
	}

	knowledgeFile := &knowledge.KnowledgeFile{
		FileName:      fileHeader.Filename,
		FilePath:      fileHeader.Filename,
		FileType:      fileType,
		FileSize:      fileHeader.Size,
		Category:      c.PostForm("category"),
		Description:   string(content),
		UploadedBy:    c.PostForm("uploaded_by"),
		UploaderName:  c.PostForm("uploader_name"),
		DownloadCount: 0,
		Status:        "active",
	}

	if err := h.knowledgeService.UploadFile(knowledgeFile); err != nil {
		middleware.LogError(c, "上传文件失败", "error", err.Error(), "context", ctx)
		response.ErrorResponse(c, response.CodeInternalError, "上传失败")
		return
	}

	middleware.LogInfo(c, "上传文件成功", "file_id", knowledgeFile.ID, "context", ctx)
	response.SuccessResponse(c, gin.H{"file_id": knowledgeFile.ID})
}

func (h *KnowledgeHandler) GetFiles(c *gin.Context) {
	middleware.LogInfo(c, "获取知识库文件列表", "path", c.Request.URL.Path,
		"method", c.Request.Method, "remote_addr", c.ClientIP())
	traceId := middleware.GetTraceId(c)
	ctx := middleware.WithTraceId(c.Request.Context(), traceId)

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	filter := &knowledge.KnowledgeFileFilter{
		Category: c.Query("category"),
		Status:   c.Query("status"),
		Page:     page,
		PageSize: pageSize,
	}

	files, total, err := h.knowledgeService.GetAllFiles(filter)
	if err != nil {
		middleware.LogError(c, "获取文件列表失败", "error", err.Error(), "context", ctx)
		response.ErrorResponse(c, response.CodeInternalError, "获取文件列表失败")
		return
	}

	fileList := make([]response.KnowledgeFileResponse, len(files))
	for i, file := range files {
		fileList[i] = response.KnowledgeFileResponse{
			ID:            file.ID,
			FileName:      file.FileName,
			FilePath:      file.FilePath,
			FileType:      file.FileType,
			FileSize:      file.FileSize,
			Category:      file.Category,
			Description:   file.Description,
			UploadedBy:    file.UploadedBy,
			UploaderName:  file.UploaderName,
			DownloadCount: file.DownloadCount,
			Status:        file.Status,
			CreatedAt:     file.CreatedAt.Format(time.RFC3339),
			UpdatedAt:     file.UpdatedAt.Format(time.RFC3339),
		}
	}

	middleware.LogInfo(c, "获取文件列表成功", "total", total, "context", ctx)
	response.SuccessResponse(c, response.KnowledgeFilesListResponse{
		List:  fileList,
		Total: total,
	})
}

func (h *KnowledgeHandler) GetFileByID(c *gin.Context) {
	middleware.LogInfo(c, "获取知识库文件详情", "path", c.Request.URL.Path,
		"method", c.Request.Method, "remote_addr", c.ClientIP())
	traceId := middleware.GetTraceId(c)
	ctx := middleware.WithTraceId(c.Request.Context(), traceId)

	id := c.Param("id")
	if id == "" {
		middleware.LogError(c, "文件ID不能为空", "context", ctx)
		response.ErrorResponse(c, response.CodeInvalidParams, "文件ID不能为空")
		return
	}

	file, err := h.knowledgeService.GetFileByID(id)
	if err != nil {
		middleware.LogError(c, "获取文件失败", "error", err.Error(), "context", ctx)
		response.ErrorResponse(c, response.CodeInternalError, "获取文件失败")
		return
	}

	fileResponse := response.KnowledgeFileResponse{
		ID:            file.ID,
		FileName:      file.FileName,
		FilePath:      file.FilePath,
		FileType:      file.FileType,
		FileSize:      file.FileSize,
		Category:      file.Category,
		Description:   file.Description,
		UploadedBy:    file.UploadedBy,
		UploaderName:  file.UploaderName,
		DownloadCount: file.DownloadCount,
		Status:        file.Status,
		CreatedAt:     file.CreatedAt.Format(time.RFC3339),
		UpdatedAt:     file.UpdatedAt.Format(time.RFC3339),
	}

	middleware.LogInfo(c, "获取文件详情成功", "file_id", id, "context", ctx)
	response.SuccessResponse(c, fileResponse)
}

func (h *KnowledgeHandler) UpdateFile(c *gin.Context) {
	middleware.LogInfo(c, "更新知识库文件", "path", c.Request.URL.Path,
		"method", c.Request.Method, "remote_addr", c.ClientIP())
	traceId := middleware.GetTraceId(c)
	ctx := middleware.WithTraceId(c.Request.Context(), traceId)

	var req request.UpdateFileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.LogError(c, "参数绑定失败", "error", err.Error(), "context", ctx)
		response.ErrorResponse(c, response.CodeInvalidParams, "参数错误")
		return
	}

	file, err := h.knowledgeService.GetFileByID(req.ID)
	if err != nil {
		middleware.LogError(c, "获取文件失败", "error", err.Error(), "context", ctx)
		response.ErrorResponse(c, response.CodeInternalError, "获取文件失败")
		return
	}

	if req.FileName != "" {
		file.FileName = req.FileName
	}
	if req.Category != "" {
		file.Category = req.Category
	}
	if req.Description != "" {
		file.Description = req.Description
	}
	if req.Status != "" {
		file.Status = req.Status
	}

	if err := h.knowledgeService.UpdateFile(file); err != nil {
		middleware.LogError(c, "更新文件失败", "error", err.Error(), "context", ctx)
		response.ErrorResponse(c, response.CodeInternalError, "更新失败")
		return
	}

	middleware.LogInfo(c, "更新文件成功", "file_id", req.ID, "context", ctx)
	response.SuccessResponse(c, gin.H{"message": "更新成功"})
}

func (h *KnowledgeHandler) DeleteFile(c *gin.Context) {
	middleware.LogInfo(c, "删除知识库文件", "path", c.Request.URL.Path,
		"method", c.Request.Method, "remote_addr", c.ClientIP())
	traceId := middleware.GetTraceId(c)
	ctx := middleware.WithTraceId(c.Request.Context(), traceId)

	id := c.Param("id")
	if id == "" {
		middleware.LogError(c, "文件ID不能为空", "context", ctx)
		response.ErrorResponse(c, response.CodeInvalidParams, "文件ID不能为空")
		return
	}

	userID := c.GetString("user_id")
	if userID == "" {
		middleware.LogError(c, "用户未登录", "context", ctx)
		response.ErrorResponse(c, response.CodeUnauthorized, "请先登录")
		return
	}

	if err := h.knowledgeService.DeleteFile(id, userID); err != nil {
		middleware.LogError(c, "删除文件失败", "error", err.Error(), "context", ctx)
		response.ErrorResponse(c, response.CodeInternalError, err.Error())
		return
	}

	middleware.LogInfo(c, "删除文件成功", "file_id", id, "context", ctx)
	response.SuccessResponse(c, gin.H{"message": "删除成功"})
}

func (h *KnowledgeHandler) DownloadFile(c *gin.Context) {
	middleware.LogInfo(c, "下载知识库文件", "path", c.Request.URL.Path,
		"method", c.Request.Method, "remote_addr", c.ClientIP())
	traceId := middleware.GetTraceId(c)
	ctx := middleware.WithTraceId(c.Request.Context(), traceId)

	id := c.Param("id")
	if id == "" {
		middleware.LogError(c, "文件ID不能为空", "context", ctx)
		response.ErrorResponse(c, response.CodeInvalidParams, "文件ID不能为空")
		return
	}

	file, err := h.knowledgeService.GetFileByID(id)
	if err != nil {
		middleware.LogError(c, "下载文件失败", "error", err.Error(), "context", ctx)
		response.ErrorResponse(c, response.CodeInternalError, "下载失败")
		return
	}

	c.Header("Content-Disposition", "attachment; filename="+file.FileName)
	c.Header("Content-Type", file.FileType)
	c.Data(200, "application/octet-stream", []byte(file.Description))

	middleware.LogInfo(c, "下载文件成功", "file_id", id, "context", ctx)
}

func (h *KnowledgeHandler) ViewFile(c *gin.Context) {
	middleware.LogInfo(c, "查看知识库文件内容", "path", c.Request.URL.Path,
		"method", c.Request.Method, "remote_addr", c.ClientIP())
	traceId := middleware.GetTraceId(c)
	ctx := middleware.WithTraceId(c.Request.Context(), traceId)

	id := c.Param("id")
	if id == "" {
		middleware.LogError(c, "文件ID不能为空", "context", ctx)
		response.ErrorResponse(c, response.CodeInvalidParams, "文件ID不能为空")
		return
	}

	file, err := h.knowledgeService.GetFileByID(id)
	if err != nil {
		middleware.LogError(c, "获取文件内容失败", "error", err.Error(), "context", ctx)
		response.ErrorResponse(c, response.CodeInternalError, "获取文件内容失败")
		return
	}

	middleware.LogInfo(c, "获取文件内容成功", "file_id", id, "context", ctx)
	response.SuccessResponse(c, gin.H{
		"id":          file.ID,
		"file_name":   file.FileName,
		"file_type":   file.FileType,
		"file_size":   file.FileSize,
		"category":    file.Category,
		"description": file.Description,
		"content":     file.Description,
	})
}

func (h *KnowledgeHandler) RegisterRoutes(r *gin.Engine) {
	api := r.Group("/api/v1/knowledge")
	{
		api.POST("/upload", h.UploadFile)
		api.GET("/files", h.GetFiles)
		api.GET("/files/:id", h.GetFileByID)
		api.PUT("/files/:id", h.UpdateFile)
		api.DELETE("/files/:id", h.DeleteFile)
		api.GET("/download/:id", h.DownloadFile)
		api.GET("/view/:id", h.ViewFile)
	}
}
