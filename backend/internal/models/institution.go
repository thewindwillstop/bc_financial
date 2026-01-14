package models

import (
	"time"

)

// Institution 机构信息表
type Institution struct {
	ID             uint      `json:"id" gorm:"primaryKey"`
	InstitutionID  string    `json:"institution_id" gorm:"uniqueIndex;size:64;comment:机构唯一标识"`
	Name           string    `json:"name" gorm:"size:128;comment:机构名称"`
	Address        string    `json:"address" gorm:"uniqueIndex;size:42;comment:区块链地址"`
	Status         int8      `json:"status" gorm:"index;default:1;comment:状态"`
	CreatedAt      time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt      time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// TableName 指定表名
func (Institution) TableName() string {
	return "institutions"
}

// InstitutionStatus 机构状态常量
const (
	InstitutionStatusDisabled int8 = 0 // 禁用
	InstitutionStatusEnabled  int8 = 1 // 启用
)

// Transactions 关联的交易
func (i *Institution) Transactions() []Transaction {
	var txs []Transaction
	// 这里需要在Service层实现,或者使用gorm的Preload
	return txs
}

// CreateInstitutionRequest 创建机构请求
type CreateInstitutionRequest struct {
	InstitutionID string `json:"institution_id" binding:"required"`
	Name          string `json:"name" binding:"required"`
	Address       string `json:"address" binding:"required"`
}

// UpdateInstitutionRequest 更新机构请求
type UpdateInstitutionRequest struct {
	Name   string `json:"name" binding:"required"`
	Status int8   `json:"status" binding:"required,oneof=0 1"`
}

// InstitutionResponse 机构响应
type InstitutionResponse struct {
	ID            uint      `json:"id"`
	InstitutionID string    `json:"institution_id"`
	Name          string    `json:"name"`
	Address       string    `json:"address"`
	Status        int8      `json:"status"`
	StatusText    string    `json:"status_text"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// GetStatusText 获取状态文本
func (i *Institution) GetStatusText() string {
	switch i.Status {
	case InstitutionStatusDisabled:
		return "禁用"
	case InstitutionStatusEnabled:
		return "启用"
	default:
		return "未知"
	}
}

// ToResponse 转换为响应格式
func (i *Institution) ToResponse() *InstitutionResponse {
	return &InstitutionResponse{
		ID:            i.ID,
		InstitutionID: i.InstitutionID,
		Name:          i.Name,
		Address:       i.Address,
		Status:        i.Status,
		StatusText:    i.GetStatusText(),
		CreatedAt:     i.CreatedAt,
		UpdatedAt:     i.UpdatedAt,
	}
}
