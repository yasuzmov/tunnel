package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	"tunnel-server/internal/model"
	"tunnel-server/internal/pkg/crypto"
	"tunnel-server/internal/pkg/jwtutil"
	"tunnel-server/internal/repository"
)

// AuthService 管理员登录与身份查询。
type AuthService struct {
	admins repository.AdminUserRepository
	secret string
	ttl    time.Duration
}

// NewAuthService 创建认证服务。
func NewAuthService(admins repository.AdminUserRepository, secret string, ttl time.Duration) *AuthService {
	return &AuthService{admins: admins, secret: secret, ttl: ttl}
}

// LoginResult 登录成功返回的数据。
type LoginResult struct {
	Token    string          `json:"token"`
	ExpireAt time.Time       `json:"expire_at"`
	User     *model.AdminUser `json:"user"`
}

// Login 校验口令并签发 JWT。用户不存在与密码错误返回同一错误，避免枚举账号。
func (s *AuthService) Login(ctx context.Context, username, password string) (*LoginResult, error) {
	user, err := s.admins.FindByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrInvalidCredential
		}
		return nil, err
	}
	if !crypto.CheckPassword(user.PasswordHash, password) {
		return nil, ErrInvalidCredential
	}
	token, err := jwtutil.Issue(s.secret, user.ID, user.Username, s.ttl)
	if err != nil {
		return nil, err
	}
	return &LoginResult{Token: token, ExpireAt: time.Now().Add(s.ttl), User: user}, nil
}

// Me 按 ID 读取当前管理员。
func (s *AuthService) Me(ctx context.Context, id uint) (*model.AdminUser, error) {
	user, err := s.admins.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("读取当前用户: %w", err)
	}
	return user, nil
}
