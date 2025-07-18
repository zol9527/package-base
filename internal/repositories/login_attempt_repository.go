package repositories

import (
	"context"
	"time"

	"package-base/internal/core/interfaces"
	"package-base/internal/models"
	"package-base/internal/options"
	repoInterfaces "package-base/internal/repositories/interfaces"

	"github.com/pkg/errors"
	"gorm.io/gorm"
)

// loginAttemptRepository 登录尝试存储实现
type loginAttemptRepository struct {
	storage interfaces.RelationalStorage
	logger  interfaces.Logger
}

// NewLoginAttemptRepository 创建登录尝试存储实现
func NewLoginAttemptRepository(
	storage interfaces.RelationalStorage,
	logger interfaces.Logger,
) repoInterfaces.LoginAttemptRepository {
	return &loginAttemptRepository{
		storage: storage,
		logger:  logger.Named("login_attempt_repository"),
	}
}

// getDB 获取数据库连接
func (r *loginAttemptRepository) getDB() *gorm.DB {
	return r.storage.GetDB()
}

// Create 创建登录尝试记录
func (r *loginAttemptRepository) Create(ctx context.Context, attempt *models.LoginAttempt) error {
	db := r.getDB()

	if err := db.WithContext(ctx).Create(attempt).Error; err != nil {
		r.logger.Error("failed to create login attempt",
			interfaces.ErrorField(err),
			interfaces.String("ip_address", attempt.IPAddress),
			interfaces.String("login_identity", attempt.LoginIdentity),
			interfaces.String("attempt_type", attempt.AttemptType),
		)
		return errors.Wrap(err, "failed to create login attempt")
	}

	r.logger.Debug("login attempt created successfully",
		interfaces.Uint("attempt_id", attempt.ID),
		interfaces.String("ip_address", attempt.IPAddress),
		interfaces.String("attempt_type", attempt.AttemptType),
	)

	return nil
}

// GetByID 根据ID获取登录尝试记录
func (r *loginAttemptRepository) GetByID(ctx context.Context, id uint) (*models.LoginAttempt, error) {
	return r.FindLoginAttempt(ctx, options.WithLoginAttemptID(id))
}

// Update 更新登录尝试记录
func (r *loginAttemptRepository) Update(ctx context.Context, attempt *models.LoginAttempt) error {
	db := r.getDB()

	if err := db.WithContext(ctx).Save(attempt).Error; err != nil {
		r.logger.Error("failed to update login attempt",
			interfaces.ErrorField(err),
			interfaces.Uint("attempt_id", attempt.ID),
		)
		return errors.Wrap(err, "failed to update login attempt")
	}

	r.logger.Debug("login attempt updated successfully",
		interfaces.Uint("attempt_id", attempt.ID),
	)

	return nil
}

// Delete 删除登录尝试记录
func (r *loginAttemptRepository) Delete(ctx context.Context, id uint) error {
	db := r.getDB()

	queryOptions := options.NewQueryLoginAttemptOptions(db, options.WithLoginAttemptID(id))
	queryDB := queryOptions.GetDB()

	if err := queryDB.WithContext(ctx).Delete(&models.LoginAttempt{}).Error; err != nil {
		r.logger.Error("failed to delete login attempt",
			interfaces.ErrorField(err),
			interfaces.Uint("attempt_id", id),
		)
		return errors.Wrap(err, "failed to delete login attempt")
	}

	r.logger.Debug("login attempt deleted successfully",
		interfaces.Uint("attempt_id", id),
	)

	return nil
}

// FindLoginAttempts 根据选项查询登录尝试列表
func (r *loginAttemptRepository) FindLoginAttempts(ctx context.Context, opts ...options.QueryLoginAttemptOption) ([]*models.LoginAttempt, error) {
	db := r.getDB()

	queryOptions := options.NewQueryLoginAttemptOptions(db, opts...)
	queryDB := queryOptions.GetDB()

	var attempts []*models.LoginAttempt
	if err := queryDB.WithContext(ctx).Find(&attempts).Error; err != nil {
		r.logger.Error("failed to find login attempts",
			interfaces.ErrorField(err),
		)
		return nil, errors.Wrap(err, "failed to find login attempts")
	}

	return attempts, nil
}

