# 🎓 基于联盟链的金融交易数据分布式对账与溯源系统

> 毕业设计项目 - 完整实现指南

---

## 📋 项目概述

### 项目简介
本项目实现了一个基于FISCO BCOS联盟链的金融交易数据分布式对账系统,通过"哈希上链,链上碰撞"机制实现隐私保护的对账功能。

### 核心特性
- ✅ **隐私保护**: 仅上传交易哈希到区块链,原始加密数据存储在MySQL
- ✅ **自动对账**: 通过哈希碰撞实现自动对账
- ✅ **完整追溯**: 区块链保证数据不可篡改
- ✅ **事件驱动**: 实时监听对账事件
- ✅ **批量处理**: 支持Excel批量导入

---

## 🏗️ 系统架构

```
┌─────────────────────────────────────────────────────────┐
│                      前端界面(待开发)                    │
└────────────────────┬────────────────────────────────────┘
                     │
                     ↓
┌────────────────────────────────────────────────────────┐
│              Go API 服务 (端口: 8080)                   │
│  ┌──────────────┐  ┌──────────────┐  ┌─────────────┐ │
│  │  Handler层   │→ │  Service层   │→ │  Model层    │ │
│  │  (参数验证)   │  │  (业务逻辑)   │  │  (数据模型)  │ │
│  └──────────────┘  └──────────────┘  └─────────────┘ │
└────────┬───────────────────────────────────────────────┘
         │
    ┌────┴────┐
    ↓         ↓
┌──────────────┐  ┌─────────────────────────────────┐
│   MySQL      │  │      FISCO BCOS 联盟链           │
│   数据库     │  │  ┌──────────────────────────┐   │
│              │  │  │  Reconciliation 智能合约  │   │
│ - 交易数据   │  │  │  0xeed55f17...003cdc3     │   │
│ - 加密存储   │  │  └──────────────────────────┘   │
│ - 对账结果   │  │                                 │
└──────────────┘  └─────────────────────────────────┘
```

---

## 📦 项目结构

```
bc_financial/
├── backend/                      # Go后端服务
│   ├── cmd/
│   │   └── api/
│   │       └── main.go          # 程序入口
│   ├── internal/
│   │   ├── config/              # 配置管理
│   │   ├── handler/             # HTTP处理器
│   │   ├── service/             # 业务逻辑层
│   │   ├── models/              # 数据模型
│   │   ├── middleware/          # 中间件
│   │   ├── utils/               # 工具函数
│   │   ├── database/            # 数据库连接
│   │   └── blockchain/          # 区块链客户端
│   ├── configs/
│   │   └── config.yaml          # 配置文件
│   ├── contracts/               # 智能合约文件
│   │   ├── Reconciliation.abi   # 合约ABI
│   │   └── Reconciliation.bin   # 合约BIN
│   └── go.mod
├── database/
│   └── schema.sql               # 数据库schema
├── fisco/
│   └── nodes/127.0.0.1/         # FISCO BCOS节点
│       ├── node0/               # 节点0-3
│       └── console/             # 控制台
└── contracts/
    └── Reconciliation.sol       # 智能合约源码
```

---

## 🚀 快速开始

### 1. 环境准备

#### 已安装的软件
- ✅ Go 1.21
- ✅ MySQL 8.0
- ✅ Java JRE (用于FISCO控制台)
- ✅ FISCO BCOS 2.9.2 (4节点)

#### 数据库配置
```sql
-- 创建数据库
CREATE DATABASE bc_reconciliation CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

-- 导入schema
mysql -u root -p bc_reconciliation < database/schema.sql;
```

---

### 2. 启动后端服务

```bash
cd /home/lin123456/colloge_project/bc_financial/backend

# 运行服务
go run cmd/api/main.go

# 服务启动在 http://localhost:8080
```

**启动成功的标志**:
```
✅ connected to FISCO BCOS  block_number=0
✅ smart contract loaded successfully  contract_address=0xeed55f17...
✅ HTTP server started  port=8080
```

---

### 3. 启动FISCO控制台

```bash
cd /home/lin123456/colloge_project/bc_financial/fisco/nodes/127.0.0.1/console
bash start.sh
```

---

## 🔌 API端点

### 健康检查
```bash
GET /health
```

