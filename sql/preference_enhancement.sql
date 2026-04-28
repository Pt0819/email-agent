-- Steam游戏推荐系统 - 偏好分析增强表
-- 数据库: email_agent

USE email_agent;

-- 偏好分析洞察记录表
-- 记录Agent在分析用户偏好时生成的洞察和决策记录
CREATE TABLE IF NOT EXISTS preference_insights (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    user_id BIGINT NOT NULL,
    event_type VARCHAR(50) NOT NULL COMMENT '触发事件类型: steam_email_sync/library_sync/playtime_update 等',
    decision_type VARCHAR(50) NOT NULL COMMENT '决策类型: profile_update/anomaly_detected/new_pattern 等',
    trigger_desc VARCHAR(255) COMMENT '触发描述',
    insight TEXT COMMENT 'Agent洞察内容',
    reasoning TEXT COMMENT '决策理由',
    actions TEXT COMMENT '执行的操作(JSON数组)',
    confidence DECIMAL(4,2) DEFAULT 0 COMMENT '决策置信度 0-1',
    is_anomaly BOOLEAN DEFAULT FALSE COMMENT '是否异常标记',
    anomaly_type VARCHAR(50) COMMENT '异常类型: extreme_playtime/new_genre_explored 等',
    game_id VARCHAR(50) COMMENT '关联游戏AppID',
    game_name VARCHAR(255) COMMENT '游戏名称',
    tags_changed TEXT COMMENT '标签变化(JSON数组)',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    deleted_at DATETIME,
    INDEX idx_user_created (user_id, created_at DESC),
    INDEX idx_event_type (event_type),
    INDEX idx_decision_type (decision_type),
    INDEX idx_is_anomaly (is_anomaly),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 用户游戏画像更新时间表（可选，用于追踪画像刷新频率）
CREATE TABLE IF NOT EXISTS preference_profile_sync (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    user_id BIGINT NOT NULL UNIQUE,
    last_sync_at DATETIME COMMENT '上次完整画像更新时间',
    sync_source VARCHAR(50) COMMENT '触发来源: manual/auto/library_sync',
    games_analyzed INT DEFAULT 0 COMMENT '本次分析的游戏数',
    tags_extracted INT DEFAULT 0 COMMENT '提取的标签数',
    anomalies_detected INT DEFAULT 0 COMMENT '检测到的异常数',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
