// Package v1 API处理器 - 摘要
package v1

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	emailRequest "email-backend/server/model/request"
	"email-backend/server/model"
	"email-backend/server/pkg/agent"
	"email-backend/server/repository"

	"github.com/gin-gonic/gin"
)

// SummaryHandler 摘要处理器
type SummaryHandler struct {
	emailRepo   *repository.EmailRepository
	agentClient *agent.Client
}

// NewSummaryHandler 创建摘要处理器
func NewSummaryHandler(emailRepo *repository.EmailRepository, agentClient *agent.Client) *SummaryHandler {
	return &SummaryHandler{
		emailRepo:   emailRepo,
		agentClient: agentClient,
	}
}

// SetupSummaryRoutes 注册摘要路由
func SetupSummaryRoutes(r *gin.RouterGroup, agentClient *agent.Client, emailRepo *repository.EmailRepository) {
	h := NewSummaryHandler(emailRepo, agentClient)
	r.GET("/summary/daily", h.DailySummary)
}

// DailySummaryResponse 每日摘要响应
type DailySummaryResponse struct {
	Date            string              `json:"date"`
	TotalEmails     int                 `json:"total_emails"`
	ByCategory      map[string]int      `json:"by_category"`
	ImportantEmails []ImportantEmail    `json:"important_emails"`
	ActionItems     []ActionItemSummary `json:"action_items"`
	SummaryText     string              `json:"summary_text"`
}

// ImportantEmail 重要邮件
type ImportantEmail struct {
	EmailID  string `json:"email_id"`
	Subject  string `json:"subject"`
	Sender   string `json:"sender"`
	Category string `json:"category"`
	Priority string `json:"priority"`
	Summary  string `json:"summary"`
}

// ActionItemSummary 行动项摘要
type ActionItemSummary struct {
	Task     string `json:"task"`
	Priority string `json:"priority"`
}

// DailySummary 获取每日摘要
func (h *SummaryHandler) DailySummary(c *gin.Context) {
	// 解析日期参数
	dateStr := c.DefaultQuery("date", time.Now().Format("2006-01-02"))

	// TODO: 从JWT获取用户ID
	userID := int64(1)

	// 使用List方法获取所有邮件（后续按日期过滤）
	emailReq := &emailRequest.ListRequest{
		UserID:   userID,
		Page:     1,
		PageSize: 100,
	}
	allEmails, _, err := h.emailRepo.List(c.Request.Context(), emailReq)
	if err != nil {
		errorResp(c, http.StatusInternalServerError, "获取邮件失败")
		return
	}

	// 按日期过滤
	startTime, _ := time.ParseInLocation("2006-01-02", dateStr, time.Local)
	endTime := startTime.Add(24 * time.Hour)

	var emails []*model.Email
	for _, e := range allEmails {
		if e.ReceivedAt.After(startTime) && e.ReceivedAt.Before(endTime) {
			emails = append(emails, e)
		}
	}

	// 如果没有邮件，返回空摘要
	if len(emails) == 0 {
		success(c, DailySummaryResponse{
			Date:            dateStr,
			TotalEmails:     0,
			ByCategory:      make(map[string]int),
			ImportantEmails: []ImportantEmail{},
			ActionItems:     []ActionItemSummary{},
			SummaryText:     "今日无邮件",
		})
		return
	}

	// 生成本地摘要（后续可接入Agent进行AI摘要）
	h.respondLocalSummary(c, dateStr, emails)
}

// respondLocalSummary 返回本地生成的摘要
func (h *SummaryHandler) respondLocalSummary(c *gin.Context, dateStr string, emails []*model.Email) {
	byCategory := make(map[string]int)
	var importantEmails []ImportantEmail
	var actionItems []ActionItemSummary

	for _, email := range emails {
		byCategory[email.Category]++

		// 紧急和高优先级邮件标记为重要
		if email.Priority == "critical" || email.Priority == "high" {
			importantEmails = append(importantEmails, ImportantEmail{
				EmailID:  strconv.FormatInt(email.ID, 10),
				Subject:  email.Subject,
				Sender:   email.SenderName,
				Category: email.Category,
				Priority: email.Priority,
				Summary:  email.Subject,
			})
		}

		// work_urgent类别的邮件添加为行动项
		if email.Category == "work_urgent" {
			actionItems = append(actionItems, ActionItemSummary{
				Task:     email.Subject,
				Priority: email.Priority,
			})
		}
	}

	summaryText := fmt.Sprintf("今日共收到%d封邮件", len(emails))
	for cat, count := range byCategory {
		summaryText += fmt.Sprintf("，%s:%d封", cat, count)
	}

	success(c, DailySummaryResponse{
		Date:            dateStr,
		TotalEmails:     len(emails),
		ByCategory:      byCategory,
		ImportantEmails: importantEmails,
		ActionItems:     actionItems,
		SummaryText:     summaryText,
	})
}
