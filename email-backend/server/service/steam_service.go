// Package service 业务逻辑层 - Steam
package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"email-backend/server/model"
	"email-backend/server/pkg/agent"
	"email-backend/server/pkg/steam"
	"email-backend/server/repository"
)

// SteamService Steam业务服务
type SteamService struct {
	steamRepo   *repository.SteamRepository
	emailRepo   *repository.EmailRepository
	agentClient *agent.Client
	steamClient *steam.Client
}

// NewSteamService 创建Steam服务
func NewSteamService(steamRepo *repository.SteamRepository, emailRepo *repository.EmailRepository, agentClient *agent.Client) *SteamService {
	// 创建Steam客户端（默认使用Mock模式）
	steamClient := steam.NewClient(&steam.Config{
		UseMock: true, // 暂不使用真实Steam API
	})

	return &SteamService{
		steamRepo:   steamRepo,
		emailRepo:   emailRepo,
		agentClient: agentClient,
		steamClient: steamClient,
	}
}

// ==================== 账号绑定 ====================

// BindSteamAccount 绑定Steam账号
func (s *SteamService) BindSteamAccount(ctx context.Context, userID int64, steamID string) (*model.SteamAccount, error) {
	// 验证SteamID格式
	if !steam.IsValidSteamID(steamID) {
		return nil, fmt.Errorf("无效的Steam ID格式")
	}

	// 检查是否已绑定
	existing, _ := s.steamRepo.FindAccountByUserID(ctx, userID)
	if existing != nil {
		return nil, fmt.Errorf("已绑定Steam账号，请先解绑")
	}

	// 获取Steam用户资料
	profile, err := s.steamClient.GetPlayerSummaries(ctx, steamID)
	if err != nil {
		return nil, fmt.Errorf("获取Steam资料失败: %w", err)
	}

	// 创建绑定记录
	account := &model.SteamAccount{
		UserID:        userID,
		SteamID:       steamID,
		SteamNickname: profile.PersonaName,
		AvatarURL:     profile.AvatarFull,
		ProfileURL:    profile.ProfileURL,
		RealName:      profile.RealName,
		Location:      profile.Location,
		IsActive:      true,
	}

	if err := s.steamRepo.CreateAccount(ctx, account); err != nil {
		return nil, fmt.Errorf("绑定失败: %w", err)
	}

	// 自动同步游戏库
	go s.SyncGameLibrary(context.Background(), userID)

	return account, nil
}

// UnbindSteamAccount 解绑Steam账号
func (s *SteamService) UnbindSteamAccount(ctx context.Context, userID int64) error {
	account, err := s.steamRepo.FindAccountByUserID(ctx, userID)
	if err != nil {
		return fmt.Errorf("未找到绑定的Steam账号")
	}

	// 删除游戏库数据
	if err := s.steamRepo.DeleteLibraryByAccount(ctx, account.ID); err != nil {
		return fmt.Errorf("清理游戏库失败: %w", err)
	}

	// 删除绑定记录
	if err := s.steamRepo.DeleteAccount(ctx, userID); err != nil {
		return fmt.Errorf("解绑失败: %w", err)
	}

	return nil
}

// GetSteamAccount 获取Steam账号信息
func (s *SteamService) GetSteamAccount(ctx context.Context, userID int64) (*model.SteamAccount, error) {
	return s.steamRepo.FindAccountByUserID(ctx, userID)
}

// ==================== 游戏库同步 ====================

// SyncGameLibrary 同步用户Steam游戏库
func (s *SteamService) SyncGameLibrary(ctx context.Context, userID int64) error {
	account, err := s.steamRepo.FindAccountByUserID(ctx, userID)
	if err != nil {
		return fmt.Errorf("未绑定Steam账号")
	}

	// 获取用户游戏库
	games, err := s.steamClient.GetOwnedGames(ctx, account.SteamID)
	if err != nil {
		return fmt.Errorf("获取游戏库失败: %w", err)
	}

	// 获取最近游玩
	recentGames, _ := s.steamClient.GetRecentlyPlayedGames(ctx, account.SteamID)

	// 构建最近游玩map
	recentMap := make(map[int]steam.RecentGameInfo)
	for _, g := range recentGames {
		recentMap[g.AppID] = g
	}

	// 转换为LibraryItem
	items := make([]*model.SteamLibraryItem, 0, len(games))
	for _, game := range games {
		item := &model.SteamLibraryItem{
			UserID:       userID,
			AccountID:    account.ID,
			GameID:       fmt.Sprintf("%d", game.AppID),
			GameName:     game.Name,
			Playtime:     game.PlaytimeForever,
			IconURL:      fmt.Sprintf("https://media.steampowered.com/steamcommunity/public/images/apps/%d/%s.jpg", game.AppID, game.ImgIconURL),
			IsSynced:     false,
		}

		// 如果有最近游玩数据
		if recent, ok := recentMap[game.AppID]; ok {
			item.Playtime2Weeks = recent.Playtime2Weeks
			if recent.LastPlayed > 0 {
				t := time.Unix(recent.LastPlayed, 0)
				item.LastPlayedAt = &t
			}
		}

		items = append(items, item)

		// 同时更新/创建SteamGame记录（用于推荐）
		syncGame := &model.SteamGame{
			UserID:   userID,
			GameID:   fmt.Sprintf("%d", game.AppID),
			GameName: game.Name,
			Genre:    steam.GetGameGenreMock(game.AppID),
			Tags:     steam.GetGameTagsMock(game.AppID),
			StoreURL: fmt.Sprintf("https://store.steampowered.com/app/%d", game.AppID),
			Playtime: game.PlaytimeForever,
			IsOwned:  true,
		}
		_ = s.steamRepo.UpsertGame(ctx, syncGame)
	}

	// 批量保存
	if err := s.steamRepo.BatchUpsertLibraryItems(ctx, items); err != nil {
		return fmt.Errorf("保存游戏库失败: %w", err)
	}

	// 更新同步时间
	now := time.Now()
	account.LastSyncAt = &now
	_ = s.steamRepo.UpdateAccount(ctx, account)

	fmt.Printf("Steam游戏库同步完成: 用户%d, %d款游戏\n", userID, len(items))
	return nil
}

