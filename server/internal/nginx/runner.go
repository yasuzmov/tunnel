// Package nginx 负责渲染、落盘、校验并重载 nginx 配置，以及调用 certbot。
// 外部命令一律走 CommandRunner，参数用数组传入，禁止拼接进 shell。
package nginx

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
)

// Result 命令输出，便于把 nginx -t 的错误原文回显到后台。
type Result struct {
	Stdout string
	Stderr string
}

// CommandRunner 可替换的命令执行器，单测用内存实现。
type CommandRunner interface {
	Run(ctx context.Context, name string, args ...string) (Result, error)
}

// ExecRunner 真实执行系统命令。
type ExecRunner struct{}

// Run 在 ctx 取消时杀掉子进程。
func (ExecRunner) Run(ctx context.Context, name string, args ...string) (Result, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	var out, errb bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &errb
	err := cmd.Run()
	res := Result{Stdout: out.String(), Stderr: errb.String()}
	if err != nil {
		return res, fmt.Errorf("%s: %w: %s", name, err, res.Stderr)
	}
	return res, nil
}
