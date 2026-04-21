// Package service 业务逻辑层 - Steam
package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"email-backend/server/model"
	"email-backend/server/pkg/agent"
	"email-backend/server/repository"
)

// SteamService Steam业务服务
type SteamService struct {
	steamRepo   *repository.SteamRepository
	emailRepo   *repository.EmailRepository
	agentClient *agent.Client
}

// NewSteamService 创建Steam服务
func NewSteamService(steamRepo *repository.SteamRepository, emailRepo *repository.EmailRepository, agentClient *agent.Client) *SteamService {
	return &SteamService{
		steamRepo:   steamRepo,
		emailRepo:   emailRepo,
		agentClient: agentClient,
	}
}

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

	return map[string]interface{}{
		"total_games":   totalGames,
		"active_deals":  activeDeals,
	}, nil
}
