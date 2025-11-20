package transfer

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"wallet/pkg/gas"
)

// Transfer 转账管理器
type Transfer struct {
	client *ethclient.Client
}

// Request 转账请求
type Request struct {
	From       common.Address
	PrivateKey *ecdsa.PrivateKey
	To         common.Address
	Amount     *big.Int // Wei
	Speed      gas.Speed
	Data       []byte // 可选，合约调用数据
}

// Result 转账结果
type Result struct {
	TxHash      common.Hash
	BlockNumber uint64
	GasUsed     uint64
	Success     bool
}

// New 创建转账管理器
func New(rpcURL string) (*Transfer, error) {
	client, err := ethclient.Dial(rpcURL)
	if err != nil {
		return nil, fmt.Errorf("connect to rpc: %w", err)
	}

	return &Transfer{client: client}, nil
}

// Execute 执行转账
func (t *Transfer) Execute(ctx context.Context, req Request) (*Result, error) {
	// 1. 验证私钥和地址匹配
	publicKey := req.PrivateKey.Public().(*ecdsa.PublicKey)
	derivedAddr := crypto.PubkeyToAddress(*publicKey)
	if derivedAddr != req.From {
		return nil, fmt.Errorf("私钥和发送地址不匹配")
	}

	// 2. 检查余额
	balance, err := t.client.BalanceAt(ctx, req.From, nil)
	if err != nil {
		return nil, fmt.Errorf("获取余额失败: %w", err)
	}

	if balance.Cmp(req.Amount) < 0 {
		return nil, fmt.Errorf("余额不足: 需要 %s, 当前 %s",
			weiToEth(req.Amount), weiToEth(balance))
	}

	// 3. 获取 nonce
	nonce, err := t.client.PendingNonceAt(ctx, req.From)
	if err != nil {
		return nil, fmt.Errorf("获取 nonce 失败: %w", err)
	}

	// 4. 估算 gas
	params, err := gas.SuggestGasParams(
		ctx,
		t.client,
		req.From,
		&req.To,
		req.Amount,
		req.Data,
		req.Speed,
	)
	if err != nil {
		return nil, fmt.Errorf("估算 gas 失败: %w", err)
	}

	// 5. 获取链 ID
	chainID, err := t.client.ChainID(ctx)
	if err != nil {
		return nil, fmt.Errorf("获取链 ID 失败: %w", err)
	}

	// 6. 创建交易
	tx := gas.CreateTransaction(nonce, &req.To, req.Amount, req.Data, params, chainID)

	// 7. 签名
	signedTx, err := types.SignTx(tx, types.LatestSignerForChainID(chainID), req.PrivateKey)
	if err != nil {
		return nil, fmt.Errorf("签名失败: %w", err)
	}

	// 8. 发送交易
	if err := t.client.SendTransaction(ctx, signedTx); err != nil {
		return nil, fmt.Errorf("发送交易失败: %w", err)
	}

	// 9. 等待确认（可选）
	receipt, err := t.waitForReceipt(ctx, signedTx.Hash())
	if err != nil {
		return nil, fmt.Errorf("等待确认失败: %w", err)
	}

	return &Result{
		TxHash:      signedTx.Hash(),
		BlockNumber: receipt.BlockNumber.Uint64(),
		GasUsed:     receipt.GasUsed,
		Success:     receipt.Status == types.ReceiptStatusSuccessful,
	}, nil
}

// waitForReceipt 等待交易确认
func (t *Transfer) waitForReceipt(ctx context.Context, txHash common.Hash) (*types.Receipt, error) {
	return t.client.TransactionReceipt(ctx, txHash)
}

// GetBalance 查询余额
func (t *Transfer) GetBalance(ctx context.Context, addr common.Address) (*big.Int, error) {
	return t.client.BalanceAt(ctx, addr, nil)
}

// Close 关闭连接
func (t *Transfer) Close() {
	if t.client != nil {
		t.client.Close()
	}
}

// 工具函数
func weiToEth(wei *big.Int) string {
	eth := new(big.Float).Quo(
		new(big.Float).SetInt(wei),
		big.NewFloat(1e18),
	)
	return eth.Text('f', 6)
}
