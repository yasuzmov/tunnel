package handler

import (
	"github.com/gin-gonic/gin"

	"tunnel-server/internal/pkg/response"
	"tunnel-server/internal/service"
)

// DomainPoolHandler 统一域名池管理，路由层已校验「域名管理」菜单权限；
// 数据按创建者隔离，在服务层校验，他人域名一律返回 404。
type DomainPoolHandler struct {
	svc *service.DomainPoolService
}

// NewDomainPoolHandler 创建域名池 handler。
func NewDomainPoolHandler(svc *service.DomainPoolService) *DomainPoolHandler {
	return &DomainPoolHandler{svc: svc}
}

// List 列出当前管理员可见的域名与绑定情况。
func (h *DomainPoolHandler) List(c *gin.Context) {
	adminID, ok := currentAdminID(c)
	if !ok {
		return
	}
	list, err := h.svc.List(c.Request.Context(), adminID)
	if err != nil {
		writeServiceError(c, err)
		return
	}
	response.OK(c, list)
}

// Config 返回创建域名时固定的主域名后缀。
func (h *DomainPoolHandler) Config(c *gin.Context) {
	response.OK(c, h.svc.Config())
}

// Create 新增域名：只需提交前缀 name。
func (h *DomainPoolHandler) Create(c *gin.Context) {
	adminID, ok := currentAdminID(c)
	if !ok {
		return
	}
	var in service.CreatePoolInput
	if err := c.ShouldBindJSON(&in); err != nil {
		response.BadRequest(c, "参数不合法")
		return
	}
	d, err := h.svc.Create(c.Request.Context(), adminID, in)
	if err != nil {
		writeServiceError(c, err)
		return
	}
	response.OK(c, d)
}

// Update 修改域名前缀与备注。
func (h *DomainPoolHandler) Update(c *gin.Context) {
	adminID, ok := currentAdminID(c)
	if !ok {
		return
	}
	id, ok := queryUint(c, "id")
	if !ok {
		response.BadRequest(c, "无效的域名 ID")
		return
	}
	var in service.UpdatePoolInput
	if err := c.ShouldBindJSON(&in); err != nil {
		response.BadRequest(c, "参数不合法")
		return
	}
	if err := h.svc.Update(c.Request.Context(), adminID, id, in); err != nil {
		writeServiceError(c, err)
		return
	}
	response.OK(c, gin.H{})
}

// Delete 删除域名。
func (h *DomainPoolHandler) Delete(c *gin.Context) {
	adminID, ok := currentAdminID(c)
	if !ok {
		return
	}
	id, ok := queryUint(c, "id")
	if !ok {
		response.BadRequest(c, "无效的域名 ID")
		return
	}
	if err := h.svc.Delete(c.Request.Context(), adminID, id); err != nil {
		writeServiceError(c, err)
		return
	}
	response.OK(c, gin.H{})
}
