// Package v1 认证API处理器
package v1

import (
	"email-backend/server/middleware"
	emailRequest "email-backend/server/model/request"
	emailResponse "email-backend/server/model/response"
	"email-backend/server/service"

	"github.com/gin-gonic/gin"
)

// AuthHandler 认证处理器
type AuthHandler struct {
	userService *service.UserService
}

// NewAuthHandler 创建认证处理器
func NewAuthHandler(userService *service.UserService) *AuthHandler {
	return &AuthHandler{
		userService: userService,
	}
}

// Register 用户注册
// @Summary 用户注册
// @Tags 认证
// @Accept json
// @Produce json
// @Param request body request.RegisterRequest true "注册请求"
// @Success 200 {object} response.Response{data=response.AuthResponse}
// @Failure 400 {object} response.Response
// @Router /api/v1/auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var req emailRequest.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		badRequest(c, "invalid request: "+err.Error())
		return
	}

	// 标准化邮箱
	req.Email = service.NormalizeEmail(req.Email)

	user, token, err := h.userService.Register(c.Request.Context(), &req)
	if err != nil {
		switch err {
		case service.ErrUserExists:
			badRequest(c, "email or username already exists")
		default:
			errorResp(c, 500, "register failed: "+err.Error())
		}
		return
	}

	// 获取token过期时间
	_, expiresAt, _ := h.userService.GenerateToken(user)

	data := h.userService.ToAuthResponse(user, token, expiresAt)
	created(c, data)
}

// Login 用户登录
// @Summary 用户登录
// @Tags 认证
// @Accept json
// @Produce json
// @Param request body request.LoginRequest true "登录请求"
// @Success 200 {object} response.Response{data=response.AuthResponse}
// @Failure 401 {object} response.Response
// @Router /api/v1/auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req emailRequest.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		badRequest(c, "invalid request: "+err.Error())
		return
	}

	// 标准化邮箱
	req.Email = service.NormalizeEmail(req.Email)

	user, token, err := h.userService.Login(c.Request.Context(), &req)
	if err != nil {
		if err == service.ErrInvalidCredentials {
			errorResp(c, 401, "invalid email or password")
		} else {
			errorResp(c, 500, "login failed: "+err.Error())
		}
		return
	}

	// 获取token过期时间
	_, expiresAt, _ := h.userService.GenerateToken(user)

	data := h.userService.ToAuthResponse(user, token, expiresAt)
	success(c, data)
}

// Me 获取当前用户信息
// @Summary 获取当前用户信息
// @Tags 认证
// @Produce json
// @Success 200 {object} response.Response{data=response.UserResponse}
// @Failure 401 {object} response.Response
// @Router /api/v1/auth/me [get]
func (h *AuthHandler) Me(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		errorResp(c, 401, "unauthorized")
		return
	}

	user, err := h.userService.GetUserByID(c.Request.Context(), userID)
	if err != nil {
		errorResp(c, 500, "get user failed: "+err.Error())
		return
	}

	data := &emailResponse.UserResponse{
		ID:        user.ID,
		UserID:    user.UserID,
		Username:  user.Username,
		Email:     user.Email,
		CreatedAt: user.CreatedAt,
	}
	success(c, data)
}

// SetupAuthRoutes 设置认证路由（公开路由，无需JWT）
func SetupAuthRoutes(r *gin.RouterGroup, userService *service.UserService) {
	h := NewAuthHandler(userService)

	auth := r.Group("/auth")
	{
		auth.POST("/register", h.Register)
		auth.POST("/login", h.Login)
		// /auth/me 在router中以JWT中间件保护注册
	}
}
