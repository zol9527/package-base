package dto

import (
	"time"
)

// ============================================================================
// 辅助结构体
// ============================================================================

// LoginLocation 登录位置信息
type LoginLocation struct {
	Country   string  `json:"country,omitempty"`
	Region    string  `json:"region,omitempty"`
	City      string  `json:"city,omitempty"`
	Latitude  float64 `json:"latitude,omitempty"`
	Longitude float64 `json:"longitude,omitempty"`
	Timezone  string  `json:"timezone,omitempty"`
}

// DeviceInfo 设备信息
type DeviceInfo struct {
	DeviceID       string `json:"device_id,omitempty"`
	DeviceType     string `json:"device_type,omitempty"` // desktop, mobile, tablet
	OS             string `json:"os,omitempty"`
	OSVersion      string `json:"os_version,omitempty"`
	Browser        string `json:"browser,omitempty"`
	BrowserVersion string `json:"browser_version,omitempty"`
	ScreenWidth    int    `json:"screen_width,omitempty"`
	ScreenHeight   int    `json:"screen_height,omitempty"`
	Fingerprint    string `json:"fingerprint,omitempty"`
}

// ============================================================================
// 登录限制相关DTO
// ============================================================================

// RecordLoginAttemptRequest 记录登录尝试请求
type RecordLoginAttemptRequest struct {
	UserID        *string                `json:"user_id,omitempty"`
	Username      *string                `json:"username,omitempty"`
	Email         *string                `json:"email,omitempty"`
	IPAddress     string                 `json:"ip_address" validate:"required,ip"`
	UserAgent     string                 `json:"user_agent"`
	AttemptType   string                 `json:"attempt_type" validate:"required"`
	Success       bool                   `json:"success"`
	FailureReason string                 `json:"failure_reason,omitempty"`
	Location      *LoginLocation         `json:"location,omitempty"`
	DeviceInfo    *DeviceInfo            `json:"device_info,omitempty"`
	SessionID     *string                `json:"session_id,omitempty"`
	ExtraData     map[string]interface{} `json:"extra_data,omitempty"`
}

// CheckLoginAttemptRequest 检查登录限制请求
type CheckLoginAttemptRequest struct {
	UserID      *string       `json:"user_id,omitempty"`
	Username    *string       `json:"username,omitempty"`
	Email       *string       `json:"email,omitempty"`
	IPAddress   string        `json:"ip_address" validate:"required,ip"`
	CheckWindow time.Duration `json:"check_window,omitempty"` // 检查时间窗口，默认使用配置
}

// LoginAttemptCheckResult 登录限制检查结果
type LoginAttemptCheckResult struct {
	Allowed           bool          `json:"allowed"`
	Blocked           bool          `json:"blocked"`
	BlockReason       string        `json:"block_reason,omitempty"`
	RemainingAttempts int           `json:"remaining_attempts"`
	BlockedUntil      *time.Time    `json:"blocked_until,omitempty"`
	SuggectedDelay    time.Duration `json:"suggested_delay,omitempty"`
	RiskLevel         string        `json:"risk_level"` // low, medium, high, critical
}

// LoginAttemptStatsRequest 登录尝试统计请求
type LoginAttemptStatsRequest struct {
	UserID      *string    `json:"user_id,omitempty"`
	IPAddress   *string    `json:"ip_address,omitempty"`
	StartTime   *time.Time `json:"start_time,omitempty"`
	EndTime     *time.Time `json:"end_time,omitempty"`
	AttemptType *string    `json:"attempt_type,omitempty"`
	SuccessOnly *bool      `json:"success_only,omitempty"`
}

