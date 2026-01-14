package config

import (
	"time"

	"github.com/spf13/viper"
)

// Config 应用配置
type Config struct {
	Server    ServerConfig    `mapstructure:"server"`
	Database  DatabaseConfig   `mapstructure:"database"`
	Blockchain BlockchainConfig `mapstructure:"blockchain"`
	Fabric    *FabricConfig    `mapstructure:"fabric"` // Fabric配置(可选)
	Log       LogConfig        `mapstructure:"log"`
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Port int    `mapstructure:"port"`
	Mode string `mapstructure:"mode"`
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	MySQL MySQLConfig `mapstructure:"mysql"`
}

// MySQLConfig MySQL配置
type MySQLConfig struct {
	Host         string `mapstructure:"host"`
	Port         int    `mapstructure:"port"`
	Username     string `mapstructure:"username"`
	Password     string `mapstructure:"password"`
	Database     string `mapstructure:"database"`
	MaxOpenConns int    `mapstructure:"max_open_conns"`
	MaxIdleConns int    `mapstructure:"max_idle_conns"`
	MaxLifetime  time.Duration `mapstructure:"max_lifetime"`
}

// BlockchainConfig 区块链配置
type BlockchainConfig struct {
	Type            string `mapstructure:"type"`             // blockchain type: "fisco" or "fabric"
	ConfigFile      string `mapstructure:"config_file"`      // for FISCO: config.toml
	ContractAddress string `mapstructure:"contract_address"` // for FISCO
	NetworkURL      string `mapstructure:"network_url"`      // for Ethereum-style
	ChainID         int64  `mapstructure:"chain_id"`         // for Ethereum-style
}

// FabricConfig Fabric专属配置
type FabricConfig struct {
	ConfigFile  string `mapstructure:"config_file"`  // SDK配置文件路径
	ChannelID   string `mapstructure:"channel_id"`   // Channel ID
	ChaincodeID string `mapstructure:"chaincode_id"` // Chaincode ID
	OrgName     string `mapstructure:"org_name"`     // 组织名称
	User        string `mapstructure:"user"`         // 用户名
}

// LogConfig 日志配置
type LogConfig struct {
	Level      string `mapstructure:"level"`
	Filename   string `mapstructure:"filename"`
	MaxSize    int    `mapstructure:"max_size"`
	MaxBackups int    `mapstructure:"max_backups"`
	MaxAge     int    `mapstructure:"max_age"`
	Compress   bool   `mapstructure:"compress"`
}

// LoadConfig 加载配置文件
func LoadConfig(path string) (*Config, error) {
	viper.SetConfigFile(path)
	viper.SetConfigType("yaml")

	// 读取配置文件
	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, err
	}

	return &config, nil
}

// GetBlockchainType 获取区块链类型
func (c *Config) GetBlockchainType() string {
	if c.Blockchain.Type != "" {
		return c.Blockchain.Type
	}
	// 默认返回 FISCO
	return "fisco"
}

// UseFabric 是否使用Fabric
func (c *Config) UseFabric() bool {
	return c.GetBlockchainType() == "fabric"
}
