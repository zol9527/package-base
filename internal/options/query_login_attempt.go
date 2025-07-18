package options

import (
	"time"

	"gorm.io/gorm"
)

// QueryLoginAttemptOption 登录尝试查询选项
type QueryLoginAttemptOption func(*gorm.DB) *gorm.DB

// LoginAttempt查询选项构建器
type QueryLoginAttemptOptions struct {
	db *gorm.DB
}

// NewQueryLoginAttemptOptions 创建登录尝试查询选项构建器
func NewQueryLoginAttemptOptions(db *gorm.DB, opts ...QueryLoginAttemptOption) *QueryLoginAttemptOptions {
	queryDB := db
	for _, opt := range opts {
		queryDB = opt(queryDB)
	}
	return &QueryLoginAttemptOptions{db: queryDB}
}

// GetDB 获取查询数据库连接
func (q *QueryLoginAttemptOptions) GetDB() *gorm.DB {
	return q.db
}

// ============================================================================
// 基础查询选项
// ============================================================================

// WithLoginAttemptID 按ID查询
func WithLoginAttemptID(id uint) QueryLoginAttemptOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("id = ?", id)
	}
}

// WithLoginAttemptUserID 按用户ID查询
func WithLoginAttemptUserID(userID string) QueryLoginAttemptOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("user_id = ?", userID)
	}
}

// WithLoginAttemptIPAddress 按IP地址查询
func WithLoginAttemptIPAddress(ipAddress string) QueryLoginAttemptOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("ip_address = ?", ipAddress)
	}
}

// WithLoginIdentity 按登录标识查询
func WithLoginIdentity(identity string) QueryLoginAttemptOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("login_identity = ?", identity)
	}
}

// WithAttemptType 按尝试类型查询
func WithAttemptType(attemptType string) QueryLoginAttemptOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("attempt_type = ?", attemptType)
	}
}

// WithFailureReason 按失败原因查询
func WithFailureReason(reason string) QueryLoginAttemptOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("failure_reason = ?", reason)
	}
}

// ============================================================================
// 时间范围查询选项
// ============================================================================

// WithTimeRange 按时间范围查询
func WithLoginAttemptTimeRange(startTime, endTime time.Time) QueryLoginAttemptOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("created_at BETWEEN ? AND ?", startTime, endTime)
	}
}

// WithSinceTime 查询指定时间之后的记录
func WithLoginAttemptSinceTime(since time.Time) QueryLoginAttemptOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("created_at >= ?", since)
	}
}

// WithBeforeTime 查询指定时间之前的记录
func WithLoginAttemptBeforeTime(before time.Time) QueryLoginAttemptOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("created_at <= ?", before)
	}
}

// WithRecentMinutes 查询最近几分钟的记录
func WithRecentMinutes(minutes int) QueryLoginAttemptOption {
	return func(db *gorm.DB) *gorm.DB {
		since := time.Now().Add(-time.Duration(minutes) * time.Minute)
		return db.Where("created_at >= ?", since)
	}
}

// WithRecentHours 查询最近几小时的记录
func WithRecentHours(hours int) QueryLoginAttemptOption {
	return func(db *gorm.DB) *gorm.DB {
		since := time.Now().Add(-time.Duration(hours) * time.Hour)
		return db.Where("created_at >= ?", since)
	}
}

// WithToday 查询今天的记录
func WithLoginAttemptToday() QueryLoginAttemptOption {
	return func(db *gorm.DB) *gorm.DB {
		today := time.Now().Truncate(24 * time.Hour)
		tomorrow := today.Add(24 * time.Hour)
		return db.Where("created_at >= ? AND created_at < ?", today, tomorrow)
	}
}

// ============================================================================
// 复合查询选项
// ============================================================================

// WithFailedAttempts 查询失败的尝试
func WithFailedAttempts() QueryLoginAttemptOption {
	return WithAttemptType("failed")
}

// WithSuccessfulAttempts 查询成功的尝试
func WithSuccessfulAttempts() QueryLoginAttemptOption {
	return WithAttemptType("success")
}

// WithBlockedAttempts 查询被阻止的尝试
func WithBlockedAttempts() QueryLoginAttemptOption {
	return WithAttemptType("blocked")
}

