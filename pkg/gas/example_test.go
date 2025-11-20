package gas_test

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"wallet/gas"
)

// 示例：发送简单转账交易
func ExampleSuggestGasParams_transfer() {
	// 连接到节点
	client, err := ethclient.Dial("https://rpc.ankr.com/eth")
	if err != nil {
		panic(err)
	}
	defer client.Close()

	ctx := context.Background()

	// 准备交易参数
	privateKey, _ := crypto.HexToECDSA("your-private-key-hex")
	publicKey := privateKey.Public().(*ecdsa.PublicKey)
	from := crypto.PubkeyToAddress(*publicKey)
	to := common.HexToAddress("0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb")
	value := big.NewInt(1e17) // 0.1 ETH
	data := []byte{}

	// 1. 获取 gas 参数建议
	params, err := gas.SuggestGasParams(
		ctx,
		client,
		from,
		&to,
		value,
		data,
		gas.Normal, // 使用推荐速度
	)
	if err != nil {
		panic(err)
	}

	// 2. 获取 nonce
	nonce, err := client.PendingNonceAt(ctx, from)
	if err != nil {
		panic(err)
	}

	// 3. 获取链 ID
	chainID, err := client.ChainID(ctx)
	if err != nil {
		panic(err)
	}

	// 4. 创建交易
	tx := gas.CreateTransaction(nonce, &to, value, data, params, chainID)

	// 5. 签名交易
	signedTx, err := types.SignTx(tx, types.LatestSignerForChainID(chainID), privateKey)
	if err != nil {
		panic(err)
	}

	// 6. 发送交易
	err = client.SendTransaction(ctx, signedTx)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Transaction sent: %s\n", signedTx.Hash().Hex())
}

// 示例：调用智能合约
func ExampleSuggestGasParams_contract() {
	client, _ := ethclient.Dial("https://bsc-dataseed.binance.org/")
	defer client.Close()

	ctx := context.Background()
	privateKey, _ := crypto.HexToECDSA("your-private-key-hex")
	publicKey := privateKey.Public().(*ecdsa.PublicKey)
	from := crypto.PubkeyToAddress(*publicKey)

	// 合约地址和调用数据（例如：ERC20 transfer）
	contractAddress := common.HexToAddress("0x...")
	// transfer(address,uint256) 的函数选择器 + 参数
	data := common.Hex2Bytes("a9059cbb000000000000000000000000742d35cc6634c0532925a3b844bc9e7595f0beb0000000000000000000000000000000000000000000000000de0b6b3a7640000")

	// 获取 gas 参数（BSC 是 legacy 链）
	params, err := gas.SuggestGasParams(
		ctx,
		client,
		from,
		&contractAddress,
		big.NewInt(0), // 合约调用通常 value = 0
		data,
		gas.Fast, // 使用快速模式
	)
	if err != nil {
		panic(err)
	}

	// 检查是否为 legacy 交易
	fmt.Printf("Is Legacy: %v\n", params.IsLegacy) // BSC 应该输出 true
	fmt.Printf("Gas Limit: %d\n", params.GasLimit)
	if params.IsLegacy {
		fmt.Printf("Gas Price: %s gwei\n", new(big.Int).Div(params.GasPrice, big.NewInt(1e9)))
	} else {
		fmt.Printf("Tip Cap: %s gwei\n", new(big.Int).Div(params.GasTipCap, big.NewInt(1e9)))
		fmt.Printf("Fee Cap: %s gwei\n", new(big.Int).Div(params.GasFeeCap, big.NewInt(1e9)))
	}

	// 后续步骤：创建交易 -> 签名 -> 发送
	// ...
}

// 示例：对比不同速度档位
func ExampleSuggestGasParams_speeds() {
	client, _ := ethclient.Dial("https://rpc.ankr.com/eth")
	defer client.Close()

	ctx := context.Background()
	from := common.HexToAddress("0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb")
	to := common.HexToAddress("0x...")
	value := big.NewInt(1e18)
	data := []byte{}

	speeds := []gas.Speed{gas.Slow, gas.Normal, gas.Fast}

	for _, speed := range speeds {
		params, err := gas.SuggestGasParams(ctx, client, from, &to, value, data, speed)
		if err != nil {
			continue
		}

		fmt.Printf("\n[%s]\n", speed)
		fmt.Printf("  Gas Limit: %d\n", params.GasLimit)
		if !params.IsLegacy {
			fmt.Printf("  Tip: %s gwei\n", new(big.Int).Div(params.GasTipCap, big.NewInt(1e9)))
			fmt.Printf("  Max Fee: %s gwei\n", new(big.Int).Div(params.GasFeeCap, big.NewInt(1e9)))
		}
	}
}
