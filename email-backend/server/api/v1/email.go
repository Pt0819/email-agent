// Package v1 API v1版本接口
package v1

import (
	"strconv"

	emailRequest "email-backend/server/model/request"
	respModel "email-backend/server/model/response"
	"email-backend/server/middleware"
	"email-backend/server/pkg/agent"
	"email-backend/server/repository"
	"email-backend/server/service"

	"github.com/gin-gonic/gin"
)

// EmailHandler 邮件处理器
type EmailHandler struct {
	emailService *service.EmailService
	agentClient  *agent.Client
}

// NewEmailHandler 创建邮件处理器
func NewEmailHandler(emailSvc *service.EmailService, agentClient *agent.Client) *EmailHandler {
	return &EmailHandler{emailService: emailSvc, agentClient: agentClient}
}

// SetupEmailRoutes 注册邮件路由
func SetupEmailRoutes(r *gin.RouterGroup, agentClient *agent.Client, emailRepo *repository.EmailRepository) {
	h := NewEmailHandler(service.NewEmailService(emailRepo), agentClient)

	emails := r.Group("/emails")
	{
		emails.GET("", h.ListEmails)
		emails.GET("/:id", h.GetEmail)
		emails.POST("/:id/classify", h.ClassifyEmail)
		emails.PUT("/:id/status", h.UpdateStatus)
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
	accountID, _ := strconv.ParseInt(c.Query("account_id"), 10, 64)

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	req := &emailRequest.ListRequest{
		Page:      page,
		PageSize:  pageSize,
		UserID:    middleware.GetUserID(c),
		AccountID: accountID,
		Category:  c.Query("category"),
		Status:    c.Query("status"),
		Keyword:   c.Query("keyword"),
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

	// 获取邮件内容
	email, err := h.emailService.GetByID(c.Request.Context(), id)
	if err != nil {
		notFound(c, "邮件不存在")
		return
	}

	// 调用Agent进行分类
	classifyReq := &agent.ClassifyRequest{
		EmailID:     idStr,
		Subject:     email.Subject,
		SenderName:  email.SenderName,
		SenderEmail: email.SenderEmail,
		Content:     email.Content,
		ReceivedAt:  email.ReceivedAt.Format("2006-01-02 15:04:05"),
	}

	agentResp, err := h.agentClient.Classify(c.Request.Context(), classifyReq)
	if err != nil {
		errorResp(c, 500, "Agent分类失败: "+err.Error())
		return
	}

	result := &respModel.ClassificationResponse{
		EmailID:    idStr,
		Category:   agentResp.Classification.Category,
		Priority:   agentResp.Classification.Priority,
		Confidence: agentResp.Classification.Confidence,
		Reasoning:  agentResp.Classification.Reasoning,
	}

	// 更新邮件分类
	if err := h.emailService.ClassifyEmail(c.Request.Context(), id, result.Category, result.Priority, result.Confidence, result.Reasoning); err != nil {
		errorResp(c, 500, "更新分类失败")
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

// UpdateStatus 更新邮件状态（标记已读、归档等）
func (h *EmailHandler) UpdateStatus(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		badRequest(c, "无效的邮件ID")
		return
	}

	var req struct {
		Status string `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		badRequest(c, "缺少status参数")
		return
	}

	// 验证status合法性
	validStatuses := map[string]bool{"read": true, "unread": true, "archived": true}
	if !validStatuses[req.Status] {
		badRequest(c, "无效的状态值，允许: read, unread, archived")
		return
	}

	switch req.Status {
	case "read":
		err = h.emailService.MarkAsRead(c.Request.Context(), id)
	case "archived":
		err = h.emailService.ArchiveEmail(c.Request.Context(), id)
	default:
		// 其他状态直接更新
		err = h.emailService.UpdateStatus(c.Request.Context(), id, req.Status)
	}

	if err != nil {
		errorResp(c, 500, "更新状态失败")
		return
	}

	success(c, gin.H{"id": id, "status": req.Status})
}