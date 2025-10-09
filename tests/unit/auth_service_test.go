package unit

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"package-base/internal/core/interfaces"
	"package-base/internal/core/utils"
	"package-base/internal/dto"
	"package-base/internal/models"
	"package-base/internal/options"
	repoIf "package-base/internal/repositories/interfaces"
	"package-base/internal/services"
	serviceIf "package-base/internal/services/interfaces"

	"gorm.io/gorm"
)

// --- mock 实现 ---

type mockLogger struct{}

func (m *mockLogger) Debug(msg string, fields ...interfaces.Field)      {}
func (m *mockLogger) Info(msg string, fields ...interfaces.Field)       {}
func (m *mockLogger) Warn(msg string, fields ...interfaces.Field)       {}
func (m *mockLogger) Error(msg string, fields ...interfaces.Field)      {}
func (m *mockLogger) Fatal(msg string, fields ...interfaces.Field)      {}
func (m *mockLogger) Sync() error                                       { return nil }
func (m *mockLogger) With(fields ...interfaces.Field) interfaces.Logger { return m }
func (m *mockLogger) Named(name string) interfaces.Logger               { return m }

// mockUserRepo
type mockUserRepo struct {
	user      *models.User
	errFind   error
	errUpdate error
	created   *models.User
}

func (m *mockUserRepo) Create(ctx context.Context, user *models.User) error {
	m.created = user
	return nil
}
func (m *mockUserRepo) GetByID(ctx context.Context, id string) (*models.User, error) { return nil, nil }
func (m *mockUserRepo) GetByUsername(ctx context.Context, username string) (*models.User, error) {
	return nil, nil
}
func (m *mockUserRepo) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	return nil, nil
}
func (m *mockUserRepo) GetByUsernameOrEmail(ctx context.Context, s string) (*models.User, error) {
	return nil, nil
}
func (m *mockUserRepo) FindByUsernameOrEmail(ctx context.Context, s string) (*models.User, error) {
	return nil, nil
}
func (m *mockUserRepo) Update(ctx context.Context, user *models.User) error           { return m.errUpdate }
func (m *mockUserRepo) UpdatePassword(ctx context.Context, userID, hash string) error { return nil }
func (m *mockUserRepo) UpdateLoginInfo(ctx context.Context, userID string) error      { return nil }
func (m *mockUserRepo) UpdateLoginMeta(ctx context.Context, userID string) error      { return nil }
func (m *mockUserRepo) ExistsByUsername(ctx context.Context, username string) (bool, error) {
	return false, nil
}
func (m *mockUserRepo) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	return false, nil
}
func (m *mockUserRepo) Delete(ctx context.Context, id string) error { return nil }
func (m *mockUserRepo) GetList(ctx context.Context, offset, limit int) ([]*models.User, int64, error) {
	return nil, 0, nil
}
func (m *mockUserRepo) UpdateStatus(ctx context.Context, userID string, status models.UserStatus) error {
	return nil
}
func (m *mockUserRepo) FindUsers(ctx context.Context, opts ...options.QueryUserOption) ([]*models.User, error) {
	return nil, nil
}
func (m *mockUserRepo) FindUser(ctx context.Context, opts ...options.QueryUserOption) (*models.User, error) {
	// 解析 options 中是否包含 ID/email/username 等，这里简单返回 m.user (模拟匹配到)
	return m.user, m.errFind
}
func (m *mockUserRepo) CountUsers(ctx context.Context, opts ...options.QueryUserOption) (int64, error) {
	return 0, nil
}
func (m *mockUserRepo) FindUsersWithPagination(ctx context.Context, opts ...options.QueryUserOption) ([]*models.User, int64, error) {
	return nil, 0, nil
}

