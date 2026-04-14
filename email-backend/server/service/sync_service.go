// Package service 业务逻辑层 - 同步服务
package service

import (
	"context"
	"fmt"
	"time"

	"email-backend/server/global"
	"email-backend/server/model"
	"email-backend/server/pkg/email/provider"
	"email-backend/server/repository"
)

// SyncService 同步服务
type SyncService struct {
	accountRepo *repository.AccountRepository
	emailRepo   *repository.EmailRepository
}

// NewSyncService 创建同步服务
func NewSyncService(accountRepo *repository.AccountRepository, emailRepo *repository.EmailRepository) *SyncService {
	return &SyncService{
		accountRepo: accountRepo,
		emailRepo:   emailRepo,
	}
}

// SyncRequest 同步请求
type SyncRequest struct {
	AccountID int64 `json:"account_id,omitempty"` // 可选，不传则同步所有账户
}

// SyncResult 同步结果
type SyncResult struct {
	AccountID   int64     `json:"account_id"`
	AccountEmail string   `json:"account_email"`
	Success     bool      `json:"success"`
	Message     string    `json:"message"`
	TotalCount  int       `json:"total_count"`
	SyncedCount int       `json:"synced_count"`
	ErrorCount  int       `json:"error_count"`
	SyncedAt    time.Time `json:"synced_at"`
}

// SyncAll 同步所有账户
func (s *SyncService) SyncAll(ctx context.Context, userID int64) ([]*SyncResult, error) {
	accounts, err := s.accountRepo.ListByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("获取账户列表失败: %w", err)
	}

	results := make([]*SyncResult, 0, len(accounts))
	for _, account := range accounts {
		if !account.SyncEnabled {
			continue
		}
		result := s.SyncAccount(ctx, account.ID)
		results = append(results, result)
	}

	return results, nil
}

// SyncAccount 同步单个账户
func (s *SyncService) SyncAccount(ctx context.Context, accountID int64) *SyncResult {
	result := &SyncResult{
		AccountID: accountID,
		SyncedAt:  time.Now(),
	}

	// 获取账户信息
	account, err := s.accountRepo.FindByID(ctx, accountID)
	if err != nil {
		result.Success = false
		result.Message = fmt.Sprintf("账户不存在: %v", err)
		return result
	}

	result.AccountEmail = account.AccountEmail

	// 解密凭证
	credential, err := global.Encryptor().Decrypt(
		account.EncryptedCredential,
		account.CredentialIV,
	)
	if err != nil {
		result.Success = false
		result.Message = fmt.Sprintf("解密凭证失败: %v", err)
		return result
	}

	// 创建Provider
	emailProvider, ok := provider.Create(account.Provider, nil)
	if !ok {
		result.Success = false
		result.Message = fmt.Sprintf("不支持的邮件提供商: %s", account.Provider)
		return result
	}
	defer emailProvider.Disconnect()

	// 连接邮箱
	err = emailProvider.Connect(ctx, account.AccountEmail, credential)
	if err != nil {
		result.Success = false
		result.Message = fmt.Sprintf("连接邮箱失败: %v", err)
		return result
	}

	// 计算同步起始时间
	since := time.Now().AddDate(0, 0, -7) // 默认同步最近7天
	if account.LastSyncAt != nil {
		since = *account.LastSyncAt
	}

	// 获取邮件
	syncResult, err := emailProvider.FetchEmails(ctx, since, 100)
	if err != nil {
		result.Success = false
		result.Message = fmt.Sprintf("获取邮件失败: %v", err)
		return result
	}

	result.TotalCount = syncResult.TotalCount
	result.SyncedCount = syncResult.SyncedCount
	result.ErrorCount = syncResult.ErrorCount

	// 保存邮件到数据库
	for _, email := range syncResult.Emails {
		emailModel := &model.Email{
			MessageID:     email.MessageID,
			UserID:        account.UserID,
			AccountID:     account.ID,
			SenderName:    email.SenderName,
			SenderEmail:   email.SenderEmail,
			Subject:       email.Subject,
			Content:       email.Content,
			ContentHTML:   email.ContentHTML,
			ContentType:   email.ContentType,
			HasAttachment: email.HasAttachment,
			ReceivedAt:    email.ReceivedAt,
			Category:      "unclassified",
			Priority:      "medium",
			Status:        "unread",
		}

		// 检查是否已存在
		existing, _ := s.emailRepo.FindByMessageID(ctx, email.MessageID)
		if existing == nil {
			if err := s.emailRepo.Create(ctx, emailModel); err != nil {
				result.ErrorCount++
				continue
			}
		}
	}

	// 更新账户最后同步时间
	now := time.Now()
	account.LastSyncAt = &now
	s.accountRepo.Update(ctx, account)

	result.Success = true
	result.Message = fmt.Sprintf("同步成功，获取 %d 封邮件", result.SyncedCount)

	return result
}

// GetSyncStatus 获取同步状态
func (s *SyncService) GetSyncStatus(ctx context.Context, userID int64) (map[string]interface{}, error) {
	accounts, err := s.accountRepo.ListByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	status := make(map[string]interface{})
	status["accounts"] = accounts

	var lastSync *time.Time
	for _, account := range accounts {
		if account.LastSyncAt != nil {
			if lastSync == nil || account.LastSyncAt.After(*lastSync) {
				lastSync = account.LastSyncAt
			}
		}
	}

	if lastSync != nil {
		status["last_sync"] = lastSync.Format(time.RFC3339)
	} else {
		status["last_sync"] = nil
	}

	return status, nil
}
