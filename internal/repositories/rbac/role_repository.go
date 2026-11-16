package rbac

import (
	"context"
	coreInterfaces "package-base/internal/core/interfaces"
	"package-base/internal/models/rbac"
	repoInterfaces "package-base/internal/repositories/interfaces"
)

// Function: RoleGormRepository
//
// Description: 基于 GORM 的角色仓储实现骨架。
//
// Parameters:
//
//	无
//
// Returns:
//
//	无
type RoleGormRepository struct {
	db     coreInterfaces.RelationalStorage
	logger coreInterfaces.Logger
}

// Function: NewRoleRepository
//
// Description: 创建角色仓储实现。
//
// Parameters:
//
//	db(coreInterfaces.RelationalStorage): 关系存储
//	logger(coreInterfaces.Logger): 日志
//
// Returns:
//
//	repoInterfaces.RoleRepository: 仓储接口实现
func NewRoleRepository(db coreInterfaces.RelationalStorage, logger coreInterfaces.Logger) repoInterfaces.RoleRepository {
	return &RoleGormRepository{db: db, logger: logger}
}

// Function: Create
//
// Description: 创建角色。
//
// Parameters:
//
//	ctx(context.Context): 上下文
//	role(*rbac.Role): 角色对象
//
// Returns:
//
//	error(error): 可能的错误
func (r *RoleGormRepository) Create(ctx context.Context, role *rbac.Role) error {
	return r.db.GetDB().WithContext(ctx).Create(role).Error
}

// Function: Update
//
// Description: 更新角色。
//
// Parameters:
//
//	ctx(context.Context): 上下文
//	role(*rbac.Role): 角色对象
//
// Returns:
//
//	error(error): 可能的错误
func (r *RoleGormRepository) Update(ctx context.Context, role *rbac.Role) error {
	return r.db.GetDB().WithContext(ctx).Save(role).Error
}

// Function: Delete
//
// Description: 删除角色（后续注意内置角色保护与外键关联）。
//
// Parameters:
//
//	ctx(context.Context): 上下文
//	id(uint): 角色ID
//
// Returns:
//
//	error(error): 可能的错误
func (r *RoleGormRepository) Delete(ctx context.Context, id uint) error {
	return r.db.GetDB().WithContext(ctx).Delete(&rbac.Role{}, id).Error
}

// Function: GetByID
//
// Description: 通过ID获取角色。
//
// Parameters:
//
//	ctx(context.Context): 上下文
//	id(uint): 角色ID
//
// Returns:
//
//	(*rbac.Role, error): 角色或错误
func (r *RoleGormRepository) GetByID(ctx context.Context, id uint) (*rbac.Role, error) {
	var m rbac.Role
	if err := r.db.GetDB().WithContext(ctx).First(&m, id).Error; err != nil {
		return nil, err
	}
	return &m, nil
}

// Function: List
//
// Description: 角色分页列表（按名称模糊）。
//
// Parameters:
//
//	ctx(context.Context): 上下文
//	keyword(string): 名称关键字
//	page(int): 页码
//	size(int): 页大小
//
// Returns:
//
//	([]*rbac.Role, int64, error): 角色集合、总数、错误
func (r *RoleGormRepository) List(ctx context.Context, keyword string, page, size int) ([]*rbac.Role, int64, error) {
	db := r.db.GetDB().WithContext(ctx).Model(&rbac.Role{})
	if keyword != "" {
		db = db.Where("name LIKE ?", "%"+keyword+"%")
	}
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if page <= 0 {
		page = 1
	}
	if size <= 0 {
		size = 10
	}
	var rows []*rbac.Role
	if err := db.Offset((page - 1) * size).Limit(size).Order("id DESC").Find(&rows).Error; err != nil {
		return nil, 0, err
	}
	return rows, total, nil
}
