package tunnel

import (
	"fmt"
	"time"

	"github.com/gorilla/websocket"
)

// Run 阻塞读取客户端帧，直到连接断开或空闲超时。
func (s *Session) Run(idle time.Duration) {
	defer s.Close(s.reasonOr(modelDisconnectClient()))
	go s.watchIdle(idle)
	for {
		_ = s.conn.SetReadDeadline(time.Now().Add(idle))
		typ, data, err := s.conn.ReadMessage()
		if err != nil {
			s.Close(modelDisconnectError())
			return
		}
		s.touchBeat()
		s.BytesIn.Add(int64(len(data)))
		if typ != websocket.BinaryMessage {
			continue
		}
		frame, err := DecodeFrame(data)
		if err != nil {
			continue
		}
		s.dispatch(frame)
	}
}

func (s *Session) watchIdle(idle time.Duration) {
	ticker := time.NewTicker(idle / 3)
	defer ticker.Stop()
	for {
		select {
		case <-s.closed:
			return
		case <-ticker.C:
			if time.Since(s.LastBeat()) > idle {
				s.Close("idle_timeout")
				return
			}
		}
	}
}

func (s *Session) dispatch(f Frame) {
	switch f.Type {
	case TypePing:
		s.handlePing(f)
	case TypePong:
		return
	case TypeHTTPResp, TypeWSOpenAck:
		s.deliverHeader(f)
	case TypeBodyChunk:
		s.deliverBody(f)
	case TypeWSMsg, TypeWSClose:
		s.deliverMsg(f)
	case TypeStreamError:
		s.deliverError(f)
	}
}

func (s *Session) handlePing(f Frame) {
	var h PingHeader
	_ = UnmarshalHeader(f, &h)
	if h.TS == 0 {
		h.TS = time.Now().UnixMilli()
	}
	_ = s.writeTyped(TypePong, 0, h, nil)
	if s.onPing != nil {
		s.onPing()
	}
}

func (s *Session) deliverHeader(f Frame) {
	st := s.getStream(f.StreamID)
	if st == nil {
		return
	}
	select {
	case st.headerCh <- f:
	default:
	}
}

func (s *Session) deliverBody(f Frame) {
	st := s.getStream(f.StreamID)
	if st == nil {
		return
	}
	var h BodyChunkHeader
	_ = UnmarshalHeader(f, &h)
	if len(f.Payload) > 0 {
		if _, err := st.pw.Write(f.Payload); err != nil {
			return
		}
	}
	if h.EOF {
		st.closeBody()
	}
}

func (s *Session) deliverMsg(f Frame) {
	st := s.getStream(f.StreamID)
	if st == nil {
		return
	}
	select {
	case st.msgCh <- f:
	case <-s.closed:
	default:
	}
}

func (s *Session) deliverError(f Frame) {
	st := s.getStream(f.StreamID)
	if st == nil {
		return
	}
	select {
	case st.headerCh <- f:
	default:
	}
	st.fail(parseStreamError(f))
}

func (s *Session) reasonOr(fallback string) string {
	if s.closeReason != "" {
		return s.closeReason
	}
	return fallback
}

func modelDisconnectClient() string { return "client_closed" }
func modelDisconnectError() string  { return "connection_error" }

// WaitHello 在超时内读取 hello 帧。鉴权已在升级前完成，这里只收客户端元数据。
func (s *Session) WaitHello(timeout time.Duration) (*HelloHeader, error) {
	_ = s.conn.SetReadDeadline(time.Now().Add(timeout))
	typ, data, err := s.conn.ReadMessage()
	if err != nil {
		return nil, fmt.Errorf("等待 hello: %w", err)
	}
	if typ != websocket.BinaryMessage {
		return nil, fmt.Errorf("hello 必须是二进制帧")
	}
	frame, err := DecodeFrame(data)
	if err != nil {
		return nil, err
	}
	if frame.Type != TypeHello {
		return nil, fmt.Errorf("首帧必须是 hello")
	}
	var h HelloHeader
	if err := UnmarshalHeader(frame, &h); err != nil {
		return nil, err
	}
	s.Hostname, s.OS, s.Version = h.Hostname, h.OS, h.Version
	s.touchBeat()
	return &h, nil
}

// RecvWS 读取一条 WebSocket 相关帧，会话关闭时返回错误。
func (s *Session) RecvWS(st *Stream) (Frame, error) {
	select {
	case <-s.closed:
		return Frame{}, fmt.Errorf("会话已关闭")
	case err := <-st.errCh:
		return Frame{}, err
	case f := <-st.msgCh:
		return f, nil
	}
}
