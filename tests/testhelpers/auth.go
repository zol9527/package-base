package testhelpers

import (
	"net"
	"package-base/internal/container"
	"package-base/internal/core/utils"
	"package-base/internal/models"
	"testing"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// Function: SeedTestUser
//
// Description: 如果测试用户 tester 不存在，则插入一条带有指定密码的用户记录；兼容 legacy users.id(int)+uuid(22) 场景。
//
// Parameters:
//
//	cont(*container.Container): 依赖注入容器，提供数据库实例
//	password(string): 明文密码
//	t(*testing.T): 测试句柄用于帮助与失败输出
//
// Returns:
//
//	err(error): 可能的错误
func SeedTestUser(t *testing.T, cont *container.Container, password string) error {
	t.Helper()
	db := cont.GetMySQLStorage().GetDB()
	var count int64
	if err := db.Model(&models.User{}).Where("username = ?", "tester").Count(&count).Error; err == nil && count > 0 {
		return nil
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), 10)
	if err != nil {
		return err
	}
	uuidFull := utils.GenerateShortUUID()
	uuid := uuidFull
	if len(uuid) > 22 {
		uuid = uuid[:22]
	}
	insertSQL := "INSERT INTO users (uuid, username, email, password_hash, status, roles, created_at, updated_at) VALUES (?,?,?,?,?,?,NOW(),NOW())"
	return db.Exec(insertSQL, uuid, "tester", "tester@example.com", string(hash), 1, 0).Error
}

// Function: EnsureSessionTable
//
// Description: 如果不存在 user_sessions 表则创建（无外键，user_id 使用 INT 以匹配既有 users.id）。
//
// Parameters:
//
//	cont(*container.Container): 容器
//
// Returns:
//
//	err(error): 可能的错误
func EnsureSessionTable(cont *container.Container) error {
	db := cont.GetMySQLStorage().GetDB()
	var count int64
	if err := db.Raw("SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = DATABASE() AND table_name = 'user_sessions'").Scan(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return nil
	}
	createSQL := `CREATE TABLE user_sessions (
        id VARCHAR(36) PRIMARY KEY,
        user_id INT NOT NULL,
        device_info VARCHAR(500),
        ip_address VARCHAR(45),
        user_agent VARCHAR(500),
        refresh_token VARCHAR(255),
        access_token VARCHAR(500),
        expires_at DATETIME(3) NOT NULL,
        is_active TINYINT(1) DEFAULT 1,
        created_at DATETIME(3) NULL,
        updated_at DATETIME(3) NULL,
        UNIQUE KEY idx_user_sessions_refresh_token (refresh_token),
        KEY idx_user_sessions_user_id (user_id)
    ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;`
	return db.Exec(createSQL).Error
}

// Function: WaitForServer
//
// Description: 等待指定 TCP 地址监听就绪，轮询尝试连接直到成功或超时，用于集成测试在发起 HTTP 请求前确保 Fiber 服务已启动。
//
// Parameters:
//
//	addr(string): 目标地址 host:port
//	timeout(time.Duration): 最长等待时间
//	interval(time.Duration): 轮询间隔
//
// Returns:
//
//	ready(bool): 是否在超时前就绪
//	err(error): 最后一次连接错误（若未就绪）
func WaitForServer(addr string, timeout, interval time.Duration) (bool, error) {
	deadline := time.Now().Add(timeout)
	for {
		conn, err := net.DialTimeout("tcp", addr, interval)
		if err == nil {
			conn.Close()
			return true, nil
		}
		if time.Now().After(deadline) {
			return false, err
		}
		time.Sleep(interval)
	}
}
