// Package service 业务逻辑层
package service

import (
	"context"

	"email-backend/server/model"
	emailRequest "email-backend/server/model/request"
	"email-backend/server/repository"
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
func (s *EmailService) ClassifyEmail(ctx context.Context, id int64, category, priority string, confidence float64) error {
	email, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}

	email.Category = category
	email.Priority = priority
	email.Confidence = confidence
	email.IsProcessed = true

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
func (s *AccountService) Create(ctx context.Context, account *model.EmailAccount) error {
	return s.repo.Create(ctx, account)
}

// Delete 删除账户
func (s *AccountService) Delete(ctx context.Context, id int64) error {
	return s.repo.Delete(ctx, id)
}