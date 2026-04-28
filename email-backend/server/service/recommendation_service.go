// Package service 业务逻辑层 - 推荐
package service

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"email-backend/server/model"
	"email-backend/server/pkg/agent"
	"email-backend/server/repository"
)

// RecommendationService 推荐业务服务
type RecommendationService struct {
	recRepo       *repository.RecommendationRepository
	steamRepo     *repository.SteamRepository
	prefRepo      *repository.PreferenceRepository
	feedbackRepo  *repository.FeedbackRepository
	agentClient   *agent.Client
}

// NewRecommendationService 创建推荐服务
func NewRecommendationService(
	recRepo *repository.RecommendationRepository,
	steamRepo *repository.SteamRepository,
	prefRepo *repository.PreferenceRepository,
	feedbackRepo *repository.FeedbackRepository,
	agentClient *agent.Client,
) *RecommendationService {
	return &RecommendationService{
		recRepo:      recRepo,
		steamRepo:    steamRepo,
		prefRepo:     prefRepo,
		feedbackRepo: feedbackRepo,
		agentClient:  agentClient,
	}
}

// ==================== 推荐生成 ====================

// GenerateRecommendations 生成个性化推荐
// 核心算法：基于用户偏好标签与游戏标签的匹配度计算
func (s *RecommendationService) GenerateRecommendations(ctx context.Context, userID int64, req *model.RecommendationGenerateRequest) (*model.RecommendationListResponse, error) {
	// 获取用户已拥有的游戏
	ownedGames, err := s.recRepo.GetOwnedGameIDs(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("获取已拥有游戏失败: %w", err)
	}

	// 获取用户偏好
	prefs, err := s.recRepo.GetUserPreferences(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("获取用户偏好失败: %w", err)
	}

	// 构建偏好权重map
	prefWeight := make(map[string]float64)
	prefSources := make(map[string]string)
	for _, p := range prefs {
		prefWeight[p.Tag] = p.Weight
		prefSources[p.Tag] = p.Source
	}

	// 获取候选游戏
	var candidates []*model.SteamGame
	if len(req.GameIDs) > 0 {
		// 指定游戏
		for _, gid := range req.GameIDs {
			game, err := s.steamRepo.FindGameByGameID(ctx, userID, gid)
			if err == nil && game != nil {
				candidates = append(candidates, game)
			}
		}
	} else if req.DealOnly {
		// 仅促销游戏
		deals, err := s.recRepo.GetActiveDealsWithoutRecommendation(ctx, userID, 100)
		if err != nil {
			return nil, fmt.Errorf("获取促销游戏失败: %w", err)
		}
		for i := range deals {
			game, err := s.steamRepo.FindGameByGameID(ctx, userID, deals[i].GameID)
			if err == nil && game != nil {
				candidates = append(candidates, game)
			}
		}
	} else {
		// 全量推荐 - 获取所有未拥有的游戏
		page := 1
		pageSize := 100
		for {
			games, total, err := s.steamRepo.ListGames(ctx, userID, page, pageSize, "")
			if err != nil {
				break
			}
			for _, g := range games {
				if !ownedGames[g.GameID] {
					candidates = append(candidates, g)
				}
			}
			if int(total) <= page*pageSize {
				break
			}
			page++
		}
	}

	// 计算匹配度并排序
	type scoredGame struct {
		game    *model.SteamGame
		score   float64
		matches []string
	}

	var scoredGames []scoredGame
	for _, game := range candidates {
		// 跳过已拥有的游戏
		if ownedGames[game.GameID] {
			continue
		}

		score, matches := s.calculateMatchScore(game, prefWeight)
		if req.MinScore > 0 && score < req.MinScore {
			continue
		}

		scoredGames = append(scoredGames, scoredGame{
			game:    game,
			score:   score,
			matches: matches,
		})
	}

	// 按匹配度排序
	sort.Slice(scoredGames, func(i, j int) bool {
		return scoredGames[i].score > scoredGames[j].score
	})

	// 取前N个
	maxCount := req.MaxCount
	if maxCount <= 0 || maxCount > 50 {
		maxCount = 20
	}
	if len(scoredGames) > maxCount {
		scoredGames = scoredGames[:maxCount]
	}

	// 构建推荐记录
	recs := make([]*model.GameRecommendation, 0, len(scoredGames))
	now := time.Now()
	for _, sg := range scoredGames {
		rec := &model.GameRecommendation{
			UserID:       userID,
			GameID:       sg.game.GameID,
			GameName:     sg.game.GameName,
			GameGenre:    sg.game.Genre,
			GameTags:     sg.game.Tags,
			CoverURL:     sg.game.CoverURL,
			StoreURL:     sg.game.StoreURL,
			MatchScore:   sg.score,
			MatchReasons: toJSON(sg.matches),
			Source:       model.RecSourceAuto,
			Status:       model.RecStatusActive,
			CreatedAt:    now,
		}

		// 检查是否有促销
		if req.DealOnly || sg.game.CoverURL != "" {
			// 尝试获取促销信息
			deals, _, _ := s.steamRepo.ListDeals(ctx, userID, 1, 1, "created_at", true)
			for _, d := range deals {
				if d.GameID == sg.game.GameID && d.IsActive {
					rec.DealID = &d.ID
					rec.DealPrice = d.DealPrice
					rec.DealDiscount = d.Discount
					rec.DealEndDate = d.EndDate
					break
				}
			}
		}

		recs = append(recs, rec)
	}

	// 批量保存
	if len(recs) > 0 {
		// 先删除旧的active推荐（避免重复）
		_ = s.recRepo.DeleteByUser(ctx, userID)
		if err := s.recRepo.BatchCreate(ctx, recs); err != nil {
			return nil, fmt.Errorf("保存推荐失败: %w", err)
		}
	}

	// 响应
	page := 1
	pageSize := len(recs)
	if pageSize == 0 {
		pageSize = 20
	}
	return s.ListRecommendations(ctx, userID, page, pageSize, "", false)
}

