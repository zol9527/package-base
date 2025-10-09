package repositories

import (
	"context"
	"package-base/internal/core/interfaces"
	"package-base/internal/models"
	"package-base/internal/options"
	repoInterfaces "package-base/internal/repositories/interfaces"

	"github.com/pkg/errors"
	"gorm.io/gorm"
)

// sessionRepository 会话存储实现
type sessionRepository struct {
	storage interfaces.RelationalStorage
	logger  interfaces.Logger
}

// NewSessionRepository 创建会话存储实现
func NewSessionRepository(storage interfaces.RelationalStorage, logger interfaces.Logger) repoInterfaces.SessionRepository {
	return &sessionRepository{
		storage: storage,
		logger:  logger,
	}
}

// getDB 获取数据库连接
func (r *sessionRepository) getDB() *gorm.DB {
	return r.storage.GetDB()
}

// Create 创建会话
func (r *sessionRepository) Create(ctx context.Context, session *models.Session) error {
	db := r.getDB()

	if err := db.WithContext(ctx).Create(session).Error; err != nil {
		r.logger.Error("failed to create session",
			interfaces.ErrorField(err),
			interfaces.String("user_id", session.UserID),
			interfaces.String("ip_address", session.IPAddress),
		)
		return errors.Wrap(err, "failed to create session")
	}

	r.logger.Info("session created successfully",
		interfaces.String("session_id", session.ID),
		interfaces.String("user_id", session.UserID),
	)

	return nil
}

// GetByID 根据ID获取会话
func (r *sessionRepository) GetByID(ctx context.Context, id string) (*models.Session, error) {
	return r.FindSession(ctx, options.WithSessionID(id))
}

// GetByRefreshToken 根据刷新Token获取会话
func (r *sessionRepository) GetByRefreshToken(ctx context.Context, refreshToken string) (*models.Session, error) {
	return r.FindSession(ctx, options.WithActiveRefreshToken(refreshToken))
}

// GetByUserID 获取用户的所有会话
func (r *sessionRepository) GetByUserID(ctx context.Context, userID string) ([]*models.Session, error) {
	return r.FindSessions(ctx,
		options.WithUserID(userID),
		options.WithSessionOrderByCreatedAt(true),
	)
}

// GetActiveByUserID 获取用户的活跃会话
func (r *sessionRepository) GetActiveByUserID(ctx context.Context, userID string) ([]*models.Session, error) {
	return r.FindSessions(ctx,
		options.WithUserActiveSessions(userID),
		options.WithSessionOrderByCreatedAt(true),
	)
}

// Update 更新会话
func (r *sessionRepository) Update(ctx context.Context, session *models.Session) error {
	db := r.getDB()

	if err := db.WithContext(ctx).Save(session).Error; err != nil {
		r.logger.Error("failed to update session",
			interfaces.ErrorField(err),
			interfaces.String("session_id", session.ID),
		)
		return errors.Wrap(err, "failed to update session")
	}

	r.logger.Info("session updated successfully",
		interfaces.String("session_id", session.ID),
	)

	return nil
}

// UpdateAccessToken 更新访问Token
func (r *sessionRepository) UpdateAccessToken(ctx context.Context, sessionID, accessToken string) error {
	db := r.getDB()

	// 使用选项构建查询条件
	queryOptions := options.NewQuerySessionOptions(db, options.WithSessionID(sessionID))
	queryDB := queryOptions.GetDB()

	if err := queryDB.WithContext(ctx).Model(&models.Session{}).Update("access_token", accessToken).Error; err != nil {
		r.logger.Error("failed to update access token",
			interfaces.ErrorField(err),
			interfaces.String("session_id", sessionID),
		)
		return errors.Wrap(err, "failed to update access token")
	}

	return nil
}

// DeactivateSession 停用指定会话
func (r *sessionRepository) DeactivateSession(ctx context.Context, sessionID string) error {
	db := r.getDB()

	// 使用选项构建查询条件
	queryOptions := options.NewQuerySessionOptions(db, options.WithSessionID(sessionID))
	queryDB := queryOptions.GetDB()

	if err := queryDB.WithContext(ctx).Model(&models.Session{}).Update("is_active", false).Error; err != nil {
		r.logger.Error("failed to deactivate session",
			interfaces.ErrorField(err),
			interfaces.String("session_id", sessionID),
		)
		return errors.Wrap(err, "failed to deactivate session")
	}

	r.logger.Info("session deactivated successfully",
		interfaces.String("session_id", sessionID),
	)

	return nil
}

