# Research: 移除 Logto 依赖并重构认证分层

## Decisions & Findings

### 1. DB 驱动策略
- Decision: 默认 SQLite (开发极速) + 通过 env `DB_DRIVER` 切换 mysql|postgres|sqlite。
- Rationale: SQLite 零外部依赖，MySQL/PostgreSQL 生产友好。
- Alternatives: 仅支持 MySQL (放弃轻量) / 抽象过度（多层工厂）→ 早期负担高。

### 2. GORM 使用
- Decision: 保留 GORM 作为 ORM/迁移工具。
- Rationale: 快速迭代 + 多数据库兼容；后续可替换为 sqlc 针对性能热点。
- Alternatives: 原生 database/sql (增加样板代码), sqlc (初期建模频繁变更)。

### 3. JWT 签发
- Decision: HS256 + `TOKEN_TTL` 环境变量控制有效期 (默认 60m)。
- Rationale: 简单对称秘钥即可满足当前需求；后续可升级 RS256。
- Alternatives: RS256 (秘钥管理复杂), 自定义 token (生态支持差)。

### 4. Session 移除
- Decision: 完全移除 session 中间件与表引用。
- Rationale: 避免伪状态污染；保持无状态水平扩展路径。
- Alternatives: 保留但禁用（残留复杂度）。

### 5. 多存储抽象 (文件/对象)
- Decision: 定义接口 `ObjectStorage` (Put/Get/Delete/Presign)；实现：local / oss。
- Rationale: 明确最小集合；延后高级特性（多分片/生命周期）。
- Alternatives: 直接散落文件操作 (难替换)、过度抽象 (增加维护成本)。

### 6. 分层
- Decision: handler→service→repository→storage 单向；DI 统一装配。
- Rationale: 降低耦合，驱动可替换性与测试可注入性。
- Alternatives: 服务直接引用 GORM 模型 (耦合 DB 结构)。

### 7. 错误码设计
- Decision: 认证错误统一 `AUTH_xxx` 前缀；通用错误 code=非0 + 统一 message。
- Rationale: 可检索性；避免枚举冲突。
- Alternatives: 数字散列（不可读）。

### 8. Logging & Tracing
- Decision: zap + request middleware 注入 trace_id (uuid)；登录失败也记录 trace_id。
- Rationale: 统一调试链路。
- Alternatives: 按需打印 (难审计)。

### 9. 性能基线
- Decision: 仅对 login 路径做 p95 采样脚本 (wrk/vegeta)；后续新增其他关键点。
- Rationale: 聚焦主要敏感路径。
- Alternatives: 早期全面基准 (投入过大)。

### 10. PostgreSQL 连接池
- Decision: 初始使用 GORM 默认；计划阶段标记 TODO 可调：`MAX_OPEN_CONNS`, `MAX_IDLE_CONNS`。
- Rationale: 延迟优化；避免 premature tuning。
- Alternatives: 立即调优 (数据不足)。

## Open TODO (非阻塞)
- TODO: 添加 presign URL 是否需要 exp 参数策略
- TODO: 引入统一错误码枚举文件
- TODO: benchmark 脚本样例 (auth/login)

## Resolved NEEDS CLARIFICATION
- JWT TTL 可配置 TOKEN_TTL 默认 60m

