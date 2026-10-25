// Package service 实现业务规则。Handler 只做参数绑定与响应封装。
package service

import "errors"

// 可被 handler 识别的业务错误。内部错误继续用 fmt.Errorf 包装。
var (
	ErrInvalidCredential = errors.New("用户名或密码错误")
	// ErrWrongOldPassword 不能映射为 401，否则前端会把已登录用户踢回登录页。
	ErrWrongOldPassword = errors.New("原密码不正确")
	ErrNotFound         = errors.New("资源不存在")
	ErrBadInput         = errors.New("参数不合法")
	ErrTunnelOffline    = errors.New("客户端尚未联通，无法配置域名")
	ErrDomainTaken      = errors.New("该域名已被其他记录占用")
	ErrDomainMismatch   = errors.New("域名不属于该穿透记录")
	ErrOperation        = errors.New("操作未完成")
	ErrForbidden        = errors.New("没有权限执行该操作")
	ErrAdminTaken       = errors.New("该用户名已存在")
	ErrLastSuperAdmin   = errors.New("必须保留至少一个超级管理员")
	ErrDomainPoolTaken  = errors.New("该域名已在域名列表中")
	ErrDomainNotInPool  = errors.New("该域名不在可用域名列表中，请联系超级管理员添加")
	ErrDomainNotAllowed = errors.New("当前管理员无权使用该域名")
	ErrDomainBound      = errors.New("该域名已绑定到穿透记录，无法删除")
)
