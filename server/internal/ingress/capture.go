package ingress

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"tunnel-server/internal/model"
	"tunnel-server/internal/pkg/httputil"
	"tunnel-server/internal/pkg/idgen"
)

// capBuf 只保留前 limit 字节，同时统计总大小。
type capBuf struct {
	buf       []byte
	limit     int
	truncated bool
	total     int64
}

func (c *capBuf) Write(p []byte) (int, error) {
	c.total += int64(len(p))
	remain := c.limit - len(c.buf)
	if remain <= 0 {
		if len(p) > 0 {
			c.truncated = true
		}
		return len(p), nil
	}
	if len(p) > remain {
		c.buf = append(c.buf, p[:remain]...)
		c.truncated = true
		return len(p), nil
	}
	c.buf = append(c.buf, p...)
	return len(p), nil
}

func newLog(r *http.Request, tunnelID uint, domain string, sessionID *uint) *model.RequestLog {
	now := time.Now()
	query := r.URL.RawQuery
	return &model.RequestLog{
		RequestID:      idgen.NewRequestID(),
		TunnelID:       tunnelID,
		SessionID:      sessionID,
		Domain:         domain,
		Method:         r.Method,
		Scheme:         httputil.RequestScheme(r),
		Host:           r.Host,
		Path:           r.URL.Path,
		QueryString:    query,
		QueryParams:    mustJSON(flattenQuery(r.URL.Query())),
		RemoteIP:       httputil.ClientIP(r),
		UserAgent:      r.UserAgent(),
		Referer:        r.Referer(),
		RequestHeaders: mustJSON(httputil.HeaderToMap(r.Header)),
		ReceivedAt:     now,
	}
}

func flattenQuery(v url.Values) map[string]any {
	out := make(map[string]any, len(v))
	for k, vals := range v {
		if len(vals) == 1 {
			out[k] = vals[0]
			continue
		}
		out[k] = vals
	}
	return out
}

func parseFormParams(contentType string, body []byte) string {
	ct := strings.ToLower(contentType)
	if strings.Contains(ct, "application/x-www-form-urlencoded") {
		vals, err := url.ParseQuery(string(body))
		if err != nil {
			return ""
		}
		return mustJSON(flattenQuery(vals))
	}
	if strings.Contains(ct, "application/json") {
		var obj any
		if json.Unmarshal(body, &obj) == nil {
			return mustJSON(obj)
		}
	}
	return ""
}

func mustJSON(v any) string {
	b, err := json.Marshal(v)
	if err != nil {
		return "{}"
	}
	return string(b)
}

func finishLog(item *model.RequestLog, errMsg string) {
	now := time.Now()
	item.CompletedAt = &now
	item.DurationMS = now.Sub(item.ReceivedAt).Milliseconds()
	if item.ForwardedAt != nil {
		item.UpstreamMS = now.Sub(*item.ForwardedAt).Milliseconds()
	}
	item.ErrorMessage = errMsg
}

func teeRead(src io.ReadCloser, cap *capBuf) io.ReadCloser {
	if src == nil {
		return http.NoBody
	}
	return &teeCloser{Reader: io.TeeReader(src, cap), c: src}
}

type teeCloser struct {
	io.Reader
	c io.Closer
}

func (t *teeCloser) Close() error {
	if t.c != nil {
		return t.c.Close()
	}
	return nil
}

func textOrEmpty(ct string, buf *capBuf) []byte {
	if !httputil.IsTextContent(ct) {
		return nil
	}
	return bytes.Clone(buf.buf)
}
