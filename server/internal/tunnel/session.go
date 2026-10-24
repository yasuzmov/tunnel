package tunnel

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
)

// Session 一条客户端长连接。同一隧道同时只允许一个 Session 转发流量。
type Session struct {
	ID         uint
	TunnelID   uint
	TunnelName string
	Domains    []string
	ClientIP   string
	Hostname   string
	OS         string
	Version    string

	conn   *websocket.Conn
	writeM sync.Mutex

	mu      sync.Mutex
	streams map[uint32]*Stream
	nextID  uint32

	closed      chan struct{}
	closeOnce   sync.Once
	closeReason string

	beatMu   sync.Mutex
	lastBeat time.Time

	// onPing 在收到客户端心跳后回调，用于刷新联通记录与推送 SSE。
	onPing func()

	Requests atomic.Int64
	BytesIn  atomic.Int64
	BytesOut atomic.Int64
}

// NewSession 包装已升级的 WebSocket。调用方随后执行 Hello 与 Run。
func NewSession(conn *websocket.Conn, tunnelID uint, name string, domains []string, clientIP string) *Session {
	return &Session{
		TunnelID:   tunnelID,
		TunnelName: name,
		Domains:    append([]string(nil), domains...),
		ClientIP:   clientIP,
		conn:       conn,
		streams:    make(map[uint32]*Stream),
		nextID:     1,
		closed:     make(chan struct{}),
		lastBeat:   time.Now(),
	}
}

// Close 幂等关闭连接并失败所有进行中的 stream。
func (s *Session) Close(reason string) {
	s.closeOnce.Do(func() {
		s.closeReason = reason
		close(s.closed)
		_ = s.conn.Close()
		s.failAll(reason)
	})
}

// CloseReason 返回关闭原因，未关闭时为空。
func (s *Session) CloseReason() string {
	return s.closeReason
}

// Closed 在会话结束时关闭。
func (s *Session) Closed() <-chan struct{} {
	return s.closed
}

func (s *Session) allocStream() *Stream {
	s.mu.Lock()
	defer s.mu.Unlock()
	id := s.nextID
	s.nextID++
	st := newStream(id)
	s.streams[id] = st
	return st
}

func (s *Session) getStream(id uint32) *Stream {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.streams[id]
}

func (s *Session) dropStream(id uint32) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.streams, id)
}

// DropStream 在 HTTP/WebSocket 转发结束后释放 stream 映射。
func (s *Session) DropStream(id uint32) {
	s.dropStream(id)
}

func (s *Session) failAll(reason string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	err := &streamErr{code: "session_closed", msg: reason}
	for id, st := range s.streams {
		st.fail(err)
		delete(s.streams, id)
	}
}

func (s *Session) touchBeat() {
	s.beatMu.Lock()
	s.lastBeat = time.Now()
	s.beatMu.Unlock()
}

// LastBeat 最近一次收到客户端帧的时间。
func (s *Session) LastBeat() time.Time {
	s.beatMu.Lock()
	defer s.beatMu.Unlock()
	return s.lastBeat
}

type streamErr struct {
	code string
	msg  string
}

func (e *streamErr) Error() string {
	if e.msg == "" {
		return e.code
	}
	return e.code + ": " + e.msg
}
