package interfaces

// Function: RBAC 概念与接口
//
// Description: 定义 RBAC 相关的跨层核心接口与共用类型（管理服务接口的抽象）。
//
// Parameters:
//    无
//
// Returns:
//    无

// Function: Action 常量
//
// Description: 权限动作枚举
//
// Parameters:
//
//	无
//
// Returns:
//
//	无
const (
	ACTION_READ   = "read"
	ACTION_WRITE  = "write"
	ACTION_DELETE = "delete"
	ACTION_MANAGE = "manage"
)

// Function: Effect 常量
//
// Description: 权限效果枚举
//
// Parameters:
//
//	无
//
// Returns:
//
//	无
const (
	EFFECT_ALLOW = "allow"
	EFFECT_DENY  = "deny"
)

// Function: RBACService
//
// Description: RBAC 管理服务接口抽象（角色/权限/绑定的管理能力）。
//
// Parameters:
//
//	无
//
// Returns:
//
//	无
type RBACService interface {
	// 角色/权限/绑定的管理接口将在后续实现阶段补充
}
