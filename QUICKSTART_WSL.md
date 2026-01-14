# 🚀 区块链对账系统 - 快速启动指南 (无 Docker 环境)

## 📋 当前状态

✅ **已完成**:
- FISCO BCOS 4节点已部署
- 后端 Go 代码已完成
- 智能合约 (Solidity) 已编写
- 配置文件已准备

❌ **环境限制**:
- Docker 未安装 (WSL2 环境)
- MySQL 未启动

---

## 🎯 快速启动方案 (使用 FISCO BCOS)

此方案可以在**没有 Docker**的情况下运行系统，适合:
- 时间紧迫的毕业设计
- 本地开发测试
- 快速演示

---

## 📝 步骤 1: 安装并启动 MySQL (10分钟)

### 1.1 安装 MySQL

```bash
sudo apt update
sudo apt install mysql-server -y

# 启动服务
sudo service mysql start

# 设置 root 密码
sudo mysql
```

在 MySQL 命令行中执行:

```sql
ALTER USER 'root'@'localhost' IDENTIFIED WITH mysql_native_password BY 'lin123456';
FLUSH PRIVILEGES;
EXIT;
```

### 1.2 创建数据库和表

```bash
cd /home/lin123456/colloge_project/bc_financial

# 创建数据库
mysql -u root -plin123456 -e "CREATE DATABASE IF NOT EXISTS bc_reconciliation;"

# 导入表结构
mysql -u root -plin123456 bc_reconciliation < database/schema.sql

# 验证
mysql -u root -plin123456 bc_reconciliation -e "SHOW TABLES;"
```

预期输出:
```
+--------------------------+
| Tables_in_bc_reconciliation |
+--------------------------+
| institutions             |
| transactions             |
| reconciliations          |
| reconciliation_logs      |
+--------------------------+
```

---

## 📝 步骤 2: 部署 FISCO BCOS 智能合约 (15分钟)

### 2.1 启动 FISCO 节点

```bash
cd /home/lin123456/colloge_project/bc_financial/fisco/nodes/127.0.0.1

# 启动所有节点
./start_all.sh

# 等待几秒
sleep 5

# 检查节点状态
ps -ef | grep fisco-bcos
```

预期输出应该看到 4 个 fisco-bcos 进程。

### 2.2 部署智能合约

```bash
cd /home/lin123456/colloge_project/bc_financial/contracts

# 复制合约到控制台目录
cp Reconciliation.sol ../fisco/nodes/127.0.0.1/console/solidity/

cd ../fisco/nodes/127.0.0.1/console

# 启动控制台
./start.sh
```

在控制台中执行:

```bash
# 编译合约
compile Reconciliation.sol

# 部署合约
deploy Reconciliation.sol

# 退出控制台
exit
```

**重要**: 记录返回的合约地址，类似:
```
Reconciliation address: 0x1234...abcd
```

### 2.3 创建配置文件

```bash
cd /home/lin123456/colloge_project/bc_financial/backend/configs

# 创建 FISCO 配置
cat > fisco_config.toml << 'EOF'
[Network]
Peers=["127.0.0.1:20200", "127.0.0.1:20201"]

[Account]
KeyFile = "./configs/sdk.key"

[Chain]
ChainID = 1

[Security]
Enable = false
EOF
```

### 2.4 更新主配置文件

编辑 `configs/config.yaml`:

```bash
vim configs/config.yaml
```

修改为:

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
    max_open_conns: 100
    max_idle_conns: 10
    max_lifetime: 3600

# 区块链配置 - 使用 FISCO BCOS
blockchain:
  type: fisco  # blockchain type: "fisco" or "fabric"
  config_file: ./configs/fisco_config.toml
  contract_address: "0x..."  # 填入步骤2.2中的合约地址

log:
  level: info
  filename: logs/app.log
  max_size: 100
  max_backups: 3
  max_age: 30
  compress: true
```

---

## 📝 步骤 3: 启动后端服务 (5分钟)

### 3.1 安装 Go 依赖

```bash
cd /home/lin123456/colloge_project/bc_financial/backend

# 下载依赖
go mod download

# 编译检查
go build -o api cmd/api/main.go
```

### 3.2 启动服务

```bash
# 方式 1: 直接运行
go run cmd/api/main.go

# 方式 2: 编译后运行
./api
```

预期输出:
```
[INFO] Connecting to MySQL database...
[INFO] Database connection established
[INFO] Connecting to FISCO BCOS...
[INFO] Connected to FISCO BCOS (block number: X)
[INFO] Smart contract loaded successfully
[INFO] Server started on :8080
```

### 3.3 测试健康检查

打开**新的终端窗口**:

```bash
curl http://localhost:8080/health
```

预期输出:
```json
{
  "status": "healthy",
  "blockchain": "fisco",
  "database": "connected",
  "contract_loaded": true
}
```

---

## 📝 步骤 4: 测试核心功能 (20分钟)

### 4.1 创建测试机构

```bash
curl -X POST http://localhost:8080/api/v1/institutions \
  -H "Content-Type: application/json" \
  -d '{
    "name": "机构A",
    "code": "INST_A",
    "msp_id": "Org1MSP"
  }'
```

### 4.2 创建交易

```bash
curl -X POST http://localhost:8080/api/v1/transactions \
  -H "Content-Type: application/json" \
  -d '{
    "biz_id": "TX202501140001",
    "institution_id": "1",
    "amount": "1000.00",
    "sender": "机构A",
    "receiver": "机构B",
    "transaction_time": "2025-01-14T10:00:00Z"
  }'
