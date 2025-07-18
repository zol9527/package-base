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

// backupCodeRepository 备用码存储实现
type backupCodeRepository struct {
	storage interfaces.RelationalStorage
	logger  interfaces.Logger
}

// NewBackupCodeRepository 创建备用码存储实例
func NewBackupCodeRepository(storage interfaces.RelationalStorage, logger interfaces.Logger) repoInterfaces.BackupCodeRepository {
	return &backupCodeRepository{
		storage: storage,
		logger:  logger,
	}
}

// getDB 获取数据库连接
func (r *backupCodeRepository) getDB() *gorm.DB {
	return r.storage.GetDB()
}

// Create 创建备用码
func (r *backupCodeRepository) Create(ctx context.Context, backupCode *models.BackupCode) error {
	db := r.getDB()
	if db == nil {
		return errors.New("数据库连接无效")
	}

	if err := db.WithContext(ctx).Create(backupCode).Error; err != nil {
		r.logger.Error("创建备用码失败",
			interfaces.ErrorField(err),
			interfaces.String("user_id", backupCode.UserID),
		)
		return errors.Wrap(err, "创建备用码失败")
	}

	r.logger.Info("成功创建备用码",
		interfaces.String("id", backupCode.ID),
		interfaces.String("user_id", backupCode.UserID),
	)

	return nil
}

// CreateBatch 批量创建备用码
func (r *backupCodeRepository) CreateBatch(ctx context.Context, backupCodes []*models.BackupCode) error {
	db := r.getDB()
	if db == nil {
		return errors.New("数据库连接无效")
	}

	if err := db.WithContext(ctx).Create(&backupCodes).Error; err != nil {
		r.logger.Error("批量创建备用码失败",
			interfaces.ErrorField(err),
			interfaces.Int("count", len(backupCodes)),
		)
		return errors.Wrap(err, "批量创建备用码失败")
	}

	r.logger.Info("成功批量创建备用码",
		interfaces.Int("count", len(backupCodes)),
	)

	return nil
}

// GetByID 根据ID获取备用码
func (r *backupCodeRepository) GetByID(ctx context.Context, id string) (*models.BackupCode, error) {
	db := r.getDB()
	if db == nil {
		return nil, errors.New("数据库连接无效")
	}

	var backupCode models.BackupCode
	if err := db.WithContext(ctx).Where("id = ?", id).First(&backupCode).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		r.logger.Error("根据ID获取备用码失败",
			interfaces.ErrorField(err),
			interfaces.String("id", id),
		)
		return nil, errors.Wrap(err, "根据ID获取备用码失败")
	}

	return &backupCode, nil
}

// GetByUserID 根据用户ID获取备用码列表
func (r *backupCodeRepository) GetByUserID(ctx context.Context, userID string) ([]*models.BackupCode, error) {
	db := r.getDB()
	if db == nil {
		return nil, errors.New("数据库连接无效")
	}

	var backupCodes []*models.BackupCode
	if err := db.WithContext(ctx).Where("user_id = ?", userID).Find(&backupCodes).Error; err != nil {
		r.logger.Error("根据用户ID获取备用码失败",
			interfaces.ErrorField(err),
			interfaces.String("user_id", userID),
		)
		return nil, errors.Wrap(err, "根据用户ID获取备用码失败")
	}

	return backupCodes, nil
}

// Update 更新备用码
func (r *backupCodeRepository) Update(ctx context.Context, backupCode *models.BackupCode) error {
	db := r.getDB()
	if db == nil {
		return errors.New("数据库连接无效")
	}

	if err := db.WithContext(ctx).Save(backupCode).Error; err != nil {
		r.logger.Error("更新备用码失败",
			interfaces.ErrorField(err),
			interfaces.String("id", backupCode.ID),
		)
		return errors.Wrap(err, "更新备用码失败")
	}

	r.logger.Info("成功更新备用码",
		interfaces.String("id", backupCode.ID),
		interfaces.String("user_id", backupCode.UserID),
		interfaces.Bool("is_used", backupCode.IsUsed),
	)

	return nil
}

// Delete 删除备用码
func (r *backupCodeRepository) Delete(ctx context.Context, id string) error {
	db := r.getDB()
	if db == nil {
		return errors.New("数据库连接无效")
	}

	if err := db.WithContext(ctx).Delete(&models.BackupCode{}, "id = ?", id).Error; err != nil {
		r.logger.Error("删除备用码失败",
			interfaces.ErrorField(err),
			interfaces.String("id", id),
		)
		return errors.Wrap(err, "删除备用码失败")
	}

	r.logger.Info("成功删除备用码",
		interfaces.String("id", id),
	)

	return nil
}

// DeleteByUserID 删除用户的所有备用码
func (r *backupCodeRepository) DeleteByUserID(ctx context.Context, userID string) error {
	db := r.getDB()
	if db == nil {
		return errors.New("数据库连接无效")
	}

	if err := db.WithContext(ctx).Delete(&models.BackupCode{}, "user_id = ?", userID).Error; err != nil {
		r.logger.Error("删除用户备用码失败",
			interfaces.ErrorField(err),
			interfaces.String("user_id", userID),
		)
		return errors.Wrap(err, "删除用户备用码失败")
	}

	r.logger.Info("成功删除用户备用码",
		interfaces.String("user_id", userID),
	)

	return nil
}

