package integration

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"package-base/internal/models"
	"package-base/tests/testhelpers"
)

// Test_AuthRefreshToken_Tampered 篡改 refresh token 应 401
func Test_AuthRefreshToken_Tampered(t *testing.T) {
	wd, _ := os.Getwd()
	root := filepath.Clean(filepath.Join(wd, "../../"))
	cfgPath := filepath.Join(root, "config.json")
	addr := "127.0.0.1:18088"

	app, cont, shutdown, err := testhelpers.StartTestApp(cfgPath, addr, true, 1500*time.Millisecond)
	if err != nil {
		t.Fatalf("启动测试应用失败: %v", err)
	}
	defer shutdown()

	// 迁移用户表 & 会话表
	if err := cont.GetMySQLStorage().Migrate(t.Context(), &models.User{}); err != nil {
		t.Fatalf("用户模型迁移失败: %v", err)
	}
	if err := testhelpers.EnsureSessionTable(cont); err != nil {
		t.Fatalf("创建会话表失败: %v", err)
	}
	password := "Passw0rd!"
	if err := testhelpers.SeedTestUser(t, cont, password); err != nil {
		t.Fatalf("插入测试用户失败: %v", err)
	}

	// 登录获得合法 refresh (使用 helper)
	_, refresh, err := testhelpers.LoginForTokens(testhelpers.GetJSON, addr, "tester", password, 3*time.Second)
	if err != nil {
		t.Fatalf("登录失败: %v", err)
	}
	if refresh == "" {
		t.Fatalf("未返回 refresh token")
	}

	// 篡改 refresh (尾部加字符)
	tampered := refresh + "x"
	refreshBody, _ := json.Marshal(map[string]string{"refresh_token": tampered})
	rStatus, _, _, _, rErr := testhelpers.GetJSON("POST", "http://"+addr+"/api/auth/refresh", refreshBody, 3*time.Second)
	if rErr == nil && rStatus == 200 {
		t.Fatalf("篡改 token 不应刷新成功")
	}
	if rStatus != 401 {
		t.Fatalf("篡改 refresh 期望 401 实际 %d", rStatus)
	}
	_ = app // 引用避免未使用；app/shutdown 已 defer
}
