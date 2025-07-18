package services

import (
	"context"
	"fmt"
	"net"
	"strings"
	"time"

	"package-base/internal/core/interfaces"
	"package-base/internal/dto"
	"package-base/internal/models"
	"package-base/internal/options"
	repoInterfaces "package-base/internal/repositories/interfaces"
	serviceInterfaces "package-base/internal/services/interfaces"

	"github.com/pkg/errors"
)

// securityService 安全服务实现
type securityService struct {
	logger           interfaces.Logger
	loginAttemptRepo repoInterfaces.LoginAttemptRepository
	redisStorage     interfaces.KeyValueStorage
	options          *serviceInterfaces.SecurityServiceOptions
}

// NewSecurityService 创建安全服务
func NewSecurityService(
	logger interfaces.Logger,
	loginAttemptRepo repoInterfaces.LoginAttemptRepository,
	redisStorage interfaces.KeyValueStorage,
	options *serviceInterfaces.SecurityServiceOptions,
) serviceInterfaces.SecurityService {
	if options == nil {
		options = getDefaultSecurityOptions()
	}

	return &securityService{
		logger:           logger,
		loginAttemptRepo: loginAttemptRepo,
		redisStorage:     redisStorage,
		options:          options,
	}
}

// getDefaultSecurityOptions 获取默认安全配置
func getDefaultSecurityOptions() *serviceInterfaces.SecurityServiceOptions {
	return &serviceInterfaces.SecurityServiceOptions{
		MaxFailedAttempts:      5,
		FailedAttemptWindow:    15 * time.Minute,
		BlockDuration:          30 * time.Minute,
		ProgressiveDelay:       true,
		MaxFailedAttemptsPerIP: 10,
		IPBlockDuration:        1 * time.Hour,
		EnableLocationCheck:    true,
		EnableDeviceCheck:      true,
		SuspiciousThreshold:    70.0,
		CleanupInterval:        1 * time.Hour,
		RetentionPeriod:        30 * 24 * time.Hour, // 30天
	}
}

// RecordLoginAttempt 记录登录尝试
func (s *securityService) RecordLoginAttempt(ctx context.Context, req *dto.RecordLoginAttemptRequest) error {
	s.logger.Info("recording login attempt",
		interfaces.String("ip_address", req.IPAddress),
		interfaces.Any("user_id", req.UserID),
		interfaces.Any("username", req.Username),
		interfaces.Any("email", req.Email),
		interfaces.String("attempt_type", req.AttemptType),
		interfaces.Bool("success", req.Success),
	)

	// 确定登录标识
	loginIdentity := s.getLoginIdentity(req)
	if loginIdentity == "" {
		return errors.New("login identity cannot be empty")
	}

	// 创建登录尝试记录
	loginAttempt := &models.LoginAttempt{
		UserID:        s.getStringValue(req.UserID),
		IPAddress:     req.IPAddress,
		UserAgent:     req.UserAgent,
		LoginIdentity: loginIdentity,
		AttemptType:   req.AttemptType,
		FailureReason: req.FailureReason,
		Country:       s.getLocationValue(req.Location, "country"),
		City:          s.getLocationValue(req.Location, "city"),
	}

	// 保存登录尝试记录
	if err := s.loginAttemptRepo.Create(ctx, loginAttempt); err != nil {
		s.logger.Error("failed to create login attempt record",
			interfaces.ErrorField(err),
			interfaces.String("ip_address", req.IPAddress),
			interfaces.String("login_identity", loginIdentity),
		)
		return errors.Wrap(err, "failed to record login attempt")
	}

	return nil
}

