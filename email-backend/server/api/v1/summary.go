// Package v1 API处理器 - 摘要
package v1

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	emailRequest "email-backend/server/model/request"
	"email-backend/server/model"
	"email-backend/server/middleware"
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
	return &SummaryHandler{emailRepo: emailRepo, agentClient: agentClient}
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
	CategoryLabels  map[string]string   `json:"category_labels"`
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

// 分类中文标签映射
var categoryLabels = map[string]string{
	"work_urgent":      "紧急工作",
	"work_normal":      "普通工作",
	"personal":         "个人邮件",
	"subscription":     "订阅邮件",
	"notification":     "系统通知",
	"promotion":        "营销推广",
	"spam":             "垃圾邮件",
	"steam_promotion":  "Steam促销",
	"steam_wishlist":   "Steam愿望单",
	"steam_news":       "Steam资讯",
	"steam_update":     "Steam更新",
	"unclassified":     "未分类",
}

// DailySummary 获取每日摘要
func (h *SummaryHandler) DailySummary(c *gin.Context) {
	dateStr := c.DefaultQuery("date", time.Now().Format("2006-01-02"))
	userID := middleware.GetUserID(c)

	// 获取用户当天邮件
	startTime, _ := time.ParseInLocation("2006-01-02", dateStr, time.Local)
	endTime := startTime.Add(24 * time.Hour)

	emailReq := &emailRequest.ListRequest{
		UserID:   userID,
		Page:     1,
		PageSize: 200,
	}
	allEmails, _, err := h.emailRepo.List(c.Request.Context(), emailReq)
	if err != nil {
		errorResp(c, http.StatusInternalServerError, "获取邮件失败")
		return
	}

	// 按日期过滤
	var emails []*model.Email
	for _, e := range allEmails {
		if e.ReceivedAt.After(startTime) && e.ReceivedAt.Before(endTime) {
			emails = append(emails, e)
		}
	}

	// 本地统计（始终生成）
	summary := h.buildLocalSummary(dateStr, emails)

	// 尝试调用Agent生成AI摘要（失败时降级为本地摘要）
	if h.agentClient != nil && len(emails) > 0 {
		h.enrichWithAI(c, &summary, emails)
	}

	success(c, summary)
}

// buildLocalSummary 生成本地统计摘要
func (h *SummaryHandler) buildLocalSummary(dateStr string, emails []*model.Email) DailySummaryResponse {
	byCategory := make(map[string]int)
	labels := make(map[string]string)
	var importantEmails []ImportantEmail
	var actionItems []ActionItemSummary

	for _, email := range emails {
		byCategory[email.Category]++
		if label, ok := categoryLabels[email.Category]; ok {
			labels[email.Category] = label
		}

		// 紧急和高优先级邮件
		if email.Priority == "critical" || email.Priority == "high" {
			importantEmails = append(importantEmails, ImportantEmail{
				EmailID:  strconv.FormatInt(email.ID, 10),
				Subject:  email.Subject,
				Sender:   email.SenderName,
				Category: email.Category,
				Priority: email.Priority,
				Summary:  email.Reasoning,
			})
		}

		// 紧急工作类别添加行动项
		if email.Category == "work_urgent" {
			actionItems = append(actionItems, ActionItemSummary{
				Task:     email.Subject,
				Priority: email.Priority,
			})
		}
	}

	// 生成摘要文本
	summaryText := fmt.Sprintf("今日共收到%d封邮件", len(emails))
	for cat, count := range byCategory {
		label := categoryLabels[cat]
		if label == "" {
			label = cat
		}
		summaryText += fmt.Sprintf("，%s%d封", label, count)
	}

	if len(emails) == 0 {
		summaryText = "今日暂无邮件"
	}

	return DailySummaryResponse{
		Date:            dateStr,
		TotalEmails:     len(emails),
		ByCategory:      byCategory,
		CategoryLabels:  labels,
		ImportantEmails: importantEmails,
		ActionItems:     actionItems,
		SummaryText:     summaryText,
	}
}

// enrichWithAI 调用Agent生成AI摘要，增强本地摘要
func (h *SummaryHandler) enrichWithAI(c *gin.Context, summary *DailySummaryResponse, emails []*model.Email) {
	// 构建邮件数据传给Agent
	emailsData := make([]map[string]interface{}, 0, len(emails))
	emailIDs := make([]string, 0, len(emails))
	for _, e := range emails {
		emailIDs = append(emailIDs, strconv.FormatInt(e.ID, 10))
		emailsData = append(emailsData, map[string]interface{}{
			"email_id":     strconv.FormatInt(e.ID, 10),
			"subject":      e.Subject,
			"sender_name":  e.SenderName,
			"sender_email": e.SenderEmail,
			"category":     e.Category,
			"priority":     e.Priority,
			"content":      truncateContent(e.Content, 200),
		})
	}

	req := &agent.SummaryRequest{
		EmailIDs:   emailIDs,
		EmailsData: emailsData,
		Date:       summary.Date,
	}

	resp, err := h.agentClient.Summary(c.Request.Context(), req)
	if err != nil {
		// Agent调用失败，保持本地摘要不变
		fmt.Printf("AI摘要生成失败(降级为本地摘要): %v\n", err)
		return
	}

	// 用AI结果增强摘要
	if resp.SummaryText != "" {
		summary.SummaryText = resp.SummaryText
	}

	// 合并AI返回的重要邮件（保留本地+AI的并集）
	if len(resp.ImportantEmails) > 0 {
		existingIDs := make(map[string]bool)
		for _, ie := range summary.ImportantEmails {
			existingIDs[ie.EmailID] = true
		}
		for _, ie := range resp.ImportantEmails {
			if !existingIDs[ie.EmailID] {
				summary.ImportantEmails = append(summary.ImportantEmails, ImportantEmail{
					EmailID:  ie.EmailID,
					Subject:  ie.Subject,
					Sender:   ie.Sender,
					Category: ie.Category,
					Priority: ie.Priority,
					Summary:  ie.Summary,
				})
			}
		}
	}

	// 合并AI返回的行动项
	if len(resp.ActionItems) > 0 {
		for _, ai := range resp.ActionItems {
			summary.ActionItems = append(summary.ActionItems, ActionItemSummary{
				Task:     ai.Task,
				Priority: ai.Priority,
			})
		}
	}
}

// truncateContent 截断邮件内容
func truncateContent(content string, maxLen int) string {
	if len(content) <= maxLen {
		return content
	}
	return content[:maxLen] + "..."
}
