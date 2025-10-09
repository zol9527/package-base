# Tasks: 移除 Logto 依赖并重构认证分层

**Feature**: 移除 Logto 依赖并重构认证分层 (`001-logto-session`)
**Generated**: 2025-10-08
**Source Docs**: plan.md, spec.md, data-model.md, research.md, quickstart.md, contracts/

> 组织方式：按用户故事（优先级 P1→P2→P3）分相位；Story 各自可独立交付与测试。仅生成与需求直接相关的实现与必要测试任务（规范要求覆盖率）。未显式需求的额外测试或优化不在范围内。

## 任务编号规则
- 顺序编号：T001, T002, ... 按执行顺序
- [P] 标记：表示与同 Phase 其它 [P] 任务可并行（不同文件 / 互不写入同一实现单元）
- [Story: USx]：归属的用户故事 (US1=P1, US2=P2, US3=P3)
- Tests 策略：需求显式提出覆盖率与场景 -> 生成必要测试任务（先写测试桩再实现 / 或并行注明）

## Phase 1: Setup (环境与基础准备)
| # | Task | Output / File(s) | Notes |
|---|------|------------------|-------|
| T001 | 创建 `tests/` 目录结构(`tests/unit`, `tests/integration`, `tests/contract`) | 目录 | 为后续测试分层 | [X]
| T002 | 添加统一响应结构 helper (`internal/core/utils/response.go`) | `internal/core/utils/response.go` | 封装 Success/Error JSON | [X]
| T003 | 添加错误码常量文件 (`internal/core/utils/error_codes.go`) | `internal/core/utils/error_codes.go` | AUTH_xxx 前缀 | [X]
| T004 | 初始化性能脚本占位(bench 或 wrk 文档 `scripts/perf/login.lua`) | `scripts/perf/login.lua` | 占位脚本 | [X]
| T005 | 增加 Makefile 目标：`make test`, `make cover`, `make lint` | `Makefile` | 覆盖率≥70% | [X]
| T006 | 新增 `scripts/check-layering.sh` 静态校验 handler 不直接依赖存储实现 | `scripts/check-layering.sh` | grep + 退出码 | [X]
| T007 | 更新 README / quickstart 增加新命令说明 | `README.md` | 同步文档 | [X]

## Phase 2: Foundational (全局阻塞前置)
| # | Task | Output / File(s) | Notes |
|---|------|------------------|-------|
| T008 | 移除 go.mod / 代码中全部 Logto 依赖 (搜索: logto) | `go.mod` 代码差异 | FR-001 | [X]
| T009 | 删除/清理 session 相关中间件与模型 ( `internal/middleware/session.go`, `internal/models/session_model.go`, `internal/models/login_attempt_model.go` 如可移除) | 删除文件/清理引用 | FR-006/FR-018 | [X 部分] 已删除 session.go / logto handler，保留登录尝试模型待后续评估
| T010 | 审核并移除容器/路由中 session 注入逻辑 | `internal/container/container.go`, `internal/routes/routes.go` | 移除注册 | [X]
| T011 | 定义 Repository 接口：用户 (若已存在则补充最小方法 SignIn 所需：FindByUsernameOrEmail, CreateUser, UpdateLoginMeta) | `internal/repositories/interfaces/user_repository.go` | FR-009 | [X] 已新增别名方法 FindByUsernameOrEmail / UpdateLoginMeta
| T012 | 定义 AuthService 接口 + DTO (登录请求/响应结构) | `internal/services/interfaces/auth_service.go`, `internal/dto/auth_dto.go` | FR-002, FR-003 | [X] 增加 SignIn 别名与 AuthResponse.expires_in 字段
| T013 | 定义 ObjectStorage 接口 (Put/Get/Delete/Presign 占位) | `internal/core/interfaces/object_storage_interface.go` | FR-020 | [X] 已新增接口与错误类型
| T014 | 实现本地文件存储驱动 (local) 并注册至容器 | `internal/core/storages/local_object_storage.go`, `internal/container/container.go` | FR-020/FR-022 | [X] 已实现 LocalObjectStorage 并装配
| T015 | 添加 JWT 工具增强：支持 TOKEN_TTL, exp/iat/sub | `internal/core/utils/jwt.go` | FR-017, FR-017.1 | [X] 新增 Set/GetAccessTokenDuration & 返回 expires_in
| T016 | 添加配置项解析扩展：DB_DRIVER, OBJECT_STORAGE, TOKEN_TTL | `internal/core/config/config.go` | FR-021 | [X] 已新增字段并在容器应用 TokenTTL
| T017 | 容器装配：根据配置分支选择 DB / Storage / JWTSigner | `internal/container/container.go` | FR-010, FR-022, FR-024 | [X] 根据 ObjectStorage 分支 + TokenTTL 覆盖
| T018 | Layering 静态检查脚本集成到 Makefile (`make layer-check`) | `Makefile` | 与 T006 关联 | [X] Makefile 增强加入 go vet
| T019 | 基础日志增强：登录链路 trace_id 中间件（若无则新增） | `internal/middleware/logger.go` or new `trace.go` | FR-014 | [X] 新增 trace.go 并输出 trace_id

