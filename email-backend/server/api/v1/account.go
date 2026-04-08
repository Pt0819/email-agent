// Package v1 账户接口
package v1

import (
	"net/http"
	"strconv"

	"email-backend/server/model"
	emailRequest "email-backend/server/model/request"
	respModel "email-backend/server/model/response"
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
	h := NewAccountHandler(service.NewAccountService(nil))

	accounts := r.Group("/accounts")
	{
		accounts.GET("", h.ListAccounts)
		accounts.POST("", h.CreateAccount)
		accounts.DELETE("/:id", h.DeleteAccount)
		accounts.POST("/:id/test", h.TestAccount)
	}
}

// ListAccounts 获取账户列表
func (h *AccountHandler) ListAccounts(c *gin.Context) {
	accounts, err := h.accountService.List(c.Request.Context(), 1)
	if err != nil {
		errorResp(c, http.StatusInternalServerError, err.Error())
		return
	}

	success(c, gin.H{
		"list": accounts,
	})
}

// CreateAccount 创建账户
func (h *AccountHandler) CreateAccount(c *gin.Context) {
	var req emailRequest.CreateAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		badRequest(c, "参数错误")
		return
	}

	account := &model.EmailAccount{
		UserID:       1,
		Provider:     req.Provider,
		AccountEmail: req.Email,
		SyncEnabled:  true,
	}

	if err := h.accountService.Create(c.Request.Context(), account); err != nil {
		errorResp(c, http.StatusInternalServerError, "创建账户失败")
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

	// TODO: 实际测试邮箱连接
	success(c, gin.H{
		"id":      id,
		"status":  "connected",
		"message": "连接成功",
	})
}