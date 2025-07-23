package container

import (
	"context"

	"package-base/internal/core/config"
	"package-base/internal/core/interfaces"
	"package-base/internal/core/storages"
	"package-base/internal/core/utils"
	"package-base/internal/handlers"
	"package-base/internal/middleware"
	"package-base/internal/repositories"
	repoInterfaces "package-base/internal/repositories/interfaces"
	"package-base/internal/services"
	serviceInterfaces "package-base/internal/services/interfaces"

	"github.com/pkg/errors"
)

// Container 依赖注入容器
type Container struct {
	config *config.Config

	// 基础设施
	logger       interfaces.Logger
	mysqlStorage interfaces.RelationalStorage
	redisStorage interfaces.KeyValueStorage

	// 核心组件
	jwtManager *utils.JWTManager
	validator  *utils.Validator

	// 存储层
	sessionRepo        repoInterfaces.SessionRepository
	tokenBlacklistRepo repoInterfaces.TokenBlacklistRepository
	loginAttemptRepo   repoInterfaces.LoginAttemptRepository

	// 服务层
	securityService serviceInterfaces.SecurityService

	// 处理器层
	logtoAuthHandler *handlers.LogtoAuthHandler

	// 中间件
	authMiddleware *middleware.AuthMiddleware
}

// NewContainer 创建依赖注入容器
func NewContainer(cfg *config.Config, logger interfaces.Logger) *Container {
	return &Container{
		config: cfg,
		logger: logger,
	}
}

// Initialize 初始化容器
func (c *Container) Initialize(ctx context.Context) error {

	// 2. 初始化存储
	if err := c.initStorages(ctx); err != nil {
		return errors.Wrap(err, "failed to initialize storages")
	}

	// 3. 初始化核心组件
	if err := c.initCoreComponents(); err != nil {
		return errors.Wrap(err, "failed to initialize core components")
	}

	// 4. 初始化存储层
	if err := c.initRepositories(); err != nil {
		return errors.Wrap(err, "failed to initialize repositories")
	}

	// 5. 初始化服务层
	if err := c.initServices(); err != nil {
		return errors.Wrap(err, "failed to initialize services")
	}

	// 6. 初始化处理器层
	if err := c.initHandlers(); err != nil {
		return errors.Wrap(err, "failed to initialize handlers")
	}

	// 7. 初始化中间件
	if err := c.initMiddlewares(); err != nil {
		return errors.Wrap(err, "failed to initialize middlewares")
	}

	c.logger.Info("container initialized successfully")
	return nil
}

// initStorages 初始化存储
func (c *Container) initStorages(ctx context.Context) error {
	// 初始化MySQL存储
	mysqlStorage, err := storages.NewMySQLStorage(c.config.Database.MySQL, c.logger)
	if err != nil {
		return errors.Wrap(err, "failed to create mysql storage")
	}

	if err := mysqlStorage.Connect(ctx); err != nil {
		return errors.Wrap(err, "failed to connect to mysql")
	}

	c.mysqlStorage = mysqlStorage

	// 初始化Redis存储
	redisStorage, err := storages.NewRedisStorage(c.config.Database.Redis)
	if err != nil {
		return errors.Wrap(err, "failed to create redis storage")
	}

	if err := redisStorage.Connect(ctx); err != nil {
		return errors.Wrap(err, "failed to connect to redis")
	}

	c.redisStorage = redisStorage

	return nil
}

// initCoreComponents 初始化核心组件
func (c *Container) initCoreComponents() error {
	// 初始化JWT管理器
	jwtManager := utils.NewJWTManager(
		c.config.JWT.Secret,
		c.config.JWT.GetAccessTokenExpire(),
		c.config.JWT.GetRefreshTokenExpire(),
	)
	c.jwtManager = jwtManager

	// 初始化验证器
	c.validator = utils.NewValidator()

	return nil
}

