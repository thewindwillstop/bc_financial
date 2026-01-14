package blockchain

import (
	"context"
	"encoding/hex"
	"fmt"
	"math/big"
	"strconv"

	"bc-reconciliation-backend/internal/config"

	"github.com/FISCO-BCOS/go-sdk/client"
	"github.com/FISCO-BCOS/go-sdk/conf"
	"github.com/FISCO-BCOS/go-sdk/core/types"
	"github.com/ethereum/go-ethereum/common"
	"go.uber.org/zap"
)

// Client FISCO BCOS客户端封装
type Client struct {
	client         *client.Client
	config         *config.BlockchainConfig
	logger         *zap.Logger
	contractHelper *ContractHelper
	contractAddr   common.Address
}

// NewClient 创建区块链客户端
func NewClient(cfg *config.BlockchainConfig, logger *zap.Logger) (*Client, error) {
	// 加载FISCO配置
	configs, err := conf.ParseConfigFile(cfg.ConfigFile)
	if err != nil {
		return nil, fmt.Errorf("failed to parse FISCO config: %w", err)
	}

	// 连接节点
	c, err := client.Dial(&configs[0])
	if err != nil {
		return nil, fmt.Errorf("failed to connect to FISCO BCOS: %w", err)
	}

	// 测试连接
	ctx := context.Background()
	blockNumber, err := c.GetBlockNumber(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get block number: %w", err)
	}

	logger.Info("connected to FISCO BCOS",
		zap.Int64("block_number", blockNumber),
		zap.String("config_file", cfg.ConfigFile))

	blockchainClient := &Client{
		client:       c,
		config:       cfg,
		logger:       logger,
		contractAddr: common.HexToAddress(cfg.ContractAddress),
	}

	// 加载智能合约ABI
	if cfg.ContractAddress != "" && cfg.ContractAddress != `""` {
		// 使用内嵌的ABI
		abiContent := getEmbeddedABI()

		helper, err := NewContractHelper(abiContent, cfg.ContractAddress)
		if err != nil {
			logger.Warn("failed to create contract helper",
				zap.Error(err),
				zap.String("contract_address", cfg.ContractAddress))
		} else {
			blockchainClient.contractHelper = helper
			logger.Info("smart contract loaded successfully",
				zap.String("contract_address", cfg.ContractAddress))
		}
	}

	return blockchainClient, nil
}

// GetClient 获取原始客户端
func (c *Client) GetClient() *client.Client {
	return c.client
}

// GetContractAddress 获取合约地址
func (c *Client) GetContractAddress() common.Address {
	return c.contractAddr
}

// GetBlockNumber 获取当前区块高度
func (c *Client) GetBlockNumber(ctx context.Context) (int64, error) {
	return c.client.GetBlockNumber(ctx)
}

// GetSystemConfig 获取系统配置
func (c *Client) GetSystemConfig(ctx context.Context, key string) (string, error) {
	configBytes, err := c.client.GetSystemConfigByKey(ctx, key)
	if err != nil {
		return "", err
	}
	return string(configBytes), nil
}

// ========== 合约调用方法 ==========

// UploadTransaction 上传交易到区块链
func (c *Client) UploadTransaction(ctx context.Context, bizId, dataHash string) (string, *types.Receipt, error) {
	if c.contractHelper == nil {
		return "", nil, fmt.Errorf("contract helper not initialized")
	}

	// 编码合约调用数据
	input, err := c.contractHelper.EncodeUploadTransaction(bizId, dataHash)
	if err != nil {
		return "", nil, fmt.Errorf("failed to encode uploadTransaction: %w", err)
	}

	// 发送交易
	receipt, err := c.sendTransaction(ctx, input)
	if err != nil {
		return "", nil, fmt.Errorf("failed to send transaction: %w", err)
	}

	c.logger.Info("transaction uploaded to blockchain",
		zap.String("tx_hash", receipt.TransactionHash),
		zap.String("biz_id", bizId),
		zap.String("block_number", receipt.BlockNumber),
		zap.String("gas_used", receipt.GasUsed))

	return receipt.TransactionHash, receipt, nil
}

