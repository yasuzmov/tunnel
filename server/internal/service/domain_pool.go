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
type DomainPoolService struct {
	pool  repository.DomainPoolRepository
	bound repository.DomainRepository
}

// NewDomainPoolService 创建域名池服务。
func NewDomainPoolService(pool repository.DomainPoolRepository, bound repository.DomainRepository) *DomainPoolService {
	return &DomainPoolService{pool: pool, bound: bound}
}

// DomainPoolView 域名池条目与当前绑定情况。
type DomainPoolView struct {
	ID              uint      `json:"id"`
	Domain          string    `json:"domain"`
	Remark          string    `json:"remark"`
	CreatedAt       time.Time `json:"created_at"`
	BoundTunnelID   uint      `json:"bound_tunnel_id"`
	BoundTunnelName string    `json:"bound_tunnel_name"`
	HTTPSEnabled    bool      `json:"https_enabled"`
	CertStatus      string    `json:"cert_status"`
	NginxStatus     string    `json:"nginx_status"`
}

// CreatePoolInput 新增域名池条目。
type CreatePoolInput struct {
	Domain string `json:"domain"`
	Remark string `json:"remark"`
}

// UpdatePoolInput 修改域名备注。
type UpdatePoolInput struct {
	Remark string `json:"remark"`
}

// List 列出域名池与绑定情况。
func (s *DomainPoolService) List(ctx context.Context) ([]DomainPoolView, error) {
	domains, err := s.pool.List(ctx, "")
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
	views := make([]DomainPoolView, 0, len(domains))
	for _, d := range domains {
		v := DomainPoolView{
			ID:        d.ID,
			Domain:    d.Domain,
			Remark:    d.Remark,
			CreatedAt: d.CreatedAt,
		}
		if b, ok := byID[d.ID]; ok {
			v.BoundTunnelID = b.TunnelID
			v.BoundTunnelName = b.TunnelName
			v.HTTPSEnabled = b.HTTPSEnabled
			v.CertStatus = b.CertStatus
			v.NginxStatus = b.NginxStatus
		}
		views = append(views, v)
	}
	return views, nil
}

// Create 校验域名格式并写入域名池。
func (s *DomainPoolService) Create(ctx context.Context, in CreatePoolInput) (*model.Domain, error) {
	host, err := validator.ValidateDomain(in.Domain)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrBadInput, err.Error())
	}
	remark, err := validateRemark(in.Remark)
	if err != nil {
		return nil, err
	}
	if _, err := s.pool.FindByDomain(ctx, host); err == nil {
		return nil, ErrDomainPoolTaken
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	d := &model.Domain{Domain: host, Remark: remark}
	if err := s.pool.Create(ctx, d); err != nil {
		return nil, err
	}
	return d, nil
}

// Update 修改备注。
func (s *DomainPoolService) Update(ctx context.Context, id uint, in UpdatePoolInput) error {
	remark, err := validateRemark(in.Remark)
	if err != nil {
		return err
	}
	if err := s.pool.UpdateRemark(ctx, id, remark); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrNotFound
		}
		return err
	}
	return nil
}

// Delete 删除域名池条目；已绑定到穿透记录的域名必须先解绑。
func (s *DomainPoolService) Delete(ctx context.Context, id uint) error {
	d, err := s.pool.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrNotFound
		}
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

func validateRemark(raw string) (string, error) {
	remark := strings.TrimSpace(raw)
	if len([]rune(remark)) > 255 {
		return "", fmt.Errorf("%w: 备注不能超过 255 个字符", ErrBadInput)
	}
	return remark, nil
}