// CheckLoginAttemptLimit 检查登录限制
func (s *securityService) CheckLoginAttemptLimit(ctx context.Context, req *dto.CheckLoginAttemptRequest) (*dto.LoginAttemptCheckResult, error) {
	result := &dto.LoginAttemptCheckResult{
		Allowed:           true,
		Blocked:           false,
		RemainingAttempts: s.options.MaxFailedAttempts,
		RiskLevel:         "low",
	}

	// 检查IP是否被阻断
	if blocked, err := s.IsIPBlocked(ctx, net.ParseIP(req.IPAddress)); err != nil {
		return nil, errors.Wrap(err, "failed to check IP block status")
	} else if blocked {
		result.Allowed = false
		result.Blocked = true
		result.BlockReason = "IP address is blocked"
		result.RiskLevel = "critical"
		return result, nil
	}

	// 确定检查窗口
	checkWindow := req.CheckWindow
	if checkWindow == 0 {
		checkWindow = s.options.FailedAttemptWindow
	}

	// 检查用户失败尝试次数
	if req.UserID != nil || req.Username != nil || req.Email != nil {
		if err := s.checkUserAttemptLimit(ctx, req, result, checkWindow); err != nil {
			return nil, err
		}
	}

	// 检查IP失败尝试次数
	if err := s.checkIPAttemptLimit(ctx, req.IPAddress, result, checkWindow); err != nil {
		return nil, err
	}

	return result, nil
}

// checkUserAttemptLimit 检查用户尝试限制
func (s *securityService) checkUserAttemptLimit(ctx context.Context, req *dto.CheckLoginAttemptRequest, result *dto.LoginAttemptCheckResult, window time.Duration) error {
	startTime := time.Now().Add(-window)

	// 构建查询选项
	opts := []options.QueryLoginAttemptOption{
		options.WithLoginAttemptSinceTime(startTime),
		options.WithAttemptType(models.AttemptTypeFailed),
	}

	// 添加用户标识条件
	if req.UserID != nil {
		opts = append(opts, options.WithLoginAttemptUserID(*req.UserID))
	}
	if req.Username != nil {
		opts = append(opts, options.WithLoginIdentity(*req.Username))
	}
	if req.Email != nil {
		opts = append(opts, options.WithLoginIdentity(*req.Email))
	}

	// 查询失败尝试次数
	count, err := s.loginAttemptRepo.CountLoginAttempts(ctx, opts...)
	if err != nil {
		return errors.Wrap(err, "failed to count user failed attempts")
	}

	// 检查是否超过限制
	if count >= int64(s.options.MaxFailedAttempts) {
		result.Allowed = false
		result.Blocked = true
		result.BlockReason = "Too many failed login attempts"
		result.RemainingAttempts = 0
		blockUntil := time.Now().Add(s.options.BlockDuration)
		result.BlockedUntil = &blockUntil
		result.RiskLevel = "high"

		// 如果启用渐进式延迟，计算建议延迟
		if s.options.ProgressiveDelay {
			result.SuggectedDelay = s.calculateProgressiveDelay(count)
		}
	} else {
		result.RemainingAttempts = s.options.MaxFailedAttempts - int(count)
		if count > int64(s.options.MaxFailedAttempts/2) {
			result.RiskLevel = "medium"
		}
	}

	return nil
}

// checkIPAttemptLimit 检查IP尝试限制
func (s *securityService) checkIPAttemptLimit(ctx context.Context, ipAddress string, result *dto.LoginAttemptCheckResult, window time.Duration) error {
	startTime := time.Now().Add(-window)

	// 查询IP失败尝试次数
	count, err := s.loginAttemptRepo.CountLoginAttempts(ctx,
		options.WithLoginAttemptIPAddress(ipAddress),
		options.WithLoginAttemptSinceTime(startTime),
		options.WithAttemptType(models.AttemptTypeFailed),
	)
	if err != nil {
		return errors.Wrap(err, "failed to count IP failed attempts")
	}

	// 检查是否超过IP限制
	if count >= int64(s.options.MaxFailedAttemptsPerIP) {
		result.Allowed = false
		result.Blocked = true
		result.BlockReason = fmt.Sprintf("Too many failed attempts from IP %s", ipAddress)
		result.RiskLevel = "critical"

		// 自动阻断IP
		if err := s.BlockIP(ctx, net.ParseIP(ipAddress), s.options.IPBlockDuration, "Automatic block due to excessive failed attempts"); err != nil {
			s.logger.Error("failed to auto-block IP",
				interfaces.ErrorField(err),
				interfaces.String("ip_address", ipAddress),
				interfaces.Int64("count", count),
			)
		}
	}

	return nil
}

