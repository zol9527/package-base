package interfaces

import (
	"context"
	rbacModel "package-base/internal/models/rbac"
)

// Function: RoleService
//
// Description: 角色管理服务接口，负责角色的创建、更新、删除与查询等业务逻辑。
//
// Parameters:
//
//	无
//
// Returns:
//
//	无
type RoleService interface {
	// Function: CreateRole
	//
	// Description: 创建一个新角色，校验名称与属性并持久化。
	//
	// Parameters:
	//    ctx(context.Context): 上下文
	//    role(*rbacModel.Role): 待创建的角色实体（仅使用 Name、Description、Enabled 字段）
	//
	// Returns:
	//    (*rbacModel.Role): 创建后的角色（包含ID与时间戳）
	//    (error): 错误信息（如名称重复等）
	CreateRole(ctx context.Context, role *rbacModel.Role) (*rbacModel.Role, error)

	// Function: UpdateRole
	//
	// Description: 更新角色的名称/描述/启用状态。
	//
	// Parameters:
	//    ctx(context.Context): 上下文
	//    id(uint): 角色ID
	//    role(*rbacModel.Role): 待更新字段（Name/Description/Enabled）
	//
	// Returns:
	//    (*rbacModel.Role): 更新后的角色
	//    (error): 错误信息（如未找到、冲突等）
	UpdateRole(ctx context.Context, id uint, role *rbacModel.Role) (*rbacModel.Role, error)

	// Function: DeleteRole
	//
	// Description: 删除指定角色（内置角色禁止删除）。
	//
	// Parameters:
	//    ctx(context.Context): 上下文
	//    id(uint): 角色ID
	//
	// Returns:
	//    (error): 错误信息
	DeleteRole(ctx context.Context, id uint) error
}
