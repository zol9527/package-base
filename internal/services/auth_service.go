package services

import (
	"context"
	"time"

	"package-base/internal/core/interfaces"
	"package-base/internal/core/utils"
	"package-base/internal/dto"
	"package-base/internal/models"
	"package-base/internal/options"
	repoInterfaces "package-base/internal/repositories/interfaces"
	serviceInterfaces "package-base/internal/services/interfaces"

	"github.com/pkg/errors"
	"gorm.io/gorm"
)

// authService 认证服务实现
type authService struct {
	userRepo           repoInterfaces.UserRepository
	sessionRepo        repoInterfaces.SessionRepository
	tokenBlacklistRepo repoInterfaces.TokenBlacklistRepository
	jwtManager         *utils.JWTManager
	validator          *utils.Validator
	logger             interfaces.Logger
}

// NewAuthService 创建认证服务实例
// Function: NewAuthService
//
// Description: 传统构造函数（向后兼容），内部自建 JWTManager。推荐在容器中使用 NewAuthServiceWithManager 以便动态调整 token 时长。
//
// Parameters:
//
//	userRepo(UserRepository): 用户仓库
//	sessionRepo(SessionRepository): 会话仓库
//	tokenBlacklistRepo(TokenBlacklistRepository): 黑名单仓库
//	jwtSecret(string): 签名密钥
//	accessTokenDuration(time.Duration): access token 时长
//	refreshTokenDuration(time.Duration): refresh token 时长
//	logger(interfaces.Logger): 日志
//
// Returns:
//
//	svc(serviceInterfaces.AuthService): 认证服务实例
func NewAuthService(
	userRepo repoInterfaces.UserRepository,
	sessionRepo repoInterfaces.SessionRepository,
	tokenBlacklistRepo repoInterfaces.TokenBlacklistRepository,
	jwtSecret string,
	accessTokenDuration time.Duration,
	refreshTokenDuration time.Duration,
	logger interfaces.Logger,
) serviceInterfaces.AuthService {
	jwtManager := utils.NewJWTManager(jwtSecret, accessTokenDuration, refreshTokenDuration)
	validator := utils.NewValidator()
	return &authService{
		userRepo:           userRepo,
		sessionRepo:        sessionRepo,
		tokenBlacklistRepo: tokenBlacklistRepo,
		jwtManager:         jwtManager,
		validator:          validator,
		logger:             logger,
	}
}

// Function: NewAuthServiceWithManager
//
// Description: 使用已经构造好的 *JWTManager 实例（便于测试动态修改 token 生命周期）。
//
// Parameters:
//
//	userRepo(UserRepository): 用户仓库
//	sessionRepo(SessionRepository): 会话仓库
//	tokenBlacklistRepo(TokenBlacklistRepository): 黑名单仓库
//	jwtManager(*utils.JWTManager): 预先创建的 JWT 管理器
//	logger(interfaces.Logger): 日志组件
//
// Returns:
//
//	svc(serviceInterfaces.AuthService): 认证服务实例
func NewAuthServiceWithManager(
	userRepo repoInterfaces.UserRepository,
	sessionRepo repoInterfaces.SessionRepository,
	tokenBlacklistRepo repoInterfaces.TokenBlacklistRepository,
	jwtManager *utils.JWTManager,
	logger interfaces.Logger,
) serviceInterfaces.AuthService {
	validator := utils.NewValidator()
	return &authService{
		userRepo:           userRepo,
		sessionRepo:        sessionRepo,
		tokenBlacklistRepo: tokenBlacklistRepo,
		jwtManager:         jwtManager,
		validator:          validator,
		logger:             logger,
	}
}

