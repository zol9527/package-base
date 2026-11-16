package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
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

// Function: Test_Role_Update_Then_Delete
//
// Description: 创建角色 -> 更新 -> 删除 的集成验证。
//
// Parameters:
//
//	t(*testing.T): 测试句柄
//
// Returns:
//
//	无
func Test_Role_Update_Then_Delete(t *testing.T) {
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
	addr := "127.0.0.1:18090"
	go app.Listen(addr)
	defer app.Shutdown()

	// 1) 创建角色
	b1, _ := json.Marshal(map[string]interface{}{"name": "Manager", "description": "manage"})
	resp1, err := http.Post("http://"+addr+"/api/rbac/roles", "application/json", bytes.NewReader(b1))
	if err != nil {
		t.Fatalf("创建失败: %v", err)
	}
	if resp1.StatusCode != http.StatusOK {
		t.Fatalf("创建期望200 实际 %d", resp1.StatusCode)
	}
	var out1 struct {
		Data struct {
			ID uint `json:"id"`
		} `json:"data"`
	}
	_ = json.NewDecoder(resp1.Body).Decode(&out1)
	_ = resp1.Body.Close()

	// 2) 更新角色
	b2, _ := json.Marshal(map[string]interface{}{"name": "Manager2", "enabled": false})
	req2, _ := http.NewRequest("PATCH", fmt.Sprintf("http://%s/api/rbac/roles/%d", addr, out1.Data.ID), bytes.NewReader(b2))
	req2.Header.Set("Content-Type", "application/json")
	resp2, err := http.DefaultClient.Do(req2)
	if err != nil {
		t.Fatalf("更新失败: %v", err)
	}
	if resp2.StatusCode != http.StatusOK {
		t.Fatalf("更新期望200 实际 %d", resp2.StatusCode)
	}
	_ = resp2.Body.Close()

	// 3) 删除角色
	req3, _ := http.NewRequest("DELETE", fmt.Sprintf("http://%s/api/rbac/roles/%d", addr, out1.Data.ID), nil)
	resp3, err := http.DefaultClient.Do(req3)
	if err != nil {
		t.Fatalf("删除失败: %v", err)
	}
	if resp3.StatusCode != http.StatusOK {
		t.Fatalf("删除期望200 实际 %d", resp3.StatusCode)
	}
	_ = resp3.Body.Close()
}
