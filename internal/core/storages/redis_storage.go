package storages

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"package-base/internal/core/interfaces"

	"github.com/duke-git/lancet/v2/convertor"
	"github.com/go-redis/redis/v8"
	"github.com/pkg/errors"
)

// RedisConfig Redis配置结构
//
// 描述: Redis连接配置
type RedisConfig struct {
	Host          string        `json:"host"`            // 主机地址
	Port          int           `json:"port"`            // 端口号
	Username      string        `json:"username"`        // 用户名
	Password      string        `json:"password"`        // 密码
	Database      int           `json:"database"`        // 数据库索引
	MaxRetries    int           `json:"max_retries"`     // 最大重试次数
	PoolSize      int           `json:"pool_size"`       // 连接池大小
	MinIdleConns  int           `json:"min_idle_conns"`  // 最小空闲连接数
	DialTimeout   time.Duration `json:"dial_timeout"`    // 连接超时
	ReadTimeout   time.Duration `json:"read_timeout"`    // 读取超时
	WriteTimeout  time.Duration `json:"write_timeout"`   // 写入超时
	PoolTimeout   time.Duration `json:"pool_timeout"`    // 连接池超时
	IdleTimeout   time.Duration `json:"idle_timeout"`    // 空闲连接超时
	IdleCheckFreq time.Duration `json:"idle_check_freq"` // 空闲检查频率
	MaxConnAge    time.Duration `json:"max_conn_age"`    // 连接最大存活时间
	KeyPrefix     string        `json:"key_prefix"`      // 键前缀
	DefaultTTL    time.Duration `json:"default_ttl"`     // 默认TTL
}

// GetType 获取存储类型
//
// 描述: 返回存储类型标识
//
// 返回:
//
//	string: 存储类型
func (c *RedisConfig) GetType() string {
	return "redis"
}

// Validate 验证配置
//
// 描述: 验证Redis配置的有效性
//
// 返回:
//
//	error: 验证错误信息
func (c *RedisConfig) Validate() error {
	if c.Host == "" {
		return errors.New("🚫 Redis主机地址不能为空")
	}
	if c.Port <= 0 || c.Port > 65535 {
		return errors.New("🚫 Redis端口号必须在1-65535之间")
	}
	if c.Database < 0 || c.Database > 15 {
		return errors.New("🚫 Redis数据库索引必须在0-15之间")
	}

	// 设置默认值
	if c.MaxRetries <= 0 {
		c.MaxRetries = 3
	}
	if c.PoolSize <= 0 {
		c.PoolSize = 10
	}
	if c.MinIdleConns <= 0 {
		c.MinIdleConns = 1
	}
	if c.DialTimeout <= 0 {
		c.DialTimeout = 5 * time.Second
	}
	if c.ReadTimeout <= 0 {
		c.ReadTimeout = 3 * time.Second
	}
	if c.WriteTimeout <= 0 {
		c.WriteTimeout = 3 * time.Second
	}
	if c.PoolTimeout <= 0 {
		c.PoolTimeout = 4 * time.Second
	}
	if c.IdleTimeout <= 0 {
		c.IdleTimeout = 5 * time.Minute
	}
	if c.IdleCheckFreq <= 0 {
		c.IdleCheckFreq = time.Minute
	}
	if c.MaxConnAge <= 0 {
		c.MaxConnAge = 30 * time.Minute
	}
	if c.DefaultTTL <= 0 {
		c.DefaultTTL = 24 * time.Hour
	}

	return nil
}

// NormalizeDuration 单位归一化
//
// 描述: 将 config.json 解析出来的秒数参数转为 time.Duration
//
// 返回:
//
//	无
func (c *RedisConfig) NormalizeDuration() {
	c.DialTimeout = c.DialTimeout * time.Second
	c.ReadTimeout = c.ReadTimeout * time.Second
	c.WriteTimeout = c.WriteTimeout * time.Second
	c.PoolTimeout = c.PoolTimeout * time.Second
	c.IdleTimeout = c.IdleTimeout * time.Second
	c.IdleCheckFreq = c.IdleCheckFreq * time.Second
	c.MaxConnAge = c.MaxConnAge * time.Second
	c.DefaultTTL = c.DefaultTTL * time.Second
}

