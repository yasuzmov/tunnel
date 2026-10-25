package handler

import (
	"github.com/gin-gonic/gin"

	"tunnel-server/internal/pkg/response"
	"tunnel-server/internal/service"
)

// DomainHandler 域名绑定与 nginx 预览。
type DomainHandler struct {
	svc *service.DomainService
}

// NewDomainHandler 创建域名 handler。
func NewDomainHandler(svc *service.DomainService) *DomainHandler {
	return &DomainHandler{svc: svc}
}

func tunnelID(c *gin.Context) (uint, bool) {
	return queryUint(c, "id")
}

func domainID(c *gin.Context) (uint, bool) {
	return queryUint(c, "did")
}

// List 列出某记录下的域名。
func (h *DomainHandler) List(c *gin.Context) {
	tid, ok := tunnelID(c)
	if !ok {
		response.BadRequest(c, "无效的记录 ID")
		return
	}
	list, err := h.svc.List(c.Request.Context(), tid)
	if err != nil {
		writeServiceError(c, err)
		return
	}
	response.OK(c, list)
}

// Create 绑定域名。
func (h *DomainHandler) Create(c *gin.Context) {
	tid, ok := tunnelID(c)
	if !ok {
		response.BadRequest(c, "无效的记录 ID")
		return
	}
	adminID, ok := currentAdminID(c)
	if !ok {
		return
	}
	var in service.CreateInputDomain
	if err := c.ShouldBindJSON(&in); err != nil {
		response.BadRequest(c, "参数不合法")
		return
	}
	d, err := h.svc.Create(c.Request.Context(), adminID, tid, in)
	if err != nil {
		writeServiceError(c, err)
		return
	}
	response.OK(c, d)
}

// Available 列出当前管理员可绑定、且尚未被占用的域名。
func (h *DomainHandler) Available(c *gin.Context) {
	adminID, ok := currentAdminID(c)
	if !ok {
		return
	}
	list, err := h.svc.AvailableDomains(c.Request.Context(), adminID)
	if err != nil {
		writeServiceError(c, err)
		return
	}
	response.OK(c, list)
}

// Delete 解绑域名。
func (h *DomainHandler) Delete(c *gin.Context) {
	tid, ok := tunnelID(c)
	did, ok2 := domainID(c)
	if !ok || !ok2 {
		response.BadRequest(c, "无效的 ID")
		return
	}
	if err := h.svc.Delete(c.Request.Context(), tid, did); err != nil {
		writeServiceError(c, err)
		return
	}
	response.OK(c, gin.H{})
}

// Preview 查看将生成的 nginx 配置。
func (h *DomainHandler) Preview(c *gin.Context) {
	tid, ok := tunnelID(c)
	did, ok2 := domainID(c)
	if !ok || !ok2 {
		response.BadRequest(c, "无效的 ID")
		return
	}
	conf, err := h.svc.Preview(c.Request.Context(), tid, did)
	if err != nil {
		writeServiceError(c, err)
		return
	}
	response.OK(c, gin.H{"config": conf})
}
