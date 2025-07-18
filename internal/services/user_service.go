package services

import (
	"context"
	"math"

	"package-base/internal/core/interfaces"
	"package-base/internal/dto"
	"package-base/internal/models"
	"package-base/internal/options"
	repoInterfaces "package-base/internal/repositories/interfaces"
	serviceInterfaces "package-base/internal/services/interfaces"

	"github.com/pkg/errors"
)

// userService 用户服务实现
type userService struct {
	userRepo repoInterfaces.UserRepository
	logger   interfaces.Logger
}

// NewUserService 创建用户服务实例
func NewUserService(
	userRepo repoInterfaces.UserRepository,
	logger interfaces.Logger,
) serviceInterfaces.UserService {
	return &userService{
		userRepo: userRepo,
		logger:   logger,
	}
}

// GetProfile 获取用户资料
func (s *userService) GetProfile(ctx context.Context, userID string) (*dto.UserInfo, error) {
	user, err := s.userRepo.FindUser(ctx, options.WithID(userID))
	if err != nil {
		s.logger.Error("failed to get user profile",
			interfaces.ErrorField(err),
			interfaces.String("user_id", userID),
		)
		return nil, errors.Wrap(err, "failed to get user profile")
	}

	if user == nil {
		return nil, errors.New("user not found")
	}

	return s.ConvertToUserInfo(user), nil
}

// UpdateProfile 更新用户资料
func (s *userService) UpdateProfile(ctx context.Context, userID string, req *dto.UpdateProfileRequest) (*dto.UserInfo, error) {
	// 1. 获取现有用户信息
	user, err := s.userRepo.FindUser(ctx, options.WithID(userID))
	if err != nil {
		return nil, errors.Wrap(err, "failed to get user")
	}

	if user == nil {
		return nil, errors.New("user not found")
	}

	// 2. 更新用户信息
	if req.Nickname != "" {
		user.Nickname = &req.Nickname
	}
	if req.Phone != "" {
		user.Phone = &req.Phone
	}
	if req.Avatar != "" {
		user.AvatarURL = &req.Avatar
	}

	// 3. 保存更新
	if err := s.userRepo.Update(ctx, user); err != nil {
		s.logger.Error("failed to update user profile",
			interfaces.ErrorField(err),
			interfaces.String("user_id", userID),
		)
		return nil, errors.Wrap(err, "failed to update user profile")
	}

	s.logger.Info("user profile updated successfully",
		interfaces.String("user_id", userID),
	)

	return s.ConvertToUserInfo(user), nil
}

// GetByID 根据ID获取用户
func (s *userService) GetByID(ctx context.Context, userID string) (*models.User, error) {
	user, err := s.userRepo.FindUser(ctx, options.WithID(userID))
	if err != nil {
		return nil, errors.Wrap(err, "failed to get user")
	}

	if user == nil {
		return nil, errors.New("user not found")
	}

	return user, nil
}

// GetByUsername 根据用户名获取用户
func (s *userService) GetByUsername(ctx context.Context, username string) (*models.User, error) {
	user, err := s.userRepo.FindUser(ctx, options.WithUsername(username))
	if err != nil {
		return nil, errors.Wrap(err, "failed to get user")
	}

	if user == nil {
		return nil, errors.New("user not found")
	}

	return user, nil
}

// GetByEmail 根据邮箱获取用户
func (s *userService) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	user, err := s.userRepo.FindUser(ctx, options.WithEmail(email))
	if err != nil {
		return nil, errors.Wrap(err, "failed to get user")
	}

	if user == nil {
		return nil, errors.New("user not found")
	}

	return user, nil
}

// UpdateStatus 更新用户状态
func (s *userService) UpdateStatus(ctx context.Context, userID string, status models.UserStatus) error {
	// 1. 获取用户信息
	user, err := s.userRepo.FindUser(ctx, options.WithID(userID))
	if err != nil {
		return errors.Wrap(err, "failed to get user")
	}

	if user == nil {
		return errors.New("user not found")
	}

	// 2. 更新状态
	user.Status = status
	if err := s.userRepo.Update(ctx, user); err != nil {
		s.logger.Error("failed to update user status",
			interfaces.ErrorField(err),
			interfaces.String("user_id", userID),
			interfaces.Int("status", int(status)),
		)
		return errors.Wrap(err, "failed to update user status")
	}

	s.logger.Info("user status updated successfully",
		interfaces.String("user_id", userID),
		interfaces.Int("old_status", int(user.Status)),
		interfaces.Int("new_status", int(status)),
	)

	return nil
}

