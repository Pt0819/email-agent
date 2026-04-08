// Package model 响应结构体
package model

// Response 统一响应结构
type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	TraceID string      `json:"trace_id,omitempty"`
}

// PageData 分页数据
type PageData struct {
	List       interface{} `json:"list"`
	Total      int64       `json:"total"`
	Page       int         `json:"page"`
	PageSize   int         `json:"page_size"`
	TotalPages int         `json:"total_pages"`
}

// EmailListResponse 邮件列表响应
type EmailListResponse struct {
	List      interface{} `json:"list"`
	Total     int64       `json:"total"`
	Page      int         `json:"page"`
	PageSize  int         `json:"page_size"`
}

// ClassificationResponse 分类响应
type ClassificationResponse struct {
	EmailID    string  `json:"email_id"`
	Category   string  `json:"category"`
	Priority   string  `json:"priority"`
	Confidence float64 `json:"confidence"`
	Reasoning  string  `json:"reasoning"`
}

// AccountResponse 账户响应
type AccountResponse struct {
	ID           int64   `json:"id"`
	Email        string  `json:"email"`
	Provider     string  `json:"provider"`
	LastSyncAt   string  `json:"last_sync_at,omitempty"`
	SyncEnabled  bool    `json:"sync_enabled"`
}

// SyncResponse 同步响应
type SyncResponse struct {
	TaskID   string `json:"task_id"`
	Status   string `json:"status"`
	Progress int    `json:"progress,omitempty"`
	Total    int    `json:"total,omitempty"`
}

// ErrorCode 错误码定义
const (
	CodeSuccess         = 0
	CodeBadRequest      = 400
	CodeUnauthorized    = 401
	CodeForbidden       = 403
	CodeNotFound        = 404
	CodeInternalError   = 500
)