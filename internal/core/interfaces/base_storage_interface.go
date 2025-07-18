package interfaces

import (
	"context"
	"time"
)

// BaseStorage 基础存储接口
//
// 描述: 定义所有存储实现必须支持的基础操作
type BaseStorage interface {
	// Connect 连接到存储
	Connect(ctx context.Context) error

	// Close 关闭存储连接
	Close() error

	// Ping 测试连接状态
	Ping(ctx context.Context) error

	// IsHealthy 检查存储健康状态
	IsHealthy(ctx context.Context) bool
}

// KeyValueStorage 键值存储接口
//
// 描述: 定义键值存储操作，适用于缓存、会话等场景
type KeyValueStorage interface {
	BaseStorage

	// Set 设置键值对
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error

	// Get 获取值
	Get(ctx context.Context, key string) (interface{}, error)

	// GetString 获取字符串值
	GetString(ctx context.Context, key string) (string, error)

	// Delete 删除键
	Delete(ctx context.Context, key string) error

	// Exists 检查键是否存在
	Exists(ctx context.Context, key string) (bool, error)

	// SetExpiration 设置键的过期时间
	SetExpiration(ctx context.Context, key string, expiration time.Duration) error

	// GetTTL 获取键的剩余生存时间
	GetTTL(ctx context.Context, key string) (time.Duration, error)

	// Keys 获取匹配模式的所有键
	Keys(ctx context.Context, pattern string) ([]string, error)

	// Clear 清空所有数据
	Clear(ctx context.Context) error
}
