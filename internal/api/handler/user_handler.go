package handler

import (
	"context"
	"reimbursement-audit/internal/api/middleware"
	"reimbursement-audit/internal/api/request"
	"reimbursement-audit/internal/api/response"
	"reimbursement-audit/internal/domain/user"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	userService user.ServiceInterface
}

func NewUserHandler(userService user.ServiceInterface) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

func (h *UserHandler) Login(c *gin.Context) {
	middleware.LogInfo(c, "用户登录请求", "path", c.Request.URL.Path,
		"method", c.Request.Method, "remote_addr", c.ClientIP())
	
	traceId := middleware.GetTraceId(c)
	ctx := middleware.WithTraceId(context.Background(), traceId)

	var req request.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.LogError(c, "JSON数据绑定失败", "error", err.Error(), "context", ctx)
		response.ErrorResponse(c, response.CodeInvalidParams, err.Error())
		return
	}

	loginResp, err := h.userService.Login(ctx, &user.LoginRequest{
		Username: req.Username,
		Password: req.Password,
	})

	if err != nil {
		middleware.LogError(c, "用户登录失败", "error", err.Error(), "context", ctx)
		response.ErrorResponse(c, response.CodeInternalError, err.Error())
		return
	}

	middleware.LogInfo(c, "用户登录成功", "username", req.Username, "context", ctx)
	
	resp := response.LoginResponse{
		Token: loginResp.Token,
		User: response.UserInfo{
			ID:       loginResp.User.ID,
			Username: loginResp.User.Username,
			Email:    loginResp.User.Email,
			RealName: loginResp.User.RealName,
			Role:     loginResp.User.Role,
		},
	}
	
	response.SuccessResponse(c, resp)
}

func (h *UserHandler) GetCurrentUser(c *gin.Context) {
	middleware.LogInfo(c, "获取当前用户信息请求", "path", c.Request.URL.Path,
		"method", c.Request.Method, "remote_addr", c.ClientIP())
	
	traceId := middleware.GetTraceId(c)
	ctx := middleware.WithTraceId(context.Background(), traceId)

	userID, exists := c.Get("user_id")
	if !exists {
		middleware.LogError(c, "获取用户ID失败", "context", ctx)
		response.ErrorResponse(c, response.CodeUnauthorized, "未授权")
		return
	}

	user, err := h.userService.GetUserByID(ctx, userID.(string))
	if err != nil {
		middleware.LogError(c, "获取用户信息失败", "error", err.Error(), "context", ctx)
		response.ErrorResponse(c, response.CodeInternalError, err.Error())
		return
	}

	middleware.LogInfo(c, "获取当前用户信息成功", "user_id", userID, "context", ctx)
	
	resp := response.UserInfo{
		ID:       user.ID,
		Username: user.Username,
		Email:    user.Email,
		RealName: user.RealName,
		Role:     user.Role,
	}
	
	response.SuccessResponse(c, resp)
}
