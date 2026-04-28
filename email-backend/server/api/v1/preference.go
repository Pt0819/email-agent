// Package v1 API处理器 - 用户偏好分析
package v1

import (
	"net/http"
	"strconv"

	"email-backend/server/middleware"
	"email-backend/server/service"

	"github.com/gin-gonic/gin"
)

// PreferenceHandler 偏好分析处理器
type PreferenceHandler struct {
	preferenceService *service.PreferenceService
}

// NewPreferenceHandler 创建偏好分析处理器
func NewPreferenceHandler(preferenceService *service.PreferenceService) *PreferenceHandler {
	return &PreferenceHandler{preferenceService: preferenceService}
}

// SetupPreferenceRoutes 注册偏好分析路由
func SetupPreferenceRoutes(r *gin.RouterGroup, preferenceService *service.PreferenceService) {
	h := NewPreferenceHandler(preferenceService)

	steam := r.Group("/steam/profile")
	{
		// 获取用户偏好画像
		steam.GET("/preference", h.GetUserProfile)

		// 触发偏好分析
		steam.POST("/analyze", h.AnalyzePreferences)

		// 获取洞察列表
		steam.GET("/insights", h.GetInsights)
	}
}

// GetUserProfile 获取用户偏好画像
// @Summary 获取用户偏好画像
// @Description 获取用户的游戏偏好聚合画像，包括偏好标签、游戏类型、近期活动等
// @Tags Preference
// @Produce json
// @Success 200 {object} model.UserGamingProfile
// @Failure 500 {object} response.Response
// @Router /api/v1/steam/profile/preference [get]
func (h *PreferenceHandler) GetUserProfile(c *gin.Context) {
	userID := middleware.GetUserID(c)

	profile, err := h.preferenceService.GetUserProfile(c.Request.Context(), userID)
	if err != nil {
		errorResp(c, http.StatusInternalServerError, "获取用户画像失败: "+err.Error())
		return
	}

	success(c, profile)
}

// AnalyzePreferences 触发偏好分析
// @Summary 触发偏好分析
// @Description 调用AI Agent分析用户游戏偏好，更新偏好标签并生成洞察
// @Tags Preference
// @Produce json
// @Success 200 {object} model.PreferenceAnalysisResult
// @Failure 500 {object} response.Response
// @Router /api/v1/steam/profile/analyze [post]
func (h *PreferenceHandler) AnalyzePreferences(c *gin.Context) {
	userID := middleware.GetUserID(c)

	result, err := h.preferenceService.AnalyzePreferences(c.Request.Context(), userID)
	if err != nil {
		errorResp(c, http.StatusInternalServerError, "偏好分析失败: "+err.Error())
		return
	}

	success(c, result)
}

// GetInsights 获取洞察列表
// @Summary 获取洞察列表
// @Description 分页获取用户的偏好分析洞察记录
// @Tags Preference
// @Produce json
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(20)
// @Success 200 {object} model.InsightListResponse
// @Failure 500 {object} response.Response
// @Router /api/v1/steam/profile/insights [get]
func (h *PreferenceHandler) GetInsights(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	userID := middleware.GetUserID(c)

	result, err := h.preferenceService.GetInsights(c.Request.Context(), userID, page, pageSize)
	if err != nil {
		errorResp(c, http.StatusInternalServerError, "获取洞察列表失败: "+err.Error())
		return
	}

	success(c, result)
}