// mockSessionRepo
type mockSessionRepo struct {
	createErr error
	created   *models.Session
	// 以下字段用于 RefreshToken 场景
	sessionByRefresh *models.Session
	sessionFetchErr  error

	// 额外字段用于更多方法覆盖
	sessionByID             *models.Session
	activeByUser            []*models.Session
	deactivatedSessionIDs   []string
	deactivatedAllForUserID []string
	countSessionsReturn     int64
	countSessionsErr        error
}

func (m *mockSessionRepo) Create(ctx context.Context, s *models.Session) error {
	if m.createErr != nil {
		return m.createErr
	}
	m.created = s
	// 初始赋值方便后续通过 refresh_token 查找
	m.sessionByRefresh = s
	return nil
}
func (m *mockSessionRepo) GetByID(ctx context.Context, id string) (*models.Session, error) {
	if m.sessionByID != nil && (m.sessionByID.ID == "" || m.sessionByID.ID == id) {
		return m.sessionByID, nil
	}
	return nil, gorm.ErrRecordNotFound
}
func (m *mockSessionRepo) GetByRefreshToken(ctx context.Context, t string) (*models.Session, error) {
	if m.sessionFetchErr != nil {
		return nil, m.sessionFetchErr
	}
	return m.sessionByRefresh, nil
}
func (m *mockSessionRepo) GetByUserID(ctx context.Context, userID string) ([]*models.Session, error) {
	return nil, nil
}
func (m *mockSessionRepo) GetActiveByUserID(ctx context.Context, userID string) ([]*models.Session, error) {
	if m.activeByUser != nil {
		return m.activeByUser, nil
	}
	return []*models.Session{}, nil
}
func (m *mockSessionRepo) Update(ctx context.Context, s *models.Session) error {
	// 覆盖内部引用，便于测试断言变化
	if m.sessionByRefresh != nil && m.sessionByRefresh.ID == s.ID {
		m.sessionByRefresh = s
	}
	return nil
}
func (m *mockSessionRepo) UpdateAccessToken(ctx context.Context, sessionID, accessToken string) error {
	return nil
}
func (m *mockSessionRepo) DeactivateSession(ctx context.Context, id string) error { return nil }
func (m *mockSessionRepo) DeactivateUserSessions(ctx context.Context, userID string) error {
	m.deactivatedAllForUserID = append(m.deactivatedAllForUserID, userID)
	return nil
}
func (m *mockSessionRepo) DeactivateOtherSessions(ctx context.Context, userID, currentSessionID string) error {
	return nil
}
func (m *mockSessionRepo) Delete(ctx context.Context, sessionID string) error { return nil }
func (m *mockSessionRepo) DeleteExpiredSessions(ctx context.Context) error    { return nil }
func (m *mockSessionRepo) CleanupInactiveSessions(ctx context.Context) error  { return nil }
func (m *mockSessionRepo) CountActiveSessionsByUserID(ctx context.Context, userID string) (int64, error) {
	return 0, nil
}
func (m *mockSessionRepo) GetSessionWithUser(ctx context.Context, sessionID string) (*models.Session, error) {
	return nil, nil
}
func (m *mockSessionRepo) FindSessions(ctx context.Context, opts ...options.QuerySessionOption) ([]*models.Session, error) {
	if m.activeByUser != nil {
		return m.activeByUser, nil
	}
	return []*models.Session{}, nil
}
func (m *mockSessionRepo) FindSession(ctx context.Context, opts ...options.QuerySessionOption) (*models.Session, error) {
	return nil, nil
}
func (m *mockSessionRepo) CountSessions(ctx context.Context, opts ...options.QuerySessionOption) (int64, error) {
	if m.countSessionsErr != nil {
		return 0, m.countSessionsErr
	}
	return m.countSessionsReturn, nil
}
func (m *mockSessionRepo) FindSessionsWithPagination(ctx context.Context, page, pageSize int, opts ...options.QuerySessionOption) ([]*models.Session, int64, error) {
	return nil, 0, nil
}

