package main

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"log"
	"math/big"

	"wallet/pkg/gas"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

func main() {
	// 选择你要测试的功能
	// 取消注释你想运行的示例：

	exampleEstimateGas()
	exampleSendTransaction()
	exampleCompareSpeed()
}

// 示例1：仅估算 gas 参数（不发送交易）
func exampleEstimateGas() {
	fmt.Println("=== 估算 Gas 参数示例 ===")
	fmt.Println()

	// 连接到以太坊主网（你可以换成其他 RPC）
	client, err := ethclient.Dial("https://rpc.ankr.com/eth")
	if err != nil {
		log.Fatal("连接节点失败:", err)
	}
	defer client.Close()

	ctx := context.Background()

	// 模拟交易参数
	from := common.HexToAddress("0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb")
	to := common.HexToAddress("0x1234567890123456789012345678901234567890")
	value := big.NewInt(1e17) // 0.1 ETH
	data := []byte{}

	// 获取 gas 参数建议
	params, err := gas.SuggestGasParams(
		ctx,
		client,
		from,
		&to,
		value,
		data,
		gas.Normal, // 推荐速度
	)
	if err != nil {
		log.Fatal("估算 gas 失败:", err)
	}

	// 打印结果
	fmt.Printf("Gas Limit: %d\n", params.GasLimit)
	fmt.Printf("交易类型: %s\n", map[bool]string{true: "Legacy", false: "EIP-1559"}[params.IsLegacy])

	if params.IsLegacy {
		fmt.Printf("Gas Price: %s Gwei\n", weiToGwei(params.GasPrice))
		fmt.Printf("预估费用: %s ETH\n", weiToEth(new(big.Int).Mul(params.GasPrice, big.NewInt(int64(params.GasLimit)))))
	} else {
		fmt.Printf("Max Priority Fee (Tip): %s Gwei\n", weiToGwei(params.GasTipCap))
		fmt.Printf("Max Fee: %s Gwei\n", weiToGwei(params.GasFeeCap))
		fmt.Printf("预估最高费用: %s ETH\n", weiToEth(new(big.Int).Mul(params.GasFeeCap, big.NewInt(int64(params.GasLimit)))))
	}
}

// 示例2：完整流程 - 发送真实交易（需要私钥）
func exampleSendTransaction() {
	fmt.Println("=== 发送交易示例 ===")
	fmt.Println()

	// ⚠️ 警告：这会发送真实交易！请确保你知道自己在做什么
	privateKeyHex := "YOUR_PRIVATE_KEY_HERE" // 替换成你的私钥（不要提交到 git！）

	if privateKeyHex == "YOUR_PRIVATE_KEY_HERE" {
		log.Fatal("❌ 请先设置你的私钥！")
	}

	// 连接节点
	client, err := ethclient.Dial("https://rpc.ankr.com/eth_sepolia") // 使用测试网
	if err != nil {
		log.Fatal("连接节点失败:", err)
	}
	defer client.Close()

	ctx := context.Background()

	// 加载私钥
	privateKey, err := crypto.HexToECDSA(privateKeyHex)
	if err != nil {
		log.Fatal("私钥格式错误:", err)
	}

	publicKey := privateKey.Public().(*ecdsa.PublicKey)
	from := crypto.PubkeyToAddress(*publicKey)
	to := common.HexToAddress("0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb")
	value := big.NewInt(1e15) // 0.001 ETH (测试网)
	data := []byte{}

	fmt.Printf("发送地址: %s\n", from.Hex())
	fmt.Printf("接收地址: %s\n", to.Hex())
	fmt.Printf("金额: 0.001 ETH\n\n")

	// 1. 估算 gas
	params, err := gas.SuggestGasParams(ctx, client, from, &to, value, data, gas.Fast)
	if err != nil {
		log.Fatal("估算 gas 失败:", err)
	}

	fmt.Printf("✓ Gas 估算完成\n")
	fmt.Printf("  Gas Limit: %d\n", params.GasLimit)
	if !params.IsLegacy {
		fmt.Printf("  Max Fee: %s Gwei\n", weiToGwei(params.GasFeeCap))
	}

	// 2. 获取 nonce
	nonce, err := client.PendingNonceAt(ctx, from)
	if err != nil {
		log.Fatal("获取 nonce 失败:", err)
	}
	fmt.Printf("✓ Nonce: %d\n", nonce)

	// 3. 获取链 ID
	chainID, err := client.ChainID(ctx)
	if err != nil {
		log.Fatal("获取链 ID 失败:", err)
	}
	fmt.Printf("✓ Chain ID: %s\n\n", chainID)

	// 4. 创建交易
	tx := gas.CreateTransaction(nonce, &to, value, data, params, chainID)

	// 5. 签名
	signedTx, err := types.SignTx(tx, types.LatestSignerForChainID(chainID), privateKey)
	if err != nil {
		log.Fatal("签名失败:", err)
	}
	fmt.Printf("✓ 交易已签名\n")

	// 6. 发送交易
	err = client.SendTransaction(ctx, signedTx)
	if err != nil {
		log.Fatal("发送交易失败:", err)
	}

	fmt.Printf("\n🚀 交易已发送！\n")
	fmt.Printf("交易哈希: %s\n", signedTx.Hash().Hex())
	fmt.Printf("在区块浏览器查看: https://sepolia.etherscan.io/tx/%s\n", signedTx.Hash().Hex())
}

