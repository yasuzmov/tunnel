// Package eventbus 提供进程内观察者总线。
// 会话上下线、域名状态与新日志都经此推送到 SSE，避免 handler 直接耦合 Hub。
package eventbus

// 事件类型常量，前后端约定同名。
const (
	EventSessionOnline  = "session.online"
	EventSessionOffline = "session.offline"
	EventSessionBeat    = "session.heartbeat"
	EventDomainUpdated  = "domain.updated"
	EventLogCreated     = "log.created"
)

// Event 是一条可 JSON 序列化的通知。
type Event struct {
	Type    string `json:"type"`
	Payload any    `json:"payload"`
}
