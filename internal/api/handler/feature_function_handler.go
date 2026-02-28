package handler

import (
	"net/http"

	"reimbursement-audit/internal/api/response"
	"reimbursement-audit/internal/domain/ruleengine"

	"github.com/gin-gonic/gin"
)

type FeatureFunctionHandler struct {
	service *ruleengine.FeatureFunctionService
}

func NewFeatureFunctionHandler(service *ruleengine.FeatureFunctionService) *FeatureFunctionHandler {
	return &FeatureFunctionHandler{
		service: service,
	}
}

func (h *FeatureFunctionHandler) ListFunctions(c *gin.Context) {
	ctx := c.Request.Context()

	functions := h.service.ListFunctions(ctx)

	response.SuccessResponse(c, functions)
}

func (h *FeatureFunctionHandler) GetFunctionSchema(c *gin.Context) {
	ctx := c.Request.Context()

	name := c.Param("name")
	if name == "" {
		response.ErrorResponse(c, http.StatusBadRequest, "特征函数名称不能为空")
		return
	}

	schema, err := h.service.GetFunctionSchema(ctx, name)
	if err != nil {
		response.ErrorResponse(c, http.StatusNotFound, "特征函数不存在")
		return
	}

	response.SuccessResponse(c, schema)
}
