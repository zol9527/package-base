package log

import (
	"fmt"
	"os"
	"package-base/internal/core/interfaces"
	"path/filepath"
	"sync"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

// ============================================================================
// #1 全局变量
// ============================================================================

var (
	globalLogger interfaces.Logger
	once         sync.Once
)

// ============================================================================
// #2 公开的数据结构
// ============================================================================

// LoggerConfig 日志配置
type LoggerConfig struct {
	// 基础配置
	Level  string // 日志级别: debug, info, warn, error, fatal
	Format string // 输出格式: json, console

	// 文件输出配置
	EnableFile bool   // 是否启用文件输出
	FilePath   string // 日志文件路径
	MaxSize    int    // 单个文件最大大小 (MB)
	MaxBackups int    // 最大备份文件数
	MaxAge     int    // 文件保留天数
	Compress   bool   // 是否压缩备份文件
	LocalTime  bool   // 是否使用本地时间

	// 控制台输出配置
	EnableConsole bool // 是否启用控制台输出
	ColorOutput   bool // 是否启用彩色输出

	// 高级配置
	EnableCaller     bool // 是否记录调用者信息
	EnableStacktrace bool // 是否记录堆栈信息
	SampleFirst      int  // 采样配置：初始采样数
	SampleThereafter int  // 采样配置：后续采样频率
}

// ============================================================================
// #3 公开的 API 函数
// ============================================================================

// DefaultConfig 返回默认配置
func DefaultConfig() *LoggerConfig {
	return &LoggerConfig{
		Level:            "info",
		Format:           "console",
		EnableFile:       false,
		FilePath:         "logs/app.log",
		MaxSize:          100,
		MaxBackups:       3,
		MaxAge:           7,
		Compress:         true,
		LocalTime:        true,
		EnableConsole:    true,
		ColorOutput:      true,
		EnableCaller:     true,
		EnableStacktrace: false,
		SampleFirst:      100,
		SampleThereafter: 100,
	}
}

// Initialize 初始化全局日志器
func Initialize(config *LoggerConfig) error {
	var err error
	once.Do(func() {
		globalLogger, err = NewLogger(config)
	})
	return err
}

// L 获取全局日志器实例
func L() interfaces.Logger {
	if globalLogger == nil {
		// 如果全局日志器未初始化，使用默认配置初始化
		_ = Initialize(DefaultConfig())
	}
	return globalLogger
}

// NewLogger 创建新的日志器实例
func NewLogger(config *LoggerConfig) (interfaces.Logger, error) {
	if config == nil {
		config = DefaultConfig()
	}

	zapLogger, err := newZapLogger(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create zap logger: %w", err)
	}

	return &ZapLogger{logger: zapLogger}, nil
}

// Debug 全局调试日志
func Debug(msg string, fields ...interfaces.Field) {
	L().Debug(msg, fields...)
}

// Info 全局信息日志
func Info(msg string, fields ...interfaces.Field) {
	L().Info(msg, fields...)
}

// Warn 全局警告日志
func Warn(msg string, fields ...interfaces.Field) {
	L().Warn(msg, fields...)
}

// Error 全局错误日志
func Error(msg string, fields ...interfaces.Field) {
	L().Error(msg, fields...)
}

// Fatal 全局致命日志
func Fatal(msg string, fields ...interfaces.Field) {
	L().Fatal(msg, fields...)
}

// Sync 同步全局日志器
func Sync() error {
	if globalLogger != nil {
		return globalLogger.Sync()
	}
	return nil
}

// ============================================================================
// #4 内部数据结构
// ============================================================================

// ZapLogger zap 日志器实现
type ZapLogger struct {
	logger *zap.Logger
}

// ============================================================================
// #5 内部方法 (接口实现)
// ============================================================================

// Debug 记录调试日志
func (l *ZapLogger) Debug(msg string, fields ...interfaces.Field) {
	l.logger.Debug(msg, convertFields(fields)...)
}

// Info 记录信息日志
func (l *ZapLogger) Info(msg string, fields ...interfaces.Field) {
	l.logger.Info(msg, convertFields(fields)...)
}

// Warn 记录警告日志
func (l *ZapLogger) Warn(msg string, fields ...interfaces.Field) {
	l.logger.Warn(msg, convertFields(fields)...)
}

// Error 记录错误日志
func (l *ZapLogger) Error(msg string, fields ...interfaces.Field) {
	l.logger.Error(msg, convertFields(fields)...)
}

// Fatal 记录致命日志
func (l *ZapLogger) Fatal(msg string, fields ...interfaces.Field) {
	l.logger.Fatal(msg, convertFields(fields)...)
}

// With 添加字段
func (l *ZapLogger) With(fields ...interfaces.Field) interfaces.Logger {
	return &ZapLogger{
		logger: l.logger.With(convertFields(fields)...),
	}
}

// Named 创建命名日志器
func (l *ZapLogger) Named(name string) interfaces.Logger {
	return &ZapLogger{
		logger: l.logger.Named(name),
	}
}

// Sync 同步日志
func (l *ZapLogger) Sync() error {
	return l.logger.Sync()
}

// ============================================================================
// #6 内部实现函数
// ============================================================================

// newZapLogger 创建zap日志器
func newZapLogger(cfg *LoggerConfig) (*zap.Logger, error) {
	// 1. 设置日志级别
	var level zapcore.Level
	if err := level.UnmarshalText([]byte(cfg.Level)); err != nil {
		level = zapcore.InfoLevel
	}

	// 2. 配置输出核心 (cores)
	var cores []zapcore.Core

	// 基础编码器配置
	baseEncoderConfig := zap.NewProductionEncoderConfig()
	baseEncoderConfig.TimeKey = "timestamp"
	baseEncoderConfig.MessageKey = "message"
	baseEncoderConfig.LevelKey = "level"
	baseEncoderConfig.CallerKey = "caller"
	baseEncoderConfig.StacktraceKey = "stacktrace"
	baseEncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	baseEncoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder

	// 2.1 控制台输出核心
	if cfg.EnableConsole {
		consoleEncoderConfig := baseEncoderConfig
		var consoleEncoder zapcore.Encoder
		if cfg.Format == "console" {
			if cfg.ColorOutput {
				consoleEncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
			}
			consoleEncoder = zapcore.NewConsoleEncoder(consoleEncoderConfig)
		} else {
			consoleEncoder = zapcore.NewJSONEncoder(consoleEncoderConfig)
		}
		cores = append(cores, zapcore.NewCore(consoleEncoder, zapcore.AddSync(os.Stdout), level))
	}

	// 2.2 文件输出核心
	if cfg.EnableFile {
		// 确保日志目录存在
		logDir := filepath.Dir(cfg.FilePath)
		if err := os.MkdirAll(logDir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create log directory: %w", err)
		}

		// 配置 lumberjack
		lumberJackLogger := &lumberjack.Logger{
			Filename:   cfg.FilePath,
			MaxSize:    cfg.MaxSize,
			MaxBackups: cfg.MaxBackups,
			MaxAge:     cfg.MaxAge,
			Compress:   cfg.Compress,
			LocalTime:  cfg.LocalTime,
		}

		// 文件输出强制使用 json 格式
		fileEncoderConfig := baseEncoderConfig
		fileEncoder := zapcore.NewJSONEncoder(fileEncoderConfig)
		cores = append(cores, zapcore.NewCore(fileEncoder, zapcore.AddSync(lumberJackLogger), level))
	}

	// 如果没有任何输出目标，添加默认的控制台输出
	if len(cores) == 0 {
		cores = append(cores, zapcore.NewCore(zapcore.NewConsoleEncoder(baseEncoderConfig), zapcore.AddSync(os.Stdout), level))
	}

	// 3. 创建复合核心
	core := zapcore.NewTee(cores...)

	// 4. 配置选项
	var options []zap.Option
	if cfg.EnableCaller {
		// AddCallerSkip(1) 是为了修正调用栈信息，因为我们封装了 zap
		options = append(options, zap.AddCaller(), zap.AddCallerSkip(1))
	}
	if cfg.EnableStacktrace {
		options = append(options, zap.AddStacktrace(zapcore.ErrorLevel))
	}

	// 5. 采样配置
	if cfg.SampleFirst > 0 && cfg.SampleThereafter > 0 {
		core = zapcore.NewSamplerWithOptions(
			core,
			time.Second,
			cfg.SampleFirst,
			cfg.SampleThereafter,
		)
	}

	// 6. 构建最终 Logger
	logger := zap.New(core, options...)
	return logger, nil
}

// convertFields 转换字段格式
func convertFields(fields []interfaces.Field) []zapcore.Field {
	zapFields := make([]zapcore.Field, len(fields))
	for i, field := range fields {
		zapFields[i] = zap.Any(field.Key(), field.Value())
	}
	return zapFields
}
