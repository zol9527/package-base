package interfaces

import (
	"context"
	"package-base/internal/dto"
	"package-base/internal/models"
	"package-base/internal/options"
)

// UserService 用户服务接口
type UserService interface {
	// GetProfile 获取用户资料
	GetProfile(ctx context.Context, userID string) (*dto.UserInfo, error)

	// UpdateProfile 更新用户资料
	UpdateProfile(ctx context.Context, userID string, req *dto.UpdateProfileRequest) (*dto.UserInfo, error)

	// GetByID 根据ID获取用户
	GetByID(ctx context.Context, userID string) (*models.User, error)

	// GetByUsername 根据用户名获取用户
	GetByUsername(ctx context.Context, username string) (*models.User, error)

	// GetByEmail 根据邮箱获取用户
	GetByEmail(ctx context.Context, email string) (*models.User, error)

	// UpdateStatus 更新用户状态
	UpdateStatus(ctx context.Context, userID string, status models.UserStatus) error

	// GetUserList 获取用户列表（分页）
	GetUserList(ctx context.Context, page, pageSize int) (*dto.PaginationResponse, error)

	// DeleteUser 删除用户
	DeleteUser(ctx context.Context, userID string) error

	// ValidateUserExists 验证用户是否存在
	ValidateUserExists(ctx context.Context, userID string) error

	// ConvertToUserInfo 转换用户模型为DTO
	ConvertToUserInfo(user *models.User) *dto.UserInfo

	// 使用查询选项的方法

	// FindUsers 根据查询选项查找用户列表
	FindUsers(ctx context.Context, opts ...options.QueryUserOption) ([]*dto.UserInfo, error)

	// FindUser 根据查询选项查找单个用户
	FindUser(ctx context.Context, opts ...options.QueryUserOption) (*dto.UserInfo, error)

	// CountUsers 根据查询选项统计用户数量
	CountUsers(ctx context.Context, opts ...options.QueryUserOption) (int64, error)

	// FindUsersWithPagination 根据查询选项查找用户列表（带分页）
	FindUsersWithPagination(ctx context.Context, opts ...options.QueryUserOption) (*dto.UserListResponse, error)

	// SearchUsers 搜索用户（支持关键词搜索）
	SearchUsers(ctx context.Context, keyword string, page, pageSize int) (*dto.UserListResponse, error)

	// GetActiveUsers 获取活跃用户列表
	GetActiveUsers(ctx context.Context, page, pageSize int) (*dto.UserListResponse, error)

	// GetInactiveUsers 获取非活跃用户列表
	GetInactiveUsers(ctx context.Context, days, page, pageSize int) (*dto.UserListResponse, error)

	// GetUsersByRole 根据角色获取用户列表
	GetUsersByRole(ctx context.Context, role int, page, pageSize int) (*dto.UserListResponse, error)

	// GetUsersByStatus 根据状态获取用户列表
	GetUsersByStatus(ctx context.Context, status int, page, pageSize int) (*dto.UserListResponse, error)
}
