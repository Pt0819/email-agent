// Package v1 账户接口
package v1

import (
	"net/http"
	"strconv"

	"email-backend/server/global"
	"email-backend/server/model"
	emailRequest "email-backend/server/model/request"
	respModel "email-backend/server/model/response"
	"email-backend/server/repository"
	"email-backend/server/service"

	"github.com/gin-gonic/gin"
)

// AccountHandler 账户处理器
type AccountHandler struct {
	accountService *service.AccountService
}

// NewAccountHandler 创建账户处理器
func NewAccountHandler(accountSvc *service.AccountService) *AccountHandler {
	return &AccountHandler{accountService: accountSvc}
}

// SetupAccountRoutes 注册账户路由
func SetupAccountRoutes(r *gin.RouterGroup) {
	accountRepo := repository.NewAccountRepository(global.DB())
	accountSvc := service.NewAccountService(accountRepo)
	h := NewAccountHandler(accountSvc)

	accounts := r.Group("/accounts")
	{
		accounts.GET("", h.ListAccounts)
		accounts.POST("", h.CreateAccount)
		accounts.GET("/:id", h.GetAccount)
		accounts.DELETE("/:id", h.DeleteAccount)
		accounts.POST("/:id/test", h.TestAccount)
	}
}

// ListAccounts 获取账户列表
func (h *AccountHandler) ListAccounts(c *gin.Context) {
	// TODO: 从JWT获取用户ID
	userID := int64(1)

	accounts, err := h.accountService.List(c.Request.Context(), userID)
	if err != nil {
		errorResp(c, http.StatusInternalServerError, err.Error())
		return
	}

	// 转换为响应格式
	list := make([]*respModel.AccountResponse, 0, len(accounts))
	for _, acc := range accounts {
		list = append(list, &respModel.AccountResponse{
			ID:          acc.ID,
			Email:       acc.AccountEmail,
			Provider:    acc.Provider,
			SyncEnabled: acc.SyncEnabled,
			LastSyncAt:  acc.LastSyncAt,
		})
	}

	success(c, gin.H{
		"list": list,
	})
}

// GetAccount 获取单个账户
func (h *AccountHandler) GetAccount(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		badRequest(c, "无效的账户ID")
		return
	}

	account, err := h.accountService.GetByID(c.Request.Context(), id)
	if err != nil {
		notFound(c, "账户不存在")
		return
	}

	success(c, &respModel.AccountResponse{
		ID:          account.ID,
		Email:       account.AccountEmail,
		Provider:    account.Provider,
		SyncEnabled: account.SyncEnabled,
		LastSyncAt:  account.LastSyncAt,
	})
}

// CreateAccount 创建账户
func (h *AccountHandler) CreateAccount(c *gin.Context) {
	var req emailRequest.CreateAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		badRequest(c, "参数错误: "+err.Error())
		return
	}

	// 验证参数
	if req.Email == "" {
		badRequest(c, "邮箱地址不能为空")
		return
	}
	if req.Credential == "" {
		badRequest(c, "授权码不能为空")
		return
	}
	if req.Provider == "" {
		badRequest(c, "邮箱类型不能为空")
		return
	}

	// TODO: 从JWT获取用户ID
	userID := int64(1)

	account := &model.EmailAccount{
		UserID:       userID,
		Provider:     req.Provider,
		AccountEmail: req.Email,
		SyncEnabled:  true,
	}

	// 创建账户并加密凭证
	err := h.accountService.Create(c.Request.Context(), account, req.Credential)
	if err != nil {
		errorResp(c, http.StatusInternalServerError, "创建账户失败: "+err.Error())
		return
	}

	created(c, &respModel.AccountResponse{
		ID:          account.ID,
		Email:       account.AccountEmail,
		Provider:    account.Provider,
		SyncEnabled: account.SyncEnabled,
	})
}

// DeleteAccount 删除账户
func (h *AccountHandler) DeleteAccount(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		badRequest(c, "无效的账户ID")
		return
	}

	if err := h.accountService.Delete(c.Request.Context(), id); err != nil {
		errorResp(c, http.StatusInternalServerError, "删除失败")
		return
	}

	success(c, nil)
}

// TestAccount 测试账户连接
func (h *AccountHandler) TestAccount(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		badRequest(c, "无效的账户ID")
		return
	}

	result, err := h.accountService.TestConnection(c.Request.Context(), id)
	if err != nil {
		errorResp(c, http.StatusInternalServerError, "测试连接失败: "+err.Error())
		return
	}

	success(c, gin.H{
		"id":      id,
		"success": result.Success,
		"message": result.Message,
	})
}
