-- 邮件分类系统数据库初始化脚本

-- 创建数据库（如果不存在）
CREATE DATABASE IF NOT EXISTS email_system DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

USE email_system;

-- 用户表
CREATE TABLE IF NOT EXISTS users (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    username VARCHAR(100) NOT NULL UNIQUE,
    email VARCHAR(255) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_username (username),
    INDEX idx_email (email)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 邮箱账户表
CREATE TABLE IF NOT EXISTS email_accounts (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    user_id BIGINT NOT NULL,
    provider VARCHAR(20) NOT NULL COMMENT '126, gmail, outlook, imap',
    account_email VARCHAR(255) NOT NULL,
    encrypted_credential TEXT NOT NULL COMMENT 'AES加密的授权码',
    credential_iv VARCHAR(64) NOT NULL COMMENT '加密IV',
    last_sync_at DATETIME,
    sync_enabled BOOLEAN DEFAULT TRUE,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    UNIQUE INDEX idx_user_email (user_id, account_email),
    INDEX idx_provider (provider),
    INDEX idx_sync_enabled (sync_enabled)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 邮件表
CREATE TABLE IF NOT EXISTS emails (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    message_id VARCHAR(255) UNIQUE NOT NULL,
    user_id BIGINT NOT NULL,
    account_id BIGINT NOT NULL,

    -- 发件人信息
    sender_name VARCHAR(255),
    sender_email VARCHAR(255) NOT NULL,

    -- 邮件内容
    subject VARCHAR(512),
    content TEXT,
    content_html TEXT,
    content_type VARCHAR(20) DEFAULT 'text/plain',

    -- 分类信息
    category VARCHAR(50) DEFAULT 'unclassified' COMMENT 'work_urgent, work_normal, personal, subscription, notification, promotion, spam, unclassified',
    priority VARCHAR(20) DEFAULT 'medium' COMMENT 'critical, high, medium, low',
    confidence_score DECIMAL(5,4) DEFAULT 0,
    classification_reason TEXT,

    -- 状态
    status VARCHAR(20) DEFAULT 'unread' COMMENT 'unread, read, processed, archived',
    is_processed BOOLEAN DEFAULT FALSE,
    has_attachment BOOLEAN DEFAULT FALSE,

    -- 时间
    received_at DATETIME NOT NULL,
    processed_at DATETIME,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (account_id) REFERENCES email_accounts(id) ON DELETE CASCADE,

    INDEX idx_user_category (user_id, category),
    INDEX idx_user_received (user_id, received_at DESC),
    INDEX idx_status (status),
    INDEX idx_account_id (account_id),
    INDEX idx_message_id (message_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 行动项表
CREATE TABLE IF NOT EXISTS action_items (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    email_id BIGINT NOT NULL,
    user_id BIGINT NOT NULL,
    task TEXT NOT NULL,
    task_type VARCHAR(50) COMMENT 'reply, review, submit, attend, prepare',
    deadline DATETIME,
    priority VARCHAR(20) DEFAULT 'medium',
    status VARCHAR(20) DEFAULT 'pending',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (email_id) REFERENCES emails(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    INDEX idx_user_status (user_id, status),
    INDEX idx_deadline (deadline)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 每日摘要表
CREATE TABLE IF NOT EXISTS daily_summaries (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    user_id BIGINT NOT NULL,
    summary_date DATE NOT NULL,
    summary_content TEXT,
    important_emails JSON,
    pending_actions JSON,
    statistics JSON,
    generated_at DATETIME,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    UNIQUE INDEX idx_user_date (user_id, summary_date),
    INDEX idx_summary_date (summary_date)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- LLM配置表
CREATE TABLE IF NOT EXISTS llm_configs (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    user_id BIGINT NOT NULL,
    provider VARCHAR(50) NOT NULL COMMENT 'deepseek, zhipu, qwen',
    model_name VARCHAR(100),
    api_key_encrypted TEXT,
    api_key_iv VARCHAR(64),
    base_url VARCHAR(255),
    config_json JSON,
    is_active BOOLEAN DEFAULT TRUE,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    INDEX idx_user_provider (user_id, provider),
    INDEX idx_is_active (is_active)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 插入默认测试用户（密码: admin123，实际应该使用bcrypt等加密）
INSERT INTO users (username, email, password_hash) VALUES
('admin', 'admin@example.com', 'admin_hash_placeholder')
ON DUPLICATE KEY UPDATE email=email;
