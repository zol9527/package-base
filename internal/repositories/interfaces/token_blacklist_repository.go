package interfaces

import (
	"context"
	"time"
)

// Function: TokenBlacklistRepository
//
// Description: Token 黑名单仓储接口，提供加入、查询、移除及清理等操作，确保失效 Token 无法继续使用。
//
// Parameters:
//
//	无
//
// Returns:
//
//	无
//
// TokenBlacklistRepository Token黑名单存储接口
type TokenBlacklistRepository interface {
	// AddToBlacklist 将token添加到黑名单
	AddToBlacklist(ctx context.Context, token string, expiration time.Duration) error

	// IsBlacklisted 检查token是否在黑名单中
	IsBlacklisted(ctx context.Context, token string) (bool, error)

	// RemoveFromBlacklist 从黑名单中移除token（用于测试或管理）
	RemoveFromBlacklist(ctx context.Context, token string) error

	// ClearExpiredTokens 清理过期的黑名单token
	ClearExpiredTokens(ctx context.Context) error

	// GetBlacklistSize 获取黑名单大小
	GetBlacklistSize(ctx context.Context) (int64, error)

	// GetBlacklistedTokens 获取黑名单中的token列表（用于管理）
	GetBlacklistedTokens(ctx context.Context, limit int) ([]string, error)
}
