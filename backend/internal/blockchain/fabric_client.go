package blockchain

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"bc-reconciliation-backend/internal/config"

	"github.com/hyperledger/fabric-sdk-go/pkg/channel"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
	"go.uber.org/zap"
)

// FabricClient Hyperledger Fabric 客户端
type FabricClient struct {
	sdk         *fabsdk.FabricSDK
	channel     *channel.Client
	channelName string
	cfg          *FabricClientConfig
	logger      *zap.Logger
}

// FabricClientConfig Fabric 客户端配置
type FabricClientConfig struct {
	ConfigFile   string `yaml:"config_file"`   // SDK配置文件路径
	ChannelID    string `yaml:"channel_id"`    // Channel ID
	ChaincodeID  string `yaml:"chaincode_id"`  // Chaincode ID
	OrgName      string `yaml:"org_name"`      // 组织名称
	User         string `yaml:"user"`          // 用户名
}

// NewFabricClient 创建 Fabric 客户端
func NewFabricClient(cfg *FabricClientConfig, logger *zap.Logger) (*FabricClient, error) {
	// 读取配置文件
	sdkConfig, err := config.FromFile(cfg.ConfigFile, "")
	if err != nil {
		return nil, fmt.Errorf("failed to load SDK config: %w", err)
	}

	// 创建 Fabric SDK
	sdk := fabsdk.New(sdkConfig)

	// 创建 channel 客户端
	channelProvider := sdk.ChannelContext(cfg.ChannelID, fabsdk.WithUser(cfg.User), fabsdk.WithOrg(cfg.OrgName))
	client := channel.New(sdk.ChannelContext(channelProvider))

	logger.Info("connected to Hyperledger Fabric",
		zap.String("channel", cfg.ChannelID),
		zap.String("chaincode", cfg.ChaincodeID),
		zap.String("org", cfg.OrgName))

	return &FabricClient{
		sdk:         sdk,
		channel:     client,
		channelName: cfg.ChannelID,
		cfg:         cfg,
		logger:      logger,
	}, nil
}

// ========== 合约调用方法 ==========

// UploadTransaction 上传交易到区块链
func (c *FabricClient) UploadTransaction(ctx context.Context, bizId, dataHash string) (*TxReceipt, error) {
	args := [][]byte{[]byte(bizId), []byte(dataHash)}

	// 执行 Chaincode
	response, err := c.channel.Execute(
		channel.Request{ChaincodeID: c.cfg.ChaincodeID, Fcn: "UploadTransaction", Args: args},
		channel.WithTargetEndpoints(peer0.org1.example.com), // 发送给背书节点
	)
	if err != nil {
		return nil, fmt.Errorf("failed to execute UploadTransaction: %w", err)
	}

	// 解析响应
	var result bool
	err = json.Unmarshal(response.Payload, &result)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	txID := response.TransactionID

	c.logger.Info("transaction uploaded to Fabric",
		zap.String("tx_id", txID),
		zap.String("biz_id", bizId),
		zap.Bool("success", result))

	return &TxReceipt{
		TxID:      txID,
		Status:    "SUCCESS",
		Timestamp: time.Now(),
	}, nil
}

// BatchUploadTransactions 批量上传交易
func (c *FabricClient) BatchUploadTransactions(ctx context.Context, bizIds, dataHashes []string) (*TxReceipt, int, error) {
	if len(bizIds) != len(dataHashes) {
		return nil, 0, fmt.Errorf("bizIds and dataHashes length mismatch")
	}

	// 构造参数
	bizIdBytes, _ := json.Marshal(bizIds)
	dataHashBytes, _ := json.Marshal(dataHashes)

	args := [][]byte{bizIdBytes, dataHashBytes}

	// 执行 Chaincode
	response, err := c.channel.Execute(
		channel.Request{ChaincodeID: c.cfg.ChaincodeID, Fcn: "BatchUploadTransactions", Args: args},
		channel.WithTargetEndpoints(peer0.org1.example.com),
	)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to execute BatchUploadTransactions: %w", err)
	}

	// 解析响应
	var successCount int
	err = json.Unmarshal(response.Payload, &successCount)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	txID := response.TransactionID

	c.logger.Info("batch uploaded transactions to Fabric",
		zap.String("tx_id", txID),
		zap.Int("success_count", successCount),
		zap.Int("total", len(bizIds)))

	return &TxReceipt{
		TxID:      txID,
		Status:    "SUCCESS",
		Timestamp: time.Now(),
	}, successCount, nil
}

