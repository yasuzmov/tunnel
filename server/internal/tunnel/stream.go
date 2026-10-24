package tunnel

import (
	"io"
	"sync"
)

// Stream 一条在隧道上复用的逻辑连接。
// HTTP 用 headerCh 等响应头，body 走 pipe；WebSocket 用 msgCh 收后续帧。
type Stream struct {
	ID       uint32
	headerCh chan Frame
	msgCh    chan Frame
	errCh    chan error
	pr       *io.PipeReader
	pw       *io.PipeWriter
	once     sync.Once
}

func newStream(id uint32) *Stream {
	pr, pw := io.Pipe()
	return &Stream{
		ID:       id,
		headerCh: make(chan Frame, 1),
		msgCh:    make(chan Frame, 256),
		errCh:    make(chan error, 1),
		pr:       pr,
		pw:       pw,
	}
}

func (s *Stream) fail(err error) {
	s.once.Do(func() {
		_ = s.pw.CloseWithError(err)
		select {
		case s.errCh <- err:
		default:
		}
	})
}

func (s *Stream) closeBody() {
	s.once.Do(func() {
		_ = s.pw.Close()
	})
}

func (s *Stream) Body() io.ReadCloser {
	return s.pr
}
