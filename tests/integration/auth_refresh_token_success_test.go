package integration

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"package-base/internal/models"
	"package-base/tests/testhelpers"
)

// Test_AuthRefreshTokenSuccess 刷新 token 成功 (获取新一对 access/refresh)
func Test_AuthRefreshTokenSuccess(t *testing.T) {
	wd, _ := os.Getwd()
	root := filepath.Clean(filepath.Join(wd, "../../"))
	cfgPath := filepath.Join(root, "config.json")
	addr := "127.0.0.1:18085"

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

	// 登录获取初始 tokens
	access1, refresh1, err := testhelpers.LoginForTokens(testhelpers.GetJSON, addr, "tester", password, 3*time.Second)
	if err != nil {
		t.Fatalf("登录失败: %v", err)
	}
	if access1 == "" || refresh1 == "" {
		t.Fatalf("登录未返回完整 token")
	}

	time.Sleep(20 * time.Millisecond) // 确保新旧 access JTI 不同（避免同秒生成概率极低但留余量）

	// 刷新
	access2, refresh2, err := testhelpers.RefreshTokens(testhelpers.GetJSON, addr, refresh1, 3*time.Second)
	if err != nil {
		t.Fatalf("刷新失败: %v", err)
	}

	if refresh2 == refresh1 {
		t.Fatalf("预期 refresh_token 已轮换")
	}
	if access2 == access1 {
		t.Fatalf("预期新的 access_token 与旧值不同")
	}

	// 额外解析检查格式（快速 JSON 验证）
	// 这里只做最小验证：结构上非空；详细声明验证在其他测试中已覆盖
	if len(access2) < 10 || len(refresh2) < 10 {
		t.Fatalf("返回 token 长度异常")
	}
}
