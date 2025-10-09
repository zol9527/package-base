package routes

import (
	"package-base/internal/container"
	"package-base/internal/core/log"
	"package-base/internal/handlers"
	"package-base/internal/middleware"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"
)

// Function: Routes
//
// Description: 路由配置结构，封装 Fiber 应用与依赖容器，提供中间件注册、API 分组以及健康检查端点初始化。
//
// Fields:
//
//	app(*fiber.App): Fiber 应用实例
//	container(*container.Container): 全局依赖注入容器，暴露服务、配置与存储访问
type Routes struct {
	app       *fiber.App
	container *container.Container
}

// Function: NewRoutes
//
// Description: 创建并返回 Routes 实例，不做副作用注册，需进一步调用 Setup 完成路由装配。
//
// Parameters:
//
//	app(*fiber.App): Fiber 应用实例
//	container(*container.Container): 依赖容器
//
// Returns:
//
//	(*Routes): Routes 结构指针
func NewRoutes(app *fiber.App, container *container.Container) *Routes {
	return &Routes{
		app:       app,
		container: container,
	}
}

// Function: Setup
//
// Description: 装配全部路由，包括全局中间件、API 分组与健康检查。调用顺序：先中间件 -> API -> 健康检查。
//
// Parameters:
//
//	(r *Routes): 接收者，路由配置实例
//
// Returns:
//
//	(void): 无返回值
func (r *Routes) Setup() {
	// 设置全局中间件
	r.setupMiddlewares()

	// 设置API路由
	r.setupAPIRoutes()

	// 设置健康检查路由
	r.setupHealthRoutes()
}

// Function: setupMiddlewares
//
// Description: 注册全局中间件链，包括 panic 恢复、request id、统一结构化日志、安全头、CORS。
//
// Parameters:
//
//	(r *Routes): 接收者
//
// Returns:
//
//	(void): 无返回值
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

	// 认证 session 组件已移除，后续需要可在此添加新的认证中间件

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

// Function: setupAPIRoutes
//
// Description: 注册业务 API 路由。当前仅包含 /api/auth 下的认证登录与刷新接口，可在此扩展版本化分组。
//
// Parameters:
//
//	(r *Routes): 接收者
//
// Returns:
//
//	(void): 无返回值
func (r *Routes) setupAPIRoutes() {
	api := r.app.Group("/api")
	authGroup := api.Group("/auth")

	// 构建 AuthHandler （需要容器中已装配 AuthService）
	logger := log.L().Named("handler.auth")
	authHandler := handlers.NewAuthHandler(r.container.GetAuthService(), logger)

	authGroup.Post("/login", authHandler.Login)
	authGroup.Post("/refresh", authHandler.RefreshToken)
}

// Function: setupHealthRoutes
//
// Description: 注册健康检查与 ping 端点，用于外部存活探测与简单连通性验证。
//
// Parameters:
//
//	(r *Routes): 接收者
//
// Returns:
//
//	(void): 无返回值
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
