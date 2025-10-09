package interfaces

import (
	"context"
	"package-base/internal/dto"
)

// Function: AuthService
//
// Description: 认证服务接口，覆盖注册、登录、刷新、登出、黑名单管理等核心能力，为上层 Handler 提供稳定契约。
//
// Parameters:
//
//	无
//
// Returns:
//
//	无
//
// AuthService 认证服务接口
type AuthService interface {
	// Register 用户注册
	Register(ctx context.Context, req *dto.RegisterRequest) (*dto.AuthResponse, error)

	// Login 用户登录
	Login(ctx context.Context, req *dto.LoginRequest) (*dto.AuthResponse, error)

	// SignIn 语义别名：用户登录（内部可直接调用 Login，实现最小 SignIn 需求的独立可读性，供后续按任务 T011/T012 规划使用）
	SignIn(ctx context.Context, req *dto.LoginRequest) (*dto.AuthResponse, error)

	// RefreshToken 刷新Token
	RefreshToken(ctx context.Context, req *dto.RefreshTokenRequest) (*dto.TokenResponse, error)

	// Logout 用户登出（单设备）
	Logout(ctx context.Context, userID, sessionID string) error

	// LogoutAll 用户登出（所有设备）
	LogoutAll(ctx context.Context, userID string) error

	// ValidateToken 验证Token
	ValidateToken(ctx context.Context, token string) (*dto.UserInfo, error)

	// ChangePassword 修改密码
	ChangePassword(ctx context.Context, userID string, req *dto.ChangePasswordRequest) error

	// GetUserSessions 获取用户会话列表
	GetUserSessions(ctx context.Context, userID string) (*dto.SessionListResponse, error)

	// DeleteSession 删除指定会话
	DeleteSession(ctx context.Context, userID, sessionID string) error

	// GetCurrentUserInfo 获取当前用户信息
	GetCurrentUserInfo(ctx context.Context, userID string) (*dto.UserInfo, error)

	// GetActiveSessions 获取用户活跃会话
	GetActiveSessions(ctx context.Context, userID string) (*dto.SessionListResponse, error)

	// GetSessionsByDevice 根据设备信息获取会话
	GetSessionsByDevice(ctx context.Context, userID, deviceInfo string) (*dto.SessionListResponse, error)

	// GetRecentSessions 获取最近的会话（指定天数内）
	GetRecentSessions(ctx context.Context, userID string, days int) (*dto.SessionListResponse, error)

	// CountUserSessions 统计用户会话数量
	CountUserSessions(ctx context.Context, userID string, activeOnly bool) (int64, error)

	// Token黑名单管理方法

	// BlacklistToken 将token添加到黑名单
	BlacklistToken(ctx context.Context, token string) error

	// IsTokenBlacklisted 检查token是否在黑名单中
	IsTokenBlacklisted(ctx context.Context, token string) (bool, error)

	// ClearExpiredBlacklistTokens 清理过期的黑名单token
	ClearExpiredBlacklistTokens(ctx context.Context) error

	// GetBlacklistStats 获取黑名单统计信息
	GetBlacklistStats(ctx context.Context) (map[string]interface{}, error)
}
