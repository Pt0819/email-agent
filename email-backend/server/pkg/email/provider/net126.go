// Package provider 邮件提供商实现 - 网易126邮箱
package provider

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"mime"
	"strings"
	"sync"
	"time"

	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
	"github.com/emersion/go-imap-id"
	"github.com/emersion/go-message"
	_ "github.com/emersion/go-message/charset" // 注册字符集解码器
)

// Net126Provider 网易126邮箱Provider
// 使用 go-imap 库实现 IMAP 协议
type Net126Provider struct {
	name       string
	server     string
	port       int
	useSSL     bool
	email      string
	credential string
	client     *client.Client
	timeout    time.Duration
	mu         sync.Mutex

	// 重试配置
	maxRetries    int
	retryInterval time.Duration
}

// NewNet126Provider 创建126邮箱Provider
func NewNet126Provider(config *ProviderConfig) EmailProvider {
	p := &Net126Provider{
		name:          "126",
		server:        "imap.126.com",
		port:          993,
		useSSL:        true,
		timeout:       30 * time.Second,
		maxRetries:    3,
		retryInterval: 2 * time.Second,
	}

	if config != nil {
		if config.Server != "" {
			p.server = config.Server
		}
		if config.Port > 0 {
			p.port = config.Port
		}
		if config.Timeout > 0 {
			p.timeout = time.Duration(config.Timeout) * time.Second
		}
	}

	return p
}

// Name 返回提供商名称
func (p *Net126Provider) Name() string {
	return p.name
}

// Connect 连接邮箱服务器（带重试机制）
func (p *Net126Provider) Connect(ctx context.Context, email, credential string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.email = email
	p.credential = credential

	var lastErr error
	for attempt := 1; attempt <= p.maxRetries; attempt++ {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		err := p.connectOnce()
		if err == nil {
			return nil
		}

		lastErr = err

		// 认证失败不重试
		if strings.Contains(err.Error(), "认证失败") ||
			strings.Contains(err.Error(), "LOGIN failed") {
			return err
		}

		if attempt < p.maxRetries {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(p.retryInterval):
				continue
			}
		}
	}

	return fmt.Errorf("连接失败(重试%d次): %w", p.maxRetries, lastErr)
}

// connectOnce 单次连接尝试
func (p *Net126Provider) connectOnce() error {
	addr := fmt.Sprintf("%s:%d", p.server, p.port)

	var c *client.Client
	var err error

	if p.useSSL {
		tlsConfig := &tls.Config{ServerName: p.server}
		c, err = client.DialTLS(addr, tlsConfig)
	} else {
		c, err = client.Dial(addr)
	}

	if err != nil {
		return fmt.Errorf("连接服务器失败: %w", err)
	}

	p.client = c
	p.client.Timeout = p.timeout

	if err := p.client.Login(p.email, p.credential); err != nil {
		p.client.Close()
		p.client = nil
		return fmt.Errorf("认证失败: %w", err)
	}

	// 发送ID命令标识客户端身份（126邮箱要求）
	// 避免被标记为"Unsafe Login"
	if err := p.sendIDCommand(); err != nil {
		p.client.Close()
		p.client = nil
		return fmt.Errorf("发送客户端标识失败: %w", err)
	}

	return nil
}

// sendIDCommand 发送IMAP ID命令标识客户端
// 126邮箱要求发送ID命令，否则SELECT会被标记为"Unsafe Login"
// 使用 go-imap-id 库的标准实现
func (p *Net126Provider) sendIDCommand() error {
	idClient := id.NewClient(p.client)
	_, err := idClient.ID(id.ID{
		"name":    "Foxmail",
		"version": "7.2",
		"vendor":  "Tencent",
	})
	return err
}

// TestConnection 测试连接
func (p *Net126Provider) TestConnection(ctx context.Context) (*ConnectionResult, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.client == nil {
		return &ConnectionResult{Success: false, Message: "未连接到服务器"}, nil
	}

	if err := p.client.Noop(); err != nil {
		return &ConnectionResult{
			Success: false,
			Message: fmt.Sprintf("连接已断开: %v", err),
		}, nil
	}

	return &ConnectionResult{Success: true, Message: "连接正常"}, nil
}

