package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"wallet/config"
	"wallet/internal/scanner"
)

func main() {
	log.Println("=== Wallet Worker 启动 ===")

	// 1. 加载配置
	cfg, err := config.LoadWithEnv("config/config.yaml")
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	if !cfg.Scanner.Enabled {
		log.Println("扫块功能未启用")
		return
	}

	// 2. 创建上下文（支持优雅退出）
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 3. 启动扫块器
	for _, chain := range cfg.Chains {
		if len(chain.RPCURLs) == 0 {
			log.Printf("跳过链 %s: 没有配置 RPC", chain.Name)
			continue
		}

		log.Printf("启动扫块器: %s (chainID=%d)", chain.Name, chain.ChainID)

		// 创建扫块器
		s, err := scanner.New(scanner.Config{
			RPCUrl:        chain.RPCURLs[0],
			StartBlock:    cfg.Scanner.StartBlock,
			ConfirmBlocks: cfg.Scanner.ConfirmBlocks,
			BatchSize:     cfg.Scanner.BatchSize,
		})
		if err != nil {
			log.Printf("创建扫块器失败: %v", err)
			continue
		}

		// 添加充值处理器（示例）
		depositHandler := scanner.NewDepositHandler(
			[]string{}, // 这里添加需要监控的地址
			func(deposit *scanner.Deposit) {
				// 处理充值逻辑
				log.Printf("💰 新充值: from=%s, amount=%s ETH, tx=%s",
					deposit.From.Hex(),
					weiToEth(deposit.Value),
					deposit.TxHash.Hex(),
				)
				// TODO: 保存到数据库、发送通知等
			},
		)
		s.AddHandler(depositHandler)

		// 启动扫块（在 goroutine 中运行）
		go func(name string, scanner *scanner.Scanner) {
			if err := scanner.Start(ctx, cfg.Scanner.ScanInterval); err != nil {
				log.Printf("扫块器 %s 错误: %v", name, err)
			}
		}(chain.Name, s)
	}

	// 4. 等待退出信号
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	<-sigChan
	log.Println("收到退出信号，正在关闭...")

	cancel()
	time.Sleep(2 * time.Second) // 等待 goroutine 清理

	log.Println("Worker 已关闭")
}

func weiToEth(wei interface{}) string {
	// 简化版本，实际应该使用 big.Int
	
	return "0.00"
}
