package middleware

import (
	"package-base/internal/core/config"
	"package-base/internal/core/interfaces"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/session/v2"
	"github.com/logto-io/go/v2/client"
	"github.com/pkg/errors"
)

// --- Logto Session 存储适配器 ---

// LogtoSessionStorage: Logto Session 存储适配器
type LogtoSessionStorage struct {
	store *session.Store
}

// GetItem: 实现 Logto Storage 接口
func (s *LogtoSessionStorage) GetItem(key string) string {
	value := s.store.Get(key)
	if value == nil {
		return ""
	}

	if str, ok := value.(string); ok {
		return str
	}

	return ""
}

// SetItem: 实现 Logto Storage 接口
func (s *LogtoSessionStorage) SetItem(key, value string) {
	s.store.Set(key, value)
	s.store.Save()
}

// --- LogtoSessionMiddleware: Logto 会话中间件 ---

// LogtoSessionMiddleware: Logto 会话中间件
type LogtoSessionMiddleware struct {
	cfg         *config.Config
	logger      interfaces.Logger
	sess        *session.Session
	logtoConfig *client.LogtoConfig
}

// NewLogtoSessionMiddleware: 创建 Logto 会话中间件实例
func NewLogtoSessionMiddleware(cfg *config.Config, logger interfaces.Logger) (*LogtoSessionMiddleware, error) {
	// 验证 Logto 配置
	if !cfg.Logto.IsEnabled() {
		return nil, errors.New("logto 未启用")
	}

	if err := cfg.Logto.Validate(); err != nil {
		return nil, errors.Wrap(err, "logto 配置无效")
	}

	// 创建 session 配置，直接使用 hard code 值
	// todo 正式环境修改 secure, domain, samesite 等配置
	sessionConfig := session.Config{
		Lookup:     "cookie:logto-session", // 固定查找方式
		Secure:     false,                  // 固定 Cookie 安全标志
		Domain:     "localhost",            // 固定 Cookie 域
		SameSite:   "lax",                  // 固定 SameSite 策略
		Expiration: 12 * time.Hour,         // 固定 session 过期时间
	}

	// 根据配置选择存储类型
	switch cfg.Session.Storage {
	case "redis":
		// todo 完成 redis 的对接
		// 如果使用 Redis 存储，需要先配置 Redis 存储
		// sessionConfig.Storage = redis.New(redis.Config{...})
		logger.Named("session").Info("🔄 使用内存存储，Redis 存储配置待实现")
	default:
		// 默认使用内存存储
		logger.Named("session").Info("📝 使用内存存储会话数据")
	}

	// 创建 session
	sess := session.New(sessionConfig)

	// 创建 Logto 客户端配置
	logtoConfig := &client.LogtoConfig{
		Endpoint:  cfg.Logto.GetEndpoint(),
		AppId:     cfg.Logto.GetAppID(),
		AppSecret: cfg.Logto.GetSecret(),
	}

	return &LogtoSessionMiddleware{
		cfg:         cfg,
		logger:      logger,
		sess:        sess,
		logtoConfig: logtoConfig,
	}, nil
}

// Handler: 处理会话中间件
func (m *LogtoSessionMiddleware) Handler() fiber.Handler {
	return func(c *fiber.Ctx) error {
		log := m.logger.Named("middleware.logto_session")

		// 获取 session store
		store := m.sess.Get(c)

		// 创建 Logto Storage 实例
		storage := &LogtoSessionStorage{store: store}

		// 创建 Logto 客户端实例
		logtoClient := client.NewLogtoClient(m.logtoConfig, storage)
		logtoClient.FetchUserInfo()

		// 将实例存储到上下文中
		c.Locals("logto_client", logtoClient)
		c.Locals("session_store", store)
		c.Locals("logto_storage", storage)

		log.Debug("Logto会话中间件初始化完成")
		return c.Next()
	}
}

// GetLogtoClient: 从上下文中获取 Logto 客户端
//
// Description: 从 fiber 上下文中提取 Logto 客户端实例
//
// Parameters:
//
//	c(*fiber.Ctx): fiber 上下文
//
// Returns:
//
//	client(*client.LogtoClient): Logto 客户端实例
//	ok(bool): 获取是否成功
func GetLogtoClient(c *fiber.Ctx) (*client.LogtoClient, bool) {
	client, ok := c.Locals("logto_client").(*client.LogtoClient)
	return client, ok
}

// GetSessionStore: 从上下文中获取 session store
//
// Description: 从 fiber 上下文中提取 session store 实例
//
// Parameters:
//
//	c(*fiber.Ctx): fiber 上下文
//
// Returns:
//
//	store(*session.Store): session store 实例
//	ok(bool): 获取是否成功
func GetSessionStore(c *fiber.Ctx) (*session.Store, bool) {
	store, ok := c.Locals("session_store").(*session.Store)
	return store, ok
}
