# 生产部署

目标环境：Linux 服务器，nginx 与本服务同机。管理进程监听 `:8080`，公网入口只绑 `127.0.0.1:8081`。

## 1. 准备

```bash
sudo useradd -r -s /usr/sbin/nologin tunnel
sudo mkdir -p /opt/tunnel /etc/nginx/conf.d/tunnel /var/www/acme
sudo chown -R tunnel:tunnel /opt/tunnel
```

把 `tunnel-server` 二进制和 `config.yaml` 放到 `/opt/tunnel/`。DSN、`jwt.secret`、bootstrap 口令务必改掉。

`config.yaml` 中：

- `nginx.enabled: true`
- `certbot.enabled: true` 并填写邮箱
- `admin.public_url` 写成客户端能访问的 `https://admin.example.com`

## 2. nginx

把 [nginx-main-snippet.conf](nginx-main-snippet.conf) 中的 `map` 与 `include` 放进 `http {}`。

管理域名可参考 [nginx-admin.conf](nginx-admin.conf)，这样客户端走 `wss://`。

服务端写入的文件位于 `/etc/nginx/conf.d/tunnel/{domain}.conf`。进程用户需要对这个目录可写，并对 `nginx -t` / `nginx -s reload` 有权限。可选 [sudoers.tunnel](sudoers.tunnel)。若走 sudo，把配置里的 `nginx_bin` / `certbot.bin` 改成包装脚本。

## 3. systemd

```bash
sudo cp deploy/tunnel-server.service /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl enable --now tunnel-server
```

证书续期交给 certbot 自带 timer，无需本服务参与。

## 4. 本机联调

Windows 开发保持 `nginx.enabled: false`。创建记录、客户端连通后，用：

```bash
curl -H "Host: test.local" http://127.0.0.1:8081/
```

直接打入口。后台仍可预览将要生成的 nginx 配置。