### 交易管理
```bash
# 创建交易
POST /api/v1/transactions
Content-Type: application/json
{
  "biz_id": "TX20250113001",
  "institution_id": 1,
  "amount": "1000.00",
  "sender": "机构A",
  "receiver": "机构B",
  "transaction_time": "2025-01-13T10:00:00Z"
}

# 上传Excel
POST /api/v1/transactions/excel
Content-Type: multipart/form-data
file: transactions.xlsx

# 查询交易列表
GET /api/v1/transactions?page=1&size=10

# 查询交易详情
GET /api/v1/transactions/{bizId}

# 批量上链
POST /api/v1/transactions/upload-chain

# 下载Excel模板
GET /api/v1/transactions/template
```

### 仪表板
```bash
# 概览数据
GET /api/v1/dashboard/overview

# 统计数据
GET /api/v1/dashboard/statistics

# 图表数据
GET /api/v1/dashboard/chart-data
```

---

## 🔐 智能合约使用指南

### 已部署的合约信息
- **合约地址**: `0xeed55f17ea7d7646681f34fe95a6a5cfe003cdc3`
- **合约文件**: `fisco/nodes/127.0.0.1/console/contracts/solidity/Reconciliation.sol`
- **Solidity版本**: ^0.4.25

### 合约功能

#### 1. 注册机构
```bash
[group:1]> call Reconciliation.sol registerInstitution "机构A" "0xbe42da26d0a2681f83f24a346df80616195ed5a0"
```

#### 2. 上传交易哈希(第一次)
```bash
[group:1]> call Reconciliation.sol uploadTransaction 0x[bizId哈希] 0x[数据哈希]
```

#### 3. 上传交易哈希(第二次 - 触发对账)
```bash
# 对方机构上传相同的业务流水号
[group:1]> call Reconciliation.sol uploadTransaction 0x[相同的bizId] 0x[数据哈希]

# 结果:
# - 哈希相同 → status=2 (对账成功)
# - 哈希不同 → status=3 (对账失败)
```

#### 4. 查询交易
```bash
[group:1]> call Reconciliation.sol getTransaction 0x[bizId]
```

#### 5. 查询统计
```bash
[group:1]> call Reconciliation.sol getStatistics
```

### 合约事件
- **DataUploaded**: 数据上传事件
- **ReconciliationEvent**: 对账完成事件
- **InstitutionRegistered**: 机构注册事件

---

## 📊 完整使用流程示例

### 场景: 两家机构进行交易对账

#### 第一步: 机构A上传数据
```bash
# 1. 通过API上传Excel到数据库
curl -X POST http://localhost:8080/api/v1/transactions/excel \
  -F "file=@transactions_a.xlsx"

# 2. 在控制台注册机构(首次)
[group:1]> call Reconciliation.sol registerInstitution "机构A" "0x[A的地址]"

# 3. 计算哈希并上链
# 在Go代码中:
dataHash := sha256(bizId + amount + timestamp)
# 例如: 0x1234...abcd (64位hex字符)

# 4. 在控制台上传
[group:1]> call Reconciliation.sol uploadTransaction 0x[bizId] 0x[dataHash]
```

#### 第二步: 机构B上传数据
```bash
# 1. 通过API上传Excel
curl -X POST http://localhost:8080/api/v1/transactions/excel \
  -F "file=@transactions_b.xlsx"

# 2. 在控制台注册机构(首次)
[group:1]> call Reconciliation.sol registerInstitution "机构B" "0x[B的地址]"

# 3. 上传相同的bizId
[group:1]> call Reconciliation.sol uploadTransaction 0x[相同bizId] 0x[dataHash]

# 系统自动对账!
```

#### 第三步: 查看结果
```bash
# 在控制台查询
[group:1]> call Reconciliation.sol getTransaction 0x[bizId]

# 返回:
# status = 2  → 对账成功 (双方数据一致)
# status = 3  → 对账失败 (数据不一致)

# 通过API查询数据库
curl http://localhost:8080/api/v1/transactions/[bizId]
```

---

## 🔧 技术实现细节

### 隐私保护机制
```
原始数据(加密存储在MySQL)
         ↓
    SHA256哈希
         ↓
区块链存储(仅哈希值)
```

### 哈希计算方法
```go
// internal/utils/hash.go
func CalculateDataHash(bizId, amount string, timestamp int64) string {
    data := fmt.Sprintf("%s:%s:%d", bizId, amount, timestamp)
    hash := sha256.Sum256([]byte(data))
    return hex.EncodeToString(hash[:])
}
```

### 对账状态
- `0` - PENDING: 待处理
- `1` - UPLOADED: 已上传(单方)
- `2` - MATCHED: 对账成功
- `3` - MISMATCH: 对账失败

---

## 📁 数据库表结构

### institutions (机构表)
```sql
- id: 机构ID
- name: 机构名称
- address: 区块链地址
- created_at: 创建时间
```

