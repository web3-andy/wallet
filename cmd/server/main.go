package main

import (
	"fmt"
	"log"
	"net/http"

	"wallet/api/http/handler"
	"wallet/config"
)

func main() {
	log.Println("=== Wallet API Server 启动 ===")

	// 1. 加载配置
	cfg, err := config.LoadWithEnv("config/config.yaml")
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	// 2. 设置路由
	mux := http.NewServeMux()

	// 健康检查
	mux.HandleFunc("/health", handler.Health)

	// API 路由（示例）
	mux.HandleFunc("/api/v1/balance", handler.GetBalance)
	mux.HandleFunc("/api/v1/transfer", handler.Transfer)
	mux.HandleFunc("/api/v1/transactions", handler.GetTransactions)

	// 3. 启动服务器
	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	log.Printf("API 服务器运行在: http://%s", addr)
	log.Printf("健康检查: http://%s/health", addr)

	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("服务器启动失败: %v", err)
	}
}