// FetchEmailList 获取邮件列表
func (p *Net126Provider) FetchEmailList(ctx context.Context, since time.Time, limit int) ([]*EmailSummary, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.client == nil {
		return nil, fmt.Errorf("未连接")
	}

	// 126邮箱要求：每次SELECT前发送ID命令
	// 避免被标记为"Unsafe Login"
	if err := p.sendIDCommand(); err != nil {
		return nil, fmt.Errorf("发送客户端标识失败: %w", err)
	}

	// 选择收件箱
	mailbox, err := p.client.Select("INBOX", true) // true = readonly
	if err != nil {
		// SELECT失败时尝试非只读方式
		mailbox, err = p.client.Select("INBOX", false)
		if err != nil {
			return nil, fmt.Errorf("选择收件箱失败: %w", err)
		}
	}
	_ = mailbox

	// 搜索邮件
	criteria := imap.NewSearchCriteria()
	if !since.IsZero() {
		criteria.Since = since
	}

	uids, err := p.client.UidSearch(criteria)
	if err != nil {
		uids, err = p.client.UidSearch(imap.NewSearchCriteria())
		if err != nil {
			return nil, fmt.Errorf("搜索邮件失败: %w", err)
		}
	}

	if len(uids) > limit {
		uids = uids[len(uids)-limit:]
	}

	if len(uids) == 0 {
		return []*EmailSummary{}, nil
	}

	uidSet := new(imap.SeqSet)
	for _, uid := range uids {
		uidSet.AddNum(uid)
	}

	fetchItems := []imap.FetchItem{
		imap.FetchEnvelope,
		imap.FetchRFC822Size,
		imap.FetchFlags,
		imap.FetchBodyStructure,
	}
	messages := make(chan *imap.Message, 10)
	done := make(chan error, 1)

	go func() {
		done <- p.client.UidFetch(uidSet, fetchItems, messages)
	}()

	summaries := make([]*EmailSummary, 0, len(uids))
	for msg := range messages {
		summary := p.parseEnvelope(msg)
		if summary != nil {
			summaries = append(summaries, summary)
		}
	}

	if err := <-done; err != nil {
		return nil, fmt.Errorf("获取邮件头失败: %w", err)
	}

	return summaries, nil
}

// parseEnvelope 解析邮件信封
func (p *Net126Provider) parseEnvelope(msg *imap.Message) *EmailSummary {
	if msg == nil || msg.Envelope == nil {
		return nil
	}

	summary := &EmailSummary{
		MessageID:  msg.Envelope.MessageId,
		Subject:    decodeSubject(msg.Envelope.Subject),
		ReceivedAt: msg.Envelope.Date,
		Size:       int(msg.Size),
	}

	if len(msg.Envelope.From) > 0 {
		from := msg.Envelope.From[0]
		summary.SenderName = sanitizeString(from.PersonalName)
		summary.SenderEmail = from.Address()
	}

	if msg.BodyStructure != nil {
		summary.HasAttachment = hasAttachment(msg.BodyStructure)
	}

	return summary
}

// hasAttachment 递归检查BodyStructure是否有附件
func hasAttachment(bs *imap.BodyStructure) bool {
	if bs == nil {
		return false
	}

	if bs.Disposition == "attachment" {
		return true
	}

	for _, part := range bs.Parts {
		if hasAttachment(part) {
			return true
		}
	}

	// 非文本非多部分的独立部分视为附件
	if len(bs.Parts) == 0 && bs.MIMEType != "text" && bs.MIMEType != "multipart" && bs.MIMEType != "" {
		return true
	}

	return false
}

// FetchEmailDetail 获取邮件详情
func (p *Net126Provider) FetchEmailDetail(ctx context.Context, messageID string) (*Email, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.client == nil {
		return nil, fmt.Errorf("未连接")
	}

	// 126邮箱要求：每次SELECT前发送ID命令
	if err := p.sendIDCommand(); err != nil {
		return nil, fmt.Errorf("发送客户端标识失败: %w", err)
	}

	// 选择收件箱
	mailbox, err := p.client.Select("INBOX", true)
	if err != nil {
		mailbox, err = p.client.Select("INBOX", false)
		if err != nil {
			return nil, fmt.Errorf("选择收件箱失败: %w", err)
		}
	}
	_ = mailbox

	// 搜索指定Message-ID
	criteria := imap.NewSearchCriteria()
	criteria.Header = map[string][]string{"Message-ID": {messageID}}

	uids, err := p.client.UidSearch(criteria)
	if err != nil {
		return nil, fmt.Errorf("搜索邮件失败: %w", err)
	}
	if len(uids) == 0 {
		return nil, fmt.Errorf("邮件不存在: %s", messageID)
	}

	uidSet := new(imap.SeqSet)
	uidSet.AddNum(uids[0])

	// 获取完整邮件
	fetchItems := []imap.FetchItem{imap.FetchEnvelope, imap.FetchBodyStructure, imap.FetchRFC822}
	messages := make(chan *imap.Message, 1)
	done := make(chan error, 1)

	go func() {
		done <- p.client.UidFetch(uidSet, fetchItems, messages)
	}()

	msg := <-messages
	if err := <-done; err != nil {
		return nil, fmt.Errorf("获取邮件失败: %w", err)
	}
	if msg == nil {
		return nil, fmt.Errorf("获取邮件内容为空")
	}

	// 构建返回对象
	email := &Email{
		MessageID:     msg.Envelope.MessageId,
		Subject:       decodeSubject(msg.Envelope.Subject),
		ReceivedAt:    msg.Envelope.Date,
		HasAttachment: hasAttachment(msg.BodyStructure),
	}

	if len(msg.Envelope.From) > 0 {
		from := msg.Envelope.From[0]
		email.SenderName = sanitizeString(from.PersonalName)
		email.SenderEmail = from.Address()
	}

	if len(msg.Envelope.To) > 0 {
		email.To = msg.Envelope.To[0].Address()
	}

	for _, cc := range msg.Envelope.Cc {
		email.CC = append(email.CC, cc.Address())
	}

	// 解析邮件正文
	section := &imap.BodySectionName{}
	body := msg.GetBody(section)
	if body != nil {
		parseBody(body, email)
	}

	return email, nil
}