// Register 用户注册
func (s *authService) Register(ctx context.Context, req *dto.RegisterRequest) (*dto.AuthResponse, error) {
	// 1. 验证输入参数
	if err := s.validator.ValidateStruct(req); err != nil {
		s.logger.Warn("registration validation failed",
			interfaces.ErrorField(err),
			interfaces.String("username", req.Username),
			interfaces.String("email", req.Email),
		)
		return nil, errors.Wrap(err, "validation failed")
	}

	// 2. 检查用户名是否已存在
	existingUser, err := s.userRepo.FindUser(ctx,
		options.WithUsername(req.Username),
	)
	if err != nil {
		s.logger.Error("failed to check username existence",
			interfaces.ErrorField(err),
			interfaces.String("username", req.Username),
		)
		return nil, errors.Wrap(err, "failed to check username")
	}
	if existingUser != nil {
		return nil, errors.New("username already exists")
	}

	// 3. 检查邮箱是否已存在
	existingUser, err = s.userRepo.FindUser(ctx,
		options.WithEmail(req.Email),
	)
	if err != nil {
		s.logger.Error("failed to check email existence",
			interfaces.ErrorField(err),
			interfaces.String("email", req.Email),
		)
		return nil, errors.Wrap(err, "failed to check email")
	}
	if existingUser != nil {
		return nil, errors.New("email already exists")
	}

	// 4. 加密密码
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		s.logger.Error("failed to hash password", interfaces.ErrorField(err))
		return nil, errors.Wrap(err, "failed to hash password")
	}

	// 5. 创建用户
	user := &models.User{
		Username:     req.Username,
		Email:        &req.Email,
		PasswordHash: &hashedPassword,
		Phone:        &req.Phone,
		Nickname:     &req.Nickname,
		Status:       models.UserStatusActive, // 默认激活状态
		Roles:        models.UserRoleTypeUser, // 默认普通用户
	}

	// 生成UUID
	user.ID = utils.GenerateUUID()
	user.Uuid = utils.GenerateShortUUID()

	if err := s.userRepo.Create(ctx, user); err != nil {
		s.logger.Error("failed to create user",
			interfaces.ErrorField(err),
			interfaces.String("username", req.Username),
		)
		return nil, errors.Wrap(err, "failed to create user")
	}

	// 6. 生成Token并创建会话
	accessToken, expiresAt, err := s.jwtManager.GenerateAccessToken(
		user.ID, user.Username, user.GetEmail(), int(user.Roles))
	if err != nil {
		s.logger.Error("failed to generate access token",
			interfaces.ErrorField(err),
			interfaces.String("user_id", user.ID),
		)
		return nil, errors.Wrap(err, "failed to generate access token")
	}

	refreshToken, _, err := s.jwtManager.GenerateRefreshToken(user.ID)
	if err != nil {
		s.logger.Error("failed to generate refresh token",
			interfaces.ErrorField(err),
			interfaces.String("user_id", user.ID),
		)
		return nil, errors.Wrap(err, "failed to generate refresh token")
	}

	// 7. 创建会话记录
	session := &models.Session{
		UserID:       user.ID,
		RefreshToken: refreshToken,
		AccessToken:  accessToken,
		ExpiresAt:    expiresAt.Add(24 * time.Hour), // 会话24小时过期
		IsActive:     true,
	}

	if err := s.sessionRepo.Create(ctx, session); err != nil {
		s.logger.Error("failed to create session",
			interfaces.ErrorField(err),
			interfaces.String("user_id", user.ID),
		)
		return nil, errors.Wrap(err, "failed to create session")
	}

	s.logger.Info("user registered successfully",
		interfaces.String("user_id", user.ID),
		interfaces.String("username", user.Username),
	)

	// 8. 返回认证响应
	userInfo := s.convertToUserInfo(user)
	return &dto.AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    expiresAt,
		ExpiresIn:    int64(time.Until(expiresAt).Seconds()),
		TokenType:    "Bearer",
		User:         *userInfo,
	}, nil
}

