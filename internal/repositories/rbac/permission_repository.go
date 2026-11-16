package rbac

import (
	"context"
	coreInterfaces "package-base/internal/core/interfaces"
	model "package-base/internal/models/rbac"
	repoInterfaces "package-base/internal/repositories/interfaces"
)

// Function: PermissionGormRepository
//
// Description: 基于 GORM 的权限仓储实现骨架。
//
// Parameters:
//
//	无
//
// Returns:
//
//	无
type PermissionGormRepository struct {
	db     coreInterfaces.RelationalStorage
	logger coreInterfaces.Logger
}

// Function: NewPermissionRepository
//
// Description: 创建权限仓储实现。
//
// Parameters:
//
//	db(coreInterfaces.RelationalStorage): 关系存储
//	logger(coreInterfaces.Logger): 日志
//
// Returns:
//
//	repoInterfaces.PermissionRepository: 仓储接口实现
func NewPermissionRepository(db coreInterfaces.RelationalStorage, logger coreInterfaces.Logger) repoInterfaces.PermissionRepository {
	return &PermissionGormRepository{db: db, logger: logger}
}

// Function: Create
//
// Description: 创建权限实体。
//
// Parameters:
//
//	ctx(context.Context): 上下文
//	perm(*model.Permission): 权限对象
//
// Returns:
//
//	error(error): 可能的错误
func (r *PermissionGormRepository) Create(ctx context.Context, perm *model.Permission) error {
	return r.db.GetDB().WithContext(ctx).Create(perm).Error
}

// Function: ReplaceRolePermissions
//
// Description: 替换角色的权限集合（全量覆盖）。
//
// Parameters:
//
//	ctx(context.Context): 上下文
//	roleID(uint): 角色ID
//	permIDs([]uint): 权限ID集合
//
// Returns:
//
//	error(error): 可能的错误
func (r *PermissionGormRepository) ReplaceRolePermissions(ctx context.Context, roleID uint, permIDs []uint) error {
	db := r.db.GetDB().WithContext(ctx)
	// 先清空映射
	if err := db.Where("role_id = ?", roleID).Delete(&model.RolePermission{}).Error; err != nil {
		return err
	}
	// 再批量插入
	if len(permIDs) == 0 {
		return nil
	}
	mappings := make([]*model.RolePermission, 0, len(permIDs))
	for _, pid := range permIDs {
		mappings = append(mappings, &model.RolePermission{RoleID: roleID, PermissionID: pid})
	}
	return db.Create(&mappings).Error
}

// Function: ListByRole
//
// Description: 按角色列出权限集合。
//
// Parameters:
//
//	ctx(context.Context): 上下文
//	roleID(uint): 角色ID
//
// Returns:
//
//	([]*model.Permission, error): 权限列表与错误
func (r *PermissionGormRepository) ListByRole(ctx context.Context, roleID uint) ([]*model.Permission, error) {
	db := r.db.GetDB().WithContext(ctx)
	var perms []*model.Permission
	err := db.Table("permissions").
		Joins("JOIN role_permissions rp ON rp.permission_id = permissions.id").
		Where("rp.role_id = ?", roleID).
		Order("permissions.id ASC").
		Find(&perms).Error
	return perms, err
}
