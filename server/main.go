// 内网穿透服务端入口。
// 负责加载配置并启动管理 API（含隧道接入）与公网流量入口两个监听端口。
package main

import (
	"context"
	"errors"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"tunnel-server/internal/app"
	"tunnel-server/internal/config"
)

func main() {
	path := flag.String("config", "config.yaml", "YAML 配置文件路径")
	flag.Parse()
	cfg, err := config.Load(*path)
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}
	instance, err := app.New(cfg)
	if err != nil {
		log.Fatalf("初始化失败: %v", err)
	}

	errCh := make(chan error, 1)
	go func() {
		errCh <- instance.Run()
	}()
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(sigCh)

	select {
	case err := <-errCh:
		if err != nil {
			log.Fatalf("服务退出: %v", err)
		}
	case sig := <-sigCh:
		log.Printf("收到信号 %s，开始关闭", sig)
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := instance.Shutdown(ctx); err != nil && !errors.Is(err, context.Canceled) {
			log.Printf("关闭未完成: %v", err)
		}
	}
}
