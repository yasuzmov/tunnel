package handler

import (
	"github.com/gin-gonic/gin"

	"tunnel-server/internal/middleware"
	"tunnel-server/internal/pkg/response"
	"tunnel-server/internal/service"
)

// AdminHandler 管理员账号与域名授权管理，仅超级管理员可用。
type AdminHandler struct {
	svc *service.AdminService
}

// NewAdminHandler 创建管理员 handler。
func NewAdminHandler(svc *service.AdminService) *AdminHandler {
	return &AdminHandler{svc: svc}
}

// List 列出全部管理员。
func (h *AdminHandler) List(c *gin.Context) {
	if !requireSuper(c, h.svc) {
		return
	}
	list, err := h.svc.List(c.Request.Context())
	if err != nil {
		writeServiceError(c, err)
		return
	}
	response.OK(c, list)
}

// Create 新增管理员。
func (h *AdminHandler) Create(c *gin.Context) {
	if !requireSuper(c, h.svc) {
		return
	}
	var in service.CreateAdminInput
	if err := c.ShouldBindJSON(&in); err != nil {
		response.BadRequest(c, "参数不合法")
		return
	}
	view, err := h.svc.Create(c.Request.Context(), in)
	if err != nil {
		writeServiceError(c, err)
		return
	}
	response.OK(c, view)
}

// Update 修改登录名与角色。
func (h *AdminHandler) Update(c *gin.Context) {
	if !requireSuper(c, h.svc) {
		return
	}
	id, ok := queryUint(c, "id")
	if !ok {
		response.BadRequest(c, "无效的管理员 ID")
		return
	}
	var in service.UpdateAdminInput
	if err := c.ShouldBindJSON(&in); err != nil {
		response.BadRequest(c, "参数不合法")
		return
	}
	view, err := h.svc.Update(c.Request.Context(), id, in)
	if err != nil {
		writeServiceError(c, err)
		return
	}
	response.OK(c, view)
}

type resetAdminPasswordReq struct {
	Password string `json:"password"`
}

// ResetPassword 重置指定管理员的口令。
func (h *AdminHandler) ResetPassword(c *gin.Context) {
	if !requireSuper(c, h.svc) {
		return
	}
	id, ok := queryUint(c, "id")
	if !ok {
		response.BadRequest(c, "无效的管理员 ID")
		return
	}
	var in resetAdminPasswordReq
	if err := c.ShouldBindJSON(&in); err != nil {
		response.BadRequest(c, "参数不合法")
		return
	}
	if err := h.svc.ResetPassword(c.Request.Context(), id, in.Password); err != nil {
		writeServiceError(c, err)
		return
	}
	response.OK(c, gin.H{})
}

// Delete 删除管理员。
func (h *AdminHandler) Delete(c *gin.Context) {
	if !requireSuper(c, h.svc) {
		return
	}
	id, ok := queryUint(c, "id")
	if !ok {
		response.BadRequest(c, "无效的管理员 ID")
		return
	}
	operatorID, ok := currentAdminID(c)
	if !ok {
		return
	}
	if err := h.svc.Delete(c.Request.Context(), id, operatorID); err != nil {
		writeServiceError(c, err)
		return
	}
	response.OK(c, gin.H{})
}

// currentAdminID 读取当前登录管理员 ID，失败时已写入响应。
func currentAdminID(c *gin.Context) (uint, bool) {
	claims := middleware.Claims(c)
	if claims == nil {
		response.Unauthorized(c, "未登录")
		return 0, false
	}
	return claims.UserID, true
}

// requireSuper 校验当前登录管理员是否为超级管理员，失败时已写入响应。
func requireSuper(c *gin.Context, svc *service.AdminService) bool {
	id, ok := currentAdminID(c)
	if !ok {
		return false
	}
	super, err := svc.IsSuper(c.Request.Context(), id)
	if err != nil {
		writeServiceError(c, err)
		return false
	}
	if !super {
		response.Forbidden(c, "仅超级管理员可执行此操作")
		return false
	}
	return true
}
