-- Steam游戏推荐系统 - 用户偏好相关表
-- 数据库: email_agent

USE email_agent;

-- 用户游戏偏好标签表
-- 记录用户喜欢的游戏类型、标签及权重
CREATE TABLE IF NOT EXISTS user_game_preferences (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    user_id BIGINT NOT NULL,
    tag VARCHAR(100) NOT NULL COMMENT '偏好标签: RPG/动作/策略/射击 等',
    weight DECIMAL(5,2) DEFAULT 1.00 COMMENT '偏好权重 0-10, 越高越喜欢',
    source VARCHAR(30) DEFAULT 'system' COMMENT '来源: wishlist/email_purchase/manual/system',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    UNIQUE INDEX idx_user_tag (user_id, tag),
    INDEX idx_weight (weight DESC)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 推荐反馈表
-- 记录用户对推荐游戏的反馈，用于推荐算法优化
CREATE TABLE IF NOT EXISTS recommendation_feedback (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    user_id BIGINT NOT NULL,
    game_id VARCHAR(50) NOT NULL COMMENT 'Steam AppID',
    game_name VARCHAR(255) NOT NULL,
    action VARCHAR(30) NOT NULL COMMENT 'clicked/purchased/ignored/wishlisted',
    deal_id BIGINT COMMENT '关联的促销记录ID',
    email_id BIGINT COMMENT '来源邮件ID',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    INDEX idx_user_game (user_id, game_id),
    INDEX idx_action (action),
    INDEX idx_created (created_at DESC)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
