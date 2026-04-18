// Package service 业务逻辑层
package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"email-backend/server/global"
	"email-backend/server/model"
	emailRequest "email-backend/server/model/request"
	"email-backend/server/pkg/email/provider"
	"email-backend/server/repository"
)

// 错误定义
var (
	ErrInvalidCredential = errors.New("无效的凭证")
	ErrProviderNotFound  = errors.New("不支持的邮件提供商")
	ErrAccountNotFound   = errors.New("账户不存在")
	ErrEncryptFailed     = errors.New("加密失败")
	ErrDecryptFailed     = errors.New("解密失败")
)

// EmailService 邮件服务
type EmailService struct {
	repo *repository.EmailRepository
}

// NewEmailService 创建邮件服务
func NewEmailService(repo *repository.EmailRepository) *EmailService {
	return &EmailService{repo: repo}
}

// GetByID 根据ID获取邮件
func (s *EmailService) GetByID(ctx context.Context, id int64) (*model.Email, error) {
	return s.repo.FindByID(ctx, id)
}

// List 获取邮件列表
func (s *EmailService) List(ctx context.Context, req *emailRequest.ListRequest) ([]*model.Email, int64, error) {
	return s.repo.List(ctx, req)
}

// Update 更新邮件
func (s *EmailService) Update(ctx context.Context, email *model.Email) error {
	return s.repo.Update(ctx, email)
}

// Create 创建邮件
func (s *EmailService) Create(ctx context.Context, email *model.Email) error {
	return s.repo.Create(ctx, email)
}

// Delete 删除邮件
func (s *EmailService) Delete(ctx context.Context, id int64) error {
	return s.repo.Delete(ctx, id)
}

// ClassifyEmail 分类邮件
func (s *EmailService) ClassifyEmail(ctx context.Context, id int64, category, priority string, confidence float64, reasoning string) error {
	email, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}

	email.Category = category
	email.Priority = priority
	email.Confidence = confidence
	email.Reasoning = reasoning
	email.IsProcessed = true

	now := time.Now()
	email.ProcessedAt = &now

	return s.repo.Update(ctx, email)
}

// MarkAsRead 标记邮件为已读
func (s *EmailService) MarkAsRead(ctx context.Context, id int64) error {
	email, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}

	email.Status = "read"
	return s.repo.Update(ctx, email)
}

// ArchiveEmail 归档邮件
func (s *EmailService) ArchiveEmail(ctx context.Context, id int64) error {
	email, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}

	email.Status = "archived"
	return s.repo.Update(ctx, email)
}

// UpdateStatus 更新邮件状态
func (s *EmailService) UpdateStatus(ctx context.Context, id int64, status string) error {
	email, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}

	email.Status = status
	return s.repo.Update(ctx, email)
}

// AccountService 账户服务
type AccountService struct {
	repo *repository.AccountRepository
}

// NewAccountService 创建账户服务
func NewAccountService(repo *repository.AccountRepository) *AccountService {
	return &AccountService{repo: repo}
}

// GetByID 根据ID获取账户
func (s *AccountService) GetByID(ctx context.Context, id int64) (*model.EmailAccount, error) {
	return s.repo.FindByID(ctx, id)
}

// List 获取账户列表
func (s *AccountService) List(ctx context.Context, userID int64) ([]*model.EmailAccount, error) {
	return s.repo.ListByUserID(ctx, userID)
}

// Create 创建账户
func (s *AccountService) Create(ctx context.Context, account *model.EmailAccount, credential string) error {
	// 加密凭证
	encrypted, iv, err := global.Encryptor().Encrypt(credential)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrEncryptFailed, err)
	}

	account.EncryptedCredential = encrypted
	account.CredentialIV = iv

	return s.repo.Create(ctx, account)
}

// Delete 删除账户
func (s *AccountService) Delete(ctx context.Context, id int64) error {
	return s.repo.Delete(ctx, id)
}

// TestConnection 测试账户连接
func (s *AccountService) TestConnection(ctx context.Context, id int64) (*provider.ConnectionResult, error) {
	// 获取账户信息
	account, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, ErrAccountNotFound
	}

	// 解密凭证
	credential, err := global.Encryptor().Decrypt(
		account.EncryptedCredential,
		account.CredentialIV,
	)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrDecryptFailed, err)
	}

	// 创建Provider
	emailProvider, ok := provider.Create(account.Provider, nil)
	if !ok {
		return nil, ErrProviderNotFound
	}

	// 连接并测试
	err = emailProvider.Connect(ctx, account.AccountEmail, credential)
	if err != nil {
		return &provider.ConnectionResult{
			Success: false,
			Message: fmt.Sprintf("连接失败: %v", err),
		}, nil
	}
	defer emailProvider.Disconnect()

	return emailProvider.TestConnection(ctx)
}

// GetDecryptedCredential 获取解密的凭证
func (s *AccountService) GetDecryptedCredential(ctx context.Context, id int64) (string, error) {
	account, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return "", ErrAccountNotFound
	}

	return global.Encryptor().Decrypt(
		account.EncryptedCredential,
		account.CredentialIV,
	)
}
