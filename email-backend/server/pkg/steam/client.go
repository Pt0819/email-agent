// Package steam Steam Web API客户端
package steam

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// Client Steam Web API客户端
type Client struct {
	apiKey     string
	httpClient *http.Client
	baseURL    string
	useMock    bool // 是否使用Mock数据
	mockData   *MockData
}

// Config 客户端配置
type Config struct {
	APIKey  string
	UseMock bool
}

// NewClient 创建Steam API客户端
func NewClient(cfg *Config) *Client {
	return &Client{
		apiKey: cfg.APIKey,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		baseURL:  "https://api.steampowered.com",
		useMock:  cfg.UseMock || cfg.APIKey == "",
		mockData: NewMockData(),
	}
}

// API响应结构

// GetOwnedGamesResponse 获取用户游戏库响应
type GetOwnedGamesResponse struct {
	Response struct {
		GameCount int        `json:"game_count"`
		Games     []GameInfo `json:"games"`
	} `json:"response"`
}

// GameInfo 游戏信息
type GameInfo struct {
	AppID        int    `json:"appid"`
	Name         string `json:"name"`
	PlaytimeForever int `json:"playtime_forever"`
	Playtime2Weeks int  `json:"playtime_2weeks,omitempty"`
	ImgIconURL   string `json:"img_icon_url"`
	HasCommunity bool   `json:"has_community_visible_assets"`
}

// GetRecentlyPlayedGamesResponse 获取最近游玩响应
type GetRecentlyPlayedGamesResponse struct {
	Response struct {
		TotalCount int              `json:"total_count"`
		Games      []RecentGameInfo `json:"games"`
	} `json:"response"`
}

// RecentGameInfo 最近游玩游戏
type RecentGameInfo struct {
	AppID         int    `json:"appid"`
	Name         string `json:"name"`
	PlaytimeForever int `json:"playtime_forever"`
	Playtime2Weeks int  `json:"playtime_2weeks"`
	ImgIconURL    string `json:"img_icon_url"`
	LastPlayed    int64  `json:"last_played"`
}

// GetPlayerSummariesResponse 获取玩家资料响应
type GetPlayerSummariesResponse struct {
	Response struct {
		Players []PlayerSummary `json:"players"`
	} `json:"response"`
}

// PlayerSummary 玩家资料
type PlayerSummary struct {
	SteamID      string `json:"steamid"`
	PersonaName  string `json:"personaname"`
	ProfileURL   string `json:"profileurl"`
	Avatar       string `json:"avatar"`
	AvatarMedium string `json:"avatarmedium"`
	AvatarFull   string `json:"avatarfull"`
	PersonaState int    `json:"personastate"`
	RealName     string `json:"realname,omitempty"`
	Location     string `json:"loccountrycode,omitempty"`
}

// GetOwnedGames 获取用户游戏库
func (c *Client) GetOwnedGames(ctx context.Context, steamID string) ([]GameInfo, error) {
	if c.useMock {
		return c.mockData.UserGames, nil
	}

	params := url.Values{
		"key":      {c.apiKey},
		"steamid":  {steamID},
		"format":   {"json"},
	}

	resp, err := c.doRequest(ctx, "IPlayerService/GetOwnedGames/v0002/", params)
	if err != nil {
		return nil, err
	}

	var result GetOwnedGamesResponse
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	return result.Response.Games, nil
}

// GetRecentlyPlayedGames 获取最近游玩（2周内）
func (c *Client) GetRecentlyPlayedGames(ctx context.Context, steamID string) ([]RecentGameInfo, error) {
	if c.useMock {
		return c.mockData.RecentGames, nil
	}

	params := url.Values{
		"key":     {c.apiKey},
		"steamid": {steamID},
		"format":  {"json"},
	}

	resp, err := c.doRequest(ctx, "IPlayerService/GetRecentlyPlayedGames/v0002/", params)
	if err != nil {
		return nil, err
	}

	var result GetRecentlyPlayedGamesResponse
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	return result.Response.Games, nil
}

