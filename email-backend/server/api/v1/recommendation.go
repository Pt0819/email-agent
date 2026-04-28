// Package v1 API处理器 - 游戏推荐
package v1

import (
	"net/http"
	"strconv"

	"email-backend/server/middleware"
	"email-backend/server/model"
	"email-backend/server/service"

	"github.com/gin-gonic/gin"
)

// RecommendationHandler 推荐处理器
type RecommendationHandler struct {
	recService *service.RecommendationService
}

// NewRecommendationHandler 创建推荐处理器
func NewRecommendationHandler(recService *service.RecommendationService) *RecommendationHandler {
	return &RecommendationHandler{recService: recService}
}

// SetupRecommendationRoutes 注册推荐路由
func SetupRecommendationRoutes(r *gin.RouterGroup, recService *service.RecommendationService) {
	h := NewRecommendationHandler(recService)

	recommendations := r.Group("/recommendations")
	{
		// 推荐列表
		recommendations.GET("", h.ListRecommendations)

		// 生成推荐
		recommendations.POST("/generate", h.GenerateRecommendations)

		// 推荐详情
		recommendations.GET("/:id", h.GetRecommendation)

		// 用户反馈
		recommendations.POST("/:id/feedback", h.ProcessFeedback)

		// 推荐统计
		recommendations.GET("/stats", h.GetStats)
	}
}

// ListRecommendations 获取推荐列表
func (h *RecommendationHandler) ListRecommendations(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	status := c.DefaultQuery("status", "all")
	dealOnly := c.DefaultQuery("deal_only", "false") == "true"

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 50 {
		pageSize = 20
	}

	userID := middleware.GetUserID(c)
	result, err := h.recService.ListRecommendations(c.Request.Context(), userID, page, pageSize, status, dealOnly)
	if err != nil {
		errorResp(c, http.StatusInternalServerError, "获取推荐列表失败")
		return
	}

	success(c, result)
}

// GenerateRecommendations 生成推荐
func (h *RecommendationHandler) GenerateRecommendations(c *gin.Context) {
	var req model.RecommendationGenerateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// 使用默认参数
		req = model.RecommendationGenerateRequest{
			MaxCount: 20,
			DealOnly: false,
			MinScore: 50,
		}
	}

	userID := middleware.GetUserID(c)
	result, err := h.recService.GenerateRecommendations(c.Request.Context(), userID, &req)
	if err != nil {
		errorResp(c, http.StatusInternalServerError, "生成推荐失败: "+err.Error())
		return
	}

	success(c, result)
}

// GetRecommendation 获取推荐详情
func (h *RecommendationHandler) GetRecommendation(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		badRequest(c, "无效的推荐ID")
		return
	}

	userID := middleware.GetUserID(c)
	result, err := h.recService.GetRecommendationByID(c.Request.Context(), id, userID)
	if err != nil {
		notFound(c, "推荐不存在")
		return
	}

	success(c, result)
}

// ProcessFeedback 处理用户反馈
func (h *RecommendationHandler) ProcessFeedback(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		badRequest(c, "无效的推荐ID")
		return
	}

	var req model.FeedbackRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		badRequest(c, "请提供反馈动作")
		return
	}

	// 验证action
	validActions := map[string]bool{
		"click": true, "purchase": true, "ignore": true, "like": true, "dislike": true,
	}
	if !validActions[req.Action] {
		badRequest(c, "无效的反馈动作")
		return
	}

	userID := middleware.GetUserID(c)
	if err := h.recService.ProcessFeedback(c.Request.Context(), userID, id, req.Action); err != nil {
		errorResp(c, http.StatusBadRequest, err.Error())
		return
	}

	success(c, gin.H{"message": "反馈已记录"})
}

// GetStats 获取推荐统计
func (h *RecommendationHandler) GetStats(c *gin.Context) {
	userID := middleware.GetUserID(c)
	result, err := h.recService.ListRecommendations(c.Request.Context(), userID, 1, 1, "", false)
	if err != nil {
		errorResp(c, http.StatusInternalServerError, "获取统计失败")
		return
	}

	success(c, result.Stats)
}
