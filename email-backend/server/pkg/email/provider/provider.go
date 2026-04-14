// Package provider 邮件提供商接口定义
package provider

import (
	"context"
	"time"
)

// EmailSummary 邮件摘要信息
type EmailSummary struct {
	MessageID     string    `json:"message_id"`     // 邮件唯一标识
	Subject       string    `json:"subject"`        // 邮件主题
	SenderName    string    `json:"sender_name"`    // 发件人名称
	SenderEmail   string    `json:"sender_email"`   // 发件人邮箱
	ReceivedAt    time.Time `json:"received_at"`    // 接收时间
	HasAttachment bool      `json:"has_attachment"` // 是否有附件
	Size          int       `json:"size"`           // 邮件大小(字节)
}

// Email 完整邮件信息
type Email struct {
	MessageID     string    `json:"message_id"`     // 邮件唯一标识
	Subject       string    `json:"subject"`        // 邮件主题
	SenderName    string    `json:"sender_name"`    // 发件人名称
	SenderEmail   string    `json:"sender_email"`   // 发件人邮箱
	To            string    `json:"to"`             // 收件人
	CC            []string  `json:"cc"`             // 抄送列表
	Content       string    `json:"content"`        // 纯文本正文
	ContentHTML   string    `json:"content_html"`   // HTML正文
	ContentType   string    `json:"content_type"`   // 内容类型
	ReceivedAt    time.Time `json:"received_at"`    // 接收时间
	HasAttachment bool      `json:"has_attachment"` // 是否有附件
}

// Attachment 附件信息
type Attachment struct {
	Filename    string `json:"filename"`     // 文件名
	ContentType string `json:"content_type"` // 内容类型
	Size        int    `json:"size"`         // 大小
}

// ConnectionResult 连接结果
type ConnectionResult struct {
	Success bool   `json:"success"` // 是否成功
	Message string `json:"message"` // 结果消息
}

// SyncResult 同步结果
type SyncResult struct {
	TotalCount    int          `json:"total_count"`    // 总邮件数
	SyncedCount   int          `json:"synced_count"`   // 已同步数
	SkippedCount  int          `json:"skipped_count"`  // 跳过数
	ErrorCount    int          `json:"error_count"`    // 错误数
	LastError     string       `json:"last_error"`     // 最后错误信息
	Emails        []Email      `json:"emails"`         // 同步的邮件列表
	Summaries     []EmailSummary `json:"summaries"`    // 邮件摘要列表
}

// EmailProvider 邮件提供商接口
// 定义了邮件服务提供商必须实现的方法
type EmailProvider interface {
	// Name 返回提供商名称
	Name() string

	// Connect 连接邮箱服务器
	// email: 邮箱地址
	// credential: 授权码/密码(已解密)
	Connect(ctx context.Context, email, credential string) error

	// TestConnection 测试连接
	// 返回连接测试结果
	TestConnection(ctx context.Context) (*ConnectionResult, error)

	// FetchEmailList 获取邮件列表
	// since: 获取此时间之后的邮件
	// limit: 最大获取数量
	FetchEmailList(ctx context.Context, since time.Time, limit int) ([]*EmailSummary, error)

	// FetchEmailDetail 获取邮件详情
	// messageID: 邮件唯一标识
	FetchEmailDetail(ctx context.Context, messageID string) (*Email, error)

	// FetchEmails 批量获取邮件(带详情)
	// since: 获取此时间之后的邮件
	// limit: 最大获取数量
	FetchEmails(ctx context.Context, since time.Time, limit int) (*SyncResult, error)

	// Disconnect 断开连接
	Disconnect() error

	// IsConnected 检查是否已连接
	IsConnected() bool
}

// ProviderConfig Provider配置
type ProviderConfig struct {
	Server   string `json:"server"`    // 服务器地址
	Port     int    `json:"port"`      // 端口
	UseSSL   bool   `json:"use_ssl"`   // 是否使用SSL
	Timeout  int    `json:"timeout"`   // 超时时间(秒)
}

// ProviderFactory Provider工厂函数类型
type ProviderFactory func(config *ProviderConfig) EmailProvider

// 全局Provider注册表
var providers = make(map[string]ProviderFactory)

// Register 注册Provider工厂
func Register(name string, factory ProviderFactory) {
	providers[name] = factory
}

// Create 创建Provider实例
func Create(name string, config *ProviderConfig) (EmailProvider, bool) {
	factory, ok := providers[name]
	if !ok {
		return nil, false
	}
	return factory(config), true
}

// ListProviders 列出所有已注册的Provider
func ListProviders() []string {
	names := make([]string, 0, len(providers))
	for name := range providers {
		names = append(names, name)
	}
	return names
}

// IsProviderAvailable 检查Provider是否可用
func IsProviderAvailable(name string) bool {
	_, ok := providers[name]
	return ok
}
