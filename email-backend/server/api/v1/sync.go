// Package v1 同步接口
package v1

import (
	"net/http"

	"email-backend/server/global"
	"email-backend/server/middleware"
	"email-backend/server/pkg/agent"
	"email-backend/server/repository"
	"email-backend/server/service"

	"github.com/gin-gonic/gin"
)

// SyncHandler 同步处理器
type SyncHandler struct {
	syncService *service.SyncService
	scheduler   *service.SyncScheduler
}

// NewSyncHandler 创建同步处理器
func NewSyncHandler(syncSvc *service.SyncService, scheduler *service.SyncScheduler) *SyncHandler {
	return &SyncHandler{syncService: syncSvc, scheduler: scheduler}
}

// SetupSyncRoutes 注册同步路由
func SetupSyncRoutes(r *gin.RouterGroup, agentClient *agent.Client, scheduler *service.SyncScheduler) {
	accountRepo := repository.NewAccountRepository(global.DB())
	emailRepo := repository.NewEmailRepository(global.DB())
	syncSvc := service.NewSyncService(accountRepo, emailRepo, agentClient)
	h := NewSyncHandler(syncSvc, scheduler)

	sync := r.Group("/sync")
	{
		sync.POST("", h.TriggerSync)
		sync.GET("/status", h.SyncStatus)

		// 调度器控制
		sync.POST("/scheduler/start", h.StartScheduler)
		sync.POST("/scheduler/stop", h.StopScheduler)
		sync.GET("/scheduler/status", h.SchedulerStatus)
		sync.PUT("/scheduler/interval", h.SetInterval)
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
	userID := middleware.GetUserID(c)

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
	userID := middleware.GetUserID(c)

	status, err := h.syncService.GetSyncStatus(c.Request.Context(), userID)
	if err != nil {
		errorResp(c, http.StatusInternalServerError, "获取状态失败")
		return
	}

	success(c, status)
}

// StartScheduler 启动调度器
func (h *SyncHandler) StartScheduler(c *gin.Context) {
	if h.scheduler == nil {
		errorResp(c, http.StatusBadRequest, "调度器未初始化")
		return
	}

	if err := h.scheduler.Start(); err != nil {
		errorResp(c, http.StatusInternalServerError, "启动调度器失败: "+err.Error())
		return
	}

	success(c, gin.H{"message": "调度器已启动"})
}

// StopScheduler 停止调度器
func (h *SyncHandler) StopScheduler(c *gin.Context) {
	if h.scheduler == nil {
		errorResp(c, http.StatusBadRequest, "调度器未初始化")
		return
	}

	h.scheduler.Stop()
	success(c, gin.H{"message": "调度器已停止"})
}

// SchedulerStatus 获取调度器状态
func (h *SyncHandler) SchedulerStatus(c *gin.Context) {
	if h.scheduler == nil {
		success(c, gin.H{"running": false})
		return
	}

	success(c, h.scheduler.GetStatus())
}

// SetIntervalRequest 设置间隔请求
type SetIntervalRequest struct {
	Interval int `json:"interval" binding:"required,min=1,max=1440"` // 分钟，1分钟-24小时
}

// SetInterval 设置同步间隔
func (h *SyncHandler) SetInterval(c *gin.Context) {
	var req SetIntervalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errorResp(c, http.StatusBadRequest, "参数错误: "+err.Error())
		return
	}

	if h.scheduler == nil {
		errorResp(c, http.StatusBadRequest, "调度器未初始化")
		return
	}

	h.scheduler.SetInterval(req.Interval)
	success(c, gin.H{
		"message":  "同步间隔已更新",
		"interval": req.Interval,
	})
}