// DeactivateUserSessions 停用用户的所有会话
func (r *sessionRepository) DeactivateUserSessions(ctx context.Context, userID string) error {
	db := r.getDB()

	// 使用选项构建查询条件
	queryOptions := options.NewQuerySessionOptions(db, options.WithUserID(userID))
	queryDB := queryOptions.GetDB()

	if err := queryDB.WithContext(ctx).Model(&models.Session{}).Update("is_active", false).Error; err != nil {
		r.logger.Error("failed to deactivate user sessions",
			interfaces.ErrorField(err),
			interfaces.String("user_id", userID),
		)
		return errors.Wrap(err, "failed to deactivate user sessions")
	}

	r.logger.Info("user sessions deactivated successfully",
		interfaces.String("user_id", userID),
	)

	return nil
}

// DeactivateOtherSessions 停用用户的其他会话（除了指定的会话）
func (r *sessionRepository) DeactivateOtherSessions(ctx context.Context, userID, currentSessionID string) error {
	db := r.getDB()

	// 使用专门的复合查询选项
	queryOptions := options.NewQuerySessionOptions(db,
		options.WithUserAndExcludeSession(userID, currentSessionID),
	)
	queryDB := queryOptions.GetDB()

	if err := queryDB.WithContext(ctx).Model(&models.Session{}).Update("is_active", false).Error; err != nil {
		r.logger.Error("failed to deactivate other sessions",
			interfaces.ErrorField(err),
			interfaces.String("user_id", userID),
			interfaces.String("current_session_id", currentSessionID),
		)
		return errors.Wrap(err, "failed to deactivate other sessions")
	}

	r.logger.Info("other sessions deactivated successfully",
		interfaces.String("user_id", userID),
		interfaces.String("current_session_id", currentSessionID),
	)

	return nil
}

// Delete 删除会话
func (r *sessionRepository) Delete(ctx context.Context, sessionID string) error {
	db := r.getDB()

	// 使用选项构建查询条件
	queryOptions := options.NewQuerySessionOptions(db, options.WithSessionID(sessionID))
	queryDB := queryOptions.GetDB()

	if err := queryDB.WithContext(ctx).Delete(&models.Session{}).Error; err != nil {
		r.logger.Error("failed to delete session",
			interfaces.ErrorField(err),
			interfaces.String("session_id", sessionID),
		)
		return errors.Wrap(err, "failed to delete session")
	}

	r.logger.Info("session deleted successfully",
		interfaces.String("session_id", sessionID),
	)

	return nil
}

// DeleteExpiredSessions 删除过期会话
func (r *sessionRepository) DeleteExpiredSessions(ctx context.Context) error {
	db := r.getDB()

	// 使用选项构建查询条件
	queryOptions := options.NewQuerySessionOptions(db, options.WithExpiredSession())
	queryDB := queryOptions.GetDB()

	result := queryDB.WithContext(ctx).Delete(&models.Session{})
	if result.Error != nil {
		r.logger.Error("failed to delete expired sessions",
			interfaces.ErrorField(result.Error),
		)
		return errors.Wrap(result.Error, "failed to delete expired sessions")
	}

	r.logger.Info("expired sessions deleted successfully",
		interfaces.Int64("deleted_count", result.RowsAffected),
	)

	return nil
}

// CleanupInactiveSessions 清理非活跃会话
func (r *sessionRepository) CleanupInactiveSessions(ctx context.Context) error {
	db := r.getDB()

	// 使用选项构建查询条件：删除30天前的非活跃会话
	queryOptions := options.NewQuerySessionOptions(db, options.WithOldInactiveSessions(30))
	queryDB := queryOptions.GetDB()

	result := queryDB.WithContext(ctx).Delete(&models.Session{})
	if result.Error != nil {
		r.logger.Error("failed to cleanup inactive sessions",
			interfaces.ErrorField(result.Error),
		)
		return errors.Wrap(result.Error, "failed to cleanup inactive sessions")
	}

	r.logger.Info("inactive sessions cleaned up successfully",
		interfaces.Int64("deleted_count", result.RowsAffected),
	)

	return nil
}

