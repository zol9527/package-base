package interfaces

import (
	"context"
	"package-base/internal/models/rbac"
)

// Function: RoleRepository
//
// Description: 角色仓储接口定义，负责角色与映射的持久化。
//
// Parameters:
//
//	无
//
// Returns:
//
//	无
type RoleRepository interface {
	Create(ctx context.Context, role *rbac.Role) error
	Update(ctx context.Context, role *rbac.Role) error
	Delete(ctx context.Context, id uint) error
	GetByID(ctx context.Context, id uint) (*rbac.Role, error)
	List(ctx context.Context, keyword string, page, size int) ([]*rbac.Role, int64, error)
}

// Function: PermissionRepository
//
// Description: 权限仓储接口定义。
//
// Parameters:
//
//	无
//
// Returns:
//
//	无
type PermissionRepository interface {
	Create(ctx context.Context, perm *rbac.Permission) error
	ReplaceRolePermissions(ctx context.Context, roleID uint, permIDs []uint) error
	ListByRole(ctx context.Context, roleID uint) ([]*rbac.Permission, error)
}

// Function: BindingRepository
//
// Description: 用户-角色绑定仓储接口。
//
// Parameters:
//
//	无
//
// Returns:
//
//	无
type BindingRepository interface {
	BindUserRoles(ctx context.Context, userID string, roleIDs []uint, scope *string) error
	ListUserRoles(ctx context.Context, userID string) ([]*rbac.Role, error)
}

// Function: AuditRepository
//
// Description: 审计记录仓储接口。
//
// Parameters:
//
//	无
//
// Returns:
//
//	无
type AuditRepository interface {
	AppendDecision(ctx context.Context, rec *rbac.AccessDecision) error
	// Function: QueryDecisions
	//
	// Description: 按可选过滤条件查询审计记录，按时间倒序，支持 limit。
	//
	// Parameters:
	//    ctx(context.Context): 上下文
	//    userID(string): 过滤用户ID（可空）
	//    resource(string): 过滤资源（可空，支持前缀匹配由上层处理，这里按精确匹配）
	//    action(string): 过滤动作（可空）
	//    limit(int): 最大返回条数（<=0 则默认 100）
	//
	// Returns:
	//    ([]*rbac.AccessDecision, error): 结果与错误
	QueryDecisions(ctx context.Context, userID, resource, action string, limit int) ([]*rbac.AccessDecision, error)
}
