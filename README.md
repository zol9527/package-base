# Package Base - 认证系统

基于 Fiber 框架的完整用户认证系统，采用现代 Go 架构设计。

## ✨ 特性

- 🚀 **高性能**: 基于 Fiber v2 框架
- 🔐 **安全**: JWT 认证、密码加密、会话管理
- 🏗️ **架构清晰**: 六边形架构、依赖注入、分层设计
- 🔍 **灵活查询**: 函数式查询选项系统
- 📊 **数据存储**: MySQL + Redis 双重存储
- 🌐 **RESTful API**: 完整的 REST API 设计
- 🛡️ **安全防护**: CORS、请求限制、安全头设置
- 📝 **完整日志**: 结构化日志记录
- 🔧 **易于配置**: 支持环境变量和配置文件

## 🏗️ 系统架构

```
├── cmd/                    # 应用程序入口
├── internal/
│   ├── container/         # 依赖注入容器
│   ├── routes/           # 路由配置
│   ├── handlers/         # HTTP 处理器层
│   ├── services/         # 业务逻辑层
│   ├── repositories/     # 数据访问层
│   ├── models/           # 数据模型
│   ├── dto/              # 数据传输对象
│   ├── options/          # 查询选项
│   ├── middleware/       # 中间件
│   └── core/             # 核心组件
│       ├── config/       # 配置管理
│       ├── storages/     # 存储实现
│       ├── utils/        # 工具函数
│       ├── log/          # 日志系统
│       └── interfaces/   # 接口定义
└── docs/                 # 文档
```

## 🚀 快速开始

### 前置要求

- Go 1.21+
- MySQL 8.0+
- Redis 6.0+

### 安装与配置

1. **克隆项目**
```bash
git clone <repository-url>
cd package-base
```

2. **安装依赖**
```bash
make init
```

3. **配置环境变量**
```bash
cp .env.example .env
# 编辑 .env 文件，配置数据库连接等信息
```

4. **配置数据库**
```bash
# 创建数据库
mysql -u root -p -e "CREATE DATABASE package_base;"

# 运行数据库迁移（如果有）
make migrate-up
```

5. **启动应用程序**
```bash
# 开发模式（热重载）
make dev

# 或者普通启动
go run cmd/server/main.go
```

应用程序将在 `http://localhost:8080` 启动。

## 📖 API 文档

### 认证相关 API

| 方法 | 路径 | 描述 | 认证 |
|------|------|------|------|
| POST | `/api/v1/auth/register` | 用户注册 | ❌ |
| POST | `/api/v1/auth/login` | 用户登录 | ❌ |
| POST | `/api/v1/auth/refresh` | 刷新Token | ❌ |
| POST | `/api/v1/auth/logout` | 用户登出 | ✅ |
| POST | `/api/v1/auth/logout-all` | 登出所有设备 | ✅ |
| PUT | `/api/v1/auth/change-password` | 修改密码 | ✅ |
| GET | `/api/v1/auth/profile` | 获取用户信息 | ✅ |
| GET | `/api/v1/auth/sessions` | 获取会话列表 | ✅ |
| GET | `/api/v1/auth/sessions/active` | 获取活跃会话 | ✅ |
| GET | `/api/v1/auth/sessions/recent` | 获取最近会话 | ✅ |
| GET | `/api/v1/auth/sessions/stats` | 获取会话统计 | ✅ |
| DELETE | `/api/v1/auth/sessions/:id` | 删除指定会话 | ✅ |

### 用户相关 API

| 方法 | 路径 | 描述 | 认证 |
|------|------|------|------|
| GET | `/api/v1/user/profile` | 获取用户资料 | ✅ |
| PUT | `/api/v1/user/profile` | 更新用户资料 | ✅ |

### 系统 API

| 方法 | 路径 | 描述 |
|------|------|------|
| GET | `/health` | 健康检查 |
| GET | `/ping` | 连接测试 |

## 🔧 开发命令

```bash
# 开发相关
make dev              # 启动开发服务器
make watch            # 文件监听模式（需要 air）

# 构建相关
make build            # 构建应用程序
make build-linux      # 构建 Linux 版本
make build-windows    # 构建 Windows 版本
make build-all        # 构建所有平台版本

# 测试相关
make test             # 运行测试
make test-cover       # 测试覆盖率
make benchmark        # 性能测试

# 代码质量
make lint             # 代码检查
make fmt              # 格式化代码
make vet              # Go vet 检查
make tidy             # 整理依赖

# 清理
make clean            # 清理构建文件
make clean-all        # 清理所有文件

# 工具
make install-tools    # 安装开发工具
make health           # 健康检查
```

## ⚙️ 配置说明

### 环境变量配置

主要的环境变量配置：

```bash
# 应用配置
APP_NAME=package-base
APP_MODE=development
SERVER_PORT=8080

# 数据库配置
MYSQL_HOST=localhost
MYSQL_PORT=3306
MYSQL_USERNAME=root
MYSQL_PASSWORD=password
MYSQL_DATABASE=package_base

# Redis 配置
REDIS_HOST=localhost
REDIS_PORT=6379

# JWT 配置
JWT_SECRET=your-super-secret-jwt-key
```

### 配置文件

也可以使用 JSON 配置文件 `config.json`，详细配置选项请参考 `config.json` 示例文件。

## 🔐 安全特性

- **JWT 认证**: 安全的 Token 认证机制
- **密码加密**: 使用 bcrypt 加密存储密码
- **会话管理**: 支持多设备会话管理
- **CORS 保护**: 跨域资源共享保护
- **请求限制**: 防止暴力破解攻击
- **安全头**: 自动添加安全相关的 HTTP 头

## 🗄️ 数据库设计

### 用户表 (users)
- 完整的用户信息管理
- 支持多种状态和角色
- 包含安全设置和偏好配置

### 会话表 (user_sessions)
- 设备信息记录
- 会话状态管理
- 过期时间控制

## 📊 监控与日志

- **结构化日志**: 使用 Zap 进行高性能日志记录
- **健康检查**: `/health` 端点监控系统状态
- **性能监控**: 内置请求ID和响应时间记录

## 🔄 开发流程

1. **创建功能分支**: `git checkout -b feature/new-feature`
2. **编写代码**: 遵循项目架构规范
3. **运行测试**: `make test`
4. **代码检查**: `make lint`
5. **提交代码**: 遵循约定式提交规范
6. **创建 PR**: 代码审查后合并

## 🚀 部署

### Docker 部署

```bash
# 构建镜像
make docker-build

# 运行容器
make docker-run

# 使用 docker-compose
make docker-compose-up
```

### 生产环境部署

1. 设置生产环境变量
2. 构建生产版本: `make build-linux`
3. 配置反向代理 (Nginx)
4. 设置进程管理 (Systemd)
5. 配置日志轮转

## 📝 许可证

MIT License

## 🤝 贡献

欢迎提交 Issue 和 Pull Request！

## 📞 支持

如有问题，请创建 Issue 或联系维护团队。

---

**🎉 享受编码的乐趣！**
