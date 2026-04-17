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
	p := &MockProvider{
		name:      "mock",
		connected: false,
	}

	// 初始化时加载模拟邮件数据
	p.loadMockEmails()

	return p
}

// loadMockEmails 加载模拟邮件数据
func (p *MockProvider) loadMockEmails() {
	now := time.Now()

	// 模拟不同类别的邮件数据
	p.MockEmails = []*Email{
		// 1. 紧急工作邮件
		{
			MessageID:     "<mock-001@mock.local>",
			SenderName:    "张三",
			SenderEmail:   "zhangsan@company.com",
			Subject:       "【紧急】项目上线前最后检查 - 请尽快确认",
			Content:       "您好，项目将于明天上午10点上线，请务必在今天下班前完成最后的代码审查和测试。如有问题请立即联系我。",
			ContentHTML:   "",
			ContentType:   "text/plain",
			HasAttachment: true,
			ReceivedAt:    now.Add(-2 * time.Hour),
		},
		// 2. 普通工作邮件
		{
			MessageID:     "<mock-002@mock.local>",
			SenderName:    "李四",
			SenderEmail:   "lisi@company.com",
			Subject:       "关于下周会议安排的通知",
			Content:       "各位同事好，下周三下午2点在会议室A召开项目进度会议，请提前准备好各自负责模块的进度报告。谢谢！",
			ContentHTML:   "",
			ContentType:   "text/plain",
			HasAttachment: false,
			ReceivedAt:    now.Add(-5 * time.Hour),
		},
		// 3. 个人邮件
		{
			MessageID:     "<mock-003@mock.local>",
			SenderName:    "王五",
			SenderEmail:   "wangwu@personal.com",
			Subject:       "周末聚会邀请",
			Content:       "嗨，这周六晚上7点我们在老地方聚会，好久没见了，大家都很想你。有空的话记得来哦！",
			ContentHTML:   "",
			ContentType:   "text/plain",
			HasAttachment: false,
			ReceivedAt:    now.Add(-24 * time.Hour),
		},
		// 4. 订阅邮件
		{
			MessageID:     "<mock-004@mock.local>",
			SenderName:    "技术周刊",
			SenderEmail:   "newsletter@techweekly.com",
			Subject:       "【技术周刊】本周热门：AI最新进展、Go并发模式、云原生实践",
			Content:       "本期内容：1. GPT-5最新动态解析 2. Go语言高并发模式实战 3. Kubernetes最佳实践分享 4. 微服务架构设计模式...",
			ContentHTML:   "<html><body><h1>技术周刊</h1><p>本期内容精彩...</p></body></html>",
			ContentType:   "text/html",
			HasAttachment: false,
			ReceivedAt:    now.Add(-48 * time.Hour),
		},
		// 5. 系统通知
		{
			MessageID:     "<mock-005@mock.local>",
			SenderName:    "系统通知",
			SenderEmail:   "noreply@system.com",
			Subject:       "您的账户安全提醒",
			Content:       "尊敬的用户，我们检测到您的账户在新设备上登录。如果这不是您本人的操作，请立即修改密码并联系客服。",
			ContentHTML:   "",
			ContentType:   "text/plain",
			HasAttachment: false,
			ReceivedAt:    now.Add(-3 * time.Hour),
		},
		// 6. 营销推广
		{
			MessageID:     "<mock-006@mock.local>",
			SenderName:    "商城优惠",
			SenderEmail:   "promo@shop.com",
			Subject:       "限时特惠！全场5折起，仅限今日",
			Content:       "亲爱的用户，今日限时特惠活动开启！全场商品5折起，更有满减优惠等你来拿。点击查看详情...",
			ContentHTML:   "<html><body><h1>限时特惠</h1><p>全场5折起...</p></body></html>",
			ContentType:   "text/html",
			HasAttachment: false,
			ReceivedAt:    now.Add(-12 * time.Hour),
		},
		// 7. 包含会议信息的邮件
		{
			MessageID:     "<mock-007@mock.local>",
			SenderName:    "会议助手",
			SenderEmail:   "meeting@company.com",
			Subject:       "会议邀请：产品评审会 - 周五 14:00",
			Content:       "您已被邀请参加产品评审会议\n\n时间：本周五 14:00-16:00\n地点：会议室B\n参会人员：产品部、研发部、设计部\n\n会议链接：https://meeting.company.com/room/123\n\n请提前准备相关材料。",
			ContentHTML:   "",
			ContentType:   "text/plain",
			HasAttachment: true,
			ReceivedAt:    now.Add(-6 * time.Hour),
		},
		// 8. 包含待办事项的邮件
		{
			MessageID:     "<mock-008@mock.local>",
			SenderName:    "项目经理",
			SenderEmail:   "pm@company.com",
			Subject:       "本周任务清单 - 请按时完成",
			Content:       "本周需要完成的任务：\n\n1. 完成API文档编写（周三前）\n2. 代码评审（周四前）\n3. 提交测试报告（周五前）\n\n如有问题请及时沟通。",
			ContentHTML:   "",
			ContentType:   "text/plain",
			HasAttachment: false,
			ReceivedAt:    now.Add(-36 * time.Hour),
		},
	}

	// 生成摘要数据
	p.MockSummaries = make([]*EmailSummary, len(p.MockEmails))
	for i, email := range p.MockEmails {
		p.MockSummaries[i] = &EmailSummary{
			MessageID:     email.MessageID,
			Subject:       email.Subject,
			SenderName:    email.SenderName,
			SenderEmail:   email.SenderEmail,
			ReceivedAt:    email.ReceivedAt,
			HasAttachment: email.HasAttachment,
			Size:          1024,
		}
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
