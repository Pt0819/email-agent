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

// ==================== Account 操作 ====================

// FindAccountByUserID 根据用户ID查找Steam账号
func (r *SteamRepository) FindAccountByUserID(ctx context.Context, userID int64) (*model.SteamAccount, error) {
	var account model.SteamAccount
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND is_active = ?", userID, true).
		First(&account).Error
	if err != nil {
		return nil, err
	}
	return &account, nil
}

// CreateAccount 创建Steam账号绑定
func (r *SteamRepository) CreateAccount(ctx context.Context, account *model.SteamAccount) error {
	return r.db.WithContext(ctx).Create(account).Error
}

// UpdateAccount 更新Steam账号信息
func (r *SteamRepository) UpdateAccount(ctx context.Context, account *model.SteamAccount) error {
	return r.db.WithContext(ctx).Save(account).Error
}

// DeleteAccount 软删除Steam账号
func (r *SteamRepository) DeleteAccount(ctx context.Context, userID int64) error {
	return r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Delete(&model.SteamAccount{}).Error
}

// ==================== Library 操作 ====================

// UpsertLibraryItem 创建或更新游戏库条目
func (r *SteamRepository) UpsertLibraryItem(ctx context.Context, item *model.SteamLibraryItem) error {
	var existing model.SteamLibraryItem
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND account_id = ? AND game_id = ?", item.UserID, item.AccountID, item.GameID).
		First(&existing).Error

	if err == gorm.ErrRecordNotFound {
		return r.db.WithContext(ctx).Create(item).Error
	}
	if err != nil {
		return err
	}

	return r.db.WithContext(ctx).Model(&existing).Updates(map[string]interface{}{
		"game_name":        item.GameName,
		"playtime":         item.Playtime,
		"playtime_2_weeks": item.Playtime2Weeks,
		"last_played_at":   item.LastPlayedAt,
		"icon_url":         item.IconURL,
	}).Error
}

// BatchUpsertLibraryItems 批量创建或更新游戏库
func (r *SteamRepository) BatchUpsertLibraryItems(ctx context.Context, items []*model.SteamLibraryItem) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, item := range items {
			var existing model.SteamLibraryItem
			err := tx.Where("user_id = ? AND account_id = ? AND game_id = ?",
				item.UserID, item.AccountID, item.GameID).First(&existing).Error

			if err == gorm.ErrRecordNotFound {
				if err := tx.Create(item).Error; err != nil {
					return err
				}
			} else if err != nil {
				return err
			} else {
				if err := tx.Model(&existing).Updates(map[string]interface{}{
					"game_name":        item.GameName,
					"playtime":         item.Playtime,
					"playtime_2_weeks": item.Playtime2Weeks,
					"last_played_at":   item.LastPlayedAt,
					"icon_url":         item.IconURL,
				}).Error; err != nil {
					return err
				}
			}
		}
		return nil
	})
}

// ListLibraryByUser 分页获取用户游戏库
func (r *SteamRepository) ListLibraryByUser(ctx context.Context, userID int64, page, pageSize int, sortBy string) ([]*model.SteamLibraryItem, int64, error) {
	var items []*model.SteamLibraryItem
	var total int64

	query := r.db.WithContext(ctx).Model(&model.SteamLibraryItem{}).Where("user_id = ?", userID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	orderClause := "playtime DESC"
	switch sortBy {
	case "name":
		orderClause = "game_name ASC"
	case "recent":
		orderClause = "last_played_at DESC"
	case "playtime":
		orderClause = "playtime DESC"
	}

	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).
		Order(orderClause).
		Find(&items).Error; err != nil {
		return nil, 0, err
	}

	return items, total, nil
}

// ListRecentPlayed 获取最近游玩的游戏
func (r *SteamRepository) ListRecentPlayed(ctx context.Context, userID int64, limit int) ([]*model.SteamLibraryItem, error) {
	var items []*model.SteamLibraryItem
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND playtime_2_weeks > ?", userID, 0).
		Order("playtime_2_weeks DESC").
		Limit(limit).
		Find(&items).Error
	return items, err
}

// DeleteLibraryByAccount 删除账号关联的所有游戏库数据
func (r *SteamRepository) DeleteLibraryByAccount(ctx context.Context, accountID int64) error {
	return r.db.WithContext(ctx).
		Where("account_id = ?", accountID).
		Delete(&model.SteamLibraryItem{}).Error
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
	return r.db.WithContext(ctx).Model(&existing).Updates(map[string]interface{}{
		"game_name":  game.GameName,
		"developer":  game.Developer,
		"publisher":  game.Publisher,
		"genre":      game.Genre,
		"tags":       game.Tags,
		"cover_url":  game.CoverURL,
		"store_url":  game.StoreURL,
		"playtime":   game.Playtime,
		"is_owned":   game.IsOwned,
	}).Error
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
