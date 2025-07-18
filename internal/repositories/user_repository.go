package repositories

import (
	"context"
	"package-base/internal/core/interfaces"
	"package-base/internal/models"
	"package-base/internal/options"
	repoInterfaces "package-base/internal/repositories/interfaces"
	"time"

	"github.com/pkg/errors"
	"gorm.io/gorm"
)

// userRepository 用户存储实现
type userRepository struct {
	storage interfaces.RelationalStorage
	logger  interfaces.Logger
}

// NewUserRepository 创建用户存储实现
func NewUserRepository(storage interfaces.RelationalStorage, logger interfaces.Logger) repoInterfaces.UserRepository {
	return &userRepository{
		storage: storage,
		logger:  logger,
	}
}

// getDB 获取数据库连接
func (r *userRepository) getDB() *gorm.DB {
	return r.storage.GetDB()
}

// Create 创建用户
func (r *userRepository) Create(ctx context.Context, user *models.User) error {
	db := r.getDB()

	if err := db.WithContext(ctx).Create(user).Error; err != nil {
		r.logger.Error("failed to create user",
			interfaces.ErrorField(err),
			interfaces.String("username", user.Username),
			interfaces.String("email", user.GetEmail()),
		)
		return errors.Wrap(err, "failed to create user")
	}

	r.logger.Info("user created successfully",
		interfaces.String("user_id", user.ID),
		interfaces.String("username", user.Username),
	)

	return nil
}

// GetByID 根据ID获取用户
func (r *userRepository) GetByID(ctx context.Context, id string) (*models.User, error) {
	r.logger.Debug("getting user by ID",
		interfaces.String("user_id", id),
	)

	user, err := r.FindUser(ctx, options.WithID(id))
	if err != nil {
		r.logger.Error("failed to get user by ID",
			interfaces.ErrorField(err),
			interfaces.String("user_id", id),
		)
		return nil, errors.Wrap(err, "failed to get user by ID")
	}

	if user == nil {
		r.logger.Debug("user not found by ID",
			interfaces.String("user_id", id),
		)
		return nil, nil
	}

	return user, nil
}

// GetByUsername 根据用户名获取用户
func (r *userRepository) GetByUsername(ctx context.Context, username string) (*models.User, error) {
	r.logger.Debug("getting user by username",
		interfaces.String("username", username),
	)

	user, err := r.FindUser(ctx, options.WithUsername(username))
	if err != nil {
		r.logger.Error("failed to get user by username",
			interfaces.ErrorField(err),
			interfaces.String("username", username),
		)
		return nil, errors.Wrap(err, "failed to get user by username")
	}

	if user == nil {
		r.logger.Debug("user not found by username",
			interfaces.String("username", username),
		)
		return nil, nil
	}

	return user, nil
}

// GetByEmail 根据邮箱获取用户
func (r *userRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	r.logger.Debug("getting user by email",
		interfaces.String("email", email),
	)

	user, err := r.FindUser(ctx, options.WithEmail(email))
	if err != nil {
		r.logger.Error("failed to get user by email",
			interfaces.ErrorField(err),
			interfaces.String("email", email),
		)
		return nil, errors.Wrap(err, "failed to get user by email")
	}

	if user == nil {
		r.logger.Debug("user not found by email",
			interfaces.String("email", email),
		)
		return nil, nil
	}

	return user, nil
}

// GetByUsernameOrEmail 根据用户名或邮箱获取用户
func (r *userRepository) GetByUsernameOrEmail(ctx context.Context, usernameOrEmail string) (*models.User, error) {
	r.logger.Debug("getting user by username or email",
		interfaces.String("username_or_email", usernameOrEmail),
	)

	user, err := r.FindUser(ctx, options.WithUsernameOrEmail(usernameOrEmail))
	if err != nil {
		r.logger.Error("failed to get user by username or email",
			interfaces.ErrorField(err),
			interfaces.String("username_or_email", usernameOrEmail),
		)
		return nil, errors.Wrap(err, "failed to get user by username or email")
	}

	if user == nil {
		r.logger.Debug("user not found by username or email",
			interfaces.String("username_or_email", usernameOrEmail),
		)
		return nil, nil
	}

	return user, nil
}

