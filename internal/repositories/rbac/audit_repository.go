package rbac

import (
	"context"
	coreInterfaces "package-base/internal/core/interfaces"
	model "package-base/internal/models/rbac"
	repoInterfaces "package-base/internal/repositories/interfaces"
)

// Function: AuditGormRepository
//
// Description: 基于 GORM 的审计记录仓储实现骨架。
//
// Parameters:
//
//	无
//
// Returns:
//
//	无
type AuditGormRepository struct {
	db     coreInterfaces.RelationalStorage
	logger coreInterfaces.Logger
}

// Function: NewAuditRepository
//
// Description: 创建审计仓储实现。
//
// Parameters:
//
//	db(coreInterfaces.RelationalStorage): 关系存储
//	logger(coreInterfaces.Logger): 日志
//
// Returns:
//
//	repoInterfaces.AuditRepository: 仓储接口实现
func NewAuditRepository(db coreInterfaces.RelationalStorage, logger coreInterfaces.Logger) repoInterfaces.AuditRepository {
	return &AuditGormRepository{db: db, logger: logger}
}

// Function: AppendDecision
//
// Description: 写入一条访问决策审计记录。
//
// Parameters:
//
//	ctx(context.Context): 上下文
//	rec(*model.AccessDecision): 审计记录
//
// Returns:
//
//	error(error): 可能的错误
func (r *AuditGormRepository) AppendDecision(ctx context.Context, rec *model.AccessDecision) error {
	return r.db.GetDB().WithContext(ctx).Create(rec).Error
}

// Function: QueryDecisions
//
// Description: 按条件查询审计记录，按创建时间倒序，支持 limit。
//
// Parameters:
//
//	ctx(context.Context): 上下文
//	userID(string): 用户ID（可空）
//	resource(string): 资源（可空）
//	action(string): 动作（可空）
//	limit(int): 限制条数（<=0 则默认 100）
//
// Returns:
//
//	([]*model.AccessDecision, error): 结果与错误
func (r *AuditGormRepository) QueryDecisions(ctx context.Context, userID, resource, action string, limit int) ([]*model.AccessDecision, error) {
	db := r.db.GetDB().WithContext(ctx)
	q := db.Table("access_decisions")
	if userID != "" {
		q = q.Where("user_id = ?", userID)
	}
	if resource != "" {
		q = q.Where("resource = ?", resource)
	}
	if action != "" {
		q = q.Where("action = ?", action)
	}
	if limit <= 0 {
		limit = 100
	}
	var rows []*model.AccessDecision
	err := q.Order("created_at DESC").Limit(limit).Find(&rows).Error
	return rows, err
}
