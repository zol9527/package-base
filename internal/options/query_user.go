package options

import (
	"time"

	"gorm.io/gorm"
)

// QueryUserOptions 用户查询选项
type QueryUserOptions struct {
	db *gorm.DB
}

// QueryUserOption 用户查询选项函数类型
type QueryUserOption func(*QueryUserOptions)

// NewQueryUserOptions 创建新的用户查询选项
func NewQueryUserOptions(db *gorm.DB, opts ...QueryUserOption) *QueryUserOptions {
	options := &QueryUserOptions{
		db: db,
	}

	// 应用所有选项
	for _, opt := range opts {
		opt(options)
	}

	return options
}

// GetDB 获取构建的数据库查询
func (q *QueryUserOptions) GetDB() *gorm.DB {
	return q.db
}

// 基础查询选项

// WithID 根据用户ID查询
func WithID(id string) QueryUserOption {
	return func(q *QueryUserOptions) {
		q.db = q.db.Where("id = ?", id)
	}
}

// WithUUID 根据用户UUID查询
func WithUUID(uuid string) QueryUserOption {
	return func(q *QueryUserOptions) {
		q.db = q.db.Where("uuid = ?", uuid)
	}
}

// WithUsername 根据用户名查询
func WithUsername(username string) QueryUserOption {
	return func(q *QueryUserOptions) {
		q.db = q.db.Where("username = ?", username)
	}
}

// WithEmail 根据邮箱查询
func WithEmail(email string) QueryUserOption {
	return func(q *QueryUserOptions) {
		q.db = q.db.Where("email = ?", email)
	}
}

// WithPhone 根据手机号查询
func WithPhone(phone string) QueryUserOption {
	return func(q *QueryUserOptions) {
		q.db = q.db.Where("phone = ?", phone)
	}
}

// WithUsernameOrEmail 根据用户名或邮箱查询
func WithUsernameOrEmail(usernameOrEmail string) QueryUserOption {
	return func(q *QueryUserOptions) {
		q.db = q.db.Where("username = ? OR email = ?", usernameOrEmail, usernameOrEmail)
	}
}

// 状态查询选项

// WithStatus 根据用户状态查询
func WithStatus(status int) QueryUserOption {
	return func(q *QueryUserOptions) {
		q.db = q.db.Where("status = ?", status)
	}
}

// WithActiveStatus 查询激活状态的用户
func WithActiveStatus() QueryUserOption {
	return func(q *QueryUserOptions) {
		q.db = q.db.Where("status = ?", 1)
	}
}

// WithInactiveStatus 查询未激活状态的用户
func WithInactiveStatus() QueryUserOption {
	return func(q *QueryUserOptions) {
		q.db = q.db.Where("status = ?", 0)
	}
}

// WithRole 根据用户角色查询
func WithRole(role int) QueryUserOption {
	return func(q *QueryUserOptions) {
		q.db = q.db.Where("roles = ?", role)
	}
}

// WithAdminRole 查询管理员角色用户
func WithAdminRole() QueryUserOption {
	return func(q *QueryUserOptions) {
		q.db = q.db.Where("roles = ?", 1)
	}
}

// WithUserRole 查询普通用户角色
func WithUserRole() QueryUserOption {
	return func(q *QueryUserOptions) {
		q.db = q.db.Where("roles = ?", 0)
	}
}

// 时间范围查询选项

// WithCreatedAfter 查询指定时间之后创建的用户
func WithCreatedAfter(createdAt time.Time) QueryUserOption {
	return func(q *QueryUserOptions) {
		q.db = q.db.Where("created_at > ?", createdAt)
	}
}

// WithCreatedBefore 查询指定时间之前创建的用户
func WithCreatedBefore(createdAt time.Time) QueryUserOption {
	return func(q *QueryUserOptions) {
		q.db = q.db.Where("created_at < ?", createdAt)
	}
}

// WithCreatedBetween 查询指定时间范围内创建的用户
func WithCreatedBetween(startTime, endTime time.Time) QueryUserOption {
	return func(q *QueryUserOptions) {
		q.db = q.db.Where("created_at BETWEEN ? AND ?", startTime, endTime)
	}
}

// WithLastLoginAfter 查询指定时间之后登录的用户
func WithLastLoginAfter(lastLoginAt time.Time) QueryUserOption {
	return func(q *QueryUserOptions) {
		q.db = q.db.Where("last_login_at > ?", lastLoginAt)
	}
}

// WithLastLoginBefore 查询指定时间之前登录的用户
func WithLastLoginBefore(lastLoginAt time.Time) QueryUserOption {
	return func(q *QueryUserOptions) {
		q.db = q.db.Where("last_login_at < ?", lastLoginAt)
	}
}

// 模糊查询选项

// WithUsernameLike 根据用户名模糊查询
func WithUsernameLike(username string) QueryUserOption {
	return func(q *QueryUserOptions) {
		q.db = q.db.Where("username LIKE ?", "%"+username+"%")
	}
}

// WithEmailLike 根据邮箱模糊查询
func WithEmailLike(email string) QueryUserOption {
	return func(q *QueryUserOptions) {
		q.db = q.db.Where("email LIKE ?", "%"+email+"%")
	}
}

// WithNicknameLike 根据昵称模糊查询
func WithNicknameLike(nickname string) QueryUserOption {
	return func(q *QueryUserOptions) {
		q.db = q.db.Where("nickname LIKE ?", "%"+nickname+"%")
	}
}

// 地理位置查询选项

// WithCountry 根据国家查询
func WithCountry(country string) QueryUserOption {
	return func(q *QueryUserOptions) {
		q.db = q.db.Where("country = ?", country)
	}
}