// 示例3：对比不同速度档位
func exampleCompareSpeed() {
	fmt.Println("=== 对比不同速度档位 ===")
	fmt.Println()

	client, err := ethclient.Dial("https://rpc.ankr.com/eth")
	if err != nil {
		log.Fatal("连接节点失败:", err)
	}
	defer client.Close()

	ctx := context.Background()
	from := common.HexToAddress("0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb")
	to := common.HexToAddress("0x1234567890123456789012345678901234567890")
	value := big.NewInt(1e18) // 1 ETH
	data := []byte{}

	speeds := []gas.Speed{gas.Slow, gas.Normal, gas.Fast}

	for _, speed := range speeds {
		params, err := gas.SuggestGasParams(ctx, client, from, &to, value, data, speed)
		if err != nil {
			log.Printf("速度 %s 估算失败: %v\n", speed, err)
			continue
		}

		fmt.Printf("【%s】\n", map[gas.Speed]string{
			gas.Slow:   "慢速 - 省钱 (1.0x)",
			gas.Normal: "标准 - 推荐 (1.1x)",
			gas.Fast:   "快速 - 秒进块 (1.5x)",
		}[speed])

		fmt.Printf("  Gas Limit: %d\n", params.GasLimit)

		if params.IsLegacy {
			gasPrice := params.GasPrice
			fmt.Printf("  Gas Price: %s Gwei\n", weiToGwei(gasPrice))
			totalFee := new(big.Int).Mul(gasPrice, big.NewInt(int64(params.GasLimit)))
			fmt.Printf("  预估费用: %s ETH\n", weiToEth(totalFee))
		} else {
			fmt.Printf("  Priority Fee: %s Gwei\n", weiToGwei(params.GasTipCap))
			fmt.Printf("  Max Fee: %s Gwei\n", weiToGwei(params.GasFeeCap))
			totalFee := new(big.Int).Mul(params.GasFeeCap, big.NewInt(int64(params.GasLimit)))
			fmt.Printf("  预估最高费用: %s ETH\n", weiToEth(totalFee))
		}
		fmt.Println()
	}
}

// 工具函数：Wei 转 Gwei
func weiToGwei(wei *big.Int) string {
	gwei := new(big.Float).Quo(
		new(big.Float).SetInt(wei),
		big.NewFloat(1e9),
	)
	return gwei.Text('f', 2)
}

// 工具函数：Wei 转 ETH
func weiToEth(wei *big.Int) string {
	eth := new(big.Float).Quo(
		new(big.Float).SetInt(wei),
		big.NewFloat(1e18),
	)
	return eth.Text('f', 6)
}
