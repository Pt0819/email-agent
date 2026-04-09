package provider

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// MockProvider Mock邮件提供商(用于测试)
type MockProvider struct {
	name       string
	email      string
	credential string
	connected  bool
	mu         sync.Mutex

	// 可配置的模拟数据
	MockEmails     []*Email
	MockSummaries  []*EmailSummary
	MockConnectErr error
}

// NewMockProvider 创建Mock Provider
func NewMockProvider(config *ProviderConfig) EmailProvider {
	return &MockProvider{
		name:      "mock",
		connected: false,
	}
}

// Name 返回提供商名称
func (p *MockProvider) Name() string {
	return p.name
}

// Connect 连接邮箱服务器
func (p *MockProvider) Connect(ctx context.Context, email, credential string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.MockConnectErr != nil {
		return p.MockConnectErr
	}

	p.email = email
	p.credential = credential
	p.connected = true

	return nil
}

// TestConnection 测试连接
func (p *MockProvider) TestConnection(ctx context.Context) (*ConnectionResult, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.MockConnectErr != nil {
		return &ConnectionResult{
			Success: false,
			Message: p.MockConnectErr.Error(),
		}, nil
	}

	return &ConnectionResult{
		Success: true,
		Message: "连接成功",
	}, nil
}

// FetchEmailList 获取邮件列表
func (p *MockProvider) FetchEmailList(ctx context.Context, since time.Time, limit int) ([]*EmailSummary, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if !p.connected {
		return nil, fmt.Errorf("未连接")
	}

	// 返回模拟数据
	if p.MockSummaries != nil {
		result := make([]*EmailSummary, 0, len(p.MockSummaries))
		for _, s := range p.MockSummaries {
			if s.ReceivedAt.After(since) {
				result = append(result, s)
			}
		}
		if len(result) > limit {
			result = result[:limit]
		}
		return result, nil
	}

	// 默认返回空列表
	return []*EmailSummary{}, nil
}

// FetchEmailDetail 获取邮件详情
func (p *MockProvider) FetchEmailDetail(ctx context.Context, messageID string) (*Email, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if !p.connected {
		return nil, fmt.Errorf("未连接")
	}

	// 查找模拟数据
	if p.MockEmails != nil {
		for _, email := range p.MockEmails {
			if email.MessageID == messageID {
				return email, nil
			}
		}
	}

	return nil, fmt.Errorf("邮件不存在: %s", messageID)
}

// FetchEmails 批量获取邮件
func (p *MockProvider) FetchEmails(ctx context.Context, since time.Time, limit int) (*SyncResult, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if !p.connected {
		return nil, fmt.Errorf("未连接")
	}

	// 返回模拟数据
	if p.MockEmails != nil {
		emails := make([]Email, 0)
		summaries := make([]EmailSummary, 0)

		for _, email := range p.MockEmails {
			if email.ReceivedAt.After(since) {
				emails = append(emails, *email)
				summaries = append(summaries, EmailSummary{
					MessageID:     email.MessageID,
					Subject:       email.Subject,
					SenderName:    email.SenderName,
					SenderEmail:   email.SenderEmail,
					ReceivedAt:    email.ReceivedAt,
					HasAttachment: email.HasAttachment,
				})
			}
		}

		if len(emails) > limit {
			emails = emails[:limit]
			summaries = summaries[:limit]
		}

		return &SyncResult{
			TotalCount:   len(emails),
			SyncedCount:  len(emails),
			SkippedCount: 0,
			ErrorCount:   0,
			Emails:       emails,
			Summaries:    summaries,
		}, nil
	}

	return &SyncResult{
		TotalCount:   0,
		SyncedCount:  0,
		SkippedCount: 0,
		ErrorCount:   0,
		Emails:       []Email{},
		Summaries:    []EmailSummary{},
	}, nil
}

// Disconnect 断开连接
func (p *MockProvider) Disconnect() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.connected = false
	return nil
}

// IsConnected 检查是否已连接
func (p *MockProvider) IsConnected() bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.connected
}

// SetMockEmails 设置模拟邮件数据
func (p *MockProvider) SetMockEmails(emails []*Email) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.MockEmails = emails
}

// SetMockSummaries 设置模拟摘要数据
func (p *MockProvider) SetMockSummaries(summaries []*EmailSummary) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.MockSummaries = summaries
}

// SetConnectError 设置连接错误
func (p *MockProvider) SetConnectError(err error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.MockConnectErr = err
}

func init() {
	// 注册Mock Provider
	Register("mock", NewMockProvider)
}
