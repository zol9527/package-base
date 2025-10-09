package repositories

import (
	"context"
	"errors"
	"package-base/internal/core/interfaces"
	"package-base/internal/models"
	"package-base/internal/options"
	repoInterfaces "package-base/internal/repositories/interfaces"
)

// Function: NewUserRepositoryMySQL
//
// Description: MySQL 用户仓库占位实现（当前未启用驱动）。所有方法返回 driver not enabled 错误，以满足任务 T035 占位需求。
//
// Parameters:
//
//	storage(interfaces.RelationalStorage): 关系型存储实例（未使用）
//	logger(interfaces.Logger): 日志实例（未使用）
//
// Returns:
//
//	(repoInterfaces.UserRepository): 用户仓库接口占位实现
func NewUserRepositoryMySQL(storage interfaces.RelationalStorage, logger interfaces.Logger) repoInterfaces.UserRepository {
	return &userRepositoryMySQL{}
}

type userRepositoryMySQL struct{}

var errDriverNotEnabled = errors.New("driver not enabled")

func (r *userRepositoryMySQL) Create(ctx context.Context, user *models.User) error {
	return errDriverNotEnabled
}
func (r *userRepositoryMySQL) GetByID(ctx context.Context, id string) (*models.User, error) {
	return nil, errDriverNotEnabled
}
func (r *userRepositoryMySQL) GetByUsername(ctx context.Context, username string) (*models.User, error) {
	return nil, errDriverNotEnabled
}
func (r *userRepositoryMySQL) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	return nil, errDriverNotEnabled
}
func (r *userRepositoryMySQL) GetByUsernameOrEmail(ctx context.Context, s string) (*models.User, error) {
	return nil, errDriverNotEnabled
}
func (r *userRepositoryMySQL) FindByUsernameOrEmail(ctx context.Context, s string) (*models.User, error) {
	return nil, errDriverNotEnabled
}
func (r *userRepositoryMySQL) Update(ctx context.Context, user *models.User) error {
	return errDriverNotEnabled
}
func (r *userRepositoryMySQL) UpdatePassword(ctx context.Context, userID, hash string) error {
	return errDriverNotEnabled
}
func (r *userRepositoryMySQL) UpdateLoginInfo(ctx context.Context, userID string) error {
	return errDriverNotEnabled
}
func (r *userRepositoryMySQL) UpdateLoginMeta(ctx context.Context, userID string) error {
	return errDriverNotEnabled
}
func (r *userRepositoryMySQL) ExistsByUsername(ctx context.Context, username string) (bool, error) {
	return false, errDriverNotEnabled
}
func (r *userRepositoryMySQL) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	return false, errDriverNotEnabled
}
func (r *userRepositoryMySQL) Delete(ctx context.Context, id string) error {
	return errDriverNotEnabled
}
func (r *userRepositoryMySQL) GetList(ctx context.Context, offset, limit int) ([]*models.User, int64, error) {
	return nil, 0, errDriverNotEnabled
}
func (r *userRepositoryMySQL) UpdateStatus(ctx context.Context, userID string, status models.UserStatus) error {
	return errDriverNotEnabled
}
func (r *userRepositoryMySQL) FindUsers(ctx context.Context, opts ...options.QueryUserOption) ([]*models.User, error) {
	return nil, errDriverNotEnabled
}
func (r *userRepositoryMySQL) FindUser(ctx context.Context, opts ...options.QueryUserOption) (*models.User, error) {
	return nil, errDriverNotEnabled
}
func (r *userRepositoryMySQL) CountUsers(ctx context.Context, opts ...options.QueryUserOption) (int64, error) {
	return 0, errDriverNotEnabled
}
func (r *userRepositoryMySQL) FindUsersWithPagination(ctx context.Context, opts ...options.QueryUserOption) ([]*models.User, int64, error) {
	return nil, 0, errDriverNotEnabled
}
