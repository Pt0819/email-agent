// Package agent Agent服务客户端
package agent

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"email-backend/server/config"
)

// Client Agent客户端
type Client struct {
	baseURL string
	apiKey  string
	timeout time.Duration
	client  *http.Client
}

// NewClient 创建Agent客户端
func NewClient(cfg *config.AgentConfig) *Client {
	return &Client{
		baseURL: cfg.URL,
		apiKey:  cfg.APIKey,
		timeout: time.Duration(cfg.Timeout) * time.Second,
		client:  &http.Client{Timeout: time.Duration(cfg.Timeout) * time.Second},
	}
}

// ClassifyRequest 分类请求
type ClassifyRequest struct {
	EmailID     string `json:"email_id"`
	Subject     string `json:"subject"`
	SenderName  string `json:"sender_name,omitempty"`
	SenderEmail string `json:"sender_email"`
	Content     string `json:"content"`
	ReceivedAt  string `json:"received_at,omitempty"`
}

// ClassifyResponse 分类响应
type ClassifyResponse struct {
	EmailID     string          `json:"email_id"`
	Classification Classification `json:"classification"`
	ProcessedAt string          `json:"processed_at"`
}

// Classification 分类结果
type Classification struct {
	Category   string  `json:"category"`
	Priority   string  `json:"priority"`
	Confidence float64 `json:"confidence"`
	Reasoning  string  `json:"reasoning"`
}

// ExtractRequest 提取请求
type ExtractRequest struct {
	EmailID     string `json:"email_id"`
	Subject     string `json:"subject"`
	SenderName  string `json:"sender_name,omitempty"`
	SenderEmail string `json:"sender_email"`
	Content     string `json:"content"`
}

// ExtractResponse 提取响应
type ExtractResponse struct {
	EmailID    string        `json:"email_id"`
	Extraction ExtractionResult `json:"extraction"`
	ProcessedAt string       `json:"processed_at"`
}

// ExtractionResult 提取结果
type ExtractionResult struct {
	ActionItems []ActionItem   `json:"action_items"`
	Meetings    []MeetingInfo  `json:"meetings"`
	Summary     string         `json:"summary"`
	Intent      string         `json:"intent"`
}

// ActionItem 行动项
type ActionItem struct {
	Task     string `json:"task"`
	TaskType string `json:"task_type"`
	Deadline string `json:"deadline,omitempty"`
	Priority string `json:"priority"`
}

// MeetingInfo 会议信息
type MeetingInfo struct {
	Title      string   `json:"title"`
	Time       string   `json:"time"`
	Location   string   `json:"location,omitempty"`
	Attendees  []string `json:"attendees"`
	MeetingURL string   `json:"meeting_url,omitempty"`
}

// HealthResponse 健康检查响应
type HealthResponse struct {
	Status     string   `json:"status"`
	Service    string   `json:"service"`
	Version    string   `json:"version"`
	LLMStatus  string   `json:"llm_status"`
	Providers  []string `json:"providers"`
}

// SummaryRequest 摘要请求
type SummaryRequest struct {
	EmailIDs   []string                 `json:"email_ids"`
	EmailsData []map[string]interface{} `json:"emails_data"`
	Date       string                   `json:"date"`
}

// SummaryResponse 摘要响应
type SummaryResponse struct {
	Date           string                 `json:"date"`
	TotalEmails    int                    `json:"total_emails"`
	ByCategory     map[string]int         `json:"by_category"`
	ImportantEmails []ImportantEmailAgent `json:"important_emails"`
	ActionItems    []ActionItemAgent      `json:"action_items"`
	SummaryText    string                 `json:"summary_text"`
}

// ImportantEmailAgent Agent返回的重要邮件
type ImportantEmailAgent struct {
	EmailID  string `json:"email_id"`
	Subject  string `json:"subject"`
	Sender   string `json:"sender"`
	Category string `json:"category"`
	Priority string `json:"priority"`
	Summary  string `json:"summary"`
}

// ActionItemAgent Agent返回的行动项
type ActionItemAgent struct {
	Task     string `json:"task"`
	Priority string `json:"priority"`
}

// Classify 调用Agent分类邮件
func (c *Client) Classify(ctx context.Context, req *ClassifyRequest) (*ClassifyResponse, error) {
	url := c.baseURL + "/api/v1/classify"

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("序列化请求失败: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	if c.apiKey != "" {
		httpReq.Header.Set("X-API-Key", c.apiKey)
	}

	resp, err := c.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("请求Agent失败: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Agent返回错误: status=%d, body=%s", resp.StatusCode, string(respBody))
	}

	var result ClassifyResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	return &result, nil
}

// Extract 调用Agent提取信息
func (c *Client) Extract(ctx context.Context, req *ExtractRequest) (*ExtractResponse, error) {
	url := c.baseURL + "/api/v1/extract"

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("序列化请求失败: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	if c.apiKey != "" {
		httpReq.Header.Set("X-API-Key", c.apiKey)
	}

	resp, err := c.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("请求Agent失败: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Agent返回错误: status=%d, body=%s", resp.StatusCode, string(respBody))
	}

	var result ExtractResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	return &result, nil
}