// LoginAttemptStatsResponse 登录尝试统计响应
type LoginAttemptStatsResponse struct {
	TotalAttempts      int64                `json:"total_attempts"`
	SuccessfulAttempts int64                `json:"successful_attempts"`
	FailedAttempts     int64                `json:"failed_attempts"`
	UniqueIPs          int64                `json:"unique_ips"`
	UniqueUsers        int64                `json:"unique_users"`
	AttemptsByType     map[string]int64     `json:"attempts_by_type"`
	AttemptsByHour     map[string]int64     `json:"attempts_by_hour"`
	TopFailureReasons  []FailureReasonStats `json:"top_failure_reasons"`
	SuspiciousIPs      []SuspiciousIPStats  `json:"suspicious_ips"`
	MostActiveUsers    []UserActivityStats  `json:"most_active_users"`
}

// FailureReasonStats 失败原因统计
type FailureReasonStats struct {
	Reason string `json:"reason"`
	Count  int64  `json:"count"`
}

// SuspiciousIPStats 可疑IP统计
type SuspiciousIPStats struct {
	IPAddress   string    `json:"ip_address"`
	FailedCount int64     `json:"failed_count"`
	LastAttempt time.Time `json:"last_attempt"`
	RiskScore   float64   `json:"risk_score"`
}

// UserActivityStats 用户活动统计
type UserActivityStats struct {
	UserID       string    `json:"user_id"`
	Username     string    `json:"username"`
	AttemptCount int64     `json:"attempt_count"`
	SuccessRate  float64   `json:"success_rate"`
	LastAttempt  time.Time `json:"last_attempt"`
}

// ============================================================================
// IP管理相关DTO
// ============================================================================

// BlockedIPInfo 被阻断IP信息
type BlockedIPInfo struct {
	IPAddress    string     `json:"ip_address"`
	BlockedAt    time.Time  `json:"blocked_at"`
	BlockedUntil *time.Time `json:"blocked_until,omitempty"`
	Reason       string     `json:"reason"`
	FailedCount  int64      `json:"failed_count"`
	Permanent    bool       `json:"permanent"`
}

// ============================================================================
// 风险分析相关DTO
// ============================================================================

// LoginRiskAnalysisRequest 登录风险分析请求
type LoginRiskAnalysisRequest struct {
	UserID     *string        `json:"user_id,omitempty"`
	Username   *string        `json:"username,omitempty"`
	Email      *string        `json:"email,omitempty"`
	IPAddress  string         `json:"ip_address" validate:"required,ip"`
	UserAgent  string         `json:"user_agent"`
	Location   *LoginLocation `json:"location,omitempty"`
	DeviceInfo *DeviceInfo    `json:"device_info,omitempty"`
}

// LoginRiskAnalysisResponse 登录风险分析响应
type LoginRiskAnalysisResponse struct {
	RiskScore        float64      `json:"risk_score"` // 0-100
	RiskLevel        string       `json:"risk_level"` // low, medium, high, critical
	RiskFactors      []RiskFactor `json:"risk_factors"`
	Recommendations  []string     `json:"recommendations"`
	RequiresAction   bool         `json:"requires_action"`
	SuggestedActions []string     `json:"suggested_actions,omitempty"`
}

// RiskFactor 风险因子
type RiskFactor struct {
	Type        string                 `json:"type"`     // ip_reputation, location_change, device_change, etc.
	Severity    string                 `json:"severity"` // low, medium, high, critical
	Score       float64                `json:"score"`    // 贡献的风险分数
	Description string                 `json:"description"`
	Details     map[string]interface{} `json:"details,omitempty"`
}

// SuspiciousActivityRequest 可疑活动请求
type SuspiciousActivityRequest struct {
	StartTime    *time.Time `json:"start_time,omitempty"`
	EndTime      *time.Time `json:"end_time,omitempty"`
	RiskLevel    *string    `json:"risk_level,omitempty"`    // low, medium, high, critical
	ActivityType *string    `json:"activity_type,omitempty"` // login_attempt, password_change, etc.
	IPAddress    *string    `json:"ip_address,omitempty"`
	UserID       *string    `json:"user_id,omitempty"`
	Limit        int        `json:"limit,omitempty"`
	Offset       int        `json:"offset,omitempty"`
}

