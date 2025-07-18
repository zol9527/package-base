package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"package-base/internal/container"
	"package-base/internal/core/config"
	"package-base/internal/routes"

	"github.com/gofiber/fiber/v2"
)

// Application 应用程序结构
type Application struct {
	app       *fiber.App
	container *container.Container
	config    *config.Config
}

// NewApplication 创建应用程序实例
func NewApplication() (*Application, error) {
	// 1. 加载配置
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "./config.json" // 默认配置文件路径
	}

	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		// 如果配置文件不存在，使用默认配置
		if os.IsNotExist(err) {
			log.Printf("Config file not found, using default config")
			cfg = config.DefaultConfig()

			// 从环境变量加载配置
			loader := config.NewLoader()
			loader.SetConfig(cfg)
			if err := loader.LoadFromEnv(); err != nil {
				return nil, fmt.Errorf("failed to load config from env: %w", err)
			}

			// 验证配置
			if err := cfg.Validate(); err != nil {
				return nil, fmt.Errorf("config validation failed: %w", err)
			}
		} else {
			return nil, fmt.Errorf("failed to load config: %w", err)
		}
	}

	// 2. 创建Fiber应用
	app := fiber.New(fiber.Config{
		ServerHeader: fmt.Sprintf("%s v%s", cfg.App.Name, cfg.App.Version),
		AppName:      fmt.Sprintf("%s v%s", cfg.App.Name, cfg.App.Version),
		BodyLimit:    parseBodyLimit(cfg.Server.BodyLimit),
		ReadTimeout:  cfg.Server.GetReadTimeout(),
		WriteTimeout: cfg.Server.GetWriteTimeout(),
		IdleTimeout:  cfg.Server.GetIdleTimeout(),
		Prefork:      cfg.Server.Prefork,
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}
			return c.Status(code).JSON(fiber.Map{
				"error": err.Error(),
				"code":  code,
			})
		},
	})

	// 3. 创建依赖注入容器
	cont := container.NewContainer(cfg)

	return &Application{
		app:       app,
		container: cont,
		config:    cfg,
	}, nil
}

// Initialize 初始化应用程序
func (a *Application) Initialize() error {
	ctx := context.Background()

	// 初始化容器
	if err := a.container.Initialize(ctx); err != nil {
		return fmt.Errorf("failed to initialize container: %w", err)
	}

	// 设置路由
	routes := routes.NewRoutes(a.app, a.container)
	routes.Setup()

	return nil
}

// Start 启动应用程序
func (a *Application) Start() error {
	// 创建监听地址
	addr := fmt.Sprintf("%s:%d", a.config.Server.Host, a.config.Server.Port)

	// 在生产模式下禁用调试信息
	if a.config.IsProduction() {
		a.app.Server().Logger = nil
	}

	// 启动HTTP服务器
	go func() {
		if err := a.app.Listen(addr); err != nil {
			log.Printf("HTTP server startup error: %v", err)
		}
	}()

	log.Printf("🚀 %s started successfully", a.config.App.Name)
	log.Printf("🌍 Listening on http://%s", addr)
	log.Printf("📖 Environment: %s", a.config.App.Mode)

	return nil
}

// Stop 停止应用程序
func (a *Application) Stop() error {
	log.Println("🛑 Shutting down server...")

	// 停止HTTP服务器
	if err := a.app.Shutdown(); err != nil {
		log.Printf("Error shutting down HTTP server: %v", err)
	}

	// 关闭容器
	if err := a.container.Close(); err != nil {
		log.Printf("Error closing container: %v", err)
	}

	log.Println("✅ Server shutdown completed")
	return nil
}

// WaitForShutdown 等待关闭信号
func (a *Application) WaitForShutdown() {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	<-quit
	log.Println("📡 Shutdown signal received")

	// 优雅关闭，超时时间30秒
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	done := make(chan bool, 1)
	go func() {
		a.Stop()
		done <- true
	}()

	select {
	case <-done:
		log.Println("✅ Graceful shutdown completed")
	case <-ctx.Done():
		log.Println("⚠️ Shutdown timeout, forcing exit")
	}
}

// parseBodyLimit 解析body大小限制
func parseBodyLimit(limit string) int {
	switch limit {
	case "1MB":
		return 1 * 1024 * 1024
	case "2MB":
		return 2 * 1024 * 1024
	case "4MB":
		return 4 * 1024 * 1024
	case "8MB":
		return 8 * 1024 * 1024
	case "16MB":
		return 16 * 1024 * 1024
	default:
		return 4 * 1024 * 1024 // 默认4MB
	}
}

func main() {
	// 创建应用程序
	app, err := NewApplication()
	if err != nil {
		log.Fatalf("Failed to create application: %v", err)
	}

	// 初始化应用程序
	if err := app.Initialize(); err != nil {
		log.Fatalf("Failed to initialize application: %v", err)
	}

	// 启动应用程序
	if err := app.Start(); err != nil {
		log.Fatalf("Failed to start application: %v", err)
	}

	// 等待关闭信号
	app.WaitForShutdown()
}
