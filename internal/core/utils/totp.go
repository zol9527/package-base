package utils

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base32"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
	"github.com/skip2/go-qrcode"
)

// TOTPManager TOTP管理器
type TOTPManager struct {
	issuer string
}

// NewTOTPManager 创建TOTP管理器
func NewTOTPManager(issuer string) *TOTPManager {
	if issuer == "" {
		issuer = "AuthSystem"
	}
	return &TOTPManager{
		issuer: issuer,
	}
}

// GenerateSecret 生成TOTP密钥
func (tm *TOTPManager) GenerateSecret(accountName string) (*otp.Key, error) {
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      tm.issuer,
		AccountName: accountName,
		SecretSize:  32, // 256位密钥
	})
	if err != nil {
		return nil, fmt.Errorf("生成TOTP密钥失败: %w", err)
	}
	return key, nil
}

// GenerateQRCode 生成QR码
func (tm *TOTPManager) GenerateQRCode(key *otp.Key, size int) ([]byte, error) {
	if size <= 0 {
		size = 256
	}

	qrCode, err := qrcode.Encode(key.String(), qrcode.Medium, size)
	if err != nil {
		return nil, fmt.Errorf("生成QR码失败: %w", err)
	}
	return qrCode, nil
}

// ValidateToken 验证TOTP令牌
func (tm *TOTPManager) ValidateToken(secret, token string) bool {
	// 允许前后30秒的时间窗口
	return totp.Validate(token, secret)
}

// ValidateTokenWithSkew 验证TOTP令牌（带时间偏移）
func (tm *TOTPManager) ValidateTokenWithSkew(secret, token string, skew uint) bool {
	opts := totp.ValidateOpts{
		Period:    30,
		Skew:      skew,
		Digits:    otp.DigitsSix,
		Algorithm: otp.AlgorithmSHA1,
	}
	valid, err := totp.ValidateCustom(token, secret, time.Now(), opts)
	return err == nil && valid
}

// GenerateCurrentToken 生成当前TOTP令牌（主要用于测试）
func (tm *TOTPManager) GenerateCurrentToken(secret string) (string, error) {
	token, err := totp.GenerateCode(secret, time.Now())
	if err != nil {
		return "", fmt.Errorf("生成TOTP令牌失败: %w", err)
	}
	return token, nil
}

// BackupCodeManager 备用码管理器
type BackupCodeManager struct {
	codeLength int
	codeCount  int
}

// NewBackupCodeManager 创建备用码管理器
func NewBackupCodeManager(codeLength, codeCount int) *BackupCodeManager {
	if codeLength <= 0 {
		codeLength = 8
	}
	if codeCount <= 0 {
		codeCount = 10
	}
	return &BackupCodeManager{
		codeLength: codeLength,
		codeCount:  codeCount,
	}
}

// GenerateBackupCodes 生成备用码
func (bcm *BackupCodeManager) GenerateBackupCodes() ([]string, error) {
	codes := make([]string, bcm.codeCount)

	for i := 0; i < bcm.codeCount; i++ {
		code, err := bcm.generateSingleCode()
		if err != nil {
			return nil, fmt.Errorf("生成备用码失败: %w", err)
		}
		codes[i] = code
	}

	return codes, nil
}

// generateSingleCode 生成单个备用码
func (bcm *BackupCodeManager) generateSingleCode() (string, error) {
	// 生成随机字节
	bytes := make([]byte, bcm.codeLength)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}

	// 转换为base32编码（去除填充符）
	code := base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(bytes)

	// 截取指定长度
	if len(code) > bcm.codeLength {
		code = code[:bcm.codeLength]
	}

	// 转换为大写并格式化（每4位加一个连字符）
	code = strings.ToUpper(code)
	return bcm.formatCode(code), nil
}

// formatCode 格式化备用码（添加连字符）
func (bcm *BackupCodeManager) formatCode(code string) string {
	if len(code) <= 4 {
		return code
	}

	var formatted strings.Builder
	for i, char := range code {
		if i > 0 && i%4 == 0 {
			formatted.WriteString("-")
		}
		formatted.WriteRune(char)
	}
	return formatted.String()
}

// ValidateBackupCode 验证备用码
func (bcm *BackupCodeManager) ValidateBackupCode(inputCode, storedCode string) bool {
	// 移除输入中的连字符和空格，转换为大写
	cleanInput := strings.ToUpper(strings.ReplaceAll(strings.ReplaceAll(inputCode, "-", ""), " ", ""))
	cleanStored := strings.ToUpper(strings.ReplaceAll(strings.ReplaceAll(storedCode, "-", ""), " ", ""))

	// 使用常量时间比较防止时序攻击
	return subtle.ConstantTimeCompare([]byte(cleanInput), []byte(cleanStored)) == 1
}

// TOTPSecretFromString 从字符串创建TOTP密钥
func (tm *TOTPManager) TOTPSecretFromString(secret string) (*otp.Key, error) {
	// 构造标准的otpauth URL
	u := &url.URL{
		Scheme: "otpauth",
		Host:   "totp",
		Path:   "/" + tm.issuer + ":user",
	}

	q := u.Query()
	q.Set("secret", secret)
	q.Set("issuer", tm.issuer)
	u.RawQuery = q.Encode()

	key, err := otp.NewKeyFromURL(u.String())
	if err != nil {
		return nil, fmt.Errorf("解析TOTP密钥失败: %w", err)
	}

	return key, nil
}

// ValidateSecretFormat 验证密钥格式
func (tm *TOTPManager) ValidateSecretFormat(secret string) bool {
	// 检查是否为有效的base32编码
	_, err := base32.StdEncoding.DecodeString(secret)
	return err == nil && len(secret) >= 16 // 至少128位
}
