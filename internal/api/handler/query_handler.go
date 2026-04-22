// query_handler.go 处理结果查询的控制器
// 功能点：
// 1. 按报销单ID查询审核报告
// 2. 按用户ID查询历史审核记录
// 3. 按时间范围查询审核记录
// 4. 支持分页查询
// 5. 支持条件组合查询
// 6. 返回结构化的审核报告数据

package handler

import (
	"net/http"
	"strconv"

	"reimbursement-audit/internal/api/request"
	"reimbursement-audit/internal/api/response"
	"reimbursement-audit/internal/application/service"
	"reimbursement-audit/internal/domain/audit"
	"reimbursement-audit/internal/domain/reimbursement"
	"reimbursement-audit/internal/pkg/logger"

	"github.com/gin-gonic/gin"
)

// QueryHandler 处理查询请求的结构体
type QueryHandler struct {
	reimbursementService service.ReimbursementApplicationServiceInterface
	reimbursementRepo    reimbursement.Repository
	auditService         audit.ServiceInterface
	logger               logger.Logger
}

// NewQueryHandler 创建查询处理器实例
func NewQueryHandler(
	reimbursementService service.ReimbursementApplicationServiceInterface,
	reimbursementRepo reimbursement.Repository,
	auditService audit.ServiceInterface,
	logger logger.Logger,
) *QueryHandler {
	return &QueryHandler{
		reimbursementService: reimbursementService,
		reimbursementRepo:    reimbursementRepo,
		auditService:         auditService,
		logger:               logger,
	}
}

// GetReimbursementByID 根据报销单ID查询
func (h *QueryHandler) GetReimbursementByID(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		response.ErrorResponse(c, http.StatusBadRequest, "报销单ID不能为空")
		return
	}

	reimbursement, err := h.reimbursementService.GetReimbursementDetail(c.Request.Context(), id)
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "获取报销单详情失败: "+err.Error())
		return
	}

	response.SuccessResponse(c, reimbursement)
}

// GetReimbursementsByUserID 根据用户ID查询
func (h *QueryHandler) GetReimbursementsByUserID(c *gin.Context) {
	userID := c.Query("user_id")
	if userID == "" {
		response.ErrorResponse(c, http.StatusBadRequest, "用户ID不能为空")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "10"))

	if page < 1 {
		page = 1
	}
	if size < 1 || size > 100 {
		size = 10
	}

	title := c.Query("title")
	status := c.Query("status")
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")

	h.logger.WithContext(c.Request.Context()).Info("根据用户ID查询报销单列表",
		logger.NewField("user_id", userID),
		logger.NewField("page", page),
		logger.NewField("size", size),
		logger.NewField("title", title),
		logger.NewField("status", status),
		logger.NewField("start_date", startDate),
		logger.NewField("end_date", endDate))

	var reimbursements []*reimbursement.Reimbursement
	var total int64
	var err error

	if userID == "all" {
		reimbursements, total, err = h.reimbursementRepo.ListAllReimbursementsWithFilters(c.Request.Context(), page, size, title, status, startDate, endDate)
	} else {
		reimbursements, total, err = h.reimbursementRepo.ListReimbursementsByUserIDWithFilters(c.Request.Context(), userID, page, size, title, status, startDate, endDate)
	}

	if err != nil {
		h.logger.WithContext(c.Request.Context()).Error("查询报销单列表失败",
			logger.NewField("error", err.Error()),
			logger.NewField("user_id", userID))
		response.ErrorResponse(c, http.StatusInternalServerError, "查询报销单列表失败: "+err.Error())
		return
	}

	response.JSONResponse(c, 200, "success", gin.H{
		"list":  reimbursements,
		"total": total,
		"page":  page,
		"size":  size,
	})
}

