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

	"reimbursement-audit/internal/application/service"

	"github.com/gin-gonic/gin"
)

// QueryHandler 处理查询请求的结构体
type QueryHandler struct {
	reimbursementService *service.ReimbursementApplicationService
}

// NewQueryHandler 创建查询处理器实例
func NewQueryHandler(reimbursementService *service.ReimbursementApplicationService) *QueryHandler {
	return &QueryHandler{
		reimbursementService: reimbursementService,
	}
}

// GetReimbursementByID 根据报销单ID查询
func (h *QueryHandler) GetReimbursementByID(c *gin.Context) {
	// 获取路径参数
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    http.StatusBadRequest,
			"message": "报销单ID不能为空",
			"data":    nil,
		})
		return
	}

	// 调用应用服务获取报销单详情
	reimbursement, err := h.reimbursementService.GetReimbursementDetail(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    http.StatusInternalServerError,
			"message": "获取报销单详情失败: " + err.Error(),
			"data":    nil,
		})
		return
	}

	// 返回结果
	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data":    reimbursement,
	})
}

// GetReimbursementsByUserID 根据用户ID查询
func (h *QueryHandler) GetReimbursementsByUserID(w http.ResponseWriter, r *http.Request) {
	// TODO: 实现根据用户ID查询报销单列表逻辑
}

// GetReimbursementsByDateRange 根据时间范围查询
func (h *QueryHandler) GetReimbursementsByDateRange(w http.ResponseWriter, r *http.Request) {
	// TODO: 实现根据时间范围查询报销单列表逻辑
}

// GetAuditReport 获取审核报告详情
func (h *QueryHandler) GetAuditReport(w http.ResponseWriter, r *http.Request) {
	// TODO: 实现获取审核报告详情逻辑
}