// calculateMatchScore 计算游戏与用户偏好的匹配度
// 返回: 匹配度分数(0-100), 匹配理由列表
func (s *RecommendationService) calculateMatchScore(game *model.SteamGame, userPrefs map[string]float64) (float64, []string) {
	if len(userPrefs) == 0 {
		// 无偏好时返回默认分数
		return 50.0, []string{"根据你的Steam游戏库推荐"}
	}

	var totalScore float64
	var maxPossibleScore float64
	matches := []string{}

	// 解析游戏标签
	gameTags := parseTags(game.Tags)

	// 匹配标签
	for tag, weight := range userPrefs {
		maxPossibleScore += weight * 2 // 每个偏好标签最高2分

		// 直接标签匹配
		for _, gameTag := range gameTags {
			if strings.EqualFold(tag, gameTag) || containsTag(gameTag, tag) {
				totalScore += weight * 2
				matches = append(matches, fmt.Sprintf("你偏好%s类游戏", tag))
				break
			}
		}

		// 标签包含匹配
		for _, gameTag := range gameTags {
			if strings.Contains(strings.ToLower(gameTag), strings.ToLower(tag)) {
				totalScore += weight
				matches = append(matches, fmt.Sprintf("包含\"%s\"标签", tag))
			}
		}
	}

	// 类型匹配
	if game.Genre != "" {
		for tag, weight := range userPrefs {
			if strings.Contains(strings.ToLower(game.Genre), strings.ToLower(tag)) {
				totalScore += weight * 1.5
				matches = append(matches, fmt.Sprintf("你常玩%s类型", tag))
			}
		}
	}

	// 计算最终分数（归一化到0-100）
	score := 50.0 // 基础分
	if maxPossibleScore > 0 {
		score = (totalScore / maxPossibleScore) * 50 // 匹配贡献最多50分
	}
	score += 50 // 基础分

	if score > 100 {
		score = 100
	}

	// 去重匹配理由
	seen := make(map[string]bool)
	uniqueMatches := []string{}
	for _, m := range matches {
		if !seen[m] {
			seen[m] = true
			uniqueMatches = append(uniqueMatches, m)
		}
	}

	// 如果没有匹配，使用通用理由
	if len(uniqueMatches) == 0 {
		uniqueMatches = append(uniqueMatches, "热门游戏推荐")
	}

	return score, uniqueMatches
}

