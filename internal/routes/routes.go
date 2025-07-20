package routes

import (
	"package-base/internal/container"
	"package-base/internal/core/log"
	"package-base/internal/middleware"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"
)

// Routes 路由配置
type Routes struct {
	app       *fiber.App
	container *container.Container
}

// NewRoutes 创建路由配置
func NewRoutes(app *fiber.App, container *container.Container) *Routes {
	return &Routes{
		app:       app,
		container: container,
	}
}

// Setup 设置路由
func (r *Routes) Setup() {
	// 设置全局中间件
	r.setupMiddlewares()

	// 设置API路由
	r.setupAPIRoutes()

	// 设置健康检查路由
	r.setupHealthRoutes()
}

// setupMiddlewares 设置全局中间件
func (r *Routes) setupMiddlewares() {
	// 恢复中间件
	r.app.Use(recover.New())

	// 请求ID中间件
	r.app.Use(requestid.New())

	// 日志中间件
	logger := log.L() // 日志处理器已经在前面初始化，现在直接使用 L() 获取
	middlewareLogger := middleware.NewLogger(logger)
	r.app.Use(middlewareLogger.Handler())

	// 设置 security headers
	r.app.Use(middleware.SecurityHeaders())

	// CORS中间件
	r.app.Use(cors.New(cors.Config{
		AllowOrigins:     "*",
		AllowMethods:     "GET,POST,HEAD,PUT,DELETE,PATCH,OPTIONS",
		AllowHeaders:     "Origin,Content-Type,Accept,Authorization,X-Request-ID",
		AllowCredentials: true,
		ExposeHeaders:    "X-Request-ID",
	}))
}

// setupAPIRoutes 设置API路由
func (r *Routes) setupAPIRoutes() {
	// API v1 路由组
	v1 := r.app.Group("/api/v1")

	// 认证路由组
	auth := v1.Group("/auth")
	r.setupAuthRoutes(auth)

	// 用户路由组
	user := v1.Group("/user")
	r.setupUserRoutes(user)

	// 2FA路由组
	twofa := v1.Group("/2fa")
	r.setupTotpRoutes(twofa)
}

// setupAuthRoutes 设置认证路由
func (r *Routes) setupAuthRoutes(group fiber.Router) {
	authHandler := r.container.GetAuthHandler()

	// 公开路由（无需认证）
	group.Post("/register", authHandler.Register)
	group.Post("/login", authHandler.Login)
	group.Post("/refresh", authHandler.RefreshToken)

	// 需要认证的路由
	protected := group.Group("")
	authMiddleware := middleware.NewAuthMiddleware(r.container.GetConfig().JWT.Secret, r.container.GetLogger())
	protected.Use(authMiddleware.Authenticate())

	protected.Post("/logout", authHandler.Logout)
	protected.Post("/logout-all", authHandler.LogoutAll)
	protected.Put("/change-password", authHandler.ChangePassword)
	protected.Get("/profile", authHandler.GetProfile)
	protected.Get("/sessions", authHandler.GetSessions)
	protected.Get("/sessions/active", authHandler.GetActiveSessions)
	protected.Get("/sessions/recent", authHandler.GetRecentSessions)
	protected.Get("/sessions/stats", authHandler.GetSessionStats)
	protected.Delete("/sessions/:sessionId", authHandler.DeleteSession)
}

// setupUserRoutes 设置用户路由
func (r *Routes) setupUserRoutes(group fiber.Router) {
	userHandler := r.container.GetUserHandler()

	// 需要认证的路由
	protected := group.Group("")
	authMiddleware := middleware.NewAuthMiddleware(r.container.GetConfig().JWT.Secret, r.container.GetLogger())
	protected.Use(authMiddleware.Authenticate())

	protected.Get("/profile", userHandler.GetProfile)
	protected.Put("/profile", userHandler.UpdateProfile)
}

// setupHealthRoutes 设置健康检查路由
func (r *Routes) setupHealthRoutes() {
	r.app.Get("/health", func(c *fiber.Ctx) error {
		// 执行健康检查
		if err := r.container.HealthCheck(c.Context()); err != nil {
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
				"status": "unhealthy",
				"error":  err.Error(),
			})
		}

		return c.JSON(fiber.Map{
			"status":  "healthy",
			"app":     r.container.GetConfig().App.Name,
			"version": r.container.GetConfig().App.Version,
		})
	})

	// Ping 端点
	r.app.Get("/ping", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"message": "pong",
		})
	})
}

// setupTotpRoutes 设置2FA TOTP路由
func (r *Routes) setupTotpRoutes(group fiber.Router) {
	totpHandler := r.container.GetTotpHandler()

	// 需要认证的路由
	protected := group.Group("")
	authMiddleware := middleware.NewAuthMiddleware(r.container.GetConfig().JWT.Secret, r.container.GetLogger())
	protected.Use(authMiddleware.Authenticate())

	// 2FA管理接口
	protected.Post("/setup", totpHandler.GenerateSecret)                    // 生成2FA设置信息
	protected.Post("/enable", totpHandler.EnableTOTP)                       // 启用2FA
	protected.Post("/disable", totpHandler.DisableTOTP)                     // 禁用2FA
	protected.Post("/verify", totpHandler.VerifyTOTP)                       // 验证TOTP令牌
	protected.Post("/backup-verify", totpHandler.VerifyBackupCode)          // 验证备用码
	protected.Post("/regenerate-backup", totpHandler.RegenerateBackupCodes) // 重新生成备用码
	protected.Get("/status", totpHandler.GetTOTPStatus)                     // 获取2FA状态
	protected.Get("/backup-status", totpHandler.GetBackupCodeStatus)        // 获取备用码状态
}
