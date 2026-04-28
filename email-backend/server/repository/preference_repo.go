// Package repository 数据访问层 - 用户偏好
package repository

import (
	"context"

	"email-backend/server/model"

	"gorm.io/gorm"
)

// PreferenceRepository 用户偏好仓库
type PreferenceRepository struct {
	db *gorm.DB
}

// NewPreferenceRepository 创建用户偏好仓库
func NewPreferenceRepository(db *gorm.DB) *PreferenceRepository {
	return &PreferenceRepository{db: db}
}

// UpsertPreference 创建或更新偏好
func (r *PreferenceRepository) UpsertPreference(ctx context.Context, pref *model.UserGamePreference) error {
	return r.db.WithContext(ctx).
		Assign(map[string]interface{}{
			"weight": gorm.Expr("weight + ?", pref.Weight),
		}).
		FirstOrCreate(pref, map[string]interface{}{
			"user_id": pref.UserID,
			"tag":     pref.Tag,
		}).Error
}

// GetPreferencesByUserID 获取用户偏好列表
func (r *PreferenceRepository) GetPreferencesByUserID(ctx context.Context, userID int64, limit int) ([]*model.UserGamePreference, error) {
	var prefs []*model.UserGamePreference
	query := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("weight DESC")
	if limit > 0 {
		query = query.Limit(limit)
	}
	err := query.Find(&prefs).Error
	return prefs, err
}

// BatchUpsertPreferences 批量创建或更新偏好
func (r *PreferenceRepository) BatchUpsertPreferences(ctx context.Context, prefs []*model.UserGamePreference) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, pref := range prefs {
			if err := tx.Assign(map[string]interface{}{
				"weight": gorm.Expr("weight + ?", pref.Weight),
			}).FirstOrCreate(pref, map[string]interface{}{
				"user_id": pref.UserID,
				"tag":     pref.Tag,
			}).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// CreateFeedback 创建推荐反馈
func (r *PreferenceRepository) CreateFeedback(ctx context.Context, feedback *model.RecommendationFeedback) error {
	return r.db.WithContext(ctx).Create(feedback).Error
}

// GetUserFeedbackHistory 获取用户反馈历史
func (r *PreferenceRepository) GetUserFeedbackHistory(ctx context.Context, userID int64, limit int) ([]*model.RecommendationFeedback, error) {
	var feedbacks []*model.RecommendationFeedback
	query := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC")
	if limit > 0 {
		query = query.Limit(limit)
	}
	err := query.Find(&feedbacks).Error
	return feedbacks, err
}

// CountFeedbackByGame 统计游戏的反馈数量
func (r *PreferenceRepository) CountFeedbackByGame(ctx context.Context, userID int64, gameID string) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&model.RecommendationFeedback{}).
		Where("user_id = ? AND game_id = ?", userID, gameID).
		Count(&count).Error
	return count, err
}

// GetUserPreferences 获取用户偏好列表（非分页）
func (r *PreferenceRepository) GetPreferences(ctx context.Context, userID int64) ([]model.UserGamePreference, error) {
	var prefs []model.UserGamePreference
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("weight DESC").
		Find(&prefs).Error
	return prefs, err
}

// FindByTag 根据标签查找偏好
func (r *PreferenceRepository) FindByTag(ctx context.Context, userID int64, tag string) (*model.UserGamePreference, error) {
	var pref model.UserGamePreference
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND tag = ?", userID, tag).
		First(&pref).Error
	if err != nil {
		return nil, err
	}
	return &pref, nil
}

// Create 创建偏好记录
func (r *PreferenceRepository) Create(ctx context.Context, pref *model.UserGamePreference) error {
	return r.db.WithContext(ctx).Create(pref).Error
}

// Update 更新偏好记录
func (r *PreferenceRepository) Update(ctx context.Context, pref *model.UserGamePreference) error {
	return r.db.WithContext(ctx).Save(pref).Error
}
