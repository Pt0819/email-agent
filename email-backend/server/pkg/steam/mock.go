// Package steam Mock数据
package steam

import "time"

// MockData Steam Mock数据
type MockData struct {
	UserGames     []GameInfo
	RecentGames   []RecentGameInfo
	PlayerSummary *PlayerSummary
	StoreDetails  map[int]*StoreGameDetails
}

// NewMockData 创建Mock数据
func NewMockData() *MockData {
	now := time.Now()

	md := &MockData{
		PlayerSummary: &PlayerSummary{
			SteamID:      "76561198012345678",
			PersonaName:  "TestGamer",
			ProfileURL:   "https://steamcommunity.com/id/testgamer/",
			Avatar:       "https://avatars.steamstatic.com/default.jpg",
			AvatarMedium: "https://avatars.steamstatic.com/default_medium.jpg",
			AvatarFull:   "https://avatars.steamstatic.com/default_full.jpg",
			PersonaState: 1,
			RealName:     "测试玩家",
			Location:     "CN",
		},
		UserGames: []GameInfo{
			{AppID: 1245620, Name: "ELDEN RING", PlaytimeForever: 18640, Playtime2Weeks: 320, ImgIconURL: "f4c0a6b1c0e3d8e9"},
			{AppID: 1091500, Name: "Cyberpunk 2077", PlaytimeForever: 12480, Playtime2Weeks: 180, ImgIconURL: "a1b2c3d4e5f6g7h8"},
			{AppID: 1174180, Name: "Red Dead Redemption 2", PlaytimeForever: 15360, Playtime2Weeks: 0, ImgIconURL: "b2c3d4e5f6g7h8i9"},
			{AppID: 892970, Name: "Valheim", PlaytimeForever: 5760, Playtime2Weeks: 0, ImgIconURL: "c3d4e5f6g7h8i9j0"},
			{AppID: 359550, Name: "Rainbow Six Siege", PlaytimeForever: 24000, Playtime2Weeks: 60, ImgIconURL: "d4e5f6g7h8i9j0k1"},
			{AppID: 570, Name: "Dota 2", PlaytimeForever: 36000, Playtime2Weeks: 0, ImgIconURL: "e5f6g7h8i9j0k1l2"},
			{AppID: 413150, Name: "Stardew Valley", PlaytimeForever: 4320, Playtime2Weeks: 240, ImgIconURL: "f6g7h8i9j0k1l2m3"},
			{AppID: 1086940, Name: "Baldur's Gate 3", PlaytimeForever: 19200, Playtime2Weeks: 480, ImgIconURL: "g7h8i9j0k1l2m3n4"},
			{AppID: 814380, Name: "Sekiro", PlaytimeForever: 6720, Playtime2Weeks: 0, ImgIconURL: "h8i9j0k1l2m3n4o5"},
			{AppID: 1593500, Name: "God of War", PlaytimeForever: 2880, Playtime2Weeks: 0, ImgIconURL: "i9j0k1l2m3n4o5p6"},
			{AppID: 374320, Name: "Dark Souls III", PlaytimeForever: 9600, Playtime2Weeks: 0, ImgIconURL: "j0k1l2m3n4o5p6q7"},
			{AppID: 220, Name: "Half-Life 2", PlaytimeForever: 960, Playtime2Weeks: 0, ImgIconURL: "k1l2m3n4o5p6q7r8"},
			{AppID: 105600, Name: "Terraria", PlaytimeForever: 3840, Playtime2Weeks: 0, ImgIconURL: "l2m3n4o5p6q7r8s9"},
			{AppID: 431960, Name: "Wallpaper Engine", PlaytimeForever: 12000, Playtime2Weeks: 0, ImgIconURL: "m3n4o5p6q7r8s9t0"},
			{AppID: 1172470, Name: "Hades", PlaytimeForever: 3360, Playtime2Weeks: 60, ImgIconURL: "n4o5p6q7r8s9t0u1"},
			{AppID: 252490, Name: "Rust", PlaytimeForever: 7200, Playtime2Weeks: 120, ImgIconURL: "o5p6q7r8s9t0u1v2"},
			{AppID: 271590, Name: "Grand Theft Auto V", PlaytimeForever: 9600, Playtime2Weeks: 0, ImgIconURL: "p6q7r8s9t0u1v2w3"},
			{AppID: 1551360, Name: "Forza Horizon 5", PlaytimeForever: 4800, Playtime2Weeks: 0, ImgIconURL: "q7r8s9t0u1v2w3x4"},
		},
		RecentGames: []RecentGameInfo{
			{AppID: 1086940, Name: "Baldur's Gate 3", PlaytimeForever: 19200, Playtime2Weeks: 480, ImgIconURL: "g7h8i9j0k1l2m3n4", LastPlayed: now.Unix()},
			{AppID: 1245620, Name: "ELDEN RING", PlaytimeForever: 18640, Playtime2Weeks: 320, ImgIconURL: "f4c0a6b1c0e3d8e9", LastPlayed: now.Add(-24 * time.Hour).Unix()},
			{AppID: 413150, Name: "Stardew Valley", PlaytimeForever: 4320, Playtime2Weeks: 240, ImgIconURL: "f6g7h8i9j0k1l2m3", LastPlayed: now.Add(-48 * time.Hour).Unix()},
			{AppID: 252490, Name: "Rust", PlaytimeForever: 7200, Playtime2Weeks: 120, ImgIconURL: "o5p6q7r8s9t0u1v2", LastPlayed: now.Add(-72 * time.Hour).Unix()},
			{AppID: 1172470, Name: "Hades", PlaytimeForever: 3360, Playtime2Weeks: 60, ImgIconURL: "n4o5p6q7r8s9t0u1", LastPlayed: now.Add(-96 * time.Hour).Unix()},
		},
	}

	return md
}

