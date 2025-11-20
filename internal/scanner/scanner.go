package scanner

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

// Scanner 区块扫描器
type Scanner struct {
	client        *ethclient.Client
	chainID       *big.Int
	startBlock    uint64
	confirmBlocks uint64
	batchSize     int
	handlers      []Handler
}

// Handler 区块处理器接口
type Handler interface {
	HandleBlock(ctx context.Context, block *types.Block) error
	HandleTransaction(ctx context.Context, tx *types.Transaction, receipt *types.Receipt) error
}

// Config 扫描器配置
type Config struct {
	RPCUrl        string
	StartBlock    uint64 // 0 表示最新区块
	ConfirmBlocks uint64 // 确认区块数
	BatchSize     int
}

// New 创建扫描器
func New(cfg Config) (*Scanner, error) {
	client, err := ethclient.Dial(cfg.RPCUrl)
	if err != nil {
		return nil, fmt.Errorf("connect to rpc: %w", err)
	}

	chainID, err := client.ChainID(context.Background())
	if err != nil {
		return nil, fmt.Errorf("get chain id: %w", err)
	}

	startBlock := cfg.StartBlock
	if startBlock == 0 {
		// 从最新区块开始
		latest, err := client.BlockNumber(context.Background())
		if err != nil {
			return nil, fmt.Errorf("get latest block: %w", err)
		}
		startBlock = latest
	}

	return &Scanner{
		client:        client,
		chainID:       chainID,
		startBlock:    startBlock,
		confirmBlocks: cfg.ConfirmBlocks,
		batchSize:     cfg.BatchSize,
		handlers:      []Handler{},
	}, nil
}

// AddHandler 添加处理器
func (s *Scanner) AddHandler(h Handler) {
	s.handlers = append(s.handlers, h)
}

// Start 开始扫描
func (s *Scanner) Start(ctx context.Context, interval time.Duration) error {
	log.Printf("扫块器启动: chainID=%s, startBlock=%d", s.chainID, s.startBlock)

	currentBlock := s.startBlock
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("扫块器停止")
			return ctx.Err()

		case <-ticker.C:
			// 获取最新区块
			latestBlock, err := s.client.BlockNumber(ctx)
			if err != nil {
				log.Printf("获取最新区块失败: %v", err)
				continue
			}

			// 计算需要扫描的区块（减去确认数）
			confirmedBlock := latestBlock - s.confirmBlocks
			if currentBlock > confirmedBlock {
				// 还没有新的已确认区块
				continue
			}

			// 批量扫描
			endBlock := currentBlock + uint64(s.batchSize)
			if endBlock > confirmedBlock {
				endBlock = confirmedBlock
			}

			log.Printf("扫描区块: %d -> %d", currentBlock, endBlock)

			for blockNum := currentBlock; blockNum <= endBlock; blockNum++ {
				if err := s.scanBlock(ctx, blockNum); err != nil {
					log.Printf("扫描区块 %d 失败: %v", blockNum, err)
					// 继续扫描下一个区块
				}
			}

			currentBlock = endBlock + 1
		}
	}
}

// scanBlock 扫描单个区块
func (s *Scanner) scanBlock(ctx context.Context, blockNum uint64) error {
	// 获取区块详情
	block, err := s.client.BlockByNumber(ctx, big.NewInt(int64(blockNum)))
	if err != nil {
		return fmt.Errorf("get block %d: %w", blockNum, err)
	}

	// 调用区块处理器
	for _, handler := range s.handlers {
		if err := handler.HandleBlock(ctx, block); err != nil {
			log.Printf("处理区块 %d 失败: %v", blockNum, err)
		}
	}

	// 处理区块中的每笔交易
	for _, tx := range block.Transactions() {
		// 获取交易收据
		receipt, err := s.client.TransactionReceipt(ctx, tx.Hash())
		if err != nil {
			log.Printf("获取交易收据失败 %s: %v", tx.Hash(), err)
			continue
		}

		// 调用交易处理器
		for _, handler := range s.handlers {
			if err := handler.HandleTransaction(ctx, tx, receipt); err != nil {
				log.Printf("处理交易 %s 失败: %v", tx.Hash(), err)
			}
		}
	}

	return nil
}

// Close 关闭扫描器
func (s *Scanner) Close() {
	if s.client != nil {
		s.client.Close()
	}
}