// GetConnectionString 获取连接字符串
//
// 描述: 构建Redis连接字符串
//
// 返回:
//
//	string: 连接字符串
func (c *RedisConfig) GetConnectionString() string {
	return fmt.Sprintf("redis://:%s@%s:%d/%d", c.Password, c.Host, c.Port, c.Database)
}

// RedisStorage Redis存储实现
//
// 描述: 基于go-redis的键值存储和缓存存储实现
type RedisStorage struct {
	client *redis.Client
	config *RedisConfig
}

// NewRedisStorage 创建Redis存储实例
//
// 描述: 根据配置创建Redis存储实例
//
// 参数:
//
//	config (*RedisConfig): Redis配置
//
// 返回:
//
//	interfaces.CacheStorage: 缓存存储接口
//	error: 错误信息
func NewRedisStorage(config *RedisConfig) (interfaces.CacheStorage, error) {
	if config == nil {
		return nil, errors.New("🚫 Redis配置不能为空")
	}

	config.NormalizeDuration() // 单位归一化

	if err := config.Validate(); err != nil {
		return nil, errors.Wrap(err, "🚫 Redis配置验证失败")
	}

	storage := &RedisStorage{
		config: config,
	}

	// 创建连接
	if err := storage.Connect(context.Background()); err != nil {
		return nil, errors.Wrap(err, "🚫 创建Redis连接失败")
	}

	return storage, nil
}

// Connect 连接到Redis
//
// 描述: 建立到Redis服务器的连接
//
// 参数:
//
//	ctx (context.Context): 上下文
//
// 返回:
//
//	error: 错误信息
func (s *RedisStorage) Connect(ctx context.Context) error {
	if s.client != nil {
		return nil
	}

	// 创建Redis客户端配置
	opts := &redis.Options{
		Addr:         fmt.Sprintf("%s:%d", s.config.Host, s.config.Port),
		Username:     s.config.Username,
		Password:     s.config.Password,
		DB:           s.config.Database,
		MaxRetries:   s.config.MaxRetries,
		PoolSize:     s.config.PoolSize,
		MinIdleConns: s.config.MinIdleConns,
		DialTimeout:  s.config.DialTimeout,
		ReadTimeout:  s.config.ReadTimeout,
		WriteTimeout: s.config.WriteTimeout,
		PoolTimeout:  s.config.PoolTimeout,
		IdleTimeout:  s.config.IdleTimeout,

		MaxConnAge: s.config.MaxConnAge,
	}

	// 创建客户端
	client := redis.NewClient(opts)

	// 测试连接
	pingCtx, cancel := context.WithTimeout(ctx, s.config.DialTimeout)
	defer cancel()

	if err := client.Ping(pingCtx).Err(); err != nil {
		client.Close()
		return errors.Wrap(err, "🚫 Redis连接测试失败")
	}

	s.client = client
	return nil
}

// Close 关闭连接
//
// 描述: 关闭Redis连接
//
// 返回:
//
//	error: 错误信息
func (s *RedisStorage) Close() error {
	if s.client == nil {
		return nil
	}

	if err := s.client.Close(); err != nil {
		return errors.Wrap(err, "🚫 关闭Redis连接失败")
	}

	s.client = nil
	return nil
}

// Ping 测试连接状态
//
// 描述: 测试Redis连接是否正常
//
// 参数:
//
//	ctx (context.Context): 上下文
//
// 返回:
//
//	error: 错误信息
func (s *RedisStorage) Ping(ctx context.Context) error {
	if s.client == nil {
		return errors.New("🚫 Redis连接未建立")
	}

	pingCtx, cancel := context.WithTimeout(ctx, s.config.ReadTimeout)
	defer cancel()

	if err := s.client.Ping(pingCtx).Err(); err != nil {
		return errors.Wrap(err, "🚫 Redis ping失败")
	}

	return nil
}

// IsHealthy 检查存储健康状态
//
// 描述: 检查Redis存储是否健康
//
// 参数:
//
//	ctx (context.Context): 上下文
//
// 返回:
//
//	bool: 是否健康
func (s *RedisStorage) IsHealthy(ctx context.Context) bool {
	return s.Ping(ctx) == nil
}

// buildKey 构建键名
//
// 描述: 根据配置的前缀构建完整键名
//
// 参数:
//
//	key (string): 原始键名
//
// 返回:
//
//	string: 完整键名
func (s *RedisStorage) buildKey(key string) string {
	if s.config.KeyPrefix == "" {
		return key
	}
	return fmt.Sprintf("%s:%s", s.config.KeyPrefix, key)
}

