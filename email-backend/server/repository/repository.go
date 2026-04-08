// Package repository 数据访问层
package repository

import (
	"context"

	"email-backend/server/model"
	emailRequest "email-backend/server/model/request"

	"gorm.io/gorm"
)

// EmailRepository 邮件数据访问
type EmailRepository struct {
	db *gorm.DB
}

// NewEmailRepository 创建邮件Repository
func NewEmailRepository(db *gorm.DB) *EmailRepository {
	return &EmailRepository{db: db}
}

// FindByID 根据ID查询
func (r *EmailRepository) FindByID(ctx context.Context, id int64) (*model.Email, error) {
	var email model.Email
	err := r.db.WithContext(ctx).
		Where("id = ?", id).
		First(&email).Error
	if err != nil {
		return nil, err
	}
	return &email, nil
}

// List 分页查询
func (r *EmailRepository) List(ctx context.Context, req *emailRequest.ListRequest) ([]*model.Email, int64, error) {
	var emails []*model.Email
	var total int64

	query := r.db.WithContext(ctx).Model(&model.Email{})

	// 条件过滤
	if req.UserID > 0 {
		query = query.Where("user_id = ?", req.UserID)
	}
	if req.Category != "" {
		query = query.Where("category = ?", req.Category)
	}
	if req.Status != "" {
		query = query.Where("status = ?", req.Status)
	}

	// 统计总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (req.Page - 1) * req.PageSize
	if err := query.
		Offset(offset).
		Limit(req.PageSize).
		Order("received_at DESC").
		Find(&emails).Error; err != nil {
		return nil, 0, err
	}

	return emails, total, nil
}

// Create 创建
func (r *EmailRepository) Create(ctx context.Context, email *model.Email) error {
	return r.db.WithContext(ctx).Create(email).Error
}

// Update 更新
func (r *EmailRepository) Update(ctx context.Context, email *model.Email) error {
	return r.db.WithContext(ctx).Save(email).Error
}

// Delete 删除
func (r *EmailRepository) Delete(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Delete(&model.Email{}, id).Error
}

// AccountRepository 账户数据访问
type AccountRepository struct {
	db *gorm.DB
}

// NewAccountRepository 创建账户Repository
func NewAccountRepository(db *gorm.DB) *AccountRepository {
	return &AccountRepository{db: db}
}

// FindByID 根据ID查询
func (r *AccountRepository) FindByID(ctx context.Context, id int64) (*model.EmailAccount, error) {
	var account model.EmailAccount
	err := r.db.WithContext(ctx).
		Where("id = ?", id).
		First(&account).Error
	if err != nil {
		return nil, err
	}
	return &account, nil
}

// ListByUserID 根据用户ID查询
func (r *AccountRepository) ListByUserID(ctx context.Context, userID int64) ([]*model.EmailAccount, error) {
	var accounts []*model.EmailAccount
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Find(&accounts).Error
	return accounts, err
}

// Create 创建
func (r *AccountRepository) Create(ctx context.Context, account *model.EmailAccount) error {
	return r.db.WithContext(ctx).Create(account).Error
}

// Delete 删除
func (r *AccountRepository) Delete(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Delete(&model.EmailAccount{}, id).Error
}