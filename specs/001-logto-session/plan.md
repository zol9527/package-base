# Implementation Plan: 移除 Logto 依赖并重构认证分层

**Branch**: `001-logto-session` | **Date**: 2025-10-08 | **Spec**: ./spec.md
**Input**: Feature specification from `./spec.md`

**Note**: This template is filled in by the `/speckit.plan` command. See `.specify/templates/commands/plan.md` for the execution workflow.

## Summary

移除第三方 Logto 认证依赖，建立自有账号密码登录（JWT HS256 可配置 TTL），清理 session 相关实现并重构分层使 handler→service→repository→storage 依赖方向单向。初期默认 SQLite（开发友好），通过配置支持 MySQL/PostgreSQL/SQLite 可插拔；对象存储通过统一接口支持 OSS / 本地。目标：降低外部耦合、提升可维护性与后续扩展（多因子、权限、更多驱动）的演进速度。

## Technical Context

**Language/Version**: Go 1.24.x  
**Primary Dependencies**: Fiber v2 (HTTP 框架), golang-jwt/jwt v5, GORM, zap, validator, redis(v8)（短期保留仅作缓存/未来用途), aliyun-oss-sdk (可选), go-redis (可选), sqlite / mysql / postgres 驱动（当前只启用 sqlite -> gorm sqlite 驱动 待引入）  
**Storage**: Phase1: SQLite (默认)；可配置 MySQL / PostgreSQL；对象存储：本地文件 + (未来) 阿里 OSS 抽象  
**Testing**: go test, table-driven, coverage thresholds (≥70% new lines, ≥90% security), integration tests via http test server  
**Target Platform**: Linux/macOS server (单进程)  
**Project Type**: backend API (单后端服务)  
**Performance Goals**: auth/login p95 <120ms；冷启动 <2s；容器镜像减小无用依赖  
**Constraints**: 无外部 Logto; 分层依赖单向；可插拔 DB/Storage；无 session 状态  
**Scale/Scope**: 早期单节点 <1k 并发；设计需支持横向扩展（无内存粘性 session）

NEEDS CLARIFICATION (转入 Phase0 研究): 
- PostgreSQL 驱动采用何种版本与参数调优（连接池尺寸）
- OSS 与本地存储统一接口是否需要签名 URL 支持（首期或延后）

## Constitution Check

| 宪章要求 | 当前状态 | 说明 |
|----------|----------|------|
| 代码质量：lint + 日志统一 | 计划合规 | 将移除 fmt 打印，统一 zap 封装 |
| 分层边界：handler 不直接访问存储 | 需实施 | 现存可能 handler 直用 repo/存储待审计 |
| 测试覆盖率阈值 | 待实施 | 添加覆盖率门禁脚本 CI TODO |
| 性能基线 p95 <120ms | 风险低 | 登录流程简单，待基准脚本 |
| 统一响应结构 | 部分 | 确认所有 handler 返回标准结构，缺失处需改 |
| 移除外部耦合(Logto) | 待实施 | 代码与依赖需清理 go.mod / imports |
| 可观测性 trace_id | 部分 | 登录链路需贯穿 trace_id 注入 |

GATE 结论：允许进入 Phase0，需在 Phase1 后复检 “分层边界 / 统一响应结构 / 覆盖率脚本”。

## Project Structure

### Documentation (this feature)

```
specs/[###-feature]/
├── plan.md              # This file (/speckit.plan command output)
├── research.md          # Phase 0 output (/speckit.plan command)
├── data-model.md        # Phase 1 output (/speckit.plan command)
├── quickstart.md        # Phase 1 output (/speckit.plan command)
├── contracts/           # Phase 1 output (/speckit.plan command)
└── tasks.md             # Phase 2 output (/speckit.tasks command - NOT created by /speckit.plan)
```

### Source Code (repository root)
<!--
  ACTION REQUIRED: Replace the placeholder tree below with the concrete layout
  for this feature. Delete unused options and expand the chosen structure with
  real paths (e.g., apps/admin, packages/something). The delivered plan must
  not include Option labels.
-->

```
# [REMOVE IF UNUSED] Option 1: Single project (DEFAULT)
src/
```
internal/
  core/
    config/
    interfaces/ (领域与跨层接口，如 repository/service contracts)
    utils/
  repositories/
    interfaces/
    (具体实现: user_repository.go ...)
  storages/
    mysql_storage.go
    redis_storage.go
    local_storage.go
    aliyun_remote_oss_storage.go
  services/
    interfaces/
    auth_service.go
    user_service.go
  handlers/
    auth_handler.go
    user_handler.go
  middleware/
    auth.go
    logger.go
  container/
    container.go (DI 装配)
cmd/server/main.go
pkg/ (若新增通用库, 暂不创建)
tests/
  integration/
  contract/
  unit/
```

**Structure Decision**: 保持单后端项目，强化 interfaces/ 分层与 container 单点装配；新增 tests/ 三类目录（若当前缺失将逐步引入）。Session 相关移除后清理中间件与模型引用；多数据库/存储通过 DI 选择实现。

## Complexity Tracking

*Fill ONLY if Constitution Check has violations that must be justified*

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| Repository pattern | 需要隔离多数据库/存储差异 | 直接 DB 访问无法无侵入切换驱动 & 测试替身困难 |
| DI 容器 | 统一装配多驱动与可插拔 Token 签发器 | 手工分散初始化增加重复与易错 |
| JWT 签发抽象 | 未来扩展 OAuth/多因子 | 直接散落生成逻辑难以集中治理安全策略 |

## Phase Breakdown

### Phase 0 (Research 完成)
- 验证多数据库抽象最小接口列表 (Connect(), AutoMigrate(User), FindUserByIdentifier, CreateUser)。
- 评估是否需要统一错误包装层（决定延后）。

### Phase 1 (Setup / Skeleton)
- 移除 go.mod 中 logto 依赖与相关 import。
- 删除 / 重构 session 中间件及其引用。
- 新增 `internal/core/interfaces` 中 repository 与 service 接口声明（若缺失）。
- 建立 `ObjectStorage` 接口 + local 实现接入容器。
- JWT 签发器封装：`internal/core/utils/jwt.go`（若已存在扩展 TTL 支持）。
- 统一响应结构中间件/或 handler 返回辅助函数。

### Phase 2 (Foundational Implementation & Test Planning)
- 实现 sqlite repository（默认）。
- 预留 mysql/postgres switch（仅装配分支与空实现占位）。
- 移除遗留 session 表访问代码（若无 side effect）。
- 集成测试：login 成功 / 失败 / 禁用用户。
- 静态检测脚本 (grep) 确认 handlers 未直接引用存储实现。
- 基准脚本（可选初版）指令说明。

### 非目标 (记录避免 scope creep)
- 刷新 token / 登出
- 角色权限/ACL
- MFA 强制流程

## Post-Design Constitution Re-Check (预填写，实施后复核)
> 备注：将新增 trace_id 覆盖测试、覆盖率阈值脚本、懒加载驱动验证与 hash 算法限定检测任务（见 tasks 后续追加部分）。
| 项目 | 状态(预期) | 说明 |
|------|-----------|------|
| 分层依赖 | 将达成 | 通过接口 + container 装配 |
| 覆盖率脚本 | 待添加 | 计划在 tasks 阶段添加 CI 覆盖率门槛 |
| 性能基线 | 可执行 | 提供 wrk 脚本示例 |
| 统一响应 | 达成 | 登录与未来受保护接口采用 helper |
| 移除 Logto | 达成 | go.mod 与代码清理 |