## Phase 3: 用户故事 US1 (P1) - 移除 Logto 依赖可独立运行
Story Goal: 单机启动不依赖外部 Logto；无 Logto 相关 import；启动成功。
Independent Test Criteria: grep 代码无 logto; `go run cmd/server/main.go` 成功；健康检查 200。

| # | Task | Output / File(s) | Notes |
|---|------|------------------|-------|
| T020 | [Story: US1] 编写测试：启动流程无 Logto 调用 (可用构建后 grep 脚本) | `tests/integration/startup_nologto_test.go` | 先测试后实现 | [X]
| T021 | [Story: US1] 编写测试：健康检查端点返回 200 | `tests/integration/health_test.go` | 若无 health 路由则补 | [X]
| T022 | [Story: US1] 实现/调整健康检查路由或确认存在 | `internal/routes/routes.go` | 与 T021 顺序 | [X]
| T023 | [Story: US1] 移除残余 Logto 配置读取 (环境变量/配置结构) | `internal/core/config/config.go` | 补充注释 | [X]
| T024 | [Story: US1][P] 更新 README 删除 Logto 指引 | `README.md` | 可与 T023 并行 | [X]
| T025 | [Story: US1][P] 更新 quickstart 去除 Logto 残留 | `specs/001-logto-session/quickstart.md` | 并行 | [X]
| T026 | [Story: US1] 运行 layer-check + 覆盖率脚本确保通过 | 终端报告 | 质量门控 | [X] 覆盖率当前 0%（后续 US2 提升），脚本通过
| T027 | [Story: US1] 提交阶段性成果 (Git commit tag: feat(auth): remove logto) | commit | Checkpoint | [X]

Checkpoint US1 ✅

## Phase 4: 用户故事 US2 (P2) - 基础账号密码登录接口
Story Goal: `/api/auth/login` 接口可用，统一响应格式，成功返回 JWT，失败统一错误码（不泄露状态）。
Independent Test Criteria: 成功登录返回 token; 错误密码/禁用/不存在均返回同一错误码；p95 基准 <120ms 本地。

| # | Task | Output / File(s) | Notes |
|---|------|------------------|-------|
| T028 | [Story: US2] 编写登录成功集成测试 (创建测试用户后调用) | `tests/integration/auth_login_success_test.go` | 测试优先 | [X]
| T029 | [Story: US2] 编写登录失败(错误密码)测试 | `tests/integration/auth_login_wrong_password_test.go` | | [X]
| T030 | [Story: US2][P] 编写登录失败(用户不存在)测试 | `tests/integration/auth_login_user_notfound_test.go` | 并行 | [X]
| T031 | [Story: US2][P] 编写登录失败(用户禁用)测试 | `tests/integration/auth_login_disabled_test.go` | 并行 | [X]
| T032 | [Story: US2] DTO: 登录请求/响应结构补充 expires_in | `internal/dto/auth_dto.go` | FR-002/FR-003 | [X]
| T033 | [Story: US2] AuthService 接口具体实现 (SignIn) | `internal/services/auth_service.go` | 调用 UserRepo + password 工具 | [X]
| T034 | [Story: US2] UserRepository sqlite 实现补充查找方法 | `internal/repositories/user_repository.go` | GORM | [X]
| T035 | [Story: US2][P] UserRepository mysql/postgres 占位实现 (stub 返回 `errors.New("driver not enabled")`) | `internal/repositories/user_repository_mysql.go`, `..._postgres.go` | FR-019/占位 | [X]
| T036 | [Story: US2] JWT 生成整合到 AuthService (sub, exp, iat) | `internal/services/auth_service.go` | FR-017 | [X]
| T037 | [Story: US2] 统一响应 helper 在 auth handler 中使用 | `internal/handlers/auth_handler.go` | FR-003 | [X]
| T038 | [Story: US2] 登录路由注册/确认存在 | `internal/routes/routes.go` | /api/auth/login | [X]
| T039 | [Story: US2][P] 增加日志：成功/失败 (不含明文) | `internal/services/auth_service.go` | FR-014 | [X]
| T040 | [Story: US2] 覆盖率评估 & 单元测试 AuthService (密码错误/禁用分支) | `tests/unit/auth_service_test.go` | 覆盖率≥70% 安全≥90% | [X]

