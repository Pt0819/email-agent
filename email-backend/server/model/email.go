// Package model 数据模型层
package model

import (
	"time"

	"gorm.io/gorm"
)

// Email 邮件模型
type Email struct {
	ID           int64          `gorm:"primaryKey;autoIncrement" json:"id"`
	MessageID    string         `gorm:"uniqueIndex;size:255;not null" json:"message_id"`
	UserID       int64          `gorm:"index;not null" json:"user_id"`
	AccountID    int64          `gorm:"index;not null" json:"account_id"`
	SenderName   string         `gorm:"size:255" json:"sender_name"`
	SenderEmail  string         `gorm:"size:255;not null" json:"sender_email"`
	Subject      string         `gorm:"size:512" json:"subject"`
	Content      string         `gorm:"type:text" json:"content"`
	ContentHTML  string         `gorm:"type:text" json:"content_html"`
	Category     string         `gorm:"size:50;default:unclassified;index" json:"category"`
	Priority     string         `gorm:"size:20;default:medium" json:"priority"`
	Confidence   float64       `gorm:"type:decimal(5,4);default:0" json:"confidence"`
	Status       string         `gorm:"size:20;default:unread;index" json:"status"`
	IsProcessed  bool           `gorm:"default:false" json:"is_processed"`
	HasAttachment bool          `gorm:"default:false" json:"has_attachment"`
	ReceivedAt   time.Time     `gorm:"not null;index" json:"received_at"`
	ProcessedAt   *time.Time    `json:"processed_at,omitempty"`
	CreatedAt     time.Time     `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt     time.Time     `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName 表名
func (Email) TableName() string {
	return "emails"
}

// EmailAccount 邮箱账户模型
type EmailAccount struct {
	ID                  int64          `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID              int64          `gorm:"index;not null" json:"user_id"`
	Provider             string         `gorm:"size:20;not null" json:"provider"`
	AccountEmail        string         `gorm:"size:255;not null" json:"account_email"`
	EncryptedCredential string         `gorm:"type:text;not null" json:"-"`
	CredentialIV        string         `gorm:"size:64" json:"-"`
	LastSyncAt          *time.Time    `json:"last_sync_at,omitempty"`
	SyncEnabled         bool           `gorm:"default:true" json:"sync_enabled"`
	CreatedAt           time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt           time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt           gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName 表名
func (EmailAccount) TableName() string {
	return "email_accounts"
}

// ActionItem 行动项模型
type ActionItem struct {
	ID          int64          `gorm:"primaryKey;autoIncrement" json:"id"`
	EmailID     int64          `gorm:"index;not null" json:"email_id"`
	UserID      int64          `gorm:"index;not null" json:"user_id"`
	Task        string         `gorm:"type:text;not null" json:"task"`
	TaskType    string         `gorm:"size:50" json:"task_type"`
	Deadline    *time.Time    `json:"deadline,omitempty"`
	Priority    string         `gorm:"size:20;default:medium" json:"priority"`
	Status      string         `gorm:"size:20;default:pending" json:"status"`
	CreatedAt   time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName 表名
func (ActionItem) TableName() string {
	return "action_items"
}