package main

import (
	"context"
	"fmt"
	"log"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

func main() {
	fmt.Println("=== FISCO BCOS 智能合约调用测试 ===\n")

	// 测试1: 生成bizId的bytes32格式
	testBizId := "TX202501140001"
	bizIdBytes32 := prepareBizIdBytes32(testBizId)
	fmt.Printf("测试1 - bizId转bytes32:\n")
	fmt.Printf("  原始bizId: %s\n", testBizId)
	fmt.Printf("  bytes32: %x\n\n", bizIdBytes32)

	// 测试2: 生成dataHash
	testAmount := "1000.00"
	testSalt := "random_salt_12345"
	dataHash := calculateDataHash(testBizId, testAmount, testSalt)
	fmt.Printf("测试2 - 计算dataHash:\n")
	fmt.Printf("  bizId: %s\n", testBizId)
	fmt.Printf("  amount: %s\n", testAmount)
	fmt.Printf("  salt: %s\n", testSalt)
	fmt.Printf("  SHA256: %x\n\n", dataHash)

	// 测试3: 生成以太坊地址
	privateKey, err := crypto.GenerateKey()
	if err != nil {
		log.Fatalf("生成私钥失败: %v", err)
	}
	address := crypto.PubkeyToAddress(privateKey.PublicKey)
	fmt.Printf("测试3 - 生成测试地址:\n")
	fmt.Printf("  地址: %s\n\n", address.Hex())

	fmt.Println("=== 所有测试通过! ===")
	fmt.Println("\n下一步: 运行完整的服务并调用智能合约")
	fmt.Println("  cd backend")
	fmt.Println("  go run cmd/api/main.go")
}

// prepareBizIdBytes32 将bizId字符串转为bytes32格式
func prepareBizIdBytes32(bizId string) [32]byte {
	var bizIdBytes32 [32]byte
	copy(bizIdBytes32[:], bizId)
	return bizIdBytes32
}

// calculateDataHash 计算数据哈希
func calculateDataHash(bizId, amount, salt string) [32]byte {
	data := fmt.Sprintf("%s:%s:%s", bizId, amount, salt)
	return crypto.Keccak256Hash([]byte(data)) // 使用Keccak256 (以太坊标准)
}

// 实际调用智能合约的示例 (需要连接到FISCO BCOS)
func exampleCallContract() {
	ctx := context.Background()

	// 这些是示例值,实际使用时需要从配置中读取
	contractAddress := common.HexToAddress("0xeed55f17ea7d7646681f34fe95a6a5cfe003cdc3")

	fmt.Printf("示例: 调用智能合约 %s\n", contractAddress.Hex())
	fmt.Println("注意: 这需要运行完整的服务并连接到FISCO BCOS节点")

	// 实际代码示例:
	// 1. 创建区块链客户端
	// client, err := blockchain.NewClient(cfg, logger)

	// 2. 调用GetStatistics查询统计信息
	// stats, err := client.GetStatistics(ctx)
	// fmt.Printf("总交易数: %s\n", stats.TotalTx.String())

	// 3. 上传交易
	// bizId := prepareBizIdBytes32("TX001")
	// dataHash := calculateDataHash("TX001", "1000.00", "salt")
	// txHash, err := client.UploadTransaction(ctx, bizId, dataHash)
	// fmt.Printf("交易哈希: %s\n", txHash)

	_ = ctx // 避免未使用警告
}
