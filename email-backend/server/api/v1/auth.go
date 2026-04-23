// Package v1 认证API处理器
package v1

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"email-backend/server/middleware"
	emailRequest "email-backend/server/model/request"
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

	data := h.userService.ToUserResponse(user)
	success(c, data)
}

// UpdateProfile 更新用户资料
// @Summary 更新用户资料
// @Tags 认证
// @Accept json
// @Produce json
// @Param request body request.UpdateProfileRequest true "更新资料请求"
// @Success 200 {object} response.Response{data=response.UserResponse}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Router /api/v1/auth/profile [put]
func (h *AuthHandler) UpdateProfile(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		errorResp(c, 401, "unauthorized")
		return
	}

	var req emailRequest.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		badRequest(c, "invalid request: "+err.Error())
		return
	}

	user, err := h.userService.UpdateProfile(c.Request.Context(), userID, req.Username)
	if err != nil {
		if err.Error() == "username already exists" {
			badRequest(c, "用户名已被使用")
		} else {
			errorResp(c, 500, err.Error())
		}
		return
	}

	data := h.userService.ToUserResponse(user)
	success(c, data)
}

// ChangePassword 修改密码
// @Summary 修改密码
// @Tags 认证
// @Accept json
// @Produce json
// @Param request body request.ChangePasswordRequest true "修改密码请求"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Router /api/v1/auth/password [put]
func (h *AuthHandler) ChangePassword(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		errorResp(c, 401, "unauthorized")
		return
	}

	var req emailRequest.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		badRequest(c, "invalid request: "+err.Error())
		return
	}

	// 验证两次密码一致
	if req.NewPassword != req.ConfirmPassword {
		badRequest(c, "两次输入的密码不一致")
		return
	}

	err := h.userService.ChangePassword(c.Request.Context(), userID, req.OldPassword, req.NewPassword)
	if err != nil {
		if err.Error() == "invalid old password" {
			badRequest(c, "旧密码错误")
		} else {
			errorResp(c, 500, err.Error())
		}
		return
	}

	success(c, map[string]string{"message": "密码修改成功"})
}

// UploadAvatar 上传头像
// @Summary 上传头像
// @Tags 认证
// @Accept multipart/form-data
// @Produce json
// @Param avatar formData file true "头像图片"
// @Success 200 {object} response.Response{data=map[string]string}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Router /api/v1/auth/avatar [post]
func (h *AuthHandler) UploadAvatar(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		errorResp(c, 401, "unauthorized")
		return
	}

	// 获取上传的文件
	file, err := c.FormFile("avatar")
	if err != nil {
		badRequest(c, "请选择要上传的头像图片")
		return
	}

	// 验证文件类型
	allowedTypes := map[string]bool{
		"image/jpeg": true,
		"image/png":  true,
		"image/gif":  true,
		"image/webp": true,
	}
	contentType := file.Header.Get("Content-Type")
	if !allowedTypes[contentType] {
		badRequest(c, "仅支持 JPG、PNG、GIF、WebP 格式")
		return
	}

	// 验证文件大小（限制 2MB）
	if file.Size > 2*1024*1024 {
		badRequest(c, "头像图片大小不能超过 2MB")
		return
	}

	// 生成文件名：user_{userID}_{timestamp}.{ext}
	ext := filepath.Ext(file.Filename)
	filename := fmt.Sprintf("user_%d_%d%s", userID, time.Now().Unix(), ext)

	// 保存文件到 uploads/avatars 目录
	uploadDir := "./uploads/avatars"
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		errorResp(c, 500, "failed to create upload directory")
		return
	}

	filePath := filepath.Join(uploadDir, filename)
	if err := c.SaveUploadedFile(file, filePath); err != nil {
		errorResp(c, 500, "failed to save file")
		return
	}

	// 生成访问URL
	avatarURL := fmt.Sprintf("/uploads/avatars/%s", filename)

	// 更新数据库
	if err := h.userService.UpdateAvatar(c.Request.Context(), userID, avatarURL); err != nil {
		errorResp(c, 500, "failed to update avatar")
		return
	}

	// 返回更新后的用户信息
	user, _ := h.userService.GetUserByID(c.Request.Context(), userID)
	data := h.userService.ToUserResponse(user)
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
