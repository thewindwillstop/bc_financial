#!/bin/bash

# 🚀 BC Reconciliation 后端服务启动脚本

echo "========================================"
echo "  金融交易对账系统 - 后端服务启动"
echo "========================================"
echo ""

# 颜色定义
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# 切换到后端目录
cd "$(dirname "$0")"

echo -e "${YELLOW}[1/5] 检查FISCO BCOS节点...${NC}"
cd ../fisco/nodes/127.0.0.1
NODE_COUNT=$(ps aux | grep fisco-bcos | grep -v grep | wc -l)
if [ $NODE_COUNT -ge 4 ]; then
    echo -e "${GREEN}✓ FISCO节点运行正常 (4个节点)${NC}"
else
    echo -e "${YELLOW}! FISCO节点未运行,正在启动...${NC}"
    bash start_all.sh
    sleep 2
fi
echo ""

echo -e "${YELLOW}[2/5] 检查MySQL数据库...${NC}"
if command -v mysql &> /dev/null; then
    echo -e "${GREEN}✓ MySQL已安装${NC}"

    # 尝试连接MySQL
    if mysql -u root -e "USE bc_reconciliation;" 2>/dev/null; then
        echo -e "${GREEN}✓ 数据库已存在${NC}"
    else
        echo -e "${YELLOW}! 需要初始化数据库${NC}"
        echo "请执行: mysql -u root -p < ../database/schema.sql"
        read -p "按Enter继续 (假设数据库已配置)..."
    fi
else
    echo -e "${RED}✗ MySQL未安装${NC}"
    exit 1
fi
echo ""

cd ../../backend

echo -e "${YELLOW}[3/5] 检查配置文件...${NC}"
if [ ! -f "configs/config.yaml" ]; then
    echo -e "${RED}✗ 配置文件不存在${NC}"
    exit 1
fi
echo -e "${GREEN}✓ 配置文件存在${NC}"
echo ""

echo -e "${YELLOW}[4/5] 下载Go依赖...${NC}"
if [ ! -d "vendor" ]; then
    go mod tidy
    if [ $? -eq 0 ]; then
        echo -e "${GREEN}✓ 依赖下载成功${NC}"
    else
        echo -e "${RED}✗ 依赖下载失败${NC}"
        exit 1
    fi
else
    echo -e "${GREEN}✓ 依赖已存在${NC}"
fi
echo ""

echo -e "${YELLOW}[5/5] 启动服务...${NC}"
echo ""
echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}  服务正在启动...${NC}"
echo -e "${GREEN}  访问地址: http://localhost:8080${NC}"
echo -e "${GREEN}  健康检查: http://localhost:8080/health${NC}"
echo -e "${GREEN}========================================${NC}"
echo ""
echo "按 Ctrl+C 停止服务"
echo ""

# 启动服务
go run cmd/api/main.go
