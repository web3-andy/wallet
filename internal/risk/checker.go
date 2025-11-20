package risk

import (
	"fmt"
	"math/big"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
)

// Checker 风控检查器
type Checker struct {
	config        *Config
	blacklist     map[common.Address]bool
	whitelist     map[common.Address]bool
	dailyAmounts  map[common.Address]*DailyAmount // 每日累计金额
	mu            sync.RWMutex
}

// Config 风控配置
type Config struct {
	Enabled           bool
	DailyLimit        *big.Int // Wei
	SingleLimit       *big.Int // Wei
	WhitelistAddrs    []string
	BlacklistAddrs    []string
	RequireManualApproval bool
}

// DailyAmount 每日金额统计
type DailyAmount struct {
	Date   string   // YYYY-MM-DD
	Amount *big.Int // 累计金额
}

// CheckResult 检查结果
type CheckResult struct {
	Passed bool
	Reason string
	Risk   RiskLevel
}

// RiskLevel 风险等级
type RiskLevel int

const (
	RiskNone   RiskLevel = 0 // 无风险
	RiskLow    RiskLevel = 1 // 低风险
	RiskMedium RiskLevel = 2 // 中风险
	RiskHigh   RiskLevel = 3 // 高风险
)

// New 创建风控检查器
func New(config *Config) *Checker {
	c := &Checker{
		config:       config,
		blacklist:    make(map[common.Address]bool),
		whitelist:    make(map[common.Address]bool),
		dailyAmounts: make(map[common.Address]*DailyAmount),
	}

	// 加载黑白名单
	for _, addr := range config.BlacklistAddrs {
		c.blacklist[common.HexToAddress(addr)] = true
	}
	for _, addr := range config.WhitelistAddrs {
		c.whitelist[common.HexToAddress(addr)] = true
	}

	return c
}

// Check 执行风控检查
func (c *Checker) Check(from, to common.Address, amount *big.Int) *CheckResult {
	if !c.config.Enabled {
		return &CheckResult{Passed: true, Risk: RiskNone}
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	// 1. 白名单检查（白名单地址直接通过）
	if c.whitelist[from] || c.whitelist[to] {
		return &CheckResult{Passed: true, Risk: RiskNone}
	}

	// 2. 黑名单检查
	if c.blacklist[from] {
		return &CheckResult{
			Passed: false,
			Reason: fmt.Sprintf("发送地址 %s 在黑名单中", from.Hex()),
			Risk:   RiskHigh,
		}
	}
	if c.blacklist[to] {
		return &CheckResult{
			Passed: false,
			Reason: fmt.Sprintf("接收地址 %s 在黑名单中", to.Hex()),
			Risk:   RiskHigh,
		}
	}

	// 3. 单笔限额检查
	if c.config.SingleLimit != nil && amount.Cmp(c.config.SingleLimit) > 0 {
		return &CheckResult{
			Passed: false,
			Reason: fmt.Sprintf("超过单笔限额: %s > %s",
				weiToEth(amount), weiToEth(c.config.SingleLimit)),
			Risk: RiskMedium,
		}
	}

	// 4. 每日限额检查
	if c.config.DailyLimit != nil {
		today := time.Now().Format("2006-01-02")
		dailyAmount, exists := c.dailyAmounts[from]

		if !exists || dailyAmount.Date != today {
			// 新的一天，重置
			c.dailyAmounts[from] = &DailyAmount{
				Date:   today,
				Amount: big.NewInt(0),
			}
			dailyAmount = c.dailyAmounts[from]
		}

		// 计算累计金额
		newTotal := new(big.Int).Add(dailyAmount.Amount, amount)
		if newTotal.Cmp(c.config.DailyLimit) > 0 {
			return &CheckResult{
				Passed: false,
				Reason: fmt.Sprintf("超过每日限额: %s > %s",
					weiToEth(newTotal), weiToEth(c.config.DailyLimit)),
				Risk: RiskMedium,
			}
		}

		// 更新累计金额
		dailyAmount.Amount = newTotal
	}

	// 5. 大额需人工审批
	if c.config.RequireManualApproval {
		threshold := new(big.Int).Div(c.config.SingleLimit, big.NewInt(2)) // 单笔限额的 50%
		if amount.Cmp(threshold) > 0 {
			return &CheckResult{
				Passed: false,
				Reason: "大额交易需要人工审批",
				Risk:   RiskLow,
			}
		}
	}

	return &CheckResult{Passed: true, Risk: RiskNone}
}

// AddToBlacklist 添加到黑名单
func (c *Checker) AddToBlacklist(addr string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.blacklist[common.HexToAddress(addr)] = true
}

// RemoveFromBlacklist 从黑名单移除
func (c *Checker) RemoveFromBlacklist(addr string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.blacklist, common.HexToAddress(addr))
}

// AddToWhitelist 添加到白名单
func (c *Checker) AddToWhitelist(addr string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.whitelist[common.HexToAddress(addr)] = true
}

// RemoveFromWhitelist 从白名单移除
func (c *Checker) RemoveFromWhitelist(addr string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.whitelist, common.HexToAddress(addr))
}

// GetDailyAmount 获取每日累计金额
func (c *Checker) GetDailyAmount(addr common.Address) *big.Int {
	c.mu.RLock()
	defer c.mu.RUnlock()

	today := time.Now().Format("2006-01-02")
	if dailyAmount, exists := c.dailyAmounts[addr]; exists && dailyAmount.Date == today {
		return new(big.Int).Set(dailyAmount.Amount)
	}
	return big.NewInt(0)
}

func weiToEth(wei *big.Int) string {
	eth := new(big.Float).Quo(
		new(big.Float).SetInt(wei),
		big.NewFloat(1e18),
	)
	return eth.Text('f', 6)
}
