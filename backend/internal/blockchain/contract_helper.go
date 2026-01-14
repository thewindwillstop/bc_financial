package blockchain

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

// ContractHelper 智能合约调用辅助类
type ContractHelper struct {
	abi          abi.ABI
	contractAddr common.Address
}

// NewContractHelper 创建合约辅助类
func NewContractHelper(abiJSON string, contractAddr string) (*ContractHelper, error) {
	parsedABI, err := abi.JSON(json.Unmarshaler{}) // 正确的方式
	if err != nil {
		return nil, fmt.Errorf("failed to parse ABI: %w", err)
	}

	return &ContractHelper{
		abi:          parsedABI,
		contractAddr: common.HexToAddress(contractAddr),
	}, nil
}

// EncodeUploadTransaction 编码 uploadTransaction 方法调用
// bizId: 业务流水号 (字符串,会被转为bytes32)
// dataHash: 数据哈希 (字符串,会被转为bytes32)
func (h *ContractHelper) EncodeUploadTransaction(bizId, dataHash string) ([]byte, error) {
	// 将字符串转为bytes32
	bizIdBytes32 := stringToBytes32(bizId)
	dataHashBytes32, err := parseHashToBytes32(dataHash)
	if err != nil {
		return nil, fmt.Errorf("invalid data hash: %w", err)
	}

	// 使用ABI编码
	data, err := h.abi.Pack("uploadTransaction", bizIdBytes32, dataHashBytes32)
	if err != nil {
		return nil, fmt.Errorf("failed to pack uploadTransaction: %w", err)
	}

	return data, nil
}

// EncodeBatchUploadTransactions 编码 batchUploadTransactions 方法调用
func (h *ContractHelper) EncodeBatchUploadTransactions(bizIds, dataHashes []string) ([]byte, error) {
	if len(bizIds) != len(dataHashes) {
		return nil, fmt.Errorf("bizIds and dataHashes length mismatch")
	}

	bizIdArray := make([][32]byte, len(bizIds))
	hashArray := make([][32]byte, len(dataHashes))

	for i, bizId := range bizIds {
		bizIdArray[i] = stringToBytes32(bizId)
	}

	for i, hash := range dataHashes {
		h, err := parseHashToBytes32(hash)
		if err != nil {
			return nil, fmt.Errorf("invalid data hash at index %d: %w", i, err)
		}
		hashArray[i] = h
	}

	data, err := h.abi.Pack("batchUploadTransactions", bizIdArray, hashArray)
	if err != nil {
		return nil, fmt.Errorf("failed to pack batchUploadTransactions: %w", err)
	}

	return data, nil
}

// EncodeRegisterInstitution 编码 registerInstitution 方法调用
func (h *ContractHelper) EncodeRegisterInstitution(name, address string) ([]byte, error) {
	addr := common.HexToAddress(address)

	data, err := h.abi.Pack("registerInstitution", name, addr)
	if err != nil {
		return nil, fmt.Errorf("failed to pack registerInstitution: %w", err)
	}

	return data, nil
}

// DecodeGetTransaction 解码 getTransaction 方法的返回值
func (h *ContractHelper) DecodeGetTransaction(data []byte) (*TransactionResult, error) {
	// 解码返回值: (bytes32, address, uint256, uint8, address, uint256)
	results, err := h.abi.Methods["getTransaction"].Outputs.Unpack(data)
	if err != nil {
		return nil, fmt.Errorf("failed to unpack getTransaction: %w", err)
	}

	if len(results) != 6 {
		return nil, fmt.Errorf("unexpected number of return values: %d", len(results))
	}

	result := &TransactionResult{
		DataHash:     results[0].([32]byte),
		Uploader:     results[1].(common.Address).Hex(),
		Timestamp:    results[2].(*big.Int),
		Status:       uint8(results[3].(uint8)),
		Counterparty: results[4].(common.Address).Hex(),
		MatchHeight:  results[5].(*big.Int),
	}

	return result, nil
}

// DecodeGetStatistics 解码 getStatistics 方法的返回值
func (h *ContractHelper) DecodeGetStatistics(data []byte) (*StatisticsResult, error) {
	results, err := h.abi.Methods["getStatistics"].Outputs.Unpack(data)
	if err != nil {
		return nil, fmt.Errorf("failed to unpack getStatistics: %w", err)
	}

	if len(results) != 4 {
		return nil, fmt.Errorf("unexpected number of return values: %d", len(results))
	}

	result := &StatisticsResult{
		TotalTx:          results[0].(*big.Int),
		TotalMatched:     results[1].(*big.Int),
		MatchRate:        results[2].(*big.Int),
		InstitutionCount: results[3].(*big.Int),
	}

	return result, nil
}

// GetContractAddress 获取合约地址
func (h *ContractHelper) GetContractAddress() common.Address {
	return h.contractAddr
}

// ========== 辅助函数 ==========

// stringToBytes32 将字符串转为bytes32
func stringToBytes32(s string) [32]byte {
	var result [32]byte
	copy(result[:], s)
	return result
}

// parseHashToBytes32 解析哈希字符串为bytes32
func parseHashToBytes32(hash string) ([32]byte, error) {
	var result [32]byte

	// 移除0x前缀(如果有)
	if len(hash) >= 2 && hash[0:2] == "0x" {
		hash = hash[2:]
	}

	// 如果是64位hex字符串
	if len(hash) == 64 {
		bytes, err := hex.DecodeString(hash)
		if err != nil {
			return result, fmt.Errorf("invalid hex: %w", err)
		}
		copy(result[:], bytes)
		return result, nil
	}

	// 否则当作普通字符串处理
	return stringToBytes32(hash), nil
}

// CalculateDataHash 计算数据哈希 (Keccak256)
func CalculateDataHash(bizId, amount, salt string) string {
	data := fmt.Sprintf("%s:%s:%s", bizId, amount, salt)
	hash := crypto.Keccak256Hash([]byte(data))
	return hash.Hex()
}

// BizIdToBytes32 将业务流水号转为bytes32格式
func BizIdToBytes32(bizId string) [32]byte {
	return stringToBytes32(bizId)
}

// ========== 数据结构 (解码结果) ==========

// TransactionResult 解码后的交易结果
type TransactionResult struct {
	DataHash     [32]byte
	Uploader     string
	Timestamp    *big.Int
	Status       uint8
	Counterparty string
	MatchHeight  *big.Int
}

// StatisticsResult 解码后的统计结果
type StatisticsResult struct {
	TotalTx          *big.Int
	TotalMatched     *big.Int
	MatchRate        *big.Int
	InstitutionCount *big.Int
}
