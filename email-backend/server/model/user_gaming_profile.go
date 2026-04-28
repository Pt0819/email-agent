// Package model 数据模型层 - 用户游戏画像
package model

import "time"

// UserGamingProfile 用户游戏画像（聚合视图，非数据库表）
// 用于前端展示用户的整体游戏偏好画像
type UserGamingProfile struct {
	UserID         int64                  `json:"user_id"`
	TopTags        []TagPreference        `json:"top_tags"`          // Top偏好标签
	TopGenres      []TagPreference        `json:"top_genres"`        // Top游戏类型
	TopDevelopers  []TagPreference        `json:"top_developers"`    // 常玩开发商
	TotalGames     int                    `json:"total_games"`       // 拥有游戏总数
	TotalPlaytime  int                    `json:"total_playtime"`    // 总游玩时长(分钟)
	RecentActivity *RecentActivitySummary `json:"recent_activity"`   // 近期活动摘要
	LastAnalyzedAt *time.Time             `json:"last_analyzed_at"`  // 上次分析时间
}

// TagPreference 标签偏好
type TagPreference struct {
	Tag     string  `json:"tag"`      // 标签名称
	Weight  float64 `json:"weight"`   // 权重 0-10
	Count   int     `json:"count"`   // 出现次数
	Source  string  `json:"source"`  // 来源
}

// RecentActivitySummary 近期活动摘要
type RecentActivitySummary struct {
	GamesPlayedLastWeek   int     `json:"games_played_last_week"`   // 上周游玩游戏数
	TotalPlaytimeLastWeek int     `json:"total_playtime_last_week"`  // 上周总时长
	MostPlayedGame        string  `json:"most_played_game"`          // 最常玩游戏
	MostPlayedGameHours   int     `json:"most_played_game_hours"`    // 最常玩游戏时长
	NewGamesAdded         int     `json:"new_games_added"`           // 新增游戏数
	GenreDistribution     []GenreCount `json:"genre_distribution"`  // 类型分布
}

// GenreCount 类型计数
type GenreCount struct {
	Genre string `json:"genre"`
	Count int    `json:"count"`
}

// PreferenceAnalysisResult 偏好分析结果
type PreferenceAnalysisResult struct {
	Success         bool          `json:"success"`
	Profile         *UserGamingProfile `json:"profile"`
	NewTags         []TagPreference   `json:"new_tags"`       // 新发现的标签
	UpdatedTags     []TagPreference   `json:"updated_tags"`   // 更新的标签
	Insights        []string          `json:"insights"`       // Agent洞察
	Error           string            `json:"error,omitempty"`
}
