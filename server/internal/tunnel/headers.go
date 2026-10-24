package tunnel

// HelloHeader 客户端握手信息。
type HelloHeader struct {
	Hostname string `json:"hostname"`
	OS       string `json:"os"`
	Version  string `json:"version"`
}

// WelcomeHeader 服务端确认，带上记录名称与已绑定域名。
type WelcomeHeader struct {
	TunnelID uint     `json:"tunnel_id"`
	Name     string   `json:"name"`
	Domains  []string `json:"domains"`
}

// PingHeader 心跳时间戳，单位毫秒。
type PingHeader struct {
	TS int64 `json:"ts"`
}

// ConfigUpdateHeader 域名变更推送。
type ConfigUpdateHeader struct {
	Domains []string `json:"domains"`
}

// HTTPReqHeader 公网 HTTP 请求元数据。
type HTTPReqHeader struct {
	Method        string              `json:"method"`
	Scheme        string              `json:"scheme"`
	Host          string              `json:"host"`
	Path          string              `json:"path"`
	Query         string              `json:"query"`
	Headers       map[string][]string `json:"headers"`
	ContentLength int64               `json:"content_length"`
}

// HTTPRespHeader 客户端回传的响应头及命中规则。
type HTTPRespHeader struct {
	Status       int                 `json:"status"`
	Headers      map[string][]string `json:"headers"`
	MatchedRule  string              `json:"matched_rule"`
	LocalTarget  string              `json:"local_target"`
}

// BodyChunkHeader eof=true 表示该 stream 的 body 结束，即使 payload 为空。
type BodyChunkHeader struct {
	EOF bool `json:"eof"`
}

// StreamErrorHeader 转发失败。
type StreamErrorHeader struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// CancelHeader 对端取消 stream。
type CancelHeader struct {
	Reason string `json:"reason"`
}

// WSOpenHeader 公网 WebSocket 握手。
type WSOpenHeader struct {
	Scheme  string              `json:"scheme"`
	Host    string              `json:"host"`
	Path    string              `json:"path"`
	Query   string              `json:"query"`
	Headers map[string][]string `json:"headers"`
}

// WSOpenAckHeader 客户端完成本地升级后的确认。
type WSOpenAckHeader struct {
	OK          bool                `json:"ok"`
	Status      int                 `json:"status"`
	Headers     map[string][]string `json:"headers"`
	MatchedRule string              `json:"matched_rule"`
	LocalTarget string              `json:"local_target"`
	Message     string              `json:"message"`
}

// WSMsgHeader 区分文本与二进制消息。
type WSMsgHeader struct {
	Binary bool `json:"binary"`
}

// WSCloseHeader 关闭码与原因。
type WSCloseHeader struct {
	Code   int    `json:"code"`
	Reason string `json:"reason"`
}
