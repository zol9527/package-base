package integration

import (
	"bytes"
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

// Function: Test_Role_Create_Success
//
// Description: 验证 POST /api/rbac/roles 能成功创建角色。
//
// Parameters:
//
//	t(*testing.T): 测试句柄
//
// Returns:
//
//	无
func Test_Role_Create_Success(t *testing.T) {
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
	addr := "127.0.0.1:18089"
	go app.Listen(addr)
	defer app.Shutdown()

	payload := map[string]interface{}{
		"name":        "Viewer",
		"description": "Read-only role",
	}
	b, _ := json.Marshal(payload)

	resp, err := http.Post("http://"+addr+"/api/rbac/roles", "application/json", bytes.NewReader(b))
	if err != nil {
		t.Fatalf("请求失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("期望 200 实际 %d", resp.StatusCode)
	}

	var out struct {
		Code int `json:"code"`
		Data struct {
			ID      uint   `json:"id"`
			Name    string `json:"name"`
			Enabled bool   `json:"enabled"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		t.Fatalf("解析响应失败: %v", err)
	}
	if out.Code != 0 || out.Data.Name != "Viewer" || !out.Data.Enabled {
		t.Fatalf("响应校验失败: %+v", out)
	}
}
