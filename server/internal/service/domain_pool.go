package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"

	"tunnel-server/internal/model"
	"tunnel-server/internal/pkg/validator"
	"tunnel-server/internal/repository"
)

// DomainPoolService 维护统一域名池。只有池中的域名才允许绑定到穿透记录。
// 域名统一为「前缀 + 固定主域名后缀」，并按创建者隔离：
// 普通管理员只能看到、修改、删除、绑定自己创建的域名；超级管理员可见全部。
type DomainPoolService struct {
	pool   repository.DomainPoolRepository
	bound  repository.DomainRepository
	admins repository.AdminUserRepository
	base   string
}

// DomainPoolDeps 构造依赖。
type DomainPoolDeps struct {
	Pool   repository.DomainPoolRepository
	Bound  repository.DomainRepository
	Admins repository.AdminUserRepository
	// BaseDomain 固定的主域名后缀，例如 fish.czypro.cn；为空时禁止新增与修改前缀。
	BaseDomain string
}

// NewDomainPoolService 创建域名池服务。
func NewDomainPoolService(d DomainPoolDeps) *DomainPoolService {
	return &DomainPoolService{
		pool:   d.Pool,
		bound:  d.Bound,
		admins: d.Admins,
		base:   strings.ToLower(strings.Trim(strings.TrimSpace(d.BaseDomain), ".")),
	}
}

// DomainPoolView 域名池条目与当前绑定情况。
type DomainPoolView struct {
	ID     uint   `json:"id"`
	Domain string `json:"domain"`
	// Name 域名前缀；历史上不符合「前缀.主域名」格式的域名为空。
	Name      string    `json:"name"`
	Remark    string    `json:"remark"`
	CreatedAt time.Time `json:"created_at"`
	CreatedBy uint      `json:"created_by"`
	// CreatorName 创建者登录名，仅超级管理员查询时填充。
	CreatorName     string `json:"creator_name,omitempty"`
	BoundTunnelID   uint   `json:"bound_tunnel_id"`
	BoundTunnelName string `json:"bound_tunnel_name"`
	HTTPSEnabled    bool   `json:"https_enabled"`
	CertStatus      string `json:"cert_status"`
	NginxStatus     string `json:"nginx_status"`
}

// DomainPoolConfig 前端创建域名时展示的固定后缀。
type DomainPoolConfig struct {
	BaseDomain string `json:"base_domain"`
}

// CreatePoolInput 新增域名池条目：只需填写前缀 name，后缀固定为主域名。
// 兼容旧接口的 domain 字段：必须是「前缀.主域名」形式。
type CreatePoolInput struct {
	Name   string `json:"name"`
	Domain string `json:"domain"`
	Remark string `json:"remark"`
}

// UpdatePoolInput 修改域名前缀与备注；name、domain 都为空时只改备注。
type UpdatePoolInput struct {
	Name   string `json:"name"`
	Domain string `json:"domain"`
	Remark string `json:"remark"`
}

// Config 返回固定后缀。
func (s *DomainPoolService) Config() DomainPoolConfig {
	return DomainPoolConfig{BaseDomain: s.base}
}

// Viewer 解析当前管理员的数据可见范围。
func (s *DomainPoolService) Viewer(ctx context.Context, adminID uint) (Viewer, error) {
	return resolveViewer(ctx, s.admins, adminID)
}

// List 列出当前管理员可见的域名与绑定情况。
func (s *DomainPoolService) List(ctx context.Context, adminID uint) ([]DomainPoolView, error) {
	v, err := s.Viewer(ctx, adminID)
	if err != nil {
		return nil, err
	}
	domains, err := s.pool.List(ctx, "", v.ownerFilter())
	if err != nil {
		return nil, err
	}
	bindings, err := s.pool.ListBindings(ctx)
	if err != nil {
		return nil, err
	}
	byID := make(map[uint]repository.DomainBinding, len(bindings))
	for _, b := range bindings {
		byID[b.DomainID] = b
	}
	var names map[uint]string
	if v.Super {
		names = s.adminNames(ctx)
	}
	views := make([]DomainPoolView, 0, len(domains))
	for _, d := range domains {
		view := s.viewOf(d)
		view.CreatorName = names[d.CreatedBy]
		if b, ok := byID[d.ID]; ok {
			view.BoundTunnelID = b.TunnelID
			view.BoundTunnelName = b.TunnelName
			view.HTTPSEnabled = b.HTTPSEnabled
			view.CertStatus = b.CertStatus
			view.NginxStatus = b.NginxStatus
		}
		views = append(views, view)
	}
	return views, nil
}

// Get 读取单个域名；他人或历史域名对普通管理员返回 ErrNotFound。
func (s *DomainPoolService) Get(ctx context.Context, adminID, id uint) (*DomainPoolView, error) {
	d, err := s.findVisible(ctx, adminID, id)
	if err != nil {
		return nil, err
	}
	view := s.viewOf(*d)
	return &view, nil
}

