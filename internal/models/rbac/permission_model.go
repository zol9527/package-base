package rbac

import "time"

// Function: Permission
//
// Description: 权限模型，描述资源-动作-效果的授权规则。
//
// Parameters:
//
//    无
//
// Returns:
//
//    无
 type Permission struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	Resource  string    `json:"resource" gorm:"not null;size:255"`
	Action    string    `json:"action" gorm:"not null;size:32"`
	Effect    string    `json:"effect" gorm:"not null;size:8"`
	Note      string    `json:"note" gorm:"size:255"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
 }
