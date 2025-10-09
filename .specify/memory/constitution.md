<!--
Sync Impact Report
Version change: 0.1.0 → 1.0.0
Modified principles: (模板初次具体化，无重命名)
Added sections: Core Principles (具体化), 体系分层约束, 研发流程与质量门槛, Governance (具体化)
Removed sections: 无（仅替换占位）
Templates requiring updates:
	.specify/templates/plan-template.md ✅（无需修改：已含 Performance/Constraints，可与新性能要求对齐）
	.specify/templates/spec-template.md ⚠（建议在 Success Criteria 中加入性能/一致性/测试覆盖示例 TODO）
	.specify/templates/tasks-template.md ⚠（建议增加质量/性能基线与代码质量静态检查任务示例 TODO）
	.specify/templates/checklist-template.md ⚠（建议增加“性能基线验证”“用户体验一致性”分类 TODO）
Follow-up TODOs:
	TODO(SPEC_TEMPLATE_ENHANCE): 在 spec-template 增补性能指标/可用性与一致性示例
	TODO(TASKS_TEMPLATE_ENHANCE): 在 tasks-template 增补静态分析、基准、UX 对齐任务示例
	TODO(CHECKLIST_TEMPLATE_ENHANCE): checklist-template 新增 Performance / UX Consistency 分类示例
--> 

# package-base 项目宪章 (Constitution)

## Core Principles

### 1. 代码质量为先 (Code Quality First)
本项目所有 Go 代码必须：
1) 通过编译与 `go vet`, `golangci-lint`（或同等组合）零阻塞级告警；
2) 保持函数/方法内聚：单函数逻辑不超过 ~50 行且单一职责；
3) 公共接口、导出结构体、关键内部包必须有中文注释（格式与项目规范一致）；
4) 禁止无前缀缩写的魔法数字/字符串；
5) 核心领域逻辑（服务层、仓储层）禁止直接使用 `fmt.Println`，统一走 logger 接口；
6) 严格遵循分层依赖方向：`internal/core` ← `internal/services` ← `internal/handlers`，上层不可反向引用下层实现细节；
7) PR 必须展示：变更动机、影响范围、回滚策略（简单描述即可）。
Rationale: 高内聚 + 明确边界提高可维护性，降低回归风险。

### 2. 测试标准与覆盖 (Test Discipline & Coverage)
1) 新增或修改的业务逻辑，必须先补/改对应单元测试（Red-Green-Refactor）；
2) 单元测试覆盖核心业务路径（Auth、Session、TOTP、存储适配器）函数级语句覆盖率目标 ≥ 70%，关键安全与鉴权路径 ≥ 90%；
3) 集成/契约测试：跨组件交互（handler→service→repo）新增行为需有集成测试；
4) 安全相关（密码、JWT、TOTP）必须含边界/失败用例；
5) 禁止随机/时间依赖造成的非确定性测试（需使用伪随机种子或注入时钟接口）。
Rationale: 可预测的回归保障，支持快速迭代与重构。

### 3. 用户体验一致性 (Consistent API & UX)
1) 所有 HTTP Handler 统一返回结构：`{ code:int, message:string, data:any, trace_id?:string }`；
2) 错误码集中枚举，语义稳定（新增不复用旧值）；
3) 分页、排序、过滤参数命名标准化：`page`, `page_size`, `sort`, `order`（asc|desc）；
4) 日志 `trace_id` 贯穿请求链路；
5) 安全失败反馈模糊化（避免泄露存在性）：认证/授权失败统一返回 401/403 + 通用 message；
6) 公共时间字段统一使用 ISO8601 (UTC)；
7) 破坏兼容的接口变更必须给出迁移策略（双写/灰度 header）。
Rationale: 一致性降低客户端适配成本并提升调试效率。

