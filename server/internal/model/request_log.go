package model

import "time"

// RequestLog 公网入口的请求概要。大字段拆到 RequestLogBody，
// 保证列表按隧道/域名/时间筛选时不被 BLOB 拖慢。
type RequestLog struct {
	ID              uint       `json:"id" gorm:"primaryKey"`
	RequestID       string     `json:"request_id" gorm:"size:36;uniqueIndex;not null"`
	TunnelID        uint       `json:"tunnel_id" gorm:"index:idx_tunnel_created;not null"`
	SessionID       *uint      `json:"session_id"`
	Domain          string     `json:"domain" gorm:"size:253;index:idx_domain_created"`
	Method          string     `json:"method" gorm:"size:16"`
	Scheme          string     `json:"scheme" gorm:"size:16"`
	Host            string     `json:"host" gorm:"size:253"`
	Path            string     `json:"path" gorm:"size:2048"`
	QueryString     string     `json:"query_string" gorm:"size:4096"`
	QueryParams     string     `json:"query_params" gorm:"type:text"`
	FormParams      string     `json:"form_params" gorm:"type:text"`
	RemoteIP        string     `json:"remote_ip" gorm:"size:64"`
	UserAgent       string     `json:"user_agent" gorm:"size:512"`
	Referer         string     `json:"referer" gorm:"size:1024"`
	RequestHeaders  string     `json:"request_headers" gorm:"type:mediumtext"`
	ResponseHeaders string     `json:"response_headers" gorm:"type:mediumtext"`
	Status          int        `json:"status" gorm:"index"`
	RequestSize     int64      `json:"request_size"`
	ResponseSize    int64      `json:"response_size"`
	ReceivedAt      time.Time  `json:"received_at"`
	ForwardedAt     *time.Time `json:"forwarded_at"`
	FirstByteAt     *time.Time `json:"first_byte_at"`
	CompletedAt     *time.Time `json:"completed_at"`
	DurationMS      int64      `json:"duration_ms"`
	UpstreamMS      int64      `json:"upstream_ms"`
	LocalTarget     string     `json:"local_target" gorm:"size:512"`
	MatchedRule     string     `json:"matched_rule" gorm:"size:256"`
	ErrorMessage    string     `json:"error_message" gorm:"size:1024"`
	IsWebsocket     bool       `json:"is_websocket"`
	WSMessageCount  int64      `json:"ws_message_count"`
	CreatedAt       time.Time  `json:"created_at" gorm:"index:idx_tunnel_created;index:idx_domain_created"`
	Body            *RequestLogBody `json:"body,omitempty" gorm:"foreignKey:RequestLogID"`
}

// TableName 指定表名。
func (RequestLog) TableName() string {
	return "request_logs"
}
