# 数据库设计文档

## 📊 数据库概述

**数据库名称**: `bc_reconciliation`
**字符集**: UTF8MB4
**存储引擎**: InnoDB
**用途**: 金融交易分布式对账系统的业务数据存储

---

## 📋 数据表清单

### 1. institutions (机构信息表)
存储金融机构的基本信息

| 字段 | 类型 | 说明 |
|------|------|------|
| id | BIGINT | 主键ID |
| institution_id | VARCHAR(64) | 机构唯一标识 |
| name | VARCHAR(128) | 机构名称 |
| address | VARCHAR(42) | 区块链地址 |
| status | TINYINT | 状态: 0-禁用, 1-启用 |
| created_at | DATETIME | 创建时间 |
| updated_at | DATETIME | 更新时间 |

**索引**:
- PRIMARY KEY (id)
- UNIQUE KEY (institution_id)
- UNIQUE KEY (address)

---

### 2. transactions (交易流水主表) ⭐核心表
存储交易明文数据(加密)和哈希值

| 字段 | 类型 | 说明 |
|------|------|------|
| id | BIGINT | 主键ID |
| biz_id | VARCHAR(64) | 业务流水号(唯一) |
| institution_id | VARCHAR(64) | 机构ID |
| amount_cipher | VARCHAR(256) | 金额密文(AES加密) |
| amount_hash | VARCHAR(64) | 金额哈希 |
| data_hash | VARCHAR(64) | 数据哈希(上链用) |
| salt | VARCHAR(64) | 随机盐 |
| receiver | VARCHAR(128) | 收款方 |
| sender | VARCHAR(128) | 付款方 |
| tx_type | TINYINT | 交易类型: 1-转账, 2-退款 |
| status | TINYINT | 状态: 0-待上链, 1-已上链, 2-对账成功, 3-对账失败 |
| created_at | DATETIME | 创建时间 |
| updated_at | DATETIME | 更新时间 |

**索引**:
- PRIMARY KEY (id)
- UNIQUE KEY (biz_id)
- INDEX (institution_id)
- INDEX (status)
- INDEX (data_hash)

**状态流转**:
```
0(待上链) → 1(已上链) → 2(对账成功) / 3(对账失败)
```

---

### 3. chain_receipts (链上锚定表)
存储交易上链后的回执信息

| 字段 | 类型 | 说明 |
|------|------|------|
| id | BIGINT | 主键ID |
| biz_id | VARCHAR(64) | 业务流水号 |
| tx_hash | VARCHAR(128) | 区块链交易哈希 |
| block_height | BIGINT | 区块高度 |
| block_hash | VARCHAR(128) | 区块哈希 |
| contract_address | VARCHAR(42) | 合约地址 |
| gas_used | BIGINT | Gas消耗 |
| status | TINYINT | 状态: 0-失败, 1-成功 |
| created_at | DATETIME | 创建时间 |

**用途**: 前端"点击验证"功能,通过tx_hash跳转到区块链浏览器

---

### 4. reconciliations (对账记录表)
记录对账成功/失败的记录

| 字段 | 类型 | 说明 |
|------|------|------|
| id | BIGINT | 主键ID |
| biz_id | VARCHAR(64) | 业务流水号 |
| party_a | VARCHAR(64) | 机构A(发起方) |
| party_b | VARCHAR(64) | 机构B(对手方) |
| status | TINYINT | 对账状态: 2-成功, 3-失败 |
| matched_at | DATETIME | 对账时间 |
| block_height | BIGINT | 对账成功时的区块高度 |
| created_at | DATETIME | 创建时间 |
| updated_at | DATETIME | 更新时间 |

**用途**: 统计对账成功率,查询对账历史

---

### 5. event_logs (事件监听日志表)
存储链上事件监听日志

| 字段 | 类型 | 说明 |
|------|------|------|
| id | BIGINT | 主键ID |
| event_type | VARCHAR(64) | 事件类型 |
| biz_id | VARCHAR(64) | 业务流水号 |
| tx_hash | VARCHAR(128) | 交易哈希 |
| block_height | BIGINT | 区块高度 |
| contract_address | VARCHAR(42) | 合约地址 |
| data | TEXT | 事件数据(JSON) |
| processed | TINYINT | 是否已处理: 0-未处理, 1-已处理 |
| created_at | DATETIME | 创建时间 |

