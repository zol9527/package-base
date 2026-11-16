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
	"package-base/internal/core/utils"
	"package-base/internal/models"
	"package-base/internal/routes"
	"package-base/tests/testhelpers"

	"github.com/gofiber/fiber/v2"
	"golang.org/x/crypto/bcrypt"
)

// Function: Test_Protected_ForbiddenByDefault
//
// Description: 默认拒绝策略下，即使已认证也应对未分配权限的资源返回 403。
//
// Parameters:
//
//	t(*testing.T): 测试句柄
//
// Returns:
//
//	无
func Test_Protected_ForbiddenByDefault(t *testing.T) {
	// 启动应用
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

	// 迁移与种子用户（使用独立用户名 rbac_user1 避免干扰）
	if err := cont.GetMySQLStorage().Migrate(t.Context(), &models.User{}); err != nil {
		t.Fatalf("用户模型迁移失败: %v", err)
	}
	if err := testhelpers.EnsureSessionTable(cont); err != nil {
		t.Fatalf("创建会话表失败: %v", err)
	}
	password := "Passw0rd!"
	db := cont.GetMySQLStorage().GetDB()
	var count int64
	if err := db.Model(&models.User{}).Where("username=?", "rbac_user1").Count(&count).Error; err != nil {
		t.Fatalf("检查用户存在失败: %v", err)
	}
	if count == 0 {
		hash, _ := bcrypt.GenerateFromPassword([]byte(password), 10)
		uuid := utils.GenerateShortUUID()
		if len(uuid) > 22 {
			uuid = uuid[:22]
		}
		insertSQL := "INSERT INTO users (uuid, username, email, password_hash, status, roles, created_at, updated_at) VALUES (?,?,?,?,?,?,NOW(),NOW())"
		if err := db.Exec(insertSQL, uuid, "rbac_user1", "rbac_user1@example.com", string(hash), 1, 0).Error; err != nil {
			t.Fatalf("插入种子用户失败: %v", err)
		}
	}

	app := fiber.New()
	r := routes.NewRoutes(app, cont)
	r.Setup()
	addr := "127.0.0.1:18086"
	go app.Listen(addr)
	defer app.Shutdown()

	// 登录获取 token
	body, _ := json.Marshal(map[string]string{"username": "rbac_user1", "password": password})
	resp, err := http.Post("http://"+addr+"/api/auth/login", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("登录失败: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("登录期望 200 实际 %d", resp.StatusCode)
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

	// 携带 token 请求受保护端点，应 403（Authorize 默认返回 false）
	req, _ := http.NewRequest("GET", "http://"+addr+"/api/protected/ping", nil)
	req.Header.Set("Authorization", "Bearer "+login.Data.AccessToken)
	resp2, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("请求失败: %v", err)
	}
	defer resp2.Body.Close()
	if resp2.StatusCode != http.StatusForbidden {
		t.Fatalf("期望 403 实际 %d", resp2.StatusCode)
	}
}