// Function: RotateTokensCAS
//
// Description: 模拟基于旧的 refresh token 原子更新为新的一对 token。若旧 token 不匹配或 session 不存在返回 false。
//
// Parameters:
//
//	ctx(context.Context): 上下文
//	oldRefreshToken(string): 旧的 refresh token
//	newRefreshToken(string): 新的 refresh token
//	newAccessToken(string): 新的 access token
//	newExpiresAt(int64): 新的过期时间（unix 秒）
//
// Returns:
//
//	updated(bool): 是否成功更新（模拟 CAS 成功）
//	err(error): 错误信息（此 mock 恒为 nil）
func (m *mockSessionRepo) RotateTokensCAS(ctx context.Context, oldRefreshToken, newRefreshToken, newAccessToken string, newExpiresAt int64) (bool, error) {
	if m.sessionByRefresh == nil || m.sessionByRefresh.RefreshToken != oldRefreshToken {
		return false, nil
	}
	m.sessionByRefresh.RefreshToken = newRefreshToken
	m.sessionByRefresh.AccessToken = newAccessToken
	// ExpiresAt 为 time.Time 类型，newExpiresAt 是 unix 秒
	m.sessionByRefresh.ExpiresAt = time.Unix(newExpiresAt, 0)
	return true, nil
}

// mockTokenBlacklistRepo
type mockTokenBlacklistRepo struct {
	err   error
	added []string
	size  int64
}

func (m *mockTokenBlacklistRepo) AddToBlacklist(ctx context.Context, token string, expiration time.Duration) error {
	m.added = append(m.added, token)
	return m.err
}
func (m *mockTokenBlacklistRepo) IsBlacklisted(ctx context.Context, token string) (bool, error) {
	for _, t := range m.added {
		if t == token {
			return true, nil
		}
	}
	return false, nil
}
func (m *mockTokenBlacklistRepo) RemoveFromBlacklist(ctx context.Context, token string) error {
	return nil
}
func (m *mockTokenBlacklistRepo) ClearExpiredTokens(ctx context.Context) error { return nil }
func (m *mockTokenBlacklistRepo) GetBlacklistSize(ctx context.Context) (int64, error) {
	return m.size, nil
}
func (m *mockTokenBlacklistRepo) GetBlacklistedTokens(ctx context.Context, limit int) ([]string, error) {
	return m.added, nil
}

// helper 创建 service
func newAuthServiceForTest(usr repoIf.UserRepository, sess repoIf.SessionRepository, bl repoIf.TokenBlacklistRepository) serviceIf.AuthService {
	jwtSecret := "test-secret"
	access := 900 * time.Second
	refresh := 3600 * time.Second
	return services.NewAuthService(usr, sess, bl, jwtSecret, access, refresh, &mockLogger{})
}

// --- 测试用例 ---

func TestAuthService_Login_Success(t *testing.T) {
	password := "Passw0rd!"
	hash, _ := utils.HashPassword(password)
	user := &models.User{ID: "u1", Username: "tester", PasswordHash: &hash, Status: models.UserStatusActive, Roles: models.UserRoleTypeUser}
	usrRepo := &mockUserRepo{user: user}
	sessRepo := &mockSessionRepo{}
	blRepo := &mockTokenBlacklistRepo{}

	svc := newAuthServiceForTest(usrRepo, sessRepo, blRepo)
	resp, err := svc.Login(context.Background(), &dto.LoginRequest{Username: "tester", Password: password})
	if err != nil {
		t.Fatalf("期望成功, 实际错误: %v", err)
	}
	if resp.AccessToken == "" || resp.RefreshToken == "" {
		t.Fatalf("token 为空: %+v", resp)
	}
	if sessRepo.created == nil {
		t.Fatalf("未创建 session")
	}
}