// Set 设置键值对
//
// 描述: 设置键值对，支持过期时间
//
// 参数:
//
//	ctx (context.Context): 上下文
//	key (string): 键名
//	value (interface{}): 值
//	expiration (time.Duration): 过期时间，0表示不过期
//
// 返回:
//
//	error: 错误信息
func (s *RedisStorage) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	if s.client == nil {
		return errors.New("🚫 Redis连接未建立")
	}

	if key == "" {
		return errors.New("🚫 键名不能为空")
	}

	// 序列化值
	data, err := s.serializeValue(value)
	if err != nil {
		return errors.Wrap(err, "🚫 序列化值失败")
	}

	// 如果没有指定过期时间，使用默认TTL
	if expiration == 0 {
		expiration = s.config.DefaultTTL
	}

	fullKey := s.buildKey(key)
	if err := s.client.Set(ctx, fullKey, data, expiration).Err(); err != nil {
		return errors.Wrap(err, "🚫 设置键值对失败")
	}

	return nil
}

// Get 获取值
//
// 描述: 根据键名获取值
//
// 参数:
//
//	ctx (context.Context): 上下文
//	key (string): 键名
//
// 返回:
//
//	interface{}: 值
//	error: 错误信息
func (s *RedisStorage) Get(ctx context.Context, key string) (interface{}, error) {
	if s.client == nil {
		return nil, errors.New("🚫 Redis连接未建立")
	}

	if key == "" {
		return nil, errors.New("🚫 键名不能为空")
	}

	fullKey := s.buildKey(key)
	data, err := s.client.Get(ctx, fullKey).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, errors.New("🔍 键不存在")
		}
		return nil, errors.Wrap(err, "🚫 获取值失败")
	}

	// 反序列化值
	value, err := s.deserializeValue(data)
	if err != nil {
		return nil, errors.Wrap(err, "🚫 反序列化值失败")
	}

	return value, nil
}

// GetString 获取字符串值
//
// 描述: 根据键名获取字符串值
//
// 参数:
//
//	ctx (context.Context): 上下文
//	key (string): 键名
//
// 返回:
//
//	string: 字符串值
//	error: 错误信息
func (s *RedisStorage) GetString(ctx context.Context, key string) (string, error) {
	if s.client == nil {
		return "", errors.New("🚫 Redis连接未建立")
	}

	if key == "" {
		return "", errors.New("🚫 键名不能为空")
	}

	fullKey := s.buildKey(key)
	result, err := s.client.Get(ctx, fullKey).Result()
	if err != nil {
		if err == redis.Nil {
			return "", errors.New("🔍 键不存在")
		}
		return "", errors.Wrap(err, "🚫 获取字符串值失败")
	}

	return result, nil
}

// Delete 删除键
//
// 描述: 删除指定的键
//
// 参数:
//
//	ctx (context.Context): 上下文
//	key (string): 键名
//
// 返回:
//
//	error: 错误信息
func (s *RedisStorage) Delete(ctx context.Context, key string) error {
	if s.client == nil {
		return errors.New("🚫 Redis连接未建立")
	}

	if key == "" {
		return errors.New("🚫 键名不能为空")
	}

	fullKey := s.buildKey(key)
	if err := s.client.Del(ctx, fullKey).Err(); err != nil {
		return errors.Wrap(err, "🚫 删除键失败")
	}

	return nil
}

// Exists 检查键是否存在
//
// 描述: 检查指定的键是否存在
//
// 参数:
//
//	ctx (context.Context): 上下文
//	key (string): 键名
//
// 返回:
//
//	bool: 是否存在
//	error: 错误信息
func (s *RedisStorage) Exists(ctx context.Context, key string) (bool, error) {
	if s.client == nil {
		return false, errors.New("🚫 Redis连接未建立")
	}

	if key == "" {
		return false, errors.New("🚫 键名不能为空")
	}

	fullKey := s.buildKey(key)
	result, err := s.client.Exists(ctx, fullKey).Result()
	if err != nil {
		return false, errors.Wrap(err, "🚫 检查键是否存在失败")
	}

	return result > 0, nil
}

