package interfaces

import "context"

// Function: BindingService
//
// Description: 用户与角色绑定服务接口，负责按照 scope 覆盖式绑定用户的角色集合。
//
// Parameters:
//
//	无
//
// Returns:
//
//	无
type BindingService interface {
	// Function: BindUserRoles
	//
	// Description: 为给定用户在指定 scope 下绑定角色集合（覆盖旧绑定）。
	//
	// Parameters:
	//    ctx(context.Context): 上下文
	//    userID(string): 用户ID
	//    roleIDs([]uint): 角色ID集合
	//    scope(*string): 作用域，可为空
	//
	// Returns:
	//    error: 绑定失败时返回错误
	BindUserRoles(ctx context.Context, userID string, roleIDs []uint, scope *string) error
}
