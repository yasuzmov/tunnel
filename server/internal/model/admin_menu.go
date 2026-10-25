package model

// AdminMenu 管理员可见的左侧菜单权限。
// 超级管理员不写此表，默认拥有全部菜单。
type AdminMenu struct {
	ID      uint   `json:"id" gorm:"primaryKey"`
	AdminID uint   `json:"admin_id" gorm:"uniqueIndex:uk_admin_menu;not null"`
	MenuKey string `json:"menu_key" gorm:"size:32;uniqueIndex:uk_admin_menu;not null"`
}

// TableName 指定表名。
func (AdminMenu) TableName() string {
	return "admin_menus"
}
