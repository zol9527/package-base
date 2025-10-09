package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"package-base/internal/container"
	"package-base/internal/core/config"
	"package-base/internal/core/log"
	"package-base/internal/models"
	"package-base/internal/routes"
	"package-base/tests/testhelpers"

	"github.com/gofiber/fiber/v2"
)

// Test_AuthLoginSuccess 成功登录返回 token (US2 T028)
func Test_AuthLoginSuccess(t *testing.T) {
	// 准备配置路径
	wd, _ := os.Getwd()
	root := filepath.Clean(filepath.Join(wd, "../../"))
	configPath := filepath.Join(root, "config.json")
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		t.Fatalf("加载配置失败: %v", err)
	}

	// 创建 logger（测试禁用文件输出）
	loggerCfg := &log.LoggerConfig{Level: cfg.Logger.Level, Format: cfg.Logger.Format, EnableConsole: false, EnableFile: false}
	logger, err := log.NewLogger(loggerCfg)
	if err != nil {
		t.Fatalf("创建 logger 失败: %v", err)
	}

	cont := container.NewContainer(cfg, logger)
	if err := cont.Initialize(t.Context()); err != nil {
		t.Fatalf("容器初始化失败: %v", err)
	}
	defer cont.Close()

	// 仅迁移用户表（会话表因现有 users.id 为 int 与模型 string 外键不兼容，单独手工创建）
	if err := cont.GetMySQLStorage().Migrate(t.Context(), &models.User{}); err != nil {
		t.Fatalf("用户模型迁移失败: %v", err)
	}
	if err := testhelpers.EnsureSessionTable(cont); err != nil {
		t.Fatalf("创建 user_sessions 表失败: %v", err)
	}

	// 准备测试用户（若不存在则使用 raw SQL 插入，避免 GORM 对 string 主键的插入逻辑触发 id 列冲突）
	password := "Passw0rd!"
	if err := testhelpers.SeedTestUser(t, cont, password); err != nil {
		t.Fatalf("插入测试用户失败: %v", err)
	}

	// 启动 Fiber 应用并注册路由
	app := fiber.New()
	r := routes.NewRoutes(app, cont)
	r.Setup()
	addr := "127.0.0.1:18081" // 避免与主 8080 冲突
	go app.Listen(addr)
	defer app.Shutdown()

	// 构造请求
	reqBody, _ := json.Marshal(map[string]string{"username": "tester", "password": password})
	resp, err := http.Post("http://"+addr+"/api/auth/login", "application/json", bytes.NewReader(reqBody))
	if err != nil {
		t.Fatalf("请求登录失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("期望 200 实际 %d", resp.StatusCode)
	}

	var parsed struct {
		Code    int             `json:"code"`
		Message string          `json:"msg"`
		Data    json.RawMessage `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		t.Fatalf("解析响应失败: %v", err)
	}
	if parsed.Code != 0 {
		t.Fatalf("期望 code=0 实际 %d message=%s", parsed.Code, parsed.Message)
	}

	// 解析 data
	var authData struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int64  `json:"expires_in"`
		User        struct {
			Username string `json:"username"`
		} `json:"user"`
	}
	if err := json.Unmarshal(parsed.Data, &authData); err != nil {
		t.Fatalf("解析 data 失败: %v", err)
	}
	if authData.AccessToken == "" || authData.ExpiresIn <= 0 || authData.User.Username != "tester" {
		t.Fatalf("返回数据不符合预期: %+v", authData)
	}

	// 追加断言: user_sessions 表存在对应会话 & users.login_count 已自增
	// 等待少许时间以确保事务提交
	time.Sleep(50 * time.Millisecond)
	db := cont.GetMySQLStorage().GetDB()
	var sessionCount int64
	if err := db.Table("user_sessions").Where("user_id = (SELECT id FROM users WHERE username=?)", "tester").Count(&sessionCount).Error; err != nil {
		t.Fatalf("查询会话计数失败: %v", err)
	}
	if sessionCount == 0 {
		t.Fatalf("期望至少 1 条会话记录, 实际 0")
	}
	var loginCount int64
	if err := db.Table("users").Select("login_count").Where("username=?", "tester").Scan(&loginCount).Error; err != nil {
		t.Fatalf("查询 login_count 失败: %v", err)
	}
	if loginCount < 1 {
		t.Fatalf("期望 login_count >=1 实际 %d", loginCount)
	}
}

// 原始 seedTestUser / ensureSessionTable 已迁移至 testhelpers