func TestAuthService_Login_UserNotFound(t *testing.T) {
	usrRepo := &mockUserRepo{user: nil}
	sessRepo := &mockSessionRepo{}
	svc := newAuthServiceForTest(usrRepo, sessRepo, &mockTokenBlacklistRepo{})
	_, err := svc.Login(context.Background(), &dto.LoginRequest{Username: "ghost", Password: "x"})
	if err == nil || err.Error() != "invalid credentials" {
		t.Fatalf("期望 invalid credentials 实际: %v", err)
	}
}

func TestAuthService_Login_WrongPassword(t *testing.T) {
	goodHash, _ := utils.HashPassword("Passw0rd!")
	user := &models.User{ID: "u1", Username: "tester", PasswordHash: &goodHash, Status: models.UserStatusActive}
	usrRepo := &mockUserRepo{user: user}
	sessRepo := &mockSessionRepo{}
	svc := newAuthServiceForTest(usrRepo, sessRepo, &mockTokenBlacklistRepo{})
	_, err := svc.Login(context.Background(), &dto.LoginRequest{Username: "tester", Password: "Bad123!"})
	if err == nil || err.Error() != "invalid credentials" {
		t.Fatalf("期望 invalid credentials 实际: %v", err)
	}
}

func TestAuthService_Login_Inactive(t *testing.T) {
	hash, _ := utils.HashPassword("Passw0rd!")
	inactive := &models.User{ID: "u1", Username: "tester", PasswordHash: &hash, Status: models.UserStatusInactive}
	usrRepo := &mockUserRepo{user: inactive}
	sessRepo := &mockSessionRepo{}
	svc := newAuthServiceForTest(usrRepo, sessRepo, &mockTokenBlacklistRepo{})
	_, err := svc.Login(context.Background(), &dto.LoginRequest{Username: "tester", Password: "Passw0rd!"})
	if err == nil || err.Error() != "account is not active" {
		t.Fatalf("期望 account is not active 实际: %v", err)
	}
}

func TestAuthService_Login_SessionCreateFail(t *testing.T) {
	hash, _ := utils.HashPassword("Passw0rd!")
	user := &models.User{ID: "u1", Username: "tester", PasswordHash: &hash, Status: models.UserStatusActive}
	usrRepo := &mockUserRepo{user: user}
	origErr := errors.New("db down")
	sessRepo := &mockSessionRepo{createErr: origErr}
	svc := newAuthServiceForTest(usrRepo, sessRepo, &mockTokenBlacklistRepo{})
	_, err := svc.Login(context.Background(), &dto.LoginRequest{Username: "tester", Password: "Passw0rd!"})
	if err == nil || !strings.Contains(err.Error(), origErr.Error()) {
		t.Fatalf("期望包含 db down 错误, 实际: %v", err)
	}
}

func TestAuthService_RefreshToken_Success(t *testing.T) {
	// 1. 先通过登录生成真实 refresh token
	loginHash, _ := utils.HashPassword("Passw0rd!")
	user := &models.User{ID: "u1", Username: "tester", PasswordHash: &loginHash, Status: models.UserStatusActive, Roles: models.UserRoleTypeUser}
	usrRepo := &mockUserRepo{user: user}
	sessRepo := &mockSessionRepo{}
	svc := newAuthServiceForTest(usrRepo, sessRepo, &mockTokenBlacklistRepo{})
	loginResp, err := svc.Login(context.Background(), &dto.LoginRequest{Username: "tester", Password: "Passw0rd!"})
	if err != nil {
		t.Fatalf("预登录失败: %v", err)
	}
	if sessRepo.created == nil {
		t.Fatalf("预登录未创建 session")
	}
	// 模拟从仓库按 refresh_token 取回 session
	sessRepo.sessionByRefresh = sessRepo.created

	// 2. 调用 RefreshToken
	oldRT := loginResp.RefreshToken
	resp, err := svc.RefreshToken(context.Background(), &dto.RefreshTokenRequest{RefreshToken: oldRT})
	if err != nil {
		t.Fatalf("期望成功 实际错误: %v", err)
	}
	if resp.AccessToken == "" || resp.RefreshToken == "" {
		t.Fatalf("返回 token 不应为空: %+v", resp)
	}
	if resp.RefreshToken == oldRT {
		t.Fatalf("应生成新的 refresh token, old=%s new=%s", oldRT, resp.RefreshToken)
	}
}

