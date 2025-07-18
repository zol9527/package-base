package log

import (
	"fmt"
	"package-base/internal/core/interfaces"
)

// FiberLoggerAdapter Fiber日志适配器
//
// 描述: 将自定义Logger适配为Fiber所需的日志接口，保持核心logger的独立性
type FiberLoggerAdapter struct {
	logger interfaces.Logger
}

// NewFiberAdapter 创建Fiber日志适配器
//
// 描述: 基于现有logger创建Fiber适配器
//
// 参数:
//
//	logger (interfaces.Logger): 自定义logger实例
//
// 返回:
//
//	*FiberLoggerAdapter: Fiber适配器实例
func NewFiberAdapter(logger interfaces.Logger) *FiberLoggerAdapter {
	return &FiberLoggerAdapter{
		logger: logger,
	}
}

// NewFiberAdapterFromGlobal 从全局logger创建Fiber适配器
//
// 描述: 使用全局logger实例创建Fiber适配器
//
// 返回:
//
//	*FiberLoggerAdapter: Fiber适配器实例
func NewFiberAdapterFromGlobal() *FiberLoggerAdapter {
	return &FiberLoggerAdapter{
		logger: L(),
	}
}

// 基础日志方法 - 支持可变参数
func (f *FiberLoggerAdapter) Debug(args ...interface{}) {
	f.logger.Debug(fmt.Sprint(args...))
}

func (f *FiberLoggerAdapter) Info(args ...interface{}) {
	f.logger.Info(fmt.Sprint(args...))
}

func (f *FiberLoggerAdapter) Warn(args ...interface{}) {
	f.logger.Warn(fmt.Sprint(args...))
}

func (f *FiberLoggerAdapter) Error(args ...interface{}) {
	f.logger.Error(fmt.Sprint(args...))
}

func (f *FiberLoggerAdapter) Fatal(args ...interface{}) {
	f.logger.Fatal(fmt.Sprint(args...))
}

func (f *FiberLoggerAdapter) Panic(args ...interface{}) {
	// 注意: 由于我们的logger没有Panic方法，使用Fatal替代
	f.logger.Fatal(fmt.Sprint(args...))
}

// 格式化日志方法
func (f *FiberLoggerAdapter) Debugf(format string, args ...interface{}) {
	f.logger.Debug(fmt.Sprintf(format, args...))
}

func (f *FiberLoggerAdapter) Infof(format string, args ...interface{}) {
	f.logger.Info(fmt.Sprintf(format, args...))
}

func (f *FiberLoggerAdapter) Warnf(format string, args ...interface{}) {
	f.logger.Warn(fmt.Sprintf(format, args...))
}

func (f *FiberLoggerAdapter) Errorf(format string, args ...interface{}) {
	f.logger.Error(fmt.Sprintf(format, args...))
}

func (f *FiberLoggerAdapter) Fatalf(format string, args ...interface{}) {
	f.logger.Fatal(fmt.Sprintf(format, args...))
}

func (f *FiberLoggerAdapter) Panicf(format string, args ...interface{}) {
	// 注意: 由于我们的logger没有Panic方法，使用Fatal替代
	f.logger.Fatal(fmt.Sprintf(format, args...))
}

// 结构化日志方法 - 支持键值对
func (f *FiberLoggerAdapter) Debugw(msg string, keysAndValues ...interface{}) {
	fields := f.convertKeysAndValues(keysAndValues...)
	f.logger.Debug(msg, fields...)
}

func (f *FiberLoggerAdapter) Infow(msg string, keysAndValues ...interface{}) {
	fields := f.convertKeysAndValues(keysAndValues...)
	f.logger.Info(msg, fields...)
}

func (f *FiberLoggerAdapter) Warnw(msg string, keysAndValues ...interface{}) {
	fields := f.convertKeysAndValues(keysAndValues...)
	f.logger.Warn(msg, fields...)
}

func (f *FiberLoggerAdapter) Errorw(msg string, keysAndValues ...interface{}) {
	fields := f.convertKeysAndValues(keysAndValues...)
	f.logger.Error(msg, fields...)
}

func (f *FiberLoggerAdapter) Fatalw(msg string, keysAndValues ...interface{}) {
	fields := f.convertKeysAndValues(keysAndValues...)
	f.logger.Fatal(msg, fields...)
}

func (f *FiberLoggerAdapter) Panicw(msg string, keysAndValues ...interface{}) {
	// 注意: 由于我们的logger没有Panic方法，使用Fatal替代
	fields := f.convertKeysAndValues(keysAndValues...)
	f.logger.Fatal(msg, fields...)
}

// convertKeysAndValues 转换键值对为Field切片
//
// 描述: 将Fiber风格的键值对转换为自定义Field接口
//
// 参数:
//
//	keysAndValues (...interface{}): 键值对参数
//
// 返回:
//
//	[]interfaces.Field: 转换后的字段切片
func (f *FiberLoggerAdapter) convertKeysAndValues(keysAndValues ...interface{}) []interfaces.Field {
	var fields []interfaces.Field

	// 确保参数成对出现
	for i := 0; i < len(keysAndValues); i += 2 {
		key := fmt.Sprintf("%v", keysAndValues[i])
		var value interface{}
		if i+1 < len(keysAndValues) {
			value = keysAndValues[i+1]
		}
		fields = append(fields, interfaces.Any(key, value))
	}

	// 处理奇数个参数的情况
	if len(keysAndValues)%2 != 0 {
		fields = append(fields, interfaces.Any("unpaired", keysAndValues[len(keysAndValues)-1]))
	}

	return fields
}

// GetLogger 获取底层logger实例
//
// 描述: 获取适配器包装的原始logger，用于直接操作
//
// 返回:
//
//	interfaces.Logger: 底层logger实例
func (f *FiberLoggerAdapter) GetLogger() interfaces.Logger {
	return f.logger
}

// SetupFiberLogger 设置Fiber全局日志器
//
// 描述: 将自定义logger设置为Fiber的全局日志器的便捷函数
//
// 参数:
//
//	config (*Config): 日志配置，如果为nil则使用默认配置
//
// 返回:
//
//	error: 设置过程中的错误
func SetupFiberLogger(config *Config) error {
	// 初始化全局logger
	if err := Initialize(config); err != nil {
		return err
	}

	// 注意: 这里需要在实际使用时导入fiber的log包并调用SetLogger
	// 由于要保持独立性，这里只返回适配器实例
	// 实际调用应该是: fiberlog.SetLogger(NewFiberAdapterFromGlobal())

	return nil
}
