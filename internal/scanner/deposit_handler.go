package scanner

import (
	"context"
	"log"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

// DepositHandler 充值处理器
type DepositHandler struct {
	watchAddresses map[common.Address]bool // 监控的地址
	callback       func(deposit *Deposit)   // 充值回调
}

// Deposit 充值信息
type Deposit struct {
	TxHash      common.Hash
	BlockNumber uint64
	From        common.Address
	To          common.Address
	Value       *big.Int
	Status      uint64 // 1=成功, 0=失败
}

// NewDepositHandler 创建充值处理器
func NewDepositHandler(addresses []string, callback func(*Deposit)) *DepositHandler {
	watchMap := make(map[common.Address]bool)
	for _, addr := range addresses {
		watchMap[common.HexToAddress(addr)] = true
	}

	return &DepositHandler{
		watchAddresses: watchMap,
		callback:       callback,
	}
}

// AddWatchAddress 添加监控地址
func (h *DepositHandler) AddWatchAddress(addr string) {
	h.watchAddresses[common.HexToAddress(addr)] = true
}

// HandleBlock 处理区块
func (h *DepositHandler) HandleBlock(ctx context.Context, block *types.Block) error {
	// 可以在这里处理区块级别的逻辑
	return nil
}

// HandleTransaction 处理交易
func (h *DepositHandler) HandleTransaction(ctx context.Context, tx *types.Transaction, receipt *types.Receipt) error {
	// 只处理监控地址的交易
	if tx.To() == nil {
		return nil // 合约创建交易
	}

	if !h.watchAddresses[*tx.To()] {
		return nil // 不是我们监控的地址
	}

	// 只处理原生代币转账（value > 0）
	if tx.Value().Cmp(big.NewInt(0)) == 0 {
		return nil
	}

	// 提取发送者地址
	from, err := types.Sender(types.LatestSignerForChainID(tx.ChainId()), tx)
	if err != nil {
		log.Printf("提取发送者地址失败: %v", err)
		return err
	}

	deposit := &Deposit{
		TxHash:      tx.Hash(),
		BlockNumber: receipt.BlockNumber.Uint64(),
		From:        from,
		To:          *tx.To(),
		Value:       tx.Value(),
		Status:      receipt.Status,
	}

	log.Printf("检测到充值: from=%s, to=%s, value=%s ETH, tx=%s",
		from.Hex(),
		tx.To().Hex(),
		weiToEth(tx.Value()),
		tx.Hash().Hex(),
	)

	// 调用回调
	if h.callback != nil {
		h.callback(deposit)
	}

	return nil
}

// weiToEth 工具函数
func weiToEth(wei *big.Int) string {
	eth := new(big.Float).Quo(
		new(big.Float).SetInt(wei),
		big.NewFloat(1e18),
	)
	return eth.Text('f', 6)
}
