package repositories

import (
	"context"

	"package-base/internal/core/interfaces"
	"package-base/internal/models"
	"package-base/internal/options"
	repoInterfaces "package-base/internal/repositories/interfaces"

	"github.com/pkg/errors"
	"gorm.io/gorm"
)

// totpSecretRepository TOTP密钥存储实现
type totpSecretRepository struct {
	storage interfaces.RelationalStorage
	logger  interfaces.Logger
}

// NewTotpSecretRepository 创建TOTP密钥存储实例
func NewTotpSecretRepository(storage interfaces.RelationalStorage, logger interfaces.Logger) repoInterfaces.TotpSecretRepository {
	return &totpSecretRepository{
		storage: storage,
		logger:  logger,
	}
}

// Create 创建TOTP密钥
func (r *totpSecretRepository) Create(ctx context.Context, totpSecret *models.TotpSecret) error {
	db := r.storage.GetDB()
	if db == nil {
		return errors.New("数据库连接无效")
	}

	if err := db.WithContext(ctx).Create(totpSecret).Error; err != nil {
		r.logger.Error("创建TOTP密钥失败",
			interfaces.ErrorField(err),
			interfaces.String("user_id", totpSecret.UserID),
		)
		return errors.Wrap(err, "创建TOTP密钥失败")
	}

	r.logger.Info("成功创建TOTP密钥",
		interfaces.String("id", totpSecret.ID),
		interfaces.String("user_id", totpSecret.UserID),
		interfaces.String("issuer", totpSecret.Issuer),
	)

	return nil
}

// GetByID 根据ID获取TOTP密钥
func (r *totpSecretRepository) GetByID(ctx context.Context, id string) (*models.TotpSecret, error) {
	db := r.storage.GetDB()
	if db == nil {
		return nil, errors.New("数据库连接无效")
	}

	var totpSecret models.TotpSecret
	if err := db.WithContext(ctx).Where("id = ?", id).First(&totpSecret).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		r.logger.Error("根据ID获取TOTP密钥失败",
			interfaces.ErrorField(err),
			interfaces.String("id", id),
		)
		return nil, errors.Wrap(err, "根据ID获取TOTP密钥失败")
	}

	return &totpSecret, nil
}

// GetByUserID 根据用户ID获取TOTP密钥
func (r *totpSecretRepository) GetByUserID(ctx context.Context, userID string) (*models.TotpSecret, error) {
	db := r.storage.GetDB()
	if db == nil {
		return nil, errors.New("数据库连接无效")
	}

	var totpSecret models.TotpSecret
	if err := db.WithContext(ctx).Where("user_id = ?", userID).First(&totpSecret).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		r.logger.Error("根据用户ID获取TOTP密钥失败",
			interfaces.ErrorField(err),
			interfaces.String("user_id", userID),
		)
		return nil, errors.Wrap(err, "根据用户ID获取TOTP密钥失败")
	}

	return &totpSecret, nil
}

// Update 更新TOTP密钥
func (r *totpSecretRepository) Update(ctx context.Context, totpSecret *models.TotpSecret) error {
	db := r.storage.GetDB()
	if db == nil {
		return errors.New("数据库连接无效")
	}

	if err := db.WithContext(ctx).Save(totpSecret).Error; err != nil {
		r.logger.Error("更新TOTP密钥失败",
			interfaces.ErrorField(err),
			interfaces.String("id", totpSecret.ID),
		)
		return errors.Wrap(err, "更新TOTP密钥失败")
	}

	r.logger.Info("成功更新TOTP密钥",
		interfaces.String("id", totpSecret.ID),
		interfaces.String("user_id", totpSecret.UserID),
		interfaces.Bool("is_active", totpSecret.IsActive),
	)

	return nil
}

