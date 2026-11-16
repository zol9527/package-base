package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"package-base/internal/core/utils"
	"package-base/internal/models"
	"package-base/tests/testhelpers"

	"golang.org/x/crypto/bcrypt"
)

// Function: Test_Audit_Query_Basic
//
// Description: 先登录获取 token，再访问受保护端点触发 RBAC 审计（403），随后查询审计接口应返回至少一条记录。
//
// Parameters:
//
//	t(*testing.T): 测试句柄
//
// Returns:
//
//	无
func Test_Audit_Query_Basic(t *testing.T) {
	// 启动测试应用
	wd, _ := os.Getwd()
	root := filepath.Clean(filepath.Join(wd, "../../"))
	cfgPath := filepath.Join(root, "config.json")
	addr := "127.0.0.1:18093"
	app, cont, shutdown, err := testhelpers.StartTestApp(cfgPath, addr, true, 3*time.Second)
	if err != nil {
		t.Fatalf("启动应用失败: %v", err)
	}
	defer shutdown()
	_ = app // 保留引用以防编译器优化

	// 准备用户/会话表，并注入专用种子用户
	if err := cont.GetMySQLStorage().Migrate(t.Context(), &models.User{}); err != nil {
		t.Fatalf("用户模型迁移失败: %v", err)
	}
	if err := testhelpers.EnsureSessionTable(cont); err != nil {
		t.Fatalf("创建会话表失败: %v", err)
	}
	username := "rbac_audit_user1"
	password := "Passw0rd!"
	db := cont.GetMySQLStorage().GetDB()
	var count int64
	if err := db.Table("users").Where("username=?", username).Count(&count).Error; err != nil {
		t.Fatalf("检查用户存在失败: %v", err)
	}
	hash, _ := bcrypt.GenerateFromPassword([]byte(password), 10)
	if count == 0 {
		uuid := utils.GenerateShortUUID()
		if len(uuid) > 22 {
			uuid = uuid[:22]
		}
		insertSQL := "INSERT INTO users (uuid, username, email, password_hash, status, roles, created_at, updated_at) VALUES (?,?,?,?,?,?,NOW(),NOW())"
		if err := db.Exec(insertSQL, uuid, username, username+"@example.com", string(hash), 1, 0).Error; err != nil {
			t.Fatalf("插入种子用户失败: %v", err)
		}
	} else {
		// 确保账号可用且密码匹配
		if err := db.Exec("UPDATE users SET password_hash=?, status=1 WHERE username=?", string(hash), username).Error; err != nil {
			t.Fatalf("更新用户状态/密码失败: %v", err)
		}
	}

	// 登录拿到 access token
	body, _ := json.Marshal(map[string]string{"username": username, "password": password})
	resp, err := http.Post("http://"+addr+"/api/auth/login", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("登录失败: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("登录期望200 实际 %d", resp.StatusCode)
	}
	var login struct {
		Code int `json:"code"`
		Data struct {
			AccessToken string `json:"access_token"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&login); err != nil {
		t.Fatalf("解析登录响应失败: %v", err)
	}
	if login.Code != 0 || login.Data.AccessToken == "" {
		t.Fatalf("登录响应异常: %+v", login)
	}

	// 使用 token 访问受保护端点（默认无权限 -> 403），RBAC 中间件会记录审计
	req, _ := http.NewRequest("GET", "http://"+addr+"/api/protected/ping", nil)
	req.Header.Set("Authorization", "Bearer "+login.Data.AccessToken)
	resp2, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("请求受保护端点失败: %v", err)
	}
	resp2.Body.Close()

	// 查询审计（可加上资源/动作过滤以增强稳定性）
	// 等待极短时间确保写入提交
	time.Sleep(30 * time.Millisecond)
	resp3, err := http.Get("http://" + addr + "/api/rbac/audit?limit=10&resource=%2Fapi%2Fprotected%2Fping&action=GET")
	if err != nil {
		t.Fatalf("查询失败: %v", err)
	}
	if resp3.StatusCode != http.StatusOK {
		t.Fatalf("审计查询期望200 实际 %d", resp3.StatusCode)
	}
	var out struct {
		Code int                      `json:"code"`
		Data []map[string]interface{} `json:"data"`
	}
	if err := json.NewDecoder(resp3.Body).Decode(&out); err != nil {
		t.Fatalf("解析失败: %v", err)
	}
	_ = resp3.Body.Close()
	if out.Code != 0 {
		t.Fatalf("响应码异常: %+v", out)
	}
	if len(out.Data) == 0 {
		t.Fatalf("期望返回至少1条审计记录，实际 0 条")
	}
}
