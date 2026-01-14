package models

import (
	"time"

)

// Reconciliation 对账记录表
type Reconciliation struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	BizID       string    `json:"biz_id" gorm:"uniqueIndex;size:64;comment:业务流水号"`
	PartyA      string    `json:"party_a" gorm:"index;size:64;comment:机构A"`
	PartyB      string    `json:"party_b" gorm:"index;size:64;comment:机构B"`
	Status      int8      `json:"status" gorm:"index;comment:对账状态"`
	MatchedAt   *time.Time `json:"matched_at,omitempty" gorm:"comment:对账时间"`
	BlockHeight *int64     `json:"block_height,omitempty" gorm:"comment:区块高度"`
	CreatedAt   time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
}

// TableName 指定表名
func (Reconciliation) TableName() string {
	return "reconciliations"
}

// ReconciliationStatus 对账状态常量
const (
	ReconciliationStatusMatched  int8 = 2 // 对账成功
	ReconciliationStatusMismatch int8 = 3 // 对账失败
)

// GetStatusText 获取状态文本
func (r *Reconciliation) GetStatusText() string {
	switch r.Status {
	case ReconciliationStatusMatched:
		return "对账成功"
	case ReconciliationStatusMismatch:
		return "对账失败"
	default:
		return "未知"
	}
}

// ReconciliationResponse 对账响应
type ReconciliationResponse struct {
	ID          uint       `json:"id"`
	BizID       string     `json:"biz_id"`
	PartyA      string     `json:"party_a"`
	PartyB      string     `json:"party_b"`
	Status      int8       `json:"status"`
	StatusText  string     `json:"status_text"`
	MatchedAt   *time.Time `json:"matched_at,omitempty"`
	BlockHeight *int64     `json:"block_height,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
}

// ToResponse 转换为响应格式
func (r *Reconciliation) ToResponse() *ReconciliationResponse {
	return &ReconciliationResponse{
		ID:          r.ID,
		BizID:       r.BizID,
		PartyA:      r.PartyA,
		PartyB:      r.PartyB,
		Status:      r.Status,
		StatusText:  r.GetStatusText(),
		MatchedAt:   r.MatchedAt,
		BlockHeight: r.BlockHeight,
		CreatedAt:   r.CreatedAt,
	}
}

// StatisticsResponse 统计数据响应
type StatisticsResponse struct {
	TotalTransactions int64   `json:"total_transactions"`
	MatchedCount      int64   `json:"matched_count"`
	MismatchCount     int64   `json:"mismatch_count"`
	MatchRate         float64 `json:"match_rate"` // 匹配率(百分比)
	PendingCount      int64   `json:"pending_count"`
	UploadedCount     int64   `json:"uploaded_count"`
}

// DailyStatistics 每日统计
type DailyStatistics struct {
	Date     string `json:"date"`
	Count    int64  `json:"count"`
	Matched  int64  `json:"matched"`
	MatchRate float64 `json:"match_rate"`
}
