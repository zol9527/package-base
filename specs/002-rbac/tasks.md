# Tasks: RBAC 权限管理与资源访问控制

Feature: RBAC 权限管理与资源访问控制 (Branch: 002-rbac)
Source Docs: spec.md, plan.md, data-model.md, contracts/openapi.yaml, research.md, quickstart.md

## Phase 1 — Setup

- T001 [Setup] 建立 RBAC 目录与骨架 [P] [X]
  - 路径：internal/{middleware,services,repositories,models}/rbac, internal/core/{interfaces,utils}
  - 产出：占位文件与注释、分层依赖校验脚本通过
- T002 [Setup] 统一错误码与响应约定扩展 [P] [X]
  - 路径：internal/core/common（或现有错误码定义处）
  - 产出：新增 401/403 错误码与消息枚举；响应结构对齐宪章
- T003 [Setup] 数据迁移脚手架与占位（明确策略） [P] [X]
  - 路径：internal/models / migrations/
  - 产出：选择并约定迁移方式（AutoMigrate 或 手写 migrations）；创建 Role, Permission, RolePermission, UserRoleBinding, AccessDecision 的迁移占位
- T003a [Setup Tests] 基础测试脚手架与工具 [P] [X]
  - 路径：tests/{unit,integration,contract}/helpers
  - 产出：mock、测试数据加载器、统一断言工具；为后续失败态测试提供基座

## Phase 2 — Foundational (All stories prerequisites)

- T004 [Foundation] 定义领域接口（RBACService, Repository 接口） [X]
  - 路径：internal/core/interfaces/rbac.go
  - 依赖：T001