// parseTags 解析JSON标签数组
func parseTags(tagsJSON string) []string {
	if tagsJSON == "" {
		return []string{}
	}

	var tags []string
	if err := json.Unmarshal([]byte(tagsJSON), &tags); err != nil {
		// 尝试逗号分隔
		return strings.Split(tagsJSON, ",")
	}
	return tags
}

// containsTag 检查tag是否包含关键字
func containsTag(tag, keyword string) bool {
	tag = strings.ToLower(strings.TrimSpace(tag))
	keyword = strings.ToLower(strings.TrimSpace(keyword))

	// 常见同义词映射
	synonyms := map[string][]string{
		"rpg":        {"role playing", "jrpg", "arpg"},
		"fps":        {"first person", "shooter"},
		"act":        {"action", "beat 'em up"},
		"strategy":   {"strategic", "turn based"},
		"simulation": {"sim", "management"},
	}

	for _, syn := range synonyms[keyword] {
		if strings.Contains(tag, syn) {
			return true
		}
	}

	return strings.Contains(tag, keyword)
}

// ==================== 推荐列表 ====================

// ListRecommendations 获取推荐列表
func (s *RecommendationService) ListRecommendations(ctx context.Context, userID int64, page, pageSize int, status string, dealOnly bool) (*model.RecommendationListResponse, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 50 {
		pageSize = 20
	}

	recs, total, err := s.recRepo.List(ctx, userID, page, pageSize, status, dealOnly)
	if err != nil {
		return nil, err
	}

	// 转换响应
	list := make([]model.RecommendationResponse, 0, len(recs))
	for _, rec := range recs {
		list = append(list, s.recRepo.ToResponse(rec))
	}

	// 获取统计
	stats, _ := s.recRepo.GetOrCreateStats(ctx, userID)
	statsSummary := &model.RecStatsSummary{
		TotalRecs:      stats.TotalRecs,
		ClickedCount:   stats.ClickedCount,
		PurchaseCount:  stats.PurchasedCount,
		CTR:            stats.CTR,
		PurchaseRate:   stats.PurchaseRate,
	}

	return &model.RecommendationListResponse{
		List:     list,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
		Stats:    statsSummary,
	}, nil
}

// GetRecommendationByID 获取推荐详情
func (s *RecommendationService) GetRecommendationByID(ctx context.Context, id int64, userID int64) (*model.RecommendationResponse, error) {
	rec, err := s.recRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("推荐不存在")
	}

	if rec.UserID != userID {
		return nil, fmt.Errorf("无权访问此推荐")
	}

	resp := s.recRepo.ToResponse(rec)
	return &resp, nil
}

// ==================== 用户反馈 ====================

