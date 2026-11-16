package rbac

import "time"

// Function: AccessDecision
//
// Description: 访问决策审计记录，记录允许/拒绝与原因。
//
// Parameters:
//    无
//
// Returns:
//    无
 type AccessDecision struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	UserID    string    `json:"user_id" gorm:"not null;size:64;index"`
	Resource  string    `json:"resource" gorm:"not null;size:255;index"`
	Action    string    `json:"action" gorm:"not null;size:32;index"`
	Allowed   bool      `json:"allowed" gorm:"not null"`
	Reason    string    `json:"reason" gorm:"size:255"`
	RequestID string    `json:"request_id" gorm:"size:64;index"`
	Client    string    `json:"client_info" gorm:"size:255"`
	CreatedAt time.Time `json:"created_at"`
 }
