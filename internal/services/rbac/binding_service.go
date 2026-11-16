package rbac

import (
	"context"
	coreIf "package-base/internal/core/interfaces"
	repoIf "package-base/internal/repositories/interfaces"
	svcIf "package-base/internal/services/interfaces"
)

// Function: NewBindingService
//
// Description: 创建绑定服务实例。
//
// Parameters:
//
//	logger(coreIf.Logger): 日志记录器
//	bindingRepo(repoIf.BindingRepository): 绑定仓储
//
// Returns:
//
//	(svcIf.BindingService): 绑定服务实现
func NewBindingService(logger coreIf.Logger, bindingRepo repoIf.BindingRepository) svcIf.BindingService {
	return &bindingService{logger: logger, bindingRepo: bindingRepo}
}

type bindingService struct {
	logger      coreIf.Logger
	bindingRepo repoIf.BindingRepository
}

// Function: BindUserRoles
//
// Description: 在指定 scope 下以覆盖方式绑定用户的角色集合。
//
// Parameters:
//
//	ctx(context.Context): 上下文
//	userID(string): 用户ID
//	roleIDs([]uint): 角色ID集合
//	scope(*string): 作用域
//
// Returns:
//
//	error: 错误信息
func (s *bindingService) BindUserRoles(ctx context.Context, userID string, roleIDs []uint, scope *string) error {
	if userID == "" {
		return simpleError("user id is required")
	}
	return s.bindingRepo.BindUserRoles(ctx, userID, roleIDs, scope)
}
