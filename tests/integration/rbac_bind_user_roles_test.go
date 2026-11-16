package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"package-base/internal/container"
	"package-base/internal/core/config"
	"package-base/internal/core/log"
	"package-base/internal/core/utils"
	"package-base/internal/routes"
	"package-base/tests/testhelpers"

	"github.com/gofiber/fiber/v2"
)

// Function: Test_User_Bind_Roles_Success
//
// Description: 创建两个角色并绑定到用户（覆盖式），随后验证覆盖与显式解绑逻辑。
//
// Parameters:
//
//	t(*testing.T): 测试句柄
//
// Returns:
//
//	无
func Test_User_Bind_Roles_Success(t *testing.T) {
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

	// 使用随机 userID（无外键约束，绑定表使用自由字符串ID）
	uuid := utils.GenerateShortUUID()
	if len(uuid) > 22 {
		uuid = uuid[:22]
	}
	userID := fmt.Sprintf("bind_%s", uuid)

	app := fiber.New()
	r := routes.NewRoutes(app, cont)
	r.Setup()
	addr := "127.0.0.1:18091"
	go app.Listen(addr)
	defer app.Shutdown()

	// 等待服务就绪
	if ok, err := testhelpers.WaitForServer(addr, 3*time.Second, 100*time.Millisecond); !ok || err != nil {
		t.Fatalf("服务未就绪: %v", err)
	}

	// 创建两个角色
	mkRole := func(name string) uint {
		b, _ := json.Marshal(map[string]interface{}{"name": name})
		resp, err := http.Post("http://"+addr+"/api/rbac/roles", "application/json", bytes.NewReader(b))
		if err != nil {
			t.Fatalf("创建角色失败: %v", err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("创建角色期望200 实际 %d", resp.StatusCode)
		}
		var out struct {
			Data struct {
				ID uint `json:"id"`
			} `json:"data"`
		}
		_ = json.NewDecoder(resp.Body).Decode(&out)
		_ = resp.Body.Close()
		return out.Data.ID
	}
	rid1 := mkRole("bind_viewer")
	rid2 := mkRole("bind_manager")

	// 绑定角色（第一次）
	body1, _ := json.Marshal(map[string]interface{}{"role_ids": []uint{rid1}, "scope": "global"})
	req1, _ := http.NewRequest("PUT", fmt.Sprintf("http://%s/api/rbac/users/%s/roles", addr, userID), bytes.NewReader(body1))
	req1.Header.Set("Content-Type", "application/json")
	resp1, err := http.DefaultClient.Do(req1)
	if err != nil {
		t.Fatalf("绑定失败: %v", err)
	}
	if resp1.StatusCode != http.StatusOK {
		t.Fatalf("绑定期望200 实际 %d", resp1.StatusCode)
	}
	_ = resp1.Body.Close()

	// 校验数据库存在1条绑定
	db := cont.GetMySQLStorage().GetDB()
	var c1 int64
	if err := db.Table("user_role_bindings").Where("user_id=? AND scope=?", userID, "global").Count(&c1).Error; err != nil {
		t.Fatalf("查询绑定计数失败: %v", err)
	}
	if c1 != 1 {
		t.Fatalf("期望1条绑定 实际 %d", c1)
	}

	// 覆盖绑定（第二次）
	body2, _ := json.Marshal(map[string]interface{}{"role_ids": []uint{rid1, rid2}, "scope": "global"})
	req2, _ := http.NewRequest("PUT", fmt.Sprintf("http://%s/api/rbac/users/%s/roles", addr, userID), bytes.NewReader(body2))
	req2.Header.Set("Content-Type", "application/json")
	resp2, err := http.DefaultClient.Do(req2)
	if err != nil {
		t.Fatalf("绑定失败: %v", err)
	}
	if resp2.StatusCode != http.StatusOK {
		t.Fatalf("绑定期望200 实际 %d", resp2.StatusCode)
	}
	_ = resp2.Body.Close()

	// 校验数据库存在2条绑定
	var c2 int64
	if err := db.Table("user_role_bindings").Where("user_id=? AND scope=?", userID, "global").Count(&c2).Error; err != nil {
		t.Fatalf("查询绑定计数失败: %v", err)
	}
	if c2 != 2 {
		t.Fatalf("期望2条绑定 实际 %d", c2)
	}

	// 显式解绑（空数组 + explicit_empty=true）
	body3, _ := json.Marshal(map[string]interface{}{"role_ids": []uint{}, "scope": "global", "explicit_empty": true})
	req3, _ := http.NewRequest("PUT", fmt.Sprintf("http://%s/api/rbac/users/%s/roles", addr, userID), bytes.NewReader(body3))
	req3.Header.Set("Content-Type", "application/json")
	resp3, err := http.DefaultClient.Do(req3)
	if err != nil {
		t.Fatalf("解绑失败: %v", err)
	}
	if resp3.StatusCode != http.StatusOK {
		t.Fatalf("解绑期望200 实际 %d", resp3.StatusCode)
	}
	_ = resp3.Body.Close()

	var c3 int64
	if err := db.Table("user_role_bindings").Where("user_id=? AND scope=?", userID, "global").Count(&c3).Error; err != nil {
		t.Fatalf("查询解绑后计数失败: %v", err)
	}
	if c3 != 0 {
		t.Fatalf("期望0条绑定 实际 %d", c3)
	}
}
