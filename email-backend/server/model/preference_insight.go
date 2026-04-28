// Package model 数据模型层 - 偏好分析洞察记录
package model

import (
	"time"

	"gorm.io/gorm"
)

// PreferenceInsight 偏好分析洞察记录
// Agent在分析用户偏好时生成的洞察和决策记录
type PreferenceInsight struct {
	ID           int64          `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID       int64          `gorm:"index;not null" json:"user_id"`
	EventType    string         `gorm:"size:50;not null" json:"event_type"`    // 触发事件类型
	DecisionType string         `gorm:"size:50;not null" json:"decision_type"`  // 决策类型
	TriggerDesc  string         `gorm:"size:255" json:"trigger_desc"`          // 触发描述
	Insight      string         `gorm:"type:text" json:"insight"`               // 洞察内容
	Reasoning    string         `gorm:"type:text" json:"reasoning"`             // 决策理由
	Actions      string         `gorm:"type:text" json:"actions"`              // 执行的操作(JSON数组)
	Confidence   float64        `gorm:"type:decimal(4,2);default:0" json:"confidence"` // 置信度
	IsAnomaly    bool           `gorm:"default:false" json:"is_anomaly"`         // 是否异常标记
	AnomalyType  string         `gorm:"size:50" json:"anomaly_type,omitempty"` // 异常类型
	GameID       string         `gorm:"size:50" json:"game_id,omitempty"`      // 关联游戏
	GameName     string         `gorm:"size:255" json:"game_name,omitempty"`   // 游戏名称
	TagsChanged  string         `gorm:"type:text" json:"tags_changed"`         // 标签变化(JSON)
	CreatedAt    time.Time      `gorm:"autoCreateTime" json:"created_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName 表名
func (PreferenceInsight) TableName() string {
	return "preference_insights"
}

// InsightResponse 洞察响应（前端展示用）
type InsightResponse struct {
	ID           int64   `json:"id"`
	EventType    string  `json:"event_type"`
	DecisionType string  `json:"decision_type"`
	TriggerDesc  string  `json:"trigger_desc"`
	Insight      string  `json:"insight"`
	Reasoning    string  `json:"reasoning"`
	Confidence   float64 `json:"confidence"`
	IsAnomaly    bool    `json:"is_anomaly"`
	AnomalyType  string  `json:"anomaly_type,omitempty"`
	GameID       string  `json:"game_id,omitempty"`
	GameName     string  `json:"game_name,omitempty"`
	TagsChanged  []TagChange `json:"tags_changed"`
	CreatedAt    string  `json:"created_at"`
}

// TagChange 标签变化
type TagChange struct {
	Tag   string  `json:"tag"`
	Old   float64 `json:"old"`
	New   float64 `json:"new"`
	Delta float64 `json:"delta"`
}

// InsightListResponse 洞察列表响应
type InsightListResponse struct {
	List  []InsightResponse `json:"list"`
	Total int64             `json:"total"`
	Page  int               `json:"page"`
}

// InsightEventType 洞察事件类型（与Agent TriggerType对应）
const (
	InsightEventSteamEmailSync = "steam_email_sync"
	InsightEventLibrarySync    = "library_sync"
	InsightEventPlaytimeUpdate = "playtime_update"
	InsightEventNewGameAdded   = "new_game_added"
	InsightEventUserFeedback   = "user_feedback"
	InsightEventGamePurchased  = "game_purchased"
	InsightEventGameWishlisted = "game_wishlisted"
	InsightEventPeriodicCheck  = "periodic_check"
	InsightEventManualTrigger  = "manual_trigger"
)

// InsightDecisionType 洞察决策类型
const (
	InsightDecisionNoAction          = "no_action"
	InsightDecisionProfileUpdate     = "profile_update"
	InsightDecisionTagWeightAdjust   = "tag_weight_adjust"
	InsightDecisionAnomalyDetected  = "anomaly_detected"
	InsightDecisionPreferenceDrift   = "preference_drift"
	InsightDecisionNewPattern        = "new_pattern"
	InsightDecisionPushNotification  = "push_notification"
	InsightDecisionGenerateRec       = "generate_recommendation"
	InsightDecisionRequestConfirm    = "request_confirm"
)

// InsightAnomalyType 异常类型
const (
	AnomalyTypeExtremePlaytime  = "extreme_playtime"
	AnomalyTypeNewGenreExplored = "new_genre_explored"
	AnomalyTypePreferenceDrift  = "preference_drift"
	AnomalyTypeSuddenDrop       = "sudden_drop"
)
