// Package repository 数据访问层 - Steam
package repository

import (
	"context"
	"time"

	"email-backend/server/model"

	"gorm.io/gorm"
)

// SteamRepository Steam数据访问
type SteamRepository struct {
	db *gorm.DB
}

// NewSteamRepository 创建Steam Repository
func NewSteamRepository(db *gorm.DB) *SteamRepository {
	return &SteamRepository{db: db}
}

// ==================== Game 操作 ====================

// FindGameByID 根据ID查询游戏
func (r *SteamRepository) FindGameByID(ctx context.Context, id int64) (*model.SteamGame, error) {
	var game model.SteamGame
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&game).Error
	if err != nil {
		return nil, err
	}
	return &game, nil
}

// FindGameByGameID 根据Steam GameID查询游戏
func (r *SteamRepository) FindGameByGameID(ctx context.Context, userID int64, gameID string) (*model.SteamGame, error) {
	var game model.SteamGame
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND game_id = ?", userID, gameID).
		First(&game).Error
	if err != nil {
		return nil, err
	}
	return &game, nil
}

// ListGames 分页查询游戏列表
func (r *SteamRepository) ListGames(ctx context.Context, userID int64, page, pageSize int, keyword string) ([]*model.SteamGame, int64, error) {
	var games []*model.SteamGame
	var total int64

	query := r.db.WithContext(ctx).Model(&model.SteamGame{}).Where("user_id = ?", userID)

	if keyword != "" {
		kw := "%" + keyword + "%"
		query = query.Where("game_name LIKE ? OR developer LIKE ?", kw, kw)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).
		Order("playtime DESC").
		Find(&games).Error; err != nil {
		return nil, 0, err
	}

	return games, total, nil
}

// UpsertGame 创建或更新游戏（按GameID去重）
func (r *SteamRepository) UpsertGame(ctx context.Context, game *model.SteamGame) error {
	var existing model.SteamGame
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND game_id = ?", game.UserID, game.GameID).
		First(&existing).Error

	if err == gorm.ErrRecordNotFound {
		return r.db.WithContext(ctx).Create(game).Error
	}
	if err != nil {
		return err
	}

	// 更新已有记录
	result := r.db.WithContext(ctx).Model(&existing).Updates(map[string]interface{}{
		"game_name":  game.GameName,
		"developer":  game.Developer,
		"publisher":  game.Publisher,
		"genre":      game.Genre,
		"tags":       game.Tags,
		"cover_url":  game.CoverURL,
		"store_url":  game.StoreURL,
		"playtime":   game.Playtime,
		"is_owned":   game.IsOwned,
	})
	return result.Error
}

// ==================== Deal 操作 ====================

// FindDealByID 根据ID查询促销
func (r *SteamRepository) FindDealByID(ctx context.Context, id int64) (*model.SteamDeal, error) {
	var deal model.SteamDeal
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&deal).Error
	if err != nil {
		return nil, err
	}
	return &deal, nil
}

// ListDeals 分页查询促销列表
func (r *SteamRepository) ListDeals(ctx context.Context, userID int64, page, pageSize int, sortBy string, activeOnly bool) ([]*model.SteamDeal, int64, error) {
	var deals []*model.SteamDeal
	var total int64

	query := r.db.WithContext(ctx).Model(&model.SteamDeal{}).Where("user_id = ?", userID)

	if activeOnly {
		query = query.Where("is_active = ?", true)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 排序
	orderClause := "created_at DESC"
	switch sortBy {
	case "discount":
		orderClause = "discount DESC"
	case "price_asc":
		orderClause = "deal_price ASC"
	case "price_desc":
		orderClause = "deal_price DESC"
	case "end_date":
		orderClause = "end_date ASC"
	}

	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).
		Order(orderClause).
		Find(&deals).Error; err != nil {
		return nil, 0, err
	}

	return deals, total, nil
}

// CreateDeal 创建促销记录
func (r *SteamRepository) CreateDeal(ctx context.Context, deal *model.SteamDeal) error {
	return r.db.WithContext(ctx).Create(deal).Error
}

// UpdateDeal 更新促销记录
func (r *SteamRepository) UpdateDeal(ctx context.Context, deal *model.SteamDeal) error {
	return r.db.WithContext(ctx).Save(deal).Error
}

// ExpireDeals 将过期促销标记为不活跃
func (r *SteamRepository) ExpireDeals(ctx context.Context) error {
	now := time.Now()
	return r.db.WithContext(ctx).
		Model(&model.SteamDeal{}).
		Where("is_active = ? AND end_date IS NOT NULL AND end_date < ?", true, now).
		Update("is_active", false).Error
}

// CountActiveDeals 统计活跃促销数量
func (r *SteamRepository) CountActiveDeals(ctx context.Context, userID int64) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&model.SteamDeal{}).
		Where("user_id = ? AND is_active = ?", userID, true).
		Count(&count).Error
	return count, err
}
