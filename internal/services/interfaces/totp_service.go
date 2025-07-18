package interfaces

import (
	"context"
)

// TotpService 2FA TOTP服务接口
type TotpService interface {
	// 生成2FA密钥和QR码
	GenerateSecret(ctx context.Context, userID, accountName string) (*TotpSetupInfo, error)

	// 验证并启用2FA
	EnableTOTP(ctx context.Context, userID, secret, token string) (*TotpSecret, []string, error)

	// 禁用2FA
	DisableTOTP(ctx context.Context, userID, token string) error

	// 验证TOTP令牌
	VerifyTOTP(ctx context.Context, userID, token string) (bool, error)

	// 验证备用码
	VerifyBackupCode(ctx context.Context, userID, code string) (bool, error)

	// 重新生成备用码
	RegenerateBackupCodes(ctx context.Context, userID string) ([]string, error)

	// 获取用户2FA状态
	GetUserTOTPStatus(ctx context.Context, userID string) (*TotpStatus, error)

	// 获取用户备用码状态
	GetBackupCodeStatus(ctx context.Context, userID string) (*BackupCodeStatus, error)
}

// TotpSetupInfo TOTP设置信息
type TotpSetupInfo struct {
	Secret    string `json:"secret"`
	QRCode    []byte `json:"qr_code"`
	QRCodeB64 string `json:"qr_code_base64"`
	URL       string `json:"url"`
}

// TotpSecret TOTP密钥信息
type TotpSecret struct {
	ID       string `json:"id"`
	UserID   string `json:"user_id"`
	Issuer   string `json:"issuer"`
	IsActive bool   `json:"is_active"`
}

// TotpStatus TOTP状态
type TotpStatus struct {
	IsEnabled bool   `json:"is_enabled"`
	Issuer    string `json:"issuer,omitempty"`
	CreatedAt string `json:"created_at,omitempty"`
}

// BackupCodeStatus 备用码状态
type BackupCodeStatus struct {
	TotalCount  int `json:"total_count"`
	UsedCount   int `json:"used_count"`
	UnusedCount int `json:"unused_count"`
}