// calculateProgressiveDelay 计算渐进式延迟
func (s *securityService) calculateProgressiveDelay(attempts int64) time.Duration {
	// 基础延迟2秒，每次失败增加指数延迟
	baseDelay := 2 * time.Second
	multiplier := attempts - int64(s.options.MaxFailedAttempts/2)
	if multiplier <= 0 {
		return baseDelay
	}

	// 最大延迟1分钟
	delay := time.Duration(multiplier*multiplier) * baseDelay
	maxDelay := 1 * time.Minute
	if delay > maxDelay {
		return maxDelay
	}

	return delay
}

// GetLoginAttemptStats 获取登录尝试统计
func (s *securityService) GetLoginAttemptStats(ctx context.Context, req *dto.LoginAttemptStatsRequest) (*dto.LoginAttemptStatsResponse, error) {
	// 构建查询选项
	var opts []options.QueryLoginAttemptOption

	if req.UserID != nil {
		opts = append(opts, options.WithLoginAttemptUserID(*req.UserID))
	}
	if req.IPAddress != nil {
		opts = append(opts, options.WithLoginAttemptIPAddress(*req.IPAddress))
	}
	if req.StartTime != nil {
		opts = append(opts, options.WithLoginAttemptSinceTime(*req.StartTime))
	}
	if req.EndTime != nil {
		opts = append(opts, options.WithLoginAttemptBeforeTime(*req.EndTime))
	}
	if req.AttemptType != nil {
		opts = append(opts, options.WithAttemptType(*req.AttemptType))
	}

	// 获取基础统计 - 使用现有方法的简化版本
	attempts, err := s.loginAttemptRepo.FindLoginAttempts(ctx, opts...)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get login attempt statistics")
	}

	// 手动计算统计信息
	stats := &dto.LoginAttemptStatsResponse{
		TotalAttempts:      int64(len(attempts)),
		SuccessfulAttempts: 0,
		FailedAttempts:     0,
		UniqueIPs:          0,
		UniqueUsers:        0,
		AttemptsByType:     make(map[string]int64),
		AttemptsByHour:     make(map[string]int64),
		TopFailureReasons:  []dto.FailureReasonStats{},
		SuspiciousIPs:      []dto.SuspiciousIPStats{},
		MostActiveUsers:    []dto.UserActivityStats{},
	}

	// 计算基础统计
	ipSet := make(map[string]bool)
	userSet := make(map[string]bool)
	for _, attempt := range attempts {
		ipSet[attempt.IPAddress] = true
		if attempt.UserID != "" {
			userSet[attempt.UserID] = true
		}

		stats.AttemptsByType[attempt.AttemptType]++
		if attempt.AttemptType == models.AttemptTypeSuccess {
			stats.SuccessfulAttempts++
		} else {
			stats.FailedAttempts++
		}
	}

	stats.UniqueIPs = int64(len(ipSet))
	stats.UniqueUsers = int64(len(userSet))

	return stats, nil
}

// CleanupOldLoginAttempts 清理旧登录尝试记录
func (s *securityService) CleanupOldLoginAttempts(ctx context.Context, olderThan time.Duration) (int64, error) {
	cutoffTime := time.Now().Add(-olderThan)

	// 使用现有的清理方法
	daysToKeep := int(olderThan.Hours() / 24)
	if daysToKeep <= 0 {
		daysToKeep = 1
	}

	err := s.loginAttemptRepo.CleanupOldAttempts(ctx, daysToKeep)
	if err != nil {
		return 0, errors.Wrap(err, "failed to cleanup old login attempts")
	}

	// 暂时返回0，因为CleanupOldAttempts不返回删除计数
	deleted := int64(0)

	s.logger.Info("cleaned up old login attempts",
		interfaces.Int64("deleted_count", deleted),
		interfaces.Any("cutoff_time", cutoffTime),
	)

	return deleted, nil
}