**事件类型**:
- `DataUploaded`: 数据上链事件
- `ReconciliationEvent`: 对账完成事件

**用途**: 断点续传机制,记录同步到哪个区块

---

### 6. system_configs (系统配置表)
存储系统级配置

| 字段 | 类型 | 说明 |
|------|------|------|
| id | BIGINT | 主键ID |
| config_key | VARCHAR(64) | 配置键 |
| config_value | TEXT | 配置值 |
| description | VARCHAR(256) | 配置说明 |
| created_at | DATETIME | 创建时间 |
| updated_at | DATETIME | 更新时间 |

**预置配置**:
- `contract_address`: 智能合约地址
- `last_sync_block`: 事件监听最后同步的区块高度
- `batch_upload_size`: 批量上传的最大数量
- `aes_encryption_key`: AES加密密钥

---

### 7. users (用户表)
存储系统用户信息

| 字段 | 类型 | 说明 |
|------|------|------|
| id | BIGINT | 主键ID |
| username | VARCHAR(64) | 用户名(唯一) |
| password_hash | VARCHAR(128) | 密码哈希(bcrypt) |
| institution_id | VARCHAR(64) | 所属机构ID |
| role | VARCHAR(32) | 角色: admin, operator, auditor |
| email | VARCHAR(128) | 邮箱 |
| status | TINYINT | 状态: 0-禁用, 1-启用 |
| created_at | DATETIME | 创建时间 |
| updated_at | DATETIME | 更新时间 |

**角色说明**:
- `admin`: 系统管理员,管理机构和用户
- `operator`: 机构操作员,录入流水和查看报表
- `auditor`: 审计员,查看全网数据

---

## 🔄 数据流转示意

```
1. 用户上传Excel
   ↓
2. Go后端解析文件
   ├─ 计算哈希: data_hash = SHA256(biz_id + amount + salt)
   ├─ AES加密: amount_cipher = AES.encrypt(amount)
   ↓
3. 写入transactions表 (status=0待上链)
   ↓
4. 调用智能合约 uploadTransaction(biz_id, data_hash)
   ↓
5. 合约返回成功
   ↓
6. 写入chain_receipts表
   ↓
7. 更新transactions表 (status=1已上链)
   ↓
8. 事件监听服务检测到对账事件
   ↓
9. 更新transactions表 (status=2对账成功 或 3对账失败)
   ↓
10. 写入reconciliations表
```

---

## 🔒 隐私保护机制

### 1. 链上存储
- **仅存储哈希**: `data_hash = SHA256(流水号 + 金额 + 随机盐)`
- **不存储明文**: 金额、流水号明文仅存于链下数据库

### 2. 链下存储
- **金额加密**: `amount_cipher = AES.encrypt(amount, key)`
- **盐值保密**: 每笔交易独立随机盐,防止彩虹表攻击

### 3. 哈希碰撞对账
```
机构A和机构B上传相同biz_id的哈希:
- Hash相同 → 对账成功 (说明金额一致)
- Hash不同 → 对账失败 (说明有人篡改)
```

---

## 📊 统计查询示例

### 1. 对账成功率
```sql
SELECT
  COUNT(*) as total,
  SUM(CASE WHEN status = 2 THEN 1 ELSE 0 END) as matched,
  SUM(CASE WHEN status = 2 THEN 1 ELSE 0 END) / COUNT(*) * 100 as rate
FROM transactions;
```

### 2. 各机构交易统计
```sql
SELECT
  i.name,
  COUNT(t.id) as total_tx,
  SUM(CASE WHEN t.status = 2 THEN 1 ELSE 0 END) as matched_tx
FROM institutions i
LEFT JOIN transactions t ON i.institution_id = t.institution_id
GROUP BY i.institution_id;
```

### 3. 每日对账趋势
```sql
SELECT
  DATE(created_at) as date,
  COUNT(*) as count,
  SUM(CASE WHEN status = 2 THEN 1 ELSE 0 END) as matched
FROM reconciliations
WHERE created_at >= DATE_SUB(NOW(), INTERVAL 30 DAY)
GROUP BY DATE(created_at)
ORDER BY date DESC;
```

---

## 🚀 初始化数据库

```bash
# 创建数据库和表
mysql -u root -p < database/schema.sql

# 验证
mysql -u root -p bc_reconciliation -e "SHOW TABLES;"
```

---

**作者**: 毕业设计项目组
**版本**: v1.0.0
**最后更新**: 2026-01-13
