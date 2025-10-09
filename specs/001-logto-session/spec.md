# Feature Specification: 移除 Logto 依赖并重构认证分层

**Feature Branch**: `001-logto-session`  
**Created**: 2025-10-08  
**Status**: Draft  
**Input**: User description: "移除 logto 依赖，重构认证体系，清理 session 与不清晰分层，建立自建账号密码认证基础"

## Clarifications
### Session 2025-10-08
- Q: 初始 JWT Access Token 过期策略？ → A: 可配置 TTL (TOKEN_TTL) 默认 60 分钟（无刷新）

## User Scenarios & Testing *(mandatory)*

### User Story 1 - 作为开发者我可以在不安装或配置 Logto 的情况下启动项目 (Priority: P1)
系统启动与基础认证相关代码不再硬依赖外部 Logto SDK / 配置；保留的用户、密码、TOTP 等内部逻辑可独立运行。

**Why this priority**: 解除外部耦合是后续自建认证的前置，避免第三方停摆阻塞开发。

**Independent Test**: 在全新环境仅提供项目 README 指定本地配置，执行 `go run cmd/server/main.go` 启动成功且无 Logto 相关错误；禁用 Logto 变量后健康检查接口返回 200。

**Acceptance Scenarios**:
1. **Given** 全新克隆仓库且未设置任何 Logto 环境变量, **When** 启动应用, **Then** 日志无 Logto 初始化失败且 HTTP 端口监听成功。
2. **Given** 当前分支代码, **When** 搜索代码基, **Then** 不存在对 Logto SDK 包导入。

---

### User Story 2 - 作为系统我可提供基础账号密码登录接口并返回标准响应包 (Priority: P2)
提供 `/api/auth/login` 接口：接受用户名/邮箱 + 密码，验证通过后返回统一响应结构（含 code/message/data）。暂不建立持久化 session，直接返回一次性 JWT（或占位 token 结构）。

**Why this priority**: 为后续自建认证与权限体系奠定最小可用登录能力。

**Independent Test**: 通过集成测试：创建测试用户（脚本/fixture），调用登录接口，得到成功响应；错误密码返回统一错误码。

**Acceptance Scenarios**:
1. **Given** 存在启用状态用户与正确密码, **When** 调用登录接口, **Then** 返回 code=0 且 data 中含有 token 字段。
2. **Given** 不存在用户或密码错误, **When** 调用登录接口, **Then** 返回非 0 错误码与通用信息（不泄露存在性）。
3. **Given** 用户被标记禁用, **When** 调用登录接口, **Then** 返回禁止登录错误码。

---

### User Story 3 - 作为维护者我可以清晰看到认证/用户/存储/处理层的分层边界 (Priority: P3)
重构目录 & 依赖：Handler 仅依赖 Service 接口，Service 依赖 Repository 接口与领域模型，Repository 依赖具体存储实现；移除反向依赖与穿透调用。

**Why this priority**: 降低后续扩展（OAuth、第三方登录、权限系统）的修改范围与回归风险。

**Independent Test**: 静态检查（grep 或简单脚本）验证 handlers 不直接 import mysql/redis 实现；所有 service 通过接口引用 repository；容器装配集中在 `internal/container`。

**Acceptance Scenarios**:
1. **Given** 重构完成, **When** 检查 `internal/handlers/*`, **Then** 无直接导入具体存储实现文件。
2. **Given** 重构完成, **When** 修改 repository 实现文件路径, **Then** service 层编译不需改业务逻辑代码（仅容器 wiring 受影响）。

---

### Edge Cases
- 多次失败登录的节流策略：初期不实现 → 记录为后续扩展需求。
- 被软删除/禁用用户尝试登录应返回统一“认证失败”而非暴露状态。
- 密码迁移：确认当前无历史旧散列格式 → 无需兼容路径（若未来引入版本字段再扩展）。
- 未来多因子：保留 TOTP 逻辑但当前登录流程不强制。

## Requirements *(mandatory)*

