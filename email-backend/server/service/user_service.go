// Package service 用户服务层
package service

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"

	"email-backend/server/global"
	"email-backend/server/model"
	emailRequest "email-backend/server/model/request"
	emailResponse "email-backend/server/model/response"
	"email-backend/server/repository"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

const (
	// UserIDLength 业务UserID长度
	UserIDLength = 16
	// TokenExpiry JWT Token有效期
	TokenExpiry = 24 * time.Hour
	// UserIDCharset 可用于UserID的字符集
	UserIDCharset = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"
)

var (
	// ErrUserExists 用户已存在
	ErrUserExists = errors.New("user already exists")
	// ErrInvalidCredentials 无效的凭据
	ErrInvalidCredentials = errors.New("invalid credentials")
	// ErrUserNotFound 用户不存在
	ErrUserNotFound = errors.New("user not found")
)

// Claims JWT Claims
type Claims struct {
	UserID int64  `json:"user_id"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
}

// UserService 用户服务
type UserService struct {
	repo *repository.UserRepository
}

// NewUserService 创建用户服务
func NewUserService(repo *repository.UserRepository) *UserService {
	return &UserService{repo: repo}
}

// Register 用户注册
func (s *UserService) Register(ctx context.Context, req *emailRequest.RegisterRequest) (*model.User, string, error) {
	// 检查邮箱是否已存在
	existing, err := s.repo.FindByEmail(ctx, req.Email)
	if err == nil && existing != nil {
		return nil, "", ErrUserExists
	}

	// 检查用户名是否已存在
	existing, err = s.repo.FindByUsername(ctx, req.Username)
	if err == nil && existing != nil {
		return nil, "", ErrUserExists
	}

	// 生成唯一的16位业务UserID
	userID, err := s.GenerateUniqueUserID(ctx)
	if err != nil {
		return nil, "", fmt.Errorf("generate user id failed: %w", err)
	}

	// 密码哈希
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, "", fmt.Errorf("hash password failed: %w", err)
	}

	user := &model.User{
		UserID:       userID,
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: string(hashedPassword),
	}

	if err := s.repo.Create(ctx, user); err != nil {
		return nil, "", fmt.Errorf("create user failed: %w", err)
	}

	// 生成JWT Token
	token, _, err := s.GenerateToken(user)
	if err != nil {
		return nil, "", fmt.Errorf("generate token failed: %w", err)
	}

	return user, token, nil
}

// Login 用户登录
func (s *UserService) Login(ctx context.Context, req *emailRequest.LoginRequest) (*model.User, string, error) {
	user, err := s.repo.FindByEmail(ctx, req.Email)
	if err != nil {
		return nil, "", ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, "", ErrInvalidCredentials
	}

	token, _, err := s.GenerateToken(user)
	if err != nil {
		return nil, "", fmt.Errorf("generate token failed: %w", err)
	}

	return user, token, nil
}

// GetUserByID 根据ID获取用户
func (s *UserService) GetUserByID(ctx context.Context, id int64) (*model.User, error) {
	return s.repo.FindByID(ctx, id)
}

// GenerateUniqueUserID 生成唯一的16位业务UserID
func (s *UserService) GenerateUniqueUserID(ctx context.Context) (string, error) {
	for i := 0; i < 10; i++ {
		userID := GenerateRandomUserID()
		exists, err := s.repo.UserIDExists(ctx, userID)
		if err != nil {
			return "", err
		}
		if !exists {
			return userID, nil
		}
	}
	return "", errors.New("failed to generate unique user id after 10 attempts")
}

// GenerateRandomUserID 生成随机16位UserID
func GenerateRandomUserID() string {
	result := make([]byte, UserIDLength)
	charsetLen := big.NewInt(int64(len(UserIDCharset)))

	for i := 0; i < UserIDLength; i++ {
		n, _ := rand.Int(rand.Reader, charsetLen)
		result[i] = UserIDCharset[n.Int64()]
	}

	return string(result)
}

// GenerateToken 生成JWT Token
func (s *UserService) GenerateToken(user *model.User) (string, time.Time, error) {
	expiresAt := time.Now().Add(TokenExpiry)

	claims := &Claims{
		UserID: user.ID,
		Email:  user.Email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "email-agent",
			Subject:   user.UserID,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(global.Config().Security.JWTSecret))
	if err != nil {
		return "", time.Time{}, err
	}

	return tokenString, expiresAt, nil
}

// ValidateToken 验证JWT Token
func (s *UserService) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(global.Config().Security.JWTSecret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

// ToAuthResponse 转换为认证响应
func (s *UserService) ToAuthResponse(user *model.User, token string, expiresAt time.Time) *emailResponse.AuthResponse {
	return &emailResponse.AuthResponse{
		Token:     token,
		ExpiresAt: expiresAt,
		User:      s.ToUserResponse(user),
	}
}

// IsEmailExists 检查邮箱是否已存在
func (s *UserService) IsEmailExists(ctx context.Context, email string) bool {
	_, err := s.repo.FindByEmail(ctx, email)
	return err == nil
}

// NormalizeEmail 标准化邮箱（小写）
func NormalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

// UpdateProfile 更新用户资料
func (s *UserService) UpdateProfile(ctx context.Context, userID int64, username string) (*model.User, error) {
	user, err := s.repo.FindByID(ctx, userID)
	if err != nil {
		return nil, ErrUserNotFound
	}

	// 检查用户名是否被其他用户占用
	existing, _ := s.repo.FindByUsername(ctx, username)
	if existing != nil && existing.ID != userID {
		return nil, errors.New("username already exists")
	}

	user.Username = username
	if err := s.repo.Update(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

// ChangePassword 修改密码
func (s *UserService) ChangePassword(ctx context.Context, userID int64, oldPassword, newPassword string) error {
	user, err := s.repo.FindByID(ctx, userID)
	if err != nil {
		return ErrUserNotFound
	}

	// 验证旧密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(oldPassword)); err != nil {
		return errors.New("invalid old password")
	}

	// 哈希新密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	user.PasswordHash = string(hashedPassword)
	return s.repo.Update(ctx, user)
}

// UpdateAvatar 更新用户头像
func (s *UserService) UpdateAvatar(ctx context.Context, userID int64, avatarURL string) error {
	user, err := s.repo.FindByID(ctx, userID)
	if err != nil {
		return ErrUserNotFound
	}

	user.AvatarURL = avatarURL
	return s.repo.Update(ctx, user)
}

// ToUserResponse 转换为用户响应（包含头像）
func (s *UserService) ToUserResponse(user *model.User) emailResponse.UserResponse {
	return emailResponse.UserResponse{
		ID:        user.ID,
		UserID:    user.UserID,
		Username:  user.Username,
		Email:     user.Email,
		AvatarURL: user.AvatarURL,
		CreatedAt: user.CreatedAt,
	}
}
