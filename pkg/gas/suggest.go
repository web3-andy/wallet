package gas

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

// Speed 速度档位
type Speed string

const (
	Slow   Speed = "slow"   // 省钱
	Normal Speed = "normal" // 推荐
	Fast   Speed = "fast"   // 秒进块
)

// GasParams 包含估算的 gas 参数
type GasParams struct {
	GasLimit  uint64
	GasPrice  *big.Int // Legacy 交易使用
	GasTipCap *big.Int // EIP-1559 交易使用
	GasFeeCap *big.Int // EIP-1559 交易使用
	IsLegacy  bool     // 是否为 Legacy 交易类型
}

// SuggestGasParams 自动填充 gas 参数（核心函数）
// 参数:
//   - ctx: 上下文
//   - client: ETH 客户端
//   - from: 发送者地址
//   - to: 接收者地址（可为 nil，表示合约创建）
//   - value: 转账金额
//   - data: 交易数据
//   - speed: 速度档位
func SuggestGasParams(
	ctx context.Context,
	client *ethclient.Client,
	from common.Address,
	to *common.Address,
	value *big.Int,
	data []byte,
	speed Speed,
) (*GasParams, error) {
	// 1. 实时估算 gas used → gasLimit
	gasLimit, err := client.EstimateGas(ctx, ethereum.CallMsg{
		From:  from,
		To:    to,
		Value: value,
		Data:  data,
	})
	if err != nil {
		return nil, fmt.Errorf("estimate gas failed: %w", err)
	}
	// +20% 缓冲
	gasLimit = gasLimit * 120 / 100

	// 2. 获取链 ID
	chainID, err := client.ChainID(ctx)
	if err != nil {
		return nil, fmt.Errorf("get chain ID failed: %w", err)
	}

	// 3. 获取最新费用建议
	suggestedGasTipCap, err := client.SuggestGasTipCap(ctx)
	if err != nil {
		return nil, fmt.Errorf("suggest gas tip cap failed: %w", err)
	}

	suggestedGasFeeCap, err := client.SuggestGasPrice(ctx) // 使用 SuggestGasPrice 更通用
	if err != nil {
		return nil, fmt.Errorf("suggest gas price failed: %w", err)
	}

	// 根据速度档位调整
	var multiplier int64
	switch speed {
	case Slow:
		multiplier = 10 // 1.0x（省钱）
	case Normal:
		multiplier = 11 // 1.1x（推荐）
	case Fast:
		multiplier = 15 // 1.5x（秒进块）
	default:
		multiplier = 11 // 默认 Normal
	}

	// 判断是否是 Legacy 链（BSC / Polygon 等）
	legacyChains := map[int64]bool{
		56:    true, // BSC
		137:   true, // Polygon
		97:    true, // BSC Testnet
		80002: true, // Polygon Amoy Testnet
	}

	if legacyChains[chainID.Int64()] {
		// Legacy 交易：只需要 gasPrice
		gasPrice := new(big.Int).Mul(suggestedGasFeeCap, big.NewInt(multiplier))
		gasPrice.Div(gasPrice, big.NewInt(10))

		// 至少加 1 gwei 保险
		oneGwei := big.NewInt(1e9)
		gasPrice.Add(gasPrice, oneGwei)

		return &GasParams{
			GasLimit: gasLimit,
			GasPrice: gasPrice,
			IsLegacy: true,
		}, nil
	}

	// EIP-1559 交易（ETH, Base, Arbitrum, Optimism, zkSync...）
	tip := new(big.Int).Mul(suggestedGasTipCap, big.NewInt(multiplier))
	tip.Div(tip, big.NewInt(10))

	// 获取当前 base fee（如果可用）
	header, err := client.HeaderByNumber(ctx, nil)
	var baseFee *big.Int
	if err == nil && header.BaseFee != nil {
		baseFee = header.BaseFee
	} else {
		// 如果获取不到 baseFee，使用 suggestedGasFeeCap 作为估算
		baseFee = new(big.Int).Set(suggestedGasFeeCap)
	}

	// feeCap = (baseFee + tip) * multiplier
	feeCap := new(big.Int).Add(baseFee, tip)
	feeCap.Mul(feeCap, big.NewInt(multiplier))
	feeCap.Div(feeCap, big.NewInt(10))

	// 确保 feeCap >= tip
	if feeCap.Cmp(tip) < 0 {
		feeCap = new(big.Int).Set(tip)
	}

	return &GasParams{
		GasLimit:  gasLimit,
		GasTipCap: tip,
		GasFeeCap: feeCap,
		IsLegacy:  false,
	}, nil
}

// CreateTransaction 根据 GasParams 创建交易（未签名）
func CreateTransaction(
	nonce uint64,
	to *common.Address,
	value *big.Int,
	data []byte,
	params *GasParams,
	chainID *big.Int,
) *types.Transaction {
	if params.IsLegacy {
		return types.NewTx(&types.LegacyTx{
			Nonce:    nonce,
			To:       to,
			Value:    value,
			Gas:      params.GasLimit,
			GasPrice: params.GasPrice,
			Data:     data,
		})
	}

	return types.NewTx(&types.DynamicFeeTx{
		ChainID:   chainID,
		Nonce:     nonce,
		To:        to,
		Value:     value,
		Gas:       params.GasLimit,
		GasTipCap: params.GasTipCap,
		GasFeeCap: params.GasFeeCap,
		Data:      data,
	})
}
