package dto

import "time"

// ============================================================================
// 认证相关DTO
// ============================================================================

// RegisterRequest 用户注册请求
type RegisterRequest struct {
	Username string `json:"username" validate:"required,min=3,max=50,alphanum"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8,max=100"`
	Phone    string `json:"phone,omitempty" validate:"omitempty,e164"`
	Nickname string `json:"nickname,omitempty" validate:"omitempty,min=1,max=50"`
}

// LoginRequest 用户登录请求
type LoginRequest struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}

// RefreshTokenRequest Token刷新请求
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

// ChangePasswordRequest 修改密码请求
type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" validate:"required"`
	NewPassword string `json:"new_password" validate:"required,min=8,max=100"`
}

// BlacklistTokenRequest Token黑名单请求
type BlacklistTokenRequest struct {
	Token  string `json:"token" validate:"required"`
	Reason string `json:"reason,omitempty"` // 可选的黑名单原因
}

// ============================================================================
// 认证相关响应
// ============================================================================

// AuthResponse 认证响应
type AuthResponse struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
	ExpiresIn    int64     `json:"expires_in"` // 访问 token 剩余秒数，便于前端调度刷新
	TokenType    string    `json:"token_type"`
	User         UserInfo  `json:"user"`
}

// TokenResponse Token响应
type TokenResponse struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
	TokenType    string    `json:"token_type"`
}

// ============================================================================
// 用户信息DTO
// ============================================================================

// UserInfo 用户信息
type UserInfo struct {
	ID       string `json:"id"`
	UUID     string `json:"uuid"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Phone    string `json:"phone,omitempty"`
	Nickname string `json:"nickname,omitempty"`
	Avatar   string `json:"avatar,omitempty"`
	Status   int    `json:"status"`
	Role     int    `json:"role"`
}

// UpdateProfileRequest 更新用户资料请求
type UpdateProfileRequest struct {
	Nickname string `json:"nickname,omitempty" validate:"omitempty,min=1,max=50"`
	Phone    string `json:"phone,omitempty" validate:"omitempty,e164"`
	Avatar   string `json:"avatar,omitempty" validate:"omitempty,url"`
}

// ============================================================================
// 会话信息DTO
// ============================================================================

// SessionInfo 会话信息
type SessionInfo struct {
	ID         string    `json:"id"`
	DeviceInfo string    `json:"device_info"`
	IPAddress  string    `json:"ip_address"`
	UserAgent  string    `json:"user_agent"`
	IsActive   bool      `json:"is_active"`
	CreatedAt  time.Time `json:"created_at"`
	ExpiresAt  time.Time `json:"expires_at"`
	IsCurrent  bool      `json:"is_current"`
}

// SessionListResponse 会话列表响应
type SessionListResponse struct {
	Sessions []SessionInfo `json:"sessions"`
	Total    int           `json:"total"`
}

// BlacklistStatsResponse Token黑名单统计响应
type BlacklistStatsResponse struct {
	TotalBlacklistedTokens int64     `json:"total_blacklisted_tokens"`
	LastUpdated            time.Time `json:"last_updated"`
}
