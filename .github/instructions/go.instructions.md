---
applyTo: '**/*.go'
description: Go 语言项目开发规范和架构设计
---

您是 Go 语言和可扩展系统开发的专家。现在需要您根据以下要求协助我完成项目开发。

## 项目基础要求

- 主要任务：开发和实现所需的 Go 模块、服务或工具，并协助用户思考实现方案。
- 编程语言：Go 始终使用最新稳定版本的 Go（当前为 Go 1.24 或更高版本）。
- 编写标准的 RESTful API，遵循 RESTful 设计原则。

## 项目技术栈要求

- 使用 `fiber` 框架开发 API 
- 使用 `wire` 进行依赖注入
- 使用 `gorm.io/gorm` 作为 ORM 框架
- 使用 `duke-git/lancet/v2` 作为工具库
- 使用 `github.com/pkg/errors` 作为错误处理库
- 如需其他常用库，优先选择社区广泛接受且维护良好的库。

## 项目结构和文件约束（严格遵守）

### 目录结构规范
```
<project_name>/
├── cmd/                  # 应用程序入口点
├── internal/
│   ├── options/          # 函数式查询数据定义文件夹
│   ├── handlers/         # HTTP 处理器层（Controller）
│   ├── services/         # 业务逻辑层（Service）
│   │   └── interfaces/   # 接口定义
│   ├── repositories/     # 数据访问层（Repository）
│   │   └── interfaces/   # 接口定义
│   ├── models/           # 数据模型定义（仅限模型定义）
│   ├── dto/              # 数据传输对象
│   └── core/             # 公共包和工具
│       └── logger/       # 自定义日志
│       └── interfaces/   # 自定义接口
│       └── config/       # 自定义配置
│       └── storage/      # 数据存储（远程、数据库、自定义）
├── tests/                # 测试文件
└── docs/                 # 文档
```

### 严格的文件约束规则

#### Models 目录约束
- **models/ 目录仅用于数据模型定义**：
  - 只能包含结构体定义、字段标签、模型验证规则
  - 禁止包含任何业务逻辑、数据库操作、HTTP 处理逻辑
  - 禁止包含方法实现（除了简单的 getter/setter 和字符串方法）
  - 模型文件命名：`<entity_name>_model.go`

#### Handlers 目录约束
- **handlers/ 目录仅用于 HTTP 请求处理**：
  - 只能包含 HTTP 请求解析、响应格式化、参数验证
  - 禁止包含复杂业务逻辑
  - 必须调用 services 层处理业务逻辑
  - 禁止直接操作数据库或调用 repositories 层
  - 处理器文件命名：`<entity_name>_handler.go`

#### Services 目录约束
- **services/ 目录用于业务逻辑实现**：
  - 包含核心业务逻辑、业务规则验证、事务管理
  - 可以调用 repositories 层进行数据操作
  - 禁止直接处理 HTTP 请求/响应
  - 禁止直接写入 models（只能读取和传递）
  - 服务文件命名：`<entity_name>_service.go`

#### Repositories 目录约束
- **repositories/ 目录用于数据访问**：
  - 只能包含数据库操作、缓存操作、外部 API 调用
  - 禁止包含业务逻辑
  - 可以创建和返回 models，但不能修改传入的 models
  - 仓库文件命名：`<entity_name>_repository.go`

#### DTO 目录约束
- **dto/ 目录用于数据传输对象**：
  - 用于不同层之间的数据传输
  - 请求/响应对象定义
  - 禁止包含业务逻辑
  - DTO 文件命名：`<entity_name>_dto.go`

### 分层架构约束

#### 调用关系约束
```
Handlers -> Services -> Repositories -> Storage(Models) -> 底层数据库
    ↓           ↓            ↓
   DTOs       DTOs        Models
```

- **严禁跨层调用**：
  - Handlers 不能直接调用 Repositories
  - Services 不能直接处理 HTTP 请求
  - Repositories 不能包含业务逻辑

#### 数据流约束
- **向下传递**：上层可以创建 DTO 传递给下层
- **向上返回**：下层返回 Models 或 DTOs 给上层
- **禁止修改**：业务层不能修改从 repositories 返回的 models
- **转换规则**：
  - Handlers: HTTP 请求 -> DTO -> Services
  - Services: DTO -> Repository 参数 -> Models
  - Repositories: 参数 -> Models -> 数据库

## 架构和代码规范要求

- 代码风格：
    - 代码必须符合 Go 官方编码标准（`go fmt`、`go lint`）。
    - 变量、函数、结构体等命名应清晰且具有表达性。
    - 遵循 Go 语言习惯和最佳实践。
    - 严禁将后端错误信息直接暴露给前端，需要使用简短的错误码和友好的提示信息。
    - API 响应消息格式固定为：`{code: xxx, msg: "xxx", data: ""}`
- 结构要求：
    - 开发高可读性、可维护性、可扩展性和可复用性的代码。
    - 保持函数体简洁；将复杂功能拆分为多个函数。
    - 合理组织包结构，遵循 Go 的包设计原则。
    - 严格遵循上述分层架构约束。
- 接口设计：
    - 每个 Service 和 Repository 都应定义对应的 `interface` 接口
    - 接口文件放在 `interfaces/` 目录
    - 最后会使用 `wire` 依赖注入实现松耦合
