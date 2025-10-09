# Data Model: Auth & User Refactor

## Entities

### User
| Field | Type | Constraints | Notes |
|-------|------|-------------|-------|
| id | string(UUID) | PK | 生成 uuid v4 |
| username | string | UNIQUE NOT NULL | 登录凭据之一 |
| email | string | UNIQUE NULLABLE | 可选，未来找回密码用 |
| password_hash | string | NOT NULL | bcrypt hash 当前版本 |
| status | enum(active,disabled) | NOT NULL default active | 控制登录可用性 |
| created_at | datetime | NOT NULL | ISO8601 UTC |
| updated_at | datetime | NOT NULL | 更新触发刷新 |

Validation:
- username 长度 3-32；限制字母数字下划线
- password_hash 来自密码工具（明文复杂度校验延后）

### AuthToken (逻辑实体)
| Field | Type | Constraints | Notes |
|-------|------|-------------|-------|
| value | string | JWT | 不落库（当前无持久化） |
| subject_id | string | FK(User.id) | 来源用户 |
| expires_at | datetime | > issued_at | 与 TOKEN_TTL 对齐 |
| scopes | []string | optional | 未来权限扩展 |

### LoginAttempt (可能移除)
当前 feature 将 session 体系移除；LoginAttempt 如仅用于风控暂保留占位：
- 若仍存在表：字段 (id, user_id, success, ip, created_at)
- 后续节流策略启用再纳入核心模型。

## Relationships
- User 1 - N LoginAttempt
- AuthToken 与 User 1 - 1（瞬时颁发，无持久化）

## State Transitions (User.status)
- active → disabled (管理员或安全策略)
- disabled → active (恢复)

## Derived / Future Fields (Not in Scope Now)
- password_algo_version (兼容未来散列升级)
- last_login_at
- mfa_enabled (bool)

## Notes
- 不存储 token blacklist：若未来需要登出/失效策略再扩展表 `token_blacklist`。
- SQLite 下 datetime 统一存 RFC3339 字符串。
