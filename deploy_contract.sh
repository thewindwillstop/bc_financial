#!/bin/bash

# Java安装和智能合约部署脚本

echo "========================================"
echo "  智能合约部署向导"
echo "========================================"
echo ""

# 检查是否已安装Java
if command -v java &> /dev/null; then
    echo "✅ Java已安装:"
    java -version 2>&1 | head -1
    echo ""
    read -p "是否继续部署合约? (y/n): " choice
    if [ "$choice" != "y" ]; then
        echo "已取消"
        exit 0
    fi
else
    echo "⚠️  未检测到Java"
    echo ""
    echo "请先安装Java JRE:"
    echo ""
    echo "  sudo apt update"
    echo "  sudo apt install -y default-jre"
    echo ""
    echo "安装完成后重新运行此脚本"
    echo ""
    exit 1
fi

echo "========================================"
echo "  开始部署智能合约"
echo "========================================"
echo ""

# 进入控制台目录
cd /home/lin123456/colloge_project/bc_financial/fisco/nodes/127.0.0.1

# 复制合约到控制台
echo "[1/5] 复制合约文件到控制台..."
cp ../../contracts/Reconciliation.sol console/contracts/solidity/
if [ $? -eq 0 ]; then
    echo "✅ 合约文件已复制"
else
    echo "❌ 复制失败"
    exit 1
fi

# 配置控制台
echo ""
echo "[2/5] 配置控制台..."
cd console

if [ ! -f "conf/config.toml" ]; then
    cp conf/config-example.toml conf/config.toml
    echo "✅ 配置文件已创建"
else
    echo "✅ 配置文件已存在"
fi

echo ""
echo "[3/5] 准备启动控制台..."
echo ""
echo "========================================"
echo "  重要提示"
echo "========================================"
echo ""
echo "控制台启动后,请执行以下命令:"
echo ""
echo "  1. 部署合约:"
echo "     [group:1]> deploy Reconciliation.sol"
echo ""
echo "  2. 查看合约地址(会显示在输出中)"
echo ""
echo "  3. 退出控制台:"
echo "     [group:1]> exit"
echo ""
echo "========================================"
echo ""

read -p "按Enter启动控制台..."

# 启动控制台
bash start.sh
