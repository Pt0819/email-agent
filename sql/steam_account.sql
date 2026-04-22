-- Steam账号绑定和游戏库表
-- 数据库: email_agent

USE email_agent;

-- Steam账号绑定表
CREATE TABLE IF NOT EXISTS steam_accounts (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    user_id BIGINT NOT NULL,
    steam_id VARCHAR(64) NOT NULL COMMENT 'SteamID64',
    steam_nickname VARCHAR(255) COMMENT 'Steam昵称',
    avatar_url VARCHAR(512) COMMENT '头像URL',
    profile_url VARCHAR(512) COMMENT '个人主页',
    real_name VARCHAR(100) COMMENT '真实姓名（公开信息）',
    location VARCHAR(50) COMMENT '国家/地区代码',
    api_key VARCHAR(255) COMMENT '可选的私有API Key',
    last_sync_at DATETIME COMMENT '最后同步时间',
    is_active BOOLEAN DEFAULT TRUE COMMENT '是否启用',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    UNIQUE INDEX idx_user_steam (user_id, steam_id),
    INDEX idx_steam_id (steam_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 用户Steam游戏库表（从API同步）
CREATE TABLE IF NOT EXISTS steam_library_items (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    user_id BIGINT NOT NULL,
    account_id BIGINT NOT NULL COMMENT '关联的steam_accounts.id',
    game_id VARCHAR(50) NOT NULL COMMENT 'Steam AppID',
    game_name VARCHAR(255) NOT NULL,
    playtime INT DEFAULT 0 COMMENT '总游玩时长(分钟)',
    playtime_2_weeks INT DEFAULT 0 COMMENT '最近两周(分钟)',
    last_played_at DATETIME COMMENT '最后游玩时间',
    icon_url VARCHAR(512) COMMENT '图标URL',
    is_synced BOOLEAN DEFAULT FALSE COMMENT '是否已同步元数据',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (account_id) REFERENCES steam_accounts(id) ON DELETE CASCADE,
    UNIQUE INDEX idx_user_game (user_id, account_id, game_id),
    INDEX idx_account (account_id),
    INDEX idx_playtime (playtime DESC)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
