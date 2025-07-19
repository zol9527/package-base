package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"package-base/internal/container"
	"package-base/internal/core/config"
	"package-base/internal/core/interfaces"
	"package-base/internal/core/log"
	"package-base/internal/dto"
	"package-base/internal/routes"

	"github.com/gofiber/fiber/v2"
)

// Application 应用程序结构
type Application struct {
	app       *fiber.App
	container *container.Container
	config    *config.Config
	logger    interfaces.Logger
}

// NewApplication 创建应用程序实例
func NewApplication() (*Application, error) {
	// 加载配置
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "./config.json" // 默认配置文件路径
	}

	fmt.Print("Loading configuration from: ", configPath, "\n")
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		// 如果配置文件不存在，使用默认配置
		if os.IsNotExist(err) {
			fmt.Printf("Config file not found")

			// 验证配置
			if err := cfg.Validate(); err != nil {
				return nil, fmt.Errorf("config validation failed: %w", err)
			}
		} else {
			return nil, fmt.Errorf("failed to load config: %w", err)
		}
	}

	// 设置 log config
	logConfig := &log.LoggerConfig{
		Level:            cfg.Logger.Level,
		Format:           cfg.Logger.Format,
		EnableFile:       cfg.Logger.EnableFile,
		FilePath:         cfg.Logger.FilePath,
		MaxSize:          cfg.Logger.MaxSize,
		MaxBackups:       cfg.Logger.MaxBackups,
		MaxAge:           cfg.Logger.MaxAge,
		Compress:         cfg.Logger.Compress,
		LocalTime:        cfg.Logger.LocalTime,
		EnableConsole:    cfg.Logger.EnableConsole,
		ColorOutput:      cfg.Logger.ColorOutput,
		EnableCaller:     cfg.Logger.EnableCaller,
		EnableStacktrace: cfg.Logger.EnableStacktrace,
	}
	// 创建基础日志 logger
	defaultLogger, err := log.NewLogger(logConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create logger: %w", err)
	}

	// 2. 创建依赖注入容器
	cont := container.NewContainer(cfg, defaultLogger)

	// 3. 创建Fiber应用（顺序调整，保证 ErrorHandler 能访问 cont）
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
			// 使用日志记录
			log := cont.GetLogger()
			log.Error("Error occurred", interfaces.ErrorField(err))
			// 返回 internal error 而不是直接返回错误信息
			return c.Status(code).JSON(dto.InternalServerError("Internal Server Error"))
		},
	})

	return &Application{
		app:       app,
		container: cont,
		config:    cfg,
		logger:    defaultLogger,
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
			logger := a.logger.Named("application.start")
			logger.Error("🔥 HTTP server startup error", interfaces.ErrorField(err))
		}
	}()

	startLogger := a.logger.Named("application.start")
	startLogger.Info("🚀 启动成功", interfaces.String("name", a.config.App.Name))
	startLogger.Info("🌍 监听地址", interfaces.String("address", fmt.Sprintf("http://%s", addr)))
	startLogger.Info("📖 运行环境", interfaces.String("mode", a.config.App.Mode))

	return nil
}

// Stop 停止应用程序
func (a *Application) Stop() error {
	stopLogger := a.logger.Named("application.stop")
	stopLogger.Info("🛑 正在关闭服务器...")

	// 停止HTTP服务器
	if err := a.app.Shutdown(); err != nil {
		stopLogger.Error("关闭HTTP服务器出错", interfaces.ErrorField(err))
	}

	// 关闭容器
	if err := a.container.Close(); err != nil {
		stopLogger.Error("关闭容器出错", interfaces.ErrorField(err))
	}

	stopLogger.Info("✅ 服务器关闭完成")
	return nil
}

// WaitForShutdown 等待关闭信号
func (a *Application) WaitForShutdown() {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	<-quit
	shutdownLogger := a.logger.Named("application.shutdown")
	shutdownLogger.Info("📡 接收到关闭信号")

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
		shutdownLogger.Info("✅ 优雅关闭完成")
	case <-ctx.Done():
		shutdownLogger.Warn("⚠️ 关闭超时，强制退出")
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
		fmt.Printf("❌ Failed to create application: %v\n", err)
		os.Exit(1)
	}

	// 初始化应用程序
	if err := app.Initialize(); err != nil {
		mainLogger := app.logger.Named("main")
		mainLogger.Fatal("❌ 初始化应用程序失败", interfaces.ErrorField(err))
	}

	// 启动应用程序
	if err := app.Start(); err != nil {
		mainLogger := app.logger.Named("main")
		mainLogger.Fatal("❌ 启动应用程序失败", interfaces.ErrorField(err))
	}

	// 等待关闭信号
	app.WaitForShutdown()
}
