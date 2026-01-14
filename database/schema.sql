-- ========================================
-- 金融交易对账系统 - 数据库建表脚本
-- 数据库: bc_reconciliation
-- 字符集: UTF8MB4
-- ========================================

CREATE DATABASE IF NOT EXISTS `bc_reconciliation` DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
USE `bc_reconciliation`;

-- ========================================
-- 表1: 机构信息表 (institutions)
-- ========================================
DROP TABLE IF EXISTS `institutions`;
CREATE TABLE `institutions` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `institution_id` VARCHAR(64) NOT NULL COMMENT '机构唯一标识',
  `name` VARCHAR(128) NOT NULL COMMENT '机构名称',
  `address` VARCHAR(42) NOT NULL COMMENT '区块链地址',
  `status` TINYINT NOT NULL DEFAULT 1 COMMENT '状态: 0-禁用, 1-启用',
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_institution_id` (`institution_id`),
  UNIQUE KEY `uk_address` (`address`),
  KEY `idx_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='机构信息表';

-- ========================================
-- 表2: 交易流水主表 (transactions)
-- ========================================
DROP TABLE IF EXISTS `transactions`;
CREATE TABLE `transactions` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `biz_id` VARCHAR(64) NOT NULL COMMENT '业务流水号',
  `institution_id` VARCHAR(64) NOT NULL COMMENT '机构ID',
  `amount_cipher` VARCHAR(256) NOT NULL COMMENT '金额密文(AES加密)',
  `amount_hash` VARCHAR(64) NOT NULL COMMENT '金额哈希(用于链上验证)',
  `data_hash` VARCHAR(64) NOT NULL COMMENT '数据哈希(SHA256,上链用)',
  `salt` VARCHAR(64) NOT NULL COMMENT '随机盐',
  `receiver` VARCHAR(128) NOT NULL COMMENT '收款方',
  `sender` VARCHAR(128) NOT NULL COMMENT '付款方',
  `tx_type` TINYINT NOT NULL DEFAULT 1 COMMENT '交易类型: 1-转账, 2-退款, 3-其他',
  `status` TINYINT NOT NULL DEFAULT 0 COMMENT '状态: 0-待上链, 1-已上链, 2-对账成功, 3-对账失败',
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_biz_id` (`biz_id`),
  KEY `idx_institution_id` (`institution_id`),
  KEY `idx_status` (`status`),
  KEY `idx_created_at` (`created_at`),
  KEY `idx_data_hash` (`data_hash`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='交易流水主表';

-- ========================================
-- 表3: 链上锚定表 (chain_receipts)
-- ========================================
DROP TABLE IF EXISTS `chain_receipts`;
CREATE TABLE `chain_receipts` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `biz_id` VARCHAR(64) NOT NULL COMMENT '业务流水号',
  `tx_hash` VARCHAR(128) NOT NULL COMMENT '区块链交易哈希',
  `block_height` BIGINT NOT NULL COMMENT '区块高度',
  `block_hash` VARCHAR(128) NOT NULL COMMENT '区块哈希',
  `contract_address` VARCHAR(42) NOT NULL COMMENT '合约地址',
  `gas_used` BIGINT NOT NULL DEFAULT 0 COMMENT 'Gas消耗',
  `status` TINYINT NOT NULL DEFAULT 1 COMMENT '状态: 0-失败, 1-成功',
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_biz_id` (`biz_id`),
  KEY `idx_tx_hash` (`tx_hash`),
  KEY `idx_block_height` (`block_height`),
  KEY `idx_contract_address` (`contract_address`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='链上锚定表';

-- ========================================
-- 表4: 对账记录表 (reconciliations)
-- ========================================
DROP TABLE IF EXISTS `reconciliations`;
CREATE TABLE `reconciliations` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `biz_id` VARCHAR(64) NOT NULL COMMENT '业务流水号',
  `party_a` VARCHAR(64) NOT NULL COMMENT '机构A(发起方)',
  `party_b` VARCHAR(64) NOT NULL COMMENT '机构B(对手方)',
  `status` TINYINT NOT NULL COMMENT '对账状态: 2-成功, 3-失败',
  `matched_at` DATETIME DEFAULT NULL COMMENT '对账时间',
  `block_height` BIGINT DEFAULT NULL COMMENT '对账成功时的区块高度',
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_biz_id` (`biz_id`),
  KEY `idx_party_a` (`party_a`),
  KEY `idx_party_b` (`party_b`),
  KEY `idx_status` (`status`),
  KEY `idx_matched_at` (`matched_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='对账记录表';

-- ========================================
-- 表5: 事件监听日志表 (event_logs)
-- ========================================
DROP TABLE IF EXISTS `event_logs`;
CREATE TABLE `event_logs` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `event_type` VARCHAR(64) NOT NULL COMMENT '事件类型',
  `biz_id` VARCHAR(64) DEFAULT NULL COMMENT '业务流水号',
  `tx_hash` VARCHAR(128) NOT NULL COMMENT '交易哈希',
  `block_height` BIGINT NOT NULL COMMENT '区块高度',
  `contract_address` VARCHAR(42) NOT NULL COMMENT '合约地址',
  `data` TEXT COMMENT '事件数据(JSON)',
  `processed` TINYINT NOT NULL DEFAULT 0 COMMENT '是否已处理: 0-未处理, 1-已处理',
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  PRIMARY KEY (`id`),
  KEY `idx_event_type` (`event_type`),
  KEY `idx_biz_id` (`biz_id`),
  KEY `idx_tx_hash` (`tx_hash`),
  KEY `idx_block_height` (`block_height`),
  KEY `idx_processed` (`processed`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='事件监听日志表';

-- ========================================
-- 表6: 系统配置表 (system_configs)
-- ========================================
DROP TABLE IF EXISTS `system_configs`;
CREATE TABLE `system_configs` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `config_key` VARCHAR(64) NOT NULL COMMENT '配置键',
  `config_value` TEXT NOT NULL COMMENT '配置值',
  `description` VARCHAR(256) DEFAULT NULL COMMENT '配置说明',
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_config_key` (`config_key`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='系统配置表';

-- ========================================
-- 表7: 用户表 (users)
-- ========================================
DROP TABLE IF EXISTS `users`;
CREATE TABLE `users` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `username` VARCHAR(64) NOT NULL COMMENT '用户名',
  `password_hash` VARCHAR(128) NOT NULL COMMENT '密码哈希(bcrypt)',
  `institution_id` VARCHAR(64) NOT NULL COMMENT '所属机构ID',
  `role` VARCHAR(32) NOT NULL DEFAULT 'operator' COMMENT '角色: admin, operator, auditor',
  `email` VARCHAR(128) DEFAULT NULL COMMENT '邮箱',
  `status` TINYINT NOT NULL DEFAULT 1 COMMENT '状态: 0-禁用, 1-启用',
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_username` (`username`),
  KEY `idx_institution_id` (`institution_id`),
  KEY `idx_role` (`role`),
  KEY `idx_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户表';

-- ========================================
-- 初始化数据
-- ========================================

-- 插入系统配置
INSERT INTO `system_configs` (`config_key`, `config_value`, `description`) VALUES
('contract_address', '', '智能合约地址'),
('last_sync_block', '0', '事件监听最后同步的区块高度'),
('batch_upload_size', '100', '批量上传的最大数量'),
('aes_encryption_key', 'your-aes-encryption-key-32bytes!!', 'AES加密密钥(32字节)');

-- ========================================
-- 索引说明
-- ========================================
-- 1. 所有表都使用InnoDB引擎,支持事务
-- 2. 所有varchar字段使用utf8mb4字符集,支持emoji
-- 3. 关键查询字段已建立索引
-- 4. 业务流水号(biz_id)在各表中保持唯一,用于关联
-- 5. 时间字段统一使用DATETIME类型
-- ========================================