- 日志记录：
    - 使用自定义的 `log` 包, 我通常定义在 `internal/pkg/logger` 这个文件夹下面
    - 从 `storage` 层开始，直到 `service` 层，需要写入 `debug` 级别的日志记录，在 `handler` 层需要写入 `info` 级别或以上的日志记录
    - 请勿直接调用 `log := logger.L().xxx()`, 而是使用 `log := logger.L().Named("xxx")` 来创建日志实例, 随后才开始使用
- 错误处理：
    - 使用 `github.com/pkg/errors` 包，并且从 `storage` 或者 `repository` 中开始使使用，在会遇到 `err` 的地方创建错误或者使用`errors.Wrap`包装给下一层，在 service 层继续包装并添加业务上下文，在顶层 handler 层使用自定义 `log` 打印，并且使用 `log` 记录完整堆栈调试使用
- 并发：必要时利用 Go 的并发特性（Goroutines、Channels）提高性能，确保线程安全。
- 安全性：避免使用不安全的函数如:（`unsafe` 包）。在设计和实现中始终优先考虑安全性。
- 数据库：
    - 在 `options` 目录下创建文件，文件名格式为`query_<entity_name>.go`, 文件中定义通用查询选项如：`type XxxxQueryOption func(*gorm.DB) (*gorm.DB, error)`，随后才定义查询选项函数，查询选项函数命名为 `With<ConditionName>`，例如 `WithID`、`WithName`、`WithAge` 等，选项函数返回通用查询选项 `XxxQueryOption`, 并使用 `Where` 等方法进行查询
    - 在 repository 层定义主查询函数，`SearchXxxQueryByOptions(opts ...QueryOption) (xxx, error)`，使用 `Find` 获取结果，并做好错误处理
    - 使用选项查询时应创建错误，如果查询有异常应中断查询，随后返回错误
    - **创建、删除、更新时一定要正确处理外键关系, 非常重要**

## 代码实现约束示例

### 正确的实现方式

```go
// models/user_model.go - 仅包含模型定义
type User struct {
    ID       uint   `json:"id" gorm:"primaryKey"`
    Username string `json:"username" gorm:"unique;not null"`
    Email    string `json:"email" gorm:"unique;not null"`
}

// repositories/user_repository.go - 仅数据访问
func (r *UserRepository) CreateUser(user *models.User) error {
    log := logger.L().Name("repository.create_user")
    if err := r.db.Create(user).Error; err != nil {
        log.Error("创建用户失败", "error", err)
        return errors.Wrap(err, "创建用户失败")
    }
    return nil
}

// services/user_service.go - 业务逻辑
func (s *UserService) RegisterUser(dto *dto.CreateUserDTO) (*models.User, error) {
    // 业务验证逻辑
    user := &models.User{
        Username: dto.Username,
        Email:    dto.Email,
    }
    return s.userRepo.CreateUser(user)
}

// handlers/user_handler.go - HTTP 处理（fiber 版本）
func (h *UserHandler) CreateUser(c *fiber.Ctx) error {
    log := logger.L().Named("handler.create_user")

    var dto dto.CreateUserDTO
    if err := c.BodyParser(&dto); err != nil {
        log.Error("参数绑定失败", "error", err)
        return c.Status(400).JSON(fiber.Map{"code": 400, "msg": "参数错误", "data": nil})
    }

    user, err := h.userService.RegisterUser(&dto)
    if err != nil {
        log.Error("用户注册失败", "error", err)
        return c.Status(500).JSON(fiber.Map{"code": 500, "msg": "用户注册失败", "data": nil})
    }

    return c.Status(200).JSON(fiber.Map{"code": 200, "msg": "创建成功", "data": user})
}
```

### 错误的实现方式

```go
// 错误：在 models 中包含业务逻辑
type User struct {
    ID uint `json:"id"`
    Username string `json:"username"`
}
func (u *User) ValidateAndSave() error { // 禁止
    // 业务逻辑不应在 model 中
}

// 错误：Handler 直接调用 Repository（fiber 版本错误示例）
func (h *UserHandler) CreateUser(c *fiber.Ctx) error {
    user := &models.User{}
    h.userRepo.Create(user) // 禁止跨层调用
    return nil
}

// 错误：Service 修改传入的 model
func (s *UserService) UpdateUser(user *models.User) error {
    user.UpdatedAt = time.Now() // 禁止修改传入的 model
    return s.userRepo.Update(user)
}
```

## 设计规划要求

- 开发前进行详细的设计规划：
    - 编写代码前，使用伪代码、结构描述、数据流图等详细描述实现方案。
    - 确认设计方案后再开始编码。
    - 明确各层的职责和数据流向。

## 关键约束和原则（严格遵守，非常重要）

- 生成的代码必须完整、可直接复制粘贴并可执行，不得包含任何 TODO、占位符或未完成的实现。
- 代码必须满足上述所有功能、结构和逻辑要求，确保在预期场景下无错误。
- **严格遵守分层架构约束**，禁止跨层调用和职责混乱。
- **严格遵守文件目录约束**，确保代码放在正确的目录中。
- 包含所有必要的导入语句和类型定义。
- 严格遵守所有指定的 Go 版本、第三方库版本和 Go 语言语法标准。
- 始终优先考虑可读性、可维护性、可扩展性和安全性。
- 如对最佳实践或实现细节不确定，请明确说明并提供多个替代方案供用户选择。
- 严格遵循此模板中的所有要求，不得添加或修改功能或提供冗余信息。
