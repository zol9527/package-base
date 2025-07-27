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

// LogtoSessionStorage: 实现 Logto Storage 接口
// Description: Logto Storage 适配器，仅操作 session store
// Parameters:
//
//	store(*session.Store): session 存储实例
//	logger(interfaces.Logger): 日志实例
//
// Returns:
//
//	无
type LogtoSessionStorage struct {
	store  *session.Store
	logger interfaces.Logger
}

// GetItem: 实现 Logto Storage 接口
// Description: 从 session store 获取指定 key 的值
// Parameters:
//
//	key(string): 要获取的键
//
// Returns:
//
//	value(string): 获取到的值（如不存在则返回空字符串）
func (s *LogtoSessionStorage) GetItem(key string) string {
	// 从 store 从获取，并避免 nil 引用错误
	value := s.store.Get(key)
	if value == nil {
		return ""
	}

	// 类型断言
	if str, ok := value.(string); ok {
		return str
	}

	return ""
}

// SetItem: 实现 Logto Storage 接口
// Description: 设置 session store 指定 key 的值
// Parameters:
//
//	key(string): 要设置的键
//	value(string): 要设置的值
//
// Returns:
//
//	无
func (s *LogtoSessionStorage) SetItem(key, value string) {
	s.store.Set(key, value)
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

	// 使用最简单的 session 配置进行调试
	sessionConfig := session.Config{
		Lookup:     "cookie:logto-session", // 固定查找方式
		Secure:     false,                  // 开发环境不使用 HTTPS，设为 false
		Domain:     "",                     // 留空，让浏览器自动处理
		SameSite:   "lax",                  // 固定 SameSite 策略
		Expiration: 12 * time.Hour,         // 固定 session 过期时间
	}

	// 根据配置选择存储类型
	switch cfg.Session.Storage {
	case "redis":
		// TODO 完成 redis 的对接
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
		Scopes:    []string{"openid", "profile", "email"}, // 添加必要的作用域
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
		// 获取 session store
		store := m.sess.Get(c)

		// 创建 Logto Storage 实例
		storage := &LogtoSessionStorage{
			store:  store,
			logger: m.logger,
		}

		// 创建 LogtoClient - 确保每个请求都有独立但正确配置的实例
		logtoClient := client.NewLogtoClient(m.logtoConfig, storage)

		// 将实例存储到上下文中
		c.Locals("logto_client", logtoClient)
		c.Locals("session_store", store)
		c.Locals("logto_storage", storage)

		// 处理请求
		err := c.Next()

		// 保存 session, 避免调用 save 时 session 保存在之前的 ID 中
		if saveErr := store.Save(); saveErr != nil {
			m.logger.Error("Failed to save session", interfaces.ErrorField(saveErr))
		}

		return err
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
