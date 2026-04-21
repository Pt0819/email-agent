// Package model 数据模型层 - Steam游戏
package model

import (
	"time"

	"gorm.io/gorm"
)

// SteamGame Steam游戏模型
type SteamGame struct {
	ID          int64          `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID      int64          `gorm:"index;not null" json:"user_id"`
	GameName    string         `gorm:"size:255;not null" json:"game_name"`
	GameID      string         `gorm:"size:50;index" json:"game_id"`           // Steam AppID
	Developer   string         `gorm:"size:255" json:"developer"`
	Publisher   string         `gorm:"size:255" json:"publisher"`
	Genre       string         `gorm:"size:255" json:"genre"`                  // 类型：动作、RPG等
	Tags        string         `gorm:"type:text" json:"tags"`                  // 标签JSON数组
	CoverURL    string         `gorm:"size:512" json:"cover_url"`              // 封面图URL
	StoreURL    string         `gorm:"size:512" json:"store_url"`              // 商店页URL
	Playtime    int            `gorm:"default:0" json:"playtime"`              // 游玩时长(分钟)
	IsOwned     bool           `gorm:"default:false" json:"is_owned"`          // 是否已拥有
	CreatedAt   time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName 表名
func (SteamGame) TableName() string {
	return "steam_games"
}

// SteamDeal Steam促销模型
type SteamDeal struct {
	ID           int64          `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID       int64          `gorm:"index;not null" json:"user_id"`
	GameID       string         `gorm:"size:50;index" json:"game_id"`
	GameName     string         `gorm:"size:255;not null" json:"game_name"`
	OriginalPrice float64       `gorm:"type:decimal(10,2);default:0" json:"original_price"`
	DealPrice    float64        `gorm:"type:decimal(10,2);default:0" json:"deal_price"`
	Discount     int            `gorm:"default:0" json:"discount"`               // 折扣百分比 0-100
	CoverURL     string         `gorm:"size:512" json:"cover_url"`
	StoreURL     string         `gorm:"size:512" json:"store_url"`
	StartDate    *time.Time     `json:"start_date,omitempty"`
	EndDate      *time.Time     `gorm:"index" json:"end_date,omitempty"`
	IsActive     bool           `gorm:"default:true;index" json:"is_active"`
	EmailID      int64          `gorm:"index" json:"email_id"`                   // 来源邮件ID
	CreatedAt    time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt    time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName 表名
func (SteamDeal) TableName() string {
	return "steam_deals"
}
