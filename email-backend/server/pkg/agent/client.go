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
	ProcessedAt time.Time       `json:"processed_at"`
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
	ProcessedAt time.Time    `json:"processed_at"`
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