// ProcessFeedback 处理用户反馈
func (s *RecommendationService) ProcessFeedback(ctx context.Context, userID int64, id int64, action string) error {
	rec, err := s.recRepo.FindByID(ctx, id)
	if err != nil {
		return fmt.Errorf("推荐不存在")
	}

	if rec.UserID != userID {
		return fmt.Errorf("无权操作此推荐")
	}

	// 更新推荐状态
	var newStatus string
	switch action {
	case "click":
		newStatus = model.RecStatusClicked
	case "purchase":
		newStatus = model.RecStatusPurchased
	case "ignore":
		newStatus = model.RecStatusIgnored
	case "like":
		// 点赞不影响主状态，但记录正反馈
		_ = s.recordPositiveFeedback(ctx, userID, rec)
		return nil
	case "dislike":
		// 点踩降低相关标签权重
		_ = s.processDislike(ctx, userID, rec)
		newStatus = model.RecStatusIgnored
	default:
		return fmt.Errorf("无效的反馈动作")
	}

	// 更新状态
	if err := s.recRepo.UpdateStatus(ctx, id, newStatus); err != nil {
		return err
	}

	// 更新统计
	statsDelta := map[string]int{}
	switch newStatus {
	case model.RecStatusClicked:
		statsDelta["clicked"] = 1
	case model.RecStatusPurchased:
		statsDelta["purchased"] = 1
	case model.RecStatusIgnored:
		statsDelta["ignored"] = 1
	}
	_ = s.recRepo.UpdateStats(ctx, userID, statsDelta)

	// 如果是购买，记录到反馈表
	if newStatus == model.RecStatusPurchased {
		feedback := &model.RecommendationFeedback{
			UserID:   userID,
			GameID:   rec.GameID,
			GameName: rec.GameName,
			Action:   model.FeedbackActionPurchased,
			DealID:   rec.DealID,
		}
		_ = s.feedbackRepo.Create(ctx, feedback)

		// 增强偏好 - 购买增加相关标签权重
		_ = s.enhancePreferenceFromPurchase(ctx, userID, rec)
	}

	return nil
}

// recordPositiveFeedback 记录正反馈，增强相关标签
func (s *RecommendationService) recordPositiveFeedback(ctx context.Context, userID int64, rec *model.GameRecommendation) error {
	// 解析标签
	tags := parseTags(rec.GameTags)
	for _, tag := range tags {
		tag = strings.TrimSpace(tag)
		if tag == "" {
			continue
		}
		// 查找或创建偏好记录
		existing, err := s.prefRepo.FindByTag(ctx, userID, tag)
		if err != nil || existing == nil {
			// 创建新偏好
			pref := &model.UserGamePreference{
				UserID:  userID,
				Tag:     tag,
				Weight:  2.0, // 初始权重
				Source:  model.PreferenceSourceSystem,
			}
			_ = s.prefRepo.Create(ctx, pref)
		} else {
			// 增加权重
			existing.Weight += 0.5
			if existing.Weight > 10 {
				existing.Weight = 10
			}
			existing.Source = model.PreferenceSourceSystem
			_ = s.prefRepo.Update(ctx, existing)
		}
	}
	return nil
}

// processDislike 处理点踩，降低相关标签权重
func (s *RecommendationService) processDislike(ctx context.Context, userID int64, rec *model.GameRecommendation) error {
	tags := parseTags(rec.GameTags)
	for _, tag := range tags {
		tag = strings.TrimSpace(tag)
		if tag == "" {
			continue
		}
		existing, err := s.prefRepo.FindByTag(ctx, userID, tag)
		if err == nil && existing != nil {
			existing.Weight -= 1.0
			if existing.Weight < 0.5 {
				existing.Weight = 0.5
			}
			_ = s.prefRepo.Update(ctx, existing)
		}
	}
	return nil
}

// enhancePreferenceFromPurchase 购买增强偏好
func (s *RecommendationService) enhancePreferenceFromPurchase(ctx context.Context, userID int64, rec *model.GameRecommendation) error {
	tags := parseTags(rec.GameTags)
	for _, tag := range tags {
		tag = strings.TrimSpace(tag)
		if tag == "" {
			continue
		}
		existing, err := s.prefRepo.FindByTag(ctx, userID, tag)
		if err != nil || existing == nil {
			pref := &model.UserGamePreference{
				UserID:  userID,
				Tag:     tag,
				Weight:  3.0, // 购买给予更高初始权重
				Source:  model.PreferenceSourceEmail,
			}
			_ = s.prefRepo.Create(ctx, pref)
		} else {
			existing.Weight += 1.0
			if existing.Weight > 10 {
				existing.Weight = 10
			}
			existing.Source = model.PreferenceSourceEmail
			_ = s.prefRepo.Update(ctx, existing)
		}
	}
	return nil
}

