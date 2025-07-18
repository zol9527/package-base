package repositories

import (
	"context"
	"fmt"
	"time"

	"package-base/internal/core/interfaces"
	repoInterfaces "package-base/internal/repositories/interfaces"

	"github.com/pkg/errors"
)

const (
	// Redis key prefixes for token blacklist
	tokenBlacklistKeyPrefix = "auth:blacklist:tokens"
	tokenBlacklistListKey   = "auth:blacklist:token_list"
)

// tokenBlacklistRepository Token黑名单存储实现
type tokenBlacklistRepository struct {
	redis  interfaces.KeyValueStorage
	logger interfaces.Logger
}

// NewTokenBlacklistRepository 创建Token黑名单存储实现
func NewTokenBlacklistRepository(
	redis interfaces.KeyValueStorage,
	logger interfaces.Logger,
) repoInterfaces.TokenBlacklistRepository {
	return &tokenBlacklistRepository{
		redis:  redis,
		logger: logger.Named("token_blacklist_repository"),
	}
}

// AddToBlacklist 将token添加到黑名单
func (r *tokenBlacklistRepository) AddToBlacklist(ctx context.Context, token string, expiration time.Duration) error {
	// 使用token作为key，设置过期时间
	tokenKey := fmt.Sprintf("%s:%s", tokenBlacklistKeyPrefix, token)

	if err := r.redis.Set(ctx, tokenKey, "blacklisted", expiration); err != nil {
		r.logger.Error("failed to add token to blacklist",
			interfaces.ErrorField(err),
			interfaces.String("token_key", tokenKey),
		)
		return errors.Wrap(err, "failed to add token to blacklist")
	}

	r.logger.Info("token added to blacklist",
		interfaces.String("token_key", tokenKey),
		interfaces.String("expiration", expiration.String()),
	)

	return nil
}

// IsBlacklisted 检查token是否在黑名单中
func (r *tokenBlacklistRepository) IsBlacklisted(ctx context.Context, token string) (bool, error) {
	tokenKey := fmt.Sprintf("%s:%s", tokenBlacklistKeyPrefix, token)

	exists, err := r.redis.Exists(ctx, tokenKey)
	if err != nil {
		r.logger.Error("failed to check token blacklist status",
			interfaces.ErrorField(err),
			interfaces.String("token_key", tokenKey),
		)
		return false, errors.Wrap(err, "failed to check token blacklist status")
	}

	r.logger.Debug("checked token blacklist status",
		interfaces.String("token_key", tokenKey),
		interfaces.Bool("is_blacklisted", exists),
	)

	return exists, nil
}

// RemoveFromBlacklist 从黑名单中移除token
func (r *tokenBlacklistRepository) RemoveFromBlacklist(ctx context.Context, token string) error {
	tokenKey := fmt.Sprintf("%s:%s", tokenBlacklistKeyPrefix, token)

	if err := r.redis.Delete(ctx, tokenKey); err != nil {
		r.logger.Error("failed to remove token from blacklist",
			interfaces.ErrorField(err),
			interfaces.String("token_key", tokenKey),
		)
		return errors.Wrap(err, "failed to remove token from blacklist")
	}

	r.logger.Info("token removed from blacklist",
		interfaces.String("token_key", tokenKey),
	)

	return nil
}

// ClearExpiredTokens 清理过期的黑名单token
func (r *tokenBlacklistRepository) ClearExpiredTokens(ctx context.Context) error {
	// 获取所有黑名单相关的key
	pattern := fmt.Sprintf("%s:*", tokenBlacklistKeyPrefix)
	keys, err := r.redis.Keys(ctx, pattern)
	if err != nil {
		r.logger.Error("failed to get blacklist keys for cleanup",
			interfaces.ErrorField(err),
		)
		return errors.Wrap(err, "failed to get blacklist keys for cleanup")
	}

	cleanedCount := 0
	for _, key := range keys {
		exists, err := r.redis.Exists(ctx, key)
		if err != nil {
			r.logger.Warn("failed to check key existence during cleanup",
				interfaces.ErrorField(err),
				interfaces.String("key", key),
			)
			continue
		}

		// 如果key不存在，说明已过期，计数器增加
		if !exists {
			cleanedCount++
		}
	}

	r.logger.Info("cleaned expired tokens from blacklist",
		interfaces.Int("cleaned_count", cleanedCount),
		interfaces.Int("total_checked", len(keys)),
	)

	return nil
}

// GetBlacklistSize 获取黑名单大小
func (r *tokenBlacklistRepository) GetBlacklistSize(ctx context.Context) (int64, error) {
	pattern := fmt.Sprintf("%s:*", tokenBlacklistKeyPrefix)
	keys, err := r.redis.Keys(ctx, pattern)
	if err != nil {
		r.logger.Error("failed to get blacklist size",
			interfaces.ErrorField(err),
		)
		return 0, errors.Wrap(err, "failed to get blacklist size")
	}

	return int64(len(keys)), nil
}

// GetBlacklistedTokens 获取黑名单中的token列表
func (r *tokenBlacklistRepository) GetBlacklistedTokens(ctx context.Context, limit int) ([]string, error) {
	if limit <= 0 {
		limit = 100 // 默认限制
	}

	pattern := fmt.Sprintf("%s:*", tokenBlacklistKeyPrefix)
	keys, err := r.redis.Keys(ctx, pattern)
	if err != nil {
		r.logger.Error("failed to get blacklisted tokens",
			interfaces.ErrorField(err),
		)
		return nil, errors.Wrap(err, "failed to get blacklisted tokens")
	}

	// 提取token部分（去掉前缀）
	tokens := make([]string, 0, len(keys))
	prefixLen := len(tokenBlacklistKeyPrefix) + 1 // +1 for ':'

	for _, key := range keys {
		if len(key) > prefixLen {
			token := key[prefixLen:]
			tokens = append(tokens, token)
		}
	}

	// 限制返回数量
	if len(tokens) > limit {
		tokens = tokens[:limit]
	}

	r.logger.Debug("retrieved blacklisted tokens",
		interfaces.Int("count", len(tokens)),
		interfaces.Int("limit", limit),
	)

	return tokens, nil
}
