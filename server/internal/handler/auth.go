package handler

import (
	"errors"

	"github.com/gin-gonic/gin"

	"tunnel-server/internal/middleware"
	"tunnel-server/internal/pkg/response"
	"tunnel-server/internal/service"
)

// AuthHandler 登录与当前用户。
type AuthHandler struct {
	svc *service.AuthService
}

// NewAuthHandler 创建认证 handler。
func NewAuthHandler(svc *service.AuthService) *AuthHandler {
	return &AuthHandler{svc: svc}
}

type loginReq struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// Login 签发 JWT。
func (h *AuthHandler) Login(c *gin.Context) {
	var req loginReq
	if err := c.ShouldBindJSON(&req); err != nil || req.Username == "" || req.Password == "" {
		response.BadRequest(c, "请输入用户名和密码")
		return
	}
	out, err := h.svc.Login(c.Request.Context(), req.Username, req.Password)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredential) {
			response.Unauthorized(c, err.Error())
			return
		}
		response.ServerError(c, "登录失败")
		return
	}
	response.OK(c, out)
}

// Me 返回当前登录管理员。
func (h *AuthHandler) Me(c *gin.Context) {
	claims := middleware.Claims(c)
	if claims == nil {
		response.Unauthorized(c, "未登录")
		return
	}
	user, err := h.svc.Me(c.Request.Context(), claims.UserID)
	if err != nil {
		writeServiceError(c, err)
		return
	}
	response.OK(c, user)
}

// ChangePassword 修改当前登录管理员的口令。
func (h *AuthHandler) ChangePassword(c *gin.Context) {
	claims := middleware.Claims(c)
	if claims == nil {
		response.Unauthorized(c, "未登录")
		return
	}
	var in service.ChangePasswordInput
	if err := c.ShouldBindJSON(&in); err != nil {
		response.BadRequest(c, "参数不合法")
		return
	}
	if err := h.svc.ChangePassword(c.Request.Context(), claims.UserID, in); err != nil {
		writeServiceError(c, err)
		return
	}
	response.OK(c, gin.H{})
}
