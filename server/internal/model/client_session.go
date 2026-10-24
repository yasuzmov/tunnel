package model

import "time"

// 断开原因。新连接顶替与重置 key 需要可观测，不能只记 unknown。
const (
	DisconnectIdle     = "idle_timeout"
	DisconnectReplaced = "replaced_by_new_client"
	DisconnectResetKey = "key_reset"
	DisconnectClient   = "client_closed"
	DisconnectError    = "connection_error"
	DisconnectDeleted  = "tunnel_deleted"
	DisconnectShutdown = "server_shutdown"
)

// ClientSession 一次客户端联通记录。在线状态同时保存在内存 Hub 与本表，
// 便于后台刷新历史；内存才是转发路径的权威来源。
type ClientSession struct {
	ID              uint       `json:"id" gorm:"primaryKey"`
	TunnelID        uint       `json:"tunnel_id" gorm:"index;not null"`
	ClientIP        string     `json:"client_ip" gorm:"size:64"`
	Hostname        string     `json:"hostname" gorm:"size:256"`
	OS              string     `json:"os" gorm:"size:64"`
	Version         string     `json:"version" gorm:"size:64"`
	ConnectedAt     time.Time  `json:"connected_at"`
	DisconnectedAt  *time.Time `json:"disconnected_at"`
	LastHeartbeatAt time.Time  `json:"last_heartbeat_at"`
	DisconnectReason string    `json:"disconnect_reason" gorm:"size:64"`
	RequestCount    int64      `json:"request_count"`
	BytesIn         int64      `json:"bytes_in"`
	BytesOut        int64      `json:"bytes_out"`
	Online          bool       `json:"online" gorm:"index"`
}

// TableName 指定表名。
func (ClientSession) TableName() string {
	return "client_sessions"
}