### Functional Requirements
- **FR-001**: 系统 MUST 移除所有 Logto 相关依赖、配置、环境变量读取与初始化逻辑。
- **FR-002**: 系统 MUST 提供账号/邮箱 + 密码登录接口 `/api/auth/login`。
- **FR-003**: 登录成功 MUST 返回统一响应包 `{code,message,data:{token}}`。
- **FR-004**: 登录失败（账户不存在/密码不匹配/禁用）MUST 返回统一错误结构且不区分具体原因。
- **FR-005**: 系统 MUST 使用现有密码散列与验证工具（内部 `utils/password.go`）进行校验。
- **FR-006**: 系统 MUST 暂不创建持久化 session 记录（去除或跳过 session 中间件逻辑）。
- **FR-007**: 系统 MUST 保留后续扩展用 TOTP 结构与仓储文件但不在当前登录路径强制校验。
- **FR-008**: Handler 层 MUST 不直接导入 `mysql_storage.go` / `redis_storage.go` 等具体实现。
- **FR-009**: Service 层 MUST 仅依赖 repository 接口（`internal/repositories/interfaces`）。
- **FR-010**: 依赖注入 MUST 集中在 `internal/container` 统一装配。
- **FR-011**: 移除的 Logto 代码 MUST 删除或替换为自建抽象，不得留下死代码。
- **FR-012**: 新增 / 变更 / 删除公开 HTTP 路径（含登录及受保护 API） MUST 在同一 PR 中同步更新 README 或 quickstart；DoD：README/quickstart 显示新增路径或标注删除项。
- **FR-013**: 提供最小 JWT 生成（或伪 token）接口封装，token 结构可扩展并含过期时间字段。
- **FR-014**: 系统 MUST 记录认证成功与失败日志（不含敏感原始密码）。
- **FR-015**: 目录结构 MUST 满足宪章“体系分层约束”。
- **FR-016**: 系统 MUST 明确当前仅支持现行散列算法（无旧格式兼容逻辑），未来若引入版本字段需新增迁移流程。
- **FR-017**: 系统 MUST 采用标准 JWT (HS256) 生成登录 token，包含 `sub, exp, iat` 字段，可扩展 `scopes`。
- **FR-017.1**: JWT Access Token 有效期由环境变量 `TOKEN_TTL` 指定（单位: 分钟，缺省=60），当前阶段不提供刷新机制。
- **FR-018**: 系统 MUST 完全移除 session 中间件与相关持久化逻辑（login_attempt/session 表若仍需要需评估保留，否则清理）。
- **FR-019**: 认证层 MUST 通过接口抽象支持未来多数据库驱动（MySQL / PostgreSQL / SQLite），禁用直接在 service 中写特定 SQL 方言；具体实现位于 `internal/storages` 或独立 `internal/repositories` 实现文件。
- **FR-020**: 存储抽象 MUST 允许新增对象存储驱动（OSS、本地文件、其他云）通过统一接口（如 `FileStorage`）插拔，不影响调用方签名。
- **FR-021**: 配置层 MUST 通过统一配置结构选择启用的数据库与存储驱动（例如 `DB_DRIVER=mysql|postgres|sqlite`；`OBJECT_STORAGE=oss|local`）。
- **FR-022**: 容器装配 MUST 依据配置分支构造正确驱动实现并注入（开闭原则：新增驱动仅新增实现 + 装配分支）。
- **FR-023**: 多数据库/多存储实现 MUST 避免在 handler/service 层出现场景判断（if mysql ...），分派逻辑只存在于装配或工厂函数。
- **FR-024**: 若某驱动未启用，相关实现包不执行昂贵初始化（懒加载或条件构造）。

### Key Entities
- **User**: 表示系统用户（id, username/email, password_hash, status, created_at, updated_at）。
- **AuthToken (概念实体)**: 登录生成的 JWT 或占位 token（value, expires_at, subject_id, scopes?）。
- **LoginAttempt (占位)**: 当前功能不强制使用；若保留用于后续风控（字段：id,user_id,success,ip,created_at），Polish 阶段评估是否移除。

## Success Criteria *(mandatory)*

### Measurable Outcomes
- **SC-001**: 新分支首次启动 5 分钟内完成改造并通过编译，运行日志无 Logto 相关 ERROR。（人工观测指标，不做自动 Gate）
- **SC-PERF**: 登录接口成功响应 p95 < 120ms（本地基准），性能基线脚本输出 JSON（含 p50,p95,reqs_per_sec）到 `perf/login_baseline.json`。
- **SC-003**: 登录失败路径（不存在用户/错误密码/禁用）统一返回同一错误码且信息不区分场景。
- **SC-004**: handlers 层零直接引用存储实现（静态分析通过）。
- **SC-005**: 所有修改通过现有测试 + 新增认证测试用例，新增行覆盖率 ≥ 70%。
- **SC-PERF**: auth/login p95 < 120ms。
- **SC-CONSISTENCY**: 新增/变更接口 100% 符合统一响应结构。
- **SC-TEST**: 新增代码覆盖率 ≥ 70%，安全/密码校验路径 ≥ 90%。
- **SC-SEC**: 签发 token 其 exp - iat = TOKEN_TTL(分钟)*60 ± 5 秒（缺省 3600），过期后访问受保护资源必须失败。
- **SC-TRACE**: 登录成功与失败响应 JSON 均含 `trace_id` 字段；对应日志行包含相同 trace_id；trace_id 通过中间件生成并同时写入响应 header `X-Trace-Id`。

## Assumptions
- 假设：现有用户表结构已可支持账号密码登录，不需要迁移。
- 假设：JWT 私钥或签名密钥将以环境变量形式提供（本次允许临时 hardcode 开发值）。
- 假设：暂不支持刷新 token / 登出。
- 假设：TOTP 逻辑延后集成，多因子开关未来由配置决定。

## Out of Scope
- 权限/角色管理重构
- 社交登录 / 第三方 OAuth
- 账号注册 / 密码找回流程
- 登录失败次数锁定策略

## Risks
- 误删仍在使用的 Logto 相关数据结构导致隐藏依赖编译失败
- JWT 临时实现若未隔离后续替换成本高
- 去除 session 中间件可能影响现有依赖其上下文的代码（需审计）
 - 多数据库驱动抽象若设计不当导致接口臃肿或性能损耗
 - 不同 SQL 方言差异（分页/DDL）可能在未来迁移时暴露缺陷
 - 对象存储接口抽象过早泛化可能引入不必要复杂度

## Architectural Decisions
1. Password Hash Compatibility: 当前无旧散列格式 → 不实现兼容/升级逻辑（Decision: Simplify, defer when versioning appears）。
2. Token Format: 使用标准 JWT (HS256)；原因：便于后续权限声明/第三方集成；迁移成本最低。
3. Session Removal: 完全移除 session 中间件与相关持久化，未来若需要重新引入则通过独立 feature 分支实现。
4. Multi-DB Strategy: 通过 repository 接口 + 驱动工厂实现；避免在业务层出现方言特有语法；测试优先保证 MySQL + SQLite。
5. Storage Abstraction: 定义统一文件/对象存储接口（上传/获取/删除/生成临时访问 URL），驱动以实现结构体注册进容器。
6. Configuration Driven Selection: 通过环境变量或配置文件决定激活驱动，未激活的驱动不初始化。

## Open Questions
无（所有 NEEDS CLARIFICATION 已决策）。
