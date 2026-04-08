// Package v1 同步接口
package v1

import (
	respModel "email-backend/server/model/response"

	"github.com/gin-gonic/gin"
)

// SyncHandler 同步处理器
type SyncHandler struct{}

// NewSyncHandler 创建同步处理器
func NewSyncHandler() *SyncHandler {
	return &SyncHandler{}
}

// SetupSyncRoutes 注册同步路由
func SetupSyncRoutes(r *gin.RouterGroup) {
	h := NewSyncHandler()

	sync := r.Group("/sync")
	{
		sync.POST("", h.TriggerSync)
		sync.GET("/status", h.SyncStatus)
	}
}

// TriggerSync 触发同步
func (h *SyncHandler) TriggerSync(c *gin.Context) {
	// TODO: 实际触发邮件同步
	success(c, &respModel.SyncResponse{
		TaskID: "sync_001",
		Status: "started",
	})
}

// SyncStatus 获取同步状态
func (h *SyncHandler) SyncStatus(c *gin.Context) {
	// TODO: 实际获取同步状态
	success(c, gin.H{
		"last_sync": "2026-04-08T10:00:00Z",
		"status":    "idle",
	})
}