// parseBody 解析邮件正文（支持嵌套multipart）
func parseBody(r io.Reader, email *Email) {
	msgReader, err := message.Read(r)
	if err != nil {
		return
	}
	parseMessagePart(msgReader, email)
}

// parseMessagePart 递归解析邮件部分
func parseMessagePart(entity *message.Entity, email *Email) {
	// 如果是multipart，递归处理每个子部分
	if mr := entity.MultipartReader(); mr != nil {
		for {
			part, err := mr.NextPart()
			if err != nil {
				break
			}
			parseMessagePart(part, email)
		}
		return
	}

	// 非multipart，读取正文
	contentType, params, _ := entity.Header.ContentType()
	charset := params["charset"]

	if !strings.HasPrefix(contentType, "text/") {
		return
	}

	data, err := io.ReadAll(entity.Body)
	if err != nil {
		return
	}

	// 如果指定了charset且不是UTF-8，尝试转换
	body := string(data)
	if charset != "" && !strings.EqualFold(charset, "utf-8") && !strings.EqualFold(charset, "us-ascii") {
		if decoded, err := decodeCharset(data, charset); err == nil {
			body = decoded
		}
	}

	// 清理无效字符
	body = sanitizeString(body)

	if strings.HasPrefix(contentType, "text/plain") && email.Content == "" {
		email.Content = body
		email.ContentType = "text/plain"
	} else if strings.HasPrefix(contentType, "text/html") && email.ContentHTML == "" {
		email.ContentHTML = body
		if email.ContentType == "" {
			email.ContentType = "text/html"
		}
	}

	// 检查附件标记
	if disp, _, _ := entity.Header.ContentDisposition(); disp == "attachment" {
		email.HasAttachment = true
	}
}

// decodeCharset 将字节从指定字符集解码为UTF-8字符串
func decodeCharset(data []byte, charset string) (string, error) {
	// go-message/charset 已注册常见字符集解码器
	// 使用mime包的WordDecoder处理
	decoder := new(mime.WordDecoder)
	decoded, err := decoder.DecodeHeader(string(data))
	if err != nil {
		return string(data), err
	}
	return decoded, nil
}

// FetchEmails 批量获取邮件
func (p *Net126Provider) FetchEmails(ctx context.Context, since time.Time, limit int) (*SyncResult, error) {
	summaries, err := p.FetchEmailList(ctx, since, limit)
	if err != nil {
		return nil, err
	}

	result := &SyncResult{
		TotalCount: len(summaries),
		Summaries:  make([]EmailSummary, 0, len(summaries)),
		Emails:     make([]Email, 0, len(summaries)),
	}

	for _, s := range summaries {
		result.Summaries = append(result.Summaries, EmailSummary{
			MessageID:     s.MessageID,
			Subject:       s.Subject,
			SenderName:    s.SenderName,
			SenderEmail:   s.SenderEmail,
			ReceivedAt:    s.ReceivedAt,
			HasAttachment: s.HasAttachment,
			Size:          s.Size,
		})
	}

	for _, summary := range summaries {
		email, err := p.FetchEmailDetail(ctx, summary.MessageID)
		if err != nil {
			result.ErrorCount++
			result.LastError = err.Error()
			continue
		}
		result.Emails = append(result.Emails, *email)
		result.SyncedCount++
	}

	return result, nil
}

// Disconnect 断开连接
func (p *Net126Provider) Disconnect() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.client != nil {
		p.client.Logout()
		p.client.Close()
		p.client = nil
	}
	return nil
}

// IsConnected 检查是否已连接
func (p *Net126Provider) IsConnected() bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.client != nil
}

// decodeSubject 解码RFC 2047 MIME编码的主题
func decodeSubject(subject string) string {
	if subject == "" {
		return ""
	}
	dec := new(mime.WordDecoder)
	decoded, err := dec.DecodeHeader(subject)
	if err != nil {
		return sanitizeString(subject)
	}
	return sanitizeString(decoded)
}

// sanitizeString 清理字符串中的无效Unicode字符（如surrogate字符）
func sanitizeString(s string) string {
	// 移除surrogate字符（U+D800 到 U+DFFF）
	var result strings.Builder
	result.Grow(len(s))
	for _, r := range s {
		if r >= 0xD800 && r <= 0xDFFF {
			// 跳过surrogate字符
			continue
		}
		result.WriteRune(r)
	}
	return result.String()
}

func init() {
	Register("126", NewNet126Provider)
}