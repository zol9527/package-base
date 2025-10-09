package integration

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"encoding/json"
	"package-base/internal/models"
	"package-base/tests/testhelpers"
)

// Test_AuthRefreshTokenInvalid 无效 refresh_token 返回 401 (US2 T0xx)
func Test_AuthRefreshTokenInvalid(t *testing.T) {
	wd, _ := os.Getwd()
	root := filepath.Clean(filepath.Join(wd, "../../"))
	cfgPath := filepath.Join(root, "config.json")
	addr := "127.0.0.1:18086"

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

	body, _ := json.Marshal(map[string]string{"refresh_token": "invalid-token-xyz"})
	status, code, msg, _, err := testhelpers.GetJSON("POST", "http://"+addr+"/api/auth/refresh", body, 3*time.Second)
	if err == nil && status == 200 {
		t.Fatalf("无效 refresh 不应成功刷新")
	}
	if status != 401 {
		t.Fatalf("期望 401 实际 %d", status)
	}
	if code != 401 {
		t.Fatalf("响应 code 期望 401 实际 %d", code)
	}
	if msg == "" {
		t.Fatalf("期望错误消息不为空")
	}
}
