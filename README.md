# 内网穿透工具

将内网 HTTP/HTTPS（含 WebSocket）服务通过公网域名暴露出去。服务端负责隧道接入、管理后台、请求日志与 nginx/certbot 自动化；Windows 客户端负责把公网请求转发到本机端口或 hosts 中配置的域名。

## 目录

- `server/`：Gin + Gorm + MySQL 服务端，内嵌管理后台静态资源
- `web/`：Vite + Vue3 + Element Plus 管理后台源码
- `client/`：Flutter Windows 客户端
- `deploy/`：systemd、nginx 示例与部署说明
- `scripts/`：交叉编译脚本

## 快速开始（本地开发）

服务端默认 `nginx.enabled=false`，可在 Windows 上直接联调，不必安装 nginx。

1. 创建 MySQL 数据库 `tunnel`
2. 复制 `deploy/config.example.yaml` 为 `server/config.yaml` 并填写 DSN
3. 构建后台：`cd web && npm install && npm run build`
4. 启动服务端：`cd server && go run . -config config.yaml`
5. 启动客户端：`cd client && flutter run -d windows`（仓库路径含中文时请用 `scripts/build-client.ps1` 编译，MSBuild 无法直接处理非 ASCII 路径）

公网入口默认监听 `127.0.0.1:8081`，可用：

```bash
curl -H "Host: test.local" http://127.0.0.1:8081/
```

模拟访问。客户端连通后，服务端会自动刷新联通记录。

## 域名管理与权限

- 域名统一挂在固定主域名下（生产为 `fish.czypro.cn`，DNS 已配置泛解析 `*.fish.czypro.cn`），由配置 `admin.domain_suffix` 指定，留空时取 `admin.public_url` 的主机名。
- 「域名管理」中创建/编辑域名只能填写前缀，后缀不可修改；普通管理员只能看到、修改、删除自己创建的域名，超级管理员可查看全部。
- 穿透记录绑定域名时，普通管理员只能选择自己创建、且未被使用的域名；超级管理员可选任意未使用的域名。服务端强制校验。
- 原「管理员管理 → 域名权限」分配功能已移除；旧的 `admin_domains` 表保留不删。升级前已有的域名 `created_by=0`，仅超级管理员可见。

功能清单与协议、API、日志字段等技术细节见 [功能与技术文档.md](功能与技术文档.md)。生产部署请参阅 [deploy/README.md](deploy/README.md)。