// SetExpiration 设置键的过期时间
//
// 描述: 为已存在的键设置过期时间
//
// 参数:
//
//	ctx (context.Context): 上下文
//	key (string): 键名
//	expiration (time.Duration): 过期时间
//
// 返回:
//
//	error: 错误信息
func (s *RedisStorage) SetExpiration(ctx context.Context, key string, expiration time.Duration) error {
	if s.client == nil {
		return errors.New("🚫 Redis连接未建立")
	}

	if key == "" {
		return errors.New("🚫 键名不能为空")
	}

	fullKey := s.buildKey(key)
	if err := s.client.Expire(ctx, fullKey, expiration).Err(); err != nil {
		return errors.Wrap(err, "🚫 设置键过期时间失败")
	}

	return nil
}

// GetTTL 获取键的剩余生存时间
//
// 描述: 获取键的剩余生存时间
//
// 参数:
//
//	ctx (context.Context): 上下文
//	key (string): 键名
//
// 返回:
//
//	time.Duration: 剩余生存时间
//	error: 错误信息
func (s *RedisStorage) GetTTL(ctx context.Context, key string) (time.Duration, error) {
	if s.client == nil {
		return 0, errors.New("🚫 Redis连接未建立")
	}

	if key == "" {
		return 0, errors.New("🚫 键名不能为空")
	}

	fullKey := s.buildKey(key)
	result, err := s.client.TTL(ctx, fullKey).Result()
	if err != nil {
		return 0, errors.Wrap(err, "🚫 获取键TTL失败")
	}

	return result, nil
}

// Keys 获取匹配模式的所有键
//
// 描述: 获取匹配指定模式的所有键
//
// 参数:
//
//	ctx (context.Context): 上下文
//	pattern (string): 匹配模式
//
// 返回:
//
//	[]string: 键列表
//	error: 错误信息
func (s *RedisStorage) Keys(ctx context.Context, pattern string) ([]string, error) {
	if s.client == nil {
		return nil, errors.New("🚫 Redis连接未建立")
	}

	if pattern == "" {
		pattern = "*"
	}

	fullPattern := s.buildKey(pattern)
	result, err := s.client.Keys(ctx, fullPattern).Result()
	if err != nil {
		return nil, errors.Wrap(err, "🚫 获取键列表失败")
	}

	// 移除键前缀
	if s.config.KeyPrefix != "" {
		prefixLen := len(s.config.KeyPrefix) + 1 // +1 for ":"
		for i, key := range result {
			if len(key) > prefixLen {
				result[i] = key[prefixLen:]
			}
		}
	}

	return result, nil
}

// Clear 清空所有数据
//
// 描述: 清空当前数据库的所有数据
//
// 参数:
//
//	ctx (context.Context): 上下文
//
// 返回:
//
//	error: 错误信息
func (s *RedisStorage) Clear(ctx context.Context) error {
	if s.client == nil {
		return errors.New("🚫 Redis连接未建立")
	}

	if err := s.client.FlushDB(ctx).Err(); err != nil {
		return errors.Wrap(err, "🚫 清空数据库失败")
	}

	return nil
}

// SetWithTags 设置带标签的缓存
//
// 描述: 设置键值对并关联标签
//
// 参数:
//
//	ctx (context.Context): 上下文
//	key (string): 键名
//	value (interface{}): 值
//	expiration (time.Duration): 过期时间
//	tags ([]string): 标签列表
//
// 返回:
//
//	error: 错误信息
func (s *RedisStorage) SetWithTags(ctx context.Context, key string, value interface{}, expiration time.Duration, tags []string) error {
	// 设置键值对
	if err := s.Set(ctx, key, value, expiration); err != nil {
		return err
	}

	// 如果有标签，建立标签到键的映射
	if len(tags) > 0 {
		for _, tag := range tags {
			tagKey := s.buildKey(fmt.Sprintf("tag:%s", tag))
			if err := s.client.SAdd(ctx, tagKey, key).Err(); err != nil {
				return errors.Wrap(err, "🚫 添加标签映射失败")
			}
			// 设置标签键的过期时间为值的过期时间
			if expiration > 0 {
				s.client.Expire(ctx, tagKey, expiration)
			}
		}
	}

	return nil
}