// Update 更新用户信息
func (r *userRepository) Update(ctx context.Context, user *models.User) error {
	db := r.getDB()

	if err := db.WithContext(ctx).Save(user).Error; err != nil {
		r.logger.Error("failed to update user",
			interfaces.ErrorField(err),
			interfaces.String("user_id", user.ID),
		)
		return errors.Wrap(err, "failed to update user")
	}

	r.logger.Info("user updated successfully",
		interfaces.String("user_id", user.ID),
	)

	return nil
}

// UpdatePassword 更新用户密码
func (r *userRepository) UpdatePassword(ctx context.Context, userID, newPasswordHash string) error {
	db := r.getDB()

	if err := db.WithContext(ctx).Model(&models.User{}).Where("id = ?", userID).Update("password_hash", newPasswordHash).Error; err != nil {
		r.logger.Error("failed to update user password",
			interfaces.ErrorField(err),
			interfaces.String("user_id", userID),
		)
		return errors.Wrap(err, "failed to update user password")
	}

	r.logger.Info("user password updated successfully",
		interfaces.String("user_id", userID),
	)

	return nil
}

// UpdateLoginInfo 更新登录信息
func (r *userRepository) UpdateLoginInfo(ctx context.Context, userID string) error {
	db := r.getDB()

	updates := map[string]interface{}{
		"login_count":    gorm.Expr("login_count + 1"),
		"last_login_at":  time.Now(),
		"last_active_at": time.Now(),
	}

	if err := db.WithContext(ctx).Model(&models.User{}).Where("id = ?", userID).Updates(updates).Error; err != nil {
		r.logger.Error("failed to update user login info",
			interfaces.ErrorField(err),
			interfaces.String("user_id", userID),
		)
		return errors.Wrap(err, "failed to update user login info")
	}

	return nil
}

// ExistsByUsername 检查用户名是否存在
func (r *userRepository) ExistsByUsername(ctx context.Context, username string) (bool, error) {
	r.logger.Debug("checking username existence",
		interfaces.String("username", username),
	)

	count, err := r.CountUsers(ctx, options.WithUsername(username))
	if err != nil {
		r.logger.Error("failed to check username existence",
			interfaces.ErrorField(err),
			interfaces.String("username", username),
		)
		return false, errors.Wrap(err, "failed to check username existence")
	}

	return count > 0, nil
}

// ExistsByEmail 检查邮箱是否存在
func (r *userRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	r.logger.Debug("checking email existence",
		interfaces.String("email", email),
	)

	count, err := r.CountUsers(ctx, options.WithEmail(email))
	if err != nil {
		r.logger.Error("failed to check email existence",
			interfaces.ErrorField(err),
			interfaces.String("email", email),
		)
		return false, errors.Wrap(err, "failed to check email existence")
	}

	return count > 0, nil
}

// Delete 删除用户（软删除）
func (r *userRepository) Delete(ctx context.Context, id string) error {
	db := r.getDB()

	if err := db.WithContext(ctx).Delete(&models.User{}, "id = ?", id).Error; err != nil {
		r.logger.Error("failed to delete user",
			interfaces.ErrorField(err),
			interfaces.String("user_id", id),
		)
		return errors.Wrap(err, "failed to delete user")
	}

	r.logger.Info("user deleted successfully",
		interfaces.String("user_id", id),
	)

	return nil
}

// GetList 获取用户列表（分页）
func (r *userRepository) GetList(ctx context.Context, offset, limit int) ([]*models.User, int64, error) {
	r.logger.Debug("getting user list",
		interfaces.Int("offset", offset),
		interfaces.Int("limit", limit),
	)

	// 使用查询选项获取用户列表和总数
	users, total, err := r.FindUsersWithPagination(ctx,
		options.WithOffset(offset),
		options.WithLimit(limit),
		options.WithOrderByCreatedAt(true), // 按创建时间倒序
	)
	if err != nil {
		r.logger.Error("failed to get user list", interfaces.ErrorField(err))
		return nil, 0, errors.Wrap(err, "failed to get user list")
	}

	return users, total, nil
}

