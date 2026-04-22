// Package model 数据模型层 - Steam账号绑定
package model

import (
	"time"

	"gorm.io/gorm"
)

// SteamAccount Steam账号绑定
type SteamAccount struct {
	ID            int64          `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID        int64          `gorm:"uniqueIndex;not null" json:"user_id"`
	SteamID       string         `gorm:"size:64;not null" json:"steam_id"`
	SteamNickname string         `gorm:"size:255" json:"steam_nickname"`
	AvatarURL     string         `gorm:"size:512" json:"avatar_url"`
	ProfileURL    string         `gorm:"size:512" json:"profile_url"`
	RealName      string         `gorm:"size:100" json:"real_name"`
	Location      string         `gorm:"size:50" json:"location"`
	APIKey        string         `gorm:"size:255" json:"-"` // 可选的私有API Key
	LastSyncAt    *time.Time     `json:"last_sync_at,omitempty"`
	IsActive      bool          `gorm:"default:true" json:"is_active"`
	CreatedAt     time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt     time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName 表名
func (SteamAccount) TableName() string {
	return "steam_accounts"
}

// SteamLibraryItem 用户Steam游戏库条目（从Steam API同步）
type SteamLibraryItem struct {
	ID           int64          `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID       int64          `gorm:"index;not null" json:"user_id"`
	AccountID    int64          `gorm:"index;not null" json:"account_id"` // 关联的SteamAccount ID
	GameID       string         `gorm:"size:50;not null" json:"game_id"`  // Steam AppID
	GameName     string         `gorm:"size:255;not null" json:"game_name"`
	Playtime     int            `gorm:"default:0" json:"playtime"`        // 总游玩时长(分钟)
	Playtime2Weeks int          `gorm:"default:0" json:"playtime_2_weeks"` // 最近两周(分钟)
	LastPlayedAt *time.Time     `json:"last_played_at,omitempty"`
	IconURL      string         `gorm:"size:512" json:"icon_url"`
	IsSynced     bool           `gorm:"default:false" json:"is_synced"` // 是否已同步元数据
	CreatedAt    time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt    time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName 表名
func (SteamLibraryItem) TableName() string {
	return "steam_library_items"
}
