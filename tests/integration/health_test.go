package integration

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"package-base/internal/container"
	"package-base/internal/core/config"
	"package-base/internal/core/interfaces"
	"package-base/internal/core/log"
	"package-base/internal/routes"

	"github.com/gofiber/fiber/v2"
)

// Test_HealthEndpoint 健康检查端点应返回 200 且 status=healthy (US1 T021)
func Test_HealthEndpoint(t *testing.T) {
	// 计算项目根目录: tests/integration -> ../../
	wd, _ := os.Getwd()
	root := filepath.Clean(filepath.Join(wd, "../../"))
	configPath := filepath.Join(root, "config.json")
	if os.Getenv("CONFIG_PATH") == "" {
		_ = os.Setenv("CONFIG_PATH", configPath)
	} else {
		configPath = os.Getenv("CONFIG_PATH")
	}
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		t.Fatalf("加载配置失败: %v", err)
	}

	// 构建 logger
	loggerConfig := &log.LoggerConfig{
		Level:            cfg.Logger.Level,
		Format:           cfg.Logger.Format,
		EnableFile:       false, // 测试中禁用文件写入
		EnableConsole:    false,
		ColorOutput:      false,
		EnableCaller:     false,
		EnableStacktrace: false,
	}
	logger, err := log.NewLogger(loggerConfig)
	if err != nil {
		t.Fatalf("创建 logger 失败: %v", err)
	}

	cont := container.NewContainer(cfg, logger)
	if err := cont.Initialize(t.Context()); err != nil { // 使用测试上下文
		t.Fatalf("容器初始化失败: %v", err)
	}
	defer cont.Close()

	fiberApp := fiber.New(fiber.Config{AppName: fmt.Sprintf("%s v%s", cfg.App.Name, cfg.App.Version)})
	r := routes.NewRoutes(fiberApp, cont)
	r.Setup()

	// 启动 Fiber (随机端口: 0) 不能直接，因为我们要固定 URL，干脆使用配置端口
	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	go func() {
		_ = fiberApp.Listen(addr)
	}()
	defer fiberApp.Shutdown()

	// 轮询 /health 最长 3 秒
	url := fmt.Sprintf("http://%s/health", addr)
	var resp *http.Response
	var lastErr error
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		resp, lastErr = http.Get(url) // #nosec G107
		if lastErr == nil {
			break
		}
		time.Sleep(50 * time.Millisecond)
	}
	if lastErr != nil {
		t.Fatalf("请求健康检查失败: %v", lastErr)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("期望状态码 200 实际 %d", resp.StatusCode)
	}

	var body struct {
		Status string `json:"status"`
		App    string `json:"app"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("解析响应失败: %v", err)
	}
	if body.Status != "healthy" {
		t.Fatalf("期望 healthy 实际 %s", body.Status)
	}
	if body.App == "" {
		t.Fatalf("期望 app 名称非空")
	}

	// 日志接口编译期保证
	_ = interfaces.String("test", "value")
}