func TestAuthService_RefreshToken_InvalidToken(t *testing.T) {
	usrRepo := &mockUserRepo{}
	sessRepo := &mockSessionRepo{sessionByRefresh: nil}
	svc := newAuthServiceForTest(usrRepo, sessRepo, &mockTokenBlacklistRepo{})
	_, err := svc.RefreshToken(context.Background(), &dto.RefreshTokenRequest{RefreshToken: "bad"})
	if err == nil || err.Error() != "invalid refresh token" {
		t.Fatalf("期望 invalid refresh token 实际: %v", err)
	}
}

func TestAuthService_RefreshToken_SessionExpired(t *testing.T) {
	// 使用真实 refresh token 然后将 session 标记过期
	hash, _ := utils.HashPassword("Passw0rd!")
	user := &models.User{ID: "u1", Username: "tester", PasswordHash: &hash, Status: models.UserStatusActive}
	usrRepo := &mockUserRepo{user: user}
	sessRepo := &mockSessionRepo{}
	svc := newAuthServiceForTest(usrRepo, sessRepo, &mockTokenBlacklistRepo{})
	loginResp, err := svc.Login(context.Background(), &dto.LoginRequest{Username: "tester", Password: "Passw0rd!"})
	if err != nil {
		t.Fatalf("预登录失败: %v", err)
	}
	if sessRepo.created == nil {
		t.Fatalf("预登录未创建 session")
	}
	// 让会话过期
	sessRepo.created.ExpiresAt = time.Now().Add(-1 * time.Minute)
	sessRepo.sessionByRefresh = sessRepo.created
	_, err = svc.RefreshToken(context.Background(), &dto.RefreshTokenRequest{RefreshToken: loginResp.RefreshToken})
	if err == nil || err.Error() != "session expired" {
		t.Fatalf("期望 session expired 实际: %v", err)
	}
}

// Function: TestAuthService_LogoutAndBlacklist
//
// Description: 覆盖 Logout 将会话 access token 加入黑名单的路径，验证黑名单仓库被调用。
//
// Parameters:
//
//	t(*testing.T): 测试句柄
//
// Returns:
//
//	无
func TestAuthService_LogoutAndBlacklist(t *testing.T) {
	// 构造带有 access token 的 session
	hash, _ := utils.HashPassword("Passw0rd!")
	user := &models.User{ID: "u1", Username: "tester", PasswordHash: &hash, Status: models.UserStatusActive}
	usrRepo := &mockUserRepo{user: user}
	sess := &models.Session{ID: "s1", UserID: user.ID, AccessToken: "at-1", IsActive: true}
	sessRepo := &mockSessionRepo{sessionByID: sess}
	bl := &mockTokenBlacklistRepo{}
	svc := newAuthServiceForTest(usrRepo, sessRepo, bl)

	if err := svc.Logout(context.Background(), user.ID, "s1"); err != nil {
		t.Fatalf("Logout 失败: %v", err)
	}
	if len(bl.added) == 0 || bl.added[0] != "at-1" {
		t.Fatalf("应将 access token 加入黑名单, 实际: %+v", bl.added)
	}
}

