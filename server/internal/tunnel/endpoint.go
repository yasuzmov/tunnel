package tunnel

import (
	"context"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"

	"tunnel-server/internal/eventbus"
	"tunnel-server/internal/model"
	"tunnel-server/internal/pkg/httputil"
	"tunnel-server/internal/repository"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  64 * 1024,
	WriteBufferSize: 64 * 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

// Endpoint 处理客户端 /tunnel/ws。升级前校验 key，让“测试连通性”能拿到明确 401。
type Endpoint struct {
	Tunnels  repository.TunnelRepository
	Domains  repository.DomainRepository
	Sessions repository.SessionRepository
	Hub      *Hub
	Bus      *eventbus.Bus
	Idle     time.Duration
}

// Handle 鉴权、升级、登记会话并阻塞直到断开。
func (e *Endpoint) Handle(c *gin.Context) {
	key := extractKey(c)
	if key == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "缺少接入 key"})
		return
	}
	tun, err := e.Tunnels.FindByKey(c.Request.Context(), key)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "接入 key 无效"})
		return
	}
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("升级隧道 WebSocket 失败: %v", err)
		return
	}
	e.serveConn(conn, tun, httputil.ClientIP(c.Request))
}

func (e *Endpoint) serveConn(conn *websocket.Conn, tun *model.Tunnel, ip string) {
	ctx := context.Background()
	list, _ := e.Domains.ListByTunnel(ctx, tun.ID)
	tun.Domains = list
	domains := domainNames(tun)
	sess := NewSession(conn, tun.ID, tun.Name, domains, ip)
	if _, err := sess.WaitHello(10 * time.Second); err != nil {
		_ = conn.Close()
		return
	}
	row, err := e.persistOnline(ctx, sess)
	if err != nil {
		sess.Close(model.DisconnectError)
		return
	}
	sess.ID = row.ID
	sess.onPing = func() {
		now := time.Now()
		_ = e.Sessions.TouchHeartbeat(context.Background(), sess.ID, now)
		e.publish(eventbus.EventSessionBeat, map[string]any{
			"session_id": sess.ID,
			"tunnel_id":  sess.TunnelID,
			"at":         now,
		})
	}
	if old := e.Hub.Register(sess); old != nil {
		old.Close(model.DisconnectReplaced)
		_ = e.Sessions.MarkOffline(context.Background(), old.ID, time.Now(), model.DisconnectReplaced, statsOf(old))
		e.publish(eventbus.EventSessionOffline, old)
	}
	_ = e.Tunnels.TouchOnline(ctx, tun.ID, time.Now())
	if err := sess.SendWelcome(); err != nil {
		sess.Close(model.DisconnectError)
		e.Hub.Unregister(sess)
		_ = e.Sessions.MarkOffline(context.Background(), sess.ID, time.Now(), model.DisconnectError, statsOf(sess))
		return
	}
	e.publish(eventbus.EventSessionOnline, row)
	sess.Run(e.Idle)
	e.Hub.Unregister(sess)
	_ = e.Sessions.MarkOffline(context.Background(), sess.ID, time.Now(), sess.CloseReason(), statsOf(sess))
	e.publish(eventbus.EventSessionOffline, sess)
}

func statsOf(s *Session) repository.SessionStats {
	return repository.SessionStats{
		RequestCount: s.Requests.Load(),
		BytesIn:      s.BytesIn.Load(),
		BytesOut:     s.BytesOut.Load(),
	}
}

func (e *Endpoint) persistOnline(ctx context.Context, sess *Session) (*model.ClientSession, error) {
	now := time.Now()
	row := &model.ClientSession{
		TunnelID:        sess.TunnelID,
		ClientIP:        sess.ClientIP,
		Hostname:        sess.Hostname,
		OS:              sess.OS,
		Version:         sess.Version,
		ConnectedAt:     now,
		LastHeartbeatAt: now,
		Online:          true,
	}
	if err := e.Sessions.Create(ctx, row); err != nil {
		return nil, err
	}
	return row, nil
}

func (e *Endpoint) publish(typ string, payload any) {
	if e.Bus == nil {
		return
	}
	e.Bus.Publish(eventbus.Event{Type: typ, Payload: payload})
}

func extractKey(c *gin.Context) string {
	h := c.GetHeader("Authorization")
	if strings.HasPrefix(strings.ToLower(h), "bearer ") {
		return strings.TrimSpace(h[7:])
	}
	return strings.TrimSpace(c.Query("key"))
}

func domainNames(t *model.Tunnel) []string {
	out := make([]string, 0, len(t.Domains))
	for _, d := range t.Domains {
		out = append(out, d.Domain)
	}
	return out
}