// GetTransaction 查询交易信息
func (c *FabricClient) GetTransaction(ctx context.Context, bizId string) (*FabricTransactionInfo, error) {
	args := [][]byte{[]byte(bizId)}

	// 查询 Chaincode (使用 Query,不用 Execute)
	response, err := c.channel.Query(
		channel.Request{ChaincodeID: c.cfg.ChaincodeID, Fcn: "GetTransaction", Args: args},
		channel.WithTargetEndpoints(peer0.org1.example.com),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query GetTransaction: %w", err)
	}

	// 解析响应
	var tx FabricTransaction
	err = json.Unmarshal(response.Payload, &tx)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &FabricTransactionInfo{
		DataHash:     tx.TxHash,
		Uploader:     tx.Uploader,
		Timestamp:    tx.Timestamp,
		Status:       tx.Status,
		Counterparty: tx.Counterparty,
		MatchHeight:  tx.MatchHeight,
	}, nil
}

// GetStatistics 获取统计信息
func (c *FabricClient) GetStatistics(ctx context.Context) (*FabricStatisticsInfo, error) {
	args := [][]byte{}

	response, err := c.channel.Query(
		channel.Request{ChaincodeID: c.cfg.ChaincodeID, Fcn: "GetStatistics", Args: args},
		channel.WithTargetEndpoints(peer0.org1.example.com),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query GetStatistics: %w", err)
	}

	var stats FabricStatistics
	err = json.Unmarshal(response.Payload, &stats)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &FabricStatisticsInfo{
		TotalTx:          stats.TotalTx,
		TotalMatched:     stats.TotalMatched,
		MatchRate:        stats.MatchRate,
		InstitutionCount: stats.InstitutionCount,
	}, nil
}

// RegisterInstitution 注册机构
func (c *FabricClient) RegisterInstitution(ctx context.Context, name, mspid string) (*TxReceipt, error) {
	args := [][]byte{[]byte(name), []byte(mspid)}

	response, err := c.channel.Execute(
		channel.Request{ChaincodeID: c.cfg.ChaincodeID, Fcn: "RegisterInstitution", Args: args},
		channel.WithTargetEndpoints(peer0.org1.example.com),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to execute RegisterInstitution: %w", err)
	}

	txID := response.TransactionID

	c.logger.Info("institution registered on Fabric",
		zap.String("tx_id", txID),
		zap.String("name", name),
		zap.String("mspid", mspid))

	return &TxReceipt{
		TxID:      txID,
		Status:    "SUCCESS",
		Timestamp: time.Now(),
	}, nil
}

// GetCurrentBlockHeight 获取当前区块高度
func (c *FabricClient) GetCurrentBlockHeight(ctx context.Context) (uint64, error) {
	channelProvider := c.sdk.ChannelContext(c.channelName)

	height, err := channelProvider.ChannelService().QueryBlockHeight()
	if err != nil {
		return 0, fmt.Errorf("failed to query block height: %w", err)
	}

	return height, nil
}

// Close 关闭连接
func (c *FabricClient) Close() {
	c.sdk.Close()
}

// ========== 数据结构 ==========

// FabricTransaction Fabric 交易结构
type FabricTransaction struct {
	TxHash       string `json:"txHash"`
	Uploader     string `json:"uploader"`
	Timestamp    int64  `json:"timestamp"`
	Status       int    `json:"status"`
	Counterparty string `json:"counterparty"`
	MatchHeight  int64  `json:"matchHeight"`
}

// FabricStatistics Fabric 统计结构
type FabricStatistics struct {
	TotalTx          int `json:"totalTx"`
	TotalMatched     int `json:"totalMatched"`
	MatchRate        int `json:"matchRate"`
	InstitutionCount int `json:"institutionCount"`
}

// FabricTransactionInfo 交易信息
type FabricTransactionInfo struct {
	DataHash     string
	Uploader     string
	Timestamp    int64
	Status       int
	Counterparty string
	MatchHeight  int64
}

// FabricStatisticsInfo 统计信息
type FabricStatisticsInfo struct {
	TotalTx          int
	TotalMatched     int
	MatchRate        int
	InstitutionCount int
}

// TxReceipt 交易收据
type TxReceipt struct {
	TxID      string    `json:"tx_id"`
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
}