// Delete 删除TOTP密钥
func (r *totpSecretRepository) Delete(ctx context.Context, id string) error {
	db := r.storage.GetDB()
	if db == nil {
		return errors.New("数据库连接无效")
	}

	if err := db.WithContext(ctx).Delete(&models.TotpSecret{}, "id = ?", id).Error; err != nil {
		r.logger.Error("删除TOTP密钥失败",
			interfaces.ErrorField(err),
			interfaces.String("id", id),
		)
		return errors.Wrap(err, "删除TOTP密钥失败")
	}

	r.logger.Info("成功删除TOTP密钥",
		interfaces.String("id", id),
	)

	return nil
}

// Find 查找TOTP密钥
func (r *totpSecretRepository) Find(ctx context.Context, opts ...options.TotpSecretQueryOption) ([]*models.TotpSecret, error) {
	db := r.storage.GetDB()
	if db == nil {
		return nil, errors.New("数据库连接无效")
	}

	query := db.WithContext(ctx).Model(&models.TotpSecret{})
	for _, opt := range opts {
		query = opt(query)
	}

	var totpSecrets []*models.TotpSecret
	if err := query.Find(&totpSecrets).Error; err != nil {
		r.logger.Error("查找TOTP密钥失败",
			interfaces.ErrorField(err),
		)
		return nil, errors.Wrap(err, "查找TOTP密钥失败")
	}

	return totpSecrets, nil
}

// Count 统计TOTP密钥数量
func (r *totpSecretRepository) Count(ctx context.Context, opts ...options.TotpSecretQueryOption) (int64, error) {
	db := r.storage.GetDB()
	if db == nil {
		return 0, errors.New("数据库连接无效")
	}

	query := db.WithContext(ctx).Model(&models.TotpSecret{})
	for _, opt := range opts {
		query = opt(query)
	}

	var count int64
	if err := query.Count(&count).Error; err != nil {
		r.logger.Error("统计TOTP密钥数量失败",
			interfaces.ErrorField(err),
		)
		return 0, errors.Wrap(err, "统计TOTP密钥数量失败")
	}

	return count, nil
}

// Exists 检查TOTP密钥是否存在
func (r *totpSecretRepository) Exists(ctx context.Context, opts ...options.TotpSecretQueryOption) (bool, error) {
	count, err := r.Count(ctx, opts...)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// EnableForUser 为用户启用TOTP
func (r *totpSecretRepository) EnableForUser(ctx context.Context, userID string) error {
	db := r.storage.GetDB()
	if db == nil {
		return errors.New("数据库连接无效")
	}

	if err := db.WithContext(ctx).Model(&models.TotpSecret{}).
		Where("user_id = ?", userID).
		Update("is_active", true).Error; err != nil {
		r.logger.Error("启用用户TOTP失败",
			interfaces.ErrorField(err),
			interfaces.String("user_id", userID),
		)
		return errors.Wrap(err, "启用用户TOTP失败")
	}

	r.logger.Info("成功启用用户TOTP",
		interfaces.String("user_id", userID),
	)

	return nil
}

// DisableForUser 为用户禁用TOTP
func (r *totpSecretRepository) DisableForUser(ctx context.Context, userID string) error {
	db := r.storage.GetDB()
	if db == nil {
		return errors.New("数据库连接无效")
	}

	if err := db.WithContext(ctx).Model(&models.TotpSecret{}).
		Where("user_id = ?", userID).
		Update("is_active", false).Error; err != nil {
		r.logger.Error("禁用用户TOTP失败",
			interfaces.ErrorField(err),
			interfaces.String("user_id", userID),
		)
		return errors.Wrap(err, "禁用用户TOTP失败")
	}

	r.logger.Info("成功禁用用户TOTP",
		interfaces.String("user_id", userID),
	)

	return nil
}

// IsEnabledForUser 检查用户是否启用了TOTP
func (r *totpSecretRepository) IsEnabledForUser(ctx context.Context, userID string) (bool, error) {
	return r.Exists(ctx, options.WithActiveTotpSecretByUserID(userID))
}