// initRepositories 初始化存储层
func (c *Container) initRepositories() error {
	// 初始化会话仓库
	c.sessionRepo = repositories.NewSessionRepository(c.mysqlStorage, c.logger)

	// 初始化Token黑名单仓库
	c.tokenBlacklistRepo = repositories.NewTokenBlacklistRepository(c.redisStorage, c.logger)

	// 初始化登录尝试仓库
	c.loginAttemptRepo = repositories.NewLoginAttemptRepository(c.mysqlStorage, c.logger)

	return nil
}

// initServices 初始化服务层
func (c *Container) initServices() error {

	// 初始化安全服务
	c.securityService = services.NewSecurityService(
		c.logger,
		c.loginAttemptRepo,
		c.redisStorage,
		nil, // 使用默认选项
	)

	return nil
}

// initHandlers 初始化处理器层
func (c *Container) initHandlers() error {

	return nil
}

// initMiddlewares 初始化中间件
func (c *Container) initMiddlewares() error {
	// 注意：中间件将在路由配置时创建，这里暂不实现
	return nil
}

// Close 关闭容器
func (c *Container) Close() error {
	var errs []error

	// 关闭存储连接
	if c.mysqlStorage != nil {
		if err := c.mysqlStorage.Close(); err != nil {
			errs = append(errs, errors.Wrap(err, "failed to close mysql storage"))
		}
	}

	if c.redisStorage != nil {
		if err := c.redisStorage.Close(); err != nil {
			errs = append(errs, errors.Wrap(err, "failed to close redis storage"))
		}
	}

	// 同步日志
	if c.logger != nil {
		if err := c.logger.Sync(); err != nil {
			errs = append(errs, errors.Wrap(err, "failed to sync logger"))
		}
	}

	if len(errs) > 0 {
		return errors.Errorf("errors occurred while closing container: %v", errs)
	}

	return nil
}

// GetConfig 获取配置
func (c *Container) GetConfig() *config.Config {
	return c.config
}

// GetLogger 获取日志实例
func (c *Container) GetLogger() interfaces.Logger {
	return c.logger
}

// GetMySQLStorage 获取MySQL存储
func (c *Container) GetMySQLStorage() interfaces.RelationalStorage {
	return c.mysqlStorage
}

// GetRedisStorage 获取Redis存储
func (c *Container) GetRedisStorage() interfaces.KeyValueStorage {
	return c.redisStorage
}

// GetJWTManager 获取JWT管理器
func (c *Container) GetJWTManager() *utils.JWTManager {
	return c.jwtManager
}

// GetValidator 获取验证器
func (c *Container) GetValidator() *utils.Validator {
	return c.validator
}

// GetSessionRepository 获取会话仓库
func (c *Container) GetSessionRepository() repoInterfaces.SessionRepository {
	return c.sessionRepo
}

// GetTokenBlacklistRepository 获取Token黑名单仓库
func (c *Container) GetTokenBlacklistRepository() repoInterfaces.TokenBlacklistRepository {
	return c.tokenBlacklistRepo
}

// GetLoginAttemptRepository 获取登录尝试仓库
func (c *Container) GetLoginAttemptRepository() repoInterfaces.LoginAttemptRepository {
	return c.loginAttemptRepo
}

// GetSecurityService 获取安全服务
func (c *Container) GetSecurityService() serviceInterfaces.SecurityService {
	return c.securityService
}

// --- Handlers ---
func (c *Container) GetLogtoAuthHandler() *handlers.LogtoAuthHandler {
	return c.logtoAuthHandler
}

// GetAuthMiddleware 获取认证中间件
func (c *Container) GetAuthMiddleware() *middleware.AuthMiddleware {
	return c.authMiddleware
}

// HealthCheck 健康检查
func (c *Container) HealthCheck(ctx context.Context) error {
	// 检查MySQL连接
	if !c.mysqlStorage.IsHealthy(ctx) {
		return errors.New("mysql storage is unhealthy")
	}

	// 检查Redis连接
	if !c.redisStorage.IsHealthy(ctx) {
		return errors.New("redis storage is unhealthy")
	}

	return nil
}
