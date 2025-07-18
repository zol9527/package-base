package interfaces

import (
	"context"
	"package-base/internal/models"
	"package-base/internal/options"
)

// UserRepository 用户存储接口
type UserRepository interface {
	// Create 创建用户
	Create(ctx context.Context, user *models.User) error

	// GetByID 根据ID获取用户
	GetByID(ctx context.Context, id string) (*models.User, error)

	// GetByUsername 根据用户名获取用户
	GetByUsername(ctx context.Context, username string) (*models.User, error)

	// GetByEmail 根据邮箱获取用户
	GetByEmail(ctx context.Context, email string) (*models.User, error)

	// GetByUsernameOrEmail 根据用户名或邮箱获取用户
	GetByUsernameOrEmail(ctx context.Context, usernameOrEmail string) (*models.User, error)

	// Update 更新用户信息
	Update(ctx context.Context, user *models.User) error

	// UpdatePassword 更新用户密码
	UpdatePassword(ctx context.Context, userID, newPasswordHash string) error

	// UpdateLoginInfo 更新登录信息（登录次数、最后登录时间等）
	UpdateLoginInfo(ctx context.Context, userID string) error

	// ExistsByUsername 检查用户名是否存在
	ExistsByUsername(ctx context.Context, username string) (bool, error)

	// ExistsByEmail 检查邮箱是否存在
	ExistsByEmail(ctx context.Context, email string) (bool, error)

	// Delete 删除用户（软删除）
	Delete(ctx context.Context, id string) error

	// GetList 获取用户列表（分页）
	GetList(ctx context.Context, offset, limit int) ([]*models.User, int64, error)

	// UpdateStatus 更新用户状态
	UpdateStatus(ctx context.Context, userID string, status models.UserStatus) error

	// 使用查询选项的方法

	// FindUsers 根据查询选项查找用户列表
	FindUsers(ctx context.Context, opts ...options.QueryUserOption) ([]*models.User, error)

	// FindUser 根据查询选项查找单个用户
	FindUser(ctx context.Context, opts ...options.QueryUserOption) (*models.User, error)

	// CountUsers 根据查询选项统计用户数量
	CountUsers(ctx context.Context, opts ...options.QueryUserOption) (int64, error)

	// FindUsersWithPagination 根据查询选项查找用户列表（带分页信息）
	FindUsersWithPagination(ctx context.Context, opts ...options.QueryUserOption) ([]*models.User, int64, error)
}