// ==================== Agent协同 ====================

// GenerateWithLLM 使用LLM生成个性化推荐理由
func (s *RecommendationService) GenerateWithLLM(ctx context.Context, userID int64, gameID string) (string, error) {
	// 获取游戏信息
	game, err := s.steamRepo.FindGameByGameID(ctx, userID, gameID)
	if err != nil {
		return "", err
	}

	// 获取用户画像
	profile, err := s.GetUserProfile(ctx, userID)
	if err != nil {
		return "", err
	}

	// 构建游戏库数据
	var gameLibrary []agent.LibraryGameData
	if game != nil {
		gameLibrary = append(gameLibrary, agent.LibraryGameData{
			GameID:   game.GameID,
			GameName: game.GameName,
			Genre:    game.Genre,
			Tags:     game.Tags,
		})
	}

	// 调用Agent生成推荐理由
	req := &agent.PreferenceAnalyzeRequest{
		UserID:       userID,
		GameLibrary:  gameLibrary,
		CurrentPrefs: buildPreferenceContext(profile),
		TriggerType:  "recommendation",
	}

	resp, err := s.agentClient.PreferenceAnalyze(ctx, req)
	if err != nil || !resp.Success {
		// 回退到本地生成
		return s.generateLocalReason(game, profile), nil
	}

	if len(resp.Insights) > 0 {
		return resp.Insights[0], nil
	}

	return s.generateLocalReason(game, profile), nil
}

// GetUserProfile 获取用户画像摘要
func (s *RecommendationService) GetUserProfile(ctx context.Context, userID int64) (*model.UserGamingProfile, error) {
	prefs, err := s.prefRepo.GetPreferences(ctx, userID)
	if err != nil {
		return nil, err
	}

	// 获取top标签
	var topTags []model.TagPreference
	maxTags := 10
	if len(prefs) < maxTags {
		maxTags = len(prefs)
	}
	for i := 0; i < maxTags; i++ {
		topTags = append(topTags, model.TagPreference{
			Tag:    prefs[i].Tag,
			Weight: prefs[i].Weight,
			Source: prefs[i].Source,
		})
	}

	return &model.UserGamingProfile{
		UserID:  userID,
		TopTags: topTags,
	}, nil
}

// buildPreferenceContext 构建偏好上下文
func buildPreferenceContext(profile *model.UserGamingProfile) []agent.PreferenceTagData {
	var prefs []agent.PreferenceTagData
	for _, t := range profile.TopTags {
		prefs = append(prefs, agent.PreferenceTagData{
			Tag:    t.Tag,
			Weight: t.Weight,
			Source: t.Source,
		})
	}
	return prefs
}

// generateLocalReason 生成本地推荐理由
func (s *RecommendationService) generateLocalReason(game *model.SteamGame, profile *model.UserGamingProfile) string {
	if game == nil {
		return "根据你的游戏偏好推荐"
	}

	var reasons []string

	// 类型匹配
	if game.Genre != "" {
		reasons = append(reasons, fmt.Sprintf("与你常玩的%s类型游戏一致", game.Genre))
	}

	// 标签匹配
	gameTags := parseTags(game.Tags)
	for i, profileTag := range profile.TopTags {
		if i >= 3 {
			break
		}
		for _, gameTag := range gameTags {
			if strings.Contains(strings.ToLower(gameTag), strings.ToLower(profileTag.Tag)) {
				reasons = append(reasons, fmt.Sprintf("包含你喜欢的\"%s\"标签", profileTag.Tag))
				break
			}
		}
	}

	if len(reasons) == 0 {
		reasons = append(reasons, "热门游戏，值得一试")
	}

	return reasons[0]
}

// ==================== 工具方法 ====================

// toJSON 将slice转换为JSON字符串
func toJSON(v interface{}) string {
	data, _ := json.Marshal(v)
	return string(data)
}