// DeleteByTags 根据标签删除缓存
//
// 描述: 删除所有关联指定标签的键
//
// 参数:
//
//	ctx (context.Context): 上下文
//	tags ([]string): 标签列表
//
// 返回:
//
//	error: 错误信息
func (s *RedisStorage) DeleteByTags(ctx context.Context, tags []string) error {
	if s.client == nil {
		return errors.New("🚫 Redis连接未建立")
	}

	if len(tags) == 0 {
		return errors.New("🚫 标签列表不能为空")
	}

	for _, tag := range tags {
		tagKey := s.buildKey(fmt.Sprintf("tag:%s", tag))

		// 获取标签关联的所有键
		keys, err := s.client.SMembers(ctx, tagKey).Result()
		if err != nil {
			return errors.Wrap(err, "🚫 获取标签关联的键失败")
		}

		// 删除关联的键
		if len(keys) > 0 {
			fullKeys := make([]string, len(keys))
			for i, key := range keys {
				fullKeys[i] = s.buildKey(key)
			}
			if err := s.client.Del(ctx, fullKeys...).Err(); err != nil {
				return errors.Wrap(err, "🚫 删除标签关联的键失败")
			}
		}

		// 删除标签键
		if err := s.client.Del(ctx, tagKey).Err(); err != nil {
			return errors.Wrap(err, "🚫 删除标签键失败")
		}
	}

	return nil
}

// GetMulti 批量获取
//
// 描述: 批量获取多个键的值
//
// 参数:
//
//	ctx (context.Context): 上下文
//	keys ([]string): 键列表
//
// 返回:
//
//	map[string]interface{}: 键值映射
//	error: 错误信息
func (s *RedisStorage) GetMulti(ctx context.Context, keys []string) (map[string]interface{}, error) {
	if s.client == nil {
		return nil, errors.New("🚫 Redis连接未建立")
	}

	if len(keys) == 0 {
		return make(map[string]interface{}), nil
	}

	// 构建完整键名
	fullKeys := make([]string, len(keys))
	for i, key := range keys {
		fullKeys[i] = s.buildKey(key)
	}

	// 批量获取
	result, err := s.client.MGet(ctx, fullKeys...).Result()
	if err != nil {
		return nil, errors.Wrap(err, "🚫 批量获取值失败")
	}

	// 构建结果映射
	resultMap := make(map[string]interface{})
	for i, value := range result {
		if value != nil {
			if strValue, ok := value.(string); ok {
				deserializedValue, err := s.deserializeValue(strValue)
				if err != nil {
					resultMap[keys[i]] = strValue // 如果反序列化失败，返回原始字符串
				} else {
					resultMap[keys[i]] = deserializedValue
				}
			} else {
				resultMap[keys[i]] = value
			}
		}
	}

	return resultMap, nil
}

// SetMulti 批量设置
//
// 描述: 批量设置多个键值对
//
// 参数:
//
//	ctx (context.Context): 上下文
//	items (map[string]interface{}): 键值映射
//	expiration (time.Duration): 过期时间
//
// 返回:
//
//	error: 错误信息
func (s *RedisStorage) SetMulti(ctx context.Context, items map[string]interface{}, expiration time.Duration) error {
	if s.client == nil {
		return errors.New("🚫 Redis连接未建立")
	}

	if len(items) == 0 {
		return nil
	}

	// 使用事务确保原子性
	pipe := s.client.TxPipeline()

	for key, value := range items {
		data, err := s.serializeValue(value)
		if err != nil {
			return errors.Wrap(err, fmt.Sprintf("🚫 序列化键 %s 的值失败", key))
		}

		fullKey := s.buildKey(key)
		pipe.Set(ctx, fullKey, data, expiration)
	}

	if _, err := pipe.Exec(ctx); err != nil {
		return errors.Wrap(err, "🚫 批量设置键值对失败")
	}

	return nil
}

// DeleteMulti 批量删除
//
// 描述: 批量删除多个键
//
// 参数:
//
//	ctx (context.Context): 上下文
//	keys ([]string): 键列表
//
// 返回:
//
//	error: 错误信息
func (s *RedisStorage) DeleteMulti(ctx context.Context, keys []string) error {
	if s.client == nil {
		return errors.New("🚫 Redis连接未建立")
	}

	if len(keys) == 0 {
		return nil
	}

	// 构建完整键名
	fullKeys := make([]string, len(keys))
	for i, key := range keys {
		fullKeys[i] = s.buildKey(key)
	}

	if err := s.client.Del(ctx, fullKeys...).Err(); err != nil {
		return errors.Wrap(err, "🚫 批量删除键失败")
	}

	return nil
}

