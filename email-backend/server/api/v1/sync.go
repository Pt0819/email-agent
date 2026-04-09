// Package v1 同步接口
package v1

import (
	"net/http"

	"email-backend/server/global"
	"email-backend/server/repository"
	"email-backend/server/service"

	"github.com/gin-gonic/gin"
)

// SyncHandler 同步处理器
type SyncHandler struct {
	syncService *service.SyncService
}

// NewSyncHandler 创建同步处理器
func NewSyncHandler(syncSvc *service.SyncService) *SyncHandler {
	return &SyncHandler{syncService: syncSvc}
}

// SetupSyncRoutes 注册同步路由
func SetupSyncRoutes(r *gin.RouterGroup) {
	accountRepo := repository.NewAccountRepository(global.DB())
	emailRepo := repository.NewEmailRepository(global.DB())
	syncSvc := service.NewSyncService(accountRepo, emailRepo)
	h := NewSyncHandler(syncSvc)

	sync := r.Group("/sync")
	{
		sync.POST("", h.TriggerSync)
		sync.GET("/status", h.SyncStatus)
	}
}

// SyncRequest 同步请求
type SyncRequest struct {
	AccountID int64 `json:"account_id,omitempty"` // 可选，不传则同步所有账户
}

// TriggerSync 触发同步
func (h *SyncHandler) TriggerSync(c *gin.Context) {
	var req SyncRequest
	c.ShouldBindJSON(&req)

	// TODO: 从JWT获取用户ID
	userID := int64(1)

	var results []*service.SyncResult
	var err error

	if req.AccountID > 0 {
		// 同步单个账户
		result := h.syncService.SyncAccount(c.Request.Context(), req.AccountID)
		results = []*service.SyncResult{result}
	} else {
		// 同步所有账户
		results, err = h.syncService.SyncAll(c.Request.Context(), userID)
		if err != nil {
			errorResp(c, http.StatusInternalServerError, "同步失败: "+err.Error())
			return
		}
	}

	// 检查结果
	allSuccess := true
	for _, r := range results {
		if !r.Success {
			allSuccess = false
			break
		}
	}

	success(c, gin.H{
		"status":      "completed",
		"all_success": allSuccess,
		"results":     results,
	})
}

// SyncStatus 获取同步状态
func (h *SyncHandler) SyncStatus(c *gin.Context) {
	// TODO: 从JWT获取用户ID
	userID := int64(1)

	status, err := h.syncService.GetSyncStatus(c.Request.Context(), userID)
	if err != nil {
		errorResp(c, http.StatusInternalServerError, "获取状态失败")
		return
	}

	success(c, status)
}