// WithProvince 根据省份查询
func WithProvince(province string) QueryUserOption {
	return func(q *QueryUserOptions) {
		q.db = q.db.Where("province = ?", province)
	}
}

// WithCity 根据城市查询
func WithCity(city string) QueryUserOption {
	return func(q *QueryUserOptions) {
		q.db = q.db.Where("city = ?", city)
	}
}

// 验证状态查询选项

// WithEmailVerified 查询已验证邮箱的用户
func WithEmailVerified() QueryUserOption {
	return func(q *QueryUserOptions) {
		q.db = q.db.Where("email_verified_at IS NOT NULL")
	}
}

// WithEmailUnverified 查询未验证邮箱的用户
func WithEmailUnverified() QueryUserOption {
	return func(q *QueryUserOptions) {
		q.db = q.db.Where("email_verified_at IS NULL")
	}
}

// WithPhoneVerified 查询已验证手机的用户
func WithPhoneVerified() QueryUserOption {
	return func(q *QueryUserOptions) {
		q.db = q.db.Where("phone_verified_at IS NOT NULL")
	}
}

// WithPhoneUnverified 查询未验证手机的用户
func WithPhoneUnverified() QueryUserOption {
	return func(q *QueryUserOptions) {
		q.db = q.db.Where("phone_verified_at IS NULL")
	}
}

// 安全选项

// WithTwoFactorEnabled 查询启用双因素认证的用户
func WithTwoFactorEnabled() QueryUserOption {
	return func(q *QueryUserOptions) {
		q.db = q.db.Where("two_factor_enabled = ?", true)
	}
}

// WithTwoFactorDisabled 查询未启用双因素认证的用户
func WithTwoFactorDisabled() QueryUserOption {
	return func(q *QueryUserOptions) {
		q.db = q.db.Where("two_factor_enabled = ?", false)
	}
}

// 数值查询选项

// WithLoginCountGreaterThan 查询登录次数大于指定值的用户
func WithLoginCountGreaterThan(count int) QueryUserOption {
	return func(q *QueryUserOptions) {
		q.db = q.db.Where("login_count > ?", count)
	}
}

// WithLoginCountLessThan 查询登录次数小于指定值的用户
func WithLoginCountLessThan(count int) QueryUserOption {
	return func(q *QueryUserOptions) {
		q.db = q.db.Where("login_count < ?", count)
	}
}

// WithLoginCountBetween 查询登录次数在指定范围内的用户
func WithLoginCountBetween(min, max int) QueryUserOption {
	return func(q *QueryUserOptions) {
		q.db = q.db.Where("login_count BETWEEN ? AND ?", min, max)
	}
}

// 排序选项

// WithOrderByCreatedAt 按创建时间排序
func WithOrderByCreatedAt(desc bool) QueryUserOption {
	return func(q *QueryUserOptions) {
		if desc {
			q.db = q.db.Order("created_at DESC")
		} else {
			q.db = q.db.Order("created_at ASC")
		}
	}
}

// WithOrderByLastLogin 按最后登录时间排序
func WithOrderByLastLogin(desc bool) QueryUserOption {
	return func(q *QueryUserOptions) {
		if desc {
			q.db = q.db.Order("last_login_at DESC")
		} else {
			q.db = q.db.Order("last_login_at ASC")
		}
	}
}

// WithOrderByLoginCount 按登录次数排序
func WithOrderByLoginCount(desc bool) QueryUserOption {
	return func(q *QueryUserOptions) {
		if desc {
			q.db = q.db.Order("login_count DESC")
		} else {
			q.db = q.db.Order("login_count ASC")
		}
	}
}

// WithOrderByUsername 按用户名排序
func WithOrderByUsername(desc bool) QueryUserOption {
	return func(q *QueryUserOptions) {
		if desc {
			q.db = q.db.Order("username DESC")
		} else {
			q.db = q.db.Order("username ASC")
		}
	}
}

// 分页选项

// WithLimit 限制查询结果数量
func WithLimit(limit int) QueryUserOption {
	return func(q *QueryUserOptions) {
		q.db = q.db.Limit(limit)
	}
}

// WithOffset 设置查询偏移量
func WithOffset(offset int) QueryUserOption {
	return func(q *QueryUserOptions) {
		q.db = q.db.Offset(offset)
	}
}

// WithPagination 设置分页
func WithPagination(page, pageSize int) QueryUserOption {
	return func(q *QueryUserOptions) {
		offset := (page - 1) * pageSize
		q.db = q.db.Offset(offset).Limit(pageSize)
	}
}

// 预加载选项

// WithoutDeleted 排除已删除的记录
func WithoutDeleted() QueryUserOption {
	return func(q *QueryUserOptions) {
		q.db = q.db.Where("deleted_at IS NULL")
	}
}

// 复合查询选项

// WithSearchKeyword 根据关键词搜索（用户名、邮箱、昵称）
func WithSearchKeyword(keyword string) QueryUserOption {
	return func(q *QueryUserOptions) {
		searchPattern := "%" + keyword + "%"
		q.db = q.db.Where("username LIKE ? OR email LIKE ? OR nickname LIKE ?",
			searchPattern, searchPattern, searchPattern)
	}
}

// WithActiveAndVerified 查询激活且已验证邮箱的用户
func WithActiveAndVerified() QueryUserOption {
	return func(q *QueryUserOptions) {
		q.db = q.db.Where("status = ? AND email_verified_at IS NOT NULL", 1)
	}
}

// WithInactiveUsers 查询非活跃用户（长时间未登录）
func WithInactiveUsers(days int) QueryUserOption {
	return func(q *QueryUserOptions) {
		cutoffTime := time.Now().AddDate(0, 0, -days)
		q.db = q.db.Where("last_login_at < ? OR last_login_at IS NULL", cutoffTime)
	}
}
