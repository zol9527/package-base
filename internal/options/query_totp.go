package options

import (
	"time"

	"gorm.io/gorm"
)

// TotpSecretQueryOption TOTP密钥查询选项函数类型
type TotpSecretQueryOption func(*gorm.DB) *gorm.DB

// 基础查询选项
func WithTotpSecretID(id string) TotpSecretQueryOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("id = ?", id)
	}
}

func WithTotpSecretUserID(userID string) TotpSecretQueryOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("user_id = ?", userID)
	}
}

func WithTotpSecretActive(isActive bool) TotpSecretQueryOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("is_active = ?", isActive)
	}
}

func WithTotpSecretIssuer(issuer string) TotpSecretQueryOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("issuer = ?", issuer)
	}
}

// 时间范围查询选项
func WithTotpSecretCreatedAfter(after time.Time) TotpSecretQueryOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("created_at > ?", after)
	}
}

func WithTotpSecretCreatedBefore(before time.Time) TotpSecretQueryOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("created_at < ?", before)
	}
}

func WithTotpSecretCreatedBetween(start, end time.Time) TotpSecretQueryOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("created_at BETWEEN ? AND ?", start, end)
	}
}

func WithTotpSecretUpdatedAfter(after time.Time) TotpSecretQueryOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("updated_at > ?", after)
	}
}

func WithTotpSecretUpdatedBefore(before time.Time) TotpSecretQueryOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("updated_at < ?", before)
	}
}

func WithTotpSecretUpdatedBetween(start, end time.Time) TotpSecretQueryOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("updated_at BETWEEN ? AND ?", start, end)
	}
}

// 排序选项
func WithTotpSecretOrderByCreatedAt(desc bool) TotpSecretQueryOption {
	return func(db *gorm.DB) *gorm.DB {
		if desc {
			return db.Order("created_at DESC")
		}
		return db.Order("created_at ASC")
	}
}

func WithTotpSecretOrderByUpdatedAt(desc bool) TotpSecretQueryOption {
	return func(db *gorm.DB) *gorm.DB {
		if desc {
			return db.Order("updated_at DESC")
		}
		return db.Order("updated_at ASC")
	}
}

func WithTotpSecretOrderByIssuer(desc bool) TotpSecretQueryOption {
	return func(db *gorm.DB) *gorm.DB {
		if desc {
			return db.Order("issuer DESC")
		}
		return db.Order("issuer ASC")
	}
}

// 分页选项
func WithTotpSecretLimit(limit int) TotpSecretQueryOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Limit(limit)
	}
}

func WithTotpSecretOffset(offset int) TotpSecretQueryOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Offset(offset)
	}
}

func WithTotpSecretPagination(page, pageSize int) TotpSecretQueryOption {
	return func(db *gorm.DB) *gorm.DB {
		if page < 1 {
			page = 1
		}
		if pageSize < 1 {
			pageSize = 10
		}
		offset := (page - 1) * pageSize
		return db.Offset(offset).Limit(pageSize)
	}
}

// 预加载选项
func WithTotpSecretPreloadUser() TotpSecretQueryOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Preload("User")
	}
}

// 复合查询选项
func WithActiveTotpSecretByUserID(userID string) TotpSecretQueryOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("user_id = ? AND is_active = ?", userID, true)
	}
}

func WithTotpSecretsByIssuerAndActive(issuer string, isActive bool) TotpSecretQueryOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("issuer = ? AND is_active = ?", issuer, isActive)
	}
}

func WithRecentTotpSecrets(days int) TotpSecretQueryOption {
	return func(db *gorm.DB) *gorm.DB {
		since := time.Now().AddDate(0, 0, -days)
		return db.Where("created_at > ?", since)
	}
}

// BackupCode查询选项
type BackupCodeQueryOption func(*gorm.DB) *gorm.DB

// 基础查询选项
func WithBackupCodeID(id string) BackupCodeQueryOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("id = ?", id)
	}
}

func WithBackupCodeUserID(userID string) BackupCodeQueryOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("user_id = ?", userID)
	}
}

func WithBackupCodeIsUsed(isUsed bool) BackupCodeQueryOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("is_used = ?", isUsed)
	}
}

func WithBackupCodeCode(code string) BackupCodeQueryOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("code = ?", code)
	}
}

// 时间范围查询选项
func WithBackupCodeCreatedAfter(after time.Time) BackupCodeQueryOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("created_at > ?", after)
	}
}

func WithBackupCodeCreatedBefore(before time.Time) BackupCodeQueryOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("created_at < ?", before)
	}
}

func WithBackupCodeCreatedBetween(start, end time.Time) BackupCodeQueryOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("created_at BETWEEN ? AND ?", start, end)
	}
}

func WithBackupCodeUsedAfter(after time.Time) BackupCodeQueryOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("used_at > ?", after)
	}
}

func WithBackupCodeUsedBefore(before time.Time) BackupCodeQueryOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("used_at < ?", before)
	}
}

func WithBackupCodeUsedBetween(start, end time.Time) BackupCodeQueryOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("used_at BETWEEN ? AND ?", start, end)
	}
}

// 排序选项
func WithBackupCodeOrderByCreatedAt(desc bool) BackupCodeQueryOption {
	return func(db *gorm.DB) *gorm.DB {
		if desc {
			return db.Order("created_at DESC")
		}
		return db.Order("created_at ASC")
	}
}

func WithBackupCodeOrderByUsedAt(desc bool) BackupCodeQueryOption {
	return func(db *gorm.DB) *gorm.DB {
		if desc {
			return db.Order("used_at DESC")
		}
		return db.Order("used_at ASC")
	}
}

// 分页选项
func WithBackupCodeLimit(limit int) BackupCodeQueryOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Limit(limit)
	}
}

func WithBackupCodeOffset(offset int) BackupCodeQueryOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Offset(offset)
	}
}

func WithBackupCodePagination(page, pageSize int) BackupCodeQueryOption {
	return func(db *gorm.DB) *gorm.DB {
		if page < 1 {
			page = 1
		}
		if pageSize < 1 {
			pageSize = 10
		}
		offset := (page - 1) * pageSize
		return db.Offset(offset).Limit(pageSize)
	}
}

// 预加载选项
func WithBackupCodePreloadUser() BackupCodeQueryOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Preload("User")
	}
}

// 复合查询选项
func WithUnusedBackupCodesByUserID(userID string) BackupCodeQueryOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("user_id = ? AND is_used = ?", userID, false)
	}
}

func WithUsedBackupCodesByUserID(userID string) BackupCodeQueryOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("user_id = ? AND is_used = ?", userID, true)
	}
}

func WithRecentUsedBackupCodes(days int) BackupCodeQueryOption {
	return func(db *gorm.DB) *gorm.DB {
		since := time.Now().AddDate(0, 0, -days)
		return db.Where("is_used = ? AND used_at > ?", true, since)
	}
}