### transactions (交易表)
```sql
- id: 交易ID
- biz_id: 业务流水号
- institution_id: 机构ID
- amount: 金额(加密)
- sender: 发送方
- receiver: 接收方
- status: 对账状态
- transaction_hash: 链上交易哈希
```

### chain_receipts (链上收据表)
```sql
- id: 收据ID
- transaction_id: 交易ID
- tx_hash: 交易哈希
- block_number: 区块号
- gas_used: Gas消耗
```

### reconciliations (对账记录表)
```sql
- id: 对账ID
- biz_id: 业务流水号
- status: 对账状态
- matched_at: 匹配时间
```

---

## ⚙️ 配置文件

### backend/configs/config.yaml
```yaml
server:
  port: 8080
  mode: debug

database:
  mysql:
    host: localhost
    port: 3306
    username: root
    password: lin123456
    database: bc_reconciliation

blockchain:
  config_file: /path/to/config.toml
  contract_address: "0xeed55f17ea7d7646681f34fe95a6a5cfe003cdc3"

log:
  level: info
  filename: logs/app.log
```

---

## 🐛 常见问题

### Q1: 服务启动失败
**A**: 检查以下几点:
1. MySQL是否启动: `sudo service mysql status`
2. FISCO节点是否运行: `ps aux | grep fisco`
3. 配置文件路径是否正确

### Q2: 合约调用失败
**A**:
1. 确认机构已注册
2. 检查地址格式(0x开头)
3. 确认哈希值是64位hex字符
4. 检查合约是否被暂停

### Q3: 哈希值如何计算?
**A**: 使用Go工具函数:
```go
bizId := "TX20250113001"
amount := "1000.00"
timestamp := time.Now().Unix()
hash := utils.CalculateDataHash(bizId, amount, timestamp)
```

### Q4: 如何查看区块链数据?
**A**:
1. 使用控制台: `call Reconciliation.sol getTransaction 0x[bizId]`
2. 查看日志: `fisco/nodes/127.0.0.1/node*/log/log*.log`
3. 使用FISCO自带的区块链浏览器

---

## 📈 系统测试

### 测试API健康
```bash
curl http://localhost:8080/health
# 预期输出: {"status":"ok","time":"..."}
```

### 测试Excel上传
```bash
# 下载模板
curl -O http://localhost:8080/api/v1/transactions/template

# 填写数据后上传
curl -X POST http://localhost:8080/api/v1/transactions/excel \
  -F "file=@test_transactions.xlsx"
```

### 测试统计查询
```bash
curl http://localhost:8080/api/v1/dashboard/statistics
```

---

## 📝 下一步开发计划

### 短期(已完成)
- ✅ 基础API服务
- ✅ 数据库存储
- ✅ 智能合约部署
- ✅ Excel导入导出

### 中期(可选)
- [ ] 完整的Go合约绑定
- [ ] 事件监听服务
- [ ] 用户认证系统
- [ ] 前端界面

### 长期(扩展)
- [ ] 多机构支持
- [ ] 数据分析报表
- [ ] 实时监控告警
- [ ] 性能优化

---

## 📚 参考文档

### FISCO BCOS文档
- 官方文档: https://fisco-bcos-documentation.readthedocs.io/
- 控制手册: `fisco/nodes/127.0.0.1/console/README.md`

### Go框架文档
- Gin框架: https://gin-gonic.com/docs/
- GORM文档: https://gorm.io/docs/

### 项目文档
- 智能合约源码: `contracts/Reconciliation.sol`
- 数据库schema: `database/schema.sql`
- API文档: 见上方"API端点"章节

---

## 🎉 项目成果

### 已实现功能
1. ✅ 完整的RESTful API (9个端点)
2. ✅ MySQL数据持久化
3. ✅ FISCO BCOS区块链集成
4. ✅ 智能合约部署并运行
5. ✅ Excel批量导入
6. ✅ 隐私保护的哈希对账
7. ✅ 事件监听框架
8. ✅ 统计分析功能

### 技术亮点
- **隐私保护**: 链上仅存储哈希,链下存储加密数据
- **自动对账**: 智能合约自动实现哈希碰撞
- **完整性保证**: 区块链确保数据不可篡改
- **高可用性**: 4节点联盟链部署
- **可扩展性**: 模块化设计,易于扩展

---

## 👨‍💻 作者信息

**项目**: 基于联盟链的金融交易数据分布式对账与溯源系统
**类型**: 毕业设计
**完成时间**: 2025年1月
**技术栈**: FISCO BCOS, Go, MySQL, Solidity

---

**祝你毕业设计顺利完成!** 🎓🎊
