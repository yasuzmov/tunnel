# 交叉编译 Linux amd64 服务端。先构建管理后台再打进单二进制。
$ErrorActionPreference = "Stop"
$root = Split-Path -Parent $PSScriptRoot
Set-Location "$root\web"
npm install
npm run build
Set-Location "$root\server"
$env:GOTOOLCHAIN = "local"
$env:CGO_ENABLED = "0"
$env:GOOS = "linux"
$env:GOARCH = "amd64"
New-Item -ItemType Directory -Force -Path "$root\bin" | Out-Null
go build -o "$root\bin\tunnel-server" .
Write-Host "已输出 $root\bin\tunnel-server"
