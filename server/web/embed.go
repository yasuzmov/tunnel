// Package web 内嵌管理后台静态资源。
// 构建流程：先在仓库 web/ 目录执行 npm run build，产物输出到本目录 dist/，
// 再编译服务端，即可得到单二进制发布包。
package web

import "embed"

// DistFS 包含管理后台构建产物。embed 要求目录在编译时存在，
// 因此仓库中保留一份占位 index.html，避免未构建前端时 go build 失败。
//
//go:embed all:dist
var DistFS embed.FS
