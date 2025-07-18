package models

import (
	"time"

	"gorm.io/gorm"
)

// LoginAttempt 登录尝试记录模型
type LoginAttempt struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	UserID        string    `gorm:"size:36;index" json:"user_id,omitempty"`        // 用户ID（可选，用于用户名/邮箱登录）
	IPAddress     string    `gorm:"size:45;not null;index" json:"ip_address"`      // IP地址
	UserAgent     string    `gorm:"size:500" json:"user_agent,omitempty"`          // 用户代理
	LoginIdentity string    `gorm:"size:255;not null;index" json:"login_identity"` // 登录标识（用户名/邮箱）
	AttemptType   string    `gorm:"size:20;not null" json:"attempt_type"`          // 尝试类型：failed, success, blocked
	FailureReason string    `gorm:"size:100" json:"failure_reason,omitempty"`      // 失败原因
	RequestPath   string    `gorm:"size:255" json:"request_path,omitempty"`        // 请求路径
	Country       string    `gorm:"size:50" json:"country,omitempty"`              // 国家
	City          string    `gorm:"size:100" json:"city,omitempty"`                // 城市
	CreatedAt     time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt     time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

// TableName 指定表名
func (LoginAttempt) TableName() string {
	return "login_attempts"
}

// BeforeCreate 创建前钩子
func (la *LoginAttempt) BeforeCreate(tx *gorm.DB) error {
	la.CreatedAt = time.Now()
	la.UpdatedAt = time.Now()
	return nil
}

// BeforeUpdate 更新前钩子
func (la *LoginAttempt) BeforeUpdate(tx *gorm.DB) error {
	la.UpdatedAt = time.Now()
	return nil
}

// IsFailedAttempt 检查是否为失败尝试
func (la *LoginAttempt) IsFailedAttempt() bool {
	return la.AttemptType == AttemptTypeFailed
}

// IsSuccessAttempt 检查是否为成功尝试
func (la *LoginAttempt) IsSuccessAttempt() bool {
	return la.AttemptType == AttemptTypeSuccess
}

// IsBlockedAttempt 检查是否为被阻止尝试
func (la *LoginAttempt) IsBlockedAttempt() bool {
	return la.AttemptType == AttemptTypeBlocked
}

// 登录尝试类型常量
const (
	AttemptTypeFailed  = "failed"
	AttemptTypeSuccess = "success"
	AttemptTypeBlocked = "blocked"
)

// 失败原因常量
const (
	FailureReasonInvalidCredentials = "invalid_credentials"
	FailureReasonUserNotFound       = "user_not_found"
	FailureReasonAccountLocked      = "account_locked"
	FailureReasonAccountDisabled    = "account_disabled"
	FailureReasonTooManyAttempts    = "too_many_attempts"
	FailureReasonIPBlocked          = "ip_blocked"
	FailureReasonInvalidToken       = "invalid_token"
)