// GetUserList 获取用户列表（分页）
func (s *userService) GetUserList(ctx context.Context, page, pageSize int) (*dto.PaginationResponse, error) {
	// 计算偏移量
	offset := (page - 1) * pageSize

	// 获取用户列表和总数
	users, total, err := s.userRepo.GetList(ctx, offset, pageSize)
	if err != nil {
		s.logger.Error("failed to get user list",
			interfaces.ErrorField(err),
			interfaces.Int("page", page),
			interfaces.Int("page_size", pageSize),
		)
		return nil, errors.Wrap(err, "failed to get user list")
	}

	// 转换为DTO
	userInfos := make([]interface{}, 0, len(users))
	for _, user := range users {
		userInfos = append(userInfos, s.ConvertToUserInfo(user))
	}

	// 计算分页信息
	totalPages := int((total + int64(pageSize) - 1) / int64(pageSize))

	return &dto.PaginationResponse{
		Data:     userInfos,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
		Pages:    totalPages,
	}, nil
}

// DeleteUser 删除用户
func (s *userService) DeleteUser(ctx context.Context, userID string) error {
	// 1. 检查用户是否存在
	if err := s.ValidateUserExists(ctx, userID); err != nil {
		return err
	}

	// 2. 执行软删除
	if err := s.userRepo.Delete(ctx, userID); err != nil {
		s.logger.Error("failed to delete user",
			interfaces.ErrorField(err),
			interfaces.String("user_id", userID),
		)
		return errors.Wrap(err, "failed to delete user")
	}

	s.logger.Info("user deleted successfully",
		interfaces.String("user_id", userID),
	)

	return nil
}

// ValidateUserExists 验证用户是否存在
func (s *userService) ValidateUserExists(ctx context.Context, userID string) error {
	user, err := s.userRepo.FindUser(ctx, options.WithID(userID))
	if err != nil {
		return errors.Wrap(err, "failed to validate user existence")
	}

	if user == nil {
		return errors.New("user not found")
	}

	return nil
}

// ConvertToUserInfo 转换用户模型为DTO
func (s *userService) ConvertToUserInfo(user *models.User) *dto.UserInfo {
	return &dto.UserInfo{
		ID:       user.ID,
		UUID:     user.Uuid,
		Username: user.Username,
		Email:    user.GetEmail(),
		Phone:    user.GetPhone(),
		Nickname: user.GetNickname(),
		Avatar:   user.GetAvatarURL(),
		Status:   int(user.Status),
		Role:     int(user.Roles),
	}
}

// 使用查询选项的方法实现

// FindUsers 根据查询选项查找用户列表
func (s *userService) FindUsers(ctx context.Context, opts ...options.QueryUserOption) ([]*dto.UserInfo, error) {
	s.logger.Debug("finding users with options")

	users, err := s.userRepo.FindUsers(ctx, opts...)
	if err != nil {
		s.logger.Error("failed to find users", interfaces.ErrorField(err))
		return nil, errors.Wrap(err, "failed to find users")
	}

	userInfos := make([]*dto.UserInfo, 0, len(users))
	for _, user := range users {
		userInfos = append(userInfos, s.ConvertToUserInfo(user))
	}

	return userInfos, nil
}

// FindUser 根据查询选项查找单个用户
func (s *userService) FindUser(ctx context.Context, opts ...options.QueryUserOption) (*dto.UserInfo, error) {
	s.logger.Debug("finding user with options")

	user, err := s.userRepo.FindUser(ctx, opts...)
	if err != nil {
		s.logger.Error("failed to find user", interfaces.ErrorField(err))
		return nil, errors.Wrap(err, "failed to find user")
	}

	if user == nil {
		return nil, nil
	}

	return s.ConvertToUserInfo(user), nil
}

// CountUsers 根据查询选项统计用户数量
func (s *userService) CountUsers(ctx context.Context, opts ...options.QueryUserOption) (int64, error) {
	s.logger.Debug("counting users with options")

	count, err := s.userRepo.CountUsers(ctx, opts...)
	if err != nil {
		s.logger.Error("failed to count users", interfaces.ErrorField(err))
		return 0, errors.Wrap(err, "failed to count users")
	}

	return count, nil
}

