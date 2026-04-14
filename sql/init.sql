-- ============================================
-- Email Agent 系统数据库初始化脚本
-- 数据库: email_system
-- 创建日期: 2026-04-09
-- ============================================

-- 创建数据库
CREATE DATABASE IF NOT EXISTS email_system
DEFAULT CHARACTER SET utf8mb4
COLLATE utf8mb4_unicode_ci;

USE email_system;

-- ============================================
-- 用户表
-- ============================================
CREATE TABLE IF NOT EXISTS users (
    id BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '用户ID',
    username VARCHAR(100) NOT NULL COMMENT '用户名',
    email VARCHAR(255) NOT NULL COMMENT '用户邮箱',
    password_hash VARCHAR(255) COMMENT '密码哈希',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    deleted_at DATETIME DEFAULT NULL COMMENT '软删除时间',

    UNIQUE KEY uk_username (username),
    UNIQUE KEY uk_email (email),
    INDEX idx_deleted_at (deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户表';

-- ============================================
-- 邮箱账户表
-- 存储 user 的邮箱账户信息和加密凭证
-- ============================================
CREATE TABLE IF NOT EXISTS email_accounts (
    id BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '账户ID',
    user_id BIGINT NOT NULL COMMENT '所属用户ID',
    provider VARCHAR(20) NOT NULL COMMENT '邮箱提供商: 126, gmail, outlook, imap',
    account_email VARCHAR(255) NOT NULL COMMENT '邮箱地址',
    encrypted_credential TEXT NOT NULL COMMENT 'AES-256-GCM加密的授权码',
    credential_iv VARCHAR(64) NOT NULL COMMENT '加密IV (Base64)',
    display_name VARCHAR(100) COMMENT '显示名称',
    last_sync_at DATETIME COMMENT '最后同步时间',
    sync_enabled BOOLEAN DEFAULT TRUE COMMENT '是否启用同步',
    sync_error TEXT COMMENT '同步错误信息',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    deleted_at DATETIME DEFAULT NULL COMMENT '软删除时间',

    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    UNIQUE KEY uk_user_email (user_id, account_email),
    INDEX idx_user_id (user_id),
    INDEX idx_provider (provider),
    INDEX idx_deleted_at (deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='邮箱账户表';

-- ============================================
-- 邮件表
-- 存储同步的邮件数据
-- ============================================
CREATE TABLE IF NOT EXISTS emails (
    id BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '邮件ID',
    message_id VARCHAR(255) NOT NULL COMMENT '邮件唯一标识 (Message-ID)',
    user_id BIGINT NOT NULL COMMENT '所属用户ID',
    account_id BIGINT NOT NULL COMMENT '所属账户ID',

    -- 发件人信息
    sender_name VARCHAR(255) COMMENT '发件人名称',
    sender_email VARCHAR(255) NOT NULL COMMENT '发件人邮箱',

    -- 邮件内容
    subject VARCHAR(512) COMMENT '邮件主题',
    content TEXT COMMENT '纯文本正文',
    content_html TEXT COMMENT 'HTML正文',
    content_type VARCHAR(20) DEFAULT 'text/plain' COMMENT '内容类型',

    -- 分类信息
    category VARCHAR(50) DEFAULT 'unclassified' COMMENT '分类: work_urgent, work_normal, personal, subscription, notification, promotion, spam',
    priority VARCHAR(20) DEFAULT 'medium' COMMENT '优先级: critical, high, medium, low',
    confidence_score DECIMAL(5,4) DEFAULT 0 COMMENT '分类置信度 (0-1)',
    classification_reason TEXT COMMENT '分类理由',

    -- 状态
    status VARCHAR(20) DEFAULT 'unread' COMMENT '状态: unread, read, archived',
    is_processed BOOLEAN DEFAULT FALSE COMMENT '是否已处理',
    has_attachment BOOLEAN DEFAULT FALSE COMMENT '是否有附件',

    -- 时间戳
    received_at DATETIME NOT NULL COMMENT '邮件接收时间',
    processed_at DATETIME COMMENT '处理时间',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    deleted_at DATETIME DEFAULT NULL COMMENT '软删除时间',

    UNIQUE KEY uk_message_id (message_id),
    INDEX idx_user_id (user_id),
    INDEX idx_account_id (account_id),
    INDEX idx_user_category (user_id, category),
    INDEX idx_user_received (user_id, received_at DESC),
    INDEX idx_status (status),
    INDEX idx_deleted_at (deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='邮件表';

-- ============================================
-- 行动项表
-- 从邮件中提取的行动项
-- ============================================
CREATE TABLE IF NOT EXISTS action_items (
    id BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '行动项ID',
    email_id BIGINT NOT NULL COMMENT '关联邮件ID',
    user_id BIGINT NOT NULL COMMENT '所属用户ID',

    task TEXT NOT NULL COMMENT '任务描述',
    task_type VARCHAR(50) COMMENT '任务类型: reply, review, submit, attend, prepare',
    deadline DATETIME COMMENT '截止时间',
    priority VARCHAR(20) DEFAULT 'medium' COMMENT '优先级: critical, high, medium, low',
    status VARCHAR(20) DEFAULT 'pending' COMMENT '状态: pending, in_progress, completed, cancelled',

    created_at DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    deleted_at DATETIME DEFAULT NULL COMMENT '软删除时间',

    FOREIGN KEY (email_id) REFERENCES emails(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,

    INDEX idx_user_id (user_id),
    INDEX idx_email_id (email_id),
    INDEX idx_user_status (user_id, status),
    INDEX idx_deadline (deadline),
    INDEX idx_deleted_at (deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='行动项表';

-- ============================================
-- 同步任务表
-- 记录同步任务状态
-- ============================================
CREATE TABLE IF NOT EXISTS sync_tasks (
    id BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '任务ID',
    user_id BIGINT NOT NULL COMMENT '所属用户ID',
    account_id COMMENT '账户ID (NULL表示同步所有账户)',

    status VARCHAR(20) DEFAULT 'pending' COMMENT '状态: pending, running, completed, failed',
    total_count INT DEFAULT 0 COMMENT '总邮件数',
    synced_count INT DEFAULT 0 COMMENT '已同步数',
    error_message TEXT COMMENT '错误信息',

    started_at DATETIME COMMENT '开始时间',
    completed_at DATETIME COMMENT '完成时间',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',

    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (account_id) REFERENCES email_accounts(id) ON DELETE SET NULL,

    INDEX idx_user_id (user_id),
    INDEX idx_account_id (account_id),
    INDEX idx_status (status),
    INDEX idx_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='同步任务表';

-- ============================================
-- 分类配置表
-- 用户自定义分类规则
-- ============================================
CREATE TABLE IF NOT EXISTS classification_configs (
    id BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '配置ID',
    user_id BIGINT NOT NULL COMMENT '所属用户ID',

    category VARCHAR(50) NOT NULL COMMENT '分类代码',
    category_name VARCHAR(100) NOT NULL COMMENT '分类名称',
    keywords TEXT COMMENT '关键词JSON数组',
    sender_patterns TEXT COMMENT '发件人模式JSON数组',
    priority VARCHAR(20) DEFAULT 'medium' COMMENT '默认优先级',
    is_enabled BOOLEAN DEFAULT TRUE COMMENT '是否启用',
    sort_order INT DEFAULT 0 COMMENT '排序顺序',

    created_at DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',

    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,

    UNIQUE KEY uk_user_category (user_id, category),
    INDEX idx_user_id (user_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='分类配置表';

-- ============================================
-- 初始化数据
-- ============================================

-- 插入默认用户 (用于开发测试)
INSERT INTO users (id, username, email, password_hash) VALUES
(1, 'default_user', 'user@example.com', '')
ON DUPLICATE KEY UPDATE username = username;

-- 插入默认分类配置
INSERT INTO classification_configs (user_id, category, category_name, priority, sort_order) VALUES
(1, 'work_urgent', '紧急工作', 'critical', 1),
(1, 'work_normal', '普通工作', 'medium', 2),
(1, 'personal', '个人邮件', 'medium', 3),
(1, 'subscription', '订阅邮件', 'low', 4),
(1, 'notification', '系统通知', 'low', 5),
(1, 'promotion', '营销推广', 'low', 6),
(1, 'spam', '垃圾邮件', 'low', 7)
ON DUPLICATE KEY UPDATE category_name = VALUES(category_name);

-- ============================================
-- 创建视图
-- ============================================

-- 邮件统计视图
CREATE OR REPLACE VIEW v_email_stats AS
SELECT
    user_id,
    category,
    COUNT(*) as total_count,
    SUM(CASE WHEN status = 'unread' THEN 1 ELSE 0 END) as unread_count,
    SUM(CASE WHEN DATE(received_at) = CURDATE() THEN 1 ELSE 0 END) as today_count,
    SUM(CASE WHEN received_at >= DATE_SUB(CURDATE(), INTERVAL 7 DAY) THEN 1 ELSE 0 END) as week_count
FROM emails
WHERE deleted_at IS NULL
GROUP BY user_id, category;

-- 账户同步状态视图
CREATE OR REPLACE VIEW v_account_sync_status AS
SELECT
    a.id as account_id,
    a.user_id,
    a.provider,
    a.account_email,
    a.sync_enabled,
    a.last_sync_at,
    a.sync_error,
    COUNT(e.id) as total_emails,
    MAX(e.received_at) as latest_email_at
FROM email_accounts a
LEFT JOIN emails e ON a.id = e.account_id AND e.deleted_at IS NULL
WHERE a.deleted_at IS NULL
GROUP BY a.id, a.user_id, a.provider, a.account_email, a.sync_enabled, a.last_sync_at, a.sync_error;

-- ============================================
-- 结束
-- ============================================
