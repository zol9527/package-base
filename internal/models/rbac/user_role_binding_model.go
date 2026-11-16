package rbac

import "time"

// Function: UserRoleBinding
//
// Description: 用户-角色绑定关系，可选 scope 与过期时间。
//
// Parameters:
//    无
//
// Returns:
//    无
 type UserRoleBinding struct {
	ID        uint       `json:"id" gorm:"primaryKey"`
	UserID    string     `json:"user_id" gorm:"not null;index;size:64"`
	RoleID    uint       `json:"role_id" gorm:"not null;index"`
	Scope     *string    `json:"scope" gorm:"size:128"`
	ExpiresAt *time.Time `json:"expires_at"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
 }