// GetAllReimbursements 获取所有报销单（管理员使用）
func (h *QueryHandler) GetAllReimbursements(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "10"))

	if page < 1 {
		page = 1
	}
	if size < 1 || size > 100 {
		size = 10
	}

	title := c.Query("title")
	status := c.Query("status")
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")

	h.logger.WithContext(c.Request.Context()).Info("获取所有报销单列表",
		logger.NewField("page", page),
		logger.NewField("size", size),
		logger.NewField("title", title),
		logger.NewField("status", status),
		logger.NewField("start_date", startDate),
		logger.NewField("end_date", endDate))

	reimbursements, total, err := h.reimbursementRepo.ListAllReimbursementsWithFilters(c.Request.Context(), page, size, title, status, startDate, endDate)
	if err != nil {
		h.logger.WithContext(c.Request.Context()).Error("查询报销单列表失败",
			logger.NewField("error", err.Error()))
		response.ErrorResponse(c, http.StatusInternalServerError, "查询报销单列表失败: "+err.Error())
		return
	}

	response.JSONResponse(c, 200, "success", gin.H{
		"list":  reimbursements,
		"total": total,
		"page":  page,
		"size":  size,
	})
}

// GetReimbursementsByDateRange 根据时间范围查询
func (h *QueryHandler) GetReimbursementsByDateRange(c *gin.Context) {
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")

	if startDate == "" || endDate == "" {
		response.ErrorResponse(c, http.StatusBadRequest, "开始日期和结束日期不能为空")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "10"))

	if page < 1 {
		page = 1
	}
	if size < 1 || size > 100 {
		size = 10
	}

	h.logger.WithContext(c.Request.Context()).Info("根据时间范围查询报销单列表",
		logger.NewField("start_date", startDate),
		logger.NewField("end_date", endDate),
		logger.NewField("page", page),
		logger.NewField("size", size))

	reimbursements, total, err := h.reimbursementRepo.ListReimbursementsByDateRange(c.Request.Context(), startDate, endDate, page, size)
	if err != nil {
		h.logger.WithContext(c.Request.Context()).Error("查询报销单列表失败",
			logger.NewField("error", err.Error()),
			logger.NewField("start_date", startDate),
			logger.NewField("end_date", endDate))
		response.ErrorResponse(c, http.StatusInternalServerError, "查询报销单列表失败: "+err.Error())
		return
	}

	response.JSONResponse(c, 200, "success", gin.H{
		"list":  reimbursements,
		"total": total,
		"page":  page,
		"size":  size,
	})
}

// GetAuditReport 获取审核报告详情
func (h *QueryHandler) GetAuditReport(c *gin.Context) {
	reimbursementID := c.Param("id")
	if reimbursementID == "" {
		response.ErrorResponse(c, http.StatusBadRequest, "报销单ID不能为空")
		return
	}

	h.logger.WithContext(c.Request.Context()).Info("获取审核报告详情",
		logger.NewField("reimbursement_id", reimbursementID))

	auditResult, err := h.auditService.GetAuditByReimbursementID(c.Request.Context(), reimbursementID)
	if err != nil {
		h.logger.WithContext(c.Request.Context()).Error("获取审核报告失败",
			logger.NewField("error", err.Error()),
			logger.NewField("reimbursement_id", reimbursementID))
		response.ErrorResponse(c, http.StatusInternalServerError, "获取审核报告失败: "+err.Error())
		return
	}

	if auditResult == nil {
		response.ErrorResponse(c, http.StatusNotFound, "审核报告不存在")
		return
	}

	response.SuccessResponse(c, auditResult)
}

// UpdateReimbursement 更新报销单
func (h *QueryHandler) UpdateReimbursement(c *gin.Context) {
	var req request.UpdateReimbursementRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "请求参数格式错误: "+err.Error())
		return
	}

	if err := req.Validate(); err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "参数校验失败: "+err.Error())
		return
	}

	h.logger.WithContext(c.Request.Context()).Info("更新报销单",
		logger.NewField("reimbursement_id", req.ID))

	result, err := h.reimbursementService.UpdateReimbursement(c.Request.Context(), &req)
	if err != nil {
		h.logger.WithContext(c.Request.Context()).Error("更新报销单失败",
			logger.NewField("error", err.Error()),
			logger.NewField("reimbursement_id", req.ID))
		response.ErrorResponse(c, http.StatusInternalServerError, "更新报销单失败: "+err.Error())
		return
	}

	response.SuccessResponse(c, result)
}
