// Package repository 数据访问层 - 偏好洞察记录
package repository

import (
	"context"

	"email-backend/server/model"

	"gorm.io/gorm"
)

// PreferenceInsightRepository 偏好洞察仓库
type PreferenceInsightRepository struct {
	db *gorm.DB
}

// NewPreferenceInsightRepository 创建偏好洞察仓库
func NewPreferenceInsightRepository(db *gorm.DB) *PreferenceInsightRepository {
	return &PreferenceInsightRepository{db: db}
}

// Create 创建洞察记录
func (r *PreferenceInsightRepository) Create(ctx context.Context, insight *model.PreferenceInsight) error {
	return r.db.WithContext(ctx).Create(insight).Error
}

// GetByUserID 获取用户洞察列表（分页）
func (r *PreferenceInsightRepository) GetByUserID(ctx context.Context, userID int64, page, pageSize int) ([]*model.PreferenceInsight, int64, error) {
	var insights []*model.PreferenceInsight
	var total int64

	query := r.db.WithContext(ctx).Model(&model.PreferenceInsight{}).
		Where("user_id = ?", userID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.
		Offset(offset).
		Limit(pageSize).
		Order("created_at DESC").
		Find(&insights).Error; err != nil {
		return nil, 0, err
	}

	return insights, total, nil
}

// GetRecentAnomalies 获取最近的异常洞察
func (r *PreferenceInsightRepository) GetRecentAnomalies(ctx context.Context, userID int64, limit int) ([]*model.PreferenceInsight, error) {
	if limit <= 0 {
		limit = 10
	}

	var insights []*model.PreferenceInsight
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND is_anomaly = ?", userID, true).
		Order("created_at DESC").
		Limit(limit).
		Find(&insights).Error
	return insights, err
}

// CountByUserID 统计用户洞察数量
func (r *PreferenceInsightRepository) CountByUserID(ctx context.Context, userID int64) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&model.PreferenceInsight{}).
		Where("user_id = ?", userID).
		Count(&count).Error
	return count, err
}

// BatchCreate 批量创建洞察
func (r *PreferenceInsightRepository) BatchCreate(ctx context.Context, insights []*model.PreferenceInsight) error {
	if len(insights) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, insight := range insights {
			if err := tx.Create(insight).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// GetLatestByEventType 获取特定事件类型的最新洞察
func (r *PreferenceInsightRepository) GetLatestByEventType(ctx context.Context, userID int64, eventType string, limit int) ([]*model.PreferenceInsight, error) {
	if limit <= 0 {
		limit = 10
	}

	var insights []*model.PreferenceInsight
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND event_type = ?", userID, eventType).
		Order("created_at DESC").
		Limit(limit).
		Find(&insights).Error
	return insights, err
}
