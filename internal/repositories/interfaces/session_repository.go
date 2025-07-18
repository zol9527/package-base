package interfaces

import (
	"context"
	"package-base/internal/models"
	"package-base/internal/options"
)

// SessionRepository 会话存储接口
type SessionRepository interface {
	// Create 创建会话
	Create(ctx context.Context, session *models.Session) error

	// GetByID 根据ID获取会话
	GetByID(ctx context.Context, id string) (*models.Session, error)

	// GetByRefreshToken 根据刷新Token获取会话
	GetByRefreshToken(ctx context.Context, refreshToken string) (*models.Session, error)

	// GetByUserID 获取用户的所有会话
	GetByUserID(ctx context.Context, userID string) ([]*models.Session, error)

	// GetActiveByUserID 获取用户的活跃会话
	GetActiveByUserID(ctx context.Context, userID string) ([]*models.Session, error)

	// Update 更新会话
	Update(ctx context.Context, session *models.Session) error

	// UpdateAccessToken 更新访问Token
	UpdateAccessToken(ctx context.Context, sessionID, accessToken string) error

	// DeactivateSession 停用指定会话
	DeactivateSession(ctx context.Context, sessionID string) error

	// DeactivateUserSessions 停用用户的所有会话
	DeactivateUserSessions(ctx context.Context, userID string) error

	// DeactivateOtherSessions 停用用户的其他会话（除了指定的会话）
	DeactivateOtherSessions(ctx context.Context, userID, currentSessionID string) error

	// Delete 删除会话
	Delete(ctx context.Context, sessionID string) error

	// DeleteExpiredSessions 删除过期会话
	DeleteExpiredSessions(ctx context.Context) error

	// CleanupInactiveSessions 清理非活跃会话
	CleanupInactiveSessions(ctx context.Context) error

	// CountActiveSessionsByUserID 统计用户活跃会话数
	CountActiveSessionsByUserID(ctx context.Context, userID string) (int64, error)

	// GetSessionWithUser 获取会话及关联的用户信息
	GetSessionWithUser(ctx context.Context, sessionID string) (*models.Session, error)

	// FindSessions 根据选项查询会话列表
	FindSessions(ctx context.Context, opts ...options.QuerySessionOption) ([]*models.Session, error)

	// FindSession 根据选项查询单个会话
	FindSession(ctx context.Context, opts ...options.QuerySessionOption) (*models.Session, error)

	// CountSessions 根据选项统计会话数量
	CountSessions(ctx context.Context, opts ...options.QuerySessionOption) (int64, error)

	// FindSessionsWithPagination 根据选项分页查询会话
	FindSessionsWithPagination(ctx context.Context, page, pageSize int, opts ...options.QuerySessionOption) ([]*models.Session, int64, error)
}
