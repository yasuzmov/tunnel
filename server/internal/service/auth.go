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

// CurrentUser 当前登录管理员的公开信息，附带实际生效的左侧菜单。
type CurrentUser struct {
	ID        uint      `json:"id"`
	Username  string    `json:"username"`
	Role      string    `json:"role"`
	Menus     []string  `json:"menus"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// LoginResult 登录成功返回的数据。
type LoginResult struct {
	Token    string       `json:"token"`
	ExpireAt time.Time    `json:"expire_at"`
	User     *CurrentUser `json:"user"`
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
	view, err := s.currentUser(ctx, user)
	if err != nil {
		return nil, err
	}
	return &LoginResult{Token: token, ExpireAt: time.Now().Add(s.ttl), User: view}, nil
}

// Me 按 ID 读取当前管理员。
func (s *AuthService) Me(ctx context.Context, id uint) (*CurrentUser, error) {
	user, err := s.admins.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("读取当前用户: %w", err)
	}
	return s.currentUser(ctx, user)
}

func (s *AuthService) currentUser(ctx context.Context, user *model.AdminUser) (*CurrentUser, error) {
	menus, err := s.effectiveMenus(ctx, user)
	if err != nil {
		return nil, err
	}
	return &CurrentUser{
		ID:        user.ID,
		Username:  user.Username,
		Role:      user.Role,
		Menus:     menus,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}, nil
}

func (s *AuthService) effectiveMenus(ctx context.Context, user *model.AdminUser) ([]string, error) {
	if user.IsSuperAdmin() {
		return EffectiveMenus(user.Role, nil), nil
	}
	stored, err := s.admins.ListMenus(ctx, user.ID)
	if err != nil {
		return nil, err
	}
	return EffectiveMenus(user.Role, stored), nil
}
