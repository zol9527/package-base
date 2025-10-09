package integration

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"package-base/internal/models"
	"package-base/tests/testhelpers"
)

// Test_AuthRefreshToken_Expired 人为缩短 refresh 有效期并等待过期，刷新应 401
func Test_AuthRefreshToken_Expired(t *testing.T) {
	wd, _ := os.Getwd()
	root := filepath.Clean(filepath.Join(wd, "../../"))
	cfgPath := filepath.Join(root, "config.json")
	addr := "127.0.0.1:18090"

	_, cont, shutdown, err := testhelpers.StartTestApp(cfgPath, addr, true, 1500*time.Millisecond)
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

	// 缩短 refresh TTL -> 1 秒
	cont.GetJWTManager().SetRefreshTokenDuration(1 * time.Second)

	// 登录获得 tokens
	access, refresh, err := testhelpers.LoginForTokens(testhelpers.GetJSON, addr, "tester", password, 3*time.Second)
	if err != nil {
		t.Fatalf("登录失败: %v", err)
	}
	if access == "" || refresh == "" {
		t.Fatalf("登录返回 token 为空")
	}

	// 等待过期
	time.Sleep(1100 * time.Millisecond)

	// 尝试刷新 -> 401
	_, _, rErr := testhelpers.RefreshTokens(testhelpers.GetJSON, addr, refresh, 3*time.Second)
	if rErr == nil {
		t.Fatalf("预期刷新失败 (refresh 已过期) 但却成功")
	}
}
