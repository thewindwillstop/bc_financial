# 🎓 基于联盟链的金融交易数据分布式对账与溯源系统

> 毕业设计项目 | FISCO BCOS | Go | MySQL | Solidity

---

## ✨ 项目简介

本项目实现了一个基于**FISCO BCOS联盟链**的金融交易数据分布式对账系统,通过**"哈希上链,链上碰撞"**机制实现隐私保护的自动对账功能。

### 核心特性

- 🔐 **隐私保护**: 仅上传交易哈希到区块链,原始加密数据存储在MySQL
- ⚡ **自动对账**: 通过智能合约哈希碰撞实现自动对账
- 🔗 **完整追溯**: 区块链保证数据不可篡改,完整可追溯
- 📊 **批量处理**: 支持Excel批量导入导出
- 🎯 **事件驱动**: 实时监听对账事件

---

## 🏗️ 系统架构

```
用户 → Go API → (MySQL数据库 + FISCO BCOS智能合约)
```

**技术栈**:
- **后端**: Go 1.21 + Gin框架
- **区块链**: FISCO BCOS 2.9.2 (4节点联盟链)
- **数据库**: MySQL 8.0
- **智能合约**: Solidity ^0.4.25
- **合约地址**: `0xeed55f17ea7d7646681f34fe95a6a5cfe003cdc3`

---

## 🚀 快速开始

### 1. 启动后端服务

```bash
cd /home/lin123456/colloge_project/bc_financial/backend
go run cmd/api/main.go
```

服务启动在 `http://localhost:8080`

### 2. 启动FISCO控制台

```bash
cd fisco/nodes/127.0.0.1/console
bash start.sh
```

### 3. 测试API

```bash
# 健康检查
curl http://localhost:8080/health

# 查询统计
curl http://localhost:8080/api/v1/dashboard/statistics
```

---

## 📖 完整文档

📘 **[PROJECT_GUIDE.md](./PROJECT_GUIDE.md)** - 完整的项目使用指南

包含内容:
- ✅ 详细的项目架构说明
- ✅ API接口文档
- ✅ 智能合约使用教程
- ✅ 完整的使用流程示例
- ✅ 技术实现细节
- ✅ 常见问题解答

📄 **[idea.md](./idea.md)** - 项目设计思路和需求分析

---

## 📡 API端点

### 健康检查
- `GET /health`

### 交易管理
- `POST /api/v1/transactions` - 创建交易
- `POST /api/v1/transactions/excel` - 上传Excel
- `GET /api/v1/transactions` - 交易列表
- `GET /api/v1/transactions/:bizId` - 查询详情
- `POST /api/v1/transactions/upload-chain` - 批量上链
- `GET /api/v1/transactions/template` - 下载Excel模板

### 仪表板
- `GET /api/v1/dashboard/overview` - 概览数据
- `GET /api/v1/dashboard/statistics` - 统计数据
- `GET /api/v1/dashboard/chart-data` - 图表数据

---

## 🔐 智能合约使用

### 快速示例

```bash
# 1. 注册机构
[group:1]> call Reconciliation.sol registerInstitution "机构A" "0x地址"

# 2. 上传交易(第一次)
[group:1]> call Reconciliation.sol uploadTransaction 0x[bizId哈希] 0x[数据哈希]

# 3. 上传交易(第二次 - 触发对账)
[group:1]> call Reconciliation.sol uploadTransaction 0x[相同bizId] 0x[数据哈希]

# 4. 查询结果
[group:1]> call Reconciliation.sol getTransaction 0x[bizId]
```

详细信息请查看 [PROJECT_GUIDE.md](./PROJECT_GUIDE.md)

---

## 📊 使用流程

### 典型场景: 两家机构对账

1. **机构A操作**
   - 通过API上传Excel到数据库
   - 在控制台注册机构
   - 计算哈希并上链

2. **机构B操作**
   - 通过API上传Excel
   - 在控制台注册机构
   - 上传相同业务流水号

3. **系统自动对账**
   - 哈希相同 → 对账成功 (status=2)
   - 哈希不同 → 对账失败 (status=3)

4. **查看结果**
   - 通过控制台查询链上数据
   - 通过API查询数据库记录

---

## 📁 项目结构

```
bc_financial/
├── backend/                 # Go后端服务
│   ├── cmd/api/            # 程序入口
│   ├── internal/           # 内部模块
│   │   ├── handler/        # HTTP处理器
│   │   ├── service/        # 业务逻辑
│   │   ├── models/         # 数据模型
│   │   ├── blockchain/     # 区块链客户端
│   │   └── utils/          # 工具函数
│   ├── contracts/          # 智能合约文件
│   └── configs/            # 配置文件
├── database/
│   └── schema.sql          # 数据库schema
├── fisco/                  # FISCO BCOS节点
│   └── nodes/127.0.0.1/    # 4节点部署
└── contracts/
    └── Reconciliation.sol  # 智能合约源码
```

---

## ✅ 已实现功能

- [x] Go RESTful API服务 (9个端点)
- [x] MySQL数据持久化
- [x] FISCO BCOS区块链集成
- [x] 智能合约部署
- [x] Excel批量导入导出
- [x] 隐私保护的哈希对账
- [x] 事件监听框架
- [x] 统计分析功能

---

## 📝 下一步计划

- [ ] 完整的Go合约绑定(当前使用控制台)
- [ ] 事件监听服务
- [ ] 用户认证系统
- [ ] 前端界面开发

---

## 🛠️ 技术栈

| 组件 | 技术 | 说明 |
|------|------|------|
| 后端框架 | Gin 1.9.1 | HTTP服务 |
| ORM | GORM 1.25.5 | 数据库操作 |
| 区块链 | FISCO BCOS 2.9.2 | 联盟链平台 |
| 智能合约 | Solidity ^0.4.25 | 合约语言 |
| 数据库 | MySQL 8.0 | 数据存储 |
| Excel | excelize/v2 | 表格处理 |

---

## 👨‍💻 作者信息

**项目**: 基于联盟链的金融交易数据分布式对账与溯源系统
**类型**: 毕业设计
**完成时间**: 2025年1月

---

**📘 详细文档请查看 [PROJECT_GUIDE.md](./PROJECT_GUIDE.md)**

**祝你使用愉快!** 🎉
