// Package repository 数据访问层 - 推荐
package repository

import (
	"context"
	"encoding/json"
	"time"

	"email-backend/server/model"

	"gorm.io/gorm"
)

// RecommendationRepository 推荐数据访问
type RecommendationRepository struct {
	db *gorm.DB
}

// NewRecommendationRepository 创建推荐Repository
func NewRecommendationRepository(db *gorm.DB) *RecommendationRepository {
	return &RecommendationRepository{db: db}
}

// ==================== 推荐记录操作 ====================

// Create 创建推荐记录
func (r *RecommendationRepository) Create(ctx context.Context, rec *model.GameRecommendation) error {
	return r.db.WithContext(ctx).Create(rec).Error
}

// BatchCreate 批量创建推荐记录
func (r *RecommendationRepository) BatchCreate(ctx context.Context, recs []*model.GameRecommendation) error {
	if len(recs) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).CreateInBatches(recs, 50).Error
}

// Update 更新推荐记录
func (r *RecommendationRepository) Update(ctx context.Context, rec *model.GameRecommendation) error {
	return r.db.WithContext(ctx).Save(rec).Error
}

// FindByID 根据ID查询推荐
func (r *RecommendationRepository) FindByID(ctx context.Context, id int64) (*model.GameRecommendation, error) {
	var rec model.GameRecommendation
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&rec).Error
	if err != nil {
		return nil, err
	}
	return &rec, nil
}

// FindByUserGame 根据用户ID和游戏ID查询推荐
func (r *RecommendationRepository) FindByUserGame(ctx context.Context, userID int64, gameID string) (*model.GameRecommendation, error) {
	var rec model.GameRecommendation
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND game_id = ? AND status = ?", userID, gameID, model.RecStatusActive).
		Order("created_at DESC").
		First(&rec).Error
	if err != nil {
		return nil, err
	}
	return &rec, nil
}

// List 分页查询推荐列表
func (r *RecommendationRepository) List(ctx context.Context, userID int64, page, pageSize int, status string, dealOnly bool) ([]*model.GameRecommendation, int64, error) {
	var recs []*model.GameRecommendation
	var total int64

	query := r.db.WithContext(ctx).Model(&model.GameRecommendation{}).Where("user_id = ?", userID)

	if status != "" && status != "all" {
		query = query.Where("status = ?", status)
	} else {
		// 默认不显示已过期的
		query = query.Where("status != ?", model.RecStatusExpired)
	}

	if dealOnly {
		query = query.Where("deal_id IS NOT NULL")
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).
		Order("match_score DESC, created_at DESC").
		Find(&recs).Error; err != nil {
		return nil, 0, err
	}

	return recs, total, nil
}

// ListActiveDeals 获取活跃的促销推荐
func (r *RecommendationRepository) ListActiveDeals(ctx context.Context, userID int64, limit int) ([]*model.GameRecommendation, error) {
	var recs []*model.GameRecommendation
	now := time.Now()

	err := r.db.WithContext(ctx).
		Where("user_id = ? AND deal_id IS NOT NULL AND status = ? AND (deal_end_date IS NULL OR deal_end_date > ?)",
			userID, model.RecStatusActive, now).
		Order("match_score DESC, deal_discount DESC").
		Limit(limit).
		Find(&recs).Error
	return recs, err
}

// DeleteExpired 删除过期推荐
func (r *RecommendationRepository) DeleteExpired(ctx context.Context) error {
	now := time.Now()
	return r.db.WithContext(ctx).
		Model(&model.GameRecommendation{}).
		Where("status = ? AND deal_end_date IS NOT NULL AND deal_end_date < ?",
			model.RecStatusActive, now).
		Update("status", model.RecStatusExpired).Error
}

// DeleteByUser 删除用户的所有推荐（重新生成前）
func (r *RecommendationRepository) DeleteByUser(ctx context.Context, userID int64) error {
	return r.db.WithContext(ctx).
		Where("user_id = ? AND status = ?", userID, model.RecStatusActive).
		Delete(&model.GameRecommendation{}).Error
}

// UpdateStatus 更新推荐状态
func (r *RecommendationRepository) UpdateStatus(ctx context.Context, id int64, status string) error {
	return r.db.WithContext(ctx).
		Model(&model.GameRecommendation{}).
		Where("id = ?", id).
		Update("status", status).Error
}

// CountByStatus 统计各状态推荐数量
func (r *RecommendationRepository) CountByStatus(ctx context.Context, userID int64, status string) (int64, error) {
	var count int64
	query := r.db.WithContext(ctx).Model(&model.GameRecommendation{}).Where("user_id = ?", userID)
	if status != "" {
		query = query.Where("status = ?", status)
	}
	err := query.Count(&count).Error
	return count, err
}

