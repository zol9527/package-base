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
	cfg := r.container.GetConfig()

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

	// logto Session 中间键
	logtoSessionMiddleware, err := middleware.NewLogtoSessionMiddleware(r.container.GetConfig(), logger)
	if err != nil {
		logger.Fatal("Failed to create Logto session middleware")
	}
	r.app.Use(logtoSessionMiddleware.Handler())

	// CORS中间件
	r.app.Use(cors.New(cors.Config{
		AllowOrigins:     cfg.Cors.AllowOrigin,
		AllowMethods:     cfg.Cors.AllowMethods,
		AllowHeaders:     cfg.Cors.AllowHeaders,
		ExposeHeaders:    cfg.Cors.ExposeHeaders,
		AllowCredentials: cfg.Cors.AllowCredentials,
		MaxAge:           cfg.Cors.MaxAge,
	}))
}

// setupAPIRoutes 设置API路由
func (r *Routes) setupAPIRoutes() {
	// API v1 路由组
	v1 := r.app.Group("/api/v1")

	// Logto 身份认证
	logtoAuthHandler := r.container.GetLogtoAuthHandler()

	// 认证路由组
	auth := v1.Group("/auth")
	logtoAuthHandler.RegisterLogtoAuthRoutes(auth)
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
