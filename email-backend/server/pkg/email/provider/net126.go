// Package provider 邮件提供商实现
package provider

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"strings"
	"time"
)

// Net126Provider 网易126邮箱Provider
// 使用原生 IMAP 协议实现
type Net126Provider struct {
	name       string
	server     string
	port       int
	useSSL     bool
	email      string
	credential string
	conn       net.Conn
	timeout    time.Duration
}

// NewNet126Provider 创建126邮箱Provider
func NewNet126Provider(config *ProviderConfig) EmailProvider {
	p := &Net126Provider{
		name:    "126",
		server:  "imap.126.com",
		port:    993,
		useSSL:  true,
		timeout: 30 * time.Second,
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

// Connect 连接邮箱服务器
func (p *Net126Provider) Connect(ctx context.Context, email, credential string) error {
	p.email = email
	p.credential = credential

	addr := fmt.Sprintf("%s:%d", p.server, p.port)

	var err error
	p.conn, err = net.DialTimeout("tcp", addr, p.timeout)
	if err != nil {
		return fmt.Errorf("连接邮箱服务器失败: %w", err)
	}

	if p.useSSL {
		tlsConfig := &tls.Config{
			ServerName:         p.server,
			InsecureSkipVerify: false,
		}
		p.conn = tls.Client(p.conn, tlsConfig)
	}

	p.conn.SetDeadline(time.Now().Add(p.timeout))

	// 读取服务器问候语
	buf := make([]byte, 1024)
	_, err = p.conn.Read(buf)
	if err != nil {
		p.Disconnect()
		return fmt.Errorf("读取服务器问候语失败: %w", err)
	}

	// 发送 LOGIN 命令
	loginCmd := fmt.Sprintf("A001 LOGIN \"%s\" \"%s\"\r\n", email, credential)
	_, err = p.conn.Write([]byte(loginCmd))
	if err != nil {
		p.Disconnect()
		return fmt.Errorf("发送登录命令失败: %w", err)
	}

	// 读取响应
	n, err := p.conn.Read(buf)
	if err != nil {
		p.Disconnect()
		return fmt.Errorf("读取登录响应失败: %w", err)
	}

	resp := string(buf[:n])
	if !strings.Contains(resp, "OK") {
		p.Disconnect()
		return fmt.Errorf("登录失败: %s", resp)
	}

	return nil
}

// TestConnection 测试连接
func (p *Net126Provider) TestConnection(ctx context.Context) (*ConnectionResult, error) {
	if p.conn == nil {
		return &ConnectionResult{
			Success: false,
			Message: "未连接到服务器",
		}, nil
	}

	p.conn.SetDeadline(time.Now().Add(p.timeout))

	// 发送 NOOP 命令
	_, err := p.conn.Write([]byte("A000 NOOP\r\n"))
	if err != nil {
		return &ConnectionResult{
			Success: false,
			Message: fmt.Sprintf("连接测试失败: %v", err),
		}, nil
	}

	// 读取响应
	buf := make([]byte, 1024)
	_, err = p.conn.Read(buf)
	if err != nil {
		return &ConnectionResult{
			Success: false,
			Message: fmt.Sprintf("连接已断开: %v", err),
		}, nil
	}

	return &ConnectionResult{
		Success: true,
		Message: "连接正常",
	}, nil
}

// FetchEmailList 获取邮件列表
func (p *Net126Provider) FetchEmailList(ctx context.Context, since time.Time, limit int) ([]*EmailSummary, error) {
	if p.conn == nil {
		return nil, fmt.Errorf("未连接")
	}

	p.conn.SetDeadline(time.Now().Add(p.timeout))

	// 选择收件箱
	_, err := p.conn.Write([]byte("A002 SELECT INBOX\r\n"))
	if err != nil {
		return nil, fmt.Errorf("选择收件箱失败: %w", err)
	}

	// 读取响应直到完成
	total := 0
	for {
		buf := make([]byte, 1024)
		p.conn.SetReadDeadline(time.Now().Add(p.timeout))
		n, err := p.conn.Read(buf)
		if err != nil {
			return nil, fmt.Errorf("读取响应失败: %w", err)
		}
		resp := string(buf[:n])

		// 解析邮件总数
		if strings.Contains(resp, "EXISTS") {
			parts := strings.Split(resp, "EXISTS")
			if len(parts) > 0 {
				numStr := strings.TrimSpace(parts[0])
				numStr = strings.Trim(numStr, " ")
				numParts := strings.Fields(numStr)
				if len(numParts) > 0 {
					fmt.Sscanf(numParts[len(numParts)-1], "%d", &total)
				}
			}
		}

		if strings.Contains(resp, "A002 OK") {
			break
		}
	}

	if total == 0 {
		return []*EmailSummary{}, nil
	}

	// 搜索指定日期之后的邮件
	sinceStr := since.Format("02-Jan-2006")
	searchCmd := fmt.Sprintf("A003 SEARCH SINCE %s NOT SEEN\r\n", sinceStr)
	_, err = p.conn.Write([]byte(searchCmd))
	if err != nil {
		return nil, fmt.Errorf("搜索邮件失败: %w", err)
	}

	// 读取搜索结果
	buf := make([]byte, 4096)
	n, err := p.conn.Read(buf)
	if err != nil {
		return nil, fmt.Errorf("读取搜索结果失败: %w", err)
	}

	resp := string(buf[:n])
	var ids []int
	fmt.Sscanf(strings.Replace(resp, "SEARCH", "", 1), "%d", &ids)

	if len(ids) == 0 {
		// 如果没有未读，返回最新的
		start := total - limit + 1
		if start < 1 {
			start = 1
		}
		for i := start; i <= total; i++ {
			ids = append(ids, i)
		}
	}

	if len(ids) > limit {
		ids = ids[len(ids)-limit:]
	}

	summaries := make([]*EmailSummary, 0, len(ids))
	for _, id := range ids {
		summary, err := p.fetchEmailSummary(uint32(id))
		if err != nil {
			continue
		}
		summaries = append(summaries, summary)
	}

	return summaries, nil
}

func (p *Net126Provider) fetchEmailSummary(seqNum uint32) (*EmailSummary, error) {
	// 获取邮件头
	cmd := fmt.Sprintf("A004 FETCH %d (ENVELOPE RFC822.SIZE)\r\n", seqNum)
	_, err := p.conn.Write([]byte(cmd))
	if err != nil {
		return nil, err
	}

	// 读取响应
	buf := make([]byte, 4096)
	n, err := p.conn.Read(buf)
	if err != nil {
		return nil, err
	}

	summary := &EmailSummary{}
	resp := string(buf[:n])

	// 简单解析 ENVELOPE
	// 格式: ENVELOPE ("date" "subject" ("from" ("name" "adl" "mailbox" "host")) ...)
	// 这里做简化处理

	// 提取 Subject
	if strings.Contains(resp, "ENVELOPE") {
		envStart := strings.Index(resp, "(")
		envEnd := strings.LastIndex(resp, ")")
		if envStart > 0 && envEnd > envStart {
			env := resp[envStart : envEnd+1]
			summary.MessageID = fmt.Sprintf("<seq-%d>", seqNum)

			// 尝试提取日期
			parts := strings.Split(env, "\"")
			if len(parts) > 2 {
				// 第一个引号内容是日期
				dateStr := parts[1]
				if t, err := time.Parse("02-Jan-2006 15:04:05 -0700", dateStr+" +0000"); err == nil {
					summary.ReceivedAt = t
				}
			}
		}
	}

	return summary, nil
}

// FetchEmailDetail 获取邮件详情
func (p *Net126Provider) FetchEmailDetail(ctx context.Context, messageID string) (*Email, error) {
	if p.conn == nil {
		return nil, fmt.Errorf("未连接")
	}

	// 选择收件箱
	p.conn.Write([]byte("A005 SELECT INBOX\r\n"))

	// 简化实现：获取最新一封邮件
	cmd := fmt.Sprintf("A006 FETCH %s (ENVELOPE BODY[TEXT])\r\n", messageID)
	_, err := p.conn.Write([]byte(cmd))
	if err != nil {
		return nil, err
	}

	email := &Email{
		MessageID: messageID,
	}

	return email, nil
}

// FetchEmails 批量获取邮件
func (p *Net126Provider) FetchEmails(ctx context.Context, since time.Time, limit int) (*SyncResult, error) {
	summaries, err := p.FetchEmailList(ctx, since, limit)
	if err != nil {
		return nil, err
	}

	summaryResults := make([]EmailSummary, 0, len(summaries))
	for _, s := range summaries {
		summaryResults = append(summaryResults, EmailSummary{
			MessageID:     s.MessageID,
			Subject:       s.Subject,
			SenderName:    s.SenderName,
			SenderEmail:   s.SenderEmail,
			ReceivedAt:    s.ReceivedAt,
			HasAttachment: s.HasAttachment,
			Size:          s.Size,
		})
	}

	result := &SyncResult{
		TotalCount: len(summaries),
		Summaries:  summaryResults,
		Emails:     make([]Email, 0, len(summaries)),
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
	if p.conn != nil {
		p.conn.Write([]byte("A999 LOGOUT\r\n"))
		p.conn.Close()
		p.conn = nil
	}
	return nil
}

// IsConnected 检查是否已连接
func (p *Net126Provider) IsConnected() bool {
	return p.conn != nil
}

func init() {
	Register("126", NewNet126Provider)
}

// decodeMIME1 decode RFC 2047 MIME encoded-word
func decodeMIME1(s string) string {
	// Simplified - just remove encoded words
	if strings.HasPrefix(s, "=?") && strings.HasSuffix(s, "?=") {
		parts := strings.Split(s, "?")
		if len(parts) >= 4 {
			return parts[3]
		}
	}
	return s
}

// quoteString for IMAP
func quoteString(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "\"", "\\\"")
	return "\"" + s + "\""
}
