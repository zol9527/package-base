package rbac

import (
	"context"
	coreInterfaces "package-base/internal/core/interfaces"
	model "package-base/internal/models/rbac"
	repoInterfaces "package-base/internal/repositories/interfaces"
)

// Function: BindingGormRepository
//
// Description: 基于 GORM 的用户-角色绑定仓储实现骨架。
//
// Parameters:
//
//	无
//
// Returns:
//
//	无
type BindingGormRepository struct {
	db     coreInterfaces.RelationalStorage
	logger coreInterfaces.Logger
}

// Function: NewBindingRepository
//
// Description: 创建绑定仓储实现。
//
// Parameters:
//
//	db(coreInterfaces.RelationalStorage): 关系存储
//	logger(coreInterfaces.Logger): 日志
//
// Returns:
//
//	repoInterfaces.BindingRepository: 仓储接口实现
func NewBindingRepository(db coreInterfaces.RelationalStorage, logger coreInterfaces.Logger) repoInterfaces.BindingRepository {
	return &BindingGormRepository{db: db, logger: logger}
}

// Function: BindUserRoles
//
// Description: 绑定用户的角色集合（同一 scope 下覆盖旧绑定）。
//
// Parameters:
//
//	ctx(context.Context): 上下文
//	userID(string): 用户ID
//	roleIDs([]uint): 角色ID集合
//	scope(*string): 作用域，可为 nil
//
// Returns:
//
//	error(error): 可能的错误
func (r *BindingGormRepository) BindUserRoles(ctx context.Context, userID string, roleIDs []uint, scope *string) error {
	db := r.db.GetDB().WithContext(ctx)
	// 删除旧绑定
	tx := db.Where("user_id = ?", userID)
	if scope != nil {
		tx = tx.Where("scope = ?", *scope)
	}
	if err := tx.Delete(&model.UserRoleBinding{}).Error; err != nil {
		return err
	}
	if len(roleIDs) == 0 {
		return nil
	}
	// 新增绑定
	rows := make([]*model.UserRoleBinding, 0, len(roleIDs))
	for _, rid := range roleIDs {
		rows = append(rows, &model.UserRoleBinding{UserID: userID, RoleID: rid, Scope: scope})
	}
	return db.Create(&rows).Error
}

// Function: ListUserRoles
//
// Description: 列出用户的角色集合（不含继承，纯绑定）。
//
// Parameters:
//
//	ctx(context.Context): 上下文
//	userID(string): 用户ID
//
// Returns:
//
//	([]*model.Role, error): 角色列表与错误
func (r *BindingGormRepository) ListUserRoles(ctx context.Context, userID string) ([]*model.Role, error) {
	db := r.db.GetDB().WithContext(ctx)
	var roles []*model.Role
	err := db.Table("roles").
		Joins("JOIN user_role_bindings b ON b.role_id = roles.id").
		Where("b.user_id = ?", userID).
		Order("roles.id ASC").
		Find(&roles).Error
	return roles, err
}
