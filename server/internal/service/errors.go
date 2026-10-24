// Package service 实现业务规则。Handler 只做参数绑定与响应封装。
package service

import "errors"

// 可被 handler 识别的业务错误。内部错误继续用 fmt.Errorf 包装。
var (
	ErrInvalidCredential = errors.New("用户名或密码错误")
	ErrNotFound          = errors.New("资源不存在")
	ErrBadInput          = errors.New("参数不合法")
	ErrTunnelOffline     = errors.New("客户端尚未联通，无法配置域名")
	ErrDomainTaken       = errors.New("该域名已被其他记录占用")
	ErrDomainMismatch    = errors.New("域名不属于该穿透记录")
	ErrOperation         = errors.New("操作未完成")
)