// BatchUploadTransactions 批量上传交易
func (c *Client) BatchUploadTransactions(ctx context.Context, bizIds, dataHashes []string) (string, *types.Receipt, error) {
	if c.contractHelper == nil {
		return "", nil, fmt.Errorf("contract helper not initialized")
	}

	if len(bizIds) != len(dataHashes) {
		return "", nil, fmt.Errorf("bizIds and dataHashes length mismatch")
	}

	// 编码合约调用数据
	input, err := c.contractHelper.EncodeBatchUploadTransactions(bizIds, dataHashes)
	if err != nil {
		return "", nil, fmt.Errorf("failed to encode batchUploadTransactions: %w", err)
	}

	// 发送交易
	receipt, err := c.sendTransaction(ctx, input)
	if err != nil {
		return "", nil, fmt.Errorf("failed to send batch transaction: %w", err)
	}

	c.logger.Info("batch uploaded transactions to blockchain",
		zap.String("tx_hash", receipt.TransactionHash),
		zap.Int("count", len(bizIds)),
		zap.String("block_number", receipt.BlockNumber))

	return receipt.TransactionHash, receipt, nil
}

// GetTransaction 查询交易信息
func (c *Client) GetTransaction(ctx context.Context, bizId string) (*TransactionInfo, error) {
	if c.contractHelper == nil {
		return nil, fmt.Errorf("contract helper not initialized")
	}

	// 编码调用数据
	bizIdBytes32 := BizIdToBytes32(bizId)
	input, err := c.contractHelper.abi.Pack("getTransaction", bizIdBytes32)
	if err != nil {
		return nil, fmt.Errorf("failed to pack getTransaction: %w", err)
	}

	// 调用合约 - 使用 FISCO SDK 的 CallContract
	// 注意: FISCO SDK 可能使用不同的CallMsg结构
	// 这里简化处理,直接调用
	result, err := c.client.CallContract(ctx, c.contractAddr, input)
	if err != nil {
		return nil, fmt.Errorf("failed to call getTransaction: %w", err)
	}

	// 解码返回值
	txResult, err := c.contractHelper.DecodeGetTransaction(result)
	if err != nil {
		return nil, fmt.Errorf("failed to decode getTransaction result: %w", err)
	}

	return &TransactionInfo{
		DataHash:     txResult.DataHash,
		Uploader:     txResult.Uploader,
		Timestamp:    txResult.Timestamp,
		Status:       txResult.Status,
		Counterparty: txResult.Counterparty,
		MatchHeight:  txResult.MatchHeight,
	}, nil
}

// GetStatistics 获取统计信息
func (c *Client) GetStatistics(ctx context.Context) (*StatisticsInfo, error) {
	if c.contractHelper == nil {
		return nil, fmt.Errorf("contract helper not initialized")
	}

	// 编码调用数据
	input, err := c.contractHelper.abi.Pack("getStatistics")
	if err != nil {
		return nil, fmt.Errorf("failed to pack getStatistics: %w", err)
	}

	// 调用合约
	result, err := c.client.CallContract(ctx, c.contractAddr, input)
	if err != nil {
		return nil, fmt.Errorf("failed to call getStatistics: %w", err)
	}

	// 解码返回值
	statsResult, err := c.contractHelper.DecodeGetStatistics(result)
	if err != nil {
		return nil, fmt.Errorf("failed to decode getStatistics result: %w", err)
	}

	return &StatisticsInfo{
		TotalTx:          statsResult.TotalTx,
		TotalMatched:     statsResult.TotalMatched,
		MatchRate:        statsResult.MatchRate,
		InstitutionCount: statsResult.InstitutionCount,
	}, nil
}

// RegisterInstitution 注册机构
func (c *Client) RegisterInstitution(ctx context.Context, name, address string) (string, *types.Receipt, error) {
	if c.contractHelper == nil {
		return "", nil, fmt.Errorf("contract helper not initialized")
	}

	// 编码合约调用数据
	input, err := c.contractHelper.EncodeRegisterInstitution(name, address)
	if err != nil {
		return "", nil, fmt.Errorf("failed to encode registerInstitution: %w", err)
	}

	// 发送交易
	receipt, err := c.sendTransaction(ctx, input)
	if err != nil {
		return "", nil, fmt.Errorf("failed to send registerInstitution transaction: %w", err)
	}

	c.logger.Info("institution registered",
		zap.String("tx_hash", receipt.TransactionHash),
		zap.String("name", name),
		zap.String("address", address))

	return receipt.TransactionHash, receipt, nil
}

