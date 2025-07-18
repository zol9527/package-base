package models

import (
	"time"

	"package-base/internal/core/utils"

	"gorm.io/gorm"
)

// TotpSecret TOTP密钥模型
type TotpSecret struct {
	ID        string    `gorm:"primaryKey;size:36" json:"id"`
	UserID    string    `gorm:"not null;index;size:36" json:"user_id"`
	Secret    string    `gorm:"not null;size:255" json:"-"` // 不在JSON中暴露
	Issuer    string    `gorm:"size:100;default:'AuthSystem'" json:"issuer"`
	IsActive  bool      `gorm:"default:false" json:"is_active"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`

	// 关联关系
	User User `gorm:"foreignKey:UserID;references:ID" json:"user,omitempty"`
}

// TableName 指定表名
func (TotpSecret) TableName() string {
	return "user_totp_secrets"
}

// BeforeCreate 创建前钩子 - 生成UUID
func (t *TotpSecret) BeforeCreate(tx *gorm.DB) error {
	if t.ID == "" {
		t.ID = generateUUID()
	}
	return nil
}

// Activate 激活TOTP
func (t *TotpSecret) Activate() {
	t.IsActive = true
}

// Deactivate 停用TOTP
func (t *TotpSecret) Deactivate() {
	t.IsActive = false
}

// BackupCode 备用码模型
type BackupCode struct {
	ID        string     `gorm:"primaryKey;size:36" json:"id"`
	UserID    string     `gorm:"not null;index;size:36" json:"user_id"`
	Code      string     `gorm:"not null;size:20;uniqueIndex" json:"-"` // 不在JSON中暴露
	IsUsed    bool       `gorm:"default:false" json:"is_used"`
	UsedAt    *time.Time `gorm:"null" json:"used_at,omitempty"`
	CreatedAt time.Time  `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time  `gorm:"autoUpdateTime" json:"updated_at"`

	// 关联关系
	User User `gorm:"foreignKey:UserID;references:ID" json:"user,omitempty"`
}

// TableName 指定表名
func (BackupCode) TableName() string {
	return "user_backup_codes"
}

// BeforeCreate 创建前钩子 - 生成UUID
func (b *BackupCode) BeforeCreate(tx *gorm.DB) error {
	if b.ID == "" {
		b.ID = generateUUID()
	}
	return nil
}

// MarkAsUsed 标记为已使用
func (b *BackupCode) MarkAsUsed() {
	b.IsUsed = true
	now := time.Now()
	b.UsedAt = &now
}

// IsExpired 检查是否过期（备用码通常不过期，但可以根据业务需要实现）
func (b *BackupCode) IsExpired() bool {
	// 备用码默认不过期，可根据业务需求修改
	return false
}

// generateUUID 内部UUID生成函数
func generateUUID() string {
	return utils.GenerateUUID()
}