// Function: TestAuthService_LogoutAll
//
// Description: 覆盖 LogoutAll 遍历活跃会话并将 access token 加入黑名单的逻辑。
//
// Parameters:
//
//	t(*testing.T): 测试句柄
//
// Returns:
//
//	无
func TestAuthService_LogoutAll(t *testing.T) {
	hash, _ := utils.HashPassword("Passw0rd!")
	user := &models.User{ID: "u1", Username: "tester", PasswordHash: &hash, Status: models.UserStatusActive}
	usrRepo := &mockUserRepo{user: user}
	active := []*models.Session{{ID: "s1", UserID: user.ID, AccessToken: "at-1", IsActive: true}}
	sessRepo := &mockSessionRepo{activeByUser: active}
	bl := &mockTokenBlacklistRepo{}
	svc := newAuthServiceForTest(usrRepo, sessRepo, bl)

	if err := svc.LogoutAll(context.Background(), user.ID); err != nil {
		t.Fatalf("LogoutAll 失败: %v", err)
	}
	if len(bl.added) == 0 {
		t.Fatalf("应将活跃会话的 access token 加入黑名单")
	}
}

// Function: TestAuthService_ValidateToken_Success
//
// Description: 验证有效的 access token 能被成功解析并返回用户信息。
//
// Parameters:
//
//	t(*testing.T): 测试句柄
//
// Returns:
//
//	无
func TestAuthService_ValidateToken_Success(t *testing.T) {
	hash, _ := utils.HashPassword("Passw0rd!")
	user := &models.User{ID: "u1", Username: "tester", PasswordHash: &hash, Status: models.UserStatusActive, Roles: models.UserRoleTypeUser}
	usrRepo := &mockUserRepo{user: user}
	sessRepo := &mockSessionRepo{}
	// 使用相同 secret，避免签名不一致
	svc := services.NewAuthService(usrRepo, sessRepo, &mockTokenBlacklistRepo{}, "secret", 5*time.Minute, 30*time.Minute, &mockLogger{})
	jm := utils.NewJWTManager("secret", 5*time.Minute, 30*time.Minute)
	at, _, _ := jm.GenerateAccessToken(user.ID, user.Username, "t@example.com", int(user.Roles))
	info, err := svc.ValidateToken(context.Background(), at)
	if err != nil || info == nil || info.ID != user.ID {
		t.Fatalf("ValidateToken 失败: info=%+v err=%v", info, err)
	}
}

// Function: TestAuthService_ChangePassword
//
// Description: 覆盖修改密码流程（含旧密码校验、hash 新密码、停用会话）。
//
// Parameters:
//
//	t(*testing.T): 测试句柄
//
// Returns:
//
//	无
func TestAuthService_ChangePassword(t *testing.T) {
	oldHash, _ := utils.HashPassword("OldP@ss1!")
	user := &models.User{ID: "u1", Username: "tester", PasswordHash: &oldHash, Status: models.UserStatusActive}
	usrRepo := &mockUserRepo{user: user}
	sessRepo := &mockSessionRepo{}
	svc := newAuthServiceForTest(usrRepo, sessRepo, &mockTokenBlacklistRepo{})
	err := svc.ChangePassword(context.Background(), user.ID, &dto.ChangePasswordRequest{OldPassword: "OldP@ss1!", NewPassword: "NewP@ss1!"})
	if err != nil {
		t.Fatalf("ChangePassword 失败: %v", err)
	}
}

// Function: TestAuthService_SessionQueriesAndCounts
//
// Description: 覆盖会话查询与统计接口（活跃会话、按设备、最近会话、计数）。
//
// Parameters:
//
//	t(*testing.T): 测试句柄
//
// Returns:
//
//	无
func TestAuthService_SessionQueriesAndCounts(t *testing.T) {
	userID := "u1"
	sessRepo := &mockSessionRepo{
		activeByUser:        []*models.Session{{ID: "s1", UserID: userID, IsActive: true}},
		countSessionsReturn: 3,
	}
	svc := newAuthServiceForTest(&mockUserRepo{}, sessRepo, &mockTokenBlacklistRepo{})
	// GetActiveSessions
	if resp, err := svc.GetActiveSessions(context.Background(), userID); err != nil || resp.Total == 0 {
		t.Fatalf("GetActiveSessions 失败: %+v err=%v", resp, err)
	}
	// GetSessionsByDevice
	if resp, err := svc.GetSessionsByDevice(context.Background(), userID, "Mac"); err != nil {
		t.Fatalf("GetSessionsByDevice 失败: %v", err)
	} else if resp.Total == 0 {
		// 此处返回 activeByUser 的数据
		t.Fatalf("GetSessionsByDevice 期望有数据")
	}
	// GetRecentSessions
	if resp, err := svc.GetRecentSessions(context.Background(), userID, 7); err != nil || resp.Total == 0 {
		t.Fatalf("GetRecentSessions 失败: %+v err=%v", resp, err)
	}
	// CountUserSessions
	if n, err := svc.CountUserSessions(context.Background(), userID, true); err != nil || n != 3 {
		t.Fatalf("CountUserSessions 返回不符合预期: %d err=%v", n, err)
	}
}

