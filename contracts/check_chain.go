package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/FISCO-BCOS/go-sdk/client"
	"github.com/FISCO-BCOS/go-sdk/conf"
)

func main() {
	// 切换到项目根目录
	os.Chdir("/home/lin123456/colloge_project/bc_financial/go-project/fisco-recon")

	fmt.Println("🚀 开始测试FISCO BCOS连接...\n")

	// 加载配置
	configs, err := conf.ParseConfigFile("config.toml")
	if err != nil {
		log.Fatalf("❌ 解析配置失败: %v", err)
	}
	fmt.Println("✅ 配置文件加载成功")

	// 连接节点
	c, err := client.Dial(&configs[0])
	if err != nil {
		log.Fatalf("❌ 连接节点失败: %v", err)
	}
	fmt.Println("✅ 成功连接到FISCO BCOS节点")

	// 测试连接
	blockNumber, err := c.GetBlockNumber(context.Background())
	if err != nil {
		log.Fatalf("❌ 获取区块高度失败: %v", err)
	}

	fmt.Printf("\n📊 当前区块高度: %d\n", blockNumber)

	// 测试基本功能
	fmt.Println("\n🔍 测试节点信息:")

	// 获取系统配置
	chainID, err := c.GetSystemConfigByKey(context.Background(), "chain_id")
	if err != nil {
		fmt.Printf("⚠️  获取chain_id失败: %v\n", err)
	} else {
		fmt.Printf("   ✅ Chain ID: %s\n", chainID)
	}

	// 获取区块信息
	block, err := c.GetBlockByNumber(context.Background(), blockNumber, false)
	if err != nil {
		fmt.Printf("⚠️  获取区块信息失败: %v\n", err)
	} else {
		fmt.Printf("   ✅ 最新区块哈希: %x\n", block.Hash)
		fmt.Printf("   ✅ 区块时间戳: %d\n", block.Timestamp)
		fmt.Printf("   ✅ 交易数量: %d\n", len(block.Transactions))
	}

	// 获取节点ID列表
	nodeIDs, err := c.GetNodeIDList(context.Background())
	if err != nil {
		fmt.Printf("⚠️  获取节点列表失败: %v\n", err)
	} else {
		fmt.Printf("   ✅ 共有 %d 个共识节点\n", len(nodeIDs))
	}

	fmt.Println("\n==================================================")
	fmt.Println("✅ FISCO BCOS连接测试通过!")
	fmt.Println("==================================================")

	fmt.Println("\n📝 下一步建议:")
	fmt.Println("   1. ✅ FISCO BCOS节点运行正常")
	fmt.Println("   2. 需要安装solc编译器来编译智能合约")
	fmt.Println("   3. 或者安装Java控制台来部署合约")
}