// GetMockGameDetails 获取Mock游戏详情
func (md *MockData) GetMockGameDetails(appID int) *StoreGameDetails {
	return &StoreGameDetails{
		Success: true,
	}
}

// GetGameGenreMock 获取游戏类型的Mock映射
func GetGameGenreMock(appID int) string {
	genres := map[int]string{
		1245620: "动作角色扮演",
		1091500: "动作冒险",
		1174180: "动作冒险",
		892970:  "生存沙盒",
		359550:  "战术射击",
		570:     "MOBA",
		413150:  "模拟经营",
		1086940: "角色扮演",
		814380:  "动作冒险",
		1593500: "动作冒险",
		374320:  "动作角色扮演",
		220:     "射击",
		105600:  "沙盒冒险",
		431960:  "工具",
		1172470: "Roguelike",
	}
	if g, ok := genres[appID]; ok {
		return g
	}
	return "其他"
}

// GetGameTagsMock 获取游戏标签的Mock映射
func GetGameTagsMock(appID int) string {
	tags := map[int]string{
		1245620: `["开放世界","魂系","暗黑奇幻","RPG","多人"]`,
		1091500: `["开放世界","赛博朋克","RPG","FPS","剧情"]`,
		1174180: `["开放世界","西部","剧情","动作","射击"]`,
		892970:  `["生存","多人","开放世界","建造","维京"]`,
		359550:  `["战术","射击","多人","竞技","FPS"]`,
		570:     `["MOBA","多人","竞技","策略","免费"]`,
		413150:  `["农场","模拟","像素","经营","多人"]`,
		1086940: `["RPG","回合制","剧情","奇幻","D&D"]`,
		814380:  `["动作","魂系","忍者","困难","单机"]`,
		1593500: `["动作","剧情","北欧","单机","冒险"]`,
		374320:  `["动作","魂系","RPG","困难","暗黑"]`,
		220:     `["FPS","剧情","经典","科幻","单机"]`,
		105600:  `["沙盒","建造","冒险","像素","2D"]`,
		431960:  `["工具","创意","自定义","桌面","壁纸"]`,
		1172470: `["Roguelike","动作","希腊神话","独立","爽快"]`,
	}
	if t, ok := tags[appID]; ok {
		return t
	}
	return `["其他"]`
}
