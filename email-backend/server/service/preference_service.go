// Package service 业务逻辑层 - 用户偏好分析
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

// PreferenceService 用户偏好分析服务
type PreferenceService struct {
	preferenceRepo *repository.PreferenceRepository
	insightRepo    *repository.PreferenceInsightRepository
	steamRepo      *repository.SteamRepository
	agentClient    *agent.Client
}

// NewPreferenceService 创建偏好分析服务
func NewPreferenceService(
	preferenceRepo *repository.PreferenceRepository,
	insightRepo *repository.PreferenceInsightRepository,
	steamRepo *repository.SteamRepository,
	agentClient *agent.Client,
) *PreferenceService {
	return &PreferenceService{
		preferenceRepo: preferenceRepo,
		insightRepo:    insightRepo,
		steamRepo:      steamRepo,
		agentClient:    agentClient,
	}
}

// GetUserProfile 构建用户画像聚合视图
func (s *PreferenceService) GetUserProfile(ctx context.Context, userID int64) (*model.UserGamingProfile, error) {
	profile := &model.UserGamingProfile{
		UserID: userID,
	}

	// 1. 获取 TopTags（按 weight 排序，取前15个）
	prefs, err := s.preferenceRepo.GetPreferencesByUserID(ctx, userID, 15)
	if err != nil {
		return nil, fmt.Errorf("获取用户偏好失败: %w", err)
	}
	profile.TopTags = s.buildTagPreferences(prefs)

	// 2. 获取 TopGenres（从 steam_games 表按 genre 分组计数）
	topGenres, err := s.getTopGenres(ctx, userID)
	if err != nil {
		// 不阻断整个流程，返回空数组
		profile.TopGenres = []model.TagPreference{}
	} else {
		profile.TopGenres = topGenres
	}

	// 3. 统计 TotalGames 和 TotalPlaytime（从 steam_library_items）
	s.buildLibraryStats(ctx, userID, profile)

	// 4. 计算 RecentActivitySummary（从 library_items 中最近2周的游玩）
	s.buildRecentActivity(ctx, userID, profile)

	// 5. 获取 LastAnalyzedAt（从 preference_insights 表获取最新一条记录）
	s.buildLastAnalyzedAt(ctx, userID, profile)

	return profile, nil
}

// AnalyzePreferences 触发偏好分析
func (s *PreferenceService) AnalyzePreferences(ctx context.Context, userID int64) (*model.PreferenceAnalysisResult, error) {
	result := &model.PreferenceAnalysisResult{
		Success: false,
	}

	// 1. 获取用户的游戏库数据
	gameLibrary, err := s.buildLibraryGameData(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("获取游戏库数据失败: %w", err)
	}

	// 2. 获取当前偏好标签
	prefs, err := s.preferenceRepo.GetPreferencesByUserID(ctx, userID, 0)
	if err != nil {
		return nil, fmt.Errorf("获取当前偏好失败: %w", err)
	}
	currentPrefs := make([]agent.PreferenceTagData, 0, len(prefs))
	for _, p := range prefs {
		currentPrefs = append(currentPrefs, agent.PreferenceTagData{
			Tag:    p.Tag,
			Weight: p.Weight,
			Source: p.Source,
		})
	}

	// 3. 构建请求并调用 Agent
	req := &agent.PreferenceAnalyzeRequest{
		UserID:       userID,
		GameLibrary:  gameLibrary,
		CurrentPrefs: currentPrefs,
		TriggerType:  model.InsightEventManualTrigger,
	}

	// 设置超时上下文
	analyzeCtx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	resp, err := s.agentClient.PreferenceAnalyze(analyzeCtx, req)
	if err != nil {
		result.Error = fmt.Sprintf("Agent分析失败: %v", err)
		return result, fmt.Errorf("调用Agent偏好分析失败: %w", err)
	}

	// 4. 处理 Agent 响应：更新/创建标签权重
	if err := s.processAgentResponse(ctx, userID, resp); err != nil {
		return nil, fmt.Errorf("处理Agent响应失败: %w", err)
	}

	// 5. 构建返回结果
	result.Success = resp.Success
	result.Insights = resp.Insights

	// 获取更新后的画像
	updatedProfile, err := s.GetUserProfile(ctx, userID)
	if err != nil {
		// 不阻断，返回空画像
		updatedProfile = &model.UserGamingProfile{UserID: userID}
	}
	result.Profile = updatedProfile

	return result, nil
}

