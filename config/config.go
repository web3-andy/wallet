package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Config 系统配置
type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Database DatabaseConfig `yaml:"database"`
	Chains   []ChainConfig  `yaml:"chains"`
	Scanner  ScannerConfig  `yaml:"scanner"`
	Collect  CollectConfig  `yaml:"collect"`
	Risk     RiskConfig     `yaml:"risk"`
}

// ServerConfig API 服务器配置
type ServerConfig struct {
	Host         string        `yaml:"host"`
	Port         int           `yaml:"port"`
	ReadTimeout  time.Duration `yaml:"read_timeout"`
	WriteTimeout time.Duration `yaml:"write_timeout"`
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Driver   string `yaml:"driver"` // postgres, mysql, sqlite
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	Database string `yaml:"database"`
	MaxConns int    `yaml:"max_conns"`
}

// ChainConfig 区块链配置
type ChainConfig struct {
	ChainID   int64    `yaml:"chain_id"`
	Name      string   `yaml:"name"` // eth, bsc, polygon
	RPCURLs   []string `yaml:"rpc_urls"`
	WSURLs    []string `yaml:"ws_urls"`
	IsTestnet bool     `yaml:"is_testnet"`
}

// ScannerConfig 扫块配置
type ScannerConfig struct {
	Enabled          bool          `yaml:"enabled"`
	StartBlock       uint64        `yaml:"start_block"`        // 起始区块（0 表示最新）
	ConfirmBlocks    uint64        `yaml:"confirm_blocks"`     // 确认区块数
	BatchSize        int           `yaml:"batch_size"`         // 批量扫描大小
	ScanInterval     time.Duration `yaml:"scan_interval"`      // 扫描间隔
	ConcurrentChains int           `yaml:"concurrent_chains"`  // 并发扫描链数
}

// CollectConfig 归集配置
type CollectConfig struct {
	Enabled        bool          `yaml:"enabled"`
	Interval       time.Duration `yaml:"interval"`        // 归集间隔
	MinAmount      string        `yaml:"min_amount"`      // 最小归集金额（ETH）
	TargetAddress  string        `yaml:"target_address"`  // 归集目标地址
	ReserveAmount  string        `yaml:"reserve_amount"`  // 保留 gas 费金额
	MaxConcurrent  int           `yaml:"max_concurrent"`  // 最大并发归集数
}

// RiskConfig 风控配置
type RiskConfig struct {
	Enabled           bool     `yaml:"enabled"`
	DailyLimit        string   `yaml:"daily_limit"`        // 单日限额
	SingleLimit       string   `yaml:"single_limit"`       // 单笔限额
	WhitelistAddrs    []string `yaml:"whitelist_addrs"`    // 白名单地址
	BlacklistAddrs    []string `yaml:"blacklist_addrs"`    // 黑名单地址
	RequireManualApproval bool `yaml:"require_manual_approval"` // 大额需人工审批
}

// Load 加载配置文件
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	return &cfg, nil
}

// LoadWithEnv 加载配置并覆盖环境变量
func LoadWithEnv(path string) (*Config, error) {
	cfg, err := Load(path)
	if err != nil {
		return nil, err
	}

	// 从环境变量覆盖敏感配置
	if dbPassword := os.Getenv("DB_PASSWORD"); dbPassword != "" {
		cfg.Database.Password = dbPassword
	}

	return cfg, nil
}
