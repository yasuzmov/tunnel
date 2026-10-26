package handler

import "testing"

func TestEventTunnelID(t *testing.T) {
	cases := []struct {
		name    string
		payload string
		want    uint
		found   bool
	}{
		{"日志/联通记录/域名使用 json 标签", `{"id":5,"tunnel_id":7}`, 7, true},
		{"隧道会话结构体无标签", `{"ID":5,"TunnelID":8}`, 8, true},
		{"心跳事件", `{"session_id":5,"tunnel_id":9,"at":"2026-01-01T00:00:00Z"}`, 9, true},
		{"无记录 ID 的事件不下发给普通管理员", `{"session_id":5}`, 0, false},
		{"非对象载荷", `"x"`, 0, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, found := eventTunnelID([]byte(tc.payload))
			if got != tc.want || found != tc.found {
				t.Fatalf("want (%d,%v), got (%d,%v)", tc.want, tc.found, got, found)
			}
		})
	}
}