// IsIPBlocked 检查IP是否被阻断
func (s *securityService) IsIPBlocked(ctx context.Context, ip net.IP) (bool, error) {
	key := fmt.Sprintf("blocked_ip:%s", ip.String())

	exists, err := s.redisStorage.Exists(ctx, key)
	if err != nil {
		return false, errors.Wrap(err, "failed to check IP block status")
	}

	return exists, nil
}

// BlockIP 阻断IP
func (s *securityService) BlockIP(ctx context.Context, ip net.IP, duration time.Duration, reason string) error {
	key := fmt.Sprintf("blocked_ip:%s", ip.String())
	value := fmt.Sprintf("blocked:%d:%s", time.Now().Unix(), reason)

	if err := s.redisStorage.Set(ctx, key, value, duration); err != nil {
		return errors.Wrap(err, "failed to block IP")
	}

	s.logger.Info("IP blocked",
		interfaces.String("ip_address", ip.String()),
		interfaces.Any("duration", duration),
		interfaces.String("reason", reason),
	)

	return nil
}

// UnblockIP 解除IP阻断
func (s *securityService) UnblockIP(ctx context.Context, ip net.IP) error {
	key := fmt.Sprintf("blocked_ip:%s", ip.String())

	if err := s.redisStorage.Delete(ctx, key); err != nil {
		return errors.Wrap(err, "failed to unblock IP")
	}

	s.logger.Info("IP unblocked",
		interfaces.String("ip_address", ip.String()),
	)

	return nil
}

// GetBlockedIPs 获取被阻断的IP列表
func (s *securityService) GetBlockedIPs(ctx context.Context) ([]*dto.BlockedIPInfo, error) {
	// TODO: 实现获取被阻断IP列表的功能
	// Redis 没有原生的模式匹配功能，需要考虑其他实现方式
	return []*dto.BlockedIPInfo{}, nil
}

// IsUserTemporarilyBlocked 检查用户是否被临时阻断
func (s *securityService) IsUserTemporarilyBlocked(ctx context.Context, userID string) (bool, error) {
	key := fmt.Sprintf("blocked_user:%s", userID)

	exists, err := s.redisStorage.Exists(ctx, key)
	if err != nil {
		return false, errors.Wrap(err, "failed to check user block status")
	}

	return exists, nil
}

// BlockUserTemporarily 临时阻断用户
func (s *securityService) BlockUserTemporarily(ctx context.Context, userID string, duration time.Duration, reason string) error {
	key := fmt.Sprintf("blocked_user:%s", userID)
	value := fmt.Sprintf("blocked:%d:%s", time.Now().Unix(), reason)

	if err := s.redisStorage.Set(ctx, key, value, duration); err != nil {
		return errors.Wrap(err, "failed to block user")
	}

	s.logger.Info("user temporarily blocked",
		interfaces.String("user_id", userID),
		interfaces.Any("duration", duration),
		interfaces.String("reason", reason),
	)

	return nil
}

// UnblockUser 解除用户阻断
func (s *securityService) UnblockUser(ctx context.Context, userID string) error {
	key := fmt.Sprintf("blocked_user:%s", userID)

	if err := s.redisStorage.Delete(ctx, key); err != nil {
		return errors.Wrap(err, "failed to unblock user")
	}

	s.logger.Info("user unblocked",
		interfaces.String("user_id", userID),
	)

	return nil
}