### 4. 性能与资源基线 (Performance & Resource Baseline)
1) 认证/会话创建接口 p95 延迟 < 120ms（本地单实例基准环境）；
2) Redis/MySQL 仓储调用必须封装超时（默认 < 300ms），禁止无限阻塞；
3) 新增数据结构或循环操作需评估 O(复杂度) 并避免非必要 N²；
4) 性能回退 ≥15% (p95 或 QPS) 需在 PR 中说明原因与补救计划；
5) 增加第三方依赖需说明：体积/初始化耗时/安全审计状况；
6) 对外可观测：关键接口暴露计数、延迟、错误率指标（预留接口/埋点函数）。
Rationale: 明确基线防止“性能债”隐性累积。

### 5. 简洁与演进 (Simplicity & Evolution)
1) YAGNI：禁止为“未来不确定需求”预先抽象；
2) 一处使用不成立抽象，多处重复才提取；
3) 功能拆分遵循“领域 → 接口 → 实现”三步；
4) 重构需证明：降低圈复杂度或删除代码 > 新增代码；
5) 守护层次边界：infra/adapter 变更不得传播到 service API。
Rationale: 防止过度设计，保持迭代速度。

## 体系分层约束 (Layered Architecture Constraints)
层次：
- Core/Domain (`internal/core/*`): 纯领域模型与接口定义，无外部框架依赖；
- Repositories & Storages (`internal/repositories`, `internal/core/storages`): 实现持久化，依赖基础驱动，不反向依赖 handlers；
- Services (`internal/services` + `interfaces`): 封装业务编排，仅依赖 Domain Interface 与 Repository Interface；
- Handlers (`internal/handlers`): HTTP/外部适配层，不包含业务规则；
- Middleware (`internal/middleware`): 横切关注点（鉴权、日志、CORS、Session 注入）；
- Utils (`internal/core/utils`): 无状态纯函数；
- Container (`internal/container`): 组合依赖注入，禁止放业务逻辑。
约束：
1) 下层禁止 import 上层实现；
2) Handler 不得直接访问存储实现，必须通过 Service；
3) Service 不直接感知具体 HTTP/框架请求对象；
4) Repository 返回领域对象或 DTO，不泄露底层驱动类型；
5) 跨层数据结构转换只发生在 Handler ↔ Service、Service ↔ Repository 边界；
6) 若需跨层事件，用接口回调或消息通道抽象，禁止直接循环依赖。

## 研发流程与质量门槛 (Workflow & Quality Gates)
阶段：规划 → 设计 → 实施 → 评审 → 基准 → 发布。
质量 Gates：
1) 规划：需求拆分为独立可测试 User Stories；
2) 设计：数据模型 + 合同（接口/DTO）评审通过；
3) 实施前：测试用例（失败态）已提交；
4) PR 校验：Lint、测试、覆盖率阈值、性能基准（关键接口）脚本通过；
5) 合并：无未决高严重级别 Issue；
6) 回滚：提供一键环境变量或镜像回退说明。
评审 Checklist（最少）：
- 是否违反分层依赖？
- 是否缺失测试或降低覆盖率？
- 新依赖风险（安全/许可证）评估？
- 是否引入潜在性能回退？
- 日志是否含敏感信息？

## Governance
治理规则：
1) 本宪章高于其他内部约定，冲突时以宪章为准；
2) 修订流程：提出 Issue → 评估影响与版本级别 → PR → 通过后更新版本与修改日期；
3) 版本策略 (语义化)：
	 - MAJOR：删除/重定义原则，破坏兼容工作流规则；
	 - MINOR：新增原则或扩展约束（本次属于首次具体化 → 1.0.0）；
	 - PATCH：文字澄清、排版、示例修正；
4) 合规审查：每两周进行一次抽样 PR 复核（至少 3 个最近合并项）；
5) 违例处理：需在 PR 描述添加“Violation Justification”表格并记录补救计划；
6) 废弃策略：标记 @Deprecated 后提供替代方案，最少保持两次 MINOR 版本宽限；
7) 指标治理：若持续 2 次审查周期性能基线不达标，列为强制改进 Sprint 任务；
8) 文档同步：修订后 48 小时内同步模板 TODO；
9) Ratification 未知原始日期，使用 TODO 追踪补录。

**Version**: 1.0.0 | **Ratified**: TODO(RATIFICATION_DATE): 尚未提供初始采纳日期 | **Last Amended**: 2025-10-08
