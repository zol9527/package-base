package integration

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"package-base/internal/container"
	"package-base/internal/core/config"
	"package-base/internal/core/log"
	"package-base/internal/routes"

	"github.com/gofiber/fiber/v2"
)

// Function: Test_Protected_Unauthorized
//
// Description: 未携带 Authorization 访问受保护接口，应返回 401。
//
// Parameters:
//
//	t(*testing.T): 测试句柄
//
// Returns:
//
//	无
func Test_Protected_Unauthorized(t *testing.T) {
	// 启动应用
	wd, _ := os.Getwd()
	root := filepath.Clean(filepath.Join(wd, "../../"))
	cfg, err := config.LoadConfig(filepath.Join(root, "config.json"))
	if err != nil {
		t.Fatalf("加载配置失败: %v", err)
	}
	logger, err := log.NewLogger(&log.LoggerConfig{Level: cfg.Logger.Level, Format: cfg.Logger.Format, EnableConsole: false, EnableFile: false})
	if err != nil {
		t.Fatalf("logger 创建失败: %v", err)
	}
	cont := container.NewContainer(cfg, logger)
	if err := cont.Initialize(t.Context()); err != nil {
		t.Fatalf("容器初始化失败: %v", err)
	}
	defer cont.Close()

	app := fiber.New()
	r := routes.NewRoutes(app, cont)
	r.Setup()
	addr := "127.0.0.1:18085"
	go app.Listen(addr)
	defer app.Shutdown()

	// 请求受保护端点（无 Authorization）
	resp, err := http.Get("http://" + addr + "/api/protected/ping")
	if err != nil {
		t.Fatalf("请求失败: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("期望 401 实际 %d", resp.StatusCode)
	}
	var parsed struct {
		Code    int    `json:"code"`
		Message string `json:"msg"`
	}
	_ = json.NewDecoder(resp.Body).Decode(&parsed)
	if parsed.Code != 401 {
		t.Fatalf("期望 code=401 实际 %d msg=%s", parsed.Code, parsed.Message)
	}
}
