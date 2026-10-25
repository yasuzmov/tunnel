package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"gorm.io/gorm"

	"tunnel-server/internal/model"
	"tunnel-server/internal/pkg/crypto"
	"tunnel-server/internal/repository"
)

// AdminService 管理后台账号管理、角色与域名授权。
type AdminService struct {
	admins repository.AdminUserRepository
	pool   repository.DomainPoolRepository
}

// NewAdminService 创建管理员服务。
func NewAdminService(admins repository.AdminUserRepository, pool repository.DomainPoolRepository) *AdminService {
	return &AdminService{admins: admins, pool: pool}
}

// AdminView 管理员列表项，附带被授权的域名。
// 超级管理员不返回具体授权，前端以「全部域名」展示。
type AdminView struct {
	ID          uint      `json:"id"`
	Username    string    `json:"username"`
	Role        string    `json:"role"`
	CreatedAt   time.Time `json:"created_at"`
	DomainIDs   []uint    `json:"domain_ids"`
	DomainNames []string  `json:"domain_names"`
}

// CreateAdminInput 新增管理员。
type CreateAdminInput struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Role     string `json:"role"`
}

// UpdateAdminInput 修改登录名与角色；Role 为空表示保持不变。
type UpdateAdminInput struct {
	Username string `json:"username"`
	Role     string `json:"role"`
}

// List 列出全部管理员。
func (s *AdminService) List(ctx context.Context) ([]AdminView, error) {
	users, err := s.admins.List(ctx)
	if err != nil {
		return nil, err
	}
	domains, err := s.pool.List(ctx, "")
	if err != nil {
		return nil, err
	}
	nameByID := make(map[uint]string, len(domains))
	for _, d := range domains {
		nameByID[d.ID] = d.Domain
	}
	views := make([]AdminView, 0, len(users))
	for _, u := range users {
		view, err := s.viewOf(ctx, u, nameByID)
		if err != nil {
			return nil, err
		}
		views = append(views, view)
	}
	return views, nil
}

// Get 读取单个管理员。
func (s *AdminService) Get(ctx context.Context, id uint) (*AdminView, error) {
	u, err := s.find(ctx, id)
	if err != nil {
		return nil, err
	}
	domains, err := s.pool.List(ctx, "")
	if err != nil {
		return nil, err
	}
	nameByID := make(map[uint]string, len(domains))
	for _, d := range domains {
		nameByID[d.ID] = d.Domain
	}
	view, err := s.viewOf(ctx, *u, nameByID)
	if err != nil {
		return nil, err
	}
	return &view, nil
}

// Create 新增管理员账号。
func (s *AdminService) Create(ctx context.Context, in CreateAdminInput) (*AdminView, error) {
	username := strings.TrimSpace(in.Username)
	if err := validateUsername(username); err != nil {
		return nil, err
	}
	if err := validatePassword(in.Password); err != nil {
		return nil, err
	}
	role, err := normalizeRole(in.Role, model.RoleAdmin)
	if err != nil {
		return nil, err
	}
	if _, err := s.admins.FindByUsername(ctx, username); err == nil {
		return nil, ErrAdminTaken
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	hashed, err := crypto.HashPassword(in.Password)
	if err != nil {
		return nil, err
	}
	u := &model.AdminUser{Username: username, PasswordHash: hashed, Role: role}
	if err := s.admins.Create(ctx, u); err != nil {
		return nil, err
	}
	return s.Get(ctx, u.ID)
}

// Update 修改登录名与角色，Role 为空时保持原角色。
func (s *AdminService) Update(ctx context.Context, id uint, in UpdateAdminInput) (*AdminView, error) {
	u, err := s.find(ctx, id)
	if err != nil {
		return nil, err
	}
	username := strings.TrimSpace(in.Username)
	if err := validateUsername(username); err != nil {
		return nil, err
	}
	role := u.Role
	if strings.TrimSpace(in.Role) != "" {
		role, err = normalizeRole(in.Role, u.Role)
		if err != nil {
			return nil, err
		}
	}
	if username != u.Username {
		if found, err := s.admins.FindByUsername(ctx, username); err == nil && found.ID != id {
			return nil, ErrAdminTaken
		} else if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
	}
	if u.IsSuperAdmin() && role != model.RoleSuperAdmin {
		if err := s.ensureAnotherSuper(ctx); err != nil {
			return nil, err
		}
	}
	if err := s.admins.UpdateProfile(ctx, id, username, role); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return s.Get(ctx, id)
}

// ResetPassword 管理员重置他人口令，不需要原口令。
func (s *AdminService) ResetPassword(ctx context.Context, id uint, password string) error {
	if _, err := s.find(ctx, id); err != nil {
		return err
	}
	if err := validatePassword(password); err != nil {
		return err
	}
	hashed, err := crypto.HashPassword(password)
	if err != nil {
		return err
	}
	if err := s.admins.UpdatePassword(ctx, id, hashed); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrNotFound
		}
		return err
	}
	return nil
}

