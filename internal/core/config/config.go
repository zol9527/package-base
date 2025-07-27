package config

import (
	"fmt"
	"time"

	"package-base/internal/core/storages"
)

// Config 应用全局配置
type Config struct {
	App      *AppConfig      `json:"app" yaml:"app"`
	Server   *ServerConfig   `json:"server" yaml:"server"`
	Database *DatabaseConfig `json:"database" yaml:"database"`
	JWT      *JWTConfig      `json:"jwt" yaml:"jwt"`
	Security *SecurityConfig `json:"security" yaml:"security"`
	Cors     *CorsConfig     `json:"cors" yaml:"cors"`
	Logger   *LoggerConfig   `json:"logger" yaml:"logger"`
	Logto    *LogtoConfig    `json:"logto" yaml:"logto"`
	Session  *SessionConfig  `json:"session" yaml:"session"`
}

// AppConfig 应用配置
type AppConfig struct {
	Name    string `json:"name" yaml:"name"`
	Version string `json:"version" yaml:"version"`
	Mode    string `json:"mode" yaml:"mode"` // development, production, test
	Debug   bool   `json:"debug" yaml:"debug"`
	Domain  string `json:"domain" yaml:"domain"` // 应用域名
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Port         int    `json:"port" yaml:"port"`
	Host         string `json:"host" yaml:"host"`
	ReadTimeout  int    `json:"read_timeout" yaml:"read_timeout"`   // 秒
	WriteTimeout int    `json:"write_timeout" yaml:"write_timeout"` // 秒
	IdleTimeout  int    `json:"idle_timeout" yaml:"idle_timeout"`   // 秒
	BodyLimit    string `json:"body_limit" yaml:"body_limit"`
	Prefork      bool   `json:"prefork" yaml:"prefork"`
}

// GetReadTimeout 获取读取超时时间
func (s *ServerConfig) GetReadTimeout() time.Duration {
	return time.Duration(s.ReadTimeout) * time.Second
}

// GetWriteTimeout 获取写入超时时间
func (s *ServerConfig) GetWriteTimeout() time.Duration {
	return time.Duration(s.WriteTimeout) * time.Second
}

// GetIdleTimeout 获取空闲超时时间
func (s *ServerConfig) GetIdleTimeout() time.Duration {
	return time.Duration(s.IdleTimeout) * time.Second
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	MySQL *storages.MySQLConfig `json:"mysql" yaml:"mysql"`
	Redis *storages.RedisConfig `json:"redis" yaml:"redis"`
}

// JWTConfig JWT配置
type JWTConfig struct {
	Secret              string `json:"secret" yaml:"secret"`
	AccessTokenExpire   int    `json:"access_token_expire" yaml:"access_token_expire"`   // 秒
	RefreshTokenExpire  int    `json:"refresh_token_expire" yaml:"refresh_token_expire"` // 秒
	Issuer              string `json:"issuer" yaml:"issuer"`
	EnableRefreshRotate bool   `json:"enable_refresh_rotate" yaml:"enable_refresh_rotate"`
}

// GetAccessTokenExpire 获取访问令牌过期时间
func (j *JWTConfig) GetAccessTokenExpire() time.Duration {
	return time.Duration(j.AccessTokenExpire) * time.Second
}

// GetRefreshTokenExpire 获取刷新令牌过期时间
func (j *JWTConfig) GetRefreshTokenExpire() time.Duration {
	return time.Duration(j.RefreshTokenExpire) * time.Second
}

// SecurityConfig 安全配置
type SecurityConfig struct {
	PasswordMinLength    int  `json:"password_min_length" yaml:"password_min_length"`
	PasswordRequireUpper bool `json:"password_require_upper" yaml:"password_require_upper"`
	PasswordRequireLower bool `json:"password_require_lower" yaml:"password_require_lower"`
	PasswordRequireDigit bool `json:"password_require_digit" yaml:"password_require_digit"`
	PasswordRequireSpec  bool `json:"password_require_spec" yaml:"password_require_spec"`
	MaxLoginAttempts     int  `json:"max_login_attempts" yaml:"max_login_attempts"`
	LoginAttemptWindow   int  `json:"login_attempt_window" yaml:"login_attempt_window"`   // 分钟
	AccountLockDuration  int  `json:"account_lock_duration" yaml:"account_lock_duration"` // 分钟
}

