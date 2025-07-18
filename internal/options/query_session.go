package options

import (
	"time"

	"gorm.io/gorm"
)

// QuerySessionOptions 会话查询选项
type QuerySessionOptions struct {
	db *gorm.DB
}

// QuerySessionOption 会话查询选项函数类型
type QuerySessionOption func(*QuerySessionOptions)

// NewQuerySessionOptions 创建新的会话查询选项
func NewQuerySessionOptions(db *gorm.DB, opts ...QuerySessionOption) *QuerySessionOptions {
	options := &QuerySessionOptions{
		db: db,
	}

	// 应用所有选项
	for _, opt := range opts {
		opt(options)
	}

	return options
}

// GetDB 获取构建的数据库查询
func (q *QuerySessionOptions) GetDB() *gorm.DB {
	return q.db
}

// 基础查询选项

// WithSessionID 根据会话ID查询
func WithSessionID(id string) QuerySessionOption {
	return func(q *QuerySessionOptions) {
		q.db = q.db.Where("id = ?", id)
	}
}

// WithExcludeSessionID 排除指定会话ID
func WithExcludeSessionID(id string) QuerySessionOption {
	return func(q *QuerySessionOptions) {
		q.db = q.db.Where("id != ?", id)
	}
}

// WithUserID 根据用户ID查询
func WithUserID(userID string) QuerySessionOption {
	return func(q *QuerySessionOptions) {
		q.db = q.db.Where("user_id = ?", userID)
	}
}

// WithRefreshToken 根据刷新Token查询
func WithRefreshToken(refreshToken string) QuerySessionOption {
	return func(q *QuerySessionOptions) {
		q.db = q.db.Where("refresh_token = ?", refreshToken)
	}
}

// WithAccessToken 根据访问Token查询
func WithAccessToken(accessToken string) QuerySessionOption {
	return func(q *QuerySessionOptions) {
		q.db = q.db.Where("access_token = ?", accessToken)
	}
}

// WithIPAddress 根据IP地址查询
func WithIPAddress(ipAddress string) QuerySessionOption {
	return func(q *QuerySessionOptions) {
		q.db = q.db.Where("ip_address = ?", ipAddress)
	}
}

// WithUserAgent 根据用户代理查询
func WithUserAgent(userAgent string) QuerySessionOption {
	return func(q *QuerySessionOptions) {
		q.db = q.db.Where("user_agent = ?", userAgent)
	}
}

// WithDeviceInfo 根据设备信息查询
func WithDeviceInfo(deviceInfo string) QuerySessionOption {
	return func(q *QuerySessionOptions) {
		q.db = q.db.Where("device_info = ?", deviceInfo)
	}
}

// 状态查询选项

// WithActiveSession 查询激活状态的会话
func WithActiveSession() QuerySessionOption {
	return func(q *QuerySessionOptions) {
		q.db = q.db.Where("is_active = ?", true)
	}
}

// WithInactiveSession 查询非激活状态的会话
func WithInactiveSession() QuerySessionOption {
	return func(q *QuerySessionOptions) {
		q.db = q.db.Where("is_active = ?", false)
	}
}

// WithActiveAndValidSession 查询激活且未过期的会话
func WithActiveAndValidSession() QuerySessionOption {
	return func(q *QuerySessionOptions) {
		q.db = q.db.Where("is_active = ? AND expires_at > ?", true, time.Now())
	}
}

// WithExpiredSession 查询已过期的会话
func WithExpiredSession() QuerySessionOption {
	return func(q *QuerySessionOptions) {
		q.db = q.db.Where("expires_at < ?", time.Now())
	}
}

// WithValidSession 查询未过期的会话
func WithValidSession() QuerySessionOption {
	return func(q *QuerySessionOptions) {
		q.db = q.db.Where("expires_at > ?", time.Now())
	}
}

// 时间范围查询选项

// WithCreatedAfter 查询指定时间之后创建的会话
func WithSessionCreatedAfter(createdAt time.Time) QuerySessionOption {
	return func(q *QuerySessionOptions) {
		q.db = q.db.Where("created_at > ?", createdAt)
	}
}

