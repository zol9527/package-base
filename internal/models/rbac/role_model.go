package rbac

import "time"

// Function: Role
//
// Description: 角色模型，定义角色基础信息与状态。
//
// Parameters:
//
//    无
//
// Returns:
//
//    无
 type Role struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	Name        string    `json:"name" gorm:"unique;not null;size:64"`
	Description string    `json:"description" gorm:"size:255"`
	BuiltIn     bool      `json:"built_in" gorm:"not null;default:false"`
	Enabled     bool      `json:"enabled" gorm:"not null;default:true"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
 }