// AnalyzeLoginRisk 分析登录风险
func (s *securityService) AnalyzeLoginRisk(ctx context.Context, req *dto.LoginRiskAnalysisRequest) (*dto.LoginRiskAnalysisResponse, error) {
	// TODO: 实现详细的风险分析逻辑
	response := &dto.LoginRiskAnalysisResponse{
		RiskScore:       10.0,
		RiskLevel:       "low",
		RiskFactors:     []dto.RiskFactor{},
		Recommendations: []string{},
		RequiresAction:  false,
	}

	return response, nil
}

// GetSuspiciousActivities 获取可疑活动
func (s *securityService) GetSuspiciousActivities(ctx context.Context, req *dto.SuspiciousActivityRequest) (*dto.SuspiciousActivityResponse, error) {
	// TODO: 实现可疑活动检测逻辑
	response := &dto.SuspiciousActivityResponse{
		Activities: []dto.SuspiciousActivity{},
		Total:      0,
		RiskSummary: dto.RiskSummary{
			TotalActivities:    0,
			RiskDistribution:   make(map[string]int64),
			TopRiskFactors:     []string{},
			MostActiveIPs:      []string{},
			RecommendedActions: []string{},
		},
	}

	return response, nil
}

// GetSecurityConfig 获取安全配置
func (s *securityService) GetSecurityConfig(ctx context.Context) (*dto.SecurityConfigResponse, error) {
	return &dto.SecurityConfigResponse{
		LoginAttemptConfig: dto.LoginAttemptConfig{
			MaxFailedAttempts:   s.options.MaxFailedAttempts,
			FailedAttemptWindow: s.options.FailedAttemptWindow,
			BlockDuration:       s.options.BlockDuration,
			ProgressiveDelay:    s.options.ProgressiveDelay,
			Enabled:             true,
		},
		IPBlockingConfig: dto.IPBlockingConfig{
			MaxFailedAttemptsPerIP: s.options.MaxFailedAttemptsPerIP,
			IPBlockDuration:        s.options.IPBlockDuration,
			Enabled:                true,
			Whitelist:              []string{},
			Blacklist:              []string{},
		},
		RiskAnalysisConfig: dto.RiskAnalysisConfig{
			EnableLocationCheck:     s.options.EnableLocationCheck,
			EnableDeviceCheck:       s.options.EnableDeviceCheck,
			SuspiciousThreshold:     s.options.SuspiciousThreshold,
			Enabled:                 true,
			LocationChangeThreshold: 100.0, // 100km
			DeviceChangeThreshold:   0.8,   // 80% similarity
		},
		CleanupConfig: dto.CleanupConfig{
			CleanupInterval: s.options.CleanupInterval,
			RetentionPeriod: s.options.RetentionPeriod,
			AutoCleanup:     true,
			BatchSize:       1000,
		},
	}, nil
}

// UpdateSecurityConfig 更新安全配置
func (s *securityService) UpdateSecurityConfig(ctx context.Context, req *dto.UpdateSecurityConfigRequest) error {
	// TODO: 实现安全配置更新逻辑
	s.logger.Info("security config update requested",
		interfaces.Any("config", req),
	)

	return nil
}

// 辅助函数

// getLoginIdentity 获取登录标识
func (s *securityService) getLoginIdentity(req *dto.RecordLoginAttemptRequest) string {
	if req.Username != nil && *req.Username != "" {
		return *req.Username
	}
	if req.Email != nil && *req.Email != "" {
		return *req.Email
	}
	if req.UserID != nil && *req.UserID != "" {
		return *req.UserID
	}
	return ""
}

// getStringValue 安全获取字符串指针的值
func (s *securityService) getStringValue(ptr *string) string {
	if ptr == nil {
		return ""
	}
	return *ptr
}

// getLocationValue 获取位置信息的值
func (s *securityService) getLocationValue(location *dto.LoginLocation, field string) string {
	if location == nil {
		return ""
	}

	switch strings.ToLower(field) {
	case "country":
		return location.Country
	case "city":
		return location.City
	case "region":
		return location.Region
	default:
		return ""
	}
}
