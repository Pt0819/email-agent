// Package model 数据模型层 - 用户游戏偏好
package model

import (
	"time"

	"gorm.io/gorm"
)

// UserGamePreference 用户游戏偏好模型
type UserGamePreference struct {
	ID        int64          `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID    int64          `gorm:"index;not null" json:"user_id"`
	Tag       string         `gorm:"size:100;not null" json:"tag"`
	Weight    float64        `gorm:"type:decimal(5,2);default:1.00" json:"weight"`
	Source    string         `gorm:"size:30;default:system" json:"source"`
	CreatedAt time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName 表名
func (UserGamePreference) TableName() string {
	return "user_game_preferences"
}

// RecommendationFeedback 推荐反馈模型
type RecommendationFeedback struct {
	ID        int64          `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID    int64          `gorm:"index;not null" json:"user_id"`
	GameID    string         `gorm:"size:50;not null" json:"game_id"`
	GameName  string         `gorm:"size:255;not null" json:"game_name"`
	Action    string         `gorm:"size:30;not null" json:"action"`
	DealID    *int64         `gorm:"index" json:"deal_id,omitempty"`
	EmailID   *int64         `gorm:"index" json:"email_id,omitempty"`
	CreatedAt time.Time      `gorm:"autoCreateTime" json:"created_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName 表名
func (RecommendationFeedback) TableName() string {
	return "recommendation_feedback"
}

// 反馈操作类型
const (
	FeedbackActionClicked    = "clicked"
	FeedbackActionPurchased  = "purchased"
	FeedbackActionIgnored    = "ignored"
	FeedbackActionWishlisted = "wishlisted"
)

// 偏好来源类型
const (
	PreferenceSourceWishlist = "wishlist"
	PreferenceSourceEmail    = "email_purchase"
	PreferenceSourceManual  = "manual"
	PreferenceSourceSystem   = "system"
)