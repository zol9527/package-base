package rbac

import (
	"context"
	coreInterfaces "package-base/internal/core/interfaces"
	models "package-base/internal/models"
	model "package-base/internal/models/rbac"
	"package-base/internal/options"
	repoInterfaces "package-base/internal/repositories/interfaces"
	serviceInterfaces "package-base/internal/services/interfaces"
)

// Function: Service
//
// Description: RBAC 服务骨架，承载授权判定与聚合。
//
// Parameters:
//
//	无
//
// Returns:
//
//	无
type Service struct {
	logger      coreInterfaces.Logger
	roleRepo    repoInterfaces.RoleRepository
	permRepo    repoInterfaces.PermissionRepository
	bindingRepo repoInterfaces.BindingRepository
	auditRepo   repoInterfaces.AuditRepository
	userRepo    repoInterfaces.UserRepository
}

// Function: NewService
//
// Description: 创建 RBAC 服务实现。
//
// Parameters:
//
//	logger(coreInterfaces.Logger): 日志
//
// Returns:
//
//	*Service: 服务实例
func NewService(logger coreInterfaces.Logger) *Service { // 兼容旧调用
	return &Service{logger: logger}
}

// Function: NewServiceWithRepos
//
// Description: 使用仓储依赖创建 RBAC 服务实例。
//
// Parameters:
//
//	logger(coreInterfaces.Logger): 日志
//	roleRepo(repoInterfaces.RoleRepository): 角色仓储
//	permRepo(repoInterfaces.PermissionRepository): 权限仓储
//	bindingRepo(repoInterfaces.BindingRepository): 绑定仓储
//	auditRepo(repoInterfaces.AuditRepository): 审计仓储
//
// Returns:
//
//	*Service: 服务实例
func NewServiceWithRepos(
	logger coreInterfaces.Logger,
	roleRepo repoInterfaces.RoleRepository,
	permRepo repoInterfaces.PermissionRepository,
	bindingRepo repoInterfaces.BindingRepository,
	auditRepo repoInterfaces.AuditRepository,
	userRepo repoInterfaces.UserRepository,
) *Service {
	return &Service{
		logger:      logger,
		roleRepo:    roleRepo,
		permRepo:    permRepo,
		bindingRepo: bindingRepo,
		auditRepo:   auditRepo,
		userRepo:    userRepo,
	}
}

// Function: Authorize
//
// Description: 授权判定占位（后续实现策略：deny>精确>通配，结合继承与聚合结果）。
//
// Parameters:
//
//	ctx(context.Context): 上下文
//	userID(string): 用户ID
//	resource(string): 资源
//	action(string): 动作
//
// Returns:
//
//	allowed(bool): 是否允许
//	reason(string): 拒绝原因
func (s *Service) Authorize(ctx context.Context, userID, resource, action string) (bool, string) {
	// 账户状态优先：禁用/暂停/未激活直接拒绝
	if s.userRepo != nil && userID != "" {
		user, err := s.userRepo.FindUser(ctx, options.WithID(userID))
		if err == nil && user != nil {
			switch user.Status {
			case models.UserStatusActive:
				// 继续进行策略评估
			case models.UserStatusSuspended:
				return false, "account_suspended"
			case models.UserStatusBanned:
				return false, "account_banned"
			default:
				return false, "account_inactive"
			}
		}
	}
	// TODO: 从仓储聚合权限并执行策略评估；当前默认拒绝
	return false, "not implemented"
}

// 确保实现接口
var _ serviceInterfaces.AuthorizationService = (*Service)(nil)

// Function: AuditDecision
//
// Description: 写入访问决策审计记录（允许/拒绝均可记录）。
//
// Parameters:
//
//	ctx(context.Context): 上下文
//	userID(string): 用户ID
//	resource(string): 资源
//	action(string): 动作
//	allowed(bool): 是否允许
//	reason(string): 原因描述
//	requestID(string): 请求ID（可选）
//	client(string): 客户端信息（可选）
//
// Returns:
//
//	error(error): 可能的错误
func (s *Service) AuditDecision(ctx context.Context, userID, resource, action string, allowed bool, reason, requestID, client string) error {
	if s.auditRepo == nil {
		return nil
	}
	rec := &model.AccessDecision{
		UserID:    userID,
		Resource:  resource,
		Action:    action,
		Allowed:   allowed,
		Reason:    reason,
		RequestID: requestID,
		Client:    client,
	}
	return s.auditRepo.AppendDecision(ctx, rec)
}
