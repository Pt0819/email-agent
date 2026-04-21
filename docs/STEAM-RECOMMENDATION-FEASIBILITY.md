# Steam游戏推荐系统 - 可行性分析

> 日期：2026-04-21
> 版本：v1.0
> 状态：可行性分析

---

## 1. 需求概述

### 1.1 核心功能

1. **Steam邮件智能解析** (一期)
   - 识别Steam相关邮件（促销、愿望单特卖、新游戏上架等）
   - 提取游戏信息（名称、价格、折扣、标签、类型）
   - 存储结构化游戏数据

2. **Steam用户偏好分析** (二期)
   - 获取用户Steam 30天游玩记录
   - 通过LLM分析游戏偏好类型
   - 构建用户兴趣画像

3. **智能游戏推荐** (三期)
   - 匹配邮件资讯与用户偏好
   - 个性化推荐理由生成
   - 推荐排序和过滤

---

## 2. 技术可行性分析

### 2.1 Steam邮件解析 ✅ 高可行性

**现有基础**：
- 126邮箱Provider已实现
- 邮件分类Agent已实现
- 信息提取Agent已实现

**需要增强**：
```
Steam邮件类型识别:
├── 促销邮件 (Special Offers)
│   └── 提取: 游戏名、原价、折扣价、折扣率、截止日期
├── 愿望单通知 (Wishlist)
│   └── 提取: 游戏名、降价信息、历史最低价
├── 新游戏上架 (New Release)
│   └── 提取: 游戏名、发行商、标签、类型
└── 更新通知 (Update)
    └── 提取: 游戏名、更新内容、版本
```

**技术方案**：
1. 增加分类类别：`steam_promotion`, `steam_wishlist`, `steam_news`
2. Steam信息提取Agent（使用LLM解析HTML邮件）
3. 游戏信息结构化存储

**可行度**: ⭐⭐⭐⭐⭐ (95%)
- 邮件获取已解决
- LLM解析HTML邮件能力已验证
- 无外部API依赖，完全可控

---

### 2.2 Steam用户数据获取 ⚠️ 中等可行性

**Steam Web API能力**：

| API | 端点 | 功能 | 限制 |
|-----|------|------|------|
| GetOwnedGames | `IPlayerService/GetOwnedGames` | 获取用户拥有的游戏 | 需要API Key |
| GetRecentlyPlayedGames | `IPlayerService/GetRecentlyPlayedGames` | 获取最近游玩游戏 | 需要API Key |
| GetPlayerSummaries | `ISteamUser/GetPlayerSummaries` | 获取用户资料 | 需要API Key |
| GetSchemaForGame | `ISteamUserStats/GetSchemaForGame` | 获取游戏成就统计 | 需要API Key |

**关键发现**：
- ✅ Steam提供官方Web API
- ✅ 可以获取30天游玩记录
- ⚠️ **不能获取"游玩时长"的详细历史**，只能获取总时长
- ⚠️ 需要用户提供Steam ID和API Key

**获取方式**：
```bash
# 获取最近游玩的游戏
GET https://api.steampowered.com/IPlayerService/GetRecentlyPlayedGames/v1/
    ?key=XXX
    &steamid=7656119XXXXX
    &count=10

# 响应示例
{
  "response": {
    "total_count": 1,
    "games": [
      {
        "appid": 730,
        "name": "Counter-Strike 2",
        "playtime_2weeks": 1200,  # 最近2周游玩时长（分钟）
        "playtime_forever": 50000, # 总游玩时长（分钟）
        "img_icon_url": "xxx",
        "img_logo_url": "xxx"
      }
    ]
  }
}
```

**技术挑战**：
1. **Steam Web API限制**
   - 请求频率限制：每分钟最多300次（足够个人使用）
   - 需要申请API Key：https://steamcommunity.com/dev/apikey

2. **数据完整性问题**
   - 无法获取游戏类型标签（需要调用Store API）
   - 需要额外请求获取游戏详细信息

**可行度**: ⭐⭐⭐⭐ (80%)
- Steam API文档完善
- 社区有成熟的SDK (go-steam, steam.py)
- 主要挑战是游戏元数据获取

---

### 2.3 游戏偏好分析 ✅ 高可行性

**数据来源**：
```
用户游玩数据:
├── 游戏名称
├── 游玩时长
├── 最近游玩时间
└── 游戏ID (AppID)

游戏元数据 (Steam Store):
├── 游戏标签 (Tags: 动作、RPG、独立游戏...)
├── 游戏类型 (Genre: 冒险、策略...)
├── 开发商/发行商
├── 发行日期
└── 用户评价 (Positive/Negative)
```

