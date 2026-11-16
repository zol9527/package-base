package rbac

import (
	"context"
	coreIf "package-base/internal/core/interfaces"
	rbacModel "package-base/internal/models/rbac"
	repoIf "package-base/internal/repositories/interfaces"
	svcIf "package-base/internal/services/interfaces"
)

// Function: NewRoleService
//
// Description: 创建角色服务实例。
//
// Parameters:
//
//	logger(coreIf.Logger): 日志记录器
//	roleRepo(repoIf.RoleRepository): 角色仓储
//
// Returns:
//
//	(svcIf.RoleService): 角色服务接口实现
func NewRoleService(logger coreIf.Logger, roleRepo repoIf.RoleRepository) svcIf.RoleService {
	return &roleService{logger: logger, roleRepo: roleRepo}
}

type roleService struct {
	logger   coreIf.Logger
	roleRepo repoIf.RoleRepository
}

// Function: CreateRole
//
// Description: 创建新角色，进行基本校验并写入数据库。
//
// Parameters:
//
//	ctx(context.Context): 上下文
//	role(*rbacModel.Role): 角色实体
//
// Returns:
//
//	(*rbacModel.Role): 创建后的角色
//	(error): 错误信息
func (s *roleService) CreateRole(ctx context.Context, role *rbacModel.Role) (*rbacModel.Role, error) {
	// 基本校验
	if role == nil || role.Name == "" {
		return nil, errBadRequest("role name is required")
	}
	// 默认值：BuiltIn=false, Enabled=true（模型已有默认）
	role.BuiltIn = false
	if !role.Enabled {
		role.Enabled = true
	}

	// 唯一性校验：利用 List 搜索重名（简化实现）
	if rows, _, err := s.roleRepo.List(ctx, role.Name, 1, 1); err == nil && len(rows) > 0 {
		for _, r := range rows {
			if r.Name == role.Name {
				return nil, errConflict("role name already exists")
			}
		}
	}

	if err := s.roleRepo.Create(ctx, role); err != nil {
		return nil, err
	}
	return role, nil
}

// Function: UpdateRole
//
// Description: 更新角色基本信息。
//
// Parameters:
//
//	ctx(context.Context): 上下文
//	id(uint): 角色ID
//	role(*rbacModel.Role): 字段载体（Name/Description/Enabled）
//
// Returns:
//
//	(*rbacModel.Role): 更新后的实体
//	(error): 错误
func (s *roleService) UpdateRole(ctx context.Context, id uint, role *rbacModel.Role) (*rbacModel.Role, error) {
	exist, err := s.roleRepo.GetByID(ctx, id)
	if err != nil {
		return nil, errNotFound("role not found")
	}
	if exist.BuiltIn {
		return nil, errForbidden("built-in role cannot be modified")
	}
	// 名称变更时做重名保护
	if role.Name != "" && role.Name != exist.Name {
		if rows, _, err := s.roleRepo.List(ctx, role.Name, 1, 1); err == nil && len(rows) > 0 {
			for _, r := range rows {
				if r.Name == role.Name {
					return nil, errConflict("role name already exists")
				}
			}
		}
		exist.Name = role.Name
	}
	if role.Description != "" {
		exist.Description = role.Description
	}
	// 启用状态若显式提供则更新
	if role.Enabled || (!role.Enabled && (exist.Enabled != role.Enabled)) {
		exist.Enabled = role.Enabled
	}
	if err := s.roleRepo.Update(ctx, exist); err != nil {
		return nil, err
	}
	return exist, nil
}

// Function: DeleteRole
//
// Description: 删除角色，保护内置角色不可删除。
//
// Parameters:
//
//	ctx(context.Context): 上下文
//	id(uint): 角色ID
//
// Returns:
//
//	(error): 错误
func (s *roleService) DeleteRole(ctx context.Context, id uint) error {
	exist, err := s.roleRepo.GetByID(ctx, id)
	if err != nil {
		return errNotFound("role not found")
	}
	if exist.BuiltIn {
		return errForbidden("built-in role cannot be deleted")
	}
	return s.roleRepo.Delete(ctx, id)
}

// 轻量错误类型定义
type simpleError string

func (e simpleError) Error() string { return string(e) }

func errBadRequest(msg string) error { return simpleError(msg) }
func errConflict(msg string) error   { return simpleError(msg) }
func errNotFound(msg string) error   { return simpleError(msg) }
func errForbidden(msg string) error  { return simpleError(msg) }