// Login 用户登录
func (s *authService) Login(ctx context.Context, req *dto.LoginRequest) (*dto.AuthResponse, error) {
	// 1. 验证输入参数
	if err := s.validator.ValidateStruct(req); err != nil {
		s.logger.Warn("login validation failed",
			interfaces.ErrorField(err),
			interfaces.String("username", req.Username),
		)
		return nil, errors.Wrap(err, "validation failed")
	}

	// 2. 根据用户名或邮箱查找用户
	var user *models.User
	var err error

	// 检查是否为邮箱格式
	if utils.ValidateEmail(req.Username) == nil {
		user, err = s.userRepo.FindUser(ctx, options.WithEmail(req.Username))
	} else {
		user, err = s.userRepo.FindUser(ctx, options.WithUsername(req.Username))
	}

	if err != nil {
		s.logger.Error("failed to get user",
			interfaces.ErrorField(err),
			interfaces.String("username", req.Username),
		)
		return nil, errors.Wrap(err, "failed to get user")
	}

	if user == nil {
		s.logger.Warn("login failed: user not found",
			interfaces.String("username", req.Username),
		)
		return nil, errors.New("invalid credentials")
	}

	// 3. 验证密码
	if err := utils.VerifyPassword(user.GetPasswordHash(), req.Password); err != nil {
		s.logger.Warn("login failed: invalid password",
			interfaces.String("user_id", user.ID),
			interfaces.String("username", req.Username),
		)
		return nil, errors.New("invalid credentials")
	}

	// 4. 检查用户状态
	if user.Status != models.UserStatusActive {
		s.logger.Warn("login failed: user not active",
			interfaces.String("user_id", user.ID),
			interfaces.Int("status", int(user.Status)),
		)
		return nil, errors.New("account is not active")
	}

	// 5. 生成新的Token
	accessToken, expiresAt, err := s.jwtManager.GenerateAccessToken(
		user.ID, user.Username, user.GetEmail(), int(user.Roles))
	if err != nil {
		s.logger.Error("failed to generate access token",
			interfaces.ErrorField(err),
			interfaces.String("user_id", user.ID),
		)
		return nil, errors.Wrap(err, "failed to generate access token")
	}

	refreshToken, _, err := s.jwtManager.GenerateRefreshToken(user.ID)
	if err != nil {
		s.logger.Error("failed to generate refresh token",
			interfaces.ErrorField(err),
			interfaces.String("user_id", user.ID),
		)
		return nil, errors.Wrap(err, "failed to generate refresh token")
	}

	// 6. 创建新的会话记录（同步执行，保证后续依赖的一致性）
	session := &models.Session{
		UserID:       user.ID,
		RefreshToken: refreshToken,
		AccessToken:  accessToken,
		ExpiresAt:    expiresAt.Add(24 * time.Hour), // 会话24小时过期
		IsActive:     true,
	}

	if err := s.sessionRepo.Create(ctx, session); err != nil {
		s.logger.Error("failed to create session",
			interfaces.ErrorField(err),
			interfaces.String("user_id", user.ID),
		)
		return nil, errors.Wrap(err, "failed to create session")
	}

	// 7. 更新用户登录统计
	user.LoginCount = user.LoginCount + 1
	user.LastLoginAt = &time.Time{}
	*user.LastLoginAt = time.Now()

	if err := s.userRepo.Update(ctx, user); err != nil {
		s.logger.Warn("failed to update user login stats",
			interfaces.ErrorField(err),
			interfaces.String("user_id", user.ID),
		)
		// 不影响登录流程，只记录警告
	}

	s.logger.Info("user logged in successfully",
		interfaces.String("user_id", user.ID),
		interfaces.String("username", user.Username),
	)

	// 8. 返回认证响应
	userInfo := s.convertToUserInfo(user)
	return &dto.AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    expiresAt,
		ExpiresIn:    int64(time.Until(expiresAt).Seconds()),
		TokenType:    "Bearer",
		User:         *userInfo,
	}, nil
}

// SignIn 语义别名：调用 Login
//
// Function: SignIn
//
// Description: 语义别名封装，直接复用 Login 实现，保持未来扩展差异化的灵活性。
//
// Parameters:
//
//	ctx(context.Context): 上下文
//	req(*dto.LoginRequest): 登录请求
//
// Returns:
//
//	resp(*dto.AuthResponse): 登录响应
//	err(error): 错误
func (s *authService) SignIn(ctx context.Context, req *dto.LoginRequest) (*dto.AuthResponse, error) {
	return s.Login(ctx, req)
}