**分析方案**：

**方案A: 规则-based分析** (简单快速)
```python
def analyze_preference(playtime_games):
    # 统计标签出现频率
    tag_scores = {}
    for game in playtime_games:
        tags = get_steam_tags(game.appid)
        for tag in tags:
            tag_scores[tag] = tag_scores.get(tag, 0) + game.playtime

    # 按游玩时长加权排序
    return sorted(tag_scores.items(), key=lambda x: x[1], reverse=True)

# 输出示例
{
    "开放世界": 1200,  # 高权重
    "RPG": 800,
    "动作": 600,
    "多人": 100        # 低权重
}
```

**方案B: LLM分析** (智能深入)
```python
prompt = f"""
根据以下Steam游玩记录，分析玩家的游戏偏好：

游玩记录：
{format_games(playtime_data)}

请分析：
1. 最喜欢的游戏类型（Top 3）
2. 偏好的游戏风格（开放世界/线性、单人/多人等）
3. 特殊偏好（独立游戏/3A大作、像素风/写实风等）

返回JSON格式。
"""

# LLM输出
{
    "top_genres": ["开放世界RPG", "动作冒险", "策略模拟"],
    "play_style": "偏爱单人深度体验，喜欢开放世界探索",
    "special_interests": ["注重剧情质量", "喜欢育碧式开放世界", "偏好科幻题材"]
}
```

**可行度**: ⭐⭐⭐⭐⭐ (90%)
- 方案A：技术成熟，实现简单
- 方案B：已有LLM基础设施

---

### 2.4 个性化推荐 ✅ 高可行性

**推荐逻辑**：
```python
def recommend_games(user_profile, steam_emails):
    """
    user_profile: {
        "top_genres": ["RPG", "开放世界"],
        "special_interests": ["科幻题材", "深度剧情"]
    }

    steam_emails: [
        {
            "game_name": "博德之门3",
            "tags": ["RPG", "开放世界", "奇幻"],
            "discount": "-50%",
            "price": "¥149"
        },
        ...
    ]
    """

    recommendations = []
    for email in steam_emails:
        # 计算匹配度
        score = calculate_match_score(user_profile, email)

        # 使用LLM生成推荐理由
        if score > threshold:
            reason = llm_generate_reason(user_profile, email)
            recommendations.append({
                "game": email["game_name"],
                "match_score": score,
                "reason": reason,
                "price_info": email
            })

    return sort_by_score(recommendations)
```

**推荐理由生成示例**：
```
输入：
- 用户偏好：开放世界RPG、科幻题材
- 游戏信息：Starfield（星空），开放世界太空RPG

LLM生成推荐理由：
"根据你最近大量游玩《荒野大镖客2》和《赛博朋克2077》的记录，
我注意到你喜欢沉浸式开放世界RPG。Steam正在促销的《Starfield》
完美匹配你的偏好——它结合了贝塞斯达的开放世界探索和深度科幻叙事，
首发打折-25%，非常值得考虑！"
```

**可行度**: ⭐⭐⭐⭐⭐ (95%)
- 推荐算法成熟
- LLM生成个性化理由效果好
- 可以引入评分/反馈优化推荐

---

## 3. 数据流设计