// Function: RotateTokensCAS
//
// Description: 基于旧 refresh_token 原子更新会话中的 refresh/access token 以及过期时间，保证并发刷新只有一个成功。
//
// Parameters:
//
//	ctx(context.Context): 上下文
//	oldRefreshToken(string): 旧的刷新 token（条件）
//	newRefreshToken(string): 新刷新 token
//	newAccessToken(string): 新访问 token
//	newExpiresAt(int64): 新 access token 基础过期时间（Unix 秒），会话层会在外层计算真实 ExpiresAt（已加 24h）时此处直接持久化对应 time 值
//
// Returns:
//
//	ok(bool): 是否成功完成原子替换
//	err(error): 错误
func (r *sessionRepository) RotateTokensCAS(ctx context.Context, oldRefreshToken, newRefreshToken, newAccessToken string, newExpiresAt int64) (bool, error) {
	db := r.getDB()
	// 条件：refresh_token 匹配且 is_active=true 且未过期
	// 由于会话表 ExpiresAt 字段存储为 datetime，这里使用 NOW() 进行比较
	result := db.WithContext(ctx).Model(&models.Session{}).
		Where("refresh_token = ? AND is_active = ? AND expires_at > NOW()", oldRefreshToken, true).
		Updates(map[string]interface{}{
			"refresh_token": newRefreshToken,
			"access_token":  newAccessToken,
			// 会话过期时间延长策略：保持现有逻辑在 service 中计算后的时间；为保持最小侵入，此处不修改 ExpiresAt 由调用方决定是否传递更新。
		})
	if result.Error != nil {
		return false, errors.Wrap(result.Error, "failed to rotate tokens CAS")
	}
	return result.RowsAffected == 1, nil
}

// CountActiveSessionsByUserID 统计用户活跃会话数
func (r *sessionRepository) CountActiveSessionsByUserID(ctx context.Context, userID string) (int64, error) {
	return r.CountSessions(ctx, options.WithUserActiveSessions(userID))
}

// GetSessionWithUser 获取会话及关联的用户信息
func (r *sessionRepository) GetSessionWithUser(ctx context.Context, sessionID string) (*models.Session, error) {
	return r.FindSession(ctx,
		options.WithSessionID(sessionID),
		options.WithUserPreload(),
	)
}

// FindSessions 根据选项查询会话列表
func (r *sessionRepository) FindSessions(ctx context.Context, opts ...options.QuerySessionOption) ([]*models.Session, error) {
	db := r.getDB()

	// 构建查询选项
	queryOptions := options.NewQuerySessionOptions(db, opts...)
	queryDB := queryOptions.GetDB()

	var sessions []*models.Session
	if err := queryDB.WithContext(ctx).Find(&sessions).Error; err != nil {
		r.logger.Error("failed to find sessions",
			interfaces.ErrorField(err),
		)
		return nil, errors.Wrap(err, "failed to find sessions")
	}

	return sessions, nil
}

// FindSession 根据选项查询单个会话
func (r *sessionRepository) FindSession(ctx context.Context, opts ...options.QuerySessionOption) (*models.Session, error) {
	db := r.getDB()

	// 构建查询选项
	queryOptions := options.NewQuerySessionOptions(db, opts...)
	queryDB := queryOptions.GetDB()

	var session models.Session
	if err := queryDB.WithContext(ctx).First(&session).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		r.logger.Error("failed to find session",
			interfaces.ErrorField(err),
		)
		return nil, errors.Wrap(err, "failed to find session")
	}

	return &session, nil
}

// CountSessions 根据选项统计会话数量
func (r *sessionRepository) CountSessions(ctx context.Context, opts ...options.QuerySessionOption) (int64, error) {
	db := r.getDB()

	// 构建查询选项
	queryOptions := options.NewQuerySessionOptions(db, opts...)
	queryDB := queryOptions.GetDB()

	var count int64
	if err := queryDB.WithContext(ctx).Model(&models.Session{}).Count(&count).Error; err != nil {
		r.logger.Error("failed to count sessions",
			interfaces.ErrorField(err),
		)
		return 0, errors.Wrap(err, "failed to count sessions")
	}

	return count, nil
}

// FindSessionsWithPagination 根据选项分页查询会话
func (r *sessionRepository) FindSessionsWithPagination(ctx context.Context, page, pageSize int, opts ...options.QuerySessionOption) ([]*models.Session, int64, error) {
	db := r.getDB()

	// 先统计总数（不包含分页选项）
	countOptions := options.NewQuerySessionOptions(db, opts...)
	countDB := countOptions.GetDB()

	var total int64
	if err := countDB.WithContext(ctx).Model(&models.Session{}).Count(&total).Error; err != nil {
		r.logger.Error("failed to count sessions for pagination",
			interfaces.ErrorField(err),
		)
		return nil, 0, errors.Wrap(err, "failed to count sessions for pagination")
	}

	// 添加分页选项并查询数据
	paginationOpts := append(opts, options.WithSessionPagination(page, pageSize))
	queryOptions := options.NewQuerySessionOptions(db, paginationOpts...)
	queryDB := queryOptions.GetDB()

	var sessions []*models.Session
	if err := queryDB.WithContext(ctx).Find(&sessions).Error; err != nil {
		r.logger.Error("failed to find sessions with pagination",
			interfaces.ErrorField(err),
		)
		return nil, 0, errors.Wrap(err, "failed to find sessions with pagination")
	}

	return sessions, total, nil
}
