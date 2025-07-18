package interfaces

import (
	"context"
	"net"
	"time"

	"package-base/internal/dto"
)

// SecurityService 安全服务接口
type SecurityService interface {
	// 登录限制相关方法
	RecordLoginAttempt(ctx context.Context, req *dto.RecordLoginAttemptRequest) error
	CheckLoginAttemptLimit(ctx context.Context, req *dto.CheckLoginAttemptRequest) (*dto.LoginAttemptCheckResult, error)
	GetLoginAttemptStats(ctx context.Context, req *dto.LoginAttemptStatsRequest) (*dto.LoginAttemptStatsResponse, error)
	CleanupOldLoginAttempts(ctx context.Context, olderThan time.Duration) (int64, error)

	// IP 地址管理
	IsIPBlocked(ctx context.Context, ip net.IP) (bool, error)
	BlockIP(ctx context.Context, ip net.IP, duration time.Duration, reason string) error
	UnblockIP(ctx context.Context, ip net.IP) error
	GetBlockedIPs(ctx context.Context) ([]*dto.BlockedIPInfo, error)

	// 用户安全状态管理
	IsUserTemporarilyBlocked(ctx context.Context, userID string) (bool, error)
	BlockUserTemporarily(ctx context.Context, userID string, duration time.Duration, reason string) error
	UnblockUser(ctx context.Context, userID string) error

	// 风险分析
	AnalyzeLoginRisk(ctx context.Context, req *dto.LoginRiskAnalysisRequest) (*dto.LoginRiskAnalysisResponse, error)
	GetSuspiciousActivities(ctx context.Context, req *dto.SuspiciousActivityRequest) (*dto.SuspiciousActivityResponse, error)

	// 配置管理
	GetSecurityConfig(ctx context.Context) (*dto.SecurityConfigResponse, error)
	UpdateSecurityConfig(ctx context.Context, req *dto.UpdateSecurityConfigRequest) error
}

// SecurityServiceOptions 安全服务配置选项
type SecurityServiceOptions struct {
	// 登录限制配置
	MaxFailedAttempts   int           // 最大失败尝试次数
	FailedAttemptWindow time.Duration // 失败尝试时间窗口
	BlockDuration       time.Duration // 阻断持续时间
	ProgressiveDelay    bool          // 是否启用渐进式延迟

	// IP 限制配置
	MaxFailedAttemptsPerIP int           // 每个IP最大失败次数
	IPBlockDuration        time.Duration // IP阻断持续时间

	// 风险分析配置
	EnableLocationCheck bool    // 是否启用地理位置检查
	EnableDeviceCheck   bool    // 是否启用设备指纹检查
	SuspiciousThreshold float64 // 可疑活动阈值

	// 清理配置
	CleanupInterval time.Duration // 清理间隔
	RetentionPeriod time.Duration // 数据保留期
}