// RefreshToken 刷新Token
func (s *authService) RefreshToken(ctx context.Context, req *dto.RefreshTokenRequest) (*dto.TokenResponse, error) {
	// 1. 验证输入参数
	if err := s.validator.ValidateStruct(req); err != nil {
		return nil, errors.Wrap(err, "validation failed")
	}

	// 2. 验证Refresh Token
	userID, err := s.jwtManager.ValidateRefreshToken(req.RefreshToken)
	if err != nil {
		s.logger.Warn("invalid refresh token", interfaces.ErrorField(err))
		return nil, errors.New("invalid refresh token")
	}

	// 3. 根据会话ID查找会话
	session, err := s.sessionRepo.GetByRefreshToken(ctx, req.RefreshToken)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get session")
	}
	if session == nil {
		s.logger.Warn("session not found for refresh token")
		return nil, errors.New("invalid refresh token")
	}

	// 4. 检查会话状态
	if !session.IsActive || session.IsExpired() {
		s.logger.Warn("session is inactive or expired",
			interfaces.String("session_id", session.ID),
			interfaces.String("user_id", session.UserID),
		)
		return nil, errors.New("session expired")
	}

	// 5. 获取用户信息
	user, err := s.userRepo.FindUser(ctx, options.WithID(userID))
	if err != nil {
		return nil, errors.Wrap(err, "failed to get user")
	}

	if user == nil {
		return nil, errors.New("user not found")
	}

	// 6. 生成新的Token
	accessToken, expiresAt, err := s.jwtManager.GenerateAccessToken(
		user.ID, user.Username, user.GetEmail(), int(user.Roles))
	if err != nil {
		return nil, errors.Wrap(err, "failed to generate access token")
	}

	newRefreshToken, _, err := s.jwtManager.GenerateRefreshToken(user.ID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to generate refresh token")
	}

	// 7. CAS 原子更新会话（避免并发多次刷新同时成功）
	ok, err := s.sessionRepo.RotateTokensCAS(ctx, req.RefreshToken, newRefreshToken, accessToken, expiresAt.Unix())
	if err != nil {
		return nil, errors.Wrap(err, "failed to rotate tokens CAS")
	}
	if !ok { // 说明并发中已被其他刷新请求更新或状态失效
		return nil, errors.New("invalid refresh token")
	}

	s.logger.Info("token refreshed successfully",
		interfaces.String("user_id", user.ID),
		interfaces.String("session_id", session.ID),
	)

	return &dto.TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
		ExpiresAt:    expiresAt,
		TokenType:    "Bearer",
	}, nil
}

// Logout 用户登出（单设备）
func (s *authService) Logout(ctx context.Context, userID, sessionID string) error {
	// 获取会话信息以获取access token
	session, err := s.sessionRepo.GetByID(ctx, sessionID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("session not found")
		}
		return errors.Wrap(err, "failed to get session")
	}

	// 验证会话属于该用户
	if session.UserID != userID {
		return errors.New("session does not belong to user")
	}

	// 将access token加入黑名单
	if session.AccessToken != "" {
		if err := s.BlacklistToken(ctx, session.AccessToken); err != nil {
			s.logger.Warn("failed to blacklist access token during logout",
				interfaces.ErrorField(err),
				interfaces.String("session_id", sessionID),
			)
			// 不中断登出流程，只记录警告
		}
	}

	// 停用指定会话
	if err := s.sessionRepo.DeactivateSession(ctx, sessionID); err != nil {
		return errors.Wrap(err, "failed to deactivate session")
	}

	s.logger.Info("user logged out successfully",
		interfaces.String("user_id", userID),
		interfaces.String("session_id", sessionID),
	)

	return nil
}