```

### 4.3 查询交易

```bash
curl http://localhost:8080/api/v1/transactions/TX202501140001
```

### 4.4 查询统计

```bash
curl http://localhost:8080/api/v1/dashboard/statistics
```

### 4.5 测试批量上链

首先创建多条交易 (在数据库中):

```bash
mysql -u root -plin123456 bc_reconciliation << 'EOF'
INSERT INTO transactions (biz_id, institution_id, amount, sender, receiver, transaction_time, status) VALUES
('TX202501140002', 1, 2000.00, '机构A', '机构B', '2025-01-14 11:00:00', 0),
('TX202501140003', 1, 3000.00, '机构A', '机构C', '2025-01-14 12:00:00', 0),
('TX202501140004', 1, 1500.00, '机构A', '机构B', '2025-01-14 13:00:00', 0);
EOF
```

然后批量上链:

```bash
curl -X POST http://localhost:8080/api/v1/transactions/upload-chain
```

---

## 🐛 常见问题排查

### 问题 1: MySQL 连接失败

**错误**: `Error 2002: Can't connect to local MySQL server`

**解决**:
```bash
sudo service mysql start
sudo service mysql status
```

### 问题 2: FISCO 节点未启动

**错误**: `failed to connect to FISCO BCOS`

**解决**:
```bash
cd /home/lin123456/colloge_project/bc_financial/fisco/nodes/127.0.0.1
./start_all.sh
sleep 3
ps -ef | grep fisco-bcos
```

### 问题 3: 智能合约未部署

**错误**: `contract helper not initialized`

**解决**:
1. 检查 `config.yaml` 中的 `contract_address` 是否已填写
2. 确保合约已在控制台中成功部署
3. 重启后端服务

### 问题 4: 编译错误

**错误**: `cannot find package`

**解决**:
```bash
cd /home/lin123456/colloge_project/bc_financial/backend
go mod tidy
go mod download
```

### 问题 5: 端口被占用

**错误**: `bind: address already in use`

**解决**:
```bash
# 查看占用 8080 端口的进程
sudo lsof -i :8080

# 杀死进程
sudo kill -9 <PID>

# 或修改配置文件使用其他端口
vim configs/config.yaml
# 将 port: 8080 改为 port: 8081
```

---

## 📊 功能测试清单

完成上述步骤后,使用此清单验证系统:

### 基础功能

- [ ] MySQL 服务运行正常
- [ ] FISCO BCOS 节点运行正常 (4个节点)
- [ ] 智能合约已部署
- [ ] 后端服务启动成功
- [ ] 健康检查 API 返回正常

### 核心业务

- [ ] 创建机构成功
- [ ] 创建交易成功 (保存到数据库)
- [ ] 单笔交易上链成功
- [ ] 批量交易上链成功
- [ ] 查询交易信息正常
- [ ] 统计信息准确

### 数据验证

```bash
# 验证数据库中的交易
mysql -u root -plin123456 bc_reconciliation -e "
  SELECT biz_id, amount, tx_hash, status
  FROM transactions
  LIMIT 5;
"

# 在 FISCO 控制台查询链上数据
cd /home/lin123456/colloge_project/bc_financial/fisco/nodes/127.0.0.1/console
./start.sh

# 在控制台中
call Reconciliation.sol 0x... getTransaction "0x544832303235303131343030303000000000000000000000000000000000000000"
exit
```

---

## 🎯 下一步

### 如果所有测试通过:

1. **开发前端界面** (可选)
   - 使用 Vue.js 或 React
   - 主要页面: 交易列表、上传界面、统计面板

2. **完善功能**
   - 添加用户认证
   - 添加日志查询
   - 添加事件监听

3. **准备答辩**
   - 编写使用文档
   - 准备演示脚本
   - 制作 PPT

### 如果遇到问题:

1. **检查日志**
```bash
tail -f logs/app.log
```

2. **查看数据库**
```bash
mysql -u root -plin123456 bc_reconciliation
SELECT * FROM transactions;
```

3. **查看 FISCO 日志**
```bash
tail -f /home/lin123456/colloge_project/bc_financial/fisco/nodes/127.0.0.1/node0/log/log*.log
```

---

## 📞 需要帮助?

如果遇到无法解决的问题:

1. **查看完整日志**
   ```bash
   cd /home/lin123456/colloge_project/bc_financial/backend
   go run cmd/api/main.go 2>&1 | tee startup.log
   ```

2. **检查环境信息**
   ```bash
   go version
   mysql --version
   ps -ef | grep fisco-bcos
   ```

3. **查看配置文件**
   ```bash
   cat configs/config.yaml
   cat configs/fisco_config.toml
   ```

---

## ✅ 成功标准

系统正常运行的标准:

1. ✅ MySQL 数据库包含 4 张表
2. ✅ FISCO BCOS 4个节点都在运行
3. ✅ 后端服务监听在 8080 端口
4. ✅ 健康检查 API 返回 `"status": "healthy"`
5. ✅ 能够成功创建交易并上链
6. ✅ 能够查询统计信息

达到这些标准后,您的系统就具备了:
- ✅ 完整的区块链对账功能
- ✅ 可演示的核心流程
- ✅ 毕业设计的最低要求

祝您顺利! 🎉
