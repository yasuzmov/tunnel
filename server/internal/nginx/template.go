package nginx

import (
	"bytes"
	"fmt"
	"text/template"

	"tunnel-server/internal/pkg/validator"
)

// RenderInput 模板变量。Domain 必须已经过 ValidateDomain。
type RenderInput struct {
	Domain            string
	Upstream          string
	ClientMaxBodySize string
	ACMEWebroot       string
	HTTPS             bool
	CertReady         bool
}

const httpTemplate = `server {
    listen 80;
    server_name {{.Domain}};
    client_max_body_size {{.ClientMaxBodySize}};

    location /.well-known/acme-challenge/ {
        root {{.ACMEWebroot}};
    }

    location / {
        proxy_http_version 1.1;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection $connection_upgrade;
        proxy_buffering off;
        proxy_read_timeout 3600s;
        proxy_send_timeout 3600s;
        proxy_pass http://{{.Upstream}};
    }
}
`

const httpsTemplate = `server {
    listen 80;
    server_name {{.Domain}};
    client_max_body_size {{.ClientMaxBodySize}};

    location /.well-known/acme-challenge/ {
        root {{.ACMEWebroot}};
    }

    location / {
        return 301 https://$host$request_uri;
    }
}

server {
    listen 443 ssl;
    http2 on;
    server_name {{.Domain}};
    client_max_body_size {{.ClientMaxBodySize}};

    ssl_certificate /etc/letsencrypt/live/{{.Domain}}/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/{{.Domain}}/privkey.pem;

    location / {
        proxy_http_version 1.1;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection $connection_upgrade;
        proxy_buffering off;
        proxy_read_timeout 3600s;
        proxy_send_timeout 3600s;
        proxy_pass http://{{.Upstream}};
    }
}
`

var (
	httpTmpl  = template.Must(template.New("http").Parse(httpTemplate))
	httpsTmpl = template.Must(template.New("https").Parse(httpsTemplate))
)

// Render 生成单个域名的 nginx 配置文本。
func Render(in RenderInput) (string, error) {
	domain, err := validator.ValidateDomain(in.Domain)
	if err != nil {
		return "", err
	}
	in.Domain = domain
	if in.ClientMaxBodySize == "" {
		in.ClientMaxBodySize = "50m"
	}
	if in.Upstream == "" {
		in.Upstream = "127.0.0.1:8081"
	}
	if in.ACMEWebroot == "" {
		in.ACMEWebroot = "/var/www/acme"
	}
	tpl := httpTmpl
	if in.HTTPS && in.CertReady {
		tpl = httpsTmpl
	}
	var buf bytes.Buffer
	if err := tpl.Execute(&buf, in); err != nil {
		return "", fmt.Errorf("渲染 nginx 模板: %w", err)
	}
	return buf.String(), nil
}
