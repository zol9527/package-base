# Implementation Plan: RBAC 权限管理与资源访问控制

**Branch**: `002-rbac` | **Date**: 2025-10-10 | **Spec**: /Users/loyd/development/project/golang/package-base/specs/002-rbac/spec.md
**Input**: Feature specification from `/specs/002-rbac/spec.md`

**Note**: 此文档由 `/speckit.plan` 工作流生成，并补充了与项目宪章对齐的实现规划。

## Summary

- 主要目标：引入基于角色的访问控制（RBAC），实现默认拒绝、精细化权限（资源+动作+效果）、角色继承、作用域绑定、访问决策审计与导出；权限变更需“立即生效”。
- 技术思路（高层）：
  - 在中间件层进行授权判定；服务层暴露“检查权限/查询有效权限”能力；仓储层管理角色、权限、映射与日志。
  - 资源命名采用“域:子域:标识”，支持通配符；评估顺序：显式拒绝 > 最精确匹配 > 通配；结合角色继承聚合后评估。
  - 记录允许/拒绝的访问决策日志，以供审计筛选与导出。

## Technical Context

**Language/Version**: Go 1.24.x（项目既有）  
**Primary Dependencies**: Fiber v2（HTTP）、GORM、zap、validator、JWT、sqlite/mysql/postgres 驱动（项目已有，仅启用 sqlite 驱动待引入说明）  
**Storage**: 复用现有数据库（按项目当前驱动）；新增 RBAC 相关表/迁移（角色、权限、映射、绑定、决策日志）  
**Testing**: Go 自带 testing + 既有集成/契约测试框架（tests/integration, tests/contract, tests/unit）  
**Target Platform**: 本地/服务端（Linux/macOS 单实例）  
**Project Type**: 后端服务（单项目仓库，internal 分层）  
**Performance Goals**: 鉴权判定对端到端延迟影响可忽略（与宪章基线对齐：认证 p95 <120ms；授权侧不引入>15% 回退）  
**Constraints**: 宪章分层约束、统一响应结构、错误码一致、日志 trace_id 贯穿；最小权限默认、强一致的权限变更生效  
**Scale/Scope**: 单租户；未来可扩多租户（本版仅预留设计说明）

## Constitution Check

- 代码质量：遵循 Go 规范、注释要求、禁止魔法数、日志统一 zap、分层依赖方向正确。
- 测试覆盖：新增授权判定与策略评估路径 ≥90%（关键安全路径），总体新增 ≥70%。
- API/UX 一致性：错误码与响应结构统一，401/403 语义稳定；分页/过滤参数规范化。
- 性能基线：关键接口 p95 不回退 >15%；必要时提供补救计划与基准脚本。
- 分层约束：Handler→Service→Repository 单向依赖；Repository 不泄露驱动；中间件做横切授权。
- 质量 Gates：
  - 规划：User Stories 已可独立测试（P1-P3）。
  - 设计：数据模型与合同需评审；
  - 实施前：失败态测试先行；
  - PR：Lint/测试/覆盖率/基准通过。
- Gate 状态：PASS（无未解决澄清项）。

## Project Structure

### Documentation (this feature)

```
specs/002-rbac/
├── plan.md
├── research.md
├── data-model.md
├── quickstart.md
└── contracts/
```

### Source Code (repository root)

```
internal/
├── middleware/           # 授权中间件（横切）
├── services/             # 权限查询/判定服务接口与实现
├── repositories/         # 角色/权限/绑定/日志的持久化
├── models/               # RBAC 实体（GORM/领域模型）
├── handlers/             # 管理端 API（角色/权限/绑定/审计导出）
└── core/                 # 接口定义、utils、常量

tests/
├── integration/
├── contract/
└── unit/
```

**Structure Decision**: 采用单项目后端结构，新增 RBAC 相关的服务/仓储/中间件与模型，严格遵守分层约束。

## Complexity Tracking

无额外复杂度违例；角色继承为必要功能，已通过显式拒绝优先与循环检测控制复杂度。

## Phase 0: Outline & Research

- Unknowns：无（规范已澄清：单租户、角色继承、变更立即生效）。
- Dependencies/Best Practices：
  - 模式：策略评估顺序、显式拒绝优先、资源通配匹配、角色继承聚合、防循环检测。
  - 审计：记录必要字段并可过滤/导出；避免敏感信息泄漏。
  - 性能：缓存与立即生效的平衡（如使用只读缓存时需失效机制与回滚策略）。

Output: /Users/loyd/development/project/golang/package-base/specs/002-rbac/research.md（整合决策与备选方案）

## Phase 1: Design & Contracts

1) data-model.md：实体与关系、校验规则、状态变更  
2) contracts/：依据 FR 导出管理 API 与查询 API（OpenAPI YAML）  
3) quickstart.md：如何开启 RBAC、创建角色、分配、验证、审计与导出

Additionally: 运行 `.specify/scripts/bash/update-agent-context.sh copilot` 更新代理上下文记录。

## 后续 Phase 2（提示）

- 任务分解与里程碑、风险与回滚策略、测试计划细化（由 `/speckit.tasks` 生成，非本命令范围）。