// sendTransaction 发送交易的内部方法
func (c *Client) sendTransaction(ctx context.Context, input []byte) (*types.Receipt, error) {
	// 获取当前区块号
	blockNumber, err := c.client.GetBlockNumber(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get block number: %w", err)
	}

	// 获取交易选项
	auth := c.client.GetTransactOpts()

	// 设置区块限制 (FISCO BCOS 特有)
	// blockLimit = currentBlock + 500
	blockLimit := big.NewInt(blockNumber + 500)

	// 创建交易
	// FISCO BCOS 交易格式
	tx := types.NewTransaction(
		auth.Nonce,
		c.contractAddr,
		big.NewInt(0),    // 金额为0
		big.NewInt(300000), // Gas limit
		auth.GasPrice,
		big.NewInt(0),    // 附加的 blockLimit 参数位置
		input,
		blockLimit,
		big.NewInt(1), // Chain ID
		[]byte{},      // Extra data
		false,         // Use SM2 crypto
	)

	// 发送交易
	receipt, err := c.client.SendTransaction(ctx, tx)
	if err != nil {
		return nil, fmt.Errorf("failed to send transaction: %w", err)
	}

	return receipt, nil
}

// Close 关闭连接
func (c *Client) Close() {
	// FISCO BCOS Go SDK会自动管理连接
}

// ========== 辅助方法 ==========

// getEmbeddedABI 获取内嵌的ABI
func getEmbeddedABI() string {
	return `[{"constant":true,"inputs":[],"name":"getStatistics","outputs":[{"name":"totalTx","type":"uint256"},{"name":"totalMatched","type":"uint256"},{"name":"matchRate","type":"uint256"},{"name":"institutionCount","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":true,"inputs":[],"name":"txCount","outputs":[{"name":"","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":false,"inputs":[],"name":"unpause","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":true,"inputs":[{"name":"bizId","type":"bytes32"}],"name":"getTransaction","outputs":[{"name":"dataHash","type":"bytes32"},{"name":"uploader","type":"address"},{"name":"timestamp","type":"uint256"},{"name":"status","type":"uint8"},{"name":"counterparty","type":"address"},{"name":"matchHeight","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":true,"inputs":[],"name":"paused","outputs":[{"name":"","type":"bool"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":true,"inputs":[{"name":"","type":"bytes32"}],"name":"transactions","outputs":[{"name":"txHash","type":"bytes32"},{"name":"uploader","type":"address"},{"name":"timestamp","type":"uint256"},{"name":"status","type":"uint8"},{"name":"counterparty","type":"address"},{"name":"matchHeight","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":true,"inputs":[{"name":"addr","type":"address"}],"name":"getInstitution","outputs":[{"name":"name","type":"string"},{"name":"institutionAddr","type":"address"},{"name":"isRegistered","type":"bool"},{"name":"uploadCount","type":"uint256"},{"name":"matchedCount","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":false,"inputs":[{"name":"bizIds","type":"bytes32[]"},{"name":"dataHashes","type":"bytes32[]"}],"name":"batchUploadTransactions","outputs":[{"name":"successCount","type":"uint256"}],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":false,"inputs":[],"name":"pause","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":false,"inputs":[{"name":"bizId","type":"bytes32"},{"name":"dataHash","type":"bytes32"}],"name":"uploadTransaction","outputs":[{"name":"","type":"bool"}],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":true,"inputs":[],"name":"owner","outputs":[{"name":"","type":"address"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":true,"inputs":[{"name":"","type":"address"}],"name":"institutions","outputs":[{"name":"name","type":"string"},{"name":"addr","type":"address"},{"name":"isRegistered","type":"bool"},{"name":"uploadCount","value":"256"},{"name":"matchedCount","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":false,"inputs":[{"name":"name","type":"string"},{"name":"addr","type":"address"}],"name":"registerInstitution","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":true,"inputs":[{"name":"","type":"bytes32"}],"name":"txExists","outputs":[{"name":"","type":"bool"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":true,"inputs":[],"name":"matchedCount","outputs":[{"name":"","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":true,"inputs":[{"name":"","type":"uint256"}],"name":"institutionList","outputs":[{"name":"","type":"address"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":false,"inputs":[{"name":"newOwner","type":"address"}],"name":"transferOwnership","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"inputs":[],"payable":false,"stateMutability":"nonpayable","type":"constructor"},{"anonymous":false,"inputs":[{"indexed":true,"name":"bizId","type":"bytes32"},{"indexed":false,"name":"dataHash","type":"bytes32"},{"indexed":true,"name":"uploader","type":"address"},{"indexed":false,"name":"timestamp","type":"uint256"}],"name":"DataUploaded","type":"event"},{"anonymous":false,"inputs":[{"indexed":true,"name":"bizId","type":"bytes32"},{"indexed":false,"name":"status","type":"uint8"},{"indexed":true,"name":"uploader","type":"address"},{"indexed":true,"name":"counterparty","type":"address"},{"indexed":false,"name":"blockHeight","type":"uint256"}],"name":"ReconciliationEvent","type":"event"},{"anonymous":false,"inputs":[{"indexed":true,"name":"institutionAddr","type":"address"},{"indexed":false,"name":"name","type":"string"},{"indexed":false,"name":"timestamp","type":"uint256"}],"name":"InstitutionRegistered","type":"event"}]`
}