// ListGameLibrary 获取用户游戏库
func (s *SteamService) ListGameLibrary(ctx context.Context, userID int64, page, pageSize int, sortBy string) ([]*model.SteamLibraryItem, int64, error) {
	return s.steamRepo.ListLibraryByUser(ctx, userID, page, pageSize, sortBy)
}

// ListRecentPlayed 获取最近游玩的游戏
func (s *SteamService) ListRecentPlayed(ctx context.Context, userID int64, limit int) ([]*model.SteamLibraryItem, error) {
	if limit <= 0 {
		limit = 10
	}
	return s.steamRepo.ListRecentPlayed(ctx, userID, limit)
}

// ==================== 原有功能 ====================

// ListGames 获取游戏列表
func (s *SteamService) ListGames(ctx context.Context, userID int64, page, pageSize int, keyword string) ([]*model.SteamGame, int64, error) {
	return s.steamRepo.ListGames(ctx, userID, page, pageSize, keyword)
}

// ListDeals 获取促销列表
func (s *SteamService) ListDeals(ctx context.Context, userID int64, page, pageSize int, sortBy string, activeOnly bool) ([]*model.SteamDeal, int64, error) {
	// 先过期处理
	_ = s.steamRepo.ExpireDeals(ctx)
	return s.steamRepo.ListDeals(ctx, userID, page, pageSize, sortBy, activeOnly)
}

// GetDealByID 获取促销详情
func (s *SteamService) GetDealByID(ctx context.Context, id int64) (*model.SteamDeal, error) {
	return s.steamRepo.FindDealByID(ctx, id)
}

// ExtractSteamInfo 从邮件中提取Steam游戏信息
func (s *SteamService) ExtractSteamInfo(ctx context.Context, emailID int64, userID int64) error {
	// 获取邮件内容
	email, err := s.emailRepo.FindByID(ctx, emailID)
	if err != nil {
		return fmt.Errorf("邮件不存在: %w", err)
	}

	// 调用Agent提取Steam信息
	req := &agent.SteamExtractRequest{
		EmailID:     fmt.Sprintf("%d", emailID),
		Subject:     email.Subject,
		SenderEmail: email.SenderEmail,
		Content:     email.Content,
		ContentHTML: email.ContentHTML,
	}

	resp, err := s.agentClient.SteamExtract(ctx, req)
	if err != nil {
		return fmt.Errorf("Agent提取失败: %w", err)
	}

	// 保存提取的游戏信息
	for _, gameInfo := range resp.Games {
		tagsJSON, _ := json.Marshal(gameInfo.Tags)

		// 保存/更新游戏
		game := &model.SteamGame{
			UserID:   userID,
			GameName: gameInfo.Name,
			GameID:   gameInfo.AppID,
			Genre:    gameInfo.Genre,
			Tags:     string(tagsJSON),
			CoverURL: gameInfo.CoverURL,
			StoreURL: gameInfo.StoreURL,
		}

		if err := s.steamRepo.UpsertGame(ctx, game); err != nil {
			continue // 单个失败不影响其他
		}

		// 如果有促销信息，保存促销
		if gameInfo.HasDeal {
			deal := &model.SteamDeal{
				UserID:        userID,
				GameID:        gameInfo.AppID,
				GameName:      gameInfo.Name,
				OriginalPrice: gameInfo.OriginalPrice,
				DealPrice:     gameInfo.DealPrice,
				Discount:      gameInfo.Discount,
				CoverURL:      gameInfo.CoverURL,
				StoreURL:      gameInfo.StoreURL,
				IsActive:      true,
				EmailID:       emailID,
			}

			if gameInfo.DealEnd != "" {
				if t, err := time.Parse("2006-01-02", gameInfo.DealEnd); err == nil {
					deal.EndDate = &t
				}
			}

			if err := s.steamRepo.CreateDeal(ctx, deal); err != nil {
				continue
			}
		}
	}

	return nil
}

// GetSteamStats 获取Steam统计概览
func (s *SteamService) GetSteamStats(ctx context.Context, userID int64) (map[string]interface{}, error) {
	activeDeals, err := s.steamRepo.CountActiveDeals(ctx, userID)
	if err != nil {
		return nil, err
	}

	_, totalGames, err := s.steamRepo.ListGames(ctx, userID, 1, 1, "")
	if err != nil {
		return nil, err
	}

	// 获取绑定状态
	account, _ := s.steamRepo.FindAccountByUserID(ctx, userID)

	result := map[string]interface{}{
		"total_games":  totalGames,
		"active_deals": activeDeals,
		"is_bound":     account != nil,
	}

	if account != nil {
		result["steam_nickname"] = account.SteamNickname
		result["avatar_url"] = account.AvatarURL
		if account.LastSyncAt != nil {
			result["last_sync"] = account.LastSyncAt.Format("2006-01-02 15:04:05")
		}
	}

	return result, nil
}