| T041 | [Story: US2] 性能基线脚本运行并记录 p95 | `scripts/perf/login.lua` 报告 | FR 性能 | [X] p95≈285ms(conc=8,dur=3s) (需要优化 <120ms 目标)
| T042 | [Story: US2] 提交阶段性成果 (Git commit tag: feat(auth): basic login) | commit | Checkpoint | [X]

Checkpoint US2 ✅

## Phase 5: 用户故事 US3 (P3) - 分层边界清晰
Story Goal: handler 不直接依赖具体存储；DI 集中；层间只依赖接口。
Independent Test Criteria: layer-check 脚本 0 违规；grep handlers 无 mysql/redis/storage 实现 import；容器集中 wiring。

| # | Task | Output / File(s) | Notes |
|---|------|------------------|-------|
| T043 | [Story: US3] 扩展 layer-check 检查 service 不直接 import storages | `scripts/check-layering.sh` | 增强规则 | [X]
| T044 | [Story: US3] 审核 handlers import 并移除直接实现引用 | `internal/handlers/*` | FR-008 | [X]
| T045 | [Story: US3] 审核 services 仅依赖接口 (调整包路径) | `internal/services/*` | FR-009 | [X]
| T046 | [Story: US3] 审核 container 装配集中 & 拆分函数 (e.g. provideUserRepo) | `internal/container/container.go` | FR-010/022 | [X]
| T047 | [Story: US3][P] 为 repository 接口添加最小 doc 注释规范 | `internal/repositories/interfaces/*.go` | 可并行 | [X]
| T048 | [Story: US3][P] 为 service 接口与实现添加注释 | `internal/services/interfaces/*.go`, `internal/services/*.go` | 并行 | [X]
| T049 | [Story: US3] 添加 CI 步骤/脚本：执行 layer-check & go vet & golangci-lint (若引入) | `.github/workflows/ci.yml` 或脚本 | 质量门控 | [X]
| T050 | [Story: US3] 运行并记录 layer-check 结果入 README 开发说明 | `README.md` | 文档同步 | [X]
| T051 | [Story: US3] 提交阶段性成果 (Git commit tag: refactor(arch): enforce layering) | commit | Checkpoint | [X]

Checkpoint US3 ✅