// GetPlayerSummaries 获取玩家资料
func (c *Client) GetPlayerSummaries(ctx context.Context, steamID string) (*PlayerSummary, error) {
	if c.useMock {
		summary := *c.mockData.PlayerSummary
		summary.SteamID = steamID
		return &summary, nil
	}

	params := url.Values{
		"key":      {c.apiKey},
		"steamids": {steamID},
		"format":   {"json"},
	}

	resp, err := c.doRequest(ctx, "ISteamUser/GetPlayerSummaries/v0002/", params)
	if err != nil {
		return nil, err
	}

	var result GetPlayerSummariesResponse
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	if len(result.Response.Players) == 0 {
		return nil, fmt.Errorf("玩家不存在")
	}

	return &result.Response.Players[0], nil
}

// ResolveVanityURL 解析自定义URL
func (c *Client) ResolveVanityURL(ctx context.Context, vanityURL string) (string, error) {
	if c.useMock {
		// Mock模式下返回测试SteamID
		return "76561198012345678", nil
	}

	params := url.Values{
		"key":      {c.apiKey},
		"vanityurl": {vanityURL},
		"format":   {"json"},
	}

	resp, err := c.doRequest(ctx, "ISteamUser/ResolveVanityURL/v0002/", params)
	if err != nil {
		return "", err
	}

	var result struct {
		Response struct {
			Success    int    `json:"success"`
			SteamID    string `json:"steamid,omitempty"`
		} `json:"response"`
	}

	if err := json.Unmarshal(resp, &result); err != nil {
		return "", fmt.Errorf("解析响应失败: %w", err)
	}

	if result.Response.Success != 1 {
		return "", fmt.Errorf("无法解析自定义URL: %s", vanityURL)
	}

	return result.Response.SteamID, nil
}

// doRequest 发送API请求
func (c *Client) doRequest(ctx context.Context, endpoint string, params url.Values) ([]byte, error) {
	reqURL := c.baseURL + "/" + endpoint + "?" + params.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API返回错误: status=%d, body=%s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	return body, nil
}

// GetGameDetails 获取游戏详情（从Steam Store API）
func (c *Client) GetGameDetails(ctx context.Context, appID int) (*StoreGameDetails, error) {
	if c.useMock {
		return c.mockData.GetMockGameDetails(appID), nil
	}

	url := fmt.Sprintf("https://store.steampowered.com/api/appdetails?appids=%d&cc=cn&l=schinese", appID)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Store API返回错误: status=%d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result map[string]json.RawMessage
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	key := strconv.Itoa(appID)
	if data, ok := result[key]; ok {
		var details StoreGameDetails
		if err := json.Unmarshal(data, &details); err != nil {
			return nil, err
		}
		return &details, nil
	}

	return nil, fmt.Errorf("未找到游戏: %d", appID)
}

// StoreGameDetails Steam商店游戏详情
type StoreGameDetails struct {
	Success bool `json:"success"`
	Data    struct {
		Type        string   `json:"type"`
		Name        string   `json:"name"`
		SteamAppID  int      `json:"steam_appid"`
		ShortDesc   string   `json:"short_description"`
		HeaderImg   string   `json:"header_image"`
		Website     string   `json:"website"`
		Developers  []string `json:"developers"`
		Publishers  []string `json:"publishers"`
		Genres      []struct {
			ID        string `json:"id"`
			Desc      string `json:"description"`
		} `json:"genres"`
		Categories  []struct {
			ID        int    `json:"id"`
			Desc      string `json:"description"`
		} `json:"categories"`
		ReleaseDate struct {
			ComingSoon bool   `json:"coming_soon"`
			Date       string `json:"date"`
		} `json:"release_date"`
	} `json:"data"`
}

// SteamID验证
func IsValidSteamID(steamID string) bool {
	if steamID == "" {
		return false
	}

	// SteamID64格式：7656119xxxxxxxxxx
	if strings.HasPrefix(steamID, "7656119") && len(steamID) == 17 {
		_, err := strconv.ParseUint(steamID, 10, 64)
		return err == nil
	}

	// SteamID格式：STEAM_X:Y:Z
	if strings.HasPrefix(steamID, "STEAM_") {
		parts := strings.Split(steamID, ":")
		if len(parts) == 3 {
			return true
		}
	}

	return false
}
