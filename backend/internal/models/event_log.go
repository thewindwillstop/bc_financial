package models

import (
	"database/sql/driver"
	"encoding/json"
	"time"

)

// EventLog 事件监听日志表
type EventLog struct {
	ID              uint                  `json:"id" gorm:"primaryKey"`
	EventType       string                `json:"event_type" gorm:"index;size:64;comment:事件类型"`
	BizID           string                `json:"biz_id,omitempty" gorm:"index;size:64;comment:业务流水号"`
	TxHash          string                `json:"tx_hash" gorm:"index;size:128;comment:交易哈希"`
	BlockHeight     int64                 `json:"block_height" gorm:"index;comment:区块高度"`
	ContractAddress string                `json:"contract_address" gorm:"index;size:42;comment:合约地址"`
	Data            EventData             `json:"data" gorm:"type:json;comment:事件数据"`
	Processed       int8                  `json:"processed" gorm:"index;default:0;comment:是否已处理"`
	CreatedAt       time.Time             `json:"created_at" gorm:"autoCreateTime"`
}

// TableName 指定表名
func (EventLog) TableName() string {
	return "event_logs"
}

// EventProcessed 事件处理状态常量
const (
	EventNotProcessed int8 = 0 // 未处理
	EventProcessed    int8 = 1 // 已处理
)

// EventData 事件数据(JSON格式)
type EventData map[string]interface{}

// Scan 实现sql.Scanner接口
func (e *EventData) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, e)
}

// Value 实现driver.Valuer接口
func (e EventData) Value() (driver.Value, error) {
	return json.Marshal(e)
}

// EventType 事件类型常量
const (
	EventTypeDataUploaded       = "DataUploaded"
	EventTypeReconciliationEvent = "ReconciliationEvent"
	EventTypeInstitutionRegistered = "InstitutionRegistered"
)

// User 用户表
type User struct {
	ID             uint      `json:"id" gorm:"primaryKey"`
	Username       string    `json:"username" gorm:"uniqueIndex;size:64;comment:用户名"`
	PasswordHash   string    `json:"-" gorm:"size:128;comment:密码哈希"` // 不暴露给前端
	InstitutionID  string    `json:"institution_id" gorm:"index;size:64;comment:所属机构ID"`
	Role           string    `json:"role" gorm:"index;size:32;default:operator;comment:角色"`
	Email          string    `json:"email" gorm:"size:128;comment:邮箱"`
	Status         int8      `json:"status" gorm:"index;default:1;comment:状态"`
	CreatedAt      time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt      time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// TableName 指定表名
func (User) TableName() string {
	return "users"
}

// UserStatus 用户状态常量
const (
	UserStatusDisabled int8 = 0 // 禁用
	UserStatusEnabled  int8 = 1 // 启用
)

// UserRole 用户角色常量
const (
	UserRoleAdmin    = "admin"    // 系统管理员
	UserRoleOperator = "operator" // 机构操作员
	UserRoleAuditor  = "auditor"  // 审计员
)

// LoginRequest 登录请求
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// RegisterRequest 注册请求
type RegisterRequest struct {
	Username      string `json:"username" binding:"required"`
	Password      string `json:"password" binding:"required,min=6"`
	InstitutionID string `json:"institution_id" binding:"required"`
	Email         string `json:"email" binding:"required,email"`
}

// LoginResponse 登录响应
type LoginResponse struct {
	Token          string `json:"token"`
	Username       string `json:"username"`
	InstitutionID  string `json:"institution_id"`
	Role           string `json:"role"`
	InstitutionName string `json:"institution_name,omitempty"`
}

// UserResponse 用户响应
type UserResponse struct {
	ID             uint      `json:"id"`
	Username       string    `json:"username"`
	InstitutionID  string    `json:"institution_id"`
	Role           string    `json:"role"`
	Email          string    `json:"email"`
	Status         int8      `json:"status"`
	StatusText     string    `json:"status_text"`
	CreatedAt      time.Time `json:"created_at"`
}

// GetStatusText 获取状态文本
func (u *User) GetStatusText() string {
	switch u.Status {
	case UserStatusDisabled:
		return "禁用"
	case UserStatusEnabled:
		return "启用"
	default:
		return "未知"
	}
}

// ToResponse 转换为响应格式
func (u *User) ToResponse() *UserResponse {
	return &UserResponse{
		ID:            u.ID,
		Username:      u.Username,
		InstitutionID: u.InstitutionID,
		Role:          u.Role,
		Email:         u.Email,
		Status:        u.Status,
		StatusText:    u.GetStatusText(),
		CreatedAt:     u.CreatedAt,
	}
}
