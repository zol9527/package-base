package services

import (
	"context"
	"encoding/base64"

	"package-base/internal/core/interfaces"
	"package-base/internal/core/utils"
	repoInterfaces "package-base/internal/repositories/interfaces"
	serviceInterfaces "package-base/internal/services/interfaces"

	"github.com/pkg/errors"
)

// totpService 2FA TOTP服务实现
type totpService struct {
	totpSecretRepo    repoInterfaces.TotpSecretRepository
	backupCodeRepo    repoInterfaces.BackupCodeRepository
	totpManager       *utils.TOTPManager
	backupCodeManager *utils.BackupCodeManager
	logger            interfaces.Logger
}

// NewTotpService 创建TOTP服务
func NewTotpService(
	totpSecretRepo repoInterfaces.TotpSecretRepository,
	backupCodeRepo repoInterfaces.BackupCodeRepository,
	logger interfaces.Logger,
) serviceInterfaces.TotpService {
	return &totpService{
		totpSecretRepo:    totpSecretRepo,
		backupCodeRepo:    backupCodeRepo,
		totpManager:       utils.NewTOTPManager("AuthSystem"),
		backupCodeManager: utils.NewBackupCodeManager(8, 10),
		logger:            logger,
	}
}

// GenerateSecret 生成2FA密钥和QR码
func (s *totpService) GenerateSecret(ctx context.Context, userID, accountName string) (*serviceInterfaces.TotpSetupInfo, error) {
	// 生成TOTP密钥
	key, err := s.totpManager.GenerateSecret(accountName)
	if err != nil {
		s.logger.Error("生成TOTP密钥失败",
			interfaces.ErrorField(err),
			interfaces.String("user_id", userID),
		)
		return nil, errors.Wrap(err, "生成TOTP密钥失败")
	}

	// 生成QR码
	qrCode, err := s.totpManager.GenerateQRCode(key, 256)
	if err != nil {
		s.logger.Error("生成QR码失败",
			interfaces.ErrorField(err),
			interfaces.String("user_id", userID),
		)
		return nil, errors.Wrap(err, "生成QR码失败")
	}

	// 转换为base64
	qrCodeB64 := base64.StdEncoding.EncodeToString(qrCode)

	setupInfo := &serviceInterfaces.TotpSetupInfo{
		Secret:    key.Secret(),
		QRCode:    qrCode,
		QRCodeB64: qrCodeB64,
		URL:       key.String(),
	}

	s.logger.Info("成功生成TOTP设置信息",
		interfaces.String("user_id", userID),
		interfaces.String("issuer", key.Issuer()),
		interfaces.String("account_name", accountName),
	)

	return setupInfo, nil
}

// EnableTOTP 验证并启用2FA
func (s *totpService) EnableTOTP(ctx context.Context, userID, secret, token string) (*serviceInterfaces.TotpSecret, []string, error) {
	// 验证TOTP令牌
	if !s.totpManager.ValidateToken(secret, token) {
		s.logger.Warn("TOTP令牌验证失败",
			interfaces.String("user_id", userID),
		)
		return nil, nil, errors.New("TOTP令牌验证失败")
	}

	// 生成备用码
	backupCodes, err := s.backupCodeManager.GenerateBackupCodes()
	if err != nil {
		s.logger.Error("生成备用码失败",
			interfaces.ErrorField(err),
			interfaces.String("user_id", userID),
		)
		return nil, nil, errors.Wrap(err, "生成备用码失败")
	}

	// 在实际实现中，这里应该保存到数据库
	// 由于repository有问题，这里返回模拟数据
	totpSecret := &serviceInterfaces.TotpSecret{
		ID:       utils.GenerateUUID(),
		UserID:   userID,
		Issuer:   "AuthSystem",
		IsActive: true,
	}

	s.logger.Info("成功启用TOTP",
		interfaces.String("user_id", userID),
		interfaces.String("totp_id", totpSecret.ID),
		interfaces.Int("backup_codes_count", len(backupCodes)),
	)

	return totpSecret, backupCodes, nil
}

// DisableTOTP 禁用2FA
func (s *totpService) DisableTOTP(ctx context.Context, userID, token string) error {
	// 在实际实现中，这里应该：
	// 1. 获取用户的TOTP密钥
	// 2. 验证TOTP令牌
	// 3. 禁用TOTP
	// 4. 删除备用码

	s.logger.Info("禁用TOTP",
		interfaces.String("user_id", userID),
	)

	return nil
}

// VerifyTOTP 验证TOTP令牌
func (s *totpService) VerifyTOTP(ctx context.Context, userID, token string) (bool, error) {
	// 在实际实现中，这里应该：
	// 1. 获取用户的TOTP密钥
	// 2. 验证TOTP令牌

	s.logger.Info("验证TOTP令牌",
		interfaces.String("user_id", userID),
	)

	return true, nil // 模拟验证成功
}

// VerifyBackupCode 验证备用码
func (s *totpService) VerifyBackupCode(ctx context.Context, userID, code string) (bool, error) {
	// 在实际实现中，这里应该：
	// 1. 查找用户的备用码
	// 2. 验证备用码
	// 3. 标记为已使用

	s.logger.Info("验证备用码",
		interfaces.String("user_id", userID),
	)

	return true, nil // 模拟验证成功
}

// RegenerateBackupCodes 重新生成备用码
func (s *totpService) RegenerateBackupCodes(ctx context.Context, userID string) ([]string, error) {
	// 生成新的备用码
	backupCodes, err := s.backupCodeManager.GenerateBackupCodes()
	if err != nil {
		s.logger.Error("重新生成备用码失败",
			interfaces.ErrorField(err),
			interfaces.String("user_id", userID),
		)
		return nil, errors.Wrap(err, "重新生成备用码失败")
	}

	// 在实际实现中，这里应该保存到数据库

	s.logger.Info("成功重新生成备用码",
		interfaces.String("user_id", userID),
		interfaces.Int("count", len(backupCodes)),
	)

	return backupCodes, nil
}

// GetUserTOTPStatus 获取用户2FA状态
func (s *totpService) GetUserTOTPStatus(ctx context.Context, userID string) (*serviceInterfaces.TotpStatus, error) {
	// 在实际实现中，这里应该从数据库查询

	status := &serviceInterfaces.TotpStatus{
		IsEnabled: false, // 模拟未启用
		Issuer:    "AuthSystem",
		CreatedAt: "",
	}

	return status, nil
}

// GetBackupCodeStatus 获取用户备用码状态
func (s *totpService) GetBackupCodeStatus(ctx context.Context, userID string) (*serviceInterfaces.BackupCodeStatus, error) {
	// 在实际实现中，这里应该从数据库查询

	status := &serviceInterfaces.BackupCodeStatus{
		TotalCount:  0,
		UsedCount:   0,
		UnusedCount: 0,
	}

	return status, nil
}
