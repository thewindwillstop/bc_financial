package models

import (
	"time"

)

// ChainReceipt 链上锚定表
type ChainReceipt struct {
	ID              uint      `json:"id" gorm:"primaryKey"`
	BizID           string    `json:"biz_id" gorm:"uniqueIndex;size:64;comment:业务流水号"`
	TxHash          string    `json:"tx_hash" gorm:"index;size:128;comment:区块链交易哈希"`
	BlockHeight     int64     `json:"block_height" gorm:"index;comment:区块高度"`
	BlockHash       string    `json:"block_hash" gorm:"size:128;comment:区块哈希"`
	ContractAddress string    `json:"contract_address" gorm:"index;size:42;comment:合约地址"`
	GasUsed         int64     `json:"gas_used" gorm:"default:0;comment:Gas消耗"`
	Status          int8      `json:"status" gorm:"default:1;comment:状态"`
	CreatedAt       time.Time `json:"created_at" gorm:"autoCreateTime"`
}

// TableName 指定表名
func (ChainReceipt) TableName() string {
	return "chain_receipts"
}

// ChainReceiptStatus 链上回执状态常量
const (
	ChainReceiptStatusFailed  int8 = 0 // 失败
	ChainReceiptStatusSuccess int8 = 1 // 成功
)

// ToResponse 转换为响应格式
func (c *ChainReceipt) ToResponse() map[string]interface{} {
	return map[string]interface{}{
		"id":               c.ID,
		"biz_id":           c.BizID,
		"tx_hash":          c.TxHash,
		"block_height":     c.BlockHeight,
		"block_hash":       c.BlockHash,
		"contract_address": c.ContractAddress,
		"gas_used":         c.GasUsed,
		"status":           c.Status,
		"created_at":       c.CreatedAt,
	}
}
