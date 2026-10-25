package repository

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	"tunnel-server/internal/model"
)

type adminUserRepo struct {
	db *gorm.DB
}

// NewAdminUserRepo 创建管理员仓储。
func NewAdminUserRepo(db *gorm.DB) AdminUserRepository {
	return &adminUserRepo{db: db}
}

// FindByUsername 按登录名查找，未找到返回 gorm.ErrRecordNotFound。
func (r *adminUserRepo) FindByUsername(ctx context.Context, username string) (*model.AdminUser, error) {
	var u model.AdminUser
	err := r.db.WithContext(ctx).Where("username = ?", username).First(&u).Error
	if err != nil {
		return nil, wrapFind("管理员", err)
	}
	return &u, nil
}

// FindByID 按主键查找。
func (r *adminUserRepo) FindByID(ctx context.Context, id uint) (*model.AdminUser, error) {
	var u model.AdminUser
	err := r.db.WithContext(ctx).First(&u, id).Error
	if err != nil {
		return nil, wrapFind("管理员", err)
	}
	return &u, nil
}

// UpdatePassword 只更新口令哈希，避免 Save 整行覆盖其他字段。
func (r *adminUserRepo) UpdatePassword(ctx context.Context, id uint, passwordHash string) error {
	res := r.db.WithContext(ctx).Model(&model.AdminUser{}).Where("id = ?", id).Update("password_hash", passwordHash)
	if res.Error != nil {
		return fmt.Errorf("更新管理员口令: %w", res.Error)
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// Create 新增管理员账号。
func (r *adminUserRepo) Create(ctx context.Context, u *model.AdminUser) error {
	if err := r.db.WithContext(ctx).Create(u).Error; err != nil {
		return fmt.Errorf("创建管理员: %w", err)
	}
	return nil
}

// UpdateProfile 只更新登录名与角色。
// MySQL 在字段值未变化时 RowsAffected 为 0，因此需要再确认记录是否存在。
func (r *adminUserRepo) UpdateProfile(ctx context.Context, id uint, username, role string) error {
	res := r.db.WithContext(ctx).Model(&model.AdminUser{}).Where("id = ?", id).
		Updates(map[string]any{"username": username, "role": role})
	if res.Error != nil {
		return fmt.Errorf("更新管理员: %w", res.Error)
	}
	if res.RowsAffected == 0 {
		var count int64
		if err := r.db.WithContext(ctx).Model(&model.AdminUser{}).Where("id = ?", id).Count(&count).Error; err != nil {
			return fmt.Errorf("确认管理员: %w", err)
		}
		if count == 0 {
			return gorm.ErrRecordNotFound
		}
	}
	return nil
}

// Delete 物理删除管理员账号；授权关系由域名池仓储级联清理。
func (r *adminUserRepo) Delete(ctx context.Context, id uint) error {
	res := r.db.WithContext(ctx).Delete(&model.AdminUser{}, id)
	if res.Error != nil {
		return fmt.Errorf("删除管理员: %w", res.Error)
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// List 列出全部管理员，按创建顺序。
func (r *adminUserRepo) List(ctx context.Context) ([]model.AdminUser, error) {
	var list []model.AdminUser
	if err := r.db.WithContext(ctx).Order("id ASC").Find(&list).Error; err != nil {
		return nil, fmt.Errorf("列出管理员: %w", err)
	}
	return list, nil
}

// CountByRole 统计某角色的管理员数量，用于避免删除/降级最后一个超级管理员。
func (r *adminUserRepo) CountByRole(ctx context.Context, role string) (int64, error) {
	var n int64
	if err := r.db.WithContext(ctx).Model(&model.AdminUser{}).Where("role = ?", role).Count(&n).Error; err != nil {
		return 0, fmt.Errorf("统计管理员角色: %w", err)
	}
	return n, nil
}

// ListMenus 返回管理员被授予的菜单权限。
func (r *adminUserRepo) ListMenus(ctx context.Context, adminID uint) ([]string, error) {
	var keys []string
	err := r.db.WithContext(ctx).Model(&model.AdminMenu{}).
		Where("admin_id = ?", adminID).
		Order("id ASC").
		Pluck("menu_key", &keys).Error
	if err != nil {
		return nil, fmt.Errorf("列出管理员菜单权限: %w", err)
	}
	return keys, nil
}

// SetMenus 覆盖式写入菜单权限：先清空再批量插入。
func (r *adminUserRepo) SetMenus(ctx context.Context, adminID uint, menus []string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("admin_id = ?", adminID).Delete(&model.AdminMenu{}).Error; err != nil {
			return fmt.Errorf("清理管理员菜单权限: %w", err)
		}
		if len(menus) == 0 {
			return nil
		}
		rows := make([]model.AdminMenu, 0, len(menus))
		for _, key := range menus {
			rows = append(rows, model.AdminMenu{AdminID: adminID, MenuKey: key})
		}
		if err := tx.Create(&rows).Error; err != nil {
			return fmt.Errorf("写入管理员菜单权限: %w", err)
		}
		return nil
	})
}

func wrapFind(name string, err error) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	return fmt.Errorf("查询%s: %w", name, err)
}
