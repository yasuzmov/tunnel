package nginx

import (
	"strings"
	"testing"
)

func TestRenderHTTP(t *testing.T) {
	conf, err := Render(RenderInput{
		Domain:            "app.example.com",
		Upstream:          "127.0.0.1:8081",
		ClientMaxBodySize: "20m",
		ACMEWebroot:       "/var/www/acme",
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		"server_name app.example.com;",
		"proxy_pass http://127.0.0.1:8081;",
		"proxy_set_header Host $host;",
		"proxy_set_header Upgrade $http_upgrade;",
		"client_max_body_size 20m;",
	} {
		if !strings.Contains(conf, want) {
			t.Fatalf("missing %q in\n%s", want, conf)
		}
	}
	if strings.Contains(conf, "listen 443") {
		t.Fatal("http template should not listen 443")
	}
}

func TestRenderHTTPS(t *testing.T) {
	conf, err := Render(RenderInput{
		Domain:    "app.example.com",
		Upstream:  "127.0.0.1:8081",
		HTTPS:     true,
		CertReady: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(conf, "listen 443 ssl") {
		t.Fatal("missing 443")
	}
	if !strings.Contains(conf, "return 301 https://$host$request_uri;") {
		t.Fatal("missing http redirect")
	}
}

func TestRenderRejectsBadDomain(t *testing.T) {
	if _, err := Render(RenderInput{Domain: "bad;rm -rf"}); err == nil {
		t.Fatal("expected reject")
	}
}