// WithCreatedBefore 查询指定时间之前创建的会话
func WithSessionCreatedBefore(createdAt time.Time) QuerySessionOption {
	return func(q *QuerySessionOptions) {
		q.db = q.db.Where("created_at < ?", createdAt)
	}
}

// WithCreatedBetween 查询指定时间范围内创建的会话
func WithSessionCreatedBetween(startTime, endTime time.Time) QuerySessionOption {
	return func(q *QuerySessionOptions) {
		q.db = q.db.Where("created_at BETWEEN ? AND ?", startTime, endTime)
	}
}

// WithUpdatedAfter 查询指定时间之后更新的会话
func WithSessionUpdatedAfter(updatedAt time.Time) QuerySessionOption {
	return func(q *QuerySessionOptions) {
		q.db = q.db.Where("updated_at > ?", updatedAt)
	}
}

// WithUpdatedBefore 查询指定时间之前更新的会话
func WithSessionUpdatedBefore(updatedAt time.Time) QuerySessionOption {
	return func(q *QuerySessionOptions) {
		q.db = q.db.Where("updated_at < ?", updatedAt)
	}
}

// WithExpiresAfter 查询指定时间之后过期的会话
func WithExpiresAfter(expiresAt time.Time) QuerySessionOption {
	return func(q *QuerySessionOptions) {
		q.db = q.db.Where("expires_at > ?", expiresAt)
	}
}

// WithExpiresBefore 查询指定时间之前过期的会话
func WithExpiresBefore(expiresAt time.Time) QuerySessionOption {
	return func(q *QuerySessionOptions) {
		q.db = q.db.Where("expires_at < ?", expiresAt)
	}
}

// 模糊查询选项

// WithIPAddressLike 根据IP地址模糊查询
func WithIPAddressLike(ipAddress string) QuerySessionOption {
	return func(q *QuerySessionOptions) {
		q.db = q.db.Where("ip_address LIKE ?", "%"+ipAddress+"%")
	}
}

// WithUserAgentLike 根据用户代理模糊查询
func WithUserAgentLike(userAgent string) QuerySessionOption {
	return func(q *QuerySessionOptions) {
		q.db = q.db.Where("user_agent LIKE ?", "%"+userAgent+"%")
	}
}

// WithDeviceInfoLike 根据设备信息模糊查询
func WithDeviceInfoLike(deviceInfo string) QuerySessionOption {
	return func(q *QuerySessionOptions) {
		q.db = q.db.Where("device_info LIKE ?", "%"+deviceInfo+"%")
	}
}

// 排序选项

// WithOrderByCreatedAt 按创建时间排序
func WithSessionOrderByCreatedAt(desc bool) QuerySessionOption {
	return func(q *QuerySessionOptions) {
		if desc {
			q.db = q.db.Order("created_at DESC")
		} else {
			q.db = q.db.Order("created_at ASC")
		}
	}
}

// WithOrderByUpdatedAt 按更新时间排序
func WithSessionOrderByUpdatedAt(desc bool) QuerySessionOption {
	return func(q *QuerySessionOptions) {
		if desc {
			q.db = q.db.Order("updated_at DESC")
		} else {
			q.db = q.db.Order("updated_at ASC")
		}
	}
}

// WithOrderByExpiresAt 按过期时间排序
func WithSessionOrderByExpiresAt(desc bool) QuerySessionOption {
	return func(q *QuerySessionOptions) {
		if desc {
			q.db = q.db.Order("expires_at DESC")
		} else {
			q.db = q.db.Order("expires_at ASC")
		}
	}
}

// 分页选项

// WithSessionLimit 限制查询结果数量
func WithSessionLimit(limit int) QuerySessionOption {
	return func(q *QuerySessionOptions) {
		q.db = q.db.Limit(limit)
	}
}

// WithSessionOffset 设置查询偏移量
func WithSessionOffset(offset int) QuerySessionOption {
	return func(q *QuerySessionOptions) {
		q.db = q.db.Offset(offset)
	}
}