- T005 [Foundation] 定义模型与验证规则（数据模型对齐） [X]
  - 路径：internal/models/rbac/*.go
  - 依赖：T003
- T006 [Foundation] 实现仓储接口骨架（增删改查+查询） [X]
  - 路径：internal/repositories/rbac/*.go
  - 依赖：T004, T005
- T007 [Foundation] 实现服务层骨架（权限聚合与判定框架） [X]
  - 路径：internal/services/rbac/*.go
  - 依赖：T006
- T008 [Foundation] 鉴权中间件骨架（注入服务、读取用户身份、拦截403/记录决策） [X]
  - 路径：internal/middleware/rbac.go
  - 依赖：T007
- T009 [Foundation] 审计记录基础设施（日志写入通道/函数） [P]
  - 路径：internal/services/rbac/audit.go, internal/repositories/rbac/audit_repository.go
  - 依赖：T006
- T009b [Foundation] 账户状态校验（禁用/锁定优先）集成 [X]
  - 路径：internal/services/rbac/authorize.go, internal/middleware/rbac.go
  - 依赖：T007
- T009c [Foundation] 内置角色（Admin/Manager/Viewer）种子数据与回滚策略 [P]
  - 路径：migrations/ 或启动初始化逻辑
  - 依赖：T003

## Phase 3 — User Story 1 (P1) 为账号分配角色并强制访问控制

- Story 目标：分配角色后立即生效；未授权返回403；决策记录。
- 独立验收：用 Viewer 读取成功、写入拒绝；切换 Manager 后写入成功；日志可见。

- T009a [US1 Tests] 授权失败态测试先行（默认拒绝/403、账户禁用/锁定） [X]
  - 路径：tests/integration/authz_us1_*.go, tests/contract/rbac_us1_*.go
  - 依赖：T003a, T007, T008
- T010 [US1] 创建角色 API（POST /rbac/roles） [X]
  - 路径：internal/handlers/rbac/roles_handler.go
  - 依赖：T007, T006
- T011 [US1] 更新/删除角色 API（PATCH/DELETE /rbac/roles/{id}) [X]
  - 路径：internal/handlers/rbac/roles_handler.go
  - 依赖：T010
- T012 [US1] 用户绑定角色 API（PUT /rbac/users/{userId}/roles）+ 绑定请求 DTO/校验（含可选 scope） [X]
  - 路径：internal/handlers/rbac/bindings_handler.go
  - 依赖：T010
- T013 [US1] 授权中间件接入路由（对受保护路径生效） [X]
  - 路径：internal/routes/routes.go
  - 依赖：T008
- T014 [US1] 有效权限查询 API（GET /rbac/effective-permissions/{userId}) [P] [X]
  - 路径：internal/handlers/rbac/permissions_query_handler.go
  - 依赖：T007
- T015 [US1] 决策日志写入（允许/拒绝）
  - 路径：internal/services/rbac/authorize.go
  - 依赖：T009
- T016 [US1] 集成示例：对现有受保护端点应用角色校验（auth/refresh-token 与 user/profile）
  - 路径：internal/handlers/* 与 internal/middleware 结合
  - 依赖：T013

## Phase 4 — User Story 2 (P2) 自定义角色与精细化权限

- Story 目标：支持资源命名与通配、动作集合（read/write/delete/manage）、显式拒绝优先、精确匹配优先。
- 独立验收：创建 billing.* + read 权限后，GET 成功、POST 403；加入 deny 规则后全部拒绝。

- T017 [US2] 权限实体与校验扩展（动作/效果）
  - 路径：internal/models/rbac/permission.go
  - 依赖：T005
- T018 [US2] 角色权限管理 API（PUT /rbac/roles/{roleId}/permissions）
  - 路径：internal/handlers/rbac/role_permissions_handler.go
  - 依赖：T017, T010
- T018a [US2] 父角色设置 API（PUT /rbac/roles/{roleId}/parents）+ 循环继承检测
  - 路径：internal/handlers/rbac/role_parents_handler.go, internal/services/rbac/role_inherit.go
  - 依赖：T010, T007
- T019 [US2] 策略评估实现：显式拒绝优先、精确匹配优先、通配匹配
  - 路径：internal/services/rbac/authorize.go
  - 依赖：T007
- T020 [US2] 资源匹配器与排序器实现（域:子域:标识 + *） [P]
  - 路径：internal/core/utils/resource_matcher.go
  - 依赖：T019
- T017a [US2 Tests] 策略评估失败态测试（deny 优先、通配与精确冲突）
  - 路径：tests/unit/rbac_eval_*.go, tests/integration/rbac_eval_*.go
  - 依赖：T003a, T019, T020

## Phase 5 — User Story 3 (P3) 审计与可追溯性

- Story 目标：提供审计查询与导出，支持基于用户/资源/动作/时间过滤。
- 独立验收：制造若干允许/拒绝记录；按条件过滤；导出文件含期望字段。

- T021 [US3] 审计查询 API（GET /rbac/audit）
  - 路径：internal/handlers/rbac/audit_handler.go
  - 依赖：T009
- T022 [US3] 审计导出 API（GET /rbac/audit/export） [P]
  - 路径：internal/handlers/rbac/audit_export_handler.go
  - 依赖：T021
- T021a [US3 Tests] 审计过滤与导出测试（含空结果与时间窗边界）
  - 路径：tests/contract/rbac_audit_*.go
  - 依赖：T003a, T021, T022

## Phase 6 — Polish & Cross-Cutting

- T023 [Polish] 错误码/响应结构一致性巡检（401/403/错误消息） [P]
- T024 [Polish] 覆盖率检查与关键路径补测（策略评估分支≥90%）
- T025 [Polish] 性能基线检查（授权判定路径不回退>15%） [P]
- T026 [Polish] 文档与 Quickstart 校对、示例更新 [P]
- T027 [Deferred] 角色/权限导入导出（标记为后续增量）

## Dependencies (Story Order)
- P1 → P2 → P3（每个 Story 可独立演示与交付，P1 为 MVP）

## Parallel Execution Examples
- US1 内并行：T014 可与 T010/T012 并行（不同文件），T015 在 T009 后并行实施。
- US2 内并行：T020 可与 T019 并行（不同文件）；T017a 与 T018a 实现后可并行。
- US3 内并行：T022 可与 T021 后端实现并行（handler 不同文件）。

## Implementation Strategy
- MVP：完成 Phase 1/2/3（US1）即可接入最小闭环：角色/绑定/中间件/有效权限/决策日志。
- 增量：US2 提升精细化控制（含角色继承）；US3 增加合规审计能力；导入导出作为后续增量。