// WithUserFailedAttempts 查询用户的失败尝试
func WithUserFailedAttempts(userID string) QueryLoginAttemptOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("user_id = ? AND attempt_type = ?", userID, "failed")
	}
}

// WithIPFailedAttempts 查询IP的失败尝试
func WithIPFailedAttempts(ipAddress string) QueryLoginAttemptOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("ip_address = ? AND attempt_type = ?", ipAddress, "failed")
	}
}

// WithUserRecentFailedAttempts 查询用户最近的失败尝试
func WithUserRecentFailedAttempts(userID string, minutes int) QueryLoginAttemptOption {
	return func(db *gorm.DB) *gorm.DB {
		since := time.Now().Add(-time.Duration(minutes) * time.Minute)
		return db.Where("user_id = ? AND attempt_type = ? AND created_at >= ?", userID, "failed", since)
	}
}

// WithIPRecentFailedAttempts 查询IP最近的失败尝试
func WithIPRecentFailedAttempts(ipAddress string, minutes int) QueryLoginAttemptOption {
	return func(db *gorm.DB) *gorm.DB {
		since := time.Now().Add(-time.Duration(minutes) * time.Minute)
		return db.Where("ip_address = ? AND attempt_type = ? AND created_at >= ?", ipAddress, "failed", since)
	}
}

// WithIdentityRecentFailedAttempts 查询登录标识最近的失败尝试
func WithIdentityRecentFailedAttempts(identity string, minutes int) QueryLoginAttemptOption {
	return func(db *gorm.DB) *gorm.DB {
		since := time.Now().Add(-time.Duration(minutes) * time.Minute)
		return db.Where("login_identity = ? AND attempt_type = ? AND created_at >= ?", identity, "failed", since)
	}
}

// ============================================================================
// 地理位置查询选项
// ============================================================================

// WithLoginAttemptCountry 按国家查询
func WithLoginAttemptCountry(country string) QueryLoginAttemptOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("country = ?", country)
	}
}

// WithLoginAttemptCity 按城市查询
func WithLoginAttemptCity(city string) QueryLoginAttemptOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("city = ?", city)
	}
}

// WithLocation 按国家和城市查询
func WithLocation(country, city string) QueryLoginAttemptOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("country = ? AND city = ?", country, city)
	}
}

// ============================================================================
// 排序和分页选项
// ============================================================================

// WithLoginAttemptOrderByCreatedAt 按创建时间排序
func WithLoginAttemptOrderByCreatedAt(desc bool) QueryLoginAttemptOption {
	return func(db *gorm.DB) *gorm.DB {
		if desc {
			return db.Order("created_at DESC")
		}
		return db.Order("created_at ASC")
	}
}

// WithLoginAttemptOrderByID 按ID排序
func WithLoginAttemptOrderByID(desc bool) QueryLoginAttemptOption {
	return func(db *gorm.DB) *gorm.DB {
		if desc {
			return db.Order("id DESC")
		}
		return db.Order("id ASC")
	}
}

// WithLoginAttemptLimit 限制查询数量
func WithLoginAttemptLimit(limit int) QueryLoginAttemptOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Limit(limit)
	}
}

// WithLoginAttemptOffset 设置查询偏移
func WithLoginAttemptOffset(offset int) QueryLoginAttemptOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Offset(offset)
	}
}

// WithLoginAttemptPagination 分页查询
func WithLoginAttemptPagination(page, pageSize int) QueryLoginAttemptOption {
	return func(db *gorm.DB) *gorm.DB {
		offset := (page - 1) * pageSize
		return db.Offset(offset).Limit(pageSize)
	}
}

// ============================================================================
// 分组和聚合选项
// ============================================================================

// WithGroupByIP 按IP地址分组
func WithGroupByIP() QueryLoginAttemptOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Group("ip_address")
	}
}

// WithGroupByUserID 按用户ID分组
func WithGroupByUserID() QueryLoginAttemptOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Group("user_id")
	}
}

// WithGroupByAttemptType 按尝试类型分组
func WithGroupByAttemptType() QueryLoginAttemptOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Group("attempt_type")
	}
}

// WithHavingFailedCountGreaterThan 筛选失败次数大于指定值的记录
func WithHavingFailedCountGreaterThan(count int) QueryLoginAttemptOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Having("COUNT(*) > ?", count)
	}
}
