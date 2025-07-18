# Logger 使用指南

这是一个基于 Zap 的高性能、可配置的日志组件，支持独立使用和与 Fiber 框架集成。

## 功能特性

- 🚀 基于 Zap 的高性能日志记录
- 📁 支持文件和控制台双重输出
- 🔄 支持日志轮转和压缩
- 🎨 支持彩色控制台输出
- 📊 支持结构化日志（字段记录）
- 🔧 完全可配置的日志级别和格式
- 🌐 支持 Fiber 框架适配
- 🔌 保持核心组件独立性

## 快速开始

### 1. 独立使用

```go
package main

import "your-project/internal/core/log/logger"

func main() {
    // 使用默认配置初始化
    err := logger.Initialize(nil)
    if err != nil {
        panic(err)
    }
    
    // 基础使用
    logger.Info("Hello World!")
    logger.Error("Something went wrong", logger.String("error", "connection failed"))
    
    // 使用命名logger
    log := logger.L().Named("service")
    log.Debug("Service started")
    
    // 同步日志（程序退出前调用）
    logger.Sync()
}
```

### 2. 自定义配置

```go
config := &logger.Config{
    Level:         "debug",
    Encoding:      "console",        // "console" 或 "json"
    Filename:      "logs/app.log",   // 日志文件路径
    MaxSize:       100,              // 最大文件大小(MB)
    MaxBackups:    5,                // 保留文件数量
    MaxAge:        30,               // 保留天数
    Compress:      true,             // 是否压缩旧文件
    EnableColor:   true,             // 是否启用颜色
    EnableConsole: true,             // 是否输出到控制台
    Caller:        true,             // 是否显示调用者信息
    Stacktrace:    false,            // 是否显示堆栈
}

err := logger.Initialize(config)
```

### 3. 与 Fiber 集成

```go
package main

import (
    "github.com/gofiber/fiber/v2"
    fiberlog "github.com/gofiber/fiber/v2/log"
    "your-project/internal/core/log/logger"
)

func main() {
    // 1. 初始化自定义logger
    err := logger.Initialize(&logger.Config{
        Level:         "info",
        Encoding:      "json",
        Filename:      "logs/fiber.log",
        EnableConsole: true,
    })
    if err != nil {
        panic(err)
    }
    
    // 2. 设置Fiber使用自定义logger
    fiberAdapter := logger.NewFiberAdapterFromGlobal()
    fiberlog.SetLogger(fiberAdapter)
    
    // 3. 现在Fiber的日志调用会使用你的自定义logger
    app := fiber.New()
    
    app.Get("/", func(c *fiber.Ctx) error {
        fiberlog.Info("处理请求")  // 使用你的logger
        return c.JSON(fiber.Map{"message": "Hello"})
    })
    
    app.Listen(":3000")
}
```

## API 参考

### 基础日志方法

```go
logger.Debug("调试信息", logger.String("key", "value"))
logger.Info("信息", logger.Int("count", 10))
logger.Warn("警告", logger.Bool("is_valid", false))
logger.Error("错误", logger.ErrorField(err))
logger.Fatal("致命错误")  // 会终止程序
```

### 格式化日志方法

```go
logger.L().Debugf("调试: %s", "message")
logger.L().Infof("信息: %d", 123)
logger.L().Errorf("错误: %v", err)
```

### 字段创建函数

```go
logger.String("key", "value")
logger.Int("key", 123)
logger.Int64("key", 123456)
logger.Float64("key", 3.14)
logger.Bool("key", true)
logger.ErrorField(err)
logger.Any("key", complexObject)
```

### 高级用法

```go
// 创建命名logger
namedLogger := logger.L().Named("module")

// 创建带预设字段的logger
contextLogger := logger.L().With(
    logger.String("service", "user"),
    logger.String("version", "1.0.0"),
)

// 在handler中使用
func CreateUser(c *fiber.Ctx) error {
    log := logger.L().Named("handler.create_user")
    log.Info("处理创建用户请求",
        logger.String("path", c.Path()),
        logger.String("method", c.Method()),
    )
    // ...
}
```

## 架构设计

- `logger.go` - 核心日志组件，完全独立
- `fiber_adapter.go` - Fiber 适配器，不影响核心组件
- `examples.go` - 使用示例和最佳实践

这种设计确保了：
1. 核心 logger 可以在任何项目中独立使用
2. Fiber 适配器提供无缝集成
3. 保持组件之间的低耦合
4. 易于扩展到其他框架

## 注意事项

1. 程序退出前调用 `logger.Sync()` 确保日志写入
2. Fatal 级别会终止程序，请谨慎使用
3. 在生产环境建议使用 JSON 格式便于解析
4. 合理设置日志轮转参数避免磁盘空间不足
