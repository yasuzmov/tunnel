package handler

import (
	"github.com/gin-gonic/gin"

	"tunnel-server/internal/pkg/response"
	"tunnel-server/internal/service"
)

// requireMenu 校验当前管理员是否有权访问指定左侧菜单。
// 超级管理员默认拥有全部菜单，普通管理员按 admin_menus 授权判断。
func requireMenu(svc *service.AdminService, menu string) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, ok := currentAdminID(c)
		if !ok {
			c.Abort()
			return
		}
		allowed, err := svc.HasMenu(c.Request.Context(), id, menu)
		if err != nil {
			writeServiceError(c, err)
			c.Abort()
			return
		}
		if !allowed {
			response.Forbidden(c, "没有权限访问该菜单")
			c.Abort()
			return
		}
		c.Next()
	}
}
