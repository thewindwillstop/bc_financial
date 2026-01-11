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
	// 调试：打印当前工作目录，确保它能找到文件
	dir, _ := os.Getwd()
	fmt.Printf("--- 调试信息 ---\n当前路径: %s\n", dir)

	// 1. 加载配置
	configs, err := conf.ParseConfigFile("config.toml")
	if err != nil {
		// 如果这里报错，说明 TOML 语法有问题或者文件找不到
		log.Fatalf("解析失败 (Syntax/Path Error): %v", err)
	}

	// 2. 打印解析出来的具体内容，看看哪个字段没合上
	fmt.Println("--- 配置解析结果 ---")
	if len(configs) > 0 {
		c := configs[0]
		fmt.Printf("CAFile 路径: '%s'\n", c.CAFile)
		
		// 关键点：看看这里是不是 0
		fmt.Printf("Chain ID: %d\n", c.ChainID)
	}
	fmt.Println("--------------------")

	// 3. 连接节点
	c, err := client.Dial(&configs[0])
	if err != nil {
		log.Fatalf("连接失败 (Dial Error): %v", err)
	}

	blockNumber, err := c.GetBlockNumber(context.Background())
	if err != nil {
		log.Fatalf("获取块高失败: %v", err)
	}

	fmt.Printf("🎉 连接成功！当前区块高度为: %d\n", blockNumber)
}
