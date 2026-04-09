package provider

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func TestProviderRegistry(t *testing.T) {
	// 清空注册表进行测试
	originalProviders := providers
	defer func() {
		providers = originalProviders
	}()

	providers = make(map[string]ProviderFactory)

	// 测试注册
	Register("test", func(config *ProviderConfig) EmailProvider {
		return &MockProvider{}
	})

	if len(providers) != 1 {
		t.Errorf("Register() failed, providers count = %d, want 1", len(providers))
	}

	// 测试创建
	provider, ok := Create("test", &ProviderConfig{})
	if !ok {
		t.Error("Create() failed, provider not found")
	}
	if provider == nil {
		t.Error("Create() returned nil provider")
	}

	// 测试不存在
	_, ok = Create("nonexistent", &ProviderConfig{})
	if ok {
		t.Error("Create() should return false for nonexistent provider")
	}

	// 测试列表
	list := ListProviders()
	if len(list) != 1 || list[0] != "test" {
		t.Errorf("ListProviders() = %v, want [test]", list)
	}

	// 测试可用性检查
	if !IsProviderAvailable("test") {
		t.Error("IsProviderAvailable() should return true for test")
	}
	if IsProviderAvailable("nonexistent") {
		t.Error("IsProviderAvailable() should return false for nonexistent")
	}
}

func TestMockProviderConnect(t *testing.T) {
	provider := NewMockProvider(&ProviderConfig{}).(*MockProvider)

	ctx := context.Background()

	// 测试连接
	err := provider.Connect(ctx, "test@example.com", "password")
	if err != nil {
		t.Errorf("Connect() error = %v", err)
	}

	if !provider.IsConnected() {
		t.Error("IsConnected() should return true after connect")
	}

	// 测试断开
	err = provider.Disconnect()
	if err != nil {
		t.Errorf("Disconnect() error = %v", err)
	}

	if provider.IsConnected() {
		t.Error("IsConnected() should return false after disconnect")
	}
}

func TestMockProviderTestConnection(t *testing.T) {
	provider := NewMockProvider(&ProviderConfig{}).(*MockProvider)
	ctx := context.Background()

	// 测试成功连接
	result, err := provider.TestConnection(ctx)
	if err != nil {
		t.Errorf("TestConnection() error = %v", err)
	}
	if !result.Success {
		t.Errorf("TestConnection() success = %v, want true", result.Success)
	}

	// 测试失败连接
	provider.SetConnectError(fmt.Errorf("connection failed"))
	result, err = provider.TestConnection(ctx)
	if err != nil {
		t.Errorf("TestConnection() error = %v", err)
	}
	if result.Success {
		t.Errorf("TestConnection() success = %v, want false", result.Success)
	}
}

func TestMockProviderFetchEmailList(t *testing.T) {
	provider := NewMockProvider(&ProviderConfig{}).(*MockProvider)
	ctx := context.Background()

	// 先连接
	provider.Connect(ctx, "test@example.com", "password")

	// 设置模拟数据
	now := time.Now()
	summaries := []*EmailSummary{
		{
			MessageID:   "msg1",
			Subject:     "Test Email 1",
			SenderEmail: "sender1@example.com",
			ReceivedAt:  now.Add(-1 * time.Hour),
		},
		{
			MessageID:   "msg2",
			Subject:     "Test Email 2",
			SenderEmail: "sender2@example.com",
			ReceivedAt:  now.Add(-2 * time.Hour),
		},
	}
	provider.SetMockSummaries(summaries)

	// 测试获取列表
	since := now.Add(-3 * time.Hour)
	result, err := provider.FetchEmailList(ctx, since, 10)
	if err != nil {
		t.Errorf("FetchEmailList() error = %v", err)
	}
	if len(result) != 2 {
		t.Errorf("FetchEmailList() count = %d, want 2", len(result))
	}

	// 测试限制数量
	result, err = provider.FetchEmailList(ctx, since, 1)
	if err != nil {
		t.Errorf("FetchEmailList() error = %v", err)
	}
	if len(result) != 1 {
		t.Errorf("FetchEmailList() with limit count = %d, want 1", len(result))
	}

	// 测试时间过滤
	since = now.Add(-30 * time.Minute)
	result, err = provider.FetchEmailList(ctx, since, 10)
	if err != nil {
		t.Errorf("FetchEmailList() error = %v", err)
	}
	if len(result) != 0 {
		t.Errorf("FetchEmailList() with recent since count = %d, want 0", len(result))
	}
}

func TestMockProviderFetchEmailDetail(t *testing.T) {
	provider := NewMockProvider(&ProviderConfig{}).(*MockProvider)
	ctx := context.Background()

	// 先连接
	provider.Connect(ctx, "test@example.com", "password")

	// 设置模拟数据
	now := time.Now()
	emails := []*Email{
		{
			MessageID:   "msg1",
			Subject:     "Test Email 1",
			SenderEmail: "sender1@example.com",
			Content:     "This is test content",
			ReceivedAt:  now,
		},
	}
	provider.SetMockEmails(emails)

	// 测试获取存在的邮件
	email, err := provider.FetchEmailDetail(ctx, "msg1")
	if err != nil {
		t.Errorf("FetchEmailDetail() error = %v", err)
	}
	if email == nil || email.Subject != "Test Email 1" {
		t.Errorf("FetchEmailDetail() email = %v, want subject 'Test Email 1'", email)
	}

	// 测试获取不存在的邮件
	_, err = provider.FetchEmailDetail(ctx, "nonexistent")
	if err == nil {
		t.Error("FetchEmailDetail() should return error for nonexistent email")
	}
}

func TestMockProviderNotConnected(t *testing.T) {
	provider := NewMockProvider(&ProviderConfig{}).(*MockProvider)
	ctx := context.Background()

	// 不连接直接调用方法应该返回错误
	_, err := provider.FetchEmailList(ctx, time.Time{}, 10)
	if err == nil {
		t.Error("FetchEmailList() should return error when not connected")
	}

	_, err = provider.FetchEmailDetail(ctx, "msg1")
	if err == nil {
		t.Error("FetchEmailDetail() should return error when not connected")
	}

	_, err = provider.FetchEmails(ctx, time.Time{}, 10)
	if err == nil {
		t.Error("FetchEmails() should return error when not connected")
	}
}

func TestMockProviderFetchEmails(t *testing.T) {
	provider := NewMockProvider(&ProviderConfig{}).(*MockProvider)
	ctx := context.Background()

	// 先连接
	provider.Connect(ctx, "test@example.com", "password")

	// 设置模拟数据
	now := time.Now()
	emails := []*Email{
		{
			MessageID:   "msg1",
			Subject:     "Test Email 1",
			SenderEmail: "sender1@example.com",
			Content:     "Content 1",
			ReceivedAt:  now.Add(-1 * time.Hour),
		},
		{
			MessageID:   "msg2",
			Subject:     "Test Email 2",
			SenderEmail: "sender2@example.com",
			Content:     "Content 2",
			ReceivedAt:  now.Add(-2 * time.Hour),
		},
	}
	provider.SetMockEmails(emails)

	// 测试批量获取
	since := now.Add(-3 * time.Hour)
	result, err := provider.FetchEmails(ctx, since, 10)
	if err != nil {
		t.Errorf("FetchEmails() error = %v", err)
	}
	if result.TotalCount != 2 {
		t.Errorf("FetchEmails() total count = %d, want 2", result.TotalCount)
	}
	if len(result.Emails) != 2 {
		t.Errorf("FetchEmails() emails count = %d, want 2", len(result.Emails))
	}
}
