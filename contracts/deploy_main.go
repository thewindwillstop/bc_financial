package main

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"os"

	"github.com/FISCO-BCOS/go-sdk/client"
	"github.com/FISCO-BCOS/go-sdk/conf"
	"github.com/FISCO-BCOS/go-sdk/core/contract"
	"github.com/ethereum/go-ethereum/common"
)

// Reconciliation 合约接口
type Reconciliation struct {
	contract.Contract
}

// 部署合约
func DeployReconciliation(client *client.Client, from common.Address) (common.Address, *Reconciliation, error) {
	// 读取合约ABI
	abiPath := "contracts/Reconciliation.abi.json"
	bytecodePath := "contracts/Reconciliation.bin"

	// 检查文件是否存在
	if _, err := os.Stat(abiPath); os.IsNotExist(err) {
		return common.Address{}, nil, fmt.Errorf("ABI file not found: %s", abiPath)
	}
	if _, err := os.Stat(bytecodePath); os.IsNotExist(err) {
		return common.Address{}, nil, fmt.Errorf("Bytecode file not found: %s", bytecodePath)
	}

	// 这里需要实际的ABI和bytecode
	// 由于没有solc编译器,我们需要其他方法
	return common.Address{}, nil, fmt.Errorf("need solc compiler to generate ABI and bytecode")
}

func main() {
	// 切换到项目根目录
	os.Chdir("/home/lin123456/colloge_project/bc_financial")

	// 加载配置
	configs, err := conf.ParseConfigFile("go-project/fisco-recon/config.toml")
	if err != nil {
		log.Fatalf("解析配置失败: %v", err)
	}

	// 连接节点
	c, err := client.Dial(&configs[0])
	if err != nil {
		log.Fatalf("连接节点失败: %v", err)
	}

	// 测试连接
	blockNumber, err := c.GetBlockNumber(context.Background())
	if err != nil {
		log.Fatalf("获取区块高度失败: %v", err)
	}

	fmt.Printf("✅ 成功连接到FISCO BCOS节点\n")
	fmt.Printf("📊 当前区块高度: %d\n", blockNumber)

	// 测试基本功能
	fmt.Println("\n🔍 测试节点信息:")

	// 获取系统配置
	chainID, err := c.GetSystemConfigByKey(context.Background(), "chain_id")
	if err != nil {
		fmt.Printf("⚠️  获取chain_id失败: %v\n", err)
	} else {
		fmt.Printf("   Chain ID: %s\n", chainID)
	}

	// 获取区块信息
	block, err := c.GetBlockByNumber(context.Background(), big.NewInt(blockNumber), false)
	if err != nil {
		fmt.Printf("⚠️  获取区块信息失败: %v\n", err)
	} else {
		fmt.Printf("   最新区块哈希: %s\n", block.Hash.Hex())
		fmt.Printf("   交易数量: %d\n", len(block.Transactions))
	}

	fmt.Println("\n✅ FISCO BCOS连接测试通过!")
	fmt.Println("\n📝 下一步:")
	fmt.Println("   1. 安装solc编译器生成合约ABI和bytecode")
	fmt.Println("   2. 或者使用web3j/控制台部署合约")
	fmt.Println("   3. 使用Go SDK调用合约")
}