// SetDomains 覆盖式配置某管理员可使用的域名。
func (s *AdminService) SetDomains(ctx context.Context, id uint, domainIDs []uint) error {
	u, err := s.find(ctx, id)
	if err != nil {
		return err
	}
	if u.IsSuperAdmin() {
		return fmt.Errorf("%w: 超级管理员默认拥有全部域名", ErrBadInput)
	}
	domains, err := s.pool.List(ctx, "")
	if err != nil {
		return err
	}
	known := make(map[uint]struct{}, len(domains))
	for _, d := range domains {
		known[d.ID] = struct{}{}
	}
	seen := make(map[uint]struct{}, len(domainIDs))
	clean := make([]uint, 0, len(domainIDs))
	for _, did := range domainIDs {
		if _, ok := known[did]; !ok {
			return fmt.Errorf("%w: 域名不存在", ErrBadInput)
		}
		if _, dup := seen[did]; dup {
			continue
		}
		seen[did] = struct{}{}
		clean = append(clean, did)
	}
	return s.pool.SetAdminDomains(ctx, id, clean)
}

// Delete 删除管理员；禁止删除自己，且必须保留至少一个超级管理员。
func (s *AdminService) Delete(ctx context.Context, id, operatorID uint) error {
	if id == operatorID {
		return fmt.Errorf("%w: 不能删除当前登录账号", ErrBadInput)
	}
	u, err := s.find(ctx, id)
	if err != nil {
		return err
	}
	if u.IsSuperAdmin() {
		if err := s.ensureAnotherSuper(ctx); err != nil {
			return err
		}
	}
	if err := s.pool.SetAdminDomains(ctx, id, nil); err != nil {
		return err
	}
	if err := s.admins.Delete(ctx, id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrNotFound
		}
		return err
	}
	return nil
}

// IsSuper 判断当前管理员是否为超级管理员。
func (s *AdminService) IsSuper(ctx context.Context, id uint) (bool, error) {
	u, err := s.find(ctx, id)
	if err != nil {
		return false, err
	}
	return u.IsSuperAdmin(), nil
}

func (s *AdminService) find(ctx context.Context, id uint) (*model.AdminUser, error) {
	u, err := s.admins.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return u, nil
}

func (s *AdminService) ensureAnotherSuper(ctx context.Context) error {
	n, err := s.admins.CountByRole(ctx, model.RoleSuperAdmin)
	if err != nil {
		return err
	}
	if n <= 1 {
		return ErrLastSuperAdmin
	}
	return nil
}

func (s *AdminService) viewOf(ctx context.Context, u model.AdminUser, nameByID map[uint]string) (AdminView, error) {
	view := AdminView{
		ID:          u.ID,
		Username:    u.Username,
		Role:        u.Role,
		CreatedAt:   u.CreatedAt,
		DomainIDs:   []uint{},
		DomainNames: []string{},
	}
	if u.IsSuperAdmin() {
		return view, nil
	}
	ids, err := s.pool.ListAdminDomainIDs(ctx, u.ID)
	if err != nil {
		return view, err
	}
	for _, id := range ids {
		view.DomainIDs = append(view.DomainIDs, id)
		if name, ok := nameByID[id]; ok {
			view.DomainNames = append(view.DomainNames, name)
		}
	}
	return view, nil
}

func validateUsername(username string) error {
	if username == "" {
		return fmt.Errorf("%w: 用户名不能为空", ErrBadInput)
	}
	if utf8.RuneCountInString(username) > 64 {
		return fmt.Errorf("%w: 用户名不能超过 64 个字符", ErrBadInput)
	}
	if strings.ContainsAny(username, " \t\r\n") {
		return fmt.Errorf("%w: 用户名不能包含空白字符", ErrBadInput)
	}
	return nil
}

func normalizeRole(raw, fallback string) (string, error) {
	role := strings.TrimSpace(raw)
	if role == "" {
		role = fallback
	}
	switch role {
	case model.RoleSuperAdmin, model.RoleAdmin:
		return role, nil
	default:
		return "", fmt.Errorf("%w: 未知的管理员角色", ErrBadInput)
	}
}
