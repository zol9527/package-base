package interfaces

import (
	"context"

	"package-base/internal/models"
	"package-base/internal/options"
)

// Function: TotpSecretRepository
//
// Description: TOTP 密钥的仓储接口，支持启用/禁用与多种查询方法，服务于二步验证场景。
//
// Parameters:
//
//	无
//
// Returns:
//
//	无
//
// TotpSecretRepository TOTP密钥存储接口
type TotpSecretRepository interface {
	// 基础CRUD操作
	Create(ctx context.Context, totpSecret *models.TotpSecret) error
	GetByID(ctx context.Context, id string) (*models.TotpSecret, error)
	GetByUserID(ctx context.Context, userID string) (*models.TotpSecret, error)
	Update(ctx context.Context, totpSecret *models.TotpSecret) error
	Delete(ctx context.Context, id string) error

	// 查询操作
	Find(ctx context.Context, opts ...options.TotpSecretQueryOption) ([]*models.TotpSecret, error)
	Count(ctx context.Context, opts ...options.TotpSecretQueryOption) (int64, error)
	Exists(ctx context.Context, opts ...options.TotpSecretQueryOption) (bool, error)

	// 业务操作
	EnableForUser(ctx context.Context, userID string) error
	DisableForUser(ctx context.Context, userID string) error
	IsEnabledForUser(ctx context.Context, userID string) (bool, error)
}

// Function: BackupCodeRepository
//
// Description: 备用验证码的仓储接口，支持批量生成、校验与统计，用于账户恢复场景。
//
// Parameters:
//
//	无
//
// Returns:
//
//	无
//
// BackupCodeRepository 备用码存储接口
type BackupCodeRepository interface {
	// 基础CRUD操作
	Create(ctx context.Context, backupCode *models.BackupCode) error
	CreateBatch(ctx context.Context, backupCodes []*models.BackupCode) error
	GetByID(ctx context.Context, id string) (*models.BackupCode, error)
	GetByUserID(ctx context.Context, userID string) ([]*models.BackupCode, error)
	Update(ctx context.Context, backupCode *models.BackupCode) error
	Delete(ctx context.Context, id string) error
	DeleteByUserID(ctx context.Context, userID string) error

	// 查询操作
	Find(ctx context.Context, opts ...options.BackupCodeQueryOption) ([]*models.BackupCode, error)
	Count(ctx context.Context, opts ...options.BackupCodeQueryOption) (int64, error)
	Exists(ctx context.Context, opts ...options.BackupCodeQueryOption) (bool, error)

	// 业务操作
	ValidateAndUse(ctx context.Context, userID, code string) (*models.BackupCode, error)
	GetUnusedByUserID(ctx context.Context, userID string) ([]*models.BackupCode, error)
	GetUsedByUserID(ctx context.Context, userID string) ([]*models.BackupCode, error)
	CountUnusedByUserID(ctx context.Context, userID string) (int64, error)
	RegenerateForUser(ctx context.Context, userID string, codes []string) error
}
