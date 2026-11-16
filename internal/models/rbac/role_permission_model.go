package rbac

import "time"

// Function: RolePermission
//
// Description: 角色与权限的多对多映射关系。
//
// Parameters:
//    无
//
// Returns:
//    无
type RolePermission struct {
	ID           uint      `json:"id" gorm:"primaryKey"`
	RoleID       uint      `json:"role_id" gorm:"not null;index"`
	PermissionID uint      `json:"permission_id" gorm:"not null;index"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}
