// Package model 请求结构体
package model

// ListRequest 列表请求
type ListRequest struct {
	Page      int    `form:"page" json:"page"`
	PageSize  int    `form:"page_size" json:"page_size"`
	UserID    int64  `form:"-" json:"-"`
	AccountID int64  `form:"account_id" json:"account_id"`
	Category  string `form:"category" json:"category"`
	Status    string `form:"status" json:"status"`
	Keyword   string `form:"keyword" json:"keyword"`
}

// CreateAccountRequest 创建账户请求
type CreateAccountRequest struct {
	Email      string `json:"email" binding:"required,email"`
	Provider   string `json:"provider" binding:"required"`
	Credential string `json:"credential" binding:"required"`
}

// ClassificationRequest 分类请求
type ClassificationRequest struct {
	EmailID    string `json:"email_id" binding:"required"`
	Subject    string `json:"subject"`
	SenderName string `json:"sender_name"`
	Sender     string `json:"sender" binding:"required"`
	Content    string `json:"content" binding:"required"`
}

// SyncRequest 同步请求
type SyncRequest struct {
	AccountID int64 `json:"account_id,omitempty"`
}

// UpdateAccountRequest 更新账户请求
type UpdateAccountRequest struct {
	SyncEnabled *bool   `json:"sync_enabled,omitempty"`
	Credential  *string `json:"credential,omitempty"`
}