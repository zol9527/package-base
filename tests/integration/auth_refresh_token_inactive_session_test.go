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

// Test_AuthRefreshToken_InactiveSession 将会话标记为非激活后刷新应失败 (401)
func Test_AuthRefreshToken_InactiveSession(t *testing.T) {
	wd, _ := os.Getwd()
	root := filepath.Clean(filepath.Join(wd, "../../"))
	cfgPath := filepath.Join(root, "config.json")
	addr := "127.0.0.1:18091"

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

	// 登录获取 refresh token
	loginBody, _ := json.Marshal(map[string]string{"username": "tester", "password": password})
	status, code, _, data, err := testhelpers.GetJSON("POST", "http://"+addr+"/api/auth/login", loginBody, 3*time.Second)
	if err != nil {
		t.Fatalf("登录请求失败: %v", err)
	}
	if status != 200 || code != 0 {
		t.Fatalf("登录期望 200/0 实际 %d/%d", status, code)
	}
	var loginData struct {
		RefreshToken string `json:"refresh_token"`
		AccessToken  string `json:"access_token"`
	}
	if err := json.Unmarshal(data, &loginData); err != nil {
		t.Fatalf("解析登录data失败: %v", err)
	}
	if loginData.RefreshToken == "" {
		t.Fatalf("未返回 refresh token")
	}

	// 直接将对应 session 置为非激活
	// 这里通过仓库按 refresh token 查询再更新 (或直接执行 SQL)。为了避免引入新依赖，使用底层 GORM。
	gdb := cont.GetMySQLStorage().GetDB()
	res := gdb.Exec("UPDATE user_sessions SET is_active = ? WHERE refresh_token = ?", false, loginData.RefreshToken)
	if res.Error != nil {
		t.Fatalf("更新会话状态失败: %v", res.Error)
	}
	if res.RowsAffected == 0 {
		t.Fatalf("未找到需要更新的会话记录")
	}

	// 尝试刷新 -> 应 401
	refreshBody, _ := json.Marshal(map[string]string{"refresh_token": loginData.RefreshToken})
	rStatus, _, _, _, rErr := testhelpers.GetJSON("POST", "http://"+addr+"/api/auth/refresh", refreshBody, 3*time.Second)
	if rErr == nil && rStatus == 200 {
		t.Fatalf("非激活会话刷新不应成功")
	}
	if rStatus != 401 {
		t.Fatalf("非激活会话刷新期望 401 实际 %d", rStatus)
	}
}