// GetInsights 获取洞察列表
func (s *PreferenceService) GetInsights(ctx context.Context, userID int64, page, pageSize int) (*model.InsightListResponse, error) {
	insights, total, err := s.insightRepo.GetByUserID(ctx, userID, page, pageSize)
	if err != nil {
		return nil, fmt.Errorf("获取洞察列表失败: %w", err)
	}

	list := make([]model.InsightResponse, 0, len(insights))
	for _, insight := range insights {
		item := model.InsightResponse{
			ID:           insight.ID,
			EventType:    insight.EventType,
			DecisionType: insight.DecisionType,
			TriggerDesc:  insight.TriggerDesc,
			Insight:      insight.Insight,
			Reasoning:    insight.Reasoning,
			Confidence:   insight.Confidence,
			IsAnomaly:    insight.IsAnomaly,
			AnomalyType:  insight.AnomalyType,
			GameID:       insight.GameID,
			GameName:     insight.GameName,
			TagsChanged:  parseTagsChanged(insight.TagsChanged),
			CreatedAt:    insight.CreatedAt.Format("2006-01-02 15:04:05"),
		}
		list = append(list, item)
	}

	return &model.InsightListResponse{
		List:  list,
		Total: total,
		Page:  page,
	}, nil
}

// GetRecentAnomalies 获取最近异常
func (s *PreferenceService) GetRecentAnomalies(ctx context.Context, userID int64, limit int) ([]*model.PreferenceInsight, error) {
	if limit <= 0 {
		limit = 10
	}
	return s.insightRepo.GetRecentAnomalies(ctx, userID, limit)
}

// ==================== 内部辅助方法 ====================

// buildTagPreferences 从DB模型构建展示模型
func (s *PreferenceService) buildTagPreferences(prefs []*model.UserGamePreference) []model.TagPreference {
	result := make([]model.TagPreference, 0, len(prefs))
	for _, p := range prefs {
		result = append(result, model.TagPreference{
			Tag:    p.Tag,
			Weight: p.Weight,
			Source: p.Source,
		})
	}
	return result
}

// getTopGenres 从 steam_games 表按 genre 分组计数获取 Top 类型
func (s *PreferenceService) getTopGenres(ctx context.Context, userID int64) ([]model.TagPreference, error) {
	// 获取用户的全部游戏（取足够多的数据做分组统计）
	games, _, err := s.steamRepo.ListGames(ctx, userID, 1, 500, "")
	if err != nil {
		return nil, err
	}

	// 按 genre 分组计数
	genreCount := make(map[string]int)
	for _, game := range games {
		if game.Genre == "" {
			continue
		}
		genreCount[game.Genre]++
	}

	// 转换为 TagPreference 列表并排序
	result := make([]model.TagPreference, 0, len(genreCount))
	for genre, count := range genreCount {
		result = append(result, model.TagPreference{
			Tag:   genre,
			Count: count,
		})
	}

	// 按 count 降序排序，取前10
	sortTagPreferences(result)
	if len(result) > 10 {
		result = result[:10]
	}

	return result, nil
}

// buildLibraryStats 从游戏库统计总游戏数和总游玩时长
func (s *PreferenceService) buildLibraryStats(ctx context.Context, userID int64, profile *model.UserGamingProfile) {
	// 获取全部游戏库（不分页）
	items, total, err := s.steamRepo.ListLibraryByUser(ctx, userID, 1, 1000, "playtime")
	if err != nil {
		profile.TotalGames = 0
		profile.TotalPlaytime = 0
		return
	}

	profile.TotalGames = int(total)

	var totalPlaytime int
	for _, item := range items {
		totalPlaytime += item.Playtime
	}
	profile.TotalPlaytime = totalPlaytime
}

