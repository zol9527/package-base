package interfaces

import (
	"context"
	"package-base/internal/models"
	"package-base/internal/options"
)

// Function: LoginAttemptRepository
//
// Description: 登录尝试记录的仓储接口，提供审计统计与风控所需的查询与聚合能力。
//
// Parameters:
//
//	无
//
// Returns:
//
//	无
//
// LoginAttemptRepository 登录尝试存储接口
type LoginAttemptRepository interface {
	// Create 创建登录尝试记录
	Create(ctx context.Context, attempt *models.LoginAttempt) error

	// GetByID 根据ID获取登录尝试记录
	GetByID(ctx context.Context, id uint) (*models.LoginAttempt, error)

	// Update 更新登录尝试记录
	Update(ctx context.Context, attempt *models.LoginAttempt) error

	// Delete 删除登录尝试记录
	Delete(ctx context.Context, id uint) error

	// 使用查询选项的方法

	// FindLoginAttempts 根据选项查询登录尝试列表
	FindLoginAttempts(ctx context.Context, opts ...options.QueryLoginAttemptOption) ([]*models.LoginAttempt, error)

	// FindLoginAttempt 根据选项查询单个登录尝试
	FindLoginAttempt(ctx context.Context, opts ...options.QueryLoginAttemptOption) (*models.LoginAttempt, error)

	// CountLoginAttempts 根据选项统计登录尝试数量
	CountLoginAttempts(ctx context.Context, opts ...options.QueryLoginAttemptOption) (int64, error)

	// FindLoginAttemptsWithPagination 根据选项分页查询登录尝试
	FindLoginAttemptsWithPagination(ctx context.Context, page, pageSize int, opts ...options.QueryLoginAttemptOption) ([]*models.LoginAttempt, int64, error)

	// 特定业务方法

	// GetRecentFailedAttemptsByIP 获取IP最近的失败尝试次数
	GetRecentFailedAttemptsByIP(ctx context.Context, ipAddress string, minutes int) (int64, error)

	// GetRecentFailedAttemptsByUser 获取用户最近的失败尝试次数
	GetRecentFailedAttemptsByUser(ctx context.Context, userID string, minutes int) (int64, error)

	// GetRecentFailedAttemptsByIdentity 获取登录标识最近的失败尝试次数
	GetRecentFailedAttemptsByIdentity(ctx context.Context, identity string, minutes int) (int64, error)

	// GetIPAttemptStats 获取IP的尝试统计
	GetIPAttemptStats(ctx context.Context, ipAddress string) (map[string]int64, error)

	// GetUserAttemptStats 获取用户的尝试统计
	GetUserAttemptStats(ctx context.Context, userID string) (map[string]int64, error)

	// CleanupOldAttempts 清理过期的登录尝试记录
	CleanupOldAttempts(ctx context.Context, daysToKeep int) error

	// GetTopFailedIPs 获取失败尝试最多的IP列表
	GetTopFailedIPs(ctx context.Context, limit int, hours int) ([]map[string]interface{}, error)

	// GetFailedAttemptsTrend 获取失败尝试趋势数据
	GetFailedAttemptsTrend(ctx context.Context, hours int) ([]map[string]interface{}, error)
}
