// Package model 数据模型层 - 游戏推荐
package model

import (
	"time"

	"gorm.io/gorm"
)

// GameRecommendation 游戏推荐记录
type GameRecommendation struct {
	ID            int64          `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID        int64          `gorm:"index;not null" json:"user_id"`
	GameID        string         `gorm:"size:50;not null" json:"game_id"`                 // Steam AppID
	GameName      string         `gorm:"size:255;not null" json:"game_name"`             // 游戏名称
	GameGenre     string         `gorm:"size:255" json:"game_genre"`                   // 游戏类型
	GameTags      string         `gorm:"type:text" json:"game_tags"`                    // 游戏标签JSON
	CoverURL      string         `gorm:"size:512" json:"cover_url"`                    // 游戏封面URL
	StoreURL      string         `gorm:"size:512" json:"store_url"`                    // 商店页面URL
	MatchScore    float64        `gorm:"type:decimal(5,2);default:0" json:"match_score"` // 匹配度分数 0-100
	MatchReasons  string         `gorm:"type:text" json:"match_reasons"`               // 推荐理由(JSON数组)
	DealID        *int64         `gorm:"index" json:"deal_id,omitempty"`               // 关联促销ID
	DealPrice     float64        `gorm:"type:decimal(10,2)" json:"deal_price"`         // 促销价格
	DealDiscount  int            `gorm:"default:0" json:"deal_discount"`               // 折扣百分比
	DealEndDate   *time.Time     `json:"deal_end_date,omitempty"`                      // 促销截止日期
	Source        string         `gorm:"size:50;default:auto" json:"source"`           // 推荐来源
	Status        string         `gorm:"size:30;default:active" json:"status"`         // 状态
	CreatedAt     time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt     time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName 表名
func (GameRecommendation) TableName() string {
	return "game_recommendations"
}

// 推荐状态常量
const (
	RecStatusActive   = "active"
	RecStatusClicked  = "clicked"
	RecStatusPurchased = "purchased"
	RecStatusIgnored  = "ignored"
	RecStatusExpired  = "expired"
)

// 推荐来源常量
const (
	RecSourceAuto    = "auto"
	RecSourceManual  = "manual"
	RecSourceSurprise = "surprise"
)

// RecommendationStats 推荐统计
type RecommendationStats struct {
	ID              int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID          int64     `gorm:"uniqueIndex;not null" json:"user_id"`
	TotalRecs       int       `gorm:"default:0" json:"total_recommendations"`
	ClickedCount    int       `gorm:"default:0" json:"clicked_count"`
	PurchasedCount  int       `gorm:"default:0" json:"purchased_count"`
	IgnoredCount    int       `gorm:"default:0" json:"ignored_count"`
	CTR             float64   `gorm:"type:decimal(5,2);default:0" json:"ctr"`
	PurchaseRate    float64   `gorm:"type:decimal(5,2);default:0" json:"purchase_rate"`
	LastUpdated     time.Time `gorm:"autoUpdateTime" json:"last_updated"`
}

// TableName 表名
func (RecommendationStats) TableName() string {
	return "recommendation_stats"
}

// RecommendationResponse 推荐响应（前端展示用）
type RecommendationResponse struct {
	ID           int64    `json:"id"`
	GameID       string   `json:"game_id"`
	GameName     string   `json:"game_name"`
	GameGenre    string   `json:"game_genre"`
	GameTags     []string `json:"game_tags"`
	CoverURL     string   `json:"cover_url"`
	StoreURL     string   `json:"store_url"`
	MatchScore   float64  `json:"match_score"`  // 0-100
	MatchReasons []string `json:"match_reasons"`
	HasDeal      bool     `json:"has_deal"`
	DealPrice    float64  `json:"deal_price"`
	DealDiscount int      `json:"deal_discount"`
	DealEndDate  string   `json:"deal_end_date,omitempty"`
	Source       string   `json:"source"`
	Status       string   `json:"status"`
	CreatedAt    string   `json:"created_at"`
}

// RecommendationListResponse 推荐列表响应
type RecommendationListResponse struct {
	List       []RecommendationResponse `json:"list"`
	Total      int64                    `json:"total"`
	Page       int                      `json:"page"`
	PageSize   int                      `json:"page_size"`
	Stats      *RecStatsSummary         `json:"stats,omitempty"`
}

// RecStatsSummary 推荐统计摘要
type RecStatsSummary struct {
	TotalRecs      int     `json:"total_recommendations"`
	ClickedCount   int     `json:"clicked_count"`
	PurchaseCount  int     `json:"purchase_count"`
	CTR            float64 `json:"ctr"`     // 点击率
	PurchaseRate   float64 `json:"purchase_rate"` // 购买转化率
}

// FeedbackRequest 反馈请求
type FeedbackRequest struct {
	Action string `json:"action" binding:"required"` // like/dislike/click/ignore
}

// RecommendationGenerateRequest 生成推荐请求
type RecommendationGenerateRequest struct {
	MaxCount   int      `json:"max_count"`   // 最大推荐数量
	DealOnly   bool     `json:"deal_only"`   // 仅推荐促销游戏
	MinScore   float64  `json:"min_score"`   // 最低匹配度
	GameIDs    []string `json:"game_ids"`    // 指定游戏ID列表（为空则全量推荐）
}
