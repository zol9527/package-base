package dto

// TOTP 2FA相关DTO

// EnableTOTPRequest 启用TOTP请求
type EnableTOTPRequest struct {
	Secret string `json:"secret" validate:"required"`
	Token  string `json:"token" validate:"required,len=6"`
}

// EnableTOTPResponse 启用TOTP响应
type EnableTOTPResponse struct {
	Success     bool     `json:"success"`
	BackupCodes []string `json:"backup_codes"`
	Message     string   `json:"message"`
}

// DisableTOTPRequest 禁用TOTP请求
type DisableTOTPRequest struct {
	Token string `json:"token" validate:"required"`
}

// VerifyTOTPRequest 验证TOTP请求
type VerifyTOTPRequest struct {
	Token string `json:"token" validate:"required"`
}

// VerifyBackupCodeRequest 验证备用码请求
type VerifyBackupCodeRequest struct {
	Code string `json:"code" validate:"required"`
}

// TotpSetupResponse TOTP设置响应
type TotpSetupResponse struct {
	Secret      string `json:"secret"`
	QRCodeB64   string `json:"qr_code_base64"`
	URL         string `json:"url"`
	Issuer      string `json:"issuer"`
	AccountName string `json:"account_name"`
}

// TotpStatusResponse TOTP状态响应
type TotpStatusResponse struct {
	IsEnabled bool   `json:"is_enabled"`
	Issuer    string `json:"issuer,omitempty"`
	CreatedAt string `json:"created_at,omitempty"`
}

// BackupCodeStatusResponse 备用码状态响应
type BackupCodeStatusResponse struct {
	TotalCount  int `json:"total_count"`
	UsedCount   int `json:"used_count"`
	UnusedCount int `json:"unused_count"`
}

// RegenerateBackupCodesResponse 重新生成备用码响应
type RegenerateBackupCodesResponse struct {
	BackupCodes []string `json:"backup_codes"`
	Count       int      `json:"count"`
	Message     string   `json:"message"`
}

// VerifyResponse 验证响应
type VerifyResponse struct {
	Valid   bool   `json:"valid"`
	Message string `json:"message"`
}
