package handler

import (
	"github.com/gin-gonic/gin"

	"tunnel-server/internal/pkg/response"
	"tunnel-server/internal/service"
)

// DomainPoolHandler 统一域名池管理，路由层已校验「域名管理」菜单权限。
type DomainPoolHandler struct {
	svc *service.DomainPoolService
}

// NewDomainPoolHandler 创建域名池 handler。
func NewDomainPoolHandler(svc *service.DomainPoolService) *DomainPoolHandler {
	return &DomainPoolHandler{svc: svc}
}

// List 列出域名池与绑定情况。
func (h *DomainPoolHandler) List(c *gin.Context) {
	list, err := h.svc.List(c.Request.Context())
	if err != nil {
		writeServiceError(c, err)
		return
	}
	response.OK(c, list)
}

// Create 新增域名。
func (h *DomainPoolHandler) Create(c *gin.Context) {
	var in service.CreatePoolInput
	if err := c.ShouldBindJSON(&in); err != nil {
		response.BadRequest(c, "参数不合法")
		return
	}
	d, err := h.svc.Create(c.Request.Context(), in)
	if err != nil {
		writeServiceError(c, err)
		return
	}
	response.OK(c, d)
}

// Update 修改域名备注。
func (h *DomainPoolHandler) Update(c *gin.Context) {
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
	if err := h.svc.Update(c.Request.Context(), id, in); err != nil {
		writeServiceError(c, err)
		return
	}
	response.OK(c, gin.H{})
}

// Delete 删除域名。
func (h *DomainPoolHandler) Delete(c *gin.Context) {
	id, ok := queryUint(c, "id")
	if !ok {
		response.BadRequest(c, "无效的域名 ID")
		return
	}
	if err := h.svc.Delete(c.Request.Context(), id); err != nil {
		writeServiceError(c, err)
		return
	}
	response.OK(c, gin.H{})
}