// LogoutAll 用户登出（所有设备）
func (s *authService) LogoutAll(ctx context.Context, userID string) error {
	// 获取用户的所有活跃会话以将access token加入黑名单
	activeSessions, err := s.sessionRepo.GetActiveByUserID(ctx, userID)
	if err != nil {
		s.logger.Warn("failed to get active sessions for blacklisting tokens",
			interfaces.ErrorField(err),
			interfaces.String("user_id", userID),
		)
		// 继续执行登出流程
	} else {
		// 将所有活跃会话的access token加入黑名单
		for _, session := range activeSessions {
			if session.AccessToken != "" {
				if err := s.BlacklistToken(ctx, session.AccessToken); err != nil {
					s.logger.Warn("failed to blacklist access token during logout all",
						interfaces.ErrorField(err),
						interfaces.String("session_id", session.ID),
					)
					// 不中断登出流程，只记录警告
				}
			}
		}
	}

	// 停用用户的所有活跃会话
	if err := s.sessionRepo.DeactivateUserSessions(ctx, userID); err != nil {
		return errors.Wrap(err, "failed to deactivate all sessions")
	}

	s.logger.Info("user logged out from all devices",
		interfaces.String("user_id", userID),
	)

	return nil
}

// ValidateToken 验证Token
func (s *authService) ValidateToken(ctx context.Context, token string) (*dto.UserInfo, error) {
	// 首先检查token是否在黑名单中
	isBlacklisted, err := s.IsTokenBlacklisted(ctx, token)
	if err != nil {
		s.logger.Warn("failed to check token blacklist status, allowing token",
			interfaces.ErrorField(err),
		)
		// 如果检查黑名单失败，继续验证token（宽松策略）
	} else if isBlacklisted {
		return nil, errors.New("token has been revoked")
	}

	// 验证Token
	claims, err := s.jwtManager.ValidateAccessToken(token)
	if err != nil {
		return nil, errors.Wrap(err, "invalid token")
	}

	// 获取用户信息
	user, err := s.userRepo.FindUser(ctx, options.WithID(claims.UserID))
	if err != nil {
		return nil, errors.Wrap(err, "failed to get user")
	}

	if user == nil {
		return nil, errors.New("user not found")
	}

	// 检查用户状态
	if user.Status != models.UserStatusActive {
		return nil, errors.New("user account is not active")
	}

	return s.convertToUserInfo(user), nil
}

// ChangePassword 修改密码
func (s *authService) ChangePassword(ctx context.Context, userID string, req *dto.ChangePasswordRequest) error {
	// 1. 验证输入参数
	if err := s.validator.ValidateStruct(req); err != nil {
		return errors.Wrap(err, "validation failed")
	}

	// 2. 获取用户信息
	user, err := s.userRepo.FindUser(ctx, options.WithID(userID))
	if err != nil {
		return errors.Wrap(err, "failed to get user")
	}

	if user == nil {
		return errors.New("user not found")
	}

	// 3. 验证旧密码
	if err := utils.VerifyPassword(user.GetPasswordHash(), req.OldPassword); err != nil {
		s.logger.Warn("change password failed: invalid old password",
			interfaces.String("user_id", userID),
		)
		return errors.New("invalid old password")
	}

	// 4. 加密新密码
	hashedPassword, err := utils.HashPassword(req.NewPassword)
	if err != nil {
		return errors.Wrap(err, "failed to hash new password")
	}

	// 5. 更新密码
	user.PasswordHash = &hashedPassword
	if err := s.userRepo.Update(ctx, user); err != nil {
		return errors.Wrap(err, "failed to update password")
	}

	// 6. 停用所有现有会话（强制重新登录）
	if err := s.sessionRepo.DeactivateUserSessions(ctx, userID); err != nil {
		s.logger.Warn("failed to deactivate sessions after password change",
			interfaces.ErrorField(err),
			interfaces.String("user_id", userID),
		)
	}

	s.logger.Info("password changed successfully",
		interfaces.String("user_id", userID),
	)

	return nil
}

// GetUserSessions 获取用户会话列表
func (s *authService) GetUserSessions(ctx context.Context, userID string) (*dto.SessionListResponse, error) {
	sessions, err := s.sessionRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get user sessions")
	}

	sessionInfos := make([]dto.SessionInfo, 0, len(sessions))
	for _, session := range sessions {
		sessionInfos = append(sessionInfos, dto.SessionInfo{
			ID:         session.ID,
			DeviceInfo: session.DeviceInfo,
			IPAddress:  session.IPAddress,
			UserAgent:  session.UserAgent,
			IsActive:   session.IsActive,
			CreatedAt:  session.CreatedAt,
			ExpiresAt:  session.ExpiresAt,
			IsCurrent:  false, // 需要在Handler层设置
		})
	}

	return &dto.SessionListResponse{
		Sessions: sessionInfos,
		Total:    len(sessionInfos),
	}, nil
}

