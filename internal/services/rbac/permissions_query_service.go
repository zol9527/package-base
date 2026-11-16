package rbac

import (
	"context"
	coreIf "package-base/internal/core/interfaces"
	repoIf "package-base/internal/repositories/interfaces"
	"package-base/internal/services/interfaces"
)

// Function: NewPermissionsQueryService
//
// Description: 创建权限查询服务实例。
//
// Parameters:
//
//	logger(coreIf.Logger): 日志
//	bindingRepo(repoIf.BindingRepository): 绑定仓储
//	permRepo(repoIf.PermissionRepository): 权限仓储
//
// Returns:
//
//	(interfaces.PermissionsQueryService): 服务实现
func NewPermissionsQueryService(logger coreIf.Logger, bindingRepo repoIf.BindingRepository, permRepo repoIf.PermissionRepository) interfaces.PermissionsQueryService {
	return &permissionsQueryService{logger: logger, bindingRepo: bindingRepo, permRepo: permRepo}
}

type permissionsQueryService struct {
	logger      coreIf.Logger
	bindingRepo repoIf.BindingRepository
	permRepo    repoIf.PermissionRepository
}

// Function: GetEffectivePermissions
//
// Description: 聚合用户绑定角色的权限集合，按 deny 优先规则合并并去重后返回。
//
// Parameters:
//
//	ctx(context.Context): 上下文
//	userID(string): 用户ID
//
// Returns:
//
//	([]interfaces.PermissionDTO): 有效权限列表
//	(error): 错误
func (s *permissionsQueryService) GetEffectivePermissions(ctx context.Context, userID string) ([]interfaces.PermissionDTO, error) {
	if userID == "" {
		return nil, simpleError("user id is required")
	}

	// 1) 找到用户绑定的所有角色
	roles, err := s.bindingRepo.ListUserRoles(ctx, userID)
	if err != nil {
		return nil, err
	}
	if len(roles) == 0 {
		return []interfaces.PermissionDTO{}, nil
	}

	// 2) 聚合所有角色的权限
	type key struct{ r, a string }
	allows := map[key]bool{}
	denies := map[key]bool{}

	for _, role := range roles {
		perms, err := s.permRepo.ListByRole(ctx, role.ID)
		if err != nil {
			return nil, err
		}
		for _, p := range perms {
			k := key{r: p.Resource, a: p.Action}
			if p.Effect == "deny" {
				denies[k] = true
			} else {
				allows[k] = true
			}
		}
	}

	// 3) 按 deny 优先合并
	out := make([]interfaces.PermissionDTO, 0, len(allows)+len(denies))
	// 先把 deny 写入；若存在 allow 相同键，仍以 deny 覆盖
	for k := range denies {
		out = append(out, interfaces.PermissionDTO{Resource: k.r, Action: k.a, Effect: "deny"})
		if allows[k] {
			delete(allows, k)
		}
	}
	// 再写入余下 allow
	for k := range allows {
		out = append(out, interfaces.PermissionDTO{Resource: k.r, Action: k.a, Effect: "allow"})
	}
	return out, nil
}
