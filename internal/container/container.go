package container

import (
	"context"
	"time"

	"package-base/internal/core/config"
	"package-base/internal/core/interfaces"
	"package-base/internal/core/storages"
	"package-base/internal/core/utils"
	rbacModel "package-base/internal/models/rbac"
	"package-base/internal/repositories"
	repoInterfaces "package-base/internal/repositories/interfaces"
	rbacRepos "package-base/internal/repositories/rbac"
	"package-base/internal/services"
	serviceInterfaces "package-base/internal/services/interfaces"
	rbacService "package-base/internal/services/rbac"

	"github.com/pkg/errors"
)

// Container 依赖注入容器
type Container struct {
	config *config.Config

	// 基础设施
	logger       interfaces.Logger
	mysqlStorage interfaces.RelationalStorage
	redisStorage interfaces.KeyValueStorage

	// 对象存储抽象
	objectStorage interfaces.ObjectStorage

	// 核心组件
	jwtManager *utils.JWTManager
	validator  *utils.Validator

	// 存储层
	userRepo           repoInterfaces.UserRepository
	sessionRepo        repoInterfaces.SessionRepository
	tokenBlacklistRepo repoInterfaces.TokenBlacklistRepository
	loginAttemptRepo   repoInterfaces.LoginAttemptRepository
	// RBAC 存储
	roleRepo    repoInterfaces.RoleRepository
	permRepo    repoInterfaces.PermissionRepository
	bindingRepo repoInterfaces.BindingRepository
	auditRepo   repoInterfaces.AuditRepository

	// 服务层
	authService      serviceInterfaces.AuthService
	securityService  serviceInterfaces.SecurityService
	roleService      serviceInterfaces.RoleService
	bindingService   serviceInterfaces.BindingService
	permQueryService serviceInterfaces.PermissionsQueryService
	// RBAC 服务
	authorizationService serviceInterfaces.AuthorizationService

	// 处理器/中间件层：外部身份服务依赖已移除，按需再扩展
	authMiddleware interface{} // 占位，后续可引入新的认证中间件
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

	// 处理器/中间件层当前暂无特定初始化逻辑

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

	// 初始化对象存储 (T017 分支化)
	switch c.config.ObjectStorage {
	case "", "local":
		localObj, err := storages.NewLocalObjectStorage("data/objects")
		if err != nil {
			return errors.Wrap(err, "failed to create local object storage")
		}
		c.objectStorage = localObj
	default:
		// 未来扩展：如 ali_oss, s3 等；当前返回占位错误以明确未实现
		c.logger.Warn("object storage driver not implemented, fallback to local", interfaces.String("driver", c.config.ObjectStorage))
		localObj, err := storages.NewLocalObjectStorage("data/objects")
		if err != nil {
			return errors.Wrap(err, "failed to create fallback local object storage")
		}
		c.objectStorage = localObj
	}

	return nil
}

// initCoreComponents 初始化核心组件
func (c *Container) initCoreComponents() error {
	// 初始化JWT管理器
	accessDur := c.config.JWT.GetAccessTokenExpire()
	if c.config.TokenTTL > 0 {
		accessDur = time.Duration(c.config.TokenTTL) * time.Second
	}
	jwtManager := utils.NewJWTManager(
		c.config.JWT.Secret,
		accessDur,
		c.config.JWT.GetRefreshTokenExpire(),
	)
	c.jwtManager = jwtManager

	// 初始化验证器
	c.validator = utils.NewValidator()

	return nil
}

// initRepositories 初始化存储层
func (c *Container) initRepositories() error {
	// 初始化用户仓库
	c.userRepo = repositories.NewUserRepository(c.mysqlStorage, c.logger)
	// 初始化会话仓库
	c.sessionRepo = repositories.NewSessionRepository(c.mysqlStorage, c.logger)

	// 初始化Token黑名单仓库
	c.tokenBlacklistRepo = repositories.NewTokenBlacklistRepository(c.redisStorage, c.logger)

	// 初始化登录尝试仓库
	c.loginAttemptRepo = repositories.NewLoginAttemptRepository(c.mysqlStorage, c.logger)

	// 初始化 RBAC 仓储（骨架）
	c.roleRepo = rbacRepos.NewRoleRepository(c.mysqlStorage, c.logger)
	c.permRepo = rbacRepos.NewPermissionRepository(c.mysqlStorage, c.logger)
	c.bindingRepo = rbacRepos.NewBindingRepository(c.mysqlStorage, c.logger)
	c.auditRepo = rbacRepos.NewAuditRepository(c.mysqlStorage, c.logger)

	// 迁移 RBAC 相关表（临时：AutoMigrate）
	_ = c.mysqlStorage.Migrate(context.Background(),
		&rbacModel.Role{},
		&rbacModel.Permission{},
		&rbacModel.RolePermission{},
		&rbacModel.UserRoleBinding{},
		&rbacModel.AccessDecision{},
	)

	return nil
}

// initServices 初始化服务层
func (c *Container) initServices() error {
	// 初始化认证服务 (复用已创建的 jwtManager 以支持动态调整时长)
	c.authService = services.NewAuthServiceWithManager(
		c.userRepo,
		c.sessionRepo,
		c.tokenBlacklistRepo,
		c.jwtManager,
		c.logger,
	)

	// 初始化安全服务
	c.securityService = services.NewSecurityService(
		c.logger,
		c.loginAttemptRepo,
		c.redisStorage,
		nil, // 使用默认选项
	)

	// 初始化 RBAC 授权服务（骨架）
	c.authorizationService = rbacService.NewServiceWithRepos(
		c.logger,
		c.roleRepo,
		c.permRepo,
		c.bindingRepo,
		c.auditRepo,
		c.userRepo,
	)

	// 初始化角色管理服务
	c.roleService = rbacService.NewRoleService(c.logger, c.roleRepo)

	// 初始化绑定服务
	c.bindingService = rbacService.NewBindingService(c.logger, c.bindingRepo)

	// 初始化权限查询服务
	c.permQueryService = rbacService.NewPermissionsQueryService(c.logger, c.bindingRepo, c.permRepo)

	return nil
}

// initHandlers 初始化处理器层
// initHandlers / initMiddlewares 当前无特定初始化逻辑

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

// GetObjectStorage 获取对象存储
func (c *Container) GetObjectStorage() interfaces.ObjectStorage {
	return c.objectStorage
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

// GetAuthService 获取认证服务
func (c *Container) GetAuthService() serviceInterfaces.AuthService {
	return c.authService
}

// GetBindingService 获取绑定服务
func (c *Container) GetBindingService() serviceInterfaces.BindingService {
	return c.bindingService
}

// GetRoleService 获取角色服务
func (c *Container) GetRoleService() serviceInterfaces.RoleService {
	return c.roleService
}

// GetAuthorizationService 获取 RBAC 授权服务
func (c *Container) GetAuthorizationService() serviceInterfaces.AuthorizationService {
	return c.authorizationService
}

// GetPermissionsQueryService 获取权限查询服务
func (c *Container) GetPermissionsQueryService() serviceInterfaces.PermissionsQueryService {
	return c.permQueryService
}

// GetPermissionRepository 获取权限仓储
func (c *Container) GetPermissionRepository() repoInterfaces.PermissionRepository {
	return c.permRepo
}

// GetAuditRepository 获取审计仓储
func (c *Container) GetAuditRepository() repoInterfaces.AuditRepository {
	return c.auditRepo
}

// 处理器/中间件 Getter：当前无特定暴露方法

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