// CORS 配置
type CorsConfig struct {
	AllowOrigin      string `json:"allow_origin" yaml:"allow_origin"`
	AllowMethods     string `json:"allow_methods" yaml:"allow_methods"`
	AllowHeaders     string `json:"allow_headers" yaml:"allow_headers"`
	AllowCredentials bool   `json:"allow_credentials" yaml:"allow_credentials"`
	ExposeHeaders    string `json:"expose_headers" yaml:"expose_headers"`
	MaxAge           int    `json:"max_age" yaml:"max_age"`
}

// LoggerConfig 日志配置
type LoggerConfig struct {
	Level            string `json:"level" yaml:"level"`
	Format           string `json:"format" yaml:"format"`
	EnableFile       bool   `json:"enable_file" yaml:"enable_file"`
	FilePath         string `json:"file_path" yaml:"file_path"`
	MaxSize          int    `json:"max_size" yaml:"max_size"`
	MaxBackups       int    `json:"max_backups" yaml:"max_backups"`
	MaxAge           int    `json:"max_age" yaml:"max_age"`
	Compress         bool   `json:"compress" yaml:"compress"`
	LocalTime        bool   `json:"local_time" yaml:"local_time"`
	EnableConsole    bool   `json:"enable_console" yaml:"enable_console"`
	ColorOutput      bool   `json:"color_output" yaml:"color_output"`
	EnableCaller     bool   `json:"enable_caller" yaml:"enable_caller"`
	EnableStacktrace bool   `json:"enable_stacktrace" yaml:"enable_stacktrace"`
}

type SessionConfig struct {
	Storage string `json:"storage" yaml:"storage"` // 存储类型，如 "memory", "redis"
}

// DefaultConfig 返回默认配置
func DefaultConfig() *Config {
	return &Config{
		App: &AppConfig{
			Name:    "package-base",
			Version: "1.0.0",
			Mode:    "development",
			Debug:   true,
			Domain:  "localhost", // 默认域名
		},
		Server: &ServerConfig{
			Port:         8080,
			Host:         "localhost",
			ReadTimeout:  10,
			WriteTimeout: 10,
			IdleTimeout:  120,
			BodyLimit:    "4MB",
			Prefork:      false,
		},
		Logto: &LogtoConfig{
			Enabled:     true,
			Endpoint:    "https://api.logto.io",
			Secret:      "your-api-key",
			AppID:       "your-app-id",
			CallbackURI: "http://localhost:8080/callback",
		},
		Database: &DatabaseConfig{
			MySQL: &storages.MySQLConfig{
				Host:            "localhost",
				Port:            3306,
				Username:        "root",
				Password:        "password",
				Database:        "package_base",
				Charset:         "utf8mb4",
				ParseTime:       true,
				Loc:             "Local",
				MaxOpenConns:    100,
				MaxIdleConns:    10,
				ConnMaxLifetime: time.Hour,
				ConnMaxIdleTime: 10 * time.Minute,
				LogLevel:        "info",
			},
			Redis: &storages.RedisConfig{
				Host:          "localhost",
				Port:          6379,
				Username:      "",
				Password:      "",
				Database:      0,
				PoolSize:      10,
				MinIdleConns:  1,
				DialTimeout:   5 * time.Second,
				ReadTimeout:   3 * time.Second,
				WriteTimeout:  3 * time.Second,
				PoolTimeout:   4 * time.Second,
				IdleTimeout:   5 * time.Minute,
				IdleCheckFreq: time.Minute,
				MaxConnAge:    30 * time.Minute,
				KeyPrefix:     "auth:",
				DefaultTTL:    24 * time.Hour,
			},
		},
		Session: &SessionConfig{
			Storage: "redis",
		},
		JWT: &JWTConfig{
			Secret:              "your-secret-key-change-in-production",
			AccessTokenExpire:   900,   // 15分钟
			RefreshTokenExpire:  86400, // 24小时
			Issuer:              "package-base",
			EnableRefreshRotate: true,
		},
		Cors: &CorsConfig{
			AllowOrigin:      "*",
			AllowMethods:     "GET,POST,HEAD,PUT,DELETE,PATCH,OPTIONS",
			AllowHeaders:     "Origin,Content-Type,Accept,Authorization,X-Request-ID",
			ExposeHeaders:    "X-Request-ID",
			AllowCredentials: true,
			MaxAge:           3600,
		},
		Security: &SecurityConfig{
			PasswordMinLength:    8,
			PasswordRequireUpper: true,
			PasswordRequireLower: true,
			PasswordRequireDigit: true,
			PasswordRequireSpec:  false,
			MaxLoginAttempts:     5,
			LoginAttemptWindow:   15, // 15分钟
			AccountLockDuration:  30, // 30分钟
		},
		Logger: &LoggerConfig{
			Level:            "info",
			Format:           "json",
			EnableFile:       false,
			FilePath:         "logs/app.log",
			MaxSize:          100,
			MaxBackups:       3,
			MaxAge:           28,
			Compress:         true,
			LocalTime:        true,
			EnableConsole:    true,
			ColorOutput:      true,
			EnableCaller:     true,
			EnableStacktrace: false,
		},
	}
}

