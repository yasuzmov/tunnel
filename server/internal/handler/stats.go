package handler

import (
	"github.com/gin-gonic/gin"

	"tunnel-server/internal/pkg/response"
	"tunnel-server/internal/service"
)

// StatsHandler 仪表盘。
type StatsHandler struct {
	svc *service.StatsService
}

// NewStatsHandler 创建统计 handler。
func NewStatsHandler(svc *service.StatsService) *StatsHandler {
	return &StatsHandler{svc: svc}
}

// Dashboard 返回在线数、记录数、今日请求与错误率。
func (h *StatsHandler) Dashboard(c *gin.Context) {
	s, err := h.svc.Dashboard(c.Request.Context())
	if err != nil {
		writeServiceError(c, err)
		return
	}
	response.OK(c, s)
}
