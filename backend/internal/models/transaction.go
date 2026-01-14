package models

import (
	"time"

	"gorm.io/gorm"
)

// PageResponse 分页响应结构
type PageResponse struct {
	Total int64       `json:"total"`
	Page  int         `json:"page"`
	Size  int         `json:"size"`
	Data  interface{} `json:"data"`
}

// Transaction 交易流水主表
type Transaction struct {
	ID             uint      `json:"id" gorm:"primaryKey"`
	BizID          string    `json:"biz_id" gorm:"uniqueIndex;size:64;comment:业务流水号"`
	InstitutionID  string    `json:"institution_id" gorm:"index;size:64;comment:机构ID"`
	AmountCipher   string    `json:"amount_cipher" gorm:"size:256;comment:金额密文"`
	AmountHash     string    `json:"amount_hash" gorm:"size:64;comment:金额哈希"`
	DataHash       string    `json:"data_hash" gorm:"index;size:64;comment:数据哈希"`
	Salt           string    `json:"-" gorm:"size:64;comment:随机盐"` // 不暴露给前端
	Receiver       string    `json:"receiver" gorm:"size:128;comment:收款方"`
	Sender         string    `json:"sender" gorm:"size:128;comment:付款方"`
	TxType         int8      `json:"tx_type" gorm:"default:1;comment:交易类型"`
	Status         int8      `json:"status" gorm:"index;default:0;comment:状态"`
	CreatedAt      time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt      time.Time `json:"updated_at" gorm:"autoUpdateTime"`

	// 关联
	ChainReceipt  *ChainReceipt `json:"chain_receipt,omitempty" gorm:"foreignKey:BizID;references:BizID"`
	Reconciliation *Reconciliation `json:"reconciliation,omitempty" gorm:"foreignKey:BizID;references:BizID"`
}

// TableName 指定表名
func (Transaction) TableName() string {
	return "transactions"
}

// TransactionStatus 交易状态常量
const (
	TxStatusPending    int8 = 0 // 待上链
	TxStatusUploaded   int8 = 1 // 已上链
	TxStatusMatched    int8 = 2 // 对账成功
	TxStatusMismatch   int8 = 3 // 对账失败
)

// BeforeCreate 创建前钩子
func (t *Transaction) BeforeCreate(tx *gorm.DB) error {
	// 可以在这里添加创建前的逻辑
	return nil
}

// CreateTransactionRequest 创建交易请求
type CreateTransactionRequest struct {
	BizID         string `json:"biz_id" binding:"required"`
	InstitutionID string `json:"institution_id" binding:"required"`
	Amount        string `json:"amount" binding:"required"` // 明文金额,后端加密
	Receiver      string `json:"receiver" binding:"required"`
	Sender        string `json:"sender" binding:"required"`
	TxType        int8   `json:"tx_type"`
}

// UploadChainRequest 上链请求
type UploadChainRequest struct {
	BizIDs []string `json:"biz_ids" binding:"required"`
}

// TransactionResponse 交易响应
type TransactionResponse struct {
	ID            uint      `json:"id"`
	BizID         string    `json:"biz_id"`
	InstitutionID string    `json:"institution_id"`
	Receiver      string    `json:"receiver"`
	Sender        string    `json:"sender"`
	TxType        int8      `json:"tx_type"`
	Status        int8      `json:"status"`
	StatusText    string    `json:"status_text"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
	// 不返回敏感信息
}

// GetStatusText 获取状态文本
func (t *Transaction) GetStatusText() string {
	switch t.Status {
	case TxStatusPending:
		return "待上链"
	case TxStatusUploaded:
		return "已上链"
	case TxStatusMatched:
		return "对账成功"
	case TxStatusMismatch:
		return "对账失败"
	default:
		return "未知"
	}
}

// ToResponse 转换为响应格式
func (t *Transaction) ToResponse() *TransactionResponse {
	return &TransactionResponse{
		ID:            t.ID,
		BizID:         t.BizID,
		InstitutionID: t.InstitutionID,
		Receiver:      t.Receiver,
		Sender:        t.Sender,
		TxType:        t.TxType,
		Status:        t.Status,
		StatusText:    t.GetStatusText(),
		CreatedAt:     t.CreatedAt,
		UpdatedAt:     t.UpdatedAt,
	}
}

// BatchUploadResult 批量上传结果
type BatchUploadResult struct {
	Total      int      `json:"total"`
	Success    int      `json:"success"`
	Failed     int      `json:"failed"`
	SuccessIDs []string `json:"success_ids"`
	FailedIDs  []string `json:"failed_ids"`
}