// ========== 数据结构 ==========

// TransactionInfo 交易信息
type TransactionInfo struct {
	DataHash     [32]byte
	Uploader     string
	Timestamp    *big.Int
	Status       uint8
	Counterparty string
	MatchHeight  *big.Int
}

// StatisticsInfo 统计信息
type StatisticsInfo struct {
	TotalTx          *big.Int
	TotalMatched     *big.Int
	MatchRate        *big.Int
	InstitutionCount *big.Int
}

// ContractResponse 合约调用响应
type ContractResponse struct {
	TxHash      string `json:"tx_hash"`
	BlockNumber string `json:"block_number"`
	GasUsed     string `json:"gas_used"`
	Status      string `json:"status"`
}

// NewContractResponse 创建合约响应
func NewContractResponse(receipt *types.Receipt) *ContractResponse {
	return &ContractResponse{
		TxHash:      receipt.TransactionHash,
		BlockNumber: receipt.BlockNumber,
		GasUsed:     receipt.GasUsed,
		Status:      "1", // FISCO BCOS 的 status 通常是字符串 "1" 表示成功
	}
}

// GetStatusMessage 获取状态消息
func (r *ContractResponse) GetStatusMessage() string {
	if r.Status == "1" {
		return "Success"
	}
	return "Failed"
}

// FormatTxHash 格式化交易哈希
func FormatTxHash(hash string) string {
	if len(hash) >= 2 && hash[0:2] == "0x" {
		return hash
	}
	return "0x" + hash
}

// ParseHexToBytes32 解析hex字符串为bytes32
func ParseHexToBytes32(hexStr string) ([32]byte, error) {
	if len(hexStr) >= 2 && hexStr[0:2] == "0x" {
		hexStr = hexStr[2:]
	}

	if len(hexStr) != 64 {
		return [32]byte{}, fmt.Errorf("invalid hex length: %d, expected 64", len(hexStr))
	}

	bytes, err := hex.DecodeString(hexStr)
	if err != nil {
		return [32]byte{}, fmt.Errorf("failed to decode hex: %w", err)
	}

	var result [32]byte
	copy(result[:], bytes)
	return result, nil
}

// ParseBlockNumber 解析区块号
func ParseBlockNumber(blockStr string) (uint64, error) {
	return strconv.ParseUint(blockStr, 10, 64)
}

// ParseGasUsed 解析Gas使用量
func ParseGasUsed(gasStr string) (uint64, error) {
	return strconv.ParseUint(gasStr, 10, 64)
}
