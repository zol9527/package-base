package interfaces

import (
	"context"
)

// Function: PermissionsQueryService
//
// Description: 权限查询服务接口，提供用户有效权限聚合查询能力（已做 deny 优先去重）。
//
// Parameters:
//
//	无
//
// Returns:
//
//	无
type PermissionsQueryService interface {
	// Function: GetEffectivePermissions
	//
	// Description: 基于用户ID聚合并返回有效权限列表（格式 resource/action/effect），对冲突规则应用 deny 优先策略。
	//
	// Parameters:
	//    ctx(context.Context): 上下文
	//    userID(string): 用户ID
	//
	// Returns:
	//    perms([]PermissionDTO): 有效权限列表
	//    err(error): 错误信息
	GetEffectivePermissions(ctx context.Context, userID string) ([]PermissionDTO, error)
}

// Function: PermissionDTO
//
// Description: 权限传输对象，返回给上层/HTTP 层使用。
//
// Parameters:
//
//	无
//
// Returns:
//
//	无
type PermissionDTO struct {
	Resource string `json:"resource"`
	Action   string `json:"action"`
	Effect   string `json:"effect"`
}
