// Package repository 数据访问层 - 推荐反馈
package repository

import (
	"context"

	"email-backend/server/model"

	"gorm.io/gorm"
)

// FeedbackRepository 推荐反馈仓库
type FeedbackRepository struct {
	db *gorm.DB
}

// NewFeedbackRepository 创建反馈仓库
func NewFeedbackRepository(db *gorm.DB) *FeedbackRepository {
	return &FeedbackRepository{db: db}
}

// Create 创建反馈记录
func (r *FeedbackRepository) Create(ctx context.Context, feedback *model.RecommendationFeedback) error {
	return r.db.WithContext(ctx).Create(feedback).Error
}

// FindByUserAndGame 根据用户和游戏查找反馈
func (r *FeedbackRepository) FindByUserAndGame(ctx context.Context, userID int64, gameID string) (*model.RecommendationFeedback, error) {
	var feedback model.RecommendationFeedback
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND game_id = ?", userID, gameID).
		Order("created_at DESC").
		First(&feedback).Error
	if err != nil {
		return nil, err
	}
	return &feedback, nil
}

// ListByUser 获取用户反馈列表
func (r *FeedbackRepository) ListByUser(ctx context.Context, userID int64, limit int) ([]*model.RecommendationFeedback, error) {
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

// CountByAction 统计某类反馈数量
func (r *FeedbackRepository) CountByAction(ctx context.Context, userID int64, action string) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&model.RecommendationFeedback{}).
		Where("user_id = ? AND action = ?", userID, action).
		Count(&count).Error
	return count, err
}
