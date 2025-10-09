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

// Test_AuthLoginWrongPassword 错误密码返回 401 (US2 T029)
func Test_AuthLoginWrongPassword(t *testing.T) {
	wd, _ := os.Getwd()
	root := filepath.Clean(filepath.Join(wd, "../../"))
	configPath := filepath.Join(root, "config.json")
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		t.Fatalf("加载配置失败: %v", err)
	}

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

	if err := cont.GetMySQLStorage().Migrate(t.Context(), &models.User{}); err != nil {
		t.Fatalf("用户模型迁移失败: %v", err)
	}
	if err := testhelpers.EnsureSessionTable(cont); err != nil {
		t.Fatalf("创建 user_sessions 表失败: %v", err)
	}

	// 种子用户（正确密码）
	realPassword := "Passw0rd!"
	if err := seedTestUserWithCustom(cont, "tester_wrong_pw", realPassword); err != nil {
		t.Fatalf("插入测试用户失败: %v", err)
	}

	app := fiber.New()
	r := routes.NewRoutes(app, cont)
	r.Setup()
	addr := "127.0.0.1:18082"
	go app.Listen(addr)
	defer app.Shutdown()

	// 使用错误密码
	reqBody, _ := json.Marshal(map[string]string{"username": "tester_wrong_pw", "password": "Wrong123!"})
	resp, err := http.Post("http://"+addr+"/api/auth/login", "application/json", bytes.NewReader(reqBody))
	if err != nil {
		t.Fatalf("请求失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("期望 401 实际 %d", resp.StatusCode)
	}

	var generic map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&generic); err != nil {
		t.Fatalf("解析响应失败: %v", err)
	}
	codeField, ok := generic["code"]
	if !ok {
		t.Fatalf("响应缺少 code 字段: %+v", generic)
	}
	valid := false
	switch v := codeField.(type) {
	case float64:
		if int(v) == 401 {
			valid = true
		}
	case string:
		// 新统一错误码为 AUTH_xxx，HTTP 状态已验证
		if v == "AUTH_INVALID_CREDENTIALS" || v == "AUTH_LOGIN_FAILED" || v == "AUTH_INVALID_REQUEST" {
			valid = true
		}
	}
	if !valid {
		t.Fatalf("意外的业务 code: %v", codeField)
	}
}

// seedTestUserWithCustom 插入指定用户名的测试用户
func seedTestUserWithCustom(cont *container.Container, username, password string) error {
	db := cont.GetMySQLStorage().GetDB()
	var count int64
	if err := db.Model(&models.User{}).Where("username = ?", username).Count(&count).Error; err == nil && count > 0 {
		return nil
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), 10)
	if err != nil {
		return err
	}
	uuidFull := utils.GenerateShortUUID()
	if len(uuidFull) > 22 {
		uuidFull = uuidFull[:22]
	}
	insertSQL := "INSERT INTO users (uuid, username, email, password_hash, status, roles, created_at, updated_at) VALUES (?,?,?,?,?,?,NOW(),NOW())"
	return db.Exec(insertSQL, uuidFull, username, username+"@example.com", string(hash), 1, 0).Error
}
