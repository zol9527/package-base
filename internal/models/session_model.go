package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Session 会话模型
//
// 描述: 用于管理用户会话信息，支持多设备登录和JWT Token管理
type Session struct {
	ID           string    `gorm:"primaryKey;size:36" json:"id"`
	UserID       string    `gorm:"not null;index;size:36" json:"user_id"`
	DeviceInfo   string    `gorm:"size:500" json:"device_info"`
	IPAddress    string    `gorm:"size:45" json:"ip_address"`
	UserAgent    string    `gorm:"size:500" json:"user_agent"`
	RefreshToken string    `gorm:"size:255;uniqueIndex" json:"-"`
	AccessToken  string    `gorm:"size:500" json:"-"`
	ExpiresAt    time.Time `gorm:"not null" json:"expires_at"`
	IsActive     bool      `gorm:"default:true" json:"is_active"`
	CreatedAt    time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt    time.Time `gorm:"autoUpdateTime" json:"updated_at"`

	// 关联关系
	User User `gorm:"foreignKey:UserID;references:ID" json:"user,omitempty"`
}

// TableName 指定表名
func (Session) TableName() string {
	return "user_sessions"
}

// BeforeCreate 创建前钩子 - 生成UUID
func (s *Session) BeforeCreate(tx *gorm.DB) error {
	if s.ID == "" {
		s.ID = uuid.New().String()
	}
	return nil
}

// IsExpired 检查会话是否过期
func (s *Session) IsExpired() bool {
	return time.Now().After(s.ExpiresAt)
}

// Activate 激活会话
func (s *Session) Activate() {
	s.IsActive = true
}

// Deactivate 停用会话
func (s *Session) Deactivate() {
	s.IsActive = false
}

// IsValid 检查会话是否有效
func (s *Session) IsValid() bool {
	return s.IsActive && !s.IsExpired()
}

// ExtendExpiration 延长会话有效期
func (s *Session) ExtendExpiration(duration time.Duration) {
	s.ExpiresAt = time.Now().Add(duration)
}