// buildRecentActivity 构建近期活动摘要
func (s *PreferenceService) buildRecentActivity(ctx context.Context, userID int64, profile *model.UserGamingProfile) {
	// 获取最近游玩（playtime_2_weeks > 0 的记录）
	recentItems, err := s.steamRepo.ListRecentPlayed(ctx, userID, 50)
	if err != nil || len(recentItems) == 0 {
		profile.RecentActivity = &model.RecentActivitySummary{
			GamesPlayedLastWeek:   0,
			TotalPlaytimeLastWeek: 0,
			GenreDistribution:     []model.GenreCount{},
		}
		return
	}

	summary := &model.RecentActivitySummary{
		GamesPlayedLastWeek: len(recentItems),
	}

	var totalPlaytime2Weeks int
	mostPlayedName := ""
	mostPlayedMinutes := 0

	// 用于统计类型分布
	genreMap := make(map[string]int)

	for _, item := range recentItems {
		totalPlaytime2Weeks += item.Playtime2Weeks
		if item.Playtime2Weeks > mostPlayedMinutes {
			mostPlayedMinutes = item.Playtime2Weeks
			mostPlayedName = item.GameName
		}

		// 尝试从 steam_games 获取 genre
		game, err := s.steamRepo.FindGameByGameID(ctx, userID, item.GameID)
		if err == nil && game.Genre != "" {
			genreMap[game.Genre]++
		}
	}

	summary.TotalPlaytimeLastWeek = totalPlaytime2Weeks
	summary.MostPlayedGame = mostPlayedName
	summary.MostPlayedGameHours = mostPlayedMinutes / 60

	// 类型分布
	genreDist := make([]model.GenreCount, 0, len(genreMap))
	for genre, count := range genreMap {
		genreDist = append(genreDist, model.GenreCount{
			Genre: genre,
			Count: count,
		})
	}
	summary.GenreDistribution = genreDist

	profile.RecentActivity = summary
}

// buildLastAnalyzedAt 获取上次分析时间
func (s *PreferenceService) buildLastAnalyzedAt(ctx context.Context, userID int64, profile *model.UserGamingProfile) {
	// 获取最新的洞察记录
	insights, _, err := s.insightRepo.GetByUserID(ctx, userID, 1, 1)
	if err != nil || len(insights) == 0 {
		return
	}
	t := insights[0].CreatedAt
	profile.LastAnalyzedAt = &t
}

// buildLibraryGameData 构建发送给 Agent 的游戏库数据
func (s *PreferenceService) buildLibraryGameData(ctx context.Context, userID int64) ([]agent.LibraryGameData, error) {
	items, _, err := s.steamRepo.ListLibraryByUser(ctx, userID, 1, 200, "playtime")
	if err != nil {
		return nil, err
	}

	result := make([]agent.LibraryGameData, 0, len(items))
	for _, item := range items {
		data := agent.LibraryGameData{
			GameID:        item.GameID,
			GameName:      item.GameName,
			Playtime:      item.Playtime,
			Playtime2Weeks: item.Playtime2Weeks,
		}

		// 获取游戏的 genre 和 tags
		game, err := s.steamRepo.FindGameByGameID(ctx, userID, item.GameID)
		if err == nil {
			data.Genre = game.Genre
			data.Tags = game.Tags // 已是 JSON 字符串
		}

		// 最后游玩时间
		if item.LastPlayedAt != nil {
			data.LastPlayedAt = item.LastPlayedAt.Format("2006-01-02 15:04:05")
		}

		result = append(result, data)
	}

	return result, nil
}