// DeleteSession 删除指定会话
func (s *authService) DeleteSession(ctx context.Context, userID, sessionID string) error {
	// 验证会话属于指定用户
	session, err := s.sessionRepo.GetByID(ctx, sessionID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("session not found")
		}
		return errors.Wrap(err, "failed to get session")
	}

	if session.UserID != userID {
		return errors.New("session does not belong to user")
	}

	// 停用会话
	if err := s.sessionRepo.DeactivateSession(ctx, sessionID); err != nil {
		return errors.Wrap(err, "failed to delete session")
	}

	s.logger.Info("session deleted successfully",
		interfaces.String("user_id", userID),
		interfaces.String("session_id", sessionID),
	)

	return nil
}

// GetCurrentUserInfo 获取当前用户信息
func (s *authService) GetCurrentUserInfo(ctx context.Context, userID string) (*dto.UserInfo, error) {
	user, err := s.userRepo.FindUser(ctx, options.WithID(userID))
	if err != nil {
		return nil, errors.Wrap(err, "failed to get user")
	}

	if user == nil {
		return nil, errors.New("user not found")
	}

	return s.convertToUserInfo(user), nil
}

// convertToUserInfo 转换用户模型为DTO
func (s *authService) convertToUserInfo(user *models.User) *dto.UserInfo {
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

// GetActiveSessions 获取用户活跃会话
func (s *authService) GetActiveSessions(ctx context.Context, userID string) (*dto.SessionListResponse, error) {
	sessions, err := s.sessionRepo.FindSessions(ctx,
		options.WithUserActiveSessions(userID),
		options.WithSessionOrderByCreatedAt(true),
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get active sessions")
	}

	sessionInfos := make([]dto.SessionInfo, 0, len(sessions))
	for _, session := range sessions {
		sessionInfos = append(sessionInfos, dto.SessionInfo{
			ID:         session.ID,
			DeviceInfo: session.DeviceInfo,
			IPAddress:  session.IPAddress,
			UserAgent:  session.UserAgent,
			IsActive:   session.IsActive,
			CreatedAt:  session.CreatedAt,
			ExpiresAt:  session.ExpiresAt,
			IsCurrent:  false, // 需要在Handler层设置
		})
	}

	return &dto.SessionListResponse{
		Sessions: sessionInfos,
		Total:    len(sessionInfos),
	}, nil
}

// GetSessionsByDevice 根据设备信息获取会话
func (s *authService) GetSessionsByDevice(ctx context.Context, userID, deviceInfo string) (*dto.SessionListResponse, error) {
	sessions, err := s.sessionRepo.FindSessions(ctx,
		options.WithSessionsByUserAndDevice(userID, deviceInfo),
		options.WithSessionOrderByCreatedAt(true),
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get sessions by device")
	}

	sessionInfos := make([]dto.SessionInfo, 0, len(sessions))
	for _, session := range sessions {
		sessionInfos = append(sessionInfos, dto.SessionInfo{
			ID:         session.ID,
			DeviceInfo: session.DeviceInfo,
			IPAddress:  session.IPAddress,
			UserAgent:  session.UserAgent,
			IsActive:   session.IsActive,
			CreatedAt:  session.CreatedAt,
			ExpiresAt:  session.ExpiresAt,
			IsCurrent:  false,
		})
	}

	return &dto.SessionListResponse{
		Sessions: sessionInfos,
		Total:    len(sessionInfos),
	}, nil
}

// GetRecentSessions 获取最近的会话（指定天数内）
func (s *authService) GetRecentSessions(ctx context.Context, userID string, days int) (*dto.SessionListResponse, error) {
	sessions, err := s.sessionRepo.FindSessions(ctx,
		options.WithUserID(userID),
		options.WithRecentSessions(days),
		options.WithSessionOrderByCreatedAt(true),
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get recent sessions")
	}

	sessionInfos := make([]dto.SessionInfo, 0, len(sessions))
	for _, session := range sessions {
		sessionInfos = append(sessionInfos, dto.SessionInfo{
			ID:         session.ID,
			DeviceInfo: session.DeviceInfo,
			IPAddress:  session.IPAddress,
			UserAgent:  session.UserAgent,
			IsActive:   session.IsActive,
			CreatedAt:  session.CreatedAt,
			ExpiresAt:  session.ExpiresAt,
			IsCurrent:  false,
		})
	}

	return &dto.SessionListResponse{
		Sessions: sessionInfos,
		Total:    len(sessionInfos),
	}, nil
}

// CountUserSessions 统计用户会话数量
func (s *authService) CountUserSessions(ctx context.Context, userID string, activeOnly bool) (int64, error) {
	var opts []options.QuerySessionOption

	if activeOnly {
		opts = append(opts, options.WithUserActiveSessions(userID))
	} else {
		opts = append(opts, options.WithUserID(userID))
	}

	count, err := s.sessionRepo.CountSessions(ctx, opts...)
	if err != nil {
		return 0, errors.Wrap(err, "failed to count user sessions")
	}

	return count, nil
}

// Token黑名单管理方法

// BlacklistToken 将token添加到黑名单
func (s *authService) BlacklistToken(ctx context.Context, token string) error {
	// 解析token以获取过期时间
	claims, err := s.jwtManager.ValidateAccessToken(token)
	if err != nil {
		s.logger.Warn("failed to parse token for blacklisting, using default expiration",
			interfaces.ErrorField(err),
		)
		// 如果解析失败，使用默认过期时间
		return s.tokenBlacklistRepo.AddToBlacklist(ctx, token, 24*time.Hour)
	}

	// 计算token的剩余过期时间
	now := time.Now()
	expiration := time.Unix(claims.ExpiresAt.Unix(), 0)
	if expiration.Before(now) {
		// token已经过期，无需加入黑名单
		s.logger.Info("token already expired, not adding to blacklist")
		return nil
	}

	ttl := expiration.Sub(now)
	if err := s.tokenBlacklistRepo.AddToBlacklist(ctx, token, ttl); err != nil {
		s.logger.Error("failed to add token to blacklist",
			interfaces.ErrorField(err),
		)
		return errors.Wrap(err, "failed to add token to blacklist")
	}

	s.logger.Info("token successfully added to blacklist",
		interfaces.String("ttl", ttl.String()),
	)

	return nil
}

// IsTokenBlacklisted 检查token是否在黑名单中
func (s *authService) IsTokenBlacklisted(ctx context.Context, token string) (bool, error) {
	isBlacklisted, err := s.tokenBlacklistRepo.IsBlacklisted(ctx, token)
	if err != nil {
		s.logger.Error("failed to check token blacklist status",
			interfaces.ErrorField(err),
		)
		return false, errors.Wrap(err, "failed to check token blacklist status")
	}

	return isBlacklisted, nil
}

// ClearExpiredBlacklistTokens 清理过期的黑名单token
func (s *authService) ClearExpiredBlacklistTokens(ctx context.Context) error {
	if err := s.tokenBlacklistRepo.ClearExpiredTokens(ctx); err != nil {
		s.logger.Error("failed to clear expired blacklist tokens",
			interfaces.ErrorField(err),
		)
		return errors.Wrap(err, "failed to clear expired blacklist tokens")
	}

	s.logger.Info("expired blacklist tokens cleared successfully")
	return nil
}

// GetBlacklistStats 获取黑名单统计信息
func (s *authService) GetBlacklistStats(ctx context.Context) (map[string]interface{}, error) {
	size, err := s.tokenBlacklistRepo.GetBlacklistSize(ctx)
	if err != nil {
		s.logger.Error("failed to get blacklist size",
			interfaces.ErrorField(err),
		)
		return nil, errors.Wrap(err, "failed to get blacklist size")
	}

	stats := map[string]interface{}{
		"total_blacklisted_tokens": size,
		"last_updated":             time.Now(),
	}

	return stats, nil
}
