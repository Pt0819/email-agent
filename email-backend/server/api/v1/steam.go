// Package v1 API处理器 - Steam
package v1

import (
	"net/http"
	"strconv"

	"email-backend/server/middleware"
	"email-backend/server/pkg/agent"
	"email-backend/server/repository"
	"email-backend/server/service"

	"github.com/gin-gonic/gin"
)

// SteamHandler Steam处理器
type SteamHandler struct {
	steamService *service.SteamService
	agentClient  *agent.Client
}

// NewSteamHandler 创建Steam处理器
func NewSteamHandler(steamService *service.SteamService, agentClient *agent.Client) *SteamHandler {
	return &SteamHandler{steamService: steamService, agentClient: agentClient}
}

// SetupSteamRoutes 注册Steam路由
func SetupSteamRoutes(r *gin.RouterGroup, agentClient *agent.Client, steamRepo *repository.SteamRepository, emailRepo *repository.EmailRepository) {
	h := NewSteamHandler(service.NewSteamService(steamRepo, emailRepo, agentClient), agentClient)

	steam := r.Group("/steam")
	{
		// Steam邮件列表（复用邮件API的category筛选）
		steam.GET("/emails", h.ListSteamEmails)

		// 游戏列表
		steam.GET("/games", h.ListGames)

		// 促销列表
		steam.GET("/deals", h.ListDeals)
		steam.GET("/deals/:id", h.GetDeal)

		// 统计概览
		steam.GET("/stats", h.GetStats)

		// 手动触发Steam邮件提取
		steam.POST("/emails/:id/extract", h.ExtractSteamInfo)
	}
}

// ListSteamEmails 获取Steam分类邮件列表
func (h *SteamHandler) ListSteamEmails(c *gin.Context) {
	// 复用邮件API，通过category筛选Steam邮件
	// steam_categories: steam_promotion, steam_wishlist, steam_news, steam_update
	success(c, gin.H{
		"message": "请使用 GET /api/v1/emails?category=steam_promotion 查询Steam邮件",
		"steam_categories": []string{
			"steam_promotion",
			"steam_wishlist",
			"steam_news",
			"steam_update",
		},
	})
}

// ListGames 获取游戏列表
func (h *SteamHandler) ListGames(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	keyword := c.Query("keyword")

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	userID := middleware.GetUserID(c)
	games, total, err := h.steamService.ListGames(c.Request.Context(), userID, page, pageSize, keyword)
	if err != nil {
		errorResp(c, http.StatusInternalServerError, "获取游戏列表失败")
		return
	}

	success(c, gin.H{
		"list":      games,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// ListDeals 获取促销列表
func (h *SteamHandler) ListDeals(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	sortBy := c.DefaultQuery("sort", "created_at")
	activeOnly := c.DefaultQuery("active", "true") == "true"

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	userID := middleware.GetUserID(c)
	deals, total, err := h.steamService.ListDeals(c.Request.Context(), userID, page, pageSize, sortBy, activeOnly)
	if err != nil {
		errorResp(c, http.StatusInternalServerError, "获取促销列表失败")
		return
	}

	success(c, gin.H{
		"list":      deals,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// GetDeal 获取促销详情
func (h *SteamHandler) GetDeal(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		badRequest(c, "无效的促销ID")
		return
	}

	deal, err := h.steamService.GetDealByID(c.Request.Context(), id)
	if err != nil {
		notFound(c, "促销信息不存在")
		return
	}

	success(c, deal)
}

// GetStats 获取Steam统计概览
func (h *SteamHandler) GetStats(c *gin.Context) {
	userID := middleware.GetUserID(c)
	stats, err := h.steamService.GetSteamStats(c.Request.Context(), userID)
	if err != nil {
		errorResp(c, http.StatusInternalServerError, "获取统计失败")
		return
	}

	success(c, stats)
}

// ExtractSteamInfo 手动触发Steam邮件信息提取
func (h *SteamHandler) ExtractSteamInfo(c *gin.Context) {
	emailID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		badRequest(c, "无效的邮件ID")
		return
	}

	userID := middleware.GetUserID(c)
	if err := h.steamService.ExtractSteamInfo(c.Request.Context(), emailID, userID); err != nil {
		errorResp(c, http.StatusInternalServerError, "Steam信息提取失败: "+err.Error())
		return
	}

	success(c, gin.H{"message": "提取完成"})
}