```
┌─────────────────────────────────────────────────────────────────┐
│                      Steam游戏推荐系统                           │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  ┌────────────────┐    ┌────────────────┐    ┌──────────────┐  │
│  │ Steam邮件      │───▶│ 邮件分类Agent  │───▶│ Steam邮件   │  │
│  │ (IMAP同步)     │    │ (识别steam_*)  │    │ 专属队列    │  │
│  └────────────────┘    └────────────────┘    └──────────────┘  │
│                                                        │         │
│                                                        ▼         │
│  ┌────────────────┐    ┌────────────────┐    ┌──────────────┐  │
│  │ Steam Web API  │───▶│ 偏好分析Agent  │───▶│ 用户画像    │  │
│  │ (游玩记录)     │    │ (LLM分析)      │    │ (Redis缓存) │  │
│  └────────────────┘    └────────────────┘    └──────────────┘  │
│                                                        │         │
│                                                        ▼         │
│  ┌────────────────┐    ┌────────────────┐    ┌──────────────┐  │
│  │ Steam Store    │───▶│ 游戏推荐Agent  │◀───│ 用户画像    │  │
│  │ (游戏元数据)   │    │ (匹配+排序)    │    │              │  │
│  └────────────────┘    └────────────────┘    └──────────────┘  │
│                                 │                               │
│                                 ▼                               │
│  ┌────────────────┐    ┌────────────────┐    ┌──────────────┐  │
│  │ Web前端        │◀───│ 推荐结果API    │    │ 个性化理由  │  │
│  │ (推荐展示)     │    │ (JSON响应)     │    │ (LLM生成)   │  │
│  └────────────────┘    └────────────────┘    └──────────────┘  │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

---

## 4. 实现路线图

### Phase 1: Steam邮件解析 (2周)

**目标**: 识别并提取Steam邮件中的游戏信息

**任务**:
1. 扩展邮件分类类别
   - `steam_promotion` (促销)
   - `steam_wishlist` (愿望单)
   - `steam_news` (资讯)

2. Steam信息提取Agent
   - 解析HTML邮件结构
   - 提取游戏名称、价格、折扣
   - 识别游戏标签

3. 游戏信息存储
   - `steam_games`表（游戏基础信息）
   - `steam_deals`表（促销信息）
   - `email_game_tags`表（邮件-游戏关联）

**验收标准**:
- 收到Steam促销邮件后，自动提取游戏信息
- 数据准确率 > 90%

---

### Phase 2: Steam数据集成 (1周)

**目标**: 获取用户Steam游玩记录

**任务**:
1. Steam账号绑定
   - 用户输入Steam ID
   - 系统获取API Key（或使用统一Key）

2. Steam数据同步
   - 定期同步用户游戏库
   - 获取最近游玩记录（30天）
   - 同步游戏元数据（标签、类型）

3. 数据展示
   - 前端显示Steam账号绑定状态
   - 游戏库列表
   - 游玩时长统计

**验收标准**:
- 可以绑定Steam账号
- 显示最近游玩的游戏

---

### Phase 3: 偏好分析 (1周)

**目标**: 分析用户游戏偏好

**任务**:
1. 规则-based分析（MVP）
   - 标签统计
   - 时长加权
   - 生成用户画像

2. LLM增强分析（可选）
   - 深度偏好解读
   - 游戏风格分析
   - 特殊兴趣识别

3. 画像展示
   - 前端可视化用户偏好
   - Top标签、类型、风格

**验收标准**:
- 生成用户偏好画像
- 偏好准确率 > 80%（用户反馈）

---

### Phase 4: 智能推荐 (2周)

**目标**: 根据偏好推荐游戏

**任务**:
1. 推荐算法
   - 标签匹配度计算
   - 个性化排序
   - 过滤已拥有游戏

2. 推荐理由生成
   - LLM生成个性化理由
   - 结合用户游玩历史
   - 强调匹配点

3. 推荐展示
   - 推荐列表页面
   - 推荐理由卡片
   - 相关度评分

4. 反馈优化
   - 用户点赞/点踩
   - 推荐效果追踪

**验收标准**:
- 生成个性化推荐列表
- 推荐理由自然流畅
- 用户满意度 > 70%

---

## 5. 数据库设计

### 5.1 新增表结构

```sql
-- Steam账号绑定
CREATE TABLE steam_accounts (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    user_id BIGINT NOT NULL,
    steam_id VARCHAR(20) UNIQUE NOT NULL,  -- Steam 64-bit ID
    steam_username VARCHAR(100),
    avatar_url VARCHAR(255),
    last_sync_at DATETIME,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id)
);

-- Steam游戏库
CREATE TABLE steam_games (
    app_id INT PRIMARY KEY,  -- Steam AppID
    name VARCHAR(255) NOT NULL,
    developers JSON,         -- 开发商列表
    publishers JSON,         -- 发行商列表
    tags JSON,               -- 游戏标签 ["动作", "RPG", "开放世界"]
    genres JSON,             -- 游戏类型 ["Action", "RPG"]
    release_date DATE,
    header_image VARCHAR(255),
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_tags ((CAST(tags AS CHAR(255)))),
    INDEX idx_genres ((CAST(genres AS CHAR(255))))
);

-- 用户游戏记录
CREATE TABLE user_game_library (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    user_id BIGINT NOT NULL,
    app_id INT NOT NULL,
    playtime_forever INT,     -- 总游玩时长（分钟）
    playtime_2weeks INT,      -- 最近2周游玩时长
    last_played_at DATETIME,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id),
    FOREIGN KEY (app_id) REFERENCES steam_games(app_id),
    UNIQUE KEY (user_id, app_id)
);

-- Steam促销信息
CREATE TABLE steam_deals (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    app_id INT NOT NULL,
    email_id BIGINT,          -- 来源邮件
    original_price DECIMAL(10,2),
    discount_price DECIMAL(10,2),
    discount_percent INT,
    start_date DATETIME,
    end_date DATETIME,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (app_id) REFERENCES steam_games(app_id),
    FOREIGN KEY (email_id) REFERENCES emails(id) ON DELETE SET NULL
);