// Function: TestAuthService_BlacklistAndStats
//
// Description: 覆盖黑名单相关方法（加入、检查、清理、统计）。
//
// Parameters:
//
//	t(*testing.T): 测试句柄
//
// Returns:
//
//	无
func TestAuthService_BlacklistAndStats(t *testing.T) {
	// 使用真实 access token，验证 BlacklistToken 能解析 exp 并写入
	jm := utils.NewJWTManager("secret", 1*time.Minute, 30*time.Minute)
	at, _, _ := jm.GenerateAccessToken("u1", "tester", "t@example.com", 0)
	bl := &mockTokenBlacklistRepo{size: 5}
	svc := services.NewAuthService(nil, nil, bl, "secret", 1*time.Minute, 30*time.Minute, &mockLogger{})

	if err := svc.BlacklistToken(context.Background(), at); err != nil {
		t.Fatalf("BlacklistToken 失败: %v", err)
	}
	if ok, _ := svc.IsTokenBlacklisted(context.Background(), at); !ok {
		t.Fatalf("应被列入黑名单")
	}
	if err := svc.ClearExpiredBlacklistTokens(context.Background()); err != nil {
		t.Fatalf("ClearExpiredBlacklistTokens 失败: %v", err)
	}
	if stats, err := svc.GetBlacklistStats(context.Background()); err != nil {
		t.Fatalf("GetBlacklistStats 失败: %v", err)
	} else if stats["total_blacklisted_tokens"].(int64) != 5 {
		t.Fatalf("统计 size 不匹配: %+v", stats)
	}
}

// Function: TestAuthService_Getters
//
// Description: 覆盖用户信息获取、用户会话列表与删除会话等查询类方法。
//
// Parameters:
//
//	t(*testing.T): 测试句柄
//
// Returns:
//
//	无
func TestAuthService_Getters(t *testing.T) {
	// GetCurrentUserInfo
	user := &models.User{ID: "u1", Username: "tester", Status: models.UserStatusActive}
	usrRepo := &mockUserRepo{user: user}
	svc := newAuthServiceForTest(usrRepo, &mockSessionRepo{}, &mockTokenBlacklistRepo{})
	if info, err := svc.GetCurrentUserInfo(context.Background(), user.ID); err != nil || info == nil || info.ID != user.ID {
		t.Fatalf("GetCurrentUserInfo 失败: %+v err=%v", info, err)
	}

	// GetUserSessions / DeleteSession 依赖 SessionRepo 行为
	sess := &models.Session{ID: "s1", UserID: user.ID}
	srepo := &mockSessionRepo{sessionByID: sess}
	svc2 := newAuthServiceForTest(usrRepo, srepo, &mockTokenBlacklistRepo{})
	if resp, err := svc2.GetUserSessions(context.Background(), user.ID); err != nil {
		t.Fatalf("GetUserSessions 失败: %v", err)
	} else if resp == nil {
		t.Fatalf("GetUserSessions 返回为空")
	}
	if err := svc2.DeleteSession(context.Background(), user.ID, "s1"); err != nil {
		t.Fatalf("DeleteSession 失败: %v", err)
	}
}