// FindLoginAttempt 根据选项查询单个登录尝试
func (r *loginAttemptRepository) FindLoginAttempt(ctx context.Context, opts ...options.QueryLoginAttemptOption) (*models.LoginAttempt, error) {
	db := r.getDB()

	queryOptions := options.NewQueryLoginAttemptOptions(db, opts...)
	queryDB := queryOptions.GetDB()

	var attempt models.LoginAttempt
	if err := queryDB.WithContext(ctx).First(&attempt).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		r.logger.Error("failed to find login attempt",
			interfaces.ErrorField(err),
		)
		return nil, errors.Wrap(err, "failed to find login attempt")
	}

	return &attempt, nil
}

// CountLoginAttempts 根据选项统计登录尝试数量
func (r *loginAttemptRepository) CountLoginAttempts(ctx context.Context, opts ...options.QueryLoginAttemptOption) (int64, error) {
	db := r.getDB()

	queryOptions := options.NewQueryLoginAttemptOptions(db, opts...)
	queryDB := queryOptions.GetDB()

	var count int64
	if err := queryDB.WithContext(ctx).Model(&models.LoginAttempt{}).Count(&count).Error; err != nil {
		r.logger.Error("failed to count login attempts",
			interfaces.ErrorField(err),
		)
		return 0, errors.Wrap(err, "failed to count login attempts")
	}

	return count, nil
}

// FindLoginAttemptsWithPagination 根据选项分页查询登录尝试
func (r *loginAttemptRepository) FindLoginAttemptsWithPagination(ctx context.Context, page, pageSize int, opts ...options.QueryLoginAttemptOption) ([]*models.LoginAttempt, int64, error) {
	db := r.getDB()

	// 先统计总数
	countOptions := options.NewQueryLoginAttemptOptions(db, opts...)
	countDB := countOptions.GetDB()

	var total int64
	if err := countDB.WithContext(ctx).Model(&models.LoginAttempt{}).Count(&total).Error; err != nil {
		r.logger.Error("failed to count login attempts for pagination",
			interfaces.ErrorField(err),
		)
		return nil, 0, errors.Wrap(err, "failed to count login attempts")
	}

	// 添加分页选项
	opts = append(opts, options.WithLoginAttemptPagination(page, pageSize))

	// 查询数据
	attempts, err := r.FindLoginAttempts(ctx, opts...)
	if err != nil {
		return nil, 0, err
	}

	return attempts, total, nil
}

// GetRecentFailedAttemptsByIP 获取IP最近的失败尝试次数
func (r *loginAttemptRepository) GetRecentFailedAttemptsByIP(ctx context.Context, ipAddress string, minutes int) (int64, error) {
	return r.CountLoginAttempts(ctx,
		options.WithIPRecentFailedAttempts(ipAddress, minutes),
	)
}

// GetRecentFailedAttemptsByUser 获取用户最近的失败尝试次数
func (r *loginAttemptRepository) GetRecentFailedAttemptsByUser(ctx context.Context, userID string, minutes int) (int64, error) {
	return r.CountLoginAttempts(ctx,
		options.WithUserRecentFailedAttempts(userID, minutes),
	)
}

// GetRecentFailedAttemptsByIdentity 获取登录标识最近的失败尝试次数
func (r *loginAttemptRepository) GetRecentFailedAttemptsByIdentity(ctx context.Context, identity string, minutes int) (int64, error) {
	return r.CountLoginAttempts(ctx,
		options.WithIdentityRecentFailedAttempts(identity, minutes),
	)
}

// GetIPAttemptStats 获取IP的尝试统计
func (r *loginAttemptRepository) GetIPAttemptStats(ctx context.Context, ipAddress string) (map[string]int64, error) {
	db := r.getDB()

	var stats []struct {
		AttemptType string `gorm:"column:attempt_type"`
		Count       int64  `gorm:"column:count"`
	}

	if err := db.WithContext(ctx).
		Model(&models.LoginAttempt{}).
		Select("attempt_type, COUNT(*) as count").
		Where("ip_address = ?", ipAddress).
		Group("attempt_type").
		Find(&stats).Error; err != nil {
		r.logger.Error("failed to get IP attempt stats",
			interfaces.ErrorField(err),
			interfaces.String("ip_address", ipAddress),
		)
		return nil, errors.Wrap(err, "failed to get IP attempt stats")
	}

	result := make(map[string]int64)
	for _, stat := range stats {
		result[stat.AttemptType] = stat.Count
	}

	return result, nil
}

