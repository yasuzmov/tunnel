package httputil

import (
	"net/http"
	"testing"
)

func TestFilterHopByHop(t *testing.T) {
	h := http.Header{}
	h.Set("Host", "a.com")
	h.Set("Connection", "keep-alive, X-Foo")
	h.Set("X-Foo", "1")
	h.Set("Upgrade", "websocket")
	h.Set("User-Agent", "curl")
	out := FilterHopByHop(h, true)
	if out.Get("Upgrade") != "" || out.Get("X-Foo") != "" || out.Get("Connection") != "" {
		t.Fatalf("hop-by-hop not stripped: %v", out)
	}
	if out.Get("User-Agent") != "curl" {
		t.Fatalf("lost user-agent")
	}
}

func TestFilterKeepUpgrade(t *testing.T) {
	h := http.Header{}
	h.Set("Upgrade", "websocket")
	out := FilterHopByHop(h, false)
	if out.Get("Upgrade") != "websocket" {
		t.Fatal("upgrade should be kept")
	}
}

func TestIsTextContent(t *testing.T) {
	if !IsTextContent("application/json; charset=utf-8") {
		t.Fatal("json should be text")
	}
	if IsTextContent("image/png") {
		t.Fatal("png should not be text")
	}
}
