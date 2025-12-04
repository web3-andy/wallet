# BIP39 & BIP44 详解

## 目录

- [什么是 BIP？](#什么是-bip)
- [BIP39 - 助记词标准](#bip39---助记词标准)
- [BIP44 - 分层确定性钱包](#bip44---分层确定性钱包)
- [实际应用场景](#实际应用场景)
- [安全最佳实践](#安全最佳实践)
- [代码示例](#代码示例)

---

## 什么是 BIP？

**BIP** = **Bitcoin Improvement Proposal**（比特币改进提案）

虽然名字叫"比特币改进提案"，但这些标准已经被整个区块链行业广泛采用，包括以太坊、BSC、Polygon 等所有主流区块链。

---

## BIP39 - 助记词标准

### 核心概念

BIP39 定义了如何将随机数（熵）转换为人类可读的单词序列（助记词），以及如何从助记词生成种子。

### 工作流程

```
随机熵 → 校验和 → 助记词 → 种子
  ↓         ↓         ↓        ↓
128位   添加4位   12个单词   512位
256位   添加8位   24个单词   512位
```

### 详细步骤

#### 1. 生成随机熵（Entropy）

```
熵大小          助记词长度
128 bits    →   12 words
160 bits    →   15 words
192 bits    →   18 words
224 bits    →   21 words
256 bits    →   24 words  (最安全)
```

**示例：**
```
熵: 13d1e6a11413e2262289a8f5528862a8
```

#### 2. 计算校验和（Checksum）

校验和 = SHA256(熵) 的前 N 位
- 128 位熵 → 4 位校验和
- 256 位熵 → 8 位校验和

**作用：** 检测助记词输入错误

#### 3. 转换为助记词

- 将"熵 + 校验和"按每 11 位分组
- 每组对应一个单词（2048 个单词列表中的一个）
- 12 个单词 = 132 位 = 128 位熵 + 4 位校验和

**示例：**
```
because monkey portion choice dilemma basic
mechanic crush vocal nephew board expand
```

#### 4. 生成种子（Seed）

使用 PBKDF2 算法从助记词生成 512 位种子：

```
种子 = PBKDF2(
    password = 助记词,
    salt = "mnemonic" + 可选密码,
    iterations = 2048,
    hash = HMAC-SHA512
)
```

**可选密码（Passphrase）的作用：**
- 相同助记词 + 不同密码 = 完全不同的钱包
- 可以创建"诱饵钱包"（输入假密码显示小额资金）
- 增加一层安全保护

### BIP39 的优势

✅ **易于备份**：只需记住 12 个单词
✅ **跨平台兼容**：所有钱包都支持
✅ **包含校验和**：能检测输入错误
✅ **支持多语言**：中文、日文、韩文等

---

## BIP44 - 分层确定性钱包

### 核心概念

BIP44 定义了如何从一个种子派生出无限多个地址，并且这些地址按照标准路径组织。

### HD 钱包层级结构

```
m / purpose' / coin_type' / account' / change / address_index
│      │           │          │         │            │
│      │           │          │         │            └─ 地址索引 (0,1,2,...)
│      │           │          │         └────────────── 0=外部(接收) 1=内部(找零)
│      │           │          └──────────────────────── 账户索引 (0,1,2,...)
│      │           └─────────────────────────────────── 币种类型
│      └─────────────────────────────────────────────── 44 表示 BIP44 标准
└────────────────────────────────────────────────────── m 表示主密钥
```

### 路径详解

#### Purpose（目的）
- `44'` - BIP44 标准
- `49'` - BIP49（隔离见证）
- `84'` - BIP84（原生隔离见证）

**注意：** `'` 表示硬化派生（Hardened Derivation），增强安全性

#### Coin Type（币种类型）

| 币种 | Coin Type | 示例路径 |
|------|-----------|----------|
| Bitcoin | 0' | m/44'/0'/0'/0/0 |
| Ethereum | 60' | m/44'/60'/0'/0/0 |
| BSC | 60' | m/44'/60'/0'/0/0 |
| Polygon | 60' | m/44'/60'/0'/0/0 |
| Litecoin | 2' | m/44'/2'/0'/0/0 |
| Dogecoin | 3' | m/44'/3'/0'/0/0 |

**说明：** EVM 兼容链都使用 `60'`（以太坊的币种编号）

#### Account（账户）

允许用户在同一个钱包中创建多个独立账户：

```
m/44'/60'/0'/0/0   ← 账户 0 的第一个地址
m/44'/60'/1'/0/0   ← 账户 1 的第一个地址
m/44'/60'/2'/0/0   ← 账户 2 的第一个地址
```

**用途：**
- 个人账户 vs 企业账户
- 不同用途分离（投资 / 日常 / 储蓄）

#### Change（找零）

- `0` - 外部链（External）：用于接收资金的地址
- `1` - 内部链（Internal）：用于找零的地址

**比特币场景：**
```
你有 1 BTC，要转 0.3 BTC 给朋友
→ 0.3 BTC 发送到朋友地址（外部地址）
→ 0.7 BTC 找零发送到你的找零地址（内部地址）
```

**以太坊场景：**
- 以太坊没有 UTXO 模型，通常只用 `0`（外部链）
- 大部分以太坊钱包不使用 `1`（找零链）

#### Address Index（地址索引）

从 0 开始的连续编号：

```
m/44'/60'/0'/0/0   ← 第 1 个地址
m/44'/60'/0'/0/1   ← 第 2 个地址
m/44'/60'/0'/0/2   ← 第 3 个地址
...
m/44'/60'/0'/0/99  ← 第 100 个地址
```

### 常见路径示例

#### 以太坊钱包

**MetaMask 默认路径：**
```
m/44'/60'/0'/0/0   ← 账户 1
m/44'/60'/0'/0/1   ← 账户 2
m/44'/60'/0'/0/2   ← 账户 3
```

**Ledger 硬件钱包：**
```
m/44'/60'/0'/0     ← Live 路径（旧版）
m/44'/60'/x'/0/0   ← Legacy 路径（x 是账户索引）
```

#### 比特币钱包

**Legacy 地址：**
```
m/44'/0'/0'/0/0
```

**隔离见证地址：**
```
m/49'/0'/0'/0/0    ← P2SH-P2WPKH
m/84'/0'/0'/0/0    ← Native SegWit (bech32)
```

---

## 实际应用场景

### 场景 1: 一个助记词管理所有资产

```
助记词: abandon abandon abandon ... art
  │
  ├─ m/44'/0'/0'/0/0   → 比特币地址 1
  ├─ m/44'/0'/0'/0/1   → 比特币地址 2
  ├─ m/44'/60'/0'/0/0  → 以太坊地址 1
  ├─ m/44'/60'/0'/0/1  → 以太坊地址 2
  └─ m/44'/2'/0'/0/0   → 莱特币地址 1
```

### 场景 2: 冷热钱包分离

```
账户 0: 热钱包（日常使用）
m/44'/60'/0'/0/0

账户 1: 冷钱包（长期存储）
m/44'/60'/1'/0/0
```

### 场景 3: 企业多部门管理

```
账户 0: 财务部
账户 1: 市场部
账户 2: 研发部
账户 3: 法务部
```

### 场景 4: 交易所批量地址生成

```go
// 为 10000 个用户生成充值地址
for i := 0; i < 10000; i++ {
    path := fmt.Sprintf("m/44'/60'/0'/0/%d", i)
    address := deriveAddress(seed, path)
    saveToDatabase(userID, address, path)
}
```

---

## 硬化派生 vs 普通派生

### 硬化派生（Hardened Derivation）

**标记：** 路径中带 `'` 或 `h`，如 `44'` 或 `44h`

**计算方式：**
```
子私钥 = HMAC-SHA512(父私钥, 索引 + 2^31)
```

**特点：**
- ✅ 更安全：即使子私钥泄露，也无法推导出父私钥
- ❌ 需要私钥才能派生
- 适用于：purpose, coin_type, account 层级

### 普通派生（Normal Derivation）

**标记：** 路径中不带 `'`，如 `0`

**计算方式：**
```
子公钥 = HMAC-SHA512(父公钥, 索引)
```

**特点：**
- ✅ 可以只用公钥派生子公钥（无需私钥）
- ❌ 如果子私钥泄露 + 父公钥泄露 = 可以推导父私钥
- 适用于：change, address_index 层级

### 安全建议

```
✅ 推荐（更安全）:
m/44'/60'/0'/0'/0'   ← 全部硬化

⚠️  标准（平衡）:
m/44'/60'/0'/0/0     ← BIP44 标准

❌ 不推荐（风险）:
m/44/60/0/0/0        ← 全部普通派生（不安全）
```

---

## 安全最佳实践

### 助记词安全

#### ✅ 应该做

1. **物理备份**
   - 写在纸上，存放在多个安全地方
   - 使用金属板刻录（防火防水）
   - 考虑使用 Shamir 秘密共享（分片备份）

2. **离线存储**
   - 永远不要在联网设备上输入助记词
   - 不要拍照、截图、存储在云端
   - 使用硬件钱包（Ledger, Trezor）

3. **验证备份**
   - 创建钱包后立即验证备份
   - 定期检查备份是否完整可读

#### ❌ 禁止做

1. **不要分享**
   - 永远不要告诉任何人你的助记词
   - 没有客服会要求你提供助记词
   - 永远不要在网页上输入助记词

2. **不要数字化**
   - 不要存储在电脑、手机上
   - 不要通过微信、邮件发送
   - 不要存储在密码管理器中

3. **不要过度依赖**
   - 不要只有一份备份
   - 不要只靠记忆（人会遗忘）

### 密码（Passphrase）使用

```go
// 相同助记词 + 不同密码 = 不同钱包
seed1 := bip39.NewSeed(mnemonic, "")           // 主钱包
seed2 := bip39.NewSeed(mnemonic, "password1")  // 冷钱包
seed3 := bip39.NewSeed(mnemonic, "password2")  // 诱饵钱包（小额）
```

**优点：**
- 增加一层保护（即使助记词泄露）
- 可以创建多个隐藏钱包
- 抗暴力破解（密码可以很长很复杂）

**缺点：**
- 密码丢失 = 钱包永久丢失
- 必须额外备份密码
- 增加使用复杂度

---

## 代码示例

### 完整流程示例

```go
package main

import (
    "fmt"
    "github.com/tyler-smith/go-bip39"
    hdwallet "github.com/miguelmota/go-ethereum-hdwallet"
)

func main() {
    // 1. 生成助记词
    entropy, _ := bip39.NewEntropy(128)
    mnemonic, _ := bip39.NewMnemonic(entropy)
    fmt.Println("助记词:", mnemonic)

    // 2. 从助记词创建 HD 钱包
    wallet, _ := hdwallet.NewFromMnemonic(mnemonic)

    // 3. 派生以太坊地址
    path := hdwallet.MustParseDerivationPath("m/44'/60'/0'/0/0")
    account, _ := wallet.Derive(path, false)

    fmt.Println("以太坊地址:", account.Address.Hex())
}
```

### 批量生成地址

```go
// 为交易所生成 1000 个充值地址
func generateDepositAddresses(mnemonic string, count int) []string {
    wallet, _ := hdwallet.NewFromMnemonic(mnemonic)
    addresses := make([]string, count)

    for i := 0; i < count; i++ {
        path := hdwallet.MustParseDerivationPath(
            fmt.Sprintf("m/44'/60'/0'/0/%d", i),
        )
        account, _ := wallet.Derive(path, false)
        addresses[i] = account.Address.Hex()
    }

    return addresses
}
```

### 恢复钱包

```go
// 从助记词恢复钱包
func recoverWallet(mnemonic string) {
    // 验证助记词
    if !bip39.IsMnemonicValid(mnemonic) {
        fmt.Println("❌ 助记词无效")
        return
    }

    // 恢复钱包
    wallet, _ := hdwallet.NewFromMnemonic(mnemonic)

    // 恢复前 5 个地址
    for i := 0; i < 5; i++ {
        path := hdwallet.MustParseDerivationPath(
            fmt.Sprintf("m/44'/60'/0'/0/%d", i),
        )
        account, _ := wallet.Derive(path, false)
        fmt.Printf("地址 %d: %s\n", i, account.Address.Hex())
    }
}
```

---

## 常见问题

### Q1: 为什么以太坊和 BSC 使用相同的 coin_type (60')?

**答：** BSC 是 EVM 兼容链，为了兼容性使用了以太坊的币种编号。同一个助记词可以生成相同的地址在以太坊和 BSC 上使用。

### Q2: 丢失助记词但还有私钥，能恢复吗？

**答：** 不能。私钥只能控制单个地址，无法反推助记词或派生其他地址。

### Q3: 助记词可以修改吗？

**答：** 不能。助记词是通过随机熵生成的，修改任何一个单词都会变成完全不同的钱包。

### Q4: 12 个单词和 24 个单词哪个更好？

**答：**
- **12 个单词**：128 位熵，2^128 种组合，安全性足够
- **24 个单词**：256 位熵，2^256 种组合，更安全但更难记

对于个人用户，12 个单词已经足够安全。

### Q5: 可以使用自定义的助记词吗？

**答：** 不推荐。自己选择的单词通常熵不足，容易被暴力破解。应该使用加密安全的随机数生成器。

### Q6: 不同钱包的地址为什么不一样？

**答：** 可能是使用了不同的派生路径。例如：
- MetaMask: `m/44'/60'/0'/0/x`
- Ledger: `m/44'/60'/x'/0/0`

---

## 参考资源

### 官方文档
- [BIP39 规范](https://github.com/bitcoin/bips/blob/master/bip-0039.mediawiki)
- [BIP44 规范](https://github.com/bitcoin/bips/blob/master/bip-0044.mediawiki)
- [BIP32 规范](https://github.com/bitcoin/bips/blob/master/bip-0032.mediawiki) (HD 钱包基础)

### 在线工具
- [BIP39 助记词生成器](https://iancoleman.io/bip39/) - 离线使用！
- [SLIP44 币种列表](https://github.com/satoshilabs/slips/blob/master/slip-0044.md)

### Go 语言库
- [go-bip39](https://github.com/tyler-smith/go-bip39) - BIP39 实现
- [go-ethereum-hdwallet](https://github.com/miguelmota/go-ethereum-hdwallet) - 以太坊 HD 钱包

### 推荐阅读
- [精通比特币（第 2 版）- 第 5 章：钱包](https://github.com/bitcoinbook/bitcoinbook)
- [以太坊技术黄皮书](https://ethereum.github.io/yellowpaper/paper.pdf)

---

## 本项目示例

查看完整的 BIP39 和 BIP44 测试代码：

```bash
# 运行测试代码
go run playground/bip39_bip44.go
```

测试代码包含：
- ✅ BIP39 助记词生成和验证
- ✅ BIP44 地址派生
- ✅ 钱包恢复
- ✅ 多账户管理
- ✅ 地址查找

---

**⚠️ 安全提醒：永远不要在生产环境中打印或记录私钥和助记词！**
