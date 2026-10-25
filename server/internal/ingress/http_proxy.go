package ingress

import (
	"context"
	"io"
	"net/http"
	"strings"
	"time"

	"tunnel-server/internal/logwriter"
	"tunnel-server/internal/model"
	"tunnel-server/internal/pkg/httputil"
	"tunnel-server/internal/tunnel"
)

func (s *Server) proxyHTTP(w http.ResponseWriter, r *http.Request, sess *tunnel.Session, rt Route) {
	limit := s.BodyLimit
	reqCap := &capBuf{limit: limit}
	body := teeRead(r.Body, reqCap)
	defer body.Close()

	item := newLog(r, rt.TunnelID, rt.Domain, sessionPtr(sess))
	item.FormParams = ""
	fwd := httputil.FilterHopByHop(r.Header, true)
	req := tunnel.HTTPReqHeader{
		Method:        r.Method,
		Scheme:        item.Scheme,
		Host:          httputil.CanonicalHost(r.Host),
		Path:          r.URL.Path,
		Query:         r.URL.RawQuery,
		Headers:       httputil.HeaderToMap(fwd),
		ContentLength: r.ContentLength,
	}

	ctx, cancel := context.WithTimeout(r.Context(), s.ResponseTimeout)
	defer cancel()
	now := time.Now()
	item.ForwardedAt = &now
	resp, respBody, err := sess.OpenHTTP(ctx, req, body)
	if err != nil {
		writeBadGateway(w, "上游无响应")
		item.Status = http.StatusBadGateway
		finishLog(item, err.Error())
		s.enqueue(rt.LogEnabled, item, reqCap, r.Header.Get("Content-Type"), nil, "")
		return
	}
	defer respBody.Close()
	s.writeHTTPResponse(w, r, rt, item, reqCap, resp, respBody)
}

func (s *Server) writeHTTPResponse(w http.ResponseWriter, r *http.Request, rt Route, item *model.RequestLog, reqCap *capBuf, resp *tunnel.HTTPRespHeader, respBody io.ReadCloser) {
	fb := time.Now()
	item.FirstByteAt = &fb
	item.Status = resp.Status
	item.LocalTarget = resp.LocalTarget
	item.MatchedRule = resp.MatchedRule
	outHeader := httputil.FilterHopByHop(httputil.MapToHeader(resp.Headers), true)
	item.ResponseHeaders = mustJSON(httputil.HeaderToMap(outHeader))
	copyHeader(w.Header(), outHeader)
	if resp.Status <= 0 {
		resp.Status = http.StatusOK
	}
	w.WriteHeader(resp.Status)
	respCap := &capBuf{limit: s.BodyLimit}
	n, err := io.Copy(io.MultiWriter(flushWriter{w}, respCap), respBody)
	item.ResponseSize = n
	item.RequestSize = reqCap.total
	item.FormParams = parseFormParams(r.Header.Get("Content-Type"), reqCap.buf)
	msg := ""
	if err != nil {
		msg = err.Error()
	}
	finishLog(item, msg)
	s.enqueue(rt.LogEnabled, item, reqCap, r.Header.Get("Content-Type"), respCap, outHeader.Get("Content-Type"))
}

func (s *Server) enqueue(enabled bool, item *model.RequestLog, reqCap *capBuf, reqCT string, respCap *capBuf, respCT string) {
	if !enabled {
		return
	}
	body := &model.RequestLogBody{
		RequestContentType:  reqCT,
		RequestTruncated:    reqCap.truncated,
		RequestBody:         textOrEmpty(reqCT, reqCap),
		ResponseContentType: respCT,
	}
	if respCap != nil {
		body.ResponseTruncated = respCap.truncated
		body.ResponseBody = textOrEmpty(respCT, respCap)
		if item.ResponseSize == 0 {
			item.ResponseSize = respCap.total
		}
	}
	if !httputil.IsTextContent(reqCT) {
		body.RequestTruncated = false
	}
	if !httputil.IsTextContent(respCT) {
		body.ResponseTruncated = false
	}
	s.Logs.Enqueue(&logwriter.Entry{Log: item, Body: body})
}

func copyHeader(dst, src http.Header) {
	for k, vs := range src {
		for _, v := range vs {
			dst.Add(k, v)
		}
	}
}

func sessionPtr(sess *tunnel.Session) *uint {
	if sess == nil || sess.ID == 0 {
		return nil
	}
	id := sess.ID
	return &id
}

func writeBadGateway(w http.ResponseWriter, msg string) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusBadGateway)
	_, _ = io.WriteString(w, "<!doctype html><meta charset=utf-8><title>502</title><h1>502 Bad Gateway</h1><p>"+htmlEscape(msg)+"</p>")
}

func htmlEscape(s string) string {
	r := strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;")
	return r.Replace(s)
}

type flushWriter struct{ http.ResponseWriter }

func (f flushWriter) Write(p []byte) (int, error) {
	n, err := f.ResponseWriter.Write(p)
	if fl, ok := f.ResponseWriter.(http.Flusher); ok {
		fl.Flush()
	}
	return n, err
}