// ==================== 统计操作 ====================

// GetOrCreateStats 获取或创建用户统计
func (r *RecommendationRepository) GetOrCreateStats(ctx context.Context, userID int64) (*model.RecommendationStats, error) {
	var stats model.RecommendationStats
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).First(&stats).Error
	if err == gorm.ErrRecordNotFound {
		stats = model.RecommendationStats{UserID: userID}
		if err := r.db.WithContext(ctx).Create(&stats).Error; err != nil {
			return nil, err
		}
		return &stats, nil
	}
	if err != nil {
		return nil, err
	}
	return &stats, nil
}

// UpdateStats 更新统计数据
func (r *RecommendationRepository) UpdateStats(ctx context.Context, userID int64, delta map[string]int) error {
	stats, err := r.GetOrCreateStats(ctx, userID)
	if err != nil {
		return err
	}

	if v, ok := delta["total"]; ok {
		stats.TotalRecs += v
	}
	if v, ok := delta["clicked"]; ok {
		stats.ClickedCount += v
	}
	if v, ok := delta["purchased"]; ok {
		stats.PurchasedCount += v
	}
	if v, ok := delta["ignored"]; ok {
		stats.IgnoredCount += v
	}

	// 计算比率
	if stats.TotalRecs > 0 {
		stats.CTR = float64(stats.ClickedCount) / float64(stats.TotalRecs) * 100
		stats.PurchaseRate = float64(stats.PurchasedCount) / float64(stats.TotalRecs) * 100
	}

	return r.db.WithContext(ctx).Save(stats).Error
}

// ==================== 推荐算法辅助查询 ====================

// GetOwnedGameIDs 获取用户已拥有的游戏ID集合
func (r *RecommendationRepository) GetOwnedGameIDs(ctx context.Context, userID int64) (map[string]bool, error) {
	var games []model.SteamGame
	err := r.db.WithContext(ctx).
		Select("game_id").
		Where("user_id = ? AND is_owned = ?", userID, true).
		Find(&games).Error
	if err != nil {
		return nil, err
	}

	result := make(map[string]bool, len(games))
	for _, g := range games {
		if g.GameID != "" {
			result[g.GameID] = true
		}
	}
	return result, nil
}

// GetUserPreferences 获取用户偏好标签及权重
func (r *RecommendationRepository) GetUserPreferences(ctx context.Context, userID int64) ([]model.UserGamePreference, error) {
	var prefs []model.UserGamePreference
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("weight DESC").
		Find(&prefs).Error
	return prefs, err
}

// GetActiveDealsWithoutRecommendation 获取未生成推荐的活跃促销
func (r *RecommendationRepository) GetActiveDealsWithoutRecommendation(ctx context.Context, userID int64, limit int) ([]model.SteamDeal, error) {
	var deals []model.SteamDeal

	// 子查询获取已有推荐的game_id
	subQuery := r.db.WithContext(ctx).
		Model(&model.GameRecommendation{}).
		Select("game_id").
		Where("user_id = ? AND status = ?", userID, model.RecStatusActive)

	err := r.db.WithContext(ctx).
		Model(&model.SteamDeal{}).
		Where("user_id = ? AND is_active = ? AND game_id NOT IN (?)",
			userID, true, subQuery).
		Order("discount DESC").
		Limit(limit).
		Find(&deals).Error
	return deals, err
}

// ==================== 响应转换 ====================

// ToResponse 将推荐记录转换为响应格式
func (r *RecommendationRepository) ToResponse(rec *model.GameRecommendation) model.RecommendationResponse {
	resp := model.RecommendationResponse{
		ID:           rec.ID,
		GameID:       rec.GameID,
		GameName:     rec.GameName,
		GameGenre:    rec.GameGenre,
		CoverURL:     rec.CoverURL,
		StoreURL:     rec.StoreURL,
		MatchScore:   rec.MatchScore,
		HasDeal:      rec.DealID != nil,
		DealPrice:    rec.DealPrice,
		DealDiscount: rec.DealDiscount,
		Source:       rec.Source,
		Status:       rec.Status,
		CreatedAt:    rec.CreatedAt.Format("2006-01-02 15:04:05"),
	}

	// 解析标签
	if rec.GameTags != "" {
		json.Unmarshal([]byte(rec.GameTags), &resp.GameTags)
	}

	// 解析推荐理由
	if rec.MatchReasons != "" {
		json.Unmarshal([]byte(rec.MatchReasons), &resp.MatchReasons)
	}

	// 促销截止日期
	if rec.DealEndDate != nil {
		resp.DealEndDate = rec.DealEndDate.Format("2006-01-02")
	}

	return resp
}