// processAgentResponse 处理 Agent 返回的分析结果
func (s *PreferenceService) processAgentResponse(ctx context.Context, userID int64, resp *agent.PreferenceAnalyzeResponse) error {
	var insightsToCreate []*model.PreferenceInsight

	// 1. 处理新增标签
	for _, tag := range resp.NewTags {
		pref := &model.UserGamePreference{
			UserID: userID,
			Tag:    tag.Tag,
			Weight: tag.Delta,
			Source: model.PreferenceSourceSystem,
		}
		if err := s.preferenceRepo.UpsertPreference(ctx, pref); err != nil {
			// 单个失败不阻断
			continue
		}
	}

	// 2. 处理更新标签
	for _, tag := range resp.UpdatedTags {
		pref := &model.UserGamePreference{
			UserID: userID,
			Tag:    tag.Tag,
			Weight: tag.Delta,
			Source: model.PreferenceSourceSystem,
		}
		if err := s.preferenceRepo.UpsertPreference(ctx, pref); err != nil {
			continue
		}
	}

	// 3. 构建标签变化 JSON（记录本次分析中的所有标签变化）
	allTagChanges := make([]map[string]interface{}, 0, len(resp.NewTags)+len(resp.UpdatedTags))
	for _, t := range resp.NewTags {
		allTagChanges = append(allTagChanges, map[string]interface{}{
			"tag":   t.Tag,
			"delta": t.Delta,
			"type":  "new",
		})
	}
	for _, t := range resp.UpdatedTags {
		allTagChanges = append(allTagChanges, map[string]interface{}{
			"tag":   t.Tag,
			"delta": t.Delta,
			"type":  "updated",
		})
	}
	tagsChangedJSON, _ := json.Marshal(allTagChanges)

	// 4. 构建操作记录 JSON
	actionsJSON, _ := json.Marshal(map[string]interface{}{
		"new_tags":     len(resp.NewTags),
		"updated_tags": len(resp.UpdatedTags),
		"insights":     len(resp.Insights),
		"anomalies":    len(resp.Anomalies),
	})

	// 5. 创建洞察记录
	reasoning := resp.Reasoning
	if len(resp.Insights) > 0 {
		// 将所有洞察合并
		insightText := ""
		for i, ins := range resp.Insights {
			if i > 0 {
				insightText += "\n"
			}
			insightText += ins
		}
		if reasoning != "" {
			reasoning = insightText
		} else {
			reasoning = insightText
		}
	}

	insight := &model.PreferenceInsight{
		UserID:       userID,
		EventType:    model.InsightEventManualTrigger,
		DecisionType: s.determineDecisionType(resp),
		TriggerDesc:  "用户手动触发偏好分析",
		Insight:      reasoning,
		Reasoning:    resp.Reasoning,
		Actions:      string(actionsJSON),
		Confidence:   0.85, // 默认置信度
		IsAnomaly:    len(resp.Anomalies) > 0,
		TagsChanged:  string(tagsChangedJSON),
	}
	insightsToCreate = append(insightsToCreate, insight)

	// 6. 为每个异常创建单独的洞察记录
	for _, anomaly := range resp.Anomalies {
		anomalyJSON, _ := json.Marshal(anomaly)
		anomalyInsight := &model.PreferenceInsight{
			UserID:       userID,
			EventType:    model.InsightEventManualTrigger,
			DecisionType: model.InsightDecisionAnomalyDetected,
			TriggerDesc:  anomaly.Description,
			Insight:      anomaly.Description,
			Reasoning:    resp.Reasoning,
			Actions:      string(anomalyJSON),
			Confidence:   0.90,
			IsAnomaly:    true,
			AnomalyType:  anomaly.Type,
			GameID:       anomaly.GameID,
			GameName:     anomaly.GameName,
		}
		insightsToCreate = append(insightsToCreate, anomalyInsight)
	}

	// 批量保存洞察
	if len(insightsToCreate) > 0 {
		if err := s.insightRepo.BatchCreate(ctx, insightsToCreate); err != nil {
			return fmt.Errorf("保存洞察记录失败: %w", err)
		}
	}

	return nil
}

// determineDecisionType 根据分析结果确定决策类型
func (s *PreferenceService) determineDecisionType(resp *agent.PreferenceAnalyzeResponse) string {
	if len(resp.Anomalies) > 0 {
		return model.InsightDecisionAnomalyDetected
	}
	if len(resp.NewTags) > 0 {
		return model.InsightDecisionNewPattern
	}
	if len(resp.UpdatedTags) > 0 {
		return model.InsightDecisionTagWeightAdjust
	}
	if resp.RecommendRec {
		return model.InsightDecisionGenerateRec
	}
	return model.InsightDecisionProfileUpdate
}

// parseTagsChanged 解析 TagsChanged JSON 字符串
func parseTagsChanged(tagsJSON string) []model.TagChange {
	if tagsJSON == "" {
		return []model.TagChange{}
	}

	var changes []model.TagChange
	if err := json.Unmarshal([]byte(tagsJSON), &changes); err != nil {
		// 尝试解析为通用 map 格式并转换
		var rawChanges []map[string]interface{}
		if err2 := json.Unmarshal([]byte(tagsJSON), &rawChanges); err2 != nil {
			return []model.TagChange{}
		}
		for _, rc := range rawChanges {
			delta := 0.0
			if d, ok := rc["delta"].(float64); ok {
				delta = d
			}
			tag := ""
			if t, ok := rc["tag"].(string); ok {
				tag = t
			}
			changes = append(changes, model.TagChange{
				Tag:   tag,
				Delta: delta,
			})
		}
	}
	return changes
}

// sortTagPreferences 按权重/计数降序排序
func sortTagPreferences(tags []model.TagPreference) {
	for i := 0; i < len(tags); i++ {
		for j := i + 1; j < len(tags); j++ {
			valI := tags[i].Weight
			if tags[i].Count > 0 {
				valI = float64(tags[i].Count)
			}
			valJ := tags[j].Weight
			if tags[j].Count > 0 {
				valJ = float64(tags[j].Count)
			}
			if valJ > valI {
				tags[i], tags[j] = tags[j], tags[i]
			}
		}
	}
}