// Find 查找备用码
func (r *backupCodeRepository) Find(ctx context.Context, opts ...options.BackupCodeQueryOption) ([]*models.BackupCode, error) {
	db := r.getDB()
	if db == nil {
		return nil, errors.New("数据库连接无效")
	}

	query := db.WithContext(ctx).Model(&models.BackupCode{})
	for _, opt := range opts {
		query = opt(query)
	}

	var backupCodes []*models.BackupCode
	if err := query.Find(&backupCodes).Error; err != nil {
		r.logger.Error("查找备用码失败",
			interfaces.ErrorField(err),
		)
		return nil, errors.Wrap(err, "查找备用码失败")
	}

	return backupCodes, nil
}

// Count 统计备用码数量
func (r *backupCodeRepository) Count(ctx context.Context, opts ...options.BackupCodeQueryOption) (int64, error) {
	db := r.getDB()
	if db == nil {
		return 0, errors.New("数据库连接无效")
	}

	query := db.WithContext(ctx).Model(&models.BackupCode{})
	for _, opt := range opts {
		query = opt(query)
	}

	var count int64
	if err := query.Count(&count).Error; err != nil {
		r.logger.Error("统计备用码数量失败",
			interfaces.ErrorField(err),
		)
		return 0, errors.Wrap(err, "统计备用码数量失败")
	}

	return count, nil
}

// Exists 检查备用码是否存在
func (r *backupCodeRepository) Exists(ctx context.Context, opts ...options.BackupCodeQueryOption) (bool, error) {
	count, err := r.Count(ctx, opts...)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// ValidateAndUse 验证并使用备用码
func (r *backupCodeRepository) ValidateAndUse(ctx context.Context, userID, code string) (*models.BackupCode, error) {
	db := r.getDB()
	if db == nil {
		return nil, errors.New("数据库连接无效")
	}

	// 开始事务
	tx := db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return nil, errors.Wrap(tx.Error, "开始事务失败")
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 查找未使用的备用码
	var backupCode models.BackupCode
	if err := tx.Where("user_id = ? AND code = ? AND is_used = ?", userID, code, false).First(&backupCode).Error; err != nil {
		tx.Rollback()
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // 备用码不存在或已使用
		}
		r.logger.Error("查找备用码失败",
			interfaces.ErrorField(err),
			interfaces.String("user_id", userID),
		)
		return nil, errors.Wrap(err, "查找备用码失败")
	}

	// 标记为已使用
	backupCode.MarkAsUsed()
	if err := tx.Save(&backupCode).Error; err != nil {
		tx.Rollback()
		r.logger.Error("标记备用码为已使用失败",
			interfaces.ErrorField(err),
			interfaces.String("id", backupCode.ID),
		)
		return nil, errors.Wrap(err, "标记备用码为已使用失败")
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		r.logger.Error("提交事务失败",
			interfaces.ErrorField(err),
		)
		return nil, errors.Wrap(err, "提交事务失败")
	}

	r.logger.Info("成功使用备用码",
		interfaces.String("id", backupCode.ID),
		interfaces.String("user_id", userID),
	)

	return &backupCode, nil
}

// GetUnusedByUserID 获取用户未使用的备用码
func (r *backupCodeRepository) GetUnusedByUserID(ctx context.Context, userID string) ([]*models.BackupCode, error) {
	return r.Find(ctx, options.WithUnusedBackupCodesByUserID(userID))
}

// GetUsedByUserID 获取用户已使用的备用码
func (r *backupCodeRepository) GetUsedByUserID(ctx context.Context, userID string) ([]*models.BackupCode, error) {
	return r.Find(ctx, options.WithUsedBackupCodesByUserID(userID))
}

// CountUnusedByUserID 统计用户未使用的备用码数量
func (r *backupCodeRepository) CountUnusedByUserID(ctx context.Context, userID string) (int64, error) {
	return r.Count(ctx, options.WithUnusedBackupCodesByUserID(userID))
}

// RegenerateForUser 为用户重新生成备用码
func (r *backupCodeRepository) RegenerateForUser(ctx context.Context, userID string, codes []string) error {
	db := r.getDB()
	if db == nil {
		return errors.New("数据库连接无效")
	}

	// 开始事务
	tx := db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return errors.Wrap(tx.Error, "开始事务失败")
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 删除用户现有的未使用备用码
	if err := tx.Delete(&models.BackupCode{}, "user_id = ? AND is_used = ?", userID, false).Error; err != nil {
		tx.Rollback()
		r.logger.Error("删除旧备用码失败",
			interfaces.ErrorField(err),
			interfaces.String("user_id", userID),
		)
		return errors.Wrap(err, "删除旧备用码失败")
	}

	// 创建新备用码
	backupCodes := make([]*models.BackupCode, len(codes))
	for i, code := range codes {
		backupCodes[i] = &models.BackupCode{
			UserID: userID,
			Code:   code,
			IsUsed: false,
		}
	}

	if err := tx.Create(&backupCodes).Error; err != nil {
		tx.Rollback()
		r.logger.Error("创建新备用码失败",
			interfaces.ErrorField(err),
			interfaces.String("user_id", userID),
		)
		return errors.Wrap(err, "创建新备用码失败")
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		r.logger.Error("提交事务失败",
			interfaces.ErrorField(err),
		)
		return errors.Wrap(err, "提交事务失败")
	}

	r.logger.Info("成功重新生成备用码",
		interfaces.String("user_id", userID),
		interfaces.Int("count", len(codes)),
	)

	return nil
}