// UpdateStatus 更新用户状态
func (r *userRepository) UpdateStatus(ctx context.Context, userID string, status models.UserStatus) error {
	r.logger.Debug("updating user status",
		interfaces.String("user_id", userID),
		interfaces.Int("status", int(status)),
	)

	// 1. 先验证用户是否存在
	user, err := r.FindUser(ctx, options.WithID(userID))
	if err != nil {
		r.logger.Error("failed to get user for status update",
			interfaces.ErrorField(err),
			interfaces.String("user_id", userID),
		)
		return errors.Wrap(err, "failed to get user for status update")
	}

	if user == nil {
		r.logger.Warn("user not found for status update",
			interfaces.String("user_id", userID),
		)
		return errors.New("user not found")
	}

	// 2. 更新状态
	db := r.getDB()
	if err := db.WithContext(ctx).Model(&models.User{}).Where("id = ?", userID).Update("status", status).Error; err != nil {
		r.logger.Error("failed to update user status",
			interfaces.ErrorField(err),
			interfaces.String("user_id", userID),
			interfaces.Int("status", int(status)),
		)
		return errors.Wrap(err, "failed to update user status")
	}

	r.logger.Info("user status updated successfully",
		interfaces.String("user_id", userID),
		interfaces.Int("old_status", int(user.Status)),
		interfaces.Int("new_status", int(status)),
	)

	return nil
}

// 使用查询选项的方法实现

// FindUsers 根据查询选项查找用户列表
func (r *userRepository) FindUsers(ctx context.Context, opts ...options.QueryUserOption) ([]*models.User, error) {
	db := r.getDB()

	queryOptions := options.NewQueryUserOptions(db, opts...)

	var users []*models.User
	if err := queryOptions.GetDB().WithContext(ctx).Find(&users).Error; err != nil {
		r.logger.Error("failed to find users with options", interfaces.ErrorField(err))
		return nil, errors.Wrap(err, "failed to find users")
	}

	return users, nil
}

// FindUser 根据查询选项查找单个用户
func (r *userRepository) FindUser(ctx context.Context, opts ...options.QueryUserOption) (*models.User, error) {
	db := r.getDB()

	queryOptions := options.NewQueryUserOptions(db, opts...)

	var user models.User
	if err := queryOptions.GetDB().WithContext(ctx).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		r.logger.Error("failed to find user with options", interfaces.ErrorField(err))
		return nil, errors.Wrap(err, "failed to find user")
	}

	return &user, nil
}

// CountUsers 根据查询选项统计用户数量
func (r *userRepository) CountUsers(ctx context.Context, opts ...options.QueryUserOption) (int64, error) {
	db := r.getDB()

	queryOptions := options.NewQueryUserOptions(db, opts...)

	var count int64
	if err := queryOptions.GetDB().WithContext(ctx).Model(&models.User{}).Count(&count).Error; err != nil {
		r.logger.Error("failed to count users with options", interfaces.ErrorField(err))
		return 0, errors.Wrap(err, "failed to count users")
	}

	return count, nil
}

// FindUsersWithPagination 根据查询选项查找用户列表（带分页信息）
func (r *userRepository) FindUsersWithPagination(ctx context.Context, opts ...options.QueryUserOption) ([]*models.User, int64, error) {
	// 先统计总数（不包含分页选项）
	countOpts := make([]options.QueryUserOption, 0, len(opts))
	for _, opt := range opts {
		// 排除分页相关的选项来统计总数
		// 这里我们需要过滤掉 Limit 和 Offset 选项
		countOpts = append(countOpts, opt)
	}

	total, err := r.CountUsers(ctx, countOpts...)
	if err != nil {
		return nil, 0, err
	}

	// 获取分页数据
	users, err := r.FindUsers(ctx, opts...)
	if err != nil {
		return nil, 0, err
	}

	return users, total, nil
}