// SuspiciousActivityResponse 可疑活动响应
type SuspiciousActivityResponse struct {
	Activities  []SuspiciousActivity `json:"activities"`
	Total       int64                `json:"total"`
	RiskSummary RiskSummary          `json:"risk_summary"`
}

// SuspiciousActivity 可疑活动
type SuspiciousActivity struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"`
	UserID      *string                `json:"user_id,omitempty"`
	IPAddress   string                 `json:"ip_address"`
	Timestamp   time.Time              `json:"timestamp"`
	RiskScore   float64                `json:"risk_score"`
	RiskLevel   string                 `json:"risk_level"`
	Description string                 `json:"description"`
	RiskFactors []RiskFactor           `json:"risk_factors"`
	Status      string                 `json:"status"` // pending, investigated, resolved, false_positive
	Details     map[string]interface{} `json:"details,omitempty"`
}

// RiskSummary 风险总结
type RiskSummary struct {
	TotalActivities    int64            `json:"total_activities"`
	RiskDistribution   map[string]int64 `json:"risk_distribution"` // low: 10, medium: 5, high: 2, critical: 1
	TopRiskFactors     []string         `json:"top_risk_factors"`
	MostActiveIPs      []string         `json:"most_active_ips"`
	RecommendedActions []string         `json:"recommended_actions"`
}

// ============================================================================
// 安全配置相关DTO
// ============================================================================

// SecurityConfigResponse 安全配置响应
type SecurityConfigResponse struct {
	LoginAttemptConfig LoginAttemptConfig `json:"login_attempt_config"`
	IPBlockingConfig   IPBlockingConfig   `json:"ip_blocking_config"`
	RiskAnalysisConfig RiskAnalysisConfig `json:"risk_analysis_config"`
	CleanupConfig      CleanupConfig      `json:"cleanup_config"`
}

// UpdateSecurityConfigRequest 更新安全配置请求
type UpdateSecurityConfigRequest struct {
	LoginAttemptConfig *LoginAttemptConfig `json:"login_attempt_config,omitempty"`
	IPBlockingConfig   *IPBlockingConfig   `json:"ip_blocking_config,omitempty"`
	RiskAnalysisConfig *RiskAnalysisConfig `json:"risk_analysis_config,omitempty"`
	CleanupConfig      *CleanupConfig      `json:"cleanup_config,omitempty"`
}

// LoginAttemptConfig 登录尝试配置
type LoginAttemptConfig struct {
	MaxFailedAttempts   int           `json:"max_failed_attempts"`
	FailedAttemptWindow time.Duration `json:"failed_attempt_window"`
	BlockDuration       time.Duration `json:"block_duration"`
	ProgressiveDelay    bool          `json:"progressive_delay"`
	Enabled             bool          `json:"enabled"`
}

// IPBlockingConfig IP阻断配置
type IPBlockingConfig struct {
	MaxFailedAttemptsPerIP int           `json:"max_failed_attempts_per_ip"`
	IPBlockDuration        time.Duration `json:"ip_block_duration"`
	Enabled                bool          `json:"enabled"`
	Whitelist              []string      `json:"whitelist"`
	Blacklist              []string      `json:"blacklist"`
}

// RiskAnalysisConfig 风险分析配置
type RiskAnalysisConfig struct {
	EnableLocationCheck     bool    `json:"enable_location_check"`
	EnableDeviceCheck       bool    `json:"enable_device_check"`
	SuspiciousThreshold     float64 `json:"suspicious_threshold"`
	Enabled                 bool    `json:"enabled"`
	LocationChangeThreshold float64 `json:"location_change_threshold"` // km
	DeviceChangeThreshold   float64 `json:"device_change_threshold"`   // similarity score
}

// CleanupConfig 清理配置
type CleanupConfig struct {
	CleanupInterval time.Duration `json:"cleanup_interval"`
	RetentionPeriod time.Duration `json:"retention_period"`
	AutoCleanup     bool          `json:"auto_cleanup"`
	BatchSize       int           `json:"batch_size"`
}