// SteamExtractRequest Steam信息提取请求
type SteamExtractRequest struct {
	EmailID     string `json:"email_id"`
	Subject     string `json:"subject"`
	SenderEmail string `json:"sender_email"`
	Content     string `json:"content"`
	ContentHTML string `json:"content_html,omitempty"`
}

// SteamGameInfo 提取的Steam游戏信息
type SteamGameInfo struct {
	AppID         string   `json:"app_id"`
	Name          string   `json:"name"`
	Genre         string   `json:"genre"`
	Tags          []string `json:"tags"`
	CoverURL      string   `json:"cover_url"`
	StoreURL      string   `json:"store_url"`
	HasDeal       bool     `json:"has_deal"`
	OriginalPrice float64  `json:"original_price"`
	DealPrice     float64  `json:"deal_price"`
	Discount      int      `json:"discount"`
	DealEnd       string   `json:"deal_end,omitempty"`
}

// SteamExtractResponse Steam提取响应
type SteamExtractResponse struct {
	EmailID     string          `json:"email_id"`
	Games       []SteamGameInfo `json:"games"`
	ProcessedAt string          `json:"processed_at"`
}

// SteamExtract 调用Agent提取Steam游戏信息
func (c *Client) SteamExtract(ctx context.Context, req *SteamExtractRequest) (*SteamExtractResponse, error) {
	url := c.baseURL + "/api/v1/steam/extract"

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("序列化请求失败: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	if c.apiKey != "" {
		httpReq.Header.Set("X-API-Key", c.apiKey)
	}

	resp, err := c.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("请求Agent失败: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Agent返回错误: status=%d, body=%s", resp.StatusCode, string(respBody))
	}

	var result SteamExtractResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	return &result, nil
}

// Health 检查Agent健康状态
func (c *Client) Health(ctx context.Context) (*HealthResponse, error) {
	url := c.baseURL + "/api/v1/health"

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	resp, err := c.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("请求Agent失败: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Agent返回错误: status=%d", resp.StatusCode)
	}

	var result HealthResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	return &result, nil
}

// Summary 调用Agent生成摘要
func (c *Client) Summary(ctx context.Context, req *SummaryRequest) (*SummaryResponse, error) {
	url := c.baseURL + "/api/v1/summary/daily"

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("序列化请求失败: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	if c.apiKey != "" {
		httpReq.Header.Set("X-API-Key", c.apiKey)
	}

	resp, err := c.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("请求Agent失败: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Agent返回错误: status=%d, body=%s", resp.StatusCode, string(respBody))
	}

	var result SummaryResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	return &result, nil
}

// ==================== 偏好分析 ====================

// PreferenceAnalyzeRequest 偏好分析请求
type PreferenceAnalyzeRequest struct {
	UserID       int64               `json:"user_id"`
	GameLibrary  []LibraryGameData   `json:"game_library"`
	CurrentPrefs []PreferenceTagData `json:"current_preferences"`
	TriggerType  string              `json:"trigger_type"`
}

// LibraryGameData 游戏库游戏数据
type LibraryGameData struct {
	GameID        string `json:"game_id"`
	GameName      string `json:"game_name"`
	Playtime      int    `json:"playtime"`
	Playtime2Weeks int   `json:"playtime_2_weeks"`
	Genre         string `json:"genre"`
	Tags          string `json:"tags"`
	LastPlayedAt  string `json:"last_played_at,omitempty"`
}

// PreferenceTagData 偏好标签数据
type PreferenceTagData struct {
	Tag    string  `json:"tag"`
	Weight float64 `json:"weight"`
	Source string  `json:"source"`
}

// PreferenceAnalyzeResponse 偏好分析响应
type PreferenceAnalyzeResponse struct {
	Success      bool             `json:"success"`
	NewTags      []TagChangeData  `json:"new_tags"`
	UpdatedTags  []TagChangeData  `json:"updated_tags"`
	Insights     []string        `json:"insights"`
	Reasoning    string           `json:"reasoning"`
	Anomalies    []AnomalyData    `json:"anomalies"`
	RecommendRec bool             `json:"recommend_rec"`
}

// TagChangeData 标签变化数据
type TagChangeData struct {
	Tag   string  `json:"tag"`
	Delta float64 `json:"delta"`
}

// AnomalyData 异常数据
type AnomalyData struct {
	Type        string `json:"type"`
	GameID      string `json:"game_id,omitempty"`
	GameName    string `json:"game_name,omitempty"`
	Description string `json:"description"`
}

// PreferenceAnalyze 调用Agent进行偏好分析
func (c *Client) PreferenceAnalyze(ctx context.Context, req *PreferenceAnalyzeRequest) (*PreferenceAnalyzeResponse, error) {
	url := c.baseURL + "/api/v1/preference/analyze"

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("序列化请求失败: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	if c.apiKey != "" {
		httpReq.Header.Set("X-API-Key", c.apiKey)
	}

	resp, err := c.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("请求Agent失败: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Agent返回错误: status=%d, body=%s", resp.StatusCode, string(respBody))
	}

	var result PreferenceAnalyzeResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	return &result, nil
}