-- Steam游戏推荐系统 - 推荐结果存储表
-- 数据库: email_agent

USE email_agent;

-- 游戏推荐记录表
-- 存储个性化游戏推荐结果和用户反馈
CREATE TABLE IF NOT EXISTS game_recommendations (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    user_id BIGINT NOT NULL,
    game_id VARCHAR(50) NOT NULL COMMENT 'Steam AppID',
    game_name VARCHAR(255) NOT NULL COMMENT '游戏名称',
    game_genre VARCHAR(255) COMMENT '游戏类型',
    game_tags TEXT COMMENT '游戏标签JSON',
    cover_url VARCHAR(512) COMMENT '游戏封面URL',
    store_url VARCHAR(512) COMMENT '商店页面URL',
    match_score DECIMAL(5,2) DEFAULT 0 COMMENT '匹配度分数 0-100',
    match_reasons TEXT COMMENT '推荐理由(JSON数组)',
    deal_id BIGINT COMMENT '关联促销ID(如果有)',
    deal_price DECIMAL(10,2) COMMENT '促销价格',
    deal_discount INT COMMENT '折扣百分比',
    deal_end_date DATETIME COMMENT '促销截止日期',
    source VARCHAR(50) DEFAULT 'auto' COMMENT '推荐来源: auto/manual/surprise',
    status VARCHAR(30) DEFAULT 'active' COMMENT '状态: active/clicked/purchased/ignored/expired',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at DATETIME,
    INDEX idx_user_status (user_id, status),
    INDEX idx_user_created (user_id, created_at DESC),
    INDEX idx_game_id (game_id),
    INDEX idx_match_score (match_score DESC),
    INDEX idx_deal (deal_id),
    INDEX idx_status (status),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 推荐反馈追踪表（扩展原有recommendation_feedback）
-- 注意: recommendation_feedback表已存在，这里只添加补充字段
-- ALTER TABLE recommendation_feedback ADD COLUMN recommendation_id BIGINT AFTER game_id;
-- ALTER TABLE recommendation_feedback ADD COLUMN match_score DECIMAL(5,2) AFTER recommendation_id;

-- 推荐统计表（可选，用于追踪推荐效果）
CREATE TABLE IF NOT EXISTS recommendation_stats (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    user_id BIGINT NOT NULL UNIQUE,
    total_recommendations INT DEFAULT 0 COMMENT '累计推荐数',
    clicked_count INT DEFAULT 0 COMMENT '点击数',
    purchased_count INT DEFAULT 0 COMMENT '购买数',
    ignored_count INT DEFAULT 0 COMMENT '忽略数',
    ctr DECIMAL(5,2) DEFAULT 0 COMMENT '点击率',
    purchase_rate DECIMAL(5,2) DEFAULT 0 COMMENT '购买转化率',
    last_updated DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
