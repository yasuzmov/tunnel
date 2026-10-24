package tunnel

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/gorilla/websocket"
)

func (s *Session) writeFrame(f Frame) error {
	raw, err := EncodeFrame(f)
	if err != nil {
		return err
	}
	s.writeM.Lock()
	defer s.writeM.Unlock()
	_ = s.conn.SetWriteDeadline(time.Now().Add(30 * time.Second))
	if err := s.conn.WriteMessage(websocket.BinaryMessage, raw); err != nil {
		return fmt.Errorf("写入隧道帧: %w", err)
	}
	s.BytesOut.Add(int64(len(raw)))
	return nil
}

func (s *Session) writeTyped(typ byte, streamID uint32, header any, payload []byte) error {
	h, err := HeaderFrom(header)
	if err != nil {
		return err
	}
	return s.writeFrame(Frame{Type: typ, StreamID: streamID, Header: h, Payload: payload})
}

// SendWelcome 在 Hello 之后回欢迎帧。
func (s *Session) SendWelcome() error {
	return s.writeTyped(TypeWelcome, 0, WelcomeHeader{
		TunnelID: s.TunnelID,
		Name:     s.TunnelName,
		Domains:  s.Domains,
	}, nil)
}

// SendConfigUpdate 推送最新域名列表。
func (s *Session) SendConfigUpdate(domains []string) {
	s.Domains = append([]string(nil), domains...)
	_ = s.writeTyped(TypeConfigUpdate, 0, ConfigUpdateHeader{Domains: domains}, nil)
}

// SendCancel 通知客户端放弃某条 stream。
func (s *Session) SendCancel(streamID uint32, reason string) {
	_ = s.writeTyped(TypeCancel, streamID, CancelHeader{Reason: reason}, nil)
}

func (s *Session) sendBody(streamID uint32, r io.Reader) error {
	buf := make([]byte, ChunkSize)
	for {
		n, err := r.Read(buf)
		if n > 0 {
			chunk := append([]byte(nil), buf[:n]...)
			if werr := s.writeTyped(TypeBodyChunk, streamID, BodyChunkHeader{EOF: false}, chunk); werr != nil {
				return werr
			}
			s.BytesIn.Add(int64(n))
		}
		if err == io.EOF {
			return s.writeTyped(TypeBodyChunk, streamID, BodyChunkHeader{EOF: true}, nil)
		}
		if err != nil {
			return err
		}
	}
}

func (s *Session) waitHeader(ctx context.Context, st *Stream) (Frame, error) {
	select {
	case <-ctx.Done():
		return Frame{}, ctx.Err()
	case <-s.closed:
		return Frame{}, fmt.Errorf("会话已关闭")
	case err := <-st.errCh:
		return Frame{}, err
	case f := <-st.headerCh:
		return f, nil
	}
}