// WithSessionPagination 设置分页
func WithSessionPagination(page, pageSize int) QuerySessionOption {
	return func(q *QuerySessionOptions) {
		offset := (page - 1) * pageSize
		q.db = q.db.Offset(offset).Limit(pageSize)
	}
}

// 预加载选项

// WithUserPreload 预加载用户信息
func WithUserPreload() QuerySessionOption {
	return func(q *QuerySessionOptions) {
		q.db = q.db.Preload("User")
	}
}

// WithoutDeletedSessions 排除已删除的会话记录
func WithoutDeletedSessions() QuerySessionOption {
	return func(q *QuerySessionOptions) {
		q.db = q.db.Where("deleted_at IS NULL")
	}
}

// 复合查询选项

// WithActiveRefreshToken 查询活跃且有效的刷新Token会话
func WithActiveRefreshToken(refreshToken string) QuerySessionOption {
	return func(q *QuerySessionOptions) {
		q.db = q.db.Where("refresh_token = ? AND is_active = ? AND expires_at > ?",
			refreshToken, true, time.Now())
	}
}

// WithUserActiveSessions 查询用户的所有活跃会话
func WithUserActiveSessions(userID string) QuerySessionOption {
	return func(q *QuerySessionOptions) {
		q.db = q.db.Where("user_id = ? AND is_active = ? AND expires_at > ?",
			userID, true, time.Now())
	}
}

// WithOldInactiveSessions 查询旧的非活跃会话（用于清理）
func WithOldInactiveSessions(days int) QuerySessionOption {
	return func(q *QuerySessionOptions) {
		cutoffTime := time.Now().AddDate(0, 0, -days)
		q.db = q.db.Where("is_active = ? AND updated_at < ?", false, cutoffTime)
	}
}

// WithSessionsByUserAndDevice 根据用户和设备信息查询会话
func WithSessionsByUserAndDevice(userID, deviceInfo string) QuerySessionOption {
	return func(q *QuerySessionOptions) {
		q.db = q.db.Where("user_id = ? AND device_info = ?", userID, deviceInfo)
	}
}

// WithSessionsByIPRange 根据IP地址范围查询会话（简单的前缀匹配）
func WithSessionsByIPRange(ipPrefix string) QuerySessionOption {
	return func(q *QuerySessionOptions) {
		q.db = q.db.Where("ip_address LIKE ?", ipPrefix+"%")
	}
}

// WithRecentSessions 查询最近的会话（指定天数内）
func WithRecentSessions(days int) QuerySessionOption {
	return func(q *QuerySessionOptions) {
		cutoffTime := time.Now().AddDate(0, 0, -days)
		q.db = q.db.Where("created_at > ?", cutoffTime)
	}
}

// WithLongLivedSessions 查询长期存在的会话（超过指定天数）
func WithLongLivedSessions(days int) QuerySessionOption {
	return func(q *QuerySessionOptions) {
		cutoffTime := time.Now().AddDate(0, 0, -days)
		q.db = q.db.Where("created_at < ? AND is_active = ?", cutoffTime, true)
	}
}

// WithUserAndExcludeSession 查询用户的其他会话（排除指定会话）
func WithUserAndExcludeSession(userID, sessionID string) QuerySessionOption {
	return func(q *QuerySessionOptions) {
		q.db = q.db.Where("user_id = ? AND id != ?", userID, sessionID)
	}
}

// WithUserActiveSessionsExcluding 查询用户的其他活跃会话（排除指定会话）
func WithUserActiveSessionsExcluding(userID, sessionID string) QuerySessionOption {
	return func(q *QuerySessionOptions) {
		q.db = q.db.Where("user_id = ? AND id != ? AND is_active = ? AND expires_at > ?",
			userID, sessionID, true, time.Now())
	}
}

// WithSessionForUpdate 查询指定会话用于更新操作
func WithSessionForUpdate(sessionID string) QuerySessionOption {
	return func(q *QuerySessionOptions) {
		q.db = q.db.Where("id = ?", sessionID)
	}
}