// GetUserAttemptStats 获取用户的尝试统计
func (r *loginAttemptRepository) GetUserAttemptStats(ctx context.Context, userID string) (map[string]int64, error) {
	db := r.getDB()

	var stats []struct {
		AttemptType string `gorm:"column:attempt_type"`
		Count       int64  `gorm:"column:count"`
	}

	if err := db.WithContext(ctx).
		Model(&models.LoginAttempt{}).
		Select("attempt_type, COUNT(*) as count").
		Where("user_id = ?", userID).
		Group("attempt_type").
		Find(&stats).Error; err != nil {
		r.logger.Error("failed to get user attempt stats",
			interfaces.ErrorField(err),
			interfaces.String("user_id", userID),
		)
		return nil, errors.Wrap(err, "failed to get user attempt stats")
	}

	result := make(map[string]int64)
	for _, stat := range stats {
		result[stat.AttemptType] = stat.Count
	}

	return result, nil
}

// CleanupOldAttempts 清理过期的登录尝试记录
func (r *loginAttemptRepository) CleanupOldAttempts(ctx context.Context, daysToKeep int) error {
	db := r.getDB()

	cutoffTime := time.Now().AddDate(0, 0, -daysToKeep)

	result := db.WithContext(ctx).
		Where("created_at < ?", cutoffTime).
		Delete(&models.LoginAttempt{})

	if result.Error != nil {
		r.logger.Error("failed to cleanup old login attempts",
			interfaces.ErrorField(result.Error),
		)
		return errors.Wrap(result.Error, "failed to cleanup old login attempts")
	}

	r.logger.Info("old login attempts cleaned up",
		interfaces.Int64("deleted_count", result.RowsAffected),
		interfaces.Int("days_to_keep", daysToKeep),
	)

	return nil
}

// GetTopFailedIPs 获取失败尝试最多的IP列表
func (r *loginAttemptRepository) GetTopFailedIPs(ctx context.Context, limit int, hours int) ([]map[string]interface{}, error) {
	db := r.getDB()

	since := time.Now().Add(-time.Duration(hours) * time.Hour)

	var results []struct {
		IPAddress string `gorm:"column:ip_address"`
		Count     int64  `gorm:"column:count"`
		Country   string `gorm:"column:country"`
		City      string `gorm:"column:city"`
	}

	if err := db.WithContext(ctx).
		Model(&models.LoginAttempt{}).
		Select("ip_address, COUNT(*) as count, MAX(country) as country, MAX(city) as city").
		Where("attempt_type = ? AND created_at >= ?", models.AttemptTypeFailed, since).
		Group("ip_address").
		Order("count DESC").
		Limit(limit).
		Find(&results).Error; err != nil {
		r.logger.Error("failed to get top failed IPs",
			interfaces.ErrorField(err),
		)
		return nil, errors.Wrap(err, "failed to get top failed IPs")
	}

	response := make([]map[string]interface{}, len(results))
	for i, result := range results {
		response[i] = map[string]interface{}{
			"ip_address":   result.IPAddress,
			"failed_count": result.Count,
			"country":      result.Country,
			"city":         result.City,
		}
	}

	return response, nil
}

// GetFailedAttemptsTrend 获取失败尝试趋势数据
func (r *loginAttemptRepository) GetFailedAttemptsTrend(ctx context.Context, hours int) ([]map[string]interface{}, error) {
	db := r.getDB()

	since := time.Now().Add(-time.Duration(hours) * time.Hour)

	var results []struct {
		Hour  string `gorm:"column:hour"`
		Count int64  `gorm:"column:count"`
	}

	if err := db.WithContext(ctx).
		Model(&models.LoginAttempt{}).
		Select("DATE_FORMAT(created_at, '%Y-%m-%d %H:00:00') as hour, COUNT(*) as count").
		Where("attempt_type = ? AND created_at >= ?", models.AttemptTypeFailed, since).
		Group("hour").
		Order("hour ASC").
		Find(&results).Error; err != nil {
		r.logger.Error("failed to get failed attempts trend",
			interfaces.ErrorField(err),
		)
		return nil, errors.Wrap(err, "failed to get failed attempts trend")
	}

	response := make([]map[string]interface{}, len(results))
	for i, result := range results {
		response[i] = map[string]interface{}{
			"hour":         result.Hour,
			"failed_count": result.Count,
		}
	}

	return response, nil
}
