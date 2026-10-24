package ingress

import (
	"context"
	"net/http"
	"time"

	"github.com/gorilla/websocket"

	"tunnel-server/internal/logwriter"
	"tunnel-server/internal/model"
	"tunnel-server/internal/pkg/httputil"
	"tunnel-server/internal/tunnel"
)

var publicUpgrader = websocket.Upgrader{
	CheckOrigin:     func(r *http.Request) bool { return true },
	ReadBufferSize:  32 * 1024,
	WriteBufferSize: 32 * 1024,
}

func (s *Server) proxyWS(w http.ResponseWriter, r *http.Request, sess *tunnel.Session, rt Route) {
	item := newLog(r, rt.TunnelID, rt.Domain, sessionPtr(sess))
	item.IsWebsocket = true
	now := time.Now()
	item.ForwardedAt = &now
	fwd := httputil.FilterHopByHop(r.Header, false)
	ctx, cancel := context.WithTimeout(r.Context(), s.ResponseTimeout)
	ack, st, err := sess.OpenWS(ctx, tunnel.WSOpenHeader{
		Scheme:  wsScheme(item.Scheme),
		Host:    httputil.CanonicalHost(r.Host),
		Path:    r.URL.Path,
		Query:   r.URL.RawQuery,
		Headers: httputil.HeaderToMap(fwd),
	})
	cancel()
	if err != nil {
		writeBadGateway(w, "无法建立本地 WebSocket")
		item.Status = http.StatusBadGateway
		finishLog(item, err.Error())
		s.Logs.Enqueue(entryOnly(item))
		return
	}
	item.LocalTarget = ack.LocalTarget
	item.MatchedRule = ack.MatchedRule
	respHeader := httputil.FilterHopByHop(httputil.MapToHeader(ack.Headers), false)
	item.ResponseHeaders = mustJSON(httputil.HeaderToMap(respHeader))
	conn, err := publicUpgrader.Upgrade(w, r, respHeader)
	if err != nil {
		sess.SendWSClose(st.ID, 1011, "upgrade_failed")
		sess.DropStream(st.ID)
		item.Status = http.StatusBadGateway
		finishLog(item, err.Error())
		s.Logs.Enqueue(entryOnly(item))
		return
	}
	defer conn.Close()
	item.Status = http.StatusSwitchingProtocols
	fb := time.Now()
	item.FirstByteAt = &fb
	count := s.relayWS(conn, sess, st)
	item.WSMessageCount = count
	finishLog(item, "")
	s.Logs.Enqueue(entryOnly(item))
}

func (s *Server) relayWS(conn *websocket.Conn, sess *tunnel.Session, st *tunnel.Stream) int64 {
	var n int64
	done := make(chan struct{})
	go func() {
		defer close(done)
		for {
			typ, data, err := conn.ReadMessage()
			if err != nil {
				_ = sess.SendWSClose(st.ID, 1000, "public_closed")
				return
			}
			n++
			_ = sess.SendWSMsg(st.ID, typ == websocket.BinaryMessage, data)
		}
	}()
	go func() {
		for {
			f, err := sess.RecvWS(st)
			if err != nil {
				_ = conn.WriteControl(websocket.CloseMessage, websocket.FormatCloseMessage(1000, ""), time.Now().Add(3*time.Second))
				_ = conn.Close()
				return
			}
			if f.Type == tunnel.TypeWSClose {
				_ = conn.Close()
				return
			}
			if f.Type != tunnel.TypeWSMsg {
				continue
			}
			var h tunnel.WSMsgHeader
			_ = tunnel.UnmarshalHeader(f, &h)
			typ := websocket.TextMessage
			if h.Binary {
				typ = websocket.BinaryMessage
			}
			n++
			if err := conn.WriteMessage(typ, f.Payload); err != nil {
				return
			}
		}
	}()
	<-done
	sess.DropStream(st.ID)
	return n
}

func wsScheme(httpScheme string) string {
	if httpScheme == "https" {
		return "wss"
	}
	return "ws"
}

func entryOnly(item *model.RequestLog) *logwriter.Entry {
	return &logwriter.Entry{Log: item, Body: &model.RequestLogBody{}}
}