// Increment 递增
//
// 描述: 对键的值进行递增操作
//
// 参数:
//
//	ctx (context.Context): 上下文
//	key (string): 键名
//	delta (int64): 递增量
//
// 返回:
//
//	int64: 递增后的值
//	error: 错误信息
func (s *RedisStorage) Increment(ctx context.Context, key string, delta int64) (int64, error) {
	if s.client == nil {
		return 0, errors.New("🚫 Redis连接未建立")
	}

	if key == "" {
		return 0, errors.New("🚫 键名不能为空")
	}

	fullKey := s.buildKey(key)
	result, err := s.client.IncrBy(ctx, fullKey, delta).Result()
	if err != nil {
		return 0, errors.Wrap(err, "🚫 递增操作失败")
	}

	return result, nil
}

// Decrement 递减
//
// 描述: 对键的值进行递减操作
//
// 参数:
//
//	ctx (context.Context): 上下文
//	key (string): 键名
//	delta (int64): 递减量
//
// 返回:
//
//	int64: 递减后的值
//	error: 错误信息
func (s *RedisStorage) Decrement(ctx context.Context, key string, delta int64) (int64, error) {
	return s.Increment(ctx, key, -delta)
}

// Lock 分布式锁
//
// 描述: 获取分布式锁
//
// 参数:
//
//	ctx (context.Context): 上下文
//	key (string): 锁键名
//	expiration (time.Duration): 锁过期时间
//
// 返回:
//
//	bool: 是否获取成功
//	error: 错误信息
func (s *RedisStorage) Lock(ctx context.Context, key string, expiration time.Duration) (bool, error) {
	if s.client == nil {
		return false, errors.New("🚫 Redis连接未建立")
	}

	if key == "" {
		return false, errors.New("🚫 锁键名不能为空")
	}

	lockKey := s.buildKey(fmt.Sprintf("lock:%s", key))
	result, err := s.client.SetNX(ctx, lockKey, "locked", expiration).Result()
	if err != nil {
		return false, errors.Wrap(err, "🚫 获取分布式锁失败")
	}

	return result, nil
}

// Unlock 释放分布式锁
//
// 描述: 释放分布式锁
//
// 参数:
//
//	ctx (context.Context): 上下文
//	key (string): 锁键名
//
// 返回:
//
//	error: 错误信息
func (s *RedisStorage) Unlock(ctx context.Context, key string) error {
	if s.client == nil {
		return errors.New("🚫 Redis连接未建立")
	}

	if key == "" {
		return errors.New("🚫 锁键名不能为空")
	}

	lockKey := s.buildKey(fmt.Sprintf("lock:%s", key))
	if err := s.client.Del(ctx, lockKey).Err(); err != nil {
		return errors.Wrap(err, "🚫 释放分布式锁失败")
	}

	return nil
}

// serializeValue 序列化值
//
// 描述: 将值序列化为字符串
//
// 参数:
//
//	value (interface{}): 要序列化的值
//
// 返回:
//
//	string: 序列化后的字符串
//	error: 错误信息
func (s *RedisStorage) serializeValue(value interface{}) (string, error) {
	switch v := value.(type) {
	case string:
		return v, nil
	case []byte:
		return string(v), nil
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64, bool:
		return convertor.ToString(v), nil
	default:
		data, err := json.Marshal(v)
		if err != nil {
			return "", errors.Wrap(err, "🚫 JSON序列化失败")
		}
		return string(data), nil
	}
}

// deserializeValue 反序列化值
//
// 描述: 将字符串反序列化为值
//
// 参数:
//
//	data (string): 序列化的字符串
//
// 返回:
//
//	interface{}: 反序列化后的值
//	error: 错误信息
func (s *RedisStorage) deserializeValue(data string) (interface{}, error) {
	// 尝试解析为JSON
	var jsonValue interface{}
	if err := json.Unmarshal([]byte(data), &jsonValue); err == nil {
		return jsonValue, nil
	}

	// 如果不是JSON，返回原始字符串
	return data, nil
}

// GetClient 获取Redis客户端
//
// 描述: 获取底层的Redis客户端
//
// 返回:
//
//	*redis.Client: Redis客户端
func (s *RedisStorage) GetClient() *redis.Client {
	return s.client
}

// GetConfig 获取配置
//
// 描述: 获取Redis配置信息
//
// 返回:
//
//	*RedisConfig: Redis配置
func (s *RedisStorage) GetConfig() *RedisConfig {
	return s.config
}
