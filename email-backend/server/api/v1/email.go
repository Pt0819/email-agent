// Package v1 API v1版本接口
package v1

import (
	"strconv"

	emailRequest "email-backend/server/model/request"
	respModel "email-backend/server/model/response"
	"email-backend/server/service"

	"github.com/gin-gonic/gin"
)

// EmailHandler 邮件处理器
type EmailHandler struct {
	emailService *service.EmailService
}

// NewEmailHandler 创建邮件处理器
func NewEmailHandler(emailSvc *service.EmailService) *EmailHandler {
	return &EmailHandler{emailService: emailSvc}
}

// SetupEmailRoutes 注册邮件路由
func SetupEmailRoutes(r *gin.RouterGroup) {
	h := NewEmailHandler(service.NewEmailService(nil))

	emails := r.Group("/emails")
	{
		emails.GET("", h.ListEmails)
		emails.GET("/:id", h.GetEmail)
		emails.POST("/:id/classify", h.ClassifyEmail)
		emails.DELETE("/:id", h.DeleteEmail)
	}
}

// HealthCheck 健康检查
func HealthCheck(c *gin.Context) {
	success(c, gin.H{
		"status":  "ok",
		"service": "email-backend",
		"version": "1.0.0",
	})
}

// ListEmails 获取邮件列表
func (h *EmailHandler) ListEmails(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	req := &emailRequest.ListRequest{
		Page:     page,
		PageSize: pageSize,
		Category: c.Query("category"),
		Status:   c.Query("status"),
	}

	emails, total, err := h.emailService.List(c.Request.Context(), req)
	if err != nil {
		errorResp(c, 500, err.Error())
		return
	}

	success(c, &respModel.EmailListResponse{
		List:     emails,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	})
}

// GetEmail 获取邮件详情
func (h *EmailHandler) GetEmail(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		badRequest(c, "无效的邮件ID")
		return
	}

	email, err := h.emailService.GetByID(c.Request.Context(), id)
	if err != nil {
		notFound(c, "邮件不存在")
		return
	}

	success(c, email)
}

// ClassifyEmail 分类邮件
func (h *EmailHandler) ClassifyEmail(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		badRequest(c, "无效的邮件ID")
		return
	}

	// TODO: 调用Agent服务进行分类
	result := &respModel.ClassificationResponse{
		EmailID:    idStr,
		Category:   "work_normal",
		Priority:   "medium",
		Confidence: 0.85,
		Reasoning:  "基于内容分析判断为普通工作邮件",
	}

	// 更新邮件分类
	if err := h.emailService.ClassifyEmail(c.Request.Context(), id, result.Category, result.Priority, result.Confidence); err != nil {
		errorResp(c, 500, "分类失败")
		return
	}

	success(c, result)
}

// DeleteEmail 删除邮件
func (h *EmailHandler) DeleteEmail(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		badRequest(c, "无效的邮件ID")
		return
	}

	if err := h.emailService.Delete(c.Request.Context(), id); err != nil {
		errorResp(c, 500, "删除失败")
		return
	}

	success(c, nil)
}