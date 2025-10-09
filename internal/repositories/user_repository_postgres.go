package repositories

import (
	"context"
	"errors"
	"package-base/internal/core/interfaces"
	"package-base/internal/models"
	"package-base/internal/options"
	repoInterfaces "package-base/internal/repositories/interfaces"
)

// Function: NewUserRepositoryPostgres
//
// Description: Postgres 用户仓库占位实现（当前未启用驱动）。所有方法返回 driver not enabled 错误，以满足任务 T035 占位需求。
//
// Parameters:
//
//	storage(interfaces.RelationalStorage): 关系型存储实例（未使用）
//	logger(interfaces.Logger): 日志实例（未使用）
//
// Returns:
//
//	(repoInterfaces.UserRepository): 用户仓库接口占位实现
func NewUserRepositoryPostgres(storage interfaces.RelationalStorage, logger interfaces.Logger) repoInterfaces.UserRepository {
	return &userRepositoryPostgres{}
}

type userRepositoryPostgres struct{}

var errDriverNotEnabledPg = errors.New("driver not enabled")

func (r *userRepositoryPostgres) Create(ctx context.Context, user *models.User) error {
	return errDriverNotEnabledPg
}
func (r *userRepositoryPostgres) GetByID(ctx context.Context, id string) (*models.User, error) {
	return nil, errDriverNotEnabledPg
}
func (r *userRepositoryPostgres) GetByUsername(ctx context.Context, username string) (*models.User, error) {
	return nil, errDriverNotEnabledPg
}
func (r *userRepositoryPostgres) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	return nil, errDriverNotEnabledPg
}
func (r *userRepositoryPostgres) GetByUsernameOrEmail(ctx context.Context, s string) (*models.User, error) {
	return nil, errDriverNotEnabledPg
}
func (r *userRepositoryPostgres) FindByUsernameOrEmail(ctx context.Context, s string) (*models.User, error) {
	return nil, errDriverNotEnabledPg
}
func (r *userRepositoryPostgres) Update(ctx context.Context, user *models.User) error {
	return errDriverNotEnabledPg
}
func (r *userRepositoryPostgres) UpdatePassword(ctx context.Context, userID, hash string) error {
	return errDriverNotEnabledPg
}
func (r *userRepositoryPostgres) UpdateLoginInfo(ctx context.Context, userID string) error {
	return errDriverNotEnabledPg
}
func (r *userRepositoryPostgres) UpdateLoginMeta(ctx context.Context, userID string) error {
	return errDriverNotEnabledPg
}
func (r *userRepositoryPostgres) ExistsByUsername(ctx context.Context, username string) (bool, error) {
	return false, errDriverNotEnabledPg
}
func (r *userRepositoryPostgres) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	return false, errDriverNotEnabledPg
}
func (r *userRepositoryPostgres) Delete(ctx context.Context, id string) error {
	return errDriverNotEnabledPg
}
func (r *userRepositoryPostgres) GetList(ctx context.Context, offset, limit int) ([]*models.User, int64, error) {
	return nil, 0, errDriverNotEnabledPg
}
func (r *userRepositoryPostgres) UpdateStatus(ctx context.Context, userID string, status models.UserStatus) error {
	return errDriverNotEnabledPg
}
func (r *userRepositoryPostgres) FindUsers(ctx context.Context, opts ...options.QueryUserOption) ([]*models.User, error) {
	return nil, errDriverNotEnabledPg
}
func (r *userRepositoryPostgres) FindUser(ctx context.Context, opts ...options.QueryUserOption) (*models.User, error) {
	return nil, errDriverNotEnabledPg
}
func (r *userRepositoryPostgres) CountUsers(ctx context.Context, opts ...options.QueryUserOption) (int64, error) {
	return 0, errDriverNotEnabledPg
}
func (r *userRepositoryPostgres) FindUsersWithPagination(ctx context.Context, opts ...options.QueryUserOption) ([]*models.User, int64, error) {
	return nil, 0, errDriverNotEnabledPg
}