// Create 按「前缀.主域名」写入域名池，并记录创建者。
func (s *DomainPoolService) Create(ctx context.Context, adminID uint, in CreatePoolInput) (*DomainPoolView, error) {
	if _, err := s.Viewer(ctx, adminID); err != nil {
		return nil, err
	}
	host, err := s.resolveHost(in.Name, in.Domain)
	if err != nil {
		return nil, err
	}
	if host == "" {
		return nil, fmt.Errorf("%w: 请填写域名前缀", ErrBadInput)
	}
	remark, err := validateRemark(in.Remark)
	if err != nil {
		return nil, err
	}
	if err := s.ensureHostFree(ctx, host, 0); err != nil {
		return nil, err
	}
	d := &model.Domain{Domain: host, Remark: remark, CreatedBy: adminID}
	if err := s.pool.Create(ctx, d); err != nil {
		return nil, err
	}
	view := s.viewOf(*d)
	return &view, nil
}

// Update 修改前缀与备注。已绑定到穿透记录的域名不允许改前缀，避免 nginx 配置与绑定失效。
func (s *DomainPoolService) Update(ctx context.Context, adminID, id uint, in UpdatePoolInput) error {
	d, err := s.findVisible(ctx, adminID, id)
	if err != nil {
		return err
	}
	remark, err := validateRemark(in.Remark)
	if err != nil {
		return err
	}
	host, err := s.resolveHost(in.Name, in.Domain)
	if err != nil {
		return err
	}
	if host == "" {
		host = d.Domain
	}
	if host != d.Domain {
		if _, err := s.bound.FindByDomain(ctx, d.Domain); err == nil {
			return fmt.Errorf("%w: 该域名已绑定到穿透记录，请先解绑再修改前缀", ErrBadInput)
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		if err := s.ensureHostFree(ctx, host, d.ID); err != nil {
			return err
		}
	}
	if err := s.pool.Update(ctx, id, host, remark); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrNotFound
		}
		return err
	}
	return nil
}

// Delete 删除域名池条目；已绑定到穿透记录的域名必须先解绑。
func (s *DomainPoolService) Delete(ctx context.Context, adminID, id uint) error {
	d, err := s.findVisible(ctx, adminID, id)
	if err != nil {
		return err
	}
	if _, err := s.bound.FindByDomain(ctx, d.Domain); err == nil {
		return ErrDomainBound
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	if err := s.pool.Delete(ctx, id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrNotFound
		}
		return err
	}
	return nil
}

// findVisible 读取域名并校验归属；无权或不存在都返回 ErrNotFound，避免泄露他人域名是否存在。
func (s *DomainPoolService) findVisible(ctx context.Context, adminID, id uint) (*model.Domain, error) {
	v, err := s.Viewer(ctx, adminID)
	if err != nil {
		return nil, err
	}
	d, err := s.pool.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	if !v.Owns(d.CreatedBy) {
		return nil, ErrNotFound
	}
	return d, nil
}

// resolveHost 把前缀拼成完整域名；兼容传入完整域名，但必须是主域名下的一级子域名。
// name、domain 都为空时返回空串，由调用方决定是否允许。
func (s *DomainPoolService) resolveHost(name, domain string) (string, error) {
	name = strings.TrimSpace(name)
	domain = strings.TrimSpace(domain)
	if name == "" && domain == "" {
		return "", nil
	}
	if s.base == "" {
		return "", fmt.Errorf("%w: 服务端未配置主域名后缀（admin.public_url / admin.domain_suffix）", ErrBadInput)
	}
	if name == "" {
		host, err := validator.ValidateDomain(domain)
		if err != nil {
			return "", fmt.Errorf("%w: %s", ErrBadInput, err.Error())
		}
		name = validator.SubdomainLabel(host, s.base)
		if name == "" {
			return "", fmt.Errorf("%w: 域名必须是 *.%s 形式，只能修改前缀", ErrBadInput, s.base)
		}
	}
	label, err := validator.ValidateLabel(name)
	if err != nil {
		return "", fmt.Errorf("%w: %s", ErrBadInput, err.Error())
	}
	host := label + "." + s.base
	if len(host) > 253 {
		return "", fmt.Errorf("%w: 域名过长", ErrBadInput)
	}
	return host, nil
}

// ensureHostFree 域名在池中（不论属于谁）或已被穿透记录绑定都视为占用；提示中不透露占用者。
func (s *DomainPoolService) ensureHostFree(ctx context.Context, host string, selfID uint) error {
	if found, err := s.pool.FindByDomain(ctx, host); err == nil {
		if found.ID != selfID {
			return ErrDomainPoolTaken
		}
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	if _, err := s.bound.FindByDomain(ctx, host); err == nil {
		return ErrDomainPoolTaken
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	return nil
}

func (s *DomainPoolService) viewOf(d model.Domain) DomainPoolView {
	return DomainPoolView{
		ID:        d.ID,
		Domain:    d.Domain,
		Name:      validator.SubdomainLabel(d.Domain, s.base),
		Remark:    d.Remark,
		CreatedAt: d.CreatedAt,
		CreatedBy: d.CreatedBy,
	}
}

// adminNames 超级管理员视角下的创建者登录名，失败时静默返回空表。
func (s *DomainPoolService) adminNames(ctx context.Context) map[uint]string {
	if s.admins == nil {
		return nil
	}
	users, err := s.admins.List(ctx)
	if err != nil {
		return nil
	}
	names := make(map[uint]string, len(users))
	for _, u := range users {
		names[u.ID] = u.Username
	}
	return names
}

func validateRemark(raw string) (string, error) {
	remark := strings.TrimSpace(raw)
	if len([]rune(remark)) > 255 {
		return "", fmt.Errorf("%w: 备注不能超过 255 个字符", ErrBadInput)
	}
	return remark, nil
}
