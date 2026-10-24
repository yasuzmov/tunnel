package tunnel

import (
	"context"
	"fmt"
	"io"
)

// OpenHTTP 向客户端发起一条 HTTP 转发，返回响应头与响应体流。
// 调用方读完 body 后必须 Close，以便释放 stream。
func (s *Session) OpenHTTP(ctx context.Context, req HTTPReqHeader, body io.Reader) (*HTTPRespHeader, io.ReadCloser, error) {
	st := s.allocStream()
	s.Requests.Add(1)
	if err := s.writeTyped(TypeHTTPReq, st.ID, req, nil); err != nil {
		s.dropStream(st.ID)
		return nil, nil, err
	}
	if body == nil {
		body = eofReader{}
	}
	if err := s.sendBody(st.ID, body); err != nil {
		s.dropStream(st.ID)
		st.fail(err)
		return nil, nil, err
	}
	frame, err := s.waitHeader(ctx, st)
	if err != nil {
		s.SendCancel(st.ID, "wait_header")
		s.dropStream(st.ID)
		return nil, nil, err
	}
	if frame.Type == TypeStreamError {
		return nil, nil, parseStreamError(frame)
	}
	var resp HTTPRespHeader
	if err := UnmarshalHeader(frame, &resp); err != nil {
		s.dropStream(st.ID)
		return nil, nil, err
	}
	return &resp, &streamBody{Stream: st, session: s}, nil
}

// OpenWS 让客户端先连本地 WebSocket，成功后再升级公网连接。
func (s *Session) OpenWS(ctx context.Context, req WSOpenHeader) (*WSOpenAckHeader, *Stream, error) {
	st := s.allocStream()
	s.Requests.Add(1)
	if err := s.writeTyped(TypeWSOpen, st.ID, req, nil); err != nil {
		s.dropStream(st.ID)
		return nil, nil, err
	}
	frame, err := s.waitHeader(ctx, st)
	if err != nil {
		s.SendCancel(st.ID, "wait_ws_ack")
		s.dropStream(st.ID)
		return nil, nil, err
	}
	if frame.Type == TypeStreamError {
		s.dropStream(st.ID)
		return nil, nil, parseStreamError(frame)
	}
	var ack WSOpenAckHeader
	if err := UnmarshalHeader(frame, &ack); err != nil {
		s.dropStream(st.ID)
		return nil, nil, err
	}
	if !ack.OK {
		s.dropStream(st.ID)
		return &ack, nil, fmt.Errorf("本地 WebSocket 握手失败: %s", ack.Message)
	}
	return &ack, st, nil
}

// SendWSMsg 向客户端转发一条 WebSocket 消息。
func (s *Session) SendWSMsg(streamID uint32, binary bool, payload []byte) error {
	return s.writeTyped(TypeWSMsg, streamID, WSMsgHeader{Binary: binary}, payload)
}

// SendWSClose 通知客户端关闭本地 WebSocket。
func (s *Session) SendWSClose(streamID uint32, code int, reason string) error {
	return s.writeTyped(TypeWSClose, streamID, WSCloseHeader{Code: code, Reason: reason}, nil)
}

type streamBody struct {
	*Stream
	session *Session
}

func (b *streamBody) Read(p []byte) (int, error) {
	return b.pr.Read(p)
}

func (b *streamBody) Close() error {
	b.session.dropStream(b.ID)
	return b.pr.Close()
}

type eofReader struct{}

func (eofReader) Read([]byte) (int, error) { return 0, io.EOF }

func parseStreamError(f Frame) error {
	var h StreamErrorHeader
	if err := UnmarshalHeader(f, &h); err != nil {
		return err
	}
	return &streamErr{code: h.Code, msg: h.Message}
}