-- 用户偏好画像
CREATE TABLE user_gaming_profile (
    user_id BIGINT PRIMARY KEY,
    favorite_tags JSON,       -- {"开放世界": 1200, "RPG": 800}
    favorite_genres JSON,
    play_style JSON,          -- LLM分析结果
    special_interests JSON,   -- LLM分析结果
    last_updated_at DATETIME,
    FOREIGN KEY (user_id) REFERENCES users(id)
);

-- 游戏推荐记录
CREATE TABLE game_recommendations (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    user_id BIGINT NOT NULL,
    app_id INT NOT NULL,
    deal_id BIGINT,           -- 关联促销（如果有）
    match_score DECIMAL(3,2), -- 匹配分数 0-1
    reason TEXT,              -- 推荐理由（LLM生成）
    feedback ENUM('like', 'dislike', 'ignore'),
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id),
    FOREIGN KEY (app_id) REFERENCES steam_games(app_id),
    FOREIGN KEY (deal_id) REFERENCES steam_deals(id) ON DELETE SET NULL,
    INDEX idx_user_score (user_id, match_score DESC)
);
```

---

## 6. API设计

### 6.1 Steam账号管理

```
POST   /api/v1/steam/bind         # 绑定Steam账号
GET    /api/v1/steam/profile      # 获取Steam资料
DELETE /api/v1/steam/unbind       # 解绑账号
```

### 6.2 游戏库管理

```
GET    /api/v1/steam/games        # 获取游戏库
GET    /api/v1/steam/games/recent # 获取最近游玩
POST   /api/v1/steam/sync         # 手动同步
```

### 6.3 偏好分析

```
GET    /api/v1/steam/profile/preference    # 获取偏好画像
POST   /api/v1/steam/profile/analyze       # 重新分析
```

### 6.4 游戏推荐

```
GET    /api/v1/recommendations             # 获取推荐列表
POST   /api/v1/recommendations/:id/feedback # 反馈（点赞/点踩）
GET    /api/v1/recommendations/deals        # 仅推荐促销
```

---

## 7. 风险与挑战

### 7.1 技术风险

| 风险 | 影响 | 概率 | 缓解措施 |
|------|------|------|----------|
| Steam API限流 | 高 | 低 | 本地缓存、定时批量同步 |
| Steam游戏元数据缺失 | 中 | 中 | 多源数据（SteamDB、IGDB）|
| HTML邮件解析失败 | 中 | 高 | 多模板适配、LLM容错 |
| 推荐准确率低 | 低 | 中 | 用户反馈循环优化 |

### 7.2 产品风险

| 风险 | 影响 | 概率 | 缓解措施 |
|------|------|------|----------|
| 用户没有Steam | 高 | 低 | 不影响原有功能 |
| Steam邮件太少 | 中 | 中 | 可关注官方Curator |
| 推荐同质化 | 低 | 中 | 多样化策略、惊喜推荐 |

---

## 8. 竞品分析

| 产品 | 优势 | 劣势 |
|------|------|------|
| **Steam官方推荐** | 数据准确、推荐精准 | 基于购买历史，不考虑促销 |
| **Deku Deals** | 促销追踪强大 | 无个性化推荐 |
| **IsThereAnyDeal** | 价格历史完善 | 无用户偏好分析 |
| **本方案** | 🆕 结合邮件促销 + 用户偏好 | 需要数据积累期 |

---

## 9. 总结

### 可行性评估：⭐⭐⭐⭐ (85/100)

**优势**:
- ✅ 技术栈完全支持
- ✅ Steam API文档完善
- ✅ LLM能力已验证
- ✅ 产品方向独特

**挑战**:
- ⚠️ Steam游戏元数据获取复杂
- ⚠️ 需要用户主动绑定Steam
- ⚠️ 邮件HTML解析需要适配

### 建议开发顺序

1. **Phase 1: Steam邮件解析** (优先级最高)
   - 独立于Steam账号，可快速验证价值
   - 为后续功能打数据基础

2. **Phase 2: Steam数据集成**
   - 验证API可行性
   - 收集真实用户偏好数据

3. **Phase 3: 偏好分析 + Phase 4: 智能推荐**
   - 完整体验
   - 持续优化

### 预计工期

- **MVP版本**: 4-6周
- **完整版本**: 8-10周

---

*文档版本: v1.0*
*分析日期: 2026-04-21*
