package integration

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"fmt"
	"package-base/internal/models"
	"package-base/tests/testhelpers"
)

// Test_AuthRefreshToken_OldRefreshReuseInvalid 刷新成功后旧 refresh_token 再次使用应 401 (token 轮换策略)
func Test_AuthRefreshToken_OldRefreshReuseInvalid(t *testing.T) {
	wd, _ := os.Getwd()
	root := filepath.Clean(filepath.Join(wd, "../../"))
	cfgPath := filepath.Join(root, "config.json")
	addr := "127.0.0.1:18087"

	_, cont, shutdown, err := testhelpers.StartTestApp(cfgPath, addr, true, 1500*time.Millisecond)
	if err != nil {
		t.Fatalf("启动测试应用失败: %v", err)
	}
	defer shutdown()

	if err := cont.GetMySQLStorage().Migrate(t.Context(), &models.User{}); err != nil {
		t.Fatalf("用户模型迁移失败: %v", err)
	}
	if err := testhelpers.EnsureSessionTable(cont); err != nil {
		t.Fatalf("创建 user_sessions 表失败: %v", err)
	}

	password := "Passw0rd!"
	if err := testhelpers.SeedTestUser(t, cont, password); err != nil {
		t.Fatalf("插入测试用户失败: %v", err)
	}

	// 登录获取旧 refresh
	_, oldRefresh, err := testhelpers.LoginForTokens(testhelpers.GetJSON, addr, "tester", password, 3*time.Second)
	if err != nil {
		t.Fatalf("登录失败: %v", err)
	}
	if oldRefresh == "" {
		t.Fatalf("未获得 refresh token")
	}

	time.Sleep(10 * time.Millisecond)

	// 首次刷新 -> 获得新 refresh
	_, newRefresh, err := testhelpers.RefreshTokens(testhelpers.GetJSON, addr, oldRefresh, 3*time.Second)
	if err != nil {
		t.Fatalf("第一次刷新失败: %v", err)
	}
	if newRefresh == oldRefresh {
		t.Fatalf("刷新后 refresh 未轮换")
	}

	// 使用旧 refresh 再试 -> 应 401
	body := fmt.Sprintf(`{"refresh_token":"%s"}`, oldRefresh)
	status, _, _, _, rErr := testhelpers.GetJSON("POST", "http://"+addr+"/api/auth/refresh", []byte(body), 3*time.Second)
	if rErr == nil && status == 200 {
		t.Fatalf("复用旧 refresh 不应成功")
	}
	if status != 401 {
		t.Fatalf("复用旧 refresh 期望401 实际 %d", status)
	}
}