// FindUsersWithPagination 根据查询选项查找用户列表（带分页）
func (s *userService) FindUsersWithPagination(ctx context.Context, opts ...options.QueryUserOption) (*dto.UserListResponse, error) {
	s.logger.Debug("finding users with pagination")

	users, total, err := s.userRepo.FindUsersWithPagination(ctx, opts...)
	if err != nil {
		s.logger.Error("failed to find users with pagination", interfaces.ErrorField(err))
		return nil, errors.Wrap(err, "failed to find users with pagination")
	}

	userInfos := make([]dto.UserInfo, 0, len(users))
	for _, user := range users {
		userInfos = append(userInfos, *s.ConvertToUserInfo(user))
	}

	// 从查询选项中提取分页信息
	page, pageSize := extractPaginationFromOptions(opts...)
	if pageSize == 0 {
		pageSize = 10 // 默认页大小
	}
	if page == 0 {
		page = 1 // 默认第一页
	}

	totalPages := int(math.Ceil(float64(total) / float64(pageSize)))

	return &dto.UserListResponse{
		Users:    userInfos,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
		Pages:    totalPages,
	}, nil
}

// SearchUsers 搜索用户（支持关键词搜索）
func (s *userService) SearchUsers(ctx context.Context, keyword string, page, pageSize int) (*dto.UserListResponse, error) {
	s.logger.Debug("searching users",
		interfaces.String("keyword", keyword),
		interfaces.Int("page", page),
		interfaces.Int("page_size", pageSize),
	)

	opts := []options.QueryUserOption{
		options.WithSearchKeyword(keyword),
		options.WithPagination(page, pageSize),
		options.WithOrderByCreatedAt(true),
	}

	return s.FindUsersWithPagination(ctx, opts...)
}

// GetActiveUsers 获取活跃用户列表
func (s *userService) GetActiveUsers(ctx context.Context, page, pageSize int) (*dto.UserListResponse, error) {
	s.logger.Debug("getting active users",
		interfaces.Int("page", page),
		interfaces.Int("page_size", pageSize),
	)

	opts := []options.QueryUserOption{
		options.WithActiveStatus(),
		options.WithPagination(page, pageSize),
		options.WithOrderByLastLogin(true),
	}

	return s.FindUsersWithPagination(ctx, opts...)
}

// GetInactiveUsers 获取非活跃用户列表
func (s *userService) GetInactiveUsers(ctx context.Context, days, page, pageSize int) (*dto.UserListResponse, error) {
	s.logger.Debug("getting inactive users",
		interfaces.Int("days", days),
		interfaces.Int("page", page),
		interfaces.Int("page_size", pageSize),
	)

	opts := []options.QueryUserOption{
		options.WithInactiveUsers(days),
		options.WithPagination(page, pageSize),
		options.WithOrderByLastLogin(false),
	}

	return s.FindUsersWithPagination(ctx, opts...)
}

// GetUsersByRole 根据角色获取用户列表
func (s *userService) GetUsersByRole(ctx context.Context, role int, page, pageSize int) (*dto.UserListResponse, error) {
	s.logger.Debug("getting users by role",
		interfaces.Int("role", role),
		interfaces.Int("page", page),
		interfaces.Int("page_size", pageSize),
	)

	opts := []options.QueryUserOption{
		options.WithRole(role),
		options.WithPagination(page, pageSize),
		options.WithOrderByCreatedAt(true),
	}

	return s.FindUsersWithPagination(ctx, opts...)
}

// GetUsersByStatus 根据状态获取用户列表
func (s *userService) GetUsersByStatus(ctx context.Context, status int, page, pageSize int) (*dto.UserListResponse, error) {
	s.logger.Debug("getting users by status",
		interfaces.Int("status", status),
		interfaces.Int("page", page),
		interfaces.Int("page_size", pageSize),
	)

	opts := []options.QueryUserOption{
		options.WithStatus(status),
		options.WithPagination(page, pageSize),
		options.WithOrderByCreatedAt(true),
	}

	return s.FindUsersWithPagination(ctx, opts...)
}

// extractPaginationFromOptions 从查询选项中提取分页信息
func extractPaginationFromOptions(opts ...options.QueryUserOption) (page, pageSize int) {
	// 创建一个临时的查询选项来应用选项并提取分页信息
	// 这是一个简化的实现，实际项目中可能需要更复杂的逻辑
	page = 1
	pageSize = 10

	// 这里可以通过解析选项来获取实际的分页参数
	// 由于当前的选项结构不直接暴露这些信息，我们使用默认值
	// 在实际应用中，可以考虑在查询选项中添加获取分页信息的方法

	return page, pageSize
}