// Validate 验证配置
func (c *Config) Validate() error {
	if c.App == nil {
		return fmt.Errorf("app config is required")
	}
	if c.Server == nil {
		return fmt.Errorf("server config is required")
	}
	if c.Database == nil {
		return fmt.Errorf("database config is required")
	}
	if c.JWT == nil {
		return fmt.Errorf("jwt config is required")
	}
	if c.Security == nil {
		return fmt.Errorf("security config is required")
	}
	if c.Logger == nil {
		return fmt.Errorf("logger config is required")
	}

	// 验证服务器配置
	if c.Server.Port <= 0 || c.Server.Port > 65535 {
		return fmt.Errorf("invalid server port: %d", c.Server.Port)
	}
	if c.Server.Host == "" {
		return fmt.Errorf("server host is required")
	}

	// 验证JWT配置
	if c.JWT.Secret == "" {
		return fmt.Errorf("jwt secret is required")
	}
	if len(c.JWT.Secret) < 32 {
		return fmt.Errorf("jwt secret must be at least 32 characters")
	}

	// 验证安全配置
	if c.Security.PasswordMinLength < 6 {
		return fmt.Errorf("password min length must be at least 6")
	}

	return nil
}

// IsProduction 检查是否为生产环境
func (c *Config) IsProduction() bool {
	return c.App.Mode == "production"
}

type LogtoConfig struct {
	Enabled     bool   `json:"enabled" yaml:"enabled"`
	Endpoint    string `json:"endpoint" yaml:"endpoint"`
	Secret      string `json:"secret" yaml:"secret"`
	AppID       string `json:"app_id" yaml:"app_id"`
	CallbackURI string `json:"redirect_uri" yaml:"redirect_uri"`
}

func (c *LogtoConfig) Validate() error {
	if c.Endpoint == "" {
		return fmt.Errorf("logto endpoint is required")
	}
	if c.Secret == "" {
		return fmt.Errorf("logto api_key is required")
	}
	if c.AppID == "" {
		return fmt.Errorf("logto app_id is required")
	}
	return nil
}

func (c *LogtoConfig) IsEnabled() bool {
	return c.Enabled
}

func (c *LogtoConfig) GetEndpoint() string {
	return c.Endpoint
}

func (c *LogtoConfig) GetSecret() string {
	return c.Secret
}

func (c *LogtoConfig) GetAppID() string {
	return c.AppID
}

func (c *LogtoConfig) GetCallbackURI() string {
	return c.CallbackURI
}