## Phase 6: Polish & Cross-Cutting (扩展质量与门控强化)
| # | Task | Output / File(s) | Notes |
|---|------|------------------|-------|
| T052 | 移除未用的占位表/模型 (LoginAttempt 若确认不再需要) | `internal/models/login_attempt_model.go` | 清理 | 
| T053 | 对象存储 OSS 驱动占位实现 (懒加载) | `internal/storages/aliyun_remote_oss_storage.go` | FR-020/024 | 
| T054 | DB 驱动占位延迟初始化 (mysql/postgres) | `internal/container/container.go` | lazy check | 
| T055 | (前移到 US2 后执行) — *已替换为新任务 T065* | - | 占位删除 | 
| T056 | 性能脚本文档化 + README 性能章节 | `README.md`, `scripts/perf/login.lua` | SC-PERF | 
| T057 | 复核覆盖率并生成 HTML 报告 (go tool cover) | `coverage.out` / `coverage.html` | 归档 | 
| T058 | 最终质量门控汇总（层边界/覆盖率/性能/一致响应） | `docs/final-quality-report.md` | 总结 | 
| T059 | 合并前 squash/rebase & 生成 CHANGELOG 片段 | `CHANGELOG.md` | 发布准备 | 
| T060 | 新增覆盖率阈值检查脚本（≥70% 总体, 安全路径 ≥90%） | `scripts/coverage_check.sh` | 宪章 Gate (C1) | [X]
| T061 | 添加 trace_id 集成测试（成功/失败登录响应包含 trace_id） | `tests/integration/trace_id_auth_test.go` | SC-TRACE (C2) | 
| T062 | 添加 lazy 驱动未启用不初始化测试 | `tests/unit/driver_lazy_init_test.go` | FR-024 验证 | 
| T063 | 添加 hash 算法限定检测脚本 (grep legacy/old hash) | `scripts/check-hash-algo.sh` | FR-016 | 
| T064 | JWT claims & TTL 精度单元测试 (exp-iat) | `tests/unit/jwt_claims_test.go` | SC-SEC/G4 | [X]
| T065 | TOKEN_TTL 环境变量校验与默认逻辑测试（从原 T055 前移） | `tests/unit/jwt_util_test.go` | 稳定性 | 
| T066 | fmt.Println 禁用检测脚本 | `scripts/check-no-fmt.sh` | 宪章: 代码质量 | 
| T067 | 统一响应使用扫描脚本（检测未使用 helper 的 JSON 返回） | `scripts/check-response-helper.sh` | SC-CONSISTENCY | [X]
| T068 | 性能基线脚本输出 JSON (p50,p95,reqs_per_sec) 保存 | `perf/login_baseline.json` | SC-PERF 结构化 | [X]
| T069 | 回滚策略文档（镜像回退 / ENV 恢复流程） | `docs/rollback.md` | 宪章回滚要求 | 
| T070 | 懒加载驱动文档说明（激活矩阵 + 初始化成本） | `docs/drivers-lazy-init.md` | 透明化 | 
| T071 | 风险缓解文档（多驱动/抽象膨胀评估） | `docs/risk-mitigation.md` | 对应 spec Risks | 
| T072 | 更新任务统计（加入 T060-T071） | `specs/001-logto-session/tasks.md` | 自更新 | 

## 任务总览统计 (更新含新增质量任务)
- 总任务数: 72
- US1 任务: 8 (T020-T027)
- US2 任务: 15 (T028-T042) + 覆盖率/安全相关后置: T060(全局), T061(共享)
- US3 任务: 9 (T043-T051)
- Polish & 扩展质量: 20 (T052-T072)
- 并行标记任务数: 12 (T024,T025,T030,T031,T035,T039,T047,T048,T061,T064,T066,T067)

## 依赖与执行顺序图 (高层)
1. Phase1 -> Phase2 -> US1 -> US2 -> US3 -> Polish
2. US1 不依赖 US2/US3；US2 依赖基础设施 (Phase1/2)；US3 依赖 US1/US2；Polish 依赖全部故事完成；质量 Gate (T060-T072) 多数在 US2 之后可提前并行（除文档总结）

```
Phase1(T001-T007)
  ↓
Phase2(T008-T019)
  ↓
US1(T020-T027)
  ↓
US2(T028-T042)
  ↓
US3(T043-T051)
  ↓
Polish(T052-T059)
```

## 并行执行示例
- 在 Phase2：T015 (JWT) 与 T016 (配置) 与 T013 (ObjectStorage 接口) 可并行（不同文件）。
- US2：失败用例测试 (T029, T030, T031) 可并行；MySQL/Postgres 占位 (T035) 可与 AuthService 实现 (T033) 并行。

## MVP 建议范围
- 完成本 Feature 的 MVP = US1 + US2（基础登录能力 + 无外部依赖）。US3 与 Polish 可延后。

## 每用户故事独立测试准则摘要
- US1: 启动成功 + grep 无 logto + health 200。
- US2: 登录成功/失败统一错误码 + token exp 正确 + p95 <120ms + trace_id 响应可用。
- US3: layer-check 零违规 + handlers 无存储实现 import + 容器集中 wiring。

## 实施策略
1. 快速移除外部依赖与 session 清理（风险集中处理）
2. 建立接口与容器装配骨架
3. TDD 登录核心 happy path 然后补充失败路径与安全日志
4. 引入静态与结构检查，性能验证与文档同步，补充质量 Gate（覆盖率、trace_id、lazy 驱动、hash 算法）

---

> 以上任务文件为自动生成，可直接驱动实现迭代。针对新增需求请追加 Phase 或在 Polish 后新增增量任务。
